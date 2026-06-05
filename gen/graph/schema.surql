-- Graph schema generated from Fabric primitives — DO NOT EDIT BY HAND.
-- Nodes = primitives (fully expanded with fields); edges = RELATION tables.

DEFINE TABLE account SCHEMAFULL;
DEFINE FIELD provider ON TABLE account TYPE string;
DEFINE FIELD provider_sub ON TABLE account TYPE string;
DEFINE FIELD email ON TABLE account TYPE option<string>;
DEFINE FIELD username ON TABLE account TYPE option<string>;

DEFINE TABLE agent SCHEMAFULL;
DEFINE FIELD name ON TABLE agent TYPE string;

DEFINE TABLE application SCHEMAFULL;
DEFINE FIELD name ON TABLE application TYPE string;
DEFINE FIELD clientId ON TABLE application TYPE option<string>;

DEFINE TABLE capability SCHEMAFULL;
DEFINE FIELD name ON TABLE capability TYPE string;
DEFINE FIELD verb ON TABLE capability TYPE string;
DEFINE FIELD inputs ON TABLE capability TYPE option<array<object>>;
DEFINE FIELD outputs ON TABLE capability TYPE option<array<object>>;
DEFINE FIELD requiresTools ON TABLE capability TYPE option<array<record>>;
DEFINE FIELD requiresResources ON TABLE capability TYPE option<array<record>>;
DEFINE FIELD maturity ON TABLE capability TYPE option<string>;

DEFINE TABLE connector SCHEMAFULL;
DEFINE FIELD name ON TABLE connector TYPE string;

DEFINE TABLE consent SCHEMAFULL;
DEFINE FIELD purpose ON TABLE consent TYPE string;
DEFINE FIELD granted ON TABLE consent TYPE option<bool>;

DEFINE TABLE constraint SCHEMAFULL;
DEFINE FIELD kind ON TABLE constraint TYPE string;
DEFINE FIELD expression ON TABLE constraint TYPE string;
DEFINE FIELD target ON TABLE constraint TYPE option<record>;
DEFINE FIELD severity ON TABLE constraint TYPE option<string>;
DEFINE FIELD onViolation ON TABLE constraint TYPE option<string>;

DEFINE TABLE control SCHEMAFULL;
DEFINE FIELD name ON TABLE control TYPE string;

DEFINE TABLE credential SCHEMAFULL;
DEFINE FIELD type ON TABLE credential TYPE string;
DEFINE FIELD claim ON TABLE credential TYPE option<object>;
DEFINE FIELD proof ON TABLE credential TYPE option<string>;

DEFINE TABLE data_type SCHEMAFULL;
DEFINE FIELD name ON TABLE data_type TYPE string;
DEFINE FIELD fields ON TABLE data_type TYPE option<array<object>>;

DEFINE TABLE dataset SCHEMAFULL;
DEFINE FIELD name ON TABLE dataset TYPE string;
DEFINE FIELD format ON TABLE dataset TYPE option<string>;

DEFINE TABLE device SCHEMAFULL;
DEFINE FIELD kind ON TABLE device TYPE string;
DEFINE FIELD fingerprint ON TABLE device TYPE option<string>;

DEFINE TABLE event SCHEMAFULL;
DEFINE FIELD type ON TABLE event TYPE string;
DEFINE FIELD occurredAt ON TABLE event TYPE datetime;
DEFINE FIELD actor ON TABLE event TYPE option<record>;
DEFINE FIELD subject ON TABLE event TYPE option<record>;
DEFINE FIELD location ON TABLE event TYPE option<record>;
DEFINE FIELD payload ON TABLE event TYPE option<object>;

DEFINE TABLE evidence SCHEMAFULL;
DEFINE FIELD kind ON TABLE evidence TYPE string;
DEFINE FIELD claim ON TABLE evidence TYPE string;
DEFINE FIELD source ON TABLE evidence TYPE option<record>;
DEFINE FIELD uri ON TABLE evidence TYPE option<string>;
DEFINE FIELD hash ON TABLE evidence TYPE option<string>;
DEFINE FIELD collectedAt ON TABLE evidence TYPE option<datetime>;

DEFINE TABLE feature SCHEMAFULL;
DEFINE FIELD name ON TABLE feature TYPE string;
DEFINE FIELD status ON TABLE feature TYPE option<string>;

