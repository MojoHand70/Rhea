package core

import (
	"strings"
	"testing"
)

// The rights vocabulary (DIRECTION 2026-10-19, "People and rights"): one
// word per right, and a wildcard over doors that stops at the people domain
// — the right to do the work never implies the right to say who may do it.
func TestRightsVocabulary(t *testing.T) {
	for _, bad := range []string{"", "door", "door:", "wizard:x", "domain:*", "state:open", "door:a b", "sees:*"} {
		if _, err := ParseRight(bad); err == nil {
			t.Errorf("%q must be refused", bad)
		}
	}
	rights, err := ParseRights("door:* view:* type:* domain:people  state:order/approved sees:sales approves:sales")
	if err != nil || len(rights) != 7 {
		t.Fatalf("parse: %v %v", rights, err)
	}
	if rights[0].String() != "door:*" || rights[3].String() != "domain:people" {
		t.Fatalf("strings: %v", rights)
	}

	cases := []struct {
		right           string
		kind, name, dom string
		covers          bool
	}{
		{"*", "door", "grant_membership", "people", true},
		{"door:*", "door", "close_period", "periods", true},
		{"door:*", "door", "grant_membership", "people", false}, // KSeF: managing rights is its own grant
		{"door:grant_membership", "door", "grant_membership", "people", true},
		{"domain:people", "door", "grant_membership", "people", true},
		{"domain:people", "view", "who-can-do-what", "people", true},
		{"domain:people", "door", "close_period", "periods", false},
		{"domain:people", "sees", "sales", "people", false}, // a domain grants doors, views and types
		{"view:*", "view", "who-can-do-what", "people", true},
		{"type:*", "type", "membership", "people", true},
		{"type:invoice", "type", "invoice", "finance", true},
		{"type:invoice", "view", "invoice", "finance", false},
		{"sees:sales", "sees", "sales", "", true},
		{"sees:sales", "approves", "sales", "", false},
		{"state:order/approved", "state", "order/approved", "sales", true},
	}
	for _, c := range cases {
		r, err := ParseRight(c.right)
		if err != nil {
			t.Fatal(err)
		}
		if got := r.Covers(c.kind, c.name, c.dom); got != c.covers {
			t.Errorf("%s covers %s:%s in %s = %v, want %v", c.right, c.kind, c.name, c.dom, got, c.covers)
		}
	}
	if !CoveredBy(rights, "door", "close_period", "periods") || !CoveredBy(rights, "door", "grant_mandate", "people") {
		t.Fatal("CoveredBy with domain:people")
	}
	if books, _ := ParseRights("door:* view:* type:*"); CoveredBy(books, "door", "grant_mandate", "people") || !CoveredBy(books, "view", "who-can-do-what", "people") {
		t.Fatal("the books' rights never reach the people doors, and seeing who may do what is not managing it")
	}
}

// Who is written in actor kinds; a person is anyone who is not the model or
// the machine's own hands.
func TestActorKindAndWho(t *testing.T) {
	for actor, kind := range map[string]string{"kk": ActorHuman, "shell:kk": ActorHuman, "cli:krzysztof": ActorHuman,
		"agent:claude-sonnet-4-6": ActorAgent, "kernel": ActorAdapter, "adapter:ksef": ActorAdapter, "channel:clock": ActorAdapter} {
		if got := ActorKind(actor); got != kind {
			t.Errorf("%s: %s, want %s", actor, got, kind)
		}
	}
	sp := ActivitySpec{Who: []string{"adapter"}}
	if sp.Admits(ActorHuman) || !sp.Admits(ActorAdapter) || !(ActivitySpec{}).Admits(ActorHuman) {
		t.Fatal("Admits")
	}
	for _, f := range []FieldDef{{Name: "approved_by", Type: "string"}, {Name: "by", Type: "string"}, {Name: "granted_by", Type: "string"}} {
		if !SignedInput(f) {
			t.Errorf("%s is signed", f.Name)
		}
	}
	for _, f := range []FieldDef{{Name: "by", Type: "int"}, {Name: "reason", Type: "string"}, {Name: "standby", Type: "string"}} {
		if SignedInput(f) {
			t.Errorf("%s is not signed", f.Name)
		}
	}
	if !ReservedEventType(EventAccessRefused) || !strings.HasPrefix(EventAccessRecorded, "access.") {
		t.Fatal("the access namespace is the kernel's")
	}
}
