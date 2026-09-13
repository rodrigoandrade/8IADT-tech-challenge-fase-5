package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"real-estate-sdr/internal/domain"
)

// ============================================================
// CONVERSATION REPOSITORY MOCK
// ============================================================

type conversationRepositoryMock struct {
	createFunc      func(context.Context, *domain.Conversation) (*domain.Conversation, error)
	getByIDFunc     func(context.Context, uuid.UUID) (*domain.Conversation, error)
	getByLeadIDFunc func(context.Context, uuid.UUID) ([]*domain.Conversation, error)
	listFunc        func(context.Context, int, int) ([]*domain.Conversation, error)
	closeFunc       func(context.Context, uuid.UUID) (*domain.Conversation, error)
	countFunc       func(context.Context) (int64, error)
}

func (m *conversationRepositoryMock) Create(
	ctx context.Context,
	conversation *domain.Conversation,
) (*domain.Conversation, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, conversation)
	}

	return nil, errors.New("create not configured")
}

func (m *conversationRepositoryMock) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Conversation, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}

	return nil, errors.New("get by id not configured")
}

func (m *conversationRepositoryMock) GetByLeadID(
	ctx context.Context,
	leadID uuid.UUID,
) ([]*domain.Conversation, error) {
	if m.getByLeadIDFunc != nil {
		return m.getByLeadIDFunc(ctx, leadID)
	}

	return nil, errors.New("get by lead id not configured")
}

func (m *conversationRepositoryMock) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.Conversation, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}

	return nil, errors.New("list not configured")
}

func (m *conversationRepositoryMock) Close(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Conversation, error) {
	if m.closeFunc != nil {
		return m.closeFunc(ctx, id)
	}

	return nil, errors.New("close not configured")
}

func (m *conversationRepositoryMock) Count(
	ctx context.Context,
) (int64, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx)
	}

	return 0, errors.New("count not configured")
}

// ============================================================
// HELPERS
// ============================================================

func validConversationForServiceTest() *domain.Conversation {
	return &domain.Conversation{
		ID:        uuid.New(),
		LeadID:    uuid.New(),
		Status:    domain.ConversationStatusActive,
		StartedAt: time.Now(),
	}
}

// ============================================================
// CREATE
// ============================================================

