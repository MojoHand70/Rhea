package shell

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"rhea/internal/network"
)

// lessonOut is one thing installations were never taught, with where the
// team stands on it.
type lessonOut struct {
	network.Untaught
	Status  string     `json:"status"` // open until the team says otherwise
	Note    string     `json:"note,omitempty"`
	NotedBy string     `json:"noted_by,omitempty"`
	NotedAt *time.Time `json:"noted_at,omitempty"`
}

type learnedOut struct {
	Question string           `json:"question"`
	Answers  []network.Answer `json:"answers"`
}

// handleLearning is the Learning Centre's read (DIRECTION 2026-10-17): what
// installations were asked and had not been taught, counted, with the team's
// notes on each; and what Rhea has learned across them, by question. It reads
// the house's network, never an installation's books. ?synthetic=1 counts
// simulated customers too.
func (s *Server) handleLearning(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	synthetic := r.URL.Query().Get("synthetic") == "1"
	if s.Network == nil {
		writeJSON(w, 200, map[string]any{"network": false})
		return
	}
	us, err := s.Network.Untaught(ctx, synthetic)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	notes, err := s.Network.Lessons(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	lessons := make([]lessonOut, 0, len(us))
	for _, u := range us {
		l := lessonOut{Untaught: u, Status: "open"}
		if n, ok := notes[u.Lesson]; ok {
			l.Status, l.Note, l.NotedBy, l.NotedAt = n.Status, n.Note, n.By, &n.At
		}
		lessons = append(lessons, l)
	}
	learn := s.Network.Learn
	if synthetic {
		learn = s.Network.LearnIncludingSynthetic
	}
	k, err := learn(ctx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	learned := make([]learnedOut, 0, len(k.Questions))
	for q, as := range k.Questions {
		learned = append(learned, learnedOut{Question: q, Answers: as})
	}
	sort.Slice(learned, func(i, j int) bool { return learned[i].Question < learned[j].Question })
	writeJSON(w, 200, map[string]any{"network": true, "synthetic": synthetic, "floor": network.Floor,
		"published": k.Installations, "lessons": lessons, "learned": learned, "statuses": network.LessonStatuses})
}

// handleNoteLesson files the team's note on a lesson: where it stands, and why.
func (s *Server) handleNoteLesson(w http.ResponseWriter, r *http.Request) {
	if s.Network == nil {
		writeErr(w, 503, fmt.Errorf("lesson: no network is connected"))
		return
	}
	var req struct {
		Lesson string `json:"lesson"`
		Status string `json:"status"`
		Note   string `json:"note"`
		By     string `json:"by"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	req.Lesson, req.Note = strings.TrimSpace(req.Lesson), strings.TrimSpace(req.Note)
	if req.Lesson == "" || len(req.Lesson) > 400 || len(req.Note) > 2000 || req.By == "" {
		writeErr(w, 400, fmt.Errorf("lesson: lesson (1–400 bytes), status and by are required, a note up to 2000"))
		return
	}
	at, err := s.Network.NoteLesson(r.Context(), req.Lesson, req.Status, req.Note, req.By)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"lesson": strings.ToLower(req.Lesson), "status": req.Status, "at": at})
}
