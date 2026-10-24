# Model Polish chart of accounts — strand 1: the statements and their positions

Research note, 2026-10-25. Input for building a model chart of accounts (*zakładowy plan kont*) that produces every Polish GAAP financial statement completely from the books.

**Scope of this strand.** The financial statements under the Accounting Act (*ustawa o rachunkowości*, UoR) and the National Accounting Standards (KSR): every position, its official numbering, its e-sprawozdanie XML element, which account groups feed it and which analytics are needed to fill it without manual work. **Out of scope** (other strands): (a) tax and statistical filings from the books (CIT-8, JPK_KR_PD tags, GUS forms); (b) model charts used in practice. Banks, insurers, investment funds and NGOs (annexes 2, 3, 6) are mentioned only where they bound the scope.

**Legal state used.** UoR consolidated text [Dz.U. 2026 poz. 522](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260000522) (legal state 26 March 2026; [PDF via ELI](https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf)) plus the three later amendments listed in §1.1. Statement position names below are quoted from the annexes in Polish and were checked against the e-sprawozdanie XSDs, which carry the same text word for word.

**Conventions used in the tables.**

| Abbreviation | Meaning |
|---|---|
| **CP** | counterparty class, the statements' three-way split: **P** = *jednostka powiązana* (same capital group, art. 3 ust. 1 pkt 43–44); **Z** = *pozostała jednostka, w której jednostka posiada zaangażowanie w kapitale* (durable equity stake, art. 3 ust. 1 pkt 37d — associates, joint ventures and other durable holdings outside the group); **O** = *pozostała jednostka* |
| **DUE** | due date kept per open item (invoice, instalment), so maturity buckets come from data, not from judgement |
| **CUR** | currency of the item |
| **MOV** | movement-type tag on a posting (e.g. acquisition, disposal, revaluation, internal transfer; issue, redemption; created, used, released) |
| **CF** | cash-flow line tag on a posting to a cash or bank account |
| **LT / ST** | long-term / short-term split measured from the balance-sheet date |
| **Groups 0–8** | account groups (*zespoły kont*) of Polish practice: 0 aktywa trwałe; 1 środki pieniężne, inwestycje krótkoterminowe, kredyty; 2 rozrachunki i roszczenia; 3 materiały i towary; 4 koszty według rodzajów; 5 koszty według typów działalności (funkcjonalne); 6 produkty i rozliczenia międzyokresowe; 7 przychody i koszty ich osiągnięcia; 8 kapitały, fundusze specjalne, rezerwy, RMP i wynik finansowy. The law prescribes **no** model chart for business entities (art. 10 ust. 1 pkt 3 lit. a UoR makes every entity set its own *zakładowy plan kont*); the 0–8 grouping is a practice convention, see e.g. [fakturaxl.pl](https://www.fakturaxl.pl/plan-kont-w-pelnej-ksiegowosci). Account numbering itself is strand (b)'s subject. |

---

## 0. Key facts in one page

| Fact | Value | Source |
|---|---|---|
| Consolidated UoR | Dz.U. 2026 poz. 522, later amended by Dz.U. 2026 poz. 640, 644, 1206 | [ELI metadata](https://api.sejm.gov.pl/eli/acts/DU/2026/522) |
| **Micro** (*jednostka mikro*) | does not exceed 2 of 3, in the year and the preceding year: **2 000 000 zł** total assets, **4 000 000 zł** net revenue, **10** FTE | art. 3 ust. 1 pkt 1a UoR |
| **Small** (*jednostka mała*) | not micro and does not exceed 2 of 3: **33 000 000 zł**, **66 000 000 zł**, **50** FTE | art. 3 ust. 1 pkt 1b |
| **Medium** (*jednostka średnia*) | neither micro nor small and does not exceed 2 of 3: **110 000 000 zł**, **220 000 000 zł**, **250** FTE | art. 3 ust. 1 pkt 1c |
| **Large** (*jednostka duża*) | exceeds 2 of the 3 medium limits (110 M / 220 M / 250) | art. 3 ust. 1 pkt 1d |
| Size limits apply from | statements for financial years starting after 31 Dec 2023 (i.e. FY 2024 onward) | art. 14 ust. 3 of the [act of 6 Dec 2024, Dz.U. 2024 poz. 1863](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20240001863); [MF Q&A](https://www.gov.pl/attachment/f9b05862-99e1-473c-95e4-c1d1c21301cb) |
| Always treated as large (for the simplifications and for CF/SCE) | entities under art. 2 ust. 1 pkt 3 (banks, insurers, investment firms, crowdfunding providers, investment and pension funds, SKOKs) and those applying for such licences, ASIs, issuers (or applicants) on an EEA regulated market, issuers on an ATS, national payment institutions, e-money institutions | art. 3 ust. 1h (as amended by [Dz.U. 2026 poz. 644](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260000644)) |
| **Cash-flow statement and statement of changes in equity required for** | every entity **other than micro and small** (and other than art. 80 ust. 4 organisations), plus public-finance-sector units meeting 2 of the art. 64 ust. 1 pkt 4 conditions. **No longer tied to the audit duty.** | art. 45 ust. 3 UoR; [MF Q&A, webinar 15 Jan 2025](https://www.gov.pl/attachment/f9b05862-99e1-473c-95e4-c1d1c21301cb) |
| **Audit duty** (other entities) | S.A. always; others meeting 2 of 3 in the **preceding** year: **≥ 50** FTE, total assets **≥ 3 125 000 EUR**, net revenue **≥ 6 250 000 EUR** | art. 64 ust. 1 pkt 3–4 |
| Audit limits in złoty for FY 2026 (calendar year, tested on FY 2025) | 3 125 000 EUR × 4,2267 = **13 208 437,50 zł**; 6 250 000 EUR × 4,2267 = **26 416 875,00 zł** (NBP mid-rate of 31 Dec 2025, table 251/A/NBP/2025). For FY 2025 (tested on FY 2024, rate 4,2730 of 31 Dec 2024): 13 353 125,00 zł and 26 706 250,00 zł | art. 3 ust. 2 UoR; [NBP API](https://api.nbp.pl/api/exchangerates/rates/a/eur/2025-12-31/?format=json); same figures in [Infor](https://ksiegowosc.infor.pl/rachunkowosc/sprawozdawczosc/6577736,zmiany-ustawy-o-rachunkowosci-podniesienie-o-25-progow-okreslajacyc.html) |
| Books duty for natural persons and partnerships (for orientation) | net revenue ≥ 2 500 000 EUR in the preceding year; for 2026: × 4,2586 (NBP, 1 Oct 2025) = **10 646 500 zł** | art. 2 ust. 1 pkt 2 and art. 3 ust. 3 UoR; [NBP API](https://api.nbp.pl/api/exchangerates/rates/a/eur/2025-10-01/?format=json) |
| Sale of materials | since FY 2024 statements: **other operating** income/cost, not revenue | art. 3 ust. 1 pkt 30a and pkt 32 lit. b; [MF](https://www.gov.pl/web/finanse/jak-zaprezentowac-w-rachunku-zyskow-i-strat-sprzedaz-materialow-po-nowelizacji-ustawy-o-rachunkowosci-ustawa-z-dnia-6-grudnia-2024-r); KSR 15 (2026) pkt 8.1 |
| e-sprawozdanie structures in force | CRWDE forms **SFJINZ(2)** / SFJINT(2) (*jednostka inna*), **SFJMAZ(2)** / SFJMAT(2) (*mała*), **SFJMIZ(2)** / SFJMIT(2) (*mikro*); `wersjaSchemy` **1-0E**, `WariantSprawozdania` **2**; domain schemas `…StrukturyDanychSprFin_v2-0E.xsd` (namespace dated 2025/07/31) | [gov.pl/web/kas/struktury-e-sprawozdan](https://www.gov.pl/web/kas/struktury-e-sprawozdan); XSDs at crd.gov.pl (§2) |
| They apply to | financial years starting **1 Jan 2025 or later**; KAS application accepts old and new structures until **30 Jun 2026**, then only the new | [MF notice 27 Nov 2025](https://podatki-arch.mf.gov.pl/komunikaty-techniczne/nowe-struktury-logiczne-dla-sprawozdan-finansowych-271125/?altTemplate=ArticlePdf&download=True); [MF notice 21 Jan 2026, upd. 26 May 2026](https://www.podatki.gov.pl/aktualnosci/komunikat-w-sprawie-nowych-struktur-logicznych-dla-sprawozdan-finansowych) |
| Amount type in the XSD | `TKwota3`: decimal, 18 digits total, 2 after the point — matches Rhea's `DECIMAL(18,2)` | `StrukturyDanychSprFin_v2-0E.xsd` |

---

## 1. Legal frame

### 1.1 The text and its amendments

| Act | What it does to the statements | In force / applies | Source |
|---|---|---|---|
| Consolidated text, Marshal's notice of 30 Mar 2026 | base text used here (legal state 26 Mar 2026) | — | [Dz.U. 2026 poz. 522](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260000522) |
| Act of 6 Dec 2024 (CSRD + thresholds) | new size definitions art. 3 ust. 1 pkt 1a–1e (thresholds +25 %); CF/SCE tied to size (art. 45 ust. 3); new annexes 1, 4, 5 texts (materials out of revenue); definition of net revenue (pkt 30a); audit limits in EUR raised to 3 125 000 / 6 250 000; books-duty limit 2 500 000 EUR | entered 1 Jan 2025; size/annex changes apply to statements for FYs starting after 31 Dec 2023; audit limits and books-duty limit to FYs starting after 31 Dec 2024 (art. 14 ust. 1–3) | [Dz.U. 2024 poz. 1863](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20240001863) |
| Act of 6 Nov 2024 (global minimum tax) | art. 37 ust. 12 (no deferred tax on top-up taxes), annex 1 notes 1.20 and 2.13, art. 45 ust. 1ba | 1 Jan 2025 | [Dz.U. 2024 poz. 1685](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20240001685) |
| Act of 27 Feb 2026 | art. 84a: CSRD "stop-the-clock" for FYs 2025–2026 (≤ 1000 FTE or ≤ 1 900 000 000 zł revenue) — no effect on the financial statements themselves | 14 Mar 2026 | [Dz.U. 2026 poz. 333](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260000333) |
| Act of 27 Mar 2026 | art. 66 ust. 6 pkt 1 only (audit firm rules) | 28 May 2026 | [Dz.U. 2026 poz. 640](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260000640) |
| Act of 17 Apr 2026 (ESAP) | art. 3 ust. 1h now also covers art. 45 ust. 3 pkt 1 and art. 49 ust. 4–6; art. 45 ust. 3 pkt 3 (ASI) repealed (ASIs now covered via ust. 1h); new chapter 6d (ESAP metadata, applies from FYs starting after 31 Dec 2026, provisions enter 10 Jan 2028) | 29 May 2026 (chapter 6d: 10 Jan 2028) | [Dz.U. 2026 poz. 644](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260000644), art. 18–19 |
| Act of 31 Jul 2026 (banking) | wording "oddziały z państw trzecich" in art. 3 ust. 1 pkt 30a, art. 64, art. 69 — banks only | 29 Sep 2026 | [Dz.U. 2026 poz. 1206](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260001206) |

No bill changing the size or audit thresholds or the annexes was found pending as of this note (*to verify* on [RCL](https://legislacja.gov.pl) before FY 2027).

### 1.2 What each entity must prepare

Art. 45 ust. 2 UoR: every financial statement contains *bilans*, *rachunek zysków i strat* and *informacja dodatkowa* (introduction plus additional notes). Art. 45 ust. 3 adds *zestawienie zmian w kapitale (funduszu) własnym* and *rachunek przepływów pieniężnych* for the entities listed there. The simplified annexes need a decision of the approving body (*organ zatwierdzający*, art. 49c ust. 1); a micro entity may instead use the small-entity simplifications (art. 49c ust. 2). Public-finance-sector units and NGOs under the public-benefit act may not use art. 49c (ust. 3).

| Entity | Bilans | RZiS | Informacja dodatkowa | Zest. zmian w kapitale | Rach. przepływów | Sprawozdanie z działalności (if art. 49 ust. 1 entity) |
|---|---|---|---|---|---|---|
| **Micro**, simplifications decided | annex 4 | annex 4 (one layout) | may be replaced by *informacje uzupełniające do bilansu*, annex 4 (art. 48 ust. 3) | no | no | may be skipped if own-share data given (art. 49 ust. 4) |
| **Micro**, no annex 4 decision | annex 1 (or annex 5, again by decision, art. 49c ust. 2) | annex 1 / annex 5, variant of choice | annex 1 / annex 5 | no | no | yes |
| **Small**, simplifications decided | annex 5 | annex 5, comparative or by-function (art. 47 ust. 4 pkt 5) | annex 5 | no | no | may be skipped if own-share data in the notes (art. 49 ust. 5) |
| **Small**, no decision | annex 1 | annex 1, variant of choice | annex 1, **at least** the annex 5 scope (art. 48 ust. 4) | no | no | yes |
| **Medium, large**, and every art. 3 ust. 1h entity | annex 1 | annex 1, variant of choice (art. 47 ust. 4 pkt 1) | annex 1 | **yes**, annex 1 | **yes**, direct or indirect method (art. 48b ust. 1) | yes |
| Public-finance-sector unit meeting 2 of the art. 64 ust. 1 pkt 4 conditions | per its rules | per its rules | per its rules | **yes** | **yes** | per its rules |

The MF confirmed in writing that CF/SCE now follow size, not audit: an sp. z o.o. that is neither micro nor small prepares them even when not audited ([MF Q&A, webinar 15 Jan 2025](https://www.gov.pl/attachment/f9b05862-99e1-473c-95e4-c1d1c21301cb)). The converse follows from the text: an audited **small** entity is not obliged to prepare them (art. 45 ust. 3 pkt 1) — *to verify* that no other act (e.g. KSH for S.A.) adds them; none was found.

Comparatives: every statement shows the current and the preceding year (art. 46 ust. 1, art. 47 ust. 1, art. 48a ust. 1, art. 48b ust. 1); a position empty in both years is omitted (art. 50 ust. 3); more detail than the annexes is allowed (art. 50 ust. 1).

### 1.3 Micro and small simplifications that change which accounts exist

| Simplification | Who may use it | Excluded | Effect on the chart | Source |
|---|---|---|---|---|
| Leases classified by tax rules (no finance-lease asset/liability at the lessee unless tax treats it so) | micro, small | — | lease accounts follow tax classification | art. 3 ust. 6 |
| Waive prudence for impairment allowances and provisions | micro, small | **capital companies, SKA, sp.j./sp.k. whose all unlimited partners are capital companies, public-finance units** | allowance/provision accounts may stay empty | art. 7 ust. 2b–2c |
| Indirect production costs in product cost regardless of capacity | micro, small | — | no unabsorbed-overhead split needed | art. 28 ust. 4a |
| No fair value, no amortised cost, no financial-instruments regulation | micro using annex 4 or the art. 49 ust. 4 exemption (mandatory, art. 28a); other micro and all small may skip the instruments regulation (optional, art. 28b) | — | no FV revaluation accounts (art. 28a); no instrument-category analytics (art. 28a/28b) | art. 28a, 28b |
| Tax depreciation of fixed assets | micro, small | **same list as art. 7 ust. 2c** | no book/tax depreciation difference | art. 32 ust. 7–8 |
| No deferred tax | micro, small | — | no deferred-tax asset/provision accounts | art. 37 ust. 10 |
| No accruals for future employee benefits (retirement etc.) | micro, small | **same list as art. 7 ust. 2c** | no retirement-provision account | art. 39 ust. 6–7 |

These record-keeping simplifications are decided by the head of the entity and do not depend on using the simplified statement annexes ([MF Q&A](https://www.gov.pl/attachment/f9b05862-99e1-473c-95e4-c1d1c21301cb)). KSR 16 *Uproszczenia dopuszczone ustawą o rachunkowości* (2025) explains all of them and applies from FY 2026 ([KSR page](https://www.gov.pl/web/finanse/krajowe-standardy-rachunkowosci)). **Consequence for Rhea:** for the first users (sp. z o.o. and JDG), the model chart must keep provision, allowance and retirement-accrual accounts available for every capital company, because the law does not let them opt out.

### 1.4 Rules in the Act that shape the chart beyond the annexes

| Rule | What it requires of the books | Source |
|---|---|---|
| *Zakładowy plan kont* is part of the written accounting policy, with the subsidiary-ledger links | the chart itself is a policy document | art. 10 ust. 1 pkt 3 lit. a |
| Subsidiary ledgers (*konta ksiąg pomocniczych*) at least for: fixed assets, WNiP and their depreciation; counterparties; employees (named payroll records); sales; purchases; costs and material assets; cash | analytic ledgers reconciled with the general ledger | art. 16–17 |
| Trial balance (*zestawienie obrotów i sald*) at least monthly; balances of all subsidiary ledgers at least at year end | period close produces both | art. 18 |
| No offsetting of different assets/liabilities, revenues/costs | debit and credit balances of settlement accounts (VAT, ZUS, counterparties) must be split in the balance sheet | art. 7 ust. 3 |
| Trade receivables and trade payables are **always short-term**, whatever the due date; other items short-term if due within 12 months | maturity from DUE, trade vs non-trade flag | art. 3 ust. 1 pkt 18 lit. c, pkt 22 |
| Cash equivalents for the cash-flow statement: monetary assets due within 3 months | deposit term at placement | art. 3 ust. 1 pkt 25; KSR 1 pkt 3.1, 4.4 |
| Net revenue = sales of goods and products incl. subsidies, rebates, net of VAT and other turnover taxes | revenue accounts net; price subsidies inside revenue | art. 3 ust. 1 pkt 30a |
| Other operating income/costs list (social activity, disposal of materials and fixed assets, investment property, write-offs, provisions, impairment, compensation and penalties, donations, random events) | each of these needs its own account or analytic in group 7 | art. 3 ust. 1 pkt 32 |
| Impairment of receivables → other operating or financial costs by nature of the receivable; reversals → other operating or financial income | allowance account analytics by receivable type | art. 35b ust. 2, art. 35c |
| Provisions → other operating costs or financial costs by circumstance; releases likewise | provision analytics by nature | art. 35d ust. 2 and 4 |
| FX differences on non-investment items → financial income/costs, or capitalised into inventory or fixed assets "in justified cases" | FX analytics: realised/unrealised, item type, capitalised flag | art. 30 ust. 4 |
| Equity booked by kind; S.A. share-issue costs reduce agio, the rest to financial costs; dopłaty (sp. z o.o.) as a separate line *kapitał rezerwowy z dopłat wspólników* and unpaid ones as *Należne dopłaty na poczet kapitału rezerwowego (wielkość ujemna)* | separate equity accounts beyond the annex lines | art. 36 ust. 1, 2b, 2e |
| Own shares sold/redeemed: differences to kapitał zapasowy, excess to prior-year loss | equity movement tags | art. 36a |
| In-year profit charges shown negative in equity | separate equity account | art. 46 ust. 3 |
| ZFŚS and other special funds not in equity are liabilities (*fundusze specjalne*) | group 8 special-fund accounts mapped to liabilities | art. 46 ust. 4 |
| Discontinued activities shown separately | flag on revenue and cost postings | art. 47 ust. 3 |
| Deferred tax: provision and asset separately (offset only with a legal right); P&L tax = current + deferred; equity-related deferred tax to equity | separate accounts for current and deferred tax and for each side | art. 37 ust. 1–9 |
| Accruals (RMK czynne/bierne); employee-benefit accruals shown as provisions | RMK analytics by nature and term | art. 39 ust. 1–3, 2a |
| Deferred income (RMP), incl. grants for fixed assets released in step with depreciation | RMP analytics by type | art. 41 |
| Financial instruments: five categories, LT/ST split, disclosures by category | category attribute on every financial asset/liability | [Rozp. MF 17 Nov 2024, Dz.U. 2024 poz. 1750](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20240001750) § 4, § 39, § 45 (applies from FYs starting after 31 Dec 2024, § 47); replaces Dz.U. 2017 poz. 277 |

---

## 2. The e-sprawozdanie XML structures

### 2.1 Legal basis and where they are published

Art. 45 ust. 1f UoR: the statement is prepared electronically and signed (qualified signature, *podpis zaufany* or *podpis osobisty*). Art. 45 ust. 1g: entities in the KRS register of entrepreneurs prepare it "w strukturze logicznej oraz formacie udostępnianych w Biuletynie Informacji Publicznej" of the Ministry of Finance. IFRS statements use ESEF/another searchable format (art. 45 ust. 1ga–1h); the MF does not publish structures for them ([gov.pl/web/kas/struktury-e-sprawozdan](https://www.gov.pl/web/kas/struktury-e-sprawozdan)).

The current structures are published as forms (*wzory*) in the CRWDE (Centralne Repozytorium Wzorów Dokumentów Elektronicznych); the gov.pl page lists them with the note "wzór stosuje się do sprawozdań finansowych sporządzonych od 01.01.2026" ([gov.pl/web/kas/struktury-e-sprawozdan](https://www.gov.pl/web/kas/struktury-e-sprawozdan), page version 27.0 of 3 Aug 2026). The MF notice of 27 Nov 2025 states the scope differently: the new forms "mają zastosowanie do sprawozdań finansowych sporządzanych za lata obrotowe które rozpoczęły się 1 stycznia 2025 r. albo później", the new versions are recommended for earlier years too, and the old structures are temporarily still allowed ([MF notice 27 Nov 2025](https://podatki-arch.mf.gov.pl/komunikaty-techniczne/nowe-struktury-logiczne-dla-sprawozdan-finansowych-271125/?altTemplate=ArticlePdf&download=True)). The later MF notice (21 Jan 2026, updated 26 May 2026) says the KAS e-Sprawozdania application would support the new structures by the end of Q1 2026 and accept both until 30 Jun 2026, then only the new ones; KRS filings go through the Ministry of Justice's *Repozytorium Dokumentów Finansowych* ([podatki.gov.pl](https://www.podatki.gov.pl/aktualnosci/komunikat-w-sprawie-nowych-struktur-logicznych-dla-sprawozdan-finansowych)).

History, for orientation: v1-2 files named e.g. `JednostkaInnaWZlotych(1)_v1-2.xsd` were published in Nov 2024 ([pit.pl, 12 Nov 2024](https://www.pit.pl/aktualnosci/ministerstwo-finansow-wprowadzamy-nowe-struktury-logiczne-dla-sprawozdan-finansowych-od-2025-r-1010431)); a version 1_3 for financial years from 1 Jan 2024 followed in Nov 2024 ([EY](https://www.ey.com/pl_pl/insights/tax/e-sprawozdania-finansowe-2025-wybor-wlasciwej-struktury-logicznej-dla-jednostki-innej)). The "wariant 2" CRWDE forms described below replace them. *To verify:* the exact older file names, since the gov.pl page no longer lists them.

### 2.2 The three structures that matter for a non-financial company

All downloaded and parsed for this note (XSDs are public, no login).

| Form (CRWDE) | Code on gov.pl | `KodSprawozdania` / `kodSystemowy` / `wersjaSchemy` / `WariantSprawozdania` | Full title | Main schema | Domain schema(s) imported |
|---|---|---|---|---|---|
| 2025/08/07/13817 | SFJINZ(2) | `SprFinJednostkaInnaWZlotych` / `SFJINZ (2)` / `1-0E` / `2` | SPRAWOZDANIE FINANSOWE DLA JEDNOSTEK INNYCH NIŻ BANKI, ZAKŁADY UBEZPIECZEŃ I ZAKŁADY REASEKURACJI (W ZŁOTYCH) | [schemat.xsd](https://crd.gov.pl/wzor/2025/08/07/13817/schemat.xsd) | [JednostkaInnaStrukturyDanychSprFin_v2-0E.xsd](http://crd.gov.pl/xml/schematy/dziedzinowe/mf/2025/07/31/eD/JednostkaInnaStruktury/JednostkaInnaStrukturyDanychSprFin_v2-0E.xsd), [StrukturyDanychSprFin_v2-0E.xsd](http://crd.gov.pl/xml/schematy/dziedzinowe/mf/2025/07/31/eD/StrukturyDanychSprFin/StrukturyDanychSprFin_v2-0E.xsd) |
| 2025/08/04/13810 | SFJINT(2) | `SprFinJednostkaInnaWTysiacach` / `SFJINT (2)` / `1-0E` / `2` | same, W TYSIĄCACH | [schemat.xsd](https://crd.gov.pl/wzor/2025/08/04/13810/schemat.xsd) | same |
| 2025/08/07/13818 | SFJMAZ(2) | `SprFinJednostkaMalaWZlotych` / `SFJMAZ (2)` / `1-0E` / `2` | SPRAWOZDANIE FINANSOWE DLA JEDNOSTEK MAŁYCH KORZYSTAJĄCYCH Z UPROSZCZEŃ ODNOSZĄCYCH SIĘ DO SPRAWOZDANIA FINANSOWEGO (W ZŁOTYCH) | [schemat.xsd](https://crd.gov.pl/wzor/2025/08/07/13818/schemat.xsd) | + [JednostkaMalaStrukturyDanychSprFin_v2-0E.xsd](http://crd.gov.pl/xml/schematy/dziedzinowe/mf/2025/07/31/eD/JednostkaMalaStruktury/JednostkaMalaStrukturyDanychSprFin_v2-0E.xsd) |
| 2025/09/10/13846 | SFJMAT(2) | `SprFinJednostkaMalaWTysiacach` / `SFJMAT (2)` / `1-0E` / `2` | same, W TYSIĄCACH | [schemat.xsd](https://crd.gov.pl/wzor/2025/09/10/13846/schemat.xsd) | same |
| 2025/08/12/13821 | SFJMIZ(2) | `SprFinJednostkaMikroWZlotych` / `SFJMIZ (2)` / `1-0E` / `2` | SPRAWOZDANIE FINANSOWE DLA JEDNOSTEK MIKRO ORAZ JEDNOSTEK, O KTÓRYCH MOWA W ART. 80 UST. 4 USTAWY O RACHUNKOWOŚCI (W ZŁOTYCH) | [schemat.xsd](https://crd.gov.pl/wzor/2025/08/12/13821/schemat.xsd) | + [JednostkaMikroStrukturyDanychSprFin_v2-0E.xsd](http://crd.gov.pl/xml/schematy/dziedzinowe/mf/2025/07/31/eD/JednostkaMikroStruktury/JednostkaMikroStrukturyDanychSprFin_v2-0E.xsd) |
| 2025/08/11/13819 | SFJMIT(2) | `SprFinJednostkaMikroWTysiacach` / `SFJMIT (2)` / `1-0E` / `2` | same, W TYSIĄCACH | [schemat.xsd](https://crd.gov.pl/wzor/2025/08/11/13819/schemat.xsd) | same |

Other forms on the same page (not parsed): SFSINZ(2)/SFSINT(2) consolidated statements of non-financial groups; SFJOPZ(2)/SFJOPT(2) NGOs (annex 6); SFASIZ(2)/SFASIT(2) ASIs (republished Feb 2026 as 2026/02/10/14162 and 2026/02/13/14163); banks, insurers, brokerage houses, SKOKs, funds, issuers (SFE(2)), and SOI(2) for organisations reporting under other law ([gov.pl](https://www.gov.pl/web/kas/struktury-e-sprawozdan)).

Which form to file: by their titles, the *Mała* form is for small entities **using** simplifications and the *Mikro* form for micro entities and art. 80 ust. 4 organisations; other entities use *Inna*. (*To verify:* which form a micro or small entity that uses **no** simplification should file — both the *Mała* and *Mikro* forms also accept annex 1 statements.) The *Mała* and *Mikro* forms let each statement be chosen separately — the XSD offers `BilansJednostkaMala` **or** `BilansJednostkaInna`, `RZiSJednostkaMala` **or** `RZiSJednostkaInna`, and for micro also the annex 5 forms — so a partial use of simplifications is expressible (schemat.xsd of 13818 and 13821).

### 2.3 Document tree (JednostkaInna, SFJINZ(2))

```
Dokument                              CRWDE envelope (struktura.xsd 2009/11/16)
 ├─ OpisDokumentu
 ├─ DaneDokumentu
 └─ TrescDokumentu
   └─ JednostkaInna
     ├─ Naglowek                      OkresOd, OkresDo, DataSporzadzenia, KodSprawozdania, WariantSprawozdania (=2)
     ├─ WprowadzenieDoSprawozdaniaFinansowego
     │    P_1  Dane identyfikujące jednostkę: P_1A firma i siedziba, P_1B forma prawna (enumeracja TFormaPrawna),
     │         P_1C PKD, P_1D NIP, P_1E KRS
     │    P_2  czas trwania działalności (if limited)      P_3  okres sprawozdania
     │    P_4  dane łączne (1/2)                            P_5  kontynuacja działalności: P_5A, P_5B, P_5C opis zagrożeń
     │    P_6  połączenie spółek: P_6A, P_6B metoda        P_7  zasady (polityka) rachunkowości: P_7A wycena, P_7B wynik,
     │                                                          P_7C sposób sporządzenia, P_7D pozostałe (free text)
     │    P_8  średnioroczne zatrudnienie w etatach (art. 64 ust. 1 pkt 4)   — required number
     │    P_9  podlega obowiązkowi badania (1/2)                                — required flag
     │    P_10 pozycje użytkownika
     ├─ Bilans                        BilansJednostkaInna            (required)
     ├─ RZiS                          RZiSJednostkaInna: RZiSKalk | RZiSPor   (required, one of two)
     ├─ ZestZmianWKapitale            ZestZmianWKapitaleJednostkaInna (optional in XSD; required by art. 45 ust. 3)
     ├─ RachPrzeplywow                RachPrzeplywowJednostkaInna: PrzeplywyBezp | PrzeplywyPosr (optional in XSD; required by art. 45 ust. 3)
     └─ DodatkoweInformacjeIObjasnieniaJednostkaInna
          DodatkoweInformacjeIObjasnienia*            TInformacjaDodatkowa = Opis (text) + Plik? (file, base64)
          InformacjaDodatkowaDotyczacaPodatkuDochodowego?   structured CIT reconciliation (see 2.5)
```

The *Mała* form adds `P_4` "Wskazanie zastosowanych uproszczeń przewidzianych dla jednostek małych" (or `P_11` when the annex 1 introduction is chosen); the *Mikro* form has `InformacjeOgolneJednostkaMikro` with `P_4` "Wskazanie zastosowanych zasad rachunkowości przewidzianych dla jednostek mikro z wyszczególnieniem wybranych uproszczeń" and `InformacjeUzupelniajaceDoBilansuJednostkaMikro` in place of the notes (schemat.xsd of 13818, 13821). The introduction's P_8 (average FTE) and P_9 (audit flag) are mandatory data that do not come from the ledger: Rhea needs them from HR and from the audit test.

### 2.4 Amount types — what every position carries

| Type | Content | Consequence for Rhea |
|---|---|---|
| `TKwotyPozycji` | `KwotaA` current year (or current period), `KwotaB` preceding year, `KwotaB1` (optional) **restated** comparatives | the books must give the prior-year figure and, after a policy change or material error (KSR 7 pkt 5.7 lit. c), a restated prior-year figure |
| `TKwota3` | decimal, totalDigits 18, fractionDigits 2 | `DECIMAL(18,2)`, złoty and grosze |
| `TPozycjaSprawozdania` | a position + any number of `PozycjaUszczegolawiajaca` (`NazwaPozycji` + `KwotyPozycji`) | the art. 50 ust. 1 extra detail (e.g. KSR 3 "Aktywa z tytułu niezakończonych umów budowlanych", KSR 8 "w tym: Obiekty w zabudowie", art. 36 ust. 2e dopłaty) goes here |
| `TKwotyPozycjiTys` | same in thousands | the W TYSIĄCACH forms (art. 45 ust. 5 allows rounding to thousands) |

Source: `StrukturyDanychSprFin_v2-0E.xsd` (types `TKwotyPozycji`, `TKwota3`, `TPozycjaSprawozdania`, `TKwotyPozycjiSprawozdania`).

### 2.5 Notes in the XML: text and files, plus one structured table

`TInformacjaDodatkowa` is `Opis` (text) and an optional `Plik` (`Nazwa` + base64 `Zawartosc`). So the notes listed in §4 are free text or attached tables — **not** XML positions. The one structured note is the CIT reconciliation, `InformacjaDodatkowaDotyczacaPodatkuDochodowego` (type `TInformacjaDodatkowaDotyczacaPodatkuDochodowego`), documented as "Wypełniają wyłącznie jednostki zobowiązane":

| Element | Name (XSD documentation) |
|---|---|
| `P_ID_1` | A. Zysk (strata) brutto za dany rok |
| `P_ID_2` | B. Przychody zwolnione z opodatkowania (trwałe różnice …), w tym: |
| `P_ID_3` | C. Przychody niepodlegające opodatkowaniu w roku bieżącym, w tym: |
| `P_ID_4` | D. Przychody podlegające opodatkowaniu w roku bieżącym, ujęte w księgach rachunkowych lat ubiegłych w tym: |
| `P_ID_5` | E. Koszty niestanowiące kosztów uzyskania przychodów (trwałe różnice …), w tym: |
| `P_ID_6` | F. Koszty nieuznawane za koszty uzyskania przychodów w bieżącym roku, w tym: |
| `P_ID_7` | G. Koszty uznawane za koszty uzyskania przychodów w roku bieżącym ujęte w księgach lat ubiegłych, w tym: |
| `P_ID_8` | H. Strata z lat ubiegłych, w tym: |
| `P_ID_9` | I. Inne zmiany podstawy opodatkowania, w tym: |
| `P_ID_10` | J. Podstawa opodatkowania podatkiem dochodowym |
| `P_ID_11` | K. Podatek dochodowy |

Items B–G and I take `Kwota` (RB current year, RP prior year optional), optional `PozycjaUzytkownika` lines (name + amounts + legal basis) and `Pozostale` "możliwość łącznego podania różnic o wartościach niższych niż 20.000 zł" (`TPozycjaInformacjiDodatkowej`); item H (losses) takes `Kwota` and user lines without legal basis (`TPozycjaInformacjiDodatkowej2`). It implements annex 1 note 2.6 and KSR 2 pkt 14.2 lit. c (which points to "wzoru sprawozdania finansowego w postaci elektronicznej"). The tag set that feeds it (permanent vs timing differences per account) belongs to strand (a). *To verify:* which entities count as "zobowiązane" — the schema offers the element in the micro and small forms too, while annexes 4 and 5 do not list the reconciliation.

### 2.6 How to reuse the parse

The position trees in §3 were extracted from the domain XSDs above with a small stdlib parser (walk `xsd:complexType` → `xsd:element`, read `xsd:documentation`, skip `PozycjaUszczegolawiajaca*`). The law's annex text and the XSD documentation strings were compared for annexes 1, 4 and 5 and match position for position, apart from capitalisation and the repealed tiret "– (uchylone)" under pasywa A in annex 4, which the XSD omits (UoR pp. 84–93, 113–119). The XSD element id is the stable key to map accounts to; e.g. `Aktywa_B_II_3_A_1` = *Aktywa B.II.3.a) – do 12 miesięcy*.

---

## 3. Every position, with its feeding accounts and the analytics it needs

Notation in the "Poz." column follows the annex: `A.` / `I.` / `1.` / `a)` / `–` (tiret). "Σ" = sum of the sub-positions, no own account. "computed" = arithmetic of other positions. The groups column is practice (see conventions); the analytics column is what a posting or open item must carry so that the position is produced without manual work.

### 3.1 Bilans — załącznik nr 1 (`BilansJednostkaInna`)

| Poz. | Nazwa (annex text) | XSD element | Account groups (practice) | Analytics needed to fill it without manual work |
|---|---|---|---|---|
|  | Aktywa razem ¹ | `Aktywa` | Σ | Aktywa razem = Pasywa razem (check) |
| A. | Aktywa trwałe | `Aktywa_A` | Σ subpositions |  |
| A.I. | Wartości niematerialne i prawne | `Aktywa_A_I` | 0 (WNiP, less umorzenie/odpisy) | per-object register (art. 17 ust. 1 pkt 1); MOV for note 1.1 |
| A.I.1. | Koszty zakończonych prac rozwojowych | `Aktywa_A_I_1` | 0 | kind = completed development costs; amortisation period (note 1.3) |
| A.I.2. | Wartość firmy | `Aktywa_A_I_2` | 0 | kind = goodwill; write-off period (note 1.3, art. 44b ust. 10) |
| A.I.3. | Inne wartości niematerialne i prawne | `Aktywa_A_I_3` | 0 | kind = other WNiP (licences, software, rights…); MOV |
| A.I.4. | Zaliczki na wartości niematerialne i prawne | `Aktywa_A_I_4` | 0 or 2 (practice varies) | advance paid flagged ‘on WNiP’ (not B.I.5) |
| A.II. | Rzeczowe aktywa trwałe | `Aktywa_A_II` | Σ |  |
| A.II.1. | Środki trwałe | `Aktywa_A_II_1` | 0 (ŚT, less umorzenie/odpisy) | per-object register; KŚT group → line a–e; own vs finance-leased (art. 3 ust. 4); MOV |
| A.II.1.a) | grunty (w tym prawo użytkowania wieczystego gruntu) | `Aktywa_A_II_1_A` | 0 | KŚT 0 land + perpetual usufruct; PWUG value separately (note 1.4) |
| A.II.1.b) | budynki, lokale, prawa do lokali i obiekty inżynierii lądowej i wodnej | `Aktywa_A_II_1_B` | 0 | KŚT 1–2 (practice mapping) |
| A.II.1.c) | urządzenia techniczne i maszyny | `Aktywa_A_II_1_C` | 0 | KŚT 3–6 (practice mapping) |
| A.II.1.d) | środki transportu | `Aktywa_A_II_1_D` | 0 | KŚT 7 (practice mapping) |
| A.II.1.e) | inne środki trwałe | `Aktywa_A_II_1_E` | 0 | KŚT 8–9 and rest (practice mapping) |
| A.II.2. | Środki trwałe w budowie | `Aktywa_A_II_2` | 0 (środki trwałe w budowie) | per project; capitalised interest and FX (note 2.7); environmental capex (note 2.9) |
| A.II.3. | Zaliczki na środki trwałe w budowie | `Aktywa_A_II_3` | 0 or 2 (practice varies) | advance paid flagged ‘on ŚT w budowie’ |
| A.III. | Należności długoterminowe | `Aktywa_A_III` | 2 | non-trade receivables with DUE > 12 m from balance date (art. 3 ust. 1 pkt 18c); trade receivables never here |
| A.III.1. | Od jednostek powiązanych | `Aktywa_A_III_1` | 2 | CP = P |
| A.III.2. | Od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale | `Aktywa_A_III_2` | 2 | CP = Z |
| A.III.3. | Od pozostałych jednostek | `Aktywa_A_III_3` | 2 | CP = O |
| A.IV. | Inwestycje długoterminowe | `Aktywa_A_IV` | Σ |  |
| A.IV.1. | Nieruchomości | `Aktywa_A_IV_1` | 0 (inwestycje w nieruchomości) | valuation model (cost or FV, art. 28 ust. 1 pkt 1a); MOV |
| A.IV.2. | Wartości niematerialne i prawne | `Aktywa_A_IV_2` | 0 | WNiP held as investment; MOV |
| A.IV.3. | Długoterminowe aktywa finansowe | `Aktywa_A_IV_3` | 0 (długoterminowe aktywa finansowe) | per instrument/investee; instrument type; category (rozp. 2024 § 4); CUR; DUE |
| A.IV.3.a) | w jednostkach powiązanych | `Aktywa_A_IV_3_A` | 0 | CP = P |
| A.IV.3.a) – | – udziały lub akcje | `Aktywa_A_IV_3_A_1` | 0 | instrument = shares |
| A.IV.3.a) – | – inne papiery wartościowe | `Aktywa_A_IV_3_A_2` | 0 | instrument = other securities |
| A.IV.3.a) – | – udzielone pożyczki | `Aktywa_A_IV_3_A_3` | 0 | instrument = loan granted; part due ≤ 12 m moves to B.III |
| A.IV.3.a) – | – inne długoterminowe aktywa finansowe | `Aktywa_A_IV_3_A_4` | 0 | instrument = other |
| A.IV.3.b) | w pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale | `Aktywa_A_IV_3_B` | 0 | CP = Z |
| A.IV.3.b) – | – udziały lub akcje | `Aktywa_A_IV_3_B_1` | 0 | instrument = shares |
| A.IV.3.b) – | – inne papiery wartościowe | `Aktywa_A_IV_3_B_2` | 0 | instrument = other securities |
| A.IV.3.b) – | – udzielone pożyczki | `Aktywa_A_IV_3_B_3` | 0 | instrument = loan granted; part due ≤ 12 m moves to B.III |
| A.IV.3.b) – | – inne długoterminowe aktywa finansowe | `Aktywa_A_IV_3_B_4` | 0 | instrument = other |
| A.IV.3.c) | w pozostałych jednostkach | `Aktywa_A_IV_3_C` | 0 | CP = O |
| A.IV.3.c) – | – udziały lub akcje | `Aktywa_A_IV_3_C_1` | 0 | instrument = shares |
| A.IV.3.c) – | – inne papiery wartościowe | `Aktywa_A_IV_3_C_2` | 0 | instrument = other securities |
| A.IV.3.c) – | – udzielone pożyczki | `Aktywa_A_IV_3_C_3` | 0 | instrument = loan granted; part due ≤ 12 m moves to B.III |
| A.IV.3.c) – | – inne długoterminowe aktywa finansowe | `Aktywa_A_IV_3_C_4` | 0 | instrument = other |
| A.IV.4. | Inne inwestycje długoterminowe | `Aktywa_A_IV_4` | 0 | other long-term investments |
| A.V. | Długoterminowe rozliczenia międzyokresowe | `Aktywa_A_V` | Σ |  |
| A.V.1. | Aktywa z tytułu odroczonego podatku dochodowego | `Aktywa_A_V_1` | 6 (RMK, deferred tax asset) or own account | origin: temporary difference / tax loss (expiry) / unused relief; via P&L vs equity (KSR 2 pkt 14.2–14.3); offset decision (art. 37 ust. 7) |
| A.V.2. | Inne rozliczenia międzyokresowe | `Aktywa_A_V_2` | 6 (RMK czynne) | portion settled > 12 m from balance date |
| B. | Aktywa obrotowe | `Aktywa_B` | Σ |  |
| B.I. | Zapasy | `Aktywa_B_I` | Σ | net of inventory write-downs (note 2.4) |
| B.I.1. | Materiały | `Aktywa_B_I_1` | 3 (materials, purchase settlement, price variances) | per item/warehouse (art. 17 ust. 2 method); write-downs |
| B.I.2. | Półprodukty i produkty w toku | `Aktywa_B_I_2` | 5 (WIP balance) and 6 | WIP per order; KSR 8 ‘w tym: Obiekty w zabudowie’ if developer |
| B.I.3. | Produkty gotowe | `Aktywa_B_I_3` | 6 (products, variances) | per product; write-downs |
| B.I.4. | Towary | `Aktywa_B_I_4` | 3 (goods, variances) | per item; write-downs |
| B.I.5. | Zaliczki na dostawy i usługi | `Aktywa_B_I_5` | 2 (advances paid) | advance paid flagged ‘on supplies/services’ (not A.I.4/A.II.3) |
| B.II. | Należności krótkoterminowe | `Aktywa_B_II` | Σ | net of receivable allowances (note 1.7) |
| B.II.1. | Należności od jednostek powiązanych | `Aktywa_B_II_1` | 2 | CP = P |
| B.II.1.a) | z tytułu dostaw i usług, o okresie spłaty: | `Aktywa_B_II_1_A` | 2 (trade receivables) | DUE per invoice → bucket |
| B.II.1.a) – | – do 12 miesięcy | `Aktywa_B_II_1_A_1` | 2 | bucket ≤ 12 m |
| B.II.1.a) – | – powyżej 12 miesięcy | `Aktywa_B_II_1_A_2` | 2 | bucket > 12 m (still short-term, art. 3 ust. 1 pkt 18c) |
| B.II.1.b) | inne | `Aktywa_B_II_1_B` | 2 | other receivables, CP = P; non-trade with DUE ≤ 12 m |
| B.II.2. | Należności od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale | `Aktywa_B_II_2` | 2 | CP = Z |
| B.II.2.a) | z tytułu dostaw i usług, o okresie spłaty: | `Aktywa_B_II_2_A` | 2 (trade receivables) | DUE per invoice → bucket |
| B.II.2.a) – | – do 12 miesięcy | `Aktywa_B_II_2_A_1` | 2 | bucket ≤ 12 m |
| B.II.2.a) – | – powyżej 12 miesięcy | `Aktywa_B_II_2_A_2` | 2 | bucket > 12 m (still short-term) |
| B.II.2.b) | inne | `Aktywa_B_II_2_B` | 2 | other receivables, CP = Z |
| B.II.3. | Należności od pozostałych jednostek | `Aktywa_B_II_3` | 2 | CP = O |
| B.II.3.a) | z tytułu dostaw i usług, o okresie spłaty: | `Aktywa_B_II_3_A` | 2 (trade receivables) | DUE per invoice → bucket |
| B.II.3.a) – | – do 12 miesięcy | `Aktywa_B_II_3_A_1` | 2 | bucket ≤ 12 m |
| B.II.3.a) – | – powyżej 12 miesięcy | `Aktywa_B_II_3_A_2` | 2 | bucket > 12 m (still short-term) |
| B.II.3.b) | z tytułu podatków, dotacji, ceł, ubezpieczeń społecznych i zdrowotnych oraz innych tytułów publicznoprawnych | `Aktywa_B_II_3_B` | 2 (rozrachunki publicznoprawne, debit balances) | per tax/authority (VAT, CIT, ZUS, customs, grants receivable); debit balances only — no offsetting with credit balances |
| B.II.3.c) | inne | `Aktywa_B_II_3_C` | 2 | other receivables (employees, deposits, claims, …) |
| B.II.3.d) | dochodzone na drodze sądowej | `Aktywa_B_II_3_D` | 2 | flag ‘dochodzone na drodze sądowej’ (litigation) per receivable |
| B.III. | Inwestycje krótkoterminowe | `Aktywa_B_III` | Σ |  |
| B.III.1. | Krótkoterminowe aktywa finansowe | `Aktywa_B_III_1` | Σ |  |
| B.III.1.a) | w jednostkach powiązanych | `Aktywa_B_III_1_A` | 1 (krótkoterminowe aktywa finansowe) | CP = P; instrument type; category (rozp. 2024 § 4); CUR |
| B.III.1.a) – | – udziały lub akcje | `Aktywa_B_III_1_A_1` | 1 | shares |
| B.III.1.a) – | – inne papiery wartościowe | `Aktywa_B_III_1_A_2` | 1 | other securities |
| B.III.1.a) – | – udzielone pożyczki | `Aktywa_B_III_1_A_3` | 1 or 2 (practice) | loans granted incl. current portion of long-term loans |
| B.III.1.a) – | – inne krótkoterminowe aktywa finansowe | `Aktywa_B_III_1_A_4` | 1 | other |
| B.III.1.b) | w pozostałych jednostkach | `Aktywa_B_III_1_B` | 1 | CP ∈ {Z, O} (only two classes here); instrument type; category; CUR |
| B.III.1.b) – | – udziały lub akcje | `Aktywa_B_III_1_B_1` | 1 | shares |
| B.III.1.b) – | – inne papiery wartościowe | `Aktywa_B_III_1_B_2` | 1 | other securities |
| B.III.1.b) – | – udzielone pożyczki | `Aktywa_B_III_1_B_3` | 1 or 2 (practice) | loans granted incl. current portion of long-term loans |
| B.III.1.b) – | – inne krótkoterminowe aktywa finansowe | `Aktywa_B_III_1_B_4` | 1 | other |
| B.III.1.c) | Środki pieniężne i inne aktywa pieniężne | `Aktywa_B_III_1_C` | 1 | CUR (note 3); restricted flag (RPP G.–); VAT account (note 1.18) |
| B.III.1.c) – | – środki pieniężne w kasie i na rachunkach | `Aktywa_B_III_1_C_1` | 1 (kasa, rachunki bankowe incl. rachunek VAT) | per cash desk / bank account; CUR |
| B.III.1.c) – | – inne środki pieniężne | `Aktywa_B_III_1_C_2` | 1 | e.g. deposits ≤ 3 m, cash in transit (practice) |
| B.III.1.c) – | – inne aktywa pieniężne | `Aktywa_B_III_1_C_3` | 1 | other monetary assets, incl. accrued interest (art. 3 ust. 1 pkt 25) |
| B.III.2. | Inne inwestycje krótkoterminowe | `Aktywa_B_III_2` | 1 | other short-term investments |
| B.IV. | Krótkoterminowe rozliczenia międzyokresowe | `Aktywa_B_IV` | 6 (RMK czynne) | portion settled ≤ 12 m; list of material items (note 1.14) |
| C. | Należne wpłaty na kapitał (fundusz) podstawowy | `Aktywa_C` | 2 or 8 (practice) | declared, unpaid contributions per shareholder (art. 36 ust. 2) |
| D. | Udziały (akcje) własne | `Aktywa_D` | 1 (practice) | per buy-back; reason, number, nominal (art. 49 ust. 2 pkt 5) |
|  | Pasywa razem ¹ | `Pasywa` | Σ |  |
| A. | Kapitał (fundusz) własny | `Pasywa_A` | Σ (8) | SCE movements (section 3.4) |
| A.I. | Kapitał (fundusz) podstawowy | `Pasywa_A_I` | 8 (kapitał podstawowy) | per shareholder and share class (note 1.8); MOV issue/redemption |
| A.II. | Kapitał (fundusz) zapasowy, w tym: | `Pasywa_A_II` | 8 (kapitał zapasowy) | MOV by source: agio / statutory profit allocation / above-statutory / loss coverage |
| A.II. – | – nadwyżka wartości sprzedaży (wartości emisyjnej) nad wartością nominalną udziałów (akcji) | `Pasywa_A_II_1` | 8 | source = share premium (agio), net of issue costs (art. 36 ust. 2b) |
| A.III. | Kapitał (fundusz) z aktualizacji wyceny, w tym: | `Pasywa_A_III` | 8 (kapitał z aktualizacji wyceny) | source; MOV; deferred tax charged to it (art. 37 ust. 9) |
| A.III. – | – z tytułu aktualizacji wartości godziwej | `Pasywa_A_III_1` | 8 | source = fair-value remeasurement (rozp. 2024 § 45 pkt 9) |
| A.IV. | Pozostałe kapitały (fundusze) rezerwowe, w tym: | `Pasywa_A_IV` | 8 (pozostałe kapitały rezerwowe) | purpose; dopłaty need their own line (art. 36 ust. 2e) |
| A.IV. – | – tworzone zgodnie z umową (statutem) spółki | `Pasywa_A_IV_1` | 8 | purpose = per articles of association |
| A.IV. – | – na udziały (akcje) własne | `Pasywa_A_IV_2` | 8 | purpose = own-share buy-back |
| A.V. | Zysk (strata) z lat ubiegłych | `Pasywa_A_V` | 8 (rozliczenie wyniku / wynik lat ubiegłych) | by year; flags ‘korekta błędu’, ‘zmiana zasad’ (KSR 7) |
| A.VI. | Zysk (strata) netto | `Pasywa_A_VI` | 8 (wynik finansowy) = close of 4/5/7 + income tax | none beyond P&L |
| A.VII. | Odpisy z zysku netto w ciągu roku obrotowego (wielkość ujemna) | `Pasywa_A_VII` | 8 | interim dividends / other in-year profit charges (art. 46 ust. 3) |
| B. | Zobowiązania i rezerwy na zobowiązania | `Pasywa_B` | Σ |  |
| B.I. | Rezerwy na zobowiązania | `Pasywa_B_I` | Σ | MOV per provision: created / used / released (note 1.11) |
| B.I.1. | Rezerwa z tytułu odroczonego podatku dochodowego | `Pasywa_B_I_1` | 8 (rezerwy) — deferred tax | type of temporary difference; via P&L vs equity (KSR 2 pkt 14.2) |
| B.I.2. | Rezerwa na świadczenia emerytalne i podobne | `Pasywa_B_I_2` | 8 (rezerwy) or 6 (bierne RMK, art. 39 ust. 2 pkt 2 → shown as rezerwy, art. 39 ust. 2a) | benefit type (retirement, jubilee…); LT/ST portion (KSR 6) |
| B.I.2. – | – długoterminowa | `Pasywa_B_I_2_1` | 8/6 | portion used > 12 m |
| B.I.2. – | – krótkoterminowa | `Pasywa_B_I_2_2` | 8/6 | portion used ≤ 12 m |
| B.I.3. | Pozostałe rezerwy | `Pasywa_B_I_3` | 8 (rezerwy) and 6 (bierne RMK under art. 39 ust. 2 pkt 2) | purpose (note 1.11; KSR 6 pkt 5.1); LT/ST |
| B.I.3. – | – długoterminowe | `Pasywa_B_I_3_1` | 8/6 | portion used > 12 m |
| B.I.3. – | – krótkoterminowe | `Pasywa_B_I_3_2` | 8/6 | portion used ≤ 12 m |
| B.II. | Zobowiązania długoterminowe | `Pasywa_B_II` | Σ | repayment schedule per contract: part due ≤ 12 m moves to B.III; buckets ≤1 / 1–3 / 3–5 / >5 years (note 1.12); secured flag (note 1.13) |
| B.II.1. | Wobec jednostek powiązanych | `Pasywa_B_II_1` | 1 or 2 | CP = P (instrument type still needed for annex 5) |
| B.II.2. | Wobec pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale | `Pasywa_B_II_2` | 1 or 2 | CP = Z (instrument type still needed for annex 5) |
| B.II.3. | Wobec pozostałych jednostek | `Pasywa_B_II_3` | Σ | CP = O |
| B.II.3.a) | kredyty i pożyczki | `Pasywa_B_II_3_A` | 1 (kredyty) / 2 (pożyczki) | lender; DUE schedule; CUR |
| B.II.3.b) | z tytułu emisji dłużnych papierów wartościowych | `Pasywa_B_II_3_B` | 2 | issue; DUE |
| B.II.3.c) | inne zobowiązania finansowe | `Pasywa_B_II_3_C` | 2 | incl. finance-lease liabilities (KSR 5 pkt 12.1.2); derivatives |
| B.II.3.d) | zobowiązania wekslowe | `Pasywa_B_II_3_D` | 2 | bills of exchange |
| B.II.3.e) | inne | `Pasywa_B_II_3_E` | 2 | other long-term liabilities (never trade payables) |
| B.III. | Zobowiązania krótkoterminowe | `Pasywa_B_III` | Σ |  |
| B.III.1. | Zobowiązania wobec jednostek powiązanych | `Pasywa_B_III_1` | 2 | CP = P |
| B.III.1.a) | z tytułu dostaw i usług, o okresie wymagalności: | `Pasywa_B_III_1_A` | 2 (trade payables) | DUE per invoice → bucket |
| B.III.1.a) – | – do 12 miesięcy | `Pasywa_B_III_1_A_1` | 2 | bucket ≤ 12 m |
| B.III.1.a) – | – powyżej 12 miesięcy | `Pasywa_B_III_1_A_2` | 2 | bucket > 12 m (still short-term, art. 3 ust. 1 pkt 22) |
| B.III.1.b) | inne | `Pasywa_B_III_1_B` | 1 or 2 | other liabilities incl. loans, CP = P |
| B.III.2. | Zobowiązania wobec pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale | `Pasywa_B_III_2` | 2 | CP = Z |
| B.III.2.a) | z tytułu dostaw i usług, o okresie wymagalności: | `Pasywa_B_III_2_A` | 2 (trade payables) | DUE per invoice → bucket |
| B.III.2.a) – | – do 12 miesięcy | `Pasywa_B_III_2_A_1` | 2 | bucket ≤ 12 m |
| B.III.2.a) – | – powyżej 12 miesięcy | `Pasywa_B_III_2_A_2` | 2 | bucket > 12 m |
| B.III.2.b) | inne | `Pasywa_B_III_2_B` | 1 or 2 | other liabilities incl. loans, CP = Z |
| B.III.3. | Zobowiązania wobec pozostałych jednostek | `Pasywa_B_III_3` | Σ | CP = O |
| B.III.3.a) | kredyty i pożyczki | `Pasywa_B_III_3_A` | 1 (kredyty bankowe incl. overdraft) / 2 (pożyczki) | incl. current portion of long-term loans; CUR |
| B.III.3.b) | z tytułu emisji dłużnych papierów wartościowych | `Pasywa_B_III_3_B` | 2 | debt securities ≤ 12 m |
| B.III.3.c) | inne zobowiązania finansowe | `Pasywa_B_III_3_C` | 2 | finance-lease current portion; other financial |
| B.III.3.d) | z tytułu dostaw i usług, o okresie wymagalności: | `Pasywa_B_III_3_D` | 2 (trade payables) | DUE → bucket; investment payables (for fixed assets) kept apart for RPP A.II.8 |
| B.III.3.d) – | – do 12 miesięcy | `Pasywa_B_III_3_D_1` | 2 | bucket ≤ 12 m |
| B.III.3.d) – | – powyżej 12 miesięcy | `Pasywa_B_III_3_D_2` | 2 | bucket > 12 m |
| B.III.3.e) | zaliczki otrzymane na dostawy i usługi | `Pasywa_B_III_3_E` | 2 (advances received) | advance received flag per customer |
| B.III.3.f) | zobowiązania wekslowe | `Pasywa_B_III_3_F` | 2 | bills of exchange |
| B.III.3.g) | z tytułu podatków, ceł, ubezpieczeń społecznych i zdrowotnych oraz innych tytułów publicznoprawnych | `Pasywa_B_III_3_G` | 2 (rozrachunki publicznoprawne, credit balances) | per tax/authority (VAT, PIT, CIT, ZUS, PFRON…) |
| B.III.3.h) | z tytułu wynagrodzeń | `Pasywa_B_III_3_H` | 2 (rozrachunki z tytułu wynagrodzeń) | per employee (art. 17 ust. 1 pkt 3) |
| B.III.3.I. | inne | `Pasywa_B_III_3_I` | 2 | other (e.g. dividends payable, credit balances on other settlements) |
| B.III.4. | Fundusze specjalne | `Pasywa_B_III_4` | 8 (fundusze specjalne, e.g. ZFŚS) | per fund (art. 46 ust. 4); MOV (allocation from profit kept apart, KSR 1) |
| B.IV. | Rozliczenia międzyokresowe | `Pasywa_B_IV` | Σ |  |
| B.IV.1. | Ujemna wartość firmy | `Pasywa_B_IV_1` | 8 (RMP) or own account | per acquisition; settlement schedule |
| B.IV.2. | Inne rozliczenia międzyokresowe | `Pasywa_B_IV_2` | 8 (RMP) and 6 (bierne RMK not shown as provisions — practice) | type: grants for fixed assets (art. 41 ust. 1 pkt 2), prepaid revenue, KSR 3 contract balances; LT/ST |
| B.IV.2. – | – długoterminowe | `Pasywa_B_IV_2_1` | 8/6 | portion > 12 m |
| B.IV.2. – | – krótkoterminowe | `Pasywa_B_IV_2_2` | 8/6 | portion ≤ 12 m |

