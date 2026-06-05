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
