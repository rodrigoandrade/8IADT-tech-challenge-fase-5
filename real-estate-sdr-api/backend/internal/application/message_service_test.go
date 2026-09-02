package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"real-estate-sdr/internal/domain"
)

type messageRepositoryMock struct {
	createFn              func(context.Context, *domain.Message) (*domain.Message, error)
	getByIDFn             func(context.Context, uuid.UUID) (*domain.Message, error)
	listFn                func(context.Context, int, int) ([]*domain.Message, error)
	listByConversationFn  func(context.Context, uuid.UUID) ([]*domain.Message, error)
	countFn               func(context.Context) (int64, error)
	countByConversationFn func(context.Context, uuid.UUID) (int64, error)
}

func (m *messageRepositoryMock) Create(
	ctx context.Context,
	message *domain.Message,
) (*domain.Message, error) {
	if m.createFn != nil {
		return m.createFn(ctx, message)
	}

	return nil, errors.New("create not configured")
}

func (m *messageRepositoryMock) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Message, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return nil, errors.New("get by id not configured")
}

func (m *messageRepositoryMock) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.Message, error) {
	if m.listFn != nil {
		return m.listFn(ctx, limit, offset)
	}

	return nil, errors.New("list not configured")
}

func (m *messageRepositoryMock) ListByConversation(
	ctx context.Context,
	conversationID uuid.UUID,
) ([]*domain.Message, error) {
	if m.listByConversationFn != nil {
		return m.listByConversationFn(ctx, conversationID)
	}

	return nil, errors.New("list by conversation not configured")
}

func (m *messageRepositoryMock) Count(
	ctx context.Context,
) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx)
	}

	return 0, errors.New("count not configured")
}

func (m *messageRepositoryMock) CountByConversation(
	ctx context.Context,
	conversationID uuid.UUID,
) (int64, error) {
	if m.countByConversationFn != nil {
		return m.countByConversationFn(ctx, conversationID)
	}

	return 0, errors.New("count by conversation not configured")
}

func newTestMessage() *domain.Message {
	return &domain.Message{
		ID:             uuid.New(),
		ConversationID: uuid.New(),
		Role:           "USER",
		Content:        "Olá, gostaria de saber mais sobre este imóvel.",
	}
}

func TestMessageService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		message := newTestMessage()

		repository := &messageRepositoryMock{
			createFn: func(
				_ context.Context,
				input *domain.Message,
			) (*domain.Message, error) {
				return input, nil
			},
		}

		service := NewMessageService(repository)

		result, err := service.Create(ctx, message)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("expected message, got nil")
		}

		if result.ID != message.ID {
			t.Fatalf(
				"expected id %s, got %s",
				message.ID,
				result.ID,
			)
		}
	})

	t.Run("nil message", func(t *testing.T) {
		service := NewMessageService(&messageRepositoryMock{})

		result, err := service.Create(
			context.Background(),
			nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}
	})

	t.Run("invalid message", func(t *testing.T) {
		message := newTestMessage()
		message.Content = ""

		service := NewMessageService(&messageRepositoryMock{})

		result, err := service.Create(
			context.Background(),
			message,
		)

		if err == nil {
			t.Fatal("expected validation error, got nil")
		}

		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("repository failure")

		repository := &messageRepositoryMock{
			createFn: func(
				_ context.Context,
				_ *domain.Message,
			) (*domain.Message, error) {
				return nil, expectedErr
			},
		}

		service := NewMessageService(repository)

		result, err := service.Create(
			context.Background(),
			newTestMessage(),
		)

		if err == nil {
			t.Fatal("expected repository error, got nil")
		}

		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"expected wrapped repository error, got %v",
				err,
			)
		}
	})
}

func TestMessageService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		message := newTestMessage()

		repository := &messageRepositoryMock{
			getByIDFn: func(
				_ context.Context,
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

		service := NewMessageService(repository)

		result, err := service.GetByID(
			context.Background(),
			message.ID,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("expected message, got nil")
		}

		if result.ID != message.ID {
			t.Fatalf(
				"expected id %s, got %s",
				message.ID,
				result.ID,
			)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		service := NewMessageService(&messageRepositoryMock{})

		result, err := service.GetByID(
			context.Background(),
			uuid.Nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("repository failure")

		repository := &messageRepositoryMock{
			getByIDFn: func(
				_ context.Context,
				_ uuid.UUID,
			) (*domain.Message, error) {
				return nil, expectedErr
			},
		}

		service := NewMessageService(repository)

		result, err := service.GetByID(
			context.Background(),
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected repository error, got nil")
		}

		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"expected wrapped repository error, got %v",
				err,
			)
		}
	})
}

