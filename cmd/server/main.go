package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v3"
	"github.com/suka712/api.sukaseven.com/internal/auth"
	"github.com/suka712/api.sukaseven.com/internal/db"
	"github.com/suka712/api.sukaseven.com/internal/health"
)

func main() {
	godotenv.Load()
	r := chi.NewRouter()

	ctx := context.Background()
	queries, pool, err := db.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	emailClient := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	authHandler := &auth.Handler{Queries: queries, EmailClient: emailClient}

	r.Get("/health", health.Health)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/email", authHandler.Email)
	})

	log.Print("Server starting on port 8080")
	http.ListenAndServe(":"+"8080", r)
}
