package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// MCP makes the data fabric agent-native at the protocol level: the runtime
// speaks the Model Context Protocol (JSON-RPC 2.0 over stdio), so any
// MCP-capable agent can discover and call the fabric's tools directly — query
// the model, traverse the graph, register agents, and record governed actions.
//
// Run with `runtime --mcp`. Messages are newline-delimited JSON on stdin/stdout
// (the MCP stdio transport); diagnostics go to stderr.

const mcpProtocolVersion = "2024-11-05"

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// mcpTool describes one callable tool: its JSON-Schema input and a handler that
// reuses the same API/Graph methods as the GraphQL resolvers.
type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
	handler     func(args map[string]any) (any, error)
}

// MCPServer serves the fabric over the MCP stdio transport.
type MCPServer struct {
	api   *API
	tools []mcpTool
	out   *json.Encoder
}

func obj(props map[string]any, required ...string) map[string]any {
	s := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func str(desc string) map[string]any { return map[string]any{"type": "string", "description": desc} }
func num(desc string) map[string]any { return map[string]any{"type": "integer", "description": desc} }
func list(desc string) map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": desc}
}

// NewMCPServer wires the tool catalogue over the runtime API.
func NewMCPServer(a *API) *MCPServer {
	s := &MCPServer{api: a, out: json.NewEncoder(os.Stdout)}
	s.tools = []mcpTool{
		{
			Name:        "fabric_classes",
			Description: "List all Fabric primitives (the canonical model's classes).",
			InputSchema: obj(map[string]any{}),
			handler:     func(map[string]any) (any, error) { return a.Model.Classes, nil },
		},
		{
			Name:        "fabric_resolve",
			Description: "Walk the schema graph from a primitive to a depth, returning the connected subgraph.",
			InputSchema: obj(map[string]any{"name": str("primitive short id, e.g. identity"), "depth": num("hops (default 2)")}, "name"),
			handler: func(args map[string]any) (any, error) {
				r, ok := a.Model.Resolve(argStr(args, "name"), argInt(args, "depth", 2))
				if !ok {
					return nil, fmt.Errorf("unknown primitive %q", argStr(args, "name"))
				}
				return r, nil
			},
		},
		{
			Name:        "fabric_graph",
			Description: "Return the whole schema graph: nodes and edges.",
			InputSchema: obj(map[string]any{}),
			handler: func(map[string]any) (any, error) {
				return map[string]any{"nodes": a.Model.Nodes, "edges": a.Model.Edges}, nil
			},
		},
		{
			Name:        "fabric_records",
			Description: "List runtime records of a table (record type).",
			InputSchema: obj(map[string]any{"table": str("table name, e.g. identity, agent, event")}, "table"),
			handler:     func(args map[string]any) (any, error) { return a.Graph.Table(argStr(args, "table")), nil },
		},
		{
			Name:        "fabric_get_record",
			Description: "Fetch one runtime record by id.",
			InputSchema: obj(map[string]any{"id": str("record id")}, "id"),
			handler:     func(args map[string]any) (any, error) { return a.Graph.Get(argStr(args, "id")), nil },
		},
		{
			Name:        "fabric_traverse",
			Description: "Walk the runtime property graph from a record.",
			InputSchema: obj(map[string]any{"id": str("start record id"), "depth": num("hops (default 2)"), "dir": str("out|in|both")}, "id"),
			handler: func(args map[string]any) (any, error) {
				dir := argStr(args, "dir")
				if dir == "" {
					dir = "both"
				}
				return a.Graph.Traverse(argStr(args, "id"), argInt(args, "depth", 2), dir), nil
			},
		},
		{
			Name:        "fabric_create_record",
			Description: "Insert a record into the multi-model store.",
			InputSchema: obj(map[string]any{"table": str("table"), "id": str("id"), "fields": map[string]any{"type": "object", "description": "arbitrary fields"}}, "table", "id"),
			handler: func(args map[string]any) (any, error) {
				fields, _ := args["fields"].(map[string]any)
				return a.Graph.Put(&Record{Table: argStr(args, "table"), ID: argStr(args, "id"), Fields: fields}, true), nil
			},
		},
		{
			Name:        "fabric_relate",
			Description: "Add a directed edge between two records.",
			InputSchema: obj(map[string]any{"from": str("source id"), "rel": str("relationship"), "to": str("target id")}, "from", "rel", "to"),
			handler: func(args map[string]any) (any, error) {
				return a.Graph.Relate(&Edge{From: argStr(args, "from"), Rel: argStr(args, "rel"), To: argStr(args, "to")}, true), nil
			},
		},
		{
			Name:        "fabric_register_agent",
			Description: "Register an autonomous agent as a first-class governed actor in the fabric.",
			InputSchema: obj(map[string]any{
				"id": str("agent id"), "name": str("display name"),
				"capabilities": list("capability ids the agent executes"),
				"objectives":   list("objective ids the agent pursues"),
				"policies":     list("policy ids governing the agent"),
			}, "id", "name"),
			handler: func(args map[string]any) (any, error) {
				return a.RegisterAgent(argStr(args, "id"), argStr(args, "name"),
					argStrList(args, "capabilities"), argStrList(args, "objectives"), argStrList(args, "policies"))
			},
		},
		{
			Name:        "fabric_agent_act",
			Description: "Record a governed agent action as an audited Event (the agent's memory); optionally apply it to the graph.",
			InputSchema: obj(map[string]any{
				"agent": str("acting agent id"), "action": str("verb, e.g. ingested, classified"),
				"targetTable": str("optional record table to create/affect"),
				"targetId":    str("optional record id"),
				"fields":      map[string]any{"type": "object", "description": "optional target fields"},
			}, "agent", "action"),
			handler: func(args map[string]any) (any, error) {
				fields, _ := args["fields"].(map[string]any)
				return a.Act(argStr(args, "agent"), argStr(args, "action"), argStr(args, "targetTable"), argStr(args, "targetId"), fields)
			},
		},
		{
			Name:        "fabric_propose",
			Description: "URAP: agent PROPOSES a change into possibility — recorded, NOT committed.",
			InputSchema: obj(map[string]any{
				"agent": str("acting agent id"), "action": str("verb, e.g. resolve.merge"),
				"targetTable": str("optional record table"), "targetId": str("optional record id"),
				"fields": map[string]any{"type": "object", "description": "optional proposed fields"},
			}, "agent", "action"),
			handler: func(args map[string]any) (any, error) {
				fields, _ := args["fields"].(map[string]any)
				return a.Propose(argStr(args, "agent"), argStr(args, "action"), argStr(args, "targetTable"), argStr(args, "targetId"), fields)
			},
		},
		{
			Name:        "fabric_admit",
			Description: "Human admits a proposal at the now-edge — the only path that commits (audited Event).",
			InputSchema: obj(map[string]any{"proposal": str("proposal id"), "admitter": str("human guard id")}, "proposal", "admitter"),
			handler: func(args map[string]any) (any, error) {
				return a.Admit(argStr(args, "proposal"), argStr(args, "admitter"))
			},
		},
		{
			Name:        "fabric_reject",
			Description: "Human rejects a proposal — discarded, nothing committed.",
			InputSchema: obj(map[string]any{"proposal": str("proposal id"), "rejecter": str("who rejected")}, "proposal"),
			handler: func(args map[string]any) (any, error) {
				return a.Reject(argStr(args, "proposal"), argStr(args, "rejecter"))
			},
		},
		{
			Name:        "fabric_resolver_scan",
			Description: "Resolver agent: deterministically PROPOSE merges for duplicate records (by keyField). Proposes only.",
			InputSchema: obj(map[string]any{"agent": str("resolver agent id"), "table": str("record table"), "keyField": str("field that identifies duplicates")}, "agent", "table", "keyField"),
			handler: func(args map[string]any) (any, error) {
				return a.ResolverScan(argStr(args, "agent"), argStr(args, "table"), argStr(args, "keyField"))
			},
		},
		{
			Name:        "fabric_register_domain",
			Description: "Register a governance domain with a platform owner; mints the owner's ed25519 signing key. Returns the domain record and the owner's public key.",
			InputSchema: obj(map[string]any{"name": str("domain name, e.g. acme.com"), "owner": str("owner identity id")}, "name", "owner"),
			handler: func(args map[string]any) (any, error) {
				rec, pub, err := a.RegisterDomain(argStr(args, "name"), argStr(args, "owner"))
				if err != nil {
					return nil, err
				}
				return map[string]any{"domain": rec, "pubkey": pub}, nil
			},
		},
		{
			Name:        "fabric_verify_identity",
			Description: "Record an identity verification (oauth|sso|domain-control|key-challenge) — gates twinning and twin admission.",
			InputSchema: obj(map[string]any{"subject": str("identity id"), "method": str("oauth|sso|domain-control|key-challenge"), "evidence": str("optional evidence")}, "subject", "method"),
			handler: func(args map[string]any) (any, error) {
				return a.VerifyIdentity(argStr(args, "subject"), argStr(args, "method"), argStr(args, "evidence"))
			},
		},
		{
			Name:        "fabric_twin_propose",
			Description: "Propose a universal Twin (profile aggregate + preference model) of ANY entity — URAP, NOT committed. Admit via fabric_admit (domain owner only, signed).",
			InputSchema: obj(map[string]any{
				"agent": str("proposing agent id"), "sourceTable": str("source record table"), "sourceId": str("source record id"),
				"preferences": map[string]any{"type": "object", "description": "preference weights, e.g. {\"sustainability\":2,\"price\":-1}"},
				"domain":      str("governing domain (optional; derived from the source if omitted)"),
			}, "agent", "sourceTable", "sourceId"),
			handler: func(args map[string]any) (any, error) {
				prefs, _ := args["preferences"].(map[string]any)
				return a.TwinPropose(argStr(args, "agent"), argStr(args, "sourceTable"), argStr(args, "sourceId"), prefs, argStr(args, "domain"))
			},
		},
		{
			Name:        "fabric_twin_decide",
			Description: "Ask a twin what it would decide among options — deterministic, justified by its preferences.",
			InputSchema: obj(map[string]any{
				"twin":    str("twin id"),
				"options": map[string]any{"type": "array", "items": map[string]any{"type": "object"}, "description": "options, each {id, attrs:{...}}"},
			}, "twin", "options"),
			handler: func(args map[string]any) (any, error) {
				return a.TwinDecide(argStr(args, "twin"), toMapList(args["options"]))
			},
		},
		{
			Name:        "fabric_verify_signature",
			Description: "Verify an ed25519 signature (hex) over a payload for a public key (hex).",
			InputSchema: obj(map[string]any{"pubkey": str("public key hex"), "payload": str("signed payload"), "signature": str("signature hex")}, "pubkey", "payload", "signature"),
			handler: func(args map[string]any) (any, error) {
				return VerifySignature(argStr(args, "pubkey"), []byte(argStr(args, "payload")), argStr(args, "signature")), nil
			},
		},
		{
			Name:        "fabric_signup",
			Description: "Register a person as an Identity record (+ audit Event).",
			InputSchema: obj(map[string]any{"name": str("full name"), "email": str("email"), "company": str("optional"), "message": str("optional use case")}, "name", "email"),
			handler:     func(args map[string]any) (any, error) { return a.signup(args) },
		},
	}
	return s
}

