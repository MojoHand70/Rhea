/* Rhea shell. Renders view notions generically: the client knows list/detail/
   analysis plus the system functions (worklist, rules). It knows nothing about
   any particular object type. */
"use strict";

const $ = (sel, el = document) => el.querySelector(sel);

let NAV = [];
let activeDomain = null;
const tabs = [];      // {id, title, render}
let activeTab = null;

async function api(path, opts) {
  const res = await fetch(path, opts);
  const body = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(body.error || res.statusText);
  return body;
}

function toast(msg, isError = false) {
  const t = $("#toast");
  t.textContent = msg;
  t.className = isError ? "error" : "";
  t.hidden = false;
  clearTimeout(t._timer);
  t._timer = setTimeout(() => { t.hidden = true; }, 5000);
}

function el(tag, attrs = {}, ...children) {
  const n = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (k === "class") n.className = v;
    else if (k.startsWith("on")) n.addEventListener(k.slice(2), v);
    else n.setAttribute(k, v);
  }
  for (const c of children) n.append(c);
  return n;
}

/* --- navigation --------------------------------------------------------- */

async function boot() {
  NAV = await api("/api/nav");
  renderActivityBar();
  if (NAV.length) selectDomain(NAV[0].domain);
}

function renderActivityBar() {
  const bar = $("#activity-bar");
  bar.replaceChildren(...NAV.map(d =>
    el("div", {
      class: "activity" + (d.domain === activeDomain ? " active" : ""),
      title: d.domain, onclick: () => selectDomain(d.domain),
    }, d.domain.slice(0, 2))));
}

function selectDomain(domain) {
  activeDomain = domain;
  renderActivityBar();
  const d = NAV.find(x => x.domain === domain);
  const menu = $("#submenu");
  menu.replaceChildren(el("h2", {}, domain));
  for (const fn of d.functions) {
    menu.append(el("h2", {}, fn.function));
    if (fn.function === "worklist") {
      menu.append(el("div", { class: "menu-item", onclick: () => openWorklist() }, "Inbox"));
    } else if (fn.function === "rules") {
      menu.append(el("div", { class: "menu-item", onclick: () => openRules() }, "All rules"));
    } else {
      for (const v of (fn.views || [])) {
        if (v.notion === "detail") continue; // details open from lists
        menu.append(el("div", {
          class: "menu-item", onclick: () => openView(v.view_id, v.title),
        }, v.title, el("span", { class: "notion" }, v.notion)));
      }
    }
  }
}

/* --- tabs ---------------------------------------------------------------- */

function openTab(id, title, render) {
  let t = tabs.find(x => x.id === id);
  if (!t) { t = { id, title, render }; tabs.push(t); }
  t.render = render;
  activeTab = t;
  renderTabBar();
  refreshActive();
}

function closeTab(t) {
  const i = tabs.indexOf(t);
  tabs.splice(i, 1);
  if (activeTab === t) activeTab = tabs[Math.max(0, i - 1)] || null;
  renderTabBar();
  refreshActive();
}

function renderTabBar() {
  $("#tab-bar").replaceChildren(...tabs.map(t =>
    el("div", { class: "tab" + (t === activeTab ? " active" : ""), role: "tab",
                onclick: () => { activeTab = t; renderTabBar(); refreshActive(); } },
      t.title,
      el("span", { class: "close", onclick: (e) => { e.stopPropagation(); closeTab(t); } }, "×"))));
}

async function refreshActive() {
  const c = $("#tab-content");
  if (!activeTab) { c.replaceChildren(el("p", { class: "hint" }, "Open something from the left.")); return; }
  c.replaceChildren(el("p", { class: "hint" }, "Loading…"));
  try { c.replaceChildren(...await activeTab.render()); }
  catch (e) { c.replaceChildren(el("p", { class: "hint" }, "Error: " + e.message)); }
}

/* --- notion renderers ----------------------------------------------------- */

