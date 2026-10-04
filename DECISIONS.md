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

## 2026-10-04 — E2 attempt: two markets in one kernel fail as pure data

- **The attempt** (`TestCohabitation`): both market packs loaded into one
  kernel over one log, all rules approved. It fails three ways, and the
  worst two are silent:
- **Silent capture.** Packs register master data from shared neutral event
  types (`account.created`, `vat_rate.defined`) with no market binding, and
  approval re-evaluates eagerly — so whichever market's registration rule
  is approved first claims the other market's events too. The German SKR03
  materializes with `pl-register-account` as its explanation. Provenance
  records the mis-explanation honestly; nothing flags it.
- **Silent misrouting.** The packs route invoices by currency (`$.currency
  eq PLN`/`EUR`) — disjoint guards, so the same-id conflict that would at
  least refuse loudly never fires. A Polish company's EUR-denominated
  domestic invoice books cleanly into the German SKR03. Currency is not a
  market, and it is certainly not a company.
- **Global ref namespace.** Both packs ship a VAT rate coded "0";
  `=ref(vat_rate, code, ...)` resolves globally, so a 0% invoice's rate is
  ambiguous and the event waits forever.
- **Verdict: E2 passed as pure data — an M2-style negative proof, nothing
  earned.** The fix is pack v2s and one seed: every rule matching a shared
  neutral event type carries a `$.market` where-guard (the binding of pack
  rules to a market is rule data); `sales_invoice` v2 makes the seller a
  `ref<company>` filled by `=ref(company, vat_id, $.seller_nip)` (the
  binding of a document to its company is a ref); posting rules name their
  book (`pl-stat`/`de-stat`), so the E1 machinery is the per-company close;
  the statutory registers join the seller's country (the read side CAN
  follow refs today — only the match side cannot); and de's zero rate is
  recoded `0-de` because resolution keys of a shared type are one global
  namespace (the E1 account-codes decision, extended). An event naming no
  market matches nothing and waits — unroutable beats misrouted.
- **Contortions carried, with named exits**: `market` is denormalized onto
  every payload because match conditions cannot read through refs (the
  seller's country) — dies when ref-reads arrive (DIRECTION, arithmetic
  destiny list). One company per market per book, because book names
  cannot compose (company × GAAP) — a second same-market company forces
  that question. A 0% invoice still cannot post (the VAT line's amount
  must be positive; conditional lines are inexpressible) — postings
  sub-language residue, left for its own fail-as-data. The KSeF adapter
  still reads all invoices; binding adapters per market joins the adapter
  contract's open questions.

## 2026-10-04 — E1: parallel books fail as pure data; `book` earns admission

- **The attempt** (`TestParallelBooksAsPureData`): a group-GAAP rule-book
  beside the statutory one, over the same `sale.recorded` events, the group
  CoA carried as accounts with prefixed codes. Posting *works* — multi-rule
  firing books both entries from one event — but the book is first-class
  nowhere: identity lives in account-code prefixes, exactly the rule-shape
  contortion DIRECTION's network note warns about, and a trial balance can
  only split books by joining codes back.
- **The fail: independent period closes are inexpressible.** `period_lock`
  is month-only and the kernel check is global: closing the statutory
  September locks the group book too, and the lock object cannot even say
  which book it means. Atomic multi-rule firing makes it total — one locked
  book refuses the event for every book.
- **Verdict: `book` earns admission into the postings sub-language** — the
  first closed-class change since cascade, judged by the vocabulary-
  governance rules: fail-as-data (this entry), invariant (book is declared
  at entry level so an entry cannot straddle books by construction; the
  lock invariant becomes per (book, month)), determinism (lock effects bake
  into the log at firing time; replay untouched), network test (one small
  field on two conventional types).
- **Scope held back deliberately**: accounts stay one population with
  globally unique codes — group CoAs have their own code schemes in
  practice, so book-scoped account resolution waits for a real collision
  to demand it (its own fail-as-data, when it comes).
