package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// API binds the canonical model and the runtime graph behind the GraphQL schema.
type API struct {
	Model *Model
	Graph *Graph
	// Keyring holds domain owners' ed25519 private keys, minted by
	// RegisterDomain. In-memory only (reference runtime; not persisted) — a
	// KMS/external signer is a later swap behind signFor.
	Keyring map[string]ed25519.PrivateKey
}

// jsonScalar carries arbitrary record fields / JSON Schema documents through
// GraphQL as a single opaque value, so we don't have to model every primitive's
// fields as a static GraphQL type.
var jsonScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "JSON",
	Description:  "Arbitrary JSON value (record fields, schema documents).",
	Serialize:    func(v any) any { return v },
	ParseValue:   func(v any) any { return v },
	ParseLiteral: parseJSONLiteral,
})

// parseJSONLiteral converts a GraphQL AST literal into a plain Go value so
// callers can pass nested objects/lists for the JSON scalar (e.g. record fields).
func parseJSONLiteral(v ast.Value) any {
	switch val := v.(type) {
	case *ast.StringValue:
		return val.Value
	case *ast.BooleanValue:
		return val.Value
	case *ast.IntValue:
		return val.Value
	case *ast.FloatValue:
		return val.Value
	case *ast.ObjectValue:
		out := map[string]any{}
		for _, f := range val.Fields {
			out[f.Name.Value] = parseJSONLiteral(f.Value)
		}
		return out
	case *ast.ListValue:
		out := make([]any, 0, len(val.Values))
		for _, item := range val.Values {
			out = append(out, parseJSONLiteral(item))
		}
		return out
	default:
		return nil
	}
}

