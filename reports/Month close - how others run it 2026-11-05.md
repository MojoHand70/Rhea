# Month close: how others run it, and what Rhea has (2026-11-05)

KK, 2026-11-05 evening: "Zamknięcie is a whole another story and a separate functionality. And we
need to do it right. Check how others are doing it, like numeric.io for example. Rhea needs to
prepare a checklist of close steps which need confirmation that those steps are done. I don't
think we have this functionality today." He is right about the last point; this note gathers
what the field does, from their public pages, read the same evening.

## What Rhea has today

- A door, `close_period` (common periods pack): a person names the month and the book; the fact
  is `period.closed`; a month closes once per book. A closed month refuses postings per event.
  The Poland pack settles on this fact: the VAT settlement of the month (vat.json) and the CIT
  advance, cumulative from January (cit.json), recomputed when an earlier month's figures move.
- A year close (`close_year`): result accounts to 860 at the balance date; approval and
  distribution doors after it.
- Tasks (`task`, with a `month_close` category) and task schedules — nothing creates close
  tasks today.
- The morning note's "Co jest otwarte": months to close, unsettled statement lines, open tasks.
- The tax calendar: the 20th (CIT/PIT advances) and the 25th (VAT, JPK_V7M), Easter-aware.

No checklist, no per-step confirmation, no evidence per step, no office-wide view of the close.

## Numeric (numeric.io, pages read 2026-11-05)

Close checklist = a task workspace: "Assign owners, instruct your agents via MCP, tag items, map
dependencies, attach workbooks, and keep a full audit log for every step" (product/close). Tags
by "department, entity, control, or anything else"; multi-entity is tags plus Saved Views. The
audit trail: "every comment, task submission, attachment, and more with timestamps". No reviewer
field, due-date offset or recurrence is named on the product pages; the blog recommends
"Preparer and reviewer tags (plus due dates for each), as well as business day due dates" and
dependencies on a timeline (blog/streamline-month-end-close, updated 2026-05-05).

Automation: reconciliations read their sources and are "evaluated against your GL. Set
materiality thresholds, track and resolve differences, and auto-submit your recs"; flux
explanations "auto-drafted" (home page: 80% of first drafts by AI, a customer quote); monitors
"check for duplicate transactions, unusual balances, and elevated period spending" on live
transaction data; agents "do the recurring work your checklist defines: accruals, sales commission
updates, and follow-ups", fed from the checklist through MCP. KPI named: days to close.
Integrations: NetSuite first, "and other ERPs"; QuickBooks via MCP. Customers named: Brex, Plaid,
Wealthfront, Riskified, Benchling — US tech, public-company closes.

## FloQast (floqast.com, pages read 2026-11-05)

"Intelligent checklists that track every task by owner, entity, task duration, dependencies and
due date"; "set owners, due dates, and review steps"; preparer and reviewer roles; tags such as
#high_risk and #journal; "automate task rollovers for the next period"; an existing checklist
can be mapped in "with the help of AI". Completion: "support, review notes, and sign-offs
attached to the work"; reconciliations with "a completeness check [that] confirms every account
is reconciled before you close". No auto-completion of a checklist task from ledger conditions is
described. Dashboards: "early, on-time, and at-risk tasks" by team member, process and entity;
close trends exported for retrospectives. AI: matching, automated reconciliations, flux
explanations "drafted for review", agents that "execute [a process described in plain language]
consistently every cycle", with "Nothing hits your books without human approval."

## Rillet (rillet.com, pages read 2026-11-05)

No checklist product page. The pitch is the continuous close: "Checklists, reconciliations, and
flux drafts run continuously to shave days off the close"; subledgers (revenue waterfall,
deferred revenue, prepaids, fixed assets) generated from contracts and bills; bank transactions
"auto-matched in seconds"; an agent that "reasons like an accountant" over the GL with traceable
sources. Marketing figures (99.7% automatic posting) unverified. No owner/reviewer/due-date model
published.

## BlackLine

Its close-management and task-management pages did not load (an empty page and an HTTP 500) when
read; nothing is claimed here about it.

## Polish practice (sources read the same evening)

Sources: [S] Symfonia, "Koniec miesiąca w małej firmie: checklista"
https://symfonia.pl/blog/finanse-i-ksiegowosc/koniec-miesiaca-w-malej-firmie-checklista-dokumentow-dla-wlasciciela-i-ksiegowej/ ·
[PP1] poradnikprzedsiebiorcy, "Co oznacza zamknięcie ksiąg rachunkowych danego miesiąca"
https://poradnikprzedsiebiorcy.pl/-co-oznacza-zamkniecie-ksiag-rachunkowych-danego-miesiaca ·
[PP2] poradnikprzedsiebiorcy, "Zamknięcie ksiąg – o czym należy pamiętać"
https://poradnikprzedsiebiorcy.pl/-zamkniecie-ksiag-rachunkowych-o-czym-nalezy-pamietac-w-firmie ·
[EY] EY outsourcing scope sheet (PDF)
https://www.ey.com/content/dam/ey-unified-site/ey-com/pl-pl/services/uslugi-ksiegowe-ey/documents/outsourcing-procesow-ey.pdf ·
[CX] Comarch XL https://pomoc.comarch.pl/xl/2024/?p=49684 · [CXT] Comarch ERP XT
https://pomoc.comarch.pl/erpxtstart/index.php/dokumentacja/jak-zamknac-miesiac-przy-wspolpracy-z-biurem-rachunkowym/ ·
[CO] Comarch Optima (biuro) https://pomoc.comarch.pl/optima/pl/2025/?p=43616 · [CA] Comarch Altum
https://pomoc.comarch.pl/altum/?p=19943 · [E1] enova365
https://enova.pl/blog/zamkniecie-miesiaca-ile-czasu-zajmuje-porownaj-sie-z-innymi/ · [E2] enova365
https://enova.pl/blog/modul-finansowo-ksiegowy-enova365-top6-najbardziej-uzytecznych-funkcji/ ·
[W1] wFirma "Rozlicz miesiąc" https://pomoc.wfirma.pl/-funkcja-rozlicz-miesiac · [W2] wFirma okresy
https://pomoc.wfirma.pl/-otwieranie-i-zamykanie-okresow-ksiegowych · [W3] wFirma komplet dokumentów
https://pomoc.wfirma.pl/-funkcja-potwierdzenia-przez-przedsiebiorce-kompletu-dokumentow · [W4] wFirma kokpit
https://pomoc.wfirma.pl/-zestawienie-informacji-o-podatkach-i-zus-klientow