- **Shipped semantics** (`TestParallelBooks`): `postings.book` is an
  optional entry-level template defaulting to `"main"`; every posting
  carries its book first-class; `period_lock` names its (book, month) and
  the kernel check intersects two single-field lookups, so the Lookup
  contract stays one field. A pre-book lock object matches no book and is
  inert from here on — history is untouched because lock effects bake into
  the log at firing time. A mixed event matching an open and a closed book
  refuses whole (never half-explained); the open-book half is a separate
  event in an open period — korekta territory, per the backfill stance.

## 2026-10-03 — M4: the second market proved the pack boundary

- **Germany is one data file** (`packs/de`): SKR03 chart, 19/7/0 rates,
  1400/8400/1776 booking, German-labeled views. The pack file and its test
  are the entire diff — kernel, pack loader and adapter contract are
  byte-for-byte what Poland runs on.
- **M4 forced one base correction, which is exactly its job**: `vat_rate`
  and `sales_invoice` sat in the PL pack but are market-neutral, and a
  German pack requiring Poland would have been a fake boundary. They moved
  to the finance base (`finance_v4` seed). Germany ships **zero object
  types**: a market can be pure rules + master data + views.
- **Cross-market exclusivity is explicit in match predicates** (the
  multi-fire idiom): PL's document rule claims PLN, DE's claims EUR; a PLN
  invoice in a German-only installation waits in the worklist instead of
  booking by accident.
- **Parked, recorded: true cohabitation of two markets in one ledger.** The
  packs' register-account rules are identical and would collide if both were
  approved (approve one); both markets' analysis views aggregate the shared
  `sales_invoice` type. Company-/market-scoped events wait until
  cohabitation is a real goal, not before.

## 2026-10-03 — M3 complete: the statutory adapter contract

- **An adapter is a user, not a kernel extension** (Alpha's "the agent is a
  user", applied to integrations). The declared contract
  (`internal/adapter`): read projected objects, talk to the outside world
  through your own client, return raw events; the runner appends them under
  actor `adapter:<name>`. The kernel never calls out, and replay never
  re-runs an adapter — its events are already in the log. This answers SPEC
  §7's parked questions: *async* by construction (a pass polls the projected
  world for what the outside world still owes the log, and pending statutory
  state is ordinary objects from ordinary pack rules); *retry* is free
  (evidence events carry deterministic dedup keys — `ksef/submit/<invoice
  id>` — so a crashed pass re-submits into a no-op); *evidence* lives in the
  event payload (KSeF reference + UPO) — the append-only log is the evidence
  store.
- **The KSeF client is an interface; the experiment ships a deterministic
  fake** (reference and UPO derived from the invoice number), so demos and
  replays reproduce. `rhea ksef` runs one pass.
- **Renamed before anything consumed it**: `ksef_invoice` →
  `sales_invoice`, event `ksef.invoice.received` → `sales.invoice.issued`,
  and the document no longer carries `ksef_ref` — the invoice is ours at
  issue time; the KSeF reference is *evidence of submission* and belongs on
  the `ksef_submission` object the adapter's event materializes. An invoice
  is "in KSeF" only if the adapter ran: exactly the statutory truth.

## 2026-10-03 — M3: the Poland pack; a market pack is one data file

- **Pack = one JSON manifest** (`packs/pl/pack.json`): object types, view
  defs, rules and master-data events, plus a `requires` list naming the base
  types it builds on (company, account, posting) without shipping them — base
  definitions belong to the base, and M4's second market proves the boundary
  by reusing them. Loaded by `rhea pack FILE` or `pack.Load`.
- **A pack installs no behavior.** Its rules land as *drafts*; its events (the
  chart of accounts, VAT rates) wait in the worklist; approving the pack's
  rules in the shell IS the installation. The approval gate needed no new
  machinery to become a pack-install UX.
- **Loading is idempotent**: every pack event must carry a dedup key
  (`pack/pl/account/201`); types and views skip on an existing
  (name, version); a rule id the system already knows is left alone — pack
  updates will arrive as pack version bumps, not silent redrafts.
- **The invariant layer audits the pack's own data**: the KSeF sales entry
  books 201 gross / 700 net / 222 VAT, so an invoice whose gross ≠ net + VAT
  is refused by the double-entry balance check — statutory arithmetic
  enforced by a domain invariant, not by pack code.
