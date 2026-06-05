#!/usr/bin/env python3
"""Fabric platform CLI — connect all nodes to the DB, and validate the graph.

The DB is SurrealDB. This tool turns the model + an instance into a live graph,
and validates that the graph is sound.

Commands:
    validate [--instance F.yaml]   validate the model graph (and an instance)
    schema                         emit the SurrealQL schema (DEFINE TABLE/FIELD)
    create <instance.yaml>         load an instance into the DB (connect all nodes)

Connecting to the DB (optional — zero extra deps, uses SurrealDB's HTTP /sql):
    SURREAL_URL  (e.g. http://localhost:8000)   SURREAL_USER  SURREAL_PASS
    SURREAL_NS   SURREAL_DB
If those are unset, `schema`/`create` print the SurrealQL so you can pipe it to
`surreal sql`. `validate` never needs the DB — it runs anywhere.

    python tools/fabric.py validate
    python tools/fabric.py validate --instance examples/edge-guardian/fabric.yaml
    python tools/fabric.py create examples/edge-guardian/fabric.yaml
"""
from __future__ import annotations

import argparse
import base64
import json
import os
import sys
import urllib.request

sys.path.insert(0, os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "codegen"))
import generate_models as gm  # noqa: E402

# Link fields on instance assets that become graph edges (RELATE a->field->b).
LINK_FIELDS = {
    "executes", "pursues", "governedBy", "guards", "speaks", "constraints",
    "contains", "grants", "memberOf", "assignedTo", "hasMemory", "holds",
    "remembers", "binds", "hosts", "exposes", "emits", "securedBy",
}
# Asset fields that point at another asset by id (scalar record links).
REF_FIELDS = {"source", "subject", "controller"}


def _table(name: str) -> str:
    import re
    return re.sub(r"[^a-z0-9]+", "_", name.lower())


# ── validation ────────────────────────────────────────────────────────────────

def validate_model(prims):
    """Returns (errors, warnings)."""
    errors, warnings = [], []
    known = {p["id"] for p in prims}
    id2n = {p["id"]: p["name"] for p in prims}

    # 1. No dangling relationship targets (hard).
    adj = {p["id"]: [] for p in prims}
    for p in prims:
        for r in p.get("relationships", []):
            t = r["target"]
            if t in known:
                adj[p["id"]].append(t)
            else:
                errors.append(f"dangling edge: {p['id']} --{r['name']}--> {t} (unknown target)")

    # 2. Cycles (SCC) — semantic cycles are allowed; report as info/warning.
    cycles = _sccs(adj)
    for c in cycles:
        warnings.append("semantic cycle (ok for an ontology): " + " -> ".join(id2n[x] for x in c))

    # 3. Every primitive has the required shape.
    for p in prims:
        for req in ("id", "name", "question", "schemaOrg", "attributes"):
            if req not in p:
                errors.append(f"{p.get('id','?')}: missing required field '{req}'")
    return errors, warnings


def validate_instance(prims, inst):
    """Validate an instance: refs resolve, touchpoints are covered."""
    errors, warnings = [], []
    assets = inst.get("fabric", {}).get("assets", [])
    by_id = {a["id"]: a for a in assets}
    prim_ids = {p["id"] for p in prims}

    def resolves(ref):
        return ref in by_id or ref in prim_ids

    for a in assets:
        # every link/ref points at a defined asset
        for f in LINK_FIELDS:
            for ref in (a.get(f) or []):
                if not resolves(ref):
                    errors.append(f"{a['id']}.{f} -> {ref} does not resolve to a known asset")
        for f in REF_FIELDS:
            ref = a.get(f)
            if ref and not resolves(ref):
                errors.append(f"{a['id']}.{f} -> {ref} does not resolve")
        # touchpoint coverage: a touchpoint must declare surface + protocol
        if a.get("type") == "fabric:primitive:touchpoint":
            if not a.get("surface"):
                errors.append(f"touchpoint {a['id']} has no surface (boundary uncovered)")
            if not a.get("protocol"):
                errors.append(f"touchpoint {a['id']} has no protocol — define one")
        # agent stability: guarded touchpoints must exist
        if a.get("type") == "fabric:primitive:agent":
            if not a.get("executes"):
                warnings.append(f"agent {a['id']} executes nothing")
    return errors, warnings


def _sccs(adj):
    import sys as _s
    _s.setrecursionlimit(10000)
    index, low, onstk, stack, idx, out = {}, {}, {}, [], [0], []
    def strong(v):
        index[v] = low[v] = idx[0]; idx[0] += 1; stack.append(v); onstk[v] = True
        for w in adj.get(v, []):
            if w not in index:
                strong(w); low[v] = min(low[v], low[w])
            elif onstk.get(w):
                low[v] = min(low[v], index[w])
        if low[v] == index[v]:
            comp = []
            while True:
                w = stack.pop(); onstk[w] = False; comp.append(w)
                if w == v:
                    break
            if len(comp) > 1:
                out.append(comp)
    for v in list(adj):
        if v not in index:
            strong(v)
    return out


