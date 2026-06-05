#!/usr/bin/env bash
#
# One-command Nextcloud-on-k3s installer.
#
#   ./install.sh <hostname> [letsencrypt-email]
#
# Examples:
#   ./install.sh cloud.example.org                 # deploy, no auto-TLS
#   ./install.sh cloud.example.org me@example.org  # deploy + Let's Encrypt
#
# It builds an ephemeral, fully-resolved copy of the manifests in a temp dir
# (your tracked files are never modified), generates strong random secrets,
# substitutes the hostname, and applies everything with `kubectl apply -k`.
# Re-running is safe: existing secrets in the cluster are preserved.
#
# Overridable via env: NEXTCLOUD_ADMIN_USER, NEXTCLOUD_ADMIN_PASSWORD.
set -euo pipefail

HOST="${1:-${NEXTCLOUD_HOST:-}}"
EMAIL="${2:-${LETSENCRYPT_EMAIL:-}}"
if [[ -z "$HOST" ]]; then
  echo "usage: $0 <hostname> [letsencrypt-email]" >&2
  exit 1
fi

command -v kubectl >/dev/null || { echo "error: kubectl not found" >&2; exit 1; }
command -v openssl >/dev/null || { echo "error: openssl not found" >&2; exit 1; }

SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cp "$SRC"/*.yaml "$WORK"/

gen() { openssl rand -base64 24 | tr -dc 'A-Za-z0-9' | head -c 32; }
ADMIN_USER="${NEXTCLOUD_ADMIN_USER:-admin}"
ADMIN_PASS="${NEXTCLOUD_ADMIN_PASSWORD:-$(gen)}"

# Real secrets (referenced by kustomization.yaml as 01-secrets.yaml).
sed -e "s|CHANGE_ME_db_password|$(gen)|" \
    -e "s|CHANGE_ME_redis_password|$(gen)|" \
    -e "s|CHANGE_ME_admin_password|${ADMIN_PASS}|" \
    -e "s|NEXTCLOUD_ADMIN_USER: admin|NEXTCLOUD_ADMIN_USER: ${ADMIN_USER}|" \
    "$WORK/01-secrets.example.yaml" > "$WORK/01-secrets.yaml"

# Substitute the public hostname everywhere it appears.
sed -i "s|nextcloud.example.com|${HOST}|g" "$WORK"/*.yaml

echo ">> applying manifests to namespace 'nextcloud' ..."
kubectl apply -k "$WORK"

if [[ -n "$EMAIL" ]]; then
  echo ">> configuring Let's Encrypt issuer for ${EMAIL} ..."
  sed "s|CHANGE_ME@example.com|${EMAIL}|" \
    "$WORK/99-cert-manager.example.yaml" | kubectl apply -f -
else
  echo ">> no email supplied — skipping cert-manager issuer (TLS won't auto-issue)"
fi

echo ">> waiting for Nextcloud to become ready (first boot can take a few minutes) ..."
kubectl -n nextcloud rollout status deploy/nextcloud --timeout=600s || true

cat <<EOF

============================================================
 Nextcloud deployed.
   URL:      https://${HOST}
   Admin:    ${ADMIN_USER}
EOF
if [[ -z "${NEXTCLOUD_ADMIN_PASSWORD:-}" ]]; then
  echo "   Password: ${ADMIN_PASS}   <-- generated, save it now"
fi
cat <<EOF
============================================================
EOF
