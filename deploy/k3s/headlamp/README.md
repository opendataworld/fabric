# Headlamp on k3s

[Headlamp](https://headlamp.dev) — a lightweight **CNCF** Kubernetes UI. Single
non-root deployment, plain-YAML, one-command. The k3s alternative to the heavier
official Dashboard.

## Deploy

```bash
cd deploy/k3s/headlamp
./install.sh
```

It applies the deployment + RBAC, then prints a **login token** and the
port-forward command.

## Access

```bash
kubectl -n headlamp port-forward svc/headlamp 8080:80
# remote:  ssh -L 8080:localhost:8080 root@<k3s-host>
```
Open **http://localhost:8080**, paste the token.

Get a token anytime:
```bash
kubectl -n headlamp create token headlamp-admin
```

## Notes

- Hardened: non-root (uid 100), dropped caps, `RuntimeDefault` seccomp,
  writable `/tmp` via emptyDir, pinned image `v0.25.0`.
- **Port-forward only** by design — a cluster-admin UI should not be public.
- `headlamp-admin` is bound to `cluster-admin`; switch the binding in
  `10-rbac.yaml` to `view` for read-only.
- Headlamp authenticates by **forwarding your bearer token** to the API server,
  so the pod's own SA stays minimal.
