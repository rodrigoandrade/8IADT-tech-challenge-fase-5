package repository

import (
	"context"

	"github.com/google/uuid"

	"real-estate-sdr/internal/domain"
)

type PropertyRepository interface {
	Create(
		ctx context.Context,
		property *domain.Property,
	) (*domain.Property, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.Property, error)

	List(
		ctx context.Context,
		limit int,
		offset int,
	) ([]*domain.Property, error)

	Update(
		ctx context.Context,
		property *domain.Property,
	) (*domain.Property, error)

	UpdateStatus(
		ctx context.Context,
		id uuid.UUID,
		status domain.PropertyStatus,
	) (*domain.Property, error)

	Count(
		ctx context.Context,
	) (int64, error)
}