¹ element required by the XSD (minOccurs ≠ 0); all others may be omitted when empty (art. 50 ust. 3 UoR).

### 3.2 Rachunek zysków i strat — wariant porównawczy (`RZiSJednostkaInna/RZiSPor`)

| Poz. | Nazwa (annex text) | XSD element | Account groups (practice) | Analytics needed to fill it without manual work |
|---|---|---|---|---|
|  | Rachunek zysków i strat (wariant porównawczy) ¹ | `RZiSPor` | Σ subpositions |  |
| A. | Przychody netto ze sprzedaży i zrównane z nimi, w tym: | `A` | Σ (7, 4/5/6) |  |
| A. – | – od jednostek powiązanych | `A_J` | 7 | CP = P on revenue lines (customer master data) |
| A.I. | Przychody netto ze sprzedaży produktów | `A_I` | 7 (sprzedaż produktów i usług) | product/service type and market (note 2.1); net of rebates; subsidies to prices included (art. 3 ust. 1 pkt 30a) |
| A.II. | Zmiana stanu produktów (zwiększenie – wartość dodatnia, zmniejszenie – wartość ujemna) | `A_II` | 6 and 5 (change of products, semi-finished, WIP) | only production-related changes; KSR 8 ‘w tym: Obiekty w zabudowie’ |
| A.III. | Koszt wytworzenia produktów na własne potrzeby jednostki | `A_III` | 4/5 (rozliczenie kosztów: own-use production) | own work capitalised (fixed assets, R&D) tagged |
| A.IV. | Przychody netto ze sprzedaży towarów | `A_IV` | 7 (sprzedaż towarów) | goods vs products; market (note 2.1). Materials sold are NOT here (art. 3 ust. 1 pkt 32 lit. b) |
| B. | Koszty działalności operacyjnej | `B` | Σ (4 + value of goods sold) |  |
| B.I. | Amortyzacja | `B_I` | 4 (amortyzacja) | only depreciation charged to operating costs; investment-property depreciation goes to E.III (art. 3 ust. 1 pkt 32 lit. c) |
| B.II. | Zużycie materiałów i energii | `B_II` | 4 (zużycie materiałów i energii) |  |
| B.III. | Usługi obce | `B_III` | 4 (usługi obce) | auditor fees by service (note 5.6); operating-lease/rent fees (KSR 5) |
| B.IV. | Podatki i opłaty, w tym: | `B_IV` | 4 (podatki i opłaty) |  |
| B.IV. – | – podatek akcyzowy | `B_IV_1` | 4 | excise flag |
| B.V. | Wynagrodzenia | `B_V` | 4 (wynagrodzenia) | board / supervisory-board pay by organ (note 5.4) |
| B.VI. | Ubezpieczenia społeczne i inne świadczenia, w tym: | `B_VI` | 4 (ubezpieczenia społeczne i inne świadczenia) |  |
| B.VI. – | – emerytalne | `B_VI_1` | 4 | pension contributions flag |
| B.VII. | Pozostałe koszty rodzajowe | `B_VII` | 4 (pozostałe koszty rodzajowe) |  |
| B.VIII. | Wartość sprzedanych towarów | `B_VIII` | 7 (wartość sprzedanych towarów) | goods only, not materials |
| C. | Zysk (strata) ze sprzedaży (A–B) | `C` | computed | A − B |
| D. | Pozostałe przychody operacyjne | `D` | Σ 7 (pozostałe przychody operacyjne) |  |
| D.I. | Zysk z tytułu rozchodu niefinansowych aktywów trwałych | `D_I` | 7 (both legs of fixed-asset disposals) | per disposal: proceeds − net book value − costs; net result goes to D.I or E.I (KSR 11 pkt 11.1) |
| D.II. | Dotacje | `D_II` | 7 | grants, incl. release of grant RMP (art. 41) |
| D.III. | Aktualizacja wartości aktywów niefinansowych | `D_III` | 7 | reversal of write-downs of non-financial assets (art. 35c); asset type |
| D.IV. | Inne przychody operacyjne | `D_IV` | 7 | other, incl. revenue from materials sold (MF Q&A 2025), released provisions (art. 35d ust. 4) |
| E. | Pozostałe koszty operacyjne | `E` | Σ 7 (pozostałe koszty operacyjne) |  |
| E.I. | Strata z tytułu rozchodu niefinansowych aktywów trwałych | `E_I` | 7 | net loss on fixed-asset disposals |
| E.II. | Aktualizacja wartości aktywów niefinansowych | `E_II` | 7 | write-downs of non-financial assets incl. trade receivables (art. 35b ust. 2), inventory, fixed assets |
| E.III. | Inne koszty operacyjne | `E_III` | 7 | other, incl. value of materials sold, provisions created, donations, penalties |
| F. | Zysk (strata) z działalności operacyjnej (C+D–E) | `F` | computed | C + D − E |
| G. | Przychody finansowe | `G` | Σ 7 (przychody finansowe) |  |
| G.I. | Dywidendy i udziały w zyskach, w tym: | `G_I` | 7 | investee flags: related? equity stake? (two independent attributes) |
| G.I.a) | Od jednostek powiązanych, w tym: | `G_I_A` | 7 | investee CP = P |
| G.I.a) – | – w których jednostka posiada zaangażowanie w kapitale | `G_I_A_1` | 7 | investee related AND with equity stake |
| G.I.b) | Od jednostek pozostałych, w tym: | `G_I_B` | 7 | investee not related |
| G.I.b) – | – w których jednostka posiada zaangażowanie w kapitale | `G_I_B_1` | 7 | investee not related, with equity stake |
| G.II. | Odsetki, w tym: | `G_II` | 7 | interest income; CF activity tag |
| G.II. – | – od jednostek powiązanych | `G_II_J` | 7 | CP = P |
| G.III. | Zysk z tytułu rozchodu aktywów finansowych, w tym: | `G_III` | 7 | per disposal, net |
| G.III. – | – w jednostkach powiązanych | `G_III_J` | 7 | CP = P |
| G.IV. | Aktualizacja wartości aktywów finansowych | `G_IV` | 7 | revaluation / reversal of write-downs of financial assets (rozp. 2024 § 42) |
| G.V. | Inne | `G_V` | 7 | other, incl. net positive FX differences (art. 42 ust. 3); FX realised/unrealised |
| H. | Koszty finansowe | `H` | Σ 7 (koszty finansowe) |  |
| H.I. | Odsetki, w tym: | `H_I` | 7 | interest expense; CF activity tag |
| H.I. – | – dla jednostek powiązanych | `H_I_J` | 7 | CP = P |
| H.II. | Strata z tytułu rozchodu aktywów finansowych, w tym: | `H_II` | 7 | per disposal, net |
| H.II. – | – w jednostkach powiązanych | `H_II_J` | 7 | CP = P |
| H.III. | Aktualizacja wartości aktywów finansowych | `H_III` | 7 | write-downs of financial assets |
| H.IV. | Inne | `H_IV` | 7 | other, incl. net negative FX, issue costs above agio (art. 36 ust. 2b) |
| I. | Zysk (strata) brutto (F+G–H) ¹ | `I` | computed | F + G − H |
| J. | Podatek dochodowy | `J` | 8 (podatek dochodowy) | current vs deferred part (art. 37 ust. 8) |
| K. | Pozostałe obowiązkowe zmniejszenia zysku (zwiększenia straty) | `K` | 8 | other mandatory charges on profit (practice) |
| L. | Zysk (strata) netto (I–J–K) ¹ | `L` | 8 (wynik finansowy) | I − J − K; equals Pasywa A.VI |

