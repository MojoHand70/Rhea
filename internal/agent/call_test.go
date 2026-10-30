package agent_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"rhea/internal/agent"
)

// fakeAPI stands in for the model's API: it keeps the last request body and
// answers every call with the message it is given. The agent reaches it
// through ANTHROPIC_BASE_URL, the same switch a local model uses.
func fakeAPI(t *testing.T, answer string) (*agent.Agent, func() map[string]any) {
	t.Helper()
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = nil
		if err := json.Unmarshal(b, &body); err != nil {
			t.Errorf("request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, answer)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("ANTHROPIC_BASE_URL", srv.URL)
	t.Setenv("ANTHROPIC_API_KEY", "test-key")
	return agent.New(), func() map[string]any { return body }
}

const said = `{"id":"msg_1","type":"message","role":"assistant","model":"m",
  "content":[{"type":"text","text":"Dzień dobry."}],
  "stop_reason":"end_turn","stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":1}}`

// The host speaking on a Claude model: the 5.5 generation refuses to switch
// thinking off (an HTTP 400), so the call carries no thinking parameter and
// asks for low effort instead — the words still come in seconds.
func TestSpeakOnAClaudeModelAsksForLowEffort(t *testing.T) {
	a, last := fakeAPI(t, said)
	t.Setenv("RHEA_MODEL", "claude-sonnet-5-5")
	words, err := a.Speak(context.Background(), "You are the host.", "Speak.")
	if err != nil {
		t.Fatal(err)
	}
	if words != "Dzień dobry." {
		t.Fatalf("words = %q", words)
	}
	body := last()
	if th, ok := body["thinking"]; ok {
		t.Errorf("Speak sent thinking %v to a Claude model; the 5.5 generation answers 400 to switching it off", th)
	}
	oc, _ := body["output_config"].(map[string]any)
	if oc["effort"] != "low" {
		t.Errorf("Speak sent output_config %v; the host speaks at low effort", body["output_config"])
	}
	if body["model"] != "claude-sonnet-5-5" {
		t.Errorf("model = %v", body["model"])
	}
}

// A model that is not Claude (a local one behind ANTHROPIC_BASE_URL) is
// switched off as before: Ollama honours the parameter, and skips the
// reasoning pass that would make the morning a minute long.
func TestSpeakOnALocalModelSwitchesThinkingOff(t *testing.T) {
	a, last := fakeAPI(t, said)
	t.Setenv("RHEA_MODEL", "qwen3.6:27b")
	if _, err := a.Speak(context.Background(), "You are the host.", "Speak."); err != nil {
		t.Fatal(err)
	}
	body := last()
	th, _ := body["thinking"].(map[string]any)
	if th["type"] != "disabled" {
		t.Errorf("Speak sent thinking %v to a local model; it is switched off", body["thinking"])
	}
	if oc, ok := body["output_config"]; ok {
		t.Errorf("Speak sent output_config %v to a local model, which knows no effort", oc)
	}
}

// Drafting and reading want the model's own default: no thinking parameter,
// no effort — on any model.
func TestCompleteLeavesTheModelsDefault(t *testing.T) {
	a, last := fakeAPI(t, said)
	t.Setenv("RHEA_MODEL", "claude-sonnet-5-5")
	if _, err := a.Complete(context.Background(), "You draft rules.", "Draft."); err != nil {
		t.Fatal(err)
	}
	body := last()
	if th, ok := body["thinking"]; ok {
		t.Errorf("Complete sent thinking %v; the model's default stands", th)
	}
	if oc, ok := body["output_config"]; ok {
		t.Errorf("Complete sent output_config %v; the model's default stands", oc)
	}
}
