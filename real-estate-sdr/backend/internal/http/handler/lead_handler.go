package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/domain"
)

type LeadHandler struct {
	service *application.LeadService
}

func NewLeadHandler(service *application.LeadService) *LeadHandler {
	return &LeadHandler{
		service: service,
	}
}

// POST /leads
func (h *LeadHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request createLeadRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	lead := &domain.Lead{
		Name:  request.Name,
		Email: request.Email,
		Phone: request.Phone,
	}

	createdLead, err := h.service.Create(r.Context(), lead)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		createdLead,
	)
}

// GET /leads/{id}
func (h *LeadHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid lead id",
		)
		return
	}

	lead, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		lead,
	)
}

// GET /leads
func (h *LeadHandler) List(w http.ResponseWriter, r *http.Request) {
	leads, err := h.service.List(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		leads,
	)
}

// PUT /leads/{id}
func (h *LeadHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid lead id",
		)
		return
	}

	var request updateLeadRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	lead := &domain.Lead{
		ID:     id,
		Name:   request.Name,
		Email:  request.Email,
		Phone:  request.Phone,
		Status: domain.LeadStatus(request.Status),
	}

	updatedLead, err := h.service.Update(r.Context(), lead)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		updatedLead,
	)
}

// PATCH /leads/{id}/status
func (h *LeadHandler) UpdateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid lead id",
		)
		return
	}

	var request updateLeadStatusRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	status := domain.LeadStatus(
		strings.TrimSpace(request.Status),
	)

	updatedLead, err := h.service.UpdateStatus(
		r.Context(),
		id,
		status,
	)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		updatedLead,
	)
}

type createLeadRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type updateLeadRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

type updateLeadStatusRequest struct {
	Status string `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func handleServiceError(
	w http.ResponseWriter,
	err error,
) {
	if errors.Is(err, application.ErrInvalidLead) {
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	if errors.Is(err, application.ErrLeadNotFound) {
		writeError(
			w,
			http.StatusNotFound,
			err.Error(),
		)
		return
	}

	writeError(
		w,
		http.StatusInternalServerError,
		"internal server error",
	)
}

func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(
		w,
		status,
		errorResponse{
			Error: message,
		},
	)
}
