package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHealthHandler_Check(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	handler := NewHealthHandler(pool)

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Check(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	expected := `{"status":"UP"}`

	if rec.Body.String() != expected+"\n" {
		t.Fatalf(
			"expected body %q, got %q",
			expected+"\n",
			rec.Body.String(),
		)
	}
}

func TestHealthHandler_Ready(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	handler := NewHealthHandler(pool)

	req := httptest.NewRequest(
		http.MethodGet,
		"/ready",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Ready(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	var response map[string]string

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if response["status"] != "UP" {
		t.Fatalf(
			"expected status UP, got %q",
			response["status"],
		)
	}

	if response["database"] != "UP" {
		t.Fatalf(
			"expected database UP, got %q",
			response["database"],
		)
	}
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not configured")
	}

	pool, err := pgxpool.New(
		context.Background(),
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"failed to create PostgreSQL pool: %v",
			err,
		)
	}

	return pool
}
