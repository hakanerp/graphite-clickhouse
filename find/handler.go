package find

import (
	"net/http"

	"github.com/lomik/graphite-clickhouse/config"
)

type Handler struct {
	config *config.Config
}

func NewHandler(config *config.Config) *Handler {
	return &Handler{
		config: config,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	f, err := New(h.config, r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.Reply(w, r, f)
}

func (h *Handler) Reply(w http.ResponseWriter, r *http.Request, f *Find) {
	switch r.URL.Query().Get("format") {
	case "pickle":
		f.WritePickle(w)
	case "protobuf":
		f.WriteProtobuf(w)
	}
}
