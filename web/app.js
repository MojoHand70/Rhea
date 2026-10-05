/* Rhea shell. Renders view notions generically: the client knows list/detail/
   analysis plus the system functions (worklist, rules, language). It knows
   nothing about any particular object type. */
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
  startLive();
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
    if (fn.function === "worklist") {
      menu.append(el("h2", {}, fn.function),
        el("div", { class: "menu-item", onclick: () => openWorklist() }, "Inbox"));
    } else if (fn.function === "rules") {
      menu.append(el("h2", {}, fn.function),
        el("div", { class: "menu-item", onclick: () => openRules() }, "All rules"));
    } else if (fn.function === "language") {
      menu.append(el("h2", {}, fn.function),
        el("div", { class: "menu-item", onclick: () => openLanguage() }, "Object types"));
    } else {
      // View groups fold (native disclosure): a pack can land a dozen types
      // as derived lists under one function, and the submenu must stay
      // walkable. Derived entries arrive alphabetical from the store.
      const group = el("details", { class: "menu-group", open: "" },
        el("summary", {}, fn.function));
      for (const v of (fn.views || [])) {
        if (v.notion === "detail") continue; // details open from lists
        group.append(el("div", {
          class: "menu-item", onclick: () => openView(v.view_id, v.title),
        }, v.title, el("span", { class: "notion" }, v.notion)));
      }
      menu.append(group);
    }
  }
}

/* --- tabs ---------------------------------------------------------------- */

function openTab(id, title, render, live = false) {
  let t = tabs.find(x => x.id === id);
  if (!t) { t = { id, title, render }; tabs.push(t); }
  t.render = render;
  t.live = live; // live tabs re-render when a projection notice arrives
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

/* --- typed values ---------------------------------------------------------
   The API carries semantics, the renderer decides formatting (the typed view
   API, DIRECTION 2026-10-05): columns declare {field, label, type}; cells are
   {v, id?} with v in canonical encoding (money as "1234.56" minor-exact
   decimal strings). This renderer formats for the browser's locale — Polish
   eyes see 1 234,56, English eyes 1,234.56, from the same response. */

const decimalSep = (() => {
  try { return new Intl.NumberFormat().formatToParts(1.1).find(p => p.type === "decimal").value; }
  catch { return "."; }
})();

function fmtMoney(v) { // "-1234.56" canonical → locale-grouped; BigInt keeps it float-free
  if (!v) return "";
  const neg = v.startsWith("-");
  const [whole, frac = "00"] = (neg ? v.slice(1) : v).split(".");
  try { return (neg ? "-" : "") + new Intl.NumberFormat().format(BigInt(whole)) + decimalSep + frac; }
  catch { return v; }
}

const numericType = (t) => t === "money" || t === "int";

/* Formats one typed cell into DOM content. A ref shows its label and is a
   door: it opens the referenced object's detail view (every type has one,
   derived if not stored). An enum renders as a chip, so state reads as
   state; an event-typed cell opens the provenance walk. Untyped columns
   (analysis without declarations) render as-is. */
function cellContent(cell, type) {
  const v = cell?.v ?? "";
  if (type === "money") return fmtMoney(v);
  if (type === "enum") return v ? el("span", { class: "chip" }, v) : "";
  if (type === "event") return v ? explainLink({ event: v }, v) : "";
  if (cell?.id && cell?.detail) {
    return el("a", { href: "#", title: cell.id, onclick: (e) => {
      e.preventDefault(); e.stopPropagation();
      openView(cell.detail, "Detail", cell.id);
    } }, v);
  }
  if (cell?.id && v !== cell.id) return el("span", { title: cell.id }, v);
  return v;
}

function explainLink(q, text) {
  return el("a", { href: "#", onclick: (e) => {
    e.preventDefault(); e.stopPropagation(); openExplain(q);
  } }, text);
}

function dataTable(columns, rows, onRowClick) {
  // Alignment is type-driven; for undeclared columns, fall back to sniffing —
  // a renderer heuristic, which is the one place heuristics belong.
  const numeric = columns.map((c, i) =>
    numericType(c.type) || (!c.type && rows.length > 0 && rows.every(r => /^-?\d+(\.\d+)?$/.test(r[i]?.v ?? ""))));
  const numCls = (i) => numeric[i] ? "num" : "";
  return el("table", {},
    el("thead", {}, el("tr", {}, ...columns.map((c, i) =>
      el("th", { class: numCls(i) }, c.label)))),
    el("tbody", {}, ...rows.map((r, ri) =>
      el("tr", onRowClick ? { class: "clickable", onclick: () => onRowClick(ri) } : {},
        ...r.map((cell, i) => el("td", { class: numCls(i) },
          cellContent(cell, columns[i].type)))))));
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
  }, true); // view tabs are live: projections re-render them in place
}

