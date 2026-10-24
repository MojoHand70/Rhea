# flowMEE API — what the public documentation shows (fetched 2026-10-24)

Source: https://api.flowmee.automee.pl/doc — a Scalar page with an embedded OpenAPI 3.0 spec,
"FlowMEE API" 2.0.0, 114 operations, 147 schemas. Kept here because API documentation changes.
Everything below is their documentation, quoted or paraphrased; my readings are marked.

## Areas (tags)

Documents, Contractors, Agencies, Users, Organizations, Accounts (chart), Rules, Settlements
(bank lines), Tags, Sync (the desktop connector), Document Processing, VAT Registries, RMK,
Payroll Lists, Fixed Assets, Invoices (issuing), Leasing schedules, App releases, Webhooks.

## Documents

- Upload: multipart, or a signed upload to their storage (25 MB; Google Cloud Storage headers),
  then processing is queued. Sources: KSEF, USER, EMAIL, REMOTE, DRIVE, SALDEO, SALES.
- Status: PENDING → NEEDS_REVIEW → APPROVED / REJECTED → BOOKED → ARCHIVED; ERP sync PENDING /
  DONE / FAILED; failed documents listed and retried.
- Per document: KSeF number, dates (issue, sale, receipt, tax point, posting), net/VAT/gross,
  totals by rate, currency and the VAT rate's date, payment status, white-list verified, VIES,
  VAT status, cost and VAT accounts, MPK, project, journal, VAT register entries per rate
  (deduction 100 or 0, manual override), JPK procedure markings (WSTO_EE, TP, TT_WNT, MPP, …),
  JPK document markings (RO, WEW, FP, IMP, VAT_RR, MK), KSeF markings (NRKSEF, OFF, BFK, DI),
  VAT transaction type (domestic, import of goods, WNT, reverse charge, import of services, WDT,
  export…), KPiR entries, ryczałt summary by rate with a control sum, KUP/NKUP amounts, VAT
  deductible, annex-15 items, duplicate flag, "note to agent".
- Lines: description, cost type, VAT rate, net, GTU_01–13, cost account, Optima category,
  KUP rate, VAT deductible rate, ryczałt rate with its basis and reasons, MPK, project, parallel
  account and amount.
- `lightning_color` green / yellow / red.
- `ai_validation`: status positive / to explain / negative, with items each positive / to
  explain / negative / not applicable / not checked, a code and a message.
- `problem_items`: scope, code, message, severity (red, yellow, info, technical), action
  (choose, confirm, fix), a question, options A–D (label, value, recommended, reason, evidence,
  comment), facts, a **rule proposal** (text, ticked by default) and a **ledger fix** (Wn/Ma
  entries and a proposed change with its reason).
- `rebook`: "replays an already-booked document through the agent webhook… letting one
  document be replayed through a different agent (FlowForge staging, a DEV n8n project, a local
  agent)". *My reading:* documents are processed by an agent service, FlowForge, called over a
  webhook.
- `related-invoice-lookup`: finds the original invoice by number or KSeF number (corrections).

## Rules (`ClientRules`)

One object per company, replaced whole (PUT), deep-merged (PATCH) or patched by section.
- `company`: name, NIP, address, country, accounting type, ERP type, integration type (API,
  EXPORT_IMPORT, SYNCHRONIZER), business context, accounting summary, tax profile, VAT
  registered / payer, small taxpayer on the cash method, conditional VAT deduction, custom VAT
  deduction %, deduction by service end date, warehouse, VAT scheme, flat-rate %, accounting
  basis, CIS role and rate (UK), capitalisation threshold, currency rate source, accounting mode
  (STANDARD, ADVANCED, CUSTOM), ERP sync interval (5/15/30/60 min), description source for
  entries, posting date source (issue, sale or receipt date).
- `costs` sections: overview, journal selection, register selection, category selection,
  NKUP/KUP, VAT deductible, Estonian CIT expense classification, item type, project assignment,
  parallel accounting, contractor account, VAT account, VAT rate, route qualification, KPiR
  cost column, payroll accounts, accounts 0xx, 1xx, 2xx, 3xx, 4xx–7xx, 5xx, 6xx, 8xx, 9xx
  (Estonian CIT), advances; RMK threshold; **unresolved questions**; issues.
