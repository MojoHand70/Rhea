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

**Shipped 2026-10-08** (see DECISIONS, "the formula language"): the first
three steps. Every `=` template is a formula compiled by the kernel and
run by its VM over exact rationals; division lives only under a declared
`round`; money, int and decimal are kinds the checker and the VM both
enforce; `$.state.item.std_cost` and `ref(...).field` read through links
typed by the catalog; folds walk the payload's lists or `objects(T, field,
value)` under the cap; every computed value is baked into its event with
the formula and the inputs it read, and the walk shows them. The first
customers — PZ line value, the PZ entry at standard cost, VAT with
statutory rounding, the due date, the stock count's book quantity — run
end to end. The fourth step waits for FIFO; formula → SQL waits for an
analysis view that needs it.

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
   **Reporting is the one place a model's query runs (KK, 2026-10-17):**
   read-only SQL on the DuckDB analysis copy, in a session locked before
   the query arrives — read-only, no files, no network, no extensions,
   configuration frozen, one SELECT, bounded time, memory and rows. It
   reads a derived copy and can change nothing; a kept report still waits
   at the gate.
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
| Tests and eval | Invariant tests, `rhea eval` corpora; simulated customers (2026-10-09: personas with generated documents and rule-free expected state, the model only paraphrasing the interview; the verdict over repeated runs and voices is all or nothing) | The whole demo story as one unattended eval run; the share of it passing in every run is the progress metric | P3 |
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

## The host is a concierge: scenarios as data (KK, 2026-10-10)

KK's image for the interaction side, verbatim: *"She should be like the
best concierge in the world. Absolute luxury. Like Winston in John Wick,
like Charon."* The Continental's rules are absolute and everyone knows
it; the concierge is discreet, anticipates, knows every guest by name
and history, speaks little, is never obsequious, and acts within the
rules without being asked twice. In Rhea the rules are the invariants
and the approved corpus; the concierge is the host of the north-star
note (system as host, owner as principal, the decision as the unit of
interaction). Rhea is "she".

KK's list of what she does — welcomes, suggests what to do today, points
at hotspots and bottlenecks, runs implementation interviews, manages
rights and access, executes actions, notifies other people, helps settle
substantive disputes — and the question: how do we approach use case
scenarios, and how does Rhea support more and more of them over time?

**The interaction side is not a second system.** It is the same six
concepts seen from the chair of the person at the screen. A scenario is
a composition, every part is data behind the approval, and so scenarios
grow the way the rule corpus grows. The morning scene ("I contacted the
bank, booked 98 invoices, 2 are strange and need your decision") has
five parts, and every scenario has the same five:

1. *A trigger.* The clock, an event, or the person.
2. *Facts the kernel computes.* A view: bookings made, documents still
   unexplained, open cases by deadline.
3. *What the person can do.* Actions offered on those facts, inline.
4. *What Rhea does alone.* Rules.
5. *What the host says.* The agent phrases the facts fresh each time;
   its words are logged as events so it never repeats itself (the
   north-star note's "never a templated welcome"). It never computes a
   number.

A scenario is therefore a set of definitions — a case type, a view, a
few actions, a few rules, sometimes an adapter — approved together like
a pack. No scenario needs code in the shell; if one seems to, the view
vocabulary is missing a kind, and the kind gets added, never the screen.

**The eight scenarios, mapped.**

| Scenario | What it is in Rhea | Missing |
|---|---|---|
| Welcomes | Briefing: facts from the kernel, narrative by the agent, logged | The narrative view kind; a "host said" door |
| What to do today | Open cases ranked by deadline and amount, plus pending approvals | The queue view kind; ranking as a formula on the case |
| Hotspots, bottlenecks | Thresholds are rules that count across a customer's documents and raise a case; trends are analysis views the agent reads and phrases | Nothing in the language; a library of hotspot rules in the pack |
| Implementation interview | Adopted 2026-10-08 ("Implementation by interview"); turns are events | The conversation loop, step 6 of that note |
| Rights and access | "Who may trigger this action" on the action itself; people and roles as objects | The mechanism — Phase 4 |
| Executes actions | Doors for the person; rules and the clock for Rhea; adapters for the outside world | Nothing |
| Notifies people | A rule routes a case to a person; an adapter delivers; evidence as events | The outbound adapter — Phase 4 |
| Substantive disputes | A case with two parties; the evidence is the walk from the number to the events and rules behind it; what-if shows the month under the other side's reading; resolution is a door (correction, acknowledgment, redraft) | A dispute case type in the pack; the walk from any number — Phase 2 |

Seven of eight need no new concept. Rights has a destination ("authz
belongs in the language", below) and no mechanism, and the growth plan
already places it in Phase 4.

**How the catalogue grows.** Three sources, one measure.

- *Packs ship scenarios* with a warrant, as they ship VAT rules:
  "overdue receivables" is a case type, a threshold rule, a queue view
  and two doors.
- *The host captures decisions.* The person handles something by hand
  twice; the host asks whether to put it down for the future and drafts
  the rule. A hotspot the host pointed out and the person confirmed
  becomes a rule the same way. This is the Junior Accountant → CFO path,
  and it needs nothing beyond the interview loop.
- *The network shares shapes.* A hotspot rule that worked for one
  company arrives behind another company's approval.

The measure is the simulated customer: each scenario is one scene in a
persona's day, and "Rhea supports this scenario" means the scene passes
unattended in every run — the growth plan's metric, unchanged.

**The one decision, settled: people are data now.** Three scenarios key
on a person — who is notified, who may act, who disputes — and the actor
on an event was a plain string. KK, 2026-10-10: "I think we need to add
person object now." Shipped the same day (see DECISIONS): a `person`
type in the first market-neutral pack (`packs/common/people.json`) with
a role, a deputy and a life; three doors; the pl pack's task vocabulary
assigning work to persons by ref; the actor string resolving to a
person by lookup, never by kernel code. The warrant is the one the
company key got: cheap now, expensive later. What stays out, by the
plan: permissions (Phase 4), login (never, in the experiment), the
notification address (with the adapter that sends).

**Order.** Phase 2 of the plan, scene 3 of the story, already covers
four scenarios: the briefing, what to do today, hotspots from the pack,
and the walk that disputes need. The interview loop sits in the same
phase. Rights and notifications stay in Phase 4.

**The two welding questions.** *Partition:* a person belongs to a
company's log like a contractor does; the same human serving several
client companies (bureau mode) is a person in each, matched by handle
across logs only in the network's sense, never by a cross-company read
on the hot path. *Containment:* the host's words are events in the
company's log and the agent reads persons only through views; a
person's name and role are instance values, which the network's
no-facts-leave test already strips.

One line: *the concierge is the same six concepts from the guest's
chair; a scenario is five parts of data behind one approval, and the
guest is now a person Rhea can name.*

## Documents in: channels, readers, and the file as the fact (KK, 2026-10-10)

KK's call: the time has come for intake, and channels first. Three
things KK added the same day: *a registry of actually implemented
channels, versioned*; *streaming from POS and IoT systems as a future
channel*; and *a dedicated treatment for high-volume channels, for
performance*. This section is the brief the intake work starts from,
the arithmetic brief's sibling: constraints and a map, not a design.
The growth plan places it in Phase 1 (KSeF and the bank statement in),
with the scanned PDF behind the gate in Phase 2.

**Prior art to read first.** `internal/adapter/adapter.go` (the
contract: an adapter is a user, acts only by returning raw events,
replay never re-runs it), `ksef.go` (outbound, the only statutory
adapter today), `clock.go` (time as a channel, re-entry gated), the
`submit_event` door and `rhea submit` (dedup by content hash), the
people pack (an actor string resolved to an object by lookup, the shape
a supplier's NIP will take), "Late understanding" above (the file must
survive its first reading), and `internal/sim` (the persona generator,
which becomes the outside world).

**The channel map.** A channel is a way documents reach the log. What
distinguishes one from another is *who reads the document and how sure
they are*; everything else is one mechanism.

| Channel | What arrives | Form | Reader | Identity (never twice) | Phase |
|---|---|---|---|---|---|
| KSeF | Purchase invoices from Polish suppliers; the UPO for our sales invoices | FA XML (FA(3) under the 2026 mandate), pulled by our NIP | Parser, exact | KSeF reference | 1 |
| Bank | Statements: payments in and out, fees, interest | MT940 or CSV export; an API feed later | Parser, exact, except the free-text line every bank writes differently | Statement number + line | 1 |
| Email | Foreign supplier invoices, subscriptions, proformas, customer orders | PDF, with or without a text layer | Model, with a confidence | File hash + message id | 2 |
| Paper and photo | Receipts, paper invoices, signed delivery notes, expense slips | Scan or photo | Model, with a confidence | File hash | 2 |
| A person | Cash sale, stock count, correction, acknowledgment | Typed in the shell | The person | The door's dedup key | exists |
| Other systems | Webshop and marketplace orders, POS exports, the payroll bureau's file, warehouse scanners | CSV, JSON, API | Parser per source, exact | Source record id | 4 |
| Streams | POS receipts as they happen; meter and sensor readings | Push: webhook, MQTT, a queue | Parser, exact; grain declared | Device id + sequence | 4 |
| Other portals | ZUS notices, tax office messages, JPK and return acknowledgments | Portal downloads | Parser, exact | Portal reference | 4 |
| The clock | Days opening | Not a document | Nobody | Date | exists (M5) |

The map is intent. The registry below is the fact: which channels a
company actually has, in which version.

**What the invariants demand.**

1. *The file is the fact; the reading is an interpretation with an
   author.* A document first enters the log as received, a
   `document.received` event with its hash, source channel, name and
   type, the bytes kept in an append-only blob store keyed by hash (a
   new append-only table; invariant 1 extends to it). Then a reader turns
   it into the typed raw event (`purchase.invoice.received`,
   `bank.statement.received`), signed by the reader (actor
   `adapter:<channel>`, `agent:<model>`, or a person) and pointing at the
   file by hash. For KSeF and the bank the reader is a parser and the
   reading is exact. For a PDF the reader is the model and the reading
   carries a confidence. The model's reading is data, never state
   (invariant 3): proposed, validated against the event schema, then
   booked or sent to the worklist. A person is a reader too: confirming
   a doubtful reading is a door, and the typed event then carries the
   person as actor.
2. *Replay never re-reads.* The adapter contract already says replay
   never re-runs an adapter. A model's reading is non-deterministic, and
   that is fine, because it is a raw event in the log with the model and
   the confidence on it. Invariant 4 holds by construction.
3. *Every raw event names its channel and version.* Derived events carry
   `(rule_id, rule_version)`; raw events through a door carry
   `(activity_name, activity_version)`, which `core.Event` documents as
   empty for adapter events. The registry fills that gap: adapter events
   stamp the channel's activity name and version. Invariant 5 reaches
   intake, and the walk reads: purchase_invoice ← rule v2 ←
   `purchase.invoice.received` (KSeF channel v1, exact) ←
   `document.received` sha256:… (pulled 2026-10-10).
4. *Readers keep everything the document says.* No curated subset. Late
   understanding is the reason: the field nobody asked for is wanted
   later, and a stored file can be re-read by a new channel version, the
   result landing as an amendment, never as a rewrite.
5. *Never twice, per channel.* Each channel declares its identity rule,
   which becomes the dedup key: the KSeF reference, statement number plus
   line, file hash plus message id, device plus sequence. Retry stays
   free, as the contract promises.
6. *Doubt goes to the worklist; the threshold is company data.* A reading
   below the channel's confidence threshold waits for a person; above it,
   it books and appears in the morning line ("booked 98, 2 need you").
   The default threshold ships in the pack with its warrant, and any
   company may change it. A booked reading that was wrong is corrected
   by amendment, like any fact.
7. *Numbers cross as strings; the parser owns the decimal.* "1 234,56"
   becomes minor units with the document's currency inside the reader.
   No float anywhere (invariant 6).
8. *Unknown parties are residue, not errors.* A supplier the company has
   never seen is matched by NIP the way an actor is matched by handle.
   No match refuses the event into the worklist, and a door creates the
   contractor from the document. The people pack is the precedent.

**The registry: a channel is an activity triggered by code (KK's
requirement, versioned).** No seventh concept. An activity today
declares inputs, the event it emits and who may trigger it. A channel is
an activity whose trigger is an adapter rather than a person, with four
more declared things: the *binding* (the adapter name and code version
the kernel must have registered; a mismatch refuses the channel at
startup, as data), the *format* it reads (FA(3), MT940, PDF), the
*identity* rule (item 5), and the *confidence policy* (exact, or a model
with a threshold). It is versioned like every definition, installed by a
pack or drafted, and passes the same approval as a rule (settled
2026-10-08: one gate, no second path). Three consequences:

- *Implemented versus connected.* The code's adapter registry says what
  Rhea can read; the active channel definitions say what this company
  has connected. Both are generic list views in the shell; neither is a
  screen.
- *Format versions coexist.* When FA(3) gives way to FA(4), or a bank
  changes its export, a new channel version activates and the old one
  supersedes; every event says which version read it.
- *The persona generator targets the registry.* It emits a document per
  channel the persona has (a FA XML file, an MT940 file, a PDF), not
  JSON events, the step already noted for the simulated customer.
  "Documents in, state out" becomes the eval, and residue over the
  intake corpus stays the gauge.

**Streams: the same contract, pushed instead of polled (KK's future
channel: POS and IoT).** The adapter contract defines a pass as a poll.
A stream channel holds a listener (webhook, MQTT, a queue) and its pass
drains what the listener buffered: still a pass, still idempotent by
device plus sequence. What a stream needs that a document does not is a
declared *grain*: one event per item (a POS receipt is a document; the
cash sale in testdata already is one) or one event per window (a
sensor's hourly reading), with the raw chunk kept as the evidence file
of item 1, keyed by device and window, re-readable later. Polish
practice offers the document-shaped grain for free: an online fiscal
register produces a daily report, and the receipts behind it are the
finer grain when stock needs them. Late readings are late facts, which
the log already distinguishes from late understanding. Phase 4; the
contract is fixed now so that nothing is designed against it.

**High-volume channels get a dedicated treatment (KK, for performance),
inside the non-goal.** No benchmarks and no caches, as SPEC says. The
treatment is a declared *volume class* on the channel (`documents` or
`stream`) and what the kernel does differently for `stream`:

1. *Batch, not one by one.* A pass appends its events in one
   transaction and projects them in one `ProcessPending`, never a round
   trip per event. Today `adapter.Run` appends one at a time; the batch
   append is this brief's cheap-now item, the company key's sibling, and
   KSeF pulls and bank statements are batches already.
2. *Residue is counted by shape, not listed by event.* A million
   receipts with a new product code must not raise a million worklist
   items. A stream channel's unexplained events fold into one case per
   (event type, reason) with a count and a sample, which is already how
   the interview clusters residue. The person decides once.
3. *Rules on stream events are keyed and capped.* The formula cap on
   linked collections and scale rule 2 (a bounded window, never the log)
   already apply; the simulation reports per-event cost for a stream
   channel so an expensive rule is seen before approval. The model is
   never on the path (scale rule 5).
4. *Aggregates are analysis views.* Sums over a month of readings live
   in DuckDB, incrementally from the log tail, never as object lists.
5. *Replay is snapshot plus tail* (scale rule 3), which a stream log is
   the first to need in practice.

**First customers.** Nordwind's purchase invoice arriving through KSeF
as FA XML: the purchase invoice document, the input VAT register and
the entry (net to 300, input VAT to the VAT account, gross to 202) as
pack drafts with warrants, then the delivery settled against it (the
300/202 pair noted in the arithmetic brief). The same month's MT940
statement: the pack's existing `bank_account` and `bank_operation`
types, then a matching rule that pairs a payment with the open invoice
of that contractor (a fold under the cap, already in the language) and
amends the invoice's status to paid. That is the Phase 1 exit of the
growth plan. Phase 2's first model-read document: a German supplier's
invoice in EUR as a PDF, the real reason a Polish company still
receives PDFs after KSeF, carrying the FX the formula language already
does. The matching grid view comes with it.

**Not questions for KK** ("Rhea suggests standards"): how a purchase
invoice is booked, how a payment is matched (amount and reference
first, then amount and contractor), what the default confidence
threshold is. Each arrives as a pack draft with its warrant; KK or the
client overrides.

**The two welding questions.** *Partition:* a channel belongs to a
company like a contractor does; the blob store carries the company key;
a stream's device belongs to one company. *Containment:* a reader's
output is validated against the event schema before it touches the log;
the model reads the file and returns JSON through the agent package,
which has no store access; a channel's credentials live behind the
adapter contract and nowhere else (security rule 5).

**Order of work.**

1. The registry: channel as activity with binding, format, identity and
   confidence policy; adapter events stamp channel name and version;
   the startup check; two generic list views.
2. The file as the fact: `document.received`, the blob store, an upload
   path in the CLI and the shell (`rhea intake FILE`, a drop zone), and
   the adapter pass appending in one transaction (the cheap-now item).
3. The persona generator emits documents: a FA XML and an MT940 file per
   persona month.
4. KSeF inbound: the FA parser as a channel; the purchase invoice, VAT
   register and entry as pack drafts; the eval runs documents in.
5. The bank statement: the MT940 channel keeping the raw free-text line;
   the matching rule; the invoice paid by amendment. Phase 1 exit.
6. The PDF reader behind the gate, with the German supplier's invoice,
   and residue counted by shape. Phase 2.
7. The first stream channel when a POS arrives. Phase 4.

Status: **drafted 2026-10-10**. The channel map agreed by KK; the
registry, streams and the high-volume treatment added by KK the same
day; KK: "let's move on" (2026-10-11). **Step 1 shipped 2026-10-11**
(see DECISIONS, "the channel registry"): a channel is an activity with
a `channel` section, the binary registers its adapters, `Bind` joins the
two before every pass, every adapter event names its channel, the clock
is the first channel (a builtin) and the Poland pack ships the KSeF
channel; `rhea channels` and the Channels tab show both lists.
**Step 2 shipped 2026-10-11** (DECISIONS, "the file is the fact"): a
`file` input kind on verbs, bytes kept in an append-only blob store under
their hash inside the trigger's transaction, `document.received` refused
without its file, the `receive_document` door (`rhea intake`, a derived
file picker in the shell), the `received_document` type in a common pack,
files linked wherever the shell shows them, and the adapter pass appending
in one transaction. **Step 3 shipped 2026-10-11** (DECISIONS, "the
simulated customer emits documents"): a FA-shaped XML per Polish
supplier's invoice named by its KSeF reference, one MT940 per month in the
"sim-bank" dialect, each naming the events it carries; `rhea sim
-documents DIR`; the FA namespace and rate slots to be checked against the
published XSD before the first real pull. **Step 4 shipped 2026-10-11**
(DECISIONS, "the KSeF reader"): `ksef_inbound` reads Faktura files from the
store into `purchase.invoice.received` naming the file, refuses a document
that contradicts itself into the worklist, and marks the document read;
declared by `packs/pl/ksef-inbound.json`; the eval installs packs, receives
files and runs readers, and both personas reach DONE with documents in.
**Step 5 shipped 2026-10-11** (DECISIONS, "the bank reader and
reconciliation by rule"): `bank_mt940` reads statements into
`bank.statement.line` with direction, accounts and the raw `:86:`, refuses
a statement whose balances do not add up; the generator pays purchase
invoices; the interviews split money in from money out; the Poland pack
(v9) books the purchase invoice from the document's own VAT, posts Wn 300 /
Wn 221 / Ma 202 and shows the purchase VAT register. **Phase 1's exit holds
on the bench**: a KSeF invoice and its bank payment flow in as files, book
with correct VAT, and reconcile by rule, both personas DONE on the
reference answers. Next: the live run with the model, documents in, then
documents-in as the default.

One line: *a channel is who reads the document and how sure they are;
the file is the fact and the reading has an author; the registry is an
activity triggered by code, versioned and approved like everything
else; a stream is a pushed channel with a declared grain, batched and
counted rather than listed.*

## The host remembers: memory is the log with a lifecycle (KK, 2026-10-11)

The case that opened it. On 2026-10-10 KK corrected a reading of mine
(the "all-or-nothing over runs" verdict was my construction, not KK's
rule). The correction went into my notes as a new note beside the old
one; nothing marked the old one dead; a day later I cited the old one as
KK's standard. Asked why it was not marked, the honest answer: appending
was cheap and retiring cost a search through prose, so the correction
was added, not applied. KK's question: can Rhea afford this? And a
requirement, verbatim in spirit: *she is the concierge, she has to
remember* — preferences, habits, customs, personality.

**Both are true, and the second is the answer to the first.** Rhea
remembers everything, and she remembers it as facts in the log and
definitions with a lifecycle, never as notes. That is what makes her
safe to remember with. The concierge who forgets is poor service; the
concierge who remembers the guest's old room after the guest moved is
worse. Winston knows both rooms, and which one is current.

**The rule.** Everything Rhea remembers is one of two things: a fact in
the event log (dated, signed, append-only) or a definition with a
lifecycle (draft → active → superseded). A correction supersedes; it
never sits beside. The old version stays readable as history, with its
date and the reason it was replaced, and nothing the host says, drafts
or does may draw on a superseded version as current. Supersession must
be as cheap as recording — one act, behind the same gate — or stale
truths accumulate the way they did in my notes.

**Where each kind of memory lives.**

- *Preferences* — how a person wants to be briefed, in which language,
  at what hour, how much detail, which name — are amendable fields on the
  person (the people pack of 2026-10-10), each amendment carrying who
  changed it and when. "You told me on 3 October you prefer the brief at
  seven" is a walk, not a recollection.
- *Habits* are patterns the kernel can count in the log — the person
  handles something by hand twice — and the host turns into a proposal
  behind the gate ("The host captures decisions", the concierge note).
  Until approved, a habit is an observation the host may mention, never
  a rule she acts on.
- *Customs* — how this company does things — are rules, scenarios and
  packs: the corpus itself, versioned.
- *Personality*, as the host meets it, is the sum of the above plus the
  person's own words in the interview, which are events. The host's own
  words are events too, so she never repeats a welcome and can be asked
  what she said.

**Where Rhea already refuses my mistake, and where she still could.**

1. *Acting.* A superseded rule cannot fire: invariant 2, enforced by the
   kernel and its tests; replay proves every booking used the version in
   force on its day. This is the part that matters most and it holds.
2. *Drafting.* The author's ask is built from the active rules, and
   rejected drafts reach it only with the reason they were rejected
   (`DraftAsk`, the conversation of 2026-10-10). Holds today; the test
   below pins it.
3. *Speaking.* The briefing phrases facts the kernel computed from
   current state; the host's words are logged. Pins with the same test
   when the narrative view kind lands.
4. *Learning.* The gap found 2026-10-11: what an installation publishes
   to the network is counted forever, and the network has no notion of a
   shape being superseded where it came from. A standard every subscriber
   abandoned would still arrive at a new subscriber with its old support.
   The fix is the same shape as everything else: a supersession is
   published too, and support follows the versions in force.

**The standing test — "the overruled rule never returns."** Approve a
rule; supersede it with a different reading; then: replay never fires
the old one after its time; the author's ask names the new one and shows
the old only as superseded, with the reason; the network counts the new
one; the briefing speaks the new one. Of the ladder's kind: if the shape
cannot carry it, the shape is wrong.

**Order.** The test for acting and drafting now (small); the network's
supersession with the next network slice; speaking with the briefing.

One line: *the concierge remembers everything, as facts with a date and
a status; a correction retires what it corrects; and the one place a
memory could lie — in the author's eyes — is pinned by a test.*

## Matching: a deterministic engine with N sides, as a kernel method (KK, 2026-10-11)

KK's call, after the bank channel landed and the matching grid came up
next: matching is one of the crucial components of any business system
and a very difficult domain — by exact amount, by amount and reference,
partially, 1:1, 1:n, n:1, n:n, accrued in time, per saldo, three-way,
with FX, with tolerance, and more modes than anyone lists up front.
*"We need to design a deterministic matching engine, which Rhea will use
as a tool."* And, from experience: *three-way matching happens in many
circumstances, not only order and receipt — the skeleton must be
generic.* Agreed: this is the biggest member of the family the
arithmetic brief named — standards with names, as kernel methods
parameterized by rules, which the model composes and never authors —
and it earns a primitive and **a dedicated session**. This section is
the handover that session starts from: constraints, a skeleton and an
order of work, not a design.

**Prior art to read first.** In the repo: the arithmetic brief above
(folds under a cap, kernel methods callable from a formula, "the
restriction is the feature"); `convert` on postings (a method in the
kernel, parameters as rule data, declared rounding, a difference to a
named account — the FX precedent); `amend` and the lifecycle (open
amounts move by provenanced events); `each` (fan-out); the bank reader
and the KSeF reader (the two sides as they arrive today); the Poland
pack's `settlement`, `bank_operation` and `contractor` types; the
warehouse pack's PZ and lines; the growth plan's "matching grid" row.
Outside: how the big systems name the standards — clearing with
tolerance groups, residual and partial items, clearing differences
(SAP); settlement and the settle-by-priority rules (Dynamics); Comarch's
rozliczenia and kompensaty; the Polish statutes behind the names —
różnice kursowe (ustawa o rachunkowości art. 30), kompensata
(kodeks cywilny art. 498). Read them for the vocabulary and the
warrants, not the architecture.

**The skeleton.** One engine, N sides, a declared structure.

1. *A side* is a keyed, capped collection of items: objects of one type,
   each carrying a vector of measures (money in a currency, quantity in
   a unit, a count), a date, the references it names, a party, and its
   *open* portion — what is not yet allocated. The bank lines of one
   account, the open invoices of one counterparty, the receipts of one
   purchase order.
2. *A structure* says how the sides relate: which measures must agree
   across which sides (the payment's money against the invoice's gross;
   the receipt's quantity against the order's; the invoice's price
   against the order's), through which converters (an FX rate at a
   date, a unit conversion), within which tolerances (an amount, a
   percentage, a number of days), and which keys join candidates (the
   party, the currency, a reference). Two sides is the common case;
   three is order, receipt and invoice, but also invoice, bank line and
   the factoring house's settlement; payroll, transfers and the ZUS
   declaration; a customs declaration, the supplier's invoice and the
   carrier's; our purchase invoice, their sales invoice and the group's
   settlement. Four is order, delivery note, invoice, payment. The
   engine knows N; the structure is data.
3. *Strategies* are the search procedures that propose matches under a
   structure — the standards with names, each deterministic and total:
   by reference (a shared key); exact (the vectors agree); subset sum
   (one item against a subset of the other side under a size cap,
   nested across sides under the same cap); oldest first with partial
   allocation (an item settles the earliest open items as far as it
   reaches, the remainder stays open); proportional allocation; within
   a time window; tolerance (a difference up to a bound is absorbed to
   a declared account); FX (measures converted at the rate of a
   declared date, the difference to a declared account); netting per
   saldo (two sides of the same party with opposite signs read as one).
   A *policy* is an ordered list of strategies with their parameters,
   bound to a structure. The first feasible match wins; the order is
   the company's choice, shipped as pack defaults with warrants.
4. *The outcome is events, baked.* Every match is one event naming the
   items on every side, the amount allocated per item and measure, the
   residual left open, every difference with its kind (tolerance, FX,
   rounding) and its account, and the strategy that found it. Ordinary
   rules explain it: the items' open amounts move by amendment, the
   settlement object materializes, the differences post. Replay never
   re-matches. A wrong match is undone by a correcting event that
   restores the open amounts, never by deletion; period locks apply.
5. *The remainder is the grid.* What no strategy settles is the N-sided
   view for a person; the pairing verb runs the same engine on the
   person's selection and emits the same event. One fact, explained the
   same way, whoever made it. The host proposes a strategy after a
   person has paired the same way twice — behind the gate.

**What the invariants demand.**

1. *No floats* (invariant 6): exact rationals inside, declared rounding
   wherever an allocation or a conversion can lose precision; money
   stays int64 minor units in the events.
2. *Total and bounded* (security rule 1, scale rule 2): keyed sides,
   a cap per side and a subset size cap per policy, declared on the
   rule; above the cap the items go to the person, never to a long
   search. Cost is paid once at firing.
3. *Baked at firing* (scale rule 3): the engine runs inside rule
   expansion, like a formula; replay reads.
4. *Every allocation explained* (invariant 5): strategy, items, measures,
   tolerances used, rate used — in the event, shown by the walk.
5. *Append-only* (invariant 1): open amounts are state moved by
   events; unmatching is a correcting event.
6. *The two welding questions.* Partition: a side never crosses
   companies — intercompany matching crosses through refs at firing, as
   E4's mirror did, and serializes there. Containment: the policy is
   data the agent may compose and the gate approves; the engine is
   kernel code; no strategy is ever expressed as a loop in data.
7. *Late sides.* The payment can arrive before the invoice. A policy
   fires on the arrival of any side's item and reads the other sides'
   open items by key, so the match happens when the last side arrives;
   a scheduled re-run (a clock-fired rule) catches what a late
   correction reopened.

**What earns the primitive — the attempt.** The ladder's rule: try it
with what exists first. By reference and by exact amount already work as
rules with a lookup. Write the bank story with rules and folds and let
it fail where it must: a subset that sums, a partial allocation that
leaves a residual, a difference booked to an account, a match that
spans N sides. Record the attempt in DECISIONS; the failure is the proof
that `match` is a closed-class effect, as postings, `convert` and
`amend` were.

**First customers, in the order the demo needs them.**

1. A bank line against sales and purchase invoices: by reference (the
   rule of today), then exact amount, then subset sum for a customer who
   pays three invoices at once, then oldest first for a part payment,
   then a two-grosz tolerance to a declared account.
2. The same with FX: an EUR invoice paid in PLN, the difference to
   755/756 at the rate of the payment date — `convert`'s rate source,
   `convert`'s rounding.
3. Netting per saldo for a party who is both customer and supplier.
4. Three-way: the purchase order (when it arrives), the PZ and the
   supplier's invoice, quantities and prices with tolerances — the
   warehouse pack's customer.
5. The generality test: one more three-way from a different corner
   (invoice, bank line, a factoring settlement, or payroll, transfers
   and the declaration), chosen so the structure is reused and nothing
   is added to the engine.

**Not questions for KK** ("Rhea suggests standards"): which strategies
in which order a Polish company runs, the tolerance bound, the
difference accounts, the time window. The pack proposes the standard
with its warrant; the client overrides.

**Order of work for the session.** The attempt; the engine's core
(sides, structure, two-way with reference, exact, subset, partial,
tolerance; the match event; open amounts; the settlement object); FX
and netting; N sides with the three-way PZ customer; the grid notion
over the remainder; the host proposing strategies — the last with the
interview loop. The eval gains a persona scene per strategy; the
simulated customer pays in every way the strategies name.

One line: *matching is one engine with N sides and a declared structure;
strategies are standards with names, ordered by a policy that is data;
every match is a baked, explained event; what the engine cannot settle
is a person's grid, and the person's pairing is the same event.*

**KK, 2026-10-11, during the session.** Two notes. *Each company can
have a different policy for payment matching* — which is exactly what
the policy being data on the company's own rule gives: the order of
strategies, the tolerance, the difference account and the time window
are the rule's, the Poland pack proposes a default with its warrant, and
the owner overrides it in the interview. And *intercompany matching for
group consolidation is one of the important kinds* — it is one of the
three-sided structures above (our purchase invoice, their sales invoice,
the group's settlement), keyed across company refs with the document
number as the reference and the group currency through the converter;
the match event is the elimination's evidence. It is the generality
customer of this session (customer 5), not an afterthought.

**Shipped 2026-10-11, the first steps** (see DECISIONS, "the matching
attempt" and "the matching engine's core"): the attempt — by reference
carries, exact fails on the one-field key, a partial on the allocation
across invoices, a tolerance on the conditional posting line, a subset
and N sides have no vocabulary; the engine's core — a pure engine with N
sides, the structure and the policy as the `matching` effect (spelled so
because a rule already has a `match`), four strategies (reference, exact,
subset, oldest first) with tolerance and a time window, one baked event
per match or non-match, consequences as ordinary rules (`each` on amend
earned), replay and simulation sharing the path, the walk showing it, the
agent taught; the persona scene — the simulated customers pay in every way
the policy names (named, exact, several at once, two grosz short, a named
part, an unnamed part the policy reads as the oldest), the generator
predicts the settlement without any rule, the interview answers carry the
owner's policy, both personas DONE as JSON and as documents; FX and
netting — the converter as a clause of the structure (read in the
trigger's currency at the settlement date's rate, written back in each
item's own, the rate's move an `fx` difference up or down), netting as a
structure with nothing new in the engine; N sides — three-way (order
lines, two receipts, the invoice line; quantity allocated, price compared
within two percent, the variance posted) and intercompany as the
generality test (a payment settling our receivable and their payable at
once, the elimination's evidence), the sides' declared `join` the one
addition; the grid — the `matching` notion over the remainder, every
side's open items with why the policy left each trigger item, and the
pairing verb running the same engine on a person's selection into the
same event; the host proposing a strategy — every pairing by hand says
which named standards would have found it, and the second pairing the
same way drafts the rule's next version with that step, warranted by the
two pairings, behind the gate. The order of work is walked. 2026-10-13:
the Poland pack's default policy shipped as drafts with a practice
warrant (pack pl v11: open amounts on the invoices, the bank operation
with its cash postings, the account as an enrichment, the two matching
rules, their consequences, the two grids) — the story test pays an
invoice through it; unpairing as a correcting event — a reversal the same
rules explain, open amounts restored, differences flipped, never a
deletion; the market stamp — a channel declares the market of the pack
that declares it and the runner stamps every fact it appends, so the
bank rule guards as the others do; the live measurement — both personas
DONE in every run over the scripted interview and three paraphrases (8
of 8 runs, 59 of 60 answers right on the first draft, the matching
bundle right in every voice); the two debts — the standing test that the
overruled rule never returns (a correction's reason on the record, shown
to the author), and the FA(3) check against the published schema (one
slot wrong, fixed); the three-run round over the voices — 24 of 24 runs
DONE across both personas, 175 of 180 answers right on the first draft,
every re-ask fixed, the matching bundle right every time it was drafted;
the "envelope" habit traced to its cause (schematics in the grammar
copied into the answer; prose snippets taken for the answer) and fixed
on both sides — concrete examples in the grammar, a decoder that reads
the answer the model meant. Owed next: the one-sided time window when a
customer asks.

**Documents in, step 6 shipped 2026-10-13** (see DECISIONS, "the scanned
invoice"): the first model-read channel — a foreign supplier's PDF read by
the model with a confidence, the channel's threshold deciding between the
typed event and a proposal for a person, the person's confirmation a door
that makes the reading theirs, the simulator writing the PDFs, the bench
reading their text layer. The channel map's row 2 (email) is real; row 2
(paper and photo) is the same reader over a scan without a text layer —
the model alone — when a customer brings one.

**The company key shipped 2026-10-13** (DECISIONS, "the company key"): the
partition key on every event and object, stamped from the caller's scope
at the door, inherited down every chain, kept by replay; the scope on the
context every store call carries, from the shell's request or the CLI's
environment; scoped reads see one company, unscoped reads the log;
enforcement still out, as the note says.

**The briefing shipped 2026-10-13** (DECISIONS, "the briefing"): the
narrative notion — the kernel's facts, the host's words in the owner's
voice, logged through the host's door so it never repeats itself, every
number checked against the facts, silence instead of a template when
there is no host; the Poland pack's "Poranek". The host speaks (the
third part of "the host remembers"); the queue and the hotspots follow.

**The queue shipped 2026-10-14** (DECISIONS, "the queue"): the queue notion
— what is open as a condition on the object, how urgent as a formula on
the object with today in it, the columns to show; the item's contextual
verbs inline, prefilled; the approvals waiting at the gate above the
items; an item the formula cannot rank listed last with the reason. The
Poland pack's "Do zrobienia dziś" and "Należności do pilnowania". The
hotspots follow: threshold rules raising cases that land in this queue.

## Rhea as a public free service: the host is the product, the kernel is open (KK, 2026-10-14)

KK's proposal, condensed: make Rhea a public service, free of charge,
with very limited capability at first — "like all AI services started" —
and a small but real wow: *yes, it's different*. The first version must
carry what the law requires, KSeF and JPK. Then grow in small chunks,
often; earn recognition; add the advanced capability; offer "easy
migrate"; consider Patronite or crowdfunding; cement security; then
monetise — enterprise, services, associated accountants, a marketplace,
ads. A free tier cannot run on paid Anthropic tokens, so a cheaper model
(Qwen, or Bielik for Polish) on Polish hardware. And the experience must
be stunning and different: *intent-oriented*, the concierge of the host
note, built apart from the kernel. Adopted 2026-10-14 with the
amendments below.

**Why free fits Rhea and not its competitors.** Every competitor pays
per user: a licence, a support seat, model tokens per conversation.
Rhea's expensive act is writing a rule, and a rule is written once per
situation, not once per company. The thousandth sole trader whose
purchase invoice arrives through KSeF costs almost nothing, because the
matching rule already exists in the network. The kernel itself costs a
Postgres row per fact. The business case for free is the network's
economics, and it cannot be copied without rebuilding the kernel.

**The timing.** KSeF has been mandatory for all but the smallest firms
since 1 April 2026; the smallest join on 1 January 2027, and penalties
start the same day. The switch moment is ten weeks from this note.

**Amendments to KK's order.**

- **Security moves to the top, not after recognition.** A tax tool holds
  KSeF credentials and bank statements. Before the first stranger signs
  up: sign-in, strict company separation (the company key exists;
  sign-in does not), encrypted credentials, backups that restore, a
  data-processing agreement, an incident plan.
- **Ads are struck.** A tax tool with ads loses the trust that is its
  only asset, and the money is tiny.
- **Accountants are the channel, not a later option.** Most Polish
  micro-firms keep a biuro rachunkowe. Free for the entrepreneur, a paid
  seat for the accountant who oversees twenty of them. This is the
  Junior Accountant → CFO path with a buyer attached, and the first
  revenue.
- **Open source replaces Patronite.** Patronite funds people, not
  infrastructure. Open-sourcing the kernel and the packs answers "free,
  and you hold my tax data?" with "you can run it yourself"; the hosted
  network stays the service. GitHub Sponsors fits better than Patronite.
  The licence decision is early, because it shapes the repo.
- **The wow is defined before it is built.** KSeF and JPK are the ticket
  to the room, not the wow; the government's free app already issues
  invoices.

**The wow: three moments nobody can fake cheaply.**

1. *Zero data entry at the start.* Give KSeF access and a bank file; a
   minute later the host says: 47 invoices in three months, 12
   customers, two unpaid, VAT so far 3,214 zł. Nobody typed anything.
2. *"Why?" on any number.* Tap the VAT amount and see the invoices and
   the rule that produced it. No Polish tool does this.
3. *Say it once, never again.* Correct the host once and the correction
   holds ("The host remembers").

**The statutory floor.** "KSeF plus JPK" is not the full legal minimum
for a sole trader — records, ZUS and PIT advances are also required. The
narrow first user is a VAT-registered sole trader who keeps an
accountant. Rhea issues invoices to KSeF, reads purchase invoices and
the bank, matches, keeps the VAT register, and hands a ready JPK_V7 to
the accountant, who files it; the accountant is the safety net while
penalties are new. Of that list only inbound KSeF exists today. Missing:
KSeF outbound (issuing), the VAT register as a view, JPK_V7 as a file
produced from it.

**Two model jobs, two costs.** Writing rules is rare, hard, and
structured; speaking to the owner is daily, easy, and Polish. Rule
writing is shared across companies, so it can stay on Anthropic at
first. The host's words and the morning briefing go to a local model.
Bielik's current family is 3.0 at 1.5B and 4.5B, plus 2.5 at 11B — too
small to write rules in the grammar, plausible for the host's words.
Qwen 27B wrote the easy corpus 9/9 but slowly. Measure, don't guess: the
eval harness and the base-URL switch exist; one afternoon answers it.
Free hardware is not the plan: no free programme found (Cyfronet says it
serves startups; worth one email). One rented GPU server in a Polish
data centre, two to four thousand złoty a month, should carry the first
thousand daily users on the shape above — an estimate to measure, not a
fact.

**The face: intent-oriented means the host shows, the owner decides.**
Built apart from the kernel, against the JSON API, in its own repo.
Three rules. A chat box is not stunning in 2026; the home screen is a
short list of decisions, one tap each, no forms, chat only as the escape
hatch. The face stays generic: it renders notions (queue, briefing,
document, case, grid), the Poland pack supplies the words, and a screen
that knows what an invoice is does not get built. The plain shell stays
alive as the developer's truth, so the beautiful face can never show
what the kernel did not say.

**The order, against the growth plan.** Phases 1–3 are unchanged; the
demo of April 2027 stays the date. Added: the three wow moments become
scenes of the demo story; KSeF outbound and JPK_V7 join Phase 2 as the
statutory floor; the security floor is the entry condition of Phase 3's
private beta — ten to twenty sole traders from KK's network, free, by
invitation; the public free tier opens in Phase 4 only when real people
finish the way the simulated ones do; the accountant seat is the first
revenue; "easy migrate", the marketplace and enterprise wait behind it.

**Decided the same day (KK, 2026-10-14).** (a) Open source it is, and
the code must meet a standard that is not laughed at: before the repo
opens, a quality floor — formatted, vetted and linted clean, a README
that explains the three-part split, a licence, and no secret anywhere
in the history (checked 2026-10-14: `.env` was never tracked and no key
is in the log). (b) Security first, in the sense of *defensible*:
sooner or later someone asks, possibly an auditor, and every answer
must already be on paper — what data is held, who can reach it, the
threat model, the log as evidence. (c) April 2027 is the *latest*, not
the date. Sequence from here: the model measurement first, then KSeF
outbound.

**Measured the same day (2026-10-14, laptop RTX 5090 24 GB, Ollama,
one run each).** Writing rules: qwen3.6:27b easy corpus NOT DONE (all
nine approved, residue zero, but the invoice state 0 of 5 — 9/9 on
2026-10-10, so one run is not a verdict) and operations DONE in 474 s
with one re-ask; Bielik 11B v3.0 (Q8_0) 0/9 and 0/6, eight drafts
refused for writing the effect as a string; Bielik 4.5B v3.0 (Q8_0) 0/9
and 0/6, refused at the gate. The host's words, five mornings each
through the live check: qwen3.6:27b 5/5 accepted and truthful — both
waiting payments named, every number exact, each morning phrased anew —
at 40–78 s thinking, and 5/5 at 3–10 s with thinking switched off (a
request parameter the agent does not send yet); Bielik 11B 0/5 — fluent
Polish in 3–8 s, but the decimals dropped from every amount, the same
text five times, and "no outgoing payments waiting" when one was;
Bielik 4.5B 3/5 by the checks, yet every answer denied both waiting
payments and opened with a template date. Two conclusions and a gap.
Rule writing stays on Anthropic; no local candidate writes the grammar,
and Bielik is out for both jobs at its current sizes. The host's words
can run locally on a Qwen-class 27B with thinking off, at a few seconds
a briefing on a laptop, which settles the free tier's cost shape. The
gap is the law's: the kernel refuses a number the facts do not carry
but not a fact the words deny — a non-empty fact must be spoken, or the
words are refused; that check is the next small step, before KSeF
outbound. (Shipped 2026-10-15, DECISIONS: the second law — a fact with
an amount in it must be spoken — and thinking off for the words.)

One line: *free is the network's economics, open source is the trust
answer, the accountant is the buyer, and the wow is zero data entry,
"why?" on every number, and a host that never asks twice.*

## Documents out: KSeF outbound, the first statutory channel out (KK, 2026-10-15)

KK: the KSeF phase touches integration, presentation, performance and
onboarding; can Novi's company account be used; and — the point that
shapes the product — an ordinary owner has to be walked by the hand
through the Ministry's portal: "go there, click that". This note is the
brief; it answers the growth plan's two welding questions and names the
decisions the first step takes.

**The three environments, and which key opens which.** The Ministry runs
three separated KSeF 2.0 systems (official integrator docs,
`srodowiska.md`). *Test* (`api-test.ksef.mf.gov.pl`) is public and
fictitious: anyone authenticates as any company with a self-made
certificate, the Ministry asks for random NIPs because the data is shared
between integrators, nothing there has legal effect. *Demo*
(`api-demo`) is configured like production, needs real credentials (a
Trusted Profile or a qualified signature of the company's
representative), and invoices there have no legal effect; real entity
data and production invoices must not be sent. *Production* (`api`) is
the real thing. So: integration and performance are built on test with
a random NIP and no account; onboarding is rehearsed on demo with Novi's
account, KK authenticating as Novi's representative; production is not a
test but the decision to make Novi the first real company on Rhea, with
Novi's accountant as the safety net — recommended as the private beta's
first member, after demo. Keys and tokens from one environment do not
work in another, which is the safety.

**Onboarding is one paste, and the walk to it is data.** KSeF 2.0 keeps
tokens (the Ministry reversed the planned deprecation; 1.0 tokens do not
carry over). The owner logs into the Ministry's taxpayer app with their
Trusted Profile, generates a token with the invoicing permissions
(`InvoiceWrite`, `InvoiceRead`) and pastes it into Rhea once; Rhea
authenticates with it (a challenge, the token encrypted with the
Ministry's published key, a short-lived access token) every time it
speaks. The walk — "go there, click that" — is a scenario in the Poland
pack, the host narrating each step, not a page in code; the first draft
of it is the walk KK takes on the test portal to produce the first test
token, written down as it is taken. A token generated on the test portal
behind "test authentication" needs no account at all.

**Secrets never enter the log.** A channel declaration is data, versioned
and approved, and every version of it is an event — so a token in a
declaration would be a secret kept forever in an append-only log. The
declaration names what the channel needs; the binding resolves it from
outside the log: today the environment (`RHEA_KSEF_ENV`, `RHEA_KSEF_NIP`,
`RHEA_KSEF_TOKEN`), later the encrypted secrets store that is the first
item of the security floor. The adapter runs with the Ministry's fake
stand-in when no token is configured, so every test and demo keeps
running without one.

**The flow, as the law of the channel.** One pass of the `ksef` adapter
is one interactive session per company: authenticate, open a session
declaring FA(3), encrypt each pending invoice with a fresh AES key
wrapped with the Ministry's key, send, poll each invoice to its final
status, fetch the UPO of each accepted one, close. The evidence is the
event: `ksef.invoice.submitted` carries the KSeF number (35 characters,
CRC-8 checked), the UPO's session reference and the UPO XML itself; the
append-only log is the evidence store, as the adapter contract says. A
rejection is evidence too: `ksef.invoice.rejected` carries the code and
the Ministry's description and details (a duplicate is code 440 with the
original KSeF number), and the pack turns it into a `ksef_submission`
with status *rejected* and a task in "Do zrobienia dziś", so a person
decides. An invoice is re-sent only when its XML changes (the event
carries the invoice's hash); a transport or authentication failure fails
the pass, and the next pass retries — retry is free by the contract.

**Performance is the limits, not the speed.** Per integration the
defaults are 120 session openings and 180 invoices an hour (June 2026
annex), raised on application for demo and production, with batch mode
recommended for volume. One session per pass per company keeps a sole
trader far below them; a hosted service with a thousand companies will
need the clocked pass to spread sessions and, above that, the batch
session — the welding question of partition is answered by the company
key (one session per company context), and containment by the channel
(the limits are the channel's, declared as data when the time comes).

**Presentation.** The KSeF number and the UPO become evidence on the
invoice, shown generically by the document and list views (the pack's
"Wysyłki KSeF" exists). An invoice shown outside KSeF needs the two
mandated QR codes on its visualization — a rendering addition to the
same view, after the floor.

**Verification.** The Ministry's checks (`weryfikacja-faktury.md`): UTF-8
without BOM, schema-valid against the declared FA(3), no processing
instructions, the issue date not after acceptance, duplicates by seller
NIP + invoice kind + number kept ten years, 1 MB per invoice, 10,000 per
session. The test environment is the validator of record; the live check
`TestKSeFLive` (gated by `RHEA_LIVE_KSEF=1`) sends one invoice there and
prints what came back.

One line: *build on test with a random NIP, rehearse onboarding on demo
with Novi, go to production only as Novi's real first month; one paste
onboards, secrets stay out of the log, and the evidence — number, UPO,
rejection — is an event.*

## Local packs, clean kernel: market code is peripheral (KK, 2026-10-15)

KK, after the first invoice reached KSeF: "I just don't want to
contaminate Rhea with local specific solutions. Later, if we fix the
Polish pack, we fix the Polish pack, not Rhea's kernel." All EU countries
will need such local packs, Germany and France first. Adopted as an
invariant (CLAUDE.md, 8), with the record and the plan below.

**What the record already said.** On 2026-10-03, closing M3: an adapter
is a user, not a kernel extension — it reads objects, talks to the
outside world through its own client, returns events; the kernel never
calls out and replay never re-runs it. Germany then proved a market can
be pure data: Poland and Germany run on byte-for-byte the same kernel.
By dependency the KSeF code was never in the kernel: the kernel packages
do not import the adapter package, the Poland pack declares the channel
as data, and the registry binds it to code by name and version.

**What it lacked, in KK's sense.** The KSeF code was compiled into the
one binary, in the shared adapter package beside the EU-wide readers; the
CLI had a verb naming a Polish institution; and a "local pack" was data
only, so the one piece of code a country brings had nowhere to live but
the kernel's repository. Twenty-seven countries would make a binary full
of every tax office in Europe.

**Move one, done the same day.** `markets/pl` holds KSeF in, KSeF out and
the FA(3) writer; `internal/adapter` keeps the contract, the runner, the
registry, the clock and the readers that are EU-wide anyway (MT940, the
model-read PDF). A market registers itself at init through
`adapter.Register` (what the binary carries) and `adapter.Provide` (an
adapter the CLI, the shell or a corpus may run by name, its client built
from the environment); the CLI's Polish verb is gone — `rhea pass ksef`
is the generic form. The binary links each market by one blank import,
and nothing else in the kernel names one. `markets/markets_test.go`
fails the build of any kernel package that imports a market. Two named
exceptions, both outside the kernel: the binary's one import line, and
the simulator, which writes a Polish persona's documents through the
Polish writer — a bench tool's debt, to move into the market when a
second market's personas exist.

**Move two, when Germany lands: adapters out of process.** The registry
already separates "implemented" from "connected"; the implemented side
becomes adapter programs discovered at startup, speaking the contract
over a local protocol (a pass asks for objects by type and files by
hash, returns events). A country pack is then data plus its own program,
versioned by the code version the channel declares, and an open-source
contributor ships France without touching the kernel repository. That is
the full meaning of "local pack"; the seam exists today, the protocol
waits for its second customer.

**Move three, the one thing that belongs in the kernel: the invoice's
meaning.** Every EU mandate but Poland's is built on the same semantic
invoice, EN 16931 — Germany's XRechnung and ZUGFeRD, France's Factur-X,
the Peppol network, the EU-wide ViDA rules from 2030. Poland's FA(3) is
its own schema. So the invoice's semantics belong in the kernel's
language once, and a market supplies only its syntax (a writer) and its
transport (a channel). Without that, every country pack reinvents the
invoice. Germany receives e-invoices since 2025 and issues from 2027 for
larger firms, 2028 for all; France receives from September 2026 and
issues from 2026 or 2027 by company size, through certified platforms
with reporting on top — dates to verify when each market starts.

One line: *a market is data plus its own code behind one contract; the
kernel imports no market, and a test says so; the invoice's meaning is
the kernel's, its syntax and transport are the market's.*

## Month end: the VAT settlement, the JPK_V7M file, the close (2026-10-15)

The second half of the statutory floor (the public-service note): Rhea
keeps the VAT register, settles the month, and hands the accountant a
ready JPK_V7M, who files it. KSeF outbound closed the first half this
morning. This note is the brief; it answers the growth plan's two welding
questions and keeps invariant 8 — the settlement is a rule, the file is a
market's writer, and nothing of either is kernel code.

**What exists.** Both VAT registers as analysis views (DuckDB, by month and
rate); postings refused into a locked (book, month) by `period_lock`
objects; `end_of_month` and `add_months` in the formula language, folds
over keyed, capped collections with `where` and reads through links; the
KSeF number and the UPO on every issued sales invoice since this morning.
What is missing: the settlement as a kernel fact, a door that closes a
month, the file.

**The law, in one paragraph.** JPK_V7M(3) is in force for periods from
February 2026 (published 19 December 2025 in the national repository,
schema `http://crd.gov.pl/wzor/2025/12/19/14090/`): one XML with a header
(form, purpose, tax office, year, month), the taxpayer (a natural person
with NIP, names and birth date, or a legal person with NIP and name, an
e-mail), the declaration (positions P_10…P_69: bases and output VAT by
rate — P_19/P_20 for 23%, P_17/P_18 for 8%, P_15/P_16 for 5%, P_13 for
0%, P_10 for exempt — P_38 total output VAT, P_42/P_43 purchases and
input VAT, P_39 the excess carried in, P_48 total input VAT, P_51 VAT to
pay, P_53 the excess, P_62 the excess carried forward), and the records
(every sales invoice with the buyer, the number, the dates, its K_
amounts by rate, and the KSeF number — or the mark BFK for an invoice
issued outside KSeF, DI for a document that is not an invoice, OFF for
the offline mode; every purchase invoice likewise with K_42/K_43; control
sums). The KSeF number in the records is required from 1 August 2026.
The brochure of February 2026 is the field-by-field spec.