¹ element required by the XSD (minOccurs ≠ 0); all others may be omitted when empty (art. 50 ust. 3 UoR).

### 3.3 Rachunek zysków i strat — wariant kalkulacyjny (`RZiSJednostkaInna/RZiSKalk`)

| Poz. | Nazwa (annex text) | XSD element | Account groups (practice) | Analytics needed to fill it without manual work |
|---|---|---|---|---|
|  | Rachunek zysków i strat (wariant kalkulacyjny) ¹ | `RZiSKalk` | Σ subpositions |  |
| A. | Przychody netto ze sprzedaży produktów i towarów, w tym: | `A` | Σ 7 |  |
| A. – | – od jednostek powiązanych | `A_J` | 7 | CP = P on revenue lines (customer master data) |
| A.I. | Przychody netto ze sprzedaży produktów | `A_I` | 7 (sprzedaż produktów i usług) | product/service type and market (note 2.1) |
| A.II. | Przychody netto ze sprzedaży towarów | `A_II` | 7 (sprzedaż towarów) | goods only, not materials |
| B. | Koszty sprzedanych produktów i towarów, w tym: | `B` | Σ 7 (koszt sprzedanych produktów, wartość sprzedanych towarów) |  |
| B. – | – jednostkom powiązanym | `B_J` | 7 | CP = P carried from the sales line to its cost of sales |
| B.I. | Koszt wytworzenia sprzedanych produktów | `B_I` | 7 fed from 5/6 (koszt wytworzenia sprzedanych produktów) | unabsorbed fixed overhead per KSR 13 |
| B.II. | Wartość sprzedanych towarów | `B_II` | 7 (wartość sprzedanych towarów) |  |
| C. | Zysk (strata) brutto ze sprzedaży (A–B) | `C` | computed | A − B |
| D. | Koszty sprzedaży | `D` | 5 (koszty sprzedaży) | nature-of-cost split still required (note 2.2) |
| E. | Koszty ogólnego zarządu | `E` | 5 (koszty ogólnego zarządu) | nature-of-cost split still required (note 2.2) |
| F. | Zysk (strata) ze sprzedaży (C–D–E) | `F` | computed | C − D − E |
| G. | Pozostałe przychody operacyjne | `G` | Σ 7 (pozostałe przychody operacyjne) |  |
| G.I. | Zysk z tytułu rozchodu niefinansowych aktywów trwałych | `G_I` | 7 (both legs of fixed-asset disposals) | per disposal: proceeds − net book value − costs; net result goes to D.I or E.I (KSR 11 pkt 11.1) |
| G.II. | Dotacje | `G_II` | 7 | grants, incl. release of grant RMP (art. 41) |
| G.III. | Aktualizacja wartości aktywów niefinansowych | `G_III` | 7 | reversal of write-downs of non-financial assets (art. 35c); asset type |
| G.IV. | Inne przychody operacyjne | `G_IV` | 7 | other, incl. revenue from materials sold (MF Q&A 2025), released provisions (art. 35d ust. 4) |
| H. | Pozostałe koszty operacyjne | `H` | Σ 7 (pozostałe koszty operacyjne) |  |
| H.I. | Strata z tytułu rozchodu niefinansowych aktywów trwałych | `H_I` | 7 | net loss on fixed-asset disposals |
| H.II. | Aktualizacja wartości aktywów niefinansowych | `H_II` | 7 | write-downs of non-financial assets incl. trade receivables (art. 35b ust. 2), inventory, fixed assets |
| H.III. | Inne koszty operacyjne | `H_III` | 7 | other, incl. value of materials sold, provisions created, donations, penalties |
| I. | Zysk (strata) z działalności operacyjnej (F+G–H) | `I` | computed | F + G − H |
| J. | Przychody finansowe | `J` | Σ 7 (przychody finansowe) |  |
| J.I. | Dywidendy i udziały w zyskach, w tym: | `J_I` | 7 | investee flags: related? equity stake? (two independent attributes) |
| J.I.a) | od jednostek powiązanych, w tym: | `J_I_A` | 7 | investee CP = P |
| J.I.a) – | – w których jednostka posiada zaangażowanie w kapitale | `J_I_A_1` | 7 | investee related AND with equity stake |
| J.I.b) | od jednostek pozostałych, w tym: | `J_I_B` | 7 | investee not related |
| J.I.b) – | – w których jednostka posiada zaangażowanie w kapitale | `J_I_B_1` | 7 | investee not related, with equity stake |
| J.II. | Odsetki, w tym: | `J_II` | 7 | interest income; CF activity tag |
| J.II. – | – od jednostek powiązanych | `J_II_J` | 7 | CP = P |
| J.III. | Zysk z tytułu rozchodu aktywów finansowych, w tym: | `J_III` | 7 | per disposal, net |
| J.III. – | – w jednostkach powiązanych | `J_III_J` | 7 | CP = P |
| J.IV. | Aktualizacja wartości aktywów finansowych | `J_IV` | 7 | revaluation / reversal of write-downs of financial assets (rozp. 2024 § 42) |
| J.V. | Inne | `J_V` | 7 | other, incl. net positive FX differences (art. 42 ust. 3); FX realised/unrealised |
| K. | Koszty finansowe | `K` | Σ 7 (koszty finansowe) |  |
| K.I. | Odsetki, w tym: | `K_I` | 7 | interest expense; CF activity tag |
| K.I. – | – dla jednostek powiązanych | `K_I_J` | 7 | CP = P |
| K.II. | Strata z tytułu rozchodu aktywów finansowych, w tym: | `K_II` | 7 | per disposal, net |
| K.II. – | – w jednostkach powiązanych | `K_II_J` | 7 | CP = P |
| K.III. | Aktualizacja wartości aktywów finansowych | `K_III` | 7 | write-downs of financial assets |
| K.IV. | Inne | `K_IV` | 7 | other, incl. net negative FX, issue costs above agio (art. 36 ust. 2b) |
| L. | Zysk (strata) brutto (I+J–K) ¹ | `L` | computed | I + J − K |
| M. | Podatek dochodowy | `M` | 8 (podatek dochodowy) | current vs deferred part (art. 37 ust. 8) |
| N. | Pozostałe obowiązkowe zmniejszenia zysku (zwiększenia straty) | `N` | 8 | other mandatory charges on profit (practice) |
| O. | Zysk (strata) netto (L–M–N) ¹ | `O` | 8 (wynik finansowy) | L − M − N; equals Pasywa A.VI |

