package sim_test

import (
	"bytes"
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/sim"
	"rhea/internal/store/storetest"
)

// The reader's view of a generated FA file: enough of the subset to check
// it against the fact it stands for.
type faRead struct {
	XMLName  xml.Name `xml:"Faktura"`
	Podmiot1 struct {
		NIP string `xml:"DaneIdentyfikacyjne>NIP"`
	} `xml:"Podmiot1"`
	Podmiot2 struct {
		NIP string `xml:"DaneIdentyfikacyjne>NIP"`
	} `xml:"Podmiot2"`
	Fa struct {
		Currency string   `xml:"KodWaluty"`
		Date     string   `xml:"P_1"`
		Number   string   `xml:"P_2"`
		WZ       []string `xml:"WZ"`
		Net      string   `xml:"P_13_1"`
		VAT      string   `xml:"P_14_1"`
		Gross    string   `xml:"P_15"`
		Lines    []struct {
			Net string `xml:"P_11"`
		} `xml:"FaWiersz"`
		Term string `xml:"Platnosc>TerminPlatnosci>Termin"`
	} `xml:"Fa"`
}

func halfUp(n, d int64) int64 {
	q, r := n/d, n%d
	if 2*r >= d {
		q++
	}
	return q
}

// Documents carry the facts (DIRECTION 2026-10-10, step 3): every purchase
// invoice from a Polish supplier is one FA file saying exactly what the
// event says; every month is one MT940 whose lines are the month's bank
// lines and whose balances add up; the same persona yields the same bytes.
func TestDocumentsCarryTheFacts(t *testing.T) {
	p, _, err := sim.Load(filepath.Join("..", "..", "testdata", "customers", "nordwind.json"))
	if err != nil {
		t.Fatal(err)
	}
	p.Months = 3
	f, err := sim.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := sim.Generate(p)
	if len(again.Documents) != len(f.Documents) {
		t.Fatal("the generator is not deterministic")
	}
	for i := range f.Documents {
		if f.Documents[i].Path != again.Documents[i].Path || !bytes.Equal(f.Documents[i].Bytes, again.Documents[i].Bytes) {
			t.Fatalf("document %d differs between runs", i)
		}
	}

	covered := map[int]bool{}
	ksef, bank := 0, 0
	for _, d := range f.Documents {
		for _, i := range d.Events {
			covered[i] = true
		}
		switch d.Source {
		case "ksef":
			ksef++
			if len(d.Events) != 1 || d.MediaType != "application/xml" || !strings.HasPrefix(d.Path, d.Date[:7]+"/ksef/") {
				t.Fatalf("ksef document %+v", d)
			}
			ev := f.Events[d.Events[0]]
			var inv faRead
			if err := xml.Unmarshal(d.Bytes, &inv); err != nil {
				t.Fatalf("%s: %v", d.Path, err)
			}
			if inv.XMLName.Space != sim.FA3Namespace {
				t.Fatalf("%s: namespace %q", d.Path, inv.XMLName.Space)
			}
			net, _ := core.ParseMoney(ev.Payload["net"].(string))
			vat := halfUp(net*23, 100)
			if inv.Fa.Number != ev.Payload["number"] || inv.Fa.Date != ev.Payload["invoice_date"] ||
				"PL"+inv.Podmiot1.NIP != ev.Payload["supplier_nip"] || "PL"+inv.Podmiot2.NIP != p.Company.VATID ||
				inv.Fa.Net != core.FormatMoney(net) || inv.Fa.VAT != core.FormatMoney(vat) ||
				inv.Fa.Gross != core.FormatMoney(net+vat) || inv.Fa.Currency != "PLN" {
				t.Fatalf("%s does not say what the event says: %+v vs %v", d.Path, inv.Fa, ev.Payload)
			}
			var lines int64
			for _, l := range inv.Fa.Lines {
				v, _ := core.ParseMoney(l.Net)
				lines += v
			}
			if lines != net {
				t.Fatalf("%s: lines sum to %d, net is %d", d.Path, lines, net)
			}
			if wz, ok := ev.Payload["for_document"].(string); ok && (len(inv.Fa.WZ) != 1 || inv.Fa.WZ[0] != wz) {
				t.Fatalf("%s: WZ = %v, want %s", d.Path, inv.Fa.WZ, wz)
			}
			ref := sim.KSeFReference(inv.Podmiot1.NIP, inv.Fa.Date, inv.Fa.Number)
			if filepath.Base(d.Path) != ref+".xml" {
				t.Fatalf("%s: not named by its KSeF reference %s", d.Path, ref)
			}
		case "bank":
			bank++
			text := string(d.Bytes)
			if !strings.HasPrefix(text, ":20:STMT") || !strings.Contains(text, ":25:"+p.Company.IBAN+"\r\n") {
				t.Fatalf("%s: header %q", d.Path, text[:40])
			}
			var sum int64
			lines := 0
			for _, ln := range strings.Split(text, "\r\n") {
				if strings.HasPrefix(ln, ":61:") {
					lines++
					amt := ln[15:strings.Index(ln, "NTRF")] // after :61:, the dates and the C/D mark
					v, err := core.ParseMoney(strings.ReplaceAll(amt, ",", "."))
					if err != nil {
						t.Fatalf("%s: amount %q: %v", d.Path, amt, err)
					}
					if ln[14] == 'D' {
						v = -v
					}
					sum += v
				}
			}
			if lines != len(d.Events) {
				t.Fatalf("%s: %d lines for %d events", d.Path, lines, len(d.Events))
			}
			var fromEvents int64
			for _, i := range d.Events {
				ev := f.Events[i]
				v, _ := core.ParseMoney(ev.Payload["amount"].(string))
				kind := "PRZELEW PRZYCHODZACY"
				if ev.Payload["direction"] == "debit" {
					v, kind = -v, "PRZELEW WYCHODZACY"
				}
				fromEvents += v
				ref := ev.Payload["reference"].(string)
				who := ev.Payload["counterparty"].(string)
				if !strings.Contains(text, ":86:"+kind+"|"+ref+"|"+who+"|"+ev.Payload["counterparty_account"].(string)) {
					t.Fatalf("%s: no :86: line for %s from %s", d.Path, ref, who)
				}
				if ev.OccurredAt[:7] != d.Date[:7] {
					t.Fatalf("%s carries a line from %s", d.Path, ev.OccurredAt)
				}
			}
			if sum != fromEvents {
				t.Fatalf("%s: lines sum to %d, events to %d", d.Path, sum, fromEvents)
			}
		default:
			t.Fatalf("unknown source %q", d.Source)
		}
	}
	if bank != p.Months {
		t.Fatalf("%d statements for %d months", bank, p.Months)
	}
	// Every Polish supplier's invoice is a KSeF document; the German one is
	// not — it arrives as a PDF, in phase 2.
	polish, foreign := 0, 0
	for i, ev := range f.Events {
		if ev.EventType != "purchase.invoice.received" {
			continue
		}
		if strings.HasPrefix(ev.Payload["supplier_nip"].(string), "PL") {
			polish++
			if !covered[i] {
				t.Fatalf("purchase invoice %v has no document", ev.Payload["number"])
			}
		} else {
			foreign++
			if covered[i] {
				t.Fatalf("foreign invoice %v got a KSeF document", ev.Payload["number"])
			}
		}
	}
	if ksef != polish || polish == 0 || foreign == 0 {
		t.Fatalf("ksef documents %d, polish invoices %d, foreign %d", ksef, polish, foreign)
	}
	// Balances carry: the last closing is the opening plus every credit
	// minus every debit.
	last := f.Documents[len(f.Documents)-1]
	var credits int64
	for _, ev := range f.Events {
		if ev.EventType == "bank.statement.line" {
			v, _ := core.ParseMoney(ev.Payload["amount"].(string))
			if ev.Payload["direction"] == "debit" {
				v = -v
			}
			credits += v
		}
	}
	closing := strings.ReplaceAll(core.FormatMoney(sim.StatementOpening+credits), ".", ",")
	if !strings.Contains(string(last.Bytes), ":62F:C") || !strings.HasSuffix(strings.TrimSpace(string(last.Bytes)), "PLN"+closing+"\r\n-") {
		t.Fatalf("last statement's closing balance: %q", string(last.Bytes))
	}

	// On disk, under the persona's name.
	dir := t.TempDir()
	n, err := sim.WriteDocuments(dir, p, f.Documents)
	if err != nil || n != len(f.Documents) {
		t.Fatalf("write: %d %v", n, err)
	}
	if _, err := os.Stat(filepath.Join(dir, p.Name, f.Documents[0].Path)); err != nil {
		t.Fatal(err)
	}
}

