#!/usr/bin/env python3
"""Generate the OpenDataWorld product catalog: an index page plus one landing
page per product, from catalog.json. Static output, shares catalog.css."""
import html
import json
import os

HERE = os.path.dirname(os.path.abspath(__file__))
DATA = os.path.join(HERE, "catalog.json")

STATUS_CLASS = {
    "Available": "ok", "Beta": "beta", "In design": "design", "Planned": "planned",
}


def esc(s: str) -> str:
    return html.escape(s, quote=True)


def page(title, body, depth_to_site_styles="../styles.css"):
    return f"""<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>{esc(title)}</title>
  <link rel="stylesheet" href="{depth_to_site_styles}" />
  <link rel="stylesheet" href="catalog.css" />
</head>
<body>
  <header class="nav">
    <a class="brand" href="../index.html"><span class="brand-mark"></span>
      <span>OpenDataWorld<small>Product Catalog</small></span></a>
    <nav class="nav-links"><a href="index.html">All products</a>
      <a href="../index.html#contact" class="btn btn-ghost">Contact</a></nav>
  </header>
  {body}
  <footer class="footer"><p>© <span id="y"></span> OpenDataWorld · Product Catalog</p></footer>
  <script>document.getElementById("y").textContent=new Date().getFullYear();</script>
</body>
</html>
"""


def status_badge(status):
    cls = STATUS_CLASS.get(status, "planned")
    return f'<span class="badge badge-{cls}">{esc(status)}</span>'


def build_index(data):
    cats = data["categories"]
    total = sum(len(c["products"]) for c in cats)
    sections = []
    for cat in cats:
        cards = []
        for p in cat["products"]:
            cards.append(f"""
        <a class="cat-card" href="{esc(p['slug'])}.html">
          <div class="cat-card-top">{status_badge(p['status'])}</div>
          <h3>{esc(p['name'])}</h3>
          <p>{esc(p['tagline'])}</p>
        </a>""")
        sections.append(f"""
    <section class="cat-section">
      <div class="cat-head"><h2>{esc(cat['name'])}</h2><p>{esc(cat['blurb'])}</p></div>
      <div class="cat-grid">{''.join(cards)}</div>
    </section>""")
    body = f"""
  <section class="hero hero-sm">
    <div class="hero-inner">
      <p class="eyebrow">Product Catalog</p>
      <h1>{esc(data['brand'])}</h1>
      <p class="lede">{esc(data['tagline'])}</p>
      <p class="hero-principle">{total} products across {len(cats)} categories</p>
    </div>
  </section>
  <main class="section">{''.join(sections)}</main>"""
    return page(f"{data['brand']} — Product Catalog", body)


def build_product(p, cat_name):
    feats = "".join(f"<li>{esc(f)}</li>" for f in p["features"])
    body = f"""
  <section class="hero hero-sm">
    <div class="hero-inner">
      <p class="eyebrow">{esc(cat_name)}</p>
      <h1>{esc(p['name'])} {status_badge(p['status'])}</h1>
      <p class="lede">{esc(p['tagline'])}</p>
      <div class="hero-cta">
        <a href="../index.html#contact" class="btn btn-primary">Talk to us</a>
        <a href="index.html" class="btn btn-ghost">← All products</a>
      </div>
    </div>
  </section>
  <main class="section product-body">
    <p class="product-desc">{esc(p['description'])}</p>
    <div class="product-cols">
      <div class="card"><h3>Highlights</h3><ul class="ticks">{feats}</ul></div>
      <div class="card"><h3>How it connects</h3><p>{esc(p['connects'])}</p></div>
    </div>
  </main>"""
    return page(f"{p['name']} — {p['brand_suffix']}", body)


def main():
    data = json.load(open(DATA))
    # index
    open(os.path.join(HERE, "index.html"), "w").write(build_index(data))
    count = 0
    for cat in data["categories"]:
        for p in cat["products"]:
            p["brand_suffix"] = data["brand"]
            open(os.path.join(HERE, f"{p['slug']}.html"), "w").write(build_product(p, cat["name"]))
            count += 1
    # stylesheet additions specific to catalog
    open(os.path.join(HERE, "catalog.css"), "w").write(CATALOG_CSS)
    print(f"Generated index.html + {count} product pages + catalog.css")


CATALOG_CSS = """/* Catalog-specific styles, layered on the site's styles.css */
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
.badge { display: inline-block; font-size: .72rem; font-weight: 700; letter-spacing: .04em; text-transform: uppercase; padding: 3px 9px; border-radius: 999px; }
.badge-ok { background: rgba(40,216,176,.16); color: var(--brand-2); }
.badge-beta { background: rgba(91,140,255,.18); color: var(--brand); }
.badge-design { background: rgba(255,196,84,.16); color: #ffc454; }
.badge-planned { background: rgba(154,164,184,.16); color: var(--muted); }
.product-body { max-width: 880px; }
.product-desc { font-size: 1.15rem; color: var(--text); margin-bottom: 28px; }
.product-cols { display: grid; grid-template-columns: 1fr 1fr; gap: 18px; }
.ticks { margin: 0; padding-left: 1.1em; }
.ticks li { margin-bottom: 6px; color: var(--muted); }
@media (max-width: 720px) { .product-cols { grid-template-columns: 1fr; } }
"""


if __name__ == "__main__":
    main()
