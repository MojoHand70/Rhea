# Rhea — CLAUDE.md

An AI-driven ERP: a "language of ERP" (events, objects, documents, activities, rules,
views), AI authoring rules with a human in the loop, a deterministic kernel executing
them. Read `SPEC.md` before writing any code. The spec is the design; do not invent
another one. `DIRECTION.md` holds the briefs, `DECISIONS.md` the decisions.

## What this is
- THE product (KK, 2026-10-20: the experiment gathered its proof — the founding ladder,
  the statutory floor, two live reviews — and the state changed). It will hold many
  companies' books, an accounting office's clients among them, and a company's books
  are that company's: a scope leak, a double booking or a wrong amount is a defect of
  the first order, never an accepted cost.
- Still: no gold-plating, no performance work ahead of a measured need, no UI beyond
  the shell pattern (the face is built apart). Iteration stays fast because the checks
  are automatic, not because they are skipped.

## Stack (fixed)
- Go 1.26+, single module `rhea`, standard library first.
- PostgreSQL 17 (system of record) in Docker container `rhea-postgres`, localhost:5432,
  db/user/password `rhea` (local-only experiment; no secrets pretence).
  Driver `github.com/jackc/pgx/v5`.
- DuckDB (read/analysis side) via `github.com/duckdb/duckdb-go/v2`, file `data/rhea.duckdb`.
- Anthropic Go SDK; model from env `RHEA_MODEL` (default `claude-sonnet-4-6`), key from
  `ANTHROPIC_API_KEY`. Never hard-code either.
- Shell: `net/http` + embedded vanilla HTML/JS/CSS from `web/`. No framework, no build step.
- SQL lives in `.sql` files embedded with `//go:embed`, never in Go string literals.

## Invariants (hard — tests must enforce them)
1. The `event` log is append-only. No UPDATE, no DELETE, anywhere. Same for `rule`,
   `object_type`, `view_def`: changes are new version rows.
2. Rules follow `draft → active → superseded`; only `active` rules execute; activation
   requires an explicit human approval activity, itself recorded as an event — the
   approval is that event, not a resting status.
3. The agent never writes state. Its output is data: definitions stored as drafts and
   activated only by a human approval, and read-only queries the kernel runs on the
   analysis copy in a locked session that cannot write, reach files or reach the network.
   Model output never runs with authority. (KK, 2026-10-17)
4. Determinism: wiping projections and replaying the event log through the same rule
   versions reproduces identical object state. A test does exactly that.
5. Every object and field is explained: provenance `(event_id, rule_id, rule_version)`
   on every derived fact. Nothing appears in state without a rule.
6. Money is `int64` minor units in Go, `DECIMAL(18,2)` in SQL, strings across the
   boundary. No floats, ever.
7. The `object` table is a rebuildable cache owned by the kernel's executor/replayer —
   the only non-append table, and nothing else may write it.
8. The kernel never imports a market. A market's code (a statutory channel, a document
   syntax) lives under `markets/<cc>/` behind the adapter contract and registers itself;
   the kernel knows no market by name. A fix to the Polish pack is a fix to the Polish
   pack, never to the kernel. `markets/markets_test.go` enforces the import direction.
9. Isolation: a company's facts and objects are seen only within a scope that names
   that company. A scope is a set of companies (an accounting office sees its clients);
   installation-wide rows (the pack's master data, with no company) are seen by every
   scope. The database enforces it (row-level security on a role every scoped request
   runs as), so a query that forgets its scope returns nothing, never another company's
   rows; the analysis copy carries the company and filters on it. A test walks every
   read path with two companies.
10. Concurrency and idempotency: the executor runs one chain at a time (one lock, in
    process and across processes); a chain is bounded in depth and size; every door and
    every reader is safe to run twice (dedup keys name what a thing is, never only its
    bytes). Tests run doors from several goroutines and feed every reader the same
    document twice.

## Layout
```
cmd/rhea/          CLI entry (rhea init | serve | submit | replay)
internal/core/     the language: Event, ObjectType, Object, Rule, ViewDef types + JSON schemas
internal/store/    Postgres event log + versioned definition stores (.sql embedded)
internal/exec/     deterministic executor: match → expand → emit derived events → project
internal/agent/    Anthropic calls; drafts rule JSON; no side effects, no store access
internal/adapter/  the adapter contract, runner and registry; EU-wide readers (MT940, PDF)
markets/<cc>/      a market's code behind the contract: pl = KSeF in/out, FA(3)
internal/pack/     market-pack loader: definitions + draft rules + master-data events
internal/project/  event log → DuckDB projections for analysis views
internal/shell/    HTTP JSON API + embedded web shell
web/               activity bar / submenu / tabs shell (vanilla)
packs/             market packs as data (pl: CoA, VAT, KSeF)
testdata/          sample raw events, expected objects, expected analysis output
SPEC.md            the design
```

## Working style
- One finding, one failing test, one fix, one commit. The failure scenario becomes a
  test before the fix; a test is never edited to fit a fix without saying so in the
  commit. Batches of fixes are where the regressions came from (DECISIONS 2026-10-20).
- Before every change, the boundary checklist: which company does this read; what if
  two run at once; what if it runs twice; which values are money. Write the answers
  into the design note or the commit.
- `./scripts/check.sh` before claiming done: gofmt, vet, and `go test ./... -count=1
  -race` with `RHEA_CI=1` against the real database. The test cache hid a red suite
  for a day; a skip for a missing database counts as red. The pre-push hook runs it.
- Work is not done until an independent review has read the diff (KK triggers it; the
  author's session does not mark its own work). Anything marked critical gets a
  reproduction before it is called fixed.
- When the spec is silent, pick the simplest option and note it in `DECISIONS.md`
  (one line: date, decision, why). Do not ask; do not gold-plate.
- Do not defer or leave TODO stubs for anything inside the current milestone; an
  "assessed and left" item gets a line in DECISIONS with its reason and, when it
  touches invariant 9 or 10, a date.
- The shell renders view notions generically. If you catch yourself writing an
  invoice-specific screen, stop — that knowledge belongs in a ViewDef.

## Definition of done (a change)
The failing test exists and passes; `./scripts/check.sh` is green; the invariant tests
still enforce all ten; `rhea replay` reproduces identical state on the test story; the
change is explained in business terms in its commit; an independent review has run.
The founding milestone (SPEC §6/M0: raw invoice event → worklist → agent-drafted rule →
human approval → materialized document in list and detail → DuckDB analysis → replay
identical) stays the story every pack's tests tell.
