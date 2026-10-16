# The United States, read for a first year — the register

**2026-10-16.** KK asked what a US pack would take. The answer was one pack for the
whole country, with the states as data inside it. This register is where that pack
would begin, as `packs/pl/REVIEW.md` was for Poland: the questions two small American
businesses meet in their first year, each with the law that answers it.

No US pack exists, so where the Polish register said what the pack does today, this one
says what Rhea's **shared ground** gives: the kernel, `packs/common` and the shared
finance definitions every market starts from. It also says where the work would
live. And each point carries a new line, **Watch**: where a change to it would first be
published. Together those lines are the first list for the law watch (KK, 2026-10-16:
Rhea must follow the law as an accountant does).

**The two businesses**, the American mirror of the Polish sole trader and sp. z o.o.:

- **Maple Lane Goods**, a sole proprietor. A single-member LLC in Indianapolis, disregarded
  for federal tax (Schedule C), cash method. Handmade home goods sold through its own
  web shop, one marketplace and craft fairs, plus some design work. A few freelancers,
  no employees, about $120,000 in its first year.
- **Hoosier Supply Inc.**, a small corporation. An Indiana C corporation (Form 1120),
  accrual method, a wholesale distributor with about $3M of revenue. It sells to
  resellers and some consumers, holds stock in one Indiana warehouse, crosses Ohio's
  sales-tax threshold in year one and California's in year two, and pays ten employees
  through a payroll provider. The S-corporation election is a variant.

**States**: Indiana (home; one statewide rate, no local sales tax, a Streamlined member),
Ohio (Streamlined, county and transit taxes), California (not Streamlined, district
taxes). Federal law throughout. **State of law**: 16 October 2026, tax year 2026, with
enacted changes named with their dates.

**How to read an entry.** Each point is a question the books must answer, in the order
a business meets it during the year, with:

- **Applies to**: sole proprietor, corporation, or everyone.
- **The law**: the citation, with the decisive words quoted.
- **Grade**:
  - *1 plain*: the text decides.
  - *2 settled*: the text is unclear, but a regulation, ruling, notice, bulletin or
    court settled it (named).
  - *3 open*: practice splits; both readings are named, with who holds each.
- **Rhea's default**: the reading the US pack should ship.
- **Rhea today**: what the shared ground already gives, and what is missing.
- **Where the work lives**:
  - *pack data*: rules, rates and definitions, with no code;
  - *market code*: a reader or writer in `markets/us`;
  - *shared ground*: something every market shares must change;
  - *later*: not needed in a first year.
- **Deadline**: the statutory date, when there is one.
- **Watch**: the official source that would publish a change first, how it publishes,
  and its address.
- **Sources**: inline.

**Caveats on sources:**

- Claims found only in secondary sources are marked "(secondary — verify)".
- Dates and amounts the researchers worked out themselves say "computed here".
- **The official state law sites refused automated readers all day**: iga.in.gov,
  codes.ohio.gov and leginfo.legislature.ca.gov. State statutes are therefore quoted
  from FindLaw's copies (current to 1 January 2026) or through the revenue
  departments' own bulletins, and each entry says which. Every state quotation must be
  re-read on the official site, in a browser, before the pack ships.
- The 2026 editions of several IRS instructions were not yet posted. Where an entry uses
  the 2025 edition, it says so.

Five researchers wrote the sections, one area each. Nothing here has been read by a
practising US accountant (CPA or enrolled agent) yet.

## Summary

Each point is counted once, under the main place its work lives.

| Area | Points | Plain | Settled | Open | Pack data | Market code | Shared ground | Later |
|---|---|---|---|---|---|---|---|---|
| Sales tax at home — Indiana | 33 | 20 | 13 | 0 | 21 | 3 | 7 | 2 |
| Selling into other states | 30 | 20 | 10 | 0 | 17 | 5 | 6 | 2 |
| Contractors, information returns, banks and money | 30 | 20 | 9 | 1 | 22 | 5 | 2 | 1 |
| Sole proprietor — books and income tax | 33 | 28 | 5 | 0 | 26 | 1 | 5 | 1 |
| Small corporation — books and income tax | 32 | 20 | 11 | 1 | 22 | 1 | 7 | 2 |
| **Total** | **158** | **108** | **48** | **2** | **108** | **15** | **27** | **8** |

The shared ground is the main place for 27 points and is touched by 47.

**What the numbers say.**

- **Most of the US pack is data.** The law prescribes no form of books. IRS Pub. 583
  says "the law does not require any specific kind of records". There is no
  e-invoice clearance and no monthly audit file. The Schedule C lines serve a sole
  proprietor as the KPiR columns serve a Polish sole trader, dated by tax year.
  108 of 158 points need only rules, rates and definitions.
- **The shared ground has to move before most of that data can land.** The US is the
  first market without VAT, so it finds every place where Poland's shape leaked into
  what all markets share. Germany cannot find these, because Germany has VAT too. The
  ten causes below cover all 47 points that touch the shared ground.
- **The market code is small.** It is 15 points:
  - readers: bank statements (BAI2, OFX, CSV), marketplace and card-processor payouts,
    the payroll provider's journal, the Streamlined rate and boundary files, the
    California rate service, the Federal Reserve's exchange rates;
  - writers: 1099s through the IRS's IRIS portal (CSV in year one; its API in year two),
    the Streamlined simplified return (XML) for Indiana and Ohio, Ohio's monthly return
    as a spreadsheet upload.
- **Indiana's monthly sales-tax return has no machine route in year one.** The return is
  keyed into INTIME by hand (STX-27), so Rhea prepares it and a person types it in.
- **Two points are open:**
  - **PAY-13**: is interest a corporation pays on a shareholder loan reportable on
    1099-INT from $600 or from $2,000? The IRS's instructions, not yet revised, say
    $600; the statute as amended in 2025 says $2,000.
  - **CO-23**: an Indiana seller whose only California activity is soliciting orders,
    plus a website. What does it owe California once its sales cross the threshold?
    The Franchise Tax Board's booklet says it "may be liable" for the $800 minimum
    even when the federal protection for sellers of goods (P.L. 86-272) applies.
    Practitioners read that protection as reaching any tax measured by net income, and
    the Board's 2022 internet guidance as void (secondary — verify).
- **This review graded the mechanics, not the grey zone.** As in Poland, whether a given
  expense is deductible at all is decided expense by expense. That is within reach of
  a practising accountant and, later, of the network.

## The root causes

The 47 shared-ground points come from ten causes. Fixing a cause fixes every point
under it. Causes 1, 3 and 4 block the most pack data; the order is for KK to confirm.

| # | Cause | Points | What fixes it |
|---|---|---|---|
| 1 | **The sale is a VAT invoice.** `sales_invoice` requires one `vat_rate` for the whole document, has no lines, no ship-to and no delivery fact, and its buyer must be a company with a tax number. A US sale is taxed line by line: by what the item is, where the buyer receives it, and whether a certificate exempts it. Rates stack (state, county, transit, district). Postage, installation and design stay untaxed only on their own lines. A fair booth or marketplace sale has no identified buyer at all, and the revenue date comes from the shipment. | STX-02, STX-11, STX-12, MST-06, MST-08, MST-12, MST-15, MST-18, MST-22, PAY-16, SP-18, CO-13 | Lines (as objects, since fields are scalar or ref) each with a kind and a tax treatment. A ship-to or pickup place per line. A buyer that may be a person or nobody, with a daily summary for anonymous sales. Tax components per jurisdiction. A delivery fact the revenue rule reads. Poland needs the consumer sale too (VAT-06). |
| 2 | **Purchase tax is always recoverable VAT.** `purchase_invoice.vat` books as an input-tax asset. Sales tax paid on a purchase is cost, and use tax is a liability the buyer assesses itself. Purchase lines are needed to classify a payment for the 1099 and to spot an asset. | MST-11, PAY-08, CO-09 | Purchase lines. A tax amount whose treatment (recoverable, cost, self-assessed) is decided by the rules, not by the type. |
| 3 | **Rates are whole numbers without dates, and a new rule version erases the old one's past.** `vat_rate.percent` is an integer, so 7.25% cannot be stored, and no rate carries a validity interval or a jurisdiction. The executor runs only the newest active version of each rule, and `effective_from` is only a lower bound ([exec.go:462](../../internal/exec/exec.go#L462), `ActiveRules`). When the 2026 mileage rate changed on 1 July (72.5¢ to 76¢), a June trip entered in July would match neither version. The same happens to Indiana county tax rates (January, October), quarterly sales-tax rates and Poland's VAT-05. | STX-10, MST-24, PAY-16, SP-16, SP-27 | Rates as dated master data (exact decimal, valid from and to, per jurisdiction), which rules read by the event's date. A rule version that keeps applying to events dated before its successor. This is also what the law watch needs: a law that changes on 1 January must not stop December's documents being booked in January. |
| 4 | **The company says what it is in Polish.** `legal_form`, `books` and `income_tax` take Polish values. Identifiers are `vat_id` and NIP fields, one per company, while a US company holds an EIN, a state account per state and later more. Owner-level facts (estimated tax; in Poland PIT and ZUS) sit on the company. Object types are versioned by name across packs, so a US `company` version would supersede Poland's in any database that loads both. Places of business and employees' work locations carry no state. | STX-01, MST-02, MST-03, SP-02, SP-03, SP-25, CO-02, CO-03 | A list of identifiers (scheme, value, valid from) and registrations per jurisdiction. Declared values a market may extend without editing one shared type. Owner facts on the person. Postal addresses with a state. A dated work location on `person`. |
| 5 | **Choices are not facts yet.** The first federal return locks in the overall method, the inventory method, the recurring-item exception and the bad-debt method. Some elections are made afresh every year (de minimis, the bonus opt-out, §179). | CO-07 | A generic dated election, declared through a door. Poland needs the same shape for the Estonian CIT (CIT-01), the VAT cash method (VAT-04) and the form of tax (FORM-04). |
| 6 | **An asset has one Polish book.** `fixed_asset` (in `packs/pl`) hard-codes two Polish depreciation lines. A US asset needs a book line, a federal line and one per state: Indiana does not follow federal bonus depreciation and caps §179 at $25,000 (DOR IB 118), and California follows neither. `task`, `task_schedule`, `vehicle`, `mileage_entry` and `stocktake` also live in `packs/pl`, with kilometres and Polish fields. | SP-01, SP-12, SP-13, SP-16, SP-17, CO-10, CO-11 | An asset register where one asset carries any number of books. Tasks, schedules, vehicles, mileage and stocktakes moved to `packs/common`, in units the market declares. |
| 7 | **Stock has a quantity but no cost.** The warehouse records quantities only: no FIFO layers, no cost-of-goods-sold posting. A maker deducts materials when the finished goods are sold (SP-10), and a distributor values stock by FIFO (CO-08). | SP-10, SP-29, CO-08, CO-27 | Valued receipts and cost layers consumed by issues, with a stocktake carrying lines and valuation. The same gap as Poland's FB-19. |
| 8 | **A secret identifier cannot go into a log that never forgets.** A contractor's tax number is usually a Social Security number. Once written to the append-only log it can never be removed, and masking it in views is not enough. Poland never hit this because a NIP is public; a PESEL would hit it too. | PAY-06, SP-03 | A sealed field kind: the value kept outside the plain log, a stable fingerprint inside it so replay stays deterministic (as the KSeF receipt's fingerprint is verified). |
| 9 | **The calendar is the month.** There is no business-day calendar, no day arithmetic in the formula language, no settlement period other than the month, and no two-year cadence. US deadlines move to the next business day (IRC §7503); Indiana assigns monthly, quarterly or annual filing; the Indiana entity report is biennial. | STX-04, CO-01, CO-19 | A business-day and holiday calendar each market supplies to the clock. Day arithmetic in formulas. Settlement periods of any length. Polish deadlines need the same roll. |
| 10 | **What every market's auditor or bank needs is filed under Poland.** A camt.053 statement reader belongs beside MT940 for every market. Every auditor asks for an export of the log and postings with provenance. Accounts need one optional mapping field for return lines (Poland's JPK_KR_PD tags need the same). | PAY-02, PAY-05, SP-09, CO-05, STX-32 | camt.053 in `internal/adapter`. A market-neutral examination export. A mapping field on `account`. And the law watch itself (STX-32), which every market shares. |

## Dated

Points with a date that does not move, or a step that must start before one.

| By | What | Points |
|---|---|---|
| **in force since 1 July 2026** | Indiana sales tax collected is "held in trust" for the state. A responsible person is personally liable, and knowingly failing to remit is a Level 6 felony. | STX-18 |
| **in force since 15 March 2026** | Indiana retailers may round cash totals to the nickel. The rounding stays out of the tax base, and the full tax is still remitted. | STX-17 |
| **now (October 2026)** | Apply for an IRIS transmitter control code: about 45 business days, so a code applied for today arrives around 22 December, in time for the 1 February 2027 filing. FIRE takes no new applications since 21 July 2026. | PAY-26 |
| **1 January 2027** | IRIS becomes the IRS's only electronic channel for 1099s. Indiana remittances are rounded down to $0.05. | PAY-26, STX-17 |
| **15 January 2027** | The sole proprietor's last 2026 federal estimated payment. | SP-25 |
| **1 February 2027** | 1099-NEC to contractors and the IRS for 2026 payments ($2,000 threshold). Platforms' 1099-K to sellers. 31 January is a Sunday. | SP-20, PAY-26, SP-19 |
| **15 March 2027** | Year-end bonuses accrued for 2026 must be paid to be deducted for 2026. Form 1120-S, for the S-corporation variant. | CO-26, CO-31 |
| **15 April 2027** | The 2026 returns: Form 1040 with Schedule C and Indiana IT-40; Form 1120. | SP-32, SP-27, CO-30 |
| **15 May 2027** | Indiana IT-20 for 2026. Form 103, business personal property, for the 1 January 2027 assessment. 15 May is a Saturday; the move to Monday is not checked. | CO-20, CO-24, SP-31 |
| **17 September 2027** | The Same Day ACH limit changes. | PAY-18 |
| **November 2027** | The Fedwire release that needs structured addresses on wires. | PAY-20 |
| **from 2027** | The $2,000 1099 threshold is indexed, rounded to $100; no 2027 figure is published yet. | PAY-09 |

## What the law watch would read

The Watch lines name the sources a US pack would follow. By kind:

- **Statute**: the U.S. Code (uscode.house.gov, "current through" a public-law number);
  Congress.gov for bills; the Indiana General Assembly (sections take effect 1 July
  unless they say otherwise, IC 1-1-3-3; 2027 is a long session from January), the Ohio
  and California legislatures, and California's election calendar for district taxes.
- **Regulation and guidance**: the eCFR (point-in-time versions with an API); the Federal
  Register (an API; FinCEN's and Treasury's rules); IRS revenue procedures (the yearly
  inflation amounts each autumn), notices, announcements (the mid-year mileage rate)
  and the Internal Revenue Bulletin; draft forms and instructions on irs.gov.
- **State revenue departments**: Indiana's information-bulletin and departmental-notice
  indexes (HTML tables whose links change when a bulletin is replaced, so a diff
  catches revisions), its monthly Tax Bulletin, and the yearly Legislative Synopsis;
  Ohio's GovDelivery bulletins; the CDTFA's quarterly "Explanation of Tax Rate
  Changes"; the FTB's pages.
- **Data**: the Streamlined rate and boundary files (CSV, quarterly, posted a month
  ahead and sometimes corrected late) and its taxability matrix (a public JSON API
  with dated versions, whose comments carry meaning the flags do not); the CDTFA
  rate files and address service; Ohio's address lookup (an API with a date, needing
  a key); the Federal Reserve's exchange rates.
- **Not public**: the Nacha Operating Rules (sold) and the FASB Codification (behind a
  login). Their changes reach Rhea through summaries until someone holds a licence.

## The sections

The five researchers' entries follow unedited. Where areas meet, the sections
cross-refer instead of repeating each other: the multistate section points to the
Indiana section for home-state detail.

## Sales tax at home — Indiana

Reviewed against the Indiana Code, title 6, article 2.5 (state gross retail and use taxes) and article 8.1 (tax administration); the Indiana Administrative Code, 45 IAC 2.2; the Department of Revenue's Sales Tax and General Tax Information Bulletins and Departmental Notices as its tax-library index listed them on 16 October 2026; the ST-103 instructions; the DOR Business FAQ, sales-tax and marketplace-facilitator pages; the DOR's SST Business Rules (last revised September 2026) and bulk-filing matrix; and the Streamlined Sales and Use Tax Agreement as amended through 19 December 2025. Law as of 16 October 2026. The General Assembly's site (iga.in.gov) is a JavaScript application that returned no text to any fetch, so the statute quotes below come from FindLaw's reproduction of the Indiana Code ("current as of January 01, 2026") and the 45 IAC quotes from Cornell LII's reproduction. Changes made in the 2026 session come from the DOR's 2026 Legislative Synopsis (revised April 2026), not from the enrolled acts. Every IC quote should be re-read against iga.in.gov before the pack ships. Forms on forms.in.gov (ST-105, ST-115, IT-40 Schedule 4) were refused (HTTP 403); they are described from the DOR pages that cite them. Abbreviations: "IC" = Indiana Code; "IAC" = Indiana Administrative Code; "DOR" = Indiana Department of Revenue; "SIB n" / "GIB n" = DOR Sales Tax / General Tax Information Bulletin number n; "DN n" = DOR Departmental Notice n; "RRMC" = Registered Retail Merchant Certificate; "TID" = Indiana Taxpayer Identification Number; "SST" / "SSUTA" = Streamlined Sales Tax / the Streamlined Sales and Use Tax Agreement; "SER" = SST Simplified Electronic Return; "SEA" / "HEA" = Senate / House Enrolled Act; "Maple Lane" = the sole proprietor, "Hoosier Supply" = the corporation.

Entries are ordered by when a business meets them in its first year: setting up, first purchases, first sales, payments, month end, later in the year, year end.

---

### STX-01 — Do we have to register before our first sale, and what does it cost?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-8-1(a): "A retail merchant may not make a retail transaction in Indiana unless the retail merchant has applied for a registered retail merchant's certificate."
  - (b): the merchant pays "a registration fee of twenty-five dollars ($25) for each place of business listed on the application". (d): the DOR issues "a separate registered retail merchant's certificate for each place of business listed on the application".
  - (g): the certificate "is valid for two (2) years after the date the registered retail merchant's certificate is originally issued or renewed". If all returns are filed and taxes paid, the DOR renews it "within thirty (30) days after the expiration date, at no cost to the retail merchant".
  - IC 6-2.5-9-2: a merchant who sells without a certificate, or after it is revoked or suspended, "commits a Class A misdemeanor".
  - A wholesaler is a retail merchant too. SIB 52: "persons engaged in making wholesale sales in Indiana are Indiana retail merchants". The DOR's sales-tax page: "You'll need one certificate on display for each business location". Registration is through INBiz (SIB 91: "Registration is $25").
  - The $100,000 threshold in IC 6-2.5-2-1(d) is for sellers without physical presence. A business with an Indiana location registers whatever its volume (SIB 89).
- **Grade:** 1 plain
- **Rhea's default:** the company declares each tax registration as a dated fact: the Indiana TID with a three-digit location number per place of business (the DOR's SER rules call it "the full 13-character Taxpayer ID (TID/Loc #)"), the RRMC issue date and its two-year renewal date. A sale dated before the registration raises a task; it is not booked silently.
- **Rhea today:** missing. The shared `company` (testdata/seed/finance_v8.json) has one `vat_id`, a Polish NIP in practice: rules resolve parties with `ref(company, vat_id, …)`. The Poland pack's `company` v7 adds `vat_status`. A US company holds several registrations at once, each with its own dates: a federal EIN, an Indiana TID per location, later an Ohio vendor's license and a California seller's permit.
- **Where the work lives:** *shared ground*. The one-tax-number-per-company assumption (`company.vat_id`, the invoice's `seller_nip`/`buyer_nip`) must become a list of registrations per jurisdiction. The Indiana entries are then *pack data*.
- **Deadline:** before the first retail transaction in Indiana (IC 6-2.5-8-1(a)). Renewal every two years is automatic while the account is compliant.
- **Watch:** IC 6-2.5-8 changes only by act of the General Assembly. Sessions run January to March or April, and a section takes effect on 1 July unless it says otherwise (IC 1-1-3-3). The DOR summarises each session in its Legislative Synopsis PDF (in.gov/dor/files/legislative-synopsis-2026.pdf, one file per year). The fee and display rule also appear on the DOR sales-tax page, which is unversioned HTML: snapshot it and diff.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-8-1/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-9-2/ ; https://www.in.gov/dor/i-am-a/business-corp/sales-tax/ ; https://www.in.gov/dA/d88598eb06/sib52.pdf ; https://www.in.gov/dA/778f09c373/sib91.pdf ; https://www.in.gov/dA/88a60658b3/sib89.pdf ; https://www.in.gov/dor/files/sst-business-rules.pdf

---

### STX-02 — Is each craft fair a new place of business, and does an event need its own permit?
- **Applies to:** sole proprietor
- **The law:**
  - GIB 102 (March 2023): the tax "also applies to transient merchants, defined in IC 25-37-1 as one who engages in temporary business in Indiana. Before selling at retail in Indiana, a merchant is required to obtain a Registered Retail Merchant Certificate (RRMC)".
  - Local licences are separate. "an RRMC is not the same as a Transient Merchant License". The county licence is needed "if the county so requires. As of the issuance date of this bulletin, Parke County was the only county with such a requirement. Some Indiana cities and towns have merchant licensing requirements as well". A seller who trades only in some months "can be permitted by the department to file on a seasonal basis".
  - SIB 70: "Persons selling arts, crafts and items not suitable or intended for human consumption are retail merchants and must register and collect sales tax", with "a one-time registration fee of $25.00".
  - A sale handed over at the booth is sourced to "that business location" (SIB 96, rule 1, IC 6-2.5-13). It is an Indiana sale whoever the buyer is.
  - IC 6-2.5-8-1(b) and (d) speak of "each place of business".
- **Grade:** 2 settled. The code does not say whether a fair booth is a "place of business" needing its own $25 location. GIB 102 and SIB 70 treat event and market sellers as holding one RRMC and filing the ST-103 under it, with no registration per event. I found no DOR text calling a booth a separate place of business (unverified either way).
- **Rhea's default:** Maple Lane holds one RRMC, for its home office. Each fair is an `event` (dates, venue, city, county) that the day's sales point to. A week before each fair, a task asks for a check of the county and city licence (Parke County is the one known). Every booth sale is an Indiana sale at 7%.
- **Rhea today:** missing. There is no event or venue. A booth sale goes to an anonymous consumer, but the shared `sales_invoice` requires `buyer: ref<company>` and the Poland pack resolves buyers by tax number. This is the same wall VAT-06 hit for Polish sales to consumers. A fair day needs one summary record (takings by tax treatment, cash and card), not one invoice per sale.
- **Where the work lives:** *shared ground* for a sale without an identified buyer (the B2B-invoice assumption) and for a daily sales summary. The event and the licence task are *pack data*.
- **Deadline:** the RRMC before the first fair; a local licence before each event that requires one.
- **Watch:** GIB 102 is on the DOR information-bulletin index, a table of number, title, publication date and effective date (in.gov/dor/resources/tax-library/information-bulletins/). County transient-merchant ordinances have no central feed; the source is each county auditor.
- **Sources:** https://www.in.gov/dA/f031b3065a/gb102.pdf ; https://www.in.gov/dA/e7027b4aa4/sib70.pdf ; https://www.in.gov/dA/f6b43ccd68/sib96.pdf ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-8-1/

---

### STX-03 — Do we report sales tax when we sell, or when we are paid?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-6-2: "A retail merchant may, without prior departmental approval, report and pay his state gross retail and use taxes on an accrual basis if he uses the accrual basis to pay and report the adjusted gross income tax". Also: "The department may, at any time, require the retail merchant to stop using the accrual basis."
  - DOR Business FAQ: "Absent the use of the accrual method for income tax, the retail merchant must use the cash method." And: "The merchant can use the cash method for sales tax and accrual method for adjusted gross income tax".
  - IC 6-2.5-2-2(a) measures the tax by "the gross retail income received by a retail merchant".
  - Against this, the 1982 rule 45 IAC 2.2-6-8(a) tells the merchant to "total gross retail income from taxable transactions made during the reporting period". (b): that is the liability "regardless of the amount of tax he actually collects".
- **Grade:** 2 settled. The statute makes accrual an option tied to the income-tax method. The DOR's FAQ settles that every other merchant reports on the cash basis, although the old rule's "transactions made" reads as accrual.
- **Rhea's default:** the reporting basis is a dated company fact. Maple Lane reports federal income on the cash method, and so its Indiana income, so it must report sales tax on the cash basis: tax falls in the period the payment is received. Hoosier Supply reports income on the accrual basis and ships on accrual, which matches its books and keeps the bad-debt deduction (STX-25) available. Cash is its other lawful choice.
- **Rhea today:** wrong shape for Maple Lane. The Poland pack puts each sale in the period of its invoice or sale date (`vat_period = month(…)`), and no rule knows a period by payment date. The matching engine already pairs payments with invoices. The tax inside each payment (payment × tax ÷ gross, the proportional split IC 6-2.5-6-9(d)(6) prescribes for later payments on bad debts) can therefore be computed from facts already in the log.
- **Where the work lives:** *pack data*: a `sales_tax_basis` company field and a period rule keyed on the matched payment date. No code.
- **Deadline:** none. The basis can be changed, but accrual needs accrual income tax.
- **Watch:** the DOR Business FAQ page is unversioned HTML (snapshot it and diff). IC 6-2.5-6 changes come by session. 45 IAC 2.2-6 is maintained in the Indiana Register (iar.iga.in.gov), which I did not reach.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-6-2/ ; https://www.in.gov/dor/i-am-a/business-corp/business-faq/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-2-2/ ; https://www.law.cornell.edu/regulations/indiana/45-IAC-2.2-6-8 ; https://www.law.cornell.edu/regulations/indiana/45-IAC-2.2-6-4

---

### STX-04 — How often do we file, and when is each return due?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-6-1(a): "each person liable for collecting the state gross retail or use tax shall file a return for each calendar month".
    - Due "not more than thirty (30) days after the end of that month" when average monthly liability in the preceding calendar year did not exceed $1,000.
    - Due "not more than twenty (20) days after the end of that month" when it did.
  - (b) lets the DOR set other periods, with returns due 20 days after the period ends.
  - (c) allows a calendar-year period "if the retail merchant's state gross retail and use tax liability in the previous calendar year does not exceed one thousand dollars ($1,000)". That return is due "not later than the last day of the month immediately following the close of that reporting period".
  - (e): "A retail merchant shall report and remit state gross retail and use taxes through the department's online tax filing program."
  - New businesses: "Upon registration and for the first full year, the merchant will be required to file sales tax returns (Form ST-103) based on their taxable sales volume for the preceding calendar year, as determined by the department" (GIB 102; SIB 70).
  - DOR Business FAQ: "All business accounts are reviewed in November and any filing status changes are made at that time". You can increase how often you file "but you can't decrease".
  - A zero return is still due (ST-103 instructions). "If the due date printed on your return falls on a weekend or holiday, the due date becomes the next legal business day."
  - Storm relief: the DOR's September 2026 bulletin gives the 21 counties hit by storms from 11 August 2026, Marion (Indianapolis) among them, "until Feb. 1, 2027, to file various federal and business tax returns and make payments". It does not say whether sales tax is included.
- **Grade:** 1 plain. A January return under the 30-day rule is due 2 March in a non-leap year (computed: 30 days after 31 January). The DOR's bulletins paraphrase the rule as "the 30th day of the month following", a day February does not have.
- **Rhea's default:**
  - The filing frequency is a dated fact taken from the DOR's assignment, not computed by Rhea.
  - Each due date is the period end plus 20 or 30 days, moved to the next business day, and checked against the date INTIME prints.
  - Every period while registered raises a filing duty, including zero periods.
  - The storm relief is not assumed to cover sales tax until the DOR says so.
- **Rhea today:** partly there. The clock fires dated duties and the period lock closes months. But the formula language has `end_of_month` and `add_months` and no way to add days. No business-day or holiday calendar exists for any market (Poland needs one too). An annual filer's return spans twelve monthly locks, and the settlement is tied to the month.
- **Where the work lives:** *shared ground*: a business-day calendar supplied per market to the clock, day arithmetic in the formula language, and settlement periods that are not months. The frequency facts are *pack data*.
- **Deadline:** monthly returns 20 or 30 days after month end; annual returns by 31 January.
- **Watch:** frequency changes arrive as INTIME correspondence after the November review. They are account-specific and not public, so they come in as received documents. Disaster extensions appear in the DOR's monthly Tax Bulletin (email through subscription.in.gov, republished as in.gov/dor/files/2026-september-tax-bulletin.pdf and so on, one per month) and on IRS.gov.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-6-1/ ; https://www.in.gov/dA/f031b3065a/gb102.pdf ; https://www.in.gov/dA/e7027b4aa4/sib70.pdf ; https://www.in.gov/dor/i-am-a/business-corp/business-faq/ ; https://www.in.gov/dA/d85d002b5d/st-103inst.pdf ; https://www.in.gov/dor/files/2026-september-tax-bulletin.pdf ; https://www.in.gov/dor/files/2026-october-tax-bulletin.pdf

---

### STX-05 — Should we register through the Streamlined system instead of, or as well as, INBiz?
- **Applies to:** everyone (both sell into other SST states; Hoosier Supply crosses Ohio's threshold in its first year)
- **The law:**
  - SIB 91: "A seller registering under the Agreement is registered in each of the member states", and "Regardless of the method used to register with the state of Indiana, the seller agrees to collect Indiana sales and use taxes for all taxable sales made into Indiana."
  - DOR SST Business Rules (September 2026): "Taxpayers with physical nexus in Indiana must complete both the SST registration and Indiana business registration forms. The application fee is $25."
  - On filing: "Model 1 & 2 sellers must file the SER in Indiana monthly, even if they were filing with Indiana using a different filing frequency before registering for Streamlined Sales Tax. These sellers are required to file one return for all locations", due the 20th.
  - "Model 4 sellers who are self-filing are required to file on INTIME. If these sellers choose to change to a different model, the seller will automatically become a monthly filer."
- **Grade:** 1 plain
- **Rhea's default:** both businesses register in Indiana through INBiz. The SST question opens when a second member state arrives (Ohio, in another section). The model chosen is a company fact that changes Indiana filing. Model 1 means a certified service provider files for the seller, and Indiana then becomes monthly by SER on the 20th. Model 4 means filing yourself, and Indiana stays on INTIME.
- **Rhea today:** missing.
- **Where the work lives:** *pack data* (the registration and model as facts). Working with a certified service provider is *later*.
- **Deadline:** none for the choice itself.
- **Watch:** the DOR's SST Business Rules PDF (in.gov/dor/files/sst-business-rules.pdf, "Last revised: September 2026", revised in place with no version index). The SST Governing Board's public notices and meeting minutes at streamlinedsalestax.org.
- **Sources:** https://www.in.gov/dA/778f09c373/sib91.pdf ; https://www.in.gov/dor/files/sst-business-rules.pdf ; https://www.in.gov/dor/i-am-a/business-corp/sales-tax/streamlined-sales-tax/

---

### STX-06 — Can we buy inventory, materials and packaging without paying tax?
- **Applies to:** everyone
- **The law:**
  - Resale, IC 6-2.5-5-8(b): exempt when the purchaser acquires property "for resale, rental, or leasing in the ordinary course of the person's business without changing the form of the property".
  - Materials, IC 6-2.5-5-6: exempt when the purchaser "acquires it for incorporation as a material part of other tangible personal property" that it manufactures, assembles or processes for sale.
  - Packaging, IC 6-2.5-5-9(d): wrapping material and empty containers acquired for use as nonreturnable packages for "selling the contents that the person adds".
  - Who may claim, IC 6-2.5-8-8(b): "The following are the only persons authorized to issue exemption certificates: (1) Retail merchants, wholesalers, and manufacturers, who are registered with the department under this chapter".
  - Forms: a wholesaler may use "an Indiana General Sales Tax Exemption Certificate (ST-105) or a Streamlined Sales Tax Agreement Certificate of Exemption" (SIB 52). A drop shipper may accept the reseller's ST-105 or F0003 (SIB 57).
  - Taking stock out for own use makes it taxable. ST-103 line 7: use tax is due if an item is "(a) removed from inventory for personal use, (b) used as a giveaway, or (c) used as a display model or promotional item not for sale".
- **Grade:** 1 plain
- **Rhea's default:**
  - Hoosier Supply gives its Indiana suppliers a blanket ST-105 for resale stock.
  - Maple Lane gives an ST-105 for materials that become part of its goods (fabric, wood, wax) and for nonreturnable packaging. Its tools, supplies and office goods are taxable.
  - Sales tax paid on a purchase is part of the purchase's cost; it is never a recoverable credit.
  - A stock withdrawal for own use, gifts or display raises use tax at cost (STX-07).
- **Rhea today:**
  - The shared `purchase_invoice` carries `vat` and `vat_rate`, and the Poland pack posts that VAT to 221 "VAT naliczony", a recoverable input tax. A US pack must post the tax to inventory or expense instead.
  - `goods_issue` (testdata/seed/warehouse.json) has a `reason` field that can mark own-use withdrawals.
  - No object records the certificate *we issued*.
- **Where the work lives:** *pack data* (posting rules, an issued-certificate record). The VAT-named fields `vat`/`vat_rate` on the shared purchase type are worth renaming to a neutral "tax".
- **Deadline:** the certificate at the time of purchase.
- **Watch:** IC 6-2.5-5 changes by session (synopsis). SIB 52 and SIB 57 on the bulletin index.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-5-8/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-5-6/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-5-9/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-8-8/ ; https://www.in.gov/dA/d88598eb06/sib52.pdf ; https://www.in.gov/dA/9aaab8e691/sib57.pdf ; https://www.in.gov/dA/d85d002b5d/st-103inst.pdf

---

### STX-07 — What do we owe when an out-of-state seller charges us no Indiana tax?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-3-2(a): the use tax "is imposed on the storage, use, or consumption of tangible personal property in Indiana" "if the property was acquired in a retail transaction", "regardless of the location of that transaction or of the retail merchant making that transaction".
  - Rate, IC 6-2.5-3-3: the tax "is imposed at the same rates as the state gross retail tax under IC 6-2.5-2-2". Credit, IC 6-2.5-3-5: a credit "equal to the amount, if any, of sales tax, purchase tax, or use tax paid to another state". IC 6-2.5-3-6(b): the user is personally liable.
  - Since 1 July 2025 (IC 6-2.5-3-11, SEA 453-2025, per the DOR's 2025 synopsis), a buyer of taxable services or other non-goods items who does not remit the full tax is "liable for the tax not collected".
  - ST-103 line 7: "Use tax is due on any purchase(s) where no sales tax was paid and the property was not held for resale or for another exempt purpose … Multiply your cost by the current use tax rate."
  - DOR Business FAQ: a registered merchant puts it "on its next sales tax return"; others use "Form ST-115". If 6% was paid elsewhere, "The manufacturer owes an additional 1% Indiana use tax".
- **Grade:** 1 plain
- **Rhea's default:** each purchase records the tax the seller charged and for which state. Use tax is self-assessed when a taxable item is first used in Indiana with no Indiana tax charged: 7% × cost, less tax paid to another state, never below zero. The posting is Dr the cost, Cr use tax payable, and it is reported on ST-103 line 7 for the period of receipt (simplest reading). Stock withdrawals (`goods_issue` with reason own use, gift or display) accrue at cost.
- **Rhea today:** missing. The rule shape is a purchase that raises a tax the buyer owes itself, the same shape Poland's import-of-services rule needs (VAT-03 found that one missing too). Here it carries no input credit.
- **Where the work lives:** *pack data*
- **Deadline:** with the ST-103 for the period.
- **Watch:** IC 6-2.5-3 changes by session. The ST-103 instructions on the DOR sales-tax forms page; the PDF URL carries a content hash (/dA/<hash>/st-103inst.pdf), so a new hash means a new revision.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-3-2/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-3-3/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-3-5/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-3-6/ ; https://www.in.gov/dor/files/legislative-synopsis-2025.pdf ; https://www.in.gov/dA/d85d002b5d/st-103inst.pdf ; https://www.in.gov/dor/i-am-a/business-corp/business-faq/

---

### STX-08 — The owner buys online for himself: whose return carries that use tax?
- **Applies to:** sole proprietor
- **The law:** the DOR Business FAQ says an individual "will pay the use tax owed on his or her annual Indiana income tax return (due April 15). Use tax may also be remitted directly to DOR as purchases occur using the ST-115". An individual can report "on their Individual Income Tax Returns (IT-40, Schedule 4 …)". The LLC is disregarded for income tax, but as a registered retail merchant it reports the *business's* use tax on its ST-103 (STX-07).
- **Grade:** 1 plain
- **Rhea's default:** Rhea books the business only. The owner's private purchases stay outside the books. A business purchase the owner paid personally is booked as the business's (with use tax on line 7 if due). Stock the owner takes home is a withdrawal at cost (STX-07). The IT-40 line belongs to the owner's personal return.
- **Rhea today:** the people pack knows the owner as a person; there is no personal tax return, which is right.
- **Where the work lives:** *later* (the owner's personal return). The business side is STX-07.
- **Deadline:** 15 April 2027 for 2026 purchases, per the FAQ's "due April 15".
- **Watch:** the IT-40 instructions, revised yearly. They are on forms.in.gov, which refused my fetch; the DOR's individual-forms page links them.
- **Sources:** https://www.in.gov/dor/i-am-a/business-corp/business-faq/ (IT-40 Schedule 4 and ST-115 not opened: forms.in.gov returned 403)

---

### STX-09 — Is the software we pay for, and the digital goods we sell, taxed?
- **Applies to:** everyone
- **The law:**
  - Remote access, IC 6-2.5-4-16.7(b): a purchase of "the right to remotely access" prewritten software "is not considered to be a transaction in which prewritten computer software is delivered electronically" and "does not constitute a retail transaction".
  - Downloads, SIB 8 (January 2023): "prewritten computer software downloaded by Indiana customers, whether sold, rented, leased, or licensed for consideration, is subject to Indiana sales or use tax". A download "merely incidental to the provision of services (less than 10% of the total price of the transaction)" leaves the service untaxed.
  - Specified digital products, IC 6-2.5-4-16.4(b): a retail transaction when the seller "electronically transfers specified digital products to an end user" and "grants to the end user the right of permanent use", which SIB 93 renders as "not conditioned upon continued payment by the purchaser".
  - SIB 93 (April 2023): "Indiana imposes sales and use tax on products transferred electronically only if the products meet the definition of 'specified digital products,' 'prewritten computer software,' or 'telecommunication services.'" Specified digital products are digital audio works, audiovisual works and "works that are generally recognized in the ordinary and usual sense as books".
  - The 2025 and 2026 DOR synopses list no change to any of this.
- **Grade:** 2 settled. The statutes decide software. SIB 93 settles that other electronically delivered items are outside the tax. That a printable sewing pattern or art file is not a "book" is my reading.
- **Rhea's default:** Hoosier Supply's SaaS subscriptions are not taxable and raise no use tax. A downloaded software licence is taxable, with use tax if the vendor charged none. Maple Lane's downloadable patterns and printables are not taxable in Indiana. An e-book sold for permanent use is taxable; subscription or streaming access is not.
- **Rhea today:** missing. No item carries a taxability; the invoice's single `vat_rate` is the only tax treatment Rhea knows.
- **Where the work lives:** *pack data*: item taxability categories, ideally the SST Library of Definitions codes.
- **Deadline:** none.
- **Watch:** SIB 8 and SIB 93 on the bulletin index. Indiana's software and digital-product lines in the SST taxability matrix (sst.streamlinedsalestax.org/TM). SSUTA §328 C requires states to notify the Board of changes, with seller relief until the first of the month at least 30 days later. Bills to tax SaaS would come in the 2027 long session; none is enacted today.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-4-16-7/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-4-16-4/ ; https://www.in.gov/dA/cccbbb7556/sib08.pdf ; https://www.in.gov/dA/94e37045c3/sib93.pdf ; https://www.in.gov/dor/files/legislative-synopsis-2026.pdf

---

### STX-10 — What does Indiana tax, and at what rate?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-2-1(a): "An excise tax, known as the state gross retail tax, is imposed on retail transactions made in Indiana." (b): "The person who acquires property in a retail transaction is liable for the tax on the transaction", and the merchant must "collect the tax as agent for the state".
  - IC 6-2.5-2-2(a): the tax is "imposed at seven percent (7%) of that gross retail income".
  - SIB 93: "Except for certain enumerated services, sales of services generally are not retail transactions". Tangible personal property "includes electricity, water, gas, steam, and prewritten computer software" (SIB 8).
  - DOR SST Business Rules: "Currently, Indiana is a single-rate state and has no special taxing rules. Indiana's sales tax rate is 7%." The rate was 6% from December 2002 to March 2008 (DOR Business FAQ).
  - Local food-and-beverage taxes exist for food and beverage sales (GIB 203, not read); neither business sells prepared food.
- **Grade:** 1 plain
- **Rhea's default:** one Indiana state rate, 7.000% from 2008-04-01, open-ended. Taxability is decided per item category and per line, never per invoice.
- **Rhea today:** wrong shape. The shared `vat_rate` (testdata/seed/finance_v4.json) is `percent: int` with no validity dates, and `sales_invoice.vat_rate` is required and single per invoice. Indiana's 7 fits an integer. Ohio's 5.75% with county rates and California's 7.25% with district taxes do not. A US sales tax is a rate per jurisdiction applied line by line, not a rate code per document.
- **Where the work lives:** *shared ground*: an integer percent and one rate per invoice (the VAT-invoice shape). The Indiana rate itself is *pack data*.
- **Deadline:** none.
- **Watch:** the rate changes only by statute (IC 6-2.5-2-2). SSUTA §304 A asks member states to "Limit the effective date of a rate change to the first day of a calendar quarter". The SST rate-and-boundary file pages (streamlinedsalestax.org/for-businesses/rate-and-boundary-files) publish member-state rates.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-2-1/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-2-2/ ; https://www.in.gov/dA/94e37045c3/sib93.pdf ; https://www.in.gov/dor/files/sst-business-rules.pdf ; https://www.in.gov/dor/i-am-a/business-corp/business-faq/

---

### STX-11 — Are our design services taxable?
- **Applies to:** sole proprietor
- **The law:**
  - SIB 94 (March 2023): "except for certain enumerated services, professional services, personal services, and services in respect to property not owned by the person rendering such services are not 'transactions of a retail merchant constituting selling at retail'".
  - SIB 92 (June 2025): gross retail income includes "any 'bona fide' charges which are made for preparation, fabrication, alteration, modification, finishing, completion, delivery, or other service performed in respect to the property transferred before its transfer". By contrast, "Separately stated charges for installation that occurs after delivery and transfer of the tangible personal property are not subject to sales tax". Example 12: a carpenter's labour building cabinets is taxable; separately stated installation after delivery is not.
  - A bundled transaction excludes "a retail sale in which a service is the true object of the transaction", and sales where the taxable product does not exceed 10% of the price (SIB 90, on SEA 258-2006).
- **Grade:** 2 settled. The code has no services clause. SIB 92 and SIB 94 settle the line: design as such is a nontaxable professional service, while design that ends in an object Maple Lane makes and sells is a pre-transfer service charge, taxable with the object even when separately stated.
- **Rhea's default:** a design job whose deliverable is the design (a digital file, a consultation) is nontaxable. When Maple Lane then makes and sells the designed object, the design charge is taxable with it. A mixed invoice keeps the design and the goods on separate lines.
- **Rhea today:** missing. There are no invoice lines and no tax treatment per line.
- **Where the work lives:** *shared ground*: invoice lines, each with a tax treatment. `ObjectType` fields are scalar or ref only (internal/core/core.go), so lines must be objects of their own. The rules are then *pack data*.
- **Deadline:** none.
- **Watch:** SIB 92 and SIB 94 on the bulletin index.
- **Sources:** https://www.in.gov/dA/2c575a6954/sib94.pdf ; https://www.in.gov/dA/92b275c1e1/sib92.pdf ; https://www.in.gov/dA/52e9b05d52/sib90.pdf

---

### STX-12 — Which sales are Indiana sales: the warehouse pickup, the parcel to Ohio?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-13, as SIB 96 (March 2024) states it:
    1. "When the product is received by the purchaser at a business location of the seller, the sale is sourced to that business location."
    2. "When the product is not received by the purchaser at a business location of the seller, the sale is sourced to the location where receipt by the purchaser … occurs, including the location indicated by instructions for delivery … known to the seller."
    3. Then, in order, an address in the seller's records, the address given at the sale (including the payment instrument's), and the ship-from address.
  - "The terms 'receive' and 'receipt' do not include possession by a shipping company on behalf of the purchaser", and "contractual terms relating to transfer of title or risk of loss are not relevant".
  - Example 2: a tool shipped from Indianapolis to Kentucky "is subject to Kentucky's sales tax".
  - ST-103 line 1 includes "retail, wholesale, manufacturing, and out-of-state sales".
- **Grade:** 1 plain
- **Rhea's default:** every sale records a place of receipt: the seller's location for counter, dock and booth sales; the ship-to address for shipments; for customer-arranged freight, the delivery address the seller knows, then the account address, payment address and ship-from in that order. Indiana tax applies only when receipt is in Indiana. Out-of-state sales go into ST-103 line 1 and out again on line 2. An Ohio reseller collecting at Hoosier Supply's dock makes an Indiana sale, and its exemption needs a certificate (STX-19).
- **Rhea today:** missing. Invoices carry no ship-to or place of receipt, and Poland's place-of-supply codes (WDT, export) are missing too (VAT-05).
- **Where the work lives:** *shared ground*: a place of receipt on every sale (the one-country invoice assumption). The sourcing ladder is *pack data*.
- **Deadline:** none.
- **Watch:** sourcing changes come through SSUTA amendments (Governing Board meetings) and then the General Assembly. SIB 96 on the bulletin index.
- **Sources:** https://www.in.gov/dA/f6b43ccd68/sib96.pdf ; https://www.in.gov/dA/d85d002b5d/st-103inst.pdf

---

### STX-13 — Is shipping and handling taxed?
- **Applies to:** everyone
- **The law:**
  - SIB 92 (June 2025), on IC 6-2.5-1-5: "'Delivery charges' consist of transportation, shipping, crating, handling, and packing, whether or not separately stated".
  - Postage is the exception: "separately stated postage charges do not constitute gross retail income and are not subject to sales tax". Postage means "the purchase price of stamps or similar charges, as incurred by the seller on behalf of its customers, for mail or parcel delivery through the United States mail, without any additional amounts added". It does not cover "any means other than through the United States mail" or "any amount representing a 'markup'".
  - Examples 1 and 2: $100 of goods plus $10 private-carrier delivery is taxed on $110. The same $10 by USPS, separately stated, is taxed on $100.
  - Mixed shipments: "If some of the items being delivered are exempt and some of the items are subject to sales tax, the delivery charge attributable to the taxable items is subject to tax."
  - "delivery charges billed and furnished by a third party are exempt".
- **Grade:** 2 settled. IC 6-2.5-1-5 defines delivery charges as part of gross retail income; SIB 92 settles postage, mixed shipments and third-party carriers. (The statute's own text was not read; the quotes are the bulletin's.)
- **Rhea's default:** a shipping or handling line is taxable when the goods are. A separately stated USPS postage line at exact cost is excluded. A flat "shipping & handling" or another carrier's charge is taxable. On a mixed basket, the taxable share follows the taxable share of the goods' price (SIB 92 Example 5 uses the exempt percentage; that it is by price is my reading).
- **Rhea today:** missing. A web-shop order arrives with one "shipping" amount and Rhea has nowhere to keep its kind.
- **Where the work lives:** *pack data*: charge types on lines and an allocation formula with declared rounding, on the shared lines of STX-11.
- **Deadline:** none.
- **Watch:** SIB 92 on the bulletin index (last revised June 2025). The delivery-charge entries in Indiana's SST taxability matrix.
- **Sources:** https://www.in.gov/dA/92b275c1e1/sib92.pdf ; https://www.in.gov/dA/52e9b05d52/sib90.pdf

---

### STX-14 — How do discounts, coupons and early-payment terms change the tax?
- **Applies to:** everyone
- **The law:**
  - SIB 58 (November 2022), on IC 6-2.5-1-5: "discounts (including cash, terms, or coupons) that are allowed by a seller are not included within gross retail income if they are not reimbursed by a third party and they are taken by a purchaser on a sale".
  - Terms and cash discounts: "with regard to time discounts (e.g., 2% discount for payment within 10 days) … only the actual amount received by the merchant is subject to Indiana sales tax. If the consumer is not actually given the discount, then sales tax must be collected on the full price paid."
  - Third-party discounts stay in the base when four conditions hold: the seller receives consideration from a third party, must pass the reduction on, can fix the amount at the sale, and the reduction is identified as a third-party one.
  - Example 1: a $1.00 item with a $0.20 manufacturer's coupon is taxed on $1.00. Example 2: a $0.25 store coupon on a $1.00 item leaves $0.75 taxable.
- **Grade:** 2 settled by SIB 58 (the bulletin applies the definition of gross retail income to each kind of discount).
- **Rhea's default:** a seller's discount or store coupon lowers the taxable base on the document. A manufacturer's coupon does not. Hoosier Supply invoices tax on the full price. When a customer takes a prompt-payment discount, Rhea drafts a credit for the discount and the tax on it, in the period it is taken (accrual). On the cash basis, the tax simply follows the money received.
- **Rhea today:** partly there. The matching engine already recognises a short payment within terms (the Poland pack's `matching_difference`). Splitting off the tax part of that difference is missing.
- **Where the work lives:** *pack data*
- **Deadline:** none.
- **Watch:** SIB 58 on the bulletin index.
- **Sources:** https://www.in.gov/dA/1db827e68a/sib58.pdf ; https://www.in.gov/dA/92b275c1e1/sib92.pdf

---

### STX-15 — What must our invoice or receipt show?
- **Applies to:** everyone
- **The law:**
  - 45 IAC 2.2-9-5(a), subject to the gasoline exceptions in IC 6-2.5-7, makes it unlawful to "(1) Display an advertised price, marked price, or publicly stated price that includes the state gross retail or use taxes" or "(2) Offer to assume or absorb part of a customer's state gross retail or use tax on a sale". (b): the person "commits a Class B infraction".
  - 45 IAC 2.2-2-2: "The retail merchant, acting as an agent for the state of Indiana, must collect the tax. The tax is bourne by the customer."
  - Each exclusion from the base holds only when separately stated: USPS postage, installation, interest and financing charges, taxes imposed on the consumer, food-serving charges and trade-in value. SIB 92: they "must be separately stated … If they are not, then the charge is subject to sales tax."
  - No Indiana section lists the contents of an invoice, as the Polish VAT act does in art. 106e.
- **Grade:** 2 settled. The code is silent on invoice contents. 45 IAC 2.2-9-5 bars tax-included prices, and SIB 92 ties each exclusion to separate statement. The rule was filed in 1982 and read via Cornell LII; I did not confirm against the Indiana Register that it is still in force.
- **Rhea's default:** every sales document shows each line's price and treatment, the tax as its own amount, and each excludable charge on its own line. An exempt sale names its certificate. No tax-included price tags at the fair: prices are pre-tax and the tax is added (cash totals may then be rounded, STX-17).
- **Rhea today:** missing. Rhea's writers render a VAT invoice (KSeF FA(3)) through `adapter.ProvideWriter` and the issued-document door; nothing renders a US receipt.
- **Where the work lives:** *market code*: a receipt and invoice writer in markets/us behind the writer seam. Its content rules are *pack data*.
- **Deadline:** at the sale.
- **Watch:** the Indiana Register (iar.iga.in.gov) for re-adoption or repeal of 45 IAC 2.2-9 (not reached here). SIB 92 on the bulletin index.
- **Sources:** https://www.law.cornell.edu/regulations/indiana/45-IAC-2.2-9-5 ; https://www.law.cornell.edu/regulations/indiana/45-IAC-2.2-2-2 ; https://www.in.gov/dA/92b275c1e1/sib92.pdf

---

### STX-16 — Do we compute tax per line or per invoice, and how do we round?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-2-2(b): "If the tax computed under subsection (a) carried to the third decimal place results in the numeral in the third decimal place being greater than four (4), the amount of the tax shall be rounded to the next additional cent."
  - (c): "A seller may elect to round the tax under subsection (b) on a transaction on an item basis or an invoice basis. However, a seller may not round the tax under subsection (b) to circumvent the tax that would otherwise be imposed on a transaction using an invoice basis."
  - SSUTA §324 B: states "shall allow sellers to elect to compute the tax due on a transaction on an item or an invoice basis" and "No member state shall require a seller to collect tax based on a bracket system."
  - SIB 90: the brackets were removed from 1 July 2010; the item or invoice election dates from 1 July 2014.
- **Grade:** 1 plain
- **Rhea's default:** the invoice basis: tax = `round(taxable total × 0.07, 2, half_up)`, computed once per document. Displayed line taxes are allocated so they sum to that figure. The election is a dated company fact. On marketplace orders, the facilitator's figures are taken as given.
- **Rhea today:** right for the arithmetic. The formula language's `round(x, 2, half_up)` is the statute's rule ("third decimal place … greater than four"), and declared rounding is already enforced. Missing: the election as a company fact. Also to check: how `half_up` treats a negative amount on a credit.
- **Where the work lives:** *pack data*
- **Deadline:** none.
- **Watch:** HEA 1406-2026 §26 (effective 1 January 2027) amends SEA 243 §112 "by removing reference to IC 6-2.5-2-2 and IC 23-15-13" (DOR synopsis; what §112 itself says was not read). SSUTA §324 amendments by the Governing Board.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-2-2/ ; https://www.streamlinedsalestax.org/docs/default-source/agreement/ssuta/ssuta-as-amended-through-12-19-25-with-hyperlinks-and-compiler-notes-at-end.pdf ; https://www.in.gov/dA/52e9b05d52/sib90.pdf ; https://www.in.gov/dor/files/legislative-synopsis-2026.pdf

---

### STX-17 — Cash at the fair with no pennies: may we round the total to a nickel?
- **Applies to:** everyone (in practice Maple Lane's fair takings)
- **The law:** per the DOR's 2026 Legislative Synopsis (the enrolled acts on iga.in.gov could not be read):
  - IC 23-15-13 (SEA 243-2026 §106, made permissive by HEA 1406 §18; in force 15 March 2026 under HEA 1406 §24). For a cash "total transaction amount" whose second decimal is not 0 or 5, a business "may round the total transaction amount" down, up, or to the nearest $0.05.
  - IC 6-2.5-1-5 (SEA 243 §2, from 15 March 2026): "amounts added or subtracted by a seller to comply with IC 23-15-13 shall not be considered in determining gross retail income". HEA 1406 §25 limits both to "cash transactions occurring after March 14, 2026".
  - The DOR's notice of 6 March 2026: "Rounding is for cash transactions only. Rounding must take place on the total price including tax." On a $6.42 total, "the tax of 42 cents is remitted to the State regardless of rounding", and the gain or loss is "added to or taken away from the retailer's income".
  - Enacted, in force 1 January 2027: IC 5-36.5 (SEA 243 §1). Amounts payable to the state are rounded "downward to the next amount divisible by $0.05". For a tax remitted by an agent, "The aggregate amount … remitted …, reduced by any collection allowances …, shall be subject to the rounding provisions".
- **Grade:** 2 settled. The cash rounding and its exclusion from the tax base are settled by IC 6-2.5-1-5 and the DOR notice. How IC 5-36.5 will appear on the ST-103 from 2027 is not yet published; I found no DOR guidance.
- **Rhea's default:** at the fair, the tax is computed on the pre-tax total (STX-16). If the owner has chosen to round, only the cash total is rounded. The difference goes to a "cash rounding" account, never to sales tax payable. Card sales are never rounded. From 1 January 2027, if INTIME cuts the remittance to the nickel, the cents go to the same account.
- **Rhea today:** missing. There is no cash-sale document. Rounding to $0.05 can be written in the formula language today as a declared rounding of x × 20 back to whole units.
- **Where the work lives:** *pack data*
- **Deadline:** cash sales after 14 March 2026; remittances from 1 January 2027.
- **Watch:** DOR email notices through subscription.in.gov (the 6 March notice is in.gov/dor/files/penny-rounding-for-retailers.pdf). The 2027 ST-103 form revision on the sales-tax forms page. The DOR's 2027 synopsis.
- **Sources:** https://www.in.gov/dor/files/legislative-synopsis-2026.pdf ; https://www.in.gov/dor/files/penny-rounding-for-retailers.pdf

---

### STX-18 — Is the tax we collect ours to use, and who answers if it is not paid over?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-2-1(b): the merchant collects "as agent for the state". 45 IAC 2.2-6-8(b): the computed tax is the liability "regardless of the amount of tax he actually collects".
  - In force 1 July 2026, per the DOR synopsis:
    - IC 6-8.1-1-4.7 defines "taxes held in trust" as a listed tax "collected or received by a taxpayer from the taxpayer's customer" which "upon receipt or accrual becomes property of the state".
    - IC 6-8.1-1-11 makes "a responsible person" any person who "is an individual conducting business as a sole proprietor or an employee, contractor, officer, or member" with a duty to remit.
    - IC 6-8.1-8-18: responsible persons "are personally liable", and one who "knowingly fails to collect or remit those taxes … commits a Level 6 felony". The business and each responsible person are "jointly and severally liable".
- **Grade:** 1 plain
- **Rhea's default:** collected tax is credited to "sales tax payable — Indiana" at the sale (or at receipt, on the cash basis), never to revenue. Only the return's payment clears it. The owner (Maple Lane) and the officer who files (Hoosier Supply) are recorded as responsible persons. The host warns when the bank balance falls below the payable.
- **Rhea today:** right in shape. The double-entry engine and the Poland pack already keep output VAT on its own account (222). The people pack can name the responsible person.
- **Where the work lives:** *pack data*
- **Deadline:** none of its own; the return dates of STX-04.
- **Watch:** IC 6-8.1 changes in the DOR's yearly synopsis (the 2026 edition carries the trust-tax package, SEA 243 §§66–77).
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-2-1/ ; https://www.law.cornell.edu/regulations/indiana/45-IAC-2.2-6-8 ; https://www.in.gov/dor/files/legislative-synopsis-2026.pdf ; https://www.in.gov/dA/f031b3065a/gb102.pdf

---

### STX-19 — A reseller says it is buying for resale: what must we get, check and keep?
- **Applies to:** corporation
- **The law:**
  - IC 6-2.5-8-8(a): an authorized buyer "may issue an exemption certificate to the seller instead of paying the tax"; a seller holding one "has no duty to collect or remit".
  - (b) lists "the only persons authorized": "(1) Retail merchants, wholesalers, and manufacturers, who are registered with the department under this chapter"; certain utility certificate holders; and "(3) Other persons who are exempt from the state gross retail tax with respect to any part of their purchases".
  - (d) allows "a blanket exemption certificate to cover exempt purchases over a stated period of time".
  - Incomplete certificates, (e): the seller is relieved only if it obtains "a fully completed exemption certificate" or "the relevant data" "within ninety (90) days after the sale". (f): after a DOR request, "within one hundred twenty (120) days", or it must prove the exemption by other means.
  - SSUTA §317:
    - B: no relief "to a seller who fraudulently fails to collect tax; to a seller who solicits purchasers to participate in the unlawful claim of an exemption".
    - D: good faith means the exemption "(i) was statutorily available …, (ii) could be applicable to the item being purchased, and (iii) is reasonable for the purchaser's type of business".
    - F: no renewal of blanket certificates may be asked while "a period of no more than twelve months elapses between sales transactions".
  - SST form F0003: "The seller is not required to verify the purchaser's ID number". A buyer not registered in the claiming state may give "your sales tax ID number issued by any state".
  - SIB 9: the F0003 "should be accepted by Indiana retailers and can be used by nonresident purchasers". IC 6-2.5-9-1: issuing or accepting a certificate to evade the tax is a Class B misdemeanor.
- **Grade:** 2 settled. IC 6-2.5-8-8(b)(1) reads as if only Indiana-registered merchants may claim resale. SSUTA §317, the F0003 instructions Indiana signs up to, and SIB 9 settle that an out-of-state reseller may use F0003 with its home-state number.
- **Rhea's default:**
  - Hoosier Supply keeps each customer's certificates as `exemption_certificate` objects: form (ST-105, F0003, NP-1), reason, ID number and issuing state, single or blanket, date received, and the file as a `received_document`.
  - A sale to a customer without a valid certificate is taxed.
  - A certificate missing at the sale opens a 90-day task.
  - A blanket certificate lapses after 12 months without a sale (the SST rule; the ST-105's own terms could not be read).
  - Each certificate is checked for plausibility (the reason fits the buyer's trade and the items), not against a registry.
- **Rhea today:** missing as an object. `received_document` (packs/common/documents.json) can hold the file and the work queues can carry the 90-day chase. Nothing links a certificate to a sale.
- **Where the work lives:** *pack data*
- **Deadline:** at the sale or within 90 days of it; within 120 days of a DOR request.
- **Watch:** ST-105 revisions on the DOR sales-tax forms page (the form itself is on forms.in.gov). SST F0003 revisions (current form "Revised 12/21/2021") and SSUTA §317 amendments by the Governing Board. Indiana's entries in the SST Tax Administration Practices matrix (sst.streamlinedsalestax.org/TAP).
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-8-8/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-9-1/ ; https://www.streamlinedsalestax.org/docs/default-source/agreement/ssuta/ssuta-as-amended-through-12-19-25-with-hyperlinks-and-compiler-notes-at-end.pdf ; https://www.streamlinedsalestax.org/docs/default-source/forms/exemption-certificateb926a7ab4a0d43e1ad4fe8eb19e79cbb.pdf?sfvrsn=857843d_5 ; https://www.in.gov/dA/e3470c7b38/sib09.pdf (ST-105 not opened: forms.in.gov returned 403)

---

### STX-20 — Manufacturers, farmers, nonprofits and governments: which papers make their purchases exempt?
- **Applies to:** corporation
- **The law:**
  - Manufacturing, IC 6-2.5-5-3(b): machinery, tools and equipment "are exempt from the state gross retail tax if the person acquiring that property acquires it for direct use" in direct production.
  - Agriculture, IC 6-2.5-5-2(a): "Transactions involving agricultural machinery, tools, and equipment are exempt from the state gross retail tax if the person acquiring that property acquires it for the person's direct use in the direct production, extraction, harvesting, or processing of agricultural commodities".
  - Mixed use, SIB 9 (June 2025): "If the purchase does not qualify for a 100% or predominate exemption, the ST-105 should not be issued to the vendor. Instead, sales tax should be paid on the purchase, and the purchaser should file a Claim for Refund (Form GA-110L)".
  - Nonprofits, SIB 10 (July 2026): a registered nonprofit buys exempt with "a Form NP-1 (Indiana Nonprofit Sales Tax Exemption Certificate) from INTIME"; "A separate Form NP-1 needs to be generated for each vendor"; and "Form ST-105s issued prior to Jan. 1, 2023, by a qualified nonprofit registered in Indiana will expire Jan. 1, 2024". A nonprofit from another state "may use a … Form F0003". Public schools keep using the ST-105.
- **Grade:** 2 settled. The exemption tests are in the code. SIB 9 and SIB 10 settle which paper each buyer uses.
- **Rhea's default:** the certificate a sale needs follows the buyer's kind:
  - A manufacturer or farmer gives an ST-105 or F0003 with its reason, accepted in good faith with a plausibility check: a farmer's claim on a tractor part passes; on office chairs it is flagged.
  - An Indiana nonprofit gives an NP-1 naming Hoosier Supply. An ST-105 from an Indiana nonprofit is refused.
  - A nonprofit from another state gives an F0003.
  - A government buyer gives what SIB 4 prescribes (SIB 4 not read).
- **Rhea today:** missing, as in STX-19.
- **Where the work lives:** *pack data*
- **Deadline:** as STX-19.
- **Watch:** SIB 64 (manufacturing; revised September 2026, effective October 2026, per the bulletin index and the October 2026 Tax Bulletin; its PDF link did not open for me), SIB 9, SIB 10 and SIB 4. HEA 1406-2026 shows the nonprofit rules move with each session.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-5-3/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-5-2/ ; https://www.in.gov/dA/e3470c7b38/sib09.pdf ; https://www.in.gov/dA/dd8bba6308/sib10.pdf ; https://www.in.gov/dor/files/2026-october-tax-bulletin.pdf

---

### STX-21 — Is anything we sell exempt by its nature: food, bundles, gift baskets?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-5-20(a): "Sales of food and food ingredients for human consumption are exempt from the state gross retail tax." (c) leaves out candy, alcoholic beverages, soft drinks, food heated by the seller, two or more ingredients mixed by the seller, food sold with utensils, tobacco and dietary supplements.
  - SIB 70: the food exemption "does not apply to items such as candy, soft drinks and prepared food items".
  - Bundles: a bundled transaction is a sale of products that are "distinct; identifiable; and sold for one nonitemized price" (SIB 94). It excludes sales where the taxable product does not exceed 10% of the price (SIB 90). A person selling tangible property in a bundle "is making a retail transaction" (IC 6-2.5-4-15, per SIB 90).
- **Grade:** 2 settled. Food is plain. A gift basket mixing exempt food with taxable goods at one price is settled by the bundled-transaction definition (SIB 94; the 10% measure itself was not read in full).
- **Rhea's default:** the item master carries a taxability category. A one-price gift basket is taxed in full unless its taxable items are 10% or less of the price. Separately priced items are taxed each by its own category.
- **Rhea today:** missing. The warehouse `item` has only `sku`, `name` and `unit`.
- **Where the work lives:** *pack data* (an item taxability field and the bundle rule).
- **Deadline:** none.
- **Watch:** SIB 29 (food, September 2025) and SIB 94 on the bulletin index. The SST Library of Definitions entries for food.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-5-20/ ; https://www.in.gov/dA/e7027b4aa4/sib70.pdf ; https://www.in.gov/dA/2c575a6954/sib94.pdf ; https://www.in.gov/dA/52e9b05d52/sib90.pdf ; https://www.in.gov/dA/190543470d/sib29.pdf

---

### STX-22 — Our marketplace collects the tax: do we still report those sales?
- **Applies to:** sole proprietor
- **The law:**
  - SIB 89 (April 2024): "When such facilitation occurs, the marketplace facilitator is considered the retail merchant of each retail transaction that is facilitated on its marketplace". The facilitator must "Collect and remit the gross retail tax, even if a seller … does not have a registered retail merchant certificate".
  - DOR marketplace page:
    - "A marketplace facilitator is considered to be the retail merchant with regard to sales by third-party sellers on its marketplace. Such sales are not legally considered to be sales by the third-party seller for Indiana sales tax purposes."
    - "If a marketplace seller is otherwise required to register … then the seller should report marketplace sales as if they are exempt sales on the ST-103."
    - "in the event the marketplace facilitator incorrectly remits the tax to the seller, the seller should remit the tax to the DOR".
    - "The marketplace seller has no obligation to collect tax on facilitated sales even if the marketplace facilitator fails to do so."
    - The facilitator may give the seller an ST-105 marked "Other" with "Marketplace Sales".
  - ST-103 line 3: "Excluded Sales − Sales included in Line 1 where the obligation to collect the tax falls on another party such as a marketplace facilitator."
- **Grade:** 2 settled. The code makes the facilitator the merchant. The DOR page and the ST-103 instructions settle that a registered seller still reports those sales, in line 1 and out again on line 3. The FAQ's older wording says "as if they are exempt"; the current form has its own excluded-sales line.
- **Rhea's default:** marketplace orders are Maple Lane's revenue at the price the buyer paid less the tax the facilitator collected. That tax never touches Maple Lane's payable. The orders go into ST-103 line 1 and line 3. A payout that contains tax raises a task, and that tax moves to the payable to be remitted. The facilitator's ST-105 is kept on file.
- **Rhea today:** missing. There is no reader for the marketplace's order and payout report. The matching engine can pair a payout (net of fees) with its orders.
- **Where the work lives:** *market code*: a reader for the marketplace's report. Its format is the platform's, not American law, so it may sit beside the shared readers in internal/adapter rather than in markets/us. The reporting rules are *pack data*.
- **Deadline:** with each ST-103.
- **Watch:** the DOR marketplace page (unversioned HTML; snapshot it and diff); SIB 89 on the bulletin index. SEA 243-2026 provides that a facilitator is not liable under both IC 6-2.5-9-3 and the new IC 6-8.1-8-18 (from 1 July 2026; synopsis).
- **Sources:** https://www.in.gov/dA/88a60658b3/sib89.pdf ; https://www.in.gov/dor/i-am-a/business-corp/remote-sellers/marketplace-facilitators/ ; https://www.in.gov/dA/d85d002b5d/st-103inst.pdf ; https://www.in.gov/dA/d88598eb06/sib52.pdf

---

### STX-23 — Is there a sales tax holiday to plan for?
- **Applies to:** everyone
- **The law:**
  - No holiday provision appears in IC 6-2.5 as summarised by the DOR's 2025 and 2026 synopses, nor in its bulletin index. SSUTA §328 F: a state with a holiday "shall … give notice in the Library section of the taxability matrix".
  - The 2026 "gas tax holiday" (DN #45 and the Governor's orders) suspended gasoline use and excise tax, not the tax on goods.
  - A secondary source says "Indiana, however, does not currently have any scheduled sales tax holidays" (secondary — verify).
- **Grade:** 1 plain (a negative finding; I found no primary statement that none exists).
- **Rhea's default:** none shipped. A holiday would be a dated exemption by item category: the same shape as a rate with validity dates (STX-10).
- **Rhea today:** `vat_rate` has no validity dates.
- **Where the work lives:** *later*
- **Deadline:** none.
- **Watch:** the General Assembly (a holiday needs a statute); the holiday section of the SST Library of Definitions.
- **Sources:** https://www.in.gov/dor/files/legislative-synopsis-2026.pdf ; https://www.in.gov/dor/files/legislative-synopsis-2025.pdf ; https://www.salestaxhandbook.com/indiana/sales-tax-holidays (secondary)

---

### STX-24 — A customer returns goods or was overcharged: how do we correct the tax?
- **Applies to:** everyone
- **The law:**
  - SIB 89: "Although not required by law, when a purchaser makes a return of tangible personal property to a retail merchant …, the merchant will often refund the sales tax collected by the merchant. The merchant will then make an adjustment on their next sales tax return to recover from the department the sales tax that they refunded."
  - ST-103 line 2: "Enter the total amount of exempt sales and/or deductions for the period."
  - IC 6-2.5-6-13(a): a customer gets a refund from the DOR if the merchant "erroneously or illegally collects" the tax, remits it, and "does not refund the taxes to the person".
  - SSUTA §325 C: a purchaser's claim against the seller does not accrue until written notice and sixty days for the seller to respond.
- **Grade:** 2 settled. I found no code section on returned goods; SIB 89 and ST-103 line 2 settle the merchant's deduction.
- **Rhea's default:** a return is a credit document pointing to the original sale, reversing the price and the tax. It is deducted on line 2 in the period of the refund; the original period is not amended. Tax found over-collected later is refunded to the customer. While it is not refunded, it stays payable.
- **Rhea today:** partly there. The amendment primitive and unpairing can reverse a matched sale. No credit document is tied to a tax deduction.
- **Where the work lives:** *pack data*
- **Deadline:** the return for the period of the refund.
- **Watch:** SIB 89 and the ST-103 instructions.
- **Sources:** https://www.in.gov/dA/88a60658b3/sib89.pdf ; https://www.in.gov/dA/d85d002b5d/st-103inst.pdf ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-6-13/ ; https://www.streamlinedsalestax.org/docs/default-source/agreement/ssuta/ssuta-as-amended-through-12-19-25-with-hyperlinks-and-compiler-notes-at-end.pdf

---

### STX-25 — A customer never pays: can we get the tax back?
- **Applies to:** corporation (accrual basis)
- **The law:**
  - IC 6-2.5-6-9(a): the merchant may deduct receivables on which it "did not collect the state gross retail or use tax from the purchaser", "has previously paid the state gross retail or use tax liability to the department", and which were written off "as an uncollectible debt for federal tax purposes under Section 166 of the Internal Revenue Code during the particular reporting period".
  - (b): an amount later collected is included again.
  - (d)(2): the deduction excludes "Financing charges or interest", "Sales or use taxes charged on the purchase price", amounts on property the seller still holds, collection expenses and repossessed property.
  - (d)(3): it is claimed on "the return for the period during which the receivable is written off". (d)(4): if it exceeds taxable sales, the merchant files a refund claim. (d)(6): payments apply "proportionally to the taxable price of the property" and its tax first.
  - ST-103 instructions: "Include on this line any bad debts expensed during this reporting period. If you incur bad debt expense that is not written off monthly, you must file a refund claim GA-110L, on an annual basis".
- **Grade:** 1 plain
- **Rhea's default:** for accrual reporters only. The write-off event the books use for the federal §166 deduction drafts a line-2 deduction of the unpaid taxable price (excluding finance charges and the tax itself) in that period. A later recovery is split proportionally, and its taxable part re-enters gross retail income. Write-offs made once a year go to a GA-110L instead. Cash-basis Maple Lane never reported unpaid tax, so it has nothing to deduct.
- **Rhea today:** missing. The matching engine knows the open amount (the Poland pack's `sales_invoice.open`). There is no write-off event and no proportional split.
- **Where the work lives:** *pack data*
- **Deadline:** the return for the period of the write-off.
- **Watch:** IC 6-2.5-6-9 by session; SSUTA §320 (bad debts) amendments; the ST-103 instructions.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-6-9/ ; https://www.in.gov/dA/d85d002b5d/st-103inst.pdf ; https://www.in.gov/dA/52e9b05d52/sib90.pdf

---

### STX-26 — What goes on the ST-103, line by line?
- **Applies to:** everyone
- **The law:** the ST-103 instructions, line by line:
  - 1, Total Sales: "retail, wholesale, manufacturing, and out-of-state sales".
  - 2, Exemptions/Deductions, including bad debts.
  - 3, Excluded Sales (marketplace).
  - 4, Taxable Sales.
  - 5, Total Tax Due: "Multiply Line 4 by the current sales tax rate".
  - 6, Collection Allowance: "only if your return and payment were made on or before the due date".
  - 7, Use Tax Due.
  - 8, Interest: "multiply the amount due by the annual interest rate and divide the result by 365. Multiply that result by the number of days the payment is late".
  - 9, Penalty: "10% of the total of Line 5 plus Line 7, or $5, whichever is greater".
  - 10, payment previously made by EFT.
  - 11, Amount Due.
  - Also: "One Payment Per Return", and a return "must be filed … even when no tax is due". 45 IAC 2.2-6-8(a) computes the liability on total taxable gross retail income times the rate.
  - Each location files separately unless the locations are consolidated under one TID (Form BT-1C, or INTIME). Consolidated filers file an annual RP-11, "due Jan. 31" (DOR Business FAQ).
- **Grade:** 1 plain
- **Rhea's default:** a `sales_tax_return` object per Indiana account (TID and location) and period, with lines 1–11. Line 2 is broken down by cause: resale, other exempt, out-of-state, returns, discounts, bad debts. Every figure is explained by the facts it sums. Line 5 is 7% of line 4, not the sum of the invoices' rounded taxes. The difference between line 5 and the payable account comes from per-invoice rounding: it goes to a rounding account, and a large one is flagged as possible over-collection (STX-24).
- **Rhea today:** the pattern exists, the object does not. Poland's `vat_settlement` is built by folds at period close, but its fields are Polish rate buckets (`sales_net_23`, `sales_vat_23`, …).
- **Where the work lives:** *pack data* (a new settlement type in the US pack). Periods longer than a month need the shared change of STX-04.
- **Deadline:** STX-04.
- **Watch:** the ST-103 instructions on the DOR sales-tax forms page (content-hash URL /dA/<hash>/st-103inst.pdf); INTIME's return screens.
- **Sources:** https://www.in.gov/dA/d85d002b5d/st-103inst.pdf ; https://www.in.gov/dor/i-am-a/business-corp/business-faq/ ; https://www.law.cornell.edu/regulations/indiana/45-IAC-2.2-6-8

---

### STX-27 — Can Rhea file the return, or must someone key it into INTIME? How is it paid?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-6-1(e): "A retail merchant shall report and remit state gross retail and use taxes through the department's online tax filing program." DN #36: "all retail merchants are required to report and remit sales tax electronically"; exceptions only through a BT-EX.
  - The DOR's bulk-filing matrix gives ST-103 a single mark, under "INTIME (manually keyed in)". It has no spreadsheet, XML, SFTP or EDI mark. (This is my reading of the PDF's column layout.)
  - DOR SST Business Rules (September 2026):
    - "Certified SER Transmitters may file the SER for Model 4 clients and for non-Streamlined clients"; "To become a certified SER transmitter, a company must first complete testing with Indiana".
    - "SERs for non-SST clients must be filed populating the full 13-character Taxpayer ID (TID/Loc #)"; "Indiana will acknowledge the receipt of SERs and payments within two business days".
    - On payment: "Payments may be made on INTIME up to a year before the due date", by ACH debit or credit (CCD+TXP format).
  - DOR Business FAQ: "if your average monthly tax exceeds $5,000, you are required to make payments by Electronic Funds Transfer". I did not locate the statute behind the $5,000 figure.
- **Grade:** 1 plain
- **Rhea's default:** in the first year Rhea prepares the return, every figure with its explanation, and a person keys it into INTIME and pays by ACH debit there. The confirmation is kept as a received document and closes the filing duty. Later, Rhea can become a certified SER transmitter: an XML SER sent to Indiana's web service after testing. That one writer would serve every SST state that accepts SERs, Ohio among them.
- **Rhea today:** writers exist for FA(3) and JPK_V7M behind `adapter.ProvideWriter`, and markets/pl has a live KSeF API client. INTIME offers a direct filer no API.
- **Where the work lives:** *market code*: an SER XML writer and transmitter client in markets/us, after Indiana's certification testing. In the first year, *pack data* (a human keying duty).
- **Deadline:** STX-04.
- **Watch:** the DOR SST Business Rules PDF (in.gov/dor/files/sst-business-rules.pdf, revised in place). The bulk-filing matrix PDF (in.gov/dor/files/dor-efile-methods-bulk.pdf). The SER schemas, which the SST technology page points to on the FTA E-Standards site (per a search summary; not fetched — verify).
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-6-1/ ; https://www.in.gov/dA/c7308ef71f/dn36.pdf ; https://www.in.gov/dor/files/dor-efile-methods-bulk.pdf ; https://www.in.gov/dor/files/sst-business-rules.pdf ; https://www.in.gov/dor/i-am-a/business-corp/business-faq/

---

### STX-28 — How much may we keep as a collection allowance?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-6-10(a): a merchant may deduct, "if timely remitted, a retail merchant's collection allowance".
  - (b) sets the rate by the merchant's total liability for "the state fiscal year ending on June 30 of the immediately preceding calendar year":
    - "Seventy-three hundredths percent (0.73%)" up to "sixty thousand dollars ($60,000)";
    - "Fifty-three hundredths percent (0.53%)" above that up to "six hundred thousand dollars ($600,000)";
    - "Twenty-six hundredths percent (0.26%)" above that.
  - ST-103 line 6: "Use this line only if your return and payment were made on or before the due date. Your Collection Allowance for this period is ___% of Line 5". The allowance is on line 5 only, not on use tax.
  - Enacted for 1 January 2027: the remittance after the allowance is rounded down to $0.05 (IC 5-36.5, STX-17).
- **Grade:** 1 plain. A new business had no liability in the measuring fiscal year, so its first rate is 0.73% (my reading of (b); the DOR prints the percentage on the return).
- **Rhea's default:** the rate is a dated fact per account and calendar year, taken from the return. Allowance = `round(line 5 × rate, 2, half_up)`; the statute does not say how to round it, so this is our declared choice. It is booked as income (Dr sales tax payable, Cr collection allowance) only when payment is on time, and a late payment reverses it.
- **Rhea today:** missing; it fits the formula language as it stands.
- **Where the work lives:** *pack data*
- **Deadline:** the allowance is lost if the return or payment is late.
- **Watch:** IC 6-2.5-6-10 by session. The ST-103 form. The DOR deleted its departmental notice on the allowance (DN #27) in August 2023, so the statute is the only source.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-6-10/ ; https://www.in.gov/dA/d85d002b5d/st-103inst.pdf ; https://www.in.gov/dor/legal-resources/tax-library/departmental-notices/

---

### STX-29 — What does a late or missed return cost?
- **Applies to:** everyone
- **The law:**
  - Interest, IC 6-8.1-10-1. DN #3 (October 2025): "The rate of interest for an underpayment of tax and an excess tax payment for calendar year 2026 will be 7%." The rate is the state general fund's yield plus two points, rounded, and is set each autumn (IB 64).
  - The 2027 rate is not verified. The DOR's notice index lists a DN #3 revision published 09/26 and effective 01/27, and the October 2026 Tax Bulletin announces "Departmental Notice 3: Interest Rates for Calendar Year 2027". The PDF served on 16 October 2026 still read 2026.
  - Penalty, IC 6-8.1-10-2.1(b): "ten percent (10%)" of the unpaid tax. (g): a late return showing no tax costs "ten dollars ($10) for each day that the return is past due", up to $250. (d): waived where the failure "was due to reasonable cause and not due to willful neglect".
  - The ST-103 instructions add a $5 minimum, and a 20% penalty after a Notice of Failure to File. The DOR's sales-tax page: "Late filed returns are subject to a penalty of up to 20%, and a minimum penalty of $5."
  - Unpaid liabilities lead to RRMC revocation (DOR sales-tax page), and selling without a certificate is a misdemeanour (STX-01). Personal liability and the felony: STX-18.
- **Grade:** 1 plain
- **Rhea's default:** a missed due date raises a task the same day. Rhea drafts the late return with interest computed daily at that year's rate, the 10%/$5 penalty, and no collection allowance. The rate for each year is master data from DN #3: 7% for 2026; 2027 to be loaded once the notice is read.
- **Rhea today:** the clock and tasks are there. Counting the days late needs date-difference arithmetic, which the formula language lacks (STX-04).
- **Where the work lives:** *pack data*, on the shared day arithmetic of STX-04.
- **Deadline:** interest runs from the due date to the payment date.
- **Watch:** DN #3 each autumn on the departmental-notices index (a table with publication and effective dates). The DOR monthly Tax Bulletin.
- **Sources:** https://www.in.gov/dA/e466396cf5/dn03.pdf ; https://www.in.gov/dA/1e0e308887/ib64.pdf ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-8-1-10-2-1/ ; https://www.in.gov/dA/d85d002b5d/st-103inst.pdf ; https://www.in.gov/dor/i-am-a/business-corp/sales-tax/ ; https://www.in.gov/dor/files/2026-october-tax-bulletin.pdf

---

### STX-30 — We overpaid the state: how do we claim it back, and by when?
- **Applies to:** everyone
- **The law:**
  - IC 6-8.1-9-1(a), as GIB 100 (September 2024) explains it: the claim must be filed "within three years after the later of" the return's due date, the date of payment, or, for a periodic tax, "31 days after the end of the calendar year that contains the taxable period". Example 1: monthly 2021 returns can be claimed until "January 31, 2025".
  - The claim goes on Form GA-110L or through INTIME. GIB 101: interest accrues if the excess is not refunded within 90 days. For sales tax, "the due date of the tax is December 31 of the year containing the taxable period".
- **Grade:** 1 plain
- **Rhea's default:** an overpayment found after filing (a duplicate payment, an error) becomes a claim with its own deadline. For every 2026 period filed and paid on time, that deadline is 31 January 2030 (computed per GIB 100, Example 1).
- **Rhea today:** missing.
- **Where the work lives:** *pack data*
- **Deadline:** three years, as above.
- **Watch:** GIB 100 and GIB 101 on the bulletin index; IC 6-8.1-9 in the DOR synopsis.
- **Sources:** https://www.in.gov/dA/fbc4676f2e/gb100.pdf ; https://www.in.gov/dA/dae39a0c16/gb101.pdf

---

### STX-31 — How long must we keep sales records and exemption certificates?
- **Applies to:** everyone
- **The law:**
  - IC 6-8.1-5-4: every taxpayer "must keep books and records so that the department can determine the amount" of its liability, source documents included, "for a period of at least three (3) years after the date the final payment of the particular tax liability was due". The period is unlimited if no return was filed and longer if the assessment period is extended. The taxpayer "must allow inspection of the books and records and returns by the department".
  - IC 6-8.1-5-2, as SIB 56 explains it: an assessment must issue "within three years of the end of the calendar year which contains the taxable period". There is no time limit for a fraudulent, unsigned, substantially blank or unfiled return.
  - SSUTA §317 A6: "The seller shall maintain proper records of exempt transactions and provide them to a member state when requested."
- **Grade:** 1 plain
- **Rhea's default:** facts are kept forever, since the log is append-only. The retention rule concerns the evidence: sales documents, fair-day summaries, marketplace reports, exemption certificates and return confirmations. Each is kept at least until the later of three years after the end of the period's calendar year (31 December 2029 for 2026 periods, computed) and three years after final payment was due. Nothing is purged while any period lacks a filed return. A certificate is kept as long as the latest sale it covers.
- **Rhea today:** right in substance. Append-only log, blob store, received and issued documents, and backup and restore.
- **Where the work lives:** *pack data* (the retention policy as data; Rhea deletes nothing anyway).
- **Deadline:** as above.
- **Watch:** IC 6-8.1-5 in the DOR synopsis (SEA 243 changed 6-8.1-5-2 retroactively to 1 January 2026, on agent-report response time); SIB 56.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-8-1-5-4/ ; https://www.in.gov/dA/9cdc9ee785/sib56.pdf ; https://www.streamlinedsalestax.org/docs/default-source/agreement/ssuta/ssuta-as-amended-through-12-19-25-with-hyperlinks-and-compiler-notes-at-end.pdf ; https://www.in.gov/dor/files/legislative-synopsis-2026.pdf

---

### STX-32 — How will we hear that Indiana changed a rate or what it taxes, and when does the change bite?
- **Applies to:** everyone
- **The law:**
  - IC 1-1-3-3: a section of an act "takes effect on July 1 next following its enactment" "unless a different time is specified in the SECTION".
  - 2026 shows how often the exception is used: HEA 1406's nonprofit change took effect 1 July 2026; the cash rounding 15 March 2026; IC 5-36.5 takes effect 1 January 2027; and parts of SEA 243 were made retroactive (synopsis).
  - IC 6-2.5-11-10 (SIB 90): "if 30 days are not provided between the enactment of a statute changing the rate and the effective date of the rate change, the department shall relieve the seller of liability for failing to collect tax at the new rate". SSUTA §304 C says the same.
  - SSUTA §328 C–D: a seller relying on the taxability matrix is relieved "until the first day of the calendar month that is at least 30 days after notice of a change … is submitted to the Governing Board". Indiana's own relief, IC 6-2.5-11-10, is narrower as the DOR states it: it covers certified service providers and sellers using a certified automated system that rely on the matrix (SIB 90, SIB 91). IC 6-2.5-11-11 relieves purchasers of penalties where the seller relied on erroneous DOR data (SIB 90).
  - DOR bulletins are "not binding on either the department or the taxpayer" where they depart from the law (every bulletin's disclaimer).
- **Grade:** 1 plain
- **Rhea's default:** every Indiana rule in the pack carries `effective_from` and a `warrant` citing the act and section. A change Rhea notices becomes a draft rule version with its effective date and is never applied before that date. A taxability decision records which matrix version it relied on, because the §328 relief depends on it.
- **Rhea today:** the hook is there. `Rule.EffectiveFrom` and `Warrant` (basis, citations, scope) in internal/core. Nothing yet watches the sources.
- **Where the work lives:** *shared ground*: the law watch itself, fetchers and differs for the sources below. The Indiana rules are *pack data*.
- **Deadline:** none.
- **Watch:** the sources, concretely:
  1. **The General Assembly**, iga.in.gov. 2027 is a long (budget) session from January. Enrolled acts and fiscal notes are at iga.in.gov/pdf-documents/124/<year>/…. The site is a JavaScript application and served my fetcher no text, so a feed needs its API (not verified).
  2. **The DOR Legislative Synopsis**, after each session: in.gov/dor/files/legislative-synopsis-<year>.pdf (2025 and 2026 exist; the 2026 one was revised April 2026).
  3. **The DOR information-bulletin index**, in.gov/dor/resources/tax-library/information-bulletins/. It is an HTML table of number, title, publication date and effective date. PDFs sit at /dA/<content-hash>/sibNN.pdf, and the hash changes when a bulletin is replaced, so diffing the table catches revisions (September 2026: SIB 12, 64, 82 and 85).
  4. **The departmental-notices index**, the same kind of table (DN #3 interest every autumn).
  5. **DOR emails**: the monthly Tax Bulletin through subscription.in.gov, republished as in.gov/dor/files/<yyyy>-<month>-tax-bulletin.pdf, plus one-off notices such as penny-rounding-for-retailers.pdf. I found no public index or RSS feed of the email archive.
  6. **The Indiana Register**, iar.iga.in.gov: 45 IAC rules and DOR rulings. Not reached.
  7. **The SST Governing Board**: Agreement amendments (2025: 20 May, 8 October, 19 December), the taxability matrix (sst.streamlinedsalestax.org/TM and /TAP) with its updates in the SST library, and the rate and boundary files.
  8. **The DOR sales-tax forms page**, ST-103 instructions at a content-hash URL.
- **Sources:** https://codes.findlaw.com/in/title-1-general-provisions/in-code-sect-1-1-3-3/ ; https://www.in.gov/dA/52e9b05d52/sib90.pdf ; https://www.in.gov/dA/778f09c373/sib91.pdf ; https://www.streamlinedsalestax.org/docs/default-source/agreement/ssuta/ssuta-as-amended-through-12-19-25-with-hyperlinks-and-compiler-notes-at-end.pdf ; https://www.in.gov/dor/resources/tax-library/information-bulletins/ ; https://www.in.gov/dor/legal-resources/tax-library/departmental-notices/ ; https://www.in.gov/dor/files/legislative-synopsis-2026.pdf ; https://www.streamlinedsalestax.org/for-businesses/state-taxability-matrix

---

### STX-33 — What does year end ask of us: annual return, frequency review, reconciliation?
- **Applies to:** everyone
- **The law:**
  - IC 6-2.5-6-1(c): an annual return is due "not later than the last day of the month immediately following the close of that reporting period", that is 31 January.
  - DOR Business FAQ: "All business accounts are reviewed in November and any filing status changes are made at that time. … You will be notified of any changes to your filing status before the beginning of the next tax year". Consolidated filers' RP-11 "is due Jan. 31 and can only be filed online".
  - Next year's allowance rate turns on the fiscal year ending 30 June (IC 6-2.5-6-10(b)), and next year's interest rate on DN #3 (STX-29).
- **Grade:** 1 plain
- **Rhea's default:** at year end Rhea:
  1. loads next year's filing frequency from the November notice, and next year's interest rate and allowance rate;
  2. reconciles the year: the sum of ST-103 line 1 against sales in the books, the sales tax payable account to zero after the last payment, and sales against the federal return, with each difference explained or tasked;
  3. raises the 31 January return for an annual filer.
- **Rhea today:** month close exists; there is no year close for sales tax.
- **Where the work lives:** *pack data*
- **Deadline:** 31 January (annual returns; RP-11).
- **Watch:** INTIME correspondence (account notices, which arrive as received documents); DN #3; the DOR Business FAQ.
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-6-1/ ; https://www.in.gov/dor/i-am-a/business-corp/business-faq/ ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-6-10/


## Selling into other states — nexus, sourcing, Ohio and California

Reviewed against what Rhea's shared ground gives a US pack today, because no US pack exists. That ground is the VAT-shaped `sales_invoice` v3 and `purchase_invoice` v1 (`testdata/seed/finance_v8.json`), `vat_rate` (code, name, percent), `fx_rate` and its rate channel (`packs/pl/nbp.json`, `markets/pl/nbp.go`), the warehouse `location` and `item`, `packs/common` (documents, people, periods), the formula language with declared rounding, the clock and the host. The law is as of 16 October 2026.

Abbreviations:
- **SSUTA**: the Streamlined Sales and Use Tax Agreement as amended through 19 December 2025, the latest consolidated text on streamlinedsalestax.org. I did not check whether the Board adopted amendments in 2026.
- **SST**: Streamlined Sales Tax. **SSTRS**: its registration system. **CSP**: Certified Service Provider. **MF**: marketplace facilitator.
- **Tech Guide**: SST Technology Guide SST160.8 (2 May 2022).
- **R&TC**: California Revenue and Taxation Code. **Reg**: Title 18 California Code of Regulations. Both are quoted from CDTFA's *Business Taxes Law Guide—Revision 2026*, because leginfo.legislature.ca.gov answered every request with a Cloudflare challenge.
- **CDTFA FAQ**: CDTFA, "Use Tax Collection Requirements Based on Sales into California Due to the Wayfair Decision".
- **ORC**: Ohio Revised Code. codes.ohio.gov refused connections all day, and so did tax.ohio.gov and legislature.ohio.gov. Ohio law is therefore quoted through Department documents on dam.assets.ohio.gov:
  - **ST 2001-01**: Information Release ST 2001-01, rev. November 2019.
  - **2019 alert**: "Substantial Nexus and Marketplace Facilitator Changes", July 2019.
  - **AR 2025**: the Department's Annual Report for FY2025.
  - The Finder, and the SST Ohio pages.
- **IC**: Indiana Code. iga.in.gov returned only an empty JavaScript shell, and forms.in.gov was blocked by Cloudflare. Indiana law is quoted through Department of Revenue bulletins: **SIB 89** = Sales Tax Information Bulletin #89, April 2024; **SIB 8** = Bulletin #8, January 2023.
- **MTC**: Multistate Tax Commission. **NNP**: its National Nexus Program.
- **STX**: the Indiana home-state section of this register. This section cross-refers to it instead of repeating Indiana detail.

Entries are ordered by when a business meets them in its first year: setting up, first purchases, first sales, payments, month and quarter end, later in the year, year end.

---

### MST-01 — Can a state where we have no office or staff make us collect its sales tax?
- **Applies to:** everyone
- **The law:**
  - South Dakota v. Wayfair, Inc., No. 17-494, decided 21 June 2018, 5–4. It is reported at 585 U.S. 162; that page number comes from secondary listings — verify. 138 S. Ct. 2080.
  - The overruling: "The Court's decisions in Quill Corp. v. North Dakota, 504 U. S. 298 (1992), and National Bellas Hess, Inc. v. Department of Revenue of Ill., 386 U. S. 753 (1967), should be, and now are, overruled."
  - The test: the first prong of Complete Auto "simply asks whether the tax applies to an activity with a substantial nexus with the taxing State". Here "the nexus is clearly sufficient based on both the economic and virtual contacts". The South Dakota act applied only to sellers that "deliver more than $100,000 of goods or services into South Dakota or engage in 200 or more separate transactions … on an annual basis".
  - Left open: "whether some other principle in the Court's Commerce Clause doctrine might invalidate the Act". The Court noted features "designed to prevent discrimination against or undue burdens":
    - a safe harbor for "limited business";
    - "no obligation to remit the sales tax may be applied retroactively";
    - SSUTA membership, with "sales tax administration software paid for by the State" and immunity "from audit liability" for sellers that use it.
  - Each state's own statute sets the reach. California wrote it into R&TC 6203(c): "any retailer that has substantial nexus with this state for purposes of the commerce clause of the United States Constitution and any retailer upon whom federal law permits this state to impose a use tax collection duty". CDTFA's note to 6203 says its 2019 amendments "shall not have any retroactive effect" (Stats. 2019, ch. 5, § 8).
  - I found no enacted federal statute on remote-seller collection. Unverified that none is pending.
- **Grade:** 1 plain for the holding. Its outer limits, such as how low a threshold may go or a threshold with no small-seller safe harbor, have not been tested in any decision I found.
- **Rhea's default:** Rhea never infers nexus. Each state's threshold is a dated pack rule, with a warrant citing that state's statute and Wayfair as the constitutional basis. A state with no rule in the pack shows as "not assessed", never as "no obligation".
- **Rhea today:** `Rule.warrant` (basis, citations, scope) and `effective_from` carry exactly this. There is no notion of a US state: `company` has one `country`. Missing.
- **Where the work lives:** *pack data*
- **Deadline:** —
- **Watch:**
  - Supreme Court slip opinions, an HTML list with one page per term (OT2025: https://www.supremecourt.gov/opinions/slipopinion/25). A term opens on the first Monday of October. There is no feed.
  - Congress for a federal remote-seller standard: api.congress.gov (JSON, key required), searching bill text for "remote seller" and "sales tax".
- **Sources:**
  - https://www.supremecourt.gov/docket/docketfiles/html/public/17-494.html
  - Slip opinion: https://www.supremecourt.gov/opinions/17pdf/17-494_j4el.pdf. It returned 404 from here, so the text was read from the copy at https://www.indianz.com/News/2018/06/21/17-494opinion.pdf
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6203.html

---

### MST-02 — Rhea knows we are in Indiana: what physical presence elsewhere would make us register from the first sale, whatever our sales?
- **Applies to:** everyone
- **The law:**
  - **Indiana** (SIB 89, restating IC 6-2.5-2-1). A merchant is physically present when it "maintains an office, place of distribution, sales location, sample location, warehouse, storage place, or other place of business which is located in Indiana … either permanently or temporarily, either directly or indirectly". It is also present when it "maintains a representative, agent, salesman, canvasser, or solicitor who … sells, delivers, installs, repairs, assembles, sets up, accepts returns of, bills, invoices, or takes orders". Then "the economic nexus threshold below does not apply and the retail merchant must be registered".
  - **Ohio** (ST 2001-01, restating ORC 5741.01(I)(2)(a)–(f)). A seller has nexus when it:
    - "uses an office, distribution facility, warehouse, storage facility, or similar place of business within Ohio, whether operated by the seller or any other person, other than a common carrier";
    - "regularly uses employees, agents, representatives, solicitors, installers, repair people, salespeople, or other individuals in Ohio";
    - uses a person in Ohio to take orders, deliver, install or assemble, or to let customers pick up;
    - "makes regular deliveries of tangible personal property into this state by means other than common carrier";
    - is an affiliate of a person with nexus, or leases property to consumers in Ohio.

    The seller may rebut the presumption by showing the activities "are not significantly associated with the seller's ability to establish or maintain the seller's market in Ohio".
  - **California:**
    - R&TC 6203(c)(1) reaches a retailer "maintaining, occupying, or using, permanently or temporarily, directly or indirectly, or through a subsidiary, or agent … an office, place of distribution, sales or sample room or place, warehouse or storage place".
    - 6203(c)(2) reaches a retailer with "any representative, agent, salesperson, canvasser, independent contractor, or solicitor operating in this state … for the purpose of selling, delivering, installing, assembling, or the taking of orders".
    - Reg 1684(b)(2): "there is a presumption that a retailer is engaged in business in this state … if the retailer has any physical presence in California". The presumption is rebuttable.
  - **Goods stored by a marketplace:**
    - California, Reg 1684.5(b)(2), Example 7: a marketplace seller that stores inventory in a California warehouse but sells only through a registered MF "will not be required to be registered".
    - Ohio (SST marketplace-seller chart, Ohio row updated 25 June 2020): a seller whose sales all go through an MF that collects "does not need to register".
    - Indiana (same chart, updated 22 June 2020): no registration, "unless the marketplace seller has physical presence in Indiana".
- **Grade:** 1 plain. Ohio's statute and California's regulation write the presumptions and their rebuttal.
- **Rhea's default:** the company declares, for each state and with dates:
  - every place it keeps goods: its own, a third-party warehouse, or a marketplace's fulfilment network;
  - every person who works there for it;
  - whether it delivers in its own vehicles.

  Any one of these raises a "physical presence" fact and a task to register before the next direct sale into that state.
  - Hoosier is present in Indiana only. But its own trucks delivering regularly into Ohio would be nexus under (I)(2)(d), whatever its sales.
  - Maple Lane: a marketplace storing its goods in Ohio or California creates no registration while every sale there runs through that marketplace.
- **Rhea today:**
  - The warehouse `location` is a code and a name, with no address or state.
  - `company` has one address and one `country`.
  - A `person` has no work location.

  Missing.
- **Where the work lives:** *shared ground*: a location and a company place of business need a postal address with a state, since the Polish types assume one country. Plus *pack data* for the state rules.
- **Deadline:**
  - Ohio: the duty "commences with the month that includes the day of the contact that establishes a regular presence" (ST 2001-01 §IV).
  - California: a seller's permit "for each place of business" before selling there (R&TC 6066).
- **Watch:**
  - Ohio information releases, revised in place with the revision date in the file name: https://dam.assets.ohio.gov/image/upload/tax.ohio.gov/sales_and_use/information_releases/
  - The version list for ORC 5741.01 at codes.ohio.gov. Each version has a dated URL; the search index shows one dated 3 April 2025, whose content I could not read.
  - CDTFA's law guide, revised yearly ("Revision 2026").
  - SIB 89, whose header names the bulletin it replaces.
- **Sources:**
  - https://www.in.gov/dor/files/sib89.pdf
  - https://dam.assets.ohio.gov/image/upload/tax.ohio.gov/sales_and_use/information_releases/ST%202001-01%20SUT%20Nexus%20Rev%20November%202019.pdf
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6203.html
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1684.html
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1684-5.html
  - https://www.streamlinedsalestax.org/for-businesses/marketplace-sellers

---

### MST-03 — An employee works from home in Ohio, or a sales rep visits California: does that pull us in?
- **Applies to:** corporation. Also the sole proprietor, through its freelance contractors.
- **The law:**
  - **Ohio, ST 2001-01.** §I (5741.01(I)(2)(b)): the seller "regularly uses employees … in Ohio for the purpose of conducting its business". §II lists nexus-creating activities "whether by the out-of-state seller or others acting on the seller's behalf", among them soliciting, delivering, installing, repair, "technical assistance or consulting service … including … design service", resolving complaints and training. It adds: "It is irrelevant whether such others reside in Ohio."
  - **Ohio safe harbours, §III.** They cover recruiting events, attending others' seminars, and "no more than seven instances of nexus creating activities in Ohio in a calendar year, and … no more than $25,000 in gross sales sitused to this state". These "are not mandated by statutory or case law; rather … administrative convenience". They do not apply once the seller passes "$100,000 or 200 or more separate transactions".
  - **California.** R&TC 6203(c)(2) reaches "any representative, agent, salesperson, canvasser, independent contractor, or solicitor operating in this state … for the purpose of selling, delivering, installing, assembling, or the taking of orders". Reg 1684(b)(2) presumes nexus from "any physical presence in California", rebuttable.
- **Grade:** 2 settled. "Regularly uses" and "for the purpose of selling" do not define themselves. Ohio settles them through the activity list and the seven-instance / $25,000 harbour in ST 2001-01, California through Reg 1684's presumption on any physical presence. Whether a back-office employee working from home in California rebuts that presumption is not answered in anything I read.
- **Rhea's default:**
  - Every person, employee or contractor, gets a dated work state. A person working outside the company's home state raises a nexus fact for that state.
  - In Ohio, Rhea counts the calendar year's visits and instances against the seven-instance / $25,000 harbour.
  - In California, any presence is presumed nexus. The task asks the owner either to rebut it, with facts, or to register.
  - Rhea raises the back-office question and does not answer it; that goes to the accountant.
  - The payroll provider's journal lists each employee's work state, so a new state appearing there is the trigger.
- **Rhea today:** `person` (people pack v2) has name, handle, role, status, deputy, names and birth date, with no work location. Contractors exist only as payees. Missing.
- **Where the work lives:** *shared ground*: the common `person` type needs a dated work location. Plus *pack data* for the Ohio and California tests.
- **Deadline:** Ohio: from the month that includes the day of the contact (ST 2001-01 §IV).
- **Watch:** as MST-02. Ohio's harbours are an information release, so a change would appear as a new revision of ST 2001-01 on dam.assets.ohio.gov.
- **Sources:** ST 2001-01 (URL in MST-02); https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6203.html ; https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1684.html

---

### MST-04 — We sell at a craft fair or exhibit at a trade show in another state: what do we need?
- **Applies to:** sole proprietor (craft fairs); corporation (trade shows). For fairs inside Indiana, see STX-02.
- **The law:**
  - **Ohio licence.** AR 2025, Exhibit 1: "Transient vendor — $50 — Retailer who makes sales in any county in which they have no fixed place of business. The license is valid statewide." H.B. 366 raised the fee from $25 to $50 "effective April 9, 2025" (AR 2025, history).
  - **Ohio harbour.** ST 2001-01 §III keeps an exhibitor out of remote-sale collection "provided that the out-of-state seller has no more than seven instances of nexus creating activities in this state in a calendar year and does not have gross receipts in Ohio in excess of $25,000".
  - **Ohio sourcing at the fair.** SSUTA §310(A)(1): "When the product is received by the purchaser at a business location of the seller, the sale is sourced to that business location", so the rate is that of the fair's county.
  - **California exception.** R&TC 6203(d): a retailer is not engaged in business by trade-show activity "for not more than 15 days, in whole or in part, in this state during any 12-month period" with no more than "$100,000 of net income from those activities in this state during the prior calendar year". But it "is liable for collection of the applicable use tax, with respect to any sale … occurring at the convention and trade show activities and … pursuant to an order taken at or during those".
  - **California permits.**
    - R&TC 6066: a permit "for each place of business".
    - R&TC 6073: the operator of a "swap meet, flea market, or special event" may be required to obtain evidence that each seller holds a valid seller's permit.
    - CDTFA FAQ: "Seasonal sellers may apply for a temporary permit using our online registration."
- **Grade:** 1 plain
- **Rhea's default:**
  - An event is a dated place of business with an address.
  - Before an event in another state, Rhea raises a task to obtain that state's licence: Ohio's transient vendor's licence, or California's temporary seller's permit.
  - Sales at the event are taxed at the event location's combined rate.
  - Each event counts toward Ohio's seven-instance / $25,000 harbour or California's 15-day / $100,000 exception, which govern remote sales into that state.
- **Rhea today:** there is no event or place-of-business object, and a sale has no place of sale. Missing.
- **Where the work lives:** *pack data*
- **Deadline:** the licence before the first sale at the event.
- **Watch:**
  - Ohio's licence fee is set by statute. Each year's changes are summarised in the "History of Major Changes" of the Department's annual report, published once a year: https://dam.assets.ohio.gov/image/upload/tax.ohio.gov/communications/publications/annual_reports/
  - California: the CDTFA law guide, revised yearly.
- **Sources:**
  - https://dam.assets.ohio.gov/image/upload/tax.ohio.gov/communications/publications/annual_reports/2025annualreport.pdf
  - ST 2001-01 (MST-02)
  - https://www.streamlinedsalestax.org/docs/default-source/agreement/ssuta/ssuta-as-amended-through-12-19-25-with-hyperlinks-and-compiler-notes-at-end.pdf
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6203.html ; https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6066.html ; https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6073.html
  - https://www.cdtfa.ca.gov/industry/wayfair/frequently-asked-questions.htm

---

### MST-05 — Indiana, for completeness: what is its threshold for sellers outside the state?
- **Applies to:** everyone. It is moot for both companies, which are present in Indiana. It matters for reading the pattern and for any affiliate outside Indiana.
- **The law:**
  - SIB 89 (effective 1 January 2024, retroactive), restating IC 6-2.5-2-1. A remote seller must register if it "has gross revenue from any combination of: the sale of tangible personal property that is delivered into Indiana; a product transferred electronically into Indiana; or a service delivered in Indiana; that exceeds $100,000 in the current or preceding calendar year."
  - "Effective January 1, 2024, Indiana only has the $100,000 threshold." Senate Enrolled Act 228 (2024) removed the 200-transaction test.
  - "The threshold applies regardless of whether any of the $100,000 of revenue was generated by taxable sales. Wholesalers are considered to be retail merchants."
  - Marketplace: "the transactions that a seller makes on a marketplace are not counted toward the seller … (except in instances where the marketplace facilitator has not met the threshold themselves)".
- **Grade:** 1 plain
- **Rhea's default:** both businesses register in Indiana before their first sale because they are present there (STX-01). The $100,000 rule ships as data, for affiliates and as the pattern's reference point.
- **Rhea today:** nothing.
- **Where the work lives:** *pack data*
- **Deadline:** —
- **Watch:**
  - Indiana General Assembly (iga.in.gov). It is a JavaScript application, and I found no feed.
  - DOR bulletins (https://www.in.gov/dor/legal-resources/tax-library/information-bulletins/). Each new bulletin's header names the one it replaces.
  - DOR GovDelivery bulletins (content.govdelivery.com/accounts/INDOR).
- **Sources:** https://www.in.gov/dor/files/sib89.pdf ; https://www.in.gov/dor/i-am-a/business-corp/remote-sellers/

---

### MST-06 — Ohio: at what point must we register, and does the 200-transaction test still apply?
- **Applies to:** everyone
- **The law:**
  - 2019 alert: "If the seller's gross receipts are greater than $100,000 or the seller has at least 200 transactions in either year, the seller has substantial nexus and is required to collect from consumers, file returns, and remit the appropriate tax."
  - ST 2001-01 §I (ORC 5741.01(I)(2)(g)–(h)): "gross receipts in excess of $100,000 or 200 or more separate transactions in the current or preceding calendar year from the sale of tangible personal property for storage, use or consumption in Ohio or providing of services the benefit of which is realized in Ohio". In force from 1 August 2019 (Am. Sub. H.B. 166).
  - The SST Ohio state page, live on 16 October 2026: "Gross receipts greater than $100,000 or 200 or more transactions in the current or preceding calendar year."
  - I could not read the current ORC 5741.01. I found no Department statement removing the transaction test.
  - Indiana (2024) and Alaska's ARSSTC (2025, MST-28) dropped their transaction tests, and the economic test in Washington's B&O nexus statute counts dollars, not transactions (RCW 82.04.067(1)(c)(i), MST-29). Ohio is the one state of the three that still counts transactions.
- **Grade:** 1 plain on the Department's own text. The current statute text is unread.
- **Rhea's default:**
  - Rhea counts Ohio receipts and Ohio transactions separately, for the calendar year to date and for the previous year. Crossing either test starts registration.
  - Maple Lane must watch the count, not the dollars. 200 web-shop orders of $40 are $8,000 (computed here), so a small shop can cross Ohio on transactions years before $100,000. Marketplace orders count too (MST-08).
- **Rhea today:** formula folds can sum and count objects in the log, but an invoice carries no ship-to state. Missing.
- **Where the work lives:** *pack data*. The ship-to state is the *shared ground* gap in MST-12.
- **Deadline:** see MST-09.
- **Watch:**
  - ORC 5741.01 at codes.ohio.gov. Each version has its own dated URL and an authenticated PDF (pattern seen: /assets/laws/revised-code/authenticated/57/5741/5741.01/4-3-2025/5741.01-4-3-2025.pdf).
  - Ohio's biennial budget bill. The annual report's history shows these changes arriving in the odd-year budgets: H.B. 166 (2019), H.B. 110 (2021), H.B. 33 (2023), H.B. 96 (2025).
  - The SST Ohio page.
- **Sources:**
  - https://dam.assets.ohio.gov/image/upload/tax.ohio.gov/ohiotaxalert/archivedalerts/substantialnexusandmarketplacefacilitatorchanges07232019.pdf
  - ST 2001-01 (MST-02)
  - https://streamlinedsalestax.org/state-tables/ohio

---

### MST-07 — California: at what point must we register?
- **Applies to:** everyone. Hoosier crosses in its second year.
- **The law:**
  - R&TC 6203(c)(4)(A): "Any retailer that, in the preceding calendar year or the current calendar year, has total combined sales of tangible personal property for delivery in this state by the retailer and all persons related to the retailer that exceed five hundred thousand dollars ($500,000)."
  - Related persons are those related "pursuant to Section 267(b) of the Internal Revenue Code" (6203(c)(4)(B)). The test has been operative since 1 April 2019 (6203(f)(1)) and has no transaction count.
  - Reg 1684.5(b)(2): a marketplace seller "shall include all sales of tangible personal property for delivery in this state, including sales made on its own behalf and sales facilitated through any marketplace facilitator's marketplace, regardless of whether the sales are taxable".
  - CDTFA FAQ: the threshold is met "based on your total sales of tangible personal property, which may include nontaxable sales, such as sales for resale".
  - A wholesaler that makes no retail sales into California need not register. But once over $500,000 with "any retail sale … you are then required to register … and begin filing sales and use tax returns to report your total sales to customers in California".
  - Districts: R&TC 7262(a)(1) applies the same $500,000 test, measured "in this state or for delivery in the state", to make the retailer "engaged in business in the district". The FAQ adds: "you are also considered engaged in business in all districts that impose a district tax".
- **Grade:** 2 settled. 6203 counts "sales of tangible personal property". Reg 1684.5 (marketplace sales, taxable or not) and the CDTFA FAQ (sales for resale) settle what the words cover.
- **Rhea's default:**
  - One running total for each calendar year: every sale of goods delivered to California, direct and marketplace, taxable, exempt and for resale, plus the sales of companies related under IRC 267(b). Services are left out.
  - Hoosier's resale-certificate sales count, so it crosses earlier than its taxable sales would suggest.
  - Once over, every district Hoosier ships into is in scope (MST-14).
- **Rhea today:** `company` has no ownership links, so related persons cannot be found; the only trace is the free-text `intragroup` on the invoice. Missing.
- **Where the work lives:** *pack data*
- **Deadline:** see MST-09.
- **Watch:**
  - CDTFA's law guide, revised yearly (https://www.cdtfa.ca.gov/lawguides/).
  - The CDTFA FAQ page.
  - The California Legislature on leginfo.legislature.ca.gov, which is behind a Cloudflare challenge. AB 147 itself was in effect on 25 April 2019 but operative from 1 April 2019.
- **Sources:**
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6203.html
  - https://www.cdtfa.ca.gov/lawguides/vol1/tutl/7262.html
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1684-5.html
  - https://www.cdtfa.ca.gov/industry/wayfair/frequently-asked-questions.htm

---

### MST-08 — What counts toward each threshold: gross or taxable sales, resale, exempt sales, services, marketplace sales?
- **Applies to:** everyone
- **The law:** each state's own measure (MST-05 to MST-07), side by side:

  | | Indiana (SIB 89) | Ohio (Department releases; SST Ohio page) | California (R&TC 6203; Reg 1684.5; CDTFA FAQ) |
  |---|---|---|---|
  | Measure | "gross revenue" | "gross receipts", but only "retail sales" count | "total combined sales of tangible personal property" |
  | Sales for resale | count: "Wholesalers are considered to be retail merchants" | excluded: "'retail sales' do not include 'sales for resale'" | count: "may include nontaxable sales, such as sales for resale" |
  | Exempt retail sales | count: "regardless of whether any of the $100,000 … was generated by taxable sales" | not stated in what I read | count: "regardless of whether the sales are taxable" |
  | Services | count: "a service delivered in Indiana" | only the services enumerated in R.C. 5739.01(B)(3) | not counted |
  | Electronic products | count: "a product transferred electronically" | not stated | not counted, as they are not tangible |
  | Marketplace sales | excluded for the seller | included: "If all sales into Ohio combined – including sales made though any marketplace facilitator – exceed $100,000 or 200 transactions, you must register" (2019 alert) | included |
  | Transaction count | none since 1 January 2024 | 200 | none |
  | Related companies | not stated | affiliates through the physical-presence presumption (I)(2)(e) | included (IRC 267(b)) |
  | Period | current or preceding calendar year | the same | the same |

- **Grade:** 2 settled. The Department pages and Reg 1684.5 settle what the statutes' words cover. One cell stays unverified: whether Ohio counts exempt retail sales, such as a sale to a charity on a certificate.
- **Rhea's default:**
  - Each sale carries what the three tests need: ship-to state, goods, service or electronic product, marketplace or direct, the resale or exemption basis, and the selling company. Each state's test is its own fold over those facts.
  - Ohio exempt retail sales are counted, because registering early is the safer mistake. That reading goes to the accountant.
- **Rhea today:** `sales_invoice` has none of these fields. Its `buyer` must be a `company`, and it has one `vat_rate` for the whole document. Missing.
- **Where the work lives:** *shared ground*: the sales document must carry a ship-to address, the kind of each line and the sales channel. The VAT invoice assumes one rate per document and a business buyer. Plus *pack data* for the three tests.
- **Deadline:** —
- **Watch:** SST's state guidance charts, one row per state, each dated "(updated m/d/yyyy)": https://www.streamlinedsalestax.org/for-businesses/remote-seller-faqs/remote-seller-state-guidance and https://www.streamlinedsalestax.org/for-businesses/marketplace-sellers
- **Sources:** as MST-05 to MST-07; https://www.streamlinedsalestax.org/for-businesses/remote-seller-faqs/remote-seller-state-guidance

---

### MST-09 — We have just crossed a threshold: from which sale do we collect, and by when must we register?
- **Applies to:** everyone. Hoosier crosses Ohio in year one and California in year two.
- **The law:**
  - **Ohio.** 2019 alert: "The duty to obtain a seller's use tax license, collect tax on taxable sales made to consumers in this state, and file returns and remit the appropriate tax begins with the month that the seller establishes a substantial nexus and applies prospectively from that date." ST 2001-01 §IV gives an example of a seller passing the safe harbour on 8 September 2019. It must file "beginning with the September return due on October 23, 2019, and would include all sales made for the entire month of September 2019".
  - **Ohio, marketplace facilitators** (2019 alert): collection starts "on the first day of the first month that begins at least thirty (30) days after meeting the thresholds".
  - **California.** Reg 1684(b)(4)(B): "required to register with the Department and begin collecting use tax immediately after the sale was made that exceeded the five hundred thousand dollar ($500,000) threshold". The CDTFA FAQ example registers on 7 July and says the seller is "not liable for collection of the use tax on sales … prior to July 7, 2019, including the $400 transaction that put you over the $500,000 threshold".
  - **Indiana.** SIB 89 names no day. The question is moot for both companies, which are present there.
- **Grade:** 2 settled. The statutes name no day. Ohio's release settles it as "the month", and its example reaches the whole month, including the days before the crossing. California's regulation settles it as "immediately after".
- **Rhea's default:**
  - Rhea warns at a margin the company sets below each test and drafts the registration task then, so the permit exists before the crossing sale.
  - Ohio: if the crossing comes first, the month's earlier untaxed Ohio sales are shown as tax the seller pays itself, following the release's example. They are never billed to customers afterwards.
  - California: collection starts with the sale after the crossing sale.
- **Rhea today:** the clock fires dated duties and the host's morning briefing can carry the warning. There is no threshold rule. Missing.
- **Where the work lives:** *pack data*
- **Deadline:** California: registration on the day of the crossing sale. Ohio: the first return, for the month of crossing, by the 23rd of the next month.
- **Watch:** Ohio information releases (MST-02); CDTFA's law guide for Reg 1684.
- **Sources:** 2019 alert and ST 2001-01 (MST-02, MST-06); https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1684.html ; https://www.cdtfa.ca.gov/industry/wayfair/frequently-asked-questions.htm

---

### MST-10 — Should we register in several states at once through the SSTRS, and can we get the tax software free?
- **Applies to:** everyone. For Indiana's own registration, see STX-05.
- **The law:**
  - **Registration.**
    - SSUTA §401(B): "By registering, the seller agrees to collect and remit sales and use taxes for all taxable sales into the states in which the seller elected to register."
    - §401(D): a state "shall not use registration with the central registration system … as the sole factor in determining whether the seller has nexus with that state for any tax".
    - The SSTRS page: "There is no fee to register." "Returns must be filed even if you have no sales in that state." It lists 23 full members, Indiana and Ohio among them, and Tennessee as associate. California is not a member.
    - Registrations: 36,132 active, of which "28.6% … contract with a CSP" (as of 30 September 2026).
  - **Free CSP service.** A "CSP-compensated seller" must register through the SSTRS, contract with a CSP, and meet five criteria in the 12 months before registering:
    - "no fixed place of business for more than 30 days in that state";
    - less than $50,000 of property and less than $50,000 of payroll there;
    - less than 25% of its total property or payroll there;
    - it was not collecting as a condition of supplying the state.

    "A remote seller that is only required to report and remit sales tax in a state because it meets the economic nexus threshold of that state qualifies for free CSP services in that state."
  - **When the CSP may charge.** For states where the seller is not CSP-compensated, and "to process non-taxable transactions if the non-taxable transactions exceed 30% of all transactions processed annually".
  - **Who offers it.** The CSPs certified and offering free services: Avalara, TaxCloud, Sovos, AccurateTax, Avior. Exactor is certified but "not currently offering free services". The current contract is the 2025–2026 one.
- **Grade:** 1 plain
- **Rhea's default:**
  - Rhea registers state by state, only where a test is met, and never in "all SST states", because every registration creates returns even at zero.
  - Hoosier would qualify for free CSP service in Ohio, where its nexus is economic only. It would not in Indiana, where it has a fixed place of business. Its resale-heavy Ohio sales may pass the 30% non-taxable line, and then the CSP may charge.
  - Rhea computes the tax itself. A CSP, if chosen, is a dated fact about who files, not a dependency.
- **Rhea today:** there is no registration object (state, account number, date, frequency, filer). Missing.
- **Where the work lives:** *pack data* for registrations. *later* for any CSP hand-off.
- **Deadline:** —
- **Watch:**
  - SST free-services page, with the CSP list and the contract period (renewed per contract term): https://www.streamlinedsalestax.org/for-businesses/free-sales-tax-services
  - SSTRS page, with the member list and the quarterly registration count.
- **Sources:**
  - https://www.streamlinedsalestax.org/for-businesses/sales-tax-registration-sstrs
  - https://www.streamlinedsalestax.org/for-businesses/free-sales-tax-services
  - SSUTA (MST-04)

---

### MST-11 — We buy goods or equipment from sellers in other states who charge us no tax: what do we owe, and to whom?
- **Applies to:** everyone. For the Indiana return, see STX-07 and STX-08.
- **The law:**
  - **Indiana.** Use tax at 7%, the state rate (STX-07). The DOR's ST-115 consumer's use tax form computes "7% of Line 1" and credits "Sales / Use tax paid to another state on these purchases". Its instructions send registered retail merchants to "ST-103, Line 7". I read these lines only through search-index copies (secondary — verify), because forms.in.gov was blocked.
  - **Ohio,** for goods used in Ohio (AR 2025): consumers "have a liability on purchases made out-of-state, by catalog or via the Internet on which no sales tax has been paid. Taxpayers with an annual consumer's use tax liability exceeding $1,000 must register for a consumer's use tax account and file returns."
  - **California,** for goods used in California. CDTFA-401 instructions (rev. 34, 1-26), line 2, "Purchases Subject to Use Tax", covers property "purchased … from an out-of-state retailer who did not collect California use tax". Businesses with no permit become "qualified purchasers" above "ten thousand dollars ($10,000) in purchases subject to use tax per calendar year", with an annual return due "on or before April 15" (R&TC 6225). Enacted change: 6225 "shall remain in effect only until January 1, 2029, and as of that date is repealed".
- **Grade:** 1 plain
- **Rhea's default:**
  - Every purchase invoice carries the tax the supplier charged and its state.
  - A taxable purchase used in a state, with no tax of that state charged, accrues use tax at that state's rate in the period of purchase: a liability, with the cost going to the expense or asset account.
  - Purchases bought for resale are not use-taxed (resale certificates to suppliers, MST-21).
- **Rhea today:** `purchase_invoice` has `vat`. The Polish rule debits it to recoverable input VAT (account 221). In the US, tax paid on a purchase is never recoverable: it is part of cost, and use tax is a self-assessed liability. Wrong for a US pack if reused as it is.
- **Where the work lives:** *shared ground*: the purchase document's tax must be able to be cost, and to be self-assessed, not only to be an input-tax asset. Plus *pack data*.
- **Deadline:** with the period's return (STX-07). California qualified purchasers: 15 April for the previous calendar year.
- **Watch:** Indiana forms catalogue (forms.in.gov, versions dated per period; ST-115 has a version for periods after 1 January 2025); the CDTFA-401 instructions, whose revision number and date are in the header; R&TC 6225's sunset on 1 January 2029.
- **Sources:**
  - https://forms.in.gov/Download.aspx?id=7177 (blocked; read via search index — secondary)
  - AR 2025 (MST-04)
  - https://www.cdtfa.ca.gov/formspubs/cdtfa401inst.pdf
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6225.html

---

### MST-12 — Which address decides the tax on a sale: ours or the buyer's?
- **Applies to:** everyone. For Indiana sales, see STX-12.
- **The law:**
  - **SSUTA §310(A).**
    - (1) "When the product is received by the purchaser at a business location of the seller, the sale is sourced to that business location."
    - (2) Otherwise, "to the location where receipt by the purchaser (or the purchaser's donee …) occurs, including the location indicated by instructions for delivery".
    - (3)–(5) are fallbacks: an address in the seller's records, then the payment-instrument address, then the place from which the goods were shipped.
    - §310.1 lets a member elect origin sourcing only when the order is received in the same state as the delivery.
  - **Ohio** (AR 2025): "For taxable sales made by Ohio vendors and delivered to an Ohio consumer, the sales tax rate is based on the location where the vendor receives the order. Sales made by out-of-state sellers are generally sourced to the location where the consumer receives the tangible personal property." Marketplace facilitators source to the consumer under R.C. 5741.05.
  - **California, retailers inside the state.** R&TC 7263: "all retail transactions are consummated at the place of business of the retailer unless the tangible personal property sold is delivered by the retailer or his agent to an out-of-state destination". A district use tax applies when a retailer engaged in business in a district "ships or delivers the property into the district" (7262(a)(2)).
  - **California, retailers outside it.** Use tax is due at the delivery address. Once over $500,000 the retailer is "engaged in business in all districts" (CDTFA FAQ), and it collects each district's use tax where it delivers.
  - **Indiana:** destination, as an SST member, with one statewide rate.
- **Grade:** 1 plain
- **Rhea's default:**
  - Every sale line carries a ship-to address, or a pickup place.
  - Pickup at the seller's own location (a fair, a trade counter) is sourced there. A delivery is sourced to the ship-to.
  - Both companies are out-of-state sellers in Ohio and California, so they are destination-sourced everywhere outside Indiana.
- **Rhea today:** `sales_invoice` has no address of any kind, and `buyer` is a required `ref<company>`. A consumer sale, or a ship-to that differs from the bill-to, cannot be written. The Polish design places the supply by the buyer's tax number and country. Missing.
- **Where the work lives:** *shared ground*: a sale needs a ship-to and pickup address on each line, and a buyer that may be a private person. This is the same gap as Poland's VAT-06.
- **Deadline:** —
- **Watch:** SSUTA amendments, adopted at Governing Board meetings. The compiled text lists their dates; in 2025 they were 20 May, 8 October and 19 December. The new compiled PDF is posted under /docs/default-source/agreement/ssuta/.
- **Sources:**
  - SSUTA (MST-04)
  - AR 2025 (MST-04)
  - https://www.cdtfa.ca.gov/lawguides/vol1/tutl/7263.html ; https://www.cdtfa.ca.gov/lawguides/vol1/tutl/7262.html
  - CDTFA FAQ (MST-07)

---

### MST-13 — Ohio: which rate applies at a given address, and where do we get it?
- **Applies to:** everyone registered in Ohio
- **The law:**
  - **Rates** (AR 2025):
    - state rate "5.75 percent since Sept. 1, 2013";
    - all 88 counties "levied permissive sales and use taxes as of August 5, 2025, ranging from 0.50 percent to 2.00 percent";
    - "eleven transit authorities … ranging from 0.10 percent to 1.00 percent".
  - **How high local rates may go.** Counties: up to 1.00% plus 0.50% (R.C. 5739.026), plus, from 1 October 2019, 0.50% for a detention facility. Transit authorities: "0.05 percent to 1.50 percent, also in 0.05 percent increments", capped against the county rate.
  - **Transit districts split counties.** The Department's bulletin of 10 December 2024 announced the COTA increase for 1 April 2025. In Delaware County the "COTA (96)" area went from 7.50% to 8%, while "non-COTA (21)" stayed at 7%: one county, two rates, two reporting codes.
  - **The Finder,** Ohio's lookup. Web search, plus a REST service at /finder/api/v1, with POST tax-rates/address, /zipcode and /coordinates. It needs an "x-api-key", and "All three operations require a date field formatted as MM/DD/YYYY". On reliance: "in conformity with the Streamlined Sales Tax Agreement … vendors and sellers may rely on this information … based on the date used for the search".
  - **SST rate and boundary files** (Tech Guide ch. 5).
    - Format: comma-delimited CSV.
    - Rate file fields: state, jurisdiction type, jurisdiction code, intrastate and interstate general rates and food/drug rates ("Always 0.nnnnn"), begin and end dates (CCYYMMDD).
    - Boundary records: "A" address, "4" ZIP+4, "Z" 5-digit.
    - Names: e.g. OHR2026Q4SEP17, meaning state, R or B, effective year and quarter, upload date.
    - Posting: "by the first day of the month prior to the beginning of each calendar quarter", and only when something changed.
    - "Calculate tax for each FIPS Code (taxing jurisdiction) separately."
  - **SSUTA lookup rules.** §305(A)(6)(a): where a ZIP+4 area holds more than one rate, "the state must assign the lowest combined tax rate". §309(A) sets the lookup order: address, then ZIP+4, then 5-digit. §306(A) relieves sellers who rely on the state's data.
- **Grade:** 1 plain
- **Rhea's default:**
  - A channel in `markets/us` ingests Ohio's SST rate and boundary files each quarter as dated events, the same mechanism as Poland's NBP rates.
  - A sale resolves standardized address → county and transit codes → rates for the sale date.
  - The Finder is a cross-check that needs a key and works one address at a call. It is not the record.
- **Rhea today:** the channel mechanism and dated `fx_rate` exist. `vat_rate` is one percent with no jurisdiction and no validity dates, and there is no boundary lookup. Missing.
- **Where the work lives:** *market code* (an SST rate and boundary CSV reader as a channel; a Finder client). Plus *pack data* for Ohio's jurisdiction codes.
- **Deadline:** new files by the 1st of March, June, September and December, for quarters starting the following month.
- **Watch:**
  - SST "Rate and Boundary Files – Updates", a list by date: https://www.streamlinedsalestax.org/for-businesses/rate-and-boundary-files/rate-and-boundary-files---updates
  - The SST "Rate and Boundary File Update Notices" mailing list, which emails when files post.
  - The Department's GovDelivery rate-change bulletins (content.govdelivery.com/accounts/OHTAX).
  - The Finder: https://thefinder.tax.ohio.gov/
- **Sources:**
  - AR 2025 (MST-04)
  - https://content.govdelivery.com/accounts/OHTAX/bulletins/3c678f3
  - https://thefinder.tax.ohio.gov/webservice
  - https://www.streamlinedsalestax.org/for-businesses/rate-and-boundary-files
  - https://www.streamlinedsalestax.org/docs/default-source/technology/technology-guide-may-2022.pdf
  - SSUTA (MST-04)

---

### MST-14 — California: which rate applies at a given address, and where do we get it?
- **Applies to:** everyone registered in California (Hoosier from year two)
- **The law:**
  - **The rate.** CDTFA rates page: "The statewide tax rate is 7.25%. In most areas of California, local jurisdictions have added district taxes … Those district tax rates range from 0.10% to 2.00%. Some areas may have more than one district tax in effect." The return splits the 7.25%: line 13 "State Tax 6.00 Percent", line 14 county 0.25%, line 15 "Local Tax 1.00 Percent", line 16 district tax through CDTFA-531-A2 (CDTFA-401 instructions rev. 34, 1-26).
  - **Downloads.** "Tax Rates Effective October 1, 2026" as xlsx (SalesTaxRates10-1-26.xlsx), "all have 3 digits after the decimal". "Some communities located within a county or a city may not be listed."
  - **The API.** GET https://services.maps.cdtfa.ca.gov/api/taxrate/GetRateByAddress?address=&city=&zip= and GetRateByLngLat. On 16 October 2026, 200 N Spring St, Los Angeles 90012 returned `{"rate":0.1025, … "tac":"19050D290000"}`. There is no date parameter: the API answers for today only.
  - **Changes this quarter** (Explanation of Tax Rate Changes, operative 1 October 2026):
    - County of Los Angeles 9.750% → 10.250%, new code D29, under "Measure ER, approved by voters in the June 2, 2026, election";
    - City of Perris 7.750% → 8.750%;
    - City of Red Bluff 7.500% → 8.500%.
- **Grade:** 1 plain
- **Rhea's default:**
  - A `markets/us` channel ingests the quarterly xlsx as dated rate events. At sale time Rhea asks the API for the address and records the TAC, the rate and the response in the sale's fact. That recorded answer is the evidence, because the API cannot answer for a past date, and replay never asks it again.
  - The rate is read from the JSON as decimal text into an exact rational. `0.1025` must never pass through a float (invariant 6).
- **Rhea today:** the channel and the market stamp exist. There is no jurisdiction or rate model. Missing.
- **Where the work lives:** *market code*
- **Deadline:** changes take effect on 1 January, 1 April, 1 July and 1 October (MST-24).
- **Watch:**
  - CDTFA rates page, published quarterly; the file name carries the effective date: https://www.cdtfa.ca.gov/taxes-and-fees/sales-use-tax-rates.htm
  - "Explanation of Tax Rate Changes", one section per quarter: https://www.cdtfa.ca.gov/taxes-and-fees/explanationoftaxratechanges.htm
  - The API. Comparing the TAC returned against the xlsx code list catches a change mid-quarter.
- **Sources:**
  - https://www.cdtfa.ca.gov/taxes-and-fees/sales-use-tax-rates.htm
  - https://www.cdtfa.ca.gov/taxes-and-fees/explanationoftaxratechanges.htm
  - https://services.maps.cdtfa.ca.gov/api/taxrate/GetRateByAddress?address=200+N+Spring+St&city=Los+Angeles&zip=90012
  - https://www.cdtfa.ca.gov/formspubs/cdtfa401inst.pdf

---

### MST-15 — What must each sale record so the tax can be proved years later?
- **Applies to:** everyone. For the face of the invoice, see STX-15.
- **The law:**
  - **Due diligence** (SSUTA §309(A)(2)). The seller has exercised it if it obtained "the complete street address and five-digit zip code" and "standardized the address using either: 1. the look-up application available from the United States Postal Service; 2. software certified by the Coding Accuracy Support System (CASS™); or 3. other software approved by the Governing Board".
  - **Without it.** §306(C): if the seller "does not request the complete street address", a state "is not required to relieve the seller".
  - **Exempt sales.** §317(A)(6): "The seller shall maintain proper records of exempt transactions and provide them to a member state when requested."
  - **Dates and jurisdictions** (Tech Guide). "The rows in the boundary and rate databases are date dependent." "Calculate tax for each FIPS Code (taxing jurisdiction) separately."
  - **Retention.**
    - California, Reg 1698(i): records "preserved for a period of not less than four years".
    - Ohio and Indiana: not read (statutes unreachable); see STX for Indiana.
- **Grade:** 1 plain
- **Rhea's default:** each taxable sale line stores:
  - the ship-to address as entered and as standardized, with the standardizer's name and version;
  - the jurisdiction codes found;
  - the rate file, or the API response, used;
  - the product's taxability category;
  - the exemption certificate, if any;
  - whether the sale was marketplace-facilitated;
  - the tax per jurisdiction before rounding.

  The replay invariant needs exactly this: the rate source is itself an event in the log.
- **Rhea today:** every derived fact already carries its provenance (event, rule, rule version), and `received_document` keeps files by hash. There are no sale lines, no address and no per-jurisdiction amounts. Missing.
- **Where the work lives:** *shared ground*: sale lines, and tax components per jurisdiction. Poland's one rate per invoice is in the way.
- **Deadline:** keep records at least 4 years (California).
- **Watch:** SST Technology Guide, whose current version (SST160.8) is linked from the rate-files page; CDTFA's law guide for Reg 1698.
- **Sources:** SSUTA (MST-04); Tech Guide (MST-13); https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1698.html

---

### MST-16 — Are our goods taxable in each state: clothing, home goods, food?
- **Applies to:** everyone. For Indiana food and bundles, see STX-21.
- **The law:** I read the SST taxability matrix (Library of Definitions) through its public JSON API: Indiana version 2026.0, published 4 August 2026, and Ohio version 2026.0, published 31 July 2026.
  - **Clothing (20010).** Taxable in Indiana (IC 6-2.5-4-1) and in Ohio (R.C. 5739.01(B)(1)). Children's diapers are exempt in Indiana "on or after 9/1/2022" (IC 6-2.5-5-57) and in Ohio "Beginning October 1, 2023" (R.C. 5739.02(B)(60)).
  - **Food (40030) in Indiana.** Food is exempt (IC 6-2.5-5-20). Candy (40010) and soft drinks (40050) are taxable, "7% sales tax". Prepared food (41000) is taxable and "may also be subject to Food and Beverage taxes".
  - **Food in Ohio.** "Food for consumption off the premises where sold is exempt" (Ohio Const. art. XII § 3; R.C. 5739.02(B)(2)). Soft drinks fall outside "food" (R.C. 5739.01(CCC)(2)(c)). Ohio's matrix marks food, candy and prepared food with the same flag it gives taxable clothing and puts the exemption in the comment column. The matrix cannot be read as true or false alone.
  - **Holidays.** Ohio holds one on the first Friday to Sunday of August: clothing up to $75, school supplies up to $20, or every item up to $500 in a surplus year (H.B. 33; SST Ohio page). Indiana's matrix answers the holiday question with 0, meaning none.
  - **California** is not an SST member and has no matrix. The CDTFA-401 instructions, line 5, deduct "Nontaxable Sales of Food Products" and say "alcoholic and carbonated beverages" and "hot prepared food products" are generally taxable. I found no clothing exemption (not verified further).
  - **What the two companies sell.** Maple Lane's home goods and Hoosier's tools and supplies are general tangible property, taxable in all three states unless the buyer or the use is exempt.
- **Grade:** 2 settled, by the states' matrix entries, their cited statutes and California's return instructions.
- **Rhea's default:**
  - An item carries a taxability category: the SST Library of Definitions code (20010, 40030 …), or a pack code where SST has none.
  - The pack maps each code to each state's treatment, with effective dates. California gets its own map.
  - A person reads each matrix row's comment once, and the pack records the treatment as a rule whose warrant is the row's cited statute.
- **Rhea today:** the warehouse `item` has sku, name and unit, and no tax category. Missing.
- **Where the work lives:** *pack data*
- **Deadline:** —
- **Watch:**
  - The SST taxability matrix API, plain JSON with no key:
    - https://sst.streamlinedsalestax.org/api/states
    - https://sst.streamlinedsalestax.org/api/forms (each version with its `publishDate`)
    - https://sst.streamlinedsalestax.org/api/forms/{formId}/rows (Indiana 2026.0 is form 14921, Ohio 2026.0 is form 14973)
  - Annual versions appear at the end of July, with updates in between (Ohio 2025.1 on 5 May 2026, 2025.2 on 31 July 2026).
  - SST's relief runs "until the first day of the calendar month that is at least 30 days after notice of a change". Comparing `publishDate` values gives that notice.
- **Sources:**
  - https://www.streamlinedsalestax.org/for-businesses/state-taxability-matrix
  - https://sst.streamlinedsalestax.org/api/forms/14921/rows ; https://sst.streamlinedsalestax.org/api/forms/14973/rows
  - https://streamlinedsalestax.org/state-tables/ohio
  - https://www.cdtfa.ca.gov/formspubs/cdtfa401inst.pdf

---

### MST-17 — Software, SaaS and digital downloads: taxable where?
- **Applies to:** everyone. Maple Lane may sell downloadable patterns; both companies buy software. For Indiana, see STX-09.
- **The law:**
  - **Indiana** (matrix and SIB 8).
    - Prewritten software is taxable however delivered: tangible (30040), "delivered electronically" (30050) and "load and leave" (30060). Custom software is exempt (30015).
    - SaaS: "Prewritten computer software … that is remotely accessed over the internet … is not considered an electronic transfer of computer software and is not considered a retail transaction" (SIB 8; IC 6-2.5-4-16.7).
    - Digital audio-visual works, audio works and books sold "to an end user with rights for permanent use" are taxable. Sold "with rights of use less than permanent" they are exempt (IC 6-2.5-4-16.4).
  - **Ohio.**
    - Prewritten software is taxable, including electronic delivery. "Custom system software for business use is a taxable computer service" (matrix 30015).
    - The tax base includes "automatic data processing, computer services and electronic information services used in business" and "a specified digital product … for permanent use or less than permanent use, regardless of whether continued payment is required" (AR 2025).
    - Enacted, in force from 1 January 2026: H.B. 96 repealed exemptions tied to electronic publishing and the 25% refund to electronic information service providers (AR 2025).
  - **California.** Reg 1502(f)(1)(D): "The sale or lease of a prewritten program is not a taxable transaction if the program is transferred by remote telecommunications from the seller's place of business, to or through the purchaser's computer and the purchaser does not obtain possession of any tangible personal property". Tax applies to tangible storage media.
- **Grade:** 2 settled, by Indiana's bulletin, both matrices, and California's Reg 1502. Whether Ohio taxes a consumer's SaaS turns on "used in business", which the annual report states but which I did not read in the statute.
- **Rhea's default:**
  - An electronic product carries its delivery (tangible, download, remote access), its rights (permanent, less than permanent, subscription) and, for Ohio, the buyer's use (business or consumer).
  - Indiana taxes permanent downloads, not SaaS. Ohio taxes downloads, and taxes SaaS for business buyers. California taxes neither unless a tangible medium passes.
  - Purchases follow the same table for use tax (MST-11).
- **Rhea today:** nothing. Missing.
- **Where the work lives:** *pack data*
- **Deadline:** —
- **Watch:** the SST matrix API (MST-16); SIB 8, whose header names what it replaces; CDTFA's law guide for Reg 1502.
- **Sources:**
  - https://www.in.gov/dor/files/sib08.pdf
  - matrix rows (MST-16)
  - AR 2025 (MST-04)
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1502.html

---

### MST-18 — Is shipping and handling taxable outside Indiana?
- **Applies to:** everyone. For Indiana, see STX-13.
- **The law:**
  - **Ohio** (matrix). Handling (11000), "Transportation, shipping, postage, and similar charges" (11010) and postage on direct mail (11022) are all "Included in Sales Price" (R.C. 5739.01(H)(1)(a)(iv)).
  - **Indiana,** for comparison. Its matrix marks the same charges included, with the comment "(USPS postage excluded from sales price, effective July 1, 2013)" (IC 6-2.5-1-7.5).
  - **California.** Reg 1628(a) excludes separately stated transportation charges for delivery "directly to the purchaser", which "shall not exceed the cost of the transportation to the retailer". It continues:
    - "If a separately stated charge is made designated 'postage and handling' or 'shipping and handling', only that portion of the charge which represents actual postage or actual shipment may be excluded."
    - "A separately stated charge designated 'handling' or 'handling charge' is not a separate statement of transportation charges. Tax applies to such charges."

    Reg 1628(b) covers sales at a "delivered price" and delivery in the retailer's own vehicles.
- **Grade:** 2 settled. California settles it by Reg 1628, and Indiana's postage carve-out by its matrix entry, the statute being unread.
- **Rhea's default:**
  - An invoice keeps goods, shipping and handling on separate lines. A shipping line carries the carrier (common carrier, USPS, own vehicle) and its actual cost.
  - Ohio: shipping and handling are taxable whenever the goods are.
  - California: separately stated common-carrier or USPS shipping at no more than actual cost is excluded. Handling, and any markup over cost, are taxable.
- **Rhea today:** there are no invoice lines. Missing.
- **Where the work lives:** *pack data*. The line kinds are the *shared ground* gap in MST-15.
- **Deadline:** —
- **Watch:** the SST matrix API (MST-16); CDTFA's law guide for Reg 1628.
- **Sources:** matrix rows (MST-16); https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1628.html

---

### MST-19 — Our design services: are they taxable outside Indiana, and do they count toward thresholds?
- **Applies to:** sole proprietor. For Indiana, see STX-11.
- **The law:**
  - **Indiana** (SIB 89): "with a few exceptions, Indiana generally does not impose sales tax on services". Services delivered in Indiana still count toward its $100,000 test.
  - **Ohio.** The tax reaches "certain sales of enumerated services". AR 2025 lists repair and installation, automatic data processing and computer services, telecommunications, lawn care, building maintenance and recreation clubs. Design is not in that list. I did not read R.C. 5739.01(B)(3) itself.
    - Only enumerated services count toward Ohio's threshold (SST Ohio page).
    - Design work done in Ohio on a client's behalf is a nexus-creating activity in ST 2001-01 §II. That makes it a presence question, not a tax-base question.
  - **California** taxes tangible property. The CDTFA-401 instructions: "Labor charges for making or fabricating a new product (such as labor charges for making a ring or furniture), or for assembling a product, are generally taxable."
- **Grade:** 2 settled for design sold on its own, which is a service outside each state's tax base. For California, a design fee bundled into a custom-made piece is part of what the instructions call fabrication. I did not read the regulation that would settle where a separately billed design fee for a custom piece falls (unverified).
- **Rhea's default:**
  - Design sold on its own is a service line, not taxed in Ohio or California, and it counts only toward Indiana's threshold.
  - Design sold with the piece it designs belongs to that piece's price and is taxed where the piece is.
  - A contract that mixes the two raises a task for the accountant.
- **Rhea today:** nothing. Missing.
- **Where the work lives:** *pack data*
- **Deadline:** —
- **Watch:** Ohio's list of enumerated services changes in the biennial budget (H.B. 110 removed employment services in 2021, AR 2025); CDTFA's law guide.
- **Sources:** SIB 89 (MST-02); AR 2025 (MST-04); ST 2001-01 (MST-02); https://www.cdtfa.ca.gov/formspubs/cdtfa401inst.pdf

---

### MST-20 — We sell through a marketplace: who collects the tax, and what do we still report?
- **Applies to:** sole proprietor (also any marketplace sales Hoosier makes). For Indiana, see STX-22.
- **The law:**
  - **Ohio.**
    - The MF collects "on all taxable retail sales made into Ohio that are facilitated by the marketplace facilitator" (2019 alert, citing R.C. 5741.01(T)).
    - The seller "only required to collect and remit tax on taxable sales through your website and catalog that are not marketplace facilitators" (alert FAQ).
    - A seller whose sales all go through an MF need not register (SST chart, Ohio row).
    - A registered seller "will include all of their sales on Line 1, 'Gross sales', of the UST1 and then include the sales that were made on the marketplace, on which tax was collected by the facilitator, on Line 2, 'Exempt sales'" (same chart).
    - Sellers with US gross receipts of $1 billion or more may take back the duty by waiver (R.C. 5741.071, alert); not relevant here.
  - **California.**
    - R&TC 6043: the MF "is the retailer selling or making the sale".
    - 6045: the seller registers "for retail sales made on its own behalf and not facilitated through a registered marketplace facilitator".
    - 6046: where the MF shows its error came from "incorrect or incomplete information provided … by the unrelated marketplace seller", the MF is relieved "and the marketplace seller is the retailer for that retail sale".
    - Reg 1684.5(c) applies these sections.
    - The CDTFA-401 instructions I read (rev. 34) name no line for facilitated sales. How a registered seller shows them is unverified.
  - **Indiana** (SIB 89): the MF "is considered the retail merchant of each retail transaction that is facilitated". Its gross retail income includes any fee it charges the buyer.
- **Grade:** 1 plain
- **Rhea's default:**
  - Maple Lane books each marketplace sale from the marketplace's settlement report as three amounts: the gross sale, the marketplace fee, and the tax the MF collected. That tax is passed through and is never Maple Lane's liability.
  - Marketplace sales count toward Ohio's and California's thresholds, not Indiana's (MST-08).
  - On returns where Maple Lane is registered, they are reported by each state's rule: Ohio lines 1 and 2; Indiana per STX-22.
  - The product data Maple Lane gives the marketplace is kept as a fact, because under 6046 wrong data makes Maple Lane the retailer.
- **Rhea today:** bank statement readers and the N-sided matching engine can match payouts to orders. There is no marketplace settlement-report reader and no notion of a sales channel on a sale. Missing.
- **Where the work lives:** *market code*: an adapter that reads the marketplace's settlement report. Plus *pack data*.
- **Deadline:** —
- **Watch:** SST marketplace-seller chart, rows dated (MST-02); CDTFA's "Tax Guide for the Marketplace Facilitator Act" (linked from the CDTFA FAQ); CDTFA's law guide for Reg 1684.5.
- **Sources:**
  - 2019 alert (MST-06)
  - https://www.streamlinedsalestax.org/for-businesses/marketplace-sellers
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6043.html ; https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6045.html ; https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6046.html
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1684-5.html
  - SIB 89 (MST-02)

---

### MST-21 — A customer in another state says it is buying for resale or is exempt: which certificate do we accept?
- **Applies to:** corporation chiefly (resellers); everyone for exempt buyers. For Indiana buyers, see STX-19 and STX-20.
- **The law:**
  - **SST certificate** (Form F0003, "Revised 12/21/2021"). It is "accepted by all 24 Streamlined member states".
    - Unless the single-purchase box is ticked, it "is considered a blanket certificate and remains effective until cancelled by the purchaser if purchases are no more than 12 months apart".
    - "The seller is not required to verify the purchaser's ID number."
    - The seller is not liable if "the fully completed exemption certificate is provided to the seller at the time of sale or within 90 days subsequent to the date of sale".
  - **SSUTA §317.** (C) relieves the seller with a certificate within 90 days. (D) gives "120 days subsequent to a request for substantiation" to get one "taken in good faith", meaning one that claims an exemption that "was statutorily available", "could be applicable to the item", and "is reasonable for the purchaser's type of business".
  - **Indiana** (SIB 89): "an Indiana General Sales Tax Exemption Certificate (ST-105) or a Streamlined Sales Tax Agreement Certificate of Exemption". Indiana is not among the states listed as accepting the MTC uniform certificate (revised 14 October 2022).
  - **Ohio** accepts the SST certificate and the MTC certificate (note 23): "The buyer must specify which one of the reasons for exemption on the certificate applies … Failure to specify the exemption reason will, on audit, result in disallowance", and "If no certificate is provided … at the time of the sale or within ninety days … it shall be presumed that the tax applies". Ohio's own blanket form is the STEC B. I saw only copies hosted by others, Rev. 7/25, not the Department's own file.
  - **California** (Reg 1668(a)). The seller is relieved if it "timely takes in good faith a certificate from the purchaser that the property is purchased for resale … from a person … who holds a California seller's permit".
    - It is timely if taken "at any time before the seller bills the purchaser …, or any time within the seller's normal billing and payment cycle, or any time at or prior to delivery".
    - It "remains in effect until revoked in writing".
    - The CDTFA FAQ: the document must say "for resale"; "Phrases such as 'nontaxable' or 'exempt' are not acceptable". The template is CDTFA-230.
    - The MTC certificate is "not valid as an exemption certificate. Its use is limited to use as a resale certificate" (MTC note 4).
  - **Records.** California keeps them at least four years (Reg 1698(i)).
- **Grade:** 2 settled, by the SSUTA, Reg 1668 and the states' notes to the MTC form.
- **Rhea's default:**
  - A certificate is an object: the buyer, the states covered, the reason (SST letters A–M), the buyer's ID for each state, single or blanket, the date received, and the signed file (a `received_document`).
  - A sale is exempt only if a certificate covering its state and reason exists, or arrives within that state's window: 90 days in Ohio and the SST states, before billing or delivery in California. Otherwise tax is charged, and a late certificate leads to a refund correction.
  - Rhea refuses the MTC form for Indiana. In California it accepts the MTC form for resale only.
- **Rhea today:** `received_document` (hash, name, media_type, source, received_on, status) can hold the signed file. There is no certificate type and no link from a sale to a certificate. Missing.
- **Where the work lives:** *pack data*
- **Deadline:** Ohio and SST: 90 days after the sale; on audit in SST states, 120 days after the state's request. California: at or before billing or delivery.
- **Watch:**
  - SST forms: the F0003 revision date is in its footer.
  - The MTC certificate page, which carries the revision date: https://www.mtc.gov/resources/uniform-sales-use-tax-exemption-certificate/
  - CDTFA forms (CDTFA-230).
- **Sources:**
  - https://www.streamlinedsalestax.org/for-businesses/exemptions ; https://www.streamlinedsalestax.org/docs/default-source/forms/exemption-certificateb926a7ab4a0d43e1ad4fe8eb19e79cbb.pdf?sfvrsn=857843d_7
  - SSUTA (MST-04)
  - https://www.mtc.gov/wp-content/uploads/2023/01/Unif-Resale-Cert-revised-10-14-22.pdf
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1668.html
  - CDTFA FAQ (MST-07)
  - https://consolidatedlabel.com/app/uploads/2026/08/OH-stec-b-blanket-exemption.pdf (secondary host — verify)

---

### MST-22 — A blanket certificate from a regular customer: how long does it last, and what about drop shipments?
- **Applies to:** corporation
- **The law:**
  - **Blanket certificates** (SSUTA §317(F)). The seller is relieved by a blanket certificate where there is "a recurring business relationship". A state "may not request from the seller renewal of blanket certificates". "A recurring business relationship exists when a period of no more than twelve months elapses between sales transactions."
  - **Drop shipments in SST states.** §317(A)(8): member states "must allow a third party vendor (e.g., drop shipper) to claim a resale exemption based on an exemption certificate provided by its customer/re-seller … regardless of whether the customer/re-seller is registered to collect and remit sales and use tax in the state where the sale is sourced". The F0003 instructions let the drop shipper accept an ID number "issued by another state".
  - **Drop shipments in California.** CDTFA FAQ: a retailer engaged in business in California, "including a retailer that meets the sales threshold", that delivers "directly to a California consumer on behalf of an unregistered supplier or vendor … is considered the retailer in the transaction". The tax is measured by "the amount the true retailer … charged the California consumer, or … the amount you charged the true retailer … plus a mark-up of 10 percent" (Reg 1706).
- **Grade:** 1 plain
- **Rhea's default:**
  - In SST states, a blanket certificate lapses in Rhea's view after 12 months with no sale to that customer. In California it lasts until revoked.
  - When Hoosier ships to a consumer in an SST state for a reseller, the reseller's certificate is enough, with an ID from any state.
  - In California, once Hoosier is over $500,000, a drop shipment for a reseller not registered in California makes Hoosier the retailer. The tax base is the reseller's price, or Hoosier's price plus 10%, and Rhea asks for the reseller's price.
- **Rhea today:** nothing. Missing.
- **Where the work lives:** *pack data*. The ship-to that differs from the buyer is the *shared ground* gap in MST-12.
- **Deadline:** —
- **Watch:** SSUTA amendments (MST-12); the CDTFA FAQ; CDTFA's law guide for Reg 1706.
- **Sources:** SSUTA (MST-04); F0003 (MST-21); https://www.cdtfa.ca.gov/industry/wayfair/frequently-asked-questions.htm

---

### MST-23 — Do we round the tax per line or per invoice when several jurisdictions apply?
- **Applies to:** everyone. For Indiana, see STX-16.
- **The law:**
  - SSUTA §324(A): "Tax computation must be carried to the third decimal place, and … rounded to a whole cent using a method that rounds up to the next cent whenever the third decimal place is greater than four."
  - §324(B): states "shall allow sellers to elect to compute the tax due on a transaction on an item or an invoice basis, and shall allow the rounding rule to be applied to the aggregated state and local taxes. No member state shall require a seller to collect tax based on a bracket system."
  - The Tech Guide computes tax "for each FIPS Code (taxing jurisdiction) separately".
  - California publishes "Tax Rate Charts (CDTFA-180 Sales Tax Reimbursement Schedules)" (rates page). I did not read California's rounding rule (unverified).
- **Grade:** 1 plain for the SST states. California is unverified.
- **Rhea's default:**
  - Rhea computes the tax per jurisdiction in exact rationals, sums state and local, and rounds the aggregate once per invoice with `round(x, 2, half_up)`.
  - The company declares invoice or item basis once.
  - California uses the same method until its own rule is read, and a task records that check.
- **Rhea today:** the formula language holds exact rationals and requires declared rounding (`half_up`, `half_even`, `down`, `up`). That fits. Present.
- **Where the work lives:** *pack data*
- **Deadline:** —
- **Watch:** SSUTA amendments (MST-12); the CDTFA rates page (CDTFA-180 charts).
- **Sources:** SSUTA (MST-04); Tech Guide (MST-13); https://www.cdtfa.ca.gov/taxes-and-fees/sales-use-tax-rates.htm

---

### MST-24 — A new quarter starts: which rates changed, and how much warning did we have?
- **Applies to:** everyone registered outside Indiana
- **The law:**
  - **State rates** (SSUTA §304(A)). Members make "a reasonable effort" to "Provide sellers with as much advance notice as practicable" and "Limit the effective date of a rate change to the first day of a calendar quarter". §304(C): with "less than thirty days" between enactment and effective date, a seller that kept collecting the old rate is relieved for up to "thirty days after the date of enactment".
  - **Local rates** (§305(A)(1)): "local rate changes will be effective only on the first day of a calendar quarter after a minimum of sixty days' notice to sellers". Boundary changes follow the same rule ((A)(3)).
  - **Data.** The databases must be available "by the first day of the month prior to the first day of a calendar quarter" (§305(D)). The Tech Guide grants liability relief for files posted later "until the first day of the following calendar quarter". §306 relieves reliance on erroneous state data.
  - **Late files in practice** (SST updates page):
    - the general Q4 2026 release posted on 1 September 2026, then Ohio's Q4 rate and boundary files on 17 September 2026 (OHR2026Q4SEP17, OHB2026Q4SEP17);
    - Ohio's Q1 2026 files posted on 4 December 2025, three days after the 1 December date, and the rate file was posted again on 27 January 2026.
  - **Ohio's notices in practice.** Bulletin of 10 December 2024 for 1 April 2025 (COTA), and of 1 May 2025 for 1 July 2025 (Lake County, 7.0% → 7.25%). That is 112 and 61 days (computed here).
  - **California districts.** R&TC 7265: "No ordinance … shall be operative on other than the first day of a calendar quarter, or prior to the first day of the first calendar quarter, commencing more than 110 days after the election on the ordinance". The measures of 2 June 2026 became operative on 1 October 2026 (MST-14). Computed here: 110 days after the general election of 3 November 2026 is 21 February 2027, so a measure passed that day can be operative no earlier than 1 April 2027.
  - **Indiana** has one statewide rate and no local sales tax (STX-10). Any change needs an act of the General Assembly.
- **Grade:** 1 plain
- **Rhea's default:**
  - Every rate is a dated event from a channel.
  - In the month before each quarter, the clock checks for new Ohio files on the SST updates page and for a new CDTFA explanation page.
  - A change that arrives less than 60 days before it takes effect, or after the quarter's posting date, raises a task. The SST relief applies, and Rhea records which file version each sale used.
- **Rhea today:** channels and the clock exist. `vat_rate` has no validity period, the same gap as Poland's VAT-05. Missing.
- **Where the work lives:** *market code* (the rate channels). Plus *shared ground*: rate master data needs validity dates and a jurisdiction.
- **Deadline:** rates take effect on 1 January, 1 April, 1 July and 1 October.
- **Watch:**
  - The SST updates page and its mailing list (MST-13).
  - Ohio GovDelivery bulletins, emailed and archived at content.govdelivery.com/accounts/OHTAX/bulletins/<id>.
  - CDTFA "Explanation of Tax Rate Changes" (MST-14).
  - California's election calendar (Secretary of State) for district measures.
  - The SST matrix `publishDate` for taxability changes (MST-16).
- **Sources:**
  - SSUTA (MST-04)
  - Tech Guide (MST-13)
  - https://www.streamlinedsalestax.org/for-businesses/rate-and-boundary-files/rate-and-boundary-files---updates
  - https://content.govdelivery.com/accounts/OHTAX/bulletins/3c678f3 ; https://content.govdelivery.com/accounts/OHTAX/bulletins/3de8394
  - https://www.cdtfa.ca.gov/lawguides/vol1/tutl/7265.html

---

### MST-25 — Our Ohio return: which return, how often, by when, and what discount?
- **Applies to:** everyone registered in Ohio (Hoosier from year one)
- **The law:**
  - **Account type** (AR 2025, Exhibit 1). "Seller — No Fee — Retailer located outside of Ohio who makes retail sales of property or services for storage, use, or consumption in Ohio". Accounts opened through the SSTRS are "Streamlined sales and streamlined use tax accounts".
  - **Monthly returns** (Exhibit 2), for vendors, sellers and transient vendors: "By the 23rd day of the month following the close of the reporting period".
  - **Semi-annual returns:** "may be authorized for vendors and sellers whose tax liability is less than $1,200 per six-month period".
  - **EFT and accelerated payment.** "Vendors are required to remit payment by electronic funds transfer (EFT) in cases where annual liability exceeds $75,000". They must also "make advance payment of 75 percent of each month's anticipated tax by the 23rd day of that month".
  - **Discount.** Timely payment entitles the vendor to "a discount of 0.75 of the amount due, up to $750" (R.C. 5739.12, 5741.12). Enacted by H.B. 96 and in force since 1 January 2026: "only receive a maximum of $750 in total vendor's discount per vendor's license per month".
  - **Zero returns.** "Once registered, a seller is required to file tax returns, even if the seller has no sales for the reporting period" (ST 2001-01).
  - **Filing** (Department tutorial, rev. January 2025). "Filing a business tax return should begin with Ohio Business Gateway and end with OH|TAX eServices". All Ohio locations "will all be filed on one UST 1 return". Taxable sales are entered county by county. A file upload is accepted, and "The UST 1 File Format Specifications in OH|TAX eServices will only accept .xlsx file type".
- **Grade:** 1 plain
- **Rhea's default:**
  - The Ohio return is built from the period's sale facts as taxable sales and tax per county and per transit code.
  - Rhea writes it as the UST 1 .xlsx upload.
  - The discount is computed at 0.75%, capped at $750 per licence per month, and only when the payment is dated by the 23rd.
  - The clock raises the duty monthly, or semi-annually once the Department authorises it.
- **Rhea today:** Poland's JPK_V7M writer shows the pattern for a statutory file writer, and the clock and period lock exist. Nothing for Ohio. Missing.
- **Where the work lives:** *market code* (a UST 1 upload writer, once the specification is read). Plus *pack data* (frequency, due date, discount).
- **Deadline:** the 23rd of the month after the period.
- **Watch:**
  - The Department's OH|TAX tutorials (dam.assets.ohio.gov/…/business/r2-tutorials/).
  - The "UST-1 Data File Upload" specification page on tax.ohio.gov, which I could not reach.
  - The annual report, for changes to the discount and the EFT threshold.
- **Sources:**
  - AR 2025 (MST-04)
  - https://dam.assets.ohio.gov/image/upload/v1732713572/tax.ohio.gov/business/r2-tutorials/file-original-and-amended-returns.pdf
  - ST 2001-01 (MST-02)

---

### MST-26 — Our California return: which return, how often, by when, and prepayments?
- **Applies to:** corporation, from year two
- **The law:**
  - **Due date.** R&TC 6451: "due and payable … quarterly on or before the last day of the month next succeeding each quarterly period".
  - **Frequency.** CDTFA "assigns a filing frequency (quarterly prepay, quarterly, monthly, fiscal yearly, yearly) based on your reported sales tax or your anticipated taxable sales". "A filing is required on or before the due date, even if there are no sales to report."
  - **Prepayments.** R&TC 6471(a): once notified in writing, a person "whose estimated measure of tax liability under this part averages seventeen thousand dollars ($17,000) or more per month" prepays "not less than 90 percent" of the tax for each of the first two months of a quarter. The second quarter has a variant that also covers the first 15 days of June. Due dates: "January prepayment — February 24", "February prepayment — March 24", and the quarterly return on 30 April (CDTFA filing dates).
  - **The form.** CDTFA-401-A, State, Local, and District Sales and Use Tax Return. District tax is computed on CDTFA-531-A2, Schedule A2, Long Form, line 16 (instructions rev. 34, 1-26).
  - **Registration.** An out-of-state retailer registers "for a Certificate of Registration-Use Tax" (Reg 1684(e)(1)).
- **Grade:** 1 plain
- **Rhea's default:**
  - The California return comes from the period's sale facts, with district tax allocated by the TAC and district code recorded on each sale (MST-14).
  - The prepayment duty starts when CDTFA's written notice is recorded as an event, and the clock fires on the 24th.
  - The return itself stays a task until CDTFA's filing channel is read.
- **Rhea today:** nothing. Missing.
- **Where the work lives:** *pack data* for the frequencies, dates and the return's figures. *later* for a writer: I verified no upload format.
- **Deadline:** the last day of the month after each quarter. Prepayments on the 24th of the month after each of a quarter's first two months.
- **Watch:** the CDTFA filing-dates page (https://www.cdtfa.ca.gov/taxes-and-fees/sales-use-tax-returns-filing-dates.htm); the CDTFA-401 instructions, with the revision number in the header; CDTFA's law guide for 6451 and 6471.
- **Sources:**
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6451.html ; https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6471.html
  - https://www.cdtfa.ca.gov/taxes-and-fees/sales-use-tax-returns-filing-dates.htm
  - https://www.cdtfa.ca.gov/formspubs/cdtfa401inst.pdf
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1684.html

---

### MST-27 — We find we should have been collecting in a state for months: how do we put it right?
- **Applies to:** everyone
- **The law:**
  - **Indiana VDA.** "Generally, the program is only available to customers that do not have a brick-and-mortar state tax filing obligation in Indiana". "Unregistered remote sellers … are now eligible." Benefits: "Written agreement to decrease the look-back period" and "Reduction of all penalties". Look-back: "Sales/Use Tax: Three full calendar years plus the current period."
  - **Ohio VDA.** The Department's FAQ (rev. 15 March 2016): the general look-back is 36 months. For tax collected from customers but not remitted, "there is no limitation to the lookback period", with a 10% penalty. Civil and criminal penalties are waived except on collected tax. ST 2001-01 §VI admits sellers "not yet registered with or contacted by the Department".
  - **California.** R&TC 6487.05: for a "qualifying retailer" the deficiency period "is limited to three years". Such a retailer is outside California, never registered, engaged in business under 6203, registers voluntarily, has not been contacted, and has reasonable cause. CDTFA FAQ: otherwise "CDTFA can assess a use tax liability up to eight years owed by a retailer that did not file a return".
  - **The MTC's joint programme.** It is free, with a "$500" minimum estimate. Its look-back charts (1 January 2025) do not list Indiana, Ohio or California among the participating states.
  - **SST amnesty** (SSUTA §402) applies only when the seller registers "within twelve months of the effective date of the state's participation". Indiana and Ohio are long-standing members, so it does not help.
- **Grade:** 1 plain. Ohio's FAQ dates from 2016, and whether a newer one exists is unverified.
- **Rhea's default:**
  - Threshold folds run over the log from the first sale, so a late-found crossing gives the exact past months, state by state.
  - Rhea drafts the disclosure package for the accountant: periods, sales and tax per period.
  - It goes to each state directly, since the MTC route does not cover these three.
- **Rhea today:** folds and tasks exist. No state rules. Missing.
- **Where the work lives:** *pack data* for the computation. *later* for the applications.
- **Deadline:** —
- **Watch:**
  - Indiana VDA page (in.gov/dor/legal-resources/voluntary-disclosure-program/).
  - Ohio VDA FAQ (dam.assets.ohio.gov/…/other/SUT/), which has a revision date in its name.
  - CDTFA's out-of-state voluntary disclosure page.
  - MTC look-back charts, dated PDFs under mtc.gov/wp-content/uploads/.
- **Sources:**
  - https://www.in.gov/dor/legal-resources/voluntary-disclosure-program/
  - https://dam.assets.ohio.gov/image/upload/tax.ohio.gov/other/SUT/ST_VDA_QA_GEN_FS.revised_3-15-16.pdf
  - https://www.cdtfa.ca.gov/lawguides/vol1/sutl/6487-05.html
  - https://www.mtc.gov/nexus/multistate-voluntary-disclosure-program/ ; https://www.mtc.gov/wp-content/uploads/2024/12/NNP-Lookback-period-chart-1-1-25r1.pdf
  - SSUTA (MST-04)

---

### MST-28 — We ship to every state: what pattern do the other 47 follow?
- **Applies to:** sole proprietor (ships everywhere); corporation if it widens
- **The law:** only as far as needed to state the pattern.
  - **No statewide sales tax.** The MTC look-back chart notes "Montana has no state sales/use tax" and "Oregon has no sales/use tax", and shows Delaware's sales/use look-back as "n/a". Alaska and New Hampshire also have no statewide sales tax (secondary — verify).
  - **Alaska, local taxes only, through one commission.** The ARSSTC: "Remote Sellers and Market Place Facilitators who meet the statewide threshold of $100,000 gross sales (the 200 transaction threshold was removed effective 1/1/2025) annually into Alaska either last year or this year, regardless of destination or taxability status of end buyer, must register with ARSSTC." The count uses "ALL sales into Alaska INCLUDING Marketplace Facilitator sales". Sellers must "Register … within 30 days of meeting the economic threshold" and keep records "for three (3) years".
  - **Home-rule and locally administered taxes.**
    - Colorado: home-rule cities with their own taxes, a shared filing portal (SUTS), and a 2024 law (HB24-1041) limiting non-portal cities' reach over remote sellers (secondary — verify; tax.colorado.gov refused access).
    - Louisiana: parish taxes, which remote sellers file on a single return through the Louisiana Sales and Use Tax Commission for Remote Sellers, at $100,000 (secondary — verify).
  - **The SST members** (MST-10) share one registration, one certificate, one rate and boundary format and one taxability matrix. The other sales-tax states do not.
- **Grade:** 1 plain where a primary source was read (MTC chart, ARSSTC). Colorado, Louisiana, Alaska's and New Hampshire's state-level status are unverified.
- **Rhea's default:**
  - Rhea counts Maple Lane's sales and orders per state, and Alaska's statewide, from day one.
  - A state passing a set margin of a common test ($100,000) raises a "read this state" task.
  - That state's rules ship only when written. No state is assumed clear.
- **Rhea today:** nothing. Missing.
- **Where the work lives:** *later*. The per-state counting is *pack data* now.
- **Deadline:** Alaska: register within 30 days of crossing.
- **Watch:** SST's remote-seller state guidance, which also covers some non-members (MST-08); ARSSTC news (https://arsstc.org/); the MTC's state tax amnesties page.
- **Sources:**
  - https://www.mtc.gov/wp-content/uploads/2024/12/NNP-Lookback-period-chart-1-1-25r1.pdf
  - https://arsstc.org/business-sellers/
  - https://zamp.com/resources/suts-the-deal-with-colorado-suts/ (secondary)
  - https://trykintsugi.com/sales-tax-guides/usa/louisiana (secondary)

---

### MST-29 — Taxes on our own gross receipts in other states (not sales tax): could any reach us?
- **Applies to:** everyone
- **The law:**
  - **Washington B&O.** RCW 82.04.067(1)(c)(i): substantial nexus with "More than one hundred thousand dollars of cumulative gross receipts from this state". (2)(a) counts "all of a person's gross income of the business attributed to this state".
  - **Ohio's commercial activity tax** (AR 2025).
    - From 1 January 2025 a person with "not more than … $6 million in taxable gross receipts" is excluded (H.B. 33). Above that, the rate is 0.26% on receipts over the exclusion.
    - "Bright-line presence" begins at "$500,000 in taxable gross receipts in Ohio", or $50,000 of property or payroll, or 25%.
    - Quarterly returns are "due by the 10th day of the second month after the end of each calendar quarter".
  - **Texas franchise tax.** For 2026 and 2027 reports: "No Tax Due Threshold $2,650,000"; rate "0.375%" for retail or wholesale; annual report "due May 15".
  - **Hawaii's general excise tax** is of this kind. Its statute was not reached (capitol.hawaii.gov blocked).
  - **Federal protection for net income tax only.** 15 U.S.C. 381(a) (P.L. 86-272) bars "a net income tax" where the only in-state activity is "the solicitation of orders … for sales of tangible personal property, which orders are sent outside the State for approval or rejection, and, if approved, are filled by shipment or delivery from a point outside the State". It protects neither gross receipts taxes nor sales tax collection. Under (b)(1) it never protects a corporation in its own state of incorporation.
- **Grade:** 1 plain
- **Rhea's default:**
  - A gross receipts tax is the seller's own tax on its receipts: an expense, never collected from buyers.
  - Rhea tracks receipts per state against each such threshold as data.
  - Neither company owes one in year one. They sell into Indiana, Ohio and California, and Hoosier's whole revenue (about $3 million) is under Ohio's $6 million exclusion (computed).
- **Rhea today:** nothing. Missing.
- **Where the work lives:** *later*
- **Deadline:** Ohio CAT: quarterly, 10 February, May, August and November, if ever above the exclusion.
- **Watch:**
  - Washington's RCW (app.leg.wa.gov), amended by session law.
  - Ohio's annual report, for CAT changes in budget bills.
  - The Texas Comptroller's franchise page, with thresholds set per two-year period (https://comptroller.texas.gov/taxes/franchise/).
- **Sources:**
  - https://app.leg.wa.gov/RCW/default.aspx?cite=82.04.067
  - AR 2025 (MST-04)
  - https://comptroller.texas.gov/taxes/franchise/
  - https://www.law.cornell.edu/uscode/text/15/381 (uscode.house.gov returned "Page Not Found" in its temporary format)

---

### MST-30 — Year end: does crossing this year keep us registered next year, and when can we close an account?
- **Applies to:** everyone
- **The law:**
  - All three tests look at "the current or preceding calendar year" (MST-05 to MST-07). Crossing in one year keeps the duty through the next.
  - **California.** Reg 1684(e)(3): a retailer remains registered "during any calendar year that it has a physical presence … or the total combined sales … exceed the five hundred thousand dollar ($500,000) threshold …, and during the following calendar year". It may stop on 1 January only if, in all of the preceding year, it had no such presence and did not exceed the threshold. (e)(4) lets it stay registered voluntarily.
  - **Ohio.** ST 2001-01 §V: "When an out-of-state seller no longer has nexus creating contacts, the out-of-state seller may cancel its license". But a new contact "within twelve months of canceling its license" is presumed "part of a regular presence", and the seller then reinstates and pays "on all its sales in this state during the interim period".
  - **Indiana.** SIB 89: a seller that "no longer meets one of the threshold due to the use of a registered marketplace facilitator" may close its account. Sellers that met only the old 200-transaction test in 2023 could close in 2024, but "will still have to file all required sales tax returns for 2024".
- **Grade:** 2 settled. California settles it by Reg 1684(e), Ohio by its release.
- **Rhea's default:**
  - On 1 January Rhea recomputes each state's test on the year just closed.
  - Registrations stay through the following year.
  - Closing an account is offered only when both years are under the test and no physical presence exists. In Ohio, Rhea shows the 12-month re-entry presumption with the offer.
- **Rhea today:** the clock can fire on 1 January, and the period lock closes months. There are no registration objects (MST-10). Missing.
- **Where the work lives:** *pack data*
- **Deadline:** 1 January each year.
- **Watch:** CDTFA's law guide for Reg 1684; Ohio information releases (MST-02); SIB 89.
- **Sources:** https://www.cdtfa.ca.gov/lawguides/vol1/sutr/1684.html ; ST 2001-01 (MST-02); SIB 89 (MST-02)


## Contractors, information returns, banks and money

Reviewed against Rhea's shared ground as it stands (the kernel's `Rule`/`Warrant`, the shared finance definitions in `testdata/seed/finance_v2–v9.json`, `packs/common/*.json`, the matching engine in `internal/match`, the MT940 and PDF readers in `internal/adapter`) and against how the Poland pack answered the same questions (`packs/pl/pack.json`, `packs/pl/nbp.json`, `packs/pl/bank-mt940.json`). Law as of 16 October 2026. Federal statute is quoted from the U.S. Code (Cornell LII, which carries the Public Law 119-21 amendments and their effective-date notes). Treasury regulations are quoted from the eCFR versioner API at 1 October 2026; the eCFR web pages refused automated access. IRS forms and publications are those posted for tax year 2026 where they exist; where only the 2025 edition exists, the entry says so. Abbreviations:
- "IRC" = Internal Revenue Code of 1986 (26 U.S.C.); "Reg." = Treasury Regulations (26 CFR, or 31 CFR for FinCEN).
- "OBBBA" = Public Law 119-21 of 4 July 2025 (the "One Big Beautiful Bill Act"). Its § 70432 is the 1099-K threshold and its § 70433 the $2,000 threshold.
- "Pub. 1099" = IRS Publication 1099 (2026), *General Instructions for Certain Information Returns*. This is the former "General Instructions", now a numbered publication (new publication of 27 March 2026).
- "1099 instructions" = *Instructions for Forms 1099-MISC and 1099-NEC*, labelled "(12/2026)" on irs.gov.
- "Rev. Proc. 2025-32" = the 2026 inflation adjustments.
- "Nacha Rules" = the Nacha Operating Rules. They bind a business through its bank (ODFI) agreement. The rule book is sold, not public, so every Nacha point is taken from nacha.org rule-summary pages and is marked secondary.
- "UCC" = Uniform Commercial Code. Indiana enacts Article 4 as IC 26-1-4 and Article 4A as IC 26-1-4.1. The uniform text is quoted (Cornell LII); the Indiana text was not fetched.
- "IC", "ORC", "CCP", "RTC" = the Indiana Code, Ohio Revised Code, California Code of Civil Procedure and Revenue and Taxation Code. All are quoted from FindLaw ("current as of January 01, 2026"), because iga.in.gov did not render, codes.ohio.gov refused the connection and leginfo.legislature.ca.gov returned 403.
- "ASC" = FASB Accounting Standards Codification. It is behind a login, so it is quoted through Deloitte's public roadmap and marked secondary.

Dates marked "computed here" apply IRC § 7503 (the next day that is not a Saturday, Sunday or "a legal holiday in the District of Columbia") to the 2027 calendar. 31 January 2027 is a Sunday; 15 February 2027 is Washington's Birthday; 28 February 2027 and 1 August 2027 are Sundays.

Entries are ordered by when a business meets them in its first year: setting up, first purchases and contractors, first sales, payments, month and quarter end, later in the year, year end.

---

### PAY-01 — Do we have to report our owners to FinCEN (the "BOI report")?
- **Applies to:** everyone
- **The law:** Corporate Transparency Act regulations, 31 CFR 1010.380(c)(1) as amended. The "reporting company" definition now reads: „(i) [Reserved] (ii) Any entity that is: (A) A corporation, limited liability company, or other entity; (B) Formed under the law of a foreign country; and (C) Registered to do business in any State or tribal jurisdiction …”. How the definition got there:
  - FinCEN's interim final rule of 26 March 2025 (90 FR 13688) removed U.S.-formed entities.
  - The final rule "Beneficial Ownership Information Reporting Requirement Revision" (91 FR 52508, 14 August 2026, RIN 1506-AB67, effective on publication) adopts that interim rule "as final and with certain limited changes". It also exempts U.S. person company applicants and lifts the duty on U.S. persons to update FinCEN ID information.
  - FinCEN's BOI page (alert of 11 August 2026): „U.S. companies are exempt from the Beneficial Ownership Information (BOI) reporting requirements”.
- **Grade:** 1 plain
- **Rhea's default:** no BOI report for an Indiana LLC or an Indiana corporation. The pack records this as a warrant-only fact on the company's legal form, so the question is answered with its source and not asked again. The duty comes back only if the company is formed abroad.
- **Rhea today:** the Poland pack's `company` (v7) carries `legal_form`; the shared `company` (finance_v2) carries only name, kind, vat_id and country. There is nowhere to say "formed under the law of" a state or country.
- **Where the work lives:** *pack data* (a `formed_in` field on the US company and one warranted rule; no filing to build)
- **Deadline:** none for domestic companies
- **Watch:**
  - The Federal Register API for FinCEN rules on 31 CFR 1010.380: https://www.federalregister.gov/api/v1/documents.json?conditions[agencies][]=financial-crimes-enforcement-network&conditions[term]=beneficial%20ownership
  - FinCEN's BOI page (alert banner): https://www.fincen.gov/boi
  - Congress.gov for any amendment to 31 U.S.C. 5336.
- **Sources:** https://www.federalregister.gov/documents/2026/08/14/2026-16576/beneficial-ownership-information-reporting-requirement-revision ; https://www.fincen.gov/boi ; https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-31.xml?section=1010.380

---

### PAY-02 — What records must we keep, for how long, and may they exist only in Rhea?
- **Applies to:** everyone
- **The law:**
  - IRC § 6001; Reg. 1.6001-1(a): any person subject to income tax, or required to file an information return, „shall keep such permanent books of account or records, including inventories, as are sufficient to establish the amount of gross income, deductions, credits, or other matters required to be shown”.
  - Reg. 1.6001-1(e): records „shall be retained so long as the contents thereof may become material in the administration of any internal revenue law”.
  - IRC § 6501(a): tax „shall be assessed within 3 years after the return was filed”, or 6 years where omitted income „is in excess of 25 percent of the amount of gross income stated in the return” (§ 6501(e)(1)(A)(i)).
  - Employment and backup-withholding records (Reg. 31.6001-1(e)(2)): „for at least four years after the due date of such tax for the return period to which the records relate, or the date such tax is paid, whichever is the later”.
  - Electronic-only records: Rev. Proc. 98-25 § 4.02(2)(a) applies its conditions to a taxpayer under $10 million of assets whenever „all or part of the information required by § 6001 is not in the taxpayer's hardcopy books and records, but is available in machine-sensible records”. Its § 5.01(1) says materiality lasts „at a minimum … until the expiration of the period of limitation for assessment, including extensions”. It also requires documentation of the system and records „capable of being processed”.
  - Scanned paper falls under Rev. Proc. 97-22 (not opened here).
  - Indiana: 1099s not sent to the state must be kept „for the statutory time period (currently three years)” (DOR Departmental Notice #5). Ohio: 1099-NEC records „for a period of at least four (4) years from the due date” (ODT 1099-NEC specifications).
  - Form 8300: kept „for at least five years from the date filed” (IRS Form 8300 reference guide).
- **Grade:** 2 settled — the regulation's "so long as … material" is open-ended; Rev. Proc. 98-25 settles the minimum and the conditions for records that exist only electronically
- **Rhea's default:** nothing is deleted (the log is append-only), so retention is met by construction. The pack adds the Rev. Proc. 98-25 duties:
  - a written description of the record system: event log, rule versions, replay;
  - an export a revenue agent can process: CSV or JSON of postings with their provenance;
  - a retention clock per tax year, which only blocks archiving and never deletes.
  - Received documents (PDF, XML) are kept as received, because they are the "hardcopy" a scan stands for.
- **Rhea today:** right in substance. Invariant 1 (append-only), invariant 4 (replay) and invariant 5 (provenance on every fact) are exactly what Rev. Proc. 98-25 asks for, and `received_document` keeps the original by hash. Missing: the system documentation and a processable export as named deliverables. The Polish pack has no equivalent rule.
- **Where the work lives:** *shared ground* (an "export for examination" of the log and postings with provenance is market-neutral) and *pack data* (the retention periods with warrants)
- **Deadline:** keep at least 3 years from filing (6 for a large omission); employment and backup-withholding records 4 years
- **Watch:**
  - IRS GuideWire (advance copies of revenue procedures, posted to https://www.irs.gov/pub/irs-drop/), subscription at https://www.irs.gov/newsroom/e-news-subscriptions
  - the weekly Internal Revenue Bulletin, https://www.irs.gov/irb
  - the eCFR versions API for § 1.6001-1: https://www.ecfr.gov/api/versioner/v1/versions/title-26.json?section=1.6001-1
- **Sources:** https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=1.6001-1 ; https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=31.6001-1 ; https://www.law.cornell.edu/uscode/text/26/6501 ; https://www.irs.gov/pub/irs-drop/rp-98-25.pdf ; https://www.in.gov/dor/files/dn05.pdf ; https://dam.assets.ohio.gov/image/upload/tax.ohio.gov/employer_withholding/2025-2026%20NYP%20Updates/2025_1099NEC_Specs_V1.pdf ; https://www.irs.gov/businesses/small-businesses-self-employed/irs-form-8300-reference-guide

---

### PAY-03 — Our books are in dollars. When a bill comes in Canadian dollars, which exchange rate do we use, and whose?
- **Applies to:** corporation (sole proprietor only if it buys or sells in a foreign currency)
- **The law:**
  - IRC § 985: all determinations are made in the functional currency. The IRS page: „The U.S. dollar is the functional currency for all taxpayers except some qualified business units (QBUs).”
  - There is no statutory rate source: „The Internal Revenue Service has no official exchange rate. Generally, it accepts any posted exchange rate that is used consistently.” The same IRS page says: „In general, use the exchange rate prevailing (i.e., the spot rate) when you receive, pay or accrue the item.”
  - Reg. 1.988-1(d)(1): the spot rate is „a fair market rate of exchange available to the public for currency under a spot contract in a free market and involving representative amounts”. Acceptable sources include „exchange rates published by the Board of Governors of the Federal Reserve System pursuant to 31 U.S.C. section 5151”.
  - Reg. 1.988-1(d)(2): inconsistent sources that distort income let the IRS pick the rate.
  - Reg. 1.988-1(d)(3): „If consistent with the taxpayer's financial accounting, a taxpayer may utilize a spot rate convention determined at intervals of one quarter year or less” for ordinary-course payables and receivables (for example, a monthly rate).
  - The IRS "yearly average currency exchange rates" (Canada dollar 2025: 1.398 per USD) are for translating foreign income, foreign taxes and QBU results. They are not a rate for daily bookkeeping.
  - Books: ASC 830-20-30-1 measures a foreign-currency transaction at the exchange rate in effect at the date it is recognized (secondary — verify; via Deloitte DART).
- **Grade:** 2 settled — no statute names a rate; Reg. 1.988-1(d) and the IRS's published position settle "any posted spot rate, used consistently"
- **Rhea's default:** a US rate channel publishes, per day, the Federal Reserve Board's H.10 daily rates (USD per CAD and others). A document is converted at the rate for its own date; Poland converts at the previous working day's rate. On a day with no rate (weekend, holiday) the last earlier rate applies. The company declares the source once, and changing it is a dated event, which is how the consistency the regulation demands becomes visible. A monthly-convention option, behind the gate, is allowed only for payables and receivables.
- **Rhea today:**
  - `fx_rate` (code, base, quote, date, rate as a string) and `posting` v3 (book amount plus `tx_amount`/`tx_currency`) are market-neutral and fit.
  - The rate channel exists for Poland only (`packs/pl/nbp.json`, `markets/pl/nbp.go`), and its warrant is the Polish "previous working day" rule. That is pack data, so the US choice of the same day does not touch the kernel.
  - Missing: a US rate reader, and a declared "rate source" on the company.
- **Where the work lives:** *market code* (an H.10 or FRED reader in `markets/us`) and *pack data* (the same-day rule, the source declaration)
- **Deadline:** none
- **Watch:**
  - H.10 release page (weekly): https://www.federalreserve.gov/releases/h10/current/default.htm. On 13 October 2026 it announced that the Data Download Program's "Build Your Package" feature will be removed the week of 9 November, ahead of the "eventual retirement of the DDP", and pointed to FRED (St. Louis Fed). The reader should be built on FRED (series DEXCAUS) from the start (not fetched here — verify the series).
  - The IRS foreign-currency page: https://www.irs.gov/individuals/international-taxpayers/foreign-currency-and-currency-exchange-rates
- **Sources:** https://www.irs.gov/individuals/international-taxpayers/foreign-currency-and-currency-exchange-rates ; https://www.irs.gov/individuals/international-taxpayers/yearly-average-currency-exchange-rates ; https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=1.988-1 ; https://www.federalreserve.gov/releases/h10/current/default.htm ; https://dart.deloitte.com/USDART/home/codification/broad-transactions/asc830-10/roadmap-foreign-currency-transactions-translations/chapter-4-foreign-currency-transactions/4-3-subsequent-measurement-foreign-currency (secondary)

---

### PAY-04 — The owner puts money in and takes money out. Is that income, expense, or neither — and what if we are a corporation?
- **Applies to:** everyone (different answers)
- **The law:**
  - Sole proprietor: Reg. 301.7701-3(b)(1)(ii). A domestic LLC with one owner, absent an election, is „Disregarded as an entity separate from its owner if it has a single owner.” Its profit is the owner's Schedule C profit whether or not it is drawn. Money the owner puts in or takes out is not income or expense of the business; nothing is reported.
  - Corporation: a distribution to a shareholder out of earnings is a dividend, reported on Form 1099-DIV for a shareholder paid „$10 or more in money or other property” (1099-DIV instructions; IRC § 6042). The $2,000 change did not touch § 6042.
  - A shareholder loan to the corporation is a liability; interest paid on it is covered in PAY-13.
  - S-corporation variant, where it changes the books: an S corporation does not issue 1099-DIV for distributions (they flow through the K-1). A shareholder who works in the business is paid wages through payroll. Distributions are a separate equity line.
- **Grade:** 1 plain
- **Rhea's default:**
  - Sole proprietor: an owner-equity account pair (contributions, draws) and a rule that a bank line identified as the owner's (by the owner's own account number) posts to equity, never to income or expense.
  - Corporation: capital stock, additional paid-in capital, retained earnings and distributions accounts. A declared dividend raises a 1099-DIV duty at $10.
  - The company's form (disregarded LLC, C corporation, S corporation) is declared with its effective date, and the equity rules are gated on it.
- **Rhea today:** the Poland pack already has the shape. `pl-explain-bank-operation` recognizes an operation as the owner's, `company` v7 has `owner` (ref<person>) and `owner_account`, and `pl-identify-company-owner` exists. The US pack can mirror it. Missing: US equity accounts and the 1099-DIV duty.
- **Where the work lives:** *pack data*
- **Deadline:** 1099-DIV to shareholders 1 February 2027 (computed here from 31 January)
- **Watch:** the IRS "About Form 1099-DIV" page "Recent developments": https://www.irs.gov/forms-pubs/about-form-1099-div ; and IRS GuideWire for any entity-classification change.
- **Sources:** https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=301.7701-3 ; https://www.irs.gov/instructions/i1099div

---

### PAY-05 — Our bank gives us statements as BAI2, camt.053, OFX or CSV. Which must Rhea read?
- **Applies to:** everyone
- **The law:** none. No US rule of law prescribes a statement format for a business customer.
  - Practice, secondary throughout:
    - BAI2, the Bankers Administration Institute format, is still the dominant US corporate statement format.
    - BAI transferred its copyright to ASC X9 in 2008. The successor is ANSI X9.121-2016 "Balance and Transaction Reporting Standard" (BTRS v3), which retired many BAI2 codes. Adoption in practice is unverified.
    - ISO 20022 camt.053 is offered by larger banks.
    - OFX/QFX and CSV come from online banking at small banks.
  - The CFPB's personal financial data rights rule (12 CFR 1033) concerns consumer accounts and does not give a business a right to a feed (not researched further).
- **Grade:** 1 plain (no rule of law governs; practice only)
- **Rhea's default:**
  - Sole proprietor (a small or regional bank): CSV and OFX/QFX first.
  - Corporation: BAI2 first, camt.053 second. Its wires already carry ISO 20022 data (PAY-20).
  - Every reader keeps the bank's raw line, as the MT940 reader keeps field :86:.
  - Matching keys: ABA routing plus account number for counterparties, check number, ACH trace number and company ID, wire IMAD/OMAD. There is no IBAN in a domestic US payment.
- **Rhea today:**
  - The MT940 reader is in shared ground (`internal/adapter/mt940.go`) and works as the model.
  - The matching engine is generic. The Poland pack keys its two matching rules by IBAN (`pl-match-incoming`, `pl-match-outgoing`) and `company.iban` is a Poland-pack field. The US keys are pack data, so nothing in the kernel stops them.
  - `bank_operation` (v3, with direction, counterparty_account, kind, open) is a Poland-pack type the US pack can copy.
  - Missing: every US reader.
- **Where the work lives:** *market code* (BAI2/BTRS reader in `markets/us`; OFX and bank-CSV readers), *shared ground* (a camt.053 reader belongs beside MT940 in `internal/adapter`, since it serves every market), *pack data* (the matching keys)
- **Deadline:** none
- **Watch:**
  - ASC X9 standards announcements for X9.121 (ANSI webstore): https://webstore.ansi.org/standards/ascx9/ansix91212016
  - The corporation's bank's own file-format notices. These are not public law and should be watched per bank.
- **Sources:** https://blog.ansi.org/ansi/x9-balance-and-transaction-reporting/ (secondary) ; https://webstore.ansi.org/standards/ascx9/ansix91212016 (secondary) ; https://eco.com/support/en/articles/17190462-bai2-vs-mt940-vs-camt-053-bank-statement-formats-explained (secondary — verify)

---

### PAY-06 — Before we pay a freelancer, what must we collect from them?
- **Applies to:** everyone
- **The law:**
  - IRC § 6109 requires the payee's TIN on an information return.
  - For nonemployee compensation, rents and similar payments, Reg. 31.3406(d)-1(d): „the payee must furnish the payee's taxpayer identification number to the payor either orally or in writing”. A certificate under penalties of perjury is not required, except after IRS notices (§ 31.3406(d)-5).
  - Reasonable-cause relief from the penalty for a missing TIN depends on solicitation, Reg. 301.6724-1(e)(1)(i): „An initial solicitation for a payee's correct TIN must be made at the time an account is opened.” Annual solicitations follow when none is received.
  - Form W-9 is the IRS's form for this: revision March 2024; "Recent developments: None at this time" (irs.gov, page updated 27 June 2026).
  - Pub. 1099, for a payee that is a single-member LLC disregarded under Reg. 301.7701-3: „enter the owner's name only on the first name line and the LLC's name on the second name line”, with the TIN type the W-9 instructions direct. If the LLC is taxed as a corporation or partnership, „enter the entity's EIN”.
  - Pub. 1099 also allows a payee's TIN to be truncated on payee statements.
- **Grade:** 2 settled — the regulation lets a TIN be given orally; the reasonable-cause regulation (§ 301.6724-1(e)) and universal practice make the signed W-9 the standard
- **Rhea's default:**
  - A counterparty cannot be paid for services until a W-9 is on file as a received document.
  - The W-9 yields facts: legal name, business name, federal tax classification (individual/sole proprietor, C corporation, S corporation, partnership, LLC with its tax code), exempt payee code, TIN type and TIN.
  - A missing W-9 at the first payment raises a backup-withholding duty (PAY-12), not a refusal to pay.
  - The same applies to Maple Lane's own W-9 given to its processor and marketplace. As a disregarded LLC it gives the owner's name and TIN, so the 1099-Ks it receives match the owner's Schedule C (PAY-15).
- **Rhea today:**
  - `received_document` (hash, source, received_on) can hold the W-9.
  - The PDF reader shows the pattern for reading it.
  - There is no TIN on any shared type: `company` has `vat_id`, the Poland pack's `contractor` has `nip`.
  - The larger gap is shared ground. A sole proprietor's TIN is often a Social Security number, a secret. The Polish NIP is public, so the log never had to hold one. An SSN written into an append-only, replayable log can never be removed.
- **Where the work lives:**
  - *pack data*: the W-9 facts and the classification enum.
  - *shared ground*: a field-level sensitivity class that keeps a secret identifier out of the plain log while keeping replay deterministic, for example a sealed value with a stable fingerprint, the way the KSeF UPO fingerprint is verified. Without it, the US pack cannot store a contractor's SSN at all. Masking a value in views is not enough.
  - *market code*: a W-9 reader (later).
- **Deadline:** at the time the relationship opens (Reg. 301.6724-1(e)(1)(i)). If no TIN came, the first annual solicitation is due „on or before December 31 of the year in which the account is opened (for accounts opened before December) or January 31 of the following year” (§ 301.6724-1(e)(1)(ii)).
- **Watch:**
  - "About Form W-9", Recent developments: https://www.irs.gov/forms-pubs/about-form-w-9
  - IRS draft forms list (a new W-9 revision appears there first): https://apps.irs.gov/app/picklist/list/draftTaxForms.html
- **Sources:** https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=31.3406(d)-1 ; https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=301.6724-1 ; https://www.irs.gov/forms-pubs/about-form-w-9 ; https://www.irs.gov/publications/p1099

---

### PAY-07 — Can we check a freelancer's name and TIN with the IRS before we file?
- **Applies to:** everyone
- **The law:**
  - IRS e-Services TIN Matching, run under Rev. Proc. 2003-9, per Pub. 2108 (Rev. 11-2024). The IRS page: „TIN Matching is only for payers and their authorized agents that submit information returns. To use this service, a payer must be listed in the IRS Payer Account File (PAF) database.”
  - Pub. 2108 § 6(13): „Authorized Payor – means a Payor who has filed Information Returns with the IRS in at least one of the two preceding tax years.”
  - Limits: interactive „up to 25 name/TIN combinations at a time”; bulk „.txt files composed of up to 100,000 name/TIN combinations”, with a response „within 24 hrs”.
  - Result codes: 0 match; 1 missing or not 9 digits; 2 not issued; 3 no match; 4 invalid; 5 duplicate; 6, 7, 8 matched on SSN, EIN or both when the TIN type is unknown.
- **Grade:** 1 plain
- **Rhea's default:** in the first year neither business is an "authorized payor". Rhea relies on the W-9 (PAY-06), runs local checks (nine digits, not all one digit, the EIN prefix list — not researched here) and treats the IRS's later mismatch notices (CP2100/972CG) as the matching. From the second year, a bulk TIN-matching file is written before every January filing.
- **Rhea today:** missing. Nothing in Poland is like it: the Polish checks are public (the VAT white list).
- **Where the work lives:** *later* (year two: *market code* for the bulk .txt request and the response reader)
- **Deadline:** none (practical: before the January filing)
- **Watch:** the TIN Matching page (https://www.irs.gov/tax-professionals/taxpayer-identification-number-tin-matching) and Pub. 2108 revisions on https://www.irs.gov/pub/irs-pdf/p2108.pdf
- **Sources:** https://www.irs.gov/tax-professionals/taxpayer-identification-number-tin-matching ; https://www.irs.gov/pub/irs-pdf/p2108.pdf

---

### PAY-08 — Is this payment reportable at all, and on 1099-NEC or 1099-MISC?
- **Applies to:** everyone
- **The law:**
  - IRC § 6041(a): „All persons engaged in a trade or business and making payment in the course of such trade or business to another person” of rents, salaries, wages, premiums, compensations and other fixed or determinable income „of $2,000 or more in any calendar year” must file a return.
  - IRC § 6041A(a) does the same for remuneration for services.
  - The 1099 instructions' test for nonemployee compensation (1099-NEC box 1a): the payee is not an employee; the payment is for services in the course of the payer's business; the payee is an individual, partnership or estate (or, in listed cases, a corporation); and the year's payments to the payee are at least $2,000.
  - 1099-MISC carries rents (box 1, $2,000), royalties (box 2, „$10 or more”), other income (box 3, $2,000), medical and health care payments (box 6, $2,000) and gross proceeds to an attorney (box 10, „$600 or more”).
  - Not reportable:
    - Reg. 1.6041-3(c): „Payments of bills for merchandise, telegrams, telephone, freight, storage, and similar charges”.
    - Reg. 1.6041-3(d): rent paid to rental agents (the agent reports).
    - Wages (W-2).
    - Payments to tax-exempt organizations and governments.
    - Payments to corporations (PAY-10a, below).
- **Grade:** 2 settled — the statute is general; the regulations' exception list and the IRS's four-condition test settle it
- **Rhea's default:** every purchase line carries a reporting class decided by rule, never guessed from the text:
  - services → NEC;
  - rent → MISC 1;
  - royalties → MISC 2;
  - attorney gross proceeds (settlements paid through counsel) → MISC 10;
  - attorney fees → NEC;
  - medical → MISC 6;
  - goods, freight, telephone, storage → none.
  - Services billed with parts are reportable in full unless the parts are separately stated (practice — secondary — verify).
  - The class sits on the purchase-invoice line, and the 1099 sums payments, not invoices (PAY-09).
- **Rhea today:** `purchase_invoice` (shared v1; Poland v6) has no line-level class and is VAT-shaped (`vat`, `vat_rate`). The rule language can carry the class as a derived field once the US purchase type has it.
- **Where the work lives:** *pack data* (a `reporting_class` per purchase line, the classification rules with warrants); *shared ground* only insofar as `purchase_invoice` assumes VAT fields (see PAY-16)
- **Deadline:** none (the class must exist before the payment)
- **Watch:** the 1099-MISC/NEC instructions and "About Form 1099-NEC" recent developments: https://www.irs.gov/instructions/i1099mec ; https://www.irs.gov/forms-pubs/about-form-1099-nec
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6041 ; https://www.law.cornell.edu/uscode/text/26/6041A ; https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=1.6041-3 ; https://www.irs.gov/instructions/i1099mec

---

### PAY-09 — What is the reporting threshold for payments we make in 2026, and will it change?
- **Applies to:** everyone
- **The law:**
  - IRC § 6041(a) as amended by OBBBA § 70433(a): „of $2,000 or more in any calendar year”; it was $600, and "taxable year" became "calendar year".
  - § 6041A(a)(2) now cross-refers: „equals or exceeds the dollar amount in effect for such calendar year under section 6041(a)”.
  - § 70433(f): the amendments „shall apply with respect to payments made after December 31, 2025”.
  - Indexing, new § 6041(h): „In the case of any calendar year after 2026, the dollar amount in subsection (a)” is increased by the § 1(f)(3) cost-of-living adjustment „determined by substituting “calendar year 2025” for “calendar year 2016””. Any increase „shall be rounded to the nearest multiple of $100”.
  - Rev. Proc. 2025-32 § 3.15: „For payments made after December 31, 2025, the base threshold under section 6041(a) is $2,000. This base threshold amount is adjusted for inflation for returns required to be filed in calendar year 2027.” The 1099 instructions (12/2026) say instead: „increased to $2,000 and may be adjusted for inflation beginning in calendar year 2027”.
  - Not changed by OBBBA:
    - gross proceeds to attorneys under § 6045(f): „$600 or more” (1099-MISC box 10);
    - royalties: $10;
    - dividends: $10 (PAY-04).
  - The regulations still say $600 (Reg. 1.6041-1 is still headed „payments of $600 or more”).
  - I found no 2027 figure on irs.gov as of 16 October 2026. Bloomberg Tax projects $2,100 (secondary — verify).
- **Grade:** 2 settled — the statute decides $2,000 for 2026 payments. The Rev. Proc. sentence could be read as indexing 2026 payments, but the statute ("calendar year after 2026") and the TY2026 instructions settle that indexing starts with payments in 2027.
- **Rhea's default:**
  - The threshold is dated master data: $2,000 for payments in 2026; the 2027 amount is entered when the IRS publishes it.
  - It is tested per payee per calendar year on payments made (cash), never on invoices accrued. This matters for the accrual-method corporation: a December invoice paid in January belongs to next year's 1099.
  - Attorney gross proceeds keep $600. Royalties and dividends keep $10.
- **Rhea today:** the formula language (exact rationals, folds under a cap) can sum a year's payments per payee, and the clock can fire the January duty. No US threshold data exists.
- **Where the work lives:** *pack data*
- **Deadline:** none (feeds PAY-28)
- **Watch:**
  - The annual inflation revenue procedure. The IRS's 2026–2027 Priority Guidance Plan lists it for October 2026 (per irs.gov search listing, not opened). It is posted to https://www.irs.gov/pub/irs-drop/ and announced on IRS GuideWire.
  - The inflation-adjustment landing page: https://www.irs.gov/inflationadjustment (not opened).
  - The § 6041 text on uscode.house.gov release points: https://uscode.house.gov/download/download.shtml
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6041 ; https://www.law.cornell.edu/uscode/text/26/6041A ; https://www.irs.gov/pub/irs-drop/rp-25-32.pdf ; https://www.irs.gov/instructions/i1099mec ; https://www.ecfr.gov/api/versioner/v1/versions/title-26.json?section=1.6041-1 ; https://pro.bloombergtax.com/insights/federal-tax/2027-projected-tax-rates-special-report/ (secondary)

---

### PAY-10 — We pay one freelancer by credit card and another through PayPal, and one is an incorporated design studio. Do we still file 1099s for them?
- **Applies to:** everyone
- **The law:**
  - Card and network payments, Reg. 1.6041-1(a)(1)(iv): transactions „that otherwise would be subject to reporting under both sections 6041 and 6050W are reported under section 6050W and not section 6041”. And: „Solely for purposes of this paragraph, the de minimis threshold for third party network transactions in § 1.6050W-1(c)(4) is disregarded”.
  - So a payment made through a third party settlement organization is excluded from the payer's 1099-NEC even when the network itself files no 1099-K because the payee is under $20,000 and 200 transactions. The 1099 instructions: such payments „must be reported on Form 1099-K by the payment settlement entity”.
  - Corporations, Reg. 1.6041-3(p)(1): payments to a corporation are excepted, „except with respect to payments made to a corporation after December 31, 1997 for attorneys' fees, and except a corporation engaged in providing medical and health care services”.
  - The 1099 instructions also keep reportable to corporations: gross proceeds to attorneys (MISC 10), medical (MISC 6), fish for resale (MISC 11) and substitute payments (MISC 8).
  - An LLC is a corporation for this purpose only if it is taxed as a C or S corporation, which the W-9 shows.
- **Grade:** 1 plain
- **Rhea's default:**
  - Every payment carries its means (check, ACH, wire, card, network). A payment by card or through a network that is a third party settlement organization is excluded from the 1099 sum, whatever the network's own threshold.
  - A bank-to-bank transfer service that is not a settlement organization stays reportable. Which apps are TPSOs is a recorded fact about the payment channel, not verified per app here.
  - A payee whose W-9 says C or S corporation is excluded unless the line's class is attorney or medical.
- **Rhea today:** payments exist as bank operations and settlements (Poland pack), with no "means" attribute. The W-9 classification is missing (PAY-06).
- **Where the work lives:** *pack data*
- **Deadline:** none
- **Watch:** Reg. 1.6041-1 and 1.6050W-1 via the eCFR versions API (https://www.ecfr.gov/api/versioner/v1/versions/title-26.json?section=1.6050W-1); the IRS "Understanding your Form 1099-K" page (https://www.irs.gov/businesses/understanding-your-form-1099-k)
- **Sources:** https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=1.6041-1 ; https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=1.6041-3 ; https://www.irs.gov/instructions/i1099mec

---

### PAY-11 — Our Canadian suppliers ship us goods and invoice in CAD. Do we owe them a 1099, or withhold anything?
- **Applies to:** corporation
- **The law:**
  - Reg. 1.6041-4(a)(1): „Returns of information are not required for payments that a payor can, prior to payment, reliably associate with documentation upon which it may rely to treat as made to a foreign beneficial owner”. That documentation is Form W-8BEN-E for an entity.
  - Withholding on foreign persons (chapter 3) reaches fixed or determinable annual or periodical income. Reg. 1.1441-2(b)(2)(i) excludes from it „Gains derived from the sale of property”, so the price of goods bought is not subject to it.
  - Services performed outside the United States are foreign-source (IRC § 862(a)(3); not quoted here — verify).
- **Grade:** 1 plain
- **Rhea's default:** a foreign supplier is documented once with a W-8BEN-E (held as a received document, with its three-year validity — not researched here). Payments to it carry no 1099 and no withholding when they are for goods, or for services performed abroad. A foreign supplier without a W-8 at payment raises a task. Its US-source services (work done in the US) would need a separate review (later).
- **Rhea today:** `company.country` exists in shared ground; foreign-status documentation does not.
- **Where the work lives:** *pack data*
- **Deadline:** before the first payment
- **Watch:** "About Form W-8BEN-E" (https://www.irs.gov/forms-pubs/about-form-w-8-ben-e, not opened); Pub. 515 annual revision (https://www.irs.gov/publications/p515, not opened)
- **Sources:** https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=1.6041-4 ; https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=1.1441-2

---

### PAY-12 — A freelancer won't give us a TIN, or the IRS says it's wrong. When do we withhold, how much, and how is it filed?
- **Applies to:** everyone
- **The law:**
  - IRC § 3406(a)(1): backup withholding applies if „(A) the payee fails to furnish his TIN to the payor in the manner required”, „(B) the Secretary notifies the payor that the TIN furnished by the payee is incorrect”, or on notified underreporting (C) or a certification failure (D).
  - Rate: „the fourth lowest rate of tax applicable under section 1(c)”, which Pub. 1099 (2026) gives as „a 24% rate”.
  - § 3406(b)(6), amended by OBBBA § 70433(d): § 6041 and § 6041A(a) payments are reportable payments only once the year's aggregate reaches „the dollar amount in effect for such calendar year under section 6041(a)” ($2,000 in 2026), or where a return was required for the payee for the prior year.
  - New § 3406(b)(8) (OBBBA § 70432(b)) does the same for third-party network settlements against the $20,000 and 200 thresholds.
  - Timing (Pub. 1099):
    - missing TIN: withhold „until the TIN is furnished”; a W-9 marked "Applied For" gives „60 days”;
    - IRS incorrect-TIN ("B") notice: „you must withhold on any reportable payment made to the account more than 30 business days after you received the notice”;
    - a payee is treated as failing to furnish a TIN „if the TIN furnished does not contain the proper number of digits” (§ 3406(h)(1)).
  - Filing: backup withholding is reported on Form 945, and deposits are made separately from Form 941 deposits („Don't combine deposits for Forms 941 … with deposits for Form 945”). The latest instructions are the 2025 edition; they apply monthly or semiweekly schedules, and no deposits are needed under $2,500 for the year.
  - California: payers required to backup-withhold federally „are also required to withhold and remit to the FTB on income sourced to California”, at „7 percent of the payment” (RTC § 18664; FTB Form 592 instructions 2026). Indiana and Ohio backup-withholding rules were not researched.
- **Grade:** 1 plain
- **Rhea's default:**
  - A payee without a usable TIN gets a "backup withholding" status from the payment at which the year's total reaches the threshold (or from the first payment if a return was required for the prior year).
  - From then on, each payment posts 24% to a federal backup-withholding liability, plus 7% to a California liability for California-source payments.
  - The liability is deposited under the Form 945 schedule, and the 1099 shows it in box 4. A B-notice starts a 30-business-day clock.
  - The 24% is read from master data tied to § 1(j)(2)(C), not hard-coded.
- **Rhea today:** the clock (dated duties), the formula language (exact rationals) and partial settlement in matching are all there. No withholding liability exists anywhere in shared ground; Poland has no backup withholding.
- **Where the work lives:** *pack data*
- **Deadline:** Form 945 for 2026: 1 February 2027, or 10 February 2027 if every deposit was on time (computed here from the 2025 instructions' 31 January / 10 February pattern)
- **Watch:** "About Form 945" and its yearly instructions (https://www.irs.gov/instructions/i945); Pub. 1281 (B-notice procedures); FTB withholding pages (https://www.ftb.ca.gov/pay/withholding/backup-withholding.html)
- **Sources:** https://www.law.cornell.edu/uscode/text/26/3406 ; https://www.irs.gov/publications/p1099 ; https://www.irs.gov/instructions/i945 ; https://www.ftb.ca.gov/pay/withholding/backup-withholding.html ; https://www.ftb.ca.gov/forms/2026/2026-592-instructions.html

---

### PAY-13 — We buy on a business credit card and borrowed from a bank and from the owner. What do the books hold, and is interest we pay reportable?
- **Applies to:** everyone (interest reporting: corporation)
- **The law:**
  - Card timing for a cash-method payer: Rev. Rul. 78-39 holds that a charge on a bank credit card is payment „in the year the credit card charge is made regardless of when the bank is repaid” (medical expenses; Rev. Rul. 78-38 the same for charity). Secondary sources apply it to cash-method business expenses (secondary — verify).
  - Interest: the small-business exception to the § 163(j) interest limit turns on the § 448(c) gross-receipts test, $32,000,000 for taxable years beginning in 2026 (Rev. Proc. 2025-32 § 3.30). Both businesses are far under it.
  - Reporting interest the corporation pays on a shareholder loan, where two texts disagree:
    - The 1099-INT instructions, a continuous-use edition dated 01/2024, still say file for „at least $600 of interest paid in the course of your trade or business”.
    - That trade-or-business interest is reported under § 6041, whose threshold OBBBA raised to $2,000, and Pub. 1099 (2026) speaks of „the minimum threshold for certain information returns increased to $2,000”.
- **Grade:** 3 open — for interest paid on a shareholder loan, $600 (unrevised 1099-INT instructions, 01/2024) or $2,000 (§ 6041(a) as amended; Pub. 1099 2026). The card-timing point is 2 settled by Rev. Rul. 78-39 by analogy.
- **Rhea's default:**
  - A card is a liability account. Each charge books the expense (cash method: deductible on the charge date) against the card liability. The monthly statement is reconciled like a bank account, and the payment to the issuer clears the liability without touching expense.
  - Loans: principal is a liability; interest accrues (corporation) or posts when paid (sole proprietor). A loan agreement is a received document whose schedule drives dated duties.
  - Interest on a shareholder loan: report on 1099-INT at $600 until the IRS revises the instructions. Over-reporting carries no penalty, and under-reporting would.
- **Rhea today:** postings, accounts and matching handle this shape. The Poland pack recognizes interest and fees on bank lines (`pl-explain-bank-operation`). A card-statement reader does not exist.
- **Where the work lives:** *pack data* (accounts, rules); *market code* (card-statement reader, which issuers export as CSV/OFX — later for the first year if lines are entered by hand)
- **Deadline:** 1099-INT to recipients 1 February 2027 (computed here)
- **Watch:** "About Form 1099-INT" recent developments (https://www.irs.gov/forms-pubs/about-form-1099-int, not opened); a revision of https://www.irs.gov/instructions/i1099int
- **Sources:** https://bradfordtaxinstitute.com/Endnotes/Rev_Rul_78-39.pdf (copy of the ruling — secondary) ; https://www.irs.gov/pub/irs-drop/rp-25-32.pdf ; https://www.irs.gov/instructions/i1099int ; https://www.irs.gov/publications/p1099

---

### PAY-14 — Our card processor and marketplace pay us out net of fees, refunds and chargebacks. What do the books record?
- **Applies to:** everyone
- **The law:**
  - Gross income from business includes gross receipts (IRC § 61). Fees are deductions (§ 162), and refunds are returns and allowances (Schedule C line 2), not netted out of receipts.
  - The processor reports gross: Reg. 1.6050W-1(a)(6) defines "gross amount" as the total „without regard to any adjustments for credits, cash equivalents, discount amounts, fees, refunded amounts or any other amounts. The dollar amount of each transaction is determined on the date of the transaction.”
  - The Schedule C instructions (2025 edition, the latest posted) say: „Be sure to check any Forms 1099 you received for business income that must be reported on this line.” Where 1099-NEC totals exceed line 1, „attach a statement explaining the difference”.
  - Corporation: ASC 606 principal-versus-agent. The seller of its own goods is principal and records revenue gross, with processor fees as expense (secondary — not researched).
- **Grade:** 2 settled for the gross-up (Reg. 1.6050W-1(a)(6), Schedule C instructions). Whether a cash-method seller has income on the card-transaction date or on the payout date at a year-end cut-off: unverified, no primary source reached.
- **Rhea's default:**
  - Each sale books gross revenue (and sales tax collected, PAY-16) against a "processor clearing" receivable on the transaction date, for both methods. That date is also the date the processor's 1099-K counts.
  - A payout report books fees, refunds, chargebacks, chargeback fees and reserve holds against the same clearing account.
  - The bank deposit then clears the net, through N-sided matching of many sales to one deposit.
  - At year end, the clearing balance is the money in transit and is shown as such.
- **Rhea today:** N-sided matching (many invoices to one bank line, netting, unpairing) is exactly the engine needed. A processor or marketplace payout reader does not exist, and there is no clearing-account rule.
- **Where the work lives:** *market code* (payout-report readers per processor and marketplace: CSV exports or APIs, one per provider behind the adapter contract) and *pack data* (the gross-up rules)
- **Deadline:** none
- **Watch:** Reg. 1.6050W-1 via the eCFR versions API (https://www.ecfr.gov/api/versioner/v1/versions/title-26.json?section=1.6050W-1); each processor's payout-report format changelog (provider documentation, not law)
- **Sources:** https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=1.6050W-1 ; https://www.irs.gov/instructions/i1040sc

---

### PAY-15 — Will we get a 1099-K, from whom, and must our books agree with it?
- **Applies to:** everyone (above all the sole proprietor)
- **The law:**
  - IRC § 6050W(e) as amended by OBBBA § 70432 ("as if included in" ARPA § 9674, so retroactively): a third party settlement organization reports a payee's third party network transactions „only if— (1) the amount … exceeds $20,000, and (2) the aggregate number of such transactions exceeds 200.” The IRS reads the earlier $5,000/$2,500/$600 phase-in (Notice 2024-85) as superseded (FTB summary).
  - Payment card transactions („any transaction in which a payment card is accepted as payment”, § 6050W(c)(2)) have no de minimis: the IRS says a card processor issues a 1099-K „no matter how many payments you got or how much they were for”.
  - Statements to payees are due „on or before January 31 of the year following the calendar year” (§ 6050W(f)).
  - A payee that gives no TIN is subject to backup withholding by the processor (1099-K instructions: „A payment made by a TPSO is a reportable payment potentially subject to section 3406 backup withholding”).
  - Clients paying Maple Lane ≥ $2,000 for design services by check or ACH issue it a 1099-NEC.
- **Grade:** 1 plain
- **Rhea's default:**
  - The card processor's 1099-K is expected every year. The marketplace's is expected only above $20,000 and 200 transactions.
  - A received 1099-K or 1099-NEC is read into facts and reconciled to the clearing-account gross by month, since 1099-K boxes 5a–5l are monthly (not verified for 2026).
  - Differences (refunds, chargebacks, sales tax, timing) are explained on a view, not hidden. The owner gets a task in early February to check them.
- **Rhea today:** the received-document pipeline and readers (KSeF, PDF) show the pattern; nothing reads a 1099.
- **Where the work lives:** *pack data* (the reconciliation view and rules); *market code* (a 1099 PDF reader — later; hand entry is enough in year one)
- **Deadline:** expect the forms by 1 February 2027 (computed here)
- **Watch:** "About Form 1099-K" and https://www.irs.gov/instructions/i1099k ; the IRS gig-economy page https://www.irs.gov/businesses/understanding-your-form-1099-k ; Congress.gov for any further § 6050W amendment
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6050W ; https://www.law.cornell.edu/uscode/text/26/3406 ; https://www.irs.gov/instructions/i1099k ; https://www.irs.gov/businesses/understanding-your-form-1099-k ; https://www.ftb.ca.gov/file/business/information-returns.html

---

### PAY-16 — Is the sales tax we collect our money? And what about sales through the marketplace?
- **Applies to:** everyone
- **The law:**
  - Indiana: the merchant „shall collect the tax as agent for the state” (IC 6-2.5-2-1(b)). A person who must remit it „holds those taxes in trust for the state and is personally liable for the payment of those taxes” (IC 6-2.5-9-3).
  - A marketplace facilitator „shall be considered the retail merchant of each retail transaction” it facilitates and must „Collect and remit the gross retail tax” (IC 6-2.5-4-18).
  - Ohio: tax „shall be paid by the consumer to the vendor” and „each vendor shall collect from the consumer, as a trustee for the state of Ohio” (ORC 5739.03(A)).
  - California: „For the privilege of selling tangible personal property at retail a tax is hereby imposed upon all retailers” (RTC § 6051). The tax is the retailer's own liability whether or not it collects reimbursement.
  - Nexus, rates and filing are another section's subject; this entry concerns only how collected tax sits in the books.
- **Grade:** 1 plain
- **Rhea's default:**
  - Sales tax charged is a liability per jurisdiction (state plus local), never revenue. Its remittance clears it.
  - Sales made through the marketplace carry the marketplace's tax, which never enters Maple Lane's liability. The sale is still Maple Lane's revenue (gross), and the marketplace's fee is an expense.
  - In California, the liability exists on taxable sales even where reimbursement was not charged.
- **Rhea today:** wrong shape for the US. The shared `sales_invoice` (finance_v8) requires `vat`, `vat_rate` (ref<vat_rate>), `seller` and `buyer` (ref<company>). The shared `vat_rate` has `percent` as an **int**.
  - US combined rates are fractional (7.25% in California; local add-ons give two decimals). An int percent cannot hold them.
  - A consumer buyer is not a `company`.
  - Sales tax is per jurisdiction on the ship-to address, not one rate per invoice.
- **Where the work lives:** *shared ground*:
  - `vat_rate.percent` must become an exact decimal or rational (the formula language already has rationals);
  - `sales_invoice` must allow a tax amount by jurisdiction without being VAT;
  - the buyer must be allowed to be a person or an anonymous consumer.
  - The Polish VAT assumption (one VAT rate per invoice, a VAT-registered buyer) is what stands in the way.
  - *pack data* for the liability accounts.
- **Deadline:** per state filing calendar (other section)
- **Watch:** Indiana DOR Information Bulletins (https://www.in.gov/dor/legal-resources/tax-library/, not opened); Ohio Tax Alerts (tax.ohio.gov); CDTFA publications; the state legislatures' session laws (Indiana acts usually take effect 1 July)
- **Sources:** https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-2-1.html ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-9-3.html ; https://codes.findlaw.com/in/title-6-taxation/in-code-sect-6-2-5-4-18.html ; https://codes.findlaw.com/oh/title-lvii-taxation/oh-rev-code-sect-5739-03.html ; https://codes.findlaw.com/ca/revenue-and-taxation-code/rtc-sect-6051.html ; /home/krzysztof/Rhea/testdata/seed/finance_v8.json ; /home/krzysztof/Rhea/testdata/seed/finance_v4.json

---

### PAY-17 — A customer pays more than $10,000 in cash. What must we do?
- **Applies to:** everyone (craft fairs; a wholesale counter)
- **The law:**
  - IRC § 6050I(a): any person „who is engaged in a trade or business, and” „who, in the course of such trade or business, receives more than $10,000 in cash in 1 transaction” „(or 2 or more related transactions)” must file a return.
  - Cash includes foreign currency and certain monetary instruments with a face amount of up to $10,000 (§ 6050I(d)). Per the IRS guide, cashier's checks, bank drafts, traveler's checks and money orders count only in a designated reporting transaction or when the business knows the customer is avoiding reporting.
  - Digital assets were added by the 2021 Infrastructure Act, but Announcement 2024-4 says they are not counted toward the $10,000 until regulations are issued.
  - Related transactions: within 24 hours, or more than $10,000 over 12 months in installments (IRS guide).
  - Filing:
    - „Generally, a business must file Form 8300 within 15 days after they receive the cash.”
    - It must be e-filed if the business must file at least 10 other information returns that year, through the „Bank Secrecy Act (BSA) Electronic Filing (E-Filing) System”.
    - A written statement goes to the payer „on or before January 31” of the next year (§ 6050I(e)).
  - Penalty for intentional disregard, for returns due in 2027: the greater of $34,930 or the cash received up to $139,500 (Rev. Proc. 2025-32 § 3.57(3)).
- **Grade:** 1 plain (digital assets: 2 settled by Announcement 2024-4)
- **Rhea's default:** cash receipts are tagged by payer. A rolling rule sums a payer's cash over 24 hours and over 12 months and raises a task "Form 8300 due" with the 15-day date, plus the January statement duty. Rhea prepares the data; the owner files in BSA E-Filing (no API used in year one).
- **Rhea today:** the clock and the formula folds can do the rolling sums. No cash-receipt type exists.
- **Where the work lives:** *pack data*; *later* (a BSA E-Filing writer — FinCEN batch XML, not researched)
- **Deadline:** 15 days after the cash is received; payer statement 1 February 2027 for 2026 receipts (computed here)
- **Watch:** "About Form 8300" (https://www.irs.gov/forms-pubs/about-form-8300); FinCEN BSA E-Filing notices (https://bsaefiling.fincen.treas.gov, not opened); the Federal Register API for IRS rules under § 6050I (digital-asset regulations)
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6050I ; https://www.irs.gov/businesses/small-businesses-self-employed/irs-form-8300-reference-guide ; https://www.irs.gov/pub/irs-drop/a-24-04.pdf ; https://www.irs.gov/pub/irs-drop/rp-25-32.pdf

---

### PAY-18 — We pay suppliers and the payroll provider by ACH. What do the ACH rules require of us as a business that sends payments?
- **Applies to:** everyone (as originator through its bank)
- **The law:** the Nacha Rules, binding through the bank's origination agreement (all points secondary — rule text not public):
  - **Fraud monitoring.** Each non-consumer Originator must „establish and implement risk-based processes and procedures” reasonably intended to identify entries suspected of being unauthorized or authorized under False Pretenses, and review them „at least annually”. False Pretenses covers business email compromise, vendor impersonation and payroll impersonation. Nacha names change controls over vendor and payroll payment instructions as an example procedure.
    - Phase 1 took effect 20 March 2026 for Originators with 2023 origination of 6 million entries or more.
    - Phase 2 took effect 19 June 2026, in practice Monday 22 June 2026, for everyone else, including both businesses here.
  - **Company Entry Description.** From 20 March 2026, „PAYROLL” for PPD wage credits (leftmost 7 characters) and „PURCHASE” for consumer e-commerce debits. This applies to whoever builds the file, usually the payroll provider or the bank.
  - **Authorization.** A debit to a business account (CCD) needs an agreement with the receiver binding it to the Rules (not fetched — verify).
  - **Same Day ACH.** Per-payment limit $1 million since 18 March 2022; Nacha: „On Sept. 17, 2027, the Same Day ACH per-payment limit will increase tenfold from $1 million to $10 million.”
  - ACH credits between businesses are funds transfers under UCC Article 4A (IC 26-1-4.1) where Regulation E does not apply.
- **Grade:** 1 plain (the Rules decide; quoted through Nacha's summaries)
- **Rhea's default:**
  - Rhea does not originate ACH in year one; the bank portal does.
  - Rhea is the change control. A supplier's new bank account is an event that holds payments to that supplier until a second person approves it, and every approval is in the log. This is the "risk-based process" the Phase 2 rule asks for, and it costs nothing extra in an approval-gated system.
  - A payment instruction arriving by email is never a source of bank details.
- **Rhea today:**
  - People and roles (`person.role`, `deputy`) and the approval gate exist.
  - Bank details live on the Poland-pack `company.iban`, with no versioned "bank account of a counterparty" whose change needs approval.
  - A NACHA file writer does not exist.
- **Where the work lives:** *pack data* (counterparty bank account as a gated, versioned fact); *later* (*market code*: a NACHA-format payment file writer if Rhea ever originates)
- **Deadline:** Phase 2 already in force (22 June 2026); Same Day ACH limit change 17 September 2027
- **Watch:** Nacha "Summary of Upcoming Rule Changes" (https://www.nacha.org/content/summary-upcoming-rule-changes) and ACH Operations Bulletins on https://www.nacha.org/news ; the corporation's bank's ACH origination agreement updates
- **Sources:** https://www.nacha.org/rules/risk-management-topics-fraud-monitoring-phase-2 (secondary) ; https://www.nacha.org/rules/risk-management-topics-company-entry-descriptions (secondary) ; https://www.nacha.org/content/abcs-ach (secondary) ; https://www.nacha.org/sites/default/files/2022-03/Same_Day_ACH_Guidance_$1M.pdf (secondary)

---

### PAY-19 — An ACH payment came back, or a customer's debit was disputed. How long can that happen, and what do the books do?
- **Applies to:** everyone (ACH debits of customers: corporation)
- **The law:** the Nacha Rules (secondary):
  - an unauthorized debit to a non-consumer account is returned (R29) so it reaches the ODFI „no later than the opening of business on the second banking day following the settlement date”;
  - an unauthorized debit to a consumer account (R10) may come back until the banking day after the 60th calendar day after settlement, with a Written Statement of Unauthorized Debit;
  - a sender's reversal must be made „within five Banking Days of the Settlement Date” and only for the errors the Rules allow;
  - an improper reversal can itself be returned (R11 consumer within 60 days, R17 non-consumer within two banking days).
  - Most administrative returns (insufficient funds, closed account) follow the general two-banking-day window (not fetched — verify).
- **Grade:** 1 plain
- **Rhea's default:** a returned or reversed entry is a bank line of opposite direction carrying the original trace number. It unpairs the original settlement (reversal = new event, original kept), reopens the invoice and raises a task with the return reason. A customer debit stays "at risk" for 2 banking days (business account) or 60 days (consumer) after settlement, shown on the cash view and not netted away.
- **Rhea today:** unpairing exists and is the right mechanism: `objects.matched` with `reversal_of`, and the same rules explain the reversal. Missing: return-reason codes and trace numbers in any reader (they come with BAI2 or camt.053, PAY-05).
- **Where the work lives:** *pack data* (reason codes, at-risk windows); *market code* (the codes come through the bank reader)
- **Deadline:** 2 banking days / 60 days / 5 banking days as above
- **Watch:** Nacha rules pages and Operations Bulletins (https://www.nacha.org/rules) ; the Federal Reserve's FedACH services notices (https://www.frbservices.org/news/communications)
- **Sources:** https://www.nacha.org/rules/differentiating-unauthorized-return-reasons (secondary) ; https://www.nacha.org/rules/reversals-and-enforcement (secondary)

---

### PAY-20 — We pay Canadian suppliers by wire. What changed with wires, and what does a wire tell the books?
- **Applies to:** corporation
- **The law:**
  - Funds transfers are governed by UCC Article 4A (Indiana IC 26-1-4.1, not fetched). Under § 4A-505, once the bank has notified the customer of a debited payment order, „the customer is precluded from asserting that the bank is not entitled to retain the payment” unless it objects „within one year”.
  - The Fedwire Funds Service moved to ISO 20022 messages on 14 July 2025. This date comes from frbservices.org search listings; the implementation-center page text I fetched does not restate it — verify.
  - The next Fedwire release, which removes the unstructured postal address, was moved from November 2026 to November 2027. Notice of 27 August 2026: „Federal Reserve Financial Services (FRFS) has decided to reschedule the Fedwire Funds Service release from November 2026 to November 2027”, following Swift's delay of its own November 2026 standards release.
- **Grade:** 1 plain
- **Rhea's default:**
  - A wire is matched by the bank's reference (IMAD/OMAD or end-to-end ID) and the beneficiary's name and account.
  - A CAD wire from a USD account records the USD actually debited, which fixes the realized exchange gain or loss (PAY-23).
  - A wire's bank fee is booked separately, never netted into the supplier payment.
  - Supplier addresses are kept structured (street, town, postal code, country), because the 2027 release will require them.
- **Rhea today:** addresses are a single string (`company.address` in the Poland pack). FX settlement with realized difference exists (`pl-post-fx-difference`).
- **Where the work lives:** *pack data* (structured address on the US counterparty, the wire matching keys); *market code* (wire details come through the BAI2/camt reader)
- **Deadline:** structured-address readiness before the November 2027 Fedwire release (date to be announced "this fall")
- **Watch:** Fedwire ISO 20022 Implementation Center and FRFS communications (https://www.frbservices.org/resources/financial-services/wires/iso-20022-implementation-center ; https://www.frbservices.org/news/communications); Swift standards release notices
- **Sources:** https://www.law.cornell.edu/ucc/4A/4A-505 ; https://www.frbservices.org/news/communications/082726-fedwire-november-release-rescheduled ; https://www.frbservices.org/resources/financial-services/wires/iso-20022-implementation-center

---

### PAY-21 — We still write checks. How long is a check good, how do we stop one, and what is "positive pay"?
- **Applies to:** everyone (above all the corporation)
- **The law:** UCC Article 4 (Indiana IC 26-1-4; uniform text quoted):
  - § 4-404: „A bank is under no obligation to a customer having a checking account to pay a check, other than a certified check”, which is presented more than six months after its date. The bank may still pay it in good faith and charge the account.
  - § 4-403(b): „A stop-payment order is effective for six months, but it lapses after 14 calendar days” „if the original order was oral and was not confirmed in a record within that period”. It may be renewed for further six-month periods.
  - Positive pay (the business sends its bank the list of checks issued, and the bank pays only those) is a bank service under the account agreement, not law. Practice — secondary.
- **Grade:** 1 plain
- **Rhea's default:**
  - A check is a payment document with number, payee, amount and date. Until it clears it is "issued, outstanding", a liability to the payee, not cash spent.
  - At 180 days uncleared, the check becomes "stale" and raises a task: reissue, void or keep. A void reverses the payment and reopens the payable; nothing is deleted.
  - The dormancy clock for unclaimed property starts from the obligation, not from the void (PAY-25).
  - A positive-pay issue file per bank is a writer for later.
- **Rhea today:** the clock fires dated duties. Reversal-as-new-event exists. No check register type.
- **Where the work lives:** *pack data* (check register, stale and stop rules); *later* (*market code*: positive-pay file writer, bank-specific)
- **Deadline:** 6 months (stale); 14 days (oral stop order unconfirmed); 6 months (written stop order)
- **Watch:** Indiana General Assembly session laws for IC 26-1-4 amendments (https://iga.in.gov); the Uniform Law Commission's UCC amendments (2022 amendments on emerging technologies — Indiana adoption not checked)
- **Sources:** https://www.law.cornell.edu/ucc/4/4-404 ; https://www.law.cornell.edu/ucc/4/4-403

---

### PAY-22 — At month end, what must the bank reconciliation prove, and how fast must we look?
- **Applies to:** everyone
- **The law:**
  - Reg. 1.6001-1(a): records „sufficient to establish the amount of gross income” (PAY-02).
  - For a Schedule C, IRS examiners are told to „Compare the total deposits with the reported Gross Income” (IRM 4.10.4.2.3.7, bank account analysis, a minimum income probe). For corporations the probe is a balance-sheet analysis, where „Cash accounts should always be analyzed to identify unusual transactions” (IRM 4.10.4.2.4.1).
  - The customer's duty to look, UCC § 4-406(c): the customer „must exercise reasonable promptness in examining the statement or the items”. Forged items by the same wrongdoer are barred after „a reasonable period of time, not exceeding 30 days”, and every claim of an unauthorized signature or alteration is barred if not reported „within one year after the statement or items are made available” (§ 4-406(f)).
  - The ACH window for an unauthorized debit to a business account is two banking days (PAY-19).
- **Grade:** 2 settled — no statute requires a reconciliation; the IRM's minimum income probes and the UCC preclusion periods settle what it must show and how quickly
- **Rhea's default:**
  - Bank lines are read daily, and unexplained debits raise a task the same day: the two-banking-day ACH window and the 30-day forgery window make a monthly look too slow.
  - The month closes (period lock) only when every bank and card account ties: opening + lines = closing, every line matched or explained.
  - The Schedule C view shows "deposits vs. gross receipts", the IRS's own probe, with the differences named: owner contributions, loans, transfers, sales tax, refunds.
- **Rhea today:** right in shape. The MT940 reader refuses a statement where opening + lines ≠ closing (`document.unreadable`); matching and the period lock exist. Missing: the deposits-vs-receipts view, and US readers (PAY-05).
- **Where the work lives:** *pack data* (views, close checklist)
- **Deadline:** daily review (practice); one year (UCC § 4-406(f))
- **Watch:** IRM 4.10.4 revisions (https://www.irs.gov/irm/part4/irm_04-010-004, transmittal dated 29 August 2025); IRM revisions are announced in the IRM "What's New" list (https://www.irs.gov/irm, not opened)
- **Sources:** https://www.irs.gov/irm/part4/irm_04-010-004 ; https://www.law.cornell.edu/ucc/4/4-406

---

### PAY-23 — We owe a Canadian supplier CAD 50,000 and the rate moved before we paid. Is the difference income, and do we revalue at month end?
- **Applies to:** corporation
- **The law:**
  - Tax: IRC § 988(c)(1)(B)(ii) makes a § 988 transaction of accruing an expense „which is to be paid or received after the date” it is accrued.
  - Gain is limited to what is „realized by reason of changes in exchange rates on or after the booking date and before the payment date” (§ 988(b)(1); loss likewise, (b)(2)). The payment date is „the date on which the payment is made or received” (§ 988(c)(3)).
  - The gain or loss „shall be computed separately and treated as ordinary income or loss” (§ 988(a)(1)(A)). It is recognized when realized, so an unpaid payable at year end produces no tax gain or loss.
  - Books: ASC 830-20-35-1 and 35-2 remeasure a foreign-currency payable at each balance-sheet date and take the change to net income (secondary — verify; via Deloitte DART). The result is a book–tax difference for any payable open at year end.
- **Grade:** 1 plain
- **Rhea's default:**
  - The payable is booked at the bill-date rate (PAY-03).
  - Payment books the realized difference to "realized FX gain/loss", the same in the book and tax ledgers.
  - Month end remeasures open CAD balances in the book ledger only ("unrealized FX"). The entry reverses on the first day of the next month and is absent from the tax ledger.
  - The year-end difference appears as a reconciling item between book income and taxable income.
- **Rhea today:**
  - Right foundation: `posting.book` lets one fact post differently in two books, and `pl-post-fx-difference` shows a realized difference at settlement.
  - Missing: month-end remeasurement rules, and a tax book distinct from the GAAP book for the US.
- **Where the work lives:** *pack data*
- **Deadline:** month end (books); none for tax until payment
- **Watch:** Reg. 1.988-2 via the eCFR versions API (https://www.ecfr.gov/api/versioner/v1/versions/title-26.json?section=1.988-2); FASB's ASC update list (Accounting Standards Updates, https://www.fasb.org/standards — not opened)
- **Sources:** https://www.law.cornell.edu/uscode/text/26/988 ; https://dart.deloitte.com/USDART/home/codification/broad-transactions/asc830-10/roadmap-foreign-currency-transactions-translations/chapter-4-foreign-currency-transactions/4-3-subsequent-measurement-foreign-currency (secondary)

---

### PAY-24 — An outside payroll provider runs our payroll. What must our books hold from it?
- **Applies to:** corporation
- **The law:**
  - Pub. 15 (2026), § 16: the employer remains responsible for deposits and filings „even if you contract with a third party to perform these acts. You remain responsible if the third party fails to perform any required action.” Also: „If the third party fails to make the deposits and payments, the IRS may assess penalties and interest on the employer's account.”
  - A payroll service provider is not liable as employer or agent; a CPEO or a § 3504 agent is (Pub. 15 § 16). Pub. 15 recommends using the EFTPS Inquiry PIN „to monitor its account and ensure the third party is making the required tax deposits”.
  - Year-end accrued pay of an accrual-method corporation: compensation received after „the 15th day of the 3rd calendar month after the end of the employer's taxable year” is presumed deferred (Reg. 1.404(b)-1T, Q&A-2(b)(1)).
  - Forms 941, 940, W-2 and the Indiana WH-1/WH-3 are the provider's filings (out of scope), but their figures are the employer's liability.
- **Grade:** 1 plain
- **Rhea's default:** each pay run's provider journal is read as one received document. It books, by employee group:
  - gross wages expense;
  - employee withholdings as liabilities: federal income tax, employee Social Security and Medicare, Indiana state and county income tax, benefit deductions;
  - employer taxes as expense and liability: employer Social Security and Medicare, FUTA, Indiana unemployment;
  - net pay as a liability, cleared by the bank debit;
  - the provider's tax impound debit as a transfer to "taxes held by provider", cleared only by the provider's deposit confirmations.
  - At month end, earned-but-unpaid wages are accrued. A December accrual paid after 15 March is flagged.
  - A provider-reported deposit is checked against EFTPS, entered from the Inquiry PIN view, and a mismatch raises a task. The employer's liability is the reason.
- **Rhea today:** postings and matching can carry it. Missing: a provider-journal reader (CSV/API, provider-specific) and US payroll liability accounts. The Poland pack's ZUS handling (`zus_account`) is the analogue.
- **Where the work lives:** *market code* (provider-journal readers, one per provider behind the adapter contract) and *pack data* (accounts, the accrual rule, the EFTPS check)
- **Deadline:** per pay run (journal); month end (accrual); provider's deposit schedule (semiweekly/monthly, Pub. 15)
- **Watch:** Pub. 15 annual revision (https://www.irs.gov/publications/p15); "Outsourcing payroll and third-party payers" (https://www.irs.gov/businesses/small-businesses-self-employed/outsourcing-payroll-and-third-party-payers, not opened)
- **Sources:** https://www.irs.gov/publications/p15 ; https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=1.404(b)-1T

---

### PAY-25 — A supplier never cashed our check, and a customer has an old credit balance. Is that our money after a while?
- **Applies to:** everyone (above all the corporation)
- **The law:**
  - Indiana, the Revised Uniform Unclaimed Property Act (IC 32-34-1.5), presumes property abandoned after:
    - „For a debt of a business association, three (3) years after the obligation to pay arises” (-4(4)), which covers outstanding vendor checks;
    - „For money owed to a customer as a result of a retail business transaction, three (3) years after the obligation arose” (-4(6));
    - one year for wages and other compensation (-4(11), paraphrase).
  - Excluded: „This chapter does not apply to a business to business credit memorandum” or „a credit balance resulting from a business to business credit memorandum” (-1(b)). An uncashed check to a business vendor is a debt, not a credit memorandum, and stays reportable.
  - The report „must be filed before November 1 of each year and cover the twelve (12) months preceding July 1 of that year” (-20(a)), „in an electronic record to the attorney general” (-18(a)), and „signed … and verified” (-19).
  - Due diligence on property of $50 or more, no more than 180 and no less than 60 days before filing (IC 32-34-1.5-23, as the State Board of Accounts' 2025 training presents it — the statute text was not reached).
  - Which state takes the property: the owner's last known address, else the holder's state of incorporation (Texas v. New Jersey, 379 U.S. 674 (1965); not fetched). Ohio and California customers' balances therefore go to those states:
    - Ohio: retail-customer credits after three years (ORC 169.02); report „before the first day of November of each year as of the preceding thirtieth day of June”, with notice to owners of $50 or more (ORC 169.03(D)–(E)).
    - California: property „unclaimed by the owner for more than three years after it became payable or distributable” (CCP § 1520(a)); report „before November 1 of each year as of June 30” (CCP § 1530(d)).
    - Neither state's business-to-business rules were checked.
- **Grade:** 2 settled — each state's text is plain; which state takes a given item is settled by the Supreme Court's priority rules
- **Rhea's default:**
  - Every open liability to a third party (outstanding check, customer credit, unpaid wages) carries a dormancy clock from the date the obligation arose, with the owner's state from its last known address.
  - On 1 July each year Rhea lists what will be presumed abandoned by 30 June, by state. It raises the due-diligence letters (in the 60–180-day window for Indiana) and the report task before 1 November (Indiana and California: before 1 November; Ohio: before the first day of November).
  - Remitting reverses the liability to "unclaimed property remitted". Nothing is ever written off to income.
  - Business-to-business credit memos are excluded in Indiana only.
- **Rhea today:** the clock and dated duties exist. No liability carries a dormancy start or an owner state. The Poland pack has no analogue (Polish law has no escheat of this kind — not researched).
  - In a first year nothing can be dormant yet (three years), so the first possible report is years away. Whether Indiana wants a "nothing to report" filing meanwhile was not checked.
  - What the first year owes is the dormancy start date on every open liability. Without it, the clock cannot be rebuilt later.
- **Where the work lives:** *pack data* (dormancy periods by state and property type, with warrants, and the start date now); *later* (the report itself, and *market code* for the NAUPA standard electronic file if the state portal does not accept a CSV — not researched)
- **Deadline:**
  - Indiana: before 1 November each year. 31 October 2027 is a Sunday; an electronic filing that day is still "before November 1", and Rhea's internal date is Friday 29 October 2027 (computed here).
  - Ohio and California: before 1 November.
  - Due diligence: 60–180 days earlier.
- **Watch:** Indiana Attorney General Unclaimed Property holder reporting guidelines (https://www.indianaunclaimed.gov/app/reporting-guidelines — returned 403 to automated access; watch manually); the Indiana Register for 10 IAC 1.5 rules (https://iar.iga.in.gov/latestArticle/10/1.5); Ohio Division of Unclaimed Funds; California State Controller holder pages (https://www.sco.ca.gov/upd_rptg_holders.html — 403)
- **Sources:** https://codes.findlaw.com/in/title-32-property/in-code-sect-32-34-1-5-1.html ; https://codes.findlaw.com/in/title-32-property/in-code-sect-32-34-1-5-4.html ; https://codes.findlaw.com/in/title-32-property/in-code-sect-32-34-1-5-20.html ; https://www.in.gov/sboa/files/Byrns_Indiana-Unclaimed_Clerks-2025.pdf ; https://codes.findlaw.com/oh/title-i-state-government/oh-rev-code-sect-169-02.html ; https://codes.findlaw.com/oh/title-i-state-government/oh-rev-code-sect-169-03.html ; https://codes.findlaw.com/ca/code-of-civil-procedure/ccp-sect-1520.html ; https://codes.findlaw.com/ca/code-of-civil-procedure/ccp-sect-1530.html

---

### PAY-26 — Through which IRS channel do we file 1099s now that FIRE is closing, and what must we apply for first?
- **Applies to:** everyone
- **The law:**
  - Pub. 1099 (2026): „Beginning tax year 2026/filing season 2027, the Information Returns Intake System (IRIS) will be the only intake system for information returns.” „FIRE will not be available for submissions (current, prior year, or corrections) once the system shuts down in 2026 for end of year.”
  - Pub. 1220 (Rev. 5-2026), QuickAlert of 7 May 2026: the IRS „will no longer accept new Information Returns (IR) Applications for Transmitter Control Codes (TCCs) beginning July 21, 2026”, and „IRIS will be the only information returns electronic filing system, including current year, prior year, or corrections, after January 1, 2027”.
  - IRIS has two doors.
    - The Taxpayer Portal (Pub. 5717, Rev. 9-2026) is „a no cost online method”. It allows keyed entry or a CSV upload („up to 250 forms per file”, IRS templates only, one form type and tax year per file), „up to 100 forms per submission”, participation in the Combined Federal/State Filing Program, corrections and extensions.
    - A2A (Pub. 5718): „The A2A channel uses REST (Representational State Transfer) APIs”, with XML returns. It requires an IRIS A2A TCC, an e-Services account, an API Client ID registered with a JSON Web Key, and passing communication testing in the IRIS Assurance Testing System.
  - Both doors need an IRIS TCC; „The typical TCC application will be processed within 45 business days” (Pub. 1099).
- **Grade:** 1 plain
- **Rhea's default:**
  - Year one: Rhea writes the IRS's IRIS Portal CSV for 1099-NEC and 1099-MISC, and the owner uploads it. There is no software certification to pass, and it fits both businesses' volumes.
  - The IRIS Portal TCC application is a dated duty now. 45 business days from 16 October 2026 is Tuesday 22 December 2026, counting federal holidays (computed here), so it must start this month for a January 2027 filing.
  - A2A (an API client in `markets/us`, like the KSeF client in `markets/pl`) is a second-year item, once Rhea files for several companies.
- **Rhea today:** the writer seam (`adapter.ProvideWriter`, `issued_document`) and a live statutory API client pattern (KSeF 2.0 with token auth and receipt evidence) exist. Nothing US.
- **Where the work lives:** *market code* (IRIS Portal CSV writer now; IRIS A2A client later)
- **Deadline:** apply for the IRIS TCC now (computed: ~45 business days before the 1 February 2027 filing)
- **Watch:**
  - IRS QuickAlerts for e-file (https://www.irs.gov/e-file-providers/subscribe-to-quickalerts), where FIRE's retirement was announced.
  - The IRIS page (https://www.irs.gov/iris).
  - Pub. 5717 and Pub. 5718 revision dates on https://www.irs.gov/pub/irs-pdf/p5717.pdf and https://www.irs.gov/pub/irs-pdf/p5718.pdf.
  - IRIS A2A schema releases (published with Pub. 5718).
- **Sources:** https://www.irs.gov/publications/p1099 ; https://www.irs.gov/pub/irs-pdf/p1220.pdf ; https://www.irs.gov/pub/irs-pdf/p5717.pdf ; https://www.irs.gov/pub/irs-pdf/p5718.pdf

---

### PAY-27 — We have only four freelancers. Must we e-file, or may we file paper?
- **Applies to:** everyone
- **The law:**
  - Reg. 301.6011-2(c)(4)(i) (T.D. 9972, for returns required on or after 1 January 2024): „all the information returns described in paragraphs (b)(1) and (2) of this section required to be filed during the calendar year are counted in the aggregate”. Paragraph (b)(1) covers the Form 1099 series and others; (b)(2) covers Form W-2.
  - Pub. 1099: „If you are required to file 10 or more information returns during the year, you must e-file. The 10-or-more requirement does not apply separately to each type of form.” Corrections follow the original: if the originals had to be e-filed, so must the corrections. Corrections do not count toward the 10.
  - Paper filed when e-filing was required is a failure to file under § 6721 (Pub. 1099, penalty list).
  - Form 8300 must be e-filed when 10 or more other information returns are required (PAY-17).
- **Grade:** 1 plain
- **Rhea's default:**
  - Corporation: its ten employees' W-2s, filed by the payroll provider under its EIN, already make 10, so every 1099 and any Form 8300 it files must be e-filed.
  - Sole proprietor: with no W-2s and fewer than 10 1099s it may file paper (Form 1096 plus scannable red Copy A). Rhea still defaults to the IRIS Portal, which is free, because paper forms cannot be printed by Rhea.
  - The count is a rule over the year's required returns, recomputed when a 1099 duty appears.
- **Rhea today:** missing.
- **Where the work lives:** *pack data*
- **Deadline:** none of its own
- **Watch:** Reg. 301.6011-2 via the eCFR versions API (https://www.ecfr.gov/api/versioner/v1/versions/title-26.json?section=301.6011-2); the Federal Register API for IRS e-filing rules
- **Sources:** https://www.ecfr.gov/api/versioner/v1/full/2026-10-01/title-26.xml?section=301.6011-2 ; https://www.irs.gov/publications/p1099

---

### PAY-28 — When are the 2026 information returns due, to the IRS and to the recipients?
- **Applies to:** everyone
- **The law:**
  - IRC § 6071(c): „Forms W–2 and W–3 and any returns or statements required by the Secretary to report nonemployee compensation” „shall be filed on or before January 31 of the year following the calendar year to which such returns relate.”
  - Recipient statements: § 6041A(e) and § 6050W(f), „on or before January 31”.
  - 1099-MISC (1099 instructions): to recipients by 31 January (15 February for boxes 8 and 10); to the IRS by 28 February on paper or 31 March electronically.
  - Form 945 and Form 8300 statements: 31 January.
  - § 7503: an act due on a Saturday, Sunday or legal holiday is timely on the next day that is none of those.
- **Grade:** 1 plain
- **Rhea's default:** the clock carries these dated duties for tax year 2026 (all dates computed here):
  - 1099-NEC to recipients and IRS: Monday 1 February 2027;
  - 1099-MISC to recipients: 1 February 2027 (boxes 8/10: Tuesday 16 February 2027);
  - 1099-MISC to IRS: 1 March 2027 on paper, Wednesday 31 March 2027 via IRIS;
  - 1099-INT and 1099-DIV to recipients: 1 February 2027;
  - Form 945: 1 February 2027 (10 February if all deposits were timely);
  - Form 8300 payer statements: 1 February 2027.
  - A preparation task fires on the first business day of January, Monday 4 January 2027 (Rhea's choice, not law).
- **Rhea today:** the clock and dated duties exist; no US calendar.
- **Where the work lives:** *pack data*
- **Deadline:** as listed
- **Watch:** the IRS tax calendar (https://www.irs.gov/businesses/small-businesses-self-employed/online-tax-calendar, not opened; Pub. 509 annual edition); OPM's federal holiday list for § 7503 (DC legal holidays, including DC Emancipation Day 16 April)
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6071 ; https://www.law.cornell.edu/uscode/text/26/7503 ; https://www.irs.gov/instructions/i1099mec ; https://www.irs.gov/instructions/i945

---

### PAY-29 — Do Indiana, Ohio or California want copies of our 1099s?
- **Applies to:** everyone (California and Ohio only when a payee or a payment is theirs)
- **The law:**
  - **Combined Federal/State Filing.** Through IRIS, IRS forwards 1099-NEC and 1099-MISC to participating states. California, Indiana and Ohio are all in the IRIS CF/SF table (Pub. 5717, Rev. 9-2026). „It is the filer's responsibility to contact the participating state(s) to verify their criteria”.
  - **Indiana.** „Information returns (such as 1099s) that do not report withholding of Indiana state or local income tax should not be submitted to the department.” Those with withholding go with Form WH-3, due „31 days after the end of the calendar year”, in „the IRS 1220 format”. Late penalty: „$10 for each late-filed W-2 and 1099 statement” (DOR Departmental Notice #5, September 2023).
  - **Ohio.** The TY2025 1099-NEC specifications (rev. November 2025):
    - files „must contain 1099-NEC information for each recipient from whom you withheld Ohio individual income tax or Ohio school district income tax”; others are not required;
    - upload through OH|TAX eServices in the Pub. 1220 layout („FIXED length of 750 positions”);
    - due „February 2, 2026” for TY2025.
  - **California.**
    - File if „The recipient is a California resident or part-year resident” or „The source of an information return transaction was in California”.
    - If you file under CF/SF „and the amounts you are reporting to IRS and us are the same, do not file with us”. Different amounts go directly to FTB, electronically through SWIFT when 10 or more.
    - The FTB guide adopts „$2,000 or more for tax year 2026” for 1099-NEC.
    - EDD Form DE 542: a business must report within „20 calendar days” any independent contractor „who provide[s] services for any of your business operations in California” once it has paid or contracted „$600 or more”. That is CUIC § 1088.8, not conformed to $2,000.
    - California 7% backup withholding: PAY-12.
- **Grade:** 2 settled — the state statutes were not read; the state revenue departments' published guidance (Indiana Departmental Notice #5, ODT specifications, FTB guidance) settles practice
- **Rhea's default:**
  - Federal 1099s go through IRIS with CF/SF ticked for any California payee. No separate Indiana or Ohio filing is made unless state tax was withheld (Rhea withholds none in year one).
  - A DE 542 duty arises only for a contractor working on California operations, which neither business has as described.
  - A Pub. 1220-format writer for state filing is not needed in year one. This is notable: the states still take the fixed-length format the IRS is retiring.
- **Rhea today:** missing.
- **Where the work lives:** *pack data*; *later* (*market code*: Pub. 1220 750-byte writer for Indiana/Ohio/California direct filing)
- **Deadline:**
  - Indiana (with withholding): 1 February 2027 (computed here).
  - Ohio (with withholding): expected 1 February 2027 (computed here; the TY2026 Ohio spec is not yet published).
  - California direct filing: FTB guide lists 28 February paper and 31 March electronic, with recipients 31 January.
  - DE 542: 20 days from payment or contract.
- **Watch:**
  - Indiana DOR withholding forms and specifications page (https://www.in.gov/dor/tax-forms/withholding-tax-forms/). Its "2025 and Prior Tax Years 1220" guide name suggests a new TY2026 layout; watch for it.
  - Ohio ODT employer withholding specifications, reissued each tax year (dam.assets.ohio.gov, path in Sources).
  - FTB "Guidance for reporting information returns" ("Last updated: 04/03/2026") and FTB Tax News monthly (https://www.ftb.ca.gov/about-ftb/newsroom/tax-news/).
  - EDD independent contractor page.
  - Pub. 5717 CF/SF state table.
- **Sources:** https://www.irs.gov/pub/irs-pdf/p5717.pdf ; https://www.in.gov/dor/files/dn05.pdf ; https://dam.assets.ohio.gov/image/upload/tax.ohio.gov/employer_withholding/2025-2026%20NYP%20Updates/2025_1099NEC_Specs_V1.pdf ; https://www.ftb.ca.gov/file/business/information-returns.html ; https://www.ftb.ca.gov/file/guide-to-information-returns-filed-with-california.html ; https://edd.ca.gov/en/payroll_taxes/Independent_Contractor_Reporting/

---

### PAY-30 — We filed a 1099 with a wrong amount or TIN. What does it cost, and how do we correct it?
- **Applies to:** everyone
- **The law:**
  - §§ 6721 (returns) and 6722 (payee statements) are separate penalties per form. Amounts for returns due and statements furnished in 2027 (Rev. Proc. 2025-32 §§ 3.57–3.58). Both businesses are "small" (average gross receipts of $5,000,000 or less):
    - $60 per return if corrected within 30 days (annual cap $244,500);
    - $130 if corrected by 1 August (cap $698,500);
    - $340 after that or never (cap $1,397,000);
    - intentional disregard: the greater of $690 or 10% of the amounts, no cap.
    - Larger filers' caps are $698,500, $2,095,500 and $4,191,500.
  - The Rev. Proc. tables print the middle tier as "August 1, 2026" for returns filed in 2027. The statute says „on or before August 1” of the filing year (§ 6721(b)(2)), so read 1 August 2027.
  - De minimis safe harbor (§ 6721(c)(3)): no correction is needed if „no single amount in error differs from the correct amount by more than $100” and „no single amount reported for tax withheld … differs … by more than $25”. The payee may elect out (§ 6722(c)(3)(B)).
  - Corrections filed by 1 August for the greater of 10 returns or ½ of 1% are spared (Pub. 1099).
  - Errors in a TIN, a payee surname or any money amount are never "inconsequential" (Pub. 1099).
  - Indiana adds $10 per late statement (PAY-29).
- **Grade:** 1 plain
- **Rhea's default:**
  - A 1099 is an issued document derived from payments. A later fact that changes it (a reclassified payment, a corrected TIN) makes Rhea draft a corrected return through the same IRIS channel, behind the gate. The original stays in the log with its receipt.
  - Rhea drafts corrections inside the $100/$25 safe harbor too, unless the owner decides otherwise, because the payee may elect out.
  - The penalty table is dated master data.
- **Rhea today:** the pattern exists in Poland: KSeF rejection becomes a task, corrections are new documents, receipts are kept. Missing: everything US.
- **Where the work lives:** *pack data* (penalty table, correction rules); *market code* (IRIS correction records)
- **Deadline:** correct within 30 days of the due date (by about 3 March 2027 for 1099-NEC, computed here), and in any case by 1 August 2027. That is a Sunday; Rhea's internal date is Friday 30 July 2027, because whether § 7503 moves a penalty-tier date is not settled by any source reached.
- **Watch:** the annual inflation Rev. Proc. (IRS GuideWire; https://www.irs.gov/pub/irs-drop/); the IRS penalty page (https://www.irs.gov/payments/information-return-penalties, "Page Last Reviewed or Updated: 15-Oct-2026", table currently ends at returns due in 2026)
- **Sources:** https://www.irs.gov/pub/irs-drop/rp-25-32.pdf ; https://www.law.cornell.edu/uscode/text/26/6721 ; https://www.irs.gov/publications/p1099 ; https://www.irs.gov/payments/information-return-penalties

---

**Sources this section could not reach:**
- The eCFR web pages and the Indiana, Ohio and California legislature sites refused automated access. The eCFR versioner API and FindLaw copies were used instead; FindLaw marks its text "current as of January 01, 2026".
- The Indiana Unclaimed Property site and the California State Controller's holder pages returned 403.
- The Nacha Operating Rules are not public; nacha.org summaries were used.
- FASB ASC is behind a login.
- The Fedwire implementation date, the TIN-type prefix rules, the 1099-K monthly boxes, W-8BEN-E validity and IC 32-34-1.5-23 were not read in primary text, and each is marked where it appears.


## Sole proprietor — books and income tax

Scope: Maple Lane Goods, a single-member LLC in Indianapolis (Marion County), Indiana, owned by one individual, disregarded for federal tax, cash method, calendar year, first year of business 2026, about $120,000 gross receipts from handmade goods (own web shop, one marketplace, craft fairs) and some design services; card processor; a laser cutter and a laptop; a home workshop; the family car to fairs; a few freelance contractors; no employees. Law as it stands on 16 October 2026, tax year 2026. Primary texts read in the relevant part: the Internal Revenue Code as currently codified with the amendments of Public Law 119-21 (read on law.cornell.edu, whose text and amendment notes carry P.L. 119-21); Treasury Regulations from the eCFR (point-in-time API, 1 October 2026); Rev. Proc. 2025-32 (2026 inflation adjustments, read in full for the items used); Notice 2026-10 and Announcement 2026-11 (2026 mileage rates); Rev. Proc. 2026-15 (2026 car depreciation caps); Notice 2026-11 (bonus depreciation after P.L. 119-21); Notice 2025-67 (2026 retirement limits); Notice 2015-82 (de minimis); Rev. Proc. 2013-13 (simplified home office); Rev. Proc. 2019-46 (mileage rules); SSA's notice "Cost-of-Living Increase and Other Determinations for 2026" (Federal Register, 3 November 2025, FR Doc. 2025-19763); the IRS's 2026 draft Schedule C (created 15 May 2026), draft Schedule C instructions (dated 8 October 2026), draft Schedule SE, Form 8995, Form 8829, Form 4562-B and Schedule 1 for 2026; the final 2026 Form 1040-ES; Instructions for Form 1099-K and Forms 1099-MISC/NEC (Rev. December 2026); Pubs. 334, 463, 560, 583, 587 and 946 (2025 editions, the latest final ones); Indiana DOR Departmental Notice #1 (effective 1 January 2026), Information Bulletins #3 and #118 (May 2026), the add-backs page and the New & Small Business Owners guide (July 2026); DLGF's memos of 23 May 2025 (assessment) and 17 July 2026 (local income tax), DLGF's 2026 personal property training deck, and Form 103-Long for 2026; the Indiana Secretary of State's business entity report page and its regulatory analysis LSA 25-155; FinCEN's BOI page. The Indiana Code itself could not be read: iga.in.gov serves the code only to a browser and refused every automated fetch, forms.in.gov and law.justia.com returned 403. Indiana statutes are therefore cited through the DOR and DLGF documents that quote them, and each entry says so. ssa.gov also returned 403; the wage base is from SSA's Federal Register notice.

Abbreviations: "IRC" = Internal Revenue Code of 1986, 26 U.S.C.; "Reg." = Treasury Regulations, 26 CFR; "P.L. 119-21" = the act of 4 July 2025 known as the One Big Beautiful Bill Act ("OBBBA" in IRS documents); "Rev. Proc. 2025-32" = the 2026 inflation adjustments; "Sch. C instr. 2026 (draft)" = the IRS draft Instructions for Schedule C (Form 1040) 2026, dated 8 October 2026; "IC" = Indiana Code; "DOR" = Indiana Department of Revenue; "DLGF" = Indiana Department of Local Government Finance; "SOS" = Indiana Secretary of State; "IB" = DOR Income Tax Information Bulletin; "SMLLC" = single-member limited liability company; "NIMS" = inventory treated as non-incidental materials and supplies (Reg. §1.471-1(b)(4)); "AFS" = applicable financial statement.

What changed for 2026–2027, in one place:
- **§179** (P.L. 119-21 §70306): $2,500,000 / $4,000,000 for 2025, indexed; **$2,560,000**, phase-out from **$4,090,000**, SUV cap **$32,000** for 2026 (Rev. Proc. 2025-32 §4.24).
- **Bonus depreciation** (P.L. 119-21 §70301): **100% permanently** for property acquired after 19 January 2025; Notice 2026-11 (14 January 2026) is the interim guidance.
- **§199A** (P.L. 119-21 §70105): no longer expires; phase-in widened to $75,000 ($150,000 joint); a **$400 minimum deduction** for at least $1,000 of active QBI from 2026. 2026 threshold **$201,750** ($403,500 joint).
- **Mileage**: **72.5¢** a mile from 1 January to 30 June 2026, **76¢** from 1 July 2026 (Announcement 2026-11, citing fuel prices). The first mid-year change in this review's sources.
- **Information returns**: 1099-NEC/MISC threshold **$2,000** for payments after 2025 (P.L. 119-21 §70433), indexed from 2027 filings; 1099-K for third-party networks back to **$20,000 and 200 transactions**, retroactively as if part of the 2021 act (§70432).
- **Social Security wage base** **$184,500** for 2026 (SSA, FR Doc. 2025-19763). The 2027 figure was not yet published in the Federal Register on 16 October 2026.
- **Schedule C 2026**: new line **16b vehicle loan interest** (16c other interest); amortization moves to new **Form 4562-B** (December 2026 revision); employer-convenience meals no longer deductible after 2025 (§274(o)).
- **Indiana**: individual rate **2.95%** for 2026, **2.90%** for 2027 (DOR); Marion County **2.02%** (Departmental Notice #1, 1 January 2026); business personal property exemption **$2,000,000** from the 1 January 2026 assessment date (IC 6-1.1-3-7.2, via DLGF); SEA 243-2026 extended Indiana's decoupling from bonus depreciation to §168(n) property (IB 118, retroactive to 4 July 2025); HEA 1210-2026 delays the local income tax overhaul by a year, to collections from 2029 (DLGF, 17 July 2026).
- **BOI**: FinCEN's final rule (Federal Register, 14 August 2026) exempts U.S. companies from beneficial ownership reporting.

What Rhea has today for this area:
- **company** is a Polish type in practice: the shared seed has v1 (name, kind, vat_id, country) and the Poland pack carries it to v7 with `legal_form` (natural_person | legal_entity), `books` (full | kpir | ryczalt), `income_tax` (cit | skala | liniowy | ryczalt) and `form_from`. The pattern (the company declares its form with a date and rules branch on it) is the right one. The enums are Polish.
- **sales_invoice** and **purchase_invoice** (testdata/seed/finance_v8.json) require `vat`, a `vat_rate` and, for sales, a `buyer` that is a company. A craft-fair cash sale or a marketplace order cannot be recorded as one.
- **posting**, **account**, **period_lock**, the warehouse types and received/issued documents are market-neutral. The warehouse's `goods_receipt` records quantity, not cost.
- **fixed_asset**, **vehicle**, **mileage_entry**, **tax_declaration**, **stocktake**, **task** and **task_schedule** live in `packs/pl/pack.json`, not in shared ground, and carry Polish fields (KŚT, kilometres).
- The executor runs **only the newest active version of each rule**, and `effective_from` is a lower bound only (`internal/exec/exec.go` matchRule; `internal/store/store.go` ActiveRules).
- Field types are string, int, decimal, date, money, enum and ref (`internal/core/core.go`). No field can be marked as sensitive (an SSN).

---

### SP-01 — How do we form the LLC in Indiana, what does it cost, and what keeps it alive?
- **Applies to:** sole proprietor (the corporation files the same biennial report)
- **The law:**
  - Formation is by Articles of Organization filed with the SOS, online through INBiz or on paper (State Form 49459). The SOS's regulatory analysis for its fee rule (LSA Document #25-155) lists a "Domestic LLC Articles of Organization $20.00" Enhanced Access fee charged on online filings in addition to the statutory fee, and puts Indiana's "LLC Registration Cost" at "$96" in its regional comparison. The commonly cited totals, **$95 online** ($75 statutory electronic fee + $20) and **$100 on paper**, come from secondary sources and a search-engine copy of the form (secondary — verify). I could not open IC 23-0.5-9 (fees) or the form itself.
  - Keeping the entity alive, SOS: for-profit entities "file every other year (biennial) in the anniversary month in which the entity was originally created". The fee is "For Profit Online: $31" and "For Profit by Mail: $50". If the report is not filed, the office "will administratively dissolve or revoke the business entity". Reinstatement is possible within 5 years.
  - Beneficial ownership: FinCEN's final rule of 14 August 2026 (FR Doc. 2026-16576) makes permanent the exemption of the interim final rule of 26 March 2025: "U.S. companies are exempt from BOI reporting requirements and therefore, are no longer required to file BOI reports."
- **Grade:** 1 plain (the formation fee amount itself is unverified)
- **Rhea's default:** The LLC is a dated fact: state of formation, formation date, SOS business ID, registered agent and principal office. A dated duty opens a task for the Business Entity Report in the anniversary month every second year. For a 2026 formation that is the anniversary month of 2028 (computed here from the SOS rule; the statute's wording on the first report was not read). Rhea records no BOI duty for a U.S. LLC.
- **Rhea today:** company v7 has `legal_form` and `owner`, but no state, formation date or registry ID. The clock can fire dated duties, but `task` and `task_schedule` are Poland-pack types.
- **Where the work lives:** *pack data* (company fields, the biennial duty); *shared ground* (move `task` and `task_schedule` from `packs/pl` to `packs/common`, so that every market can schedule a duty without copying Polish types)
- **Deadline:** Business Entity Report in the anniversary month, every second year
- **Watch:** SOS fees are now set by rule (HEA 1623-2023), so a change appears as an LSA document in the Indiana Register (https://iar.iga.in.gov) and on the SOS entity-report page (https://www.in.gov/sos/business/business-entity-reports/). Changes to the Business Organization Code (IC 23-0.5) come from the General Assembly (https://iga.in.gov) and usually take effect 1 July. BOI: https://www.fincen.gov/boi and FinCEN documents in the Federal Register (https://www.federalregister.gov/agencies/financial-crimes-enforcement-network).
- **Sources:** https://www.in.gov/sos/business/business-entity-reports/ ; https://www.in.gov/sos/business/files/Regulatory-Analysis-Business-Entity-LSA-25-155-OMB-2024-73R.pdf ; https://www.fincen.gov/boi ; https://www.federalregister.gov/documents/2026/08/14/2026-16576/beneficial-ownership-information-reporting-requirement-revision

---

### SP-02 — The IRS ignores our LLC: what does "disregarded" mean for the books?
- **Applies to:** sole proprietor
- **The law:**
  - Reg. §301.7701-3(b)(1): "unless the entity elects otherwise, a domestic eligible entity is— … (ii) Disregarded as an entity separate from its owner if it has a single owner."
  - Reg. §301.7701-2(c)(2)(i): "a business entity that has a single owner and is not a corporation under paragraph (b) of this section is disregarded as an entity separate from its owner."
  - Exceptions: the entity "is treated as a corporation with respect to taxes imposed under Subtitle C—Employment Taxes" (§301.7701-2(c)(2)(iv)(B)) and is separate for certain excise taxes (c)(2)(v).
  - Sch. C instr. 2026 (draft): "If you are the sole member of a domestic LLC, file Schedule C … unless you have elected to treat the domestic LLC as a corporation."
  - Indiana: an LLC files "IT-40, IT-40PNR or IT-65" "depending on how the LLC elects to be taxed" (DOR, New & Small Business Owners guide, July 2026).
- **Grade:** 1 plain
- **Rhea's default:**
  - The company declares two dated facts: its state-law form (LLC) and its federal tax classification (disregarded, from the formation date). Every income-tax rule branches on the classification, never on the form.
  - The LLC's ledger is the owner's Schedule C business. Money the owner puts in or takes out is equity (contributions, draws), never income or expense, and there is no owner's salary.
- **Rhea today:**
  - Polish `legal_form` (natural_person | legal_entity) cannot say "a legal entity under state law that is the owner for income tax". `income_tax` and `books` are Polish enums.
  - The Polish design (declared form + `form_from` + rules gated on it) carries over unchanged.
- **Where the work lives:** *shared ground*. The company type's market enums (`legal_form`, `books`, `income_tax`, `vat_id`) need a market-neutral shape: a dated tax-classification fact per jurisdiction. And object types version in one line per name: the pack loader (`internal/pack/pack.go`) stores `company` vN whichever pack ships it, so a US `company` version and the Polish v7 cannot both be current in one database. *pack data* for the US classification values.
- **Deadline:** none (the default applies from formation)
- **Watch:** the check-the-box regulations change only through Treasury decisions published in the Federal Register (https://www.federalregister.gov/agencies/internal-revenue-service). The IRS SMLLC page (https://www.irs.gov/businesses/small-businesses-self-employed/single-member-limited-liability-companies) reflects them.
- **Sources:** https://www.ecfr.gov/current/title-26/chapter-I/subchapter-F/part-301/section-301.7701-3 ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-F/part-301/section-301.7701-2 ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf ; https://www.in.gov/dor/files/new-small-business-handbook.pdf

---

### SP-03 — Do we need an EIN, and whose number goes on the W-9 we give the marketplace and clients?
- **Applies to:** sole proprietor
- **The law:**
  - Reg. §301.6109-1(h)(2)(i): a disregarded single-owner entity "must use its owner's taxpayer identifying number (TIN) for federal tax purposes."
  - IRS, Single Member LLCs: one "that does not have employees and does not have an excise tax liability does not need an EIN. It should use the name and TIN of the single member owner for federal tax purposes. However, if a single-member LLC … needs an EIN to open a bank account or if state tax law requires … the LLC can apply for and obtain an EIN."
  - Form W-9 instructions: "information for disregarded entities is reported as the owner's name on line 1, and the disregarded entity's name is entered on line 2."
  - Sch. C instr. 2026 (draft), line D: enter the LLC's EIN only if issued "for a qualified retirement plan; to file employment, excise, alcohol, tobacco, or firearms returns; or as a payer of gambling winnings. If you do not have such an EIN, leave line D blank."
- **Grade:** 1 plain
- **Rhea's default:** The owner's TIN is the business's income-tax identifier. An LLC EIN, if the owner gets one for the bank or for Indiana sales tax registration, is a second identifier. It goes on no W-9 and on Schedule C line D only for one of the listed uses (a solo 401(k) would be one, SP-28).
- **Rhea today:** `company.vat_id` is a string read with NIP semantics. Nothing says that the TIN belongs to the owner (a person), and no field can be marked sensitive, so an SSN would show wherever the field is shown.
- **Where the work lives:** *pack data* (identifier fields with their role); *shared ground* (a sensitive field kind that views, the briefing and the network never show in full; no such marker exists among the kernel's field types)
- **Deadline:** before the first W-9 is requested (the marketplace and the processor ask at onboarding)
- **Watch:** Form W-9 revisions (https://www.irs.gov/forms-pubs/about-form-w-9); the IRS SMLLC page above.
- **Sources:** https://www.ecfr.gov/current/title-26/chapter-I/subchapter-F/part-301/section-301.6109-1 ; https://www.irs.gov/businesses/small-businesses-self-employed/single-member-limited-liability-companies ; https://www.irs.gov/pub/irs-pdf/iw9.pdf ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf

---

### SP-04 — Which tax year do the books close on?
- **Applies to:** sole proprietor
- **The law:** IRC §441(b)(1): the taxable year is "the taxpayer's annual accounting period, if it is a calendar year or a fiscal year". §441(g): "the taxpayer's taxable year shall be the calendar year if— (1) the taxpayer keeps no books; (2) the taxpayer does not have an annual accounting period". A disregarded entity has no year of its own; Schedule C is part of the owner's Form 1040, and §6072(a) makes calendar-year returns due "on or before the 15th day of April following the close of the calendar year".
- **Grade:** 1 plain
- **Rhea's default:** The calendar year, fixed. Monthly period locks roll up to it. No fiscal-year option is shipped for a disregarded LLC owned by an individual.
- **Rhea today:** `period_lock` and `close_period` (packs/common/periods.json) are market-neutral and serve as they are.
- **Where the work lives:** *pack data*
- **Deadline:** the year ends 31 December; the return is due 15 April 2027
- **Watch:** stable statute; nothing periodic.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/441 ; https://www.law.cornell.edu/uscode/text/26/6072

---

### SP-05 — May we use the cash method, and how is the method chosen?
- **Applies to:** sole proprietor
- **The law:**
  - IRC §446(c): "a taxpayer may compute taxable income under any of the following methods of accounting— (1) the cash receipts and disbursements method; (2) an accrual method".
  - Reg. §1.446-1(c)(1)(i): "all items which constitute gross income (whether in the form of cash, property, or services) are to be included for the taxable year in which actually or constructively received. Expenditures are to be deducted for the taxable year in which actually made."
  - Reg. §1.446-1(d)(1): "The method first used in accounting for business income and deductions in connection with each trade or business, as evidenced in the taxpayer's income tax return in which such income or deductions are first reported, must be consistently followed thereafter."
  - §448(a) bars the cash method only for a C corporation, a partnership with a C corporation partner, or a tax shelter, so it does not reach an individual's Schedule C.
  - Sch. C instr. 2026 (draft), line F: "Unless you are a small business taxpayer … you must use an accrual method for sales and purchases of inventory items." A change later needs Form 3115.
- **Grade:** 1 plain
- **Rhea's default:** Cash method, declared with its start date, and fixed by the first return (line F). Income dates are receipt dates (SP-18) and expense dates are payment dates. A purchase on a credit card counts as paid when charged (Rev. Rul. 78-39, not fetched here — verify). Inventory is the exception, handled in SP-10.
- **Rhea today:** the matching engine's settlements already carry payment dates, so cash-basis dating is a reading of facts Rhea has. Postings, however, follow invoice dates in the Polish rules.
- **Where the work lives:** *pack data* (rules that date income and expense by settlement)
- **Deadline:** chosen on the 2026 return, due 15 April 2027
- **Watch:** accounting-method changes appear as revenue procedures in the Internal Revenue Bulletin (https://www.irs.gov/irb), mainly the annual list of automatic changes.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/446 ; https://www.law.cornell.edu/uscode/text/26/448 ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.446-1 ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf

---

### SP-06 — Are we a "small business taxpayer", and what does that buy us?
- **Applies to:** everyone (the corporation needs the same test to use the cash method at all)
- **The law:**
  - IRC §448(c)(1): the test is met "if the average annual gross receipts of such entity for the 3-taxable-year period ending with the taxable year which precedes such taxable year does not exceed $25,000,000", indexed (§448(c)(4)). Rev. Proc. 2025-32 §4.30 sets it for taxable years beginning in 2026 at "$32,000,000".
  - §448(c)(3)(A): a business not in existence for the whole period is tested "on the basis of the period during which such entity (or trade or business) was in existence". (B): a short year is annualized. (C): gross receipts are "reduced by returns and allowances".
  - For an individual, §471(c)(3) and §263A(i)(2) apply the test "as if each trade or business of such taxpayer were a corporation or partnership".
  - The tax-shelter exclusion (§448(a)(3), (d)(3)) reaches a "syndicate": an entity other than a C corporation "if more than 35 percent of the losses … are allocable to limited partners or limited entrepreneurs" (§1256(e)(3)(B)). An owner who runs the business day to day actively participates in management, which §1256(e)(3)(C) treats as not holding as a limited entrepreneur (that reading of (C) is mine — verify).
- **Grade:** 1 plain. That a first year with no prior year passes: the statute's "period during which … was in existence" is empty, and I found no contrary authority.
- **Rhea's default:** Compute the test each year from the books' gross receipts (Schedule C line 1 less line 2), averaged over the years in existence. Record the result as a dated fact that the inventory (SP-10) and capitalization rules read. At $120,000 the test passes with a margin of more than two orders of magnitude, so the fact is near-constant. The test is still recorded, because three provisions depend on it.
- **Rhea today:** missing. The formula language can fold the year's receipts; nothing records the result.
- **Where the work lives:** *pack data*
- **Deadline:** none of its own (tested for each tax year)
- **Watch:** the amount is in the annual inflation revenue procedure, published each October–November in the Internal Revenue Bulletin and announced in the IRS newsroom (https://www.irs.gov/newsroom; the 2026 one is https://www.irs.gov/pub/irs-drop/rp-25-32.pdf).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/448 ; https://www.law.cornell.edu/uscode/text/26/471 ; https://www.law.cornell.edu/uscode/text/26/263A ; https://www.law.cornell.edu/uscode/text/26/1256 ; https://www.irs.gov/pub/irs-drop/rp-25-32.pdf

---

### SP-07 — Is there a required layout for the books, like the Polish KPiR?
- **Applies to:** sole proprietor
- **The law:**
  - IRC §6001: "Every person liable for any tax imposed by this title … shall keep such records … as the Secretary may from time to time prescribe."
  - Reg. §1.6001-1(a): "such permanent books of account or records, including inventories, as are sufficient to establish the amount of gross income, deductions, credits, or other matters required to be shown by such person in any return".
  - Pub. 583: "Except in a few cases, the law does not require any specific kind of records." Single-entry or double-entry is the owner's choice.
  - The return is the fixed form. The 2026 draft Schedule C asks for Part I income (lines 1–7), Part II expenses (lines 8–27b, home on line 30), Part III cost of goods sold (lines 33–42), Part IV vehicle (43–47b) and Part V other expenses listed by type (48). For 2026 it adds line 16b "Vehicle loan" interest and moves line 16b "Other" to 16c. Amortization moves to the new Form 4562-B.
  - Some items the owner pays are not Schedule C lines: half of the SE tax (Schedule 1 line 15), SEP and other self-employed plans (line 16), self-employed health insurance (line 17), estimated tax payments, and the owner's draws.
- **Grade:** 1 plain
- **Rhea's default:**
  - No statutory book is mandated, unlike the KPiR, whose 19 columns are fixed by regulation (Polish review KPIR-01). The fixed artifact is the return, so the US counterpart of the KPiR is a ledger whose expense and income accounts each map to one Schedule C line, plus the Part III, IV and V schedules and the Form 4562 and 8829 schedules.
  - The line map is dated by tax year, because lines move (16b/16c in 2026). Owner-level items (half SE tax, health insurance, retirement, estimated payments) are equity or owner accounts, outside Schedule C.
  - Meals are booked at full cost; the 50% limit is applied in the Schedule C view (line 24b is "Deductible meals").
- **Rhea today:** `account` and `posting` are generic double entry, and a ViewDef can sum postings by a mapped line. The Poland pack posts to the Polish model chart (pl-post-sales-invoice and pl-post-purchase-invoice). There is no US chart and no Schedule C view.
- **Where the work lives:** *pack data* (a US chart keyed to Schedule C lines with tax-year validity; a Schedule C ViewDef)
- **Deadline:** from the first transaction
- **Watch:** IRS draft forms (https://www.irs.gov/draft-tax-forms, "IRS.gov/DraftForms"). The draft Schedule C appears each spring–summer and the draft instructions in the autumn (the 2026 draft instructions are dated 8 October 2026). Line changes show in the instructions' "What's New".
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6001 ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.6001-1 ; https://www.irs.gov/pub/irs-pdf/p583.pdf ; https://www.irs.gov/pub/irs-dft/f1040sc--dft.pdf ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf ; https://www.irs.gov/pub/irs-dft/f1040s1--dft.pdf

---

### SP-08 — We spent money before the shop opened (and paid the LLC's filing fee): when is it deductible?
- **Applies to:** sole proprietor (the corporation also has §248 organizational costs; a disregarded LLC does not)
- **The law:**
  - IRC §195(b)(1): if elected, a deduction "for the taxable year in which the active trade or business begins in an amount equal to the lesser of— (i) the amount of start-up expenditures … or (ii) $5,000, reduced (but not below zero) by the amount by which such start-up expenditures exceed $50,000", and "the remainder … ratably over the 180-month period beginning with the month in which the active trade or business begins".
  - §195(c)(1): start-up expenditures are amounts for "investigating the creation or acquisition of an active trade or business, or … creating an active trade or business" that would be deductible if paid in an existing business.
  - Sch. C instr. 2026 (draft): "For amortization that begins in 2026, complete and attach Form 4562-B." The new form is a December 2026 revision.
  - Equipment bought before opening is not a start-up expenditure. It is depreciated from the date placed in service (SP-12).
- **Grade:** 1 plain for §195. Whether a disregarded LLC's state filing fee is a §195 start-up cost or an ordinary §162 business license is not settled in any text I found: §248 and §709 do not apply to a disregarded entity. Treating it as a start-up cost if paid before the business begins, and as a line 23 license afterwards, is my reading (verify with a practitioner).
- **Rhea's default:** The owner declares the date the active business began. Costs dated before it that would otherwise be deductible collect as start-up expenditures, take the up to $5,000 deduction on the first return, and amortize the rest over 180 months. The opening date is a fact the owner states, not a date Rhea infers.
- **Rhea today:** missing. There is no business-start fact (the Polish CEIDG start date is not modelled either) and no amortization schedule.
- **Where the work lives:** *pack data* (start-up classification rule, amortization schedule); *market code* later if Rhea writes Form 4562-B
- **Deadline:** election on the timely filed 2026 return
- **Watch:** §195 amounts are not indexed. The Form 4562-B instructions are at https://www.irs.gov/draft-tax-forms until final.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/195 ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf ; https://www.irs.gov/pub/irs-dft/f4562b--dft.pdf

---

### SP-09 — What records must we keep, in what form, and for how long?
- **Applies to:** everyone
- **The law:**
  - IRC §6001 and Reg. §1.6001-1(a) (SP-07). Reg. §1.6001-1(e): records "shall be retained so long as the contents thereof may become material in the administration of any internal revenue law."
  - Pub. 583, Table 3 (period of limitations): 3 years in general; 6 years if income "more than 25% of the gross income shown on the return" is left out; 7 years for a bad-debt or worthless-securities claim; not limited for a fraudulent or unfiled return.
  - Computerized records (Pub. 583, from Rev. Proc. 98-25): the records "must reconcile with your books and return" and must "identify the underlying source documents". The system's documentation must show "Controls used to prevent the unauthorized addition, alteration, or deletion of retained records" and the "Charts of accounts". Electronic storage of paper follows Rev. Proc. 97-22 (cited in Pub. 583; not fetched).
  - Travel, gifts and listed property have their own substantiation rule (§274(d), SP-17).
- **Grade:** 1 plain
- **Rhea's default:** Every source document (receipt, marketplace statement, processor report, 1099, contract) is a received document with its hash, linked to the facts it supports. Nothing is deleted. The retention period is a dated fact per tax year: 3 years from filing, extended to 6 or 7 where those rules apply.
- **Rhea today:** largely right. The append-only log, versioned rules and blob store with hashes are close to the Rev. Proc. 98-25 controls, and the received-document door with a required source exists (documents pack v3). Missing: a retention fact per year and an export of the "complete description" of the system the procedure asks for.
- **Where the work lives:** *pack data* (retention rule); *shared ground* (a system-description export that every market's auditor can ask for; Poland's JPK and the US machine-sensible rules want the same thing)
- **Deadline:** retention runs from the filing date of each return
- **Watch:** Rev. Proc. 98-25 and 97-22 are long-standing; changes would come as revenue procedures in the Internal Revenue Bulletin (https://www.irs.gov/irb). Pub. 583 is revised irregularly (current: Rev. December 2024).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6001 ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.6001-1 ; https://www.irs.gov/pub/irs-pdf/p583.pdf

---

### SP-10 — We make goods from materials: must we keep inventory, and when do material costs become deductible?
- **Applies to:** everyone with goods for sale (the corporation passes the same test)
- **The law:**
  - IRC §471(c)(1): a taxpayer meeting the §448(c) test may skip §471(a), and its method "shall not be treated as failing to clearly reflect income if such method either— (i) treats inventory as non-incidental materials and supplies, or (ii) conforms to … an applicable financial statement … or, if the taxpayer does not have any applicable financial statement …, the books and records of the taxpayer prepared in accordance with the taxpayer's accounting procedures."
  - §263A(i)(1): for such a taxpayer "this section shall not apply" (no uniform capitalization).
  - NIMS, Reg. §1.471-1(b)(4)(i): costs "are recovered through cost of goods sold only in the taxable year in which the inventory is used or consumed in the taxpayer's business, or in the taxable year in which the taxpayer pays for or incurs the cost of the inventory, whichever is later", and inventory "is used or consumed … in the taxable year in which the taxpayer provides the inventory to its customer."
    - (b)(4)(ii): identification by "a specific identification method, a first-in, first-out (FIFO) method, or an average cost method"; includible costs are "the direct material costs of the property produced".
    - NIMS inventory costs are not eligible for the de minimis safe harbor (b)(4)(i).
  - Books-and-records method, Reg. §1.471-1(b)(6)(i): the method "used for inventory in the taxpayer's books and records that properly reflect its business activities for non-tax purposes". Example 1 of (b)(6)(iii): a taxpayer that takes a year-end physical count and uses it in its books "must use the physical inventory count".
  - Sch. C instr. 2026 (draft): a small business taxpayer "can choose not to keep an inventory, but you must still use a method of accounting for inventory that clearly reflects income".
- **Grade:** 2 settled. The statute does not say when NIMS costs are deducted; Reg. §1.471-1(b)(4) (T.D. 9942, 86 FR 254, 5 January 2021) settles it: when the goods reach the customer, not when the materials are bought.
- **Rhea's default:**
  - NIMS, with FIFO identification of direct materials. A material purchase goes into stock at cost. Its cost moves to cost of goods sold (line 4, Part III) when the finished piece is sold, or when it is paid, whichever is later.
  - Labour (the owner's own time is never a cost), the workshop, tools and overhead are deducted when paid, because NIMS covers direct materials only.
  - The books-and-records method is offered only if the owner wants Rhea's inventory figures to drive the tax figures, and then any year-end count Rhea records binds it (SP-29).
- **Rhea today:** the warehouse has `item`, `goods_receipt`, `goods_issue` and `stock_movement`, all with quantity only. NIMS needs a cost per receipt and FIFO layers relieved on sale. The Poland pack's `pz_line` has a `value`, but it is Polish and outside shared ground.
- **Where the work lives:** *shared ground* (cost layers on the warehouse: a valued receipt and a FIFO or average relief on issue; any market with inventory needs them); *pack data* (the NIMS rule, Part III view)
- **Deadline:** method adopted on the first return (Schedule C line 33 and Part III)
- **Watch:** inventory regulations change by Treasury decision in the Federal Register (https://www.federalregister.gov/agencies/internal-revenue-service). The §448(c) amount is indexed in the annual revenue procedure (SP-06).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/471 ; https://www.law.cornell.edu/uscode/text/26/263A ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.471-1 ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf

---

### SP-11 — Can we simply expense small equipment and tools (the de minimis safe harbor)?
- **Applies to:** everyone
- **The law:**
  - Reg. §1.263(a)-1(f)(1)(ii): a taxpayer without an AFS may deduct an amount paid for tangible property if "(B) The taxpayer has at the beginning of the taxable year accounting procedures treating as an expense for non-tax purposes— (1) Amounts paid for property costing less than a specified dollar amount; or (2) Amounts paid for property with an economic useful life … of 12 months or less; (C) The taxpayer treats the amount paid for the property as an expense on its books and records in accordance with these accounting procedures; and (D) The amount paid for the property does not exceed $500 per invoice (or per item as substantiated by the invoice) or other amount as identified in published guidance".
  - Notice 2015-82: the limit "for a taxpayer without an AFS is increased from $500 to $2,500", for taxable years beginning on or after 1 January 2016.
  - (f)(5): the election is annual, "by attaching a statement to the taxpayer's timely filed original Federal tax return", titled "Section 1.263(a)-1(f) de minimis safe harbor election". It covers all qualifying amounts in the year.
  - Sch. C instr. 2026 (draft): de minimis amounts go only in Part V ("Only deduct these amounts as other expenses").
- **Grade:** 2 settled (the regulation sets $500; Notice 2015-82, under the regulation's own delegation, sets $2,500)
- **Rhea's default:**
  - A written expensing policy dated 1 January (or the first day of business) is recorded as the company's fact before the first purchase: items or invoices of $2,500 or less are expensed.
  - The election statement is generated for each return. Materials under NIMS are excluded (SP-10).
  - The safe harbor is not depreciation, so Indiana has nothing to add back (SP-13). That is the reading Rhea ships.
- **Rhea today:** missing. No expensing policy fact and no classification between expense, materials and asset.
- **Where the work lives:** *pack data* (policy fact, classification rule, election statement as an issued document)
- **Deadline:** the policy must exist at the beginning of the tax year; the election goes with the timely filed return (15 April 2027, or with extension)
- **Watch:** the $2,500 amount changes only by an IRS notice or regulation in the Internal Revenue Bulletin (https://www.irs.gov/irb). It has not changed since Notice 2015-82.
- **Sources:** https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.263(a)-1 ; https://www.irs.gov/pub/irs-drop/n-15-82.pdf ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf

---

### SP-12 — The laser cutter and the laptop: §179, bonus depreciation or regular depreciation?
- **Applies to:** everyone (the corporation has no §179 business-income issue of the same kind but the same choices)
- **The law:**
  - §179 for 2026, Rev. Proc. 2025-32 §4.24: the aggregate "cannot exceed $2,560,000" and is reduced by the amount by which §179 property placed in service in 2026 "exceeds $4,090,000"; SUVs are capped at $32,000. These follow P.L. 119-21 §70306 ($2,500,000 / $4,000,000 for years beginning after 2024, indexed after 2025).
  - §179 limits, IRC: §179(b)(3)(A), the deduction "shall not exceed the aggregate amount of taxable income of the taxpayer for such taxable year which is derived from the active conduct by the taxpayer of any trade or business", with a carryover in (B). §179(d)(1)(C): property must be "acquired by purchase for use in the active conduct of a trade or business".
  - Pub. 946: for mixed-use property "you can elect the section 179 deduction only if you use the property more than 50% for business in the year you place it in service".
  - Bonus depreciation, IRC §168(k)(1)(A): "an allowance equal to 100 percent of the adjusted basis of the qualified property". P.L. 119-21 §70301(c): the amendments "shall apply to property acquired after January 19, 2025". Notice 2026-11 (IR-2026-06, 14 January 2026) is the interim guidance.
  - Pub. 946: "Unless you elect out, you must take a 100% special depreciation allowance". The election out is by class of property for a year (§168(k)(7)).
  - Recovery class: Pub. 946 Table B-1 puts computers in class 00.12, a 5-year recovery period. The laser cutter's class depends on the owner's activity under Rev. Proc. 87-56: for example, 24.4 "Manufacture of Wood Products, and Furniture" is 7-year.
  - Since 2018 computers are not listed property: §280F(d)(4)(A) lists passenger automobiles, other transportation property, entertainment and recreation property, and property the Secretary specifies.
- **Grade:** 2 settled. The 100% rate and dates are plain. "Acquired after January 19, 2025" (binding-contract rule) and component elections are settled by Notice 2026-11 pending regulations. The laser cutter's class is settled by Rev. Proc. 87-56 once the activity is named.
- **Rhea's default:**
  - Each purchase over the de minimis amount opens a fixed-asset proposal: placed-in-service date, cost, business-use share, class and recovery period. The owner accepts it; it is never automatic.
  - Rhea proposes §179 for the business share of the laser cutter and the laptop, up to the year's business income, because of Indiana (SP-13). Bonus applies only to what §179 does not take.
  - An asset under 50% business use gets no §179 and no bonus; it gets straight-line under ADS (§280F(b)) if listed.
- **Rhea today:** the Poland pack's `fixed_asset` (KŚT, Polish rates, linear/degressive/one_off) has no rules. It has no US fields: MACRS class, recovery period, convention, method, §179 amount, bonus amount, business-use share or a separate state basis.
- **Where the work lives:** *pack data* (a US asset type and depreciation rules: MACRS tables as dated data, §179 and bonus rules); *shared ground* (the asset register must hold more than one depreciation book per asset — federal and Indiana here, tax and accounting in Poland; today `fixed_asset` is a single Polish line)
- **Deadline:** §179 is elected on the return for the year placed in service (Form 4562); the election out of bonus also goes on that return
- **Watch:** §179 amounts are in the annual inflation revenue procedure (October–November, https://www.irs.gov/irb). Proposed and final §168(k) regulations after P.L. 119-21 will appear in the Federal Register (https://www.federalregister.gov/agencies/internal-revenue-service). Form 4562 drafts are at https://www.irs.gov/draft-tax-forms.
- **Sources:** https://www.irs.gov/pub/irs-drop/rp-25-32.pdf ; https://www.law.cornell.edu/uscode/text/26/179 ; https://www.law.cornell.edu/uscode/text/26/168 ; https://www.law.cornell.edu/uscode/text/26/280F ; https://www.irs.gov/pub/irs-drop/n-26-11.pdf ; https://www.irs.gov/newsroom/treasury-irs-issue-guidance-on-the-additional-first-year-depreciation-deduction-amended-as-part-of-the-one-big-beautiful-bill ; https://www.irs.gov/pub/irs-pdf/p946.pdf

---

### SP-13 — Does Indiana accept the federal depreciation, or do we keep a second depreciation book?
- **Applies to:** everyone with Indiana income tax
- **The law:**
  - IB 118 (May 2026, effective "July 4, 2025 (retroactive)", references IC 6-3-1-3.5, IC 6-3-1-33): "Since 2002, Indiana has largely decoupled from the federal allowances for bonus depreciation and Section 179 expensing. Under Senate Enrolled Act 243 (2026), Indiana extended the decoupling to qualified production property".
  - §179: "Indiana caps the overall Indiana Section 179 Allowance at $25,000 against such property. In the first year, the difference between the Federal Section 179 Allowance and the Indiana Section 179 Allowance is an add-back … In subsequent years, the difference … will be permitted as a deduction." Reporting code 105.
  - Bonus, DOR add-backs page: "Figure the net income (or loss) which would have been included in federal adjusted gross income had the bonus depreciation method not been used. The difference … must be added back." IB 118 uses code 104.
  - That SEA 243-2026 moved Indiana's Code reference date to 1 January 2026 comes from secondary sources (RSM, EY), unverified against the act, which I could not open.
- **Grade:** 1 plain (DOR bulletin; the statutes behind it, IC 6-3-1-33 and 6-3-1-3.5, were not read directly)
- **Rhea's default:**
  - Every asset carries an Indiana basis and schedule beside the federal one. The year's code 104 and 105 adjustments are computed and handed to the preparer.
  - Because §179 up to $25,000 needs no Indiana adjustment, Rhea proposes §179 rather than bonus while the year's §179 stays within $25,000.
  - Worked example (computed here): bonus on an $8,000 laser cutter in 7-year class gives an Indiana add-back of $8,000 − $1,143.20 (14.29% half-year MACRS) = $6,856.80, reversed over the following years. §179 on the same machine gives no adjustment.
- **Rhea today:** missing (one depreciation line per asset at most, and no US asset type)
- **Where the work lives:** *pack data* (Indiana schedule and adjustment rules); *shared ground* (multiple depreciation books per asset, as SP-12)
- **Deadline:** with the IT-40 (15 April 2027)
- **Watch:** Indiana conformity is set each session by the General Assembly (January–March in even years, January–April in odd years; acts often retroactive, as SEA 243-2026 was, to 4 July 2025): https://iga.in.gov. DOR republishes IB 118 and the add-backs page after each change: https://www.in.gov/dor/legal-resources/tax-library/ and https://www.in.gov/dor/i-am-a/individual/indiana-add-backs/.
- **Sources:** https://www.in.gov/dor/files/ib118.pdf ; https://www.in.gov/dor/i-am-a/individual/indiana-add-backs/ ; https://www.irs.gov/pub/irs-pdf/p946.pdf (Table A-1) ; https://rsmus.com/insights/tax-alerts/2026/indiana-federal-tax-conformity-bay-latest-tax-bill.html (secondary)

---

### SP-14 — The family car, the laptop the owner already had, a purchase on the personal card: what is business and what is personal?
- **Applies to:** sole proprietor
- **The law:**
  - IRC §262(a): "Except as otherwise expressly provided in this chapter, no deduction shall be allowed for personal, living, or family expenses." Mixed-use costs are split by business use.
  - Property moved from personal to business use, Pub. 946: "your depreciable basis is the lesser of the following. 1. The FMV of the property on the date of the change in use. 2. Your original cost or other basis adjusted".
  - §179 needs property "acquired by purchase for use in the active conduct of a trade or business" (§179(d)(1)(C)), so a converted personal asset gets none. Pub. 946 lists "Property converted from personal use to business use in the same or later tax year" in its bonus depreciation discussion; the reuse-and-recycling exception it mentions does not fit a used car or laptop (my reading).
  - The SMLLC's own bank account is not a tax requirement. Money between the owner and the business is neither income nor expense, because the owner and the business are one taxpayer (SP-02).
- **Grade:** 1 plain
- **Rhea's default:**
  - A personal asset brought in is entered at the lower of its documented cost and its value on the day of the change, dated, with no §179 and no bonus.
  - A business expense paid from a personal card is booked as an owner contribution plus the expense. A personal purchase paid from the business account is an owner draw.
  - Mixed-use items (phone, internet, laptop) carry a stated business-use share with the date and basis of the estimate. The share is the owner's statement and Rhea never infers it.
- **Rhea today:** the Polish review found the same gap (PIT-05, PIT-13). Nothing records business-use shares, conversions or owner equity movements; the matching engine can pair a personal-card receipt to nothing, which is the case to model.
- **Where the work lives:** *pack data* (owner contribution and draw accounts; business-use share fact; conversion rule)
- **Deadline:** at conversion; shares confirmed at year end
- **Watch:** stable rules; Pub. 946 is revised each year (https://www.irs.gov/forms-pubs/about-publication-946).
- **Sources:** https://www.irs.gov/pub/irs-pdf/p946.pdf ; https://www.law.cornell.edu/uscode/text/26/179 ; https://www.law.cornell.edu/uscode/text/26/262

---

### SP-15 — The workshop at home: how much of the home is deductible, and which method?
- **Applies to:** sole proprietor
- **The law:**
  - IRC §280A(c)(1): a deduction is allowed for the part of the dwelling "exclusively used on a regular basis— (A) as the principal place of business for any trade or business of the taxpayer, (B) as a place of business which is used by patients, clients, or customers …, or (C) in the case of a separate structure which is not attached to the dwelling unit, in connection with the taxpayer's trade or business." The principal place of business "includes a place of business which is used by the taxpayer for the administrative or management activities … if there is no other fixed location".
  - §280A(c)(2): storage space for "the inventory or product samples of the taxpayer held for use in the taxpayer's trade or business of selling products at retail or wholesale" qualifies without exclusivity, "but only if the dwelling unit is the sole fixed location of such trade or business".
  - §280A(c)(5): the deduction may not exceed the gross income from the use less other business deductions; the excess carries forward under the regular method.
  - Simplified method, Rev. Proc. 2013-13: "The prescribed rate is $5.00", on business square footage "not to exceed 300 square feet". The election is per year and a change between methods is not a change of accounting method (Sch. C instr. 2026 (draft)). Under it depreciation of the home "is deemed to be zero". A part-year use is averaged monthly, with "-0-" for a month of under 15 days. The disallowed excess is not carried over.
  - Regular method: Form 8829 (actual share of mortgage interest, taxes, insurance, utilities, repairs and depreciation; a 2026 draft exists).
- **Grade:** 1 plain
- **Rhea's default:**
  - The owner states the workshop's square footage, the home's total, the date use began, and whether it is exclusive or a qualifying storage use. Rhea computes both methods each year and proposes the larger allowable one.
  - Worked example (computed here): 200 sq ft used from 1 April gives (9 × 200) / 12 = 150 average sq ft × $5 = $750 simplified.
  - Home costs are owner-level facts (the mortgage interest belongs to the owner's Schedule A otherwise), so the regular method's inputs come from the owner, not from the business ledger.
- **Rhea today:** missing. No home, area or use facts; the Polish review left the home to PIT-05 and PIT-13.
- **Where the work lives:** *pack data* (home-use facts, both computations, a Form 8829 schedule as a view)
- **Deadline:** with the return (line 30)
- **Watch:** the $5 rate changes only by IRS revenue procedure (unchanged since Rev. Proc. 2013-13; https://www.irs.gov/irb). Pub. 587 is revised annually (https://www.irs.gov/forms-pubs/about-publication-587).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/280A ; https://www.irs.gov/pub/irs-drop/rp-13-13.pdf ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf ; https://www.irs.gov/pub/irs-dft/f8829--dft.pdf ; https://www.irs.gov/pub/irs-pdf/p587.pdf

---

### SP-16 — Driving the family car to fairs: standard mileage rate or actual costs, and which rate on which day?
- **Applies to:** sole proprietor
- **The law:**
  - Notice 2026-10 §3: "The standard mileage rate for transportation or travel expenses for 2026 is 72.5 cents per mile for all miles of business use". §4: the depreciation part is "35 cents per mile for 2026".
  - Announcement 2026-11 (IRB 2026-29, 13 July 2026), "result[ing] from recent increases in the price of fuel": "(1) Business 76 cents per mile". These "apply to deductible transportation expenses paid or incurred … on or after July 1, 2026"; Notice 2026-10 rates "continue to apply … before July 1, 2026".
  - Sch. C instr. 2026 (draft), line 9: "Multiply your business miles driven from January 1, 2026, to June 30, 2026, by 72.5 cents per mile; Multiply your business miles driven from July 1, 2026, to December 31, 2026, by 76 cents per mile; and Add to this amount your parking fees and tolls". Business-use loan interest goes on new line 16b and personal property tax on the vehicle on line 23.
  - Rev. Proc. 2019-46 §4.05(3): the rate is not available for a car on which the taxpayer has claimed accelerated depreciation, §179 or bonus. "By using the business standard mileage rate, the taxpayer has elected to exclude the automobile, if owned, from MACRS". Pub. 463: the rate must be chosen "in the first year the car is available for use in your business".
  - Actual costs instead: the business share of fuel, repairs, insurance, registration and depreciation, capped by §280F(a). Rev. Proc. 2026-15 caps first-year depreciation for a car placed in service in 2026 at $20,300 with bonus (Table 1) or $12,300 without (Table 2), then $19,800, $11,900 and $7,160.
  - Commuting is not business. Travel from a home that is the principal place of business under §280A(c)(1)(A) to another work location is deductible (Sch. C instr. 2026 (draft), line 44b).
- **Grade:** 1 plain
- **Rhea's default:**
  - The standard mileage rate for the family car, chosen in 2026 so that it stays available. The rate is dated master data with a validity interval (72.5¢ for 1 January–30 June 2026, 76¢ for 1 July–31 December 2026), and each trip takes the rate valid on its date.
  - Parking and tolls are added at cost. The owner's share of loan interest goes on line 16b, never on line 9.
  - Worked example (computed here): 1,000 miles in the first half and 1,500 in the second give $725.00 + $1,140.00 = $1,865.00.
- **Rhea today:**
  - The Poland pack has `vehicle` and `mileage_entry` (date, vehicle, route, km) with no rules and no purpose field.
  - More important, a rate cannot be a rule version. The executor runs only the newest active version of a rule and treats `effective_from` as a lower bound, so after a 76¢ version replaces the 72.5¢ one, a June trip entered in July matches neither. VAT-05 in the Polish review needs the same dated master data for VAT rates.
- **Where the work lives:** *shared ground* (dated master data with a validity interval that rules read by event date — one mechanism for mileage rates, VAT rates and tax tables; the vehicle and mileage types moved to common, in miles or with a unit); *pack data* (US rates and caps)
- **Deadline:** the method is fixed on the 2026 return
- **Watch:** the standard mileage notice comes out each December or January as an IRS notice in the IRB, announced in the newsroom (https://www.irs.gov/tax-professionals/standard-mileage-rates). 2026 shows a mid-year announcement is possible, so the watch must run all year. The §280F caps come as a revenue procedure each February–March (Rev. Proc. 2026-15: https://www.irs.gov/pub/irs-drop/rp-26-15.pdf).
- **Sources:** https://www.irs.gov/pub/irs-drop/n-26-10.pdf ; https://www.irs.gov/irb/2026-29_irb ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf ; https://www.irs.gov/pub/irs-drop/rp-19-46.pdf ; https://www.irs.gov/pub/irs-drop/rp-26-15.pdf ; https://www.irs.gov/pub/irs-pdf/p463.pdf

---

### SP-17 — What must the mileage log say, and is a log made later good enough?
- **Applies to:** sole proprietor
- **The law:**
  - IRC §274(d): no deduction for travel, gifts or "any listed property (as defined in section 280F(d)(4))" — a passenger car is listed — "unless the taxpayer substantiates by adequate records or by sufficient evidence corroborating the taxpayer's own statement (A) the amount … (B) the time and place … (C) the business purpose".
  - Reg. §1.274-5T(b)(6), listed property: "(B) Uses. The amount of each business/investment use … based on the appropriate measure (i.e., mileage for automobiles …), and the total use of the listed property for the taxable period", "(ii) Time. Date of the expenditure or use", "(iii) Business or investment purpose".
  - Reg. §1.274-5T(c)(1): "A contemporaneous log is not required, but a record … made at or near the time of the expenditure or use, supported by sufficient documentary evidence, has a high degree of credibility not present with respect to a statement prepared subsequent thereto".
  - Reg. §1.274-5(c)(2)(iii)(A): documentary evidence is required for lodging and "any other expenditure of $75 or more".
  - Schedule C Part IV asks: "Do you have evidence to support your deduction?" and "If 'Yes,' is the evidence written?" It also asks for business, commuting and other miles for the year.
- **Grade:** 1 plain
- **Rhea's default:** Each trip is an event made on the day: date, start and end, miles, purpose (the fair's name), linked to the fair's booking or sales. Odometer readings at 1 January (or first business use) and 31 December give total miles. A trip entered more than a set number of days after its date is flagged as a late record, because the regulation weighs it lower. The fair's sales and booth invoice corroborate the trip.
- **Rhea today:** Polish `mileage_entry` has no purpose and counts kilometres. Rhea's log already records when each fact was entered, which is exactly the "at or near the time" evidence the regulation values.
- **Where the work lives:** *pack data* (trip type with purpose, odometer facts, Part IV view); *shared ground* (the vehicle and mileage types, as SP-16)
- **Deadline:** at or near the time of each trip; the year's totals at year end
- **Watch:** §274 regulations change by Treasury decision (https://www.federalregister.gov/agencies/internal-revenue-service); Pub. 463 is revised yearly (https://www.irs.gov/forms-pubs/about-publication-463).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/274 ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.274-5T ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.274-5 ; https://www.irs.gov/pub/irs-dft/f1040sc--dft.pdf

---

### SP-18 — A buyer pays through the marketplace or the card processor: when is it income, and is it the gross sale or the net payout?
- **Applies to:** everyone (accrual for the corporation differs on timing, not on gross versus net)
- **The law:**
  - Reg. §1.451-2(a): income "is constructively received by him in the taxable year during which it is credited to his account, set apart for him, or otherwise made available so that he may draw upon it at any time … However, income is not constructively received if the taxpayer's control of its receipt is subject to substantial limitations or restrictions."
  - Sch. C instr. 2026 (draft), line F: "Income is constructively received when it is credited to your account or made available to you without restriction."
  - Gross, not net: Schedule C line 1 is "Gross receipts or sales" and line 2 "Returns and allowances". Processor and marketplace fees are expenses ("commissions and fees", line 10).
  - Sales tax, Sch. C instr. 2026 (draft), line 23: taxes "imposed on the buyer that you were required to collect and pay over to state or local governments … are not included in gross receipts or sales nor are they a deductible expense". Indiana's guide calls sales tax a "trust tax". The statute (IC 6-2.5) was not read here.
  - Form 1099-NEC from a design client: "If the total amounts that were reported in box 1a of Forms 1099-NEC are more than the total you are reporting on line 1, attach a statement explaining the difference."
- **Grade:** 2 settled. The marketplace and processor case is settled by Reg. §1.451-2(a): an amount in the seller's balance that the seller may withdraw is received, and an amount held in reserve is not. No ruling on marketplace reserves specifically was found; applying the regulation's "substantial limitations" to a reserve is my reading.
- **Rhea's default:**
  - A sale is income on the date it is credited to the seller's available balance with the marketplace or processor, at the gross sale price before fees, excluding sales tax collected for the state. A reserve or hold is income when released.
  - The bank payout is a transfer from the processor balance, not income. Fees, refunds and chargebacks are separate facts: line 10, line 2 and line 2.
  - Craft-fair cash is income on the day.
  - The design services go on the same Schedule C, as one business with one set of books (Reg. §1.446-1(d)(2): no separate business without "a complete and separable set of books and records").
- **Rhea today:**
  - The matching engine can pair a payout to the sales it settles and derive the fee difference (`matching_difference` exists in the Poland pack).
  - The `sales_invoice` type cannot carry these sales: it requires `vat`, `vat_rate` and a company `buyer`. A marketplace order has a consumer buyer, a US sales tax that is not VAT, and the marketplace, not the seller, is often the party that collects the tax.
- **Where the work lives:** *shared ground* (a sale document that does not presume VAT or a business buyer: gross amount, tax collected for whom, channel, optional buyer; today's sales_invoice is VAT-shaped); *market code* (readers for marketplace and processor settlement reports, as MT940 is for Polish banks); *pack data* (the receipt-date and gross rules)
- **Deadline:** none of its own
- **Watch:** constructive receipt is a long-standing regulation (Federal Register for any change). Schedule C line 1 guidance and the gig-economy pages are at https://www.irs.gov/businesses/gig-economy-tax-center and in each year's draft instructions (https://www.irs.gov/draft-tax-forms).
- **Sources:** https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.451-2 ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.446-1 ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf ; https://www.in.gov/dor/files/new-small-business-handbook.pdf

---

### SP-19 — We will get Forms 1099-K from the marketplace and the processor: what do they mean for our income?
- **Applies to:** everyone
- **The law:**
  - IRC §6050W(e), as amended by P.L. 119-21 §70432: a third party settlement organization reports third-party network transactions "only if— (1) the amount which would otherwise be reported … exceeds $20,000, and (2) the aggregate number of such transactions exceeds 200." §70432(a)(2): it takes effect "as if included in section 9674 of the American Rescue Plan Act", i.e. the $600 rule never applied.
  - The threshold covers third-party network transactions only. A merchant acquirer reports payment card transactions with no minimum; §6050W(d) excludes card arrangements from the "third party payment network" definition.
  - Instructions for Form 1099-K (Rev. December 2026), box 1a: "'Gross amount' means the total dollar amount of total reportable payment transactions … without regard to any adjustments for credits, cash equivalents, discount amounts, fees, refunded amounts, shipping amounts, or any other amounts."
  - The 2026 draft Schedule 1 has a line for "the amount reported to you on Form(s) 1099-K that was included in error or for personal items".
- **Grade:** 1 plain
- **Rhea's default:** A 1099-K is a received document checked against the books, not a source of income. Rhea reconciles each form's box 1a and monthly boxes to its own gross sales for that channel, with the difference explained: refunds, shipping charged, sales tax the platform collected, timing at year end. A card processor's form is expected whatever the volume; a marketplace form only above $20,000 and 200 transactions. A missing expected form is a task, not an error.
- **Rhea today:** the received-document door and blob store can take the form. There is no 1099 reader and no reconciliation view.
- **Where the work lives:** *market code* (a 1099-K reader for PDF or the platform's CSV in markets/us); *pack data* (the reconciliation rule and view)
- **Deadline:** platforms furnish forms by 31 January (for 2026: 1 February 2027, since 31 January is a Sunday — computed); reconciliation before the return
- **Watch:** Form 1099-K instructions are revised each December (https://www.irs.gov/forms-pubs/about-form-1099-k). Threshold changes come from Congress (Congress.gov) and IRS newsroom FAQs (https://www.irs.gov/newsroom/form-1099-k-frequently-asked-questions-general-information).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6050W ; https://www.irs.gov/pub/irs-pdf/i1099k.pdf ; https://www.irs.gov/pub/irs-dft/f1040s1--dft.pdf

---

### SP-20 — We pay freelancers: which forms, from what amount, and by when?
- **Applies to:** everyone
- **The law:**
  - IRC §6041(a), as amended by P.L. 119-21 §70433: a person in business paying another "compensations, remunerations … of $2,000 or more in any calendar year" must file an information return. Rev. Proc. 2025-32 §2.15: for payments "made after December 31, 2025, the base threshold … is $2,000", indexed for returns filed in 2027.
  - Sch. C instr. 2026 (draft), line 11: "You must file Form 1099-NEC to report contract labor payments of $2,000 or more during the year." Lines I and J ask whether 1099s were required and filed.
  - Instructions for Forms 1099-MISC and 1099-NEC (Rev. December 2026): "Payments made with a credit card or payment card and certain other types of payments, including third-party network transactions, must be reported on Form 1099-K by the payment settlement entity under section 6050W and are not subject to reporting on Form 1099-MISC" (or 1099-NEC).
  - Due date, §6071(c): returns "required … to report nonemployee compensation shall be filed on or before January 31". 31 January 2027 is a Sunday, so 1 February 2027 (computed, §7503).
  - E-filing is required at 10 or more information returns in aggregate (T.D. 9972, per Sch. C instr. 2026 (draft)). For returns due in 2027, the late or incorrect filing penalty for a business with receipts of $5,000,000 or less is $340 per return (Rev. Proc. 2025-32 §4.57).
  - The contractor's TIN comes from Form W-9; backup withholding applies when no TIN is furnished (§3406, cross-referenced in §6041 as amended; not quoted here).
- **Grade:** 1 plain
- **Rhea's default:**
  - A contractor is a person or company with a W-9 on file before the first payment. Rhea tracks each contractor's calendar-year total paid by cash, check or bank transfer, leaving out card and PayPal-type payments.
  - Crossing $2,000 opens a 1099-NEC task. The forms are drafted in January from the year's facts.
  - The 2026 threshold ($2,000) and the 2027 indexed one are dated master data.
- **Rhea today:** the Polish review's people and contractor types have names and a NIP, with no W-9, no TIN status and no payment-method distinction. There is no 1099 writer.
- **Where the work lives:** *pack data* (contractor facts, threshold rule, tasks); *market code* (a 1099-NEC writer and IRS e-file transmitter, like JPK_V7M for Poland)
- **Deadline:** 1 February 2027 for 2026 payments (furnish to the contractor and file with the IRS)
- **Watch:** the indexed threshold appears in the annual inflation revenue procedure (October–November, https://www.irs.gov/irb). Form instructions are revised each December (https://www.irs.gov/forms-pubs/about-form-1099-nec). The e-file rules are at https://www.irs.gov/filing/e-file-information-returns.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6041 ; https://www.law.cornell.edu/uscode/text/26/6071 ; https://www.irs.gov/pub/irs-drop/rp-25-32.pdf ; https://www.irs.gov/pub/irs-pdf/i1099mec.pdf ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf

---

### SP-21 — Lunch with a buyer, coffee at a fair, a client's concert tickets: what is deductible?
- **Applies to:** everyone
- **The law:**
  - IRC §274(n)(1): "The amount allowable as a deduction under this chapter for any expense for food or beverages shall not exceed 50 percent of the amount of such expense".
  - §274(a)(1)(A): no deduction "With respect to an activity which is of a type generally considered to constitute entertainment, amusement, or recreation".
  - Sch. C instr. 2026 (draft), line 24b: the meal must be ordinary and necessary, "not lavish or extravagant", "You or your employee was present", provided to "a current or potential business customer, client, consultant, or similar business contact". At an entertainment event, food must be "purchased separately from the entertainment" or stated separately (Reg. §1.274-11, -12).
  - New for 2026: "no deduction is allowed for expenses paid or incurred after 2025 by an employer, including a sole proprietor, for meals provided to employees at an employer-operated facility or meals provided for the convenience of the employer" (§274(o)). It does not bite with no employees.
  - Receipts are required at $75 or more (Reg. §1.274-5(c)(2)(iii)).
- **Grade:** 1 plain
- **Rhea's default:** Meals are booked at full cost in a meals account with the attendee and purpose; line 24b takes 50%. Entertainment is booked to a non-deductible account, kept for the record. The owner's own meal on a day trip is personal unless it is a business meal with a contact; the owner's meals while away overnight follow SP-22.
- **Rhea today:** missing (no meal or entertainment classification)
- **Where the work lives:** *pack data*
- **Deadline:** —
- **Watch:** §274 amendments come from Congress (P.L. 119-21 last). Regulations and the per diem notice come out each September or October in the IRB (https://www.irs.gov/irb).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/274 ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.274-5

---

### SP-22 — An out-of-town craft fair with a hotel night: what travel costs count, and can we use per diem?
- **Applies to:** sole proprietor
- **The law:**
  - Sch. C instr. 2026 (draft), line 24a: "lodging and transportation connected with overnight travel for business while away from your tax home. In most cases, your tax home is your main place of business". Meals go on 24b (50%).
  - Standard meal allowance: "Instead of deducting the actual cost of your meals while traveling away from home, you can use the standard meal allowance", which "is the federal meals and incidental expenses (M&IE) rate" (GSA). "However, you must still keep records to prove the time, place, and business purpose of your travel."
  - Pub. 463: the allowance may be used "whether you are an employee or self-employed". For a self-employed person "Your allowable lodging expense deduction is your actual cost". The first and last travel day are three-quarters or prorated.
  - High-cost localities: Notice 2025-54 for 1 October 2025–30 September 2026 and Notice 2026-60 from 1 October 2026.
  - Elements to prove (Reg. §1.274-5T(b)(2)): amount, dates of departure and return, destination, business purpose. Receipts are required for all lodging.
- **Grade:** 1 plain
- **Rhea's default:** A trip is an event (dates, destination, fair) to which lodging, transport and meals attach. Rhea offers the M&IE rate for the destination and date as dated master data, against actual meals, and books the larger defensible figure at 50%. Lodging is always actual.
- **Rhea today:** missing
- **Where the work lives:** *pack data* (trip type, GSA M&IE table as dated data); *market code* later (a reader for GSA per diem files)
- **Deadline:** —
- **Watch:** GSA publishes per diem rates for each federal fiscal year (1 October) at https://www.gsa.gov/travel/plan-book/per-diem-rates. The IRS special per diem notice comes each September or October (Notice 2026-60: https://www.irs.gov/pub/irs-drop/n-26-60.pdf).
- **Sources:** https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf ; https://www.irs.gov/pub/irs-pdf/p463.pdf ; https://www.irs.gov/pub/irs-drop/n-26-60.pdf ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.274-5T

---

### SP-23 — The owner buys health insurance: is it a business expense?
- **Applies to:** sole proprietor
- **The law:**
  - IRC §162(l)(1): a self-employed individual is "allowed as a deduction … an amount equal to the amount paid during the taxable year for insurance which constitutes medical care for— (A) the taxpayer, (B) the taxpayer's spouse, (C) the taxpayer's dependents, and (D) any child … who as of the end of the taxable year has not attained age 27."
  - (2)(A): not more than "the taxpayer's earned income … derived by the taxpayer from the trade or business with respect to which the plan … is established".
  - (2)(B): not "for any calendar month for which the taxpayer is eligible to participate in any subsidized health plan maintained by any employer of the taxpayer or of the spouse".
  - It is claimed on Schedule 1 line 17 ("Self-employed health insurance deduction", 2026 draft), not on Schedule C. The IRS's Form 7206 computes it.
  - For §199A it reduces qualified business income (Reg. §1.199A-3(b)(1)(vi)).
- **Grade:** 1 plain
- **Rhea's default:** Premiums are owner-level facts by month, with an "eligible for employer-subsidized plan" flag per month (the owner's or spouse's job). Rhea computes the deduction up to the business's earned income and hands it to the return. It never posts to a Schedule C line.
- **Rhea today:** missing (the people pack has no insurance or household facts)
- **Where the work lives:** *pack data*
- **Deadline:** with the return
- **Watch:** stable statute; Form 7206 instructions yearly (https://www.irs.gov/forms-pubs/about-form-7206).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/162 ; https://www.irs.gov/pub/irs-dft/f1040s1--dft.pdf ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.199A-3

---

### SP-24 — How much self-employment tax will the owner pay, and what is deductible?
- **Applies to:** sole proprietor
- **The law:**
  - IRC §1401(a): "a tax equal to 12.4 percent of the amount of the self-employment income". §1401(b)(1): "2.9 percent". §1401(b)(2)(A): an additional "0.9 percent of the self-employment income … in excess of" $200,000 for a single filer ($250,000 joint).
  - §1402(a)(12): net earnings are reduced by "one-half of the sum of the rates imposed by subsections (a) and (b) of section 1401" (7.65%). The 2026 draft Schedule SE line 4a: "multiply line 3 by 92.35% (0.9235)".
  - §1402(b)(1): the 12.4% part stops at the contribution and benefit base less wages. SSA: "$184,500 for remuneration paid in 2026 and self-employment income earned in tax years beginning in 2026" (FR Doc. 2025-19763). §1402(b)(2) and §6017: nothing is due, and no return is required for it, under $400.
  - §164(f)(1): a deduction of "one-half of the taxes imposed by section 1401 (other than the taxes imposed by section 1401(b)(2))", on Schedule 1 line 15.
- **Grade:** 1 plain
- **Rhea's default:**
  - SE tax is computed from Schedule C line 31 as the year runs, for the estimated payments (SP-25). The base and rates are dated data; the base is reduced by any W-2 wages the owner also earns, which the owner states.
  - Worked example (computed here): a net profit of $40,000 gives $36,940.00 × 15.3% = $5,651.82 of SE tax and a $2,825.91 deduction.
  - The tax is an owner-level liability and never a business expense.
- **Rhea today:** missing. The formula language's exact rationals and declared rounding fit the computation. If the preparer rounds to whole dollars (the Form 1040 instructions allow it; not fetched here), that is a declared rounding.
- **Where the work lives:** *pack data*
- **Deadline:** paid through estimated tax; the return by 15 April 2027
- **Watch:** SSA announces the next year's base with the COLA in October and publishes it in the Federal Register in October–November (https://www.federalregister.gov/agencies/social-security-administration). The 2027 notice had not appeared by 16 October 2026. Draft Schedule SE: https://www.irs.gov/draft-tax-forms.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/1401 ; https://www.law.cornell.edu/uscode/text/26/1402 ; https://www.law.cornell.edu/uscode/text/26/164 ; https://www.law.cornell.edu/uscode/text/26/6017 ; https://www.federalregister.gov/documents/2025/11/03/2025-19763/cost-of-living-increase-and-other-determinations-for-2026 ; https://www.irs.gov/pub/irs-dft/f1040sse--dft.pdf

---

### SP-25 — Quarterly federal estimated tax in the first year: how much, and by when?
- **Applies to:** sole proprietor (the corporation pays under §6655)
- **The law:**
  - IRC §6654(c)(2): installments are due "April 15", "June 15", "September 15" and "January 15 of the following taxable year". For 2026, Form 1040-ES gives 15 April 2026, 15 June 2026, 15 September 2026 and 15 January 2027. The last is not needed "if you file your 2026 tax return by February 1, 2027, and pay the entire balance due with your return."
  - §6654(d)(1)(B): each installment is 25% of the lesser of "90 percent of the tax shown on the return for the taxable year" or "100 percent of the tax shown on the return of the individual for the preceding taxable year". The prior-year route "shall not apply if the preceding taxable year was not a taxable year of 12 months or if the individual did not file a return". (C)(i): 110% if the prior year's AGI "exceeds $150,000".
  - §6654(e)(1): no penalty if the tax shown less withholding "is less than $1,000". (e)(2): none if the preceding year was 12 months and the individual "did not have any liability for tax" for it.
  - The annualized method (§6654(d)(2), Form 2210 Schedule AI) suits income that starts mid-year. The underpayment rate for individuals is 7% for the quarter from 1 October 2026 (IR-2026-98).
  - The owner's whole tax counts (Schedule C, SE tax, other income, less W-2 withholding), not only the business.
- **Grade:** 1 plain
- **Rhea's default:**
  - At the start, the owner states the 2025 tax, the 2025 AGI and any withholding for 2026. Rhea sets the required annual payment to the safe harbor, 100% or 110% of 2025 tax, or zero under §6654(e)(2). The four installments are dated duties for the clock.
  - With no 2025 return, Rhea uses 90% of a running projection, with an annualized option.
  - Each payment is an owner-level fact (not a business expense) matched to the bank line.
- **Rhea today:** the clock fires dated duties and the Poland pack shows the pattern (PIT advances, ZUS). The US amounts and the owner-level facts (prior-year tax, AGI, withholding) are missing.
- **Where the work lives:** *pack data* (schedule, safe harbors, projection); *shared ground* (owner-level tax facts that are not the company's: Poland's PIT and ZUS sit with the person too, but today live on `company`)
- **Deadline:** 15 April, 15 June, 15 September 2026; 15 January 2027
- **Watch:** Form 1040-ES is issued each January (https://www.irs.gov/forms-pubs/about-form-1040-es). Interest rates are announced quarterly in the IRS newsroom (https://www.irs.gov/payments/quarterly-interest-rates). Disaster postponements come as IRS news releases (https://www.irs.gov/newsroom/tax-relief-in-disaster-situations).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6654 ; https://www.irs.gov/pub/irs-pdf/f1040es.pdf ; https://www.irs.gov/pub/irs-pdf/i2210.pdf ; https://www.irs.gov/newsroom/interest-rates-remain-the-same-for-the-fourth-quarter-of-2026

---

### SP-26 — Indiana estimated tax (state and county): does it follow the federal rules?
- **Applies to:** sole proprietor
- **The law:**
  - IB 3 (March 2023, references IC 6-3-4-4.1): estimated payments "must be made by an individual who: (1) Receives income from which Indiana adjusted gross income tax or local income tax is not properly withheld; and (2) Has a total unpaid state and local income tax liability that is $1,000 or more for the taxable year". Payments are due "April 15, June 15, September 15, and January 15", by ES-40 or INTIME. The taxpayer "is required to separately identify the amount of state adjusted gross income tax and local income tax", though "a failure to allocate a payment properly will not result in a penalty".
  - No penalty if each installment is a quarter of the lesser of "90 percent of the tax shown on the current year's return" or "100 percent of the tax shown on the previous year's return", 110% above $150,000 federal AGI. The penalty "will be 10 percent of the underpayment for that period".
  - DOR's estimated payments page: "if you owe $1,000 or more in state and county tax for the year that's not covered by withholding taxes, you need to make estimated tax payments". DOR's deadline page lists "File income tax return and pay balance due in lieu of making fourth estimated tax installment payment".
- **Grade:** 1 plain (from DOR's bulletin; IC 6-3-4-4.1 itself not read)
- **Rhea's default:** The same four dates as the federal schedule, computed on Indiana AGI at 2.95% plus Marion County's 2.02% (SP-27), each payment split into its state and county parts. The penalty is a flat 10% per quarter, not interest, so a late payment's cost is known on the day it is late.
- **Rhea today:** missing (as SP-25)
- **Where the work lives:** *pack data*
- **Deadline:** 15 April, 15 June, 15 September 2026; 15 January 2027
- **Watch:** DOR's estimated payments page (https://www.in.gov/dor/i-need-to/make-a-payment/estimated-payments/) and IB 3 (https://www.in.gov/dor/files/ib03.pdf). Statutory changes come from the General Assembly (https://iga.in.gov).
- **Sources:** https://www.in.gov/dor/files/ib03.pdf ; https://www.in.gov/dor/i-need-to/make-a-payment/estimated-payments/ ; https://www.in.gov/dor/i-am-a/individual/tax-filing-deadlines/

---

### SP-27 — Indiana income tax and the Marion County tax: what rate, set by whom, changing when?
- **Applies to:** sole proprietor
- **The law:**
  - DOR rates page: "The Indiana Individual adjusted gross income tax rate for 2026 is 2.95% and will adjust in 2027 to 2.90%." The rate is in IC 6-3-2-1 (not read).
  - County tax, Departmental Notice #1 (effective 1 January 2026): "Both the county of residence and the county of principal business or employment of an individual are determined on Jan. 1 of the calendar year in which the individual's taxable year begins." Marion is county 49 at "0.0202". DOR: "County income tax rates may be adjusted in January and October."
  - The base: Indiana AGI starts from federal AGI with Indiana's modifications (IC 6-3-1-3.5, per IB 118), less exemptions ($1,000 each, IB 3). The county tax is on the same base.
  - How county rates are set: local income tax (LIT) rates are adopted by county ordinance under IC 6-3.6. Senate Enrolled Act 1-2025 rebuilt the LIT; HEA 1210-2026 (signed 12 March 2026) delays it. DLGF: "the current LIT ordinances would expire on January 1, 2029, and counties and eligible municipalities may adopt new LIT ordinances starting on July 1, 2028, for collections starting in 2029". From then a municipal LIT rate (IC 6-3.6-6-22) can apply by residence in a municipality.
  - The return is IT-40, due 15 April (DOR deadlines page).
- **Grade:** 1 plain
- **Rhea's default:**
  - The owner's Indiana county of residence on 1 January is a dated fact each year; a move in the year changes nothing until the next 1 January. State and county rates are dated master data, read from Departmental Notice #1 each January and October.
  - From the 2029 collections, Rhea will need the municipality too. The address must resolve to the municipality, not just the county.
- **Rhea today:** missing (the people pack has no residence history)
- **Where the work lives:** *pack data* (rates, residence fact, IT-40 figures); *shared ground* (dated master data with validity, as SP-16)
- **Deadline:** IT-40 by 15 April 2027 (extended: 15 November 2027 per DOR's deadline page)
- **Watch:**
  - DOR Departmental Notice #1, republished each 1 January and 1 October with changed counties starred: https://www.in.gov/dor/files/dn01.pdf. Rates page: https://www.in.gov/dor/resources/tax-rates-and-reports/rates-fees-and-penalties/.
  - County ordinances are adopted by the county council or LIT council, then certified. DLGF memos at https://www.in.gov/dlgf/ track the LIT overhaul.
  - The state rate schedule is changed by the General Assembly (https://iga.in.gov); further cuts are conditional on revenue triggers (secondary — verify).
- **Sources:** https://www.in.gov/dor/resources/tax-rates-and-reports/rates-fees-and-penalties/ ; https://www.in.gov/dor/files/dn01.pdf ; https://www.in.gov/dor/files/ib03.pdf ; https://www.in.gov/dlgf/files/2026-memos/Legislation-Affecting-Local-Income-Taxes.pdf ; https://www.in.gov/dor/i-am-a/individual/tax-filing-deadlines/

---

### SP-28 — A retirement plan for the owner: SEP-IRA or solo 401(k), how much, and is it a business expense?
- **Applies to:** sole proprietor (the corporation sponsors plans for employees, including an owner-employee)
- **The law:**
  - 2026 limits, Notice 2025-67: "The limitation for defined contribution plans under section 415(c)(1)(A) is increased in 2026 from $70,000 to $72,000". Elective deferrals under §402(g)(1) rise "from $23,500 to $24,500"; the age-50 catch-up rises "from $7,500 to $8,000". The compensation limit rises "from $350,000 to $360,000".
  - Compensation for the owner is "earned income": net earnings from self-employment after the deduction for half the SE tax and after the contribution itself (IRS one-participant 401(k) page; Pub. 560's rate table). A 25% plan rate therefore works out to 20% of that net (Pub. 560: "20% of your net earnings from self-employment").
  - SEP, IRS: "You can set up a SEP for a year as late as the due date (including extensions) of your business income tax return", and contributions "must be made by the due date (including extensions)".
  - Solo 401(k), Pub. 560 (2025): "For 2023 and later years, a sole proprietor with no employees can adopt a section 401(k) plan after the end of the tax year, provided the plan is adopted by the tax filing deadline (without regard to extensions)." A 401(k) needs an EIN (SP-03, Schedule C line D).
  - The deduction is on Schedule 1 line 16 ("Self-employed SEP, SIMPLE, and qualified plans"). Schedule C line 19 is for employees' plans. It reduces qualified business income (Reg. §1.199A-3(b)(1)(vi)).
- **Grade:** 1 plain
- **Rhea's default:** Rhea computes the year's maximum for each plan type from the projected Schedule C, so the owner can choose before the deadline. Contributions are owner-level facts (draws from the business account), never Schedule C expenses. The plan's adoption date is a fact.
- **Rhea today:** missing
- **Where the work lives:** *pack data*
- **Deadline:** SEP: the return due date with extensions (15 October 2027 if extended); a solo 401(k) adopted after year end: by 15 April 2027
- **Watch:** retirement limits come each November in an IRS notice and the newsroom (https://www.irs.gov/retirement-plans/cola-increases-for-dollar-limitations-on-benefits-and-contributions). Pub. 560 is revised yearly (https://www.irs.gov/forms-pubs/about-publication-560).
- **Sources:** https://www.irs.gov/pub/irs-drop/n-25-67.pdf ; https://www.irs.gov/retirement-plans/one-participant-401k-plans ; https://www.irs.gov/retirement-plans/plan-sponsor/simplified-employee-pension-plan-sep ; https://www.irs.gov/pub/irs-pdf/p560.pdf ; https://www.irs.gov/pub/irs-dft/f1040s1--dft.pdf

---

### SP-29 — Year end: must we count what is left, and how is cost of goods sold figured?
- **Applies to:** everyone with goods
- **The law:**
  - Schedule C Part III (2026 draft): "33 Method(s) used to value closing inventory: Cost / Lower of cost or market / Other"; "35 Inventory at beginning of year"; "36 Purchases less cost of items withdrawn for personal use"; "37 Cost of labor. Do not include any amounts paid to yourself"; "38 Materials and supplies"; "39 Other costs"; "41 Inventory at end of year"; "42 Cost of goods sold".
  - Under NIMS (SP-10), the year's cost of goods sold is the direct material cost of goods provided to customers in the year that has also been paid. Materials on hand at 31 December stay in stock (Reg. §1.471-1(b)(4)(i)).
  - Under the books-and-records method, a physical count used in the books must be used for tax (Reg. §1.471-1(b)(6)(iii), Example 1).
  - Items the owner takes home for personal use come out of purchases (line 36).
- **Grade:** 2 settled. Whether a NIMS taxpayer must take a physical count is not stated; Reg. §1.471-1(b)(4)(ii) requires consistent specific identification, FIFO or average, which presumes knowing what is on hand. That a count is not compulsory but the identification must be supportable is my reading.
- **Rhea's default:** On closing the year, open a stocktake task dated 31 December: count materials and finished goods. The count adjusts the FIFO layers, and differences go to a loss (breakage, waste) or a personal-use withdrawal. Part III is a view over the stock facts: opening = last year's closing (zero in 2026), purchases, closing, cost of goods sold.
- **Rhea today:** the Poland pack's `stocktake` (date, comment, status, value) has no lines and no rules. The Polish review asks for the same count (KPIR-10).
- **Where the work lives:** *shared ground* (stocktake with item lines and valuation, shared by KPiR and Schedule C); *pack data* (the Part III view)
- **Deadline:** count as at 31 December 2026; on the return by 15 April 2027
- **Watch:** inventory regulations (Federal Register, https://www.federalregister.gov/agencies/internal-revenue-service); draft Schedule C each year (https://www.irs.gov/draft-tax-forms).
- **Sources:** https://www.irs.gov/pub/irs-dft/f1040sc--dft.pdf ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.471-1 ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf

---

### SP-30 — The 20% qualified business income deduction: does it apply, and what changed for 2026?
- **Applies to:** sole proprietor (a C corporation has no §199A)
- **The law:**
  - IRC §199A(a): for "a taxpayer other than a corporation, except as provided in subsection (i), … a deduction … equal to the lesser of— (1) the combined qualified business income amount of the taxpayer, or (2) an amount equal to 20 percent of the excess (if any) of— (A) the taxable income … over (B) the net capital gain".
  - P.L. 119-21 §70105 struck the rule that the section "shall not apply to taxable years beginning after December 31, 2025" and widened the phase-in to "$75,000 ($150,000 in the case of a joint return)". New §199A(i): the deduction is "the greater of— (A) the amount … determined without regard to this subsection, or (B) $400" for a taxpayer whose QBI from active qualified businesses "is at least $1,000", for "taxable years beginning after December 31, 2025" (§70105(c)).
  - Rev. Proc. 2025-32 §4.26 for 2026: threshold "$201,750" (single), "$403,500" (joint); the phase-in ends at $276,750 / $553,500. Below the threshold the W-2 wage and property limits and the specified-service exclusion do not apply (§199A(b)(3)(A), (d)(3)).
  - QBI is net of "the deductible portion of the tax on self-employment income under section 164(f), the self-employed health insurance deduction under section 162(l), and the deduction for contributions to qualified retirement plans" (Reg. §1.199A-3(b)(1)(vi)).
  - The 2026 draft Form 8995 adds line 16, "Minimum deduction for active qualified business income".
- **Grade:** 1 plain
- **Rhea's default:** QBI = Schedule C line 31 less half the SE tax, the self-employed health insurance and the retirement contribution. The deduction is 20% of that, capped at 20% of taxable income less net capital gain (owner-level facts). Rhea flags when projected taxable income approaches $201,750, the point where the design services' character (specified service or not) would start to matter.
- **Rhea today:** missing
- **Where the work lives:** *pack data*
- **Deadline:** with the return (Form 8995)
- **Watch:** thresholds come each October–November in the inflation revenue procedure (https://www.irs.gov/irb); the $400 and $1,000 amounts are indexed from 2027 (Rev. Proc. 2025-32 §2.12). Form 8995 drafts: https://www.irs.gov/draft-tax-forms.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/199A ; https://www.irs.gov/pub/irs-drop/rp-25-32.pdf ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-A/part-1/section-1.199A-3 ; https://www.irs.gov/pub/irs-dft/f8995--dft.pdf

---

### SP-31 — Indiana taxes business equipment: do we owe personal property tax on the laser cutter?
- **Applies to:** everyone with business property in Indiana
- **The law:**
  - IC 6-1.1-3-7.2(d), as amended by SEA 1-2025 and HEA 1427-2025, quoted by DLGF: "if the acquisition cost of a taxpayer's total business personal property in a county is less than: (1) eighty thousand dollars ($80,000) for assessment dates before 2026; and (2) two million dollars ($2,000,000) for the 2026 assessment date, and each assessment date thereafter; the taxpayer's business personal property in the county for that assessment date is exempt from taxation."
  - (e): an eligible taxpayer "shall include … on the taxpayer's personal property tax return: (1) A declaration that the taxpayer's business personal property in the county is exempt". (f): having done so once, the taxpayer "is not required to file a personal property return … unless or until the taxpayer no longer qualifies".
  - Form 103-Long for 2026: the return speaks "as of" "JANUARY 1, 2026". It "must be filed with the township assessor, if any, or the county assessor … not later than May 15, 2026", with "Form 104", and "To establish an exemption in a new county, new or existing businesses must file a personal property return".
  - DLGF (memo of 23 May 2025, on HEA 1427 §§13, 15, 16): "taxpayers can no longer file personal property tax returns through PPOP-IN after 2025".
- **Grade:** 1 plain (statute quoted through DLGF; IC 6-1.1 not read directly)
- **Rhea's default:**
  - The business holds property on the 1 January assessment date from 2027. Rhea opens a task to file Form 103 with Form 104 by 15 May 2027 declaring the exemption, with the county's total acquisition cost from the fixed-asset register.
  - After that, Rhea checks each 1 January that the total stays under $2,000,000 and files nothing while it does. If the business already held equipment on 1 January 2026, the first declaration was due 15 May 2026 and is late: a task, not a silent pass.
  - Whether the inventory exemption covers materials and stock was not researched here.
- **Rhea today:** missing (the fixed-asset register would supply acquisition cost by county)
- **Where the work lives:** *pack data*
- **Deadline:** 15 May of the year after the business first holds property on 1 January (here 15 May 2027), extendable by 30 days
- **Watch:** DLGF memos after each session (https://www.in.gov/dlgf/) and its yearly Form 103/104 revision (State Forms 11405, 10068). The General Assembly (https://iga.in.gov) changes IC 6-1.1, often retroactively to 1 January, as SEA 1 and HEA 1427 did in 2025.
- **Sources:** https://www.in.gov/dlgf/files/2025-memos/250523-Cockerill-Memo-Legislation-Affecting-Assessment-Matters.pdf ; https://www.in.gov/dlgf/files/2026-level-i-certification/2026-Level-I-Personal-Property.pdf ; https://www.in.gov/counties/howard/departments/assessor/103-Long-Business-Tangible-Personal-Property-2026.pdf.pdf (the state form as posted by a county)

---

### SP-32 — Year end: what must the books hand the return preparer, and by when are the returns due?
- **Applies to:** sole proprietor
- **The law:**
  - Form 1040 with Schedule C, Schedule SE, Schedule 1 and the supporting forms is due 15 April 2027 (§6072(a)); an extension (Form 4868, not researched here) moves filing, not payment, to 15 October 2027.
  - IT-40 is due 15 April 2027, extended to 15 November 2027 (DOR).
  - Supporting forms for this business:
    - Form 4562 (depreciation, §179; listed property);
    - Form 4562-B (amortization beginning in 2026, new);
    - Form 8829 or the simplified worksheet;
    - Form 8995 (QBI) and Form 7206 (self-employed health insurance);
    - the de minimis election statement (Reg. §1.263(a)-1(f)(5));
    - Indiana add-back codes 104 and 105 (IB 118).
  - Information returns issued (1099-NEC, SP-20) and received (1099-K, 1099-NEC, SP-19) feed Schedule C lines I and J and line 1.
- **Grade:** 1 plain
- **Rhea's default:**
  - A year-close package that is a view, not a form:
    - Schedule C by line with every figure explained to its documents;
    - Part III with the count, and Part IV with the mileage log totals;
    - the asset schedule, federal and Indiana;
    - the home-office computation;
    - the 1099 reconciliation;
    - estimated payments made (federal, Indiana state and county);
    - the owner-level items (half SE tax, health insurance, retirement, QBI inputs).
  - Rhea does not file the 1040. The preparer files it, and Rhea keeps the filed figures as a fact so that next year's safe harbors (SP-25, SP-26) read them.
- **Rhea today:** the period close and issued-document seam (JPK_V7M precedent) exist. `tax_declaration` is a Polish enum (jpk_v7, vat7, pit, zus_dra, pcc3, other).
- **Where the work lives:** *pack data* (the package view, a US declaration type); *market code* (an export to preparer software, for example a structured file per form, if a preparer asks; later)
- **Deadline:** returns 15 April 2027; 1099-NEC 1 February 2027
- **Watch:** IRS draft forms page (https://www.irs.gov/draft-tax-forms) from spring; final forms each December–January (https://www.irs.gov/forms-instructions); DOR forms (https://www.in.gov/dor/tax-forms/) each January.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6072 ; https://www.irs.gov/pub/irs-dft/i1040sc--dft.pdf ; https://www.in.gov/dor/i-am-a/individual/tax-filing-deadlines/ ; https://www.in.gov/dor/files/ib118.pdf

---

### SP-33 — Later: should the LLC elect S corporation status, and what would change in the books?
- **Applies to:** sole proprietor (compare the C corporation section)
- **The law:**
  - IRC §1362(b)(1): the election may be made "(A) at any time during the preceding taxable year, or (B) at any time during the taxable year and on or before the 15th day of the 3d month of the taxable year". For a calendar 2027 election: by 15 March 2027 (computed).
  - Reg. §301.7701-3(c)(1)(v)(C): "An eligible entity that timely elects to be an S corporation under section 1362(a)(1) is treated as having made an election under this section to be classified as an association". Reg. §301.6109-1(h)(2)(ii): the entity then "must acquire an EIN and not use the TIN of the single owner".
  - Employment taxes on the owner's wages follow from the corporation treatment for Subtitle C (Reg. §301.7701-2(c)(2)(iv)). The requirement that those wages be reasonable is IRS position and case law, not researched here.
- **Grade:** 1 plain for the mechanics; whether to elect is a business judgment, not a legal question
- **Rhea's default:**
  - Not before the first year closes: the comparison needs a full year's profit.
  - If elected, the classification fact changes on its effective date (SP-02), and rules branch from that date:
    - a separate entity return (Form 1120-S, Indiana IT-20S);
    - payroll for the owner (W-2, Form 941, Indiana withholding);
    - shareholder basis and distributions instead of draws;
    - health insurance as wages; retirement through payroll;
    - estimated tax for SE tax disappears for that income.
  - The ledger, the asset register and the Schedule C line map carry over as history. Nothing is restated, because the classification is a dated fact.
- **Rhea today:** the dated-classification design (SP-02) is what would make this a rule branch rather than a migration; it does not exist yet.
- **Where the work lives:** *later*
- **Deadline:** 15 March of the year it is to take effect (late-election relief exists; not researched)
- **Watch:** Form 2553 and instructions (https://www.irs.gov/forms-pubs/about-form-2553); revenue procedures on late elections in the IRB (https://www.irs.gov/irb).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/1362 ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-F/part-301/section-301.7701-3 ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-F/part-301/section-301.6109-1 ; https://www.ecfr.gov/current/title-26/chapter-I/subchapter-F/part-301/section-301.7701-2


## Small corporation — books and income tax

Scope: "Hoosier Supply Inc.", an Indiana C corporation in its first year (calendar 2026). It is a wholesale distributor with about $3M of revenue, accrual method, FIFO stock in one Indiana warehouse, a leased building, a forklift, a van and racking, a bank line of credit whose lender wants annual statements, ten employees on an outside payroll, and two owners. It sells into Ohio from year one and crosses California's threshold in year two. The section was reviewed against what Rhea's shared ground gives today (`internal/core`, `testdata/seed/finance_v3…v9.json`, `packs/common/periods.json`, `testdata/seed/warehouse.json`), and against the Poland pack's `company` v7, `fixed_asset` v1, `lease_contract` v1 and `tax_declaration` v1 (`packs/pl/pack.json` v23), because a US pack would be built beside them.

Texts used, as of 16 October 2026:
- **Federal statute.** Internal Revenue Code text from Cornell LII. The official uscode.house.gov says it is current through Public Law 119-120 (7 October 2026), but it was serving a temporary format and returned "Page Not Found" for section pages.
- **Federal regulations and IRS guidance.** eCFR title 26 (up to date as of 14 October 2026). Rev. Proc. 2025-32 (the 2026 inflation amounts), Notice 2026-11 (bonus depreciation), Rev. Proc. 2026-15 (vehicle limits), Notice 2015-82, Rev. Proc. 98-25 and Rev. Proc. 2025-28. The IRS 2025 Instructions for Form 1120, 1120-S, 4562 and 2220 are the newest on irs.gov; the 2026 editions are not yet published. Also Pub. 542 (Rev. January 2024) and Pub. 946 (2025).
- **Indiana.** DOR Information Bulletins #11 (December 2022), #12 (October 2023), #118 (May 2026) and #303 (December 2023), the DLGF Form 103-Short for the 1 January 2026 assessment, and the Secretary of State's pages. The Indiana Code was read only in the 2015 unofficial copy on statecodesfiles.justia.com: iga.in.gov serves a script-only page to a non-browser client, and forms.in.gov and inbiz.in.gov refused (403).
- **Ohio.** Ohio Department of Taxation Information Release CAT 2023-01 (21 August 2023). codes.ohio.gov and tax.ohio.gov could not be reached.
- **California.** FTB Form 100 Booklet 2025, FTB Publication 1050 (Rev. 05-2022) and the FTB "Doing business in California" page (last updated 7 November 2025). leginfo.legislature.ca.gov refused (403), so no R&TC text was read directly.
- **Standards.** FASB and AICPA texts are paywalled or refused (fasb.org returned 403). Every summary of ASC or SSARS below is secondary.

Abbreviations: "§" = Internal Revenue Code (26 U.S.C.); "Reg." = Treasury Regulations (26 C.F.R.); "P.L. 119-21" = Public Law 119-21 of 4 July 2025 ("One Big Beautiful Bill Act"); "AFS" = applicable financial statement, §451(b)(3); "IC" = Indiana Code; "DOR" = Indiana Department of Revenue; "IB" = a DOR Income Tax Information Bulletin; "DLGF" = Indiana Department of Local Government Finance; "SOS" = Secretary of State; "CAT" = Ohio Commercial Activity Tax; "FTB" = California Franchise Tax Board; "R&TC" = California Revenue and Taxation Code; "ASC" = FASB Accounting Standards Codification; "ASU" = Accounting Standards Update; "SSARS" = AICPA Statements on Standards for Accounting and Review Services; "book" = the books kept for the lender (GAAP); "tax book" = the federal tax basis.

Entries run in the order the company meets them: setting up, first purchases, first sales, payments, quarter ends, later in the year, year end. Nothing here has been read by a practising accountant or attorney yet.

---

### CO-01 — We are incorporating in Indiana: what must the Articles say, when do we exist, and what keeps us in good standing?
- **Applies to:** corporation
- **The law:**
  - IC 23-1-21-2(a) (2015 text): the articles „must set forth: (1) a corporate name for the corporation that satisfies the requirements of IC 23-1-23-1; (2) the number of shares the corporation is authorized to issue; (3) the street address of the corporation's initial registered office in Indiana and the name of its initial registered agent at that office; and (4) the name and address of each incorporator.”
  - IC 23-1-21-3(a): „Unless a delayed effective date is specified, the corporate existence begins when the articles of incorporation are filed.”
  - Filing: Articles of Incorporation, State Form 4159, on paper or through INBiz. The SOS forms page lists the form but no fee. A search excerpt of the form shows a $100 paper fee (secondary — verify; forms.in.gov refused). Secondary sources give $95–$98 online (secondary — verify).
  - Business Entity Report, SOS page: „For-profit entities file every other year (biennial) in the anniversary month in which the entity was originally created.” Fees: „For Profit by Mail: $50”, „For Profit Online: $31”. Reports „can be submitted up to 90 [days] prior” to the month due. „After five months of a business entity report being overdue, the business will be administratively dissolved.”
  - The 2015 Code puts the biennial report in IC 23-1-53-3. Where the current Code puts it was not read.
  - INBiz, seen only as a search excerpt: the first report is due two years after formation, at $32 online (secondary — verify; it conflicts with the SOS page's $31).
- **Grade:** 1 plain
- **Rhea's default:** incorporation is one dated fact on the company: the state, the date existence begins (the filing date or the delayed effective date), authorised shares, registered agent and incorporators. The Business Entity Report becomes a clock duty in the anniversary month every second year, with a task opening 90 days before. Missing it ends in dissolution, which no later rule can undo.
- **Rhea today:**
  - `company` v7 has name, address, `legal_form` (`natural_person | legal_entity`) and `form_from`. It has no state of incorporation, existence date, registered agent or authorised shares.
  - The clock and `task_schedule` know `weekly | monthly | yearly | other`; there is no "every second year".
- **Where the work lives:** *pack data* (the incorporation door and rule, the biennial duty). A two-year cadence is a small *shared ground* addition to the clock.
- **Deadline:** Business Entity Report every two years in the anniversary month. If the company was formed in 2026, the first is due in that month of 2028 (computed here from the SOS rule and the INBiz statement — verify).
- **Watch:**
  - Indiana General Assembly, https://iga.in.gov. The session runs January to March or April; each act carries its own effective date.
  - SOS Business Services Division, https://www.in.gov/sos/business/ (fee schedule and forms).
  - Indiana Register, title 75 (Secretary of State rules), https://iar.iga.in.gov.
- **Sources:** https://statecodesfiles.justia.com/indiana/2015/title-23/article-1/chapter-21/chapter-21.pdf (unofficial 2015 copy); https://www.in.gov/sos/business/business-entity-reports ; https://www.in.gov/sos/business/division-forms/business-forms/ ; https://statecodesfiles.justia.com/indiana/2015/title-23/article-1/chapter-53/chapter-53.pdf

---

### CO-02 — Do we need an EIN before anything else, and must we report our owners to FinCEN?
- **Applies to:** everyone (a sole proprietor needs an EIN once it has employees)
- **The law:**
  - IRS, "Get an employer identification number" (reviewed 16 October 2026): „form your entity through the secretary of state before you apply for an EIN”. The applicant must be „the responsible party in control of the entity or its authorized representative”, with that person's SSN or ITIN. „You can apply for only 1 EIN per responsible party per day.” „You never have to pay a fee for an EIN.” Applicants who cannot apply online use phone, fax or mail (Form SS-4).
  - FinCEN: the interim final rule of 26 March 2025 „removes the requirement for U.S. companies and U.S. persons to report beneficial ownership information”. A final rule „making permanent the … BOI reporting exemptions … became effective on August 14, 2026”: „U.S. companies are exempt from BOI reporting requirements and therefore, are no longer required to file BOI reports.”
- **Grade:** 1 plain
- **Rhea's default:** the EIN is the company's federal identifier, recorded as a dated fact with the IRS assignment notice as its document. Every federal filing, every Indiana account (IB 11 prints the EIN on Form IT-6) and the payroll provider key on it. A domestic corporation has no BOI duty, so no task.
- **Rhea today:** the only tax identifier on `company` is `vat_id`, and the sales invoice carries `seller_nip` and `buyer_nip`. These are VAT-shaped names for what in the US is an EIN plus one account number per state agency.
- **Where the work lives:** *shared ground*. The company's identifiers are named for VAT (`vat_id`, `*_nip`). A market-neutral list of identifiers (scheme, value, valid-from) would serve Poland (NIP, REGON, KRS) and the US (EIN, Indiana DOR account, later the FTB entity ID) alike.
- **Deadline:** none in statute. In practice before the first payroll and the first federal deposit.
- **Watch:** IRS EIN page; FinCEN BOI page https://www.fincen.gov/boi ; Federal Register, FinCEN, https://www.federalregister.gov
- **Sources:** https://www.irs.gov/businesses/small-businesses-self-employed/get-an-employer-identification-number ; https://www.fincen.gov/boi

---

### CO-03 — Are we taxed as a C corporation, or should the two owners elect S status — and what in the books changes?
- **Applies to:** corporation
- **The law:**
  - A state-law corporation is a corporation for federal tax: Form 8832 instructions, "Domestic default rule". Form 8832 is for an "eligible entity" such as an LLC, not for an Indiana corporation.
  - The C corporation pays §11(b): „21 percent of taxable income.”
  - S status, §1361(b)(1): a „small business corporation” is a domestic corporation that does not „(A) have more than 100 shareholders, (B) have as a shareholder a person (other than an estate, a trust …, or an organization …) who is not an individual, (C) have a nonresident alien as a shareholder, and (D) have more than 1 class of stock.”
  - When the election is made, §1362(b)(1): „(A) at any time during the preceding taxable year, or (B) at any time during the taxable year and on or before the 15th day of the 3d month of the taxable year.”
  - A new corporation's first year, Reg. 1.1362-6(a)(2)(ii)(C): it „begins on the date that the corporation has shareholders, acquires assets, or begins doing business, whichever is the first to occur.”
  - Late relief: Rev. Proc. 2013-30, a Form 2553 filed „within 3 years and 75 days” of the intended effective date (Form 2553 instructions).
- **Grade:** 1 plain
- **Rhea's default:** the company declares its federal tax classification (`c_corporation | s_corporation`; an LLC may also be `partnership | disregarded`) with the date it takes effect and the IRS acceptance letter as evidence. Corporate income tax, estimates and deferred tax are gated on C. The S variant (CO-31) switches them off and turns on AAA and shareholder basis. Nothing in the log can infer the election, so it must be declared.
- **Rhea today:** `company` v7 already has the right shape: a declared, dated form that rules branch on (`books`, `income_tax`, `form_from`). The values are Polish: `income_tax: cit | skala | liniowy | ryczalt`, `books: full | kpir | ryczalt`.
- **Where the work lives:** *shared ground*. The declared-form enums on the shared `company` type are Poland's. Object types are versioned globally by name (`internal/pack/pack.go` skips a name and version that already exist), so a US `company` v8 would supersede Poland's in any installation that loads both. Each market needs its own declared values (a market-scoped profile, or enums a market may extend) without both editing one type.
- **Deadline:** Form 2553 no later than 2 months and 15 days after the first tax year begins (§1362(b)(1)(B); Reg. 1.1362-6(a)(2)(ii)(C)).
- **Watch:**
  - IRS Form 2553 and Form 8832 pages (form revisions).
  - Internal Revenue Bulletin for relief procedures (successors to Rev. Proc. 2013-30), https://www.irs.gov/irb.
  - Congress for §11 and §1361 changes, https://www.congress.gov.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/11 ; https://www.law.cornell.edu/uscode/text/26/1361 ; https://www.law.cornell.edu/uscode/text/26/1362 ; https://www.ecfr.gov/current/title-26/section-1.1362-6 ; https://www.irs.gov/pub/irs-pdf/i2553.pdf ; https://www.irs.gov/pub/irs-pdf/f8832.pdf

---

### CO-04 — Does any law make our books follow GAAP, or a prescribed chart of accounts?
- **Applies to:** everyone
- **The law:** No federal or Indiana statute prescribes GAAP or a chart of accounts for a private company.
  - §446(a): „Taxable income shall be computed under the method of accounting on the basis of which the taxpayer regularly computes his income in keeping his books.”
  - Reg. 1.446-1(a)(2): „no uniform method of accounting can be prescribed for all taxpayers. Each taxpayer shall adopt such forms and systems as are, in his judgment, best suited to his needs.” Also: „A method of accounting which reflects the consistent application of generally accepted accounting principles … will ordinarily be regarded as clearly reflecting income”.
  - Reg. 1.446-1(a)(4): the records include „a reconciliation of any differences between such books and his return.”
  - §6001: „Every person liable for any tax … shall keep such records”. Reg. 1.6001-1(a): „permanent books of account or records, including inventories”.
  - Indiana corporate law asks only for records and for statements on request.
    - IC 23-1-52-1(b) (2015 text): „A corporation shall maintain appropriate accounting records.”
    - IC 23-1-53-1(a): annual statements are prepared „On written request of any shareholder”, and „If financial statements are prepared for the corporation on the basis of generally accepted accounting principles, the annual financial statements must also be prepared on that basis.”
    - (b): without an accountant's report, an officer states „whether the statements were prepared on the basis of generally accepted accounting principles and, if not, describing the basis of preparation”.
  - In practice the basis is set by contract: the lender's covenant names GAAP or a "tax basis" special-purpose framework (secondary — verify against the actual loan agreement).
- **Grade:** 1 plain (no law mandates a basis or a chart)
- **Rhea's default:**
  - The company records the basis its lender requires as a fact, taken from the loan agreement.
  - Hoosier keeps its books on accrual GAAP (book `us-gaap`).
  - The federal tax basis is the same postings plus adjustment entries in a second book (`us-fed`), one adjustment rule per difference. Schedule M-1 then falls out as a view, and so does the reconciliation that Reg. 1.446-1(a)(4) asks for.
  - Indiana and, later, California keep only their own adjustments (CO-11).
- **Rhea today:** the kernel already supports this. `posting` v3 carries a free `book` string (`pl-stat`, `de-stat` exist), and `period_lock` is per book. No rule yet posts one event to two books, or posts adjustments that are only differences.
- **Where the work lives:** *pack data* (book names, adjustment rules, the M-1 view). The kernel needs nothing new.
- **Deadline:** —
- **Watch:** Federal Register, Treasury and IRS (changes to Reg. 1.446-1), https://www.federalregister.gov/agencies/internal-revenue-service ; Indiana General Assembly (IC 23-1-52/53).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/446 ; https://www.ecfr.gov/current/title-26/section-1.446-1 ; https://www.law.cornell.edu/uscode/text/26/6001 ; https://statecodesfiles.justia.com/indiana/2015/title-23/article-1/chapter-52/chapter-52.pdf ; https://statecodesfiles.justia.com/indiana/2015/title-23/article-1/chapter-53/chapter-53.pdf

---

### CO-05 — Which chart of accounts should the pack propose for a small distributor?
- **Applies to:** everyone
- **The law:**
  - None prescribes a chart (CO-04). What every chart must feed is fixed by the return.
  - Form 1120 instructions (2025) on Schedule L: „The balance sheets should agree with the corporation's books and records.” A corporation with cost of goods sold must „Complete and attach Form 1125-A”.
  - Form 1125-E is required where total receipts „are $500,000 or more”.
- **Grade:** 1 plain (the chart is custom, not law)
- **Rhea's default:** ship a four-digit chart as the company's own, each account carrying the Form 1120 line (page 1 or Schedule L) and the Form 1125-A line it maps to.
  - Ranges: 1000s assets, 2000s liabilities, 3000s equity, 4000s revenue, 5000s cost of goods sold, 6000s–7000s operating expenses. The 1000–7000 ranges are a common small-business convention (secondary — verify). Rhea proposes 8000s for other income and expense and 9000s for income taxes; no source sets those two.
  - Accounts year one needs:
    - assets: cash; receivables with an allowance; FIFO inventory; prepaid expenses; machinery and equipment (forklift, racking), vehicles and furniture, each with accumulated depreciation; the right-of-use asset;
    - liabilities: payables; accrued liabilities; the line of credit; the lease liability; federal and Indiana income tax payable; deferred tax; loans from shareholders;
    - equity: common stock, paid-in capital, retained earnings, dividends;
    - income and costs: sales; returns, allowances and rebates; cost of goods sold with freight-in; interest; current and deferred income tax expense.
- **Rhea today:** `account` v1 (`code`, `name`, `type`) is the right shape. It has no field for a return-line mapping. Poland's chart lives in `packs/pl/pack.json`.
- **Where the work lives:** *pack data*. One optional mapping field on `account` is a light *shared ground* change, and Poland's JPK_KR_PD tags need the same (CIT-08).
- **Deadline:** —
- **Watch:** IRS Form 1120 page (line changes each year), https://www.irs.gov/forms-pubs/about-form-1120 ; IRS draft forms, https://www.irs.gov/draft-tax-forms
- **Sources:** https://www.irs.gov/pub/irs-pdf/i1120.pdf ; https://www.ecfr.gov/current/title-26/section-1.446-1 ; https://beancount.io/it/blog/2026/04/03/chart-of-accounts-what-it-is-how-to-set-one-up (secondary — verify); https://www.patronaccounting.com/blog/chart-of-accounts-numbering-system (secondary — verify)

---

### CO-06 — May a C corporation use the cash method, and what does the $32 million gross receipts test unlock?
- **Applies to:** corporation (the sole proprietor applies the same test per business, §471(c)(3))
- **The law:**
  - §448(a): for a „C corporation … taxable income shall not be computed under the cash receipts and disbursements method”. The exception in (b)(3) applies if the entity „meets the gross receipts test of subsection (c)”.
  - §448(c)(1): „average annual gross receipts of such entity for the 3-taxable-year period ending with the taxable year which precedes such taxable year does not exceed $25,000,000”. That amount is indexed under (c)(4) and rounded to the nearest $1,000,000.
  - (c)(3)(A): an entity not in existence for the three years is tested on „the period during which such entity … was in existence”. (B) annualises short years. (C) reduces receipts by „returns and allowances”.
  - (c)(2) aggregates businesses under common control (§52, §414).
  - Rev. Proc. 2025-32 §4.30: for taxable years beginning in 2026 the test is met if the average „does not exceed $32,000,000.”
  - The same test unlocks the inventory exemption (§471(c)), the UNICAP exemption (§263A(i)) and the interest-limit exemption (§163(j)(3)).
- **Grade:** 1 plain. A first year has no prior year to average; I found no text that fails a new corporation on that account. Hoosier passes either way at about $3M.
- **Rhea's default:** accrual, because the lender reads GAAP books and the books keep inventory. Rhea computes the gross receipts test every year from the log: three-year average, annualised, net of returns, with the owners' other controlled businesses declared and added. It stores the result as a dated fact that the four exemptions read. The 2026 amount, $32,000,000, is a dated rate-table entry.
- **Rhea today:** folds and exact-rational formulas can average receipts. There is no fact for "businesses under common control", and no rate table with a validity period for US amounts.
- **Where the work lives:** *pack data*
- **Deadline:** the overall method is adopted on the first return (Reg. 1.446-1(e)(1)).
- **Watch:** the annual inflation revenue procedure in the Internal Revenue Bulletin (Rev. Proc. 2025-32 published the 2026 amounts in October 2025), https://www.irs.gov/irb
- **Sources:** https://www.law.cornell.edu/uscode/text/26/448 ; https://www.irs.gov/pub/irs-drop/rp-25-32.pdf ; https://www.ecfr.gov/current/title-26/section-1.448-2

---

### CO-07 — Which choices does our first return lock in, and where do we record them?
- **Applies to:** everyone
- **The law:** Reg. 1.446-1(e)(1): „A taxpayer filing his first return may adopt any permissible method of accounting”. Later changes need consent (Form 3115). Choices made on or with Hoosier's first return:
  - **Methods, kept until changed with consent:**
    - the overall method (CO-06) and the inventory method (CO-08);
    - the recurring item exception, Reg. 1.461-5(a): a taxpayer „may adopt the recurring item exception … as method of accounting for one or more types of recurring items”;
    - the §451(c) deferral election for advance payments, which „shall be treated as a method of accounting” (§451(c)(2)(B));
    - the bad-debt method (Reg. 1.166-1(b)).
  - **Annual elections, made again each year:**
    - the de minimis safe harbor, Reg. 1.263(a)-1(f)(5): a statement titled „Section 1.263(a)-1(f) de minimis safe harbor election” attached to „the taxpayer's timely filed original Federal tax return (including extensions)”; „A taxpayer may not revoke an election”;
    - the bonus election-out, §168(k)(7): „with respect to any class of property for any taxable year”; it „may be revoked only with the consent of the Secretary”;
    - §179 expensing (Form 4562).
  - **One-off elections:** S status (CO-03).
- **Grade:** 1 plain
- **Rhea's default:** a register of elections. Each election is a dated fact (kind, tax year, scope such as an asset class, who decided, and the statement filed as its document), and rules read it instead of assuming. Method-type elections stay in force until a change-of-method fact; annual ones raise a task before each return is due. The agent may propose an election with its warrant, but only a person makes it.
- **Rhea today:** rules have `effective_from` and a `warrant`, and the gate records approval. An election, though, is a fact about the company that rules branch on, not a rule. Today only `company.form_from` comes close.
- **Where the work lives:** *shared ground*. A generic dated election declared through a door is the same shape Poland needs for the Estonian CIT (CIT-01), the VAT cash method (VAT-04) and the form of tax (FORM-04).
- **Deadline:** with the timely filed 2026 return, including extensions (15 October 2027 if extended).
- **Watch:** IRS Form 1120 and 3115 instructions; Federal Register for amendments to Reg. 1.263(a)-1 and 1.461-5; https://www.irs.gov/irb for automatic-change procedures.
- **Sources:** https://www.ecfr.gov/current/title-26/section-1.446-1 ; https://www.ecfr.gov/current/title-26/section-1.461-5 ; https://www.ecfr.gov/current/title-26/section-1.263(a)-1 ; https://www.law.cornell.edu/uscode/text/26/168 ; https://www.law.cornell.edu/uscode/text/26/451

---

### CO-08 — We buy stock for the warehouse: must we keep inventories, at what cost, and do the UNICAP rules apply?
- **Applies to:** everyone with stock
- **The law:**
  - §471(a): inventories are required where „necessary in order clearly to determine the income”.
  - Small business exemption, §471(c)(1): a taxpayer meeting the §448(c) test may treat inventory „as non-incidental materials and supplies”, or follow its AFS „or, if the taxpayer does not have any applicable financial statement …, the books and records of the taxpayer prepared in accordance with the taxpayer's accounting procedures”.
  - Reg. 1.471-1(b)(6)(ii) limits the books method: „a taxpayer may not recover any costs that have not been paid or incurred under the taxpayer's overall method … inventory costs must satisfy the all events test, including economic performance”.
  - §263A(i)(1): for a taxpayer meeting the §448(c) test, „this section shall not apply”. Hoosier is outside UNICAP.
  - Valuation, Reg. 1.471-2(c): „(1) cost and (2) cost or market, whichever is lower”.
  - Reg. 1.471-2(d): intermingled goods „will be deemed to be the goods most recently purchased” (FIFO). „Book inventories … should be verified by physical inventories at reasonable intervals”.
  - LIFO (note only): the conformity rule, Reg. 1.472-2(e)(1), forbids any other method „for credit purposes or for purposes of reports to shareholders”. LIFO for tax would mean LIFO in the lender's statements.
  - Books: ASC 330; FIFO stock is measured at the lower of cost and net realisable value (ASU 2015-11) (secondary — verify).
- **Grade:** 2 settled — the books-and-records option is settled by Reg. 1.471-1(b)
- **Rhea's default:**
  - Perpetual FIFO at cost in one stock ledger for both books: one cost layer per receipt (net invoice price plus freight-in), consumed in order by issues.
  - Cost of goods sold posts from the layers at each issue.
  - The year-end count adjusts quantities (CO-27).
  - Regular §471 (not the §471(c) materials-and-supplies option), because the books already do it and it keeps tax equal to the books.
  - §263A off while the gross receipts test holds.
  - LIFO refused unless a person declares it, with the conformity consequence stated.
- **Rhea today:** the warehouse definitions count quantities by item and location (`goods_receipt`, `goods_issue`, `stock_movement`). Poland's PZ values receipts (Wn 330 / Ma 300). No issue is valued, no cost layer exists and no rule posts cost of goods sold. Poland's FB-19 asks for the same.
- **Where the work lives:** *shared ground*. FIFO cost layers consumed by issues are a stateful cost-flow capability (in the warehouse definitions or the fold engine) that every market with stock needs. The US valuation choices themselves are *pack data*.
- **Deadline:** the inventory at 31 December (count close to year end).
- **Watch:** Federal Register (§471/§263A regulations); FASB ASUs on Topic 330, https://www.fasb.org/standards/accounting-standards-updates (refused to scripts; read in a browser).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/471 ; https://www.law.cornell.edu/uscode/text/26/263A ; https://www.ecfr.gov/current/title-26/section-1.471-1 ; https://www.ecfr.gov/current/title-26/section-1.471-2 ; https://www.ecfr.gov/current/title-26/section-1.472-2

---

### CO-09 — The forklift, van, racking and small tools: which purchases are assets and which are expenses?
- **Applies to:** everyone
- **The law:**
  - De minimis safe harbor, Reg. 1.263(a)-1(f)(1):
    - With an AFS, the amount must „not exceed $5,000 per invoice (or per item as substantiated by the invoice)”, and the taxpayer needs „at the beginning of the taxable year written accounting procedures treating as an expense”.
    - Without an AFS, the regulation's „$500 … or other amount as identified in published guidance”. Notice 2015-82 raised it: „the limitation for a taxpayer without an AFS is increased from $500 to $2,500” (taxable years beginning on or after 1 January 2016). I found no later increase.
    - Not for „property that is or is intended to be included in inventory” ((f)(2)(i)).
  - An AFS is, among others, „an audited financial statement of the taxpayer which is used for— (I) credit purposes” (§451(b)(3)(A)(ii)). A compiled or reviewed statement is not audited, so a lender asking for review keeps Hoosier at $2,500.
- **Grade:** 1 plain
- **Rhea's default:**
  - A capitalisation policy, adopted as a dated fact before the first day of the tax year: expense items up to $2,500 per invoice or item, or $5,000 from the first year the lender requires audited statements.
  - Above the threshold, a purchase line becomes a fixed-asset candidate (a task for a person to confirm the class); below it, an expense.
  - The annual election statement is a clock task for each return.
  - Stock for resale never goes through this path.
- **Rhea today:** `purchase_invoice` v6 is VAT-shaped (`net`, `vat`, `gross`, `vat_rate`) and has no lines. The Poland pack's `fixed_asset` v1 carries Polish fields (`kst`, `amortization_type`, `coefficient`), and no rule creates one.
- **Where the work lives:** *pack data* (the policy, the asset-candidate rule). Purchase lines, needed here and for sales tax, are *shared ground* (see CO-13).
- **Deadline:** written policy in place by the first day of the tax year; election with the 2026 return.
- **Watch:** Internal Revenue Bulletin (the $2,500 changes only „in published guidance in the Federal Register or in the Internal Revenue Bulletin”, per the regulation), https://www.irs.gov/irb
- **Sources:** https://www.ecfr.gov/current/title-26/section-1.263(a)-1 ; https://www.irs.gov/pub/irs-drop/n-15-82.pdf ; https://www.law.cornell.edu/uscode/text/26/451

---

### CO-10 — How do we depreciate for federal tax now that P.L. 119-21 restored 100% bonus depreciation?
- **Applies to:** everyone
- **The law:**
  - Bonus, §168(k)(1)(A) as amended by P.L. 119-21 §70301(b)(1)(A): the first-year allowance is „100 percent of the adjusted basis of the qualified property”.
    - Effective date, P.L. 119-21 §70301(c)(1): it applies „to property acquired after January 19, 2025”. A property is not acquired after the date a written binding contract is signed ((c)(4)). The election out, §168(k)(7), is per „class of property for any taxable year”.
    - Notice 2026-11 is interim guidance. Taxpayers may rely on it, and on Reg. 1.168(k)-2 with substituted dates, until proposed regulations are published. I found none in the Federal Register as of 16 October 2026 (unverified that none exist).
  - Classes, Pub. 946 Table B-2:
    - asset class 57.0 „Distributive Trades and Services: Includes assets used in wholesale and retail trade”, class life 9, GDS 5 years, ADS 9 (forklift, racking);
    - Table B-1 00.241 „Light General Purpose Trucks … (actual weight less than 13,000 pounds)”, 5 years (the van);
    - 00.11 office furniture, 7 years.
  - §179, Rev. Proc. 2025-32 §4.24: for 2026 „$2,560,000”, reduced by cost above „$4,090,000”; SUVs capped at „$32,000”. The deduction is limited to taxable income from the active business (§179(b)(3)).
  - The van: if it is a passenger automobile under §280F, Rev. Proc. 2026-15 caps first-year depreciation at $20,300 with bonus and $12,300 without. Whether Hoosier's cargo van is a qualified non-personal-use vehicle outside the cap was not checked.
- **Grade:** 2 settled — the statute is plain; its application is settled for now by Notice 2026-11, pending regulations
- **Rhea's default:**
  - Each asset records cost, placed-in-service date, asset class, recovery period, convention, and one depreciation line per book.
  - Book: straight line over the policy useful life.
  - Federal: 100% bonus unless a class is elected out (CO-07), with §179 elected on the first $25,000 of eligible cost. §179 costs nothing federally when bonus is 100%, but it is the only immediate deduction Indiana allows (CO-11). This is Rhea's proposal, not a source.
  - The half-year convention, or mid-quarter when more than 40% of the year's basis is placed in service in the last quarter (Pub. 946), still drives the book-like state and elected-out schedules.
  - Form 4562 is rendered from the federal book.
- **Rhea today:** no depreciation rules. The clock (monthly) can post depreciation, the formula language holds MACRS percentages exactly, and postings carry a book.
- **Where the work lives:** *shared ground*. An asset register in which one asset carries several depreciation books: book, federal, each state. The Poland pack's `fixed_asset` hard-codes two (`prior_tax_depreciation`, `prior_balance_depreciation`) and Polish classification fields. The MACRS tables and the class list are *pack data*.
- **Deadline:** Form 4562 with the return for the first year an asset is depreciated.
- **Watch:**
  - Federal Register for the §168(k) proposed regulations Notice 2026-11 announces, https://www.federalregister.gov/agencies/internal-revenue-service.
  - The autumn inflation revenue procedure (§179) and the early-year §280F revenue procedure (Rev. Proc. 2026-15), both in https://www.irs.gov/irb.
  - Form 4562 page, https://www.irs.gov/forms-pubs/about-form-4562.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/168 ; https://www.irs.gov/pub/irs-drop/n-26-11.pdf ; https://www.law.cornell.edu/uscode/text/26/179 ; https://www.irs.gov/pub/irs-drop/rp-25-32.pdf ; https://www.irs.gov/pub/irs-pdf/p946.pdf ; https://www.irs.gov/pub/irs-drop/rp-26-15.pdf ; https://www.irs.gov/pub/irs-pdf/i4562.pdf

---

### CO-11 — Do Indiana, and later California, accept the federal bonus and §179 deductions?
- **Applies to:** everyone (Indiana); corporation (California Form 100)
- **The law:**
  - Indiana, IB 118 (May 2026, effective 4 July 2025, retroactive): „Since 2002, Indiana has largely decoupled from the federal allowances for bonus depreciation and Section 179 expensing.”
    - §179: „Indiana caps the overall Indiana Section 179 Allowance at $25,000 … In the first year, the difference between the Federal Section 179 Allowance and the Indiana Section 179 Allowance is an add-back”, recovered in later years (Code 105).
    - Bonus: the adjustment is „the federal depreciation claimed on the property minus depreciation that would have been computed if an election had been made to not apply the bonus depreciation”. This gives an add-back in the first year and subtractions after (Code 104).
    - SEA 243 (2026) extends the decoupling to §168(n) qualified production property.
    - One slip to verify: the bulletin says Indiana follows „the $2,500,000 threshold for phase-out under IRC Section 179(b)(2)”. In the statute $2,500,000 is the (b)(1) limit; the (b)(2) phase-out starts at $4,000,000 ($4,090,000 for 2026). Read it as a slip and ask DOR.
  - California, Form 100 Booklet 2025: California does not conform to „IRC Section 168(k) relating to the depreciation deduction for certain assets” or „The enhanced IRC Section 179 expensing election”. The differences are computed on form FTB 3885, and on depreciation „California law is substantially different from federal law for corporations”.
- **Grade:** 1 plain
- **Rhea's default:** from the first asset, every asset keeps an Indiana depreciation line: federal MACRS with bonus off, and §179 capped at $25,000. The IT-20 add-back and subtraction is the difference between the federal and Indiana lines. A California line opens in the first California year (2027) under FTB 3885 rules.
- **Rhea today:** as CO-10. No state books.
- **Where the work lives:** *shared ground* (the multi-book asset register, CO-10), with the Indiana and California rules as *pack data*.
- **Deadline:** with each state return (CO-20, CO-23).
- **Watch:**
  - DOR IB 118, revised with each conformity act, https://www.in.gov/dor/legal-resources/tax-library/information-bulletins/income-tax-information-bulletins/.
  - Indiana General Assembly (the annual Internal Revenue Code conformity bill), https://iga.in.gov.
  - FTB legal rulings and notices, https://www.ftb.ca.gov/tax-pros/law/legal-rulings/index.html.
- **Sources:** https://www.in.gov/dor/files/reference/ib118.pdf ; https://www.ftb.ca.gov/forms/2025/2025-100-booklet.pdf

---

### CO-12 — We lease the warehouse: what goes on the balance sheet, and what does tax deduct?
- **Applies to:** everyone with a lease (ASC 842 binds only GAAP books)
- **The law:**
  - Books: ASC 842. The lessee recognises a right-of-use asset and a lease liability; an operating lease has a single, generally straight-line lease cost. Leases of 12 months or less may be left off the balance sheet by policy, and a private company may elect a risk-free discount rate. Private companies apply it for fiscal years beginning after 15 December 2021 (ASU 2020-05). All secondary — verify.
  - Tax, §162(a)(3), deduction for „rentals or other payments required to be made as a condition to the continued use or possession … of property to which the taxpayer has not taken or is not taking title or in which he has no equity.”
  - Accrual timing, §461(h)(2)(A)(iii): for „the use of property by the taxpayer, economic performance occurs as the taxpayer uses such property.”
  - §467 rental agreement:
    - (d)(1)(B) covers an agreement under which „there are increases in the amount to be paid as rent”;
    - (d)(2) excludes agreements whose total consideration „does not exceed $250,000”;
    - (b)(1) then makes rent accrue as the agreement allocates it, except for disqualified leasebacks and long-term agreements.
- **Grade:** 2 settled — the tax text is plain and §467 settles escalating rents; the book side rests on ASC 842 (secondary)
- **Rhea's default:**
  - At commencement, the book records the right-of-use asset and the lease liability, discounted at the rate the lease's policy fact names (rate implicit, incremental borrowing rate, or the private-company risk-free election). Lease cost is straight line.
  - The tax book deducts rent as the lease's rent schedule allocates it to each period of use (§467 when the lease escalates and exceeds $250,000).
  - The difference is an M-1 item and a deferred tax balance. The lease contract is a document with its schedule as data.
- **Rhea today:** the Poland pack has `lease_contract` v1, shaped for Polish car leases (`vat_deduction_limit`, `vat_deducted`), with no rules. Present-value arithmetic over a payment schedule should fit the exact-rational formula language (not checked).
- **Where the work lives:** *pack data*
- **Deadline:** —
- **Watch:** FASB ASUs on Topic 842 (and Private Company Council alternatives), https://www.fasb.org/standards/accounting-standards-updates ; Federal Register (§467 regulations).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/162 ; https://www.law.cornell.edu/uscode/text/26/461 ; https://www.law.cornell.edu/uscode/text/26/467 ; https://dart.deloitte.com/USDART/home/publications/archive/deloitte-publications/heads-up/2020/fasb-defers-effective-dates-revenue-leases (secondary)

---

### CO-13 — We ship an order: in which year is the sale income for tax, and is the invoice date what counts?
- **Applies to:** everyone on accrual
- **The law:**
  - §451(a): income is included „for the taxable year in which received …, unless, under the method of accounting used …, such amount is to be properly accounted for as of a different period.”
  - §451(b)(1)(C): the all events test is met when „all the events have occurred which fix the right to receive such income and the amount of such income can be determined with reasonable accuracy.”
  - AFS rule, (b)(1)(A): the test is met no later than when the item „is taken into account as revenue in— (i) an applicable financial statement”. By (B)(i) it does not apply „to a taxpayer which does not have a financial statement described in clause (i) or (ii)”.
  - Reg. 1.446-1(c)(1)(ii)(C): a business „may account for sales of the taxpayer's product when the goods are shipped, when the product is delivered or accepted, or when title to the goods passes to the customers, whether or not billed, depending on the method regularly employed in keeping the taxpayer's books.”
- **Grade:** 1 plain
- **Rhea's default:**
  - A sale is income on shipment or delivery, as the sale's terms say (shipping point or destination).
  - That date is the one the book uses for revenue under ASC 606 (CO-14), so the two agree.
  - The invoice date does not decide.
  - While Hoosier has no audited statement, §451(b) adds nothing. If the lender later requires an audit, the AFS rule applies from that year (CO-30).
- **Rhea today:** `sales_invoice` v7 has `issue_date`, `sale_date` (the VAT "data sprzedaży"), `vat_period`, `net`/`vat`/`gross`, `vat_rate`, `seller_nip`/`buyer_nip` and no lines. It has no shipment or delivery date, no shipping terms and no ship-to address, and revenue posts from the invoice. `goods_issue` records quantity, not the sale it fulfils.
- **Where the work lives:** *shared ground*. The sales and purchase documents are VAT-shaped. A US sale has no VAT, and its revenue date, apportionment state (CO-20) and sales-tax situs (another section) come from the shipment and its terms. The documents need lines, a ship-to, and a delivery fact the revenue rule reads.
- **Deadline:** —
- **Watch:** Federal Register (Reg. 1.451-3 and §451 changes), https://www.federalregister.gov/agencies/internal-revenue-service
- **Sources:** https://www.law.cornell.edu/uscode/text/26/451 ; https://www.ecfr.gov/current/title-26/section-1.446-1 ; https://www.ecfr.gov/current/title-26/section-1.451-1

---

### CO-14 — In the books, when is a sale revenue, and what do customer rebates and returns do to it?
- **Applies to:** corporation (GAAP books)
- **The law:**
  - Books: ASC 606 (Revenue from Contracts with Customers), effective for private companies for annual periods beginning after 15 December 2018, optionally deferred to 2019 by ASU 2020-05. Revenue for goods is recognised when control passes. Rebates, returns and prompt-pay discounts are variable consideration, estimated and constrained (secondary — verify).
  - Tax, Reg. 1.461-4(g)(3): for „a rebate, refund, or similar payment to another person … economic performance occurs as payment is made to the person to which the liability is owed.”
- **Grade:** 2 settled — the book side by ASC 606 (secondary); the tax side by Reg. 1.461-4(g)(3)
- **Rhea's default:**
  - In the book, a sale's revenue is reduced at the time of sale by expected volume rebates and returns: a refund liability, with a returns asset at cost.
  - The tax book holds back the rebate accrual until the rebate is paid or credited.
  - Estimates are a declared policy (rate per customer programme), never guessed by the agent from data.
- **Rhea today:** no rebate or return rules. Polish corrections (VAT-27) are invoice-shaped.
- **Where the work lives:** *pack data*
- **Deadline:** —
- **Watch:** FASB ASUs on Topic 606; Federal Register (Reg. 1.461-4).
- **Sources:** https://www.ecfr.gov/current/title-26/section-1.461-4 ; https://dart.deloitte.com/USDART/home/publications/archive/deloitte-publications/heads-up/2020/fasb-defers-effective-dates-revenue-leases (secondary)

---

### CO-15 — A customer pays a deposit before we ship: is it income when received?
- **Applies to:** everyone on accrual
- **The law:**
  - §451(c)(1): an accrual taxpayer includes an advance payment in the year received or, if it elects, includes only the portion „required under subsection (b)” that year and „the remaining portion of such advance payment in gross income in the taxable year following the taxable year in which such payment is received.” The election is a method of accounting ((c)(2)(B)).
  - Without an AFS, Reg. 1.451-8(d)(3)(i): the taxpayer „includes the advance payment in gross income for the taxable year of receipt to the extent that it is earned in that taxable year and includes the remaining portion … in the next succeeding taxable year.” Earned means the all events test „without regard to when the amount is received” ((d)(3)(ii)).
- **Grade:** 1 plain
- **Rhea's default:**
  - A deposit is a contract liability in the book.
  - For tax, Rhea proposes the non-AFS deferral method on the first return in which a deposit arrives. The deposit is income in the year received only for the goods shipped that year; the rest is income next year at the latest, shipped or not.
  - The election is an entry in the elections register (CO-07).
- **Rhea today:** the matching engine already pairs payments with documents. Poland's prepayment rules are VAT-driven (VAT-19).
- **Where the work lives:** *pack data*
- **Deadline:** the election on the first return in which an advance payment is received.
- **Watch:** Federal Register (Reg. 1.451-8).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/451 ; https://www.ecfr.gov/current/title-26/section-1.451-8

---

### CO-16 — A customer will not pay: an allowance in the books, a deduction for tax — when?
- **Applies to:** everyone
- **The law:**
  - Tax, §166(a)(1): „There shall be allowed as a deduction any debt which becomes worthless within the taxable year.” (a)(2), partial worthlessness: „in an amount not in excess of the part charged off within the taxable year”.
  - The reserve method was repealed: „[(c) Repealed. Pub. L. 99–514, title VIII, § 805(a), Oct. 22, 1986]”. Reg. 1.166-1(a)(2) still mentions „a reasonable addition to a reserve”; that text is stale against the statute.
  - Evidence, Reg. 1.166-2(a): „all pertinent evidence, including the value of the collateral, if any, securing the debt and the financial condition of the debtor”. (b): legal action is not required. Recovery is income (Reg. 1.166-1(f)).
  - Books: ASC 326 (expected credit losses) requires an allowance on trade receivables. ASU 2025-05 lets an entity assume balance-sheet-date conditions persist for current receivables. Private companies may also consider collections after the balance-sheet date. Both apply to annual periods beginning after 15 December 2025, so from 2026 (secondary — verify).
- **Grade:** 2 settled — when a debt is worthless is a question of facts, settled by Reg. 1.166-2's evidence rules
- **Rhea's default:**
  - The book carries an allowance by ageing bucket, with the rates as a declared policy (and the ASU 2025-05 elections as declared facts). It never enters the tax book.
  - The tax deduction is per invoice, when a person records it wholly or partly worthless with the evidence as a document, and the book charges it off at the same moment.
  - A later payment on a written-off invoice is income.
- **Rhea today:** the matching engine knows each invoice's open amount and age. Poland's bad-debt relief (VAT-30) is VAT-shaped. No allowance rule exists.
- **Where the work lives:** *pack data*
- **Deadline:** the deduction belongs to the year the debt becomes worthless, not the year a person notices.
- **Watch:** FASB ASUs on Topic 326; Federal Register (Reg. 1.166).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/166 ; https://www.ecfr.gov/current/title-26/section-1.166-1 ; https://www.ecfr.gov/current/title-26/section-1.166-2 ; https://kpmg.com/us/en/frv/reference-library/2025/fasb-issues-asu-simplified-credit-loss-guidance.html (secondary)

---

### CO-17 — Is all the interest on our line of credit deductible?
- **Applies to:** everyone with business debt
- **The law:**
  - §163(j)(3): the business-interest limit „shall not apply” to a taxpayer that „meets the gross receipts test of section 448(c)” (other than a tax shelter). For 2026 that test is $32,000,000 (CO-06).
  - P.L. 119-21 §70303 restored the depreciation add-back in adjusted taxable income (§163(j)(8)(A)(v)) „to taxable years beginning after December 31, 2024”. §70341 changed (j) further for taxable years beginning after 31 December 2025, including that business interest excludes interest capitalised under §263(g) or §263A(f).
- **Grade:** 1 plain
- **Rhea's default:** Hoosier is exempt while it passes the gross receipts test. Interest is deducted as it accrues, from the lender's statements. If the test fails, Rhea raises a task, because the §163(j) computation is not in the first-year pack.
- **Rhea today:** bank statements arrive through the MT940 reader and the matching engine. Nothing tests §163(j).
- **Where the work lives:** *pack data*
- **Deadline:** —
- **Watch:** Congress; Federal Register (§163(j) regulations).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/163

---

### CO-18 — The owners put money in and take money out: capital, loans, dividends, salaries
- **Applies to:** corporation
- **The law:**
  - Dividends, §316(a): „any distribution of property made by a corporation to its shareholders … out of its earnings and profits”. „Every distribution is made out of earnings and profits to the extent thereof”. Form 1099-DIV is filed for each person paid „$10 or more” (Instructions for Form 1099-DIV, Rev. January 2024).
  - Loans, §7872(c)(1)(C): the below-market loan rules apply to „Any below-market loan directly or indirectly between a corporation and any shareholder”. (c)(3) leaves out any day on which outstanding loans between the two „does not exceed $10,000.”
  - Salaries, §162(a)(1): deduction for „a reasonable allowance for salaries or other compensation for personal services actually rendered”. For a C corporation the risk is pay too high, treated as a dividend; for an S corporation, pay too low (CO-31).
  - Retained profits: §531 imposes 20% on accumulated taxable income beyond reasonable business needs, with a minimum credit of „$250,000” less accumulated earnings and profits (§535(c)).
  - Whether an owner's advance is debt or equity is decided by courts on many factors (not fetched here — verify).
- **Grade:** 2 settled — the statute is plain; debt versus equity is settled by case law
- **Rhea's default:**
  - A shareholder register: each owner, shares and date.
  - Separate accounts per owner: paid-in capital, loans to the company, loans from the company, dividends declared.
  - A loan exists only with a note as a document (principal, rate, schedule). A loan carrying less than the applicable federal rate over $10,000 raises a §7872 task.
  - A distribution is a dividend up to earnings and profits, which the tax book tracks. It raises a 1099-DIV task in January.
  - Owners' wages come through the payroll provider's journal like anyone's.
- **Rhea today:** the Poland pack has `company.owner` and `owner_account`, and `packs/common/people.json` gives persons. There is no shareholder register.
- **Where the work lives:** *pack data*
- **Deadline:** 1099-DIV for distributions of $10 or more (the due dates are not checked here).
- **Watch:** IRS monthly applicable federal rates revenue rulings in https://www.irs.gov/irb ; Form 1099-DIV instructions.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/316 ; https://www.law.cornell.edu/uscode/text/26/7872 ; https://www.law.cornell.edu/uscode/text/26/162 ; https://www.law.cornell.edu/uscode/text/26/531 ; https://www.law.cornell.edu/uscode/text/26/535 ; https://www.irs.gov/pub/irs-pdf/i1099div.pdf

---

### CO-19 — Quarterly federal estimated tax: which dates, how much, and is there a safe harbor in our first year?
- **Applies to:** corporation
- **The law:**
  - §6655(c): four installments. For a calendar year they fall on April 15, June 15, September 15 and December 15 (Pub. 542: „the 15th day of the 4th, 6th, 9th, and 12th months”). A weekend or holiday moves a date to the next business day (Pub. 542).
  - §6655(d)(1): each is „25 percent of the required annual payment”, the lesser of „100 percent of the tax shown on the return for the taxable year” and „100 percent of the tax shown on the return … for the preceding taxable year”. But „Clause (ii) shall not apply if the preceding taxable year was not a taxable year of 12 months, or the corporation did not file a return for such preceding taxable year showing a liability for tax.” In its first year the company has no prior-year safe harbor.
  - Large corporations, (d)(2): a corporation with taxable income of „$1,000,000 or more” in any of the three prior years may use last year's tax only „for purposes of determining the amount of the 1st required installment”.
  - (e) allows the annualised income installment.
  - (f): no penalty if the year's tax „is less than $500.”
  - The worksheet is now in Pub. 542: „Form 1120-W … and the Instructions for Form 1120-W are now historical. The 2022 revisions were the last revisions”.
  - The penalty is computed on Form 2220. Payments go through EFTPS (Pub. 542).
- **Grade:** 1 plain
- **Rhea's default:**
  - Four dated duties on the clock.
  - Each installment is the annualised-income amount from the tax book's year to date (§6655(e)), because the first year has no prior-year safe harbor and Hoosier's income will grow through the year.
  - Each payment is an event matched to its bank debit.
  - Form 2220 is computed at year end from the four.
  - A short first year moves the dates (not covered here).
- **Rhea today:** the clock fires dated duties (`yearly` works for four fixed dates) but does not move a date off a weekend or holiday. `tax_declaration.kind` lists only Polish declarations.
- **Where the work lives:** *pack data* (rules and duties). Moving a deadline to the next business day is *shared ground*, and Polish deadlines need it too.
- **Deadline:** 2026: 15 April (Wednesday), 15 June (Monday), 15 September (Tuesday), 15 December (Tuesday). Weekdays computed here.
- **Watch:** Pub. 542 and Form 2220 instructions, https://www.irs.gov/forms-pubs/about-publication-542 ; quarterly underpayment-rate news releases (§6621); Congress (§6655).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6655 ; https://www.irs.gov/pub/irs-pdf/p542.pdf ; https://www.irs.gov/pub/irs-pdf/i2220.pdf

---

### CO-20 — Indiana's corporate income tax: at what rate, on which share of our income, and when is the return due?
- **Applies to:** corporation
- **The law:**
  - Rate, IB 12: 4.9% „After June 30, 2021”. IB 11: „the adjusted gross income tax rate is 4.9% of Indiana adjusted gross income after apportionment.” IB 12 lists no later step. The 2025 and 2026 acts were not searched for a rate change (IB 118, May 2026, mentions none).
  - Apportionment, IB 12: a corporation „that has business income from both within and outside Indiana must apportion its income by means of the single-factor receipts formula under IC 6-3-2-2”. „Destination sales by an Indiana seller to a purchaser outside Indiana (other than the U.S. Government) will not be included in the numerator of the sales factor.”
  - Indiana has no throwback for goods: commentary reports its repeal from 2016 (secondary — verify). IC 6-3-2-2 itself was not read.
  - Doing business includes „Maintenance of an office, a warehouse”.
  - The starting point is federal taxable income, with Indiana's add-backs (bonus and §179, CO-11). Other modifications, such as the add-back of state income taxes, were not checked.
  - Return, IB 12: IT-20 is due „the 15th day of the fifth month following the close of the taxable year”. A federal Form 7004 extension is accepted (IB 11: returns „posted within one month after the last date indicated on the federal extension form will be considered filed on a timely basis”), but 90% of the tax must be prepaid by the original date.
  - E-filing is required above $1,000,000 of federal gross income (2025 IT-20 booklet, seen only as a search excerpt; forms.in.gov refused — verify).
- **Grade:** 2 settled. Out-of-state destination sales leave the numerator without any condition that the company be taxable there; this is settled by DOR IB 12 (statute text not reached).
- **Rhea's default:**
  - Indiana adjusted gross income = federal taxable income from the tax book + Indiana depreciation differences.
  - Apportioned by Indiana receipts ÷ all receipts, each sale sitused by its ship-to state.
  - Taxed at 4.9%.
  - The Ohio and California sales leave the Indiana numerator from the first sale.
- **Rhea today:** no ship-to state on a sale (CO-13), no state tax.
- **Where the work lives:** *pack data*, once the shared ship-to fact exists (CO-13).
- **Deadline:** IT-20 for 2026 due 15 May 2027. That is a Saturday; the move to Monday 17 May is assumed, not checked.
- **Watch:**
  - DOR information bulletins and departmental notices, https://www.in.gov/dor/legal-resources/tax-library/.
  - Indiana General Assembly (rate in IC 6-3-2-1; annual conformity), https://iga.in.gov.
- **Sources:** https://www.in.gov/dor/files/ib12.pdf ; https://www.in.gov/dor/files/ib11.pdf ; https://secure.in.gov/dor/files/gb303.pdf ; https://www.klcpas.com/?p=1862 (secondary)

---

### CO-21 — Indiana estimated payments: threshold, dates, and whether a first-year corporation owes them
- **Applies to:** corporation
- **The law:**
  - IB 11: corporations „must make estimated payments if the adjusted gross income tax exceeds $2,500 for the taxable year”. Payments are „due no later than the twentieth day of the fourth, sixth, ninth and twelfth month of the taxpayer's taxable year”. Electronic payment is required when the payment „exceeds $5,000”, and an INTIME payment fulfils that.
  - Penalty: 10%, avoided by quarterly payments equal to „25% of the final tax liability for the current year, or 25% of the corporation's liability for the previous tax year” (IB 12), or the §6655(e) annualised amount.
  - For short years „the penalty shall be computed in a manner consistent with Section 6655” (IB 11).
  - A first-year corporation „no estimated tax payment is required”: this is from a search excerpt of the 2025 IT-20 booklet only, unverified (forms.in.gov refused).
- **Grade:** 2 settled by DOR, if the booklet excerpt is accurate. Read against IB 11's reference to §6655, which has no first-year safe harbor, it needs a direct read before shipping.
- **Rhea's default:** pay Indiana estimates from the first year, on the 20th, at 25% of projected Indiana tax (annualised), always through INTIME so the $5,000 electronic rule never bites. If the booklet's first-year exemption is confirmed, the only cost of paying anyway is cash paid early.
- **Rhea today:** as CO-19.
- **Where the work lives:** *pack data*
- **Deadline:** 2026: 20 April (Monday), 20 June (Saturday), 20 September (Sunday), 20 December (Sunday). Weekdays computed here; Indiana's weekend rule not checked.
- **Watch:** DOR IB 11 and the yearly IT-20 booklet; Indiana General Assembly (IC 6-3-4-4.1).
- **Sources:** https://www.in.gov/dor/files/ib11.pdf ; https://www.in.gov/dor/files/ib12.pdf ; https://forms.in.gov/Download.aspx?id=16945 (not reached)

---

### CO-22 — Our Ohio sales cross Ohio's nexus line: does the corporation owe an Ohio tax?
- **Applies to:** everyone selling into Ohio
- **The law:**
  - Ohio Department of Taxation, Information Release CAT 2023-01 (21 August 2023), after Am. Sub. H.B. 33:
    - „Effective for tax periods beginning on and after January 1, 2024, the CAT annual minimum tax is eliminated.”
    - „For calendar year 2025 and thereafter, the exclusion amount is increased to $6 million.”
    - „Taxpayers with … $6 million or less in 2025 will not be required to file a return with the Department of Taxation.”
    - „The CAT rate of 0.26% remains unchanged” and applies above the exclusion.
    - A taxpayer that later exceeds it „must reactivate its CAT account”.
  - The 2025 budget (H.B. 96) made the CAT's net operating loss credit nonrefundable; it did not change the exclusion (secondary — verify).
  - The CAT's own nexus test (a $500,000 Ohio-receipts bright line, ORC 5751.01) and Ohio municipal net-profit taxes were not read; codes.ohio.gov was unreachable.
- **Grade:** 1 plain
- **Rhea's default:** no CAT. Rhea keeps a running total of Ohio-sitused receipts for each calendar year against $6,000,000, with a task at a margin. At about $3M of total revenue, Hoosier cannot reach it in 2026.
- **Rhea today:** nothing. It needs the ship-to state (CO-13).
- **Where the work lives:** *pack data*
- **Deadline:** none while under the exclusion.
- **Watch:**
  - Ohio Department of Taxation CAT information releases (PDFs on https://dam.assets.ohio.gov).
  - Ohio General Assembly biennial budget: odd years, H.B. 33 in 2023 and H.B. 96 in 2025, the next in 2027.
- **Sources:** https://dam.assets.ohio.gov/image/upload/tax.ohio.gov/commercial_activities/information_releases/CAT_2023-01_info_release.pdf ; https://www.cohenco.com/knowledge-center/insights/august-2025/ohio-begins-broader-tax-restructuring-with-new-budget-bill-what-you-should-know (secondary)

---

### CO-23 — California in year two: when are we "doing business", what do we owe, and does P.L. 86-272 protect a seller of goods?
- **Applies to:** corporation
- **The law:**
  - Doing business, FTB page (updated 7 November 2025). A corporation is doing business if it does any of these:
    - it does „any transaction for the purpose of financial gain within California”;
    - it is organised or commercially domiciled there;
    - its California sales, property or payroll exceed the thresholds.
  - Thresholds: for 2025, California sales above $757,070 (or 25% of total sales), and property or payroll above $75,707 (or 25%). The page shows no 2026 figure, and the 2027 figure that Hoosier will meet is not published.
  - P.L. 86-272, same page: „potentially applies to companies located outside of California whose only in-state activity is the solicitation of sale of tangible personal property… Businesses that qualify … are exempt from state taxes that are based on your net income. These entities, however, still may be considered to be doing business in California and may be liable for filing and paying the applicable amounts.”
  - Form 100 Booklet 2025:
    - the rate is 8.84%; the minimum franchise tax is $800;
    - „A corporation that incorporated or qualified through the California SOS … is not subject to the minimum franchise tax for its first taxable year … This does not apply to corporations that are not qualified by the California SOS”;
    - for a foreign corporation not qualified and not doing business, there is no minimum tax, „However, if corporations meet the sale, property, or payroll threshold for 'doing business' under R&TC Section 23101(b), corporations may be subject to the minimum franchise tax”;
    - Form 100 is due the 15th day of the 4th month, and a corporation may file by the 15th day of the 11th month without a written request;
    - estimated tax is 30% / 40% / none / 30% on the 15th day of the 4th, 6th, 9th and 12th months (Form 100-ES; the 9th-month installment is nil).
  - Internet activity: FTB Pub. 1050 (Rev. 05-2022) lists online activities FTB treats as beyond solicitation. A San Francisco Superior Court ruling in *American Catalog Mailers Association v. FTB* reportedly invalidated that guidance for want of rulemaking under the Administrative Procedure Act (secondary — verify; appeal status unknown).
- **Grade:** 3 open.
  - Whether an Indiana seller whose only California activity is solicitation, with a website, owes anything is disputed.
  - FTB's booklet says a company over the threshold „may be subject to the minimum franchise tax” and „may be liable” even when protected.
  - Practitioner commentary reads P.L. 86-272 as reaching any tax measured by net income, and the 2022 internet guidance as void (secondary — verify).
- **Rhea's default:**
  - From 1 January of the year Hoosier's California sales cross the threshold (expected 2027), it is doing business: Form 100 and at least $800.
  - Tax at 8.84% on California-apportioned income, unless a person declares P.L. 86-272 protection as a dated fact with its warrant (only solicitation; orders approved and shipped from outside California; website activity listed).
  - The first-year minimum-tax exemption is not assumed, because Hoosier will not have qualified with the California SOS.
- **Rhea today:** nothing.
- **Where the work lives:** *later* (2027), as *pack data* then.
- **Deadline:** for 2027, estimates on 15 April, 15 June and 15 December 2027 (dates computed here); Form 100 on 15 April 2028.
- **Watch:**
  - FTB "Doing business in California" page (the indexed thresholds; the 2026 and 2027 figures to come).
  - FTB Tax News, https://www.ftb.ca.gov/about-ftb/newsroom/tax-news/index.html.
  - FTB legal rulings and notices.
  - Office of Tax Appeals opinions, https://ota.ca.gov/opinions/.
  - California Legislature, https://leginfo.legislature.ca.gov (refused to scripts).
- **Sources:** https://www.ftb.ca.gov/file/business/doing-business-in-california.html ; https://www.ftb.ca.gov/forms/2025/2025-100-booklet.pdf ; https://www.ftb.ca.gov/forms/2026/2026-100-es.pdf ; https://www.ftb.ca.gov/forms/misc/1050.pdf ; https://www.anchin.com/articles/taxpayers-prevail-in-battle-against-california/ (secondary)

---

### CO-24 — Indiana business personal property tax: do the forklift and racking need a Form 103?
- **Applies to:** everyone with business equipment in Indiana
- **The law:**
  - DLGF Form 103-Short for the 1 January 2026 assessment: „For taxpayers with less than $2,000,000 in acquisition costs to report within the county, IC 6-1.1-3-7.2 exempts this property.” The exemption is claimed by checking the box, entering total acquisition cost and completing „only sections I, II, and IV”, and „you must also file a Form 104.”
  - „If you filed a return and claimed this exemption in the previous assessment year and you continue to qualify …, no return is required.” But „To establish an exemption in a new county, new or existing businesses must file a personal property return during their first year in that new county.”
  - Due „not later than May 15”, with up to 30 days' extension on written request.
  - Property „placed in service after January 1, 2025, is not subject to the 30% minimum valuation limitation” (outside older TIF areas).
  - The threshold rose from $80,000 to $2,000,000 for 2026 by Senate Bill 1 and House Bill 1427 (2025) (Allen County page — secondary).
  - Leased or not-owned equipment goes on Forms 103-O and 103-N.
- **Grade:** 1 plain
- **Rhea's default:**
  - At each 1 January assessment date, Rhea sums acquisition cost by county from the asset register.
  - In the first year it holds property on 1 January (1 January 2027 if Hoosier held none on 1 January 2026), a task to file the exemption claim (103 + 104) by 15 May. After that, a yearly check only.
  - The van is left out (vehicles are not on Form 103; their excise was not checked).
- **Rhea today:** nothing. The asset register is missing (CO-10).
- **Where the work lives:** *pack data*
- **Deadline:** 15 May 2027 for the 1 January 2027 assessment (a Saturday; whether it moves to Monday was not checked).
- **Watch:** DLGF forms and memos, https://www.in.gov/dlgf/forms/dlgf-forms/ and https://www.in.gov/dlgf/assessments/personal-property ; Indiana General Assembly (IC 6-1.1-3).
- **Sources:** https://secure.in.gov/counties/martin/files/SF-11274-R46-103-Short-2026-Business-Tangible-Personal-Property-Return-SECURED.pdf ; https://www.in.gov/dlgf/assessments/personal-property ; https://www.allencounty.in.gov/804/Personal-Property (secondary)

---

### CO-25 — Research and software costs: does §174A reach a distributor in its first year?
- **Applies to:** everyone
- **The law:**
  - §174A(a), added by P.L. 119-21 §70302: „Notwithstanding section 263, there shall be allowed as a deduction any domestic research or experimental expenditures which are paid or incurred by the taxpayer during the taxable year.” Software: „any amount paid or incurred in connection with the development of any software shall be treated as a research or experimental expenditure.”
  - Effective for amounts paid „in taxable years beginning after December 31, 2024” (§70302(e)). A taxpayer may elect instead to amortise over „not less than 60 months” (§174A(c)).
  - The retroactive election for 2022–2024 (§70302(f); Rev. Proc. 2025-28) cannot reach a company founded in 2026.
- **Grade:** 1 plain
- **Rhea's default:** not in the first-year pack. A distributor does no research. The one first-year case is paying someone to develop software for Hoosier, such as a custom ordering portal: that cost is deductible at once for tax, while the book may capitalise it as internal-use software (ASC 350-40, secondary). Rhea raises a task if a purchase is classed as software development.
- **Rhea today:** nothing.
- **Where the work lives:** *later*
- **Deadline:** —
- **Watch:** Internal Revenue Bulletin (§174A guidance after Rev. Proc. 2025-28); Federal Register.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/174A ; https://www.irs.gov/pub/irs-drop/rp-25-28.pdf ; https://www.irs.gov/pub/irs-pdf/i1120.pdf

---

### CO-26 — At year end we accrue bonuses, rebates, taxes and unbilled services: which of them are deductible this year?
- **Applies to:** everyone on accrual
- **The law:**
  - §461(h)(1): the all events test is not met „any earlier than when economic performance with respect to such item occurs”. For services to the taxpayer, performance occurs „as such person provides such services” ((h)(2)(A)(i)).
  - Recurring items, Reg. 1.461-5(b)(1): a liability counts this year if the facts are fixed by year end and economic performance occurs by the earlier of „the date the taxpayer files a timely (including extensions) return” and „the 15th day of the 9th calendar month after the close of that taxable year”. The item must be recurring, and either immaterial or better matched.
  - Taxes, Reg. 1.461-4(g)(6)(i): „economic performance occurs as the tax is paid”. Rebates: as paid (CO-14).
  - Compensation, Reg. 1.404(b)-1T, A-2(b)(1): pay „received after the 15th day of the 3rd calendar month after the end of the employer's taxable year” is presumed deferred compensation, deductible when paid.
  - Related parties, §267(a)(2): an accrual payer's deduction waits until the amount „is includible in the gross income” of a related cash-method payee. An individual is related to a corporation „more than 50 percent in value of the outstanding stock of which is owned, directly or indirectly” by him (§267(b)(2)). Family holdings are attributed (§267(c)).
- **Grade:** 2 settled by Reg. 1.461-4, 1.461-5 and 1.404(b)-1T
- **Rhea's default:**
  - The book accrues everything at year end.
  - The tax book reverses:
    - bonuses not paid by 15 March;
    - any accrual to an owner who, with family, holds more than 50%, until paid (two unrelated 50/50 owners are not related; spouses would be);
    - unpaid rebates;
    - state income tax beyond what was paid (estimates are payments);
    - services not yet performed, except recurring items under the recurring item exception adopted on the first return (CO-07).
  - Each reversal is a rule reading a payment or performance date that arrives after 31 December. Rhea's amendable facts with ruled backfill (the shipped "late understanding") let a payment on 10 March 2027 settle the 2026 tax book.
- **Rhea today:** the matching engine dates payments. The period lock closes months, and only the year for good. There are no accrual rules.
- **Where the work lives:** *pack data*
- **Deadline:** year-end bonuses paid by 15 March 2027 (a Monday). Economic performance for recurring items by the earlier of the filed return and 15 September 2027 (computed here from the regulation).
- **Watch:** Federal Register (Reg. 1.461-4/-5, 1.404(b)-1T); Congress (§461, §267).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/461 ; https://www.ecfr.gov/current/title-26/section-1.461-4 ; https://www.ecfr.gov/current/title-26/section-1.461-5 ; https://www.ecfr.gov/current/title-26/section-1.404(b)-1T ; https://www.law.cornell.edu/uscode/text/26/267

---

### CO-27 — The year-end stock count, shrinkage and slow-moving goods: what changes in the books and in tax?
- **Applies to:** everyone with stock
- **The law:**
  - §471(b) permits shrinkage estimates confirmed „by a physical count only after the last day of the taxable year” if the taxpayer „normally does a physical count of inventories at each location on a regular and consistent basis”.
  - Reg. 1.471-2(c): goods „unsalable at normal prices or unusable in the normal way because of damage, imperfections, shop wear, changes of style, odd or broken lots” are valued „at bona fide selling prices less direct cost of disposition”. „Bona fide selling price means actual offering of goods during a period ending not later than 30 days after inventory date.”
  - Reg. 1.471-2(d): book inventories are „verified by physical inventories at reasonable intervals and adjusted to conform therewith.”
  - Books: ASC 330 write-down to net realisable value, which is not reversed (secondary — verify).
- **Grade:** 1 plain
- **Rhea's default:**
  - A stocktake per location at year end, each counted line a fact. The differences post to shrinkage in both books.
  - A write-down of slow-moving goods goes to the book only.
  - The tax book follows only for goods actually offered at a reduced price within 30 days, or scrapped, with the offer as evidence.
  - The rest of the write-down is an M-1 item and deferred tax.
- **Rhea today:** the Poland pack has `stocktake` v1 (date, comment, status, value) and warehouse quantities, with no FIFO values (CO-08).
- **Where the work lives:** *pack data*, on the *shared ground* of CO-08.
- **Deadline:** count at or near 31 December.
- **Watch:** Federal Register (§471 regulations); FASB ASUs on Topic 330.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/471 ; https://www.ecfr.gov/current/title-26/section-1.471-2

---

### CO-28 — Closing the year: which statements does the lender want, and why does "audited" change our taxes?
- **Applies to:** corporation
- **The law:**
  - No statute sets a private company's closing date or statements beyond IC 23-1-53-1 (statements on a shareholder's written request; CO-04). The loan agreement does.
  - The AICPA SSARS give three levels without audit (secondary — verify):
    - preparation (AR-C 70): no report, a no-assurance legend on each page;
    - compilation (AR-C 80): a one-paragraph report, no assurance;
    - review (AR-C 90): limited assurance.
  - An audit is under AICPA auditing standards.
  - Only an audited statement used for credit is an AFS (§451(b)(3)(A)(ii)). An audit therefore switches on:
    - §451(b) (income no later than in the AFS);
    - the $5,000 de minimis amount (Reg. 1.263(a)-1(f)(1)(i));
    - the AFS route in §471(c);
    - the AFS deferral method in Reg. 1.451-8(c).
- **Grade:** 1 plain as to what makes an AFS. The level of assurance is a term of the loan agreement.
- **Rhea's default:**
  - The company declares the lender's reporting terms as a fact: framework (GAAP or tax basis), level (compiled, reviewed, audited), and due date after year end.
  - "Audited" flips the AFS rules from that year onward.
  - Rhea renders the balance sheet, income statement, statement of changes in equity and cash-flow statement from the book. The accountant's report is a document attached to them.
  - The year close is a period lock on the book after the year-end facts (CO-26, CO-27) are in.
- **Rhea today:** `packs/common/periods.json` closes months per book, and only the year for good. The views and the briefing (narrative ViewDef) can render statements. There is no cash-flow view and no declared lender terms.
- **Where the work lives:** *pack data*
- **Deadline:** as the loan agreement sets.
- **Watch:** AICPA Accounting and Review Services Committee (new SSARS) and Auditing Standards Board; FASB and the Private Company Council, https://www.fasb.org.
- **Sources:** https://www.law.cornell.edu/uscode/text/26/451 ; https://www.ecfr.gov/current/title-26/section-1.263(a)-1 ; https://www.journalofaccountancy.com/issues/2017/jan/preparation-and-compilation-standards/ (secondary)

---

### CO-29 — Do the lender's statements need deferred tax, and how does Rhea keep book and tax apart without two ledgers?
- **Applies to:** corporation (GAAP books)
- **The law:**
  - Books: ASC 740 requires current and deferred income tax: a deferred tax asset or liability on each temporary difference, at enacted rates. GAAP has no private-company exemption, which is one reason some lenders accept tax-basis statements (secondary — verify).
  - Tax: Reg. 1.446-1(a)(4) asks for „a reconciliation of any differences between such books and his return”, filed as Schedule M-1 (CO-30).
  - Hoosier's first-year differences are mostly timing differences (deferred tax): 100% bonus depreciation against straight line (a large liability), Indiana's own depreciation, the lease, the allowance, and accrued bonuses and rebates.
- **Grade:** 2 settled — ASC 740 (secondary)
- **Rhea's default:**
  - Deferred tax = enacted rate × (book basis − tax basis), per balance, and per asset for fixed assets, read straight from the two books (CO-04). The rates are 21% federal and 4.9% Indiana on the Indiana-apportioned share, Indiana tax being deductible federally (computed by Rhea, not a source).
  - Current tax = the year's federal and Indiana tax less estimates paid.
  - Both are year-end rules gated on the company being a C corporation.
- **Rhea today:** books as a posting dimension, exact arithmetic, year lock. No tax-provision rule.
- **Where the work lives:** *pack data*
- **Deadline:** with the lender's statements.
- **Watch:** FASB ASUs on Topic 740; Congress (rate changes are deferred-tax remeasurements on enactment).
- **Sources:** https://www.ecfr.gov/current/title-26/section-1.446-1 ; https://www.law.cornell.edu/uscode/text/26/11

---

### CO-30 — The federal return: Form 1120 at 21%, which schedules, which attachments, and by when?
- **Applies to:** corporation
- **The law:**
  - Rate, §11(b): „21 percent of taxable income.” Due date, §6072(a): fiscal-year returns are filed „on or before the 15th day of the fourth month following the close of the fiscal year”; for a calendar year „the 15th day of April”.
  - Extension, Reg. 1.6081-3(a): „an automatic 6-month extension” on Form 7004. It extends filing, not payment.
  - Instructions for Form 1120 (2025):
    - corporations with total receipts and total assets „less than $250,000 are not required to complete Schedules L, M-1, and M-2”;
    - total assets „of $10 million or more on the last day of the tax year must file Schedule M-3 … instead of Schedule M-1”;
    - Form 1125-A for cost of goods sold; Form 1125-E when total receipts „are $500,000 or more”;
    - e-filing is required „if the corporation files 10 or more returns of any type during the calendar year (including income tax, employment tax, excise tax, and information returns)” (Reg. 301.6011-5);
    - the minimum penalty for a return more than 60 days late, for returns due in 2026, is the smaller of the tax due or $525.
  - The 2026 instructions were not yet published on 16 October 2026.
- **Grade:** 1 plain
- **Rhea's default:**
  - Hoosier files Form 1120 with Schedules L, M-1 and M-2, Forms 1125-A, 1125-E and 4562, and the election statements (CO-07).
  - Its W-2s and Forms 941 alone pass the 10-return count, so it e-files (counted here from ten employees).
  - Rhea renders the return's data from the tax book: page 1 and 1125-A from the account mapping (CO-05), Schedule L from the book's balance sheet, M-1 from the adjustment book.
  - It hands this to the preparer or an authorised e-file provider. Transmission through IRS Modernized e-File is out of scope for the first year.
- **Rhea today:** the writer seam (`adapter.ProvideWriter` / `Render` / `Issue`, used for JPK_V7M) is the place for a return writer. `tax_declaration.kind` is Polish.
- **Where the work lives:** *market code* (a Form 1120 data writer in `markets/us`, behind the writer seam), fed by *pack data*.
- **Deadline:** 15 April 2027 (Thursday); extended, 15 October 2027 (Friday).
- **Watch:** IRS Form 1120 page and draft forms (the 2026 edition), https://www.irs.gov/forms-pubs/about-form-1120 and https://www.irs.gov/draft-tax-forms ; IRS MeF schemas and business rules (irs.gov e-file provider pages); Congress (§11).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/11 ; https://www.law.cornell.edu/uscode/text/26/6072 ; https://www.ecfr.gov/current/title-26/section-1.6081-3 ; https://www.irs.gov/pub/irs-pdf/i1120.pdf

---

### CO-31 — The S-corporation variant: what changes in the books if the owners elect?
- **Applies to:** corporation (S variant)
- **The law:**
  - Eligibility and timing: CO-03.
  - Return, §6072(b): S corporation returns are filed „on or before the 15th day of March following the close of the calendar year”. Form 1120-S with a Schedule K-1 for each shareholder.
  - Distributions without earnings and profits, §1368(b): for an S corporation with no accumulated earnings and profits (one that was S from its first year), a distribution „shall not be included in gross income to the extent that it does not exceed the adjusted basis of the stock”, and the excess is gain.
  - With earnings and profits (an election after C years), §1368(c)(1): the part „which does not exceed the accumulated adjustments account shall be treated in the manner provided by subsection (b)”. The 1120-S instructions: „S corporations with accumulated earnings and profits (AE&P) must maintain the AAA”, and „The AAA may have a negative balance at year end.”
  - Basis, §1367(a)(2): each shareholder's basis decreases (not below zero) by distributions that were not income under §1368.
  - Shareholder pay, IRS (page reviewed 3 March 2026): „S corporations must pay reasonable compensation to a shareholder-employee in return for services that the employee provides to the corporation before non-wage distributions may be made”. Its factors include „Duties and responsibilities”, „Time and effort devoted to the business” and „What comparable businesses pay for similar services”.
  - Accruals, §267(e)(1)(B)(ii): for an S corporation „any person who owns (directly or indirectly) any of the stock” is related, so accrued pay to any owner is deductible only when paid.
- **Grade:** 1 plain (reasonable compensation is a question of facts, settled by the IRS factors and courts)
- **Rhea's default:**
  - A declared S election switches off the C-corporation tax rules (provision, deferred tax, estimates, CO-19 and CO-29).
  - It turns on: the AAA (from the first S year, because §1368(c) needs it once there are C-year earnings); per-shareholder basis (stock and loans) fed by K-1 items, which decides how much of a distribution is tax-free under §1368(b); allocation per share per day; a check that distributions are pro rata (one class of stock); and a check that each working owner draws wages through payroll before distributions.
  - The book's equity shows AAA apart from retained earnings.
  - Indiana (IT-20S) and California rules for S corporations are not covered here.
- **Rhea today:** as CO-18; nothing S-specific.
- **Where the work lives:** *pack data* (a variant gated on the declared election).
- **Deadline:** 1120-S for 2026 on 15 March 2027 (Monday).
- **Watch:** Form 1120-S instructions; IRS S-corporation compensation page; Congress (§1361–1368).
- **Sources:** https://www.law.cornell.edu/uscode/text/26/6072 ; https://www.law.cornell.edu/uscode/text/26/1367 ; https://www.law.cornell.edu/uscode/text/26/1368 ; https://www.law.cornell.edu/uscode/text/26/267 ; https://www.irs.gov/pub/irs-pdf/i1120s.pdf ; https://www.irs.gov/businesses/small-businesses-self-employed/s-corporation-compensation-and-medical-insurance-issues

---

### CO-32 — How long do we keep the books, and may they exist only electronically?
- **Applies to:** everyone
- **The law:**
  - Reg. 1.6001-1(e): records „shall be retained so long as the contents thereof may become material in the administration of any internal revenue law.”
  - §6501(a): tax „shall be assessed within 3 years after the return was filed”.
  - Form 1120 instructions (2025): „Usually, records that support an item of income, deduction, or credit on the return must be kept for 3 years from the date the return is due or filed, whichever is later. Keep records that verify the corporation's basis in property for as long as they are needed”.
  - Electronic books, Rev. Proc. 98-25 §3.02(2): a taxpayer with assets under $10 million must comply if „all or part of the information required by § 6001 is not in the taxpayer's hardcopy books and records, but is available in machine-sensible records”. §3.01(3): „all requirements of § 6001 that apply to hardcopy books and records apply as well to machine-sensible books and records”.
  - Indiana's own retention rule (IC 6-8.1-5-4) was not read.
- **Grade:** 2 settled — electronic books by Rev. Proc. 98-25
- **Rhea's default:** Rhea's books are machine-sensible, so Rev. Proc. 98-25 applies in full.
  - The append-only log, each rule version with its approval, and every document are the records and their documentation. They are kept at least until the later of three years after the return's due or filing date and the life of the related asset basis.
  - Nothing is deleted: the log cannot delete, and backup and restore keep it whole.
  - A retention view shows, per year, the earliest date a record could go. Disposal stays a human decision outside Rhea.
- **Rhea today:** right in shape. The append-only event log, replay that reproduces state, versioned rules with approvals, and backup and restore are what §6001 and Rev. Proc. 98-25 ask for.
- **Where the work lives:** *pack data* (the retention view only)
- **Deadline:** three years from the later of due date and filing, longer for basis.
- **Watch:** Internal Revenue Bulletin (successors to Rev. Proc. 98-25 and 97-22), https://www.irs.gov/irb
- **Sources:** https://www.ecfr.gov/current/title-26/section-1.6001-1 ; https://www.law.cornell.edu/uscode/text/26/6501 ; https://www.eitc.irs.gov/pub/irs-drop/rp-98-25.pdf ; https://www.irs.gov/pub/irs-pdf/i1120.pdf


