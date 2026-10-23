# AI-native ERP and finance-platform entrants, as of 23 Oct 2026

Research for the target catalog of packs (`PACKS.md`), 2026-10-23. Companion to "Products similar
to Rhea" (the architecture side) and to "Incumbent ERP vendors 2026-10" (the old guard). Every fact
carries its source [n]; a matrix "N" means no evidence on the pages read, not a vendor's denial;
customer counts and percentages are the vendors' own claims unless noted.

## Vendors

**Light (Copenhagen)**
- **What it sells:** one ledger for all entities, shared by accounting, payables, receivables, cards and consolidation; global bill pay, spend management, subscription management, reporting, workflows, consolidation, multibook [1]. Procurement: purchase requests and orders only. Payroll through a Finch integration. No inventory, manufacturing or projects [7, a Light-sponsored review].
- **For:** multi-entity finance teams in global businesses [1].
- **AI:** "Agents do the work. You set policy and approve." Approval rules by entity, vendor or amount; "every action is logged to the agent that took it" [1]; the manifesto: "Every correction is training" [4].
- **Price:** not published [7]. **Implementation:** 4–16 weeks to the first close, "as little as 4 weeks" [1].
- **Funding:** $30M Series A led by Balderton (Sep 2025), $43M in total [5][6]; about 100 staff in July 2026 [3].
- **2026:** agents live 28 Apr with IFRS; Custom Agents and a Claude connector 13 Aug; App Store 1 Sep; KPMG partnership 8 Sep; withholding tax and intercompany invoicing 14 Oct [3].
- **Localisation:** the strongest in the cohort — 21 country pages (Peppol, XRechnung/ELSTER, SdI, SII, SAF-T, MTD) [2]. **Gap:** Poland not listed, KSeF not mentioned [2].

**Rillet (US)**
- **What it sells:** GL, AR (contract and usage invoicing, aging, reminders), bank reconciliation, close, ASC 606 revenue recognition, consolidation, reporting [8][9]. AP by connecting an existing AP tool; payroll via Rippling [9].
- **For:** mid-market, enterprise and public companies [8]; 600+ customers, from tech and AI towards biotech, healthcare, fintech, logistics, professional services [10].
- **AI:** the Aura agents; claimed 99.7% of journal entries booked automatically [9]; "Accounting Superintelligence" with human approval and an audit trail [10].
- **Price:** quoted per customer [9]; a third party reports roughly $20–35k a year (uncertain) [13]. **Implementation:** 4–6 weeks, CPA-led [9].
- **Funding:** $100M Series C led by ICONIQ at $1B (Aug 2026), over $200M in total [10].
- **2026:** EY alliance 29 Apr [11]; partner of more than half of the Accounting Today top-20 CPA firms [10]; replaces Oracle Fusion, SAP, Workday, Great Plains, NetSuite [10].
- **Gaps:** no FP&A, procurement or inventory found [9]; tax via Avalara [8]; no EU e-invoicing found.

**Campfire (US)**
- **What it sells:** GL with multi-entity consolidation, revenue automation, close (checklists, flux, reconciliations), reporting, Forecast; 180+ currencies, unlimited entities [14].
- **For:** tech companies (Replit, Decagon, PostHog) [15]; about 100 customers and 12 staff in mid-2025 [17].
- **AI:** Ember Agents (beta, Mar 2026) run continuously on matching, AP/AR, anomalies, accruals; confidence thresholds decide what goes to review; customers "expand autonomy as trust grows"; a proprietary "Accounting Intelligence" model [15].
- **Price:** not public [19]. **Implementation:** 3–10 weeks, per a third party [19].
- **Funding:** $65M Series B co-led by Accel and Ribbit (Oct 2025), $100M in total [18][16].
- **2026:** London office (22 Jun) with three UK partner firms; UK support only as "HMRC requirements" [16].
- **Gaps:** inventory through a Doss integration [20]; no HR, inventory or manufacturing module [19].

