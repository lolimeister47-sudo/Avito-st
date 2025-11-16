package httpadapter

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"prservice/internal/adapter/http/api"
)

func NewRouter(server api.ServerInterface) http.Handler {
	r := chi.NewRouter()

	// healthcheck
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// все маршруты из OpenAPI
	api.HandlerFromMux(server, r)

	return r
}