¹ element required by the XSD (minOccurs ≠ 0); all others may be omitted when empty (art. 50 ust. 3 UoR).

### 3.4 Zestawienie zmian w kapitale (funduszu) własnym (`ZestZmianWKapitaleJednostkaInna`)

| Poz. | Nazwa (annex text) | XSD element | Account groups (practice) | Analytics needed to fill it without manual work |
|---|---|---|---|---|
| I. | Kapitał (fundusz) własny na początek okresu (BO) ¹ | `I` | 8 (all equity accounts) — prior-year closing | = prior-year Pasywa A (BZ) |
| I. – | – zmiany przyjętych zasad (polityki) rachunkowości | `I_1` | 8 | MOV = ‘zmiana zasad (polityki) rachunkowości’ (KSR 7) |
| I. – | – korekty błędów | `I_2` | 8 | MOV = ‘korekta błędu’ (KSR 7 pkt 5.7) |
| I.a. | Kapitał (fundusz) własny na początek okresu (BO), po korektach | `IA` | computed | I ± I.– |
| 1. | Kapitał (fundusz) podstawowy na początek okresu | `IA_1` | 8 (kapitał podstawowy) | opening balance |
| 1.1. | Zmiany kapitału (funduszu) podstawowego | `IA_1_1` | Σ |  |
| 1.1.a) | zwiększenie (z tytułu) | `IA_1_1_A` | 8 | MOV = increase (by title) |
| 1.1.a) – | – wydania udziałów (emisji akcji) | `IA_1_1_A_1` | 8 | MOV = share issue |
| 1.1.b) | zmniejszenie (z tytułu) | `IA_1_1_B` | 8 | MOV = decrease (by title) |
| 1.1.b) – | – umorzenia udziałów (akcji) | `IA_1_1_B_1` | 8 | MOV = redemption |
| 1.2. | Kapitał (fundusz) podstawowy na koniec okresu | `IA_1_2` | 8 | closing balance |
| 2. | Kapitał (fundusz) zapasowy na początek okresu | `IA_2` | 8 (kapitał zapasowy) | opening balance |
| 2.1. | Zmiany kapitału (funduszu) zapasowego | `IA_2_1` | Σ |  |
| 2.1.a) | zwiększenie (z tytułu) | `IA_2_1_A` | 8 | MOV = increase |
| 2.1.a) – | – emisji akcji powyżej wartości nominalnej | `IA_2_1_A_1` | 8 | MOV = agio |
| 2.1.a) – | – podziału zysku (ustawowo) | `IA_2_1_A_2` | 8 | MOV = statutory profit allocation (e.g. KSH art. 396 for S.A.) |
| 2.1.a) – | – podziału zysku (ponad wymaganą ustawowo minimalną wartość) | `IA_2_1_A_3` | 8 | MOV = allocation above statutory minimum |
| 2.1.b) | zmniejszenie (z tytułu) | `IA_2_1_B` | 8 | MOV = decrease |
| 2.1.b) – | – pokrycia straty | `IA_2_1_B_1` | 8 | MOV = loss coverage |
| 2.2. | Stan kapitału (funduszu) zapasowego na koniec okresu | `IA_2_2` | 8 | closing balance |
| 3. | Kapitał (fundusz) z aktualizacji wyceny na początek okresu – zmiany przyjętych zasad (polityki) rachunkowości | `IA_3` | 8 (kapitał z aktualizacji wyceny) | opening balance (incl. policy-change effect) |
| 3.1. | Zmiany kapitału (funduszu) z aktualizacji wyceny | `IA_3_1` | Σ |  |
| 3.1.a) | zwiększenie (z tytułu) | `IA_3_1_A` | 8 | MOV = increase (by title) |
| 3.1.b) | zmniejszenie (z tytułu) | `IA_3_1_B` | 8 | MOV = decrease (by title) |
| 3.1.b) – | – zbycia środków trwałych | `IA_3_1_B_1` | 8 | MOV = transfer on sale of fixed assets |
| 3.2. | Kapitał (fundusz) z aktualizacji wyceny na koniec okresu | `IA_3_2` | 8 | closing balance |
| 4. | Pozostałe kapitały (fundusze) rezerwowe na początek okresu | `IA_4` | 8 (pozostałe kapitały rezerwowe) | opening balance |
| 4.1. | Zmiany pozostałych kapitałów (funduszy) rezerwowych | `IA_4_1` | Σ |  |
| 4.1.a) | zwiększenie (z tytułu) | `IA_4_1_A` | 8 | MOV = increase |
| 4.1.b) | zmniejszenie (z tytułu) | `IA_4_1_B` | 8 | MOV = decrease |
| 4.2. | Pozostałe kapitały (fundusze) rezerwowe na koniec okresu | `IA_4_2` | 8 | closing balance |
| 5. | Zysk (strata) z lat ubiegłych na początek okresu | `IA_5` | 8 (wynik lat ubiegłych) | split profit side vs loss side |
| 5.1. | Zysk z lat ubiegłych na początek okresu | `IA_5_1` | 8 | retained profit, opening |
| 5.1. – | – zmiany przyjętych zasad (polityki) rachunkowości | `IA_5_1_1` | 8 | MOV = policy change |
| 5.1. – | – korekty błędów | `IA_5_1_2` | 8 | MOV = error correction |
| 5.2. | Zysk z lat ubiegłych na początek okresu, po korektach | `IA_5_2` | 8 | after corrections |
| 5.2.a) | zwiększenie (z tytułu) | `IA_5_2_A` | 8 | MOV = increase |
| 5.2.a) – | – podziału zysku z lat ubiegłych | `IA_5_2_A_1` | 8 | MOV = prior-year profit carried forward |
| 5.2.b) | zmniejszenie (z tytułu) | `IA_5_2_B` | 8 | MOV = decrease (dividend, transfer to capitals…) |
| 5.3. | Zysk z lat ubiegłych na koniec okresu | `IA_5_3` | 8 | closing |
| 5.4. | Strata z lat ubiegłych na początek okresu | `IA_5_4` | 8 | accumulated loss, opening |
| 5.4. – | – zmiany przyjętych zasad (polityki) rachunkowości | `IA_5_4_1` | 8 | MOV = policy change |
| 5.4. – | – korekty błędów | `IA_5_4_2` | 8 | MOV = error correction |
| 5.5. | Strata z lat ubiegłych na początek okresu, po korektach | `IA_5_5` | 8 | after corrections |
| 5.5.a) | zwiększenie (z tytułu) | `IA_5_5_A` | 8 | MOV = increase |
| 5.5.a) – | – przeniesienia straty z lat ubiegłych do pokrycia | `IA_5_5_A_1` | 8 | MOV = prior-year loss carried forward |
| 5.5.b) | zmniejszenie (z tytułu) | `IA_5_5_B` | 8 | MOV = decrease (loss covered) |
| 5.6. | Strata z lat ubiegłych na koniec okresu | `IA_5_6` | 8 | closing |
| 5.7. | Zysk (strata) z lat ubiegłych na koniec okresu | `IA_5_7` | computed | 5.3 − 5.6 |
| 6. | Wynik netto | `IA_6` | 8 (wynik finansowy) |  |
| 6.a) | zysk netto | `IA_6_A` | 8 | net profit |
| 6.b) | strata netto | `IA_6_B` | 8 | net loss |
| 6.c) | odpisy z zysku | `IA_6_C` | 8 | in-year profit charges (= Pasywa A.VII) |
| II. | Kapitał (fundusz) własny na koniec okresu (BZ) ¹ | `II` | computed | = Pasywa A (BZ) |
| III. | Kapitał (fundusz) własny, po uwzględnieniu proponowanego podziału zysku (pokrycia straty) ¹ | `III` | not a ledger figure | needs the management’s proposed profit distribution (note 1.10) |

