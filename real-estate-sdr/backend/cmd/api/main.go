package main

import (
	"context"
	"log"
	"net/http"
	"os"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
	"time"

	"real-estate-sdr/internal/application"
	web "real-estate-sdr/internal/http"
	"real-estate-sdr/internal/infrastructure/postgres"
)

func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not configured")
	}

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	log.Println("PostgreSQL connection established successfully")

	queries := generated.New(pool)

	leadRepository := postgres.NewLeadRepository(queries)

	leadService := application.NewLeadService(leadRepository)

	router := web.NewRouter(leadService)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("HTTP server listening on :8080")

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
