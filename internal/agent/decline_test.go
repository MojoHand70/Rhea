package agent_test

import (
	"context"
	"strings"
	"testing"
)

const declined = `{"id":"msg_2","type":"message","role":"assistant","model":"m","content":[],
  "stop_reason":"refusal","stop_details":{"type":"refusal","category":"general_harms","explanation":"This request was declined."},
  "stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":0}}`

// A decline is an answer with no words. It is an error naming the API's
// reason, never an empty draft that fails later as "no JSON" or a morning
// that falls silent without saying why.
func TestADeclineIsAnErrorWithItsReason(t *testing.T) {
	a, _ := fakeAPI(t, declined)
	t.Setenv("RHEA_MODEL", "claude-sonnet-5-5")
	_, err := a.Complete(context.Background(), "You draft rules.", "Draft.")
	if err == nil {
		t.Fatal("a declined answer returned no error")
	}
	for _, want := range []string{"declined", "general_harms", "This request was declined."} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not carry %q", err, want)
		}
	}
}
