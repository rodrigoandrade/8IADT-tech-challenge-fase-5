package postgres

import (
	"context"
	"errors"
	"fmt"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"real-estate-sdr/internal/domain"
	"real-estate-sdr/internal/repository"
)

type LeadRepository struct {
	queries *generated.Queries
}

func NewLeadRepository(queries *generated.Queries) *LeadRepository {
	return &LeadRepository{
		queries: queries,
	}
}

func (r *LeadRepository) Create(
	ctx context.Context,
	lead *domain.Lead,
) (*domain.Lead, error) {
	if lead == nil {
		return nil, errors.New("lead cannot be nil")
	}

	result, err := r.queries.CreateLead(
		ctx,
		generated.CreateLeadParams{
			Name:  lead.Name,
			Email: lead.Email,
			Phone: lead.Phone,
		},
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "uq_leads_email" {
				return nil, repository.ErrLeadAlreadyExists
			}
		}

		return nil, fmt.Errorf(
			"create lead: %w",
			err,
		)
	}

	return mapGeneratedLeadToDomain(result), nil
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
			"get lead by id: %w",
			err,
		)
	}

	return mapGeneratedLeadToDomain(result), nil
}

func (r *LeadRepository) List(
	ctx context.Context,
) ([]*domain.Lead, error) {
	results, err := r.queries.ListLeads(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"list leads: %w",
			err,
		)
	}

	leads := make([]*domain.Lead, 0, len(results))

	for _, result := range results {
		leads = append(
			leads,
			mapGeneratedLeadToDomain(result),
		)
	}

	return leads, nil
}

func (r *LeadRepository) Update(
	ctx context.Context,
	lead *domain.Lead,
) (*domain.Lead, error) {
	if lead == nil {
		return nil, errors.New("lead cannot be nil")
	}

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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrLeadNotFound
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "uq_leads_email" {
				return nil, repository.ErrLeadAlreadyExists
			}
		}

		return nil, fmt.Errorf(
			"update lead: %w",
			err,
		)
	}

	return mapGeneratedLeadToDomain(result), nil
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrLeadNotFound
		}

		return nil, fmt.Errorf(
			"update lead status: %w",
			err,
		)
	}

	return mapGeneratedLeadToDomain(result), nil
}

func (r *LeadRepository) Count(
	ctx context.Context,
) (int64, error) {
	count, err := r.queries.CountLeads(ctx)
	if err != nil {
		return 0, fmt.Errorf(
			"count leads: %w",
			err,
		)
	}

	return count, nil
}

func mapGeneratedLeadToDomain(
	result generated.Lead,
) *domain.Lead {
	return &domain.Lead{
		ID:        result.ID,
		Name:      result.Name,
		Email:     result.Email,
		Phone:     result.Phone,
		Status:    domain.LeadStatus(result.Status),
		CreatedAt: timestamptzToTime(result.CreatedAt),
		UpdatedAt: timestamptzToTime(result.UpdatedAt),
	}
}

func timestamptzToTime(
	value pgtype.Timestamptz,
) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time
}