# ── SurrealQL generation (connect all nodes) ───────────────────────────────────

def _sql_val(v):
    if isinstance(v, bool):
        return "true" if v else "false"
    if isinstance(v, (int, float)):
        return str(v)
    return json.dumps(str(v))


def instance_to_surql(inst):
    """Turn an instance's assets into CREATE (nodes) + RELATE (edges)."""
    assets = inst.get("fabric", {}).get("assets", [])
    by_id = {a["id"]: a for a in assets}
    out = [f"-- {inst.get('fabric',{}).get('name','instance')} — connect all nodes"]

    def thing(ref):
        a = by_id.get(ref)
        tbl = _table(a["type"].split(":")[-1]) if a else "thing"
        return f"type::thing('{tbl}', {json.dumps(ref)})"

    # nodes
    for a in assets:
        if a.get("type") == "fabric:primitive:relationship":
            continue  # emitted as an edge below
        tbl = _table(a["type"].split(":")[-1])
        sets = []
        for k, v in a.items():
            if k in ("id", "type") or k in LINK_FIELDS:
                continue
            sets.append(f"{k} = {_sql_val(v)}")
        setclause = (" SET " + ", ".join(sets)) if sets else ""
        out.append(f"CREATE type::thing('{tbl}', {json.dumps(a['id'])}){setclause};")

    # edges from link fields
    for a in assets:
        for f in LINK_FIELDS:
            for ref in (a.get(f) or []):
                out.append(f"RELATE {thing(a['id'])}->{_table(f)}->{thing(ref)};")
    # explicit relationship assets
    for a in assets:
        if a.get("type") == "fabric:primitive:relationship":
            pred = _table(a.get("predicate", "relates"))
            out.append(f"RELATE {thing(a['source'])}->{pred}->{thing(a['target'])};")
    return "\n".join(out) + "\n"


def surreal_sql(statements: str):
    """POST SurrealQL to SurrealDB's HTTP /sql endpoint (stdlib only)."""
    url = os.environ["SURREAL_URL"].rstrip("/") + "/sql"
    auth = base64.b64encode(f"{os.environ['SURREAL_USER']}:{os.environ['SURREAL_PASS']}".encode()).decode()
    req = urllib.request.Request(url, data=statements.encode(), method="POST", headers={
        "Authorization": f"Basic {auth}",
        "Accept": "application/json",
        "surreal-ns": os.environ.get("SURREAL_NS", "fabric"),
        "surreal-db": os.environ.get("SURREAL_DB", "fabric"),
        "Content-Type": "text/plain",
    })
    with urllib.request.urlopen(req, timeout=30) as r:
        return r.read().decode()


def _have_db():
    return all(os.environ.get(k) for k in ("SURREAL_URL", "SURREAL_USER", "SURREAL_PASS"))


# ── driver ──────────────────────────────────────────────────────────────────────

def main():
    ap = argparse.ArgumentParser(description="Fabric platform: connect all nodes + validate the graph")
    sub = ap.add_subparsers(dest="cmd", required=True)
    v = sub.add_parser("validate"); v.add_argument("--instance")
    sub.add_parser("schema")
    c = sub.add_parser("create"); c.add_argument("instance")
    args = ap.parse_args()

    import yaml
    prims = gm.load_primitives()

    if args.cmd == "validate":
        errors, warnings = validate_model(prims)
        if args.instance:
            ie, iw = validate_instance(prims, yaml.safe_load(open(args.instance)))
            errors += ie; warnings += iw
        for w in warnings:
            print(f"  war: {w}")
        nodes = len(prims)
        edges = sum(len(p.get("relationships", [])) for p in prims)
        if errors:
            for e in errors:
                print(f"  ERROR: {e}", file=sys.stderr)
            print(f"\nFAIL — {nodes} nodes, {edges} edges, {len(errors)} error(s).", file=sys.stderr)
            sys.exit(1)
        print(f"\nOK — {nodes} nodes, {edges} edges, 0 dangling, {len(warnings)} semantic cycle(s).")
        return

    if args.cmd == "schema":
        sql = gm.gen_graph_surql(prims)
        if _have_db():
            print(surreal_sql(sql))
        else:
            sys.stdout.write(sql)
        return

    if args.cmd == "create":
        inst = yaml.safe_load(open(args.instance))
        ie, iw = validate_instance(prims, inst)
        for w in iw:
            print(f"  war: {w}", file=sys.stderr)
        if ie:
            for e in ie:
                print(f"  ERROR: {e}", file=sys.stderr)
            print("refusing to load an invalid instance.", file=sys.stderr)
            sys.exit(1)
        sql = instance_to_surql(inst)
        if _have_db():
            print(surreal_sql(sql))
            print(">> loaded into SurrealDB", file=sys.stderr)
        else:
            sys.stdout.write(sql)
            print(">> no SURREAL_URL set — printed SurrealQL (pipe to `surreal sql`)", file=sys.stderr)
        return


if __name__ == "__main__":
    main()