func TestMessageService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		messages := []*domain.Message{
			newTestMessage(),
			newTestMessage(),
		}

		repository := &messageRepositoryMock{
			listFn: func(
				_ context.Context,
				limit int,
				offset int,
			) ([]*domain.Message, error) {
				if limit != 10 {
					t.Fatalf("expected limit 10, got %d", limit)
				}

				if offset != 20 {
					t.Fatalf("expected offset 20, got %d", offset)
				}

				return messages, nil
			},
		}

		service := NewMessageService(repository)

		result, err := service.List(
			context.Background(),
			10,
			20,
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
	})

	t.Run("invalid limit", func(t *testing.T) {
		service := NewMessageService(&messageRepositoryMock{})

		result, err := service.List(
			context.Background(),
			0,
			0,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}
	})

	t.Run("negative limit", func(t *testing.T) {
		service := NewMessageService(&messageRepositoryMock{})

		result, err := service.List(
			context.Background(),
			-1,
			0,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}
	})

	t.Run("negative offset", func(t *testing.T) {
		service := NewMessageService(&messageRepositoryMock{})

		result, err := service.List(
			context.Background(),
			10,
			-1,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("repository failure")

		repository := &messageRepositoryMock{
			listFn: func(
				_ context.Context,
				_ int,
				_ int,
			) ([]*domain.Message, error) {
				return nil, expectedErr
			},
		}

		service := NewMessageService(repository)

		result, err := service.List(
			context.Background(),
			10,
			0,
		)

		if err == nil {
			t.Fatal("expected repository error, got nil")
		}

		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"expected wrapped repository error, got %v",
				err,
			)
		}
	})
}

func TestMessageService_ListByConversation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		conversationID := uuid.New()
		messages := []*domain.Message{
			newTestMessage(),
			newTestMessage(),
		}

		repository := &messageRepositoryMock{
			listByConversationFn: func(
				_ context.Context,
				id uuid.UUID,
			) ([]*domain.Message, error) {
				if id != conversationID {
					t.Fatalf(
						"expected conversation id %s, got %s",
						conversationID,
						id,
					)
				}

				return messages, nil
			},
		}

		service := NewMessageService(repository)

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
	})

	t.Run("invalid conversation id", func(t *testing.T) {
		service := NewMessageService(&messageRepositoryMock{})

		result, err := service.ListByConversation(
			context.Background(),
			uuid.Nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("repository failure")

		repository := &messageRepositoryMock{
			listByConversationFn: func(
				_ context.Context,
				_ uuid.UUID,
			) ([]*domain.Message, error) {
				return nil, expectedErr
			},
		}

		service := NewMessageService(repository)

		result, err := service.ListByConversation(
			context.Background(),
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected repository error, got nil")
		}

		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"expected wrapped repository error, got %v",
				err,
			)
		}
	})
}

func TestMessageService_Count(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repository := &messageRepositoryMock{
			countFn: func(
				_ context.Context,
			) (int64, error) {
				return 15, nil
			},
		}

		service := NewMessageService(repository)

		result, err := service.Count(
			context.Background(),
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result != 15 {
			t.Fatalf(
				"expected count 15, got %d",
				result,
			)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("repository failure")

		repository := &messageRepositoryMock{
			countFn: func(
				_ context.Context,
			) (int64, error) {
				return 0, expectedErr
			},
		}

		service := NewMessageService(repository)

		result, err := service.Count(
			context.Background(),
		)

		if err == nil {
			t.Fatal("expected repository error, got nil")
		}

		if result != 0 {
			t.Fatalf("expected zero result, got %d", result)
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"expected wrapped repository error, got %v",
				err,
			)
		}
	})
}

func TestMessageService_CountByConversation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		conversationID := uuid.New()

		repository := &messageRepositoryMock{
			countByConversationFn: func(
				_ context.Context,
				id uuid.UUID,
			) (int64, error) {
				if id != conversationID {
					t.Fatalf(
						"expected conversation id %s, got %s",
						conversationID,
						id,
					)
				}

				return 7, nil
			},
		}

		service := NewMessageService(repository)

		result, err := service.CountByConversation(
			context.Background(),
			conversationID,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result != 7 {
			t.Fatalf(
				"expected count 7, got %d",
				result,
			)
		}
	})

	t.Run("invalid conversation id", func(t *testing.T) {
		service := NewMessageService(&messageRepositoryMock{})

		result, err := service.CountByConversation(
			context.Background(),
			uuid.Nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if result != 0 {
			t.Fatalf("expected zero result, got %d", result)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("repository failure")

		repository := &messageRepositoryMock{
			countByConversationFn: func(
				_ context.Context,
				_ uuid.UUID,
			) (int64, error) {
				return 0, expectedErr
			},
		}

		service := NewMessageService(repository)

		result, err := service.CountByConversation(
			context.Background(),
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected repository error, got nil")
		}

		if result != 0 {
			t.Fatalf("expected zero result, got %d", result)
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"expected wrapped repository error, got %v",
				err,
			)
		}
	})
}
