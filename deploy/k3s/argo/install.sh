#!/usr/bin/env bash
#
# One-command install of Argo Workflows + the agent-composer template on k3s.
#
#   ./install.sh
#
# Access is via port-forward (the quick-start server runs without login).
set -euo pipefail

command -v kubectl >/dev/null || { echo "error: kubectl not found" >&2; exit 1; }
SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ">> applying Argo Workflows ..."
kubectl apply -k "$SRC"

echo ">> waiting for the WorkflowTemplate CRD to be established ..."
kubectl wait --for=condition=Established --timeout=90s \
  crd/workflowtemplates.argoproj.io >/dev/null 2>&1 || true

echo ">> applying the agent-composer WorkflowTemplate ..."
kubectl apply -f "$SRC/agent-composer-template.yaml"

echo ">> waiting for Argo to become ready ..."
kubectl -n argo rollout status deploy/argo-server --timeout=120s || true

cat <<EOF

============================================================
 Argo Workflows is deployed.

 Open the UI (keep the port-forward running):

   kubectl -n argo port-forward svc/argo-server 2746:2746
   # remote? also:  ssh -L 2746:localhost:2746 <user>@<k3s-host>

 then browse  https://localhost:2746  (accept the self-signed cert).

 Run the sample agent pipeline:

   argo submit --from workflowtemplate/agent-composer -n argo --watch
   # or click "Submit" on agent-composer in the UI

 Edit agent-composer-template.yaml to replace the echo steps with real
 containers (model call, tools, retriever) and re-apply.
============================================================
EOF
