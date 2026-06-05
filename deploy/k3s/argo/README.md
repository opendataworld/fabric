# Argo Workflows on k3s — agent composer

[Argo Workflows](https://argoproj.github.io/workflows/) (CNCF *graduated*) as the
engine for **composing agents as pipelines** — each step a container, wired into
a DAG (plan → retrieve/act → observe → review). Vendored as plain YAML
(quick-start-minimal, v3.6.2), same packaging as the rest of `deploy/k3s/`.

## What's here

| File | Purpose |
|------|---------|
| `upstream-argo.yaml` | Vendored Argo Workflows (controller, server, CRDs, RBAC) |
| `agent-composer-template.yaml` | Sample `agent-composer` WorkflowTemplate (a DAG agent loop) |
| `kustomization.yaml` | Base + pinned images |
| `install.sh` | Applies the base, waits for CRDs, applies the template |

## Install

```bash
cd deploy/k3s/argo
./install.sh
```

## Use it

Open the UI:
```bash
kubectl -n argo port-forward svc/argo-server 2746:2746
# remote: ssh -L 2746:localhost:2746 you@your-k3s-host
```
Browse <https://localhost:2746> (self-signed cert). Run the sample:
```bash
argo submit --from workflowtemplate/agent-composer -n argo --watch
```

## Composing a real agent

`agent-composer-template.yaml` is a DAG where each task is just a container.
Replace the `argosay` echo steps with real work — an LLM call, a tool
invocation, a retriever — passing data via parameters/artifacts between steps.
Argo handles ordering, parallelism (`retrieve` + `act` run concurrently),
retries, and history. Loops/branches map to `withItems`, `when:`, and recursion.

## Notes

- **quick-start-minimal** runs the server without login — fine behind a
  port-forward, **do not expose it publicly** as-is. For real use, set
  `--auth-mode=client` and put it behind an authenticating ingress.
- Pinned to v3.6.2; bump the `images:` block + re-vendor `upstream-argo.yaml`
  to upgrade.
