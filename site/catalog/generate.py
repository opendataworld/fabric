#!/usr/bin/env python3
"""Generate the OpenDataWorld catalogs: a hub plus three catalogs (Products,
Features, Solutions), each with an index and one landing page per item.
All driven by the *.json data files; static output sharing the site styling."""
import glob
import html
import json
import os

HERE = os.path.dirname(os.path.abspath(__file__))

# (json file, kind/slug-prefix, nav label)
CATALOGS = [
    ("catalog.json", "products", "Products"),
    ("features.json", "features", "Features"),
    ("solutions.json", "solutions", "Solutions"),
]

STATUS_CLASS = {
    "Available": "ok", "Beta": "beta", "In design": "design", "Planned": "planned",
    "Core": "core", "Cloud": "cloud",
}


def esc(s):
    return html.escape(str(s), quote=True)


def nav():
    return """
  <header class="nav">
    <a class="brand" href="index.html"><span class="brand-mark"></span>
      <span>OpenDataWorld<small>Catalog</small></span></a>
    <nav class="nav-links">
      <a href="products.html">Products</a>
      <a href="features.html">Features</a>
      <a href="solutions.html">Solutions</a>
      <a href="../index.html" class="btn btn-ghost">Home</a>
    </nav>
  </header>"""


def page(title, body):
    return f"""<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>{esc(title)}</title>
  <link rel="stylesheet" href="../styles.css" />
  <link rel="stylesheet" href="catalog.css" />
</head>
<body>{nav()}
  {body}
  <footer class="footer"><p>© <span id="y"></span> OpenDataWorld · Catalog</p></footer>
  <script>document.getElementById("y").textContent=new Date().getFullYear();</script>
</body>
</html>
"""


# Gartner Peer Insights market each offering competes in (slug -> market name).
MARKETS = {
    # products
    "fabric-db": "Cloud Database Management Systems",
    "knowledge-graph": "Cloud Database Management Systems",
    "identity-fabric": "Identity Governance and Administration",
    "data-catalog": "Metadata Management Solutions",
    "data-integration": "Data Integration Tools",
    "analytics": "Analytics and Business Intelligence Platforms",
    "data-observability": "Data Observability Tools",
    "data-quality": "Augmented Data Quality Solutions",
    "governance-platform": "Data and Analytics Governance Platforms",
    "entity-registry": "Master Data Management Solutions",
    "feature-store": "Data Science and Machine Learning Platforms",
    "data-marketplace": "Data Marketplaces and Exchanges",
    "agent-runtime": "AI Agent Development Platforms",
    # solutions
    "customer-360": "Customer Data Platforms",
    "master-data-management": "Master Data Management Solutions",
    "data-governance": "Data and Analytics Governance Platforms",
    "data-lineage": "Metadata Management Solutions",
    "real-time-analytics": "Analytics and Business Intelligence Platforms",
    "access-audit": "Identity Governance and Administration",
    "data-exchange": "Data Marketplaces and Exchanges",
    "ai-agent-platform": "AI Agent Development Platforms",
    "knowledge-graph-search": "Cloud Database Management Systems",
    "graph-rag": "Data Science and Machine Learning Platforms",
    "entity-resolution": "Master Data Management Solutions",
}


def badge(status):
    return f'<span class="badge badge-{STATUS_CLASS.get(status, "planned")}">{esc(status)}</span>'


def market_tag(slug):
    m = MARKETS.get(slug)
    return f'<p class="market-tag">Gartner market · {esc(m)}</p>' if m else ""


# Governed feature lifecycle (see meta/feature-lifecycle.yaml).
LIFECYCLE = ["Planned", "In design", "Beta", "Available", "Deprecated"]


def lifecycle_stepper(status):
    """Render a pipeline stepper. Returns '' for non-lifecycle tags (Core/Cloud)."""
    if status not in LIFECYCLE:
        return ""
    cur = LIFECYCLE.index(status)
    steps = "".join(
        f'<li class="{"done" if i < cur else "now" if i == cur else "todo"}">{esc(s)}</li>'
        for i, s in enumerate(LIFECYCLE))
    return f'<ol class="lifecycle">{steps}</ol>'


