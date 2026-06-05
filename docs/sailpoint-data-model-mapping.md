# SailPoint Identity Security Cloud (IGA) → Fabric Primitives

How SailPoint ISC's governance objects (v3 API) map onto Fabric primitives.
Where Okta answers *who you are* (authentication/IdP), SailPoint answers *what
you're entitled to, and is it governed* (IGA). Both are just `Source`s behind
`Touchpoint`s; the fabric holds the unified, governed truth as **State**.

## Object mapping

| SailPoint ISC object | Fabric primitive | Notes |
|---|---|---|
| **Identity** | `Identity` | The person/actor. |
| **Identity Profile** | `Schema` + `Policy` | Rules mapping source accounts → identity attributes. |
| **Source** | `Source` (+ `Connector`) | A connected system (e.g. Active Directory). |
| **Connector** | `Connector` | Integration to a source. |
| **Account** | `Account` | An identity's account on a source. |
| **Entitlement** | `Permission` | A single access right on a source. |
| **Access Profile** | `Permission` bundle (≈ source-scoped `Role`) | Groups entitlements on a source. |
| **Role** | `Role` | Groups access profiles into an assignable role. |
| **Lifecycle State** | **`State`** | active / inactive / leaver … — *everything important is a State*. |
| **Access Request** | `Event` + `Journey` | The request flow (emergent path of states). |
| **Access Request Approval** | `Policy` (approval gate) + `Constraint` | Governs whether access is granted. |
| **Requestable Object** | `Capability` | What can be requested. |
| **Certification / Campaign** | `Control` + `Evidence` | Attestation that access is correct. |
| **SOD Policy** | `Policy` + `Constraint` | Separation-of-duties rule over entitlement combinations. |
| **SOD Violation** | `Risk` (+ `Event`/`State`) | A detected breach of an SOD policy. |
| **Segment** | `Group` | Access segmentation. |
| **Workflow** | `Pipeline` / `Journey` | Orchestrated governance flow. |
| **Work Item** | `Event` + `State` | A task with a status. |
| **Transform** | `Capability` at a `Touchpoint` | Attribute transformation = binding at a boundary. |
| **MFA / Authenticators** | `Credential` | Auth factors. |
| **OAuth Clients / Personal Access Tokens** | `Credential` + `Application` | Programmatic access. |
| **Password Policies** | `Policy` + `Constraint` | |
| **Account / Source / Activity Usages** | `Event` + `Metric` | Audit + measurement. |
| **Service Desk Integration** | `Connector` + `Touchpoint` | External ITSM boundary. |

## Governance relationships (SailPoint → Fabric edges)

```
Entitlement (Permission)  ──on──▶  Source
Access Profile            ──groups──▶  Entitlement(s)
Role                      ──groups──▶  Access Profile(s)
Lifecycle State (State)   ──grants──▶  Access Profile(s) / Role(s)
Certification (Control)   ──reviews──▶  Entitlement / Access Profile / Role
SOD Policy (Policy)       ──constrains──▶  Entitlement combinations  (Constraint)
Access Request (Event)    ──advances──▶  State   (under an approval Policy)
```

## How an agent uses this

The **Enterprise Identity Fabric Agent** (and a governance peer) treat SailPoint
as one more `Source` behind a `Touchpoint` (SCIM / ISC v3 API, speaking its
`Protocol`):

1. **Entitlements/Access Profiles/Roles** land as `Permission`/`Role` granted to
   the resolved `Identity` — joining the Okta-side accounts on the same person.
2. **Lifecycle State** is a `State` in the graph: a leaver transition is an
   `Event` that advances State and triggers de-provisioning — governed, observable.
3. **SOD Policies** are `Policy`+`Constraint`; a violation is a `Risk`. The
   agent's path is emergent but **bounded by these** — it can never grant a
   combination an SOD `Constraint` forbids.
4. **Certifications** are `Control`+`Evidence` — the proof that access is correct,
   persisted in the graph.

So authentication (Okta) and governance (SailPoint) converge on **one governed
Identity with one entitlement state** — the fabric is the system of record, the
IdP/IGA are sources behind touchpoints.
