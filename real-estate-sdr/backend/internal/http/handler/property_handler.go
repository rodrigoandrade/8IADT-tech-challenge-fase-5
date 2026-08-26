package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"real-estate-sdr/internal/repository"
	"strconv"

	"github.com/google/uuid"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/domain"
)

type PropertyHandler struct {
	service *application.PropertyService
}

func NewPropertyHandler(
	service *application.PropertyService,
) *PropertyHandler {
	return &PropertyHandler{
		service: service,
	}
}

type createPropertyRequest struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Type          string  `json:"type"`
	Transaction   string  `json:"transaction_type"`
	Price         float64 `json:"price"`
	Bedrooms      int     `json:"bedrooms"`
	Bathrooms     int     `json:"bathrooms"`
	ParkingSpaces int     `json:"parking_spaces"`
	Area          float64 `json:"area"`
	City          string  `json:"city"`
	Neighborhood  string  `json:"neighborhood"`
	Address       string  `json:"address"`
	Status        string  `json:"status"`
}

type updatePropertyRequest struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Type          string  `json:"type"`
	Transaction   string  `json:"transaction_type"`
	Price         float64 `json:"price"`
	Bedrooms      int     `json:"bedrooms"`
	Bathrooms     int     `json:"bathrooms"`
	ParkingSpaces int     `json:"parking_spaces"`
	Area          float64 `json:"area"`
	City          string  `json:"city"`
	Neighborhood  string  `json:"neighborhood"`
	Address       string  `json:"address"`
	Status        string  `json:"status"`
}

type updatePropertyStatusRequest struct {
	Status string `json:"status"`
}

// POST /properties
func (h *PropertyHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request createPropertyRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	property := &domain.Property{
		Title:         request.Title,
		Description:   request.Description,
		Type:          domain.PropertyType(request.Type),
		Transaction:   domain.TransactionType(request.Transaction),
		Price:         request.Price,
		Bedrooms:      request.Bedrooms,
		Bathrooms:     request.Bathrooms,
		ParkingSpaces: request.ParkingSpaces,
		Area:          request.Area,
		City:          request.City,
		Neighborhood:  request.Neighborhood,
		Address:       request.Address,
		Status:        domain.PropertyStatus(request.Status),
	}

	createdProperty, err := h.service.Create(
		r.Context(),
		property,
	)
	if err != nil {
		handlePropertyServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		createdProperty,
	)
}

// GET /properties/{id}
func (h *PropertyHandler) GetByID(
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
			"invalid property id",
		)
		return
	}

	property, err := h.service.GetByID(
		r.Context(),
		id,
	)
	if err != nil {
		handlePropertyServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		property,
	)
}

// GET /properties
func (h *PropertyHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	limit := 20
	offset := 0

	if value := r.URL.Query().Get("limit"); value != "" {
		parsedLimit, err := strconv.Atoi(value)
		if err != nil {
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
		if err != nil {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid offset",
			)
			return
		}

		offset = parsedOffset
	}

	properties, err := h.service.List(
		r.Context(),
		limit,
		offset,
	)
	if err != nil {
		handlePropertyServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		properties,
	)
}

// PUT /properties/{id}
func (h *PropertyHandler) Update(
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
			"invalid property id",
		)
		return
	}

	var request updatePropertyRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	property := &domain.Property{
		ID:            id,
		Title:         request.Title,
		Description:   request.Description,
		Type:          domain.PropertyType(request.Type),
		Transaction:   domain.TransactionType(request.Transaction),
		Price:         request.Price,
		Bedrooms:      request.Bedrooms,
		Bathrooms:     request.Bathrooms,
		ParkingSpaces: request.ParkingSpaces,
		Area:          request.Area,
		City:          request.City,
		Neighborhood:  request.Neighborhood,
		Address:       request.Address,
		Status:        domain.PropertyStatus(request.Status),
	}

	updatedProperty, err := h.service.Update(
		r.Context(),
		property,
	)
	if err != nil {
		handlePropertyServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		updatedProperty,
	)
}

// PATCH /properties/{id}/status
func (h *PropertyHandler) UpdateStatus(
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
			"invalid property id",
		)
		return
	}

	var request updatePropertyStatusRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	property, err := h.service.UpdateStatus(
		r.Context(),
		id,
		domain.PropertyStatus(request.Status),
	)
	if err != nil {
		handlePropertyServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		property,
	)
}

func handlePropertyServiceError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(
		err,
		domain.ErrInvalidProperty,
	):
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(
		err,
		domain.ErrInvalidPropertyType,
	):
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(
		err,
		domain.ErrInvalidTransactionType,
	):
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(
		err,
		domain.ErrInvalidPropertyStatus,
	):
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(
		err,
		repository.ErrPropertyNotFound,
	):
		writeError(
			w,
			http.StatusNotFound,
			err.Error(),
		)

	case errors.Is(
		err,
		repository.ErrPropertyAlreadyExists,
	):
		writeError(
			w,
			http.StatusConflict,
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
