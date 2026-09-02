package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"real-estate-sdr/internal/domain"
	"real-estate-sdr/internal/repository"
)

var (
	ErrConversationNotFound      = errors.New("conversation not found")
	ErrConversationAlreadyClosed = errors.New("conversation already closed")
)

type ConversationService struct {
	repository repository.ConversationRepository
}

func NewConversationService(
	repository repository.ConversationRepository,
) *ConversationService {
	return &ConversationService{
		repository: repository,
	}
}

func (s *ConversationService) Create(
	ctx context.Context,
	conversation *domain.Conversation,
) (*domain.Conversation, error) {
	if conversation == nil {
		return nil, domain.ErrInvalidConversation
	}

	if err := conversation.Validate(); err != nil {
		return nil, fmt.Errorf(
			"invalid conversation: %w",
			err,
		)
	}

	created, err := s.repository.Create(
		ctx,
		conversation,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create conversation: %w",
			err,
		)
	}

	return created, nil
}

func (s *ConversationService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Conversation, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf(
			"invalid conversation id",
		)
	}

	conversation, err := s.repository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConversationNotFound
		}

		return nil, fmt.Errorf(
			"get conversation: %w",
			err,
		)
	}

	return conversation, nil
}

func (s *ConversationService) GetByLeadID(
	ctx context.Context,
	leadID uuid.UUID,
) ([]*domain.Conversation, error) {
	if leadID == uuid.Nil {
		return nil, fmt.Errorf(
			"invalid lead id",
		)
	}

	conversations, err := s.repository.GetByLeadID(
		ctx,
		leadID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get conversations by lead: %w",
			err,
		)
	}

	return conversations, nil
}

func (s *ConversationService) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.Conversation, error) {
	if limit <= 0 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	conversations, err := s.repository.List(
		ctx,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list conversations: %w",
			err,
		)
	}

	return conversations, nil
}

func (s *ConversationService) Close(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Conversation, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf(
			"invalid conversation id",
		)
	}

	// First verify that the conversation exists.
	conversation, err := s.repository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConversationNotFound
		}

		return nil, fmt.Errorf(
			"get conversation before close: %w",
			err,
		)
	}

	// The conversation exists, but is already closed.
	if conversation.Status == domain.ConversationStatusClosed {
		return nil, ErrConversationAlreadyClosed
	}

	// Close the active conversation.
	closedConversation, err := s.repository.Close(
		ctx,
		id,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// This can happen if another request closed the
			// conversation between GetByID and Close.
			return nil, ErrConversationAlreadyClosed
		}

		return nil, fmt.Errorf(
			"close conversation: %w",
			err,
		)
	}

	return closedConversation, nil
}

func (s *ConversationService) Count(
	ctx context.Context,
) (int64, error) {
	count, err := s.repository.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf(
			"count conversations: %w",
			err,
		)
	}

	return count, nil
}
