# Rules-as-Data Engines, Decision Management, and Policy-as-Data — Survey vs. Rhea's Rule Layer (as of October 2026)

Rhea's rule layer, for comparison: (1) LLM authors rules as versioned strict-JSON *data*, never code; (2) draft → human-approved → active → superseded lifecycle, only approved+active versions execute; (3) deterministic kernel executes; (4) rule versions + append-only event log give reproducible replay and per-field provenance.

## Key Question 1 — Decision management platforms: do any offer LLM-drafted rules with approval lifecycles?

### Takeaway
Yes, and 2025–2026 is the inflection point: GoRules now ships an AI copilot that drafts decision graphs as JSON data with diff-review, auto-generated tests, and human-gated versioned release — the closest commercial match to Rhea's loop found anywhere. IBM ODM gained an AI toolkit ("Bob mode") that generates whole rule projects from policy text, and DecisionRules added a Gemini-backed assistant that builds decision tables from plain language. None pair this with event sourcing or per-field provenance.

### Cited Findings

**GoRules / ZEN Engine (closest commercial match)**
- ZEN Engine is a cross-platform open-source business rules engine written in Rust with bindings for Node.js, Python, Go, Java, Kotlin, .NET; decisions are stored as portable JSON ("JSON Decision Model", JDM) — rules are literally data files, loaded and executed by the engine — [gorules/zen on GitHub](https://github.com/gorules/zen); [GoRules architecture docs](https://docs.gorules.io/developers/overview/architecture)
- The GoRules BRMS layer provides semantic-versioned releases, environment-scoped deployments with rollback, branching for parallel rule development, and "publish, release, and deploy" as explicit human actions with an audit trail; rules can also be exported to Git and bundled via CI/CD — [GoRules architecture docs](https://docs.gorules.io/developers/overview/architecture); [Platform overview — "From Policy to Production"](https://gorules.io/overview)
- GoRules AI (marketed as "AI Business Rules Engine"): describe a policy in natural language, or drop in a policy PDF or spreadsheet, and the AI builds or modifies decision tables, expressions, and entire decision graphs; the assistant's proposal is shown **as a diff on the canvas which the user accepts or rejects** — [GoRules AI page](https://gorules.io/ai); [gorules.io](https://gorules.io/)
- GoRules AI "generates scenario suites for every change and runs them before anything reaches a reviewer," and the stated division of labor is: "AI drafts, tests, and explains, while releasing stays with your reviewers — the same approvals and audit trail as any human change" — [GoRules AI page](https://gorules.io/ai)
- GoRules positions itself for governing AI systems with deterministic policy ("AI Governance Software That Enforces Policy") — [gorules.io/use-cases/ai-governance](https://gorules.io/use-cases/ai-governance)
- *Closest resemblance to Rhea:* the full authoring loop — LLM drafts rule-as-JSON-data → human reviews a diff and accepts → versioned, audited release → deterministic engine executes. This is essentially Rhea invariants 2 and 3 as a product.
- *Sharpest difference:* ZEN/GoRules is a stateless request/response decision evaluator. There is no append-only business event log, no replay that reproduces object/document state through historical rule versions, and no per-derived-field provenance `(event_id, rule_id, rule_version)` — the audit trail covers *rule changes*, not *derived facts*. GoRules evaluates decisions; Rhea's kernel *builds state* from them.

**IBM ODM — "Bob mode" / ODM AI Toolkit (announced June 2026)**
- The ODM AI Toolkit ("ODM Rule Designer Bob mode") generates a complete, production-ready ODM rules project from a single natural-language prompt or a policy document: the XOM (Java domain model), BOM with natural-language vocabulary, business rules in Business Action Language (BAL), ruleflow orchestration, and deployment configuration — [IBM Community blog, Mathias Mouly, 2026-06-08](https://community.ibm.com/community/user/blogs/mathias-mouly1/2026/06/08/odm-rule-designer-bob-mode)
- "Every project is validated with the ODM Build Command and imports directly into Rule Designer — no manual cleanup required"; the toolkit aims to "remove the repetitive scaffolding so teams can focus on the decisions that matter" — [IBM Community blog](https://community.ibm.com/community/user/blogs/mathias-mouly1/2026/06/08/odm-rule-designer-bob-mode)
- A built-in GenAI assistant generating decision tables/BAL rules from prompts is *not* part of the standard ODM feature set; IBM's GenAI capabilities historically lived in separate products — [Nected ODM overview](https://www.nected.ai/tools/ibm-odm-overview)
- A Medium article by Pierre Feillet (IBM) on generative AI for business automation describes LLMs extracting business rules from plain-text policies, which "once reviewed by a human, can automate decisions with traceability and determinism" — [Medium, Pierre Feillet](https://medium.com/@pierrefeillet/approaches-in-using-generative-ai-for-business-automation-the-path-to-comprehensive-decision-3dd91c57e38f) (note: direct fetch returned HTTP 403; this characterization comes from the search-result summary and should be treated as secondhand)
- *Closest resemblance:* AI authors complete rule content from policy documents; ODM's Decision Center product line has long had enterprise rule-lifecycle governance.
- *Sharpest difference:* the generated artifact is partly **code** (a Java XOM) plus BAL (a controlled natural language), not pure validated data; generation targets a developer IDE (Rule Designer) rather than landing as a *draft state* inside a governed store; the blog post does not describe any approval gate between generation and deployment, and there is no event-sourced replay or field provenance.

**DecisionRules.io AI Assistant**
- The AI Assistant generates Decision Tables, Scripting Rules, Lookup Tables, and Decision Flows from natural-language descriptions, building "the appropriate rule or flow directly in your active workspace"; the output is rule configuration data in DecisionRules' proprietary format, not standalone source code; it runs on "Google Gemini under enterprise privacy terms" with payloads that "never train public models"; marketing says "your team stays in control of every decision that goes live" — [DecisionRules article: AI for Business Rules](https://www.decisionrules.io/en/articles/ai-for-business-rules-from-plain-language-to-working-logic/)
- The article does **not** specify a human approval step, draft/published lifecycle, version control, audit logging, or determinism guarantees for AI-generated rules — [same source](https://www.decisionrules.io/en/articles/ai-for-business-rules-from-plain-language-to-working-logic/)
- *Closest resemblance:* LLM → rule-as-data draft inside a rules-engine workspace, with "control of what goes live" rhetoric.
- *Sharpest difference:* no documented approval lifecycle tied to the AI output (governance unspecified in published material), and no event log/replay/provenance story.

**Camunda (8.8 era, 2025–2026)**
- Camunda Modeler ships AI copilots for FEEL and for BPMN; Camunda Copilot provides modeling suggestions, generates BPMN diagrams from natural-language prompts, creates documentation, and includes a form builder; FEEL is used inside DMN tables, gateways, connectors, and forms — [Camunda blog: The Power of Camunda Copilots (Aug 2025)](https://camunda.com/blog/2025/08/the-power-of-camunda-copilots/)
- With Camunda 8.8 you "choose between modeling deterministic logic using DMN decision tables, or invoking AI agents via connectors that integrate with external LLM APIs" — Camunda explicitly frames **deterministic DMN vs. AI-agent-at-runtime** as the design choice — [Camunda blog: AI agent or rule-based DMN? (Jul 2025)](https://camunda.com/blog/2025/07/ai-agent-or-based-rule-dmn-ai-powered-orchestration/)
- *Closest resemblance:* DMN decision tables are declarative data (XML) executed by a deterministic engine, and an AI copilot helps author FEEL expressions inside them.
- *Sharpest difference:* the copilot assists a human modeler interactively rather than producing drafts that enter an approval lifecycle; and Camunda's headline AI story is the *inverse* of Rhea's — put the LLM at **runtime** (agent connectors) versus Rhea's LLM at **authoring time** with a deterministic runtime.

**Decisions.com**
- Decisions describes its rules engine as "the deterministic core behind intelligent decisioning," which "governs when AI or agents can act," within a low-code platform for building, testing, and governing rules — [Decisions rules engine page](https://decisions.com/platform/rules-engine); [Decisions blog on AI-powered rules engine](https://decisions.com/redefining-intelligent-decisioning-with-a-low-code-ai-powered-rules-engine/)
- *Closest resemblance:* deterministic rules governing AI actions (rules as the guardrail layer).
- *Sharpest difference:* the direction is rules-govern-AI, not AI-authors-rules; no published LLM-drafting-with-approval loop was found.

### Inferences
- The decision-management industry converged in 2025–2026 on "AI drafts, human approves, deterministic engine executes" as a marketing and product pattern (GoRules most completely, IBM partially, DecisionRules nominally) — Rhea's authoring loop is no longer unique as a *loop*; what remains unique is what sits downstream of it (see KQ5).
- GoRules' diff-review-accept on canvas is functionally equivalent to Rhea's draft→approved transition, though GoRules does not appear to model it as a first-class state machine on rule versions with the approval itself recorded as an event.

### Gaps
- Drools/Kogito, OpenRules, FICO (Blaze/Platform), and Red Hat decision tooling: not individually investigated within the call budget; no evidence either way on LLM authoring found in the searches run.
- Whether DecisionRules' AI output lands as a formally distinct "draft" version state (its docs at docs.decisionrules.io were not fetched).
- Whether IBM Decision Center imposes an approval gate on Bob-generated projects (the announcement blog is silent on governance integration).

## Key Question 2 — Policy-as-data (OPA, Oso, Cedar, Kyverno): any LLM-authoring-with-approval story?

### Takeaway
Cedar is the standout: its evaluator is formally deterministic and formally analyzable, and 2026 research (AutoCedar; autoformalization pipelines) uses LLM generator-critic loops to synthesize Cedar policies with *machine* verification standing in where Rhea puts a *human* approver. No equivalent LLM-authoring story surfaced for OPA, Oso, or Kyverno.

### Cited Findings
- Cedar's authorizer is deterministic: "guaranteed to terminate and always produce the same authorization decision for a given request, hierarchy, and set of policies"; AWS chose Cedar for securing agentic workflows in Bedrock AgentCore, where policies deterministically reject unexpected LLM-produced tool arguments — [AWS Security Blog: Why Policy in Amazon Bedrock AgentCore chose Cedar](https://aws.amazon.com/blogs/security/why-policy-in-amazon-bedrock-agentcore-chose-cedar-for-securing-agentic-workflows/)
- Cedar was built with a verification-guided approach and has a symbolic compiler translating policies to SMT-lib, "sound, complete, and decidable," enabling proofs that two policy sets authorize exactly the same requests — [Cedar language paper (Amazon Science)](https://assets.amazon.science/96/a8/1b427993481cbdf0ef2c8ca6db85/cedar-a-new-language-for-expressive-fast-safe-and-analyzable-authorization.pdf); [How We Built Cedar (arXiv 2407.01688)](https://arxiv.org/pdf/2407.01688); [AWS: Prove It, Part 2 — formal logic, Cedar policies](https://aws.amazon.com/aws-startups/learn/prove-it-part-2-formal-logic-cedar-policies-and-the-economics-of-verification/)
- "AutoCedar: An Agentic Framework for Verifier-Guided Access Control Policy Synthesis" (2026) — an agentic framework where LLMs synthesize Cedar access-control policies under verifier guidance — [arXiv 2607.03656](https://arxiv.org/html/2607.03656v1)
- "Autoformalization of Agent Instructions into Policy-as-Code" (2026) describes an LLM generator-critic pipeline translating agent prompts, MCP tool descriptions, and natural-language policy documents into formally verified policies: a **hard critic** does deterministic syntax/schema/contradiction checking and a **soft critic** (LLM-as-judge) scores semantic alignment with the original instructions — [arXiv 2606.26649](https://arxiv.org/html/2606.26649)
- *Closest resemblance (Cedar research stack):* LLM drafts a declarative, data-like policy; a deterministic verifier gates acceptance; the runtime is formally deterministic. This mirrors Rhea's invariant 3 (model output is data, validated before storage) with formal methods replacing the human approver.
- *Sharpest difference:* domain is authorization allow/deny only — policies constrain actions, they don't *derive state*; there is no versioned activation lifecycle with human approval events, no event sourcing, and no provenance on derived facts (there are no derived facts).

### Inferences
- The Cedar line of work validates Rhea's core bet from the formal-methods side: constrained declarative languages are what make LLM authoring safe, because drafts can be checked mechanically before activation. Rhea's JSON-schema validation is a lightweight version of Cedar's hard critic.
- A hybrid is conceivable and unoccupied: verifier-guided *plus* human-approval-as-recorded-event. Neither the Cedar papers nor any product found combines both gates.

### Gaps
- OPA/Rego: no LLM-authoring-with-approval product offering found in these searches (not deeply searched; Styra's product line not investigated).
- Oso and Kyverno: not individually investigated within the call budget.

## Key Question 3 — Tax/billing/pricing engines with rules-as-data: any AI rule authoring?

### Takeaway
Tax engines moved first among vertical rule systems: Avalara (Sept–Oct 2025) lets users create and edit compliance rules via AI prompts under its ALFA agentic framework, and Fonoa pairs AI regulatory monitoring with "your team sets the rules" human oversight over deterministic tax logic. Neither publishes a formal draft→approve→active version lifecycle for AI-drafted rules, and billing/pricing vendors (Zuora, Metronome, Orb, Stigg) went uninvestigated.

### Cited Findings
- Avalara introduced agentic AI for tax and compliance (Sept 2025): users can "create, edit, and visualize complex compliance rules using simple AI prompts with no coding," powered by ALFA (Avalara LLM Framework for Agentic Applications), which combines large and small language models; AI-guided onboarding prefils profiles, recommends registrations, and configures rules automatically — [Avalara blog (Sept 2025)](https://www.avalara.com/blog/en/north-america/2025/09/avalara-introduces-agentic-ai.html); [Digital Commerce 360 (Oct 2025)](https://www.digitalcommerce360.com/2025/10/08/avalara-introduces-ai-agents-to-automate-ecommerce-tax-and-compliance/); [Accounting Today](https://www.accountingtoday.com/news/avalara-touts-agentic-ai-workforce-for-tax-and-compliance); [Avalara agentic AI product page](https://www.avalara.com/us/en/products/ai-compliance.html)
- Fonoa Agents are "an agentic execution layer for indirect tax" running the compliance lifecycle "with human supervision built in"; positioning: "your team sets the rules while AI handles the rest" — [Fonoa Agents](https://www.fonoa.com/platform/agents)
- Fonoa Knowledge (AI regulatory intelligence) continuously monitors global indirect-tax changes, maps them to the company's profile, and delivers "cited, actionable insights" — i.e., AI *proposes*, humans *decide*, the deterministic tax engine *executes* — [Introducing Fonoa Knowledge](https://www.fonoa.com/resources/blog/introducing-fonoa-knowledge); [Fonoa: Where AI Belongs in Your Tax Stack — And Where It Doesn't](https://www.fonoa.com/resources/blog/where-ai-belongs-in-your-tax-stack)
- *Closest resemblance (Fonoa):* explicit architectural separation — AI in the advisory/monitoring layer, determinism in the calculation layer, human in between. This is Rhea's three-part split applied to one statutory domain (and a direct conceptual cousin of Rhea's `internal/adapter/` statutory adapters).
- *Sharpest difference:* the rule content is substantially vendor-maintained rather than tenant-authored; no public draft→approved→active→superseded version lifecycle for AI-drafted rules; no user-facing append-only event log with replay or per-field provenance.

### Inferences
- Avalara's "configure rules automatically" during AI onboarding suggests AI writes live configuration *without* a formalized approval state machine — the opposite of Rhea's invariant 2 — though public materials are too thin to say definitively what review exists.

### Gaps
- Zuora, Metronome, Orb, Stigg: no searches run against these vendors within the call budget; whether any offers AI authoring of billing/pricing rules is unanswered.
- Avalara's internal review/audit mechanics for AI-created rules: not documented in the public announcements found.

## Key Question 4 — Low-code/no-code platforms: when AI builds the workflow, is the artifact data or code? Approval gate? Deterministic/versioned?

### Takeaway
Split picture: n8n's AI builder emits workflow JSON (data) but with no formal approval lifecycle — the community itself flags "who reviews the workflow?"; Windmill's artifacts are literal code (TS/Python/Go/Bash) governed by Git. Microsoft Copilot Studio has a real admin-approval workflow, but it gates *distribution* — logic changes can be republished without re-approval, the inverse of Rhea's invariant.

### Cited Findings
- n8n workflows are expressed and exported as JSON; AI builders (first- and third-party) generate n8n workflow/node JSON from natural-language descriptions, which users import, after which the workflow is "ready to be activated" — [n8n-Workflow-Builder-Ai (GitHub)](https://github.com/farhansrambiyan/n8n-Workflow-Builder-Ai); [Inside n8n's AI Workflow Builder architecture deep-dive](https://medium.com/@rajveer.rathod1301/inside-n8ns-ai-workflow-builder-a-complete-architecture-deep-dive-f2eeb2d57ec8)
- The review gap is an open community concern: "n8n: When AI Writes the Workflow, Who Reviews the Workflow?" and "Workflow JSON Is Generated Code" argue AI-generated workflow JSON needs code-style review that the platform doesn't impose — [dev.to](https://dev.to/hosseinhezami/n8n-when-ai-writes-the-workflow-who-reviews-the-workflow-g22); [dev.to](https://dev.to/romiteld/workflow-json-is-generated-code-1nfk)
- n8n's JSON export is version-controllable but "the round-trip is workable rather than designed-in"; Windmill by contrast is code-first (TypeScript, Python, Go, Bash scripts become runnable, schedulable, observable artifacts) with Git as first-class source of truth, so review and CI are native — [automationatlas.io n8n vs Windmill 2026](https://automationatlas.io/guides/n8n-vs-windmill-2026-comparison/); [stacksandflows.com comparison](https://stacksandflows.com/compare/windmill-vs-n8n/)
- Microsoft Copilot Studio: makers "submit for admin approval"; admins review and approve agents in the Microsoft 365 admin center before org-wide availability; **but** updating agent *details* (name, icon, description) requires new approval while "content-only changes (topics, logic, KB, etc.) can be republished without re-approval" — [Microsoft Learn: Agent Store](https://learn.microsoft.com/en-us/microsoft-365/copilot/copilot-agent-store); [spknowledge.com (Sept 2026)](https://spknowledge.com/2026/09/21/microsoft-copilot-studio-agents-knowledge-tools-mcp-publishing/)
- *Closest resemblance (n8n):* AI emits a declarative JSON artifact that a human then chooses to activate.
- *Closest resemblance (Copilot Studio):* an explicit submit→review→approved state machine with named approver roles.
- *Sharpest differences:* n8n has no approval lifecycle or rule-version state model and workflow execution hits live external systems (not deterministic/replayable); Windmill's artifact is code, which Rhea's invariant 3 forbids; Copilot Studio's approval gates *who can find the agent*, not *which logic version executes* — logic re-publication bypasses the human gate entirely, and the runtime (LLM topics/generative answers) is non-deterministic.

### Inferences
- Across low-code platforms, "activation" is a deployment act, not a governed state transition on a versioned logic artifact; none of the platforms found treats the approval itself as a recorded, replayable event.
- Airplane.dev (named in the assignment) was discontinued after its team's acquisition by Airtable in early 2024 — background knowledge, not verified by a source in this research; treat accordingly.

### Gaps
- Zapier, Workato, Retool, ServiceNow Creator: not individually searched within the call budget; their 2025–2026 AI-builder governance specifics are unanswered.
- Whether n8n cloud's first-party AI Workflow Builder adds any review gate beyond canvas preview: not confirmed from primary docs.

## Key Question 5 — Does any product combine the full Rhea loop (LLM drafts declarative rule → human approves → versioned activation → deterministic execution with full provenance)?

### Takeaway
No product found implements the full loop. GoRules comes closest on the authoring/governance half (LLM-drafted JSON rules, diff review, human-gated versioned release, deterministic engine, audit trail) but stops at decision evaluation — it has no append-only business event log, no replay that reproduces object state through historical rule versions, and no per-field provenance on derived facts. The event-sourcing half of Rhea exists as a well-understood pattern and in 2026 agent-systems research, but nobody found has fused the two halves.

### Cited Findings
- GoRules' published loop — natural language/PDF in, decision tables and graphs out, proposal reviewed as an accept/reject diff, AI-generated scenario tests run "before anything reaches a reviewer," and "releasing stays with your reviewers — the same approvals and audit trail as any human change" — covers Rhea steps 1–3 of the loop but not step 4 — [GoRules AI](https://gorules.io/ai); [GoRules architecture](https://docs.gorules.io/developers/overview/architecture)
- Event sourcing's guarantees match Rhea's replay invariant generically: every state change persisted as an immutable event in an append-only log; current state derived by replaying events in order; projections rebuilt from scratch by replay — [Event Sourcing in Backend Systems](https://thebackenddevelopers.substack.com/p/event-sourcing-in-backend-systems); [arc42 quality: Event Sourcing](https://quality.arc42.org/approaches/event-sourcing)
- 2026 research applies exactly this fusion to *agents* rather than *business rules*: "The Log is the Agent: Event-Sourced Reactive Graphs for Auditable, Forkable Agentic Systems" claims an event-sourced reactive substrate yields "auditable lineage, deterministic replay, and cheap counterfactual forks," with each state transition referring to the inputs it consumed — [arXiv 2605.21997](https://arxiv.org/pdf/2605.21997)
- Camunda frames deterministic DMN and runtime AI agents as alternatives to choose between — i.e., the market's dominant framing puts the LLM at runtime, where Rhea deliberately puts it at authoring time only — [Camunda blog (Jul 2025)](https://camunda.com/blog/2025/07/ai-agent-or-based-rule-dmn-ai-powered-orchestration/)
- The Cedar autoformalization line replaces Rhea's human approver with a deterministic verifier + LLM judge before policies take effect — machine-gated rather than human-gated activation — [arXiv 2606.26649](https://arxiv.org/html/2606.26649); [AutoCedar, arXiv 2607.03656](https://arxiv.org/html/2607.03656v1)

### Inferences
- The competitive map as of Oct 2026: (a) authoring loop with human gate — commoditizing fast (GoRules, IBM Bob mode, DecisionRules, Avalara); (b) deterministic rules-as-data execution — mature and widespread (ZEN/JDM, DMN, Cedar); (c) append-only event log + replay + per-field provenance *driven by those versioned rules* — found nowhere as a product. Rhea's defensible novelty is specifically (c) welded to (a) and (b): rules that don't just *decide* but *derive durable state*, with every derived field carrying `(event_id, rule_id, rule_version)` and whole-state reproducibility by replay.
- A second distinguishing detail: in Rhea the *approval itself is an event in the same log* (invariant 2). No surveyed product records approval as a replayable event; GoRules has an audit trail (adjacent but separate from execution data), Copilot Studio an admin console state.

### Gaps
- Private/stealth startups: this survey covers publicly documented products and papers; an unlaunched product with the full loop would not surface.
- Oracle (OIC/decision services), SAP (BRF+/Signavio), Pega, Sparkling Logic, Trisotech's AI features, and Gartner's decision-intelligence vendor coverage: not reached within the call budget; Trisotech in particular (DMN vendor employing Bruce Silver as a blogger) may have AI-assist features that went unexamined. The DMCommunity.org challenge archives were not directly searched.
- The Pierre Feillet (IBM) Medium article — likely the best single articulation of the "LLM extracts rules, human reviews, deterministic engine executes" thesis from inside IBM — returned HTTP 403 and could only be characterized secondhand.
