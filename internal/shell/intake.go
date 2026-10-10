package shell

import (
	"net/http"
	"strconv"
)

// The file is the fact (DIRECTION 2026-10-10): a stored file is read back
// by its hash, with the media type it was received as. The shell links any
// file fact it renders here; it knows no particular document.
func (s *Server) handleBlob(w http.ResponseWriter, r *http.Request) {
	b, err := s.Store.GetBlob(r.Context(), r.PathValue("hash"))
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	w.Header().Set("Content-Type", b.MediaType)
	w.Header().Set("Content-Length", strconv.FormatInt(b.Size, 10))
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable") // content-addressed: it never changes
	w.Write(b.Bytes)
}
