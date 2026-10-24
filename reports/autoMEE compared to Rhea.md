# autoMEE (flowMEE) compared to Rhea

*2026-10-24. What autoMEE's product does, function by function against FUNCTIONS.md. Sources:
their site and price lists, every accounting post on their blog in 2026 (about 60 of 229 posts;
the rest are from their years as an AI agency), their terms, and their public API documentation
(114 operations, the most exact list of what the product does). Everything about flowMEE is
what they say in public; nothing was tried. Where I read something into their API rather than
their words, it is marked* my reading.

## In short

- **flowMEE does most of the work an office does on a document before it reaches the books, and
  sells it today.** Documents come in from KSeF, mail, a drive or their API. The AI reads each
  one, checks it (NIP, VAT status, white list, duplicates, sums) and proposes the booking line by
  line: account, MPK, project, KUP/NKUP, VAT deduction, VAT register, JPK markings, KPiR column,
  ryczałt rate. It asks a person only when unsure, then sends the entry to the office's program.
  Bank lines arrive live, are paired with invoices and split (Stripe, collective payments).
  There are approval paths, a collections module with an AI that phones debtors, and a live P&L.
- **It is not the books.** Optima, Symfonia, enova, InsERT and seven other programs keep the
  ledger, the VAT return, the JPK, the taxes and the close. Rhea is the books: those parts
  work in Rhea today (functions 36, 37, 49, 50, 57, 59, 64–68).
- **The real difference is who decides each document.** In flowMEE an AI agent decides every
  document. It is guided by the client's rules, which are plain sentences, and by the client's
  past bookings. In Rhea a rule decides, the same way every time. The model only drafts a rule
  where none exists, and a person approves it first. Each design has its price; see "How each
  one decides".
- **Where they are ahead is where FUNCTIONS.md says "not yet" or "kernel only":** documents
  coming in (22–24), bank (26–27, 29), the checks (77–81), dekretacja (82), issuing invoices
  (30–33), KPiR/ryczałt classification (42–46). And about fifteen functions FUNCTIONS.md does
  not name at all (list at the end).
- **They learn from the client's own past bookings on day one.** Rhea cannot do that today,
  because it needs the migration readers (73–75), which are not built.
- **Their onboarding is Rhea's Implementation by Interview, already selling.** First they read
  the client's past bookings. Then the agent asks the client what the history does not settle.
  The answers become rules in plain words. Before a contract they show, in numbers, how many
  documents would go through by themselves. Rhea adopted the same design on 2026-10-08, but the
  live interview is not built, and FUNCTIONS.md has no function for it. See "Onboarding".
- **They sell against rules by name.** Their post of 2026-10-23 says rule-based automation
  "starts generating its own work". New case → new rule → new exception → new fix. Rhea will
  be asked this question.
- *My reading of their API:* they are growing parts of a ledger of their own. They keep per
  document VAT register entries, journals, a chart of accounts by year, a fixed-asset register
  that posts depreciation, RMK schedules, sales invoices sent to KSeF, and a live P&L. They
  also sell to companies directly now, not only to offices. The line between "layer" and
  "books" is moving towards Rhea's ground.

## The company

