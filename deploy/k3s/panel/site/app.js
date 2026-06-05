// k3s Stack Catalog — dependency-free, driven entirely by catalog.json.
// Edit catalog.json (or the generated ConfigMap) to recatalogue; no rebuild.
(function () {
  "use strict";

  const els = {
    title: document.getElementById("winTitle"),
    subtitle: document.getElementById("subtitle"),
    sidebar: document.getElementById("sidebar"),
    catalog: document.getElementById("catalog"),
    counts: document.getElementById("counts"),
    search: document.getElementById("search"),
  };

  let data = { categories: [] };
  let activeCategory = "All";
  let query = "";

  function escapeHtml(s) {
    return String(s == null ? "" : s).replace(/[&<>"']/g, (c) => ({
      "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
    }[c]));
  }

  function matches(item) {
    if (!query) return true;
    const hay = (item.name + " " + (item.description || "") + " " + (item.image || "") + " " + (item.maturity || "")).toLowerCase();
    return hay.includes(query);
  }

  const MATURITY = { graduated: "graduated", incubating: "incubating", sandbox: "sandbox" };

  function cardHtml(item) {
    const status = item.status === "deployed" ? "deployed" : "available";
    const m = MATURITY[item.maturity];
    const cncf = m ? `<span class="badge-cncf maturity-${m}">${m}</span>` : "";
    const inner = `
      <div class="card-head">
        <div class="logo">${escapeHtml(item.logo || "📦")}</div>
        <div class="card-name">${escapeHtml(item.name)}</div>
      </div>
      <div class="card-desc">${escapeHtml(item.description || "")}</div>
      <div class="image">${escapeHtml(item.image || "")}</div>
      <div class="card-meta">
        <span class="status"><i class="dot ${status}"></i>${status}</span>
        ${cncf}
      </div>`;
    return item.docs
      ? `<a class="card" href="${escapeHtml(item.docs)}" target="_blank" rel="noopener">${inner}</a>`
      : `<div class="card">${inner}</div>`;
  }

  function render() {
    // Sidebar
    const cats = ["All", ...data.categories.map((c) => c.name)];
    els.sidebar.innerHTML = cats
      .map((c) => `<button data-cat="${escapeHtml(c)}" class="${c === activeCategory ? "active" : ""}">${escapeHtml(c)}</button>`)
      .join("");

    // Cards, grouped by category
    const groups = data.categories.filter((c) => activeCategory === "All" || c.name === activeCategory);
    let html = "";
    let shown = 0, deployed = 0, total = 0, cncf = 0;
    data.categories.forEach((c) => c.items.forEach((i) => {
      total++;
      if (i.status === "deployed") deployed++;
      if (MATURITY[i.maturity]) cncf++;
    }));

    groups.forEach((c) => {
      const items = c.items.filter(matches);
      if (!items.length) return;
      shown += items.length;
      html += `<h2 class="category-title">${escapeHtml(c.name)}</h2>`;
      html += `<div class="grid">${items.map(cardHtml).join("")}</div>`;
    });

    els.catalog.innerHTML = html || `<div class="empty">No components match “${escapeHtml(query)}”.</div>`;
    els.counts.textContent = `${total} components · ${cncf} CNCF · ${deployed} deployed · ${shown} shown`;
  }

  els.search.addEventListener("input", (e) => { query = e.target.value.trim().toLowerCase(); render(); });
  els.sidebar.addEventListener("click", (e) => {
    const btn = e.target.closest("button[data-cat]");
    if (!btn) return;
    activeCategory = btn.getAttribute("data-cat");
    render();
  });

  fetch("catalog.json")
    .then((r) => { if (!r.ok) throw new Error(r.status); return r.json(); })
    .then((json) => {
      data = json;
      if (json.title) { document.title = json.title; els.title.textContent = json.title; }
      if (json.subtitle) els.subtitle.textContent = json.subtitle;
      render();
    })
    .catch((err) => {
      els.catalog.innerHTML = `<div class="empty">Failed to load catalog.json (${escapeHtml(err.message)}).</div>`;
    });
})();
