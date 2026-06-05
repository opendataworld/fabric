# Proof: cross-vendor resolution in one traversal

The use case that demonstrates the gap — and the go-to-market in one picture.

## The claim

> "Show me everything a person can touch — their **Okta** accounts, their
> **SailPoint** entitlements/roles, *and* their **CyberArk** privileged secrets
> — in one query, over one model."

An incumbent can't answer this in one shot: it's three separate products, three
data models, and a relational join-depth wall between them. On the fabric it's
**one graph traversal.**

## The demo

Run [`../examples/cross-vendor-resolution/demo.surql`](../examples/cross-vendor-resolution/demo.surql)
on any SurrealDB:

```surql
SELECT
    displayName,
    ->same_as->account.subject                AS okta_and_entra_accounts,
    ->has_role->role->grants->permission.name AS sailpoint_access,
    ->can_use->credential.name                AS cyberark_secrets
FROM identity:alice;
```

```text
displayName              : 'Alice Reyes'
okta_and_entra_accounts  : ['alice@corp.com', 'alice.reyes@corp']   ← authn (Okta/Entra)
sailpoint_access         : ['AD Employees', 'AD Developers']        ← governance (SailPoint)
cyberark_secrets         : ['prod-db root']                         ← privileged (CyberArk)
```

One identity, three vendors, **one traversal** — because they're all nodes in
one multi-model graph.

> Note: the query is written for SurrealDB v2 graph syntax. It has **not** been
> executed in this build environment (no SurrealDB available here); run it on a
> SurrealDB instance to see the result.

## The strategy this encodes: fit in, fill the gaps

This is **not rip-and-replace.** The incumbents stay exactly where they are —
they're the `Source`s behind `Touchpoint`s:

```
   Okta          SailPoint        CyberArk          (incumbents — they stay)
   (authn)       (governance)     (privileged)
     │ Touchpoint   │ Touchpoint     │ Touchpoint
     ▼              ▼                ▼
  ┌───────────────────────────────────────────────┐
  │   FABRIC  — the graph / AI-native layer         │   (we fit in here)
  │   resolves across them; fills the gaps          │
  │   they structurally can't: one identity,        │
  │   one traversal, agentic remediation            │
  └───────────────────────────────────────────────┘
```

- **We don't fight their experience** — we sit *on top of* it and consume it.
- **We fill the gap** — the cross-vendor, multi-model, agent-driven questions
  no single-model product can answer.
- **When they respond** (build it, or acquire a startup to bolt it on), the
  edge holds: graph + multi-model + AI-native has to be the *substrate*, not a
  feature acquired and stitched onto a relational core. You can't buy your way
  out of an architecture.

We don't play defense. We integrate — and own the layer they can't become.

## The wedge: step reduction

The moment the customer adopts you isn't a feature comparison — it's friction.
They run the incumbent flow and count the steps:

```
INCUMBENT (today):  form  →  request (SailPoint)  →  approval  →
                    provision (Okta)  →  check-out secret (CyberArk)  →
                    verify  →  close ticket            ≈ 7 manual steps, 3 systems

FABRIC (agent over the graph):   "grant Alice prod-db access"
                                 → agent walks the graph, applies Policy,
                                   advances State            ≈ 1 step
```

Then they ask the question that sells it: **"why are there so many steps — can
I reduce this?"**

Yes — because every enterprise product is *data model + states + **control
flows***, and the control flow is exactly what collapses when an **agent
traverses the graph** instead of a human walking three systems and filling
forms. Each step an incumbent makes a human do is a step the fabric makes a
State transition an agent performs under Policy.

- Incumbents *can't* reduce the steps: the steps exist **because** the data is
  siloed across products and entered by hand. The steps are the architecture.
- The fabric reduces them to ~1 because the systems are one graph and the work
  is agentic.

**Land on step reduction, expand to the whole control flow.** That's the wedge.

## Optimization: steps come down gradually

Step reduction isn't a one-time switch — it's **continuous optimization.** The
fabric records every control flow as States and Events, so it can see where the
steps are and collapse them over time: a 7-step flow becomes 5, then 3, then 1
as more of it is expressed as agent-traversable Policy over the graph. The graph
*is* the telemetry; optimization is reading it and cutting.

## These are hypotheses, not claims

Everything above (the wedge, the moat, "incumbents can't reduce steps",
gradual optimization) is a **hypothesis to validate** — not a proven fact. State
them so they can be falsified:

| # | Hypothesis | How we'd test it (and what kills it) |
|---|---|---|
| H1 | Customers feel acute step-friction in cross-vendor identity flows | Interviews / time-on-task. *Killed if* flows are already ~1-step for them. |
| H2 | Agent-over-graph cuts a ~7-step flow toward ~1 | Build it, measure steps before/after. *Killed if* the agent adds as many steps (governance, exceptions) as it removes. |
| H3 | Incumbents *can't* close the gap quickly | Watch their roadmaps/acquisitions. *Killed if* one ships graph-native multi-model + agentic in a release. |
| H4 | Step reduction compounds via optimization | Longitudinal step-count per flow. *Killed if* steps plateau (governance/edge-cases floor it). |
| H5 | "Fit in, fill gaps" beats rip-and-replace as GTM | Win-rate of layer-on-top vs replacement deals. |

The architecture (graph + multi-model + AI-native) is real and on PR #3. The
**business** around it is hypotheses — and we should treat them as such:
instrument, measure, and let the evidence move them, the same way the paper we
read found ATD dropping at "fixes/improvements" only *after* measuring.


