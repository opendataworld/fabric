package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// A domain is the unit of ownership and governance. Each domain has a platform
// owner (an Identity) and an ed25519 public key; the owner is the sole authority
// that may admit twin proposals for entities in that domain, and every such
// admission is signed. Like proposal/event, a domain is a runtime record table,
// not an ontology class.

// RegisterDomain mints an ed25519 keypair for the domain's owner, stores the
// private key in the in-memory keyring, and records the domain with the owner's
// public key. Returns the domain record and the owner's public key (hex, shown
// once for out-of-band distribution / verification).
func (a *API) RegisterDomain(name, ownerID string) (*Record, string, error) {
	name = strings.TrimSpace(name)
	ownerID = strings.TrimSpace(ownerID)
	if name == "" || ownerID == "" {
		return nil, "", fmt.Errorf("domain name and owner are required")
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("generate key: %w", err)
	}
	if a.Keyring == nil {
		a.Keyring = map[string]ed25519.PrivateKey{}
	}
	a.Keyring[ownerID] = priv
	pubHex := hex.EncodeToString(pub)

	id := "domain:" + name
	rec := a.Graph.Put(&Record{
		Table: "domain",
		ID:    id,
		Fields: map[string]any{
			"name":      name,
			"owner":     ownerID,
			"pubkey":    pubHex,
			"createdAt": time.Now().UTC().Format(time.RFC3339),
		},
	}, true)
	a.Graph.Relate(&Edge{From: ownerID, Rel: "owns", To: id}, true)
	return rec, pubHex, nil
}

// domainOf derives a record's governing domain: an explicit `domain` field wins,
// otherwise the domain part of an `email`, otherwise empty.
func (a *API) domainOf(rec *Record) string {
	if rec == nil {
		return ""
	}
	if d, ok := rec.Fields["domain"].(string); ok && strings.TrimSpace(d) != "" {
		return strings.TrimSpace(d)
	}
	if email, ok := rec.Fields["email"].(string); ok {
		if at := strings.LastIndex(email, "@"); at >= 0 && at < len(email)-1 {
			return email[at+1:]
		}
	}
	return ""
}
