#!/usr/bin/env bash
#
# One-command Headlamp (CNCF Kubernetes UI) on k3s.
#
#   ./install.sh                # deploy; access via port-forward
#
# Prints a login token and the port-forward command. Access is via
# port-forward by design (don't expose a cluster-admin UI publicly).
set -euo pipefail

command -v kubectl >/dev/null || { echo "error: kubectl not found" >&2; exit 1; }
SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ">> applying Headlamp ..."
kubectl apply -k "$SRC"
kubectl -n headlamp rollout status deploy/headlamp --timeout=120s || true

TOKEN="$(kubectl -n headlamp create token headlamp-admin --duration=8760h 2>/dev/null || true)"

cat <<EOF

============================================================
 Headlamp deployed.

 Access (keep the port-forward running):
   kubectl -n headlamp port-forward svc/headlamp 8080:80
   # remote:  ssh -L 8080:localhost:8080 <user>@<k3s-host>
 then open  http://localhost:8080

 Login token:
   ${TOKEN:-"(run: kubectl -n headlamp create token headlamp-admin)"}
============================================================
EOF
