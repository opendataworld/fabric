package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// A Twin is a universal, passive mirror of ANY entity in the fabric — a person,
// org, device, dataset, capability, any record. It is a profile aggregate (the
// connected subgraph around its source) plus a preference model that answers
// "what would this entity decide?" deterministically, with justification. A twin
// is never an autonomous actor: it is built by PROPOSING (URAP) and only the
// platform owner of its domain may ADMIT it, signing the now-edge. The Resolver
// canonicalizes identities; the Twin aggregates the canonical entity.

// TwinPropose aggregates the graph around any source entity and PROPOSES a twin
// (it commits nothing — admit does). domain may be passed explicitly; otherwise
// it is derived from the source record (e.g. an identity's email domain).
func (a *API) TwinPropose(agentID, sourceTable, sourceID string, preferences map[string]any, domain string) (*Record, error) {
	agentID = strings.TrimSpace(agentID)
	sourceTable = strings.TrimSpace(sourceTable)
	sourceID = strings.TrimSpace(sourceID)
	if sourceTable == "" || sourceID == "" {
		return nil, fmt.Errorf("sourceTable and sourceId are required")
	}
	src := a.Graph.Get(sourceID)
	if strings.TrimSpace(domain) == "" {
		domain = a.domainOf(src)
	}
	// Aggregate: the connected subgraph around the source (the profile mirror).
	tr := a.Graph.Traverse(sourceID, 2, "both")
	aggregate := make([]any, 0, len(tr.Records))
	for _, r := range tr.Records {
		if r.ID != sourceID {
			aggregate = append(aggregate, r.ID)
		}
	}
	twinID := "twin:" + sourceTable + ":" + sourceID
	profile := map[string]any{
		"source":      map[string]any{"table": sourceTable, "id": sourceID},
		"domain":      domain,
		"verified":    a.isVerified(sourceID),
		"aggregate":   aggregate,
		"edgeCount":   len(tr.Edges),
		"preferences": preferences,
		"builtAt":     time.Now().UTC().Format(time.RFC3339),
	}
	return a.Propose(agentID, "twin.build", "twin", twinID, profile)
}

// admitTwin is the twin-specific commit invoked by Admit for action
// "twin.build": it enforces domain-owner governance + identity verification,
// writes the twin record and its edges, and signs the now-edge. Returns the
// signature, the owner's public key, and the canonical signed payload (all
// stored on the admit Event for later verification).
func (a *API) admitTwin(prop *Record, admitter, twinTable, twinID string, f map[string]any) (sig, pub, payload string, err error) {
	domain, _ := f["domain"].(string)
	if strings.TrimSpace(domain) == "" {
		return "", "", "", fmt.Errorf("twin proposal has no domain — cannot determine the owning authority")
	}
	dom := a.Graph.Get("domain:" + domain)
	if dom == nil || dom.Table != "domain" {
		return "", "", "", fmt.Errorf("unknown domain %q — register it first", domain)
	}
	owner, _ := dom.Fields["owner"].(string)
	if admitter != owner {
		return "", "", "", fmt.Errorf("only the platform owner of domain %q (%s) may admit this twin", domain, owner)
	}
	if !a.isVerified(owner) {
		return "", "", "", fmt.Errorf("domain owner %q is not identity-verified", owner)
	}
	// Commit the twin and wire its edges.
	a.Graph.Put(&Record{Table: twinTable, ID: twinID, Fields: f}, true)
	if src, ok := f["source"].(map[string]any); ok {
		if sid, _ := src["id"].(string); sid != "" {
			a.Graph.Relate(&Edge{From: twinID, Rel: "mirrors", To: sid}, true)
		}
	}
	a.Graph.Relate(&Edge{From: twinID, Rel: "governedBy", To: dom.ID}, true)
	// Sign the now-edge commitment.
	payloadBytes := canonicalPayload(map[string]any{
		"proposal": prop.ID,
		"twin":     twinID,
		"domain":   domain,
		"source":   f["source"],
		"hash":     hashFields(f),
	})
	sig, pub, err = a.signFor(owner, payloadBytes)
	if err != nil {
		return "", "", "", err
	}
	return sig, pub, string(payloadBytes), nil
}

// TwinDecide answers "what would this entity decide?" deterministically: it
// scores each option by the twin's preference weights (score = Σ weight·attr),
// returning a ranked list, the chosen option (argmax; ties broken by id), and a
// justification (the winning preference contributions, largest first).
func (a *API) TwinDecide(twinID string, options []map[string]any) (map[string]any, error) {
	tw := a.Graph.Get(strings.TrimSpace(twinID))
	if tw == nil || tw.Table != "twin" {
		return nil, fmt.Errorf("unknown twin %q", twinID)
	}
	prefs := toFloatMap(tw.Fields["preferences"])
	type scored struct {
		id    string
		score float64
		terms map[string]float64
	}
	results := make([]scored, 0, len(options))
	for _, opt := range options {
		id := toStr(opt["id"])
		attrs := toFloatMap(opt["attrs"])
		terms := map[string]float64{}
		var score float64
		for k, w := range prefs {
			if av, ok := attrs[k]; ok {
				c := w * av
				score += c
				terms[k] = c
			}
		}
		results = append(results, scored{id, score, terms})
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		return results[i].id < results[j].id
	})
	ranked := make([]any, 0, len(results))
	for _, r := range results {
		ranked = append(ranked, map[string]any{"id": r.id, "score": r.score, "terms": floatMapToAny(r.terms)})
	}
	out := map[string]any{"twin": twinID, "ranked": ranked}
	if len(results) > 0 {
		out["choice"] = results[0].id
		out["score"] = results[0].score
		out["justification"] = topTerms(results[0].terms)
	}
	return out, nil
}

// --- numeric coercion helpers ---
// GraphQL passes numbers as strings through the JSON scalar; MCP/JSON pass
// float64; tests pass Go numerics. toFloat tolerates all three.

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f, err == nil
	}
	return 0, false
}

func toFloatMap(v any) map[string]float64 {
	out := map[string]float64{}
	m, ok := v.(map[string]any)
	if !ok {
		return out
	}
	for k, val := range m {
		if f, ok := toFloat(val); ok {
			out[k] = f
		}
	}
	return out
}

func floatMapToAny(m map[string]float64) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// topTerms renders a preference→contribution map as a list ordered by
// contribution (largest first), the human-readable justification for a choice.
func topTerms(terms map[string]float64) []any {
	type t struct {
		k string
		v float64
	}
	ts := make([]t, 0, len(terms))
	for k, v := range terms {
		ts = append(ts, t{k, v})
	}
	sort.Slice(ts, func(i, j int) bool {
		if ts[i].v != ts[j].v {
			return ts[i].v > ts[j].v
		}
		return ts[i].k < ts[j].k
	})
	out := make([]any, 0, len(ts))
	for _, x := range ts {
		out = append(out, map[string]any{"preference": x.k, "contribution": x.v})
	}
	return out
}

// toMapList coerces a JSON list argument into []map[string]any for TwinDecide.
func toMapList(v any) []map[string]any {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		if m, ok := it.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}
