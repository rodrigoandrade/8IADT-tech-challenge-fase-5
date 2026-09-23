package application

import (
	"context"
	"errors"

	"real-estate-sdr/internal/domain"
	"real-estate-sdr/internal/repository"
)

type MessageService struct {
	messageRepository      repository.MessageRepository
	conversationRepository repository.ConversationRepository
	aiService              AIService
}

func NewMessageService(
	messageRepository repository.MessageRepository,
	conversationRepository repository.ConversationRepository,
	aiService AIService,
) *MessageService {
	return &MessageService{
		messageRepository:      messageRepository,
		conversationRepository: conversationRepository,
		aiService:              aiService,
	}
}

// Create cria uma mensagem normalmente, sem chamar o serviço de IA.
func (s *MessageService) Create(
	ctx context.Context,
	message *domain.Message,
) (*domain.Message, error) {

	if message == nil {
		return nil, errors.New("message is required")
	}

	if err := message.Validate(); err != nil {
		return nil, err
	}

	if message.ID == [16]byte{} {
		// Não faz nada aqui caso o ID seja gerado no repository.
		// Mantemos a responsabilidade de geração conforme a implementação atual.
	}

	return s.messageRepository.Create(ctx, message)
}

// CreateAndGenerateResponse cria a mensagem do lead,
// recupera o histórico da conversa, chama o serviço de IA
// e persiste a resposta do assistente.
func (s *MessageService) CreateAndGenerateResponse(
	ctx context.Context,
	message *domain.Message,
) (*domain.AIResponse, error) {

	if message == nil {
		return nil, errors.New("message is required")
	}

	if err := message.Validate(); err != nil {
		return nil, err
	}

	// 1. Persiste a mensagem enviada pelo lead.
	createdMessage, err := s.messageRepository.Create(ctx, message)
	if err != nil {
		return nil, err
	}

	if createdMessage == nil {
		return nil, errors.New("message repository returned an empty message")
	}

	// 2. Recupera a conversa.
	conversation, err := s.conversationRepository.GetByID(
		ctx,
		message.ConversationID,
	)
	if err != nil {
		return nil, err
	}

	if conversation == nil {
		return nil, errors.New("conversation not found")
	}

	// 3. Recupera todo o histórico da conversa.
	messages, err := s.messageRepository.ListByConversation(
		ctx,
		message.ConversationID,
	)
	if err != nil {
		return nil, err
	}

	// 4. Envia conversa + histórico para o AI Service.
	aiResponse, err := s.aiService.GenerateResponse(
		ctx,
		conversation,
		messages,
	)
	if err != nil {
		return nil, err
	}

	if aiResponse == nil {
		return nil, errors.New("AI service returned an empty response")
	}

	// 5. Persiste a resposta do assistente.
	assistantMessage := &domain.Message{
		ConversationID: message.ConversationID,
		Role:           domain.MessageRoleAssistant,
		Content:        aiResponse.Message,
	}

	if _, err := s.messageRepository.Create(
		ctx,
		assistantMessage,
	); err != nil {
		return nil, err
	}

	return aiResponse, nil
}

func (s *MessageService) GetByID(
	ctx context.Context,
	id [16]byte,
) (*domain.Message, error) {
	return s.messageRepository.GetByID(ctx, id)
}

func (s *MessageService) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.Message, error) {
	return s.messageRepository.List(ctx, limit, offset)
}

func (s *MessageService) ListByConversation(
	ctx context.Context,
	conversationID [16]byte,
) ([]*domain.Message, error) {
	return s.messageRepository.ListByConversation(
		ctx,
		conversationID,
	)
}

func (s *MessageService) Count(
	ctx context.Context,
) (int64, error) {
	return s.messageRepository.Count(ctx)
}

func (s *MessageService) CountByConversation(
	ctx context.Context,
	conversationID [16]byte,
) (int64, error) {
	return s.messageRepository.CountByConversation(
		ctx,
		conversationID,
	)
}