**The settlement is a rule, not code.** Every invoice gets a `vat_period`
at booking — the month it settles in, `month(date)`, a market-neutral
field on the base types (a purchase invoice's deduction month is a
choice the owner may amend later; the default is the invoice month). A
door, `close_period`, is a common verb (packs/common/periods): a person
names the month and the books, and `period.closed` is the fact. The
Poland pack reacts with rules: `period_lock` for the statutory book, and
`vat_settlement` — one object per month whose fields are folds over the
invoices of that month by rate (`sum(i in objects(sales_invoice,
vat_period, $.period) where i.vat_rate.code = "23": i.vat)`), the excess
carried in from the previous month's settlement, VAT to pay or the
excess forward. Every number on the declaration is then a kernel fact
with provenance, explained by the rule the owner approved, reproduced by
replay — invariant 5 applied to the tax return. A declaration never
contains a number the log cannot derive.

**The file is a writer.** A market writer is the second seam beside the
adapter: `adapter.ProvideWriter(name, fn)` where fn reads objects and
returns a named file. `markets/pl` provides `jpk_v7m`: the settlement for
the positions, the invoices for the records, the company and its owner
for the taxpayer block, the KSeF submissions for the numbers. The kernel
renders any provided writer generically — `rhea file <writer>
key=value…` and `GET /api/files/<writer>?…` — and never names one. A
file handed to the accountant is a fact: the common documents pack gains
the door `issue_document` (writer, parameters) and the type
`issued_document` (name, hash, size, writer, parameters), the bytes kept
in the blob store under their hash, the event the evidence. The file is
rendered from the objects at that moment and sealed by its hash; a later
correction is a new file, a new fact.

