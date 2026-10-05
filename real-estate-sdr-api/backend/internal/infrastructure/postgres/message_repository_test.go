package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"real-estate-sdr/internal/domain"
)

func TestMessageRepository(t *testing.T) {
	ctx := context.Background()

	pool := testPool(t)
	queries := newQueries(pool)

	repository := NewMessageRepository(queries)

	leadID := uuid.New()

	leadRepository := NewLeadRepository(queries)

	lead := &domain.Lead{
		ID:    leadID,
		Name:  "Message Test Lead",
		Email: "message-test-" + uuid.NewString() + "@example.com",
		Phone: "11999999999",
	}

	createdLead, err := leadRepository.Create(ctx, lead)
	if err != nil {
		t.Fatalf("create lead: %v", err)
	}

	conversationRepository := NewConversationRepository(queries)

	conversation := &domain.Conversation{
		LeadID: createdLead.ID,
		Status: domain.ConversationStatusActive,
	}

	createdConversation, err := conversationRepository.Create(
		ctx,
		conversation,
	)
	if err != nil {
		t.Fatalf(
			"create conversation: %v",
			err,
		)
	}

	t.Run("Create", func(t *testing.T) {
		message := &domain.Message{
			ConversationID: createdConversation.ID,
			Role:           domain.MessageRoleUser,
			Content:        "Quero conhecer os imóveis disponíveis.",
		}

		result, err := repository.Create(
			ctx,
			message,
		)
		if err != nil {
			t.Fatalf(
				"create message: %v",
				err,
			)
		}

		if result == nil {
			t.Fatal("expected message, got nil")
		}

		if result.ID == uuid.Nil {
			t.Fatal("expected generated message ID")
		}

		if result.ConversationID != createdConversation.ID {
			t.Fatalf(
				"expected conversation ID %s, got %s",
				createdConversation.ID,
				result.ConversationID,
			)
		}

		if result.Role != domain.MessageRoleUser {
			t.Fatalf(
				"expected role %s, got %s",
				domain.MessageRoleUser,
				result.Role,
			)
		}

		if result.Content != message.Content {
			t.Fatalf(
				"expected content %q, got %q",
				message.Content,
				result.Content,
			)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		message := &domain.Message{
			ConversationID: createdConversation.ID,
			Role:           domain.MessageRoleAssistant,
			Content:        "Claro! Vou verificar os imóveis disponíveis.",
		}

		created, err := repository.Create(
			ctx,
			message,
		)
		if err != nil {
			t.Fatalf(
				"create message: %v",
				err,
			)
		}

		result, err := repository.GetByID(
			ctx,
			created.ID,
		)
		if err != nil {
			t.Fatalf(
				"get message by id: %v",
				err,
			)
		}

		if result == nil {
			t.Fatal("expected message, got nil")
		}

		if result.ID != created.ID {
			t.Fatalf(
				"expected ID %s, got %s",
				created.ID,
				result.ID,
			)
		}

		if result.ConversationID != createdConversation.ID {
			t.Fatalf(
				"expected conversation ID %s, got %s",
				createdConversation.ID,
				result.ConversationID,
			)
		}

		if result.Role != domain.MessageRoleAssistant {
			t.Fatalf(
				"expected role %s, got %s",
				domain.MessageRoleAssistant,
				result.Role,
			)
		}

		if result.Content != message.Content {
			t.Fatalf(
				"expected content %q, got %q",
				message.Content,
				result.Content,
			)
		}
	})

	t.Run("ListByConversation", func(t *testing.T) {
		first := &domain.Message{
			ConversationID: createdConversation.ID,
			Role:           domain.MessageRoleUser,
			Content:        "Primeira mensagem.",
		}

		second := &domain.Message{
			ConversationID: createdConversation.ID,
			Role:           domain.MessageRoleAssistant,
			Content:        "Segunda mensagem.",
		}

		_, err := repository.Create(ctx, first)
		if err != nil {
			t.Fatalf(
				"create first message: %v",
				err,
			)
		}

		_, err = repository.Create(ctx, second)
		if err != nil {
			t.Fatalf(
				"create second message: %v",
				err,
			)
		}

		results, err := repository.ListByConversation(
			ctx,
			createdConversation.ID,
		)
		if err != nil {
			t.Fatalf(
				"list messages by conversation: %v",
				err,
			)
		}

		if len(results) < 3 {
			t.Fatalf(
				"expected at least 3 messages, got %d",
				len(results),
			)
		}

		for _, result := range results {
			if result.ConversationID != createdConversation.ID {
				t.Fatalf(
					"expected conversation ID %s, got %s",
					createdConversation.ID,
					result.ConversationID,
				)
			}
		}
	})

	t.Run("List", func(t *testing.T) {
		results, err := repository.List(
			ctx,
			100,
			0,
		)
		if err != nil {
			t.Fatalf(
				"list messages: %v",
				err,
			)
		}

		if len(results) == 0 {
			t.Fatal(
				"expected at least one message",
			)
		}
	})

	t.Run("Count", func(t *testing.T) {
		count, err := repository.Count(ctx)
		if err != nil {
			t.Fatalf(
				"count messages: %v",
				err,
			)
		}

		if count < 4 {
			t.Fatalf(
				"expected at least 4 messages, got %d",
				count,
			)
		}
	})

	t.Run("CountByConversation", func(t *testing.T) {
		count, err := repository.CountByConversation(
			ctx,
			createdConversation.ID,
		)
		if err != nil {
			t.Fatalf(
				"count messages by conversation: %v",
				err,
			)
		}

		if count < 4 {
			t.Fatalf(
				"expected at least 4 messages for conversation, got %d",
				count,
			)
		}
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		_, err := repository.GetByID(
			ctx,
			uuid.New(),
		)

		if err == nil {
			t.Fatal(
				"expected error when message does not exist",
			)
		}
	})

	t.Run("ListByConversation_Empty", func(t *testing.T) {
		results, err := repository.ListByConversation(
			ctx,
			uuid.New(),
		)
		if err != nil {
			t.Fatalf(
				"list messages by unknown conversation: %v",
				err,
			)
		}

		if len(results) != 0 {
			t.Fatalf(
				"expected no messages, got %d",
				len(results),
			)
		}
	})
}
