# Direction notes

Where Rhea is headed beyond the SPEC milestone ladder, and what we learned
from Alpha. SPEC.md stays the design; this records intent so it survives
outside chat history. Written 2026-10-02.

## Two phases of the product story (KK, 2026-10-01)

**(a) Building the system is a conversation.** The user talks to the AI about
their business — the documents they collect, their chart of accounts, banks,
reconciliation, departments, cost centres — and the AI matches that against
its country knowledge (Polish CoA templates, VAT, KSeF…) and proposes
definitions: ObjectTypes, Rules, ViewDefs, eventually whole market packs
(M3), all as drafts behind the same human approval gate rules already have.
The AI "generates visuals during the conversation" without drawing anything:
it drafts a ViewDef and the generic shell renders it with sample data — the
mock-up *is* the artifact; approving it makes it live. Today's agent drafts
one rule from one sample event; phase (a) needs a multi-turn conversation
producing coherent *bundles*. That loop deserves its own design note before
building. Alpha's discovery-session screenplay (its doc 010) is the interview
protocol such an agent should run.

**(b) Operating the system is worklists.** Incoming events, decisions
waiting, notifications — "what do I do now?" next to "how are we doing?".
Rhea has the skeleton (worklist, approvals, simulation); Alpha's cases model
is the maturity target.

## What Alpha teaches (surveyed 2026-10-01)

Alpha (`MojoHand70/Alpha`) is a hand-built ERP for a real client — the
empirical counterpart to Rhea's language experiment. Worth porting in spirit:

- **Cases, not a worklist view** (Alpha doc 017 §5): the work item is an
  *object* — kind, subject, status, deadline, thread, and the one or two
  verbs that resolve it, inline. Raised by the system from events, never
  typed in; closes itself when the world moves on. "An empty list means the
  day's work is done."
- **Live screens** (017 §2): the projector is the only writer, so it is the
  only announcer — a notice per projection write (Postgres NOTIFY), one SSE
  stream per browser tab, rows patch in place. Nobody refreshes. Rhea's
  executor is already the single writer; the same mechanism drops in.
- **The agent is a user** (012): a role with a worklist and typed verbs,
  `actor = agent:<model>` on everything it causes, never on a read path,
  never deciding what a rule decides. Rhea's actor column follows this.
- **The shell** (016): omnibox (⌘K, `>` runs verbs by name), activity bar
  ordered by the life of a business object rather than by department, tabs
  as URLs, assistant panel with the current tab as context.

## The network: Rhea learns explanations (KK, 2026-10-03)

The endgame concept, adopted as direction. Rhea runs for thousands of
smaller and bigger EU businesses, free (monetization deliberately deferred —
KK), and becomes wiser by collecting and generalizing *rules*, never data.

- **What flows upward is explanations, not facts.** Rule-shapes, anonymized
  and parameterized (stripping instance literals is itself an AI task).
  Business data never leaves; structure does — GDPR-friendly by
  construction.
