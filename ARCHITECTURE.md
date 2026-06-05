# Fabric — Architecture & First Principles

This document states the design philosophy of the platform built on the Fabric
primitives. It is the *why* behind the model; the YAML primitives are the *what*.

> These principles were derived deliberately. They are not defaults — each one
> is a design decision with a cost the alternative would have incurred.

---

## 0. The law: everything important is a State

**If it matters, it is a `State` in the graph. If it is not in the graph, it
does not officially matter.**

Everything else follows from this:

- The core is fundamentally a **state store** (the graph / SurrealDB).
- The *only* way importance changes is an **`Event`** — `Event advances State`.
- **Memory** is the recorded states and events the runtime can read back.
- A **`Journey`** is the emergent *trace* of states — descriptive, never prescribed.
- Importance ⇒ observable, persisted, governed — because it is a State, not
  hidden in some service's RAM.

`State is real.`

---

## 1. The core is the graph

```
CORE (protocol-agnostic)
  Runtime = the graph (SurrealDB substrate)
          + Identity        (who acts / is acted upon)
          + Memory          (recorded Events / State, read back)
          + Agents          (which act on the graph)
  Runtime delivers an SLA   (guaranteed Objectives: availability, latency, durability)
```

- **SurrealDB is the substrate** — the one non-tool. Everything else is a tool.
- The core **never embeds a wire protocol**. It speaks graph, not HTTP.
- The core is a small, *stable* kernel (high fan-in, low fan-out) — which, in
  Martin's dependency metrics, is exactly what keeps architectural debt low.

---

## 2. Everything else is a tool; protocols are sidecars

```
        ┌──────────────── SIDECARS (protocols, = agents) ───────────────┐
        │  https · grpc · mcp · kafka · sql · custom …                   │
        │  each binds/translates a Protocol at the edge                  │
        └───────────────────────────┬────────────────────────────────────┘
                                     │  Touchpoints (speak a Protocol)
        ┌─────────────────────────────▼──────────────────────────────────┐
        │  CORE = the graph + identity + memory + agents                  │
        └───────────────────────────────────────────────────────────────────┘
```

- **Protocols are sidecars, never core.** Adding/changing a protocol means
  deploying an adapter — it must never require changing core code.
- **Sidecars *are* agents.** A protocol adapter is an Agent that guards a
  Touchpoint and does the binding. The edge is all agents; the core is the graph.

---

## 3. Touchpoints & protocols (the boundaries)

- A **`Touchpoint`** is a junction where two or more things interact. It is
  **only needed at a boundary** — where the things differ in *type* or
  *surface/protocol*. Same type + same surface on both sides is **seamless**
  (e.g. agent↔agent over **A2A** is *not* a touchpoint).
- Every touchpoint declares a **surface** and a **protocol**. If no protocol
  exists for a boundary, a **`Protocol` is defined** (a reusable asset).
- A boundary with **no covering touchpoint is a design gap** — it must be
  reported: *"the system design does not cover that touchpoint."*
- The **agent language is the binding language** that translates across
  heterogeneous touchpoints (agent↔tool uses **MCP**).
- **The binding language must be ambiguity-free.** Every term resolves to
  exactly one meaning — grounded in the graph's IRIs (each attribute carries a
  `schema:sameAs` mapping in the JSON-LD export). No term is overloaded; meaning
  is the same on both sides of a touchpoint, so translation is deterministic.
  Ambiguity at a boundary *is* a defect, not a tolerance.

---

## 4. Agents are autonomous; the path is emergent

- An agent's path through the state/journey graph is **emergent and
  non-mandatory**. The graph is *possibility space*, not a required flow.
- An agent is bounded **only by Policy / Constraint** — never by a prescribed
  route. Within those bounds it is free.
- This is the opposite of a rigid pipeline: the DAG is optional, the agent decides.
- **An agent's definition depends only on the stable core** (existing State,
  Identity, the graph) — **never on future or not-yet-real things.** You cannot
  define an agent on what doesn't exist yet. Depend toward stability; that is
  what keeps the agent layer free of the Unstable Dependency smell.

---

## 5. Debt discipline (why this shape)

Per architectural-debt research, the dangerous smells are **Cyclic**,
**Hub-Like**, and **Unstable** dependencies — at the *component* layer.

- **Semantic cycles are fine.** The model is an ontology; `Policy ↔ Constraint ↔
  Objective` referencing each other is *meaning*, not a dependency cycle.
- **The component layer stays acyclic.** "Protocols are sidecars" makes
  dependencies flow one way (sidecar → core), structurally preventing the
  core↔adapter cycle — the #1 smell.
- **The core avoids being a hub/god-object** because behaviour lives at the
  edge in many small agents, not in the kernel.

So the sidecar/core split isn't aesthetic — it's the ATD-minimizing decision.

---

## 6. The product: the Agent Composer

The platform ships as the **Agent Composer**. It has one hard admission rule:

> **Inside the composer, only the agent language is allowed.**

- The composer is a **closed world**: a single, ambiguity-free language is the
  only thing spoken inside it. One term, one meaning — always.
