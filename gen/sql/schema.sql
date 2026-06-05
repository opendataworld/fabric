-- Generated from Fabric primitives — DO NOT EDIT BY HAND.

CREATE TABLE IF NOT EXISTS account (
    id TEXT PRIMARY KEY NOT NULL,
    provider TEXT NOT NULL,
    provider_sub TEXT NOT NULL,
    email TEXT,
    username TEXT
);

CREATE TABLE IF NOT EXISTS agent (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS application (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    clientId TEXT
);

CREATE TABLE IF NOT EXISTS capability (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    verb TEXT NOT NULL,
    inputs JSONB,
    outputs JSONB,
    requiresTools JSONB,
    requiresResources JSONB,
    maturity TEXT
);

CREATE TABLE IF NOT EXISTS connector (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS consent (
    id TEXT PRIMARY KEY NOT NULL,
    purpose TEXT NOT NULL,
    granted BOOLEAN
);

CREATE TABLE IF NOT EXISTS constraint (
    id TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    expression TEXT NOT NULL,
    target TEXT,
    severity TEXT,
    onViolation TEXT
);

CREATE TABLE IF NOT EXISTS control (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS credential (
    id TEXT PRIMARY KEY NOT NULL,
    type TEXT NOT NULL,
    claim JSONB,
    proof TEXT
);

CREATE TABLE IF NOT EXISTS data_type (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    fields JSONB
);

CREATE TABLE IF NOT EXISTS dataset (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    format TEXT
);

CREATE TABLE IF NOT EXISTS device (
    id TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    fingerprint TEXT
);

CREATE TABLE IF NOT EXISTS event (
    id TEXT PRIMARY KEY NOT NULL,
    type TEXT NOT NULL,
    occurredAt TIMESTAMPTZ NOT NULL,
    actor TEXT,
    subject TEXT,
    location TEXT,
    payload JSONB
);

CREATE TABLE IF NOT EXISTS evidence (
    id TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    claim TEXT NOT NULL,
    source TEXT,
    uri TEXT,
    hash TEXT,
    collectedAt TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS feature (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    status TEXT
);

CREATE TABLE IF NOT EXISTS field_group (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    fields JSONB
);

CREATE TABLE IF NOT EXISTS group (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS identity (
    id TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    displayName TEXT,
    controller TEXT,
    credentials JSONB
);

CREATE TABLE IF NOT EXISTS journey (
    id TEXT PRIMARY KEY NOT NULL,
    subject TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS location (
    id TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    geometry JSONB,
    address TEXT,
    uri TEXT
);

CREATE TABLE IF NOT EXISTS market (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS metric (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    value DOUBLE PRECISION,
    unit TEXT
);

CREATE TABLE IF NOT EXISTS objective (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    statement TEXT NOT NULL,
    metrics JSONB,
    targetDate TEXT,
    priority TEXT,
    status TEXT
);

CREATE TABLE IF NOT EXISTS permission (
    id TEXT PRIMARY KEY NOT NULL,
    action TEXT NOT NULL,
    effect TEXT
);

CREATE TABLE IF NOT EXISTS pipeline (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    mode TEXT
);

CREATE TABLE IF NOT EXISTS policy (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    effect TEXT NOT NULL,
    combine TEXT,
    constraints JSONB NOT NULL,
    scope JSONB,
    precedence INTEGER,
    owner TEXT
);

CREATE TABLE IF NOT EXISTS product (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    status TEXT
);

CREATE TABLE IF NOT EXISTS protocol (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    version TEXT,
    format TEXT,
    spec TEXT,
    status TEXT
);

CREATE TABLE IF NOT EXISTS relationship (
    id TEXT PRIMARY KEY NOT NULL,
    predicate TEXT NOT NULL,
    source TEXT NOT NULL,
    target TEXT NOT NULL,
    directed BOOLEAN,
    validDuring TEXT,
    weight DOUBLE PRECISION
);

CREATE TABLE IF NOT EXISTS resource (
    id TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    unit TEXT NOT NULL,
    capacity DOUBLE PRECISION,
    consumed DOUBLE PRECISION,
    owner TEXT
);

CREATE TABLE IF NOT EXISTS risk (
    id TEXT PRIMARY KEY NOT NULL,
    category TEXT NOT NULL,
    statement TEXT NOT NULL,
    likelihood TEXT,
    impact TEXT,
    severity TEXT,
    status TEXT
);

CREATE TABLE IF NOT EXISTS role (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS runtime (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    version TEXT
);

CREATE TABLE IF NOT EXISTS schema (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    baseClass TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS session (
    id TEXT PRIMARY KEY NOT NULL,
    started TIMESTAMPTZ,
    ip TEXT
);

CREATE TABLE IF NOT EXISTS solution (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS source (
    id TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    endpoint TEXT
);

CREATE TABLE IF NOT EXISTS state (
    id TEXT PRIMARY KEY NOT NULL,
    subject TEXT NOT NULL,
    value TEXT NOT NULL,
    lifecycle TEXT,
    since TIMESTAMPTZ,
    allowedTransitions JSONB
);

CREATE TABLE IF NOT EXISTS tenant (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS thing (
    id TEXT PRIMARY KEY NOT NULL,
    type TEXT NOT NULL,
    name TEXT,
    description TEXT,
    createdAt TIMESTAMPTZ,
    metadata JSONB
);

CREATE TABLE IF NOT EXISTS time (
    id TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    instant TIMESTAMPTZ,
    start TIMESTAMPTZ,
    end TIMESTAMPTZ,
    duration TEXT,
    timezone TEXT
);

CREATE TABLE IF NOT EXISTS touchpoint (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    surface TEXT NOT NULL,
    protocol TEXT NOT NULL,
    format TEXT,
    direction TEXT,
    endpoint TEXT,
    protocolVersion TEXT
);
