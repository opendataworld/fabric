# Fabric DB

The SurrealDB-backed runtime for the Fabric data model — a minimal, honest
implementation of the four planes (Connect · Catalog · Govern · Activate).

> **Status:** reference scaffold (alpha). Not compiled in CI here (no Rust
> toolchain / crate access in this environment). Targets `surrealdb` 2.x.

## Run

1. Start a SurrealDB instance (local example):
   ```
   surreal start --user root --pass root memory
   ```
2. Configure the connection via env (never hardcode the cloud endpoint):
   ```
   export SURREAL_URL=ws://localhost:8000
   export SURREAL_USER=root
   export SURREAL_PASS=root
   export SURREAL_NS=fabric
   export SURREAL_DB=fabric
   ```
3. (Optional) seed the schema generated from the primitives:
   ```
   surreal import --conn $SURREAL_URL --user $SURREAL_USER --pass $SURREAL_PASS \
     --ns $SURREAL_NS --db $SURREAL_DB ../gen/graph/schema.surql
   ```
4. Run:
   ```
   cargo run
   ```

## What it does
- Connects to SurrealDB via `surrealdb::engine::any`.
- Creates `Identity` and `Account` nodes.
- Relates them with an `owns` graph edge and appends an immutable `Event`.
- Selects identities back.

The node/edge schema is generated from the canonical primitives by
`../codegen/generate_models.py --target graph` (`gen/graph/schema.surql`).