- **Packs are the distribution vehicle.** The network clusters rule-shapes
  by market and event shape; consensus patterns above a support threshold
  (never a single business's pattern — it may encode their secrets) promote
  into pack version bumps. Thousands of Polish bakeries book flour the same
  way; the pack learns it once.
- **The approval gate is the trust boundary.** A network suggestion arrives
  at every subscriber as a *draft*, simulated against their own history
  before approval. The mechanism that makes one business safe makes the
  network safe: collective wisdom, locally falsified.
- **The agent gets wiser two ways**: the pack library grows as explicit,
  auditable knowledge; and drafting gains priors from similar businesses —
  "3,400 like you explain this event shape this way; you deviate because
  you're cash-method."
- **The signals are already recorded**: approvals, rejections, edits of
  drafts (the richest — a human correcting the AI), and worklist residue.
  Residue across the network is demand: "7% of PL businesses receive events
  the pack cannot explain" is the pack roadmap writing itself.
- **The design principle this imposes now**: consensus over small
  declarative specs is tractable; over code it is hopeless. Everything that
  keeps rules tiny data — sub-languages instead of rule-level programs —
  is what makes the network effect computable. This retroactively justifies
  the arithmetic stance below.
- **Phase (a) is the collection mechanism**: every discovery conversation
  that ends in approved rules is labeled training data for the flywheel.
  Cold start: packs we author ourselves (begun 2026-10-03, pl + de).

One line: *Rhea never learns facts; it learns explanations — packs are the
distribution vehicle, the approval gate is the trust boundary.*

## The UX is a replaceable interpreter (KK, 2026-10-03)

Alpha's shell (activity bar, submenu, tabs, omnibox, cases) stays the
reference UX — but as *a* client, never the UI. The contract, in one
sentence: **a shell is anything that can render the notions and offer the
activities.** What can be seen and what can be done are both declared as
data; every surface — web shell, TUI, native, voice, the phase-(a) agent
itself — is an interpreter of the same two vocabularies.

- **Typed view API.** Today `/api/views` returns pre-formatted strings
  (money already through FormatMoney) — presentation leaking into the API.
  The declared contract carries semantics: typed values (money as minor
  units + currency), plus the notion spec; renderers decide formatting.
  Server-side formatting was experiment-phase safety, not the contract.
- **Default visualization is derived from the ObjectType** — fields, types,
  label_field and is_document already say enough for a serviceable list and
  detail with zero ViewDefs. Objects *have* their visualization, by
  derivation. ViewDefs become exceptions, authored only where seeing is
  genuinely a point of view: localization, role, emphasis. M4 is the proof
  views cannot live on the type: one base `sales_invoice`, Polish and
  German eyes on it.
- **The complaint question dissolves into four concepts the language
  already has.** *Registering* a complaint is an Activity (a declared verb
  with typed inputs emitting a raw event — what the omnibox runs by name);
  *that it happened* is the Event; *the complaint* is an Object — a
  document with a lifecycle (open → investigating → resolved, as
  projections); *that someone must act* is a **case**: another object,
  raised by a rule from the event, carrying deadline and resolving verbs,
  closing itself by projection when the world moves on. "An empty list
  means the day's work is done" becomes a property of the event log, not
  of a screen.
- **Cases need no kernel** — with cascade, lifecycle amendments and `each`,
  a case is a conventional object type plus rules, like `posting` and
  `period_lock` before it. That is the next falsifiability test in waiting:
  express Alpha's case model as pure data.
- Rough order of work: derived default views (kills ViewDef boilerplate),
  typed view API (the replaceability contract), Activities as data (the
  verbs — also what the agent needs to know what can be done), cases as
  pure data.

## Vocabulary governance: small grammar, open dictionary (KK, 2026-10-04)

The thesis depends on the language staying small while ERP is an ocean. The
resolution: the vocabulary has layers, and they grow by different rules —
like a natural language's closed class (prepositions: one per century) and
open class (nouns: daily).

- **The open class** — types, rules, views, packs, naming conventions — is
  pure data with unlimited growth, governed only by the approval gate.
  Richness lives here and needs no further safeguarding: three days in it
  held double-entry, warehousing and two statutory markets.
- **The closed class** — the six concepts, the template expressions, the
  sub-languages (postings, `each`, the coming costing and amendments), the
  notion library — grows rarely and only by proof. Admission requires all
  four: (1) **fail-as-data first** — a domain demonstrably inexpressible
  with the current vocabulary (postings, `each` and cascade each earned
  their place this way; M2 and M4 are the negative proofs where nothing was
  earned); (2) **no invariant, no primitive** — a primitive without an
  enforceable kernel law is sugar, and sugar is how languages rot;
  (3) a **determinism proof** — replay survives it; (4) the **network
  test** — small and declarative enough that thousands of uses cluster.
  The language earns primitives the way science earns laws: by exhausted
  attempts to do without them.
- **The grammar is append-only by physics, not policy.** The log must
  replay forever, so a primitive's semantics can never change once any log
  depends on it — only be superseded. Primitives have the rule lifecycle:
  experimental → active → superseded-but-honored-in-replay-forever.
  Deprecation means "stop authoring," never "remove."
- **Completeness has two yardsticks.** Theoretical: REA as a checklist, not
  a design (SPEC §7) — read against it, our one untested concept is the
  **commitment**: events that *should* happen (orders, reservations,
  budgets, production plans). Expect it to become unavoidable at M5, where
  plans are commitments on a time axis. Empirical: the network — worklist
  residue at scale is a completeness measurement, and **rule-shape
  contortions** (thousands of businesses independently using the same
  awkward encoding, as we carried precomputed values before deciding on
  arithmetic) are missing primitives announcing themselves. Gaps become
  data.
- **The agent is the gardener.** Dialect drift, not bloat, is the threat to
  the open class — a thousand names for the same thing blinds the network's
  clustering. Conventions are never enforced: the agent authors with the
  canonical ontology (seeded by the conventional names `posting`,
  `account`, `period_lock`), so they propagate through drafting. Deviation
  that spreads is itself a promotion signal.
- **Stewardship is the same gate, one level up.** Today: KK, DECISIONS.md
  and the falsifiability habit. At scale: the language gets its own
  worklist (fed by residue and contortions), candidate primitives are
  simulated against the corpus, and a human editorial layer — plausibly
  accountants and auditors, not programmers — approves. Rhea has never
  needed a second governance mechanism for anything; needing one here
  would be a design smell.

One line: *richness grows as data behind the gate we have; grammar grows
only by proof behind the same gate, one level up — and the log makes the
grammar append-only whether we like it or not.*

## What "big" means and the rules that keep us eligible (KK, 2026-10-04)

Adopted as positioning; KK's one-year target (2027-10) is for this picture
to hold end to end. Source: a conversation dissecting what makes S/4,
Fusion or D365 "big" — enterprise structure model, process breadth on one
data model, localization depth, encoded edge cases, controls and
auditability, ecosystem. Its sharpest findings: only SAP owns its statutory
substance (the rest built localization *plumbing* and outsourced the
content to partner overlays); edge cases split into *industry* ones
(encoded once, amortized across thousands of customers) and
*company-specific* ones (the implementation killers — configuration
becomes customization in disguise, every upgrade becomes a migration); and
the structural flaw: regulation and company specifics bolted onto a
monolith, and the bolts are where ERPs fail. Its closing sentence — "a
design where statutory rules are first-class, versioned components would
attack exactly that weakness" — describes Rhea without knowing it.

The rules of eligibility — each one maps an ingredient of "big" to our
mechanism, and each is already in SPEC or a direction note, which is itself
the finding:

1. **The kernel never learns a domain** (their process breadth). Their
   version: every operational event has its accounting consequence,
   hardcoded over decades. Ours: one language, the consequence as rule
   cascade with provenance. Fail-as-data first, forever — an
   invoice-specific branch in the kernel is the first day of a mid-market
   ERP.
2. **A market is a data file** (their localization depth — the real moat).
   Statutory substance and statutory *semantics* (per-line vs per-document
   VAT rounding, numbering) are pack data behind the adapter contract,
   never kernel code. The network maintains the substance as consensus
   rule-shapes behind every subscriber's own gate; worklist residue is a
   completeness gauge no big ERP can measure about itself.
3. **One extension mechanism, zero customization layer** (their edge
   cases). Industry edge cases are open-class data the network learns
   once; company-specific ones are the same kind of object as pack rules —
   versioned, simulated, replayed, upgrade-safe because the grammar is
   append-only by physics. In big ERP, custom is debt; in Rhea, custom is
   data with the same guarantees as standard. Never a second mechanism.
4. **The invariant layer is the audit story** (their controls — the one
   dimension where we are ahead, not chasing). Their audit trail is a
   feature; ours is the physics: append-only, provenance on every fact,
   approval as an event, period locks, backfill ruled, auditable by
   replay. Never traded for convenience.
5. **The network replaces the labour market's function** (their
   ecosystem). Buyers of big ERP partly purchase consultants who carry
   edge-case knowledge in their heads; phase (a) does the implementation
   interview, consensus rule-shapes carry the ISV knowledge as auditable
   data, the editorial layer is the new certification. If implementing
   Rhea ever requires code, the ecosystem reverts to the old labour
   market.
