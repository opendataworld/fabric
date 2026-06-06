# Fabric — Foundation Primitive Layer

> **Status: `v0.2.0-alpha`** — the model, generators, API, identity connectors,
> a runnable Go graph runtime (GraphQL + MCP), and product-composition tooling
> (the Product MVP Builder) are real and runnable; persistent storage and full
> governance execution are still maturing. See [CHANGELOG](CHANGELOG.md).

Fabric is the canonical semantic foundation for the OpenDataWorld platform:
an open, schema.org-aligned model from which schemas, entities, knowledge,
identities, governance, datasets, agents, and publishing are all composed.

Everything is an asset. Everything is governed. Everything is connected
through a canonical model.

New in **v0.2.0-alpha**: a runnable Go graph runtime (GraphQL + MCP) and the
[Product MVP Builder](site/build.html).

## Root primitives

Each primitive answers one fundamental question and is defined as a rich,
schema.org-aligned YAML model conforming to
[`meta/primitive-schema.yaml`](meta/primitive-schema.yaml).

| Primitive | Question | Model |
|-----------|----------|-------|
| **Thing** | What is it? | [`things/thing-model.yaml`](things/thing-model.yaml) |
| **Identity** | Who or what? | [`identities/identity-model.yaml`](identities/identity-model.yaml) |
| **Time** | When? | [`time/time-model.yaml`](time/time-model.yaml) |
| **Location** | Where? | [`locations/location-model.yaml`](locations/location-model.yaml) |
| **Relationship** | How connected? | [`relationships/relationship-model.yaml`](relationships/relationship-model.yaml) |
| **State** | Current condition? | [`states/state-model.yaml`](states/state-model.yaml) |
| **Event** | What happened? | [`events/event-model.yaml`](events/event-model.yaml) |
| **Capability** | What can be done? | [`capabilities/capability-model.yaml`](capabilities/capability-model.yaml) |
| **Constraint** | What is limited? | [`constraints/constraint-model.yaml`](constraints/constraint-model.yaml) |
| **Policy** | What is allowed? | [`policies/policy-model.yaml`](policies/policy-model.yaml) |
| **Objective** | Why? | [`objectives/objective-model.yaml`](objectives/objective-model.yaml) |
| **Resource** | What is consumed? | [`resources/resource-model.yaml`](resources/resource-model.yaml) |
| **Evidence** | What proves it? | [`evidence/evidence-model.yaml`](evidence/evidence-model.yaml) |
| **Risk** | What could go wrong? | [`risks/risk-model.yaml`](risks/risk-model.yaml) |

**Extended nodes** (grounded in the identity integrations and XDM alignment):
`Account` · `Source` · `Credential` · `Session` · `Device` (identity graph) and
`Schema` · `FieldGroup` · `DataType` (XDM-style composition).

**Domain nodes** (the platform describes itself — *everything is data*):
`Dataset` · `Product` · `Solution` · `Feature` · `Market` · `Connector` ·
`Pipeline` · `Agent` · `Journey` · `Metric` · `Tenant` · `Consent` · `Control`.

**IAM nodes** (generalizing Keycloak & Casdoor): `Role` · `Permission` ·
`Group` · `Application`.

The model forms a clean graph — **42 nodes, 120 edges, 0 dangling** (verified by
`python tools/fabric.py validate`).

## What you can run today

- **Data Model API** — `python api/server.py` (schema registry over the primitives).
- **Code generator** — `python codegen/generate_models.py --target all` (Python,
  TypeScript, JSON Schema, SQL, and SurrealDB graph schema).
- **Fabric Runtime** — `runtime/` Go multi-model graph engine with a GraphQL API
  and MCP server (`runtime --mcp`); agents are first-class governed actors.
- **Product MVP Builder** — [`site/build.html`](site/build.html): compose a
  `Product` from the live model and export aligned YAML + a catalog snippet.
- **Liferay on Kubernetes** — [`deploy/liferay/`](deploy/liferay/): a starter
  Helm chart realizing the matching MVP preset.

## How the primitives compose

```
Identity ── holds ──▶ Capability ── serves ──▶ Objective
   │                      │                        ▲
subjectTo             constrainedBy            boundedBy
   ▼                      ▼                        │
 Policy ◀── composedOf ── Constraint ─────────────┘

Event ──advances──▶ State        (a sequence of States over Time = a Journey)
Time / Location ──▶ scope temporal & spatial Constraints
```

Key design realizations carried over from the design notes:

- **Constraint is deeper than Policy.** A Policy is a governed *collection* of
  Constraints; Controls, Validation, Eligibility, Approval Gates, Boundaries,
  Budgets, and Rate Limits are all specializations of Constraint.
- **Capability answers "what can be done"; Objective answers "why."** Together
  with the who/when/where/how/what primitives they form a complete semantic
  operating system rather than a loose collection of schemas.

## Root layer status

The root layer is complete, and the graph has been extended with identity,
composition, and domain nodes so the platform is self-describing — **42 nodes total**.

## Layout

```
fabric/
├── meta/            # primitive meta-schema (schema-of-schemas)
├── things/
├── identities/
├── time/
├── locations/
├── relationships/
├── states/
├── events/
├── capabilities/
├── constraints/
├── policies/
├── objectives/
├── api/             # Data Model API (Python, REST)
├── fabric-db/       # SurrealDB runtime scaffold (Rust, design-stage)
└── runtime/         # data-fabric runtime (Go): multi-model graph, GraphQL
```

## Runtime

[`runtime/`](runtime) is a runnable Go data-fabric runtime: an in-memory
multi-model graph (documents + typed edges) seeded self-describingly from
`gen/model.json` and served over **GraphQL**. See
[`runtime/README.md`](runtime/README.md). `cd runtime && go run .` then open
`http://localhost:8088/`.
