package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MessageRoleUser      = "USER"
	MessageRoleAssistant = "ASSISTANT"
	MessageRoleSystem    = "SYSTEM"
)

var (
	ErrNilMessage            = errors.New("message is required")
	ErrMessageConversationID = errors.New("message conversation is required")
	ErrMessageRole           = errors.New("message role is required")
	ErrInvalidMessageRole    = errors.New("invalid message role")
	ErrMessageContent        = errors.New("message content is required")
)

type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	Role           string
	Content        string
	CreatedAt      time.Time
}

func (m *Message) Validate() error {
	if m == nil {
		return ErrNilMessage
	}

	if m.ConversationID == uuid.Nil {
		return ErrMessageConversationID
	}

	m.Role = strings.ToUpper(strings.TrimSpace(m.Role))

	if m.Role == "" {
		return ErrMessageRole
	}

	if !IsValidMessageRole(m.Role) {
		return ErrInvalidMessageRole
	}

	m.Content = strings.TrimSpace(m.Content)

	if m.Content == "" {
		return ErrMessageContent
	}

	return nil
}

func IsValidMessageRole(role string) bool {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case MessageRoleUser,
		MessageRoleAssistant,
		MessageRoleSystem:
		return true
	default:
		return false
	}
}
