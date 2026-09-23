package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/domain"
	"real-estate-sdr/internal/repository"
)

// ============================================================
// MESSAGE REPOSITORY MOCK
// ============================================================

type testMessageRepository struct {
	repository.MessageRepository

	createFn              func(context.Context, *domain.Message) (*domain.Message, error)
	getByIDFn             func(context.Context, uuid.UUID) (*domain.Message, error)
	listFn                func(context.Context, int, int) ([]*domain.Message, error)
	listByConversationFn  func(context.Context, uuid.UUID) ([]*domain.Message, error)
	countFn               func(context.Context) (int64, error)
	countByConversationFn func(context.Context, uuid.UUID) (int64, error)
}

func (m *testMessageRepository) Create(
	ctx context.Context,
	message *domain.Message,
) (*domain.Message, error) {
	if m.createFn != nil {
		return m.createFn(ctx, message)
	}

	return nil, errors.New("create not configured")
}

func (m *testMessageRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Message, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return nil, errors.New("get by id not configured")
}

func (m *testMessageRepository) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.Message, error) {
	if m.listFn != nil {
		return m.listFn(ctx, limit, offset)
	}

	return nil, errors.New("list not configured")
}

func (m *testMessageRepository) ListByConversation(
	ctx context.Context,
	conversationID uuid.UUID,
) ([]*domain.Message, error) {
	if m.listByConversationFn != nil {
		return m.listByConversationFn(ctx, conversationID)
	}

	return nil, errors.New("list by conversation not configured")
}

func (m *testMessageRepository) Count(
	ctx context.Context,
) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx)
	}

	return 0, errors.New("count not configured")
}

func (m *testMessageRepository) CountByConversation(
	ctx context.Context,
	conversationID uuid.UUID,
) (int64, error) {
	if m.countByConversationFn != nil {
		return m.countByConversationFn(ctx, conversationID)
	}

	return 0, errors.New("count by conversation not configured")
}

// ============================================================
// CONVERSATION REPOSITORY MOCK
// ============================================================

type testConversationRepository struct {
	repository.ConversationRepository

	getByIDFn func(
		context.Context,
		uuid.UUID,
	) (*domain.Conversation, error)
}

func (m *testConversationRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Conversation, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return nil, errors.New("get conversation by id not configured")
}

// ============================================================
// AI SERVICE MOCK
// ============================================================

type testAIService struct {
	generateResponseFn func(
		context.Context,
		*domain.Conversation,
		[]*domain.Message,
	) (*domain.AIResponse, error)
}

func (m *testAIService) GenerateResponse(
	ctx context.Context,
	conversation *domain.Conversation,
	messages []*domain.Message,
) (*domain.AIResponse, error) {
	if m.generateResponseFn != nil {
		return m.generateResponseFn(
			ctx,
			conversation,
			messages,
		)
	}

	return nil, errors.New("generate response not configured")
}

// ============================================================
// HELPERS
// ============================================================

func newExternalTestMessage() *domain.Message {
	return &domain.Message{
		ID:             uuid.New(),
		ConversationID: uuid.New(),
		Role:           domain.MessageRoleUser,
		Content:        "Estou procurando um apartamento na zona sul.",
	}
}

func newExternalTestConversation() *domain.Conversation {
	return &domain.Conversation{
		ID:     uuid.New(),
		LeadID: uuid.New(),
		Status: domain.ConversationStatusActive,
	}
}

func newExternalTestService(
	messageRepository repository.MessageRepository,
	conversationRepository repository.ConversationRepository,
	aiService application.AIService,
) *application.MessageService {
	return application.NewMessageService(
		messageRepository,
		conversationRepository,
		aiService,
	)
}

// ============================================================
// CREATE
// ============================================================

