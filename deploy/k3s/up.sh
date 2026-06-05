#!/usr/bin/env bash
#
#  The whole stack, one command.
#
#      ./up.sh [base-domain] [letsencrypt-email]
#
#  Examples:
#      ./up.sh                                  # NO domain needed — auto-detects
#                                               # your IP and uses <ip>.sslip.io
#      ./up.sh example.com you@example.com      # real DNS + automatic TLS
#      ./up.sh 203.0.113.10.sslip.io            # a specific IP, instant DNS
#
#  Brings up, in dependency order, on one domain:
#      cert-manager (TLS)                -> prerequisite
#      Nextcloud      nextcloud.<domain> -> file-sync app + Postgres + Redis
#      Catalog panel  catalog.<domain>   -> CNCF-landscape catalog UI
#      Zot registry   registry.<domain>  -> OCI registry (UI + signature verify)
#      K8s Dashboard  dashboard.<domain> -> cluster control panel (basic-auth)
#      Argo Workflows (port-forward)     -> agent-composer pipelines
#
#  Each component is idempotent, so re-running is safe. Components that need a
#  missing tool (e.g. htpasswd for the registry) are skipped, not fatal.
set -uo pipefail

BASE="${1:-}"
EMAIL="${2:-}"
command -v kubectl >/dev/null || { echo "error: kubectl not found" >&2; exit 1; }

# No domain given? Auto-detect the server IP and use sslip.io — zero DNS setup.
if [[ -z "$BASE" ]]; then
  IP="$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}' 2>/dev/null)"
  [[ -z "$IP" ]] && IP="$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' 2>/dev/null)"
  [[ -z "$IP" ]] && IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
  if [[ -z "$IP" ]]; then
    echo "could not auto-detect an IP — pass one explicitly: $0 <domain-or-ip.sslip.io>" >&2
    exit 1
  fi
  BASE="${IP}.sslip.io"
  echo ">> no domain given — using ${BASE} (auto DNS via sslip.io, nothing to configure)"
fi
SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CM_VERSION="v1.16.2"

step() { printf '\n\033[1;34m== %s ==\033[0m\n' "$*"; }
run()  { echo "+ $*"; "$@" || echo "  (step failed — continuing)"; }

# ---------------------------------------------------------------------------
step "1/6  cert-manager (TLS)"
if ! kubectl get ns cert-manager >/dev/null 2>&1; then
  run kubectl apply -f "https://github.com/cert-manager/cert-manager/releases/download/${CM_VERSION}/cert-manager.yaml"
  run kubectl -n cert-manager rollout status deploy/cert-manager-webhook --timeout=180s
else
  echo "  cert-manager already present"
fi
if [[ -n "$EMAIL" ]]; then
  echo "  applying letsencrypt-prod ClusterIssuer for ${EMAIL}"
  sed "s|CHANGE_ME@example.com|${EMAIL}|" "$SRC/nextcloud/99-cert-manager.example.yaml" | kubectl apply -f - \
    || echo "  (issuer apply failed — continuing)"
else
  echo "  no email supplied — skipping issuer (services use Traefik's self-signed cert)"
fi

step "2/6  Nextcloud  (nextcloud.${BASE})"
run bash "$SRC/nextcloud/install.sh" "nextcloud.${BASE}" "$EMAIL"

step "3/6  Catalog panel  (catalog.${BASE})"
run bash "$SRC/panel/install.sh" "catalog.${BASE}" "$EMAIL"

step "4/6  Zot registry  (registry.${BASE})"
run bash "$SRC/registry/install.sh" "registry.${BASE}" "$EMAIL"

step "5/6  Kubernetes Dashboard  (dashboard.${BASE})"
run bash "$SRC/dashboard/install.sh" "dashboard.${BASE}"

step "6/6  Argo Workflows  (agent composer)"
run bash "$SRC/argo/install.sh"

# ---------------------------------------------------------------------------
SCHEME="https"
cat <<EOF

############################################################
#  Stack is up on  ${BASE}
#
#    Nextcloud   ${SCHEME}://nextcloud.${BASE}
#    Catalog     ${SCHEME}://catalog.${BASE}
#    Registry    ${SCHEME}://registry.${BASE}
#    Dashboard   ${SCHEME}://dashboard.${BASE}   (set basic-auth first)
#    Argo        kubectl -n argo port-forward svc/argo-server 2746:2746
#
#  Credentials for each app were printed by its installer above — save them.
#  DNS: point *.${BASE} (or each host) at the cluster's ingress IP.
############################################################
EOF
