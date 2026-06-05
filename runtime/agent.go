package main

import (
	"fmt"
	"strings"
	"time"
)

// Agent support makes the data fabric agent-native: autonomous actors are
// first-class, governed records in the same multi-model graph as everything
// else. This mirrors the canonical Agent primitive (agents/agent-model.yaml):
// an Agent extends Identity and executes Capabilities, pursues Objectives, is
// governedBy Policies, and accrues Events as memory.

// RegisterAgent creates (or replaces) an agent record and wires its governance
// edges. Capability/objective/policy ids are linked with the canonical
// relationship names so the runtime graph matches the schema graph.
func (a *API) RegisterAgent(id, name string, capabilities, objectives, policies []string) (*Record, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	if id == "" || name == "" {
		return nil, fmt.Errorf("agent id and name are required")
	}
	rec := a.Graph.Put(&Record{
		Table: "agent",
		ID:    id,
		Fields: map[string]any{
			"name":         name,
			"capabilities": toAnySlice(capabilities),
			"objectives":   toAnySlice(objectives),
			"policies":     toAnySlice(policies),
			"registeredAt": time.Now().UTC().Format(time.RFC3339),
		},
	}, true)

	// Self-describing: tie the instance to its primitive.
	a.Graph.Relate(&Edge{From: id, Rel: "instanceOf", To: "primitive:agent"}, true)
	for _, c := range capabilities {
		a.Graph.Relate(&Edge{From: id, Rel: "executes", To: c}, true)
	}
	for _, o := range objectives {
		a.Graph.Relate(&Edge{From: id, Rel: "pursues", To: o}, true)
	}
	for _, p := range policies {
		a.Graph.Relate(&Edge{From: id, Rel: "governedBy", To: p}, true)
	}
	return rec, nil
}

// Act records an agent action as an audited Event (the agent's memory) and,
// optionally, applies a change to the graph (create a target record, relate to
// it). Governance is honest: the agent must be registered, and the action Event
// carries the policies the agent is governedBy — we record the governance
// context rather than pretending to evaluate arbitrary policy logic.
func (a *API) Act(agentID, action, targetTable, targetID string, fields map[string]any) (*Record, error) {
	agentID = strings.TrimSpace(agentID)
	action = strings.TrimSpace(action)
	if action == "" {
		return nil, fmt.Errorf("action is required")
	}
	agent := a.Graph.Get(agentID)
	if agent == nil || agent.Table != "agent" {
		return nil, fmt.Errorf("unknown agent %q — register it first", agentID)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Optionally apply the effect of the action.
	if targetTable != "" && targetID != "" {
		a.Graph.Put(&Record{Table: targetTable, ID: targetID, Fields: fields}, true)
		a.Graph.Relate(&Edge{From: agentID, Rel: action, To: targetID}, true)
	}

	// Audit Event = the agent's memory of what it did, under which governance.
	eventID := fmt.Sprintf("event:agent.act:%s:%d", shortID(agentID), time.Now().UnixNano())
	event := a.Graph.Put(&Record{
		Table: "event",
		ID:    eventID,
		Fields: map[string]any{
			"type":       "agent.act",
			"actor":      agentID,
			"action":     action,
			"target":     targetID,
			"governedBy": agent.Fields["policies"],
			"occurredAt": now,
		},
	}, true)
	a.Graph.Relate(&Edge{From: agentID, Rel: "hasMemory", To: eventID}, true)
	return event, nil
}

func shortID(id string) string {
	if i := strings.LastIndex(id, ":"); i >= 0 && i < len(id)-1 {
		return id[i+1:]
	}
	return id
}

func toAnySlice(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}