- Anything from outside enters **only through a Touchpoint**, where a Protocol
  sidecar (an agent) **translates it into the agent language**. Foreign
  protocols, formats, and dialects stop at the edge; they never leak inward.
- This is what makes composition *safe to sell*: because the interior speaks
  one unambiguous language over the State graph, composed agents are
  deterministic and verifiable. Ambiguity can only exist at a boundary, and the
  boundary is exactly where a Protocol is defined to remove it.

In short: **the edge translates; the core composes — in one language only.**

## 7. Why this sells: every enterprise product reduces to three things

**Every enterprise product = a data model + states + control flows.**

That's the whole thesis. Strip any enterprise SaaS down and you find exactly:

| The three parts | Fabric provides | Example (the identity stack) |
|---|---|---|
| **Data model** | primitives + `Schema`/`FieldGroup`/`DataType` → SurrealDB schema, JSON-LD | Users, Accounts, Entitlements, Safes |
| **States** | `State` (everything important is a State) + `Event` advances it | lifecycle (joiner/leaver), JIT "elevated", "last-rotated" |
| **Control flows** | `Policy`·`Constraint`·`Capability`·`Objective`·`Journey`·`Agent` | approvals, certifications, rotation, SOD enforcement |

We *demonstrated* this: **Okta, SailPoint, and CyberArk** — three very different
enterprise products — each decomposed cleanly into (data model → primitives),
(states → `State`), (control flows → `Policy`/`Agent`/`Journey`). See
`docs/*-data-model-mapping.md`.

So the platform doesn't model *one* product — it models the **substrate every
enterprise product is built from.** A new product is: load its data model, name
its states, wire its control flows as governed agent paths. That is what the
**Agent Composer** sells.

## 8. What makes this different: AI-native · multi-model · graph

Cortex, and the incumbents we mapped (Okta/SailPoint/CyberArk), are
service-oriented, relational-backed systems with AI bolted *on*. This is the
inversion:

- **AI-native** — agents aren't a feature, they're the **operating model**. The
  core runtime *is* graph + identity + memory + **agents**; sidecars *are*
  agents; the product is an **Agent Composer**; paths are emergent, not scripted.
  Incumbents add an "AI assistant" beside a fixed app; here the agents *are* the
  app. Tellingly, incumbents **still make humans fill forms** to enter data —
  the symptom of a single-model system that can't take rich input. The fabric
  ingests multi-model input and **agents populate the `State`** — the form is
  obsolete.
- **Multi-model** — one **SurrealDB** substrate is graph + document +
  relational (+ time-series + vector) at once. No stitching three databases
  together; the data model, the state store, and the relationships live in one
  place. This also means **multi-model *input*** — the fabric ingests graph,
  document, relational, and event/vector data as first-class. Incumbents froze
  on a single relational model years ago and never innovated past it; they
  can't take multi-model input, so it gets flattened or dropped at the door.
- **Graph** — the core **is** a graph. Everything important is a `State`
  *node*; control flows are *edges*; identity resolution is *traversal*.
  Relationships are first-class, not foreign keys bolted across tables.

These three reinforce each other: a **graph** substrate that's **multi-model**
is the only thing an **AI-native** agent layer can reason and act over uniformly
— one ambiguity-free language, one system of record, agents all the way to the
edge.

> The incumbents are products on a database. This is a graph an AI operates.

### Where incumbents get crippled

No incumbent supports all three features — so each breaks at the use cases that
*need* them:

| Use case | Why incumbents break | Why the fabric doesn't |
|---|---|---|
| **Deep identity resolution** ("all access for this person across N providers, M hops") | relational join depth explodes; data siloed per product | native graph **traversal** over one substrate |
| **Cross-domain correlation** (authn × entitlements × privileged secrets × behavior in one answer) | three separate systems, no shared model | **multi-model** graph holds it all as one |
| **Blast-radius / impact analysis** ("if this entitlement changes, what's affected?") | unbounded reachability is impractical in SQL | graph **reachability** is the natural query |
| **Open-ended remediation** ("figure it out and fix it") | only fixed playbooks/workflows | **AI-native** agents, emergent paths |
| **Semantic integration across vendors** | schema mismatch, overloaded terms | one **ambiguity-free** language, `sameAs` IRIs |

The pattern: the moment a use case needs *relationships, more than one data
model, or open-ended agency*, a service-on-a-relational-DB is crippled — it can
bolt on an assistant, but it cannot become a graph an AI operates.

> Honest caveat: incumbents are mature and battle-tested at their core use
> cases; this fabric is a coherent model + agents (on PR), not yet a running
> product. The advantage is structural, not yet proven in production.

## In one breath

> The core is the graph. Everything important is a State; Events advance it;
> Memory is what was recorded. Identity and Agents are core; everything else is
> a tool. Protocols live in sidecars (which are agents) at Touchpoints on the
> boundaries; define one where none exists. Agents are autonomous — the path is
> emergent, bounded only by governance. The runtime delivers an SLA.
