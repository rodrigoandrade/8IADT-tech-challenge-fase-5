package main

import (
	"context"
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

	pool, err := postgres.NewPool(
		ctx,
		databaseURL,
	)
	if err != nil {
		log.Fatalf(
			"failed to connect to PostgreSQL: %v",
			err,
		)
	}

	defer pool.Close()

	queries := generated.New(pool)

	// =========================================================
	// REPOSITORIES
	// =========================================================

	leadRepository := postgres.NewLeadRepository(
		queries,
	)

	propertyRepository := postgres.NewPropertyRepository(
		queries,
	)

	conversationRepository := postgres.NewConversationRepository(
		queries,
	)

	messageRepository := postgres.NewMessageRepository(
		queries,
	)

	// =========================================================
	// SERVICES
	// =========================================================

	leadService := application.NewLeadService(
		leadRepository,
	)

	propertyService := application.NewPropertyService(
		propertyRepository,
	)

	conversationService := application.NewConversationService(
		conversationRepository,
	)

	messageService := application.NewMessageService(
		messageRepository,
	)

	// =========================================================
	// ROUTER
	// =========================================================

	router := web.NewRouter(
		leadService,
		propertyService,
		conversationService,
		messageService,
		pool,
	)

	// =========================================================
	// HTTP SERVER
	// =========================================================

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// =========================================================
	// START SERVER
	// =========================================================

	go func() {
		log.Printf(
			"HTTP server listening on :%s",
			port,
		)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf(
				"HTTP server failed: %v",
				err,
			)
		}
	}()

	// =========================================================
	// GRACEFUL SHUTDOWN
	// =========================================================

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Println("shutting down HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf(
			"HTTP server shutdown error: %v",
			err,
		)
	}

	log.Println("HTTP server stopped")
}
