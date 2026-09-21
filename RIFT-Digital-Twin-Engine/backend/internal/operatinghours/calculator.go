// Package operatinghours turns a user-entered weekly schedule (the user
// types in open/close times for each day, entirely themselves — RIFT invents
// nothing) into a live answer: is this twin/entity open right now, when does
// that next change, and how long is left until then.
package operatinghours

import (
	"fmt"
	"strconv"
	"time"

	"rift/internal/model"
)

// Compute evaluates an OperatingHours schedule at time `at` and returns the
// live OperatingStatus. It looks up to 8 days ahead to find the next
// open/close transition so it always finds an answer even for schedules
// that are closed for several consecutive days.
func Compute(h model.OperatingHours, at time.Time) (model.OperatingStatus, error) {
	loc := time.UTC
	if h.TimeZone != "" {
		l, err := time.LoadLocation(h.TimeZone)
		if err != nil {
			return model.OperatingStatus{}, fmt.Errorf("unknown time zone %q: %w", h.TimeZone, err)
		}
		loc = l
	}
	local := at.In(loc)
	isOpenNow := isOpenAt(h, local)

	// Scan forward minute-granularity boundaries by checking each day's
	// open/close instants for up to 8 days, which is cheap and always correct
	// even across "closed all week" schedules.
	next, nextOpen := findNextChange(h, local)

	return model.OperatingStatus{
		IsOpen:            isOpenNow,
		Now:               local,
		NextChangeAt:      next,
		NextChangeIsOpen:  nextOpen,
		TimeUntilNext:     humanDuration(next.Sub(local)),
		TimeUntilNextSecs: int64(next.Sub(local).Seconds()),
	}, nil
}

func isOpenAt(h model.OperatingHours, local time.Time) bool {
	day := h.Schedule[strconv.Itoa(int(local.Weekday()))]
	if day.Closed || day.Open == "" || day.Close == "" {
		// still check "past midnight" window from the previous day
		return isOpenPastMidnightFromYesterday(h, local)
	}
	openT, err1 := parseAtDate(local, day.Open)
	closeT, err2 := parseAtDate(local, day.Close)
	if err1 != nil || err2 != nil {
		return false
	}
	if closeT.Before(openT) || closeT.Equal(openT) {
		// window crosses midnight, e.g. 22:00 -> 06:00
		closeT = closeT.Add(24 * time.Hour)
	}
	if (local.Equal(openT) || local.After(openT)) && local.Before(closeT) {
		return true
	}
	return isOpenPastMidnightFromYesterday(h, local)
}

func isOpenPastMidnightFromYesterday(h model.OperatingHours, local time.Time) bool {
	yesterday := local.AddDate(0, 0, -1)
	day := h.Schedule[strconv.Itoa(int(yesterday.Weekday()))]
	if day.Closed || day.Open == "" || day.Close == "" {
		return false
	}
	openT, err1 := parseAtDate(yesterday, day.Open)
	closeT, err2 := parseAtDate(yesterday, day.Close)
	if err1 != nil || err2 != nil {
		return false
	}
	if !closeT.Before(openT) && !closeT.Equal(openT) {
		return false // did not cross midnight
	}
	closeT = closeT.Add(24 * time.Hour)
	return local.After(openT) && local.Before(closeT)
}

func findNextChange(h model.OperatingHours, local time.Time) (time.Time, bool) {
	currentlyOpen := isOpenAt(h, local)
	for d := 0; d <= 8; d++ {
		day := local.AddDate(0, 0, d)
		sched := h.Schedule[strconv.Itoa(int(day.Weekday()))]
		if sched.Closed || sched.Open == "" || sched.Close == "" {
			continue
		}
		openT, err1 := parseAtDate(day, sched.Open)
		closeT, err2 := parseAtDate(day, sched.Close)
		if err1 != nil || err2 != nil {
			continue
		}
		if closeT.Before(openT) || closeT.Equal(openT) {
			closeT = closeT.Add(24 * time.Hour)
		}
		for _, candidate := range []struct {
			t      time.Time
			isOpen bool
		}{{openT, true}, {closeT, false}} {
			if candidate.t.After(local) && candidate.isOpen != currentlyOpen {
				return candidate.t, candidate.isOpen
			}
			// even if isOpen state matches, the very next boundary after "now"
			// that flips relative to currentlyOpen is what we want; if it
			// doesn't flip, keep scanning later boundaries/days.
		}
	}
	// No schedule at all / permanently closed: report "now" with no change.
	return local, currentlyOpen
}

func parseAtDate(date time.Time, hhmm string) (time.Time, error) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location()), nil
}

func humanDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}
