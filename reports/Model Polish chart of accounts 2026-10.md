# A model Polish chart of accounts

*2026-10-25. KK asked for research toward a model Polish chart that satisfies Polish GAAP (the
Accounting Act, UoR, and the national standards, KSR), with enough analytical accounts for the
complete statements and every submission the law requires ("F1 or similar"). Three research notes
from primary sources (the law in force, the Ministry's XML schemas and tag dictionaries, GUS's 2026
programme) sit in `research_notes/Model Polish chart of accounts/`; this report puts them together
and proposes the chart. The chart itself is data: `model-chart.csv` in the same folder, built by
`modelchart.py` from the Ministry's tag dictionary. It is a research proposal: no practising
accountant has reviewed it yet.*

## In short

- **The law asks each company for its own chart, and offers no model.** UoR art. 10 requires a
  *zakładowy plan kont* in the accounting policy; model charts are optional (art. 83), and the only
  ministerial one in force is for banks. Offices and software vendors each use their own, broadly
  alike in groups 0–8. There is room for a standard.
- **The state has defined the finest split it needs.** JPK_KR_PD, the new books file, requires every
  account to carry one of **244 tags** (Annex 7 of Dz.U. 2024 poz. 1314), and accounts used for the
  income tax one of **28 tax tags**. An account that mixes things needing different tags must be
  split. Companies filing VAT monthly send it for **2026** (by 31 July 2027); everyone else from
  **2027**. Small and micro companies get no simpler list.
- **The proposal: a chart built on those tags.** 117 synthetic accounts, 464 accounts in all, every
  analytic account carrying exactly one tag. All 244 tags are covered (two umbrella tags are left
  unused on purpose, as the Ministry's FAQ asks when their parts are kept apart). The tags name the
  balance-sheet and profit-and-loss positions in both variants, so **one mark per account gives the
  JPK_KR_PD file, the balance sheet and the P&L**; a company on this chart files JPK_KR_PD with no
  tagging by hand. That is what could make it the standard.
- **158 accounts are what every sp. z o.o. needs;** the rest come with the business: employing
  people, trade, production, the by-function P&L, foreign currency, groups, investments, partnerships,
  Estonian CIT.
- **F-01/I-01 ("F1") exists** in 2026: companies with 50 or more people file quarterly, 10–49
  half-yearly. It asks for costs by nature even from a company reporting by function, and for a few
  splits the chart builds in (energy, subcontracting, wages from employment contracts, training,
  business travel, insurance, excise, exports, provisions released and created, receivables written
  off). SP, the yearly survey, adds little more; F-02 no longer exists.
- **Tax travels with the accounts.** What Rhea's rules can tell at posting (representation, fines,
  donations, the non-deductible part of a car, unpaid wages and contributions, accrued interest) goes
  to analytic accounts with their tax tag; R&D relief, IP box and book-to-tax items without a posting
  go to off-balance registers in group 9, which carry their own tax tags.
- **Every counterparty needs two marks:** *related* (the same capital group) and *equity stake held*.
  With the due date and the country, they decide which account a posting goes to.
- **Rhea's chart today is far from it:** 55 accounts, no analytics anywhere, group 4 numbered one place
  off every published chart, *Podatki i opłaty* missing, no VAT-settlement or CIT account, no deferred
  income (840), contra accounts typed as assets, current and deferred tax on one account.
- **Changing a chart during the year:** adding accounts is allowed at any time; a change that moves an
  item to another statement position is a policy change and takes effect from the first day of the
  year; a closed year cannot change (art. 8 ust. 2, art. 12 ust. 5, KSR 7).

## The idea: one tag per account