6. **AI authors, never executes** (the AiRP claim, below). Containment is
   what makes the claim auditable; drop it and the brand collapses into
   every other "AI ERP" press release.
7. **Enterprise structure arrives as data too** (their structure model —
   the gap, promoted to work below).

Positioning, two decisions:

- **"New era of ERP" means time-to-depth, not stack age.** The warning the
  claim must survive: a new architecture alone makes no challenger; the
  moat is accumulated business semantics, earned over decades. So "they
  are old" is true in one precise sense: their accumulation mechanism —
  semantics as code, welded into a monolith — cannot be sped up. Ours —
  semantics as data, at network speed — is the flywheel, and its rate is
  the metric. First evidence already on record: Germany as one data file,
  days after Poland, versus partner-overlay-years.
- **AiRP — AI Resource Planning.** Not an agent on top of old machinery;
  the ARC reactor. The reactor powers the suit but is not the suit's
  hands: AI is what makes a pure-metamodel ERP economically viable for the
  first time (SPEC §1) — without the AI the language is unaffordable to
  author (REA died of this), without the language the AI is unauditable.
  A copilot on SAP can only *operate* existing semantics; AiRP's AI
  *authors* semantics, and the network makes it collectively wiser. The
  reactor works because it is contained — kernel sole writer, model output
  as data, human gate — and the containment is what survives
  due-diligence.

