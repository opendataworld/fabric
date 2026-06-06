#!/usr/bin/env python3
"""Fabric data model generator.

Reads the canonical foundation primitives (the `*/<name>-model.yaml` files) and
generates concrete data models in multiple targets:

    python      -> dataclasses
    typescript  -> interfaces
    jsonschema  -> one JSON Schema per primitive
    sql         -> CREATE TABLE DDL

This turns the canonical spec into real, usable artifacts. One source of truth
(the primitives), many generated outputs.

Usage:
    python codegen/generate_models.py --target all --out gen
    python codegen/generate_models.py --target python
"""
from __future__ import annotations

import argparse
import glob
import json
import os
import re

try:
    import yaml
except ImportError:
    raise SystemExit("PyYAML required: pip install pyyaml")

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


# ── load primitives ────────────────────────────────────────────────────────────

def load_primitives() -> list[dict]:
    prims = []
    for path in sorted(glob.glob(os.path.join(REPO, "*", "*-model.yaml"))):
        if os.path.basename(os.path.dirname(path)) == "meta":
            continue
        m = yaml.safe_load(open(path))
        if isinstance(m, dict) and str(m.get("id", "")).startswith("fabric:primitive:"):
            prims.append(m)
    return prims


def class_name(primitive: dict) -> str:
    return primitive["name"].replace(" ", "")


def attrs(primitive: dict) -> list[dict]:
    return primitive.get("attributes", [])


# ── type mapping ────────────────────────────────────────────────────────────────
# Fabric attribute types -> per-target types.

def _base(fabric_type: str) -> str:
    t = fabric_type.strip().lower()
    m = re.match(r"list<(.+)>", t)
    return m.group(1) if m else t


def _is_list(fabric_type: str) -> bool:
    return fabric_type.strip().lower().startswith("list<")


PY = {
    "string": "str", "uri": "str", "iso8601-duration": "str", "geojson": "str",
    "ref": "str", "enum": "str", "param": "dict", "metric": "dict",
    "datetime": "datetime", "number": "float", "integer": "int",
    "boolean": "bool", "map": "dict",
}
TS = {
    "string": "string", "uri": "string", "iso8601-duration": "string", "geojson": "string",
    "ref": "string", "enum": "string", "param": "Record<string, unknown>",
    "metric": "Record<string, unknown>", "datetime": "string", "number": "number",
    "integer": "number", "boolean": "boolean", "map": "Record<string, unknown>",
}
JS = {  # json schema
    "string": {"type": "string"}, "uri": {"type": "string", "format": "uri"},
    "iso8601-duration": {"type": "string"}, "geojson": {"type": "string"},
    "ref": {"type": "string"}, "enum": {"type": "string"},
    "param": {"type": "object"}, "metric": {"type": "object"},
    "datetime": {"type": "string", "format": "date-time"},
    "number": {"type": "number"}, "integer": {"type": "integer"},
    "boolean": {"type": "boolean"}, "map": {"type": "object"},
}
SQL = {
    "string": "TEXT", "uri": "TEXT", "iso8601-duration": "TEXT", "geojson": "JSONB",
    "ref": "TEXT", "enum": "TEXT", "param": "JSONB", "metric": "JSONB",
    "datetime": "TIMESTAMPTZ", "number": "DOUBLE PRECISION", "integer": "INTEGER",
    "boolean": "BOOLEAN", "map": "JSONB",
}
SURREAL = {  # SurrealQL field types
    "string": "string", "uri": "string", "iso8601-duration": "string", "geojson": "object",
    "ref": "record", "enum": "string", "param": "object", "metric": "object",
    "datetime": "datetime", "number": "number", "integer": "int",
    "boolean": "bool", "map": "object",
}


# ── generators ──────────────────────────────────────────────────────────────────

def gen_python(prims: list[dict]) -> str:
    out = ['"""Generated from Fabric primitives — DO NOT EDIT BY HAND."""',
           "from __future__ import annotations", "",
           "from dataclasses import dataclass, field",
           "from datetime import datetime", "from typing import Optional", "", ""]
    for p in prims:
        out.append("@dataclass")
        out.append(f"class {class_name(p)}:")
        out.append(f'    """{p.get("question","")}  ({p["id"]})"""')
        required, optional = [], []
        for a in attrs(p):
            base = PY.get(_base(a["type"]), "str")
            ty = f"list[{base}]" if _is_list(a["type"]) else base
            if a.get("required"):
                required.append(f"    {a['name']}: {ty}")
            else:
                default = "field(default_factory=list)" if _is_list(a["type"]) else "None"
                opt_ty = f"list[{base}]" if _is_list(a["type"]) else f"Optional[{ty}]"
                optional.append(f"    {a['name']}: {opt_ty} = {default}")
        body = required + optional
        out.extend(body if body else ["    pass"])
        out.append("")
        out.append("")
    return "\n".join(out)


