package http

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/http/handler"
)

func NewRouter(
	leadService *application.LeadService,
	propertyService *application.PropertyService,
	conversationService *application.ConversationService,
	messageService *application.MessageService,
	pool *pgxpool.Pool,
) http.Handler {
	mux := http.NewServeMux()

	leadHandler := handler.NewLeadHandler(leadService)
	propertyHandler := handler.NewPropertyHandler(propertyService)
	conversationHandler := handler.NewConversationHandler(conversationService)
	messageHandler := handler.NewMessageHandler(messageService)

	// =========================================================
	// LEADS
	// =========================================================

	mux.HandleFunc("POST /leads", leadHandler.Create)
	mux.HandleFunc("GET /leads", leadHandler.List)
	mux.HandleFunc("GET /leads/{id}", leadHandler.GetByID)
	mux.HandleFunc("PUT /leads/{id}", leadHandler.Update)
	mux.HandleFunc("PATCH /leads/{id}/status", leadHandler.UpdateStatus)

	mux.HandleFunc(
		"GET /leads/{id}/conversations",
		conversationHandler.ListByLead,
	)

	// =========================================================
	// PROPERTIES
	// =========================================================

	mux.HandleFunc("POST /properties", propertyHandler.Create)
	mux.HandleFunc("GET /properties", propertyHandler.List)
	mux.HandleFunc("GET /properties/{id}", propertyHandler.GetByID)
	mux.HandleFunc("PUT /properties/{id}", propertyHandler.Update)
	mux.HandleFunc(
		"PATCH /properties/{id}/status",
		propertyHandler.UpdateStatus,
	)

	// =========================================================
	// CONVERSATIONS
	// =========================================================

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
		"PATCH /conversations/{id}/close",
		conversationHandler.Close,
	)

	mux.HandleFunc(
		"PUT /conversations/{id}",
		conversationHandler.Close,
	)

	mux.HandleFunc(
		"PATCH /conversations/{id}",
		conversationHandler.Close,
	)

	// =========================================================
	// MESSAGES
	// =========================================================

	mux.HandleFunc(
		"POST /messages",
		messageHandler.Create,
	)

	mux.HandleFunc(
		"GET /messages",
		messageHandler.List,
	)

	mux.HandleFunc(
		"GET /messages/count",
		messageHandler.Count,
	)

	mux.HandleFunc(
		"GET /messages/{id}",
		messageHandler.GetByID,
	)

	mux.HandleFunc(
		"GET /conversations/{id}/messages",
		messageHandler.ListByConversation,
	)

	mux.HandleFunc(
		"GET /conversations/{id}/messages/count",
		messageHandler.CountByConversation,
	)

	// =========================================================
	// HEALTH
	// =========================================================

	mux.HandleFunc(
		"GET /health",
		func(w http.ResponseWriter, r *http.Request) {
			writeJSON(
				w,
				http.StatusOK,
				map[string]string{
					"status": "ok",
				},
			)
		},
	)

	// =========================================================
	// READINESS
	// =========================================================

	mux.HandleFunc(
		"GET /ready",
		func(w http.ResponseWriter, r *http.Request) {
			if pool == nil {
				writeHTTPError(
					w,
					http.StatusServiceUnavailable,
					"database unavailable",
				)
				return
			}

			if err := pool.Ping(r.Context()); err != nil {
				writeHTTPError(
					w,
					http.StatusServiceUnavailable,
					"database unavailable",
				)
				return
			}

			writeJSON(
				w,
				http.StatusOK,
				map[string]string{
					"status":   "UP",
					"database": "UP",
				},
			)
		},
	)

	return mux
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

func writeHTTPError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(
		w,
		status,
		map[string]string{
			"error": message,
		},
	)
}