¹ element required by the XSD (minOccurs ≠ 0); all others may be omitted when empty (art. 50 ust. 3 UoR).

### 3.5 Rachunek przepływów pieniężnych — metoda pośrednia (`RachPrzeplywowJednostkaInna/PrzeplywyPosr`)

| Poz. | Nazwa (annex text) | XSD element | Account groups (practice) | Analytics needed to fill it without manual work |
|---|---|---|---|---|
|  | Rachunek przepływów pieniężnych (metoda pośrednia) ¹ | `PrzeplywyPosr` | Σ subpositions |  |
| A. | Przepływy środków pieniężnych z działalności operacyjnej | `A` | Σ |  |
| A.I. | Zysk (strata) netto | `A_I` | 8 (wynik finansowy) | = RZiS net result |
| A.II. | Korekty razem | `A_II` | Σ | explain each difference vs balance-sheet change in note 4 |
| A.II.1. | Amortyzacja | `A_II_1` | 4 (amortyzacja) / 0 (umorzenie movement) | all depreciation incl. investment property |
| A.II.2. | Zyski (straty) z tytułu różnic kursowych | `A_II_2` | 7 (FX differences) | FX unrealised, and FX on investing/financing items; tag by nature |
| A.II.3. | Odsetki i udziały w zyskach (dywidendy) | `A_II_3` | 7 (interest, dividends) | interest/dividends classed as investing or financing (reversed here) |
| A.II.4. | Zysk (strata) z działalności inwestycyjnej | `A_II_4` | 7 | gains/losses on investing activity (fixed-asset and financial-asset disposals, revaluation) |
| A.II.5. | Zmiana stanu rezerw | `A_II_5` | 8/6 (rezerwy) | exclude deferred tax booked to equity |
| A.II.6. | Zmiana stanu zapasów | `A_II_6` | 3, 5, 6 | exclude non-cash moves (contributions in kind, transfers to fixed assets) |
| A.II.7. | Zmiana stanu należności | `A_II_7` | 2 | exclude investing receivables (e.g. from fixed-asset sales) and non-cash conversions (KSR 1) |
| A.II.8. | Zmiana stanu zobowiązań krótkoterminowych, z wyjątkiem pożyczek i kredytów | `A_II_8` | 2, 8 (special funds) | exclude loans, debt securities, financial liabilities, investment payables, equity-booked tax, ZFŚS allocation from profit (KSR 1) |
| A.II.9. | Zmiana stanu rozliczeń międzyokresowych | `A_II_9` | 6, 8 (RMK/RMP) | exclude deferred tax to equity, negative goodwill in year of origin, grants received in kind/in year (KSR 1) |
| A.II.10. | Inne korekty | `A_II_10` | various | other non-cash items (KSR 1 pkt 8.4 item 10) |
| A.III. | Przepływy pieniężne netto z działalności operacyjnej (I±II) | `A_III` | computed | I ± II |
| B. | Przepływy środków pieniężnych z działalności inwestycyjnej | `B` | Σ |  |
| B.I. | Wpływy | `B_I` | Σ |  |
| B.I.1. | Zbycie wartości niematerialnych i prawnych oraz rzeczowych aktywów trwałych | `B_I_1` | 1 (cash/bank) ↔ 2 (receivable from asset sale) | CF tag; amounts net of VAT (KSR 1 pkt 6.4) |
| B.I.2. | Zbycie inwestycji w nieruchomości oraz wartości niematerialne i prawne | `B_I_2` | 1 ↔ 2 | CF tag; investment property / WNiP held as investment |
| B.I.3. | Z aktywów finansowych, w tym: | `B_I_3` | Σ |  |
| B.I.3.a) | w jednostkach powiązanych | `B_I_3_A` | 1 | CF tag + CP = P |
| B.I.3.b) | w pozostałych jednostkach | `B_I_3_B` | Σ | CP ≠ P |
| B.I.3.b) – | – zbycie aktywów finansowych | `B_I_3_B_1` | 1 | CF tag: disposal of financial assets |
| B.I.3.b) – | – dywidendy i udziały w zyskach | `B_I_3_B_2` | 1 | CF tag: dividends received |
| B.I.3.b) – | – spłata udzielonych pożyczek długoterminowych | `B_I_3_B_3` | 1 | CF tag: repayment of long-term loans granted |
| B.I.3.b) – | – odsetki | `B_I_3_B_4` | 1 | CF tag: interest received (investing) |
| B.I.3.b) – | – inne wpływy z aktywów finansowych | `B_I_3_B_5` | 1 | CF tag: other |
| B.I.4. | Inne wpływy inwestycyjne | `B_I_4` | 1 | CF tag: other investing inflows |
| B.II. | Wydatki | `B_II` | Σ |  |
| B.II.1. | Nabycie wartości niematerialnych i prawnych oraz rzeczowych aktywów trwałych | `B_II_1` | 1 ↔ 2 (investment payables) | CF tag; net of deductible VAT (KSR 1 pkt 6.4) |
| B.II.2. | Inwestycje w nieruchomości oraz wartości niematerialne i prawne | `B_II_2` | 1 ↔ 2 | CF tag |
| B.II.3. | Na aktywa finansowe, w tym: | `B_II_3` | Σ |  |
| B.II.3.a) | w jednostkach powiązanych | `B_II_3_A` | 1 | CF tag + CP = P |
| B.II.3.b) | w pozostałych jednostkach | `B_II_3_B` | Σ | CP ≠ P |
| B.II.3.b) – | – nabycie aktywów finansowych | `B_II_3_B_1` | 1 | CF tag: purchase of financial assets |
| B.II.3.b) – | – udzielone pożyczki długoterminowe | `B_II_3_B_2` | 1 | CF tag: long-term loans granted |
| B.II.4. | Inne wydatki inwestycyjne | `B_II_4` | 1 | CF tag: other investing outflows |
| B.III. | Przepływy pieniężne netto z działalności inwestycyjnej (I–II) | `B_III` | computed | I − II |
| C. | Przepływy środków pieniężnych z działalności finansowej | `C` | Σ |  |
| C.I. | Wpływy | `C_I` | Σ |  |
| C.I.1. | Wpływy netto z wydania udziałów (emisji akcji) i innych instrumentów kapitałowych oraz dopłat do kapitału | `C_I_1` | 1 ↔ 8/2 | CF tag: equity issue net, dopłaty |
| C.I.2. | Kredyty i pożyczki | `C_I_2` | 1 ↔ 1/2 (loans) | CF tag: loans drawn |
| C.I.3. | Emisja dłużnych papierów wartościowych | `C_I_3` | 1 ↔ 2 | CF tag: debt securities issued |
| C.I.4. | Inne wpływy finansowe | `C_I_4` | 1 | CF tag: other financing inflows |
| C.II. | Wydatki | `C_II` | Σ |  |
| C.II.1. | Nabycie udziałów (akcji) własnych | `C_II_1` | 1 | CF tag: own-share buy-back |
| C.II.2. | Dywidendy i inne wypłaty na rzecz właścicieli | `C_II_2` | 1 ↔ 2 (dividends payable) | CF tag: dividends |
| C.II.3. | Inne, niż wypłaty na rzecz właścicieli, wydatki z tytułu podziału zysku | `C_II_3` | 1 | CF tag: other profit distributions |
| C.II.4. | Spłaty kredytów i pożyczek | `C_II_4` | 1 | CF tag: loan repayments (principal incl. realised FX, KSR 1 pkt 4.2) |
| C.II.5. | Wykup dłużnych papierów wartościowych | `C_II_5` | 1 | CF tag: debt redemption |
| C.II.6. | Z tytułu innych zobowiązań finansowych | `C_II_6` | 1 | CF tag: other financial liabilities |
| C.II.7. | Płatności zobowiązań z tytułu umów leasingu finansowego | `C_II_7` | 1 | CF tag: finance-lease payments (capital part) |
| C.II.8. | Odsetki | `C_II_8` | 1 | CF tag: interest paid (financing) — policy choice disclosed (KSR 1 pkt 4.2) |
| C.II.9. | Inne wydatki finansowe | `C_II_9` | 1 | CF tag: other financing outflows |
| C.III. | Przepływy pieniężne netto z działalności finansowej (I–II) | `C_III` | computed | I − II |
| D. | Przepływy pieniężne netto razem (A.III±B.III±C.III) | `D` | computed | A.III ± B.III ± C.III |
| E. | Bilansowa zmiana stanu środków pieniężnych, w tym: | `E` | 1 (balance change of B.III.1.c) | cash definition for RPP (≤ 3 m equivalents, KSR 1 pkt 4.4/4.7) |
| E. – | – zmiana stanu środków pieniężnych z tytułu różnic kursowych | `E_1` | 7 (FX on cash balances) | FX differences tagged ‘on cash’ — unrealised |
| F. | Środki pieniężne na początek okresu ¹ | `F` | 1 | opening cash per RPP definition |
| G. | Środki pieniężne na koniec okresu (F±D), w tym: ¹ | `G` | 1 | closing cash per RPP definition |
| G. – | – o ograniczonej możliwości dysponowania | `G_1` | 1 | restricted-cash flag (e.g. VAT account, escrow — practice) |

¹ element required by the XSD (minOccurs ≠ 0); all others may be omitted when empty (art. 50 ust. 3 UoR).

### 3.6 Rachunek przepływów pieniężnych — metoda bezpośrednia (`RachPrzeplywowJednostkaInna/PrzeplywyBezp`)

| Poz. | Nazwa (annex text) | XSD element | Account groups (practice) | Analytics needed to fill it without manual work |
|---|---|---|---|---|
|  | Rachunek przepływów pieniężnych (metoda bezpośrednia) ¹ | `PrzeplywyBezp` | Σ subpositions |  |
| A. | Przepływy środków pieniężnych z działalności operacyjnej | `A` | Σ | direct method also needs the indirect reconciliation in note 4 |
| A.I. | Wpływy | `A_I` | Σ |  |
| A.I.1. | Sprzedaż | `A_I_1` | 1 ↔ 2 (customers) | CF tag: receipts from sales incl. VAT (KSR 1) |
| A.I.2. | Inne wpływy z działalności operacyjnej | `A_I_2` | 1 | CF tag: other operating receipts (incl. VAT refunds) |
| A.II. | Wydatki | `A_II` | Σ |  |
| A.II.1. | Dostawy i usługi | `A_II_1` | 1 ↔ 2 (suppliers) | CF tag: payments for supplies and services incl. VAT |
| A.II.2. | Wynagrodzenia netto | `A_II_2` | 1 ↔ 2 (payroll) | CF tag: net wages |
| A.II.3. | Ubezpieczenia społeczne i zdrowotne oraz inne świadczenia | `A_II_3` | 1 ↔ 2 (ZUS) | CF tag: social and health insurance |
| A.II.4. | Podatki i opłaty o charakterze publicznoprawnym | `A_II_4` | 1 ↔ 2 (tax office) | CF tag: taxes and public levies incl. VAT paid and income tax |
| A.II.5. | Inne wydatki operacyjne | `A_II_5` | 1 | CF tag: other operating payments |
| A.III. | Przepływy pieniężne netto z działalności operacyjnej (I–II) | `A_III` | computed | I − II |
| B. | Przepływy środków pieniężnych z działalności inwestycyjnej | `B` | Σ |  |
| B.I. | Wpływy | `B_I` | Σ |  |
| B.I.1. | Zbycie wartości niematerialnych i prawnych oraz rzeczowych aktywów trwałych | `B_I_1` | 1 (cash/bank) ↔ 2 (receivable from asset sale) | CF tag; amounts net of VAT (KSR 1 pkt 6.4) |
| B.I.2. | Zbycie inwestycji w nieruchomości oraz wartości niematerialne i prawne | `B_I_2` | 1 ↔ 2 | CF tag; investment property / WNiP held as investment |
| B.I.3. | Z aktywów finansowych, w tym: | `B_I_3` | Σ |  |
| B.I.3.a) | w jednostkach powiązanych | `B_I_3_A` | 1 | CF tag + CP = P |
| B.I.3.b) | w pozostałych jednostkach | `B_I_3_B` | Σ | CP ≠ P |
| B.I.3.b) – | – zbycie aktywów finansowych | `B_I_3_B_1` | 1 | CF tag: disposal of financial assets |
| B.I.3.b) – | – dywidendy i udziały w zyskach | `B_I_3_B_2` | 1 | CF tag: dividends received |
| B.I.3.b) – | – spłata udzielonych pożyczek długoterminowych | `B_I_3_B_3` | 1 | CF tag: repayment of long-term loans granted |
| B.I.3.b) – | – odsetki | `B_I_3_B_4` | 1 | CF tag: interest received (investing) |
| B.I.3.b) – | – inne wpływy z aktywów finansowych | `B_I_3_B_5` | 1 | CF tag: other |
| B.I.4. | Inne wpływy inwestycyjne | `B_I_4` | 1 | CF tag: other investing inflows |
| B.II. | Wydatki | `B_II` | Σ |  |
| B.II.1. | Nabycie wartości niematerialnych i prawnych oraz rzeczowych aktywów trwałych | `B_II_1` | 1 ↔ 2 (investment payables) | CF tag; net of deductible VAT (KSR 1 pkt 6.4) |
| B.II.2. | Inwestycje w nieruchomości oraz wartości niematerialne i prawne | `B_II_2` | 1 ↔ 2 | CF tag |
| B.II.3. | Na aktywa finansowe, w tym: | `B_II_3` | Σ |  |
| B.II.3.a) | w jednostkach powiązanych | `B_II_3_A` | 1 | CF tag + CP = P |
| B.II.3.b) | w pozostałych jednostkach | `B_II_3_B` | Σ | CP ≠ P |
| B.II.3.b) – | – nabycie aktywów finansowych | `B_II_3_B_1` | 1 | CF tag: purchase of financial assets |
| B.II.3.b) – | – udzielone pożyczki długoterminowe | `B_II_3_B_2` | 1 | CF tag: long-term loans granted |
| B.II.4. | Inne wydatki inwestycyjne | `B_II_4` | 1 | CF tag: other investing outflows |
| B.III. | Przepływy pieniężne netto z działalności inwestycyjnej (I–II) | `B_III` | computed | I − II |
| C. | Przepływy środków pieniężnych z działalności finansowej | `C` | Σ |  |
| C.I. | Wpływy | `C_I` | Σ |  |
| C.I.1. | Wpływy netto z wydania udziałów (emisji akcji) i innych instrumentów kapitałowych oraz dopłat do kapitału | `C_I_1` | 1 ↔ 8/2 | CF tag: equity issue net, dopłaty |
| C.I.2. | Kredyty i pożyczki | `C_I_2` | 1 ↔ 1/2 (loans) | CF tag: loans drawn |
| C.I.3. | Emisja dłużnych papierów wartościowych | `C_I_3` | 1 ↔ 2 | CF tag: debt securities issued |
| C.I.4. | Inne wpływy finansowe | `C_I_4` | 1 | CF tag: other financing inflows |
| C.II. | Wydatki | `C_II` | Σ |  |
| C.II.1. | Nabycie udziałów (akcji) własnych | `C_II_1` | 1 | CF tag: own-share buy-back |
| C.II.2. | Dywidendy i inne wypłaty na rzecz właścicieli | `C_II_2` | 1 ↔ 2 (dividends payable) | CF tag: dividends |
| C.II.3. | Inne, niż wypłaty na rzecz właścicieli, wydatki z tytułu podziału zysku | `C_II_3` | 1 | CF tag: other profit distributions |
| C.II.4. | Spłaty kredytów i pożyczek | `C_II_4` | 1 | CF tag: loan repayments (principal incl. realised FX, KSR 1 pkt 4.2) |
| C.II.5. | Wykup dłużnych papierów wartościowych | `C_II_5` | 1 | CF tag: debt redemption |
| C.II.6. | Z tytułu innych zobowiązań finansowych | `C_II_6` | 1 | CF tag: other financial liabilities |
| C.II.7. | Płatności zobowiązań z tytułu umów leasingu finansowego | `C_II_7` | 1 | CF tag: finance-lease payments (capital part) |
| C.II.8. | Odsetki | `C_II_8` | 1 | CF tag: interest paid (financing) — policy choice disclosed (KSR 1 pkt 4.2) |
| C.II.9. | Inne wydatki finansowe | `C_II_9` | 1 | CF tag: other financing outflows |
| C.III. | Przepływy pieniężne netto z działalności finansowej (I–II) | `C_III` | computed | I − II |
| D. | Przepływy pieniężne netto, razem (A.III±B.III±C.III) | `D` | computed | A.III ± B.III ± C.III |
| E. | Bilansowa zmiana stanu środków pieniężnych, w tym: | `E` | 1 (balance change of B.III.1.c) | cash definition for RPP (≤ 3 m equivalents, KSR 1 pkt 4.4/4.7) |
| E. – | – zmiana stanu środków pieniężnych z tytułu różnic kursowych | `E_1` | 7 (FX on cash balances) | FX differences tagged ‘on cash’ — unrealised |
| F. | Środki pieniężne na początek okresu ¹ | `F` | 1 | opening cash per RPP definition |
| G. | Środki pieniężne na koniec okresu (F±D), w tym: ¹ | `G` | 1 | closing cash per RPP definition |
| G. – | – o ograniczonej możliwości dysponowania | `G_1` | 1 | restricted-cash flag (e.g. VAT account, escrow — practice) |

