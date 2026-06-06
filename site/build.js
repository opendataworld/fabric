// Product MVP Builder — composes a Fabric Product spec from the live model.
// Static, no backend: reads data/model.json for capabilities, persists the
// work-in-progress to localStorage, and generates Product/Feature YAML +
// a catalog.json snippet for export.

const STORE_KEY = "fabric.mvp.builder.v1";
const MODEL_URL = "data/model.json";

// Lifecycle stages mirror meta/feature-lifecycle.yaml (State primitive).
const STAGES = ["Planned", "In design", "Beta", "Available", "Deprecated", "Retired"];

// A worked example so the builder is never an empty page.
const EXAMPLE = {
  name: "Customer Feedback Collector",
  slug: "customer-feedback-collector",
  tagline: "Capture, route, and govern customer feedback as first-class assets.",
  description:
    "An MVP that turns scattered customer feedback into governed, connected assets — collected across channels, tagged by sentiment, and routed to the right owner with a full audit trail.",
  market: "Customer Experience",
  status: "In design",
  capabilities: ["event", "identity", "consent", "policy", "metric"],
  customCaps: ["Sentiment tagging", "Channel ingestion"],
  features: [
    { name: "Feedback intake form", stage: "Beta" },
    { name: "Channel connectors (email, chat, web)", stage: "In design" },
    { name: "Sentiment & topic tagging", stage: "Planned" },
    { name: "Routing & ownership rules", stage: "Planned" },
    { name: "Consent-aware storage", stage: "In design" },
  ],
};

// One-click starter templates — each is a complete worked MVP.
const PRESETS = [
  EXAMPLE,
  {
    name: "Investor Pitch Deck",
    slug: "investor-pitch-deck",
    tagline: "Assemble, version, and govern a fundraising deck as a living asset.",
    description:
      "An MVP that treats the pitch deck as a governed asset: a structured outline, metric-bound slides pulled from source data, version history, and access controls for the data room.",
    market: "Fundraising",
    status: "Planned",
    capabilities: ["metric", "evidence", "policy", "event", "identity"],
    customCaps: ["Slide templating", "Data-room sharing"],
    features: [
      { name: "Deck outline & sections", stage: "In design" },
      { name: "Metric-bound slides", stage: "Planned" },
      { name: "Version history", stage: "Planned" },
      { name: "Data-room access controls", stage: "Planned" },
    ],
  },
  {
    name: "Cloud Control Panel",
    slug: "cloud-control-panel",
    tagline: "One governed console to provision, observe, and police cloud resources.",
    description:
      "An MVP control plane: a unified panel over cloud Resources with policy-enforced provisioning, live State, and an Event audit trail across accounts.",
    market: "Platform Engineering",
    status: "In design",
    capabilities: ["resource", "policy", "constraint", "event", "state", "metric"],
    customCaps: ["Provisioning", "Cost dashboards"],
    features: [
      { name: "Resource inventory", stage: "Beta" },
      { name: "Policy-guarded provisioning", stage: "In design" },
      { name: "Live health & state", stage: "In design" },
      { name: "Cost & usage dashboards", stage: "Planned" },
      { name: "Audit trail", stage: "Planned" },
    ],
  },
  {
    name: "Liferay on Kubernetes",
    slug: "liferay-on-kubernetes",
    tagline: "Run and govern Liferay DXP as a managed, Kubernetes-native deployment.",
    description:
      "An MVP that packages Liferay DXP for Kubernetes: Helm-based provisioning, autoscaling and persistent storage as governed Resources, SSO, and live health/observability with a full audit trail.",
    market: "Digital Experience Platforms",
    status: "In design",
    capabilities: ["resource", "application", "connector", "state", "event", "policy", "metric"],
    customCaps: ["Helm packaging", "Autoscaling"],
    features: [
      { name: "Helm chart & values", stage: "Beta" },
      { name: "Autoscaling & HPA", stage: "In design" },
      { name: "Persistent storage", stage: "In design" },
      { name: "SSO integration", stage: "Planned" },
      { name: "Observability dashboards", stage: "Planned" },
      { name: "Zero-downtime upgrades", stage: "Planned" },
    ],
  },
];

const $ = (s, r = document) => r.querySelector(s);
const $$ = (s, r = document) => [...r.querySelectorAll(s)];
const slugify = (s) =>
  s.toLowerCase().trim().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "");
