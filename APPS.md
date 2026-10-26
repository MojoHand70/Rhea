# APPS.md — every tab an App

Step 1 of "Rhea Apps and business packs" (DIRECTION, KK 2026-10-23): every tab an App. First what
an App definition is and holds; then the register — the Apps Rhea already has without saying so:
every tab, sheet and screen the face opens, and the definitions behind it, placed in an App; for
each, its layer — what never changes (core), what the law imposes (the market pack), what the
business chooses (business packs), and the platform that carries them all — what it shows, does,
keeps and books, and where its knowledge lives today. It is the map step 1 builds from, and the
list of what still has to move out of the face.

Counted from the definitions as they stand (finance base v1–v9, the common packs, the Poland pack
v44): 54 types, 54 views, 157 rules, 53 doors; and about thirty kinds of screen the face draws from its own
code.

## The App definition

*KK, 2026-10-26 (DECISIONS): an App is one definition holding its parts, which the kernel unpacks
once at approval into the pieces it runs; an App may have several tabs; tabs inside a desk tab go
one level deep, never deeper. The rest of this section is Claude's proposal; where it is a choice,
it says so.*

**On the desk.** An App opens as one desk tab; its tabs run along the top inside it, and it opens on
the first. One record — an invoice, a bank line — opens in a desk tab of its own, as today, so two
can sit side by side; its sections are tabs inside it. Those are the only two places tabs nest, and
never a level deeper. A person may pull an App's tab out into a desk tab of its own, to set it beside
another; that is their arrangement, the App does not change. On the phone the same tabs are a row
along the top of the sheet. Odoo draws an invoice the same way: a list, and one invoice's form with
its own tabs.