def build_index(data, kind):
    cats = data["categories"]
    total = sum(len(c["products"]) for c in cats)
    sections = []
    for cat in cats:
        cards = "".join(
            f"""
        <a class="cat-card" href="{kind}-{esc(p['slug'])}.html">
          <div class="cat-card-top">{badge(p['status'])}</div>
          <h3>{esc(p['name'])}</h3><p>{esc(p['tagline'])}</p>
        </a>""" for p in cat["products"])
        sections.append(f"""
    <section class="cat-section">
      <div class="cat-head"><h2>{esc(cat['name'])}</h2><p>{esc(cat['blurb'])}</p></div>
      <div class="cat-grid">{cards}</div>
    </section>""")
    body = f"""
  <section class="hero hero-sm"><div class="hero-inner">
    <p class="eyebrow">{esc(data.get('title','Catalog'))}</p>
    <h1>{esc(data['brand'])}</h1>
    <p class="lede">{esc(data['tagline'])}</p>
    <p class="hero-principle">{total} {kind} across {len(cats)} categories</p>
  </div></section>
  <main class="section">{''.join(sections)}</main>"""
    return page(f"{data['brand']} — {data.get('title','Catalog')}", body)


def build_item(p, cat_name, siblings, kind, data):
    second_heading = data.get("second_heading", "How it connects")
    feats = "".join(f"<li>{esc(f)}</li>" for f in p["features"])
    related = "".join(
        f'<a class="related-chip" href="{kind}-{esc(s["slug"])}.html">{esc(s["name"])}'
        f'<span>{esc(s["tagline"])}</span></a>' for s in siblings)
    related_block = f"""
    <section class="related"><h3>Related in {esc(cat_name)}</h3>
      <div class="related-grid">{related}</div></section>""" if related else ""
    body = f"""
  <section class="hero hero-sm"><div class="hero-inner">
    <p class="eyebrow">{esc(cat_name)}</p>
    <h1>{esc(p['name'])} {badge(p['status'])}</h1>
    {market_tag(p['slug'])}
    {lifecycle_stepper(p['status'])}
    <p class="lede">{esc(p['tagline'])}</p>
    <div class="hero-cta">
      <a href="../index.html#contact" class="btn btn-primary">Talk to us</a>
      <a href="{kind}.html" class="btn btn-ghost">← All {kind}</a>
    </div>
  </div></section>
  <main class="section product-body">
    <p class="product-desc">{esc(p['description'])}</p>
    <div class="product-cols">
      <div class="card"><h3>Highlights</h3><ul class="ticks">{feats}</ul></div>
      <div class="card"><h3>{esc(second_heading)}</h3><p>{esc(p['connects'])}</p></div>
    </div>{related_block}
  </main>"""
    return page(f"{p['name']} — {data['brand']}", body)


def build_hub(meta):
    cards = "".join(
        f"""
      <a class="hub-card" href="{kind}.html">
        <h2>{esc(label)}</h2>
        <p class="hub-count">{count} {kind}</p>
        <p>{esc(tagline)}</p>
      </a>""" for kind, label, count, tagline in meta)
    body = f"""
  <section class="hero hero-sm"><div class="hero-inner">
    <p class="eyebrow">Catalog</p>
    <h1>OpenDataWorld Catalog</h1>
    <p class="lede">Everything is an asset. Everything is governed. Everything is connected.</p>
  </div></section>
  <main class="section"><div class="hub-grid">{cards}</div></main>"""
    return page("OpenDataWorld — Catalog", body)


def main():
    for f in glob.glob(os.path.join(HERE, "*.html")):
        os.remove(f)
    meta = []
    for fname, kind, label in CATALOGS:
        data = json.load(open(os.path.join(HERE, fname)))
        open(os.path.join(HERE, f"{kind}.html"), "w").write(build_index(data, kind))
        count = 0
        for cat in data["categories"]:
            for p in cat["products"]:
                siblings = [s for s in cat["products"] if s["slug"] != p["slug"]]
                open(os.path.join(HERE, f"{kind}-{p['slug']}.html"), "w").write(
                    build_item(p, cat["name"], siblings, kind, data))
                count += 1
        meta.append((kind, label, count, data["tagline"]))
        print(f"  {label}: index + {count} pages")
    open(os.path.join(HERE, "index.html"), "w").write(build_hub(meta))
    open(os.path.join(HERE, "catalog.css"), "w").write(CATALOG_CSS)
    print(f"Hub + {sum(m[2] for m in meta)} item pages generated.")