An example. Rhea books a sales invoice to a company in the same capital group, due in 30 days. The
counterparty is marked *related*; the due date is within 12 months; so the receivable goes to
**201-1-1** *Należności z tytułu dostaw i usług od jednostek powiązanych — do 12 miesięcy*, whose tag
is **`BABII1a_D12`**. That one mark says where the amount stands in the balance sheet (B.II.1.a, "do
12 miesięcy") and how the account appears in JPK_KR_PD. The revenue goes to **700-1**, tag
`RPAI_RKAI_POW`: the same line in both P&L variants. Nothing is tagged at the end of the year,
because the account was tagged when the chart was made.

What a tag does not give: the cash-flow statement (it needs each cash posting's kind, or the changes
of balances by the indirect method), the notes (*informacja dodatkowa*), and the shortened forms of
small and micro companies, which group the same positions more coarsely (the statements note, §3.12,
to confirm line by line).

## What the law and the filings ask

| | Who | What it asks of the chart | Note |
|---|---|---|---|
| Balance sheet, P&L, notes (annex 1) | every company not using a simplification | the annex positions, both P&L variants | statements.md §3 |
| Cash flow, changes in equity | every company that is neither small nor micro, audited or not (since 2024) | cash-flow kind on cash postings; equity movement types | statements.md §1.2, §3.4–3.6 |
| Small (annex 5), micro (annex 4) forms | small: up to 33 mln zł assets / 66 mln zł revenue / 50 people; micro: 2 / 4 / 10 — two of three, this year and last | coarser groupings of the same positions | statements.md §1.2–1.3 |
| e-sprawozdanie (XML) | every company, to KRS or Szef KAS | the position identifiers of SFJINZ(2), SFJMAZ(2), SFJMIZ(2) | statements.md §2 |
| **JPK_KR_PD** (+ JPK_ST_KR, fixed assets) | CIT taxpayers keeping books: monthly-VAT filers for 2026, everyone for 2027 | one Annex-7 tag per account, one tax tag where it serves the tax, the parent account; counterparty code, NIP, country; KSeF numbers; currency amounts; fixed assets with tax and book values side by side | filings.md §1 |
| CIT-8 and attachments | every CIT taxpayer | capital gains apart from other income; deductible apart from not; financing costs; R&D by its nine categories; IP-box per right | filings.md §2 |
| F-01/I-01, SP | 10 or more people | costs by nature always; the splits above | filings.md §4 |
| NBP | positions with non-residents above PLN 10 mln (quarterly) | residents apart from non-residents; the counterparty's country | filings.md §5 |
| TPR-C | related parties above the thresholds | related-party analytics | filings.md §6 |

## The chart

Synthetic accounts, with the number of analytic accounts under each (the last levels of settlement
accounts are the counterparty and the document, kept as links, not as accounts). Scope: **W** every
full-books company; **Z** employers; **H** trade with stock; **P** production; **K** by-function
P&L; **$** foreign currency; **G** related parties; **F** financial investments; **S** partnerships;
**E** Estonian CIT; **X** other cases. "Rhea dziś" is today's code; in bold where it differs.

| Konto | Nazwa | Analityki | Zakres | Rhea dziś |
|---|---|---|---|---|
| 010 | Środki trwałe | 5 | W | 010 |
| 020 | Wartości niematerialne i prawne | 3 | W | 020 |
| 030 | Długoterminowe aktywa finansowe | 12 | F | 030 |
| 035 | Odpisy aktualizujące długoterminowe aktywa finansowe | 12 | F | — |
| 040 | Inwestycje w nieruchomości i prawa oraz inne inwestycje długoterminowe | 3 | F | — |
| 071 | Umorzenie środków trwałych | 5 | W | 071 |
| 072 | Odpisy aktualizujące środki trwałe i środki trwałe w budowie | 7 | X | — |
| 075 | Umorzenie wartości niematerialnych i prawnych | 3 | W | 075 |
| 076 | Odpisy aktualizujące wartości niematerialne i prawne | 4 | X | — |
| 077 | Umorzenie i odpisy aktualizujące inwestycji długoterminowych | 5 | F | — |
| 080 | Środki trwałe w budowie | — | W | 080 |
| 081 | Zaliczki na środki trwałe w budowie i na wartości niematerialne i prawne | 2 | W | — |
| 100 | Kasa | — | W | 100 |
| 130 | Rachunek bieżący | — | W | 130 |
| 131 | Rachunki bankowe w walutach obcych | — | $ | — |
| 132 | Rachunek VAT | — | W | — |
| 134 | Kredyty bankowe | 2 | X | — |
| 135 | Inne rachunki bankowe | 2 | X | — |
| 136 | Inne aktywa pieniężne | — | X | — |
| 139 | Środki pieniężne w drodze | — | W | **149** |
| 140 | Krótkoterminowe aktywa finansowe | 9 | F | — |
| 145 | Odpisy aktualizujące krótkoterminowe aktywa finansowe | 11 | F | — |
| 201 | Rozrachunki z odbiorcami | 6 | W | 201 |
| 202 | Rozrachunki z dostawcami | 6 | W | 202 |
| 205 | Rozrachunki z tytułu zaliczek na dostawy i usługi | 4 | W | — |
| 221 | VAT naliczony i jego rozliczenie | 2 | W | 221 |
| 222 | VAT należny | — | W | 222 |
| 223 | Rozrachunki z urzędem skarbowym z tytułu VAT | 2 | W | — |
| 224 | Rozrachunki z tytułu podatku dochodowego | 3 | W | — |
| 225 | Pozostałe rozrachunki publicznoprawne | 6 | W | 225 |
| 229 | Rozrachunki z ZUS i innymi instytucjami | 7 | Z | 229 |
| 231 | Rozrachunki z tytułu wynagrodzeń | — | Z | 231 |
| 234 | Pozostałe rozrachunki z pracownikami | 2 | Z | — |
| 240 | Pożyczki otrzymane | 6 | W | 240 |
| 241 | Rozliczenie niedoborów, szkód i nadwyżek | — | X | — |
| 242 | Rozrachunki z właścicielami (wspólnikami spółek osobowych) | — | S | — |
| 243 | Rozrachunki z udziałowcami (akcjonariuszami) | 4 | W | **241** |
| 244 | Zobowiązania z tytułu leasingu finansowego | 2 | X | — |
| 246 | Należności dochodzone na drodze sądowej | — | X | — |
| 247 | Rozliczenia z operatorami płatności | — | X | — |
| 248 | Zobowiązania z tytułu emisji dłużnych papierów wartościowych i wekslowe | 4 | X | — |
| 249 | Pozostałe rozrachunki | 10 | W | 249 |
| 250 | Należne wpłaty na kapitał podstawowy | — | W | — |
| 280 | Odpisy aktualizujące należności | 15 | W | 280 |
| 300 | Rozliczenie zakupu | — | H | 300 |
| 310 | Materiały | — | P | 310 |
| 330 | Towary | — | H | 330 |
| 340 | Odchylenia od cen ewidencyjnych materiałów i towarów | 2 | X | — |
| 349 | Odpisy aktualizujące wartość zapasów | 2 | H | 349 |
| 400 | Amortyzacja | 4 | W | **401** |
| 401 | Zużycie materiałów i energii | 4 | W | **402** |
| 402 | Usługi obce | 10 | W | **403** |
| 403 | Podatki i opłaty | 8 | W | — |
| 404 | Wynagrodzenia | 4 | Z | 404 |
| 405 | Ubezpieczenia społeczne i inne świadczenia | 8 | Z | 405 |
| 409 | Pozostałe koszty rodzajowe | 7 | W | 409 |
| 490 | Rozliczenie kosztów | — | W | 490 |
| 500 | Koszty działalności podstawowej — produkcyjnej i usługowej | — | P | — |
| 510 | Koszty działalności podstawowej — handlowej | — | K | — |
| 520 | Koszty wydziałowe | — | P | — |
| 527 | Koszty sprzedaży | — | K | — |
| 530 | Koszty działalności pomocniczej | — | P | — |
| 550 | Koszty ogólnego zarządu | — | K | — |
| 580 | Rozliczenie kosztów działalności | — | P | — |
| 600 | Produkty gotowe | — | P | — |
| 601 | Półprodukty | — | P | — |
| 602 | Produkty w toku | — | P | — |
| 620 | Odchylenia od cen ewidencyjnych produktów | 2 | P | — |
| 621 | Odpisy aktualizujące wartość produktów | 2 | P | — |
| 640 | Czynne rozliczenia międzyokresowe kosztów | 2 | W | 640 |
| 641 | Bierne rozliczenia międzyokresowe kosztów | 3 | W | 641 |
| 650 | Aktywa z tytułu odroczonego podatku dochodowego | — | W | 650 |
| 651 | Odpisy aktualizujące aktywa z tytułu odroczonego podatku dochodowego | — | X | — |
| 652 | Odpisy aktualizujące czynne rozliczenia międzyokresowe | 2 | X | — |
| 700 | Sprzedaż produktów (wyrobów i usług) | 2 | W | 700 |
| 701 | Koszt sprzedanych produktów | 2 | K | — |
| 730 | Sprzedaż towarów | 2 | H | 730 |
| 731 | Wartość sprzedanych towarów w cenie nabycia | 2 | H | 731 |
| 750 | Przychody finansowe | 13 | W | 750 |
| 751 | Koszty finansowe | 10 | W | 751 |
| 760 | Pozostałe przychody operacyjne | 9 | W | 760 |
| 761 | Pozostałe koszty operacyjne | 12 | W | 761 |
| 790 | Obroty wewnętrzne | — | X | — |
| 791 | Koszt obrotów wewnętrznych | — | X | — |
| 795 | Zmiana stanu produktów | — | P | — |
| 796 | Koszt wytworzenia produktów na własne potrzeby jednostki | — | P | — |
| 801 | Kapitał podstawowy | — | W | 801 |
| 802 | Kapitał zapasowy | 4 | W | 802 |
| 803 | Pozostałe kapitały rezerwowe | 5 | X | — |
| 805 | Udziały (akcje) własne | — | X | — |
| 813 | Kapitał z aktualizacji wyceny | 2 | X | 813 |
| 820 | Rozliczenie wyniku finansowego | 6 | W | 820 |
| 830 | Rezerwy na zobowiązania | 4 | W | 830 |
| 835 | Rezerwa z tytułu odroczonego podatku dochodowego | — | W | 835 |
| 840 | Rozliczenia międzyokresowe przychodów | 3 | W | — |
| 851 | Zakładowy fundusz świadczeń socjalnych | — | X | — |
| 852 | Inne fundusze specjalne | — | X | — |
| 860 | Wynik finansowy | — | W | 860 |
| 870 | Podatek dochodowy | 2 | W | 870 |
| 871 | Pozostałe obowiązkowe zmniejszenia zysku (zwiększenia straty) | — | X | — |
| 910 | Obce środki trwałe (najem, dzierżawa, leasing operacyjny, użyczenie) | — | X | — |
| 911 | Środki trwałe i WNiP w likwidacji lub wycofane z używania | — | X | — |
| 912 | Składniki majątku o niskiej wartości w używaniu | — | X | — |
| 920 | Należności warunkowe | — | X | — |
| 921 | Zobowiązania warunkowe (poręczenia, gwarancje, weksle) | — | X | — |
| 922 | Zabezpieczenia ustanowione na majątku | — | X | — |
| 930 | Zapasy obce | — | X | — |
| 941 | Przychody zwolnione z opodatkowania (różnice trwałe) | 4 | W | — |
| 942 | Przychody niepodlegające opodatkowaniu w roku bieżącym | — | W | — |
| 943 | Przychody podatkowe roku bieżącego ujęte w księgach lat ubiegłych | — | W | — |
| 944 | Przychody podatkowe nieujmowane w księgach | — | X | — |
| 945 | Koszty niestanowiące KUP (różnice trwałe) | 4 | W | — |
| 946 | Koszty nieuznawane za KUP w roku bieżącym | — | W | — |
| 947 | KUP roku bieżącego ujęte w księgach lat ubiegłych | — | W | — |
| 948 | KUP nieujmowane w księgach (np. nadwyżka jednorazowej amortyzacji podatkowej) | — | X | — |
| 949 | Ulgi: B+R i IP Box | 3 | W | — |
| 999 | Rozliczenie kont pozabilansowych | — | X | — |

Every row of `model-chart.csv` carries: the code, its parent and level, the Polish name, the nature
(active, passive, both, cost, revenue, contra, clearing, off-balance), the scope, the Annex-7 tag and
the position it names, the tax tag where there is one, the F-01 split, today's Rhea code, and a note
where something is to verify.

## What Rhea's account must carry

KK's A-codes, with the law naming most of them:

- the **Annex-7 tag** (S_12_1) and the **tax tag** (S_12_3), per financial year;
- the **parent** and level (JPK S_3; UoR's "link to the general ledger");
- the **nature**, which decides how a balance is closed and presented (Rhea's `type` mixes nature and
  role today, and types contra accounts as assets);
- the **scope** — which companies the account belongs to;
- the **F-01 line** for cost and revenue accounts;
- the **cash-flow kind** for cash and bank accounts;
- **validity** from and to, and the name's history.

And on the counterparty and the posting: related, equity stake held, country and residency, NIP; the
KSeF number, the currency amount, the due date. Rhea's rules then pick the analytic account from
these marks. *One thing to check in the kernel before the pack is written:* that a rule's condition
can read a mark of the counterparty, not only a field of the fact.

## From today's chart to the new pack

KK decided on 2026-10-25 that Helios, the role-model demo company, gets a full, new Polish pack. It
would be built on this chart: the full-books pack's rules rewritten to book to the analytic accounts
(the supplier's invoice to 202-3-1 when the supplier is neither related nor an investee and the
invoice is due within a year), the accounts carrying their tags from the start, and the files that
read them added to the plan: the JPK_KR_PD with JPK_ST_KR, the e-sprawozdanie XML, F-01/I-01
(FUNCTIONS.md 120, 121, 62). Helios's present log cannot be wiped (it is append-only), so the new
Helios is a new company with the same story.

Today's 55 accounts map onto the new chart code by code (`charts-in-practice.md` §4): 42 stay as they
are, with analytics added; 401–403 move to 400–402 and 403 becomes *Podatki i opłaty*; 030 becomes all
long-term financial assets; 149 becomes 139; dividends move from 241 to 243; asset advances from 249 to
081; CIT from 225 to 224; deferred tax gets its own analytic of 870; 756 dissolves into the FX accounts
of 750 and 751.

## For KK to decide

1. **One tag per account as the principle of Rhea's Polish chart.**
   *Example:* receivables split by related / investee / other and by up to or over 12 months, six
   accounts where today there is one.
   (a) Yes: the chart is built from the Ministry's tags, so the books and the file never disagree.
   (b) A shorter chart, tagged at year end by hand or by Rhea.
   *Recommendation: (a).* It is the point of a standard, and the law's split is what an auditor and
   the tax office will read.
2. **Group 4 numbered as every published chart does** (400 Amortyzacja, 401 Zużycie materiałów i
   energii, 402 Usługi obce, 403 Podatki i opłaty) in the new pack.
   *Recommendation: yes.* The new Helios has no history to break, and an accountant reading "403"
   expects taxes.
3. **Who reviews it before it is called a standard.**
   *Recommendation:* one practising accountant reads `model-chart.csv` and the points below, then
   SKwP, the accountants' association. SKwP asked the Ministry for a simpler dictionary for small
   companies, so they know exactly this problem.

## What is still to verify

From the three notes, merged; the details and sources are there.

1. Whether "up to / over 12 months" counts the original term or the time remaining at the balance
   date.
2. Where accrued costs for services received but not invoiced are shown (other accruals or
   liabilities); 641-1 is tagged as other accruals for now.
3. How a two-sided account is tagged when its balance changes side: an overpaid receivable, VAT and
   CIT settlement accounts, account 300.
4. The tax tag of reversed write-offs and released provisions (760-3, 760-4-4), of revaluation of
   financial assets (750-4, 751-3), and whether a write-off on receivables (761-2) is a timing or a
   permanent difference in each case.
5. Unrealised FX differences: taxed only under the tax method of CIT art. 9b, so their tax tag
   follows the company's choice.
6. KUP and NKUP as analytic accounts (proposed for what rules book) or as off-balance registers
   (proposed for R&D, IP box and items without a posting): an accountant's preference.
7. Where conversion roundings go (Rhea's 756 today).
8. How *dopłaty* are shown (a separate equity line, art. 36 ust. 2e) and tagged; the ZFŚS threshold.
9. Which entities must fill the structured CIT note of the e-sprawozdanie, especially small and
   micro.
10. The tags of account 500 (its balance is work in progress) and 510.
11. Renaming and renumbering accounts mid-year: no explicit rule; reading of the 2010 KSR Committee
    position.
12. Vendor default charts (Optima, Symfonia, enova, Rewizor) and SKwP's own, not public; a client's
    export would show them.
13. Whether the Ministry adds Estonian-CIT tax tags, a dictionary for small companies, or an IFRS
    dictionary.

## Sources

The three notes, each with its sources, primary first:
[statements.md](../research_notes/Model%20Polish%20chart%20of%20accounts/statements.md) (UoR Dz.U. 2026
poz. 522 and its amendments, the e-sprawozdanie XSDs, KSR 1–16),
[filings.md](../research_notes/Model%20Polish%20chart%20of%20accounts/filings.md) (JPK_KR_PD(1) XSD and
Dz.U. 2024 poz. 1314 with its FAQ, CIT-8(34), the GUS programme Dz.U. 2025 poz. 1842 and the 2026
forms, the NBP regulation), and
[charts-in-practice.md](../research_notes/Model%20Polish%20chart%20of%20accounts/charts-in-practice.md)
(UoR art. 8, 10, 16–17, KSR 7, the KSR Committee's 2010 position, published model charts, Rhea's pack
`pl` v44). The chart: [model-chart.csv](../research_notes/Model%20Polish%20chart%20of%20accounts/model-chart.csv),
built by [modelchart.py](../research_notes/Model%20Polish%20chart%20of%20accounts/modelchart.py).
