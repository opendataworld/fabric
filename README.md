# Fabric — Foundation Primitive Layer

Fabric is the canonical semantic foundation for the OpenDataWorld platform:
an open, schema.org-aligned model from which schemas, entities, knowledge,
identities, governance, datasets, agents, and publishing are all composed.

Everything is an asset. Everything is governed. Everything is connected
through a canonical model.

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

## Remaining root gaps (queued)

`Resource` · `Evidence` · `Risk` — to be added next.

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
└── objectives/
```