function dataTable(columns, rows, onRowClick) {
  const numeric = columns.map((_, i) => rows.every(r => /^-?\d+(\.\d+)?$/.test(r[i] ?? "")));
  return el("table", {},
    el("thead", {}, el("tr", {}, ...columns.map((c, i) =>
      el("th", { class: numeric[i] ? "num" : "" }, c)))),
    el("tbody", {}, ...rows.map((r, ri) =>
      el("tr", onRowClick ? { class: "clickable", onclick: () => onRowClick(ri) } : {},
        ...r.map((v, i) => el("td", { class: numeric[i] ? "num" : "" }, v ?? ""))))));
}

function openView(viewId, title, objectId) {
  const tabId = objectId ? `${viewId}:${objectId}` : viewId;
  openTab(tabId, objectId ? `${title} ${objectId}` : title, async () => {
    const q = objectId ? `?object_id=${encodeURIComponent(objectId)}` : "";
    const data = await api(`/api/views/${encodeURIComponent(viewId)}${q}`);
    switch (data.view.notion) {
      case "list": return renderList(data);
      case "detail": return renderDetail(data);
      case "analysis": return renderAnalysis(data);
      default: return [el("p", { class: "hint" }, `Notion ${data.view.notion} not renderable.`)];
    }
  });
}

function renderList(d) {
  const open = (ri) => {
    if (!d.detail_view_id) return;
    openView(d.detail_view_id, "Detail", d.object_ids[ri]);
  };
  return [
    el("h1", {}, d.view.title),
    d.rows.length
      ? dataTable(d.columns, d.rows, d.detail_view_id ? open : null)
      : el("p", { class: "hint" }, "Nothing here yet."),
  ];
}

function renderDetail(d) {
  const out = [el("h1", {}, `${d.view.title} — ${d.object.object_id}`)];
  for (const s of d.sections) {
    out.push(el("div", { class: "section" },
      el("h3", {}, s.Title),
      el("div", { class: "kv" }, ...s.Fields.flatMap(f =>
        [el("div", { class: "k" }, f.Label), el("div", {}, f.Value)]))));
  }
  out.push(el("p", { class: "provenance" }, "Explained by: " + d.provenance));
  return out;
}

function renderAnalysis(d) {
  return [
    el("h1", {}, d.view.title),
    d.rows.length ? dataTable(d.columns, d.rows) : el("p", { class: "hint" }, "No data yet."),
  ];
}

/* --- system functions: worklist, rules ------------------------------------ */

function openWorklist() {
  openTab("worklist", "Worklist", async () => {
    const d = await api("/api/worklist");
    if (!d.events || !d.events.length) {
      return [el("h1", {}, "Worklist"), el("p", { class: "hint" }, "Inbox zero — every event is explained.")];
    }
    const out = [
      el("h1", {}, "Worklist"),
      el("p", { class: "hint" }, "Raw events no rule explains yet. Pick one and ask for a rule."),
    ];
    for (const ev of d.events) {
      out.push(el("div", { class: "draft-form" },
        el("div", { class: "row" },
          el("strong", {}, `#${ev.event_id} ${ev.event_type}`),
          el("span", { class: "hint" }, ev.occurred_at)),
        el("details", {}, el("summary", {}, "payload"),
          el("pre", {}, JSON.stringify(ev.payload, null, 2))),
        draftForm(ev, d.object_types)));
    }
    return out;
  });
}

function draftForm(ev, objectTypes) {
  const intent = el("textarea", {
    placeholder: "Describe the rule in plain language, e.g. “PLN invoices become invoice documents; total is the sum of line amounts.”",
  });
  const typeSel = el("select", {}, ...objectTypes.map(t => el("option", { value: t }, t)));
  const btn = el("button", {
    class: "primary",
    onclick: async () => {
      if (!intent.value.trim()) { toast("Describe the rule first.", true); return; }
      btn.disabled = true; btn.textContent = "Asking the agent…";
      try {
        const rule = await api("/api/rules/draft", {
          method: "POST", headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            intent: intent.value, sample_event_id: ev.event_id, object_type: typeSel.value,
          }),
        });
        toast(`Draft ${rule.rule_id} v${rule.version} ready for approval.`);
        openRules();
      } catch (e) { toast(e.message, true); }
      btn.disabled = false; btn.textContent = "Draft rule";
    },
  }, "Draft rule");
  return el("div", {}, intent,
    el("div", { class: "row", style: "margin-top:8px" },
      el("span", { class: "hint", style: "margin:0" }, "materialize as"), typeSel, btn));
}

