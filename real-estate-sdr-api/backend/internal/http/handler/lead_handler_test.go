package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/domain"
)

type fakeLeadRepository struct {
	createFunc       func(context.Context, *domain.Lead) (*domain.Lead, error)
	getByIDFunc      func(context.Context, uuid.UUID) (*domain.Lead, error)
	listFunc         func(context.Context) ([]*domain.Lead, error)
	updateFunc       func(context.Context, *domain.Lead) (*domain.Lead, error)
	updateStatusFunc func(context.Context, uuid.UUID, domain.LeadStatus) (*domain.Lead, error)
	countFunc        func(context.Context) (int64, error)
}

func (f *fakeLeadRepository) Create(
	ctx context.Context,
	lead *domain.Lead,
) (*domain.Lead, error) {
	if f.createFunc != nil {
		return f.createFunc(ctx, lead)
	}

	return nil, errors.New("create not implemented")
}

func (f *fakeLeadRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Lead, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}

	return nil, errors.New("get by id not implemented")
}

func (f *fakeLeadRepository) List(
	ctx context.Context,
) ([]*domain.Lead, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx)
	}

	return nil, errors.New("list not implemented")
}

func (f *fakeLeadRepository) Update(
	ctx context.Context,
	lead *domain.Lead,
) (*domain.Lead, error) {
	if f.updateFunc != nil {
		return f.updateFunc(ctx, lead)
	}

	return nil, errors.New("update not implemented")
}

func (f *fakeLeadRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.LeadStatus,
) (*domain.Lead, error) {
	if f.updateStatusFunc != nil {
		return f.updateStatusFunc(ctx, id, status)
	}

	return nil, errors.New("update status not implemented")
}

func (f *fakeLeadRepository) Count(
	ctx context.Context,
) (int64, error) {
	if f.countFunc != nil {
		return f.countFunc(ctx)
	}

	return 0, errors.New("count not implemented")
}

func newTestHandler(
	repository *fakeLeadRepository,
) *LeadHandler {
	service := application.NewLeadService(repository)

	return NewLeadHandler(service)
}

