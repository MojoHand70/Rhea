# Commercial AI-native ERP and accounting products resembling Rhea (as of October 2026)

Rhea's six-point signature being tested against: (1) AI authors business rules as versioned strict-JSON *data*; (2) draft → human-approved → active rule lifecycle; (3) deterministic kernel over an append-only event log; (4) wipe-and-replay reproduces identical state; (5) per-field provenance (event_id, rule_id, rule_version); (6) statutory domain shipped as data packs.

Headline: no commercial vendor found matches the full architecture. The two closest partial matches are **Lensing** (plain-language business-logic authoring + sandbox simulation + approval before deployment — Rhea's point 1+2, without points 3–5) and the open-source **ERPClaw** (hash-chained append-only ledger + AI-generated modules gated by rule-checks and human approval — points 2+3-ish, without rules-as-data, replay, or per-field provenance). Everyone else puts AI at the *transaction* layer (drafting journal entries, reconciliations, accruals for approval), not the *rule* layer.

## Key question 1: What do Rillet, DualEntry, Doss, Keel, Nominal, Campfire, Light, Everest/Lensing (and peers) actually automate — and does any let AI author business rules/logic with human approval?

### Takeaway
The funded AI-native ERPs automate accounting *work* (journal entries, reconciliation, accruals, consolidation, close) under human approval — the AI drafts transactions, not rules. Only Lensing (ex-Everest) markets AI authoring of business logic itself ("build agents and apps in plain English" with sandbox-then-approve), and only Light's manifesto gestures at runtime-learned behavior; neither expresses rules as inspectable versioned data the way Rhea does.

### Cited Findings

**Rillet** (full-stack AI-native GL, mid-market SaaS)
- Founded 2023 by ex-N26 US CEO Nicolas Kopp, launched publicly 2024; $25M Series A led by Sequoia (May 2025) after a $13.5M seed; $70M Series B co-led by a16z and ICONIQ — [ERP Research AI-native roundup](https://www.erpresearch.com/en-us/ai-native-erp); [daily.dev](https://daily.dev/posts/rillet-raises-25m-from-sequoia-to-automate-general-ledger-systems-using-ai)
- $100M Series C at $1B valuation led by ICONIQ (Aug 2026), total >$200M, 600+ customers (Neuralink, Skild AI, Mercor), "agent activity growing about 70 percent month over month" — [TNW](https://thenextweb.com/news/rillet-100m-series-c-1bn-accounting-ai); [Pulse2](https://pulse2.com/rillet-raises-100-million-series-c-at-1-billion-valuation-as-ai-native-erp-tops-600-customers/); [Rillet blog](https://www.rillet.com/blog/rillet-raises-100m-series-c-at-1b-valuation-to-build-accounting-superintelligence)
- Positioning: "accounting superintelligence with AI agents doing finance work inside a real-time general ledger, with human approval and a full audit trail"; agents "Complete accounting tasks end-to-end" with "Audit trail by default" (Aura AI); "AI agents that automate journals, accruals, reconciliations, and management reports"; "continuously reconciled ledgers with immutable audit logs" — [Rillet blog](https://www.rillet.com/blog/rillet-raises-100m-series-c-at-1b-valuation-to-build-accounting-superintelligence); [ERPClaw vendor comparison](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/); [RoboCFO landscape](https://robocfo.ai/frameworks/ai-native-erp-landscape)
- Closest to Rhea: human-approval gates on every agent action plus default audit trail. Sharpest difference: AI drafts *transactions*, not rules; ledger logic is vendor-coded and there is no rules-as-data, replay, or determinism story in any source found.

**DualEntry** (full-stack, migration-wedge, mid-market multi-entity)
- Founded 2023; $90M Series A Oct 2025 (Lightspeed/Khosla); claims to automate "90% of traditional workflows"; "24-hour go-live" NetSuite-migration positioning — [ERPClaw vendor comparison](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/); [RoboCFO landscape](https://robocfo.ai/frameworks/ai-native-erp-landscape)
- Architecture claims: "Unified ledger with real-time data sync; immutable audit trails with multi-step approvals and role-based permissions"; period locks — [RoboCFO landscape](https://robocfo.ai/frameworks/ai-native-erp-landscape)
- Closest to Rhea: immutable audit trail + multi-step approvals. Sharpest difference: AI does transaction matching and flux explanations; business logic is vendor-configured for fast migration, not runtime-definable; no rule authoring, replay, or provenance claims.

**Campfire** (full-stack, seed-to-Series-B startups)
- $65M Series B Oct 2025 (Accel/Ribbit), $100M+ total — [ERPClaw vendor comparison](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/)
- Proprietary "Large Accounting Model (LAM)" claiming "95%+ accuracy" in reconciliations; "Ember AI copilot built on Claude"; ASC 606 revenue automation; "autonomous reconciliation, auto-categorization, and anomaly detection" with staged approval — [RoboCFO landscape](https://robocfo.ai/frameworks/ai-native-erp-landscape); [ERPClaw vendor comparison](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/)
- Closest to Rhea: AI executes accounting tasks with staged human approval. Sharpest difference: the "logic" is a proprietary ML model's weights — the opposite of Rhea's inspectable versioned rules; no immutability/replay claims found.

**Light (light.inc)** (Copenhagen; multi-entity/multi-country "agentic ERP")
- Raised €25M/$30M (Sept 2025) "to replace legacy finance systems with AI-native platform" — [EU-Startups](https://www.eu-startups.com/2025/09/light-raises-e25-million-to-replace-legacy-finance-systems-with-ai-native-platform/); [ArcticStartup](https://arcticstartup.com/light-raises-e25-million/)
- Single-ledger design: "entities are dimensions of one ledger, consolidation posts continuously with eliminations and FX"; "AI agents doing the bookkeeping under your approval policies"; performance claim of 280M records processed in under a second — [light.inc](https://light.inc/); [ERP Research Light review](https://www.erpresearch.com/en-us/light-erp)
- "Organic software" manifesto: "Every correction is training. When a controller recodes an invoice, adjusts an accrual, or overrides a match, the system takes that as instruction"; the system "reshapes itself around the business as it is today"; "Organic software hands them the pen. When the system takes instruction in plain language and learns from how the work is actually done, shaping it stops being a technical act" — [Light manifesto](https://light.inc/manifesto)
- The manifesto does **not** address AI-generated schemas/rules as artifacts, determinism guarantees, or ledger immutability — [Light manifesto](https://light.inc/manifesto)
- Closest to Rhea: the strongest *philosophical* match — system behavior shaped at runtime from human corrections and plain-language instruction, under approval policies. Sharpest difference: the learning is implicit (model/automation adapts from corrections) rather than explicit versioned rule artifacts; no draft/approve lifecycle for the learned behavior, no replay or provenance guarantee anywhere in their public material.

**Doss** (ops-first "Adaptive Resource Platform", not ledger-first)
- Founded 2022; $18M raised Apr 2025; $55M Series B Mar 2026 (Madrona/Premji) — [SiliconANGLE](https://siliconangle.com/2025/04/15/doss-raises-18m-accelerate-ai-driven-erp-development-global-expansion/); [MLQ](https://mlq.ai/news/doss-raises-55m-for-ai-natives-operations-cloud-targeting-erp-workflows/); [ERPClaw vendor comparison](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/)
- "Composable operations platform that lets teams build modules for each business function using three no-code building blocks: tables, forms, and workflows"; deploys "in as little as two days" — [search synthesis citing doss.com](https://www.doss.com/products/platform)
- "Internal DSL allows workflow generation"; "immutable operations log with separation-of-duties via rules engine"; AI "auto-generates custom workflows", demand-triggered POs, approval workflows for purchase orders — [RoboCFO landscape](https://robocfo.ai/frameworks/ai-native-erp-landscape)
- Operates "around your existing general ledger... plugs into existing ERPs rather than replacing them"; "self-configuring setup" but no custom rule authoring by AI mentioned — [ERPClaw vendor comparison](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/)
- Closest to Rhea: runtime-definable modules as data (tables/forms/workflows DSL) plus an "immutable operations log" claim. Sharpest difference: the humans configure via no-code (AI assists); there is no AI-authored, versioned, approval-gated rule artifact, no accounting kernel, and no replay/determinism claim.

**Keel (keel.so)** (London; ops layer between no-code and ERP)
- Founded 2022 by Benoit Machefer, Tom Frew, Jon Bretman (ex-Echo pharmacy); $6M seed Oct 2024 (Earlybird, LocalGlobe) — [Tech.eu](https://tech.eu/2024/10/23/keel-raises-6m-to-bridge-the-gap-between-no-code-and-erp-solutions/); [Vestbee](https://www.vestbee.com/insights/articles/keel-secures-6-m)
- Explicitly "not accounting software — Xero and QuickBooks handle that. Keel replaces the operational layer: orders, inventory, production, and fulfilment"; open-source, API-first backend engine with auto-generated admin tools — [TechFundingNews](https://techfundingnews.com/keel-lands-6m-to-power-the-next-generation-of-operational-software/)
- "AI-assisted workflow generation from business requirements; no fixed schema — teams define custom data models"; "software dev best practices including version control and sandbox testing" — [RoboCFO landscape](https://robocfo.ai/frameworks/ai-native-erp-landscape)
- Closest to Rhea: runtime-definable schemas and workflows with versioning and sandboxing discipline. Sharpest difference: the artifacts are code/config authored by ops teams (AI assists at generation time); no event log, no replay, no provenance, no accounting domain.

**Nominal** (augmentation layer over existing ERPs)
- Founded 2023; ~$30M, Series A Jul 2025; "Shadow GL mirrors official books"; "read-only mode preserves audit trails"; "detailed audit trails for every AI action"; SOC 1 Type I — [RoboCFO landscape](https://robocfo.ai/frameworks/ai-native-erp-landscape)
- "Operates as an AI layer on top of established ERP systems" rather than replacing them — [search synthesis](https://www.erpresearch.com/en-us/ai-native-erp)
- Closest to Rhea: every AI action individually audit-trailed, and the shadow-GL is structurally a derived projection over source systems (loosely analogous to Rhea's rebuildable `object` cache). Sharpest difference: it sits on *mutable* incumbent ERPs, the AI emits journal entries not rules, and there is no rules-as-data or replay concept.

**Everest Systems → Lensing** (top-of-market, ex-SAP-HANA founders) — **closest commercial analog on the rule-authoring axis**
- Founded 2020 by Franz Faerber (26-year SAP veteran, "key architect of the SAP HANA in-memory database"), Sandeep Chopra (Deloitte/Veeva), CTO Joachim Fitzer; out of stealth Nov 2024 with $140M from Sutter Hill, Altimeter, Redpoint, D1 — [Everest co-founders post](https://www.everest-systems.com/post/the-co-founders-vision-behind-everest-erp-built-for-saas-businesses); [Dealroom](https://dealroom.co/companies/everest-systems/); rebranded to Lensing (everest-systems.com 301-redirects to lensing.ai) — [lensing.ai](https://lensing.ai/)
- AiSpecify is "a full software development lifecycle platform" where "business users gain the power to build entire applications, specification to launch"; "Build agents and apps in plain English, no custom language or waiting on IT" — [lensing.ai](https://lensing.ai/)
- "Live Sandbox to simulate any changes before approving them"; tagline "This is enterprise-native AI, the exact opposite of vibe-coding"; multi-entity, multibook system of record with intercompany eliminations — [lensing.ai](https://lensing.ai/)
- "AiSpecify describes business logic in plain language for automatic workflow configuration... Supports SOX change management control via sandbox testing" — [RoboCFO landscape](https://robocfo.ai/frameworks/ai-native-erp-landscape)
- Lensing's public pages "don't specify detailed versioning, audit trails, or replay mechanisms for configurations" — [lensing.ai](https://lensing.ai/)
- Closest to Rhea: the only funded vendor whose core pitch is AI turning plain-language *specifications of business logic* into running configuration, simulated in a sandbox and human-approved before deployment — structurally parallel to Rhea's draft → simulate → approve → active. Sharpest difference: the artifact is application/workflow configuration on a metadata platform (possibly generated code), not strict-JSON rules executed by a deterministic kernel over an append-only event log; no determinism, replay, or per-field provenance claims; no public statement that only approved versions execute as an enforced invariant.
- Note: the research brief's "Tactiq/Everest" — no evidence found that "Tactiq" is an ERP product; Everest's actual rebrand is **Lensing** (see Gaps).

**ERPClaw (AvanSaber)** — open source, not a funded startup, but the closest *architectural* match found
- "Every business operation, from `add-customer` to `submit-sales-invoice` to `stripe-post-payout-gl`, is a discrete action an AI agent can call"; "Posted entries are never edited: a correction is a reversal plus a new entry, and the general ledger is hash-chained so an integrity check catches tampering"; "High-impact actions refuse to run without explicit `--user-confirmed` flag"; "The optional ERPClaw OS engine generates new modules in a sandbox, checks them against the system's accounting rules, and deploys only when you approve"; GPL v3, no venture funding — [ERPClaw vendor comparison (self-description)](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/)
- Caveat: erpclaw.ai both sells this architecture and publishes the comparison content cited here, so treat its self-assessment as marketing.
- Closest to Rhea: append-only hash-chained ledger, validated write-time posting engine, and AI-generated extensions that deploy only after automated rule-checking plus explicit human approval. Sharpest difference: the AI generates *modules* (code), not rules-as-data; no claim that wiping state and replaying reproduces identical state, and no per-field provenance triple.

**Digits** (SMB, "Autonomous General Ledger")
- AGL announced Mar 2025; "dozens of specialized AI models and agents... trained in-house on a proprietary dataset of over $825 billion in small-business transactions"; "specialized, non-generative predictive AI models for the core tasks of classification and reconciliation" (explicitly to avoid LLM hallucination); auto-books "up to 95% of entries"; agents "pausing only when human judgment is necessary" — [Accounting Today](https://www.accountingtoday.com/news/digits-announces-autonomous-general-ledger); [CPA Practice Advisor](https://www.cpapracticeadvisor.com/2025/03/11/digits-autonomous-general-ledger-will-take-on-quickbooks-company-says/157183/); [Yahoo Finance](https://finance.yahoo.com/news/digits-launches-first-ai-agents-140000473.html)
- Closest to Rhea: deliberate rejection of generative models for core posting in favor of deterministic-leaning predictive models, with human pause points. Sharpest difference: the "logic" lives in trained model weights — unexplainable and unversionable in Rhea's sense; no rules, no replay, no provenance artifacts.

**Category context**
- "Venture investors have put approaching half a billion dollars into the category in just over a year"; challenger list includes Rillet, Pennylane, Ramp, Light, DualEntry, Campfire, Everest, Digits — [ERP Research](https://www.erpresearch.com/en-us/ai-native-erp); [CIO.com](https://www.cio.com/article/4225287/6-erp-trends-for-2026-and-beyond-cios-rethink-core-systems.html)

### Inferences
- The category's consensus pattern is "AI drafts transactions, humans approve, immutable audit trail" — i.e., Rhea's invariant 2 applied to *transactions*. Nobody applies it to *rules*: the rule layer everywhere remains vendor code (Rillet/DualEntry/Campfire/Light) or human no-code config (Doss/Keel), with Lensing's AiSpecify the lone "AI authors the logic" pitch.
- "Immutable audit trail" is table stakes marketing in this category, but in every case it means a log *about* mutable state, not an event log *from which* state is derived. That inversion (state as a pure function of log + rules) is the part of Rhea's architecture with no commercial occupant.
- Two unfunded/open-source projects (ERPClaw, Keel) are closer to Rhea's architecture than any of the funded unicorn-track vendors, suggesting the funded players compete on automation outcomes (close speed, headcount) rather than on architectural guarantees.

### Gaps
- "Tactiq" as an ERP/accounting company: no evidence found; the only widely-known Tactiq is a meeting-transcription tool. Everest Systems' rebrand is Lensing. The name may be a conflation; not searched further within call budget.
- Campfire's founding team/date and Nominal's exact round size conflict across sources (RoboCFO says Campfire founded 2022 / Nominal ~$30M; other coverage differs); treat funding figures other than press-release-sourced ones (Rillet, Doss, Keel, Everest) as approximate.
- DualEntry's and Rillet's internal data models (mutable Postgres vs anything event-sourced) are not publicly documented; the "immutable audit trail" claims could not be verified below the marketing layer.

## Key question 2: Does any vendor use event sourcing, deterministic replay, or per-field provenance as an architectural guarantee?

### Takeaway
No. Audit trails and "immutability" claims are universal; event sourcing as the system of record, wipe-and-replay determinism, and per-field provenance triples appear in no commercial AI-native ERP's public material. Event-sourced ledgers exist only as an engineering pattern in open-source demos and architecture literature.

### Cited Findings
- ERPClaw is the only product found enforcing append-only at write time: "the general ledger is hash-chained so an integrity check catches tampering"; corrections are reversal-plus-new-entry — [ERPClaw](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/). Even it makes no replay-reproduces-state or per-field provenance claim.
- Rillet: "immutable audit logs" / "audit trail by default", but "model specifics absent" and determinism "not addressed" — [RoboCFO](https://robocfo.ai/frameworks/ai-native-erp-landscape); [ERPClaw](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/)
- DualEntry: "immutable audit trails with multi-step approvals"; data architecture otherwise undocumented — [RoboCFO](https://robocfo.ai/frameworks/ai-native-erp-landscape)
- Nominal: "detailed audit trails for every AI action", read-only shadow GL — per-action provenance, not per-field — [RoboCFO](https://robocfo.ai/frameworks/ai-native-erp-landscape)
- Doss: "immutable operations log with separation-of-duties via rules engine" — [RoboCFO](https://robocfo.ai/frameworks/ai-native-erp-landscape)
- Light's manifesto wants "text and numbers" connected so "the system can explain itself," but "avoids technical specifics about audit trails or schema governance" and says nothing about determinism or immutability — [Light manifesto](https://light.inc/manifesto)
- Event sourcing with replay ("rebuild by replaying the events... safely deleted and completely rebuilt by replaying the event stream") is well-established as a *pattern*, with accounting/audit cited as the canonical use case — but the search surfaced only open-source demos, blog posts, and a bachelor thesis, no commercial ERP vendor — [Baytech](https://www.baytechconsulting.com/blog/event-sourcing-explained-2025); [example repo](https://github.com/renanbambam/event-sourcing-ledger); [HAW Hamburg thesis](https://reposit.haw-hamburg.de/bitstream/20.500.12738/16034/1/BA_Evaluation_Use_of_Event-Sourcing.pdf)

### Inferences
- The category conflates three distinct guarantees Rhea separates: an audit *log* (everyone claims it), an append-only *source of record* (only ERPClaw), and *derivability* — state as a replayable function of log + versioned rules (nobody). Rhea's invariant 4 (replay reproduces identical state) and invariant 5 (per-field (event_id, rule_id, rule_version)) have no commercial precedent found in this survey.
- Because every funded vendor's AI acts directly on ledger state with an audit trail, their "explainability" is narrative (the agent logs what it did) rather than structural (the state itself carries its derivation). This is the sharpest architectural differentiation available to Rhea.

### Gaps
- Absence of evidence caveat: engineering blogs for Rillet, DualEntry, Campfire, and Light were not individually crawled (call budget); one of them could use event sourcing internally without marketing it. No public claim exists, which is itself the finding.
- Fintech infrastructure plays known to use immutable/double-entry ledger APIs (e.g., Modern Treasury, TigerBeetle, Formance) were out of scope (not ERPs, no AI rule story) and were not surveyed; worth a follow-up if the question is "who sells an immutable ledger" rather than "who sells an ERP."

## Key question 3: Are there 2025–2026 entrants explicitly claiming "AI writes the ERP's logic" or "self-configuring / self-learning ERP"?

### Takeaway
The phrases exist, but almost always mean less than they say: "self-configuring" mostly refers to AI-assisted setup/configuration (Doss, Opkey test scripts) and "self-learning" to models adapting from corrections (Light). Lensing's AiSpecify is the only funded product claiming business users + AI author whole applications from plain-language specs with sandbox-approval, and Light's manifesto is the only one claiming the system *rewrites its own behavior* from corrections.

### Cited Findings
- Lensing AiSpecify: "business users gain the power to build entire applications, specification to launch... Build agents and apps in plain English"; "Live Sandbox to simulate any changes before approving them" — [lensing.ai](https://lensing.ai/)
- Light: "Every correction is training... the system takes that as instruction"; it "reshapes itself around the business as it is today, not as it was on the day of the kickoff meeting" — [Light manifesto](https://light.inc/manifesto)
- Doss ARP "uses AI agents to dynamically configure workflows, build integrations, and optimise processes... without heavy manual setup" — [Versori](https://www.versori.com/post/5-ai-native-erps-reimagining-enterprise-software); ERPClaw's assessment: Doss has "self-configuring setup but no custom rule authoring mentioned" — [ERPClaw](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/)
- Opkey (ERP-lifecycle, not an ERP): "agentic AI-powered ERP Lifecycle Optimization Platform... AI-powered, self-configuring test scripts", claims up to 50% cost and 85% testing-time reduction — [CIO.com](https://www.cio.com/article/3833003/startup-opkey-launches-agentic-ai-platform-for-erp-lifecycle-optimization.html)
- "Lensing: 'No-code agent building and AI agents across the platform' suggests users can define business logic"; classified as the only vendor with a "builder model" execution position — [ERPClaw](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/)
- McKinsey frames the disruption as AI-driven but process-level ("five ways AI is disrupting ERP"), not logic-authoring — [McKinsey](https://www.mckinsey.com/capabilities/mckinsey-technology/our-insights/the-end-of-erp-as-we-know-it-five-ways-ai-is-disrupting-erp)
- Adjacent composable/headless players surfaced but not verified in depth: tailor.tech (headless ERP content marketing) — [Tailor](https://www.tailor.tech/resources/posts/ai-in-erp-how-artificial-intelligence-is-shaping-the-future-of-erp-software); erp.ai ("Business Systems & Headless SaaS") — [erp.ai](https://www.erp.ai/)

### Inferences
- "AI writes the logic" exists in two commercial flavors, both different from Rhea's: (a) generation-time authoring of apps/config from specs (Lensing, Keel, Doss) where the output is conventional config/code thereafter, and (b) continuous implicit learning (Light, Digits) where the "logic" is never an inspectable artifact at all. Rhea's flavor — AI emits explicit, versioned, schema-validated rule *data* that a kernel executes — sits between these and is claimed by no one found.
- The "human approval of AI-authored logic" step exists commercially only as sandbox/change-management practice (Lensing's Live Sandbox, Keel's sandbox testing, ERPClaw's deploy-on-approval), not as a first-class recorded lifecycle (draft → approved → active → superseded) enforced by the execution engine.

### Gaps
- Y Combinator directory was not directly queried (call budget); 2025–2026 YC batches may contain pre-seed "AI authors the ERP" companies below press visibility. The analyst roundups (ERP Research, RoboCFO, ERPClaw, Soberan, CodeBrewTools) cross-covered the same ~10 names, suggesting coverage saturation at the funded tier.
- No Hacker News thread-level evidence gathered on architecture discussions of these vendors.

## Key question 4: What do incumbents' AI layers (SAP Joule, Oracle/NetSuite, Microsoft Dynamics, Odoo) actually do — assistance or logic-authoring?

### Takeaway
Incumbent AI is agent-orchestration over existing vendor-coded transactions. SAP's Joule Studio is the most logic-authoring-adjacent — a GA low/no-code builder for custom agents and skills — but humans author the agents and the agents invoke pre-existing ERP logic; NetSuite Next adds approval-gated agentic workflows on top of a forms-based core. Nothing resembling AI-authored executable business rules with a versioned lifecycle was found.

### Cited Findings
- SAP Joule Studio agent builder went GA in Q1 2026: "enterprises define goals, permissions, triggers, and workflows using low-code and no-code capabilities" — [AIMultiple](https://aimultiple.com/sap-ai-agents); [SAP Community](https://community.sap.com/t5/technology-blog-posts-by-members/sap-joule-studio-an-overview-of-intent-based-ai-agent-development/ba-p/14496888)
- At Sapphire 2026 SAP unveiled Joule Studio 2.0: "a fully managed, zero-infrastructure agent builder — with 12 months of free design-time access, a €100 million partner fund, Cursor IDE support" — [Savic Tech](https://www.savictech.com/insights/sap-joule-studio-2-managed-agent-builder-citizen-developer-2026/); [SAP Sapphire guide](https://www.sap.com/topics/events/sapphire/innovation-news-guide-2026)
- NetSuite Next: unveiled at SuiteWorld Oct 2025, planned for late 2026, rollout began 2026.2 (US/Canada); "agentic workflows for payment proposals, vendor selection, and reconciliations can be approved step by step or allowed to act on their own. Still forms-based underneath"; "Ask Oracle" natural-language assistant across a unified data model — [ERPClaw](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/); [HouseBlend](https://www.houseblend.io/articles/agentic-erp-sap-sapphire-2026-netsuite-next)
- NetSuite Next's per-step-approval-or-autonomy toggle is the incumbents' only human-in-the-loop granularity claim found — [ERPClaw](https://www.erpclaw.ai/blog/5-ai-native-erps-that-earn-the-label/)

### Inferences
- Joule Studio's output (an agent with goals/permissions/triggers) is the incumbents' nearest thing to runtime-definable logic, but it is (a) human-authored with AI assistance, inverted from Rhea where AI authors and humans approve, and (b) orchestration *over* the ERP's fixed transactional logic, not rules that *produce* state. The deterministic core stays vendor-coded in every incumbent.
- Incumbent messaging has converged on "agents with optional autonomy + approval," i.e., the same transaction-level HITL pattern as the startups — which further isolates rule-level HITL as unclaimed territory.

### Gaps
- Microsoft Dynamics 365 Copilot/agents and Odoo's AI features returned no usable results within the call budget; no sourced assessment is included for them. Known public positioning (Copilot Studio agents, Odoo AI automation) was deliberately omitted as unsourced in this survey — a follow-up single search each would close this.
- No vendor was found shipping statutory content (CoA, VAT, e-invoicing/KSeF-style) as swappable *data packs* (Rhea point 6); incumbents ship localization as vendor code/config modules. This claim is an inference from absence — no source directly addresses the packaging question, so treat point 6 as unverified rather than confirmed unoccupied.