**The file.** One file per App, in its pack: `packs/<pack>/apps/<app>.json`, in the JSON packs use
today. A query longer than a line sits beside it in a `.sql` file named after its tab, so a person
can read it (Claude's choice). The pack file keeps what is the pack's as a whole — what it is, what
it needs, its history and warrant, the facts it brings (master data: the VAT rates, the chart, the
roles) — and lists its Apps in order. The channels' files (KSeF in, MT940, NBP, the PDF reader) stay
as they are until step 5 makes them connectors.

**What it holds.** Step 1 fills the first group. The parts of later steps are in the format from the
start and stay empty until their step.

| Part | What it is | Unpacks into |
|---|---|---|
| `app`, `version`, `title` | its name for good, its version, its title in the pack's language (another pack's words may rename it, as they rename screens today) | the App kind (new) |
| `layer` | core, law, business or platform | — |
| `data` | the kinds of record it keeps, each with its fields and today's states | object types |
| `uses` | the kinds of record of other Apps it reads or its rules write | — (checked: each is some App's `data`) |
| `tabs` | its tabs, in order; the first opens it | views |
| `record` | what one of its records shows, as tabs | detail views |
| `actions` | what a person or a channel does with it | doors |
| `rules` | what makes and changes its records and books its facts — a booking rule belongs to the App whose fact it books, not to the journal it writes | rules |
| `variables` | step 2 | |
| `workflow` | step 3; until then a record's states stay on its kind of record, as today | |
| `triggers`, `autonomy` | step 4 | |
| `connectors` | step 5 | |

**Every part keeps what it has today**, word for word: its id, its version, its syntax — a view's
notion and spec, a rule's match and effect, a door's inputs. The App adds what holds them together
and invents no language. So nothing that names a part breaks: the facts, the provenance on every
value (fact, rule, rule version), the rights (`view:*`, `door:*`), a person's saved desk.

**A tab the desk draws from its own code** is written `{ "desk": "<name>" }` — the flow of sales
invoices, the document sheet, Dziś's cards — until step 2 gives it a definition. The register below
and "What the face knows that is not its own" list them.

**One owner.** Every kind of record, view, door and rule of a pack is in exactly one App of that
pack; a check fails on a part in none or in two.

**Across packs** (Claude's choice). A pack may bring a later version of another pack's App, whole, as
it brings later versions of a kind of record today. The core has Faktury zakupu: purchase invoices
at version 1 and their list. The Poland pack brings Faktury zakupu version 2: purchase invoices at
version 6, the Polish rules and doors, and the core's list carried unchanged. When packs install per
company, a company sees the version its packs bring — the same rule as for a kind of record's
versions, settled for both at the split.

**Where it sits in the drawer** (Claude's choice): from its first tab, as the drawer files screens
today — its domain the activity, its kind the folder (Dokumenty i zapisy, Raporty, Kartoteki). A
tab that is a briefing is also listed under Dziś, opening its App on that tab. No field of its own:
the category and folder of KK's sketch are read from the first tab rather than written twice.

**Who may run it.** In step 1 a person sees an App when they may see its first tab, and within it
the tabs they may see. A right on the App as a whole (`app:<id>` in a role) comes after step 1,
approved like any change to the roles.

**Approval.** An App is drafted and approved with its pack, in one bundle. The kernel checks it
before anyone approves — every part valid, one owner, every `uses` found — stores it append-only, a
new version a new row (invariant 1), and unpacks it once. Nothing is interpreted per click, so a
screen opens as fast as today.

### Terminy podatkowe — stateless, one tab

```json
{
  "app": "terminy-podatkowe", "version": 1, "title": "Terminy podatkowe", "layer": "law",
  "tabs": [
    { "view_id": "kalendarz-podatkowy", "version": 2, "notion": "analysis",
      "title": "Terminy podatkowe", "domain": "finance", "function": "briefing",
      "spec": {
        "sql_file": "kalendarz-podatkowy.sql",
        "columns": [
          { "field": "due",    "label": "Termin",   "type": "date" },
          { "field": "what",   "label": "Co",       "type": "string" },
          { "field": "period", "label": "Za okres", "type": "string" },
          { "field": "amount", "label": "Kwota",    "type": "money" },
          { "field": "note",   "label": "Uwaga",    "type": "string" } ] } } ]
}
```

The loader puts the `.sql` file's text where `sql_file` stands, so the view it unpacks is today's.
A briefing, so it sits under Dziś. Step 2 adds, for this one, the day as a variable a person may
set; today it is `{{today}}` inside the query.

### Faktury sprzedaży — several tabs, a record, rules

```json
{
  "app": "faktury-sprzedazy", "version": 2, "title": "Faktury sprzedaży", "layer": "law",
  "data": [
    { "name": "sales_invoice", "version": 8, "domain": "finance", "is_document": true,
      "label_field": "number",
      "lifecycle": { "field": "status", "transitions": { "open": ["open", "paid"], "paid": ["open"] } },
      "amendable": ["open"], "fields": ["…its 19 fields, as today"] },
    { "name": "ksef_submission", "version": 3, "…": "as today" } ],
  "uses": ["company", "vat_rate", "account", "posting"],
  "tabs": [
    { "view_id": "sales-invoice-list", "version": 1, "notion": "list", "…": "as today" },
    { "desk": "flow", "title": "Przepływ" },
    { "view_id": "naleznosci-do-pilnowania", "version": 1, "notion": "queue", "function": "briefing", "…": "as today" },
    { "view_id": "ksef-submissions", "version": 2, "notion": "list", "…": "as today" } ],
  "record": {
    "of": "sales_invoice",
    "tabs": [
      { "desk": "document-sheet", "title": "Faktura" },
      { "view_id": "sales-invoice-detail", "version": 1, "notion": "detail", "…": "as today" } ] },
  "actions": [ { "name": "ksef", "…": "as today: the KSeF channel out" } ],
  "rules": [ { "rule_id": "pl-book-sales-invoice", "…": "as today" },
             "…nine more: the postings by kind of sale, with VAT and without, and the KSeF answer and its task" ]
}
```

Version 2: the core's version 1 holds sales invoices at version 3 and nothing else. It opens on the
list, under Księgowość → Dokumenty i zapisy; Należności do pilnowania is also listed under Dziś. The
rule settling an invoice against a payment books the bank's fact, so it is Wyciągi i rozliczenia's.
The desk does not read the pack's detail view today: it draws one invoice from its own code (the
document sheet, 288 lines). Step 1 keeps both; step 2 moves what the sheet knows into the record's
tabs, and the sheet goes.

## The register

Its columns are the definition's parts: *shows* is `tabs` and `record`, *acts* is `actions`, *keeps*
is `data`, *books* is `rules`.

**Core** — what never changes: the kernel and its base vocabulary.

| App | shows | acts | keeps | books | knowledge today |
|---|---|---|---|---|---|
| Dziennik | journal, posting-detail | — | posting | — (each booking rule is its App's) | data |
| Plan kont | account-list, account-detail; *face: the account's ledger sheet* | set_tax_treatment | account | 2 | data; the ledger sheet is the core's own view |
| Zestawienie obrotów i sald | trial-balance, group-trial-balance | — | — | — | data |
| Zamknięcie miesiąca | *face: Dziś's month-end card* | close_period | period_lock | 1 | **face code**: which month, its VAT figures |
| Zamknięcie roku | — | close_year | year_close_clearing, _cost, _revenue, _tax | 4 | data |

**Law** — the Poland pack.

| App | shows | acts | keeps | books | knowledge today |
|---|---|---|---|---|---|
| Faktury sprzedaży | sales-invoice-list, sales-invoice-detail, naleznosci-do-pilnowania; *face: Przepływ, the invoice sheet* | the KSeF channel out | sales_invoice, ksef_submission | ~12 | **face code**: the flow's four states, the sheet's layout |
| Faktury zakupu | purchase-invoice-list; *face: the invoice sheet* | the KSeF channel in, set_cost_account | purchase_invoice | ~5 | **face code**: the sheet's cost-account action |
| Wyciągi i rozliczenia | wplaty-do-rozliczenia, wyplaty-do-rozliczenia; *face: Dziś's bank-line card* | classify_bank_operation, approve_statement, the MT940 reader | bank_operation, settlement, matching_difference | ~40 | **face code**: the classification card |
| VAT | rejestr-vat-sprzedazy, rejestr-vat-zakupu, rozliczenie-vat, deklaracja-vat; *face: the register-cell sheet* | close_period, the JPK_V7M writer | vat_settlement | 1 | data; the cell sheet is generic "why?" |
| CIT | zaliczki-cit, zaliczka-cit, podatek-odroczony | declare_cit_status, recompute_cit_advance, recognise/reverse_deferred_tax | cit_status, cit_limit, cit_advance, deferred_tax | ~6 | data |
| Środki trwałe | srodki-trwale, amortyzacja | accept_fixed_asset, dispose_fixed_asset | fixed_asset, depreciation_charge | ~6 | data |
| Prace rozwojowe | prace-rozwojowe | capitalise/amortise_development_costs | development_costs | 3 | data |
| Rozliczenia międzyokresowe | rmk; *face: "Rozłóż w czasie" on the purchase sheet* | spread_cost | prepaid_cost, prepaid_release | 3 | **face code**: the spreading form |
| Rezerwy i odpisy | rezerwy, odpisy-naleznosci | create/use/release_provision, write_down_receivable, settle_write_down, write_down/reverse_inventory_write_down | provision, receivable_write_down, inventory_write_down | ~12 | data |
| Pożyczki | pozyczki | grant_loan, receive_loan, repay_loan, charge_loan_interest | loan | ~10 | data |
| Waluty | fx-rates, wycena-bilansowa-walut | revalue_fx_balances, the NBP channel | fx_rate, fx_revaluation | 1 | data |
| Kapitały i wynik | kapitaly, podzial-wyniku, korekty-zasad, wycena-bez-kontynuacji | distribute_result, adjust_for_policy_change, measure_without_going_concern | result_distribution, policy_adjustment, liquidation_measurement | ~6 | data |
| Zaliczki na środki trwałe | — | pay_asset_advance, settle_asset_advance | asset_advance | 3 | data |
| Kontrahenci i moja firma | company-list, company-detail; *face: Firma (the company page)* | enter_company, identify_company, declare_company_form, declare_company_accounts | company | ~12 | **face code**: the company page, built from invoices |
| Stawki VAT | vat-rate-list | — | vat_rate | 1 | data |
| Terminy podatkowe | kalendarz-podatkowy | — | — | — | data |
| Raporty | wynik, wynik-miesiace, naleznosci-wiek, invoice-by-customer, intercompany-positions; *face: the result sheet, the receivables sheet* | — | — | — | **face code**: the result sheet (330 lines composing the P&L) |

**Platform** — the chassis every App runs on.

| App | shows | acts | keeps | knowledge today |
|---|---|---|---|---|
| Dziś | *face*: decisions, her morning, the figures | — | — | **face code**: each decision card knows its App's business |
| Poranek | poranek, poranek-ksiegowosc (narratives) and the facts they speak from (rano, poranek-*) | — | — | data |
| Zadania | do-zrobienia | — | task | data |
| Ludzie | who-can-do-what; *face*: Kto ma dostęp, Zaproszenia, Moje klucze, Zaproś | the people pack's doors | person, role, membership, mandate, substitution, group | data + generic face |
| Firmy | *face*: the switch, Dodaj klienta, Załóż firmę | grant_mandate, grant_membership, enter_company | — | generic face |
| Dokumenty | issued-documents | receive_document, issue_document, the readers and writers | received_document, issued_document | data |
| Dlaczego | *face*: a field's origin, a rule, a chain of facts | — | — | generic face |
| Rozmowa | *face*: her conversation, kept answers and reports | — | — | generic face |
| Szukaj, Ustawienia, Galeria | *face* | — | — | the search box knows invoices, bank lines and accounts by name |

**Business** — none yet. The first business pack (Rhea for Accounting Offices) is step 7.

(The finance base's generic `invoice` — invoice-list, invoice-detail, invoice-by-customer — is the
founding milestone's, kept for its story; the Poland pack's invoices supersede it in use.)

## What the face knows that is not its own

Each moves out — law to the pack, business to an App — when the step that gives it a home lands;
the face's check and the live walk guard what the person sees.

1. **Przepływ faktur sprzedaży** — the four states and what each means, written in the desk's code.
   The sales-invoice App's workflow, step 3.
2. **Dziś's decision cards** — the month-end card (which month, its VAT), the bank-line
   classification card, the approval card: each is an App's proposal. Dziś becomes the place every
   App's *propose* steps appear, drawn generically — steps 1 and 4.
3. **The document sheet** (288 lines) — the invoice's layout, the bank line, "Rozłóż w czasie", the
   cost account: the Apps' detail views and actions, step 2.
4. **The result and receivables sheets** — statement views in the Poland pack, step 2.
5. **The company page** — the company App's view over its documents, step 2.
6. **The search box** — each App says what of it is searchable, step 1.
7. **The starter questions and the stand-in domain words** — the pack's words, step 1.

## Order within step 1

*Proposed by Claude, 2026-10-26; each step on KK's go.*

1. **The format** — "The App definition" above; KK's three answers in DECISIONS, 2026-10-26.
2. **The kernel.** The App as a definition kind: stored append-only, a new version a new row,
   drafted and approved with its pack in one bundle. The loader reads a pack's App files and unpacks
   each into today's pieces. The checks: every part in exactly one App of its pack, every `uses`
   found, a tab naming a view or a desk screen. `/api/apps`, served through what the company is
   offered and walked with two companies, like every read. The test: each pack, converted,
   unpacks into exactly the definitions approved today — the same ids, versions and content — so
   a database built as white01 drafts nothing, and replay reproduces identical state.
3. **The packs rewritten as App files** — the core, the common packs, Poland, Germany, the US — by
   a script, held by step 2's test, following the register above. Where a row's boundary is
   unclear, the ownership rule decides and the row is corrected.
4. **The desk.** The drawer lists Apps; an App opens as one desk tab with its tabs, a record in its
   own with its tabs, one level; a tab pulled out into a desk tab of its own; a desk saved before
   still opens, a view's id leading to its App; an App's page of what it consists of — the
   designer page's first, read-only form. Gallery first, then the face's check and the live walk.
5. **The platform's Apps** — Dziś, Ludzie, Dokumenty, Dlaczego, Rozmowa, Szukaj, Ustawienia — as
   definitions in the core, their tabs the desk's own; what is searchable and the starter words
   move to the packs (items 6–7 above).

**Before the split, for KK.** "Core, market packs and business packs" (DIRECTION) splits the Poland
pack into Podstawa, Pełna księgowość and Środki trwałe — Polska, then adds KPiR and ryczałt.
Proposed: rewrite the packs as Apps first, so the split moves whole Apps between packs rather than
loose parts. The split's own steps stand otherwise.
