package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/domain"
)

type MessageHandler struct {
	service *application.MessageService
}

type createMessageRequest struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
}

func NewMessageHandler(service *application.MessageService) *MessageHandler {
	return &MessageHandler{
		service: service,
	}
}

// Create godoc
// POST /messages
func (h *MessageHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request createMessageRequest

	if err := decodeJSON(r, &request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	message := domain.Message{
		ConversationID: request.ConversationID,
		Role:           request.Role,
		Content:        request.Content,
	}

	created, err := h.service.Create(r.Context(), &message)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// GetByID godoc
// GET /messages/{id}
func (h *MessageHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := extractUUIDFromPath(r, "messages")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid message id")
		return
	}

	message, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			writeJSONError(w, http.StatusNotFound, "message not found")
			return
		}

		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, message)
}

// List godoc
// GET /messages?limit=20&offset=0
func (h *MessageHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	messages, err := h.service.List(
		r.Context(),
		limit,
		offset,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, messages)
}

// ListByConversation godoc
// GET /conversations/{id}/messages
func (h *MessageHandler) ListByConversation(w http.ResponseWriter, r *http.Request) {
	conversationID, err := extractUUIDFromPath(r, "conversations")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	messages, err := h.service.ListByConversation(
		r.Context(),
		conversationID,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, messages)
}

// Count godoc
// GET /messages/count
func (h *MessageHandler) Count(w http.ResponseWriter, r *http.Request) {
	count, err := h.service.Count(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{
		"count": count,
	})
}

// CountByConversation godoc
// GET /conversations/{id}/messages/count
func (h *MessageHandler) CountByConversation(w http.ResponseWriter, r *http.Request) {
	conversationID, err := extractUUIDFromPath(r, "conversations")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	count, err := h.service.CountByConversation(
		r.Context(),
		conversationID,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{
		"count": count,
	})
}

func parsePagination(r *http.Request) (int, int, error) {
	const (
		defaultLimit = 20
		maxLimit     = 100
	)

	limit := defaultLimit
	offset := 0

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			return 0, 0, &paginationError{"invalid limit"}
		}

		if parsed > maxLimit {
			parsed = maxLimit
		}

		limit = parsed
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return 0, 0, &paginationError{"invalid offset"}
		}

		offset = parsed
	}

	return limit, offset, nil
}

type paginationError struct {
	message string
}

func (e *paginationError) Error() string {
	return e.message
}

func extractUUIDFromPath(
	r *http.Request,
	resource string,
) (uuid.UUID, error) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	for i, part := range parts {
		if part == resource && i+1 < len(parts) {
			return uuid.Parse(parts[i+1])
		}
	}

	return uuid.Nil, fmt.Errorf("invalid %s id", resource)
}
