-- Cross-vendor identity resolution — the proof.
-- One person. Three vendors (Okta authn · SailPoint IGA · CyberArk PAM).
-- ONE graph traversal. An incumbent needs three products and can't join them.
--
-- Run on any SurrealDB:  surreal sql --ns demo --db fabric < demo.surql
-- (or paste into Surrealist).  Self-contained; uses graph RELATE edges.

-- ── the unified identity ────────────────────────────────────────────────
CREATE identity:alice SET kind = 'person', displayName = 'Alice Reyes';

-- ── Okta (authentication): linked accounts on sources ───────────────────
CREATE source:okta  SET name = 'Okta';
CREATE source:entra SET name = 'Microsoft Entra';
CREATE account:okta_alice  SET source = source:okta,  subject = 'alice@corp.com';
CREATE account:entra_alice SET source = source:entra, subject = 'alice.reyes@corp';
RELATE identity:alice->same_as->account:okta_alice;
RELATE identity:alice->same_as->account:entra_alice;

-- ── SailPoint (governance): entitlements grouped into roles ─────────────
CREATE permission:ad_employees  SET name = 'AD Employees';
CREATE permission:ad_developers SET name = 'AD Developers';
CREATE role:engineer SET name = 'Engineer';
RELATE role:engineer->grants->permission:ad_employees;
RELATE role:engineer->grants->permission:ad_developers;
RELATE identity:alice->has_role->role:engineer;

-- ── CyberArk (privileged access): a vaulted credential in a safe ────────
CREATE resource:prod_safe  SET name = 'Prod Safe';
CREATE credential:db_root   SET name = 'prod-db root', kind = 'privileged';
RELATE resource:prod_safe->contains->credential:db_root;
RELATE identity:alice->can_use->credential:db_root;

-- ── THE QUERY: everything Alice can touch, all three vendors, one shot ──
SELECT
    displayName,
    ->same_as->account.subject                       AS okta_and_entra_accounts,
    ->has_role->role->grants->permission.name        AS sailpoint_access,
    ->can_use->credential.name                       AS cyberark_secrets
FROM identity:alice;

-- Expected result (one row):
--   displayName              : 'Alice Reyes'
--   okta_and_entra_accounts  : ['alice@corp.com', 'alice.reyes@corp']
--   sailpoint_access         : ['AD Employees', 'AD Developers']
--   cyberark_secrets         : ['prod-db root']
