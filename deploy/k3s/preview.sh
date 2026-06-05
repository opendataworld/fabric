#!/usr/bin/env bash
#
# Quick local preview of a deployed stack via kubectl port-forward.
# Run this ON the k3s host (over SSH), then open the printed SSH tunnel from
# your laptop — no Ingress, DNS, or TLS required.
#
#   ./preview.sh           # preview the catalog panel  (default)
#   ./preview.sh nextcloud # preview Nextcloud
#
set -euo pipefail

TARGET="${1:-panel}"
case "$TARGET" in
  panel)     NS=stack-catalog; SVC=stack-catalog; RPORT=80;  LPORT=8080 ;;
  nextcloud) NS=nextcloud;     SVC=nextcloud;     RPORT=80;  LPORT=8081 ;;
  *) echo "usage: $0 [panel|nextcloud]" >&2; exit 1 ;;
esac

command -v kubectl >/dev/null || { echo "error: kubectl not found" >&2; exit 1; }
if ! kubectl -n "$NS" get svc "$SVC" >/dev/null 2>&1; then
  echo "error: service $SVC not found in namespace $NS — deploy it first:" >&2
  echo "       kubectl apply -k $(dirname "$0")/${TARGET/panel/panel}" >&2
  exit 1
fi

# Best-effort hint for the SSH tunnel command.
HOST_IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
USER_NAME="$(whoami)"

cat <<EOF
============================================================
 Previewing '$TARGET'  ($NS/$SVC :$RPORT  ->  localhost:$LPORT)

 On your LAPTOP, open an SSH tunnel (keep it running):

     ssh -L ${LPORT}:localhost:${LPORT} ${USER_NAME}@${HOST_IP:-<this-host>}

 then browse:   http://localhost:${LPORT}

 Press Ctrl-C here to stop the port-forward.
============================================================
EOF

exec kubectl -n "$NS" port-forward "svc/$SVC" "${LPORT}:${RPORT}" --address 127.0.0.1