| | |
|---|---|
| Who | autoMEE Sp. z o.o., Gdańsk; founded 2023 as an AI automation agency (chatbots, voice AI, recruitment — the blog until late 2025), accounting from 2026. CEO Dominik Laskowski; Chief of AI Kacper Tokarczyk; COO, CRO, customer success, business development (team page). A London entity, autoMEE Ltd, is listed on startup-seeker. |
| Money | Seed of about 10 mln zł (US$2.5–2.6M) led by Level2 Ventures, announced Sept 2026; Level2 had also invested in April 2024. |
| Users | 700+ companies through partner offices at the end of Sept 2026; ARR goal 7 mln zł by end of 2026 (Dealroom, xyz.pl). Named offices: KPFIG (106 client firms moved at once, their case study), Moore Polska, Business Assistant (testimonials). |
| Markets | Poland, and the UK in the API (a `region` of pl/uk; Xero and QuickBooks fields on bank lines; the UK construction scheme CIS and directors' loans; one endpoint "UK-only"). |
| Buyers | Accounting offices first; CFOs and finance teams of companies of 20+ people, groups, franchise networks, shared-service centres; now also small firms directly (starterMEE). |
| How they sell | Calculator of hours saved; a free session on the client's own documents and chart ("we show you in numbers how many would book themselves"); events in Gdańsk; partner programme (resell, refer, implement); referral 80 % off for 3 months; white-label for offices. |
| Liability | Capped at 6× the month's fee; "responsibility stays with the company and the person who approved the entry" (terms §11, dekretacja post). |

### Prices (public, net)

| Who | Price |
|---|---|
| Accounting office | **990 zł/month**: 10 clients included, then 5 zł per client; booking 100 pages included, then from 0.65 zł/page; bank lines 0.05 zł each; collections 0.50 zł per minute of calls; approval workflow unlimited; white-label 7,985 zł once. |
| Their calculator | 5 accountants, 100 clients, 10,000 documents a month → **13,615 zł/month**, about 1.36 zł a document all in. |
| Company — starterMEE | 985 zł/month: 5 users, approval workflow, AI sorting, kanban board, issuing invoices, KSeF unlimited. |
| Company — optiMEE | from 2,485 zł/month: 10 users, AI cost allocation, 350 bank lines (then 0.10 zł), booking 250 pages (then 1 zł/page). |
| Company — ultraMEE | quoted: live P&L, invoice checked against contract and goods receipt (PZ), voice collections. |

## Function by function

Numbers are FUNCTIONS.md's. "Rhea" is the status FUNCTIONS.md gives today.

### People, offices, setting up (A, B, C)

| | flowMEE | Rhea |
|---|---|---|
| Sign-in, people, roles | Users per company with roles; rights per company or location ("a branch manager sees only his branch"). | 1–9 work (keys, invitations, roles, substitutes, groups). |
| An office and its clients | An office ("agency") with many client companies; **one list of open questions across all clients**; each client its own chart and rules; a new client can be set up as a copy of another (chart, journals, VAT registers); white-label. | 10–13 work; 14 (the office's client list) not yet; 15 kernel only. No list of open questions across clients. |
| The company's profile | The AI writes the company's profile (business context, tax profile, industry, sales structure, warehouse, compliance notes) and keeps it with its rules; flags for VAT payer, small taxpayer on the cash method, Estonian CIT. | 16 through its generic form. |
| Chart of accounts | Read from the office's program, by year; MPK, projects, allocation keys, journals, VAT registers read too. | 21 partly (seen, not added; still shared). |
| KSeF connection | Yes (fetching and sending; "unlimited" in the cheapest plan). | 20 not yet. |

### Documents coming in (D) — they are ahead

| | flowMEE | Rhea |
|---|---|---|
| From KSeF | Fetched and booked; also as FA(3) files pushed through the API. | 22 kernel only (the reader exists, nothing fetches). |
| Upload, photo, scan | Upload in the app and through the API (25 MB); photos from a phone; receipts, including the simplified invoice up to 450 zł with the buyer's NIP. | 23 kernel only. |
| Mail | A dedicated address per company. | 24 not yet. |
| Other sources | A shared drive; Saldeo; documents in foreign languages ("most European languages"); leasing schedules; payroll lists (read and their totals cross-checked; payroll is not computed). | — |
| Bank | **Live bank feeds (open banking/PSD2)** plus statements; pairs payments with invoices by amount, counterparty, account and title; splits collective and partial payments; Stripe/PayPal/marketplace payouts split into sales and fees by a rule in a sentence; FX differences; a category, tags and a responsible person on each line; "99 % automatically" (their claim). Ready for the KSeF number in transfer titles from 1 Jan 2027. | 26 kernel only; 27 partly (one bank's detail lines); 28 works; 29 kernel only. |

### Checking a document (D2) — they are ahead

| | flowMEE | Rhea |
|---|---|---|
| 77 formal, 78 arithmetic | Net/VAT/gross checked against each other per rate; missing fields; each check recorded per item as positive / to explain / negative / not applicable / not checked. | not yet |
| 79 substantive | Approval by the person responsible for the budget is their substantive check; invoice against contract and PZ in the top plan. | not yet |
| 80 counterparty | NIP, VAT status, white list at intake; **a known supplier's bank account that changed stops the invoice**; VIES; duplicates, including the same receipt photographed twice. | partly (white list at onboarding) |
| 81 VAT deduction | Deductible share per line; reverse charge, import of services, WNT recognised. | not yet |
| 82 dekretacja | Per line: account, MPK, project, KUP/NKUP %, VAT deductible %, VAT register, JPK procedure and document markings, GTU, KPiR column, ryczałt rate with its reasons. A green/yellow/red light per document. Unsure → a question with options A–D, one recommended, with reason and evidence, and a proposed fix to the entry. The accountant's answer becomes a rule (the box is ticked by default). | not yet as such; supplier cost account (38) and RMK (39) work; C2, C5, C6 exist at one rate. |
| Corrections | Finds the original invoice even if booked before flowMEE, works out the difference line by line, books it like the original, proposes the VAT period, updates the balance. | 32 to verify |

### Documents going out (E)

| | flowMEE | Rhea |
|---|---|---|
| 30 issue an invoice | Yes: next number, lines, several rates, foreign currency at the NBP rate, exemption with its legal basis, pro forma; KSeF types FA, correction, RR (farmers), advance, final. | not yet on the desk (the kernel books an invoice from its fact). |
| 31 send to KSeF | Yes, with status (queued, sent, accepted, failed) and the reason when issued outside KSeF. | kernel only (KSeF's test system). |
| 32, 33 | Corrections yes; recurring invoices not seen. | to verify; not yet. |

### The books, taxes and the close (F–K) — Rhea is ahead, because flowMEE leaves them to the office's program

| | flowMEE | Rhea |
|---|---|---|
| Journal, trial balance, an entry's origin (34–37) | Kept in the office's program. flowMEE keeps journals and the entries it sends, and posts to cost-by-type and cost-by-function accounts at once. | 36, 37 work; 34, 35 not yet. |
| RMK (39) | Proposes the schedule from the invoice (insurance, subscription, licence…), sent to the program period by period. | works |
| Fixed assets (40) | A full register: draft → active → suspended → disposed; linear, degressive, one-off; depreciation runs that post; improvements and revaluations; a disposal preview with the gain or loss. | through the generic form, not yet tried. |
| KPiR, ryczałt (42–48) | Line by line: KPiR column and ryczałt rate; the book itself is kept in the program. PIT, ZUS: not seen. | not yet |
| VAT registers, settlement, JPK (49–55) | VAT register entries per rate on each document, JPK markings, GTU; the return and JPK are made and sent in the program. | 49, 50 work; 51 kernel only; 52–54 not yet. |
| Taxes (56, 57) | Not seen (no CIT, no tax calendar). | 56 kernel only; 57 works. |
| Month and year close (58–63) | "Continuous close" is their pitch, but no close and no lock of their own; no statements. | 59 works; 61, 63 through the generic form; 58, 60, 62 not yet. |

### Reports, questions, learning (L, M, N)

| | flowMEE | Rhea |
|---|---|---|
| Questions in words | "How much did we spend on marketing against budget?" — in the live P&L (top plan). | 64 works, every figure opening its source. |
| Live P&L, budgets, forecast | Yes (top plan): P&L as documents arrive, budgets with an alert before a limit is passed, MPK and project margins, an end-of-month forecast of missing recurring costs, report packs to the board. | — (68 receivables ageing works; no budgets, no forecast) |
| Why is this here (66) | "See why the document got this assignment": the rules applied, the analysis, the history of who approved. | works: the fact, the rule, its version, its legal basis. |
| Rhea's morning (67) | — | works |
| Learning (69–72, 83) | From the client's past bookings (uploaded at the start); every correction becomes a rule; rules visible and editable in plain Polish; an AI check of the rules and approval paths for conflicts and coherence, scored none to excellent. | 69 works; 70 kernel only; 71, 72, 83 not yet; rules versioned, with warrants, company vs Rhea rules. |
| Bringing data in (73–76) | Not needed: they read the office's program. | not yet |

### Integrations

Comarch ERP Optima, Symfonia, enova365, InsERT (Subiekt/Rewizor, nexo, GT, Navireo), RAKS,
WAPRO, Comarch Księgowość 360, Streamsoft Prestiż and Verto; export to wFirma and Asseco
Softlab; Saldeo as a source; Xero and QuickBooks (UK). Three ways in: an API, file
export/import, or a desktop synchroniser that pulls finished entries every 5–60 minutes and
reports back what the program did with each. A public REST API for software houses and
platforms: "send a document, get back the data, the booking proposal, the checks and the
exceptions". Billed per document.

## How each one decides

**flowMEE.** An AI agent reads each document and decides how to book it. Its guide is the
client's rules, about thirty sections of plain sentences: "journal selection", "KUP/NKUP",
"VAT deductible", "KPiR column", "accounts 4xx–7xx", "bank matching", "Stripe and other
gateways", "unresolved questions"…. Each section holds a list of sentences, notes and who
changed it last. Saving replaces the whole set; the API shows no versions, no date from
which a rule applies, and no approval step. The client's past bookings are uploaded at the
start ("the system reproduces your rules on earlier documents"). Every document is a model
call; each gets a confidence light; the sure ones go through by themselves.

**Rhea.** A rule decides. The rule is data with a version, a date it applies from and its
legal basis. It is drafted by the model only where no rule explains a fact, and it works only
after a person approves it. After that no model is involved: the same document is always
booked the same way, every figure names the rule and version that put it there, and the whole
history can be replayed from the log.

What follows from that, in practice:

| Question | flowMEE | Rhea |
|---|---|---|
| The same document twice | Probably the same answer; not guaranteed (a model decides each time; they have a "rebook" that runs it through the agent again). | The same answer, by construction (replay test). |
| A rule changes in June | The sentence is changed; whether May's bookings stay under May's rule is not visible from outside. | A new version from June; May stays booked under the old one. |
| The auditor asks why | The agent's reasons and the approval history. | The fact, the rule, its version, its legal basis; the period can be re-derived. |
| Two rules contradict each other | An AI checks the rule set and scores it. | (not checked in this report) |
| The first month at a new office | Strong: learns from the office's own history. | Weak today: starts from the pack's rules and learns what the pack lacks; history needs 73–75. |
| Cost of the ten-thousandth routine invoice | A model call (priced per page). | No model call once a rule exists (a PDF still needs one to be read). |
| Client data and the model | Every document, and the past bookings, go to the model; the provider is not named; files are stored on Google Cloud (from the API); the terms say nothing on where the AI runs. | KSeF XML and bank files need no model; a PDF does; a rule draft carries one sample; a local model is possible. |

Their 23 October post argues the case against rules: "the system has more and more
automation, and you, the accountant, still do a lot by hand". It describes a spiral: a new
case, then a new rule, then an exception, then a fix. *My suggestion:* Rhea's answer is that
Rhea writes and keeps the rules herself, and the accountant only approves them. The rules are
then the very audit trail that flowMEE's approach does not leave.

## Onboarding: their version of Implementation by Interview

KK, 2026-10-24: their onboarding survey "is exactly our Implement by interview". The evidence
supports that. Their step-by-step implementation post (2026-02-02) is a general article and
says nothing specific. The concrete evidence is in their product pages and their API:

1. **A form first.** A meeting booked on the site leads to "create a free account before the
   meeting". The API keeps on each company whether the onboarding form was submitted, whether
   onboarding started, and how the automatic setup went (queued, running, succeeded, failed).
   A free account has record limits. The form's questions are behind the login; I could not
   see them.
2. **The company described.** The client writes its business context. The AI then writes the
   rest of the profile: accounting summary, tax profile, industry, sales structure, warehouse,
   compliance notes. It regenerates them later and "never overwrites the user's business
   context". A mode is chosen: standard, advanced or custom rules.
3. **Evidence first.** They read the chart of accounts, the counterparties and the MPK from the
   office's program ("nothing is booked yet"). They load the client's past bookings, and "the
   model reproduces your rules: which supplier to which account, what is a tax cost, how you
   split costs to projects".
4. **Then the questions.** "An agent-training run starts at once; the client later answers the
   agent's questions in Zasady księgowe" (the API, on the past bookings). The rules keep a list
   of "unresolved questions" for costs and for revenue. The answers, in plain Polish, are the
   rules.
5. **A number before go-live.** "We take a sample of your invoices and your chart, and show in
   numbers how many would book themselves, with the list of those that would still need a
   person". Then a pilot on one document type or a few companies, measuring the share that
   goes through untouched. Afterwards the share is reported per company, "straight into an
   SLA report".
6. **People do much of it.** Their Customer Success team configures the first company with the
   client. Many answers on their pages say "we agree this at implementation".

Against Rhea's note (DIRECTION, "Implementation by interview", adopted 2026-10-08):

| | flowMEE | Rhea's design |
|---|---|---|
| Evidence first | The client's past bookings from its program, plus a sample of documents. | The client's last months of documents (KSeF, bank, chart) as facts in the log. |
| What the questions are about | What the history does not settle. | What the pack's rules leave unexplained, the largest cluster first. |
| Starting point | The client's own history (and a copy of another client's setup). | The pack's standard, with the law as its warrant. |
| An answer becomes | A sentence in the client's rules. | Versioned definitions, approved as one set, the conversation turn their cause in the log. |
| Done means | An automation share measured on a sample, then a pilot. | Nothing left unexplained on the intake corpus, counting what was knowingly set aside. |
| Who runs it | Their team with the client, the agent asking. | Rhea herself; a person approves. |
| Status | Selling. | Designed; sets of rules approved as a unit, agent drafts of such sets, and simulated clients through a scripted interview are built (8 of 8 runs done, 59 of 60 answers right on the first draft); the live interview with a real client is not. |

Two things they have that the note does not: **learning from the office's history** (Rhea
would need the migration readers, 73–75), and **the number shown before the contract**. That
number is Rhea's residue, used as a sales tool. *My reading:* the idea is not Rhea's alone any
more. What remains Rhea's is the governance: every answer becomes a versioned rule with its
warrant, and every implementation can be replayed.

## What flowMEE does that FUNCTIONS.md does not name

Candidates only, for you to decide on; not added to FUNCTIONS.md.

1. **Approval paths** for cost documents: who approves what by amount, category,
   department, supplier or project; several steps; substitutes; reminders and escalations;
   approval from the phone or straight from the notification; set in one sentence ("above
   10,000 zł to the CFO").
2. **MPK and projects**: cost centres and projects per line, split by allocation keys (area,
   headcount, share of revenue), several levels, changes from a date.
3. **Budgets** against MPK and projects, with an alert before the limit.
4. **One list of open questions across all of an office's clients.**
5. **Payment gateways and marketplaces**: Stripe, PayU, PayPal, Allegro, Amazon payouts split
   into sales, fees and returns.
6. **Live bank feeds** (open banking) beside statement files.
7. **Receipts from photos**, and the simplified invoice up to 450 zł.
8. **Foreign-language invoices**, with import of services, WNT and reverse charge recognised.
9. **A known supplier's bank account changed** → stop; **VIES** for EU counterparties.
10. **The KSeF number in transfer titles** from 1 Jan 2027, and the collective identifier.
11. **Payroll lists** read and their totals checked (payroll itself not computed).
12. **Collections**: reminders, SMS, and an AI that phones the debtor and records the
    promised date; stops when the money arrives; DSO reports.
13. **Live P&L with an end-of-month forecast** and report packs for the board.
14. **Invoice against contract and goods receipt (PZ)** — a three-way check.
15. **A public API** for software houses and platforms, billed per document; **white-label**
    for offices.
16. **Onboarding a new client**: Implementation by Interview itself. It is adopted in
    DIRECTION, but no function in FUNCTIONS.md says who runs it, where, or when it is done.
17. Smaller: GTU and JPK markings per line; RR, advance and final invoices; posting to cost by
    type and by function at once; Estonian CIT classification; costs marked for re-invoicing;
    intercompany documents marked at booking.

What they say they **do not** do: work out the VAT proportion, recognise revenue by stage of
completion, compute food cost, issue re-invoices themselves, qualify IP Box costs, compute
payroll. For each, they deliver the data and the office's program or adviser does the rest.

## What I could not verify

- Any accuracy figure. Their own numbers ("under 1 % errors", "99 % of bank lines", "90–95 % of
  documents") have no independent check; they promise to measure on the client's sample before
  a contract. Several older posts quote other firms' numbers as if general.
- Which model provider they use, and where it runs.
- Whether rule changes keep the earlier rule for earlier bookings (not visible in the API).
- How much of the UK region is live.
- The two reports they offer (offices 2026, companies 2026) are behind a form; not read.

## Sources

- Site: [automee.pl](https://automee.pl/), [Księgowość bez klikania](https://automee.pl/ksiegowosc-bez-klikania/), [for offices](https://automee.pl/biura-rachunkowe/), [Obieg AI](https://automee.pl/obieg-ai/), [Rozrachunki AI](https://automee.pl/rozrachunki-ai/), [Windykacja AI](https://automee.pl/windykacja-ai/), [RZiS na żywo](https://automee.pl/rzis-na-zywo/), [prices for offices](https://automee.pl/cennik-biura/), [prices for companies](https://automee.pl/cennik-firmy/), [partners](https://automee.pl/partnerzy/), [about](https://automee.pl/o-nas/), [terms](https://automee.pl/warunki-korzystania-z-uslugi/), [privacy](https://automee.pl/polityka-prywatnosci/), [referral](https://automee.pl/program-polecen-flowmee/), [events](https://automee.pl/eventy/), [meeting booked](https://automee.pl/umowione-spotkanie/), [implementation step by step, 2026-02-02](https://automee.pl/blog/ai-w-ksiegowosci-wdrozenie-ai-krok-po-kroku-jako-praktyczny-przewodnik-automatyzacji-procesow-ksiegowych/) — fetched 2026-10-24.
- API: [api.flowmee.automee.pl/doc](https://api.flowmee.automee.pl/doc) (OpenAPI 3, "FlowMEE API" 2.0.0), fetched 2026-10-24; summary in `research_notes/autoMEE compared to Rhea/flowmee-api.md`.
- Blog, 2026 accounting posts, among them: [API](https://automee.pl/blog/api-ksiegowe/), [automatic booking](https://automee.pl/blog/automatyczne-ksiegowanie-faktur/), [approval of cost invoices](https://automee.pl/blog/akceptacja-faktur-kosztowych/), [dekretacja](https://automee.pl/blog/dekretacja-faktur/), [MPK](https://automee.pl/blog/rozliczanie-kosztow-mpk/), [bank statements](https://automee.pl/blog/ksiegowanie-wyciagow-bankowych/), [corrections](https://automee.pl/blog/ksiegowanie-faktur-korygujacych/), [receipts](https://automee.pl/blog/ksiegowanie-paragonow/), [foreign invoices](https://automee.pl/blog/ksiegowanie-faktur-zagranicznych/), [white list](https://automee.pl/blog/weryfikacja-kontrahentow-biala-lista/), [KSeF cost invoices](https://automee.pl/blog/ksef-faktury-kosztowe/), [KSeF number in transfers](https://automee.pl/blog/ksef-rozrachunki-2027/), [integrations](https://automee.pl/blog/integracje/), [Optima](https://automee.pl/blog/integracje-comarch-erp-optima/), [enova365](https://automee.pl/blog/integracje-enova365/), [shared-service centres](https://automee.pl/blog/centrum-uslug-wspolnych/), [e-commerce payouts](https://automee.pl/blog/ksiegowosc-w-e-commerce-jak-rozliczac-wyplaty-zbiorcze/), the sector posts (medical, property managers, software houses, construction, hotels, networks, groups, transport), [AI vs rules, 2026-10-23](https://automee.pl/blog/ai-w-ksiegowosci-jak-sztuczna-inteligencja-zmienia-prace-ksiegowych-i-biur-rachunkowych/), [their ranking of AI tools for offices](https://automee.pl/blog/najlepsze-narzedzia-ai-dla-biur-rachunkowych-w-polsce-w-2026-roku-ranking-i-porownanie/), [KPFIG case](https://automee.pl/blog/ksiegowosc-ai-z-flowmee-i-ksef-jak-polskie-biura-rachunkowe-realnie-automatyzuja-obsluge-klientow/).
- Press: [Dealroom, seed](https://dealroom.co/news/157853-automee-raises-2-6m-seed-to-bring-real-time-accounting-to-polands-smes/), [xyz.pl, 10 mln zł](https://xyz.pl/?p=255867), [startup-seeker, autoMEE Ltd](https://startup-seeker.com/company/automee~uk).
- Rhea: [FUNCTIONS.md](../FUNCTIONS.md), [packs/pl/pack.json](../packs/pl/pack.json) (41 object types, 38 views, 138 rules, 39 actions).