¹ element required by the XSD (minOccurs ≠ 0); all others may be omitted when empty (art. 50 ust. 3 UoR).

### 3.7 Bilans — załącznik nr 5, jednostka mała (`BilansJednostkaMala`)

| Poz. | Nazwa (annex text) | XSD element | Account groups (practice) | Analytics needed to fill it without manual work |
|---|---|---|---|---|
|  | Aktywa razem ¹ | `Aktywa` | Σ |  |
| A. | Aktywa trwałe | `Aktywa_A` | Σ subpositions |  |
| A.I. | Wartości niematerialne i prawne | `Aktywa_A_I` | 0 | = annex 1 A.I |
| A.II. | Rzeczowe aktywa trwałe, w tym: | `Aktywa_A_II` | 0 | = annex 1 A.II (incl. advances) |
| A.II. – | – środki trwałe | `Aktywa_A_II_1` | 0 | = annex 1 A.II.1 |
| A.II. – | – środki trwałe w budowie | `Aktywa_A_II_2` | 0 | = annex 1 A.II.2 |
| A.III. | Należności długoterminowe | `Aktywa_A_III` | 2 | = annex 1 A.III |
| A.IV. | Inwestycje długoterminowe, w tym: | `Aktywa_A_IV` | 0 | = annex 1 A.IV |
| A.IV. – | – nieruchomości | `Aktywa_A_IV_1` | 0 | = annex 1 A.IV.1 |
| A.IV. – | – długoterminowe aktywa finansowe | `Aktywa_A_IV_2` | 0 | = annex 1 A.IV.3 |
| A.V. | Długoterminowe rozliczenia międzyokresowe | `Aktywa_A_V` | 6 | = annex 1 A.V |
| B. | Aktywa obrotowe | `Aktywa_B` | Σ subpositions |  |
| B.I. | Zapasy | `Aktywa_B_I` | 3, 5, 6, 2 | = annex 1 B.I |
| B.II. | Należności krótkoterminowe, w tym: | `Aktywa_B_II` | 2 | = annex 1 B.II |
| B.II.a) | a) z tytułu dostaw i usług, w tym: | `Aktywa_B_II_A` | 2 | trade receivables of ALL counterparty classes |
| B.II.a) – | – do 12 miesięcy | `Aktywa_B_II_A_1` | 2 | bucket ≤ 12 m |
| B.II.a) – | – powyżej 12 miesięcy | `Aktywa_B_II_A_2` | 2 | bucket > 12 m |
| B.III. | Inwestycje krótkoterminowe, w tym: | `Aktywa_B_III` | 1 | = annex 1 B.III |
| B.III.a) | a) krótkoterminowe aktywa finansowe, w tym: | `Aktywa_B_III_A` | 1 | = annex 1 B.III.1 |
| B.III.a) – | – środki pieniężne w kasie i na rachunkach | `Aktywa_B_III_A_1` | 1 | = annex 1 B.III.1.c first tiret |
| B.IV. | Krótkoterminowe rozliczenia międzyokresowe | `Aktywa_B_IV` | 6 | = annex 1 B.IV |
| C. | Należne wpłaty na kapitał (fundusz) podstawowy | `Aktywa_C` | 2/8 | = annex 1 C |
| D. | Udziały (akcje) własne | `Aktywa_D` | 1 | = annex 1 D |
|  | Pasywa razem ¹ | `Pasywa` | Σ |  |
| A. | Kapitał (fundusz) własny | `Pasywa_A` | Σ subpositions |  |
| A.I. | Kapitał (fundusz) podstawowy | `Pasywa_A_I` | 8 | = annex 1 |
| A.II. | Kapitał (fundusz) zapasowy, w tym: | `Pasywa_A_II` | 8 | = annex 1 |
| A.II. – | – nadwyżka wartości sprzedaży (wartości emisyjnej) nad wartością nominalną udziałów (akcji) | `Pasywa_A_II_1` | 8 | = annex 1 |
| A.III. | Kapitał (fundusz) z aktualizacji wyceny, w tym: | `Pasywa_A_III` | 8 | = annex 1 |
| A.III. – | – z tytułu aktualizacji wartości godziwej | `Pasywa_A_III_1` | 8 | = annex 1 |
| A.IV. | Pozostałe kapitały (fundusze) rezerwowe | `Pasywa_A_IV` | 8 | = annex 1 A.IV |
| A.V. | Zysk (strata) z lat ubiegłych | `Pasywa_A_V` | 8 |  |
| A.VI. | Zysk (strata) netto | `Pasywa_A_VI` | 8 |  |
| A.VII. | Odpisy z zysku netto w ciągu roku obrotowego (wielkość ujemna) | `Pasywa_A_VII` | 8 |  |
| B. | Zobowiązania i rezerwy na zobowiązania | `Pasywa_B` | Σ subpositions |  |
| B.I. | Rezerwy na zobowiązania, w tym: | `Pasywa_B_I` | 8/6 | = annex 1 B.I |
| B.I. – | – rezerwa na świadczenia emerytalne i podobne | `Pasywa_B_I_1` | 8/6 | = annex 1 B.I.2 (LT + ST) |
| B.II. | Zobowiązania długoterminowe, w tym: | `Pasywa_B_II` | 1/2 | = annex 1 B.II |
| B.II. – | – z tytułu kredytów i pożyczek | `Pasywa_B_II_1` | 1/2 | loans of ALL counterparty classes (instrument-type tag on related-party liabilities too) |
| B.III. | Zobowiązania krótkoterminowe, w tym: | `Pasywa_B_III` | 1/2/8 | = annex 1 B.III |
| B.III.a) | a) z tytułu kredytów i pożyczek | `Pasywa_B_III_A` | 1/2 | loans of ALL counterparty classes |
| B.III.b) | b) z tytułu dostaw i usług, w tym: | `Pasywa_B_III_B` | 2 | trade payables of ALL counterparty classes |
| B.III.b) – | – Do 12 miesięcy | `Pasywa_B_III_B_1` | 2 | bucket ≤ 12 m |
| B.III.b) – | – Powyżej 12 miesięcy | `Pasywa_B_III_B_2` | 2 | bucket > 12 m |
| B.III.c) | c) fundusze specjalne | `Pasywa_B_III_C` | 8 | = annex 1 B.III.4 |
| B.IV. | Rozliczenia międzyokresowe | `Pasywa_B_IV` | 8/6 | = annex 1 B.IV |

¹ element required by the XSD (minOccurs ≠ 0); all others may be omitted when empty (art. 50 ust. 3 UoR).

### 3.8 RZiS — załącznik nr 5, wariant porównawczy (`RZiSJednostkaMala/RZiSPor`)

| Poz. | Nazwa (annex text) | XSD element | Account groups (practice) | Analytics needed to fill it without manual work |
|---|---|---|---|---|
|  | Rachunek zysków i strat (wariant porównawczy) ¹ | `RZiSPor` | Σ subpositions |  |
| A. | Przychody netto ze sprzedaży i zrównane z nimi | `A` | Σ subpositions |  |
| A.I. | Przychody netto ze sprzedaży | `A_I` | 7 | products + goods (annex 1 A.I + A.IV) |
| A.II. | Zmiana stanu produktów (zwiększenie – wartość dodatnia, zmniejszenie – wartość ujemna) | `A_II` | 6/5 | = annex 1 A.II |
| A.III. | Koszt wytworzenia produktów na własne potrzeby jednostki | `A_III` | 4/5 | = annex 1 A.III |
| B. | Koszty działalności operacyjnej | `B` | Σ subpositions |  |
| B.I. | Amortyzacja | `B_I` | 4 |  |
| B.II. | Zużycie materiałów i energii | `B_II` | 4 |  |
| B.III. | Usługi obce | `B_III` | 4 |  |
| B.IV. | Wynagrodzenia | `B_IV` | 4 |  |
| B.V. | Ubezpieczenia społeczne i inne świadczenia, w tym: | `B_V` | 4 |  |
| B.V. – | – emerytalne | `B_V_1` | 4 | pension flag |
| B.VI. | Pozostałe koszty, w tym: | `B_VI` | 4 + 7 | taxes and fees + other + value of goods sold |
| B.VI. – | – wartość sprzedanych towarów | `B_VI_1` | 7 | value of goods sold |
| C. | Zysk (strata) ze sprzedaży (A - B) | `C` | computed |  |
| D. | Pozostałe przychody operacyjne, w tym: | `D` | 7 | = annex 1 D |
| D. – | – aktualizacja wartości aktywów niefinansowych | `D_1` | 7 | = annex 1 D.III |
| E. | Pozostałe koszty operacyjne, w tym: | `E` | 7 | = annex 1 E |
| E. – | – aktualizacja wartości aktywów niefinansowych | `E_1` | 7 | = annex 1 E.II |
| F. | Przychody finansowe, w tym: | `F` | 7 | = annex 1 G |
| F.I. | Dywidendy i udziały w zyskach od jednostek, w których jednostka posiada zaangażowanie w kapitale, w tym: | `F_I` | 7 | dividends ONLY from investees with an equity stake (flag) |
| F.I. – | – od jednostek powiązanych, w których jednostka posiada zaangażowanie w kapitale | `F_I_1` | 7 | … of which related |
| F.II. | Odsetki, w tym: | `F_II` | 7 |  |
| F.II. – | – od jednostek powiązanych | `F_II_1` | 7 | CP = P |
| F.III. | Zysk z tytułu rozchodu aktywów finansowych, w tym: | `F_III` | 7 |  |
| F.III. – | – w jednostkach powiązanych | `F_III_1` | 7 | CP = P |
| F.IV. | Aktualizacja wartości aktywów finansowych | `F_IV` | 7 |  |
| G. | Koszty finansowe, w tym: | `G` | 7 | = annex 1 H |
| G.I. | Odsetki, w tym: | `G_I` | 7 |  |
| G.I. – | – dla jednostek powiązanych | `G_I_1` | 7 | CP = P |
| G.II. | Strata z tytułu rozchodu aktywów finansowych, w tym: | `G_II` | 7 |  |
| G.II. – | – w jednostkach powiązanych | `G_II_1` | 7 | CP = P |
| G.III. | Aktualizacja wartości aktywów finansowych | `G_III` | 7 |  |
| H. | Zysk (strata) brutto (C + D - E + F - G) ¹ | `H` | computed |  |
| I. | Podatek dochodowy | `I` | 8 | no deferred tax if art. 37 ust. 10 used |
| J. | Zysk (strata) netto (H - I) ¹ | `J` | 8 |  |

¹ element required by the XSD (minOccurs ≠ 0); all others may be omitted when empty (art. 50 ust. 3 UoR).

### 3.9 RZiS — załącznik nr 5, wariant kalkulacyjny (`RZiSJednostkaMala/RZiSKalk`)

| Poz. | Nazwa (annex text) | XSD element | Account groups (practice) | Analytics needed to fill it without manual work |
|---|---|---|---|---|
|  | Rachunek zysków i strat (wariant kalkulacyjny) ¹ | `RZiSKalk` | Σ subpositions |  |
| A. | Przychody netto ze sprzedaży produktów i towarów | `A` | 7 | products + goods |
| B. | Koszty sprzedanych produktów i towarów | `B` | 7 | cost of products sold + value of goods sold |
| C. | Koszty sprzedaży | `C` | 5 |  |
| D. | Koszty ogólnego zarządu | `D` | 5 |  |
| E. | Zysk (strata) ze sprzedaży (A - B - C - D) | `E` | computed |  |
| F. | Pozostałe przychody operacyjne, w tym: | `F` | 7 |  |
| F. – | – aktualizacja wartości aktywów niefinansowych | `F_1` | 7 | revaluation of non-financial assets |
| G. | Pozostałe koszty operacyjne, w tym: | `G` | 7 |  |
| G. – | – aktualizacja wartości aktywów niefinansowych | `G_1` | 7 | revaluation of non-financial assets |
| H. | Przychody finansowe, w tym: | `H` | 7 |  |
| H.I. | Dywidendy i udziały w zyskach od jednostek, w których jednostka posiada zaangażowanie w kapitale, w tym: | `H_I` | 7 | dividends ONLY from investees with an equity stake |
| H.I. – | – od jednostek powiązanych, w których jednostka posiada zaangażowanie w kapitale | `H_I_1` | 7 | … of which related |
| H.II. | Odsetki, w tym: | `H_II` | 7 |  |
| H.II. – | – od jednostek powiązanych | `H_II_1` | 7 | CP = P |
| H.III. | Zysk z tytułu rozchodu aktywów finansowych, w tym: | `H_III` | 7 |  |
| H.III. – | – w jednostkach powiązanych | `H_III_1` | 7 | CP = P |
| H.IV. | Aktualizacja wartości aktywów finansowych | `H_IV` | 7 |  |
| I. | Koszty finansowe, w tym: | `I` | 7 |  |
| I.I. | Odsetki, w tym: | `I_I` | 7 |  |
| I.I. – | – dla jednostek powiązanych | `I_I_1` | 7 | CP = P |
| I.II. | Strata z tytułu rozchodu aktywów finansowych, w tym: | `I_II` | 7 |  |
| I.II. – | – w jednostkach powiązanych | `I_II_1` | 7 | CP = P |
| I.III. | Aktualizacja wartości aktywów finansowych | `I_III` | 7 |  |
| J. | Zysk (strata) brutto (E + F - G + H - I) ¹ | `J` | computed |  |
| K. | Podatek dochodowy | `K` | 8 |  |
| L. | Zysk (strata) netto (J - K) ¹ | `L` | 8 |  |

¹ element required by the XSD (minOccurs ≠ 0); all others may be omitted when empty (art. 50 ust. 3 UoR).

### 3.10 Bilans — załącznik nr 4, jednostka mikro (`BilansJednostkaMikro`)

| Poz. | Nazwa (annex text) | XSD element | Account groups (practice) | Analytics needed to fill it without manual work |
|---|---|---|---|---|
|  | Aktywa razem ¹ | `Aktywa` | Σ |  |
| A. | Aktywa trwałe, w tym środki trwałe | `Aktywa_A` | 0 (+ long-term 2, 6) | all fixed assets; ‘w tym środki trwałe’ = annex 1 A.II.1 |
| B. | Aktywa obrotowe, w tym: | `Aktywa_B` | 1, 2, 3, 5, 6 | all current assets |
| B. – | - zapasy | `Aktywa_B_1` | 3, 5, 6 | = annex 1 B.I |
| B. – | - należności krótkoterminowe | `Aktywa_B_2` | 2 | = annex 1 B.II |
| C. | Należne wpłaty na kapitał (fundusz) podstawowy | `Aktywa_C` | 2/8 |  |
| D. | Udziały (akcje) własne | `Aktywa_D` | 1 |  |
|  | Pasywa razem ¹ | `Pasywa` | Σ |  |
| A. | Kapitał (fundusz) własny, w tym: | `Pasywa_A` | 8 |  |
| A. – | - kapitał (fundusz) podstawowy | `Pasywa_A_1` | 8 | = annex 1 A.I |
| B. | Zobowiązania i rezerwy na zobowiązania, w tym: | `Pasywa_B` | 1, 2, 6, 8 | all liabilities, provisions and accruals |
| B. – | - rezerwy na zobowiązania | `Pasywa_B_1` | 8/6 | = annex 1 B.I |
| B. – | - zobowiązania z tytułu kredytów i pożyczek | `Pasywa_B_2` | 1/2 | loans of all classes, LT + ST |

¹ element required by the XSD (minOccurs ≠ 0); all others may be omitted when empty (art. 50 ust. 3 UoR).

### 3.11 RZiS — załącznik nr 4, jednostka mikro (`RZiSJednostkaMikro`)

| Poz. | Nazwa (annex text) | XSD element | Account groups (practice) | Analytics needed to fill it without manual work |
|---|---|---|---|---|
| A. | Przychody podstawowej działalności operacyjnej i zrównane z nimi, w tym: | `A` | 7 (+ change in products) | core-activity revenue (products + goods) |
| A. – | - zmiana stanu produktów (zwiększenie - wartość dodatnia, zmniejszenie - wartość ujemna) | `A_1` | 6/5 | change in products |
| B. | Koszty podstawowej działalności operacyjnej | `B` | Σ 4 (+ 7 goods sold) |  |
| B.I. | Amortyzacja | `B_I` | 4 |  |
| B.II. | Zużycie materiałów i energii | `B_II` | 4 |  |
| B.III. | Wynagrodzenia, ubezpieczenia społeczne i inne świadczenia | `B_III` | 4 | wages + social insurance combined |
| B.IV. | Pozostałe koszty | `B_IV` | 4 + 7 | all other operating costs (practice: incl. services, taxes, goods sold — *to verify*) |
| C. | Pozostałe przychody i zyski, w tym: | `C` | 7 | other operating AND financial income together |
| C. – | - aktualizacja wartości aktywów | `C_1` | 7 | revaluation of assets |
| D. | Pozostałe koszty i straty, w tym: | `D` | 7 | other operating AND financial costs together |
| D. – | - aktualizacja wartości aktywów | `D_1` | 7 | revaluation of assets |
| E. | Podatek dochodowy | `E` | 8 |  |
| F. | Zysk/strata netto (A-B+C-D-E) ¹ | `F` | 8 | A − B + C − D − E |

¹ element required by the XSD (minOccurs ≠ 0); all others may be omitted when empty (art. 50 ust. 3 UoR).


### 3.12 Reading the annex 5 and annex 4 tables against annex 1

The simplified balance sheets are roll-ups of annex 1, with three cuts that annex 1 does **not** give and the books must therefore carry anyway:

1. **Loans (*kredyty i pożyczki*) across all counterparty classes.** Annex 5 shows long-term and short-term liabilities "z tytułu kredytów i pożyczek" in total, and annex 4 shows all loans in one line; annex 1 hides loans from related parties and from Z-class entities inside "Wobec jednostek powiązanych" / "inne". So every liability needs an instrument-type attribute even when its counterparty is related.
2. **Trade receivables and payables across all counterparty classes**, still split ≤ 12 / > 12 months.
3. **Dividends by equity stake, not by group membership.** Annex 5 shows "Dywidendy i udziały w zyskach od jednostek, w których jednostka posiada zaangażowanie w kapitale, w tym: – od jednostek powiązanych, w których jednostka posiada zaangażowanie w kapitale"; annex 1 splits by related / other and then by stake. Hence **related** and **equity stake** are two independent attributes of a counterparty, not one three-valued field (CP = P/Z/O is derived from them: P if related; Z if not related and with stake; O otherwise).

The micro RZiS merges operating and financial "other" items into C *Pozostałe przychody i zyski* and D *Pozostałe koszty i straty*; the chart needs nothing extra for it, as long as financial and other-operating accounts can be summed.

---

## 4. Notes (*informacja dodatkowa*) that need account-level data

