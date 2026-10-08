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

### Rhea suggests standards (KK, 2026-10-08)

Asked "how do you book a PZ?", KK answered past the question: those
questions have no single answer ("it depends", "everyone does it a bit
differently"), and **the role of Rhea is to suggest standards.** If Rhea
knows how something is booked, she proposes it, and the client accepts or
gives their own rule. Over time she learns from thousands of clients, from
the web, from documents and from scientific work, until she can say:
*"Listen, 90% of companies in Poland in this industry do it like this."*
That never obliges the client.

- **Rhea leads, the client decides.** The agent stops asking open
  questions it could answer. It proposes the standard as a draft, says
  why, and asks for a yes or the client's own rule. The approval gate is
  unchanged, and only who speaks first changes. This is the host vision
  applied to implementation.
- **Knowledge has many sources, and each is named.** Statute and
  regulation, accounting standards (UoR, KSR, IFRS), textbooks and
  research, documents, the web, pack authors, and the network's
  consensus over approved rules. They differ in kind and in strength, so
  they are never blended anonymously.
- **Every suggestion carries its warrant**, as every fact carries its
  provenance: the source kind, the citation, the scope (market,
  industry, size) and, for the network, the actual support ("2,140 of
  2,380 Polish wholesalers"). *A support figure is only ever a count of
  real approved rules, never estimated or invented.* Before the network
  exists, the honest warrant is the standard itself ("customary under
  the Polish accounting act; Poland pack default"). A fabricated "90%"
  would spend the trust the whole design is built to earn.
- **The client's own rule wins, and teaches.** An override is an
  ordinary approved rule for that client. The deviation and its reason
  ("we value at standard cost") are the richest signal the network gets,
  and they may become a minority standard of their own once enough
  clients share them.
- **The industry is part of the scope.** "In this industry" needs the
  client's profile (PKD code, size, accounting method) as data the
  consensus is scoped by. It is collected in the interview like
  everything else.

One line: *Rhea speaks first with a standard and its warrant; the client
answers with a yes or their own rule; both are explanations, and the
network learns from the difference.*

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
  express Alpha's case model as pure data. **Passed 2026-10-05** (see
  DECISIONS): the attempt failed exactly at "closes itself" — identity
  physics forbids touching an existing object, and the closure marker left
  the status lying — which admitted the amendment (`effect.amend`, consent
  via declared lifecycles, transitions validated like balance); on top of
  it the whole model is types + rules + activities, rendered with zero
  case-specific screens.
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
  plans are commitments on a time axis. **Tested at M5 (2026-10-05, see
  DECISIONS): the commitment stayed pure data — a schedule is an object
  with a lifecycle plus the clock adapter plus two rules; the closed class
  earned nothing. Production-depth plans will re-falsify on the same
  mechanism when that domain arrives.** Empirical: the network — worklist
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
- **Shipped 2026-10-05, readable simulation diffs** (see DECISIONS): the
  approval gate reads field-level, typed changes — before → after with the
  old value struck, refs resolved against the hypothetical world, the newly
  explained events named, changed and removed objects linking to themselves
  as they are today. The visual ladder's remaining rung: omnibox +
  activities-as-data, which waits on Activities being declared in the
  language (the replaceable-interpreter contract's second vocabulary).
- **Shipped 2026-10-05, omnibox + activities as data** (see DECISIONS) —
  the ladder's last rung, and the second vocabulary declared (section
  below). The shell offers verbs the way it renders notions: a Verbs
  catalog under Language (status, who, emits, events through each door),
  trigger forms derived from the declaration the way list/detail derive
  from the ObjectType, contextual verbs on every detail whose type a ref
  input names, and ⌘K jumping to any view or running any verb (`>` filters
  to verbs). The worklist and the provenance walk name each raw event's
  door — the walk starts one hop earlier than the rules.

One line: *the design language is tokens plus a closed set of notion
primitives — the generic shell makes beauty a data change, so experts can
restyle without touching behavior.*

## Activities as data: the second vocabulary (KK, 2026-10-05)

Adopted as direction, from a design pass over the open questions carried
in from another conversation. Activity has been one of the six concepts
since the founding spec (§2: inputs, the raw event it emits, who may
trigger), so this is **not a closed-class admission** — it materializes a
declared concept, and the admission ceremony applies to its *shape*, via
the retrofit test below. The omnibox rung and the phase-(a) agent's "what
can be done" vocabulary both wait on exactly this.

- **A door, never the door.** SPEC §2 already names `submit_event` as an
  M0 activity: the free-form submit *is* the universal, degenerate
  activity, and it can never be removed — fail-as-data requires
  submitting what nobody declared yet, and the network measures
  completeness by worklist residue; close the open door and the gauge
  goes blind. A declared activity is a named, typed specialization
  emitting a raw event of a declared type. Adapters remain the third
  door (code behind the contract), never through activities.
- **The schema belongs to the verb, not the event.** Inputs are
  `FieldDef`s (no second type vocabulary), validated at trigger time —
  before the event exists. The emitted event is a raw event like any
  other: schemaless in the log, still needing a rule to explain it,
  indistinguishable from a free-form submit of the same type except for
  the door stamp. ActivityDefs are never consulted during replay; they
  are not in the determinism path at all. ObjectType schemas govern
  derived state (kernel-enforced); activity input schemas govern the
  form a human fills (door-enforced courtesy).
- **Provenance symmetry is the invariant that earns the keep** (no
  invariant, no primitive). Derived events carry `(rule_id,
  rule_version)`; raw events emitted through an activity carry
  `(activity_id, activity_version)` — every event names its door, from
  day one (KK, this conversation: retrofitting provenance later is the
  kind of hole replay never forgives). Invariant 2 becomes literally
  checkable (`rule.approved`'s door is `approve_rule`, who: human), and
  the provenance walk starts one hop earlier: which declared verb
  created this raw fact — a question no other ERP can answer about its
  raw side.
- **Rule lifecycle, with one semantic note.** draft → approved → active
  → superseded, the same gate as everything; for an activity, `active`
  means *offered and triggerable*, not *executes in replay* —
  superseding one never touches the events it emitted. `rhea init`
  seeds the builtins and records their activation events: the gate's
  own birth is in the log. Phase (a) requires the gate anyway — drafted
  bundles include activities ("register a complaint" is one).
- **The retrofit falsifiability test**, run over all three shell verbs
  (KK, this conversation: `draft_rule` is in), predictions kept for the
  record. `approve_rule` nearly carries already: the handler appends
  raw `rule.approved` *first*, then inserts the version row — the
  retrofit moves the flip into an invariant-layer kernel reaction to
  the event; the `rule_id` input is a string contortion (FieldDef refs
  point at ObjectTypes, not definitions), exit named at
  ref-to-definition. `submit_event` strains honestly — an open
  event-type input plus a schemaless payload — and the passthrough
  form is the answer: declaring the open gate as data makes the escape
  hatch itself versioned and auditable rather than ambient.
  `draft_rule` is the predicted negative case: SPEC M0 never listed
  it, it emits no event today, and it is not an event-emitter by
  nature but a request to another actor. Prediction: the shape refuses
  it, and the refusal is the finding — either drafting routes to the
  phase-(a) agent-as-user design, or the test forces
  `rule.draft_requested` consumed by the agent from its own worklist,
  and agent-as-user arrives earlier than planned. Either way,
  DECISIONS.md records which.
- **What cases will demand of the shape**: ref-typed inputs and
  `list<ref>` (the reconciliation specimen's partial matches — one
  operation settles several documents and vice versa), plus the
  surfacing convention: a shell offers an activity on an object when
  the activity declares a ref input of that object's type — how "the
  one or two verbs that resolve it, inline" renders generically. The
  guard smell is flagged, not designed: "resolving verbs only while the
  case is open" must reuse the rule predicate vocabulary or stay out
  (transition validation refuses the emitted event) — never a second
  condition mechanism.
- **`who` stays inert**: actor-pattern strings (`human`, `agent:*`),
  displayed, never a Go enum with semantics, nothing the kernel acts
  on. Actor attribution on every event is the real mechanism today;
  when authz-in-the-language runs its own falsifiability test
  (post-M2), policies-as-objects interpret the same strings.

One line: *an activity is a declared door — typed at the verb, schemaless
in the log, stamped on every event it emits — and the open door is itself
the first declared activity.*

**Shipped 2026-10-05** (see DECISIONS): the retrofit ran over all three
verbs with the predicted findings — approve_rule's flip became a kernel
reaction (and turned atomic in passing), submit_event became the declared
passthrough, draft_rule's declaration carried while its fulfillment
located the boundary exactly: reactions are deterministic, so the agent
cannot be one; the ask is recorded as rule.draft_requested and consumed
outside the kernel. approve_activity closed the governance loop, init
records the gate's own birth, and the door stamp landed day one.

## Implementation by interview: the phase-(a) design note (KK, 2026-10-07)

The note "Two phases" above asked for before phase (a) gets built. Its seed
is KK's, verbatim:

> Implement by Interview — not a configuration, it's corpus acquisition —
> that gives you something no classic implementation has ever had: a
> measurable definition of done.

A classic implementation ends when the consultants' budget does. Rhea's
can end on a number the physics already computes: **worklist residue**. An
implementation is the client's real events, submitted as facts, being
explained by approved rules until nothing is left unexplained except what
was explicitly accepted. "Are we live?" stops being an opinion. The same
number, across clients, is the network's flywheel rate.

Status: **adopted 2026-10-08.** KK settled the four open questions
below, each on the proposed option.

- **Evidence first, answers second.** The interview starts by intake, not
  by questions: the client's last months of documents (KSeF invoices,
  bank statements, the CoA export) enter as raw events through the open
  door. They are facts, so they belong in the log. They land in the
  worklist as fail-as-data, which is exactly the state the interview
  starts from. Nothing new is needed: this is M0's demo story at corpus
  scale.
- **Residue drives the conversation.** The agent clusters the worklist
  by event shape and asks about the largest unexplained cluster first.
  Every question has a measurable payoff: the share of the corpus it
  would explain. The market pack goes first (its drafts, simulated
  against the corpus, usually explain the bulk); the interview is about
  the remainder, which is the company-specific edge cases that eligibility
  rule 3 says are the implementation killers.
- **The bundle is the unit of approval.** One answer ("we book cash sales
  from the till daily, by VAT rate") usually needs a type, a rule or
  three, maybe a view. The agent proposes them together; the simulation
  diff shows the whole bundle's effect on the corpus; one approval
  activates the bundle or nothing. Members stay ordinary versioned
  definitions, each with its own provenance, so the bundle is only the
  approval's scope, not a new kind of definition.
- **Accepted residue is explained too.** "Ignore this" must not be a hole
  in invariant 5. Accepting a residue cluster is a rule like any other: it
  materializes an `acknowledged` object (event, reason, who decided)
  through a conventional type. It is pure data, and the residue counter
  can tell "explained" from "explained as deliberately out of scope".
- **The conversation is in the log.** Each turn, human and agent alike,
  is a raw event through a declared activity (`interview.said`). That
  gives the agent continuity (the host vision's "it never asks twice" and
  "it doesn't repeat itself"), gives every drafted definition a cause one
  hop back (the turn that asked for it), and makes every finished
  interview labeled training data for the network: question, evidence,
  approved explanation.
- **The agent stays outside the kernel**, exactly where the activities
  retrofit located it: it consumes the log (turns, worklist, simulation
  results) through read-only tools and returns strict JSON drafts the
  kernel validates. Tool use is reading, never writing; invariant 3 is
  unchanged. Its runner is the shell, as with `rule.draft_requested`,
  until agent-as-user gets its own worklist.
- **Done is a declared state, not a feeling.** Implementation closes when
  residue on the intake corpus is zero, counting acknowledged events as
  explained. Ongoing operation inherits the same gauge: new residue is
  the host's morning question ("2 are strange and need your decision").

**The gap this exposes.** Rules and activities have the draft → active
lifecycle; `object_type` and `view_def` rows do not. Packs install them
directly, which was harmless while a human wrote every pack. A bundle
drafted by the agent cannot bypass the gate, so types and views need the
same lifecycle (a status column and the same approval reaction). That
extends invariant 2 to every definition; it adds no new primitive.

**Settled questions (KK, 2026-10-08).**

1. *Everything drafts.* Pack-shipped types and views land as drafts like
   pack rules and activities, and installing a pack is one bundle
   approval. Invariant 2 covers every definition, with no second path.
2. *Intake goes into the real log.* The client's documents are facts and
   start their system of record. No sandbox and no promotion step.
3. *No partial approval.* A bundle activates whole or not at all. A
   rejection carries a reason and the agent redrafts, so the correction
   stays on the record as the richest network signal.
4. *Residue-first protocol.* Intake → pack baseline → largest unexplained
   cluster → views and activities last. Alpha's doc 010 supplies the
   tone and the questions, not the order.

**Order of work.**

1. *Teach the agent the whole closed class.* Today it drafts one
   object-or-postings rule from one event. It should also draft cascade
   (derived event types), amend, `convert`, `book`, `each`, object types
   with lifecycles, activities and views, all validated by the existing
   gates. This does not depend on the open questions.
2. *An eval harness.* A fixed corpus (raw events plus intents) runs the
   agent's drafts through simulation and reports explained and residual
   counts per event type. This makes "the AI authors" a number that can
   be tracked from release to release, and it is the interview's
   definition of done in miniature.
   *Shipped 2026-10-07* with step 1 (see DECISIONS): first live score
   9/9 on the M0 story, too easy to discriminate.
3. *The bundle* (inserted 2026-10-08): types and views under the gate,
   one approval activating a set of definitions, a dry run over the set.
   The eval surfaced the forcing case: one event with two consequences
   cannot be implemented one approval at a time. *Shipped 2026-10-08*
   (see DECISIONS): packs install as one bundle, and the forcing case
   passes. The agent proposing bundles moves to step 6.
4. *A harder corpus* (convert, book, amend, `each`, the two-consequence
   case) so the score can tell models and prompts apart. *Shipped
   2026-10-08* with the agent proposing bundles (see DECISIONS). It does
   discriminate, and it surfaced two language decisions for KK: ruled
   backfill (late understanding of an already-explained event) and ref
   re-reference in cascades.
5. *The live full-story demo* with KK, showing a bundle come to life.
6. *The interview loop*: turns as events, residue clustering, the agent
   proposing bundles.

One line: *an implementation is a corpus being explained; the interview
is how, residue is when, and the gate is who.*

## Late understanding: backfill, enrichment, evolution (KK, 2026-10-08)

KK's question, after approving the Poland pack: nobody can settle a
definition up front. Should a bank account carry the IBAN separately, or
the bank's country? It is unknowable today, the same holds for every
other definition, and it will hold for every client. Can definitions be
amended later, extended, reduced?

The answer is the experiment's oldest invariant: nothing approved is
final, it is version 1. Types have already moved this way (invoice v1→v2,
posting v1→v3, sales_invoice v1→v3, the Poland pack at v6), and every
object records the type version it was built under. A schema change is
a bundle: new versions, a dry run against the client's real history,
then approval. In a classic ERP the data model is settled up front
because changing it later is a migration. In Rhea it is a reviewed
proposal. That is a selling point, and it is the same mechanism that
delivers pack updates to every subscriber.

What is missing is narrower and has one shape: **understanding that
arrives after the facts.** The eval found it first: an approved
withdrawal rule that never fires, because the withdrawal was already
explained another way. There are three variants.

Status: **adopted 2026-10-08.** KK agreed to every proposal; the settled
questions are at the end, with question 5 revised after KK asked "can we
add to the past, when the past is closed?".

**1. Late rule: ruled backfill.** The facts are old and the explanation
is new. DIRECTION fixed the stance on 2026-10-03: backfill is the
promotion of a simulation diff into the log, approved like everything
else, and a locked month is a locked month. The proposal makes it
buildable:

- *Additive only.* Backfill fires (event, rule) pairs that never fired:
  a rule approved after the event, whose derived events carry no
  rooted chain from it. The log can decide this, since every derived
  event names its cause and its rule. It never rewrites an existing
  derivation. Correcting a wrong past explanation stays the human
  korekta in an open period, as decided. This covers the forcing case
  (the case resolves on the old withdrawal) without touching identity
  physics.
- *A deliberate act, sibling to the clock's catch-up.* Approving a
  bundle changes the future. When its dry run shows that the bundle
  would also explain old events, the shell offers a second, separate
  gated act: `approve_backfill` emits `backfill.approved` with the
  pairs it promotes, and the kernel's reaction fires exactly those.
  Steady state stays automatic, and rewriting the past needs a human,
  the same physics as re-entry after a clock gap.
- *The log never closes; periods do.* Backfill does not change what
  happened. It records today a new understanding of an old fact: a
  derived event with today's `recorded_at`, caused by the old event.
  Periods are accounting periods, locked per (book, month), and they
  govern postings. A month-end close freezes what the books say about
  September. It does not freeze whether a September complaint's case is
  resolved.
- *Dated by the fact while its period is open.* A backfilled
  consequence carries the original event's business date, so it lands
  where it belongs.
- *Never into a closed period; offered forward instead.* A pair that
  would post into a locked (book, month) is not written there. The dry
  run proposes it for the first open period, dated on the backfill and
  linked to its original event: the korekta, or prior-period adjustment,
  that accounting already practices. The human approves the re-dating
  explicitly in the same act. The ledger stays closed, and the walk still
  shows which September fact caused the October entry.
- *Replay stays boring.* Backfilled derived events are ordinary derived
  events, later in the log. Replay re-projects them in log order, and
  invariant 4 holds by construction.

**2. Late fact: enrichment.** The thing is old and the information is
new. The existing bank accounts get their IBANs; no old event ever
carried them, so backfill cannot help. This is a new fact and enters as
one (`bank_account.details_provided`), explained by a rule that sets
fields on an existing object: the amendment, which already exists.
What does not fit is its consent. Today a type that declares a lifecycle
lets rules set *any* of its fields, and a type without one allows none.
Consent is all or nothing, and it is spelled as a status machine.

- *Proposal: field-level consent.* The type declares `amendable:
  ["iban", "bank_country"]`, the fields rules may set after
  materialization. The lifecycle field stays implicitly amendable and
  keeps its transition law. Everything else is fixed at birth, which
  matches how the business treats it: an invoice number never changes,
  an IBAN legitimately does. This is ObjectType vocabulary, not a
  closed-class admission: it refines the amendment's consent, the third
  primitive's own shape.
- *Every enrichment is a fact with provenance.* The walk shows "IBAN
  set by `bank_account.details_provided` #812 through rule
  `enrich-bank-account` v1", never a silent edit.

**3. Late schema: evolution.** The language itself grows. This already
works, and the proposal turns practice into stance:

- *A type version never rewrites old objects.* They keep the version
  they were born under. Views render them through the newest version,
  and a field missing from an older version shows as "not recorded
  under v1" rather than blank. Old objects gain new fields only by
  enrichment (a new fact) or backfill (a new explanation), never by
  migration.
- *Subtracting never deletes.* A field dropped in v2 stays in history
  and in the walk. Rules still setting it break, and the bundle's dry
  run shows it before approval, because the overlay re-validates the
  active rules against the drafted type. The type change and its rule
  redrafts travel in one bundle.
- *There is no rename.* A rename is a new field plus enrichment or
  backfill, so the log says what happened.

**The attempt, predictions kept for the record.** A bank-account story
as a falsifiability test of the M2 kind:

1. `bank_account` v1 (number, bank name), three accounts from
   `bank_account.opened` events.
2. A bundle raises it to v2 with optional `iban` and `bank_country`. The
   old accounts keep v1 and render "not recorded under v1".
3. Three `bank_account.details_provided` events arrive, and an
   enrichment rule tries to set the IBANs. **Predicted fail:** the
   amendment is refused (no lifecycle, no consent), which admits
   `amendable`.
4. The operations corpus's withdrawal task, rerun. **Predicted fail:**
   the approved rule never sees the explained event, which builds the
   backfill door. After both, the corpus reaches DONE.
5. Replay reproduces identical state throughout, and the walk shows
   every IBAN's provenance.

**Settled questions (KK, 2026-10-08).**

1. *Backfill is additive only*: rules fire on past events they never
   saw. Corrections of past derivations stay the human korekta;
   rewriting a derivation would need object identity to supersede
   itself.
2. *Backfill is a separate deliberate act* after the bundle's approval,
   the sibling of the clock's catch-up, never an option hidden in the
   approval.
3. *Enrichment consent is an explicit field list* on the type
   (`amendable`). Existing types tighten: the case type declares
   `resolution`.
4. *Old objects stay at their type version forever.* New fields reach
   them only by enrichment or backfill.
5. *Dating (revised after KK's "can we add to the past, when the past is
   closed?")*: in an open period, a backfilled consequence is dated at
   the original event. In a closed period it is never written there; it
   is offered for the first open period, dated on the backfill and
   linked to its cause, as an explicit choice in the same approval.
   Non-financial consequences (cases, statuses, enrichment) are not
   governed by period locks unless a later decision brings some under
   them. The amendment's exemption from locks stays a named exit.

**Shipped 2026-10-08** (see DECISIONS): the attempt failed as predicted
on enrichment and backfill, and evolution carried. Enrichment earned
`amendable`. Backfill is a door with a dry run, additive, offered forward
past closed periods, per chain in v1. The operations corpus reached DONE
in two of three live runs.

Ref re-reference, out of scope here, was decided the same day (KK,
option A: a follow-up points at exactly what caused it). It shipped as
carried links plus lookups one step through a link (see DECISIONS).

One line: *nothing approved is final; late understanding enters as a new
explanation (backfill), a new fact (enrichment) or a new version
(evolution), each gated, each in the log, none rewriting history.*

## Arithmetic: a formula language of our own (KK, 2026-10-08)

KK's call, after the PZ showed that value = quantity × price has no way
in: approach arithmetic as sfmt and Swan did, with our own tokenizer,
parser, formula compiler and a VM that runs compiled formulas. It gets
**a dedicated session.** This section is the brief that session starts
from. It records constraints, not a design.

**Prior art to read first.** sfmt's engine (`~/projects/sfmt/pkg/engine`):
`070-formula_vm.go` (byte opcodes, a stack VM), `072-formula_compiler.go`
(and its tests), `073-formula_aggregation.go`, `074-formula_sql.go` (the
same formula compiled to SQL), with `docs/FORMULA_INTELLISENSE_IMPLEMENTATION.md`
and `docs/calc-engine-cleanup-audit.md` alongside. Swan
(`~/projects/Swan`) carries formulas as element data (`formulaText`,
derived class C), which is the "formula as data, compiled by the kernel"
shape Rhea needs.

**What Rhea's invariants demand of it.** These differ from sfmt:

1. *No floats, ever* (invariant 6). sfmt's VM computes in `float64`;
   Rhea's cannot. Money is int64 minor units, quantities are integers or
   fixed-point decimals, rates are exact decimals (`fx_rate` is already
   a string). Port the architecture, not the number core.
2. *Rounding is declared, never implied.* Division and multiplication
   by rates are where determinism dies (DIRECTION, 2026-10-03). Every
   operation that can lose precision names its rounding (`half_up`
   today, others joining by proof), as `convert` already does. Rounding
   is statutory and therefore pack data: per-line versus per-document
   VAT rounding differs by country.
3. *Computed at firing, baked into the event.* The VM runs inside rule
   expansion. Results are written into the derived event, so replay
   never recomputes and invariant 4 holds by construction. DuckDB keeps
   analysis arithmetic only; sfmt's formula→SQL path is the natural fit
   for that side.
4. *Formulas are data, small and comparable.* They live in rule
   templates (an `=` template kind), are validated and type-checked at
   draft time against the fields they read and write (money × int →
   money, money ÷ money → rate), and stay tiny enough for the network
   to cluster. The AI composes formulas; it never authors loops
   (algorithmics such as FIFO stay kernel sub-languages).
5. *Every value is explained.* Invariant 5 extends to computed values:
   the walk should be able to show which inputs and which formula
   produced a number.
6. *Reads through links come with it.* The PZ also showed the agent
   reaching for "the price of the line's item" and "the PZ's warehouse":
   field reads through a link (`item.std_cost`). Name resolution is the
   parser's job, so it belongs in the same design.

**First customers, to test it against:** PZ line value (qty × price,
rounded), the PZ entry (Wn 330 / Ma 300 at that value, KK to confirm the
policy), VAT from net × rate with statutory rounding, and due dates
(date + payment term days, named by the cases test).

**Not questions for KK** (corrected 2026-10-08, see "Rhea suggests
standards"): how a PZ is valued and booked has many answers, and Rhea's
job is to propose the standard with its warrant. The Polish default (PZ
at purchase price, Wn 330 / Ma 300, settled against the invoice 300 /
202 with input VAT) arrives as a pack draft any client may replace. The
formula language must serve all the common variants: the delivery
states the value, a price list computes it, or the value comes later
from the purchase invoice.

**Iteration: folds over named collections, never loops (KK, agreed
2026-10-08).** KK's question: goods counting and production scheduling
will need looping, the simplest form being "for all order lines" — can
the formula language avoid it? Agreed: iteration cannot be avoided, and
general loops stay out. The distinction is *bounded iteration over a
named collection* versus *a loop that runs until a condition holds*.
The first is total — it always terminates and costs the size of the
collection; the second is what makes AL and ABAP customizations slow
and unauditable. sfmt is the mirror: its formula text has no loops, yet
its aggregation VM walks a whole dimension to sum it. Rhea already
iterates in three places — `each` fans a document into line objects,
`sum` folds an array, cascade iterates generations under a depth cap —
and the formula language extends exactly that, nothing more.

What the language gets:

- *Comprehensions and folds*: `sum`, `count`, `min`, `max`, `any`,
  `all`, a filter, each over a collection with a per-element expression.
  `sum(line in $.lines: line.qty * line.price)` is "for all order lines".
- *One `fold` with an accumulator*, for scans such as forward scheduling
  (operations in routing order, start = previous end), still bounded by
  the number of operations.
- *No `while`, no recursion, no mutable variables.* Termination by
  construction, so the VM stays total (security rule 1); the model can
  author it, the network can cluster it, and the walk explains it as
  "the sum of these twelve values".

The performance risk sits in the collection, not the loop, so the
design handle is *which collections a formula may walk*. Three kinds:

1. *The event's own payload* — order lines, counted items. Free, and the
   first customer.
2. *Objects one step through a link, keyed and filtered* — the open
   invoices of this customer, the movements of this item in this
   warehouse. This reads state, so it is declared on the rule as a named
   read (as `ref` is today), resolved by index, ordered by log sequence
   for determinism, and refused into the worklist above a per-rule size
   cap — the cascade depth cap's sibling. Scale rule 2 holds because the
   read is keyed, never a scan. **Admitted in phase 1, with the cap from
   day one**: the stock count (book quantity of the item in the
   warehouse) and bank matching (the open invoices of a customer) are
   both in the demo story.
3. *The whole state.* Never.

Whatever a fold costs is paid once at firing and baked into the event
(scale rule 3): replay never loops, views never loop.

What stays out, deliberately: algorithms that carry state across
iterations with ordering or backtracking — FIFO layers, allocating a
header amount across lines with the remainder on the last (largest
remainder), finite-capacity scheduling, MRP netting. These are standards
with names; a client chooses one and never writes one ("Rhea suggests
standards"). They land as kernel methods parameterized by rules and
callable from a formula (`=fifo_cost(item, warehouse, qty)`), the
postings and `convert` precedent. "The AI composes; it never authors
loops" stands unchanged.

The claim this sharpens: every ERP has a calculation engine. The
advantage is that every number carries its inputs and its formula, and
that iteration is only ever over named, capped, visible collections. The
restriction is the feature.

Order of work for the formula session: scalar formulas with declared
rounding and reads through links (the brief above); folds over the
payload; folds over linked collections with the cap; the first kernel
method when FIFO arrives in phase 4.

## Scale and security are welded, not added (KK, 2026-10-08)

KK's question: how does Rhea carry thousands of companies and millions
of transactions a day one day, when performance and security are
explicit non-goals today (SPEC §1)? The answer: the non-goal stands, and
the welding is elsewhere. Rhea already has the shape that large
transaction systems converge on — an append-only log, deterministic
replay, rebuildable projections, one writer — and containment is already
the security model. What is missing is not work but *refusal*: naming
the properties that make scale and isolation possible later, and
declining any feature that spends them. The rules below are of the same
kind as the seven eligibility rules: each maps an ingredient of scale or
trust to a mechanism we already have, and says what would break it.

**Scale rules.**

1. **The company is the unit of scale.** Every event, object, rule and
   view belongs to one company. If the store carries that key on every
   call, then sharding, per-company databases and per-company replay are
   deployment choices later, never a redesign. Today the company is a
   ref inside payloads (E2 made it plural: two `kind: self` companies in
   one log). This is the one item that is cheap now and expensive later:
   **promote the company ref from a payload field to a partition key the
   store knows about**, threaded as an argument rather than a global.
   The isolation enforcement that rides on it stays out (rule S2).
2. **The hot path reads a bounded window, never the whole log.** Firing
   one event may touch that event, the active rules for its type, and
   the objects it references — `ProcessPending` already works this way.
   The full scans are honest experiment shortcuts, already recorded as
   such in DECISIONS: the DuckDB projection rebuilt on every request
   (→ incremental, from the log tail), `Replay` reading every derived
   event (→ snapshot plus tail, rule 3), the object cache rebuilt whole
   (→ the same snapshot). The habit to keep: every full scan gets a
   DECISIONS line naming its production shape.
3. **Replay never re-evaluates.** Deltas are baked at firing time, so
   replay is a pure read of the log — no rule evaluation, no ref
   resolution, no arithmetic (the formula brief's rule 3 extends this).
   That is what keeps audit-by-replay affordable at volume. The
   production shape is a snapshot of the object cache plus replay of the
   tail; invariant 4 survives it because a snapshot is a cached prefix,
   and the full replay stays the test that proves the snapshot honest.
4. **Determinism buys parallelism for free.** Events of different
   companies, or events sharing no objects, can fire concurrently,
   because outcome depends only on log order within a company. What
   protects it: no global mutable state in the executor, no
   cross-company read at fire time (cross-company is the network's job,
   explicitly), and identity and sequences per company — object identity
   is `<type>-<source_event_id>` today, derived from one global log, which
   is fine until the log is partitioned and must then derive from the
   company's own order. E4's intercompany cascade is the known exception
   and crosses companies through refs it resolves at firing, baked
   before replay — so it serializes at that one point, by design.
5. **The model is never on the hot path.** The agent runs at authoring
   time, on worklist residue; a million invoices a day never touch it.
   Model cost scales with *novelty*, not with volume — the economic
   claim beneath AiRP, and worth stating as a rule: no rule, formula or
   adapter may call the model at fire time. Rule matching stays
   indexable (by event type first), so no rule feature may require
   evaluating every rule against every event.

**Security rules.**

1. **Interpret, never execute.** Model output is data, the kernel
   validates it, the gate sits before activation (invariant 3). The first
   real test of this is the formula session: the arithmetic language
   must be *total* — no loops, no I/O, bounded cost per evaluation —
   and run by the kernel's own VM, never an eval. The same holds for
   every later sub-language: algorithmics are kernel code, never data.
2. **The partition key is the isolation key.** The company key of scale
   rule 1 is what row-level security keys on and what every store call
   carries as the principal's scope. Permissions, when they come, belong
   on activities ("who may trigger this verb"), which are already
   declared, versioned and gated — never a second permission system.
   The parked line "authz belongs in the language" (below) is the same
   destination.
3. **Append-only by proof, not by convention.** Today the log is
   append-only because tests forbid UPDATE and DELETE. A hash chain over
   the log — each event carrying the hash of its predecessor — turns the
   audit story into tamper evidence a third party can verify. It touches
   no language and lands whenever multi-company does.
4. **The network is the real security surface.** Rule shapes leave the
   company; facts never do (the network's first slice already strips
   instance values and floors rare shapes). The welding is a test that
   asserts no payload value of any event appears in what the network
   stores; incoming shapes remain untrusted data behind the same gate.
5. **Adapters are the only outbound side effects.** Credentials, retries
   and evidence live behind the adapter contract and nowhere else, with
   evidence recorded as events — the parked SPEC §7 question answered
   from the security side.

**How it becomes a habit.** Two questions join the definition of done
for any new area or direction:

- *Partition:* does this still work with a thousand companies in one
  log, and is there a cross-company read on the hot path?
- *Containment:* does any new input get executed rather than
  interpreted, and can any data leave the gate or the company?

A "yes" is allowed; it costs one DECISIONS line naming the production
shape. What we do not do now, and SPEC is right about it: no sharding,
no caches, no benchmarks, no auth code. The welding is in the data model
(the company key), in the two questions, and in protecting three
properties — bounded reads, baked deltas, one writer.

One line: *scale is the company as the key and replay that never
thinks; security is a kernel that interprets and a network that learns
shapes, never facts — both are already here, and the work is to refuse
what would spend them.*

## The growth plan: six months to a public demo, twelve to "big" (KK, 2026-10-08)

KK asked how to structure development from here: steady progress on
every front, or focus. The answer adopted: **the demo story sets the
depth of every area, and no area gets work the story does not need.**
What the world can see in six months is one business, one country, end
to end; what it can see in twelve is the picture the seven eligibility
rules describe. This note is the plan: areas, priorities, timing, and
what "done" means for each. It replaces no design; each item still
starts from its own brief or direction note.

**The story the demo tells** (Poland, one company, one person at the
screen):

1. The owner describes their business; Rhea proposes the company as
   data with warrants; the owner approves. (Implementation by interview.)
2. Real documents arrive: a KSeF invoice, a scanned supplier invoice, a
   bank statement. Rhea turns each into events and books them, with
   every number explained. (Intake and calculation.)
3. The owner sees what needs deciding and how the business stands:
   worklist, reconciliation, books, briefing. (Views and decisions.)
4. Month end closes. A rule shape learned elsewhere arrives behind the
   owner's gate, and the owner approves it. (The network.)

**Priorities.** P1 blocks the story; P2 the story needs but in a thin
form; P3 is welded now because it is cheap now and expensive later;
P4 is after the demo.

| Area | Today | Done for the demo | P |
|---|---|---|---|
| Calculation | No arithmetic; `convert` only | Formula language per the arithmetic brief: integer/decimal VM, declared rounding, reads through links, folds over the document and over linked collections under a cap, baked at firing; PZ value, VAT, due dates, stock count as customers | P1 |
| Intake | JSON raw events via CLI/API; KSeF stub; clock | Design note first ("documents in"), then three real paths: KSeF, bank statement (MT940), scanned PDF. Extraction is interpretation: a proposed raw event with confidence, validated against the schema, low confidence to the worklist | P1 |
| Views | Generic shell, derived defaults, pl/de eyes, formatted strings in the API | Typed view API (semantics in, formatting in the renderer); one new notion, the **matching grid** (two sides, a pairing verb), with bank reconciliation as its first customer | P2 |
| Decisions | Worklist, approval, simulation diff, explain, cases as data | The briefing as narrative ViewDef; the walk from any number to its inputs and formula | P2 |
| Interview | Agent drafts bundles from one sample; eval 9/9 easy, 6/6 operations | Multi-turn conversation producing a coherent company (types, rules, views, activities) as one gated bundle, driven by the phase-(a) note | P2 |
| Market packs | pl (CoA, VAT, KSeF, warehouse PZ), de proof; warrant on suggestions | Poland deep enough for the story: VAT registers and settlement, bank booking, month-end entries, each as pack drafts with warrants. Germany stays the "one data file" proof | P2 |
| Network | First slice: rule shapes, no facts; live flywheel, 16/19 converged | The closing scene only: a second subscriber, a shape arriving behind the gate, the no-facts-leave test | P2 |
| Isolation and scale | Direction note; company as payload ref | Company key promoted to a store-level partition key, threaded as an argument. Nothing else: no sharding, no auth, no benchmarks | P3 |
| Backup and integrity | Nothing | Backup is the event log plus definitions; a test restores from that alone and replay reproduces state. Hash chain when multi-company lands | P3 |
| Tests and eval | Invariant tests, `rhea eval` corpora | The whole demo story as one unattended eval run; the share of it passing is the progress metric | P3 |
| Enterprise structure | E1–E5 shipped | Nothing for the demo; a real group after it | P4 |
| Notifications | Nothing | An outbound adapter over cases and the worklist (mail or chat), evidence as events | P4 |
| Agent interface | Nothing | An MCP server as one more interpreter of the two vocabularies: views to read, activities to call, the gate unchanged | P4 |
| Algorithmic sub-languages | `each`, postings | FIFO valuation, allocation, stock count as kernel method vocabulary parameterised by rules | P4 |

**Timing.** Four phases; the unit of work stays the session, one
scene of the story at a time, touching whichever area that scene needs.

- **Phase 1, October to November 2026: numbers and intake.** Formula
  session; company key; intake design note; KSeF and bank statement in;
  typed view API. Exit: a KSeF invoice and its bank payment flow in,
  book with correct VAT, and reconcile by rule.
- **Phase 2, December 2026 to January 2027: seeing and deciding.**
  Matching grid; scanned PDF behind the gate; interview as a
  conversation; Polish pack depth; backup-and-restore test; briefing.
  Exit: a new company can be stood up from an interview and run a
  month of real documents.
- **Phase 3, February to March 2027: the story end to end.** Full-story
  eval unattended; network closing scene with two subscribers; demo
  rehearsal; marketing site aligned with what runs. Exit: **public demo,
  April 2027.**
- **Phase 4, April to September 2027: the big picture.** Germany with
  real depth; enterprise structure on a real group; notifications; MCP;
  hash-chained log; permissions on activities; FIFO and allocation.
  Target: the seven eligibility rules demonstrable end to end,
  **October 2027.**

**Rules of the plan.**

- An area gets work only when a scene of the story asks for it. "It
  would be nice" is a P4 line, not a session.
- Every new area starts with a brief in this file, like arithmetic did,
  and answers the two welding questions (partition, containment).
- The views list is counted in notions, not screens: list, document,
  ledger, matching grid, queue, tree, narrative. Reconciliation,
  intercompany matching and bank matching are one notion. A screen that
  knows what an invoice is does not get built.
- The metric is one number: how much of the demo story the eval runs
  unattended, end to end.

One line: *the story sets the depth; six months buys one country end to
end, twelve buys the picture of "big", and everything the story does not
ask for waits.*

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
  The AI composes sub-languages; it never authors loops. Refined
  2026-10-08 (the iteration note in the arithmetic brief): bounded folds
  over named, capped collections are formula vocabulary; a loop that runs
  until a condition holds never is.
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
  Made buildable in "Late understanding" above (adopted 2026-10-08):
  additive pairs only, a deliberate act sibling to catch-up, and a pair
  bound for a closed period offered forward into the first open one.
- **Conflicts stay human.** Same-object-id claims refuse the event into the
  worklist; semantic double-booking is simulation's job to reveal before
  approval; static conflict detection is parked (SPEC §7).
- **Authz belongs in the language, eventually.** Users, groups and access
  policies should be objects and rules like everything else — a good
  falsifiability test of its own, post-M2. Until then: actor attribution
  only, no auth (experiment non-goal).
- **Scheduling ships with its hand on the switch** (KK, 2026-10-05:
  thirty years of implementations say scheduling is where automation
  outruns organic, unorganized reality — leave a space to switch it off).
  The off switches are mechanisms the language already has, at three
  granularities, and M5 must preserve them rather than invent a toggle:
  per behavior, supersede the rule (the lifecycle is the kill switch,
  with the log answering "who turned it off and why"); per schedule,
  amend the schedule object's status (active ⇄ paused as declared
  transitions — the amendment's first operational customer); globally,
  stop the clock adapter — time stops arriving and the log says so
  honestly. The constraint this adds to M5: **re-entry is gated. Steady
  state is automatic, bursts need a human.** After a gap the adapter
  never floods the missed days in; it proposes them — "N days unopened;
  opening them fires this" — behind the same simulation-then-approval
  gate everything passes, a deliberate act with a dry-run, sibling to
  ruled backfill. Trust comes from the same physics: every automated
  firing provenance-chains to its time event, and a misfiring schedule
  fails as data into the worklist, never silently. **Shipped 2026-10-05**
  (see DECISIONS, M5): all three switches live, re-entry gated exactly as
  stated — RunClock opens one day, a burst gates, the dry-run reads like
  a rule approval, CatchUp is the deliberate act.
- **Multi-company is a ref, multi-tenant is infrastructure.** The operating
  company will be a `company` object (`kind: self`) that events reference;
  tenancy stays out of the experiment. Promoted (KK, 2026-10-04): enterprise
  structure is the named gap against big ERP and now has a destination —
  parallel GAAPs as parallel rule-books explaining one event log,
  intercompany as cascade across company refs, FX in the arithmetic
  sub-language. A falsifiability test of the M2 kind: zero kernel changes.
  Planned (2026-10-04): the enterprise-structure ladder above, E1–E5.
- **The company key is the partition key** (KK, 2026-10-08, "Scale and
  security are welded"). The one scale item that is cheap now and
  expensive later: the company ref promoted from payload field to a key
  the store carries on every call — isolation and sharding ride on it
  later, enforced by nothing yet. Alongside: snapshot-plus-tail replay,
  a hash-chained log, and the network test that no fact ever leaves.
