package application

import (
	"context"
	"errors"
	"real-estate-sdr/internal/repository"
	"testing"
	"time"

	"github.com/google/uuid"

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

func TestLeadService_Create(t *testing.T) {
	ctx := context.Background()

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

	service := NewLeadService(repository)

	lead := &domain.Lead{
		Name:  "João Silva",
		Email: "joao@example.com",
		Phone: "11999999999",
	}

	createdLead, err := service.Create(ctx, lead)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if createdLead == nil {
		t.Fatal("expected created lead, got nil")
	}

	if createdLead.ID != leadID {
		t.Errorf(
			"expected ID %s, got %s",
			leadID,
			createdLead.ID,
		)
	}

	if createdLead.Status != domain.LeadStatusNew {
		t.Errorf(
			"expected status %q, got %q",
			domain.LeadStatusNew,
			createdLead.Status,
		)
	}
}

func TestLeadService_Create_InvalidLead(t *testing.T) {
	ctx := context.Background()

	repositoryCalled := false

	repository := &fakeLeadRepository{
		createFunc: func(
			_ context.Context,
			lead *domain.Lead,
		) (*domain.Lead, error) {
			repositoryCalled = true
			return lead, nil
		},
	}

	service := NewLeadService(repository)

	tests := []struct {
		name string
		lead *domain.Lead
	}{
		{
			name: "nil lead",
			lead: nil,
		},
		{
			name: "missing name",
			lead: &domain.Lead{
				Email: "joao@example.com",
				Phone: "11999999999",
			},
		},
		{
			name: "missing email",
			lead: &domain.Lead{
				Name:  "João Silva",
				Phone: "11999999999",
			},
		},
		{
			name: "missing phone",
			lead: &domain.Lead{
				Name:  "João Silva",
				Email: "joao@example.com",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repositoryCalled = false

			_, err := service.Create(ctx, test.lead)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.Is(err, ErrInvalidLead) {
				t.Errorf(
					"expected ErrInvalidLead, got %v",
					err,
				)
			}

			if repositoryCalled {
				t.Error("repository should not have been called")
			}
		})
	}
}

func TestLeadService_GetByID(t *testing.T) {
	ctx := context.Background()
	leadID := uuid.New()

	expectedLead := &domain.Lead{
		ID:     leadID,
		Name:   "João Silva",
		Email:  "joao@example.com",
		Phone:  "11999999999",
		Status: domain.LeadStatusNew,
	}

	repository := &fakeLeadRepository{
		getByIDFunc: func(
			_ context.Context,
			id uuid.UUID,
		) (*domain.Lead, error) {
			if id != leadID {
				t.Errorf(
					"expected ID %s, got %s",
					leadID,
					id,
				)
			}

			return expectedLead, nil
		},
	}

	service := NewLeadService(repository)

	result, err := service.GetByID(ctx, leadID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected lead, got nil")
	}

	if result.ID != expectedLead.ID {
		t.Errorf(
			"expected ID %s, got %s",
			expectedLead.ID,
			result.ID,
		)
	}
}

func TestLeadService_GetByID_InvalidID(t *testing.T) {
	ctx := context.Background()

	repositoryCalled := false

	repository := &fakeLeadRepository{
		getByIDFunc: func(
			_ context.Context,
			_ uuid.UUID,
		) (*domain.Lead, error) {
			repositoryCalled = true
			return nil, nil
		},
	}

	service := NewLeadService(repository)

	_, err := service.GetByID(ctx, uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidLead) {
		t.Errorf(
			"expected ErrInvalidLead, got %v",
			err,
		)
	}

	if repositoryCalled {
		t.Error("repository should not have been called")
	}
}

func TestLeadService_List(t *testing.T) {
	ctx := context.Background()

	expectedLeads := []*domain.Lead{
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
			return expectedLeads, nil
		},
	}

	service := NewLeadService(repository)

	result, err := service.List(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != len(expectedLeads) {
		t.Fatalf(
			"expected %d leads, got %d",
			len(expectedLeads),
			len(result),
		)
	}
}

func TestLeadService_Update(t *testing.T) {
	ctx := context.Background()

	leadID := uuid.New()

	lead := &domain.Lead{
		ID:     leadID,
		Name:   "João Silva",
		Email:  "joao@example.com",
		Phone:  "11999999999",
		Status: domain.LeadStatusNew,
	}

	repository := &fakeLeadRepository{
		updateFunc: func(
			_ context.Context,
			input *domain.Lead,
		) (*domain.Lead, error) {
			return input, nil
		},
	}

	service := NewLeadService(repository)

	result, err := service.Update(ctx, lead)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ID != leadID {
		t.Errorf(
			"expected ID %s, got %s",
			leadID,
			result.ID,
		)
	}
}

func TestLeadService_UpdateStatus(t *testing.T) {
	ctx := context.Background()

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

	service := NewLeadService(repository)

	result, err := service.UpdateStatus(
		ctx,
		leadID,
		domain.LeadStatusQualified,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Status != domain.LeadStatusQualified {
		t.Errorf(
			"expected status %q, got %q",
			domain.LeadStatusQualified,
			result.Status,
		)
	}
}

func TestLeadService_UpdateStatus_InvalidInput(t *testing.T) {
	ctx := context.Background()

	repositoryCalled := false

	repository := &fakeLeadRepository{
		updateStatusFunc: func(
			_ context.Context,
			_ uuid.UUID,
			_ domain.LeadStatus,
		) (*domain.Lead, error) {
			repositoryCalled = true
			return nil, nil
		},
	}

	service := NewLeadService(repository)

	tests := []struct {
		name   string
		id     uuid.UUID
		status domain.LeadStatus
	}{
		{
			name:   "invalid id",
			id:     uuid.Nil,
			status: domain.LeadStatusQualified,
		},
		{
			name:   "invalid status",
			id:     uuid.New(),
			status: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repositoryCalled = false

			_, err := service.UpdateStatus(
				ctx,
				test.id,
				test.status,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.Is(err, ErrInvalidLead) {
				t.Errorf(
					"expected ErrInvalidLead, got %v",
					err,
				)
			}

			if repositoryCalled {
				t.Error("repository should not have been called")
			}
		})
	}
}

func TestLeadService_Count(t *testing.T) {
	ctx := context.Background()

	repository := &fakeLeadRepository{
		countFunc: func(
			_ context.Context,
		) (int64, error) {
			return 10, nil
		},
	}

	service := NewLeadService(repository)

	count, err := service.Count(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 10 {
		t.Errorf(
			"expected count 10, got %d",
			count,
		)
	}

}

func TestLeadService_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()

	leadID := uuid.New()

	repository := &fakeLeadRepository{
		getByIDFunc: func(
			_ context.Context,
			_ uuid.UUID,
		) (*domain.Lead, error) {
			return nil, repository.ErrLeadNotFound
		},
	}

	service := NewLeadService(repository)

	_, err := service.GetByID(ctx, leadID)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrLeadNotFound) {
		t.Errorf(
			"expected ErrLeadNotFound, got %v",
			err,
		)
	}
}
