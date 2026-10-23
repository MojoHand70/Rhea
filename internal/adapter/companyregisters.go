package adapter

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// A CompanyRegister is a public record of companies a market reads by tax id — a
// statistics office's business register, a tax office's list of VAT payers —
// so a person adding a company types its number and confirms what comes
// back, and nobody retypes an address (DIRECTION 2026-10-19, "People and
// rights": onboarding). A lookup writes nothing and books nothing: what the
// person confirms reaches the log through a door, the register's answer
// named in it as the source, so replay never asks the register again.
// Markets provide their registers from their own package at init; the
// kernel knows none by name (invariant 8).
type CompanyRegister interface {
	Name() string
	Look(ctx context.Context, taxID string) (RegisterEntry, error)
}

// RegisterEntry is what a register says of one tax id. Fields carry the company in a
// small shared vocabulary — name, short_name, address, street, building,
// unit, postcode, city, country, legal_form (natural_person, legal_entity, or
// empty where the register does not decide it), legal_form_name, vat_status,
// accounts (comma-separated), registered_on, closed_on — and whatever the
// register adds under its own names (a REGON, a KRS number). Ref is the
// register's own mark of this answer, kept as evidence when it has one.
type RegisterEntry struct {
	Register string            `json:"register"`
	TaxID    string            `json:"tax_id"`
	Found    bool              `json:"found"`
	Fields   map[string]string `json:"fields"`
	Ref      string            `json:"ref,omitempty"`
	ReadAt   string            `json:"read_at"`
}

var companyRegisters = map[string]CompanyRegister{}

// ProvideCompanyRegister makes a market's register readable by its name.
func ProvideCompanyRegister(r CompanyRegister) { companyRegisters[r.Name()] = r }

// CompanyRegisterNames lists the registers this binary can read, by name.
func CompanyRegisterNames() []string {
	out := make([]string, 0, len(companyRegisters))
	for n := range companyRegisters {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// LookUpCompany asks one register what it says of a tax id: digits only, as every
// register keeps them ("526-025-09-95" and "PL5260250995" are the same).
func LookUpCompany(ctx context.Context, name, taxID string) (RegisterEntry, error) {
	r, ok := companyRegisters[name]
	if !ok {
		return RegisterEntry{}, fmt.Errorf("no register named %q in this binary", name)
	}
	var b strings.Builder
	for _, c := range taxID {
		if c >= '0' && c <= '9' {
			b.WriteRune(c)
		}
	}
	if b.Len() == 0 {
		return RegisterEntry{}, fmt.Errorf("a tax id has digits: %q", taxID)
	}
	return r.Look(ctx, b.String())
}
