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

	// ---------------------------------------------------------
	// LEAD
	// ---------------------------------------------------------

	leadRepository := postgres.NewLeadRepository(
		queries,
	)

	leadService := application.NewLeadService(
		leadRepository,
	)

	// ---------------------------------------------------------
	// PROPERTY
	// ---------------------------------------------------------

	propertyRepository := postgres.NewPropertyRepository(
		queries,
	)

	propertyService := application.NewPropertyService(
		propertyRepository,
	)

	// ---------------------------------------------------------
	// HTTP
	// ---------------------------------------------------------

	router := web.NewRouter(
		leadService,
		propertyService,
		pool,
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("HTTP server listening on :8080")

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf(
			"HTTP server failed: %v",
			err,
		)
	}
}