def gen_typescript(prims: list[dict]) -> str:
    out = ["// Generated from Fabric primitives — DO NOT EDIT BY HAND.", ""]
    for p in prims:
        out.append(f"/** {p.get('question','')}  ({p['id']}) */")
        out.append(f"export interface {class_name(p)} {{")
        for a in attrs(p):
            base = TS.get(_base(a["type"]), "string")
            ty = f"{base}[]" if _is_list(a["type"]) else base
            opt = "" if a.get("required") else "?"
            out.append(f"  {a['name']}{opt}: {ty};")
        out.append("}")
        out.append("")
    return "\n".join(out)


def gen_jsonschema(p: dict) -> dict:
    props, required = {}, []
    for a in attrs(p):
        base = JS.get(_base(a["type"]), {"type": "string"})
        schema = {"type": "array", "items": base} if _is_list(a["type"]) else dict(base)
        if a.get("description"):
            schema["description"] = a["description"]
        props[a["name"]] = schema
        if a.get("required"):
            required.append(a["name"])
    out = {
        "$schema": "https://json-schema.org/draft/2020-12/schema",
        "$id": f"https://opendataworld.org/fabric/schema/{p['id'].split(':')[-1]}.json",
        "title": p["name"], "description": p.get("description", "").strip(),
        "type": "object", "properties": props,
    }
    if required:
        out["required"] = required
    return out


def gen_sql(prims: list[dict]) -> str:
    out = ["-- Generated from Fabric primitives — DO NOT EDIT BY HAND.", ""]
    for p in prims:
        table = re.sub(r"[^a-z0-9]+", "_", p["name"].lower())
        out.append(f"CREATE TABLE IF NOT EXISTS {table} (")
        cols = []
        for a in attrs(p):
            col = SQL.get(_base(a["type"]), "TEXT")
            if _is_list(a["type"]):
                col = "JSONB"
            constraint = " NOT NULL" if a.get("required") else ""
            pk = " PRIMARY KEY" if a["name"] == "id" else ""
            cols.append(f"    {a['name']} {col}{pk}{constraint}")
        out.append(",\n".join(cols))
        out.append(");")
        out.append("")
    return "\n".join(out)


def _table(name: str) -> str:
    return re.sub(r"[^a-z0-9]+", "_", name.lower())


def short(pid: str) -> str:
    return pid.split(":")[-1]


def gen_graph_surql(prims: list[dict]) -> str:
    id2table = {p["id"]: _table(p["name"]) for p in prims}
    out = ["-- Graph schema generated from Fabric primitives — DO NOT EDIT BY HAND.",
           "-- Nodes = primitives (fully expanded with fields); edges = RELATION tables.", ""]
    for p in prims:
        tbl = id2table[p["id"]]
        out.append(f"DEFINE TABLE {tbl} SCHEMAFULL;")
        for a in attrs(p):
            if a["name"] == "id":  # SurrealDB manages the record id itself
                continue
            base = SURREAL.get(_base(a["type"]), "string")
            ty = f"array<{base}>" if _is_list(a["type"]) else base
            if not a.get("required"):
                ty = f"option<{ty}>"
            out.append(f"DEFINE FIELD {a['name']} ON TABLE {tbl} TYPE {ty};")
        out.append("")
    out.append("-- Relationship edges")
    for p in prims:
        src = id2table[p["id"]]
        for r in p.get("relationships", []):
            tgt = id2table.get(r["target"])
            if not tgt:
                continue
            out.append(f"DEFINE TABLE {_table(r['name'])} TYPE RELATION IN {src} OUT {tgt};")
    return "\n".join(out) + "\n"


def gen_graph_mermaid(prims: list[dict]) -> str:
    id2name = {p["id"]: p["name"].replace(" ", "") for p in prims}
    out = ["graph LR"]
    for p in prims:
        src = p["name"].replace(" ", "")
        for r in p.get("relationships", []):
            tgt = id2name.get(r["target"])
            if tgt:
                out.append(f"  {src} -->|{r['name']}| {tgt}")
    return "\n".join(out) + "\n"


def gen_graph_cypher(prims: list[dict]) -> str:
    """Neo4j Cypher to load the model graph — for Neo4j Browser / Bloom."""
    known = {p["id"] for p in prims}
    out = ["// Neo4j Cypher generated from Fabric primitives — DO NOT EDIT BY HAND.",
           "// Load in Neo4j Browser, then explore in Bloom.", ""]
    for p in prims:
        q = (p.get("question", "") or "").replace("'", "\\'")
        out.append(f"CREATE (:Primitive {{id:'{short(p['id'])}', name:'{p['name']}', question:'{q}'}});")
    out.append("")
    for p in prims:
        for r in p.get("relationships", []):
            if r["target"] not in known:
                continue
            rel = re.sub(r"[^A-Z0-9]+", "_", r["name"].upper())
            out.append(
                f"MATCH (a:Primitive {{id:'{short(p['id'])}'}}), "
                f"(b:Primitive {{id:'{short(r['target'])}'}}) "
                f"CREATE (a)-[:{rel}]->(b);")
    return "\n".join(out) + "\n"