DEFINE TABLE field_group SCHEMAFULL;
DEFINE FIELD name ON TABLE field_group TYPE string;
DEFINE FIELD fields ON TABLE field_group TYPE option<array<object>>;

DEFINE TABLE group SCHEMAFULL;
DEFINE FIELD name ON TABLE group TYPE string;

DEFINE TABLE identity SCHEMAFULL;
DEFINE FIELD kind ON TABLE identity TYPE string;
DEFINE FIELD displayName ON TABLE identity TYPE option<string>;
DEFINE FIELD controller ON TABLE identity TYPE option<record>;
DEFINE FIELD credentials ON TABLE identity TYPE option<array<record>>;

DEFINE TABLE journey SCHEMAFULL;
DEFINE FIELD subject ON TABLE journey TYPE record;

DEFINE TABLE location SCHEMAFULL;
DEFINE FIELD kind ON TABLE location TYPE string;
DEFINE FIELD latitude ON TABLE location TYPE option<number>;
DEFINE FIELD longitude ON TABLE location TYPE option<number>;
DEFINE FIELD geometry ON TABLE location TYPE option<object>;
DEFINE FIELD address ON TABLE location TYPE option<string>;
DEFINE FIELD uri ON TABLE location TYPE option<string>;

DEFINE TABLE market SCHEMAFULL;
DEFINE FIELD name ON TABLE market TYPE string;

DEFINE TABLE metric SCHEMAFULL;
DEFINE FIELD name ON TABLE metric TYPE string;
DEFINE FIELD value ON TABLE metric TYPE option<number>;
DEFINE FIELD unit ON TABLE metric TYPE option<string>;

DEFINE TABLE objective SCHEMAFULL;
DEFINE FIELD name ON TABLE objective TYPE string;
DEFINE FIELD kind ON TABLE objective TYPE string;
DEFINE FIELD statement ON TABLE objective TYPE string;
DEFINE FIELD metrics ON TABLE objective TYPE option<array<object>>;
DEFINE FIELD targetDate ON TABLE objective TYPE option<record>;
DEFINE FIELD priority ON TABLE objective TYPE option<string>;
DEFINE FIELD status ON TABLE objective TYPE option<record>;

DEFINE TABLE permission SCHEMAFULL;
DEFINE FIELD action ON TABLE permission TYPE string;
DEFINE FIELD effect ON TABLE permission TYPE option<string>;

DEFINE TABLE pipeline SCHEMAFULL;
DEFINE FIELD name ON TABLE pipeline TYPE string;
DEFINE FIELD mode ON TABLE pipeline TYPE option<string>;

DEFINE TABLE policy SCHEMAFULL;
DEFINE FIELD name ON TABLE policy TYPE string;
DEFINE FIELD effect ON TABLE policy TYPE string;
DEFINE FIELD combine ON TABLE policy TYPE option<string>;
DEFINE FIELD constraints ON TABLE policy TYPE array<record>;
DEFINE FIELD scope ON TABLE policy TYPE option<object>;
DEFINE FIELD precedence ON TABLE policy TYPE option<int>;
DEFINE FIELD owner ON TABLE policy TYPE option<record>;

DEFINE TABLE product SCHEMAFULL;
DEFINE FIELD name ON TABLE product TYPE string;
DEFINE FIELD status ON TABLE product TYPE option<string>;

DEFINE TABLE protocol SCHEMAFULL;
DEFINE FIELD name ON TABLE protocol TYPE string;
DEFINE FIELD version ON TABLE protocol TYPE option<string>;
DEFINE FIELD format ON TABLE protocol TYPE option<string>;
DEFINE FIELD spec ON TABLE protocol TYPE option<string>;
DEFINE FIELD status ON TABLE protocol TYPE option<string>;

DEFINE TABLE relationship SCHEMAFULL;
DEFINE FIELD predicate ON TABLE relationship TYPE string;
DEFINE FIELD source ON TABLE relationship TYPE record;
DEFINE FIELD target ON TABLE relationship TYPE record;
DEFINE FIELD directed ON TABLE relationship TYPE option<bool>;
DEFINE FIELD validDuring ON TABLE relationship TYPE option<record>;
DEFINE FIELD weight ON TABLE relationship TYPE option<number>;

DEFINE TABLE resource SCHEMAFULL;
DEFINE FIELD kind ON TABLE resource TYPE string;
DEFINE FIELD unit ON TABLE resource TYPE string;
DEFINE FIELD capacity ON TABLE resource TYPE option<number>;
DEFINE FIELD consumed ON TABLE resource TYPE option<number>;
DEFINE FIELD owner ON TABLE resource TYPE option<record>;

