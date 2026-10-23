package shell_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"testing"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/shell"
	"rhea/internal/store/storetest"
)

// The file is the fact, over HTTP: a document comes in through the derived
// form of the receive_document door (its bytes as base64), and comes back
// out by its hash with the media type it was received as.
func TestIntakeOverHTTP(t *testing.T) {
	s := storetest.New(t)
	srv := &shell.Server{Store: s, Exec: &exec.Executor{Store: s},
		DuckPath: filepath.Join(t.TempDir(), "rhea.duckdb")}
	ts := serve(t, srv.Handler())

	data := []byte("%PDF-1.7 a scan")
	hash := core.HashBytes(data)
	body, _ := json.Marshal(map[string]any{"inputs": map[string]any{
		"file":   map[string]any{"name": "scan.pdf", "media_type": "application/pdf", "content_base64": base64.StdEncoding.EncodeToString(data)},
		"source": "scan",
	}})
	res, err := ts.Client().Post(ts.URL+"/api/activities/receive_document/trigger", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	json.NewDecoder(res.Body).Decode(&out)
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("trigger: %d %v", res.StatusCode, out)
	}

	res, err = ts.Client().Get(ts.URL + "/api/blobs/" + hash)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 || res.Header.Get("Content-Type") != "application/pdf" || string(got) != string(data) {
		t.Fatalf("blob: %d %s %q", res.StatusCode, res.Header.Get("Content-Type"), got)
	}
	if res, _ := ts.Client().Get(ts.URL + "/api/blobs/sha256:nothing"); res.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown hash: %d", res.StatusCode)
	}

	// The worklist shows the fact with the file it names.
	res, _ = ts.Client().Get(ts.URL + "/api/worklist")
	var wl struct {
		Events []struct {
			Type    string          `json:"event_type"`
			Payload json.RawMessage `json:"payload"`
		} `json:"events"`
	}
	json.NewDecoder(res.Body).Decode(&wl)
	res.Body.Close()
	if len(wl.Events) != 1 || wl.Events[0].Type != core.EventDocumentReceived || !bytes.Contains(wl.Events[0].Payload, []byte(hash)) {
		t.Fatalf("worklist = %+v", wl.Events)
	}
}
