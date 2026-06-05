#!/usr/bin/env bash
#
# One-command install of the Kubernetes Dashboard control panel on k3s.
#
#   ./install.sh                      # secure default: port-forward access only
#   ./install.sh dashboard.you.com    # ALSO expose at a URL (TLS + basic-auth)
#
# With no args you get port-forward access (recommended). Passing a hostname
# additionally applies 20-ingress.example.yaml to expose it publicly behind
# Traefik TLS + a basic-auth gate — you must still set the basic-auth secret.
set -euo pipefail

HOST="${1:-}"
command -v kubectl >/dev/null || { echo "error: kubectl not found" >&2; exit 1; }
SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ">> applying Kubernetes Dashboard ..."
kubectl apply -k "$SRC"

if [[ -n "$HOST" ]]; then
  echo ">> exposing at https://${HOST} (TLS + basic-auth) ..."
  if kubectl -n kubernetes-dashboard get secret dashboard-basic-auth >/dev/null 2>&1; then
    sed "s|dashboard.example.com|${HOST}|g" "$SRC/20-ingress.example.yaml" | kubectl apply -f -
  else
    sed -e "s|dashboard.example.com|${HOST}|g" \
        -e '/^  users:/d' "$SRC/20-ingress.example.yaml" | kubectl apply -f -
    echo "   !! set basic-auth before relying on this:" >&2
    echo "      htpasswd -nb admin 'PASSWORD' | base64 -w0   # then put under data.users" >&2
    echo "      kubectl -n kubernetes-dashboard edit secret dashboard-basic-auth" >&2
  fi
fi

echo ">> waiting for the dashboard to become ready ..."
kubectl -n kubernetes-dashboard rollout status deploy/kubernetes-dashboard --timeout=120s || true

# Give the token controller a moment to populate the Secret.
TOKEN=""
for _ in $(seq 1 10); do
  TOKEN="$(kubectl -n kubernetes-dashboard get secret admin-user-token -o jsonpath='{.data.token}' 2>/dev/null | base64 -d || true)"
  [[ -n "$TOKEN" ]] && break
  sleep 2
done
[[ -z "$TOKEN" ]] && TOKEN="$(kubectl -n kubernetes-dashboard create token admin-user 2>/dev/null || echo '<run: kubectl -n kubernetes-dashboard create token admin-user>')"

cat <<EOF

============================================================
 Kubernetes Dashboard is deployed.

 1) Start a port-forward (keep it running):

      kubectl -n kubernetes-dashboard port-forward svc/kubernetes-dashboard 8443:443

    Accessing from your laptop over SSH? Tunnel it too:

      ssh -L 8443:localhost:8443 <user>@<k3s-host>

 2) Open  https://localhost:8443  (accept the self-signed cert).

 3) Choose "Token" and paste:

$TOKEN

 NOTE: this token is cluster-admin — treat it like a root password.
============================================================
EOF
