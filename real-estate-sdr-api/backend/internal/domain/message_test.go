package domain

import (
	"testing"

	"github.com/google/uuid"
)

func validMessage() *Message {
	return &Message{
		ID:             uuid.New(),
		ConversationID: uuid.New(),
		Role:           MessageRoleUser,
		Content:        "Quero comprar um apartamento.",
	}
}

func TestMessage_Validate(t *testing.T) {
	message := validMessage()

	if err := message.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMessage_Validate_NilMessage(t *testing.T) {
	var message *Message

	err := message.Validate()

	if err != ErrNilMessage {
		t.Fatalf(
			"expected ErrNilMessage, got %v",
			err,
		)
	}
}

func TestMessage_Validate_MissingConversation(t *testing.T) {
	message := validMessage()
	message.ConversationID = uuid.Nil

	err := message.Validate()

	if err != ErrMessageConversationID {
		t.Fatalf(
			"expected ErrMessageConversationID, got %v",
			err,
		)
	}
}

func TestMessage_Validate_MissingRole(t *testing.T) {
	message := validMessage()
	message.Role = ""

	err := message.Validate()

	if err != ErrMessageRole {
		t.Fatalf(
			"expected ErrMessageRole, got %v",
			err,
		)
	}
}

func TestMessage_Validate_InvalidRole(t *testing.T) {
	message := validMessage()
	message.Role = "INVALID"

	err := message.Validate()

	if err != ErrInvalidMessageRole {
		t.Fatalf(
			"expected ErrInvalidMessageRole, got %v",
			err,
		)
	}
}

func TestMessage_Validate_MissingContent(t *testing.T) {
	message := validMessage()
	message.Content = ""

	err := message.Validate()

	if err != ErrMessageContent {
		t.Fatalf(
			"expected ErrMessageContent, got %v",
			err,
		)
	}
}

func TestMessage_Validate_WhitespaceContent(t *testing.T) {
	message := validMessage()
	message.Content = "   "

	err := message.Validate()

	if err != ErrMessageContent {
		t.Fatalf(
			"expected ErrMessageContent, got %v",
			err,
		)
	}
}

func TestMessage_Validate_NormalizesRole(t *testing.T) {
	message := validMessage()
	message.Role = "  user  "

	if err := message.Validate(); err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if message.Role != MessageRoleUser {
		t.Fatalf(
			"expected role %q, got %q",
			MessageRoleUser,
			message.Role,
		)
	}
}

func TestMessage_Validate_NormalizesContent(t *testing.T) {
	message := validMessage()
	message.Content = "  Quero comprar um apartamento.  "

	if err := message.Validate(); err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	expected := "Quero comprar um apartamento."

	if message.Content != expected {
		t.Fatalf(
			"expected content %q, got %q",
			expected,
			message.Content,
		)
	}
}

func TestMessage_Validate_User(t *testing.T) {
	message := validMessage()
	message.Role = MessageRoleUser

	if err := message.Validate(); err != nil {
		t.Fatalf(
			"expected no error for USER, got %v",
			err,
		)
	}
}

func TestMessage_Validate_Assistant(t *testing.T) {
	message := validMessage()
	message.Role = MessageRoleAssistant

	if err := message.Validate(); err != nil {
		t.Fatalf(
			"expected no error for ASSISTANT, got %v",
			err,
		)
	}
}

func TestMessage_Validate_System(t *testing.T) {
	message := validMessage()
	message.Role = MessageRoleSystem

	if err := message.Validate(); err != nil {
		t.Fatalf(
			"expected no error for SYSTEM, got %v",
			err,
		)
	}
}

func TestIsValidMessageRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{
			name: "USER",
			role: MessageRoleUser,
			want: true,
		},
		{
			name: "ASSISTANT",
			role: MessageRoleAssistant,
			want: true,
		},
		{
			name: "SYSTEM",
			role: MessageRoleSystem,
			want: true,
		},
		{
			name: "lowercase user",
			role: "user",
			want: true,
		},
		{
			name: "lowercase assistant",
			role: "assistant",
			want: true,
		},
		{
			name: "lowercase system",
			role: "system",
			want: true,
		},
		{
			name: "invalid",
			role: "INVALID",
			want: false,
		},
		{
			name: "empty",
			role: "",
			want: false,
		},
		{
			name: "whitespace",
			role: "   ",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidMessageRole(tt.role)

			if got != tt.want {
				t.Fatalf(
					"expected %v, got %v",
					tt.want,
					got,
				)
			}
		})
	}
}
