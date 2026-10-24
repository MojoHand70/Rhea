# Rhea MVP — the functions

The MVP: an accounting office keeps its whole mix of clients in Rhea — sole traders (KPiR,
ryczałt) and companies (full books) — running alongside its own program until Rhea can carry it
alone; KSeF reaches micro-firms on 1 January 2027. This is every function that takes, in plain
words, so we can see what the system does, tick what works and build the rest one by one. The
business generator comes after this map, to exercise what it names.

How to read one function:

- **The name**, then who does it and where on the desk.
- **What goes in → what comes out.**
- **Now:** *works* — on the desk, checked by the face's checks or the live walk; *partly* — some
  of it works; *kernel only* — Rhea can do it, but there is nothing on the desk for a person;
  *not yet*; *to verify* — I could not confirm it from the code.
- **The boxes** are its tests: tick one when you have tried it and it does what it says.
- **Built as an App** marks a document type, a flow or a company's own table: an App on M8, shipped
  standard and reshaped by the client. **Business pack** marks a collection of Apps for one kind of
  business or work (payroll, collections), designed after M8. Everything else is product: the
  channels, the checks, the law's objects, the export, the API, the interview.

## The order of work: catch up with flowMEE, then pass it

*Proposed by Claude, 2026-10-24, at KK's request; waiting for KK.* What flowMEE has and Rhea lacks:
`reports/autoMEE compared to Rhea.md` (internal: it names a competitor). The two new working parts
are designed in `EXPORT.md` (Rhea beside the office's program) and `API.md` (Rhea's API).

**Apps are how Rhea grows** (KK, 2026-10-24). Every invoice type and every document type, with its
workflow, is a SmartApp: its variables and formulas decide, for one, who approves an invoice. A
company's own kind of record (MPK, projects, budgets) is a SmartApp too, with only its data tables
defined: "Rhea, please add a new data object called MPK…". Rhea ships standard Apps, and the client
changes their design for any logic and any data. So approval paths, MPK, budgets and the live result
are not built by hand: they are Apps on the App machinery (M8), which comes first. What is not an
App is product: the law's channels, the checks, the export, the API.

**Machinery first.** A piece of machinery is finished before any function stands on it: its boxes
ticked, `./scripts/check.sh` green, replay identical, an independent review read. Each wave starts
with its machinery and ends with its functions; no function is built on a piece still being built.

| Wave | Machinery first | Then the functions | Done when | Against flowMEE |
|---|---|---|---|---|
| 1. The App machinery | M8, M7 | 100 (the first standard Apps), 101 (MPK and projects, as tables) | KK's MPK sentence becomes an approved App with its table, carried on a booking line; a cost invoice's App sends it for approval by a formula; lines at several rates book right | what none of them has: any logic and any data, per client, approved and replayable |
| 2. Documents in, checked, booked right | M3, M5 | 20, 22–27, 29–31, 55, 77–82, 87, 88, 92–95, 116 | a client's month of KSeF invoices and bank statements comes in by itself, is checked and booked by the invoice Apps, and a person sees only the questions | catches up on their core: intake, checks, dekretacja, the bank |
| 3. Beside the office's program, and onboarding | M1, M9, M11, M10 | 11, 14, 15, 21, 70, 73–76, 84–86, 103–114 | an office keeps Optima, runs a client in Rhea and imports the month; a new client is set up by interview, the share Rhea books by herself shown before it commits | catches up on how they sell, and passes them: Rhea keeps the books, compares them with the program's each month, and the office can move its books over |
| 4. Cost control and money | M12 | 32, 33, 60, 89–91, 96–98, 102, 115, 117 | the standard Apps for corrections, recurring invoices, receipts, payroll lists, budgets and the live result run for a client; the office asks its clients and a reminder reaches each once | catches up on the rest of their list |
| 5. Rhea's books for everyone | M6 | 34, 35, 42–48, 51–54, 56, 58, 61–63, 71, 72, 83 | an office closes a sole trader's and a company's month in Rhea and sends the JPK with its UPO | passes them for good: they leave all of this to the program |

*Outside the waves:* business packs, designed on M8 once it stands — collections (99), payroll
(119); with each country's invoicing pack — 118.

*Waits for KK:* the chart design (M1 opens wave 3; export needs each company's own chart); the
questions at the end of EXPORT.md and API.md.

## Machinery — what the functions stand on

Each piece names the functions (by number below) it unlocks, and is built in this order. A piece
of machinery is built only for the functions it names.

**Done**

- **The core shipped by the kernel** — a new server starts with the ledger, accounts, periods,
  people, documents and bank accounts, approved by a person (6752aee). Unlocks: every function on a
  new installation.
- **Packs installed per company** — a pack runs only where a company installed it; what a company
  sees follows its packs; Pakiety; booking the past after an install (d0e3264 … e6a3d85, 87fc2f2).
  Unlocks: 17, 18, and every function a pack brings to some companies and not others.
- **A definition belongs to the packs that declare it; words merged from every pack a company has**
  (c26aa46, a638021). Unlocks: splitting a pack without renaming anything.
- **M2. A form for any action** — the desk draws the form from the action's own definition (its
  inputs, their types, what is required), offers it where the company has the action and the person
  may, and shows what the kernel answers; no face code per function. Unlocks: 7, 8, 9, 15, 16, 19,
  25, 29, 30, 32, 33, 35, 40, 41, 43, 45, 61, 63, 70. Done in rheaface 83d4806 — in the search
  box and on a record's sheet; on the desk today for 7, 8, 9, 16, 19, 40, 41, 61, 63, whose actions
  exist; 25, 30, 32, 33, 35, 43, 45 wait for their own actions to be written in their packs; 29 and 70
  are the kernel's own actions, not opened by the desk.
  - [ ] a pack's new action appears on the desk with its form, with no change to the face
  - [ ] the form is offered only where the company has the action and the person may open it
  - [ ] what the kernel refuses is said in Polish, field by field

**To build** — in the order of the waves above.

*Wave 1*

**M8. SmartApps** (DIRECTION, "Rhea Apps and business packs" and "The App designer", 2026-10-23;
APPS.md, steps 2–6) — the App as one definition: variables and formulas; a visualization (a table
edited in its tab, a form, a board, a statement); its own data tables; a workflow (states,
transitions, actions, who may do what in each state, a running process keeping its version and
moved to a new one by default); triggers and autonomy (act, propose, earn it); connectors; the
designer (anyone asks, Rhea drafts, a designer or the owner puts it into production). A company's
own kind of record is an App with only its data tables defined (KK, 2026-10-24): fields, a code
fixed once made and a name that can change, a hierarchy, links to other tables, records other Apps
use and a booking line carries. Every invoice type and document type, with its workflow, is an App
(KK, 2026-10-24); who approves is a formula in it, not a machine of its own. An App is approved like
any definition; running it is business, recorded as facts. Unlocks: 98–102, 58, and every document
type as an App (25, 30, 32, 33, 87, 89, 97).
- [ ] KK's sentence — "add MPK: it can make hierarchies, it has a code and a name; once created the
  code cannot change, because bookings use it, but the MPK can be renamed; an MPK belongs to a
  business function and a region; we already have regions, so join them" — becomes an approved
  App with its table, edited in its tab