// Serve runs the stdio read/dispatch/write loop until EOF.
func (s *MCPServer) Serve() error {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}
		s.dispatch(&req)
	}
	if err := sc.Err(); err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (s *MCPServer) dispatch(req *rpcRequest) {
	switch req.Method {
	case "initialize":
		s.reply(req.ID, map[string]any{
			"protocolVersion": mcpProtocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "fabric-runtime", "version": "0.1.0"},
		})
	case "tools/list":
		defs := make([]map[string]any, len(s.tools))
		for i, t := range s.tools {
			defs[i] = map[string]any{"name": t.Name, "description": t.Description, "inputSchema": t.InputSchema}
		}
		s.reply(req.ID, map[string]any{"tools": defs})
	case "tools/call":
		s.callTool(req)
	case "ping":
		s.reply(req.ID, map[string]any{})
	default:
		// Notifications (no id) such as notifications/initialized need no reply.
		if len(req.ID) > 0 {
			s.replyErr(req.ID, -32601, "method not found: "+req.Method)
		}
	}
}

func (s *MCPServer) callTool(req *rpcRequest) {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	_ = json.Unmarshal(req.Params, &params)
	for _, t := range s.tools {
		if t.Name != params.Name {
			continue
		}
		result, err := t.handler(params.Arguments)
		if err != nil {
			s.reply(req.ID, map[string]any{
				"content": []map[string]any{{"type": "text", "text": "error: " + err.Error()}},
				"isError": true,
			})
			return
		}
		b, _ := json.MarshalIndent(result, "", "  ")
		s.reply(req.ID, map[string]any{
			"content": []map[string]any{{"type": "text", "text": string(b)}},
		})
		return
	}
	s.reply(req.ID, map[string]any{
		"content": []map[string]any{{"type": "text", "text": "unknown tool: " + params.Name}},
		"isError": true,
	})
}

func (s *MCPServer) reply(id json.RawMessage, result any) {
	if len(id) == 0 {
		return
	}
	s.out.Encode(rpcResponse{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *MCPServer) replyErr(id json.RawMessage, code int, msg string) {
	s.out.Encode(rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}})
}

func argStr(args map[string]any, k string) string {
	if v, ok := args[k].(string); ok {
		return v
	}
	return ""
}

func argInt(args map[string]any, k string, def int) int {
	switch v := args[k].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return def
}

func argStrList(args map[string]any, k string) []string {
	items, ok := args[k].([]any)
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
