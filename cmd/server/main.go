package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v3"
	"github.com/suka712/api.sukaseven.com/internal/auth"
	"github.com/suka712/api.sukaseven.com/internal/db"
	"github.com/suka712/api.sukaseven.com/internal/health"
	"github.com/suka712/api.sukaseven.com/util"
)

func main() {
	godotenv.Load()
	util.RequireEnvs()

	ctx := context.Background()
	queries, pool, err := db.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	emailClient := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	authHandler := &auth.Handler{Queries: queries, EmailClient: emailClient}

	r := chi.NewRouter()

	r.Use(cors.Handler(util.CorsOptions()))

	r.Get("/health", health.Health)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/email", authHandler.Email)
		r.Post("/otp", authHandler.OTP)
		r.Get("/session", authHandler.Session)
	})

	port := os.Getenv("PORT")
	log.Printf("✨ Server starting on port %s...", port)
	http.ListenAndServe(":"+port, r)
}
