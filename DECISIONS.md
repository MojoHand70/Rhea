# DECISIONS.md

One line per decision: date, decision, why. Where `SPEC.md` is silent, choices land here
so the next session does not re-derive them.

## 2026-10-01 — bootstrap

- **Postgres runs in Docker (`rhea-postgres`, postgres:17, localhost:5432, volume
  `rhea-pgdata`), not via apt.** Docker is already the established pattern on this
  machine (novi stack runs in it) and keeps the experiment removable; the plan's apt
  option is superseded. Credentials db/user/password `rhea` — local-only experiment.
- **Rule conflict resolution: first match wins by explicit integer `priority`
  (ascending), ties broken by rule id.** SPEC §7 parks real conflict detection; M0 needs
  something deterministic and dumb.
- **Business date for rule effectivity is the event's `occurred_at` date**, carried by
  every raw event; derived events inherit the causing event's business date.

## 2026-10-01 — milestone 0

- **The `rule.*` event-type namespace is system activities** (approvals); the worklist
  query excludes it, so approvals do not sit forever as "unexplained" business events.
- **Money display formatting stays in integer arithmetic end to end**: Go formats minor
  units with `core.FormatMoney`; analysis SQL uses `printf('%d.%02d', x // 100, x % 100)`.
  The DuckDB driver hands DECIMAL back as a float and trims trailing zeros, so DECIMAL
  never crosses into display.
- **Analysis views rebuild the DuckDB projection on every request.** Always fresh, and
  at experiment scale the rebuild costs nothing; incremental projection is an M1+ concern.
- **Shell listens on :8070** — 8080/8083–8085/8091 are taken by the novi stack on this box.
- **Object identity is `<type>-<source_event_id>`**, derived from the log so replay
  reproduces it without any sequence state.
- **A list view finds its detail view by matching `object_type`** across view defs; the
  first match wins. Good enough until a type legitimately has two detail views.

## 2026-10-01 — refs and master data

- **Master data is objects materialized from events**, same as documents: `company` is
  an ObjectType, `company.registered` a raw event, a normal rule books it. No special
  master-data machinery; counterparty vs operating company is the `kind` enum
  (customer|supplier|self). Multi-tenant stays out (experiment non-goal); multi-company
  will be a `ref<company>` field when a domain needs it.
- **`ref<T>` fields hold the target's object_id and are filled only by
  `=ref(T, field, $.path)`**, which resolves against the object cache at fire time.
  Unresolved or ambiguous refs abort the firing and the event waits in the worklist —
  that *is* the unknown-master-data flow. Resolved ids are baked into derived events, so
  replay never re-resolves and determinism is untouched. Direct-id templates are
  rejected: referential integrity only through resolution.
- **`ProcessPending` iterates to a fixpoint** (passes until nothing books), so master
  data and its dependents can arrive in any order or the same batch; errors reported
  from the final pass only.
- **ObjectType gains optional `label_field`**; the shell renders a ref as the target's
  label, falling back to the raw value (which also covers v1 objects whose field
  predates the ref). Ref targets are validated syntactically at load, not for
  existence — definition files may load in any order.
- **Old-schema objects stay as they are**: invoice v1 rows keep the customer name
  string; the v2 analysis view joins on refs and so aggregates only v2 invoices.
  Schema evolution of live objects remains parked (SPEC §7).

## 2026-10-01 — double-entry sub-language (M1)

- **Postings are an effect, lines are objects.** `effect.postings` (currency +
  lines of account / one-of-debit-credit money templates) expands into
  `posting` objects — one per line, grouped by an `entry-<event_id>` key, via
  the ordinary derived-event path. Replay and simulation needed no changes;
  provenance per line comes free (Sunbeetle invariant 3). No `list<>` field
  type yet; one object per line avoids it.
- **Sunbeetle ports**: balance ΣD=ΣC checked after expansion, before booking —
  an unbalanced expansion is a rule error and nothing books; the period lock
  is a pre-insert check in the poster (Go, one place); account types take
  Sunbeetle's full ten (debtor/creditor etc. — AR/AP are filters, not tables),
  which costs nothing since enums are data.
- **Accounts resolve by `code`** — a line's account is a code template; the
  kernel resolves it against `account` objects (chart of accounts = master
  data from events, like companies). `posting` / `account` / `period_lock`
  are conventional type names the kernel knows; their definitions are seeded
  data like any other.
- **Period locks are events** (`period.locked` → `period_lock` object via a
  normal rule); a posting whose business month has a lock object is refused
  and waits in the worklist. No unlock in M1 — the language has no state
  updates, and re-opening a period is exactly the kind of thing that should
  hurt. One currency per entry; no base-currency translation yet.
- **Lookup returns all matches now** — ref() demands exactly one, the period
  lock asks "any?"; one primitive serves both, in store and simulator alike.

## 2026-10-01 — rule simulation (M1)

- **Simulation runs in the kernel (Go), not DuckDB.** SPEC M1 says "dry-run in
  DuckDB", but the simulator reuses the executor's own match/expand/ref path
  (`matchRule`, `Expand`), so a dry run cannot drift from live firing. DuckDB
  stays the analysis read side. Recorded as a deliberate deviation.
- **The diff is replay vs replay**: all raw business events through the active
  rules, then through active + the candidate's latest version, both fully in
  memory with refs resolving against the simulated world. Diffing against the
  *actual* cache would mix the candidate's effect with drift from superseded
  rule versions; replay-vs-replay isolates the rule being judged.
- **Firing-order bug fixed on the way**: `latest_rules.sql` ordered rules by
  `rule_id` (a DISTINCT ON artifact), not the decided priority-ascending
  order. An outer ORDER BY now enforces it; the simulator sorts identically.

- **Drafting a new version must not suspend the running one** — found by the
  first live simulation: `ActiveRules` used to read each rule's *latest*
  version's status, so a pending draft silently deactivated the rule. The
  executing set is now the newest **active** version per rule
  (`active_rules.sql`); only a latest version of `superseded` retires a rule.

## 2026-10-01 — actor attribution

- **Every event names its actor**: `cli:<os user>` (override `RHEA_ACTOR`),
  `shell` (until identity exists), `agent:<model>` on drafted rule rows,
  `kernel` on derived events (rule provenance explains the rest), and the
  approver's name on `rule.approved`. NULL on events from before the column
  existed — honestly unknown, never backfilled. Authz stays deferred; this is
  audit, not access control.
