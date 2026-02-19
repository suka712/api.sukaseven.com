package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/suka712/api.sukaseven.com/internal/health"
)

func main() {
	r := chi.NewRouter()
	
	r.Get("/health", health.HandleHealth)

	http.ListenAndServe(":" + "8080", r)
}
