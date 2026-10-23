package shell_test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"rhea/internal/adapter"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// a register as a market would provide it: one company it knows, one tax id
// it refuses
type fakeRegister struct{}

func (fakeRegister) Name() string { return "test-register" }
func (fakeRegister) Look(_ context.Context, id string) (adapter.RegisterEntry, error) {
	if id == "9999999999" {
		return adapter.RegisterEntry{}, errors.New("the register is closed tonight")
	}
	return adapter.RegisterEntry{Register: "test-register", TaxID: id, Found: id == "5260250995", Fields: map[string]string{"name": "Alfa sp. z o.o."}}, nil
}

// TestARegisterSaysWhoATaxIDIs: the registers a market provides are named,
// and one is asked by tax id however the person wrote it; an unknown
// register, a missing id and a register's own refusal are said so. Nothing
// is written: the log is as it was.
func TestARegisterSaysWhoATaxIDIs(t *testing.T) {
	adapter.ProvideCompanyRegister(fakeRegister{})
	s := storetest.New(t)
	logged := func() int {
		t.Helper()
		n := 0
		for _, kind := range []string{"raw", "derived"} {
			evs, err := s.EventsByKind(context.Background(), kind)
			if err != nil {
				t.Fatal(err)
			}
			n += len(evs)
		}
		return n
	}
	before := logged()
	ts := serve(t, (&shell.Server{Store: s}).Handler())
	get := func(path string, out any) int {
		t.Helper()
		res, err := ts.Client().Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if out != nil {
			json.NewDecoder(res.Body).Decode(out)
		}
		return res.StatusCode
	}
	var names struct{ Registers []string }
	if code := get("/api/registers", &names); code != 200 || !slices.Contains(names.Registers, "test-register") {
		t.Fatalf("the registers this binary reads: %d %v", code, names.Registers)
	}
	var e adapter.RegisterEntry
	if code := get("/api/registers/test-register?id=PL%20526-025-09-95", &e); code != 200 || !e.Found || e.TaxID != "5260250995" || e.Fields["name"] != "Alfa sp. z o.o." {
		t.Fatalf("a tax id as a person writes it: %d %+v", code, e)
	}
	var refused struct{ Error string }
	for _, c := range []struct {
		path string
		want int
	}{{"/api/registers/no-such?id=5260250995", 404}, {"/api/registers/test-register", 400}, {"/api/registers/test-register?id=9999999999", 424}} {
		if code := get(c.path, &refused); code != c.want {
			t.Fatalf("%s: %d, want %d (%s)", c.path, code, c.want, refused.Error)
		}
	}
	if refused.Error != "the register is closed tonight" {
		t.Fatalf("a register's refusal in its own words: %q", refused.Error)
	}
	if after := logged(); after != before {
		t.Fatalf("a lookup writes nothing: %d facts before, %d after", before, after)
	}
}
