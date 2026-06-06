# Liferay on Kubernetes — Helm chart

A starter Helm chart that deploys [Liferay](https://www.liferay.com/) DXP/Portal
on Kubernetes. It realizes the **Liferay on Kubernetes** product MVP defined in
the [Fabric Product MVP Builder](../../site/build.html):

| Feature (from the MVP)      | How it's implemented here |
|-----------------------------|---------------------------|
| Helm chart & values         | this chart + `values.yaml` |
| Autoscaling & HPA           | `templates/hpa.yaml` (`autoscaling.enabled`) |
| Persistent storage          | `templates/pvc.yaml` mounted at `/opt/liferay/data` |
| SSO integration             | `sso.oidc.*` env hooks in `templates/deployment.yaml` |
| Observability dashboards    | Prometheus scrape annotations (`metrics.podAnnotations`) |
| Zero-downtime upgrades      | `RollingUpdate` with `maxUnavailable: 0` |

> **Status:** starter / demo. The image tag, database, SSO, and metrics endpoint
> must be configured for any real environment. See the warnings below.

## Prerequisites

- Kubernetes 1.23+
- Helm 3.8+
- A default `StorageClass` (for persistence), and `metrics-server` (for the HPA)
- A real Liferay image tag — pin `image.tag` (see
  [Docker Hub](https://hub.docker.com/r/liferay/portal/tags))

## Install

```bash
# from the repo root
helm install liferay ./deploy/liferay \
  --namespace liferay --create-namespace

# watch it boot (Liferay's first start takes a few minutes)
kubectl get pods -n liferay -w
```

## Common overrides

```bash
helm upgrade --install liferay ./deploy/liferay -n liferay \
  --set image.tag=7.4.3.132-ga132 \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=liferay.example.com \
  --set autoscaling.enabled=true \
  --set database.external.enabled=true \
  --set database.external.url='jdbc:postgresql://postgres:5432/lportal' \
  --set database.external.username=liferay \
  --set database.external.existingSecret=liferay-db
```

## Validate without a cluster

```bash
helm lint ./deploy/liferay
helm template liferay ./deploy/liferay | kubectl apply --dry-run=client -f -
```

## ⚠️ Before production

- **Database:** the chart defaults to Liferay's embedded **HSQL** (demo only).
  Set `database.external.enabled=true` and point at a managed Postgres/MySQL.
- **Secrets:** prefer `existingSecret` references over plaintext values.
- **Image:** pin a specific, supported `image.tag`.
- **Metrics:** wire a real Liferay metrics endpoint before relying on the
  Prometheus annotations / HPA memory target.
- **Clustering:** running `replicaCount > 1` requires Liferay cluster config
  (Cluster Link / unicast) and `ReadWriteMany` storage — not enabled here.
