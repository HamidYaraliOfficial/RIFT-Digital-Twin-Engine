// Package rootcause implements the Root Cause Analysis Engine and the Event
// Correlation Engine: given an incident's events and the dependency graph,
// it proposes ranked candidate causes with evidence and a confidence score.
// It never claims certainty — every candidate is explicitly a hypothesis.
package rootcause

import (
	"fmt"
	"sort"
	"time"

	"rift/internal/graph"
	"rift/internal/model"
)

// Analyze looks at the events inside [windowStart, windowEnd] touching any
// entity connected to focusEntityID, and ranks candidates by: (1) temporal
// precedence (earlier severe events score higher), (2) graph proximity
// (closer upstream dependencies score higher), (3) severity.
func Analyze(g *graph.Graph, events []model.Event, focusEntityID string, windowStart, windowEnd time.Time) []model.RootCause {
	// Build the set of entities upstream of focusEntityID (things focus
	// depends on) plus focusEntityID itself — a failure there could explain focus's symptoms.
	upstream := map[string]int{focusEntityID: 0}
	for _, id := range g.Reachable(focusEntityID) {
		// Reachable follows outgoing edges (contains/feeds/powers); also walk
		// Dependents to capture "X depends_on focus" -> focus depends on nothing new,
		// but the reverse (what focus depends on) is best expressed by inverse edges,
		// which for our directional convention are already the outgoing side for
		// depends_on/powered_by (focus -> other means focus depends_on other).
		if _, ok := upstream[id]; !ok {
			upstream[id] = 1
		}
	}

	type scored struct {
		rc    model.RootCause
		score float64
	}
	var candidates []scored
	seen := map[string]bool{}
	for _, e := range events {
		if e.Timestamp.Before(windowStart) || e.Timestamp.After(windowEnd) {
			continue
		}
		entity := e.SourceEntityID
		if entity == "" {
			entity = e.TargetEntityID
		}
		if entity == "" {
			continue
		}
		depth, related := upstream[entity]
		if !related && entity != focusEntityID {
			continue
		}
		key := entity + "|" + e.Type
		if seen[key] {
			continue
		}
		seen[key] = true

		sevWeight := map[model.Severity]float64{model.SeverityInfo: 1, model.SeverityWarning: 2, model.SeverityCritical: 4}[e.Severity]
		proximity := 1.0 / float64(depth+1)
		recency := 1.0 - float64(e.Timestamp.Sub(windowStart))/float64(windowEnd.Sub(windowStart)+1)
		score := sevWeight*2 + proximity*3 + recency*2

		candidates = append(candidates, scored{
			rc: model.RootCause{
				EntityID: entity,
				Reason:   fmt.Sprintf("%s event on %s preceded/overlapped the incident window", e.Type, entity),
				Evidence: []string{fmt.Sprintf("event %s at %s (severity=%s)", e.ID, e.Timestamp.Format(time.RFC3339), e.Severity)},
			},
			score: score,
		})
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })

	maxScore := 0.0
	for _, c := range candidates {
		if c.score > maxScore {
			maxScore = c.score
		}
	}
	out := make([]model.RootCause, 0, len(candidates))
	for i, c := range candidates {
		if i >= 8 {
			break
		}
		conf := 0.5
		if maxScore > 0 {
			conf = clamp01(c.score / maxScore)
		}
		c.rc.Confidence = round2(conf)
		out = append(out, c.rc)
	}
	return out
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func round2(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