CATALOG_CSS = """/* Catalog styles, layered on the site's styles.css */
.hero-sm { padding: 56px 24px 40px; }
.cat-section { margin: 0 auto 48px; max-width: 1080px; }
.cat-head { margin-bottom: 18px; }
.cat-head h2 { font-size: 1.5rem; }
.cat-head p { color: var(--muted); margin: 0; }
.cat-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 16px; }
.cat-card { background: var(--surface); border: 1px solid var(--border); border-radius: 14px; padding: 20px; transition: transform .12s ease, border-color .2s ease; }
.cat-card:hover { transform: translateY(-2px); border-color: var(--brand); }
.cat-card-top { margin-bottom: 8px; }
.cat-card h3 { margin: 0 0 6px; }
.cat-card p { color: var(--muted); margin: 0; font-size: .92rem; }
.hub-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(260px, 1fr)); gap: 18px; max-width: 1080px; margin: 0 auto; }
.hub-card { background: var(--surface); border: 1px solid var(--border); border-radius: 16px; padding: 30px; transition: transform .12s ease, border-color .2s ease; }
.hub-card:hover { transform: translateY(-3px); border-color: var(--brand); }
.hub-card h2 { margin: 0 0 6px; }
.hub-count { color: var(--brand-2); font-weight: 700; margin: 0 0 10px; }
.hub-card p:last-child { color: var(--muted); margin: 0; }
.badge { display: inline-block; font-size: .72rem; font-weight: 700; letter-spacing: .04em; text-transform: uppercase; padding: 3px 9px; border-radius: 999px; }
.badge-ok { background: rgba(40,216,176,.16); color: var(--brand-2); }
.badge-beta { background: rgba(91,140,255,.18); color: var(--brand); }
.badge-design { background: rgba(255,196,84,.16); color: #ffc454; }
.badge-planned { background: rgba(154,164,184,.16); color: var(--muted); }
.badge-core { background: rgba(40,216,176,.16); color: var(--brand-2); }
.badge-cloud { background: rgba(91,140,255,.18); color: var(--brand); }
.product-body { max-width: 880px; }
.product-desc { font-size: 1.15rem; color: var(--text); margin-bottom: 28px; }
.product-cols { display: grid; grid-template-columns: 1fr 1fr; gap: 18px; }
.ticks { margin: 0; padding-left: 1.1em; }
.ticks li { margin-bottom: 6px; color: var(--muted); }
.related { margin-top: 36px; }
.related h3 { font-size: 1.05rem; color: var(--muted); margin-bottom: 14px; }
.related-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 12px; }
.related-chip { background: var(--surface); border: 1px solid var(--border); border-radius: 10px; padding: 12px 14px; display: flex; flex-direction: column; transition: border-color .2s ease; }
.related-chip:hover { border-color: var(--brand); }
.related-chip span { color: var(--muted); font-size: .82rem; margin-top: 3px; }
.market-tag { display: inline-block; margin: 4px 0 0; font-size: .82rem; color: var(--brand); border: 1px solid var(--border); border-radius: 999px; padding: 4px 12px; }
.lifecycle { list-style: none; display: flex; flex-wrap: wrap; gap: 8px; padding: 0; margin: 14px 0 0; }
.lifecycle li { font-size: .74rem; font-weight: 600; letter-spacing: .03em; padding: 4px 10px; border-radius: 999px; border: 1px solid var(--border); color: var(--muted); }
.lifecycle li.done { color: var(--brand-2); border-color: rgba(40,216,176,.4); }
.lifecycle li.now { color: #04121c; background: linear-gradient(135deg, var(--brand), var(--brand-2)); border-color: transparent; }
.lifecycle li.todo { opacity: .55; }
@media (max-width: 720px) { .product-cols { grid-template-columns: 1fr; } }
"""


if __name__ == "__main__":
    main()