DEFINE TABLE risk SCHEMAFULL;
DEFINE FIELD category ON TABLE risk TYPE string;
DEFINE FIELD statement ON TABLE risk TYPE string;
DEFINE FIELD likelihood ON TABLE risk TYPE option<string>;
DEFINE FIELD impact ON TABLE risk TYPE option<string>;
DEFINE FIELD severity ON TABLE risk TYPE option<string>;
DEFINE FIELD status ON TABLE risk TYPE option<record>;

DEFINE TABLE role SCHEMAFULL;
DEFINE FIELD name ON TABLE role TYPE string;

DEFINE TABLE runtime SCHEMAFULL;
DEFINE FIELD name ON TABLE runtime TYPE string;
DEFINE FIELD version ON TABLE runtime TYPE option<string>;

DEFINE TABLE schema SCHEMAFULL;
DEFINE FIELD name ON TABLE schema TYPE string;
DEFINE FIELD baseClass ON TABLE schema TYPE record;

DEFINE TABLE session SCHEMAFULL;
DEFINE FIELD started ON TABLE session TYPE option<datetime>;
DEFINE FIELD ip ON TABLE session TYPE option<string>;

DEFINE TABLE solution SCHEMAFULL;
DEFINE FIELD name ON TABLE solution TYPE string;

DEFINE TABLE source SCHEMAFULL;
DEFINE FIELD kind ON TABLE source TYPE string;
DEFINE FIELD name ON TABLE source TYPE string;
DEFINE FIELD endpoint ON TABLE source TYPE option<string>;

DEFINE TABLE state SCHEMAFULL;
DEFINE FIELD subject ON TABLE state TYPE record;
DEFINE FIELD value ON TABLE state TYPE string;
DEFINE FIELD lifecycle ON TABLE state TYPE option<string>;
DEFINE FIELD since ON TABLE state TYPE option<datetime>;
DEFINE FIELD allowedTransitions ON TABLE state TYPE option<array<string>>;

DEFINE TABLE tenant SCHEMAFULL;
DEFINE FIELD name ON TABLE tenant TYPE string;

DEFINE TABLE thing SCHEMAFULL;
DEFINE FIELD type ON TABLE thing TYPE string;
DEFINE FIELD name ON TABLE thing TYPE option<string>;
DEFINE FIELD description ON TABLE thing TYPE option<string>;
DEFINE FIELD createdAt ON TABLE thing TYPE option<datetime>;
DEFINE FIELD metadata ON TABLE thing TYPE option<object>;

DEFINE TABLE time SCHEMAFULL;
DEFINE FIELD kind ON TABLE time TYPE string;
DEFINE FIELD instant ON TABLE time TYPE option<datetime>;
DEFINE FIELD start ON TABLE time TYPE option<datetime>;
DEFINE FIELD end ON TABLE time TYPE option<datetime>;
DEFINE FIELD duration ON TABLE time TYPE option<string>;
DEFINE FIELD timezone ON TABLE time TYPE option<string>;

DEFINE TABLE touchpoint SCHEMAFULL;
DEFINE FIELD name ON TABLE touchpoint TYPE string;
DEFINE FIELD surface ON TABLE touchpoint TYPE string;
DEFINE FIELD protocol ON TABLE touchpoint TYPE string;
DEFINE FIELD format ON TABLE touchpoint TYPE option<string>;
DEFINE FIELD direction ON TABLE touchpoint TYPE option<string>;
DEFINE FIELD endpoint ON TABLE touchpoint TYPE option<string>;
DEFINE FIELD protocolVersion ON TABLE touchpoint TYPE option<string>;

