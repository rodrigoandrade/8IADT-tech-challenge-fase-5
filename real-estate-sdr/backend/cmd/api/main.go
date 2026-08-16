package main

import (
	"context"
	"log"
	"os"

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
		log.Fatal(err)
	}
	defer pool.Close()

	log.Println("PostgreSQL connection established successfully")
}