func TestConversationService_Create(t *testing.T) {
	conversation := validConversationForServiceTest()

	var received *domain.Conversation

	repository := &conversationRepositoryMock{
		createFunc: func(
			ctx context.Context,
			input *domain.Conversation,
		) (*domain.Conversation, error) {
			received = input
			return input, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.Create(
		context.Background(),
		conversation,
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result != conversation {
		t.Fatal(
			"expected returned conversation to be repository result",
		)
	}

	if received != conversation {
		t.Fatal(
			"expected repository to receive same conversation",
		)
	}
}

func TestConversationService_Create_NilConversation(t *testing.T) {
	repository := &conversationRepositoryMock{}

	service := NewConversationService(repository)

	result, err := service.Create(
		context.Background(),
		nil,
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if !errors.Is(err, domain.ErrInvalidConversation) {
		t.Fatalf(
			"expected ErrInvalidConversation, got %v",
			err,
		)
	}
}

func TestConversationService_Create_InvalidConversation(t *testing.T) {
	conversation := validConversationForServiceTest()

	conversation.LeadID = uuid.Nil

	repositoryCalled := false

	repository := &conversationRepositoryMock{
		createFunc: func(
			ctx context.Context,
			input *domain.Conversation,
		) (*domain.Conversation, error) {
			repositoryCalled = true
			return input, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.Create(
		context.Background(),
		conversation,
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	if !repositoryCalled {
		return
	}

	t.Fatal(
		"repository should not be called for invalid conversation",
	)
}

func TestConversationService_Create_RepositoryError(t *testing.T) {
	conversation := validConversationForServiceTest()

	expectedErr := errors.New("repository error")

	repository := &conversationRepositoryMock{
		createFunc: func(
			ctx context.Context,
			input *domain.Conversation,
		) (*domain.Conversation, error) {
			return nil, expectedErr
		},
	}

	service := NewConversationService(repository)

	result, err := service.Create(
		context.Background(),
		conversation,
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

// ============================================================
// GET BY ID
// ============================================================

func TestConversationService_GetByID(t *testing.T) {
	id := uuid.New()

	conversation := validConversationForServiceTest()
	conversation.ID = id

	var receivedID uuid.UUID

	repository := &conversationRepositoryMock{
		getByIDFunc: func(
			ctx context.Context,
			inputID uuid.UUID,
		) (*domain.Conversation, error) {
			receivedID = inputID
			return conversation, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.GetByID(
		context.Background(),
		id,
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result != conversation {
		t.Fatal(
			"expected returned conversation from repository",
		)
	}

	if receivedID != id {
		t.Fatalf(
			"expected ID %s, got %s",
			id,
			receivedID,
		)
	}
}

func TestConversationService_GetByID_InvalidID(t *testing.T) {
	repositoryCalled := false

	repository := &conversationRepositoryMock{
		getByIDFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {
			repositoryCalled = true
			return nil, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.GetByID(
		context.Background(),
		uuid.Nil,
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if err == nil {
		t.Fatal(
			"expected invalid ID error, got nil",
		)
	}

	if repositoryCalled {
		t.Fatal(
			"repository should not be called with invalid ID",
		)
	}
}

func TestConversationService_GetByID_RepositoryError(t *testing.T) {
	expectedErr := errors.New("repository error")

	repository := &conversationRepositoryMock{
		getByIDFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {
			return nil, expectedErr
		},
	}

	service := NewConversationService(repository)

	_, err := service.GetByID(
		context.Background(),
		uuid.New(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestConversationService_GetByID_NotFound(t *testing.T) {
	repository := &conversationRepositoryMock{
		getByIDFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {
			return nil, pgx.ErrNoRows
		},
	}

	service := NewConversationService(repository)

	result, err := service.GetByID(
		context.Background(),
		uuid.New(),
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf(
			"expected ErrConversationNotFound, got %v",
			err,
		)
	}
}

// ============================================================
// GET BY LEAD ID
// ============================================================

func TestConversationService_GetByLeadID(t *testing.T) {
	leadID := uuid.New()

	conversations := []*domain.Conversation{
		validConversationForServiceTest(),
		validConversationForServiceTest(),
	}

	var receivedLeadID uuid.UUID

	repository := &conversationRepositoryMock{
		getByLeadIDFunc: func(
			ctx context.Context,
			inputLeadID uuid.UUID,
		) ([]*domain.Conversation, error) {
			receivedLeadID = inputLeadID
			return conversations, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.GetByLeadID(
		context.Background(),
		leadID,
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(result) != len(conversations) {
		t.Fatalf(
			"expected %d conversations, got %d",
			len(conversations),
			len(result),
		)
	}

	if receivedLeadID != leadID {
		t.Fatalf(
			"expected lead ID %s, got %s",
			leadID,
			receivedLeadID,
		)
	}
}

func TestConversationService_GetByLeadID_InvalidID(t *testing.T) {
	repositoryCalled := false

	repository := &conversationRepositoryMock{
		getByLeadIDFunc: func(
			ctx context.Context,
			leadID uuid.UUID,
		) ([]*domain.Conversation, error) {
			repositoryCalled = true
			return nil, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.GetByLeadID(
		context.Background(),
		uuid.Nil,
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if err == nil {
		t.Fatal(
			"expected invalid lead ID error, got nil",
		)
	}

	if repositoryCalled {
		t.Fatal(
			"repository should not be called with invalid lead ID",
		)
	}
}

func TestConversationService_GetByLeadID_RepositoryError(t *testing.T) {
	expectedErr := errors.New("repository error")

	repository := &conversationRepositoryMock{
		getByLeadIDFunc: func(
			ctx context.Context,
			leadID uuid.UUID,
		) ([]*domain.Conversation, error) {
			return nil, expectedErr
		},
	}

	service := NewConversationService(repository)

	_, err := service.GetByLeadID(
		context.Background(),
		uuid.New(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

// ============================================================
// LIST
// ============================================================

func TestConversationService_List(t *testing.T) {
	conversations := []*domain.Conversation{
		validConversationForServiceTest(),
		validConversationForServiceTest(),
	}

	var receivedLimit int
	var receivedOffset int

	repository := &conversationRepositoryMock{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) ([]*domain.Conversation, error) {
			receivedLimit = limit
			receivedOffset = offset

			return conversations, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.List(
		context.Background(),
		10,
		20,
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(result) != len(conversations) {
		t.Fatalf(
			"expected %d conversations, got %d",
			len(conversations),
			len(result),
		)
	}

	if receivedLimit != 10 {
		t.Fatalf(
			"expected limit 10, got %d",
			receivedLimit,
		)
	}

	if receivedOffset != 20 {
		t.Fatalf(
			"expected offset 20, got %d",
			receivedOffset,
		)
	}
}

func TestConversationService_List_DefaultPagination(t *testing.T) {
	var receivedLimit int
	var receivedOffset int

	repository := &conversationRepositoryMock{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) ([]*domain.Conversation, error) {
			receivedLimit = limit
			receivedOffset = offset

			return nil, nil
		},
	}

	service := NewConversationService(repository)

	_, err := service.List(
		context.Background(),
		0,
		-10,
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if receivedLimit != 20 {
		t.Fatalf(
			"expected default limit 20, got %d",
			receivedLimit,
		)
	}

	if receivedOffset != 0 {
		t.Fatalf(
			"expected default offset 0, got %d",
			receivedOffset,
		)
	}
}

func TestConversationService_List_RepositoryError(t *testing.T) {
	expectedErr := errors.New("repository error")

	repository := &conversationRepositoryMock{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) ([]*domain.Conversation, error) {
			return nil, expectedErr
		},
	}

	service := NewConversationService(repository)

	_, err := service.List(
		context.Background(),
		10,
		0,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

// ============================================================
// CLOSE
// ============================================================

func TestConversationService_Close(t *testing.T) {
	id := uuid.New()

	conversation := validConversationForServiceTest()
	conversation.ID = id

	closedConversation := *conversation

	now := time.Now()

	closedConversation.Status = domain.ConversationStatusClosed
	closedConversation.EndedAt = &now

	var receivedGetID uuid.UUID
	var receivedCloseID uuid.UUID

	repository := &conversationRepositoryMock{
		getByIDFunc: func(
			ctx context.Context,
			inputID uuid.UUID,
		) (*domain.Conversation, error) {

			receivedGetID = inputID

			return conversation, nil
		},

		closeFunc: func(
			ctx context.Context,
			inputID uuid.UUID,
		) (*domain.Conversation, error) {

			receivedCloseID = inputID

			return &closedConversation, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.Close(
		context.Background(),
		id,
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result != &closedConversation {
		t.Fatal(
			"expected returned closed conversation",
		)
	}

	if receivedGetID != id {
		t.Fatalf(
			"expected GetByID ID %s, got %s",
			id,
			receivedGetID,
		)
	}

	if receivedCloseID != id {
		t.Fatalf(
			"expected Close ID %s, got %s",
			id,
			receivedCloseID,
		)
	}
}

func TestConversationService_Close_InvalidID(t *testing.T) {
	repositoryCalled := false

	repository := &conversationRepositoryMock{
		getByIDFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {
			repositoryCalled = true
			return nil, nil
		},

		closeFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {
			repositoryCalled = true
			return nil, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.Close(
		context.Background(),
		uuid.Nil,
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if err == nil {
		t.Fatal(
			"expected invalid conversation ID error",
		)
	}

	if repositoryCalled {
		t.Fatal(
			"repository should not be called with invalid ID",
		)
	}
}

func TestConversationService_Close_NotFound(t *testing.T) {
	repository := &conversationRepositoryMock{
		getByIDFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {
			return nil, pgx.ErrNoRows
		},

		closeFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {
			t.Fatal(
				"Close should not be called when conversation does not exist",
			)

			return nil, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.Close(
		context.Background(),
		uuid.New(),
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf(
			"expected ErrConversationNotFound, got %v",
			err,
		)
	}
}

func TestConversationService_Close_GetByIDError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	repository := &conversationRepositoryMock{
		getByIDFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {
			return nil, expectedErr
		},

		closeFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Conversation, error) {
			t.Fatal(
				"Close should not be called when GetByID fails",
			)

			return nil, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.Close(
		context.Background(),
		uuid.New(),
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestConversationService_Close_AlreadyClosed(t *testing.T) {
	id := uuid.New()

	now := time.Now()

	conversation := validConversationForServiceTest()

	conversation.ID = id
	conversation.Status = domain.ConversationStatusClosed
	conversation.EndedAt = &now

	repository := &conversationRepositoryMock{
		getByIDFunc: func(
			ctx context.Context,
			inputID uuid.UUID,
		) (*domain.Conversation, error) {
			return conversation, nil
		},

		closeFunc: func(
			ctx context.Context,
			inputID uuid.UUID,
		) (*domain.Conversation, error) {
			t.Fatal(
				"Close should not be called for already closed conversation",
			)

			return nil, nil
		},
	}

	service := NewConversationService(repository)

	result, err := service.Close(
		context.Background(),
		id,
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if !errors.Is(err, ErrConversationAlreadyClosed) {
		t.Fatalf(
			"expected ErrConversationAlreadyClosed, got %v",
			err,
		)
	}
}

func TestConversationService_Close_RepositoryError(t *testing.T) {
	id := uuid.New()

	conversation := validConversationForServiceTest()
	conversation.ID = id

	expectedErr := errors.New("close repository error")

	repository := &conversationRepositoryMock{
		getByIDFunc: func(
			ctx context.Context,
			inputID uuid.UUID,
		) (*domain.Conversation, error) {
			return conversation, nil
		},

		closeFunc: func(
			ctx context.Context,
			inputID uuid.UUID,
		) (*domain.Conversation, error) {
			return nil, expectedErr
		},
	}

	service := NewConversationService(repository)

	result, err := service.Close(
		context.Background(),
		id,
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestConversationService_Close_RaceAlreadyClosed(t *testing.T) {
	id := uuid.New()

	conversation := validConversationForServiceTest()
	conversation.ID = id

	repository := &conversationRepositoryMock{
		getByIDFunc: func(
			ctx context.Context,
			inputID uuid.UUID,
		) (*domain.Conversation, error) {
			return conversation, nil
		},

		closeFunc: func(
			ctx context.Context,
			inputID uuid.UUID,
		) (*domain.Conversation, error) {
			return nil, pgx.ErrNoRows
		},
	}

	service := NewConversationService(repository)

	result, err := service.Close(
		context.Background(),
		id,
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %v",
			result,
		)
	}

	if !errors.Is(err, ErrConversationAlreadyClosed) {
		t.Fatalf(
			"expected ErrConversationAlreadyClosed, got %v",
			err,
		)
	}
}

// ============================================================
// COUNT
// ============================================================

func TestConversationService_Count(t *testing.T) {
	repository := &conversationRepositoryMock{
		countFunc: func(
			ctx context.Context,
		) (int64, error) {
			return 42, nil
		},
	}

	service := NewConversationService(repository)

	count, err := service.Count(
		context.Background(),
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if count != 42 {
		t.Fatalf(
			"expected count 42, got %d",
			count,
		)
	}
}

func TestConversationService_Count_RepositoryError(t *testing.T) {
	expectedErr := errors.New("count error")

	repository := &conversationRepositoryMock{
		countFunc: func(
			ctx context.Context,
		) (int64, error) {
			return 0, expectedErr
		},
	}

	service := NewConversationService(repository)

	_, err := service.Count(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}
