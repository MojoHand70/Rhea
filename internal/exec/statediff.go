package exec

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"rhea/internal/core"
)

// StateDiff compares two readings of the object cache — before and after a
// replay — object by object, through the JSON form both the store and the
// derived events speak, and names every difference: an object lost, one
// that appeared, one whose state or provenance moved. Empty means replay
// reproduced the cache (invariant 4); rhea replay exits non-zero otherwise.
func StateDiff(before, after []core.Object) []string {
	index := func(objs []core.Object) map[string]core.Object {
		out := make(map[string]core.Object, len(objs))
		for _, o := range objs {
			b, _ := json.Marshal(o)
			var n core.Object
			_ = json.Unmarshal(b, &n)
			out[o.ID] = n
		}
		return out
	}
	was, is := index(before), index(after)
	var diffs []string
	for id, b := range was {
		a, ok := is[id]
		if !ok {
			diffs = append(diffs, fmt.Sprintf("%s: lost in replay", id))
			continue
		}
		if !reflect.DeepEqual(a, b) {
			diffs = append(diffs, fmt.Sprintf("%s: before %s, after %s", id, short(b), short(a)))
		}
	}
	for id := range is {
		if _, ok := was[id]; !ok {
			diffs = append(diffs, fmt.Sprintf("%s: appeared in replay", id))
		}
	}
	sort.Strings(diffs)
	return diffs
}

func short(o core.Object) string {
	b, _ := json.Marshal(o)
	if len(b) > 200 {
		return string(b[:200]) + "…"
	}
	return string(b)
}
