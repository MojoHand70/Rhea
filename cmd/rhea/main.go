// rhea: CLI for the Rhea kernel.
//
//	rhea init                create the schema in Postgres
//	rhea load FILE           draft definitions (object types, view defs) into one bundle
//	rhea pack FILE           load a market pack as one draft bundle + master-data events
//	rhea approve BUNDLE      activate a bundle: every member, or none
//	rhea backfill [-dry]     let the active rules explain the past further:
//	                         additive only; closed periods are offered forward
//	rhea publish             share this installation's explanations with the
//	                         network: rule shapes, never data
//	rhea network             what Rhea has learned across installations
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
//	                         prints residue (default testdata/eval/corpus.json).
//	                         A persona file (testdata/customers/) runs as a
//	                         simulated customer: -voice N picks a paraphrase,
//	                         -months N shortens the year
//	rhea sim PERSONA         the simulated customer: what it generates; -corpus FILE
//	                         writes the compiled corpus, -paraphrase N asks the model
//	                         to rephrase the interview, -play submits its documents
//	                         into THIS installation for a live demo
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
	"sort"
	"strings"
	"time"

	"rhea/internal/adapter"
	"rhea/internal/agent"
	"rhea/internal/core"
	"rhea/internal/eval"
	"rhea/internal/exec"
	"rhea/internal/network"
	"rhea/internal/pack"
	"rhea/internal/project"
	"rhea/internal/shell"
	"rhea/internal/sim"
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

	case "publish":
		nw, err := network.Open(ctx, network.DSN())
		if err != nil {
			log.Fatal(err)
		}
		defer nw.Close()
		name, err := s.Installation(ctx)
		if err != nil {
			log.Fatal(err)
		}
		rules, err := s.ActiveRules(ctx)
		if err != nil {
			log.Fatal(err)
		}
		n, err := nw.Publish(ctx, name, rules)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s published %d explanation shape(s) — no business data left this installation\n", name, n)

	case "network":
		fs := flag.NewFlagSet("network", flag.ExitOnError)
		synthetic := fs.Bool("synthetic", false, "count simulated customers too (never what a client sees)")
		fs.Parse(os.Args[2:])
		nw, err := network.Open(ctx, network.DSN())
		if err != nil {
			log.Fatal(err)
		}
		defer nw.Close()
		learn := nw.Learn
		if *synthetic {
			learn = nw.LearnIncludingSynthetic
		}
		k, err := learn(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d installation(s) have published; an answer is knowledge once %d share it\n", k.Installations, network.Floor)
		qs := make([]string, 0, len(k.Questions))
		for q := range k.Questions {
			qs = append(qs, q)
		}
		sort.Strings(qs)
		for _, q := range qs {
			fmt.Println(q)
			for _, a := range k.Questions[q] {
				mark := "  "
				if !a.Surface {
					mark = "· " // below the floor: that client's own business
				}
				fmt.Printf("  %s%d of %d  %s\n", mark, a.Count, a.Of, a.Shape.Fingerprint)
			}
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
		// Learning is optional: without a network Rhea runs alone.
		if nw, err := network.Open(ctx, network.DSN()); err == nil {
			defer nw.Close()
			srv.Network = nw
			log.Printf("network: %s", network.DSN())
		} else {
			log.Printf("network unavailable, running alone: %v", err)
		}
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
		learn := fs.Bool("network", false, "draw on what Rhea learned, then publish this run's explanations")
		voice := fs.Int("voice", 0, "persona only: run the k-th paraphrase of the interview (0: as scripted)")
		months := fs.Int("months", 0, "persona only: override the persona's months")
		fs.Parse(os.Args[2:])
		path := "testdata/eval/corpus.json"
		if fs.NArg() > 0 {
			path = fs.Arg(0)
		}
		var c eval.Corpus
		var dir string
		if sim.IsPersona(path) {
			// A simulated customer: facts and expected state from the
			// generator, the interview as scripted or as paraphrased.
			p, pdir, err := sim.Load(path)
			if err != nil {
				log.Fatal(err)
			}
			if *months > 0 {
				p.Months = *months
			}
			voices, _, err := sim.LoadVoices(path)
			if err != nil {
				log.Fatal(err)
			}
			var facts sim.Facts
			if c, facts, err = sim.Compile(p, &voices, *voice); err != nil {
				log.Fatal(err)
			}
			dir = pdir
			fmt.Print(sim.Summary(p, facts))
		} else {
			var err error
			if c, dir, err = eval.Load(path); err != nil {
				log.Fatal(err)
			}
		}
		// Never the system of record: the run gets its own log, dropped after.
		tmp, cleanup, err := store.Throwaway(ctx, "rhea_eval")
		if err != nil {
			log.Fatal(err)
		}
		defer cleanup()
		var nw *network.Network
		var k *network.Knowledge
		if *learn {
			if nw, err = network.Open(ctx, network.DSN()); err != nil {
				log.Fatal(err)
			}
			defer nw.Close()
			// A simulated run draws on simulated runs too; a real corpus
			// never sees them.
			learnFrom := nw.Learn
			if network.Synthetic(c.Name) {
				learnFrom = nw.LearnIncludingSynthetic
			}
			known, err := learnFrom(ctx)
			if err != nil {
				log.Fatal(err)
			}
			k = &known
			fmt.Printf("drawing on %d installation(s)\n", known.Installations)
		}
		rep, err := eval.Run(ctx, tmp, agent.New(), agent.Model(), c, dir, k)
		if err != nil {
			log.Fatal(err)
		}
		if nw != nil {
			rules, err := tmp.ActiveRules(ctx)
			if err != nil {
				log.Fatal(err)
			}
			name := fmt.Sprintf("eval-%s-%s", c.Name, time.Now().Format("20060102-150405"))
			if network.Synthetic(c.Name) { // stays marked: never a real client's figure
				name = fmt.Sprintf("%s-%s", c.Name, time.Now().Format("20060102-150405"))
			}
			if _, err := nw.Publish(ctx, name, rules); err != nil {
				log.Fatal(err)
			}
			defer fmt.Printf("published as %s\n", name)
		}
		if *verbose {
			fmt.Print(rep.Verbose())
		} else {
			fmt.Print(rep)
		}

	case "sim":
		fs := flag.NewFlagSet("sim", flag.ExitOnError)
		months := fs.Int("months", 0, "override the persona's months")
		paraphrase := fs.Int("paraphrase", 0, "ask the model for N rephrasings of every interview answer and save them beside the persona")
		voice := fs.Int("voice", 0, "compile with the k-th paraphrase (0: the scripted answers)")
		out := fs.String("corpus", "", "write the compiled corpus to this file")
		play := fs.Bool("play", false, "submit the persona's documents into THIS installation through the door, for a live demo")
		fs.Parse(os.Args[2:])
		if fs.NArg() != 1 {
			log.Fatal("usage: rhea sim [-months N] [-paraphrase N] [-voice N] [-corpus FILE] [-play] PERSONA")
		}
		path := fs.Arg(0)
		p, _, err := sim.Load(path)
		if err != nil {
			log.Fatal(err)
		}
		if *months > 0 {
			p.Months = *months
		}
		if *paraphrase > 0 {
			a := agent.New()
			v, err := sim.Paraphrase(ctx, a.Complete, agent.Model(), p, *paraphrase)
			if err != nil {
				log.Fatal(err)
			}
			if err := sim.SaveVoices(path, v); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("saved %d voices per answer to %s\n", *paraphrase, sim.VoicesPath(path))
		}
		voices, _, err := sim.LoadVoices(path)
		if err != nil {
			log.Fatal(err)
		}
		c, facts, err := sim.Compile(p, &voices, *voice)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(sim.Summary(p, facts))
		if *out != "" {
			b, err := json.MarshalIndent(c, "", "  ")
			if err != nil {
				log.Fatal(err)
			}
			if err := os.WriteFile(*out, append(b, '\n'), 0o644); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("corpus written to %s\n", *out)
		}
		if *play {
			// The demo's business: the same facts, through the real door
			// into the live log, dedup-keyed so playing twice adds nothing.
			// The interview stays the human's: it is printed as the script.
			played, waiting := 0, 0
			for i, ev := range facts.Events {
				tr, err := x.TriggerActivity(ctx, "submit_event", map[string]any{
					"event_type": ev.EventType, "occurred_at": ev.OccurredAt, "payload": ev.Payload,
				}, "sim:"+p.Name, fmt.Sprintf("sim:%s:%d:%d", p.Name, p.Seed, i))
				if err != nil {
					if store.IsDuplicate(err) {
						continue
					}
					log.Fatal(err)
				}
				played++
				waiting += len(tr.Errors)
			}
			fmt.Printf("played %d document(s) into this installation; the interview script:\n", played)
			for i, a := range c.Tasks {
				fmt.Printf("%2d. %s\n", i+1, a.Intent)
			}
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
