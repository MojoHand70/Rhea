# PL corpus — evidence-derived market vocabulary

This file is the accumulation point for market knowledge that arrives as
**evidence** — screens of production systems, statutory forms, accountant
practice — rather than as spec. Each evidence session appends a dated entry;
the vocabulary, status ladders and rule intents below are the distillate.

The corpus feeds three consumers:

1. **The pack** (`pack.json`): object types distilled here land as pack data,
   get derived list/detail views automatically, and wait for rules.
2. **The agent**: this file is drafting context. A rule intent plus a target
   object type from here is a complete drafting request.
3. **The notion library**: screens of real products are the best specimens of
   the unbuilt notions (reconciliation, scheduling) — recorded here as specs
   until the kernel renders them.

Discipline: **screens are evidence, not design.** We take vocabulary, field
semantics, status ladders and rule intents. We never take screen layouts; the
shell renders notions generically, and an invoice-specific screen is a bug.

---

## Evidence sessions

- **2026-10-05** — 39 screens of a production Polish SME finance suite,
  embedded in a major bank's online banking (category evidence; vendor
  irrelevant). Coverage: dashboard, KSeF (archive, certificates, settings,
  automation), sales documents, cost documents, bank operations and
  reconciliation, payment planning, receivables monitoring, ZUS/taxes,
  month close, JPK, fixed assets, vehicles, mileage, leases, stocktake,
  task center with recurring schedules. All vocabulary below traces to this
  session unless marked otherwise.

---

## Vocabulary — object types

Status: **pack vN** (shipped in pack.json version N), **base** (already in
the base finance seed), **candidate** (recorded, not yet shipped).

| Type | Status | Evidence (original terms) |
|---|---|---|
| contractor | pack v5 | Kontrahent: NIP, REGON, adres, rodzaj/termin płatności, metoda kasowa, "bardzo ważny", rachunki bankowe, monitoring płatności |
| bank_account | pack v5 | Rachunek bankowy: nazwa (konto komfort/plus/max), numer, waluta (PLN/EUR/USD), bank |
| bank_operation | pack v5 | Operacja bankowa: data wykonania, kontrahent, tytuł, kwota, rachunek, typ (przelew/karta/opłata), "oznacz jako prywatny" |
| settlement | pack v5 | Rozliczenie / łączenie faktur z przelewami: operacja ↔ dokument, kwota, data |
| task | pack v5 | Zadanie: termin i opóźnienie, priorytet (problem/krytyczny/wysoki/średni/niski), kategoria, status, zlecający, wykonawca |
| task_schedule | pack v5 | Harmonogram zadań: typ (cotygodniowy/comiesięczny/coroczny/inny), następne/ostatnie wykonanie, systemowe/własne |
| recurring_invoice_template | pack v5 | Faktura cykliczna: nazwa, faktura źródłowa, data najbliższego wystawienia, wystawiać od/do, wysyłka do KSeF, wysyłka e-mail |
| fixed_asset | pack v5 | Środek trwały: numer ŚT, KST, wartość podatkowa/bilansowa zakupu, data przyjęcia, typ/stopa/współczynnik amortyzacji, likwidacja (data, przyczyna), sprzedaż, umorzenie z poprzedniego systemu |
| vehicle | pack v5 | Pojazd: numer rejestracyjny, opis, typ, pojemność silnika, własność, aktywny, domyślny |
| mileage_entry | pack v5 | Ewidencja przebiegu: data, początek, koniec, ilość km, pojazd |
| lease_contract | pack v5 | Umowa leasingowa: typ (operacyjny/finansowy), numer, początek/koniec, netto/VAT/brutto, limit odliczenia VAT, VAT odliczony/pozostały |
| tax_declaration | pack v5 | Deklaracja: rodzaj (JPK-V7, PIT, ZUS DRA, PCC-3), okres, termin płatności, status deklaracji, status płatności, do zapłaty |
| stocktake | pack v5 | Remanent: data inwentaryzacji, komentarz, status, wartość magazynu |
| sales_invoice | base | Dokument sprzedażowy; evidence adds: tryb formularza, mechanizm podzielonej płatności, metoda kasowa VAT, JPK V7 flag — candidates for a type version bump when rules need them |
| purchase_invoice | base | Dokument kosztowy; evidence adds: noty korygujące, koszty stałe (recurring costs → same template mechanism as recurring_invoice_template) |
| company | base | Overlaps contractor; merge/promotion deferred until a rule needs both unified (see DECISIONS 2026-10-05) |
| ksef_submission | pack v1 | Evidence adds the full status ladder and certificate kinds (below) — candidates for a version bump |
| correction_note | candidate | Nota korygująca: data wystawienia, numer, numer dokumentu źródłowego, kontrahent, transakcja |
| jpk_file | candidate | JPK: okres od/do, rodzaj deklaracji, podatek należny/naliczony, liczba wierszy sprzedaży/zakupu, rodzaj generowania, data kalkulacji — likely an analysis view over postings + a generated document, not hand-kept state |

