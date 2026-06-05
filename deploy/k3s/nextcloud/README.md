# Nextcloud on k3s

A self-contained, cloud-native deployment of [Nextcloud](https://nextcloud.com)
for a [k3s](https://k3s.io) cluster. Plain Kubernetes manifests (no Helm),
applied with Kustomize, tuned for k3s defaults (Traefik ingress, the
`local-path` storage class) and hardened to CNCF best practices.

## What you get

| Component | Manifest | Notes |
|-----------|----------|-------|
| Namespace | `00-namespace.yaml` | `nextcloud` |
| Secrets | `01-secrets.example.yaml` | DB / Redis / admin credentials (copy & fill) |
| PostgreSQL 16 | `02-postgres.yaml` | StatefulSet + 10Gi `local-path` PVC |
| Redis 7 | `03-redis.yaml` | password-auth file locking + cache |
| Nextcloud 30 | `04-nextcloud.yaml` | apache web pod + cron sidecar + 20Gi PVC |
| Ingress + TLS | `05-ingress.yaml` | Traefik, cert-manager, security headers, CalDAV/CardDAV redirects |
| NetworkPolicies | `06-networkpolicy.yaml` | default-deny + explicit allows |
| PodDisruptionBudget | `07-pdb.yaml` | `minAvailable: 1` |
| cert-manager issuer | `99-cert-manager.example.yaml` | optional Let's Encrypt `ClusterIssuer` |

### Cloud-native / CNCF posture

- **Recommended labels** (`app.kubernetes.io/*`) on every object.
- **Hardened pods**: `runAsNonRoot` where the image allows, `seccompProfile:
  RuntimeDefault`, `allowPrivilegeEscalation: false`, dropped capabilities,
  read-only root filesystem on Redis.
- **Resource requests/limits** and **liveness / readiness / startup probes**
  on every workload.
- **NetworkPolicies** for zero-trust east-west traffic.
- **CNCF projects**: Kubernetes (k3s is a CNCF-certified distribution),
  Traefik (ingress), cert-manager (TLS), and optionally Prometheus —
  the web Service is ready to scrape via the `nextcloud-exporter`.

## Prerequisites

- A running k3s cluster (`kubectl` pointing at it) with the bundled Traefik
  ingress and `local-path` storage class (both are k3s defaults).
- A DNS record for your hostname pointing at the cluster's ingress IP.
- For automatic TLS: [cert-manager](https://cert-manager.io) installed.

## Deploy

```bash
cd deploy/k3s/nextcloud

# 1. Credentials (the real file is git-ignored)
cp 01-secrets.example.yaml 01-secrets.yaml
#   edit 01-secrets.yaml — replace every CHANGE_ME
#   generate strong values with:  openssl rand -base64 24

# 2. Set your hostname in 04-nextcloud.yaml (NEXTCLOUD_TRUSTED_DOMAINS)
#    and 05-ingress.yaml (host + tls.hosts), replacing nextcloud.example.com

# 3. (optional) TLS via Let's Encrypt
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
cp 99-cert-manager.example.yaml 99-cert-manager.yaml   # set your email
kubectl apply -f 99-cert-manager.yaml

# 4. Deploy everything
kubectl apply -k .

# 5. Watch it come up
kubectl -n nextcloud get pods -w
```

First boot installs Nextcloud into the data volume (the `startupProbe` allows
up to ~5 min). Once `kubectl -n nextcloud get pods` shows the web pod `Ready`,
browse to `https://<your-host>` and log in with the admin credentials from
your secret.

## Operations

```bash
# occ (Nextcloud admin CLI) — runs as the web user
kubectl -n nextcloud exec deploy/nextcloud -c nextcloud -- \
  su -s /bin/sh www-data -c "php occ status"

# tail logs
kubectl -n nextcloud logs deploy/nextcloud -c nextcloud -f

# DB shell
kubectl -n nextcloud exec -it sts/nextcloud-postgres -- \
  psql -U nextcloud -d nextcloud
```

### Backups

`local-path` volumes live on a single node — back them up. Snapshot the two
PVCs (`nextcloud-data`, `data-nextcloud-postgres-0`) and/or run
`pg_dump`/`occ files:scan` on a schedule. For multi-node durability, swap the
`storageClassName` for a CNCF storage layer such as Longhorn or Rook/Ceph.

## Customization

| Want to change | Where |
|----------------|-------|
| Hostname | `04-nextcloud.yaml` env + `05-ingress.yaml` |
| Storage sizes | `*.yaml` `resources.requests.storage` |
| Storage backend | `storageClassName` in PVCs |
| Upload / memory limits | `PHP_UPLOAD_LIMIT`, `PHP_MEMORY_LIMIT` |
| Image versions | `image:` tags (pin a digest for reproducibility) |
| Non-root web tier | switch to the `nextcloud:30-fpm` image + nginx sidecar |

## Notes & trade-offs

- **Single web replica.** `local-path` is `ReadWriteOnce` and node-local, so
  the web Deployment uses `Recreate` and the PVC must not be shared. To scale
  out, move `/var/www/html` to a `ReadWriteMany` backend and use external
  Redis for shared locking.
- **NetworkPolicies need a policy-aware CNI.** On stock flannel-only k3s they
  are inert (fail-open); install Cilium or Calico to enforce them.
- The **apache image is not fully rootless** (binds :80 and chowns on first
  boot). The `-fpm` variant behind nginx gives a non-root web tier if your
  compliance baseline requires it.
