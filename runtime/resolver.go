package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// URAP propose→admit. An agent PROPOSES a change into possibility (a proposal
// record); it never commits to reality. A human ADMITS the proposal — the only
// path that applies the change and records an immutable Event. The agent
// proposes; the human guards the now-edge. (See docs / ARCHITECTURE.)

// Propose records an agent's intended change as a proposal WITHOUT applying it.
// The target record is untouched until a human admits.
func (a *API) Propose(agentID, action, targetTable, targetID string, fields map[string]any) (*Record, error) {
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
	propID := fmt.Sprintf("proposal:%s:%d", shortID(agentID), time.Now().UnixNano())
	prop := a.Graph.Put(&Record{
		Table: "proposal",
		ID:    propID,
		Fields: map[string]any{
			"status":      "proposed",
			"actor":       agentID,
			"action":      action,
			"targetTable": targetTable,
			"targetId":    targetID,
			"fields":      fields,
			"governedBy":  agent.Fields["policies"],
			"proposedAt":  now,
		},
	}, true)
	a.Graph.Relate(&Edge{From: agentID, Rel: "proposed", To: propID}, true)
	return prop, nil
}

// Admit is the human's act at the now-edge: it applies the proposal's change
// (commit forward) and records an immutable Event. Only admit mutates reality.
func (a *API) Admit(proposalID, admitter string) (*Record, error) {
	proposalID = strings.TrimSpace(proposalID)
	admitter = strings.TrimSpace(admitter)
	if admitter == "" {
		return nil, fmt.Errorf("admitter (the human guard) is required")
	}
	prop := a.Graph.Get(proposalID)
	if prop == nil || prop.Table != "proposal" {
		return nil, fmt.Errorf("unknown proposal %q", proposalID)
	}
	if s, _ := prop.Fields["status"].(string); s != "proposed" {
		return nil, fmt.Errorf("proposal %q already %s", proposalID, s)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	action, _ := prop.Fields["action"].(string)
	tt, _ := prop.Fields["targetTable"].(string)
	tid, _ := prop.Fields["targetId"].(string)
	f, _ := prop.Fields["fields"].(map[string]any)

	// eventExtra carries action-specific fields onto the immutable admit Event
	// (e.g. the ed25519 signature material for a signed twin commitment).
	eventExtra := map[string]any{}

	// Commit forward: apply the proposed change to reality.
	switch action {
	case "resolve.merge":
		// Relate each duplicate to the canonical record; mark it merged.
		// (Canonical record is preserved, not overwritten.)
		if dups, ok := f["duplicates"].([]any); ok {
			for _, d := range dups {
				ds := fmt.Sprintf("%v", d)
				a.Graph.Relate(&Edge{From: ds, Rel: "canonicalizedAs", To: tid}, true)
				if dr := a.Graph.Get(ds); dr != nil {
					dr.Fields["mergedInto"] = tid
					dr.Fields["status"] = "merged"
					a.Graph.Put(dr, true)
				}
			}
		}
	case "twin.build":
		// Twin governance: only the verified platform owner of the twin's
		// domain may admit, and the commitment is ed25519-signed.
		sig, pub, payload, err := a.admitTwin(prop, admitter, tt, tid, f)
		if err != nil {
			return nil, err
		}
		eventExtra["twin"] = tid
		eventExtra["signature"] = sig
		eventExtra["pubkey"] = pub
		eventExtra["payload"] = payload
	default:
		if tt != "" && tid != "" {
			a.Graph.Put(&Record{Table: tt, ID: tid, Fields: f}, true)
			actor, _ := prop.Fields["actor"].(string)
			a.Graph.Relate(&Edge{From: actor, Rel: action, To: tid}, true)
		}
	}

	eventID := fmt.Sprintf("event:proposal.admitted:%s:%d", shortID(proposalID), time.Now().UnixNano())
	eventFields := map[string]any{
		"type":       "proposal.admitted",
		"actor":      admitter,
		"action":     action,
		"proposal":   proposalID,
		"occurredAt": now,
	}
	for k, v := range eventExtra {
		eventFields[k] = v
	}
	event := a.Graph.Put(&Record{Table: "event", ID: eventID, Fields: eventFields}, true)
	prop.Fields["status"] = "admitted"
	prop.Fields["admittedBy"] = admitter
	a.Graph.Put(prop, true)
	a.Graph.Relate(&Edge{From: proposalID, Rel: "admittedBy", To: admitter}, true)
	return event, nil
}

// Reject discards a proposal without applying it, recording an Event.
func (a *API) Reject(proposalID, rejecter string) (*Record, error) {
	prop := a.Graph.Get(strings.TrimSpace(proposalID))
	if prop == nil || prop.Table != "proposal" {
		return nil, fmt.Errorf("unknown proposal %q", proposalID)
	}
	if s, _ := prop.Fields["status"].(string); s != "proposed" {
		return nil, fmt.Errorf("proposal %q already %s", proposalID, s)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	eventID := fmt.Sprintf("event:proposal.rejected:%s:%d", shortID(proposalID), time.Now().UnixNano())
	event := a.Graph.Put(&Record{Table: "event", ID: eventID, Fields: map[string]any{
		"type": "proposal.rejected", "actor": strings.TrimSpace(rejecter), "proposal": prop.ID, "occurredAt": now,
	}}, true)
	prop.Fields["status"] = "rejected"
	a.Graph.Put(prop, true)
	return event, nil
}

// ResolverScan is the Resolver agent's Capability: deterministically find
// duplicate records in a table (same value of keyField) and PROPOSE merging
// each duplicate into the canonical record (the lowest id in the group). It
// proposes only — nothing is committed. Deterministic: same input → same
// proposals (groups and ids are sorted).
func (a *API) ResolverScan(agentID, table, keyField string) ([]*Record, error) {
	table = strings.TrimSpace(table)
	keyField = strings.TrimSpace(keyField)
	if table == "" || keyField == "" {
		return nil, fmt.Errorf("table and keyField are required")
	}
	if r := a.Graph.Get(strings.TrimSpace(agentID)); r == nil || r.Table != "agent" {
		return nil, fmt.Errorf("unknown agent %q — register it first", agentID)
	}
	groups := map[string][]string{}
	for _, r := range a.Graph.Table(table) {
		v, ok := r.Fields[keyField]
		if !ok {
			continue
		}
		if s, _ := r.Fields["status"].(string); s == "merged" {
			continue // already resolved
		}
		k := fmt.Sprintf("%v", v)
		groups[k] = append(groups[k], r.ID)
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var proposals []*Record
	for _, k := range keys {
		ids := groups[k]
		if len(ids) < 2 {
			continue
		}
		sort.Strings(ids)
		canonical, dups := ids[0], ids[1:]
		prop, err := a.Propose(agentID, "resolve.merge", table, canonical, map[string]any{
			"key":        keyField + "=" + k,
			"canonical":  canonical,
			"duplicates": toAnySlice(dups),
		})
		if err != nil {
			return nil, err
		}
		proposals = append(proposals, prop)
	}
	return proposals, nil
}
