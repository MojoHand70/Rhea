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

**To build**

**M1. Split the Poland pack** into Polska — podstawa, Fakturowanie — Polska, Pełna księgowość and
Środki trwałe — Polska (step 3, parts 3–5): each company's own chart, the six links from the base
back into full books cut, the cut itself. *Waits for:* the chart design, agreed with KK first.
Unlocks: 21, 42–48, 53, 54.
- [ ] a sole trader's invoice is never booked into full books
- [ ] a full-books company's books are exactly as before the split (replay)
- [ ] each company's chart is its own

**M2. A form for any action** — the desk draws the form from the action's own definition (its
inputs, their types, what is required), offers it where the company has the action and the person
may, and shows what the kernel answers; no face code per function. Unlocks: 7, 8, 9, 15, 16, 19,
25, 29, 30, 32, 33, 35, 40, 41, 43, 45, 61, 63, 70. *Now: done (rheaface 83d4806) — in the search
box and on a record's sheet; on the desk today for 7, 8, 9, 16, 19, 40, 41, 61, 63, whose actions
exist; 25, 30, 32, 33, 35, 43, 45 wait for their own actions to be written in their packs; 29 and 70
are the kernel's own actions, not opened by the desk.*
- [ ] a pack's new action appears on the desk with its form, with no change to the face
- [ ] the form is offered only where the company has the action and the person may open it
- [ ] what the kernel refuses is said in Polish, field by field

**M3. Channels that run by themselves** — KSeF fetched every day, a statement read when it
arrives, mail per client read into its client; each pass the company's own, each document once.
Unlocks: 22, 23, 24, 26, 27.
- [ ] an invoice issued in KSeF is in the books next morning without anyone asking
- [ ] the same document arriving twice is one document

**M4. A checklist screen** — items computed from the books, each opening what it is about, the
open ones holding a close back. Unlocks: 58, later the year's close (61, 62).
- [ ] the month's checklist lists what blocks the close, and empties as it is done

**M6. Rhea's research step** — when a document meets no rule, Rhea may search the web (the acts,
the Ministry's pages, interpretations, guides) with a question that carries no client data, and
returns what she found with its sources, and a draft rule whose warrant they are; she writes nothing
herself (invariant 3) — the draft waits for a person. Unlocks: 83, and with it every rule she has not
been taught yet.
- [ ] her question carries no client data
- [ ] the draft names its sources, primary law first, and what is still to verify

**M5. Sending to the state from the desk** — the company's KSeF connection set up on the desk,
the JPK sent with its UPO kept, a note to the client. Unlocks: 20, 31, 52, 60.
- [ ] a company connects its KSeF on the desk; its token or certificate is never another's
- [ ] the JPK is sent and its UPO kept with the month

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

## E. Documents going out

**30. Issue a sales invoice** — the owner or the accountant · not yet on the desk
Buyer (by NIP), dates, lines — item, unit, quantity, unit price, discount, rate, codes — payment
→ the invoice with its number, its totals by rate worked out by Rhea, booked, sent to KSeF.
*Now: not yet (the kernel books an invoice from its fact; there is no form).*
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
- [ ] the correction refers to its invoice and corrects the registers

**33. Recurring invoices** — the owner · not yet on the desk
A template, a cadence → each month's invoice drafted for approval. *Now: not yet (the type exists).*
- [ ] next month's invoice is drafted by itself

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
