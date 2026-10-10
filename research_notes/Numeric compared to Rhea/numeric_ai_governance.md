# How Numeric's AI works and is governed (numeric.io, Numeric Technologies, San Francisco, CEO Parker Gilbert) — state as of 8 October 2026

Scope note: "Numeric" here is only the close/ERP-replacement vendor at numeric.io. All quotes are verbatim from the cited page unless marked "paraphrase". Vendor marketing is labelled as such. Items the sources do not establish are listed under Gaps, not inferred.

Reading guide for the Rhea comparison: Numeric's own slogan is "AI writes the rules. Rules do the accounting." On paper that is the same authoring-time/executing-time split Rhea makes (AI drafts a rule as data, a human approves it, a deterministic engine applies it). The notes below show where that holds (cash-matching rules, agentic-subledger treatments, revenue-recognition rules, transaction monitors, Prism "sign-off → rule"), where it does not (flux narratives, technical-accounting memos, MCP-driven agents, App Studio code are produced by an LLM in the live loop each period and are reviewed as text), and where mechanism is unverifiable (no public rule schema, no public rule versioning, no published prompt/model-version logging for Numeric's own AI, no determinism/replay guarantee stated anywhere).

## Context: company facts and timeline

### Takeaway
Numeric is a 2020-founded San Francisco company that sold close management and flux analysis alongside an ERP from 2021-2025, added cash management (2025), an MCP server (mid-June 2026), a technical accounting assistant (April 2026), and on 22 September 2026 launched a "Financial Data Platform" positioned as a NetSuite replacement with "agentic subledgers", App Studio, Agent Builder and Prism.