- **Localization is data**: Polish titles and labels in the pack's views;
  `function` keys stay English since they are grouping keys shared with base
  submenus. Wzorcowy plan kont mapping to the ten account types: Kasa and
  Rachunek bieżący are `bank`, VAT and rozrachunki publicznoprawne `tax`,
  Rozliczenie zakupu/kosztów `clearing`, zespół 4 `expense`, zespół 7
  `revenue`.
- **Still open in M3**: the statutory adapter (KSeF submission) as code
  behind a declared contract — the SPEC §7 adapter-contract question.

## 2026-10-03 — cause-qualified cascade identity

- **The forcing domain arrived the same day**: `each` × cascade. Two line
  movements from one receipt both fire the valuation rule, and per-root
  posting ids collided — refusing exactly the story the cascade exists for
  (receipt → N movements → N valuation entries). The open edge below is
  hereby closed.
- **A cascaded firing's ids build on the causing object's id, not the root
  event's**: objects `<type>-<cause object id>`, postings
  `posting-<cause id>-<rule>-<n>`, entries `entry-<cause id>-<rule>`. Root
  firings keep `<type>-<root event id>` — every pre-existing id is unchanged.
  The causing object's id is itself rooted, so identity still needs no
  sequence state and simulation still reproduces cascaded ids exactly.
- **The traded-away guard, as recorded**: cascade cycles no longer
  self-collide (each generation mints a fresh qualified id), so the
  16-generation depth cap is now the loop guard for cascaded rules. Its
  error names the looping rule and the id it keeps materializing — the id
  shows the loop, one type name per generation. Collision still guards every
  same-cause and root-level ambiguity.
- **Each line's valuation is its own balanced journal entry** (per-cause
  entry keys); balance stays enforced per firing, as always.

## 2026-10-03 — multi-line documents: lines are objects, via `each`

- **Lines stay objects; `list<>` fields are deliberately not built.** The fork
  was real: SPEC §2 names `list<...>`, but embedding lines in a document field
  would have cost per-line provenance (invariant 5), needed a new detail-view
  renderer, new validation, and JSON-array gymnastics on the analysis side —
  while M1 already decided "postings are an effect, lines are objects" and got
  all of that for free. So the posting pattern is generalized instead: an
  object effect may set `each`, a `=$.path` fan-out, and materializes one
  object per array element. Revisit `list<>` only if display demands true
  embedding.
- **Each-templates see `{doc, line, n}`** — the whole event payload, the
  element, the 1-based line number — explicitly scoped, no merge magic.
  Refs resolve per line (`=ref(item, sku, $.line.item)`).
- **Line ids are rooted and numbered**: `<type>-<root event id>-<n>`. Two
  each-rules claiming the same type collide on line 1 like any same-id
  ambiguity and ask a human. An empty fan-out is a rule error — a document
  with no lines is malformed, and the whole event waits in the worklist (the
  header does not book either; atomicity as everywhere).
- **Parent linkage is a business key, not a synthetic ref**: line objects
  carry the document's natural key (`grn`, invoice number) from the payload.
  A kernel-made parent ref would need a new primitive and sibling visibility
  at generation 0; the paper world's own key costs nothing. Revisit when a
  domain has genuinely keyless documents.
- **Open edge, recorded**: a cascade rule firing on several sibling derived
  events of one root still collides (rooted ids are per root, not per cause).
  Valuing each line's movement via a cascaded postings rule will demand
  cause-qualified identity — which would trade away collision-as-loop-guard
  (the depth cap would become the real guard). Decide when a domain forces
  it, not before.

## 2026-10-03 — rule cascade

- **Rules match derived events now — with zero language changes.** A cascade
  rule is an ordinary rule whose `match.event_type` is `object.materialized`,
  with conditions over the materialization payload (`$.object_type`,
  `$.state.*`). The kernel evaluates the rule set against each firing's
  derived event, generation by generation, until nothing fires: receipt →
  stock movement → valuation postings. All-matching-fire was the stepping
  stone; this is the recorded end-state.
