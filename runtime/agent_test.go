package main

import "testing"

func TestRegisterAgentWiresGovernance(t *testing.T) {
	api, _ := newTestAPI(t)
	_, err := api.RegisterAgent("agent:ingestor", "Ingestor",
		[]string{"capability"}, []string{"objective"}, []string{"policy"})
	if err != nil {
		t.Fatal(err)
	}
	rec := api.Graph.Get("agent:ingestor")
	if rec == nil || rec.Table != "agent" {
		t.Fatalf("agent record not stored: %+v", rec)
	}
	// Governance + self-describing edges must exist out of the agent. The
	// targets are references (no record required), so assert on edges.
	tr := api.Graph.Traverse("agent:ingestor", 1, "out")
	want := map[string]string{"executes": "capability", "pursues": "objective", "governedBy": "policy", "instanceOf": "primitive:agent"}
	for rel, to := range want {
		if !hasEdge(tr.Edges, rel, to) {
			t.Fatalf("agent should have edge %s->%s: %+v", rel, to, tr.Edges)
		}
	}
}

func hasEdge(edges []*Edge, rel, to string) bool {
	for _, e := range edges {
		if e.Rel == rel && e.To == to {
			return true
		}
	}
	return false
}

func TestRegisterAgentValidation(t *testing.T) {
	api, _ := newTestAPI(t)
	if _, err := api.RegisterAgent("", "x", nil, nil, nil); err == nil {
		t.Fatal("empty agent id should error")
	}
}

func TestActRequiresRegistration(t *testing.T) {
	api, _ := newTestAPI(t)
	if _, err := api.Act("agent:ghost", "ingest", "", "", nil); err == nil {
		t.Fatal("acting as an unregistered agent should error")
	}
}

func TestActRecordsAuditedMemory(t *testing.T) {
	api, _ := newTestAPI(t)
	if _, err := api.RegisterAgent("agent:a", "A", nil, nil, []string{"policy:gdpr"}); err != nil {
		t.Fatal(err)
	}
	event, err := api.Act("agent:a", "ingested", "dataset", "ds:sales", map[string]any{"rows": 10})
	if err != nil {
		t.Fatal(err)
	}
	if event.Table != "event" || event.Fields["actor"] != "agent:a" {
		t.Fatalf("unexpected audit event: %+v", event)
	}
	// The action's effect (the dataset) and the governance context are recorded.
	if api.Graph.Get("ds:sales") == nil {
		t.Fatal("agent action did not create the target record")
	}
	if event.Fields["governedBy"] == nil {
		t.Fatal("audit event should carry governance context")
	}
	// The event is reachable as the agent's memory.
	tr := api.Graph.Traverse("agent:a", 1, "out")
	if !hasRecord(tr.Records, event.ID) {
		t.Fatalf("event should be linked as agent memory: %v", recordIDs(tr.Records))
	}
}
