package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"real-estate-sdr/internal/domain"
	"real-estate-sdr/internal/repository"
)

var (
	ErrInvalidLead  = errors.New("invalid lead")
	ErrLeadNotFound = errors.New("lead not found")
)

type LeadService struct {
	repository repository.LeadRepository
}

func NewLeadService(repository repository.LeadRepository) *LeadService {
	return &LeadService{
		repository: repository,
	}
}

func (s *LeadService) Create(
	ctx context.Context,
	lead *domain.Lead,
) (*domain.Lead, error) {
	if err := validateLead(lead); err != nil {
		return nil, err
	}

	if lead.Status == "" {
		lead.Status = domain.LeadStatusNew
	}

	createdLead, err := s.repository.Create(ctx, lead)
	if err != nil {
		return nil, fmt.Errorf("create lead: %w", err)
	}

	return createdLead, nil
}

func (s *LeadService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Lead, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf(
			"%w: id is required",
			ErrInvalidLead,
		)
	}

	lead, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrLeadNotFound) {
			return nil, ErrLeadNotFound
		}

		return nil, fmt.Errorf("get lead: %w", err)
	}

	return lead, nil
}

func (s *LeadService) List(
	ctx context.Context,
) ([]*domain.Lead, error) {
	leads, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list leads: %w", err)
	}

	return leads, nil
}

func (s *LeadService) Update(
	ctx context.Context,
	lead *domain.Lead,
) (*domain.Lead, error) {
	if err := validateLead(lead); err != nil {
		return nil, err
	}

	if lead.ID == uuid.Nil {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidLead)
	}

	updatedLead, err := s.repository.Update(ctx, lead)
	if err != nil {
		return nil, fmt.Errorf("update lead: %w", err)
	}

	return updatedLead, nil
}

func (s *LeadService) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.LeadStatus,
) (*domain.Lead, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidLead)
	}

	if status == "" {
		return nil, fmt.Errorf("%w: status is required", ErrInvalidLead)
	}

	lead, err := s.repository.UpdateStatus(ctx, id, status)
	if err != nil {
		return nil, fmt.Errorf("update lead status: %w", err)
	}

	return lead, nil
}

func (s *LeadService) Count(
	ctx context.Context,
) (int64, error) {
	count, err := s.repository.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count leads: %w", err)
	}

	return count, nil
}

func validateLead(lead *domain.Lead) error {
	if lead == nil {
		return fmt.Errorf("%w: lead is required", ErrInvalidLead)
	}

	if strings.TrimSpace(lead.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidLead)
	}

	if strings.TrimSpace(lead.Email) == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidLead)
	}

	if strings.TrimSpace(lead.Phone) == "" {
		return fmt.Errorf("%w: phone is required", ErrInvalidLead)
	}

	return nil
}
