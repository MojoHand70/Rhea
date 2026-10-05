package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"rhea/internal/core"
	"rhea/internal/exec"
	"rhea/internal/store"
)

// The clock (SPEC M5): time passing becomes facts in the log. The clock is
// an adapter — it reads projected objects and acts only by appending raw
// events under its own actor — but it carries its own runner instead of
// Pass: time must interleave with processing, because a day's consequences
// (a schedule advanced by a rule) decide the next day's due set.
//
// Division of labor, per the adapter contract: calendars are the clock's
// knowledge. The clock owns the wall-clock read, the cadence arithmetic
// ("+1 month", clamped to month ends) and the due-set scan; the kernel owns
// none of it. A schedule fires as a schedule.fired raw event with the
// advanced next_run baked in the payload, and ordinary rules — pure data —
// raise the work and amend the schedule. Dedup keys make every day, month
// and occurrence idempotent, so a crashed or repeated run is a no-op.
//
// The switch (DIRECTION 2026-10-05): steady state is automatic, bursts need
// a human. Run opens at most one pending day; a gap is reported, simulated
// on request (SimulateCatchUp), and opened only by an explicit CatchUp.

const (
	// EventScheduleFired announces one due occurrence of one schedule. Not a
	// time.* event: an unexplained firing is real residue — a schedule whose
	// work nobody declared — and belongs in the worklist.
	EventScheduleFired = "schedule.fired"

	// ScheduleType is the conventional vocabulary the clock scans — a type
	// the packs ship, like posting or account; the kernel does not know it.
	ScheduleType = "task_schedule"
)

// Clock is the system's only wall-clock reader. Now is injectable so tests
// own time completely.
type Clock struct {
	Now func() time.Time
}

func (c Clock) today() string {
	now := c.Now
	if now == nil {
		now = time.Now
	}
	return now().UTC().Format("2006-01-02")
}

// ClockStatus says where time stands: the last opened day and every day
// still unopened through today, oldest first.
type ClockStatus struct {
	Today      string   `json:"today"`
	LastOpened string   `json:"last_opened,omitempty"`
	Pending    []string `json:"pending"`
}

// ClockResult says what one clock run did.
type ClockResult struct {
	Opened  []string `json:"opened"`
	Fired   int      `json:"fired"` // schedule occurrences emitted
	Booked  int      `json:"booked"`
	Gated   bool     `json:"gated"` // a burst awaited a human instead of opening
	Pending []string `json:"pending,omitempty"`
	Errors  []error  `json:"-"`
}

// ClockState reads where time stands. A log that never opened a day starts
// today — the genesis day is the day the clock first runs.
func ClockState(ctx context.Context, s *store.Store, c Clock) (ClockStatus, error) {
	st := ClockStatus{Today: c.today()}
	last, ok, err := s.LatestEventOfType(ctx, core.EventDayOpened)
	if err != nil {
		return st, err
	}
	from := st.Today
	if ok {
		var p struct {
			Date string `json:"date"`
		}
		if err := json.Unmarshal(last.Payload, &p); err != nil {
			return st, err
		}
		st.LastOpened = p.Date
		next, err := nextDay(p.Date)
		if err != nil {
			return st, err
		}
		from = next
	}
	for d := from; d <= st.Today; {
		st.Pending = append(st.Pending, d)
		n, err := nextDay(d)
		if err != nil {
			return st, err
		}
		d = n
	}
	return st, nil
}

// RunClock is the steady state: at most one pending day opens by itself;
// more than one is a burst and waits for a human (the gate).
func RunClock(ctx context.Context, s *store.Store, x *exec.Executor, c Clock) (ClockResult, error) {
	st, err := ClockState(ctx, s, c)
	if err != nil {
		return ClockResult{}, err
	}
	if len(st.Pending) > 1 {
		return ClockResult{Gated: true, Pending: st.Pending}, nil
	}
	return openDays(ctx, s, x, st.LastOpened, st.Pending)
}

