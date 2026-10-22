package adapter

import (
	"context"
	"encoding/json"
	"strings"

	"rhea/internal/core"
)

// One purchase invoice is one fact, whichever door it came through (the
// whole-tree review of 2026-10-21, D2): the KSeF reader keyed an invoice
// "ksef_inbound/…" and the PDF reader "pdf_invoice/…", so a supplier's
// invoice pulled from KSeF and e-mailed as a PDF booked twice. Every reader
// of purchase invoices keys one here, on the seller's tax id and the
// number, each written one way.

// PurchaseInvoiceKey is a purchase invoice's identity, or "" when the
// seller's tax id or the number is missing (the reader keys its file then).
func PurchaseInvoiceKey(taxID, number string) string {
	id, num := taxIDOf(taxID), numberOf(number)
	if id == "" || num == "" {
		return ""
	}
	return "purchase_invoice/" + id + "/" + num
}

// taxIDOf writes a tax id one way: its letters and digits, upper case,
// without the two-letter country prefix an EU VAT id carries — "PL
// 805-808-30-55" and "8058083055" are one seller.
func taxIDOf(s string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(s) {
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	id := b.String()
	if len(id) > 2 && isUpper(id[0]) && isUpper(id[1]) && id[2] >= '0' && id[2] <= '9' {
		id = id[2:]
	}
	return id
}

func isUpper(c byte) bool { return c >= 'A' && c <= 'Z' }

// numberOf writes an invoice number one way: upper case, spaces collapsed.
func numberOf(s string) string { return strings.ToUpper(strings.Join(strings.Fields(s), " ")) }

// LoggedInvoices is every purchase invoice the pass's company already holds,
// by identity — whichever door logged it and under whatever key, the keys
// before PurchaseInvoiceKey included.
func LoggedInvoices(ctx context.Context, reads Reads) (map[string]bool, error) {
	logged := map[string]bool{}
	er, ok := reads.(interface {
		EventsOfTypes(ctx context.Context, types []string) ([]core.Event, error)
	})
	if !ok {
		return logged, nil
	}
	evs, err := er.EventsOfTypes(ctx, []string{"purchase.invoice.received"})
	if err != nil {
		return nil, err
	}
	for _, e := range evs {
		var p struct {
			SupplierNIP string `json:"supplier_nip"`
			Number      string `json:"number"`
		}
		if json.Unmarshal(e.Payload, &p) != nil {
			continue
		}
		if k := PurchaseInvoiceKey(p.SupplierNIP, p.Number); k != "" {
			logged[k] = true
		}
	}
	return logged, nil
}
