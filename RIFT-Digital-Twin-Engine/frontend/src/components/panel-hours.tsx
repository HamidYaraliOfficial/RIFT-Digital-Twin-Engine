"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import type { DayHours, OperatingHoursEntry } from "@/lib/types";

const DEFAULT_SCHEDULE: Record<string, DayHours> = {
  "0": { closed: true, open: "", close: "" },
  "1": { closed: false, open: "09:00", close: "18:00" },
  "2": { closed: false, open: "09:00", close: "18:00" },
  "3": { closed: false, open: "09:00", close: "18:00" },
  "4": { closed: false, open: "09:00", close: "18:00" },
  "5": { closed: false, open: "09:00", close: "18:00" },
  "6": { closed: true, open: "", close: "" },
};

// Fully user-entered weekly schedule: the person types in their own open and
// close times for each day (and can mark a day fully closed); RIFT then
// computes, live, whether the twin is open right now and exactly how long
// until the next change — nothing here is guessed or hard-coded.
export function HoursPanel({ twinId }: { twinId: string }) {
  const { t } = useLocale();
  const [entries, setEntries] = useState<OperatingHoursEntry[]>([]);
  const [label, setLabel] = useState("Main Schedule");
  const [timeZone, setTimeZone] = useState("UTC");
  const [schedule, setSchedule] = useState<Record<string, DayHours>>(DEFAULT_SCHEDULE);
  const [, forceTick] = useState(0);

  const load = () => api.get<OperatingHoursEntry[]>(`/api/twins/${twinId}/operating-hours/status`).then(setEntries);

  useEffect(() => {
    load();
    const iv = setInterval(load, 15000);
    const clockIv = setInterval(() => forceTick((x) => x + 1), 1000);
    return () => {
      clearInterval(iv);
      clearInterval(clockIv);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [twinId]);

  const setDay = (day: string, patch: Partial<DayHours>) => {
    setSchedule((s) => ({ ...s, [day]: { ...s[day], ...patch } }));
  };

  const save = async () => {
    await api.post(`/api/twins/${twinId}/operating-hours`, { label, timeZone, schedule });
    load();
  };

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <div className="card p-5">
        <h3 className="text-lg font-semibold">{t.hours.title}</h3>
        <p className="mb-4 text-sm text-[var(--ink-muted)]">{t.hours.subtitle}</p>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <input className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" placeholder={t.hours.label} value={label} onChange={(e) => setLabel(e.target.value)} />
          <input className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" placeholder={t.hours.timezone} value={timeZone} onChange={(e) => setTimeZone(e.target.value)} />
        </div>
        <div className="mt-4 space-y-2">
          {t.hours.days.map((dayName, idx) => {
            const key = String(idx);
            const d = schedule[key];
            return (
              <div key={key} className="flex flex-wrap items-center gap-2 rounded-lg bg-[var(--surface-3)] px-3 py-2 text-sm">
                <span className="w-24 shrink-0 font-medium">{dayName}</span>
                <label className="flex items-center gap-1 text-xs text-[var(--ink-muted)]">
                  <input type="checkbox" checked={d.closed} onChange={(e) => setDay(key, { closed: e.target.checked })} />
                  {t.hours.closedAllDay}
                </label>
                {!d.closed && (
                  <>
                    <input type="time" className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1" value={d.open} onChange={(e) => setDay(key, { open: e.target.value })} />
                    <span className="text-[var(--ink-muted)]">–</span>
                    <input type="time" className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1" value={d.close} onChange={(e) => setDay(key, { close: e.target.value })} />
                  </>
                )}
              </div>
            );
          })}
        </div>
        <button className="btn btn-accent mt-4" onClick={save}>{t.hours.save}</button>
      </div>

      <div className="space-y-3">
        {entries.length === 0 && <div className="card p-5 text-[var(--ink-muted)]">{t.common.none}</div>}
        {entries.map((entry) => (
          <div key={entry.schedule.id} className="card p-5">
            <div className="flex items-center justify-between">
              <h4 className="font-semibold">{entry.schedule.label}</h4>
              <span className={`badge ${entry.status.isOpen ? "status-operational" : "status-failed"}`}>
                {entry.status.isOpen ? t.hours.openNow : t.hours.closedNow}
              </span>
            </div>
            <p className="mt-2 text-sm text-[var(--ink-muted)]">
              {t.hours.nextChange}: {entry.status.nextChangeIsOpen ? t.hours.open : t.hours.close} {t.hours.in} {entry.status.timeUntilNext}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}
