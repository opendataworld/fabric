# CyberArk (PAM) → Fabric Primitives

How CyberArk's Privileged Access Management model maps onto Fabric primitives —
the third layer of the identity-security stack:

| Layer | Provider | Question | Doc |
|---|---|---|---|
| Authentication / IdP | Okta | *Who are you?* | `okta-data-model-mapping.md` |
| Governance / IGA | SailPoint ISC | *What may you have, is it governed?* | `sailpoint-data-model-mapping.md` |
| **Privileged Access / Secrets** | **CyberArk** | ***Privileged access — vaulted, rotated, brokered?*** | **this doc** |

> CyberArk's page (cyberark.com/why-identity-security) was egress-blocked from
> the build environment, so this reflects the well-known CyberArk PAM model.

## Object mapping

| CyberArk object | Fabric primitive | Notes |
|---|---|---|
| **Vault / Safe** | `Resource` (secured container) + `Policy` | A governed collection of secrets with an access policy. |
| **Privileged Account** | `Account` (kind: privileged) on a `Source` | High-privilege account on a target system. |
| **Secret / Password / Key** | `Credential` | The vaulted secret itself. |
| **Credential rotation** | `Event` → advances `State` (under a rotation `Policy`) | Each rotation is an event; "last-rotated" is State. |
| **CPM (Central Policy Manager)** | `Agent` (rotation agent) + `Capability` (rotate) | The thing that rotates credentials on policy. |
| **PSM (Privileged Session Manager)** | `Agent` + `Touchpoint` | Brokers/records the privileged-session boundary. |
| **Privileged Session (recorded)** | `Session` + `Evidence` | The session, plus its recording as proof. |
| **Just-In-Time (JIT) elevation** | time-boxed `Permission`/`Capability` via `Event`; `State` = "elevated"; `Constraint` = TTL | Elevation *is a State*, bounded by a time Constraint. |
| **Access / Rotation Policy** | `Policy` + `Constraint` | Who may retrieve; how often to rotate. |
| **PTA (Privileged Threat Analytics)** | `Risk` + Security `Agent` + `Event` | Detected privileged-access threats. |
| **Target system / connection** | `Source` + `Touchpoint`→`Protocol` | SSH/RDP/DB boundary the secret unlocks. |

## Relationships (CyberArk → Fabric edges)

```
Safe (Resource)        ──contains──▶  Credential(s)
Privileged Account     ──on──▶  Source
Credential             ──unlocks──▶  Touchpoint (SSH/RDP/DB)
CPM Agent              ──rotates (Event)──▶  Credential   (advances State)
PSM Agent              ──guards──▶  Touchpoint (privileged session)
JIT grant (Event)      ──advances──▶  State "elevated"  (bounded by TTL Constraint)
Access Policy          ──governs──▶  Safe / Account
```

## Agents this adds to the roster

- **Rotation Agent** (CPM): executes `rotate`, pursues "no stale credential",
  governed by a rotation `Policy`; every rotation is an `Event` advancing State.
- **Session-Broker Agent** (PSM): guards the privileged-session `Touchpoint`,
  records the `Session` as `Evidence`, enforces least-privilege.

Both are the same shape as the Edge Guardian / Security / Protocol agents:
defined only on the stable core, guarding a Touchpoint, path emergent, bounded
by Policy/Constraint.

## The convergence

A single resolved **`Identity`** in the graph now carries:
- its **accounts** across IdPs (Okta),
- its **entitlements / roles / lifecycle State** (SailPoint),
- its **privileged access & vaulted credentials** (CyberArk).

…all as **State**, all **governed**, all reached through **Touchpoints speaking
Protocols** bound by sidecar agents. The three vendors are `Source`s; the fabric
is the system of record.
