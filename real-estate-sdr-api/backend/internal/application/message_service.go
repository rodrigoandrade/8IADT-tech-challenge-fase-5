package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"real-estate-sdr/internal/domain"
	"real-estate-sdr/internal/repository"
)

type MessageService struct {
	messageRepository repository.MessageRepository
}

func NewMessageService(
	messageRepository repository.MessageRepository,
) *MessageService {
	return &MessageService{
		messageRepository: messageRepository,
	}
}

func (s *MessageService) Create(
	ctx context.Context,
	message *domain.Message,
) (*domain.Message, error) {
	if message == nil {
		return nil, fmt.Errorf("message is required")
	}

	if err := message.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	createdMessage, err := s.messageRepository.Create(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	return createdMessage, nil
}

func (s *MessageService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Message, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("message id is required")
	}

	message, err := s.messageRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get message by id: %w", err)
	}

	return message, nil
}

func (s *MessageService) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.Message, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}

	messages, err := s.messageRepository.List(
		ctx,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}

	return messages, nil
}

func (s *MessageService) ListByConversation(
	ctx context.Context,
	conversationID uuid.UUID,
) ([]*domain.Message, error) {
	if conversationID == uuid.Nil {
		return nil, fmt.Errorf("conversation id is required")
	}

	messages, err := s.messageRepository.ListByConversation(
		ctx,
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list messages by conversation: %w",
			err,
		)
	}

	return messages, nil
}

func (s *MessageService) Count(
	ctx context.Context,
) (int64, error) {
	count, err := s.messageRepository.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count messages: %w", err)
	}

	return count, nil
}

func (s *MessageService) CountByConversation(
	ctx context.Context,
	conversationID uuid.UUID,
) (int64, error) {
	if conversationID == uuid.Nil {
		return 0, fmt.Errorf("conversation id is required")
	}

	count, err := s.messageRepository.CountByConversation(
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
