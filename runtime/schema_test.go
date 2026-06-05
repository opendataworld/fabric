package main

import (
	"testing"

	"github.com/graphql-go/graphql"
)

// newTestAPI builds an ephemeral runtime (no durable log) seeded from the real
// canonical model, so GraphQL tests exercise the same schema as production.
func newTestAPI(t *testing.T) (*API, graphql.Schema) {
	t.Helper()
	model, err := LoadModel()
	if err != nil {
		t.Fatalf("LoadModel: %v", err)
	}
	g := NewGraph(nil)
	model.Seed(g)
	api := &API{Model: model, Graph: g}
	schema, err := api.BuildSchema()
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	return api, schema
}

func do(t *testing.T, schema graphql.Schema, q string) *graphql.Result {
	t.Helper()
	r := graphql.Do(graphql.Params{Schema: schema, RequestString: q})
	if len(r.Errors) > 0 {
		t.Fatalf("query errors: %v", r.Errors)
	}
	return r
}

func TestQueryClasses(t *testing.T) {
	_, schema := newTestAPI(t)
	r := do(t, schema, `{ classes { id name } }`)
	data := r.Data.(map[string]any)
	classes := data["classes"].([]any)
	if len(classes) < 35 {
		t.Fatalf("expected >=35 classes, got %d", len(classes))
	}
}

func TestQueryResolve(t *testing.T) {
	_, schema := newTestAPI(t)
	r := do(t, schema, `{ resolve(name:"identity", depth:2) { resolved } }`)
	resolved := r.Data.(map[string]any)["resolve"].(map[string]any)["resolved"].([]any)
	found := false
	for _, v := range resolved {
		if v == "capability" {
			found = true
		}
	}
	if !found {
		t.Fatalf("resolve(identity,2) should reach capability: %v", resolved)
	}
}

func TestMutationSignupRoundTrip(t *testing.T) {
	api, schema := newTestAPI(t)
	r := do(t, schema, `mutation { signup(name:"Grace Hopper", email:"grace@navy.mil") { id table fields } }`)
	rec := r.Data.(map[string]any)["signup"].(map[string]any)
	if rec["table"] != "identity" {
		t.Fatalf("signup record table = %v, want identity", rec["table"])
	}
	id := rec["id"].(string)

	// The identity, its event, and the audit edges must now be in the graph.
	if api.Graph.Get(id) == nil {
		t.Fatalf("signup did not store identity %s", id)
	}
	if len(api.Graph.Table("event")) != 1 {
		t.Fatalf("signup did not append an audit event")
	}
	tr := api.Graph.Traverse(id, 1, "out")
	if !hasRecord(tr.Records, "primitive:identity") {
		t.Fatalf("signup identity should link to its primitive: %v", recordIDs(tr.Records))
	}
}

func TestMutationSignupValidation(t *testing.T) {
	_, schema := newTestAPI(t)
	r := graphql.Do(graphql.Params{Schema: schema, RequestString: `mutation { signup(name:"", email:"nope") { id } }`})
	if len(r.Errors) == 0 {
		t.Fatal("signup with bad input should error")
	}
}
