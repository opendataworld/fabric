package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

// newTestMCP builds an MCP server whose output is captured in a buffer.
func newTestMCP(t *testing.T) (*MCPServer, *bytes.Buffer) {
	t.Helper()
	api, _ := newTestAPI(t)
	s := NewMCPServer(api)
	var buf bytes.Buffer
	s.out = json.NewEncoder(&buf)
	return s, &buf
}

func decode(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", buf.String(), err)
	}
	return resp
}

func TestMCPInitialize(t *testing.T) {
	s, buf := newTestMCP(t)
	s.dispatch(&rpcRequest{Method: "initialize", ID: json.RawMessage(`1`)})
	res := decode(t, buf)["result"].(map[string]any)
	if res["protocolVersion"] != mcpProtocolVersion {
		t.Fatalf("protocolVersion = %v", res["protocolVersion"])
	}
}

func TestMCPToolsListExposesAgentTools(t *testing.T) {
	s, buf := newTestMCP(t)
	s.dispatch(&rpcRequest{Method: "tools/list", ID: json.RawMessage(`2`)})
	tools := decode(t, buf)["result"].(map[string]any)["tools"].([]any)
	names := map[string]bool{}
	for _, tl := range tools {
		names[tl.(map[string]any)["name"].(string)] = true
	}
	for _, want := range []string{"fabric_classes", "fabric_resolve", "fabric_register_agent", "fabric_agent_act"} {
		if !names[want] {
			t.Fatalf("tools/list missing %q; got %v", want, names)
		}
	}
}

func TestMCPToolsCallResolve(t *testing.T) {
	s, buf := newTestMCP(t)
	params, _ := json.Marshal(map[string]any{"name": "fabric_resolve", "arguments": map[string]any{"name": "identity", "depth": 2}})
	s.dispatch(&rpcRequest{Method: "tools/call", ID: json.RawMessage(`3`), Params: params})
	result := decode(t, buf)["result"].(map[string]any)
	if result["isError"] == true {
		t.Fatalf("resolve call errored: %v", result)
	}
	text := result["content"].([]any)[0].(map[string]any)["text"].(string)
	if !bytes.Contains([]byte(text), []byte("capability")) {
		t.Fatalf("resolve result should mention capability: %s", text)
	}
}

func TestMCPNotificationNoReply(t *testing.T) {
	s, buf := newTestMCP(t)
	s.dispatch(&rpcRequest{Method: "notifications/initialized"}) // no id
	if buf.Len() != 0 {
		t.Fatalf("notification should not produce a reply, got %q", buf.String())
	}
}
