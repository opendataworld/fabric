# Okta Data Model → Fabric Primitives

How Okta's core objects map onto Fabric primitives. This grounds the
**Enterprise Identity Fabric Agent** (`examples/enterprise-identity-fabric`) in a
real IdP data model, and shows the fabric identity nodes already cover it.

> Source: the standard Okta data model
> (developer.okta.com/docs/concepts/okta-data-model). The page wasn't fetchable
> from the build environment (egress-blocked), so this reflects the well-known
> model; adjust if Okta's published concepts have shifted.

## Object mapping

| Okta object | Fabric primitive | Notes |
|---|---|---|
| **Org** | `Tenant` | The top-level container / isolation boundary. |
| **User** | `Identity` (kind: person) | Profile + lifecycle status (STAGED/ACTIVE/SUSPENDED/DEPROVISIONED) → modeled as `State`. |
| **Group** | `Group` | Collections used for access assignment. |
| **Application** | `Application` | Integrated apps (OIDC/SAML/SWA). |
| **Authorization Server** | `Capability` + `Policy` | Issues tokens (a capability) under access policies. |
| **Inbound IdP / federation** | `Source` + `Touchpoint`→`Protocol` | Each external IdP is a boundary speaking OIDC/SAML. |
| **Linked account (per IdP)** | `Account` | The provider-specific identity; `sameAs` → the unified `Identity`. |
| **Session** | `Session` | An authenticated session. |
| **Authenticator / Factor (MFA)** | `Credential` | Auth factors held by the identity. |
| **Device** | `Device` | Registered device. |
| **Policy** | `Policy` | Sign-on / password / MFA-enrollment policy. |
| **Policy Rule** | `Constraint` | A policy is a governed collection of constraints. |
| **Profile / Universal Directory schema** | `Schema` + `FieldGroup` + `DataType` | Profile attribute definitions. |
| **System Log entry** | `Event` | Audit/lifecycle events — and `Event advances State`. |
| **Consent / grant** | `Consent` | OAuth consent / account-linking grant. |

## Key relationships (Okta → Fabric edges)

| Okta relationship | Fabric edge |
|---|---|
| User memberOf Group | `Identity` —memberOf→ `Group` |
| User/Group assigned to Application | `Identity`/`Group` —assignedTo→ `Application` |
| User has Sessions / Factors / Devices | `Identity` → `Session` / `Credential` / `Device` |
| Policy has Rules | `Policy` —composedOf→ `Constraint` |
| Policy applies to Group/App | `Policy` —governs→ `Group`/`Application` |
| Account federates from IdP | `Account` —from→ `Source`; reached via a `Touchpoint` |
| User identity resolved across IdPs | `Account` —sameAs→ unified `Identity` |

## How the Identity Fabric Agent uses this

The **Enterprise Identity Fabric Agent**:

1. **Guards a Touchpoint per IdP** — OIDC, SAML, SCIM — each speaking its own
   `Protocol` (bound by a protocol sidecar into the one agent language).
2. **Reads accounts** (`Account` ← `Source`) and **resolves** them via `sameAs`
   into a single unified `Identity` in the graph.
3. **Governed by consent** — links only happen where a `Consent` grant +
   `Constraint` allow it.
4. Everything important is a **State**: the resolved identity, each linked
   account, each lifecycle transition is persisted and governed in the graph —
   not hidden in the IdP.

So Okta (or Entra, Auth0, Google, …) is just a `Source` behind a `Touchpoint`;
the fabric holds the *unified, governed* truth.
