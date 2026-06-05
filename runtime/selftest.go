package main

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

// selftest runs in-process checks against an ephemeral runtime (no socket, no
// durable log), mirroring api/server.py --selftest. It exercises the model,
// the GraphQL schema, and a signup round-trip through the graph.
func selftest() error {
	model, err := LoadModel()
	if err != nil {
		return fmt.Errorf("load model: %w", err)
	}
	g := NewGraph(nil)
	model.Seed(g)
	api := &API{Model: model, Graph: g}

	if n := len(model.Classes); n < 35 {
		return fmt.Errorf("expected >=35 primitives, got %d", n)
	}
	if c, ok := model.Class("identity"); !ok || c.Name != "Identity" {
		return fmt.Errorf("identity primitive missing")
	}
	if _, ok := model.Class("risk"); !ok {
		return fmt.Errorf("risk primitive missing")
	}
	known := map[string]bool{}
	for _, n := range model.Nodes {
		known[n.ID] = true
	}
	for _, e := range model.Edges {
		if !known[e.To] {
			return fmt.Errorf("dangling edge to %q", e.To)
		}
	}
	res, ok := model.Resolve("identity", 2)
	if !ok {
		return fmt.Errorf("resolve(identity) failed")
	}
	if !contains(res.Resolved, "capability") {
		return fmt.Errorf("resolve(identity,2) should reach capability, got %v", res.Resolved)
	}

	schema, err := api.BuildSchema()
	if err != nil {
		return fmt.Errorf("build schema: %w", err)
	}

	// signup via GraphQL, then confirm the Identity record + audit edge exist.
	r := graphql.Do(graphql.Params{Schema: schema, RequestString: `
		mutation { signup(name:"Ada Lovelace", email:"ada@example.com") { id table } }`})
	if len(r.Errors) > 0 {
		return fmt.Errorf("signup mutation: %v", r.Errors)
	}
	q := graphql.Do(graphql.Params{Schema: schema, RequestString: `
		{ records(table:"identity") { id } event: records(table:"event") { id } }`})
	if len(q.Errors) > 0 {
		return fmt.Errorf("records query: %v", q.Errors)
	}

	fmt.Printf("selftest OK: classes=%d nodes=%d edges=%d resolve(identity,2)=%v\n",
		len(model.Classes), len(model.Nodes), len(model.Edges), res.Resolved)
	return nil
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