- [ ] a booking line carries an MPK, and a report by MPK equals the books
- [ ] a cost invoice's App sends it for approval by a formula over its amount, category, MPK and
  supplier; a rejected invoice books nothing
- [ ] a running process moves to its App's new version by default; a breaking change is answered
  in the draft before the version can go to production
- [ ] an App's table is its company's own (the isolation walk); replay reproduces identical state

**M7. Documents with lines and several rates** — a document keeps its lines (item, unit,
quantity, unit price, discount, rate, codes); the form takes lines (M2's lines input); rules run per
line; VAT worked out by rate (S3) or taken as printed (C1); the registers, the JPK and the bookings
by rate; each line booked by what it is for. Several rates on one invoice: KK, DECISIONS 2026-10-24.
Unlocks: 25, 30, 32, 77, 78, 81, 82, 87, 88, 97, 101, 106.
- [ ] an invoice at 23% and 8% gives a register row and the right VAT for each rate
- [ ] a purchase invoice split into two costs books two rows whose nets add up to the invoice's net
- [ ] replay reproduces the same lines and bookings

*Wave 2*

**M3. Channels that run by themselves** — KSeF fetched every day, a statement read when it
arrives, mail per client read into its client; each pass the company's own, each document once.
Unlocks: 22, 23, 24, 26, 27, 87 (a receipt sent by mail), 91 (a live bank feed, through a
licensed provider).
- [ ] an invoice issued in KSeF is in the books next morning without anyone asking
- [ ] the same document arriving twice is one document

**M5. Sending to the state from the desk** — the company's KSeF connection set up on the desk,
the JPK sent with its UPO kept, a note to the client. Unlocks: 20, 31, 52, 60. For a company whose books are kept in the office's program (103), the
JPK part is off: the program sends it.
- [ ] a company connects its KSeF on the desk; its token or certificate is never another's
- [ ] the JPK is sent and its UPO kept with the month

*Wave 3*

**M1. Split the Poland pack** into Polska — podstawa, Fakturowanie — Polska, Pełna księgowość and
Środki trwałe — Polska (step 3, parts 3–5): each company's own chart, the six links from the base
back into full books cut, the cut itself. *Waits for:* the chart design, agreed with KK first.
Unlocks: 21, 42–48, 53, 54, and 103–109: a company whose books are kept in the office's program
has that program's chart as its own (EXPORT.md).
- [ ] a sole trader's invoice is never booked into full books
- [ ] a full-books company's books are exactly as before the split (replay)
- [ ] each company's chart is its own

**M9. Export to the office's program** (EXPORT.md) — for each company, where its books are kept;
a writer per program in the market's code (Optima first), the entries in the company's own chart; a
batch is a fact, nothing is exported twice, a change after export travels as a correction, the
program's answer is a fact. Unlocks: 103–109. *Waits for:* M1 (each company's own chart).
- [ ] the same month exported twice adds nothing, also from two exports at once
- [ ] replay after a month of exports exports nothing
- [ ] a real Optima takes the file and its register totals by rate equal Rhea's

**M11. The interview loop** (DIRECTION, "Implementation by interview", step 6) — the client's
evidence in, the pack's rules over it, what is left unexplained grouped and asked about largest
first, every turn a fact, the agent's drafts approved as one set, the share explained shown
throughout. Unlocks: 85, 86.
- [ ] the simulated clients are done in every run, now through the desk
- [ ] the share explained equals the count in the books

**M10. Rhea's API** (API.md) — `/v1` over the same kernel functions the desk uses; a key is a
grant like a person's; every write carries its idempotency key; events from a cursor and by
subscription; each company's spec drawn from its packs; a program never approves. Unlocks: 110–113,
and the export's connector.
- [ ] the isolation walk passes over `/v1` with two companies
- [ ] a key cannot approve a rule, a bundle, a pack, a lesson or a booking decision
- [ ] the same write twice is one fact, also from several goroutines

*Wave 4*

**M12. Messages out** — mail (then SMS) to a person or a client, from a question, a path, a
close or a due date; each message is a fact and never sent twice; the answer comes back through a
door. Unlocks: 60, 99, the reminders of 100.
- [ ] each step sends once, also when two passes run at once
- [ ] replay sends nothing

*Wave 5*

**M6. Rhea's research step** — when a document meets no rule, Rhea may search the web (the acts,
the Ministry's pages, interpretations, guides) with a question that carries no client data, and
returns what she found with its sources, and a draft rule whose warrant they are; she writes nothing
herself (invariant 3) — the draft waits for a person. Unlocks: 83, and with it every rule she has not
been taught yet.
- [ ] her question carries no client data
- [ ] the draft names its sources, primary law first, and what is still to verify

**M4. The month's checklist** — no longer machinery of its own: an App on M8 (58), its items
computed from the books, each opening what it is about, the open ones holding a close back.
- [ ] the month's checklist lists what blocks the close, and empties as it is done

---

## A. Signing in and people

**1. Sign in** — anyone · the sign-in page
A key (passkey) → the person is in, acting for their company. *Now: works.*
- [ ] signing in with a key opens the desk on the company last worked for
- [ ] signing out leaves the sign-in page

**2. Join from an invitation** — the invited person · the link they received
The link → a key enrolled, the person in the company with the role the invitation carried. *Now: works.*
- [ ] the link enrols a key once and lands on the desk
- [ ] the link works seven days and only once

**3. Invite a person** — a company's owner · Ludzie → „Zaproś osobę”
Name, e-mail, role → a link to send; the person holds the role from today. *Now: works.*
- [ ] the link is shown to copy
- [ ] the person appears in „Kto ma dostęp” before they first come in

**4. Withdraw an invitation** — the owner · Ludzie → Zaproszenia → „Unieważnij”
An unused link → the link opens nothing, the person's access ends. *Now: works.*
- [ ] the link no longer opens
- [ ] the person is gone from „Kto ma dostęp”

**5. See who has access and who may do what** — the owner · Ludzie → „Kto ma dostęp”, „Kto co może”
→ every person, their role, where it comes from (membership, the office, a substitution), from when; every right by person. *Now: works.*
- [ ] each row opens the grant it rests on
- [ ] a person without the right to see this is told so

