// Package graph implements the Dependency Graph Engine: traversal, shortest
// path, reachability, dependency depth, critical path and failure
// propagation over the Relationship edges stored in the Registry. It also
// backs the Impact Analysis Engine.
package graph

import (
	"rift/internal/model"
)

// Graph is a lightweight adjacency-list view built on demand from a
// Relationship slice; it never mutates or owns the Registry.
type Graph struct {
	out map[string][]*model.Relationship // sourceID -> outgoing edges
	in  map[string][]*model.Relationship // targetID -> incoming edges
}

func Build(relationships []*model.Relationship) *Graph {
	g := &Graph{out: map[string][]*model.Relationship{}, in: map[string][]*model.Relationship{}}
	for _, r := range relationships {
		g.out[r.SourceID] = append(g.out[r.SourceID], r)
		g.in[r.TargetID] = append(g.in[r.TargetID], r)
	}
	return g
}

// Neighbors returns the entities directly reachable from id, following
// outgoing edges only (i.e. "id contains X", "id feeds X").
func (g *Graph) Neighbors(id string) []string {
	var out []string
	for _, r := range g.out[id] {
		out = append(out, r.TargetID)
	}
	return out
}

// Dependents returns entities that point at id (i.e. "X depends_on id",
// "X powered_by id") — the set impact analysis needs when id goes down.
func (g *Graph) Dependents(id string) []string {
	var out []string
	for _, r := range g.in[id] {
		if r.Type == model.RelDependsOn || r.Type == model.RelPoweredBy || r.Type == model.RelFeeds || r.Type == model.RelControlledBy || r.Type == model.RelConsumes {
			out = append(out, r.SourceID)
		}
	}
	return out
}

// ImpactAnalysis performs a breadth-first traversal of Dependents starting
// at rootID, returning every entity transitively affected along with the
// hop distance ("dependency depth") and one illustrative path.
type ImpactResult struct {
	EntityID string   `json:"entityId"`
	Depth    int      `json:"depth"`
	Path     []string `json:"path"`
}

func (g *Graph) ImpactAnalysis(rootID string, maxDepth int) []ImpactResult {
	visited := map[string]bool{rootID: true}
	type qitem struct {
		id    string
		depth int
		path  []string
	}
	queue := []qitem{{rootID, 0, []string{rootID}}}
	var results []ImpactResult
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.depth >= maxDepth {
			continue
		}
		for _, dep := range g.Dependents(cur.id) {
			if visited[dep] {
				continue
			}
			visited[dep] = true
			path := append(append([]string{}, cur.path...), dep)
			results = append(results, ImpactResult{EntityID: dep, Depth: cur.depth + 1, Path: path})
			queue = append(queue, qitem{dep, cur.depth + 1, path})
		}
	}
	return results
}

// ShortestPath performs an unweighted BFS from src to dst over outgoing
// edges and returns the path (inclusive), or nil if unreachable.
func (g *Graph) ShortestPath(src, dst string) []string {
	if src == dst {
		return []string{src}
	}
	visited := map[string]bool{src: true}
	parent := map[string]string{}
	queue := []string{src}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, n := range g.Neighbors(cur) {
			if visited[n] {
				continue
			}
			visited[n] = true
			parent[n] = cur
			if n == dst {
				return reconstruct(parent, src, dst)
			}
			queue = append(queue, n)
		}
	}
	return nil
}

func reconstruct(parent map[string]string, src, dst string) []string {
	path := []string{dst}
	cur := dst
	for cur != src {
		cur = parent[cur]
		path = append([]string{cur}, path...)
	}
	return path
}

// Reachable returns every entity reachable from id via outgoing edges
// (used e.g. to answer "what does this power grid segment feed?").
func (g *Graph) Reachable(id string) []string {
	visited := map[string]bool{id: true}
	queue := []string{id}
	var out []string
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, n := range g.Neighbors(cur) {
			if !visited[n] {
				visited[n] = true
				out = append(out, n)
				queue = append(queue, n)
			}
		}
	}
	return out
}

// Depth returns the maximum dependency chain length starting at id
// (longest path over Dependents), used for "dependency depth" reporting.
func (g *Graph) Depth(id string) int {
	max := 0
	var visit func(cur string, d int, seen map[string]bool)
	visit = func(cur string, d int, seen map[string]bool) {
		if d > max {
			max = d
		}
		for _, dep := range g.Dependents(cur) {
			if seen[dep] {
				continue
			}
			seen[dep] = true
			visit(dep, d+1, seen)
		}
	}
	visit(id, 0, map[string]bool{id: true})
	return max
}