// CatchUpClock is the human act: open every pending day (through the given
// date, or all of them), in order, each day's consequences processed before
// the next day's due set is read.
func CatchUpClock(ctx context.Context, s *store.Store, x *exec.Executor, c Clock, through string) (ClockResult, error) {
	st, err := ClockState(ctx, s, c)
	if err != nil {
		return ClockResult{}, err
	}
	days := st.Pending
	if through != "" {
		cut := 0
		for _, d := range days {
			if d > through {
				break
			}
			cut++
		}
		days = days[:cut]
	}
	return openDays(ctx, s, x, st.LastOpened, days)
}

func openDays(ctx context.Context, s *store.Store, x *exec.Executor, lastOpened string, days []string) (ClockResult, error) {
	var res ClockResult
	appendNew := func(ev core.Event) (bool, error) {
		ev.Kind = core.KindRaw
		ev.Actor = "adapter:clock"
		switch _, err := s.AppendEvent(ctx, ev); {
		case err == nil:
			return true, nil
		case store.IsDuplicate(err):
			return false, nil
		default:
			return false, err
		}
	}
	prev := lastOpened
	for _, d := range days {
		day, _ := json.Marshal(map[string]any{"date": d})
		if _, err := appendNew(core.Event{Type: core.EventDayOpened, OccurredAt: d,
			Payload: day, DedupKey: "time/day/" + d}); err != nil {
			return res, err
		}
		if prev == "" || d[:7] != prev[:7] {
			month, _ := json.Marshal(map[string]any{"month": d[:7], "date": d})
			if _, err := appendNew(core.Event{Type: core.EventMonthOpened, OccurredAt: d,
				Payload: month, DedupKey: "time/month/" + d[:7]}); err != nil {
				return res, err
			}
		}
		prev = d

		// The due set, read fresh each day: the advance a rule booked
		// yesterday decides what fires today.
		scheds, err := s.ObjectsByType(ctx, ScheduleType)
		if err != nil {
			return res, err
		}
		for _, fired := range dueOccurrences(scheds, d) {
			payload, err := json.Marshal(fired)
			if err != nil {
				return res, err
			}
			added, err := appendNew(core.Event{Type: EventScheduleFired, OccurredAt: d,
				Payload: payload, DedupKey: "time/fired/" + fired.Schedule + "/" + fired.Occurrence})
			if err != nil {
				return res, err
			}
			if added {
				res.Fired++
			}
		}

		booked, errs := x.ProcessPending(ctx)
		res.Booked += booked
		res.Errors = append(res.Errors, errs...)
		res.Opened = append(res.Opened, d)
	}
	return res, nil
}

// Fired is the payload of one schedule.fired event: a snapshot of what the
// clock knew — which occurrence is due, on which day it fired, and the
// advanced next_run the schedule's amend rule will set. Everything a rule
// needs rides in the payload; nothing reads state at match time.
type Fired struct {
	Schedule   string `json:"schedule"`
	Name       string `json:"name"`
	Occurrence string `json:"occurrence"` // the due date being honored
	Date       string `json:"date"`       // the day it actually fired
	NextRun    string `json:"next_run"`   // where the cadence moves the schedule
}

// dueOccurrences scans the projected schedules for everything due on or
// before the day, ordered by object id for determinism. One occurrence per
// schedule per day: an overdue schedule catches up one cadence step per
// opened day, late and in order, never silently skipped — the re-entry gate
// is what warns the human about the burst.
func dueOccurrences(scheds []core.Object, day string) []Fired {
	sort.Slice(scheds, func(i, j int) bool { return scheds[i].ID < scheds[j].ID })
	var out []Fired
	for _, o := range scheds {
		status, _ := o.State["status"].(string)
		nextRun, _ := o.State["next_run"].(string)
		cadence, _ := o.State["cadence"].(string)
		name, _ := o.State["name"].(string)
		if status != "active" || nextRun == "" || nextRun > day {
			continue
		}
		advanced, err := nextOccurrence(nextRun, cadence)
		if err != nil {
			continue // an unreadable schedule never fires; it shows in its own detail
		}
		out = append(out, Fired{Schedule: o.ID, Name: name,
			Occurrence: nextRun, Date: day, NextRun: advanced})
	}
	return out
}

