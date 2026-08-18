package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"real-estate-sdr/internal/infrastructure/postgres"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
	"syscall"
	"time"

	"real-estate-sdr/internal/application"
	web "real-estate-sdr/internal/http"
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

	leadRepository := postgres.NewLeadRepository(queries)

	leadService := application.NewLeadService(
		leadRepository,
	)

	router := web.NewRouter(
		leadService,
		pool,
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		log.Println("HTTP server listening on :8080")

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatalf(
				"HTTP server failed: %v",
				err,
			)
		}
	}()

	<-serverCtx.Done()

	log.Println("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf(
			"HTTP server shutdown failed: %v",
			err,
		)
	} else {
		log.Println("HTTP server stopped gracefully")
	}

	log.Println("Application stopped")
}
