package shell_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"rhea/internal/agent"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// Authoring is a conversation (KK, 2026-10-10): a rejected bundle carries
// its reason on the record, and the next ask — "draft again" — shows the
// agent what was refused and why. The redraft names what it answers, every
// bundle keeps its question, and the record reads as the exchange it was.
func TestRedraftAnswersTheRejection(t *testing.T) {
	ctx := context.Background()
	s := storetest.New(t)
	x := &exec.Executor{Store: s}
	var seen []string // every request the agent received
	calls := 0
	a := &agent.Agent{Complete: func(ctx context.Context, system, user string) (string, error) {
		seen = append(seen, user)
		calls++
		if calls == 1 {
			return `{"bundle_id":"memos","description":"memos are kept","warrant":{"basis":"client"},
			  "object_types":[{"name":"memo","domain":"work","fields":[{"name":"text","type":"string","required":true}]}]}`, nil
		}
		return `{"bundle_id":"notes","description":"notes are kept","warrant":{"basis":"client"},
		  "object_types":[{"name":"note","domain":"work","fields":[{"name":"text","type":"string","required":true}]}]}`, nil
	}}
	srv := &shell.Server{Store: s, Exec: x, Agent: a, DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	post := func(path string, body any, out any) int {
		t.Helper()
		b, _ := json.Marshal(body)
		res, err := ts.Client().Post(ts.URL+path, "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if out != nil {
			json.NewDecoder(res.Body).Decode(out)
		}
		return res.StatusCode
	}

	// 1. The first ask lands a bundle that keeps its question.
	var first struct {
		ID       string `json:"bundle_id"`
		Question string `json:"question"`
	}
	if code := post("/api/bundles/draft", map[string]any{"intent": "we keep short memos"}, &first); code != 200 {
		t.Fatalf("first draft: %d", code)
	}
	if first.ID != "memos" || first.Question != "we keep short memos" {
		t.Fatalf("first bundle = %+v", first)
	}
	// A draft that was not rejected cannot be "drafted again".
	if code := post("/api/bundles/draft", map[string]any{"after": "memos"}, nil); code != 400 {
		t.Fatalf("redraft of an undecided bundle: %d, want 400", code)
	}

	// 2. Rejected with a reason, on the record.
	if code := post("/api/bundles/memos/reject", map[string]any{"reason": "call them notes, not memos", "rejected_by": "krzysztof"}, nil); code != 200 {
		t.Fatalf("reject: %d", code)
	}
	var listed []struct {
		ID        string `json:"bundle_id"`
		Status    string `json:"status"`
		Rejection *struct {
			Reason string `json:"reason"`
			By     string `json:"by"`
		} `json:"rejection"`
	}
	res, err := ts.Client().Get(ts.URL + "/api/bundles")
	if err != nil {
		t.Fatal(err)
	}
	json.NewDecoder(res.Body).Decode(&listed)
	res.Body.Close()
	if len(listed) != 1 || listed[0].Status != "superseded" || listed[0].Rejection == nil ||
		listed[0].Rejection.Reason != "call them notes, not memos" || listed[0].Rejection.By != "krzysztof" {
		t.Fatalf("listed = %+v", listed)
	}

	// 3. "Draft again" with no new words: the same question, and the agent
	// sees the refused proposal with its reason.
	var second struct {
		ID       string `json:"bundle_id"`
		Question string `json:"question"`
		After    string `json:"after"`
	}
	if code := post("/api/bundles/draft", map[string]any{"after": "memos"}, &second); code != 200 {
		t.Fatalf("redraft: %d", code)
	}
	if second.ID != "notes" || second.Question != "we keep short memos" || second.After != "memos" {
		t.Fatalf("second bundle = %+v", second)
	}
	if len(seen) != 2 {
		t.Fatalf("agent asked %d times", len(seen))
	}
	ask := seen[1]
	for _, want := range []string{"Refused by krzysztof: call them notes, not memos", `"bundle_id":"memos"`, `"name":"memo"`} {
		if !strings.Contains(ask, want) {
			t.Fatalf("the redraft ask lacks %q:\n%s", want, ask)
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(ask), "User intent: we keep short memos") {
		t.Fatalf("the ask must end with the question:\n%s", ask)
	}
	// 4. A second rejection: the chain reads oldest first, from the record
	// alone, and carries the question the exchange began with.
	if code := post("/api/bundles/notes/reject", map[string]any{"reason": "and keep the author", "rejected_by": "krzysztof"}, nil); code != 200 {
		t.Fatalf("reject notes: %d", code)
	}
	conv, question, err := shell.Conversation(ctx, s, "notes")
	if err != nil {
		t.Fatal(err)
	}
	if question != "we keep short memos" || len(conv) != 2 ||
		conv[0].Reason != "call them notes, not memos" || conv[1].Reason != "and keep the author" ||
		!strings.Contains(string(conv[1].Proposal), `"name":"note"`) {
		t.Fatalf("conversation = %+v, question %q", conv, question)
	}
}
