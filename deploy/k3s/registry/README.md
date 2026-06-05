# Zot registry on k3s

A lightweight, **OCI-native container registry** ([Zot](https://zotregistry.dev))
for your k3s cluster — single container, web UI, search, and built-in
**cosign / notation signature verification**. A much smaller footprint than
Harbor, packaged the same plain-YAML way as the rest of `deploy/k3s/`.

## What's here

| File | Purpose |
|------|---------|
| `01-config.yaml` | Zot config (storage, htpasswd auth, anonymous read, UI/search/trust extensions) |
| `02-htpasswd.example.yaml` | Push credential (copy → `02-htpasswd.yaml`, git-ignored) |
| `10-zot.yaml` | Hardened non-root Deployment + 20Gi PVC + Service |
| `20-ingress.yaml` | Traefik ingress + cert-manager TLS |
| `install.sh` | One-command deploy (generates a bcrypt credential, sets host) |

## Deploy

```bash
cd deploy/k3s/registry
./install.sh registry.yourdomain.com you@email.com
```

(Needs `htpasswd` from `apache2-utils` to mint the bcrypt credential, plus
cert-manager + a `letsencrypt-prod` issuer for TLS.) The installer prints the
push username/password and a `docker login` example.

## Use it

```bash
docker login registry.yourdomain.com -u pushuser
docker tag myapp registry.yourdomain.com/myapp:1.0
docker push registry.yourdomain.com/myapp:1.0
```

- **Pulls** are anonymous (read-only); **pushes** need the credential.
- Browse **https://registry.yourdomain.com** for the web UI (image list, tags,
  vulnerability/search, signature status).

## Signature verification

The `trust` extension is on, so Zot understands **cosign** and **notation**
signatures. Sign a pushed image and the UI flags it as signed:

```bash
cosign sign registry.yourdomain.com/myapp:1.0
```

This is the registry side of the supply-chain story — pair it with a cluster
admission policy (Kyverno `verifyImages` / sigstore policy-controller) so k3s
only runs signed images.

## Notes

- Hardened: non-root (uid 1000), read-only root fs (writable `/tmp` + data
  PVC), all caps dropped, `RuntimeDefault` seccomp.
- `install.sh` is idempotent — the credential is generated once and reused on
  re-runs (it won't clobber an existing one).
- Storage is a node-local `local-path` PVC; back it up or switch the
  `storageClassName` for multi-node durability.
