# k3s Stack Catalog

A lightweight, desktop-style **catalog UI** for the cloud-native components
running on (or available to) your k3s cluster — laid out as CNCF-landscape-style
grouped logo-cards. Dependency-free static site (no build step, no framework),
served by a hardened non-root nginx and deployable in one command.

![layout](https://img.shields.io/badge/UI-vanilla%20JS-blue) ![cncf](https://img.shields.io/badge/style-CNCF%20landscape-5b6ee1)

## What it is

- **`site/`** — the entire UI: `index.html`, `styles.css`, `app.js`, and
  **`catalog.json`** (the catalog data). The page is rendered at runtime from
  `catalog.json`, so cataloguing is pure data — edit one file, no rebuild.
- **`10-panel.yaml`** — non-root nginx (`nginx-unprivileged`, port 8080,
  read-only root fs, dropped caps, `RuntimeDefault` seccomp), Service, and a
  Traefik Ingress with optional cert-manager TLS.
- **`kustomization.yaml`** — turns `site/*` into a content-hashed `panel-site`
  ConfigMap (edits auto-trigger a rollout) and pins the image.
- **`install.sh`** — one-command deploy.

## Deploy

```bash
cd deploy/k3s/panel
./install.sh catalog.example.com you@example.com   # TLS
# or
./install.sh catalog.example.com                   # no auto-TLS
# or manually, after setting the host in 10-panel.yaml:
kubectl apply -k .
```

## Cataloguing

`site/catalog.json` is the single source of truth:

```json
{
  "title": "k3s Stack Catalog",
  "subtitle": "…",
  "categories": [
    {
      "name": "Application",
      "items": [
        {
          "name": "Nextcloud",
          "logo": "☁️",
          "description": "Self-hosted file sync.",
          "image": "nextcloud:30-apache",
          "status": "deployed",     // deployed | available
          "maturity": "graduated",  // graduated | incubating | sandbox | non-cncf
          "docs": "https://nextcloud.com",
          "manifest": "../nextcloud"
        }
      ]
    }
  ]
}
```

Add/extend categories and items, re-apply, and the panel rolls out the new
catalog automatically. The UI provides category filtering, free-text search,
per-card deployed/available status, and CNCF **maturity** badges
(graduated / incubating / sandbox).

It ships pre-catalogued from the **CNCF landscape** taxonomy — Orchestration &
Management, App Definition & Development, Runtime, Provisioning & Security,
Observability & Analysis, Service Mesh & Networking — with real projects
(Kubernetes, etcd, Helm, Argo, Cilium, cert-manager, Prometheus, OpenTelemetry,
Istio, Falco, …) tagged by maturity, and the companion
[`../nextcloud`](../nextcloud) stack items marked *deployed*. Maturity is a
snapshot; refresh against <https://landscape.cncf.io>.

## Preview

- **Local:** `cd site && python3 -m http.server 8080` → open <http://localhost:8080>
  (serve it — `fetch('catalog.json')` is blocked on `file://`).
- **Hosted (clickable URL):** the repo's GitHub Pages workflow publishes this
  panel at **`<pages-url>/panel/`** (assembled from `site/` here; deploys on
  push to `main`).
- **On the cluster, no DNS:** run [`../preview.sh`](../preview.sh) on the k3s
  host to port-forward the Service and get a ready-made `ssh -L` tunnel command.

## Notes

- Hardened: runs non-root with a read-only root filesystem (writable `tmp`/cache
  via `emptyDir`), all capabilities dropped.
- The catalog is descriptive, not a live cluster client — it does not call the
  Kubernetes API, so it needs no RBAC and exposes no cluster data. Wire the
  `status` field from your CI/GitOps if you want it to reflect live state.
