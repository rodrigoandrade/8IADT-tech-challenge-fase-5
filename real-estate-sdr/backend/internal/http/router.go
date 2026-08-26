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
	conversationService *application.ConversationService,
	pool *pgxpool.Pool,
) http.Handler {
	mux := http.NewServeMux()

	leadHandler := handler.NewLeadHandler(
		leadService,
	)

	propertyHandler := handler.NewPropertyHandler(
		propertyService,
	)

	conversationHandler := handler.NewConversationHandler(
		conversationService,
	)

	healthHandler := handler.NewHealthHandler(
		pool,
	)

	// Health
	mux.HandleFunc(
		"GET /health",
		healthHandler.Check,
	)

	mux.HandleFunc(
		"GET /ready",
		healthHandler.Ready,
	)

	// Leads
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

	// Properties
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

	// Conversations
	mux.HandleFunc(
		"POST /conversations",
		conversationHandler.Create,
	)

	mux.HandleFunc(
		"GET /conversations",
		conversationHandler.List,
	)

	mux.HandleFunc(
		"GET /conversations/{id}",
		conversationHandler.GetByID,
	)

	mux.HandleFunc(
		"GET /leads/{leadID}/conversations",
		conversationHandler.ListByLead,
	)

	mux.HandleFunc(
		"PATCH /conversations/{id}/close",
		conversationHandler.Close,
	)

	return mux
}
