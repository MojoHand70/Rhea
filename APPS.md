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

**A tab of another pack's screen** is written by its name alone, `{ "view": "account-list" }`: the
screen stays that pack's. It is how a later version of another pack's App shows what that pack
defines (Claude's choice, 2026-10-26).

**A tab the desk draws from its own code** is written `{ "desk": "<name>" }` — the flow of sales
invoices, the document sheet, Dziś's cards — until step 2 gives it a definition. The register below
and "What the face knows that is not its own" list them.

**One owner.** Every kind of record, view, door and rule of a pack is in exactly one App of that
pack; a check fails on a part in none or in two.

**Across packs** (Claude's choice). A pack may bring a later version of another pack's App, whole, as
it brings later versions of a kind of record today. The core has Faktury zakupu: purchase invoices
at version 1 and their list. The Poland pack brings Faktury zakupu version 2: purchase invoices at
version 6, the Polish rules and doors, and the core's list named as its tab. When packs install per
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
  "app": "tax-calendar", "version": 1, "title": "Terminy podatkowe", "layer": "law",
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
  "app": "sales-invoices", "version": 2, "title": "Faktury sprzedaży", "layer": "law",
  "data": [
    { "name": "sales_invoice", "version": 8, "domain": "finance", "is_document": true,
      "label_field": "number",
      "lifecycle": { "field": "status", "transitions": { "open": ["open", "paid"], "paid": ["open"] } },
      "amendable": ["open"], "fields": ["…its 19 fields, as today"] },
    { "name": "ksef_submission", "version": 3, "…": "as today" },
    { "name": "recurring_invoice_template", "version": 1, "…": "as today" } ],
  "uses": ["posting", "task"],
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
             "…eleven more: the postings by kind of sale, with VAT and without, the cost of goods and materials sold, the KSeF answer and its task" ]
}
```

Version 2: the core's version 1 holds sales invoices at version 3 and nothing else. It opens on the
list, under Księgowość → Dokumenty i zapisy; Należności do pilnowania is also listed under Dziś. The
rule settling an invoice against a payment books the bank's fact, so it is Rachunki i karty's. Its
uses are what its rules write that another App keeps: the journal's postings, and the task a KSeF
refusal opens.
The desk does not read the pack's detail view today: it draws one invoice from its own code (the
document sheet, 288 lines). Step 1 keeps both; step 2 moves what the sheet knows into the record's
tabs, and the sheet goes.

## The register

The Apps the packs are written as (2026-10-26): every kind of record, view, door and rule of a pack
in exactly one of its Apps, generated from the App files. A pack that brings a later version of
another pack's App names what that pack defines — a screen as a tab by its name (*named* below), a
kind of record in its uses — so each pack declares exactly what it declared before: Okresy's Period
close names the core's month lock, the Poland pack's Plan kont, Kontrahenci, Faktury zakupu, Stawki
VAT and Waluty name the core's screens and kinds. The channels' packs — KSeF in, MT940, NBP, the PDF
reader — have doors and no tabs and stay as they are until step 5 makes them connectors.

**Core** — `packs/core/core.json` v3, 13 Apps.

| App | title | v | layer | tabs | record | keeps | uses | doors | rules |
|---|---|---|---|---|---|---|---|---|---|
| `invoices` | Invoices | 1 | core | invoice-list, invoice-by-customer | invoice: invoice-detail | invoice | — | — | — |
| `parties` | Companies | 1 | core | company-list | company: company-detail | company, contractor | — | — | — |
| `chart-of-accounts` | Chart of accounts | 1 | core | account-list | account: account-detail | account | — | — | — |
| `journal` | Journal | 1 | core | journal | posting: posting-detail | posting | — | — | — |
| `trial-balance` | Trial balance | 1 | core | trial-balance, group-trial-balance, intercompany-positions | — | — | — | — | — |
| `period-close` | Period close | 1 | core | — | — | period_lock | — | — | — |
| `vat-rates` | VAT rates | 1 | core | — | — | vat_rate | — | — | — |
| `fx-rates` | FX rates | 1 | core | fx-rates | — | fx_rate | — | — | — |
| `sales-invoices` | Sales invoices | 1 | core | — | — | sales_invoice | — | — | — |
| `purchase-invoices` | Purchase invoices | 1 | core | purchase-invoice-list | — | purchase_invoice | — | — | — |
| `bank-accounts` | Bank accounts and cards | 1 | core | — | — | bank_account, bank_operation, settlement, matching_difference | — | — | — |
| `tasks` | Tasks | 1 | platform | do-zrobienia | — | task, task_schedule | — | — | — |
| `packs` | Packs | 1 | platform | *desk: packs* | — | installed_pack | — | install_pack, switch_off_pack | 2 |

**Ludzie** — `packs/common/people.json` v4, 1 App.

| App | title | v | layer | tabs | record | keeps | uses | doors | rules |
|---|---|---|---|---|---|---|---|---|---|
| `people` | People | 1 | platform | *desk: access*, who-can-do-what, *desk: invites* | — | person, role, membership, mandate, substitution, group | — | register_person, appoint_deputy, set_person_status, identify_person, declare_role, change_role, grant_membership, end_membership, grant_mandate, end_mandate, appoint_substitute, end_substitution, declare_group | 15 |

**Okresy** — `packs/common/periods.json` v3, 1 App.

| App | title | v | layer | tabs | record | keeps | uses | doors | rules |
|---|---|---|---|---|---|---|---|---|---|
| `period-close` | Period close | 2 | core | — | — | — | period_lock | close_period | 1 |

**Dokumenty** — `packs/common/documents.json` v5, 1 App.

| App | title | v | layer | tabs | record | keeps | uses | doors | rules |
|---|---|---|---|---|---|---|---|---|---|
| `documents` | Documents | 1 | platform | issued-documents | — | received_document, issued_document | — | — | 3 |

**Poland** — `packs/pl/pack.json` v46, 24 Apps.

| App | title | v | layer | tabs | record | keeps | uses | doors | rules |
|---|---|---|---|---|---|---|---|---|---|
| `chart-of-accounts` | Plan kont | 2 | core | account-list *(named)* | account: account-detail *(named)*, *desk: ledger-sheet* | account | — | set_tax_treatment | 2 |
| `vat-rates` | Stawki VAT | 2 | law | vat-rate-list | — | — | vat_rate | — | 1 |
| `fx-rates` | Waluty | 2 | law | fx-rates *(named)*, wycena-bilansowa-walut | — | fx_revaluation | fx_rate, posting | revalue_fx_balances | 5 |
| `parties` | Kontrahenci i moja firma | 2 | law | company-list *(named)* | company: company-detail *(named)*, *desk: company-page* | company, contractor | — | identify_company, declare_company_form, declare_company_accounts, enter_company | 8 |
| `sales-invoices` | Faktury sprzedaży | 2 | law | sales-invoice-list, *desk: flow*, naleznosci-do-pilnowania, ksef-submissions | sales_invoice: *desk: document-sheet*, sales-invoice-detail | sales_invoice, ksef_submission, recurring_invoice_template | posting, task | ksef | 12 |
| `purchase-invoices` | Faktury zakupu | 2 | law | purchase-invoice-list *(named)* | purchase_invoice: *desk: document-sheet* | purchase_invoice | posting, company | set_cost_account | 6 |
| `bank-accounts` | Rachunki i karty | 2 | law | wplaty-do-rozliczenia, wyplaty-do-rozliczenia | bank_operation: *desk: document-sheet* | bank_account, bank_operation, settlement, matching_difference | posting, sales_invoice, purchase_invoice | classify_bank_operation | 19 |
| `vat` | VAT | 1 | law | rozliczenie-vat, rejestr-vat-sprzedazy, rejestr-vat-zakupu | vat_settlement: deklaracja-vat | vat_settlement | — | — | 1 |
| `cit` | CIT | 1 | law | zaliczki-cit, podatek-odroczony | cit_advance: zaliczka-cit | cit_limit, cit_status, cit_advance, deferred_tax | posting | declare_cit_status, recompute_cit_advance, recognise_deferred_tax, reverse_deferred_tax | 12 |
| `tax-declarations` | Deklaracje podatkowe | 1 | law | — | — | tax_declaration | — | — | — |
| `tax-calendar` | Terminy podatkowe | 1 | law | kalendarz-podatkowy | — | — | — | — | — |
| `fixed-assets` | Środki trwałe | 1 | law | srodki-trwale, amortyzacja | — | fixed_asset, depreciation_charge, asset_advance | posting | accept_fixed_asset, dispose_fixed_asset, move_from_construction, pay_asset_advance, settle_asset_advance | 13 |
| `vehicles` | Samochody i leasing | 1 | law | — | — | vehicle, mileage_entry, lease_contract | — | — | — |
| `development-costs` | Prace rozwojowe | 1 | law | prace-rozwojowe | — | development_costs | posting | capitalise_development_costs, amortise_development_costs | 4 |
| `prepaid-costs` | Rozliczenia międzyokresowe | 1 | law | rmk | — | prepaid_cost, prepaid_release | posting | spread_cost | 5 |
| `provisions` | Rezerwy i odpisy | 1 | law | rezerwy, odpisy-naleznosci | — | provision, receivable_write_down, inventory_write_down | posting | create_provision, use_provision, release_provision, write_down_receivable, settle_write_down, write_down_inventory, reverse_inventory_write_down | 19 |
| `stocktake` | Inwentaryzacja | 1 | law | — | — | stocktake | — | — | — |
| `loans` | Pożyczki | 1 | law | pozyczki | — | loan | posting | receive_loan, grant_loan, repay_loan, charge_loan_interest | 12 |
| `equity` | Kapitały i wynik | 1 | law | kapitaly, podzial-wyniku, korekty-zasad, wycena-bez-kontynuacji | — | result_distribution, policy_adjustment, liquidation_measurement | posting | approve_statement, distribute_result, pay_dividend, adjust_for_policy_change, measure_without_going_concern, provide_for_cessation | 13 |
| `year-close` | Zamknięcie roku | 1 | law | — | — | year_close_cost, year_close_revenue, year_close_tax, year_close_clearing | posting | close_year | 6 |
| `reports` | Raporty | 1 | law | wynik, wynik-miesiace, naleznosci-wiek | — | — | — | — | — |
| `tasks` | Zadania | 2 | platform | do-zrobienia | — | task, task_schedule | — | — | — |
| `morning` | Poranek | 1 | platform | poranek, poranek-ksiegowosc, rano, poranek-wynik, poranek-pieniadze, poranek-otwarte | — | — | — | — | — |
| `words` | Słownik | 1 | platform | slownik | — | — | — | — | — |

**Poland — warehouse** — `packs/pl/warehouse.json` v2, 1 App.

| App | title | v | layer | tabs | record | keeps | uses | doors | rules |
|---|---|---|---|---|---|---|---|---|---|
| `goods-receipts` | Przyjęcia zewnętrzne (PZ) | 1 | law | pz-list | — | pz, pz_line | stock_movement, posting | — | 4 |

**Germany** — `packs/de/pack.json` v4, 2 Apps.

| App | title | v | layer | tabs | record | keeps | uses | doors | rules |
|---|---|---|---|---|---|---|---|---|---|
| `chart-of-accounts-skr03` | Kontenrahmen SKR03 | 1 | law | — | — | — | account, vat_rate | — | 2 |
| `sales-invoices-de` | Ausgangsrechnungen | 1 | law | rechnungen-list, ust-je-monat | sales_invoice: rechnung-detail | — | sales_invoice, posting | — | 2 |
Where the first register (2026-10-23, by reading) and the packs disagreed, the packs decided:
approving a statement is the year's financial statement and lives in Kapitały i wynik, not with the
bank's statements; advances on fixed assets are part of Środki trwałe; the kinds of record no rule
or screen uses yet have small Apps of their own (Samochody i leasing, Inwentaryzacja, Deklaracje
podatkowe) so that each still has its list; the result sheet and the receivables sheet open from a
figure, not as a tab, and wait for step 2 as screens.

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
2. **The kernel** — *done 2026-10-26* (7d29c10, cf281df, 3826a0e, 18cd638, 4d23690, 2766677). The App as a definition kind: stored append-only, a new version a new row,
   drafted and approved with its pack in one bundle. The loader reads a pack's App files and unpacks
   each into today's pieces. The checks: every part in exactly one App of its pack, every `uses`
   found, a tab naming a view or a desk screen. `/api/apps`, served through what the company is
   offered and walked with two companies, like every read. The test: each pack, converted,
   unpacks into exactly the definitions approved today — the same ids, versions and content — so
   a database built as white01 drafts nothing, and replay reproduces identical state.
3. **The packs rewritten as App files** — *done 2026-10-26*: the core, the common packs, Poland and
   its warehouse pack, Germany (the US has no pack file yet) — by
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
