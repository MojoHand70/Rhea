# API.md — Rhea's API

A design, not built (Claude, 2026-10-24, at KK's request). It waits for KK's answers to the
questions at the end. It rests on the machine as it is: one door for facts, actions and views
declared as data, the same rights and scope for every caller.

## What a program can do with Rhea, in plain words

A program is anything that is not a person at the desk: an office's own system, a shop, a
software house's platform, another accounting program, a script, an AI agent. Through the API
it can:

- **send Rhea a document** (a KSeF invoice, a PDF, a photo, a bank statement) and get back what
  Rhea read, what she checked, how she booked it, or the question she is waiting on;
- **read anything a person could read in that company**: every list, every record, every
  report, and why each figure is there (the fact, the rule, its version, the law);
- **do what a person could do with an action**: issue an invoice, record a payment, add a
  supplier. It uses the same actions the desk offers, with the same checks;
- **be told when something happens**: a document booked, a question waiting, a month closed;
- **take a file**: the JPK, a statement, the month's entries for another program (EXPORT.md).

It **cannot approve**: a rule, a pack, a lesson, or the accountant's decision on how a document
is booked. Those stay with a signed-in person (invariant 2: the approval is a person's act).

**What sets it apart.** flowMEE's API has 114 operations, each written by hand. Rhea's API is
drawn from the definitions, the way M2 draws the desk's forms. Every view a company has can be
read, and every action it has can be called. A pack's new version reaches the API with no code
change, and a business pack written by an office gets its API on the day it is installed.

## Principles

1. **A program is a user** (the adapter contract's own words: "an adapter is a user, not a
   kernel extension"). It reads what its rights allow and acts only by facts through doors. The
   kernel never calls a program back except to tell it that something happened.
2. **The same rights, scope and refusals as a person.** A key is a grant in a company: a role,
   from a date, until it is withdrawn. An office's key sees the office's clients (a scope set,
   invariant 9) and acts for one client per call. Row-level security applies unchanged: a
   query that forgets its scope returns nothing.
3. **Every write is a fact with its idempotency key** (invariant 10). The caller names what the
   thing is (the invoice's number and the seller's NIP, the statement's number), never only its
   bytes. The same call twice is one fact, and the answer says so. A write without a key is
   refused.
4. **Programs never approve.** Rule, definition, pack and lesson approvals, a path's approval
   step, and the accountant's signature on a booking (UoR art. 21 ust. 1a) are refused to a
   key. A program may only send the facts those decisions rest on.
5. **Money is a string with its currency** (`"1230.45"`, `"PLN"`); dates are ISO; no floats
   (invariant 6).
6. **Everything explains.** Any figure in any answer can be asked "why".
7. **Refusals in the company's language, field by field**, as the desk shows them (M2).
8. **Stable outside, free inside.** `/v1` is a thin layer over the same kernel functions the
   desk uses (trigger an activity, read a view, explain). The desk's internal `/api` can keep
   changing. Within `/v1` changes are additive; a pack change that would break a field is caught
   when the pack is designed (the designer's "breaking" class, DIRECTION 2026-10-23).
9. **The log is the feed.** Events out come from the company's append-only log, in its order,
   so a program that missed a day reads from where it stopped.

## Who calls, and how

- **A key per program per company** (or per office). The owner creates it in Ludzie →
  "Programy": a name, a role, the companies, from when. The key is shown once; Rhea keeps only
  its hash. The grant is a fact; withdrawing it is a fact, and the key opens nothing from that
  moment (functions 4 and 6, for programs).
- `Authorization: Bearer <key>`. The actor of every fact the key writes is
  `program:<name>`, the way an adapter's is `adapter:<name>`. "Who did this" always has an
  answer.
- **Limits:** a body at most 32 MB (the kernel's limit today); a rate per key; the KSeF limits
  are Rhea's to keep, never passed on to the caller.
- **A test company** for integrators: its own installation, KSeF's test system, the clock that
  can be moved (`/api/time/simulate` exists).

## The resources

`{c}` is the company: its NIP, or its key within an office.

| | Call | What it does | Built on today |
|---|---|---|---|
| Who | `GET /v1/me` | the key, its role, its companies, its rights | `/api/whoami`, `/api/rights` |
| | `GET /v1/companies` | the companies in scope (an office's clients) | the scope set |
| The map | `GET /v1/companies/{c}/spec` | an OpenAPI document of this company's views and actions, drawn from its installed packs | M2's form from the definition |
| | `GET /v1/companies/{c}/packs` | installed packs and versions | `/api/packs` |
| Read | `GET /v1/companies/{c}/views` | the views the key may read | `/api/nav` |
| | `GET /v1/companies/{c}/views/{view}?filter…&after=` | a view's rows, paged | `/api/views/{id}` |
| | `GET /v1/companies/{c}/records/{type}/{id}` | one record with its fields | `/api/objects/{id}` |
| | `GET /v1/companies/{c}/why?record=…&field=…` | the fact, the rule, its version, its legal basis | `/api/explain` |
| | `GET /v1/companies/{c}/rules` | the rules in force, with warrants, read-only | `/api/rules` |
| Act | `GET /v1/companies/{c}/actions` | the actions the key may call, each with its inputs | `/api/activities` |
| | `POST /v1/companies/{c}/actions/{name}` | call it; `Idempotency-Key` required | `/api/activities/{name}/trigger` |
| Documents | `POST /v1/companies/{c}/documents` | send a file (KSeF XML, PDF, photo, MT940); the reader is chosen by its kind | the intake doors |
| | `GET /v1/companies/{c}/documents/{id}` | status, what was read, the checks, the booking, the open question | intake + views |
| | `GET /v1/companies/{c}/documents/{id}/file` | the file as received | `/api/blobs/{hash}` |
| Questions | `GET /v1/companies/{c}/questions` | what waits for a person, and why | `/api/worklist`, `/api/untaught` |
| | `POST /v1/companies/{c}/ask` | a question in words; the answer with its sources, read-only | `/api/answer` |
| Files | `GET /v1/companies/{c}/files` | the files Rhea can write here (JPK, statements, exports) | `/api/files` |
| | `GET /v1/companies/{c}/files/{writer}?period=` | render one | `/api/files/{writer}` |
| | `POST /v1/companies/{c}/files/{writer}/handed` | record that it was handed on (to the office's program, to the state) | `/api/files/{writer}/issue` |
| Exports | `POST /v1/companies/{c}/exports/{program}/{batch}/answer` | the other program's answer: its document numbers or its refusals (EXPORT.md) | — |
| Told | `GET /v1/companies/{c}/events?after=` | the company's facts in order, from a cursor | the log |
| | `POST /v1/companies/{c}/subscriptions` | a URL, the kinds wanted, a secret | — |
| Health | `GET /v1/health` | up or not | `/api/health` |

### Three calls, as a program sees them

**Send a document.**

```
POST /v1/companies/5260250995/documents
Idempotency-Key: ksef/5213000000-20261023-0A1B2C-7F
Content-Type: application/xml

<Faktura …>  (an FA(3) file)
→ 202 {"document": "doc_7Q…", "status": "received"}
```

Later, by `GET` or by a subscription:

```
{"document": "doc_7Q…", "status": "booked",
 "read": {"seller": {"nip": "5213000000", "name": "…"}, "number": "FV/10/2026/123",
          "issued": "2026-10-23", "gross": {"amount": "1230.00", "currency": "PLN"},
          "by_rate": [{"rate": "23", "net": "1000.00", "vat": "230.00"}]},
 "checks": [{"check": "white_list", "result": "passed", "rule": "pl-white-list", "version": 3}],
 "booked": [{"account": "402", "debit": "1000.00"}, {"account": "221", "debit": "230.00"},
            {"account": "202", "credit": "1230.00"}],
 "why": "/v1/companies/5260250995/why?record=doc_7Q…"}
```

or `"status": "waiting"`, with the question a person must answer and who it waits for.

**Call an action.**

```
POST /v1/companies/5260250995/actions/issue_invoice
Idempotency-Key: shop-order-88213
{"buyer_nip": "…", "issued": "2026-10-24", "lines": [{"item": "…", "qty": "2", "unit_price": "100.00", "rate": "23"}]}
→ 200 {"fact": "evt_…", "done": true, "records": ["invoice/FV-2026-10-0042"]}
→ 422 {"refused": [{"field": "buyer_nip", "says": "NIP ma 10 cyfr"}]}
```

The same call again answers `"repeated": true` with the first fact.

**Be told.** A subscription sends each new fact of the kinds asked for, signed with the
subscription's secret, retried until answered, each carrying its event id. A receiver drops a
delivery it has seen; Rhea delivers at least once and never claims exactly once.

## What it builds on, and what is new

Built: the door for facts with dedup keys, activities with their inputs declared, views,
explain, the worklist, writers, the scope set and row-level security, the refusal as a recorded
fact, the request log.

New:
- **The program principal**: grants for keys, a hash kept, withdrawal, the actor
  `program:<name>`; the face's sign-in stays as it is.
- **`/v1`**: the stable layer, `{c}` in the path, `Idempotency-Key` on writes.
- **The spec drawn per company** from its packs (M2's logic, written out as OpenAPI).
- **Documents as one call**: the file kept, the reader chosen by kind, the status gathered from
  intake, checks, booking and the worklist.
- **Events by cursor and subscriptions**: the outbound side, as an adapter (DIRECTION's
  "notifications" note), each delivery a fact so none is sent twice by Rhea's own doing.
- **Approval refusal for keys**: one check at the approval doors, tested.

An MCP server for AI agents comes later, as one more reader of the same two vocabularies
(DIRECTION, "Agent interface"). An agent holds a key like any program, and like any program it
cannot approve (invariant 3).

## Tests that must hold before anyone outside gets a key

- [ ] the isolation walk runs over `/v1` with two companies: a key of A reads nothing of B,
  through every resource, including the feed and the files
- [ ] an office key acts only for its clients, one per call, and its refusal is a fact
- [ ] every write without `Idempotency-Key` is refused; the same write twice is one fact,
  from several goroutines at once
- [ ] a key cannot approve a rule, a bundle, a pack, a lesson or a booking decision
- [ ] a withdrawn key opens nothing from that moment
- [ ] every fact a key wrote names `program:<name>` as its actor
- [ ] the spec for a company lists exactly the views and actions its packs give it
- [ ] money in every answer is a string; no float anywhere in `/v1`
- [ ] replay after a day of API traffic reproduces identical state

## For KK to decide

1. **May a program approve anything?**
   *Example:* an office's own system wants to confirm the bookings an accountant already
   checked there.
   (a) never; approvals stay with a signed-in person in Rhea. (b) A person's own key may
   approve as that person. *My recommendation: (a)* for now. The law asks for the
   accountant's signature, and invariant 2 makes approval a person's act. (b) can come later
   with a second factor.
2. **Who may create a key?** (a) the company's owner; (b) the owner or the office; (c) any
   member with the right. *Recommendation: (b)*, the same people who may install a pack.
3. **When does the API open?** *Recommendation:* after the tests above are green and the export
   (EXPORT.md) uses it. The first outside caller is then one of our own.
