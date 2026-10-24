# Filings fed from the books of a company keeping full books — and what each asks of the chart

Researched 2026-10-25 for Rhea's model Polish chart of accounts. Strand: tax, statistical and other
submissions fed from *księgi rachunkowe* (not the financial statements' positions, not model charts in
practice — two other notes cover those). Primary sources first: Dziennik Ustaw (via the Sejm/ELI API),
podatki.gov.pl and gov.pl/KAS (structures, XSD, brochures, forms), GUS (the 2026 programme and the 2026
forms), NBP (the reporting regulation). Where I read something myself rather than quote it, it says
"my reading". Anything not confirmed is marked *to verify*.

The JPK_KR_PD tag list in Appendix A was extracted from the official XSD
([JPK_KR_PD(1) schema 1-1, gov.pl](https://www.gov.pl/attachment/4ec8f106-032c-42f2-9fb5-15b5ec5ca670))
and cross-checked code by code against Annex 7 of the tag regulation
([Dz.U. 2024 poz. 1314, text](https://api.sejm.gov.pl/eli/acts/DU/2024/1314/text.html)): all 272 codes
(244 statement tags incl. `INNE`, plus 28 PD tags) match; the only differences were PDF-to-text artefacts
(`0` read for `O`).

---

## 0. The answer in one table

| Filing | Who files (company keeping full books) | When | Fed from the books | What it asks of the chart |
|---|---|---|---|---|
| **JPK_KR_PD** (+ **JPK_ST_KR**) | CIT taxpayers keeping books: phased in — PGK and revenue > €50m from tax years starting after 31.12.2024; JPK_V7M filers after 31.12.2025; everyone else (incl. JPK_V7K) after 31.12.2026 | Yearly, by the end of the **7th month** after the tax year (from 1.7.2026) | The whole ledger: journal, every posting, trial balance, counterparties, fixed-asset register, book-to-tax bridge | **Every account tagged** with a statement tag (`S_12_1`, mandatory) from the MF dictionary, a **PD tax tag** (`S_12_3`) where the account serves the tax computation; homogeneous analytic accounts; parent-account pointer; NIP and KSeF number on entries |
| **CIT-8(34)** + attachments | Every CIT taxpayer on general rules | Yearly, end of the **3rd month**; advances monthly (or quarterly) by the 20th | Revenue and tax-deductible costs, split **capital gains / other sources**, Poland / abroad; KUP/NKUP; tax depreciation; losses; bad-debt relief; debt-financing costs | Revenue/cost accounts that separate capital-gains sources; KUP vs NKUP (permanent vs timing) accounts; R&D and IP-box separation; financing-cost accounts; investment fund in reserve capital |
| **CIT-8E(4)** (Estonian CIT) | Companies on *ryczałt od dochodów spółek* | Yearly, end of 3rd month; hidden profits / non-business expenses monthly by the 20th | Distributed profit, hidden profits, non-business expenses, undisclosed transactions | Equity split: profits before vs during ryczałt (dedicated JPK tags `…_ES_…`); accounts that isolate hidden-profit and non-business items |
| **JPK_V7M / JPK_V7K** | VAT taxpayers | Monthly (or quarterly declaration), by the 25th | VAT **registers**, not the chart | Only VAT settlement accounts and the reconciliation of registers to them |
| **GUS F-01/I-01** | Entities keeping books with **≥ 50** persons working (quarterly) or **10–49** (half-yearly), most PKD sections | 20 Apr / 20 Jul / 20 Oct 2026, 10 Feb 2027 | P&L **comparative variant** (costs by nature) YTD, selected balance items, capex, headcount | Costs **by nature** always available; extra splits: export, energy, subcontracting, excise, wages from employment, social insurance, training, travel, insurance premiums, released provisions, interest types, FX net, subsidies, imports, capex by KŚT group |
| **GUS SP** (annual) / **SP-3** (≤9, sample) | ≥ 10 persons working (SP); ≤ 9 sampled (SP-3) | SP by 4 May; SP-3 by 17 March (for 2026) | Full balance + P&L comparative, fixed assets by KŚT group with movements, capex financing, local units | As F-01 plus overdue receivables/payables, foreign loans, owner types of share capital, operating-lease and agency-work services, VAT payable/refundable |
| **GUS F-02** | — | — | — | **Not in the 2023–2026 statistical programmes** (only F-02/dk for cultural institutions); its content is in SP |
| **NBP** (balance of payments) | Residents whose foreign assets+liabilities on the NBP forms exceed **PLN 10m** (quarterly) / **PLN 300m** (monthly); **> PLN 3m** of trade items with foreign trade; FDI links ≥ 10% (annual) | 26 days after quarter / 20 after month; FDI by 31 May | Receivables, payables, loans, deposits, equity vs **non-residents** | Non-resident split of receivables, payables, loans, cash at foreign banks; equity and result for FDI forms |
| **TPR-C(6)** | Related entities with documentable controlled transactions (thresholds PLN 10m goods/financial, PLN 2m services/other) | End of **11th month** | Financial data and controlled transactions | Related-party analytics on revenues, costs, receivables, payables, loans |
| **Financial statement** to KRS (or Szef KAS) | Every company | 15 days after approval (approval ≤ 6 months) | The statements | Covered by the statements note |

The details, sources and the chart consequences follow.

---

## 1. JPK_KR_PD — the books themselves, with account tags

### 1.1 What it is and its legal basis

- Art. 9 ust. 1c of the CIT Act: taxpayers keeping *księgi rachunkowe* must keep them with software and
  send them to the head of the tax office after the tax year, electronically, in the logical structure
  under art. 193a § 2 Ordynacja podatkowa
  ([CIT Act, consolidated text Dz.U. 2026 poz. 554](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)).
  Art. 9 ust. 1e applies the same to *spółki niebędące osobami prawnymi* whose partners are not only
  natural persons; ust. 1g to each company in a PGK (same source).
- Art. 9 ust. 5 lets the Minister set the additional data and the exempted groups (same source). Done by:
  - **Rozporządzenie Ministra Finansów z dnia 16 sierpnia 2024 r. w sprawie dodatkowych danych, o które
    należy uzupełnić prowadzone księgi rachunkowe podlegające przekazaniu na podstawie ustawy o podatku
    dochodowym od osób prawnych** ([Dz.U. 2024 poz. 1314](https://api.sejm.gov.pl/eli/acts/DU/2024/1314/text.html)),
    in force 1.1.2025. § 2 ust. 1 lists the additional data: (1) the counterparty's NIP if it has one;
    (2) the KSeF number of invoices issued by the taxpayer; (3) **account tags (*znaczniki identyfikujące
    konta ksiąg*)** from one of seven dictionaries (Annexes 1–7); (4) fixed-asset/intangible data
    (acquisition, acceptance into use, removal, inventory number); (5) the **difference between the
    accounting result and the tax base**, split into eight items (same source).
  - § 2 ust. 3: tags are assigned "w oparciu o kryteria klasyfikacyjne przyjęte dla zapisów zdarzeń
    w księgach, wynikające z przepisów o rachunkowości" (same source).
  - § 5 ust. 1: for years starting in 2025 the books may omit items (1), (2), (4) and (5) — **not the
    tags (3)** (same source). § 5 ust. 2: IFRS reporters may omit the tags; the amending regulation
    **Dz.U. 2025 poz. 1828** moves that IFRS relief to tax years starting before **1.1.2028** and reworks
    the fixed-asset items (adds "przyczyna wykreślenia z ewidencji")
    ([Dz.U. 2025 poz. 1828](https://eli.gov.pl/eli/DU/2025/1828/ogl/pol/pdf)).
  - **Rozporządzenie MF z 13 grudnia 2024 r. w sprawie zwolnienia z obowiązku przesyłania części ksiąg
    rachunkowych…** exempts the fixed-asset register (JPK_ST_KR) for tax years starting in 2025
    ([Dz.U. 2024 poz. 1861](https://api.sejm.gov.pl/eli/acts/DU/2024/1861/text.pdf)).
  - PIT taxpayers keeping books have a parallel regulation of 6.09.2025 (Dz.U. 2025 poz. 1311), cited by
    the MF FAQ Q126 ([podatki.gov.pl FAQ](https://www.podatki.gov.pl/podatki-firmowe/jednolity-plik-kontrolny/jpk_pd/pytania-i-odpowiedzi-jpk_pd));
    *title not verified*.
- Current structure: **JPK_KR_PD (1), schema version 1-1**; the change from 1-0 to 1-1 only aligned the
  dictionaries with Dz.U. 2024 poz. 1314 ([KAS: Struktury JPK w podatkach dochodowych](https://www.gov.pl/web/kas/struktury-jpk-w-podatkach-dochodowych);
  [brochure update valid from 1.7.2026](https://www.gov.pl/attachment/b16d282e-e1df-47df-ba7f-cfa973f4412e)).
  XSD namespace `http://jpk.mf.gov.pl/wzor/2024/09/04/09041/` (from the XSD itself).

### 1.2 Who must send it, from which year

From the MF brochure update valid from 1.7.2026 ([gov.pl attachment](https://www.gov.pl/attachment/b16d282e-e1df-47df-ba7f-cfa973f4412e))
and FAQ Q10 (updated 2025-10-28) ([FAQ](https://www.podatki.gov.pl/podatki-firmowe/jednolity-plik-kontrolny/jpk_pd/pytania-i-odpowiedzi-jpk_pd)):
first for the tax year (for non-corporate partnerships: financial year) **starting after**:

| Starting after | CIT taxpayers and *spółki niebędące osobami prawnymi* | Structures |
|---|---|---|
| 31.12.2024 | (a) tax capital groups (PGK); (b) taxpayers whose prior-year revenue exceeded **€50m** (tax revenue, not accounting — FAQ Q45; base year = the year before the first obliged year — FAQ Q46) | JPK_KR_PD; JPK_ST_KR optional for 2025 |
| 31.12.2025 | taxpayers obliged to file **JPK_V7M** | JPK_KR_PD + JPK_ST_KR |
| 31.12.2026 | **all others**, including **JPK_V7K** filers | JPK_KR_PD + JPK_ST_KR |

- The obligation is assessed at the date the books are opened (FAQ Q134, 2025-11-12, same FAQ).
- Each company of a PGK files its own JPK_KR_PD and JPK_ST_KR (FAQ Q41).
- **Exempt** (art. 9 ust. 1d CIT): taxpayers subjectively exempt under art. 6 ust. 1 (except family
  foundations); taxpayers allowed to file a paper return (art. 27a); those keeping simplified records
  ([CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf); FAQ Q127). Budget units and local
  governments are outside (FAQ Q51).
- **Estonian CIT taxpayers file too** (art. 9 ust. 1c refers to the art. 28r declaration); they use the
  statement tags but **no PD tags**, and put **0** in the RPD node (FAQ Q114).
- **IFRS reporters file too**; tags optional for them until tax years starting before 1.1.2028 (FAQ Q149,
  2026-06-24; Dz.U. 2025 poz. 1828).
- For **2025** all taxpayers were relieved from NIP, KSeF number and RPD values; for books of years
  starting after 31.12.2025 the RPD node is required (FAQ Q149; Dz.U. 2024 poz. 1314 § 5).

For Rhea's typical client: an sp. z o.o. filing JPK_V7M sends its first JPK_KR_PD for **2026**; one filing
JPK_V7K (or not VAT-registered) for **2027** (my reading of the table above).

### 1.3 When and how

- **Deadline**: from 1.7.2026 art. 9 ust. 1c reads "do końca **siódmego** miesiąca po zakończeniu roku
  podatkowego" (and ust. 1e "siódmego" instead of "trzeciego")
  ([ustawa z 15 maja 2026 r., Dz.U. 2026 poz. 779](https://api.sejm.gov.pl/eli/acts/DU/2026/779/text.pdf)).
  Before that a regulation extended the deadline to the end of the 7th month for years starting after
  31.12.2024 and ending before 1.4.2026, and to the end of July 2026 for phase-1 years ending before
  31.12.2025 ([Dz.U. 2026 poz. 188](https://api.sejm.gov.pl/eli/acts/DU/2026/188/text.pdf)). So a calendar
  2026 year is due **31 July 2027** — four months *after* the CIT-8 (end of March).
- The UPL-1 power of attorney for e-declarations also covers signing JPK_KR_PD (new art. 9 ust. 1h, in
  force 16.6.2026, same act 2026/779).
- Sent through the MF API or the *Klient JPK WEB* app (FAQ Q8); limits 200 GB per file by API, 100 MB by
  Klient JPK_WEB (FAQ Q118/Q122, spec 5.1.0); current interface spec 5.6.0
  ([podatki.gov.pl, Pliki do pobrania JPK_PD](https://www.podatki.gov.pl/podatki-firmowe/jednolity-plik-kontrolny/jpk_pd/pliki-do-pobrania-jpk_pd)).
- The file may be split into period files, continuous with no gaps or overlaps, every day of the year
  covered; RPD filled only in the last period's file, "0" in the others (FAQ Q6, Q39, Q150).
- A correction resends all data of the file it corrects (FAQ Q139).
- JPK_ST_KR is sent with JPK_KR_PD, cannot be split (FAQ Q4, Q12).

### 1.4 The structure — nodes and fields

Read from the XSD ([JPK_KR_PD(1) v1-1](https://www.gov.pl/attachment/4ec8f106-032c-42f2-9fb5-15b5ec5ca670))
and the brochure ([Broszura JPK_KR_PD(1)](https://www.gov.pl/attachment/22010dbd-5e27-433b-84c5-ec99cee7487c)).

| Node | Fields | What the chart/ledger must hold |
|---|---|---|
| `Naglowek` | form code, purpose (1 first / 2 correction), file period `DataOd`/`DataDo`, financial year `RokDataOd/Do`, **tax year** `RokPdDataOd/Do` (if different), default currency, tax office code | Financial year and tax year as separate attributes of the company |
| `Podmiot1` | identifiers, address, `Znacznik_EST` (Estonian CIT), `Znacznik_MSSF` (IFRS) | Company flags: Estonian CIT, IFRS |
| `Kontrahent` (0..n) | `T_1` counterparty code in the system, `T_2` country of tax ID, `T_3` tax ID | A counterparty master with stable internal codes, country and NIP; only counterparties with a NIP that had transactions in the period (FAQ Q146); employees only for their business transactions (FAQ Q123) |
| `ZOiS` (trial balance), one of `ZOiS1`…`ZOiS8` by entity type | `S_1` account id (final posting account), `S_2` name, `S_3` **parent account id**, `S_4/S_5` opening balance Dr/Cr, `S_6/S_7` period turnover, `S_8/S_9` YTD turnover, `S_10/S_11` closing balance, **`S_12_1` tag (mandatory), `S_12_2` second tag (optional), `S_12_3` PD tag** | Hierarchical accounts (synthetic → analytic, the parent "not always synthetic" — FAQ Q5/Q13/Q36; for a synthetic account `S_1` = `S_3`); every account with balance or turnover appears, synthetics included (FAQ Q141, Q44/Q47); tag attributes on every account |
| `Dziennik` (journal, 1..n) | `D_1` entry no. (continuous in the year), `D_2` journal (partial journals, e.g. "Zakup"), `D_3` counterparty code (= `T_1`), `D_4` document number, `D_5` document type, `D_6` operation date, `D_7` document date, `D_8` posting date, `D_9` person responsible, `D_10` description, `D_11` amount, `D_12` **KSeF number** | Per journal entry: counterparty, document no. and type, three dates, author, KSeF number for KSeF invoices (required from 2026 where the document has one — FAQ Q112) |
| `KontoZapis` (postings within each journal entry) | `Z_1` line no., `Z_2` line description, `Z_3` account (= `S_1`), `Z_4` Dr PLN, `Z_5` Dr foreign amount, `Z_6` Dr currency, `Z_7` Cr PLN, `Z_8` Cr foreign amount, `Z_9` Cr currency | Postings only on final accounts; foreign-currency amount and ISO currency on each line |
| `Ctrl` | `C_1`–`C_5` counts and sums | — |
| `RPD` *Rozliczenie podatku dochodowego na zasadach ogólnych* | `K_1` exempt revenue (permanent), `K_2` revenue not taxable this year, `K_3` revenue taxable this year booked in prior years, `K_4` non-deductible costs (permanent), `K_5` costs not deductible this year, `K_6` costs deductible this year booked in prior years, `K_7` taxable revenue not booked, `K_8` deductible costs not booked | The book-to-tax bridge, entered manually, must reconcile with the PD-tagged result and off-balance accounts (FAQ Q22/Q38); not split capital gains / other (FAQ Q40) |

### 1.5 The tags (*znaczniki kont*) — rules

From the FAQ ([podatki.gov.pl](https://www.podatki.gov.pl/podatki-firmowe/jednolity-plik-kontrolny/jpk_pd/pytania-i-odpowiedzi-jpk_pd)),
question numbers as published:

- **One tag per account in `S_12_1`, mandatory**; `S_12_2` optional; two tags only in exceptional cases,
  and then only balance-sheet / P&L names (Q23, Q61, Q138).
- **Lowest level of the dictionary**: e.g. `BABI1`, `BABI1_O`, `BABI1_A` are distinct tags for materials,
  their price variances and their write-downs (Q48).
- **Homogeneous accounts**: where one account mixes events that need different tags, "podzielić
  zdarzenia na jednorodne grupy, zaewidencjonować każdą grupę na wyodrębnionym koncie analitycznym" and
  tag each (Q128, Q138 — 2026-04-29). The FAQ notes this requirement comes from tax law, not the
  Accounting Act (Q138).
- **Both cost teams are tagged in every variant**: accounts kept for the comparative variant (team 4) and
  for the calculation variant (team 5) — and for a third layout if kept — are all mapped (Q108, Q128).
- **Synthetic accounts** repeat their analytics' tag if it fits both, otherwise `INNE` (Q147, 2026-06-24).
- **`INNE`** only when the content fits no tag, e.g. account 490 *Rozliczenie kosztów* (Q61, Q7).
- **PD tags in `S_12_3`** only on accounts used to settle income tax (Q17, Q23). Book depreciation and
  tax-deductible depreciation accounts are **not** PD-tagged; only write-offs that are **not** KUP get a PD
  tag (exception: R&D relief) (Q32). Each analytic account may carry one PD tag only (Q50, Q53).
- **Off-balance accounts** used for book/tax differences: `INNE` in `S_12_1` and a `…_PB` (POZABILANSOWE)
  PD tag in `S_12_3`; their entries go in `KontoZapis`; other off-balance accounts optional (Q87, Q143).
  Off-balance accounts for bad-debt relief, hypothetical interest and excess one-off tax depreciation
  have **no** PD tag → `INNE` (Q130–Q132).
- Benefits not booked: no PD tag; taxable revenue goes to RPD `K_7`, deductible costs to `K_8` (Q96,
  corrected 2025-12-18).
- Sale of materials: `RPDIV_RKGIV_P` / `RPEIII_RKHIII_P` where the entity keeps separate accounts for
  the ZCP / goodwill items, otherwise `RPDIV_RKGIV` / `RPEIII_RKHIII` (Q105, Q133).
- Estonian CIT: statement tags yes, PD tags no, RPD = 0 (Q114).

### 1.6 The dictionaries

Seven dictionaries, Annexes 1–7 of Dz.U. 2024 poz. 1314, each a `ZOiS` variant in the XSD; value counts are
the enumerations in the XSD:

| Annex | Entities | XSD node / type | Values |
|---|---|---|---|
| 1 | banks | `ZOiS1` / `TMapKontaBanki` | 131 |
| 2 | insurers and reinsurers | `ZOiS2` / `TMapKontaUbezp` | 205 |
| 3 | public-benefit organisations reporting under Annex 6 UoR | `ZOiS3` / `TMapKontaPP` | 109 |
| 4 | investment funds | `ZOiS4` / `TMapKontaFI` | 39 |
| 5 | brokerage houses | `ZOiS5` / `TMapKontaDM` | 209 |
| 6 | SKOK | `ZOiS6` / `TMapKontaSKOK` | 85 |
| 7 | **pozostałe jednostki** — every ordinary company | `ZOiS7` / `TMapKontaPOZ` | **244** (135 assets, 51 equity & liabilities, 57 P&L, `INNE`) |
| — | IFRS reporters | `ZOiS8` (free text, optional until 2028) | — |
| all | tax tags PD (`S_12_3`) | `TMapKontaPD` | **28** (13 PD + 15 POZABILANSOWE) |

- **Micro and small entities get no simpler dictionary**: entities reporting under Annexes 1, **4 and 5**
  UoR use Annex 7 and `ZOiS7` (FAQ Q94). The accountants' association SKwP asked MF for a simplified
  dictionary for micro/small entities, noting that Annex-7 tags force analytics that the simplifications
  were meant to avoid, and for explicit Estonian-CIT tax tags (hidden profits, non-business expenses)
  ([SKwP opinion of 15.10.2025](https://skwp.pl//content/uploads/2026/03/odpowiedz-dla-MFiG_61.pdf)); the
  enacted amendment (Dz.U. 2025 poz. 1828) did not add either.
- A separate IFRS dictionary is in consultation (FAQ Q115) — *status to verify*.

**How the Annex-7 codes are built** (my reading of the dictionary descriptions, not an official key):
`B`=bilans, `BA`/`BP` = assets / equity & liabilities, then the Annex-1 balance-sheet letters and numerals
(`BAAII1c` = A.II.1.c *urządzenia techniczne i maszyny*); `R`=wynik with `RP…` = comparative-variant
position and `RK…` = calculation-variant position (`RPAI_RKAI_POW` is the same line in both variants).
Suffixes: `_W` gross value, `_U` accumulated depreciation (*umorzenie*), `_A` write-downs (*odpisy
aktualizujące*), `_O` price variances (*odchylenia od cen ewidencyjnych*), `_D12`/`_P12` maturity up to /
over 12 months, `_D`/`_K` long / short term, `ULA` shares, `IPW` other securities, `UP` loans granted,
`IDAF`/`IKAF` other long/short financial assets, `SPKR` cash in hand and at bank, `POW`/`POZ` related /
other parties, `_RK` FX differences, `_ES_…` Estonian-CIT splits, `ZCP` organised part of an enterprise,
`WF` goodwill items, `AKC` excise, `E` pension contributions.

The full list is in **Appendix A**.

### 1.7 What JPK_KR_PD requires of the chart (consolidated)

Where the Annex-7 dictionary is **finer than the Annex-1 statement**, the chart needs separate accounts
(read from the dictionary in Appendix A):

1. **Fixed assets and intangibles**: gross value, accumulated depreciation and write-downs on separate
   accounts for each line — development costs, goodwill, other intangibles, advances; land, buildings,
   machinery, vehicles, other fixed assets; construction in progress and its advances; long-term
   investments in property and in intangibles (`_W`, `_U`, `_A`).
2. **Inventories**: each of materials, WIP, finished goods, goods: base, price variances (`_O`),
   write-downs (`_A`); advances for deliveries.
3. **Receivables**: trade receivables split **related / capital-engagement / other** and **≤ 12 / > 12
   months**, each with its write-down account; tax and social receivables (`BABII3b` — VAT input lands
   here, my reading); other; **receivables in court** (`BABII3d`) with write-downs.
4. **Financial assets**: shares, other securities, loans granted, other financial assets — each split by
   related / capital-engagement / other, long and short term, each with write-downs; cash split into cash
   in hand and at bank, other cash, other monetary assets.
5. **Equity**: share capital; supplementary capital split into **share premium** (`BPAIIWS`) and other;
   revaluation reserve split **fair value** / other; other reserves split **statutory**, **for own
   shares**, other; **Estonian-CIT splits** of supplementary capital, reserves and retained earnings (profit
   from years before vs during ryczałt; undistributed profits / uncovered losses before vs during); net
   profit; interim profit distributions.
6. **Provisions**: deferred tax; pension provisions long / short; other provisions long / short.
7. **Liabilities**: long-term by related / capital-engagement / other, and within "other": loans and
   borrowings, debt securities, other financial liabilities (e.g. finance leases — my reading), bills of
   exchange, other; short-term trade payables by group and **≤ 12 / > 12 months**; advances received;
   bills; taxes and social security (`BPBIII3g` — VAT output lands here, my reading); wages; other;
   special funds.
8. **Accruals**: negative goodwill; other accruals long / short.
9. **P&L**: revenue from products and from goods/materials split **related / other**; costs by nature,
   with **excise separate from other taxes** and **pension contributions** within social security
   (`RPBVI_E`); value of goods and materials sold split related / other; **calculation variant**: cost of
   products sold related / other, selling costs, general and administrative costs; other operating income
   and costs: disposals of non-financial fixed assets (related / other), subsidies, revaluations, and in
   "other" the **ZCP disposal**, **negative-goodwill / goodwill write-off** and the rest; finance income and
   costs: dividends from related and from other entities, each split into entities where the company
   has a capital engagement vs the rest (`RPGIa…_k/_p`, `RPGIb…_k/_p`), **interest** related / other,
   disposals of financial assets related / other, revaluation, **FX gains and FX losses as separate
   accounts** (`RPGV_RKJV_RK`, `RPHIV_RKKIV_RK` — gross, while the statement shows the net excess; my
   reading), other;
   income tax **current / deferred**; other mandatory reductions of profit.
10. **Tax (PD) tags**: separate accounts (or off-balance accounts) for revenue exempt / not taxable / not
    taxable this year, for **NKUP permanent** (split by the four PD4 sub-tags) and **NKUP timing** (PD5),
    for **R&D-relief costs** (PD7) and **IP-box** revenue and costs (PD8_1, PD8_2). One PD tag per account
    means a cost account cannot hold both deductible and non-deductible items (FAQ Q50/Q53 + Q128).
11. **Ledger mechanics**: parent pointers, account names, continuous journal numbering per year, partial
    journals named, counterparty codes with NIP and country, KSeF numbers, document type, three dates,
    person responsible, foreign-currency amount and currency on each posting line.

### 1.8 JPK_ST_KR — the fixed-asset register sent with it

From the XSD ([JPK_ST_KR(1) v1-0](https://www.gov.pl/attachment/336fb09b-ea93-475d-a895-0ef5b8b72a0b)) and the FAQ:
per asset/intangible — inventory number (`E_1`), acquisition / production date (`E_2`/`E_3`), acceptance
into use (`E_4`), OT document (`E_5`), acquisition type (`E_6`: S sale contract, D donation, N in-kind
contribution, W own production, F VAT invoice, I other), name, **KŚT symbol** (`E_8`), **tax depreciation
method** up to three (`E_9_1..3`: D degressive, L straight-line, J one-off, I other, X none), tax rates
(`E_10A`, `E_10B`, `E_11`), **tax initial value** (`E_12`) and its revaluations/changes (`E_13`–`E_15`),
frequency (`E_17`: M, K, R, J, S, I, X), **tax depreciation that is not KUP** (`E_18`), tax depreciation
for the year and cumulative (`E_19`, `E_20`), **book initial value** (`E_21`) and changes (`E_22`–`E_24`),
book rates (`E_25A/B`), book depreciation for the year and cumulative (`E_26`, `E_27`), removal date,
reason (`E_29`: A reclassification, S sale, X shortage, K theft, D donation, N in-kind contribution, I
other) and document, KSeF number of the sale invoice (`E_31`), KŚT reclassification flag (`E_32`).
Assets accepted before 1.1.2025 may show "brak obowiązku" / 1900-01-01 for the old data (FAQ Q49;
Dz.U. 2025 poz. 1828 § 4). Estonian-CIT companies fill it with "X – brak amortyzacji" etc. (FAQ Q137).

Chart consequence: the register is a subledger, not accounts, but it must **carry tax and book values side
by side** and reconcile to the `_W`/`_U`/`_A` accounts of 1.7 (my reading).

### 1.9 Relation to JPK_KR "on demand"

JPK_KR_PD does **not** abolish **JPK_KR** on demand (art. 193a Ordynacja podatkowa); JPK_KR still applies
e.g. to periods before JPK_KR_PD (FAQ Q16). JPK_KR has no parent-account field; it carries
team/category/sub-category codes instead (FAQ, cited in search snippet of the same page — *to verify the
exact JPK_KR field names*). Other on-demand structures (bank statements, warehouse, invoices) are listed
under "JPK na żądanie" on podatki.gov.pl — *not examined here*.

---

## 2. CIT — the return, its attachments, advances, Estonian CIT

### 2.1 CIT-8 and its timing

- **CIT-8(34)** — *Zeznanie o wysokości osiągniętego dochodu (poniesionej straty) i należnego podatku
  dochodowego od osób prawnych*; filed electronically by the **end of the 3rd month** after the tax year
  (art. 27 ust. 1, 1c CIT) ([CIT-8(34) form](https://www.gov.pl/attachment/b6901dbb-05bf-4d5b-b1be-52fd86fc753d);
  [CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)). A draft **CIT-8(35)** (bank-sector
  changes) was in consultation until 7.7.2026 ([gov.pl](https://www.gov.pl/web/finanse/konsultacje-podatkowe-wzoru-formularza-cit-8));
  the year it applies from is *to verify*.
- **Advances**: monthly by the 20th of the next month (art. 25 ust. 1, 1a); **quarterly** for small
  taxpayers (sales revenue incl. VAT ≤ €2m in the prior year, art. 4a pkt 10) and start-ups in year one
  (art. 25 ust. 1b, 1c); **simplified** 1/12 of the tax shown in the return filed in the prior year (or the
  year before) (art. 25 ust. 6) (all [CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)).
  There is no monthly CIT declaration: the advances due and paid per month/quarter are reported in
  CIT-8 part G (poz. 151–246) ([CIT-8(34)](https://www.gov.pl/attachment/b6901dbb-05bf-4d5b-b1be-52fd86fc753d)).

### 2.2 The CIT-8(34) lines that come from the books

From the form ([CIT-8(34)](https://www.gov.pl/attachment/b6901dbb-05bf-4d5b-b1be-52fd86fc753d)):

| Part | Content | Chart need |
|---|---|---|
| B.3 poz. 19–36 | method for revenue date (art. 12 ust. 3g); **debt-financing costs excluded** under art. 15c (poz. 23) and art. 16 ust. 1 pkt 13f (poz. 24); FX differences by the accounting method (art. 9b); assets moved abroad; **hypothetical interest** (art. 15cb, poz. 35–36) | Financing-cost accounts complete enough for the art. 15c cap (PLN 3m or 30% of tax EBITDA; "koszty finansowania dłużnego" include interest, fees, commissions, the interest part of lease instalments, late-payment penalties, hedging — art. 15c ust. 1, 12); interest-type revenue identifiable (art. 15c ust. 13) |
| D.1 poz. 52–61 | **revenue** in two columns: *z zysków kapitałowych* / *z innych źródeł*; rows: Polish sources, foreign dividends and art. 21-type income, other foreign sources, IP contributions | Revenue accounts (and finance income) that separate **capital-gains** revenue (dividends, disposals of shares, etc.) from operating revenue, and **foreign-source** revenue |
| D.2 poz. 62–77 | **tax-deductible costs** in the same two columns; deferred interest / financing costs from prior years | KUP vs NKUP separation at account level (also needed for PD tags) |
| D.3–D.4 | income/loss per source; IP-box and SEZ adjustments | — |
| E.1 poz. 96–113 | exempt income, **prior-year losses** (by source), other deductions, investment deductions (from CIT-8/O) | Loss records by source (outside the chart) |
| E.3 poz. 116–119 | deductions from the base: new technology, **R&D relief** (from CIT/BR), others | R&D cost accounts by CIT/BR category (2.3) |
| E.5 poz. 121–130 | **bad-debt relief** (art. 18f): unpaid receivables / unpaid liabilities 90 days past due date | Due dates and settlement status per receivable/payable (subledger) |
| F | 19% / 9% rate, tax due, deductions | — |
| G | advances due and paid per month/quarter; tax on revenue from buildings (art. 24b) | CIT advance analytics per period |
| H–I | settlement, plus CIT/BR, CIT/WW, CIT/8S, CIT/IP, CIT/KW, CIT/PD, **CIT/M** amounts | — |
| L poz. 316–317 | entering / leaving **Estonian CIT** next / previous year | — |
| L poz. 318–320 | write-offs to an **investment fund separated within reserve capital** (art. 15 ust. 1hb) | A dedicated reserve-capital sub-account **and** a dedicated bank account: the fund must be funded "na wyodrębniony wyłącznie w tym celu" settlement account (art. 15 ust. 1hb pkt 2, [CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)) |

### 2.3 Attachments that exist now

The MF list of CIT forms ([podatki.gov.pl — formularze CIT](https://www.podatki.gov.pl/cit/e-deklaracje-cit/formularze-cit))
shows, among others: **CIT-8/O(20)** (deductions and exempt income), **CIT/BR(10)** (R&D relief),
**CIT/IP(5)** (IP box), **CIT/M(2)** (minimum tax), **CIT/8S(7)** (SEZ permit income), **CIT/8SP(7)** (income
under a support decision), **CIT/KW(4)** (transformation income on entering ryczałt), **CIT/PD(2)**
(shifted income), **CIT/WW(3)** (virtual currencies), **CIT/MIT(3)** (tax on revenue from buildings),
**CIT/PM(1)**, **CIT/NZI(2)**, **CIT/PGK(7)**, **CIT-D(8)** (donations), **CIT-CSR(2)**, **CIT-8ST(1)**
(branches), **CIT-8E(4)** (Estonian CIT), **CIT-8AB(9)** (PGK), **TPR-C(6)**, **IFT-2R(12)**, **CIT-10Z(7)**,
**ZAW-RD(2)** (notice of choosing ryczałt). The CIT-8(34) attachment box also lists CIT/F and CIT/WOT
([CIT-8(34)](https://www.gov.pl/attachment/b6901dbb-05bf-4d5b-b1be-52fd86fc753d)). There is **no form
named "CIT-ST"** in the list; the branch information is **CIT-8ST** (my reading of the list).

Those fed from the books, and what they ask:

- **CIT/BR(10)**: qualified R&D costs in nine categories — salaries and contributions (art. 18d ust. 2
  pkt 1, 1a), materials and raw materials (pkt 2), specialist equipment not fixed assets (pkt 2a),
  expertise/advisory (pkt 3), paid use of research equipment (pkt 4), service of using research equipment
  (pkt 4a), patents and protection rights (pkt 5), depreciation of development costs (ust. 2a), other
  depreciation (ust. 3) ([CIT/BR(10)](https://www.podatki.gov.pl/media/cn1nighs/cit_br_1007.pdf)). The Act
  requires R&D costs to be **separated in the records** (art. 9 ust. 1b, [CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)).
  Chart: an R&D cost dimension (or sub-accounts) by these categories, with the PD7 tag.
- **CIT/IP(5)**: art. 24e requires **each qualified IP right separated** in the records, with revenue,
  costs and income per right ([CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)). Chart:
  an IP-right dimension; PD8_1/PD8_2 tags.
- **CIT/M(2)**: minimum tax (10% of a base) for companies with a **loss** from operating sources or income
  ≤ **2%** of operating revenue (art. 24ca ust. 1); the base includes 1.5% of operating revenue and excess
  **debt-financing costs to related parties** ([CIT/M(2)](https://www.podatki.gov.pl/media/0nlhw5go/cit-m0206.pdf);
  [CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)). Chart: same splits as CIT-8 plus
  related-party financing costs.
- **CIT-8/O(20)**: exempt income, losses by source, other deductions, investment deductions, dividend data,
  payments to foreign legal persons, holding-company dividends ([CIT-8/O(20)](https://www.podatki.gov.pl/media/czfdxieo/cit-8o-20-07.pdf)).
- **CIT/8S, CIT/8SP**: income from SEZ / support-decision activity — the activity must be separable in
  revenue and costs (my reading; *details not examined*).
- **CIT-8ST(1)** (by 31 March): headcount under employment contracts per branch and municipality —
  payroll data, **not** from the chart ([CIT-8ST(1)](https://www.podatki.gov.pl/media/fl5pwqca/cit-8st_-1-12.pdf);
  art. 28 CIT).
- **Withholding tax**: WHT paid by the 7th of the next month; **IFT-2R** to the non-resident and the tax
  office by the end of the **3rd month** of the next year (art. 26 ust. 3, 3a); annual declaration
  **CIT-10Z** by the end of the **1st month** (art. 26a ust. 1) ([CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)).
  Chart: payments to non-residents identifiable by type (interest, royalties, intangible services,
  dividends) and by recipient country.

### 2.4 Estonian CIT (*ryczałt od dochodów spółek*) and CIT-8E

- Eligible: sp. z o.o., S.A., P.S.A., sp.k., S.K.A. owned only by natural persons; < 50% of prior-year
  revenue (incl. VAT) from listed **passive sources** (receivables, interest, lease interest, guarantees,
  IP, financial instruments, related-party transactions with no added value); at least 3 FTE employees
  who are not partners (or equivalent spend on other contracts); no shares in other companies; **no IFRS**
  for the ryczałt period; notice ZAW-RD by the end of the first month (art. 28j ust. 1)
  ([CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)). Chart: revenue analytics that can
  measure the passive-revenue share.
- Taxed income (art. 28m ust. 1): **distributed profit** and profit used to cover pre-ryczałt losses;
  **hidden profits** (*ukryte zyski* — e.g. loans to shareholders, donations, **representation
  expenses**, excess over market value in related transactions, art. 28m ust. 3); **non-business expenses**
  (*wydatki niezwiązane z działalnością gospodarczą*); change in asset value on mergers; **net profit** on
  leaving ryczałt; **undisclosed transactions** (*nieujawnione operacje gospodarcze*). For cars and
  similar assets not used only for business, 50% of the expenses and depreciation is excluded from hidden
  profits and non-business expenses (fully excluded if used only for business; the taxpayer proves it)
  (art. 28m ust. 4 pkt 2, ust. 4a, ust. 5) (same source).
- Payment (art. 28t): distributed profit — by the end of the 3rd month of the year after the profit
  resolution; hidden profits and non-business expenses — **by the 20th of the month after** the payment;
  undisclosed transactions — end of the 3rd month after the year (same source).
- **CIT-8E(4)** — *Deklaracja o wysokości osiągniętego dochodu i należnego ryczałtu od dochodów spółek*;
  by the end of the 3rd month of the next tax year (art. 28r ust. 1); electronic only; sections D.1
  distributed profit, D.2 hidden profits and non-business expenses **per month**, D.3 change in asset
  value, D.4 undisclosed transactions, D.5 net profit on leaving (per year of ryczałt)
  ([CIT-8E(4)](https://www.podatki.gov.pl/media/dhjhcy34/cit-8e_-4-9.pdf)).
- JPK_KR_PD: `Znacznik_EST`; statement tags yes; PD tags no; RPD = 0 (FAQ Q114); equity tags with
  `_ES_ZP` / `_ES_ZR` / `_ES_NZP` / `_ES_NPR` / `_ES_NSP` / `_ES_NSR` (Appendix A).
- MF guidance: *Przewodnik do ryczałtu od dochodów spółek* (tax explanations, 23.12.2021)
  ([gov.pl](https://www.gov.pl/web/finanse/ryczalt-od-dochodow-spolek--objasnienia-podatkowe)).
- **Chart needs**: (a) retained earnings and reserves split by the year the profit was earned — before vs
  during ryczałt — because the tax base depends on it and the JPK tags require it; (b) accounts (or a
  flag) for hidden-profit items (representation, donations, shareholder loans and their interest,
  related-party price excess) and non-business expenses, posted monthly to drive the monthly payment;
  (c) a 50% split for mixed-use cars and similar assets (my reading of the above).

---

## 3. VAT — JPK_V7M / JPK_V7K are fed from the registers, not the chart

JPK_V7M (monthly) and JPK_V7K (quarterly declaration, monthly records) consist of a **records part**
(*ewidencja* — the VAT purchase and sales registers) and a **declaration part** (VAT-7/VAT-7K positions),
due by the **25th** of the following month; legal basis the VAT Act and the regulation of 15.10.2019
(Dz.U. 2019 poz. 1988) ([podatki.gov.pl — JPK_VAT z deklaracją](https://www.podatki.gov.pl/podatki-firmowe/jednolity-plik-kontrolny/jpk_vat-z-deklaracja/jpk_vat-z-deklaracja)).
Current structures **JPK_V7M(3) / JPK_V7K(3)** apply from **1.2.2026** (version 2 from 1.1.2022 to
31.1.2026) ([Pliki do pobrania](https://www.podatki.gov.pl/podatki-firmowe/jednolity-plik-kontrolny/jpk_vat-z-deklaracja/pliki-do-pobrania));
version 3 adds the KSeF number or the markers OFF / BFK / DI per document (secondary source:
[Infor](https://ksiegowosc.infor.pl/wiadomosci/7537062,nowy-jpk-vat-od-rozliczenia-za-luty-2026-r-dodatkowe-symbole-w-ewidencji-dot-ksef.html)
— *to verify against the MF brochure*). The file is built **per document** from the VAT registers (dates,
counterparty, tax bases and VAT by rate, GTU and procedure markings), so the chart does not feed it.

What the chart still needs (my reading): settlement accounts for **input VAT** and **output VAT**
(traditionally 221 / 222; in the JPK tags they land in `BABII3b` and `BPBIII3g`), an account for VAT
**settled or refunded** with the office, deferred/unsettled VAT (e.g. input VAT not yet deductible), and a
monthly reconciliation of registers to these accounts. JPK_KR_PD links the two worlds through the **KSeF
number** on journal entries (`D_12`). GUS SP also asks for **VAT payable and VAT refundable for the
calendar year** (section 4.3).

---

## 4. GUS — statistical reports fed from the books

Legal basis: art. 30 ust. 1 pkt 3 of the Act on Public Statistics and the yearly programme: **Rozporządzenie
Rady Ministrów z dnia 16 grudnia 2025 r. w sprawie programu badań statystycznych statystyki publicznej na
rok 2026** ([Dz.U. 2025 poz. 1842](https://api.sejm.gov.pl/eli/acts/DU/2025/1842/text.pdf); amended by
Dz.U. 2026 poz. 1136, not affecting these forms). Part II of its annex lists each form, who must file,
deadlines. Forms are filed on the GUS portal raport.stat.gov.pl.

### 4.1 F-01/I-01 — yes, "F1" still exists in 2026, under that name

**F-01/I-01 — Sprawozdanie o przychodach, kosztach i wyniku finansowym oraz o nakładach na środki trwałe**
(PBSSP 2026, Part II lp. 47; survey 1.61.01 *Bieżące wyniki finansowe i nakłady na środki trwałe
przedsiębiorstw oraz aktywa i pasywa finansowe*) ([Dz.U. 2025 poz. 1842](https://api.sejm.gov.pl/eli/acts/DU/2025/1842/text.pdf);
[2026 form](https://form.stat.gov.pl/formularze/2026/passive/F-01_I-01.pdf)).

| | |
|---|---|
| Who | Entities **keeping księgi rachunkowe** with **50 or more persons working** — quarterly; with **10–49** — half-yearly; main activity in PKD 2025 sections A (not individual farms), B–I, J (not cultural institutions with legal personality), K, L (not banks, SKOK, insurers, brokers, investment and pension funds and their managers), M, N, O, Q (not higher education), R (not SP ZOZ), S (not cultural institutions with legal personality), T (not trade unions, religious and political organisations); not budget units, executive agencies, state research units. Full observation (all who qualify). "Persons working" (*pracujący*), not only employees, as defined in the form's notes to row 77 |
| Deadlines for 2026 | ≥ 50: **20 April** (Q1), **20 July** (H1), **20 October** (3 quarters) 2026, **10 February 2027** (year). 10–49: **20 July 2026** (H1) and **10 February 2027** (year) |
| Amounts | thousands of PLN, **cumulative from the start of the calendar year**; companies with a non-calendar financial year report for the calendar year; foreign branches included ([form](https://form.stat.gov.pl/formularze/2026/passive/F-01_I-01.pdf)) |
| Basis | to be filled under the Accounting Act; adapted to **Annex 1** UoR; preliminary figures allowed if statements are not approved (same form) |

The same scope existed in 2025 (quarterly ≥ 50, half-yearly 10–49) ([PBSSP 2025, Dz.U. 2024 poz. 1653](https://api.sejm.gov.pl/eli/acts/DU/2024/1653/text.pdf)).

**Content of the 2026 form** ([F-01/I-01](https://form.stat.gov.pl/formularze/2026/passive/F-01_I-01.pdf)):

- Questions: type of records (only full books continue), foreign shareholders, UoR vs IFRS, IFRS 16.
- **Part I, Dział 1 — Rachunek zysków i strat, "tylko jedną wersję … tj. porównawczą"** — rows 01–65:
  net revenue from products (row 02) *of which export* (03), change in products (04), own work capitalised
  (05), revenue from goods (06) *of which export* (07); operating costs **by nature**: depreciation (09),
  materials and energy (10) *of which energy* (11), external services (12) *of which subcontracting* (13),
  taxes and charges (14) *of which excise* (15/16), wages (17) *of which from employment* (18), social
  insurance and other benefits (19) *of which social-insurance contributions* (20) and *training* (21),
  other costs by nature (22) *of which business travel* (23) and *property and personal insurance
  premiums* (24), value of goods sold (25); other operating income (28–33) incl. gain on disposal of
  non-financial fixed assets, revaluation, subsidies, other *of which released provisions* (33); other
  operating costs (34–39) incl. loss on disposal, revaluation, other *of which provisions for future
  liabilities* (38) and *receivables written off in insolvency proceedings* (39); finance income (42–50):
  dividends, interest *of which on loans granted and term deposits* (45) and *on loans to other
  enterprises* (46), gain on disposal of financial assets, revaluation, other *of which net FX gain* (50);
  finance costs (51–57): interest *of which on credits and loans* (53), loss on disposal, revaluation,
  other *of which net FX loss* (57); income tax current / deferred (60–62); other mandatory reductions
  (63); net result (64–65).
- **Supplementary data**: subsidies to core activity (66) and to products (67); **cash spent on debt
  service** — capital and interest paid to banks and non-financial enterprises, incl. finance-lease
  instalments (68); **imports** total (69) *of which raw materials for production* (70) and *goods for
  resale* (71) (EU acquisitions count as imports); excise (72–75); average employment in FTE (76),
  persons working (77).
- **Dział 2 — Aktywa** (opening 1.1.2026 and period end): non-current assets with long-term receivables
  and investments; inventories (materials, WIP, finished goods, goods, advances); short-term receivables
  *of which trade* and *tax/social*; short-term investments *of which financial assets* and *cash*;
  prepayments; called-up capital; own shares. An **overdraft is shown as a short-term loan, not negative
  cash** (instructions to row 15).
- **Dział 3 — Wybrane źródła finansowania**: provisions (*of which deferred tax*), long-term liabilities
  (*of which credits and loans, debt securities*), short-term liabilities without special funds (*of which
  trade, credits and loans, debt securities, advances received, taxes and social security, wages*),
  accruals.
- **Part II — Nakłady na środki trwałe**: investments started (number, cost estimate, *of which
  improvements*); **capex by KŚT group** — land (0), buildings and civil engineering (1–2) *of which
  drainage*, machinery (3–6), vehicles (7), tools and equipment (8), livestock (9) — *of which imported*,
  plus **purchases of used fixed assets** separately; **financing costs and FX differences** of capex
  liabilities; other capex costs; *of which construction works*; **intangibles by type** (copyrights and
  licences, software, mineral-deposit documentation, licences without software, patents and trademarks,
  completed development, goodwill); investments in property/intangibles held as investments; **leasing**
  (lessee: gross value of new fixed assets taken under finance / operating lease, *of which vehicles*).

**What F-01/I-01 requires of the chart** (my reading of the above):

1. **Costs by nature always**, even for a company that reports the calculation variant in its statements
   — team-4 accounts are not optional for a company with ≥ 10 persons working.
2. Sub-accounts or analytics for: energy within materials; **subcontracting** within external services;
   **excise** within taxes; wages **from employment** vs civil contracts; social-insurance contributions
   and **training** within benefits; **business travel** and **insurance premiums** within other costs.
3. **Export** split of product and goods revenue (incl. intra-EU supplies).
4. Other operating: **released provisions**, **provisions created**, **receivables written off in
   insolvency**; subsidies (with the core-activity / product split).
5. Finance: interest on **loans granted and deposits**, interest from **loans to non-financial
   enterprises**, interest on **credits and loans** taken; FX gains and losses kept so that the **net**
   figure per period can be derived.
6. **Imports** (raw materials, goods for resale) identifiable — a purchase-side analytic (counterparty
   country / document type), not necessarily separate accounts.
7. Capex by **KŚT group** with import and "used asset" flags; capitalised financing costs; intangibles by
   type; lease inflows by type.
8. Cash-flow-type item: **debt service paid** (principal + interest) in the period.
9. All figures **year-to-date by calendar year** even if the financial year differs.

### 4.2 F-02 — not in the current programmes

The PBSSP for 2023, 2024, 2025 and 2026 contain **no "F-02" form** for enterprises — only **F-02/dk**
(*roczne sprawozdanie o finansach instytucji kultury*) for cultural institutions (searched the full texts:
[2026](https://api.sejm.gov.pl/eli/acts/DU/2025/1842/text.pdf), [2025](https://api.sejm.gov.pl/eli/acts/DU/2024/1653/text.pdf),
[2024](https://api.sejm.gov.pl/eli/acts/DU/2023/2747/text.pdf), [2023](https://api.sejm.gov.pl/eli/acts/DU/2022/2453/text.pdf)).
The annual balance sheet + P&L of enterprises is collected through **SP** (4.3), and for the annual
enterprise survey (1.61.05) GUS also takes, from Ministry of Finance systems, the "elektroniczne
ustrukturyzowane sprawozdania finansowe jednostek wpisanych do rejestru przedsiębiorców KRS i podatników
PIT prowadzących księgi rachunkowe…" (data set lp. 26.31) ([PBSSP 2026](https://api.sejm.gov.pl/eli/acts/DU/2025/1842/text.pdf)).
When F-02 last existed is *to verify*.

### 4.3 SP and SP-3 — the annual enterprise surveys

- **SP — Roczna ankieta przedsiębiorstwa** (lp. 158): entities with **10 or more persons working** in the
  same PKD sections, full observation; for 2026 by **4 May 2027** ([Dz.U. 2025 poz. 1842](https://api.sejm.gov.pl/eli/acts/DU/2025/1842/text.pdf)).
  The form published in 2026 (for 2025, due 4 May 2026) ([SP form](https://form.stat.gov.pl/formularze/2026/passive/SP.pdf))
  contains: full **balance sheet** opening/closing with extra rows (cash *in hand*, **foreign** loans
  granted, bank credits *of which foreign*, loans *of which foreign*, IFRS right-of-use assets); a
  supplement with **accumulated depreciation** of PP&E and intangibles, **overdue** short-term receivables,
  overdue long- and short-term liabilities, and **share capital by owner type** (State Treasury, state
  legal persons, local government, domestic natural persons, other domestic private, foreign,
  dispersed); **P&L comparative variant** with F-01-type splits plus **long-term rental and operating
  lease services**, **temporary-agency services**, **occupational-medicine costs**, goods bought for
  resale but used internally, **revenue from sale of materials** and **value of materials sold**; a
  supplement with subsidies from the state/local budget and from EU funds, **VAT payable and VAT
  refundable for the calendar year**, debt service paid, losses in goods not captured by stock-taking;
  **fixed assets by KŚT group** with opening/closing gross value, increases and decreases (liquidation,
  sale, other), capex, used-asset purchases, **depreciation**, **finance leases**, **capex financing
  sources**, capex by direction; local units; trade data; foreign capital; business functions.
  Survey 1.61.05 also names the calculation-variant P&L, statement of changes in equity and cash flow among
  its subjects ([Dz.U. 2025 poz. 1842](https://api.sejm.gov.pl/eli/acts/DU/2025/1842/text.pdf)).
- **SP-3 — Sprawozdanie o działalności gospodarczej przedsiębiorstw** (lp. 159): **9 or fewer** persons
  working, **4% random sample**; by **17 March 2027** for 2026 (same source;
  [SP-3 form](https://form.stat.gov.pl/formularze/2026/passive/SP-3.pdf)).

Chart needs beyond 4.1: fixed-asset movements by KŚT group (from the register), overdue analysis
(from subledgers), **foreign** vs domestic lenders/borrowers, owner types of share capital (master data),
sale of materials separated from goods, rental/operating-lease and agency-work services within external
services, VAT payable/refundable per calendar year.

### 4.4 Other GUS forms that touch the books

All from PBSSP 2026 Part II ([Dz.U. 2025 poz. 1842](https://api.sejm.gov.pl/eli/acts/DU/2025/1842/text.pdf)):

| Form | Who | When | From the books |
|---|---|---|---|
| **DG-1** *Meldunek o działalności gospodarczej* (lp. 39) | 10–49 persons working, PKD B–K, M, N (not 72, 75), O, S, T (not 94), div. 02, class 03.11 — the programme shows a full-observation row and a 10% sample row with the same wording (*to verify which applies to whom*) | Monthly by the **5th working day** (December by 11.1.2027) | Monthly revenue from products and from goods, wages; industrial firms: export split ([DG-1 form](https://form.stat.gov.pl/formularze/2026/passive/DG-1.pdf)) — chart needs **monthly** revenue products/goods and export |
| **RF-01** *Kwartalne sprawozdanie o aktywach i pasywach finansowych* (lp. 133) | ≥ 50 persons working, keeping books, PKD 2025 B–H, L (non-financial), M, N, O | Quarterly: 4 May, 3 Aug, 2 Nov 2026, 22 Feb 2027 | Cash in PLN vs foreign currency; deposits at **residents vs non-residents**; debt securities by issuer type and original maturity (≤ 1 y / > 1 y); loans granted to **non-financial enterprises, households, others, non-residents**; quoted / unquoted shares by issuer sector; on the liability side credits and loans by lender sector (non-financial enterprises, MFIs, PFR and other public lenders…) ([RF-01 form](https://form.stat.gov.pl/formularze/2026/passive/RF-01.pdf)) — chart needs **counterparty sector, residency and original maturity** on financial assets and liabilities |
| **IF** *Sprawozdanie o wykorzystaniu instrumentów finansowych w przedsiębiorstwach niefinansowych* (lp. 72) | ≥ 50, PKD B–H, M–O, holding financial instruments, keeping full books | Yearly by 30 July 2027 | Cash flow, **currency structure of revenue and costs**, derivatives |
| **PNT-01** *Sprawozdanie o działalności badawczej i rozwojowej (B+R)* (lp. 116) | Any entity doing R&D, regardless of size | Yearly by 31 March 2027 | R&D expenditure — same R&D cost separation as CIT/BR |
| **PNT-02** *Sprawozdanie o innowacjach* (lp. 117) | ≥ 50 in B–E; samples of 10–49 and of some services | Yearly by 10 March 2027 | Includes P&L (comparative) items |
| **DNU-K / DNU-R** *sprawozdanie o międzynarodowej wymianie usług* (lp. 40–41) | DNU-K: residents with revenue **or** expenditure from services with non-residents **≥ PLN 2,000,000** (2025 or 2026), purposive sample; DNU-R: others trading services with non-residents, purposive sample | DNU-K quarterly by the 20th after the quarter; DNU-R by 31 March 2027 | Services bought from / sold to non-residents by type and country — chart needs **non-resident service** analytics (counterparty country + service type) |
| **GP** *sprawozdanie statystyczne dla grup przedsiębiorstw* (lp. 63) | Companies in groups, any size | Yearly by 30 July 2027 | Balance data, group structure |
| **AP-01** (lp. 7) | ≥ 10 with public capital | Yearly by 18 May 2027 | — |
| **F-03** *sprawozdanie o stanie i ruchu środków trwałych* (lp. 51) | Not ordinary companies: farming div. 01 without foreign capital, financial institutions, higher education, public units | — | — (ordinary companies report fixed assets in SP) |
| Z-03, Z-05, Z-06 etc. | employment surveys | — | Payroll, not the chart |

---

## 5. NBP — balance of payments and international investment position

Legal basis: **Rozporządzenie Ministra Rozwoju i Finansów z 9 sierpnia 2017 r. w sprawie przekazywania
Narodowemu Bankowi Polskiemu danych niezbędnych do sporządzania bilansu płatniczego oraz międzynarodowej
pozycji inwestycyjnej**, issued under art. 30 ust. 3 Prawo dewizowe; consolidated text
[Dz.U. 2024 poz. 1746](https://api.sejm.gov.pl/eli/acts/DU/2024/1746/text.pdf). Data come **from the
books** (§ 2 ust. 3 pkt 1); filed on the NBP reporting portal with an NBP-issued certificate (§ 13).
Amounts in foreign currency at the NBP average rate of the last working day of the period (§ 2 ust. 5–6).

| Rule | Who (residents other than public sector, investment firms, individuals) | Frequency, deadline | Forms |
|---|---|---|---|
| § 7 ust. 1 | total assets + liabilities **shown on the forms of Annexes 1–5** > **PLN 300m** at year end | **monthly**, 20 days after the month | Annex 1: AZ-DEP (deposits in foreign banks), AZ-NZR (real estate abroad), AZ-UDZ (shares in non-residents), AZ-KRE (loans to non-residents, finance-lease receivables), AZ-REP, AZ-POZ (other financial assets and receivables), PZ-KRE (**credits and loans from non-residents, finance-lease liabilities**), PZ-REP, PZ-POZ (other financial liabilities), IP-ASA / IP-ASP / IP-SYM (derivatives); Annex 2–3: PW-… securities; Annex 4: **AZ-KRH** (trade receivables from non-residents, advances paid), **PZ-KRH** (trade payables to non-residents, advances received); Annex 5: PZ-UDZ (equity held by non-residents), PW-PAN |
| § 7 ust. 2 | that total > **PLN 10m** and ≤ PLN 300m | **quarterly**, 26 days after the quarter | same |
| § 7 ust. 3 | those of ust. 1–2 holding ≥ 10% votes in a foreign entity or with a foreign branch, or in which a non-resident holds ≥ 10%; branches of foreign enterprises | **yearly by 31 May** | **AZ-IB** (outward FDI), **PZ-IB** (inward FDI: shareholders, transactions, **equity at year end, profit/loss for the year, dividends paid**) |
| § 8 | below the § 7 thresholds but with assets/liabilities **related to foreign trade** > **PLN 3m** at year end | quarterly, 26 days | Annex 4 (AZ-KRH, PZ-KRH) |
| § 9 | "niezależnie od" §§ 4, 5, 7: residents referred to there with long-term loans from non-residents / finance-lease liabilities to non-residents > **PLN 3m** at quarter end (whether this reaches residents below the § 7 thresholds is *to verify*) | quarterly, 26 days | **PZ-KDH** |
| § 11 | issuers of debt securities placed abroad > PLN 3m | 10 days | PW-EMI, PW-ZME |

Thresholds are on the **foreign** positions reported on the forms, not the balance-sheet total (my reading
of § 7 ust. 1: "łączna kwota aktywów i pasywów określonych w formularzach"). The annex-to-form mapping above
is from the order of the forms in the published text. NBP's explanatory notes v2.0 (June 2025) exist on
nbp.pl but could not be downloaded here (*not read*).

**Chart needs**: every receivable, payable, loan, deposit, lease and equity position identifiable as
**resident / non-resident** with the **counterparty country** (and for FDI the group relation), trade vs
financial, original maturity (long-term loans), currency; equity components and the year's result for
PZ-IB (my reading).

---

## 6. Other submissions KK should know about

- **Annual financial statement** — filed by the head of the unit in the court register (KRS / Repozytorium
  Dokumentów Finansowych) **within 15 days of approval** (art. 69 ust. 1 UoR), approval no later than
  **6 months** after the balance-sheet date (art. 53 ust. 1) ([UoR, Dz.U. 2026 poz. 522](https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf));
  taxpayers not filing to KRS send it to **Szef KAS** within 15 days of approval, in the logical structure
  of art. 45 ust. 1g UoR (art. 27 ust. 2 CIT) ([CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)).
  Positions: see the statements note.
- **Transfer pricing**: local file by the end of the **10th month**, documentation thresholds per
  homogeneous transaction **PLN 10m** (goods, financial), **PLN 2m** (services, other), lower for tax
  havens (art. 11k); **TPR-C(6)** information by the end of the **11th month**, built from the local file
  or the financial statements, with general financial data and the controlled transactions (art. 11t)
  ([CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf); form list
  [podatki.gov.pl](https://www.podatki.gov.pl/cit/e-deklaracje-cit/formularze-cit)). Chart needs: related
  parties identifiable on revenue, cost, receivable, payable and loan accounts — the same POW/POZ split the
  JPK tags need — plus transaction-type analytics (goods / services / financial / other).
- **Payment-terms report** (*sprawozdanie o stosowanych terminach zapłaty w transakcjach handlowych*,
  art. 13a of the Act on counteracting excessive delays): taxpayers in the MF list of CIT taxpayers with
  revenue > €50m (art. 27b ust. 2 pkt 2 CIT), by **30 April**: values paid/received on time and late in
  bands ≤ 5, 6–30, 31–60, 61–120, > 120 days, excluding intra-group transactions
  ([Dz.U. 2023 poz. 1790](https://api.sejm.gov.pl/eli/acts/DU/2023/1790/text.pdf); no amendment after 2023
  listed in the ELI record). Needs due dates and payment dates in the subledgers, not chart accounts.
- **Real-estate companies** (*spółka nieruchomościowa*): information on shareholders to Szef KAS by the end
  of the 3rd month (art. 27 ust. 1e CIT; CIT-8 poz. 321 asks the status) ([CIT Act](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)).
  The status itself depends on balance-sheet real-estate values (art. 4a pkt 35 — *not examined*).
- **Tax strategy report** (former art. 27c CIT): **repealed** from 7.8.2025 by the act of 25.6.2025
  (Dz.U. 2025 poz. 1074) ([CIT Act footnote 60](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf)).
- **Payroll-only filings** — ZUS DRA/RCA, PIT-4R/PIT-11, PFRON declarations (DEK-I-0 etc.), GUS Z-forms,
  CIT-8ST: fed from payroll and HR, not from the chart. The chart only needs the payroll cost and
  liability accounts (wages, ZUS, PIT, PFRON) that these postings land on (my reading; PFRON forms *not
  examined*).
- **Intrastat**, **CbC-R**, **MDR**: not fed from the chart (documents, group data, arrangements) — *not
  examined; thresholds not verified*.

---

## 7. What the model chart must carry — one list for the builder

A synthesis of sections 1–6 (my reading). Each item names the filing that drives it.

**Account attributes (on every account, versioned with the chart):**

1. `jpk_tag` — one Annex-7 tag (`S_12_1`), mandatory; optional second tag only for exceptions (JPK_KR_PD).
2. `jpk_pd_tag` — one of 28 PD tags where the account serves the tax computation (JPK_KR_PD).
3. `parent` — the account it rolls into, for `S_3` (JPK_KR_PD).
4. `cit_source` — capital gains / other sources (CIT-8 D.1–D.3) for revenue and cost accounts.
5. `kup_status` — deductible / non-deductible permanent (PD4 family) / non-deductible this year (PD5) /
   deductible this year booked earlier; revenue: taxable / exempt (PD1 family) / not taxable this year
   (PD2) (CIT-8, RPD, PD tags).
6. `gus_line` — the F-01/I-01 and SP comparative-variant line (cost **by nature** for every cost
   account, incl. team-5 accounts mapped back to nature where the company keeps both) (F-01/I-01, SP).
7. `party_relation` — related / with capital engagement / other, where the dictionary splits by it
   (JPK tags, TPR, CIT/M).
8. `maturity` — ≤ 12 / > 12 months for trade receivables and payables; long / short term for provisions,
   accruals, financial assets and liabilities; original maturity for RF-01 (JPK tags, RF-01, NBP).
9. `value_role` — gross value / accumulated depreciation / write-down / price variance (`_W`, `_U`, `_A`,
   `_O`) (JPK tags).
10. Flags: **R&D** (CIT/BR category), **IP right** (CIT/IP), **financing cost** (art. 15c, CIT/M),
    **interest-type revenue**, **excise**, **energy**, **subcontracting**, **training**, **business travel**,
    **insurance premiums**, **released / created provisions**, **Estonian-CIT hidden profit / non-business
    expense**, **investment fund** (art. 15 ust. 1hb), **FX gain / FX loss** kept separate.

**Posting / counterparty dimensions (on entries, not accounts):**

11. Counterparty with internal code, **NIP**, **country**, residency, institutional sector (non-financial
    enterprise, household, bank, public lender…), related-party flag (JPK_KR_PD, NBP, RF-01, TPR, DNU).
12. **KSeF number**, document number and type, operation / document / posting dates, person responsible,
    journal (JPK_KR_PD).
13. Currency and foreign-currency amount on every posting line (JPK_KR_PD, NBP, IF).
14. Export / import / intra-EU flags on revenue and purchases (F-01/I-01, SP, DG-1).
15. Due date and settlement date (art. 18f bad-debt relief, payment-terms report, overdue items in SP).

**Subledger that must reconcile to the chart:** the fixed-asset register with **tax and book values side by
side**, KŚT symbol, methods, acquisition and removal reasons (JPK_ST_KR, CIT-8, F-01/I-01 capex, SP).

**Periods:** financial year and tax year separately (JPK_KR_PD header); calendar-year YTD views for GUS
(F-01/I-01 asks for calendar year even when the financial year differs); monthly and quarterly cuts (DG-1,
F-01/I-01, NBP, CIT advances).

---

## 8. What remains to verify

- Title and content of the PIT tag regulation of 6.09.2025 (Dz.U. 2025 poz. 1311) — only cited via FAQ Q126.
- Status of the separate **IFRS tag dictionary** (FAQ Q115 says "in consultation").
- Whether MF will add **Estonian-CIT tax tags** or a **micro/small dictionary** (SKwP asked; not enacted
  as of Dz.U. 2025 poz. 1828).
- JPK_KR (on demand) field names for team/category codes (from a search snippet of the FAQ, not read in
  full).
- The year from which **CIT-8(35)** applies.
- JPK_V7(3) KSeF markers (OFF, BFK, DI) — confirmed only by a secondary source; read the MF brochure
  "JPK_VAT z deklaracją od 1 lutego 2026".
- **DG-1**: which rows (full observation vs 10% sample) apply to which 10–49 entities.
- When **F-02** was last in the programme (absent 2023–2026).
- NBP explanatory notes v2.0 (June 2025) — not downloadable here; the regulation text was used instead.
  Open point: whether § 9 (PZ-KDH, long-term foreign loans > PLN 3m) also binds residents below the § 7
  thresholds.
- CIT/8S and CIT/8SP record-keeping requirements for SEZ / support-decision activity — not examined.
- Intrastat thresholds, CbC-R, PFRON forms — not examined (not chart-fed).
- The 2026 SP form shown on form.stat.gov.pl is the one **for 2025**; the form for 2026 (due 4.5.2027) may
  differ.

---

## 9. Sources (primary first)

- CIT Act, consolidated text — [Dz.U. 2026 poz. 554](https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf); amendment [Dz.U. 2026 poz. 779](https://api.sejm.gov.pl/eli/acts/DU/2026/779/text.pdf) (JPK deadline, UPL-1).
- Tag regulation — [Dz.U. 2024 poz. 1314](https://api.sejm.gov.pl/eli/acts/DU/2024/1314/text.html); amendment [Dz.U. 2025 poz. 1828](https://eli.gov.pl/eli/DU/2025/1828/ogl/pol/pdf); JPK_ST_KR exemption 2025 [Dz.U. 2024 poz. 1861](https://api.sejm.gov.pl/eli/acts/DU/2024/1861/text.pdf); deadline extension [Dz.U. 2026 poz. 188](https://api.sejm.gov.pl/eli/acts/DU/2026/188/text.pdf).
- KAS — [Struktury JPK w podatkach dochodowych](https://www.gov.pl/web/kas/struktury-jpk-w-podatkach-dochodowych); [JPK_KR_PD XSD v1-1](https://www.gov.pl/attachment/4ec8f106-032c-42f2-9fb5-15b5ec5ca670); [JPK_ST_KR XSD v1-0](https://www.gov.pl/attachment/336fb09b-ea93-475d-a895-0ef5b8b72a0b); [brochure JPK_KR_PD(1)](https://www.gov.pl/attachment/22010dbd-5e27-433b-84c5-ec99cee7487c); [errata](https://www.gov.pl/attachment/8b2eb009-c827-409d-8cf1-0548f4290abc); [update from 1.7.2026](https://www.gov.pl/attachment/b16d282e-e1df-47df-ba7f-cfa973f4412e).
- podatki.gov.pl — [JPK_PD FAQ](https://www.podatki.gov.pl/podatki-firmowe/jednolity-plik-kontrolny/jpk_pd/pytania-i-odpowiedzi-jpk_pd); [JPK_PD downloads](https://www.podatki.gov.pl/podatki-firmowe/jednolity-plik-kontrolny/jpk_pd/pliki-do-pobrania-jpk_pd); [CIT forms](https://www.podatki.gov.pl/cit/e-deklaracje-cit/formularze-cit); [JPK_VAT z deklaracją](https://www.podatki.gov.pl/podatki-firmowe/jednolity-plik-kontrolny/jpk_vat-z-deklaracja/jpk_vat-z-deklaracja); [JPK_VAT downloads](https://www.podatki.gov.pl/podatki-firmowe/jednolity-plik-kontrolny/jpk_vat-z-deklaracja/pliki-do-pobrania).
- Forms — [CIT-8(34)](https://www.gov.pl/attachment/b6901dbb-05bf-4d5b-b1be-52fd86fc753d); [CIT-8/O(20)](https://www.podatki.gov.pl/media/czfdxieo/cit-8o-20-07.pdf); [CIT/BR(10)](https://www.podatki.gov.pl/media/cn1nighs/cit_br_1007.pdf); [CIT/M(2)](https://www.podatki.gov.pl/media/0nlhw5go/cit-m0206.pdf); [CIT-8E(4)](https://www.podatki.gov.pl/media/dhjhcy34/cit-8e_-4-9.pdf); [CIT-8ST(1)](https://www.podatki.gov.pl/media/fl5pwqca/cit-8st_-1-12.pdf); [CIT-8(35) consultation](https://www.gov.pl/web/finanse/konsultacje-podatkowe-wzoru-formularza-cit-8); [Estonian CIT guide](https://www.gov.pl/web/finanse/ryczalt-od-dochodow-spolek--objasnienia-podatkowe).
- GUS — PBSSP 2026 [Dz.U. 2025 poz. 1842](https://api.sejm.gov.pl/eli/acts/DU/2025/1842/text.pdf); PBSSP 2025 [Dz.U. 2024 poz. 1653](https://api.sejm.gov.pl/eli/acts/DU/2024/1653/text.pdf); 2024 [Dz.U. 2023 poz. 2747](https://api.sejm.gov.pl/eli/acts/DU/2023/2747/text.pdf); 2023 [Dz.U. 2022 poz. 2453](https://api.sejm.gov.pl/eli/acts/DU/2022/2453/text.pdf); forms [F-01/I-01](https://form.stat.gov.pl/formularze/2026/passive/F-01_I-01.pdf), [SP](https://form.stat.gov.pl/formularze/2026/passive/SP.pdf), [SP-3](https://form.stat.gov.pl/formularze/2026/passive/SP-3.pdf), [DG-1](https://form.stat.gov.pl/formularze/2026/passive/DG-1.pdf), [RF-01](https://form.stat.gov.pl/formularze/2026/passive/RF-01.pdf); [GUS Q1 2026 results (F-01/I-01 coverage)](https://stat.gov.pl/obszary-tematyczne/podmioty-gospodarcze-wyniki-finansowe/przedsiebiorstwa-niefinansowe/wyniki-finansowe-przedsiebiorstw-w-1-kwartale-2026-r-,43,1.html).
- NBP regulation — [Dz.U. 2024 poz. 1746](https://api.sejm.gov.pl/eli/acts/DU/2024/1746/text.pdf).
- Accounting Act, consolidated — [Dz.U. 2026 poz. 522](https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf).
- Late-payments act — [Dz.U. 2023 poz. 1790](https://api.sejm.gov.pl/eli/acts/DU/2023/1790/text.pdf).
- Professional: [SKwP opinion 15.10.2025](https://skwp.pl//content/uploads/2026/03/odpowiedz-dla-MFiG_61.pdf); [Infor on JPK_V7(3)](https://ksiegowosc.infor.pl/wiadomosci/7537062,nowy-jpk-vat-od-rozliczenia-za-luty-2026-r-dodatkowe-symbole-w-ewidencji-dot-ksef.html).

---

## Appendix A — JPK_KR_PD tag dictionary for ordinary companies (Annex 7, `TMapKontaPOZ`) and the PD tags

Source: the enumerations of `TMapKontaPOZ` and `TMapKontaPD` in the
[JPK_KR_PD(1) v1-1 XSD](https://www.gov.pl/attachment/4ec8f106-032c-42f2-9fb5-15b5ec5ca670), identical in
codes to Annex 7 of [Dz.U. 2024 poz. 1314](https://api.sejm.gov.pl/eli/acts/DU/2024/1314/text.html). The
description is the dictionary's own path, levels joined with "›".

#### Bilans – aktywa (BA…) — 135 znaczników

| Znacznik | Opis znacznika (pełna ścieżka ze słownika) |
|---|---|
| `BAAI1_W` | AKTYWA › Aktywa trwałe › Wartości niematerialne i prawne › Koszty zakończonych prac rozwojowych |
| `BAAI1_A` | AKTYWA › Aktywa trwałe › Wartości niematerialne i prawne › Koszty zakończonych prac rozwojowych › Odpisy aktualizujące wartość kosztów zakończonych prac rozwojowych |
| `BAAI1_U` | AKTYWA › Aktywa trwałe › Wartości niematerialne i prawne › Koszty zakończonych prac rozwojowych › Umorzenie zakończonych prac rozwojowych |
| `BAAI2_W` | AKTYWA › Aktywa trwałe › Wartości niematerialne i prawne › Wartość firmy |
| `BAAI2_A` | AKTYWA › Aktywa trwałe › Wartości niematerialne i prawne › Wartość firmy › Odpisy aktualizujące wartość firmy |
| `BAAI2_U` | AKTYWA › Aktywa trwałe › Wartości niematerialne i prawne › Wartość firmy › Umorzenie wartości firmy |
| `BAAI3_W` | AKTYWA › Aktywa trwałe › Wartości niematerialne i prawne › Inne wartości niematerialne i prawne |
| `BAAI3_A` | AKTYWA › Aktywa trwałe › Wartości niematerialne i prawne › Inne wartości niematerialne i prawne › Odpisy aktualizujące inne wartości niematerialne i prawne |
| `BAAI3_U` | AKTYWA › Aktywa trwałe › Wartości niematerialne i prawne › Inne wartości niematerialne i prawne › Umorzenie innych wartości niematerialnych i prawnych |
| `BAAI4_W` | AKTYWA › Aktywa trwałe › Wartości niematerialne i prawne › Zaliczki na wartości niematerialne i prawne |
| `BAAI4_A` | AKTYWA › Aktywa trwałe › Wartości niematerialne i prawne › Zaliczki na wartości niematerialne i prawne › Odpisy aktualizujące zaliczki na wartości niematerialne i prawne |
| `BAAII1a_W` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Grunty (w tym prawo użytkowania wieczystego gruntu) |
| `BAAII1a_A` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Grunty (w tym prawo użytkowania wieczystego gruntu) › Odpisy aktualizujące wartość gruntów (w tym prawa użytkowania wieczystego gruntu) |
| `BAAII1a_U` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Grunty (w tym prawo użytkowania wieczystego gruntu) › Umorzenie gruntów (w tym prawa użytkowania wieczystego gruntu) |
| `BAAII1b_W` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Budynki, lokale i obiekty inżynierii lądowej i wodnej |
| `BAAII1b_A` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Budynki, lokale i obiekty inżynierii lądowej i wodnej › Odpisy aktualizujące wartość budynków, lokali i obiektów inżynierii lądowej i wodnej |
| `BAAII1b_U` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Budynki, lokale i obiekty inżynierii lądowej i wodnej › Umorzenie budynków, lokali i obiektów inżynierii lądowej i wodnej |
| `BAAII1c_W` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Urządzenia techniczne i maszyny |
| `BAAII1c_A` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Urządzenia techniczne i maszyny › Odpisy aktualizujące wartość urządzeń technicznych i maszyn |
| `BAAII1c_U` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Urządzenia techniczne i maszyny › Umorzenie urządzeń technicznych i maszyn |
| `BAAII1d_W` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Środki transportu |
| `BAAII1d_A` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Środki transportu › Odpisy aktualizujące wartość środków transportu |
| `BAAII1d_U` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Środki transportu › Umorzenie środków transportu |
| `BAAII1e_W` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Inne środki trwałe |
| `BAAII1e_A` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Inne środki trwałe › Odpisy aktualizujące wartość innych środków trwałych |
| `BAAII1e_U` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe › Inne środki trwałe › Umorzenie innych środków trwałych |
| `BAAII2_W` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe w budowie |
| `BAAII2_A` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Środki trwałe w budowie › Odpisy aktualizujące wartość środków trwałych w budowie |
| `BAAII3_W` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Zaliczki na środki trwałe w budowie |
| `BAAII3_A` | AKTYWA › Aktywa trwałe › Rzeczowe aktywa trwałe › Zaliczki na środki trwałe w budowie › Odpisy aktualizujące wartość zaliczek na środki trwałe w budowie |
| `BAAIII1_W` | AKTYWA › Aktywa trwałe › Należności długoterminowe › Od jednostek powiązanych |
| `BAAIII1_A` | AKTYWA › Aktywa trwałe › Należności długoterminowe › Od jednostek powiązanych › Odpisy aktualizujące wartość należności od jednostek powiązanych |
| `BAAIII2_W` | AKTYWA › Aktywa trwałe › Należności długoterminowe › Od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale |
| `BAAIII2_A` | AKTYWA › Aktywa trwałe › Należności długoterminowe › Od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale › Odpisy aktualizujące wartość należności od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale |
| `BAAIII3_W` | AKTYWA › Aktywa trwałe › Należności długoterminowe › Od pozostałych jednostek |
| `BAAIII3_A` | AKTYWA › Aktywa trwałe › Należności długoterminowe › Od pozostałych jednostek › Odpisy aktualizujące wartość należności od pozostałych jednostek |
| `BAAIV1_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Nieruchomości |
| `BAAIV1_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Nieruchomości › Odpisy aktualizujące wartość inwestycji długoterminowych w nieruchomości |
| `BAAIV1_U` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Nieruchomości › Umorzenie inwestycji długoterminowych w nieruchomości |
| `BAAIV2_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Wartości niematerialne i prawne |
| `BAAIV2_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Wartości niematerialne i prawne › Odpisy aktualizujące wartość inwestycji długoterminowych w wartości niematerialne i prawne |
| `BAAIV2_U` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Wartości niematerialne i prawne › Umorzenie wartości inwestycji długoterminowych w wartości niematerialne i prawne |
| `BAAIV3a_ULA_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W jednostkach powiązanych › Udziały lub akcje |
| `BAAIV3a_ULA_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W jednostkach powiązanych › Udziały lub akcje › Odpisy aktualizujące wartość udziałów lub akcji w jednostkach powiązanych |
| `BAAIV3a_IPW_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W jednostkach powiązanych › Inne papiery wartościowe |
| `BAAIV3a_IPW_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W jednostkach powiązanych › Inne papiery wartościowe › Odpisy aktualizujące wartość innych papierów wartościowych w jednostkach powiązanych |
| `BAAIV3a_UP_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W jednostkach powiązanych › Udzielone pożyczki |
| `BAAIV3a_UP_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W jednostkach powiązanych › Udzielone pożyczki › Odpisy aktualizujące wartość udzielonych pożyczek w jednostkach powiązanych |
| `BAAIV3a_IDAF_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W jednostkach powiązanych › Inne długoterminowe aktywa finansowe |
| `BAAIV3a_IDAF_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W jednostkach powiązanych › Inne długoterminowe aktywa finansowe › Odpisy aktualizujące wartość innych długoterminowych aktywów finansowych w jednostkach powiązanych |
| `BAAIV3b_ULA_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale › Udziały lub akcje |
| `BAAIV3b_ULA_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale › Udziały lub akcje › Odpisy aktualizujące wartość udziałów lub akcji w pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale |
| `BAAIV3b_IPW_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale › Inne papiery wartościowe |
| `BAAIV3b_IPW_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale › Inne papiery wartościowe › Odpisy aktualizujące wartość innych papierów wartościowych w pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale |
| `BAAIV3b_UP_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale › Udzielone pożyczki |
| `BAAIV3b_UP_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale › Udzielone pożyczki › Odpisy aktualizujące wartość udzielonych pożyczek w pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale |
| `BAAIV3b_IDAF_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale › Inne długoterminowe aktywa finansowe |
| `BAAIV3b_IDAF_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale › Inne długoterminowe aktywa finansowe › Odpisy aktualizujące wartość innych długoterminowych aktywów finansowych w pozostałych jednostkach, w których jednostka posiada zaangażowanie w kapitale |
| `BAAIV3c_ULA_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach › Udziały lub akcje |
| `BAAIV3c_ULA_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach › Udziały lub akcje › Odpisy aktualizujące wartość udziałów lub akcji w pozostałych jednostkach |
| `BAAIV3c_IPW_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach › Inne papiery wartościowe |
| `BAAIV3c_IPW_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach › Inne papiery wartościowe › Odpisy aktualizujące wartość innych papierów wartościowych w pozostałych jednostkach |
| `BAAIV3c_UP_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach › Udzielone pożyczki |
| `BAAIV3c_UP_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach › Udzielone pożyczki › Odpisy aktualizujące wartość udzielonych pożyczek w pozostałych jednostkach |
| `BAAIV3c_IDAF_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach › Inne długoterminowe aktywa finansowe |
| `BAAIV3c_IDAF_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Długoterminowe aktywa finansowe › W pozostałych jednostkach › Inne długoterminowe aktywa finansowe › Odpisy aktualizujące wartość innych długoterminowych aktywów finansowych w pozostałych jednostkach |
| `BAAIV4_W` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Inne inwestycje długoterminowe |
| `BAAIV4_A` | AKTYWA › Aktywa trwałe › Inwestycje długoterminowe › Inne inwestycje długoterminowe › Odpisy aktualizujące wartość innych inwestycji długoterminowych |
| `BAAV1_W` | AKTYWA › Aktywa trwałe › Długoterminowe rozliczenia międzyokresowe › Aktywa z tytułu odroczonego podatku dochodowego |
| `BAAV1_A` | AKTYWA › Aktywa trwałe › Długoterminowe rozliczenia międzyokresowe › Aktywa z tytułu odroczonego podatku dochodowego › Odpisy aktualizujące aktywa z tytułu odroczonego podatku dochodowego |
| `BAAV2_W` | AKTYWA › Aktywa trwałe › Długoterminowe rozliczenia międzyokresowe › Inne rozliczenia międzyokresowe |
| `BAAV2_A` | AKTYWA › Aktywa trwałe › Długoterminowe rozliczenia międzyokresowe › Inne rozliczenia międzyokresowe › Odpisy aktualizujące inne rozliczenia międzyokresowe |
| `BABI1` | AKTYWA › Aktywa obrotowe › Zapasy › Materiały |
| `BABI1_O` | AKTYWA › Aktywa obrotowe › Zapasy › Materiały › Odchylenia od cen ewidencyjnych materiałów |
| `BABI1_A` | AKTYWA › Aktywa obrotowe › Zapasy › Materiały › Odpisy aktualizujące wartość materiałów |
| `BABI2` | AKTYWA › Aktywa obrotowe › Zapasy › Półprodukty i produkty w toku |
| `BABI2_O` | AKTYWA › Aktywa obrotowe › Zapasy › Półprodukty i produkty w toku › Odchylenia od cen ewidencyjnych półproduktów i produktów w toku |
| `BABI2_A` | AKTYWA › Aktywa obrotowe › Zapasy › Półprodukty i produkty w toku › Odpisy aktualizujące wartość półproduktów i produktów w toku |
| `BABI3` | AKTYWA › Aktywa obrotowe › Zapasy › Produkty gotowe |
| `BABI3_O` | AKTYWA › Aktywa obrotowe › Zapasy › Produkty gotowe › Odchylenia od cen ewidencyjnych produktów gotowych |
| `BABI3_A` | AKTYWA › Aktywa obrotowe › Zapasy › Produkty gotowe › Odpisy aktualizujące wartość produktów gotowych |
| `BABI4` | AKTYWA › Aktywa obrotowe › Zapasy › Towary |
| `BABI4_O` | AKTYWA › Aktywa obrotowe › Zapasy › Towary › Odchylenia od cen ewidencyjnych towarów |
| `BABI4_A` | AKTYWA › Aktywa obrotowe › Zapasy › Towary › Odpisy aktualizujące wartość towarów |
| `BABI5` | AKTYWA › Aktywa obrotowe › Zapasy › Zaliczki na dostawy i usługi |
| `BABI5_A` | AKTYWA › Aktywa obrotowe › Zapasy › Zaliczki na dostawy i usługi › Odpisy aktualizujące zaliczki na dostawy i usługi |
| `BABII1a_D12` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od jednostek powiązanych › Z tytułu dostaw i usług › o okresie spłaty do 12 miesięcy |
| `BABII1a_D12_A` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od jednostek powiązanych › Z tytułu dostaw i usług › o okresie spłaty do 12 miesięcy › Odpisy aktualizujące wartość należności z tytułu dostaw i usług od jednostek powiązanych w okresie spłaty do 12 miesięcy |
| `BABII1a_P12` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od jednostek powiązanych › Z tytułu dostaw i usług › o okresie spłaty powyżej 12 miesięcy |
| `BABII1a_P12_A` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od jednostek powiązanych › Z tytułu dostaw i usług › o okresie spłaty powyżej 12 miesięcy › Odpisy aktualizujące wartość należności z tytułu dostaw i usług od jednostek powiązanych w okresie spłaty powyżej 12 miesięcy |
| `BABII1b` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od jednostek powiązanych › Inne |
| `BABII1b_A` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od jednostek powiązanych › Inne › Odpisy aktualizujące wartość innych należności od jednostek powiązanych |
| `BABII2a_D12` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale › Z tytułu dostaw i usług › o okresie spłaty do 12 miesięcy |
| `BABII2a_D12_A` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale › Z tytułu dostaw i usług › o okresie spłaty do 12 miesięcy › Odpisy aktualizujące wartość należności z tytułu dostaw i usług od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale, o okresie spłaty do 12 miesięcy |
| `BABII2a_P12` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale › Z tytułu dostaw i usług › o okresie spłaty powyżej 12 miesięcy |
| `BABII2a_P12_A` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale › Z tytułu dostaw i usług › o okresie spłaty powyżej 12 miesięcy › Odpisy aktualizujące wartość należności z tytułu dostaw i usług od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale, o okresie spłaty powyżej 12 miesięcy |
| `BABII2b` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale › Inne |
| `BABII2b_A` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale › Inne › Odpisy aktualizujące wartość innych należności od pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale |
| `BABII3a_D12` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek › Z tytułu dostaw i usług › o okresie spłaty do 12 miesięcy |
| `BABII3a_D12_A` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek › Z tytułu dostaw i usług › o okresie spłaty do 12 miesięcy › Odpisy aktualizujące wartość należności z tytułu dostaw i usług od pozostałych jednostek o okresie spłaty do 12 miesięcy |
| `BABII3a_P12` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek › Z tytułu dostaw i usług › o okresie spłaty powyżej 12 miesięcy |
| `BABII3a_P12_A` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek › Z tytułu dostaw i usług › o okresie spłaty powyżej 12 miesięcy › Odpisy aktualizujące wartość należności z tytułu dostaw i usług od pozostałych jednostek o okresie spłaty powyżej 12 miesięcy |
| `BABII3b` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek › Z tytułu podatków, dotacji, ceł, ubezpieczeń społecznych i zdrowotnych oraz innych świadczeń |
| `BABII3b_A` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek › Z tytułu podatków, dotacji, ceł, ubezpieczeń społecznych i zdrowotnych oraz innych świadczeń › Odpisy aktualizujące wartość należności z tytułu podatków, dotacji, ceł, ubezpieczeń, społecznych i zdrowotnych oraz innych świadczeń |
| `BABII3c_W` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek › Inne |
| `BABII3c_A` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek › Inne › Odpisy aktualizujące wartość innych należności |
| `BABII3d_W` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek › Dochodzone na drodze sądowej |
| `BABII3d_A` | AKTYWA › Aktywa obrotowe › Należności krótkoterminowe › Należności od pozostałych jednostek › Dochodzone na drodze sądowej › Odpisy aktualizujące wartość należności dochodzonych na drodze sądowej |
| `BABIII1a1_ULA_W` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W jednostkach powiązanych › Udziały lub akcje |
| `BABIII1a1_ULA_A` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W jednostkach powiązanych › Udziały lub akcje › Odpisy aktualizujące wartość udziałów lub akcji w jednostkach powiązanych |
| `BABIII1a2_IPW_W` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W jednostkach powiązanych › Inne papiery wartościowe |
| `BABIII1a2_IPW_A` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W jednostkach powiązanych › Inne papiery wartościowe › Odpisy aktualizujące wartość innych papierów wartościowych w jednostkach powiązanych |
| `BABIII1a3_UP_W` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W jednostkach powiązanych › Udzielone pożyczki |
| `BABIII1a3_UP_A` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W jednostkach powiązanych › Udzielone pożyczki › Odpisy aktualizujące wartość udzielonych pożyczek w jednostkach powiązanych |
| `BABIII1a4_IKAF_W` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W jednostkach powiązanych › Inne krótkoterminowe aktywa finansowe |
| `BABIII1a4_IKAF_A` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W jednostkach powiązanych › Inne krótkoterminowe aktywa finansowe › Odpisy aktualizujące wartość innych krótkoterminowych aktywów finansowych w jednostkach powiązanych |
| `BABIII1b1_ULA_W` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W pozostałych jednostkach › Udziały lub akcje |
| `BABIII1b1_ULA_A` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W pozostałych jednostkach › Udziały lub akcje › Odpisy aktualizujące wartość udziałów lub akcji w pozostałych jednostkach |
| `BABIII1b2_IPW_W` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W pozostałych jednostkach › Inne papiery wartościowe |
| `BABIII1b2_IPW_A` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W pozostałych jednostkach › Inne papiery wartościowe › Odpisy aktualizujące wartość innych papierów wartościowych w pozostałych jednostkach |
| `BABIII1b3_UP_W` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W pozostałych jednostkach › Udzielone pożyczki |
| `BABIII1b3_UP_A` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W pozostałych jednostkach › Udzielone pożyczki › Odpisy aktualizujące wartość udzielonych pożyczek w pozostałych jednostkach |
| `BABIII1b4_IKAF_W` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W pozostałych jednostkach › Inne krótkoterminowe aktywa finansowe |
| `BABIII1b4_IKAF_A` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › W pozostałych jednostkach › Inne krótkoterminowe aktywa finansowe › Odpisy aktualizujące wartość innych krótkoterminowych aktywów finansowych w pozostałych jednostkach |
| `BABIII1c1_SPKR` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › Środki pieniężne i inne aktywa pieniężne › Środki pieniężne w kasie i na rachunkach |
| `BABIII1c2_ISP` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › Środki pieniężne i inne aktywa pieniężne › Inne środki pieniężne |
| `BABIII1c2_ISP_A` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › Środki pieniężne i inne aktywa pieniężne › Inne środki pieniężne › Odpisy aktualizujące wartość innych środków pieniężnych |
| `BABIII1c3_IAP` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › Środki pieniężne i inne aktywa pieniężne › Inne aktywa pieniężne |
| `BABIII1c3_IAP_A` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Krótkoterminowe aktywa finansowe › Środki pieniężne i inne aktywa pieniężne › Inne aktywa pieniężne › Odpisy aktualizujące wartość innych aktywów pieniężnych |
| `BABIII2_W` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Inne inwestycje krótkoterminowe |
| `BABIII2_A` | AKTYWA › Aktywa obrotowe › Inwestycje krótkoterminowe › Inne inwestycje krótkoterminowe › Odpisy aktualizujące wartość innych inwestycji krótkoterminowych |
| `BABIV` | AKTYWA › Aktywa obrotowe › Krótkoterminowe rozliczenia międzyokresowe |
| `BABIV_A` | AKTYWA › Aktywa obrotowe › Krótkoterminowe rozliczenia międzyokresowe › Odpisy aktualizujące wartość krótkoterminowych rozliczeń międzyokresowych |
| `BAC` | AKTYWA › Należne wpłaty na kapitał (fundusz) podstawowy |
| `BAD` | AKTYWA › Udziały (akcje) własne |

#### Bilans – pasywa (BP…) — 51 znaczników

| Znacznik | Opis znacznika (pełna ścieżka ze słownika) |
|---|---|
| `BPAI` | PASYWA › Kapitał (fundusz) własny › Kapitał (fundusz) podstawowy |
| `BPAIIWS` | PASYWA › Kapitał (fundusz) własny › Kapitał (fundusz) zapasowy › Z nadwyżki wartości sprzedaży (wartości emisyjnej) nad wartością nominalną udziałów (akcji) |
| `BPAII_INN` | PASYWA › Kapitał (fundusz) własny › Kapitał (fundusz) zapasowy › Z innych źródeł |
| `BPAII_ES_ZP` | PASYWA › Kapitał (fundusz) własny › Kapitał (fundusz) zapasowy › Kwota zysku wypracowanego w latach przed opodatkowaniem ryczałtem od dochodów spółek |
| `BPAII_ES_ZR` | PASYWA › Kapitał (fundusz) własny › Kapitał (fundusz) zapasowy › Kwota zysku wypracowanego w latach opodatkowania ryczałtem od dochodów spółek |
| `BPAIIIAWG` | PASYWA › Kapitał (fundusz) własny › Kapitał (fundusz) z aktualizacji wyceny › Z tytułu aktualizacji wartości godziwej |
| `BPAIII_INN` | PASYWA › Kapitał (fundusz) własny › Kapitał (fundusz) z aktualizacji wyceny › Z innego tytułu |
| `BPAIVU` | PASYWA › Kapitał (fundusz) własny › Pozostałe kapitały (fundusze) rezerwowe › Tworzone zgodnie z umową (ze statutem) |
| `BPAIVUW` | PASYWA › Kapitał (fundusz) własny › Pozostałe kapitały (fundusze) rezerwowe › Na udziały (akcje) własne |
| `BPAIV_INN` | PASYWA › Kapitał (fundusz) własny › Pozostałe kapitały (fundusze) rezerwowe › Inne |
| `BPAIV_ES_ZP` | PASYWA › Kapitał (fundusz) własny › Pozostałe kapitały (fundusze) rezerwowe › Kwota zysku wypracowanego w latach przed opodatkowaniem ryczałtem od dochodów spółek |
| `BPAIV_ES_ZR` | PASYWA › Kapitał (fundusz) własny › Pozostałe kapitały (fundusze) rezerwowe › Kwota zysku wypracowanego w latach opodatkowania ryczałtem od dochodów spółek |
| `BPAV` | PASYWA › Kapitał (fundusz) własny › Zysk (strata) z lat ubiegłych |
| `BPAV_ES_NZP` | PASYWA › Kapitał (fundusz) własny › Zysk (strata) z lat ubiegłych › Kwota niepodzielonych zysków wypracowanych w latach przed opodatkowaniem ryczałtem od dochodów spółek |
| `BPAV_ES_NPR` | PASYWA › Kapitał (fundusz) własny › Zysk (strata) z lat ubiegłych › Kwota niepodzielonych zysków wypracowanych w latach opodatkowania ryczałtem od dochodów spółek |
| `BPAV_ES_NSP` | PASYWA › Kapitał (fundusz) własny › Zysk (strata) z lat ubiegłych › Kwota niepokrytych strat poniesionych w latach przed opodatkowaniem ryczałtem od dochodów spółek |
| `BPAV_ES_NSR` | PASYWA › Kapitał (fundusz) własny › Zysk (strata) z lat ubiegłych › Kwota niepokrytych strat poniesionych w latach opodatkowania ryczałtem od dochodów spółek |
| `BPAVI` | PASYWA › Kapitał (fundusz) własny › Zysk (strata) netto |
| `BPAVII` | PASYWA › Kapitał (fundusz) własny › Odpisy z zysku netto w ciągu roku obrotowego (wielkość ujemna) |
| `BPBI1` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Rezerwy na zobowiązania › Rezerwa z tytułu odroczonego podatku dochodowego |
| `BPBI2_D` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Rezerwy na zobowiązania › Rezerwa na świadczenia emerytalne i podobne › Długoterminowa |
| `BPBI2_K` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Rezerwy na zobowiązania › Rezerwa na świadczenia emerytalne i podobne › Krótkoterminowa |
| `BPBI3_D` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Rezerwy na zobowiązania › Pozostałe rezerwy › Długoterminowe |
| `BPBI3_K` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Rezerwy na zobowiązania › Pozostałe rezerwy › Krótkoterminowe |
| `BPBII1` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania długoterminowe › Wobec jednostek powiązanych |
| `BPBII2` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania długoterminowe › Wobec pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale |
| `BPBII3a` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania długoterminowe › Wobec pozostałych jednostek › Kredyty i pożyczki |
| `BPBII3b` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania długoterminowe › Wobec pozostałych jednostek › Z tytułu emisji dłużnych papierów wartościowych |
| `BPBII3c` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania długoterminowe › Wobec pozostałych jednostek › Inne zobowiązania finansowe |
| `BPBII3d` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania długoterminowe › Wobec pozostałych jednostek › Zobowiązania wekslowe |
| `BPBII3e` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania długoterminowe › Wobec pozostałych jednostek › Inne |
| `BPBIII1a_D12` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec jednostek powiązanych › Z tytułu dostaw i usług › O okresie wymagalności do 12 miesięcy |
| `BPBIII1a_P12` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec jednostek powiązanych › Z tytułu dostaw i usług › O okresie wymagalności powyżej 12 miesięcy |
| `BPBIII1b` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec jednostek powiązanych › Inne |
| `BPBIII2a_D12` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale › Z tytułu dostaw i usług › O okresie wymagalności do 12 miesięcy |
| `BPBIII2a_P12` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale › Z tytułu dostaw i usług › O okresie wymagalności powyżej 12 miesięcy |
| `BPBIII2b` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek, w których jednostka posiada zaangażowanie w kapitale › Inne |
| `BPBIII3a` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek › Kredyty i pożyczki |
| `BPBIII3b` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek › Z tytułu emisji dłużnych papierów wartościowych |
| `BPBIII3c` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek › Inne zobowiązania finansowe |
| `BPBIII3d_D12` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek › Z tytułu dostaw i usług › O okresie wymagalności do 12 miesięcy |
| `BPBIII3d_P12` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek › Z tytułu dostaw i usług › O okresie wymagalności powyżej 12 miesięcy |
| `BPBIII3e` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek › Zaliczki otrzymane na dostawy i usługi |
| `BPBIII3f` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek › Zobowiązania wekslowe |
| `BPBIII3g` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek › Z tytułu podatków, ceł, ubezpieczeń i innych świadczeń |
| `BPBIII3h` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek › Z tytułu wynagrodzeń |
| `BPBIII3i` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Wobec pozostałych jednostek › Inne |
| `BPBIII4` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Zobowiązania krótkoterminowe › Fundusze specjalne |
| `BPBIV1` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Rozliczenia międzyokresowe › Ujemna wartość firmy |
| `BPBIV2_D` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Rozliczenia międzyokresowe › Inne rozliczenia międzyokresowe › Długoterminowe |
| `BPBIV2_K` | PASYWA › Zobowiązania i rezerwy na zobowiązania › Rozliczenia międzyokresowe › Inne rozliczenia międzyokresowe › Krótkoterminowe |

#### Rachunek zysków i strat (RP… = wariant porównawczy, RK… = wariant kalkulacyjny; kod łączony = ta sama pozycja w obu wariantach) — 57 znaczników

| Znacznik | Opis znacznika (pełna ścieżka ze słownika) |
|---|---|
| `RPAI_RKAI_POW` | WYNIKOWE › Przychody netto ze sprzedaży i zrównane z nimi / przychody netto ze sprzedaży produktów, towarów i materiałów › Przychody netto ze sprzedaży produktów › Od jednostek powiązanych |
| `RPAI_RKAI_POZ` | WYNIKOWE › Przychody netto ze sprzedaży i zrównane z nimi / przychody netto ze sprzedaży produktów, towarów i materiałów › Przychody netto ze sprzedaży produktów › Od pozostałych jednostek |
| `RPAII` | WYNIKOWE › Przychody netto ze sprzedaży i zrównane z nimi › Zmiana stanu produktów |
| `RPAIII` | WYNIKOWE › Przychody netto ze sprzedaży i zrównane z nimi › Koszt wytworzenia produktów na własne potrzeby |
| `RPAIV_RKAII_POW` | WYNIKOWE › Przychody netto ze sprzedaży i zrównane z nimi / przychody netto ze sprzedaży produktów, towarów i materiałów › Przychody netto ze sprzedaży towarów i materiałów › Od jednostek powiązanych |
| `RPAIV_RKAII_POZ` | WYNIKOWE › Przychody netto ze sprzedaży i zrównane z nimi / przychody netto ze sprzedaży produktów, towarów i materiałów › Przychody netto ze sprzedaży towarów i materiałów › Od pozostałych jednostek |
| `RPBI` | WYNIKOWE › Koszty działalności operacyjnej › Amortyzacja |
| `RPBII` | WYNIKOWE › Koszty działalności operacyjnej › Zużycie materiałów i energii |
| `RPBIII` | WYNIKOWE › Koszty działalności operacyjnej › Usługi obce |
| `RPBIV_AKC` | WYNIKOWE › Koszty działalności operacyjnej › Podatki i opłaty › Podatek akcyzowy |
| `RPBIV_POZ` | WYNIKOWE › Koszty działalności operacyjnej › Podatki i opłaty › Pozostałe podatki i opłaty |
| `RPBV` | WYNIKOWE › Koszty działalności operacyjnej › Wynagrodzenia |
| `RPBVI` | WYNIKOWE › Koszty działalności operacyjnej › Ubezpieczenia społeczne i inne świadczenia |
| `RPBVI_E` | WYNIKOWE › Koszty działalności operacyjnej › Ubezpieczenia społeczne i inne świadczenia › Emerytalne |
| `RPBVII` | WYNIKOWE › Koszty działalności operacyjnej › Pozostałe koszty rodzajowe |
| `RPBVIII_RKBII_POW` | WYNIKOWE › Koszty działalności operacyjnej / koszty sprzedanych produktów, towarów i materiałów › Wartość sprzedanych towarów i materiałów › Dla jednostek powiązanych |
| `RPBVIII_RKBII_POZ` | WYNIKOWE › Koszty działalności operacyjnej / koszty sprzedanych produktów, towarów i materiałów › Wartość sprzedanych towarów i materiałów › Dla pozostałych jednostek |
| `RKBI_POW` | WYNIKOWE › Koszty sprzedanych produktów, towarów i materiałów › Koszt wytworzenia sprzedanych produktów › Jednostkom powiązanym |
| `RKBI_POZ` | WYNIKOWE › Koszty sprzedanych produktów, towarów i materiałów › Koszt wytworzenia sprzedanych produktów › Pozostałym jednostkom |
| `RKD` | WYNIKOWE › Koszty sprzedaży |
| `RKE` | WYNIKOWE › Koszty ogólnego zarządu |
| `RPDI_RKGI_POW` | WYNIKOWE › Pozostałe przychody operacyjne › Przychody ze zbycia niefinansowych aktywów trwałych › Od jednostek powiązanych |
| `RPDI_RKGI_POZ` | WYNIKOWE › Pozostałe przychody operacyjne › Przychody ze zbycia niefinansowych aktywów trwałych › Od pozostałych jednostek |
| `RPDII_RKGII` | WYNIKOWE › Pozostałe przychody operacyjne › Dotacje |
| `RPDIII_RKGIII` | WYNIKOWE › Pozostałe przychody operacyjne › Aktualizacja wartości aktywów niefinansowych |
| `RPDIV_RKGIV` | WYNIKOWE › Pozostałe przychody operacyjne › Inne pozostałe przychody operacyjne |
| `RPDIV_RKGIV_ZCP` | WYNIKOWE › Pozostałe przychody operacyjne › Inne pozostałe przychody operacyjne › Ze zbycia zorganizowanej części przedsiębiorstwa |
| `RPDIV_RKGIV_WF` | WYNIKOWE › Pozostałe przychody operacyjne › Inne pozostałe przychody operacyjne › Odpis ujemnej wartości firmy |
| `RPDIV_RKGIV_P` | WYNIKOWE › Pozostałe przychody operacyjne › Inne pozostałe przychody operacyjne › Pozostałe |
| `RPEI_RKHI_POW` | WYNIKOWE › Pozostałe koszty operacyjne › Koszty zbycia niefinansowych aktywów trwałych › Dla jednostek powiązanych |
| `RPEI_RKHI_POZ` | WYNIKOWE › Pozostałe koszty operacyjne › Koszty zbycia niefinansowych aktywów trwałych › Dla pozostałych jednostek |
| `RPEII_RKHII` | WYNIKOWE › Pozostałe koszty operacyjne › Aktualizacja wartości aktywów niefinansowych |
| `RPEIII_RKHIII` | WYNIKOWE › Pozostałe koszty operacyjne › Inne |
| `RPEIII_RKHIII_ZCP` | WYNIKOWE › Pozostałe koszty operacyjne › Inne › Koszty zbycia zorganizowanej części przedsiębiorstwa |
| `RPEIII_RKHIII_WF` | WYNIKOWE › Pozostałe koszty operacyjne › Inne › Odpis wartości firmy |
| `RPEIII_RKHIII_P` | WYNIKOWE › Pozostałe koszty operacyjne › Inne › Pozostałe |
| `RPGIa_RKJIa_POW_k` | WYNIKOWE › Przychody finansowe › Dywidendy i udziały w zyskach › Od jednostek powiązanych › W których jednostka posiada zaangażowanie w kapitale |
| `RPGIa_RKJIa_POW_p` | WYNIKOWE › Przychody finansowe › Dywidendy i udziały w zyskach › Od jednostek powiązanych › W innych jednostkach powiązanych |
| `RPGIb_RKJIb_POZ_k` | WYNIKOWE › Przychody finansowe › Dywidendy i udziały w zyskach › Od pozostałych jednostek › W których jednostka posiada zaangażowanie w kapitale |
| `RPGIb_RKJIb_POZ_p` | WYNIKOWE › Przychody finansowe › Dywidendy i udziały w zyskach › Od pozostałych jednostek › W innych pozostałych jednostkach |
| `RPGII_RKJII_POW` | WYNIKOWE › Przychody finansowe › Odsetki › Od jednostek powiązanych |
| `RPGII_RKJII_POZ` | WYNIKOWE › Przychody finansowe › Odsetki › Od pozostałych jednostek |
| `RPGIII_RKJIII_POW` | WYNIKOWE › Przychody finansowe › Przychód z tytułu rozchodu aktywów finansowych › W jednostkach powiązanych |
| `RPGIII_RKJIII_POZ` | WYNIKOWE › Przychody finansowe › Przychód z tytułu rozchodu aktywów finansowych › W pozostałych jednostkach |
| `RPGIV_RKJIV` | WYNIKOWE › Przychody finansowe › Aktualizacja wartości aktywów finansowych |
| `RPGV_RKJV_RK` | WYNIKOWE › Przychody finansowe › Inne › Przychody z tytułu dodatnich różnic kursowych |
| `RPGV_RKJV_POZ` | WYNIKOWE › Przychody finansowe › Inne › Pozostałe |
| `RPHI_RKKI_POW` | WYNIKOWE › Koszty finansowe › Odsetki › Dla jednostek powiązanych |
| `RPHI_RKKI_POZ` | WYNIKOWE › Koszty finansowe › Odsetki › Dla pozostałych jednostek |
| `RPHII_RKKII_POW` | WYNIKOWE › Koszty finansowe › Koszty z tytułu rozchodu aktywów finansowych › W jednostkach powiązanych |
| `RPHII_RKKII_POZ` | WYNIKOWE › Koszty finansowe › Koszty z tytułu rozchodu aktywów finansowych › W pozostałych jednostkach |
| `RPHIII_RKKIII` | WYNIKOWE › Koszty finansowe › Aktualizacja wartości aktywów finansowych |
| `RPHIV_RKKIV_RK` | WYNIKOWE › Koszty finansowe › Inne › Koszty z tytułu ujemnych różnic kursowych |
| `RPHIV_RKKIV_POZ` | WYNIKOWE › Koszty finansowe › Inne › Pozostałe |
| `RPJ_RKM_B` | WYNIKOWE › Podatek dochodowy › Bieżący |
| `RPJ_RKM_O` | WYNIKOWE › Podatek dochodowy › Odroczony |
| `RPK_RKN` | WYNIKOWE › Pozostałe obowiązkowe zmniejszenia zysku (zwiększenia straty) |

#### Pozostałe — 1 znacznik

| Znacznik | Opis znacznika (pełna ścieżka ze słownika) |
|---|---|
| `INNE` | Inne, niewskazane w słowniku |

#### Znaczniki podatkowe PD (pole S_12_3) — 28 znaczników

| Znacznik | Opis znacznika |
|---|---|
| `PD1` | Przychody zwolnione z opodatkowania (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) |
| `PD1_1` | Przychody zwolnione z opodatkowania (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Przychody, w stosunku do których nie stosuje się przepisów ustawy |
| `PD1_2` | Przychody zwolnione z opodatkowania (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Przychody, które nie są zaliczane do przychodów podatkowych |
| `PD1_3` | Przychody zwolnione z opodatkowania (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Dochody (przychody) wolne |
| `PD2` | Przychody niepodlegające opodatkowaniu w roku bieżącym |
| `PD4` | Koszty niestanowiące kosztów uzyskania przychodów (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) |
| `PD4_1` | Koszty niestanowiące kosztów uzyskania przychodów (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Koszty związane z przychodami w stosunku do których przepisów ustawy o podatku dochodowym nie stosuje się |
| `PD4_2` | Koszty niestanowiące kosztów uzyskania przychodów (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Koszty związane z przychodami, które nie są zaliczane do przychodów podatkowych |
| `PD4_3` | Koszty niestanowiące kosztów uzyskania przychodów (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Koszty związane z dochodami (przychodami) wolnymi |
| `PD5` | Koszty nieuznawane za koszty uzyskania przychodów w bieżącym roku |
| `PD7` | Koszty uzyskania przychodów poniesione na działalność badawczo-rozwojową (ulga B+R) |
| `PD8_1` | Kwalifikowane prawa własności intelektualnej (IP Box) – przychody będące przychodami podatkowymi |
| `PD8_2` | Kwalifikowane prawa własności intelektualnej (IP Box) – koszty stanowiące koszty uzyskania przychodu |
| `PD1_PB` | POZABILANSOWE › Przychody zwolnione z opodatkowania (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) |
| `PD1_PB_1` | POZABILANSOWE › Przychody zwolnione z opodatkowania (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Przychody, w stosunku do których nie stosuje się przepisów ustawy |
| `PD1_PB_2` | POZABILANSOWE › Przychody zwolnione z opodatkowania (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Przychody, które nie są zaliczane do przychodów podatkowych |
| `PD1_PB_3` | POZABILANSOWE › Przychody zwolnione z opodatkowania (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Dochody (przychody) wolne |
| `PD2_PB` | POZABILANSOWE › Przychody niepodlegające opodatkowaniu w roku bieżącym |
| `PD3_PB` | POZABILANSOWE › Przychody podlegające opodatkowaniu w roku bieżącym, ujęte w księgach rachunkowych lat ubiegłych |
| `PD4_PB` | POZABILANSOWE › Koszty niestanowiące kosztów uzyskania przychodów (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) |
| `PD4_PB_1` | POZABILANSOWE › Koszty niestanowiące kosztów uzyskania przychodów (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Koszty związane z przychodami, w stosunku do których nie stosuje się przepisów ustawy |
| `PD4_PB_2` | POZABILANSOWE › Koszty niestanowiące kosztów uzyskania przychodów (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Koszty związane z przychodami, które nie są zaliczane do przychodów podatkowych |
| `PD4_PB_3` | POZABILANSOWE › Koszty niestanowiące kosztów uzyskania przychodów (trwałe różnice pomiędzy zyskiem/stratą dla celów rachunkowych a dochodem/stratą dla celów podatkowych) › Koszty związane z dochodami (przychodami) wolnymi |
| `PD5_PB` | POZABILANSOWE › Koszty nieuznawane za koszty uzyskania przychodów w bieżącym roku |
| `PD6_PB` | POZABILANSOWE › Koszty uznawane za koszty uzyskania przychodów w roku bieżącym, ujęte w księgach lat ubiegłych |
| `PD7_PB` | POZABILANSOWE › Koszty uzyskania przychodów poniesione na działalność badawczo-rozwojową (ulga B+R) |
| `PD8_PB_1` | POZABILANSOWE › Kwalifikowane prawa własności intelektualnej (IP Box) – przychody będące przychodami podatkowymi |
| `PD8_PB_2` | POZABILANSOWE › Kwalifikowane prawa własności intelektualnej (IP Box) – koszty stanowiące koszty uzyskania przychodu |
