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
  tenancy stays out of the experiment.
