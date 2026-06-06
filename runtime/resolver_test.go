package main

import "testing"

func resolverAPI(t *testing.T) *API {
	t.Helper()
	model, err := LoadModel()
	if err != nil {
		t.Fatalf("LoadModel: %v", err)
	}
	g := NewGraph(nil)
	model.Seed(g)
	a := &API{Model: model, Graph: g}
	if _, err := a.RegisterAgent("agent:resolver", "Resolver", []string{"capability"}, nil, []string{"fabric:policy:tool-validity"}); err != nil {
		t.Fatalf("RegisterAgent: %v", err)
	}
	// three identities: two share an email (duplicates), one is unique.
	a.Graph.Put(&Record{Table: "identity", ID: "identity:b", Fields: map[string]any{"email": "ada@x.io"}}, false)
	a.Graph.Put(&Record{Table: "identity", ID: "identity:a", Fields: map[string]any{"email": "ada@x.io"}}, false)
	a.Graph.Put(&Record{Table: "identity", ID: "identity:c", Fields: map[string]any{"email": "grace@x.io"}}, false)
	return a
}

// Propose must NOT commit: after a scan, the duplicate records are untouched.
func TestResolverProposeDoesNotCommit(t *testing.T) {
	a := resolverAPI(t)
	props, err := a.ResolverScan("agent:resolver", "identity", "email")
	if err != nil {
		t.Fatalf("ResolverScan: %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("want 1 merge proposal (the ada@x.io group), got %d", len(props))
	}
	if s, _ := props[0].Fields["status"].(string); s != "proposed" {
		t.Fatalf("proposal status = %q, want proposed", s)
	}
	// deterministic canonical = lowest id = identity:a
	if c, _ := props[0].Fields["targetId"].(string); c != "identity:a" {
		t.Fatalf("canonical = %q, want identity:a", c)
	}
	// URAP: nothing committed — the duplicate is unchanged.
	if dup := a.Graph.Get("identity:b"); dup.Fields["mergedInto"] != nil {
		t.Fatalf("propose committed a change (identity:b.mergedInto set) — violates URAP")
	}
}

// Admit is the only path that commits: the human admits, the merge applies,
// and an immutable Event is recorded.
func TestAdmitCommits(t *testing.T) {
	a := resolverAPI(t)
	props, _ := a.ResolverScan("agent:resolver", "identity", "email")
	ev, err := a.Admit(props[0].ID, "human:guard")
	if err != nil {
		t.Fatalf("Admit: %v", err)
	}
	if ty, _ := ev.Fields["type"].(string); ty != "proposal.admitted" {
		t.Fatalf("event type = %q, want proposal.admitted", ty)
	}
	dup := a.Graph.Get("identity:b")
	if dup.Fields["mergedInto"] != "identity:a" {
		t.Fatalf("after admit, identity:b.mergedInto = %v, want identity:a", dup.Fields["mergedInto"])
	}
	if p := a.Graph.Get(props[0].ID); p.Fields["status"] != "admitted" {
		t.Fatalf("proposal status = %v, want admitted", p.Fields["status"])
	}
	// re-admit must fail (commit is final)
	if _, err := a.Admit(props[0].ID, "human:guard"); err == nil {
		t.Fatalf("re-admitting a committed proposal should fail")
	}
}

// Reject discards without committing.
func TestRejectDoesNotCommit(t *testing.T) {
	a := resolverAPI(t)
	props, _ := a.ResolverScan("agent:resolver", "identity", "email")
	if _, err := a.Reject(props[0].ID, "human:guard"); err != nil {
		t.Fatalf("Reject: %v", err)
	}
	if dup := a.Graph.Get("identity:b"); dup.Fields["mergedInto"] != nil {
		t.Fatalf("reject committed a change — should be a no-op on reality")
	}
	if p := a.Graph.Get(props[0].ID); p.Fields["status"] != "rejected" {
		t.Fatalf("proposal status = %v, want rejected", p.Fields["status"])
	}
}

// Determinism: a second scan over the same state yields the same proposal shape.
func TestResolverDeterministic(t *testing.T) {
	a1 := resolverAPI(t)
	a2 := resolverAPI(t)
	p1, _ := a1.ResolverScan("agent:resolver", "identity", "email")
	p2, _ := a2.ResolverScan("agent:resolver", "identity", "email")
	if len(p1) != len(p2) || len(p1) != 1 {
		t.Fatalf("nondeterministic count: %d vs %d", len(p1), len(p2))
	}
	if p1[0].Fields["targetId"] != p2[0].Fields["targetId"] {
		t.Fatalf("nondeterministic canonical: %v vs %v", p1[0].Fields["targetId"], p2[0].Fields["targetId"])
	}
}