The notes are text and attached tables in the XML (§2.5), but most of them are numbers that should come straight from the books. Below, every numbered item of the annexes that needs ledger data, with the analytic dimension that produces it. Items that are pure narrative (policies, going-concern description, mergers' descriptions) are listed only where they need a figure.

### 4.1 Annex 1 — *Dodatkowe informacje i objaśnienia* (full entities and small entities not using annex 5)

Text of annex 1, UoR pp. 93–97 ([Dz.U. 2026 poz. 522](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260000522)).

| Item | Requirement (abridged, Polish terms quoted) | Data the books must hold |
|---|---|---|
| 1.1 | "szczegółowy zakres zmian wartości grup rodzajowych środków trwałych, wartości niematerialnych i prawnych oraz inwestycji długoterminowych" — opening, increases/decreases by "aktualizacji wartości, nabycia, rozchodu, przemieszczenia wewnętrznego", closing; same for accumulated depreciation | MOV tag on every posting to gross value, depreciation and impairment of groups 0 accounts, per "grupa rodzajowa" (balance-sheet sub-line or KŚT group) |
| 1.2 | impairment write-downs of fixed assets in the year, separately for non-financial and financial long-term assets | impairment accounts split non-financial / financial |
| 1.3 | completed development costs and goodwill, with write-off periods (art. 33 ust. 3, art. 44b ust. 10) | asset kind + useful life on the asset register |
| 1.4 | "wartość gruntów użytkowanych wieczyście" | PWUG flag on land |
| 1.5 | non-depreciated fixed assets used under rental, lease and other contracts (incl. leasing) | **off-balance** register of leased-in assets with value (group 9 or off-balance accounts — practice) |
| 1.6 | number and value of securities and rights held (incl. convertibles, warrants, options) with the rights they give | instrument register (quantity, value) |
| 1.7 | receivable allowances: opening, increases, use, release, closing | MOV on allowance accounts |
| 1.8 | ownership structure of share capital; number and nominal value of subscribed shares, incl. preferred | shareholder register (not ledger) linked to share-capital account |
| 1.9 | movements of kapitał zapasowy, rezerwowy and z aktualizacji wyceny — **only if no SCE is prepared** | MOV on equity accounts (same as SCE) |
| 1.10 | proposed profit distribution / loss coverage | management decision data (feeds SCE III) |
| 1.11 | provisions by purpose: opening, increases, use, release, closing | provision purpose + MOV |
| 1.12 | long-term liabilities by remaining contractual maturity: ≤ 1 year, > 1–3, > 3–5, > 5 years | repayment schedule (DUE per instalment) on each liability |
| 1.13 | total liabilities secured on the entity's assets, nature and form of security | secured flag + collateral description on the liability |
| 1.14 | material accruals and deferrals; RMK from discount on financial liabilities (art. 39 ust. 4) | RMK/RMP analytics by nature |
| 1.15 | items shown in more than one balance-sheet position (esp. LT/ST split of receivables and liabilities) | DUE-based LT/ST split, derivable automatically |
| 1.16 | contingent liabilities incl. guarantees and sureties (also bills), separately: secured ones, pensions, towards related or associated entities | **off-balance** register of contingent liabilities with CP and security |
| 1.17 | non-financial assets at fair value: assumptions, FV and revaluation effects by category, table of revaluation-reserve movements | FV flag; revaluation postings tagged; MOV on revaluation reserve |
| 1.17a | hidden profits (*ukryte zyski*) under the Estonian CIT (art. 28m ust. 1 pkt 2 CIT) | tag on postings (strand (a)) |
| 1.18 | cash on the VAT account (split payment), art. 62a Prawo bankowe / art. 3b SKOK act | VAT account as its own cash sub-account |
| 1.19 | PSA shares issued for non-cash contributions (labour, services, non-transferable rights) | share register |
| 1.20 | statement that art. 37 ust. 12 applies (no deferred tax on top-up taxes) | flag |
| 2.1 | revenue by type of activity and by geographical market, where materially different | revenue analytics: product/service line + market (country/region) |
| 2.2 | **by-function RZiS only**: own-use production costs and costs by nature (amortyzacja, materiały i energia, usługi obce, podatki i opłaty, wynagrodzenia, ubezpieczenia i inne świadczenia w tym emerytalne, pozostałe) | nature-of-cost data even when costs are booked by function (group 4, or nature analytic on group 5) |
| 2.3 | fixed-asset impairment amounts and reasons | impairment postings with reason |
| 2.4 | inventory write-downs | allowance accounts on inventory |
| 2.5 | revenue, costs and result of activities discontinued or to be discontinued | discontinued-activity flag (art. 47 ust. 3) |
| 2.6 | reconciliation of taxable base to gross result | structured P_ID table (§2.5), strand (a) |
| 2.7 | cost of fixed assets under construction, incl. interest and FX capitalised in the year | capitalised interest / FX tags on group 0 investment accounts |
| 2.8 | interest and FX capitalised into goods or products in the year | same tags on inventory |
| 2.9 | capex on non-financial fixed assets: incurred last year and **planned** next year; environmental capex separately | capex postings tagged; environmental flag; plan figures (not ledger) |
| 2.10 | income or cost items of extraordinary value or incidental | incidental flag |
| 2.11 | research and development costs not capitalised (art. 33 ust. 2) | R&D flag on cost postings |
| 2.12 | food handed to NGOs / food-waste fee | tag |
| 2.13 | global, domestic and UTPR top-up taxes for the period | tax-type analytic on income-tax accounts |
| 3 | FX rates used to measure items in foreign currencies | CUR + rate per valuation |
| 4 | structure of cash for the cash-flow statement; for the direct method the indirect reconciliation of operating cash flow; reasons for differences between balance-sheet changes and cash-flow changes | cash definition, CF tags, the exclusions in §3.5 |
| 5.1 | off-balance arrangements (nature and business purpose) | off-balance register |
| 5.2 | transactions with **related parties as defined in IAS** (*strony powiązane*, Reg. 1606/2002) **on non-market terms**, with amounts and nature of the relationship | a second, IAS 24 related-party flag on counterparties (wider than UoR "jednostki powiązane": includes key management and their families) + non-arm's-length flag on the transaction |
| 5.3 | average employment by occupational group | HR data |
| 5.4 | remuneration (incl. profit-based) paid or due to members of management, supervisory and administrative bodies, per body; pension obligations to former members | payroll analytic: body (zarząd / rada nadzorcza / other) |
| 5.5 | advances, credits, loans and similar benefits to members of those bodies, with terms, interest, repayments, write-offs; guarantees given on their behalf | CP flag "member of body" on receivables and guarantees |
| 5.6 | audit firm fees for the year, separately: statutory audit, other assurance, tax advisory, other | usługi obce analytic by audit-firm service type |
| 6.1 | revenues and costs from prior-year errors booked to equity, amounts and kind | error-correction tag (KSR 7) |
| 6.2 | material events after the balance-sheet date | narrative + amounts |
| 6.3 | changes of accounting policy, their effect on result and equity | policy-change tag (KSR 7) |
| 6.4 | figures making the prior year comparable | restated comparatives (`KwotaB1`) |
| 7.1 | joint ventures not consolidated: share, assets, liabilities, revenues/costs, commitments | joint-venture dimension on postings |
| 7.2 | transactions with **jednostki powiązane** | CP = P on all postings (already needed for the face) |
| 7.3 | list of companies where the entity holds an equity stake or ≥ 20 % of votes, with % held, their equity and net result | investee register (not ledger) |
| 7.4–7.6 | consolidation exemptions; parent entities preparing consolidated statements; entities where the unit is an unlimited partner | group master data |
| 8 | mergers: acquired company, shares issued, price, net assets at FV, goodwill | merger data |
| 9 | going-concern uncertainties | narrative |
| 10 | other material information | narrative |

### 4.2 Annex 5 — notes of a small entity using the simplifications

Same source, UoR pp. 117–118. Items needing ledger data, with the annex 1 counterpart:

| Item | Requirement (abridged) | = annex 1 item / difference |
|---|---|---|
| 1 | movement table of fixed assets, WNiP, long-term investments and their depreciation | = 1.1 |
| 2 | impairment of fixed assets, non-financial vs financial | = 1.2 |
| 3 | goodwill and its write-off period | part of 1.3 |
| 4 | fair-value items (financial **and** non-financial): assumptions; FV and revaluation effects **separately for long- and short-term**; derivatives shown as ST assets/liabilities; revaluation-reserve table | wider than 1.17 (includes financial instruments, since small entities may skip the instruments regulation, art. 28b) |
| 5 | RMK from discount on financial liabilities | part of 1.14 |
| 6 | long-term liabilities with remaining maturity **> 5 years**; secured liabilities | narrower than 1.12; = 1.13 |
| 7 | items in more than one balance-sheet position | = 1.15 |
| 8 | contingent liabilities and guarantees, with the separate categories | = 1.16 |
| 9 | cost of fixed assets under construction incl. capitalised interest/FX | = 2.7 |
| 10 | interest/FX capitalised into inventory | = 2.8 |
| 11 | incidental or extraordinary-value items | = 2.10 |
| 12 | average employment (no split by group) | narrower than 5.3 |
| 13 | advances and loans to members of bodies | = 5.5 |
| 14 | events after the balance-sheet date | = 6.2 |
| 15 | policy changes | = 6.3 |
| 16 | comparability figures | = 6.4 |
| 17 | parent preparing consolidated statements at the lowest level | part of 7.5 |
| 18 | own shares (if no management report, art. 49 ust. 5) | art. 49 ust. 2 pkt 5 data |

### 4.3 Annex 4 — *Informacje uzupełniające do bilansu* of a micro entity

UoR pp. 113–114: (1) all financial liabilities incl. debt instruments, guarantees, sureties and contingent liabilities not in the balance sheet, with secured items, pensions and related/associated entities separately; (2) advances and credits to members of bodies; (3) own shares (reason, number, nominal, consideration, holdings). Same data as annex 1 items 1.16, 5.5 and art. 49 ust. 2 pkt 5.

### 4.4 Disclosures under the financial-instruments regulation (full entities)

[Rozp. MF 17 Nov 2024 (Dz.U. 2024 poz. 1750)](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20240001750), § 45 ust. 1: for **all** financial assets and liabilities **by category** (§ 4 ust. 1: *przeznaczone do obrotu*; *pożyczki udzielone i należności własne*; *utrzymywane do terminu wymagalności*; *dostępne do sprzedaży*; *zobowiązania finansowe inne niż przeznaczone do obrotu*): number, value and terms; fair-value methods and the level 1/2/3 hierarchy; revaluation-reserve table for instruments; interest-rate and credit-risk information. § 39: instruments split LT/ST on the balance sheet. § 42: revaluation and interest presented by measurement basis. § 45 ust. 8: impairment and reversals by category; ust. 9: carrying amount of financial assets pledged as security. Micro and small may skip the regulation (art. 28a, 28b UoR). **Chart consequence:** a category attribute and a fair-value-level attribute on every financial instrument account or open item.

### 4.5 Disclosures required by KSRs

| KSR | Disclosure needing ledger data | Source |
|---|---|---|
| KSR 1 | cash structure; for the direct method the indirect reconciliation; explanation of differences between balance-sheet and cash-flow changes; classification of interest and commissions if moved to operating | KSR 1 pkt 4.2, 4.4, 5.6, 8.9–8.12 |
| KSR 2 (2024) | deferred-tax provision movement (P&L vs equity vs goodwill vs other), closing split; deferred-tax asset details: from tax losses (with expiry dates), from unused reliefs, from temporary differences, allowances; reconciliation; recommended asset movement table | KSR 2 pkt 14.1–14.3 |
| KSR 3 | per construction contract: revenues and costs by the KSR 3 method, invoiced receivables, costs incurred, provisions for losses | KSR 3 pkt 9.1–9.8 |
| KSR 5 | operating leases: value of leased-in assets, future payments ≤ 1 / 1–3 / > 3 years, sublease income, contingent payments; finance leases: method of splitting capital/interest | KSR 5 pkt 12.1.1–12.2 |
| KSR 6 | per group of provisions and bierne RMK: purpose, opening, increases (incl. unwinding of discount), used, released, closing; contingent liabilities with estimated effect | KSR 6 pkt 5.1–5.4 |
| KSR 7 | per prior-year error: kind, revenues/costs behind the equity correction, correction for the current year and each prior year | KSR 7 pkt 5.8 |
| KSR 11 | recommended: FV of fixed assets if materially different; gross value of fully depreciated assets in use; NBV of temporarily idle assets | KSR 11 pkt 11.3 |
| KSR 15 (2026) | revenue by type and by market; may be by goods type, geography, customer group, performance obligation | KSR 15 pkt 8.2 |

---

## 5. The KSRs that shape a chart of accounts

KSRs may be applied in matters the Act does not regulate (art. 10 ust. 3 UoR). The Ministry's page lists **sixteen** ([gov.pl/web/finanse/krajowe-standardy-rachunkowosci](https://www.gov.pl/web/finanse/krajowe-standardy-rachunkowosci), page version 8.0 of 24 Apr 2026). Publication data below are from the journal header of each PDF downloaded from that page.

| KSR | Title | Current text (journal) | Applies | Separate accounts or analytics it implies |
|---|---|---|---|---|
| 1 | Rachunek przepływów pieniężnych | update, uchwała 16/2017; Dz. Urz. MRiF 2017 poz. 216 ([PDF](https://www.gov.pl/attachment/c21830fc-9205-4f33-8ddd-a0cdd5d30099)) | — | **CF tag on every cash/bank posting** (direct data can come "ze szczegółowej ewidencji księgowej prowadzonej do kont środków pieniężnych", pkt 5.5); cash vs cash-equivalent (≤ 3 m) flag; restricted-cash flag; separate **investment receivables and payables** (excluded from working-capital changes, pkt 8.4 items 7–8); interest and dividends tagged operating/investing/financing; FX differences split realised/unrealised and by activity (pkt 4.2, 6.6, 7.4); VAT on investing flows: investing lines net of VAT, VAT kept in operating when the direct method is used (pkt 6.4–6.5); non-cash transactions (leases, contributions in kind, debt-to-equity) flagged to keep them out (pkt 4.8) |
| 2 | Podatek dochodowy | updated, uchwała 9/2024; Dz. Urz. MF 2024 poz. 81 ([PDF](https://www.gov.pl/attachment/ab289f43-2138-4903-a88d-41db3342c656)) | FYs from 1 Jan 2025, earlier allowed | separate accounts: **current tax** (P&L), **deferred tax** (P&L), deferred tax charged to equity; **deferred-tax asset** and **deferred-tax provision** kept apart (offset only on presentation, art. 37 ust. 7); asset analytics: from losses (with expiry), from unused reliefs, from temporary differences, allowance on the asset; per temporary difference the book and tax values (needs tax-value attribute on assets/liabilities — strand (a)); income-tax receivable vs payable |
| 3 | Niezakończone usługi budowlane | update, uchwała 14/2018; Dz. Urz. MF 2018 poz. 76 ([PDF](https://www.gov.pl/attachment/8b0eaa9b-1295-4751-99d2-5611fe36dc56)) | — | per contract: estimated revenue, invoiced, costs incurred, estimated costs; four balance accounts shown gross ("w rozwartym szyku", pkt 8.10): *Należności niezafakturowane*, *RMK czynne* (both in a recommended extra line "Aktywa z tytułu niezakończonych umów budowlanych" under aktywa B.IV), *RMP*, *RMK bierne* (in "Rozliczenia międzyokresowe umów budowlanych" under pasywa B.IV between B.IV.1 and B.IV.2) (pkt 9.1–9.6); provision for expected loss on a contract (as bierne RMK, pkt 8.9) |
| 4 | Utrata wartości aktywów | uchwała 2/12; Dz. Urz. MF 2012 poz. 15 ([PDF](https://www.gov.pl/attachment/4c74806b-00ac-4a6b-a00c-27c5f6cc9646)) | — | impairment-allowance accounts per asset class and per cash-generating unit; reversal postings kept apart (art. 35c); presentation: non-financial → other operating, financial → financial (art. 3 ust. 1 pkt 32 lit. f, art. 35b–35c UoR). The standard prescribes no accounts (no account references found in its text) |
| 5 | Leasing, najem i dzierżawa | update, uchwała 13/2018; Dz. Urz. MF 2018 poz. 77 ([PDF](https://www.gov.pl/attachment/07da18b3-1ae8-4fde-9239-7e264ed39e24)) | — | finance-leased assets in fixed assets with an own-vs-leased flag; lease liability split LT/ST and capital/interest; operating-lease assets **off-balance** at the lessee; future-payment buckets (≤ 1, 1–3, > 3 years); lessor: net investment as financial asset, LT/ST; contingent rents (pkt 12) |
| 6 | Rezerwy, bierne rozliczenia międzyokresowe kosztów, zobowiązania warunkowe | uchwała 3/2014; Dz. Urz. MF 2014 poz. 12 ([PDF](https://www.gov.pl/attachment/bc48b7e4-7102-4458-bdd9-545997b2e38b)) | — | provisions and bierne RMK by **purpose** and by **LT/ST** (B.I.2 retirement and similar; B.I.3 other, incl. art. 39 ust. 2 pkt 2 accruals); counter-entry depends on nature: operating cost (by function), other operating, financial (pkt 3.4); MOV created/used/released; contingent-liability register off-balance. Note: pkt 3.4 lit. d still mentions *straty nadzwyczajne*, which no longer exist in annex 1 for non-financial entities |
| 7 | Zmiany zasad (polityki) rachunkowości, wartości szacunkowych, poprawianie błędów, zdarzenia następujące po dniu bilansowym – ujęcie i prezentacja | update, uchwała 3/2018; Dz. Urz. MF 2019 poz. 2 ([PDF](https://www.gov.pl/attachment/9da792cf-ca83-4ace-9779-17aa810a8548)) | — | material prior-year errors and policy changes booked to **zysk (strata) z lat ubiegłych** with a tag (feeds SCE I.–, 5.1.–, 5.4.–) and restated comparatives; immaterial ones through the current P&L (pkt 5.3–5.7) |
| 8 | Działalność deweloperska | update 2020, uchwała 10/2020; Dz. Urz. MFFiPR 2020 poz. 38 ([PDF](https://www.gov.pl/attachment/585caf0f-71da-4bd0-9601-90179ba93884)) | — | per development project; land in a project shown in "Półprodukty i produkty w toku" with extra line "w tym: Obiekty w zabudowie", also in RZiS (pkt 10.1); land sale classified by the land's prior classification (goods vs fixed asset) (pkt 10.2) |
| 9 | Sprawozdanie z działalności | update, file dated 10 Jan 2018 ([PDF](https://www.gov.pl/attachment/0b95026a-9043-4682-9067-77ecd9f985f8)); the Committee is working on an amendment after the 6 Dec 2024 act ([KSR page](https://www.gov.pl/web/finanse/krajowe-standardy-rachunkowosci)) | — | KPIs from the books (not read in detail — *to verify*) |
| 10 | Umowy o partnerstwie publiczno-prywatnym oraz umowy koncesji na roboty budowlane lub usługi | update, file dated 2 Aug 2017 ([PDF](https://www.gov.pl/attachment/33853a90-f477-46db-b54e-6a69d9fa87aa)) | — | sector-specific; not read in detail (*to verify* if a PPP client appears) |
| 11 | Środki trwałe | uchwała 4/2017; Dz. Urz. MRiF 2017 poz. 105 ([PDF](https://www.gov.pl/attachment/92b030f0-544b-4114-b877-f050ea186ebf)) | — | asset register per **inventory object** (and components), with KŚT number, rate, dates; recommended classification cuts: **KŚT group, own vs finance-leased, place of use** (pkt 10.5); register must serve deferred tax and tax needs (pkt 10.6); disposal result = proceeds − NBV − costs, net, into D.I/E.I (pkt 11.1); liquidated assets kept off-balance until disposed (pkt 9.7); document types OT/LT/MT/PK-AM (pkt 10.2) |
| 12 | Działalność rolnicza | file dated 26 Jul 2018 ([PDF](https://www.gov.pl/attachment/fce99e23-d840-4108-a855-eab81b8f05ee)) | — | sector-specific; not read (*to verify* if farming clients appear) |
| 13 | Koszt wytworzenia jako podstawa wyceny produktów | uchwała 6/2019; Dz. Urz. MF 2019 poz. 60 ([PDF](https://www.gov.pl/attachment/5761f089-bda5-4a79-9d16-f27b204ce989)) | — | production costs by cost object and **cost centre**; indirect production costs split **variable vs fixed** (or planned rates at normal capacity) so unabsorbed fixed overhead goes to the period, not to inventory (pkt 8.9–8.10) |
| 14 | Kontynuacja działalności oraz rachunkowość jednostek przy braku kontynuowania działalności | uchwała 6/2021; Dz. Urz. MFFiPR 2021 poz. 119 ([PDF](https://www.gov.pl/attachment/69f68d95-61f2-4bc2-a72e-7632a076d655)) | — | on loss of going concern: remeasurement and the art. 29 provision for liquidation costs; otherwise no standing accounts (not read in detail) |
| 15 | Przychody ze sprzedaży wyrobów, półproduktów, towarów i materiałów | update 2026, uchwała 8/2026; Dz. Urz. MFiG 2026 poz. 53 ([PDF](https://www.gov.pl/attachment/143e3deb-6668-4506-b191-b0efcce109f7)) | statements with balance-sheet date after its entry into force (published 22 Apr 2026), earlier allowed | revenue separated **products / goods / materials** (materials to other operating, gross or net per policy, pkt 8.1); revenue analytics by type and market (pkt 8.2); performance obligations and allocation of price (examples) |
| 16 | Uproszczenia dopuszczone ustawą o rachunkowości | uchwała 13/2025; Dz. Urz. MFiG 2025 poz. 29 ([PDF](https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a)) | FYs from 1 Jan 2026, earlier allowed | which accounts a micro/small entity may leave unused (fixed assets, intangibles, financial instruments, inventories, provisions, deferred tax), and materiality thresholds to document |

---

## 6. The analytic dimensions the chart must carry

Collected from §3–§5. Each row is one attribute that, when present on postings or open items, removes a manual step. "Where" names the statement positions that need it.

| # | Dimension (values) | Kept on | Where it is needed | Legal basis |
|---|---|---|---|---|
| D1 | **Related** (*jednostka powiązana*: same group) — yes/no | counterparty master | Bilans A.III.1, A.IV.3.a, B.II.1, B.III.1.a, B.II.1/B.III.1 pasywa; RZiS A "w tym od jednostek powiązanych", B (kalk.), financial income/costs "w tym"; RPP B.I.3.a, B.II.3.a; note 7.2 | art. 3 ust. 1 pkt 43–44 |
| D2 | **Equity stake held** (*zaangażowanie w kapitale*) — yes/no | counterparty master | Bilans "pozostałe jednostki, w których jednostka posiada zaangażowanie w kapitale"; RZiS dividends "w tym"; annex 5 dividends; note 7.3 | art. 3 ust. 1 pkt 37d |
| D3 | **IAS related party** (*strona powiązana*) — yes/no, nature of relation; plus non-market-terms flag on the transaction | counterparty master / transaction | note 5.2 | annex 1 pt 5.2 |
| D4 | **Member of a governing body** (zarząd / rada nadzorcza / other) | counterparty or employee master | notes 5.4, 5.5; annex 4 pt 2; annex 5 pt 13 | annex 1 pt 5.4–5.5 |
| D5 | **Due date** per open item and repayment schedule per contract | open items | trade ≤ 12 / > 12 m; LT/ST split of non-trade items; note 1.12 buckets (≤ 1, 1–3, 3–5, > 5 y); annex 5 > 5 y | art. 3 ust. 1 pkt 18c, 22; annex 1 pt 1.12 |
| D6 | **Trade vs non-trade** settlement; **advance** paid/received and its object (WNiP / ŚT w budowie / supplies) ; **investment** receivable/payable (fixed-asset purchase or sale) | settlement accounts / open items | balance-sheet lines A.I.4, A.II.3, B.I.5, B.III.3.e pasywa; RPP A.II.7–8 exclusions | art. 3 ust. 1 pkt 18c, 22; KSR 1 pkt 8.4 |
| D7 | **Instrument type** (shares / other securities / loan / other; credit / loan / debt security / bill / other financial / lease) | financial assets and liabilities | Bilans A.IV.3, B.III.1, B.II.3, B.III.3; annex 5 and 4 loan lines | annexes 1, 4, 5 |
| D8 | **Financial-instrument category** (5 categories) and **FV level** (1/2/3) | financial instruments | notes under rozp. 2024 § 45 | rozp. MF 17 Nov 2024 |
| D9 | **Currency** and FX rate used | open items, cash accounts | note 3; FX differences | art. 30; annex 1 pt 3 |
| D10 | **FX difference type**: realised / unrealised; on cash / operating / investing / financing item; capitalised | FX postings | RZiS G.V/H.IV (net), RPP A.II.2, E.–; notes 2.7–2.8 | art. 30 ust. 4; KSR 1 |
| D11 | **Fixed-asset class** (KŚT group → balance-sheet line a–e), own vs finance-leased, place of use, PWUG flag, investment-property flag | asset register | Bilans A.II.1.a–e, A.IV.1; notes 1.1, 1.4 | art. 17 ust. 1 pkt 1; KSR 11 pkt 10.5 |
| D12 | **Movement type** for assets: BO, acquisition, revaluation, disposal, internal transfer, depreciation, impairment, reversal, BZ | postings to group 0 | note 1.1, 1.2; RPP B lines | annex 1 pt 1.1 |
| D13 | **Equity movement type**: issue, redemption, agio, statutory allocation, above-statutory allocation, loss coverage, revaluation increase/decrease (sale of assets), policy change, error correction, profit carried, dividend, in-year charge | postings to group 8 equity | SCE all lines | annex 1 SCE; art. 36, 36a |
| D14 | **Provision purpose** and **movement** (created / used / released / unwinding), LT/ST | provision accounts | Bilans B.I.1–3; note 1.11; KSR 6 | art. 35d; KSR 6 |
| D15 | **Accrual/deferral nature** and term (RMK czynne/bierne, RMP: grants, prepaid revenue, negative goodwill, KSR 3 contract balances), LT/ST | group 6/8 | Bilans A.V, B.IV, pasywa B.IV; RPP A.II.9; note 1.14 | art. 39, 41; KSR 3 |
| D16 | **Deferred-tax origin** (temporary difference / tax loss with expiry / relief) and **P&L vs equity** | deferred-tax accounts | Bilans A.V.1, B.I.1; RZiS income tax; KSR 2 notes | art. 37; KSR 2 |
| D17 | **Cost by nature** (7 annex groups, excise, pension contributions) **and by function** (production / selling / administration, cost centre, variable vs fixed overhead) | cost postings | RZiS por. B.I–VII; RZiS kalk. B, D, E; note 2.2; KSR 13 | art. 47; KSR 13 |
| D18 | **Revenue type** (products-services / goods / materials) and **market** (country or region) and product line | revenue postings | RZiS A lines; note 2.1; KSR 15 pkt 8.2 | art. 3 ust. 1 pkt 30a, 32b |
| D19 | **Other-operating type**: fixed-asset disposal (both legs), grant, revaluation of non-financial assets, other (incl. materials sold) | group 7 other-operating | RZiS D/E (por.), G/H (kalk.) | art. 3 ust. 1 pkt 32 |
| D20 | **Financial income/cost type**: dividend, interest, disposal gain/loss, revaluation, FX, other | group 7 financial | RZiS G/H (por.), J/K (kalk.); RPP A.II.3–4 | art. 42 ust. 3 |
| D21 | **Cash-flow line tag** on every cash/bank posting; cash-equivalent and restricted flags; VAT account | cash and bank accounts | RPP all lines (direct); B and C (both methods); E–G; note 1.18 | art. 48b; KSR 1 |
| D22 | **Discontinued activity** flag | revenue/cost postings | note 2.5 | art. 47 ust. 3 |
| D23 | **Incidental / extraordinary-value** flag | any posting | note 2.10 / annex 5 pt 11 | annex 1 pt 2.10 |
| D24 | **Capitalised interest/FX**, **environmental capex**, **R&D expensed** flags | investment and cost postings | notes 2.7, 2.8, 2.9, 2.11 | annex 1 |
| D25 | **Auditor-fee service type** (statutory audit / other assurance / tax advice / other) | usługi obce postings | note 5.6 | annex 1 pt 5.6 |
| D26 | **Litigation** flag on receivables | receivables | Bilans B.II.3.d | annex 1 |
| D27 | **Public-law settlement** type (VAT, CIT, PIT, ZUS, customs, grants) and sign | group 2 public-law accounts | Bilans B.II.3.b vs B.III.3.g | art. 7 ust. 3; annex 1 |
| D28 | **Off-balance registers**: leased-in assets, contingent liabilities, guarantees, liquidated assets | off-balance accounts (group 9 in many charts — practice) | notes 1.5, 1.16, 5.1; annex 4 pt 1; KSR 5, 6, 11 | annex 1; KSR 11 pkt 9.7 |
| D29 | **Restatement**: policy change / error tag with the year it belongs to | equity and P&L postings | `KwotaB1`; SCE I.–; notes 6.1, 6.3, 6.4 | KSR 7 |
| D30 | **Project / contract** (construction, development, joint venture) | postings | KSR 3, KSR 8 notes; note 7.1 | KSR 3, 8 |

Non-ledger inputs the statements also need (so the model must store them somewhere): average FTE (P_8, note 5.3), audit flag (P_9), proposed profit distribution (SCE III, note 1.10), shareholder and share-class register (note 1.8), investee list with their equity and result (note 7.3), capex plan (note 2.9), events after the balance-sheet date (note 6.2).

---

## 7. Account groups 0–8 and the positions they feed (summary)

| Group (practice) | Feeds (annex 1 positions) |
|---|---|
| 0 aktywa trwałe | Bilans A.I, A.II, A.IV (and advances A.I.4, A.II.3 where kept here); depreciation/impairment netted in |
| 1 środki pieniężne, krótkoterminowe aktywa finansowe, kredyty bankowe | Bilans B.III, aktywa D (own shares, practice), pasywa B.II.3.a / B.III.3.a (bank credit, practice); every RPP line |
| 2 rozrachunki i roszczenia | Bilans A.III, B.I.5 (advances), B.II, pasywa B.II, B.III (except special funds), aktywa C (practice) |
| 3 materiały i towary | Bilans B.I.1, B.I.4 |
| 4 koszty według rodzajów | RZiS por. B.I–B.VII; note 2.2 for the by-function variant |
| 5 koszty według typów działalności | RZiS kalk. B.I (via 6/7), D, E; Bilans B.I.2 (WIP balance) |
| 6 produkty i rozliczenia międzyokresowe | Bilans B.I.2–3, A.V, B.IV; pasywa B.IV / B.I (accruals shown as provisions) |
| 7 przychody i koszty ich osiągnięcia | RZiS A, B.VIII (por.) / B (kalk.), other operating, financial |
| 8 kapitały, fundusze specjalne, rezerwy, RMP, wynik | Bilans pasywa A, B.I, B.III.4, B.IV; RZiS income tax and net result; SCE |
| 9 (no fixed use; often off-balance or management accounting) | notes 1.5, 1.16 (off-balance registers) — practice varies |

The exact accounts inside each group are strand (b)'s subject.

---

## 8. What remains to verify

1. **Which entities "zobowiązane" fill the structured CIT note** (`InformacjaDodatkowaDotyczacaPodatkuDochodowego`), especially micro and small entities using annexes 4/5 — the XSD offers it in all three forms; annexes 4 and 5 do not list it. Ask the MF e-sprawozdania channel or find the current MF FAQ (strand (a) overlap).
2. **Maturity basis for the "do 12 / powyżej 12 miesięcy" split of trade receivables and payables** — remaining term from the balance-sheet date or the original payment term. The law says "o okresie spłaty" / "o okresie wymagalności"; KSR 1 pkt 8.4 speaks of payment terms "od dnia powstania"; practice to be confirmed before fixing the rule in Rhea.
3. **Mapping of KŚT groups to balance-sheet lines A.II.1.a–e** — used in practice, not stated in the Act; to be written into Rhea's accounting policy template.
4. **Where bierne RMK under art. 39 ust. 2 pkt 1** (services received, not yet invoiced) are presented — art. 39 ust. 2a moves only pkt 2 items to provisions; practice shows pkt 1 items either in pasywa B.IV.2 or among liabilities.
5. **Advances paid for intangibles and for fixed assets under construction** — group 0 or group 2 in the chart; both are used in practice.
6. **Micro RZiS line B.IV "Pozostałe koszty"** — whether services, taxes and the value of goods sold all belong there (by elimination they must, since annex 4 has no other line; KSR 16 has a worked example not read in full).
7. **Exact names and dates of the superseded e-sprawozdanie versions** (v1-2, v1-3) — only secondary sources found.
8. **ASI forms republished in Feb 2026** (SFASIZ(2)/SFASIT(2), CRWDE 14162/14163) — what changed; not relevant unless an ASI client appears.
9. **KSR 9, 10, 12, 14** — not read in detail; only titles and journal data confirmed.
10. **Pending legislation** — no bill changing thresholds or annexes found; check RCL and the Sejm before FY 2027.
11. **Small audited entities without CF/SCE** — follows from art. 45 ust. 3 as now worded; no MF statement found addressing exactly this case.

---

## 9. Sources

Primary, law and official:
- Ustawa o rachunkowości, consolidated text: [Dz.U. 2026 poz. 522 (ISAP)](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260000522), [PDF via ELI API](https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf), [ELI metadata incl. later amendments](https://api.sejm.gov.pl/eli/acts/DU/2026/522)
- Amendments: [Dz.U. 2024 poz. 1863](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20240001863), [Dz.U. 2024 poz. 1685](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20240001685), [Dz.U. 2026 poz. 333](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260000333), [Dz.U. 2026 poz. 640](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260000640), [Dz.U. 2026 poz. 644](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260000644), [Dz.U. 2026 poz. 1206](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20260001206)
- [Rozporządzenie MF z 17 listopada 2024 r. w sprawie uznawania i metod wyceny oraz ujawniania i prezentacji instrumentów finansowych, Dz.U. 2024 poz. 1750](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20240001750)
- KAS, [Struktury e-Sprawozdań](https://www.gov.pl/web/kas/struktury-e-sprawozdan); CRWDE forms [13817](https://crd.gov.pl/wzor/2025/08/07/13817/schemat.xsd), [13810](https://crd.gov.pl/wzor/2025/08/04/13810/schemat.xsd), [13818](https://crd.gov.pl/wzor/2025/08/07/13818/schemat.xsd), [13846](https://crd.gov.pl/wzor/2025/09/10/13846/schemat.xsd), [13821](https://crd.gov.pl/wzor/2025/08/12/13821/schemat.xsd), [13819](https://crd.gov.pl/wzor/2025/08/11/13819/schemat.xsd); domain schemas [JednostkaInna v2-0E](http://crd.gov.pl/xml/schematy/dziedzinowe/mf/2025/07/31/eD/JednostkaInnaStruktury/JednostkaInnaStrukturyDanychSprFin_v2-0E.xsd), [JednostkaMala v2-0E](http://crd.gov.pl/xml/schematy/dziedzinowe/mf/2025/07/31/eD/JednostkaMalaStruktury/JednostkaMalaStrukturyDanychSprFin_v2-0E.xsd), [JednostkaMikro v2-0E](http://crd.gov.pl/xml/schematy/dziedzinowe/mf/2025/07/31/eD/JednostkaMikroStruktury/JednostkaMikroStrukturyDanychSprFin_v2-0E.xsd), [StrukturyDanychSprFin v2-0E](http://crd.gov.pl/xml/schematy/dziedzinowe/mf/2025/07/31/eD/StrukturyDanychSprFin/StrukturyDanychSprFin_v2-0E.xsd)
- MF, [notice of 27 Nov 2025 on new structures](https://podatki-arch.mf.gov.pl/komunikaty-techniczne/nowe-struktury-logiczne-dla-sprawozdan-finansowych-271125/?altTemplate=ArticlePdf&download=True); [notice of 21 Jan 2026, updated 26 May 2026](https://www.podatki.gov.pl/aktualnosci/komunikat-w-sprawie-nowych-struktur-logicznych-dla-sprawozdan-finansowych)
- MF, [Pytania i odpowiedzi — webinar 15 Jan 2025 on the 6 Dec 2024 amendment](https://www.gov.pl/attachment/f9b05862-99e1-473c-95e4-c1d1c21301cb); [sale of materials after the amendment](https://www.gov.pl/web/finanse/jak-zaprezentowac-w-rachunku-zyskow-i-strat-sprzedaz-materialow-po-nowelizacji-ustawy-o-rachunkowosci-ustawa-z-dnia-6-grudnia-2024-r)
- MF, [Krajowe Standardy Rachunkowości](https://www.gov.pl/web/finanse/krajowe-standardy-rachunkowosci) and the KSR PDFs linked in §5
- NBP mid-rates: [EUR 1 Oct 2025](https://api.nbp.pl/api/exchangerates/rates/a/eur/2025-10-01/?format=json) (4,2586), [EUR 31 Dec 2025](https://api.nbp.pl/api/exchangerates/rates/a/eur/2025-12-31/?format=json) (4,2267), [EUR 1 Oct 2024](https://api.nbp.pl/api/exchangerates/rates/a/eur/2024-10-01/?format=json) (4,2846), [EUR 31 Dec 2024](https://api.nbp.pl/api/exchangerates/rates/a/eur/2024-12-31/?format=json) (4,2730)

Note: ISAP pages may refuse automated clients (HTTP 403); every act above is also served by the Sejm ELI API as `https://api.sejm.gov.pl/eli/acts/DU/<year>/<position>/text.pdf`.

Professional (explanation only):
- Infor, [podniesienie o 25 % progów](https://ksiegowosc.infor.pl/rachunkowosc/sprawozdawczosc/6577736,zmiany-ustawy-o-rachunkowosci-podniesienie-o-25-progow-okreslajacyc.html); Prawo.pl, [30 Dec 2024](https://www.prawo.pl/podatki/bedzie-wiekszy-limit-dla-prowadzenia-ksiag-rachunkowych,526618.html)
- EY, [e-Sprawozdania finansowe 2025 — wybór właściwej struktury logicznej](https://www.ey.com/pl_pl/insights/tax/e-sprawozdania-finansowe-2025-wybor-wlasciwej-struktury-logicznej-dla-jednostki-innej); pit.pl, [nowe struktury logiczne, 12 Nov 2024](https://www.pit.pl/aktualnosci/ministerstwo-finansow-wprowadzamy-nowe-struktury-logiczne-dla-sprawozdan-finansowych-od-2025-r-1010431)
- fakturaxl.pl, [plan kont w pełnej księgowości](https://www.fakturaxl.pl/plan-kont-w-pelnej-ksiegowosci) (group names of practice only)
