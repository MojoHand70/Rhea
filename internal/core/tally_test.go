package core

import (
	"strings"
	"testing"
)

func tallyCatalog() Catalog {
	types := map[string]ObjectType{
		"posting": {Name: "posting", Fields: []FieldDef{
			{Name: "account", Type: "ref<account>"}, {Name: "side", Type: "enum", Values: []string{"debit", "credit"}},
			{Name: "amount", Type: "money"}, {Name: "date", Type: "date"}, {Name: "book", Type: "string"}}},
		"account": {Name: "account", Fields: []FieldDef{{Name: "code", Type: "string"}, {Name: "type", Type: "enum", Values: []string{"revenue", "expense"}}}},
		"sales_invoice": {Name: "sales_invoice", Fields: []FieldDef{
			{Name: "buyer", Type: "ref<company>"}, {Name: "net", Type: "money"}, {Name: "issue_date", Type: "date"}, {Name: "number", Type: "string"}}},
		"company": {Name: "company", Fields: []FieldDef{{Name: "name", Type: "string"}}},
	}
	return func(n string) (ObjectType, bool) { t, ok := types[n]; return t, ok }
}

// TestTallyRefuses: a drafted tally names only what the catalog holds and
// says only what the kernel can add up; anything else is refused before a
// query is written.
func TestTallyRefuses(t *testing.T) {
	cat := tallyCatalog()
	ok := TallySpec{Of: "posting", Value: TallyDebitCredit, By: []TallyGroup{{Field: "date", By: "month", Label: "Miesiąc"}},
		Lines: []TallyLine{{Line: "Przychody", Value: TallyCreditDebit, Where: []TallyCond{{Field: "account.type", Op: "eq", Value: "revenue"}}},
			{Line: "Koszty", Op: "-", Where: []TallyCond{{Field: "account.type", Op: "eq", Value: "expense"}}}, {Line: "Wynik", Op: "="}}}
	if err := ok.Check(cat); err != nil {
		t.Fatalf("the monthly result: %v", err)
	}
	bad := map[string]TallySpec{
		"no such type":         {Of: "salary", Value: "amount"},
		"no amount":            {Of: "posting"},
		"amount not money":     {Of: "sales_invoice", Value: "number"},
		"signed without sides": {Of: "sales_invoice", Value: TallyDebitCredit},
		"amount through link":  {Of: "posting", Value: "account.code"},
		"no such field":        {Of: "sales_invoice", Value: "net", Where: []TallyCond{{Field: "seller", Op: "eq", Value: "x"}}},
		"link not a link":      {Of: "sales_invoice", Value: "net", Where: []TallyCond{{Field: "number.name", Op: "eq", Value: "x"}}},
		"two steps":            {Of: "posting", Value: "amount", Where: []TallyCond{{Field: "account.type.x", Op: "eq", Value: "x"}}},
		"a name that is code":  {Of: "posting", Value: "amount", Where: []TallyCond{{Field: "book') OR 1=1 --", Op: "eq", Value: "x"}}},
		"unknown op":           {Of: "posting", Value: "amount", Where: []TallyCond{{Field: "book", Op: "like", Value: "%"}}},
		"in without a list":    {Of: "posting", Value: "amount", Where: []TallyCond{{Field: "book", Op: "in", Value: "x"}}},
		"value not scalar":     {Of: "posting", Value: "amount", Where: []TallyCond{{Field: "book", Op: "eq", Value: map[string]any{"a": 1}}}},
		"month of a string":    {Of: "posting", Value: "amount", By: []TallyGroup{{Field: "book", By: "month", Label: "x"}}},
		"by year":              {Of: "posting", Value: "amount", By: []TallyGroup{{Field: "date", By: "year", Label: "x"}}},
		"group without label":  {Of: "posting", Value: "amount", By: []TallyGroup{{Field: "book"}}},
		"three groupings": {Of: "sales_invoice", Value: "net", By: []TallyGroup{{Field: "buyer", Label: "a"}, {Field: "number", Label: "b"},
			{Field: "issue_date", By: "month", Label: "c"}}},
		"same grouping twice": {Of: "posting", Value: "amount", By: []TallyGroup{{Field: "book", Label: "a"}, {Field: "book", Label: "b"}}},
		"sum first":           {Of: "posting", Value: "amount", Lines: []TallyLine{{Line: "Wynik", Op: "="}}},
		"sum with a where":    {Of: "posting", Value: "amount", Lines: []TallyLine{{Line: "A"}, {Line: "B", Op: "=", Where: []TallyCond{{Field: "book", Op: "exists"}}}}},
		"line without name":   {Of: "posting", Value: "amount", Lines: []TallyLine{{Line: " "}}},
		"two lines one name":  {Of: "posting", Value: "amount", Lines: []TallyLine{{Line: "A"}, {Line: "A"}}},
		"line op":             {Of: "posting", Value: "amount", Lines: []TallyLine{{Line: "A", Op: "*"}}},
	}
	for name, s := range bad {
		if err := s.Check(cat); err == nil {
			t.Errorf("%s: accepted", name)
		}
		if _, _, err := s.Compile(cat); err == nil {
			t.Errorf("%s: compiled", name)
		}
	}
}

// TestTallyQuotes: a value is a literal, whatever it says, and the
// columns are the kernel's, labelled by the tally.
func TestTallyQuotes(t *testing.T) {
	s := TallySpec{Of: "sales_invoice", Value: "net", Where: []TallyCond{{Field: "buyer.name", Op: "eq", Value: "O'Brien'); DROP TABLE objects; --"}},
		By: []TallyGroup{{Field: "buyer.name", Label: "Klient"}, {Field: "issue_date", By: "month", Label: "Miesiąc"}}, AmountLabel: "Netto"}
	sql, cols, err := s.Compile(tallyCatalog())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, "'O''Brien''); DROP TABLE objects; --'") || strings.Count(sql, "DROP") != 1 {
		t.Fatalf("the value is not one literal: %s", sql)
	}
	if len(cols) != 3 || cols[0].Field != "buyer_name" || cols[0].Label != "Klient" || cols[1].Field != "month" || cols[2].Field != "amount" ||
		cols[2].Type != "money" || cols[2].Label != "Netto" || len(cols[2].SumOf) != 1 || cols[2].SumOf[0] != "net" {
		t.Fatalf("columns: %+v", cols)
	}
	if !strings.Contains(sql, "LEFT JOIN objects l0 ON l0.object_id = json_extract_string(x.state,'$.buyer')") || !strings.Contains(sql, " AS "+RowsOfColumn+" ") {
		t.Fatalf("the link or the lineage is missing: %s", sql)
	}
}
