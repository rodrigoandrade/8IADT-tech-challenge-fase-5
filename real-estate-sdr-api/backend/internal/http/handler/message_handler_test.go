package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"real-estate-sdr/internal/application"
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

func newTestMessageService(
	repository *messageRepositoryMock,
) *application.MessageService {
	return application.NewMessageService(
		repository,
		nil,
		nil,
	)
}

func newTestMessage() *domain.Message {
	return &domain.Message{
		ID:             uuid.New(),
		ConversationID: uuid.New(),
		Role:           "USER",
		Content:        "Gostaria de saber mais sobre este imóvel.",
	}
}

func TestMessageHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := newTestMessage()

		repository := &messageRepositoryMock{
			createFn: func(
				_ context.Context,
				message *domain.Message,
			) (*domain.Message, error) {
				message.ID = expected.ID
				return message, nil
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		payload := []byte(`{
			"conversation_id": "` + expected.ConversationID.String() + `",
			"role": "USER",
			"content": "Gostaria de saber mais sobre este imóvel."
		}`)

		req := httptest.NewRequest(
			http.MethodPost,
			"/messages",
			strings.NewReader(string(payload)),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		handler.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf(
				"expected status %d, got %d: %s",
				http.StatusCreated,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		service := newTestMessageService(
			&messageRepositoryMock{},
		)

		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodPost,
			"/messages",
			strings.NewReader(`{"conversation_id":`),
		)

		rec := httptest.NewRecorder()

		handler.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("service error", func(t *testing.T) {
		repository := &messageRepositoryMock{
			createFn: func(
				_ context.Context,
				_ *domain.Message,
			) (*domain.Message, error) {
				return nil, errors.New("database failure")
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		body := `{
			"conversation_id": "` + uuid.New().String() + `",
			"role": "USER",
			"content": "Mensagem"
		}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/messages",
			strings.NewReader(body),
		)

		rec := httptest.NewRecorder()

		handler.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d: %s",
				http.StatusBadRequest,
				rec.Code,
				rec.Body.String(),
			)
		}
	})
}

func TestMessageHandler_GetByID(t *testing.T) {
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

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/messages/"+message.ID.String(),
			nil,
		)

		rec := httptest.NewRecorder()

		handler.GetByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		service := newTestMessageService(
			&messageRepositoryMock{},
		)

		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/messages/invalid-id",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.GetByID(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repository := &messageRepositoryMock{
			getByIDFn: func(
				_ context.Context,
				_ uuid.UUID,
			) (*domain.Message, error) {
				return nil, errors.New("message not found")
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		id := uuid.New()

		req := httptest.NewRequest(
			http.MethodGet,
			"/messages/"+id.String(),
			nil,
		)

		rec := httptest.NewRecorder()

		handler.GetByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf(
				"expected status %d, got %d: %s",
				http.StatusNotFound,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repository := &messageRepositoryMock{
			getByIDFn: func(
				_ context.Context,
				_ uuid.UUID,
			) (*domain.Message, error) {
				return nil, errors.New("database unavailable")
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/messages/"+uuid.New().String(),
			nil,
		)

		rec := httptest.NewRecorder()

		handler.GetByID(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})
}

func TestMessageHandler_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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

				return []*domain.Message{
					newTestMessage(),
					newTestMessage(),
				}, nil
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/messages?limit=10&offset=20",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("default pagination", func(t *testing.T) {
		repository := &messageRepositoryMock{
			listFn: func(
				_ context.Context,
				limit int,
				offset int,
			) ([]*domain.Message, error) {
				if limit != 20 {
					t.Fatalf(
						"expected default limit 20, got %d",
						limit,
					)
				}

				if offset != 0 {
					t.Fatalf(
						"expected default offset 0, got %d",
						offset,
					)
				}

				return []*domain.Message{}, nil
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/messages",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusOK,
				rec.Code,
			)
		}
	})

	t.Run("invalid limit", func(t *testing.T) {
		service := newTestMessageService(
			&messageRepositoryMock{},
		)

		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/messages?limit=invalid",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.List(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("invalid offset", func(t *testing.T) {
		service := newTestMessageService(
			&messageRepositoryMock{},
		)

		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/messages?offset=-1",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.List(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("service error", func(t *testing.T) {
		repository := &messageRepositoryMock{
			listFn: func(
				_ context.Context,
				_ int,
				_ int,
			) ([]*domain.Message, error) {
				return nil, errors.New("database failure")
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/messages",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.List(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})
}

func TestMessageHandler_ListByConversation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		conversationID := uuid.New()

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

				return []*domain.Message{
					newTestMessage(),
				}, nil
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/conversations/"+conversationID.String()+"/messages",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.ListByConversation(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("invalid conversation id", func(t *testing.T) {
		service := newTestMessageService(
			&messageRepositoryMock{},
		)

		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/conversations/invalid-id/messages",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.ListByConversation(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("service error", func(t *testing.T) {
		repository := &messageRepositoryMock{
			listByConversationFn: func(
				_ context.Context,
				_ uuid.UUID,
			) ([]*domain.Message, error) {
				return nil, errors.New("database failure")
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/conversations/"+uuid.New().String()+"/messages",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.ListByConversation(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})
}

func TestMessageHandler_Count(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repository := &messageRepositoryMock{
			countFn: func(
				_ context.Context,
			) (int64, error) {
				return 42, nil
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/messages/count",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.Count(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusOK,
				rec.Code,
			)
		}

		if !strings.Contains(rec.Body.String(), `"count":42`) {
			t.Fatalf(
				"expected count 42, got %s",
				rec.Body.String(),
			)
		}
	})

	t.Run("service error", func(t *testing.T) {
		repository := &messageRepositoryMock{
			countFn: func(
				_ context.Context,
			) (int64, error) {
				return 0, errors.New("database failure")
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/messages/count",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.Count(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})
}

func TestMessageHandler_CountByConversation(t *testing.T) {
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

				return 8, nil
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/conversations/"+conversationID.String()+"/messages/count",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.CountByConversation(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		if !strings.Contains(rec.Body.String(), `"count":8`) {
			t.Fatalf(
				"expected count 8, got %s",
				rec.Body.String(),
			)
		}
	})

	t.Run("invalid conversation id", func(t *testing.T) {
		service := newTestMessageService(
			&messageRepositoryMock{},
		)

		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/conversations/invalid-id/messages/count",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.CountByConversation(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("service error", func(t *testing.T) {
		repository := &messageRepositoryMock{
			countByConversationFn: func(
				_ context.Context,
				_ uuid.UUID,
			) (int64, error) {
				return 0, errors.New("database failure")
			},
		}

		service := newTestMessageService(repository)
		handler := NewMessageHandler(service)

		req := httptest.NewRequest(
			http.MethodGet,
			"/conversations/"+uuid.New().String()+"/messages/count",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.CountByConversation(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})
}
