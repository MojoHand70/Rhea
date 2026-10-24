# EXPORT.md — Rhea beside the office's program

A design, not built (Claude, 2026-10-24). KK asked: "Can we think about Rhea working mode, where
Rhea prepares export to other financial systems?" KK has made no decision about Rhea working
as a layer over another ledger. This note designs the mode so that it can be decided; the
questions are at the end.

## The mode in plain words

For each company, one choice: **where are the official books kept?**

- **In Rhea**, as today.
- **In the office's program** (Optima, Symfonia, enova, Rewizor…). Rhea does everything up to
  the entry: she takes documents in, checks them, books them by rules, pairs the bank, spreads
  costs and depreciates. Then she hands the month's entries to the program. The office keeps
  its program for the official books, the VAT return and JPK, the taxes and the close.

**Why.** An office changes its program rarely: years of data, habits, people trained on it.
With export, an office can start with Rhea tomorrow without leaving its program, which is how
flowMEE sells today. It can move the books to Rhea when it is ready: "an on-ramp, not a cliff",
in Numeric's words. Export is also the shortest road to the first offices. The statutory floor
(sending the JPK, the KPiR book, PIT and CIT, the close) stays in the program while Rhea's own
floor is being finished.

## What does not change: Rhea keeps her own books

*My proposal* (question 1): in export mode Rhea still books every fact by rules into her own
books, and the export is those books written in the program's language. That gives four things
a layer cannot have:

- every exported entry can be explained (the fact, the rule, its version, the law) and
  replayed;
- moving the official books to Rhea is a switch, not a migration: the history is already in
  Rhea;
- each month Rhea can compare her books with the program's (below);
- the same rules serve a company in either mode, so nothing is built twice.

## What goes to the program, and what stays there

| Goes to the program | Stays in the program |
|---|---|
| Purchase and sales documents into the VAT registers, by rate, with JPK markings and GTU | the official ledger |
| The entries (Wn/Ma per line, MPK and project) where the program takes entries | the VAT return and sending the JPK |
| KPiR entries (column) and the ryczałt record, for sole traders | PIT, CIT, ZUS |
| Bank statements with their pairings | the month's and year's close, the statements |
| RMK parts, depreciation, the month-end entries Rhea made (FX revaluation, provisions) | |

For a company in export mode Rhea does not send the JPK and does not close officially. M5
(sending to the state) is off for it. Her VAT settlement becomes a cross-check.

## Two depths, per program

