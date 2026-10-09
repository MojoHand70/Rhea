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
