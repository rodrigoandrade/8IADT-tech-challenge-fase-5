package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"real-estate-sdr/internal/domain"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
	"real-estate-sdr/internal/repository"
)

type LeadRepository struct {
	queries *generated.Queries
}

func NewLeadRepository(queries *generated.Queries) repository.LeadRepository {
	return &LeadRepository{
		queries: queries,
	}
}

func (r *LeadRepository) Create(
	ctx context.Context,
	lead *domain.Lead,
) (*domain.Lead, error) {
	result, err := r.queries.CreateLead(ctx, generated.CreateLeadParams{
		Name:  lead.Name,
		Email: lead.Email,
		Phone: lead.Phone,
	})
	if err != nil {
		return nil, fmt.Errorf("create lead: %w", err)
	}

	return toDomainLead(result), nil
}

func (r *LeadRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Lead, error) {
	result, err := r.queries.GetLeadByID(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrLeadNotFound
		}

		return nil, fmt.Errorf(
			"get lead by id %s: %w",
			id,
			err,
		)
	}

	return toDomainLead(result), nil
}

func (r *LeadRepository) List(
	ctx context.Context,
) ([]*domain.Lead, error) {
	results, err := r.queries.ListLeads(ctx)
	if err != nil {
		return nil, fmt.Errorf("list leads: %w", err)
	}

	leads := make([]*domain.Lead, 0, len(results))

	for _, result := range results {
		leads = append(leads, toDomainLead(result))
	}

	return leads, nil
}

func (r *LeadRepository) Update(
	ctx context.Context,
	lead *domain.Lead,
) (*domain.Lead, error) {
	result, err := r.queries.UpdateLead(
		ctx,
		generated.UpdateLeadParams{
			ID:     lead.ID,
			Name:   lead.Name,
			Email:  lead.Email,
			Phone:  lead.Phone,
			Status: string(lead.Status),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"update lead %s: %w",
			lead.ID,
			err,
		)
	}

	return toDomainLead(result), nil
}

func (r *LeadRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.LeadStatus,
) (*domain.Lead, error) {
	result, err := r.queries.UpdateLeadStatus(
		ctx,
		generated.UpdateLeadStatusParams{
			ID:     id,
			Status: string(status),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"update lead status %s: %w",
			id,
			err,
		)
	}

	return toDomainLead(result), nil
}

func (r *LeadRepository) Count(
	ctx context.Context,
) (int64, error) {
	count, err := r.queries.CountLeads(ctx)
	if err != nil {
		return 0, fmt.Errorf("count leads: %w", err)
	}

	return count, nil
}

func toDomainLead(lead generated.Lead) *domain.Lead {
	return &domain.Lead{
		ID:        lead.ID,
		Name:      lead.Name,
		Email:     lead.Email,
		Phone:     lead.Phone,
		Status:    domain.LeadStatus(lead.Status),
		CreatedAt: lead.CreatedAt.Time,
		UpdatedAt: lead.UpdatedAt.Time,
	}
}
