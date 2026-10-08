// Package network is Rhea learning across installations (DIRECTION, "The
// network: Rhea learns explanations"; first slice 2026-10-08). What flows
// upward is explanations, never facts: each installation publishes the
// shapes of its approved rules — what a rule explains and how, with instance
// values stripped — and the network counts which explanation how many
// installations chose for the same question. Those counts come back as the
// agent's priors and as the real support figures on suggestions ("3 of 4
// installations explain it this way"). A shape held by fewer installations
// than the floor never surfaces: one business's pattern may encode its
// secrets. Publications are append-only snapshots; an installation's latest
// one is its current knowledge.
package network

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"rhea/internal/core"
)

//go:embed schema.sql
var schemaSQL string

//go:embed latest.sql
var latestSQL string

// Floor is the fewest installations that must share a shape before it is
// knowledge: below it, a shape is one client's own business.
const Floor = 2

// SyntheticPrefix names a simulated customer's installation. What it
// publishes is stored marked and never counted for a real client: a hundred
// simulated businesses agreeing would be an invented statistic.
const SyntheticPrefix = "sim:"

// Synthetic reports whether an installation name is a simulated customer's.
func Synthetic(installation string) bool { return strings.HasPrefix(installation, SyntheticPrefix) }

// Network is the shared store of published shapes.
type Network struct {
	Pool *pgxpool.Pool
}

// DSN is the network's connection string, overridable via RHEA_NETWORK_DSN.
func DSN() string {
	if dsn := os.Getenv("RHEA_NETWORK_DSN"); dsn != "" {
		return dsn
	}
	return "postgres://rhea:rhea@127.0.0.1:5432/rhea_network"
}

// Open connects and makes sure the schema exists.
func Open(ctx context.Context, dsn string) (*Network, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		pool.Close()
		return nil, err
	}
	return &Network{Pool: pool}, nil
}

func (n *Network) Close() { n.Pool.Close() }

// Shape is one approved rule as the network sees it: the question it answers
// (Key), the anonymized answer (Spec), and the answer's identity.
type Shape struct {
	Key         string        `json:"key"`
	Fingerprint string        `json:"fingerprint"`
	Spec        core.RuleSpec `json:"spec"`
}

// structural are the condition paths whose values are vocabulary, not
// business data: which type cascaded, which market, which currency. Every
// other condition value is an instance value and is stripped.
var structural = map[string]bool{"$.object_type": true, "$.market": true, "$.currency": true}

// ShapeOf anonymizes a rule. Kept: the matched event, structural condition
// values, the effect with its templates, literals (enum values, account
// codes, book names — the market's shared vocabulary). Dropped: id,
// description, priority, dates, and every other condition value. Stripping
// instance literals well is ultimately an AI task (DIRECTION); this is the
// deterministic first cut, guarded by the floor.
func ShapeOf(spec core.RuleSpec) (Shape, error) {
	anon := spec
	anon.Match.Where = nil
	for _, c := range spec.Match.Where {
		if !structural[c.Path] {
			c.Value = "?"
		}
		anon.Match.Where = append(anon.Match.Where, c)
	}
	sort.Slice(anon.Match.Where, func(i, j int) bool { return anon.Match.Where[i].Path < anon.Match.Where[j].Path })
	b, err := json.Marshal(anon) // map keys marshal sorted: the encoding is canonical
	if err != nil {
		return Shape{}, err
	}
	return Shape{Key: KeyOf(spec), Fingerprint: fingerprint(b), Spec: anon}, nil
}

// KeyOf names the question a rule answers: the event it explains (a cascade
// names the type it cascades from) and what kind of consequence, into what.
// Two rules with one key are two answers to the same question.
func KeyOf(spec core.RuleSpec) string {
	q := spec.Match.EventType
	for _, c := range spec.Match.Where {
		if c.Path == "$.object_type" {
			q += ":" + fmt.Sprint(c.Value)
		}
	}
	switch e := spec.Effect; {
	case e.Postings != nil:
		book := e.Postings.Book
		if book == "" {
			book = "main"
		}
		return q + " → postings:" + book
	case e.Amend != nil:
		return q + " → amend:" + e.Amend.Type
	default:
		return q + " → object:" + e.Object.Type
	}
}

// Publish records an installation's current knowledge: the shapes of its
// active rules. Business data never leaves; only explanations do.
func (n *Network) Publish(ctx context.Context, installation string, rules []core.Rule) (int, error) {
	seen := map[string]bool{}
	var shapes []Shape
	for _, r := range rules {
		sh, err := ShapeOf(r.Spec)
		if err != nil {
			return 0, err
		}
		if seen[sh.Key+sh.Fingerprint] {
			continue
		}
		seen[sh.Key+sh.Fingerprint] = true
		shapes = append(shapes, sh)
	}
	b, err := json.Marshal(shapes)
	if err != nil {
		return 0, err
	}
	_, err = n.Pool.Exec(ctx, `INSERT INTO publication (installation, shapes, synthetic) VALUES ($1, $2, $3)`,
		installation, b, Synthetic(installation))
	return len(shapes), err
}