**Validation.** The Ministry's XSD validates the file; there is no test
environment for JPK. A local validator (lxml against the national
schemas, in a scratch environment) checked this morning's FA(3) and will
check the first JPK; the story test keeps a golden file. The accountant's
own tool is the validator of record until the signed-upload API is worth
the certificate.

**Welding.** Partition: a settlement is a company's (the company key
stamps it); containment: the close is a door with a gate behind it like
every verb, and a closed month refuses postings per event — "a closed
month is a closed month".

One line: *the settlement is a rule with provenance, the close is a door,
the file is a market's writer rendered by a kernel that never names it.*

## The plan after the statutory floor: test hard, then the face (KK, 2026-10-15)

KK, the night the floor closed: defer the open-source floor; do
backup-restore now. Then go through the whole Poland pack again and
verify the rules are complete. Then a testing phase, not skimped — two
more weeks and the tokens it takes: a *business generator* that makes
test companies and document traffic and moves them over several months
and through year-end. In parallel the first-class, intent-based modern
UI, mobile first — "if we want Rhea to be different, mobile will be the
#1 interface", no shortcut. KK looks for the infrastructure platform for
live deployment. And security. The demo rehearsal with Novi's account
waits for the face: without a UI there is nothing to rehearse but the
login, and the login's only question (does Novi's KSeF authorization
work on the demo environment) can be answered in five minutes whenever
the time comes.

**The order, replacing the one in the public-service note for the
coming weeks.**

1. **Backup and restore** — the claim the kernel makes is that the
   event log plus the definitions rebuild everything; a test restores
   from that alone into an empty database and replay reproduces identical
   state. Done the same night (DECISIONS).
2. **The Poland pack, read whole.** Every rule, type, view, door and
   channel re-read against the story a sole trader lives through in a
   year: what is missing, what is wrong, what the corpus never asked.
   The output is a list with a verdict per item and the pack version
   that closes it.
3. **The business generator.** The simulator grows from two personas
   with one scene each into a generator of companies: a shape (trade,
   size, customers, suppliers, banks, currencies), a calendar (months,
   quarter ends, the year end with its closings and filings), and
   traffic (invoices in and out, payments, statements, corrections,
   late arrivals, duplicates, the odd mistake). Every company runs
   through the whole floor — KSeF out and in, matching, the month end,
   JPK, the year end — and the verdict is the one the personas already
   have: DONE or not, per run, per voice, with every number explained.
   The generator is also the demo's content.
4. **The face, mobile first.** Built apart, against the JSON API, in its
   own repository, intent-based: the home screen is the decisions of the
   day, one tap each; the briefing; "why?" on every number; the doors as
   forms the pack declares; chat only as the escape hatch. The plain
   shell stays the developer's truth. The mobile screen is the primary
   one, not an adaptation.
5. **Security**, with the platform KK finds: sign-in, strict company
   separation on the company key, encrypted secrets (the KSeF token
   first), backups that restore (item 1), the data-processing agreement
   and the incident plan on paper before the first stranger.

The open-source floor (licence, README, formatting) comes after these.
KK, the same night: the kernel stays open source in the strategy, and it
opens the same day the first version opens to the street — the floor is
done before that day, not before the testing phase. The shape KK
proposes: open the kernel and the base Poland pack; keep the face and
perhaps some specialised packs proprietary. Agreed, with three lines
drawn. The open part must be enough to run alone — the kernel, the plain
shell as its working face, the common packs, and the base Poland pack
*with* its statutory channels (KSeF in and out, JPK), because "you can
run it yourself" is the trust answer and a sole trader who runs it needs
exactly those. The pack format and the loader stay open so anyone can
write a pack; what stays proprietary is the face, the industry packs
(warehouse, intercompany, trade-specific rules) and the hosted network
of learned rule shapes — the network is the moat, not the packs. The
boundary with the proprietary face is the JSON API, in another process
and another repository, so a copyleft licence on the kernel (AGPL) never
reaches it; kernel code is never linked into the face.

One line: *restore first, then read the pack whole, then generate
businesses and test for two weeks without skimping, while the mobile
face is built apart and the platform is found; security before the first
stranger, and the kernel opens the day the street does.*

## Reading the Poland pack: the law's open questions, two kinds of company (KK, 2026-10-16)

KK, on item 2: much of the Polish law the pack implements is ambiguous,
and the hope is that Rhea helps create more accounting standards in
Poland. And the first deployment phase will have sole traders as users,
so the pack must support them — as a variant of the Poland pack, not a
separate pack.

**How the pack is read.** Not rule by rule but question by question:
walk the year of two companies — a sole trader (KPiR or ryczałt) and a
small sp. z o.o. on full books — and list every point where the books
must decide something. Each point gets a grade: *1 plain* (the text
decides; cite the article), *2 settled* (the text is unclear but the
Ministry's explanations, a general interpretation, an NSA resolution,
the TSUE or a KSR settled it; cite that), *3 open* (practice splits;
Rhea ships a reading, names the alternatives and who holds them, and
the client chooses). The sources, strongest first: the consolidated
statute (ISAP, ELI) and the VAT Directive with the TSUE behind it; the
Ministry's own reading (objaśnienia podatkowe, general interpretations,
the KSeF and JPK manuals, the XSDs — the schema often settles what the
statute leaves open); individual interpretations (Eureka — binding only
on the asker, but the best map of where taxpayers were unsure); the
courts (CBOSA); the KSR and the Committee's positions; practice
(commentaries, SKwP, KIDP, one practising accountant). Every claim is
quoted from a source, never recalled. The register lives in
`packs/pl/REVIEW.md` with a verdict per point (right, wrong, missing,
later) and the pack version that closes it; the grade-3 list goes to a
practising accountant as business questions, not as rules.

**Every rule carries its own warrant** (shipped 2026-10-16): the pack
had one warrant for 29 rules, written when it booked only the bank. Now
a rule names its source and the reading it chose, the warrant travels
with every version, and "why?" on any number ends at the article.

**Sole traders are a variant of the Poland pack.** Most of the pack is
the same for both: invoices, VAT, KSeF, JPK_V7, bank matching,
contractors. What differs is how revenue and cost are recorded (KPiR
columns, the ryczałt record, or accounts), the income tax (PIT advances
on the scale, linear or ryczałt; CIT), the owner's ZUS, and the year
end. The form of books and of income tax is a dated fact about the
company, said through a door the way its legal form is: a sole trader
who crosses the threshold moves to full books on 1 January by one more
event, and replay shows which form held when. The one kernel step it
needs is already on the destiny list: a rule's match conditions reading
through a link to the company record (today they see only the event —
the reason `market` is stamped onto every payload). Sole traders are
the case that earns it.

**How Rhea helps make standards.** Two legal doors exist. The
accounting act (art. 10 ust. 3) lets a company apply the national
standards and positions of the Accounting Standards Committee where the
act is silent, and IFRS only where no KSR exists — the act expects its
gaps to be filled by standards, and someone has to propose them. The
Tax Ordinance (art. 14a § 2) lets anyone but a public body ask the
Minister for a general interpretation, and the request must show the
law applied inconsistently in decisions, judgments or interpretations
on the same facts. Rhea brings what a standard-setter lacks: the
register of open questions with the competing readings and their
sources (the evidence § 2 asks for); later, real counts of how
companies actually decided (counted approved rules, never estimates);
and readings that run — replay a company's year under reading A and
reading B and show the difference in the tax due, before anyone adopts
either. A company's approved rules with their warrants are also its
written accounting policy (art. 10 ust. 1), which Rhea can print. And
when the base Poland pack opens, its readings become public commentary
practitioners can contest and cite. Rhea's readings are always proposals
with their sources, never the law; the client's approval stays.

One line: *read the pack as the year's questions, graded by how settled
the law is; every rule names its source; the sole trader is a dated fact
about the company, not a second pack; and the open questions, written
down with their readings and run on real books, are what Rhea brings to
the people who write Polish standards.*

## The face: how it looks, the shapes objects take, talking about them (KK, 2026-10-16)

KK, on item 4 of the plan: the face cannot be avoided, and besides the
AI idea it must engage and amaze — different, elegant, simple but
complete. Objects will stack up (documents, accounts, journals and
transactions, T-entries, the trial balance, the statements, cash flow,
aging, fixed assets, clients, suppliers, partners, employees), and
people will want to comment on them and discuss them. KK brought a list
of the 2026 interface trends; this is the reading KK agreed to.

**The trends, read against Rhea.** Half the list is fashion that will
date. Glass behind figures hurts reading; a loud palette fights the
colours bookkeeping already uses for meaning (overdue, open, matched,
refused); spatial computing misses people who run a business from a
phone. The other half is what Rhea already is. Generative UI (Google's
A2UI: the AI composes a screen from a fixed set of safe components) is
the view the agent drafts and a closed set of surface primitives draws,
with a record of every step. Agents doing tasks is the host. Explainable
AI is provenance — the one thing nobody else can show, because nobody
else keeps it — so it is the face's main visual idea, not a footnote.
Calm technology is the decisions of the day: a quiet day is a nearly
empty screen, and empty is the luxury. Kept with a twist: depth only
where it means something (a layer sits above the thing it came from);
the bento grid for the morning briefing only, never for a ledger; the
human touch as the trade's own materials — the ruled ledger, the
T-account, the stamp, the signature — not fake roughness; voice and
camera for getting things in (a photo of a receipt, a forwarded PDF, a
spoken answer to her question), not for moving around.

**The look comes from the concierge.** Luxury is restraint: space,
beautiful type, exact figures, nothing that shouts. A calm warm ground;
dark figures in a typeface whose digits align; one colour that belongs
to her, used rarely. Depth only for "where it came from", motion only
for "something happened". Three moves carry the difference:

1. *The number opens.* Any amount, anywhere, unfolds layer above layer
   into its origin — the account, the entries, the document, the
   original file, the rule and its approval, the article in its warrant
   — as one continuous zoom down and back, never a jump to another page.
   `/api/explain` is the walk; the zoom takes it one layer at a time.
2. *Things find each other.* A payment slides onto its invoice and they
   become one settled item; a closed month looks closed.
3. *The sentence is the dashboard.* The briefing is prose whose every
   figure is live and opens like any other number — the host's second
   law already makes every spoken figure a kernel fact.

