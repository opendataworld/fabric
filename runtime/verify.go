package main

import (
	"fmt"
	"strings"
	"time"
)

// Identity verification gates the twin layer: both the entity being twinned and
// the domain owner who admits must hold a verified identity. A verification is a
// runtime record asserting that a subject was verified by some method — the
// foothold for richer linked-account resolution (the identity-fabric) later.

var verifyMethods = map[string]bool{
	"oauth": true, "sso": true, "domain-control": true, "key-challenge": true,
}

// VerifyIdentity records a verification for a subject identity. method must be a
// known verification method (oauth|sso|domain-control|key-challenge).
func (a *API) VerifyIdentity(subject, method, evidence string) (*Record, error) {
	subject = strings.TrimSpace(subject)
	method = strings.TrimSpace(method)
	if subject == "" {
		return nil, fmt.Errorf("subject identity is required")
	}
	if !verifyMethods[method] {
		return nil, fmt.Errorf("unknown verification method %q (want oauth|sso|domain-control|key-challenge)", method)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	id := fmt.Sprintf("verification:%s:%d", shortID(subject), time.Now().UnixNano())
	rec := a.Graph.Put(&Record{
		Table: "verification",
		ID:    id,
		Fields: map[string]any{
			"subject":    subject,
			"method":     method,
			"status":     "verified",
			"evidence":   evidence,
			"verifiedAt": now,
		},
	}, true)
	a.Graph.Relate(&Edge{From: subject, Rel: "verifiedBy", To: id}, true)
	return rec, nil
}

// isVerified reports whether subject has at least one verified verification.
func (a *API) isVerified(subject string) bool {
	for _, r := range a.Graph.Table("verification") {
		if s, _ := r.Fields["subject"].(string); s != subject {
			continue
		}
		if st, _ := r.Fields["status"].(string); st == "verified" {
			return true
		}
	}
	return false
}
