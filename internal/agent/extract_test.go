package agent

import (
	"encoding/json"
	"strings"
	"testing"
)

// A model that prefaces its answer with prose carrying a JSON snippet still
// has the answer as the largest object it wrote: the decoder takes that one,
// never the snippet (the habit the round of 2026-10-13 showed).
func TestExtractJSONTakesTheAnswerNotTheSnippet(t *testing.T) {
	raw := `I will try {"strategy": "reference"} first, as in {"rule_id": "example", "description": "a snippet"}. Here is the bundle:
{"bundle_id": "bank", "description": "d", "rules": [{"rule_id": "r", "spec": {"match": {"event_type": "x"}}}]}
Done.`
	got := extractJSON(raw)
	if !strings.HasPrefix(got, `{"bundle_id"`) || !strings.HasSuffix(got, `}]}`) {
		t.Fatalf("took the wrong object: %s", got)
	}
	if extractJSON(`no json here`) != `no json here` || extractJSON(`{"a": 1}`) != `{"a": 1}` {
		t.Fatal("plain cases")
	}
	// Nothing decodes: the text from the first brace, for the real error.
	if extractJSON(`x {"a": `) != `{"a": ` {
		t.Fatalf("broken json: %q", extractJSON(`x {"a": `))
	}
	// A bundle that breaks in the middle is not replaced by the largest
	// piece that happens to parse: the syntax error must surface.
	broken := `{"bundle_id": "b", "description": "d", "object_types": [{"name": "t", "fields": [{"name": "x", "type": "string"}]}], "rules": [{"rule_id": "r", <side>: 1}]}`
	if got := extractJSON(broken); got != broken {
		t.Fatalf("a broken bundle must surface its error, got %q", got)
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(extractJSON(broken)), &v); err == nil || !strings.Contains(err.Error(), "invalid character '<'") {
		t.Fatalf("the real error: %v", err)
	}
}
