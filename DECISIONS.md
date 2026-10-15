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
