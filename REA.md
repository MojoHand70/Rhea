# REA.md — OeBTO / ISO 15944-4 alignment

Rhea's language descends from REA (McCarthy 1982). ISO/IEC 15944-4:2015
standardizes REA as the Open-edi Business Transaction Ontology (OeBTO).
This file records the term-by-term cross-check (2026-10-05) so the language
can borrow standard vocabulary deliberately and diverge on purpose.

Sources: McCarthy, "The REA Accounting and Economic Ontology — Its Use in
ISO 15944-4" (2013 ISO presentation, valueflo.ws/linked-docs/REA-Ontology_ISO-15944-4--BillMcCarthy_20131107.pdf);
ISO/IEC 15944-4:2015 abstract (iso.org/standard/67199.html; full text paywalled).
Notable from the presentation: McCarthy credits REA as "the basis for newer
types of ERP systems (Workday and others)"; accounting artifacts are derived —
"debits/credits/ledgers are not necessary in such a model."

## The three OeBTO levels and where Rhea puts them

OeBTO stratifies into three levels (slide 18):

1. **"What has occurred"** — Economic Resource, Economic Event, Economic
   Agent; relations duality, stock-flow {inflow, outflow}, participation
   {from, to} {inside, outside}. → Rhea's **event log + derived objects**.
   Same stance: events are primary, everything else derives.
2. **"What could be or should be"** (type/policy level) — Resource Type,
   Event Type, Agent Type; relations typify, policy, specify. → Rhea's
   **object_type + rules**. OeBTO declares a "repository for business
   rules" but gives it no execution semantics, no versioning, no approval,
   no provenance; Rhea implements the policy level as versioned, AI-authored,
   human-approved data executed by the kernel. This is the lineage claim in
   one line: Rhea is OeBTO's policy level made executable.
3. **"What is planned or scheduled"** — Economic Commitment, Economic
   Agreement/Contract; relations specify, fulfills, reciprocal, trigger;
   Agreement governs Business Process. → Rhea's **weakest coverage**; see
   candidates below.

## Term map

| OeBTO term | Rhea construct | Status |
|---|---|---|
| Economic Event | event (append-only log) | aligned — identical stance |
| Economic Resource / Agent | object types (contractor, bank_account, fixed_asset, …) | aligned at pack level; the kernel stays domain-free, so the REA taxonomy (Goods/Services/Rights; Individual/Organization/Public Administration) is pack vocabulary, never kernel structure |
| Resource/Event/Agent **Type** + typify/policy | object_type + rules | aligned — Rhea executes what OeBTO only declares |
| **Duality** (requited event pairs: shipment↔payment, give↔take) | emerges from rules (settlement; intercompany twin postings from one root fact) | aligned in effect, unnamed in the language; name it as a rule-pattern term, not a kernel word |
| **Economic Claim** — materializes from an event, settles by an event; states materialized→settled | derived receivable/payable state; settlement events; payment ladder; aging buckets | aligned — corpus reconciliation spec is claims machinery; OeBTO confirms claims are derived, never primary |
| **Economic Commitment / Agreement** (fulfills, reciprocal, governs) | recurring_invoice_template, task_schedule, lease_contract, recurring costs | **candidate** — these types are commitments in OeBTO terms; see below |
| Business Transaction phases (Planning→Identification→Negotiation→Actualization→Post-Actualization) | status ladders per document (corpus) | noted, not adopted — evidence-derived ladders are finer-grained and real; the ISO macro-ladder adds nothing executable |
| **Regulator** constrains Business Transaction; Third Party; mediated transaction | statutory adapters (KSeF) | aligned — KSeF is a mediated transaction with a Regulator in OeBTO's exact sense; usable for positioning and adapter-contract naming |
| Trading-partner vs independent view of one inter-enterprise event | intercompany derivation from a single root fact ("nothing to reconcile") | aligned — Rhea materializes the independent view, which is why the two sides cannot disagree |
| Exchange (give/take) vs Conversion (use/produce) processes | finance pack = exchange; warehouse pack (receipts, stock, valuation) = conversion | noted — future manufacturing vocabulary can borrow use/produce event naming |

## Candidate adoptions (parked, not shipped)

1. **Commitment as a pack-level semantic.** Template/schedule/contract types
   share one shape: terms + cadence, with later events *fulfilling* them
   (M5 time events firing rules that emit events from templates is literally
   OeBTO's commitment→fulfills→event). If/when M5 lands, consider a
   `fulfills` provenance convention linking emitted events to their
   commitment object, and the word "commitment" in corpus vocabulary for
   this family. No kernel change required.
2. **Duality as a rule-pattern name** in corpus/drafting context (the agent
   benefits: "draft the dual" is a complete instruction for requital rules).
3. **Regulator/mediated-transaction vocabulary** for the adapter contract
   docs — free standard alignment for the statutory boundary.

## Deliberately not taken

- The resource/agent subtype taxonomy into the kernel — invariant: the
  kernel never learns a domain.
- The five transaction phases as a workflow notion — Rhea has no workflow
  engine; ladders are derived fields from events (CORPUS.md discipline).
- OeBTO's collaboration-space/BSI messaging architecture — inter-enterprise
  protocol, out of experiment scope.

## Positioning lines (sourced)

- "Rhea's core language aligns with ISO/IEC 15944-4, the ISO standardization
  of the REA accounting ontology — including its declared business-rules
  repository, which the standard never gave execution semantics. Rhea does."
  (Say *aligns with / descends from*, never *compliant with*.)
- "McCarthy's ISO presentation credits REA as the basis of Workday's
  generation of ERP systems; the ontology shipped as object models, but the
  policy level stayed hand-authored. Rhea is the policy level, authored by
  AI under human approval." (Attribute the Workday claim to McCarthy.)
- No REA + LLM work exists anywhere as of 2026-10 (landscape survey,
  reports/Products similar to Rhea.md) — the lineage is unclaimed.
