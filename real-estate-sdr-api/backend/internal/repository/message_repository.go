package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"real-estate-sdr/internal/domain"
)

var ErrMessageNotFound = errors.New("message not found")

type MessageRepository interface {
	Create(
		ctx context.Context,
		message *domain.Message,
	) (*domain.Message, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.Message, error)

	List(
		ctx context.Context,
		limit int,
		offset int,
	) ([]*domain.Message, error)

	ListByConversation(
		ctx context.Context,
		conversationID uuid.UUID,
	) ([]*domain.Message, error)

	Count(
		ctx context.Context,
	) (int64, error)

	CountByConversation(
		ctx context.Context,
		conversationID uuid.UUID,
	) (int64, error)
}
