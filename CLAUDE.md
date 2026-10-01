# Rhea — CLAUDE.md

Experiment toward an AI-driven ERP: a "language of ERP" (events, objects, documents,
activities, rules, views), AI authoring rules with a human in the loop, a deterministic
kernel executing them. Read `SPEC.md` before writing any code. The spec is the design;
do not invent another one.

## What this is / is not
- IS: a proof that the three-part split (language / AI author / deterministic kernel)
  can run a business domain end to end, extensible at runtime, auditable by replay.
- IS NOT: a product. No auth, no multi-tenant, no performance work, no UI polish beyond
  the shell pattern. Throwaway status is explicit; optimise for iteration speed.

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
2. Rules follow `draft → approved → active → superseded`; only `active` rules execute;
   activation requires an explicit human approval activity, itself recorded as an event.
3. The agent never writes state. It returns strict JSON; the kernel validates against
   schema and stores drafts. Model output is data, never executed.
4. Determinism: wiping projections and replaying the event log through the same rule
   versions reproduces identical object state. A test does exactly that.
5. Every object and field is explained: provenance `(event_id, rule_id, rule_version)`
   on every derived fact. Nothing appears in state without a rule.
6. Money is `int64` minor units in Go, `DECIMAL(18,2)` in SQL, strings across the
   boundary. No floats, ever.
7. The `object` table is a rebuildable cache owned by the kernel's executor/replayer —
   the only non-append table, and nothing else may write it.

## Layout
```
cmd/rhea/          CLI entry (rhea init | serve | submit | replay)
internal/core/     the language: Event, ObjectType, Object, Rule, ViewDef types + JSON schemas
internal/store/    Postgres event log + versioned definition stores (.sql embedded)
internal/exec/     deterministic executor: match → expand → emit derived events → project
internal/agent/    Anthropic calls; drafts rule JSON; no side effects, no store access
internal/project/  event log → DuckDB projections for analysis views
internal/shell/    HTTP JSON API + embedded web shell
web/               activity bar / submenu / tabs shell (vanilla)
testdata/          sample raw events, expected objects, expected analysis output
SPEC.md            the design
```

## Working style
- Small commits, one step per PR-sized chunk. Run `go test ./...` before claiming done.
- When the spec is silent, pick the simplest option and note it in `DECISIONS.md`
  (one line: date, decision, why). Do not ask; do not gold-plate.
- Do not defer or leave TODO stubs for anything inside the current milestone.
- The shell renders view notions generically. If you catch yourself writing an
  invoice-specific screen, stop — that knowledge belongs in a ViewDef.

## Definition of done (milestone 0)
The demo story in SPEC §6/M0 runs end to end: submitted raw invoice event → worklist →
agent-drafted rule → human approval in the shell → materialized Invoice document visible
in list and detail views → DuckDB analysis view aggregates it → `rhea replay` reproduces
identical state → `go test ./...` green including invariant tests.
