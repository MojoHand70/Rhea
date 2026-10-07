// rhea: CLI for the Rhea kernel.
//
//	rhea init                create the schema in Postgres
//	rhea load FILE           load definitions (object types, view defs) — data, at runtime
//	rhea pack FILE           load a market pack: definitions + draft rules + master-data events
//	rhea ksef                run one pass of the KSeF statutory adapter (fake client)
//	rhea clock [-catchup] [-through DATE]
//	                         advance time: steady state opens one day by itself;
//	                         a gap waits for -catchup (the burst gate)
//	rhea submit FILE         submit a raw event (dedup by file content hash)
//	rhea serve [-addr :8070] run the shell
//	rhea replay              rebuild object cache + DuckDB from the event log
//	rhea eval [CORPUS]       measure the agent as an author: draft → simulate →
//	                         approve over a corpus on a throwaway database;
//	                         prints residue (default testdata/eval/corpus.json)
package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"

	"rhea/internal/adapter"
	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/eval"
	"rhea/internal/exec"
	"rhea/internal/pack"
	"rhea/internal/project"
	"rhea/internal/shell"
	"rhea/internal/store"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatal("usage: rhea init|load|submit|serve|replay")
	}
	ctx := context.Background()
	s, err := store.Open(ctx, store.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	x := &exec.Executor{Store: s}

	switch os.Args[1] {
	case "init":
		if err := s.Init(ctx); err != nil {
			log.Fatal(err)
		}
		fmt.Println("schema ready; builtin activities seeded (the gate's birth is in the log)")

	case "load":
		if len(os.Args) < 3 {
			log.Fatal("usage: rhea load FILE")
		}
		n, err := loadDefinitions(ctx, s, os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("loaded %d definitions\n", n)

	case "pack":
		if len(os.Args) < 3 {
			log.Fatal("usage: rhea pack FILE")
		}
		sum, err := pack.Load(ctx, s, os.Args[2], actor())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("pack %s: %d types, %d views, %d draft rules, %d draft activities, %d events loaded; %d already present\n",
			sum.Pack, sum.Types, sum.Views, sum.Rules, sum.Activities, sum.Events, sum.Skipped)
		if sum.Rules > 0 || sum.Activities > 0 {
			fmt.Println("rules and activities are drafts — approve them in the shell to bring the pack to life")
		}

	case "ksef":
		// One pass of the Poland pack's statutory adapter, against the fake
		// KSeF client (the real API stays outside the experiment).
		run, err := adapter.Run(ctx, s, x, adapter.KSeF{Client: adapter.FakeKSeF{}})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ksef: %d submitted, %d retries already logged, %d booked\n",
			run.Submitted, run.Duplicates, run.Booked)
		for _, e := range run.Errors {
			fmt.Println("worklist:", e)
		}
		if run.PassErr != nil {
			log.Fatal(run.PassErr)
		}

	case "clock":
		fs := flag.NewFlagSet("clock", flag.ExitOnError)
		catchup := fs.Bool("catchup", false, "open every pending day (the human act after a gap)")
		through := fs.String("through", "", "catch up only through this date (YYYY-MM-DD)")
		fs.Parse(os.Args[2:])
		c := adapter.Clock{}
		var res adapter.ClockResult
		if *catchup {
			res, err = adapter.CatchUpClock(ctx, s, x, c, *through)
		} else {
			res, err = adapter.RunClock(ctx, s, x, c)
		}
		if err != nil {
			log.Fatal(err)
		}
		if res.Gated {
			fmt.Printf("gated: %d days unopened (%s … %s) — a burst needs a human.\n",
				len(res.Pending), res.Pending[0], res.Pending[len(res.Pending)-1])
			fmt.Println("dry-run the catch-up in the shell (Time tab), or run: rhea clock -catchup")
			return
		}
		fmt.Printf("clock: opened %d day(s), fired %d occurrence(s), booked %d\n",
			len(res.Opened), res.Fired, res.Booked)
		for _, e := range res.Errors {
			fmt.Println("worklist:", e)
		}

	case "submit":
		if len(os.Args) < 3 {
			log.Fatal("usage: rhea submit FILE")
		}
		id, booked, err := submit(ctx, s, x, os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("event %d accepted; booked %d\n", id, booked)

	case "serve":
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		addr := fs.String("addr", ":8070", "listen address")
		fs.Parse(os.Args[2:])
		srv := &shell.Server{Store: s, Exec: x, Agent: agent.New(), DuckPath: project.Path()}
		log.Printf("rhea shell on http://localhost%s (model %s)", *addr, agent.Model())
		log.Fatal(http.ListenAndServe(*addr, srv.Handler()))

	case "replay":
		objs, err := x.Replay(ctx)
		if err != nil {
			log.Fatal(err)
		}
		n, err := project.Rebuild(ctx, s, project.Path())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("replayed %d objects (postgres cache), %d rows (duckdb)\n", len(objs), n)

	case "eval":
		path := "testdata/eval/corpus.json"
		if len(os.Args) > 2 {
			path = os.Args[2]
		}
		c, dir, err := eval.Load(path)
		if err != nil {
			log.Fatal(err)
		}
		// Never the system of record: the run gets its own log, dropped after.
		tmp, cleanup, err := store.Throwaway(ctx, "rhea_eval")
		if err != nil {
			log.Fatal(err)
		}
		defer cleanup()
		rep, err := eval.Run(ctx, tmp, agent.New(), agent.Model(), c, dir)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(rep)

	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}

// Definitions file: {"object_types": [...], "view_defs": [...]} — the
// "extend at runtime" path for everything that is not a rule.
func loadDefinitions(ctx context.Context, s *store.Store, path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var defs struct {
		ObjectTypes []core.ObjectType `json:"object_types"`
		ViewDefs    []core.ViewDef    `json:"view_defs"`
	}
	if err := json.Unmarshal(b, &defs); err != nil {
		return 0, err
	}
	n := 0
	for _, t := range defs.ObjectTypes {
		if err := s.InsertObjectType(ctx, t); err != nil {
			return n, fmt.Errorf("object type %s: %w", t.Name, err)
		}
		n++
	}
	for _, v := range defs.ViewDefs {
		if err := s.InsertViewDef(ctx, v); err != nil {
			return n, fmt.Errorf("view def %s: %w", v.ID, err)
		}
		n++
	}
	return n, nil
}

// actor identifies who runs this CLI: RHEA_ACTOR, or cli:<os user>.
func actor() string {
	if a := os.Getenv("RHEA_ACTOR"); a != "" {
		return a
	}
	if u, err := user.Current(); err == nil {
		return "cli:" + u.Username
	}
	return "cli:unknown"
}

// Event file: {"event_type": ..., "occurred_at": ..., "payload": {...}}.
// Dedup key is the content hash: the same file can never book twice. The
// submission goes through the submit_event door — the open door declared as
// data — so the event carries its activity stamp like any other.
func submit(ctx context.Context, s *store.Store, x *exec.Executor, path string) (int64, int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	var ev struct {
		EventType  string         `json:"event_type"`
		OccurredAt string         `json:"occurred_at"`
		Payload    map[string]any `json:"payload"`
	}
	if err := json.Unmarshal(b, &ev); err != nil {
		return 0, 0, err
	}
	t, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
		"event_type": ev.EventType, "occurred_at": ev.OccurredAt, "payload": ev.Payload,
	}, actor(), fmt.Sprintf("file/%x", sha256.Sum256(b))[:21])
	if err != nil {
		return 0, 0, err
	}
	if len(t.Errors) > 0 {
		return t.EventID, t.Booked, t.Errors[0]
	}
	return t.EventID, t.Booked, nil
}
