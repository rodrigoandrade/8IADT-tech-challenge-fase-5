package postgres

import (
	"context"
	"errors"
	"fmt"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
	"real-estate-sdr/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"real-estate-sdr/internal/domain"
)

type ConversationRepository struct {
	queries *generated.Queries
}

func NewConversationRepository(
	queries *generated.Queries,
) *ConversationRepository {
	return &ConversationRepository{
		queries: queries,
	}
}

func (r *ConversationRepository) Create(
	ctx context.Context,
	conversation *domain.Conversation,
) (*domain.Conversation, error) {
	if conversation == nil {
		return nil, domain.ErrInvalidConversation
	}

	if err := conversation.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate conversation: %w",
			err,
		)
	}

	result, err := r.queries.CreateConversation(
		ctx,
		generated.CreateConversationParams{
			LeadID: conversation.LeadID,
			Status: string(conversation.Status),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create conversation: %w",
			err,
		)
	}

	return toDomainConversation(result), nil
}

func (r *ConversationRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Conversation, error) {
	result, err := r.queries.GetConversationByID(
		ctx,
		id,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"conversation not found: %w",
				err,
			)
		}

		return nil, fmt.Errorf(
			"get conversation by id: %w",
			err,
		)
	}

	return toDomainConversation(result), nil
}

func (r *ConversationRepository) GetByLeadID(
	ctx context.Context,
	leadID uuid.UUID,
) ([]*domain.Conversation, error) {
	results, err := r.queries.ListConversationsByLead(
		ctx,
		leadID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list conversations by lead: %w",
			err,
		)
	}

	conversations := make(
		[]*domain.Conversation,
		0,
		len(results),
	)

	for _, result := range results {
		conversations = append(
			conversations,
			toDomainConversation(result),
		)
	}

	return conversations, nil
}

func (r *ConversationRepository) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.Conversation, error) {
	results, err := r.queries.ListConversations(
		ctx,
		generated.ListConversationsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list conversations: %w",
			err,
		)
	}

	conversations := make(
		[]*domain.Conversation,
		0,
		len(results),
	)

	for _, result := range results {
		conversations = append(
			conversations,
			toDomainConversation(result),
		)
	}

	return conversations, nil
}

func (r *ConversationRepository) Close(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Conversation, error) {
	result, err := r.queries.CloseConversation(
		ctx,
		id,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"conversation not found or already closed: %w",
				err,
			)
		}

		return nil, fmt.Errorf(
			"close conversation: %w",
			err,
		)
	}

	return toDomainConversation(result), nil
}

func (r *ConversationRepository) Count(
	ctx context.Context,
) (int64, error) {
	count, err := r.queries.CountConversations(ctx)
	if err != nil {
		return 0, fmt.Errorf(
			"count conversations: %w",
			err,
		)
	}

	return count, nil
}

func toDomainConversation(
	conversation generated.Conversation,
) *domain.Conversation {
	result := &domain.Conversation{
		ID:     conversation.ID,
		LeadID: conversation.LeadID,
		Status: domain.ConversationStatus(
			conversation.Status,
		),
	}

	if conversation.StartedAt.Valid {
		result.StartedAt = conversation.StartedAt.Time
	}

	if conversation.EndedAt.Valid {
		endedAt := conversation.EndedAt.Time
		result.EndedAt = &endedAt
	}

	return result
}

var _ repository.ConversationRepository = (*ConversationRepository)(nil)
