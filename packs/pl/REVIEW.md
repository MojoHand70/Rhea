# The Poland pack, read whole — the register

**2026-10-16.** Item 2 of the plan after the statutory floor (DIRECTION, "Reading the
Poland pack"). The pack (v18) and `markets/pl` read against the year two companies
live through: a **sole trader** on KPiR or ryczałt, and a **small sp. z o.o.** on full
books. One Poland pack serves both; the company's declared form of books and tax is
the variant (KK, 2026-10-16).

**How to read an entry.** Each point is a question the books must answer, in the
order a business meets it during the year, with:

- **Applies to** — which company it concerns.
- **The law** — the act and unit, with the decisive words quoted from the
  consolidated text (ELI/ISAP, state of law September–October 2026).
- **Grade** — *1 plain* (the text decides), *2 settled* (the text is unclear, but
  the Ministry, a court, the TSUE, ZUS or a KSR settled it — named), *3 open*
  (practice splits — both readings named, with who holds each).
- **Rhea's default** — the reading the pack should ship.
- **Pack today** — the rule, type or Go file that does it, or *missing*, or
  *wrong — why*.
- **Verdict** — *right*, *wrong*, *missing* or *later* (not needed in a first year).
- **Deadline** — the statutory date, when there is one.

Claims found only in secondary sources are marked "(secondary — verify)".
Amounts the reviewers computed themselves (złoty equivalents of euro limits from NBP
rates, a deadline counted in days) say so. Each section ends its sources inline.

Five researchers, one area each, wrote the sections below. Nothing here has been read
by a practising accountant yet.

## Summary

| Area | Points | Right | Wrong | Missing | Later | Plain | Settled | Open |
|---|---|---|---|---|---|---|---|---|
| VAT — sales and purchases | 33 | 2 | 8 | 20 | 3 | 29 | 4 | 0 |
| KSeF and JPK | 41 | 5 | 11 | 24 | 1 | 38 | 3 | 0 |
| Sole trader — books and income tax | 46 | 1 | 1 | 41 | 3 | 43 | 3 | 0 |
| The owner and money | 48 | 2 | 12 | 31 | 3 | 44 | 2 | 2 |
| Full books and CIT | 44 | 5 | 5 | 31 | 3 | 41 | 3 | 0 |
| **Total** | **212** | **15** | **37** | **147** | **13** | **195** | **15** | **2** |

**What the numbers say.**

- **The law a small firm meets in its first year is mostly plain.** 195 of 212 points
  are decided by the text. Where something felt ambiguous, it was usually one of two
  things: the law changed in 2026 (the KSeF obligation, a new KPiR regulation,
  corrections without the agreed-terms condition, new car limits, the advances at
  9%), or a choice the law leaves to the company (the cost method, the
  simplifications in the accounting policy). A choice is not an ambiguity: the
  company makes it, Rhea records it as an approved rule.
- **Two points are genuinely open**, and both are small and frequent — the kind a
  standard could settle:
  - **MON-19**, penny differences: may a grosz under- or overpaid be written off, and
    is it a tax cost or income? On the nearest question, the Poznań tax chamber
    (2014) and the Minister of Finance (2014) took opposite views.
  - **MON-23**, set-off above 15 000 zł: does netting count as a "payment" under the
    rule that disallows cash-like payments above that amount? Interpretations have
    been reported both ways (secondary sources).
- **This review graded the mechanics, not the classic grey zone.** Whether a given
  expense is a tax cost at all — the general clause in the PIT and CIT acts,
  decided expense by expense — is where much of the interpretation traffic is. It
  is out of reach of a year's walk through the law and within reach of a practising
  accountant and, later, of the network.

## The root causes

The 37 wrong points come from nine causes. Fixing a cause fixes every point under it.
The pack versions are a proposed order, for KK to confirm.

| # | Cause | Wrong points | What fixes it | Proposed |
|---|---|---|---|---|
| 1 | **Dates.** The sales invoice has no sale date and takes its VAT month from the issue date; a purchase takes its VAT month from the invoice date, not from receipt (a KSeF invoice counts as received on the day KSeF numbers it). The same missing sale date decides the PIT and CIT revenue date. | VAT-08, VAT-14, KSEF-13, JPK-04, JPK-10, FB-12 | Sale date (P_6) on the sales invoice; receipt date on the purchase; VAT month by the tax point or by receipt; the same dates feed income tax. | v19 |
| 2 | **The company never says what it is.** No VAT status (an exempt company still gets a VAT settlement and a JPK), no form of books, no form of tax, no cost method or cash method, no monthly-or-quarterly choice. | VAT-03, KSEF-11, KPIR-08, JPK-11 (and VAT-01, FORM-*) | A dated door on the company; rules branch on it. **Kernel step first:** a rule's match conditions read through a link to the company (on the destiny list). | kernel + v19 |
| 3 | **Every bank line is booked as a customer or supplier payment** before anyone says what it is. Owner deposits and withdrawals, share capital, ZUS, taxes, fees and interest all land on 201/202. | MON-07, MON-08, MON-18, ZUS-08, OWN-02, OWN-03, OWN-05, OWN-07 | A statement line is unexplained until classified (invoice, ZUS, tax account, fee, owner, capital); postings follow the classification. For a KPiR trader a bank line is not a book entry at all. | v20 |
| 4 | **Accounts.** Purchased services and expenses stay on 300 for ever, so costs never reach the profit and loss account and profit is overstated. Account 756 mixes penny write-offs with exchange differences. All purchase VAT is deducted. | FB-06, FB-14, MON-19, VAT-09 | Costs to group 4 (services and expenses do not go through 300); 760/761 for settlement differences, 750/751 for exchange differences; non-deductible VAT into the cost. | v20 |
| 5 | **The month lock refuses late documents.** The act requires them booked, in the first open month; only the year closes for good. | FB-16 | The lock forbids posting *into* a closed month; a late document is booked in the first open one. | v20 |
| 6 | **Currency.** Payments post euro amounts into the PLN book; invoices convert at the issue date's rate rather than the NBP rate of the day before; realised exchange differences are never booked; the VAT settlement adds euro VAT to złoty VAT. No NBP rate source exists. | VAT-18, KSEF-10, MON-16, MON-17, FB-13 | An NBP rate channel in `markets/pl`; convert at the day-before rate; convert payments; book realised differences; settle VAT in złoty only. | v21 |
| 7 | **FA(3), both ways.** The writer handles only a Polish-NIP buyer on a plain domestic invoice: one rate per invoice, an invented single line, a fixed split-payment flag, no VAT in złoty for foreign currency, a non-legal exemption basis, no consumers, foreign VAT numbers in the NIP field. The reader ignores the invoice kind, so a supplier's correction is booked as a fresh invoice or refused, never adjusting the original. | VAT-16, VAT-17, VAT-29, KSEF-06, KSEF-07, KSEF-09 | Lines and rates as sold; P_6, P_14_xW, P_18A, P_19A; consumer and foreign buyers; corrections (KOR) and advances (ZAL/ROZ) written, and read back as adjustments to their original. | markets/pl |
| 8 | **JPK markings.** Every invoice without a KSeF number is marked BFK; DI and OFF are never used; buyers without a NIP are not marked "BRAK"; no GTU, TP or document types. | JPK-05, JPK-09 | Markings by the invoice's actual KSeF state; the missing codes; JPK_V7K and corrections alongside. | markets/pl |
| 9 | **Payments from 2027 quote the KSeF number** (VAT art. 108g); matching keys only on the invoice number, and the sales invoice does not carry its KSeF number. | MON-13 | KSeF number on the sales invoice and in the matching policy. | v21 |

## Dated

Points with a date that does not move.

| By | What | Points |
|---|---|---|
| **in force since 1 Feb 2026** | Everyone receives invoices through KSeF; Rhea reads only files dropped into it, with no download from the KSeF API, and a PDF and a KSeF copy of the same invoice could be booked twice. | KSEF-02 |
| **31 Dec 2026** | KSeF tokens end; Rhea authenticates only by token. Certificate authentication before then. | KSEF-21 |
| **1 Jan 2027** | KSeF penalties apply. An invoice reaching KSeF later than its date is offline24 and must be sent by the next working day; Rhea tracks no such deadline. A draft (UD477) would move the penalties to 2028 — not law. | KSEF-15, KSEF-16 |
| **1 Jan 2027** | Transfers between VAT payers quote the KSeF number; the split-payment list (annex 15) is replaced. | MON-13, KSEF-23, VAT-22 |
| **30 Apr 2027** | JPK_PKPIR or JPK_EWP plus JPK_ST for 2026, for sole traders filing JPK_V7. | see the sole-trader section |
| **31 Jul 2027** | JPK_KR_PD for 2026, for a sp. z o.o. filing JPK_V7M. | CIT-08, FB-11, FB-15 |

## What is missing, in blocks

The 147 missing points group into blocks of work, roughly in the order a first user
needs them:

- **The sole trader's records** — KPiR (19 columns from 2026) or the ryczałt record,
  PIT advances, PIT-36/36L/28, ZUS (no object yet), and their annual files.
- **Payment terms** — invoices carry no due date (the KSeF reader reads it; no rule
  keeps it). Late interest, the 40/70/100 EUR compensation, the 90-day bad-debt
  correction and ageing all run from it.
- **Paying suppliers** — the white list check and split payment.
- **Fixed assets** — a book and a tax schedule for each asset, the register, cars.
- **The year end on full books** — a gap-free numbered journal, a trial balance with
  opening balances and turnover for the period and the year to date, CIT advances,
  the financial statement and its filing, the written accounting policy.
- **KSeF beyond the happy path** — offline modes, QR codes on printed invoices,
  corrections, receiving through the API.

## The sections

The five researchers' entries follow unedited. Entries in different sections overlap
where areas meet (the ZUS-and-money section lists its overlaps); the findings agree.


## VAT — sales and purchases

Reviewed against the Poland pack v18 (`packs/pl/pack.json`), the KSeF reader/writer and the JPK writer in `markets/pl/`. Law as of 16 October 2026: the consolidated VAT act is the Marshal's notice of 1 September 2026 (Dz.U. 2026 poz. 1263, legal status 26 August 2026). The act of 4 September 2026 (Dz.U. 2026 poz. 1270) amends it, mostly from 1 January 2027. Abbreviations: "VAT act" = ustawa z 11 marca 2004 r. o podatku od towarów i usług; "OP" = Ordynacja podatkowa (t.j. Dz.U. 2026 poz. 622, amended by Dz.U. 2026 poz. 846 from 1 October 2026); "KSeF manual" = Ministry of Finance, Podręcznik KSeF 2.0 cz. II; "FA(3) brochure" = MF, Broszura informacyjna dotycząca struktury logicznej FA(3), August 2026.

Entries are ordered by when a business meets them in its first year: setting up, first purchases, first sales, payments, month end, later in the year, year end.

---

### VAT-01 — Do we register as an active VAT payer, or stay exempt?
- **Applies to:** everyone
- **The law:** VAT act art. 96 ust. 1: „Podmioty, o których mowa w art. 15 i art. 15a, są obowiązane przed dniem wykonania pierwszej czynności określonej w art. 5 złożyć naczelnikowi urzędu skarbowego zgłoszenie rejestracyjne, z zastrzeżeniem ust. 3”. Under ust. 3 the subjectively exempt (art. 113) and those making only art. 43 exempt sales „mogą złożyć zgłoszenie rejestracyjne”. Art. 113 ust. 4 lets a business give up the exemption only by written notice „przed początkiem miesiąca, w którym rezygnują ze zwolnienia” (or before the first sale for a new business). Art. 113 ust. 13 bars some activities from the exemption altogether: among others „prawnicze”, „w zakresie doradztwa” (with listed exceptions), „jubilerskie” and „ściągania długów”, plus the goods it lists.
- **Grade:** 1 plain
- **Rhea's default:** the company declares its VAT status (active / exempt under art. 113 / exempt under art. 43) with the date it takes effect, and every VAT rule is gated on that status. The status changes how sales, purchases and the month close behave, and nothing in the log can infer it.
- **Pack today:** missing. `company` v4 has no VAT status or registration date, and `pl-settle-vat` produces a VAT settlement (and the JPK writer a JPK_V7M) for every closed month whatever the company's status.
- **Verdict:** missing
- **Deadline:** registration before the first taxable sale (art. 96 ust. 1); leaving the exemption, before the start of the month (art. 113 ust. 4)
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 96, 113)

### VAT-02 — Are we still under the exemption limit, and what happens on the day we cross it?
- **Applies to:** VAT-exempt
- **The law:** VAT act art. 113 ust. 1 (amount changed by Dz.U. 2025 poz. 896, in force 1 January 2026): „Zwalnia się od podatku sprzedaż dokonywaną przez podatnika posiadającego siedzibę działalności gospodarczej na terytorium kraju, u którego wartość sprzedaży, z wyłączeniem podatku, nie przekroczyła w poprzednim ani bieżącym roku podatkowym kwoty 240 000 zł”. The limit was 200 000 zł before 2026. Transitional art. 2 of Dz.U. 2025 poz. 896 lets firms with 2025 sales between 200 000 and 240 000 zł use it. Ust. 9 applies a limit proportional to the period of activity for a business starting mid-year. Ust. 5: the exemption „traci moc począwszy od czynności, którą przekroczono tę kwotę”. Ust. 2 pkt 3 leaves out sales of depreciable fixed assets. Art. 96 ust. 5 pkt 2 requires registration „przed dniem, w którym podatnik traci prawo do zwolnienia”. Enacted change: from 1 January 2027 a new art. 113 ust. 1a also counts art. 13 ust. 3 transfers (own goods moved to another EU state) (Dz.U. 2026 poz. 1270 art. 1 pkt 29). That act's art. 5 measures 2026 sales under the old rules.
- **Grade:** 1 plain. Whether the start-of-year proportion is counted in days or in months: unverified.
- **Rhea's default:** a running total of sales excluding VAT for the calendar year, against 240 000 zł (or the proportional limit in the first year). A warning task at a set margin, and a task to register before the sale that would cross the limit. The limit is plain arithmetic over facts already in the log.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** VAT registration before the day the exemption is lost (art. 96 ust. 5 pkt 2)
- **Sources:** https://eli.gov.pl/eli/DU/2025/896/ogl ; https://eli.gov.pl/eli/DU/2026/1263/ogl ; https://eli.gov.pl/eli/DU/2026/1270/ogl

### VAT-03 — What must an exempt sole trader still do for VAT?
- **Applies to:** VAT-exempt
- **The law:**
  - Daily sales record, VAT act art. 109 ust. 1: such taxpayers „są obowiązani prowadzić ewidencję sprzedaży za dany dzień, niepóźniej jednak niż przed dokonaniem sprzedaży w dniu następnym”.
  - Invoices: not required (art. 106b ust. 2), except on the buyer's request made „w terminie 3 miesięcy, licząc od końca miesiąca” of the sale (art. 106b ust. 3). Any invoice must name the legal basis of the exemption (art. 106e ust. 1 pkt 19).
  - KSeF: art. 106ga ust. 1 obliges „podatnicy”. The exempt are exempted only under art. 113a (EU SME scheme, ust. 2 pkt 6), not under art. 113. Until 31 December 2026 art. 145m allows paper or electronic invoices when the month's invoiced gross is „mniejsza lub równa 10 000 zł”.
  - No JPK_V7: art. 99 ust. 7.
  - Buying services from a foreign business: the exempt buyer still owes the VAT (art. 17 ust. 1 pkt 4). It must register as VAT-UE (art. 97 ust. 3 pkt 1) and file a return „do 25. dnia miesiąca następującego po miesiącu, w którym powstał obowiązek podatkowy” (art. 99 ust. 9). It has no deduction (art. 88 ust. 4).
  - Intra-EU acquisitions of goods are not taxed below 50 000 zł a year (art. 10 ust. 1 pkt 2 lit. c).
- **Grade:** 1 plain
- **Rhea's default:** an exempt company gets no vat_settlement and no JPK. It gets a daily sales record view, invoices on request carrying P_19=1 and P_19A „art. 113 ust. 1 ustawy o VAT”, and an import-of-services rule that raises the VAT-9M obligation and payment.
- **Pack today:** wrong.
  - `pl-settle-vat` runs on every `period.closed`, so an exempt company would get a settlement and a JPK file.
  - The FA(3) writer (`markets/pl/fa3.go`) states the exemption basis as the free text „zwolnienie z VAT” in P_19C. That is not a legal basis: P_19A is the field for the article, per the KSeF manual's section on exempt sales (P_19A/P_19B/P_19C) and art. 106e ust. 1 pkt 19.
  - There is no daily sales record and no VAT-9M.
- **Verdict:** wrong
- **Deadline:** daily record before the next day's sales; an invoice on request if asked within 3 months; VAT-9M and payment by the 25th of the month after the tax point
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 10, 17, 88, 97, 99, 106b, 106e, 106ga, 109, 145m); https://ksef.podatki.gov.pl/media/3zidazbw/podrecznik-ksef-2-0-cz-ii-wystawianie-i-otrzymywane-faktur-w-ksef.pdf

### VAT-04 — Do we account for VAT when paid (cash method) or when we sell?
- **Applies to:** all VAT payers (small taxpayers only)
- **The law:**
  - VAT act art. 21 ust. 1: a small taxpayer may elect that the tax point arises „z dniem otrzymania całości lub części zapłaty” for sales to active VAT payers. For other buyers it is the same, but „niepóźniej niż 180. dnia” from delivery. The election needs notice „w terminie do końca miesiąca poprzedzającego okres, za który będzie stosował tę metodę”. It binds for at least 12 months (ust. 3).
  - Small taxpayer, art. 2 pkt 25 lit. a: sales including VAT not above the equivalent of 2 000 000 euro in the previous year, at the NBP rate „na pierwszy dzień roboczy października poprzedniego roku podatkowego, w zaokrągleniu do 1000 zł”. For 2026 that is 2 000 000 × 4,2586 (NBP table 190/A/NBP/2025 of 1 October 2025) = 8 517 200, so 8 517 000 zł (computed here from the NBP rate and the statutory rounding).
  - Consequences: the invoice must carry „metoda kasowa” (art. 106e ust. 1 pkt 16). The firm files quarterly (art. 99 ust. 2), but not in its first 12 months as an active VAT payer (art. 99 ust. 3a pkt 1). Its own input VAT is deductible only once it pays (art. 86 ust. 10e).
  - Enacted for 1 January 2027: new art. 89a ust. 3a on bad-debt relief for cash-method small taxpayers (Dz.U. 2026 poz. 1270 art. 1 pkt 17).
  - That a business in its first year counts as a small taxpayer: commentary only (secondary — verify).
- **Grade:** 1 plain
- **Rhea's default:** ship accrual only. The company declares its method. A company that declares the cash method gets a refusal and a task until cash-method rules exist, because a silent accrual result would be a wrong return.
- **Pack today:** the `contractor` type has a per-counterparty `cash_method` enum that no rule reads. The cash method is a choice the taxpayer makes, not a property of a counterparty. The FA(3) writer hard-codes P_16 = "2" (no cash method).
- **Verdict:** later (with the refusal guard now)
- **Deadline:** notice by the end of the month before the first cash-method period; leaving it after 12 months, by notice before the end of the quarter
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 2 pkt 25, 21, 86 ust. 10e, 99, 106e); https://api.nbp.pl/api/exchangerates/rates/a/eur/2025-10-01/?format=json ; https://eli.gov.pl/eli/DU/2026/1270/ogl

### VAT-05 — Which VAT rates and rate codes apply, and could 23% change for 2027?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 146ef ust. 1: from 1 January 2024 „do końca roku, w którym suma wydatków” on defence under the Homeland Defence Act „jest wyższa niż 3 % wartości produktu krajowego brutto”, the basic rate „wynosi 23 %” and the reduced rate „wynosi 8 %”. Otherwise art. 41 sets 22% and 7%.
  - Ust. 2: the Minister announces the end of the period in Monitor Polski „w terminie do dnia 31 października roku, dla którego zostały spełnione warunki”.
  - The MF's answer to interpellation 13053 (November 2025) reads it the same way: the rates „powrócą do poziomu … 22% i 7%, gdy wydatki na obronność … nie przekroczą 3%”.
  - I found no announcement of an end date as of 16 October 2026 (unverified that none exists).
  - FA(3) rate codes (P_12), per the FA(3) brochure: „0 KR” (domestic 0%), „0 WDT”, „0 EX”, „zw”, „oo” (domestic reverse charge), „np I” (outside Poland, excluding art. 100 ust. 1 pkt 4 and OSS), „np II” (art. 100 ust. 1 pkt 4 services).
- **Grade:** 1 plain (the 2027 rate depends on a fact the Minister announces)
- **Rhea's default:** vat_rate stays master data with an effective-from and effective-to date, so a rate change is a new dated event. The codes 0 WDT, 0 EX, oo, np I and np II are added beside 23/8/5/0/zw.
- **Pack today:** `pl-register-vat-rate` and five `vat_rate.defined` events (23, 8, 5, 0, zw) with no validity period. The KSeF reader (`ParseFA`) collapses P_13_6_2 (WDT) and P_13_6_3 (export) into "0". No np or oo.
- **Verdict:** missing
- **Deadline:** MF announcement, if any, by 31 October 2026
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 41, 146ef); https://api.sejm.gov.pl/sejm/term10/interpellations/attachment/ATTDNDK93/i13053-o1.pdf ; https://ksef.podatki.gov.pl/media/nuclnu0w/broszura-informacyjna-dotyczaca-struktury-logicznej-fa-3.pdf

### VAT-06 — Do we need a cash register for sales to consumers?
- **Applies to:** everyone selling to private persons
- **The law:**
  - VAT act art. 111 ust. 1: „Podatnicy dokonujący sprzedaży na rzecz osób fizycznych nieprowadzących działalności gospodarczej oraz rolników ryczałtowych są obowiązani prowadzić ewidencję sprzedaży przy zastosowaniu kas rejestrujących”.
  - Ust. 2 sanction: 30% of the input VAT for the period without a register.
  - Exemption by the Regulation of 17 December 2024 (Dz.U. 2024 poz. 1902, amended Dz.U. 2026 poz. 420), § 3 ust. 1 pkt 1–2, „nie dłużej jednak niż do dnia 31 grudnia 2027 r.”: B2C sales in the previous year (or, for a new business, expected and proportional) that do not exceed „kwoty 20 000 zł”. § 5 ust. 1: the exemption lapses „po upływie dwóch miesięcy następujących po miesiącu” in which the limit was crossed.
  - The goods and services excluded from the exemption whatever the turnover (§ 4): not checked here (unverified).
- **Grade:** 1 plain
- **Rhea's default:** a running B2C sales total against 20 000 zł (proportional in the first year), with a task before the two-month grace ends. Rhea does not run a register. It watches the threshold, and a register's daily reports would arrive through a channel.
- **Pack today:** missing. The FA(3) writer refuses a buyer without a NIP, and `pl-book-sales-invoice` resolves the buyer by `vat_id`, so a sale to a private person cannot go through the pack at all.
- **Verdict:** missing
- **Deadline:** register in use by the end of the second month after the month the limit is crossed
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 111); https://eli.gov.pl/eli/DU/2024/1902/ogl ; https://eli.gov.pl/eli/DU/2026/420/ogl

### VAT-07 — Can we deduct VAT on purchases made before VAT registration?
- **Applies to:** all VAT payers (new businesses)
- **The law:**
  - VAT act art. 88 ust. 4: deduction is not available to taxpayers „którzy nie są zarejestrowani jako podatnicy VAT czynni”.
  - Art. 86 ust. 13 allows a late deduction by correcting the return for the period the right arose, within 5 years.
  - CJEU C-385/09 Nidera Handelscompagnie (21 October 2010): the Directive precludes refusing deduction solely because the taxable person was not identified for VAT before using the goods, when it registers within a reasonable time.
  - Polish courts and later tax authority practice allow the deduction after registration (NSA I FSK 1600/07 and interpretations — secondary — verify).
- **Grade:** 2 settled — the text alone reads as a bar, but CJEU C-385/09 Nidera and the administrative courts settled that registration is a formal condition for exercising the right, not for having it
- **Rhea's default:** purchase invoices dated before the registration date are held as "deduction pending registration". Once registered, Rhea drafts the deduction as a correction of the period in which the right arose (art. 86 ust. 13), behind the human gate. Which period to claim is a judgement the accountant confirms.
- **Pack today:** missing. There is no registration date, so the pack cannot tell.
- **Verdict:** missing
- **Deadline:** 5 years from the start of the year the right arose (art. 86 ust. 13)
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 86, 88); https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:62009CJ0385 ; https://ksiegowosc.infor.pl/podatki/interpretacje-podatkowe-orzecznictwo/97530,Odliczenie-VAT-z-faktur-otrzymanych-przed-rejestracja.html (secondary)

### VAT-08 — In which month do we deduct the VAT on a purchase invoice, and when does a KSeF invoice count as received?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 86 ust. 10: the right arises „w rozliczeniu za okres, w którym w odniesieniu do nabytych … towarów i usług powstał obowiązek podatkowy”, i.e. the supplier's tax point.
  - Ust. 10b pkt 1: „niewcześniej niż w rozliczeniu za okres, w którym podatnik otrzymał fakturę”.
  - Ust. 11: if not deducted then, the buyer may deduct „w deklaracji podatkowej za jeden z trzech następnych okresów rozliczeniowych” (two for quarterly filers). Ust. 13: later still, by correction, within 5 years.
  - Receipt: art. 106na ust. 3, „Faktura ustrukturyzowana jest uznana za otrzymaną przy użyciu Krajowego Systemu e-Faktur w dniu przydzielenia w tym systemie numeru identyfikującego tę fakturę”. Ust. 4 covers delivery outside KSeF (actual receipt). Art. 106nda ust. 11 covers offline invoices (also the number date).
  - KSeF manual §5.1, example 10: an invoice sent at 23:56 on 1 February and numbered at 00:01 on 2 February is received on 2 February.
  - From a cash-method seller (P_16 = 1), the seller's tax point is payment (art. 21 ust. 1), so under ust. 10 the buyer deducts no earlier than the period of payment.
- **Grade:** 1 plain
- **Rhea's default:** vat_period of a purchase = the later of (a) the month of the supplier's tax point and (b) the month of receipt. (a) is the FA(3) P_6 date, else P_1, and the payment date for P_16 = 1. (b) is the KSeF number assignment date for KSeF invoices, or the actual receipt date for others. The accountant may defer by up to three periods, recorded as an amendment with a reason.
- **Pack today:** wrong.
  - `pl-book-purchase-invoice` and `pl-book-purchase-invoice-ksef` set `vat_period = month($.invoice_date)`. The rule's own description says „odliczenie w miesiącu otrzymania”, but the formula never sees a receipt date.
  - The KSeF reader (`markets/pl/ksef_inbound.go`) emits no receipt date, no P_6 and no P_16, and sets the event's `occurred_at` to the issue date.
  - A paper or PDF invoice dated 31 January and received on 3 February is deducted in January, which art. 86 ust. 10b forbids.
- **Verdict:** wrong
- **Deadline:** none fixed. The right lasts three following periods in the current return, then 5 years by correction.
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 21, 86, 106na, 106nda); https://ksef.podatki.gov.pl/media/3zidazbw/podrecznik-ksef-2-0-cz-ii-wystawianie-i-otrzymywane-faktur-w-ksef.pdf

### VAT-09 — Which purchase VAT is not deductible (hotels, restaurants, private use, exempt sales, foreign VAT)?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 86 ust. 1: deduction exists only „W zakresie, w jakim towary i usługi są wykorzystywane do wykonywania czynności opodatkowanych”.
  - Art. 88 ust. 1 pkt 4: no deduction for „usług noclegowych i gastronomicznych” (narrow exceptions).
  - Art. 88 ust. 3a: no deduction where the issuer does not exist, the transaction is not taxable or is exempt, or the invoice states activities not performed or wrong amounts.
  - Art. 88 ust. 4: not for those not registered as active.
  - Input VAT is „podatek” under the Polish act, so VAT charged by another member state is not Polish input VAT (art. 86 ust. 2 pkt 1).
  - The JPK regulation (Dz.U. 2019 poz. 1988, annex, pt I.5.1): input VAT is reported „wyłącznie … w takiej wysokości, w jakiej podatnikowi przysługuje prawo do odliczenia”. Its § 5 separates purchases of fixed assets from other purchases.
- **Grade:** 1 plain
- **Rhea's default:** every purchase carries a deductible share (100 / 50 / 0) and a reason. Defaults are 0 for accommodation, catering, private use, purchases for exempt sales and foreign-VAT invoices, and 100 otherwise, with the accountant able to override. The non-deductible VAT is booked to cost, not to 221. Fixed-asset purchases are flagged for their own JPK positions.
- **Pack today:** wrong. `pl-post-purchase-invoice` debits every invoice's VAT to 221. `pl-settle-vat` sums every purchase's VAT into `purchase_vat`/`input_vat`. The JPK writer puts all of it into P_42/P_43 („other purchases”) with no fixed-asset split.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 86, 88); https://eli.gov.pl/eli/DU/2019/1988/ogl

### VAT-10 — How much VAT can we deduct on a car, and when must we file VAT-26?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 86a ust. 1: for expenses on „pojazdy samochodowe” the input tax „stanowi 50 % kwoty podatku”. This covers purchase, lease or rent, fuel, repairs and running costs (ust. 2).
  - 100% only if the vehicle is used exclusively for business (ust. 3 pkt 1 lit. a), proven by usage rules and a mileage log (ust. 4 pkt 1, ust. 7), or by construction.
  - Ust. 12: the VAT-26 information is filed „w terminie do 25. dnia miesiąca następującego po miesiącu, w którym poniosą pierwszy wydatek związany z tymi pojazdami”. If late, exclusive use counts only from the month of filing (ust. 13).
  - Vehicle = ≤ 3.5 t (art. 2 pkt 34).
- **Grade:** 1 plain
- **Rhea's default:** a vehicle declares mixed or exclusive business use. On mixed use, 50% of the VAT on every expense linked to it is deductible and 50% goes to cost. Exclusive use gives 100% only from the month a VAT-26 event is recorded and with a mileage log. A task is raised for the VAT-26 by the 25th.
- **Pack today:** the `vehicle`, `mileage_entry` and `lease_contract` (with `vat_deduction_limit`) types exist, but no rule reads them. Car invoices are deducted at 100% through `pl-post-purchase-invoice`.
- **Verdict:** missing
- **Deadline:** VAT-26 by the 25th of the month after the month of the first vehicle expense
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 2 pkt 34, 86a)

### VAT-11 — Can a till receipt with our NIP, or a simplified invoice, support a deduction?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 106e ust. 5 pkt 3: where „kwota należności ogółem nie przekracza kwoty 450 zł albo kwoty 100 euro” the invoice may omit buyer details and line details if it lets the tax per rate be determined.
  - Art. 106b ust. 5: an invoice for a sale registered on a cash register goes to a VAT taxpayer „wyłącznie, jeżeli paragon … zawiera numer, za pomocą którego nabywca … jest zidentyfikowany”.
  - Art. 145n ust. 1: from 1 February to 31 December 2026, taxpayers obliged to use KSeF may still issue „paragony fiskalne uznane za faktury wystawione zgodnie z art. 106e ust. 5 pkt 3” and register-printed invoices.
  - A receipt without the buyer's NIP is not an invoice, and art. 86 ust. 2 pkt 1 needs an invoice.
- **Grade:** 1 plain
- **Rhea's default:** a purchase document kind "receipt-invoice (≤450 zł, NIP on receipt)" is booked like an invoice outside KSeF, received on the day it is received. A plain receipt books the cost with 0% deductible.
- **Pack today:** missing. `pl-book-purchase-invoice` handles non-KSeF invoices with a number and supplier NIP but has no document kind, so a receipt either fails the ref or is treated as a full invoice.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 106b, 106e, 145n)

### VAT-12 — We buy software, cloud or advertising from a foreign company: who pays the VAT?
- **Applies to:** all VAT payers; VAT-exempt
- **The law:**
  - VAT act art. 28b ust. 1: B2B services are supplied where „podatnik będący usługobiorcą posiada siedzibę działalności gospodarczej”.
  - Art. 17 ust. 1 pkt 4: the Polish buyer is the taxpayer when the provider has no seat or fixed establishment in Poland.
  - Tax point: art. 19a ust. 1 (performance), ust. 3 (periodic billing: end of each period) and ust. 8 (prepayment).
  - FX: art. 31a ust. 1 (NBP rate of the last working day before the tax point).
  - Deduction: art. 86 ust. 2 pkt 4 lit. a, in the same period on condition that the taxpayer „uwzględni kwotę podatku należnego z tytułu tych transakcji w deklaracji podatkowej, w której jest on obowiązany rozliczyć ten podatek” (art. 86 ust. 10b pkt 3).
  - The white-list rules do not apply (OP art. 117ba § 3 pkt 5).
  - Exempt buyers: see VAT-03 (VAT-UE registration, VAT-9M, no deduction).
- **Grade:** 1 plain
- **Rhea's default:** a purchase of services from a supplier with a non-PL VAT number becomes an import-of-services fact.
  - Output VAT is computed at the Polish rate on the net amount, converted at the NBP rate for the last working day before the tax point.
  - An active payer gets an equal input VAT in the same period.
  - An exempt company gets the output only, with a VAT-9M task.
  - Foreign VAT shown on the document is never Polish input VAT.
- **Pack today:** missing. Foreign suppliers' PDFs come in through the `pdf_invoice` channel (`packs/pl/pdf-invoice.json`) as `purchase.invoice.received`. `pl-book-purchase-invoice` and `pl-post-purchase-invoice` then book whatever VAT the document shows (zero, or foreign VAT) as Polish input VAT on 221, and nothing self-assesses output VAT.
- **Verdict:** missing
- **Deadline:** in the return for the period of the tax point (25th of the next month); VAT-9M likewise for the exempt
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 17, 19a, 28b, 31a, 86, 99); https://eli.gov.pl/eli/DU/2026/622/ogl (OP art. 117ba)

### VAT-13 — We buy goods from a supplier in another EU country: intra-EU acquisition
- **Applies to:** all VAT payers; VAT-exempt (above 50 000 zł)
- **The law:**
  - Tax point: VAT act art. 20 ust. 5, „z chwilą wystawienia faktury przez podatnika podatku od wartości dodanej, niepóźniej jednak niż 15. dnia miesiąca następującego po miesiącu, w którym dokonano dostawy”.
  - VAT-UE registration is required „przed dniem dokonania … pierwszego wewnątrzwspólnotowego nabycia” (art. 97 ust. 1).
  - Deduction in the same period if the output is declared in the right return (art. 86 ust. 10b pkt 2).
  - Monthly recapitulative statement (VAT-UE) by the 25th (art. 100 ust. 1 pkt 2, ust. 3).
  - The exempt are outside acquisition taxation up to 50 000 zł a year (art. 10 ust. 1 pkt 2).
- **Grade:** 1 plain
- **Rhea's default:** a purchase of goods from an EU VAT number at 0% becomes an intra-EU acquisition fact, with output and input VAT, a VAT-UE line, and a 50 000 zł counter for exempt firms.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** VAT-UE registration before the first acquisition; VAT-UE statement by the 25th of the following month
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 10, 20, 86, 97, 100)

### VAT-14 — In which month is the VAT on our sale due (the tax point)?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 19a ust. 1: „Obowiązek podatkowy powstaje z chwilą dokonania dostawy towarów lub wykonania usługi”.
  - Ust. 3, periodic services: performed „z upływem każdego okresu, do którego odnoszą się te płatności lub rozliczenia”.
  - Ust. 5 pkt 4: on issue of the invoice for utilities, telecom, „najmu, dzierżawy, leasingu”, security, „stałej obsługi prawnej i biurowej”. Ust. 7: if the invoice is late, on the deadline or the payment due date.
  - Ust. 8: on receipt of an advance.
  - The invoice may be issued up to the 15th of the next month (art. 106i ust. 1) and must show the sale date if it differs (art. 106e ust. 1 pkt 6; FA(3) P_6).
- **Grade:** 1 plain
- **Rhea's default:** `vat_period = month(tax_point)`, where tax_point is:
  - the sale date (P_6) for ordinary goods and services;
  - the end of the billing period for periodic services;
  - the issue date for the art. 19a ust. 5 pkt 4 kinds, capped at the payment due date;
  - the receipt date for advances.
  The issue date is the tax point only when the invoice says no other date.
- **Pack today:** wrong.
  - `pl-book-sales-invoice` sets `vat_period = month($.issue_date)`, and `sales_invoice` v5 has no sale-date field.
  - A September sale invoiced on 10 October lands in October's return, but belongs in September's.
  - The `rejestr-vat-sprzedazy` view also groups by issue month.
- **Verdict:** wrong
- **Deadline:** the return for the month of the tax point (25th of the following month)
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 19a, 106e, 106i); https://ksef.podatki.gov.pl/media/nuclnu0w/broszura-informacyjna-dotyczaca-struktury-logicznej-fa-3.pdf (P_6)

### VAT-15 — By when must the invoice be issued, and which date counts as issue in KSeF?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 106i ust. 1: „Fakturę wystawia się niepóźniej niż 15. dnia miesiąca następującego po miesiącu, w którym dokonano dostawy towaru lub wykonano usługę”. Ust. 2 applies the same to advances. Ust. 3 sets special deadlines (30 days for construction; payment term for utilities, rent and the like). Ust. 7: never „wcześniej niż 60. dnia przed” delivery.
  - Art. 106na ust. 1: „Fakturę ustrukturyzowaną uznaje się za wystawioną w dniu jej przesłania do Krajowego Systemu e-Faktur”.
  - Art. 106nda ust. 16: an invoice sent later than its P_1 date counts as an offline invoice. Its issue date is then P_1 (ust. 10), and it must be sent „niepóźniej niż w następnym dniu roboczym” (ust. 2).
  - KSeF manual §1.2, examples 1–3: same-day sending makes P_1 = the issue date; next-day sending is offline24 with issue date P_1; a future P_1 is rejected.
  - From 1 January 2027 art. 106ni penalises a failure to issue in KSeF or a late offline sending: up to 100% of the VAT shown, or 18.7% of the gross where no VAT is shown.
- **Grade:** 1 plain
- **Rhea's default:** a deadline rule raises a task for any sale without an accepted KSeF submission as the 15th of the following month approaches. Any submission accepted after the next working day after P_1 is flagged as a late offline invoice.
- **Pack today:** partial. `pl-record-ksef-submission` keeps `sent_at`/`accepted_at` and `pl-ksef-rejection-task` turns a refusal into a task. There is no issue-deadline rule, and without a sale date the deadline cannot be computed.
- **Verdict:** missing
- **Deadline:** the 15th of the month after delivery (special cases per art. 106i ust. 3); offline invoices to KSeF by the next working day
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 106i, 106na, 106nda, 106ni); https://ksef.podatki.gov.pl/media/3zidazbw/podrecznik-ksef-2-0-cz-ii-wystawianie-i-otrzymywane-faktur-w-ksef.pdf

### VAT-16 — Does our sales invoice carry everything art. 106e requires?
- **Applies to:** all VAT payers
- **The law:** VAT act art. 106e ust. 1 requires, among others:
  - pkt 3: „imiona i nazwiska lub nazwy podatnika i nabywcy … oraz ich adresy”;
  - pkt 6: the sale date if different;
  - pkt 7: „nazwę (rodzaj) towaru lub usługi”;
  - pkt 8–9: quantity and unit net price;
  - pkt 12–14: the rate, and the net and tax „z podziałem na” rates;
  - pkt 16 „metoda kasowa”, pkt 18 „odwrotne obciążenie”, pkt 18a „mechanizm podzielonej płatności” where applicable;
  - pkt 19: the exemption's legal basis.
- **Grade:** 1 plain
- **Rhea's default:** `sales_invoice` gets lines (name, unit, quantity, unit net price, rate, net), per-rate totals, a sale date and both parties' addresses. The FA(3) annotations are derived from facts (method, reverse charge, split payment) and are never constants.
- **Pack today:** wrong.
  - `sales_invoice` v5 holds one `vat_rate` per invoice, no lines and no sale date.
  - When given no lines, the FA(3) writer invents one: „Sprzedaż wg faktury <numer>”, quantity 1. That fails pkt 7–9.
  - It writes the party's name where the address is missing (fails pkt 3).
  - It hard-codes P_16, P_18 and P_18A to "2", and writes the exemption basis as free text (see VAT-03).
  - A sale at two rates cannot be expressed at all.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 106e); https://ksef.podatki.gov.pl/media/nuclnu0w/broszura-informacyjna-dotyczaca-struktury-logicznej-fa-3.pdf

### VAT-17 — How are VAT amounts rounded: per line, per invoice, in the return?
- **Applies to:** all VAT payers
- **The law:**
  - Invoice: VAT act art. 106e ust. 1 pkt 14 gives the tax „od sumy wartości sprzedaży netto” per rate. Ust. 10 allows tax per line, in which case „łączna kwota podatku może być ustalona w wyniku podsumowania jednostkowych kwot podatku”. Ust. 7 allows computing from gross. Ust. 11: „Kwoty wykazywane w fakturze zaokrągla się do pełnych groszy, przy czym końcówki poniżej 0,5 grosza pomija się, a końcówki od 0,5 grosza zaokrągla się do 1 grosza”.
  - KSeF does not check arithmetic: „KSeF nie dokonuje obliczeń matematycznych” (KSeF manual §1.4.2).
  - Return: OP art. 63 § 1 rounds tax bases and amounts to whole złoty, under 50 gr down and 50 gr or more up. The JPK regulation annex pt I.1.1: „Kwoty z poszczególnych pozycji deklaracji zaokrągla się do pełnych złotych”. Its § 4 pkt 7 defines total output tax as the „sumę wysokości podatku należnego” of the positions.
  - Dz.U. 2026 poz. 846 rewrote OP art. 63 § 3 from 1 October 2026; that concerns crediting payments, not returns.
- **Grade:** 1 plain
- **Rhea's default:** invoice VAT is computed per rate on the sum of net amounts unless the seller declares the per-line method, and Rhea checks each incoming and outgoing invoice's VAT against it. Return: each position is rounded on its own, and totals are sums of rounded positions.
- **Pack today:** wrong.
  - `pl-settle-vat` rounds the grand totals: round(output) − round(input + carry).
  - The JPK writer sets P_38 = whole(output_vat), while P_37 sums rounded bases. Two rates with 0.40 zł each round to 0 + 0 by position but 1 as a total, so P_38 can differ by 1 zł from P_16 + P_18 + P_20.
  - The event's `vat` is never checked against net × rate.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 106e); https://eli.gov.pl/eli/DU/2026/622/ogl (OP art. 63); https://eli.gov.pl/eli/DU/2026/846/ogl ; https://eli.gov.pl/eli/DU/2019/1988/ogl ; https://ksef.podatki.gov.pl/media/3zidazbw/podrecznik-ksef-2-0-cz-ii-wystawianie-i-otrzymywane-faktur-w-ksef.pdf

### VAT-18 — We invoice in euros: which rate converts the VAT, and what must the invoice show?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 31a ust. 1: conversion „według kursu średniego danej waluty obcej ogłoszonego przez Narodowy Bank Polski na ostatni dzień roboczy poprzedzający dzień powstania obowiązku podatkowego”, or the ECB rate if chosen.
  - Ust. 2: an invoice issued before the tax point uses the rate of the working day before issue.
  - Ust. 2a–2c: an optional income-tax method, binding for 12 months.
  - Art. 106e ust. 11: „Kwoty podatku wykazuje się w złotych”.
  - FA(3): the VAT in PLN goes in P_14_xW (KSeF manual, section on invoices in a foreign currency: „jedynie kwota podatku w polu P_14_xW przeliczona jest na polskie złote, zgodnie z art. 106e ust. 11 ustawy”).
- **Grade:** 1 plain
- **Rhea's default:** each foreign-currency invoice stores `vat_pln` per rate, computed with the NBP rate of the last working day before the tax point (keeping the table number as evidence). The settlement and registers sum `vat_pln`. The FA(3) writer fills P_14_xW.
- **Pack today:** wrong.
  - `pl-settle-vat` sums `i.vat` in the invoice's own currency, adding euros to złoty.
  - The FA(3) writer emits no P_14_xW.
  - `pl-post-sales-invoice` and `pl-post-purchase-invoice` convert with the `fx_rate` keyed to the issue date itself (`convert.date = $.issue_date` / `$.invoice_date`), not the last working day before the tax point. The pack does not say the rate feed shifts dates.
  - The VAT registers ignore currency.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 31a, 106e); https://ksef.podatki.gov.pl/media/3zidazbw/podrecznik-ksef-2-0-cz-ii-wystawianie-i-otrzymywane-faktur-w-ksef.pdf

### VAT-19 — A customer pays us before we deliver: is VAT due already?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 19a ust. 8: if payment „w szczególności przedpłatę, zaliczkę” is received before delivery, the tax point arises „z chwilą jej otrzymania w odniesieniu do otrzymanej kwoty”. Exceptions are the art. 19a ust. 5 pkt 4 kinds.
  - Invoice: art. 106b ust. 1 pkt 4; art. 106i ust. 2 by the 15th of the next month.
  - No advance invoice is needed if the delivery happens in the same month (art. 106b ust. 1a).
  - A split payment of an advance is marked „zaliczka” (art. 108a ust. 1c).
- **Grade:** 1 plain
- **Rhea's default:** an incoming payment linked to an order or agreement, with no delivered invoice, becomes an advance fact. Its VAT is gross × rate / (100 + rate) in the month received, with an advance-invoice task by the 15th. The final invoice nets the advance.
- **Pack today:** missing. An unmatched receipt stays an unsettled `bank_operation`, and no VAT arises.
- **Verdict:** missing
- **Deadline:** advance invoice by the 15th of the month after receipt
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 19a, 106b, 106i, 108a)

### VAT-20 — We sell services to a business abroad (common for IT sole traders): what goes on the invoice and the returns?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 28b ust. 1: the place of supply is the buyer's seat, so the service is outside Polish VAT.
  - Invoice: „odwrotne obciążenie” (art. 106e ust. 1 pkt 18). For EU buyers, the seller's number prefixed PL and the buyer's EU VAT number (pkt 24).
  - The invoice is still issued in KSeF (art. 106ga ust. 1) and delivered to the foreign buyer „w sposób z nim uzgodniony” (art. 106gb ust. 4), with the QR code when used outside KSeF (ust. 5).
  - EU B2B services go on the monthly recapitulative statement (VAT-UE), filed „do 25. dnia miesiąca następującego po miesiącu, w którym … powstał obowiązek podatkowy” (art. 100 ust. 1 pkt 4, ust. 3), after VAT-UE registration (art. 97 ust. 3 pkt 2, ust. 10).
  - The register records them under the domestic tax-point rules (art. 109 ust. 3a).
  - FA(3): P_12 „np II”, net in P_13_9 (FA(3) brochure).
- **Grade:** 1 plain
- **Rhea's default:** a buyer with an EU VAT number and a service kind gives rate code np II, reverse-charge annotation, P_13_9, a VAT-UE statement line and the JPK position for supplies outside the country (field number not checked here, unverified). A non-EU buyer gives np I.
- **Pack today:** missing.
  - There is no np code.
  - The FA(3) writer has only a NIP field for the buyer (no EU VAT number or country-ID element), strips a "PL" prefix, and refuses buyers without a NIP.
  - There is no VAT-UE statement.
- **Verdict:** missing
- **Deadline:** VAT-UE by the 25th of the month after the tax point
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 28b, 97, 100, 106e, 106ga, 106gb, 109); https://ksef.podatki.gov.pl/media/nuclnu0w/broszura-informacyjna-dotyczaca-struktury-logicznej-fa-3.pdf

### VAT-21 — In 2026, must a small business issue every invoice through KSeF?
- **Applies to:** all VAT payers; VAT-exempt
- **The law:**
  - VAT act art. 106ga ust. 1 (from 1 February 2026): „Podatnicy są obowiązani wystawiać faktury ustrukturyzowane przy użyciu Krajowego Systemu e-Faktur”. Receiving through KSeF applies from the same date (art. 106gb ust. 1).
  - Art. 145m ust. 1: from 1 April to 31 December 2026, paper or electronic invoices are allowed if the month's invoiced gross „jest mniejsza lub równa 10 000 zł”. The right is lost „począwszy od faktury, którą przekroczono” that value (ust. 2).
  - From 1 January 2027 there is no such relief, and art. 106ni penalties apply.
- **Grade:** 1 plain
- **Rhea's default:** always issue through KSeF. It is allowed now and compulsory from 2027, and it removes the 10 000 zł counting problem.
- **Pack today:** right. The `ksef` channel (activity v2) sends every sales invoice as FA(3).
- **Verdict:** right
- **Deadline:** 1 January 2027 (KSeF compulsory for all; penalties)
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 106ga, 106gb, 106ni, 145l, 145m)

### VAT-22 — When must we use or accept split payment, and how is the 15 000 zł threshold read?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 108a ust. 1a: for goods or services listed in annex 15 „udokumentowane fakturą, w której kwota należności ogółem przekracza kwotę 15 000 zł”, payers „są obowiązani zastosować mechanizm podzielonej płatności”.
  - The seller must mark such invoices „mechanizm podzielonej płatności” (art. 106e ust. 1 pkt 18a). Failing to do so costs an additional liability of 30% of the VAT on annex 15 items (ust. 12).
  - Paying a marked invoice outside split payment loses the cost for PIT (art. 22p ust. 1 pkt 3) and CIT (art. 15d ust. 1 pkt 3).
  - The transfer message must give the invoice number, and for structured invoices „numer identyfikujący tę fakturę w Krajowym Systemie e-Faktur” (art. 108a ust. 3 pkt 3). This wording applies to payments from 1 January 2027 (art. 17 of the act of 16 June 2023, Dz.U. 2023 poz. 1598).
  - Annex 15 is replaced from 1 January 2027 (Dz.U. 2026 poz. 1270 art. 1 pkt 38). Art. 6 of that act keeps the old list for deliveries and invoices straddling the change.
- **Grade:** 2 settled — the MF's tax explanations of 23 December 2019 settled it. The threshold is the invoice's total gross („przekroczenie wartości brutto 15 000 zł przez kwotę należności ogółem na fakturze”), not the contract value. The obligation covers only the VAT on annex-15 lines (their example: 690 zł of VAT on a 3 000 zł net line).
- **Rhea's default:**
  - Lines carry an annex-15 flag (CN/PKWiU). A sales invoice over 15 000 zł gross with any annex-15 line is marked.
  - A matching purchase raises a task to pay that VAT by split payment.
  - Outgoing bank operations record whether split payment was used.
  - From 2027 the payment message carries the KSeF number.
- **Pack today:** missing. The FA(3) writer always writes P_18A = "2", nothing classifies goods by annex 15, and `bank_operation` v2 has no split-payment flag.
- **Verdict:** missing
- **Deadline:** at payment; new annex 15 and message content from 1 January 2027
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 106e, 108a, transitional art. 17 in the notice); https://www.gov.pl/attachment/36857f08-376f-4a11-ac2a-136ee1ca8957 ; https://eli.gov.pl/eli/DU/2026/1270/ogl ; https://eli.gov.pl/eli/DU/2026/592/ogl (PIT art. 22p); https://eli.gov.pl/eli/DU/2026/554/ogl (CIT art. 15d)

### VAT-23 — Before we pay a supplier, must we check its account on the white list?
- **Applies to:** everyone (VAT liability: any art. 15 taxpayer; cost consequence: sole trader KPiR, sp. z o.o. full books — not ryczałt, which has no costs)
- **The law:**
  - VAT act art. 96b ust. 1, 3 pkt 13: the MF list of VAT taxpayers with their accounts.
  - OP art. 117ba § 1: the buyer is jointly liable „całym swoim majątkiem” for the supplier's VAT arrears on that supply when payment for a transaction under Prawo przedsiębiorców art. 19 (a single transaction above „15 000 zł”) went „na rachunek inny niż zawarty na dzień zlecenia przelewu w wykazie”.
  - Exceptions: notice to the tax office „w terminie 7 dni od dnia zlecenia przelewu” (§ 3 pkt 2), split payment (pkt 4), and acquisitions, imports and reverse charge (pkt 5).
  - Cost: PIT art. 22p ust. 1 pkt 2 and CIT art. 15d ust. 1 pkt 2 exclude the cost „w tej części”. If already booked, the taxpayer reduces costs or increases revenue in the month of the transfer (ust. 2).
  - New § 3a (from 1 October 2026, Dz.U. 2026 poz. 846): a power of attorney to sign returns covers the 7-day notice.
  - The MF API returns a „requestId” for every query, and queries are logged (API specification 1.5.0).
- **Grade:** 1 plain
- **Rhea's default:** for an outgoing payment to an active VAT supplier on a transaction above 15 000 zł, query the list (wl-api.mf.gov.pl) on the order date and keep the requestId as evidence. On a mismatch, raise a task: pay by split payment or file the notice within 7 days. The cost rule reverses the cost if neither happened.
- **Pack today:** missing. `pl-supplier-account` takes the IBAN from the invoice and `pl-match-outgoing` matches by it. Nothing checks the list.
- **Verdict:** missing
- **Deadline:** notice within 7 days of ordering the transfer
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 96b); https://eli.gov.pl/eli/DU/2026/622/ogl (OP art. 117ba); https://eli.gov.pl/eli/DU/2026/846/ogl ; https://eli.gov.pl/eli/DU/2025/1480/ogl (Prawo przedsiębiorców art. 19); https://eli.gov.pl/eli/DU/2026/592/ogl ; https://eli.gov.pl/eli/DU/2026/554/ogl ; https://www.gov.pl/attachment/6ce75dd2-97ef-4ef5-b87f-f686c00228de

### VAT-24 — When do we file the VAT return and pay, monthly or quarterly?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 99 ust. 1: monthly returns „w terminie do 25. dnia miesiąca następującego po każdym kolejnym miesiącu”. Art. 103 ust. 1: payment by the same day.
  - Quarterly filing is open to small taxpayers (art. 99 ust. 2–3). It is closed for „12 miesięcy począwszy od miesiąca, w którym dokonana została rejestracja” (ust. 3a pkt 1).
  - The records go with the return as JPK_V7 (art. 109 ust. 3b, art. 99 ust. 11c).
- **Grade:** 1 plain
- **Rhea's default:** monthly for every new active payer. A return-and-payment task is due on the 25th once the month is closed, and the settlement produces the amount.
- **Pack today:** monthly is right (`pl-settle-vat` on `period.closed`). The `tax_declaration` type has `due_date` and `payment_status`, but no rule creates the return or the 25th deadline.
- **Verdict:** missing
- **Deadline:** 25th of the following month (return and payment)
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 99, 103, 109)

### VAT-25 — Input VAT exceeds output VAT: carry it forward or ask for a refund?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 87 ust. 1: the taxpayer may reduce output tax „za następne okresy lub do zwrotu różnicy na rachunek bankowy”.
  - Ust. 2: a refund within 40 days of filing. It was 60 days before the KSeF amendment (Dz.U. 2023 poz. 1598) took effect on 1 February 2026; that act's art. 12 applies the new art. 87 to periods from that date.
  - Ust. 5a: 180 days if no taxable activity in the period. Ust. 6a: 25 days to the VAT account.
- **Grade:** 1 plain
- **Rhea's default:** carry forward. A refund only on an explicit human decision recorded as an event, because a refund claim triggers verification and has its own timing.
- **Pack today:** `pl-settle-vat` always carries forward (`vat_carry`). The JPK writer fills P_53 and P_62, with no refund option.
- **Verdict:** right
- **Deadline:** the claim is made in the return (25th)
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 87, transitional art. 12 in the notice)

### VAT-26 — We issued a KSeF invoice to the wrong NIP, or twice: can we cancel it?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 106na ust. 1 (an invoice is issued when sent); art. 106j (correction invoices).
  - KSeF manual §1.4.3: „W KSeF nie jest możliwe anulowanie wystawionej faktury”. For a wrong buyer NIP the seller „zobowiązany jest do wystawienia faktury korygującej "do zera" z tym błędnym NIP nabywcy”, then a new original invoice to the right buyer, and should not „wystawiać faktury korygującej błędny numer NIP na inny prawidłowy”.
  - Correction notes (noty korygujące) are abolished from 1 February 2026 (manual §1.4.2).
- **Grade:** 2 settled — the act only says corrections are by correction invoice. The Ministry's KSeF manual settled the procedure for a wrong buyer or a duplicate.
- **Rhea's default:** an accepted wrong invoice is corrected to zero by a KSeF correction referencing its KSeF number, and a new invoice is issued if needed. Both go through the human gate.
- **Pack today:** missing. Rejections are handled (`pl-record-ksef-rejection`, `pl-ksef-rejection-task`). An accepted but wrong invoice has no path. CORPUS still lists a `correction_note` candidate, which no longer exists in law.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 106j, 106na); https://ksef.podatki.gov.pl/media/3zidazbw/podrecznik-ksef-2-0-cz-ii-wystawianie-i-otrzymywane-faktur-w-ksef.pdf

### VAT-27 — We gave a discount or took goods back: when does our output VAT go down?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 29a ust. 10 pkt 1–3 (discounts, returns, refunded advances).
  - Ust. 13 (since 1 February 2026): the reduction „dokonuje się za okres rozliczeniowy, w którym podatnik wystawił fakturę korygującą w postaci faktury ustrukturyzowanej”.
  - Ust. 13a: for a correction outside KSeF, the period „w którym podatnik otrzymał potwierdzenie otrzymania faktury korygującej przez nabywcę”. Ust. 13b–13c cover offline corrections (the period sent to KSeF). Ust. 15 lists exceptions (exports, intra-EU supplies, utilities; ust. 15 pkt 6 documented delivery attempt, limited by ust. 15c).
  - What KSeF changed: the wording until 31 January 2026 (Dz.U. 2025 poz. 775) required that „z posiadanej przez tego podatnika dokumentacji wynika, że uzgodnił on z nabywcą … warunki obniżenia podstawy opodatkowania … oraz warunki te zostały spełnione”. That condition is gone. A structured correction counts in its issue month.
  - Art. 11 of the 2023 act keeps the old wording for corrections issued before 1 February 2026.
  - Ust. 14 applies the same to a VAT amount stated too high.
- **Grade:** 1 plain
- **Rhea's default:** a correction is its own KSeF document referencing the original's KSeF number. Its effect lands in the month of its KSeF issue date. For corrections outside KSeF (foreign buyers outside KSeF), it lands in the month of the buyer's confirmation. Rhea keeps the agreement as an attached fact for income-tax and audit purposes, not as a VAT condition.
- **Pack today:** missing. No correction document type and no rule. `sales_invoice` is amendable only in `open`.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 29a; transitional art. 11 in the notice); https://eli.gov.pl/eli/DU/2025/775/ogl (wording before 1 February 2026)

### VAT-28 — We undercharged and must invoice more: which month takes the increase?
- **Applies to:** all VAT payers
- **The law:** VAT act art. 29a ust. 17: „W przypadku gdy podstawa opodatkowania uległa zwiększeniu, korekty tej podstawy dokonuje się w rozliczeniu za okres rozliczeniowy, w którym zaistniała przyczyna zwiększenia podstawy opodatkowania”. Ust. 18 covers exports and intra-EU supplies.
- **Grade:** 2 settled — „przyczyna” is not defined. Tax authority practice since the 2021 (SLIM VAT) wording has settled it:
  - an error present when the invoice was issued (wrong price, rate or quantity) goes back to the original invoice's period;
  - a cause arising after the sale (a price rise agreed later) goes in the month it arose.
  
  For example, individual interpretation 0112-KDIL1-3.4012.301.2021.1.JK of 6 September 2021 (secondary — verify).
- **Rhea's default:** every increasing correction states its cause. "Error at issue" re-opens the original period and needs a corrected return for that month, with interest exposure. "Later event" lands in the month of the event.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** corrected return for the original period where the cause was an error
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 29a); https://ksiegowosc.infor.pl/podatki/vat/faktura/5180973,Faktury-korygujace-in-plus-zmiany-od-2021-roku.html (secondary); https://www.pit.pl/?p=5025 (secondary)

### VAT-29 — A supplier corrects its invoice: when does our input VAT change?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 86 ust. 19a: for a decrease under art. 29a ust. 13–13c the buyer „jest obowiązany do zmniejszenia kwoty podatku naliczonego w rozliczeniu za okres, w którym otrzymał fakturę korygującą”.
  - For a KSeF correction, received is the day its KSeF number is assigned (art. 106na ust. 3).
  - An increase is deducted like any invoice (art. 86 ust. 10, 10b pkt 1).
- **Grade:** 1 plain
- **Rhea's default:** an incoming correction (FA(3) RodzajFaktury = KOR) links to the original by KSeF number and adjusts input VAT in the month of its KSeF number date.
- **Pack today:** wrong. The KSeF reader (`ParseFA`) does not read RodzajFaktury. A correction (KOR) or an advance invoice (ZAL) is either booked through `pl-book-purchase-invoice-ksef` as a fresh purchase invoice (own open amount, `vat_period = month(invoice_date)`), or, when its before/after lines do not add up to the net, refused as unreadable. In neither case does it adjust the original invoice.
- **Verdict:** wrong
- **Deadline:** the return for the month of receipt
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 86, 106na)

### VAT-30 — A customer has not paid for 90 days past due: can we reduce our output VAT?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 89a ust. 1a: uncollectibility is presumed when the debt „nie została uregulowana lub zbyta w jakiejkolwiek formie w ciągu 90 dni od dnia upływu terminu jej płatności”.
  - Conditions: the creditor is an active VAT payer on the day before filing (ust. 2 pkt 3 lit. a), and „od daty wystawienia faktury … nie upłynęły 3 lata, licząc od końca roku, w którym została wystawiona” (ust. 2 pkt 5). For non-business debtors the extra conditions of ust. 2a apply.
  - Correction in the period the 90 days lapse, if unpaid by the filing day (ust. 3). Output VAT is re-increased on later payment (ust. 4). The relief is reported in the return (ust. 5).
  - From 1 January 2027: ust. 3a for cash-method small taxpayers.
- **Grade:** 1 plain
- **Rhea's default:** when a sales invoice is unpaid 90 days after its due date, Rhea proposes the relief as a draft for the human gate (it is an option, not a duty). On later payment the re-increase is automatic.
- **Pack today:** missing. `sales_invoice` has no payment due date: the FA(3) writer's term is never stored on the object. The `naleznosci-do-pilnowania` queue ranks by issue date.
- **Verdict:** missing
- **Deadline:** from the period in which the 90th day after the due date falls; within 3 years from the end of the year of issue
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 89a); https://eli.gov.pl/eli/DU/2026/1270/ogl

### VAT-31 — We have not paid a supplier for 90 days past due: must we give back the input VAT?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 89b ust. 1: if the invoice is not paid „w terminie 90 dni od dnia upływu terminu jej płatności”, the debtor „jest obowiązany do korekty odliczonej kwoty podatku … w rozliczeniu za okres, w którym upłynął 90 dzień”.
  - No correction if paid by the last day of that period (ust. 1a). Partial payment leaves a proportional correction (ust. 2). On later payment the VAT is deducted again (ust. 4).
- **Grade:** 1 plain
- **Rhea's default:** a mandatory rule. Every purchase invoice carries its due date. In the period where the 90th day after it falls, an unpaid remainder at period end reduces input VAT by its share. A matched payment later restores it.
- **Pack today:** missing. The KSeF reader puts `payment_term` in the event, but `pl-book-purchase-invoice(-ksef)` does not map it, and `purchase_invoice` v4 has no due-date field.
- **Verdict:** missing
- **Deadline:** the return for the period in which the 90th day after the due date falls
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 89b)

### VAT-32 — We take goods for private use or give gifts to clients: is that a taxable supply?
- **Applies to:** all VAT payers
- **The law:**
  - VAT act art. 7 ust. 2: a supply includes „przekazanie lub zużycie towarów na cele osobiste podatnika” and „wszelkie inne darowizny”, if input VAT was deductible.
  - Art. 7 ust. 3–4: gifts of small value are excluded. That means up to 100 zł net per person per year with a register of recipients, or a unit cost up to 20 zł without one.
  - Art. 8 ust. 2 does the same for services, such as private use of business assets.
- **Grade:** 1 plain
- **Rhea's default:** a door for "taken for private use / given away". Output VAT on the purchase price at the time is computed unless the small-gift test passes, with a gift register per recipient.
- **Pack today:** missing
- **Verdict:** later
- **Deadline:** in the return for the month of the transfer
- **Sources:** https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 7, 8)

### VAT-33 — What changes on 1 January 2027 that the pack must already plan for?
- **Applies to:** everyone
- **The law:**
  - (a) KSeF compulsory for every taxpayer: the art. 145m relief ends on 31 December 2026. KSeF penalties apply from 1 January 2027 (art. 106ni).
  - (b) The split-payment message carries the KSeF number for payments from 1 January 2027 (art. 17 of the 2023 act, Dz.U. 2023 poz. 1598).
  - (c) A new annex 15 (split-payment list) and new annexes 16–17 (Dz.U. 2026 poz. 1270 art. 1 pkt 38–39; in force 1 January 2027 per its art. 9). Art. 6 keeps the old list for deliveries and invoices straddling the date.
  - (d) art. 113 ust. 1a counts art. 13 ust. 3 transfers toward the exemption limit, with 2026 measured under the old rules (art. 5 of that act).
  - (e) art. 89a ust. 3a: bad-debt relief for cash-method small taxpayers.
  - (f) Whether 23%/8% continue depends on the art. 146ef announcement due by 31 October 2026 (see VAT-05).
- **Grade:** 1 plain
- **Rhea's default:** carry each change as a dated rule version (`effective_from: 2027-01-01`) drafted now, behind the gate, so 2026 replays unchanged.
- **Pack today:** KSeF-out already in place, which is right for (a). (b)–(e) are missing (no split-payment or exemption-limit rules exist yet to version).
- **Verdict:** later
- **Deadline:** 1 January 2027
- **Sources:** https://eli.gov.pl/eli/DU/2026/1270/ogl ; https://eli.gov.pl/eli/DU/2026/1263/ogl (art. 106ni, 145m, transitional art. 17 in the notice)

## KSeF and JPK

Law as of 16 October 2026. The base texts used are the VAT act in its consolidated text of 1 September 2026 (Dz.U. 2026 poz. 1263, law as of 26 August 2026), the amending acts it consolidates (Dz.U. 2023 poz. 1598, 2024 poz. 852 and 1721, 2025 poz. 1203), the December 2025 regulations (Dz.U. 2025 poz. 1740, 1742, 1800, 1815), the Ministry's KSeF 2.0 technical documentation (github.com/CIRFMF/ksef-docs), the JPK_V7(3) brochure (January 2026) and the Ministry's marking guide. Two corrections to secondary sources: in the Dziennik Ustaw, poz. **1740** is the regulation on cases exempt from KSeF and poz. **1742** amends the invoicing regulation. ksef.podatki.gov.pl lists them the other way round. Drafts that are not law are marked as drafts.

Entries are in the order a business meets them during the year: setup, first sale, first purchase, exceptions, first month-end, after filing, the turn of 2026/2027, and annual filings.

---

### KSEF-01 — From when must we issue invoices in KSeF, and may we still issue some outside it in 2026?
- **Applies to:** everyone (VAT payers and VAT-exempt taxpayers alike)
- **The law:** VAT act art. 106ga ust. 1: „Podatnicy są obowiązani wystawiać faktury ustrukturyzowane przy użyciu Krajowego Systemu e-Faktur.” This has been in force since 1 Feb 2026 under art. 23 of Dz.U. 2023 poz. 1598, as amended by Dz.U. 2024 poz. 852 („z dniem 1 lutego 2026 r.”). The 2025 act (Dz.U. 2025 poz. 1203) rewrote the transitional articles as follows:
  - Art. 145l: until 31 Mar 2026, taxpayers may stay outside KSeF „jeżeli łączna wartość sprzedaży wraz z kwotą podatku u tych podatników nie przekroczyła w 2024 r. kwoty 200 000 000 zł”.
  - Art. 145m ust. 1: from 1 Apr to 31 Dec 2026 they may issue paper or e-invoices „jeżeli łączna wartość sprzedaży wraz z kwotą podatku u tych podatników udokumentowana tymi fakturami wystawionymi w danym miesiącu jest mniejsza lub równa 10 000 zł”. Under ust. 2 the right is lost „począwszy od faktury, którą przekroczono wartość”.
  - Art. 145n: until 31 Dec 2026, invoices from cash registers and receipts with a NIP that count as invoices are allowed.

  Who is excluded (art. 106ga ust. 2):
  - issuers with no establishment in Poland (pkt 1–2)
  - OSS/IOSS (pkt 3)
  - consumers as buyers (pkt 4)
  - the cases in Dz.U. 2025 poz. 1740 (pkt 5): motorway tolls, single tickets, CRCO, some exempt financial services, self-billing by buyers without a NIP
  - the EU SME scheme, art. 113a (pkt 6)

  VAT-exempt taxpayers are included. The Ministry's FAQ for sole traders says „każdy podatnik (czynny i zwolniony)”. History: the 2023 act set 1 Jul 2024. The 2024 act moved it to 1 Feb 2026. The 2025 act split it into 1 Feb and 1 Apr 2026 and added the 10,000 zł allowance.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Send every invoice to a Polish NIP buyer through KSeF from the company's first day, and do not use the 10,000 zł/month allowance. The allowance is optional, tracking it invoice by invoice costs more than it saves, and it expires on 31 Dec 2026.
- **Pack today:** markets/pl/ksef.go (`KSeF.Pass` sends every `sales_invoice` without an accepted `ksef_submission`), activity `ksef`, rules `pl-record-ksef-submission` and `pl-record-ksef-rejection`. The company needs no "obligated from" date under this default.
- **Verdict:** right
- **Deadline:** 1 Feb 2026 (sales above 200 mln zł in 2024); 1 Apr 2026 (everyone else, including VAT-exempt). The allowances end 31 Dec 2026.
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2025/1203/D2025000120301.pdf ; https://dziennikustaw.gov.pl/DU/2024/852/D2024000085201.pdf ; https://dziennikustaw.gov.pl/DU/2025/1740/D2025000174001.pdf ; https://ksef.podatki.gov.pl/jdg-i-msp/ ; https://ksef.podatki.gov.pl/informacje-ogolne-ksef-20/podstawy-prawne-oraz-kluczowe-terminy/

### KSEF-02 — Must we receive our purchase invoices through KSeF, and how do they reach the books?
- **Applies to:** everyone (any buyer with a NIP, including VAT-exempt)
- **The law:** VAT act art. 106gb ust. 1: „Faktura ustrukturyzowana jest wystawiana i otrzymywana przy użyciu Krajowego Systemu e-Faktur”. The buyer's acceptance requirement (art. 106na ust. 2) was repealed by Dz.U. 2023 poz. 1598 and now reads „(uchylony)”. Ministry FAQ: „Obowiązek otrzymywania faktur w KSeF będzie istniał od 1.02.2026 r.” Invoices go outside KSeF only for the buyers in art. 106gb ust. 4: foreign buyers, the 113a scheme, buyers without an identifier, and consumers. Retrieval is through the API (`POST /invoices/query/metadata`, `POST /invoices/exports` in the Ministry's KSeF 2.0 docs).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Pull purchase invoices from the KSeF API by the company's NIP (subject 2) at least daily. Treat KSeF as the only source of a domestic purchase invoice. When a PDF copy arrives by e-mail, match it to its KSeF twin by seller NIP and number, or by the KSeF number printed under its QR code. Never book both.
- **Pack today:** packs/pl/ksef-inbound.json with markets/pl/ksef_inbound.go only reads FA files that someone has put into the document store (the channel declares „pobieranie z API KSeF dołączy z klientem”). The KSeF number comes only from the file name. markets/pl/ksef2.go has no download calls. packs/pl/pdf-invoice.json books PDFs separately, so the same invoice can be booked twice.
- **Verdict:** missing
- **Deadline:** 1 Feb 2026
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2023/1598/D2023000159801.pdf ; https://ksef.podatki.gov.pl/jdg-i-msp/ ; https://github.com/CIRFMF/ksef-docs/blob/main/pobieranie-faktur/pobieranie-faktur.md

### KSEF-03 — Who may log in to KSeF and send invoices for the company: the owner, an employee, the accountant?
- **Applies to:** sole trader KPiR | sole trader ryczałt | sp. z o.o. full books
- **The law:** VAT act art. 106nb lists the users. The KSeF regulation (Dz.U. 2025 poz. 1815) sets out the methods and permissions:
  - **Login methods (§ 6 ust. 1):** the national login node (trusted profile and similar, in force from 1 Apr 2026 per § 15 pkt 2), a qualified signature, a qualified seal, or a KSeF certificate. Under § 13 ust. 1 pkt 2 the old „podpis zaufany” route was allowed only until 31 Mar 2026.
  - **Companies without a seal (§ 5 ust. 1):** a company that cannot use a qualified seal files ZAW-FA „do właściwego naczelnika urzędu skarbowego”.
  - **One person only (§ 5 ust. 3):** under ZAW-FA „uprawnienia te może posiadać jedna osoba fizyczna”. That person grants further permissions electronically.
  - **Third parties (§ 3 ust. 3 pkt 1):** a „podmiot wskazany przez podatnika”, such as an accounting office, may receive the right to issue or read invoices.
  - **Owner (Ministry docs, uprawnienia.md):** the owner is recognised automatically by NIP or by the linked PESEL.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** A sole trader logs in with their own trusted profile or qualified signature and grants Rhea (or the accountant) only InvoiceWrite and InvoiceRead. An sp. z o.o. uses its qualified seal; without one, a single ZAW-FA person grants permissions. Keep who holds which permission as company master data, so the books can say who sent each invoice.
- **Pack today:** missing. No object records KSeF permissions. markets/pl/ksef2.go takes NIP and token from `RHEA_KSEF_NIP` and `RHEA_KSEF_TOKEN`; keeping the secret outside the log is right.
- **Verdict:** missing
- **Deadline:** before the first KSeF invoice (1 Apr 2026 for most)
- **Sources:** https://dziennikustaw.gov.pl/DU/2025/1815/D2025000181501.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/uprawnienia.md ; https://github.com/CIRFMF/ksef-docs/blob/main/uwierzytelnianie.md

### JPK-01 — Monthly or quarterly VAT return: JPK_V7M or JPK_V7K?
- **Applies to:** all VAT payers
- **The law:**
  - **Monthly by default:** VAT act art. 99 ust. 1 requires a return „do 25. dnia miesiąca następującego po każdym kolejnym miesiącu”.
  - **Who may go quarterly:** under ust. 2, small taxpayers on the cash method file quarterly. Under ust. 3, other small taxpayers may file quarterly „po uprzednim zawiadomieniu na piśmie naczelnika urzędu skarbowego, najpóźniej do 25. dnia drugiego miesiąca kwartału”. The VAT-R form has the box for this (field 54, „podatnik wybiera (wybrał) możliwość rozliczania się za okresy kwartalne”; Dz.U. 2026 poz. 1152).
  - **Who may not (ust. 3a):** taxpayers in their first 12 months after VAT registration, and those supplying Annex 15 goods above 50,000 zł in any month of the quarter or the 4 quarters before.
  - **Small taxpayer (art. 2 pkt 25 lit. a):** sales that „nie przekroczyła w poprzednim roku podatkowym … równowartości 2 000 000 euro”, converted at the NBP rate „na pierwszy dzień roboczy października poprzedniego roku podatkowego, w zaokrągleniu do 1000 zł”. My own arithmetic from NBP tables (no Ministry figure checked):
    - 2026: 4.2586 (table 190/A/NBP/2025) → 8,517,000 zł
    - 2027: 4.3770 (table 191/A/NBP/2026) → 8,754,000 zł
  - **How quarterly works:** art. 109 ust. 3c sends the records „za pierwszy i drugi miesiąc danego kwartału – w terminie do 25. dnia miesiąca następującego”, and the third month together with the quarter's declaration. The brochure says the same: months 1–2 carry Naglowek, Podmiot1 and Ewidencja only; month 3 adds a Deklaracja covering the whole quarter.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** File monthly (V7M) unless the company declares the quarterly choice. Refuse quarterly in the first 12 months after VAT registration. On quarterly, months 1–2 are records-only files and month 3 carries records plus the quarter's declaration.
- **Pack today:** markets/pl/jpk.go writes only JPK_V7M(3). `company` has no settlement-period field. `pl-settle-vat` settles one month at a time.
- **Verdict:** missing
- **Deadline:** notify by the 25th of the 2nd month of the first quarterly quarter
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2026/1152/D2026000115201.pdf ; https://www.podatki.gov.pl/media/wgbkrejs/broszura-jpk_vat-z-deklaracj%C4%85-od-1-lutego-2026-r.pdf ; https://api.nbp.pl/api/exchangerates/rates/a/eur/2025-10-01/ ; https://api.nbp.pl/api/exchangerates/rates/a/eur/2026-10-01/

### KSEF-04 — Which date is the invoice's issue date when KSeF numbers it later?
- **Applies to:** everyone
- **The law:** Two articles meet here:
  - **Same-day sending:** art. 106na ust. 1: „Fakturę ustrukturyzowaną uznaje się za wystawioną w dniu jej przesłania do Krajowego Systemu e-Faktur.”
  - **Later sending:** art. 106nda ust. 16 (added by Dz.U. 2025 poz. 1203): „Za fakturę, o której mowa w ust. 1, uznaje się także fakturę ustrukturyzowaną, jeżeli data jej przesłania do Krajowego Systemu e-Faktur jest późniejsza niż data, o której mowa w art. 106e ust. 1 pkt 1, wskazana na tej fakturze”. Under ust. 10 such an invoice's issue date is „datę, o której mowa w art. 106e ust. 1 pkt 1, wskazaną przez podatnika na tej fakturze”.
  - **Ministry docs:**
    - „Za datę wystawienia system KSeF przyjmuje wyłącznie wartość zawartą w polu P_1”. KSeF also marks an invoice offline on its own when the P_1 day is earlier than the day of sending.
    - The date inside a KSeF number is „data przyjęcia faktury … do dalszego przetwarzania”, not P_1.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** P_1 is the issue date everywhere: records, JPK `DataWystawienia`, views. `sent_at` and `accepted_at` are evidence only. Sending on a later day than P_1 makes the invoice an offline24 invoice (KSEF-15).
- **Pack today:**
  - jpk.go uses `sales_invoice.issue_date`, which is right.
  - The description of `pl-record-ksef-submission` and the comment in ksef.go call `sent_at` „dzień wystawienia faktury ustrukturyzowanej”. That is true only when the invoice is sent the same day.
  - `KSeFReference` in fa3.go (the fake) builds the number's date from P_1, which the real system does not.
- **Verdict:** right (fix the description and the date in the fake number)
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/tryby-offline.md ; https://github.com/CIRFMF/ksef-docs/blob/main/offline/automatyczne-okreslanie-trybu-offline.md ; https://github.com/CIRFMF/ksef-docs/blob/main/faktury/numer-ksef.md

### KSEF-05 — We send the customer a PDF or a printout: which QR codes must it carry?
- **Applies to:** everyone (anyone who uses a KSeF invoice outside KSeF)
- **The law:**
  - **When a code is required:** VAT act art. 106gb ust. 5: „Podatnik jest obowiązany do oznaczenia faktury ustrukturyzowanej kodem umożliwiającym dostęp do tej faktury w Krajowym Systemie e-Faktur oraz umożliwiającym weryfikację danych zawartych na tej fakturze, w przypadku: 1) udostępnienia jej nabywcy, o którym mowa w ust. 4, w sposób inny niż przy użyciu Krajowego Systemu e-Faktur, lub 2) użycia tej faktury poza Krajowym Systemem e-Faktur.” Under ust. 5a the duty also binds a „podmiot inny niż podatnik”.
  - **Offline invoices:** art. 106nda ust. 6 requires two codes. Under ust. 7 the taxpayer must download „certyfikat umożliwiający potwierdzenie tożsamości wystawcy”.
  - **Regulation Dz.U. 2025 poz. 1815, § 9:**
    - **First code (ust. 1):** a verification code with the API resource address, the P_1 date, the seller's NIP and the „wyróżnik faktury” (the SHA-256-based fingerprint defined in § 2 pkt 6). Beneath it goes the KSeF number, „zamieszczanym bezpośrednio pod kodem weryfikującym”.
    - **Before an offline invoice is sent (ust. 3):** the same code, labelled „napisem „OFFLINE””.
    - **Format (ust. 9):** QR „zgodnego z normą ISO/IEC 18004:2024”.
    - **Second code on offline invoices (ust. 10):** context type and value, seller NIP, the KSeF certificate identifier and the fingerprint, signed with the KSeF certificate, labelled „CERTYFIKAT”.
    - **Structured delivery (ust. 8):** if the invoice is sent as structured data, a separate link or graphic goes with it.
  - **Ministry spec (kody-qr.md):**
    - **KOD I:** `https://qr.ksef.mf.gov.pl/invoice/{NIP}/{DD-MM-RRRR}/{SHA-256 of the file, Base64URL}`.
    - **KOD II:** `https://qr.ksef.mf.gov.pl/certificate/{context type}/{context value}/{seller NIP}/{certificate serial}/{hash}/{signature}`. The signature covers the URL without `https://`, using RSASSA-PSS (SHA-256, MGF1, 32-byte salt, key at least 2048 bits) or ECDSA P-256.
    - **Certificate:** KOD II needs a certificate of type Offline: „Certyfikat typu `Authentication` nie może być używany do generowania KODU II”.
    - **Online vs offline:** an online invoice carries one code; an offline invoice carries two.
  - **Certificates:** downloadable from 1 Nov 2025 (art. 17b of the 2023 act as amended), valid at most 2 years (§ 6 ust. 2). None can be issued in a session opened with a token (§ 13 ust. 2).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Every PDF, print or e-mail copy of a sales invoice carries KOD I, built from the exact XML sent:
  - production host `qr.ksef.mf.gov.pl`
  - P_1 as DD-MM-RRRR
  - SHA-256 in Base64URL
  - the KSeF number underneath, or "OFFLINE" until the number exists

  Offline copies also carry KOD II, signed with an Offline-type KSeF certificate and labelled "CERTYFIKAT".
- **Pack today:** missing. Rhea renders no invoice. `ksef_submission.invoice_hash` is standard Base64, not Base64URL, but the exact XML is kept, so the code can be derived. DECISIONS.md lists the QR codes as owed.
- **Verdict:** missing
- **Deadline:** from the first invoice handed over outside KSeF
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2025/1815/D2025000181501.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/kody-qr.md ; https://github.com/CIRFMF/ksef-docs/blob/main/certyfikaty-KSeF.md

### KSEF-06 — Invoice for a consumer: through KSeF or not?
- **Applies to:** all VAT payers | VAT-exempt (selling to consumers)
- **The law:**
  - **Not obligatory:** art. 106ga ust. 2 pkt 4 lifts the obligation for invoices „na rzecz nabywcy towarów lub usług będącego osobą fizyczną nieprowadzącą działalności gospodarczej”.
  - **But allowed:** ust. 4: „W przypadku, o którym mowa w ust. 2 pkt 1, 2 i 4, podatnicy mogą wystawiać faktury ustrukturyzowane.”
  - **If issued in KSeF:** the invoice is handed over „w sposób z nim uzgodniony” (art. 106gb ust. 4 pkt 6), with a QR code (ust. 5). It counts as received on the day of actual receipt (art. 106na ust. 4).
  - **FA(3):** the buyer is identified by `BrakID`.
  - **In the JPK:** a consumer invoice issued outside KSeF is marked BFK (Ministry marking guide), and NrKontrahenta is „BRAK”.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Issue consumer invoices in KSeF too, as ust. 4 allows. That keeps one path, KSeF stores them for 10 years, and the JPK row carries a KSeF number. Hand the consumer a PDF with KOD I. Issuing outside KSeF with BFK stays lawful; record which route the company chose.
- **Pack today:**
  - fa3.go `WriteFA3` refuses with "the buyer needs a NIP".
  - `KSeF.Pass` reports this as a problem on every pass.
  - jpk.go then marks the invoice BFK with an empty NrKontrahenta.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://ksef.podatki.gov.pl/jdg-i-msp/ ; https://www.podatki.gov.pl/media/lzafohup/jpk_vat-z-deklaracja-stosowanie-znacznikow.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/faktury/schemy/FA/schemat_FA(3)_v1-0E.xsd

### KSEF-07 — Invoice for a foreign buyer: KSeF, and how does the buyer get it?
- **Applies to:** all VAT payers
- **The law:**
  - **Obligatory:** the exclusions in art. 106ga ust. 2 cover foreign issuers, not foreign buyers, so a Polish taxpayer's B2B invoice to a foreign buyer is issued in KSeF.
  - **Delivery:** art. 106gb ust. 4 pkt 1–3: „– faktura ustrukturyzowana jest udostępniana nabywcy w sposób z nim uzgodniony”, with the QR code (ust. 5). It counts as received on the day of actual receipt (art. 106na ust. 4).
  - **Ministry FAQ:** for „fakturach dla konsumentów albo dla zagranicznych odbiorców” the parties agree how to deliver, for example by PDF.
  - **FA(3), `Podmiot2`:** one of `NIP` | `KodUE`+`NrVatUE` | `KodKraju`+`NrID` | `BrakID`. The zero rate is split into P_13_6_1 (domestic), P_13_6_2 (intra-EU supply) and P_13_6_3 (export).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Send it to KSeF and deliver a PDF with KOD I by e-mail. Identify the buyer by `KodUE`+`NrVatUE` inside the EU, or `KodKraju`+`NrID` outside it. Choose the zero-rate slot by the kind of supply.
- **Pack today:**
  - `faIdent` in fa3.go has only `NIP`.
  - `NIPDigits` strips only "PL", so "DE…" lands in `<NIP>` and KSeF will reject it.
  - `faRateSlot` maps rate "0" to 6_1 only.
  - Nothing delivers the invoice.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://ksef.podatki.gov.pl/jdg-i-msp/ ; https://github.com/CIRFMF/ksef-docs/blob/main/faktury/schemy/FA/schemat_FA(3)_v1-0E.xsd

### KSEF-08 — The goods were delivered (or the service finished) on another day than the invoice date: where does that date go?
- **Applies to:** all VAT payers | VAT-exempt
- **The law:**
  - **On the invoice:** art. 106e ust. 1 pkt 6 requires „datę dokonania lub zakończenia dostawy towarów lub wykonania usługi lub datę otrzymania zapłaty, o której mowa w art. 106b ust. 1 pkt 4, o ile taka data jest określona i różni się od daty wystawienia faktury”. FA(3) puts it in P_6 (one date for the whole invoice) or P_6A (per line).
  - **In the JPK:** the field is DataSprzedazy (regulation § 10 ust. 2 pkt 1 lit. e, as amended by Dz.U. 2024 poz. 13).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Carry a supply date on the sales event. Write P_6 and DataSprzedazy whenever it differs from P_1. The supply date, not P_1, also drives the VAT month (see JPK-04; the VAT reviewer owns the tax-point rule).
- **Pack today:** `sales.invoice.issued` carries only `issue_date`. fa3.go never writes P_6. jpk.go has no DataSprzedazy.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2024/13/D2024000001301.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/faktury/schemy/FA/schemat_FA(3)_v1-0E.xsd

### KSEF-09 — Does the invoice need the words "mechanizm podzielonej płatności"?
- **Applies to:** all VAT payers (selling Annex 15 goods or services to a taxpayer)
- **The law:** art. 106e ust. 1 pkt 18a: „w przypadku faktur, w których kwota należności ogółem przekracza kwotę 15 000 zł … obejmujących dokonaną na rzecz podatnika dostawę towarów lub świadczenie usług, o których mowa w załączniku nr 15 do ustawy – wyrazy „mechanizm podzielonej płatności””. FA(3) records this in `Adnotacje/P_18A` (1 = yes, 2 = no).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Set P_18A=1 when the gross exceeds 15,000 zł, at least one line is an Annex 15 item, and the buyer is a taxpayer. This needs an item classification, the same one the GTU codes need (JPK-07).
- **Pack today:** fa3.go hard-codes P_18A "2".
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/faktury/schemy/FA/schemat_FA(3)_v1-0E.xsd

### KSEF-10 — An invoice in euros: what must be stated in złoty?
- **Applies to:** all VAT payers
- **The law:** art. 106e ust. 11: „Kwoty podatku wykazuje się w złotych.” FA(3) has P_14_xW: „W przypadku gdy faktura jest wystawiona w walucie obcej - kwota podatku … przeliczona zgodnie z przepisami działu VI”, plus `KursWaluty`.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** For any non-PLN invoice, write P_14_xW and the exchange rate used. The VAT reviewer decides which rate applies (art. 31a, the NBP rate).
- **Pack today:** fa3.go writes `KodWaluty` and the foreign amounts only. `sales_invoice` has a currency but no rate. Per project memory, the currency type is not built yet.
- **Verdict:** wrong. A non-PLN invoice would go out without VAT in PLN. Whether KSeF rejects it for that is unverified.
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/faktury/schemy/FA/schemat_FA(3)_v1-0E.xsd

### KSEF-11 — A VAT-exempt sole trader issues an invoice: what legal basis goes on it?
- **Applies to:** VAT-exempt
- **The law:**
  - **The invoice must state the provision:** art. 106e ust. 1 pkt 19: „w przypadku dostawy towarów lub świadczenia usług zwolnionych od podatku na podstawie art. 43 ust. 1, art. 113 ust. 1 i 9 … – wskazanie: a) przepisu ustawy albo aktu wydanego na podstawie ustawy, na podstawie którego podatnik stosuje zwolnienie od podatku”.
  - **FA(3) fields:**
    - P_19A is for the provision of the act: „należy wskazać przepis ustawy albo aktu wydanego na podstawie ustawy”.
    - P_19B is for the directive.
    - P_19C is for „inną podstawę prawną”.
  - **KSeF:** exempt taxpayers are in KSeF (KSEF-01).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Put the provision in P_19A, taken from the company's declared exemption: "art. 113 ust. 1 ustawy o VAT" for the turnover exemption, or the art. 43 ust. 1 point for an exempt activity.
- **Pack today:** fa3.go writes P_19C „zwolnienie z VAT” on every exempt invoice. That is no legal basis, and it is in the wrong field.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/faktury/schemy/FA/schemat_FA(3)_v1-0E.xsd

### KSEF-12 — The customer pays in advance: which kind of invoice goes to KSeF?
- **Applies to:** all VAT payers
- **The law:**
  - **Advance invoice:** art. 106b ust. 1 pkt 4 requires an invoice for „otrzymanie przez niego całości lub części zapłaty przed dokonaniem czynności”.
  - **Final invoice:** art. 106f ust. 3: the final invoice „powinna również zawierać numery identyfikujące w Krajowym Systemie e-Faktur faktury wystawione przed wydaniem towaru lub wykonaniem usługi”.
  - **FA(3) `RodzajFaktury`:** VAT, KOR, ZAL (advance), ROZ (settling), UPR, KOR_ZAL, KOR_ROZ.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** An advance payment produces a ZAL invoice. The final invoice is ROZ and lists the KSeF numbers of the ZAL invoices it settles.
- **Pack today:** fa3.go always writes `RodzajFaktury` "VAT". No advance event exists.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/faktury/schemy/FA/schemat_FA(3)_v1-0E.xsd

### KSEF-13 — When has the company received a purchase invoice, and from which month may it deduct the VAT?
- **Applies to:** all VAT payers
- **The law:**
  - **Receipt date:** art. 106na ust. 3: „Faktura ustrukturyzowana jest uznana za otrzymaną przy użyciu Krajowego Systemu e-Faktur w dniu przydzielenia w tym systemie numeru identyfikującego tę fakturę.” Offline invoices follow the same rule (art. 106nda ust. 11).
  - **Deduction:** art. 86 ust. 10b pkt 1: the right to deduct „powstaje nie wcześniej niż w rozliczeniu za okres, w którym podatnik otrzymał fakturę”.
  - **Practical consequence:** an invoice dated 31 March but numbered on 1 April is received in April.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** The purchase's VAT month is the later of two months: the month of the supplier's tax point and the month the KSeF number was assigned. Take the assignment date from the KSeF metadata, or failing that from the date inside the KSeF number. A later deduction remains allowed (art. 86 ust. 11, the VAT reviewer's area).
- **Pack today:** `pl-book-purchase-invoice` and `pl-book-purchase-invoice-ksef` set `vat_period = month($.invoice_date)`, which is P_1. That deducts a month early whenever P_1 and the KSeF number straddle a month end.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/faktury/numer-ksef.md

### KSEF-14 — An invoice from a foreign supplier: KSeF or not, and how is it marked?
- **Applies to:** all VAT payers
- **The law:**
  - **Outside KSeF:** foreign issuers are outside the obligation (art. 106ga ust. 2 pkt 1). Ust. 3: „W przypadku, o którym mowa w ust. 2, wystawia się faktury elektroniczne lub faktury w postaci papierowej.”
  - **JPK marking (Ministry guide, purchases):** BFK covers invoices „wystawionych przy wewnątrzwspólnotowym nabyciu towarów (WNT) oraz imporcie usług”. Where no invoice exists, the record rests on an internal document (WEW) marked DI.
- **Grade:** 2 settled — the regulation only says BFK = „faktura elektroniczna oraz faktura w postaci papierowej”; the Ministry's marking guide (podatki.gov.pl, „JPK_VAT z deklaracją – stosowanie … znaczników”) settles it
- **Rhea's default:** Book it from the PDF/e-mail channel and mark it BFK on the purchase side. Mark BFK also on the self-assessed output row. With no invoice, use an internal document: WEW + DI.
- **Pack today:**
  - The purchase row is right: pdf-invoice.json produces a `purchase_invoice` without `ksef_ref`, which jpk.go marks BFK.
  - For VAT reviewer — reverse charge is not handled:
    - jpk.go writes no output rows for reverse-charge purchases.
    - It puts every purchase's net and VAT into K_42/K_43.
- **Verdict:** right (for the KSeF marking)
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://www.podatki.gov.pl/media/lzafohup/jpk_vat-z-deklaracja-stosowanie-znacznikow.pdf ; https://dziennikustaw.gov.pl/DU/2025/1800/D2025000180001.pdf

### KSEF-15 — An invoice did not reach KSeF on its issue day (offline24): by when must it go?
- **Applies to:** everyone
- **The law:**
  - **The rule:** art. 106nda ust. 1 lets a taxpayer issue the invoice electronically on the template. Under ust. 2 it must then be sent „niezwłocznie, nie później niż w następnym dniu roboczym po dniu jej wystawienia … do Krajowego Systemu e-Faktur w celu przydzielenia numeru”.
  - **Late = offline:** under ust. 16, a structured invoice sent after its P_1 date is itself such an invoice.
  - **A KSeF failure in between:** art. 106nha applies.
  - **Ministry FAQ:** „musisz ją dosłać do systemu najpóźniej następnego dnia roboczego”.
  - **Penalty:** art. 106ni ust. 1 pkt 3, from 1 Jan 2027 (KSEF-22).
  - **"Dzień roboczy":** I found no definition in the VAT act (unverified). The Ministry's technical example counts Monday to Friday.
- **Grade:** 1 plain — the text decides
- **Rhea's default:**
  - The KSeF channel runs at least daily.
  - Any sales invoice whose P_1 day has passed without a KSeF number becomes an offline24 invoice.
  - Its due date is the next working day after P_1, by the Polish holiday calendar.
  - It goes with `offlineMode: true`.
  - It surfaces as a high-priority task before the deadline, never as a silent retry past it.
- **Pack today:**
  - ksef.go sends whenever a pass happens to run, with no `offlineMode` and no deadline.
  - A failed session "retries next pass".
  - There is no working-day calendar.
  - No task exists for an invoice that simply has not been sent.
- **Verdict:** missing
- **Deadline:** the next working day after the issue date (P_1)
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2025/1203/D2025000120301.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/tryby-offline.md ; https://ksef.podatki.gov.pl/jdg-i-msp/

### KSEF-16 — KSeF is down: what differs between maintenance, a failure and a total failure?
- **Applies to:** everyone
- **The law:**
  - **Announcements (art. 106ne):** the Ministry announces failures in BIP and the API (ust. 1), a total failure in the mass media (ust. 3), and unavailability in BIP (ust. 4).
  - **Unavailability (art. 106nh ust. 2):** send „nie później niż w następnym dniu roboczym po dniu zakończenia okresu trwania niedostępności Krajowego Systemu e-Faktur”.
  - **Failure (art. 106nf ust. 4):** „W terminie 7 dni roboczych od dnia zakończenia awarii … podatnik jest obowiązany do przesłania” the invoices. A new failure announcement restarts the count (ust. 5).
  - **Total failure (art. 106ng):** the taxpayer „w okresie trwania awarii wystawia faktury w postaci papierowej lub faktury elektroniczne”. Nothing is sent later (art. 106nf ust. 7; art. 106nha pkt 2: „podatnik nie przesyła tej faktury”).
  - **JPK markings (Ministry guide):** OFF only for failure-mode (106nf) invoices without a number at filing; BFK for total-failure invoices.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Record each Ministry announcement (kind, start, end) as an event, so a rule computes every affected invoice's deadline and its JPK marking. Until a reader for BIP exists, a person enters the announcement.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** next working day after unavailability ends; 7 working days after a failure ends
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/tryby-offline.md ; https://www.podatki.gov.pl/media/lzafohup/jpk_vat-z-deklaracja-stosowanie-znacznikow.pdf

### KSEF-17 — KSeF rejected an invoice: what now?
- **Applies to:** everyone
- **The law:** The act says nothing about rejections. A rejected invoice has no number, and if it is re-sent on a later day than P_1 it becomes an offline24 invoice under art. 106nda ust. 16, with that invoice's deadline (KSEF-15). The Ministry's technical documentation adds the „korekta techniczna”: it „umożliwia ponowne przesłanie faktury wystawionej w trybie offline, która po przesłaniu do systemu KSeF została odrzucona z powodu błędów technicznych”. It is sent with `hashOfCorrectedInvoice`, and „nie jest dozwolone korygowanie treści faktury”.
- **Grade:** 2 settled — the law is silent on rejections; the Ministry's KSeF 2.0 technical documentation (offline/korekta-techniczna.md) settles the resubmission route
- **Rhea's default:** A rejection is a task due before that invoice's offline24 deadline. If the invoice was already handed to the buyer as an offline invoice, the fixed XML goes as a technical correction (offline mode, `hashOfCorrectedInvoice` = hash of the rejected file). Otherwise it is a plain resend. Changes to content need a correction invoice, which is possible only once a number exists.
- **Pack today:**
  - `pl-record-ksef-rejection` and `pl-ksef-rejection-task` are right in spirit: a high-priority task due on the rejection date, and the same XML is never re-sent.
  - Technical correction is missing: no `hashOfCorrectedInvoice` and no `offlineMode`.
  - The task date is not the legal deadline.
- **Verdict:** missing
- **Deadline:** that invoice's offline24 deadline
- **Sources:** https://github.com/CIRFMF/ksef-docs/blob/main/offline/korekta-techniczna.md ; https://github.com/CIRFMF/ksef-docs/blob/main/tryby-offline.md ; https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf

### KSEF-18 — We must correct an invoice that is already in KSeF.
- **Applies to:** everyone
- **The law:**
  - **Who corrects:** the seller issues the correction (art. 106j ust. 1).
  - **It must reference the original:** ust. 2 pkt 2a: „numer identyfikujący w Krajowym Systemie e-Faktur fakturę, której dotyczy faktura korygująca”.
  - **Form:** ust. 4: „Fakturę korygującą fakturę ustrukturyzowaną wystawia się w postaci faktury ustrukturyzowanej albo faktury, o której mowa w art. 106nda ust. 1.”
  - **Only after numbering:** art. 106nda ust. 12.
  - **Seller's VAT month:** art. 29a ust. 13: the base is reduced „za okres rozliczeniowy, w którym podatnik wystawił fakturę korygującą w postaci faktury ustrukturyzowanej”. No buyer confirmation is needed; that rule now applies only to corrections outside KSeF (ust. 13a).
  - **Buyer's VAT month:** art. 86 ust. 19a: the buyer reduces input VAT „w rozliczeniu za okres, w którym otrzymał fakturę korygującą”.
  - **FA(3):** `RodzajFaktury` KOR, with `DaneFaKorygowanej/NrKSeFFaKorygowanej`.
  - **Ministry FAQ:** there is no limit on the number of corrections.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** A correction is a new sales event of kind KOR that references the original's KSeF number and goes through the same channel. The seller reduces output VAT in the month the correction is issued in KSeF. The buyer reduces input VAT in the month its KSeF number is assigned. An accepted invoice is never edited in place.
- **Pack today:** missing. No correction event or type. fa3.go writes only "VAT". `sales_invoice` is amendable only in `open`.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2025/1203/D2025000120301.pdf ; https://ksef.podatki.gov.pl/jdg-i-msp/ ; https://github.com/CIRFMF/ksef-docs/blob/main/faktury/schemy/FA/schemat_FA(3)_v1-0E.xsd

### KSEF-19 — Can the buyer fix our mistake with a correction note (nota korygująca)?
- **Applies to:** everyone
- **The law:** No. Art. 106k now reads „(uchylony)”: Dz.U. 2023 poz. 1598, art. 1 pkt 15 („uchyla się art. 106k”), in force 1 Feb 2026. By art. 15 of that act, notes issued before then follow the old art. 106k ust. 2 and art. 106l. Accounting notes (noty księgowe, for example for contractual penalties) are outside VAT and outside KSeF (secondary — verify).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Offer no correction-note path. A buyer who spots an error raises a task asking the seller for a correction invoice. Notes for invoices from before 1 Feb 2026 are booked under the old rules.
- **Pack today:** nothing implemented, and nothing needed
- **Verdict:** right
- **Deadline:** since 1 Feb 2026
- **Sources:** https://dziennikustaw.gov.pl/DU/2023/1598/D2023000159801.pdf ; https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://www.podatki.biz/artykuly/noty-korygujace-po-wprowadzeniu-obowiazkowego-ksef_34_59625.htm

### KSEF-20 — May we attach a detailed specification (załącznik) to a KSeF invoice?
- **Applies to:** all VAT payers (with complex quantity or price data)
- **The law:**
  - **Allowed for:** „czynności o złożonej liczbie danych w zakresie jednostek miary i ilości” (art. 106gba ust. 1).
  - **Notification:** made „wyłącznie za pośrednictwem konta w e-Urzędzie Skarbowym” (ust. 3).
  - **Use:** only after confirmation (ust. 7). The confirmation comes within 3 working days of the UPO (regulation 1815 § 11 ust. 3).
  - **One only:** § 12 ust. 2: „może posiadać tylko jeden załącznik”.
  - **Dates:** notifications from 1 Jan 2026 (art. 17a of the 2023 act as amended); use from 1 Feb 2026.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Not offered; a small business's invoices fit in the invoice lines.
- **Pack today:** missing (fa3.go has no `Zalacznik`)
- **Verdict:** later
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2025/1815/D2025000181501.pdf

### JPK-02 — By when is the JPK due, and what moves the date?
- **Applies to:** all VAT payers
- **The law:**
  - **Monthly:** art. 99 ust. 1, „do 25. dnia miesiąca następującego po każdym kolejnym miesiącu”. The records go „łącznie z deklaracją” (art. 109 ust. 3b).
  - **Quarterly:** records monthly, declaration with the quarter's last month (art. 109 ust. 3c).
  - **Weekends and holidays:** Tax Ordinance art. 12 § 5: „Jeżeli ostatni dzień terminu przypada na sobotę lub dzień ustawowo wolny od pracy, za ostatni dzień terminu uważa się następny dzień po dniu lub dniach wolnych od pracy”.
  - **Suspension:** art. 99 ust. 7a means no return for suspended months, with the exceptions in ust. 7b.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Compute each period's due date (the 25th, rolled past Saturdays, Sundays and holidays) and open a filing task carrying it.
- **Pack today:** the `tax_declaration` type has `due_date`, but no rule fills it. The writer runs only on demand (`rhea file`).
- **Verdict:** missing
- **Deadline:** the 25th of the following month (quarterly: records monthly, declaration by the 25th after the quarter)
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2026/622/D2026000062201.pdf

### JPK-03 — We are VAT-exempt: do we file a JPK_V7 at all?
- **Applies to:** VAT-exempt
- **The law:**
  - **No JPK_V7:** art. 99 ust. 7: „Przepisu ust. 1 nie stosuje się do podatników, u których sprzedaż jest zwolniona od podatku na podstawie art. 113 albo art. 113a, oraz podmiotów wykonujących wyłącznie czynności zwolnione … chyba że” they claim a refund or must make the listed adjustments.
  - **Daily sales record instead:** turnover-exempt taxpayers keep a daily sales record (art. 109 ust. 1).
  - **Still in KSeF:** they still invoice through KSeF (KSEF-01).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** For a VAT-exempt company, settle no VAT, write no JPK_V7, and keep the daily sales record. Its invoices still go through KSeF with the P_19A basis (KSEF-11).
- **Pack today:** `pl-settle-vat` runs on every `period.closed` for book `pl-stat`, whatever the VAT status. jpk.go does not check status either. `company` has no VAT-status field.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf

### JPK-04 — In which month's records does each invoice go?
- **Applies to:** all VAT payers
- **The law:**
  - **Sales by tax point:** regulation annex, part II pkt 13, as worded by Dz.U. 2025 poz. 1800: „Podatnicy ujmują w ewidencji faktury i inne dowody według daty powstania obowiązku podatkowego.”
  - **Purchases by deduction right:** art. 86 ust. 10 and 10b (KSEF-13).
  - **Exception:** invoices to receipts (FP) go in the month they are issued (art. 109 ust. 3d).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** A sales record's month is the month of the tax point; the VAT reviewer supplies that rule. A purchase record's month is the deduction month. Neither is "the month of P_1".
- **Pack today:** `pl-book-sales-invoice` sets `vat_period = month(issue_date)`, and jpk.go filters by `vat_period`.
- **Verdict:** wrong (wherever the tax point and the issue month differ)
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2025/1800/D2025000180001.pdf ; https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf

### JPK-05 — Which KSeF marking does each record carry: KSeF number, OFF, BFK or DI?
- **Applies to:** all VAT payers
- **The law:** Regulation § 10 ust. 5 and § 11 ust. 8, as worded by Dz.U. 2025 poz. 1800:
  - „„OFF” – faktura, o której mowa w art. 106nf ust. 1 ustawy, która na dzień złożenia ewidencji nie posiada numeru identyfikującego”
  - „„BFK” – faktura elektroniczna oraz faktura w postaci papierowej”
  - „„DI” – dowód inny niż faktura”

  The annex, part II:
  - **pkt 14:** the KSeF number is shown „o ile został on nadany i nastąpiło to do dnia przesłania ewidencji”.
  - **pkt 15:** offline24 and unavailability invoices without a number at filing are marked „DI”. After the number arrives, „podatnik koryguje ewidencję”.

  Ministry marking guide:
  - **DI (sales side):** includes „faktur wystawionych w trybie offline24 … które na dzień przesłania ewidencji nie posiadają numeru”.
  - **OFF:** „Podatnik nie ma obowiązku dokonania korekty ewidencji JPK_VAT po uzyskaniu numeru KSeF”.
  - **BFK:** invoices issued lawfully outside KSeF:
    - before 1 Feb 2026
    - in February and March 2026 by taxpayers obliged from April
    - by taxpayers not obliged
    - during a total failure
    - WNT and import-of-services invoices
- **Grade:** 2 settled — the regulation lists the marks; the Ministry's marking guide (podatki.gov.pl) and the JPK_V7(3) brochure of January 2026 settle which case takes which
- **Rhea's default:**
  - **NrKSeF:** whenever a number exists on the day the file is sent.
  - **OFF:** failure-mode invoices without a number.
  - **DI:** offline24 and unavailability invoices without a number, corrected later with the number.
  - **BFK:** only invoices issued lawfully outside KSeF.
  - **Never BFK** for an invoice that is merely late, pending or rejected.
- **Pack today:**
  - jpk.go, sales: the number from an accepted `ksef_submission`, otherwise „BFK”.
  - jpk.go, purchases: `ksef_ref`, otherwise „BFK”.
  - No OFF or DI exists, and no correction follows when a number arrives.
- **Verdict:** wrong
- **Deadline:** each filing
- **Sources:** https://dziennikustaw.gov.pl/DU/2025/1800/D2025000180001.pdf ; https://www.podatki.gov.pl/media/lzafohup/jpk_vat-z-deklaracja-stosowanie-znacznikow.pdf ; https://www.podatki.gov.pl/media/wgbkrejs/broszura-jpk_vat-z-deklaracj%C4%85-od-1-lutego-2026-r.pdf

### JPK-06 — Does a sales record need GTU codes?
- **Applies to:** all VAT payers
- **The law:** Regulation Dz.U. 2019 poz. 1988, § 10 ust. 3 (codes „01”–„13”). For example, ust. 3 pkt 2 lit. b covers services „o charakterze niematerialnym – wyłącznie: doradczych, księgowych, prawnych, zarządczych, szkoleniowych, marketingowych, firm centralnych (head offices), reklamowych, badania rynku i opinii publicznej, w zakresie badań naukowych i prac rozwojowych – oznaczenie „12””. The brochure lists the current PKWiU scope. The codes are set „dla całej faktury poprzez zaznaczenie „1””. They are not used on RO or WEW records or on reverse-charge purchases.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Each product or service in master data carries its GTU, or none. An invoice record gets a 1 for each GTU present on any of its lines. A consultant's invoice, for instance, typically carries GTU_12.
- **Pack today:** missing. There is no item master data with GTU, and jpk.go writes no GTU fields.
- **Verdict:** missing
- **Deadline:** each filing
- **Sources:** https://dziennikustaw.gov.pl/DU/2019/1988/D2019000198801.pdf ; https://www.podatki.gov.pl/media/wgbkrejs/broszura-jpk_vat-z-deklaracj%C4%85-od-1-lutego-2026-r.pdf

### JPK-07 — We sell to a related company or person: the TP marking and other procedure codes.
- **Applies to:** all VAT payers (mostly sp. z o.o. inside groups, and sole traders selling to family companies)
- **The law:**
  - **TP:** regulation § 10 ust. 4 pkt 3: „istniejących powiązań między nabywcą a dokonującym dostawy towarów lub usługodawcą, o których mowa w art. 32 ust. 2 pkt 1 ustawy – oznaczenie „TP””.
  - **The V7(3) list:** WSTO_EE, IED, TP, TT_WNT, TT_D, MR_T, MR_UZ, I_42, I_63, B_SPV, B_SPV_DOSTAWA, B_MPV_PROWIZJA. The brochure has no MPP.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Mark the counterparty as related under art. 32 ust. 2 pkt 1. That sets TP on every record with it. The other procedures come when a pack needs them.
- **Pack today:** `pl-book-sales-invoice` fills `sales_invoice.intragroup`, but jpk.go ignores it.
- **Verdict:** missing
- **Deadline:** each filing
- **Sources:** https://dziennikustaw.gov.pl/DU/2019/1988/D2019000198801.pdf ; https://www.podatki.gov.pl/media/wgbkrejs/broszura-jpk_vat-z-deklaracj%C4%85-od-1-lutego-2026-r.pdf

### JPK-08 — Sales recorded by a cash-register report, an internal document or an invoice to a receipt.
- **Applies to:** all VAT payers (with cash registers or self-supply; common for sole traders selling to consumers)
- **The law:** Regulation § 10 ust. 5, as worded by Dz.U. 2025 poz. 1800:
  - pkt 1 lit. c: „„FP” – faktura, o której mowa w art. 106h ust. 1 ustawy”
  - pkt 2: „a) „DI” – dowód inny niż faktura, b) „RO” – zbiorczy dowód sprzedaży z kas rejestrujących, c) „WEW” – dowód wewnętrzny”

  Further rules:
  - **§ 11c:** until 31 Dec 2026, FP also covers invoices issued from cash registers.
  - **Art. 109 ust. 3d:** FP invoices do not increase sales.
  - **Purchase-side document types:** MK, VAT_RR, WEW.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** A monthly cash-register report becomes one RO record marked DI. An internal document becomes WEW plus DI. An invoice to a receipt is FP and is left out of the totals.
- **Pack today:** missing. There is no cash-register report type, and jpk.go writes no `TypDokumentu`.
- **Verdict:** missing
- **Deadline:** each filing
- **Sources:** https://dziennikustaw.gov.pl/DU/2025/1800/D2025000180001.pdf ; https://www.podatki.gov.pl/media/wgbkrejs/broszura-jpk_vat-z-deklaracj%C4%85-od-1-lutego-2026-r.pdf

### JPK-09 — A record's counterparty has no Polish NIP (a consumer, a foreign firm).
- **Applies to:** all VAT payers
- **The law:**
  - **Regulation:** § 10 ust. 2 pkt 1 lit. a: the number is „zapisywany poprzez wyodrębnienie kodu kraju oraz kodu cyfrowo-literowego”.
  - **Brochure, NrKontrahenta:** „W przypadku braku - w polu tym należy wpisać „BRAK””.
  - **Brochure, KodKrajuNadaniaTIN:** „Podaje się literowy kod kraju”.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Write "BRAK" when the counterparty has no number. Split a foreign number's country prefix into KodKrajuNadaniaTIN.
- **Pack today:** jpk.go `nipOf` strips only "PL" and writes "" when there is no number. It has no KodKrajuNadaniaTIN.
- **Verdict:** wrong
- **Deadline:** each filing
- **Sources:** https://dziennikustaw.gov.pl/DU/2019/1988/D2019000198801.pdf ; https://www.podatki.gov.pl/media/wgbkrejs/broszura-jpk_vat-z-deklaracj%C4%85-od-1-lutego-2026-r.pdf

### JPK-10 — What a purchase record must carry: the KSeF number and the date of receipt.
- **Applies to:** all VAT payers
- **The law:**
  - **KSeF number required:** regulation § 11 ust. 4a, as worded by Dz.U. 2025 poz. 1800: „Dane dotyczące faktur, o których mowa w ust. 3 pkt 1 i 5, obejmują numer identyfikujący fakturę w Krajowym Systemie e-Faktur.” Until then the 2024 text said „mogą obejmować”.
  - **Receipt date required:** § 11 ust. 4 pkt 5 (Dz.U. 2024 poz. 13): „datę otrzymania faktury lub innego dowodu”.
  - **Brochure, DataWplywu:** „Podaje się datę otrzymania dowodu zakupu, o ile jest inna niż DataZakupu”.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Take the KSeF number from the KSeF download itself, not from a file name. Take the receipt date from the number's assignment date, and write it as DataWplywu whenever it differs from P_1.
- **Pack today:**
  - ksef_inbound.go sets `ksef_ref` only when the file name is a KSeF number; `pl-book-purchase-invoice-ksef` stores it.
  - jpk.go writes DataZakupu = `invoice_date` and no DataWplywu.
  - A missing number becomes BFK.
- **Verdict:** wrong
- **Deadline:** each filing
- **Sources:** https://dziennikustaw.gov.pl/DU/2025/1800/D2025000180001.pdf ; https://dziennikustaw.gov.pl/DU/2024/13/D2024000001301.pdf ; https://www.podatki.gov.pl/media/wgbkrejs/broszura-jpk_vat-z-deklaracj%C4%85-od-1-lutego-2026-r.pdf

### JPK-11 — A month with nothing to report, or a month not yet closed.
- **Applies to:** all VAT payers
- **The law:**
  - **Every month:** art. 99 ust. 1 requires a return for each month; only suspension lifts it (ust. 7a).
  - **Brochure, V7M:** filers are „obowiązani są do wypełniania wszystkich elementów JPK_V7M w pliku XML, tj. Naglowek, Podmiot1, Deklaracja oraz Ewidencja”. A declaration may be left out only in a records-only correction.
  - **P_51:** „W przypadku braku – w polu należy wykazać „0””.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** File a zero return every month unless the business is suspended. Refuse to render an original JPK_V7M without its declaration.
- **Pack today:** when no `vat_settlement` exists, jpk.go renders the records without a Deklaracja. An original V7M in that shape is not a valid filing.
- **Verdict:** wrong
- **Deadline:** the 25th
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://www.podatki.gov.pl/media/wgbkrejs/broszura-jpk_vat-z-deklaracj%C4%85-od-1-lutego-2026-r.pdf

### JPK-12 — How is the JPK signed and sent, who may sign it, and what proves it was filed?
- **Applies to:** all VAT payers (and, later, the annual JPK_PKPIR, JPK_EWP and JPK_KR_PD files)
- **The law:**
  - **How it is sent:** art. 99 ust. 11c: the file „jest przesyłany w sposób określony w przepisach wydanych na podstawie art. 193a § 3 Ordynacji podatkowej”.
  - **Ministry interface spec 5.6.0 (10 Sep 2026):** the metadata are authenticated by „podpisu kwalifikowanego (polski lub europejski)”, „podpisu zaufanego”, or „AuthData zawierającego zaszyfrowane dane autoryzujące”. For JPK_V7M(3) and V7K(3): „podpis kwalifikowany, podpis zaufany, dane autoryzujące”.
  - **Status codes:**
    - 200 means „Przetwarzanie dokumentu zakończone poprawnie, pobierz UPO”.
    - 420 means „Brak aktualnego pełnomocnictwa/upoważnienia do podpisywania dokumentu”.
  - **Who may sign:** Tax Ordinance art. 80a § 1 lets a „pełnomocnika podatnika” sign. Under § 2a, the power of attorney for e-declarations goes to the tax office that keeps the taxpayer register. The form name UPL-1 is from secondary sources — verify.
  - **The same power covers the books:** under Dz.U. 2026 poz. 779, that power of attorney also covers signing the books sent as JPK.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Rhea prepares and validates the file, and a person signs it: the owner by trusted profile or authorisation data, or the accountant holding the power of attorney. Later Rhea may send through the JPK API itself. Either way the UPO is kept as the evidence of filing, as it is for KSeF.
- **Pack today:** `rhea file` and `rhea issue` render and store the file (`issued_document`). Nothing sends it, and no UPO or status is recorded.
- **Verdict:** missing
- **Deadline:** the 25th
- **Sources:** https://www.podatki.gov.pl/media/bngboylg/specyfikacja-interfejs%C3%B3w-us%C5%82ug-jpk-wersja-560.pdf ; https://dziennikustaw.gov.pl/DU/2026/622/D2026000062201.pdf ; https://dziennikustaw.gov.pl/DU/2026/779/D2026000077901.pdf

### JPK-13 — Correcting a JPK already filed.
- **Applies to:** all VAT payers
- **The law:**
  - **Right to correct:** Tax Ordinance art. 81 § 1: taxpayers „mogą skorygować uprzednio złożoną deklarację”. Under § 2 a correction is made by filing a correcting declaration. Under art. 81b § 1 pkt 1 the right is suspended during a tax audit.
  - **Duty for the records:** VAT act art. 109 ust. 3e: „W terminie 14 dni od dnia: 1) stwierdzenia przez podatnika, że przesłana ewidencja zawiera błędy lub dane niezgodne ze stanem faktycznym … podatnik jest obowiązany przesłać korektę ewidencji”.
  - **Form (brochure):**
    - CelZlozenia „2 – korekta”.
    - A records correction is „nowy, kompletny oraz zawierający poprawione dane plik XML. Niedopuszczalne jest złożenie pliku zawierającego jedynie dane korygowane”.
    - A correction may cover the declaration only or the records only.
    - P_ORDZU (the reason) is optional.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** A correction is a new version of the period's file (CelZlozenia 2), derived from the log as it now stands and kept beside the original. Rhea opens a task with the 14-day deadline when something changes in a filed month, for example a DI record receiving its KSeF number, or a late invoice arriving.
- **Pack today:** jpk.go fixes CelZlozenia at 1. The period lock (packs/common/periods) refuses late invoices to closed months but offers no correction path.
- **Verdict:** missing
- **Deadline:** 14 days from finding the error
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/622/D2026000062201.pdf ; https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://www.podatki.gov.pl/media/wgbkrejs/broszura-jpk_vat-z-deklaracj%C4%85-od-1-lutego-2026-r.pdf

### JPK-14 — The tax office calls on us to correct or explain the JPK.
- **Applies to:** all VAT payers
- **The law:**
  - **The call:** VAT act art. 109 ust. 3f: „naczelnik urzędu skarbowego wzywa podatnika do ich skorygowania, wskazując te błędy”.
  - **14 days to answer:** ust. 3g: „Podatnik w terminie 14 dni od dnia doręczenia wezwania … przesyła … ewidencję skorygowaną w zakresie błędów wskazanych w wezwaniu lub składa wyjaśnienia”.
  - **Penalty:** ust. 3h allows „karę pieniężną w wysokości 500 zł za każdy błąd”.
  - **Sole traders:** ust. 3i excludes a sole trader already liable under the fiscal penal code for the same act.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** A letter from the office is an event. It opens a task with the 14-day deadline counted from delivery, listing the records it names. The answer is either a correction (JPK-13) or a written explanation, kept as evidence.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** 14 days from delivery of the call
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf

### JPK-15 — We found our own error after filing: how do we stay clear of a fiscal-penal case (czynny żal or correction)?
- **Applies to:** everyone
- **The law:**
  - **Correction protects:** fiscal penal code (KKS) art. 16a § 1: „Nie podlega karze … sprawca czynu zabronionego dotyczącego złożenia deklaracji lub przesłania księgi, jeżeli po jego popełnieniu została złożona organowi podatkowemu prawnie skuteczna korekta deklaracji lub księgi”. Under § 2 the tax must be paid „niezwłocznie”. Under § 3 it does not help once proceedings have started.
  - **Czynny żal:** art. 16 § 1 (self-reporting) is the route for other acts, including a late filing. Under § 4a it may be filed through e-Urząd Skarbowy. Under § 5 it is ineffective once the authority already knows.
  - **KSeF lapses:** no penal case follows (VAT act art. 106ni ust. 4: „nie wszczyna się postępowania w sprawach o przestępstwa skarbowe lub wykroczenia skarbowe”).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** A correction that raises the tax due is followed by payment of the difference the same day; interest belongs to the Tax Ordinance reviewer. A late or missed filing gets a czynny żal draft for a person to sign. KSeF lapses need none.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** before the authority acts
- **Sources:** https://dziennikustaw.gov.pl/DU/2025/633/D2025000063301.pdf ; https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf

### KSEF-21 — Our KSeF token stops working: when must we switch to certificates?
- **Applies to:** everyone (any company integrating through the API)
- **The law:**
  - **Tokens until year end:** regulation Dz.U. 2025 poz. 1815, § 13 ust. 1 pkt 1: „Do: 1) dnia 31 grudnia 2026 r. do uwierzytelniania się podmiotów korzystających z Krajowego Systemu e-Faktur może służyć ciąg znaków alfanumerycznych, o którym mowa w § 5 ust. 1 pkt 4 rozporządzenia Ministra Finansów z dnia 27 grudnia 2021 r.”. That string is the KSeF token.
  - **No certificates via a token:** § 13 ust. 2: after token authentication „nie będzie możliwe wytworzenie certyfikatu KSeF”.
  - **Certificates:** § 6 ust. 1 pkt 4 names the KSeF certificate, valid „nie dłużej niż 2 lata” (§ 6 ust. 2).
  - **Pending:** in consultations from 9 Jun 2026 the Ministry proposed keeping tokens after 31 Dec 2026 (podatki.biz, inforlex; secondary — verify). I found no amending regulation as of 16 Oct 2026.
- **Grade:** 1 plain — the text decides (a proposal to extend is pending, not enacted)
- **Rhea's default:** Build certificate authentication (an Authentication-type KSeF certificate, enrolled after the owner's qualified signature or trusted profile, used through XAdES) before 31 Dec 2026. Keep the token only as a fallback if the regulation is amended. The same enrolment gives the Offline certificate that KSEF-05 needs.
- **Pack today:** markets/pl/ksef2.go authenticates only with a KSeF token (`/auth/ksef-token`). `ClientFromEnv` requires `RHEA_KSEF_TOKEN`.
- **Verdict:** missing
- **Deadline:** 31 Dec 2026
- **Sources:** https://dziennikustaw.gov.pl/DU/2025/1815/D2025000181501.pdf ; https://github.com/CIRFMF/ksef-docs/blob/main/certyfikaty-KSeF.md ; https://www.podatki.biz/artykuly/beda-zmiany-w-ksef-tokeny-maja-pozostac_55_62973.htm

### KSEF-22 — What does a KSeF lapse cost, and from when?
- **Applies to:** everyone
- **The law:**
  - **The penalty:** VAT act art. 106ni ust. 1. The office imposes „karę pieniężną w wysokości do 100 % kwoty podatku wykazanego na tej fakturze wystawionej poza Krajowym Systemem e-Faktur, a w przypadku faktury bez wykazanego podatku – karę pieniężną w wysokości do 18,7 % kwoty należności ogółem”. It applies for an invoice issued outside KSeF, sent late, or not on the template.
  - **In force:** the consolidated text carries the footnote „Wejdzie w życie z dniem 1 stycznia 2027 r.” (art. 23 pkt 4 of the 2023 act, as worded by Dz.U. 2025 poz. 1203).
  - **No penal case:** art. 106ni ust. 4.
  - **Pending draft:** UD477 (Ministry of Finance and Economy; published 22 Sep 2026; Council of Ministers adoption planned for Q4 2026) would apply the penalty provisions „dopiero od dnia 1 stycznia 2028 r.”, with reminders and checks by KAS first. It is not law.
- **Grade:** 1 plain — the text decides (the UD477 draft could change the date)
- **Rhea's default:** Plan for penalties from 1 Jan 2027. Each invoice that is late or never sent becomes a high-priority task showing its exposure: up to its VAT, or up to 18.7% of the gross when it shows no VAT.
- **Pack today:** missing. Nothing tracks lateness (KSEF-15) or computes exposure.
- **Verdict:** missing
- **Deadline:** 1 Jan 2027 (the UD477 draft would make it 1 Jan 2028)
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2025/1203/D2025000120301.pdf ; https://www.gov.pl/web/premier/projekt-ustawy-o-zmianie-ustawy-o-podatku-od-towarow-i-uslug-oraz-ustawy-o-zmianie-ustawy-o-podatku-od-towarow-i-uslug-oraz-niektorych-innych-ustaw

### KSEF-23 — Paying and being paid: must the transfer title carry the KSeF number?
- **Applies to:** all VAT payers (payments between two active VAT payers)
- **The law:**
  - **The duty:** art. 108g ust. 1: the paying active VAT payer „jest obowiązany do podania numeru identyfikującego te faktury w Krajowym Systemie e-Faktur lub identyfikatora zbiorczego nadanego przez Krajowy System e-Faktur”.
  - **From when:** art. 17 of Dz.U. 2023 poz. 1598, as worded by Dz.U. 2025 poz. 1203: „stosuje się do płatności dokonanych od dnia 1 stycznia 2027 r.”. The same date applies to split payment (art. 108a).
  - **Exception:** invoices kept out of KSeF by a failure (ust. 4).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Outgoing payment proposals put the supplier invoice's KSeF number in the title. Incoming matching reads a KSeF number in a title as a reference to the sales invoice that carries it.
- **Pack today:** `pl-match-incoming` and `pl-match-outgoing` match `"reference": "number"` only. `sales_invoice` has no KSeF number (it lives on `ksef_submission`). Rhea writes no payment orders.
- **Verdict:** missing
- **Deadline:** payments made from 1 Jan 2027
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2025/1203/D2025000120301.pdf

### JPK-16 — Does a sole trader send the KPiR or the revenue register as a JPK (JPK_PKPIR, JPK_EWP), and when?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - **KPiR:** PIT act art. 24a ust. 1e, as worded by Dz.U. 2026 poz. 779 from 1 Jul 2026. The book is kept by computer and sent „po zakończeniu roku podatkowego w terminie: 1) do dnia upływu terminu złożenia zeznania, o którym mowa w art. 45 ust. 1 – w przypadku podatkowej księgi przychodów i rozchodów oraz ewidencji środków trwałych”. Art. 45 ust. 1 sets that deadline at 30 April.
  - **Ryczałt:** lump-sum act art. 15 ust. 12 covers the revenue register and the fixed-asset list, due by the PIT-28 deadline (art. 21 ust. 2 pkt 2: „do dnia 30 kwietnia roku następującego po roku podatkowym”).
  - **First year:** art. 66 ust. 1 of the 2021 Polski Ład act (Dz.U. 2021 poz. 2105), as worded by Dz.U. 2022 poz. 1265 and amended by Dz.U. 2023 poz. 1059. The first file covers the tax year starting after:
    - 31 Dec 2025 for those „obowiązane przesyłać ewidencję, o której mowa w art. 109 ust. 3 … zgodnie z art. 109 ust. 3b”, that is, monthly JPK_V7M filers;
    - 31 Dec 2026 for everyone else.
  - **KAS confirms:** monthly-JPK taxpayers keep electronic books from 2026 and send them in 2027; the rest send from 2028.
  - **Structures:** JPK_PKPIR(3), JPK_EWP(4), JPK_ST(1). Whether the September 2025 regulations exempt any group is unverified.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** From day one, keep the KPiR or revenue register as objects derived in the structure's own terms, so the annual file is a writer like jpk.go. For a sole trader filing V7M, the first file is due 30 Apr 2027.
- **Pack today:** missing. There are no KPiR or EWP objects or writers; see also review file 3.
- **Verdict:** missing
- **Deadline:** 30 Apr 2027 (tax year 2026, monthly-JPK filers); 30 Apr 2028 (tax year 2027, everyone else)
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/779/D2026000077901.pdf ; https://dziennikustaw.gov.pl/DU/2026/592/D2026000059201.pdf ; https://dziennikustaw.gov.pl/DU/2025/843/D2025000084301.pdf ; https://dziennikustaw.gov.pl/DU/2022/1265/D2022000126501.pdf ; https://dziennikustaw.gov.pl/DU/2023/1059/D2023000105901.pdf ; https://www.gov.pl/web/ias-rzeszow/elektroniczne-ksiegi-i-ewidencje-podatkowe

### JPK-17 — Does an sp. z o.o. send its books as JPK_KR_PD, and when?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - **The duty:** CIT act art. 9 ust. 1c: „Podatnicy prowadzący księgi rachunkowe są obowiązani prowadzić te księgi przy użyciu programów komputerowych oraz przesyłać właściwemu naczelnikowi urzędu skarbowego te księgi po zakończeniu roku podatkowego”.
  - **The deadline:** since 1 Jul 2026 (Dz.U. 2026 poz. 779) it is „do końca siódmego miesiąca po zakończeniu roku podatkowego”.
  - **First year:** art. 66 ust. 2 of the 2021 act as amended (Dz.U. 2023 poz. 1059). The first year is the one starting after:
    - „31 grudnia 2025 r. – w przypadku podatników … obowiązanych do przesyłania ewidencji, o której mowa w art. 109 ust. 3 …, zgodnie z art. 109 ust. 3b”, that is, monthly JPK filers;
    - „31 grudnia 2026 r. – w przypadku pozostałych”.
  - **Fixed assets:** JPK_ST_KR goes with it.
  - **Not read:** the 2024 regulation on additional ledger markings (Dz.U. 2024 poz. 1314).
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Write the general ledger, with the additional markings from the 2024 regulation, through a writer. For a calendar-year company filing V7M the first file covers 2026 and is due 31 Jul 2027.
- **Pack today:** missing. Journal entries exist, but there is no JPK_KR_PD writer and no extra markings; see also review file 5.
- **Verdict:** missing
- **Deadline:** 31 Jul 2027 (tax year 2026, monthly-JPK filers); 31 Jul 2028 (tax year 2027, others)
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/554/D2026000055401.pdf ; https://dziennikustaw.gov.pl/DU/2026/779/D2026000077901.pdf ; https://dziennikustaw.gov.pl/DU/2023/1059/D2023000105901.pdf ; https://www.gov.pl/web/ias-rzeszow/elektroniczne-ksiegi-i-ewidencje-podatkowe

### KSEF-24 — Must we keep our own copies of KSeF invoices?
- **Applies to:** everyone
- **The law:**
  - **KSeF keeps them:** VAT act art. 112aa ust. 1: invoices „są przechowywane w Krajowym Systemie e-Faktur przez okres 10 lat, licząc od końca roku, w którym zostały wystawione. Przepisów art. 112 i art. 112a nie stosuje się.”
  - **Unless limitation runs longer:** under ust. 2, keep them outside KSeF.
  - **Accounting act:** art. 73a (added by Dz.U. 2023 poz. 1598, art. 2) waives the storage rules of art. 71 ust. 1 and art. 73 for evidence kept in KSeF.
  - **Tax Ordinance:** art. 193a § 1a: the office does not demand structured invoices through JPK.
- **Grade:** 1 plain — the text decides
- **Rhea's default:** Keep the exact XML and the UPO anyway. It is cheap, and it makes replay and the QR code possible. Do not treat Rhea's copy as the statutory archive.
- **Pack today:** ksef.go keeps `invoice_xml` and `upo_xml` in the evidence event, and `ksef_submission` keeps the hash. Purchase files are kept as blobs under their hash.
- **Verdict:** right
- **Deadline:** —
- **Sources:** https://dziennikustaw.gov.pl/DU/2026/1263/D2026000126301.pdf ; https://dziennikustaw.gov.pl/DU/2023/1598/D2023000159801.pdf ; https://dziennikustaw.gov.pl/DU/2026/622/D2026000062201.pdf

## Sole trader — books and income tax

Scope: a natural person on KPiR (skala or liniowy) or on ryczałt, in the first year. Law as it stands on 16 October 2026. Primary texts read in full or in the relevant part: PIT act, the unified text prepared by Kancelaria Sejmu and dated 2026-09-16 (t.j. Dz.U. 2026 poz. 592 with amendments up to poz. 1123); ryczałt act, unified text (t.j. Dz.U. 2025 poz. 843 with 2026 poz. 779, 1098); accounting act, unified text dated 2026-10-08 (t.j. Dz.U. 2026 poz. 522 with amendments); the KPiR regulation Dz.U. 2025 poz. 1299; the ewidencja przychodów regulation Dz.U. 2025 poz. 1294; the additional-data regulation Dz.U. 2025 poz. 1311; Polski Ład (Dz.U. 2021 poz. 2105) art. 66; the amending acts Dz.U. 2021 poz. 2269, 2024 poz. 1863, 2025 poz. 1858 and 2026 poz. 779. All were fetched from the Sejm ELI API (api.sejm.gov.pl), which serves the same files as ISAP.

What changed for 2026–2027, in one place:
- **New KPiR regulation from 1 January 2026** (Dz.U. 2025 poz. 1299). It replaces the 2019 regulation and has 19 columns instead of 17. The new columns are the KSeF number (col. 3) and the counterparty's tax identifier (col. 5). Invoice-based revenue may now be entered by the 20th of the next month; under the old rule it was entered daily.
- **New ewidencja przychodów regulation from 1 January 2026** (Dz.U. 2025 poz. 1294). It adds a KSeF number column and the counterparty's tax identifier.
- **Electronic books.** The KPiR, the ewidencja przychodów and the fixed-asset register must be kept in software and sent once a year as JPK. VAT payers that file JPK_V7 start with tax year 2026. Everyone else starts with tax year 2027 (Polski Ład art. 66 ust. 1).
- **From 1 July 2026** (Dz.U. 2026 poz. 779) the sending deadline is split: KPiR and the fixed-asset register by the PIT return deadline, full books by 31 July.
- **Cash method of revenue (metoda kasowa).** The limit rose from 1 000 000 zł to 2 000 000 zł from 1 January 2026 (Dz.U. 2025 poz. 1858).
- **Passenger cars, from 1 January 2026:** a 100 000 zł limit for combustion cars emitting at least 50 g CO2/km, 150 000 zł below 50 g/km, and 225 000 zł for electric and hydrogen cars (Dz.U. 2021 poz. 2269).
- **Health contribution (liniowy).** The cap on the amount that may be deducted or counted as a cost is 14 100 zł for 2026 (M.P. 2025 poz. 1274).
- **Full-books threshold for natural persons:** 2 500 000 euro from financial year 2025, tested on 2024 revenue (Dz.U. 2024 poz. 1863 art. 14).
- **PLN equivalents of the euro thresholds.** Each is converted at the NBP mid rate on the first working day of October of the preceding year:
  - For 2026, the rate of 1 Oct 2025 is 4.2586: full books 10 646 500 zł; small taxpayer 8 517 000 zł (rounded to 1000); ryczałt limit 8 517 200 zł; ryczałt quarterly limit 851 720 zł; one-off depreciation 213 000 zł (rounded).
  - For 2027, the rate of 1 Oct 2026 is 4.3770: full books 10 942 500 zł; small taxpayer 8 754 000 zł; ryczałt limit 8 754 000 zł; ryczałt quarterly limit 875 400 zł; one-off depreciation 219 000 zł.

What the pack has today for this area:
- **company v4** knows `legal_form` (natural_person / legal_entity) and `owner`. Nothing records the form of tax, the form of books, the cost method, the cash method or how often advances are paid.
- **Types without rules.** CORPUS shipped `fixed_asset`, `vehicle`, `mileage_entry`, `lease_contract`, `tax_declaration` (kind `pit`, `zus_dra`) and `stocktake` as types. No rule reads or writes any of them.
- **Postings.** Every company is posted to the full-books chart: pl-post-sales-invoice gives 201/700/222, and pl-post-purchase-invoice gives 300/221/202.
- **Sales invoices.** `sales_invoice` has `issue_date` only, with no date of supply, and pl-book-sales-invoice sets `vat_period` to the month the invoice was issued.

---

### FORM-01 — How is the owner's business income taxed, and what happens if they never say?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** PIT act art. 9a ust. 1: „Dochody osiągnięte przez podatników ze źródła, o którym mowa w art. 10 ust. 1 pkt 3, są opodatkowane na zasadach określonych w art. 27, z zastrzeżeniem ust. 2 i 3, chyba że podatnicy złożą (...) oświadczenie o zastosowaniu opodatkowania w formie ryczałtu od przychodów ewidencjonowanych”. Art. 9a ust. 2 adds the 19% flat tax (art. 30c) as an option chosen by statement. Ryczałt act art. 9 ust. 3: „W przypadku rozpoczęcia działalności gospodarczej w trakcie roku podatkowego i niezłożenia w określonym terminie oświadczenia (...) podatnik jest obowiązany do założenia właściwych ksiąg i opłacania podatku na zasadach określonych w ustawie o podatku dochodowym.” The forms are skala (art. 27: 12% up to 120 000 zł minus 3600 zł, then 10 800 zł plus 32% of the excess), liniowy 19% (art. 30c ust. 1) and ryczałt. Karta is closed to new entrants (FORM-03).
- **Grade:** 1 plain
- **Rhea's default:** With no declaration on file, treat the business as skala on KPiR, because the act makes that the default. Liniowy or ryczałt applies only when the owner has declared it, with the date and channel of the statement (CEIDG or the tax office); Rhea never infers it.
- **Pack today:** missing. `company` has `legal_form` and `owner` but no tax-form field and no rule.
- **Verdict:** missing
- **Deadline:** see FORM-04
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU19910800350 ; https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf

### FORM-02 — Who may not choose ryczałt or the flat 19%?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - Ryczałt act art. 8 ust. 1: „Opodatkowania w formie ryczałtu od przychodów ewidencjonowanych (...) nie stosuje się do podatników: 1) opłacających podatek w formie karty podatkowej (...); 2) korzystających, na podstawie odrębnych przepisów, z okresowego zwolnienia od podatku dochodowego; 3) osiągających w całości lub w części przychody z tytułu: a) prowadzenia aptek, (...) c) działalności w zakresie kupna i sprzedaży wartości dewizowych, (...) f) działalności w zakresie handlu częściami i akcesoriami do pojazdów mechanicznych; 4) wytwarzających wyroby opodatkowane podatkiem akcyzowym (...)”. Pkt 5 excludes certain moves of a business between spouses.
  - Ryczałt act art. 8 ust. 2: the right to ryczałt is lost when the trader serves a former or current employer in the work done under employment in the previous or current year.
  - Ryczałt act art. 6 ust. 4: available if prior-year revenue was „nieprzekraczającej 2 000 000 euro”, and for businesses starting in the year „bez względu na wysokość przychodów”.
  - PIT act art. 9a ust. 3 (flat tax): the right is lost for services to a former or current employer matching work the trader „wykonywał lub wykonuje w roku podatkowym” under employment. The former pkt 1 (previous year) is repealed.
- **Grade:** 1 plain
- **Rhea's default:** Before accepting a declared ryczałt or liniowy, ask about these exclusions (PKD of the activity, a former-employer client, the prior-year revenue against 2 000 000 euro). Record the answers as facts, so that a later revenue line from an excluded activity raises a task.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** none of its own (tested when the form is chosen and at each revenue)
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf

### FORM-03 — Can a new business still choose the tax card (karta podatkowa)?
- **Applies to:** sole trader ryczałt (as a family of flat forms)
- **The law:** Ryczałt act art. 25 ust. 1 pkt 1: tax-card taxation applies only if the taxpayers „kontynuują stosowanie opodatkowania w tej formie po dniu 31 grudnia 2021 r. oraz nie zrezygnowali z opodatkowania w tej formie po dniu 31 grudnia 2021 r. i nie utracili prawa do opodatkowania w tej formie po tym dniu”.
- **Grade:** 1 plain
- **Rhea's default:** Do not offer karta. A first-phase user starting now cannot have it, and an existing karta user keeps no books, so Rhea has nothing to compute.
- **Pack today:** nothing, which is correct
- **Verdict:** right
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf

### FORM-04 — By when must the form be chosen, and how is it changed for next year?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - PIT act art. 9a ust. 2 (flat tax statement): „do 20. dnia miesiąca następującego po miesiącu, w którym został osiągnięty pierwszy przychód z tego tytułu w roku podatkowym, albo do końca roku podatkowego, jeżeli pierwszy taki przychód został osiągnięty w grudniu tego roku podatkowego”.
  - Art. 9a ust. 2b: „Dokonany wybór (...) dotyczy również lat następnych, chyba że w kolejnych latach podatnik w terminie określonym w ust. 2 zawiadomi (...) o rezygnacji”.
  - Art. 9a ust. 2a allows a change of mind before the deadline. Art. 9a ust. 2c allows the statement to be filed through CEIDG.
  - Ryczałt act art. 9 ust. 1, 1b and 1c set the same deadline and the same rollover, also through CEIDG. Art. 7 ust. 1: for a business starting in the year, ryczałt applies „od dnia uzyskania pierwszego przychodu”.
- **Grade:** 1 plain
- **Rhea's default:** Model the choice as a dated fact effective for a tax year that carries over until changed. On the first revenue of a year, open a task due on the 20th of the following month asking the owner to confirm or change the form, because that is the only window.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** 20th of the month after the month of the first revenue in the year; 31 December if the first revenue falls in December.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf

### FORM-05 — Which record must the business keep: KPiR, the ryczałt revenue record, or full books?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - PIT act art. 24a ust. 1: natural persons „są obowiązane prowadzić podatkową księgę przychodów i rozchodów (...) albo księgi rachunkowe”.
  - Art. 24a ust. 3: „Obowiązek prowadzenia [podatkowej księgi przychodów i rozchodów] nie dotyczy osób, które: 1) opłacają podatek dochodowy w formach zryczałtowanych”.
  - Ryczałt act art. 15 ust. 1: ryczałt taxpayers keep „ewidencję przychodów” and „wykaz środków trwałych oraz wartości niematerialnych i prawnych”.
  - Voluntary full books: PIT art. 24a ust. 5 (the trader „informuje o prowadzeniu ksiąg rachunkowych w zeznaniu”); accounting act art. 2 ust. 2 (notification „przed rozpoczęciem roku obrotowego”).
  - Compulsory full books: KPIR-11.
- **Grade:** 1 plain
- **Rhea's default:** Derive the form of books from the declared tax form and the threshold test: skala or liniowy means KPiR, ryczałt means ewidencja przychodów. Full books apply only when compulsory or declared. The current full-books postings should be gated so a KPiR trader does not mistake them for the statutory record.
- **Pack today:** missing. The pack posts every company to the wzorcowy plan kont (pl-post-sales-invoice, pl-post-purchase-invoice) whatever its form; that is harmless as management books but is not a KPiR or an ewidencja.
- **Verdict:** missing
- **Deadline:** from the first day of business, or 1 January
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1994/591/text/U/D19940591Lj.pdf

### FORM-06 — Does the owner want to be taxed on cash received (metoda kasowa) instead of on invoices issued?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - Who may choose. PIT act art. 14c ust. 1 pkt 1: prior-year revenue „nie przekroczyły kwoty 2 000 000 zł”. This was raised from 1 000 000 zł by Dz.U. 2025 poz. 1858 with effect from 1 January 2026.
  - Deadline. Art. 14c ust. 1 pkt 3: statement „do dnia 20 lutego roku podatkowego, a w przypadku podatnika rozpoczynającego prowadzenie działalności gospodarczej w trakcie roku podatkowego – w terminie do 20. dnia miesiąca następującego po miesiącu rozpoczęcia działalności”.
  - Revenue date. Art. 14c ust. 2: „za datę powstania przychodu uważa się dzień uregulowania należności, niepóźniejszy niż dzień: 1) upływu 2 lat, licząc od dnia wystawienia faktury”.
  - Scope. Art. 14c ust. 4: only invoiced B2B revenue. Art. 22 ust. 4a: costs from B2B transactions fall when paid.
  - Invoice register. Art. 24a ust. 1da requires one: „datę wystawienia faktury, numer faktury, kwotę należności wynikającą z faktury oraz datę uregulowania należności”.
  - Ryczałt. Ryczałt act art. 6 ust. 1g allows the same method; art. 15 ust. 11a requires the same register.
- **Grade:** 1 plain
- **Rhea's default:** Off unless declared. When declared, the matching engine's settlements are exactly the payment dates the method needs, so revenue and cost dates follow the settlement facts. The 2-year backstop is a dated rule.
- **Pack today:** missing. Settlements (pl-record-incoming-settlement) already carry the payment date, but no rule turns them into PIT revenue.
- **Verdict:** missing
- **Deadline:** 20 February of the tax year, or the 20th of the month after the business starts
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2025/1858/text/O/D20251858.pdf

### KPIR-01 — Which KPiR rules apply in 2026?
- **Applies to:** sole trader KPiR
- **The law:**
  - Rozporządzenie Ministra Finansów i Gospodarki z 6 września 2025 r. w sprawie prowadzenia podatkowej księgi przychodów i rozchodów (Dz.U. 2025 poz. 1299), § 25: „Rozporządzenie wchodzi w życie z dniem 1 stycznia 2026 r.” It replaces the 2019 regulation (Dz.U. 2019 poz. 2544).
  - The model book has 19 columns. Col. 3 holds the „numer identyfikujący fakturę w Krajowym Systemie e-Faktur” and col. 5 the counterparty's tax identifier, which „obejmuje również kod kraju nadania identyfikatora”. The 2019 model had 17 columns.
  - § 4 ust. 1: „Podatnik jest obowiązany prowadzić księgę rzetelnie i w sposób niewadliwy”. § 4 ust. 4 pkt 1 tolerates unrecorded or misrecorded revenue up to 0.5% of the year's revenue.
- **Grade:** 1 plain
- **Rhea's default:** Model the KPiR on the 2026 template exactly: 19 columns, with KSeF number and counterparty tax ID per line. The pack already holds `ksef_ref` on purchase invoices and KSeF numbers on submissions, so the new columns cost nothing.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** in force from 1 January 2026
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf ; https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20250001299 ; https://api.sejm.gov.pl/eli/acts/DU/2019/2544/text/O/D20192544.pdf

### KPIR-02 — Must the KPiR be kept in software and sent to the tax office, and from which year?
- **Applies to:** sole trader KPiR | sole trader ryczałt (for the ewidencja, see RYCZ-01)
- **The law:**
  - PIT act art. 24a ust. 1e, as worded from 1 July 2026 (Dz.U. 2026 poz. 779). Those keeping KPiR, full books or the fixed-asset register „są obowiązane prowadzić te księgi i tę ewidencję przy użyciu programów komputerowych oraz przesyłać je właściwemu naczelnikowi urzędu skarbowego po zakończeniu roku podatkowego w terminie: 1) do dnia upływu terminu złożenia zeznania, o którym mowa w art. 45 ust. 1 – w przypadku podatkowej księgi przychodów i rozchodów oraz ewidencji środków trwałych (...) 2) do dnia 31 lipca – w przypadku ksiąg rachunkowych”.
  - Polski Ład (Dz.U. 2021 poz. 2105) art. 66 ust. 1: the books are kept in software „i przesyłane po raz pierwszy za rok podatkowy rozpoczynający się po dniu: 1) 31 grudnia 2025 r. – w przypadku podmiotów, które są obowiązane przesyłać ewidencję, o której mowa w art. 109 ust. 3 [ustawy o VAT] (...) 2) 31 grudnia 2026 r. – w przypadku podmiotów innych niż określone w pkt 1”.
  - KAS names the files JPK_PKPIR and JPK_ST (for KPiR) and JPK_EWP and JPK_ST (for ryczałt), sent for the first time in 2027.
  - The fixed-asset register for KPiR gains the extra data in Dz.U. 2025 poz. 1311 § 3: KSeF number of a sale invoice, document of acceptance or removal, type of acquisition document, date of production.
- **Grade:** 1 plain
- **Rhea's default:** Every KPiR trader keeps the book in Rhea from day one. A VAT-registered trader gets JPK_PKPIR plus JPK_ST for 2026 by 30 April 2027; a VAT-exempt trader starts with 2027. JPK_PKPIR should be a market writer (markets/pl) like JPK_V7M.
- **Pack today:** missing (the writer seam exists, with JPK_V7M as precedent)
- **Verdict:** missing
- **Deadline:** VAT payers: the 2026 book by 30 April 2027. Others: the 2027 book by 30 April 2028. The deadline is the art. 45 ust. 1 return deadline.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/779/text/O/D20260779.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2021/2105/text/U/D20212105Lj.pdf ; https://www.gov.pl/web/kas/nowe-zasady-dokumentacji-ksiegowej-w-podatku-pit-od-2026-r-wazne-informacje ; https://api.sejm.gov.pl/eli/acts/DU/2025/1311/text/O/D20251311.pdf

### KPIR-03 — On the first day, what must be opened: the book and an opening stocktake?
- **Applies to:** sole trader KPiR
- **The law:**
  - KPiR regulation § 19: „Podatnik jest obowiązany założyć księgę (...) na dzień 1 stycznia roku podatkowego lub na dzień rozpoczęcia działalności w trakcie roku podatkowego.”
  - § 20 ust. 1: a stocktake („spis z natury”) is required „na dzień 1 stycznia, na koniec każdego roku podatkowego, na dzień rozpoczęcia działalności w ciągu roku podatkowego, a także w razie utraty w ciągu roku podatkowego prawa do opłacania podatku na podstawie ustawy (...) o zryczałtowanym podatku dochodowym”.
  - § 20 ust. 2: the previous 31 December stocktake serves as the opening one.
- **Grade:** 1 plain
- **Rhea's default:** On the day a KPiR business starts, open the book and record an opening stocktake. Its value may be zero for a service business; that zero is a fact the owner states, not a default Rhea assumes.
- **Pack today:** missing. The `stocktake` type exists (date, comment, status, value) without lines or rules.
- **Verdict:** missing
- **Deadline:** day of starting the business, or 1 January
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf

### KPIR-04 — Which cost method: costs on the invoice date (uproszczona, called "kasowa") or matched to the revenue's year (memoriałowa)?
- **Applies to:** sole trader KPiR
- **The law:**
  - PIT act art. 22 ust. 4: „Koszty uzyskania przychodów są potrącane tylko w tym roku podatkowym, w którym zostały poniesione”.
  - Art. 22 ust. 6: the full-books rules of ust. 5–5c apply to KPiR „pod warunkiem że stale w każdym roku podatkowym księgi te będą prowadzone w sposób umożliwiający wyodrębnienie kosztów uzyskania przychodów odnoszących się tylko do tego roku podatkowego”.
  - Art. 22 ust. 6b: „za dzień poniesienia kosztu (...) uważa się dzień wystawienia faktury (rachunku) lub innego dowodu stanowiącego podstawę do zaksięgowania (ujęcia) kosztu”.
  - Under the memoriał method, direct costs fall in the year of the revenue (ust. 5) and indirect costs on the day incurred (ust. 5c). KPiR col. 17 may carry „wydatki odnoszące się do przychodów miesiąca lub roku następnego”.
- **Grade:** 2 settled. Ust. 6b literally points at the ust. 6 (memoriał) taxpayers. The Minister of Finance's general interpretation No. DD2.8202.3.2022 of 23 February 2023 (Dz.Urz. MF poz. 21) states that for KPiR taxpayers the cost date is fixed by art. 22 ust. 6b, i.e. the invoice date, under either method.
- **Rhea's default:** The uproszczona method (cost on the date on the cost document), because no notification is needed and it is the common case. The memoriał method only when declared, and then kept for the whole year.
- **Pack today:** missing. `purchase_invoice.invoice_date` exists and could carry the cost date.
- **Verdict:** missing
- **Deadline:** chosen for the whole year; no statutory notice
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://www.gov.pl/attachment/3b31dae7-d219-442b-a8a1-e1e5bd567f00 ; https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf

### RYCZ-01 — What revenue record does a ryczałt trader keep, how fast, and is it sent as JPK_EWP?
- **Applies to:** sole trader ryczałt
- **The law:**
  - Ryczałt act art. 15 ust. 2: „Obowiązek prowadzenia ewidencji powstaje od dnia, od którego ma zastosowanie opodatkowanie w formie ryczałtu”.
  - Art. 15 ust. 12: the ewidencja and the wykaz must be kept in software and sent after the year „w terminie do dnia upływu terminu złożenia zeznania, o którym mowa w art. 21 ust. 2 pkt 2”. First year as KPIR-02 (Polski Ład art. 66 ust. 1).
  - The regulation of 6 September 2025 (Dz.U. 2025 poz. 1294) applies from 1 January 2026 (§ 12). § 4 ust. 1: entries are made from invoices, or from „wystawionego na koniec dnia dowodu wewnętrznego (...) z podziałem na przychody objęte poszczególnymi stawkami ryczałtu”. § 6: entries „najpóźniej do 20. dnia każdego miesiąca za miesiąc poprzedni”.
  - The model record has columns for the KSeF number, the counterparty tax ID and nine rates (17, 15, 14, 12.5, 12, 10, 8.5, 5.5, 3 %).
  - Penalty for a missing or defective record, art. 17 ust. 2: five times the rate, „nie może być wyższy niż 75 % przychodu”.
- **Grade:** 1 plain
- **Rhea's default:** Keep the ewidencja as a projection of sales invoices and day reports, split by ryczałt rate, filled by the 20th, and render JPK_EWP plus JPK_ST through the writer seam.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** entries by the 20th of the next month. JPK_EWP: VAT payers for 2026 by 30 April 2027; others for 2027 by 30 April 2028.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1294/text/O/D20251294.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf ; https://www.gov.pl/web/kas/nowe-zasady-dokumentacji-ksiegowej-w-podatku-pit-od-2026-r-wazne-informacje

### RYCZ-02 — Which ryczałt rate applies to this kind of sale?
- **Applies to:** sole trader ryczałt
- **The law:**
  - Ryczałt act art. 12 ust. 1 sets the rates:
    - 17% „przychodów osiąganych w zakresie wolnych zawodów”, as defined in art. 4 ust. 1 pkt 11 (a closed list of professions, performed personally);
    - 15%: listed services, including legal, accounting, advertising and management consulting;
    - 14%: health care, architectural and engineering services, specialised design;
    - 12%: listed IT services;
    - 10%: own-account real estate trading;
    - 8.5% up to 100 000 zł and 12.5% above it: private rental and some listed services;
    - 8.5%: other services;
    - 5.5%: production, construction, freight above 2 t;
    - 3%: trade and gastronomy;
    - 2%: art. 6 ust. 1d products.
  - Art. 12 ust. 2: „pobiera się bez pomniejszania przychodu o koszty uzyskania”.
  - Art. 12 ust. 3: where several rates apply and the record does not separate them, 8.5% applies, or the highest applicable rate if a higher-rate activity is also present.
- **Grade:** 1 plain. Rates are tied to PKWiU groupings, and placing a particular service in the right grouping is a factual call (not researched here).
- **Rhea's default:** Every sale line carries a declared activity and rate. A line without one raises a task rather than defaulting, because the art. 12 ust. 3 fallback is punitive.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf

### RYCZ-03 — Besides the revenue record, what must a ryczałt trader keep?
- **Applies to:** sole trader ryczałt
- **The law:** Ryczałt act art. 15 ust. 1: taxpayers „są obowiązani posiadać i przechowywać dowody zakupu towarów, prowadzić wykaz środków trwałych oraz wartości niematerialnych i prawnych oraz, odrębnie za każdy rok podatkowy, ewidencję przychodów”. Regulation 2025 poz. 1294 § 8 ust. 3: entries in the wykaz are made „najpóźniej w miesiącu przyjęcia ich do używania”. § 8 ust. 1 lists the contents: lp., KSeF number of a sale invoice, documents, dates, KŚT symbol, initial value, rate, updated value, liquidation or sale.
- **Grade:** 1 plain
- **Rhea's default:** Keep purchase invoices and a fixed-asset list for ryczałt traders even though no cost is deducted. The list matters for a later switch of form and for the 10% rate on sales of listed assets (art. 12 ust. 10).
- **Pack today:** missing. `fixed_asset` type exists, with no rules.
- **Verdict:** missing
- **Deadline:** the wykaz by the month the asset is put into use
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2025/1294/text/O/D20251294.pdf

### PIT-01 — On which date is a sale income for PIT, and how does that differ from the VAT tax point?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - PIT act art. 14 ust. 1c: „Za datę powstania przychodu (...) uważa się (...) dzień wydania rzeczy, zbycia prawa majątkowego lub wykonania usługi, albo częściowego wykonania usługi, niepóźniej niż dzień: 1) wystawienia faktury albo 2) uregulowania należności.”
  - Art. 14 ust. 1e: for services settled in periods, the last day of the period, „nierzadziej niż raz w roku”.
  - Art. 14 ust. 1i: otherwise, the day of payment.
  - Art. 14 ust. 3 pkt 1 excludes „pobranych wpłat lub zarachowanych należności na poczet dostaw towarów i usług, które zostaną wykonane w następnych okresach sprawozdawczych”.
  - Ryczałt follows the same rules (ryczałt act art. 6 ust. 1 refers to art. 14).
  - VAT act art. 19a ust. 1: „Obowiązek podatkowy powstaje z chwilą dokonania dostawy towarów lub wykonania usługi, z zastrzeżeniem ust. 1a, 1b, 5 i 7–11”. Art. 19a ust. 8: an advance payment creates the obligation „z chwilą jej otrzymania w odniesieniu do otrzymanej kwoty”.
  - So a VAT advance is a tax point, but not PIT revenue, and the special VAT tax points in ust. 5 have no PIT twin.
- **Grade:** 1 plain
- **Rhea's default:** PIT revenue date = the earliest of the supply or performance date, the invoice date and the payment date. Advances are excluded until performance. Under the cash method the payment date applies (FORM-06).
- **Pack today:** missing. `sales_invoice` has only `issue_date` (no date of supply), and pl-book-sales-invoice derives only `vat_period` from the issue month. A supply-date field is needed before any PIT rule can be written.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2004/535/text/U/D20040535Lj.pdf

### PIT-02 — How much of a sale is revenue, and what else counts as business revenue?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - PIT act art. 14 ust. 1: revenue is „kwoty należne, choćby nie zostały faktycznie otrzymane, po wyłączeniu wartości zwróconych towarów, udzielonych bonifikat i skont”. For VAT payers it is „przychód pomniejszony o należny podatek od towarów i usług”.
  - Art. 14 ust. 2 adds business revenue: pkt 1, sales of fixed assets and of equipment above 1500 zł; pkt 4, „otrzymane kary umowne”; pkt 5, „odsetki od środków pieniężnych na rachunkach (...) prowadzonych w związku z wykonywaną działalnością gospodarczą”.
  - Art. 14 ust. 1m: corrections not due to error are made „w okresie rozliczeniowym, w którym została wystawiona faktura korygująca”.
  - KPiR col. 10 is for „pozostałych przychodów”.
- **Grade:** 1 plain
- **Rhea's default:** Revenue is the net amount of the invoice (gross for a VAT-exempt trader). Interest credited on the business account is revenue (KPiR col. 10). A correcting invoice changes revenue in the month it is issued.
- **Pack today:** missing. `bank_operation.kind` has no interest kind, and correction invoices are a CORPUS candidate only.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf

### PIT-03 — An invoice in euro: which exchange rate turns it into złoty for PIT?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** PIT act art. 11a ust. 1: „Przychody w walutach obcych przelicza się na złote według kursu średniego walut obcych ogłaszanego przez Narodowy Bank Polski z ostatniego dnia roboczego poprzedzającego dzień uzyskania przychodu.” Ust. 2 applies the same rule to costs, as of the day the cost is incurred. KPiR regulation § 6 ust. 4: the converted amount must be written „w wolnych polach dowodu lub w załączniku do dowodu”.
- **Grade:** 1 plain
- **Rhea's default:** Convert at the NBP table A mid rate of the working day before the PIT revenue or cost date (not the invoice date when the two differ), and keep the rate as provenance on the line.
- **Pack today:** missing. `fx_rate` is a required type; no PIT conversion rule.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf

### RYCZ-04 — For a ryczałt trader, in which month does a sale count?
- **Applies to:** sole trader ryczałt
- **The law:** Ryczałt act art. 6 ust. 1: ryczałt is charged on business revenue „o których mowa w art. 7a ust. 4 lub art. 14 ustawy o podatku dochodowym”. So the PIT revenue date of art. 14 ust. 1c applies (PIT-01). Art. 6 ust. 1g: „podatnik może wybrać metodę kasową rozliczania przychodów zgodnie z art. 14c ust. 1”. The model ewidencja allows the payment date in „Uwagi” for traders paying quarterly, who must watch the 200 000 euro limit of receipts.
- **Grade:** 1 plain
- **Rhea's default:** Same date rule as PIT-01. Keep the receipt date alongside it, because the quarterly-payment limit (RYCZ-05) counts receipts, not amounts due.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2025/1294/text/O/D20251294.pdf

### KPIR-05 — Which document may support a KPiR entry, and when may the owner write one themselves?
- **Applies to:** sole trader KPiR
- **The law:**
  - KPiR regulation § 6 ust. 3: entries are based on invoices, on art. 22p correction documents, or on other documents meeting minimum data.
  - § 8 ust. 2: internal documents („dowody wewnętrzne”) „mogą dotyczyć wyłącznie” a closed list. It includes pkt 4, travel diets of the owner and staff; pkt 6, „wydatków związanych z opłatami za czynsz, energię elektryczną, telefon, wodę, gaz i centralne ogrzewanie, w części przypadającej na działalność gospodarczą”; pkt 7, court and notarial fees; pkt 8, parking tickets.
  - § 9 ust. 1: goods received before the invoice are entered from a signed description, later completed with the invoice's KSeF number.
  - § 7 pkt 2: „dowody opłat pocztowych i bankowych” are valid documents.
- **Grade:** 1 plain
- **Rhea's default:** Accept KSeF invoices, bank statements (for bank fees) and the § 8 list. An owner-made internal document outside that list is refused with the reason.
- **Pack today:** missing. Inbound KSeF invoices and bank lines exist as facts; no KPiR entry rule.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf

### KPIR-06 — By when must a KPiR entry be made?
- **Applies to:** sole trader KPiR
- **The law:**
  - Revenue. KPiR regulation § 11 ust. 2: entries from issued invoices and fiscal reports are made „najpóźniej do 20. dnia każdego miesiąca następującego po miesiącu uzyskania tych przychodów”. Under the 2019 regulation (§ 18 ust. 2) this was „jeden raz dziennie”.
  - Internal end-of-day documents. § 11 ust. 3 pkt 1: „raz dziennie po zakończeniu dnia, nie później niż przed rozpoczęciem działalności w dniu następnym”.
  - Costs. § 10 and § 15 only refer to the act; the 2019 daily rule (§ 16) is gone.
  - Late cost documents. MF general interpretation DD2.8202.3.2022: a cost document received late „może być zaksięgowany (ujęty) w PKPiR w miesiącu, w którym został otrzymany przez podatnika (...) do końca roku podatkowego”, with the cost date unchanged in col. 2.
- **Grade:** 2 settled (late cost documents: MF general interpretation of 23 February 2023). Revenue timing is plain.
- **Rhea's default:** Post revenue lines as invoices are booked, and flag any month not complete by the 20th. A late cost invoice is entered in the month it arrives with its own date in col. 2, which avoids correcting closed months.
- **Pack today:** missing (the period lock in packs/common/periods is the natural hook)
- **Verdict:** missing
- **Deadline:** invoice revenue by the 20th of the next month; internal day documents the same day
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2019/2544/text/O/D20192544.pdf ; https://www.gov.pl/attachment/3b31dae7-d219-442b-a8a1-e1e5bd567f00

### KPIR-07 — Which KPiR column does an expense go to?
- **Applies to:** sole trader KPiR
- **The law:** KPiR model, explanations:
  - item 10: col. 12 is for „zakupu materiałów i towarów handlowych według cen zakupu”;
  - item 11: col. 13 is for „kosztów ubocznych związanych z zakupem”;
  - item 12: col. 14 is for gross wages, including civil-law contracts;
  - item 13: col. 15 is for „pozostałych wydatków (...) z wyjątkiem wydatków, których zgodnie z ustawą o podatku dochodowym nie uznaje się za koszty”, including rent, energy, fuel, depreciation and employer social contributions;
  - item 17: col. 18 holds R&D costs under art. 26e.
  Cols 12–13 feed the year-end stocktake formula (item 22).
- **Grade:** 1 plain
- **Rhea's default:** Classify each purchase line as goods or materials (col. 12), side cost (col. 13) or other (col. 15) from the supplier and line content, through a rule the owner can correct. The split matters only where stock is held, so services default to col. 15.
- **Pack today:** missing. pl-post-purchase-invoice posts everything to account 300.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf

### KPIR-08 — Does the trader record purchases net or gross?
- **Applies to:** sole trader KPiR | VAT-exempt | all VAT payers
- **The law:**
  - KPiR regulation § 2 pkt 2: purchase price means the price „pomniejszoną o podatek od towarów i usług, podlegający odliczeniu zgodnie z odrębnymi przepisami”.
  - PIT act art. 23 ust. 1 pkt 43 lit. a: input VAT is a cost „jeżeli podatnik zwolniony jest od podatku od towarów i usług”, or to the extent it is not deductible.
  - Revenue for VAT payers is net (art. 14 ust. 1).
  - KPiR explanation 15: after losing the art. 113 exemption, costs are corrected „o tę część naliczonego podatku od towarów i usług, którą odlicza od podatku należnego”.
- **Grade:** 1 plain
- **Rhea's default:** A VAT-exempt trader records purchases gross, with no input-VAT line. A VAT payer records net, plus any non-deductible part. Exemption status is a declared, dated fact.
- **Pack today:** wrong for VAT-exempt traders. pl-post-purchase-invoice always posts the VAT to 221 „VAT naliczony”; for an art. 113 trader that VAT is part of the cost. This overlaps the VAT section.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf

### KPIR-09 — Sales to people without an invoice: how do they get into the book?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - KPiR regulation § 11 ust. 1: entries come from invoices, from fiscal reports, or „wystawionego na koniec dnia dowodu wewnętrznego (...) albo prowadzonego zestawienia sprzedaży (ewidencji sprzedaży)”.
  - § 17 ust. 1: the zestawienie must hold the „datę uzyskania przychodu nieudokumentowanego fakturą oraz wysokość tego przychodu”. § 17 ust. 3 lifts the duty for cash-register users.
  - § 12: cash-register traders enter daily or monthly fiscal reports, and § 12 ust. 3 forbids double-entering invoices issued for receipts.
  - Ryczałt: regulation 2025 poz. 1294 § 4 says the same, split by rate.
- **Grade:** 1 plain
- **Rhea's default:** Take daily fiscal report totals (or an owner day-total) as one revenue line per day. Exclude invoices that document an earlier receipt.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** daily (§ 18), or the 20th of the next month for monthly fiscal reports
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2025/1294/text/O/D20251294.pdf

### PIT-04 — Which expenses are never deductible, whatever the receipt says?
- **Applies to:** sole trader KPiR
- **The law:** PIT act art. 22 ust. 1: costs are those „poniesione w celu osiągnięcia przychodów lub zachowania albo zabezpieczenia źródła przychodów, z wyjątkiem kosztów wymienionych w art. 23”. The items of art. 23 ust. 1 a sole trader meets first:
  - pkt 1 lit. b: buying fixed assets; these are depreciated instead;
  - pkt 10: „wartości własnej pracy podatnika, jego małżonka i małoletnich dzieci”. However, „kosztem uzyskania przychodów jest jednak wynagrodzenie małżonka podatnika i małoletnich dzieci” under employment and similar titles;
  - pkt 15: fines;
  - pkt 23: „kosztów reprezentacji, w szczególności poniesionych na usługi gastronomiczne, zakup żywności oraz napojów, w tym alkoholowych”;
  - pkt 43: deductible input VAT;
  - pkt 58: health contributions, except the liniowy share (PIT-14).
- **Grade:** 1 plain
- **Rhea's default:** Rules mark these lines as non-cost, with the article as provenance. Gastronomy invoices default to "representation, not a cost" until the owner says otherwise (for example, a business trip).
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf

### PIT-05 — An expense used both privately and for the business: how much counts?
- **Applies to:** sole trader KPiR
- **The law:**
  - PIT act art. 22 ust. 1 (the purpose test, see PIT-04).
  - KPiR regulation § 8 ust. 2 pkt 6: home utilities may be entered by an internal document „w części przypadającej na działalność gospodarczą; podstawą sporządzenia tego dowodu jest dokument obejmujący całość opłat na te cele”.
  - Art. 22f ust. 4: a partly used property is depreciated by the share of usable floor area.
  - Cars: PIT-07.
- **Grade:** 1 plain. Only the business share counts, and the text gives no key except floor area for property; the key is a fact the owner states.
- **Rhea's default:** Each mixed-use cost source (flat, phone, internet) carries an owner-declared business share with its basis (area, time), applied by rule and shown on every derived line. Bank operations marked private never become costs.
- **Pack today:** missing. `bank_operation.settlement_status` has `private` (CORPUS rule intent 8), but no rule excludes such lines from costs.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf

### PIT-06 — Paying a supplier: when does the way of paying cost the deduction (white list, cash over 15 000 zł, split payment)?
- **Applies to:** sole trader KPiR
- **The law:**
  - PIT act art. 22p ust. 1: no cost „w tej części, w jakiej płatność dotycząca transakcji określonej w art. 19 ustawy (...) Prawo przedsiębiorców: 1) została dokonana bez pośrednictwa rachunku płatniczego lub 2) została dokonana przelewem na rachunek inny niż zawarty na dzień zlecenia przelewu w wykazie podmiotów, o którym mowa w art. 96b ust. 1 ustawy o podatku od towarów i usług (...) lub 3) pomimo zawarcia na fakturze wyrazów „mechanizm podzielonej płatności” (...) została dokonana z pominięciem mechanizmu podzielonej płatności”.
  - Prawo przedsiębiorców art. 19 applies where „jednorazowa wartość transakcji, bez względu na liczbę wynikających z niej płatności, przekracza 15 000 zł”.
  - Art. 22p ust. 2: costs are reduced, or revenue increased, in the month of payment.
  - Art. 22p ust. 4 pkt 1: no consequence if a notice is filed „w terminie 7 dni od dnia zlecenia przelewu”. The same ust. 4 pkt 3 covers payment through split payment.
  - KPiR § 6 ust. 5: the reduction is entered with a minus sign.
- **Grade:** 1 plain
- **Rhea's default:** On an outgoing payment for a transaction above 15 000 zł, check the payee IBAN against the white list as of the order date. If it is absent, open a 7-day task to file the ZAW-NR notice; otherwise post the art. 22p reduction.
- **Pack today:** missing. Matching knows payee IBANs (pl-match-outgoing, company.iban) but has no white-list check.
- **Verdict:** missing
- **Deadline:** notice within 7 days of ordering the transfer
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2018/646/text/U/D20180646Lj.pdf

### PIT-07 — A passenger car also used privately: what share of its running costs is deductible?
- **Applies to:** sole trader KPiR
- **The law:**
  - PIT act art. 23 ust. 1 pkt 46a excludes „25 % poniesionych wydatków (...) z tytułu kosztów używania samochodu osobowego (...) jeżeli samochód osobowy jest wykorzystywany również do celów niezwiązanych z działalnością gospodarczą”.
  - Pkt 46: for a private car not among business assets, the costs „w wysokości 20 % stanowią jednak koszty uzyskania przychodów”.
  - Ust. 5f: „W przypadku nieprowadzenia przez podatnika ewidencji, o której mowa w art. 86a ust. 4 ustawy o podatku od towarów i usług, uznaje się, że samochód osobowy jest wykorzystywany również do celów niezwiązanych z działalnością”. Ust. 5g: this does not apply where the VAT act imposes no such record (with one exception).
  - Ust. 5a: non-deductible VAT forms part of the limited amount.
- **Grade:** 1 plain for the 75% / 20% / 100% split. Whether depreciation of a mixed-use car falls inside the 75% limit: unverified. Secondary sources describe the MF objaśnienia podatkowe of 9 April 2020 as limiting the 75% rule to running costs, but the primary PDF on podatki.gov.pl returned 404 (secondary — verify).
- **Rhea's default:** A business car is mixed-use (75% of running costs) unless the vehicle has a VAT-26 notice and a mileage log on file, in which case 100%. A private car used in the business gets 20%. The vehicle's use status is a declared, dated fact.
- **Pack today:** missing. `vehicle` (with ownership own/leasing/rental/private) and `mileage_entry` types exist; there is no use-status field and no rules.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://www.podatki.gov.pl/media/5973/objaśnienia-podatkowe-samochody-9-kwietnia-2020-r.pdf (404 on fetch)

### PIT-08 — A passenger car's price: what value limit caps depreciation, lease fees and insurance from 2026?
- **Applies to:** sole trader KPiR
- **The law:**
  - PIT act art. 23 ust. 1 pkt 4: depreciation above the car's value is not a cost, beyond:
    - „a) 225 000 zł – w przypadku samochodu osobowego będącego pojazdem elektrycznym (...) b) 150 000 zł – jeśli emisja CO2 silnika spalinowego (...) wynosi mniej niż 50 g na kilometr, c) 100 000 zł – jeśli emisja CO2 silnika spalinowego (...) jest równa lub wyższa niż 50 g na kilometr”.
  - Pkt 47a: lease and rental fees are cut in the same proportion. Pkt 47: insurance is cut by the 150 000 zł proportion.
  - Lit. b–c took effect on 1 January 2026 (Dz.U. 2021 poz. 2269 art. 41 pkt 4). Art. 30 of that act keeps the old rules only „do pojazdów wprowadzonych do ewidencji środków trwałych (...) przed dniem wejścia w życie”.
- **Grade:** 2 settled. For cars leased before 2026, the MF (deputy minister's answer to interpelacja no. 12000) and the Director of KIS (interpretation 0115-KDIT3.4011.157.2026.1.DP of 8 April 2026) hold that the new limits apply from 1 January 2026, because the transitional rule covers only depreciated assets. One taxpayer is reported to contest this on constitutional grounds; no court ruling was found (secondary — verify both documents).
- **Rhea's default:** The limit follows the car's CO2 emission from the vehicle record. It applies to leases from 2026 whatever the contract date, and the old 150 000 zł applies only to cars entered in the asset register before 2026, as the MF reads it.
- **Pack today:** missing. `vehicle` lacks CO2 emission, value and drive type, and `lease_contract` holds only VAT limits.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2021/2269/text/O/D20212269.pdf ; https://www.prawo.pl/podatki/stare-umowy-leasingu-a-limit-kosztow-od-2026-r,1542853.html (secondary)

### PIT-09 — Buying equipment: is it an expense now or a fixed asset?
- **Applies to:** sole trader KPiR | sole trader ryczałt (wykaz only)
- **The law:**
  - PIT act art. 22d ust. 1: taxpayers may skip depreciation of items „których wartość początkowa (...) nie przekracza 10 000 zł; wydatki poniesione na ich nabycie stanowią wówczas koszty uzyskania przychodów w miesiącu oddania ich do używania”.
  - Art. 22f ust. 3: for items up to 10 000 zł, one-off depreciation „w miesiącu oddania do używania (...) albo w miesiącu następnym” is also allowed.
  - Art. 22a ust. 1: fixed assets are those „o przewidywanym okresie używania dłuższym niż rok”.
  - Art. 22e: an item over 10 000 zł expensed as short-lived but used beyond a year must be moved to fixed assets, with interest.
  - Art. 14 ust. 2 pkt 1 lit. b: selling an expensed item over 1500 zł is business revenue.
- **Grade:** 1 plain
- **Rhea's default:** Purchases up to 10 000 zł net (gross for a VAT-exempt trader) are expensed in the month of use. Above that, they open a fixed-asset proposal for the owner to accept. The 10 000 zł line is tested on the initial value, not the invoice total.
- **Pack today:** missing. `fixed_asset` type exists, with no rules.
- **Verdict:** missing
- **Deadline:** the month the item is put into use
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf

### PIT-10 — What goes in the fixed-asset register, and by when?
- **Applies to:** sole trader KPiR
- **The law:**
  - PIT act art. 22n ust. 2: the KPiR trader's register contains „co najmniej” lp., purchase date, date of putting into use, acquisition document, description, KŚT symbol, initial value, rate, annual and cumulative write-offs, updated value and write-offs, improvements, and liquidation or sale.
  - Art. 22n ust. 4: entries „najpóźniej w miesiącu przekazania ich do używania. Późniejszy termin wprowadzenia uznaje się za ujawnienie środka trwałego”.
  - Art. 22n ust. 6: „W razie braku ewidencji (...) dokonywane odpisy amortyzacyjne nie stanowią kosztów uzyskania przychodów.”
  - From 2026 the register also holds the extra data of Dz.U. 2025 poz. 1311 § 3 and is sent as JPK_ST (KPIR-02).
- **Grade:** 1 plain
- **Rhea's default:** An accepted asset becomes a register entry dated in its month of use, holding every art. 22n field plus the § 3 extras. Acceptance is the owner's door, not an automatic rule.
- **Pack today:** missing. `fixed_asset` v1 lacks the date entered in the register, improvement value, the acquisition-document reference and the KSeF number of the sale.
- **Verdict:** missing
- **Deadline:** by the end of the month the asset is put into use
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2025/1311/text/O/D20251311.pdf

### PIT-11 — How is a fixed asset depreciated: classification, rate, method, first month?
- **Applies to:** sole trader KPiR
- **The law:**
  - PIT act art. 22h ust. 1 pkt 1: write-offs run „począwszy od pierwszego miesiąca następującego po miesiącu, w którym ten środek lub wartość wprowadzono do ewidencji”.
  - Art. 22h ust. 2: the method is chosen per asset „przed rozpoczęciem ich amortyzacji; wybraną metodę stosuje się do pełnego zamortyzowania”.
  - Art. 22h ust. 4: write-offs may be „w równych ratach co miesiąc albo w równych ratach co kwartał, albo jednorazowo na koniec roku podatkowego”.
  - Art. 22i ust. 1: rates come from the Wykaz stawek (annex 1). It gives 30 for „Zespoły komputerowe” and 20 for „Samochody osobowe”.
  - Art. 22k ust. 1: degressive depreciation for KŚT groups 3–6 and 8 machines and transport, excluding passenger cars, at a coefficient up to 2.0.
  - Classification follows the KŚT (Rozporządzenie Rady Ministrów z 3 października 2016 r., Dz.U. 2016 poz. 1864, in force).
- **Grade:** 1 plain
- **Rhea's default:** Straight-line, monthly, at the annex rate, starting the month after entry. Degressive and individual rates only when the owner chooses them before the first write-off.
- **Pack today:** missing. `fixed_asset` has `kst`, `amortization_type` (linear/degressive/one_off/none), `rate` and `coefficient`; no rules (CORPUS rule intent 5).
- **Verdict:** missing
- **Deadline:** monthly with the advance, or at year end if the annual option is chosen
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU20160001864

### PIT-12 — May a new or small business write off an asset in one go?
- **Applies to:** sole trader KPiR
- **The law:**
  - De minimis one-off. PIT act art. 22k ust. 7: taxpayers „w roku podatkowym, w którym rozpoczęli prowadzenie działalności gospodarczej (...) oraz mali podatnicy” may write off at once assets in KŚT groups 3–8 „z wyłączeniem samochodów osobowych (...) do wysokości nieprzekraczającej w roku podatkowym równowartości kwoty 50 000 euro”.
    - Ust. 11 excludes those who ran a business (or whose spouse did, with joint property) in the two preceding years.
    - Ust. 12 converts at the NBP rate of the first working day of October of the previous year, rounded to 1000 zł: 213 000 zł for 2026 and 219 000 zł for 2027, from NBP mid rates of 4.2586 and 4.3770.
    - Ust. 10 makes it de minimis aid.
  - New assets. Art. 22k ust. 14: one-off for „fabrycznie nowych środków trwałych zaliczonych do grupy 3–6 i 8” up to 100 000 zł a year, subject to the ust. 15 value conditions.
  - Ust. 20: one basis per asset.
- **Grade:** 1 plain
- **Rhea's default:** Offer, never apply silently. The de minimis route uses up the owner's de minimis allowance, which Rhea cannot see, so the owner decides per asset.
- **Pack today:** missing (`amortization_type` has `one_off`, no rule)
- **Verdict:** missing
- **Deadline:** in the year the asset is entered in the register
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.nbp.pl/api/exchangerates/rates/a/eur/2025-10-01/ ; https://api.nbp.pl/api/exchangerates/rates/a/eur/2026-10-01/

### PIT-13 — The owner brings in a laptop or car they already owned, or works from home: what value, and is the home depreciable?
- **Applies to:** sole trader KPiR
- **The law:**
  - PIT act art. 22g ust. 1 pkt 1: on purchase, „cenę ich nabycia, a jeżeli były używane przez podatnika przed wprowadzeniem ich do ewidencji (...) i nie były wcześniej amortyzowane – cenę ich nabycia, niewyższą jednak od ich wartości rynkowej”.
  - Art. 22g ust. 8: if the price cannot be established, the value is the trader's own valuation „z uwzględnieniem cen rynkowych środków trwałych tego samego rodzaju z grudnia roku poprzedzającego rok założenia ewidencji”.
  - Art. 22c pkt 2: residential buildings and flats are not depreciable.
  - Art. 22f ust. 4: part-business property is depreciated by floor-area share.
- **Grade:** 1 plain
- **Rhea's default:** A private asset enters at its documented purchase price, capped at market value, or at the owner's dated valuation if there is no document. A flat or house is never depreciated, and the owner's home costs go through PIT-05.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** register by the month of putting into use
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf

### PIT-14 — The owner's own ZUS: are social and health contributions deducted, counted as costs, or neither, under each form?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - Social contributions. PIT act art. 26 ust. 1 pkt 2 lit. a deducts from income social contributions „zapłaconych w roku podatkowym bezpośrednio na własne ubezpieczenia emerytalne, rentowe, chorobowe oraz wypadkowe podatnika oraz osób z nim współpracujących”. Art. 26 ust. 13a allows this if they „nie zostały zaliczone do kosztów uzyskania przychodów”, so the trader may take them either as a cost or as a deduction, not both.
  - Health contributions on skala. Art. 23 ust. 1 pkt 58 excludes them from costs except „z tytułu pozarolniczej działalności gospodarczej opodatkowanej zgodnie z art. 30c”. Art. 26 has no health item, and art. 27b is repealed, so on skala they are neither a cost nor a deduction.
  - Health contributions on liniowy. Art. 30c ust. 2 pkt 2 allows a deduction, or a cost under art. 23 ust. 1 pkt 58, up to a combined cap that „nie może przekroczyć w roku podatkowym 8700 zł”, indexed yearly (ust. 2b–2c). M.P. 2025 poz. 1274 fixes the 2026 cap at 14 100 zł.
  - Ryczałt. Ryczałt act art. 11 ust. 1a: „Przychody (...) mogą być pomniejszone o 50 % składek na ubezpieczenie zdrowotne zapłaconych w roku podatkowym”. Art. 11 ust. 1 deducts social contributions through the PIT art. 26 ust. 1 pkt 2.
- **Grade:** 1 plain
- **Rhea's default:**
  - Social contributions: deduct from income (ryczałt: from revenue) when paid, but propose cost treatment when the year is heading for a loss, so the amount survives as a loss carry-forward. The owner decides.
  - Health contributions: none on skala; on liniowy, a deduction capped at the year's announced amount (14 100 zł in 2026); on ryczałt, 50% of contributions paid.
- **Pack today:** missing (`tax_declaration` kind `zus_dra` exists, no rules)
- **Verdict:** missing
- **Deadline:** counted when paid, monthly in the advance
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://eli.gov.pl/eli/MP/2025/1274/ogl/pol/pdf ; https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf

### PIT-15 — Is the owner a "small taxpayer" (mały podatnik) this year?
- **Applies to:** sole trader KPiR
- **The law:** PIT act art. 5a pkt 20: a small taxpayer is one whose „wartość przychodu ze sprzedaży (wraz z kwotą należnego podatku od towarów i usług) nie przekroczyła w poprzednim roku podatkowym (...) równowartości 2 000 000 euro; przeliczenia (...) według średniego kursu euro ogłaszanego przez Narodowy Bank Polski na pierwszy dzień roboczy października poprzedniego roku podatkowego, w zaokrągleniu do 1000 zł”. That gives 8 517 000 zł for status in 2026 (2025 sales) and 8 754 000 zł for 2027 (2026 sales), from NBP mid rates 4.2586 and 4.3770. Start-ups get the same advance and one-off options directly (art. 44 ust. 3g, art. 22k ust. 7).
- **Grade:** 1 plain
- **Rhea's default:** Compute the status each January from last year's gross sales and the fetched NBP rate, as a fact with provenance. Status gates quarterly advances (PIT-16) and the de minimis write-off (PIT-12).
- **Pack today:** missing (`fx_rate` exists)
- **Verdict:** missing
- **Deadline:** fixed at the start of each tax year
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.nbp.pl/api/exchangerates/rates/a/eur/2025-10-01/ ; https://api.nbp.pl/api/exchangerates/rates/a/eur/2026-10-01/

### PIT-16 — Are income-tax advances paid monthly or quarterly, and by when?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - PIT act art. 44 ust. 3: „Podatnicy (...) są obowiązani wpłacać zaliczki miesięczne”.
  - Art. 44 ust. 3g and ust. 3h (liniowy): small taxpayers and start-ups „mogą wpłacać zaliczki kwartalne”.
  - Art. 44 ust. 3i: the choice is reported „w zeznaniu (...) składanym za rok podatkowy, w którym stosowali kwartalny sposób”.
  - Art. 44 ust. 6: monthly advances „w terminie do 20 dnia każdego miesiąca za miesiąc poprzedni”, quarterly ones by the 20th after the quarter, and „Zaliczkę za ostatni miesiąc lub ostatni kwartał roku podatkowego podatnik wpłaca w terminie do 20 stycznia następnego roku podatkowego.”
  - Ryczałt act art. 21 ust. 1, 1a and 1c: the same deadlines and the same reporting in the return. Under art. 21 ust. 1b, quarterly payment is open only to start-ups and to those whose prior-year receipts „nie przekroczyły kwoty stanowiącej równowartość 200 000 euro”: 851 720 zł for 2026 and 875 400 zł for 2027.
- **Grade:** 1 plain
- **Rhea's default:** Monthly unless the owner declares quarterly and qualifies. Each month or quarter emits a dated tax_declaration (kind pit, or ryczałt) with its due date on the 20th, as the VAT settlement already does.
- **Pack today:** missing (`tax_declaration` kind `pit` exists, no rule; the month-end close and the `vat_settlement` pattern are reusable)
- **Verdict:** missing
- **Deadline:** 20th of the next month, or of the month after the quarter; December or Q4 by 20 January
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf

### PIT-17 — On the tax scale (skala), how is the monthly advance computed?
- **Applies to:** sole trader KPiR
- **The law:**
  - PIT act art. 44 ust. 2: the advance base for KPiR is „różnica pomiędzy wynikającym z tych ksiąg przychodem i kosztami jego uzyskania”, with a monthly stocktake only if one is taken.
  - Art. 44 ust. 3 pkt 1: the obligation starts „poczynając od miesiąca, w którym dochody te przekroczyły kwotę stanowiącą iloraz kwoty zmniejszającej podatek oraz najniższej stawki podatku”, i.e. 3600 zł ÷ 12% = 30 000 zł.
  - Pkt 2–3: the advance is the tax on year-to-date income „według zasad określonych w art. 26 i art. 27”, minus earlier advances.
  - Art. 27 ust. 1 scale: up to 120 000 zł, „12 % minus kwota zmniejszająca podatek 3600 zł”; above it, „10 800 zł + 32 % nadwyżki ponad 120 000 zł”.
- **Grade:** 1 plain
- **Rhea's default:** Compute the cumulative year-to-date tax each month from KPiR income less the art. 26 deductions, apply the scale and subtract earlier advances, in whole złoty as the Ordynacja rounds (rounding rule not researched here).
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** as PIT-16
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf

### PIT-18 — On the flat 19% (liniowy), how is the advance computed?
- **Applies to:** sole trader KPiR
- **The law:** PIT act art. 30c ust. 1: the tax „wynosi 19 % podstawy obliczenia podatku”, with no tax-free amount. Art. 30c ust. 2: the base may be reduced only by social contributions, health contributions up to the cap (PIT-14), IKZE payments and vocational-training donations. Art. 44 ust. 3f: advances equal „różnicy między podatkiem należnym od dochodu osiągniętego od początku roku, obliczonym zgodnie z art. 30c, a sumą należnych zaliczek za poprzednie miesiące”. Art. 6 ust. 8 pkt 1 lit. a: no joint filing.
- **Grade:** 1 plain
- **Rhea's default:** Same cumulative engine as PIT-17 with a flat 19% and the art. 30c ust. 2 deduction list. No tax-free amount and no reliefs.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** as PIT-16
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf

### RYCZ-05 — On ryczałt, how is the monthly payment computed?
- **Applies to:** sole trader ryczałt
- **The law:** Ryczałt act art. 21 ust. 1: „Podatnicy są obowiązani za każdy miesiąc obliczać ryczałt od przychodów ewidencjonowanych i wpłacać go (...) w terminie do dnia 20 następnego miesiąca”. Art. 12 ust. 2: the rate applies to revenue „bez pomniejszania przychodu o koszty uzyskania”. Art. 21 ust. 3: monthly or quarterly payments may already reflect the art. 11 deductions: 50% of health contributions, social contributions and losses. Art. 11 ust. 3: deductions are split across rates in proportion to revenue.
- **Grade:** 1 plain
- **Rhea's default:** Each month: revenue per rate, less deductions allocated pro rata to the rates, times each rate, summed. Like the VAT settlement, every position is a fold with provenance.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** 20th of the next month (quarterly: 20th after the quarter); December or Q4 by 20 January
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf

### RYCZ-06 — Can a ryczałt trader lose ryczałt in the middle of the year, and what happens then?
- **Applies to:** sole trader ryczałt
- **The law:**
  - Ryczałt act art. 8 ust. 2: on a sale to a former or current employer, the trader „traci w roku podatkowym prawo do opodatkowania w formie ryczałtu (...) i, poczynając od dnia uzyskania tego przychodu do końca roku podatkowego, opłaca podatek dochodowy na ogólnych zasadach”.
  - Art. 8 ust. 3 has the same effect on a first revenue from an excluded activity.
  - Art. 22 ust. 1: from that day, the trader must „zaprowadzić właściwe księgi”.
  - KPiR regulation § 20 ust. 1 requires a stocktake on that day.
  - The 2 000 000 euro limit (art. 6 ust. 4 pkt 1) tests the previous year, so exceeding it during the year ends ryczałt only from the next year.
- **Grade:** 1 plain
- **Rhea's default:** A revenue line from an excluded client or activity triggers a task. Confirming it switches the company to skala and KPiR from that date, with a stocktake. At year end, last year's revenue is tested against 2 000 000 euro for next year.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** from the day of the triggering revenue
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf

### PIT-19 — Can advances be paid as a fixed twelfth of an earlier year's tax (zaliczki uproszczone)?
- **Applies to:** sole trader KPiR
- **The law:** PIT act art. 44 ust. 6b: advances may be „1/12 kwoty obliczonej (...) od dochodu z pozarolniczej działalności gospodarczej wykazanego w zeznaniu (...) złożonym: 1) w roku podatkowym poprzedzającym dany rok podatkowy albo 2) (...) o dwa lata”. Art. 44 ust. 6e: this „nie mają zastosowania do podatników, którzy po raz pierwszy podjęli działalność w roku podatkowym albo w roku poprzedzającym rok podatkowy”.
- **Grade:** 1 plain
- **Rhea's default:** Not offered in the first two years. Later, offered as a declared option.
- **Pack today:** missing
- **Verdict:** later
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf

### PIT-20 — Selling a business asset, or a former business asset within six years: is it business revenue?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** PIT act art. 14 ust. 2 pkt 1 makes sales of fixed assets and of equipment over 1500 zł business revenue. Art. 10 ust. 2 pkt 3: this holds „nawet jeżeli przed zbyciem zostały wycofane z działalności gospodarczej, a między pierwszym dniem miesiąca następującego po miesiącu, w którym składniki majątku zostały wycofane z działalności i dniem ich odpłatnego zbycia, nie upłynęło 6 lat”. Art. 24 ust. 2 sets the income on sale. Ryczałt act art. 12 ust. 1 pkt 7 lit. f and ust. 10 set the rates for such sales (3% for movables; 10% for listed assets and rights).
- **Grade:** 1 plain
- **Rhea's default:** Keep the withdrawal date on every asset taken out of the business, and flag a sale within six years as business revenue.
- **Pack today:** missing (`fixed_asset` has `sale_date`, `liquidation_date`)
- **Verdict:** later
- **Deadline:** —
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf

### KPIR-10 — Year end: the closing stocktake and the year's income from the book
- **Applies to:** sole trader KPiR
- **The law:**
  - KPiR regulation § 20 ust. 1: a stocktake is required „na koniec każdego roku podatkowego”.
  - § 21 ust. 2: it lists the items, unit, quantity, unit price and value, and ends with „Spis zakończono na pozycji ...” and the owner's signature.
  - § 22 ust. 1: valuation is at purchase price or acquisition cost, or market price if lower. § 22 ust. 7: valuation is done „najpóźniej w terminie 14 dni od dnia zakończenia spisu z natury”. § 22 ust. 6: the value is entered in the book.
  - PIT act art. 24 ust. 2: income is revenue less costs, adjusted by the difference between closing and opening stocktakes.
  - Model explanation 22: costs = opening stocktake + cols 12 and 13 − closing stocktake + col. 16.
  - Several accounting-software publishers state that a service business without stock still records a stocktake of zero (secondary — verify). The text has no exemption.
- **Grade:** 1 plain
- **Rhea's default:** On closing the year, open a stocktake task dated 31 December. Accept a stated zero for service businesses, require item lines otherwise, and compute the year's income by the model formula.
- **Pack today:** missing. `stocktake` has `value` but no item lines and no rule; the year close is a door to build on `close_period`.
- **Verdict:** missing
- **Deadline:** stocktake as at 31 December; valuation within 14 days
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text/O/D20251299.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://www.infakt.pl/blog/?p=55358 (secondary)

### PIT-21 — The annual return: which form, and by when must it be filed and paid?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - PIT act art. 45 ust. 1 (skala): the return is filed „w terminie od dnia 15 lutego do dnia 30 kwietnia roku następującego po roku podatkowym”.
  - Art. 45 ust. 1a pkt 2: a separate return in the same window for income „opodatkowanych na zasadach określonych w art. 30c”.
  - Art. 45 ust. 4: the balance is paid „przed upływem terminu określonego na złożenie zeznania”.
  - Ryczałt act art. 21 ust. 2 pkt 2: the same window.
  - The forms are PIT-36 (skala), PIT-36L (liniowy) and PIT-28 (ryczałt). KAS states that business returns are pre-filled in Twój e-PIT but not accepted automatically.
  - Ryczałt act art. 21c changes on 1 January 2027 (Dz.U. 2026 poz. 1098): pre-filled returns move to „odrębnych przepisach” (not researched further).
- **Grade:** 1 plain
- **Rhea's default:** Produce the year's return figures as a draft `tax_declaration` (PIT-36, PIT-36L or PIT-28 by declared form) for the owner or accountant to file. It also records the quarterly-advance and cash-method choices that the return must report (art. 44 ust. 3i, art. 14 ust. 1j).
- **Pack today:** missing. `tax_declaration` has one generic `pit` kind.
- **Verdict:** missing
- **Deadline:** 15 February – 30 April of the next year; tax paid by 30 April
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf ; https://www.gov.pl/web/finanse/pit-36l ; https://www.gov.pl/web/finanse/kas-przypomina-twoj-e-pit-dla-przedsiebiorcow

### PIT-22 — A loss year: how is it carried forward?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** PIT act art. 9 ust. 3: the loss may reduce income from the same source „w najbliższych kolejno po sobie następujących pięciu latach podatkowych, z tym że kwota obniżenia w którymkolwiek z tych lat nie może przekroczyć 50 % wysokości tej straty”. Alternatively it may be used once „o kwotę nieprzekraczającą 5 000 000 zł”, with the rest within the five years at 50% per year. Liniowy uses the same rule (art. 30c ust. 2 refers to art. 9 ust. 3). Ryczałt act art. 11 ust. 1 lets a ryczałt trader deduct from revenue „stratę, o której mowa w art. 9 ust. 3–3b ustawy o podatku dochodowym”.
- **Grade:** 1 plain
- **Rhea's default:** Record each year's loss as a fact with its five-year window. Propose the deduction, defaulting to the 50% route; the 5 000 000 zł one-off is an owner choice.
- **Pack today:** missing
- **Verdict:** later (first used in year two)
- **Deadline:** in the annual returns of the next five years
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1998/930/text/U/D19980930Lj.pdf

### PIT-23 — Can the owner file jointly with a spouse?
- **Applies to:** sole trader KPiR
- **The law:**
  - PIT act art. 6 ust. 2: spouses in a marriage with joint property all year „mogą być (...) na wspólny wniosek wyrażony w zeznaniu podatkowym, opodatkowani łącznie od sumy swoich dochodów”.
  - Art. 6 ust. 8: joint filing „nie ma zastosowania, w przypadku gdy chociażby jeden z małżonków (...) stosuje przepisy: a) art. 30c lub b) ustawy o zryczałtowanym podatku dochodowym, z wyjątkiem art. 6 ust. 1a tej ustawy”. Art. 6 ust. 1a of the ryczałt act covers private rental.
- **Grade:** 1 plain
- **Rhea's default:** Offer joint filing in the PIT-36 draft only when the business is on skala and the owner says the spouse is not on liniowy or business ryczałt. Rhea cannot see the spouse's income, so the result is a hand-off, not a computation.
- **Pack today:** missing (`person` v2 has names and birth date; no marital facts)
- **Verdict:** missing
- **Deadline:** in the return (by 30 April)
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf

### KPIR-11 — Does this year's revenue force full books next year?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:**
  - Accounting act art. 2 ust. 1 pkt 2: natural persons are covered „jeżeli ich przychody netto ze sprzedaży towarów i produktów za poprzedni rok obrotowy wyniosły co najmniej równowartość w walucie polskiej 2 500 000 euro”. The threshold was raised from 2 000 000 euro by Dz.U. 2024 poz. 1863. Its art. 14 ust. 1 applies it „po raz pierwszy do roku obrotowego rozpoczynającego się po dniu 31 grudnia 2024 r.”
  - Accounting act art. 3 ust. 3: conversion „po średnim kursie ogłoszonym przez Narodowy Bank Polski, na pierwszy dzień roboczy października roku poprzedzającego rok obrotowy”. That gives 10 646 500 zł for the 2026 obligation (2025 revenue) and 10 942 500 zł for 2027 (2026 revenue), from NBP mid rates 4.2586 and 4.3770.
  - PIT act art. 24a ust. 4 imposes the same obligation for tax when „przychody, w rozumieniu art. 14, za poprzedni rok podatkowy wyniosły (...) co najmniej równowartość kwoty określonej w euro w przepisach o rachunkowości”. Ust. 6 uses the same conversion date.
  - Full books are sent as JPK by 31 July (art. 24a ust. 1e pkt 2, from 1 July 2026). Ryczałt traders keep ryczałt below 2 000 000 euro (RYCZ-06); full books would replace their ewidencja only on this test.
- **Grade:** 1 plain. The two acts measure different revenue (net sales vs PIT revenue). Each text decides its own obligation, so either one triggers the switch. No source disputing this was found.
- **Rhea's default:** At year end, test both measures against the fetched rate, and on either breach open a task that switches the company to full books from 1 January. The pack's existing full-books chart then becomes the statutory record.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** full books from 1 January of the year after the breach
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/1994/591/text/U/D19940591Lj.pdf ; https://api.sejm.gov.pl/eli/acts/DU/2024/1863/text/O/D20241863.pdf ; https://api.sejm.gov.pl/eli/acts/DU/1991/350/text/U/D19910350Lj.pdf ; https://api.nbp.pl/api/exchangerates/rates/a/eur/2025-10-01/ ; https://api.nbp.pl/api/exchangerates/rates/a/eur/2026-10-01/

## The owner and money

Ordered by when a business meets each point: founding and the first month, then every month as money moves, then the year end and after. Consolidated texts read from the Sejm ELI API (api.sejm.gov.pl / eli.gov.pl), state of law 16 October 2026; quotes are from those texts unless marked secondary.

### OWN-01 — Does a sole trader need a separate business bank account, and does it matter which kind?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** No provision found that obliges a sole trader to keep a separate business account. What decides the consequences: Prawo przedsiębiorców art. 19 (B2B payments over the limit go "za pośrednictwem rachunku płatniczego przedsiębiorcy" — any payment account, also a personal one); Prawo bankowe art. 49 ust. 1 pkt 1 and ust. 2 pkt 3 („rachunki rozliczeniowe … oraz prowadzone dla nich … rachunki VAT” — kept for, among others, „osób fizycznych prowadzących działalność zarobkową na własny rachunek”) against art. 49 ust. 1 pkt 3 / ust. 3 (personal „rachunki oszczędnościowo-rozliczeniowe” only for natural persons); art. 62a ust. 1: „Dla rachunku rozliczeniowego bank prowadzi rachunek VAT.”; ustawa o VAT art. 96b ust. 3 pkt 13: the white list shows only „numery rachunków rozliczeniowych, o których mowa w art. 49 ust. 1 pkt 1 ustawy … – Prawo bankowe … wskazanych w zgłoszeniu identyfikacyjnym lub zgłoszeniu aktualizacyjnym”.
- **Grade:** 1 plain — no obligation exists; the consequences of a personal account follow from the texts above: no VAT account (so no split payments can land), and the number never appears on the white list (it can still be reported — the NIP act's list of accounts, art. 5 ust. 4, has no „rozliczeniowy” qualifier). The government portal says the same: „Rachunek firmowy (a nie prywatny) będzie ci potrzebny jeśli: tylko numery firmowych rachunków bankowych są publikowane w Wykazie podatników VAT; dokonujesz transakcji z zastosowaniem mechanizmu podzielonej płatności” (biznes.gov.pl 00120).
- **Rhea's default:** accept a personal account, but when the company is an active VAT payer tell the owner once, with the two consequences (business customers paying over 15,000 zł to an unlisted account lose the cost unless they notify their tax office; split payments need a VAT account), and recommend a business (rozliczeniowy) account reported in CEIDG.
- **Pack today:** `bank_account` (name, number, currency, bank) has no kind (rozliczeniowy / osobisty / VAT); `company.iban` holds one number. Missing.
- **Verdict:** missing
- **Deadline:** none for opening; reporting see MON-01.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/38/text.pdf (Prawo bankowe, t.j. Dz.U. 2026 poz. 38); https://api.sejm.gov.pl/eli/acts/DU/2026/1263/text.pdf (VAT, t.j. Dz.U. 2026 poz. 1263); https://api.sejm.gov.pl/eli/acts/DU/2025/1480/text.pdf (Prawo przedsiębiorców, t.j. Dz.U. 2025 poz. 1480); https://www.biznes.gov.pl/pl/portal/00120

### MON-01 — When must the business tell the tax office which bank accounts it uses?
- **Applies to:** everyone
- **The law:** ustawa o zasadach ewidencji i identyfikacji podatników i płatników (t.j. Dz.U. 2026 poz. 151): art. 5 ust. 2b pkt 2 and ust. 2c — a KRS company files only the supplementary data, including „wykaz rachunków bankowych … z wyjątkiem rachunku VAT”, „w terminie 21 dni od dnia dokonania wpisu podmiotu do Krajowego Rejestru Sądowego”; art. 5 ust. 4 and 5a — a sole trader's data, including the list of accounts, go through the CEIDG form; art. 9 ust. 1 — changes „niepóźniej niż w terminie 7 dni od dnia, w którym nastąpiła zmiana danych”.
- **Grade:** 1 plain
- **Rhea's default:** when a bank account is registered or first appears on a statement, open a task to report it (CEIDG for a sole trader, the tax office for a company) and check the white list for the company's own number; the account is "reported" only when the person confirms.
- **Pack today:** missing (no rule or task; `bank_account` has no reported flag).
- **Verdict:** missing
- **Deadline:** 21 days after the KRS entry (company); 7 days after any change (art. 9 ust. 1).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/151/text.pdf

### OWN-02 — The founders of a sp. z o.o. pay in the share capital: what is the bank line, and what tax does the company owe on it?
- **Applies to:** sp. z o.o. full books
- **The law:** KSH art. 154 § 1: „Kapitał zakładowy spółki powinien wynosić co najmniej 5000 złotych.” (t.j. Dz.U. 2024 poz. 18; the 2026 amendment Dz.U. 2026 poz. 176 does not touch art. 154). Ustawa o PCC (t.j. Dz.U. 2026 poz. 191): art. 1 ust. 1 pkt 1 lit. k („umowy spółki”); art. 6 ust. 1 pkt 8 lit. a (base: „wartość kapitału zakładowego”); art. 6 ust. 9 (less the notary's fee with VAT and the court fee for the registration); art. 7 ust. 1 pkt 9: „od umowy spółki – 0,5 %”; art. 4 pkt 9 (taxpayer „na spółce”); art. 10 ust. 1 (declare and pay „w terminie 14 dni od dnia powstania obowiązku podatkowego”), art. 10 ust. 2: „Notariusze są płatnikami podatku od czynności cywilnoprawnych dokonywanych w formie aktu notarialnego.”
- **Grade:** 1 plain
- **Rhea's default:** the incoming capital is the company's equity (Wn 130 / Ma 801, or through a receivable from shareholders for capital not yet paid), never a customer payment; the PCC is a tax liability with a 14-day due date unless the deed was notarial and the notary collected it.
- **Pack today:** account 801 exists but no rule uses it; every incoming line posts Wn 130 / Ma 201 (pl-post-incoming-payment), so the capital lands on customer receivables. Wrong.
- **Verdict:** wrong
- **Deadline:** PCC 14 days from the agreement (art. 10 ust. 1) when no notary collected it.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2024/18/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/176/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/191/text.pdf

### ZUS-01 — A sole trader starts: what must he report to ZUS, and does he pay social contributions in the first six months?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** Prawo przedsiębiorców art. 18 ust. 1: a natural person starting a business for the first time, or again after „co najmniej 60 miesięcy od dnia jej ostatniego zawieszenia lub zakończenia”, and not working for a former employer in the same tasks in the current or previous year, „nie podlega obowiązkowym ubezpieczeniom społecznym przez okres 6 miesięcy od dnia podjęcia działalności gospodarczej”; ust. 2 — he may give it up by registering. Health insurance stays compulsory (ustawa o świadczeniach art. 66 ust. 1 pkt 1a). Registration: ustawa o sus art. 36 ust. 4 — „w terminie 7 dni od daty powstania obowiązku ubezpieczenia”. ZUS counts the six months as full calendar months: „Jeśli rozpoczniesz działalność pierwszego dnia miesiąca, to uwzględniasz ten miesiąc jako pierwszy z 6 miesięcy … Jeśli natomiast rozpoczniesz działalność w trakcie miesiąca, to okres ulgi liczysz od kolejnego miesiąca kalendarzowego.” biznes.gov.pl: health-only registration on ZUS ZZA with title code starting 05 40.
- **Grade:** 2 settled — the act says "6 months from the day of starting"; ZUS's published reading (full calendar months, the starting month's remainder added) settled the count.
- **Rhea's default:** the company says its start date and whether the conditions hold (first business / 60 months / no former employer); Rhea computes the relief's end the ZUS way, schedules the switch to preferential contributions, and opens the 7-day registration task.
- **Pack today:** missing — `company` has no start date, no ZUS status; no ZUS object, no schedule.
- **Verdict:** missing
- **Deadline:** registration within 7 days of starting (art. 36 ust. 4).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1480/text.pdf; https://eli.gov.pl/api/acts/DU/2026/199/text/T/D20260199L.pdf (ustawa o sus, t.j. Dz.U. 2026 poz. 199); https://www.zus.pl/-/ulga-na-start-preferencyjna-podstawa-dzialalnosc-nieewidencjonowana-jakie-sa-warunki-uprawnienia-i-skutk-1; https://www.biznes.gov.pl/pl/portal/00285

### ZUS-02 — The sole shareholder of a one-person sp. z o.o.: does he pay ZUS, and at what rate?
- **Applies to:** sp. z o.o. full books (one-shareholder companies)
- **The law:** ustawa o sus art. 8 ust. 6 pkt 4: „Za osobę prowadzącą pozarolniczą działalność uważa się: … wspólnika jednoosobowej spółki z ograniczoną odpowiedzialnością”; insured from the KRS entry (art. 13 pkt 4a). The reliefs name only art. 8 ust. 6 pkt 1 (persons running a business under Prawo przedsiębiorców): preferential base art. 18a ust. 1, Mały ZUS Plus art. 18c ust. 1, contribution holiday art. 17a ust. 1; ulga na start (Prawo przedsiębiorców art. 18) is for „Przedsiębiorca będący osobą fizyczną, który podejmuje działalność gospodarczą”, which the shareholder is not — so full contributions (ZUS-05). Health: ustawa o świadczeniach art. 81 ust. 2za — base „kwota przeciętnego miesięcznego wynagrodzenia w sektorze przedsiębiorstw w czwartym kwartale roku poprzedniego”: 9228,64 zł, 830,58 zł a month in 2026, flat, no annual reconciliation (biznes.gov.pl 00288).
- **Grade:** 2 settled — the exclusion from ulga na start rests on reading art. 18 PP; ZUS places the shareholder in the full-contribution group (ZUS 2026 table, group II).
- **Rhea's default:** a sp. z o.o. with one shareholder who is a natural person gets the shareholder's ZUS duty as a task on the person (it is his own liability as payer for himself, not the company's), at full rates.
- **Pack today:** missing (the pack knows `company.owner` only for natural-person companies).
- **Verdict:** missing
- **Deadline:** registration 7 days from the KRS entry; payment by the 20th.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/199/text/T/D20260199L.pdf; https://eli.gov.pl/api/acts/DU/2025/1461/text/T/D20251461L.pdf; https://www.biznes.gov.pl/pl/portal/00288

### ZUS-03 — Should the sole trader join voluntary sickness insurance, and what happens if he pays late?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** ustawa o sus art. 11 ust. 2: „Dobrowolnie ubezpieczeniu chorobowemu podlegają na swój wniosek osoby objęte obowiązkowo ubezpieczeniami emerytalnym i rentowym”; art. 14 ust. 1a — from the date in the application only if filed in time. Ustawa zasiłkowa (t.j. Dz.U. 2026 poz. 854) art. 4 ust. 1 pkt 2: benefit „po upływie 90 dni nieprzerwanego ubezpieczenia chorobowego – jeżeli jest ubezpieczony dobrowolnie”. Late payment no longer ends the insurance (art. 14 ust. 2 pkt 2 repealed from 1 January 2022, Dz.U. 2021 poz. 1621); instead, art. 2a ust. 1 ustawy zasiłkowej: no benefits while contribution arrears exceed „1 % minimalnego wynagrodzenia” on the day the right arises (48,06 zł in 2026), lost if not paid within 6 months (ust. 2).
- **Grade:** 1 plain
- **Rhea's default:** ask once at start (the choice changes the monthly amount); warn whenever a ZUS arrear above 1% of the minimum wage is open, because it blocks sickness pay.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** application with the registration (art. 14 ust. 1a).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/199/text/T/D20260199L.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/854/text.pdf; https://www.zus.pl/-/prawo-do-swiadczen-mimo-spoznienia-w-oplacaniu-skladek

### MON-02 — Does a sole trader want "cash-basis PIT", so that a sale is income only when the customer pays?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** PIT art. 14c ust. 1 (t.j. Dz.U. 2026 poz. 592): the cash method applies if revenue in the previous year did not exceed „2 000 000 zł”, the taxpayer keeps no full books, and files a written choice „w terminie do dnia 20 lutego roku podatkowego, a w przypadku podatnika rozpoczynającego prowadzenie działalności gospodarczej w trakcie roku podatkowego – w terminie do 20. dnia miesiąca następującego po miesiącu rozpoczęcia działalności”; ust. 2: revenue arises on „dzień uregulowania należności, niepóźniejszy niż dzień: 1) upływu 2 lat, licząc od dnia wystawienia faktury”; ust. 3 — partial payments and advance payments count; ust. 4 — only B2B revenue documented by invoices; art. 22 ust. 4a — costs from B2B transactions are deducted „w roku podatkowym, w którym zostało uregulowane zobowiązanie”. Ryczałt: ustawa o ryczałcie art. 6 ust. 1g (the same choice) and art. 15 ust. 11a (a register of such invoices with „datę uregulowania należności”). Changed: the limit was raised from 1 000 000 zł to 2 000 000 zł from 1 January 2026 (Dz.U. 2025 poz. 1858).
- **Grade:** 1 plain
- **Rhea's default:** the company declares the choice as a fact with its date; when it is on, the settlement date of a B2B invoice (the matching already records it) becomes the revenue date, with the 2-year cap from the invoice date as a dated fallback; off by default.
- **Pack today:** missing — revenue timing comes from `issue_date` only; settlements carry a date but nothing reads it for income tax.
- **Verdict:** missing
- **Deadline:** 20 February of the tax year; a business starting mid-year: the 20th of the month after it starts.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/1858/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/843/text.pdf (ryczałt, t.j.)

### MON-03 — VAT on the cash method: the tax point moves to the day the customer pays
- **Applies to:** all VAT payers that are small taxpayers (art. 2 pkt 25)
- **The law:** VAT art. 21 ust. 1: a small taxpayer „może wybrać metodę rozliczeń polegającą na tym, że obowiązek podatkowy … powstaje: 1) z dniem otrzymania całości lub części zapłaty” (sales to active VAT payers), „2) z dniem otrzymania całości lub części zapłaty, niepóźniej niż 180. dnia, licząc od dnia wydania towaru lub wykonania usługi” (others), after notifying the tax office „w terminie do końca miesiąca poprzedzającego okres”; ust. 3 — no exit before 12 months; art. 86 ust. 10e — such a taxpayer deducts input VAT „niewcześniej niż w rozliczeniu za okres, w którym podatnik dokonał zapłaty”; the invoice carries „metoda kasowa” (art. 106e ust. 1 pkt 16).
- **Grade:** 1 plain
- **Rhea's default:** the company declares the choice; the settlement date from the matching sets the VAT period of sales and purchases; the 180-day cap is a dated fallback for non-VAT customers.
- **Pack today:** `contractor.cash_method` exists but no rule reads it; the VAT period is fixed at booking (pl-book-sales-invoice, pl-book-purchase-invoice). Missing. (The VAT review owns the period rules; this entry is the payment side.)
- **Verdict:** missing
- **Deadline:** notification by the end of the month before the first period.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/1263/text.pdf

### MON-04 — At the start: which method for exchange differences — the tax method or the accounting method?
- **Applies to:** sp. z o.o. full books; sole trader KPiR (tax method only)
- **The law:** CIT art. 9b ust. 1: differences are set „na podstawie: 1) art. 15a, albo 2) przepisów o rachunkowości, pod warunkiem że w okresie, o którym mowa w ust. 3, sporządzane przez podatników sprawozdania finansowe będą badane przez firmy audytorskie”; ust. 3 — at least three tax years, announced „w zeznaniu … składanym za rok podatkowy, w którym rozpoczęli jej stosowanie”. PIT art. 14b ust. 1–2: the tax method (art. 24c), the accounting method only for „Podatnicy prowadzący księgi rachunkowe” with audited statements.
- **Grade:** 1 plain
- **Rhea's default:** the tax method (art. 15a / 24c) — a small, unaudited company cannot choose the other; for full books the accounting differences (UoR art. 30) are kept beside the tax ones, never mixed.
- **Pack today:** missing — the pack books no exchange differences at all (see MON-16, MON-17).
- **Verdict:** missing
- **Deadline:** none (the choice is announced in the first year's annual return).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf

### MON-05 — The first sale to a consumer: does the sole trader need a cash register?
- **Applies to:** all VAT payers; VAT-exempt (the duty attaches to sales to consumers, not to VAT status)
- **The law:** VAT art. 111 ust. 1 (sales to „osób fizycznych nieprowadzących działalności gospodarczej” are recorded on a cash register). Exemptions: rozporządzenie MF z 17 grudnia 2024 r. (Dz.U. 2024 poz. 1902, amended by Dz.U. 2026 poz. 420), in force from 1 January 2025, exemptions „nie dłużej jednak niż do dnia 31 grudnia 2027 r.” § 3 ust. 1 pkt 1–2: consumer sales in the previous year not above „20 000 zł”, and a business starting consumer sales after 31 December 2024 if its expected consumer sales, „w proporcji do okresu wykonywania tych czynności w danym roku podatkowym, nie przekroczy kwoty 20 000 zł”; § 5 ust. 1: the exemption „traci moc po upływie dwóch miesięcy następujących po miesiącu, w którym wartość sprzedaży … przekroczyła kwotę 20 000 zł”; § 4 ust. 1: no exemption for listed goods (e.g. computers, phones, electronics, fuel, alcohol, tobacco) and services (e.g. car repair, legal, tax advice, hairdressing and beauty, restaurants, medical services of doctors). Annex poz. 42: a service for consumers is exempt if paid in full „za pośrednictwem poczty, banku lub spółdzielczej kasy oszczędnościowo-kredytowej … a z ewidencji i dowodów dokumentujących zapłatę jednoznacznie wynika, jakiej konkretnie czynności dotyczyła”; poz. 41 the same for mail-order goods, also naming the buyer. The 2026 amendment changed only poz. 38 (parking) and added poz. 59 (coin machines without power). The value is counted without VAT (§ 3 ust. 3 pkt 2), and the limit exemption is closed to anyone who had to use a register the year before (§ 3 ust. 2 pkt 1). Coming: from 1 January 2027 a receipt with the buyer's NIP no longer stands in for an invoice for a business buyer — the KSeF transitional rule allowing it (VAT art. 145n ust. 1) runs „od dnia 1 lutego 2026 r. do dnia 31 grudnia 2026 r.”.
- **Grade:** 1 plain
- **Rhea's default:** track consumer sales against the 20,000 zł limit (pro rata for a starting business) and raise a task two months ahead of the loss of the exemption; for a business relying on poz. 41/42, keep each consumer payment matched to one specific sale — the settlement record is the very evidence the exemption asks for, so "oldest first" and multi-invoice guesses must not be used for those payments.
- **Pack today:** missing (no consumer-sales counter, no buyer type distinguishing consumers; the matching policy may settle consumer payments by amount or oldest-first).
- **Verdict:** missing
- **Deadline:** the register must be in use when the exemption lapses: after two months following the month in which consumer sales passed 20,000 zł (§ 5 ust. 1).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2024/1902/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/420/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/1263/text.pdf

### MON-06 — With a cash register, must the business also accept cards?
- **Applies to:** everyone with a cash register duty
- **The law:** Prawo przedsiębiorców art. 19a ust. 1: „Przedsiębiorca zapewnia możliwość dokonywania zapłaty w każdym miejscu, w którym działalność gospodarcza jest faktycznie wykonywana … przy użyciu instrumentu płatniczego”; ust. 2: not for one „który nie ma obowiązku prowadzenia ewidencji sprzedaży przy zastosowaniu kas rejestrujących”. In force since 1 January 2022 (Dz.U. 2021 poz. 2105).
- **Grade:** 1 plain
- **Rhea's default:** when the cash-register duty starts (MON-05), add the card-acceptance duty to the same task; card settlements from the acquirer then arrive net of fees (MON-15).
- **Pack today:** missing
- **Verdict:** later — a reminder, not a booking
- **Deadline:** from the day the cash-register duty applies.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1480/text.pdf

### MON-07 — A line arrives on the bank statement: what is it before it is anything?
- **Applies to:** everyone
- **The law:** UoR art. 20 ust. 1: „Do ksiąg rachunkowych okresu sprawozdawczego należy wprowadzić, w postaci zapisu, każde zdarzenie, które nastąpiło w tym okresie” — each in the account its nature requires; the KPiR regulation (Dz.U. 2025 poz. 1299, in force 1 January 2026) § 7 pkt 2 makes „dowody opłat pocztowych i bankowych” accounting evidence for costs. The law knows customer payments, supplier payments, taxes, contributions, fees, interest, owner money, loans, transfers between own accounts, cash deposits — each books differently.
- **Grade:** 1 plain
- **Rhea's default:** classify every line first (customer payment, supplier payment, tax, ZUS, bank fee, interest, owner, loan, own transfer, cash deposit/withdrawal, unknown), by rules on counterparty account (ZUS and tax micro-accounts, own accounts, owner's account) and the bank's transaction code; post and match only after classification; an unknown line waits for a person.
- **Pack today:** pl-bank-operation sets `kind` to "transfer" for every line; pl-post-incoming-payment posts every credit Wn 130 / Ma 201 and pl-post-outgoing-payment every debit Wn 202 / Ma 130; the MT940 reader understands only the sim-bank dialect ("PRZELEW PRZYCHODZACY/WYCHODZACY"). Wrong.
- **Verdict:** wrong
- **Deadline:** none (KPiR entries for other revenue by the 20th of the next month, KPiR regulation § 14).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf (UoR, t.j. Dz.U. 2026 poz. 522); https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text.pdf

### ZUS-04 — After the first six months: how much are the preferential social contributions in 2026, and for how long?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** ustawa o sus art. 18a ust. 1: for „pierwszych 24 miesięcy kalendarzowych od dnia rozpoczęcia wykonywania działalności gospodarczej” the base is „zadeklarowana kwota, nieniższa jednak niż 30 % kwoty minimalnego wynagrodzenia”; ust. 2 — not for someone who ran a business in the last 60 months or works for a former employer; art. 18aa ust. 3 — after ulga na start the 24 months run from the date of coverage. Minimum wage 2026: „4806 zł” (Dz.U. 2025 poz. 1242). ZUS 2026: base 1441,80 zł; emerytalne 281,44 zł, rentowe 115,34 zł, wypadkowe 24,08 zł (at 1,67%), chorobowe (voluntary) 35,32 zł; total 456,18 zł with sickness, 420,86 zł without. Fundusz Pracy is not due on a base below the minimum wage (ustawa o rynku pracy i służbach zatrudnienia, Dz.U. 2025 poz. 620, art. 259 ust. 1). Switching codes: deregister 05 40 on ZUS ZWUA, register 05 70 on ZUS ZUA (biznes.gov.pl).
- **Grade:** 1 plain
- **Rhea's default:** the 2026 amounts as master data with their source; Rhea proposes the monthly liability from the declared scheme and base; the accident rate is a company parameter defaulting to 1,67%.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** the 20th of the following month (ZUS-07).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/199/text/T/D20260199L.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/1242/text.pdf; https://www.zus.pl/baza-wiedzy/skladki-wskazniki-odsetki/skladki/wysokosc-skladek-na-ubezpieczenia-spoleczne; https://api.sejm.gov.pl/eli/acts/DU/2025/620/text.pdf

### ZUS-05 — Full social contributions in 2026
- **Applies to:** sole trader KPiR | sole trader ryczałt (after the 24 months, or from the start without the reliefs); sp. z o.o. full books (sole shareholder, ZUS-02)
- **The law:** ustawa o sus art. 18 ust. 8: the base is „zadeklarowana kwota, nieniższa jednak niż 60 % prognozowanego przeciętnego wynagrodzenia miesięcznego”; forecast 2026: 9420 zł (Obwieszczenie MRPiPS, M.P. 2025 poz. 1206). ZUS 2026, base 5652 zł: emerytalne 1103,27 zł, rentowe 452,16 zł, wypadkowe 94,39 zł, chorobowe 138,47 zł, FP and FS 138,47 zł; total 1926,76 zł, 1788,29 zł without sickness.
- **Grade:** 1 plain
- **Rhea's default:** as ZUS-04, with the 2026 table; switch automatically when the 24 months end (the date is computed from ZUS-01).
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** the 20th of the following month.
- **Sources:** https://api.sejm.gov.pl/eli/acts/MP/2025/1206/text.pdf; https://www.zus.pl/baza-wiedzy/skladki-wskazniki-odsetki/skladki/wysokosc-skladek-na-ubezpieczenia-spoleczne

### ZUS-06 — How much health contribution does the sole trader pay in 2026, by form of taxation?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** ustawa o świadczeniach opieki zdrowotnej (t.j. Dz.U. 2025 poz. 1461): skala — 9% (art. 79 ust. 1) of the previous month's income (art. 81 ust. 2), with the minimum wage as the monthly floor of the base (art. 81 ust. 2d); liniowy — art. 79a ust. 1: „4,9 % podstawy wymiaru składki, jednak nie mniej niż kwota odpowiadająca 9 %: 1) minimalnego wynagrodzenia obowiązującego w pierwszym dniu roku składkowego”; ryczałt — art. 81 ust. 2e–2f, three bands by revenue since the start of the year, on 60%, 100% and 180% of the Q4 2025 average wage (9228,64 zł). ZUS (2 February 2026): skala/liniowy at least „432,54 zł” for February 2026 to January 2027 (314,96 zł for January 2026); ryczałt 498,35 zł (revenue up to 60 000 zł), 830,58 zł (above 60 000 up to 300 000 zł), 1495,04 zł (above 300 000 zł). What changed: the 2025 floor of 75% of the minimum wage was for 2025 only (Dz.U. 2024 poz. 1915 art. 4) and has ended; the 2025 bill lowering the health contribution from 2026 (ustawa z 4 kwietnia 2025 r.) was vetoed by the President on 7 May 2025 (druk 1236) and no new law was found.
- **Grade:** 1 plain
- **Rhea's default:** compute the monthly contribution from the declared form and the books (skala, liniowy: last month's income from the KPiR; ryczałt: the band reached by year-to-date revenue, moving up the month the band is crossed), with the floor; amounts as dated master data.
- **Pack today:** missing — the pack has no form-of-taxation field on `company` at all, so nothing can branch on it.
- **Verdict:** missing
- **Deadline:** the 20th of the following month.
- **Sources:** https://eli.gov.pl/api/acts/DU/2025/1461/text/T/D20251461L.pdf; https://www.zus.pl/-/informacja-w-sprawie-podstawy-wymiaru-składki-oraz-kwoty-składki-na-ubezpieczenie-zdrowotne-w-2026-r.; https://api.sejm.gov.pl/eli/acts/DU/2024/1915/text.pdf; https://api.sejm.gov.pl/sejm/term10/prints/1236/1236.pdf

### ZUS-07 — Every month: when are contributions due, and must a declaration be filed?
- **Applies to:** sole trader KPiR | sole trader ryczałt | sp. z o.o. full books
- **The law:** ustawa o sus art. 47 ust. 1: „do 15 dnia następnego miesiąca – dla płatników składek posiadających osobowość prawną; 4) do 20 dnia następnego miesiąca – dla pozostałych płatników składek” (ust. 14 applies it to health, FP and FS too); art. 47 ust. 2: a payer paying only for himself „przesyła jedynie deklarację rozliczeniową”. The old exemption from monthly declarations was repealed from 1 January 2022 (Dz.U. 2021 poz. 2105); ZUS: „Od składek należnych za 01/2022 każdy przedsiębiorca będący płatnikiem składek ma obowiązek comiesięcznego składania dokumentów rozliczeniowych … bez względu na wskazaną formę opodatkowania.”
- **Grade:** 1 plain
- **Rhea's default:** a monthly ZUS DRA object per payer (the `tax_declaration` kind `zus_dra` already exists), drafted from the month's figures and approved by a person, due on the 20th (sole trader, shareholder) or 15th (a sp. z o.o. paying for employees), with a task three days ahead.
- **Pack today:** `tax_declaration` (kind `zus_dra`, `due_date`, `payment_status`) and `task_schedule` exist as types; no rule creates a declaration or a schedule. Missing.
- **Verdict:** missing
- **Deadline:** the 20th (or 15th) of the following month.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/199/text/T/D20260199L.pdf; https://www.zus.pl/o-zus/o-nas/programy-transformacji-cyfrowej-zus/zmiany-od-2022-r./faq

### ZUS-08 — Paying ZUS from the bank: one transfer, and what does it settle?
- **Applies to:** everyone who pays contributions
- **The law:** ustawa o sus art. 47 ust. 4: social, health and fund contributions „opłaca się jedną wpłatą na przydzielony przez Zakład numer rachunku składkowego” (since 1 January 2018); the split of one payment: rozporządzenie RM z 21 września 2017 r. (t.j. Dz.U. 2022 poz. 1771) § 7 ust. 1 — „proporcjonalnego procentowego podziału wpłaty według kwot składek przypadających do zapłaty na poszczególne fundusze”, and § 12 — within a fund „począwszy od należności o najwcześniejszym terminie płatności”. Prawo bankowe art. 62b ust. 2 pkt 2 lit. b: contributions may be paid from the VAT account.
- **Grade:** 1 plain
- **Rhea's default:** a debit to the payer's contribution account number (NRS) is a ZUS payment, settled against the open ZUS declarations oldest first, split by fund in proportion as ZUS does; read from the main or the VAT account.
- **Pack today:** the line posts Wn 202 / Ma 130 as if a supplier were paid (pl-post-outgoing-payment) and stays unsettled; the chart has 405 (social insurance cost) and 225 but no ZUS settlement account. Wrong.
- **Verdict:** wrong
- **Deadline:** the 20th (or 15th).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/199/text/T/D20260199L.pdf; https://api.sejm.gov.pl/eli/acts/DU/2022/1771/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/38/text.pdf

### MON-08 — Paying taxes from the bank: where does the money go, and what does the payment settle?
- **Applies to:** everyone
- **The law:** Ordynacja podatkowa (t.j. Dz.U. 2026 poz. 622) art. 61b § 1: payment to the tax office „jest realizowana przy użyciu rachunku bankowego urzędu skarbowego albo identyfikującego podatnika, płatnika lub urząd skarbowy indywidualnego rachunku podatkowego (mikrorachunek podatkowy)”; art. 62 § 1: a payment is credited „zgodnie ze wskazaniem podatnika”, and without it to the earliest due liability; § 2: advance payments from the earliest due. Prawo bankowe art. 62b ust. 2 pkt 2 lit. a: VAT, PIT and CIT may be paid from the VAT account.
- **Grade:** 1 plain
- **Rhea's default:** a debit to the company's micro-account is a tax payment, matched to the tax liability it names (the transfer's symbol and period), not to a supplier invoice; a payment from the VAT account is read from that account's statement.
- **Pack today:** `tax_declaration` type exists (kinds jpk_v7, vat7, pit, zus_dra, pcc3) but no rule creates it or matches payments to it; `vat_settlement` holds `vat_due` but nothing settles it; the payment posts Wn 202 / Ma 130. Wrong.
- **Verdict:** wrong
- **Deadline:** each tax's own date (outside this area).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/622/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/38/text.pdf

### MON-09 — Cash in the till: are cash receipts (KP), cash payments (KW) and a cash report required?
- **Applies to:** sole trader KPiR | sole trader ryczałt | sp. z o.o. full books
- **The law:** No act names KP, KW or a "raport kasowy" (the consolidated UoR text and the KPiR regulation in force from 1 January 2026, Dz.U. 2025 poz. 1299, were searched; neither contains them). Full books: UoR art. 20 ust. 1 (every event recorded), art. 20 ust. 2 pkt 3 (internal evidence „dotyczące operacji wewnątrz jednostki”) and art. 20 ust. 3 pkt 1 (collective evidence „służące do dokonania łącznych zapisów zbioru dowodów źródłowych, które muszą być w dowodzie zbiorczym pojedynczo wymienione”) — the cash report is such a collective document; art. 24 ust. 1: books kept „rzetelnie, bezbłędnie, sprawdzalnie i bieżąco”. KPiR: the book records revenue and costs, not cash; sales are entered from invoices, fiscal reports, or a daily internal document (KPiR regulation § 11 ust. 1), costs from the documents in § 6–9.
- **Grade:** 1 plain — not required by name; for full books cash movements must be evidenced and booked like any other event, and the cash report is the customary collective document that does it; for KPiR and ryczałt there is no cash book at all.
- **Rhea's default:** full books: each cash movement is an event with its evidence (the invoice or receipt, or an internal KP/KW the person confirms), a cash report per period derived as a view, and a refusal of any cash balance below zero; KPiR/ryczałt: no cash ledger, only the cost and revenue documents (and the 15,000 zł B2B limit in MON-10).
- **Pack today:** account 100 Kasa exists in the chart; no cash object, rule or view. Missing.
- **Verdict:** missing
- **Deadline:** none statutory for the report; KPiR entries by the 20th of the next month (§ 11 ust. 2, § 14).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text.pdf

### MON-10 — Paying or being paid in cash over 15,000 zł between businesses
- **Applies to:** sole trader KPiR | sp. z o.o. full books (cost rule); everyone (the duty itself)
- **The law:** Prawo przedsiębiorców art. 19: payments go „za pośrednictwem rachunku płatniczego przedsiębiorcy, w każdym przypadku gdy: 1) stroną transakcji, z której wynika płatność, jest inny przedsiębiorca oraz 2) jednorazowa wartość transakcji, bez względu na liczbę wynikających z niej płatności, przekracza 15 000 zł”, foreign currency at the NBP average rate „z ostatniego dnia roboczego poprzedzającego dzień dokonania transakcji”. PIT art. 22p ust. 1 pkt 1 / CIT art. 15d ust. 1 pkt 1: no cost „w tej części, w jakiej płatność … została dokonana bez pośrednictwa rachunku płatniczego”; ust. 2: if already a cost, reduce costs (or increase revenue) „w miesiącu, w którym … została dokonana płatność bez pośrednictwa rachunku płatniczego”. The KPiR regulation § 6 ust. 5 enters the reduction „ze znakiem minus (-) lub kolorem czerwonym”. What changed: the cut to 8,000 zł enacted for 1 January 2024 (Dz.U. 2021 poz. 2105 art. 22 pkt 1) was repealed before it took effect (Dz.U. 2023 poz. 1414 art. 1 pkt 4), as was the planned 20,000 zł limit for consumers; no 2025–2026 amendment touches art. 19.
- **Grade:** 1 plain
- **Rhea's default:** a B2B purchase whose transaction value exceeds 15,000 zł and which is paid (even in part, even in instalments) in cash: the cash part is not a cost; Rhea warns before a cash payment is recorded and books the reversal in the month of payment if it happens.
- **Pack today:** missing (no cash payments at all; `contractor.cash_method` is the VAT cash method, not this).
- **Verdict:** missing
- **Deadline:** cost reversal in the month of the cash payment (art. 22p ust. 2 / 15d ust. 2).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2025/1480/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2023/1414/text.pdf

### MON-11 — Paying a supplier more than 15,000 zł: must the account be on the white list?
- **Applies to:** sole trader KPiR | sp. z o.o. full books (cost); all VAT payers (joint VAT liability)
- **The law:** PIT art. 22p ust. 1 pkt 2 / CIT art. 15d ust. 1 pkt 2: no cost for the part of a payment in an art. 19 transaction made „przelewem na rachunek inny niż zawarty na dzień zlecenia przelewu w wykazie podmiotów, o którym mowa w art. 96b ust. 1 ustawy o podatku od towarów i usług” when the supplier is an active VAT payer; PIT art. 22p ust. 4 pkt 1 (and CIT art. 15d ust. 4): not if the payer files, at the first payment to that account, the notification under Ordynacja art. 117ba § 3 pkt 2 „w terminie 7 dni od dnia zlecenia przelewu”. Ordynacja art. 117ba § 1: the VAT-registered buyer „odpowiada solidarnie całym swoim majątkiem wraz z dostawcą” for the supplier's VAT arrears on that supply when paid to an unlisted account; § 3 pkt 1: not for transactions other than art. 19 ones (15,000 zł). History: the notice was due within 3 days to the issuer's tax office until 30 June 2020, and within 7 days to the payer's own office since 1 July 2020 (Dz.U. 2020 poz. 1065); from 1 October 2026 a power of attorney to sign returns also covers signing this notice (Ordynacja art. 117ba § 3a, Dz.U. 2026 poz. 846). The list is updated „w dni robocze, raz na dobę” (VAT art. 96b ust. 8).
- **Grade:** 1 plain
- **Rhea's default:** before an outgoing payment for an art. 19 transaction is ordered, check the white list for the supplier's account on the order date and keep the answer as evidence; if the account is not listed, warn, and if paid anyway, open a task to file the notification within 7 days; a payment found on the statement without a prior check gets the check after the fact with the order date.
- **Pack today:** missing — pl-match-outgoing matches by IBAN taken from KSeF invoices or registration, but nothing checks or records the white-list state.
- **Verdict:** missing
- **Deadline:** notification within 7 days of the transfer order (PIT art. 22p ust. 4 pkt 1; Ordynacja art. 117ba § 3 pkt 2).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/622/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/1263/text.pdf

### MON-12 — Paying an invoice marked "mechanizm podzielonej płatności": what must the payment look like?
- **Applies to:** all VAT payers (payer side); sole trader KPiR | sp. z o.o. full books (cost consequence)
- **The law:** VAT art. 108a ust. 1a: for goods or services from załącznik nr 15, on an invoice whose „kwota należności ogółem przekracza kwotę 15 000 zł”, taxpayers „są obowiązani zastosować mechanizm podzielonej płatności”; ust. 1d: not to the extent the amounts are offset; ust. 2: the VAT part goes to the VAT account, the net part to the account „dla których jest prowadzony rachunek VAT”; ust. 3: paid in złoty with the bank's split-payment message stating the VAT amount, the gross amount, „numer faktury” and the supplier's tax number — for payments from 1 January 2027, for a structured invoice „numer identyfikujący tę fakturę w Krajowym Systemie e-Faktur” (ust. 3 pkt 3) or, for several, „identyfikator zbiorczy nadawany przez Krajowy System e-Faktur” (ust. 3c pkt 2), per art. 17 of the KSeF act as reworded by Dz.U. 2025 poz. 1203; ust. 3a: one payment may cover a supplier's invoices from one day to one month. Cost: PIT art. 22p ust. 1 pkt 3 / CIT art. 15d ust. 1 pkt 3 — no cost where, despite the invoice saying „mechanizm podzielonej płatności”, the payment bypassed it. Prawo bankowe art. 62b ust. 2: the VAT account may be debited for the VAT part of a supplier payment, and for VAT, PIT, CIT and ZUS (pkt 2 lit. a–b).
- **Grade:** 1 plain
- **Rhea's default:** a purchase invoice carrying the MPP mark is paid only by a split payment (Rhea's payment proposal fills the message from the invoice: VAT, gross, the invoice number — its KSeF number for payments from 1 January 2027 — and the supplier's NIP); the outgoing statement lines from the main and the VAT account together settle it; a non-split payment of such an invoice reverses the cost in the month paid.
- **Pack today:** `purchase_invoice` has no MPP flag and no due date; the VAT account is not modelled (one `bank_account`, no kind); the outgoing matching reads only the main account's lines and matches by `number`, not `ksef_ref`. Missing.
- **Verdict:** missing
- **Deadline:** none (cost reversal in the month of a non-split payment, art. 22p ust. 2 / 15d ust. 2).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/1263/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/38/text.pdf

### MON-13 — From 2027 every transfer between VAT payers must quote the KSeF number: will customer payments still find their invoices?
- **Applies to:** all VAT payers
- **The law:** VAT art. 108g ust. 1 (applies to payments from 1 January 2027): an active VAT payer paying another for structured invoices by transfer „jest obowiązany do podania numeru identyfikującego te faktury w Krajowym Systemie e-Faktur lub identyfikatora zbiorczego nadanego przez Krajowy System e-Faktur”; ust. 3 — the payee's status is taken from the white list „na dzień dokonywania płatności”. Art. 108a ust. 3 pkt 3 and ust. 3c pkt 2 (the split-payment message carries the KSeF number, or the KSeF collective identifier for several invoices) apply to payments from the same day. Both by art. 17 of the KSeF act (Dz.U. 2023 poz. 1598) as reworded by Dz.U. 2025 poz. 1203: „stosuje się do płatności dokonanych od dnia 1 stycznia 2027 r.” Meanwhile: art. 108a ust. 1b — a seller that must put the mark „jest obowiązany do przyjęcia płatności … z zastosowaniem mechanizmu podzielonej płatności”; art. 106e ust. 1 pkt 18a — the words „mechanizm podzielonej płatności” on invoices above 15,000 zł for załącznik nr 15 goods and services; Prawo bankowe art. 62b ust. 1 pkt 1 — the VAT part lands on the VAT account.
- **Grade:** 1 plain
- **Rhea's default:** the sales invoice keeps its KSeF number (and the MPP mark when due); incoming matching reads the transfer title against the KSeF number first, the invoice number second, and a collective identifier against the invoices it covers; outgoing payments Rhea proposes quote the purchase invoice's KSeF number; the main-account and VAT-account lines of one split payment are joined before matching the gross.
- **Pack today:** wrong on three counts — the FA(3) writer (markets/pl/fa3.go) always writes P_18A = 2, so an invoice that must carry the MPP mark never does; pl-match-incoming and pl-match-outgoing compare the title only with `number` (the sales invoice's KSeF number lives on `ksef_submission`; `purchase_invoice.ksef_ref` is not used for matching); the VAT-account statement is not read, so wherever a bank shows only the net on the main account (how each bank presents a split payment in MT940 was not verified) the invoice is settled partially.
- **Verdict:** wrong
- **Deadline:** 1 January 2027 (KSeF number in transfers).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/1263/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/1203/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/38/text.pdf

### MON-14 — A payment arrives before the invoice: what is it?
- **Applies to:** all VAT payers; sole trader KPiR | sole trader ryczałt; sp. z o.o. full books
- **The law:** VAT art. 19a ust. 8: if before delivery „otrzymano całość lub część zapłaty, w szczególności przedpłatę, zaliczkę … obowiązek podatkowy powstaje z chwilą jej otrzymania w odniesieniu do otrzymanej kwoty”; art. 106i ust. 2: the invoice is due „niepóźniej niż 15. dnia miesiąca następującego po miesiącu, w którym otrzymano całość lub część zapłaty”. PIT art. 14 ust. 3 pkt 1: not revenue are „pobranych wpłat lub zarachowanych należności na poczet dostaw towarów i usług, które zostaną wykonane w następnych okresach” (for income tax the advance is not yet revenue — except under cash-basis PIT, art. 14c ust. 3 pkt 2).
- **Grade:** 1 plain
- **Rhea's default:** an incoming line from a customer with no open invoice to match becomes an advance: VAT in the month received and a task to issue the advance invoice by the 15th of the next month; not revenue for income tax unless the company chose cash-basis PIT.
- **Pack today:** the line stays unsettled (matching finds nothing) after posting Wn 130 / Ma 201; no VAT, no task. Missing.
- **Verdict:** missing
- **Deadline:** advance invoice by the 15th of the month after the payment (art. 106i ust. 2).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/1263/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf

### MON-15 — A card or online-payment operator pays out the customer's money minus its commission: how much is revenue?
- **Applies to:** everyone selling through payment operators
- **The law:** PIT art. 14 ust. 1 and CIT art. 12 ust. 3: revenue is the amount due „po wyłączeniu wartości zwróconych towarów, udzielonych bonifikat i skont” — an operator's commission is none of these.
- **Grade:** 1 plain
- **Rhea's default:** revenue stays the gross invoice amount; the payout matches the invoices it covers plus a fee line, the fee booked as a cost on the operator's document; the commission is never a matching "difference".
- **Pack today:** the matching tolerates only 2 grosze and only on incoming payments by reference; a net payout stays unsettled or settles invoices partially. Missing.
- **Verdict:** missing
- **Deadline:** none
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf (CIT, t.j. Dz.U. 2026 poz. 554)

### MON-16 — An invoice in euro: at which rate is it booked?
- **Applies to:** everyone
- **The law:** PIT art. 11a ust. 1: „Przychody w walutach obcych przelicza się na złote według kursu średniego walut obcych ogłaszanego przez Narodowy Bank Polski z ostatniego dnia roboczego poprzedzającego dzień uzyskania przychodu”; ust. 2 the same for costs („dzień poniesienia kosztu”); CIT art. 15a ust. 6 defines the NBP average rate as the one „z ostatniego dnia roboczego poprzedzającego dzień uzyskania przychodu lub poniesienia kosztu” and CIT art. 12 ust. 2 and art. 15 ust. 1 say the same for a company's revenue and costs. UoR art. 30 ust. 2 pkt 2: other operations at the NBP average rate „z dnia poprzedzającego ten dzień”; art. 9: „Księgi rachunkowe prowadzi się w języku polskim i w walucie polskiej.” KPiR regulation § 6 ust. 4: a document only in foreign currency must be converted under the income-tax rules and the result written on it or attached. Ryczałt: the ryczałt act has no conversion rule of its own (art. 6 ust. 1c applies PIT art. 24c to exchange differences, negative ones reducing revenue); how art. 11a reaches ryczałt revenue was not verified.
- **Grade:** 1 plain for income tax ("last business day before"); for the books, UoR's „z dnia poprzedzającego” is read in practice as the last rate published before the day (no rate is published on non-working days) — the same date in effect.
- **Rhea's default:** convert at the NBP average rate of the last business day before the revenue or cost date, the rate keyed by the date it applies to, kept as an event with its NBP table number; the converted PLN amounts are part of the invoice's facts, so replay never re-converts.
- **Pack today:** wrong — pl-post-sales-invoice and pl-post-purchase-invoice convert to PLN at the `fx_rate` keyed by the issue (invoice) date itself (`EUR/PLN/<issue_date>`; the kernel looks the rate up by that exact date), so the rate is D-1 only if a rate publisher keys rates by the day they apply to — none exists in the repository; and income tax wants the day before the revenue date (art. 14 ust. 1c: the day of delivery, „niepóźniej niż dzień: 1) wystawienia faktury albo 2) uregulowania należności”), not before the issue date. pl-post-incoming-payment and pl-post-outgoing-payment have no `convert`, so a euro payment posts euro amounts into the PLN book against UoR art. 9, and 201/202 end up holding PLN from the invoice and EUR from the payment.
- **Verdict:** wrong
- **Deadline:** none
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text.pdf

### MON-17 — The euro invoice is paid: is there an exchange difference, and at which rate?
- **Applies to:** everyone with foreign-currency invoices or accounts
- **The law:** CIT art. 15a ust. 2 pkt 1 / ust. 3 pkt 1 (PIT art. 24c the same): a positive or negative difference arises when the revenue converted at the NBP average rate differs from its value „w dniu jego otrzymania, przeliczonej według faktycznie zastosowanego kursu waluty z tego dnia” (pkt 2: costs on the day of payment); ust. 4: „uwzględnia się kursy faktycznie zastosowane w przypadku sprzedaży lub kupna walut obcych oraz otrzymania należności lub zapłaty zobowiązań. W pozostałych przypadkach, a także gdy … nie jest możliwe uwzględnienie faktycznie zastosowanego kursu waluty w danym dniu, stosuje się kurs średni ogłaszany przez Narodowy Bank Polski z ostatniego dnia roboczego poprzedzającego ten dzień”; ust. 7: the day of payment is „dzień uregulowania zobowiązań w jakiejkolwiek formie, w tym w wyniku potrącenia”; pkt 3 of ust. 2–3: own currency funds in and out at the actual rates, ust. 8 — the order of valuing them „według przyjętej metody stosowanej w rachunkowości”. UoR art. 30 ust. 2 pkt 1: payments at the rate „faktycznie zastosowanym w tym dniu”; art. 30 ust. 4: differences go to „przychodów lub kosztów finansowych”.
- **Grade:** 1 plain for which rate; for a euro payment landing on a euro account (no conversion happens) the fallback of ust. 4 applies — KIS 0112-KDIL2-2.4011.576.2020.2.AA (9 December 2020): „w odniesieniu do środków wpływających na rachunek walutowy powinien zastosować średni kurs ogłaszany przez Narodowy Bank Polski z ostatniego dnia roboczego poprzedzającego dzień otrzymania należności, natomiast w momencie kiedy dokonuje wymiany waluty w banku powinien zastosować kurs faktycznie zastosowany przez bank w dniu tej wymiany”; no contrary ruling was found. FIFO is not required: ust. 8 takes the ordering method from the accounting policy (KIS 0112-KDWL.4011.91.2023.2.JK, 22 December 2023, points to UoR art. 34 ust. 4: weighted average, FIFO, LIFO).
- **Rhea's default:** PLN account receiving euro: the actual rate is the bank's, read from the statement (PLN amount ÷ currency amount); euro account: the NBP average rate of the previous business day; the difference posts to financial income or costs and is a tax difference under art. 15a / 24c; own-currency-fund differences follow the ordering method declared in the accounting policy (FIFO by default); a withdrawal of currency for private purposes creates no difference (the same KIS ruling).
- **Pack today:** wrong — pl-match-incoming and pl-match-outgoing declare no `convert`, so the engine's existing FX difference (exercised in internal/exec/matching_test.go with the payment-date rate) is never used by the Poland pack; and the test's rate source is the payment date, not "the actual rate, else the last business day before".
- **Verdict:** wrong
- **Deadline:** none
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf; https://eureka.mf.gov.pl/informacje/podglad/423507; https://eureka.mf.gov.pl/informacje/podglad/574335

### MON-18 — Bank charges and interest on the account: how are they booked and taxed?
- **Applies to:** everyone
- **The law:** Charges: a cost under PIT art. 22 ust. 1 / CIT art. 15 ust. 1 when connected with the business; the bank's statement is the evidence (KPiR regulation § 7 pkt 2: „dowody opłat pocztowych i bankowych”). Interest received, sole trader: PIT art. 14 ust. 2 pkt 5 — business revenue includes „odsetki od środków pieniężnych na rachunkach rozliczeniowych, o których mowa w przepisach prawa bankowego … prowadzonych w związku z wykonywaną działalnością gospodarczą”, and art. 30a ust. 1 pkt 3 (the 19% flat tax on interest from accounts) applies „z zastrzeżeniem art. 14 ust. 2 pkt 5” — so interest on a business (rozliczeniowy) account is business revenue, interest on a personal account is the 19% flat tax, outside the business. One limit: interest on a separate term-deposit account (not a deposit made on the business account itself) was held to be capital income under art. 30a (Dyrektor Izby Skarbowej w Katowicach, IBPB-1-1/4511-264/15/EN, 24 August 2015: „odsetek od lokat terminowych na rachunkach lokat terminowych nie można zaliczyć do źródła przychodów z działalności gospodarczej”). Ryczałt: such interest is taxed at 3% (ustawa o ryczałcie art. 12 ust. 1 pkt 7 lit. e, covering PIT art. 14 ust. 2 pkt 2, 5–10). CIT: interest received is revenue (art. 12 ust. 1 pkt 1). Full books: UoR art. 30 ust. 4 and the P&L layout place interest among financial income and costs.
- **Grade:** 1 plain (charges have no article of their own; they fall under the general cost clause)
- **Rhea's default:** a bank-fee line becomes a cost (KPiR "pozostałe wydatki"; full books: a cost account per the company's policy) with the statement as evidence; interest on a business account (and deposits made on it) is business revenue of the month credited (KPiR, ryczałt at 3%, CIT), interest on a separate deposit account or a personal account is not business revenue; a bank that withheld 19% on business-account interest is flagged for the owner.
- **Pack today:** pl-bank-operation always sets `kind` "transfer"; a fee or an interest credit posts as a supplier or customer payment (Wn 202 / Ma 130, Wn 130 / Ma 201) and waits unmatched. Wrong.
- **Verdict:** wrong
- **Deadline:** KPiR entry by the 20th of the next month (§ 14 for other revenue).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/1299/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/843/text.pdf; https://eureka.mf.gov.pl/informacje/podglad/270297

### MON-19 — The customer paid one or two grosze too little or too much: may the difference be written off, where, and is it a tax cost or income?
- **Applies to:** everyone (books: sp. z o.o. full books; tax: all)
- **The law:** Books: UoR art. 4 ust. 4 — simplifications in the accounting policy „jeżeli nie wywiera to istotnie ujemnego wpływu” on the true and fair view; art. 3 ust. 1 pkt 32 lit. d — other operating costs and income („pozostałych kosztach i przychodach operacyjnych”) include those connected „z odpisaniem należności i zobowiązań przedawnionych, umorzonych, nieściągalnych”, while art. 30 ust. 4 sends exchange differences to financial items; art. 20 ust. 5 — an entry generated by an algorithm from data already in the books counts as a source document if its origin and the person responsible can be established. Tax, a short payment written off: PIT art. 23 ust. 1 pkt 20 / CIT art. 16 ust. 1 pkt 25 lit. a — a written-off receivable is a cost only in the part previously booked as revenue and only if its uncollectability is documented, e.g. by a „protokołem sporządzonym przez podatnika, stwierdzającym, że przewidywane koszty procesowe i egzekucyjne związane z dochodzeniem wierzytelności byłyby równe albo wyższe od jej kwoty” (PIT art. 23 ust. 2 pkt 3 / CIT art. 16 ust. 2 pkt 3). An overpayment kept: revenue only once the liability to the customer is forgiven or time-barred („wartość umorzonych lub przedawnionych zobowiązań”, PIT art. 14 ust. 2 pkt 6; CIT art. 12 ust. 1 pkt 3 lit. a; ryczałt at 3%, art. 12 ust. 1 pkt 7 lit. e).
- **Grade:** 3 open — books: the write-off is a policy simplification (art. 4 ust. 4), and its place among other operating items follows art. 3 ust. 1 pkt 32 lit. d (no KSR or Ministry position found; the account numbers 760/761 are convention — art. 10 ust. 1 pkt 3 lit. a leaves the chart to the company); but one practitioner source holds that a not-yet-time-barred liability of a few grosze may not be written off by the company's own decision at all (infor.pl, I. Pazio, 7 February 2013, older law — secondary, verify). Tax: no KIS ruling squarely on customer penny under- or overpayments was found; on the nearest question (rounding differences of the VAT return booked to other operating items) practice split — Dyrektor Izby Skarbowej w Poznaniu ILPB3/423-496/13-4/PR (17 January 2014) accepted them as tax revenue and costs, and the Minister of Finance changed that interpretation (DD10/033/356/RDX/14/RWPD-126497, 19 November 2014): they „nie podlegają … zaliczeniu do przychodów oraz kosztów uzyskania przychodów”. For payments the statute gives the route: a short payment is a cost only with the art. 16 ust. 2 pkt 3 / art. 23 ust. 2 pkt 3 protocol; an overpayment is revenue only once „umorzonych lub przedawnionych”.
- **Rhea's default:** the 2-grosze tolerance is a declared accounting-policy simplification; a short difference posts to other operating costs (761) with a generated protocol (amount, invoice, reason: recovery costs exceed the amount) as its document under art. 20 ust. 5, and is a tax cost on that protocol; an over difference stays a liability to the customer (or posts to other operating income, 760, under the policy) but is not tax revenue until time-barred or released by the customer — a tax difference Rhea tracks; never netted, never on the exchange-difference account.
- **Pack today:** wrong — pl-record-difference, pl-post-short-difference (Wn 756 / Ma 201) and pl-post-over-difference (Wn 201 / Ma 756) send both directions to one account 756 "Różnice kursowe i zaokrąglenia", typed expense, which also takes the conversion rounding plug of the invoice postings, so settlement write-offs sit with exchange items (financial by art. 30 ust. 4) and gains net against losses; the tolerance is applied only on incoming payments matched by reference (outgoing penny differences are never written off); no protocol, no tax flag.
- **Verdict:** wrong
- **Deadline:** none
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf; https://eureka.mf.gov.pl/informacje/podglad/290981; https://ksiegowosc.infor.pl/rachunkowosc/sprawozdawczosc/139598,Co-zrobic-z-kilkugroszowymi-zobowiazaniami-na-dzien-bilansowy.html (secondary)

### OWN-03 — The owner of a sole proprietorship moves money between the business and himself: is it income or cost?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** PIT art. 14 ust. 1: revenue is „kwoty należne” from the business; art. 22 ust. 1: costs are „koszty poniesione w celu osiągnięcia przychodów lub zachowania albo zabezpieczenia źródła przychodów”. An owner's transfer to or from his own pocket is neither. (A sole proprietorship is not a separate person; there is no provision on "withdrawals" because there is nothing to tax.)
- **Grade:** 1 plain
- **Rhea's default:** the owner (or the host, asking him) classifies such a line as an owner transfer: no revenue, no cost, no VAT; in KPiR and the ryczałt register it produces no entry and only takes the line out of the matching set.
- **Pack today:** every incoming line posts Wn 130 / Ma 201 and every outgoing line Wn 202 / Ma 130 (pl-post-incoming-payment, pl-post-outgoing-payment) before any classification; an owner deposit becomes a customer payment and a withdrawal a supplier payment. Wrong.
- **Verdict:** wrong
- **Deadline:** none
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf

### OWN-04 — Can the sole trader pay himself a salary as a business cost?
- **Applies to:** sole trader KPiR
- **The law:** PIT art. 23 ust. 1 pkt 10: „Nie uważa się za koszty uzyskania przychodów: … wartości własnej pracy podatnika, jego małżonka i małoletnich dzieci” — with the exception for a spouse's or minor child's pay under employment and similar titles.
- **Grade:** 1 plain
- **Rhea's default:** any "salary" the owner pays himself is an owner withdrawal (OWN-03), never a cost; a host question that proposes otherwise is refused with this article.
- **Pack today:** no rule makes it a cost, but no guard either; the line would post as a supplier payment (see OWN-03). Missing.
- **Verdict:** missing
- **Deadline:** none
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf

### OWN-05 — One account for private and business money: what does "mark as private" do?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** PIT art. 22 ust. 1 (a cost only if incurred to earn or keep revenue); VAT art. 86 ust. 1: input VAT is deductible „W zakresie, w jakim towary i usługi są wykorzystywane do wykonywania czynności opodatkowanych”. Nothing requires business costs to be paid from a business account: the only payment-channel rule is the payment account for B2B transactions over 15,000 zł (PIT art. 22p ust. 1 pkt 1 — „bez pośrednictwa rachunku płatniczego”), and a personal account is a payment account.
- **Grade:** 1 plain
- **Rhea's default:** "mark as private" is the owner's decision, recorded as an event with his name: the line leaves the matching set and produces no KPiR, ryczałt or VAT entry; a business invoice paid from the private account still matches if the owner declares that account as one Rhea reads; refuse "private" for a legal entity (a company has no private sphere — see OWN-07).
- **Pack today:** `bank_operation.settlement_status` has the value `private` and the lifecycle allows unsettled → private, but no activity or rule sets it, and the payment postings fire when the line is materialized, before anyone could mark it. Wrong.
- **Verdict:** wrong
- **Deadline:** none
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/1263/text.pdf

### OWN-06 — The owner takes goods or uses business things for himself: is there VAT?
- **Applies to:** all VAT payers (sole traders above all)
- **The law:** VAT art. 7 ust. 2 pkt 1: a supply is also „przekazanie lub zużycie towarów na cele osobiste podatnika … – jeżeli podatnikowi przysługiwało, w całości lub w części, prawo do obniżenia kwoty podatku należnego o kwotę podatku naliczonego”; art. 8 ust. 2 pkt 1 — the same for private use of business goods (a service).
- **Grade:** 1 plain
- **Rhea's default:** a door for the owner to say "I took this for myself"; when input VAT was deducted on the item, Rhea records output VAT on it; cars are left to the car rules (mixed-use limits are another area).
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** VAT for the month of the withdrawal (general VAT timing; not separately verified here).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/1263/text.pdf

### OWN-07 — Money between a sp. z o.o. and its owner: what is allowed and how is it booked?
- **Applies to:** sp. z o.o. full books
- **The law:** KSH art. 189 § 1: „W czasie trwania spółki nie wolno zwracać wspólnikom wniesionych wkładów”; art. 191 § 1 (profit only from the annual financial statements, by resolution); art. 15 § 1: a loan „z członkiem zarządu … wymaga zgody zgromadzenia wspólników”. PCC art. 9 pkt 10 lit. i: exempt are loans granted „przez wspólnika (akcjonariusza) spółce kapitałowej”; a loan the other way is taxed at „0,5 %” (art. 7 ust. 1 pkt 4), paid by „biorącym pożyczkę” (art. 4 pkt 7) within 14 days (art. 10 ust. 1).
- **Grade:** 1 plain
- **Rhea's default:** a company line to or from the owner must be one of: shareholder loan in (liability, no PCC), loan out (receivable from the owner, consent evidence if he is on the board, PCC on the owner), reimbursement against documents, dividend after a resolution, or capital; "private" is refused; an unexplained outflow to the owner stays a receivable from him until explained.
- **Pack today:** no shareholder settlement account in the chart (no 24x), no rule; lines go to 201/202. Wrong.
- **Verdict:** wrong
- **Deadline:** PCC on a company loan to the owner: 14 days (paid by the owner).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2024/18/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/191/text.pdf

### MON-20 — What payment term may a business agree with another business, and from when does it run?
- **Applies to:** everyone (transactions between businesses)
- **The law:** ustawa o przeciwdziałaniu nadmiernym opóźnieniom w transakcjach handlowych (t.j. Dz.U. 2023 poz. 1790; no amendment since): art. 7 ust. 2: the term „nie może przekraczać 60 dni, liczonych od dnia doręczenia dłużnikowi faktury lub rachunku … chyba że strony w umowie wyraźnie ustalą inaczej i pod warunkiem że ustalenie to nie jest rażąco nieuczciwe wobec wierzyciela”; art. 7 ust. 2a: no exception when the debtor is a large enterprise and the creditor a micro, small or medium one; art. 8 ust. 2: 30 days when the debtor is a public entity (60 for a public medical entity); art. 11a ust. 2: the debtor bears the burden of proving a longer term is not grossly unfair.
- **Grade:** 1 plain (whether a longer agreed term is "grossly unfair" is a court's judgment, but the 60-day default and the large-debtor cap are plain)
- **Rhea's default:** every invoice carries its due date (from the KSeF document's payment term, else the contractor's agreed days, else the person sets it); an agreed term over 60 days is flagged, and refused as a sales term when the company is micro, small or medium and the customer is a large enterprise (art. 7 ust. 2a).
- **Pack today:** `sales_invoice` and `purchase_invoice` have no due date; `contractor.payment_term_days` sits on a type the invoices do not reference; the KSeF reader (markets/pl/ksef_inbound.go) reads `TerminPlatnosci` into `payment_term`, but no rule maps it, so it is dropped. Missing — and everything that runs from a due date (interest, compensation, the 90-day bad-debt rules, ageing) cannot run.
- **Verdict:** missing
- **Deadline:** the due date itself.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2023/1790/text.pdf

### MON-21 — A business customer pays late: what interest and fixed compensation can be claimed?
- **Applies to:** everyone (transactions between businesses)
- **The law:** ustawa o przeciwdziałaniu nadmiernym opóźnieniom w transakcjach handlowych: art. 7 ust. 1 — interest is due „bez wezwania … za okres od dnia wymagalności świadczenia pieniężnego do dnia zapłaty” once the creditor has performed and was not paid on time; art. 4 pkt 3 lit. b — the rate is the NBP reference rate „i dziesięciu punktów procentowych” (plus eight only when the debtor is a public medical entity); art. 11b — the reference rate of 1 January for January–June and of 1 July for July–December; art. 11c — the minister announces it. Announced: 14% a year for 1 January–30 June 2026 (M.P. 2025 poz. 1257) and 13,75% for 1 July–31 December 2026 (M.P. 2026 poz. 642). Art. 10 ust. 1: compensation „bez wezwania” of „40 euro – gdy wartość świadczenia pieniężnego nie przekracza 5000 złotych; 70 euro – gdy … jest wyższa niż 5000 złotych, ale niższa niż 50 000 złotych; 100 euro – gdy … jest równa lub wyższa od 50 000 złotych”; ust. 1a — converted at the NBP average euro rate of „ostatniego dnia roboczego miesiąca poprzedzającego miesiąc, w którym świadczenie pieniężne stało się wymagalne”.
- **Grade:** 1 plain
- **Rhea's default:** for each overdue B2B sales invoice, compute the interest (rate per half-year as announced, kept as dated master data) and the compensation, and show them as a claim the owner may send; book nothing as revenue until the owner decides to charge and, for KPiR and ryczałt, until it is received (income-tax timing in MON-22).
- **Pack today:** missing (no due date — MON-20; no rate table).
- **Verdict:** missing
- **Deadline:** none (the claim runs from the day after the due date).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2023/1790/text.pdf; https://api.sejm.gov.pl/eli/acts/MP/2025/1257/text.pdf; https://api.sejm.gov.pl/eli/acts/MP/2026/642/text.pdf

### MON-22 — Late-payment interest charged to customers, or paid to suppliers and the tax office: when is it income or cost?
- **Applies to:** sole trader KPiR | sp. z o.o. full books (income tax); sp. z o.o. full books (books)
- **The law:** PIT art. 14 ust. 3 pkt 2 / CIT art. 12 ust. 4 pkt 2: not revenue are „kwot naliczonych, lecz nieotrzymanych odsetek od należności”; PIT art. 23 ust. 1 pkt 32 / CIT art. 16 ust. 1 pkt 11: not costs are „naliczonych, lecz niezapłaconych albo umorzonych odsetek od zobowiązań”; PIT art. 23 ust. 1 pkt 18 / CIT art. 16 ust. 1 pkt 21: never costs are „odsetek za zwłokę z tytułu nieterminowych wpłat należności budżetowych i innych należności, do których stosuje się przepisy … Ordynacja podatkowa” (whether this reaches ZUS late-payment interest was not verified). Ryczałt: the rate for interest received from customers was not verified.
- **Grade:** 1 plain
- **Rhea's default:** interest the company charges is income for tax only when received (the matching shows when); interest it pays to suppliers is a cost when paid; tax-arrears interest is never a cost; full books may accrue interest in the books under the company's policy, with the tax timing kept apart.
- **Pack today:** missing (no interest objects; MON-21).
- **Verdict:** missing
- **Deadline:** none
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf

### MON-23 — Netting a customer's invoice against his own invoice instead of paying: allowed, and what does it settle?
- **Applies to:** everyone
- **The law:** Kodeks cywilny (t.j. Dz.U. 2026 poz. 795) art. 498 § 1: either party „może potrącić swoją wierzytelność z wierzytelności drugiej strony, jeżeli przedmiotem obu wierzytelności są pieniądze … a obie wierzytelności są wymagalne”; § 2: „obie wierzytelności umarzają się nawzajem do wysokości wierzytelności niższej”; art. 499: „Potrącenia dokonywa się przez oświadczenie złożone drugiej stronie. Oświadczenie ma moc wsteczną od chwili, kiedy potrącenie stało się możliwe.” VAT art. 108a ust. 1d: split payment is not required „w zakresie, w jakim kwoty należności są potrącane” (KIS 0114-KDIP1-1.4012.1025.2025.2.MM of 4 February 2026, read only in a commercial database — secondary, verify: „potrącenie (kompensata) stanowi jedną z form zapłaty”, the remainder after set-off goes by split payment). CIT art. 15a ust. 7: for exchange differences the day of payment is „dzień uregulowania zobowiązań w jakiejkolwiek formie, w tym w wyniku potrącenia wierzytelności”.
- **Grade:** 3 open for one question only — whether a set-off above 15,000 zł counts as a "payment" under PIT art. 22p / CIT art. 15d: reading A, it is not a payment, so no cost is lost (tax-chamber interpretations reported from 2016–17, e.g. 3063-ILPB2.4510.156.2016.1.PS and 2461-IBPB-1-1.4510.367.2016 — secondary, verify in EUREKA); reading B, tax offices have treated it as a cash-like payment (a practitioner warning in rp.pl — secondary, verify). The civil law itself is grade 1.
- **Rhea's default:** netting is a door with the set-off statement as its evidence, dated as the statement says; it settles both invoices to the lower amount with one settlement record per side; for 15,000 zł+ transactions follow reading A but tell the owner the point is not settled.
- **Pack today:** the matching engine nets intercompany items, but the Poland pack has no set-off door and no rule; the date and the statement are not recorded. Missing.
- **Verdict:** missing
- **Deadline:** none
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/795/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/1263/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf

### MON-24 — A customer has not paid 90 days after the due date: can the income tax be reduced?
- **Applies to:** sole trader KPiR | sole trader ryczałt | sp. z o.o. full books
- **The law:** PIT art. 26i ust. 1 pkt 1 / CIT art. 18f ust. 1 pkt 1: the tax base „może być zmniejszona o zaliczaną do przychodów należnych wartość wierzytelności … która nie została uregulowana lub zbyta”, in the return for the year „w którym upłynęło 90 dni od dnia upływu terminu zapłaty określonego na fakturze (rachunku) lub w umowie”; already in the advances: PIT art. 44 ust. 17 pkt 1 / CIT art. 25 ust. 19 pkt 1 („począwszy od okresu rozliczeniowego, w którym upłynęło 90 dni”). Conditions (PIT art. 26i ust. 10 / CIT art. 18f ust. 10): the debtor is not in restructuring, bankruptcy or liquidation; „nie upłynęły 2 lata, licząc od końca roku kalendarzowego, w którym została wystawiona faktura”; both sides' income taxed in Poland; not between related parties (CIT art. 18f ust. 18). If paid later, the relief is reversed (art. 18f ust. 7 / PIT equivalent). Ryczałt: ustawa o ryczałcie art. 11 ust. 4 pkt 1 — revenue reduced in the annual return, same 90 days and conditions (ust. 10). Cash-basis PIT: not for receivables taxed on payment (PIT art. 44 ust. 25 pkt 1, art. 26i ust. 20–21; ryczałt art. 11 ust. 20). Unlike the VAT relief, the income-tax relief has no condition that both sides be active VAT payers. In force since 1 January 2020 (Dz.U. 2019 poz. 1649).
- **Grade:** 1 plain
- **Rhea's default:** from each sales invoice's due date compute the day 90 days later; on that day, if still open and the conditions hold, propose the reduction in the period's advance (KPiR, full books) or the annual return (ryczałt); reverse it when the payment arrives.
- **Pack today:** missing (no due date, see MON-20).
- **Verdict:** missing
- **Deadline:** the advance for the period in which the 90 days pass; the annual return.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/843/text.pdf

### MON-25 — The company has not paid a supplier 90 days after the due date: must it add the cost back?
- **Applies to:** sole trader KPiR | sole trader ryczałt | sp. z o.o. full books
- **The law:** PIT art. 26i ust. 1 pkt 2 / CIT art. 18f ust. 1 pkt 2: the tax base „podlega zwiększeniu o zaliczaną do kosztów uzyskania przychodów wartość zobowiązania … które nie zostało uregulowane”, 90 days after the due date; in the advances PIT art. 44 ust. 17 pkt 2 / CIT art. 25 ust. 19 pkt 2; reversed when paid (CIT art. 18f ust. 8). Ryczałt art. 11 ust. 4 pkt 2: revenue increased by unpaid liabilities „z tytułu zakupu towaru”. Not for costs on the cash method (PIT art. 44 ust. 25 pkt 2). The older debtor rule (PIT art. 24d, CIT art. 15b) is repealed — „(uchylony)” in the 2026 texts, since 1 January 2016 (Dz.U. 2015 poz. 1197).
- **Grade:** 1 plain
- **Rhea's default:** from each purchase invoice's due date, 90 days later, if still open: add the cost back in the period's advance (or the ryczałt annual return for goods), and reverse in the period it is paid; the matching already tells when it was paid.
- **Pack today:** missing (no due date).
- **Verdict:** missing
- **Deadline:** the advance for the period in which the 90 days pass.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/554/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/843/text.pdf

### ZUS-09 — May the sole trader skip one month of social contributions ("wakacje składkowe")?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** ustawa o sus art. 17a ust. 1 (in force since 1 November 2024, Dz.U. 2024 poz. 863): on request, the payer is exempt from his own pension, disability and accident contributions „za wskazany we wniosku miesiąc kalendarzowy”, if in the previous month he insured „niewięcej niż dziesięciu ubezpieczonych”, his revenue in at least one of the last 2 years did not exceed the PLN equivalent of 2 million euro (or he had none), and he does not work for a former employer; sickness contribution included only if voluntarily insured (ust. 3); once a year, de minimis aid (art. 17b); health contribution not exempt. Art. 36d ust. 1: the request „w miesiącu poprzedzającym miesiąc kalendarzowy, który ma być objęty zwolnieniem”, electronically (ZUS RWS per ZUS). Only for art. 8 ust. 6 pkt 1 persons — not for the sp. z o.o. shareholder.
- **Grade:** 1 plain
- **Rhea's default:** offer it once a year; the owner picks the month; Rhea files nothing itself but opens the task in the month before and drops the social part from that month's liability once ZUS confirms.
- **Pack today:** missing
- **Verdict:** later — a convenience, not a booking error
- **Deadline:** the request in the month before the chosen month.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/199/text/T/D20260199L.pdf; https://www.zus.pl/baza-wiedzy/biezace-wyjasnienia-komorek-merytorycznych/wakacje-skladkowe-pytania-i-odpowiedzi

### ZUS-10 — Mały ZUS Plus: contributions based on last year's income
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** ustawa o sus art. 18c ust. 1: for those whose „roczny przychód z pozarolniczej działalności gospodarczej uzyskany w poprzednim roku kalendarzowym nie przekroczył kwoty 120 000 złotych”; base = half the previous year's average monthly income (ust. 3–4), not above 60% of the forecast average wage nor below 30% of the minimum wage (ust. 2); not for those who ran the business less than 60 days the previous year (ust. 11 pkt 3); new from 1 January 2026 (Dz.U. 2025 poz. 769): at most „36 miesięcy kalendarzowych w każdym okresie 60 miesięcy kalendarzowych” (ust. 11a). Opt-in by ZUS ZUA/ZZA with code 05 90 by 31 January (ZUS; the statute's art. 36 ust. 4b covers January starters). Annual data in the January DRA (art. 18c ust. 9).
- **Grade:** 1 plain
- **Rhea's default:** for an existing sole trader moving to Rhea, compute eligibility and the base from last year's books each January and remind before 31 January; not offered in the first year of a new business (no previous year).
- **Pack today:** missing
- **Verdict:** later — not available in a new business's first year
- **Deadline:** 31 January.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/199/text/T/D20260199L.pdf; https://www.zus.pl/-/„mały-zus-plus-nowe-zasady-od-2026-r.

### MON-26 — Year end: are open euro receivables, payables and euro bank balances revalued?
- **Applies to:** sp. z o.o. full books (accounting); sole trader KPiR (no revaluation for tax)
- **The law:** UoR art. 30 ust. 1 pkt 1: „Nierzadziej niż na dzień bilansowy wycenia się wyrażone w walutach obcych: 1) składniki aktywów … i pasywów – po obowiązującym na ten dzień średnim kursie ogłoszonym dla danej waluty przez Narodowy Bank Polski”; ust. 4: the differences go to financial income or costs. For income tax only realised differences count (CIT art. 15a, PIT art. 24c), unless the audited-accounts method of CIT art. 9b ust. 1 pkt 2 was chosen (MON-04); KIS 0112-KDIL2-2.4011.576.2020.2.AA: the tax rules „nie nakładają na podatników obowiązku wyceny walut i ustalenia różnic kursowych na koniec roku podatkowego”.
- **Grade:** 1 plain
- **Rhea's default:** on the balance-sheet date, revalue every open foreign-currency item and account at that day's NBP average rate as a book-only (not tax) entry, reversed on the first day of the next year (reversal is common practice, not verified as required); nothing for KPiR or ryczałt.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** the balance-sheet date (31 December for a calendar year).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf; https://eureka.mf.gov.pl/informacje/podglad/423507

### MON-27 — Year end: count the cash, confirm the bank balances
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 26 ust. 1: on the last day of each financial year, inventory of „aktywów pieniężnych (z wyjątkiem zgromadzonych na rachunkach bankowych)” by physical count (pkt 1), and of assets „zgromadzonych na rachunkach bankowych” and receivables „drogą otrzymania od banków i uzyskania od kontrahentów potwierdzeń prawidłowości wykazanego w księgach rachunkowych jednostki stanu tych aktywów” (pkt 2); receivables from persons not keeping books and public-law ones by comparison with documents (pkt 3); art. 26 ust. 3 pkt 1 — the start-3-months-before / finish-by-15-January window excludes cash („z wyłączeniem aktywów pieniężnych”), so cash is counted on the last day itself. Differences are settled „w księgach rachunkowych tego roku obrotowego, na który przypadał termin inwentaryzacji” (art. 27 ust. 2). Records kept 5 years (art. 74 ust. 2 pkt 7).
- **Grade:** 1 plain
- **Rhea's default:** a year-end checklist: the cash count as a signed event against the cash ledger, the bank's closing balance from the last statement compared with account 130, balance confirmations sent to business customers (individuals and tax authorities by documents), every difference explained before the close.
- **Pack today:** missing (the MT940 reader already checks opening + lines = closing per statement, but nothing compares the statement balance with the books at year end).
- **Verdict:** missing
- **Deadline:** 31 December (cash count on that day).
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf

### ZUS-11 — How do the owner's contributions lower his income tax, by form?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** PIT art. 26 ust. 1 pkt 2 lit. a: deduct from income contributions „zapłaconych w roku podatkowym bezpośrednio na własne ubezpieczenia emerytalne, rentowe, chorobowe oraz wypadkowe”; liniowy the same (art. 30c ust. 2 pkt 1); or as a cost — the text allows it („jeżeli nie zostały zaliczone do kosztów uzyskania przychodów”, art. 26 ust. 13a, art. 30c ust. 3), and biznes.gov.pl: „zapłacone składki na ubezpieczenie społeczne możesz odliczyć od dochodu albo zaliczyć do kosztów uzyskania przychodów”. Health contribution: skala — neither cost nor deduction (art. 23 ust. 1 pkt 58 excludes it from costs except for liniowy); liniowy — cost or deduction, jointly capped: the act's „8700 zł” indexed yearly, for 2026 „14 100 zł” (Obwieszczenie MFiG, M.P. 2025 poz. 1274); ryczałt — ustawa o ryczałcie art. 11 ust. 1 (social contributions from revenue) and art. 11 ust. 1a: revenue „mogą być pomniejszone o 50 % składek na ubezpieczenie zdrowotne zapłaconych w roku podatkowym”. Always the year paid (December's contributions paid in January count in January's year).
- **Grade:** 1 plain — art. 26 ust. 13a and art. 30c ust. 3 make cost and deduction alternatives (never both); the government portal (biznes.gov.pl 00230) reads it the same way.
- **Rhea's default:** deduction from income (ryczałt: from revenue) by payment date, proposing cost treatment instead when the year heads for a loss so the amount survives in the loss carried forward — the owner decides; FP/FS as costs; the health contribution per form with the 2026 liniowy cap; ryczałt 50% of paid health contributions.
- **Pack today:** missing (no form of taxation, no PIT computation).
- **Verdict:** missing
- **Deadline:** monthly advances (the 20th) and the annual return by 30 April.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/592/text.pdf; https://api.sejm.gov.pl/eli/acts/MP/2025/1274/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2025/843/text.pdf; https://www.biznes.gov.pl/pl/portal/00230

### ZUS-12 — The annual health-contribution reconciliation: when, and who?
- **Applies to:** sole trader KPiR | sole trader ryczałt
- **The law:** ustawa o świadczeniach art. 81 ust. 2k: the annual amount „wykazywana jest w dokumencie rozliczeniowym składanym za miesiąc, w którym upływa termin złożenia zeznania” (PIT art. 45 ust. 1: 30 April) — i.e. the April DRA, due 20 May; ust. 2l — a top-up is paid by that month's deadline. Skala, liniowy and ryczałt reconcile; karta podatkowa and the sp. z o.o. shareholder do not. ZUS for 2025: „Do 20 maja 2026 r. … roczne rozliczenie składki … za 2025 r. … w … ZUS DRA lub ZUS RCA za kwiecień 2026 r.”, refund requests (RZS-R) by 1 June 2026. New DRA/RCA forms from the April 2026 settlement (Dz.U. 2026 poz. 518) add the choice to include fixed-asset sales in the annual base (art. 81 ust. 2ze).
- **Grade:** 1 plain
- **Rhea's default:** in April, recompute the year's health base from the closed books per form, compare with what was paid, and draft the April DRA with the top-up or the refund request.
- **Pack today:** missing
- **Verdict:** missing (the first reconciliation of a 2026 start falls in May 2027)
- **Deadline:** 20 May (April DRA); refund request by 1 June (2026 dates).
- **Sources:** https://eli.gov.pl/api/acts/DU/2025/1461/text/T/D20251461L.pdf; https://www.zus.pl/-/roczne-rozliczenie-składki-zdrowotnej-za-2025-r.; https://www.zus.pl/-/od-1-maja-2026-r.-obowiązują-nowe-wzory-formularzy-zus-dra-i-zus-rca

### MON-28 — How long must books, documents and bank statements be kept?
- **Applies to:** everyone
- **The law:** Ordynacja podatkowa art. 86 § 1: taxpayers keeping tax books „przechowują księgi i związane z ich prowadzeniem dokumenty do czasu upływu okresu przedawnienia zobowiązania podatkowego”; art. 70 § 1: „Zobowiązanie podatkowe przedawnia się z upływem 5 lat, licząc od końca roku kalendarzowego, w którym upłynął termin płatności podatku.” UoR art. 74 ust. 2: „księgi rachunkowe – 5 lat”, „pozostałe dowody księgowe … – 5 lat”, loans and trade contracts „przez 5 lat od początku roku następującego po roku obrotowym, w którym operacje … zostały ostatecznie zakończone, spłacone, rozliczone lub przedawnione”; ust. 3: counted „od początku roku następującego po roku obrotowym”; art. 74 ust. 1: approved financial statements at least 5 years from the year after approval. UoR art. 73a (in force 1 February 2026): invoices kept in KSeF are exempt from the unit's own keeping duties, but if KSeF's retention ends before the UoR periods, the unit keeps them itself until those periods end.
- **Grade:** 1 plain — two clocks run side by side: the tax clock (5 years from the end of the year the tax fell due; can be suspended or interrupted — art. 70 § 2 onwards, not read in detail here) and the accounting clock (5 years from the start of the following year).
- **Rhea's default:** never delete (the log is append-only anyway); compute for every document and book period the later of the two dates and show it; flag documents held only in KSeF when KSeF's retention would end first.
- **Pack today:** append-only log and blob store keep everything, so nothing is lost; no retention date is computed or shown. Right by construction; the computed date is missing (minor).
- **Verdict:** right
- **Deadline:** the computed retention end per document.
- **Sources:** https://api.sejm.gov.pl/eli/acts/DU/2026/622/text.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/522/text.pdf

### ZUS-13 — How long can ZUS claim unpaid contributions (and how long to keep the papers)?
- **Applies to:** everyone who pays contributions
- **The law:** ustawa o sus art. 24 ust. 4: „Należności z tytułu składek ulegają przedawnieniu po upływie 5 lat, licząc od dnia, w którym stały się wymagalne, z zastrzeżeniem ust. 5–6”; ust. 5h added from 8 July 2026 (Dz.U. 2026 poz. 473) suspends it during an appeal against a labour-inspectorate decision.
- **Grade:** 1 plain
- **Rhea's default:** keep ZUS declarations and payment evidence at least until 5 years after each contribution fell due (beside the tax clock in MON-28).
- **Pack today:** the append-only log keeps everything; no retention date shown. Right by construction.
- **Verdict:** right
- **Deadline:** none
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/199/text/T/D20260199L.pdf; https://api.sejm.gov.pl/eli/acts/DU/2026/473/text.pdf

## Full books and CIT

Scope: a small sp. z o.o. on full books (ustawa o rachunkowości, "UoR") paying CIT, in its first business year (2026), reviewed against `packs/pl/pack.json` v18, `packs/pl/warehouse.json` v1, `packs/common/periods.json` v1 and the base finance seeds (`testdata/seed/finance_v3…v9.json`).

Texts used: UoR consolidated text **Dz.U. 2026 poz. 522** (law as of 26 March 2026; the later amendment Dz.U. 2026 poz. 333 deals only with sustainability reporting); CIT act consolidated text **Dz.U. 2026 poz. 554** (law as of 18 March 2026); KSR 16 "Uproszczenia dopuszczone ustawą o rachunkowości" (Dz. Urz. MFiG 2025 poz. 29, first applies to years starting on or after 1 January 2026); updated KSR 2 "Podatek dochodowy" (Dz. Urz. MF 2024 poz. 81). Where a later change is known only from a Ministry page, the entry says so.

Entries run in the order a company meets them: founding, set-up, every month, during the year, year end, the following spring.

---

### FB-01 — Which year do our books and our tax run on, and can a company founded in the second half of the year make its first year longer?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 3 ust. 1 pkt 9: „rok kalendarzowy lub inny okres trwający 12 kolejnych pełnych miesięcy kalendarzowych, stosowany również do celów podatkowych. Rok obrotowy lub jego zmiany określa statut lub umowa, na podstawie której utworzono jednostkę. Jeżeli jednostka rozpoczęła działalność w drugiej połowie przyjętego roku obrotowego, to można księgi rachunkowe i sprawozdanie finansowe za ten okres połączyć z księgami rachunkowymi i sprawozdaniem finansowym za rok następny.” CIT art. 8 ust. 2a: „W przypadku podjęcia po raz pierwszy działalności w drugiej połowie roku kalendarzowego i wybrania roku podatkowego pokrywającego się z rokiem kalendarzowym, pierwszy rok podatkowy może trwać od dnia rozpoczęcia działalności do końca roku kalendarzowego następującego po roku, w którym rozpoczęto działalność.” CIT art. 8 ust. 5: a non-calendar tax year is announced in the CIT-8 for the first year.
- **Grade:** 1 plain
- **Rhea's default:** the calendar year unless the articles of association say otherwise. A company that starts after 30 June is asked one question: join the first months to next year? The answer is recorded as a fact. The year has to be a fact because every year-end close, the 9% revenue limit (which is pro-rated for a year that is not 12 months, see CIT-02) and the JPK_KR_PD file depend on it.
- **Pack today:** missing. There is no fiscal-year object, and the periods pack knows only months (`lock-period`, `month()`).
- **Verdict:** missing
- **Deadline:** no deadline of its own. A non-calendar tax year is reported in the first year's CIT-8 (CIT art. 8 ust. 5).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf

### FB-02 — Must a new sp. z o.o. keep full books from its very first transaction, and by when must the books be opened?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 2 ust. 1 pkt 1 applies the act to „spółek handlowych (osobowych i kapitałowych, w tym również w organizacji)”. Art. 12 ust. 1 pkt 1: books are opened „na dzień rozpoczęcia działalności, którym jest dzień pierwszego zdarzenia wywołującego skutki o charakterze majątkowym lub finansowym … – w ciągu 15 dni od dnia zaistnienia tych zdarzeń.” CIT art. 1 ust. 1 taxes „dochodów osób prawnych i spółek kapitałowych w organizacji”.
- **Grade:** 1 plain
- **Rhea's default:** a company that declares itself a capital company gets full books (the accrual rules on the statutory book) from its first event, including the period "w organizacji" before the KRS entry. It never gets the KPiR or ryczałt variant. The form of books and the form of tax are declared facts on the company, and they are the key the Poland pack branches on.
- **Pack today:** missing. `company` v4 has `legal_form` with only `natural_person | legal_entity`, and no field for the form of books or the form of tax. Every company gets the same `pl-stat` rules.
- **Verdict:** missing
- **Deadline:** books opened within 15 days of the first event with property or financial effect (UoR art. 12 ust. 1).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf

### CIT-01 — Classic CIT or the Estonian lump sum (ryczałt od dochodów spółek), and by when must we choose?
- **Applies to:** sp. z o.o. full books
- **The law:** CIT art. 28j ust. 1 pkt 7: the taxpayer must file „zawiadomienie o wyborze opodatkowania ryczałtem … w terminie do końca pierwszego miesiąca pierwszego roku podatkowego, w którym ma być opodatkowany ryczałtem”. The other conditions in art. 28j ust. 1:
  - pkt 4: the shareholders are „wyłącznie osoby fizyczne”;
  - pkt 5: the company holds no shares in other companies;
  - pkt 2: less than 50% of revenue is passive;
  - pkt 3: at least 3 employees in full-time equivalents. For a new company this „nie dotyczy roku rozpoczęcia tej działalności i 2 lat podatkowych bezpośrednio po nim następujących” (ust. 2 pkt 2).

  Rates (art. 28o ust. 1): „10 % podstawy opodatkowania – w przypadku małego podatnika oraz podatnika rozpoczynającego prowadzenie działalności”, 20% otherwise. Companies formed by transformation, or by contribution of a business worth more than 10 000 EUR, are excluded in their first year and the next, for at least 24 months (art. 28k ust. 1 pkt 5). The choice runs for 4-year periods that renew automatically (art. 28f).
- **Grade:** 1 plain
- **Rhea's default:** classic CIT unless the board declares the Estonian regime. The choice is a dated fact. Rhea raises a task before the end of the first month, because missing that date costs a whole year. The pack must assume neither regime.
- **Pack today:** missing. Nothing in the pack covers CIT. `tax_declaration.kind` lists `jpk_v7, vat7, pit, zus_dra, pcc3, other`, with no CIT-8 and no ZAW-RD.
- **Verdict:** missing
- **Deadline:** the ZAW-RD notice by the end of the first month of the first Estonian year.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf

### FB-03 — Who answers for the books when software drafts the rules, and could the hosted service count as the one keeping them?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - UoR art. 4 ust. 5: „Kierownik jednostki … ponosi odpowiedzialność za wykonywanie obowiązków w zakresie rachunkowości określonych ustawą, w tym z tytułu nadzoru, również w przypadku, gdy określone obowiązki w zakresie rachunkowości – z wyłączeniem odpowiedzialności za przeprowadzenie inwentaryzacji w formie spisu z natury – zostaną powierzone innej osobie lub przedsiębiorcy, o którym mowa w art. 11 ust. 2, za ich zgodą. Przyjęcie odpowiedzialności przez inną osobę lub przedsiębiorcę powinno być stwierdzone w formie pisemnej. W przypadku gdy kierownikiem jednostki jest organ wieloosobowy, a nie została wskazana osoba odpowiedzialna, odpowiedzialność ponoszą wszyscy członkowie tego organu.”
  - Art. 4a ust. 1–2: the board and any supervisory board must ensure the statements meet the act, and they are jointly liable to the company.
  - Art. 11 ust. 1: „Księgi rachunkowe są prowadzone przez jednostkę.” Under ust. 2 they may be entrusted only to an entrepreneur under art. 76a ust. 3. Under art. 76a ust. 1, bookkeeping as a service covers the activities in art. 4 ust. 3 pkt 2–6, and art. 76h ust. 1 requires civil-liability insurance for it.
  - Art. 11a: if the books are kept outside the company's seat, the tax office must be told „w terminie 15 dni od dnia ich wydania”.
  - Art. 77: „Kto wbrew przepisom ustawy dopuszcza do: 1) nieprowadzenia ksiąg rachunkowych, prowadzenia ich wbrew przepisom ustawy lub podawania w tych księgach nierzetelnych danych … – podlega grzywnie albo karze pozbawienia wolności do lat 2, albo obu tym karom łącznie.”
- **Grade:** 1 plain, as to who is responsible. Unverified: whether books held on a provider's server count as kept "outside the seat" under art. 11a. I found no authority either way.
- **Rhea's default:** the company keeps its own books and Rhea is its tool. Every rule that posts to the statutory book is approved by a named person of the company: a board member, or someone who accepted the duty in writing, that acceptance recorded as a fact. The agent and the hosting operator never approve. If Rhea, or a partner accountant, ever approves rules on a client's behalf, that is bookkeeping as a service: the art. 76a ust. 3 conditions, art. 76h insurance and the written acceptance under art. 4 ust. 5 all apply.
- **Pack today:** the kernel's gate (draft → approved → active, with the approval itself an event) does what the law needs. Not yet there: a record of who may approve rules for the statutory book. Today any human actor can.
- **Verdict:** right
- **Deadline:** 15 days to notify the tax office where the books are kept, if they are kept outside the seat (art. 11a).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-04 — What must our written accounting policy contain, who adopts it, and how much of it do Rhea's approved rules already say?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - UoR art. 10 ust. 1 requires documentation in Polish of: „1) określenia roku obrotowego i wchodzących w jego skład okresów sprawozdawczych; 2) metod wyceny aktywów i pasywów oraz ustalania wyniku finansowego; 3) sposobu prowadzenia ksiąg rachunkowych, w tym co najmniej: a) zakładowego planu kont … b) wykazu ksiąg rachunkowych, a przy prowadzeniu ksiąg rachunkowych przy użyciu komputera – wykazu zbiorów danych … c) … opisu systemu informatycznego, zawierającego wykaz programów, procedur lub funkcji … wraz z opisem algorytmów i parametrów oraz programowych zasad ochrony danych … a ponadto określenie wersji oprogramowania i daty rozpoczęcia jego eksploatacji; 4) systemu służącego ochronie danych i ich zbiorów…”
  - Ust. 2: „Kierownik jednostki ustala w formie pisemnej i aktualizuje dokumentację”.
  - Ust. 3: „W sprawach nieuregulowanych przepisami ustawy, przyjmując zasady (politykę) rachunkowości, jednostki mogą stosować krajowe standardy rachunkowości wydane przez Komitet Standardów Rachunkowości. W przypadku braku odpowiedniego standardu krajowego, jednostki … mogą stosować MSR.”
  - Art. 4 ust. 4 allows simplifications „jeżeli nie wywiera to istotnie ujemnego wpływu”, and art. 4 ust. 4a sets the materiality principle.
  - Art. 8 ust. 2: changes take effect from the first day of the year.
  - KSR 16 pt 4.2 lit. b: a simplification „powinno być przyjęte w zasadach (polityce) rachunkowości”. Chapter 4 of KSR 16 asks for materiality thresholds to be set and disclosed.
- **Grade:** 1 plain for what the policy must contain. The materiality thresholds inside it are grade 2, settled by KSR 16 (Dz. Urz. MFiG 2025 poz. 29).
- **Rhea's default:** the policy is a document rendered from facts:
  - the approved rule versions with their effective dates (these are the "algorytmy i parametry", the version and the "data rozpoczęcia eksploatacji" of art. 10 ust. 1 pkt 3 lit. c);
  - the chart of accounts, from the account events;
  - the year (FB-01);
  - the valuation methods, materiality thresholds and chosen simplifications, as declared parameters.

  The board adopts it in one approval event. Every later rule approval is a dated policy update. KSR applies wherever the UoR is silent.
- **Pack today:** the rules plus the top-level `warrant` ("basis": "practice") already state pkt 3 lit. a and c as data. Missing: the year (pkt 1), the valuation methods (pkt 2: fixed assets, stock, foreign currency, receivables), the materiality thresholds, and a rendered policy document adopted by the head of the entity. The `warrant` cites only the wzorcowy plan kont, VAT art. 106e and UoR art. 4 ust. 4.
- **Verdict:** missing
- **Deadline:** none in the statute. The policy must exist while books are kept, and changes take effect from the first day of a year (art. 8 ust. 2).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a ; https://www.gov.pl/web/finanse/krajowe-standardy-rachunkowosci

### FB-05 — Is the wzorcowy plan kont the law, and is the pack's chart a sound default for a small sp. z o.o.?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 10 ust. 1 pkt 3 lit. a requires the company's own chart (zakładowy plan kont). Art. 83 ust. 1: „W celu ujednolicenia zasad grupowania operacji gospodarczych i ograniczenia nakładu pracy związanego z ustaleniem zakładowych planów kont mogą być stosowane wzorcowe plany kont.” Ministerial model charts under ust. 2 exist only for banks, investment firms, funds, insurers and similar bodies.
- **Grade:** 1 plain. The model chart is customary, not statutory; only the company's own chart is required.
- **Rhea's default:** ship the model numbering as the default company chart, stated in the policy as the company's own. Complete it so that every account a small sp. z o.o. needs in year one exists. Customary numbers, to check against whichever published model is adopted:
  - CIT payable (220 or an analytic of 225) and social security (229);
  - employee and other settlements (234, 240), and shareholder loans (an analytic of 240 or 249);
  - fixed assets under construction (080);
  - prepaid costs (640);
  - cost of goods sold (731);
  - financial income and costs (750/751), other operating income and costs (760/761);
  - profit distribution (820), income tax (870).
- **Pack today:** 26 accounts: 010, 071, 100, 130, 201, 202, 221, 222, 225, 231, 300, 310, 330, 401–405, 409, 490, 700, 730, 756, 760, 801, 860. None of the accounts listed above exists. The numbering is sound; the set is not closed (see FB-06, FB-14, FB-28).
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.podatki.biz/artykuly/co-musisz-wiedziec-aby-dobrze-skonstruowac-zakladowy-plan-kont_23_61011.htm (secondary — verify, used only for the customary numbering)

### FB-06 — Where does a purchased service or expense land: on 300 for ever, or in costs?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 6 ust. 1: „W księgach rachunkowych jednostki należy ująć wszystkie osiągnięte, przypadające na jej rzecz przychody i obciążające ją koszty związane z tymi przychodami dotyczące danego roku obrotowego, niezależnie od terminu ich zapłaty.” Art. 4 ust. 2: transactions are recorded by their economic substance. Art. 24 ust. 2: books are reliable when entries reflect the real state („odzwierciedlają stan rzeczywisty”).
- **Grade:** 1 plain
- **Rhea's default:** a purchase books to its cost by nature (Wn 40x, by a cost type the rule reads from the document or from the supplier's default) or to an asset (310, 330, 080, 010). Account 300 is used only to clear goods or materials when the delivery note (PZ) and the invoice arrive separately, and it must come back to zero. A month-end check reports any balance left on 300.
- **Pack today:** wrong. `pl-post-purchase-invoice` books every purchase as "Wn 300 netto, Wn 221 / Ma 202". Only `pl-pz-line-posts` (warehouse pack: Wn 330 / Ma 300) ever credits 300. Services, rent, fuel and accounting fees therefore never leave 300. Cost group 4 stays empty, and the profit, the CIT advances and the CIT-8 are all overstated. Accounts 401–409 exist but no rule uses them.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-07 — Do we record costs only by nature (group 4) or also by function (group 5), and which profit-and-loss layout do we use?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 47 ust. 4 pkt 1: the profit-and-loss account is drawn up „w wariancie kalkulacyjnym albo porównawczym, zależnie od wyboru dokonanego przez kierownika jednostki”. Pkt 5 gives the same choice for small entities using Annex 5. Art. 10 ust. 1 pkt 2 puts the method of measuring the result into the policy.
- **Grade:** 1 plain
- **Rhea's default:** costs by nature only (group 4) and the comparative layout (porównawczy) for a services or trading company, recorded as a policy parameter. A manufacturer can add group 5 and account 490 later as a rule set.
- **Pack today:** missing. Accounts 401–409 and 490 exist, no rule posts to them (FB-06), and there is no profit-and-loss view.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-08 — How does a company open its books: the share capital, first balances, and a company moving to Rhea mid-life?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 5 ust. 1: „Wykazane w księgach rachunkowych na dzień ich zamknięcia stany aktywów i pasywów należy ująć w tej samej wysokości, w otwartych na następny rok obrotowy księgach rachunkowych.” Art. 36 ust. 1: „Kapitały (fundusze) własne ujmuje się w księgach rachunkowych z podziałem na ich rodzaje i według zasad określonych przepisami prawa, postanowieniami statutu lub umowy o utworzeniu jednostki.” Art. 18 ust. 1 pkt 2: the trial balance shows „salda kont na dzień otwarcia ksiąg rachunkowych”. Art. 19 ust. 1: a list of assets and liabilities, confirmed by a count, is required from units that did not previously keep books.
- **Grade:** 1 plain
- **Rhea's default:** an "open the books" door.
  - New company: the share capital as subscribed (Ma 801, against a receivable from shareholders until paid) and as paid (Wn 130).
  - Company moving in: the previous system's closing balances per account, entered as one opening fact. It must balance and must equal the prior closing (art. 5 ust. 1). Fixed assets carry their accumulated book and tax depreciation.
- **Pack today:** missing. There is no opening-balance door or rule. Account 801 exists, and `fixed_asset` already has `prior_tax_depreciation` and `prior_balance_depreciation` (the right shape).
- **Verdict:** missing
- **Deadline:** 15 days from the first event (UoR art. 12 ust. 1).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-09 — What must the system itself guarantee when the books are kept by computer?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - UoR art. 13 ust. 5: „Przy prowadzeniu ksiąg rachunkowych przy użyciu komputera należy zapewnić automatyczną kontrolę ciągłości zapisów, przenoszenia obrotów lub sald.”
  - Art. 13 ust. 6: „Księgi rachunkowe należy wydrukować niepóźniej niż na koniec roku obrotowego. Za równoważne z wydrukiem uznaje się przeniesienie treści ksiąg rachunkowych na informatyczny nośnik danych, zapewniający trwałość zapisu informacji, przez czas niekrótszy od wymaganego dla przechowywania ksiąg rachunkowych.”
  - Art. 13 ust. 4 pkt 1: the books must be marked with the company's name, the kind of book and „nazwą programu przetwarzania”.
  - Art. 23 ust. 1: „…należy stosować właściwe procedury i środki chroniące przed zniszczeniem, modyfikacją lub ukryciem zapisu.”
  - Art. 71 ust. 2: data protection includes „systematycznym tworzeniu rezerwowych kopii zbiorów danych”.
  - Art. 72 ust. 1: books may be kept as data sets on electronic media if art. 71 ust. 2 is met.
- **Grade:** 1 plain
- **Rhea's default:** the append-only log is the book. At year end the yearly export (the JPK_KR_PD file and the trial balance) is written to the document store under its hash. That export is the "transfer to a durable medium". Backups run on a schedule. Every printed book view carries the company, the kind of book, the program name and version, the period and the print date.
- **Pack today:** the kernel already has the core: the append-only event log and definitions (invariant 1), deterministic replay, and backup/restore (commit 3ce7595). Gaps: no scheduled backup, no year-end durable export, and no header marks on the journal and trial-balance views.
- **Verdict:** right
- **Deadline:** books printed, or transferred to a durable medium, by the end of the financial year (art. 13 ust. 6).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-10 — Is a rule-generated entry backed by a valid accounting document, and who is "responsible" for it?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - UoR art. 20 ust. 2: „Podstawą zapisów w księgach rachunkowych są dowody księgowe stwierdzające dokonanie operacji gospodarczej”. Internal documents count (pkt 3).
  - Art. 20 ust. 5: „Przy prowadzeniu ksiąg rachunkowych przy użyciu komputera za równoważne z dowodami źródłowymi uważa się zapisy w księgach rachunkowych, wprowadzane automatycznie … lub tworzone według algorytmu (programu) na podstawie informacji zawartych już w księgach, przy zapewnieniu, że … 1) uzyskają one trwale czytelną postać zgodną z treścią odpowiednich dowodów księgowych; 2) możliwe jest stwierdzenie źródła ich pochodzenia oraz ustalenie osoby odpowiedzialnej za ich wprowadzenie; 3) stosowana procedura zapewnia sprawdzenie poprawności przetworzenia odnośnych danych oraz kompletności i identyczności zapisów; 4) dane źródłowe w miejscu ich powstania są odpowiednio chronione, w sposób zapewniający ich niezmienność…”
  - Art. 21 ust. 1a pkt 2: the coding stamp (dekretacja) may be omitted „jeżeli wynika to z techniki dokumentowania zapisów księgowych”.
  - Art. 73a (in force since 1 February 2026): the storage duties do not apply to „dowodów księgowych przechowywanych w Krajowym Systemie e-Faktur”.
- **Grade:** 1 plain
- **Rhea's default:** each posting's provenance is the art. 20 ust. 5 chain:
  - the event id leads to the document, stored under its hash;
  - the rule id and version are the coding (dekretacja);
  - the approval event of that rule version names the responsible person, and that person is shown on the entry.

  Human doors such as `close_period` and `identify_company` are internal documents.
- **Pack today:** right. Provenance is invariant 5. Received and issued documents are kept under their hash (`ksef_submission` keeps `invoice_xml`, and the documents pack keeps files). Replay proves identical processing, which is what art. 20 ust. 5 pkt 3 asks.
- **Verdict:** right
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-11 — Do our journal entries carry what an entry must carry, and are they numbered without gaps?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - UoR art. 14 ust. 2: „Zapisy w dzienniku muszą być kolejno numerowane, a sumy zapisów (obroty) liczone w sposób ciągły.”
  - Art. 14 ust. 4: „Przy prowadzeniu ksiąg rachunkowych przy użyciu komputera zapis księgowy powinien posiadać automatycznie nadany numer pozycji, pod którą został wprowadzony do dziennika, a także dane pozwalające na ustalenie osoby odpowiedzialnej za treść zapisu.”
  - Art. 23 ust. 2: an entry contains at least the date of the transaction, the kind and number of the document (and its date if different), „zrozumiały tekst, skrót lub kod opisu operacji”, the amount and date of the entry, and the accounts.
  - Art. 23 ust. 3: foreign-currency entries show both currencies.
- **Grade:** 1 plain
- **Rhea's default:** the statutory journal numbers entries per book and per year, in log order. Each posting carries the kind and number of its document, the document date, a description, the entry date and the approver of the rule. The journal view shows running totals.
- **Pack today:** missing. `posting` v3 has `entry, line, book, account, side, amount, currency, date, tx_amount, tx_currency`. It has no sequential journal number, no document kind or number, no description, and no responsible person (reachable only through provenance). The base `journal` view lists postings without running totals. The JPK_KR_PD journal section needs exactly these fields (MF brochure, section 7).
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.podatki.gov.pl/media/if5hycxm/broszura_informacyjna-dotycz%C4%85ca-struktury-jpk_kr_pd-1-26082024.pdf

### CIT-02 — Do we pay 9% or 19% CIT this year?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - CIT art. 19 ust. 1 pkt 2, worded by Dz.U. 2025 poz. 1658 and in force since 1 January 2026: „9 % podstawy opodatkowania od przychodów (dochodów) innych niż z zysków kapitałowych – w przypadku podatników, u których przychody osiągnięte w roku podatkowym nie przekroczyły wyrażonej w złotych kwoty odpowiadającej równowartości 2 000 000 euro … a) kwotę 2 000 000 euro przelicza się na złote według średniego kursu euro ogłaszanego przez Narodowy Bank Polski na pierwszy dzień roboczy roku podatkowego, w zaokrągleniu do 1000 zł, b) jeżeli rok podatkowy podatnika był dłuższy albo krótszy niż 12 miesięcy, kwotę 2 000 000 euro ustala się jako iloczyn 1/12 tej kwoty i liczby pełnych miesięcy w tym roku podatkowym”.
  - Ust. 1d: the 9% rate requires „status małego podatnika”.
  - Ust. 1e: „Warunek, o którym mowa w ust. 1d, nie dotyczy podatników rozpoczynających prowadzenie działalności, z zastrzeżeniem ust. 1a i 1c, w roku rozpoczęcia tej działalności.”
  - Ust. 1a: companies formed by transformation, or by contribution of a business worth more than 10 000 EUR, are excluded in their first year and the next.
  - Art. 4a pkt 10: a small taxpayer is one whose prior-year sales including VAT did not exceed 2 000 000 EUR.

  For 2026 (podatki.gov.pl): the small-taxpayer limit is 8 517 000 zł (EUR rate 4,2586 of 1 October 2025), and the in-year 9% limit is 8 431 000 zł (EUR rate 4,2156 of 2 January 2026).
- **Grade:** 1 plain
- **Rhea's default:** the status is computed from facts.
  - First tax year: 9%, unless the company was formed by transformation or contribution. Onboarding asks that one question.
  - Later years: the small-taxpayer test on prior-year gross sales.
  - In-year revenue is watched against the limit, pro-rated for a short or long first year.
  - Capital-gains income is always taxed at 19%.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf ; https://www.podatki.gov.pl/podatki-firmowe/cit/cit-klasyczny/stawki-i-limity

### CIT-03 — Monthly or quarterly CIT advances, at which rate, and how are they booked?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - CIT art. 25 ust. 1a: „Zaliczki miesięczne … podatnik wpłaca w terminie do 20 dnia każdego miesiąca za miesiąc poprzedni”.
  - Ust. 1b: „Podatnicy rozpoczynający działalność, w pierwszym roku podatkowym … oraz mali podatnicy, mogą wpłacać zaliczki kwartalne”. Quarterly advances are due by the 20th of the month after the quarter (ust. 1c), and the choice is reported in the CIT-8 (ust. 1e).
  - Ust. 1f, worded from 1 January 2026: „…przy obliczaniu zaliczek miesięcznych i kwartalnych podatnicy, o których mowa w art. 19 ust. 1 pkt 2 i 5, stosują stawkę podatkową określoną odpowiednio w art. 19 ust. 1 pkt 1, 3 albo 4.”
  - Ust. 1g: start-ups in their first year, and small taxpayers, may use 9% for the months or quarters in which cumulative revenue stays within 2 000 000 EUR.
  - Ust. 6: simplified advances of 1/12 of the last return's tax. Ust. 8: not available to companies that started in the previous or the current year.
  - On the accounting side, UoR art. 37 ust. 8 makes the income tax charge a current part plus a deferred part.
- **Grade:** 1 plain
- **Rhea's default:** monthly advances. Quarterly only if the board chooses it. The advance (cumulative tax on cumulative taxable income, less advances already due) is computed by a rule on the month close as a fact with provenance, the same pattern as `pl-settle-vat`. It is booked Wn 870 / Ma 220 (customary numbers), and the payment is matched by the bank rule.
- **Pack today:** missing. There is no CIT advance rule, no income tax account and no CIT payable account.
- **Verdict:** missing
- **Deadline:** the 20th of the following month (monthly) or of the month after the quarter (quarterly). The last period's advance is due by the 20th of the first month of the next year, unless the return is filed and the tax paid first.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf ; https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-12 — On which date does a sale become revenue: the invoice date or the date of sale?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 6 ust. 1: revenue and costs „dotyczące danego roku obrotowego, niezależnie od terminu ich zapłaty”. Art. 20 ust. 1: „Do ksiąg rachunkowych okresu sprawozdawczego należy wprowadzić, w postaci zapisu, każde zdarzenie, które nastąpiło w tym okresie sprawozdawczym.” Art. 23 ust. 2 pkt 1–2: an entry carries „datę dokonania operacji gospodarczej”, plus the document date where it differs.
- **Grade:** 1 plain
- **Rhea's default:** revenue and the receivable are dated at the date of sale or completion stated on the invoice (FA(3) carries it when it differs from the issue date). The issue date is used only when no sale date is given. The year-end cut-off reads the sale date.
- **Pack today:** wrong. `sales_invoice` v5 has no sale-date field. `pl-book-sales-invoice` and `pl-post-sales-invoice` key revenue to `issue_date`, so a December service invoiced on 5 January lands in the wrong year.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-13 — At which rate does a foreign-currency invoice or payment enter the books, and where do realised exchange differences go?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 30 ust. 2: „Wyrażone w walutach obcych operacje gospodarcze ujmuje się w księgach rachunkowych na dzień ich przeprowadzenia – … odpowiednio po kursie: 1) faktycznie zastosowanym w tym dniu, wynikającym z charakteru operacji – w przypadku sprzedaży lub kupna walut oraz zapłaty należności lub zobowiązań; 2) średnim ogłoszonym dla danej waluty przez Narodowy Bank Polski z dnia poprzedzającego ten dzień – w przypadku zapłaty należności lub zobowiązań, jeżeli nie jest zasadne zastosowanie kursu, o którym mowa w pkt 1, a także w przypadku pozostałych operacji.” Ust. 4: differences „przy zapłacie należności i zobowiązań w walutach obcych … zalicza się odpowiednio do przychodów lub kosztów finansowych”. Art. 9: books are kept „w walucie polskiej”. Art. 23 ust. 3: entries show both currencies.
- **Grade:** 1 plain
- **Rhea's default:**
  - Invoices: the NBP average rate of the business day before the transaction date. The rate source must publish that rate under each date.
  - Bank operations on a currency account: the bank's actual rate (pkt 1), or NBP D-1 when the bank gives none.
  - On settlement, the difference between the invoice's PLN and the payment's PLN is posted to 750/751.
- **Pack today:** wrong.
  - `pl-post-sales-invoice` and `pl-post-purchase-invoice` convert to PLN at the `fx_rate` keyed by the issue or invoice date. The kernel leaves the D-1 lag to "the rate publisher", but no NBP publisher exists in the repository.
  - `pl-post-incoming-payment` and `pl-post-outgoing-payment` have no `convert` clause, so a EUR payment posts EUR amounts into the PLN book (art. 9), and a realised difference is never booked.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-14 — Where do penny differences on settlements and conversion roundings go?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 3 ust. 1 pkt 32 lit. d places among other operating costs and income („pozostałych kosztach i przychodach operacyjnych”) those connected „z odpisaniem należności i zobowiązań przedawnionych, umorzonych, nieściągalnych”. Art. 30 ust. 4 sends exchange differences to financial income or costs. Art. 4 ust. 4 allows simplifications in the policy. Art. 7 ust. 3: „Nie można kompensować ze sobą wartości różnych co do rodzaju aktywów i pasywów, przychodów i kosztów związanych z nimi…”
- **Grade:** 1 plain for exchange roundings (financial items, art. 30 ust. 4). For tolerated settlement shortfalls, the text (pkt 32 lit. d: a written-off, forgiven receivable) points to other operating items. That is my reading: no KSR or Ministry position on penny differences was found.
- **Rhea's default:** settlement penny differences within the policy tolerance (2 gr, declared as a materiality simplification) post to 761 (short) or 760 (over). Exchange roundings and differences post to 751/750. Gains and losses sit in separate accounts so the statement shows them gross.
- **Pack today:** wrong. One account, 756 "Różnice kursowe i zaokrąglenia" (typed expense), takes three things: short differences (`pl-post-short-difference`, Wn 756), over differences (`pl-post-over-difference`, Ma 756) and the conversion rounding plug. In the model chart 75x is the financial group, so settlement shortfalls land among financial items, and gains and losses net inside one account.
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### CIT-04 — Do our payments meet the tax conditions (bank account, VAT white list, split payment)?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - CIT art. 15d ust. 1: „Podatnicy nie zaliczają do kosztów uzyskania przychodów kosztu w tej części, w jakiej płatność dotycząca transakcji określonej w art. 19 ustawy … – Prawo przedsiębiorców: 1) została dokonana bez pośrednictwa rachunku płatniczego lub 2) została dokonana przelewem na rachunek inny niż zawarty na dzień zlecenia przelewu w wykazie podmiotów, o którym mowa w art. 96b ust. 1 ustawy o podatku od towarów i usług …, lub 3) … została dokonana z pominięciem mechanizmu podzielonej płatności”.
  - Ust. 4 pkt 1: the safe harbour is a notice to the tax office within 7 days.
  - Prawo przedsiębiorców art. 19: a payment must go through a payment account where the counterparty is an entrepreneur and „jednorazowa wartość transakcji, bez względu na liczbę wynikających z niej płatności, przekracza 15 000 zł”.
- **Grade:** 1 plain
- **Rhea's default:** an outgoing payment above 15 000 zł to an active VAT payer is checked against the white list on the order date, and the check result is a fact. A miss raises a task (notice within 7 days) and marks the cost non-deductible.
- **Pack today:** missing. `pl-match-outgoing` matches by IBAN only; there is no white-list check and no split-payment flag.
- **Verdict:** missing
- **Deadline:** the notice within 7 days of ordering the transfer (CIT art. 15d ust. 4 pkt 1).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf ; https://eli.gov.pl/api/acts/DU/2025/1480/text/T/D20251480L.pdf

### FB-15 — Must we produce a trial balance every month, and when do our books count as current?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 18 ust. 1: „…sporządza się na koniec każdego okresu sprawozdawczego, nierzadziej niż na koniec miesiąca, zestawienie obrotów i sald, zawierające: 1) symbole lub nazwy kont; 2) salda kont na dzień otwarcia ksiąg rachunkowych, obroty za okres sprawozdawczy i narastająco od początku roku obrotowego oraz salda na koniec okresu sprawozdawczego; 3) sumę …”. Art. 24 ust. 5: „Księgi rachunkowe uznaje się za prowadzone bieżąco, jeżeli: 1) pochodzące z nich informacje umożliwiają sporządzenie w terminie … deklaracji podatkowych oraz dokonanie rozliczeń finansowych; 2) zestawienia obrotów i sald kont księgi głównej są sporządzane … nierzadziej niż na koniec miesiąca, w terminie, o którym mowa w pkt 1, a za rok obrotowy – niepóźniej niż do 85 dnia po dniu bilansowym; 3) ujęcie wpłat i wypłat gotówką … następuje w tym samym dniu”.
- **Grade:** 1 plain
- **Rhea's default:** the month close (`period.closed`) produces the trial balance as a fact: opening balance, month turnover, year-to-date turnover and closing balance per account, with totals equal to the journal. It is ready before the CIT advance of the 20th and the VAT file of the next month. The annual trial balance is ready by day 85.
- **Pack today:** missing. The base `trial-balance` view sums every posting ever made: no opening balance, no period, no year-to-date. Nothing produces it on the close.
- **Verdict:** missing
- **Deadline:** monthly, in time for that month's tax settlements (the CIT advance is due on the 20th). The annual trial balance is due by the 85th day after the balance date: 26 March 2027 for FY2026 (my arithmetic).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-16 — What happens to a cost invoice that arrives after we closed the month?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - UoR art. 20 ust. 1: every event of a period goes into that period's books.
  - Art. 21 ust. 1 pkt 6: the coding assigns the month („przez wskazanie miesiąca oraz sposobu ujęcia dowodu w księgach rachunkowych”).
  - Art. 24 ust. 3: books are free of error when „wprowadzono do nich kompletnie i poprawnie wszystkie zakwalifikowane do zaksięgowania w danym miesiącu dowody księgowe”.
  - Art. 25 ust. 2 provides for corrections after a month is closed.
  - Art. 12 ust. 5: „Zamknięcie ksiąg rachunkowych polega na nieodwracalnym wyłączeniu możliwości dokonywania zapisów księgowych…”. This closing is at year end (art. 12 ust. 2 pkt 1), within 3 months.
- **Grade:** 1 plain
- **Rhea's default:** the monthly close locks the month against ordinary postings, but a late document is never refused. It goes to the first open month, keeping its transaction date. If it is material under the policy threshold, it goes through a correction door for the closed month instead. December stays open until the year's cut-off is done and the statement is drawn up.
- **Pack today:** wrong. `lock-period` makes the kernel refuse any posting dated in a locked month of `pl-stat`; the story test shows "the late invoice refused by the lock". The document is then never booked, which leaves the books incomplete (art. 24 ust. 3), and in December leaves it out of the year (art. 6).
- **Verdict:** wrong
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-17 — How do we correct a wrong entry?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 25 ust. 1 pkt 2: correction „przez wprowadzenie do ksiąg rachunkowych dowodu zawierającego korekty błędnych zapisów, dokonywane tylko zapisami dodatnimi albo tylko ujemnymi.” Ust. 2: „W razie ujawnienia błędów po zamknięciu miesiąca lub prowadzenia ksiąg rachunkowych przy użyciu komputera, dozwolone są tylko korekty dokonane w sposób określony w ust. 1 pkt 2.” Art. 54 ust. 3: a material error of a prior year goes to equity as „zysk (strata) z lat ubiegłych”.
- **Grade:** 1 plain
- **Rhea's default:** a correction is a new event, i.e. a new document. Its rule posts the reversal and then the right entry; nothing is ever edited. The policy names the reversal style: negative amounts (red storno) or opposite-side entries (black storno).
- **Pack today:** right. The log is append-only, unpairing posts a `reversal_of`, and amendments are new events. Not yet stated: the reversal style, and the prior-year error path (needed later).
- **Verdict:** right
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-18 — Must we spread prepaid costs (insurance, subscriptions) and accrue costs not yet invoiced?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 39 ust. 1: „Jednostki dokonują czynnych rozliczeń międzyokresowych kosztów, jeżeli dotyczą one przyszłych okresów sprawozdawczych.” Ust. 2 requires accruals for probable obligations of the period. Art. 6 ust. 2: matching. Art. 39 ust. 6 lets small entities skip accruals for future employee benefits, but ust. 7 says that „nie stosuje się do: … 2) spółek kapitałowych”. KSR 16 pt 12.10: a materiality threshold below which spreading is unnecessary (example 12.1 uses 3 500 zł). Example 12.2: a year's insurance paid in December may be charged 1/12 or moved wholly to next year.
- **Grade:** 2 settled. The duty is plain; the threshold below which spreading is skipped is settled by KSR 16 (Dz. Urz. MFiG 2025 poz. 29) pt 12.9–12.10.
- **Rhea's default:** a prepaid cost above the policy threshold books to 640, and a rule releases it monthly over the period the document states. Below the threshold it goes straight to cost. Month-end accruals only for material, known items.
- **Pack today:** missing. There is no account 640 and no spreading rule, and `purchase_invoice` has no service-period field.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a ; https://www.gov.pl/attachment/bc48b7e4-7102-4458-bdd9-545997b2e38b

### FB-19 — Do we keep stock records for goods, or expense them at purchase, and what books the cost of goods sold?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 17 ust. 2: „Kierownik jednostki … podejmuje decyzję o stosowaniu jednej z następujących metod …: 1) ewidencję ilościowo-wartościową …; 2) ewidencję ilościową …; 3) ewidencję wartościową …; 4) odpisywania w koszty wartości materiałów i towarów na dzień ich zakupu … połączone z ustalaniem stanu tych składników aktywów i jego wyceny oraz korekty kosztów o wartość tego stanu, niepóźniej niż na dzień bilansowy.” KSR 16 pt 11.10–11.11: expensing at purchase is meant for items such as fuel or office supplies, and is not recommended for core production materials.
- **Grade:** 1 plain. The head of the entity chooses.
- **Rhea's default:** a services company expenses minor materials at purchase (pkt 4) and counts them at year end. A trading company keeps quantity-and-value records (delivery in, issue out) and books the cost of goods sold (Wn 731 / Ma 330) at each issue. The method is declared per stock group in the policy.
- **Pack today:** missing. The warehouse pack posts deliveries Wn 330 / Ma 300. Every sale credits 700 (products), never 730 (goods). Nothing ever moves 330 to cost: there is no account 731 and no issue rule.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a

### FB-20 — What counts as a fixed asset, and what is our low-value threshold?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 3 ust. 1 pkt 15: fixed assets are „rzeczowe aktywa trwałe i zrównane z nimi, o przewidywanym okresie ekonomicznej użyteczności dłuższym niż rok, kompletne, zdatne do użytku i przeznaczone na potrzeby jednostki”. Art. 32 ust. 6 allows a simplified or one-off write-off „dla środków trwałych o niskiej jednostkowej wartości początkowej”. KSR 16 pt 8.2–8.3: the threshold is set in the policy, with regard to KSR 11 pt 4.29 and tax rules. CIT art. 16f ust. 3: assets up to 10 000 zł may be written off at once.
- **Grade:** 2 settled. The threshold practice is settled by KSR 16 pt 8.2–8.4 and KSR 11.
- **Rhea's default:** a policy threshold of 10 000 zł net, the same as CIT art. 16f ust. 3, so that books and tax agree. Below it: cost, or a one-off write-off kept in a quantity register. Above it: account 010 (through 080 if built in-house), with an asset card (`fixed_asset`) created from the purchase invoice.
- **Pack today:** missing. The `fixed_asset` type exists (number, KŚT class, rates, dates), but no rule creates one from a purchase, and there is no account 080.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a ; https://www.gov.pl/attachment/92b030f0-544b-4114-b877-f050ea186ebf ; https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf

### FB-21 — Book depreciation or tax depreciation: may a small sp. z o.o. simply use the tax rates?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - UoR art. 32 ust. 3: „Na dzień przyjęcia środka trwałego do używania należy ustalić okres lub stawkę i metodę jego amortyzacji. Poprawność stosowanych okresów i stawek amortyzacji środków trwałych powinna być przez jednostkę okresowo weryfikowana…”
  - Art. 32 ust. 7: „Jednostka mikro i jednostka mała mogą dokonywać odpisów amortyzacyjnych lub umorzeniowych od środków trwałych według zasad określonych w przepisach podatkowych.”
  - Art. 32 ust. 8: „Przepisu ust. 7 nie stosuje się do: … 2) spółek kapitałowych”. Ust. 7–8 are worded by Dz.U. 2024 poz. 1863 and apply first to statements for years starting after 31 December 2023 (art. 14 ust. 3 of that act).
  - KSR 16 example 8.4(1): „Spółka z o.o., która jest jednostką małą. Rozwiązanie nieprawidłowe.” Example 8.4(2): whether tax rates may be used is judged per asset.
  - CIT art. 16k ust. 7: in the start-up year, and for small taxpayers, a one-off tax write-off for KŚT groups 3–8 (passenger cars excluded), up to 50 000 EUR a year. That is 213 000 zł for 2026 (podatki.gov.pl).
- **Grade:** 1 plain. KSR 16 confirms.
- **Rhea's default:** each asset has two schedules.
  - Book: the economic life is set per asset when it is put into use. Tax rates may be used only where that assessment matches them.
  - Tax: the KŚT rate or a one-off write-off. This schedule feeds the CIT base and the JPK_ST_KR fixed-asset file.

  Depreciation posts monthly from the book schedule, Wn 401 / Ma 071.
- **Pack today:** missing. `fixed_asset` has the right shape (`rate`, `coefficient`, `prior_tax_depreciation`, `prior_balance_depreciation`), but there is no depreciation rule; CORPUS lists it only as rule intent 5.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a ; https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf ; https://www.podatki.gov.pl/podatki-firmowe/cit/cit-klasyczny/stawki-i-limity

### FB-22 — Is our car lease an operating lease or a finance lease in the books?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 3 ust. 4 lists seven conditions; meeting any one makes the lessee carry the asset. Art. 3 ust. 6: „Jednostka mikro i jednostka mała mogą dokonywać kwalifikacji umów, o których mowa w ust. 4, według zasad określonych w przepisach podatkowych i nie stosować przepisów ust. 4 i 5.” Capital companies are not excluded here. KSR 16 pt 5.4 c lists the form-based exclusions, and they cover only art. 7 ust. 2b, art. 32 ust. 7 and art. 39 ust. 6.
- **Grade:** 1 plain
- **Rhea's default:** a micro or small sp. z o.o. classifies leases by the tax rules (art. 3 ust. 6), declared in the policy, so books and tax share one classification. Operating-lease instalments go to costs, subject to the CIT car limits (CIT-05).
- **Pack today:** missing. The `lease_contract` type exists with no rules.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a ; https://www.gov.pl/attachment/07da18b3-1ae8-4fde-9239-7e264ed39e24

### CIT-05 — How much of the company car's cost is tax-deductible?
- **Applies to:** sp. z o.o. full books
- **The law:** CIT art. 16 ust. 1 pkt 4: depreciation is not deductible on the part of the car's value above „a) 225 000 zł – w przypadku samochodu osobowego będącego pojazdem elektrycznym … b) 150 000 zł – jeśli emisja CO2 silnika spalinowego … wynosi mniej niż 50 g na kilometr, c) 100 000 zł – jeśli emisja CO2 silnika spalinowego … jest równa lub wyższa niż 50 g na kilometr”. Pkt 51 excludes „25 % poniesionych wydatków … z tytułu kosztów używania samochodu osobowego … – jeżeli samochód osobowy jest wykorzystywany również do celów niezwiązanych z działalnością gospodarczą”. The 100 000 zł tier applies to cars entered into the register from 1 January 2026, and lease payments face matching limits under art. 16 ust. 1 pkt 49a (secondary — verify both).
- **Grade:** 1 plain
- **Rhea's default:** the vehicle record carries its CO2 band and its use: business-only with a mileage log, or mixed. Rules mark the non-deductible part of depreciation, lease and running costs as a tax difference, not as a book entry.
- **Pack today:** missing. `vehicle` and `mileage_entry` exist with no rules and no CO2 field.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf ; https://akademialtca.pl/blog/zmiana-limitow-amortyzacji-samochodow-od-2026-r-pozostaje-aktualna (secondary — verify)

### CIT-06 — Which costs in the books are not tax costs (or not yet), and how does the system keep the two apart?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - CIT art. 16 ust. 1 excludes, among others:
    - pkt 15: „podatku dochodowego”;
    - pkt 18: fines and penalties;
    - pkt 27: „rezerw … oraz odpisów aktualizujących … jeżeli obowiązek ich tworzenia w ciężar kosztów nie wynika z innych ustaw”;
    - pkt 28: „kosztów reprezentacji, w szczególności poniesionych na usługi gastronomiczne, zakup żywności oraz napojów, w tym alkoholowych”;
    - pkt 57: unpaid wages and contract fees.
  - Art. 15 ust. 4e: „Za dzień poniesienia kosztu uzyskania przychodów … uważa się dzień, na który ujęto koszt w księgach rachunkowych (zaksięgowano) na podstawie otrzymanej faktury (rachunku) … z wyjątkiem sytuacji gdy dotyczyłoby to ujętych jako koszty rezerw albo biernych rozliczeń międzyokresowych kosztów.”
  - Art. 18f ust. 1: in the yearly return the base is reduced for unpaid receivables and „podlega zwiększeniu o zaliczaną do kosztów uzyskania przychodów wartość zobowiązania … które nie zostało uregulowane … w zeznaniu podatkowym składanym za rok podatkowy, w którym upłynęło 90 dni od dnia upływu terminu zapłaty określonego na fakturze”.
- **Grade:** 1 plain
- **Rhea's default:** every cost account, or every rule, carries a tax treatment: deductible, non-deductible, or deductible on payment. The same marker is the account's CIT marker (S_12_3) in JPK_KR_PD. The CIT base is the book result plus or minus the marked differences, computed at each advance. Invoices carry their due date, so the 90-day adjustment runs off facts.
- **Pack today:** missing. Accounts carry no tax marker, and neither `sales_invoice` nor `purchase_invoice` has a due date, so the 90-day rule cannot be computed.
- **Verdict:** missing
- **Deadline:** the 90-day adjustments are made in the yearly return (art. 18f ust. 1).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf

### FB-23 — What must we count or confirm at year end, when, and who may do it?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - UoR art. 26 ust. 1, at the last day of the year:
    - pkt 1: a physical count („spisu ich ilości z natury”) of cash, stock and fixed assets;
    - pkt 2: confirmations from banks and counterparties of bank balances and receivables („potwierdzeń prawidłowości wykazanego w księgach … stanu”);
    - pkt 3: verification against documents for disputed receivables, public-law balances and the like.
  - Art. 26 ust. 3 pkt 1: the dates count as met if the count „rozpoczęto niewcześniej niż 3 miesiące przed końcem roku obrotowego, a zakończono do 15 dnia następnego roku”. Cash, and stock expensed under art. 17 ust. 2 pkt 4, are excluded from that window. Under pkt 3, fixed assets on guarded premises may be counted once every 4 years.
  - Art. 27 ust. 2: differences are settled in that year's books.
  - Art. 4 ust. 5: the board cannot hand off responsibility for the physical count.
- **Grade:** 1 plain
- **Rhea's default:** the year-end checklist generates:
  - a cash count, if account 100 is used;
  - a bank confirmation from the statement balance;
  - balance confirmations sent to customers with open receivables;
  - verification of public-law and employee balances;
  - a stock and fixed-asset count where held.

  Differences enter as facts together with the board's decision.
- **Pack today:** missing. The `stocktake` type (date, comment, status, value) has no rules, and there is no balance-confirmation door.
- **Verdict:** missing
- **Deadline:** for most items, the window from 1 October 2026 to 15 January 2027. Cash and expensed stock are counted at 31 December 2026.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-24 — Do we revalue foreign-currency balances at year end?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 30 ust. 1: „Nierzadziej niż na dzień bilansowy wycenia się wyrażone w walutach obcych: 1) składniki aktywów … i pasywów – po obowiązującym na ten dzień średnim kursie ogłoszonym dla danej waluty przez Narodowy Bank Polski”. Ust. 4: the differences go to financial income or costs.
- **Grade:** 1 plain
- **Rhea's default:** on the year close (and monthly, if the policy chooses), a valuation rule revalues open foreign-currency receivables, payables and currency cash at the NBP rate valid on the balance date. The difference posts to 750/751, reversed on 1 January if the policy says so; for tax only realised differences count.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** the balance date (31 December 2026).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-25 — Must we write down doubtful receivables?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 35b ust. 1 pkt 3 requires a write-down on „należności kwestionowanych przez dłużników oraz z których zapłatą dłużnik zalega, a według oceny sytuacji majątkowej i finansowej dłużnika spłata należności w umownej kwocie nie jest prawdopodobna”. Ust. 2: the charge goes to other operating or financial costs. Art. 7 ust. 2b lets micro and small entities waive prudence for write-downs, but ust. 2c: „Przepisu ust. 2b nie stosuje się do: … 2) spółek kapitałowych” (KSR 16 pt 5.4 c). For tax, write-downs are non-deductible unless the CIT act lists them (CIT art. 16 ust. 1 pkt 27), with separate relief under art. 18f.
- **Grade:** 1 plain
- **Rhea's default:** at year end the receivables-ageing view proposes write-downs for receivables overdue beyond the policy threshold, or disputed. The board decides. The write-down posts to 761 against an allowance account and is marked non-deductible.
- **Pack today:** missing. The ageing view (`naleznosci-do-pilnowania`) exists, but invoices have no due date and there is no write-down rule.
- **Verdict:** missing
- **Deadline:** the balance date.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a ; https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf

### FB-26 — Must a small sp. z o.o. set up provisions and accrue known costs at year end?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 35d ust. 1: „Rezerwy tworzy się na: 1) pewne lub o dużym stopniu prawdopodobieństwa przyszłe zobowiązania, których kwotę można w sposób wiarygodny oszacować…”. Art. 39 ust. 2 requires accrued costs. Art. 7 ust. 2b would let small entities skip provisions, but ust. 2c excludes „spółek kapitałowych”. KSR 16 pt 12.4 and example 12.3: a small sp. z o.o. may not use art. 39 ust. 6 either, and may skip a provision only below its materiality threshold. KSR 6 governs provisions and accruals.
- **Grade:** 1 plain. KSR 16 confirms.
- **Rhea's default:** the year-end checklist asks about known obligations: unused leave, adviser fees relating to the year, court cases, bonuses. Those above the materiality threshold are posted (Wn costs / Ma provision or accrual) and marked deductible only when paid (CIT-06).
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** the balance date.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a ; https://www.gov.pl/attachment/bc48b7e4-7102-4458-bdd9-545997b2e38b

### FB-27 — Do we calculate deferred tax?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 37 ust. 1 sets the duty. Art. 37 ust. 10: „Jednostka mikro i jednostka mała mogą odstąpić od ustalania aktywów i rezerw z tytułu odroczonego podatku dochodowego.” No form-based exclusion applies, so capital companies may use it. KSR 16 pt 12.7: „Zastosowanie uproszczenia wymaga decyzji kierownika jednostki poprzedzonej wstępną oceną wpływu tego uproszczenia na obraz jednostki prezentowany w sprawozdaniu finansowym”. KSR 2 (2024) gives the method when deferred tax is calculated.
- **Grade:** 1 plain. KSR 16 says who decides.
- **Rhea's default:** a micro or small sp. z o.o. skips deferred tax. This is recorded as a board decision in the policy, with a one-line impact note. If the company grows to medium size, deferred tax becomes a rule over the CIT-06 markers.
- **Pack today:** missing. Only the decision record is needed; no rule.
- **Verdict:** missing
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a ; https://www.gov.pl/attachment/ab289f43-2138-4903-a88d-41db3342c656

### FB-28 — How do we close the year: the result, the income tax charge, and closing the books?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 12 ust. 2 pkt 1: books are closed on the last day of the year „niepóźniej niż w ciągu 3 miesięcy”. Ust. 4: „Ostateczne zamknięcie i otwarcie ksiąg rachunkowych jednostki kontynuującej działalność powinno nastąpić najpóźniej w ciągu 15 dni od dnia zatwierdzenia sprawozdania finansowego za rok obrotowy.” Ust. 5: closing is irreversible. Art. 37 ust. 8: the tax charge has a current and a deferred part. Art. 7 ust. 2: events that come to light before the close are taken into account.
- **Grade:** 1 plain
- **Rhea's default:** a "close the year" door, used after the year-end checklist (FB-23 to FB-27). Its rules compute the year's CIT (Wn 870 / Ma 220 for the amount above the advances), carry revenues and costs to 860, and lock the year. A second door after approval (FB-31) moves 860 per the shareholders' resolution and opens next year from the closing balances (art. 5 ust. 1).
- **Pack today:** missing. There is no year close. Account 860 exists; there is no income tax account (870) and no profit-distribution account (820). The periods pack locks months only.
- **Verdict:** missing
- **Deadline:** books closed within 3 months of the balance date; the final close within 15 days of approval.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-29 — Are we a micro or a small entity, and who decides that we file a simplified statement?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - UoR art. 3 ust. 1 pkt 1a: a micro entity has not exceeded, in the year and the year before (a new company: in its first year), two of „a) 2 000 000 zł – w przypadku sumy aktywów bilansu na koniec roku obrotowego, b) 4 000 000 zł – w przypadku przychodów netto ze sprzedaży towarów i produktów za rok obrotowy, c) 10 osób – w przypadku średniorocznego zatrudnienia”.
  - Pkt 1b: a small entity's limits are 33 000 000 zł, 66 000 000 zł and 50 people.
  - Both are worded by Dz.U. 2024 poz. 1863 and apply first to statements for years starting after 31 December 2023 (art. 14 ust. 3 of that act).
  - Art. 49c ust. 1: „Decyzję o sporządzeniu sprawozdania finansowego z zastosowaniem uproszczeń … – podejmuje organ zatwierdzający.” For a sp. z o.o. that is the shareholders' meeting.
  - Art. 45 ust. 3: the cash-flow statement and the statement of changes in equity are required only from entities other than micro and small.
  - Art. 49 ust. 5: a small entity may skip the management report if its notes give the information on own shares. Art. 48 ust. 3: a micro entity may replace the notes with the Annex 4 information.
  - KSR 16 pt 5.5–5.6: the decision is in writing and is taken before the statement is drawn up.
- **Grade:** 1 plain
- **Rhea's default:** Rhea computes the size test from facts (assets, revenue, headcount); a first-year company is tested on its first year. Rhea drafts the shareholders' resolution choosing Annex 4 or 5 and skipping the management report, to be adopted before the statement is prepared.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** before the statement is drawn up (KSR 16 pt 5.6).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a

### FB-30 — Who prepares and signs the annual financial statement, in what format, and by when?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - UoR art. 52 ust. 1: „Kierownik jednostki zapewnia sporządzenie rocznego sprawozdania finansowego niepóźniej niż w ciągu 3 miesięcy od dnia bilansowego”.
  - Art. 52 ust. 2: „Sprawozdanie finansowe podpisują – podając zarazem datę podpisu – osoba, której powierzono prowadzenie ksiąg rachunkowych, i kierownik jednostki, a jeżeli jednostką kieruje organ wieloosobowy – wszyscy członkowie tego organu albo co najmniej jedna osoba wchodząca w skład tego organu w sposób, o którym mowa w ust. 2b.” Under ust. 2b the other members file declarations.
  - Art. 45 ust. 1f: „Sprawozdanie finansowe sporządza się w postaci elektronicznej oraz opatruje się kwalifikowanym podpisem elektronicznym, podpisem zaufanym albo podpisem osobistym.” Ust. 1g: companies in the KRS use the Ministry's logical structure.
  - Art. 45 ust. 2: the statement consists of the balance sheet, the profit-and-loss account and the notes. Art. 48 ust. 1 pkt 1: the notes include the accounting policy.
  - The Ministry's new structures apply to statements for years starting on or after 1 January 2025 (podatki.gov.pl, 27 November 2025).
- **Grade:** 1 plain
- **Rhea's default:** a market writer renders the e-sprawozdanie XML (Annex 1, 4 or 5 structure) from the year-end trial balance, through an account-to-line mapping declared in the pack. The notes' policy section is rendered from FB-04. Signing happens outside Rhea, and the signed file comes back as a document under its hash.
- **Pack today:** missing. There is no statement writer and no account-to-line mapping; the JPK_V7M writer seam is the model to follow.
- **Verdict:** missing
- **Deadline:** 31 March 2027 for FY2026.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://podatki-arch.mf.gov.pl/komunikaty-techniczne/nowe-struktury-logiczne-dla-sprawozdan-finansowych-271125/ ; https://www.gov.pl/web/kas/struktury-e-sprawozdan

### CIT-07 — When do we file the CIT-8 and pay the year's tax?
- **Applies to:** sp. z o.o. full books
- **The law:** CIT art. 27 ust. 1: taxpayers file a return „o wysokości dochodu osiągniętego (straty poniesionej) w roku podatkowym – do końca trzeciego miesiąca roku następnego i w tym terminie wpłacić podatek należny albo różnicę między podatkiem należnym od dochodu wykazanego w zeznaniu a sumą zapłaconych zaliczek”. Ust. 1c: filed electronically. The return also reports the choice of quarterly or simplified advances (art. 25 ust. 1e, 7a) and a non-calendar year (art. 8 ust. 5).
- **Grade:** 1 plain
- **Rhea's default:** the CIT-8 figures are a fact computed from the closed year: book result, plus or minus the CIT-06 differences, less losses. Filing through the Ministry's e-services stays outside Rhea at first. Rhea tracks the deadline as a task and matches the payment.
- **Pack today:** missing. `tax_declaration` has no CIT-8 kind.
- **Verdict:** missing
- **Deadline:** 31 March 2027 for FY2026.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf

### FB-31 — Who approves the statement, and when can profit be distributed?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 53 ust. 1: „Roczne sprawozdanie finansowe jednostki … podlega zatwierdzeniu przez organ zatwierdzający, niepóźniej niż 6 miesięcy od dnia bilansowego.” Ust. 4: a company that is not audited may distribute or cover the result after approval. Art. 68 ust. 1: a sp. z o.o. must make the statement available to shareholders „najpóźniej na 15 dni przed zgromadzeniem wspólników”. For an Estonian-CIT company, the tax arises on the distribution resolution (KSR 2 pt 5.3.3).
- **Grade:** 1 plain
- **Rhea's default:** the approval is a fact entered through a door: the resolution approving the statement and distributing the profit or covering the loss. Its rule posts the distribution from 860 (reserve capital, dividends payable) and starts the 15-day filing task (FB-32).
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** approval by 30 June 2027 for FY2026; the documents to shareholders 15 days before the meeting.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://www.gov.pl/attachment/ab289f43-2138-4903-a88d-41db3342c656

### FB-32 — Where do we file the approved statement?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 69 ust. 1: „Kierownik jednostki składa we właściwym rejestrze sądowym: 1) roczne sprawozdanie finansowe, … 3) odpis uchwały bądź postanowienia organu zatwierdzającego o zatwierdzeniu rocznego sprawozdania finansowego i podziale zysku lub pokryciu straty, 4) sprawozdanie z działalności – w przypadku jednostek, o których mowa w art. 49 ust. 1 … – w terminie 15 dni od dnia zatwierdzenia rocznego sprawozdania finansowego.” Ust. 2: if the statement is not approved in time, it is filed within 15 days after the 6-month deadline. CIT art. 27 ust. 2: the separate filing with the head of the tax administration (Szef KAS) „nie dotyczy … podmiotów, które na podstawie odrębnych przepisów mają obowiązek przekazywania sprawozdań finansowych do Krajowego Rejestru Sądowego”.
- **Grade:** 1 plain
- **Rhea's default:** Rhea assembles the filing package (the signed XML, the resolution, and the management report or the art. 49 ust. 5 note) and a task with the deadline. Filing through the KRS portal is a human act, recorded as an event. No separate filing to the tax administration.
- **Pack today:** missing
- **Verdict:** missing
- **Deadline:** 15 days after approval (15 July 2027 if approved on 30 June 2027).
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf ; https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf

### CIT-08 — Must we send our books to the tax office as JPK_KR_PD, from which year, by when, and what must our accounts carry?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - CIT art. 9 ust. 1c: „Podatnicy prowadzący księgi rachunkowe są obowiązani prowadzić te księgi przy użyciu programów komputerowych oraz przesyłać właściwemu naczelnikowi urzędu skarbowego te księgi po zakończeniu roku podatkowego…”. Estonian-CIT companies send them too (the same provision refers to their declaration).
  - Phase-in, as the Ministry summarises art. 66 ust. 2 of Dz.U. 2021 poz. 2105: first for tax years starting after 31 December 2025 for taxpayers „obowiązanych do przesyłania plików JPK_V7M”; after 31 December 2026 for the others, „w tym podmiotów zobowiązanych do składania plików JPK_V7K”.
  - Deadline: under the regulation Dz.U. 2026 poz. 188 and a CIT amendment in force from 1 July 2026, the file is due „do końca siódmego miesiąca po zakończeniu roku podatkowego”.
  - Per account: marker S_12_1 is required (except for IFRS reporters); S_12_2 and S_12_3 (the CIT marker) are marked optional in the brochure.
  - The fixed-asset file JPK_ST_KR goes alongside; the Ministry's exemption from it covered only tax years starting in 2025.
- **Grade:** 1 plain. Caveat: the 7-month deadline comes from the Ministry's brochure update. The consolidated CIT text I read (Dz.U. 2026 poz. 554, law as of 18 March 2026) still ties the deadline to the return, and the Dz.U. citation of the amending statute is unverified.
- **Rhea's default:** Rhea's log already holds what JPK_KR_PD needs. Ship a writer in `markets/pl`, beside the JPK_V7M writer, producing:
  - the trial balance: opening balance, turnovers, closing balance and markers per account;
  - the journal: numbered entries with document references (FB-11);
  - the account entries (KontoZapis).

  The markers are declared on the pack's account events.
- **Pack today:** missing. Account events carry no S_12 markers, and postings lack journal numbers and document references.
- **Verdict:** missing
- **Deadline:** for a company filing JPK_V7M, FY2026 is due by 31 July 2027. A company filing JPK_V7K starts with FY2027.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf ; https://www.gov.pl/web/kas/struktury-jpk-w-podatkach-dochodowych ; https://www.podatki.gov.pl/media/umahd4he/aktualizacja-broszury-jpk_kr_pd-1-aktualna-od-01072026r.pdf ; https://www.podatki.gov.pl/media/if5hycxm/broszura_informacyjna-dotycz%C4%85ca-struktury-jpk_kr_pd-1-26082024.pdf

### FB-33 — How long do we keep the books and documents, and what about KSeF invoices?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 74 ust. 1: approved statements are kept at least 5 years, counted from the start of the year after the year of approval. Ust. 2 pkt 1: books 5 years; pkt 8: other documents 5 years. Ust. 3: periods run from the start of the year after the year the records concern. Art. 73a ust. 1, in force since 1 February 2026: „Przepisów art. 71 ust. 1 oraz art. 73 ust. 1, 2 i 4 nie stosuje się do dowodów księgowych przechowywanych w Krajowym Systemie e-Faktur”. Ust. 2: if KSeF keeps them for less time, the company keeps them for the rest of the period.
- **Grade:** 1 plain
- **Rhea's default:** nothing is ever deleted (invariant 1). KSeF invoices are kept as XML under their hash anyway, which covers art. 73a ust. 2.
- **Pack today:** right
- **Verdict:** right
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### FB-34 — Will our statement need an audit?
- **Applies to:** sp. z o.o. full books
- **The law:** UoR art. 64 ust. 1 pkt 4: the statements of „pozostałych jednostek, które w poprzedzającym roku obrotowym … spełniły co najmniej dwa z następujących warunków: a) średnioroczne zatrudnienie … co najmniej 50 osób, b) suma aktywów bilansu … co najmniej 3 125 000 euro, c) przychody netto … co najmniej 6 250 000 euro”. Lit. b and c apply first to years starting after 31 December 2024 (art. 14 ust. 2 of Dz.U. 2024 poz. 1863).
- **Grade:** 1 plain
- **Rhea's default:** at each year end, Rhea runs the prior-year test and raises a task to appoint an auditor when two of the three are met. A first-year company has no preceding year and is never in scope.
- **Pack today:** missing
- **Verdict:** later
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf

### CIT-09 — If we chose the Estonian lump sum, what changes in the books?
- **Applies to:** sp. z o.o. full books
- **The law:**
  - CIT art. 28d ust. 1: books and statements under the UoR, kept „w sposób zapewniający prawidłowe określenie: 1) wysokości zysku (straty) netto, podstawy opodatkowania i wysokości należnego podatku oraz 2) w kapitale własnym: a) kwoty zysków niepodzielonych i kwoty zysków podzielonych odniesione na kapitały, wypracowanych w latach opodatkowania ryczałtem, oraz b) kwoty niepokrytych strat”.
  - Art. 28j ust. 1 pkt 6: no IFRS. Art. 28m: the tax base is distributed profit, hidden profits and non-business expenses.
  - Art. 28t ust. 1 pkt 4: tax on hidden profits by the 20th of the following month. Art. 28r ust. 1: the declaration by the end of the third month.
  - KSR 2 pt 5.3.1: „Jednostka nie tworzy rezerw z tytułu odroczonego podatku dochodowego”. Pt 5.3.3: the tax liability and the tax cost „są ujmowane w momencie podjęcia decyzji o wypłacie dywidendy”.
- **Grade:** 2 settled. The books treatment is settled by KSR 2 (KSR resolution no. 9/2024) pt 5.3.
- **Rhea's default:** an Estonian-CIT company keeps the same books. The split of equity by the period profits were earned is a rule over 820/860. Hidden profits and non-business expenses are flagged per transaction each month, with a task by the 20th. No deferred tax and no CIT advances.
- **Pack today:** missing
- **Verdict:** later (once classic CIT ships)
- **Deadline:** the 20th of the following month for hidden profits; the declaration by the end of the third month of the next year.
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf ; https://www.gov.pl/attachment/ab289f43-2138-4903-a88d-41db3342c656

### CIT-10 — Does the minimum income tax apply to us?
- **Applies to:** sp. z o.o. full books
- **The law:** CIT art. 24ca ust. 1: a 10% minimum tax for companies with a loss, or with income of at most 2% of revenue. Ust. 14: „Przepisu ust. 1 nie stosuje się do podatników: 1) w roku podatkowym, w którym rozpoczęli działalność, oraz w kolejno następujących po sobie dwóch latach podatkowych”.
- **Grade:** 1 plain
- **Rhea's default:** not computed until the fourth tax year. The CIT-06 markers will feed it then.
- **Pack today:** missing
- **Verdict:** later
- **Deadline:** —
- **Sources:** https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf
