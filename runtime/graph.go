package main

import (
	"sort"
	"sync"
)

// Record is a multi-model document: a typed record (a row in `table`) carrying
// free-form fields. The same store also holds graph edges between records, so a
// single Graph is simultaneously a document store and a property graph — the
// "multi-model graph runtime" the Fabric model is designed for.
type Record struct {
	Table  string         `json:"table"`
	ID     string         `json:"id"`
	Fields map[string]any `json:"fields,omitempty"`
}

// Edge is a typed, directed relationship between two records (referenced by id).
type Edge struct {
	Rel   string         `json:"rel"`
	From  string         `json:"from"`
	To    string         `json:"to"`
	Props map[string]any `json:"props,omitempty"`
}

// Graph is the in-memory multi-model store: records indexed by id, plus an
// adjacency index for fast traversal in either direction. It is safe for
// concurrent use and writes through an optional append-only Log for durability.
type Graph struct {
	mu      sync.RWMutex
	records map[string]*Record
	out     map[string][]*Edge
	in      map[string][]*Edge
	edges   []*Edge
	log     *Log
}

// NewGraph returns an empty graph. Pass a Log to persist instance mutations;
// pass nil for an ephemeral, in-memory-only runtime.
func NewGraph(log *Log) *Graph {
	return &Graph{
		records: map[string]*Record{},
		out:     map[string][]*Edge{},
		in:      map[string][]*Edge{},
		log:     log,
	}
}

// Put inserts or replaces a record. When persist is true the write is appended
// to the log; replay passes persist=false so the log is not rewritten.
func (g *Graph) Put(r *Record, persist bool) *Record {
	g.mu.Lock()
	g.records[r.ID] = r
	g.mu.Unlock()
	if persist && g.log != nil {
		g.log.Append(logEntry{Op: "put", Record: r})
	}
	return r
}

// Relate adds a directed edge. Records are not required to exist first; this
// keeps ingestion order-independent, matching how graph data actually arrives.
func (g *Graph) Relate(e *Edge, persist bool) *Edge {
	g.mu.Lock()
	g.edges = append(g.edges, e)
	g.out[e.From] = append(g.out[e.From], e)
	g.in[e.To] = append(g.in[e.To], e)
	g.mu.Unlock()
	if persist && g.log != nil {
		g.log.Append(logEntry{Op: "relate", Edge: e})
	}
	return e
}

// Get returns a record by id, or nil if absent.
func (g *Graph) Get(id string) *Record {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.records[id]
}

// Table returns all records of one table (type), sorted by id for stable output.
func (g *Graph) Table(table string) []*Record {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var out []*Record
	for _, r := range g.records {
		if r.Table == table {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Tables returns the distinct table names present, sorted.
func (g *Graph) Tables() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	seen := map[string]bool{}
	for _, r := range g.records {
		seen[r.Table] = true
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// Traversal is the connected subgraph returned by a walk from a start record.
type Traversal struct {
	Start   string    `json:"start"`
	Depth   int       `json:"depth"`
	Records []*Record `json:"records"`
	Edges   []*Edge   `json:"edges"`
}

// Traverse walks the property graph breadth-first from id up to depth hops.
// dir selects which edges to follow: "out", "in", or "both".
func (g *Graph) Traverse(id string, depth int, dir string) *Traversal {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if dir == "" {
		dir = "both"
	}
	visited := map[string]bool{id: true}
	var walkEdges []*Edge
	type item struct {
		id string
		d  int
	}
	queue := []item{{id, 0}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.d >= depth {
			continue
		}
		var next []*Edge
		if dir == "out" || dir == "both" {
			next = append(next, g.out[cur.id]...)
		}
		if dir == "in" || dir == "both" {
			next = append(next, g.in[cur.id]...)
		}
		for _, e := range next {
			walkEdges = append(walkEdges, e)
			peer := e.To
			if e.To == cur.id {
				peer = e.From
			}
			if !visited[peer] {
				visited[peer] = true
				queue = append(queue, item{peer, cur.d + 1})
			}
		}
	}
	ids := make([]string, 0, len(visited))
	for v := range visited {
		ids = append(ids, v)
	}
	sort.Strings(ids)
	recs := make([]*Record, 0, len(ids))
	for _, v := range ids {
		if r := g.records[v]; r != nil {
			recs = append(recs, r)
		}
	}
	return &Traversal{Start: id, Depth: depth, Records: recs, Edges: walkEdges}
}

// Stats reports store size for /health and the GraphQL stats query.
func (g *Graph) Stats() map[string]any {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return map[string]any{
		"records": len(g.records),
		"edges":   len(g.edges),
		"tables":  len(g.Tables()),
	}
}