**DualEntry (New York)**
- **What it sells:** GL, AR, AP, cash, tax management, purchase orders, order management, revenue recognition, subscription billing, fixed assets, consolidation, multi-book, planning and budgeting, treasury [21].
- **For:** from $5M-revenue companies to NYSE-listed firms, "thousands of users" [25].
- **AI:** "Agentic AI ERP", always-on auditing, anomaly detection; on approvals only "Tailor approvals to your needs" [21].
- **Price:** three tiers by entities (3, 20, unlimited; the top adds FP&A and advanced revenue recognition), implementation included, no prices published [22].
- **Implementation:** "Go live in 24 hours" and "4–6 weeks" both appear [21]; NextDay Migration from NetSuite, Intacct, SAP, Dynamics, QuickBooks Desktop, PeopleSoft [23].
- **Funding:** $90M Series A led by Lightspeed and Khosla (Oct 2025), over $100M in total [23]; a ~$415M valuation is an aggregator estimate, uncertain [24].
- **2026:** a $6M "ERP Rescue Fund" paying off legacy ERP contracts [26]; European HQ in London, no VAT or e-invoicing mentioned [25].
- **Gaps:** no inventory or payroll found.

**Numeric (US)**
- **What it sells:** began as close management on NetSuite (checklists, reconciliations, flux), then cash matching, billing and revenue recognition, collections, and an ERP replacement [27][28].
- **For:** fintech and software (Brex, Plaid, OpenAI, Wealthfront) [29].
- **AI:** Agent Builder (e.g. chasing overdue invoices), MCP access, "review only the exceptions" [27].
- **Price:** quoted, by transaction volume and seats [28]. **Funding:** $51M Series B led by IVP (Nov 2025) [29]; ~$89M total, uncertain [30].
- **2026:** the Financial Data Platform (22 Sep) keeps each business event as a full record with debits and credits as one output; rules called "flows" tag dimensions [31] — the closest architecture to Rhea's (see "Numeric compared to Rhea").
- **Gaps:** no AP, procurement or operations.

**Digits (US)**
- **What it sells:** the "Agentic General Ledger": bookkeeping, close, accruals, bill pay, invoicing, reporting; payroll by connection [32].
- **For:** small businesses, startups and accounting firms; 7,000+ accountants and owners claimed [32].
- **AI:** proprietary models trained on $825B of small-business transactions, claimed to beat GPT-4o by 54% [34].
- **Price:** $65, $100 or $250 a month; firms pay per client, their own books free [33]; from 7 Apr 2026 outcome-based pricing for firms — pay only for clients where at least 95% of transactions need no human touch [35].
- **Implementation:** self-serve, 30-day trial [33]. **Funding:** "nearly $100M" [36]; no round found after 2022 (uncertain).
- **Gaps:** US small business only; no revenue recognition, consolidation or localisation found.

**Lensing (formerly Everest Systems)**
- **What it sells (claimed):** order-to-cash, record-to-report, revenue recognition (ASC 606/IFRS 15), multi-book with localisation, inventory, professional-services automation, HR, asset management [37].
- **For:** complex-revenue companies, especially SaaS; "$1B running through the platform" [37].
- **AI:** AiSpecify, a "full software development lifecycle platform", and a Live Sandbox to "simulate any changes before approving them" [37]. **Implementation:** "weeks, not months" [37].
- **Funding:** $140M out of stealth (Sutter Hill, Altimeter, Redpoint, D1) [38].
- **Gaps:** no public documentation, breadth unverified [39]; a handful of G2 reviews [40].

**Others worth knowing**
- **Doss:** AI-native operations ERP (inventory, procurement, orders), the GL left to an accounting system; $55M Series B, Mar 2026 [41][42].
- **Basis:** AI agents sold to accounting firms; $100M at $1.15B (Feb 2026); about 30% of the US top-25 firms [43].
- **Accrual:** AI tax-return preparation for firms; launched Feb 2026 with $75M [44]; acquiring Puzzle's accounting-firm business (Puzzle claims 7,000+ users) [45].
- **Maximor:** close agents on top of existing ERPs; $9M seed [52].
- **Pennylane (France):** the European reference: €175M Series E (Jan 2026) at a reported $4.25B [46][47]; 6,500+ accounting firms and 1.5M businesses claimed; France, Germany, Spain; "certified" for France's September 2026 e-invoicing mandate [48]; free for firms that keep their clients' books on it [49].

