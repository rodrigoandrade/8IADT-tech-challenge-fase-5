package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func validConversation() *Conversation {
	return &Conversation{
		ID:        uuid.New(),
		LeadID:    uuid.New(),
		Status:    ConversationStatusActive,
		StartedAt: time.Now(),
		EndedAt:   nil,
	}
}

func TestConversation_Validate(t *testing.T) {
	conversation := validConversation()

	if err := conversation.Validate(); err != nil {
		t.Fatalf(
			"expected valid conversation, got error: %v",
			err,
		)
	}
}

func TestConversation_Validate_NilConversation(t *testing.T) {
	var conversation *Conversation

	err := conversation.Validate()

	if err != ErrInvalidConversation {
		t.Fatalf(
			"expected ErrInvalidConversation, got %v",
			err,
		)
	}
}

func TestConversation_Validate_DefaultStatus(t *testing.T) {
	conversation := validConversation()
	conversation.Status = ""

	if err := conversation.Validate(); err != nil {
		t.Fatalf(
			"expected valid conversation, got error: %v",
			err,
		)
	}

	if conversation.Status != ConversationStatusActive {
		t.Fatalf(
			"expected default status %q, got %q",
			ConversationStatusActive,
			conversation.Status,
		)
	}

	if conversation.EndedAt != nil {
		t.Fatal("expected active conversation to have nil EndedAt")
	}
}

func TestConversation_Validate_MissingLead(t *testing.T) {
	conversation := validConversation()
	conversation.LeadID = uuid.Nil

	err := conversation.Validate()

	if err != ErrConversationLeadRequired {
		t.Fatalf(
			"expected ErrConversationLeadRequired, got %v",
			err,
		)
	}
}

func TestConversation_Validate_InvalidStatus(t *testing.T) {
	conversation := validConversation()
	conversation.Status = ConversationStatus("INVALID")

	err := conversation.Validate()

	if err != ErrInvalidConversationStatus {
		t.Fatalf(
			"expected ErrInvalidConversationStatus, got %v",
			err,
		)
	}
}

func TestConversation_Validate_ActiveWithEndedAt(t *testing.T) {
	conversation := validConversation()

	now := time.Now()
	conversation.EndedAt = &now

	err := conversation.Validate()

	if err == nil {
		t.Fatal(
			"expected validation error for active conversation with ended_at",
		)
	}

	expected := "active conversation cannot have ended_at"

	if err.Error() != expected {
		t.Fatalf(
			"expected error %q, got %q",
			expected,
			err.Error(),
		)
	}
}

func TestConversation_Validate_ClosedWithoutEndedAt(t *testing.T) {
	conversation := validConversation()
	conversation.Status = ConversationStatusClosed
	conversation.EndedAt = nil

	err := conversation.Validate()

	if err == nil {
		t.Fatal(
			"expected validation error for closed conversation without ended_at",
		)
	}

	expected := "closed conversation must have ended_at"

	if err.Error() != expected {
		t.Fatalf(
			"expected error %q, got %q",
			expected,
			err.Error(),
		)
	}
}

func TestConversation_Validate_ClosedWithEndedAt(t *testing.T) {
	conversation := validConversation()
	conversation.Status = ConversationStatusClosed

	now := time.Now()
	conversation.EndedAt = &now

	if err := conversation.Validate(); err != nil {
		t.Fatalf(
			"expected valid closed conversation, got error: %v",
			err,
		)
	}
}

func TestConversation_Close(t *testing.T) {
	conversation := validConversation()

	before := time.Now()

	if err := conversation.Close(); err != nil {
		t.Fatalf(
			"expected conversation to close successfully, got error: %v",
			err,
		)
	}

	after := time.Now()

	if conversation.Status != ConversationStatusClosed {
		t.Fatalf(
			"expected status %q, got %q",
			ConversationStatusClosed,
			conversation.Status,
		)
	}

	if conversation.EndedAt == nil {
		t.Fatal("expected EndedAt to be set")
	}

	if conversation.EndedAt.Before(before) ||
		conversation.EndedAt.After(after) {
		t.Fatalf(
			"unexpected EndedAt value: %v",
			conversation.EndedAt,
		)
	}
}

func TestConversation_Close_AlreadyClosed(t *testing.T) {
	conversation := validConversation()

	now := time.Now()

	conversation.Status = ConversationStatusClosed
	conversation.EndedAt = &now

	err := conversation.Close()

	if err != ErrConversationAlreadyClosed {
		t.Fatalf(
			"expected ErrConversationAlreadyClosed, got %v",
			err,
		)
	}
}

func TestConversation_Close_InvalidStatus(t *testing.T) {
	conversation := validConversation()
	conversation.Status = ConversationStatus("INVALID")

	err := conversation.Close()

	if err != ErrInvalidConversationStatus {
		t.Fatalf(
			"expected ErrInvalidConversationStatus, got %v",
			err,
		)
	}
}

func TestIsValidConversationStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   ConversationStatus
		expected bool
	}{
		{
			name:     "active",
			status:   ConversationStatusActive,
			expected: true,
		},
		{
			name:     "closed",
			status:   ConversationStatusClosed,
			expected: true,
		},
		{
			name:     "invalid",
			status:   ConversationStatus("INVALID"),
			expected: false,
		},
		{
			name:     "empty",
			status:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidConversationStatus(tt.status)

			if got != tt.expected {
				t.Fatalf(
					"expected %v, got %v",
					tt.expected,
					got,
				)
			}
		})
	}
}