- `revenue` sections: journal, register, category, sale type, sale category, ryczałt rate,
  project, parallel accounting, contractor account, VAT account, KPiR revenue column, accounts
  7xx, advances; unresolved questions; issues.
- `settlements` sections: account assignment, categorisation, matching, document requirements,
  statement-required providers, gateway operators.
- `extraction`: document type classification.
- Each section: `rules` (a list of strings), `notes`, `updatedBy`. No version, no
  effective date, no approval state in the schema.

## Onboarding and learning

- Organization: `form_submitted`, `onboarding_started`, `provisioning_status` (queued, running,
  succeeded, failed), demo limits (records, documents, transactions remaining), region pl/uk,
  `live_results_enabled`, `snapshot_source_organization_id` with copy flags (chart, journals,
  VAT registers) — a new company set up as a copy of another.
- `agent-training`: "flowSYNC pushes an InsERT organization's posting history here and an
  agent-training run starts at once; the client later answers the agent's questions in Zasady
  księgowe." (JSON or ZIP up to 35 MB, format owned by FlowForge.)
- `company-profile-regeneration` webhook: writes business context, accounting summary, tax
  profile, industry, warehouse, compliance notes and sales structure into the rules; "flowMEE
  never overwrites the user's business context".
- AI checks with a score (NONE, POOR, PARTIAL, GOOD, EXCELLENT) and findings (INFO, WARNING,
  ERROR): approval-path coherence, approval-path conflicts, assignment-rule conflicts.

## Bank (`Settlement`)

Transaction date, title, details, amount, currency, advance, status (MATCHED, MANUAL_REVIEW,
PARTIALLY_MATCHED, MANUALLY_SETTLED, TO_CONFIRM), category (revenue, direct and operating
expenses, salaries, taxes and contributions, directors' loan, financing, assets, internal
transfers, special), counterparty, FX rate, statement number, IBAN, matched documents with
allocated amounts, splits, gateway payout, FX hint, decision (CONFIRMED, FLAGGED, SETTLED
MANUALLY, UNSETTLED), responsible person, tags, last reminder sent; Optima, InsERT, Xero and
QuickBooks transaction data.

## Issuing invoices

Invoice types FA and pro forma; KSeF types FA, FA_KOREKTA, FA_RR, FA_RR_KOREKTA, FA_ZAL,
FA_ROZ; KSeF channel and the reason when outside KSeF (no RR permission, KSeF unavailable,
technical impossibility, manual fallback); KSeF send status (not sent → queued → sending →
submitted → accepted / failed); VAT exemption legal basis (art. 113 ust. 1 and 9, art. 43
ust. 1 pkt …); next number; NBP rate for a date; revenue summary; "mark sent to client"
(UK only). Contractors carry the farmer flag and the RR permission status.

## Fixed assets, RMK, payroll, leasing

- Fixed assets: DRAFT → ACTIVE → SUSPENDED / FULLY_DEPRECIATED / DISPOSED; KŚT group and code;
  linear, degressive, one-off; useful life; depreciation accounts; non-deductible account and
  tax-deductible %; depreciation runs (dry run possible) that post; revisions (improvement,
  value correction, rate or method change, reclassification, audit adjustment) with a preview;
  disposal (sale, liquidation, donation, theft or loss, contribution in kind) with a preview of
  gain or loss; reversal of a posted entry; register with totals; summary by MPK, journal or
  project; idempotency key per invoice line.
- RMK: a proposal per source line (insurance, subscription, licence, domain and hosting,
  prepaid service, other), period from–to, ready / to verify; corrections; period documents
  sent to the ERP when an accountant clicks "send".
- Payroll lists: employer data, printed totals of every column, employees with pay components;
  "the printed totals are cross-checked against the sum of the employee rows… the mismatch
  FLAGS the document for review, it never rejects it".
- Leasing schedules: uploads processed by FlowForge.

## Integrations (ERP types)

OPTIMA, SYMFONIA, INSERT, INSERT_NEXO, INSERT_GT, NAVIREO, RAKS, ENOVA, KSIEGOWOSC_365,
STREAMSOFT_PRESTIZ, STREAMSOFT_VERTO. The connector pulls booked documents, RMK documents,
settlements and whole statements in the ERP's own format and reports back DONE / FAILED with
the ERP's document number. Their desktop apps are distributed through release channels
(App releases).
