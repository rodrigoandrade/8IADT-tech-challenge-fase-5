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

type MessageRepository struct {
	queries *generated.Queries
}

func NewMessageRepository(
	queries *generated.Queries,
) *MessageRepository {
	return &MessageRepository{
		queries: queries,
	}
}

func (r *MessageRepository) Create(
	ctx context.Context,
	message *domain.Message,
) (*domain.Message, error) {
	if message == nil {
		return nil, errors.New("message is required")
	}

	result, err := r.queries.CreateMessage(
		ctx,
		generated.CreateMessageParams{
			ConversationID: message.ConversationID,
			Role:           message.Role,
			Content:        message.Content,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create message: %w",
			err,
		)
	}

	return mapGeneratedMessage(result), nil
}

func (r *MessageRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Message, error) {
	result, err := r.queries.GetMessageByID(
		ctx,
		id,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrMessageNotFound
		}

		return nil, fmt.Errorf(
			"get message by id: %w",
			err,
		)
	}

	return mapGeneratedMessage(result), nil
}

func (r *MessageRepository) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.Message, error) {
	results, err := r.queries.ListMessages(
		ctx,
		generated.ListMessagesParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list messages: %w",
			err,
		)
	}

	messages := make(
		[]*domain.Message,
		0,
		len(results),
	)

	for _, result := range results {
		messages = append(
			messages,
			mapGeneratedMessage(result),
		)
	}

	return messages, nil
}

func (r *MessageRepository) ListByConversation(
	ctx context.Context,
	conversationID uuid.UUID,
) ([]*domain.Message, error) {
	results, err := r.queries.ListMessagesByConversation(
		ctx,
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list messages by conversation: %w",
			err,
		)
	}

	messages := make(
		[]*domain.Message,
		0,
		len(results),
	)

	for _, result := range results {
		messages = append(
			messages,
			mapGeneratedMessage(result),
		)
	}

	return messages, nil
}

func (r *MessageRepository) Count(
	ctx context.Context,
) (int64, error) {
	count, err := r.queries.CountMessages(ctx)
	if err != nil {
		return 0, fmt.Errorf(
			"count messages: %w",
			err,
		)
	}

	return count, nil
}

func (r *MessageRepository) CountByConversation(
	ctx context.Context,
	conversationID uuid.UUID,
) (int64, error) {
	count, err := r.queries.CountMessagesByConversation(
		ctx,
		conversationID,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"count messages by conversation: %w",
			err,
		)
	}

	return count, nil
}

func mapGeneratedMessage(
	message generated.Message,
) *domain.Message {
	return &domain.Message{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Role:           message.Role,
		Content:        message.Content,
		CreatedAt:      message.CreatedAt.Time,
	}
}
