package slashfix

import (
	"net/http"
	"strings"
)

type SkipSlashMux struct {
	Mux *http.ServeMux
}

func NewSkipSlashMux() *SkipSlashMux {
	return &SkipSlashMux{http.NewServeMux()}
}
func (h *SkipSlashMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	h.Mux.HandleFunc(pattern, handler)
}

func (h *SkipSlashMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, "//", "/", -1)
	h.Mux.ServeHTTP(w, r)
}