/* --- live screens -----------------------------------------------------------
   The executor is the single writer, so it is the single announcer: one SSE
   stream per browser tab, one notice per committed projection write. A notice
   is a signal, never data — live tabs re-read through the API. Tabs holding
   human state (worklist drafts, simulation results) stay manual. */

function startLive() {
  const es = new EventSource("/api/live");
  let timer;
  es.addEventListener("projection", () => {
    clearTimeout(timer); // bursts collapse into one re-render
    timer = setTimeout(() => { if (activeTab?.live) refreshActive(); }, 200);
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
      el("h3", {}, s.title),
      el("div", { class: "kv" }, ...s.fields.flatMap(f =>
        [el("div", { class: "k" }, f.label),
         el("div", f.type === "money" ? { class: "num-inline" } : {}, cellContent(f, f.type))]))));
  }
  const p = d.provenance;
  out.push(el("p", { class: "provenance" },
    `Explained by: event ${p.event_id} via rule ${p.rule_id} v${p.rule_version} — `,
    explainLink({ object: d.object.object_id }, "walk the explanation")));
  return out;
}

/* --- system surface: the provenance walk -----------------------------------
   Invariant 5 as an interaction: any object or event opens the full story of
   its chain's root raw event — every consequence it caused, each hop naming
   the exact rule version that explains it, each object a door to its detail.
   Both sides of an intercompany position walk to the same root. */