**The gap, promoted from parked to work**: enterprise structure — multiple
legal entities, parallel GAAPs, intercompany, multi-currency. It
reconciles with the endgame line exactly: a GAAP *is* an explanation
system, so parallel accounting is two rule-books explaining the same event
log differently — "learns explanations, never facts" applied to accounting
standards. Intercompany is cascade across company refs; FX and its
statutory rounding land in the arithmetic sub-language. The named
falsifiability test, of the M2 kind: express parallel GAAPs as pure data
over one log, zero kernel changes.

One line: *big is accumulated business semantics — theirs accumulated as
code over decades, Rhea accumulates them as data at network speed, and the
seven rules are what keep that claim auditable.*

## The enterprise-structure ladder (KK, 2026-10-04)

Adopted as the plan for the gap named above: five stages, each an M2-style
falsifiability test, each proving one ingredient of the big-ERP structure
model. Dependencies are clean enough to interleave with the main ladder —
E1 can run before M5, and E3 is what finally forces the arithmetic
sub-language build.

- **E1 — parallel books.** Shipped 2026-10-04 (see DECISIONS): the
  data-only attempt failed on independent closes as predicted, `book`
  earned admission as an entry-level qualifier defaulting to "main",
  period locks are per (book, month) via intersected single-field
  lookups, and one generic trial-balance view splits per book. Two
  posting rule-books (PL statutory + group GAAP) over the same sale
  events: two trial balances, independent period locks both directions,
  a mixed event with one closed book refused whole, replay reproducing
  both ledgers. Proves: a GAAP is a rule-book; parallel accounting is
  native. The original expectation, kept for the record:
  postings resolve accounts by `code` against one flat population and the
  period lock is keyed by month alone — no book dimension anywhere. So E1
  is a genuine fail-as-data experiment: first attempt it with what exists
  (prefixed codes, per-book lock objects), and let the `book` qualifier
  *earn* admission into the postings sub-language — invariants balance per
  entry per book and lock per (book, month), determinism proof, trivially
  network-clusterable. The first closed-class change since cascade;
  deserves the `each` discipline — the failed data-only attempt goes to
  DECISIONS.md before the primitive lands.
