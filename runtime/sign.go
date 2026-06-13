package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Signing makes the now-edge non-repudiable. When the platform owner of a domain
// admits a twin proposal, the runtime SIGNS the canonical commitment with the
// owner's ed25519 key; the admit Event stores the signature, the owner's public
// key, and the signed payload so anyone can verify the twin state was admitted
// by that owner and not tampered with afterwards. All crypto is in-runtime and
// self-contained (Go stdlib) — no external dependencies.

// canonicalPayload renders fields as deterministic JSON. Go's json.Marshal sorts
// map keys at every level, so the same fields always yield the same bytes — the
// exact bytes that get signed.
func canonicalPayload(fields map[string]any) []byte {
	b, _ := json.Marshal(fields)
	return b
}

// hashFields returns a hex SHA-256 over the canonical encoding of fields, so a
// signature can commit to full twin state compactly and tamper-evidently.
func hashFields(fields map[string]any) string {
	sum := sha256.Sum256(canonicalPayload(fields))
	return hex.EncodeToString(sum[:])
}

// signFor signs payload with the private key held for ownerID, returning the
// signature and the owner's public key (both hex). Keys are minted by
// RegisterDomain.
func (a *API) signFor(ownerID string, payload []byte) (sigHex, pubHex string, err error) {
	priv, ok := a.Keyring[ownerID]
	if !ok {
		return "", "", fmt.Errorf("owner %q has no signing key — register its domain first", ownerID)
	}
	sig := ed25519.Sign(priv, payload)
	pub := priv.Public().(ed25519.PublicKey)
	return hex.EncodeToString(sig), hex.EncodeToString(pub), nil
}

// VerifySignature reports whether sigHex is a valid ed25519 signature over
// payload for the public key pubHex. Pure and exported for the verify
// endpoint/tool and tests.
func VerifySignature(pubHex string, payload []byte, sigHex string) bool {
	pub, err := hex.DecodeString(pubHex)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return false
	}
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}
	return ed25519.Verify(ed25519.PublicKey(pub), payload, sig)
}