### Cited Findings
- Co-founded by Parker Gilbert with Duke classmates Andrew Bihl (CTO) and Anthony Alvernaz; Gilbert previously ran finance at Hearth. — [CPA Practice Advisor, 20 Nov 2025](https://www.cpapracticeadvisor.com/2025/11/20/ai-accounting-platform-numeric-raises-51m-series-b/173638/)
- $51M Series B led by IVP announced 19 Nov 2025 (Menlo Ventures, Founders Fund, 8VC et al. participating); three products live at that point: Close Management, Analytics, Cash Management; plans to "launch additional products throughout 2026 and 2027". — [PR Newswire, 19 Nov 2025](https://www.prnewswire.com/news-releases/numeric-raises-51m-series-b-expanding-from-close-management-to-comprehensive-finance-platform-302619774.html); [Numeric blog](https://www.numeric.io/blog/numeric-raises-51m-series-b)
- $28M Series A led by Menlo Ventures, announced autumn 2024. Menlo's own write-up at the time described Numeric as "an accounting automation platform that sits alongside an ERP". — [Menlo Ventures](https://menlovc.com/perspective/rewiring-the-financial-data-graph-our-series-a-investment-in-numeric/)
- Named customers across sources: OpenAI, Brex, Plaid, Wealthfront, Notion, Asana, Public.com, Clipboard Health, DailyPay, Benchling, STASH, GOAT, Built, Riskified. — [Numeric blog](https://www.numeric.io/blog/numeric-raises-51m-series-b); [numeric.io home](https://www.numeric.io/)
- Financial Data Platform (FDP) launched 22 Sep 2026 as "an ERP replacement built for enterprise companies with complex accounting"; "hundreds of NetSuite companies already on Numeric". Gilbert: "I view 'AI-native ERPs' largely as inheriting the prior generation of infrastructure, but with AI features bolted on. Same house, new paint." — [ACCESS Newswire via FinancialContent, 22 Sep 2026](https://www.financialcontent.com/article/accwirecq-2026-9-22-numeric-launches-financial-data-platform-the-erp-replacement-built-for-the-agent-era)
- Accounting Today (25 Sep 2026, Chris Gaetano) reports Gilbert clarified the FDP "is specifically a replacement for NetSuite" covering GL, subledgers, billing, revenue, cash, close, reporting, not CRM/inventory; "It's an on-ramp, not a cliff." — [Accounting Today](https://www.accountingtoday.com/news/numeric-launches-financial-data-platform-to-replace-erps)
- MCP server launched mid-June 2026 ("cumulative since launch (mid-June 2026)"). — [Numeric blog, 31 Jul 2026](https://www.numeric.io/blog/numeric-mcp-usage-report)
- Technical Accounting Assistant announced 1 Apr 2026 (author Alec Schon). — [Numeric blog](https://www.numeric.io/blog/our-new-technical-accounting-assistant)
- Headcount approx. 125 (third-party directory data; unverified). — [search summary of directory listings](https://www.cfoshortlist.com/vendors/numeric)

### Inferences
- The product has moved in two years from "AI-assisted close checklist beside your ERP" to "we are the ledger", which matters for the comparison: only in the FDP does Numeric own the system of record and so only there can its governance claims about rules, events and audit trails apply to the numbers themselves rather than to annotations of NetSuite data.

### Gaps
- No public statement of total customer count found.
- The legal entity name ("Numeric Technologies, Inc.") was not confirmed on a fetched page; the terms page was not fetched.

## Q1. Inventory of AI features (what each does, with examples)

### Takeaway
As of October 2026 Numeric ships roughly ten AI-labelled capabilities. They split into (a) LLM-generated text/analysis in the live close loop (flux narratives, technical accounting memos, MCP chat/agents, App Studio-generated apps, bank statement parsing) and (b) LLM-drafted rules/treatments that are then applied by a rules engine (cash-matching rules, transaction monitors drafted from natural language, agentic-subledger treatments, revenue rules, Prism "sign-off → rule").

### Cited Findings
Flux / variance analysis (oldest AI feature, pre-2025)
- "In a single click, AI combs through all of your transactions as well as external sources to surface core drives of variance." Explanations can be formatted as "bullet points, an essay, or even poetry"; users "simply edit, then approve". Data sources: transaction-line detail, integrations, file uploads (e.g. payroll CSV, marketing budget via NetSuite SavedSearch). — [numeric.io/product/variance-analysis](https://www.numeric.io/product/variance-analysis)
- Per-account templates: "Set custom templates for variance analysis explanations for each account... AI combs through every transaction in your GL and surfaces core drivers of variance for you to simply edit and approve." Materiality thresholds decide which variances need commentary. "Numeric's flux agent mirrors how accountants work— pivoting by specific NetSuite dimensions for each account, referencing entity-level explanations, and taking into account payroll or revenue data detail from outside the GL." — [numeric.io/solutions/variance-analysis-software](https://numeric.io/solutions/variance-analysis-software)
- Homepage metric: "80%" of flux first drafts written by AI (logo-adjacent, attribution to DailyPay inferred from layout; the DailyPay case-study page itself contains no percentage). — [numeric.io home](https://www.numeric.io/); [DailyPay case](https://www.numeric.io/cases/numeric-x-dailypay)
- Brex: "It gives an overall basis of what's changed, and then I just have to fill in the blanks." (Stephanie Lau) — [Brex case](https://www.numeric.io/cases/brex-x-numeric)
- STASH reports 90% flux coverage of GL accounts (search summary of case page). — [STASH case](https://www.numeric.io/cases/numeric-x-stash)
- Help-center Flux collection has three articles (thresholds, sidebar, balance-monitoring alerts); none documents how the AI draft is generated. — [help.numeric.io Flux](https://help.numeric.io/collections/9886677638-flux)

Cash matching / bank reconciliation
- Rules engine handles "1-to-1 to many-to-many matches"; an AI assistant creates rules "in seconds"; "AI agents watch your manual matches and propose new rules" (paraphrase of page wording "Increase match rate each month with suggestions from AI agents"); claims "90+% of your cash recs" automated, "three times the industry standard"; Numeric "drafts entries for exceptions and posts them in seconds". — [numeric.io/product/cash-matching](https://www.numeric.io/product/cash-matching); [numeric.io home](https://www.numeric.io/)
- Brex: match rate "from 30% to over 90%"; offshore team's manual match share fell from ~70% to under 20% in the first month; close from 6 to 4 business days. — [PR Newswire](https://www.prnewswire.com/news-releases/numeric-raises-51m-series-b-expanding-from-close-management-to-comprehensive-finance-platform-302619774.html); [Brex case](https://www.numeric.io/cases/brex-x-numeric)
- Cash Management tier includes "AI parsing of bank statements" and "automated matching for up to 95% of bank transactions". — [numeric.io/pricing](https://www.numeric.io/pricing)

Journal entry automation
- "automatically drafts and posts the journal entries you need to keep every cash account accurate and up to date" (fees, transfers, intercompany); posts to NetSuite in batches; each entry "tied to an underlying rule and preparer"; downloadable audit-trail report. — [numeric.io/product/journal-entry-automation](https://www.numeric.io/product/journal-entry-automation)

Account reconciliation
- "You can automatically submit any account that ties out and remains under your set materiality threshold." "Numeric will point you directly to the GL transactions that result in accounts no longer reconciling." No AI described on the page. — [numeric.io/product/reconcile](https://www.numeric.io/product/reconcile)
- Third-party review describes "risk-ranked exceptions, and auto-certification for low-risk accounts". — [CFO Shortlist, 6 Oct 2026](https://www.cfoshortlist.com/vendors/numeric)

Transaction monitors (anomaly detection)
- Rule-based, user-defined filters/thresholds: "Create monitors that reflect your team's accounting policies"; "Automatically flag any transaction above a size threshold for review." Page describes no AI/ML detection. — [numeric.io/product/transaction-monitors](https://www.numeric.io/product/transaction-monitors)
- Numeric SME Ben Sheridan: "You can use natural language to request a certain type of monitor and the AI will draft and spin it up for you." — [RightRev webinar recap, 16 Dec 2025](https://www.rightrev.com/agentic-finance-whats-real-whats-hype-what-to-implement/)
- Monitors "work just like NetSuite Saved Searches" but "more intuitive to build and work with real-time data". — [Numeric blog, 17 Dec 2025](https://www.numeric.io/blog/numeric-data-foundation-built-for-your-close-and-your-controls)

Technical Accounting Assistant (1 Apr 2026)
- Asks clarifying questions, identifies topics "from a taxonomy of hundreds", decides between quick Q&A and a full memorandum, "flags fact patterns that would change the conclusion"; Peer Review tab reviews uploaded memos/financial statements; Disclosures tab pulls comparable SEC filings; cites ASC codification and Big 4 guides "refreshed monthly"; can refer to technical accounting firms. "It doesn't replace anyone's knowledge." "This release stays in the analysis layer: memos, disclosures, policy." — [Numeric blog](https://www.numeric.io/blog/our-new-technical-accounting-assistant)
- Described elsewhere as "a technical accounting bot" that "answers US GAAP, IFRS, and UK GAAP questions", "trained on GAAP guidance and public filings". — [Numeric blog, 30 Jun 2026](https://www.numeric.io/blog/ai-accounting-tools)

"Numeric Intelligence"
- "Numeric Intelligence analyzes the close to spot bottlenecks and rebalance workload." — [Numeric blog, 30 Jun 2026](https://www.numeric.io/blog/ai-accounting-tools). The AI-audit-trail post lists "Numeric Intelligence (drafts flux, matches transactions, captures the evidence trail)". — [Numeric blog, 31 Jul 2026](https://www.numeric.io/blog/ai-audit-trail)

Numeric MCP server (mid-June 2026)
- Endpoint https://api.numeric.io/mcp; setup documented for Claude.ai/Desktop/Code; ChatGPT, Gemini, Cursor also mentioned. Tools include read (get_workspace_context, list_tasks, get_task_events, build_report, get_report_data, query_transaction_lines, get_flux_explanations ...) and write (create_task, edit_task, assign_task, submit_task — submit/approve/return — add_task_comment, update_flux_explanation). Advanced use cases: "Accruals bot that calculates estimates, updates schedules, posts entries, and submits tasks for review"; "Auditor prep packet built from task event logs". — [Numeric blog, 20 Apr 2026](https://www.numeric.io/blog/numeric-mcp-server)
- Usage (share of adopting companies, cumulative since mid-June 2026): 79.7% have AI check close task status, 66.7% build report data from scratch, 42.3% drill into transaction lines; one company created ~50 tasks and assigned 100+ via MCP; "another team writes Claude Skills into task procedure tabs". — [Numeric blog, 31 Jul 2026](https://www.numeric.io/blog/numeric-mcp-usage-report)
- "Claude Skills Library" / MCP Use Case Library: 18 templated workflows incl. "Accrual Identification & Journal Entry", "Auto-Draft Flux Explanations", "Journal Entry Posting (support-to-NetSuite)", "Audit Evidence Export", "SOP Documenter" (records how a task was performed "with gaps flagged instead of filled in"). — [numeric.io/claude-skills-library](https://www.numeric.io/claude-skills-library)

Financial Data Platform components (22 Sep 2026)
- Agentic Subledgers: when a record arrives, the system "determines the accounting treatment, builds a schedule, updates balances, and links support"; "you never have to post a JE again"; for uploaded documents the agent "proposes the correct accounting treatment for your approval"; covers fixed assets, leases, prepaids, accruals, deferred commissions, card spend, payables, equity payouts; sources Ramp, Brex, Stripe. — [numeric.io/product/agentic-subledgers](https://www.numeric.io/product/agentic-subledgers)
- Custom Ledgers: "Upload your documents or connect sources and let Numeric draft the repeatable workflow for you." "Policies are documented, followed, and enforced by the system." — [numeric.io/financial-data-platform](https://www.numeric.io/financial-data-platform)
- Prism: "the first screen you open in Numeric"; daily digest of overnight agent activity; "Every card comes with a suggested resolution"; "Prism is scoped to your seat"; "For recurring tasks, a user's sign-off can be turned into a rule so later instances proceed without them" (paraphrase). — [numeric.io/product/prism](https://www.numeric.io/product/prism)
- App Studio: "draft a few lines, then let agents build a secure application in Numeric"; each app "versioned and logged", covering "the prompts and code that created it" and "the data it reads"; "Spell out what you need and App Studio plans it, writes the code, and compiles it against your live financial data." — [numeric.io/product/app-studio](https://www.numeric.io/product/app-studio); [numeric.io/product/ai-in-numeric](https://www.numeric.io/product/ai-in-numeric)
- Agent Builder / background agents: "Delegate background tasks to agents, like chasing overdue invoices"; "Build custom agents and applications with AI"; background agents "to review transactions above a threshold or help with collections". — [numeric.io home](https://www.numeric.io/); [Accounting Today](https://www.accountingtoday.com/news/numeric-launches-financial-data-platform-to-replace-erps)
- Revenue recognition: "Your revenue policies are defined as rules in the system"; "Numeric derives the revenue schedule for each straight from the contract"; "Every number traces back to the contract term, performance obligation, and business event behind it." — [numeric.io/product/revenue-recognition](https://www.numeric.io/product/revenue-recognition)

### Inferences
- The feature set is bimodal. Rule-producing features (matching rules, monitors, subledger treatments, revenue rules) fit an "AI authors, engine executes" model. Text-producing features (flux, memos, MCP agent outputs, App Studio code) are per-period LLM generation reviewed as prose or code.
- "Agentic subledgers" is the closest analogue to Rhea's rule-as-data: a treatment is proposed once, approved, then re-applied to like records ("Approve once and every record like it follows suit"). Whether the stored artifact is a declarative rule, a prompt, or generated code is not disclosed (see Q2 gaps).

### Gaps
- No feature-level documentation (help center) was retrievable for AI flux, agentic subledgers, Prism or App Studio; the MCP help article is behind a login (redirects to numeric.portal.usepylon.com). Feature descriptions above are from marketing pages and blog posts only.
- No public example of an actual AI-drafted rule (its fields/conditions) was found; the Brex quote about rule visibility is the only evidence rules are human-readable.

## Q2. Where the AI runs: live execution loop vs. authoring of reusable artifacts

### Takeaway
Numeric explicitly claims an authoring/execution split for the number-producing paths ("Models never compute a number or draft a journal entry"; "AI is probabilistic in suggesting rules... it's deterministic software which is creating the accounting"), while the narrative paths (flux, memos, MCP agents, App Studio) run an LLM inside each period's close loop. There is no published statement that the rules engine is deterministic in the replay sense, nor any description of the rule format.

### Cited Findings
- "AI writes the rules. Rules do the accounting." "Whether it's cash matching or revenue recognition, rules drive the result. Models never compute a number or draft a journal entry." "Agents draft these rules for you. They read a record, propose a treatment, and bring it for review. Approve once and every record like it follows suit." — [numeric.io/product/ai-in-numeric](https://www.numeric.io/product/ai-in-numeric)
- "Rather than pointing an LLM at context-less journal entries, Numeric keeps every financial record whole." "AI takes your recurring work and only brings back what needs human judgment." — [numeric.io/product/ai-in-numeric](https://www.numeric.io/product/ai-in-numeric)
- Gilbert (Series B): "We route the right work to the right operator: AI to interpret unstructured data, code to handle calculations with precision, and humans to add context." Company framing: "combines AI for pattern recognition with deterministic code for calculations and human oversight for exceptions." — [PR Newswire, 19 Nov 2025](https://www.prnewswire.com/news-releases/numeric-raises-51m-series-b-expanding-from-close-management-to-comprehensive-finance-platform-302619774.html)
- Ben Sheridan (Numeric accounting SME), 16 Dec 2025: "AI is probabilistic in suggesting rules which might be successful in the engine"; "it's deterministic software which is creating the accounting"; on LLMs: "It's not necessarily going to always give me the same answer every single time." — [RightRev webinar recap](https://www.rightrev.com/agentic-finance-whats-real-whats-hype-what-to-implement/)
- FDP data model: stores "each business event as a complete record... with debits and credits as one projection of that record rather than the record itself"; "Background agents, applications, and custom ledgers with hardcoded audit trails are a prompt away". Accountants "define the rules once and review the exceptions". — [ACCESS Newswire, 22 Sep 2026](https://www.financialcontent.com/article/accwirecq-2026-9-22-numeric-launches-financial-data-platform-the-erp-replacement-built-for-the-agent-era)
- Accounting Today: "flows" are rule sets that tag each business event with dimensions ("if the vendor tag is X, do Y") and route it to a treatment (e.g. fixed-asset or lease ledger); once a ledger exists, "journal entries are generated monthly through 'hard-coded rules'"; "Agentic subledger features let AI draft rules for a new ledger, and controllers approve and refine them." — [Accounting Today, 25 Sep 2026](https://www.accountingtoday.com/news/numeric-launches-financial-data-platform-to-replace-erps)
- Reports: report logic is "deterministic, formulaic source code"; "The logic is defined, repeatable, and auditable." — [Numeric blog, 17 Dec 2025 (Chris Canoles)](https://www.numeric.io/blog/numeric-data-foundation-built-for-your-close-and-your-controls)
- Prism: a user's sign-off on a recurring task "can be turned into a rule so later instances proceed without them" (paraphrase). — [numeric.io/product/prism](https://www.numeric.io/product/prism)
- Agentic subledgers: when a document changes, "re-uploading it reuses the treatment you previously defined, and downstream accounting updates" (paraphrase). — [numeric.io/product/agentic-subledgers](https://www.numeric.io/product/agentic-subledgers)
- Live-loop AI: flux drafts are generated "in a single click" each period from current transactions (see Q1); MCP "Accruals bot... calculates estimates, updates schedules, posts entries, and submits tasks for review". — [numeric.io/product/variance-analysis](https://www.numeric.io/product/variance-analysis); [Numeric blog, 20 Apr 2026](https://www.numeric.io/blog/numeric-mcp-server)
- "AI-drafted rules support bank reconciliation and journal entry batch posting"; FDP has "a 'deterministic rules engine' that classifies accounting impact". — [Numeric blog, 30 Jun 2026](https://www.numeric.io/blog/ai-accounting-tools)
- Task procedures are "reusable, human-authored instructions that AI agents follow". — [Numeric blog, 31 Jul 2026](https://www.numeric.io/blog/ai-audit-trail)

### Inferences
- Two different senses of "deterministic" are used: (1) rule evaluation by code rather than a model (Sheridan, Series B), and (2) report logic as formulaic source code (Canoles). Neither is a replay guarantee; no page claims that re-running the event history through the same rules reproduces state.
- The "accruals bot" MCP use case contradicts the spirit of "Models never compute a number": an external LLM client calculates an estimate and posts an entry via Numeric's task API. The ai-in-numeric claim is scoped to Numeric's in-product engine paths; the MCP route is governed by user permissions and task review instead (Q3).
- "Hard-coded rules" and "hardcoded audit trails" in launch copy, read with App Studio's "writes the code, and compiles it", suggest some AI-authored artifacts are generated code rather than declarative rule data. This is inference; Numeric does not say.

### Gaps
- Rule representation (declarative data vs. generated code vs. stored prompt) is not disclosed for any feature.
- No statement on whether rules are versioned, whether a rule version is pinned to the entries it produced, or whether rules can be re-run historically.
- No public eng blog posts: numeric.substack.com ("Numeric Engineering", Andrew Bihl) exposed no post list when fetched.

## Q3. Human-in-the-loop: what is reviewed, by whom, approval recording, unattended posting, audit trail for AI content

### Takeaway
Numeric's stated controls are workflow controls inherited from close management (preparer/reviewer/second reviewer, Review Notes that can block period close, auto-submit only under a materiality threshold, "no entry posts without the proper approvals"), plus "approve once" for AI-drafted rules/treatments and Prism as an exception queue. MCP agents act with the user's own permissions and their writes are recorded as task events. Numeric does not publish what is logged about its own AI generations (prompt, model version, parameters) and its own guidance says "'A human reviews the output' is not a control."

### Cited Findings
- Journal entry automation: no entry posts "without the proper approvals"; "preparer-reviewer workflows that catch inaccurate rules before they run"; each entry "tied to an underlying rule and preparer". — [numeric.io/product/journal-entry-automation](https://www.numeric.io/product/journal-entry-automation)
- Reconciliation: auto-submit only for accounts that tie out "under your set materiality threshold". — [numeric.io/product/reconcile](https://www.numeric.io/product/reconcile)
- Homepage: "Reconcile the balance sheet automatically. Your team reviews only the exceptions." — [numeric.io home](https://www.numeric.io/)
- Flux: AI drafts, user "simply edit, then approve". — [numeric.io/product/variance-analysis](https://www.numeric.io/product/variance-analysis)
- Agentic subledgers: agent "proposes the correct accounting treatment for your approval". Rules for a new ledger: "controllers approve and refine them". FDP "escalates automatically 'anywhere accountant review or input is needed'". — [numeric.io/product/agentic-subledgers](https://www.numeric.io/product/agentic-subledgers); [Accounting Today](https://www.accountingtoday.com/news/numeric-launches-financial-data-platform-to-replace-erps)
- Close checklist with agents: "You review what comes back instead of assembling it, and every step gets recorded to Numeric with the rest of the close." "Track every comment, task submission, attachment, and more with timestamps in a comprehensive audit trail." Auditors can get read-only access. — [numeric.io/product/close](https://www.numeric.io/product/close)
- Review Notes: reviewer opens, preparer responds, reviewer closes; "Period close can be configured to block if Review Notes are open or unresolved"; Numeric "maintains an audit trail of all activities and changes within your close process". — [Numeric blog, 17 Dec 2025](https://www.numeric.io/blog/numeric-data-foundation-built-for-your-close-and-your-controls)
- MCP: "Once connected, Claude has the same access as your Numeric user account." "Claude shows you exactly what will change before executing." "Numeric serves as the audit trail for everything they touch." submit_task supports submit/approve/return; auditor prep packet "Flags any tasks where the same person submitted and approved". — [Numeric blog, 20 Apr 2026](https://www.numeric.io/blog/numeric-mcp-server)
- AI-audit-trail post on Numeric MCP: "an AI agent acts as the user, inherits their permissions, and its actions land in Numeric's audit log. A companion evidence step can post logic, source references, and sign-off back to the task." (paraphrase of page) — [Numeric blog, 31 Jul 2026](https://www.numeric.io/blog/ai-audit-trail)
- MCP usage report: "A person still decides what belongs on the checklist and reviews it before the period starts." — [Numeric blog, 31 Jul 2026](https://www.numeric.io/blog/numeric-mcp-usage-report)
- Numeric's own standard for an AI audit trail (five elements): inputs and source references; "Prompt, model version, and parameters" needed "to reconstruct a non-deterministic result"; output and confidence signals; human review, overrides and approvals; immutable timestamps and change history. "'A human reviews the output' is not a control." "Evidence reconstructed months later from a chat transcript that no longer exists is audit theater." "The system-emitted audit log is generated by your platform, not the model." The post does not say Numeric logs prompts/model versions/parameters for its own features. — [Numeric blog, 31 Jul 2026](https://www.numeric.io/blog/ai-audit-trail)
- App Studio: apps are "versioned and logged" including "the prompts and code that created it" and "the data it reads"; when an auditor asks how a number was produced, "the answer is a record in the system". — [numeric.io/product/app-studio](https://www.numeric.io/product/app-studio)
- Build-vs-buy guidance (Numeric blog, 18 Mar 2026): "Build control structures around AI outputs, not into them."; a cited controller requires two human reviewers for AI-prepared work; a Numeric solutions engineer found an AI-built report "had silently mixed local and consolidated currency". — [Numeric blog](https://www.numeric.io/blog/build-vs-buy-in-the-age-of-ai)
- Sheridan on rule review: "you can actually look at these fields and conditions and validate that yes, this makes sense."; "give folks the tools to audit these outputs before deploying them." — [RightRev webinar recap](https://www.rightrev.com/agentic-finance-whats-real-whats-hype-what-to-implement/)
- Brex on rule visibility: "The visibility of what it actually is doing is readily apparent, so we know if it's functioning as intended or not." (Kyle Gillingwater) — [Brex case](https://www.numeric.io/cases/brex-x-numeric)
- DailyPay: "We have a theoretically permanent record of everything that we've done in the system" (Sarah Hill, Controller). — [DailyPay case](https://www.numeric.io/cases/numeric-x-dailypay)
- User complaint: the flux module "forces sign-off on every assigned line whether or not it needs an explanation" (G2 reviewer, summarized from search snippets; G2 page itself returned 403). — [G2 reviews](https://www.g2.com/products/numeric/reviews)

### Inferences
- The approval of an AI-drafted rule is a workflow event in the checklist/task system, not a first-class versioned object with its own lifecycle; nothing published describes rule states (draft/approved/active/superseded) or supersession.
- Can AI post without review? For in-product rules: yes once the rule is approved (that is the point of "approve once"). For MCP agents: yes, if the user's own permissions allow it; the only published guard is a preview and the task audit log. For flux/memo text: no, these are drafts until a human approves/signs off.
- Numeric's audit-trail standard (prompt + model version + parameters) is more demanding than what any Numeric product page claims to record; the gap between the essay and the product pages is itself a finding.

### Gaps
- Whether Numeric stores, for an AI-drafted flux explanation or rule, the prompt, model identifier, model version, retrieved inputs and the diff between AI draft and approved text, is not published.
- Role-based permission model for agents (beyond "same access as your user") and separation of duties enforcement for agent-submitted vs. human-approved items: not published outside the auditor prep packet heuristic.

## Q4. Learning: institutional memory, how knowledge is stored, whether it is inspectable/versioned by the customer

### Takeaway
Numeric does not claim model training or fine-tuning on a customer's closes. "Learning" is realised as (a) rule suggestions mined from manual matches, (b) treatments approved once and reused, (c) human-authored context attached to tasks (task procedures, notes, links, Claude Skills) that agents read, and (d) per-account flux templates and prior entity-level explanations used as context. These artifacts are visible to the team inside Numeric; versioning of them is not described except for App Studio apps.

### Cited Findings
- "AI agents watch your manual matches and propose new rules, which... continuously improves your match rate over time" (paraphrase of page). — [numeric.io/product/cash-matching](https://www.numeric.io/product/cash-matching)
- "Approve once and every record like it follows suit." — [numeric.io/product/ai-in-numeric](https://www.numeric.io/product/ai-in-numeric)
- Prism: sign-off on a recurring task can be turned into a rule (paraphrase). — [numeric.io/product/prism](https://www.numeric.io/product/prism)
- Flux: per-account custom templates; flux agent "referencing entity-level explanations". — [numeric.io/solutions/variance-analysis-software](https://numeric.io/solutions/variance-analysis-software)
- Context-as-data essay (David Mansour, 18 May 2026): "Teams also use Numeric to pin documentation and Claude Skills right to the task they're working on." "Users can add any relevant context to their tasks in Numeric -- whether that's in the form of links, attachments, or written notes." A Numeric Solutions Manager describes an "edge case repository" an agent could read, presented as a practice rather than a shipped feature. The article makes no claim that Numeric itself learns or retains memory. — [Numeric blog](https://www.numeric.io/blog/your-ai-agents-are-starving-for-context)
- Task procedures: "reusable, human-authored instructions that AI agents follow". — [Numeric blog, 31 Jul 2026](https://www.numeric.io/blog/ai-audit-trail)
- Benchling stores task-specific Claude Skills in Numeric task descriptions so Claude writes flux "the way their team expects"; "skills and context stored visibly in Numeric so the team can see what's being used" (paraphrase). — [Benchling case](https://www.numeric.io/cases/numeric-x-benchling)
- Custom Ledgers: "Upload your documents or connect sources and let Numeric draft the repeatable workflow for you." — [numeric.io/financial-data-platform](https://www.numeric.io/financial-data-platform)
- FDP: "AI proposes accounting treatments based on a team's policies" (paraphrase). — [Numeric blog, 30 Jun 2026](https://www.numeric.io/blog/ai-accounting-tools)
- Technical accounting assistant inputs: ASC codification, Big 4 guides "refreshed monthly", uploaded memos, SEC filings; company policies and past memos are not mentioned as inputs. — [Numeric blog, 1 Apr 2026](https://www.numeric.io/blog/our-new-technical-accounting-assistant)
- Vendor claim (2024): Numeric's AI vendors "(including OpenAI and Anthropic) are not allowed to use customer data to train their models" (paraphrase). — [Numeric blog, 26 Sep 2024](https://www.numeric.io/blog/implementation-of-ai)
- Sheridan: "You see a lot of embedded AI where it's actually not allowed to train or retain certain information". — [RightRev webinar recap](https://www.rightrev.com/agentic-finance-whats-real-whats-hype-what-to-implement/)

### Inferences
- Numeric's durable knowledge lives in three customer-visible stores: rules (matching, monitors, flows, ledger treatments), task-attached text (procedures, skills, notes), and flux templates. None is described as versioned with history, except App Studio apps ("versioned and logged").
- The "learning loop" for matching is suggestion-based: the model proposes a rule from observed manual matches; the human accepts; the engine applies. That is structurally the Rhea flywheel, minus any published versioning, provenance-to-rule-version on each match, or a replay test.

### Gaps
- No disclosure of how "AI agents watch your manual matches" works (retrieval over match history vs. heuristic rule induction vs. LLM prompt).
- No statement that past approved flux explanations are retrieved as few-shot context, though "referencing entity-level explanations" suggests it.
- No customer-facing version history for rules, monitors or task procedures was found in any source.

## Q5. Accuracy, hallucination and reliability claims; published evaluations; disclaimers

### Takeaway
Numeric publishes automation rates (90%+ cash auto-match; 80% of flux first drafts AI-written at one customer) but no accuracy, error-rate or hallucination figures for any AI feature, no evaluation methodology, and no in-product disclaimer text that I could find. Its own blog acknowledges "plausible-looking outputs" as the top failure mode.

### Cited Findings
- "90%+ auto-match rate"; legacy tools "often deliver match rates below 30%"; Brex "from 30% to over 90%". — [PR Newswire](https://www.prnewswire.com/news-releases/numeric-raises-51m-series-b-expanding-from-close-management-to-comprehensive-finance-platform-302619774.html)
- "90+% of your cash recs" "three times the industry standard"; no methodology given. — [numeric.io/product/cash-matching](https://www.numeric.io/product/cash-matching)
- Brex automating "over 95% of its cash reconciliations" (paraphrase); no accuracy rate given for AI outputs. — [Numeric blog, 30 Jun 2026](https://www.numeric.io/blog/ai-accounting-tools)
- Gilbert: "AI is powerful, but imperfect. Most importantly, accounting can't tolerate errors compounding." — [CPA Practice Advisor](https://www.cpapracticeadvisor.com/2025/11/20/ai-accounting-platform-numeric-raises-51m-series-b/173638/)
- Marketing: "Because those records are context-rich, live, and accurate, your LLM's output is always useful, reliable and ready-to-share." — [numeric.io/product/ai-in-numeric](https://www.numeric.io/product/ai-in-numeric)
- FDP page claims AI with "real accuracy" with no evidence offered. — [numeric.io/financial-data-platform](https://www.numeric.io/financial-data-platform)
- Technical accounting assistant post has no accuracy disclaimer and does not mention hallucination; it flags items that "need to be verified". — [Numeric blog, 1 Apr 2026](https://www.numeric.io/blog/our-new-technical-accounting-assistant)
- 2024 rollout guidance: "Bad data in, bad data out—it's that simple."; "Keep a Human in the Loop"; no mention of hallucination. — [Numeric blog, 26 Sep 2024](https://www.numeric.io/blog/implementation-of-ai)
- Build-vs-buy post: plausible-looking outputs are the highest-risk failure mode; an AI-built report "silently mixed local and consolidated currency"; "the audit world hasn't figured out how to evaluate AI-touched outputs." — [Numeric blog, 18 Mar 2026](https://www.numeric.io/blog/build-vs-buy-in-the-age-of-ai)
- Third-party reviewer: "The page makes no quantified accuracy claims for the AI"; advises buyers to "compare its flux explanations with your last two closes, and ask to see a wrong draft and the correction workflow". — [CFO Shortlist, 6 Oct 2026](https://www.cfoshortlist.com/vendors/numeric)
- A review site claims the AI "sometimes struggles with highly complex intercompany transactions; a qualified accountant must always review these reconciliations" — claims a 90-day test but gives no methodology; weight lightly. — [bestaiaccounting.com](https://bestaiaccounting.com/reviews/numeric-review/)

### Inferences
- Numeric's reliability argument is architectural (keep models out of arithmetic; route to code) rather than empirical. For the comparison, this means Numeric and Rhea make the same argument, and neither can be compared on published eval numbers.

### Gaps
- No published evaluation set, benchmark, error rate, or hallucination rate for flux, memos, matching-rule suggestions or treatment proposals.
- In-product disclaimer wording (if any) not visible from public pages.

## Q6. LLM providers and models

### Takeaway
Numeric's subprocessor list names OpenAI, Anthropic and Groq (the last "available upon request") under "AI Data Processing", configurable per workspace and per product module; Numeric does not publicly name which model powers which feature and makes no claim of proprietary models. Its MCP/skills ecosystem is Claude-first in documentation.

### Cited Findings
- Subprocessors table: "OpenAI — USA — AI Data Processing"; "Anthropic — USA — AI Data Processing"; "Groq* — USA — AI Data Processing"; "Subprocessors with a * above are available upon request." "AI Data Processors can be configured by workspace and can vary by product module." Changes via "your Numeric Solutions Manager". Page showed "last updated 22 days ago" when fetched on 8 Oct 2026 (an earlier version was dated 5 Feb 2025 per search snippet). — [help.numeric.io subprocessors](https://help.numeric.io/articles/8337651077-subprocessors)
- 2024: AI vendors "(including OpenAI and Anthropic)" may not train on customer data (vendor claim). — [Numeric blog, 26 Sep 2024](https://www.numeric.io/blog/implementation-of-ai)
- Privacy policy (updated 23 Sep 2025) does not mention AI, LLMs, OpenAI or Anthropic. — [numeric.io/legal/privacy-policy](https://www.numeric.io/legal/privacy-policy)
- MCP setup documented for Claude.ai, Claude Desktop, Claude Code (`claude mcp add --transport http numeric https://api.numeric.io/mcp`); ChatGPT, Gemini, Cursor mentioned as other clients. — [Numeric blog, 20 Apr 2026](https://www.numeric.io/blog/numeric-mcp-server)
- "Claude Skills Library" link in site navigation; the page refers generically to "your AI model". — [numeric.io/claude-skills-library](https://www.numeric.io/claude-skills-library)
- Product pages for ai-in-numeric, financial-data-platform, app-studio, technical accounting assistant do not name a model. — [numeric.io/product/ai-in-numeric](https://www.numeric.io/product/ai-in-numeric); [numeric.io/product/app-studio](https://www.numeric.io/product/app-studio)

### Inferences
- Groq as a subprocessor implies inference on open-weight models (Groq hosts Llama-class models) for some workloads; no Numeric statement confirms which.
- Per-module configurability of AI processors suggests a provider-abstraction layer, which is consistent with no single model version being pinned per feature publicly.

### Gaps
- Which provider/model backs flux, matching-rule suggestion, technical accounting assistant, App Studio or agentic subledgers is undisclosed.
- No statement on model version pinning or change notification to customers.

## Q7. Auditability and compliance: SOC, SOX, auditor-facing features, determinism/reproducibility/provenance claims

### Takeaway
Numeric states it is SOC 1 and SOC 2 certified (SOC 1 Type 2 cited for data immutability), offers auditor read-only access, timestamped task/comment/attachment audit trails, preparer/reviewer sign-offs, Review Notes, and in the FDP "Every action taken has an audit trail" plus auditor-viewable source code. Provenance claims are at the level of "traces back to the business event/contract/rule"; there is no published claim of replayable determinism or per-field provenance to a rule version.

### Cited Findings
- Pricing page: Numeric is SOC 1 and SOC 2 certified (type not stated). — [numeric.io/pricing](https://www.numeric.io/pricing)
- "The data transfer is lossless and immutable—numbers don't change when they move from your ERP to our platform." Supported by SOC 1 Type 2 report (per article). Numeric "doesn't process transactions" in the close-management product; it warehouses NetSuite data. — [Numeric blog, 17 Dec 2025](https://www.numeric.io/blog/numeric-data-foundation-built-for-your-close-and-your-controls)
- Close: "Track every comment, task submission, attachment, and more with timestamps in a comprehensive audit trail."; auditors can have read-only access. — [numeric.io/product/close](https://www.numeric.io/product/close)
- JE automation: each entry "tied to an underlying rule and preparer"; downloadable "complete audit trail" report. — [numeric.io/product/journal-entry-automation](https://www.numeric.io/product/journal-entry-automation)
- FDP: "Every action taken has an audit trail"; auditors "can also view and download source code"; "custom ledgers with hardcoded audit trails". — [Accounting Today](https://www.accountingtoday.com/news/numeric-launches-financial-data-platform-to-replace-erps); [ACCESS Newswire](https://www.financialcontent.com/article/accwirecq-2026-9-22-numeric-launches-financial-data-platform-the-erp-replacement-built-for-the-agent-era)
- Revenue: "Every number traces back to the contract term, performance obligation, and business event behind it." — [numeric.io/product/revenue-recognition](https://www.numeric.io/product/revenue-recognition)
- Agentic subledgers: "stores the complete financial record behind each transaction"; "tracing a transaction to its source takes less than a minute". — [numeric.io/product/agentic-subledgers](https://www.numeric.io/product/agentic-subledgers)
- App Studio: apps "versioned and logged" with prompts, code and data read; easier for an auditor to inspect than the spreadsheets they replace (paraphrase). — [numeric.io/product/app-studio](https://www.numeric.io/product/app-studio); [numeric.io/product/ai-in-numeric](https://www.numeric.io/product/ai-in-numeric)
- MCP: get_task_events "pulls the full event history for audit trails"; auditor prep packet from task event logs; segregation-of-duties flag. — [Numeric blog, 20 Apr 2026](https://www.numeric.io/blog/numeric-mcp-server)
- Numeric's own regulatory framing: SOX 404(a)/(b) cover AI-performed steps affecting financial reporting; PCAOB AS 1105 requires entity-produced information to be complete and accurate; "the PCAOB has not endorsed any AI-assisted accounting workflow yet"; recommends 7+ year retention and asking vendors whether they log data sources and model versions, keep records immutable, and fall within SOC 1 Type 2 scope. Quote (Chris Canoles, former EY audit partner): "Auditors are wary of agentic AI in financial workflows". — [Numeric blog, 31 Jul 2026](https://www.numeric.io/blog/ai-audit-trail)
- Customer: "How GOAT built a PWC audit-ready close" (case study title). — [numeric.io home](https://www.numeric.io/)
- Third-party: "SOX program management is lighter than FloQast's"; "expect more auditor education than with incumbents"; vendor claims such as "auditor-defensible by design" are "vendor positioning, not verified results". — [CFO Shortlist, 6 Oct 2026](https://www.cfoshortlist.com/vendors/numeric)

### Inferences
- Numeric's audit trail is an activity log (who did what, when) plus source linkage (which record/contract/rule). It is not described as an append-only event log from which state is rebuilt; the FDP's "business event as complete record... debits and credits as one projection" is the nearest statement and is architectural marketing without a replay/determinism claim.
- "Auditors can view and download source code" (FDP) implies that in the FDP the accounting logic for custom ledgers is code, inspectable but presumably not a declarative rule DSL.

### Gaps
- No public SOC report scope, no trust center page found (searches returned none at numeric.io).
- No published statement of determinism/reproducibility (replay) anywhere; no per-field provenance to (event, rule, rule version).
- Whether the FDP's business-event store is append-only / immutable is not stated.

## Q8. Pricing and packaging of AI features

### Takeaway
Pricing is quote-only. Four packages (ERP Replacement; Close Management & Analytics; Cash Management; Billing & Revenue). AI is not sold as an add-on line; "In-app AI (Prism)" is listed only in ERP Replacement; the MCP and technical accounting assistant are not listed in any tier. Third-party price figures conflict and are unverified.

### Cited Findings
- No prices published; factors include transaction volume and seats; all products purchasable separately except ERP Replacement (full suite); onboarding/migration fees "scoped upfront"; "Implementation alignment is a required step" for FDP. ERP Replacement lists "policy enforcement through hard-coded workflows" and "In-app AI (Prism)"; Close Management & Analytics lists auto-reconciliations, transaction monitoring, flux analysis, review notes; Cash Management lists automated matching "up to 95%", JE automation, bank feeds, AI bank-statement parsing. — [numeric.io/pricing](https://www.numeric.io/pricing)
- Numeric's own comparison: "Platform charge plus per-seat fees; scales with team size and complexity." — [numeric.io/blog/numeric-vs-ledge](https://www.numeric.io/blog/numeric-vs-ledge)
- Third-party estimate: "$20K–$80K+ per year, depending on users, entities, and AI feature tiers" (market-reported). — [CFO Shortlist](https://www.cfoshortlist.com/vendors/numeric)
- Conflicting third-party claim: "Essentials from $30/user/mo; Growth & Enterprise (custom quote)"; not found on numeric.io. — [aiagentsquare.com](https://aiagentsquare.com/agents/numeric-ai)

### Inferences
- "AI feature tiers" in the CFO Shortlist estimate is not corroborated by the official pricing page; treat as unverified.

### Gaps
- No usage caps, token limits or AI-specific SKUs are published.

## Q9. Known limitations and critiques (users, analysts)

### Takeaway
Independent criticism is thin (few G2 reviews, no Reddit threads surfaced). The recurring complaints are workflow friction (mandatory sign-off on every flux line), integration/sync errors during close, performance, limited ERP fields for flux, and that Numeric needed an ERP underneath (true until the FDP). Analysts flag vendor youth, lighter SOX tooling, enterprise/global-close ceiling, and auditor unfamiliarity; one independent guide scores "AI Innovation" 90/100 while urging buyers to test for wrong drafts.

### Cited Findings
- G2 (summarized from search snippets; the review page returned 403 on fetch): a controller calls "the AI flux analysis tool" their favourite feature; complaints: flux module forces sign-off on every assigned line; integrations "throw random error messages about failed task updates during the busiest MEC refresh days"; "internal failure errors" the most common during closes; manual sync and slow reconciliation page; request for more ERP fields (contract IDs, product codes) in flux; request for YoY variance. — [G2 reviews](https://www.g2.com/products/numeric/reviews)
- CFO Shortlist (6 Oct 2026): strong for lean teams on modern ERPs; weaker for "complex global closes, deep SOX programs, or legacy ERPs"; "close orchestration trails FloQast's maturity"; "high-volume matching is not the product's strength"; scores: AI Innovation 90, Core FP&A 12, Scalability 55, Reporting 60, Data Integration 68 (methodology not disclosed; site promotes its own app). — [CFO Shortlist](https://www.cfoshortlist.com/vendors/numeric)
- Not a standalone GL before the FDP; "you are paying for two layers of software". — [bestaiaccounting.com](https://bestaiaccounting.com/reviews/numeric-review/)
- Numeric's own staff acknowledge AI-built artifacts can fail silently (currency mix-up) and that auditors have not settled how to evaluate AI-touched outputs. — [Numeric blog, 18 Mar 2026](https://www.numeric.io/blog/build-vs-buy-in-the-age-of-ai)
- Numeric's own audit-trail essay concedes GAAS evidence standards become "harder to meet with non-deterministic models" and that review alone "is not a control". — [Numeric blog, 31 Jul 2026](https://www.numeric.io/blog/ai-audit-trail)
- Benchling's controller's "initial reaction was skepticism" toward AI via MCP (paraphrase). — [Benchling case](https://www.numeric.io/cases/numeric-x-benchling)
- Trade press on ERP replacement generally: "Replacing an ERP system is often one of the most disruptive technology projects inside a finance function" (search summary of IT Brief coverage). — [IT Brief UK](https://itbrief.co.uk/story/numeric-launches-financial-data-platform-to-replace-erps)

### Inferences
- No user critique found targets AI accuracy specifically; complaints are about workflow rigidity and sync reliability. This likely reflects that AI outputs are drafts behind human sign-off, so errors are absorbed by review rather than surfacing as complaints.

### Gaps
- No Reddit r/accounting or r/FPandA threads about Numeric surfaced in two searches.
- No analyst coverage from The Information or Fortune found; no Gartner/Forrester assessment found.
- G2 review count and rating could not be read directly (403); Numeric's own blog cites "4.8 out of 5 across 100+ reviews" (vendor claim, via search snippet of numeric.io/blog/numeric-vs-ledge).

## Summary table for the Rhea comparison (all cells sourced above)

| Dimension | Numeric (Oct 2026) | Evidence quality |
|---|---|---|
| Where AI sits | Both: authoring-time for rules/treatments/monitors ("AI writes the rules. Rules do the accounting."); live-loop for flux text, memos, MCP agents, App Studio code | Vendor wording, consistent across pages and a staff webinar |
| What AI produces | Rule suggestions, treatment proposals, monitor definitions, flux narratives, memos/Q&A, generated apps/code, task actions via MCP | Vendor pages |
| How output is checked | Human edit/approve for text; "approve once" for rules; preparer/reviewer/second reviewer; Review Notes gate period close; materiality-thresholded auto-submit; MCP preview before write | Vendor pages; one G2 complaint confirms mandatory sign-off exists |
| Durable, inspectable learning | Rules, flux templates, task procedures/skills/notes, approved treatments; visible in-app; versioning undocumented except App Studio apps; no fine-tuning claimed; vendors barred from training | Vendor pages |
| Determinism / replay | "deterministic code for calculations", "deterministic rules engine", report logic "deterministic, formulaic source code"; no replay/reproducibility guarantee stated | Vendor wording only |
| Provenance | Entry tied to rule and preparer; number traces to contract/obligation/business event; App Studio logs prompts+code+data; no per-field (event, rule, version) tuple published | Vendor wording |
| AI-content audit trail | Task event log, timestamps, sign-offs; Numeric's own standard asks for prompt/model version/parameters but product pages do not claim to store them | Vendor wording; gap acknowledged by Numeric's essay |
| Models | OpenAI, Anthropic, Groq (on request) as subprocessors; configurable per workspace/module; no feature-to-model mapping | Help-center subprocessor page |
| Compliance | SOC 1 and SOC 2 (SOC 1 Type 2 cited); auditor read-only access; auditor-downloadable source code in FDP | Vendor pages |
| Published accuracy | None; only automation rates (90%+ match, 80% flux first drafts) | Vendor metrics without methodology |