function openExplain(q) {
  const key = q.object ? `explain:o:${q.object}` : `explain:e:${q.event}`;
  const title = q.object ? `Why ${q.object}` : `Why event #${q.event}`;
  openTab(key, title, async () => {
    const qs = q.object ? `object=${encodeURIComponent(q.object)}`
                        : `event=${encodeURIComponent(q.event)}`;
    const d = await api(`/api/explain?${qs}`);
    const onPath = new Set(d.path);
    const asked = d.path[d.path.length - 1];

    const render = (n) => {
      const cls = "explain-head" + (onPath.has(n.event_id) ? " on-path" : "") +
        (n.event_id === asked ? " asked" : "");
      const head = el("div", { class: cls },
        el("span", { class: "explain-event" }, `#${n.event_id} ${n.event_type}`),
        el("span", { class: "hint" }, ` ${n.occurred_at}${n.actor ? " · " + n.actor : ""}`));
      if (n.rule_id) {
        head.append(el("div", { class: "explain-line" },
          `via rule ${n.rule_id} v${n.rule_version}`,
          n.rule_description ? el("span", { class: "hint" }, ` — ${n.rule_description}`) : ""));
      }
      if (n.object) {
        const o = n.object;
        const text = o.label || o.object_id;
        head.append(el("div", { class: "explain-line" }, "materialized ",
          o.detail_view_id
            ? el("a", { href: "#", title: o.object_id, onclick: (e) => {
                e.preventDefault(); openView(o.detail_view_id, "Detail", o.object_id);
              } }, text)
            : text,
          el("span", { class: "hint" }, ` ${o.object_type}`)));
      }
      if (n.payload) {
        head.append(el("details", {}, el("summary", {}, "the fact"),
          el("pre", {}, JSON.stringify(n.payload, null, 2))));
      }
      const node = el("div", { class: "explain-node" }, head);
      if (n.children.length) {
        node.append(el("div", { class: "explain-children" }, ...n.children.map(render)));
      }
      return node;
    };

    return [
      el("h1", {}, title),
      el("p", { class: "hint" },
        `Everything below follows from event #${d.root_event_id}. ` +
        "The highlighted chain leads to what you asked about; every hop names the rule version that explains it."),
      render(d.tree),
    ];
  }, true); // chains grow when rules fire; the walk stays current
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
  // The diff endpoint is not typed yet (readable diffs are their own stage);
  // wrap its plain strings into the shapes dataTable speaks.
  const cols = (names) => names.map(n => ({ label: n }));
  const cells = (vals) => vals.map(v => ({ v }));
  const objTable = (title, objs) => {
    if (!objs.length) return;
    box.append(el("h3", {}, title), dataTable(cols(["object", "rule", "state"]),
      objs.map(o => cells([o.object_id, `${o.rule_id} v${o.rule_version}`, JSON.stringify(o.state)]))));
  };
  objTable("Would materialize", added);
  if (changed.length) {
    box.append(el("h3", {}, "Would change"), dataTable(cols(["object", "before", "after"]),
      changed.map(c => cells([c.after.object_id, JSON.stringify(c.before.state), JSON.stringify(c.after.state)]))));
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
              toast(`${r.rule_id} active (v${res.rule.version}); booked ${res.booked} pending event(s).` +
                (res.errors ? ` Still failing: ${res.errors}` : ""));
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

/* --- system function: language --------------------------------------------
   The language explaining itself: every object type the kernel recognizes,
   with what produces it, what shows it, and what it relates to. The catalog
   is how a system that learns by explanation explains its own vocabulary. */

function openLanguage() {
  openTab("language", "Language", async () => {
    const types = await api("/api/types");
    if (!types.length) return [el("h1", {}, "Language"), el("p", { class: "hint" }, "No object types yet.")];
    const cols = ["type", "domain", "kind", "fields", "produced by", "views", "instances"];
    const tbody = el("tbody", {});
    for (const t of types) {
      tbody.append(el("tr", { class: "clickable", onclick: () => openType(t.name) },
        el("td", {}, t.name),
        el("td", {}, t.domain),
        el("td", {}, t.is_document ? "document" : "master data"),
        el("td", { class: "num" }, String(t.fields.length)),
        el("td", { class: "num" }, String(t.produced_by.length)),
        el("td", { class: "num" }, String(t.views.length)),
        el("td", { class: "num" }, String(t.instances))));
    }
    return [
      el("h1", {}, "Language — object types"),
      el("p", { class: "hint" },
        "Every type the kernel recognizes. A type with no producing rule is vocabulary waiting for rules."),
      el("table", {},
        el("thead", {}, el("tr", {}, ...cols.map((c, i) =>
          el("th", i >= 3 ? { class: "num" } : {}, c)))),
        tbody),
    ];
  }, true); // instance counts move with the projections
}

function openType(name) {
  openTab(`type:${name}`, name, async () => {
    const types = await api("/api/types");
    const t = types.find(x => x.name === name);
    if (!t) return [el("p", { class: "hint" }, `Type ${name} not found.`)];

    const out = [el("h1", {}, `${t.name} v${t.version}`),
      el("p", { class: "hint" },
        `${t.is_document ? "document" : "master data"} · domain ${t.domain}` +
        (t.label_field ? ` · labeled by ${t.label_field}` : "") +
        ` · ${t.instances} instance${t.instances === 1 ? "" : "s"}`)];

    out.push(el("div", { class: "section" }, el("h3", {}, "Fields"),
      el("table", {},
        el("thead", {}, el("tr", {}, el("th", {}, "field"), el("th", {}, "type"),
          el("th", {}, "required"), el("th", {}, "values"))),
        el("tbody", {}, ...t.fields.map(f =>
          el("tr", {}, el("td", {}, f.name), el("td", {}, f.type),
            el("td", {}, f.required ? "yes" : ""), el("td", {}, (f.values || []).join(", "))))))));

    const produced = el("div", { class: "section" }, el("h3", {}, "Produced by"));
    if (!t.produced_by.length) {
      produced.append(el("p", { class: "hint" },
        "No rule produces this type yet — vocabulary waiting for rules."));
    } else {
      for (const r of t.produced_by) {
        produced.append(el("p", {},
          el("strong", {}, r.event_type || "(any event)"), ` → ${r.rule_id} v${r.version} `,
          el("span", { class: "hint" }, `(${r.status}) ${r.description || ""}`)));
      }
    }
    out.push(produced);

    out.push(el("div", { class: "section" }, el("h3", {}, "Views"),
      ...(t.views.length
        ? t.views.map(v => el("p", {},
            // A detail needs an object; it opens from its list, not from here.
            v.notion === "detail"
              ? v.title
              : el("a", { href: "#", onclick: (e) => { e.preventDefault(); openView(v.view_id, v.title); } }, v.title),
            el("span", { class: "hint" },
              ` ${v.notion}${v.derived ? ", derived from the type" : ""}${v.notion === "detail" ? " — opens from its list" : ""}`)))
        : [el("p", { class: "hint" }, "No views — which cannot happen: list and detail derive from the type.")])));

    const rel = el("div", { class: "section" }, el("h3", {}, "Relations"));
    for (const e of t.references) {
      rel.append(el("p", {}, `${e.field} → `,
        el("a", { href: "#", onclick: (ev) => { ev.preventDefault(); openType(e.to_type); } }, e.to_type)));
    }
    for (const e of t.referenced_by) {
      rel.append(el("p", {},
        el("a", { href: "#", onclick: (ev) => { ev.preventDefault(); openType(e.from_type); } }, e.from_type),
        `.${e.field} → ${t.name}`));
    }
    if (!t.references.length && !t.referenced_by.length) {
      rel.append(el("p", { class: "hint" }, "Stands alone — no ref fields in either direction."));
    }
    out.push(rel);
    return out;
  }, true);
}

boot().catch(e => toast(e.message, true));
