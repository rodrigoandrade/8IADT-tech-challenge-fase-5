package postgres

import (
	"context"
	"errors"
	"os"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"real-estate-sdr/internal/domain"
)

func TestConversationRepository(t *testing.T) {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not configured")
	}

	pool, err := NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf(
			"failed to connect to PostgreSQL: %v",
			err,
		)
	}
	defer pool.Close()

	queries := generated.New(pool)

	leadRepository := NewLeadRepository(queries)
	conversationRepository := NewConversationRepository(queries)

	// ---------------------------------------------------------
	// CREATE LEAD
	// ---------------------------------------------------------

	lead := &domain.Lead{
		Name:   "Conversation Test Lead",
		Email:  "conversation-test-" + uuid.NewString() + "@example.com",
		Phone:  "11999999999",
		Status: domain.LeadStatusNew,
	}

	createdLead, err := leadRepository.Create(ctx, lead)
	if err != nil {
		t.Fatalf(
			"failed to create lead: %v",
			err,
		)
	}

	defer func() {
		_, err := pool.Exec(
			ctx,
			"DELETE FROM leads WHERE id = $1",
			createdLead.ID,
		)

		if err != nil {
			t.Errorf(
				"failed to cleanup lead %s: %v",
				createdLead.ID,
				err,
			)
		}
	}()

	// ---------------------------------------------------------
	// CREATE CONVERSATION
	// ---------------------------------------------------------

	conversation := &domain.Conversation{
		LeadID:    createdLead.ID,
		Status:    domain.ConversationStatusActive,
		StartedAt: time.Now(),
	}

	createdConversation, err := conversationRepository.Create(
		ctx,
		conversation,
	)
	if err != nil {
		t.Fatalf(
			"failed to create conversation: %v",
			err,
		)
	}

	if createdConversation.ID == uuid.Nil {
		t.Fatal("expected conversation ID to be generated")
	}

	if createdConversation.LeadID != createdLead.ID {
		t.Fatalf(
			"expected lead ID %s, got %s",
			createdLead.ID,
			createdConversation.LeadID,
		)
	}

	if createdConversation.Status != domain.ConversationStatusActive {
		t.Fatalf(
			"expected status %q, got %q",
			domain.ConversationStatusActive,
			createdConversation.Status,
		)
	}

	if createdConversation.EndedAt != nil {
		t.Fatal(
			"expected active conversation to have nil EndedAt",
		)
	}

	if createdConversation.StartedAt.IsZero() {
		t.Fatal("expected StartedAt to be populated")
	}

	conversationID := createdConversation.ID

	defer func() {
		_, err := pool.Exec(
			ctx,
			"DELETE FROM conversations WHERE id = $1",
			conversationID,
		)

		if err != nil {
			t.Errorf(
				"failed to cleanup conversation %s: %v",
				conversationID,
				err,
			)
		}
	}()

	// ---------------------------------------------------------
	// GET BY ID
	// ---------------------------------------------------------

	foundConversation, err := conversationRepository.GetByID(
		ctx,
		conversationID,
	)
	if err != nil {
		t.Fatalf(
			"failed to get conversation: %v",
			err,
		)
	}

	if foundConversation.ID != conversationID {
		t.Fatalf(
			"expected conversation ID %s, got %s",
			conversationID,
			foundConversation.ID,
		)
	}

	if foundConversation.LeadID != createdLead.ID {
		t.Fatalf(
			"expected lead ID %s, got %s",
			createdLead.ID,
			foundConversation.LeadID,
		)
	}

	if foundConversation.Status != domain.ConversationStatusActive {
		t.Fatalf(
			"expected status %q, got %q",
			domain.ConversationStatusActive,
			foundConversation.Status,
		)
	}

	if foundConversation.EndedAt != nil {
		t.Fatal(
			"expected active conversation to have nil EndedAt",
		)
	}

	// ---------------------------------------------------------
	// GET BY LEAD ID
	// ---------------------------------------------------------

	conversations, err := conversationRepository.GetByLeadID(
		ctx,
		createdLead.ID,
	)
	if err != nil {
		t.Fatalf(
			"failed to get conversations by lead: %v",
			err,
		)
	}

	found := false

	for _, item := range conversations {
		if item.ID == conversationID {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf(
			"conversation %s was not found for lead %s",
			conversationID,
			createdLead.ID,
		)
	}

	// ---------------------------------------------------------
	// LIST
	// ---------------------------------------------------------

	list, err := conversationRepository.List(
		ctx,
		10,
		0,
	)
	if err != nil {
		t.Fatalf(
			"failed to list conversations: %v",
			err,
		)
	}

	found = false

	for _, item := range list {
		if item.ID == conversationID {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf(
			"conversation %s was not found in list",
			conversationID,
		)
	}

	// ---------------------------------------------------------
	// COUNT
	// ---------------------------------------------------------

	count, err := conversationRepository.Count(ctx)
	if err != nil {
		t.Fatalf(
			"failed to count conversations: %v",
			err,
		)
	}

	if count <= 0 {
		t.Fatalf(
			"expected conversation count greater than zero, got %d",
			count,
		)
	}

	// ---------------------------------------------------------
	// CLOSE
	// ---------------------------------------------------------

	closedConversation, err := conversationRepository.Close(
		ctx,
		conversationID,
	)
	if err != nil {
		t.Fatalf(
			"failed to close conversation: %v",
			err,
		)
	}

	if closedConversation.ID != conversationID {
		t.Fatalf(
			"expected conversation ID %s, got %s",
			conversationID,
			closedConversation.ID,
		)
	}

	if closedConversation.Status != domain.ConversationStatusClosed {
		t.Fatalf(
			"expected status %q, got %q",
			domain.ConversationStatusClosed,
			closedConversation.Status,
		)
	}

	if closedConversation.EndedAt == nil {
		t.Fatal(
			"expected EndedAt to be populated after closing",
		)
	}

	if closedConversation.EndedAt.Before(
		closedConversation.StartedAt,
	) {
		t.Fatalf(
			"expected EndedAt >= StartedAt, got started_at=%v ended_at=%v",
			closedConversation.StartedAt,
			closedConversation.EndedAt,
		)
	}

	// ---------------------------------------------------------
	// GET CLOSED CONVERSATION
	// ---------------------------------------------------------

	foundConversation, err = conversationRepository.GetByID(
		ctx,
		conversationID,
	)
	if err != nil {
		t.Fatalf(
			"failed to get closed conversation: %v",
			err,
		)
	}

	if foundConversation.Status != domain.ConversationStatusClosed {
		t.Fatalf(
			"expected status %q, got %q",
			domain.ConversationStatusClosed,
			foundConversation.Status,
		)
	}

	if foundConversation.EndedAt == nil {
		t.Fatal(
			"expected EndedAt to remain populated",
		)
	}

	// ---------------------------------------------------------
	// CLOSE AGAIN
	// ---------------------------------------------------------

	_, err = conversationRepository.Close(
		ctx,
		conversationID,
	)

	if err == nil {
		t.Fatal(
			"expected error when closing an already closed conversation",
		)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Logf(
			"expected wrapped pgx.ErrNoRows, got: %v",
			err,
		)
	}
}