Conventions carried from the base seeds: English field names, Polish only in
authored view labels; enums for flags (`yes`/`no`) and kinds; `money` for
amounts; rates as decimal strings (the `int` percent on vat_rate does not fit
amortization's 2.5%).

## Status ladders (rule material)

Observed state machines; each transition is an event, each status a derived
field — these are rules, not workflow engines:

- **sales document**: wprowadzony → zaksięgowany → (storno | skorygowany)
- **KSeF dispatch**: do wysłania → wysłany-oczekuje → przyjęty przez KSeF →
  potwierdzony UPO; błąd wysyłki | odrzucony | wyłączone z KSeF; plus
  "numer KSeF wprowadzony ręcznie" (manual override is a legal state, an
  event like any other)
- **payment**: nieopłacone → częściowo opłacone → opłacone; orthogonal:
  oczekujące na rozliczenie
- **bank operation settlement**: nierozliczone → rozliczone | prywatne
  (private expense is a terminal settlement state, not a deletion)
- **task**: niewykonane → wykonane | odrzucone | anulowane | zablokowane;
  opóźnione is derived from due_date vs time, not stored
- **declaration**: niezatwierdzona → zatwierdzona → skorygowana; payment
  status rides separately (nieopłacone/opłacone/częściowo)
- **month close**: open → kontrola uruchomiona → dokumenty zweryfikowane →
  zamknięty (maps onto period_lock; the wizard is a worklist over
  precondition rules, not a bespoke screen)

## Aging buckets (analysis material)

Receivables and payables both bucket identically: w terminie | 1–14 |
15–30 | 31–60 | 61–180 | >180 dni po terminie. One analysis view
parameterized by direction, not two screens.

## Notion specimens

### reconciliation (notion spec'd "later" in SPEC §2; first real specimen)

Two collections side by side, matched/unmatched:
- **left**: documents (sales/purchase invoices, declarations) with amount,
  contractor, due date
- **right**: bank operations with amount, counterparty string, title, date
- **match proposals**: rules score candidate pairs (amount equality, NIP in
  title, invoice number in title, date proximity); a human confirms via an
  activity; confirmation emits a settlement event
- **remainder surfaces**: "unknown contractor: 15 operations / 45 287 PLN",
  "unmatched payments: 500 PLN" — analysis tiles over the unmatched set
- partial matches allowed (one operation settles several documents and vice
  versa); "mark as private" removes an operation from the working set

### scheduling (M5; two specimens)

- **task schedules**: cadence (weekly/monthly/yearly), next/last run,
  requester/assignee; firing a schedule = a time event matching a rule that
  emits a task event. Determinism requires time passing to be **in the log**
  (`time.day_opened` / `time.month_opened` injected by a scheduler adapter);
  rules never read the wall clock.
- **payment planner**: 4 forward weeks, each with opening balance, inflows
  (receivables due), outflows (payables + declarations due), closing
  balance. Pure analysis over due dates and bank balances — needs no
  scheduler, only the bank_operation and settlement vocabulary.

## Rule intents (drafting backlog, PL)

Each line is a drafting request for the agent against the types above:

1. Bank operation with NIP in title matching an open invoice of equal gross →
   propose settlement (worklist; approval emits settlement).
2. Sales invoice unpaid N days past termin płatności → emit task
   (category: contractor_payments) with due date.
3. Draft documents exist and month-end is within 5 days → emit task
   "zaksięguj robocze dokumenty" (the observed suite ships exactly this nudge
   as an opaque recommendation; here it is an auditable rule).
4. KSeF submission accepted → derive UPO status onto the invoice;
   rejected → task (category: invoices, priority: high).
5. Fixed asset accepted with amortization_type=linear → emit monthly
   amortization posting events on month_opened until fully depreciated
   (blocked on M5 time events).
6. Recurring invoice template active and day_of_month reached → emit
   invoice.issued event from template (blocked on M5).
7. Lease (operational) with VAT deduction limit → cap deductible VAT per
   posting; remainder to costs.
8. Operation marked private → excluded from VAT register and cost totals
   (settlement state drives the analysis views' WHERE clause).
9. Declaration approved and unpaid with due date within 7 days → task
   (category: taxes).
10. Month close requested → verify preconditions (no drafts, no unmatched
    operations in period) → period_lock event; close checklist is the
    worklist filtered to these rules' outputs.

## Open language questions surfaced by this evidence

- **Union refs**: a settlement references "a document" that may be a
  sales_invoice, purchase_invoice or tax_declaration. The language has
  `ref<type>` only. Shipped as `document_type` + `document_id` strings;
  whether the language needs `ref<a|b>` is parked (SPEC §7 material).
- **Gapless statutory numbering** (A1/9/2026 → A1/10/2026): already named a
  domain invariant in SPEC §3; evidence confirms it must be kernel-level
  sequence state, not rule output.
- **Time in the log**: every recurring behavior observed (schedules, cyclic
  invoices, amortization, aging) reduces to rules matching injected time
  events. Confirms the M5 design direction; nothing observed needs more.
- **Multi-currency accounts** (PLN/EUR/USD): money stays int64 minor units +
  currency field; FX for reporting already exists (fx_rate). No change
  needed, recorded as confirmation.

## Deliberately not taken

- Screen layouts, navigation trees, tab structures — the shell's notions
  render these; copying them would be the bug CLAUDE.md warns about.
- Payment execution (koszyk płatności, zapłata z rachunku) — moving money is
  the bank's side of the boundary; Rhea records settlements as events.
- Certificates/identity plumbing (KSeF cert upload, invite accountant,
  roles) — out of experiment scope; actor attribution exists, authz doesn't.
- OCR mechanics — intake is an adapter emitting raw document events; the
  worklist already is the "do zatwierdzenia" inbox.