No Polish source gives one numbered checklist; [S] and [PP1] say the law sets no procedure or
date ("ustawa nie określa jednolitej procedury ani daty"). Assembled from what each names, in the
practitioners' order:

1. Kompletność dokumentów: the client confirms the month's documents are all delivered —
   Comarch's "Zamknij miesiąc" [CXT], wFirma's "KOMPLET DOKUMENTÓW → POTWIERDŹ" [W3], [S]; Comarch's
   suggested automatic cut-off is the 15th of the next month [CO].
2. Everything posted before the close: "do momentu zamknięcia miesiąca jednostka powinna
   wprowadzić do ksiąg wszystkie zapisy" [PP1]; daily statement posting [EY].
3. Kasa, bank, kredyty uzgodnione — monthly [PP2]; the most time-consuming step [E1].
4. Rozrachunki uzgodnione — monthly [PP2]; balance confirmations and ageing [EY] [CX]; ulga na
   złe długi, przedawnienie, kompensaty [PP2].
5. Rejestry VAT → JPK_V7M by the 25th [S]; rates, tax point, GTU, WDT documents, VAT-UE [EY] [W1].
6. Listy płac i ZUS — DRA and contributions on the 5th/15th/20th [S]; PIT-4R, PPK [W1].
7. Amortyzacja, księgowa i podatkowa [EY]; enova checks every asset has its depreciation [E2].
8. RMK czynne i bierne, rezerwy — monthly [PP2] [EY] [E1].
9. Magazyn i produkcja: stock at the date against the stock account [CX]; production in
   progress, price variances [PP2].
10. Różnice kursowe — monthly, balance-sheet and tax [PP2]; revaluation per currency [CX].
11. Zestawienie obrotów i sald at least monthly [S]; "per saldo" must be zero, dzienniki agree
    [CX]; group 9 against revenues and costs [PP2]. No source names 300 or 490 by number.
12. Zaliczka CIT by the 20th — the close is usually timed to it [PP1]; permanent and temporary
    differences monthly [PP2]; computation, payment instruction, KUP/NKUP review [EY].
13. Zamknięcie okresu w programie — the period lock (below).
14. Raport dla klienta lub zarządu [S] [EY]; wFirma publishes to the client and sends the
    tax/ZUS note [W1]. A formal client sign-off on balances was not found in any source.

Legal anchors the sources name: UoR art. 25 ust. 1 pkt 1 and ust. 2 — corrections "nie mogą
nastąpić po zamknięciu miesiąca", afterwards only a correcting voucher [PP1]; UoR art. 12
(yearly and event closings) [PP1] [PP2]; art. 13 ust. 1 [PP1]; art. 26 ust. 1 (stocktake) [PP2];
art. 24 ust. 5 pkt 2 (the trial balance monthly, within 85 days) stated by [S] without the
number; the KSR stance "w sprawie niektórych zasad prowadzenia ksiąg rachunkowych", chapter VI,
which defines a month's close as the entity's own exclusion of further entries [PP1]; VAT act
art. 145m (the KSeF relief to 2026-12-31) and art. 100 (VAT-UE) [S]. Nobody cites the VAT act's
art. 99/109 or updop art. 25; the 25th and the 20th are stated bare.

Software: Comarch XT/Optima give the client a "Zamknij miesiąc" button, optional or mandatory with
auto-close on a default day (15), checking the buffer and unsent KSeF invoices first, blocking
new documents by date; only the accountant reopens; the office sees "Klient zamknął m-c" [CXT]
[CO]. Comarch XL "Zamknij okres" is irreversible [CX]; Altum has "Zamknięty wstępnie" and
"Zamknięty", permission-gated [CA]. wFirma walks a month through otwarty → do księgowania → w
trakcie księgowania → zamknięty → ostatecznie zaksięgowany, warns of unposted drafts, and "Rozlicz
miesiąc" generates every filing and publishes to the client [W1] [W2]; a cockpit colours each
client's declarations grey/red/green [W4]. **A checklist with confirmations exists in enova365
Workflow:** "konfigurując proces zamknięcia miesiąca, możesz także zaznaczyć kroki, które musi
wykonać księgowy … oznaczyć je jako wykonane, żeby móc zamknąć miesiąc", beside automatic checks
(documents approved, cost processes finished, depreciation posted, outliers) [E2] [E1]. Saldeo's
help refused the read; nothing found for mOrganizer or Fakturownia.

## What this says for Rhea

The field's checklist is a task list ticked by people, with reconciliations and a completeness
gate around it; the one Polish instance (enova365) is the same shape inside a workflow module.
Rhea holds the books, so most steps are conditions over facts she evaluates herself, with the
facts as evidence; a person confirms only what the books cannot see. The list itself is nobody's
standard in Poland, so the pack's list is a standard Rhea proposes (DECISIONS 2026-10-16 on
helping create Polish standards). The proposal page for KK, 2026-11-05:
https://claude.ai/artifact/Ltd19upV3BhUz1ncocWxSo.
