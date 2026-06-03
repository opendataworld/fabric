# SurrealDB ↔ Fabric feature coverage map

The platform's stance on "implement all features": **SurrealDB already implements
the database features — Fabric uses them** via the TypeScript client (`web/fabric.ts`)
and the generated schema (`gen/graph/schema.surql`). This table is the honest
coverage map: capability → how Fabric uses it → the catalog feature it backs.

| SurrealDB capability | How Fabric uses it | Catalog feature |
|----------------------|--------------------|-----------------|
| **Multi-model** (document/graph/relational/KV) | One engine for all 39 nodes + edges | Multi-Model Data |
| **Record IDs + RELATE** (graph edges) | The `Relationship` primitive; `owns`, `linked_via`, etc. | Graph Model |
| **SurrealQL** | All reads/writes/traversals | SurrealQL |
| **Live queries** (`LIVE SELECT`) | Live graph + Fabric Agent push (`live()`); replaces the SSE prototype | Live Queries (Data Sync) |
| **Change feeds** (CDC + replay from versionstamp) | Lineage/audit replay; rebuild state from history | Change Feeds (CDC) |
| **Events & triggers** (`DEFINE EVENT`, async events) | Enforce `Policy`/`Control` on write; emit audit `Event`s | Events & Triggers |
| **Record/field permissions** | `Policy`/`Constraint`/`Permission` enforced in-engine | Row-Level Permissions |
| **Authentication** (record/system users, JWT, access) | The `Identity`/`Account`/`Application` runtime | Authentication |
| **Namespaces & databases** | Multi-tenancy — the `Tenant` node | Multi-Tenancy |
| **Schemafull/schemaless** | `Schema`/`FieldGroup` governance vs. agility | Schemafull / Schemaless |
| **Functions** (`DEFINE FUNCTION`) | `Constraint` expressions, derived `Metric`s | Functions |
| **Full-text search** | Discovery over cataloged `Dataset`s | Full-Text Search |
| **Vector search** (HNSW/MTREE) | `Semantic Search`, Graph RAG | Vector & Embeddings |
| **Geospatial** | The `Location` primitive | Geospatial |
| **GraphQL** (`DEFINE CONFIG GRAPHQL AUTO`) | Auto GraphQL API from the schema — free `Platform API`/GraphQL | (Platform API product) |
| **Query optimization** (indexes, async events) | Performance of graph/resolve queries at scale | (Performance) |
| **ACID transactions** (snapshot isolation) | Atomic governed writes | ACID Transactions |
| **Deployment** (embedded / RocksDB / IndexedDB / Cloud) | Local-first, "browser as server", and managed Fabric DB | Embedded / Distributed / Cloud |
| **SDKs** (incl. JS/TS, WASM) | Single-language TS client, browser-direct | SDKs |

## Architecture consequence
Because SurrealDB provides all of the above and supports **direct browser
connections** with in-engine permissions, Fabric needs **no separate API
server**. One language (TypeScript) + SurrealDB = the whole runtime. The Python
API (`api/server.py`) and Rust sketch (`fabric-db/`) are legacy reference
prototypes.

## What Fabric actually builds (the value on top of SurrealDB)
- The **canonical 39-node data model** (the hard, usually-failed part) and its
  generated schema/code in many targets.
- The **identity fabric** (14 connectors) that populates the graph.
- **Governance-by-construction** wiring (Policy/Constraint/Event as first-class).
- The **catalogs, agent, and visualizations** that make it usable and sellable.

SurrealDB is the engine; Fabric is the canonical model, governance, identity,
and experience layered on top.

## Durable execution (does SurrealDB replace Temporal?) — No

SurrealDB provides a durable **event log** (change feeds + versionstamp replay)
and **reactive triggers** (async `DEFINE EVENT`). It does **not** provide durable
*execution*: resumable workflow code surviving crashes, automatic retries/
timeouts/heartbeats, durable long timers, or deterministic workflow replay.

| Need | Provided by |
|------|-------------|
| Append-only event log / event sourcing | **SurrealDB** (change feeds) |
| Reactive triggers on data change | **SurrealDB** (async events) |
| Resumable multi-step workflows, retries, long timers, sagas | **Durable execution engine** (Temporal / Restate / Inngest) |

For `Agent Runtime` and `Pipeline` orchestration, add a durable execution engine
on top — or build a minimal orchestrator that uses SurrealDB change feeds as the
journal. Async events are for lightweight reactivity, not orchestration.
