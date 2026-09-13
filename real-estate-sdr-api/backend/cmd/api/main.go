package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"real-estate-sdr/internal/infrastructure/ai"
	"syscall"
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

	// =========================================================
	// DATABASE
	// =========================================================

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

	// =========================================================
	// MIGRATION MODE
	// =========================================================
	//
	// Docker executa:
	//
	//     /app/api migrate
	//
	// Atualmente o projeto não possui um mecanismo de migration
	// implementado em Go. Portanto, neste modo apenas validamos
	// que o banco está acessível e encerramos com sucesso.
	//
	// As migrations/schema devem ser executadas pela infraestrutura
	// do PostgreSQL.
	//

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		log.Println("database connection verified")
		log.Println("migration step completed")

		return
	}

	// =========================================================
	// REPOSITORIES
	// =========================================================

	queries := generated.New(pool)

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

	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		log.Fatal("AI_SERVICE_URL is not configured")
	}

	aiService := ai.NewHTTPClient(aiServiceURL, nil)

	messageService := application.NewMessageService(
		messageRepository,
		conversationRepository,
		aiService,
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

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}

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
