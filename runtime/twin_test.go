package main

import "testing"

// twinAPI builds an ephemeral runtime with: a proposing agent, a domain
// (acme.com) whose owner (identity:ada) is identity-verified, and an entity to
// twin (identity:bob in acme.com).
func twinAPI(t *testing.T) *API {
	t.Helper()
	model, err := LoadModel()
	if err != nil {
		t.Fatalf("LoadModel: %v", err)
	}
	g := NewGraph(nil)
	model.Seed(g)
	a := &API{Model: model, Graph: g}
	if _, err := a.RegisterAgent("agent:twinner", "Twinner", []string{"capability"}, nil, nil); err != nil {
		t.Fatalf("RegisterAgent: %v", err)
	}
	if _, _, err := a.RegisterDomain("acme.com", "identity:ada"); err != nil {
		t.Fatalf("RegisterDomain: %v", err)
	}
	if _, err := a.VerifyIdentity("identity:ada", "domain-control", "dns-txt"); err != nil {
		t.Fatalf("VerifyIdentity: %v", err)
	}
	a.Graph.Put(&Record{Table: "identity", ID: "identity:bob", Fields: map[string]any{
		"email": "bob@acme.com", "displayName": "Bob",
	}}, false)
	return a
}

// Propose must NOT commit: no twin record exists until the owner admits.
func TestTwinProposeDoesNotCommit(t *testing.T) {
	a := twinAPI(t)
	prop, err := a.TwinPropose("agent:twinner", "identity", "identity:bob", map[string]any{"price": -1.0}, "")
	if err != nil {
		t.Fatalf("TwinPropose: %v", err)
	}
	if s, _ := prop.Fields["status"].(string); s != "proposed" {
		t.Fatalf("status = %q, want proposed", s)
	}
	// domain derived from bob@acme.com
	if d, _ := prop.Fields["fields"].(map[string]any)["domain"].(string); d != "acme.com" {
		t.Fatalf("derived domain = %q, want acme.com", d)
	}
	if a.Graph.Get("twin:identity:identity:bob") != nil {
		t.Fatalf("propose committed a twin — violates URAP")
	}
}

// Only the platform owner of the twin's domain may admit; others are rejected
// and nothing is committed.
func TestTwinAdmitRequiresDomainOwner(t *testing.T) {
	a := twinAPI(t)
	prop, _ := a.TwinPropose("agent:twinner", "identity", "identity:bob", nil, "")
	if _, err := a.Admit(prop.ID, "identity:eve"); err == nil {
		t.Fatalf("admit by a non-owner should fail")
	}
	if a.Graph.Get("twin:identity:identity:bob") != nil {
		t.Fatalf("a failed admit must not commit a twin")
	}
}

// The owner admits: the twin commits and the admit Event carries a valid
// ed25519 signature; a tampered payload fails verification.
func TestTwinAdmitByOwnerCommitsAndSigns(t *testing.T) {
	a := twinAPI(t)
	prop, _ := a.TwinPropose("agent:twinner", "identity", "identity:bob", nil, "")
	ev, err := a.Admit(prop.ID, "identity:ada")
	if err != nil {
		t.Fatalf("owner admit: %v", err)
	}
	if a.Graph.Get("twin:identity:identity:bob") == nil {
		t.Fatalf("owner admit should commit the twin")
	}
	pub, _ := ev.Fields["pubkey"].(string)
	sig, _ := ev.Fields["signature"].(string)
	payload, _ := ev.Fields["payload"].(string)
	if pub == "" || sig == "" || payload == "" {
		t.Fatalf("admit Event missing signature material: %+v", ev.Fields)
	}
	if !VerifySignature(pub, []byte(payload), sig) {
		t.Fatalf("signature should verify against the signed payload")
	}
	if VerifySignature(pub, []byte(payload+"x"), sig) {
		t.Fatalf("tampered payload must not verify")
	}
}

// Identity verification gates admission: an unverified owner cannot admit.
func TestTwinAdmitRequiresVerifiedOwner(t *testing.T) {
	a := twinAPI(t)
	if _, _, err := a.RegisterDomain("beta.io", "identity:carol"); err != nil {
		t.Fatalf("RegisterDomain: %v", err)
	}
	a.Graph.Put(&Record{Table: "identity", ID: "identity:dan", Fields: map[string]any{"email": "dan@beta.io"}}, false)
	prop, _ := a.TwinPropose("agent:twinner", "identity", "identity:dan", nil, "")
	if _, err := a.Admit(prop.ID, "identity:carol"); err == nil {
		t.Fatalf("admit by an unverified owner should fail")
	}
}

// The preference model is deterministic and justified.
func TestTwinDecideDeterministicAndJustified(t *testing.T) {
	a := twinAPI(t)
	a.Graph.Put(&Record{Table: "twin", ID: "twin:x", Fields: map[string]any{
		"preferences": map[string]any{"sustainability": 2.0, "price": -1.0},
	}}, false)
	options := []map[string]any{
		{"id": "green", "attrs": map[string]any{"sustainability": 5.0, "price": 3.0}}, // 2*5 - 1*3 = 7
		{"id": "cheap", "attrs": map[string]any{"sustainability": 1.0, "price": 1.0}}, // 2*1 - 1*1 = 1
	}
	out, err := a.TwinDecide("twin:x", options)
	if err != nil {
		t.Fatalf("TwinDecide: %v", err)
	}
	if out["choice"] != "green" {
		t.Fatalf("choice = %v, want green", out["choice"])
	}
	out2, _ := a.TwinDecide("twin:x", options)
	if out2["choice"] != out["choice"] {
		t.Fatalf("nondeterministic choice: %v vs %v", out2["choice"], out["choice"])
	}
	if j, ok := out["justification"].([]any); !ok || len(j) == 0 {
		t.Fatalf("decision should carry a justification: %v", out["justification"])
	}
}

// Twin of everything: a non-identity entity can be twinned (with an explicit
// domain) and admitted by that domain's owner.
func TestTwinOfEverything(t *testing.T) {
	a := twinAPI(t)
	a.Graph.Put(&Record{Table: "dataset", ID: "dataset:sales", Fields: map[string]any{"rows": 100}}, false)
	prop, err := a.TwinPropose("agent:twinner", "dataset", "dataset:sales", nil, "acme.com")
	if err != nil {
		t.Fatalf("TwinPropose(dataset): %v", err)
	}
	if _, err := a.Admit(prop.ID, "identity:ada"); err != nil {
		t.Fatalf("owner admit of dataset twin: %v", err)
	}
	if a.Graph.Get("twin:dataset:dataset:sales") == nil {
		t.Fatalf("twin of a non-identity entity should commit")
	}
}