// BuildSchema constructs the GraphQL schema exposing the data fabric: the
// self-describing model (classes/graph/resolve) and the live runtime
// (records/traverse) plus mutations (signup/createRecord/relate).
func (a *API) BuildSchema() (graphql.Schema, error) {
	classType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Class",
		Fields: graphql.Fields{
			"id":       &graphql.Field{Type: graphql.String},
			"name":     &graphql.Field{Type: graphql.String},
			"question": &graphql.Field{Type: graphql.String},
			"schema": &graphql.Field{
				Type:        jsonScalar,
				Description: "Generated JSON Schema for this primitive.",
				Resolve: func(p graphql.ResolveParams) (any, error) {
					c := p.Source.(Class)
					raw, err := a.Model.Schema(c.ID)
					if err != nil {
						return nil, nil
					}
					var out any
					_ = json.Unmarshal(raw, &out)
					return out, nil
				},
			},
		},
	})

	metaEdgeType := graphql.NewObject(graphql.ObjectConfig{
		Name: "MetaEdge",
		Fields: graphql.Fields{
			"from": &graphql.Field{Type: graphql.String},
			"rel":  &graphql.Field{Type: graphql.String},
			"to":   &graphql.Field{Type: graphql.String},
		},
	})
	metaNodeType := graphql.NewObject(graphql.ObjectConfig{
		Name: "MetaNode",
		Fields: graphql.Fields{
			"id":   &graphql.Field{Type: graphql.String},
			"name": &graphql.Field{Type: graphql.String},
		},
	})
	graphType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Graph",
		Fields: graphql.Fields{
			"nodes": &graphql.Field{Type: graphql.NewList(metaNodeType)},
			"edges": &graphql.Field{Type: graphql.NewList(metaEdgeType)},
		},
	})
	resolutionType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Resolution",
		Fields: graphql.Fields{
			"start":    &graphql.Field{Type: graphql.String},
			"depth":    &graphql.Field{Type: graphql.Int},
			"resolved": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"edges":    &graphql.Field{Type: graphql.NewList(metaEdgeType)},
		},
	})

	recordType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Record",
		Fields: graphql.Fields{
			"table":  &graphql.Field{Type: graphql.String},
			"id":     &graphql.Field{Type: graphql.String},
			"fields": &graphql.Field{Type: jsonScalar},
		},
	})
	edgeType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Edge",
		Fields: graphql.Fields{
			"rel":   &graphql.Field{Type: graphql.String},
			"from":  &graphql.Field{Type: graphql.String},
			"to":    &graphql.Field{Type: graphql.String},
			"props": &graphql.Field{Type: jsonScalar},
		},
	})
	traversalType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Traversal",
		Fields: graphql.Fields{
			"start":   &graphql.Field{Type: graphql.String},
			"depth":   &graphql.Field{Type: graphql.Int},
			"records": &graphql.Field{Type: graphql.NewList(recordType)},
			"edges":   &graphql.Field{Type: graphql.NewList(edgeType)},
		},
	})

	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"health": &graphql.Field{
				Type:    jsonScalar,
				Resolve: func(p graphql.ResolveParams) (any, error) { return a.health(), nil },
			},
			"classes": &graphql.Field{
				Type:        graphql.NewList(classType),
				Description: "All Fabric primitives.",
				Resolve:     func(p graphql.ResolveParams) (any, error) { return a.Model.Classes, nil },
			},
			"class": &graphql.Field{
				Type:        classType,
				Description: "One primitive by short id, e.g. \"identity\".",
				Args:        graphql.FieldConfigArgument{"name": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					c, ok := a.Model.Class(p.Args["name"].(string))
					if !ok {
						return nil, nil
					}
					return c, nil
				},
			},
			"graph": &graphql.Field{
				Type:        graphType,
				Description: "The schema graph: the model describing itself.",
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return map[string]any{"nodes": a.Model.Nodes, "edges": a.Model.Edges}, nil
				},
			},
			"resolve": &graphql.Field{
				Type:        resolutionType,
				Description: "Walk the schema graph from a primitive.",
				Args: graphql.FieldConfigArgument{
					"name":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"depth": &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 2},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					r, ok := a.Model.Resolve(p.Args["name"].(string), p.Args["depth"].(int))
					if !ok {
						return nil, nil
					}
					return r, nil
				},
			},
			"tables": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "Distinct record tables in the runtime store.",
				Resolve:     func(p graphql.ResolveParams) (any, error) { return a.Graph.Tables(), nil },
			},
			"records": &graphql.Field{
				Type:        graphql.NewList(recordType),
				Description: "All runtime records of a table.",
				Args:        graphql.FieldConfigArgument{"table": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return a.Graph.Table(p.Args["table"].(string)), nil
				},
			},
			"record": &graphql.Field{
				Type:        recordType,
				Description: "One runtime record by id.",
				Args:        graphql.FieldConfigArgument{"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return a.Graph.Get(p.Args["id"].(string)), nil
				},
			},
			"traverse": &graphql.Field{
				Type:        traversalType,
				Description: "Walk the runtime property graph from a record.",
				Args: graphql.FieldConfigArgument{
					"id":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"depth": &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 2},
					"dir":   &graphql.ArgumentConfig{Type: graphql.String, DefaultValue: "both"},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return a.Graph.Traverse(p.Args["id"].(string), p.Args["depth"].(int), p.Args["dir"].(string)), nil
				},
			},
			"twinDecide": &graphql.Field{
				Type:        jsonScalar,
				Description: "Ask a twin what it would decide among options — deterministic, justified by its preferences.",
				Args: graphql.FieldConfigArgument{
					"twin":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"options": &graphql.ArgumentConfig{Type: jsonScalar},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return a.TwinDecide(p.Args["twin"].(string), toMapList(p.Args["options"]))
				},
			},
			"verifySignature": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Verify an ed25519 signature (hex) over a payload for a public key (hex).",
				Args: graphql.FieldConfigArgument{
					"pubkey":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"payload":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"signature": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return VerifySignature(p.Args["pubkey"].(string), []byte(p.Args["payload"].(string)), p.Args["signature"].(string)), nil
				},
			},
		},
	})

	mutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"signup": &graphql.Field{
				Type:        recordType,
				Description: "Register an alpha tester as an Identity record (+ audit Event edge).",
				Args: graphql.FieldConfigArgument{
					"name":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"email":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"company": &graphql.ArgumentConfig{Type: graphql.String},
					"message": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return a.signup(p.Args)
				},
			},
			"createRecord": &graphql.Field{
				Type:        recordType,
				Description: "Insert an arbitrary record into the multi-model store.",
				Args: graphql.FieldConfigArgument{
					"table":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"id":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"fields": &graphql.ArgumentConfig{Type: jsonScalar},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					fields, _ := p.Args["fields"].(map[string]any)
					return a.Graph.Put(&Record{
						Table:  p.Args["table"].(string),
						ID:     p.Args["id"].(string),
						Fields: fields,
					}, true), nil
				},
			},
			"registerAgent": &graphql.Field{
				Type:        recordType,
				Description: "Register an autonomous agent as a first-class governed actor.",
				Args: graphql.FieldConfigArgument{
					"id":           &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"name":         &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"capabilities": &graphql.ArgumentConfig{Type: graphql.NewList(graphql.String)},
					"objectives":   &graphql.ArgumentConfig{Type: graphql.NewList(graphql.String)},
					"policies":     &graphql.ArgumentConfig{Type: graphql.NewList(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return a.RegisterAgent(
						p.Args["id"].(string), p.Args["name"].(string),
						strList(p.Args["capabilities"]), strList(p.Args["objectives"]), strList(p.Args["policies"]),
					)
				},
			},
			"agentAct": &graphql.Field{
				Type:        recordType,
				Description: "Record an agent action as an audited Event (the agent's memory); optionally apply it.",
				Args: graphql.FieldConfigArgument{
					"agent":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"action":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"targetTable": &graphql.ArgumentConfig{Type: graphql.String},
					"targetId":    &graphql.ArgumentConfig{Type: graphql.String},
					"fields":      &graphql.ArgumentConfig{Type: jsonScalar},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					fields, _ := p.Args["fields"].(map[string]any)
					return a.Act(
						p.Args["agent"].(string), p.Args["action"].(string),
						toStr(p.Args["targetTable"]), toStr(p.Args["targetId"]), fields,
					)
				},
			},
			"relate": &graphql.Field{
				Type:        edgeType,
				Description: "Add a directed edge between two records.",
				Args: graphql.FieldConfigArgument{
					"from": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"rel":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"to":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return a.Graph.Relate(&Edge{
						From: p.Args["from"].(string),
						Rel:  p.Args["rel"].(string),
						To:   p.Args["to"].(string),
					}, true), nil
				},
			},
			"propose": &graphql.Field{
				Type:        recordType,
				Description: "Agent proposes a change into possibility (URAP) — recorded, NOT committed.",
				Args: graphql.FieldConfigArgument{
					"agent":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"action":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"targetTable": &graphql.ArgumentConfig{Type: graphql.String},
					"targetId":    &graphql.ArgumentConfig{Type: graphql.String},
					"fields":      &graphql.ArgumentConfig{Type: jsonScalar},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					fields, _ := p.Args["fields"].(map[string]any)
					return a.Propose(p.Args["agent"].(string), p.Args["action"].(string),
						toStr(p.Args["targetTable"]), toStr(p.Args["targetId"]), fields)
				},
			},
			"admit": &graphql.Field{
				Type:        recordType,
				Description: "Human admits a proposal at the now-edge — the only path that commits.",
				Args: graphql.FieldConfigArgument{
					"proposal": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"admitter": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return a.Admit(p.Args["proposal"].(string), p.Args["admitter"].(string))
				},
			},
			"reject": &graphql.Field{
				Type:        recordType,
				Description: "Human rejects a proposal — discarded, nothing committed.",
				Args: graphql.FieldConfigArgument{
					"proposal": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"rejecter": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return a.Reject(p.Args["proposal"].(string), toStr(p.Args["rejecter"]))
				},
			},
			"registerDomain": &graphql.Field{
				Type:        recordType,
				Description: "Register a governance domain with a platform owner; mints the owner's ed25519 signing key (pubkey on the record).",
				Args: graphql.FieldConfigArgument{
					"name":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"owner": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					rec, _, err := a.RegisterDomain(p.Args["name"].(string), p.Args["owner"].(string))
					return rec, err
				},
			},
			"verifyIdentity": &graphql.Field{
				Type:        recordType,
				Description: "Record an identity verification (oauth|sso|domain-control|key-challenge).",
				Args: graphql.FieldConfigArgument{
					"subject":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"method":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"evidence": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return a.VerifyIdentity(p.Args["subject"].(string), p.Args["method"].(string), toStr(p.Args["evidence"]))
				},
			},
			"twinPropose": &graphql.Field{
				Type:        recordType,
				Description: "Propose a universal Twin (profile aggregate + preference model) of any entity — URAP, NOT committed. Admit via `admit` (domain owner only, signed).",
				Args: graphql.FieldConfigArgument{
					"agent":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"sourceTable": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"sourceId":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"preferences": &graphql.ArgumentConfig{Type: jsonScalar},
					"domain":      &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					prefs, _ := p.Args["preferences"].(map[string]any)
					return a.TwinPropose(p.Args["agent"].(string), p.Args["sourceTable"].(string),
						p.Args["sourceId"].(string), prefs, toStr(p.Args["domain"]))
				},
			},
			"resolverScan": &graphql.Field{
				Type:        graphql.NewList(recordType),
				Description: "Resolver agent: deterministically PROPOSE merges for duplicate records (by keyField).",
				Args: graphql.FieldConfigArgument{
					"agent":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"table":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"keyField": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return a.ResolverScan(p.Args["agent"].(string), p.Args["table"].(string), p.Args["keyField"].(string))
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: query, Mutation: mutation})
}

func (a *API) health() map[string]any {
	s := a.Graph.Stats()
	s["status"] = "ok"
	s["classes"] = len(a.Model.Classes)
	return s
}

// signup mirrors api/server.py: Connect→Catalog→Govern→Activate. It mints a
// deterministic Identity id, stores it as a record, and links an immutable
// Event so the action is auditable in the graph.
func (a *API) signup(args map[string]any) (*Record, error) {
	name := strings.TrimSpace(toStr(args["name"]))
	email := strings.ToLower(strings.TrimSpace(toStr(args["email"])))
	at := strings.LastIndex(email, "@")
	if name == "" || at < 1 || !strings.Contains(email[at:], ".") {
		return nil, fmt.Errorf("valid name and email are required")
	}
	sum := sha256.Sum256([]byte(email))
	uid := "did:fabric:user:" + hex.EncodeToString(sum[:])[:16]
	now := time.Now().UTC().Format(time.RFC3339)

	identity := a.Graph.Put(&Record{
		Table: "identity",
		ID:    uid,
		Fields: map[string]any{
			"kind":         "person",
			"displayName":  name,
			"email":        email,
			"company":      args["company"],
			"useCase":      args["message"],
			"registeredAt": now,
		},
	}, true)

	eventID := "event:identity.signup:" + uid[len(uid)-8:]
	a.Graph.Put(&Record{
		Table: "event",
		ID:    eventID,
		Fields: map[string]any{
			"type":       "identity.signup",
			"actor":      uid,
			"occurredAt": now,
		},
	}, true)
	a.Graph.Relate(&Edge{From: uid, Rel: "performed", To: eventID}, true)
	// Tie the instance back to its primitive so the graph stays self-describing.
	a.Graph.Relate(&Edge{From: uid, Rel: "instanceOf", To: "primitive:identity"}, true)
	return identity, nil
}

func toStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// strList coerces a GraphQL list argument into []string, tolerating nil.
func strList(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		if s, ok := it.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
