# APPS.md — every tab an App

Step 1 of "Rhea Apps and business packs" (DIRECTION, KK 2026-10-23): the Apps Rhea already has
without saying so. Every tab, sheet and screen the face opens, and the definitions behind it, is
grouped into an App; for each, its layer — what never changes (core), what the law imposes (the
market pack), what the business chooses (business packs), and the platform that carries them all —
what it shows, does, keeps and books, and where its knowledge lives today. It is the map step 1
builds from, and the list of what still has to move out of the face.

Counted from the definitions as they stand (finance base v1–v9, the common packs, the Poland pack
v44): 54 types, 54 views, 157 rules, 53 doors; and about thirty kinds of screen the face draws from its own
code.

## The App, as step 1 declares it

Grouping only — nothing behaves differently. An App definition names:

- **app** and **version**; **title** from the pack's words; **folder** — the pack's domain, then
  the drawer's kind (Dokumenty i zapisy, Raporty, Kartoteki);
- **layer** — core, law, business or platform;
- **shows** — its views, the first one its tab; while the face still draws a screen from its own
  code, that screen, by name;
- **acts** — the doors it offers; **keeps** — the types it owns; **books** — the rules that make
  and change its types and book its facts (a booking rule belongs to the App whose fact it books,
  not to the journal it writes);
- **run** — who may run it: in step 1 the right to what it shows; a right on the App itself later.

Stored as a definition kind of its own, versioned and approved with its pack; served at
`/api/apps`; the drawer lists Apps; an App opens on its first view and has a page showing what it
consists of — the designer page's first, read-only form. Variables, workflow, triggers and autonomy
are steps 2–4.

## The register

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

1. This register, and its App boundaries agreed.
2. The kernel: the App as a definition kind — stored, versioned, approved with its pack — and
   `/api/apps`.
3. The packs declare their Apps (the rows above as data); the face's drawer from Apps; an App's
   page of what it consists of.
4. The platform's Apps declared by the face in one place, so every tab is an App; the searchable
   and the words (items 6–7 above) move to the packs.
