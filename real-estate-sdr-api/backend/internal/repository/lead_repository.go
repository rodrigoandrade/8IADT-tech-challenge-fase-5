package repository

import (
	"context"

	"github.com/google/uuid"

	"real-estate-sdr/internal/domain"
)

//var ErrLeadNotFound = errors.New("lead not found")

type LeadRepository interface {
	Create(ctx context.Context, lead *domain.Lead) (*domain.Lead, error)

	GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error)

	List(ctx context.Context) ([]*domain.Lead, error)

	Update(ctx context.Context, lead *domain.Lead) (*domain.Lead, error)

	UpdateStatus(
		ctx context.Context,
		id uuid.UUID,
		status domain.LeadStatus,
	) (*domain.Lead, error)

	Count(ctx context.Context) (int64, error)
}
