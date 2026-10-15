package adapter

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"rhea/internal/exec"
	"rhea/internal/store"
)

// Issue renders a writer's file and hands it over through the
// issue_document door: the bytes kept under their hash, the event the
// evidence, the documents pack's rule making the issued_document. The
// writer and its parameters are recorded so the file can be rendered again
// and compared — a later difference is a correction, a new fact.
func Issue(ctx context.Context, s *store.Store, x *exec.Executor, writer string, params map[string]string, actor, occurredAt string) (File, exec.Triggered, error) {
	f, err := Render(ctx, s, writer, params)
	if err != nil {
		return File{}, exec.Triggered{}, err
	}
	inputs := map[string]any{
		"writer": writer, "params": ParamsText(params),
		"file": map[string]any{"name": f.Name, "media_type": f.MediaType, "bytes": f.Bytes},
	}
	if occurredAt != "" {
		inputs["occurred_at"] = occurredAt
	}
	t, err := x.TriggerActivity(ctx, "issue_document", inputs, actor, "")
	if err != nil {
		return f, exec.Triggered{}, err
	}
	return f, t, nil
}

// ParamsText spells parameters one way — sorted "k=v" pairs joined by
// commas — so the same request is the same text on the record.
func ParamsText(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	return strings.Join(parts, ",")
}

// ParseParams reads "k=v" arguments, the CLI's and the API's spelling.
func ParseParams(args []string) (map[string]string, error) {
	out := map[string]string{}
	for _, a := range args {
		k, v, ok := strings.Cut(a, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("parameter %q is not key=value", a)
		}
		out[k] = v
	}
	return out, nil
}
