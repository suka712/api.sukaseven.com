package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/suka712/api.sukaseven.com/internal/health"
)

func main() {
	r := chi.NewRouter()

	r.Get("/health", health.Health)

	log.Print("Server starting on port 8080")
	http.ListenAndServe(":"+"8080", r)
}