-- Relationship edges
DEFINE TABLE ownedby TYPE RELATION IN account OUT identity;
DEFINE TABLE fromsource TYPE RELATION IN account OUT source;
DEFINE TABLE linkedto TYPE RELATION IN account OUT account;
DEFINE TABLE hassession TYPE RELATION IN account OUT session;
DEFINE TABLE executes TYPE RELATION IN agent OUT capability;
DEFINE TABLE pursues TYPE RELATION IN agent OUT objective;
DEFINE TABLE governedby TYPE RELATION IN agent OUT policy;
DEFINE TABLE hasmemory TYPE RELATION IN agent OUT event;
DEFINE TABLE guards TYPE RELATION IN agent OUT touchpoint;
DEFINE TABLE authenticates TYPE RELATION IN application OUT account;
DEFINE TABLE registeredin TYPE RELATION IN application OUT tenant;
DEFINE TABLE governedby TYPE RELATION IN application OUT policy;
DEFINE TABLE heldby TYPE RELATION IN capability OUT identity;
DEFINE TABLE constrainedby TYPE RELATION IN capability OUT constraint;
DEFINE TABLE governedby TYPE RELATION IN capability OUT policy;
DEFINE TABLE servesobjective TYPE RELATION IN capability OUT objective;
DEFINE TABLE integrates TYPE RELATION IN connector OUT source;
DEFINE TABLE implements TYPE RELATION IN connector OUT capability;
DEFINE TABLE grantedby TYPE RELATION IN consent OUT identity;
DEFINE TABLE coversdata TYPE RELATION IN consent OUT dataset;
DEFINE TABLE enforcedby TYPE RELATION IN consent OUT policy;
DEFINE TABLE composesinto TYPE RELATION IN constraint OUT policy;
DEFINE TABLE limits TYPE RELATION IN constraint OUT capability;
DEFINE TABLE enforces TYPE RELATION IN control OUT constraint;
DEFINE TABLE mitigates TYPE RELATION IN control OUT risk;
DEFINE TABLE evidencedby TYPE RELATION IN control OUT evidence;
DEFINE TABLE assertedabout TYPE RELATION IN credential OUT identity;
DEFINE TABLE issuedby TYPE RELATION IN credential OUT identity;
DEFINE TABLE evidencedby TYPE RELATION IN credential OUT evidence;
DEFINE TABLE usedby TYPE RELATION IN data_type OUT field_group;
DEFINE TABLE conformsto TYPE RELATION IN dataset OUT schema;
DEFINE TABLE derivedfrom TYPE RELATION IN dataset OUT source;
DEFINE TABLE governedby TYPE RELATION IN dataset OUT policy;
DEFINE TABLE hasstate TYPE RELATION IN dataset OUT state;
DEFINE TABLE usedby TYPE RELATION IN device OUT identity;
DEFINE TABLE locatedat TYPE RELATION IN device OUT location;
DEFINE TABLE records TYPE RELATION IN event OUT capability;
DEFINE TABLE triggers TYPE RELATION IN event OUT state;
DEFINE TABLE substantiates TYPE RELATION IN evidence OUT objective;
DEFINE TABLE derivedfrom TYPE RELATION IN evidence OUT event;
DEFINE TABLE supports TYPE RELATION IN evidence OUT policy;
DEFINE TABLE partof TYPE RELATION IN feature OUT product;
DEFINE TABLE realizes TYPE RELATION IN feature OUT capability;
DEFINE TABLE hasstate TYPE RELATION IN feature OUT state;
DEFINE TABLE usesdatatype TYPE RELATION IN field_group OUT data_type;
DEFINE TABLE composedinto TYPE RELATION IN field_group OUT schema;
DEFINE TABLE contains TYPE RELATION IN group OUT identity;
DEFINE TABLE hasrole TYPE RELATION IN group OUT role;
DEFINE TABLE withintenant TYPE RELATION IN group OUT tenant;
DEFINE TABLE holds TYPE RELATION IN identity OUT capability;
DEFINE TABLE subjectto TYPE RELATION IN identity OUT policy;
DEFINE TABLE locatedat TYPE RELATION IN identity OUT location;
DEFINE TABLE sequenceof TYPE RELATION IN journey OUT state;
DEFINE TABLE foridentity TYPE RELATION IN journey OUT identity;
DEFINE TABLE over TYPE RELATION IN journey OUT time;
DEFINE TABLE bounds TYPE RELATION IN location OUT constraint;
DEFINE TABLE categorizes TYPE RELATION IN market OUT product;
DEFINE TABLE measures TYPE RELATION IN metric OUT objective;
DEFINE TABLE evidencedby TYPE RELATION IN metric OUT evidence;
DEFINE TABLE servedby TYPE RELATION IN objective OUT capability;
DEFINE TABLE boundedby TYPE RELATION IN objective OUT constraint;
DEFINE TABLE measuredby TYPE RELATION IN objective OUT event;
DEFINE TABLE subobjectiveof TYPE RELATION IN objective OUT objective;
DEFINE TABLE permits TYPE RELATION IN permission OUT capability;
DEFINE TABLE over TYPE RELATION IN permission OUT resource;
DEFINE TABLE composesinto TYPE RELATION IN permission OUT policy;
DEFINE TABLE from TYPE RELATION IN pipeline OUT source;
DEFINE TABLE produces TYPE RELATION IN pipeline OUT dataset;
DEFINE TABLE emits TYPE RELATION IN pipeline OUT event;
DEFINE TABLE governedby TYPE RELATION IN pipeline OUT policy;
DEFINE TABLE composedof TYPE RELATION IN policy OUT constraint;
DEFINE TABLE governs TYPE RELATION IN policy OUT capability;
DEFINE TABLE appliesto TYPE RELATION IN policy OUT identity;
DEFINE TABLE packages TYPE RELATION IN product OUT capability;
DEFINE TABLE addresses TYPE RELATION IN product OUT market;
DEFINE TABLE hasstate TYPE RELATION IN product OUT state;
DEFINE TABLE carries TYPE RELATION IN protocol OUT schema;
DEFINE TABLE specifiedby TYPE RELATION IN protocol OUT evidence;
DEFINE TABLE connects TYPE RELATION IN relationship OUT thing;
DEFINE TABLE consumedby TYPE RELATION IN resource OUT capability;
DEFINE TABLE boundedby TYPE RELATION IN resource OUT constraint;
DEFINE TABLE ownedby TYPE RELATION IN resource OUT identity;
DEFINE TABLE threatens TYPE RELATION IN risk OUT thing;
DEFINE TABLE mitigatedby TYPE RELATION IN risk OUT constraint;
DEFINE TABLE governedby TYPE RELATION IN risk OUT policy;
DEFINE TABLE evidencedby TYPE RELATION IN risk OUT evidence;
DEFINE TABLE grants TYPE RELATION IN role OUT permission;
DEFINE TABLE assignedto TYPE RELATION IN role OUT identity;
DEFINE TABLE binds TYPE RELATION IN runtime OUT identity;
DEFINE TABLE hosts TYPE RELATION IN runtime OUT agent;
DEFINE TABLE remembers TYPE RELATION IN runtime OUT event;
DEFINE TABLE holds TYPE RELATION IN runtime OUT state;
DEFINE TABLE exposes TYPE RELATION IN runtime OUT touchpoint;
DEFINE TABLE guarantees TYPE RELATION IN runtime OUT objective;
DEFINE TABLE basedonclass TYPE RELATION IN schema OUT thing;
DEFINE TABLE includesfieldgroup TYPE RELATION IN schema OUT field_group;
DEFINE TABLE governedby TYPE RELATION IN schema OUT policy;
DEFINE TABLE foraccount TYPE RELATION IN session OUT account;
DEFINE TABLE ondevice TYPE RELATION IN session OUT device;
DEFINE TABLE during TYPE RELATION IN session OUT time;
DEFINE TABLE composes TYPE RELATION IN solution OUT product;
DEFINE TABLE addresses TYPE RELATION IN solution OUT market;
DEFINE TABLE servesobjective TYPE RELATION IN solution OUT objective;
DEFINE TABLE provides TYPE RELATION IN source OUT account;
DEFINE TABLE governedby TYPE RELATION IN source OUT policy;
DEFINE TABLE locatedat TYPE RELATION IN source OUT location;
DEFINE TABLE advancedby TYPE RELATION IN state OUT event;
DEFINE TABLE gatedby TYPE RELATION IN state OUT constraint;
DEFINE TABLE isolates TYPE RELATION IN tenant OUT identity;
DEFINE TABLE owns TYPE RELATION IN tenant OUT resource;
DEFINE TABLE governedby TYPE RELATION IN tenant OUT policy;
DEFINE TABLE relatedto TYPE RELATION IN thing OUT thing;
DEFINE TABLE scopes TYPE RELATION IN time OUT constraint;
DEFINE TABLE speaks TYPE RELATION IN touchpoint OUT protocol;
DEFINE TABLE exposedby TYPE RELATION IN touchpoint OUT capability;
DEFINE TABLE connects TYPE RELATION IN touchpoint OUT thing;
DEFINE TABLE securedby TYPE RELATION IN touchpoint OUT credential;
DEFINE TABLE governedby TYPE RELATION IN touchpoint OUT policy;
DEFINE TABLE emits TYPE RELATION IN touchpoint OUT event;
DEFINE TABLE locatedat TYPE RELATION IN touchpoint OUT location;