## Coverage matrix

Y covered, P partial (claimed or via a partner), N not found.

| | GL/close | AP | AR | Rev rec/subs | Consol. | FP&A | Procure | Inv/WMS | Mfg | Projects | HR/pay | CRM | Treasury | Tax/e-inv local. |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Light | Y | Y | Y | P | Y | P | P | N | N | N | N | N | P | P (21 countries, no PL) |
| Rillet | Y | P | Y | Y | Y | N | N | N | N | N | N | N | N | P (Avalara) |
| Campfire | Y | Y | Y | Y | Y | P | P | P (Doss) | N | N | N | N | P? | N |
| DualEntry | Y | Y | Y | Y | Y | P | P | N | N | N | N | N | P | P (unspecified) |
| Numeric | Y | N | P | Y | N? | N | N | N | N | N | N | N | P | N |
| Digits | Y (SMB) | Y | P | N | N | N | N | N | N | N | N | N | N | N |
| Lensing | Y | P | Y | Y | Y | N | P | P* | N | P* | P* | N | N | P* |
| Doss | N | N | N | N | N | N | Y | Y | P | N | N | N | N | N |
| Pennylane | Y | P | Y | N | N | N | N | N | N | N | N | N | P | Y (FR) |

\* Claimed with no public documentation.

## Patterns

- **They sell the replacement of NetSuite or QuickBooks** [10][17][23][26][28][36].
- **Weeks, not months:** vendor-led, CPA-staffed implementation in 4–16 weeks, often included [1][9][22]; DualEntry claims a 24-hour migration [23].
- **The finance office only:** all stop at record-to-report and order-to-cash; operations by partners (Doss with Campfire [20]), payroll and CRM by integration [9][32].
- **One AI story:** agents draft, a person approves by policy or confidence, every action attributed [1][15][35]; the learning lives in model weights or proprietary models [4][15][34], not in inspectable versioned rules. Only Lensing (sandbox, then approve) [37] and Numeric (event records plus flows) [31] point toward logic as data.
- **VC-backed US tech first** [10][15][29]; quote-based pricing, only Digits publishes [33].
- **Channel through audit and CPA firms:** EY–Rillet [11], KPMG–Light [3], Campfire's UK partners [16]; agents sold to firms: Basis [43], Accrual [44], Digits' per-client and outcome pricing [35].
- **Europe means London first** [16][25], with no EU tax depth found.

## Where they are weak

- **Against a full ERP:** no manufacturing anywhere; inventory only through Doss or Lensing's unverified claim; procurement stops at purchase orders; native payroll, CRM and projects absent [7][19][13].
- **Against Polish needs:** none lists Poland or KSeF (Light: 21 countries without Poland [2]). KSeF mandatory since 1 Feb 2026 for large taxpayers, 1 Apr 2026 for other VAT payers, micro-businesses from 1 Jan 2027 [50]; no JPK on any vendor page read; the cohort's centre of gravity is US GAAP and ASC 606 [9][37].
- **Accounting offices as a channel:** the US cohort sells to CPA firms serving mid-market clients; nobody targets offices running many sole traders and small companies on Polish statute. Pennylane proves the model in Europe — thousands of firms, free for firms, an e-invoicing mandate as the wedge — in France, Germany and Spain, not Poland [48][49].

## Sources

