# The Agent Marketplace — platform, tools, and protocols

A federated, community-run venue where **agents** (and people, and businesses)
**list, discover, and transact**. It is built the Fabric way: *a new product is
its data model + its states + its control flows, wired as governed agent paths.*
Nothing here is a bespoke app — the marketplace is **configuration over the
graph**.

- **Model:** [`marketplaces/`](../marketplaces/marketplace-model.yaml) ·
  [`operating-models/`](../operating-models/operating-model-model.yaml) ·
  [`offers/`](../offers/offer-model.yaml) · [`orders/`](../orders/order-model.yaml) ·
  [`payments/`](../payments/payment-model.yaml)
- **Runnable instance:** [`examples/agent-marketplace/fabric.yaml`](../examples/agent-marketplace/fabric.yaml)

```bash
python tools/fabric.py validate --instance examples/agent-marketplace/fabric.yaml
python tools/fabric.py create   examples/agent-marketplace/fabric.yaml   # -> SurrealQL (connect all nodes)
```

## The platform

A **`Marketplace`** is the venue (the platform itself). It `hosts` Offers,
`enables` a configurable set of Operating Models, `federatesVia` Matrix,
`transactsVia` the commerce + payment protocols, and `exposes` many Touchpoints.
The marketplace is `operatedBy` a Tenant and `governedBy` trust-and-safety Policy.

## Configurable operating models (multiple selling motions)

A **`OperatingModel`** is one selling motion, defined as a *value*, not code:

| Motion | Seller → Buyer | Example offer |
|---|---|---|
| **b2b**   | organization → organization | Acme sells "Summarizer Pro" API to another business |
| **b2c**   | organization → person       | a vendor sells a subscription to a consumer |
| **c2c**   | person → person             | an indie maker sells human translation |
| **d2c**   | organization → person       | a brand sells directly, no intermediary |
| **b2b2c** | organization → (business) → person | a platform reseller fronts a consumer |
| **p2p**   | agent → agent               | Seller Agent Y sells a skill pack to Buyer Agent X |

A marketplace **enables a subset** of these. Switching a motion on or off is
flipping `enabled` on its asset — never a code change. The motion fixes the
seller/buyer Identity *kinds* at the boundary, which is exactly how one platform
serves many models at once. Add a new motion by defining one more
`OperatingModel` asset; the core never changes.

## The protocols (sidecars at Touchpoints)

Per the architecture, **protocols are sidecars, never core**; each is realized by
an Agent that guards a Touchpoint and translates at the boundary.

| Protocol | Role | Carried by |
|---|---|---|
| **Matrix** | federation across community instances | `Federation Agent` @ `touchpoint:matrix` |
| **Agent Commerce Protocol (ACP)** | listings & orders (`Offer` → `Order`) | `Discovery` / `Commerce` agents @ api / a2a / mcp |
| **Agent Payments Protocol (AP2)** | mandate-authorized settlement (`Payment`) | `Settlement` / `Escrow` agents @ `touchpoint:pay-edge` |
| **MCP** | agent ↔ tool | `touchpoint:tool-mcp` |
| **A2A** | agent ↔ agent (seamless) | `touchpoint:agent-a2a` |

**Commerce flow (ACP):** a seller's `Offer` is discovered by a buyer agent and
turned into an `Order` (the commitment). **Settlement (AP2):** the `Order` is
settled by a `Payment` that MUST be `authorizedBy` a mandate (`Consent`),
`securedBy` a `Credential`, and that `draws` on a budget `Resource` — no mandate,
no capture; over-budget is blocked.

## Multi-channel, multi-surface

The venue is reachable on many `Touchpoint`s, each declaring a `surface` and a
`protocol`:

`ui` (web) · `api` (REST/GraphQL) · `cli` · `agent` (A2A) · `tool` (MCP) ·
`webhook` (settlement callbacks) · `event` (Matrix federation).

Adding a channel = adding a Touchpoint. The core stays protocol-agnostic; the
binding lives in the sidecar agent that guards the edge.

## Control theory: every operational agent is a closed-loop controller

The marketplace doesn't run scripted pipelines — it runs **governed feedback
loops**. The Fabric primitives map onto classical control exactly:

| Control concept | Fabric primitive | In the instance |
|---|---|---|
| **Setpoint / reference** | `Objective` | `objective:settlement` (settle within SLA, under mandate, within budget) |
| **Measured variable / sensor** | `Metric` | `metric:settle-latency`, `metric:settle-success` |
| **Controller** | `Agent` applying `Policy` | `agent:settlement` under `policy:settlement-control` |
| **Controller bounds** | `Constraint` | `constraint:settle-sla`, `constraint:budget-bound`, `constraint:mandate-required` |
| **Actuator** | `Capability` | `capability:settle`, `capability:escrow-release` |
| **Plant + dynamics** | `Order`/`Payment` `State`, advanced by `Event` | `state:order-accepted` → `event:payment-settled` → `state:settled` |
| **Feedback** | `Event` updates `Metric` → controller re-acts | latency/success feed the next control action |

Four loops run concurrently:

1. **Liquidity / pricing loop** — `agent:discovery` drives `metric:fill-rate`
   toward `objective:liquidity`, bounded by `constraint:price-floor`.
2. **Fulfillment loop** — `agent:commerce` drives Orders to `fulfilled`.
3. **Settlement loop** — `agent:settlement` drives Payments to `settled` within
   SLA, with `agent:escrow` holding funds until release conditions clear.
4. **Trust loop** — `agent:arbiter` drives `metric:dispute-rate` /
   `metric:seller-rating` toward `objective:trust`, actuating refunds and
   listing suspensions, bounded by `constraint:reputation-floor`.

The path through each loop is **emergent** (architecture §4): the agent is bounded
only by Policy/Constraint, never by a prescribed route — it is error-correcting
control, not a fixed DAG.

## Trust, reputation & disputes (community / c2c / p2p)

Community selling lives or dies on trust, so the marketplace carries an explicit
trust layer — built entirely from existing primitives, no new node types:

- **`Risk`** — what can go wrong (`non-delivery`, `fraud`, `chargeback`), each
  `threatens` a transaction and is `mitigatedBy` a Constraint.
- **`Control`** — an enforced safeguard: `kyc`, `escrow-hold`, and
  `reputation-gate` each `enforces` a Constraint, `mitigates` a Risk, and is
  `evidencedBy` Evidence.
- **`Evidence`** — receipts, KYC attestations, and reviews make trust *auditable
  rather than asserted* (`derivedFrom` Events, `substantiates` Objectives).
- **`Metric`** — `seller-rating` and `dispute-rate` are the trust loop's sensors.
- **Dispute lifecycle** — `disputed → arbitration → resolved / refunded` as
  States advanced by Events, traced by a **`Journey`** (`journey:dispute`).
- **`Arbiter` agent** — the trust controller: opens disputes, arbitrates,
  refunds, and suspends listings, bounded by `policy:dispute-resolution`.

This is what makes c2c and p2p motions safe: a low-reputation seller is gated by
`constraint:reputation-floor`, escrow holds funds until delivery Evidence
exists, and every dispute outcome is a governed State change, not a support
ticket lost in someone's inbox.

## Why this is safe to sell

Because everything important is a `State` in the graph, advanced only by `Event`,
every listing, order, authorization, capture, and refund is **observable,
persisted, and governed** — not hidden in a service's RAM. The interior speaks
one ambiguity-free language; foreign protocols (Matrix, ACP, AP2) stop at the
Touchpoint where a sidecar translates them in. The edge translates; the core
composes.