// A generated file is a document like any other: through the door, kept
// under its hash, registered by the documents pack, waiting for a reader.
func TestGeneratedDocumentThroughTheDoor(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	sum, err := pack.Load(ctx, s, filepath.Join("..", "..", "packs", "common", "documents.json"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, errs, err := x.ApproveBundle(ctx, sum.Bundle, "krzysztof", "2026-01-01"); err != nil || len(errs) != 0 {
		t.Fatalf("approve: %v %v", err, errs)
	}
	p, _, _ := sim.Load(filepath.Join("..", "..", "testdata", "customers", "helios.json"))
	p.Months = 1
	f, err := sim.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Documents {
		tr, err := x.TriggerActivity(ctx, "receive_document", map[string]any{
			"file":   map[string]any{"name": filepath.Base(d.Path), "media_type": d.MediaType, "bytes": d.Bytes},
			"source": d.Source, "occurred_at": d.Date,
		}, "sim:"+p.Name, "")
		if err != nil || tr.Booked != 1 || len(tr.Errors) != 0 {
			t.Fatalf("%s: %+v %v", d.Path, tr, err)
		}
	}
	objs, _ := s.ObjectsByType(ctx, "received_document")
	if len(objs) != len(f.Documents) {
		t.Fatalf("received %d, generated %d", len(objs), len(f.Documents))
	}
}
