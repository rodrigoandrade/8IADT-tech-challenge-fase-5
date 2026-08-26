package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"real-estate-sdr/internal/domain"
	"real-estate-sdr/internal/repository"
)

type PropertyService struct {
	repository repository.PropertyRepository
}

func NewPropertyService(
	repository repository.PropertyRepository,
) *PropertyService {
	return &PropertyService{
		repository: repository,
	}
}

func (s *PropertyService) Create(
	ctx context.Context,
	property *domain.Property,
) (*domain.Property, error) {
	if property == nil {
		return nil, fmt.Errorf(
			"%w: property is required",
			domain.ErrInvalidProperty,
		)
	}

	if err := property.Validate(); err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			domain.ErrInvalidProperty,
			err,
		)
	}

	createdProperty, err := s.repository.Create(
		ctx,
		property,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create property: %w",
			err,
		)
	}

	return createdProperty, nil
}

func (s *PropertyService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Property, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf(
			"%w: id is required",
			domain.ErrInvalidProperty,
		)
	}

	property, err := s.repository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get property by id: %w",
			err,
		)
	}

	return property, nil
}

func (s *PropertyService) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.Property, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	properties, err := s.repository.List(
		ctx,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list properties: %w",
			err,
		)
	}

	return properties, nil
}

func (s *PropertyService) Update(
	ctx context.Context,
	property *domain.Property,
) (*domain.Property, error) {
	if property == nil {
		return nil, fmt.Errorf(
			"%w: property is required",
			domain.ErrInvalidProperty,
		)
	}

	if property.ID == uuid.Nil {
		return nil, fmt.Errorf(
			"%w: id is required",
			domain.ErrInvalidProperty,
		)
	}

	if err := property.Validate(); err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			domain.ErrInvalidProperty,
			err,
		)
	}

	updatedProperty, err := s.repository.Update(
		ctx,
		property,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"update property: %w",
			err,
		)
	}

	return updatedProperty, nil
}

func (s *PropertyService) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.PropertyStatus,
) (*domain.Property, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf(
			"%w: id is required",
			domain.ErrInvalidProperty,
		)
	}

	if !domain.IsValidPropertyStatus(status) {
		return nil, fmt.Errorf(
			"%w: invalid status",
			domain.ErrInvalidPropertyStatus,
		)
	}

	updatedProperty, err := s.repository.UpdateStatus(
		ctx,
		id,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"update property status: %w",
			err,
		)
	}

	return updatedProperty, nil
}

func (s *PropertyService) Count(
	ctx context.Context,
) (int64, error) {
	count, err := s.repository.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf(
			"count properties: %w",
			err,
		)
	}

	return count, nil
}
