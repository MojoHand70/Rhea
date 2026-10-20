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

## 2026-10-09 — the first measurement against the bar: both simulated customers DONE in every run

- **The measurement:** each persona, four voices (the scripted interview
  and three model paraphrases), three runs each, one month of documents,
  the verdict all or nothing. Eight rounds in all — Helios six, Nordwind
  two — about a hundred live runs and seven hundred interview answers
  drafted by claude-sonnet-4-6 and judged against the generator's
  rule-free expected state.
- **Verdicts, round by round.** Helios: 6, 11, 11, 8, 10, then **12 of
  12**. Nordwind: 3, then **12 of 12**. Every count, every lifecycle
  move and every computed sum matched in the final rounds: VAT, gross,
  PZ line values, book stock from the fold.
- **The arithmetic was never wrong, in any run.** Not one formula was
  refused by the typed gate or the VM, and no computed sum ever came
  out wrong when the formula fired. Every loss was vocabulary or
  reach, and each was fixed once, where it belonged:
  1. an unrequested verb with an enum lacking values (prompt: no verb
     unless people record by hand);
  2. a rule setting a field its type lacked (prompt, then the seed:
     a purchase invoice carries its VAT rate, optional);
  3. the book named in the rule id and left off the entry, three times
     in three under one phrasing (prompt: a named book goes on every
     entry);
  4. `$.root` unreadable inside `each` (kernel: the chain's root stays
     readable inside each, as on the cascade payload);
  5. fields named `book_qty` where the expectation said `book` (the
     owner's answers name the fields that will be checked, as the
     operations corpus already did);
  6. bank receipts booked to accounts the chart does not have, which
     nobody asked for (prompt: post only what the question asks, never
     to an absent account);
  7. `ref(company, kind, self)` with the literal unquoted (prompt: the
     quoted-literal idiom, our own company);
  8. `$.vat_rate.percent` read through a raw code string (prompt: a raw
     value is text, never a link; resolve first);
  9. a withdrawal rule matching a re-materialization that never happens
     because amendments end chains — approved on a dry run that
     honestly said "explains 0, adds 0" (prompt: an amended status
     never re-materializes; react to the raw event).
- **What this says about the loop.** The gate gets stricter with every
  wrong draft, and the all-or-nothing verdict is what makes a one-in-
  twelve habit visible at all; averaged, every round after the first
  would have looked fine. Nine habits in a day is the shape of the
  work now: not more language, but the author learning its grammar's
  edges, one refusal at a time. The paraphrase guard needed three
  turns of its own to tell a fact from a hyphen.
- **The recorded approver reads the dry run now** (KK asked whether the
  "policy" was a rule or code: code, in the eval only, standing in for
  a person; never a kernel rule, since definitions alone legitimately
  change nothing). It refuses a rule whose dry run explains nothing,
  adds nothing and changes nothing *while what it matches is already
  there* — raw events of its type in the log, objects of the type it
  cascades from in state. A rule for the future passes: the simple
  corpus approves the invoice posting cascade before any invoice is a
  document, and a person would too. (9) is caught by this. The seeds
  now carry two general-case fields (`mirror_of` optional, `vat_rate`)
  the intercompany ladder had left out.

## 2026-10-09 — the eval's verdict is all or nothing

- **KK: "In accounting there is no such thing as a document that passes
  one time and not the other."** Right, and the variance never was in
  booking: a rule version on an event yields the same derived event every
  time, baked, and replay proves it. The variance is in authoring — the
  model drafting from a plain-language answer, once, behind the gate. A
  draft that is right most of the time is a defect to drive to zero, not
  a statistic to average; my earlier "read the metric over several runs"
  was wrong in spirit.
- **The rule:** `rhea eval -runs N` repeats a corpus on fresh companies,
  `-voices` runs it once per saved paraphrase, and the verdict
  (`eval.Verdict`) is DONE only when every run in every voice is DONE —
  "NOT DONE: 2 of 3 runs done — an author that passes most of the time is
  not done". The growth plan's metric reads against this bar: the share
  of the story that passes in every run.
- **What drives authoring variance to zero** is already the design, and
  each wrong draft sharpens it: the gate gets stricter (today: names,
  warrants, optional fields); a known question gets the known answer (the
  network reuses, packs pre-fill); the human approves what remains.
- **Refused drafts stay readable**: a draft the gate refuses comes back
  with its error (`agent.DraftRule`, `DraftBundle`) and the eval keeps it
  on the row, printed under "refused" in verbose mode — an authoring
  failure can be read without a rerun.

## 2026-10-09 — first live runs: two simulated customers, DONE

- **Five live runs (claude-sonnet-4-6, one month of each persona), the
  formula language's live proof.** In every draft the model wrote the
  arithmetic exactly as the kernel wants it: VAT as `round($.net *
  ref(vat_rate, code, $.vat_rate).percent / 100, 2, half_up)`, gross as
  net plus that, the due date as `$.issue_date + $.payment_days`, the PZ
  line value as qty × price, and the stock count's book quantity as a
  fold over `objects(stock_movement, …)` filtered by location and date.
  Not one formula was refused by the typed gate or by the VM. The
  language is authorable from a plain-language answer, first time.
- **What failed was vocabulary, three times, each fixed once:**
  1. The seeded `purchase_invoice` required `mirror_of`, E4's
     intercompany link, so a plain supplier invoice could not exist; the
     model tried to resolve it against a sales invoice, then a PZ. The
     field is optional now (`finance_v8`): a purchase invoice mirrors a
     group sales invoice only when there is one. The prompt also says a
     type may be declared again without a field the client's documents
     cannot fill.
  2. The model claimed warrant basis `network` with no network connected
     (Rhea refused: no support could be counted). The prompt now says
     the basis exists only when the request lists learned answers.
  3. The model named new types its own way (`pz_document`) and filled
     an optional ref with an empty string. The owner's answers now name
     the types, as the operations corpus already did, and the prompt
     says to leave optional fields out rather than fill them with
     nothing. The run count: Helios 5/6, 5/6, 5/6, then 6/6; Nordwind
     6/9, then 9/9 — every count, lifecycle move and computed sum
     matching the generator's rule-free expectation.
- **The eval does not print a draft the gate refused** (it is never
  stored); the refusal message had to say enough. It did. A verbose
  refused-draft line is cheap and would have saved one run — noted, not
  done.
- **Variance is real and the measure is the share of runs** — the same
  persona passed and failed the same task on different runs before the
  fixes. The growth plan's metric (share of the story passing unattended)
  should be read over several runs and several voices, which is what
  `-voice` is for. Paraphrases not yet generated; next live session.

## 2026-10-09 — simulated customers: a business as data, ground truth without the model

- **KK (2026-10-08): "Can we simulate customers?" — yes, the third way.**
  A persona (`testdata/customers/<name>.json`, `internal/sim`) is a
  business as data: its company, customers and suppliers, catalogue,
  chart, VAT rates, monthly volumes, booking policy as line counts, and
  the interview — every answer with its known-good reference bundle. A
  deterministic generator (seeded PCG) turns it into a year of documents
  and the state a correct implementation ends in, computed without any
  rule: counts per type, lifecycle moves (paid invoices, resolved cases),
  and sums of computed values (VAT, gross, PZ value, book stock). The
  model's only role is paraphrasing the answers; facts never come from
  it, and the scripted expected state judges every run.
- **Two personas from the demo story.** `nordwind` (trading: deliveries
  → PZ with valued lines, stock in and out, purchase and sales invoices
  with VAT computed and rounded, bank lines settling invoices, complaints
  and cases, two stock counts whose book quantity is a fold over the
  movements up to the count date) and `helios` (services: no warehouse,
  costs on 402, revenue on 701). Same questions, different answers — the
  network's closing scene has its two subscribers. Nine months of
  Nordwind are 262 events; the reference test plays two months of each
  through the real kernel to DONE.
- **The corpus gained `expect_sum`**: totals of one field over a type,
  money as decimal strings — the check a count cannot make on a computed
  value. `rhea eval` accepts a persona file (`-voice N` picks a
  paraphrase, `-months N` shortens the year); `rhea sim PERSONA` prints
  what it generates, `-corpus FILE` writes the compiled corpus,
  `-paraphrase N` asks the model and saves the voices beside the persona
  (`<name>.voices.json`, checked token by token: every number, code and
  status word must survive), `-play` submits the documents into the live
  installation through the door, dedup-keyed, and prints the interview as
  the demo's script.
- **Synthetic installations never count.** Publications from a `sim:`
  installation are stored marked (`publication.synthetic`); `Learn`
  leaves them out, `LearnIncludingSynthetic` is what a simulated run
  draws on, `rhea network -synthetic` shows them. A hundred simulated
  businesses agreeing would be an invented statistic (DIRECTION, "Rhea
  suggests standards"); the test asserts a real client's shape never
  surfaces on their strength. Plain-corpus eval runs still publish as
  real, as the five flywheel runs did — unchanged, recorded.
- **Found on the way:** rules take effect with the period, so a
  persona's master data is dated the first day, not the day before;
  deliveries land early in the month so stock is never issued before it
  arrived in date order.
- **Not run live** (no key this session): the first `rhea eval
  testdata/customers/nordwind.json` with the model is the formula
  language's live proof, and `-paraphrase 3` then `-voice 1..3` is the
  phrasing test. Next for the simulated customer: the generator emitting
  the documents themselves (a KSeF file, an MT940 line, a scanned PDF)
  when the intake brief lands, so one persona feeds intake and
  calculation alike.

## 2026-10-08 — the formula language: arithmetic as data, explained by its inputs

- **Every `=` template is a formula** (`core/formula.go`, `core/vm.go`):
  tokenizer, parser, static checker, compiler to a small instruction set,
  and a stack VM — the sfmt architecture ported, with Rhea's number core.
  The template kinds the validator knew stay as laws: a `ref<T>` field
  takes a bare path or a `ref()`, never a literal and never a computed id.
  Everything else computes. Existing templates mean exactly what they did.
- **Numbers are exact rationals (`math/big`); money is sticky.** Kinds:
  int, money, decimal (a new `decimal` field type: rates, unit costs),
  string, date, bool, list, object, ref. money × int → money, money ÷
  money → decimal, money × money refused, int ÷ int → decimal. No floats:
  a non-integral JSON number is refused, decimals travel as strings.
- **Rounding is declared, never implied.** `/` is admitted only under a
  `round(x, places, method)` somewhere above it — a static law. A result
  stored into a money field must fit two places, into an int field a whole
  number: a three-place unit cost times a quantity refuses the firing into
  the worklist with "declare round(…, 2, half_up)". Methods: `half_up`
  (convert's: half away from zero), `half_even`, `down` (toward zero),
  `up` (away from zero) — others join by proof. Places capped at six.
- **Reads through links are typed by the catalog.** `$.state.item.std_cost`
  follows the ref the causing object holds; `ref(item, sku,
  $.line.item).std_cost` reads after a lookup. The executor names which
  payload prefixes hold typed state (`$.state` on a cascade, `$.doc.state`
  inside `each`), so money reads as money and the walk continues through
  ref fields via the Getter. Untyped raw reads are what they say: a number
  is a number, a string is a decimal in waiting. A bare `=$.path` copy
  keeps every coercion it had — minor-unit numbers on money fields
  included — so nothing existing changed meaning.
- **Folds, never loops.** `sum/count/min/max/any/all(v in coll [where p]
  [: e])` and one `fold(v in coll, acc = init: e)`. Collections are the
  payload's own lists or `objects(T, field, value)`: a keyed read of state
  in log order (`FindObjectIDsByField` now orders by `source_event_id`),
  refused above the cap (`core.DefaultMaxCollection` = 1000;
  `Executor.MaxCollection` lowers it in tests, nothing raises it per rule).
  A formula declares its reads by writing them (`Formula.Reads()`). No
  while, no recursion, no assignment; the only backward jump is a fold's,
  and a step budget guards the mechanism itself.
- **Every computed value is explained in the log.** `MaterializedObject.Calc`
  and `AmendedObject.Calc` (and a posting's computed amount) carry, per
  computed field, the formula as written and every input read by concrete
  path — `$.lines[2].qty`, `$.state.item` → the id, `$.state.item.std_cost`,
  `ref(vat_rate, code, 23)` → the id, `objects(stock_movement, item,
  item-7)` → the ids. Baked at firing, so replay never recomputes (scale
  rule 3) and the walk — `/api/explain` and the shell — shows "value =
  formula where inputs". Copies carry no calc.
- **The typed draft gate.** `RuleSpec.Check(catalog)` runs wherever a
  catalog exists — the agent's rule and bundle validation, the pack loader
  — and refuses unknown fields through links, unlawful kinds (money ×
  money, a declared string in arithmetic, an amount into a date field),
  undeclared rounding, malformed folds. `Validate(target)` keeps the
  data-free laws for every other caller.
- **First customers pass** (`exec/formula_test.go`): the PZ line valued at
  qty × the delivery's price; the PZ entry Wn 330 / Ma 300 at the item's
  standard cost read through the line's link, rounded on purpose; VAT as
  round(net × rate / 100, 2, half_up) with the rate read after a lookup;
  the due date as issue date + payment days (the strain recorded in the
  cases tests on 2026-10-05 is gone: `=$.state.registered_on + 14`); the
  stock count's book quantity as a fold over the item's movements in the
  location, the cap refusing. Replay identical.
- **The agent** reads the grammar with examples (PZ, VAT, due date, stock
  on hand). The operations corpus task 4 now values each PZ line at
  quantity × the delivery's unit price; the reference passes offline. No
  live run this session (no key).
- **Deferred, named:** formula → SQL for DuckDB analysis views (sfmt 074's
  shape) when an analysis view needs one; kernel methods callable from a
  formula (`fifo_cost(item, warehouse, qty)`) when FIFO arrives in phase 4;
  a per-rule cap only if a real collection needs it.

## 2026-10-08 — the flywheel, live: what five runs taught

- **Five live runs of the operations corpus with `rhea eval -network`**
  (claude-sonnet-4-6), each run publishing as its own installation into
  `rhea_network`.
  - Run 1 drew on nothing, run 2 on one installation (below the floor),
    and run 3 onward on real priors.
  - **Consensus forms fast:** after three runs, 16 of 19 questions had
    one answer given identically by every installation (3 of 3; 5 of 5
    after five runs): master data, invoices into both books, goods flows,
    complaints and cases, and the withdrawal.
  - The model is consistent enough that shape fingerprints converge
    without any semantic normalization.
- **Learned knowledge is applied at the first opportunity.** With
  priors, the agent explains residue the network knows about before the
  question asks for it. By run 5, tasks 5 and 6 answered "already
  answered by …" because earlier bundles had already taken up the
  learned complaint handling. The state checks confirm it was right.
- **This made "already answered" an answer** (`agent.ErrAlreadyAnswered`).
  An empty bundle with a description naming the answering rules is not a
  failure. The shell says "Nothing to add", and the eval marks ✓ and
  leaves judgment to the state check. The prompt now tells the agent to
  check the active rules first and never propose a second rule doing
  what one already does. Run 3 had done exactly that and was rightly
  refused as a conflict.
- **What the network does not know stays hard.** The PZ was solved in
  one run only, so it sits below the floor, never surfaces, and failed
  in four of five runs on reading a field through a link. That gap is
  scoped into the formula session.
- **Robustness:** `extractJSON` now decodes the first complete JSON
  object instead of cutting first-`{` to last-`}`, which a trailing
  fence with braces broke in run 5.

## 2026-10-08 — Rhea learns: the first slice of the network

- **KK: "If it learns, it uses what it learned — otherwise why would it
  learn?"** Learning is the thesis (DIRECTION, the network). Until now
  only the per-installation loop existed; this is the first cross-
  installation slice, built against several local databases standing in
  for clients.
- **Explanations leave, facts never do.** `network.ShapeOf` reduces an
  approved rule to the question it answers (its key: the event, the
  cascaded type, the kind of consequence and its target or book) and an
  anonymized answer. It drops the id, description, priority and dates,
  and replaces every condition value except structural ones
  (`$.object_type`, `$.market`, `$.currency`) with "?". It keeps the
  market vocabulary (account codes, book names, enum literals). The
  fingerprint is a SHA-256 of the canonical JSON. This is a deterministic
  first cut; DIRECTION says real anonymization is an AI task, and the
  floor guards the gap.
- **The network store** is a separate database (`rhea_network`,
  `RHEA_NETWORK_DSN`) holding append-only `publication` snapshots; an
  installation's latest snapshot is its current knowledge. The
  installation is named by `RHEA_INSTALLATION` or the database name.
- **`Learn` counts:** per question, which answer how many installations
  give, out of how many answer it at all. **Floor = 2:** a shape held by
  one installation never surfaces, because one business's pattern may
  encode its secrets.
- **Rhea uses what she learned, in three places.** (1) Priors in the
  agent's ask, following cascades from the residue's event types to the
  questions about what they create. (2) Verified support: a proposer may
  claim basis `network`; `StoreBundleDraft` verifies a rule really
  matches a surfaced answer and writes the counted support, and an
  unverifiable claim is refused (`ValidateStored`: no network warrant
  without counted support). (3) "learned: N of M installations" on each
  rule in the Bundles tab, counted live.
- **Surfaces:** `rhea publish`, `rhea network` (questions, answers,
  counts; below-floor answers marked), `rhea serve` connects when the
  network is reachable and otherwise runs alone, and `rhea eval
  -network` draws on the network and publishes the run's installation,
  a measurable flywheel.
- **Test** (`shell/network_test.go`): four installations publish (three
  post invoices 201/702, written differently, conditioned on different
  customers; one posts 201/730). No customer name or description
  reaches the network. Rhea learns 3 of 4, and the 1-of-4 answer stays
  below the floor and away from the agent. A fifth installation's agent
  sees the prior; its network-based proposal is stored with a verified
  3 of 4; a network claim for 201/730 is refused; the client's own rule
  (201/700) wins and, once published, the count reads 3 of 5.

## 2026-10-08 — the PZ as a Polish standard: the pl-warehouse pack

- **`packs/pl/warehouse.json` (`pl-warehouse` v1)**, kept separate from
  the Poland pack so installations without a warehouse are not forced to
  take it. It requires the finance and warehouse base types. A delivery
  becomes a PZ (supplier, the supplier's WZ number, warehouse, date,
  currency); each delivered item becomes a `pz_line` pointing at its PZ
  (cascade over `$.root.lines`); each line takes stock in and posts Wn
  330 Towary / Ma 300 Rozliczenie zakupu at purchase price in book
  `pl-stat`.
- **The variant without arithmetic:** the delivery states each line's
  value. Valued later (by the purchase invoice) and price-list valuation
  wait for the formula language, and the pack's description says so.
- **The warrant is practice, cited modestly:** PZ per delivery,
  valuation at purchase price ("ustawa o rachunkowości, art. 28", cited
  at the article level, not guessed to the paragraph), and the 330/300
  pair with the invoice settlement 300, 221 / 202. KK to verify it
  against practice; any client may replace it.
- **First test of the pack package** (`internal/pack/pack_test.go`):
  seeds, then the Poland pack and pl-warehouse, each approved as one
  bundle with a cited warrant and no support. One delivery gives 1 PZ, 2
  lines, 2 movements, and Wn 330 / Ma 300 of 370.40 each. Amounts are
  read as int64 minor units, never floats.

## 2026-10-08 — the warrant: every suggestion says where it comes from

- **`core.Warrant` on bundles** (a `warrant` JSONB column): basis
  (statute | standard | practice | pack | model | client | network),
  citations, scope (market, industry). Required on agent bundles, and
  packs default to their own authority. The Bundles tab shows "Why: …"
  next to Approve. This is DIRECTION's "Rhea suggests standards" made
  concrete: the agent proposes the customary standard when practice is
  open, rather than leaving it out or asking.
- **Support is counted, never claimed.** `Warrant.Support` ("N of M in
  population") is Rhea's to fill from approved rules across
  installations. `ValidateProposed` refuses any proposal carrying
  support or claiming the `network` basis, and `InsertBundle` enforces
  it for every door. A `statute` or `standard` warrant needs its
  citation. The prompt asks for no percentages and no invented
  citations. KK's "90%" was a metaphor for what learning should make
  true; the slot exists so the learning loop has somewhere honest to
  write it.
- **Why the guard exists, recorded because KK asked "if it learns, why
  wouldn't it use what it learned?":** it does. Learned knowledge is
  exactly what fills Support. The guard only stops the language model
  from fabricating the evidence in the moment. Cross-client learning (the
  network) is designed, not built. A first slice (rule shapes collected
  from several local installations, anonymized, clustered and counted,
  fed back as priors) is proposed as the next step after the PZ pack.

## 2026-10-08 — the PZ: a document with lines, and what it exposed

- **KK: "Yes, absolutely we need a PZ document."** A delivery gets one PZ
  (goods received note) with one line per item. Each line points at its
  PZ and moves that item's stock in.
- **Cascades read the chain's root fact under `$.root`.** The lines live
  in the delivery, not in the PZ header that cascades them, so the
  header's cascade fans out with `each: =$.root.lines[*]` and each line
  points back with `=$.doc.object_id`. `$.root` is read-only: it is
  never written into derived events (asserted), and replay is untouched.
  `exec/pz_test.go` covers one delivery → one PZ, two lines, two
  movements.
- **The operations corpus task 4 now asks for the PZ shape** (types `pz`
  and `pz_line` named in the question; the delivery carries the supplier
  and its WZ number). The reference passes. **Live: 0 of 3.** The agent
  looked the warehouse up by name instead of code once, and twice tried
  to read "the PZ's warehouse" through the line's link, which the
  language cannot do (it reads `$.root`, not fields of linked objects).
  Not prompt-tuned yet.
- **Two gaps that KK's question "how did we teach the agent which
  accounts to book the PZ?" makes concrete:**
  1. *Booking policy is never taught.* The agent knows accounts only from
     the question's own words and the chart of accounts (codes, names,
     kinds), never policy such as "PZ: Wn 330 / Ma 300". The interview
     must ask for it, and a pack can offer the market default as a draft.
  2. *There is no arithmetic.* A PZ entry is quantity × purchase price,
     and templates copy and sum but cannot multiply. This is the first
     business case that forces the arithmetic sub-language (DIRECTION,
     2026-10-03), together with reading fields through links
     (`item.std_cost`).

## 2026-10-08 — carried links: a follow-up points at exactly what caused it

- **KK's decision (option A, posed in business terms):** a follow-up may
  point at exactly the thing that caused it. This is ref re-reference,
  the exit named since E2/E4 and the model's first reach four times
  over. A `ref<T>` field now takes either `=ref(...)` (a lookup by
  value) or `=$.path` carrying an object id: in a cascade, `$.object_id`
  (the causing object) or `$.state.<ref field>` (a link it holds).
  Literal ids stay refused at validation, so a rule never hard-codes a
  link.
- **Vouched at expansion** (`vouchRef`), the courtesy the amendment
  already pays its target. A carried id must be of the declared kind
  (identity is typed by construction, `<type>-…`) and must exist, in the
  world or earlier in the same chain. This applies to materializations
  and to amendment sets.
- **The reading side, part of the same decision:** a lookup may take one
  step through a link, `=ref(case, complaint, ref(complaint, number,
  $.number))`, meaning "the case of complaint R-1". Without it, exact
  links make follow-ups unfindable from events that name only a number.
  The live eval showed exactly that: the case linked properly and the
  withdrawal could not find it.
- **Prompt:** the ref limit is replaced by the capability (link to the
  cause, find through the link). "An event is explained once" became
  "rules approved later reach explained events only through a
  human-approved backfill". "Create only the objects the question asks
  for" was added after my own receipt example nudged the model into extra
  goods receipts.
- **Live, operations corpus, three runs:** DONE once. The withdrawal now
  works in all three, directly or by backfill through the link. The two
  NOT DONE runs both turn delivery lines into goods receipts (3 where the
  corpus expects 1). That is a business question for KK (does a delivery
  get a goods-receipt document, the Polish PZ?), not a language gap. The
  corpus is left as is until KK answers.
- `core_test` now asserts the new law: a carried id is admitted and a
  literal is refused. `exec/link_test.go` covers the receipt → movement
  cascade copying links, case → complaint, the withdrawal closing the
  case through its link, a wrong-kind id, a missing id, and replay.

## 2026-10-08 — ruled backfill: the past explained further, never rewritten

- **`PlanBackfill` and the `approve_backfill` door** (`backfill.approved`,
  a reserved namespace). Every explained raw event is re-expanded under
  the active rules. Its existing derivations come from `ChainOf` (a
  recursive walk over `cause_event_id`, `queries/chain.sql`) and are
  replayed as booked nodes with their baked state, never re-expanded,
  because re-judging an old amendment against today's status would
  refuse a transition that already happened. Only rules that never fired
  at a given chain point expand. A node carries `booked` (its existing
  event id), and `book()` skips it and uses it as a cause. Additive by
  construction (KK, question 1).
- **A deliberate act after approval** (KK, question 2). The plan is
  computed before the approval event is appended, the event names every
  chain (event, rules, forwarded or not, refused), and after commit the
  chains book, each atomically, before the ordinary worklist pass. A
  second backfill finds nothing and says so.
- **The log never closes; periods do** (KK, question 5, revised). A
  chain whose expansion meets a locked (book, month) (now a typed
  `PeriodLockedError`) is re-expanded dated on the backfill: offered
  forward and caused by its original fact. **Forwarding is per chain in
  v1**: an open book's entry travels forward with its closed sibling,
  because a chain books atomically. Forwarding per pair is the named
  refinement.
- **Time events stay out of backfill**: a late rule on time belongs to
  the clock's catch-up, behind its own gate.
- **Agent bundles without a sample apply from the log's first fact**
  (`EarliestFactDate`), not from today. A rule answering an interview
  explains the history, and backfill only reaches events on or after a
  rule's `effective_from`.
- **Surfaces:** `rhea backfill [-dry]`, `POST /api/backfill/simulate`
  and `/approve`, and a "The past" panel on the Bundles tab (the dry
  run, forwarded chains explained, Approve backfill). The eval plays the
  human's second act: after a clean bundle approval, a clean backfill
  plan is approved.
- **Found on the way:** lists ordered by `object_id` as text, so
  `invoice-10` sorted before `invoice-9`. One more seeded builtin
  shifted ids and the M0 demo test caught it. Lists now follow the log
  (`source_event_id`, then id).
- **Prompt:** the ref limit now names `$.object_id` too. The model
  reached for ref re-reference a fourth time (case → complaint) and
  failed task 5 on it.
- **Live (claude-sonnet-4-6), operations corpus, three runs: DONE, DONE,
  NOT DONE.** Task 6 now reads "booked 0, backfilled 1": approval
  changes nothing and the backfill resolves the case. The third run
  over-materialized (a delivery also became a goods receipt), which is
  modeling variance, not a language gap.
- **Tests:** `late_test.go` now ends in its exit (three KYC checks by
  backfill, caused and dated by their openings, replay identical).
  `backfill_test.go` covers the closed period offered forward (8
  postings dated on the backfill, each walking back to its September
  invoice) and the eval's forcing case (a late amendment reaches the
  case after an earlier amendment is replayed, not re-judged).

## 2026-10-08 — enrichment: amendment consent is per field

- **`ObjectType.amendable`** lists the fields rules may set after
  materialization. The lifecycle field stays implicitly amendable under
  its transition law, and every other field is fixed at birth. A type
  consents to amendment by declaring a lifecycle, amendable fields, or
  both. This refines the amendment's shape (the third primitive); it is
  vocabulary on ObjectType, not a closed-class admission. KK settled
  question 3 on it.
- **Existing types tightened, as decided.** Case types declare
  `resolution`, schedule types declare `next_run` (the schedule advances
  itself). Every test that broke did so for exactly that reason.
- **Kernel bug found on the way:** `expandAmend` assumed a lifecycle was
  always present (`lc := objType.Lifecycle // non-nil`) and panicked on
  an enrichment-only type. The transition law now applies only when a
  lifecycle exists.
- **The attempt test now tells the after story.** The contortion (a
  lifecycle invented to buy consent) no longer opens the account number
  ("fixed at birth"). `bank_account` v3 declares `amendable: [iban,
  bank_country]`, the waiting IBAN books on the next pass as an
  `object.amended` with provenance, and the account keeps
  `type_version` 1.
- **The agent learned it.** Both prompts explain per-field consent, and
  the bundle prompt explains extension: redeclaring a type by name makes
  it the next version.

## 2026-10-08 — late-understanding attempt: evolution carries, enrichment and backfill fail

- **The attempt is kept as executable evidence** in
  `internal/exec/late_test.go`, a bank-account story run against the
  adopted note (DIRECTION, "Late understanding").
- **Evolution carries as pure data.** `bank_account` v2 adds optional
  `iban` and `bank_country`, and the three existing accounts keep
  `type_version` 1. Nothing earned.
- **Enrichment fails as predicted.** The IBAN arrives as a fact
  (`bank_account.details_provided`), and the amending rule is refused at
  expansion ("declares no lifecycle"). The event waits in the worklist.
  The contortion is recorded too: a lifecycle invented only to buy
  consent (`state: open`, no transitions) carries the waiting IBAN, and
  it also validates a rule that rewrites the account *number*, the very
  identity the IBAN was keyed by. Consent spelled as a status machine is
  all or nothing. The exit is `amendable`, per field.
- **Backfill fails as predicted.** A KYC check rule approved after the
  accounts were opened: its dry run sees three checks, and its approval
  delivers none. Explained events are never revisited, so understanding
  that arrives after the facts has no way in. The exit is the backfill
  door.

## 2026-10-08 — the agent proposes bundles; a corpus that discriminates

- **`DraftBundle`**: the agent answers one interview question with a
  bundle: new object types (with lifecycles), rules, optional list and
  detail views, and optional verbs. The kernel assigns versions. The
  validation gate checks types with the type gate, rules against the
  catalog overlaid with the bundle's own types, views against their
  type's fields (list and detail only: analysis SQL and scheduling are
  not agent-authored), and verbs with the activity gate (never a system
  namespace). `shell.StoreBundleDraft` lands the drafts (a taken bundle
  id gets a suffix) and is shared by the shell and the eval. Stores run
  sequentially, not in one transaction: the gate already ran, and a half
  landing leaves honest orphan drafts.
- **The agent sees the residue**: `Ask.Residue` clusters the worklist by
  event type (count plus one sample), the groundwork for the
  residue-first protocol. A bundle needs no sample event.
- **`draft_bundle` door** (`bundle.draft_requested`; the sample event is
  optional), `POST /api/bundles/draft`, a "Draft bundle" button on the
  worklist form, and a free-standing ask box on the Bundles tab.
- **The prompt now states the language's limits**: no ref re-reference in
  cascades (use a second rule on the raw event), and an event is
  explained once.
- **The eval learned bundle tasks** (`"mode": "bundle"`) and field-level
  expectations (`expect_where`), plus `rhea eval -v` to print every
  draft. New corpus `testdata/eval/operations.json`: six bundle tasks
  covering master data with FX, invoices into two books with `convert`,
  goods with stock movements, `each` deliveries, new lifecycle types,
  and amend on withdrawal. The references prove it done.
- **Live score (claude-sonnet-4-6), operations corpus.** First run: 4/6,
  residue 2 of 23 (goods refused on the ref re-reference gap). With the
  limits in the prompt: 6/6 approved, residue 0 of 23, state NOT DONE:
  case resolved 0 of 1.
- **Finding 1: residue zero is not done.** Task 5's bundle explained the
  withdrawal in its own reasonable way (amending the complaint to
  `withdrawn`). Task 6's correct rule (resolve the case) was approved
  over an event already explained, so it never fires, and the dry run
  says so honestly (explains 0, adds 0). In an interview, understanding
  arrives in an order the events did not. This is the forcing case for
  **ruled backfill** (DIRECTION: a simulation diff promoted into the log
  by a gated activity). Until then, the definition of done is residue
  zero *and* the human's acceptance of state, which is what
  `expect_where` stands in for.
- **Finding 2: the ref re-reference gap is what the model reaches for
  first.** Cascading a stock movement off a materialized receipt is the
  natural modeling, and the model drafted it in both unprompted runs.
  This is the third time the exit has come up (E2, E4, here).

## 2026-10-08 — the bundle: one approval, many definitions

- **Object types and views joined the gate** (KK: everything drafts).
  Both tables gained `status` (draft | active | superseded; rows from
  before default to active). Their version is the schema's identity
  (objects record `type_version`), so approval appends the `active` row
  of the *same* version instead of minting one. The primary key became
  (name, version, status). Every reader (`GetObjectType`,
  `ListObjectTypes`, `LatestViewDefs`) is active-only; a draft type is not
  language yet. `InsertObjectType`/`InsertViewDef` stay as the
  active-insert bootstrap and test path, the way tests insert active rules.
- **`bundle` is a versioned, append-only definition**: an id, a
  description, and members (kind, name, version) pointing at draft rows
  of the four vocabularies. It adds the approval's scope and nothing else.
  Two builtin doors: `approve_bundle` emits `bundle.approved`, and its
  kernel reaction appends every member's active row plus the bundle's in
  one transaction, then runs one worklist pass. `reject_bundle` requires
  a reason and emits `bundle.rejected`. `bundle.*` joined the reserved
  system-verb namespaces.
- **No partial approval, enforced at the doors.** `approve_rule` and
  `approve_activity` refuse any draft that belongs to a bundle. A draft
  belongs to one bundle for life: it can't join a second, and after a
  rejection it stays a rejected draft. A redraft is a new version in a
  new bundle.
- **Rejection supersedes only the bundle, never its members.** Appending
  a superseded version of a draft rule would retire the active version it
  was redrafting (a rule whose latest version is superseded is retired).
  Membership already locks the drafts, so touching them would gain
  nothing.
- **The bundle's dry run overlays its draft types.** `SimulateBundle`
  runs the active rules plus the bundle's drafts on a private copy of the
  executor whose type lookup sees the bundle's draft types (the booking
  executor never does). The diff renderer gets the same overlay.