1. **Into the registers.** Rhea puts documents into the program's VAT registers (and KPiR),
   and the program's own posting schemes make the entries. The office's habits stay as they
   are.
   - **Optima** has read a JPK_VAT file into its VAT registers since version 2018.1.1. The
     import simplifies payment form, transaction type, currency, rate, description and KPiR
     column, and sets every purchase as deductible ([Comarch help](https://pomoc.comarch.pl/optima/pl/2026/dokumentacja/czy-istnieje-mozliwosc-importu-plikow-jpk_vat-do-rejestrow-vat/)).
     Rhea already writes the JPK_V7M file (function 51, kernel only), so a first test needs
     almost nothing new. Optima's own document format can carry more (a category per line,
     which drives Optima's posting schemes). flowMEE sends that format, per their API; the
     format itself is still *to verify*.
   - **Rewizor GT** reads an EPP communication file and books it by its "import and dekretacja
     schemes" ([Kamsoft's description](https://gtit.pl/dane/kamsoft_gt/Opis%20Programu%20KAMSOFT%20GT.pdf)).
   - **Symfonia** has its own import structure; **enova365** imports XML (both *to verify*).
2. **Finished entries.** Rhea sends the entries she made, and the program books exactly those,
   bypassing its own schemes. This needs the program's import of journal entries (*to verify*
   per program).

*Recommendation:* depth 1 in Optima through the JPK file first; then Optima's document format
with categories; depth 2 where an office wants Rhea's entries exactly.

## The client's chart and lists

Rhea must book to the client's own accounts in the program. At onboarding she reads the
program's chart, counterparties, categories and MPK from what the program can export. For full
books, the JPK_KR carries the chart and the year's entries. Which of the client's accounts each
of Rhea's bookings goes to is taught once per company; the office teaches it once for all its
clients that share a chart (function 70).

This is the chart question (M1) arriving for real: in export mode, the company's chart is the
program's chart. It fits KK's direction of 2026-10-24 ("a company has its own books; a
company with its own chart teaches Rhea which of its accounts is which"). That direction is
still under discussion; the export needs it answered.

The JPK_KR also gives the past year's entries. The interview can then propose the rules that
reproduce them (learning from history, function 86), and Rhea's books start from the opening
balances (76).

## How the entries travel

1. **A file the accountant imports** (a writer, like the JPK): nothing to install, works with
   any edition, and the accountant sees what goes in. First.
2. **A connector beside the program**, for offices with many clients, like flowMEE's
   synchroniser. It is installed on the office's computer. Later. Through API.md it is just
   another program with a key.
3. **The program's own API** where there is one (cloud editions, Comarch Księgowość 360,
   wFirma; *to verify*).

## Sound by construction

- **An export is a fact**: the batch, its period, and the entries in it by id and hash. An
  entry once exported is never in a later batch unchanged: the dedup key names the entry, so
  the same month exported twice adds nothing (invariant 10).
- **A change after export** (a correction, a late document in a sent month) travels the way the
  program expects: a correction document or a reversing entry in the next batch. It is never a
  silent second original.
- **The program's answer is a fact**: its document numbers, or its refusals. It comes through
  a door: the accountant uploads the import report, or the connector posts it
  (`/v1/…/exports/{program}/{batch}/answer`). A refused entry is a question on the worklist.
- **The writer is deterministic in the books**: the same books and period give the same bytes
  (the writer contract). Replay never exports again; the export facts are already in the log,
  as KSeF sending is.
- **Isolation**: a batch is one company's; an office exports per client; the dedup key carries
  the company (invariant 9).
- **Money**: amounts by rate as Rhea booked them. The VAT is the supplier's (C1) and must not be
  worked out again by the program on import (*to verify* per program).
- **The market's code**: Optima, Symfonia, enova and InsERT are Polish programs, so their
  formats live in `markets/pl`. Xero and QuickBooks would live in their own market. The kernel
  never names a program (invariant 8).

## The monthly comparison

Each month Rhea reads the program's own JPK_V7, and its JPK_KR for full books: files the office
makes anyway. She compares them with what she exported. Documents missing, amounts changed in
the program, a document typed in by hand as well (booked twice), another VAT period: the
differences are the month's questions. A layer cannot do this, because it keeps no books; the
program alone cannot either. It stands on the JPK readers (73, 74).

## Moving the books to Rhea

When the office decides, from a date: Rhea's books become the official ones and the export
stops. The opening balances are Rhea's own, already booked, checked against the program's
closing ones (JPK_KR). M5 sends the JPK from that month. The year's start is the natural date.

## Order of work (a proposal)

1. Optima: read the chart and lists; the mapping taught once; the JPK_V7M file into Optima's
   registers. One office with Optima to try it on.
2. The import report back as a fact; the monthly comparison.
3. Optima's own document format with categories, or a connector.
4. Rewizor GT, Symfonia and enova, in the order the offices we meet use them.
5. Finished entries (depth 2).

## Tests before an office imports a Rhea file into its real books

- [ ] the same month exported twice exports nothing new, also from two exports at once
- [ ] a correction after export travels as a correction, never as a second original
- [ ] company A's file holds nothing of company B; an office's batch is per client
- [ ] Optima takes the file on a real installation, and its register totals by rate equal Rhea's
- [ ] replay after a month of exports reproduces identical state and exports nothing
- [ ] the monthly comparison finds a document typed in by hand in the program

## For KK to decide

1. **In export mode, does Rhea keep her own books too?**
   *Example:* a client's October goes to Optima. Does Rhea also hold October's journal?
   (a) Yes: the export is her books written for the program. (b) No: she only prepares entries
   and keeps none.
   *My recommendation: (a).* Each entry stays explained and replayable, the switch to Rhea's
   books costs nothing, and the monthly comparison becomes possible.
2. **Which program first?** Optima, Rewizor GT, Symfonia, enova.
   *Recommendation: Optima.* It is the most common program in the offices we target (Comarch
   holds about a quarter of the Polish market), and the JPK file path exists today.
3. **How do the entries reach it first?** (a) a file the accountant imports; (b) a connector;
   (c) the program's API.
   *Recommendation: (a).* Nothing to install, and a real office can try it in days.
4. **The chart.** Export mode needs each company's own chart (M1), which is under discussion.
   Shall we settle the chart design next, with export as its first use?