func TestExternalMessageServiceCreate(t *testing.T) {
	message := newExternalTestMessage()

	called := false

	messageRepository := &testMessageRepository{
		createFn: func(
			ctx context.Context,
			input *domain.Message,
		) (*domain.Message, error) {

			called = true

			if input != message {
				t.Fatal("expected same message")
			}

			return input, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.Create(
		context.Background(),
		message,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !called {
		t.Fatal("expected repository Create to be called")
	}

	if result != message {
		t.Fatal("expected returned message")
	}
}

func TestExternalMessageServiceCreateNilMessage(t *testing.T) {
	called := false

	messageRepository := &testMessageRepository{
		createFn: func(
			ctx context.Context,
			input *domain.Message,
		) (*domain.Message, error) {

			called = true

			return input, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.Create(
		context.Background(),
		nil,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if err == nil {
		t.Fatal("expected error")
	}

	if called {
		t.Fatal("repository should not be called")
	}
}

func TestExternalMessageServiceCreateRepositoryError(t *testing.T) {
	message := newExternalTestMessage()

	messageRepository := &testMessageRepository{
		createFn: func(
			ctx context.Context,
			input *domain.Message,
		) (*domain.Message, error) {

			return nil, errors.New("database unavailable")
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.Create(
		context.Background(),
		message,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if err == nil {
		t.Fatal("expected repository error")
	}
}

// ============================================================
// GET BY ID
// ============================================================

func TestExternalMessageServiceGetByID(t *testing.T) {
	message := newExternalTestMessage()

	messageRepository := &testMessageRepository{
		getByIDFn: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Message, error) {

			if id != message.ID {
				t.Fatalf(
					"expected id %s, got %s",
					message.ID,
					id,
				)
			}

			return message, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.GetByID(
		context.Background(),
		message.ID,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != message {
		t.Fatal("expected message returned by repository")
	}
}

func TestExternalMessageServiceGetByIDInvalidID(t *testing.T) {
	called := false

	expectedMessage := newExternalTestMessage()

	messageRepository := &testMessageRepository{
		getByIDFn: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Message, error) {

			called = true

			if id != uuid.Nil {
				t.Fatalf(
					"expected uuid.Nil, got %s",
					id,
				)
			}

			return expectedMessage, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.GetByID(
		context.Background(),
		uuid.Nil,
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !called {
		t.Fatal(
			"expected repository to be called",
		)
	}

	if result != expectedMessage {
		t.Fatal(
			"expected repository message",
		)
	}
}

// ============================================================
// LIST
// ============================================================

func TestExternalMessageServiceList(t *testing.T) {
	expected := []*domain.Message{
		newExternalTestMessage(),
		newExternalTestMessage(),
	}

	messageRepository := &testMessageRepository{
		listFn: func(
			ctx context.Context,
			limit int,
			offset int,
		) ([]*domain.Message, error) {

			if limit != 20 {
				t.Fatalf("expected limit 20, got %d", limit)
			}

			if offset != 10 {
				t.Fatalf("expected offset 10, got %d", offset)
			}

			return expected, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.List(
		context.Background(),
		20,
		10,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != len(expected) {
		t.Fatalf(
			"expected %d messages, got %d",
			len(expected),
			len(result),
		)
	}
}

func TestExternalMessageServiceListInvalidLimit(t *testing.T) {
	called := false

	messageRepository := &testMessageRepository{
		listFn: func(
			ctx context.Context,
			limit int,
			offset int,
		) ([]*domain.Message, error) {

			called = true

			if limit != 0 {
				t.Fatalf(
					"expected limit 0, got %d",
					limit,
				)
			}

			if offset != 0 {
				t.Fatalf(
					"expected offset 0, got %d",
					offset,
				)
			}

			return nil, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.List(
		context.Background(),
		0,
		0,
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !called {
		t.Fatal(
			"expected repository to be called",
		)
	}

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}
}

func TestExternalMessageServiceListInvalidOffset(t *testing.T) {
	called := false

	messageRepository := &testMessageRepository{
		listFn: func(
			ctx context.Context,
			limit int,
			offset int,
		) ([]*domain.Message, error) {

			called = true

			if limit != 20 {
				t.Fatalf(
					"expected limit 20, got %d",
					limit,
				)
			}

			if offset != -1 {
				t.Fatalf(
					"expected offset -1, got %d",
					offset,
				)
			}

			return nil, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.List(
		context.Background(),
		20,
		-1,
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !called {
		t.Fatal(
			"expected repository to be called",
		)
	}

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}
}

// ============================================================
// LIST BY CONVERSATION
// ============================================================

func TestExternalMessageServiceListByConversation(t *testing.T) {
	conversationID := uuid.New()

	expected := []*domain.Message{
		{
			ID:             uuid.New(),
			ConversationID: conversationID,
			Role:           domain.MessageRoleUser,
			Content:        "Olá",
		},
		{
			ID:             uuid.New(),
			ConversationID: conversationID,
			Role:           domain.MessageRoleAssistant,
			Content:        "Olá! Como posso ajudar?",
		},
	}

	messageRepository := &testMessageRepository{
		listByConversationFn: func(
			ctx context.Context,
			id uuid.UUID,
		) ([]*domain.Message, error) {

			if id != conversationID {
				t.Fatalf(
					"expected conversation id %s, got %s",
					conversationID,
					id,
				)
			}

			return expected, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.ListByConversation(
		context.Background(),
		conversationID,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Fatalf(
			"expected 2 messages, got %d",
			len(result),
		)
	}
}

func TestExternalMessageServiceListByConversationInvalidID(t *testing.T) {
	called := false

	messageRepository := &testMessageRepository{
		listByConversationFn: func(
			ctx context.Context,
			id uuid.UUID,
		) ([]*domain.Message, error) {

			called = true

			if id != uuid.Nil {
				t.Fatalf(
					"expected uuid.Nil, got %s",
					id,
				)
			}

			return nil, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.ListByConversation(
		context.Background(),
		uuid.Nil,
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !called {
		t.Fatal(
			"expected repository to be called",
		)
	}

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}
}

// ============================================================
// COUNT
// ============================================================

func TestExternalMessageServiceCount(t *testing.T) {
	messageRepository := &testMessageRepository{
		countFn: func(
			ctx context.Context,
		) (int64, error) {

			return 42, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.Count(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != 42 {
		t.Fatalf(
			"expected 42, got %d",
			result,
		)
	}
}

// ============================================================
// COUNT BY CONVERSATION
// ============================================================

func TestExternalMessageServiceCountByConversation(t *testing.T) {
	conversationID := uuid.New()

	messageRepository := &testMessageRepository{
		countByConversationFn: func(
			ctx context.Context,
			id uuid.UUID,
		) (int64, error) {

			if id != conversationID {
				t.Fatalf(
					"expected conversation id %s, got %s",
					conversationID,
					id,
				)
			}

			return 8, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.CountByConversation(
		context.Background(),
		conversationID,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != 8 {
		t.Fatalf(
			"expected 8, got %d",
			result,
		)
	}
}

func TestExternalMessageServiceCountByConversationInvalidID(t *testing.T) {
	called := false

	messageRepository := &testMessageRepository{
		countByConversationFn: func(
			ctx context.Context,
			id uuid.UUID,
		) (int64, error) {

			called = true

			if id != uuid.Nil {
				t.Fatalf(
					"expected uuid.Nil, got %s",
					id,
				)
			}

			return 0, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.CountByConversation(
		context.Background(),
		uuid.Nil,
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !called {
		t.Fatal(
			"expected repository to be called",
		)
	}

	if result != 0 {
		t.Fatalf(
			"expected 0, got %d",
			result,
		)
	}
}

// ============================================================
// CREATE AND GENERATE RESPONSE
// ============================================================

func TestExternalMessageServiceCreateAndGenerateResponse(t *testing.T) {
	conversation := newExternalTestConversation()

	userMessage := &domain.Message{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		Role:           domain.MessageRoleUser,
		Content:        "Estou procurando um apartamento na zona sul.",
	}

	aiResponse := &domain.AIResponse{
		Message:         "Qual é a faixa de preço que você pretende investir?",
		Intent:          "property_interest",
		NextAction:      "ask_budget",
		TransactionType: "purchase",
	}

	var createdMessages []*domain.Message

	messageRepository := &testMessageRepository{
		createFn: func(
			ctx context.Context,
			message *domain.Message,
		) (*domain.Message, error) {

			createdMessages = append(
				createdMessages,
				message,
			)

			return message, nil
		},

		listByConversationFn: func(
			ctx context.Context,
			conversationID uuid.UUID,
		) ([]*domain.Message, error) {

			return []*domain.Message{
				userMessage,
			}, nil
		},
	}

	conversationRepository := &testConversationRepository{
		getByIDFn: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {

			if id != conversation.ID {
				t.Fatalf(
					"expected conversation id %s, got %s",
					conversation.ID,
					id,
				)
			}

			return conversation, nil
		},
	}

	aiCalled := false

	aiService := &testAIService{
		generateResponseFn: func(
			ctx context.Context,
			inputConversation *domain.Conversation,
			messages []*domain.Message,
		) (*domain.AIResponse, error) {

			aiCalled = true

			if inputConversation != conversation {
				t.Fatal("expected same conversation")
			}

			if len(messages) != 1 {
				t.Fatalf(
					"expected 1 message, got %d",
					len(messages),
				)
			}

			if messages[0] != userMessage {
				t.Fatal("expected user message")
			}

			return aiResponse, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		conversationRepository,
		aiService,
	)

	result, err := service.CreateAndGenerateResponse(
		context.Background(),
		userMessage,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != aiResponse {
		t.Fatal("expected AI response")
	}

	if !aiCalled {
		t.Fatal("expected AI service to be called")
	}

	if len(createdMessages) != 2 {
		t.Fatalf(
			"expected 2 created messages, got %d",
			len(createdMessages),
		)
	}

	// First message must be the lead message.
	if createdMessages[0] != userMessage {
		t.Fatal("expected first message to be user message")
	}

	// Second message must be the AI response.
	assistantMessage := createdMessages[1]

	if assistantMessage.Role != domain.MessageRoleAssistant {
		t.Fatalf(
			"expected assistant role, got %s",
			assistantMessage.Role,
		)
	}

	if assistantMessage.ConversationID != conversation.ID {
		t.Fatal("expected same conversation ID")
	}

	if assistantMessage.Content != aiResponse.Message {
		t.Fatalf(
			"expected %q, got %q",
			aiResponse.Message,
			assistantMessage.Content,
		)
	}
}

// ============================================================
// CREATE AND GENERATE RESPONSE - NIL MESSAGE
// ============================================================

func TestExternalMessageServiceCreateAndGenerateResponseNilMessage(
	t *testing.T,
) {
	called := false

	messageRepository := &testMessageRepository{
		createFn: func(
			ctx context.Context,
			message *domain.Message,
		) (*domain.Message, error) {

			called = true

			return message, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.CreateAndGenerateResponse(
		context.Background(),
		nil,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if err == nil {
		t.Fatal("expected error")
	}

	if called {
		t.Fatal("repository should not be called")
	}
}

// ============================================================
// CREATE AND GENERATE RESPONSE - MESSAGE REPOSITORY ERROR
// ============================================================

func TestExternalMessageServiceCreateAndGenerateResponseMessageRepositoryError(
	t *testing.T,
) {
	message := newExternalTestMessage()

	messageRepository := &testMessageRepository{
		createFn: func(
			ctx context.Context,
			input *domain.Message,
		) (*domain.Message, error) {

			return nil, errors.New("message repository error")
		},
	}

	service := newExternalTestService(
		messageRepository,
		&testConversationRepository{},
		&testAIService{},
	)

	result, err := service.CreateAndGenerateResponse(
		context.Background(),
		message,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if err == nil {
		t.Fatal("expected repository error")
	}
}

// ============================================================
// CREATE AND GENERATE RESPONSE - CONVERSATION ERROR
// ============================================================

func TestExternalMessageServiceCreateAndGenerateResponseConversationError(
	t *testing.T,
) {
	message := newExternalTestMessage()

	messageRepository := &testMessageRepository{
		createFn: func(
			ctx context.Context,
			input *domain.Message,
		) (*domain.Message, error) {

			return input, nil
		},
	}

	conversationRepository := &testConversationRepository{
		getByIDFn: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {

			return nil, errors.New("conversation repository error")
		},
	}

	service := newExternalTestService(
		messageRepository,
		conversationRepository,
		&testAIService{},
	)

	result, err := service.CreateAndGenerateResponse(
		context.Background(),
		message,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if err == nil {
		t.Fatal("expected conversation repository error")
	}
}

// ============================================================
// CREATE AND GENERATE RESPONSE - CONVERSATION NOT FOUND
// ============================================================

func TestExternalMessageServiceCreateAndGenerateResponseConversationNotFound(
	t *testing.T,
) {
	message := newExternalTestMessage()

	messageRepository := &testMessageRepository{
		createFn: func(
			ctx context.Context,
			input *domain.Message,
		) (*domain.Message, error) {

			return input, nil
		},
	}

	conversationRepository := &testConversationRepository{
		getByIDFn: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {

			return nil, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		conversationRepository,
		&testAIService{},
	)

	result, err := service.CreateAndGenerateResponse(
		context.Background(),
		message,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if err == nil {
		t.Fatal("expected conversation not found error")
	}
}

// ============================================================
// CREATE AND GENERATE RESPONSE - LIST MESSAGES ERROR
// ============================================================

func TestExternalMessageServiceCreateAndGenerateResponseListMessagesError(
	t *testing.T,
) {
	conversation := newExternalTestConversation()

	message := &domain.Message{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		Role:           domain.MessageRoleUser,
		Content:        "Olá",
	}

	messageRepository := &testMessageRepository{
		createFn: func(
			ctx context.Context,
			input *domain.Message,
		) (*domain.Message, error) {

			return input, nil
		},

		listByConversationFn: func(
			ctx context.Context,
			conversationID uuid.UUID,
		) ([]*domain.Message, error) {

			return nil, errors.New("list messages error")
		},
	}

	conversationRepository := &testConversationRepository{
		getByIDFn: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {

			return conversation, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		conversationRepository,
		&testAIService{},
	)

	result, err := service.CreateAndGenerateResponse(
		context.Background(),
		message,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if err == nil {
		t.Fatal("expected list messages error")
	}
}

// ============================================================
// CREATE AND GENERATE RESPONSE - AI ERROR
// ============================================================

func TestExternalMessageServiceCreateAndGenerateResponseAIError(
	t *testing.T,
) {
	conversation := newExternalTestConversation()

	message := &domain.Message{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		Role:           domain.MessageRoleUser,
		Content:        "Olá",
	}

	messageRepository := &testMessageRepository{
		createFn: func(
			ctx context.Context,
			input *domain.Message,
		) (*domain.Message, error) {

			return input, nil
		},

		listByConversationFn: func(
			ctx context.Context,
			conversationID uuid.UUID,
		) ([]*domain.Message, error) {

			return []*domain.Message{
				message,
			}, nil
		},
	}

	conversationRepository := &testConversationRepository{
		getByIDFn: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {

			return conversation, nil
		},
	}

	aiService := &testAIService{
		generateResponseFn: func(
			ctx context.Context,
			inputConversation *domain.Conversation,
			messages []*domain.Message,
		) (*domain.AIResponse, error) {

			return nil, errors.New("AI service error")
		},
	}

	service := newExternalTestService(
		messageRepository,
		conversationRepository,
		aiService,
	)

	result, err := service.CreateAndGenerateResponse(
		context.Background(),
		message,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if err == nil {
		t.Fatal("expected AI service error")
	}
}

// ============================================================
// CREATE AND GENERATE RESPONSE - NIL AI RESPONSE
// ============================================================

func TestExternalMessageServiceCreateAndGenerateResponseNilAIResponse(
	t *testing.T,
) {
	conversation := newExternalTestConversation()

	message := &domain.Message{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		Role:           domain.MessageRoleUser,
		Content:        "Olá",
	}

	messageRepository := &testMessageRepository{
		createFn: func(
			ctx context.Context,
			input *domain.Message,
		) (*domain.Message, error) {

			return input, nil
		},

		listByConversationFn: func(
			ctx context.Context,
			conversationID uuid.UUID,
		) ([]*domain.Message, error) {

			return []*domain.Message{
				message,
			}, nil
		},
	}

	conversationRepository := &testConversationRepository{
		getByIDFn: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {

			return conversation, nil
		},
	}

	aiService := &testAIService{
		generateResponseFn: func(
			ctx context.Context,
			inputConversation *domain.Conversation,
			messages []*domain.Message,
		) (*domain.AIResponse, error) {

			return nil, nil
		},
	}

	service := newExternalTestService(
		messageRepository,
		conversationRepository,
		aiService,
	)

	result, err := service.CreateAndGenerateResponse(
		context.Background(),
		message,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if err == nil {
		t.Fatal("expected nil AI response error")
	}
}