- **Packs land as one bundle** (`pack-<name>-v<version>`), and approving
  it is the installation. `rhea load` drafts into
  `load-<file>-<content hash>`. New CLI verbs `rhea approve BUNDLE` and
  `rhea reject BUNDLE REASON`, plus a Bundles tab under rules in the
  shell (members inline, simulate, approve all, reject with reason). Pack
  tests now approve bundles. Intercompany activates
  `pl-record-ksef-submission`, which it used to skip; no behavior change.
- **The forcing case passes** (`exec/bundle_test.go`): one
  `goods.received` gets a receipt (whose type arrives in the same bundle)
  and a stock movement from one approval, both provenanced to the same
  event. Replay reproduces the state.

## 2026-10-07 — the eval harness: residue as the agent's score

- **`rhea eval [CORPUS]` measures the agent as an author.** A corpus
  (`testdata/eval/corpus.json`) is an implementation script as data:
  seeds, intake events, intents with sample types, and the expected
  object counts. Per task: ask → validate → store draft → dry run →
  approve if the dry run shows no errors (the human's role played by a
  policy, actor `eval`). Verdict: residue over the intake plus a state
  check. It runs on a throwaway database (`store.Throwaway`, now also
  behind `storetest`) because the system of record is append-only. It
  lives in a new package `internal/eval` because it orchestrates agent,
  shell context and executor, which no existing package should import
  together.
- **Every corpus carries reference answers, and a test proves them
  done.** The recorded agent replays the references through the same
  path. A corpus its own references cannot finish measures nothing.
- **First live score (claude-sonnet-4-6): 9/9, residue 0 of 18,**
  including a cascade posting it had not been taught before today. The
  corpus is the M0 story and too easy to discriminate. A harder corpus
  (convert, book, amend over lifecycles, `each`) is the next
  measurement.
- **Finding: one event, two consequences cannot be implemented one
  approval at a time.** `goods.received` needs a receipt *and* a stock
  movement. Approving one rule explains the event, so the second rule
  only fires on future events (backfill is ruled), and a cascade cannot
  carry the ref (`ref<item>` will not take `$.state.item`, an id; the
  ref re-reference exit again). The corpus leaves stock movements out.
  This is evidence for the bundle as the approval scope (DIRECTION,
  implementation by interview): approving the two rules together fires
  both on the pending event.

## 2026-10-07 — the agent drafts against the whole language

- **The agent sees the language as it stands, not one pre-chosen type.**
  `DraftRule` takes an `Ask`: the full object-type catalog, the active
  rules, master data for every ref-target type (capped at 30 per type),
  the sample event and the intent. Choosing the effect and its target
  type is part of the authoring. The caller (`shell.DraftAsk`) gathers
  the context; the agent still has no store access (invariant 3).
- **The prompt teaches the whole rule grammar**: cascade on
  `object.materialized` (`$.object_type` guard, `$.state.*`,
  `$.object_id`), amend (lifecycle consent, never a literal target),
  `book` and `convert`, alongside object/`each`/postings. Validation
  adds one catalog gate: an object or amend effect must target a
  declared type. Why: phase (a)'s order of work, step 1. The agent was
  frozen at M0's grammar while the closed class grew three primitives.
- **The `draft_rule` door is unchanged; its `object_type` input is now
  a hint.** Making it optional would need a v2 builtin, and builtins
  seed v1 only, with no evolution path. The human still points at a
  type, and the agent may answer with a posting or an amendment
  instead. Exit named: builtin versioning, which is needed when the
  interview's own doors (`interview.said`) arrive.

## 2026-10-05 — M5: time is in the log, and the commitment stayed data

The founding ladder's last milestone. The closed class earned nothing —
the strongest outcome: time, schedules and the gated switch all landed as
an adapter, conventional vocabulary, and rules.

- **The original expectation, kept for the record**: the rules-only
  attempt (previous entry) would fail on due-set fan-out and cadence
  arithmetic — it did, plus one surprise (a quiet day is a rule error:
  the encoding could not even say "usually, nothing happens"); REA's
  commitment was expected to become unavoidable — it arrived and stayed
  pure data; date comparisons or ref-reads in match were expected to be
  forced — they were not, because the adapter carries the scan.
- **Time passes as facts**: `time.day_opened` / `time.month_opened`
  (months open with their first opened day), appended by the clock —
  an adapter with its own runner, because a day's consequences decide
  the next day's due set, so opening interleaves with processing. Rules
  never read the wall clock; `Clock.Now` is the system's one wall-clock
  read, injectable, so even the tests own time.
- **Calendars are the clock's knowledge** (the adapter contract carrying
  what templates must not): the due scan over `task_schedule` objects —
  conventional vocabulary, like posting and account; the kernel does not
  know it — and the cadence arithmetic, monthly/yearly clamped to month
  ends (a schedule on the 31st runs Nov 30). `schedule.fired` carries
  the occurrence and the baked advanced `next_run`; ordinary rules raise
  the work (cases, composing with the cases test) and amend the schedule
  — determinism by construction, nothing re-computed at replay.
- **The commitment verdict**: a schedule — an event that should happen —
  is an object with a lifecycle plus the clock plus two rules. Negative
  proof; no primitive. Plans at production depth (capacity, MRP) are
  pack-and-rules depth on the same mechanism, to be falsified when a
  production domain arrives.
- **The switch, as constrained** (DIRECTION): per behavior — supersede
  the rule; per schedule — pause/resume by amendment (active ⇄ paused,
  declared transitions); globally — stop the clock, and the log shows
  the gap honestly. **Re-entry is gated**: RunClock opens at most one
  pending day (steady state); a burst gates, SimulateCatchUp dry-runs
  the opening through the same diff renderer as rule approval (the
  dry run stands in for the advance rule itself, labeled), and only an
  explicit CatchUp opens the days — in order, each day processed before
  the next day's scan. Missed occurrences fire late, one cadence step
  per opened day, never silently skipped — the gate is what warned the
  human about the burst.
- **Quiet days are not questions**: the worklist hides `time.*`; the
  executor's set (new `pending.sql`) keeps it, so time fires rules and
  failed firings retry as data. Quiet days re-scan forever — accepted,
  performance is an explicit non-goal. Simulation exempts time from the
  unexplained count the same way. An unexplained `schedule.fired` is
  real residue and stays in the worklist.
