package shell

import (
	"net/http"
	"strconv"
	"strings"
)

// The file is the fact (DIRECTION 2026-10-10): a stored file is read back
// by its hash, with the media type it was received as. The shell links any
// file fact it renders here; it knows no particular document.
func (s *Server) handleBlob(w http.ResponseWriter, r *http.Request) {
	if !s.allowSystem(w, r, ViewFiles) {
		return
	}
	b, err := s.Store.GetBlob(r.Context(), r.PathValue("hash"))
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	w.Header().Set("Content-Type", b.MediaType)
	w.Header().Set("Content-Length", strconv.FormatInt(b.Size, 10))
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable") // content-addressed: it never changes
	// A stored file is a sender's bytes served from the shell's own origin:
	// an emailed invoice.html could otherwise script the approval doors as
	// the person looking at it. The browser may not sniff the type, the
	// document runs in a sandbox with no script and no origin, and anything
	// that is not a PDF or a raster image downloads instead of rendering.
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	if !inlineMediaType(b.MediaType) {
		w.Header().Set("Content-Disposition", "attachment")
	}
	w.Write(b.Bytes)
}

// inlineMediaType is what the browser may render in place: a PDF or a
// raster image. HTML, SVG and everything else download.
func inlineMediaType(mt string) bool {
	switch strings.ToLower(strings.TrimSpace(strings.SplitN(mt, ";", 2)[0])) {
	case "application/pdf", "image/png", "image/jpeg", "image/gif", "image/webp":
		return true
	}
	return false
}