**Six shapes, not thirty screens.** The face never learns what an
invoice is: each object type takes one of a few shapes, the pack says
which, and designing the shapes well styles every object that will ever
exist — the closed primitive set of the design language, grown by the
shapes objects take. *Card*: who or what it is, key facts, what is
open, its history (client, supplier, partner, employee, fixed asset,
account). *Paper*: the document as a document — header, lines, totals —
with the original file beside it (invoice, PZ/WZ, bank statement, JPK,
KSeF receipt). *Ledger*: dense rows, aligned figures, running balance,
grouped by day (the journal, transactions, an account's transactions).
*Statement*: a tree of totals that folds and unfolds, each line opening
into its ledger (trial balance, income statement, balance sheet, cash
flow). *Buckets*: a ledger split by age or period (aging, VAT by
period). *Chain*: one business event as linked steps (order → WZ →
invoice → payment → match), shown on every object so the gaps stand
out. The T-account is not a place to browse but the explanation of one
booking — "this is what this invoice did to your books": for the owner,
money moving between named pots; one switch shows the accountant debit
and credit. No network graph: it wins the demo and loses Tuesday.

**Talking about objects.** Anyone comments on an object, or on one
figure on it ("why is this in 761?"). A comment is a fact like any
other: in the object's history, through replay. She is in every thread,
and "why?" is answered at once from provenance, before the accountant
has to reply. A thread can end in a decision — "book this supplier to
402 from now on" becomes a draft rule at the gate, the correction she
remembers; that is what "say it once, never again" looks like on
screen. It matters because the accountant is the channel: the email and
photo traffic between an owner and their accountant moves onto the
objects themselves. The kernel has no comment door yet; it will be a
common door and its rule, like any other.

**Two people, two densities.** The owner on the phone: decisions, the
briefing, "why?", photos in; a statement there is one column of amounts
in a folding tree, the last period a swipe away. The accountant at the
desk: density, keyboard, ledgers side by side, many companies — where
the paid seat lives, so not an afterthought. The same six shapes at two
densities; the phone is the primary screen, not an adaptation.

**The first build.** "The number opens", on a phone, against the real
JSON API, on a simulated company's data. If that one moment feels right,
the rest of the face follows from it.

One line: *the look comes from the concierge, not from the trend list —
calm, exact, one colour of her own, depth only for where a number came
from and motion only for what happened; every object takes one of six
shapes the pack declares; comments live on the objects and can end in a
rule; and the first thing built is a number that opens.*

**The first build (2026-10-16).** KK: "let's not wait". One phone page
in its own private repository (`~/rheaface`, MojoHand70/rheaface),
Polish, on Helios Studio: nine months through the real kernel and pack
v22 into the database `rhea_face`, captured from the JSON API — objects,
views, every explanation chain, rules with warrants, doors, approvals, and
the host's real morning words. Today is the decisions (close September,
say what the stray 100 zł is), her words with all 27 figures live, the
morning's figures; any figure opens out of itself into its account, its
entries, the document, the fact, the rule and the article. The September
close shown on the page is the kernel's own result, captured ahead; the
page writes nothing. What building it found the kernel owes the face:
(1) "why?" on a view cell — explain answers for objects, but a register
cell or a trial-balance balance has no composition in the API, so the page
composes it from the period's invoices or the account's postings and
checks the sum; the kernel should serve it. (2) The shapes, the
statements (which positions make the VAT settlement) and the words (field
labels, the chart's group names) are a stand-in table in the page; they
belong in the pack. (3) Comments have no door; the thread on each object
is the page's own, her answers composed from provenance without a model.
(4) Her morning words pass both laws but list all fifteen open invoices
one by one — a concierge says the oldest and the total; the narrative's
voice wants a brevity rule. (5) Unpaid supplier invoices since January
have no door to offer as a decision (nothing says "paid outside the
bank" or "remind"), so the page can only show them. The page's own check
also surfaced a true gap worth showing: open sales invoices 101 475,00,
account 201 101 375,00 — the 100,00 is the stray payment sitting on 201
with no invoice, until someone says what it is.

**KK's first impressions (2026-10-16): intent and usage, not only
decisions.** The look holds. Lists lead with the counterparty, the
document number below. Two corrections to the brief. *Chat is not
hidden*: "chat only as the escape hatch" was too strict — the
conversation is one tap away on every screen, because it is where the
person says what they want. *The face learns from use*: what a person
asks often becomes a one-tap function, and a person shapes the face by
saying so ("show the counterparty first"). Built the same day: Rhea in
the middle of the tab bar; the model only picks one act from a closed
set (a figure, a list, open a thing, the list layout, keep a shortcut, go
to a screen, or nothing) and the face shows it from the books, so her
words carry no digit — the agent proposes, the deterministic side shows,
the same split as the kernel's; the second identical ask, she offers to
keep it as a shortcut on Today, which shows its live figure and opens
without the model; "what I remember about you" lists it all and forgets
on request. In the page the browser keeps this; in Rhea it is the host
remembering: a person's preference or shortcut is a fact in the log
about that person, a correction supersedes it, and nothing about it
needs the gate — while a change to what everyone sees (a shared view,
a new column the pack declares) is a new view version drafted and
approved like a rule. What the kernel owes for it: a person-scoped
preference door (layout, shortcut, its withdrawal) and a view a person
can own.

**From the prototype to the product (KK: "let's go", 2026-10-16).** The
drawing carries over; what changes is underneath. Five steps, in order:
(1) the kernel answers everything the face reads — the objects with
their provenance, who approved each rule, the host's words, "why?" on a
view's cell — so nothing is read past the API (done the same day,
DECISIONS); (2) the face talks to the kernel live, asking for what is on
screen, from its own small server in front of the kernel, and its
decisions write through the doors that already exist; (3) the pack
supplies the words — field labels, the shape each type takes, the
statements (the VAT settlement's layout, profit, ageing), and the
accounts' roles (which is the bank, which the receivables), so the face
computes no figure of its own; (4) Rhea understands words on her own
side: the "understand, then pick one act" step behind one endpoint of
the kernel, the same closed set and the same no-digits check, a
candidate for the local model; with it a person-scoped preference door
and a comment door; (5) installable on the phone as a web app first
(native only if that proves too little), behind the security floor. Then
the accountant's desk density. KK decided the same day: **a web app**
(installable from the browser, no app store); **Anthropic** understands
the person's words; **hosting on white01 at OVH** for now. The kernel
stays bound to the machine; the face's server is the only door from the
internet, and it opens a closed set of the kernel's routes.

**Step 2 done the same day, with the understanding step of step 4.** The
face's own server (Go, in the face's repository, no kernel code) serves
the web app — installable, with a manifest, an offline shell and her mark
as the icon — opens only the reads the page needs and three doors (close
a month, say what a bank line is, a supplier's cost account), signs every
decision with the configured person, keeps the company scope its own, and
refuses to listen beyond the machine without a password. The page boots
from the API and fetches a chain, a fact or a register cell only when a
sheet opens; the kernel's live stream refreshes what is on screen.
Decisions write: September closed through its door gave the same VAT the
kernel computed before; the stray 100 zł classified as the owner's money
moved off the receivables. One call placed differently from the plan:
*the understanding step lives in the face's server, not the kernel* — the
acts it chooses among (a figure, a list, a layout, a shortcut, a screen)
are the face's vocabulary, the kernel stays smaller, and the open kernel
needs no model to run; the model is Anthropic's (`FACE_MODEL`, default
Claude Opus 5.5 at low effort), its answer is data the page checks
against the closed set, and her words still carry no digit. Found on the
way and fixed in the kernel: two analysis reads at once collided on the
DuckDB rebuild (DECISIONS). **Live on white01, 2026-10-17: https://app.rhea.plus.** White01 is
shared with Novi's running services, so Rhea took nothing of theirs: its
own system user, its database in its own container, kernel and face
listening on the machine only, and one site added to the Caddy that
already serves Novi (validated before reload; Novi answered throughout).
The name is an A record in Cloudflare, DNS only, so Caddy holds the
Let's Encrypt certificate and the live stream is not buffered. The books
there are the Helios simulation, moved with the kernel's own backup and
restore (626 objects rebuilt). The kernel cannot be reached from outside;
the face asks for a password until sign-in arrives with the security
step, and signs every decision as krzysztof.

**Step 3 done, 2026-10-17.** The face names nothing and computes no
figure of its own. The Poland pack (v24) carries the words — every type,
field, value, event and door in Polish, the shape each type takes, the
groups of the chart — as a `words` view; the VAT declaration as a
`statement` view the kernel fills with the settlement's own figures; and
three reports the face used to add up itself: the morning's figures,
profit since January, receivables by age on the kernel's day, each row
naming what it is made of so the face opens it. The kernel's figures
matched the face's old sums to the grosz. Live on white01 the same day,
the books untouched. Still the face's: its own chrome (tabs, sentences
around the books), presentation (bar widths, a ledger's running balance),
and which views make which screen — by the views' names, for now; the
pack declaring its screens is the next refinement. Found on the way: the
host sometimes writes amounts the Polish way (44 900,00) and the face read
only the last group — fixed in the face.

## Teaching her a figure; what she was never taught reaches the house (KK, 2026-10-17)

The case. KK asked the face on his phone „Możesz pokazać wynik za sierpień
tylko?” and she answered that she shows the result only since January —
true, because the Poland pack had no monthly result. Pack v25 taught it
(approved by KK on his phone, the first approval there). KK's three points:

- **She names the missing lesson.** When she cannot, she says she was never
  taught it and that it is not in the pack she works with — not a nearby
  answer passed off as the asked one. Shipped in the face the same day.
- **A person teaches her.** Today a figure is SQL a pack's author writes by
  hand; the AI author may not draft it, because model-written SQL run
  against the books is model output executed (invariant 3). So a figure
  gets a fill-in shape — a **tally**: what is added up (the bookings, or one
  document type, and which amount), which of them (conditions on fields,
  one step through a link: the booking's account is revenue), grouped how
  (a field, or the month of a date; two at most), in which lines (+, −, and
  = for the lines above). The kernel writes the query from checked parts,
  so a drafted tally is data. Then: she says she was never taught it, offers
  „Mam się tego nauczyć?”, the author drafts the tally, the card waits on
  Today, the person approves, and she shows it from then on.
- **Rhea learns from every client what she was never taught — with people
  in the loop.** Residue across the network is demand (above); so is every
  „nie nauczono mnie”, and it comes with words. An untaught moment is a fact
  in the installation's log (the person's words, which stay home), and her
  general description of the missing skill reaches the network, where the
  Rhea team reads them across installations, counted. The loop closes three
  ways: the owner teaches her there and then; the team ships the lesson in
  the pack (v25 was exactly that, by hand); the network learns the shapes
  people taught (as it learns rule shapes) and offers them as priors.

**The standard proposed for what leaves** (Rhea suggests standards): her
general description only, never the person's words; refused at the door if
it carries a digit or a name the installation holds (a company, a person) —
then it stays home, marked unshared; the owner sees everything she reported.
A client's consent to report is part of onboarding (KK, 2026-10-17): asked
once when the client joins, not built before then.

**Steps.** (1) The tally in the kernel: compiled on every read of an
analysis view (the view, "why?" on a cell, the host's facts); proved by
writing the monthly result as a tally and getting the pack's own rows,
booking for booking. (2) The author drafts tallies: the grammar shows the
shape, validation checks it against the catalog with the bundle's new
types. (3) The untaught moment: a door that records it, the network that
hears the description, `rhea network` listing what installations were never
taught. (4) The face: her reply says what is missing, the offer to learn,
the draft through the kernel's draft door, the card, and taught reports in
what she can show.

**Steps 1–4 done, 2026-10-17** (Rhea 4123502, 0f637e0, 4bbca7b, b46e4c8;
rheaface d8d7460). Found on the way and answered: the author once drafted
"sales by customer" over bookings grouped by account code — valid, wrong —
so the grammar learned which to add up, and the card shows the figure's rows
computed now (`/api/bundles/{id}/preview`) with „Nie tak” and a reason, which
refuses the draft and has her redraft answering it. Live on Helios the whole
loop ran: untaught → recorded and shared → drafted → previewed → corrected →
approved → shown. **Still open:** the third way the loop closes — the network
learning the shapes people taught and offering them as priors, as it does
rule shapes. (Consent to report is onboarding's, KK.) **The Learning
Centre** (KK's name, 2026-10-17; Rhea 1347c0e, rheaface f36ee00) is the
team's page: at `/learning` beside the face, what installations were never
taught, counted, each with where the team says it stands (open, planned for
a pack, done, not for us) and a note, and what Rhea has learned across
installations — read from the network, never from an installation's books.

## Reporting: she queries the books and lays them out (KK, 2026-10-17)

KK, the morning after the tally: Rhea's reporting should match a complex SQL
query over a collection of DuckDB databases, with financial visualisation —
KPI dashboards, financial statements, listings, pivot tables — and a report
may have several visual components. Teaching her `count` as one keyword of
many is the wrong road: use the model's ability to write the query, and
visualise what comes back.

**KK's decisions.** (1) Invariant 3 reworded (CLAUDE.md): her output is
data — drafts activated by a person, and read-only queries the kernel runs
locked on the analysis copy; model output never runs with authority. (2) A
question is answered at once, without approval, when it only reads; what is
kept as a report goes through the gate. (3) The collection is the
business's own databases of objects — stock, workforce, sales quantities, a
real-estate firm's buildings — whatever object types it has.

**The collection.** One DuckDB schema per domain, one typed table per object
type (`sales.sales_invoice`, `finance.posting`, `warehouse.stock_item`,
`real_estate.building`): a column per field with its type, money as
DECIMAL(18,2) (invariant 6), dates as DATE, links as the linked object's id,
the provenance beside every row; the JSON `objects` table stays for the
packs' hand-written SQL. A type the business teaches her is a table the next
time the copy is built. The copy is rebuilt only when the log has moved.

**The locked session.** DuckDB opened read-only; external access, extension
install and autoload off; configuration locked; the query parsed and refused
unless it is one SELECT; a time limit, a memory limit, a row cap. What a
query can reach is the copy.

**Two speeds.** Asked now: the author writes the query and a plain reading
of what it counts, the kernel runs it locked and records the answer (what
was asked, the query, the reading) so what she told someone can be read
back; the face shows it marked as computed on request, with "why?" when the
rows name their objects. Kept: „zachowaj jako raport” drafts a report — a
page of components, each a query and a visual — previewed on the card,
approved or corrected, then in the books for everyone and for dashboards.

**Visuals.** KPI tiles, a financial statement (sections, subtotals, signs,
negatives in parentheses), a listing, a pivot (DuckDB's own PIVOT), a bar or
line chart. The query computes; the face lays out. The tally stays for the
views already approved in it; the author stops drafting it.

**Steps.** (1) The collection and the locked session. (2) Asked now: the
author's query, the kernel's run, the record. (3) Reports with components,
drafted, previewed, approved. (4) The face's visuals and its "ask now" and
"keep it". (5) The author's grammar moves from the tally to queries.

**All five done, 2026-10-17/18** (Rhea 590a09b, 4848d75, 4b1bf01, c9a8b7e;
rheaface step 4). Live on Helios: „Kto jest moim największym klientem w tym
roku?” was answered with a query in seconds — two customers, their net and
share, as bars — kept as a report through the card and shown as a page; a
sales dashboard asked for in one sentence came back as three components that
all ran. Found and closed on the way: averages came back as floats and a
query read the machine's date — both are now refusals the author fixes.
Open: the Poland pack keeps nearly every type in one domain, so its
collection is one schema; the domains are a pack change.

## Packs built together: a workspace for the profession, and packs of firms (KK, 2026-10-18)

KK, the evening the first pack pages went behind the gate: every pack says it
was prepared by AI and is an initial version; users see packs develop; packs
are always behind a gate, also when the experiment site gives way to a normal
one; accountants and auditors will be invited to comment on pack rules and
vote toward consensus — a regular system for building knowledge together; and
audit firms might build packs of their own.

**KK's decisions.** (1) Every pack carries a stage, a disclaimer in its
language and its history; an initial pack reads its version as 0.x (done
2026-10-18, a9ef1f8: `TestEveryPackSaysWhatItIs` holds every pack to it).
(2) Country pages are in the country's language and live on rhea.plus behind
the gate, never elsewhere. (3) A full secure login with an audit trail comes
before any of the workspace is built — not now, not yet.

**What is discussed.** A rule with its warrant, and a review point with its
question, its law and its default. Each has a thread. A contribution is a
question; an objection, which names its legal basis; or a proposal — the
change in the reviewer's words, which the author drafts as the rule's next
version and the kernel previews on sample books, so a reviewer judges what
the change does to the books, never JSON.

**How it is decided.** Not by votes alone: what a statute says is not a
majority's to decide. Rough consensus with a named editor per pack: a
proposal passes when it cites its basis, enough verified reviewers agree
(three to start), and no verified reviewer's objection stands unanswered
after a window (two weeks). The editor — Rhea's team at first — records the
decision and its reasons; an objection overruled is answered in writing. A
disagreement that stays makes the point open (grade 3, both readings and who
holds each, as the registers already do) and the pack offers both readings
as the company's choice through a door, defaulting to the cautious one with
its warrant. Votes, objections and decisions are facts in the log, with the
person and their qualification.

**Who.** Verified professionals, their name and qualification on everything
they do: in Poland auditors and tax advisers against their public registers
(KIBR, KIDP) and accountants through the office they work for; in Germany
Wirtschaftsprüfer and Steuerberater through their chambers; in the US CPAs
and enrolled agents. Invited, as the room is today.

**Versions.** The review register is the checklist: a pack reaches 1.0 when
its reviewers agree the first-year floor is right, point by point, each point
carrying who reviewed it. Every version's history names the proposals that
made it and who agreed.

**Packs of firms.** The country base pack stays shared and consensus-built —
the statutory floor, open in the open-core plan. A firm's pack is a layer on
a stated base version: its methodology, industry practice and accounting
policies, every rule signed by the firm. When the base moves, the conflicts
with the firm's rules are reported as pack upgrades already report rules
evolved locally. Independence decides who builds what: an audit firm must not
author the books its own audit clients run on (the self-review threat; for
public-interest entities the EU Audit Regulation, art. 5, bans bookkeeping
and designing financial-information systems). So two roles: firm packs come
from advisory arms and accounting offices, for clients they do not audit; and
audit firms give assurance — an opinion on a base pack version, which every
company on that version benefits from and which can be what takes a pack to
1.0. Replay and provenance make auditing a company on an attested pack
cheaper: that is the case to the firms. Firm packs may be paid, with a share
to Rhea; the base stays shared. The terms say who answers for what: the
pack's author for its content, the company for approving a rule into its
books.

**Steps.** (0) Done: stage, disclaimer and history in every pack; the country
pages in their languages on rhea.plus behind the gate. (1) A pilot, if KK
wants one before the login exists: three to five invited accountants read
the Polish page, and a comment form on each rule records the comment under
the reviewer's invite (name and email) and emails it to KK — no votes; it
shows what reviewers comment on. (2) The login: personal accounts, the
qualification check, an audit trail of who saw and did what. (3) Threads and
proposals, the author drafting each proposal as a previewed next version.
(4) Consensus: votes and objections as facts, the editor's decisions with
reasons, open points as company choices, versions naming their reviewers,
1.0. (5) Firm packs as signed layers on a base version, with the conflict
report and the terms. (6) Assurance: an audit firm's opinion on a pack
version.

**Open for KK.** The consensus threshold and whether a change needs a
particular qualification (an auditor or a tax adviser among those who agree);
who the editor is after Rhea's team; what reviewers get — recognition, early
access, pay; whether a firm's pack is listed for every company or kept for
the firm's own clients.

## Smart documents: a process as one definition (KK, 2026-10-19)

*Superseded in name and definition by "Rhea Apps and business packs" (KK, 2026-10-23): a smart
document is now an App, and a Rhea app of smart documents a business pack. The language pieces,
the invariants and the proofs below stand, read with "App" for "smart document".*

KK's call, after the US register showed what an invoice line costs today
and the talk turned to Rhea's language as object modelling (classes,
objects, attributes, methods): a document should carry its own
calculation — *"a small compute programme with variables, formulas, query
capability (check other objects) and control commands. This is the anima
of an object"* — and that is the smart document KK has described before.
*"This is THE most powerful mechanism of Rhea to manage company business
processes and workflows. This is what makes Rhea a capable ERP. This way
we can manage any process, calculate any values"*: a smart document takes
values from the ledger, recalculates an accrual and posts it on approval;
smart documents make whole WMSs, OMSs and document flows, *"any Rhea
subsystem. These subsystems become Rhea apps."* The bar is big ERP, not the
sole trader: Rhea's definitions are a meta-description of an instance, and
a simple company simply needs fewer of them. **A dedicated session.** This
section is the handover that session starts from: constraints, an
inventory, proofs and an order of work, not a design.

**KK's definition.** A smart document has:

- a current state, a diagram of the states it may move between, and a
  terminal state;
- per state, the user or group who may change values, who may see them,
  and who may change the state;
- variables, scalars or collections, each set by input, by a formula, by a
  query, or by a combination; the document may loop through a collection;
- a snapshot of its variables at every state entry and exit;
- a definition and instances, an instance being a workflow started;
- the power to start another workflow, and to move another instance between
  states when the current user is entitled to;
- a declarative view;
- the ability to work on data tables and to use APIs.

All of it is declarative.

**What it is in Rhea's terms.** SPEC's Document grown up: a business paper
with an identity and a lifecycle, now also carrying its variables, its
calculation, its actions and who may take them. Read in the language Rhea
already has, one smart-document definition is one approved package of the
four vocabularies — a type with a lifecycle, the activities that act on
it, the rules that explain those acts, the views that show it — written as
a single unit and turned by the kernel into those pieces. Every guarantee
then holds by construction: the log, approval, provenance on every value,
replay. A Rhea app is a package of smart documents.

**Rhea already has most of the parts, scattered.** The Poland pack's
prepaid costs (v30) is a smart document built by hand: two types
(`prepaid_cost`, lifecycle active → done; `prepaid_release`), one action
(`spread_cost`), five rules and a view, spread through `packs/pl/pack.json`.
Nobody can see it as one thing, draw its states or have the interview
build one. The CIT advance (v27–v29) is another.

| KK's smart document | Rhea today |
|---|---|
| States, paths, a terminal state | `LifecycleDef` on a type (`internal/core/core.go`): the field and its allowed moves, refused otherwise. A terminal state is only implied (no way out). |
| Definition and instances; starting a workflow | A type and its objects; an action records an event and a rule creates the instance. |
| Variables by input, formula, query | Action inputs; the formula language (exact, declared rounding, reads through links, `balance(book, code, from, to)` on the ledger); `Amendable` fields a rule may set later. No variable belongs to the document: a value is whatever field a rule writes. |
| Collections and loops | Bounded folds and `each`, including over a keyed collection (`=objects(T, field, value)`), under `MaxCollection`. Fields are scalar or a link, so a collection is separate objects. |
| Snapshots at state entry and exit | The log keeps every change and as-of is replay with a cutoff; each computed value carries its `Calc` (formula and inputs). No recorded snapshot of all variables at a move. |
| Start another workflow; move another's state | A rule may create another object (cascade, depth cap 16) and amend another's lifecycle, the kernel validating the move. No entitlement check. |
| Who may edit, see, move | `Activity.Who` is declared and displayed, never enforced (`internal/core/activity.go`). `person` has a `role` string and a deputy. No groups, no per-state permissions, no login (CLAUDE.md: no auth; the recorded destination "Authz belongs in the language"). |
| Declarative view | View notions as data: list, detail, worklist, statement, narrative, queue, words, reports. |
| Data tables | The analysis collection: one DuckDB table per type, read in the locked session ("Reporting" above). |
| APIs | Channels: code behind the adapter contract, emitting events (KSeF in and out, the NBP rates). |

**What is new.**

1. *The definition itself* — states, variables and how each is set, the
   actions allowed in each state and who may take them, what each action
   causes, the view — and its translation into the four vocabularies,
   approved as one package.
2. *Document variables.* Named, each declaring its source (input, formula,
   query, or a formula over inputs and queries) and when it is computed:
   when the document is created, when an input it reads changes (a
   payment matched), on the clock (month end), or when someone looks, as
   of a date, without storing. A variable is assigned once per
   computation, like a line of a tax form; collections are the document's
   own parts (lines, line taxes, the delivery address).
3. *Permissions per state*: who edits which variables, who sees them, who
   may take which move — checked against the acting person, a refusal
   recorded.
4. *Snapshots*: at each state entry and exit, the variables as they stood,
   baked into the event with their `Calc`, so "why?" on any number opens
   the worksheet line by line.

**The language pieces it folds in** (agreed in talk with KK, 2026-10-17 to
2026-10-19). The boundary they serve: *anything a client asks for in an
implementation is data the interview builds; code is only for engines and
file formats* — the toolbox (exact calculation, double entry, matching,
totals over time, the clock, readers and writers, queues and tasks, her
words). A request that needs a new tool is a gap in the toolbox, never in
the client.

- *Descriptors* (business dimensions): an attribute defined once with its
  set of values, often a tree (region → country → district; a product
  category tree), attached from a library to master data, transaction data
  and accounts alike. Today every field lives inside one type. Prior art
  for the shape: SAP's data dictionary (a value range defined once, reused
  in many tables), Dynamics' extended data types and financial dimensions.
- *Bookings that carry descriptors.* The posting line is the one class the
  kernel fixes: `PostingLine` is account, debit, credit
  (`internal/core/core.go`), with book and currency on the entry. It must
  become an ordinary class with the kernel's laws kept (balance, the month
  lock), so a client's chart with segments is a natural account plus
  descriptors on the line, and a 60/40 split is a rule writing two lines.
- *Parts*: "an invoice has lines; a line has taxes", declared on the class,
  so views, totals and file writers walk from the document to its parts;
  descriptors flow down unless a part says otherwise.
- *Specialisation*: an export invoice is a sales invoice; sales and
  purchase invoices are documents with lines; a client's document extends
  the pack's without editing it. Today types are flat and versioned by name
  across packs, which is why a US `company` version would supersede
  Poland's (`packs/us/REVIEW.md`, cause 4).
- *Account roles*: rules book to a role ("receivables from customers",
  "sales revenue"), the instance maps roles to its own chart. The Poland
  pack books to literal numbers (201, 700, 222); `company.cost_account`
  and its siblings are the start of the mapping.

**What the invariants demand of it.**

1. *The document never writes the books itself.* It reads, computes and
   decides; an action records an event; a rule books it. In KK's accrual,
   the document reads the ledger and computes; "approve" is an event; the
   posting names the approval, the rule and the version.
2. *Calculations always finish.* Bounded iteration over a named, capped
   collection; no `while`, no recursion (the arithmetic brief, agreed
   2026-10-08). Named algorithms (FIFO, allocation with the remainder,
   MRP) stay kernel methods a document calls.
3. *An API call is two facts*: the request out, the answer in, both in the
   log through a channel. A calculation never calls out mid-way, or replay
   would get another answer. A data-table read is the locked analysis copy
   or a keyed read as of the event's place in the log.
4. *Definitions pass the gate; instances do not.* A new or changed
   definition is drafted (by the interview or a person) and approved as one
   package; starting a workflow, filling it in and moving it are business,
   recorded as events. Model output never runs with authority (invariant 3).
5. *Running instances keep their version.* Today the executor runs only the
   newest active version of a rule, and `effective_from` is only a lower
   bound (`internal/exec/exec.go`, `matchRule`), so a new version stops the
   old one even for earlier-dated events (`packs/us/REVIEW.md`, cause 3). A
   workflow started under definition v1 must be able to finish under v1;
   moving it to v2 is a deliberate, approved act. This is the session's
   first design question, and the US register's cause 3 is the same fix.

**Proofs, in order.**

1. *Prepaid costs as one smart document.* Rebuild v30's two types, one
   action, five rules and view as one definition; on Helios the books are
   identical to the grosz and replay is identical. Proves the unit loses
   nothing.
2. *KK's accrual*: values read from the ledger, the accrual recalculated,
   posted on approval. A candidate is the Polish pair of v30, accrued costs
   (RMK bierne): the year-end audit fee or unused holidays. KK to name the
   one his clients live by.
3. *A document flow with permissions*: a purchase request; above an amount,
   approved by someone other than the requester; then the order and the
   receipt. Segregation of duties is the test: the requester's own approval
   is refused and the refusal is in the log.
4. *The interview drafts a smart document as a unit* from a plain-language
   description, the gate showing its state diagram and worksheet.

**Order of work** (each a commit or two): the definition and its
translation, with no new behaviour (proof 1); variables, parts and
snapshots; permissions per state with roles as data; acting on other
instances under entitlement; proofs 2 and 3; the author's grammar (proof
4); the view of a smart document on the face — its states, its worksheet,
"why?" on every variable. Descriptors, bookings with descriptors,
specialisation and account roles each earn their own step after proof 3,
in the order a big client needs them; APIs as two facts last.

**Prior art to read first.** In the repo: SPEC's Document; `LifecycleDef`
and `Amendable` (`internal/core/core.go`); `Activity` and `Who`
(`internal/core/activity.go`); the formula VM and `Calc`
(`internal/core/vm.go`); `each` over keyed collections and `balance()`
(commits 4c85b9b, 24b6763); the cascade cap and `MaxCollection`
(`internal/exec/exec.go`); bundles (`internal/core/bundle.go`); view
notions (`internal/core/views.go`); the people pack (`person.role`,
deputies); prepaid costs and the CIT advance in `packs/pl/pack.json`; the
arithmetic and matching briefs above (folds, never loops; a kernel method
parameterized by data); `packs/us/REVIEW.md` causes 1, 3, 4 and 6. Outside,
for vocabulary and warrants, not architecture: BPMN 2.0 (OMG) for states,
flows and tasks; how the big systems name approval by amount (SAP release
strategies, Dynamics workflows); KK's Swan, which carries formulas as
element data.

**Open for KK**, each with the standard Rhea would propose:

- *Roles and groups.* Proposed: a role is data (accountant, approver,
  warehouse clerk); a person holds roles with dates; a group is a named set
  of persons or roles; the definition grants per state. `person.role` today
  is a single string.
- *Permissions before the login.* Proposed: declare and check them against
  the acting person from the first step, refusals recorded and tested,
  saying honestly that identity is attributed until the login lands
  ("Packs built together": the login with audit comes first there too).
- *"Data tables".* Proposed: a collection variable is the document's own
  table; reading the company's data goes through the locked analysis copy
  or a keyed read. KK to say if he meant more (a document that maintains a
  table others read, as a price list does).
- *The accrual for proof 2* (above).

## People and rights: who may do what, and the doors people come in through (KK, 2026-10-19)

KK, the evening of the status check against the eleven lines: the to-do
list must carry the notions of user, role, group, supervisor, substitution
and permissions. Users will want to *see* who can do what in the system;
accountants will not accept Rhea without it, and the road into accounting
offices (biura rachunkowe) needs these objects and roles. Then onboarding on
every level: Rhea onboards an office or a client company; an office onboards
its own employees, its clients, and its clients' employees; a company
onboards its employees, and its suppliers, customers and partners. Rights
must be enforced, not only shown; some super admin handles exceptions; by
invitation or by self-nomination; a modern login, with a key; a person who
works for several entities; and in the Poland pack, the company's details
pulled from GUS by its NIP. "This is the sign-in step. Without it we cannot
progress nor proceed."

**What it rests on, already decided.** Permissions belong on activities
("who may trigger this verb"), declared, versioned and gated — never a
second permission system ("Scale and security are welded", rule 2); the
company key is the isolation key. `Activity.Who` exists and is displayed,
never enforced. The people pack (v2) has the person with a handle, a single
`role` string, a status lifecycle (active ⇄ away → left) and a deputy whose
routing a rule already does. "Packs built together" put a full login with
an audit trail before any workspace; "Smart documents" proposed roles as
data, a person holding roles with dates, groups, and rights declared and
checked from the first step with identity attributed until the login lands.
This note is that step.

**How the market does it, so the words are the accountant's.** KSeF 2.0 has
exactly KK's ladder: the taxpayer (the owner, authenticated as the company)
holds all rights; it grants rights to people and to an authorised entity —
the accounting office — with or without the right to pass them on; the
office then grants its employees rights in the client's context, indirectly.
The rights are few and named: manage rights, issue invoices, read invoices.
Beyond KSeF, the office holds the client's tax mandate (UPL-1) and a service
agreement; the office's own books are a company like any other. The company
register: GUS's REGON base (the BIR API, with a key GUS issues free on
request) gives name, address, REGON, PKD and legal form by NIP; the
Ministry's VAT white list (no key, daily limits) gives VAT status and bank
accounts. An accountant adds a client by typing a NIP and confirming what
comes back; nobody retypes an address.

