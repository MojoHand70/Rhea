# The auditor's curriculum — the register

**2026-10-19.** The syllabus of the biegły rewident examination (KRBR resolution
1695/34/2026 of 27 January 2026, in force on approval by PANA) read as a map of what the
Poland pack must know to be provable at an auditor's level (DECISIONS 2026-10-19, "The
auditor's exam"). Every numbered point of the ten subjects is one row. The Commission's
sample papers with their model solutions are the test suite (`rhea run pibr-exam`); this
register is the curriculum the suite samples from.

**How to read a row.**

- **Kind** — what in Rhea answers the point:
  *books* — rules, types, views, activities and formulas: the pack as data;
  *kernel* — the kernel by construction (the append-only journal, double entry that
  refuses an unbalanced entry, provenance on every fact, replay, period locks, the human
  gate); *engine* — code an engine or a writer must provide (a statement's XML, a
  depreciation schedule, consolidation, a derived cash-flow statement, iterative
  arithmetic); *host* — knowledge and judgement the host answers in words from the
  statute and the standards, no rule to write (economics, the law as such, the
  profession's own rules).
- **Pack today** — *covered* (the ids that do it), *partial* (what exists and what is
  missing), *missing*, or *—* where no pack content is expected (host rows).
- **Note** — what is missing or what the pack's reading is; "named by the host" marks a
  lesson the host itself named as missing when it abstained on the 2025 paper.

Prepared by AI from the syllabus text and the pack at v31, the rows refreshed for v32–v37
on 2026-10-20; not yet read by a practising auditor. Sector accounting (banks, SKOK, investment firms, insurers, pension funds,
public finance) is marked *host* throughout: the pack serves a small company on full
books or a sole trader (REVIEW.md), and those sectors are not its first-phase users; the
2025 papers test them heavily, which is where the bare model and the host alike fail.

## Summary

| Subject | Points | books | kernel | engine | host | covered | partial | missing | 2025 paper: bare model | 2025 paper: host |
|---|---|---|---|---|---|---|---|---|---|---|
| 1. Teoria i zasady rachunkowości | 41 | 13 | 14 | 7 | 7 | 12 | 18 | 8 | 105 of 120 (87%) passed | 94 of 120 (78%) passed |
| 2. Ekonomia i kontrola wewnętrzna | 23 | 0 | 2 | 0 | 21 | 1 | 2 | 0 | 89 of 98 (90%) passed | 73 of 98 (74%) passed |
| 3. Prawo | 65 | 2 | 0 | 2 | 61 | 0 | 7 | 5 | 96 of 108 (88%) passed | 78 of 108 (72%) not passed |
| 4. Prawo podatkowe I | 34 | 7 | 0 | 0 | 27 | 1 | 6 | 2 | 68 of 80 (85%) passed | 52 of 80 (65%) not passed |
| 5. Finanse i zarządzanie finansami | 53 | 4 | 0 | 3 | 46 | 1 | 3 | 4 | 82 of 88 (93%) not passed | 74 of 88 (84%) not passed |
| 6. Rachunkowość finansowa | 54 | 33 | 2 | 0 | 19 | 2 | 22 | 11 | 71 of 98 (72%) not passed | 64 of 98 (65%) not passed |
| 7. Prawo podatkowe II | 58 | 11 | 0 | 2 | 45 | 3 | 8 | 3 | 64 of 100 (64%) not passed | 43 of 100 (43%) not passed |
| 8. Rachunek kosztów i rachunkowość zarządcza | 23 | 5 | 0 | 5 | 13 | 0 | 5 | 6 | 66 of 72 (91%) passed | 49 of 72 (68%) passed |
| 9. Sprawozdania finansowe i ich analiza | 43 | 8 | 0 | 9 | 26 | 0 | 9 | 8 | 38 of 50 (76%) not passed | 38 of 50 (76%) not passed |
| 10. Rewizja finansowa oraz etyka | 27 | 0 | 2 | 0 | 25 | 2 | 2 | 0 | 29 of 44 (65%) passed | 25 of 44 (56%) not passed |
| **Total** | **421** | **83** | **20** | **28** | **290** | **22** | **82** | **47** | 708 of 858 (82%), 6 papers of 10 | 590 of 858 (68%), 3 of 10; 0 of 429 from a rule |

**What the numbers say.** Three subjects are the books' own — theory, financial
accounting, cost accounting — and there the pack covers the bookkeeper's half (documents,
VAT, bank, settlements, CIT advances, prepaid costs) and lacks the auditor's half (fixed
assets and depreciation, inventories and cost of manufacture, equity and the distribution
of result, provisions, deferred tax, impairment, leasing, financial instruments, year-end
closing, the statements themselves). The statements subject needs one writer (the
Ministry's e-Sprawozdania XML) and the derived statements (cash flow, changes in equity,
notes) before any of its points can be covered. Tax II is the pack's nearest gap with the
most rules to write (tax depreciation, dividends, the payer's duties, CIT-8). Economics,
law, finance of banks and insurers, and the audit profession's own rules are the host's:
answered from the statute in words, never a rule — and the place where the statute by
article in front of the host (DECISIONS 2026-10-19) will turn "memory" into "statute".

**2026-10-20, after pack v32–v37.** The auditor's half has begun: fixed assets accepted,
depreciated and disposed (v32), the year's close, the statement's approval and the
result's distribution (v33), provisions, accruals and receivable write-downs (v34),
goods and materials out and inventories written down (v36), construction costs and
advances for assets (v37) — five rows of financial accounting moved from missing to
partial and eleven partial rows grew. Still missing in the books' rows: deferred tax,
impairment of fixed assets, the lessee's postings, financial instruments, loans, equity
contributions, the cost of manufacture and the cost method from receipts.

## 1. Teoria i zasady rachunkowości

| Point | The point | Kind | Pack today | Note |
|---|---|---|---|---|
| 1.1.1 | Źródła regulacji prawa bilansowego | host | — | the pack's warrants cite UoR, KSR, KSH by article; the host says which governs |
| 1.1.2 | Rachunkowość jako system informacyjny jednostki | kernel | covered | the event log and its projections are the information system; every fact explained (invariant 5) |
| 1.1.3 | Pojęcia, funkcje, części składowe rachunkowości | host | — | |
| 1.1.4 | Założenia koncepcyjne, cechy jakościowe, zasady nadrzędne, metody wyceny | host | partial | accrual and matching are how the pack posts (sale date, RMK); going concern and the valuation methods are rows 6.1.1 |
| 1.1.5 | Harmonizacja: MSR, dyrektywy UE | host | — | |
| 1.2.1.1 | Równanie bilansowe | kernel | covered | an entry that does not balance is refused (exam task I-2 proved it live) |
| 1.2.1.2 | Kryteria kwalifikacji aktywów i zobowiązań | books | partial | `account.type` (asset, liability, equity, revenue, expense); the criteria themselves are the host's |
| 1.2.1.3 | Wycena początkowa i bilansowa składników bilansu | books | partial | invoices, bank, FX at the NBP rate, prepaid costs; fixed assets at cost less depreciation (v32), inventories at the value the document states or the lower net selling price (v36), receivables less write-downs (v34), provisions at their estimate (v34); missing: financial instruments, the cost method from receipts — see 6.1.1 |
| 1.2.1.4 | Struktura aktywów bilansu | engine | missing | the balance sheet layout of UoR annex 1 as a view over accounts; the 2025 tasks name lines like B.III.1.b |
| 1.2.1.5 | Struktura pasywów bilansu | engine | missing | as above |
| 1.2.1.6 | Sporządzanie bilansu, w tym uproszczonego (mała, mikro) | engine | missing | the company's size is declared (`declare_company_form`); the statement writer is not |
| 1.2.2.1 | Ujęcie przychodów i kosztów w RZiS | books | partial | sales on the sale date, purchases, prepaid costs released by month; missing: cost of goods sold, changes in inventories, accruals of liabilities |
| 1.2.2.2 | Struktura rachunku zysków i strat | engine | partial | `wynik`, `wynik-miesiace` show the result by months; the statutory layout (comparative and calculation variants) is missing |
| 1.2.2.3 | Sporządzanie RZiS, w tym uproszczonego | engine | missing | |
| 1.2.3.1 | Rachunek memoriałowy a kasowy | kernel | covered | documents post on their own dates, payments are their own events, matched later |
| 1.2.3.2 | Kategorie przepływów pieniężnych | engine | missing | the cash-flow statement, derived (indirect method) — see 9.2.5 |
| 1.3.1 | Rodzaje zdarzeń i ich wpływ na bilans | kernel | covered | an event matches a rule that posts; the effect on the balance is the posting |
| 1.3.2 | Dokumentowanie zdarzeń; treść dowodów; korekta dowodów | books | partial | `received_document`, FA(3) fields, KSeF; corrections are amendments, never edits; the formal content of a voucher (UoR art. 21) is not checked as such |
| 1.3.3 | Kontrola i kwalifikowanie dowodów do ksiąg | kernel | covered | nothing posts without a rule; what no rule explains waits in the worklist; a rejected invoice becomes a task (`pl-ksef-rejection-task`) |
| 1.3.4 | Zapisy księgowe: treść i struktura | kernel | covered | `posting`: entry, line, book, account, side, amount, currency, date; provenance on each |
| 1.3.5 | Konto: cechy, rodzaje, konta bilansowe, wynikowe, pozabilansowe | books | partial | balance and result accounts by `account.type`; off-balance accounts missing (the banks' task II-1 needs them) |
| 1.3.6 | Ewidencja typowych operacji | books | partial | sales, purchases, bank, VAT, CIT advances, RMK, PZ; fixed assets in, depreciated and out (v32), the year's close and the result's distribution (v33), provisions and write-downs (v34), goods and materials out at the invoice's value (v36), construction costs and advances for assets (v37); missing: payroll, loans, contributions of capital |
| 1.4.1 | Zakres podmiotowy obowiązku prowadzenia ksiąg | books | partial | the form of books is declared (`pl-declare-company-form`); the thresholds of UoR art. 2 are the host's |
| 1.4.2 a | Dziennik | kernel | covered | the event log: append-only, chronological, numbered (invariant 1) |
| 1.4.2 b | Konta księgi głównej, podwójny zapis, otwieranie i zamykanie | kernel | partial | double entry by construction; opening balances and the year-end closing of accounts are missing (todo "import opening balances") |
| 1.4.2 c | Konta ksiąg pomocniczych | books | partial | settlements per contractor and document (open amounts), `fixed_asset`, stock by PZ; no formal subsidiary-ledger notion |
| 1.4.2 d | Zestawienie obrotów i sald | engine | missing | a trial balance over postings — one analysis view |
| 1.4.2 e | Uzgadnianie ksiąg | kernel | covered | replay reproduces the state (invariant 4); matching reconciles bank and documents |
| 1.4.2 f | Poprawianie błędów księgowych | kernel | covered | a correction is a new event that supersedes; nothing is overwritten (UoR art. 25 by construction) |
| 1.4.3 a | Realizacja zasad podwójnego zapisu w systemie IT | kernel | covered | |
| 1.4.3 b | Zgodność systemu z przepisami | host | partial | UoR art. 10 ust. 1 pkt 3 asks for a description of the system; the register and SPEC are that description's material, not the document |
| 1.4.3 c | Dostosowanie funkcji informacyjnych | books | covered | views and briefings are data (ViewDef) |
| 1.4.4 | Księgi rzetelne, bezbłędne, sprawdzalne, bieżące | kernel | covered | verifiable by provenance and replay; error-free by determinism; current by the clock channel and queues; reliable by rules a person approved |
| 1.4.5 | Miejsce prowadzenia i przechowywanie ksiąg | host | partial | backup and restore exist; retention periods (UoR art. 74) are not enforced |
| 1.4.6 | Odpowiedzialność za prowadzenie ksiąg | host | partial | the gate records who approved what; the manager's responsibility (art. 4 ust. 5) is the host's to explain |
| 1.5.1 | Inwentaryzacja: zasady, metody, terminy | books | partial | `stocktake` exists; methods (spis z natury, potwierdzenie sald, weryfikacja) and the statutory dates are missing |
| 1.5.2 | Różnice inwentaryzacyjne i ich ujęcie | books | missing | shortages, surpluses, compensation, their postings |
| 1.6.1 | Dokumentacja zasad (polityki) rachunkowości | books | partial | the choices are approved rules (cost method, tax treatment, form); the policy document itself is not produced |
| 1.6.2 | Obieg i kontrola dowodów | kernel | partial | doors, tasks and queues are the circulation; a written description is missing |
| 1.6.3 | Dokumentacja inwentaryzacji | books | missing | |
| 1.6.4 | Dokumentacja kontroli wewnętrznej | kernel | partial | every approval and every fact is on the record; the description of the control system is missing |

## 2. Ekonomia i kontrola wewnętrzna

| Point | The point | Kind | Pack today | Note |
|---|---|---|---|---|
| 2.1.1 | Narzędzia analizy ekonomicznej | host | — | indicators and averages over the books are analysis views (DuckDB); the theory is the host's |
| 2.1.2 | Pojęcia i kategorie ekonomiczne | host | — | |
| 2.1.3 | Rynek: popyt, podaż, równowaga, elastyczność | host | — | |
| 2.1.4 | Teoria wyboru konsumenta | host | — | |
| 2.1.5 | Mikroekonomiczna analiza przedsiębiorstwa | host | — | |
| 2.1.6 | Struktury rynku | host | — | |
| 2.1.7 | Efektywność i zawodność rynku | host | — | |
| 2.1.8 | Dochód narodowy | host | — | |
| 2.1.9 | Polityka monetarna | host | — | |
| 2.1.10 | Polityka budżetowa | host | — | |
| 2.1.11 | Bilans płatniczy, handel zagraniczny, kurs walutowy | host | — | the NBP channel publishes the rates the books use; the theory is the host's |
| 2.2.1 | Rynek finansowy i jego instrumenty (weksel, czek, obligacje, akcje, giełda, fundusze) | host | — | a bill of exchange or a bond as a document the books carry is row 6.1.8 f |
| 2.2.2 | Instytucje finansowe | host | — | |
| 2.2.3 | Nadzór nad rynkiem finansowym | host | — | |
| 2.3.1 | Przedsiębiorstwo i jego otoczenie | host | — | |
| 2.3.2 | Zarządzanie przedsiębiorstwem | host | — | |
| 2.3.3 | Kierowanie ludźmi | host | — | |
| 2.3.4 | Podejmowanie decyzji | host | — | |
| 2.4.1 | Ryzyko w działalności gospodarczej | host | — | |
| 2.4.2 | System kontroli wewnętrznej i audyt wewnętrzny | kernel | partial | the gate, provenance and replay are a control system; segregation of duties per state arrives with smart documents (DIRECTION 2026-10-19) |
| 2.4.3 | Kontrola wewnętrzna a jakość ksiąg i sprawozdań | kernel | covered | every fact explained, every change a new version, locks on closed months; the control is the architecture (invariants 1–7) |
| 2.4.4 | Ład organizacyjny: prawa właścicieli i nadzorujących | host | partial | persons, deputies and the owner are data (people pack); governance itself is the host's |
| 2.4.5 | Rola interesariuszy | host | — | |

## 3. Prawo

| Point | The point | Kind | Pack today | Note |
|---|---|---|---|---|
| 3.1.1 | Źródła prawa | host | — | |
| 3.1.2 | Systematyka prawa | host | — | |
| 3.1.3 | Wykładnia prawa | host | — | |
| 3.1.4 | Obowiązywanie prawa, reguły kolizyjne | host | — | a rule's `effective_from` is the pack's way of applying law from its date |
| 3.1.5 | Norma a przepis | host | — | |
| 3.1.6 | Prawo wspólnotowe | host | — | |
| 3.2.1 | Podmioty, przedstawicielstwo, czynności prawne, terminy, przedawnienie | host | partial | limitation of claims is a date the receivables' ageing could watch (`naleznosci-wiek`); not a rule yet |
| 3.2.2 | Własność i prawa rzeczowe | host | — | |
| 3.2.3 a | Powstawanie i wykonywanie zobowiązań | host | — | |
| 3.2.3 b | Wielość dłużników i wierzycieli | host | — | |
| 3.2.3 c | Umowy: treść i forma | host | — | |
| 3.2.3 d | Bezpodstawne wzbogacenie | host | — | |
| 3.2.3 e | Czyny niedozwolone | host | — | |
| 3.2.3 f | Wykonanie i skutki niewykonania; potrącenie, odnowienie, zmiana stron | host | partial | set-off is matching with a reversal; the 15 000 zł cashless rule is REVIEW MON-23 |
| 3.2.3 g | Ochrona wierzyciela przy niewypłacalności | host | — | |
| 3.2.3 h | Umowy nazwane (sprzedaż, dostawa, dzieło, najem, leasing, pożyczka, rachunek bankowy, zlecenie, komis, ubezpieczenie, spółka, poręczenie, darowizna…) | host | partial | leases (`lease_contract`), sales and bank accounts are objects; loans, commission sales, guarantees, donations are not |
| 3.2.4 | Naruszenia prawa | host | — | |
| 3.3.1 | Zasady prawa pracy | host | — | |
| 3.3.2 | Stosunek pracy | host | — | |
| 3.3.3 | Wynagrodzenie i inne świadczenia | engine | missing | payroll: the list of pay, PIT advances, ZUS, their postings — an engine plus rules, not yet begun |
| 3.3.4 | Obowiązki pracodawcy i pracownika | host | — | |
| 3.3.5 | Odpowiedzialność materialna pracowników | host | — | |
| 3.3.6 | Czas pracy | host | — | |
| 3.3.7 | Urlopy | host | — | |
| 3.3.8 | Uprawnienia rodzicielskie | host | — | |
| 3.3.9 | Zatrudnianie młodocianych | host | — | |
| 3.3.10 | BHP | host | — | |
| 3.3.11 | Spory ze stosunku pracy | host | — | |
| 3.3.12 | Wykroczenia przeciwko prawom pracownika | host | — | |
| 3.3.13 | Przedawnienie w prawie pracy | host | — | |
| 3.3.14 | Podleganie ubezpieczeniom społecznym | host | — | |
| 3.3.15 | Ustalanie składek | engine | missing | with payroll (3.3.3); the owner's own ZUS is REVIEW "The owner and money" |
| 3.3.16 | Zgłoszenia, konta, rozliczanie składek i zasiłków | books | partial | the ZUS account is declared and payments to it are recognised (`pl-post-explained-bank-operation`, 229); declarations (DRA) are missing |
| 3.3.17 | Obowiązki ubezpieczonych | host | — | |
| 3.3.18 | Zwrot nienależnych świadczeń, odsetki | host | — | |
| 3.3.19 | Kontrola ZUS | host | — | |
| 3.3.20 | Naruszenia prawa | host | — | |
| 3.4.1 | Spółki osobowe i kapitałowe | host | partial | the company's legal form is declared (`company.legal_form`) |
| 3.4.2 | Powstawanie spółek | books | missing | share capital, contributions, founding costs as postings — row 6.6.1 a |
| 3.4.3 | Organy; prawa wspólników i akcjonariuszy | host | missing | resolutions on the distribution of result and dividends as events — row 6.1.7 c |
| 3.4.4 | Rozwiązanie i likwidacja | host | — | |
| 3.4.5 | Łączenie, podział, przekształcanie | host | missing | row 6.6.1 b |
| 3.4.6 | Naruszenia prawa | host | — | |
| 3.5.1 | Postępowanie upadłościowe i restrukturyzacyjne: przepisy ogólne | host | — | |
| 3.5.2 | Ogłoszenie upadłości | host | — | |
| 3.5.3 | Skutki ogłoszenia upadłości | host | — | |
| 3.5.4 | Postępowanie po ogłoszeniu upadłości | host | — | |
| 3.5.5 | Zgłoszenie i ustalenie wierzytelności | host | partial | a creditor's claim and its write-off when the proceedings end (exam task I-3 posted it: Dt 761 / Ct 200) |
| 3.5.6 | Układ w upadłości | host | — | |
| 3.5.7 | Likwidacja masy | host | — | |
| 3.5.8 | Podział funduszów masy | host | — | |
| 3.5.9 | Zakończenie i umorzenie postępowania | host | — | |
| 3.5.10 | Międzynarodowe postępowanie upadłościowe | host | — | |
| 3.5.11 | Odrębne postępowania upadłościowe | host | — | |
| 3.5.12 | Postępowania restrukturyzacyjne: przepisy ogólne | host | — | |
| 3.5.13 | Postępowania restrukturyzacyjne: przepisy szczególne | host | — | |
| 3.5.14 | Międzynarodowe postępowanie restrukturyzacyjne | host | — | |
| 3.5.15 | Odrębne postępowania restrukturyzacyjne | host | — | |
| 3.5.16 | Naruszenia prawa | host | — | |
| 3.6.1 | Warunki działalności: równość, rozliczenia bezgotówkowe, koncesje | host | partial | the cashless-payment limit and its tax effect are REVIEW MON-23 (open) |
| 3.6.2 | Przedsiębiorstwa osób fizycznych, państwowe, spółdzielnie, banki, ubezpieczyciele: powstawanie i likwidacja | host | — | |
| 3.6.3 | Ochrona środowiska w działalności | host | — | |
| 3.6.4 | Organizacje społeczne, fundacje, stowarzyszenia, przedstawicielstwa | host | — | |
| 3.6.5 | Prawo dewizowe | host | — | |
| 3.6.6 | Naruszenia prawa | host | — | |

## 4. Prawo podatkowe — część I

| Point | The point | Kind | Pack today | Note |
|---|---|---|---|---|
| 4.1.1 | Istota i cechy podatku | host | — | |
| 4.1.2 | Źródła prawa podatkowego | host | — | |
| 4.1.3 | Organy podatkowe, KAS, właściwość | host | partial | the company's tax office is declared (`pl-identify-company`) |
| 4.1.4 | Interpretacje urzędowe | host | — | a warrant may cite one; the law watch (todo) would follow them |
| 4.1.5 | Współdziałanie, porozumienie inwestycyjne | host | — | |
| 4.2.1 | Obowiązek a zobowiązanie podatkowe | host | partial | the VAT obligation date is a rule (`pl-book-sales-invoice`, art. 19a) |
| 4.2.2 | Powstawanie i wygasanie zobowiązań | books | partial | tax liabilities on 225 and their payment from the micro-account; expiry by payment only |
| 4.2.3 | Zaległość, odsetki za zwłokę, opłata prolongacyjna | books | missing | interest on arrears is a formula with a published rate; a late payment should compute it |
| 4.2.4 | Nadpłata | books | partial | VAT carried forward (`vat_settlement`); overpayment claims are not |
| 4.2.5 | Ulgi w spłacie | host | — | |
| 4.2.6 | Odpowiedzialność za zobowiązania | host | — | |
| 4.2.7 | Zabezpieczenie wykonania | host | — | |
| 4.2.8 | Deklaracje, informacje, rachunki | books | partial | `tax_declaration`, `deklaracja-vat`, JPK_V7M written and validated; CIT-8, PIT-36 missing |
| 4.2.9 | Przedawnienie | host | — | |
| 4.2.10 | Przeciwdziałanie unikaniu opodatkowania | host | — | |
| 4.2.11 | STIR, wyłudzenia skarbowe | host | — | the white list of VAT payers (payments to unlisted accounts) is a check the pack could do; missing |
| 4.3.1 | Postępowanie podatkowe | host | — | the 2025 paper's task 1 is this: a corrected return during proceedings, the appeal and its deadline |
| 4.3.2 | Szacowanie podstawy | host | — | |
| 4.3.3 | Czynności sprawdzające, kontrola podatkowa | host | — | |
| 4.3.4 | Kontrola celno-skarbowa | host | — | |
| 4.3.5 | Tajemnica skarbowa | host | — | |
| 4.3.6 | Ewidencja i identyfikacja podatników (NIP) | books | covered | companies by NIP, `identify_company` |
| 4.4.1 | VAT: zakres, podatnicy, rejestracja, obowiązek, zwolnienia, miejsce świadczenia, podstawa, stawki, import, sprzedaż, podatek naliczony, odpowiedzialność nabywcy, obliczenie i zwrot, deklaracje, procedury szczególne | books | partial | registration status and exemptions (art. 113, art. 43), rates, obligation date, registers, JPK_V7M, KSeF in and out, monthly settlement, carry-forward; missing: place of supply and intra-EU trade, import VAT, reverse charge, split payment, refunds, special procedures (REVIEW "VAT — sales and purchases") |
| 4.4.2 | Podatek akcyzowy | host | — | outside the pack's users |
| 4.4.3 | Opłata skarbowa, PCC | books | missing | PCC on loans and capital increases is a small, frequent rule for a sp. z o.o. |
| 4.4.4 | Podatek od gier | host | — | outside |
| 4.5.1 | Prawo celne: prawa i obowiązki, AEO, WIT | host | — | |
| 4.5.2 | Taryfa, pochodzenie, wartość celna | host | — | |
| 4.5.3 | Dług celny | host | — | import VAT and duty on a customs document are a reader and a rule; missing (4.4.1) |
| 4.5.4 | Zgłoszenie celne | host | — | |
| 4.5.5 | Procedury specjalne | host | — | |
| 4.6.1 | Przestępstwa i wykroczenia skarbowe | host | — | |
| 4.6.2 | Kary | host | — | |
| 4.6.3 | Zaniechanie ukarania (czynny żal) | host | — | |

## 5. Finanse i zarządzanie finansami

| Point | The point | Kind | Pack today | Note |
|---|---|---|---|---|
| 5.1.1 | Kapitały własne | books | missing | equity accounts and their movements — row 6.1.2 c |
| 5.1.2 | Rozliczenia gotówkowe i bezgotówkowe, w walutach obcych | books | covered | bank operations, matching, FX at the NBP rate and realised differences; cash (kasa) is missing |
| 5.1.3 | Zasilanie finansowe, analiza i prognoza przepływów | engine | missing | the cash forecast (todo) |
| 5.1.4 | Wartość pieniądza w czasie: FV, PV, annuity, perpetuity | books | partial | the formula language has exact arithmetic and declared rounding; whether it has powers and roots for discounting is to be checked |
| 5.1.5 | NPV, IRR, ocena projektu | engine | missing | NPV is a formula; IRR needs iteration — an engine |
| 5.1.6 | Decyzje długoterminowe: źródła finansowania, koszt kapitału, dźwignia, dywidendy | host | partial | leases are objects; loans, bonds, factoring are not |
| 5.1.7 | Decyzje krótkoterminowe: płynność, zapasy, należności, zobowiązania, kapitał obrotowy | books | partial | open amounts, receivables ageing, due dates watched; stock and payables ageing missing |
| 5.1.8 | Planowanie finansowe | engine | missing | |
| 5.1.9 | Zarządzanie ryzykiem finansowym | host | — | |
| 5.1.10 | Restrukturyzacja i upadłość: aspekty finansowe | host | — | |
| 5.2.1 | Wycena metodami majątkowymi | host | — | net assets are a view over the books once equity exists |
| 5.2.2 | Wycena metodami dochodowymi (DCF, EVA) | host | — | |
| 5.2.3 | Wycena metodami mieszanymi | host | — | |
| 5.2.4 | Wycena innymi metodami | host | — | |
| 5.3.1 | Środki publiczne, nadwyżka, deficyt | host | — | |
| 5.3.2 | Jednostki sektora finansów publicznych | host | — | |
| 5.3.3 | Gospodarka finansowa jednostek sektora | host | — | |
| 5.3.4 | Jednostki budżetowe, zakłady, fundusze celowe, agencje | host | — | |
| 5.3.5 | Budżet państwa i JST | host | — | |
| 5.3.6 | Środki z budżetu UE | host | — | |
| 5.3.7 | Partnerstwo publiczno-prywatne | host | — | |
| 5.4.1 | System bankowy | host | — | banks are not the pack's users; rows 5.4.1–5.4.22 are the host's |
| 5.4.2 | Bank Gospodarstwa Krajowego | host | — | |
| 5.4.3 | Grupy i konglomeraty finansowe | host | — | |
| 5.4.4 | SKOK | host | — | |
| 5.4.5 | Inne instytucje finansowe | host | — | |
| 5.4.6 | Fundusze własne banków | host | — | |
| 5.4.7 | Kapitał wewnętrzny | host | — | |
| 5.4.8 | Wymogi kapitałowe | host | — | |
| 5.4.9 | Bufory kapitałowe | host | — | |
| 5.4.10 | Domiar kapitałowy, współczynnik wypłacalności | host | — | |
| 5.4.11 | Zarządzanie ryzykiem i kontrola wewnętrzna banku | host | — | |
| 5.4.12 | Kategorie ryzyka bankowego, normy ostrożnościowe | host | — | named by the host on the 2025 paper (credit exposures) |
| 5.4.13 | Akceptowalny poziom ryzyka | host | — | |
| 5.4.14 | Identyfikacja, pomiar, raportowanie ryzyka | host | — | |
| 5.4.15 | Zarządzanie aktywami i pasywami | host | — | |
| 5.4.16 | Limity wewnętrzne, normy koncentracji | host | — | |
| 5.4.17 | Operacje i instrumenty bankowe | host | — | |
| 5.4.18 | System gwarancji depozytów | host | — | |
| 5.4.19 | System rekompensat | host | — | |
| 5.4.20 | Finansowanie nadzoru | host | — | |
| 5.4.21 | Restrukturyzacja i uporządkowana likwidacja | host | — | |
| 5.4.22 | Role instytucji w resolution | host | — | |
| 5.5.1 | System ubezpieczeń: formy, rodzaje | host | — | insurers are not the pack's users; rows 5.5.1–5.5.10 are the host's |
| 5.5.2 | Operacje ubezpieczeniowe | host | — | |
| 5.5.3 | Kapitały i fundusze zakładu ubezpieczeń | host | — | |
| 5.5.4 | Lokowanie środków | host | — | |
| 5.5.5 | Ryzyko, ORSA | host | — | |
| 5.5.6 | Reasekuracja, funkcje kluczowe | host | — | |
| 5.5.7 | Wypłacalność, rezerwy techniczne, środki własne | host | — | |
| 5.5.8 | Relacje środków własnych do wymogów | host | — | |
| 5.5.9 | UFG | host | — | |
| 5.5.10 | Inne instytucje rynku ubezpieczeniowego | host | — | |

## 6. Rachunkowość finansowa

| Point | The point | Kind | Pack today | Note |
|---|---|---|---|---|
| 6.1.1 a | Wycena przy założeniu kontynuacji działania | books | partial | documents at their amounts, FX at the NBP rate, prepaid costs; fixed assets at cost less depreciation (v32, art. 28 ust. 1 pkt 1), inventories at cost or the lower net selling price (v36, art. 28 ust. 1 pkt 6), receivables less write-downs (v34, art. 35b), provisions at a justified estimate (v34, art. 35d); missing: financial instruments (amortised cost, market price), the cost method from receipts, equity and loans at nominal value as rules |
| 6.1.1 b | Wycena w warunkach utraty kontynuacji działania | books | missing | art. 29: net realisable value, the reserve for the consequences; the 2025 paper's first question |
| 6.1.1 c | Wycena składników w walutach obcych | books | partial | initial measurement at the NBP rate of the day before (art. 30 ust. 2) and realised differences are covered; the balance-date revaluation of receivables, payables and cash (art. 30 ust. 1) is missing — exam task I-2 drafted it live |
| 6.1.2 a | Aktywa trwałe: WNiP, środki trwałe, należności i inwestycje długoterminowe, RMK długoterminowe | books | partial | `fixed_asset` accepted from 080, depreciated monthly to 071, disposed (v32); a cost that is not building leaves 080 for 761 and an advance for an asset in a foreign currency keeps its payment-day złoty to settlement, no exchange difference (v37, KSR 11 pkt 6.18); `vehicle`, `lease_contract` exist; missing: improvements, long-term investments, the lessee's postings |
| 6.1.2 b | Aktywa obrotowe: zapasy, należności, inwestycje krótkoterminowe, RMK | books | partial | goods in by PZ at purchase price; goods and materials out at the value the sales invoice states (v36: 731/330, 761/310); stock written down to net selling price against 349 and the write-down reversed (v36); receivables with open amounts and their write-downs (v34); prepaid costs; missing: the cost method from receipts (FIFO, average), products and work in progress, short-term investments (exam task I-1 drafted the write-down live) |
| 6.1.2 c | Kapitały własne: podstawowy, zapasowy, z aktualizacji, rezerwowe, wynik, odpisy z zysku | books | partial | the result approved from 860 to 820 and distributed to reserve capital, dividends or a loss covered from reserve (v33: 802, 820, 241); missing: share capital and called-up contributions (dopłaty), the revaluation reserve — the 2025 paper asks about both |
| 6.1.2 d | Zobowiązania i rezerwy: rezerwy, zobowiązania długo- i krótkoterminowe, RMK bierne | books | partial | payables with open amounts; provisions for liabilities (830) and accruals of liabilities (641) created, used and released (v34, art. 35d, art. 39 ust. 2 pkt 2); missing: loans received and granted, the long- and short-term split at presentation — exam task I-3 drafted a warranty provision live |
| 6.1.3 a | Różnice kursowe | books | partial | realised differences on settlement (`pl-post-fx-difference`, 750/751); unrealised at the balance date missing |
| 6.1.3 b | Przeszacowanie wartości | books | missing | revaluation of fixed assets by regulation; reclassification to investment property (named by the host) |
| 6.1.3 c | Wycena w wartości godziwej | books | missing | investment property, financial instruments |
| 6.1.3 d | Utrata wartości aktywów (KSR 4) | books | partial | write-downs of receivables (v34, art. 35b) and of inventories to net selling price, reversed when the cause ends (v36, art. 28 ust. 1 pkt 6, art. 34 ust. 5, art. 35c); missing: impairment of fixed assets and goodwill (KSR 4) |
| 6.1.3 e | Ostrożna wycena i współmierność | books | partial | prepaid costs (RMK czynne, KSR) and the sale date carry matching; prudence: provisions (v34), write-downs of receivables (v34) and inventories (v36); missing: impairment of fixed assets |
| 6.1.4 a | Przychody ze sprzedaży produktów i usług | books | covered | `pl-book-sales-invoice`, `pl-post-sales-invoice` (sale date, art. 19a; 700/201/222) |
| 6.1.4 b | Przychody ze sprzedaży towarów | books | covered | as above; the cost of goods sold against them at the value the invoice states (v36: 731/330); the automatic cost method from receipts (FIFO, average) is the next inventories lesson (6.1.2 b) |
| 6.1.4 c | Sprzedaż hurtowa, detaliczna, wysyłkowa, elektroniczna, komisowa, tranzytowa | books | partial | invoiced sales only; retail by receipts (kasa fiskalna), commission and transit sales missing |
| 6.1.4 d | Reklamacje | books | missing | a complaint and its credit note; the operations corpus has a `complaint` case, the pack does not |
| 6.1.5 a | Pozostałe przychody operacyjne: dotacje, darowizny, sprzedaż środków trwałych, rozwiązane rezerwy | books | partial | 760 receives overpayment differences, released provisions (v34), reversed inventory write-downs (v36), the sale of materials and of fixed assets (v36, art. 3 ust. 1 pkt 32); missing: grants, donations (the 2025 paper's donation question was the theory paper's unkeyed one) |
| 6.1.5 b | Pozostałe koszty operacyjne: aktualizacja aktywów niefinansowych, odszkodowania, kary | books | partial | 761 receives underpayment differences, the net value of a disposed asset (v32), write-downs of receivables (v34) and inventories (v36), materials sold (v36), a casualty's cost or idle capacity moved off a construction (v37, art. 3 ust. 1 pkt 32 lit. i); missing: penalties and damages — exam task I-3 posted a write-off and a destroyed-inventory loss live |
| 6.1.6 a | Przychody finansowe: dywidendy, odsetki, aktualizacja aktywów finansowych, dodatnie różnice kursowe | books | partial | FX gains (750) and bank interest (NINT) covered; dividends received, interest on loans given, reversals of write-downs missing |
| 6.1.6 b | Koszty finansowe: odsetki, aktualizacja, ujemne różnice kursowe | books | partial | FX losses (751), bank fees; interest on loans and leases, write-downs of financial assets missing (exam task I-1: Dt 751) |
| 6.1.7 a | Wynik w wariancie porównawczym | books | partial | costs by nature (4xx) are how the pack posts; `close_year` (v33, art. 12 ust. 2 pkt 1) gives every cost, revenue and tax account's balance to 860 on the balance sheet date and `approve_statement` moves it to 820; missing: the change in products (490) |
| 6.1.7 b | Wynik w wariancie kalkulacyjnym | books | partial | the cost of goods sold on 731 (v36); costs by function (5xx) missing |
| 6.1.7 c | Podział wyniku | books | partial | `distribute_result` (v33, art. 53 ust. 4; KSH art. 191–192): to reserve capital, to dividends, a loss covered from reserve, the rest undistributed on 820; `pay_dividend` with the 19% withheld (PIT art. 30a); missing: the dividend cap of KSH art. 191 § 4 as a refusal, interim dividends |
| 6.1.8 a | Umowy długoterminowe (KSR 3) | books | missing | |
| 6.1.8 b | Leasing (KSR 5) | books | partial | `lease_contract` holds the contract; the lessee's postings (finance lease: asset and liability, interest, depreciation) are missing |
| 6.1.8 c | Utrata wartości aktywów | books | partial | as 6.1.3 d |
| 6.1.8 d | Odroczony podatek dochodowy (KSR 2) | books | missing | temporary differences between the tax treatment per account (`pl-set-tax-treatment`) and the books: the data is half there |
| 6.1.8 e | Dotacje rządowe | books | missing | |
| 6.1.8 f | Instrumenty finansowe | books | missing | amortised cost, market price, the rozporządzenie on financial instruments; the 2025 paper's task I-1 (short-term investments) |
| 6.1.8 g | Zmiany zasad rachunkowości (KSR 7) | kernel | partial | a changed rule is a new version from its date, earlier facts keep theirs (DIRECTION, the US register's cause 3); the retrospective restatement of comparatives is not |
| 6.1.8 h | Zdarzenia po dacie bilansu (KSR 7) | books | partial | period locks and late documents exist; adjusting events into the closed year (exam task I-3 did it: a bankruptcy ended on 30 January booked in 2024) need a rule about the open balance year |
| 6.1.8 i | Błąd lat ubiegłych (KSR 7) | kernel | partial | a correction is a new event with its own date; the posting to retained earnings (820) for a material prior-year error is missing |
| 6.1.9 | Rachunkowość spółek publicznych | host | — | |
| 6.2.1 a | Ciągłość zapisów i porównywalność przy zmianie podstawy | host | — | |
| 6.2.1 b | MSSF 1 | host | — | |
| 6.2.1 c | Uzgodnienia UoR–MSR | host | — | |
| 6.3.1 | Banki, domy maklerskie, fundusze, SKOK, instytucje płatnicze: wycena | host | — | not the pack's users; the 2025 paper's block II failed for both sitters (25%) |
| 6.3.2 | Banki i SKOK: operacje specyficzne (kredyty, papiery, depozyty, zobowiązania warunkowe, odsetki i prowizje, odpisy, rezerwy) | host | — | exam task II-1 (a SKOK: 32 accounts of the sector chart, 8 operations, 20 lines, including off-balance entries) was DONE through the kernel on the first sitting: the language carries a sector chart; the lessons are not in the pack |
| 6.3.3 | Inwentaryzacja w instytucjach finansowych | host | — | |
| 6.3.4 | Różnice wyceny w instytucjach finansowych | host | — | |
| 6.3.5 | Instrumenty pochodne, rachunkowość zabezpieczeń | host | — | |
| 6.3.6 | Wynik finansowy instytucji finansowych | host | — | |
| 6.3.7 | Firmy inwestycyjne, instytucje płatnicze | host | — | |
| 6.4.1 | Zakłady ubezpieczeń i fundusze emerytalne: wycena | host | — | not the pack's users; block III of the 2025 paper: bare model 83%, host 44% |
| 6.4.2 | Operacje ubezpieczeniowe: lokaty, reasekuracja, rezerwy techniczne, składki, odszkodowania, akwizycja | host | — | named by the host on five questions |
| 6.4.3 | Różnice wyceny w zakładach ubezpieczeń | host | — | |
| 6.4.4 | Pozostałe koszty i przychody techniczne | host | — | |
| 6.4.5 | Wynik techniczny i finansowy | host | — | |
| 6.4.6 | Fundusze emerytalne | host | — | |
| 6.5.1 | Jednostki sektora finansów publicznych z osobowością prawną | host | — | public universities, SPZOZ: named by the host |
| 6.5.2 | Rachunkowość budżetowa, JST | host | — | |
| 6.6.1 a | Zakładanie i przekształcanie jednostek | books | missing | share capital and contributions, founding costs; a sole trader becoming a sp. z o.o. is the first-phase users' own story |
| 6.6.1 b | Łączenie, podział, przejęcia | books | missing | |
| 6.6.1 c | Restrukturyzacja, likwidacja, upadłość | books | missing | the liquidation basis of valuation (6.1.1 b) |

## 7. Prawo podatkowe — część II

| Point | The point | Kind | Pack today | Note |
|---|---|---|---|---|
| 7.1.1 | Podatnicy CIT i PIT | books | covered | `pl-declare-company-form`, `pl-declare-cit-status` |
| 7.1.2 | Przedmiot opodatkowania | host | — | |
| 7.1.3 | Rok podatkowy | books | partial | the calendar year is assumed; a company's own fiscal year is not declared |
| 7.1.4 | Przychody podatkowe | books | partial | per account (`pl-set-tax-treatment`); the date of revenue and the exceptions of art. 12 are the host's |
| 7.1.5 | Zwolnienia przedmiotowe | host | — | |
| 7.1.6 | Szczególne zasady ustalania dochodu w PIT | books | partial | the sole trader's forms (KPiR, ryczałt) are declared; REVIEW "Sole trader" counts 41 of 46 points missing |
| 7.1.7 | Koszty uzyskania przychodu | books | partial | KUP/NKUP per account, vehicles and mileage; the general clause and the catalogue of art. 16 are the host's |
| 7.1.8 | Amortyzacja podatkowa | engine | missing | the schedule per asset (tax and balance-sheet rates), the KŚT; `fixed_asset` holds the data |
| 7.1.9 | Opodatkowanie leasingu | books | partial | `lease_contract`; the operating/finance split for tax is not computed |
| 7.1.10 | Podstawa i wysokość podatku | books | covered | 9% and 19%, the small-taxpayer limit (`cit_limit`, `pl-cit-advance`) |
| 7.1.11 | Obowiązki płatnika PIT | engine | missing | with payroll (3.3.3) |
| 7.1.12 | Podatek od dywidend | books | missing | with the distribution of result (6.1.7 c) |
| 7.1.13 | Przedsiębiorstwa powiązane | host | partial | the `intragroup` flag and intercompany matching exist |
| 7.1.14 | Ceny transferowe | host | — | documentation thresholds could be watched; not a rule yet |
| 7.1.15 | Pobór: zaliczki i rozliczenie roczne | books | partial | monthly advances recomputed (`pl-recompute-cit-advance`); the annual settlement and CIT-8 are missing |
| 7.1.16 | Deklaracje, zeznania, informacje | books | partial | `tax_declaration` for VAT; CIT-8, IFT, PIT-11 missing |
| 7.2.1 | Podatek od nieruchomości: podatnicy | host | — | a company's own property tax is a scheduled payment the pack could carry (rows 7.2.1–7.2.5); not a rule yet |
| 7.2.2 | Przedmiot i podstawa | host | — | |
| 7.2.3 | Powstanie i wygaśnięcie obowiązku | host | — | |
| 7.2.4 | Ulgi i wyłączenia | host | — | |
| 7.2.5 | Wymiar i pobór | host | — | |
| 7.3.1 | Podatek od środków transportowych: podatnicy | host | — | `vehicle` exists; the tax on trucks above 3.5 t is not computed |
| 7.3.2 | Przedmiot i podstawa | host | — | |
| 7.3.3 | Powstanie i wygaśnięcie obowiązku | host | — | |
| 7.3.4 | Ulgi i wyłączenia | host | — | |
| 7.3.5 | Wymiar i pobór | host | — | |
| 7.4.1 | Podatek rolny: podatnicy | host | — | outside the pack's users |
| 7.4.2 | Przedmiot i podstawa | host | — | |
| 7.4.3 | Powstanie i wygaśnięcie obowiązku | host | — | |
| 7.4.4 | Ulgi i wyłączenia | host | — | |
| 7.4.5 | Wymiar i pobór | host | — | |
| 7.5.1 | Podatek leśny: podatnicy | host | — | outside |
| 7.5.2 | Przedmiot i podstawa | host | — | |
| 7.5.3 | Powstanie i wygaśnięcie obowiązku | host | — | |
| 7.5.4 | Ulgi i wyłączenia | host | — | |
| 7.5.5 | Wymiar i pobór | host | — | |
| 7.6.1 | Rezydencja podatkowa | host | — | |
| 7.6.2 | Źródło i miejsce położenia | host | — | |
| 7.6.3 | Podwójne opodatkowanie i jego unikanie | host | — | |
| 7.6.4 | Umowy o unikaniu podwójnego opodatkowania | host | — | the 2025 paper's unkeyed question on the OECD model |
| 7.6.5 | Przedsiębiorstwa powiązane | host | — | |
| 7.6.6 | Konwencja Modelowa OECD | host | — | |
| 7.6.7 | Przeliczanie walut obcych dla podatku | books | covered | the NBP rate of the day before (`pl-register-fx-rate`, the VAT act art. 31a, CIT art. 12 ust. 2) |
| 7.6.8 | Harmonizacja podatkowa | host | — | |
| 7.6.9 | Zakaz dyskryminacji, procedura wzajemnego porozumiewania | host | — | |
| 7.6.10 | Wymiana informacji podatkowych | host | — | |
| 7.7.1 | Wpłaty z zysku: podmioty | host | — | state-owned companies: outside |
| 7.7.2 | Zaliczki i rozliczenie roczne | host | — | |
| 7.7.3 | Obowiązki raportowe | host | — | |
| 7.7.4 | Terminy | host | — | |
| 7.7.5 | Korekty | host | — | |
| 7.7.6 | Ograniczenia i zwolnienia | host | — | |
| 7.8.1 | Podatek od niektórych instytucji finansowych: podatnicy | host | — | outside |
| 7.8.2 | Przedmiot | host | — | |
| 7.8.3 | Podstawa | host | — | |
| 7.8.4 | Stawka | host | — | |
| 7.8.5 | Pobór | host | — | |
| 7.8.6 | Zwolnienia | host | — | |

## 8. Rachunek kosztów i rachunkowość zarządcza

| Point | The point | Kind | Pack today | Note |
|---|---|---|---|---|
| 8.1.1 | Koszty i pojęcia bliskoznaczne | host | — | |
| 8.1.2 | Rachunek kosztów w systemie informacyjnym; ewolucja | host | — | |
| 8.1.3 | Kwalifikacja kosztów dla rachunkowości finansowej | books | partial | costs by nature (4xx, `pl-set-cost-account`); by function (5xx) missing |
| 8.1.4 | Koszty a struktura organizacyjna | books | partial | descriptors on bookings exist in the language (cost centres as data); no pack rule uses them yet |
| 8.1.5 | Przekroje ewidencyjne a układ RZiS; uproszczenia | books | partial | as 6.1.7 a/b |
| 8.1.6 | Kalkulacja kosztów: rodzaje i metody | engine | missing | job, process, joint-product costing — formulas over descriptors, an engine where allocation iterates |
| 8.1.7 | Koszty działalności pomocniczej i ich rozliczanie | engine | missing | |
| 8.1.8 | Koszt wytworzenia; kalkulacja kosztu produktów | engine | missing | UoR art. 28 ust. 3; the products' valuation |
| 8.1.9 | Niewykorzystane zdolności produkcyjne | books | missing | named by the host on the 2025 paper |
| 8.1.10 | Rachunek kosztów standardowych | engine | missing | |
| 8.1.11 | Klasyfikacja kosztów dla zarządzania | host | — | |
| 8.1.12 | Rachunek kosztów zmiennych | host | — | |
| 8.1.13 | Obiekty kosztów; controlling | books | partial | descriptors; no controlling views |
| 8.1.14 | Rachunek kosztów działań (ABC) | host | — | |
| 8.1.15 | Target costing, kaizen, cykl życia | host | — | |
| 8.2.1 | Rola rachunkowości zarządczej w decyzjach | host | — | |
| 8.2.2 | Pomiar kosztów i korzyści | host | — | |
| 8.2.3 | Analiza koszty–rozmiary–zysk, marże, czynniki ograniczające, zamówienia specjalne | host | — | a break-even view is a formula over the books once costs are classified |
| 8.2.4 | Ośrodki odpowiedzialności, mierniki | host | partial | descriptors could name the centres |
| 8.2.5 | Planowanie i budżetowanie, ocena wykonania | engine | missing | budgets as data against actuals — a smart document (DIRECTION 2026-10-19) |
| 8.2.6 | Rachunek efektywności decyzji długookresowych | host | — | as 5.1.5 |
| 8.2.7 | Nowoczesne systemy pomiaru dokonań, Balanced Scorecard | host | — | |
| 8.2.8 | Lean accounting | host | — | |

## 9. Sprawozdania finansowe i ich analiza

| Point | The point | Kind | Pack today | Note |
|---|---|---|---|---|
| 9.1.1 | Funkcje sprawozdań | host | — | |
| 9.1.2 | Normy sprawozdawczości wg UoR i MSR | host | — | |
| 9.1.3 | Proces sporządzania sprawozdań | engine | missing | the year-end close: closing entries, the result account, the statements, the e-signature and filing |
| 9.1.4 | Terminy sporządzania i zatwierdzania (jednostkowe, łączne, skonsolidowane) | books | partial | the tax calendar view watches dates; the statutory dates (3 months to prepare, 6 to approve, 15 days to file with the KRS) are not scheduled |
| 9.1.5 | Odpowiedzialność za sprawozdawczość | host | — | |
| 9.2.1 | Części składowe sprawozdania | engine | missing | the e-Sprawozdania Finansowe XML writer in `markets/pl`, the way JPK is written |
| 9.2.2 | Bilans: ciągłość, układ, treść pozycji | engine | missing | the annex 1 layout as a view over account balances; the 2025 tasks name lines (B.III.1.b) |
| 9.2.3 | RZiS: okresy, układ, treść | engine | partial | `wynik` views; the statutory layout missing |
| 9.2.4 | Informacja dodatkowa: wprowadzenie, dodatkowe informacje i objaśnienia | engine | missing | notes as narrative views over the books (the briefing notion) |
| 9.2.5 | Rachunek przepływów pieniężnych | engine | missing | derived by the indirect method from postings — an engine |
| 9.2.6 | Zestawienie zmian w kapitale własnym | engine | missing | needs equity (6.1.2 c) |
| 9.2.7 | Sprawozdanie z działalności | host | — | |
| 9.2.8 | Sprawozdanie z płatności na rzecz administracji | host | — | |
| 9.2.9 | Jednostki z wewnętrznymi jednostkami sporządzającymi sprawozdania | host | — | |
| 9.2.10 | Wymagania prawa podatkowego wobec sprawozdań | books | partial | the statement goes with CIT-8; both missing |
| 9.3.1 | Obowiązek sporządzania sprawozdań skonsolidowanych | host | — | |
| 9.3.2 | Elementy sprawozdania skonsolidowanego | engine | missing | consolidation is an engine over several companies' books (the company key exists) |
| 9.3.3 | Zwolnienia z konsolidacji | host | — | |
| 9.3.4 | Metody konsolidacji | host | — | |
| 9.3.5 | Metoda pełna: eliminacje, dywidendy, wartość firmy, udziały niekontrolujące | engine | partial | intercompany matching and the `intragroup` flag give the eliminations their data |
| 9.3.6 | Metoda proporcjonalna | host | — | |
| 9.3.7 | Metoda praw własności | host | — | |
| 9.3.8 | Konsolidacja grup mieszanych | host | — | |
| 9.3.9 | Test na utratę wartości firmy | host | — | |
| 9.3.10 | Dokumentacja konsolidacyjna | host | — | |
| 9.3.11 | Waluta funkcjonalna, różnice kursowe z konsolidacji | host | — | |
| 9.4.1 | Obowiązki informacyjne spółek z rynku głównego GPW | host | — | |
| 9.4.2 | NewConnect | host | — | |
| 9.4.3 | Znaczące pakiety, władze spółek | host | — | |
| 9.4.4 | Inni emitenci | host | — | |
| 9.5.1 | Analiza finansowa: pojęcie, metody, cele | host | — | |
| 9.5.2 | Źródła informacji | host | — | |
| 9.5.3 | Przygotowanie sprawozdań do analizy | host | — | |
| 9.5.4 | Analiza przychodów | books | partial | `wynik-miesiace`, the morning views |
| 9.5.5 | Analiza kosztów | books | partial | as above |
| 9.5.6 | Analiza wyniku i rentowności | books | partial | `wynik`; ratios are formulas over the statements once they exist |
| 9.5.7 | Sytuacja majątkowa i finansowa, pokrycie majątku | books | missing | needs the balance sheet |
| 9.5.8 | Kapitał obrotowy | books | partial | receivables and payables open amounts; stock missing |
| 9.5.9 | Płynność i siła finansowa | books | partial | `naleznosci-wiek`, `poranek-pieniadze`; the ratios proper missing |
| 9.5.10 | Próg rentowności, dźwignie | host | — | as 8.2.3 |
| 9.5.11 | Symptomy zagrożenia upadłością | host | — | |
| 9.5.12 | Tworzenie wartości | host | — | |
| 9.5.13 | Pozycja na rynku kapitałowym | host | — | |

## 10. Rewizja finansowa oraz etyka

| Point | The point | Kind | Pack today | Note |
|---|---|---|---|---|
| 10.1.1 | Obowiązek badania i ogłaszania sprawozdań | host | partial | the thresholds of UoR art. 64 against the company's size are a rule the pack could carry; missing |
| 10.1.2 | Biegły rewident: zasady działania i odpowiedzialność | host | — | the profession's own; rows 10.1.2–10.1.11 and 10.4 are never a rule — and Rhea signs nothing |
| 10.1.3 | Firma audytorska; JZP | host | — | |
| 10.1.4 | Etyka zawodowa | host | — | |
| 10.1.5 | Krajowe standardy badania: przedmiot, struktura | host | — | |
| 10.1.6 | Kontrola jakości firm audytorskich | host | — | |
| 10.1.7 | Doskonalenie zawodowe | host | — | |
| 10.1.8 | Samorząd zawodowy | host | — | |
| 10.1.9 | Postępowanie dyscyplinarne | host | — | |
| 10.1.10 | Nadzór publiczny (PANA) | host | — | |
| 10.1.11 | Nadzór nad JZP | host | — | |
| 10.2.1 | Cel badania i przeglądu | host | — | |
| 10.2.2 | Planowanie; dowody badania | kernel | covered | provenance on every fact and replay of the whole log are audit evidence by construction; what an auditor samples, Rhea reproduces |
| 10.2.3 | Szczególne problemy: pierwsze badanie, inwentaryzacja, zobowiązania warunkowe, szacunki, dane porównawcze, zdarzenia po dniu bilansowym, kontynuacja, sprawozdanie z działalności, małe jednostki, podmioty powiązane, nadużycia | host | partial | subsequent events (period locks, late documents) and related parties (`intragroup`) are data the auditor can read; the judgements are the host's |
| 10.2.4 | Badanie sprawozdań skonsolidowanych | host | — | |
| 10.2.5 | Programy badania pozycji sprawozdania | host | — | |
| 10.2.6 | Wykorzystanie pracy innych biegłych, audytu wewnętrznego, rzeczoznawców | host | — | |
| 10.2.7 | Rezultaty badania: sprawozdanie z badania, opinia | host | — | |
| 10.2.8 | Komunikacja z komitetem audytu | host | — | |
| 10.2.9 | Technologie informacyjne w rewizji | kernel | covered | a read-only analysis session that cannot write, reach files or the network (invariant 3); replay as the auditor's re-performance |
| 10.2.10 | Dokumentacja badania i archiwizacja | host | — | |
| 10.3.1 | Inne usługi atestacyjne: kwalifikacja | host | — | |
| 10.3.2 | Wykonywanie innych usług atestacyjnych | host | — | |
| 10.3.3 | Sprawozdanie z innej usługi atestacyjnej | host | — | |
| 10.3.4 | Usługi pokrewne: uzgodnione procedury, kompilacja | host | — | |
| 10.4.1 | Zasady etyki zawodowej | host | — | |
| 10.4.2 | Niezależność biegłego rewidenta | host | — | |

## Sources

- KRBR resolution 1695/34/2026 of 27 January 2026 on the detailed syllabus (pibr.org.pl).
- KRBR resolution 841/20/2024 on the literature and the basic legal acts per subject.
- The Commission's sample papers with model solutions, 2025 (pibr.org.pl/pl/egzaminy-pytania);
  the sittings in research_notes/pibr-egzaminy/results/round2 (kept out of git).
- The Poland pack v31, the common packs, `markets/pl`; REVIEW.md for the first-year points.
