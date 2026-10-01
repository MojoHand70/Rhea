# Rhea — founding specification

Rhea is an experiment toward an AI-driven ERP. This document is the design; code follows
it. When the spec is silent, the simplest option wins and goes to `DECISIONS.md`.

## §1 Thesis and non-goals

**Thesis.** An ERP can be built as three strictly separated parts:

1. A **language of ERP** — a small set of concepts (events, objects, documents,
   activities, rules, views) expressive enough to model any domain of a company, from
   finance through sales, procurement and warehousing to production scheduling.
2. An **AI layer** that authors and interprets rules at volume, with a human in the loop.
   This is what makes a pure-metamodel ERP economically viable for the first time: prior
   attempts (REA and its descendants) had the right concepts but no affordable way to
   author rules at the required precision and scale.
3. A **deterministic kernel** that executes rules, enforces invariants, and is the only
   thing that ever changes state. AI output is data, never code.

The system is **AI-first, not AI-only**: humans eye things. Objects materialize on screen
through a library of view notions; the AI and the rules decide *what* exists, views decide
*how it is seen*, and humans approve rules before they act.

The system **extends at runtime**: new object types, rules and views are data loaded into
a running kernel. No stop-the-world deploys to teach it a new document type.

Long-term, multi-country operation is a data problem: a **market pack** is a bundle of
rules, document schemas, tax logic and statutory adapters, not a code fork.

**Non-goals (for the experiment phase).** No product, no auth, no multi-tenant, no
performance work, no UI polish beyond what the shell pattern needs, no promise of
completeness in any domain. Throwaway status is explicit; iteration speed wins.

## §2 The language core

All six concepts are *data*, stored in the system of record, loadable at runtime.

### Event
An immutable fact. The **only** source of state change in the system. Two kinds:
- **raw**: arrived from outside (a document received, a user action, an integration).
- **derived**: emitted by the kernel when a rule fired. Carries provenance:
  `(cause_event_id, rule_id, rule_version)`.

Events are append-only forever. Objects are projections of events; wipe the projections,
replay the log, and identical state must reappear (determinism invariant).

### ObjectType
A schema as data: a name, a domain, a set of typed fields (string, int, money, date,
ref<ObjectType>, enum, list<...>). Money is minor units + currency code; no floats
anywhere. ObjectTypes are versioned like rules.

### Object
A projection of events, typed by an ObjectType. Every object and every field value is
traceable to the (event, rule, rule_version) that produced it. Objects are cached in a
table for serving but are derived state, rebuildable from the log.

### Document
An Object whose type is flagged `document`: it represents a business paper (invoice,
order, delivery note) with an external identity and a lifecycle. Documents are the main
things humans eye.

### Activity
A named action a human or agent can take, declared as data: inputs, the raw event it
emits, who may trigger it. In M0 the only activities are `submit_event` and
`approve_rule`.

### Rule
The unit of system behavior: `match` (a predicate over an event) + `effect` (templates
for derived events / object materialization). Properties:
- versioned, append-only; a change is a new version, never an edit;
- lifecycle `draft → approved → active → superseded`; only `active` rules execute;
  activation is a human approval activity, recorded as an event like everything else;
- effective-dated (`effective_from`, evaluated against the event's business date);
- authored by the AI layer or by hand — the kernel cannot tell the difference and
  validates both identically.

### ViewDef
A view definition: binds an ObjectType (or a query) to a **view notion** with layout
hints. The notion library:

| Notion | What it renders | Milestone |
|---|---|---|
| list | tabular collection with columns, sort, filter | M0 |
| detail | one object: document layout / form / 360 | M0 |
| action | a trigger surface for an Activity (button + inputs + confirm) | M0 |
| analysis | aggregated query result (table/chart) from the read side | M0 |
| scheduling | time-axis placement of objects (plans, capacity) | M5 |
| reconciliation | two collections matched/unmatched side by side | later |

The shell renders ViewDefs generically; there are no hand-written screens per object.

## §3 Boundaries

```
AI layer      →  proposes (rules, object types, views) as strict JSON
human         →  approves / rejects (an Activity, recorded as an event)
kernel        →  validates, stores, executes; the only writer of state
```

- The agent has **no store access and no side effects**. It returns JSON; the kernel
  schema-validates it and stores it as `draft`.
- Beneath all rules sits the **invariant layer** — things no rule may override and the
  AI can never author around: append-only log, rule versioning, provenance on every
  fact, determinism on replay, money as integers. Domain invariants (double-entry
  balance, gapless statutory numbering, period locks) join this layer as domains arrive
  (M1+).

## §4 Storage

- **PostgreSQL** — the system of record and write side: `event` (append-only), `rule`,
  `object_type`, `view_def` (all versioned, append-only), plus the `object` projection
  cache (rebuildable, the one non-append table, owned exclusively by the kernel's
  replayer/executor).
- **DuckDB** — the read/analysis side: projections built by replaying the event log;
  serves analysis views. Later: what-if simulation ("replay August under the proposed
  rule version before approving it").

## §5 Shell

A single web shell, served by the kernel (`net/http`, embedded vanilla HTML/JS/CSS, no
build step):

- **Activity bar** (left edge): one icon per domain (M0: Finance).
- **Submenu** (panel next to it): the selected domain's functions (M0: Documents,
  Rules, Analysis, Worklist).
- **Tabs** (main area): each opened function or object becomes a tab; tab content is an
  object-tailored editor rendered generically from a ViewDef.

The shell knows nothing about invoices; it knows how to render the six notions.

## §6 Milestone ladder

- **M0 — the vertical slice.** One domain (Finance), one flow: raw invoice event lands
  in worklist → user asks for a rule in plain language → agent drafts Rule JSON →
  human approves in the shell → kernel materializes an Invoice document → list +
  detail views show it → DuckDB analysis view aggregates it → replay reproduces state.
- **M1 — rule simulation + finance depth.** Dry-run a draft rule against the historical
  log in DuckDB with a diff view; double-entry sub-language and its invariants
  (balance, period lock) ported from the Sunbeetle lessons.
- **M2 — the falsifiability test.** Express a genuinely distant domain (warehousing:
  goods receipts, stock, locations) **without any kernel changes**. If the kernel must
  change, the language failed and gets redesigned before going further.
- **M3 — first market pack.** Poland: CoA template, VAT rules, KSeF document schema as
  data; one statutory adapter as code behind a declared contract.
- **M4 — second market.** The proof that M3's pack boundary was real.
- **M5 — scheduling.** Production/planning objects and the scheduling view notion.

## §7 Open questions (parked, not blocking)

- How closely to align the language core with REA's resource/event/agent/commitment
  vocabulary, vs. staying pragmatic.
- Schema evolution of live objects when an ObjectType gets a new version (migration
  rules? lazy upgrade on replay?).
- The adapter contract for statutory integrations (sync/async, retry, evidence storage).
- Rule conflict detection when multiple active rules match one event (M0: first match by
  priority, recorded in DECISIONS.md).
