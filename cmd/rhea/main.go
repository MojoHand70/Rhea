// rhea: CLI for the Rhea kernel.
//
//	rhea init                create the schema in Postgres
//	rhea load FILE           draft definitions (object types, view defs) into one bundle
//	rhea pack FILE           load a market pack as one draft bundle + master-data events
//	rhea approve BUNDLE      activate a bundle: every member, or none
//	rhea backfill [-dry]     let the active rules explain the past further:
//	                         additive only; closed periods are offered forward
//	rhea reject BUNDLE REASON
//	                         reject a bundle; the reason stays on the record
//	rhea ksef                run one pass of the KSeF statutory adapter (fake client)
//	rhea clock [-catchup] [-through DATE]
//	                         advance time: steady state opens one day by itself;
//	                         a gap waits for -catchup (the burst gate)
//	rhea submit FILE         submit a raw event (dedup by file content hash)
//	rhea serve [-addr :8070] run the shell
//	rhea replay              rebuild object cache + DuckDB from the event log
//	rhea eval [-v] [CORPUS]  measure the agent as an author: draft → simulate →
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
	"path/filepath"
	"strings"
	"time"

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
		n, bundle, err := loadDefinitions(ctx, s, os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("drafted %d definitions\n", n)
		if bundle != "" {
			fmt.Printf("bundle %s waits for approval — rhea approve %s, or the shell's Bundles tab\n", bundle, bundle)
		}

	case "pack":
		if len(os.Args) < 3 {
			log.Fatal("usage: rhea pack FILE")
		}
		sum, err := pack.Load(ctx, s, os.Args[2], actor())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("pack %s: %d types, %d views, %d rules, %d activities drafted, %d events loaded; %d already present\n",
			sum.Pack, sum.Types, sum.Views, sum.Rules, sum.Activities, sum.Events, sum.Skipped)
		if sum.Bundle != "" {
			fmt.Printf("bundle %s waits for approval — approving it installs the pack (rhea approve %s)\n", sum.Bundle, sum.Bundle)
		}

	case "approve":
		if len(os.Args) < 3 {
			log.Fatal("usage: rhea approve BUNDLE")
		}
		b, booked, procErrs, err := x.ApproveBundle(ctx, os.Args[2], actor(), today())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("bundle %s active: %d members; booked %d pending event(s)\n", b.ID, len(b.Members), booked)
		for _, e := range procErrs {
			fmt.Println("worklist:", e)
		}

	case "backfill":
		fs := flag.NewFlagSet("backfill", flag.ExitOnError)
		dry := fs.Bool("dry", false, "show what would happen, change nothing")
		fs.Parse(os.Args[2:])
		plan, err := x.PlanBackfill(ctx, today())
		if err != nil {
			log.Fatal(err)
		}
		for _, c := range plan.Chains {
			fmt.Printf("event %d %s (%s): %s", c.EventID, c.EventType, c.OccurredAt, strings.Join(c.Rules, ", "))
			if c.Forwarded != "" {
				fmt.Printf(" — offered forward to %s (%s)", c.BookedAt, c.Forwarded)
			}
			fmt.Println()
		}
		for _, e := range plan.Errors {
			fmt.Println("refused:", e)
		}
		fmt.Printf("%d past event(s) to explain further: %d object(s) added, %d changed\n",
			len(plan.Chains), len(plan.Diff.Added), len(plan.Diff.Changed))
		if *dry || len(plan.Chains) == 0 {
			return
		}
		n, errs, err := x.ApproveBackfill(ctx, actor(), today())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("backfilled %d past event(s)\n", n)
		for _, e := range errs {
			fmt.Println("worklist:", e)
		}

	case "reject":
		if len(os.Args) < 4 {
			log.Fatal("usage: rhea reject BUNDLE REASON")
		}
		b, err := x.RejectBundle(ctx, os.Args[2], os.Args[3], actor(), today())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("bundle %s rejected; its drafts stay on the record, a redraft is a new bundle\n", b.ID)

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
		fs := flag.NewFlagSet("eval", flag.ExitOnError)
		verbose := fs.Bool("v", false, "print every draft the agent proposed")
		fs.Parse(os.Args[2:])
		path := "testdata/eval/corpus.json"
		if fs.NArg() > 0 {
			path = fs.Arg(0)
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
		if *verbose {
			fmt.Print(rep.Verbose())
		} else {
			fmt.Print(rep)
		}

	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}

// Definitions file: {"object_types": [...], "view_defs": [...]} — the
// "extend at runtime" path for everything that is not a rule. Definitions
// land as drafts in one bundle named after the file and its content (KK,
// 2026-10-08: everything drafts); versions already present are skipped.
func loadDefinitions(ctx context.Context, s *store.Store, path string) (int, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, "", err
	}
	var defs struct {
		ObjectTypes []core.ObjectType `json:"object_types"`
		ViewDefs    []core.ViewDef    `json:"view_defs"`
	}
	if err := json.Unmarshal(b, &defs); err != nil {
		return 0, "", err
	}
	var members []core.Member
	for _, t := range defs.ObjectTypes {
		if _, _, err := s.GetObjectTypeVersion(ctx, t.Name, t.Version); err == nil {
			continue
		}
		if err := store.InsertObjectTypeRow(ctx, s.Pool, t, core.StatusDraft); err != nil {
			return len(members), "", fmt.Errorf("object type %s: %w", t.Name, err)
		}
		members = append(members, core.Member{Kind: core.KindObjectType, Name: t.Name, Version: t.Version})
	}
	for _, v := range defs.ViewDefs {
		if _, _, err := s.GetViewDefVersion(ctx, v.ID, v.Version); err == nil {
			continue
		}
		if err := store.InsertViewDefRow(ctx, s.Pool, v, core.StatusDraft); err != nil {
			return len(members), "", fmt.Errorf("view def %s: %w", v.ID, err)
		}
		members = append(members, core.Member{Kind: core.KindViewDef, Name: v.ID, Version: v.Version})
	}
	if len(members) == 0 {
		return 0, "", nil
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	bundle, err := s.InsertBundle(ctx, core.Bundle{
		ID:          fmt.Sprintf("load-%s-%x", base, sha256.Sum256(b))[:len("load-"+base)+9],
		Description: "Definitions from " + filepath.Base(path),
		Members:     members, CreatedBy: actor(),
	})
	if err != nil {
		return len(members), "", err
	}
	return len(members), bundle.ID, nil
}

func today() string { return time.Now().Format("2006-01-02") }

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