def gen_model_json(prims: list[dict]) -> dict:
    """Static model export so the frontend can run with zero backend."""
    known = {p["id"] for p in prims}
    classes = [{"id": short(p["id"]), "name": p["name"], "question": p.get("question", "")}
               for p in prims]
    nodes = [{"id": short(p["id"]), "name": p["name"]} for p in prims]
    edges = []
    for p in prims:
        for r in p.get("relationships", []):
            if r["target"] in known:
                edges.append({"from": short(p["id"]), "rel": r["name"], "to": short(r["target"])})
    return {"classes": classes, "nodes": nodes, "edges": edges}


def gen_jsonld(prims: list[dict]) -> dict:
    """All nodes as a schema.org-aligned JSON-LD @graph (the model as linked data)."""
    known = {p["id"] for p in prims}
    ctx = {
        "schema": "https://schema.org/",
        "rdfs": "http://www.w3.org/2000/01/rdf-schema#",
        "fabric": "https://opendataworld.io/fabric#",
    }
    for p in prims:  # merge each primitive's declared JSON-LD context terms
        ctx.update(p.get("jsonldContext", {}) or {})

    graph = []
    for p in prims:
        so = p.get("schemaOrg", {}) or {}
        node = {
            "@id": p["id"],
            "@type": so.get("type", "rdfs:Class"),
            "rdfs:label": p["name"],
            "rdfs:comment": p.get("question") or p.get("description", ""),
        }
        if so.get("url"):
            node["schema:sameAs"] = {"@id": so["url"]}
        props = []
        for a in attrs(p):
            prop = {
                "@id": f"fabric:{_table(p['name'])}.{a['name']}",
                "rdfs:label": a["name"],
                "fabric:dataType": a["type"],
                "fabric:required": bool(a.get("required")),
            }
            if a.get("schemaOrg"):
                prop["schema:sameAs"] = {"@id": a["schemaOrg"]}
            props.append(prop)
        if props:
            node["fabric:attribute"] = props
        rels = []
        for r in p.get("relationships", []):
            if r["target"] not in known:
                continue
            rd = {"@id": f"fabric:{_table(r['name'])}", "rdfs:label": r["name"],
                  "schema:rangeIncludes": {"@id": r["target"]}}
            if r.get("schemaOrg"):
                rd["schema:sameAs"] = {"@id": r["schemaOrg"]}
            rels.append(rd)
        if rels:
            node["fabric:relationship"] = rels
        graph.append(node)
    return {"@context": ctx, "@graph": graph}


def graph_fit_report(prims: list[dict]) -> tuple[int, int, list[str]]:
    """Returns (nodes, edges, dangling) — does the model fit a graph cleanly?"""
    known = {p["id"] for p in prims}
    edges, dangling = 0, []
    for p in prims:
        for r in p.get("relationships", []):
            if r["target"] in known:
                edges += 1
            else:
                dangling.append(f"{p['id']} --{r['name']}--> {r['target']} (unknown target)")
    return len(prims), edges, dangling


# ── driver ──────────────────────────────────────────────────────────────────────

def write(path: str, content: str):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    open(path, "w").write(content)
    print(f"  wrote {os.path.relpath(path, REPO)}")


def main():
    ap = argparse.ArgumentParser(description="Fabric data model generator")
    ap.add_argument("--target",
                    choices=["python", "typescript", "jsonschema", "sql", "graph", "json", "jsonld", "all"],
                    default="all")
    ap.add_argument("--out", default="gen")
    args = ap.parse_args()

    prims = load_primitives()
    out = os.path.join(REPO, args.out)
    print(f"Loaded {len(prims)} primitives. Generating target={args.target}")

    t = args.target
    if t in ("python", "all"):
        write(os.path.join(out, "python", "models.py"), gen_python(prims))
    if t in ("typescript", "all"):
        write(os.path.join(out, "typescript", "models.ts"), gen_typescript(prims))
    if t in ("jsonschema", "all"):
        for p in prims:
            name = p["id"].split(":")[-1]
            write(os.path.join(out, "jsonschema", f"{name}.json"),
                  json.dumps(gen_jsonschema(p), indent=2) + "\n")
    if t in ("sql", "all"):
        write(os.path.join(out, "sql", "schema.sql"), gen_sql(prims))
    if t in ("graph", "all"):
        write(os.path.join(out, "graph", "schema.surql"), gen_graph_surql(prims))
        write(os.path.join(out, "graph", "model.mmd"), gen_graph_mermaid(prims))
        write(os.path.join(out, "graph", "model.cypher"), gen_graph_cypher(prims))
    if t in ("jsonld", "all"):
        write(os.path.join(out, "jsonld", "model.jsonld"),
              json.dumps(gen_jsonld(prims), indent=2) + "\n")
    if t in ("json", "all"):
        model = gen_model_json(prims)
        doc = json.dumps(model, indent=2) + "\n"
        write(os.path.join(out, "model.json"), doc)
        # Also emit into the site so it can run fully static (no backend).
        write(os.path.join(REPO, "site", "data", "model.json"), doc)
        nodes, edges, dangling = graph_fit_report(prims)
        print(f"  graph fit: {nodes} nodes, {edges} edges, {len(dangling)} dangling")
        for d in dangling:
            print(f"    ! {d}")
    print("Done.")


if __name__ == "__main__":
    main()
