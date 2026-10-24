# Invoices in practice — what an invoice holds, how it is issued and booked (Poland, 2026)

Research for FUNCTIONS.md, functions 25 (enter a cost invoice) and 30 (issue a sales invoice),
2026-10-24, from the web: the Ministry's brochures and KSeF handbook, Comarch Optima's help, and
accounting guides (iFirma, Poradnik Przedsiębiorcy, Infor, FakturaXL). Guides are not the law: each
point marked *to verify* needs the act, the regulation or the Ministry's schema before code rests
on it.

## 1. What an invoice holds (VAT act art. 106e, KSeF FA(3))

**On the invoice as a whole:** date of issue; a number in one or more series that identifies it
once; seller and buyer with their NIP; the date of supply (or of payment received) if it differs;
the totals by rate — the net of each rate and its VAT — and the gross due. KSeF does not check the
arithmetic, but the totals must agree with the lines.

**On each line** (FA(3) `FaWiersz`):

| FA(3) | What it is | |
|---|---|---|
| NrWiersza | the line's number | required |
| P_7 | name of the goods or service (up to 512 characters) | required |
| P_8A | unit of measure ("szt.", "godz.", "kg") | required when a quantity is given |
| P_8B | quantity | |
| P_9A | net unit price (P_9B: gross unit price, when the invoice is priced gross) | |
| P_10 | the amount of a discount or markdown, unless already inside the unit price | optional |
| P_11 | the line's net value (P_11A: its gross, on a gross-priced invoice) | |
| P_12 | the VAT rate: 23, 8, 5, 0 (domestic, intra-EU, export), zw, oo, np | required |
| GTIN, PKWiU, CN, Indeks, GTU | codes — EAN, classifications, the seller's own SKU, the JPK groups | optional |

**Totals by rate:** P_13_1 / P_14_1 net and VAT at 23%, P_13_2 / P_14_2 at 8%, P_13_3 / P_14_3 at
5%, P_13_6 at 0%, P_13_7 exempt, P_13_8–P_13_10 outside the scope or reverse charge, P_13_11 the
margin scheme; P_14_xW the VAT in złoty on an invoice in another currency; P_15 the gross due.

**VAT is worked out by rate:** the VAT act asks for the tax on the sum of each rate's net values
(art. 106e ust. 1 pkt 14), rounded to the grosz (ust. 11); the seller may instead state the VAT of
each line and add them up (ust. 10). An invoice priced gross works the other way — from gross unit
prices down. *To verify:* ust. 7–8 and how FA(3) marks a gross-priced invoice.

## 2. Quantities and discounts