// nextOccurrence is the cadence arithmetic — calendar knowledge, owned by
// the clock, never by templates. Monthly and yearly clamp to the target
// month's last day (a schedule on the 31st runs Nov 30, not Dec 1).
func nextOccurrence(from, cadence string) (string, error) {
	t, err := time.Parse("2006-01-02", from)
	if err != nil {
		return "", err
	}
	switch cadence {
	case "daily":
		return t.AddDate(0, 0, 1).Format("2006-01-02"), nil
	case "weekly":
		return t.AddDate(0, 0, 7).Format("2006-01-02"), nil
	case "monthly":
		return addMonthsClamped(t, 1).Format("2006-01-02"), nil
	case "yearly":
		return addMonthsClamped(t, 12).Format("2006-01-02"), nil
	}
	return "", fmt.Errorf("unknown cadence %q", cadence)
}

func addMonthsClamped(t time.Time, months int) time.Time {
	y, m, d := t.Date()
	first := time.Date(y, m+time.Month(months), 1, 0, 0, 0, 0, time.UTC)
	if last := first.AddDate(0, 1, -1).Day(); d > last {
		d = last
	}
	return time.Date(first.Year(), first.Month(), d, 0, 0, 0, 0, time.UTC)
}

func nextDay(d string) (string, error) {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return "", fmt.Errorf("bad date %q", d)
	}
	return t.AddDate(0, 0, 1).Format("2006-01-02"), nil
}

// SimulateCatchUp is the gate's dry run: what opening the pending days
// would fire and book, computed in memory against the live log, nothing
// written. The clock assumes its own advance (a fired schedule moves one
// cadence step) when laying out later days — the diff then shows what the
// rules actually do with every event, including fired occurrences nobody
// explains, which appear as unexplained residue.
func SimulateCatchUp(ctx context.Context, s *store.Store, x *exec.Executor, c Clock, through string) (exec.SimDiff, []string, error) {
	st, err := ClockState(ctx, s, c)
	if err != nil {
		return exec.SimDiff{}, nil, err
	}
	days := st.Pending
	if through != "" {
		cut := 0
		for _, d := range days {
			if d > through {
				break
			}
			cut++
		}
		days = days[:cut]
	}
	scheds, err := s.ObjectsByType(ctx, ScheduleType)
	if err != nil {
		return exec.SimDiff{}, nil, err
	}
	// Local copies: the dry run advances next_run itself, standing in for
	// the amend rule a live catch-up relies on.
	local := make([]core.Object, len(scheds))
	for i, o := range scheds {
		state := make(map[string]any, len(o.State))
		for k, v := range o.State {
			state[k] = v
		}
		o.State = state
		local[i] = o
	}
	maxID, err := s.MaxEventID(ctx)
	if err != nil {
		return exec.SimDiff{}, nil, err
	}
	var hypo []core.Event
	nextID := maxID
	add := func(typ, date string, payload any) {
		nextID++
		b, _ := json.Marshal(payload)
		hypo = append(hypo, core.Event{ID: nextID, Kind: core.KindRaw, Type: typ,
			OccurredAt: date, Payload: b, Actor: "adapter:clock"})
	}
	prev := st.LastOpened
	for _, d := range days {
		add(core.EventDayOpened, d, map[string]any{"date": d})
		if prev == "" || d[:7] != prev[:7] {
			add(core.EventMonthOpened, d, map[string]any{"month": d[:7], "date": d})
		}
		prev = d
		for _, fired := range dueOccurrences(local, d) {
			add(EventScheduleFired, d, fired)
			for i := range local {
				if local[i].ID == fired.Schedule {
					local[i].State["next_run"] = fired.NextRun
				}
			}
		}
	}
	diff, err := x.SimulateEvents(ctx, hypo)
	return diff, days, err
}