const esc = (s) => String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
const yamlStr = (s) => (/[:#{}\[\],&*?|<>=!%@`"']/.test(s) || s === "") ? JSON.stringify(s) : s;

let state = load();
let model = { nodes: [] };
let step = 0;
let exportTab = "product";

function blank() {
  return { name: "", slug: "", tagline: "", description: "", market: "", status: "In design",
           capabilities: [], customCaps: [], features: [] };
}
function load() {
  try { const s = JSON.parse(localStorage.getItem(STORE_KEY)); if (s && s.features) return s; }
  catch (e) { /* ignore */ }
  return { ...EXAMPLE };
}
function save() {
  try { localStorage.setItem(STORE_KEY, JSON.stringify(state)); } catch (e) {}
  flagSaved();
}
let savedTimer;
function flagSaved() {
  const f = $("#saved-flag");
  f.innerHTML = "<b>✓</b> Saved locally";
  clearTimeout(savedTimer);
  savedTimer = setTimeout(() => (f.textContent = ""), 1600);
}

/* ---------- init ---------- */
document.getElementById("year").textContent = new Date().getFullYear();

// Status select
const statusSel = $("#p-status");
STAGES.forEach((s) => {
  const o = document.createElement("option");
  o.value = s; o.textContent = s; statusSel.appendChild(o);
});

fetch(MODEL_URL)
  .then((r) => r.json())
  .then((m) => { model = m; renderCapabilities(); })
  .catch(() => { model = { nodes: [] }; renderCapabilities(); });

renderPresets();
hydrateForm();
renderFeatures();
bind();
goTo(0);

/* ---------- presets ---------- */
function renderPresets() {
  const wrap = $("#preset-picker");
  if (!wrap) return;
  wrap.innerHTML = "";
  PRESETS.forEach((p) => {
    const b = document.createElement("button");
    b.className = "pick" + (state.slug === p.slug ? " on" : "");
    b.innerHTML = `<b>${esc(p.name)}</b><span>${esc(p.tagline)}</span><span class="tick">✓</span>`;
    b.onclick = () => {
      state = JSON.parse(JSON.stringify(p));
      save(); hydrateForm(); renderCapabilities(); renderFeatures(); renderPresets();
    };
    wrap.appendChild(b);
  });
}

/* ---------- step 1: form ---------- */
function hydrateForm() {
  $("#p-name").value = state.name;
  $("#p-slug").value = state.slug;
  $("#p-tagline").value = state.tagline;
  $("#p-desc").value = state.description;
  $("#p-market").value = state.market;
  $("#p-status").value = state.status;
  updateId();
}
function updateId() {
  const slug = state.slug || slugify(state.name) || "…";
  $("#p-id").textContent = "fabric:product:" + slug;
}

/* ---------- step 2: capabilities ---------- */
function renderCapabilities() {
  const picker = $("#cap-picker");
  const q = ($("#cap-search").value || "").toLowerCase();
  picker.innerHTML = "";
  const nodes = (model.nodes || []).filter((n) =>
    !q || n.name.toLowerCase().includes(q) || n.id.toLowerCase().includes(q));
  if (!nodes.length) {
    picker.innerHTML = '<p class="empty" style="grid-column:1/-1">No matching model nodes.</p>';
  }
  nodes.forEach((n) => {
    const on = state.capabilities.includes(n.id);
    const b = document.createElement("button");
    b.className = "pick" + (on ? " on" : "");
    b.innerHTML = `<b>${esc(n.name)}</b><span class="mono-id">${esc(n.id)}</span><span class="tick">✓</span>`;
    b.onclick = () => {
      const i = state.capabilities.indexOf(n.id);
      if (i >= 0) state.capabilities.splice(i, 1); else state.capabilities.push(n.id);
      b.classList.toggle("on");
      capCount(); save();
    };
    picker.appendChild(b);
  });
  // custom capability chips
  state.customCaps.forEach((c) => {
    const b = document.createElement("button");
    b.className = "pick on";
    b.innerHTML = `<b>${esc(c)}</b><span class="mono-id">custom</span><span class="tick">✕</span>`;
    b.title = "Remove custom capability";
    b.onclick = () => { state.customCaps = state.customCaps.filter((x) => x !== c); renderCapabilities(); capCount(); save(); };
    picker.appendChild(b);
  });
  capCount();
}
function capCount() {
  $("#cap-count").textContent = state.capabilities.length + state.customCaps.length;
}

/* ---------- step 3: features ---------- */
function lifeRow(stage) {
  const i = STAGES.indexOf(stage);
  return STAGES.map((s, j) => {
    const cls = j < i ? "done" : j === i ? "now" : "todo";
    return `<span class="${cls}">${esc(s)}</span>`;
  }).join("");
}
function renderFeatures() {
  const list = $("#feature-list");
  list.innerHTML = "";
  if (!state.features.length) {
    list.innerHTML = '<p class="empty">No features yet. Add the smallest set that delivers the promise.</p>';
    return;
  }
  state.features.forEach((f, idx) => {
    const row = document.createElement("div");
    row.className = "feature-row";
    row.innerHTML = `
      <div class="field">
        <label>Feature name</label>
        <input type="text" value="${esc(f.name)}" data-i="${idx}" data-k="name" placeholder="e.g. Feedback intake form" />
        <div class="liferow">${lifeRow(f.stage)}</div>
      </div>
      <div class="field">
        <label>Lifecycle stage</label>
        <select data-i="${idx}" data-k="stage">${STAGES.map((s) => `<option ${s === f.stage ? "selected" : ""}>${s}</option>`).join("")}</select>
      </div>
      <button class="icon-btn" data-del="${idx}" title="Remove feature">✕</button>`;
    list.appendChild(row);
  });
  $$("#feature-list input, #feature-list select").forEach((el) => {
    el.addEventListener("input", (e) => {
      const i = +e.target.dataset.i, k = e.target.dataset.k;
      state.features[i][k] = e.target.value;
      if (k === "stage") renderFeatures();
      save();
    });
  });
  $$("#feature-list [data-del]").forEach((b) =>
    b.addEventListener("click", () => { state.features.splice(+b.dataset.del, 1); renderFeatures(); save(); }));
}

/* ---------- step 4: review / export ---------- */
function caps() {
  return [
    ...state.capabilities.map((id) => "fabric:primitive:" + id),
    ...state.customCaps.map((c) => "fabric:capability:" + slugify(c)),
  ];
}
function slug() { return state.slug || slugify(state.name) || "untitled"; }

function productYaml() {
  const s = slug();
  const lines = [
    "# " + (state.name || "Untitled") + " — generated by the Fabric Product MVP Builder",
    "id: " + yamlStr("fabric:product:" + s),
    "name: " + yamlStr(state.name || "Untitled"),
    "version: \"0.1.0\"",
    "layer: \"domain\"",
    "question: \"What packaged offering is this?\"",
    "extends: [\"fabric:primitive:product\"]",
    "description: " + yamlStr(state.description || state.tagline || ""),
    "schemaOrg: { type: \"schema:Product\", url: \"https://schema.org/Product\" }",
    "attributes:",
    "  - { name: id, type: string, required: true }",
    "  - { name: name, type: string, required: true }",
    "  - { name: tagline, type: string, required: false }",
    "  - { name: status, type: string, required: false }",
    "tagline: " + yamlStr(state.tagline || ""),
    "status: " + yamlStr(state.status),
    "relationships:",
  ];
  caps().forEach((c) =>
    lines.push("  - { name: packages, target: " + yamlStr(c) + ", cardinality: \"0..*\" }"));
  state.features.forEach((f) =>
    lines.push("  - { name: hasFeature, target: " + yamlStr("fabric:feature:" + slugify(f.name)) + ", cardinality: \"0..*\" }"));
  if (state.market)
    lines.push("  - { name: addresses, target: " + yamlStr("fabric:market:" + slugify(state.market)) + ", cardinality: \"0..*\" }");
  lines.push("  - { name: hasState, target: \"fabric:primitive:state\", cardinality: \"0..1\" }");
  return lines.join("\n") + "\n";
}

function featuresYaml() {
  const ps = slug();
  if (!state.features.length) return "# No features defined yet.\n";
  return state.features.map((f) => {
    return [
      "# " + (f.name || "Feature"),
      "id: " + yamlStr("fabric:feature:" + slugify(f.name)),
      "name: " + yamlStr(f.name || "Feature"),
      "version: \"0.1.0\"",
      "layer: \"domain\"",
      "question: \"What product capability is this?\"",
      "extends: [\"fabric:primitive:feature\"]",
      "status: " + yamlStr(f.stage),
      "relationships:",
      "  - { name: partOf, target: " + yamlStr("fabric:product:" + ps) + ", cardinality: \"0..*\" }",
      "  - { name: hasState, target: \"fabric:primitive:state\", cardinality: \"0..1\" }",
    ].join("\n");
  }).join("\n---\n") + "\n";
}

function catalogJson() {
  const featNames = state.features.map((f) => f.name).filter(Boolean);
  const obj = {
    slug: slug(),
    name: state.name || "Untitled",
    status: state.status,
    tagline: state.tagline || "",
    description: state.description || "",
    features: featNames,
    connects: state.market ? ("Addresses the " + state.market + " market.") : "",
  };
  return JSON.stringify(obj, null, 2) + "\n";
}

function currentExport() {
  return exportTab === "product" ? productYaml()
       : exportTab === "features" ? featuresYaml()
       : catalogJson();
}
function exportFilename() {
  return exportTab === "product" ? slug() + "-model.yaml"
       : exportTab === "features" ? slug() + "-features.yaml"
       : slug() + ".catalog.json";
}

function renderReview() {
  const allCaps = caps();
  $("#summary").innerHTML = `
    <h3>${esc(state.name || "Untitled product")}</h3>
    <p class="tagline">${esc(state.tagline || "No tagline yet.")}</p>
    <dl>
      <dt>ID</dt><dd class="mono-id">fabric:product:${esc(slug())}</dd>
      <dt>Status</dt><dd>${esc(state.status)}</dd>
      <dt>Market</dt><dd>${esc(state.market || "—")}</dd>
      <dt>Capabilities</dt><dd><div class="chips">${allCaps.length ? allCaps.map((c) => `<span class="chip">${esc(c)}</span>`).join("") : "—"}</div></dd>
      <dt>Features</dt><dd><div class="chips">${state.features.length ? state.features.map((f) => `<span class="chip">${esc(f.name || "?")} · ${esc(f.stage)}</span>`).join("") : "—"}</div></dd>
    </dl>`;
  $("#code-out").textContent = currentExport();
}

/* ---------- wizard navigation ---------- */
function commitForm() {
  state.name = $("#p-name").value;
  state.slug = $("#p-slug").value ? slugify($("#p-slug").value) : "";
  state.tagline = $("#p-tagline").value;
  state.description = $("#p-desc").value;
  state.market = $("#p-market").value;
  state.status = $("#p-status").value;
}
function validateStep1() {
  let ok = true;
  [["#p-name", state.name || $("#p-name").value], ["#p-tagline", state.tagline || $("#p-tagline").value]]
    .forEach(([sel, val]) => {
      const bad = !String(val).trim();
      $(sel).classList.toggle("invalid-field", bad);
      if (bad) ok = false;
    });
  return ok;
}
function goTo(n) {
  step = Math.max(0, Math.min(3, n));
  $$(".panel").forEach((p) => p.classList.toggle("active", +p.dataset.panel === step));
  $$("#steps li").forEach((li) => {
    const s = +li.dataset.step;
    li.classList.toggle("active", s === step);
    li.classList.toggle("done", s < step);
  });
  $("#prev-btn").style.visibility = step === 0 ? "hidden" : "visible";
  $("#next-btn").textContent = step === 3 ? "Done" : "Continue";
  if (step === 3) renderReview();
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function bind() {
  // step 1 live binding
  $("#p-name").addEventListener("input", () => { state.name = $("#p-name").value; updateId(); save(); });
  $("#p-slug").addEventListener("input", () => { state.slug = slugify($("#p-slug").value); updateId(); save(); });
  ["#p-tagline", "#p-desc", "#p-market", "#p-status"].forEach((sel) =>
    $(sel).addEventListener("input", () => { commitForm(); save(); }));

  // step 2
  $("#cap-search").addEventListener("input", renderCapabilities);
  $("#cap-custom").addEventListener("keydown", (e) => {
    if (e.key === "Enter" && e.target.value.trim()) {
      e.preventDefault();
      const v = e.target.value.trim();
      if (!state.customCaps.includes(v)) state.customCaps.push(v);
      e.target.value = "";
      renderCapabilities(); save();
    }
  });

  // step 3
  $("#add-feature").addEventListener("click", () => {
    state.features.push({ name: "", stage: "Planned" }); renderFeatures(); save();
  });

  // step 4 export tabs + actions
  $$("#export-tabs button").forEach((b) =>
    b.addEventListener("click", () => {
      exportTab = b.dataset.tab;
      $$("#export-tabs button").forEach((x) => x.classList.toggle("active", x === b));
      $("#code-out").textContent = currentExport();
    }));
  $("#copy-btn").addEventListener("click", async () => {
    try { await navigator.clipboard.writeText(currentExport()); flash($("#copy-btn"), "Copied!"); }
    catch (e) { flash($("#copy-btn"), "Press ⌘/Ctrl+C"); }
  });
  $("#download-btn").addEventListener("click", () => {
    const blob = new Blob([currentExport()], { type: "text/plain" });
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = exportFilename();
    a.click(); URL.revokeObjectURL(a.href);
  });
  $("#reset-btn").addEventListener("click", () => {
    if (!confirm("Clear this MVP and start from a blank product?")) return;
    state = blank(); save(); hydrateForm(); renderCapabilities(); renderFeatures(); goTo(0);
  });

  // stepper clicks
  $$("#steps li").forEach((li) =>
    li.addEventListener("click", () => { if (step === 0 && !validateStep1()) return; commitForm(); goTo(+li.dataset.step); }));

  // prev / next
  $("#prev-btn").addEventListener("click", () => goTo(step - 1));
  $("#next-btn").addEventListener("click", () => {
    if (step === 0) { commitForm(); if (!validateStep1()) return; }
    if (step === 3) { alert("Your MVP spec is ready — copy or download it below."); return; }
    goTo(step + 1);
  });
}
function flash(btn, msg) {
  const old = btn.textContent; btn.textContent = msg;
  setTimeout(() => (btn.textContent = old), 1400);
}