/* The dry run (SPEC M1): what approving this rule would change, from an
   in-memory replay of the whole log. Rendered generically, like everything. */
function renderSimDiff(d) {
  const added = d.added || [], changed = d.changed || [], removed = d.removed || [];
  const before = (d.unexplained_before || []).length, after = (d.unexplained_after || []).length;
  const box = el("div", { class: "draft-form" },
    el("strong", {}, `If approved: ${added.length} object(s) materialize, ` +
      `${changed.length} change, ${removed.length} disappear; ` +
      `unexplained events ${before} → ${after}.`));
  const objTable = (title, objs) => {
    if (!objs.length) return;
    box.append(el("h3", {}, title), dataTable(["object", "rule", "state"],
      objs.map(o => [o.object_id, `${o.rule_id} v${o.rule_version}`, JSON.stringify(o.state)])));
  };
  objTable("Would materialize", added);
  if (changed.length) {
    box.append(el("h3", {}, "Would change"), dataTable(["object", "before", "after"],
      changed.map(c => [c.after.object_id, JSON.stringify(c.before.state), JSON.stringify(c.after.state)])));
  }
  objTable("Would disappear", removed);
  if ((d.errors || []).length) box.append(el("h3", {}, "Simulation errors"), el("pre", {}, d.errors.join("\n")));
  return box;
}

function openRules() {
  openTab("rules", "Rules", async () => {
    const rules = await api("/api/rules");
    const out = [el("h1", {}, "Rules"),
      el("p", { class: "hint" }, "Every rule version is a new row; approval activates. Nothing books without one.")];
    if (!rules.length) { out.push(el("p", { class: "hint" }, "No rules yet.")); return out; }
    const tbody = el("tbody", {});
    for (const r of rules) {
      const actions = el("td", {});
      const simBox = el("div", {});
      if (r.status === "draft") {
        const s = el("button", {
          onclick: async () => {
            s.disabled = true; s.textContent = "Simulating…";
            try {
              const d = await api(`/api/rules/${encodeURIComponent(r.rule_id)}/simulate`, { method: "POST" });
              simBox.replaceChildren(renderSimDiff(d));
            } catch (e) { toast(e.message, true); }
            s.disabled = false; s.textContent = "Simulate";
          },
        }, "Simulate");
        actions.append(s, " ");
        const b = el("button", {
          onclick: async () => {
            b.disabled = true;
            try {
              const res = await api(`/api/rules/${encodeURIComponent(r.rule_id)}/approve`, {
                method: "POST", headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ approved_by: "human" }),
              });
              toast(`${r.rule_id} active (v${res.rule.version}); booked ${res.booked} pending event(s).`);
              refreshActive();
            } catch (e) { toast(e.message, true); b.disabled = false; }
          },
        }, "Approve");
        actions.append(b);
      }
      tbody.append(
        el("tr", {},
          el("td", {}, r.rule_id), el("td", { class: "num" }, String(r.version)),
          el("td", {}, el("span", { class: "status " + r.status }, r.status)),
          el("td", {}, r.created_by), el("td", {}, r.description), actions),
        el("tr", {}, el("td", { colspan: "6" },
          el("details", {}, el("summary", {}, "spec"),
            el("pre", {}, JSON.stringify(r.spec, null, 2))),
          simBox)));
    }
    out.push(el("table", {},
      el("thead", {}, el("tr", {},
        el("th", {}, "rule"), el("th", { class: "num" }, "ver"), el("th", {}, "status"),
        el("th", {}, "author"), el("th", {}, "description"), el("th", {}, ""))),
      tbody));
    return out;
  });
}

boot().catch(e => toast(e.message, true));
