#!/usr/bin/env bash
#
# One-command deploy for the k3s Stack Catalog panel.
#
#   ./install.sh <hostname> [letsencrypt-email]
#
# Builds an ephemeral, hostname-resolved copy of the manifests (tracked files
# untouched) and applies it with `kubectl apply -k`.
set -euo pipefail

HOST="${1:-${CATALOG_HOST:-}}"
EMAIL="${2:-${LETSENCRYPT_EMAIL:-}}"
if [[ -z "$HOST" ]]; then
  echo "usage: $0 <hostname> [letsencrypt-email]" >&2
  exit 1
fi
command -v kubectl >/dev/null || { echo "error: kubectl not found" >&2; exit 1; }

SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cp -r "$SRC"/. "$WORK"/

sed -i "s|catalog.example.com|${HOST}|g" "$WORK"/*.yaml
[[ -z "$EMAIL" ]] && sed -i '/cert-manager.io\/cluster-issuer/d' "$WORK/10-panel.yaml"

echo ">> applying stack-catalog ..."
kubectl apply -k "$WORK"
kubectl -n stack-catalog rollout status deploy/stack-catalog --timeout=120s || true

echo
echo "Catalog deployed: https://${HOST}"