// Answer is one way installations explain a question, and how many do.
type Answer struct {
	Shape   Shape `json:"shape"`
	Count   int   `json:"count"` // installations explaining it this way
	Of      int   `json:"of"`    // installations explaining the question at all
	Surface bool  `json:"surface"`
}

// Knowledge is what the network has learned: per question, its answers by
// support, strongest first.
type Knowledge struct {
	Installations int                 `json:"installations"`
	Questions     map[string][]Answer `json:"questions"`
}

// Learn counts the latest publication of every real installation — what a
// client's Rhea draws on. Simulated customers are left out.
func (n *Network) Learn(ctx context.Context) (Knowledge, error) { return n.learn(ctx, false) }

// LearnIncludingSynthetic also counts simulated customers: the knowledge a
// simulated run draws on, so the flywheel can be exercised without a real
// client's figures ever moving.
func (n *Network) LearnIncludingSynthetic(ctx context.Context) (Knowledge, error) {
	return n.learn(ctx, true)
}

func (n *Network) learn(ctx context.Context, synthetic bool) (Knowledge, error) {
	rows, err := n.Pool.Query(ctx, latestSQL, synthetic)
	if err != nil {
		return Knowledge{}, err
	}
	defer rows.Close()
	k := Knowledge{Questions: map[string][]Answer{}}
	askers := map[string]int{}     // question → installations answering it
	votes := map[string]*Answer{}  // question|fingerprint → answer
	order := map[string][]string{} // question → fingerprints, first seen
	for rows.Next() {
		var installation string
		var raw []byte
		if err := rows.Scan(&installation, &raw); err != nil {
			return k, err
		}
		var shapes []Shape
		if err := json.Unmarshal(raw, &shapes); err != nil {
			return k, fmt.Errorf("publication of %s: %w", installation, err)
		}
		k.Installations++
		asked := map[string]bool{}
		for _, sh := range shapes {
			if !asked[sh.Key] {
				asked[sh.Key] = true
				askers[sh.Key]++
			}
			id := sh.Key + "|" + sh.Fingerprint
			if votes[id] == nil {
				votes[id] = &Answer{Shape: sh}
				order[sh.Key] = append(order[sh.Key], sh.Fingerprint)
			}
			votes[id].Count++
		}
	}
	if err := rows.Err(); err != nil {
		return k, err
	}
	for q, fps := range order {
		for _, fp := range fps {
			a := *votes[q+"|"+fp]
			a.Of = askers[q]
			a.Surface = a.Count >= Floor
			k.Questions[q] = append(k.Questions[q], a)
		}
		sort.SliceStable(k.Questions[q], func(i, j int) bool { return k.Questions[q][i].Count > k.Questions[q][j].Count })
	}
	return k, nil
}

// SupportFor is what the network knows about one rule: how many
// installations answer its question exactly this way. ok is false when the
// shape is below the floor — then Rhea knows nothing she may say.
func (k Knowledge) SupportFor(spec core.RuleSpec) (core.Support, bool) {
	sh, err := ShapeOf(spec)
	if err != nil {
		return core.Support{}, false
	}
	for _, a := range k.Questions[sh.Key] {
		if a.Shape.Fingerprint == sh.Fingerprint && a.Surface {
			return core.Support{Count: a.Count, Of: a.Of, Population: "installations"}, true
		}
	}
	return core.Support{}, false
}

// Priors are the surfaced answers to questions about the given event types —
// what the agent may lean on, with the real counts. They follow cascades: an
// answer that materializes an invoice brings the questions about
// object.materialized:invoice (how invoices are posted, what they raise).
func (k Knowledge) Priors(eventTypes []string) []Answer {
	reached := map[string]bool{}
	for _, t := range eventTypes {
		reached[t] = true
	}
	keys := make([]string, 0, len(k.Questions))
	for q := range k.Questions {
		keys = append(keys, q)
	}
	sort.Strings(keys)
	var out []Answer
	taken := map[string]bool{}
	for grew := true; grew; {
		grew = false
		for _, q := range keys {
			if taken[q] || !reached[strings.SplitN(q, " → ", 2)[0]] {
				continue
			}
			taken[q] = true
			for _, a := range k.Questions[q] {
				if !a.Surface {
					continue
				}
				out = append(out, a)
				if t := a.Shape.Spec.Effect.Object.Type; t != "" && !reached["object.materialized:"+t] {
					reached["object.materialized:"+t] = true
					grew = true
				}
			}
		}
	}
	return out
}

// fingerprint is the identity of a canonical encoding.
func fingerprint(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}
