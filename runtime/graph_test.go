package main

import (
	"path/filepath"
	"testing"
)

func TestGraphPutAndGet(t *testing.T) {
	g := NewGraph(nil)
	g.Put(&Record{Table: "identity", ID: "a", Fields: map[string]any{"name": "Ada"}}, false)
	got := g.Get("a")
	if got == nil || got.Fields["name"] != "Ada" {
		t.Fatalf("Get(a) = %+v, want name=Ada", got)
	}
	if g.Get("missing") != nil {
		t.Fatalf("Get(missing) should be nil")
	}
}

func TestGraphTableSorted(t *testing.T) {
	g := NewGraph(nil)
	g.Put(&Record{Table: "x", ID: "b"}, false)
	g.Put(&Record{Table: "x", ID: "a"}, false)
	g.Put(&Record{Table: "y", ID: "c"}, false)
	xs := g.Table("x")
	if len(xs) != 2 || xs[0].ID != "a" || xs[1].ID != "b" {
		t.Fatalf("Table(x) not sorted: %+v", xs)
	}
	if tables := g.Tables(); len(tables) != 2 || tables[0] != "x" || tables[1] != "y" {
		t.Fatalf("Tables() = %v", tables)
	}
}

func TestGraphTraverse(t *testing.T) {
	g := NewGraph(nil)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.Put(&Record{Table: "n", ID: id}, false)
	}
	g.Relate(&Edge{From: "a", Rel: "to", To: "b"}, false)
	g.Relate(&Edge{From: "b", Rel: "to", To: "c"}, false)
	g.Relate(&Edge{From: "c", Rel: "to", To: "d"}, false)

	out := g.Traverse("a", 2, "out")
	if !hasRecord(out.Records, "c") || hasRecord(out.Records, "d") {
		t.Fatalf("depth-2 out traversal should reach c but not d: %v", recordIDs(out.Records))
	}
	in := g.Traverse("d", 1, "in")
	if !hasRecord(in.Records, "c") || hasRecord(in.Records, "b") {
		t.Fatalf("depth-1 in traversal from d should reach only c: %v", recordIDs(in.Records))
	}
}

func TestLogReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "graph.jsonl")
	log, err := OpenLog(path)
	if err != nil {
		t.Fatal(err)
	}
	g := NewGraph(log)
	g.Put(&Record{Table: "identity", ID: "x", Fields: map[string]any{"k": "v"}}, true)
	g.Relate(&Edge{From: "x", Rel: "performed", To: "e"}, true)
	log.Close()

	log2, err := OpenLog(path)
	if err != nil {
		t.Fatal(err)
	}
	defer log2.Close()
	g2 := NewGraph(log2)
	if err := log2.Replay(g2); err != nil {
		t.Fatal(err)
	}
	if g2.Get("x") == nil {
		t.Fatal("replayed record x missing")
	}
	if st := g2.Stats(); st["edges"].(int) != 1 {
		t.Fatalf("replayed edges = %v, want 1", st["edges"])
	}
}

func hasRecord(rs []*Record, id string) bool {
	for _, r := range rs {
		if r.ID == id {
			return true
		}
	}
	return false
}

func recordIDs(rs []*Record) []string {
	ids := make([]string, len(rs))
	for i, r := range rs {
		ids[i] = r.ID
	}
	return ids
}
