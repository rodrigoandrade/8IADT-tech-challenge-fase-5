package web

import (
	"net/http"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/http/handler"
)

func NewRouter(
	leadService *application.LeadService,
) http.Handler {
	mux := http.NewServeMux()

	leadHandler := handler.NewLeadHandler(leadService)

	mux.HandleFunc("POST /leads", leadHandler.Create)
	mux.HandleFunc("GET /leads", leadHandler.List)
	mux.HandleFunc("GET /leads/{id}", leadHandler.GetByID)
	mux.HandleFunc("PUT /leads/{id}", leadHandler.Update)
	mux.HandleFunc("PATCH /leads/{id}/status", leadHandler.UpdateStatus)

	return mux
}
