package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"real-estate-sdr/internal/application"
	web "real-estate-sdr/internal/http"
	"real-estate-sdr/internal/infrastructure/postgres"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
)

func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not configured")
	}

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf(
			"failed to connect to PostgreSQL: %v",
			err,
		)
	}
	defer pool.Close()

	log.Println(
		"PostgreSQL connection established successfully",
	)

	queries := generated.New(pool)

	// Repositories
	leadRepository := postgres.NewLeadRepository(
		queries,
	)

	propertyRepository := postgres.NewPropertyRepository(
		queries,
	)

	conversationRepository := postgres.NewConversationRepository(
		queries,
	)

	// Services
	leadService := application.NewLeadService(
		leadRepository,
	)

	propertyService := application.NewPropertyService(
		propertyRepository,
	)

	conversationService := application.NewConversationService(
		conversationRepository,
	)

	// HTTP Router
	router := web.NewRouter(
		leadService,
		propertyService,
		conversationService,
		pool,
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println(
		"HTTP server listening on :8080",
	)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf(
			"HTTP server failed: %v",
			err,
		)
	}
}
