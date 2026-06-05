#!/usr/bin/env bash
#
# One-command deploy of the Zot registry on k3s.
#
#   ./install.sh <hostname> [letsencrypt-email]
#
# Generates a push credential (idempotent — reused on re-runs), sets the
# hostname, and applies everything. Prints the login + push instructions.
set -euo pipefail

HOST="${1:-${REGISTRY_HOST:-}}"
EMAIL="${2:-${LETSENCRYPT_EMAIL:-}}"
if [[ -z "$HOST" ]]; then
  echo "usage: $0 <hostname> [letsencrypt-email]" >&2
  exit 1
fi
command -v kubectl >/dev/null || { echo "error: kubectl not found" >&2; exit 1; }

SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cp "$SRC"/*.yaml "$WORK"/

PUSH_USER="${REGISTRY_USER:-pushuser}"
PUSH_PASS=""

# Idempotent credential: reuse the in-cluster secret if present.
if kubectl -n registry get secret zot-htpasswd >/dev/null 2>&1; then
  echo ">> existing registry credential found — reusing it"
  sed -i '/02-htpasswd.yaml/d' "$WORK/kustomization.yaml"
else
  command -v htpasswd >/dev/null || {
    echo "error: 'htpasswd' (apache2-utils) is required to generate a bcrypt credential." >&2
    echo "       install it, or create 02-htpasswd.yaml manually (see the example)." >&2
    exit 1
  }
  PUSH_PASS="${REGISTRY_PASSWORD:-$(LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 24)}"
  LINE="$(htpasswd -nbB "$PUSH_USER" "$PUSH_PASS")"
  sed "s|pushuser:\$2y\$05\$CHANGE_ME_replace_with_a_real_bcrypt_hash|${LINE}|" \
    "$WORK/02-htpasswd.example.yaml" > "$WORK/02-htpasswd.yaml"
fi

sed -i "s|registry.example.com|${HOST}|g" "$WORK"/*.yaml

echo ">> applying registry ..."
kubectl apply -k "$WORK"

if [[ -n "$EMAIL" ]]; then
  echo ">> (cert-manager issuer is expected to exist as 'letsencrypt-prod')"
fi

kubectl -n registry rollout status deploy/zot --timeout=120s || true

cat <<EOF

============================================================
 Zot registry deployed:  https://${HOST}
   Web UI:   https://${HOST}
   Push:     docker login ${HOST} -u ${PUSH_USER}
EOF
if [[ -n "$PUSH_PASS" ]]; then
  echo "             password: ${PUSH_PASS}   <-- generated, save it now"
fi
cat <<EOF
   Example:  docker tag myimage ${HOST}/myimage:tag && docker push ${HOST}/myimage:tag

 Anonymous pull is allowed; push requires the credential above.
 Signature verification (cosign/notation) is enabled — push a cosign
 signature and the UI will show the image as signed.
============================================================
EOF