- **The whole chain books atomically**, in one transaction with the root
  event's own firings — a locked period on the valuation refuses the movement
  too; an event is never half-explained. Cascaded derived events name the
  intermediate materialization event as `cause_event_id`, so provenance walks
  back to the root through the log; the object cache carries the same source.
- **Cascaded object identity stays rooted in the raw event**
  (`<type>-<root id>`, postings `posting-<root>-<rule>-<n>`), not in the
  intermediate event's id. Two reasons: simulation reproduces live identity
  exactly without sequence state, and a cyclic rule set re-claims ids it
  already owns — cycles surface as same-id conflicts for a human, not as
  recursion. A depth cap (16 generations) remains as a backstop with a clear
  error, not as the loop prevention. One object of a type per root event is
  the semantics; a rule firing on two sibling derived events collides by id
  and asks a human.
- **Each generation sees state as of before the root event plus all earlier
  generations, never its own siblings** — so a cascade rule's `ref()` can
  resolve the very movement created one generation up, while sibling rules
  stay order-independent. One `expandChain` is shared verbatim by the live
  executor and the simulator (the simulator's duplicated firing loop is
  gone); the simulator attributes cascaded objects to the root event, since
  intermediate events have no ids until actually booked.
- **Money inside derived payloads is minor units as numbers.** `Template.Eval`
  on money fields now takes integral numbers as minor units; decimal strings
  remain the boundary form for raw events. Non-integral numbers are errors.
- **The `object.materialized` namespace is the kernel's**: raw events may not
  use it (store-level guard), or a submitted fake could trigger cascade rules
  for facts that never happened.
- **Backfill stays open.** Cascade did not add retroactivity: a cascade rule
  approved later never re-fires already-explained events, same as before.

## 2026-10-02 — M2: the falsifiability test passed

- **Warehousing runs on the unchanged kernel.** Locations and items are master
  data, receipts and issues are documents, stock movements are objects,
  stock-on-hand is an analysis view; a goods receipt multi-fires into a
  document plus a movement; the supplier is a cross-domain `ref<company>`.
  The M2 commit adds only a seed file, five events and a test — `git diff`
  against the kernel is empty. The language held.
- **Single-line goods documents** for now (one item per receipt/issue) — the
  language still has no `list<>` fields; same simplification as invoices
  having no line-level objects. Lines are the likely next language need.
- **No retroactivity**: an event is explained by the rules active when it was
  evaluated; a rule approved later never re-fires old events. Multi-fire makes
  the sequencing visible — approve the full rule set before the events arrive,
  or let blocked refs hold events back until the set is complete. Backfill
  belongs with the cascade design (post-M2), not before it.
- **Stock may go negative.** No stock invariant yet; that is worklist/case
  territory, not a kernel rule.

## 2026-10-02 — multi-rule firing

- **Every matching rule fires; the event books atomically.** First-match-wins
  (M0 placeholder) is superseded: an invoice event becomes a document AND a
  ledger entry in one transaction, all firings or none. Priority now means
  order, not conflict resolution. Expansions see state as of before the
  event; intra-event dependencies wait for cascade.
- **Cascade (rules matching derived events, with atomicity) is the intended
  end-state — KK's call.** All-matching-fire is the stepping stone; revisit
  after M2, which will likely demand it (receipt → stock move → valuation).
- **Collisions**: two rules materializing the same object id refuse the whole
  event, which waits for a human — for documents and master data that is the
  semantics (one object per type per event, two claimants = real ambiguity).
  Posting ids and entry keys are rule-qualified (`posting-<event>-<rule>-<n>`)
  so several ledger rules may book one event (VAT beside revenue) without
  false conflict. Semantic double-booking across types is invisible
  structurally — surfacing it pre-approval is what simulation is for; static
  conflict detection stays parked (SPEC §7).
- **Exclusivity is explicit in match predicates now** (PLN vs ne-PLN). The
  fallback-by-priority idiom is gone deliberately: silently suppressing a
  matching rule is how ambiguity hides. An overlapping fallback on the same
  type collides by id and asks a human.

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