**6. End a person's access** — the owner · Ludzie → „Zakończ”
→ the membership ends yesterday; from now the person sees nothing. *Now: works.*
- [ ] the person is off the list
- [ ] nobody can end their own access

**7. Change a person's role** — the owner · the search box, or the record's sheet
Person, new role → from today they may what the new role allows. *Now: on the desk through its generic form (M2) — not yet tried with this action.*
- [ ] the new role shows in „Kto ma dostęp”
- [ ] what the old role allowed and the new does not is refused

**8. Appoint a substitute** — the owner · the search box, or the record's sheet
Who is away, who stands in, from–to → the substitute holds the absent person's rights for those days. *Now: on the desk through its generic form (M2) — not yet tried with this action.*
- [ ] the substitute acts in the company during the days, not after
- [ ] the substitution shows in „Kto ma dostęp”

**9. Team groups** — the office's owner · the search box, or the record's sheet
A group's name, its people, its clients → a supervisor sees the group's work. *Now: on the desk through its generic form (M2) — not yet tried with this action.*
- [ ] a person sees the clients of their group

## B. Offices and clients

**10. Found a company** — a person Rhea's staff invited · the link → „Załóż firmę”
NIP → the company from GUS and the VAT list, the founder its owner. *Now: works.*
- [ ] the form fills itself from GUS and the VAT list
- [ ] the founder lands on the new company's desk as its owner

**11. Be an office** — the office's owner
An office is a company that keeps others' books: its people work in its clients through the office's mandate. *Now: partly — an office is any company holding mandates; nothing names it an office.*
- [ ] an office's owner is offered „Dodaj klienta”; a company that is no office is not

**12. Add a client** — the office's owner · company switch → „Dodaj klienta”
NIP → the client company (from GUS and the VAT list), the office's mandate over it, a link for the client's owner. *Now: works.*
- [ ] the form fills itself from GUS and the VAT list
- [ ] confirmed, the client is in the switch
- [ ] the owner's link lets the client's owner in
- [ ] a client's own person is not offered „Dodaj klienta”

**13. Switch between companies** — anyone with more than one · the company switch
→ the desk works for the chosen company alone. *Now: works.*
- [ ] the desk opens again on the company last worked for
- [ ] nothing of one client shows on another's desk

**14. The office's list of clients** — the office's owner · not yet on the desk
→ every client: its form, books, who keeps them, new and lost this year, what each pays. *Now: not yet.*
- [ ] every client of the office is listed once
- [ ] new and lost clients this year are counted

**15. Who keeps which client** — the office's owner · not yet on the desk
Accountant, client → the accountant works in that client. *Now: kernel only (through groups).*
- [ ] an accountant sees only the clients given to them

**16. Record the company's form** — the owner or the office · the search box, or the record's sheet
Books (full, KPiR, ryczałt), VAT status (active, exempt), income tax (CIT, skala, liniowy, ryczałt), urząd skarbowy → what the company's books and returns follow. *Now: on the desk through its generic form (M2) — not yet tried with this action.*
- [ ] the form is shown on the company's page
- [ ] Rhea proposes the packs the form needs