- **Quantity and unit** are part of every line (art. 106e: "miara i ilość dostarczonych towarów
  albo zakres wykonanych usług").
- **A discount given at the sale** may be shown two ways, both correct: the price already lowered
  (one net unit price after the discount), or the price before it with the discount's amount in
  P_10 and the line's value after it. The discount lowers the taxable base (art. 29a ust. 7).
- **A discount given after the sale** (a rebate for the quarter, a price agreed down later) is never
  a change to the invoice: it is a correction invoice (faktura korygująca) referring to it.
- **Keeping the price before and after:** FA(3) carries the net unit price, the discount amount and
  the line value, so the price before the discount, the discount and the price after it are all on
  the document — kept as the document is kept. *To verify:* whether P_9A is the price before or
  after the discount when P_10 is filled (the guides disagree).

## 3. Issuing an invoice

The seller enters the buyer (by NIP — GUS and the VAT list fill the rest), the dates, the lines
(item, unit, quantity, net or gross unit price, discount, rate, codes), the payment (method, due
date, bank account; the split-payment mark where it applies); the program works out each line's
value, the totals by rate and the gross; gives the next number in the series; sends the FA(3) to
KSeF, which returns the KSeF number — from 2026 the invoice exists when KSeF has accepted it.

## 4. Booking a cost invoice

**Full books** (Comarch Optima's booking scheme for purchases, as an example): the gross to the
supplier's settlement account (Ma 202-…); the deductible VAT to input VAT (Wn 221-…), only when the
deduction is "TAK" or "WARUNKOWO"; the net to the cost account its **category** decides (fuel and
energy to 411, external services to 429 in that example) — and where VAT is not deductible, the
gross goes to the cost. An invoice is split into rows by **what each part is for**, each with its
net; the VAT rate is not what splits it — the VAT is taken from the document as it is.

**VAT deduction:** in full, none, or limited — 50% for a passenger car used for the business and
privately (the net and the VAT entered in the register are halved).

**The JPK (JPK_V7, from 1 February 2026):** purchases are split not by rate but into fixed assets
(K_40 net, K_41 VAT) and all other purchases (K_42 net, K_43 VAT), with corrections in K_44–K_47.
*Today Rhea's JPK writes only K_42/K_43.*

**KPiR (new template from 1 January 2026, 19 columns):** a new column 3 for the KSeF number and
column 5 for the counterparty's tax ID; the other columns moved by two — purchase of goods and
materials in column 12 (was 10), other expenses in column 15 (was 13), column 16 the sum of costs.
The price of goods is the amount due less deductible VAT and less discounts. The same expense may
go to different columns by its purpose (packaging bought to sell is goods; for the office it is
another expense); transport of goods is an incidental cost; transport of a fixed asset is added to
its initial value. *To verify:* every column number against the regulation of 6 September 2025.

## 5. What this means for Rhea

- **An invoice keeps its lines, not only its totals** — item, unit, quantity, unit price, discount,
  value, rate, codes — as FA(3) carries them, and the totals by rate. A client who wants the price
  before and after a discount has it: it is on the line.
- **A sales invoice:** lines in, totals by rate worked out by Rhea's rule (net-priced or
  gross-priced, as the invoice says), the number from the series, sent to KSeF.
- **A cost invoice entered by hand:** the document's totals by rate (to check it against), then the
  booking split into rows by what each part is for — a cost category (or account, or KPiR column) —
  each row with its net, its VAT and how much of it is deductible, and whether it is a fixed asset.
  The VAT register and the JPK take the deductible part and the fixed-asset split from those rows.
- **A cost invoice from KSeF** arrives with its lines; the accountant (or a lesson Rhea learned)
  gives it the same split.

## Sources

- [Ministry of Finance — JPK_VAT z deklaracją, broszura od 1 lutego 2026](https://www.podatki.gov.pl/media/wgbkrejs/broszura-jpk_vat-z-deklaracj%C4%85-od-1-lutego-2026-r.pdf)
- [Ministry of Finance — KSeF 2.0 podręcznik, część II: wystawianie i otrzymywanie faktur](https://ksef.podatki.gov.pl/media/cark5buy/podrecznik-ksef-20-czesc-ii-wystawianie-i-otrzymywanie-faktur-w-ksef-19022026.pdf)
- [JPK.info.pl — dane na fakturze KSeF (FA(3) fields)](https://jpk.info.pl/faktury/e-faktura-ustrukturyzowana-ksef/dane/)
- [B2B Router — mapping invoice fields to KSeF FA(3)](https://developer.b2brouter.net/docs/mapping_json_invoice_to_ksef)
- [GOBL — FA(3) addon, tax categories](https://docs.gobl.org/addons/pl-favat-v3.md)
- [Comarch Optima — wielowariantowy wzorzec księgowania zakupów](https://pomoc.comarch.pl/optima/pl/2026/?p=4913)
- [Poradnik Przedsiębiorcy — kolumny 10, 11 i 13 KPiR](https://poradnikprzedsiebiorcy.pl/-kolumny-10-11-i-13-kpir-charakterystyka)
- [iFirma — zakup opakowań, w której kolumnie KPiR](https://www.ifirma.pl/blog/zakup-opakowan-w-ktorej-kolumnie-w-kpir-nalezy-go-ujac.html)
- [iFirma — koszty uboczne zakupu w KPiR](https://www.ifirma.pl/blog/co-to-jest-koszt-uboczny-zakupu-dla-prowadzacego-kpir.html)
- [Infor — wydatki na paliwo w ewidencji i JPK VAT (50%)](https://ksiegowosc.infor.pl/podatki/vat/ewidencja-vat/778101,Wydatki-na-paliwo-w-ewidencji-i-JPK-VAT.html)
- [FakturaXL — rabat na fakturze](https://www.fakturaxl.pl/rabat-na-fakturze)
- [Infor — rabaty przedsprzedażowe na fakturze](https://ksiegowosc.infor.pl/podatki/vat/faktura/696582,Rabaty-przedsprzedazowe-na-fakturze.html)
- [Podatki.biz — treść faktury przy upustach](https://www.podatki.biz/artykuly/tresc-faktury-w-przypadku-udzielania-upustow_14_21707.htm)
- [Infor — jak zaokrąglać VAT na fakturze](https://ksiegowosc.infor.pl/podatki/vat/faktura/729678,Jak-zaokraglac-podatek-VAT-na-fakturze.html)
- [Poradnik Przedsiębiorcy — elementy faktur](https://poradnikprzedsiebiorcy.pl/-elementy-faktur-co-trzeba-umiescic-na-fakturze)
