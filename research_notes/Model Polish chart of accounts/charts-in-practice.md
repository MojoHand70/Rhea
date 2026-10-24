# Model Polish chart of accounts — the law on the chart, charts in practice, and Rhea's gap

Researched 2026-10-25 for Rhea's model Polish chart of accounts (*wzorcowy plan kont*). Strand: **the law on
the chart itself, the model charts used in practice and their analytical structure, a gap analysis of the
`pl` pack's chart, and a proposed skeleton**. Two sibling notes cover the statements' positions and the
filings (see [filings.md](filings.md) for JPK_KR_PD tags, CIT-8, GUS and NBP — not repeated here).

**Status: research proposal for review by a practising accountant (główny księgowy / biegły rewident).
Nothing here is a decision.** Primary sources first (Dziennik Ustaw via ELI, KSR published by the Ministry,
the KSR Committee's own position); professional commentary and vendor documentation second. Anything not
confirmed is marked *to verify*. Where a conclusion is my own reading, it says "my reading". Downloaded
files were treated as data only.

---

## 0. The answers in brief

- **The law requires each company's own chart, not a model one.** UoR art. 10 ust. 1 pkt 3 lit. a asks for
  a *zakładowy plan kont* with the list of general-ledger accounts, the classification principles, the
  principles of the subsidiary (analytic) ledgers and their link to the general ledger. Model charts are
  optional (art. 83 ust. 1); the only ministerial model chart found in force under art. 83 ust. 2 is the
  banks' one, and none for ordinary companies; the public-sector chart rests on the public-finance act. Commercial practice runs on publishers'
  and software vendors' model charts, which share groups 0–8 and broadly the same three-digit numbers.
- **Mid-year changes.** Adding accounts — analytic or synthetic — during the year is allowed: the head of
  the entity "ustala w formie pisemnej i aktualizuje" the documentation (art. 10 ust. 2), and a rule for
  events that did not occur before is not a policy change (KSR 7 pkt 3.9). Changing the chart at account
  level is not automatically a policy change "o ile bez zmian pozostają zasady klasyfikacji zdarzeń
  gospodarczych i agregacji danych księgowych", provided the old-to-new mapping is documented and approved
  by the head of the entity (Stanowisko KSR 2010 pkt 37). A change that alters classification to statement
  positions, valuation, the result or presentation **is** a policy change and takes effect "ze skutkiem od
  pierwszego dnia roku obrotowego, bez względu na datę podjęcia decyzji" (art. 8 ust. 2; KSR 7 pkt 3.3,
  3.7, 3.8). Changing calculation algorithms mid-year is not permitted (Stanowisko 2010 pkt 38). A closed
  year cannot be changed at all (art. 12 ust. 5).
- **Rhea's chart (pack `pl` v44): 55 accounts (50 synthetic, 5 NKUP analytics), 6 tax treatments, no
  statement mapping.** Missing for every full-books company: *Podatki i opłaty* (an RZiS line), *Rozliczenia
  międzyokresowe przychodów*, a VAT settlement account, a CIT account apart from 225, *Pozostałe
  rozrachunki z pracownikami*, *Rozrachunki z udziałowcami* beyond dividends, *Należne wpłaty na kapitał*,
  advances on fixed assets apart from 249, advances on supplies apart from trade settlements, the VAT bank
  account, the off-balance group (and, for most companies, bank loans). Group 4
  is numbered one place off every published chart checked (Rhea 401 = Amortyzacja; everyone else 400).
  No account has the analytics or attributes the statements and JPK_KR_PD need (relation, maturity,
  currency, counterparty, gross/depreciation/write-down, long/short term, current/deferred tax).
- **Skeleton proposed in §4: 112 synthetic accounts in groups 0–9** (54 for every full-books company,
  the rest by circumstance), each with its analytic levels.

---

## 1. The law on the chart

Statute: *ustawa z dnia 29 września 1994 r. o rachunkowości* (UoR), consolidated text
[Dz.U. 2026 poz. 522](https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf) (state of law at
26 March 2026; the act's ISAP page is
[WDU19941210591](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU19941210591)). All quotations below
are from that text unless another source is named.

### 1.1 What the chart must contain, who sets it, in what language

- **Art. 10 ust. 1 pkt 3 lit. a** — the accounting-policy documentation must describe the way the books are
  kept, "w tym co najmniej: a) zakładowego planu kont, ustalającego wykaz kont księgi głównej, przyjęte
  zasady klasyfikacji zdarzeń, zasady prowadzenia kont ksiąg pomocniczych oraz ich powiązania z kontami
  księgi głównej". Four things, then: the **list of synthetic accounts**, the **classification principles**,
  the **principles of the analytic ledgers**, and **how analytics link to synthetics**.
- **Art. 10 ust. 1 pkt 3 lit. b–c** — for computer books, also the list of data sets and "opisu systemu
  informatycznego, zawierającego wykaz programów, procedur lub funkcji … wraz z opisem algorytmów
  i parametrów … a ponadto określenie wersji oprogramowania i daty rozpoczęcia jego eksploatacji". For
  Rhea this is where rule versions and their effective dates belong.
- **Art. 10 ust. 2** — "Kierownik jednostki ustala w formie pisemnej i aktualizuje dokumentację, o której
  mowa w ust. 1." The chart is set **and updated** by the head of the entity, in writing.
- **Art. 10 ust. 3** — where the Act is silent, entities "mogą stosować krajowe standardy rachunkowości".
- **Art. 9** — "Księgi rachunkowe prowadzi się w języku polskim i w walucie polskiej." The KSR Committee
  reads this as covering "nazwy kont" first of all
  ([Stanowisko KSR 2010, pkt 14 lit. a](https://www.gov.pl/attachment/ecd3a097-3241-4492-b891-16b811b85aae)).
- **Art. 83 ust. 1** — "W celu ujednolicenia zasad grupowania operacji gospodarczych i ograniczenia nakładu
  pracy związanego z ustaleniem zakładowych planów kont mogą być stosowane wzorcowe plany kont." **Ust. 2**
  lets the Minister of Finance issue model charts by regulation for banks (pkt 1), entities under the
  securities law (pkt 2), investment funds (pkt 3), insurers and pension funds (pkt 4), "dla pozostałych
  jednostek" (pkt 6) and SKOK (pkt 7). The one found in force is the bank chart
  ([Dz.U. 2019 poz. 946](https://eli.gov.pl/eli/DU/2019/946/ogl), consolidated text). **No regulation under
  pkt 6 for ordinary companies was found** (*to verify* whether one ever existed before 1995 — not found).
  The public-sector chart (*plan kont dla jednostek budżetowych*) is issued under art. 40 ust. 4 pkt 1
  lit. a of the public-finance act, not under art. 83
  ([Dz.U. 2020 poz. 342](https://eli.gov.pl/eli/DU/2020/342/ogl/pol/pdf); a newer consolidated text,
  Dz.U. 2026 poz. 909, was found by search but not read — *to verify*).

### 1.2 Synthetic and analytic accounts, and which analytics are mandatory

- **Art. 13 ust. 1** — the books are the journal, the general ledger, the subsidiary ledgers, the trial
  balances ("zestawienia: obrotów i sald kont księgi głównej oraz sald kont ksiąg pomocniczych") and the
  inventory.
- **Art. 15 ust. 1** — "Konta księgi głównej zawierają zapisy o zdarzeniach w ujęciu systematycznym. Na
  kontach księgi głównej obowiązuje ujęcie zarejestrowanych uprzednio lub równocześnie w dzienniku zdarzeń,
  zgodnie z zasadą podwójnego zapisu."
- **Art. 16 ust. 1** — "Konta ksiąg pomocniczych zawierają zapisy będące uszczegółowieniem i uzupełnieniem
  zapisów kont księgi głównej. Prowadzi się je w ujęciu systematycznym jako wyodrębniony system ksiąg,
  kartotek (zbiorów kont), komputerowych zbiorów danych, uzgodniony z saldami i zapisami na kontach księgi
  głównej." Ust. 2 allows natural units in the subsidiary ledgers.
- **Art. 17 ust. 1 — the mandatory subsidiary ledgers.** "Konta ksiąg pomocniczych prowadzi się
  w szczególności dla:
  1) środków trwałych, w tym także środków trwałych w budowie, wartości niematerialnych i prawnych oraz
     dokonywanych od nich odpisów amortyzacyjnych lub umorzeniowych;
  2) rozrachunków z kontrahentami;
  3) rozrachunków z pracownikami, a w szczególności jako imienną ewidencję wynagrodzeń pracowników
     zapewniającą uzyskanie informacji, z całego okresu zatrudnienia;
  4) operacji sprzedaży (kolejno numerowane własne faktury i inne dowody, ze szczegółowością niezbędną do
     celów podatkowych);
  5) operacji zakupu (obce faktury i inne dowody, ze szczegółowością niezbędną do wyceny składników aktywów
     i do celów podatkowych);
  6) kosztów i istotnych dla jednostki składników aktywów;
  7) operacji gotówkowych w przypadku prowadzenia kasy."
  "W szczególności" makes the list a minimum, not a closed set.
- **Art. 17 ust. 2** — for current tangible assets (stock) the head of the entity chooses one method:
  quantity-and-value, quantity only, value only (retail points and stores), or expensing on purchase with a
  count and cost adjustment "niepóźniej niż na dzień bilansowy" (pkt 4).
- **Art. 18 ust. 1–2** — a trial balance of the general ledger at least monthly, with opening balances,
  period and year-to-date turnover and closing balances; "Co najmniej na dzień zamknięcia ksiąg
  rachunkowych sporządza się zestawienia sald wszystkich kont ksiąg pomocniczych".
- **Art. 23 ust. 2 pkt 5** — every entry must carry "oznaczenie kont, których dotyczy"; **ust. 3** — entries
  in foreign currency are made "w sposób umożliwiający ustalenie kwoty operacji w walucie polskiej i obcej"
  (hence analytics or attributes per currency).
- **Art. 7 ust. 3** — "Nie można kompensować ze sobą wartości różnych co do rodzaju aktywów i pasywów,
  przychodów i kosztów związanych z nimi" — so a two-sided account that mixes kinds (e.g. VAT refund due and
  CIT payable on one account) needs analytics that keep the kinds apart.