func TestLeadHandler_Create(t *testing.T) {
	leadID := uuid.New()

	repository := &fakeLeadRepository{
		createFunc: func(
			_ context.Context,
			lead *domain.Lead,
		) (*domain.Lead, error) {
			return &domain.Lead{
				ID:        leadID,
				Name:      lead.Name,
				Email:     lead.Email,
				Phone:     lead.Phone,
				Status:    lead.Status,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	handler := newTestHandler(repository)

	body := `{
		"name": "João Silva",
		"email": "joao@example.com",
		"phone": "11999999999"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/leads",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	var response domain.Lead

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if response.ID != leadID {
		t.Errorf(
			"expected ID %s, got %s",
			leadID,
			response.ID,
		)
	}

	if response.Name != "João Silva" {
		t.Errorf(
			"expected name %q, got %q",
			"João Silva",
			response.Name,
		)
	}

	if response.Status != domain.LeadStatusNew {
		t.Errorf(
			"expected status %q, got %q",
			domain.LeadStatusNew,
			response.Status,
		)
	}
}

func TestLeadHandler_Create_InvalidJSON(t *testing.T) {
	repository := &fakeLeadRepository{}

	handler := newTestHandler(repository)

	body := `{
		"name": "João Silva",
		"email":
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/leads",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestLeadHandler_Create_InvalidLead(t *testing.T) {
	repository := &fakeLeadRepository{}

	handler := newTestHandler(repository)

	body := `{
		"name": "",
		"email": "joao@example.com",
		"phone": "11999999999"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/leads",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestLeadHandler_GetByID(t *testing.T) {
	leadID := uuid.New()

	repository := &fakeLeadRepository{
		getByIDFunc: func(
			_ context.Context,
			id uuid.UUID,
		) (*domain.Lead, error) {
			return &domain.Lead{
				ID:     id,
				Name:   "João Silva",
				Email:  "joao@example.com",
				Phone:  "11999999999",
				Status: domain.LeadStatusNew,
			}, nil
		},
	}

	handler := newTestHandler(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/leads/"+leadID.String(),
		nil,
	)

	request.SetPathValue("id", leadID.String())

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var response domain.Lead

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if response.ID != leadID {
		t.Errorf(
			"expected ID %s, got %s",
			leadID,
			response.ID,
		)
	}
}

func TestLeadHandler_GetByID_InvalidID(t *testing.T) {
	repository := &fakeLeadRepository{}

	handler := newTestHandler(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/leads/invalid-id",
		nil,
	)

	request.SetPathValue("id", "invalid-id")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestLeadHandler_List(t *testing.T) {
	leads := []*domain.Lead{
		{
			ID:     uuid.New(),
			Name:   "João Silva",
			Email:  "joao@example.com",
			Phone:  "11999999999",
			Status: domain.LeadStatusNew,
		},
		{
			ID:     uuid.New(),
			Name:   "Maria Silva",
			Email:  "maria@example.com",
			Phone:  "11888888888",
			Status: domain.LeadStatusContacted,
		},
	}

	repository := &fakeLeadRepository{
		listFunc: func(
			_ context.Context,
		) ([]*domain.Lead, error) {
			return leads, nil
		},
	}

	handler := newTestHandler(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/leads",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var response []*domain.Lead

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if len(response) != 2 {
		t.Fatalf(
			"expected 2 leads, got %d",
			len(response),
		)
	}
}

func TestLeadHandler_Update(t *testing.T) {
	leadID := uuid.New()

	repository := &fakeLeadRepository{
		updateFunc: func(
			_ context.Context,
			lead *domain.Lead,
		) (*domain.Lead, error) {
			return lead, nil
		},
	}

	handler := newTestHandler(repository)

	body := `{
		"name": "João Silva Atualizado",
		"email": "joao.updated@example.com",
		"phone": "11977776666",
		"status": "CONTACTED"
	}`

	request := httptest.NewRequest(
		http.MethodPut,
		"/leads/"+leadID.String(),
		strings.NewReader(body),
	)

	request.SetPathValue("id", leadID.String())

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var response domain.Lead

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if response.ID != leadID {
		t.Errorf(
			"expected ID %s, got %s",
			leadID,
			response.ID,
		)
	}

	if response.Name != "João Silva Atualizado" {
		t.Errorf(
			"expected name %q, got %q",
			"João Silva Atualizado",
			response.Name,
		)
	}

	if response.Status != domain.LeadStatusContacted {
		t.Errorf(
			"expected status %q, got %q",
			domain.LeadStatusContacted,
			response.Status,
		)
	}
}

func TestLeadHandler_Update_InvalidID(t *testing.T) {
	repository := &fakeLeadRepository{}

	handler := newTestHandler(repository)

	body := `{
		"name": "João Silva",
		"email": "joao@example.com",
		"phone": "11999999999",
		"status": "CONTACTED"
	}`

	request := httptest.NewRequest(
		http.MethodPut,
		"/leads/invalid-id",
		strings.NewReader(body),
	)

	request.SetPathValue("id", "invalid-id")

	recorder := httptest.NewRecorder()

	handler.Update(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestLeadHandler_UpdateStatus(t *testing.T) {
	leadID := uuid.New()

	repository := &fakeLeadRepository{
		updateStatusFunc: func(
			_ context.Context,
			id uuid.UUID,
			status domain.LeadStatus,
		) (*domain.Lead, error) {
			return &domain.Lead{
				ID:     id,
				Name:   "João Silva",
				Email:  "joao@example.com",
				Phone:  "11999999999",
				Status: status,
			}, nil
		},
	}

	handler := newTestHandler(repository)

	body := `{
		"status": "QUALIFIED"
	}`

	request := httptest.NewRequest(
		http.MethodPatch,
		"/leads/"+leadID.String()+"/status",
		strings.NewReader(body),
	)

	request.SetPathValue("id", leadID.String())

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var response domain.Lead

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if response.Status != domain.LeadStatusQualified {
		t.Errorf(
			"expected status %q, got %q",
			domain.LeadStatusQualified,
			response.Status,
		)
	}
}

func TestLeadHandler_UpdateStatus_InvalidID(t *testing.T) {
	repository := &fakeLeadRepository{}

	handler := newTestHandler(repository)

	body := `{
		"status": "QUALIFIED"
	}`

	request := httptest.NewRequest(
		http.MethodPatch,
		"/leads/invalid-id/status",
		strings.NewReader(body),
	)

	request.SetPathValue("id", "invalid-id")

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestLeadHandler_UpdateStatus_InvalidStatus(t *testing.T) {
	leadID := uuid.New()

	repository := &fakeLeadRepository{}

	handler := newTestHandler(repository)

	body := `{
		"status": ""
	}`

	request := httptest.NewRequest(
		http.MethodPatch,
		"/leads/"+leadID.String()+"/status",
		strings.NewReader(body),
	)

	request.SetPathValue("id", leadID.String())

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
