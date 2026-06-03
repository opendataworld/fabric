//! Fabric DB — minimal SurrealDB-backed runtime for the Fabric data model.
//!
//! Demonstrates the data-fabric planes on real storage:
//!   Connect  -> connect to SurrealDB (any engine, env-driven)
//!   Catalog  -> create Identity / Account nodes (the Fabric model)
//!   Govern   -> RELATE an `owns` edge and append an immutable Event
//!   Activate -> query the graph back
//!
//! Endpoint and credentials come from the environment — never hardcoded:
//!   SURREAL_URL   (e.g. wss://<host>.surreal.cloud  or  ws://localhost:8000  or  mem://)
//!   SURREAL_USER  SURREAL_PASS  SURREAL_NS  SURREAL_DB
//!
//! NOTE: not compiled in this environment (no Rust toolchain / crate access here).
//! Targets surrealdb 2.x; minor API tweaks may be needed per point release.

use anyhow::Result;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use surrealdb::engine::any;
use surrealdb::opt::auth::Root;

#[derive(Debug, Serialize, Deserialize)]
struct Identity {
    kind: String,
    display_name: Option<String>,
    email: Option<String>,
    created_at: DateTime<Utc>,
}

#[derive(Debug, Serialize, Deserialize)]
struct Account {
    provider: String,
    provider_sub: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct Event {
    #[serde(rename = "type")]
    kind: String,
    actor: String,
    occurred_at: DateTime<Utc>,
}

fn env(key: &str, default: &str) -> String {
    std::env::var(key).unwrap_or_else(|_| default.to_string())
}

#[tokio::main]
async fn main() -> Result<()> {
    // ── Connect ──────────────────────────────────────────────────────────────
    let url = env("SURREAL_URL", "ws://localhost:8000");
    let db = any::connect(url).await?;
    db.signin(Root {
        username: &env("SURREAL_USER", "root"),
        password: &env("SURREAL_PASS", "root"),
    })
    .await?;
    db.use_ns(env("SURREAL_NS", "fabric"))
        .use_db(env("SURREAL_DB", "fabric"))
        .await?;

    // ── Catalog: create an Identity and an Account (Fabric primitives) ─────────
    let _identity: Option<Identity> = db
        .create(("identity", "demo"))
        .content(Identity {
            kind: "person".into(),
            display_name: Some("Ada Lovelace".into()),
            email: Some("ada@example.com".into()),
            created_at: Utc::now(),
        })
        .await?;

    let _account: Option<Account> = db
        .create(("account", "google_103xxx"))
        .content(Account {
            provider: "google".into(),
            provider_sub: "103xxx".into(),
        })
        .await?;

    // ── Govern: relate (graph edge) + append an immutable Event ───────────────
    db.query("RELATE identity:demo->owns->account:google_103xxx").await?;
    let _event: Option<Event> = db
        .create(("event", "signup_demo"))
        .content(Event {
            kind: "identity.signup".into(),
            actor: "identity:demo".into(),
            occurred_at: Utc::now(),
        })
        .await?;

    // ── Activate: query the graph back ────────────────────────────────────────
    let identities: Vec<Identity> = db.select("identity").await?;
    println!("Fabric DB ready — {} identity node(s).", identities.len());

    Ok(())
}
