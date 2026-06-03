#!/usr/bin/env python3
"""Fabric Data Model API — a schema registry over the canonical primitives.

Inspired by the Adobe XDM Schema Registry API: classes, schemas, and graph as
addressable resources. Zero dependencies beyond PyYAML (reuses codegen).

Endpoints:
    GET /                      API index
    GET /health               liveness
    GET /classes              list primitives (classes)
    GET /classes/{name}       full primitive model
    GET /schemas/{name}       JSON Schema for a primitive
    GET /graph                {nodes, edges} of the whole model
    GET /resolve/{name}?depth=2   traverse the graph from a node

Run:
    python api/server.py                 # serve on :8088
    python api/server.py --selftest      # in-process checks, no socket
"""
from __future__ import annotations

import json
import os
import sys
from collections import deque
from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib.parse import urlparse, parse_qs

sys.path.insert(0, os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "codegen"))
import generate_models as gm  # noqa: E402


def short(pid: str) -> str:
    return pid.split(":")[-1]


def _index():
    prims = gm.load_primitives()
    return {p["id"]: p for p in prims}, {short(p["id"]): p for p in prims}


def list_classes():
    prims = gm.load_primitives()
    return [{"id": p["id"], "name": p["name"], "question": p.get("question", ""),
             "schemaOrg": p.get("schemaOrg", {}).get("type")} for p in prims]


def get_class(name: str):
    _, by_short = _index()
    return by_short.get(name)


def get_schema(name: str):
    p = get_class(name)
    return gm.gen_jsonschema(p) if p else None


def get_graph():
    prims = gm.load_primitives()
    known = {p["id"] for p in prims}
    nodes = [{"id": short(p["id"]), "name": p["name"]} for p in prims]
    edges = []
    for p in prims:
        for r in p.get("relationships", []):
            if r["target"] in known:
                edges.append({"from": short(p["id"]), "rel": r["name"],
                              "to": short(r["target"]), "cardinality": r.get("cardinality")})
    return {"nodes": nodes, "edges": edges}


def resolve(name: str, depth: int = 2):
    """BFS the relationship graph from `name`, returning the connected subgraph."""
    graph = get_graph()
    if name not in {n["id"] for n in graph["nodes"]}:
        return None
    adj = {}
    for e in graph["edges"]:
        adj.setdefault(e["from"], []).append(e)
    visited, path = {name}, []
    q = deque([(name, 0)])
    while q:
        node, d = q.popleft()
        if d >= depth:
            continue
        for e in adj.get(node, []):
            path.append(e)
            if e["to"] not in visited:
                visited.add(e["to"])
                q.append((e["to"], d + 1))
    return {"start": name, "depth": depth, "resolved": sorted(visited), "edges": path}


import hashlib
import time as _time

DATA_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "data")
SIGNUPS = os.path.join(DATA_DIR, "signups.jsonl")
EVENTS = os.path.join(DATA_DIR, "events.jsonl")

# Power the app with our own data model: build signups as Identity nodes.
sys.path.insert(0, os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "gen", "python"))
try:
    from models import Identity as IdentityModel  # generated dataclass
except Exception:
    IdentityModel = None


def _append(path: str, record: dict):
    os.makedirs(DATA_DIR, exist_ok=True)
    with open(path, "a") as f:
        f.write(json.dumps(record) + "\n")


def signup(payload: dict):
    """Register an alpha tester as an Identity node — Connect→Catalog→Govern→Activate."""
    email = (payload.get("email") or "").strip().lower()
    name = (payload.get("name") or "").strip()
    if not name or "@" not in email or "." not in email.split("@")[-1]:
        return 400, {"error": "valid name and email are required"}

    uid = "did:fabric:user:" + hashlib.sha256(email.encode()).hexdigest()[:16]
    now = _time.strftime("%Y-%m-%dT%H:%M:%SZ", _time.gmtime())

    # Catalog: construct via the generated data model (the app IS powered by it).
    if IdentityModel:
        identity = IdentityModel(id=uid, kind="person", displayName=name)
        identity_doc = {"id": identity.id, "kind": identity.kind,
                        "displayName": identity.displayName}
    else:
        identity_doc = {"id": uid, "kind": "person", "displayName": name}

    record = {**identity_doc, "email": email,
              "company": payload.get("company"), "useCase": payload.get("message"),
              "registeredAt": now}
    # Govern: append-only store + immutable audit Event.
    _append(SIGNUPS, record)
    event = {"id": f"event:identity.signup:{uid[-8:]}", "type": "identity.signup",
             "actor": uid, "occurredAt": now}
    _append(EVENTS, event)
    # Activate: return the created identity.
    return 201, {"status": "registered", "identity": identity_doc, "event": event,
                 "message": f"Welcome to the Fabric alpha, {name}."}