**84. One queue of decisions across all clients** — the office's accountant · Dziś, the office's desk
→ every client's documents and bank lines waiting for a decision, in one queue, each opening in its
own client, so an accountant with forty clients does not open forty desks. *Now: to verify (the
kernel serves an office's scope; Dziś is drawn per company).*
- [ ] a question of client A opens in client A, never in B
- [ ] answered, it leaves the queue for everyone in the office

**115. The office talks with its clients** — the office's people and the client's · on a document, a month, an amount to pay
A question on a document with its options ("a private purchase or the business's?"), a reminder to
all clients ("documents by the 5th"), a short survey, an approval asked of the client → each message
and answer a fact, attached to what it is about; an answer the books can use is used (the client's
"yes, received", 79). Not a free chat at first. *Now: not yet.* *(KK's idea, 2026-10-25; the shape
is Claude's proposal.)*
- [ ] the client answers on the phone with one tap, and the document moves on
- [ ] a reminder to all clients reaches each of them once

*Stands on:* M12.

**85. Set a new client up by interview** — Rhea, with the client's owner or the office · the new company's desk
An intelligent walkthrough, easy and structured (KK, 2026-10-25): a fixed path of topics — the
company, its books and taxes, its documents and channels, its bank, its people, the Apps it needs —
one question at a time, each with its options and Rhea's recommendation, in plain words. Evidence
first: what she can read from the client's documents, the registers (GUS, the white list) and its
program's files she does not ask; she asks about what is left, the largest group first, saying how
many documents each answer settles. The answers become rules, approved as one set; the share Rhea
books by herself is shown throughout, and before the client commits (DIRECTION, "Implementation by
interview", 2026-10-08). *Now: not yet (simulated clients only: 8 of 8 runs done).*
- [ ] before anything is approved, the owner sees how many of their documents Rhea would book by herself
- [ ] each question says how many documents its answer would explain
- [ ] done means nothing left unexplained, counting what was knowingly set aside

*Stands on:* M11, M3, the readers of 73 and 74 for the history.

**86. Learn from the client's past bookings** — Rhea, at the start
The past year's entries from the client's program (JPK_KR, JPK_PKPIR) → the rules that would
reproduce them, each proposed with how many past entries it explains ("this supplier to 402-03: 48
of 48"); approved, they book the new documents. *Now: not yet.*
- [ ] a proposed rule reproduces the past entries it claims, and the count is shown
- [ ] nothing is learned without an approval

*Stands on:* 73, 74, M11.

**114. Re-interview** — the owner or the office, or Rhea offering it · any topic of the interview, any day
"We bought a car", "we are a VAT payer now", "fuel goes to another account" → Rhea opens that topic,
shows what she was told, when and by whom, asks only what changes, shows what the change would do to
the books, and the new answers apply from a date as new versions, the past keeping the old ones
(KK, 2026-10-25). She offers one herself when the evidence shifts: a new kind of document keeps
arriving, turnover nears the VAT limit. *Now: not yet (versions from a date, simulation and booking
the past exist in the kernel).*
- [ ] a change applies from its date; the months before stay booked as they were
- [ ] the person sees what the change does to the books before approving it
- [ ] an answer overruled never comes back

*Stands on:* M11.

## C. Setting a company up

**17. Install packs** — the owner or the office · Ustawienia → Pakiety
A pack, from which day → its rules book the company's facts from that day. *Now: works.*
- [ ] Zainstalowane and Dostępne are shown, each pack with what it needs
- [ ] installed, the pack's actions are offered
- [ ] installed from an earlier day, Rhea proposes booking the past

**18. Switch a pack off** — the owner or the office · Pakiety → „Wyłącz”
Last day → it books nothing after; what it booked stays readable. *Now: works.*
- [ ] its actions are gone; its screens stay

**19. The company's bank accounts** — the owner or the office · the search box, or the record's sheet
Bank, IBAN, currency, purpose → statements are read into the right account. *Now: on the desk through its generic form (M2) — not yet tried with this action.*
- [ ] each account is listed with its bank and currency

**20. Connect KSeF** — the owner · not yet on the desk
The company's KSeF token or certificate → Rhea sends and fetches invoices for it. *Now: not yet (an operator's command today).*
- [ ] connected, Rhea says so and from when
- [ ] a company's connection is never another company's

**21. Chart of accounts** — the accountant · Księgi → Plan kont
See the chart; add an account; set an account's tax treatment. *Now: partly — seeing works; adding is not on the desk; the chart is still loaded as rows every company sees (to change: each company's own).*
- [ ] the chart is the company's own
- [ ] a new account can be added and is booked to

## D. Documents coming in

Every way a document comes in is a channel: product, not an App. The reader is chosen by the
document's kind, and one document is one document whichever channel brought it (built). The
channels (KK, 2026-10-25: enumerated):

| Channel | Function | Now |
|---|---|---|
| KSeF — purchase invoices, and sales invoices the client issued elsewhere | 22 | kernel only: the reader; nothing fetches by itself |
| Upload on the desk — PDF, XML, a photo, a statement | 23 | kernel only |
| Mail — an address per company | 24 | not yet |
| The phone's camera — receipts, paper invoices | 87 | not yet |
| A watched cloud folder — Drive, OneDrive, Dropbox | 116 | not yet |
| Bank statement files — MT940, CAMT.053, a bank's CSV | 26, 27 | MT940 kernel only, one bank's detail lines read |
| A live bank feed, through a licensed provider | 91 | not yet |
| Payment gateways and marketplaces — Stripe, PayU, Przelewy24, PayPal, Allegro, Amazon | 90 | not yet |
| The client's program — JPK_KR, JPK_PKPIR, JPK_EWP, JPK_V7, its chart | 73, 74, 104 | not yet |
| Tables — CSV, Excel, the columns' mapping taught once | 75 | not yet |
| Rhea's API | 111 | not yet |
| Payroll programs' lists | 89 | not yet |
| Cash registers and e-receipts — the day's takings | 117 | not yet |
| E-invoice networks abroad — Peppol, XRechnung | 118 | not yet |

Lookups, not documents: GUS and the VAT white list (built, read at onboarding), NBP rates (built),
VIES (95, not yet).

**22. Purchase invoices from KSeF** — Rhea, every day
→ every invoice issued to the company in KSeF, read, booked as a purchase invoice. *Now: kernel only (the reader exists; nothing fetches by itself).*
- [ ] an invoice issued in KSeF appears without anyone uploading it
- [ ] the same invoice fetched twice is one invoice

**23. Upload a document** — anyone in the company · not yet on the desk
A PDF or a photo → the document kept; an invoice read from it and proposed. *Now: kernel only.*
- [ ] the uploaded file is kept and opens
- [ ] an invoice in it is read and waits for confirmation

**24. A mail address per client** — the client · e-mail
Mail with documents → the documents kept in that client. *Now: not yet.*
- [ ] mail to the client's address lands in that client only

**25. Enter a cost invoice by hand** — the accountant · not yet on the desk
The document as it is — supplier, number, dates, its totals by rate as printed (and its lines, when
they are wanted) — then the booking: rows by what each part is for, each with its net, its VAT and
how much of the VAT may be deducted → a purchase invoice in the VAT register and the books.
*Now: not yet.*
*Built as an App* (100).
- [ ] it is in the VAT purchase register and the books
- [ ] the same invoice twice is refused
- [ ] one invoice split into two costs books two rows; their nets add up to the invoice's net
- [ ] a car used privately too: half the VAT deducted, the other half in the cost

*What Rhea must learn* — each a rule with its warrant, drafted by Rhea from the sources
(`reports/Invoices in practice 2026-10.md`), checked, and approved by KK as an accountant would;
*exists* marks a rule that does a part of it today, at one rate per invoice.
- C1 *The document as it is.* The VAT of each rate is taken as the supplier printed it, never
  worked out again; Rhea warns when a rate's VAT is off its net by more than the rounding.
  Fakturowanie — Polska.
- C2 *The month of the deduction:* the later of the supply and the receipt of the invoice (VAT act
  art. 86 ust. 10, 10b). Podstawa. *Exists* (the purchase invoice's VAT period).
- C3 *What each part is for decides where it goes:* in full books a cost account by category, or
  goods and materials, or a fixed asset; in KPiR a column — goods, incidental costs, other expenses
  (the KPiR regulation of 6 September 2025, columns *to verify*). Pełna księgowość; KPiR. *Exists*
  in part (a supplier's cost account, else 403).
- C4 *How much VAT may be deducted:* all, none, or 50% for a passenger car used for the business
  and privately (art. 86a); the deductible part to input VAT (221), the rest to the cost with the
  net. Podstawa decides the share; the regime books it.
- C5 *Fixed asset or other purchase,* for the JPK: K_40/K_41 or K_42/K_43 (the Ministry's JPK_V7
  brochure, 1 February 2026). Podstawa. *Exists* for other purchases only.
- C6 *The booking in full books:* each row's net to its account, the deductible VAT to 221, the
  gross to the supplier (202). Pełna księgowość. *Exists* at one account per invoice.
- *Lessons from the accountant,* as company rules: "this supplier's invoices go to 402-03", "fuel
  for this car at 50%", "this one is a fixed asset" — proposed after the accountant has decided the
  same way, kept once approved.

*Stands on:* M2's lines input, and the purchase invoice keeping its booking rows.

**26. Bank statement in** — the accountant · not yet on the desk
An MT940 file → its lines, matched to the invoices they pay. *Now: kernel only.*
- [ ] lines match their invoices by themselves
- [ ] the same statement twice is one statement

**27. Read a Polish bank's detail lines** — Rhea
The statement's title lines → who paid, for which invoice. *Now: partly (one dialect read).*
- [ ] the counterparty and the invoice number are read from each Polish bank's file

**28. Say what an unmatched bank line is** — the accountant · Dziś
A line nobody explains → its kind (a fee, a tax, the owner's), booked. *Now: works.*
- [ ] the card offers the kinds; one tap books it

**29. Pair a payment with an invoice by hand** — the accountant · not yet on the desk
Line, invoice(s) → settled; a difference booked. *Now: kernel only.*
- [ ] the invoice shows as paid

**87. A receipt or a paper invoice from the phone** — anyone in the company · the phone's camera, mail
A photo or a scan → seller, date, lines and rates read; a receipt up to 450 zł with the buyer's NIP
taken as a simplified invoice; booked by its App. *Now: not yet (the PDF reader reads invoices).*
- [ ] a receipt with the buyer's NIP under 450 zł lands in the VAT purchase register
- [ ] the same receipt sent twice, also as another photo, is one (93)

*Built as an App* (100) for what it books; the camera is a channel.

**88. An invoice from abroad** — Rhea
An invoice in another language or currency → read; import of services, intra-EU acquisition (WNT) or
import of goods recognised; reverse charge in both VAT registers; the exchange rate for VAT and for
income tax, each by its rule. *Now: partly (currencies, pl v19–v22; import of services, WNT and
reverse charge are not in the pack).*
- [ ] a German software subscription books output and input VAT in the same month
- [ ] the NBP rate of the working day before is used, and named

*Built as an App* (100); reading the language is the channel's.

**89. A payroll list from a payroll program** — the accountant · upload, mail
The payroll program's list → its column totals checked against its rows; the entries booked
(salaries, ZUS, PIT, PPK). *Now: not yet.*
- [ ] a list whose totals do not tie is stopped, the column named

*Built as an App* (100), the list coming through a channel. Calculating payroll is a business pack's
(119).

**119. Payroll calculated** — the accountant or the client's HR · Kadry i płace
Employees, contracts, the month's time and absences → gross to net, ZUS, the PIT advance, PPK, the
payroll list and its bookings. *Now: not yet.*
- [ ] a month's list matches a payroll program's for the same employees

*Business pack* (Kadry i płace, KK, 2026-10-25): Apps on M8; the law's rates and limits are the
market pack's data.

**90. Payouts from payment gateways and marketplaces** — Rhea
A payout from Stripe, PayU, PayPal or Allegro, with the operator's report → the sales it pays, the
fee (a cost) and the returns, each sale paired with its invoice. *Now: not yet.*
- [ ] a payout's sales, fees and returns add up to the amount on the bank line
- [ ] the fee is booked as a cost, never netted off the sales

**91. A live bank feed** — Rhea, through a licensed account-information provider
→ the company's bank lines as they happen, not once a day from a file. *Now: not yet (statement files
only).*
- [ ] a line in the bank appears within minutes, once

**92. The KSeF number in transfers** (from 1 January 2027) — Rhea, part of the KSeF invoice (product)
Paying: the transfer title carries the invoice's KSeF number, or one collective identifier for
several invoices of one supplier. Receiving: a payment is paired by the KSeF number in its title.
*Now: not yet.*
- [ ] a payment with a KSeF number in its title pairs with that invoice without a person
- [ ] a payment order for three invoices of one supplier carries one collective identifier

**116. A watched cloud folder** — the client · its Drive, OneDrive or Dropbox folder
A file dropped in the company's folder → read into that company, once. *Now: not yet.*
- [ ] a file dropped twice is one document
- [ ] client A's folder never feeds client B

**117. Cash registers and e-receipts** — Rhea, daily
The day's takings from online cash registers and e-receipts → sales by rate, booked; the takings
paired with the deposits and card settlements. *Now: not yet (to verify: what the Ministry's central
repository of online cash registers gives a taxpayer).*
- [ ] a day's takings by rate equal the register's daily report

**118. E-invoices from networks abroad** — Rhea, with each country's invoicing pack
Peppol, Germany's XRechnung and ZUGFeRD and the like → read as invoices, in a company whose
country's pack is installed. *Now: not yet.*
- [ ] an XRechnung invoice books like a KSeF one

## D2. Checking a document (every document, in and out)

An accountant checks every document before it reaches the books: formally, arithmetically and on
its substance, then the tax questions, then decides how it is booked (dekretacja). Rhea does the same
with rules — each check a rule that runs on every document and records what it found, with its
warrant; what she cannot check herself she asks a person, and the answer is a fact.

**77. Formal check** — Rhea, on every document
Does it carry what its kind must — for any accounting document the type, number, parties, what
happened, its value, the dates (UoR art. 21); for an invoice also NIPs, the number in its series,
dates of issue and supply, lines, rates (VAT act art. 106e) → passed, or what is missing named. *Now:
not yet.*
- [ ] an invoice without the buyer's NIP is stopped, the missing item named

**78. Arithmetic check** — Rhea, on every document
Quantity × price, the lines' values, each rate's net and VAT, the gross agree to the grosz (UoR art.
22) → passed, or where it is off. *Now: not yet.*
- [ ] a line whose value is not quantity × price is shown, with the difference

**79. Substantive check** — Rhea, then a person
Did it happen, does it concern the business, was it received? Rhea checks what the books know — an
order, a goods receipt, a contract, a supplier the company works with, a cost that fits its
business — and asks the person responsible for the rest; the answer is recorded. *Now: not yet.*
- [ ] an invoice from a supplier the company never bought from is asked about
- [ ] the person's "yes, received" is kept with the invoice

**80. Counterparty check** — Rhea, on every document
The supplier is an active VAT taxpayer; above 15,000 zł its bank account is on the white list
(CIT art. 15d, PIT art. 22p, Ordynacja art. 117ba); split payment where it is required; the invoice
came through KSeF. *Now: partly (the white list is read when a company is added).*
- [ ] a payment above 15,000 zł to an account off the white list is stopped before it is made

**81. VAT deduction check** — Rhea, on every cost invoice
Does the invoice give the right to deduct (VAT act art. 86), or is it one of those that do not (art.
88)? → deductible in full, in part, or not. *Now: not yet.*
- [ ] an invoice for an exempt sale is not deducted

**82. Dekretacja** — Rhea proposes, the accountant decides
The month and the way of booking — accounts or KPiR column, the VAT register — proposed by the
booking rules (C1–C6, S8–S10); the accountant's approval is the signature the act asks for (UoR art.
21 ust. 1a): who, when, kept as a fact. After the same decision approved enough times, a lesson
books it by itself. *Now: not yet.*
- [ ] every booked invoice shows who decided how it was booked, and when

**83. Rhea looks up what she was not taught** — Rhea
When no rule checks or books a document, she searches as an accountant would — the acts first, the
Ministry's explanations, interpretations, professional guides — never with the client's data in her
question; she writes down what she found with its sources, drafts the rule with that warrant, and
proposes it: to the office for its client, to Rhea's staff for every company. Approved, it is a rule,
kept for good — and learned once for everyone it applies to. *Now: not yet (today she says she was
never taught it, and the question reaches Rhea's staff).*
- [ ] a document no rule books gets a proposed rule, its sources named
- [ ] her search never carries the client's name, NIP or amounts
- [ ] approved once, the next such document is booked without asking

**93. Duplicate check** — Rhea, on every document
The same document twice, whatever door it came through, is one document; a near duplicate (the same
receipt photographed twice; the same seller, amount and date under another number) is stopped and
asked about. *Now: partly (one purchase invoice is one fact, by the seller's NIP and its number, from
every door).*
- [ ] the same receipt as two photos is stopped before it is booked

**94. A supplier's bank account changed** — Rhea, on every cost document, and again when payments are prepared
An account on the invoice other than the supplier's earlier ones → one of three (Claude's proposal,
2026-10-25): on the white list for the same NIP, a note only; not on the white list, held before
payment (paid above 15,000 zł, it loses the tax-cost deduction and can make the company liable for
the supplier's VAT); the invoice came as a PDF by mail rather than through KSeF, held until a person
confirms the change by phoning the supplier on a number already known (a swapped account arrives by
mail; a KSeF invoice comes from the supplier's own authorised session). Each confirmation is a fact:
who, when, how. The comparison is the base's; the white list is the Polish pack's. *Now: not yet.*
- [ ] an invoice with a new account off the white list is held, the change named
- [ ] a confirmed change is kept with who confirmed it and how, and the account is known from then

**95. VIES check** — Rhea, on every EU counterparty
→ the counterparty's VAT number checked in VIES on the day and kept with the document (an intra-EU
supply at 0% needs it). *Now: not yet.*
- [ ] an intra-EU supply to an invalid VAT number is not at 0%

**96. An invoice against its order and goods receipt** — Rhea, on a purchase invoice
The invoice's lines against the order and the PZ → quantities and prices agree, or the differences
are shown. *Now: kernel only (the matching engine's three-way, 2026-10-11, `TestMatchingThreeWay`: the order's,
the receipts' and the invoice's lines; the quantity allocated, the price within 2 %, the difference
booked).* Product; Apps use it.
- [ ] an invoice for more than was received is held, the line named

## E. Documents going out

**30. Issue a sales invoice** — the owner or the accountant · not yet on the desk
Buyer (by NIP), dates, lines — item, unit, quantity, unit price, discount, rate, codes — payment
→ the invoice with its number, its totals by rate worked out by Rhea, booked, sent to KSeF.
*Now: not yet (the kernel books an invoice from its fact; there is no form).*
*Built as an App* (100).
- [ ] the invoice gets the next number
- [ ] it is in the VAT sales register and the books
- [ ] lines at 23% and 8% give a total for each rate, VAT worked out on each rate's sum
- [ ] a discount shows on its line: price before, the discount, the value after

*What Rhea must learn* — each a rule with its warrant, drafted by Rhea from the sources, checked,
approved by KK; *exists* marks a rule that does a part of it today, at one rate per invoice.
- S1 *The number:* the next in the company's series, identifying the invoice once (VAT act art.
  106e ust. 1 pkt 2). Fakturowanie — Polska.
- S2 *A line's value:* quantity × net unit price, less its discount unless the discount is already
  inside the price (art. 106e ust. 1 pkt 8–11); on an invoice priced gross, from the gross unit
  price (ust. 7–8, *to verify*). Fakturowanie — Polska.
- S3 *The totals by rate:* each rate's net is the sum of its lines' nets, its VAT that sum times the
  rate, rounded to the grosz (art. 106e ust. 1 pkt 14, ust. 11) — or each line's VAT stated and
  added up (ust. 10), as the invoice says. Fakturowanie — Polska.
- S4 *The rate of each line,* and for an exempt line the legal basis of the exemption on the
  invoice (art. 106e ust. 1 pkt 19, *to verify*). Fakturowanie — Polska.
- S5 *An invoice in another currency:* the VAT in złoty at the NBP rate of the working day before
  (*to verify* the article). Podstawa. *Exists* (the invoice's złoty amounts).
- S6 *The month the VAT is due:* by the date of supply or the invoice's (art. 19a). Podstawa.
  *Exists* (the sales invoice's VAT period).
- S7 *Split payment:* the mark where an item of annex 15 and over 15,000 zł gross (*to verify*).
  Fakturowanie — Polska.
- S8 *The booking in full books:* the gross to the customer (201), the net to sales by kind
  (goods, services, other), the VAT to 222. Pełna księgowość. *Exists* at one rate.
- S9 *The entry in KPiR:* the net to the sales column. KPiR.
- S10 *The VAT register and the JPK by rate:* each rate's net and VAT in its pair (K_19/K_20 for
  23%, K_17/K_18 for 8%, …). Podstawa. *Exists* at one rate.
- S11 *Sent to KSeF,* the KSeF number kept on the invoice. Fakturowanie — Polska. *Exists* (the
  channel, KSeF's test system).

*Stands on:* M2's lines input, the sales invoice keeping its lines, M5 for sending.

**31. Send it to KSeF** — Rhea, on issue
→ the KSeF number, or the reason it was refused. *Now: kernel only (tested on KSeF's test system).*
- [ ] the KSeF number shows on the invoice
- [ ] a refusal is shown with its reason

**32. Correct an invoice** — the owner or the accountant · not yet on the desk
The invoice, what changes → a correction invoice, booked, sent to KSeF. *Now: to verify.*
*Built as an App* (100).
- [ ] the correction refers to its invoice and corrects the registers

**33. Recurring invoices** — the owner · not yet on the desk
A template, a cadence → each month's invoice drafted for approval. *Now: not yet (the type exists).*
*Built as an App* (100).
- [ ] next month's invoice is drafted by itself

**97. Other kinds of invoice** — the owner or the accountant · not yet on the desk
Advance (zaliczkowa), final (rozliczeniowa), RR for a farmer, pro forma (not an invoice: nothing is
booked), each with its KSeF type. *Now: not yet.*
*Built as an App* (100).
- [ ] the final invoice deducts the advances invoiced before it
- [ ] a pro forma books nothing

## F. Full books (spółki)

**34. Define a journal** — the chief accountant · not yet on the desk
Name, which documents go in it → entries filed by journal. *Now: not yet.*
- [ ] each entry belongs to one journal

**35. A PK document (polecenie księgowania)** — the accountant · not yet on the desk
Date, description, lines Wn / Ma → a balanced entry, booked. *Now: not yet.*
- [ ] an unbalanced PK is refused
- [ ] a PK into a closed month is refused

**36. The journal and an entry's origin** — the accountant · Księgi → Dziennik; any entry → „Dlaczego”
→ every entry; for each, the fact and the rule it came from. *Now: works.*
- [ ] every entry opens its document and its rule

**37. Trial balance** — the accountant · Księgi → Zestawienie obrotów i sald
→ every account's turnover and balance for a period. *Now: works.*
- [ ] debits equal credits

**38. Another cost account for a supplier** — the accountant · a purchase invoice → „Inne konto kosztów”
Supplier, account → that supplier's invoices booked there from today. *Now: works.*
- [ ] the next invoice of that supplier goes to the new account

**39. Spread a cost over months (RMK)** — the accountant · a purchase invoice → „Rozłóż koszt w czasie”
From which month, how many months → each month's part booked at its end. *Now: works.*
- [ ] each month takes its part; the remainder is shown

**40. Fixed assets** — the accountant · the search box, or the record's sheet
Accept (OT), depreciation each month, dispose (LT). *Now: on the desk through its generic form (M2) — not yet tried with this action.*
- [ ] depreciation is booked each month by itself
- [ ] the register lists each asset with its value

**41. Provisions, write-downs, loans, FX revaluation** — the accountant · the search box, or the record's sheet
*Now: on the desk through its generic form (M2) — not yet tried with this action.*
- [ ] each is offered where it applies and booked

## G. KPiR and ryczałt (sole traders)

**42. The KPiR book** — Rhea · Księgi → KPiR
Sales and cost invoices, bank lines where they count → entries in the book's columns. *Now: not yet.*
- [ ] each invoice lands in its column
- [ ] the month's totals match the invoices

**43. A KPiR entry by hand** — the accountant · KPiR → „Dodaj wpis”
Date, document, column, amount → the entry. *Now: not yet.*
- [ ] it shows in the book in its order

**44. The KPiR month** — the accountant · KPiR
→ the month's totals, carried from the beginning of the year. *Now: not yet.*
- [ ] the totals from the beginning of the year carry on

**45. Stock count (remanent)** — the accountant · KPiR
Date, the stock's value → it counts in the year's income. *Now: not yet.*
- [ ] the year's income takes it

**46. The ryczałt record** — Rhea · Księgi → Ewidencja przychodów
Sales → revenue by rate. *Now: not yet.*
- [ ] each sale lands under its rate

**47. PIT advance** — Rhea, each month or quarter
The book's income, the tax form (skala, liniowy, ryczałt) → the advance to pay, when. *Now: not yet.*
- [ ] the advance matches a hand calculation

**48. The owner's ZUS and health contribution** — Rhea, each month
The owner's basis → what to pay, when. *Now: not yet.*
- [ ] the amounts match ZUS's for the year

## H. VAT and JPK

**49. VAT registers** — the accountant · Księgi → rejestr VAT sprzedaży, zakupu
→ every invoice of the month with its VAT. *Now: works.*
- [ ] a month's totals match the invoices

**50. The month's VAT settlement** — Rhea, on the month's close
→ VAT due or to carry over. *Now: works.*
- [ ] it matches the registers

**51. The JPK_V7M file** — the accountant · not yet on the desk
→ the file for the month, as the Ministry's schema wants it. *Now: kernel only.*
- [ ] the file passes the Ministry's schema

**52. Send the JPK** — the accountant · not yet on the desk
→ sent, with its UPO. *Now: not yet.*
- [ ] the UPO is kept with the month

**53. Quarterly VAT (JPK_V7K)** — Rhea, for a small taxpayer · *Now: not yet.*
- [ ] the quarter is settled and filed

**54. A VAT-exempt company** — Rhea · *Now: not yet.*
- [ ] no VAT registers, no VAT on the close card

**55. Check a supplier's account against the white list** — Rhea, before a payment · *Now: to verify (the white list is read at onboarding).*
- [ ] a payment to an account off the list is flagged

## I. Taxes

**56. CIT advance** — Rhea, each month · *Now: kernel only (shown in her morning).*
- [ ] it matches a hand calculation

**57. The tax calendar** — anyone · Terminy podatkowe
→ what is due and when, the company's own. *Now: works.*
- [ ] every due date of the month is there

## J. Month close

**58. The month-close checklist** — the accountant · Dziś or Zamknięcie miesiąca
→ what is left before the month can close: documents waiting, bank lines unexplained, VAT checked. *Now: not yet.*
*Built as an App* (100).
- [ ] each item opens what it is about
- [ ] the month cannot close while an item is open, unless the accountant says so

**59. Close the month** — the accountant · Dziś → „Zamknij”
→ the month locked, VAT settled. *Now: works.*
- [ ] nothing can be booked into a closed month
- [ ] a company with no books is offered no month to close

**60. Tell the client what to pay** — Rhea, after the close
→ "you pay X by the 25th", to the client. *Now: not yet.*
- [ ] the client receives the amounts and the dates

## K. Year close

**61. Close the year** — the chief accountant · the search box, or the record's sheet
→ cost and revenue accounts closed into the result. *Now: on the desk through its generic form (M2) — not yet tried with this action.*
- [ ] the result equals the year's profit and loss

**62. Financial statements (bilans, rachunek zysków i strat)** — the chief accountant · not yet on the desk
*Now: not yet (the result is shown; the statements are not drawn).*
- [ ] they balance and match the trial balance

**63. Distribute the result** — the owners · the search box, or the record's sheet · *Now: on the desk through its generic form (M2) — not yet tried with this action.*
- [ ] the distribution is booked as decided

## L. Reports and questions

**64. Ask Rhea about the books** — anyone · the box, „Zapytaj Rheę”
A question → the answer, from the books, every figure opening its source. *Now: works.*
- [ ] she says when she was never taught something

**65. Keep an answer as a report** — anyone · the answer → „Zachować”
→ a report, waiting for approval, then in Księgi. *Now: works.*
- [ ] approved, it is in the menu for everyone of the company

**66. Why is this here** — anyone · any figure or record → „Dlaczego”
→ the fact, the rule, its version, its legal basis. *Now: works.*
- [ ] every booked figure answers

**67. Her morning** — anyone · Dziś
→ what changed, what waits, in a few sentences. *Now: works.*
- [ ] every amount in her words opens

**68. Receivables and payables** — anyone · Dziś, Księgi
→ who owes, how long. *Now: works.*
- [ ] the ages match the invoices' dates

**98. The result live, with a forecast to the month's end** — the owner · Dziś, Księgi
→ the profit and loss as documents arrive, every line opening its documents; the costs still to come
(rent, subscriptions, payroll) forecast from the past and marked as forecast until their document
arrives. *Now: partly (the result is in her morning; no forecast).*
- [ ] a forecast line is replaced by the real one when its document arrives
- [ ] every figure opens its source

*Built as an App;* stands on M8.

**99. Collections** — Rhea, after the due date
Reminders by mail and SMS on the company's schedule, stopped when the money arrives; promises to pay
kept; a phone call by voice later, if wanted. *Now: not yet.*
- [ ] paid the day before, no reminder goes
- [ ] the same step never sends twice

*Business pack* (collections, to be designed, KK, 2026-10-25): Apps on M8, sending through M12.

## M. Rhea learning

**69. Approve or refuse what Rhea learned** — the owner or the office · Dziś
→ approved, it works from its date; refused with a reason, it is not proposed again. *Now: works.*
- [ ] a refused lesson does not come back

**70. Teach once for many clients** — the office · not yet on the desk
A lesson, the clients → each gets it as its own. *Now: kernel only.*
- [ ] each chosen client shows the lesson as its own

**71. Rhea's changes on Dziś** — anyone · Dziś · *Now: not yet.*
- [ ] a change of Rhea's rules is told to each company it reaches

**72. A warning where a company's rule overrides one resting on the law** — the office · *Now: not yet.*
- [ ] the warning names the law

## N. Bringing data in

**73. Migrate a company (JPK_KR)** — the office · not yet on the desk · *Now: not yet.*
- [ ] the year's balances and entries match the old program's

**74. Migrate a sole trader (JPK_PKPIR, JPK_EWP)** — the office · *Now: not yet.*
- [ ] the book matches the old program's

**75. Import from CSV or Excel** — the office · *Now: not yet.*
- [ ] customers and suppliers come in once each

**76. Opening balances** — the accountant · *Now: not yet.*
- [ ] the trial balance opens with them

## O. Apps: the standard flows and a company's own tables

**100. Standard Apps for the common flows** — Rhea ships them; the client's designer reshapes them · Aplikacje
Five to ten sample Apps, each a document type with its workflow; the client may change any logic
and any data in them. Proposed first ten: the cost invoice (read, checked, sent for approval by a
formula, booked), the sales invoice (issued, sent to KSeF), the correction, the recurring invoice,
the receipt and the expense claim, the payroll list, the budget, the month's checklist, the live
result, the office's monthly settlement with its client. *Now: not yet (APPS.md maps today's tabs
into Apps, step 1).*
- [ ] who approves a cost invoice is a formula the client changes, with no change to the code
- [ ] a client's changed App moves its running invoices to the new version, or keeps them, as
  the designer chose

*Built as an App;* stands on M8.

**101. A company's own tables: MPK, projects, regions…** — the owner or the accountant asks Rhea in words · Aplikacje
"Rhea, please add MPK…" → an App with only its data tables: fields, a code fixed once made and a
name that can change, a hierarchy, links to the tables the company already has; edited in its tab;
its records used by other Apps and carried on booking lines, a line split by keys. *Now: partly
(Rhea drafts new record types for approval; a field is fixed at birth unless declared changeable;
links between types exist; no hierarchy, no editable tab, no code on a booking line).*
- [ ] KK's MPK sentence (M8) gives the table he described
- [ ] a line split 60/40 books two rows whose nets add up to the line's
- [ ] the report by MPK equals the books

*Built as an App;* stands on M8.

**102. Budgets** — the owner · Aplikacje
A budget per MPK or project per month → spent, committed (approved, not yet booked), left, and a
warning before the limit. *Now: not yet.*
- [ ] the warning comes before the document that passes the limit is booked

*Built as an App* (a table of budgets and the formulas over it); stands on M8.

## P. Beside the office's program (EXPORT.md)

For a company whose books are kept in the office's program, Rhea books and exports; the JPK, the
taxes and the close are the program's (51–54, 56, 58–63 are off for that company).

**103. Where the books are kept** — the office · the company's page
In Rhea, or in the office's program (Optima first), from a date. *Now: not yet.*
- [ ] a company whose books are in the program is offered no JPK to send and no month to close

**104. Read the client's chart and lists from its program** — the office · when the client is set up
The program's export (chart, counterparties, categories, MPK; the JPK_KR) → the company's own
accounts and lists in Rhea. *Now: not yet.*
- [ ] every account of the program's chart is the company's own, none another's

*Stands on:* M1.

**105. The account for each of Rhea's bookings** — the accountant · once per company, or once for the office's clients that share a chart
→ each of Rhea's bookings has its account in the client's chart; a booking without one waits as a
question. *Now: not yet.*
- [ ] nothing is exported to an account the client's chart does not have

**106. Export the month to the program** — the accountant · Księgi → „Eksport do programu”
A period → the file the program imports (Optima: into its VAT registers through the JPK file first);
the batch kept as a fact. *Now: not yet (the JPK_V7M writer exists, kernel only).*
- [ ] the same month exported twice adds nothing
- [ ] Optima's register totals by rate equal Rhea's

*Stands on:* M9, M7.

**107. The program's answer, and changes after export** — the accountant (the import report), or the connector
→ the program's document numbers kept; its refusals as questions; a document changed after export
goes as a correction in the next batch. *Now: not yet.*
- [ ] a correction never travels as a second original

**108. Compare the program's books with Rhea's** — Rhea, each month
The program's own JPK_V7 (and JPK_KR) → what is missing, changed, typed in twice, or in another
period. *Now: not yet.*
- [ ] a document typed in by hand in the program is found

*Stands on:* 73.

**109. Move the books to Rhea** — the office · the company's page
From a date, Rhea's books are the official ones; the export stops; the opening balances are checked
against the program's closing ones. *Now: not yet.*
- [ ] the opening balances equal the program's closing balances

## Q. Rhea's API (API.md)

**110. A key for a program** — the owner or the office · Ludzie → „Programy”
A name, a role, the companies, from when → a key shown once; withdrawn, it opens nothing. *Now: not
yet.*
- [ ] a key of company A reads nothing of B
- [ ] a key cannot approve

**111. Send a document, get back the result** — a program · the API
A file with its idempotency key → what Rhea read, checked and booked, or the question waiting.
*Now: not yet (the intake doors exist).*
- [ ] the same document sent twice is one

**112. Be told what happened** — a program · the API
A subscription → each new fact of the kinds asked for, signed, retried until answered; or read from a
cursor. *Now: not yet.*
- [ ] a program that missed a day reads on from where it stopped

**113. Read any view, call any action** — a program · the API
→ every view and action the company's packs give it, with the same rights and refusals as a person;
a pack's new view or action is there with no code change. *Now: not yet (the desk's own API does this
for the face).*
- [ ] a pack's new action appears in the company's API spec
