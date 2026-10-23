# Installable packs — where the code stands

Review of 2026-10-23, read-only, at Rhea `d2b1f5f` and rheaface `ebcfb4f`, before any pack is
separated or built (KK: "Shouldn't we review our code first to align with installable packs?").
Three readers: the kernel; the face; the packs and their tests. The findings called defects in §0
were read again in the code before being written here. Sizes: S small, M medium, L large. Steps
refer to the revised order in §6.

## In one paragraph

Today a pack is installed for the whole installation: approving a pack's bundle activates its
rules, types, views and doors for every company. The only per-company definition the kernel knows
is the company rule (`rule.company`, `forCompany`, 2026-10-23). Record types, views, doors, the
chart of accounts and pack versions exist once, at one version, for everyone; nothing records which
pack a definition came from, or which packs a company runs. The face trusts what the kernel lists
and in a dozen places reaches for full-books screens and doors by name. So installing per company
is the foundation, not a settings tab, and it comes before the split. The split is blocked by six
links from the Polish base back into full books. KPiR needs one kernel decision about what a KPiR
entry is. The core vocabulary is shipped by nothing in the kernel.

## 0. Fix now — defects today, with or without packs

1. **Her instructions name Helios for every company.** rheaface `web/app.js:1190` ("…a small
   Polish company, Helios Studio sp. z o.o.") and `:1204` ("Months with documents: 2026-01 to
   2026-09"). On white01 a client's question is understood as Helios's. The company's name and its
   months come from the loaded world. S.
2. **Rules and bundles can be written for everyone from a company's session.**
   `internal/store/schema.sql:292` and `:314`: the policies `rule_scope` and `bundle_scope` have
   `USING (rhea_in_scope(company))` and no `WITH CHECK`, so Postgres checks new rows with
   `rhea_in_scope`, which is true for a row with no company. A scoped session can insert a
   Rhea-wide draft rule or bundle. `event` and `object` check writes with `rhea_may_write`
   (`schema.sql:278–281`) since the "office acting for no client" fix. Only drafts are reachable
   (activating Rhea's rules is the house's), still invariant 9. `WITH CHECK
   (rhea_may_write(company))` on both, tested from a company scope. S.
3. **An office's shortcuts cross its clients.** `web/app.js:71–75`: asks, pins and declined asks
   are kept under `rhea.face.<name>` with no company, while the desk keys its arrangement and
   threads per company (`desk.js:57–58`). A shortcut naming client A's customer shows on client B's
   Dziś and computes on B's books. Keyed by company as the desk does. S.

## 1. Installing per company — the foundation

**What exists to build on.** Company rules (`forCompany` `internal/exec/exec.go:519–532`, the
triggers `rule_company_fixed` and `rule_replaces_rheas` `schema.sql:341–379`, `companysOwn`
`exec/activity.go:175–206`); bundle approval and its card (`exec/activity.go:580–659`; face
`app.js:1543–1552`); `SimulateBundle` running in the caller's scope (`exec/simulate.go:87`), which
can show what installing adds; `Replay` re-applying baked derived events (`exec.go:1138–1190`), so
an install choice cannot break determinism; objects already record `type_version`.

**What is missing (kernel).**

- **1.1 Install as a fact.** Nothing records "company C installs pack P at version N from day D"
  (`Companies()` is `SELECT DISTINCT company FROM event`, `store.go:1067`). Install, upgrade and
  switch-off become dated, company-stamped events: they replay and they explain. M.
- **1.2 A pack that says what it is.** `Manifest` (`internal/pack/pack.go:30–51`) has no kind,
  market, title, needs or Apps; `requires` lists object types only and checks they exist
  (`207–211`). A version's full member list is stored nowhere: upgrade bundles carry only what
  changed. The manifest gains kind / market / needs; a pack-version manifest lists its members with
  their versions. M.
- **1.3 A catalog of available packs.** Packs are files on the server loaded from the CLI
  (`cmd/rhea/main.go:162–188`); there is no API, so Pakiety's "Dostępne" column has no source. M.
- **1.4 Two approvals.** Today the house's approval of `pack-<p>-v<n>` activates it for everyone
  (`pack.go:302–372`, `exec/activity.go:580–659`). The house publishes a version into the catalog;
  a company's owner approves installing it (a proposal on Dziś). An install is a fact, not a bundle
  of definitions, so `companysOwn`'s refusal of shared types above v1 does not stand in its way. M.
- **1.5 One catalog per company.** `/api/nav` (`internal/shell/shell.go:413–451`), `/api/types`
  (`language.go:73–140`), `/api/activities` (`activities.go:33–74`), words (`answer.go:66–82`),
  Rhea's changes (`rheas_changes.sql`), the author's catalog (`draftask.go:23–27` — `ActiveRules`
  without `forCompany`, so even replaced rules show), matching (`matching.go:54`), explain,
  narrative, rights: about eighteen call sites read definitions installation-wide (`reads.go:154`
  says so). One seam, `Catalog(ctx, company)`, and every site through it. M.
- **1.6 Rules selected by what the company installed.** `forCompany` gives Rhea's rules plus the
  company's own; a pack's rules are Rhea's, so they fire for every company. Markets are kept apart
  today by a market binding on 18 entry rules (invoices, bank line, accounts, VAT rates, CIT limit);
  the other 120 Polish rules follow from objects those create. That holds for Poland beside Germany
  now (`cohabitation_test.go`), and breaks once a second market creates a shared type (its bank
  lines would get Polish postings). `forCompany` also selects Rhea's rules by the company's installs
  at the event's date. L.
  *Not by copying pack rules into company rules:* `rule_company_fixed` would force a new id per
  company, `rule_replaces_rheas` would forbid a lesson replacing its own pack copy, `Promotable`
  (`promote.go:56–80`) would count pack copies as companies' own readings and propose them for
  promotion, and upgrades would lose their lineage. For DECISIONS: a pack's rule stays Rhea's, the
  install selects it; "the company's own" means a rule that runs for that company alone.
- **1.7 Offered before allowed.** `TriggerActivity` resolves the installation's door
  (`exec/activity.go:222`); `Authorize` checks roles and domain (`rights.go:455–545`). A door the
  company has not installed is refused before rights are asked, in the trigger, the door listing and
  the rights view (`rights.go:323`). S once 1.5 exists.
  In the face the purchase-invoice sheet offers "Inne konto kosztów" (`set_cost_account`, accounts
  `/^4/`) and "Rozłóż koszt w czasie (RMK)" (`spread_cost`, onto 640) on every purchase invoice
  (`app.js:486–487, 533–570`), and the face server opens both (`internal/face/server.go:43`).
  Harmless while every company keeps full books; for a KPiR company one tap books to 640 in books
  with no chart.
- **1.8 Installing mid-year books the past.** Backfill exists and the company may approve it itself
  (`companysOwn`), but `PlanBackfill` skips time events (`exec/backfill.go:110`): a module's monthly
  depreciation for past months would not be booked. M.
- **1.9 Readers per installed market.** `adapter.Run` fans out over every company
  (`internal/adapter/adapter.go:182–205`): KSeF and NBP passes run for companies without the Polish
  base. S.
- **1.10 Switching off has no date.** `rule.retired` (`exec/activity.go:455–478`) supersedes now,
  for every date; "stops from that day, what it booked stays" needs the install's end date read in
  `forCompany`. M.

**Several versions alive at once** (kept apart: decision D1). `object_type`, `view_def` and
`activity` carry no company and resolve to one newest active version for the installation
(`schema.sql:67–100`; `GetObjectType` `store.go:790`, `ListObjectTypes` `823`, `LatestViewDefs`
`880`, `active_activities.sql`); rules have one active version per id (`active_rules.sql`);
`matchRule` skips an event dated before the newest version's `effective_from` (`exec.go:544`) —
the US review's "rule versions lose earlier-dated history"; `Replaces` names an id, not a version
(`exec.go:522`), so a company's override keeps shadowing after Rhea's rule changes, with no warning.
A company staying on an older pack version needs all four. L.

**The face's half (rheaface).**

- *Views by role, not by id.* `LIVE_VIEWS` fetches twelve view ids on every load and refresh and
  turns a 404 into an empty view (`app.js:1741, 1760`): missing looks like empty. Columns are read
  by position (`app.js:195, 802, 817, 1033–1042, 1517, 1680`): a view whose columns are reordered
  shows the wrong column as the amount, with no error. `viewOf(notion, pred)` exists (`app.js:19`).
  M — and the split's "no change in behaviour" depends on it.
- *Doors from definitions.* The face server's allow-list (`server.go:43–49`) lacks
  `declare_company_form` and any KPiR or install door; every pack's door for people would need a
  face deploy. The server keeps stamping the person, signing and limiting size; the list comes from
  the doors the company is offered. M.
- *Refresh re-reads definitions when they change* (`app.js:1796` re-reads objects and views only);
  `trimActs` keeps `allowed` (`app.js:1755`). S.
- *Pakiety.* `settingsEl` (`desk.js:1183–1203`) gains a tab kind `{k:'packs'}`; the waiting card
  (`app.js:1543–1552`) becomes an install card (title, kind, version, needs, Apps, the company it
  installs into, Wyłącz); `mayDo` (`desk.js:831`) is the pattern for who may install;
  `bundleSource`'s regex (`app.js:1457–1461`) and "Nowa umiejętność" (`1545`) give way to the
  pack's own title; a read route for the catalog and a forwarded install route with the person
  stamped (`server.go:58–60`). M.
- *A company with nothing installed.* No card says so (DIRECTION requires one); the firm's name is
  blank (`company` is the Poland pack's type, `app.js:1508`); Odśwież with no morning behind it. The
  desk's fresh layouts are built from `trial-balance`, `rejestr-vat-sprzedazy`, `rozliczenie-vat`,
  `sales-invoice-list` and the `sales_invoice` flow (`desk.js:30–38`); the flow is offered in five
  places (`desk.js:588, 844, 1023, 1105, 1116`); the phone's Dokumenty is fixed to Sprzedaż / Zakupy
  / Bank (`app.js:1707–1716`); the Księgi menu lists only loaded views with rows
  (`desk.js:572–602`). Each S.
- *Onboarding with nothing installed.* `enter_company` is a Poland-pack door
  (`internal/face/onboard.go:19, 239–251`); the key "PL"+NIP and `country: "PL"` are written in
  (`onboard.go:41–56, 240`); the form records the legal form only — the VAT status read from the
  white list is shown, not recorded, and there is no choice of books or tax form
  (`app.js:2054–2071`). Order: the core company record, the market base, `declare_company_form`,
  then the regime proposed on Dziś. M.
- *Asking her.* `localMatch` assumes the year 2026 (`app.js:1217`) and ignores words taken from
  Helios's customers (`1218`); starter questions about Agencja Nova and August's VAT are shown to
  every company (`1386`); metrics are offered whatever is installed (`1193`). S.

## 2. The core, shipped by the kernel

- The founding vocabulary (`testdata/seed/finance*.json`: 9 types, 14 views) is loaded by nothing
  in the kernel. The face's build tool inserts its types and views straight into the store, active,
  without an approval (rheaface `tools/facebuild.go:43–57`). white01 got them from a restore of the
  Helios simulation; `deploy-white01.sh` loads only the three common packs and `pl/pack.json`
  (`112–113, 176`). A fresh server fails with "requires object type …"; `nbp.json` and
  `warehouse.json` are never deployed.
- It mixes core and law: `account`, `posting`, `period_lock`, `company` v1, `fx_rate` (the kernel
  names it: `exec/postings.go:33`, `core/vm.go`) and `invoice` sit beside `vat_rate`,
  `sales_invoice` and `purchase_invoice`.
- The common packs (people, periods, documents) are platform: rights need people's roles before
  anything is installed. They are core, for the whole installation, never installed per company.
- Rachunki i karty's types live in the Poland pack: `bank_account`, `bank_operation`, `settlement`,
  `matching_difference`, with `contractor`, `task`, `do-zrobienia`, `classify_bank_operation` and
  the rules `pl-company-account`, `pl-explain-bank-operation`, `pl-classify-bank-operation`. Core's
  `company` lacks the `iban` that `pl-company-account` sets (only the pack's `company` v7 has it).

A kernel-owned, embedded, versioned core (as `SeedActivities`, `store/activities.go:196`) loaded by
`rhea init`, with the core pieces moved out of the Poland pack under their present ids. M. This was
step 5 ("core cleanup"); it moves first, because "what a company with nothing installed has" must
exist before installing can mean anything.

## 3. Splitting the Poland pack

**What goes where** (v44 plus side files). POD = Polska — podstawa, PK = Pełna księgowość,
ŚT = Środki trwałe — Polska.

| kind (total) | POD | PK | ŚT | core | other |
|---|---|---|---|---|---|
| types (41) | 10 | 20 | 2 | 8 | 1 (Magazyn) |
| views (38) | 14 | 21 | 2 | 1 | – |
| rules (138) | 26 | 105 | 4 | 3 | – |
| doors (39) | 5 | 31 | 2 | 1 | – |
| master-data events (67) | 5 (VAT rates) | 62 (55 accounts, 6 tax treatments, 1 CIT limit) | – | – | – |

- **POD types:** `company` v7, `sales_invoice` v8, `purchase_invoice` v6, `ksef_submission`,
  `vat_settlement`, `tax_declaration`, `recurring_invoice_template`, and `vehicle`, `mileage_entry`,
  `lease_contract` (later Pojazdy). **Views:** vat-rate-list, sales-invoice-list/-detail,
  ksef-submissions, rejestr-vat-sprzedazy/-zakupu, rozliczenie-vat, deklaracja-vat,
  kalendarz-podatkowy, naleznosci-do-pilnowania, naleznosci-wiek, wplaty-/wyplaty-do-rozliczenia,
  slownik (to be split). **Doors:** ksef, identify_company, declare_company_form,
  declare_company_accounts, enter_company; side files ksef_inbound, bank_mt940, nbp, pdf_invoice.
- **PK:** CIT (`cit_limit`, `cit_status`, `cit_advance`, `deferred_tax`), RMK, year close, result
  distribution, equity, provisions, write-downs, loans, FX revaluation, assets in progress,
  development costs; the six mornings (they speak from postings in `pl-stat`, so each regime brings
  its own); 31 doors from `set_cost_account` to `amortise_development_costs`; 105 rules, of which 10
  post invoices, 7 post bank lines and differences, 4 post fixed-asset facts.
- **ŚT:** `fixed_asset`, `depreciation_charge`; srodki-trwale, amortyzacja; accept/dispose doors;
  4 register rules.
- **Core:** see §2. **Future Magazyn — Polska:** `stocktake` and `warehouse.json` (its posting rule
  `pl-pz-line-posts` to PK).
- **Words** (`slownik` v17): types 20 core / 11 POD / 20 PK / 2 ŚT / 1 Magazyn; all nine account
  groups PK; the Polish words for core types go with Podstawa.

**Links that point the wrong way — cut before the split.**

- *Podstawa → Pełna księgowość*
  - `pl-settle-vat` fires only on `period.closed` with `book = "pl-stat"`: a KPiR company would
    never get its VAT settled. VAT is settled when a month closes, whatever the books.
  - `kalendarz-podatkowy` reads `cit_advance` and filters on `pl-stat`; the calendar reads
    `tax_declaration` objects each regime fills (the type exists, unused).
  - Matching names accounts: `pl-match-incoming` (761, 750, 751), `pl-match-outgoing` (750/751),
    `pl-record-difference` (750/751/756/760); and the kernel checks the difference account against
    existing accounts (`exec/matching.go:579`; `core/matching.go:398, 429`). A difference carries
    its kind; the regime books it. Kernel change, S.
  - `company.cost_account` is a PK field on POD's type, with `set_cost_account`.
  - The words carry PK types, account groups and notes; the pack's warrant cites both regimes' law.
- *Środki trwałe → Pełna księgowość:* `pl-accept-fixed-asset` and `pl-depreciate-fixed-assets`
  stamp `book: "pl-stat"` on the module's objects. PK's four ŚT posting rules go to a small file
  installed only with both.
- *Core → Podstawa:* `pl-company-account` sets `company.iban`; `finance_v9`'s
  `group-trial-balance` and `intercompany-positions` read `purchase_invoice`.

**What the loader must learn first** (no file moves, no test changes).

- A pack names the pack it succeeds. Ownership is `created_by == "pack:"+name` exactly
  (`pack.go:268–283, 321–336`) and `Load` never retires a rule: a rule moved under a new name would
  leave the old and the new both active — a double booking. Every rule id, view id, door name and
  master-data dedup key (`pack/pl/…`) stays byte-identical.
- One owner per type, or fields added by extension. The pack's `company` v7 replaces core's v2
  wholesale; Podstawa, Pełna księgowość and KPiR all want fields on `company` and on the invoices. M.
- Words merged across the company's packs (core, base, module, regime). The kernel's `words()`
  (`answer.go:66–82`), the face (`app.js:1769–1770`) and `capture.py:61` read the first words view
  only. Domains and functions join the words; the face's stand-ins `DOMAIN_WORD` / `FUNCTION_WORD`
  (`desk.js:1007–1010`) go. M.
- Existing companies get install facts for Podstawa, Pełna księgowość and Środki trwałe, so nothing
  changes for them.
- The chart of accounts is installation-wide master data (`pack.go:377`) seen by every scope
  (`rhea_in_scope` of no company is true, `schema.sql:257`): a KPiR company sees Helios's chart; a
  second chart beside it makes `ResolveRef` by code ambiguous (`core/expr.go:57–69`); existing
  postings name installation account ids, so Helios's chart cannot be re-stamped. Decision D3.
- `packs/de` requires `vat_rate` and `sales_invoice`: putting them in Polska — podstawa makes
  Germany depend on Poland. Decision D2.

**Cut order:** the loader first → Środki trwałe out → the core pieces into the core → Podstawa's
links cut, Podstawa out → what remains is Pełna księgowość (its book stays `pl-stat`, its ids stay
`pl-…`). Germany (`packs/de` v3) cuts cleanly into Basis and Bilanzierung, with no cycles.

## 4. KPiR

- **What a KPiR entry is.** The period lock guards postings only (`exec/postings.go:83–103,
  316–327`); a posting must balance and each line has one side (`core/core.go:653`). Either KPiR
  entries are postings in a single-entry book, or a month's close locks every dated register entry.
  Proposed standard: a closed month in the KPiR locks as the journal does, every entry carries its
  document and its rule. M.
- **Polish accounting law in the kernel.** The late-document policy cites UoR art. 20/24 — forward
  to the first open month, never into the next year (`exec.go:257–275`, `firstOpenMonth`
  `postings.go:334`). Against invariant 8, and KPiR differs: a regime pack's parameter. S–M.
- **Matching differences without a chart** — see §3. S.
- **The author's prompt teaches Polish double entry** (`internal/agent/agent.go:201, 268–283,
  321–365`: 201/702/750/755/756, `seller_nip`, "accounting rules" meaning postings). Its examples
  come from the installed regime's words and notes. M.
- **The month's close in the face.** `nextOpenPeriod` reads `period_lock` or the first posting
  (`app.js:1465–1477`) and `bookOf` reads postings (`145–150`): a KPiR company never gets a
  month-close card, and an empty book would meet the "input book is required" refusal. The close
  card always promises "Po zamknięciu rozliczę VAT" (`1515–1525`); the tiles say "księga …"
  (`1577`); the phone's Księgi is the trial balance, VAT, result and ageing under "Każdy zapis jest
  zbilansowany: Wn = Ma" (`1679–1703`). The regime says which books exist and which month is open
  in each; the face draws them by role. M.
- **The invoice paper** always says "Faktura VAT" with a VAT line and "Okres VAT"
  (`app.js:313–370`), and the document's books show postings only (`420–429`). The exempt seller's
  wording comes with ryczałt; the paper as an App view with Apps step 2.

## 5. Tests and tools that pin today's layout

- `internal/shell/poland_test.go`: 41 types / 38 views / 138 rules / 39 doors / 67 events on first
  load, a reload skipping 323; `pack-pl-v44` three times; 55 accounts; 5 VAT rates.
  `reads_test.go:149`; `intercompany_test.go:104–108`; `cohabitation_test.go:124–136`
  (`pack-pl-v44`, `pack-de-v3`, 69 accounts). `words_test.go` TestThePackSpeaks expects one
  `slownik` naming every type, field and value.
- The four-file loop (people, periods, documents, `pl/pack.json`): `newCITWorld` (`cit_test.go`)
  used by eleven files, and the morning, answer, company_entered, reads and words tests.
  `markets/pl/ksef_test.go` installPoland, used by seven tests. `internal/pack/pack_test.go`;
  `stage_test.go` (at least ten pack files); `germany_test.go` (0 / 3 / 4 / 17);
  `markets/pl/pibr.go:354` (three common packs, pl, warehouse, nbp).
- The book name `pl-stat` across exec, core, sim and backfill tests and the face (`live-smoke.js:71`,
  `facebuild.go`, `bookOf`). Pełna księgowość keeps it, and these stay as they are.
- The face's tooling: `snapshot/*.json` is all Helios full books (`capture.py` captures one
  installation and its first words view; `build.py:25` lists view ids); `gridcheck.js` asserts
  full-books names (`363, 383–388, 408–424, 433–434, 446`) and fakes an empty company by stripping
  postings (`606–608`); `gallery.js` samples need postings, an account `/^13/`, `vat_settlement` and
  a waiting `pack-pl-v42`; `walk.mjs:64` installs everything for the installation and `106–108`
  give all three companies full books. Four snapshots are needed (Helios with full books, a KPiR
  sole trader, an exempt ryczałt firm, a company with nothing installed) and the walk installs per
  company.

A test changes only where it names a bundle or a count, and its commit says so.

## 6. The order of work, revised

0. The three fixes in §0, one commit each.
1. **The core, shipped by the kernel** (§2) — was step 5.
2. **Installing per company, at one version for everyone** (§1): install facts, the pack manifest
   and catalog, two approvals, the catalog seam, rules selected by install, offered before allowed,
   past months booked on a mid-year install, readers per market; and the face's half (views by
   role, columns by field, doors from definitions, Pakiety, the install card, the nothing-installed
   card, onboarding). Existing companies get installs of today's Poland pack: nothing changes for
   them, and replay proves it. Built with the company-rules session, whose mechanism it extends.
3. **Split the Poland pack** (§3): the loader first, then Środki trwałe, then Podstawa once its
   links are cut; Pełna księgowość is what remains.
4. **KPiR with the sole trader's month** (§4).
5. **Ewidencja przychodów (ryczałt).**

Later: a company staying on an older version (D1); rule history by date (the US review's cause,
also "documents keep the rule version of their date"); switching off with a date if step 2 does not
need it; the PDF reader's Polish dialect (debt of 2026-10-20); rights per pack or App; the rule set
per company (measure before acting).

## 7. Decisions for KK

- **D1. When Rhea publishes a new version of a pack, does every company that has it move at
  once, or may a company stay behind?** Recommendation: everyone moves, and Dziś shows what changed
  (as Rhea's rule changes are shown now). A company that wants its own way for one rule already has
  it: a company rule replacing Rhea's — with the warning, already owed, when the rule it overrides
  rests on the law and changes. Staying on a whole old version comes later, if an office asks; it
  roughly doubles step 2.
- **D2. Is an invoice Polish or European?** `packs/de` needs the invoice and the VAT rate too.
  Recommendation: a small European layer owns the sales invoice, the purchase invoice and the VAT
  rate, on the warrant of the VAT Directive 2006/112/EC art. 226 (what every EU invoice must carry);
  Polska — podstawa adds KSeF, FA(3) and the Polish fields, the German base adds its own.
- **D3. Does every company have its own chart of accounts?** Proposed standard: yes. The chart comes
  with Pełna księgowość and belongs to the company that installs it — UoR art. 10 ust. 1 pkt 3 lit. a
  makes the zakładowy plan kont part of each entity's own accounting policy. Helios's present chart
  stays Helios's (existing postings name it); a KPiR company has none and sees none.

**KK's answers (2026-10-23):** D1 yes. D2 — invoicing is a pack per country (Tylko plans invoicing
in PL, DE, NL, FR, CH, DK, NO); the invoice's common shape stays core, proposed with it. D3 — each
company its own chart, with segments, dimensions and groupings for reporting, the country standard
as a proposal the interview changes; Rhea's rules book to an account's role. The brief is in
DIRECTION, "Invoicing per country, and each company's own chart"; the split becomes four packs
(Podstawa, Fakturowanie — Polska, Pełna księgowość, Środki trwałe — Polska), with account roles and
the company's own chart before it.

## Corrections to the readers

- "120 of 138 Polish rules would book a German company's facts" was overstated: the entry rules
  carry the market binding and the rest follow from objects those create (§1.6); the risk is latent,
  not live.