1. https://light.inc
2. https://light.inc/countries
3. https://light.inc/blog
4. https://light.inc/manifesto
5. https://www.balderton.com/news/light-raises-30m-series-a-to-replace-legacy-finance-systems-with-ai-native-platform/
6. https://tech.eu/2025/09/25/light-raises-30m-series-a-to-redefine-finance-for-hypergrowth-companies/
7. https://www.erpresearch.com/en-us/light-erp (sponsored by Light)
8. https://www.rillet.com
9. https://info.rillet.com/what-is-rillet
10. https://pulse2.com/rillet-raises-100-million-series-c-at-1-billion-valuation-as-ai-native-erp-tops-600-customers/
11. https://www.ey.com/en_us/newsroom/2026/04/ey-announces-alliance-with-rillet-to-provide-ai-native-finance-transformation-with-risk-and-controls-built-in
13. https://www.erpresearch.com/en-us/ai-native-erp
14. https://campfire.ai
15. https://fortune.com/press-releases/campfire-ember-agents-ai-workers-finance-accounting-2026-03-12
16. https://fortune.com/press-releases/campfire-london-office-ai-native-erp-uk-finance-2026-06-22/
17. https://techcrunch.com/2025/06/30/tiny-ai-erp-startup-campfire-is-winning-so-many-startups-from-netsuite-accel-led-a-35m-series-a
18. https://fintech.global/2025/10/16/campfire-secures-65m-series-b-to-transform-finance-ai/
19. https://www.erpresearch.com/en-us/campfire-erp
20. https://www.doss.com/integrations/campfire
21. https://www.dualentry.com
22. https://www.dualentry.com/pricing
23. https://www.unite.ai/dualentry-raises-90-million-to-reinvent-business-accounting-software-with-ai
24. https://www.caplight.com/company/dualentry
25. https://www.dualentry.com/blog/dualentry-european-headquarters-london
26. https://www.dualentry.com/blog/6m-rescue-fund-launch
27. https://www.numeric.io
28. https://www.numeric.io/pricing
29. https://www.numeric.io/blog/numeric-raises-51m-series-b
30. https://fintech.global/?p=222853
31. https://www.accountingtoday.com/news/numeric-launches-financial-data-platform-to-replace-erps
32. https://digits.com
33. https://digits.com/pricing
34. https://accountingtoday.com/news/digits-announces-autonomous-general-ledger
35. https://www.digits.com/blog/outcome-based-pricing/
36. https://www.globenewswire.com/news-release/2025/03/10/3039814/0/en/AI-Startup-Digits-Takes-on-QuickBooks-with-the-World-s-First-Autonomous-General-Ledger-for-Accounting-Xero-Co-founder-Craig-Walker-Joins-Digits.html
37. https://lensing.ai/
38. https://www.everest-systems.com/everest-systems-emerges-from-stealth-with-140-million-funding
39. https://erpresearch.com/erp/everest-systems
40. https://www.g2.com/products/lensing/reviews
41. https://pulse2.com/doss-55-million-raised-for-ai-native-operations-cloud-platform
42. https://erpresearch.com/en-gb/doss-erp
43. https://www.cpapracticeadvisor.com/2026/02/24/basis-raises-100-million-to-deploy-ai-agents-for-accounting-firms/178759/
44. https://www.cpapracticeadvisor.com/2026/02/05/startup-accrual-officially-launches-with-75m-in-funding-to-bring-ai-native-automation-to-accounting/
45. https://www.01net.it/accrual-to-acquire-puzzle-expanding-its-ai-platform-into-client-accounting-services/
46. https://tech.eu/2026/01/20/french-accounting-software-platform-pennylane-raises-200m
47. https://pulse2.com/pennylane-175-million-funding/amp/
48. https://www.pennylane.com/en
49. https://www.pennylane.com/fr/expert-comptable/pdp
50. https://www.ey.com/en_gl/technical/tax-alerts/poland-signs-into-law-mandatory-national-e-invoicing-system
52. https://techeconomy.ng/maximor-raises-9m-finance-automation

## Caveats

None of the vendors publishes its internal architecture. Some figures come from aggregators and are
marked uncertain (DualEntry's valuation, Numeric's total funding, Rillet's price benchmark, Digits
with no round since 2022). The ERP Research review of Light is sponsored by Light.
