package adapter

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"rhea/internal/core"
)

// KSeFClient is the boundary to the national e-invoice API: submit one
// invoice, get back the KSeF element reference and the UPO (urzędowe
// poświadczenie odbioru). The real API stays outside the experiment; demos
// and tests plug the Fake.
type KSeFClient interface {
	Submit(ctx context.Context, invoiceNumber string, gross int64, currency string) (ksefRef, upo string, err error)
}

// KSeF is the Poland pack's statutory adapter: every sales_invoice without a
// ksef_submission gets submitted, and the evidence comes back as a
// ksef.invoice.submitted event keyed by the invoice's object id — so a retry
// after a crash re-submits into a dedup no-op.
type KSeF struct {
	Client KSeFClient
	Today  func() string // business date of submissions; defaults to the wall clock
}

func (k KSeF) Name() string { return "ksef" }

func (k KSeF) Pass(ctx context.Context, reads Reads) ([]core.Event, error) {
	invoices, err := reads.ObjectsByType(ctx, "sales_invoice")
	if err != nil {
		return nil, err
	}
	subs, err := reads.ObjectsByType(ctx, "ksef_submission")
	if err != nil {
		return nil, err
	}
	done := make(map[string]bool, len(subs))
	for _, s := range subs {
		if id, ok := s.State["invoice"].(string); ok {
			done[id] = true
		}
	}
	today := time.Now().Format("2006-01-02")
	if k.Today != nil {
		today = k.Today()
	}
	var out []core.Event
	for _, inv := range invoices {
		if done[inv.ID] {
			continue
		}
		number, _ := inv.State["number"].(string)
		currency, _ := inv.State["currency"].(string)
		gross, ok := minorUnits(inv.State["gross"])
		if number == "" || !ok {
			return out, fmt.Errorf("invoice %s has no usable number/gross", inv.ID)
		}
		ksefRef, upo, err := k.Client.Submit(ctx, number, gross, currency)
		if err != nil {
			// Events gathered so far still go to the log; the failed invoice
			// is simply still owed next pass.
			return out, fmt.Errorf("submit %s: %w", number, err)
		}
		payload := fmt.Sprintf(`{"number":%q,"ksef_ref":%q,"upo":%q,"date":%q}`,
			number, ksefRef, upo, today)
		out = append(out, core.Event{
			Type: "ksef.invoice.submitted", OccurredAt: today,
			Payload: []byte(payload), DedupKey: "ksef/submit/" + inv.ID,
		})
	}
	return out, nil
}

// minorUnits reads a money value out of projected object state, where JSONB
// hands numbers back as float64.
func minorUnits(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	}
	return 0, false
}

// FakeKSeF answers like the national system would, deterministically from
// the invoice number — demos and replays stay reproducible.
type FakeKSeF struct{}

func (FakeKSeF) Submit(_ context.Context, number string, _ int64, _ string) (string, string, error) {
	h := sha256.Sum256([]byte(number))
	return fmt.Sprintf("KSEF-%X", h[:6]), fmt.Sprintf("UPO-%X", h[6:12]), nil
}