- **Tax law adds analytics the UoR does not.** CIT act ([Dz.U. 2026 poz. 554](https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf))
  art. 9 ust. 1: records "w sposób zapewniający określenie wysokości dochodu (straty), podstawy opodatkowania
  i wysokości należnego podatku za rok podatkowy"; ust. 1b: R&D relief claimants must "wyodrębnić koszty
  działalności badawczo-rozwojowej"; art. 24e ust. 1 pkt 1: IP Box taxpayers must "wyodrębnić każde
  kwalifikowane prawo własności intelektualnej w prowadzonej ewidencji rachunkowej". The JPK_KR_PD
  regulation ([Dz.U. 2024 poz. 1314](https://eli.gov.pl/eli/DU/2024/1314/ogl/pol/pdf)) § 2 ust. 1 adds the
  counterparty NIP and the KSeF number on entries (pkt 1–2), account tags (pkt 3), fixed-asset data (pkt 4)
  and the book-to-tax differences in eight categories (pkt 5 lit. a–h); § 2 ust. 3: the tags are assigned
  "w oparciu o kryteria klasyfikacyjne przyjęte dla zapisów zdarzeń w księgach, wynikające z przepisów
  o rachunkowości". What the tags require of the chart is in [filings.md §1.5](filings.md).
- **The statements drive further analytics** (Annex 1 UoR; positions are the sibling strand's subject):
  the balance sheet splits receivables and payables by relation (*jednostki powiązane* / *jednostki,
  w których jednostka posiada zaangażowanie w kapitale* / *pozostałe*) and trade items by maturity ("do 12
  miesięcy" / "powyżej 12 miesięcy"); the cash-flow statement shows cash "o ograniczonej możliwości
  dysponowania"; the notes ask for "środki pieniężne zgromadzone na rachunku VAT" (notes ust. 1 pkt 18),
  "strukturę rzeczową (rodzaje działalności) i terytorialną (rynki geograficzne) przychodów netto" (ust. 2
  pkt 1), costs by nature even from entities using the by-function RZiS (ust. 2 pkt 2), and the
  reconciliation of the tax base to the accounting result (ust. 2 pkt 6).

### 1.3 Accounting-policy changes

- **Art. 5 ust. 1** — policy is applied "w sposób ciągły, dokonując w kolejnych latach obrotowych
  jednakowego grupowania operacji gospodarczych, wyceny aktywów i pasywów … ustalania wyniku finansowego
  i sporządzania sprawozdań finansowych tak, aby za kolejne lata informacje z nich wynikające były
  porównywalne."
- **Art. 8 ust. 2** — "W celu rzetelnego i jasnego przedstawienia sytuacji jednostka może, ze skutkiem od
  pierwszego dnia roku obrotowego, bez względu na datę podjęcia decyzji, zmienić dotychczas stosowane
  rozwiązania na inne, przewidziane ustawą." The notes must state the reasons, quantify the effect on the
  result and keep the prior year comparable; "Skutki zmiany przyjętych zasad (polityki) rachunkowości
  odnosi się na kapitał (fundusz) własny i wykazuje jako zysk (stratę) z lat ubiegłych."
- **KSR 7** ([Dz. Urz. MF 2019 poz. 2](https://www.gov.pl/attachment/9da792cf-ca83-4ace-9779-17aa810a8548),
  uchwała KSR nr 3/2018):
  - pkt 3.1 — policy comprises "a) zasady klasyfikacji i grupowania transakcji i zdarzeń, metody wyceny
    aktywów i pasywów oraz ustalania wyniku finansowego i prezentacji … do których odnosi się niniejszy
    standard, b) techniczno-organizacyjny sposób prowadzenia ksiąg rachunkowych i ich ochrony, do których
    nie odnosi się standard." The chart's mechanics fall under (b), its classification rules under (a).
  - pkt 3.3 — the policy states the classification of events "do odpowiednich pozycji sprawozdania
    finansowego", valuation methods, the method of measuring the result and the way of presentation, "i
    stosuje je w sposób ciągły"; pkt 3.8 — "Zmianę każdego z wymienionych w pkt 3.3. standardu elementów …
    uznaje się za zmianę przyjętych zasad (polityki) rachunkowości."
  - pkt 3.7 — a voluntary change is applied retrospectively; "Nową zasadę stosuje się od pierwszego dnia
    roku obrotowego. Zależnie od decyzji kierownika jednostki jest to pierwszy dzień bieżącego lub
    następnego roku obrotowego."
  - pkt 3.9 — "Nie stanowią zmian przyjętych zasad (polityki) rachunkowości m. in.: a) przyjęcie
    w odniesieniu do nowych zdarzeń i transakcji zasad (polityki) rachunkowości różniących się od poprzednio
    stosowanych do podobnych, ale nie identycznych zdarzeń i transakcji, b) przyjęcie nowej zasady
    w odniesieniu do zdarzeń, które uprzednio nie występowały lub występowały, ale były nieistotne, a stały
    się istotne, c) zmiany wysokości wartości szacunkowych."
- **Stanowisko KSR w sprawie niektórych zasad prowadzenia ksiąg rachunkowych** (załącznik do uchwały nr 5/10
  z 13 kwietnia 2010; [text](https://www.gov.pl/attachment/ecd3a097-3241-4492-b891-16b811b85aae); published
  in Dz. Urz. MF 2010 nr 6 poz. 26 per
  [podatki.biz](https://www.podatki.biz/artykuly/zmiana-systemu-finansowo-ksiegowego-w-trakcie-roku-obrotowego_23_33155.htm)).
  It is still current guidance: KSR 16 (2025) pt 7.2 refers readers to it
  ([Dz. Urz. MFiG 2025 poz. 29](https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a)).
  - pkt 36 — "Ustawa nie zawiera ograniczeń, co do dokonywania zmian systemu FK w ciągu roku obrotowego"
    but requires testing, reconciliation and continuity; "Kwestie, na które należy zwrócić uwagę przy
    zmianie systemu FK to plan kont i algorytmy (kalkulacje) wbudowane w system FK."
  - **pkt 37** — "Zmiana planu kont na poziomie poszczególnych kont księgi głównej nie pociąga za sobą
    (automatycznie) zmian zasad (polityki) rachunkowości, o ile bez zmian pozostają zasady klasyfikacji
    zdarzeń gospodarczych i agregacji danych księgowych. Konieczne jest jednak zapewnienie, aby zmiany planu
    kont były: a) odpowiednio udokumentowane opisem powiązań pomiędzy dotychczasowym i nowym planem kont
    włączonym do przyjętych zasad (polityki) rachunkowości; b) opis ten został zatwierdzony przez kierownika
    jednostki."
  - **pkt 38** — "Zmiany algorytmów w ciągu roku obrotowego, jako naruszające zasadę ciągłości nie są
    dopuszczalne"; new algorithms must be shown by tests to give the old results (pkt 39 allows correcting
    entries to the old algorithm until year end, or recomputing from the start of the year, by a decision
    in the policy).
  - pkt 41–42 — data may move to a new system as detailed entries, turnovers or balances; with aggregated
    transfer the link between detailed and aggregated entries must be tested, verifiable and approved.

### 1.4 May a company add, rename or change accounts during the financial year?

| Change during the year | Answer | Basis |
|---|---|---|
| Add an **analytic** account (a new customer, a new tax title, a new bank account) | **Yes**, as an update of the documentation | art. 10 ust. 2 ("aktualizuje"); art. 16 ust. 1; practice: [Poradnik Przedsiębiorcy, M. Sądej](https://poradnikprzedsiebiorcy.pl/-zakladowy-plan-kont-a-konta-nieuzywane-przez-jednostke): "Takiej zmiany można dokonać również w trakcie roku obrotowego" |
| Add a **synthetic** account for events that did not occur before (first loan, first production) | **Yes**; a new rule for new or newly material events is not a policy change | KSR 7 pkt 3.9 lit. a–b; art. 10 ust. 2 |
| Split or renumber accounts, classification unchanged | **Yes, not automatically a policy change**, if the old-to-new mapping is documented and approved by the head of the entity | Stanowisko KSR 2010 pkt 37; [Poradnik Przedsiębiorcy, K. Szopa, 29.05.2025](https://poradnikprzedsiebiorcy.pl/-zmiana-zakladowego-planu-kont-jednostki-kiedy-stanowi-zmiane-polityki-rachunkowosci): such a change "może być dokonana w trakcie roku obrotowego", formalised by "aneksu lub … zarządzenia kierownika jednostki" |
| **Rename** an account (same content) | **No explicit rule found.** My reading: a documentation update under art. 10 ust. 2, names in Polish (art. 9), the change recorded so earlier printouts stay explainable | art. 10 ust. 2; Stanowisko pkt 14, 37 — *to verify with an accountant* |
| Change that moves events to **different statement positions**, or changes valuation, the result or presentation (e.g. switching to group 5 and the by-function RZiS, changing the stock method) | **Policy change**: effective from the first day of the (current or next) year whatever the decision date; comparatives restated; disclosed | art. 8 ust. 2; KSR 7 pkt 3.3, 3.7, 3.8; art. 5 ust. 1 |
| Change a **calculation algorithm** (depreciation, stock valuation) mid-year | **Not permitted** mid-year; only from a year start, or by recomputing from the start of the year | Stanowisko 2010 pkt 38–39 |
| Change accounts of a **closed year** | **No** — closing is "nieodwracalnym wyłączeniu możliwości dokonywania zapisów księgowych" | art. 12 ust. 5 |

Practice leans to the year start: "najlepszym momentem na ich wprowadzenie jest początek roku obrotowego"
([Infor, P. Łomińska, 29.01.2014](https://ksiegowosc.infor.pl/raport-dnia/143512,Zakladowy-plan-kont-kiedy-dokonywac-zmian.html)).
Software enforces the same: Comarch ERP Optima keeps the chart per accounting period and lets an analytic
account be added only where "na danym koncie nie ma księgowań"
([Optima, Plan kont](https://pomoc.comarch.pl/optima/pl/2026/dokumentacja/plan-kont/)); wFirma: "Plan kont
zawsze przypisywany jest do danego roku obrotowego", and a new model chart can be adopted only when a new
financial year is opened ([wFirma, Plan kont](https://pomoc.wfirma.pl/-plan-kont);
[wFirma, Plan kont 2.0](https://pomoc.wfirma.pl/-wzorcowy-plan-kont-2-0-w-systemie-wfirma-pl)).

**What this means for Rhea (my reading).** The chart is a versioned fact of a company per financial year:
account events carry an effective date; adding accounts mid-year is an ordinary approved event; renumbering
mid-year needs a mapping document approved by the head of the entity and must keep year-to-date turnovers
continuous (art. 18 ust. 1); any rule version that changes classification, valuation or a calculation takes
effect from the first day of a financial year, never mid-year (art. 8 ust. 2; Stanowisko pkt 38); a closed
year's chart is frozen. Rhea's rule versions with effective dates are exactly the "algorytmy i parametry …
wersji oprogramowania i daty rozpoczęcia jego eksploatacji" of art. 10 ust. 1 pkt 3 lit. c.

---

## 2. Model charts in practice

### 2.1 What exists and what was read

| Source | Who / kind | Numbering | Read? |
|---|---|---|---|
| [GOFIN, *Wzorcowy wykaz kont (pełny)*](https://indeksksiegowan.gofin.pl/firma/wzorcowy-wykaz-kont) (Indeks Księgowań, updated 13.10.2026) | Wydawnictwo Podatkowe GOFIN (Gazeta Podatkowa, Zeszyty Metodyczne Rachunkowości) — the main free public model chart | **two-digit** synthetics (01, 02 … 87) with sub-accounts (07-1, 22-0 …); off-balance accounts as x9 in each group | **Yes, in full** |
| [jak-ksiegowac.pl, *Plan kont 2026*](https://www.jak-ksiegowac.pl/files/Plan-kont.pdf) | Teaching site; each account marked A / P / AP / Pr / K / Kor and B / R / W | three-digit | Yes, in full |
| [Wikibooks, *Przykładowy Zakładowy Plan Kont*](https://pl.wikibooks.org/wiki/Rachunkowo%C5%9B%C4%87/Konta/Plany_kont/Przyk%C5%82adowy_Zak%C5%82adowy_Plan_Kont) | Community textbook | three-digit; still lists 770/771 *Zyski/Straty nadzwyczajne* | Yes |
| [University of Warsaw (WNE) teaching chart, 2010](http://coin.wne.uw.edu.pl/akocia/rach_fi/Zakladowy%20plan%20kont.doc) | Teaching material | three-digit, legacy (*Fundusz założycielski*, 770/771) | Yes |
| Public-sector *plan kont dla jednostek budżetowych* ([Dz.U. 2020 poz. 342](https://eli.gov.pl/eli/DU/2020/342/ogl/pol/pdf), Załącznik nr 3) | Ministry, binding for budget units only | three-digit; gives each account an official description incl. "Ewidencja szczegółowa … powinna zapewnić …" | Yes (list and analytic rules) |
| [wFirma, *Plan kont 2.0*](https://pomoc.wfirma.pl/-wzorcowy-plan-kont-2-0-w-systemie-wfirma-pl) and its [1.0→2.0 comparison](https://pomoc.wfirma.pl/download/wzorcowy-plan-kont-2-0-w-systemie-wfirma-pl.pdf/636385) | Cloud product used by small firms and offices | three-digit with analytic codes | Changes only; the full list is not public |
| Comarch ERP Optima *Wzorcowy plan kont* ([Plan kont](https://pomoc.comarch.pl/optima/pl/2026/dokumentacja/plan-kont/), [exercise](https://pomoc.comarch.pl/optima/pl/2024_5/?p=4705)) | Market-leading desktop ERP for offices | three-digit; dictionary accounts | Mechanics and exercise accounts only; the generated list is not public — *to verify* |
| Symfonia, enova365, Rewizor (GT / nexo) | Vendors | — | Each generates a model chart (university course syllabi; for Rewizor nexo a [retailer listing](https://www.empik.com/insert-rewizor-nexo-1stan-insert,p1624244911,multimedia-p), seen only as a search snippet: "możliwość generowania wzorcowego planu kont … Konto może posiadać stałą lub zmienną długość numerów analityki"); **contents not public — *to verify*** |
| Wolters Kluwer LEX *Wzorcowy plan kont* ([product page](https://wolterskluwer.com/pl-pl/solutions/lex/ksiegowosc/baza-prawa/wzorcowy-plan-kont)); ODDK *Plan kont z komentarzem* (J. Gierusz) | Commercial commentaries; per search snippets, descriptions of each account with typical Wn/Ma entries and links to statement positions | — | Paywalled; WK page returned 403; the ODDK title is not confirmed; **not reviewed — *to verify*** |
| SKwP (Stowarzyszenie Księgowych w Polsce) | — | — | **No public SKwP model chart found** (*to verify*). SKwP's 2025 opinion to MF on JPK tags is cited in [filings.md §1.6](filings.md) |

The same nine groups appear everywhere (GOFIN's names): *Zespół 0 – Aktywa trwałe; 1 – Środki pieniężne,
rachunki bankowe oraz inne krótkoterminowe aktywa finansowe; 2 – Rozrachunki i roszczenia; 3 – Materiały
i towary; 4 – Koszty według rodzajów i ich rozliczenie; 5 – Koszty według typów działalności i ich
rozliczenie; 6 – Produkty i rozliczenia międzyokresowe; 7 – Przychody i koszty związane z ich osiągnięciem;
8 – Kapitały (fundusze) własne, fundusze specjalne i wynik finansowy.* Group 9 is used for off-balance
accounts (wFirma 2.0; [Infor, 2006](https://ksiegowosc.infor.pl/porady-prawne/29724,Jak-zaksiegowac-koszty-niebedace-kosztami-uzyskania-przychodow.html):
990/991/999) or, in the public sector, for budget execution; GOFIN instead keeps off-balance accounts as
09, 29, 39.

### 2.2 Group by group — the synthetic accounts in the published charts

GOFIN's two-digit accounts are shown as published (its "40-0" corresponds to a three-digit "400").
"JB" = the public-sector chart. "—" = not in that chart.

**Zespół 0 — Aktywa trwałe**

| Content | GOFIN | jak-ksiegowac | Wikibooks | JB (MF) | Rhea |
|---|---|---|---|---|---|
| Środki trwałe | 01 | 010 | 010 | 011 (013 pozostałe) | 010 |
| Wartości niematerialne i prawne | 02 | 020 | 020 | 020 | 020 |
| Długoterminowe aktywa finansowe | 03 (03-5 odpisy) | 030 (079 odpisy) | 031, 032 (035 odpisy) | 030 (073 odpisy) | 030 *Udzielone pożyczki* |
| Inwestycje w nieruchomości i prawa | 04 | 040 (077 odpisy) | — | — | — |
| Umorzenie ŚT / WNiP | 07-1 / 07-2 (07-3 inwestycje) | 070 / 075 | 071 / 075 | 071, 072 | 071 / 075 |
| Odpisy z tytułu trwałej utraty wartości | 07-4…07-6 | 078 | 079 | — | — |
| Środki trwałe w budowie | 08 | 080 | 080 (081 zaliczki) | 080 | 080 |
| Off-balance (obce ŚT, ŚT w likwidacji) | 09-0, 09-1 | 091–093 | 091, 092 | — | — |

**Zespół 1 — Środki pieniężne …**

| Content | GOFIN | jak-ksiegowac | Wikibooks | JB | Rhea |
|---|---|---|---|---|---|
| Kasa | 10 | 100 | 101 krajowa, 102 walutowa | 101 | 100 |
| Rachunek bieżący | 13-0 | 130 | 130, 131 | 130 | 130 |
| Rachunek walutowy | 13-1 | — | — | — | — |
| Kredyty bankowe | 13-4 | 138 długo-, 139 krótkoterminowe | 134 | 134 | — |
| Inne rachunki bankowe | 13-5 | 135 inne środki pieniężne | 132 | 139 | — |
| Środki pieniężne w drodze | 13-9 | 145 | 139 | 141 | **149** |
| Krótkoterminowe aktywa finansowe | 14 (14-5 odpisy) | 140 (141 odpisy), 144 weksle obce | 141, 145 inne aktywa pieniężne | 140 | — |
| Rachunek VAT | — | — | — | — | — (wFirma: 132-[konto VAT]) |

**Zespół 2 — Rozrachunki i roszczenia**

| Content | GOFIN | jak-ksiegowac | Wikibooks | JB | Rhea |
|---|---|---|---|---|---|
| Odbiorcy / dostawcy | 20 / 21 | 200 / 201 | 200 / 210 (205 zaliczki) | 201 (one account) | 201 / 202 |
| VAT: rozrachunek z US / naliczony / należny | 22-0 / 22-1 / 22-2 | 221 (221-1 należny, 221-2 naliczony), 223 do rozliczenia | 220 / 221 / 222 (226, 227 następny miesiąc) | — | — / 221 / 222 |
| Pozostałe rozrachunki publicznoprawne | 22-3 | 220 (220-1 ZUS, 220-2 US), 225 | 223 ZUS, 224 CIT, 225 PIT-4 | 225 budżety, 229 pozostałe | 225 (taxes), 229 ZUS |
| Wynagrodzenia / pozostałe z pracownikami | 23-0 / 23-4 | 230 / 234 | 231 / 234 | 231 / 234 | 231 / — |
| Pożyczki | 24-0 | — | 242 | — | 240 otrzymane |
| Niedobory, szkody, nadwyżki | 24-1 | 242, 243 (231 należności z tyt. niedoborów) | 246, 247 | — | — |
| Właściciele spółek osobowych / udziałowcy | 24-2 / 24-3 | — | 248 | — | 241 *dywidendy* |
| Wewnątrzzakładowe | 24-5 | — | 250 | — | — |
| Należności dochodzone sądownie | 24-6 | 244 | 241 sporne | — | — |
| Pozostałe | 24-9 | 240 | 249 | 240 (245 wpływy do wyjaśnienia) | 249 |
| Należne wpłaty na kapitał | — | 250 | — | — | — |
| Odpisy aktualizujące należności | 28 | 290 | 260 | 290 | 280 |
| Off-balance (warunkowe, weksle) | 29-0…29-2 | 291 | 291–294 | — | — |

**Zespół 3 — Materiały i towary**

| Content | GOFIN | jak-ksiegowac | Wikibooks | JB | Rhea |
|---|---|---|---|---|---|
| Rozliczenie zakupu | 30 | 300–303 (ŚT, materiały, usługi, towary) | 301–304 | 300 | 300 |
| Materiały (opakowania, w przerobie) | 31 (31-1, 31-6, 31-9) | 310 | 311, 319 | 310 | 310 |
| Towary | 33 | 330 hurt, 335 detal | 331–333 | 330 | 330 |
| Odchylenia od cen ewidencyjnych | 34-1…34-6 | 341–344 | 320, 334, 335, 345 | 340 | — |
| Odpisy aktualizujące zapasy | 34-7 | 345 | 336, 337 | — | 349 |
| Zapasy obce (off-balance) | 39-0 | 390 | 391 | — | — |

**Zespół 4 — Koszty według rodzajów** — the clearest consensus:

| Content | GOFIN | jak-ksiegowac | Wikibooks | JB (MF) | Rhea |
|---|---|---|---|---|---|
| Amortyzacja | 40-0 | 400 | 400 | 400 | **401** |
| Zużycie materiałów i energii | 40-1 | 401 | 401 | 401 | **402** |
| Usługi obce | 40-2 | 402 | 402 | 402 | **403** |
| Podatki i opłaty | 40-3 | 403 | 403 | 403 | **missing** |
| Wynagrodzenia | 40-4 | 404 | 404 | 404 | 404 |
| Ubezpieczenia społeczne i inne świadczenia | 40-5 | 405 | 405 | 405 | 405 |
| Pozostałe koszty rodzajowe | 40-9 | 409 (406 podróże) | 409 (407 reprezentacja i reklama) | 409 | 409 |
| Rozliczenie kosztów | 49 | 490 | 490 | 490 | 490 |

**Zespół 5 — Koszty według typów działalności** (kept by producers and by-function reporters)

| Content | GOFIN | jak-ksiegowac | Wikibooks | JB |
|---|---|---|---|---|
| Działalność podstawowa — produkcyjna (usługowa) | 50 | 500 | 501 | 500 |
| Działalność podstawowa — handlowa | 51 | 528 koszty handlowe | 521 | — |
| Koszty wydziałowe | 52-1 | 520 | 510 | — |
| Koszty zakupu | — | 522 | 511 | — |
| Koszty sprzedaży | 52-7 | 527 | 515 | — |
| Działalność pomocnicza | 53 | 530 | 531 | 530 |
| Koszty zarządu | 55 | 550 | 551 | 550 |
| Rozliczenie kosztów działalności | 58 | 580 | 580 | 580 |

**Zespół 6 — Produkty i rozliczenia międzyokresowe**

| Content | GOFIN | jak-ksiegowac | Wikibooks | JB | Rhea |
|---|---|---|---|---|---|
| Produkty gotowe / półprodukty / produkty w toku | 60-0 / 60-1 / 60-2 | 601 / 602 / 603 | 600 / 605 / 630 | 600 | — |
| Odchylenia, odpisy na produkty | 62-0, 62-1 | 621–623 | 620 | 620 | — |
| RMK czynne / bierne | 64-0 / 64-1 | 640, 641 / 642 | 641, 642 | 640 (both) | 640 / 641 |
| Aktywa z tytułu odroczonego podatku | 65-0 | — | — | — | 650 |
| Inne rozliczenia międzyokresowe | 65-1 | — | 642 | — | — |

**Zespół 7 — Przychody i koszty związane z ich osiągnięciem**

| Content | GOFIN | jak-ksiegowac | Wikibooks | JB | Rhea |
|---|---|---|---|---|---|
| Sprzedaż produktów / koszt sprzedanych produktów | 70-0 / 70-1 | 701, 702 / 711, 712 | 700, 703 / 711 | 700 (both) | 700 / — |
| Sprzedaż towarów / wartość sprzedanych towarów | 73-0 / 73-1 | 730 / 731 | 731, 732 / 735 | 730 | 730 / 731 |
| Sprzedaż materiałów / wartość sprzedanych materiałów | — | 740 / 741 | 740 / 745 | — | (Rhea: 760/761) |
| Przychody / koszty finansowe | 75-0 / 75-1 | 750 / 751 | 750 / 751 | 750 / 751 | 750 / 751 (+751-1) |
| Pozostałe przychody / koszty operacyjne | 76-0 / 76-1 | 760 / 761 | 760 / 761 | 760 / 761 | 760 / 761 (+761-1…3) |
| Zyski / straty nadzwyczajne | — | — | 770 / 771 | — | — |
| Obroty wewnętrzne / koszt obrotów wewnętrznych | 79-0 / 79-1 | — | 790 / 791 | — | — |
| Różnice kursowe i zaokrąglenia | — | — | — | — | **756** |

**Zespół 8 — Kapitały, fundusze specjalne, rezerwy, wynik**

| Content | GOFIN | jak-ksiegowac | Wikibooks | JB | Rhea |
|---|---|---|---|---|---|
| Kapitał podstawowy | 80 | 801 | 800 | 800 fundusz jednostki | 801 |
| Kapitał zapasowy / rezerwowy / z aktualizacji wyceny | 81-1 / 81-2 / 81-3 | 802 / 803 / 804 | 801 / 802 / 803 | — | 802 / — / 813 |
| Rozliczenie wyniku finansowego | 82 | 820 | 820 | 820 | 820 |
| Rezerwy: odroczony podatek / pozostałe | 83-0 / 83-1 | 840 | 830 | 840 (with RMP) | 835 / 830 |
| Rozliczenia międzyokresowe przychodów | 84 | 845 | 840 | 840 | **missing** |
| ZFŚS / inne fundusze specjalne | 85-0 / 85-1 | 851 / 852 | 851 / 852 | 851, 853, 855 | — |
| Wynik finansowy | 86 | 860 | 860 | 860 | 860 |
| Podatek dochodowy / inne obciążenia wyniku | 87 | 870 / 871 | 871 / 872 | 870 | 870 |

**Obsolete items still found in published charts.** *Zyski / straty nadzwyczajne* (770/771) have no RZiS
line in Annex 1 today, and wFirma removed them in Plan kont 2.0. *Sprzedaż materiałów* (740/741) is
presented, since KSR 15's 2026 update, under other operating income and costs: "Przychody ze sprzedaży
materiałów prezentowane są … w wariancie porównawczym - w pozycji D. IV. Inne przychody operacyjne", and
their carrying value under E.III ([KSR 15, Dz. Urz. MFiG 2026 poz. 53, pkt 8.1](https://www.gov.pl/attachment/143e3deb-6668-4506-b191-b0efcce109f7));
Rhea's rule already books such sales to 760. *Fundusz założycielski / przedsiębiorstwa* (UW 2010) are
state-enterprise legacy.

### 2.3 The analytical structure used in practice

- **Counterparties — per counterparty, per currency, per maturity, per document.** Comarch's model
  exercise uses 201-2-1 "do 12 miesięcy" with customer analytics 201-2-1-ALOZA, a separate synthetic 203
  *Rozrachunki walutowe* with 203-AURELIA in EUR, and 202-2-1-BIUROWIEC for suppliers
  ([Optima exercise](https://pomoc.comarch.pl/optima/pl/2024_5/?p=4705)). Dictionary-linked (*słownikowe*)
  accounts generate analytics from master data — "Kontrahenci, Artykuły, Pracownicy, Magazyny, Urzędy,
  Banki, Stawki VAT, Środki trwałe" ([Comarch Altum, Konta słownikowe](https://pomoc.comarch.pl/altum/documentation/konta-slownikowe/));
  currency becomes its own level (202-COMARCH_EUR), and once 201-X-EUR exists, posting to 201-X is blocked.
  wFirma's *konta kartotekowe* draw on kontrahent, pracownik, udziałowiec, pojazd, konto bankowe, konto VAT,
  system płatności, środek trwały i WNiP, kasa and magazyn, with an analytic filter by KŚT group
  ([wFirma, Plan kont](https://pomoc.wfirma.pl/-plan-kont)). The public-sector chart states the rule in the
  register's own words: for 201, "Ewidencja szczegółowa … powinna zapewnić możliwość ustalenia należności
  i zobowiązań krajowych i zagranicznych według poszczególnych kontrahentów"; for group 2 generally, by
  counterparty "oraz, jeśli dotyczą rozliczeń w walutach obcych – według poszczególnych walut"
  ([Dz.U. 2020 poz. 342](https://eli.gov.pl/eli/DU/2020/342/ogl/pol/pdf)).
- **Public-law settlements — per tax title and authority.** wFirma 2.0 splits 223 into analytics such as
  223-06 PIT (per shareholder), 223-07 *CIT estoński*, 223-08 *Podatek od nieruchomości*, 223-09 *PCC*,
  223-10 *Rozrachunki celne*, 223-11 *Pozostałe podatki*; Comarch's exercise splits ZUS into 227-51
  *ubezpieczenie społeczne*, 227-52 *zdrowotne*, 227-53 *Fundusz Pracy i FGŚP*; the public-sector chart:
  225 "według każdego z tytułów rozrachunków z budżetem odrębnie", 229 "według tytułów rozrachunków oraz
  podmiotów". VAT: *naliczony* split into current period and later periods (Wikibooks 226/227 for "w
  następnym miesiącu"; GOFIN 22-1 *VAT naliczony i jego rozliczenie*).
- **Employees — per person and per title**, which art. 17 ust. 1 pkt 3 requires for wages; 234 per title
  (advances, business trips, loans from ZFŚS) per JB 234.
- **Loans — by counterparty type.** wFirma 2.0 added 246 *Pożyczki otrzymane* (246-01…03: udziałowcy,
  kontrahenci, pozostałe) and 247 *Pożyczki udzielone* (247-01…04: udziałowcy, kontrahenci, pracownicy,
  pozostałe).
- **Costs by nature with KUP / NKUP.** Two practices coexist:
  1. NKUP analytics under group 4: "koszty NKUP ujmuje się na kontach analitycznych prowadzonych do kont 40-0,
     40-1, 40-2, 40-3, 40-4, 40-5, 40-6" ([Gazeta Podatkowa, D. Przybyszewska, 27.08.2024, via pit.pl](https://www.pit.pl/?p=11769));
     wFirma 1.0 had 404-01-01 *umowa o pracę KUP* / 404-01-02 *NKUP*.
  2. An off-balance tax register: GOFIN's article uses off-balance 91 *Koszty ujęte w księgach bieżącego roku
     – NKUP* and 93 for non-taxable revenue (same source); Infor (2006) uses 990 *Wyłączenia z przychodów*,
     991 *Wyłączenia z kosztów*, 999 *Rozliczenie kont pozabilansowych*; **wFirma 2.0 moved the whole tax
     view to group 9** ("Ujęcie podatkowe przeniesiono na konta pozabilansowe zespołu 9"), dropping the
     KUP/NKUP analytics in 404 and renaming 750/751/760/761 analytics from "NKUP/NP" to statement-driven
     ones (*odsetki*, *dodatnie/ujemne niezrealizowane bilansowe różnice kursowe*, *przychody/koszt sprzedaży
     niefinansowych aktywów trwałych*, *dotacje*, *darowizny*), with 940–982 *Konta ewidencji różnic
     pomiędzy wynikiem bilansowym a podatkowym*.
  JPK_KR_PD's rules on PD tags (one per analytic account; off-balance difference accounts tagged `…_PB`)
  shape this choice — [filings.md §1.5](filings.md).
- **Revenue — by kind and market**, for the notes' "struktura rzeczowa … i terytorialna"; JB 700: "podział
  sprzedaży dostosowany do zasad wymiaru podatków … oraz według przedmiotów i kierunków sprzedaży". Split by
  VAT rate on 7xx is not required by the UoR (VAT registers carry rates); *to verify* whether it is common.
- **Fixed assets — per balance-sheet class / KŚT group, then per object.** UW 2010: "0111 - Grunty 0112 -
  Budynki i budowle"; wFirma filters 0xx analytics by KŚT; JB 011 analytics must allow "ustalenie wartości
  początkowej poszczególnych obiektów … należyte obliczenie umorzenia i amortyzacji".
- **Equity — by kind, and for Estonian CIT by period.** wFirma 2.0 added 802-01…802-04: profits and
  uncovered losses "w latach nieobjętych" and "w trakcie opodatkowania ryczałtem od dochodów spółek".
- **Cost centres (MPK) and projects** are a common analytic of group 5 and sometimes group 4 in practice
  (JB 500/550 analytics by activity and structure); in Rhea's terms these are T-codes, not account segments
  (DIRECTION, D3).

### 2.4 The cost-recording variants

Three set-ups, described in [Infor / Gazeta Prawna, A. Pokojska, 11.04.2007](https://ksiegowosc.infor.pl/poradniki/1185,2,Jak-ujmowac-koszty-w-rachunkowosci.html):

1. **Group 4 only** — "pozostaje wariant porównawczy rachunku zysków i strat". Group 5 absent; 490 still
   needed for *zmiana stanu produktów*, *koszt wytworzenia produktów na własne potrzeby* and transfers to
   RMK (Rhea uses it this way).
2. **Group 5 only** — such entities "będą tworzyły rachunek zysków i strat w wariancie kalkulacyjnym". The
   notes still require costs by nature from by-function reporters (Annex 1, notes ust. 2 pkt 2), so in
   practice the group-5 accounts carry cost-by-nature analytics (my reading).
3. **Groups 4 and 5 together, linked by 490** — either RZiS variant; the author notes that in practice only
   these entities have the choice. 490 "Dzięki niemu możliwe jest … przeniesienie i rozliczenie kosztów
   zaewidencjonowanych wcześniej na kontach zespołu 4".

The choice of variant is the head of the entity's (art. 47 ust. 4 pkt 1 and 5) and belongs in the policy
(art. 10 ust. 1 pkt 2). JPK_KR_PD tags every account in either variant, and its P&L tags name the position
in both variants at once (e.g. `RPAI_RKAI_POW`) — [filings.md §1.5–1.6](filings.md).

### 2.5 Simplifications small and micro entities use, and what they do to the chart

| Simplification (who) | Effect on the chart | Basis |
|---|---|---|
| Expense stock on purchase, count at year end (all, by the head's decision) | 310/330 used only at the balance date; no 300 matching | art. 17 ust. 2 pkt 4; KSR 16 pt 11.4 |
| Quantity-only or value-only subsidiary ledgers for stock | fewer stock analytics; value-only for retail | art. 17 ust. 2 pkt 2–3 |
| Partial journals; subsidiary-ledger trial balance only at the balance date; summary documents; codes in entries | ledger mechanics, not accounts | KSR 16 pt 7.1 lit. a–g |
| No deferred tax (micro, small) | no 650 / 835 | art. 37 ust. 10 |
| Leases classified by tax rules (micro, small) | operating-lease assets off balance (group 9) | art. 3 ust. 6 |
| Tax depreciation in the books (micro, small — **not** spółki kapitałowe, SKA, certain partnerships) | no book/tax depreciation difference | art. 32 ust. 7–8 |
| No accruals for future employee benefits (micro, small — **not** spółki kapitałowe) | fewer 641 analytics | art. 39 ust. 6–7 |
| No fair-value measurement (micro under art. 28a; micro/small option under 28b) | fewer revaluation accounts (813 rarely used) | art. 28a, 28b |
| Simplified statements (Annex 4 micro, Annex 5 small) | fewer statement positions; **but** JPK_KR_PD still uses the full Annex-7 tag dictionary for them (FAQ Q94), so the analytics the simplification spared come back | art. 47 ust. 4 pkt 4–5; [filings.md §1.6](filings.md) |

For a **sp. z o.o.** three simplifications are closed by statute — tax depreciation (art. 32 ust. 8 pkt 2),
skipping employee-benefit accruals (art. 39 ust. 7 pkt 2) and waiving prudence for write-downs and
provisions (art. 7 ust. 2b–2c: "Przepisu ust. 2b nie stosuje się do: … 2) spółek kapitałowych"); deferred
tax and tax-based lease classification remain open to small and micro companies.

### 2.6 What JPK_KR_PD does to the analytic structure (brief; the filings strand is authoritative)

The Annex-7 dictionary goes to the **lowest statement level** and the FAQ requires **homogeneous accounts**,
one tag per account ([filings.md §1.5](filings.md)). So the chart must split at account level: gross value /
umorzenie / odpisy aktualizujące (`_W` / `_U` / `_A`); trade receivables and payables by relation and by
maturity (`BABII1a_D12`, `BABII1a_P12`, … `BPBIII3d_P12`); revenue and several costs into related / other
(`_POW` / `_POZ`); *podatki i opłaty* with excise apart (`RPBIV_AKC`); social security "w tym emerytalne"
(`RPBVI_E`); provisions and deferred income long / short (`_D` / `_K`); income tax current / deferred
(`RPJ_RKM_B` / `_O`); equity for Estonian CIT (`…_ES_…`). I decoded these codes independently from the
regulation's PDF and they match the filings strand's list.

---

## 3. Rhea's current chart — gap analysis

Read from `/home/krzysztof/Rhea/packs/pl/pack.json` (pack `pl` v44, read-only): `account.created` events
for 55 accounts and `account.tax_treatment.set` for 6. The `account` object type (v2) has `code`, `name`,
`type` ∈ {asset, liability, equity, revenue, expense, debtor, creditor, tax, bank, clearing, income_tax}
and an amendable `tax_treatment` ∈ {deductible, non_deductible, taxable, not_taxable}. Year-end closing
rules (`pl-close-year-*`) close accounts **by `type`** to 860. No view maps accounts to the balance sheet or
RZiS (REVIEW FB-07 noted the missing P&L view).

### 3.1 The 55 accounts, against the published convention

| Group | Rhea accounts (code name, type) | Against practice |
|---|---|---|
| 0 | 010 Środki trwałe (asset); 020 WNiP (asset); 030 *Udzielone pożyczki* (asset); 071 Umorzenie ŚT (asset); 075 Umorzenie WNiP (asset); 080 ŚT w budowie (asset) | Numbers match Wikibooks. 030 is *Długoterminowe aktywa finansowe* everywhere; Rhea puts all loans granted there, short-term ones included. 071/075 typed `asset` though contra (credit balance) |
| 1 | 100 Kasa (bank); 130 Rachunek bieżący (bank); 149 Środki pieniężne w drodze (bank) | 149 is used for this by none of the charts read (GOFIN 13-9, Wikibooks 139, jak-ksiegowac 145, JB 141) |
| 2 | 201 odbiorcy (debtor); 202 dostawcy (creditor); 221 VAT naliczony (tax); 222 VAT należny (tax); 225 Pozostałe rozrachunki publicznoprawne (tax); 229 ZUS (creditor); 231 wynagrodzenia (creditor); 240 Pożyczki otrzymane (creditor); 241 *Rozrachunki z właścicielami z tytułu dywidend* (creditor); 249 Pozostałe rozrachunki (debtor); 280 Odpisy aktualizujące należności (asset) | 201/202 as Optima; 221/222 as GOFIN/Wikibooks. 225 carries CIT advances, dividend withholding tax and tax payments together. 241 is *niedobory* in GOFIN and *sporne* in Wikibooks; dividends belong under *rozrachunki z udziałowcami* (GOFIN 24-3). 280 typed `asset` though contra |
| 3 | 300 Rozliczenie zakupu (clearing); 310 Materiały; 330 Towary; 349 Odpisy aktualizujące zapasy (asset) | Fine; 349 typed `asset` though contra |
| 4 | 401 Amortyzacja; 402 Zużycie materiałów i energii; 403 Usługi obce; 404 Wynagrodzenia; 405 Ubezpieczenia społeczne i inne świadczenia; 409 Pozostałe koszty rodzajowe; 409-1 Koszty reprezentacji (NKUP); 490 Rozliczenie kosztów (clearing) | **401–403 shifted one place** from every chart checked (400 Amortyzacja, 401 Zużycie, 402 Usługi obce, 403 Podatki i opłaty); **no *Podatki i opłaty*** — an RZiS line in the comparative variant (B.IV "w tym: podatek akcyzowy"). Rhea books bank fees to 403 *Usługi obce* (`pl-post-explained-bank-operation`), which a practitioner reading "403" will take for *Podatki i opłaty*. 404/405 unused by any rule (no payroll yet) |
| 6 | 640 Czynne RMK (asset); 641 Bierne RMK (creditor); 650 Aktywa z tytułu odroczonego podatku (asset) | As GOFIN 64-0/64-1/65-0. No long/short split (balance sheet A.V.2 vs B.IV) |
| 7 | 700 Sprzedaż produktów; 730 Sprzedaż towarów; 731 Wartość sprzedanych towarów (expense); 750 Przychody finansowe; 751 Koszty finansowe; 751-1 Odsetki budżetowe (NKUP); 756 *Różnice kursowe i zaokrąglenia* (expense); 760 Pozostałe przychody operacyjne; 761 Pozostałe koszty operacyjne; 761-1 Kary i grzywny (NKUP); 761-2 Darowizny (NKUP); 761-3 Odpisy aktualizujące należności (NKUP) | 756 is the rounding account of FX conversion in the posting rules (`rounding_account: "756"`), typed `expense` but two-sided; no published chart has it. FX differences belong to financial income/costs (art. 30 ust. 4); where conversion roundings belong is *to verify* with an accountant |
| 8 | 801 Kapitał podstawowy; 802 Kapitał zapasowy; 813 Kapitał z aktualizacji wyceny; 820 Rozliczenie wyniku; 830 Rezerwy na zobowiązania (creditor); 835 Rezerwa z tytułu odroczonego podatku (creditor); 860 Wynik finansowy; 870 Podatek dochodowy (income_tax) | Mixed conventions (801/802 like jak-ksiegowac; 813 like GOFIN 81-3) — harmless. 870 receives both current CIT and deferred tax (`pl-post-deferred-tax-*`); RZiS and JPK need them apart (art. 37 ust. 8) |
| 9 | — | No off-balance accounts at all |

REVIEW FB-05 (2026-10-16) judged "the numbering is sound"; this comparison qualifies that for group 4.

### 3.2 Synthetic accounts a complete chart has that Rhea lacks

**Needed by every full-books company (sp. z o.o. baseline):**

| Missing | Why | Basis |
|---|---|---|
| 403 Podatki i opłaty | RZiS comparative B.IV "w tym: podatek akcyzowy"; property tax, PFRON, fees, PCC | Annex 1; all charts read |
| 840 Rozliczenia międzyokresowe przychodów | Balance sheet B.IV.2 (long/short); grants for fixed assets, income received in advance for later periods | Annex 1; GOFIN 84 |
| 223 Rozrachunki z US z tytułu VAT | VAT payable / excess carried / refund — kept apart from other taxes (art. 7 ust. 3); Rhea's `pl-settle-vat` computes the settlement but posts nothing, so 221/222 are never closed against it, and tax payments land on 225 | GOFIN 22-0; Wikibooks 220 |
| 224 Rozrachunki z tytułu podatku dochodowego (CIT) | CIT advances and the annual liability, apart from PIT-4 and withholding taxes now on 225 | Wikibooks 224; REVIEW FB-05 |
| 234 Pozostałe rozrachunki z pracownikami | advances, business trips — art. 17 ust. 1 pkt 3 | all charts |
| 243 Rozrachunki z udziałowcami | dividends, interim dividends, *dopłaty* and their return, per shareholder (Rhea has dividends only, on 241) | GOFIN 24-3 |
| 250 Należne wpłaty na kapitał podstawowy | Balance sheet C — every new company until capital is paid (REVIEW FB-08) | Annex 1; jak-ksiegowac 250 |
| 081 Zaliczki na środki trwałe w budowie i WNiP | Balance sheet A.II.3 and A.I.4; Rhea books asset advances to 249 (`pl-post-asset-advance`), i.e. among receivables | Annex 1; Wikibooks 081 |
| 205 Rozrachunki z tytułu zaliczek na dostawy i usługi | advances paid go to *Zapasy* B.I.5, advances received to B.III.3.e — not to trade receivables/payables | Annex 1; Wikibooks 205 |
| 132 Rachunek VAT | notes ust. 1 pkt 18; cash of restricted use | Annex 1; wFirma 132 |
| 9xx off-balance: *zobowiązania warunkowe*, *obce środki trwałe* (operating leases), book-to-tax register | disclosure in notes; JPK_KR_PD off-balance tax accounts | GOFIN 09/29/39; wFirma 9xx; JPK reg. § 2 ust. 1 pkt 5 |

**Needed by some companies:** 134 *Kredyty bankowe*; 131 *Rachunki walutowe*; 135 *Inne rachunki bankowe*
(ZFŚS, deposits, escrow); 140/145 short-term financial assets and write-downs; 035 write-downs of long-term
financial assets; 040/077 investment property; 072/076 impairment of fixed assets and intangibles;
241 *Rozliczenie niedoborów, szkód i nadwyżek*; 242 partners of partnerships; 245 branches; 246 *Należności
dochodzone na drodze sądowej* (balance sheet B.II.3.d); 247 payment operators (e-commerce); 340 price
variances; group 5 (500–580) for producers and by-function reporters; 600–621 products; 701 *Koszt
sprzedanych produktów*; 790/791 internal turnover; 803 *Pozostałe kapitały rezerwowe*; 851/852 special
funds; 871 *Pozostałe obowiązkowe zmniejszenia zysku*.

### 3.3 Missing analytics

| Account(s) | Analytic levels a complete chart has | Rhea today |
|---|---|---|
| 010 / 071 | balance-sheet class (a grunty … e inne) → object; 071 mirrors 010 | none (assets exist as `fixed_asset` objects, not as ledger analytics) |
| 020 / 075 | prace rozwojowe / wartość firmy / inne → item | none |
| 030 | relation → kind (udziały, papiery, pożyczki, inne) → item | none |
| 080 | investment task → kind of outlay (incl. capitalised interest and FX, notes ust. 2 pkt 7) | none |
| 100 / 130 | cash desk / bank account → currency | none |
| 201 / 202 / 280 | relation → maturity (do / powyżej 12 m-cy) → currency → counterparty (→ document) | none; one 280 for all write-downs |
| 221 / 222 | period / deductible now or later / mechanism (WNT, import usług, odwrotne obciążenie) | none |
| 225 / 229 | tax title and authority / ZUS fund (social, health, FP+FS, FGŚP, PPK, PFRON) | none |
| 231 / 234 | title → person | none |
| 240 / 030 | received or granted → relation (shareholders, related, other) → long / short part → contract | none |
| 4xx | cost kind sub-lines (excise in 403; "w tym emerytalne" in 405) and KUP / NKUP | NKUP only for 409-1 |
| 640 / 641 / 830 / 840 | title → long / short term | none |
| 700 / 730 | products or services → relation → market (kraj / UE / poza UE) → kind of activity | none |
| 750 / 751 / 760 / 761 | the RZiS sub-lines (dividends, interest related / other, gain or loss on disposal, revaluation, FX, other) and tax status | NKUP analytics only (751-1, 761-1…3) |
| 802 / 820 | agio / from profit / Estonian-CIT periods; prior-year results by year / interim dividends / error and policy corrections | none |
| 870 | current / deferred | none |

### 3.4 Attributes the account will need

The pack has `type` and `tax_treatment`. A complete chart needs (sources in brackets; JPK detail in
[filings.md §7](filings.md)):

| Attribute | Why | Rhea |
|---|---|---|
| **Nature**: aktywne / pasywne / aktywno-pasywne (two balances) / kosztowe / przychodowe / korygujące / rozliczeniowe / pozabilansowe | the closing logic and statement sign depend on it (jak-ksiegowac legend A, P, AP, Pr, K, Kor; wFirma: bilansowe, rozliczeniowe, wynikowe, pozabilansowe) | `type` mixes nature with role; contra accounts typed `asset`; two-sided 221/222/225/756 have one type |
| **Level, parent, segment definition** | "powiązania z kontami księgi głównej" (art. 10 ust. 1 pkt 3 lit. a); JPK `S_3` parent | flat codes with "-1" suffixes |
| **Dictionary link** (kontrahent, pracownik, udziałowiec, rachunek bankowy, rachunek VAT, kasa, magazyn, środek trwały, pojazd, operator płatności, urząd / tytuł, umowa pożyczki, zadanie inwestycyjne) | art. 17 ust. 1; Comarch / wFirma practice | none |
| **Currency** of the analytic | art. 23 ust. 3; Optima | on postings only |
| **Open-item (rozrachunkowe) flag** | art. 17 ust. 1 pkt 2; matching | `settlement` objects exist, not on the account |
| **Statement position** — balance sheet (Annex 1, 4, 5) per debit / credit balance; RZiS comparative **and** by-function; value role (gross / umorzenie / odpis) | Annex 1; art. 7 ust. 3 | none |
| **JPK_KR_PD tags** `S_12_1`, `S_12_2`, `S_12_3`, per year | Dz.U. 2024 poz. 1314 § 2 — [filings.md](filings.md) | none |
| **Cash-flow treatment** — operating / investing / financing; indirect-method change line; restricted cash (rachunek VAT, ZFŚS, escrow) | Annex 1 RPP (G "o ograniczonej możliwości dysponowania"); notes ust. 1 pkt 18 | none |
| **Income tax**: KUP / NKUP (permanent, temporary), revenue taxable / not, category a–h of § 2 ust. 1 pkt 5; R&D (art. 9 ust. 1b); IP Box (art. 24e) | CIT art. 9 ust. 1 | `tax_treatment` on 6 accounts |
| **Relation** (powiązane / zaangażowanie / pozostałe) and **maturity / term** (≤ 12 m / > 12 m; long / short) | Annex 1; JPK `_POW`/`_POZ`, `_D12`/`_P12`, `_D`/`_K` | none |
| **Validity** (from / to, financial year), active / inactive, name history | art. 10 ust. 2; Stanowisko pkt 37; Optima status, wFirma per-year chart | `account.created` only |
| **Required T-codes** (MPK, project) | practice (*to verify*); KK: T-codes are not accounts | none |
| **Role** (what the account is for, so rules book to roles) | DIRECTION D3 (proposed, not decided) | none |

---

## 4. Proposed skeleton — *research proposal for review by a practising accountant*

**Numbering principle.** Keep Rhea's existing code wherever at least one published chart uses it for the
same content; otherwise use the number most charts share (§2.2). Numbers are customary — the law leaves
them to the entity (art. 10 ust. 1 pkt 3 lit. a) — so renumbering is a choice for KK and the accountant,
and free under the role design if D3 is adopted.

**Scope:** **W** every full-books company (the sp. z o.o. baseline; "W (VAT)" if a VAT payer); **Z**
employer or payer of civil-law contracts; **H** trade with stock; **P** production (own products, group 5/6);
**K** by-function RZiS; **B** long-term contracts (KSR 3); **$** foreign currency; **G** related parties /
groups; **F** financial investments; **S** partnerships; **E** Estonian CIT; **X** other specific case;
**M** micro / small may omit.

Analytic levels read "1:" (first level) "→ 2:" (second) and so on; the last level of settlement accounts
is the counterparty or person, and below it the open document.

### Zespół 0 — Aktywa trwałe

| Konto | Nazwa | Analityka | Zakres | Rhea dziś |
|---|---|---|---|---|
| 010 | Środki trwałe | 1: grupa bilansowa (grunty, w tym prawo użytkowania wieczystego / budynki, lokale … / urządzenia techniczne i maszyny / środki transportu / inne) or KŚT group → 2: obiekt (nr inwentarzowy) | W | 010 |
| 020 | Wartości niematerialne i prawne | 1: koszty zakończonych prac rozwojowych / wartość firmy / inne WNiP → 2: składnik | W | 020 |
| 030 | Długoterminowe aktywa finansowe | 1: jednostki powiązane / z zaangażowaniem / pozostałe → 2: udziały lub akcje / inne papiery wartościowe / udzielone pożyczki / inne → 3: składnik, kontrahent | F, G (and any company lending long-term) | 030 *Udzielone pożyczki* (rename) |
| 035 | Odpisy aktualizujące długoterminowe aktywa finansowe | as 030 | F | — |
| 040 | Inwestycje w nieruchomości i prawa | 1: nieruchomości / WNiP / inne inwestycje długoterminowe → 2: składnik | F | — |
| 071 | Umorzenie środków trwałych | as 010 | W | 071 |
| 072 | Odpisy aktualizujące środki trwałe (trwała utrata wartości) | as 010; ŚT w budowie as a further line | X | — |
| 075 | Umorzenie wartości niematerialnych i prawnych | as 020 | W | 075 |
| 076 | Odpisy aktualizujące wartości niematerialne i prawne | as 020 | X | — |
| 077 | Umorzenie i odpisy aktualizujące inwestycji w nieruchomości i prawa | as 040 | F | — |
| 080 | Środki trwałe w budowie | 1: zadanie inwestycyjne → 2: rodzaj nakładu (incl. odsetki, różnice kursowe) | W | 080 |
| 081 | Zaliczki na środki trwałe w budowie i na WNiP | 1: na ŚT w budowie / na WNiP → 2: dostawca → 3: waluta | W | (249 today) |

### Zespół 1 — Środki pieniężne, rachunki bankowe oraz inne krótkoterminowe aktywa finansowe

| Konto | Nazwa | Analityka | Zakres | Rhea dziś |
|---|---|---|---|---|
| 100 | Kasa | 1: kasa → 2: waluta | W (if cash kept) | 100 |
| 130 | Rachunek bieżący | 1: rachunek bankowy | W | 130 |
| 131 | Rachunki bankowe w walutach obcych | 1: rachunek → 2: waluta | $ | — |
| 132 | Rachunek VAT | 1: rachunek VAT (per settlement account) | W (VAT) | — |
| 134 | Kredyty bankowe | 1: umowa → 2: część długoterminowa / krótkoterminowa → 3: kapitał / odsetki | X (most) | — |
| 135 | Inne rachunki bankowe | 1: ZFŚS / lokaty / rachunek powierniczy / pomocniczy (flag: ograniczona dyspozycja) | X | — |
| 139 | Środki pieniężne w drodze | 1: rodzaj (przelewy własne, wpłaty gotówki, rozliczenia kart) | W | 149 |
| 140 | Krótkoterminowe aktywa finansowe | 1: jednostki powiązane / pozostałe → 2: udziały lub akcje / inne papiery / udzielone pożyczki / inne / inne aktywa pieniężne → 3: składnik | F | — |
| 145 | Odpisy aktualizujące krótkoterminowe aktywa finansowe | as 140 | F | — |

### Zespół 2 — Rozrachunki i roszczenia

| Konto | Nazwa | Analityka | Zakres | Rhea dziś |
|---|---|---|---|---|
| 201 | Rozrachunki z odbiorcami | 1: relacja (powiązane / zaangażowanie / pozostałe) → 2: termin (do / powyżej 12 m-cy) → 3: waluta → 4: kontrahent → 5: dokument | W | 201 |
| 202 | Rozrachunki z dostawcami | as 201 (okres wymagalności) | W | 202 |
| 205 | Rozrachunki z tytułu zaliczek na dostawy i usługi | 1: udzielone / otrzymane → 2: relacja → 3: kontrahent | W | — |
| 221 | VAT naliczony i jego rozliczenie | 1: do odliczenia w bieżącym okresie / w następnych okresach / do korekty → 2 (optional): stawka | W (VAT) | 221 |
| 222 | VAT należny | 1: okres → 2: tryb (krajowy, WNT, import usług, odwrotne obciążenie, import towarów) | W (VAT) | 222 |
| 223 | Rozrachunki z urzędem skarbowym z tytułu VAT | 1: do zapłaty / nadwyżka do przeniesienia / zwrot → 2: okres | W (VAT) | — |
| 224 | Rozrachunki z tytułu podatku dochodowego | 1: zaliczki / podatek roczny / ryczałt od dochodów spółek | W | (225 today) |
| 225 | Pozostałe rozrachunki publicznoprawne | 1: tytuł (PIT-4 od wynagrodzeń, PIT-8AR, podatek u źródła, PCC, podatek od nieruchomości i środków transportu, cła, opłaty) → 2: organ | W | 225 |
| 229 | Rozrachunki z ZUS i innymi instytucjami | 1: ubezpieczenia społeczne / zdrowotne / Fundusz Pracy i Fundusz Solidarnościowy / FGŚP / FEP / PPK / PFRON | Z (and board members' contributions) | 229 |
| 231 | Rozrachunki z tytułu wynagrodzeń | 1: tytuł (umowa o pracę, zlecenie, dzieło, zarząd, rada nadzorcza) → 2: osoba | Z | 231 |
| 234 | Pozostałe rozrachunki z pracownikami | 1: tytuł (zaliczki, delegacje, pożyczki z ZFŚS, potrącenia) → 2: osoba | Z | — |
| 240 | Pożyczki | 1: otrzymane / udzielone krótkoterminowe → 2: relacja (udziałowcy, powiązane, pozostali) → 3: część długo- / krótkoterminowa → 4: umowa → 5: kapitał / odsetki | W (most) | 240 *otrzymane* |
| 241 | Rozliczenie niedoborów, szkód i nadwyżek | 1: niedobory / szkody / nadwyżki → 2: sprawa | X | 241 is *dywidendy* — move to 243 |
| 242 | Rozrachunki z właścicielami (wspólnikami spółek osobowych) | 1: wspólnik → 2: tytuł | S | — |
| 243 | Rozrachunki z udziałowcami (akcjonariuszami) | 1: udziałowiec → 2: tytuł (dywidenda, zaliczka na dywidendę, dopłaty i ich zwrot, umorzenie udziałów) | W | (241 today) |
| 245 | Rozrachunki wewnątrzzakładowe | 1: oddział | X | — |
| 246 | Należności dochodzone na drodze sądowej | 1: relacja → 2: dłużnik → 3: sprawa | X | — |
| 247 | Rozliczenia z operatorami płatności | 1: operator | X (e-commerce) | — |
| 249 | Pozostałe rozrachunki | 1: tytuł (kaucje i wadia, wpływy do wyjaśnienia, odszkodowania, dotacje do otrzymania, sprzedaż składników majątku, inne) → 2: relacja → 3: termin → 4: kontrahent | W | 249 |
| 250 | Należne wpłaty na kapitał podstawowy | 1: udziałowiec | W (until paid) | — |
| 280 | Odpisy aktualizujące należności | 1: pozycja należności (relacja × termin, sporne / sądowe, inne) → 2: dłużnik | W | 280 |

### Zespół 3 — Materiały i towary

| Konto | Nazwa | Analityka | Zakres | Rhea dziś |
|---|---|---|---|---|
| 300 | Rozliczenie zakupu | 1: materiały / towary / usługi → 2: dostawca, dokument | H, P | 300 |
| 310 | Materiały | 1: magazyn → 2: grupa / indeks (opakowania, materiały w drodze as lines) | P (X) | 310 |
| 330 | Towary | 1: magazyn (hurt / detal; w drodze) → 2: grupa / indeks | H | 330 |
| 340 | Odchylenia od cen ewidencyjnych materiałów i towarów | 1: materiały / towary / marża detaliczna / VAT w detalu | X (stałe ceny, detal) | — |
| 349 | Odpisy aktualizujące wartość zapasów | 1: materiały / towary | H, P | 349 |

### Zespół 4 — Koszty według rodzajów i ich rozliczenie

| Konto | Nazwa | Analityka | Zakres | Rhea dziś |
|---|---|---|---|---|
| 400 | Amortyzacja | 1: ŚT / WNiP / jednorazowa (niska wartość) → 2: KUP / NKUP | W | 401 |
| 401 | Zużycie materiałów i energii | 1: materiały, paliwo, energia i media, wyposażenie, materiały biurowe → 2: KUP / NKUP | W | 402 |
| 402 | Usługi obce | 1: najem i dzierżawa, leasing, transport, remonty, księgowe i prawne, IT i telekom, opłaty bankowe, usługi niematerialne od powiązanych → 2: KUP / NKUP | W | 403 |
| 403 | Podatki i opłaty | 1: podatek akcyzowy / od nieruchomości / od środków transportu / opłaty skarbowe i sądowe / PFRON / użytkowanie wieczyste / PCC → 2: KUP / NKUP | W | **missing** |
| 404 | Wynagrodzenia | 1: umowy o pracę / cywilnoprawne / zarząd / rada nadzorcza → 2: KUP / NKUP | Z | 404 |
| 405 | Ubezpieczenia społeczne i inne świadczenia | 1: składki pracodawcy, w tym emerytalne / FP i FS / FGŚP / PPK / odpis na ZFŚS / BHP / szkolenia / inne → 2: KUP / NKUP | Z | 405 |
| 409 | Pozostałe koszty rodzajowe | 1: podróże służbowe, reprezentacja, reklama, ubezpieczenia majątkowe, składki członkowskie, samochody osobowe, inne → 2: KUP / NKUP | W | 409, 409-1 |
| 490 | Rozliczenie kosztów | 1: zmiana stanu produktów / koszt wytworzenia na własne potrzeby / RMK / przeniesienie na zespół 5 | W | 490 |

*KUP / NKUP as an analytic level or as an off-balance register (group 9) — choose with the tags strand
(§2.3, [filings.md §1.5](filings.md)).*

### Zespół 5 — Koszty według typów działalności i ich rozliczenie (P or K only)

| Konto | Nazwa | Analityka | Zakres |
|---|---|---|---|
| 500 | Koszty działalności podstawowej — produkcyjnej i usługowej | 1: rodzaj działalności / zlecenie / produkt → 2: koszty bezpośrednie by nature | P, K |
| 510 | Koszty działalności podstawowej — handlowej | 1: rodzaj → 2: by nature | K (H) |
| 520 | Koszty wydziałowe | 1: wydział → 2: by nature | P |
| 527 | Koszty sprzedaży | 1: kanał / rodzaj → 2: by nature | K |
| 530 | Koszty działalności pomocniczej | 1: jednostka → 2: by nature | P |
| 550 | Koszty ogólnego zarządu | 1: by nature (notes ust. 2 pkt 2) | K |
| 580 | Rozliczenie kosztów działalności | 1: produkt / zlecenie | P |

### Zespół 6 — Produkty i rozliczenia międzyokresowe

| Konto | Nazwa | Analityka | Zakres | Rhea dziś |
|---|---|---|---|---|
| 600 | Produkty gotowe | 1: magazyn → 2: produkt | P | — |
| 601 | Półprodukty | 1: produkt | P | — |
| 602 | Produkty w toku | 1: zlecenie | P, B | — |
| 620 | Odchylenia od cen ewidencyjnych produktów | 1: produkty / półprodukty | P | — |
| 621 | Odpisy aktualizujące wartość produktów | 1: grupa | P | — |
| 640 | Czynne rozliczenia międzyokresowe kosztów | 1: tytuł (ubezpieczenia, licencje, czynsze z góry, koszty emisji, wycena kontraktów — B) → 2: długo- / krótkoterminowe → 3: umowa | W | 640 |
| 641 | Bierne rozliczenia międzyokresowe kosztów | 1: tytuł (usługi niezafakturowane, urlopy, premie, badanie sprawozdania, naprawy gwarancyjne) → 2: długo- / krótkoterminowe | W | 641 |
| 650 | Aktywa z tytułu odroczonego podatku dochodowego | 1: przez wynik / przez kapitał (art. 37 ust. 9) | W (M may omit) | 650 |
| 651 | Inne rozliczenia międzyokresowe | 1: tytuł → 2: długo- / krótkoterminowe | X | — |

### Zespół 7 — Przychody i koszty związane z ich osiągnięciem

| Konto | Nazwa | Analityka | Zakres | Rhea dziś |
|---|---|---|---|---|
| 700 | Sprzedaż produktów | 1: wyroby / usługi → 2: relacja (powiązane / pozostali) → 3: rynek (kraj / UE / poza UE) → 4: rodzaj działalności | W | 700 |
| 701 | Koszt sprzedanych produktów | as 700 (relacja for "jednostkom powiązanym") | K, P | — |
| 730 | Sprzedaż towarów | as 700 | H | 730 |
| 731 | Wartość sprzedanych towarów w cenie nabycia | 1: relacja | H | 731 |
| 740 | Sprzedaż materiałów (optional) | presented under D.IV / G.IV (KSR 15 pkt 8.1); Rhea's 760 analytic suffices | X | (760) |
| 741 | Wartość sprzedanych materiałów (optional) | presented under E.III / H.III | X | (761) |
| 750 | Przychody finansowe | 1: dywidendy (powiązane / zaangażowanie / pozostałe) / odsetki (powiązane / pozostałe) / zysk z rozchodu aktywów finansowych / aktualizacja / dodatnie różnice kursowe (zrealizowane / niezrealizowane) / inne → 2: przychód podatkowy / nie | W | 750 |
| 751 | Koszty finansowe | 1: odsetki (powiązane / pozostałe; budżetowe) / strata z rozchodu / aktualizacja / ujemne różnice kursowe / prowizje / inne → 2: KUP / NKUP | W | 751, 751-1 |
| 760 | Pozostałe przychody operacyjne | 1: zysk z rozchodu niefinansowych aktywów trwałych (powiązane / pozostałe) / dotacje / aktualizacja wartości aktywów niefinansowych / inne (sprzedaż materiałów, odszkodowania, przedawnione zobowiązania) → 2: przychód podatkowy / nie | W | 760 |
| 761 | Pozostałe koszty operacyjne | 1: strata z rozchodu / aktualizacja (odpisy na należności, zapasy, ŚT) / inne (darowizny, kary, niedobory, wartość sprzedanych materiałów) → 2: KUP / NKUP | W | 761, 761-1…3 |
| 790 | Obroty wewnętrzne | 1: oddział | X | — |
| 791 | Koszt obrotów wewnętrznych | 1: oddział | X | — |

Rhea's 756 *Różnice kursowe i zaokrąglenia* has no counterpart: FX differences → 750/751 analytics; where
conversion roundings go is *to verify*.

### Zespół 8 — Kapitały (fundusze) własne, fundusze specjalne, rezerwy i wynik finansowy

| Konto | Nazwa | Analityka | Zakres | Rhea dziś |
|---|---|---|---|---|
| 801 | Kapitał podstawowy | 1: rodzaj (zakładowy / akcyjny / wkłady wspólników) → 2: udziałowiec, seria | W | 801 |
| 802 | Kapitał zapasowy | 1: nadwyżka wartości emisyjnej nad nominalną / z podziału zysku / inne → 2 (E): zyski i straty z lat przed / w trakcie ryczałtu | W | 802 |
| 803 | Pozostałe kapitały rezerwowe | 1: tworzone zgodnie z umową / na udziały własne / inne | X | — |
| 813 | Kapitał z aktualizacji wyceny | 1: z aktualizacji wartości godziwej / inne | X | 813 |
| 820 | Rozliczenie wyniku finansowego | 1: zysk / strata z lat ubiegłych (per rok) / podział zysku / pokrycie straty / korekty błędów i zmian zasad (KSR 7) / odpisy z zysku netto w ciągu roku | W | 820 |
| 830 | Rezerwy na zobowiązania | 1: na świadczenia emerytalne i podobne / pozostałe (tytuł) → 2: długo- / krótkoterminowe | W | 830 |
| 835 | Rezerwa z tytułu odroczonego podatku dochodowego | 1: przez wynik / przez kapitał | W (M may omit) | 835 |
| 840 | Rozliczenia międzyokresowe przychodów | 1: tytuł (przychody przyszłych okresów, dotacje do aktywów, nieodpłatnie otrzymane aktywa, ujemna wartość firmy) → 2: długo- / krótkoterminowe | W | **missing** |
| 851 | Zakładowy fundusz świadczeń socjalnych | 1: stan, zwiększenia, zmniejszenia → 2: rodzaj działalności socjalnej | X (if a fund is kept) | — |
| 852 | Inne fundusze specjalne | 1: fundusz | X | — |
| 860 | Wynik finansowy | (optional) lines of the result | W | 860 |
| 870 | Podatek dochodowy | 1: część bieżąca / część odroczona (art. 37 ust. 8) → 2: CIT / ryczałt od dochodów spółek | W | 870 (no split) |
| 871 | Pozostałe obowiązkowe zmniejszenia zysku | 1: tytuł | X | — |

### Zespół 9 — Konta pozabilansowe

| Konto | Nazwa | Analityka | Zakres |
|---|---|---|---|
| 910 | Obce środki trwałe (najem, dzierżawa, leasing operacyjny, użyczenie) | 1: umowa | X (common) |
| 911 | Środki trwałe i WNiP w likwidacji / wycofane | 1: obiekt | X |
| 912 | Składniki majątku o niskiej wartości w używaniu (ewidencja ilościowa) | 1: osoba / miejsce | X |
| 920 | Należności warunkowe | 1: tytuł → 2: kontrahent | X |
| 921 | Zobowiązania warunkowe (poręczenia, gwarancje, weksle) | 1: tytuł → 2: beneficjent | W (if any) |
| 922 | Zabezpieczenia ustanowione na majątku (hipoteki, zastawy) | 1: umowa | X |
| 930 | Zapasy obce | 1: właściciel | X |
| 941–948 | Różnice między wynikiem bilansowym a podatkowym, the eight categories of § 2 ust. 1 pkt 5 lit. a–h: 941 przychody zwolnione (trwałe); 942 przychody niepodlegające opodatkowaniu w bieżącym roku; 943 przychody podatkowe bieżącego roku ujęte w księgach lat ubiegłych; 944 przychody podatkowe nieujmowane w księgach; 945 koszty NKUP (trwałe); 946 koszty niebędące KUP w bieżącym roku; 947 KUP bieżącego roku ujęte w księgach lat ubiegłych; 948 KUP nieujmowane w księgach | 1: tytuł | W (CIT) — or attributes instead, per the tags strand |
| 999 | Rozliczenie kont pozabilansowych | only if the system needs double entry off balance (Infor 2006; GOFIN's article: "Konto techniczne") | X |

### Size

| Group | Synthetic accounts | **W** | **Z** | Other scopes |
|---|---|---|---|---|
| 0 | 12 | 6 — 010, 020, 071, 075, 080, 081 | — | 6 |
| 1 | 9 | 4 — 100, 130, 132, 139 | — | 5 |
| 2 | 21 | 13 — 201, 202, 205, 221, 222, 223, 224, 225, 240, 243, 249, 250, 280 | 3 — 229, 231, 234 | 5 |
| 3 | 5 | — | — | 5 |
| 4 | 8 | 6 — 400, 401, 402, 403, 409, 490 | 2 — 404, 405 | — |
| 5 | 7 | — | — | 7 |
| 6 | 9 | 3 — 640, 641, 650 | — | 6 |
| 7 | 12 | 5 — 700, 750, 751, 760, 761 | — | 7 |
| 8 | 13 | 8 — 801, 802, 820, 830, 835, 840, 860, 870 | — | 5 |
| 9 | 16 | 9 — 921 (if any), 941–948 | — | 7 |
| **Total** | **112** | **54** | **5** | **53** |

Without group 9 the **W** core is 45 accounts — roughly what a services sp. z o.o. on the comparative RZiS,
with no stock and no foreign currency, opens in year one (plus the five Z accounts once it pays anyone).
Rhea's chart already has 35 of those 45 numbers, three of them (401–403) with a different meaning; the
other ten are 081, 132, 139 (Rhea uses 149), 205, 223, 224, 243 (Rhea uses 241), 250, 400 (so that
*Podatki i opłaty* can sit at 403) and 840.

### Mapping of Rhea's 55 accounts to the skeleton

| Rhea | Skeleton | Action |
|---|---|---|
| 010, 020, 071, 075, 080, 100, 130, 201, 202, 221, 222, 225, 229, 231, 249, 280, 300, 310, 330, 349, 404, 405, 409, 490, 640, 641, 650, 700, 730, 731, 750, 751, 760, 761, 801, 802, 813, 820, 830, 835, 860, 870 | same | keep; add analytics (§3.3); fix `type` of contra accounts (071, 075, 280, 349) |
| 409-1, 751-1, 761-1, 761-2, 761-3 | analytics of 409, 751, 761 | keep as analytics (or move the NKUP view to 945/946) |
| 401 Amortyzacja → 400; 402 Zużycie → 401; 403 Usługi obce → 402 | 400–402 | renumber (or keep and document); add 403 *Podatki i opłaty* |
| 030 *Udzielone pożyczki* | 030 *Długoterminowe aktywa finansowe* + 240 for short-term loans granted | rename; split by term |
| 149 | 139 | renumber (optional) |
| 241 *dywidendy* | 243-dywidendy | renumber |
| 240 *Pożyczki otrzymane* | 240 *Pożyczki* | rename; add received / granted, relation and term analytics |
| 756 | 750 / 751 FX analytics; roundings *to verify* | dissolve |
| (249 for asset advances) | 081 | reclassify |
| (225 for CIT) | 224 | reclassify |
| (870 for deferred tax) | 870 odroczony analytic | split |

---

## 5. What remains to verify

1. **SKwP model chart** — none found publicly; ask SKwP or a member.
2. **Vendor default charts** (Comarch ERP Optima, Symfonia, enova365, Rewizor) — only mechanics are
   public; the generated lists need a licensed install or a client's export.
3. **Commercial commentaries** (Wolters Kluwer LEX *Wzorcowy plan kont*, ODDK *Plan kont z komentarzem*) —
   paywalled; not reviewed.
4. **Renaming and renumbering mid-year** — no statutory text; my reading rests on art. 10 ust. 2, art. 18
   and Stanowisko 2010 pkt 37, 41–42; confirm with a practising accountant or auditor.
5. **A pre-1995 ministerial model chart for enterprises** — not found; irrelevant to today's law but
   possibly the origin of the shared numbering.
6. **Rounding differences from currency conversion** (Rhea 756) — other operating or financial? Practice
   and an MF tax position exist only for VAT-return roundings (REVIEW, rounding finding).
7. **Presentation of bierne RMK** (641) in the balance sheet — the statements strand / KSR 6.
8. **How maturity (do / powyżej 12 m-cy) is assigned** — at the balance date by remaining or original
   term — the statements strand.
9. **KUP / NKUP as analytics or as a group-9 register** — decide with the JPK tags strand (PD tags one per
   analytic account; off-balance `…_PB` tags).
10. **The newer consolidated text of the public-sector chart** (Dz.U. 2026 poz. 909) — found by search, not
    read.
11. **ZFŚS obligation threshold** and **dopłaty** presentation in a sp. z o.o. — not researched here.
12. **Whether 7xx analytics by VAT rate are common practice** — not confirmed.

---

## Sources

**Law and standards (primary)**
- Ustawa o rachunkowości, t.j. [Dz.U. 2026 poz. 522](https://eli.gov.pl/api/acts/DU/2026/522/text/T/D20260522L.pdf) (ISAP [WDU19941210591](https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU19941210591)) — art. 5, 7, 8, 9, 10, 12, 13, 15–18, 23, 28a–28b, 30, 32, 37, 39, 47, 50, 83; Annex 1.
- Ustawa o podatku dochodowym od osób prawnych, t.j. [Dz.U. 2026 poz. 554](https://eli.gov.pl/api/acts/DU/2026/554/text/T/D20260554L.pdf) — art. 9 ust. 1, 1b, 1c; art. 24e.
- Rozporządzenie MF z 16.08.2024 w sprawie dodatkowych danych … ([Dz.U. 2024 poz. 1314](https://eli.gov.pl/eli/DU/2024/1314/ogl/pol/pdf)) — § 2, § 5, Annex 7.
- KSR 7 — [Dz. Urz. MF 2019 poz. 2](https://www.gov.pl/attachment/9da792cf-ca83-4ace-9779-17aa810a8548); KSR 15 (2026) — [Dz. Urz. MFiG 2026 poz. 53](https://www.gov.pl/attachment/143e3deb-6668-4506-b191-b0efcce109f7); KSR 16 — [Dz. Urz. MFiG 2025 poz. 29](https://www.gov.pl/attachment/feba9252-201d-4357-93b3-b3dbeae1c59a); [list of KSR](https://www.gov.pl/web/finanse/krajowe-standardy-rachunkowosci).
- [Stanowisko KSR w sprawie niektórych zasad prowadzenia ksiąg rachunkowych (uchwała 5/10, 13.04.2010)](https://www.gov.pl/attachment/ecd3a097-3241-4492-b891-16b811b85aae).
- Wzorcowy plan kont dla banków, t.j. [Dz.U. 2019 poz. 946](https://eli.gov.pl/eli/DU/2019/946/ogl).
- Plan kont dla jednostek budżetowych, rozporządzenie MRiF z 13.09.2017, t.j. [Dz.U. 2020 poz. 342](https://eli.gov.pl/eli/DU/2020/342/ogl/pol/pdf).

**Model charts and vendor practice**
- [GOFIN — Wzorcowy wykaz kont (pełny)](https://indeksksiegowan.gofin.pl/firma/wzorcowy-wykaz-kont).
- [jak-ksiegowac.pl — Plan kont (PDF)](https://www.jak-ksiegowac.pl/files/Plan-kont.pdf), [page](https://www.jak-ksiegowac.pl/plan-kont.php).
- [Wikibooks — Przykładowy Zakładowy Plan Kont](https://pl.wikibooks.org/wiki/Rachunkowo%C5%9B%C4%87/Konta/Plany_kont/Przyk%C5%82adowy_Zak%C5%82adowy_Plan_Kont).
- [WNE UW — Zakładowy plan kont (teaching, 2010)](http://coin.wne.uw.edu.pl/akocia/rach_fi/Zakladowy%20plan%20kont.doc).
- wFirma — [Plan kont](https://pomoc.wfirma.pl/-plan-kont); [Plan kont 2.0](https://pomoc.wfirma.pl/-wzorcowy-plan-kont-2-0-w-systemie-wfirma-pl); [comparison 1.0 / 2.0 (PDF)](https://pomoc.wfirma.pl/download/wzorcowy-plan-kont-2-0-w-systemie-wfirma-pl.pdf/636385).
- Comarch — [Optima: Plan kont](https://pomoc.comarch.pl/optima/pl/2026/dokumentacja/plan-kont/); [Optima: Ćwiczenie: Plan kont](https://pomoc.comarch.pl/optima/pl/2024_5/?p=4705); [Optima: Jak przygotować plik JPK_KR_PD](https://pomoc.comarch.pl/optima/pl/2026/dokumentacja/jak-przygotowac-plik-jpk_kr_pd/); [Altum: Konta słownikowe](https://pomoc.comarch.pl/altum/documentation/konta-slownikowe/).
- [Rewizor nexo retailer listing](https://www.empik.com/insert-rewizor-nexo-1stan-insert,p1624244911,multimedia-p); [Wolters Kluwer LEX — Wzorcowy plan kont](https://wolterskluwer.com/pl-pl/solutions/lex/ksiegowosc/baza-prawa/wzorcowy-plan-kont) (not readable, 403).

**Professional commentary**
- K. Szopa, [Zmiana zakładowego planu kont jednostki – kiedy stanowi zmianę polityki rachunkowości?](https://poradnikprzedsiebiorcy.pl/-zmiana-zakladowego-planu-kont-jednostki-kiedy-stanowi-zmiane-polityki-rachunkowosci), Poradnik Przedsiębiorcy, 29.05.2025.
- M. Sądej, [Zakładowy plan kont a konta nieużywane przez jednostkę](https://poradnikprzedsiebiorcy.pl/-zakladowy-plan-kont-a-konta-nieuzywane-przez-jednostke), Poradnik Przedsiębiorcy.
- P. Łomińska, [Zakładowy plan kont – kiedy dokonywać zmian?](https://ksiegowosc.infor.pl/raport-dnia/143512,Zakladowy-plan-kont-kiedy-dokonywac-zmian.html), Infor, 29.01.2014.
- [Zmiana systemu finansowo-księgowego w trakcie roku obrotowego](https://www.podatki.biz/artykuly/zmiana-systemu-finansowo-ksiegowego-w-trakcie-roku-obrotowego_23_33155.htm), podatki.biz, 03.11.2016.
- A. Pokojska, [Jak ujmować koszty w rachunkowości](https://ksiegowosc.infor.pl/poradniki/1185,2,Jak-ujmowac-koszty-w-rachunkowosci.html), Infor / Gazeta Prawna, 11.04.2007.
- D. Przybyszewska, [Ustalenie dochodu do opodatkowania na podstawie zapisów w księgach rachunkowych](https://www.pit.pl/?p=11769), Gazeta Podatkowa, 27.08.2024.
- E. Szczepankiewicz, [Jak zaksięgować koszty niebędące kosztami uzyskania przychodów](https://ksiegowosc.infor.pl/porady-prawne/29724,Jak-zaksiegowac-koszty-niebedace-kosztami-uzyskania-przychodow.html), Infor / Monitor Księgowego, 12.04.2006.
- [Poradnik Przedsiębiorcy — Plan kont – jak dostosować go do potrzeb przedsiębiorstwa?](https://poradnikprzedsiebiorcy.pl/-plan-kont-jak-dostosowac-go-do-potrzeb-przedsiebiorstwa) (K. Szopa, 12.05.2023).

**Rhea (read-only)**
- `/home/krzysztof/Rhea/packs/pl/pack.json` (v44): `account.created`, `account.tax_treatment.set`, the `account` type, posting rules.
- `/home/krzysztof/Rhea/packs/pl/REVIEW.md` FB-04 to FB-08; `/home/krzysztof/Rhea/DIRECTION.md` "Invoicing per country, and each company's own chart" (D3).
- Sibling note: [filings.md](filings.md).
