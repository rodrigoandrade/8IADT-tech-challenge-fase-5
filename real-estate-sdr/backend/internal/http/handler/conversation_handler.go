package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/domain"
)

type ConversationHandler struct {
	service *application.ConversationService
}

func NewConversationHandler(
	service *application.ConversationService,
) *ConversationHandler {
	return &ConversationHandler{
		service: service,
	}
}

// POST /conversations
func (h *ConversationHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request createConversationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	conversation := &domain.Conversation{
		LeadID: request.LeadID,
		Status: domain.ConversationStatusActive,
	}

	createdConversation, err := h.service.Create(
		r.Context(),
		conversation,
	)
	if err != nil {
		handleConversationServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		createdConversation,
	)
}

// GET /conversations/{id}
func (h *ConversationHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := uuid.Parse(
		r.PathValue("id"),
	)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid conversation id",
		)
		return
	}

	conversation, err := h.service.GetByID(
		r.Context(),
		id,
	)
	if err != nil {
		handleConversationServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		conversation,
	)
}

// GET /conversations
func (h *ConversationHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	limit := 20
	offset := 0

	if value := r.URL.Query().Get("limit"); value != "" {
		parsedLimit, err := strconv.Atoi(value)
		if err != nil || parsedLimit <= 0 {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid limit",
			)
			return
		}

		limit = parsedLimit
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		parsedOffset, err := strconv.Atoi(value)
		if err != nil || parsedOffset < 0 {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid offset",
			)
			return
		}

		offset = parsedOffset
	}

	conversations, err := h.service.List(
		r.Context(),
		limit,
		offset,
	)
	if err != nil {
		handleConversationServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		conversations,
	)
}

// GET /leads/{leadID}/conversations
func (h *ConversationHandler) ListByLead(
	w http.ResponseWriter,
	r *http.Request,
) {
	leadID, err := uuid.Parse(
		r.PathValue("leadID"),
	)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid lead id",
		)
		return
	}

	conversations, err := h.service.GetByLeadID(
		r.Context(),
		leadID,
	)
	if err != nil {
		handleConversationServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		conversations,
	)
}

// PATCH /conversations/{id}/close
func (h *ConversationHandler) Close(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := uuid.Parse(
		r.PathValue("id"),
	)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid conversation id",
		)
		return
	}

	conversation, err := h.service.Close(
		r.Context(),
		id,
	)
	if err != nil {
		handleConversationServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		conversation,
	)
}

type createConversationRequest struct {
	LeadID uuid.UUID `json:"lead_id"`
}

func handleConversationServiceError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(
		err,
		application.ErrConversationNotFound,
	):
		writeError(
			w,
			http.StatusNotFound,
			err.Error(),
		)

	case errors.Is(
		err,
		application.ErrConversationAlreadyClosed,
	):
		writeError(
			w,
			http.StatusConflict,
			err.Error(),
		)

	case errors.Is(
		err,
		domain.ErrInvalidConversation,
	):
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(
		err,
		domain.ErrConversationLeadRequired,
	):
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(
		err,
		domain.ErrInvalidConversationStatus,
	):
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	default:
		writeError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
	}
}
