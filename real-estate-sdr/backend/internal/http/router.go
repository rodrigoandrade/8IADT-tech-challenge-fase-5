package web

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/http/handler"
)

func NewRouter(
	leadService *application.LeadService,
	propertyService *application.PropertyService,
	pool *pgxpool.Pool,
) http.Handler {
	mux := http.NewServeMux()

	leadHandler := handler.NewLeadHandler(leadService)
	propertyHandler := handler.NewPropertyHandler(propertyService)
	healthHandler := handler.NewHealthHandler(pool)

	// ---------------------------------------------------------
	// HEALTH
	// ---------------------------------------------------------

	mux.HandleFunc(
		"GET /health",
		healthHandler.Check,
	)

	mux.HandleFunc(
		"GET /ready",
		healthHandler.Ready,
	)

	// ---------------------------------------------------------
	// LEADS
	// ---------------------------------------------------------

	mux.HandleFunc(
		"POST /leads",
		leadHandler.Create,
	)

	mux.HandleFunc(
		"GET /leads",
		leadHandler.List,
	)

	mux.HandleFunc(
		"GET /leads/{id}",
		leadHandler.GetByID,
	)

	mux.HandleFunc(
		"PUT /leads/{id}",
		leadHandler.Update,
	)

	mux.HandleFunc(
		"PATCH /leads/{id}/status",
		leadHandler.UpdateStatus,
	)

	// ---------------------------------------------------------
	// PROPERTIES
	// ---------------------------------------------------------

	mux.HandleFunc(
		"POST /properties",
		propertyHandler.Create,
	)

	mux.HandleFunc(
		"GET /properties",
		propertyHandler.List,
	)

	mux.HandleFunc(
		"GET /properties/{id}",
		propertyHandler.GetByID,
	)

	mux.HandleFunc(
		"PUT /properties/{id}",
		propertyHandler.Update,
	)

	mux.HandleFunc(
		"PATCH /properties/{id}/status",
		propertyHandler.UpdateStatus,
	)

	return mux
}