**The notions, in Rhea's terms — all data, all in the log.**

- *Person*: one per human, as the people pack has it, with as many
  memberships as they have workplaces. The handle signs every act.
- *Entity*: a company — an office is a company too (it has its own books).
- *Membership*: a person in an entity with a role, from a date, to a date.
  A person who works for several entities is one person with several
  memberships; the desk shows them side by side.
- *Mandate*: an office keeps an entity's books, from a date, with a scope of
  rights the client granted (KSeF's authorised entity). An office employee's
  rights to a client come through the mandate and the employee's role in
  the office — never by a direct membership in the client, unless the
  client grants one.
- *Role*: a named set of rights, data in a pack (owner, accountant, chief
  accountant, invoicing clerk, reader), a company able to declare its own.
- *Right*: to a door (a verb), to a view or report, to a type's objects or
  a field, to a smart document's state (edit, see, move) — the same
  vocabulary the smart documents brief asks for, within an entity's scope.
- *Group*: a named set of persons in an entity, for rights granted at once
  and for a supervisor's reach.
- *Supervisor*: a right, not a notion of its own — "sees the work of" and
  "approves for" a group; amounts above a threshold route to it.
- *Substitution*: the deputy the pack already has, with dates and a scope
  (which entities), the work routed while it lasts, nothing after.
- *The house*: Rhea's team, an entity like the others, with rights over
  packs, the network and installations — and none over any entity's books
  except as granted below.

**Enforcement — the face says who, the kernel says whether.** The face's
server authenticates the person and is the only door from outside; the
kernel takes the person on every call and decides: memberships in the scope
entity — direct, through a mandate, or as a deputy within dates — give
roles, roles give rights, and a door whose right the person lacks is refused
with the reason, the refusal a recorded fact. Reads are filtered the same
way: objects, views, kept reports and the locked analysis copy answer only
within the person's entities. Rights apply to the person as the pack's
`Who` applies to actor kinds today, so no second system appears: `Who` grows
from "human, agent" into the roles. When faces other than ours exist on the
open API, the kernel verifies a signed token rather than trusting a header;
on white01, bound to the machine, the header from the face is enough.

**Seeing who can do what.** A view, rendered generically like every other:
rows the persons and roles of an entity, columns the doors and views, each
cell a right that opens — the number-opens move applied to rights — to the
membership, mandate, role or substitution that grants it and the person who
granted it, on which day. The same question asked on an object ("who may
move this?") answers from the same facts. This is what an accountant will
show an auditor, and it is the difference from a settings page: every right
has a reason, like every entry.

**Sign-in.** Passkeys (WebAuthn): a key on the person's phone or computer,
nothing to remember, nothing to phish. Enrolment is an invitation: a single-
use link by e-mail that expires, the person registers their first key and
their account is the membership the invitation carried. Recovery is a second
key, or a re-invitation by the entity's owner, recorded. Every sign-in,
failed sign-in and sign-out is a fact under the person. No passwords. The
face's server keeps the sessions; the kernel never sees a credential.

**Onboarding, level by level — who, where, when, what for, and what they
see when it fails.**

1. *Rhea onboards an office or a company.* The person who will own it, on
   the sign-in page's "Załóż konto": NIP, the GUS channel fills name,
   address, form and REGON, the white list adds VAT status and accounts,
   the person confirms — the `identify_company` and `declare_company_form`
   doors pre-filled — and registers the first key as the owner. In the
   private beta the page opens from KK's invitation link; in public, by
   self-registration with the e-mail verified. Fails: a NIP GUS does not
   know (type the details, flagged unverified); an entity already in Rhea
   (the way to its owner or office, never a second copy). Consent for what
   may reach the house is asked here (KK, 2026-10-17).
2. *An office onboards its employees.* The office's owner or a manager, on
   the office's People screen, "Zaproś": e-mail, role, groups; the invitee
   enrols a key. Fails: an expired or used invitation says so.
3. *An office onboards its clients.* An office person with the right, on
   Clients, "Dodaj klienta": NIP → GUS → confirm; the client entity is
   created with a mandate from a date and its scope. The client's owner is
   invited to claim their entity; if they never sign in, the office acts
   under the mandate and the owner's seat stays empty, visible. The KSeF
   authorisation the client granted the office is recorded when it is.
4. *An office onboards a client's employees.* Under the mandate, on the
   client's People screen; typically the client's owner sees everything and
   approves, an employee issues invoices. The client's owner can do the
   same, and revoke.
5. *A company onboards its employees.* As 2, within the company.
6. *A company onboards suppliers, customers and partners.* They are
   contractors — objects the pack already has — not users. A login for one
   of them is a guest membership limited to their own documents, and it
   arrives with the smart document that needs it (a supplier uploading its
   invoices, a customer seeing their account), not before.

**The house and exceptions.** No standing power over any entity's books.
An entity's owner grants the house timed support access through a door —
hours and a reason — and every act under it is signed as the house's person
on the owner's grant, visible on the entity's Today and in its audit, and
ends when the time does. Exceptions at Rhea's level — an owner locked out,
an entity claimed twice, an office that stopped serving a client who cannot
reach their own books — are house doors, each with a reason, each a fact
the entity sees. The owner regains an entity the way it was claimed.

**What the invariants demand.** Granting and revoking are doors, not the
gate: a human act, enforced by rights, never an AI draft. Every grant,
revocation, mandate, substitution and sign-in is an event; replay rebuilds
who-may-do-what identically, and the view is a projection of it. The kernel
never writes a right on its own and never trusts a claim about the person
beyond the face that authenticated them. The GUS and white-list readers are
market code under `markets/pl`, channels registered like KSeF; the kernel
knows no register by name (invariant 8). Secrets — keys, sessions — live in
the face's server, never in the log.

**Order of work.** (1) The notions in the common people pack (v3): role,
membership, group, mandate, deputy with dates and scope, and the rights
vocabulary (doors, views, types, states); every pack's doors keep `Who`,
now in roles. (2) Enforcement in the kernel at every door and every read,
refusals recorded, tests that a person without the right is refused and the
refusal replays. (3) The "who can do what" view, each right opening to its
grant. (4) Passkey sign-in and invitations in the face's server, sign-ins
recorded. (5) The onboarding screens: claim an entity (NIP → the GUS and
white-list channels in `markets/pl`), invite people, add a client with its
mandate. (6) The house: support access, the exception doors. (7) Novi
onboards itself as the first real entity, then one accounting office in a
pilot — the usability test of the whole ladder.

**Open for KK**, each with the standard Rhea proposes:

- *Invitation or self-nomination.* Proposed: self-nomination only at the
  top (an office or a company registering itself; invitation-only in the
  beta), everything below by invitation from someone who holds the right.
- *Claiming an entity.* Proposed: in the beta, KK's invitation; in public,
  the e-mail verified and the NIP checked against GUS, and when it matters
  (a dispute, a second claim) a KSeF-authenticated step, the way the
  Ministry proves who acts for a company.
- *Passkeys only.* Proposed: no passwords at all; an e-mail link only to
  enrol and to recover, both single-use and recorded.
- *Supervisor.* Proposed: two rights on a group — sees the work of, and
  approves for — rather than a hierarchy object; thresholds on amounts
  route to whoever holds "approves for".
- *Guests.* Proposed: a contractor gets a login only when a smart document
  needs one, with rights to its own documents alone.
- *The super admin.* Proposed: the house with no standing access, timed
  support access granted by the owner, exceptions as doors with reasons,
  all visible to the entity.
- *What an office employee sees of a client.* Proposed: what the mandate's
  scope allows and their office role grants, nothing by default.
- *One person across installations.* Later; today every entity lives on
  one installation and a person's memberships are there.

**What shipped (2026-10-20, the kernel's half).** Steps (1), (2) and (3) of the
order of work: the notions in the common people pack v3 (role, membership, mandate,
substitution, group; the pack's roles as master data; the vocabulary of rights with
`door:*` stopping at the people domain as KSeF's "manage rights" does); enforcement at
every door and every read (`internal/rights` computes a person's grants from the log on
the books' day; the shell makes them the request's scope; the executor checks the door's
Who and the person's grant in the company acted for, signs every `…_by` input with the
person, and records a refusal as `access.refused`; a company nobody holds is claimed by
its first grant); and the who-can-do-what view as the `rights` notion, every cell the
grant that opens it. `rhea serve -require-person` makes the kernel serve signed-in
persons only; `rhea people` is the operator's hand at the first owner. The decisions are
in DECISIONS under the day. Next: (4) passkeys and invitations in the face, every
sign-in reported through `record_access`; (5) the onboarding screens with the GUS and
white-list channels in `markets/pl`; (6) the house; (7) Novi, then an office.

## Isolation in the database: a scope is a set of companies (KK, 2026-10-20)

KK, after the two reviews: "We are developing a big ERP and we cannot afford such
things like summing all companies or giving one company data of another company."
And: "an Accounting Office can see many companies — its clients." Invariant 9.

**Who sees what.** A *scope* is the set of companies a request acts for. A company's
scope is itself. An accounting office's scope is its clients, every one of them, and
when it writes a fact it writes it for one named client (the *acting* company, one of
the set). Rows with no company — the pack's master data (chart of accounts, VAT
rates), the kernel's own verbs (a rule approved, a bundle rejected) — are the
installation's and are seen by every scope. Nothing else crosses.

**Where it is enforced.** In the database, the way invariant 1 is: a role every
scoped request runs as (`rhea_app`, no bypass) with row-level security on `event` and
`object` — a row is visible, and may be written, when its company is in the request's
scope or it is the installation's. The pool switches role and sets the scope on every
connection it hands to a scoped request, and resets both when the connection returns.
A query that forgets its scope returns nothing. The kernel itself — the executor's
replay, backup and restore, the pack loader, init — runs unscoped as the owner and
sees the whole log; a request from a person never does.

**The analysis copy.** The DuckDB file is built per scope from what the scope sees,
under the same role, so an office's analysis holds its clients and nothing else and a
pack's SQL needs no filter of its own: `SELECT SUM(amount) FROM finance.posting` is
the scope's sum. Every projected table also carries `company`, for a pack that wants
to show its clients side by side. The copy's key is the scope's own mark.

**Until sign-in.** The scope comes from the request today (`X-Rhea-Company`, one key
or a comma list; the face sets its company's), and a request with none is unscoped —
the owner's whole log, as before. The people-and-rights session (the sign-in step)
makes every request carry a principal whose grants are its scope, and an unscoped
request then sees nothing. That is the day invariant 9 is whole; the database side is
ready for it now.

**What is owed.** Facts from before the partition key carry no company and read as
the installation's: a single-company database (rhea_pl, rhea_face) restores with
`rhea restore -company <key>` stamping what lacks one — the log is append-only, so a
stamp is a restore, never an UPDATE. The blob store is keyed by hash and carries no
company; a hash is unguessable and reaches a reader only through that company's
events, and a company column there follows when files are listed. The office's
sign-in, its grants per client and the client's view of who may see its books are the
sign-in session's.

## The face, whole: one box, one conversation, many tabs (KK, 2026-10-21)

KK: *"it is finally time to approach the real Rhea UI."* Rhea is AI first —
using her should feel like Claude in VS Code or Claude Desktop — and at the
same time a business system that opens data-heavy grids, documents that look
like documents, 360s of one customer, flows and timelines. Artifacts open in
tabs like a browser's or VS Code's, and also sit in the conversation as a quick
proof. The face follows intent and learns the person's preferences the way
Winston learns a guest's habits. It must be elegant, recognizable, modern but
not too much, and above all consistent — *"every time we add a new object or
presentation, you have to check if such components, like pickers, toggles,
buttons, already exist"* — and fast: any change suspected to touch speed is
measured and optimised. Conversations are kept, named so they are recognized,
and grouped. Tabs rearrange by dragging, as in VS Code, from the very
beginning. The face must hold Rhea Apps built on smart-document flows (KK
designed all of Qalcwise's flows and will bring them) and, later, the
enterprise domains and the custom apps of clients' own business models
(`targetRhea.md` lists fifteen). Alpha's shell is the start; KK: *"I love the
Omnisearch… it is faster than AI. No thinking, just millisecond access."*

**The formula: one box, one conversation, many tabs, and every artifact in
three sizes.** *One box*: the omnisearch is the front door to everything —
objects, views, actions, past conversations — answered in milliseconds from
an index on the page, no model involved; only Tab, or no match, sends the same
words to her. The fast path never waits for the model: she is the second row
in the box, never a gate in front of it. *One conversation*: she is always
there — a panel on the right at the desk, the home screen on the phone. *Many
tabs*: whatever is opened lives in a tab with its own address; tabs survive a
reload, travel as a link, and the tab in front is what she looks at when asked
"why is this so high?". *Three sizes*: one artifact, one set of data, one
component — inline in the conversation (a card, five rows, one figure), as a
tab (the whole working surface), pinned on Today (a live figure every
morning). "Open in tab" grows it; dragging a tab into the conversation hands it
to her. Nothing is drawn twice, so nothing can disagree. Two arrangements one
key apart: the conversation in the middle with the artifact beside it (Claude
Desktop), or the tabs in the middle with the conversation on the right (VS
Code); the owner on the phone lives in the first, an accountant with forty
clients in the second. Her answers are drawn, not written — the rule the face
already keeps, now everywhere: *she chooses what to show, the books compute
what is shown*; a figure is an artifact from the books with "why?" under it,
and her words frame it and carry none of their own.

**Tab groups.** VS Code's editor groups and floating windows, in the face from
the first build. The arrangement is a tree — splits, each holding a group,
each group holding tabs, each tab an address — never a flat row with splits
bolted on later. Drag to reorder, to an edge to split, between groups, into
the conversation (context for her), out of the window (a second monitor, kept
in step with the first); every gesture has a key and a menu entry, because
dragging is slow for experts. The arrangement is the person's, remembered per
device; named arrangements follow ("month end": trial balance, VAT register
and bank side by side), she notices which one a job uses and offers it, and
"put the bank statement next to the open invoices" is an arrangement she
proposes and the books fill. On the phone the groups fold into one stack over
the same tree. Hidden tabs are not drawn: thirty open must feel like one.
Written in the face's own plain JavaScript, so the component gallery and the
speed budgets cover it.

**Shapes for every domain.** Fifteen domains and clients' own apps cannot be
fifteen sets of hand-made screens — that is how large systems become
incoherent. The six shapes (card, paper, ledger, statement, buckets, chain)
grow by what enterprise work needs, into a closed set: *grid* (the data-heavy
table, keyboard-complete, pasting to and from a spreadsheet — where an
accountant lives), *360* (one party with everything around it), *timeline*,
*board* (a smart document's instances by state), *flow* (its states, where this
instance stands, who moves it next), *calendar*, *chart and dashboard*. A Rhea
App declares which shapes show its documents; a client's custom app is a new
composition, never a new screen; a new shape is rare and reviewed like a new
kernel primitive. KK's Qalcwise flows decide what board, flow and the paper
with its state bar must hold.

**Learning without moving the furniture.** Interfaces that rearrange
themselves destroy the muscle memory that makes an expert fast; predictability
beats cleverness. Winston remembers the coffee; he does not move the lobby. The
frame never moves by itself. What is learned lands in places made for it: the
box's ranking (what this person opens, and when), Today's shortcuts and
figures, defaults in forms, a "usually next" after a step. Said preferences
are kept where the person sees them ("what I remember about you") and undone
there; a habit becomes a change only after she asks once ("you open the ageing
every Monday — shall I put it on Monday's Today?").

**Conversations.** All kept, as the person's facts. She names each one and
renames it when it ends, so the title is the outcome, not the question ("why
is Nordwind's balance off?" becomes "Nordwind: double payment, refund
asked"). No folders she invents — they drift and lose trust; the company
already files itself by its parties, its months and its processes, so a
conversation about an invoice lives on that invoice, in the supplier's 360 and
in its month, beside the person's own pins and folders. The box finds them all.

**Type and the brand, fixed now.** The brand was never fixed and two looks
exist: the kernel's shell and rhea.plus (teal, the glowing kernel mark, the
wordmark in IBM Plex Serif) and the face (ledger-paper green-grey, green-black
ink, brass for her, a ring-and-disc icon). They become one now, before the
desk is built. The serif of her voice reads as Claude's and goes. The type is
KK's Novi.fm idea: one face for words, a monospace for figures — tables are
mostly columns of numbers, and a monospaced column aligns by itself. It also
says something: *monospace is a fact from the books* — every amount, date,
document number, account and tax id in mono, inside her sentences too, and
every one of them opens "why?". The narrative face carries the interface and
her voice; the mono is chosen as its partner (x-height and weight that sit
together in one line), digits first (a 1 that is not a 7, a 0 that is not an
O, good at 13 px), Polish drawn properly (ogonki and ł), German, the rest of
Latin Europe later. Chosen by looking at Rhea's own screens, not specimen
sheets; bought from the foundry with a licence for the web app and for the
wordmark, not from a subscription marketplace; one variable file per family,
served by the face itself (fonts from Google's servers are slower and, after a
Munich court in 2022, a consent problem in Germany). Then the colour, then the
mark and the wordmark in the chosen face; the kernel's shell and rhea.plus
follow.

**The type, chosen the same day.** KK, on the comparison page (Rhea's own
screens with Helios's books): *Schibsted Grotesk* for words and *Martian Mono*
in its narrow cut (width 87.5) for figures, the figures at 88 % of the word
size — exactly Schibsted's x-height over Martian's, so one line keeps one
x-height; Martian's zero carries its mark by default. Both are under the Open
Font License, so nothing is bought and the open kernel can ship the same files:
the open part and the face keep one look. The face took them the same day
(rheaface 7590dcf), served from its own files and from no third party; the
kernel's shell and rhea.plus follow with the colour and the mark.

**The colour, chosen the same day.** KK: take Alpha's schemes. Alpha's model
carries over whole: two settings, light or dark and a *palette*, and a palette
is only the paper and the greys; the accent is defined once and every palette
inherits it; one palette, Classic, moves the accent on purpose (VS Code's blue)
and says so; every colour step is measured against the paper it sits on. Six
palettes: **Ink**, the default (blue-black ink on cool blue-grey paper, deep
navy in the dark — KK: blue with gold rather than today's green), Ledger
(today's green), Paper (Alpha's warm Tylko neutrals), Neutral, Steel (NoviUI's
graphite) and Classic, whose dark greys carry a hint of cool because pure grey
read brown beside the blue. Her gold works by role: **champagne** fills buttons
and her mark, with navy text and no border (KK: the deeper gold with dark text
looked muddy, and the edge was not wanted — the label carries the button); a
deeper gold draws thin lines, 3:1 on every paper; the deepest carries small
text, 4.5:1, on its own wash too — which today's brass failed. One rule
stricter than Alpha: the faintest captions reach 3.5:1. Every palette, both
themes, passes. The palette is the person's setting; until sign-in carries it,
the browser keeps it, as in Alpha.

**The mark, chosen the same day: Orbit.** Rhea is a moon of Saturn: her gold
moon on its orbit round the kernel, something in motion, the way she works
while the owner is away. Of four directions (Orbit, Total — a coin over the
double rule under a final total, R. — the initial with a gold full stop, and
the face's old ring, which read as a radio button) KK took Orbit. The brand's
colours do not move with the palette (Alpha's rule): a navy tile, paper lines,
a champagne moon; bare on light paper the moon takes a deeper gold, or
champagne would vanish. The ring has a real gap where the moon passes, so the
mark needs no halo and sits on any ground. The word beside it is "Rhea" in
Schibsted Grotesk SemiBold, "AI Resource Planning" under it in Martian Mono;
the logo files carry both as outlines. The brand went into all three places at
once: the face (palettes after Alpha's model with Ink the default, chosen by
the person in "Co o Tobie pamiętam"; the mark in the tab bar and as the app's
icon), the kernel's own shell (tokens.css, its fonts, the logo files) and
rhea.plus. The sources — palettes, the logo writer, the three comparison pages
— live in rheaface `tools/brand`, and the face's colour tokens are written from
them, never by hand.

**The desk, first build (2026-10-21; rheaface 815a3f0).** Served at
`/desk.html` beside the phone, built from the face's own code: `openFrom`
hands every opening to the desk, so each sheet the phone has opens as a tab
(one artifact, three sizes, in practice). The box answers from an index on the
page and times itself (on Helios 41 of 191 in 0,1 ms); tabs live in a tree of
groups, dragged, split, sent to her, popped out; grids come from the pack's
views (sorted, filtered, only visible rows drawn, totals, rows copied for a
spreadsheet); a firm's 360 gathers what refers to it; the first flow shows
sales invoices by where each stands (KSeF's answer, the open amount) and who
moves it next. Read on the way: a pack view titled in English ("Trial
balance") shows so on the desk — the pack's words, not the face's; and the
flow's "after 30 days" counts from issue, as the books have no payment term.
Still owed by this section's order: the component gallery and the checks,
before the next screen; KK's Qalcwise flows into the flow shape.

**The table's layouts (2026-10-22; rheaface 7c3d8b9, 8e408d5).** Three of the
closed set's shapes — board, calendar, timeline — are built as layouts of the
one table, not as screens of their own, with cards beside them: the same rows,
filters and order, chosen in the table's toolbar and saved with the view (KK:
Airtable's habits). A layout is offered from the column types alone — a board
needs a state or a reference, a calendar a date, a timeline two — and one not
offered stays in its place, greyed, saying what it needs, so the frame does
not move from one list to the next. A state's board keeps the pack's order of
values and an unused step as an empty column, which is what a smart
document's board needs; the first flow is now this board under its strip of
steps, the shape KK's Qalcwise flows will fill. The first of the checks
exists: `node tools/gridcheck.mjs` drives headless Chrome through its
debugging port (real frames and a real clock — `--dump-dom` runs on virtual
time, where a budget measures nothing), clicks, keys, saves and reopens every
layout, draws every list the desk offers in each, and holds 100 000 rows to
under 4 000 drawn elements in two frames per layout. Still owed: the gallery
page itself, the token check, a chart as the table's next layout.

**The gallery and the checks (2026-10-22; rheaface 675d9d6, 35bbf9e).** The
three consistency items stand. (1) The gallery is a tab of the desk
(desk.html?gallery; rhea-gallery.html with the Helios books), drawn by the
desk's own parts and the face's renderers, never by copies: 38 exhibits in
seven sections, each live in light and dark side by side, 59 states a button
away, the phone as two real phones put into each state. (2) `node
tools/check.mjs` fails on a colour or font outside the tokens, on a hand edit
of the generated token block, on a var(--…) that names nothing, on a font size
off the scale, and on any stylesheet class no exhibit shows in any state (382
of 384; the two drawn only on live books set aside with that reason). (3) The
face's CLAUDE.md: the gallery first, tokens only. The walk through every state
found three real faults, each fixed with its check: a phone sheet opened just
after closing vanished, a table tool needed two clicks after a column's menu, a
disabled button looked enabled; and three rules no code used. Open, KK's
choice: the face uses twenty-seven font sizes where a designed scale has about
nine; the check holds the twenty-seven, so nothing new creeps in, until the
scale is chosen and becomes tokens.

**Consistency and speed are checks, not promises.** Several sessions build the
face, and none remembers the last; a promise to reuse does not survive that.
(1) One living gallery page shows every component in every state, light and
dark, phone and desk; a new component lands there first. (2) The face's check
fails on a colour, size or font outside the tokens, and on a component not in
the gallery. (3) The face's CLAUDE.md: before any new screen, look in the
gallery; reuse or extend, never fork. (4) Speed budgets are tests: the box
answers within one frame (16 ms) on a large company's index; a first screen
in under a second on a middling phone; a 100 000-row grid scrolls smoothly
because only visible rows are drawn; a cap on what the page downloads, set
from the first measurement. A broken budget is red like a broken test. Plain
JavaScript and no framework stay — the biggest speed asset there is.

**Order of work** (KK: "we are fine to start"). (1) The brand: the type
comparison on Rhea's own screens, then colour, then mark and wordmark. (2) The
desk, first, as the face's second density — not a second product; the kernel's
shell stays a developer's tool. A clickable desk on Helios: the box, tab
groups, the conversation with an inline card that grows into a tab, a grid, a
360, one flow. (3) The gallery and the checks before the second screen. (4)
KK's Qalcwise flows into board, flow and the paper with its state bar.

One line: *one box that answers before she thinks, one conversation, tabs in
groups as VS Code has them, every artifact inline, in a tab or pinned; she
chooses, the books compute; a closed set of shapes for every domain and every
client's app; the frame never moves, the learning lands in its places;
monospace is a fact from the books; consistency and speed are tests.*

## The desk in order: what is where, and whose (KK, 2026-10-22)

KK, after the gallery: "Everything here is very chaotic, functions are mixed between
drawer cards, not sorted, we mix metadata with transactions, general mess." And: if an
accounting office works here, where does it choose the client? Where are users, their
responsibilities, rights and permissions managed? Are there user groups? How does the
house invite or create new clients? (The quick questions, the first of his points, moved
the same day into her panel, about the tab in front — rheaface 7e71db9.)

**What it rests on, already decided.** The notions are the people pack's (v3): person,
role, membership, mandate, substitution and group, each a fact, with the who-can-do-what
view (the `rights` notion, every cell opening to the grant behind it) — "People and
rights", above. The onboarding ladder there (Rhea onboards an office or a company by NIP
through GUS; an office adds its staff, its clients under a mandate, its clients' staff)
stands; its screens and the GUS and white-list channels are its steps (5) and (6), not
built. Acting for one company sees that company alone (the review's B1, 2026-10-22). The
frame never moves; the face knows no business. What is missing is the face's half: where
each of these lives on the desk. Today the drawer files views by the pack's function
folders, so a domain mixes documents, reports and master data, and titles still in
English ("Trial balance", "FX rates", "Invoices") show as they are; the desk names its
one company as fixed text; people, rights and keys are on the phone only; nothing creates
a company.

**How the market does it.** QuickBooks Online Accountant: a "Go to QuickBooks" dropdown in
the toolbar lists the firm's clients, the client's name then stays in the toolbar with
"Back to practice", and the firm's home is a client list. wFirma's office list: a client
is entered by its name, up to five pinned at the top. Xero: the organisation's name at the
top left switches organisations. Odoo's Accounting menu keeps customers' work (invoices,
credit notes, payments), vendors' work, reporting and configuration apart. One company at a
time is the rule in all of them; the switch is always in the same place, at the top.

**Proposed, each the standard Rhea suggests.**

1. *The drawer by kind, then by function.* Every domain's drawer has the same three
   folders in the same order: **Dokumenty i zapisy** (lists of documents and entries — the
   work: invoices, bank lines, postings, flows), **Raporty** (registers, statements,
   analyses — computed, read-only), **Kartoteki** (master data: firms, the chart of
   accounts, VAT rates, exchange rates, people's roles). The kind comes from the data
   alone — a list of a type the pack marks as a document, a list of any other type, a
   computed view — so the face stays ignorant of business; the pack's function (VAT,
   KSeF, Rozrachunki) groups within a kind only where a kind holds many. Every title
   comes from the pack's words in its language: the Poland pack gives Polish titles to
   the views still in English, and the pack's check fails on a title not in the pack's
   language.
2. *The company switch at the top, one company at a time.* The company's name in the
   top bar becomes the switch: it names the company acted for and opens the companies
   the person may act for — search, pinned and recent first. Choosing one turns the whole
   desk to it — the drawer, the tabs, her panel, the box; each company keeps its own
   arrangement of tabs, so returning restores it. One company at a time, as the market
   does (KK chose it): several clients side by side is where a booking lands on the
   wrong books. *Not in this step* (KK, 2026-10-22): the office's worklist over all its
   clients — what is done, what remains, notifications about a client — and which of its
   accountants answers for which clients. They are designed together with case
   management and the smart documents' workflows, not built ahead of them; the switch
   names companies, nothing waits in it yet.
3. *People and rights in a "Ludzie" activity.* For the company acted for: **Osoby** (who,
   in which role, from when, invited by whom), **Uprawnienia** (the who-can-do-what view,
   every cell its grant), **Grupy**, **Zastępstwa**, **Zaproszenia** waiting, **Moje
   klucze** — the phone's screens drawn as tabs. In an office, the same for its staff,
   and which staff serve which client.
4. *Groups: there, and used for teams.* The pack has them; the face shows them. An office
   gives a client to a team (a group) rather than to persons one by one; a supervisor's
   two rights ("sees the work of", "approves for") rest on a group, as "People and
   rights" proposed.
5. *New companies, built as the ladder says.* The sign-in page's **Załóż konto** (in the
   beta, from KK's invitation link): NIP, GUS fills the rest, the owner confirms and
   enrols a key. An office's **Klienci → Dodaj klienta**: NIP → GUS → confirm → the
   mandate from a date → the client's owner invited to claim their seat. The house keeps
   no standing power; its exception doors, each with a reason, are the house company's
   own desk. This needs the GUS and white-list channels in `markets/pl` first.

**Order of work.** (1) The drawer by kind, with the Poland pack's Polish titles — the
fastest relief (done the same day: Rhea 68182f2, pack v42, rheaface e4cc52b). (2) The
company switch, on a test world of an office with two clients (done the same day: Rhea
d1614b9 names the companies in whoami, rheaface 5c2a471 the switch; checked on a stubbed
person of three companies and the kernel's two-company test; walked live the next day,
where a person of several companies coming to the desk was sent back to sign in — fixed,
rheaface a22627c). (3) Ludzie, from the phone's screens and the rights view (done
2026-10-23: rheaface 88874a6 the waiting links listed and withdrawn, 1b0f671 Ludzie's
Dostęp — Kto ma dostęp, Kto co może, Zaproszenia — and the person's own menu at the top
with their keys and signing out; walked live on an office with two clients by
`tools/walk.mjs`, the face's live check: the kernel at a named commit on a fresh database,
sign-in with a key Chrome simulates). Steps (2) and (3) left one question for KK: a
withdrawn link leaves the membership it carried, so the person keeps access they never
used until someone ends it; the market's standard (a revoked invitation takes the pending
user with it) is the one to propose. Groups (4) show as Kartoteki and in Kto ma dostęp; a
client given to a team waits for the worklist's design. (4)
The GUS and white-list channels, then Załóż konto and Dodaj klienta (done 2026-10-23: Rhea
46a682b the registers as a read the kernel serves — GUS's BIR 1.1, the Ministry's VAT list —
and 42e7c99 the Poland pack's enter_company door, v44; rheaface 00158e4 the form, Dodaj
klienta at the switch's foot, Załóż konto from Rhea's staff's link; walked live with the
registers stood in by their recorded answers; GUS on white01 waits for the key GUS issues on
request). Each in the gallery
first, checked as the face's check asks. Apart, to design first with case management and
the smart documents: the office's worklist over its clients and who answers for whom.

## What never changes, what the law imposes, what the business chooses (KK, 2026-10-23)

KK, after a proposal that would have written one accounting office's business model into the
face — a fixed client list, three pricing models, the fields of an agreement: *"will every
office company have to behave the way we defined their business model?"* Rhea is an AI-ERP;
the principles that keep it one, in KK's words and the decisions already made:

- **The face knows no business.** It renders what the definitions say; a screen that needs to
  know what an invoice is belongs in a view.
- **Implementation by interview** (2026-10-07): not a configuration — the company's own evidence
  first, the unexplained residue drives the questions, done when nothing is left unexplained.
- **A Rhea app is a package of smart documents** (2026-10-19): the Accounting Office, eCommerce,
  SaaS, Professional Services apps — smart documents with their workflow and business objects
  supporting one business model, Rhea's standard as the starting point, the interview shaping it
  to each company, every company ending with its own version, approved and explained.
- **Some things do not change:** the ledger, the account, the journal, the balance, the
  accounting period — the kernel and its base vocabulary.
- **Some things the law imposes:** the market's pack — VAT, KSeF, JPK, CIT, the statutory chart.
- **The rest is the business model, and a business model is processes and data:** apps, held as
  data, adapted by the interview.

What this asks of the work: the platform — who is who, who may do what, which company, how a
company enters Rhea, drawing any view and any smart document, reading across an office's clients
within its mandates — is the face's and the kernel's; anything that says how a kind of business
works is an app. The face holds business knowledge today that belongs elsewhere (the sales
invoice's flow states, Today's VAT month-end and bank-line cards, the starter questions, the
stand-in domain words): each moves out — law to the pack, business to an app — with the face's
check and the live walk guarding what the person sees.

## Rhea Apps and business packs (KK, 2026-10-23)

**The names.** An **App** (a Rhea Smart App) is a small program run in Rhea. A **business pack**
is a collection of Apps for one business model: *Rhea for Accounting Offices Business Pack*, and
packs for eCommerce, a SaaS company, a professional services company. "Smart document" goes: in
95% of cases there is no paper. *Document* means the paper again — an App may digitize one coming
in (a reader: KSeF, a bank statement, a scan) or generate one going out (a writer: an agreement,
an invoice, a goods receipt). The layers are the ones above: what never changes is the kernel and
its base vocabulary (ledger, account, journal, balance, accounting period); what the law imposes
is the market pack; the business model — processes and data — is business packs of Apps, shaped
for each company by the interview.

**What an App is** (KK's words, then the parts agreed in talk the same day).

- An App is a program that manipulates data, and it knows how to calculate.
- It may run through a defined flow of states, but does not have to: most Apps — nine in ten —
  are stateless.
- It may have a visualization; the simplest is a table, which the face already has. Every tab of
  the face is an App.
- It respects the security of every object it touches, and it says who may run it.
- Rhea designs an App with the person, interactively — a sandbox, a visual artefact — and the
  definitions stay light and nimble: performance is of utmost importance.

**An App's definition consists of:**

1. *Variables* — scalars or collections, each set by input, by a formula, by a query, or by a
   combination; parameters among them (the month a report is for).
2. *Formulas* — Rhea's formula language: exact money with declared rounding, reading across
   linked records; loops only through collections, so a calculation always finishes.
3. *Visualization* — interactive, and it may depend on the variables: a table (and its layouts:
   cards, board, calendar, timeline), a form, a statement, her words.
4. *Workflow* (optional) — states, transitions between them, the actions performed on a
   transition, and who may do what in each state: change which values, see them, move it on.
5. *Its own data structures* — an App to create a client defines the client.
6. *External data structures* — another App's data, or Rhea core's, read within the person's
   rights.
7. *Connectors* — an API fetch, an e-mail, KSeF.
8. *Triggers* — what sets it going with nobody at the screen: a fact arrives (an invoice from
   KSeF, a bank line, a scan, another App's action), the clock (a day opens, a month starts, a
   deadline comes), another App's state changes (the client's month is closed), a condition on
   data becomes true (an amount over a limit, a payment overdue, stock below its minimum).
9. *Autonomy* for each automatic action — **act** (do it), **propose** (one tap on Dziś), or
   **earn it** (propose first, act on its own after a run of accepted proposals, propose again
   at the first overrule: the learning brief's trust, 2026-10-22).
10. *Who may run it* — a right on the App as a whole.

**Automatic by default.** KK: an App that waited for a button would be the opposite of the goal.
The path every change takes is fixed — *a fact arrives, an approved rule matches it, the rule
records its result, the books follow* — and nobody presses anything: the person approves the
App once, and from then on it runs on every fact that matches. A person in the loop is a choice
the App declares, step by step, never a limit of the engine. The automation is the approved
rules, not the AI: the AI designs; when an App runs, nothing is guessed, and every automatic step
is explained and replays.

**What the invariants ask of an App** (the smart-documents note's, carried over).

- *It never writes the books itself.* It reads, computes and decides; an action records a fact;
  a rule books it — so every value names the fact, the rule and the version behind it.
- *It always finishes:* loops over named, capped collections; no `while`, no recursion; named
  algorithms (FIFO, allocation with the remainder) stay kernel methods an App calls.
- *A connector call is two facts*, the request out and the answer in, through a channel; nothing
  calls out in the middle of a calculation, so replay never asks again.
- *Definitions pass the gate; running an App does not.* A new or changed App is drafted (by the
  interview or a person) and approved as one package; starting, filling and moving it are
  business, recorded as facts. Model output never runs with authority.
- *A running process keeps its version*; moving it to a new one is a deliberate, approved act.
- *It sees only what its person may see*, within the company acted for (or, for an office,
  the clients its mandates let in); chains stay bounded, every trigger safe to fire twice.

**How it is written and shown.** One definition per App, in the format packs already use (JSON:
the AI writes it reliably, the kernel validates it before anyone approves), with a section per
part above; formulas in the formula language; read-only queries in SQL over the analysis copy.
At approval the kernel turns it once into the pieces it runs — types, actions, rules, views —
so nothing is interpreted per click; values are computed when facts happen, not on every look;
only what is on screen is drawn; speed budgets are checks. The face shows a definition as a
page — its variables and where each comes from, the state diagram, who may do what, a live
preview on the company's own data — and the same page is the sandbox while Rhea and the person
design it. The format as agreed (KK, 2026-10-26) is in `APPS.md`, "The App definition";
the sketches below were its first draft.

**Two Apps, sketched.**

*Terminy podatkowe* — stateless, as it is today (the Poland pack's `kalendarz-podatkowy`, v2):
a table under Księgowość in the morning briefing; its source one query over the books (the VAT
and income-tax days, moved past weekends and Polish holidays, Easter computed); no workflow; who
may see it from the people pack's roles. As an App definition:

```json
{ "app": "terminy-podatkowe", "version": 2, "title": "Terminy podatkowe", "folder": "Księgowość/Poranek",
  "run": "view", "variables": { "today": "=today()" },
  "data": { "query": "…the statutory days, moved past weekends and holidays…" },
  "view": { "table": { "columns": ["rodzaj", "okres", "termin", "kwota"], "editable": false } } }
```

*Rozliczenie miesięczne klienta* — the accounting office's monthly settlement, with a workflow
(in the office's business pack, every number of it the interview's to shape): an instance per
client per month; states *liczone → gotowe → zafakturowane → zapłacone*, and *sporne*; variables
— documents booked in the client's books this month (a query within the mandate), employees on
payroll, the rate (a formula over the agreement's price bands), net, VAT, the payment term;
triggers — the month starts (the clock opens one per active agreement), the client's month is
closed (→ gotowe), the client's payment is matched in the office's bank (→ zapłacone); autonomy
— the invoice is issued and sent through KSeF on its own up to the agreed fee, proposed above
it; rights — the accountant sees it and corrects the count with a reason, the office's manager
moves it on, the client sees it once invoiced. A person touches the exceptions alone.

**Rhea today, and what is missing.** Formulas, data structures, external data within rights,
security by company and right, facts-and-rules automation, the clock, connectors as channels,
drafting by Rhea and the approval card's live preview: there. Missing: the App as one unit;
an App's own variables and parameters, and visualizations built from them (forms, inputs);
rights per state, snapshots at each move, a running process keeping its version; triggers and
autonomy declared in one place; generic connectors as configuration rather than code; the
designer page; a right on the App as a whole.

**Order of work.**

1. *Every tab an App:* today's views, actions and rules grouped into App definitions with no
   change in behaviour — the folder, the title, who may run it. The register is `APPS.md`:
   every App Rhea already has, its layer, and what the face knows that is not its own. KK
   turned this step first to what Rhea is with nothing installed and what each pack brings:
   "Core, market packs and business packs" below sets its order.
2. *Variables, parameters and interactive visualization* — a table and a form driven by an
   App's variables.
3. *Workflow in the kernel* — rights per state, snapshots, a running process keeping its
   version (the smart-documents note's first question).
4. *Triggers and autonomy* — declared per App; act, propose, earn it.
5. *Generic connectors* — an API fetch and an e-mail as configured channels.
6. *The designer page* — the sandbox and the visual artefact; Rhea designs an App by interview.
   Its design: "The App designer" below.
7. *The first business pack: Rhea for Accounting Offices* — its Apps from the market's standard,
   with their warrants, shaped for each office by the interview; with it, the office's read
   across its clients within its mandates.

Speed budgets throughout. Still to settle, each with the standard proposed: an App the law
imposes (the FA(3) invoice, the JPK file) ships in the market pack and a company may add to it —
its own fields, its own approval step — but not change its statutory part; values are computed
at four moments — on input, when a fact they read changes, on the clock, or only when looked at.

## Core, market packs and business packs (KK, 2026-10-23)

KK, before the first App is declared: *"we should define what comes with Rhea without any
business pack — what is the core Rhea. A user opens Rhea without any pack installed, what does he
see? Then what comes in the Market Pack."* The answer, agreed the same day.

**Rhea with nothing installed** is the five things that never change and the platform: sign-in,
the person's companies (the switch), Ludzie (people and rights); Księgi — Dziennik, Plan kont
(empty until a pack brings a chart or the company builds its own with Rhea), Zestawienie obrotów
i sald, Okresy (open and close a month), Kontrahenci and Moja firma (name, tax id, address);
**Rachunki i karty** — the company's own bank accounts (bank, IBAN, currency, purpose: main,
VAT, foreign currency, savings), its cards (credit or debit, the account they settle from, the
statement day, the limit), its cash desks, statements in (MT940; ISO 20022 CAMT.053 and card
statements later) and matching, a kernel method; Dokumenty — files received and issued, with the
readers that are no country's (MT940, scanned PDFs); Dziś — what waits for approval, and a card
saying no pack is installed, pointing to Ustawienia → Pakiety; Rhea herself — conversation,
"why?", search — who says she was taught nothing yet and can learn from scratch by the interview.
What the founding vocabulary carries that is law (invoices, VAT rates, exchange rates) moves to
the market pack — amended in "Invoicing per country, and each company's own chart" below: the
invoice's common shape stays core, its law moves to each country's invoicing pack.

**A country is a base, one regime pack, and modules.**

- **Polska — podstawa**: the law every Polish company meets whatever its books — VAT (rates,
  registers, settlement, JPK_V7M/K; switched off by an exempt company's status), KSeF in and out
  and the FA(3) invoice, the tax calendar, the NBP rates, GUS and the VAT white list, the
  company's identity and declared form (urząd skarbowy, legal form; books full, KPiR or ryczałt;
  tax CIT, skala, liniowy or ryczałt), the vehicle mileage log (a VAT matter), the Polish banks'
  file dialects, split payment and the VAT account, the white-list check of a supplier's account,
  the Polish words.
- **One regime pack per company**, by its books: **Pełna księgowość** (UoR: the chart, double
  entry, CIT with Estonian CIT as a choice inside — a tax regime that needs full books — deferred
  tax, provisions, RMK, year-end close, financial statements, micro and small simplified as
  options, equity; most of today's pack v44); **KPiR** (the book, 19 columns from 2026, the stock
  count, the equipment register, PIT advances on skala and liniowy, IP Box records as an option,
  JPK_PKPIR); **Ewidencja przychodów** (ryczałt: the revenue register by rate, ryczałt advances,
  JPK_EWP); later **Organizacje pozarządowe** (uproszczona ewidencja przychodów i kosztów) and
  **Działalność nierejestrowana** (uproszczona ewidencja sprzedaży). Karta podatkowa is not
  offered: closed to new entrants since 2022, its existing users need only VAT and employee
  records.
- **Modules**, each country's own, needed in more than one regime: **Środki trwałe — Polska**
  first (the register of fixed assets and intangibles, KŚT, the statutory rates, OT and LT, the
  one-off write-off up to 10,000 zł, improvements, sale and disposal, tax and balance
  depreciation), later **Kadry i płace** (employees, ZUS, PIT-11), possibly Magazyn and Pojazdy.
- **The rule that keeps them apart: modules compute and keep their registers; the regime pack
  decides how anything reaches the books.** Depreciation the fixed-asset module computes is 400/070
  postings in full books, a cost line in KPiR, and nothing under ryczałt, whose register is still
  kept; a bank line is postings in full books and no book entry at all in KPiR. So one module
  serves every regime.

Which a company installs: sp. z o.o., S.A. and spółka komandytowa — Podstawa and Pełna
księgowość; a sole trader or a civil or general partnership on skala or liniowy below €2m —
Podstawa and KPiR (IP Box records if used); on ryczałt — Podstawa and Ewidencja przychodów; above
€2m or by choice — Pełna księgowość in its PIT variant; an NGO with no business activity and
income up to 100,000 zł — Organizacje pozarządowe; unregistered activity — Podstawa's sales
record; with employees — Kadry i płace; an accounting office — its own regime for its own books
and the business pack *Rhea for Accounting Offices*. Rhea proposes them at onboarding — GUS gives
the legal form, the person confirms the tax form — and the person approves.

**Installed per company.** A pack installs into the company acted for, never the whole
installation: a sole trader on KPiR and a sp. z o.o. on full books share an installation, and an
office's clients mix regimes. Definitions of types and views stay shared; the rules a pack brings
are the company's own (the company rules of 2026-10-22), and what the company sees follows what it
installed. **Ustawienia → Pakiety** has two columns, *Dostępne* and *Zainstalowane*; each pack
says its kind (market base, regime, module, business), version, what it brings (its Apps), who
made it and what it needs (KPiR needs Podstawa); **Zainstaluj** shows what will be added and
arrives as a proposal on Dziś, approving it makes it active; **Uaktualnij** takes a newer version;
**Wyłącz** stops a pack from that day, what it booked staying.

**The Marketplace, gradually.** Rhea's own packs first — Podstawa, Pełna księgowość and Środki
trwałe split out of today's pack with no change for existing companies, then KPiR, Ewidencja
przychodów, Rhea for Accounting Offices — then partners' (offices, consultants: the pack
workspace brief of 2026-10-18).

**Order of work** (step 1 of "Rhea Apps and business packs"):

1. *The core:* what stays in the kernel and its base vocabulary, what moves to the market pack;
   Rachunki i karty added.
2. *Installation per company and the Pakiety tab* — with the company-rules work.
3. *Today's Poland pack split* into Podstawa, Pełna księgowość and Środki trwałe — Polska, no
   change in behaviour for existing companies.
4. *KPiR.*
5. *Ewidencja przychodów.*

**Step 1 designed: the core, shipped by the kernel (2026-10-23).** The readiness review found
the core shipped by nothing: the founding vocabulary reached white01 through a restore of the
Helios simulation, the face's build tool inserts it active without an approval, a fresh server
fails at its first pack ("requires object type …"). The design, the simplest that keeps the
invariants:

- *The core is a pack the kernel carries in its binary* (`packs/core`, embedded): the founding
  vocabulary at its latest versions — company, account, posting, period_lock, fx_rate (the kernel
  names these), invoice, sales_invoice, purchase_invoice, vat_rate, with their views — the common
  packs Ludzie, Okresy and Dokumenty, and from the Poland pack Kontrahenci and Rachunki i karty's
  types (contractor, bank_account, bank_operation, settlement, matching_difference, task,
  task_schedule, with do-zrobienia), each byte-identical to the version the Poland pack carries
  today, so nothing changes where they already are. They load, and are approved, in the order
  Ludzie (a task names a person), the founding vocabulary, Okresy (its rule writes the period
  lock), Dokumenty; an approval out of that order is refused — a rule is approved only onto a type
  already approved or brought in the same bundle.
- *`rhea init` loads it as any pack is loaded:* what the database lacks is drafted into one bundle,
  what it holds identically is skipped (white01: nothing to draft), nothing activates without a
  person. On a new server, before any house exists, the operator approves it from the command line,
  as packs are approved there today; a later core version arrives as an upgrade the house approves
  (every company moves, D1). The kernel's own doors stay seeded active as before — they are the
  gate itself; the core's rules are behaviour, and behaviour waits for a person (invariant 2).
- *Not moved yet:* the rules and doors of Rachunki i karty still in the Poland pack
  (`pl-company-account`, `pl-explain-bank-operation`, `pl-classify-bank-operation`,
  `classify_bank_operation`) move with the loader's succession in step 3 — moved now they would
  read as conflicts; the law inside what moved — the invoices' and VAT rates' Polish fields, the
  contractor's NIP, REGON and cash method — is cut out when types can be extended (step 3).
- *What stops:* the face's build tool inserting types active, the walk loading the founding files
  by hand, the deploy copying the common packs — each takes the core from the binary.
- *Proved by tests:* a fresh database after `init` holds a waiting core bundle and, approved,
  everything the Poland pack requires (the Poland pack loads on it); `init` twice drafts nothing; a
  database built from the founding files, as white01 was, drafts nothing.

**Step 2 designed: installing per company (2026-10-24).** Built on the company rules of
2026-10-22: a pack's rules stay Rhea's rules, and what a company installed decides which of them
run for it — never copies of a pack's rules into each company (the readiness review, §1.6).

- *A pack says what it is.* Its file names its kind (core, market, regime, module, business), its
  market, the packs it needs (KPiR needs Polska — podstawa), and how it is installed: for the
  whole installation or per company. The approved pack version keeps this with it, so the
  installation knows its catalog. A pack whose needs are not available is refused at loading.
- *What runs where.* The core and the common packs run for every company, always. Today's Poland
  and Germany packs declare themselves for the whole installation, so nothing changes on white01
  or in the tests; the split (step 3) gives the new packs the per-company form, and every company
  that has today's Poland pack then gets the installs that keep its books as they are.
- *An install is a fact in the company's log*: "pack P installed from date D", and "switched off
  from date E", by a person who may (the company's owner, or an office under its mandate, as for a
  company rule) — explained, replayed, never a setting. One version for everyone (D1): an install
  names the pack, and the company follows its approved version.
- *Rules chosen by install.* For a company's fact dated D, a rule of a per-company pack runs only
  if the company had the pack installed on D; its own rules and Rhea's other rules run as today.
  Switched off, a pack stops from that day and what it booked stays. A fact dated before the
  install is not booked by the pack, unless the person approves booking the past (the backfill
  that exists).
- *What a company is served.* Screens, types, doors, words and Rhea's announced changes from
  per-company packs the company has not installed are not served to it, and such a door is refused
  before anyone's rights are asked. One place in the kernel decides it, which every read uses.
- *Pakiety.* Ustawienia → Pakiety lists the catalog in two columns, Dostępne and Zainstalowane:
  each pack's name, kind, version, what it needs and what it brings. Zainstaluj shows what the
  pack adds and asks from when (by default the company's first open month); confirmed by someone
  who may install, the pack is installed. Rhea's own proposals (at onboarding, from step 3) wait on
  Dziś as cards whose approval is the same act. Wyłącz asks from when.

Built in small steps, each with its failing test first: the pack's own description; installs and
switch-offs as facts; rules chosen by install (two companies, one install, a switch-off, replay
identical); what is served (every read path walked with two companies); the catalog; the face's
Pakiety tab, its install card and the card for a company with nothing installed. Left for step 3:
words merged across packs, readers run only for companies with their pack, screens found by their
role rather than their id.

*Built 2026-10-24:* the kernel in five commits (d0e3264 what a pack is, 39f976a installs as facts,
a7dc65f rules chosen by install, 7311a07 what a company is served, e6a3d85 the catalog), the face's
Pakiety tab and Dziś card in rheaface a4b055c; the live walk installs a pack from Pakiety. After an
install dated back, what waited for a rule is booked by the install itself, and Pakiety proposes
booking what other rules had booked and the new pack books further — the backfill the company
approves for itself (87fc2f2, rheaface 9b53b44).

**Step 3 designed: splitting the Poland pack (2026-10-24).** Four packs come out of today's:
Polska — podstawa (market), Fakturowanie — Polska (market, needs Podstawa), Pełna księgowość
(regime, needs Podstawa) and Środki trwałe — Polska (module, needs Podstawa) — the classification
of the readiness review, §3, with Rachunki i karty already in the core. In this order, each part
with its failing test first:

1. *A definition belongs to the packs whose files declare it.* Loading a pack records every
   definition its file declares — rules, doors, types, views, the identical ones included — with
   the pack's version, effective when that version is approved; which pack a definition belongs to
   is read from there, no longer from who wrote its row (a rule carried into a new pack unchanged
   still names its old pack as its writer). A pack names the packs it succeeds, so it may carry
   their rules forward and change them. Nothing is renamed: every rule id, view id, door name and
   master-data key stays as it is — a renamed rule would leave the old and the new both active,
   a double booking.
2. *Words from every pack the company has*, merged: the core's, then the market's, the regime's,
   the modules'.
3. *Each company's own chart* (D3) — *its design under discussion since 2026-10-24, see the D3
   brief; what follows is the first proposal, not decided*: the regime brings its country's standard chart as a proposal;
   installing makes it the company's; Helios's present chart becomes Helios's. A-codes declared by
   the packs that need them (tax treatment, statement line, the JPK_KR_PD tag); each account's
   role, and Rhea's rules booking to roles, never to numbers.
4. *The six links from the base back into full books cut*: VAT is settled when a month closes,
   whatever the books; the tax calendar reads declarations each regime fills; a matching
   difference carries its kind and the regime books it (the kernel no longer asks for an account);
   the supplier's cost account moves to the regime; the words and the warrant split.
5. *The cut*: four pack files, installed per company; every company that runs today's Poland pack
   gets the four installed from the beginning, so its books stay as they are — a replay proves it;
   tests change only where they name a bundle or a count.

Then step 4, KPiR.

Each pack declares its Apps as it is built or split (`APPS.md` is the register they start from).
The target catalog these packs grow into — a vision, measured against the AI-native newcomers and
the old guard — is `PACKS.md`, with the market research in `reports/`.

## Invoicing per country, and each company's own chart (KK, 2026-10-23)

KK's answers to the three questions of the readiness review (`reports/Installable packs readiness
2026-10-23.md`, §7), read before any pack is split.

**A new version of a pack reaches every company that has it.** When Rhea publishes a pack's new
version, every company that installed it moves to it, and Dziś shows what changed — as Rhea's rule
changes are shown now. A company that wants its own way for one rule has it already: its own rule
replacing Rhea's, with the warning (owed) when the rule it replaces rests on the law and changes.
A company staying on a whole older version comes later, if an office asks for it.

**Invoicing is a pack per country.** KK: *"Our Tylko client is planning to have invoicing in PL,
DE, NL, FR, CH, DK, NO and other countries. They look for an invoicing solution which can support
that and there is not too many doing that (none?). I think we need a separate per country
invoicing pack."*

- *What an invoicing pack holds:* the country's invoice law — what an invoice must say, numbering,
  wording and language, the rates it charges, corrections, the e-invoice format and the channel,
  and the dates its mandate bites. Fakturowanie — Polska (FA(3), KSeF in and out — leaving
  Podstawa); Rechnungsstellung — Deutschland (XRechnung, ZUGFeRD; receiving since 2025, issuing
  from 2027–2028); Facturation — France (Factur-X, UBL or CII through a registered platform, from
  September 2026); Nederland, Schweiz (outside the EU: MWST, the QR-bill), Danmark, Norge — each
  written from the country's law when its pack is made, its dates checked then.
- *Apart from the country's base:* a company installs the invoicing pack of every country it
  invoices in, without that country's tax returns or books — a Polish company invoicing in Germany
  installs Rechnungsstellung — Deutschland and keeps Polish books.
- *Which pack an invoice follows is the law's, never the person's choice:* in the EU the invoicing
  rules of the country where the seller is established, or registered and supplying from (VAT
  Directive 2006/112/EC art. 219a, with its exceptions for reverse charge and the one-stop shop);
  outside the EU each country's own. Rhea picks it from the seller's registrations and the supply.
- *What they share stays in the core* (proposed with this decision): the invoice's common shape —
  seller, buyer, lines, tax per line, currency, dates, number; the semantic model of EN 16931,
  which Norway's EHF follows too — so a company invoicing in seven countries keeps one sales
  ledger, one list of customers, one receivables. This amends the same day's "invoices move to the
  market pack": the shape stays core, the law moves to each country's invoicing pack. An EU layer
  could not own it — Switzerland and Norway are on the list.

**Each company has its own chart of accounts.** KK: *"Each company has its own chart of accounts.
Even more, we need to enable account numbers to have so called 'segments' — group of digits, where
segments can later be used for analytics. Each account can have additional analytical coding /
grouping for reporting purposes. I think we should offer a comprehensive country standard, which
later, based on interview can be modified by a client."*

- *The chart is the company's.* The regime pack brings its country's standard chart as a proposal;
  installing makes it the company's own, and the interview changes it — accounts added, renamed,
  split into analytics, or a migrating client's own chart kept. Helios's present chart becomes
  Helios's; a KPiR company has none and sees none.
- *Account numbers have segments.* The company declares the structure — say a three-digit synthetic
  account, then analytic segments (401-01-003) — and what each segment means (a cost type, a
  counterparty, a region); reports group and filter by any segment.
- *Account attributes — A-codes.* Beyond its number, each account carries attributes of its own
  that group it for reporting: the statement line it maps to (bilans, rachunek zysków i strat), in
  Poland the account tag JPK_KR_PD asks for (phasing in from the 2025 tax year), the tax treatment
  (one exists already: `account.tax_treatment`), a group's reporting line. The pack that needs an
  attribute declares it; the standard chart sets its values and the interview changes them. This
  is also how a pack adds to the account without re-versioning its type — the readiness review's
  §3: today the Poland pack's `account` v2 exists only to add `tax_treatment`. KK corrected the
  first reading: codes on the transaction — cost centre, project — are T-codes, a different thing
  and not this decision (they belong to Controlling, later).
- *Under discussion since 2026-10-24, not decided:* how Rhea's rules reach a company's own accounts.
  The bullet below is my proposal of 2026-10-23; KK's direction since is that there is no shared
  chart — a company has its own books and shares only Rhea's common rules — and that a company
  with its own chart teaches Rhea which of its accounts is which (her rule books to 401, the
  client keeps 401-001). The design waits until it is checked against the architecture and the
  law and agreed with KK (FUNCTIONS.md, M1).
- *Rhea's rules book to what an account is for, not to its number* (proposed, not decided). A rule books "receivables from
  customers" or "cost of external services"; the company's chart says which of its accounts plays
  that role, the standard's mapping being the default — the role carried as one more account
  attribute, declared by the regime pack. Without it a client's changed chart breaks Rhea's rules.
  Every literal account in today's full-books rules becomes a role (the readiness review, §3).
- *Country standards Rhea offers*, each checked when its pack is written: charts the law imposes —
  France's Plan comptable général, Spain's Plan General de Contabilidad, Belgium's PCMN, Czechia's
  směrná účtová osnova; standards of practice — Germany's SKR03 and SKR04, the Netherlands' RGS,
  Switzerland's KMU-Kontenrahmen, Norway's NS 4102. Poland's law sets none (UoR art. 10 ust. 1
  pkt 3 lit. a leaves the zakładowy plan kont to each entity's policy): Rhea's Polish standard is
  the model chart of practice, groups 0–9.

**What changes in the order of work.** The split (step 3) cuts four packs from today's: Polska —
podstawa, Fakturowanie — Polska, Pełna księgowość and Środki trwałe — Polska. Before it, accounts
get attributes (A-codes) and roles, and the chart becomes the company's (the installation-wide
chart was one of the review's blockers). Segments in the account number come after KPiR, before an
office brings over a client's chart. Other countries' invoicing packs follow demand; Tylko's list
is the first — Tylko has not chosen Rhea (KK, 2026-10-23): it will have a company registered in
each country, wants invoicing now, sells 80% to consumers and 20% to businesses, and may want the
books in Rhea too.

## The App designer (KK, 2026-10-23)

KK opened the designer in talk the same day — step 6 of "Rhea Apps and business packs": who
designs, the panes, versions, and what happens to an App's running processes when its production
version changes. KK's list, the shape proposed in the talk, and KK's four decisions (DECISIONS
2026-10-23, "The App designer").

**The Designer.** A role in the people pack beside owner, chief accountant, accountant, reader and
support, and a paid one: the grant is what is billed. Only a designer, or the owner (who holds
everything), designs an App and puts a version into production. Its rights are a new family in
the vocabulary beside door:, view: and type: (proposed: `app:design`, `app:release`).

**Rhea designs; the panes check and adjust.** The usual way in is a sentence — "design me an App
for purchase invoices" — and Rhea drafts the whole App. The designer is first the page that shows
that draft; its panes are where a person checks it and changes details. Rhea and the panes edit
the one definition (JSON, validated before anyone approves): whatever can be clicked can be said,
and the other way round; nothing lives only in a pane. The preview runs the draft on the company's
own data and writes nothing.

**Who asks, who accepts, who releases** (KK's decisions).

- Anyone may ask Rhea for an App or a change. Rhea prepares the draft as a proposal; the person
  who asked may accept it — "this is what I meant" — and the acceptance is recorded on the draft.
- A designer verifies it and puts it into production; the owner may too. KK: in 99% of cases the
  designer, after an oral agreement. The production card names who agreed — the requester's
  acceptance when there is one, otherwise the person the designer names, the designer included
  (KK) — kept with the approval.
- A designer from outside the company (a partner building Apps for many firms, later the
  Marketplace) sees in the sandbox only what their own rights show; otherwise Rhea's sample
  company.

**General properties.** Name, category, folder; read-only (a report) or data input; with or
without a workflow; with or without a visual UI (without one, an App that runs on triggers alone);
who may run it. The kind declared decides which panes the designer opens, and the kernel checks
it: a read-only App with an action is refused.

**Visual.** Components in flex rows; a container holds components and arranges them the flex way
(proposed: one concept — a row is a container laid across, the page a container laid down). Width
relative to the parent: 1/6, 1/4, 1/3, 1/2, 2/3, 3/4, 5/6, 100% — twelfths, so rows line up across
the page — Fit or Fixed; min and max width in px; height auto or fixed in px. Gaps come off before
the fractions, so 1/2 + 1/2 always fits. Min width makes the phone: four quarter-width fields with
a 160 px minimum stand in one line on a desk and wrap on a phone, with no separate phone design;
narrower than a phone, everything is 100%. Components carry no colour, font or padding of their
own — width, height, label and a variant from a closed list — so every App looks like Rhea, a
brand change is made once, and the face's consistency check covers Apps. The repository is the
desk's gallery: the table with its layouts (first), the document sheet, the statement, figures,
Rhea's words, the board; form fields (text, number, money, date, choice, a record picker, yes/no,
file); container, tabs, button. A component nobody has is a request to Rhea's makers, never built
inside an App.

**Variables and queries.** Every property of a component — its default value, whether it shows
(the whole, or a part with properties of its own), whether it may be edited, whether it is
required — and every transition's condition, every action's condition and parameters, is a fixed
value or a formula over the App's variables; the state is one more variable. A variable is set by
input (with a default formula until the person types), by a formula, by a query, or by a parameter
the App opens with (the month). For speed: a component's formula reads variables and never runs a
query; queries run when the App opens or a parameter changes, formulas recompute as the person
types. The designer refuses circles (A from B, B from A) and shows for each variable what it reads
and what reads it. Queries only read — the books and other Apps' data within the person's rights,
the App's own tables. Writing happens only on a transition: an action records a fact, approved
rules book it.

**Workflow.** States; transitions, the label being the button; who may move; a condition; actions,
each with its condition and parameters. A transition is fired by a person or by a trigger (a fact,
the clock, another App's state, a condition coming true), each automatic one with its autonomy
(act, propose, earn it). Who sees and who edits which field in each state is one table of ticks,
fields by states, not a condition on every field. A snapshot of the variables at each move.

**Data tables.** Object types the App owns, each the company's by construction: a designer cannot
make a table another company sees.

**Versions.** Rhea's lifecycle as it is: the development version is a draft, the production
version the active one, earlier ones superseded; putting a version into production is an
approval, a fact with who and when. *Restore production to development* starts a new draft as a
copy of what runs, the unfinished draft set aside and kept. *Back to an older version* is a new
version with the old content (v7 = v5), the history whole.

**Fixed identities.** Every state, transition, field, variable and component gets an identity when
it is made; its label is free text. Renaming is never removing: "Weryfikacja" renamed
"Sprawdzenie" is the same state. When Rhea edits by talk it keeps identities, and a draft that
drops one and adds another under the same label is asked "is this a rename?" before anything else.
Most upgrades are automatic because of this.

**Running processes: upgrade by default** (KK). A running process is its list of facts — started,
filled, moved — so upgrading reads the same facts under the new version and records one more:
"moved from v4 to v5". Nothing is rewritten; what happened under v4 says v4 forever, and replay
reads each process under the version it had at each point. After the move the new version holds in
full, rights included: if v5 says only the manager approves above 10,000 zł, a 12,000 zł invoice
waiting for the accountant now waits for the manager — usually the reason for the change. One
switch per version, *let the running ones finish on the old one*, serves a change dated by law.
Warrant for the shape: workflow engines keep an instance on its version and move it by an explicit
plan mapping old steps to new ones, checked before it runs (Camunda's process instance migration);
Rhea draws the plan itself.

**Breaking changes are found in design** (KK: "what I'm worried are breaking changes, which must be
identified in design"). The designer compares the draft with production all the time, by identity,
and classes every change the moment it is made — marked beside it in its pane, counted on the
company's own data — and a list, *Zmiany względem produkcji*, gathers them in three groups.

- *Free* — nothing already there notices: layout, labels, help texts, a new optional field, a new
  state or transition, a new variable, wider rights.
- *Visible* — nothing stops, but something already there changes, and the designer shows it: a
  changed formula (each running process's values, before and after), narrower rights (who now
  waits for whom), a new condition on a transition (processes that could move and now cannot).
- *Breaking* — something already there cannot go on as it is without a decision:
  1. *the flow*: a state removed, merged or split with processes in it; a process left with no way
     out;
  2. *the data*: a field removed (its values stay in the history; the new version neither shows
     nor computes them); a field's type changed (the values that do not convert, listed: "3 records
     say 'brak' where an amount is now required"); a new required field that records lack; a
     choice removed that records hold; a link pointed at another table;
  3. *what others rely on* — the dangerous kind, invisible in the App's own panes: a field of a fact
     the App records that an approved rule reads (the books), and a variable, field or state that
     another App's formula, query or trigger, or a report, reads. The designer knows every reader,
     the rules and Apps being definitions.

Each breaking change carries its answer in the draft itself: a mapping (Rhea proposes, the designer
confirms), a fill formula or a default for a new required field, a conversion with the failing
values listed, the removal confirmed, or "these finish on the old version". The facts an App
records are never changed in place: a field may be added to a fact, never removed or retyped; a
different meaning is a new fact, and the rules that book it come in the same package (the
event-sourcing practice: a recorded fact never changes meaning — Greg Young, *Versioning in an
Event Sourced System*). A draft with a breaking change unanswered cannot go to production; the
button says why. *Try the upgrade* in the sandbox runs it over a copy of every running process and
record and shows what moves, what changes and what stays — the list the production card repeats:

> v5 of Faktury kosztowe: 41 in progress. 38 move on their own. 2 sit in "Weryfikacja", which v5
> no longer has — Rhea proposes "Do akceptacji". 1 stays on v4 and finishes there: v5 requires an
> MPK it lacks; it moves once someone gives one.

**What a new version does to what is already there** (Claude's proposal, 2026-10-27; KK: "Yes, i
think you can add this"). In code, data sits in tables and a migration script rewrites them, once. In
Rhea the facts stay as they happened and the records are what the rules make of them, so a version's
own rules are its migration: nobody writes a script, the look-back applies the rules to the facts the
books already hold, every record it adds says which old fact and which rule made it, and replay
reproduces it. KK asked whether this works on invoices only: on every fact — documents, bank lines,
what people typed into forms — and on every App, a client's own included. Three kinds, each its own
tool:

1. *It only adds* — something new drawn from facts already there: Klienci and Dostawcy from past
   invoices, subscriptions from a supplier billing the same amount every month. The look-back: a
   person asks Rhea, or she offers it once herself; she says what it would add; a person approves
   (DECISIONS 2026-10-27). Built; at volume it works in portions.
2. *It would have decided differently* — another account, another rate. What was booked is never
   rewritten: the difference is a correction entry, an accountant's deliberate act, as a korekta is.
   The look-back deliberately leaves it alone. Not built.
3. *It needs what no fact holds* — credit limits for 37 existing clients. No rule can invent them;
   Rhea proposes what she can see (each client's usual days between issue and due date) and a person
   confirms, many at a time — the fill a breaking change carries, done with the person. Not built.

The production card says which kind a version is — "Ta wersja tylko dodaje: 37 klientów, żadnych
zapisów" — and a version of the second or third kind says so and has no one-step approval. Two
limits: rules on dates (the 10th of each month) stay out of the look-back, their catch-up being the
clock's; a field that moves (the supplier's cost account, from the firm to Dostawcy) fills its new
place from the old facts, and the old place keeps its history, as an old paper form does.

**The first App to design with it: purchase invoices described and approved.** Most purchase
invoices from Polish suppliers already arrive through KSeF, so Rhea's first question is which job
is meant — describing and approving those that arrive (opis merytoryczny, MPK, akceptacja), or
entering those from abroad. For the first: *Faktury kosztowe — opis i akceptacja*,
Księgowość/Zakupy, data input, with a workflow; the Poland pack's purchase invoice plus the App's
own table *opis* (MPK, projekt, opis); variables — gross, a limit of 5,000 zł the office changes,
over the limit, the supplier's account on the white list; the layout — supplier 1/2 · NIP 1/4 ·
date 1/4, the lines 100%, opis 2/3 · totals 1/3, the white-list warning shown only when the account
is off the list; the flow — Nowa → Opisana → (Do akceptacji, over the limit) → Zaakceptowana →
Zaksięgowana, and Odrzucona; an invoice arriving from KSeF opens one in Nowa on its own; nobody
approves their own invoice; the approval records a fact the pack's rules already book.

**What it needs from the kernel**, in the order of "Rhea Apps and business packs": a running
process knowing its version, and an old version still running for the processes that stay (step
3); identities and the comparison with production; the readers of every fact and field, known from
the definitions. The read-only designer page comes early (`APPS.md`, step 1); the editable panes
after the App runs (step 6).

## The learning process: a lesson in words, taught once for many clients (KK, 2026-10-22)

KK, in the talk about the MVP: the one key that opens an accounting office is automation
that lets its accountant keep more clients with the same effort — any automation toward
that is the key. And: "We still need to implement the learning process real good. What
we have today is machine language." With it, a community around "a self-learning AI
system", perhaps sharing "Rhea's useful prompts". The buyer is the office, with sole
traders and companies in one system (KK, the same day).

**What she learns today, and how it reads.** Five ways, all real:

- *Rules drafted by the author* from what waits in the worklist, approved by a person.
  The card shows the rule's description and its warrant — prose the author wrote beside
  the rule, not read back from it, so it can say more or less than the rule does — and
  the formula "as the rule wrote it" (rheaface `web/app.js`, the why panel). That is
  the machine language.
- *Facts said through doors*: which account a supplier's invoices go to
  (`set_cost_account`), whether an account is a tax cost (`set_tax_treatment`), what a
  bank line is (`classify_bank_operation`), the company's tax and ZUS accounts.
- *What a person does*: after two pairings of the same kind, she drafts the matching
  rule herself (`internal/shell/activities.go`, `proposed`). Only matching learns this
  way.
- *Corrections*: an overruled reading never returns (`internal/agent`, `Corrections`;
  "The host remembers").
- *Figures taught on request* and *what she was never taught* reaching the house (the
  Learning Centre); the network counts rule shapes across installations above a floor
  of two (`internal/network`).

**What is wrong with it for an office.**

1. The person who approves a lesson cannot read it.
2. The reach is wrong both ways. A rule has no company (`rule` in the store has no such
   column): once active it books for every company of the installation, whoever
   approved it. A fact said through a door is one company's: an office with eighty
   clients says eighty times that Orlen is fuel. Neither is "this office's clients".
3. She learns from what people do only in matching.
4. A lesson has no record — how many documents it booked, how often it was overruled —
   so trust cannot grow lesson by lesson: every model reading of a PDF waits (the gate
   at 1.00) and every active rule books unconditionally.

**The lesson, proposed.** What she was taught, as a person reads it:

- *The sentence* — Polish, read back from the definition itself by the kernel (a
  template per kind of effect, the warrant's source quoted), never written beside it.
  A model may smooth it under the host's laws: every number and name it says must be
  the rule's, and every condition of the rule must be spoken (the second law, applied
  to lessons). So she cannot claim to have learned more, or less, than she will do.
- *The examples* — the documents it would have changed, with the amounts before and
  after: the simulation the kernel already runs for a bundle.
- *The reach* — two kinds of rules (KK, 2026-10-22; DECISIONS). A *company rule* is
  approved inside its company and books only there. A *global rule* is approved by Rhea
  staff and books for every company from its date, announced to each with what it means
  for them. An office teaching its clients approves once, and each client it picks gets
  the lesson as its own company rule. Only staff promote a company rule to global, and
  only on proof: the same rule held by at least ten companies kept by at least three
  offices or installations, counted by the network from the rules alone (simulated
  companies excluded), optionally with an external biegły rewident's opinion.
- *The origin* — who taught it, when, from which document or which words.
- *The record* — documents it booked, times it was overruled, the last of each.
- *The trust* — it *proposes* (each result waits for one tap) or it *books* (results go
  straight through and are listed afterwards). A lesson starts by proposing and books
  once its record says so; a person may move it either way; an overrule moves it back.

The formula stays one tap behind the sentence, for the one who wants it.

**Three ways she learns, one card.** (1) *From what the accountant does*: the same
correction twice — a supplier's invoices moved from 409 to 402, a bank line classified
the same way — and she drafts the lesson, the two corrections its examples; matching's
two pairings, generalised to every door that corrects a booking. (2) *From what the
accountant says*: "Orlen to zawsze paliwo" in her box; the author drafts, the card is
the same. (3) *From other offices*: shapes approved by at least the floor of
installations come as priors, the card saying how many approved them — real counts,
never estimates.

**What the invariants demand.**

- *A lesson is a definition*: drafted, approved by a person, versioned. The sentence is
  shown; the definition runs (invariant 3).
- *Reach and isolation* (invariant 9): a rule gains its company, blank for a global
  rule. A company rule fires only on its company's facts, is seen only in a scope that
  names its company, and never reads another; a firing for client A never reads client
  B. A company rule that answers what a global rule answers names the global rule it
  replaces for that company: the company's reading books, with a warning on its card
  and on the office's desk where the global rule rests on a statute, and the two never
  both book — that would be a double booking.
- *One chart, or roles*: "Orlen → 402" across clients holds only where 402 means the
  same in every client's chart. Lessons across clients need the smart documents'
  account roles (book to "fuel and energy", the client's chart maps the role), or they
  stay per client until those land.
- *Earlier documents keep their version*: a lesson changed in June books May's late
  invoice under May's version — the smart documents' first question, the same fix.
- *The record is derived*: counted from the log (facts booked by a rule version,
  overrule events), never kept as state of its own. The switch between proposing and
  booking is an event, so replay through the same events gives the same books
  (invariant 4).
- *The boundary checklist*. Which company: the event's, and the lesson's reach is
  checked at firing. Two at once: two accountants teaching the same lesson make two
  drafts; approving the second offers it as the next version of the first, the lesson
  keyed by what it explains and its reach, never by its text. Twice: the same
  correction recorded twice proposes once. Money: an amount a lesson fixes (50% of a
  car's VAT) is an exact fraction in the formula language, and the sentence prints it
  from the rule.

**Proofs, in order.**

1. *The sentence*: every rule of the Poland pack read back in Polish; a test that each
   sentence names every condition and every amount of its rule, and that a changed rule
   changes its sentence. The card shows the sentence and the examples; the formula sits
   behind "szczegóły".
2. *Two kinds of rules, taught once for the office*: a test world of two offices and
   four clients. A company rule approved in client A books A's documents and no one
   else's. The office approves "Orlen → fuel" once for A and C, and each gets its own
   company rule; B, not picked, still waits. A global rule books for all four; C's own
   rule replacing it books for C with the warning, and the two never both book.
   Promotion is refused to anyone but staff and below ten companies in three offices.
   No read crosses a scope; replay is identical.
3. *Learned from doing*: two alike corrections on any client bring the card; accepted,
   the third document books by itself.
4. *Trust earned*: a lesson proposes, earns booking, is overruled once and proposes
   again; the record shows it all.
5. *Measured*: the business generator runs a quarter of an office's mixed clients with
   lessons on and off. The number is the accountant's touches per document, month by
   month — the learning curve — with wrong amounts at zero. It is the figure for the
   offices and for the community.

**The community.** Prompts are the on-ramp: questions that work, published on rhea.plus
and LinkedIn, each checked before it goes out, never a client's data. The face already
keeps her questions per kind of thing (rheaface 7e71db9), so "the most asked" is
counted, not guessed. The lasting asset is the library of lessons: the network's lessons
in their Polish sentences, each with its legal source, how many offices approved it,
and — with consent — the accountant who taught it first. An accountant's name on a
lesson five hundred firms use is advertising for their office, a reason to join that
prompts do not give. This is the pack workspace (DIRECTION 2026-10-18) opened to the
profession; the Learning Centre is its back room.

**Settled by KK (2026-10-22)**, as proposed:

- *When a lesson earns booking*: a run of accepted proposals the office sets, twenty by
  default, none overruled — a number to tune on the generator, not a claim.
- *The client's veto*: a client's owner sees the lessons that book their company and may
  switch one off for it; the office sees that it was.
- *Credit in the library*: the accountant opts in; a client is never named.

**Order of work**, after documents come in by themselves — without documents arriving
there is nothing to learn from — except the two kinds of rules, which come first of all,
before a second owner or an office is admitted anywhere: (1) the sentence and the card;
(2) the office teaching many clients at once; (3) learning from doing beyond matching; (4) trust earned; (5) the measure in the
generator; (6) the library on rhea.plus.

**The two kinds of rules: built in the kernel, 2026-10-23.** A company's rule books its
own facts alone (18091a2) and replaces Rhea's for it alone, never beside it (41311ee);
what is drafted or learned for a company is the company's (6eee0aa, d0fb4fb); a company
approves its own, Rhea's staff — the house — the rest (c8cc8c8); an office teaches its
clients in one act (7cbd83c); a client switches a lesson off (a4c82d0); a reading ten
companies under three keepers hold becomes Rhea's by staff (d6d9010); Rhea's changes are
the installation's facts, listed to every company (36b31ba). Left: the face's half (the
changes on Today, the statute warning, the client picker, the switch-off, the staff's
promotions page), and sorting an installation's existing rules into Rhea's and a company's
before its second company is admitted.

One line: *a lesson is a sentence she can be held to, taught once for the office, earned
into automation by its record, and shared as the profession's library.*

## Tasks, groups and the phone: who acts on an approval, and how they hear of it (KK, 2026-10-27)

On Helios's desk Rachunki bankowe showed nothing while 53 statement lines were in the books: the
core's bank account record exists and nothing in the Poland pack fills it. Asked, Rhea said she
was never taught it and drafted a skill that makes any account registered live at once. KK: "such
important metadata like bank accounts should be approved by someone responsible for bank accounts,
this might be an accountant or chief accountant. Accountant should check and validate the data
before accepted to the system" — an App that adds a bank account, does the mechanical checks, and
has a workflow with an approval step. Agreed as the first App with a workflow (step 3 of "Rhea
Apps and business packs"). With it KK named the two things every approval lacks: how a person
learns there is something to do, and who "someone responsible for banks" is.

**Two approvals, not one.** Teaching Rhea how bank accounts are registered is approved once, by
whoever releases Apps. Accepting each bank account is approved every time, by the people
responsible for banks. The draft on Dziś had only the first.

**A table is defined by one App and used by all** (KK, correcting Claude): "a Data Table can be
defined by one app, but this Data Table should be accessible system-wide. Other apps can use it
both for read and write." APPS.md's one owner is where a table is defined, never who may touch it.
Because any App may write a table, an approval guards the table, not one App's screen: a bank
account written by any App arrives proposed, and only the approval step makes it active.

**The bank account, as the proof** (proposed, Claude; KK: in Rachunki bankowe, "add a button to
launch a tab").
- *Proposed* from three places: a person, by the button; a statement from an account nobody
  registered (every line names its account); the white list showing an account the company
  declared for its NIP that is not registered.
- *Checked* at once by the kernel and the market, no model: the IBAN's checksum and length for its
  country (ISO 13616, every market); the bank read from the number; not registered already; the
  currency; in Poland, on the white list for the company's NIP, the Ministry's requestId kept as
  the proof of the check. Each result shows on the record, not only pass or fail.
- *Approved* by a member of the responsible group who did not propose it, unless nobody else is in
  it — then alone, and the record says so. At approval the ledger account it books to (130-01,
  130-02…): every statement line books to 130 today, right with one account, wrong with two.
- *Closed* with a date; a statement for it after that day is flagged.

The supplier's bank account goes the same way next. Today a purchase invoice's account replaces the
supplier's without a check (the Poland pack's "Rachunek dostawcy z faktury zakupu (KSeF) uzupełnia
IBAN kontrahenta"), the shape of the common "we have changed banks" fraud. No money leaves through
Rhea yet, so it is owed before payments, not urgent.

**Roles are the system's; the business is groups** (KK): "By roles I meant the built in roles in
the system, and we have to have them … I think these are enough. The rest we can manage by
Groups."
- *Rhea Designer* — designs Apps and runs interviews.
- *Tenant Owner* — owns the tenant's contract with Rhea.
- *Office Manager* — manages the accounting office's parameters and master data.
- *Company Manager* — owns a company, which may be an accounting office's client.
- *User*.
- *Guest*.
- *Rhea Support* — Rhea's own staff, who approve Rhea's shared rules ("That's the Rhea Support
  Role. Not a group." KK, answering open point 2); in a company's books only by that company's
  timed grant, as decided 2026-10-21.

"A group has no role assigned" (KK). "We mostly should use groups. Roles are more suitable for
commercial ownerships of contracts and technical parameters setup" (KK). A group holds people and
other groups; an App's steps, and who may use it, name groups. This replaces the role of 2026-10-19
("People and rights" above: a named set of rights, data in a pack, a company able to declare its
own): of the people pack's five roles, owner becomes Company Manager, support becomes Rhea Support,
and chief accountant, accountant and reader become groups (KK: "Chief Accountant, Accountant -
should be groups").

**Groups.** KK: "If Rhea, while designing an app hears 'This should be approved by someone
responsible for Banks', she could or even should create a dedicated User Group for that purpose and
then ask who to add to this Group."
- Rhea proposes the group with the App's draft and asks who is in it, naming the company's people:
  "Kto ma być w Bankowości?"
- Who adds a person to a group that approves (KK): the Rhea Designer, the Tenant Owner, the Office
  Manager, the Company Manager.
- A client's step may name a group of its accounting office, under the mandate (KK: yes) — one
  accountant spread over many clients.
- Proposed (Claude): members are people or groups, and a circle is refused; membership has dates,
  as a company membership does, so the books answer who could approve on a given day; an App does
  not go into production with a step whose group is empty.
- A system group *All users* (KK: "if by chance you need to add all users to an app"), whose
  members Rhea keeps and nobody adds by hand: a company's own people, never a Guest, never another
  company's; an office's people reach a client through the office's groups (Claude's proposal, KK
  confirmed).
- A User in no group sees Dziś and their own tasks, nothing else (KK, agreeing).
- The groups a company starts with — Chief Accountant (Główny księgowy), Accountant (Księgowość),
  Reader (Podgląd) — hold what today's roles of those names hold, with its Company Manager in each
  until handed over, so a sole trader does everything on day one (Claude's proposal, KK confirmed).

**Zadania.** KK: "the 'Dziś' activity needs a new function: Worklist or we can call it 'Zadania' or
'Todos', which would be the list of tasks a user has to do. A normal table but with a) inline
buttons and possibility of 'bulk action' = select several tasks (checkbox them) and if they have
the same actions to do like eg. Approve or Reject, display those buttons on toolbar. The same Todo
lists should be in every activity but prefiltered to apps from that activity." Dziś's card tells
of them. Proposed (Claude):
- One list read from what waits — every App step waiting for a person or a group is a task — and
  kept nowhere of its own, so it cannot disagree with the books and replay rebuilds it.
- Dziś's decision cards (a skill waiting, a statement line nothing settled, the month to close, the
  look-back) become tasks, so the card and the list always agree.
- Columns: what, the App, the record (it opens), since, by when, for whom (the person, or the
  group's name).
- Twelve ticked are twelve acts, each signed and each checked on its own; one refused stops none,
  and the toolbar says which and why.
- A group's task is settled by the first member who acts; the others see who did.

**The phone.** KK: "One of the things i would ultimately like to have is a very good notification
on mobiles … we should even require that."
- Required for everyone who approves (KK): Rhea puts nobody into a group that approves before one
  of their devices receives notifications.
- Proposed (Claude): Web Push through the face, which already installs as an app (its manifest and
  service worker) — Android at once, an iPhone from its home screen (iOS 16.4 and later). What it
  needs is the device's permission, asked when the passkey is enrolled on the phone, not the phone's
  number; no SMS.
- Proposed (Claude): a notification opens the task and never approves from the lock screen —
  approving stays a signed act of a signed-in person; the lock screen names the kind of task and
  the company, never an amount or an account number; tasks arriving within minutes make one
  notification.

**Order** (proposed): groups → a workflow step naming who may move it (step 3) → Zadania → the
phone. The bank account App proves all four; the supplier's account follows.

**The tenant** (KK agreed the proposal): a level the kernel does not have yet — the contract with
Rhea. A tenant holds an accounting office with its clients, or companies directly; the Tenant Owner
manages the contract and, by that role alone, sees no company's books.

**A statement from an account not yet approved** (KK agreed the proposal): its lines wait, and one
approval releases them all; nothing books to an account nobody has checked.

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
  only, no auth (experiment non-goal). Brought forward by smart documents
  (KK, 2026-10-19): who may edit, see and move a document is part of its
  definition, per state.
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
- **Matching is a kernel method with N sides** (KK, 2026-10-11, the
  matching brief above): a deterministic engine, total and bounded, with
  the structure (sides, measures, converters, tolerances, keys) and the
  policy (an ordered list of named strategies) as data; every match a
  baked, explained event; the grid is the remainder. Earned the
  closed-class effect `matching` by the attempt (DECISIONS 2026-10-11);
  the engine's core shipped the same day; a dedicated session.
- **The company key is the partition key** (KK, 2026-10-08, "Scale and
  security are welded"). The one scale item that is cheap now and
  expensive later: the company ref promoted from payload field to a key
  the store carries on every call — isolation and sharding ride on it
  later, enforced by nothing yet. Alongside: snapshot-plus-tail replay,
  a hash-chained log, and the network test that no fact ever leaves.
- **Free is the network's economics; open source is the trust answer**
  (KK, 2026-10-14, "Rhea as a public free service" above). A rule is
  written once per situation, not once per company, so a free tier costs
  the network almost nothing per user; the kernel and packs go open
  source, the hosted network stays the service; security and the
  statutory floor (KSeF outbound, JPK_V7) precede the first stranger;
  accountants are the channel; ads are struck; the face is built apart,
  generic, with the plain shell as the developer's truth. Decided the
  same day: open source with a quality floor, security as defensibility
  before an audit, April 2027 the latest; measurement first, then KSeF
  outbound.
- **A process is one definition: the smart document** (KK, 2026-10-19,
  "Smart documents" above). States with their paths and a terminal state;
  per state who edits, sees and moves; variables set by input, formula or
  query, scalars or collections, snapshotted at every move; instances as
  workflows that may start and move others under entitlement; a
  declarative view — one definition the kernel turns into the four
  vocabularies, approved as one package, so the guarantees hold by
  construction. It never writes the books itself, its calculations always
  finish, an API call is two facts, running instances keep their version.
  Rhea's apps (WMS, OMS, document flows) are packages of smart documents.
  A dedicated session; the first proof rebuilds prepaid costs as one.
