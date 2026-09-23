package repository

import (
	"context"

	"github.com/google/uuid"

	"real-estate-sdr/internal/domain"
)

type ConversationRepository interface {
	Create(
		ctx context.Context,
		conversation *domain.Conversation,
	) (*domain.Conversation, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.Conversation, error)

	GetByLeadID(
		ctx context.Context,
		leadID uuid.UUID,
	) ([]*domain.Conversation, error)

	List(
		ctx context.Context,
		limit int,
		offset int,
	) ([]*domain.Conversation, error)

	Close(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.Conversation, error)

	Count(
		ctx context.Context,
	) (int64, error)
}