- **E2 — the entity dimension.** Shipped 2026-10-04 (see DECISIONS): an
  M2-style negative proof — pure data, nothing earned. The attempt failed
  silently twice (first-approved registration rule captures the other
  market's master data; currency-as-market misroutes a Polish EUR invoice
  into the SKR03) and loudly once (both packs' "0" VAT rate is globally
  ambiguous). The fix is pack v2s: market where-guards on shared-type
  rules, the seller as a ref on `sales_invoice` v2, books per market so
  E1 is the per-company close, registers joining the seller's country,
  globally unique resolution keys. Two self companies, one kernel, one
  log, one replay. Un-parked "multi-market cohabitation in one ledger".
  Contortions carried with named exits: payload `market` dies when
  ref-reads arrive; book names cannot compose company × GAAP yet; a 0%
  invoice cannot post (conditional lines); adapters don't bind per market
  yet. The original plan, kept for the record: `company` objects go
  plural (`kind: self` more than once), events carry a company ref, packs
  bind per company — one company under the PL pack, one under DE. Proves:
  enterprise structure is refs, not tenancy. Depends on E1.
- **E3 — currency.** Shipped 2026-10-04 (see DECISIONS): the attempt
  failed as the predicted hollow-explanation contortion (precomputed PLN
  amounts on the event; the system's own rate table decorative), and the
  postings sub-language earned the `convert` clause — `{to, date,
  rounding, rounding_account}`, method vocabulary in kernel code, choice
  and parameters as rule data, no expression syntax anywhere. Rates are
  fx_rate master data from append-only events; conversion runs at firing
  time in integer arithmetic so replay never re-converts; the declared
  half_up residue books to a declared account as a visible plug line;
  identity conversion lets one rule explain domestic and foreign
  documents; a missing rate refuses into the worklist. The rounding
  stance and account are pl pack data (account 756). State reads arrived
  as core.Getter, shared by executor and simulator. Still open from the
  original plan: period-end revaluation as a rule (ties into M5
  scheduling), and ref-reads in match conditions (E2's market field still
  waits on them). The original expectation, kept for the record:
  foreign-currency events post in transaction and functional currency;
  FX rate tables (NBP, ECB) are master-data events, so every conversion
  is replay-deterministic by construction.
- **E4 — intercompany.** Shipped 2026-10-04 (see DECISIONS): another
  negative proof — pure data, nothing earned. The attempt failed on one
  seam (a ref cannot be re-referenced: direct ids rejected by design,
  value resolution hands vat_id an id), fixed by the recorded-contortion
  family: sales_invoice v3 echoes intragroup and the resolution keys,
  exit at ref-reads. Two group-level mirror rules — what phase (a)
  drafts for a subsidiary pair — raise the buyer's purchase document and
  its entry from the seller's invoice materialization, E3's convert
  composing with the cascade: Alfa books converted PLN, Beta books EUR,
  one raw event, all atomic. The killer demo is now a test assertion:
  every object on both sides provenance-chains to the same root event,
  so receivable and payable cannot disagree — nothing to reconcile, only
  to display. The original plan, kept for the record: cascade across
  company refs, provenance crossing the boundary, transfer pricing as
  the rule that prices the derived event. Depends on E2.
- **E5 — consolidation.** Shipped 2026-10-04 (see DECISIONS): no attempt
  ceremony — E5 asked nothing of the closed class. One read-side change
  (the DuckDB projection carries `root_event_id`: invariant 5 arriving
  whole on the analysis side) and two analysis ViewDefs:
  intercompany-positions (both sides matched by shared root, difference
  zero by construction — the reconciliation screen with nothing to
  reconcile) and the group trial balance in EUR (translation at the
  latest rate, eliminations by provenance except tax-typed accounts, the
  transaction-vs-closing residue as a visible CTA row). The write-side
  form — elimination entries in a group book — waits on book
  composition; statutory translation methods are pack depth. The
  original plan, kept for the record: the group is one more explanation,
  a consolidation rule-book whose eliminations match on intercompany
  provenance, starting on the analysis side. Depends on E4.

**The ladder is complete** (2026-10-04, one day end to end): E1 earned
`book`, E3 earned `convert`; E2, E4 and E5 passed as pure data — two
primitives, three negative proofs, the grammar still small. The gap named
in "What 'big' means" above is closed: parallel books, entities as refs,
functional currency, intercompany matching by construction, and the group
as one more explanation — one kernel, one log, one replay.

One line: *the enterprise structure model is five proofs — book, entity,
currency, boundary-crossing cause, group-as-explanation — each pure data
unless it earns a primitive by the governance rules.*

## Humans buy with eyes: the design language (KK, 2026-10-05)

Adopted as direction. Rillet, Light, Pigment, Airtable win deals on
beautiful, consistent UI; concept-first positioning does not exempt the
demo from this — the demo is how the concept is seen. We are not designers,
so the requirement is not beauty now but a **foundation that external
design experts can restyle later without touching behavior**.

- **The mechanism is the thesis, one level down.** The replaceable-
  interpreter contract already separates what can be seen from how it is
  shown; the design language is the same split inside the web interpreter:
  a **token layer** (color, type scale, spacing, radii, elevation, density,
  layout sizes — CSS custom properties, the lingua franca design tooling
  exports) and a **closed set of surface primitives** (shell chrome:
  activity bar, submenu, tabs, toast; notion surfaces: data table, kv
  detail, status chip, provenance line, forms, diff). Presentation is data.
- **The generic shell is a design-system multiplier.** Because the shell
  renders notions generically, the primitive set is closed and small:
  styling it once styles every screen that will ever exist, including
  ViewDefs the AI drafts in phase (a). A bespoke-screens app needs a design
  system as discipline; Rhea gets it by construction.
- **Rules**: no color or size literal outside `tokens.css`; tokens come in
  two layers (primitive palette → semantic aliases) so a rebrand is a
  palette swap and a redesign is an alias remap; no framework, no build
  step (vanilla custom properties are exactly what designers' token tools
  emit); semantic class names per primitive. The light color scheme ships
  as a pure semantic remap — the first "external restyle" is us, proving
  the layer works.
- **Hiring a designer later** means handing over `tokens.css` and the
  primitive inventory (documented in its header), not the app.
- **Visual order of work** (2026-10-05 conversation): token foundation
  first; then typed view API + derived default views — the *semantic* half
  of the design language (money, refs, statuses know what they are; the
  renderer decides how they look); then the provenance walk ("why?" on
  every fact — the signature interaction no other ERP can offer); live
  screens via NOTIFY/SSE; readable field-level simulation diffs; omnibox +
  activities-as-data. Polish is whatever makes the mechanism visible;
  nothing invoice-specific, ever.
- **Shipped 2026-10-05** (see DECISIONS): tokens, then the typed view API +
  derived default views. Columns declare `{field, label, type}`, cells cross
  as `{v, id?}` in canonical encoding, provenance is data, the shell formats
  money per browser locale — Polish and English eyes now literally see the
  same response differently, M4's lesson landing in the renderer. Objects
  have their visualization by derivation (period_lock got its first screen
  without anyone authoring one), and the analysis specs lost every `printf`:
  SQL emits data, columns say what it means. Next on the ladder: the
  provenance walk.
- **Shipped 2026-10-05, the provenance walk** (see DECISIONS): any object,
  any detail, any event-typed analysis cell opens the chain — the root raw
  fact and its whole consequence tree, every hop naming the exact rule
  version, every object a door to its detail, the asked-about chain
  highlighted. Both sides of an intercompany position walk to the same root
  on screen now, not just in a test. Refs became doors (`list → ref → detail
  → walk` closes the loop). Next: live screens via NOTIFY/SSE, then readable
  simulation diffs.
- **Shipped 2026-10-05, live screens** (see DECISIONS): the executor
  announces every committed booking from inside the transaction, one SSE
  stream per browser tab carries the notices, and live tabs — views, the
  Language catalog, the walk — re-render in place. Tabs holding human state
  stay manual. Submit an event in one pane and watch the invoice, its
  postings and the trial balance appear unrefreshed. Next: readable
  field-level simulation diffs, then omnibox + activities-as-data.

One line: *the design language is tokens plus a closed set of notion
primitives — the generic shell makes beauty a data change, so experts can
restyle without touching behavior.*

## Language decisions with a recorded destination

- **Rule cascade.** KK's call (2026-10-02): rules matching *derived* events,
  with atomicity, is the intended end-state (receipt → stock move →
  valuation posting). Shipped 2026-10-03 (see DECISIONS): cause-qualified
  rooted identity, chain-visible refs, per-line cascade valuation; loops hit
  a depth-cap error naming the looping rule.
- **Arithmetic is a language problem; effects compute in the kernel**
  (agreed 2026-10-03). Effect arithmetic runs at firing time and is baked
  into derived events — that is where provenance and replay-fixity live;
  a number computed in a database at read time explains nothing. DuckDB
  keeps *analysis* arithmetic only. What's missing is notation as data:
  operators, reads through refs (`item.std_cost` — refs resolve to ids only
  today), and an explicit rounding stance — division is where determinism
  dies, and rounding is itself statutory (per-line vs per-document VAT
  rounding differs by country: pack data). Algorithmics — FIFO valuation,
  production allocation, stock counting — arrive as *kernel sub-languages
  parameterized by rules* (the postings and `each` precedent):
  deterministic method vocabulary in code, choice and parameters as data.
  The AI composes sub-languages; it never authors loops.
- **Lifecycle: status is a projection, never an update** (KK, 2026-10-03:
  every object has a life, and month-end status must be answerable). An
  invoice becomes "paid" because a payment event, matched by a rule, emits
  a derived amendment referencing it — append-only untouched, provenance
  per amendment. ObjectTypes declare status fields and allowed transitions
  as data; the kernel validates transitions like it validates balance (the
  invariant layer grows). As-of is replay with a cutoff — "what status at
  month-end" is a parameter on machinery determinism already paid for.
  Amendments carry business dates and respect period locks.
- **Backfill is ruled** (KK, 2026-10-03: "a closed month is a closed
  month"). Backfill is the promotion of a simulation diff into the log,
  as an explicit approval-gated activity recorded as an event — the log
  stays honest about when understanding arrived vs when facts occurred.
  A backfill firing into a locked period is refused per event, no
  exceptions; the open-period correction (korekta) is the human
  alternative. Nothing retroactive ever happens silently.
- **Conflicts stay human.** Same-object-id claims refuse the event into the
  worklist; semantic double-booking is simulation's job to reveal before
  approval; static conflict detection is parked (SPEC §7).
- **Authz belongs in the language, eventually.** Users, groups and access
  policies should be objects and rules like everything else — a good
  falsifiability test of its own, post-M2. Until then: actor attribution
  only, no auth (experiment non-goal).
- **Multi-company is a ref, multi-tenant is infrastructure.** The operating
  company will be a `company` object (`kind: self`) that events reference;
  tenancy stays out of the experiment. Promoted (KK, 2026-10-04): enterprise
  structure is the named gap against big ERP and now has a destination —
  parallel GAAPs as parallel rule-books explaining one event log,
  intercompany as cascade across company refs, FX in the arithmetic
  sub-language. A falsifiability test of the M2 kind: zero kernel changes.
  Planned (2026-10-04): the enterprise-structure ladder above, E1–E5.