- **`time.*` is not reserved**: a forged day through the open door is
  possible and visible (actor column, dedup keys are the clock's); the
  namespace gets reserved to adapter doors if that honesty ever proves
  insufficient.
- **The scheduling notion shipped** — the founding table's last entry:
  spec is `{object_type, date_field}`, label and status derive from the
  ObjectType like derived views; the shell renders days in order, past
  and today marked; `view_def`'s CHECK re-learned the notion list by
  idempotent migration. The Time surface (Clock tab) is a native system
  function beside worklist/rules/language.
- Named exits that stay named: condition-over-state-×-time ("invoice
  unpaid N days") still waits on ref-reads or date compares in match —
  M5 did not force them; "until fully depreciated" waits on the
  arithmetic sub-language; list filters still belong to analysis views.

## 2026-10-05 — cases are pure data: the falsifiability test passed

Alpha's case model expressed as object types, rules and activities — the
whole story in `internal/shell/cases_test.go`, over HTTP, with zero
case-specific screens and no kernel change beyond the amendment admitted
below.

- **The expression**: `case` is a conventional type (kind, subject as the
  recorded union-ref contortion, title, status with a declared lifecycle,
  due_date); a cascade rule raises it from the subject's materialization —
  raised by the system, never typed in, by construction; `resolve_case` is
  an activity with a `ref<case>` input, offered inline on the generic
  detail by the surfacing convention; a resolution rule amends the case
  through `=$.case` (the id the door vouched); and "closes itself when the
  world moves on" is one more amend rule matching the world's event,
  resolving the case by `=ref(case, subject_id, …)` — human resolution and
  world-moved-on resolution are the same mechanism.
- **"An empty list means the day's work is done"** is an analysis view
  (`WHERE status = 'open'`) — a property of the event log, not a screen.
  ListSpec still has no filters; the read side carries the question until
  a list filter earns its place.
- **The detail and the walk explain the whole life** (invariant 5 both
  halves): the materialization says why the case exists, each amendment —
  listed on the detail with its rule version and delta, shown in the walk
  as "amended <case>: status → resolved" — says why it is what it is now.
- **What the shell needed: nothing case-shaped.** Derived views render the
  case, contextual verbs offer the resolution, the omnibox runs it. The
  generic machinery shipped for activities was sufficient — which is
  itself the thesis holding.
- Deferred with named exits: due dates computed from terms wait on the
  arithmetic sub-language; time-raised cases ("unpaid N days") wait on M5
  time events; list-inline verbs are renderer polish, not language.

## 2026-10-05 — the amendment earns admission: status is a projection

The third closed-class change (after `book` and `convert`), admitted by all
four governance rules: the cases attempt below is the fail-as-data proof;
the invariant is consent + declared transitions; the determinism proof and
the network test land with the shape.

- **The shape**: `effect.amend` = `{type, target, set}` — tiny and
  declarative (the network test). Target is `=$.path` (an id the door or a
  prior resolution vouched) or `=ref(...)`, never a literal. Set is field
  templates, evaluated with the same typed machinery as materialization
  fields, values baked into the derived event at firing time.
- **The event**: `object.amended`, kind derived, payload
  `{object_id, object_type, set}` — the delta IS the explanation, with
  `(cause_event_id, rule_id, rule_version)` like every derived fact. Raw
  `object.amended` is refused at append alongside `object.materialized`;
  both join `ReservedEventType`.
- **The invariant (no invariant, no primitive)**: amendment is
  *consent-based* — only a type that declares `lifecycle: {field,
  transitions}` may be amended, and a Set touching the lifecycle field must
  move along a declared transition, judged at expansion where the current
  status is known — the way balance is judged. Initial status at
  materialization stays unconstrained beyond the enum.
- **Determinism**: replay re-applies baked deltas in log order (`state ||
  delta` in the cache; the DuckDB rebuild merges in memory); a test wipes
  and replays to identical state. Simulation applies amendments in its
  in-memory world, so the approval gate shows the move field-level
  (before → after) before it is real.
- **Amendments end their branch of the chain**: nothing fires on
  `object.amended`, and a rule matching it is refused at validation rather
  than dying silently. A cascade-on-amendment joins the language when a
  domain fails without it, not before.
- **Ambiguity stays human**: two rules amending one object on one event is
  a conflict — the event refuses whole, same as same-id materialization.
  Within a chain, later generations read amended state (the get overlay
  merges earlier amendments); lookups still see materialized values only.
- **Fail-as-data all the way down**: an amendment whose target does not
  exist waits in the worklist and applies on a later pass — the fixpoint
  settles out-of-order arrivals (resolution before complaint) in one
  ProcessPending.
- **Scoped out, named exits**: amendments do not check period locks (the
  stance lands when a financial domain amends — postings are untouched by
  this effect); as-of replay with a cutoff is machinery already paid for,
  built when a question asks it; the agent is not yet taught to draft
  amend effects (phase (a) vocabulary work).

## 2026-10-05 — cases attempt: raising carries, closing fails as pure data

The next falsifiability test (DIRECTION: express Alpha's case model as pure
data) — the attempt, run with the language as it stood, kept as executable
evidence in `internal/exec/case_test.go`.

- **The original expectation, kept for the record**: raising would carry
  (cascade from any materialization), the subject would strain into the
  union-ref contortion, resolving verbs were already shipped (contextual
  activities), and the attempt would fail at "closes itself when the world
  moves on" — forcing the amendment that DIRECTION 2026-10-03 already
  recorded as a destination ("status is a projection, never an update").
- **What carried**: a complaint event books the document, a cascade rule
  raises the case from the document's own materialization — raised by the
  system from events, never typed in, by construction (objects only exist
  through rules); provenance walks the work item back to the root fact.
  The resolving verb is a contextual activity with a `ref<case>` input,
  shipped with activities-as-data.
- **The recorded contortions**: subject as `(subject_type, subject_id)`
  strings — `ref<T>` names one target and a case's subject is any object;
  exit at union refs (SPEC §7, parked). Due dates cannot be computed
  ("+14 days" is inexpressible); exit at the arithmetic sub-language.
  "Open cases only" is a list-level filter ListSpec does not have; the
  analysis notion carries it in SQL until a list filter earns its place.
- **Where it failed — two facts, sharper than predicted**:
  1. *Identity physics forbids even touching the case.* Ids are
     cause-qualified, so a closing rule mints `case-<new event>` — a
     duplicate, both open. No rule form can name an existing object; the
     conflict guard never even fires.
  2. *The closure marker is the hollow-explanation contortion* (E3's
     shape): a `case_closure` object records the truth while the case's
     own status field lies ("open", forever), and reality retreats into an
     anti-join on the read side. A state field that can never tell the
     truth is not explained state — invariant 5 refuses the shape.
- **The exit**: the amendment — a rule effect that moves an existing
  object's declared lifecycle along declared transitions, as a derived
  event with provenance, replayed like everything else. Earns admission by
  the four rules: this entry is the fail-as-data proof; the invariant is
  kernel-validated transitions declared on the ObjectType; the determinism
  proof and the network test (a tiny declarative `amend` clause) come with
  the primitive.

## 2026-10-05 — activities as data: the doors are declared

- **The shape**: `Activity` = name, lifecycle, domain, description, spec
  {inputs: FieldDefs, emits, who}. Versioned append-only rows (`activity`
  joins the invariant-1 trigger family); the rule lifecycle, where `active`
  means *offered and triggerable*, never *executes in replay* — activities
  are not in the determinism path, superseding one touches no history.
- **The schema belongs to the verb, not the event**: inputs validate at
  trigger time, before the event exists; the emitted event is schemaless in
  the log and a test asserts an event through the door is payload-identical
  to the same fact appended free-form. Money inputs validate via ParseMoney
  but cross as the canonical decimal string — the payload is boundary data.
- **The door stamp**: raw events gain `(activity_name, activity_version)` —
  the provenance symmetry to `(rule_id, rule_version)` on derived events.
  Adapter and pack events stay honestly unstamped: their door is the actor
  and the dedup key.
- **Retrofit findings** (predictions in DIRECTION 2026-10-05, outcomes here):
  - `approve_rule` carried as predicted: the old handler already appended
    `rule.approved` first, so the retrofit only moved the version flip into
    a kernel *reaction* keyed on the event type — and into the trigger's
    transaction, making approval event + version row atomic, which they
    never were before. `rule_id` stays a string input (FieldDef refs point
    at ObjectTypes); exit named at ref-to-definition.
  - `submit_event` carried via the passthrough form: one `json` input whose
    value IS the payload, everything else envelope-consumed ("a passthrough
    activity has nothing else to say about the payload"). The open door is
    now versioned, auditable data — and it refuses kernel namespaces.
  - `draft_rule`: the *declaration* carried; the fulfillment could not be a
    reaction, because reactions are deterministic kernel code and the agent
    is not. The ask is now recorded (`rule.draft_requested` — drafting
    finally leaves a trace in the log); the shell consumes the request
    synchronously as the agent's runner; true agent-as-user waits for phase
    (a). The boundary drew itself sharper: the kernel reacts to events, it
    never converses.
  - `approve_activity`: the gate applied to the gate, no second governance
    mechanism. Bootstrap: init seeds the builtins v1 active, each activation
    an `activity.approved` event stamped `approve_activity` v1 — the
    self-referential fixed point, recorded honestly (actor `kernel`,
    approved_by `init`).
- **Forgery guards at both layers**: declared activities cannot emit into
  `rule.*`/`activity.*` (the store refuses); a raw event in those namespaces
  without a door stamp is refused at append; the passthrough refuses them as
  resolved types. The system-verb namespaces are spoken only by their doors.
- **System verbs exclude symmetrically** (`core.ReservedEventType`):
  `activity.*` joins `rule.*` in the worklist query and the simulator's
  unexplained count.
- **Ref inputs are vouched**: the named object must exist and be the
  declared type — the shape cases will lean on (a case's resolving verbs
  take a ref to the case). `list<ref>` waits for reconciliation's partial
  matches to demand it.
- **Packs ship activities as drafts** — status is not a pack's to set;
  approving them is the installation, same as rules.
- Found on the way: list rows order by `object_id` as text, so two-digit
  event ids reordered a test's expectation; the test now finds its row. Id
  ordering stays a non-goal.

## 2026-10-05 — the market corpus: evidence becomes vocabulary

- **Market knowledge arriving as evidence (screens of production systems,
  statutory forms) accumulates in `packs/<market>/CORPUS.md`**, one dated
  evidence session at a time; pack object types, rule intents and notion
  specimens are distilled from it. Screens are evidence, never design: the
  corpus takes vocabulary, field semantics and status ladders, not layouts.
  Why: the agent needs drafting context and the pack needs a provenance trail
  for where its vocabulary came from; a corpus file is both.
- **Vocabulary lands in the evidencing market's pack, not the base seed.**
  `contractor` ships in packs/pl beside the base's `company`; merging them is
  deferred until a rule needs both unified. Why: the M4 lesson applied to
  vocabulary — promotion to base is earned by a second market needing it, not
  anticipated.
- **A settlement references its document as `document_type` + `document_id`
  strings.** The language has `ref<type>` but a settlement may point at a
  sales_invoice, purchase_invoice or tax_declaration; whether `ref<a|b>`
  (union refs) earns admission is parked as SPEC §7 material.
- **The language explains itself natively: `GET /api/types` + a `language`
  system function beside worklist and rules.** Each type carries its
  explanation — producing rules (with consumed event types), views (derived
  ones marked), ref edges both directions, instance count. Not a ViewDef:
  definitions are not objects, and views render objects. Why: a system that
  learns by explanation must be able to explain its own vocabulary; "what
  does Rhea recognize?" was answerable only by grepping seed files.

## 2026-10-05 — the simulation diff is field-level and typed

- **What the human reads at the approval gate is now fields, not JSON.** The
  kernel's SimDiff crosses the boundary through the same cell encoding as
  every view (money canonical, refs labeled, enums chipped): added and
  removed objects as field lists, changed objects with before → after per
  field and the changed ones flagged. The diff also names **the events the
  rule would newly explain** — the other half of the approval question.
- **Refs resolve against the hypothetical world first**: an added object may
  reference another object that only exists if the rule is approved, so the
  resolver consults the simulated set before the cache (`cellWith`, the one
  encoding path, parameterized by resolver).
- **No doors into a world that does not exist**: diff cells carry no detail
  links, and added objects none either; changed and removed objects link to
  themselves "as they are today", which is exactly what a reviewer wants to
  compare against.

## 2026-10-05 — live screens: the single writer is the single announcer

- **One notice per committed booking** (Alpha's lesson, DIRECTION): the
  executor sends `pg_notify` inside the booking transaction — Postgres
  delivers it only on commit, so a rolled-back write never announces itself
  (a test proves the boundary). The payload names the touched object types;
  a notice is a signal, never data — listeners re-read through the ordinary
  API. Replay announces once, as "replay".
- **`GET /api/live` is one SSE stream per browser tab**, fed by a LISTEN
  connection from the pool, with a ping comment every 25s. The client
  re-renders the active tab when a notice lands (bursts debounced to one
  re-render).
- **Live is a property of the tab, not the shell**: view tabs (list, detail,
  analysis), the Language catalog and the provenance walk re-render in
  place; tabs holding human state — worklist drafts, simulation results —
  stay manual, because a re-render must never eat what a human is typing.
  Row-level patching is deliberately skipped: re-rendering a tab is enough
  at experiment scale, and the signal protocol would not change.

## 2026-10-05 — the provenance walk: invariant 5 as an interaction

- **`GET /api/explain?object=<id>` (or `?event=<id>`)** answers "why does this
  exist?" with the whole story: the chain's root raw event and its full
  consequence tree — every derived event, each hop naming its **exact** rule
  version (descriptions come from all rule versions, superseded included:
  provenance explains history). The raw root alone carries its payload — the
  fact everything else explains. Each materialized object in the tree carries
  its label and detail view id, so the walk is made of doors, not dead ends.
  Like worklist and rules, this is a native shell surface: it renders the
  invariant layer, not a ViewDef.
- **Ref cells became doors too**: a ref cell now carries `detail` — the view
  that opens the referenced object — which derived defaults guarantee exists
  for every type. List → ref → detail → walk is a closed loop.
- **Analysis columns may declare type `event`**: the cell renders as a link
  into the walk. intercompany-positions declares its root_event column so —
  the reconciliation screen with nothing to reconcile now opens the proof on
  click, which is the E4/E5 demo as an interaction instead of a test log.
- Store grew `GetEvent`, `RuleDescriptions` (keyed by exact version) — both
  one-query reads; the walk indexes the derived log per request, the same
  scale stance as the analysis rebuild.

## 2026-10-05 — the typed view API and derived default views

- **The view API carries semantics; renderers format.** Columns declare
  `{field, label, type}`; cells cross as `{v, id?}` with `v` in the canonical
  boundary encoding — money as the "1234.56" decimal string (invariant 6),
  never locale-formatted — and a ref cell carries the referenced object id
  next to its resolved label. Detail provenance crosses as data `(event_id,
  rule_id, rule_version)`; the renderer phrases it. The web shell formats
  money per browser locale (BigInt string surgery, no floats): Polish and
  English eyes see the same response differently, which is the replaceable-
  interpreter contract doing its job.
- **Money renders bare; currency stays a sibling column.** A money field does
  not name its currency field; binding them on FieldDef is a later language
  change if a view ever needs the pairing.
- **Default views are derived from the ObjectType** (version 0, ids
  `derived:list:<type>` / `derived:detail:<type>`), grouped in the submenu as
  "documents" or "master data" by `is_document`. A stored list or detail for
  the type switches the derived one off: ViewDefs are the exceptions
  (localization, role, emphasis), never boilerplate. period_lock and
  stock_movement's detail got screens today without anyone authoring one.
- Analysis columns gain an optional declared `type`; a declared money column
  arrives from DuckDB as BIGINT minor units and crosses as the decimal
  string. Undeclared columns stay bare strings for the renderer.
- **Analysis SQL emits data, not formatting**: every `printf` money wrapper
  left the analysis specs; the assertions did not change by one character,
  because the same canonical strings now cross the boundary from the typed
  contract instead of from SQL — which is the proof the contract holds. Seed
  fixtures were edited in place with versions untouched (the append-only law
  binds the database, not testdata; bumping versions across staged seed files
  would scramble the stages); the changed pack views were bumped (rejestr v3,
  ust v3; pl pack v5, de pack v3) because packs are the distribution vehicle
  and version discipline is the point there.

## 2026-10-05 — design tokens: the aesthetic contract is a file

- **`web/tokens.css` is the shell's whole design language**: two layers of CSS
  custom properties — primitive palette and scales, then semantic aliases —
  and `app.css` may reference only the semantic layer; no color or size
  literal outside the token file. Vanilla custom properties, no framework, no
  build step (SPEC stack holds; it is also the format designers' token tools
  emit). The light color scheme ships as a pure semantic remap — the standing
  proof that restyling Rhea means editing one file, never selectors or JS.
  Why: humans buy with eyes (DIRECTION 2026-10-05); the generic shell makes
  the styled surface a closed set of primitives, so the foundation for an
  external designer costs one file now and nothing later.

- **No attempt ceremony this time**: E5 asked nothing of the closed class.
  The one code change is read-side — the DuckDB projection now carries
  `root_event_id` (one ordered pass; causes precede effects in the log) —
  which is invariant 5 finally arriving whole on the analysis side.
  Everything else is two analysis ViewDefs (finance_v9): the group lives
  entirely in open-class data, exactly as DIRECTION prescribed ("starts
  life on the analysis side").
- **intercompany-positions**: both sides of each position, matched by
  their shared root event at original transaction amounts, difference
  zero by construction — the reconciliation screen with nothing to
  reconcile, only to display.
- **group-trial-balance (EUR)**: PLN translates at the latest EUR/PLN
  rate; postings whose root is an intercompany root are eliminated unless
  their account is tax-typed — intra-group positions vanish, each
  entity's claim against its tax office is third-party and survives. The
  residue of booking at the transaction rate but translating at closing
  surfaces as a visible CTA row — the rounding-plug aesthetic at group
  level: differences are displayed, never leaked.
- **Simplifications recorded, exits named**: closing rate = latest known
  rate (statutory translation methods — closing/average/historical per
  line item — are pack depth); one currency pair assumed (general
  multi-currency translation is analysis-side work when a third currency
  arrives); elimination-by-exclusion balances only because the CTA row
  absorbs the residue — elimination *entries* in a group book are the
  write-side form, waiting on book composition (E2 residue); equity
  consolidation and minority interests stay outside the experiment.
- **The enterprise-structure ladder is complete.** Five rungs in one day:
  E1 earned `book`, E3 earned `convert`; E2, E4 and E5 passed as pure
  data. Two primitives, three negative proofs — the grammar stayed small
  while the gap named in "What big means" closed end to end.

## 2026-10-04 — E4 attempt: intercompany fails on one gap — a ref cannot be re-referenced

- **The attempt** (`TestIntercompany`): entity A's sale raising entity B's
  purchase by cascade — the mirror rule matches the sales invoice's
  materialization and books the buyer's side. Everything fits the existing
  vocabulary except one seam, with two jaws: the invoice's state carries
  seller/buyer as resolved object ids, and a direct-id template into the
  purchase's supplier ref is rejected by design (referential integrity only
  through resolution), while resolving by value hands `=ref(company,
  vat_id, $.state.seller)` an object id. Atomicity makes the gap total —
  the broken mirror refuses the whole root event, so A's own document does
  not book either: never half-explained across entities, but intercompany
  stays inexpressible.
- **A second seam**: the invoice object cannot say it is intragroup — the
  attempt's mirror is keyed to a literal invoice number, an absurdity that
  sharpens the verdict.
- **Verdict: no primitive is earned** — the governance bar is
  *inexpressible*, and this is expressible with denormalization, so E4 is
  another negative proof. The fix as data: `sales_invoice` v3 carries
  `intragroup` and the resolution keys (`seller_nip`, `buyer_nip`) as
  plain fields, so downstream rules re-resolve by value. The contortion
  joins the recorded family (payload `market`, composite codes) with the
  same named exit: ref-reads — reading through a ref, or passing a
  chain-provenance-sound id — on the arithmetic destiny list. When that
  arrives, the nip echo fields die with the market flag.
- **Shipped** (`TestIntercompany`): sales_invoice v3 (finance_v8) echoes
  `intragroup`/`seller_nip`/`buyer_nip`; pl pack v4 fills them; the mirror
  rules are group-level data (what a phase-(a) agent drafts for a
  subsidiary pair): ic-mirror-purchase raises the buyer's
  `purchase_invoice` from the sales invoice's materialization — supplier,
  owner and `mirror_of` all proper refs — and ic-mirror-post books it in
  the buyer's book, E3's convert composing with cascade (Alfa's entry in
  converted PLN, Beta's in EUR, one event). The killer-demo property is a
  test assertion: every object on both sides provenance-chains to the
  same root raw event, so the receivable and the payable cannot disagree
  — intercompany reconciliation matches by construction; there is nothing
  to reconcile, only to display. Transfer pricing is just the mirror's
  templates (here: invoice amounts). Residue, named: the general mirror
  needs per-pair rules until books compose (E2) and refs re-reference;
  statutory intercompany VAT treatment (reverse charge, WDT 0%) is pack
  depth blocked on conditional posting lines (E2 residue); an
  intercompany reconciliation *view* is display work for the
  reconciliation notion (SPEC, "later").

## 2026-10-04 — E3 attempt: currency fails as pure data; `convert` earns admission

- **The attempt** (`TestCurrencyAsPureData`): booking a EUR invoice into a
  PLN ledger with today's vocabulary means the precomputed-amounts
  contortion DIRECTION's network note predicted — the event carries
  `gross_pln`/`net_pln`/`vat_pln` computed outside the system. It books,
  and the explanation is hollow: the system holds EUR/PLN 4.3215 for the
  invoice date, the event claims amounts computed at 4.50, and the ledger
  agrees with the claim without a murmur. The rate table is decorative;
  provenance points at a raw event that carries its own conversion; replay
  reproduces numbers by copying, not explaining; and the entry has no
  transaction-currency trace, so nothing can ever tie the PLN posting back
  to the EUR document, let alone to a rate.
- **Verdict: conversion must run at firing time in the kernel,
  parameterized by rule data** — the sub-language precedent (postings,
  `each`), NOT general expression syntax. The postings template earns an
  optional `convert` clause: `{to, date, rounding, rounding_account}` —
  deterministic method vocabulary in code (rate lookup, integer
  multiply-divide, declared rounding, plug line), choice and parameters as
  data. The AI composes clauses; it never authors arithmetic.
- **Rates are master data**: a conventional `fx_rate` object (code "<from>/
  <to>/<date>", base, quote, date, rate-as-decimal-string), materialized
  from `fx.rate.published` events by a plain rule — append-only facts, so
  every conversion is replay-fixed by construction. The composite `code` is
  carried by the publisher (the resolution-keys-are-one-namespace decision,
  third use). The statutory D-1 subtlety (NBP's table applies to the next
  day) is the rate publisher's job: it publishes under application dates —
  calendars stay out of the kernel.
- **Rounding is statutory, so it is declared**: per-line `half_up` (the one
  method until another earns its way in), and when per-line rounding breaks
  the functional balance, the kernel books the difference to the declared
  `rounding_account` as an explicit plug line — ledgers stay balanced by
  invariant, and the rounding residue is visible, not hidden.
- **The ledger is functional**: with `convert`, line amounts evaluate and
  balance in transaction currency, then book in functional currency, each
  posting carrying `tx_amount`/`tx_currency` (posting v3). `to` equal to
  the transaction currency is the identity conversion — one rule explains
  domestic and foreign invoices alike, finishing what E2's "currency does
  not route" started. A missing rate refuses the event into the worklist:
  unexplainable beats guessed.
- **Shipped** (`TestConvert`, pl pack v3): the rounding stance and the
  rounding account are statutory and therefore pack data — pl-post gains
  `convert{to PLN, date =$.issue_date, half_up, 756}` and account 756
  (różnice kursowe) joins the wzorcowy plan kont. State reads arrive as
  `core.Getter`, shared verbatim by executor and simulator with chain
  overlay — the kernel's one new surface. In cohabitation the pl-stat
  trial balance is currency-coherent for the first time: PLN plus
  converted-PLN, never EUR mixed in. Rate registration stays a base rule
  (rates have no market; a per-pack rule would recreate E2's capture):
  base rules — register-company, register-fx-rate — still have no home,
  and a base pack is recorded residue for later.

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

## 2026-10-05 — REA/OeBTO ingestion

- **ISO/IEC 15944-4 (OeBTO) cross-checked against the language; recorded in
  `REA.md`.** The standard is voluntary ontology, not compliance surface; we
  take vocabulary alignment (events/types/policy level, claims, regulator),
  park three candidate adoptions (commitment semantics for template/schedule/
  contract types at M5, duality as a rule-pattern name, regulator vocabulary
  for adapter docs), and refuse the subtype taxonomy into the kernel. Why:
  free standard lineage for positioning and agent drafting context; zero
  kernel impact.

## 2026-10-10 — the agent's output cap

- **The agent's `max_tokens` goes from 4000 to 16000.** Models that think
  before answering (Haiku 5.5, Sonnet 5.5 and later) spend part of the cap on
  thinking, and a bundle with two posting books or PZ lines was cut off at
  4000: "unexpected end of JSON input" on three of four nordwind voices and
  three of four helios voices under claude-haiku-5-5. At 16000 the operations
  corpus went from NOT DONE to DONE on the same model. The cap is a ceiling,
  not a target: Sonnet 4.6 answers are unchanged. Why: a cap that truncates
  measures the cap, not the author.

## 2026-10-10 — authoring is a conversation; the all-or-nothing verdict was a misreading

- **KK: the sentence "in accounting there is no document that passes one
  time and not the other" was about booking, not authoring.** One document
  with four rules must not be booked under three when the fourth fails —
  and the kernel already does exactly that: the whole chain books or
  nothing does (`evaluate`, pinned by TestMultiRuleFiring and the
  parallel-books test). The eval verdict I built on that sentence on
  2026-10-09 — every run of every paraphrase DONE or NOT DONE — was my
  construction, and it punished a refusal the gate caught (nothing booked)
  exactly like a draft that booked wrong state. The bar stays absolute
  where money is; the author gets the second chance a person gives.
- **The loop, in the product:** a rejected bundle keeps its reason on the
  record (the reject_bundle event), and "Draft again" on a rejected bundle
  asks with `after`: the shell reads the chain of rejected bundles behind
  it from the store (`shell.Conversation`, three deep), each with its
  reason and the proposal as the agent shaped it, and the agent answers
  the latest refusal instead of starting over. Every bundle now keeps the
  question it answers (`question`) and the rejected bundle it redrafts
  (`after_bundle`); an empty re-ask means "the same question again".
- **The loop, on the test bench:** the recorded approver sends a refused
  draft back once (`maxReasks = 1`) with the reason — a validation
  refusal, unparseable JSON, a failing dry run, "explains nothing" — and
  rejects the stored draft on the record first, so the bench walks the
  same path as the shell. The verdict keeps "DONE in every run" as the
  definition of done, drops the sermon, and prints the counts that matter
  now: first draft right on N of M questions, re-asked K, of those fixed J.
  Rule-mode tasks (the M0 corpus) stay one-shot.
- **Why:** the eval's harshness on the author was measuring my rule, not
  KK's; what KK wants is the exchange one has with an accountant — propose,
  explain, refuse with a reason, propose again — and a count of how often
  the first proposal was right is the honest number about the author.

## 2026-10-10 — people as data: the person is master data, the actor resolves to it

- **A `person` object type ships in a market-neutral pack, `packs/common/people.json`**:
  name, handle, role, a status with a lifecycle (active ⇄ away, either → left, nothing
  back from left), a deputy as `ref<person>`; role and deputy amendable. Three doors —
  `register_person`, `appoint_deputy`, `set_person_status` — and the three rules that
  explain their events. First pack outside `packs/<market>/`: `common/` holds what every
  market builds on. KK, 2026-10-10: "add person object now" — three scenarios of the
  interaction brief (who is notified, who may act, who disputes) key on a person, and
  identity retrofitted later is the hole replay never forgives.
- **The actor convention is data, not code.** A human signs events as `<door>:<handle>`
  (`shell:kk`, `cli:anna`); `person.handle` holds the part after the colon, and the
  person is found by the ordinary field lookup. No kernel code knows the type exists;
  `internal/pack/people_test.go` is the convention's only enforcement. The agent, the
  adapters and the kernel keep their own actor prefixes and resolve to nobody.
- **The pl pack's `task` and `task_schedule` move to v2: `requester` and `assignee` are
  `ref<person>`**, and the pack (v7) requires `person`. Every pack-loading test installs
  `common/people` first, so "requires" stays honest: a type that assigns work builds on
  the person.
- **No permissions, no login, no address.** `who` on activities stays inert (Phase 4 of
  the growth plan); the notification address arrives with the outbound adapter, as an
  amendable field, when Phase 4 asks for it. Email is not a field until something sends.

## 2026-10-11 — the channel registry: a channel is an activity triggered by code

- **A channel is an activity with a `channel` section** (`core.ChannelSpec`): the adapter
  it binds to and the code version it needs, the format it reads, what makes a document
  the same one twice, the reader's confidence (`exact`, or `model` with a threshold in
  (0, 1]), the event types it may append, and a volume class (`documents` or `stream`,
  declared now, honored when the high-volume treatment lands). No inputs: code triggers
  it. Versioned, approved and listed like every verb — the registry is the activity
  store, and no seventh concept (DIRECTION 2026-10-10, "Documents in", step 1 of its order).
- **The binary registers what it can read** (`adapter.Register` from each adapter's init:
  `ksef` v1, `clock` v1), and **`adapter.Bind` joins the two before every pass**: no active
  channel for the adapter, two of them, or a code version other than the binary's refuses
  the pass before anything is read. The clock binds the same way, so a log without the
  clock builtin refuses to open a day — `rhea init` seeds it (run once on existing DBs).
- **Every event a channel appends names the channel**: `activity_name` and
  `activity_version` on adapter events, the pair the schema documented as NULL for
  adapters. Invariant 5 reaches the raw side: KSeF evidence reads "ksef v1 by adapter:ksef".
- **The declaration is the contract**: an event of a type the channel never declared, or
  one without a dedup key, refuses the whole pass and appends nothing — an adapter bug,
  not residue.
- **The clock is the first channel, a builtin** seeded active at init (`calendar`, exact;
  `time.day_opened`, `time.month_opened`, `schedule.fired`). The namespace test exempts
  channels: time.* is deliberately unreserved (M5). **The Poland pack ships the KSeF
  channel** (pack v8, activity `ksef`): a draft like any verb, connected by the pack's
  approval; until then the adapter is implemented but not connected, and a pass is refused.
- **Two generic lists**: `adapter.Registry` joins declared channels with the binary's
  adapters (states: connected, awaiting approval, not implemented, version mismatch, not
  connected) → `GET /api/channels`, the Channels tab under Language, `rhea channels`, and
  one startup line per active channel in `rhea serve`.
- **Left for the brief's later steps**: `document.received` and the blob store, the batch
  append (step 2); the confidence threshold enforced for model readers (step 6 — declared
  and displayed today, inert like `who`).

## 2026-10-11 — the file is the fact: a blob store, a second open door, the batch pass

- **A file enters through a verb input of type `file`** (`core.FileInput`, admitted by
  activity inputs alongside `json`): over the wire `{name, media_type, content_base64}`,
  in-process `{name, media_type, bytes}`. The door computes the hash, keeps the bytes in
  the blob store inside the trigger's transaction, and the payload carries only the facts
  about the file — `{hash, name, media_type, size}` under the input's name. The bytes
  never enter the payload. The hash is the trigger's default dedup key when the caller
  gives none (`<door>/<hash>`): the same file through the same door is the same fact,
  unless the caller names it otherwise. Cap: 32 MB.
- **The blob store is an append-only table keyed by content hash** (`blob`; invariant 1's
  trigger covers it). `PutBlob` verifies the hash against the bytes and is a no-op on a
  known hash; `GetBlob` reads it back; `GET /api/blobs/{hash}` serves it with the media
  type it was received as, immutable. No company key yet (P3).
- **`document.received` is not reserved, but it is checked**: the store refuses one whose
  payload names no file or a file it does not hold — whoever emits it (the open door, a
  channel, a pack), the file comes first. The brief's rule "the file is the fact" is thus
  enforced at the one write into the log, not by a namespace. (M5's reasoning holds: who
  said it is visible via the actor; the missing bytes were the only lie worth refusing.)
- **`receive_document` is the second open door, a builtin**: inputs `file` (required),
  `source` (required — the person says where it came from: email, scan, the bank's site)
  and `occurred_at`. Required because a template reading `=$.source` fails on a missing
  path, and the packs' convention is to map only what the event always carries. The
  namespace test now exempts the two open doors. `rhea intake [-source S] [-date D] FILE`
  (source defaults to `upload`); the shell derives a file picker for any `file` input,
  so the drop zone is a derived form and no screen knows about documents.
- **`packs/common/documents.json`**: `received_document` (document-flagged; hash, name,
  media_type, size, source, status received → read | set_aside) and one rule registering
  it from `document.received`. The reader of step 4 will amend its status; the shell links
  any hash-shaped value and any file fact in a payload to the stored file.
- **An adapter pass appends in one transaction** (`Store.AppendEvents`: known dedup keys
  are skipped and counted via `ON CONFLICT DO NOTHING`, any other failure rolls the whole
  batch back; `ErrDuplicate` joins `IsDuplicate`). The clock keeps its per-day interleave
  by design. This is the high-volume treatment's cheap-now item, done.
- **Local DBs**: `rhea init` re-run on `rhea` and `rhea_pl` (the blob table and the new
  door); the documents pack loaded into both as a draft bundle for KK.

## 2026-10-11 — the simulated customer emits documents: FA XML and MT940, the facts as files

- **`Facts.Documents`** (`internal/sim/documents.go`): after generation, every purchase
  invoice from a Polish supplier becomes one FA-shaped XML file, as KSeF hands it to the
  buyer, and every month becomes one MT940 statement of the company's account. Each
  document names the events it carries by index, so the eval of step 4 can feed the file
  in place of the JSON fact and expect the same state. `rhea sim -documents DIR PERSONA`
  writes them under `DIR/<persona>/<YYYY-MM>/{ksef,bank}/`; the summary counts them.
- **The FA subset, not the schema.** Parties (NIP without the country prefix, name,
  country), dates, number, the WZ it settles, net and VAT in the slot of the one rate the
  persona uses (`P_13_1`/`P_14_1` for 23%, `_2` for 8%, `_3` for 5%, `_6` for 0%), gross,
  lines (name, unit, quantity, unit price, line net, rate) and payment (term, the
  supplier's account), under the FA(3) namespace with `KodFormularza="FA (3)"`. The
  file is named by its KSeF reference, derived deterministically from seller NIP, date
  and number in the `<NIP>-<YYYYMMDD>-<12 hex>-<2 hex>` shape. **The namespace, the
  schema version and the rate slots must be checked against the published FA(3) XSD
  before the first real KSeF pull** (step 4): the fake client and the reader are the
  experiment's; the files are shaped so a real reader would not need a different test.
- **The MT940 dialect is "sim-bank."** Tags `:20:` (STMTyyyymm), `:25:` (the company's
  IBAN), `:28C:` (n/year), `:60F:`/`:62F:` (balances, carried month to month from a
  fixed opening of 10 000,00), `:61:` (value date, entry date, C, amount with a comma,
  NTRF, NONREF, //sequence) and `:86:` as `PRZELEW PRZYCHODZACY|<title>|<counterparty>|
  <account>`. Every bank writes `:86:` differently; the reader keeps the raw line
  (DIRECTION, the one trap). CRLF, UTF-8; a bank's export may be CP1250 — a reader's
  concern, not the generator's.
- **Personas carry accounts**: `iban` on the company, customers and suppliers (derived
  from the VAT id when absent), and the bank line's counterparty now names the paying
  customer instead of "customer transfer" — a statement never says less. Expected state
  is unchanged: the documents are a representation, not new facts.
- **Foreign suppliers are not in KSeF.** Omikron GmbH's invoices get no XML; the summary
  says how many arrive as PDF, for the model-read channel of phase 2. Until then they
  enter as JSON through the open door, which is exactly "what no channel reads yet".
- **Sales invoices stay JSON**: they are the company's own act, not a document received.
  Payments of purchase invoices (debit lines) come with step 5, where the expected state
  gains "purchase invoice paid".

## 2026-10-11 — the KSeF reader: documents in, state out

- **`ksef_inbound` is the first reader** (`internal/adapter/ksef_inbound.go`): one pass
  scans `received_document` objects still `received` whose bytes are a Faktura, reads
  each into a `purchase.invoice.received` carrying the document's own numbers (net, VAT,
  gross, rate, number, dates, parties, the WZ it settles, payment term, the supplier's
  account) plus `document` (the file's hash) and `ksef_ref` (from the file name when it
  has the KSeF shape), and a `document.read` beside it. Exact: no interpretation. A
  document that contradicts itself — net plus VAT not the gross, lines not the net — is
  not read but reported as `document.unreadable` with the reason, into the worklist; an
  XML that is not an invoice is left alone. The pull from the KSeF API joins this reader
  when a client exists; until then files reach the store through the door.
- **The reader is declared by its own pack**, `packs/pl/ksef-inbound.json` (statutory
  warrant, requires `received_document`): the market's channel, approved like any verb.
  **The documents pack moves to v2**: `document.read` amends the received document to
  `read` by hash lookup. `document.*` events are a channel's bookkeeping, explained by
  the pack; `document.unreadable` has no rule on purpose.
- **Readers are looked up by name** (`adapter.Lookup`), so a corpus or `rhea pass READER`
  can run one; the outbound KSeF submission keeps its client and its verb. `Reads` gained
  `GetBlob`: an adapter reads objects and the files they name, still no writes.
- **The eval runs documents in.** A corpus gained `packs` (installed and approved before
  the intake), `files` (through `receive_document`) and `readers` (one pass each);
  `eval.Intake` is the shared intake, used by `rhea sim -play` too. Residue counts what
  came through the open doors and the channels. `sim.CompileDocuments(…, ["ksef"])`
  feeds the KSeF invoices as files, drops their JSON facts, expects every received
  document `read`; `rhea eval -documents-in PERSONA`, `rhea sim -play -documents-in`.
  **Both personas reach DONE with documents in** on the reference answers
  (`TestPersonaReferenceIsDoneWithDocuments`), with the same rules as before: the reader's
  event has the generator's shape plus the document's extra fields.
- **Personas name their packs** (`packs`, relative to the persona file). Foreign
  suppliers' invoices stay JSON until the PDF reader; the bank statement stays JSON until
  its reader (step 5), when the matching rule and "purchase invoice paid" arrive.
- **Not done, by choice:** the purchase-invoice booking rules as Poland-pack drafts (the
  reference rules of the personas book them today; the pack's drafts with warrants come
  when the pack gains its purchase side, with the VAT purchase register), and the default
  of documents-in for `rhea eval` — flipped after one live run proves it.

## 2026-10-11 — the bank reader and reconciliation by rule: Phase 1's exit, on the bench

- **`bank_mt940` is the second reader** (`internal/adapter/mt940.go`): every entry of a
  stored statement becomes one `bank.statement.line` — value date, amount (never
  negative) with `direction` credit or debit, the title as `reference`, the counterparty
  and its account, our account, the statement number, the line number, the `:86:` text
  raw, and `document` (the file's hash) — plus a `document.read`. The reader checks the
  statement against itself (opening plus credits minus debits is the closing) and reports
  a contradiction as `document.unreadable`. It understands one `:86:` dialect, sim-bank
  (`kind|title|counterparty|account`); any other bank's line becomes the title whole,
  nothing guessed — a real bank's dialect is a second parser behind the same channel.
  Declared by `packs/pl/bank-mt940.json`; `rhea pass bank_mt940`.
- **The generator pays both ways.** A share of purchase invoices is paid by us (debit
  lines naming the supplier's invoice), beside the customers' payments; every bank line
  carries direction, accounts and counterparty; statements carry the debits. The expected
  state gains "purchase invoice paid". Both personas' interviews say so: the purchase
  invoice has a status (open → paid), and the bank answer distinguishes money in (pays a
  sales invoice) from money out (pays a purchase invoice) — two rules with a `where` on
  direction, because one rule for both would fail its lookup on the other side's lines.
  The KSeF reader and the generator now name the buyer (`buyer_nip`) on a purchase
  invoice, so the owner is a lookup like the supplier.
- **Readers run in the order the sources are given** (`ksef`, then `bank`): invoices
  before the lines that pay them. `rhea eval -documents-in` and `rhea sim -play
  -documents-in` feed both.
- **Phase 1's exit holds on the bench:** a KSeF invoice and its bank payment flow in as
  files, book with the document's VAT, and reconcile by rule — both personas DONE on the
  reference answers with documents in (`TestPersonaReferenceIsDoneWithDocuments`). The
  live proof (the model drafting from the paraphrased answers, documents in) is the next
  measurement; documents-in becomes the default after it.
- **Matching is by reference only.** A line that names no invoice stays in the worklist;
  matching by amount and counterparty over the open invoices (the fold under the cap)
  arrives with the matching grid in Phase 2, where a person pairs what rules cannot.
- **The Poland pack gains its purchase side (v9)**: `pl-book-purchase-invoice` (supplier
  and owner by NIP lookup, amounts as the document states them — the KSeF reader hands
  the document's own VAT and gross, so the pack trusts the document, as the sales side
  trusts the KSeF sales invoice), `pl-post-purchase-invoice` (Wn 300 net, Wn 221 input
  VAT, Ma 202 gross, converted like the sales entry) and the analysis view
  `rejestr-vat-zakupu`, the sales register mirrored on the owner. No "paid" rules in the
  pack: the base invoice types carry no status yet; the personas' bundles add it, and the
  base seed will when the matching grid lands.
- **Live, 2026-10-11, documents in** (`rhea eval -documents-in -months 1 -v`, one run per
  persona, scripted voice, claude-sonnet-4-6): Helios 6 of 6 questions approved on the
  first draft, 29 intake facts, residue 0, state matches; Nordwind 9 of 9, 52 facts,
  residue 0, state matches. Both DONE, no re-asks. The reader's extra fields (the
  document's VAT and gross, the hash, the KSeF reference, the direction) did not distract
  the author. One run each is the gate for the default, not the all-or-nothing verdict:
  that verdict stays a round of twelve runs over the voices, to be run before the demo.
- **Documents in is now the default** for a persona: `rhea eval PERSONA` and `rhea sim
  -play PERSONA` feed the KSeF invoices and the bank statements as files; `-documents-in=false`
  plays the JSON facts, which the reference test still proves separately.

## 2026-10-11 — the matching attempt: by reference carries, everything else fails as data

- **The attempt** (`TestMatchingAttempt`, kept as executable evidence): the bank story
  with what exists — rules, lookups, folds, amendments, postings — and the places it
  must fail, per the brief. Six customers: by reference, by exact amount, a partial
  payment, a tolerance, a subset that sums, a match across N sides.
- **By reference carries** — the rule of today, the personas' bank-payments bundle: a
  line naming an invoice amends it to paid, and a read through the link even moves
  the open amount (`ref(sales_invoice, number, $.state.reference).open - $.state.amount`).
- **By exact amount fails on the key.** A lookup keys on one field, so "the open
  invoice of this customer with this amount" can only be said as "any invoice with
  this amount". It works once, by luck; the moment another customer owes the same
  123.00 the lookup is ambiguous, and atomicity makes the gap total: the bank line
  itself does not materialize. A paid invoice never leaves the key either, so the
  ambiguity is permanent. The side must be keyed by party and filtered by openness
  — a collection, not a lookup.
- **A partial payment fails across invoices.** One named invoice's open amount moves
  by amendment, and "paid when nothing stays open, else as it is" is the type's to
  consent to — a lifecycle that declares `open → open` admits it, as data (with the
  lifecycle as the personas declare it, `open → paid` only, the whole payment is
  refused and nothing half-applies). What no data expresses is the remainder
  settling the next oldest invoice: an amendment has exactly one target.
- **A tolerance fails on the posting line.** The two-grosz difference needs a line
  that exists only when there is a difference; lines are not conditional and amounts
  must be positive, so the rule books a short payment and refuses every exact one
  ("amount must be positive, got 0").
- **A subset that sums and a match across N sides cannot even be attempted:** no
  fold enumerates subsets, and no effect touches N objects by an allocation.
- **Verdict: `matching` is a closed-class effect**, as `book`, `convert` and `amend`
  were — a kernel method with N sides and a declared structure, strategies as named
  standards ordered by a policy that is data (DIRECTION, the matching brief). The
  attempt also names one small generalization the consequences of a match need in
  the ordinary language: an amendment over a fan-out (`each` on amend — one event
  that moves N objects). The amendment law itself stands: a status set to what it
  already is is a transition the type declares, never a kernel exception — a second
  resolution of a resolved case still waits for a person.
- **The effect is spelled `matching`, not `match`.** A rule already has a `match`
  (its predicate); one word with two meanings in one JSON object is a trap for the
  author and the reader. `matching` sits beside `postings` — a noun for the engine's
  work, as `postings` is for the ledger's.

## 2026-10-11 — the review of step 5: what the bank reader owed a real bank

The code review of the step-5 commit found fifteen things; the twelve that needed no
decision are fixed here, the rest (stale voices, the registers in foreign currency, a
"paid" rule in the pack) wait for a word from KK.

- **A file holds several statements, a statement holds several pages.** The reader
  read a file as exactly one statement and overwrote the balances as it went, so a
  bank's quarterly export, or a month split over pages (`:62M:`/`:60M:`), was refused as
  "contradicts itself". Now every `:20:` is a statement, every balance page is checked
  on its own (opening plus entries equals closing), a page must open where the page
  before closed, and an entry outside a page is an error. The line numbers run through
  the file (the dedup key is file and line), the read is dated by the last closing.
- **A SWIFT header is not a reason to skip the file.** `{1:…}` block lines, a `:940:`
  line and the `-}` terminator are dropped; the first tag must still be `:20:`.
- **An amount pads to two decimals.** The format allows `1500,5`; the money parser does
  not. One helper reads every amount, balances and entries alike.
- **A reversal says so.** `RC`/`RD` in `:61:` kept only its direction's effect; the line
  now carries `reversal: true` (absent otherwise, so every existing event is unchanged),
  so a rule can tell a bounced transfer from a payment. The field is on the event, not
  in the channel pack, which stays at version 1.
- **The dialect needs a known kind in front.** Any `|` in `:86:` made the reader guess
  the sim-bank layout; a foreign bank's "TRANSFER | ref FAB/1/2026" lost its first half.
  Now only four fields behind `PRZELEW PRZYCHODZACY`/`PRZELEW WYCHODZACY` are the
  dialect; everything else is the title whole, one line.
- **The bank channel reads the bank door.** The pass fetched every received text file
  from every source to sniff it, and left a bank-sourced file that was no statement
  received forever, told to no one. Now it reads only `source = bank`; a file through
  that door that is not a statement becomes `document.unreadable` with the reason
  ("the file starts with …, not :20:"). A statement that came through another door is
  that door's business.
- **The buyer's NIP carries the buyer's country.** The KSeF reader wrote "PL" in front
  of whatever the buyer's NIP was, and a bare "PL" when there was none. Now the prefix
  is the buyer's `KodKraju` (PL when the document names none, as for the seller), and an
  invoice without a buyer NIP carries no `buyer_nip` at all — the lookup then fails on
  a missing field, not on a made-up one.
- **A purchase invoice is stamped with its market: the buyer's.** On a sales invoice
  the market is the seller's — the book it lands in; on a purchase invoice it is the
  buyer's, for the same reason. The KSeF reader stamps `market: pl` on every invoice it
  reads (KSeF is the Polish system, and the inbound channel reads what is addressed to
  a Polish buyer); the generator stamps its purchase facts likewise, the German
  supplier's included, since Nordwind buys for its Polish books.
- **The two purchase rules guard on the market, like the sales rules; the pack is v10.**
  Without the guard a two-market installation had the Polish pack claim a German
  subsidiary's supplier invoice — posting it into `pl-stat` or parking it on a failed
  lookup. The cohabitation test now books Alfa's supplier invoice and leaves Beta's in
  the worklist, where a German purchase side will find it.
- **One shape for `purchase.invoice.received`.** The generator's JSON fact carried the
  net and the rate only; the KSeF reader hands the document's own VAT and gross, and
  the pack trusts those — so an invoice in the first shape parked in the worklist on a
  missing field. The generator now emits `vat`, `gross` and `market` too, and the JSON
  path and the document path give the same event. A persona's own rules still compute
  the VAT from the net; the eval checks they agree with the document.
- **One `core.Abs`.** The executor and the generator each kept a private copy of the
  same int64 magnitude; it lives in core beside the money formatters now.
- **One settle loop in the generator.** Sales and purchases were paid by two copies of
  the same loop, differing in the direction and in how the party was kept (a name to
  look up against a Party). Both invoice sides are one `invoiceFact` carrying its party,
  settled by one closure; the random draws come in the same order, so every generated
  persona is byte-for-byte what it was.


## 2026-10-11 — the matching engine's core: N sides, a declared structure, strategies as data

- **One engine, pure** (`internal/match`): a Problem is the trigger item, the other
  sides' candidates, the measures and the policy; Solve runs the steps in order and
  the first feasible match wins. Numbers are exact rationals; nothing is a float.
  Every search is bounded: the side cap (`core.DefaultMatchingCap` = 200 open
  candidates; above it "a person decides"), the subset size cap (`max_items`, default
  4, at most 8) and a node budget (100 000; above it the step says "above its cap" and
  the policy moves on). The same problem always yields the same result: candidates
  are ordered by date then log order, a subset is the fewest items first and among
  those the oldest first, exact takes the oldest agreeing item.
- **Four strategies, standards with names**: `reference` (the items the trigger's
  reference names as whole tokens — "FV/1" is not named by "FV/12"; feasible when
  their open amounts agree, or with `partial` whenever something is named: a short
  payment settles part, an over-payment keeps its rest open on the line), `exact`
  (one item per side agreeing on every allocated measure), `subset` (the fewest items
  summing to the trigger), `oldest_first` (the trigger settles the earliest open
  items as far as it reaches; one allocated measure). Parameters any step may take:
  `tolerance` and `tolerance_percent` per measure with a `difference_account` (the
  difference is absorbed and recorded, positive, with a direction: "short" when the
  trigger brought less than the side settles, "over" when more), `within_days`.
  Deferred, named: `proportional` allocation (largest remainder) when a customer
  asks; FX converters and netting are the next step of the session.
- **The structure is the template** (`core.MatchingTemplate`, effect `matching`):
  a name (what the match event carries, so consequence rules key on it and two
  triggers over one structure share them), the trigger side, N sides — each a type,
  an `items` read (a keyed `objects(...)` written against the trigger's state; the
  trigger side has none), `open` per allocated measure (an amendable field: the
  match's consequence is an amendment), optional `compare` per compared measure (a
  price: must agree, never consumed), `currency`, `date`, `reference`, `cap` — and the
  policy. Measures are inferred from the fields: every side declares the same
  allocated measures; a measure is money, int or decimal, the same on every side. The
  rule's match must be `object.materialized` where `$.object_type` is the trigger's
  type; one rule per trigger side (the late-sides case is a second rule over the same
  name, written against its own trigger). Two gates: `Validate` (shape, strategies,
  parameters, the trigger named in the match) and `Check` (types, field kinds,
  amendability, the items read a collection of the side's type).
- **The outcome is one event, baked.** `objects.matched` carries the name, the
  strategy and step, the trigger, the measures' kinds, every item on every side with
  `allocated`, `open_before`, `open_after` (money as minor units), currency, date and
  reference, every difference with its kind, side, amount, direction, currency and
  account (resolved at firing: a declared account must exist), every step tried, and
  the inputs the engine read (the candidates per side — every object under the key,
  settled ones included — and the tolerances). `objects.unmatched` carries every step
  tried with its reason, or the one reason no step could run: a side above its cap, a
  side that cannot be read (a payer no company carries the account of — the line
  materializes and waits; its own chain is never held hostage by the match).
- **Ordinary rules explain the match.** A match is a chain point like a
  materialization: rules matching `objects.matched` (where `$.matching` eq the name)
  fire in the next generation. The open amounts move by `amend` with `each` over
  `$.sides.<side>[*]` — the one generalization the attempt named: one event that
  moves N objects, `target` and `set` evaluating against {doc, line, n, root}. The
  settlement is an `object` with `each` over the same items; the differences are
  objects with `each` over `$.differences[*]` (guarded by `$.differences[0]` exists,
  since an empty fan-out is a rule error) that a cascade posts. A status set to what
  it already is is a transition the type declares (`open → open` under a part
  payment), never a kernel exception. An amendment of an object materialized
  earlier in the same chain is a sequence, not a fight: the conflict guard now
  admits exactly that (a bank line born, matched and settled by one fact) and still
  refuses two materializations or two amendments of one object in one chain.
- **Replay never re-matches**: the match events touch no object; the amendments and
  materializations they caused are baked, and replay re-applies them. The
  simulator shares the path verbatim (`SimulateEvents` shows an invoice paid and the
  settlement added, nothing booked); backfill replays prior match events as chain
  points, so consequences approved later reach old matches through the gate.
  Ids of what a match causes build on `<name>-<trigger id>`, rooted like every id.
- **The walk shows it**: `/api/explain` carries `matched` / `unmatched` on the node
  and the shell renders the strategy, every item with allocated and open, every
  difference, every step tried before. The agent's grammar gains EFFECT 4 with the
  bank story and its consequence rules.
- **First customer passes** (`TestMatchingBankStory`): by reference; exact keyed by
  the payer (Orion's identical 123.00 is no candidate — the attempt's ambiguity is
  gone); three invoices at once, the fewest that explain 750.00; two grosz short
  within the tolerance, the difference as an object posting Wn 755 / Ma 201; a part
  payment oldest first leaving 400.00 open, settled exactly later; a payer who owes
  nothing unmatched with every step's reason; an unknown payer's line open, its chain
  intact; replay identical; the dry run shared. `TestMatchingGate` pins the refusals.
- **A named limit**: a side's read is one field, so it returns a customer's whole
  history under the key and the engine filters the open ones; the VM's collection cap
  (1000) bounds the read. When a customer's history outgrows it, the keyed-open read
  (a two-field key, or an open-items projection) is the scale item — paid then.
- **KK, during the session**: each company has its own payment-matching policy —
  which is what the policy being data on the company's rule gives; the pack only
  proposes. And intercompany matching for group consolidation is one of the
  important kinds — the brief lists it among the three-sided structures; it becomes
  this session's generality customer (customer 5), not an afterthought.

## 2026-10-11 — the persona scene: the simulated customer pays in every way the policy names

- **The generator pays in every way** (`sim.payments`): each sales invoice chosen to be
  paid (`paid_share`) is paid one of six ways by a fixed draw — named whole (the rule
  of today, most of them), unnamed at the exact amount, several invoices of one
  customer in one unnamed line (two or three, grouped per customer), named two grosz
  short, named in part with the rest named later, or an unnamed part payment that the
  policy reads as the customer's oldest open invoice, the rest named later. Purchase
  invoices are paid by us whole, naming them. One stray line per run, from a payer no
  company carries the account of, stays open for the person.
- **The expected state is predicted without any rule.** The generator walks the
  planned payments in date order with every invoice's open amount and records what
  the policy will settle: paid invoices, a settlement per allocation, a difference per
  short payment, the line settled or open. A scenario the policy could read two ways —
  an unnamed amount two open invoices share, a sum a smaller or a second subset reaches,
  a part that is itself a whole amount — falls back to naming the invoice; the fall-back
  is the honest answer, not a weaker check. In the eval every invoice of the period
  exists before any line is matched (the intake precedes the interview), so "open"
  counts them all; a live run sees only the past ones, and the scenarios hold either
  way: the oldest open invoice is a past one, and fewer candidates make nothing
  ambiguous. The facts carry the bank story's identity, checked a second way from the
  events: money in = the gross of the sales invoices it paid − the differences written
  off + what it paid of invoices still open + the stray line; money out = the gross of
  the purchase invoices it paid (`TestFactsAreConsistent`).
- **The personas carry the policy** (`testdata/customers/*.json`): the master-data
  answer adds the company's IBAN (the key a bank line is matched by; the generator
  emits it); the invoice answers add the open amount ("starts at the gross and goes
  down as payments are matched; a partly paid invoice stays open" — `open` amendable,
  `open → open` declared); the bank answer states the owner's policy in business words
  and its reference bundle is the first full matching bundle: `bank_line`, `settlement`
  (the payment, the document type and id, the amount, the date), `matching_difference`;
  two matching rules over one shape — money in against the payer's open sales
  invoices with the policy *reference; reference ± 0.02 → 755; exact; subset ≤ 4;
  reference partial; oldest first*, money out against the supplier's open purchase
  invoices without the tolerance; the consequences as ordinary rules (the open
  amounts by `each` over the match's items, the settlement per item, the difference
  as an object, the short one posted Wn 755 / Ma 201 in pl-stat). Account 755 joins
  both charts; the posting policy counts two lines per difference.
- **Both personas DONE** on the reference answers, facts as JSON and as documents
  (`TestPersonaReferenceIsDone`, `…WithDocuments`): every bank line but the stray one
  settled, every settlement and difference counted and summed, every invoice's
  status as predicted. The voices are regenerated for the changed answers.
- **Named, not done here**: the Poland pack's own version of the policy as drafts
  with warrants (the pack's invoice types gain the open amount first); a one-sided
  time window ("not issued after the payment") when a customer asks — `within_days`
  is symmetric; a live run over the voices with the model drafting the matching
  bundle — the agent knows the effect, the measurement is owed with the twelve-run
  round.

## 2026-10-11 — FX and netting: the converter is a clause of the structure; netting is a structure

- **The second customer, an EUR invoice paid in PLN** (`TestMatchingFX`, `TestConverter`):
  the structure gains `convert` — `{functional, date_side, rounding, rate_up_account,
  rate_down_account}`, the FX precedent of postings' `convert` as a clause. The engine
  runs it as a pre-pass and a post-pass around the unchanged strategies: every candidate's
  open money is read in the trigger's currency at the fx_rate of the settlement date —
  the date of the item on `date_side`: the trigger's own when it is that side (a payment
  against invoices), each candidate's own when it is on that side (an invoice against the
  payments waiting for it) — so exact, subset, reference and oldest first compare like
  with like; then every allocation is written back in the item's own currency (a whole
  item takes its own open exactly, a partial one the rounded equivalent, by the declared
  method), with the equivalent kept beside it in the event (`equivalent`,
  `equivalent_currency`). The rate's move between an item's own date and its settlement
  — what the allocated amount was worth in the functional currency then and now — is an
  `fx` difference, `up` or `down`, against the matching account; whether that is a gain
  or a loss depends on the side, and the rules that post it know (a receivable settled at
  a higher rate: Wn 201 / Ma 750). Rates are master data read as postings read them
  (`fx_rate` by code, the inverse pair inverted), every rate read is an input of the
  match, and a missing rate refuses the firing into the worklist — unexplainable beats
  guessed, as E3 decided. 1000 EUR at 4.3215 is 4321.50 PLN exactly; a 2000 PLN part is
  462.80 EUR, its settled value the 2000.00 actually allocated, its booked value
  1990.04 at 4.30, the difference 9.96 up. Replay reproduces the converted match without
  a rate in sight.
- **The converter works on one allocated money measure** (the structure declares no
  other), and when `date_side` is not the trigger's side and the trigger converts, its
  settlement date is the latest allocated item's on that side — a simplification named
  here, enough for two sides, revisited if a three-sided FX structure ever needs more.
- **Solve never mutates its problem**: the converted view is the engine's own copy — a
  bug the converter test caught on its second solve.
- **The third customer, netting per saldo** (`TestMatchingNetting`): nothing new in the
  engine. A structure whose trigger is our purchase invoice and whose other side is the
  same party's open sales invoices (`objects(sales_invoice, buyer, $.state.supplier)`),
  with exact, subset and oldest first: their 500 nets against our 300 and 200; a 400
  payable against a 1000 receivable nets whole and leaves 600 open. The match is the
  kompensata: a `netting` document per allocated receivable, posting Wn 202 / Ma 201.
  The kodeks cywilny art. 498 warrant belongs to the pack draft when the pack ships it.
- **The agent** reads the convert clause and the netting structure in its grammar.

## 2026-10-11 — N sides: three-way with the PZ, intercompany as the generality test

- **The join is the one thing the N-sided customers asked of the structure**: a side
  keyed by the order alone returns every line of that order, so the candidates must
  also agree with the trigger on the item. `join: {"<candidate field>": "<trigger
  field>"}` on a side — the keys that join candidates beyond the one the read is keyed
  by (DIRECTION, the brief's "which keys join candidates") — typed by the gate (both
  fields exist, the same type), applied at firing (a candidate that disagrees is not on
  the side at all), recorded as an input. No engine change.
- **The fourth customer, three-way** (`TestMatchingThreeWay`): the purchase order's
  lines, the receipts' lines (two PZ for one line) and the supplier's invoice lines,
  headers from the facts and lines by cascade; the invoice line triggers; the order's
  and the receipts' lines are read by the order and joined on the item; the quantity is
  allocated (an `int` measure), the price compared within two percent (a `compare`
  measure on the sides that carry it; receipts carry none); the exact step fails on the
  receipts (60 and 40) and says so, the subset finds both; every line on every side
  closes by the same `each` amendment; a `three_way_match` record per invoice line
  (`$.sides.order[0].object_id` — an indexed path into the match); the price gap sits on
  the order line in the event (`item` on the difference), becomes a `price_variance`
  object with `amount` = per unit × the quantity allocated (a formula over the match's
  numbers) and posts Wn 341 / Ma 300.
- **The fifth customer, intercompany, the generality test** (`TestMatchingIntercompany`;
  KK's note): Alfa's bank line from Beta settles Alfa's receivable and Beta's payable
  at once — three sides from a different corner, the payer's documents keyed by the
  payer and joined on us (`seller`/`supplier` = the line's `owner`), so Gamma's invoice
  to Beta with the same number stays open; the match is the elimination's evidence, an
  `intercompany_settlement` naming all three. Nothing was added to the engine. The
  sides cross companies through refs at firing, as the partition note says; the company
  key as a partition key (P3) will serialize there.
- **The agent** reads `join` in its grammar.

## 2026-10-11 — the grid: the remainder is a view, and a person's pairing is the same event

- **The `matching` notion** (`core.MatchingSpec{rule}`, `shell/matching.go`, the schema's
  notion CHECK): the N-sided view of what no strategy settled. The view names a matching
  rule; the rule's structure says what the sides are and which fields matter — the
  label, the reference, the date, what is open, its currency, what is compared — so the
  view adds no opinion of its own, and reconciliation, bank matching and intercompany
  matching are one notion (the growth plan's count). Every side lists its items with
  something open; a trigger item still open carries why the policy left it, from the
  latest `objects.unmatched` for it, step by step. Typed cells as every view's.
- **The pairing verb** `pair_items` (a builtin, `matching.paired`, a `pairing` JSON input
  `{matching, trigger, items}`): the kernel's reaction runs the matching rule's structure
  on the person's selection under the engine's `selected` strategy — never a policy's
  step: the chosen items oldest first as far as the trigger reaches when one measure is
  allocated, exact agreement when several — and books the same chain the policy would
  have: the `objects.matched` event (strategy `selected`, step 0, the structure as its
  rule), then the rules that explain it, in the transaction that appends the pairing.
  One fact, explained the same way, whoever made it: the walk shows the person as the
  root's actor and the structure as the match's rule. A pairing that names nothing on a
  side, or a trigger of the wrong kind, is refused before anything is appended; the
  `matching.` namespace is the kernel's (reserved, never residue). The chain gained a
  seed — a first node the root carries — and booking inside a caller's transaction.
- **The grid in the shell**: N columns, one per side, a radio on the trigger side and
  checkboxes on the others, the why beside an open trigger item, one button. The
  bundle gate admits a matching view that names a matching rule of the same answer;
  both personas' bank bundles carry the incoming and outgoing grids and stay DONE
  (`TestMatchingGridAndPairing`, `TestPersonaReference…`).
- **Named, not done here**: unpairing — a correcting event that restores the open
  amounts, explained by the same rules with the allocations negated and the
  differences' directions flipped (the lifecycle then declares `paid → open` as
  consent); a write-off verb for a difference a person accepts; the host proposing a
  strategy after a person has paired the same way twice.

## 2026-10-11 — the host proposes a strategy after a person has paired the same way twice

- **A pairing says which standards would have found it.** On every pairing by hand the
  kernel runs each learnable standard alone — reference, reference partial, exact, subset,
  oldest first — over the sides as the policy reads them, and records in the match event
  (`standards`) those that would have chosen exactly the same items on every side. The
  first time, nothing more happens: it is noted in the event.
- **The second time, the host proposes.** When an earlier pairing of the same structure
  was explained by the same standard, the policy lacks it, and it was never proposed
  before, the pairing's reaction drafts the matching rule's next version with the step
  appended (subset with the size the pairing needed, at least the default) and a bundle
  `learned-<structure>-<standard>` with a client warrant naming both pairings, in the
  pairing's own transaction — behind the gate: the owner approves it in Bundles, or
  rejects it, and a rejected proposal never returns (the bundle id is one per structure
  and standard; "the overruled rule never returns"). The verb's response and the
  shell's toast name the proposal.
- **Proven in the grid test**: two part payments paired by hand with the oldest open
  invoice are both what oldest first would have done; the second drafts version 2 of
  the rule with oldest first as its last step, warranted by the two events; the owner
  approves; the third part payment matches by itself, by the learned step, and the grid
  shows the invoice's open amount moved.
- **What this is not**: the interview loop (the brief's "the last with the interview
  loop") — the host's proposal is deterministic kernel code over the engine's own
  standards; the model is not asked. When the host learns a *structure* (a new side, a
  new key) from what a person did, that is the interview's job and the agent's.

## 2026-10-13 — the Poland pack's default matching policy: pack pl v11

- **The pack proposes the standard; the owner overrides** (DIRECTION, "Rhea suggests
  standards"; KK 2026-10-11: each company its own policy). Pack pl v11 ships, as drafts
  in its one bundle with a practice warrant: the invoice types with an open amount and
  a status (`sales_invoice` v4, `purchase_invoice` v2 — open starts at the gross, the
  booking rules fill it), the bank operation as the channel delivers it
  (`bank_operation` v2: direction, title, counterparty and its account, our account as
  the statement names it, open, `settlement_status` unsettled → settled/private), a
  `matching_difference`, the company's account (`company` v2, `iban` amendable), and
  fifteen rules: the account as an enrichment — copied from the registration when it
  carries one (a cascade on the company's materialization reading `$.root.iban`) and
  from the supplier's KSeF invoice (`supplier_account`); the operation from the line,
  posting Wn 130 / Ma 201 on money in and Wn 202 / Ma 130 on money out; two matching
  rules — money in against the payer's open sales invoices with *reference; reference
  ± 0.02 → 756; exact; subset ≤ 4; reference partial; oldest first*, money out against
  the supplier's open purchase invoices without the tolerance, both keyed by the
  counterparty's account; the settlements of invoices and operations, a `settlement`
  per allocated document (the pack's own type, unchanged), the difference as an object
  posting Wn 756 / Ma 201; and the two grids, "Wpłaty do rozliczenia" and "Wypłaty do
  rozliczenia". The pack's chart keeps 756 ("różnice kursowe i zaokrąglenia") for the
  tolerance, as it keeps it for convert's rounding.
- **The story test** (`TestM3PolandPack`) plays it: the buyer registered with an
  account, a payment naming the invoice posts to the bank and settles the invoice, the
  same party paid as a supplier settles the purchase, the trial balance carries 130 and
  the cleared 201 and 202, two settlements name their documents, and a payment nothing
  fits waits on the grid with every step's reason.
- **Two things the pack cannot say yet, named**: the bank-operation rule carries no
  market guard — the bank line carries no market, and `where` sees only the payload
  (the E2 lesson): the channel that appends a line should stamp the market of the pack
  that declares it, in the registry; until then a second market's bank pack would
  conflict on every line. And the customer's account is known only when the
  registration or an invoice carries it — the loop that learns a payer's account from
  the first payment matched by reference (the statement names the invoice; the invoice
  names the buyer) wants a read through a link from the match event, which is untyped;
  the white-list adapter (wykaz podatników VAT) is the statutory source when it comes.
- **Evolution stays additive.** The first cut made `open` and `status` required on the
  pack's new invoice versions, and the German pack's booking rule and the intercompany
  mirror — other authors' rules on the same shared types — refused at once ("required
  field open missing"). The fields are optional: the pack's own booking rules fill them,
  an invoice born by another rule or under an older version simply has nothing open —
  the engine reads a missing open amount as "nothing open", never a candidate, never an
  error (a trigger without one is still malformed and refuses). SPEC §7's schema
  evolution question answered the late-understanding way: a new version adds, it never
  demands of the rules that came before it.
- **Upgrading an installed pack** stays out of scope: the loader leaves known rule ids
  alone, so v11's reshaped booking rules reach a fresh installation only. The
  experiment reinstalls.

## 2026-10-13 — unpairing: a correcting event the same rules explain

- **The unpairing verb** `unpair_items` (a builtin, `matching.unpaired`, inputs: the
  match event and a reason) undoes one match by a reversal: an `objects.matched` with
  strategy `reversal`, `reversal_of` naming the match, the same structure and trigger,
  every allocation negated with the open amount as it is now before and restored
  after (later matches may have moved it), the equivalents negated, every difference
  with its direction flipped (short ↔ over, up ↔ down). The kernel seeds it into the
  chain under the match's own rule version — the structure that explains both — and
  the rules that explained the match explain its undoing: the open amounts move back
  by the same amendments, a settlement for the negative amount records the undoing,
  the flipped difference posts the opposite entry (the pack gained
  `pl-post-over-difference`, Wn 201 / Ma 756, which also books a tolerated overpayment).
  Never a deletion: the match, the reversal and both sets of consequences stay in the
  log; replay re-applies them in order.
- **Consent as data, again**: an invoice open again after its match is undone is a
  transition its type declares (`paid → open`, `settled → unsettled`), as the partial
  payment's `open → open` was. The pack's types and the personas' declare it; the
  personas' invoice answers say so.
- **A trigger is matched more than once over its life** — the policy's match, undone,
  then paired by hand — so a match's chain point is no longer one per structure and
  trigger: the event carries its sequence on the trigger (`seq`, the matches before it),
  and the chain point is `<structure>-<trigger>[-<seq>][-undo-<event>]`, so what each
  match causes (settlements, differences) never collides with an earlier match's;
  backfill keys prior firings the same way. The test's re-pairing found this as a
  duplicate object id.
  The match event now records which field holds each side's open amount
  (`inputs["open.<side>.<measure>"]`), so a reversal reads the amounts as they are now
  without the structure.
- **A verb's fact is dated like a document's.** The pack's and the personas' settlement
  rules date themselves by the root fact's `date`, and a pairing's or an unpairing's
  event carried none — the grid test had not recorded a settlement, so it never showed.
  Both verbs' payloads now carry `date`, the act's business date, beside the person's
  inputs; a settlement made by hand is dated the day it was made.
- **Refused**: undoing a match twice, undoing a reversal (pair again instead), undoing
  an event that is not a match. `TestMatchingUnpairing`: a short payment matched with
  its difference posted, undone — invoice and line open again, the negative
  settlement, the flipped difference and its opposite entry netting the postings to
  zero — then paired again by hand; replay identical.

## 2026-10-13 — a channel stamps the market of the pack that declares it

- **The gap** (named with pack v11): the bank line carries no market, `where` sees only
  the payload, so the pack's bank-operation rule could not guard on the market the way
  every other pack rule does since E2 — and a second market's bank pack would have
  fired on every Polish line. The KSeF reader had side-stepped it by writing `market:
  pl` into every invoice it reads, which is code knowing a market.
- **The channel declares it.** `ChannelSpec.market` is the market whose facts the
  channel delivers — the pack's, since the pack declares the channel — and the runner
  stamps it into every event payload the channel appends that carries none, beside the
  channel's name and version on the envelope: the market is provenance of the raw
  side, declared as data, never guessed by a reader. The Polish channels declare `pl`;
  the pack's bank-operation rule guards on it; the registry shows it. A line submitted
  through the open door says its market itself, as the invoices in the story test do.


## 2026-10-13 — live: the model drafts the matching bundle from the interview

- **Both personas DONE on one live run each**, scripted voice, claude-sonnet-4-6,
  `rhea eval -months 1 -v` with the documents in: Helios six of six answers approved
  on the first draft (31 intake facts, residue 0, state matches), Nordwind nine of nine
  (52 facts, residue 0, state matches). The bank answer became one bundle of three
  types, eleven rules and the grid: `bank_line` open on its amount, two matching rules
  keyed by the payer's account — incoming with *reference; exact ± 0.02 → 755; exact;
  subset ≤ 4; reference partial; oldest first*, outgoing without the tolerance — the
  amendments over the match's items, the settlement per item, the difference and its
  posting. The model put the tolerance on the exact step where the reference answer
  puts it on the reference step; the generator's short payments name their invoice,
  and exact-within-two-grosz found the same invoice, so the state agreed. Different
  policy, same standard: the grammar's example shaped the draft, as it should.
- **What this measures**: that the fourth effect is authorable from an owner's words,
  first time, with the structure's keys, the fan-out amendments and the guarded
  difference rule all composed correctly. The twelve-run round over the voices is the
  measurement proper; the voices are being regenerated for the changed answers.
- **The round over the voices, Helios** (`rhea eval -voices -months 1`, the scripted
  interview and three paraphrases, one run each): DONE in every run, 4 of 4; first
  draft right on 24 of 24 questions, none re-asked — the matching bundle drafted
  correctly in every voice, the policy composed from the owner's words each time.
  **Nordwind**: DONE in every run, 4 of 4; first draft right on 35 of 36, one answer
  re-asked once (the model returned a bundle without its id and description — a slip
  of form, refused by the gate and fixed on the re-ask), the matching bundle right in
  every voice. One run per voice this time; the three-run round over the voices, the
  measurement the growth plan counts by, is the next one.
- **The paraphrase check learned that "four" is "4"**: three retries in a row dropped
  the digit of "at most 4" because the model spells small numbers out, which keeps the
  fact. `Preserved` now accepts a number word for a digit up to twelve, and still
  reports a number that is gone.

## 2026-10-13 — two debts: the overruled rule never returns; the FA(3) check

- **The standing test** (`TestOverruledRuleNeverReturns`, DIRECTION "The host
  remembers"): a reading approved, then replaced by another with a reason. Acting: the
  old reading explained what arrived in its time and never fires again, the new one
  explains what came after, replay reproduces both as they happened. Drafting: the
  author's ask names the new version as the active rule and shows the old one only as
  a correction — version, what it said, who replaced it, why — and the prompt says in
  so many words that corrected readings are never proposed again. What this needed:
  `approve_rule` takes an optional `reason` (the correction's, on the record in the
  `rule.approved` event; the shell's approve endpoint passes it), and the ask gains
  `Corrections`, read from the approval events of every active rule. Speaking (the
  briefing) and learning (the network's supersession) wait for their slices, as the
  note said.
- **The FA(3) check, against the schema the CRD publishes** (`crd.gov.pl/wzor/2025/06/
  25/13775/schemat.xsd`, the Ministry's brochure): the namespace, `kodSystemowy "FA
  (3)"`, `wersjaSchemy "1-0E"`, `WariantFormularza 3`, `RodzajFaktury VAT`, the
  parties' `DaneIdentyfikacyjne` and `Adres`, the line fields `P_7`, `P_8A`, `P_8B`,
  `P_9A`, `P_11`, `P_12`, `Fa/WZ`, `Platnosc/TerminPlatnosci/Termin` and
  `Platnosc/RachunekBankowy/NrRB` are as the sim writes and the reader reads. The rate
  slots: `P_13_1/P_14_1` the standard rate (23 or 22%), `_2` the first reduced (8 or
  7%), `_3` the second (5%) — right. **One error**: FA(3) has no `P_13_6`; a zero-rated
  sale goes in `P_13_6_1` (domestic), `P_13_6_2` (intra-EU supply) or `P_13_6_3`
  (export), and none of them has a VAT element. The sim wrote `P_13_6` and a `P_14_6`;
  the reader read `P_13_6` and nothing else. Fixed on both sides: the sim writes
  `P_13_6_1` with no VAT element for 0%, the reader takes all three zero slots as 0%
  and `P_13_7` as exempt. Neither persona uses a zero rate, so the bench never saw it
  — the schema did.

## 2026-10-13 — the three-run round over the voices, Helios

- `rhea eval -voices -runs 3 -months 1`: the scripted interview and three paraphrases,
  three runs each — **DONE in every run, 12 of 12; first draft right on 71 of 72
  questions, 1 re-asked and fixed.** The one re-ask was the master-data answer in one
  run: the model wrote two company rules, one for a registration with an account and
  one without, and the dry run refused them for materializing the same company —
  the conflict guard doing its job, the re-ask fixing it. The matching bundle was right
  on the first draft in all twelve runs.
- **Nordwind: DONE in every run, 12 of 12; first draft right on 104 of 108, 4 re-asked
  and all fixed.** Three of the four were the bank answer in one voice — the longest
  answer of the interview — coming back as a bundle without its id and description,
  refused by the gate's first check and fixed on the re-ask; the fourth the PZ answer
  once. Nothing wrong in substance: the matching policy was composed right every time
  it was drafted. **The round, both personas: 24 of 24 runs DONE, 175 of 180 answers
  right on the first draft, every re-ask fixed.** The one author habit this round
  caught: a long answer sometimes loses the bundle's envelope — the next prompt fix,
  measured before it is kept.


## 2026-10-13 — the scanned invoice: the first model-read channel, behind the gate

- **The channel** (`packs/pl/pdf-invoice.json`, adapter `pdf_invoice`, intake brief step
  6): a supplier's invoice as a PDF — the German supplier after KSeF, the real reason a
  Polish company still receives PDFs. The reader asks the model for the invoice's fields
  and how sure it is (`agent.Read`: the PDF's bytes as a document block beside the
  instruction; the answer is data), validates the answer against the event's shape and
  the invoice's own arithmetic (net plus VAT is the gross, or `document.unreadable`),
  and applies the channel's confidence policy, which the runner hands the reader from
  its declaration (`Bound`): at or above the threshold the reading is
  `purchase.invoice.received`, the same shape as the KSeF reader's, naming its reader
  and confidence and file; below it a `document.reading_proposed` waits in the
  worklist, with the proposed event, the confidence and the reason. Without a model
  the reader reads the text layer and says so (reader `text-layer`, confidence 1.00):
  the bench runs the channel without a model call. The channel says its market; the
  pack's threshold is 0.90, a default with its warrant, any company's to change.
- **A person is a reader too.** `confirm_reading` is a builtin door: confirming a
  proposal makes the appended event the proposed typed event itself — the person as
  its actor, the proposal as its cause (so it leaves the worklist), the document read
  — and the proposal stays in the log as what the model thought. Confirmed twice is
  refused. A wrong booked reading is corrected by amendment, as the brief says; a
  wrong proposal is left unconfirmed and the fact submitted through the open door —
  a "correct and confirm" door is named, not built.
- **The simulator emits the PDFs**: one page of labelled lines in the supplier's words
  (a German layout), written by a PDF writer of our own — uncompressed, one Type1 font,
  a correct cross-reference table — so a person, a text extractor and the model read
  the same thing; each names the fact it carries. `rhea sim`, `-documents-in` and the
  eval feed the sources in the order `ksef`, `pdf`, `bank`: invoices before the lines
  that pay them. Nordwind's German supplier now arrives as PDF; Helios has no foreign
  supplier. Both personas stay DONE with the PDFs in (`TestPersonaReferenceIsDone…`).
- **Live, the model reading the PDFs** (`rhea eval -months 2 -v`, Nordwind, scripted
  voice, claude-sonnet-4-6, documents in): the two months' German invoices entered as
  PDFs and were read by the model (the eval hands the channel the run's agent), the
  readings booked by the pack's and the persona's rules — 80 intake facts, residue 0,
  state matches, DONE; one re-ask on the bank answer, the envelope slip, fixed. The
  numbers the model read were the numbers the facts carry, or the state would not
  have matched. The confidences it gave are not in the report yet — the next pass
  prints the readers and their confidences beside the residue. The worklist shows a
  proposal with its reason and one button.

## 2026-10-13 — the company key: the partition key the store carries on every call

- **The one scale item cheap now and expensive later** (DIRECTION 2026-10-08, scale
  rule 1, security rule 2): the company is promoted from a payload field to a key the
  store knows. `event.company` and `object.company` (indexed with the id and the type),
  `core.Event.Company` and `core.Object.Company`. A door stamps the fact with the
  caller's scope; everything derived from it — the chain's events, the objects, the
  amendments, the matches — inherits it in `bookTx`; replay keeps it. Threaded as an
  argument, not a global: the context every store call already carries holds the scope
  (`store.WithCompany`, `store.CompanyOf`); the shell takes it from the request
  (`X-Rhea-Company` or `?company=`), the CLI from `RHEA_COMPANY`. The key is a string by
  convention the company's VAT id — it must exist before the company's first object does.
- **A scoped read sees one company's facts alone; an unscoped read, the whole log.**
  The reads of facts and objects — the worklist and the pending set, the events by
  kind, the objects by type, the lookups rules resolve by, the counts, the whole cache
  — filter by the scope when there is one, and behave exactly as before when there is
  none. Nothing enforces a scope yet, as the note says: isolation and sharding ride on
  this column later; the executor's own reads at fire time stay unscoped, which is the
  cross-company read the note names (E4's intercompany cascade crosses through refs
  there, by design) and the next partition question.
- **`TestPartitionKey`**: two companies in one log, each fact stamped from its door and
  inherited by its chain; each company's reads see its own notes, lookups and counts;
  a fact no rule explains waits in its own company's worklist; replay keeps the key.
  History from before the key, and facts appended without a scope, carry none and
  show in every scoped read's complement — the whole-log read.

## 2026-10-13 — the briefing: the narrative notion, the host's words as facts

- **The narrative notion** (`core.NarrativeSpec{facts, voice}`, `shell/narrative.go`): the
  concierge's first scenario (DIRECTION 2026-10-10), with its five parts as data. The
  trigger is the person opening the view; the facts are the views the spec names —
  analysis tables and matching grids, computed by the kernel as every view is; the
  host is the agent, handed the facts as JSON, the owner's voice, the date and what it
  said before, and asked to phrase them in a few sentences (`agent.Narrate`,
  `NarrateSystem`); what it says is logged through the `host_said` door (a builtin the
  agent speaks through, `host.said` in a kernel namespace, never residue), so the next
  words know the last — never a templated welcome. The shell shows the words, then
  the facts beneath, each as its own view.
- **The host never computes a number.** Every number in its words is checked against
  the facts digit for digit (separators dropped, "1 234,56" is "1234.56", small
  counts and dates' parts pass); a foreign number refuses the words — they are not
  logged, the view says why, and the facts speak alone. The model's text is data,
  validated, as every other output of it (invariant 3). Without an agent, or when it
  fails, the view is silent and the facts are shown: the briefing degrades to the
  tables, never to a template.
- **The pack ships "Poranek"** (pl v12): the sales and purchase VAT registers and the
  two grids, in Polish, "jak najlepszy concierge". The story test reads it without a
  host; `TestBriefing` reads it with one: the words from the facts, logged by the
  agent; the second time told what it said; computed numbers refused and not logged;
  silence without a host.
- **Named, not done**: the queue notion ("what to do today": open cases ranked by a
  formula) and the hotspot rules in the pack — the next two scenarios; the briefing as
  a scene in the personas' day (the eval asking the view and checking the host's
  numbers against the facts), with the live measurement of the host's words.

## 2026-10-13 — the author's habit named: the grammar's schematics leak into the answer

- **The envelope slip was never about the envelope.** The refusals in the round said
  "bundle needs bundle_id and description" because the decoder had silently stopped
  earlier; the verbose reruns, once a refusal kept the text around the fault, showed
  the real habit: on the longest answer the model copies the grammar's *schematic*
  notation — `"sides": {<side>: [{"object_id", ..., "allocated": {<measure>: amount},
  "open_before": {...}}]}` — verbatim into its JSON, placeholders and all, and the
  answer stops being JSON. My first fix, an arrow annotation beside the shape
  ("<- ALWAYS first"), was copied the same way and made it worse. The rule for the
  grammar, kept from now on: **every shape the author is shown is a concrete instance
  that parses, never a schematic with angle brackets, ellipses or arrows**; prose says
  what varies. The match event is now shown as the real payload of a payment
  settling two invoices with two grosz written off. `rhea eval -v` keeps the text
  around a syntax error in the refusal, so the next habit is read, not guessed.
- **The decoder reads the answer the model meant.** Two more runs of the same voice
  showed the rest of the habit: the model prefaces its bundle with prose carrying JSON
  snippets (`{"strategy": "reference"}`, even a `{"rule_id": …, "description": …}`
  example), and the decoder took the first balanced object — a snippet — so the gate
  said "no envelope", or worse, read a snippet with a description as "already answered"
  and approved nothing. The decoder now takes the largest object in the text that
  looks like an answer (one carrying `bundle_id` or `rule_id`); failing that, the first
  object that parses, so the gate names what is missing; and when the first object
  fails to parse and nothing else looks like an answer, the text from that brace, so
  the real syntax error is reported where it is — never "no envelope" for a bundle that
  broke in the middle (`TestExtractJSONTakesTheAnswerNotTheSnippet`).
- **Verified** on the phrasing that slipped three times in the round: DONE, nine of
  nine answers right on the first draft, the matching bundle approved first time.
  The model's text is data; reading it well is the kernel's job, not the prompt's.

## 2026-10-14 — the queue: what to do today, ranked by a formula on the thing itself

- **The queue notion** (`core.QueueSpec{object_type, where, rank, columns, limit,
  approvals}`, `shell/queue.go`): the concierge's second scenario (DIRECTION
  2026-10-10) with nothing new in the language. What is open is a condition over the
  object's state (`$.state.status eq open`); how urgent it is is a formula over the
  object with today in the payload (`=$.today - $.state.due_date`: the formula
  language's date arithmetic, so "days overdue" is arithmetic the kernel already has,
  never code), highest first, ties in log order; what to show is a column list. Views
  compute nothing into the log: a rank is read-side arithmetic over baked state, with
  the same lookups and reads through links a rule's formula has.
- **The verbs ride on the item.** The queue declares none: every item carries the
  detail view's contextual verbs — every active activity with a ref input of the
  item's type — as buttons that open the door prefilled with the item. One mechanism
  for "the one or two verbs that resolve it, inline", wherever an object is shown.
- **The approvals sit above the items** when the view asks (`approvals: true`): the
  draft bundles waiting at the gate, each a link to the gate. The person's day begins
  with what only they can do.
- **An item the formula cannot rank is listed, not dropped**: a task with no due date
  ranks last, with the formula's reason where its rank would be, and the queue stands.
  The formula language has no null literal to guard with, and a queue that fails
  because one item is incomplete would hide the other nineteen.
- **The pack ships two** (pl v13): "Do zrobienia dziś" — open tasks ranked by days past
  due, weighted ten a day, plus a premium for priority, with the approvals — and
  "Należności do pilnowania" — open receivables ranked by days outstanding. The story
  test finds the one open receivable in the second; `TestQueue` checks the ranking
  (13, 4, −6 days), the paid one absent, the columns, the verbs that take an invoice
  and none that do not, the approval waiting.
- **Named, not done**: the hotspot rules in the pack (thresholds raising cases, which
  then appear in this queue); a queue over several types at once (a day has invoices
  and tasks in it — one view per type for now, the simplest option); the queue as a
  scene in the personas' day.


## 2026-10-14 — the live check of the host's words: one recorded morning, a gated test

- **The check is a measurement, not an assertion.** `TestHostWordsLive` in
  `internal/shell` runs only with `RHEA_LIVE_HOST=1` (`RHEA_LIVE_HOST_RUNS`, default 5):
  it hands the model named by `RHEA_MODEL` one recorded morning and asks it to speak
  several times in a row, each time told what it said before, and prints every answer
  with the kernel's verdict — refused for a number the facts do not carry (the same
  `foreignNumbers` the shell applies), refused for repeating itself, accepted otherwise.
  The simplest option: no company is stood up, no event is logged, and the model switch
  is the one the eval already uses (`ANTHROPIC_BASE_URL`, `RHEA_MODEL`).
- **The morning is recorded, not invented.** `testdata/eval/briefing.json` is the
  Poland story's briefing step verbatim — the four facts the kernel handed the host
  (both VAT registers, both payment grids) with the pack's Polish voice and a date in
  the story's month. It was captured from the running story test once and committed as
  data; a richer morning (a persona's) can join it later as a second entry.
- **Why now:** the public-service note (DIRECTION 2026-10-14) splits the model's work
  into two jobs, writing rules and speaking to the owner, and KK asked for the
  measurement before KSeF outbound. The rule corpora already had `rhea eval`; the
  host's words had no bench. Local candidates measured through Ollama's Anthropic
  endpoint: qwen3.6:27b, Bielik 11B v3.0 and Bielik 4.5B v3.0 (Q8_0).


## 2026-10-15 — the host never denies a fact: a second law, and thinking off for the words

- **The gap the small models exposed, closed.** The host's words had one law — a number
  the facts do not carry refuses the words — and the measurement of 2026-10-14 found
  its blind side: Bielik said "no outgoing payments waiting" over a grid with a
  100.00 payment on it, and the kernel let it through, because a denial carries no
  number. The second law: *a fact with an amount in it must be spoken, or the words
  are refused.* `unspokenFacts` in `internal/shell` mirrors `foreignNumbers`: a fact is
  spoken when at least one of its amounts appears in the words digit for digit; a
  fact's amounts are the non-zero cells of its typed money and count columns (an
  analysis view's declared columns, a matching grid's open amounts), or every number
  its cells carry when no column is typed. The refusal names the fact and its
  amounts: "they pass over wypłaty bez rozliczenia (100.00)".
- **An empty fact asks nothing.** No rows, no open items, only zeros: "nothing waits"
  is then the truth, not a denial. The simplest option; the alternative — the host
  must mention every fact, empty ones included — would make a quiet month a list of
  absences.
- **One amount is enough.** A register with three rate rows is spoken when one of its
  amounts is; a count ("2 notes") names a fact as well as a total does. The floor is
  deliberately low — the law is against denial, not for completeness — and it can be
  raised without touching the shape.
- **The host is told the law** in `NarrateSystem`: never deny or pass over a fact, speak
  at least one amount of every fact that has one, call only an empty fact empty. The
  door's description says the same. The live check (`TestHostWordsLive`) applies the
  new verdict between the foreign-number check and the repetition check.
- **Thinking off for the words.** `Agent.Speak` is the same call as `Complete` with
  `thinking: {type: disabled}` sent explicitly; `Narrate` speaks through it (falling
  back to `Complete` when a test injects only that). Drafting and reading keep the
  model's default. Measured on the local endpoint before wiring: qwen3.6:27b warm
  answers a one-line question in 0.9 s with the parameter and 3.6 s with a thinking
  block without it; Ollama 0.40 honours the parameter on its Anthropic endpoint, and
  Anthropic's own API accepts "disabled" as the default it already is.
- **Tests:** `TestHostLaw` runs the recorded morning (`testdata/eval/briefing.json`)
  through both laws with the shapes the measurement produced — the denial refused by
  name, the truthful morning accepted, dropped decimals still foreign, a register
  passed over named with its amounts, an emptied grid callable empty, zeros asking
  nothing, an untyped table counting every number. `TestBriefing` gains the step: a
  quiet day over two notes is refused, "2 notes" alone is accepted.
- **Numbers are compared as values, not digits.** The first law read "1000 zł" as a
  number the facts (1000.00) do not carry and refused it; nothing was rounded, so that
  refusal was false. `canonicalNumber` now drops the zeros after the decimal point:
  "100", "100,00" and "100.00" are one value, "1234" is still not "1234.56" (a decimal
  that was there and is gone is a rounding). Both laws share the comparison. The
  2026-10-14 measurement's "decimals dropped from every amount" against Bielik 11B was
  this false refusal; its other two findings (the same text five times, the denied
  payment) stand.
- **Measured with the laws as they now stand** (2026-10-15, same laptop, Ollama 0.40,
  five mornings each, thinking off through `Speak`): qwen3.6:27b 4/5 accepted at 4–5 s
  a morning, against 40–78 s yesterday with thinking on — the one refusal was "nothing
  has changed since the last report", which passed over all four facts, the second law
  doing its work; Bielik 11B 1/5 at 3–8 s — told the law, it now names the waiting
  payment, the first morning is accepted, and the four after it are the same text word
  for word, refused for repeating itself.
- **A limit, named.** Bielik's accepted morning still said "no outgoing payments
  waiting". The law let it through because the story's one bank operation sits open on
  *both* payment grids, so the 100 it named for the incoming side also covered the
  outgoing one: when two facts share every amount, numbers cannot tell a denial of one
  from a mention of the other. The remedy is in the grid, not the law — corrected the
  same day: the pack's two rules already match on the operation's direction (credit
  in, debit out), but the grid view lists every open object of the side's type and
  ignores the trigger rule's where, so one credit shows on both grids; the grid
  should apply the trigger's where to its trigger side (a small fix, owed). A law
  that attributes an amount to a fact would need to read sentences, which the kernel
  does not do.
- **Live databases keep the door's old text.** Builtin activities are seeded once and
  left alone; the law lives in the shell's check, not in the door's description, so
  nothing needs migrating.


## 2026-10-15 — numbers as people write them, one reader; the grid lists what its rule would fire on

- **One prose reader of numbers, in core.** KK asked whether number cleaning, trimming
  and comparing belong in the arithmetic engine. No: the engine computes on exact
  values only, and reading a written number is a reader's job. Every channel already
  reads its own fixed spelling into the strict boundary form (`ParseMoney`: a dot, two
  decimals); what was missing was the reader for prose — the host's words today, a
  mail or a chat line tomorrow — which lived in the shell as a regex.
  `core.WrittenNumbers` / `core.ReadWrittenNumber` now hold it: spaces (plain,
  non-breaking, narrow) group thousands; a dot or a comma is the decimal mark; both
  kinds present, the last one is the decimal mark; one kind several times groups
  thousands unless the last tail is not three digits; a lone mark is the decimal mark
  (the continental reading: "1,234" is one and 234 thousandths — a thousands group is
  written with a space or beside a decimal mark). The value is canonical (no grouping,
  a dot, no leading zeros, no trailing zeros after the point), so comparison is value
  equality. A sign is not read. The host's two laws use it; the channel readers keep
  their strict forms, where a stray character should fail rather than be read.
- **The grid's trigger side applies the rule's where.** The limit named this morning
  was the grid's: it listed every open object of the trigger's type, so a credit sat on
  the outgoing grid whose rule matches only debits. `matchingData` now evaluates the
  trigger rule's where on the same payload shape the executor fires on
  (`core.EvalCondition` over `{object_id, object_type, state}`) and lists only what the
  rule would fire on. The Poland story asserts the outgoing grid is empty while the
  open credit waits on the incoming one. The recorded morning
  (`testdata/eval/briefing.json`) keeps its older shape — it is a fixture of the
  shared-amount case, which the law test still exercises.
- **Currencies (answered, not built).** Conversion exists where it belongs — postings'
  `convert` and matching's `convert`, the rate a fact, the date and rounding declared,
  the residue to a named account, a missing rate to the worklist. What the formula
  language lacks is the *type*: money does not know its currency, so a rule could add
  EUR to PLN unrefused. Planned, not now: money carries its currency; mixing is refused
  at the gate; `convert` joins the formula vocabulary over the same rate facts; minor
  units per currency as a data table when the first such currency appears. Due with
  the first euro purchase invoice — the VAT register wants it in PLN at the NBP average
  rate of the last business day before the tax point (VAT act art. 31a), JPK_V7 wants
  PLN — as a pack rule with its warrant plus an NBP rates channel; placed after the
  statutory floor and before the private beta.


## 2026-10-15 — KSeF outbound: the national API 2.0 client, FA(3) written from objects, rejection as evidence

- **The brief first** (DIRECTION, "Documents out"): test environment with a random NIP
  and no account, demo with Novi's account for the onboarding rehearsal, production
  only as Novi's real first month; onboarding is one pasted token and the walk to it is
  pack data; secrets never enter the log; one interactive session per pass; the
  evidence — number, UPO, rejection — is an event; performance is the limits.
- **The client is the standard library.** `adapter.KSeF2` speaks KSeF API 2.0 as the
  Ministry's docs lay it out: the published keys (`/security/public-key-certificates`,
  by usage), a challenge, the token encrypted with its timestamp (RSA-OAEP SHA-256,
  `token|timestampMs`), the authentication operation polled past "in progress", the
  access token redeemed; a session declaring FA (3) 1-0E with a fresh AES-256 key and
  IV wrapped with the Ministry's symmetric-key certificate; each invoice AES-CBC
  PKCS#7, both hashes and sizes; each invoice polled to a final code; the UPO XML
  fetched per accepted invoice; the session closed (a session that will not close
  times out by itself). Error bodies are read in both of the API's shapes, so a
  refusal carries the Ministry's own words. Three environments by short name
  (`KSeFEnvironments`); keys of one never work in another.
- **The FA(3) writer is shared.** `adapter.WriteFA3` renders a sales invoice in the
  schema's order with the mandatory annotation block (cash method, self-invoicing,
  reverse charge, split payment each "no"; exemption, new means of transport and
  margin each "none", the exemption declared when the rate is `zw`), the rate in its
  slot (23/22 → 1, 8/7 → 2, 5 → 3, 0 → 6_1 as "0 KR", zw → 7, no VAT element on the
  zero slots), one line for the whole net when no lines are given, the payment block
  only when a term or an account is known, the address falling back to the name
  (`company` v3 gains an optional `address`). The simulator writes through the same
  function, so what a persona receives and what Rhea sends are one document — the
  simulator's own copy of the types is gone. The Ministry's verification doc is the
  spec: UTF-8, no BOM, no processing instructions, the declared schema.
- **The adapter, code version 2.** `KSeF.Pass` reads sales invoices, companies and VAT
  rates, writes each unsent invoice, hashes it, sends all of them in one session, and
  returns `ksef.invoice.submitted` (number, KSeF number, the UPO's session reference as
  `upo`, the UPO XML as `upo_xml`, the date, the hash) or `ksef.invoice.rejected`
  (number, code, the reason in one line, the details, the date, the hash). An invoice
  with an accepted submission is done; one whose exact XML was rejected is not sent
  again — a person changes something, the hash changes, it goes. A session that fails
  in transport or authentication is a pass error with nothing appended; the next pass
  retries, as the contract says. Dedup keys: `ksef/submit/<invoice id>`,
  `ksef/reject/<invoice id>/<hash prefix>`. A KSeF number is CRC-8 checked
  (`KSeFNumberValid`, polynomial 0x07) and the fake's references now carry a valid
  checksum.
- **The pack, v14.** `ksef_submission` v2: `status` (accepted | rejected), `ksef_ref`
  and `upo` optional, `reason`, `invoice_hash`. Rules: `pl-record-ksef-submission` v2
  (status accepted, hash), `pl-record-ksef-rejection` (status rejected, the reason),
  `pl-ksef-rejection-task` (a cascade: a rejected submission is a task for its date,
  priority high, category invoices, the reason as description — it lands in "Do
  zrobienia dziś"). The `ksef` channel v2 declares code version 2, both events and
  the market. Existing databases keep v13's rows (a pack load leaves known ids alone);
  the two shells need a migration before the next live demo, as the queue did.
- **The CLI.** `rhea ksef` uses the national API when `RHEA_KSEF_TOKEN` is set, with
  `RHEA_KSEF_ENV` (test | demo | prod) and `RHEA_KSEF_NIP`, and the deterministic fake
  otherwise; the fake takes a `Reject` map so a demo can show a refusal.
- **Tests.** `TestKSeF2AgainstTheFakeMinistry`: an httptest server that implements the
  endpoints to the letter, decrypts what the client sent with its own keys and checks
  the hashes — the plain XML the Ministry would read is byte-for-byte what Rhea wrote;
  polling past "in progress" twice; a duplicate (440) with the original number in the
  details; a wrong token refused in the Ministry's words. `TestKSeFOutbound` runs the
  adapter through the runner with the pack: two invoices, one accepted into a
  submission with the number and UPO, one rejected into a submission and a task; the
  second pass sends nothing; a later invoice goes alone. `TestWriteFA3`: the order,
  the annotation block, the slots, zw and 0 KR, the refusals. `TestKSeFNumber`: the
  documented example validates. `TestKSeFLive` (gated by `RHEA_LIVE_KSEF=1`) sends one
  invoice to the test environment and prints what came back; never against prod.
- **Owed.** The live run itself: it needs a test-environment token, which only a
  one-time signed authentication can mint — KK takes the portal walk (the onboarding
  rehearsal, written down as taken) and the token goes into `.env`. The QR codes on an
  invoice shown outside KSeF. The migration of the live shells to v14. A pass that dies
  between sending and status would re-send next time and meet a 440 duplicate carrying
  the original number — the rejection then tells the person; rare, noted.
- **LIVE, 2026-10-15 evening: the first invoice Rhea issued to KSeF.** KK took the portal
  walk on the test taxpayer app (test authentication, the random NIP 8058083055,
  "Tokeny" → "Generuj token" with issuing and reading) and pasted the token into `.env`.
  The token's own shape is three pipe-separated parts (a reference, `nip-<NIP>`, a
  64-hex secret), so the value must be quoted in `.env` — a shell reads the pipes as
  pipes; the whole string is the token the API wants. First live send: authenticated,
  session open, invoice sent, **rejected 450** with the Ministry's exact words — the
  buyer block (`Podmiot2`) was incomplete: FA(3) requires `JST` and `GV` (local
  government unit, VAT group; "2" for no) after the address, which the sim's writer
  never had and no local validator could tell us. Added; second live send: **accepted
  200**, KSeF number `8058083055-20261015-84B870400000-11` (CRC-8 valid), the UPO XML
  naming the session, the token reference, the invoice number and the acceptance time.
  Twelve seconds end to end for one invoice, most of it the Ministry's processing. The
  test environment is the validator of record, as the brief said: it found in one
  round what a schema read by eye missed.


## 2026-10-15 — local packs, clean kernel: the Polish code moves to markets/pl; the evidence keeps the document and the Ministry's times

- **Invariant 8** (CLAUDE.md; DIRECTION "Local packs, clean kernel"): the kernel never
  imports a market. KK: "I don't want to contaminate Rhea with local specific
  solutions; if we fix the Polish pack, we fix the Polish pack, not Rhea's kernel."
  `markets/pl` now holds KSeF in, KSeF out and the FA(3) writer; `internal/adapter`
  keeps the contract, the runner, the registry, the clock and the EU-wide readers
  (MT940, the model-read PDF) plus the three names every reader of the document store
  shares (`ReceivedDocumentType`, `EventDocumentRead`, `EventDocumentUnreadable`).
  `adapter.Provide(name, factory)` is the exported seam beside `Register`: a market
  registers what the binary carries and what can be run by name; the CLI's Polish
  verb is gone (`rhea pass ksef` is the generic form, the client built from the
  environment by the market, a misnamed environment a client that refuses). The binary
  links each market by one blank import. `markets/markets_test.go` lists every
  package's imports and fails if a kernel package imports a market, or a market
  another market. Two named exceptions outside the kernel: the binary's import line
  and the simulator, which writes a Polish persona's documents through the Polish
  writer — to move into the market when a second market's personas exist.
- **The evidence is complete** (KK's question: how do we store the invoice, the date,
  the proof). The acceptance event now carries the FA(3) document exactly as sent
  (`invoice_xml`) beside the receipt, so "show me what you sent" is answerable from the
  log; the Ministry's two times are lifted from the receipt into fields — `sent_at`
  (DataPrzeslaniaDokumentu: the day a structured invoice counts as issued, VAT act
  art. 106na) and `accepted_at` (DataNadaniaNumeruKSeF); and the receipt's own
  fingerprint of the document (SkrotDokumentu, base64 SHA-256) must equal the
  fingerprint of what was sent, and its KSeF number the session's — otherwise the
  evidence is refused: nothing booked, the pass error names the invoice
  (`TestKSeFOutbound`, the lying receipt). `ParseUPO` reads the receipt as the
  Ministry returns it. Pack v15: `ksef_submission` v3 with `sent_at` and
  `accepted_at`, the recording rule v3, the "Wysyłki KSeF" view v2 showing status and
  the Ministry's time. The live check reads the real receipt back and verifies it.
- **Not done, by choice:** the document is kept in the event, not the blob store — the
  append-only log is the evidence store and an adapter returns events only; a few
  kilobytes per invoice is the right price for one rule. Live shells still run v13;
  the migration to v15 is owed before the next demo.


## 2026-10-15 — month end, part one: the settlement is a rule, the close is a door

- **`month(d)` joins the formula language**: the period a date falls in, "YYYY-MM" — the
  key `period_lock`, the registers (DuckDB `substr`) and a settlement share. One
  function, typed to return a string, documented in the grammar the agent is told.
  The simplest option over a substring function: a period is a thing the language
  names, not a slice of text.
- **The close is a common door** (`packs/common/periods.json`, pack `common-periods`):
  `close_period` takes any day of the month and the book; `period.closed` is the fact;
  the common rule `lock-period` makes the `period_lock` (month, book) that postings
  already refuse into. Market-neutral by construction — every VAT regime closes months —
  so no market declares its own.
- **The settlement is the Poland pack's rule** (v16): invoices carry `vat_period` from
  booking (`=month($.issue_date)` on sales, `=month($.invoice_date)` on purchases — the
  deduction month is the invoice month until a person amends it); on `period.closed`
  for book `pl-stat`, `pl-settle-vat` materializes one `vat_settlement` whose every
  field is a fold over the month's invoices — bases and VAT by rate through the link to
  the rate's code, output VAT, purchase VAT, the excess carried in from the previous
  month's settlement (`objects(vat_settlement, period, month(add_months($.period,
  -1)))`, empty for the first month), VAT due or the excess forward. The declaration's
  positions are thus kernel facts with provenance: the walk shows the rule, the event
  and the inputs each fold read, and replay reproduces them (the story's determinism
  check covers the settlement). The list view "Rozliczenie VAT" shows the month.
- **The VAT period lives on the pack's invoice versions** (sales_invoice v5,
  purchase_invoice v3), not on a base seed: the pack already owns the later versions
  of both types (open amounts, status), and a competing base version of the same
  number collided — the loader validates a rule against the type it knows. Noted for
  the base's market-neutral read when Germany lands: `vat_period` is neutral, the pack
  is merely where the type's current version happens to be.
- **The worklist is re-read when the world changes, and the lock applies to what was
  late**: closing September re-tried the unbalanced September invoice still waiting,
  which is now refused by the lock and says so. The story asserts it, and that a late
  posting after the close is refused per event naming the lock.
- **Next:** the file — the writer seam, `markets/pl` rendering JPK_V7M(3) from the
  settlement and the invoices, the common documents pack's `issue_document` door and
  `issued_document` type, validated against the Ministry's XSD with the scratch
  validator (lxml; this morning's FA(3) validates against the schema offline too).


## 2026-10-15 — month end, part two: the file is a market's writer, handing it over is a door

- **The writer seam** (`internal/adapter/writers.go`): beside the adapter (talks to an
  outside system, returns events) stands the writer — reads objects and parameters,
  returns a named file — `adapter.ProvideWriter`, `LookupWriter`, `WriterNames`,
  `Render`. The kernel renders any writer by name and knows none: `rhea file WRITER
  k=v…` (to stdout), `GET /api/files` (what the binary renders), `GET
  /api/files/{writer}?k=v` (the file). Deterministic in the objects: the same state
  and parameters give the same bytes (`created` is a parameter so a test can pin it).
- **Handing a file over is the kernel's third open door.** `issue_document` (builtin,
  beside `submit_event` and `receive_document`): the writer, its parameters (spelled
  one way, sorted `k=v` pairs), the file; the bytes kept under their hash in the same
  transaction, `document.issued` the evidence. The documents pack v3 makes it an
  `issued_document` (name, hash, media type, size, writer, parameters) with its list
  view. `adapter.Issue` renders and hands over in one step: `rhea issue WRITER k=v…`,
  `POST /api/files/{writer}/issue`. A later correction is a new file, a new fact; the
  recorded writer and parameters let anyone render again and compare.
- **JPK_V7M(3) in `markets/pl`** (`jpk.go`, provided as `jpk_v7m`): the header (form
  code and schema version fixed by the XSD, purpose 1, the tax office, year, month),
  the taxpayer (a legal person: NIP, full name, e-mail; a natural person: NIP, given
  and family name, birth date, e-mail — read from the company of kind self and its
  owner person), the declaration from the month's `vat_settlement` in whole złoty
  (P_10 exempt, P_13 zero, P_15/16 5%, P_17/18 8%, P_19/20 23%, P_37 bases, P_38
  output VAT, P_39 carried in, P_42/43 purchases, P_48 input VAT, P_51 to pay, P_53
  and P_62 the excess forward; empty optional positions left out, as the brochure
  says), the records from the month's invoices with their grosze (each sales invoice
  with the buyer, the number, the issue date, its KSeF number from the accepted
  submission or the mark BFK when it was issued outside KSeF, the K_ amounts in the
  rate's column; each purchase with the supplier, the number, the date, its KSeF
  number when the KSeF reader carried one or BFK, K_42/K_43), and the control sums.
  Without a settlement for the month the file has records only — it says what the
  log can derive. The writer refuses a company that has not said its tax office and
  e-mail, and a rate without a column. The one arithmetic it does is the form's own:
  whole złoty by the tax ordinance's rounding (art. 63 §1).
- **The company says who it is** (pack v17): `company` gains `email`, `tax_office`,
  `legal_form` (natural_person | legal_entity) and `owner` (a person), all amendable
  with `address` and `iban`; the interview fills them, a correction amends them. A
  natural person's names and birth date wait for their door on the people pack —
  fields nobody can fill are not shipped.
- **A purchase invoice from KSeF keeps its number** (pack v17): `purchase_invoice`
  gains `ksef_ref`; the KSeF reader already carried the reference in its event, and
  the booking rule splits in two — `pl-book-purchase-invoice-ksef` where `$.ksef_ref`
  exists, `pl-book-purchase-invoice` where it is `absent`. **`absent` joins the
  condition ops** (the mirror of `exists`): a rule may state that a field is not
  there; without it two rules could not divide one event type by the presence of a
  field, and a mapped path that is missing fails the rule.
- **Validated against the Ministry's schema, kept as golden files.** The scratch
  validator (lxml, the national XSDs resolved locally) says VALID for the JPK the
  story renders and for the FA(3) sample; both are committed under
  `markets/pl/testdata` and the tests fail on any byte of drift — a change is a
  change to re-validate. The validator itself stays outside the suite (no libxml in
  the build); when a Go validator is worth its weight it joins. The validator found
  what reading the schema by eye missed: the taxpayer's NIP and name are in the JPK
  namespace, not the shared types' — the same lesson as the buyer block this morning.
- **The story** closes September, renders the file by name, checks the positions
  (P_51 = 115, the records with the KSeF number and the BFK mark), hands it over,
  finds the issued_document with the writer and the parameters, and reads the bytes
  back under their hash. **Owed:** the live shells on v13 (now four pack versions
  behind: a pack-upgrade mechanism is due before the private beta, since the loader
  leaves known ids alone by design); the natural person's door; the accountant's
  hand-over as a person (the issued document names no recipient yet).


## 2026-10-15 — the two gaps closed: pack upgrades behind the gate, the natural person's door

- **A pack upgrades its own definitions.** The loader kept a known rule id or activity
  name untouched by design ("its versions are not a pack's to touch"), which protected
  a company's local evolution and also froze every installation at the pack version it
  was born with — the two live shells were four versions behind. The distinction the
  loader lacked was authorship, and the store already had it: every version the pack
  drafted carries `created_by: pack:<name>`, and approval's active row keeps it. Now a
  known definition whose latest version is the pack's own and whose content differs is
  drafted as the next version into the upgrade bundle (`pack-<name>-v<k>`, "Upgrade to
  …"), behind the same gate as an install; an unchanged one is already present; one a
  person or the agent evolved since is reported as a conflict and left alone, for a
  person to resolve by hand (`Summary.Upgraded`, `Summary.Conflicts`; `rhea pack`
  prints both). Object types and views were always keyed by version and upgraded by
  the install path. Approving the bundle supersedes the old versions as any rule
  approval does. `TestPackUpgrade`: install, upgrade two of three, the shipped version
  runs until approval, reload is a no-op, a local evolution blocks the pack's change
  with its name on the conflict.
- **The natural person's door.** People pack v2: `person` v2 gains given name, family
  name and birth date (amendable) and the door `identify_person` with its rule. Poland
  pack v18: the door `identify_company` — e-mail, tax office code, legal form, and for
  a natural person the owner as a person — with two rules split on `$.owner` present
  or `absent`. The JPK writer renders the natural person's block from them; the test
  walks both doors and reads the block back. The interview will fill these the way it
  fills the rest.
- **The live shells.** Both databases load the four current packs through the upgrade
  path (people v2, periods v1, documents v3, Poland v18); the bundles wait at the gate
  and are approved from the CLI as the person who runs them.


## 2026-10-15 — backup and restore: the log and the definitions are the whole system

- **What a backup is.** The append-only tables, every row: object types and views in
  every version with their statuses, rules and activities likewise, bundles with their
  warrants and questions, files under their hashes, and the event log with its ids,
  causes, rule and door provenance, actors, companies and times. Not the object table:
  it is the projection replay rebuilds (invariants 4 and 7), so a restore ends with a
  replay. `Store.Backup` writes JSON lines — a header, then one line per row as
  Postgres itself spells the row (`row_to_json`) — so a backup reads without Rhea.
- **What a restore is.** Into an empty system of record (the schema in place, no event
  and no definition — `InitSchema`, the schema without the seeded verbs, since the
  verbs arrive with the backup): one transaction, every row inserted with its id and
  time (`json_populate_record`, `OVERRIDING SYSTEM VALUE` on the event identity), the
  event sequence moved past the last id so the log continues, then the replay. A
  restore refuses a log that is not empty and a line naming a table it does not take.
  `rhea backup [FILE]`, `rhea restore FILE`.
- **The test** (`TestBackupRestore`) builds a world — two type versions, an active rule
  and a draft at the gate in a bundle with a question, three events fired, a file
  through the door — backs it up, restores it into a bare throwaway database, replays,
  and finds: identical objects with identical provenance, every row of every table the
  same byte for byte (a backup of the restore equals the backup, header aside), the
  file under its hash, the doors present, the draft still waiting, the log continuing
  at the next id and the restored world still booking. `TestRestoreLiveBackup`
  (gated by `RHEA_LIVE_BACKUP`) restores a real backup; the Poland demo database was
  backed up and restored into a throwaway the same night.
- **Open-core, decided with KK the same night** (DIRECTION, the plan note): the kernel
  and the base Poland pack with its statutory channels open on the day the street
  opens; the face, the industry packs and the hosted network stay proprietary; the
  boundary is the JSON API, so the kernel's licence never reaches the face.

## 2026-10-16 — a rule's own warrant

- **What.** `core.Rule` carries an optional `Warrant` — the same type a bundle carries
  (basis, citations, scope; support counted, never proposed). A bundle's warrant speaks
  for the whole proposal; a rule's speaks for itself and travels with every version,
  because approval and the pairing verb append versions by copying the rule. Stored in
  a new `rule.warrant` JSONB column (additive `ALTER TABLE … ADD COLUMN IF NOT EXISTS`;
  `rhea init` applies it — both live databases migrated the same day). Backup and
  restore carry it with no change: they spell rows as Postgres does.
- **Why on the rule, not by splitting the pack into bundles.** The pack had one warrant
  for 29 rules; the review of the Poland pack (DIRECTION, 2026-10-16) needs a source per
  reading, and "why?" on a number walks fact → rule, so the rule is where the law must
  be found. Splitting the pack into a bundle per decision would have meant an approval
  per decision for one install.
- **The pack format.** A pack rule may name `warrant`; the loader validates it as
  proposed (a statute or standard needs its citation, no support figures) and a changed
  warrant is a changed rule — the upgrade drafts it behind the gate like any change.
- **Where it shows.** The provenance walk names the explaining rule's source under the
  rule (`rule_warrant`, `Store.RuleNotes` replacing `RuleDescriptions`); the rules list
  shows it under the description; the bundle card uses the same one-line rendering.
- **The test** (`TestRuleWarrant`): a toy pack's rule cites an article; the approved
  version keeps it; a fact it explains walks to it; a corrected citation is drafted as
  the upgrade; a statute without an article and a pack claiming support never land.

## 2026-10-16 — a condition reads through a link

- **What.** A match condition may be a formula: `{"formula": "=ref(company, vat_id,
  $.seller_nip).books = \"kpir\""}` — any formula-language expression that yields true or
  false, reading the payload and, through `ref()` and links, the objects the event
  names. Path conditions are unchanged; a condition is one or the other. Earned by the
  Poland pack review (DIRECTION, 2026-10-16): one pack serves a sole trader and a
  company by what the company says it is, and a rule must ask the company record, not
  the event. The exit the destiny list recorded for ref-reads in match conditions.
- **Where it reads.** The executor evaluates conditions in the environment the rule's
  effect gets (`x.env` over the chain-visible lookup and getter): the state at this
  point in the log, so replay decides every condition the same way (invariant 4 — the
  test replays and finds each sale in the same book). A field the record lacks is an
  error, as everywhere in the formula language, unless the formula names its stand-in
  with `coalesce` (below; corrected the same day — the first note said such a field
  compared as null, which the language never did); a link that does not resolve is an
  error and the event waits in the worklist — unroutable beats misrouted, as for
  markets.
- **The gate.** `Validate` admits a formula condition only if it parses and its
  static result can be a truth value; `Check` re-runs it against the catalog, so a
  read through a link must name a field the record declares. Contexts without state
  (the queue filter, the matching grid's trigger preview) evaluate formula conditions
  on the payload alone (`EvalCondition`); the executor uses `Condition.Holds`.
- **Known limit.** A condition reads the record as it stands at its point in the log,
  not as of the event's business date. A late document from before a change of the
  company's form, processed after it, reads the new form. Rare (a year boundary and a
  late document) and correctable by a person; the exit is an as-of read, if the
  business generator shows it matters.
- **The agent** knows the new condition from its grammar.
- **The test** (`TestConditionReadsThroughALink`): one company switches from full books
  to KPiR between two sales; each sale lands in its book by the rule the company's
  record selects, a sale from an unregistered seller waits, replay reproduces it, and
  the gate refuses a condition that is not true-or-false, one that is both formula and
  path, one without its "=", and one reading a field the company does not declare.

## 2026-10-16 — coalesce: a missing value is an error unless the formula names its stand-in

- **What.** `coalesce(a, b)` joins the formula language: `a`, unless it is not there,
  then `b`. Inside its first argument, path and field reads are optional — a key the
  payload or the record lacks reads as null instead of failing (`opPath`/`opField`
  with b = 1, set lexically by the compiler); everywhere else a missing value stays an
  error. The same stance as rounding: precision is lost only where declared, a value
  is defaulted only where the formula says what stands in.
- **Why.** Invoices state some facts only when they differ from another: FA(3) P_6,
  the sale date, appears only when it is not the issue date, so the tax point is
  `coalesce($.sale_date, $.issue_date)`. And a company that never declared its form has
  the law's default: `coalesce(ref(company, vat_id, $.seller_nip).vat_status,
  "active")`. Splitting each such rule on `exists`/`absent` would double the pack.
- **What stays strict.** A `ref()` that resolves to nothing is still an error inside
  coalesce — the record must exist; only its field may be silent. Arithmetic and
  ordering on a null still fail with the read's origin named.
- **Tests:** `TestFormulaCoalesce` (stated and silent sale dates, the later of tax point
  and receipt, arity, kinds); `TestConditionReadsThroughALink` lists a company that
  never declared its books and finds its sale in the ledger by the stand-in.

## 2026-10-16 — Poland pack v19: the dates that decide the VAT month, and the company says what it is

The first pack version out of the review (packs/pl/REVIEW.md, causes 1 and 2).

- **The sale's month is the sale's.** `sales_invoice` v6 carries `sale_date` (FA(3) P_6;
  the issue date when the invoice states none) and `vat_period = month(sale date)` —
  the tax point of art. 19a ust. 1, where the issue month had stood. A September sale
  invoiced on 3 October settles in September. The registers group by `vat_period`.
  The FA(3) writer writes P_6 when the sale was on another day; the JPK writes
  DataSprzedazy. Periodic services, utilities and rent, and advances keep their own
  tax points (art. 19a ust. 3, 5, 7, 8) — later versions, named in the rule's warrant.
- **A purchase is deductible from the later of the supplier's tax point and its
  receipt** (art. 86 ust. 10 and 10b pkt 1). `purchase_invoice` v5 carries `sale_date`
  and `received_date`; `vat_period = month(max(...))`. A KSeF invoice is received the day
  KSeF numbered it (art. 106na ust. 3): the reader takes the date inside the KSeF
  number; any other file was received the day it reached the store, which the
  document door now stamps into its payload (`date`, as verb facts already carried)
  and the documents pack v4 keeps as `received_document.received_on`; the PDF reader
  passes it on too. The JPK writes DataWplywu where receipt and issue differ. The
  later deduction art. 86 ust. 11 allows is a person's choice, an amendment with a
  reason — not modelled yet; a cash-method supplier (P_16) neither.
- **The company says what it is** through `declare_company_form`: VAT status (active or
  exempt) with the provision of an exemption, books (full, KPiR, ryczałt), income tax
  (CIT, skala, liniowy, ryczałt), and from when. Two rules, split on the provision's
  presence; both read a formula condition that refuses what the law does not allow —
  ryczałt tax without the ryczałt record, CIT without full books, an exemption without
  its provision — so an impossible declaration matches nothing and waits. Legal form
  against tax form is not checked yet (the payload names the company by id, and the
  rule cannot read its legal form without a ref-read by id).
- **What reads it now.** Purchase VAT is booked as deductible only for a buyer whose
  status is active (`coalesce(..., "active")`: undeclared is the law's default for a
  registered company); an exempt buyer's purchase books gross (PIT art. 23 ust. 1 pkt
  43 lit. a). The VAT settlement fires only when no company of kind self is exempt (art.
  99 ust. 7). The JPK writer refuses an exempt company. The FA(3) writer writes P_19 =
  1 with P_19A, the declared provision — "zwolnienie z VAT" in P_19C was no legal basis
  — and refuses an exempt invoice whose seller declared none. Books and income tax are
  declared now and read when the sole trader's records arrive.
- **Invoices without VAT book at all.** A posting line must be positive, so a zero-VAT
  invoice (exempt, 0%, an exempt seller) failed its whole event — an insurance invoice
  could not be booked. The posting rules split on `$.vat > 0`: with VAT, the three
  lines; without, two (201/700, and 300/202 gross).
- **The JPK refuses a month without its declaration** (JPK-11): an original JPK_V7M
  is the records and the declaration together, and the declaration is the close's.
- **The invoice dates are optional on the shared types.** The German and intercompany
  packs write the same `sales_invoice` and `purchase_invoice`; a required sale date
  broke them. The Polish rules always fill both dates.
- **Validated:** an exempt invoice with P_19A and a sale-dated invoice with P_6 against
  the FA(3) XSD; the September and October JPK files with DataSprzedazy, DataWplywu,
  NrKSeF and BFK against the JPK_V7M(3) XSD — the Ministry's schemas, lxml, 2026-10-16.
  The golden files are unchanged.
- **Tests:** `TestTheDatesThatDecideTheVATMonth` (a September sale invoiced in October,
  a KSeF purchase numbered in October, a mailed one received in October — through the
  reader, the close of both months and both JPK files); `TestAnExemptSoleTrader` (two
  impossible declarations wait, the exempt sole trader's sale and purchase book without
  VAT, its invoice names art. 113 ust. 1, its close settles nothing, its JPK is refused).

## 2026-10-16 — a late document is never refused: it books in the first open month

- **What.** A closed month still takes no posting — but a document dated in it is no
  longer refused into the worklist. When a chain meets a period lock, the executor
  re-dates the chain to the first day of the first month after it, in the same year,
  that the locked book has open, and books it there (`firstOpenMonth`, the same move
  backfill already made for a chain bound for a closed period). The raw fact keeps its
  own date in the log; its derived events and entries carry the booking date. Review
  FB-16: the accounting act requires every event in the books (art. 20 ust. 1, art. 24
  ust. 3), so "a closed month is a closed month, even for what was late" (the M3 story)
  is superseded — the closed month stays closed, the document is booked.
- **Not across the year.** If every later month of the year is closed, the document
  waits with the reason: a year closes only with its statement, and a December
  document must not slip into January by itself. December's VAT close and the year's
  close are still one lock in Rhea — separating the VAT month from the book's year is
  owed (the year end).
- **A chain moves whole.** An event whose chain posts in two books, one closed, books
  in the first month the closed book has open — both halves together, as an event is
  never half-explained (TestParallelBooks).
- **What a late invoice owes.** Its VAT month is still the sale's: a late September sale
  belongs to September's return, which was settled at the close. The JPK writer
  therefore refuses to render an original whose records no longer agree with the
  month's settlement (sales VAT against `output_vat`, purchase VAT against
  `purchase_vat`) and names what is owed: a correction of the return (CelZlozenia 2),
  which Rhea does not write yet. Before, the same invoice was simply lost to the books.
- **Tests:** TestDoubleEntry, TestCascade, TestParallelBooks and TestCohabitation now
  find the late document in the next open month and nothing in the closed one; the
  Poland story files September's JPK, then books a late September invoice in October
  and finds the original refused for a correction.

## 2026-10-16 — Poland pack v20: purchases land in costs, penny differences in other operating items

The review's cause 4 (packs/pl/REVIEW.md, FB-06, FB-14, MON-19); the late-document half of
cause 5 shipped in the kernel the same day.

- **A purchase is a cost.** Every purchase used to book Wn 300 and only the warehouse
  pack's PZ ever cleared 300, so services, rent and fuel never reached the profit and
  loss account. Now the debit's account is a formula: a purchase naming a delivery note
  (`for_document`) still waits on 300 for its PZ; any other goes to the supplier's
  declared cost account, or 403 (services) when none is declared. 403 is the default
  because a small firm's typical unexplained invoice is a service, and a wrong kind of
  cost leaves profit, the CIT advance and the CIT-8 right — staying on 300 left them
  all wrong. `set_cost_account` (company v6, `cost_account`) declares a supplier's
  account; an account outside the chart matches nothing and waits silently (an `any`
  over the chart, not a `ref` — a failing ref would resurface on every pass).
- **Penny differences.** The incoming policy's tolerance account is 761 (other
  operating costs) and the difference record picks the account by direction: a shortfall
  761, an overpayment 760 (other operating income) — not 756, which is the financial
  group's exchange account (art. 3 ust. 1 pkt 32 lit. d against art. 30 ust. 4). A
  reversal flips the direction but keeps the policy's account, so the record rule asks
  whether the difference undoes one (`coalesce($.doc.reversal_of, 0) > 0`) and sends it
  back through the account it came in on; a difference from before v20 keeps 756 both
  ways. The posting rules are unchanged: they post to the record's account. The tax
  side of a grosz (MON-19) is the review's one genuinely open question and stays in the
  warrant, undecided.
- **The chart** gains 750, 751 and 761. 756 now holds only the conversion rounding plug
  of the convert clause (one account by the sub-language's design); realised exchange
  differences arrive with the currency version.
- **Test:** `TestWherePurchasesAndPenniesLand` — fuel on 402 by the declared account, a
  service on 403, goods with a WZ on 300, an account outside the chart waiting; a grosz
  short on 761, two grosze over on 760, the short match undone and its grosz back out of
  761. The M3 story's service purchase moved from 300 to 403.

## 2026-10-16 — a door's ref input reads as a link

- **What.** A raw fact that came through a declared door has the door's shape: when a
  rule fires on it, the payload's root is typed by a synthetic type built from the
  door's inputs (`door:<name>`, the exact version the fact recorded), so a `ref<T>`
  input reads as a link — `$.line.amount`, `coalesce($.company.legal_form, "")` — in
  conditions and effects alike. The door already vouches that the object exists and is
  of the declared type, so this is the "direct ids rejected by design" stance (E4)
  kept: only ids the door vouched are followed, never an id a payload merely carries.
  Facts rules derive, facts with no door, and the open door's passthrough (a `json`
  input) read as before.
- **Why now.** The bank work needs a person's word on a bank line ("this is the owner's
  deposit") to book the line's own amount, and the company form needs the company's
  legal form; both name an object through a door.
- **How.** The firing's context carries the door (`withDoor`), and the formula
  environment types the root from it (`doorShape`, reading `GetActivityVersion`); a copy
  through a link is a copy — no Calc, the walk explains it by rule, fact and object.
- **Test:** `TestADoorsRefInputReadsAsALink` — a person says a line is the owner's
  money; the rule reads the line's direction in its condition and its title and amount
  in its effect; a debit matches nothing; replay reproduces it.

## 2026-10-16 — Poland pack v21: a bank line says what it is before it is booked

The review's cause 3 (packs/pl/REVIEW.md, MON-07, MON-08, MON-18, ZUS-08, OWN-02/03/05/07).

- **Recognised on arrival.** `pl-bank-operation` gives each line its kind by one formula,
  in order: a fee (`NCHG`, `NCOM`) or interest (`NINT`) by the transaction code the MT940
  reader now keeps from :61: (`code`; an additive field, so the reader's code version
  stays 1, as the KSeF reader's additive fields did); a transfer between the company's
  own accounts (the counterparty is a company of kind self by its IBAN); a payment to
  the tax micro-account, the ZUS contribution account or from/to the owner's private
  account — each as the company declared it through `declare_company_accounts` (company
  v7; an omitted account keeps its value, read through the door's link). Everything
  else is a `transfer`: a customer's or supplier's payment, booked and matched as before
  — so the personas, whose lines are all trade payments, see no change.
- **Booked where it belongs.** One rule posts every explained kind on its own side of
  130: tax 225, ZUS 229, own transfer 149 (cash in transit), owner 249 (settlements to be
  explained — a loan, capital or a dividend for a company; a sole trader's own money),
  bank fee 403, interest received 750, paid 751. An explained line is settled at once
  (open 0) and never enters the matching. Accounts 149, 229, 249 join the chart.
- **A person's word.** A transfer nothing settled can be classified through
  `classify_bank_operation` (owner, tax, ZUS, own, fee, interest): the line moves off the
  trade settlements by its open amount (a credit Wn 201 / Ma the kind's account, a debit
  Wn the account / Ma 202) and stops waiting; a second word on an explained line matches
  nothing. Reading the line's amount and direction is the door link shipped the same day.
- **Not yet.** Tax and ZUS payments are not matched to the liabilities they pay (a VAT
  settlement does not post its closing entry yet, ZUS declarations are not objects); the
  VAT account of split payment is not modelled; card lines keep their old path.
- **The company form against the legal form.** Both declaration rules now also require a
  legal person to declare CIT on full books and a natural person not to declare CIT,
  reading `$.company.legal_form` through the door's link (undeclared: no constraint).
- **A debt found.** A literal in a money field is stored as written — `"0"` became the text
  "0", not 0 minor units — because literals bypass coercion except for enums. The rules
  here use `"=0"`. Paid the same day: a literal in a money, int or decimal field now
  runs through the formula engine and is coerced by its field like any computed value
  (`TestLiteralNumbersAreCoerced`); a literal that is not a number is refused, not
  stored. No pack wrote a number as a literal, so nothing booked differently.
- **Test:** `TestABankLineSaysWhatItIs` — seven lines: VAT to the micro-account, ZUS,
  the owner's deposit, a charge, interest, a transfer to the own account, a stranger's
  700; each on its account, only the stranger's waiting; a person says it was the
  owner's and it moves from 201 to 249; a second word matches nothing; a legal person
  declaring the scale waits, CIT on full books is taken.

## 2026-10-16 — fx(): a rate inside a formula

- `fx(from, to, d)` joins the formula language: the rate converting `from` into `to` for a
  document dated `d`, read from the same `fx_rate` master data postings' convert and the
  matching engine read — by code `from/to/d`, the inverse pair inverted, exactly 1 when
  the currencies are one, an error when no rate was published (the event waits:
  unexplainable beats guessed). The rate read is an input of the formula's Calc.
- **Why.** VAT on a foreign-currency invoice is owed in złoty (VAT art. 31a, art. 106e
  ust. 11): the invoice itself must carry its VAT in PLN, so the rule that books it must
  compute `round($.vat * fx($.currency, "PLN", $.sale_date), 2, half_up)` — a value on
  the document, not only a conversion inside a posting.
- **What the date means is the market's.** The kernel reads the code it is given; the
  Poland pack's rates channel decides that the rate published under date `d` is the one
  a document dated `d` converts at (the NBP average of the last working day before it).
- **Test:** `TestFormulaFX`.

## 2026-10-16 — the NBP channel: a rate for every day a document can be dated

- **What.** `markets/pl/nbp.go`, adapter `nbp` (code version 1), declared by its own small
  pack `packs/pl/nbp.json` (`pl-nbp` v1: the channel and `pl-register-fx-rate`). A pass
  reads table A (the NBP's average rates) from the public Web API and publishes, for every
  calendar day of its window and every currency it follows (EUR, USD, GBP, CHF by
  default), `fx.rate.published` with code `CUR/PLN/day` and the rate of the last table
  published *before* that day — the rate Polish law converts a document of that day at
  (VAT art. 31a ust. 1, UoR art. 30 ust. 2 pkt 2, PIT art. 11a, CIT art. 12 ust. 2).
  Weekends and holidays fall back to the last table by construction; the table's
  number and date ride on the event. The window defaults to the last 31 days through
  today; `RHEA_NBP_FROM` reaches further back.
- **Why the day, not the table.** The kernel's convert, the matching engine and `fx()` all
  read a rate by `from/to/date`; making the date the document's date puts "the working
  day before" in one place — the market's channel — and keeps every rule simple.
- **Exact.** Rates decode as `json.Number` and stay the bank's decimal text; no floats.
  The API refuses Go's default user agent ("forbidden by administrative rules"), so the
  client names itself.
- **Tests.** `TestNBPRatesForEveryDay` (a weekend on Friday's table, Monday on Friday's,
  Tuesday on Monday's; a second pass adds nothing); `TestNBPLive` (gated by
  `RHEA_LIVE_NBP`) read the real tables for 21–30 September 2026 on the day it shipped.

## 2026-10-16 — Poland pack v22: currencies

The review's cause 6 (packs/pl/REVIEW.md, VAT-18, KSEF-10, MON-16, MON-17, FB-13).

- **The invoice carries złoty.** `sales_invoice` v8 and `purchase_invoice` v7 gain `net_pln`
  and `vat_pln`: `round($.net * fx($.currency, "PLN", sale date), 2, half_up)` — the NBP
  average of the last working day before the sale (the `pl-nbp` channel publishes that
  rate under the sale's date). For złoty invoices the rate is exactly 1.
- **One rate for the books and the VAT.** The invoice entries convert at the same date
  (the sale date, where the issue date stood), so 222 and 221 hold exactly the invoice's
  VAT in złoty and the VAT account cannot drift from the return. Bank lines — trade
  payments, explained kinds, a person's reclassification — convert at the rate of their
  own day.
- **Realised exchange differences.** Both payment matchings now declare `convert`
  (functional PLN, settled at the payment's date). The difference record picks 750 when
  the rate rose on what a customer owes or fell on what we owe, 751 otherwise, and a
  reversal goes back through the account it came in on; `pl-post-fx-difference` squares
  the receivable or the payable against it. The difference record now covers outgoing
  payments too.
- **The settlement, the registers, the JPK and the invoice in złoty.** Every sum of the VAT
  settlement reads `coalesce(i.vat_pln, i.vat)` (invoices from before v22 are złoty
  invoices); the registers likewise; the JPK writer reads the złoty amounts; the FA(3)
  writer states P_14_xW for an invoice in another currency (art. 106e ust. 11) and refuses
  one without it.
- **A known limit.** The matching engine dates an invoice by its issue date (it also orders
  by it), while the books convert at the sale date. A foreign-currency invoice issued on
  a later day than its sale leaves the rate gap of those days on the receivable when it is
  paid. Rare; an invoice side dated by its rate date is the fix when the engine can read
  a missing date as none.
- **Validated:** the euro invoice's FA(3) with P_14_1W against the Ministry's XSD.
- **Test:** `TestAnInvoiceInEuro` — NBP rates through the channel, a euro sale and purchase
  at Monday's average, the customer paying a day later (gain 6.15 zł on 750), the supplier
  paid on 1 October (loss 1.84 zł on 751), both settlements square to zero in złoty,
  September settled and filed in złoty, the invoice sent with its VAT in złoty, and the
  customer's match undone with its gain back out of 750.

## 2026-10-16 — the face reads only the API: objects, approvals, the host's words, and "why?" on a cell

- **Why.** The face is built apart (DIRECTION, "The face"): its first snapshot had to read
  three things from the database — the objects with their provenance, who approved each
  rule, what the host said — and it composed a VAT register cell from the invoices by
  itself. Step 1 of making the prototype the product: the API answers all four, and the
  face's capture now reads nothing else.
- **`GET /api/objects?type=T` and `GET /api/objects/{id}`** — the objects as the kernel
  holds them (money in minor units, refs as ids) with `source_event_id`, `rule_id`,
  `rule_version`. Scoped by the company like every read. No paging yet: the face asks
  per type; paging comes with the live face (step 2).
- **`/api/rules` carries each active version's `approval`** — `approved_by`,
  `approved_at` (the business date), `bundle` when it came as a bundle member, `reason`
  for a correction, `event_id`. Read from `rule.approved` (names the version it made) and
  `bundle.approved` (names the drafts it activated; each became the next version).
  Definitions are installation-wide, so approvals are read across the whole log.
- **`GET /api/said?view=V&n=N`** — what the host said, newest first, as the log keeps it.
  Reading the narrative view still speaks anew; a face showing this morning's words reads
  them here instead of asking her to speak again.
- **A view says what its rows are made of.** An analysis view's SQL may end with a column
  named `_of` (`core.RowsOfColumn`): the comma-separated ids of the objects the row
  aggregates (`string_agg(object_id, ',')`). The kernel takes it out of the columns and
  serves it beside the rows as `row_objects`; the host's facts never carry it (her words
  are checked against the facts' numbers, and an id is not one). A column may declare
  `sum_of` — the fields it sums, first present wins, as the SQL coalesces them. Chosen over
  a structured analysis notion: views stay SQL, the author declares lineage in the same
  query that aggregates, and the kernel, not the face, checks it.
- **`GET /api/explain?view=V&row=N&col=F`** — "why?" on a cell: the row's objects in log
  order with their labels, each one's part and which field gave it, the total and `sums`:
  whether the parts make the cell. A view without `_of` is refused, never guessed.
- **Poland pack v23** — both VAT registers (sprzedaży v6, zakupu v4) declare `_of` and
  `sum_of` (net: net_pln|net, vat: vat_pln|vat, gross). Same rows, same columns.
- **Test:** `TestFaceReads` — two invoices, the objects with provenance, a pack rule
  approved through its bundle and a rule through its own door, the register row naming
  both invoices, "why?" on September's VAT: 230.00 + 115.00 = 345.00 and the check holds,
  a column that sums nothing lists without parts, a view without lineage refused, the
  host's register facts free of ids, her words newest first.

## 2026-10-16 — the face live: an object names its root; analysis reads take turns

- **`root_event_id` on `/api/objects`.** Each object carries the raw event its chain starts
  from (a recursive walk of `cause_event_id`, `queries/object_roots.sql`). A face finds the
  document a posting books — the invoice, the bank line — by the root they share, without
  reading every chain; chains are fetched only when a sheet opens.
- **One analysis read at a time.** Each analysis read rebuilds the DuckDB projection and
  queries it; two at once collided on the rebuild ("Catalog write-write conflict") or could
  read a table mid-rebuild. Found by the live face, which asks for its reports together. The
  shell now holds one lock around rebuild-and-query. Performance stays a non-goal; this is
  correctness. `TestFaceReads` asks for eight reports at once.

## 2026-10-17 — step 3: the pack speaks the face's words and the kernel computes its figures

- **Why.** The face carried a table of Polish words, a table of which shape each type takes,
  the VAT declaration's layout, the account numbers that play the bank and the receivables, and
  computed the morning's figures, profit and ageing by itself (DIRECTION, the plan from the
  prototype to the product, step 3). A face built apart must name nothing and compute nothing.
- **A `words` notion.** A view whose spec is a market's dictionary: types (label, plural, the
  shape a face draws — paper, card, ledger, statement), fields (`type.field` for one type's),
  enum values, events, doors, the chart's account groups. The kernel names nothing in any
  language; `WordsSpec.Check` keeps a face able to use it. The Poland pack's `slownik` names all
  28 types, every field and every value — `TestThePackSpeaks` fails on anything left unnamed.
- **A `statement` notion.** One object's own figures laid out as sections of rows, each with its
  operation (+, −, =) and a note; served with `?object_id=` as the kernel encodes the figures.
  It computes nothing: the VAT declaration rounds to whole złoty and must not be made to add up
  to the grosze above it. `deklaracja-vat` replaces the face's layout.
- **`{{today}}` in analysis SQL** becomes the kernel clock's day as a DATE, so a report that
  depends on the day reads the kernel's day, not the machine's.
- **Three reports the face used to compute**, each row saying what it is made of: `rano` (the
  morning's figures — bank, receivables, payables, the last VAT — and what was done),
  `wynik` (profit since January by account type), `naleznosci-wiek` (receivables by age on the
  kernel's day). The account numbers (130) and the book (pl-stat) live in the pack's SQL now.
- **Schema.** The notion CHECK learns `words` and `statement` through its idempotent migration;
  `rhea init` brings an existing database along (done on rhea_face and white01).
- **Poland pack v24.** Five views, nothing else; an installed v23 upgrades in place.

## 2026-10-17 — the result month by month (Poland pack v25)

- **Why.** KK asked the face on his phone „Możesz pokazać wynik za sierpień tylko?”; she said
  she shows the result only since January — true: `wynik` sums every booking, and nothing in the
  books cut it by month.
- **`wynik-miesiace`.** Revenue, costs and the result for each month of the books, in `wynik`'s
  own rows (op, line, amount) with the month in front. Each row is made of its month's bookings
  (posting ids, not accounts: an account's balance is the year's, wrong under a month's figure).
  No `sum_of`: a booking's amount is unsigned and its side decides, so a plain sum would not check.
- **The month is the booking's.** A late document booked into the first open month counts there,
  where the closed months left it; the document keeps its own date in the log. The months add up
  to `wynik` — the test checks both.
- **Poland pack v25.** One view, nothing else; an installed v24 upgrades through the gate.

## 2026-10-17 — a figure she can be taught: the tally (step 1)

- **Why.** An analysis view was SQL a pack's author writes; the AI author may not write SQL that
  runs against the books (invariant 3), so nobody could teach her a figure in words (DIRECTION,
  "Teaching her a figure").
- **`analysis` with a `tally` instead of `sql`.** No new notion: the same view, the same readers.
  `TallySpec` names the type added up, the amount (a money field, or `debit-credit` /
  `credit-debit` on a type with a side and an amount — the kernel's postings), conditions
  (`eq ne in exists absent`, on a field or one step through a link), at most two groupings (a
  field or the month of a date), and lines (+, −, = for the lines above). The kernel writes the
  query: every name must be one the catalog holds, every value becomes a quoted literal.
- **Compiled at read time** against the catalog as it stands, in one place (`analysisSpec`) for
  the view, "why?" and the host's facts; a view carrying both SQL and a tally is refused.
- **Columns are the kernel's** (the groupings, then op, line, amount, and the row lineage),
  labelled by the tally (`label`, `line_label`, `amount_label`) — the kernel names nothing in
  any language. A plain sum of one money field declares `sum_of`, so "why?" checks it.
- **Proof.** The monthly result written as a tally gives the pack's own SQL rows and bookings;
  the pack keeps its SQL (no second approval of the same skill on white01).

## 2026-10-17 — the author drafts figures (step 2)

- **The grammar** gains FIGURES: an `analysis` view with a `tally`, one worked example (the
  result by month, English, the book named `main`), the ops, the link step, the months, the
  lines, and "never SQL; a figure needs no rule".
- **Validation**: an analysis view the author drafts must carry a tally and no SQL, and the tally
  must check against the catalog with the bundle's own types.
- **The books in the ask**: the author saw the active rules only by what they match, so it could
  not know a figure over bookings must keep to `pl-stat`; the ask now names the books the active
  rules post to (literal books only).
- **Live**, on Helios with the kernel's author model: „Pokaż sprzedaż netto według klientów,
  miesiąc po miesiącu” drafted a valid tally in 8 s (sales invoices, `net_pln`, by month and
  `buyer.name`, Polish labels); approved, 18 rows, "why?" adds up. A fallback amount
  (`net_pln` else `net`, as the VAT registers coalesce) is not built: the Poland pack fills
  `net_pln` on every invoice.

## 2026-10-17 — what she was never taught reaches the house (step 3)

- **KK's principle.** Rhea learns from every client what she was never taught — people in the
  loop: the Rhea team reads it and helps her learn (DIRECTION, "Teaching her a figure").
- **A built-in door, `note_untaught`** (`host.untaught`): the person's words (stay home), her
  general description of the missing skill, who asked, and `shared` yes/no with
  `kept_home_because` — the fact says what left. `POST /api/untaught` decides before it records;
  `GET /api/untaught` lists them newest first, so the owner sees everything she reported.
- **The standard for what leaves** (Rhea suggests standards; KK may override): the description
  only, and only when `core.KeptHome` finds no digit, no `@`, and no word (or stem of a word, so
  „Novej” is caught by Nova) of a company's or a person's name the installation holds. It errs
  on keeping a text home. Without a network, or when it cannot be reached, the fact says so.
- **The house**: an append-only `untaught` table in the network (installation, description,
  when, synthetic), read counted — one row per description regardless of case and outer spaces,
  installations and times — by `rhea network` under "never taught yet". Simulated customers are
  marked and left out, as with rule shapes.
- **Consent to report** is part of onboarding (KK, 2026-10-17): asked when a client joins; nothing to build before then.

## 2026-10-17 — approving what you can read: the waiting figure previewed (step 4, the kernel's part)

- **Found live.** Asked „Pokaż sprzedaż netto według klientów, miesiąc po miesiącu”, the author
  once drafted revenue bookings grouped by account code and labelled the code „Klient” — valid,
  and wrong. The grammar's only example was over bookings.
- **The grammar** gains a document example (sales by customer and month) and the rule of thumb:
  a figure about customers, suppliers or documents adds up the documents through their links; a
  figure about the books adds up the postings through their account. Six live drafts after it
  (sales by customer ×3, purchases by supplier, costs by account, output VAT by rate): all sound.
- **`GET /api/bundles/{id}/preview`**: a waiting bundle's analysis views run now, the first
  eight rows and the count, an error per figure rather than for the bundle — so the card shows
  rows, and „700” under „Klient” is plain before anyone approves it.
- **The correction loop was already the kernel's**: `reject_bundle` with a reason, then a draft
  `after` it reuses the question and shows the author the refused proposal and the reason. The
  test walks it for a figure.

## 2026-10-17 — the morning by role, the result first (Poland pack v26)

- **KK.** The morning focused on VAT, "a necessary evil, not the centre"; CIT and the result
  (revenue against costs, month and quarter) matter more; a CEO wants something different from
  a chief accountant; industries and companies differ (SaaS, RMK, a dominant cost).
- **A narrative names the roles it speaks to** (`for`). A person hears the narrative whose `for`
  names their role, anyone else the one that names none; `GET /api/said?person=<handle>` reads
  that person's morning. Roles are stored as values like every other (`owner`, `accountant`);
  the owner's morning names no role, so it stays `poranek` and a face reading it changes nothing.
- **The owner's morning** speaks from four reports: the result (this month so far, last month,
  since January, before tax), money (bank, receivables as a total with the oldest invoice, not a
  list; payables), the tax deadlines, what is open. **The accountant's** (`poranek-ksiegowosc`)
  leads with what is open (months to close, unexplained bank lines, tasks past due), then the
  deadlines, the result last. VAT is a deadline with an amount, not the headline. "Rano w
  liczbach" leads with the result of the month and the year.
- **Deadlines** (`kalendarz-podatkowy`): VAT and JPK_V7M on the 25th, the income tax advance on
  the 20th of the next month, moved to the next working day off Saturdays, Sundays and the
  statutory holidays (Ordynacja art. 12 § 5; Easter Monday and Corpus Christi computed, 24
  December included). Monthly advances assumed; quarterly and simplified ones come with CIT.
  The advance has no amount and says so ("księgi jeszcze jej nie liczą"): never estimated as
  result × rate, because tax-deductible and non-deductible costs, losses and the rate are not in
  the pack yet (REVIEW: CIT-02, -03, -06, -07). ZUS has no deadline row until it has an object.
- **Found live** (Sonnet 4.6, both mornings accepted by both laws): told not to repeat herself,
  the host said "September and the year unchanged" instead of the figures, so a result that
  holds day to day would drop out. The voices now say: the result in numbers every morning,
  only the wording changes. "October closed at −200" for a running month: the row now says
  "miesiąc trwa".

## 2026-10-17 — the Learning Centre, the kernel's part

- **KK's name**: the Learning Centre — where the Rhea team reads what installations were never
  taught and helps her learn it (not "team page").
- **The house only.** `GET /api/network/learning` reads the network, never an installation's
  books: each lesson (her description, counted by installations and times, first and last seen,
  which installations — the house knows its own), the team's latest note on it, and what Rhea has
  learned by question with support. `?synthetic=1` counts simulated customers too. Without a
  network it says so.
- **The team's notes** (`POST /api/network/lessons`): a lesson is `open`, `planned`, `done` or
  `not_for_us`, with a note and who said it; an append-only `lesson_note` table in the network,
  filed under the lesson's key (its description, case and outer spaces aside); the latest note
  is where it stands.
- **The page itself** is the face's (the hosted network is the proprietary side of open core);
  the kernel serves only JSON.

## 2026-10-17 — reporting, step 1: the collection and the locked session

- **The collection.** The read side gains one DuckDB schema per domain (`SchemaOf`: the domain
  as a plain name, "general" otherwise) and one typed table per object type: `id`, a column per
  declared field (money DECIMAL(18,2), int BIGINT, decimal DECIMAL(38,10), date DATE, string /
  enum / ref VARCHAR — a link is the linked id), and the provenance `_type_version _event _rule
  _rule_version _occurred_at _root_event _state` (the whole state as JSON, so an undeclared value
  is still there). The JSON `objects` table stays for the packs' SQL. Names are only the
  catalog's plain names, quoted; a value that cannot be read as its type is NULL in its column.
- **Rebuilt only when it moved.** `_rhea_meta` records what the copy was built from: the layout,
  the log's last event, the company scope, a hash of the catalog; equal means the copy is read as
  it is. `rhea replay` wipes the copy first.
- **The locked session**, for every query — the packs' and hers: DuckDB opened read-only with
  `enable_external_access=false`, extension install and autoload off, `lock_configuration=true`,
  1 GB, two threads; `json_serialize_sql` must find exactly one SELECT (so `SELECT 1; SELECT 2`
  and PRAGMA are refused even though the session could do no harm); 30 s; more than 10 000 rows
  is refused rather than cut short. Probed: files, COPY, ATTACH, INSTALL, SET, CREATE, INSERT all
  refused; CTEs, window functions and `PIVOT … FOR … IN` pass.
- **Exact text for decimals**: the driver's own rendering drops trailing zeros (0.10 → 0.1), so the
  kernel writes a decimal from its digits and scale.

## 2026-10-17 — the CIT advance, KUP and NKUP (Poland pack v27)

- **KK.** "CIT, KUP and NKUP are very important." The review's defaults (CIT-02, CIT-03, CIT-06)
  are built: the advance is a fact computed when a month closes, like the VAT settlement.
- **The kernel reads the ledger.** A year of postings is above any collection cap, so the
  formula language gains `balance(book, code, from, to)` — debits less credits of one account in
  a book over a range of dates, one aggregate the store computes, not a fold — and
  `start_of_year(d)`. A rule folds over the chart (a few dozen accounts), never over postings.
  The read is an input in the Calc; replay never recomputes, so it reproduces what was read.
- **Accounts say what they are for income tax** (`account` v2, `tax_treatment`: deductible,
  non_deductible, taxable, not_taxable; amendable through `set_tax_treatment`). Unmarked, a cost
  is a tax cost and a revenue is taxable — the CIT act's own shape (art. 15 ust. 1: costs are
  tax costs except art. 16). A cost cannot be declared a revenue: the declaration waits. New NKUP
  accounts: 409-1 entertainment (art. 16 ust. 1 pkt 28), 751-1 interest on tax arrears and
  761-1 fines (pkt 18), 761-2 donations (pkt 14); 870 income tax, a type of its own
  (`income_tax`), so "result before tax" never sees the charge.
- **9% or 19%** from facts: `declare_cit_status` per tax year (when the business started, how
  the company was formed, last year's gross sales). 9% in the year business started unless
  formed by transformation or contribution (then not in that year or the next), else for a
  small taxpayer; and only while the year's revenue stays within the in-year limit. No status
  declared: 19%. The newest declaration for a year stands. Limits per year are data
  (`cit_limit.defined`; 2026: 8 517 000 and 8 431 000 zł from podatki.gov.pl).
- **The advance**: income from January (revenue less not-taxable, less costs plus NKUP), the
  base in whole złoty, the year's tax, less the advances for earlier months of the year, never
  below zero (no refund); booked Wn 870 / Ma 225 so a payment from the tax micro-account
  clears it. The charge books on the close's date, in the first open month.
- **Not yet**: quarterly and simplified advances, losses carried forward, deductible-when-paid
  (wages, ZUS) and the 90-day unpaid-invoice adjustment (art. 15b/18f), car limits, the CIT-8,
  income outside the main source (capital gains at 19%). Prior-year sales stay on the
  declaration, not on the status object (a money field cannot be empty).
- **The morning** names the advance before VAT; told the month is running, she no longer says
  October "closed".

## 2026-10-17 — reporting, step 2: a question answered from the books now

- **`POST /api/answer`** (KK: answered at once when it only reads): the author writes one SELECT
  over the collection, a plain reading of what it counts and a visual; the kernel runs it in the
  locked session on the copy as it stands. DuckDB's refusal goes back to the author once (the
  retry sees the query and DuckDB's words); a second refusal is a 422, never a guess. `cannot`
  passes through as the author's reason.
- **What the author sees**: the collection table by table, each column with its DuckDB type, the
  field's kind, where a link points (`→ finance.company.id`) and the market's own name for the
  type, the field and an enum's values (from the active words view); the books the rules post to;
  today. The provenance columns are described once, not per table.
- **The record**: a built-in door, `host_answered` (`host.answered`): the question, the query,
  the reading, the visual, the first 50 rows as told, the row count, who asked. Nothing in the
  books changes; what she told someone can be read back.
- **"Why?"** keeps working: a query's `_of` column (the ids behind a row) is split out as
  `row_objects`, never shown. Each column comes with its DuckDB type, so a face tells money
  (DECIMAL(…,2)) from a count.
- **A visual** (`core.Visual`: kpi, table, statement, bar, line, pivot, with the columns playing
  each part and formats) is checked against the rows' columns; one that names a missing column
  falls back to the rows as a table rather than failing the answer.
- **Found**: the Poland pack puts nearly every type in the `finance` domain (invoices, companies,
  tasks, vehicles), so its collection is one schema; giving sales, purchases, bank, tax and fleet
  their own domains is a pack change, left while the pack is being edited for CIT.
- **Live on Helios** (the kernel's author, 5–15 s each): August's invoices and their net, the
  largest customer this year, the result month by month (grosz for grosz the pack's own report),
  unpaid sales invoices older than thirty days, the average invoice by quarter, each customer's
  share; "how many employees" came back as "cannot" — the books hold no workforce. Two defects,
  both now refusals the author fixes in the retry: an average came back as DOUBLE (a float, against
  invariant 6) — a float column is refused, ROUND(…, 2)::DECIMAL(18,2) asked for; and a query read
  CURRENT_DATE, the machine's day, not the books' — the clock functions are refused and the day is
  given as DATE 'YYYY-MM-DD'.

## 2026-10-17 — reporting, step 3: reports, kept and drafted

- **A `report` notion** (schema CHECK learns it through its idempotent migration): a page of 1–12
  components, each a title, a reading, a read-only query and a visual (`core.ReportSpec`). Served
  by running every component through the same guards as an answer (`runQuery`: the books' day,
  the locked session, no float shown, `_of` taken out as each row's objects, a misnamed visual
  shown as rows); a component whose query fails carries its error and the others still show.
- **Keeping an answer** (`POST /api/answers/{event}/keep`, KK: what is kept goes through the
  gate): the recorded query, reading and visual become a one-component report in a draft bundle
  — no second model call; the kept query is the one that answered. Warrant: the client, the day
  and the question.
- **The card** previews every component of a waiting report, money marked as money.
- **The author drafts reports** from words: the grammar's REPORTS section (a dashboard example:
  a KPI and a bar by customer and month) replaces the tally's FIGURES; a drafted report is checked
  for components, queries, readings and visual kinds; whether each query runs is seen on the
  preview. The bundle ask now carries the collection (with the market's names) and today. Tally
  views already approved keep working; the author no longer drafts them.

## 2026-10-18 — her morning on demand, backfill in order, a third law (Poland pack v28)

- **Why the morning did not change** (KK, after deploying v27): nothing asked her to speak — the
  phone showed her words of 16 October; Helios had never declared its tax form; advances come
  with a month's close. Helios is now declared (sp. z o.o., CIT, full books, active VAT from
  1 January 2026; status: started 1 January 2026, new — 9%), signed by KK, at his word.
- **`POST /api/speak`** asks the host to speak now: `person=<handle>` that person's morning (the
  face's "Odśwież"), without it every narrative (the 06:30 `rhea-morning.timer` the face's deploy
  installs). KK: "both". A handle no person carries hears the general morning — the face signs
  as a person not yet registered — instead of an error.
- **Backfill plans in order.** It planned every past event against the books as they stood, so
  a chain could not see the chains planned before it: February's advance would not see
  January's and charge the year again (the same held for a late VAT carry). Each chain is now
  planned against what the earlier ones leave — their objects and amendments read before the
  store's, their postings added to `balance()` — and books exactly as planned.
- **The third law: she names only what today's facts carry.** Live on Helios she named an
  invoice ("HOS/2/2026 z 14 stycznia") from her own earlier words; every number in it passed. A
  token with a letter and a digit that reads as a name (a slash or underscore, or a capital
  beside a digit) must stand in the facts as written; refused otherwise. What she said before is
  wording to avoid, never a source — her instructions now say so.
- **v28**: the previous month's row says whether it is closed, and neither voice calls a month
  closed unless its row does.

## 2026-10-18 — one correction for her words; recomputing a closed month's advance (Poland pack v29)

- **Found live, after KK closed September and pressed "Odśwież"**: refused — she named
  "HOS/2/2026" again, copying it from her own earlier words. The third law held, but a refusal
  left nothing new. Now the sentences of her earlier words that name what today's facts do not
  carry are not shown to her, and refused words are asked for once more with the reason; a
  second refusal stands.
- **September's advance was the year's tax** (28 674 zł): January to August were closed before
  the CIT rules existed, so no advance stood before it. The CIT act defines an advance against
  the advances *due* for earlier months (art. 25 ust. 1); when the books learn of those later,
  the month is recomputed. `recompute_cit_advance` (period) amends the advance — the year's tax
  stays, the advances before it are read again — and books the difference on the day (Wn 225 /
  Ma 870 when it shrinks, the reverse when it grows); recomputed again, nothing moves. The live
  sequence: ruled backfill for January–August, then recompute September.

## 2026-10-18 — prepaid costs (RMK czynne), month-end work at the month's opening (Poland pack v30)

- **KK**: some of his clients live by RMK. A purchase invoice that covers several months is
  spread through `spread_cost` (invoice, start, months; the account and amount when not the
  supplier's account and the invoice's net): its cost moves to 640 on the day, and comes back to
  the cost account in equal parts. A person decides what to spread — the materiality threshold
  is the company's policy (KSR 16 pt 12.10, FB-18); Rhea does not spread on its own yet (an
  invoice carrying its period, FA(3) OkresFa, could later).
- **When**: a part is released when the next month opens (`time.month_opened`), dated on the last
  day of the month that ended — month-end work before the close, so the close and its CIT
  advance already see it. Cumulative, like the advance: a schedule spread after its months began
  gives them back at once; the last part completes the amount; a schedule fully released is done.
  CIT follows the book cost month by month (CIT art. 15 ust. 4d spreads costs crossing the year).
- **The kernel**: an `each` may walk a keyed collection, `=objects(T, field, value)`, its elements
  arriving as a cascade carries an object ({object_type, object_id, state}, the state typed);
  `months_between(a, b)`; an entry may carry its own business date (`postings.date`), and a
  month's lock binds what comes after the fact that locked it — so a close's own entries go into
  the month it closes, while a late document or a backfill finds the month closed and goes
  forward. The CIT charge now lands in the month it closes (KK's option B).
- **white01 never ran the clock**: the 06:30 timer now opens the day before she speaks, so months
  open (and RMK parts release) on their own. Deferred revenue (RMP, the SaaS case — its CIT
  revenue date is contested) and accrued costs (RMK bierne) come later.

## 2026-10-18 — the first users: every unanswered moment reaches the Learning Centre

- **Found by KK's first users.** (1) „ile zapłaciłem ZUS w pierwszym kwartale 2026?” — the
  author wrote Postgres (`to_char`), fixed it once, then hit DuckDB's rule that ORDER BY after a
  UNION names only output columns; the face showed DuckDB's English. (2) A forecast she was never
  taught — she said so, answered the nearest thing with a query, and nothing reached the Learning
  Centre: the face dropped her "missing" when it queried.
- **The author learns the dialect**: no `to_char` (strftime, quarter), no `::money`, a total row
  by a sort column in a subquery or ROLLUP; and which side a payment is (money out of the bank, the
  debit side of the liability it settled — never the credit side). Two fixes after a refusal, not one.
- **Every moment she could not fully answer is recorded by the kernel**, of a kind: `untaught`
  (never taught; she showed the nearest thing or nothing — the face now passes her "missing" with
  the question), `cannot` (the books do not hold it), `failed` (her query would not run after
  three tries). The general words are the face's "missing" or the author's new `lesson` (the kind
  of question, no names, digits or dates), never the person's question; a failure carries only the
  class of the refusal (`Binder Error`, `Catalog Error: no function to_char`, `a float column`…).
  A failure answers 200 with `failed` and the error, so the face can say it plainly.
- **The kernel's doors move with the kernel**: `rhea init` gave a built-in door only its first
  version, so a changed door never reached a running database. Now a kernel-created door whose
  definition changed gets its next version, approved by the kernel (`activity-approve/<name>/<v>`);
  an unchanged door is never written again (compared through JSON) — the test inits twice.
  `note_untaught` v2 adds `kind` and `detail`; the network's `untaught` table learns both columns.

## 2026-10-18 — the Poland pack says how its books are read (v31)

- **Why.** The ZUS case: market facts (a ZUS payment is a bank operation of kind `zus` going out,
  booked Wn 229; owed is the credit side) were left for the author to guess, and once it summed
  what was owed. Market knowledge belongs in the pack (invariant 8), not the kernel's prompt.
- **`WordsSpec.Notes`**: the market's dictionary may carry notes on how its books are read; the
  kernel shows them to the reporting author beside the collection, for answers and drafted
  reports, and names none of them itself.
- **Poland pack v31**: `slownik` v5 carries thirteen notes, in Polish like the pack — ZUS and tax
  payments as outgoing bank operations (Wn 229, Wn 225), the bank kinds and their accounts, paid
  to suppliers and received from customers through settlements, złoty amounts (`net_pln`,
  `vat_pln`), the VAT month (`vat_period`), open amounts, the result before tax by account type
  and booking month, RMK on 640, the VAT settlement, the CIT advance, declarations, closed months,
  company kinds. History entry 31 in Polish (the other session's rule: every version says what it adds).
- **Live on Helios**: "ile zapłaciłem ZUS w pierwszym kwartale 2026?" now reads `kind = 'zus'` going
  out, taxes read `kind = 'tax'`, customers' payments read settlements against sales invoices.

## 2026-10-19 — KSeF certificate login, step 1: the signed login, proven on TEST

- **Why.** From 1 January 2027 a KSeF token no longer logs in (regulation Dz.U. 2025 poz.
  1815, § 13 ust. 1 pkt 1); a certificate does. The review's KSEF-21.
- **The signer** (`markets/pl/xades.go`, standard library only). The login document
  `AuthTokenRequest` (schema auth 2.0: the challenge, the context NIP,
  `certificateSubject`) gets an enveloped XAdES-BES signature: a reference to the
  document (enveloped-signature + exclusive C14N) and to the signed properties (signing
  time, the signing certificate's SHA-256 digest and issuer/serial), SignedInfo signed
  with ECDSA P-256 (R‖S, XMLDSig 1.1) or RSA-SHA256, the certificate in KeyInfo. The
  documents are ours and small, so they are built as a tree that prints its own
  exclusive canonical form — namespaces where first visibly used, attributes sorted,
  start and end tags always; the transmitted document is that form, so what a verifier
  recomputes is what was digested. Checked independently with Python's `signxml`
  (XMLDSig and XAdES verification, both references, the certificate digest) before the
  first live call.
- **The client.** `KSeF2.Signer` (a certificate and its key): when set, the challenge is
  answered with the signed document at `POST /auth/xades-signature` (XML in, the same
  202 out); status, redeem and the session are the token path's, unchanged. The channel
  reads `RHEA_KSEF_CERT` and `RHEA_KSEF_KEY` (PEM; PKCS#8, EC or RSA key; an encrypted key
  or one that is not the certificate's is refused); a certificate wins over a token,
  half a certificate refuses. The token stays as the fallback until it stops.
- **TEST only, for self-made certificates.** `SelfSignedSeal(nip, organization)` makes a
  company seal (organizationIdentifier `VATPL-<NIP>`, EC P-256), which the Ministry's
  TEST environment accepts for any made-up NIP; DEMO and production need a qualified
  or a KSeF-issued certificate.
- **Live.** `TestKSeFCertificateLive` (gated by `RHEA_LIVE_KSEF_CERT`) logged in to TEST with
  a seal for 8058083055 on the first attempt and sent one invoice in the same session:
  accepted, KSeF number 8058083055-20261019-4834E8800000-BF, the receipt's fingerprint
  ours — and the receipt now names the fingerprint of the signed login document
  (`SkrotDokumentuUwierzytelniajacego`).
- **Offline tests.** `TestKSeF2CertificateLogin` (the fake Ministry verifies the login by
  text surgery, not by the signer's canonicalizer: the document's digest with the
  signature cut out, SignedInfo with its namespace on itself, the ECDSA value with the
  seal's key; a seal for another NIP is refused); `TestCertSignerPEM`;
  `TestClientFromEnvCertificate`.
- **Next (step 2).** Asking KSeF for its own certificate — the login type for Novi, the
  Offline type for QR code II — which needs a signed login first (a token cannot ask,
  § 13 ust. 2); rehearsed on TEST with the self-made seal, then once with the owner's
  signature on DEMO.

## 2026-10-19 — KSeF certificate login, step 2: Rhea asks KSeF for its own certificates

- **What.** `KSeF2.EnrollCertificate(access, kind, name)` (`markets/pl/ksefcert.go`): after a
  signed login, the limits, the identity KSeF read from the signature
  (`/certificates/enrollments/data`), an EC P-256 key generated here, a PKCS#10 request
  carrying exactly that identity, the request (`Authentication` — the login certificate
  that replaces the token — or `Offline`, for the second QR code), polled to its serial,
  retrieved and checked to be for the key that asked. The key never leaves Rhea. After
  a token login KSeF gives no data (§ 13 ust. 2) and the error says so; a name KSeF would
  refuse (21405: letters incl. Polish, digits, spaces, `-`, `_`) is refused before asking.
- **Logging in, in pieces.** The login is split into the challenge, the start (signed
  document or token) and the finish (poll, redeem); `Login`, `LoginRequest` (the unsigned
  document with a fresh challenge, for a person to sign elsewhere within its ten minutes)
  and `LoginSigned` (a document signed elsewhere) are exported for the tool.
- **A command seam.** `adapter.ProvideCommand` and `rhea run NAME …`: a market's tool that
  talks to a statutory system for the owner but writes nothing to the log; handled
  before the store opens; the kernel and the CLI name no command (invariant 8).
  `markets/pl` provides `ksef-certificate`: `test-seal` (a seal for a made-up NIP, TEST
  only), `login-file` (the unsigned login document), `request -kind login|offline -name
  … -out DIR [-signed FILE]` (with the signed file, or with the certificate in
  `RHEA_KSEF_CERT`/`RHEA_KSEF_KEY`); files are written once and never overwritten, the key
  readable by its owner only, and the tool prints the two variables that switch Rhea to
  the issued certificate. `ClientFromEnv` and the tool read one configuration
  (`ksef2FromEnv`).
- **Rehearsed live on TEST, the route Novi will take minus the owner's signature.**
  `TestKSeFCertificateEnrollLive` (gated by `RHEA_LIVE_KSEF_ENROLL`, since each run creates
  certificates): signed login with a self-made seal, a login certificate issued (serial
  01F33EC518362167, two years, issuer "TEST CCK KSeF", key usage digital signature) and an
  Offline one (non-repudiation), then a login with the issued certificate and an invoice
  accepted (8058083055-20261019-48B661C00000-B0). KSeF adds "(uwierzytelnienie)" /
  "(offline)" to the subject's common name. The tool itself was run the same way.
- **Not yet proven.** The `-signed` route with a real trusted-profile or qualified
  signature: it needs the owner's signature on DEMO (Novi), the one step left for a
  person. The Offline certificate's use (signing QR code II) arrives with the QR codes.
- **Test:** `TestEnrollCertificate` against a fake certificate authority in the fake
  Ministry (requests must carry exactly the returned identity, issued on the second
  poll; the issued certificate logs in; nothing after a token login; the login file and a
  login signed elsewhere).

## 2026-10-19 — KSeF certificates for users: parked until the real Rhea UI

- KK: everything a user does happens in Rhea's interface; the command-line tool
  (`rhea run ksef-certificate`) is developer tooling for TEST only, never a user step.
- The user flow, as the market does it: the owner or the accountant generates the
  certificate in the Ministry's KSeF application (Profil Zaufany or a qualified
  signature), saves the `.crt`, the `.key` and the password, and uploads them on a
  "Połączenie z KSeF" screen in Rhea, which checks the connection at once, says plainly
  what is wrong if it fails, keeps the key encrypted outside the books, and warns before
  the two years end. The token moves to the same screen until 31 December.
- Owed when the UI is built: reading a password-protected key, encrypted storage of the
  certificate and password, the kernel's upload/test/status endpoints, the screen, the
  expiry warning; to verify on TEST: renewal by Rhea with the current certificate.
- Shipped and kept: certificate login and enrollment (c41f1c7, c960f19), proven on TEST.

## 2026-10-19 — The auditor's exam as the yardstick: `rhea run pibr-exam`

- KK: Rhea signs nothing, but its capability in accounting is to be proven at or above a
  biegły rewident's, on the exam itself, as the headline and the ERP's credibility — and
  later in every market with a pack. The claim to make is about the rulebook, never the
  model: every answer resting on an approved rule and its article, replayable, graded by
  the Commission's own key.
- The material: the Commission does not publish the exam items (the bank is reused); it
  publishes sample papers with model solutions, 75 PDFs over ten subjects and some ten
  years (pibr.org.pl/pl/egzaminy-pytania), the January 2026 syllabus (KRBR 1695/34/2026,
  342 numbered points, 225 lettered) and the reading list with the acts and standards per
  subject (KRBR 841/20/2024). All three are in research_notes/pibr-egzaminy/, kept out of
  git: the papers are the Commission's and the repo is headed for open source. ArsLege's
  bank is online only under a personal-use licence that forbids retrieval storage — a drill
  for a person, never input for Rhea.
- The harness: `internal/exam` is the generic sitting — a Paper of keyed single-choice
  questions and situational tasks, a Scoring, answerers, a per-block tally and a report
  that says where every answer came from. The paper's syntax and the marking are the
  market's (invariant 8): `markets/pl/pibr.go` parses the Commission's PDF through
  pdftohtml's XML (the right option is the bold one; headings come letter-spaced and are
  collapsed) and carries the regulation's marking — +2 right, −1 wrong, 0 blank, a block
  passes at 60% of its maximum on its own — and the command. Test questions only for now;
  the tasks are parsed with their model solutions and counted, to be sat through the
  kernel next (facts → events, the task's chart → account roles, rules post, formulas
  compute, the view names the balance line).
- Two sitters, reported side by side: the bare model, which may leave a question blank as
  a candidate may; and the host with the rulebook open — the pack's rules with their
  warrants — which must label each answer "rule" (a rule it names and the citation it
  relied on; a rule not in the rulebook is re-labelled memory, so the share from rules is
  never inflated), "memory" (the statute it knows itself, with the article, saying the
  pack was never taught it) or "abstain" (never taught; names the missing lesson).
- Honest limits: an answer from memory is still the model's; the number that matters is
  the share from rules and its accuracy, and it is small today because the pack is a
  bookkeeper's, not an auditor's. The 2025 financial accounting paper has 50 questions in
  three blocks (one unkeyed: the Commission marked no option on III-7) and 10 tasks. The
  bare model's score is the baseline every later claim is measured against, published
  beside Rhea's.
- **Measured, 2026-10-19, financial accounting 2025, claude-sonnet-4-6, 49 keyed questions,
  two sittings.** Bare model: 71 of 98 (72%) then 68 of 98 (69%), never blank, NOT PASSED
  both times — block I (the Accounting Act) 80% passed, block II (banks and SKOK) 25%
  failed, block III (insurers, public sector) 100% then 83%. Host with the rulebook open:
  first 44 of 98 (44%) with 21 abstentions and 3 replies the reader dropped; after the
  lenient reader and the rule "answer from memory when more likely right than wrong",
  61 of 98 (62%), block I 80% passed, II 25%, III 44%. From rules: 0 of 49 in both
  sittings. From memory: 45, 35 of them right. Abstained: 5, each naming the lesson the
  pack lacks. The reading: the model knows the Accounting Act well and sector accounting
  (banks, insurers, public bodies) and IFRS unreliably; the Poland pack, a bookkeeper's,
  has nothing to say on an auditor's paper yet — which is the point of the yardstick. The
  exam now generates the pack's missing-lesson list: every memory or abstain note names
  one, 50 for this paper. The bare model moved three points between two sittings of the
  same questions: one sitting is not a measurement; rounds, as with eval. A sitting of
  both is about 100 calls and three minutes. Results kept in
  research_notes/pibr-egzaminy/results/ (out of git).
- Next: the situational tasks through the kernel; the syllabus graded into the pack's
  review register; the other 2025 papers (the harness takes any of the 75); the host
  reading the statute's text once retrieval by article lands, so "memory" becomes
  "statute" with a dated version.

## 2026-10-19 — The auditor's exam, step 3: a situational task through the kernel

- The claim that matters is the one no chatbot can print: the ERP did the task with its own
  hands. `rhea run pibr-exam -paper FILE -task I-1` reads the task and the Commission's
  model solution into a case (`internal/exam.Case`: the company's chart as it is given,
  the facts as events, the instructions as intents, the solution's entries as expected
  postings, the rest as questions for the host), written to testdata/exam/<paper>/<task>.json
  for a person to review — the file is kept and reused, the model reads only once — and
  sits it as an implementation in miniature on a throwaway log: the finance base seeded,
  the chart and the facts as intake, the author drafting a bundle per intent, the recorded
  approver reading each dry run, then the books read against the solution. The reading is
  the model's (as it reads a PDF into a fact); the drafting is the author's normal loop;
  the posting is the kernel's; the comparison is exact in values.
- Two tasks of financial accounting 2025 DONE on the first sitting after review. I-1 (a
  write-down of shares at the balance sheet date): the chart approved, Dt 751 / Ct 145
  10 000, nothing unexplained. I-2 (foreign currency: an advance and its settlement, a
  prepayment under KSR 11 pkt 6.18, a penalty received and revalued — six events, twelve
  lines): seven bundles approved, one after a re-ask in which the kernel refused an entry
  that did not balance and the author balanced it; every amount the Commission's table
  gives is in the books.
- Three things learned. A bundle drafted without a sample was dated 2026-01-01 and never
  reached a case set in 2024: eval now dates it from the first fact in the log. The
  solution's lines on one account add up: `expect_postings` compares, per account and
  side, the solution's total with the books' — an entry compressed or split still agrees.
  Not every solution is entries: I-2's is a table of amounts, rates and differences, so
  the case's entries are the reviewer's reading of those amounts against the task's chart,
  said so in the case's notes; the review is a real step, and the model's first reading
  of I-2 had one line gross on one side and net on the other.
- A case may carry only facts whose entries the solution lists (a purchase the task does
  not book is context, folded into the event it bears on), and every event must be named
  by an intent — otherwise the case can never be done. The parser now says why a question
  has no key (no option in bold, or two), and the page footer ". str. 3" of some papers no
  longer glues itself to the last option.
- **Measured, 2026-10-19, round two: all ten 2025 papers under the final parser,
  claude-sonnet-4-6, 429 keyed questions, 858 points.** Bare model: 708 points (82%),
  passes 6 papers of 10 — theory, economics and internal control, law, tax I, cost
  accounting, audit and ethics; fails financial accounting (the banks' block), finance
  (block III), tax II (46% on block I) and statements (block I). Host with the rulebook
  open: 590 points (68%), passes 3 of 10 — theory, economics, cost accounting; 0 of 429
  answers from a rule on any paper, 380 from memory, 63 abstentions naming the lesson.
  The host's deficit against the bare model is its abstentions: the rulebook has nothing
  on these subjects, and the sharpened rule still leaves it blank where it is unsure.
  Task I-3 (events after the balance sheet date under KSR 7, two no-entry decisions, 20
  accounts, 13 lines) DONE on the model's own reading, reviewed afterwards against the
  Commission's table: three tasks of financial accounting 2025 for three. Parser debt: 14
  questions across the ten papers read as two options in bold and stay unkeyed; one
  question in finance (III-8) swallowed its neighbours' options; to check by eye.
  Reports in research_notes/pibr-egzaminy/results/round2/.
- Parser debt paid the same hour: the papers that print their footer as "Strona 4" (a
  capital S) had it glued onto the last option; the footer pattern is now case-insensitive
  — and the audit paper's footer sets its word and its number four pixels apart, so the
  baseline grouping widened to five and the footer pattern takes either alone. The ten
  papers carry 5 unkeyed questions instead of 14: two the Commission's own (no option, or
  every option, in bold), one on finance that swallowed its neighbours' options, two to
  check by eye. The next round sits them.

## 2026-10-19 — The auditor's curriculum as the pack's register: packs/pl/EXAM.md

- The January 2026 syllabus read as a map of what the Poland pack must know: 421 rows
  (every numbered point of the ten subjects, the lettered sub-items where the books live
  — theory, financial accounting, the law's contracts), each graded by what in Rhea
  answers it — *books* (the pack as data: 83), *kernel* (by construction: 20), *engine*
  (code a writer or an engine must provide: 28), *host* (knowledge and judgement answered
  in words from the statute, never a rule: 290) — and, for the first three kinds, what
  the pack holds today: covered 22, partial 77, missing 52. The 2025 sitting stands
  beside each subject. Prepared by AI, not yet read by an auditor, as REVIEW.md was.
- What it says: the books' own subjects — theory, financial accounting, cost accounting —
  are covered on the bookkeeper's half and missing the auditor's half: fixed assets and
  depreciation, inventories and cost of manufacture, equity and the distribution of the
  result, provisions, deferred tax, impairment, leasing, financial instruments, year-end
  closing, the statements themselves (one writer, the e-Sprawozdania XML, and the derived
  statements). Tax II is the nearest gap with the most rules to write. Economics, the law
  as such, the finance of banks and insurers and the audit profession's own rules are the
  host's — 290 of 421 rows — which is why the statute by article in front of the host is
  the next lever, and why no rule will ever answer those papers.
- Sector charts hold. Task II-1 of financial accounting 2025 — a SKOK with the sector's
  model chart, 32 accounts, 8 operations, 20 lines, off-balance entries included — was
  DONE through the kernel on the model's own reading, not yet reviewed by a person. Task
  III-1 (insurers) is not a bookkeeping task: a review of two insurers' accounting
  policies with verdicts and recommended changes; the harness now says so plainly ("a
  task without entries — the host's to answer in words") and keeps the case with its
  answers for the host, instead of failing. Five tasks read so far: four sat and done,
  one the host's.

## 2026-10-19 — The statute in front of the host: `internal/statute`, `rhea run pl-acts`

- Retrieval by citation, not by similarity (DECISIONS 2026-10-19, "Are the rules enough"):
  the host names the provisions it would open, is handed their text, and may answer
  "statute" only citing a provision it was handed and quoting its decisive words; a
  citation of anything else is counted as memory. Every statute answer therefore carries
  the act, the article, the quoted words and the consolidation the text was made from.
  No vector store, nothing outside the stack: the library is JSON files of acts split
  into articles, the host's consult step is one model call that returns citations, and
  the lookup is by the act's name and the article's number.
- The acts are a market's. `rhea run pl-acts fetch` takes each act the Poland pack cites
  from the Sejm's ELI service (api.sejm.gov.pl/eli: the metadata and the Chancellery's
  consolidated text, type U) and reads the PDF into articles, the page headers dropped,
  the consolidation basis from the first page's margin kept whole. Ten acts today — UoR,
  VAT, CIT, PIT, Ordynacja, KSH, KC, KP, the auditors' act, PCC — 4 413 articles, in
  data/acts/pl (out of git; regenerable; the texts are public domain). `internal/statute`
  is generic: an act, its names, its articles, a lookup, a citation parser.
- **Measured, host with the statute open, three 2025 papers, claude-sonnet-4-6.** Theory:
  94 → 111 of 120 (78% → 92%), above the bare model's 105; 41 of 60 answers from quoted
  statute, 40 right. Financial accounting: 64 → 67 of 98; statute 18, 16 right; wrong
  answers 8 → 5. Tax I: 52 → 51 points but PASSED (both blocks over 60%); statute 17,
  16 right. Across the three: statute answers 72 right of 76 (95%), memory 48 of 55 (87%).
  Where the act is in the library the host now reads it rather than remembers it; where
  it is not (banks, insurers, KSR, the regulations), it still remembers or abstains. The
  share from rules is unchanged at zero; this step makes the host's words verifiable, not
  the pack's rules larger.
- Owed: the KSR and the regulations the papers cite (the Ministry publishes them outside
  ELI); the rulebook's own warrants resolved to text in front of the host; a second sitting
  of each paper, as always.

## 2026-10-19 — Pack v32: fixed assets, the first auditor's lesson in the pack

- The register's most-tested missing row (EXAM.md 6.1.2 a; financial accounting 2025,
  question I-4) as pack rules, modelled on the prepaid-cost mechanism: a door
  `accept_fixed_asset` (the OT: name, inventory number, KŚT, initial value, the day of
  acceptance, the annual rate from the Wykaz) makes a `fixed_asset` (v2: annual rate,
  start, status with a lifecycle active → depreciated → disposed) and books Wn 010 /
  Ma 080 on the day of acceptance; when a month opens, every active asset gets a
  `depreciation_charge` for the month that ended — cumulative, initial value × rate ×
  months since the start ÷ 12, capped at the initial value, minus what was charged
  before, so a skipped month is caught up at once — booked Wn 401 / Ma 071 on the last
  day of its month; an asset whose charges reach its value is depreciated and charged no
  more; a door `dispose_fixed_asset` (the LT) moves it to disposed and books Wn 071 for
  the accumulated depreciation, Wn 761 for the net value, Ma 010 for the initial value,
  a fully depreciated asset without the zero cost line. Account 080 Środki trwałe w
  budowie joins the chart; views `srodki-trwale` and `amortyzacja`; the dictionary names
  it all. Every rule's warrant quotes the provision from the act in the library: UoR art.
  32 ust. 1 (systematic, planned, from acceptance to the equalling of charges and value
  or liquidation), ust. 3 (rate and method set on the day of acceptance), ust. 7 (a small
  entity may depreciate by the tax rules — Rhea's default for a small firm: one rate,
  tax and balance), CIT art. 16h ust. 1 pkt 1 (from the month after entry into the
  register) and 16i ust. 1 (the Wykaz), KSR 11.
- Simplifications of this version, said in the history: one rate (tax = balance), the
  linear method, no charge in the month of disposal, improvements (art. 31 ust. 1) and
  the sale of an asset in the next version. A degressive method, a separate balance
  rate (with the deferred tax it creates) and low-value one-off write-offs (art. 32 ust.
  6) are the rows that follow.
- The kernel learned one thing: a door may ask for a decimal (`activity` input type
  "decimal", a string across the boundary like money) — a rate, a coefficient; the shell
  shows it as a number. Without it the rate had no honest type (int cannot say 2,5 for a
  building). Doors take decimals from here on.
- Test: `TestFixedAssets` in the shell — a laptop at 20% from 15 September: 200 a month
  from October, a skipped January caught up with February's charge of 400, scrapped on 10
  February with 800 off 071, 11 200 into 761 and 12 000 off 010, no charge after; a
  printer at an absurd 600% depreciated in two months and scrapped without a cost line.
  The pack's pinned counts moved (26 types, 28 views, 60 rules, 12 doors, 50 events).
- **The first answers from rules (2026-10-19, pack v32 with the statute open).** On the
  v32 sitting the host answered question I-4 — a small entity may depreciate by the tax
  rules if its policy says so — by quoting UoR art. 32 ust. 7 from the statute in front
  of it, the very provision the new depreciation rule's warrant carries. An answer that
  rests on a provision a pack rule is warranted by is the pack's lesson: the host now
  labels it "rule" and names the rule (the same act and article, the unit the warrant's
  or inside it; a note says how it was first answered), so a reader opens the rule and
  its warrant beside the answer. Measured: financial accounting 2 of 49 from rules (both
  `pl-accept-fixed-asset`: art. 3 ust. 1 pkt 15 and art. 32 ust. 7), theory 3 of 60
  (`pl-post-explained-bank-operation` on art. 3 ust. 1 pkt 32, `pl-set-cost-account` on
  art. 10), all five right; the host 68 and 108 points on the two papers (one sitting
  each; the financial accounting host moved 77 → 68 between two sittings of the same
  questions, the usual reminder). The share from rules is no longer zero, and it moves
  the only way it can: a lesson at a time, each with its article.
- Owed: the live databases are on pack v31 — `rhea pack packs/pl/pack.json` drafts the
  upgrade bundle and a person approves it in the shell; not done from here.

## 2026-10-19 — Pack v33: the year-end close, the statement's approval, the result's distribution

- The second auditor's lesson, and the one the first needed: a distribution without a
  close would debit an account nothing had credited. Four doors, each a step the law
  names. `close_year` (UoR art. 12 ust. 2 pkt 1): every cost, revenue and income-tax
  account gives its balance on the balance sheet date to 860 — one close object per
  account, three types by the side of the result (`year_close_cost`, `_revenue`, `_tax`),
  because the kernel refuses two rules minting the same object from one event, and
  rightly: one fact, one rule. `approve_statement` (art. 53 ust. 1; KSH art. 231) moves
  the result from 860 to 820 on the day of approval, profit or loss. `distribute_result`
  (art. 53 ust. 4; KSH art. 191–192) is the resolution: so much to reserve capital (Wn
  820 / Ma 802), so much to dividends (Wn 820 / Ma 241), so much of a loss covered from
  reserve (Wn 802 / Ma 820); what is left stays on 820 as undistributed. `pay_dividend`
  (PIT art. 30a ust. 1 pkt 4 and art. 41 ust. 4; CIT art. 22) books the gross off 241,
  the net to 130 and the 19% withheld to 225, or no tax line at a rate of 0 under the
  exemption. Accounts 802, 820 and 241 join the chart; views `kapitaly` and
  `podzial-wyniku`; the dictionary names it all; the warrants quote the provisions.
- Simplifications said in the history: the comparative variant without the change in
  products (490 stays out), no interim dividend, no contributions in kind. The close is a
  door, not an automatic step on the December lock, because the month's own close posts
  the CIT advance on the same event and a balance read then would miss it; closing the
  year is a deliberate act after the adjustments, as art. 12 ust. 4 has it.
- Test: `TestYearEndAndResult` — 1 500 of revenue and 650,01 of costs close into a profit
  of 849,99; approval on 30 April moves it to 820; the resolution sends 68 to reserve and
  750 to dividends, 31,99 stays; 700 paid to a person leaves 567 at the bank and 133 on
  225, 50 paid to an exempt company carries no tax line. Pinned counts: 30 types, 30
  views, 73 rules, 16 doors, 53 events, 42 accounts.
- **Measured on v33 (one sitting each).** Financial accounting: 4 of 49 answers from
  rules — three right (the fixed-asset rule twice, the bank-operation rule on other
  operating costs), one wrong: II-8, a SKOK's net result, answered through the close
  rule's warrant on UoR art. 42, which opens "w jednostkach innych niż banki" — the host
  stretched a rule to a sector it excludes; the exam catches that too. Theory: 5 of 60,
  all right, one of them the reserve-capital rule cited by name. Two papers: 9 of 109
  from rules against 5 before v33, 8 right.

## 2026-10-19 — Pack v34: provisions, accruals and receivable write-downs

- The third auditor's lesson (EXAM.md 6.1.2 d, 6.1.3 d, 6.1.5, 6.1.8 c), five doors.
  `create_provision` makes a provision for a liability (UoR art. 35d ust. 1: a certain
  or highly probable future obligation, reliably estimable — a lawsuit, a guarantee)
  charged to other operating costs against 830, or an accrual (art. 39 ust. 2 pkt 2:
  warranty repairs, benefits to persons not yet known) charged to the cost it belongs to
  against 641. `use_provision` is art. 35d ust. 3: the obligation arrives as an invoice
  the pack has already booked to a cost, so the provision shrinks and the cost comes off
  — the cost was in the provision. `release_provision` is art. 35d ust. 4: what proved
  unnecessary goes to other operating revenues on the day it did; an accrual released
  reduces the cost it charged (KSR 6). `write_down_receivable` is art. 35b: a sales
  invoice's receivable written down for the debtor's insolvency, a dispute, lateness; a
  tax cost (761) only when uncollectability is made probable in the CIT act's sense —
  insolvency, liquidation, restructuring (art. 16 ust. 1 pkt 26a and ust. 2a) — else
  761-3, non-deductible. `settle_write_down` ends it: the debtor paid and the write-down
  returns to revenue, or the receivable is uncollectable and leaves the books against the
  write-down (art. 35b ust. 3–4). Accounts 830, 641, 280 and 761-3 join the chart; views
  `rezerwy` and `odpisy-naleznosci`; the dictionary names every value.
- Two small things learned. A lifecycle refuses a status set to its current value, so a
  type whose amendment may leave the status as it is declares the self-transition, as
  the bank operation does. A money field born at zero takes the literal "0.00" in the
  template, since the formula language has no money literal.
- Said in the history: the written-off receivable does not yet close the invoice's open
  amount in settlements (a write-off as a settlement kind is owed); write-downs of
  inventories come with the inventories lesson.
- Test: `TestProvisionsAndWriteDowns` — a lawsuit provided for at 10 000, used for 6 000
  and released for 4 000; a 5 000 warranty accrual used for 1 500 and released for 3 500
  against 403; a bankrupt customer's 1 230 written down as a tax cost and written off
  against 201; a late customer's 615 written down as non-deductible and reversed to 760
  when paid. Pinned counts: 32 types, 32 views, 88 rules, 21 doors, 58 events, 46
  accounts.
- **Measured on v34 (one sitting each).** Financial accounting: 6 of 49 from rules, 5
  right (the fixed-asset rule twice, the write-down rule on art. 35b ust. 2, the
  provision rule twice on art. 39 ust. 2), the same II-8 wrong — the close rule's art. 42
  stretched to a SKOK; block I at 90%, 73 of 98 overall. Theory: 7 of 60, all right, 111
  of 120. The share from rules on these two papers, pack by pack: 0 → 5 → 9 → 13 of 109.

## 2026-10-19 — Upgrading the live databases: orphan drafts are adopted, not skipped

- Asked to bring the live databases to pack v34, I found them where they were, not where
  I had said: rhea_pl on v18 with a stale v19 draft bundle never approved; rhea_face on
  v31; the default rhea without the Poland pack. Loading v34 into rhea_pl failed partway
  — its schema predated the dictionary's "words" notion, and `rhea init` (idempotent;
  the schema file carries the migration) had never been re-run there. The failed load
  left eighteen types and four views as drafts with no bundle, and a re-run would have
  skipped them as "already present" and shipped a bundle without them: an installed pack
  with its types still drafts.
- The loader now adopts an orphan — a draft version already on the record that no draft
  or active bundle holds, left by a load that failed partway or freed by a rejection —
  into its bundle (`Summary.Adopted`), and the bundle's guard lets a member of a rejected
  bundle join a new one; a member a pending or active bundle holds is still that
  bundle's. On rhea_pl the v34 load then adopted 133 orphans (the two failed loads' rows
  and v19's freed drafts) and drafted `pack-pl-v34`; on rhea_face it drafted the same
  bundle directly. `TestALoadAdoptsOrphanDrafts` covers the orphan, the pending bundle
  that keeps its own, and the approval that activates the adopted draft.
- The approval is not mine to give: the gate wants a person (invariant 2), and the
  agent approving the bundle it drafted is the one thing the gate exists to prevent — the
  session's classifier said as much before I had thought it through. Both bundles wait
  for KK in the shell (`./shell.sh 8071` for rhea_pl; a shell with RHEA_PG_DSN on
  rhea_face) or by `rhea approve pack-pl-v34`. Backups of both databases from before the
  load are in data/backups/. The server at white01 is not reachable from this machine.
- Owed: `rhea pack` as a screen in the shell (KK 2026-10-19: every user step in the
  app); `rhea init`'s migration to run on `rhea serve` as well, so an old database never
  refuses a new notion again.

## 2026-10-19 — The standards in front of the host; the learning-rate test

- KK, evening: the quiz score is the model's; what matters is what Rhea knows, and
  "she learned nothing all day" at a person's teaching pace. Two answers built the same
  evening. First, the standards: `rhea run pl-standards fetch` takes the sixteen Krajowe
  Standardy Rachunkowości from the Ministry's journal and reads each into its points
  (KSR 11 alone 272; some fourteen hundred in all), the library loads acts and standards
  from several folders, "KSR 11 pkt 6.18" is a citation the host can name and be handed.
  Reading is most of what the curriculum asks and it costs nothing to teach. Second, the
  learning-rate test: `pibr-exam -learn SITTING -block I` turns every non-rule answer of a
  prior sitting into a lesson for the author — the question, the key, the provision the
  host read — lets the author draft definitions on a throwaway database with the
  recorded approver reading each dry run, sits the block again with that rulebook, and
  reports the share from rules before and after against the clock. The pass line KK
  and I agreed: a real move in an hour of machine time with rules that stand against the
  text; else the approach fails at the thing he needs and stopping is right.
- Also this evening, sitting five ArsLege company-law questions pasted by hand, the host
  answered five of five from the statute, naming the article it opened and quoting the
  words (KSH art. 201, 237, 247, 252, 286) — the thing the product sells, seen plainly.
- **Round three, 2026-10-19 night: all ten 2025 papers, three sittings each, the host with
  the acts open (standards not yet), pack v34, claude-sonnet-4-6.** Six papers passed on
  every sitting — theory 90%, law 90%, statements 78%, cost accounting 75%, economics
  75%, tax I 73% — and four failed on every sitting: finance 78% (banks' block), financial
  accounting 68% (banks' and insurers' blocks), tax II 57%, audit and ethics 56%. 75% of
  points over thirty sittings; the spread within a paper is a few points (theory
  111/109/106), so a sitting is now a measurement. Where the answers came from over all
  thirty: statute 432 (92% right), memory 670 (90%), rule 42 (95%), abstain 185. The
  statute share is low on the papers without an act to open — economics, finance, cost
  accounting, audit — which is where the standards and the KSB now in reach will tell.
- **The learning-rate test, two runs, 2026-10-19 night.** Financial accounting block I, 30
  questions, 25 lessons (every answer of the v34 sitting not from a rule). Run one: 25
  bundles approved on the throwaway in 14 m 27 s, the share from rules 5 → 15 of 30, all
  fifteen right. Run two, drafts kept: 25 approved in 14 m 20 s, 5 → 17 of 30, all right;
  six lessons "already answered" by a bundle drafted earlier in the same run. The score
  stayed at 54 of 60 both times — she already read those answers — what moved is whose
  they are. The author drafted some 60 rules, 20 doors and 15 types per run, each bundle
  warranted by the article: going-concern valuation (art. 29), fire damage during
  construction (art. 28 ust. 3), finished-goods write-downs, price deviations on goods,
  reclassification to investment property (art. 31 ust. 4), advances for fixed assets,
  loans at nominal value (art. 28 ust. 11), idle capacity, development costs and the
  distribution restriction (art. 33), deferred tax (art. 37), a change of inventory
  policy (art. 8), FX equity investments, sales of materials. Five of twenty-five rest on
  IFRS and are not a small company's pack; the rest go to review. The pass line KK and I
  set (twenty points in an hour, rules that stand against the text) is cleared twice
  over; the slower part is now the person reading the drafts. Next: the review of the
  block I drafts into pack v35 with its story tests, then the same loop block by block
  and paper by paper, each measured the same way.

## 2026-10-20 — Fixing day: the review's fifty-five findings

- KK handed over a read-only review of the tree at 29c85d9 (55 findings, each with a
  failure scenario and the simplest fix). Confirmed first: the suite was not green —
  `go test ./...` uncached failed TestBundleRejection every run, hidden by the test cache
  since d4d243d changed the bundle rule (a rejected bundle frees its drafts) and left the
  older test asserting the opposite. Then the findings in order of harm, one commit a
  batch, every fix with a test where one could have teeth.
- Decided on the way, where the finding asked for a decision:
  - The `approved` status was a constant never written. A rule goes
    draft → active → superseded; the approval is the recorded event, not a resting
    status (invariant 2 says so now in CLAUDE.md and SPEC.md).
  - A bare number copied into a money field: minor units in a derived payload (the
    kernel's own convention — an object's state, a match's open amounts), refused in a
    raw fact, where money crosses as text (invariant 6). The executor marks each fire
    with its event's kind (`core.Env.Raw`); refusing everywhere broke the matching
    rules every persona writes (`=$.line.open_after.money`).
  - Path conditions compare exactly: a number against a number or against money text
    compares as numbers; two texts compare as text, so an identifier "0100" is not "100".
  - A door accepts a decimal typed as a JSON number and refuses one that carries the
    sixteen or seventeen digits only a computed float has.
  - The executor runs one chain at a time under one Postgres advisory lock, across
    handlers and processes; a trigger's worklist pass runs under the lock it holds.
  - Replay is the whole log's whatever company scope it runs under, and `rhea replay`
    compares the cache before and after (exit non-zero on a difference).
  - A read of a narrative view speaks nothing; `POST /api/speak` (now also `?view=`)
    speaks. Three tests that read a briefing to make the host speak were changed.
  - A market reaches the kernel through a named list of packages (adapter, core,
    statute, pack, agent, exam); the exam harness pulls the eval and the shell behind
    it, and moving the Polish exam parts under markets/pl is the debt that shortens the
    list. The import test scans test binaries too.
- Assessed and left, with the reason: the client's three unrenderable notions and the
  hard-coded worklist verb (#33), domain names in the shell's untaught/narrative code
  (#34), the DuckDB rebuild per analysis request (#35, an accepted cost per CLAUDE.md),
  the N+1 ref cells (#37), Polish knowledge in the kernel's MT940 and PDF readers (#43 —
  a hook for markets to register dialects, a session of its own), floats on the
  projection's money path (#44, numerically safe today), the two long switch functions
  (#50), the duplicated helpers (#53), signal handling and cwd-relative paths (#54), and
  `rhea load` without an adopt step (#31's second half). Each is a line in the review,
  none corrupts state.
- Running the suite is `go test ./... -count=1`: the cache hid a failing kernel test for
  a day. storetest fails instead of skipping when RHEA_CI is set; cmd/rhea has a smoke test.

## 2026-10-20 — The second review, after the fixing day

- KK handed over a second read-only review at 94d28da (54 findings, 8 critical), run by
  six parallel reviewers. The first review's fixes held; the new gaps sat one level
  deeper, and three criticals were in the fixing day's own work (the executor lock, the
  network's open, the pack loader's bump check). Two commits: the eight criticals, then
  the high and medium findings with the cheap lows.
- Decided on the way:
  - A stored file is served as a sender's bytes: nosniff, a sandboxing CSP, and a
    download unless it is a PDF or a raster image. The shell's own origin is the human
    gate's origin; an emailed invoice.html must never script the approval doors.
  - The executor's lock has an in-process gate before the pool: one goroutine per
    process waits on Postgres holding a connection. The advisory lock stays for the CLI
    beside the server.
  - A chain is bounded by size (ten thousand consequences of one event) as well as
    depth; a number by magnitude (4096 bits in numerator or denominator). Both refuse
    into the worklist naming the rule.
  - Two texts that both read as decimals compare as numbers in formulas. Identifiers
    are text; money in a raw fact is text with a point, and "100.00" is more than
    "99.00".
  - A KSeF pass books every final result of a failed session and names what was sent
    without a final word. Owed: polling the named reference next time instead of
    resending it (the channel's event list would grow by one type, a pack change).
  - A refused invoice is fingerprinted on its content with the creation time fixed, and
    that is the rejection's `invoice_hash`; the exact sent hash rides beside it as
    `sent_hash`.
  - The accepted date of a KSeF invoice is the UPO's received timestamp in Polish
    time, not the clock that ran the pass (VAT act, art. 106na).
  - Inbound documents deduplicate on what they are: a KSeF invoice on the seller's VAT
    id and the number, a PDF on the supplier's VAT id and the number, an MT940 line on
    the account, the statement and the line. Three tests that fed the same documents
    twice and expected two bookings were changed to expect one.
  - A backfill approval books its chains in its own transaction, as a bundle's members
    activate; one failing chain refuses the approval.
  - The JPK totals are sums of the rounded positions, and the file's creation time is
    the period's last instant unless given, so the writer is deterministic.
  - The DuckDB copy's key is the log's highest id with its count, and carries no
    company scope. Owed: a company column in the projected tables, so pack SQL can
    filter; today a scoped analysis sums every company.
  - The web shell can make the host speak again (a button posting to /api/speak), shows
    when the words were said, and dates a person's fact by the books' day read from
    the kernel's clock.
- Assessed and left, each a line in the review: a forwarded late document matching
  rules by the forwarded date (#10 — the chain is re-expanded on the new date by
  design of the forward; to revisit with the dry run), the dry run's own balance and
  matching (#12), an approval activating a version it does not name under a concurrent
  draft (#13 — the executor lock now covers the insert path from the doors; a direct
  store insert can still race), full-log scans in the pairing verbs (#14), duplicated
  fan-out code (#16), a derived payload's plain integer read as minor units on a copy
  and major units in arithmetic (#18 — the kernel's convention; packs copy or compare
  to zero), dead helpers and the formula cache (#26), SQL in Go literals (#30, about
  thirty statements, mechanical), doubtful PDFs re-read every pass (#47), status codes
  beyond 404 (#41 in part), the remaining small client defects (#42 in part), and the
  MT940 reader's globals (#54 in part).

## 2026-10-20 — THE product, and isolation enforced in the database

- KK: "We are now building THE product. We gathered enough proof to be able to confirm
  the state change." CLAUDE.md rewritten: what this is, invariants 9 (isolation) and
  10 (concurrency and idempotency), the working style the two reviews taught (one
  finding, one failing test, one fix, one commit; the boundary checklist; the check
  script; an independent review before done), the definition of done for a change.
- No hosted CI: `./scripts/check.sh` (gofmt, vet, `go test ./... -count=1 -race` with
  RHEA_CI=1 against the real database) runs on every push through `.githooks/pre-push`
  (`git config core.hooksPath .githooks`, once per clone). Same protection, no credits.
- Isolation (DIRECTION, "Isolation in the database"): a scope is a set of companies;
  an accounting office's scope is its clients, acting for one named client when it
  writes. Rows with no company are the installation's and every scope's. The database
  enforces it: role `rhea_app` with row-level security on event and object; the pool
  switches role and sets the scope on every connection a scoped request gets and
  resets both on return; a failure to set either discards the connection. The analysis
  copy is built per scope, under the scope, into the scope's own file, and every
  projected table carries `company`. The store's own filters admit the installation's
  rows now, as the policy does (the kernel's seeded verbs are every scope's).
- Until sign-in, the scope comes from the request header (one key, or a comma list for
  an office) and an unscoped request is the owner's whole log, as before; the sign-in
  session makes a scope mandatory. Owed: `rhea restore -company` to stamp facts from
  before the partition key; a company column on the blob store when files are listed.
- Tests: the database refusing another company's rows and writes under raw SQL
  (store), the copy per scope (project), two companies at the shell's doors with an
  office scope summing both (shell).

## 2026-10-20 — The sign-in step: who may do what, as data, enforced at every door

- KK: "the time has come to tackle the sign-in domain" (DIRECTION 2026-10-19, "People
  and rights"). Shipped in the kernel: the notions as data, enforcement at every door
  and every read, the who-can-do-what view. The face's passkeys and screens follow in
  rheaface; the GUS channel in markets/pl; the house's doors after.
- The common people pack v3 adds five types in the people domain: `role` (a name and
  its rights), `membership` (a person in a company with a role from a date, to a date,
  with a group and a reason, signed by who granted it), `mandate` (an office keeps a
  client's books within a scope of rights — KSeF's authorised entity), `substitution` (a
  deputy holds a person's rights between two dates, in one company or all), `group` (a
  named set of a company's people). Nine doors grant and end them. The pack's roles are
  master data: owner `*`, chief_accountant `door:* view:* type:* domain:people`,
  accountant and support `door:* view:* type:*`, reader `view:* type:*`. The person's
  own doors (register, identify, deputy, status) are installation-wide.
- The rights vocabulary (`core.Right`): `*`, `domain:<d>`, `door:<name>`, `view:<id>`,
  `type:<name>`, `state:<type>/<state>` (vocabulary now, enforced with smart documents),
  `sees:<group>`, `approves:<group>` (the supervisor's two rights, vocabulary now).
  `door:*` never reaches the people domain — KSeF: managing rights is its own grant,
  never implied by the right to do the work; `view:*` and `type:*` do (seeing who may do
  what is not managing it). A role's rights are a space-separated string.
- Why roles are objects and not Who on the door: definitions are the installation's,
  roles are a company's (an office's accountant is not its client's), so a company's own
  role could not live on a door without changing every company's door. `Who` keeps its
  job — which kind of actor (human, agent, adapter) a door admits — and is now enforced
  for a signed-in person: the clock is nobody's to move, the host's words are not a
  person's to forge. One question at every door: Who admits a person, and a grant covers
  the door in the company acted for.
- The grants are computed from the log on the books' day (`internal/rights`): a
  membership gives its role's rights in its company; an office member reaches every
  client under a current mandate with the role's rights narrowed by the mandate's
  scope; a deputy holds the substituted person's own grants (never their substitutions,
  so nothing chains). A company's own role of a name wins over the pack's. Dates decide,
  not status: an ended membership answers who could do what on the days it held, and
  to revoke someone now the owner ends it with yesterday as the last day.
- A person's grants are their scope (invariant 9 made whole): the shell resolves the
  person named in `X-Rhea-Person`, sets the scope to their companies, and acts for the
  one named in `X-Rhea-Company` (the only one, when they have one). A person nobody has
  let in anywhere is scoped to a key no company carries and sees the installation's
  rows only. On the desk open on all their companies, a read must be theirs in every
  one; a write names its company.
- A company nobody holds — no current membership or mandate names it — is open to its
  first grant, which is its claim: a new owner grants herself, an office grants itself
  the mandate. A second claimant is refused. In the beta this is enough (invitation
  only); in public the e-mail is verified, the NIP checked against GUS, and a dispute
  goes through a KSeF-authenticated step (the brief's standard, not built).
- A refusal is a fact: `access.refused` under the company refused, signed `kernel`,
  naming the person, the door and the reason; the one reserved event the kernel appends
  through no door. Door refusals and company refusals are recorded; a refused read is
  403 and not a fact (a read is not an act). `access.recorded` is the face's report of
  sign-ins through the builtin `record_access` door (who: adapter, installation-wide),
  the audit under the person. Both namespaces are kept out of the worklist.
- The kernel signs: with a person on the request, every string input named `by` or
  `…_by` carries the person's handle whatever the caller sent, and the event's actor is
  the handle. A door with a string input `company` must be acted for that company; a
  door with a string input `office` (grant_mandate) must be the office's person's too.
- Reads are gated by the same vocabulary: a view by its id (a list, detail or statement
  also by its type), objects by their type, and the system surfaces as named system
  views — worklist, gate (rules, bundles), files, channels, said, answer, learning,
  untaught — so `view:*`, `domain:system` and `*` grant them all and a role can name
  one. Nav, types, activities (with `allowed` per door for the person), time and the
  live stream are definitions, open to any signed-in person.
- `rhea serve -require-person` (or RHEA_REQUIRE_PERSON=1) answers 401 to a request
  naming no person, except the machine's own routes: the clock, the morning, the face's
  sign-in report, the time and live reads. On white01 the face is the only door; on the
  bench the owner's shell runs unscoped as before.
- The who-can-do-what view is a notion (`rights`, the pack's `who-can-do-what`, also
  `GET /api/rights?company=`): rows the persons with rights in a company, columns every
  door, view and type, each cell the grant that opens it — membership, mandate or
  substitution — with who granted it, from when, and the grant it rests on. The kernel's
  web shell renders it; the face renders it its own way.
- `rhea people register|grant|who` is the operator's hand at the first owner: nobody can
  invite the first person of an installation, so the operator does it from the machine
  through the same doors, signed cli:<user>. Everyone after that comes in through the
  face.
- Persons are the installation's rows (one human, many workplaces; a name and a handle
  are not a company's books); memberships, mandates, substitutions and groups are the
  company's. Owed: hiding persons without a membership in the scope when the first
  office with strangers' clients needs it (invariant 9, 2026-10-20).
- The Polish pack v35 names the new notions in its words; the formula language's
  `coalesce` reads a door's optional strings, and a grant with an optional last day is
  a rule pair (absent/exists), the pack's existing way.
- Tests: the vocabulary (core); the pack loads and its roles book (pack); grants through
  memberships, mandates, substitutions, dates, ending and a company's own role (rights);
  at the shell: a stranger refused, the grant signed by the owner, a person's facts the
  installation's, an accountant refused at the people doors with the refusal recorded,
  the clock refused to a person, allowed doors on the list, reads scoped and refused
  with the refusal recorded, a reader who reads only, the mandate's reach and its scope,
  the desk on all companies, the claim and the second claimant, the view's cells opening
  to the grant, replay identical with the refusals kept; persons-only serving.

## 2026-10-20 — Pack v36: inventories — what was sold, and stock written down

- The first of the author's block I drafts reviewed into the pack (lessons I-6 and I-29;
  EXAM.md 6.1.2 b, 6.1.3 d–e, 6.1.4 b–c). The author had drafted a separate document
  type for sales of materials; that duplicates the sales invoice, so the lesson became a
  field: a sales invoice says what was sold (`kind`: services, goods, materials, a fixed
  asset, other), and the revenue lands by kind — 700 for services as before and when
  nothing is said, 730 for goods, 760 for the rest (UoR art. 3 ust. 1 pkt 32: the sale
  of materials and fixed assets is other operating revenue). The value of what was sold
  comes off the shelf on the day of sale when the invoice states it: goods Wn 731 / Ma
  330 (art. 42 ust. 2), materials Wn 761 / Ma 310. The author's finished-goods write-down
  became a write-down of any stock: `write_down_inventory` to the net selling price (art.
  28 ust. 1 pkt 6, art. 34 ust. 5) against a correcting account 349, so stock stays at
  cost, and `reverse_inventory_write_down` when the cause ends (art. 35c) back to 760. A
  price above value is no write-down and books nothing. Accounts 731 and 349 join the
  chart; sales_invoice v8.
- Boundary checklist: reads the caller's scope only (doors and events carry the company);
  twice: the caller's dedup key, as every door; money: net, vat, gross, carrying value,
  the write-down and its reversal, all money fields, strings across the boundary.
- Parked from the same batch, with reasons: I-8 (work in progress) and I-19 (idle
  capacity) need production, which the pack does not have; I-11 (standard prices on
  goods) is a policy the pack does not use — goods at purchase price; I-5 (trade
  receivables always short-term) is presentation and belongs to the balance sheet lesson;
  I-12 (reclassification to investment property) waits for investment property as an
  object with its fair-value model; I-21, I-24, I-26, I-30 rest on IFRS and are not a
  small company's pack. The automatic cost method (FIFO, weighted average from receipts)
  is the next inventories lesson; this version takes the value from the document.
- Test: `TestInventories` — a service, goods bought for 600 sold for 900, materials sold,
  a laptop sold, each on its account; a seasonal stock written down 1 200 and the
  write-down reversed; a price above value refused. Pinned counts: 33 types, 32 views,
  98 rules, 23 doors, 60 events, 48 accounts.

## 2026-10-20 — Pack v37: a fixed asset under construction, and an advance for it

- Lessons I-3 and I-14 of the author's block I drafts (EXAM.md 6.1.2 a, 6.1.5 b). Two
  things the construction account must not keep: a cost that is not building — a fire's
  clean-up, idle capacity — and an exchange difference on an advance. `move_from_construction`
  names a cost on 080 that is not part of the asset's cost (UoR art. 28 ust. 8 says what
  the cost of construction is; ust. 3 keeps idle capacity and production losses out of a
  cost of manufacture) with its reason (casualty, idle capacity, other) and books it Wn
  761 / Ma 080 — other operating costs, where art. 3 ust. 1 pkt 32 lit. i puts the
  consequences of casualties. An advance paid for a fixed asset in a foreign currency is
  an `asset_advance`: its złoty are the złoty of the day it was paid (`pay_asset_advance`,
  Wn 249 / Ma 130), and when the asset arrives the advance settles at those same złoty
  (`settle_asset_advance`, Wn 080 / Ma 249) with no exchange difference — KSR 11 pkt 6.18:
  an advance for a fixed asset is a non-monetary item, translated once at payment and never
  revalued, at the balance date or at delivery. The author had the settlement revalue the
  advance at delivery and post the difference to 750/751; the standard says the opposite
  and the rule follows the standard.
- Boundary checklist: the caller's scope only; twice: the caller's dedup key as every
  door, and a settlement amends an advance by its id so the second settlement of the same
  advance is the same amendment; money: the amount moved, the advance in foreign currency
  and in złoty, strings across the boundary.
- Test: `TestConstructionAndAdvances` — two invoices on 080 through the supplier's cost
  account, 3 000 of a fire's clean-up moved to 761, an advance of 10 000 EUR paid at
  42 569 złoty and settled at 42 569 with 750 and 751 untouched. Pinned counts: 34 types,
  32 views, 103 rules, 26 doors, 60 events, 48 accounts.

## 2026-10-20 — An optional field copied from an input the door did not get stays unset

- Found writing the loans lesson: a door's optional input (a loan's term, its lender, its
  rate) copied into an optional field by naming it (`=$.due_date`) failed the whole rule
  when the person left the input out — "$.due_date: missing" — and the loan never
  materialized. The pack's way for strings was `coalesce($.number, "")`, a stand-in; a
  date, a ref or a decimal has nothing that could stand in. The v37 advance's optional
  supplier carried the same latent defect.
- The kernel's rule now: in an object template, an optional field copied from a path that
  is not there is left unset (`exec.Expand`, on `core.ErrMissing`); a required field from
  an absent input still fails, and a computation over an absent input fails whatever the
  field — nothing stands in for a value a formula needs. Determinism holds: the derived
  event carries the state as expanded, and replay expands the same.
- Test: `TestExpandLeavesAnOptionalFieldUnsetWhenItsInputIsAbsent` — four optional fields
  of four types unset from an absent input, set when given; the required field and the
  computation refused.
