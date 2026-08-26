package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ConversationStatus string

const (
	ConversationStatusActive ConversationStatus = "ACTIVE"
	ConversationStatusClosed ConversationStatus = "CLOSED"
)

var (
	ErrInvalidConversation       = errors.New("invalid conversation")
	ErrInvalidConversationStatus = errors.New("invalid conversation status")
	ErrConversationLeadRequired  = errors.New("conversation lead is required")
	ErrConversationAlreadyClosed = errors.New("conversation is already closed")
)

type Conversation struct {
	ID        uuid.UUID          `json:"id"`
	LeadID    uuid.UUID          `json:"lead_id"`
	Status    ConversationStatus `json:"status"`
	StartedAt time.Time          `json:"started_at"`
	EndedAt   *time.Time         `json:"ended_at,omitempty"`
}

func (c *Conversation) Validate() error {
	if c == nil {
		return ErrInvalidConversation
	}

	if c.LeadID == uuid.Nil {
		return ErrConversationLeadRequired
	}

	if c.Status == "" {
		c.Status = ConversationStatusActive
	}

	if !IsValidConversationStatus(c.Status) {
		return ErrInvalidConversationStatus
	}

	switch c.Status {
	case ConversationStatusActive:
		if c.EndedAt != nil {
			return errors.New(
				"active conversation cannot have ended_at",
			)
		}

	case ConversationStatusClosed:
		if c.EndedAt == nil {
			return errors.New(
				"closed conversation must have ended_at",
			)
		}
	}

	return nil
}

func (c *Conversation) Close() error {
	if c == nil {
		return ErrInvalidConversation
	}

	if c.Status == ConversationStatusClosed {
		return ErrConversationAlreadyClosed
	}

	if c.Status != ConversationStatusActive {
		return ErrInvalidConversationStatus
	}

	now := time.Now()

	c.Status = ConversationStatusClosed
	c.EndedAt = &now

	return nil
}

func IsValidConversationStatus(
	status ConversationStatus,
) bool {
	switch status {
	case ConversationStatusActive,
		ConversationStatusClosed:
		return true

	default:
		return false
	}
}
