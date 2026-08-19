package web

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/http/handler"
)

func NewRouter(
	leadService *application.LeadService,
	pool *pgxpool.Pool,
) http.Handler {
	mux := http.NewServeMux()

	leadHandler := handler.NewLeadHandler(leadService)
	healthHandler := handler.NewHealthHandler(pool)

	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /ready", healthHandler.Ready)

	mux.HandleFunc("POST /leads", leadHandler.Create)
	mux.HandleFunc("GET /leads", leadHandler.List)
	mux.HandleFunc("GET /leads/{id}", leadHandler.GetByID)
	mux.HandleFunc("PUT /leads/{id}", leadHandler.Update)
	mux.HandleFunc("PATCH /leads/{id}/status", leadHandler.UpdateStatus)

	return mux
}