ROUTES_DOC = {
    "/": "this index",
    "/health": "liveness",
    "/classes": "list primitives (classes)",
    "/classes/{name}": "full primitive model",
    "/schemas/{name}": "JSON Schema for a primitive",
    "/graph": "nodes + edges of the model",
    "/resolve/{name}?depth=N": "traverse the graph from a node",
    "POST /signup": "register an alpha tester (creates an Identity node)",
    "/signups": "count of registered testers (no PII)",
}


def route(path: str, query: dict):
    parts = [p for p in path.split("/") if p]
    if not parts:
        return 200, {"service": "fabric-data-model-api", "routes": ROUTES_DOC}
    head = parts[0]
    if head == "health":
        return 200, {"status": "ok"}
    if head == "classes":
        if len(parts) == 1:
            return 200, list_classes()
        c = get_class(parts[1])
        return (200, c) if c else (404, {"error": f"class '{parts[1]}' not found"})
    if head == "schemas" and len(parts) == 2:
        s = get_schema(parts[1])
        return (200, s) if s else (404, {"error": f"schema '{parts[1]}' not found"})
    if head == "graph":
        return 200, get_graph()
    if head == "resolve" and len(parts) == 2:
        depth = int(query.get("depth", ["2"])[0])
        r = resolve(parts[1], depth)
        return (200, r) if r else (404, {"error": f"node '{parts[1]}' not found"})
    if head == "signups":
        n = sum(1 for _ in open(SIGNUPS)) if os.path.exists(SIGNUPS) else 0
        return 200, {"registered": n}
    return 404, {"error": "not found", "routes": ROUTES_DOC}


class Handler(BaseHTTPRequestHandler):
    def _send(self, status, body):
        payload = json.dumps(body, indent=2).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def do_OPTIONS(self):
        self.send_response(204)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")
        self.end_headers()

    def do_GET(self):
        u = urlparse(self.path)
        status, body = route(u.path, parse_qs(u.query))
        self._send(status, body)

    def do_POST(self):
        u = urlparse(self.path)
        length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(length) if length else b"{}"
        try:
            payload = json.loads(raw or b"{}")
        except json.JSONDecodeError:
            return self._send(400, {"error": "invalid JSON"})
        if u.path.rstrip("/") == "/signup":
            return self._send(*signup(payload))
        self._send(404, {"error": "not found", "routes": ROUTES_DOC})

    def log_message(self, *args):
        pass  # quiet


def selftest():
    classes = list_classes()
    assert len(classes) >= 11, f"expected >=11 primitives, got {len(classes)}"
    assert get_class("identity")["name"] == "Identity"
    assert get_class("risk")["name"] == "Risk"  # expanded node
    assert get_schema("policy")["title"] == "Policy"
    g = get_graph()
    assert len(g["nodes"]) == len(classes)
    assert all(e["to"] in {n["id"] for n in g["nodes"]} for e in g["edges"]), "dangling edge"
    r = resolve("identity", 2)
    assert "capability" in r["resolved"]
    print("selftest OK:", {"classes": len(classes), "edges": len(g["edges"]),
                           "resolve(identity,2)": r["resolved"]})


def main():
    if "--selftest" in sys.argv:
        selftest()
        return
    port = int(os.environ.get("PORT", "8088"))
    print(f"Fabric Data Model API on http://localhost:{port}  (Ctrl-C to stop)")
    HTTPServer(("0.0.0.0", port), Handler).serve_forever()


if __name__ == "__main__":
    main()
