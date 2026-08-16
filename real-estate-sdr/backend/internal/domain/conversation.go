package domain

import (
	"time"

	"github.com/google/uuid"
)

type ConversationStatus string

const (
	ConversationStatusActive ConversationStatus = "ACTIVE"
	ConversationStatusClosed ConversationStatus = "CLOSED"
)

type Conversation struct {
	ID        uuid.UUID          `json:"id"`
	LeadID    uuid.UUID          `json:"lead_id"`
	Status    ConversationStatus `json:"status"`
	StartedAt time.Time          `json:"started_at"`
	EndedAt   *time.Time         `json:"ended_at,omitempty"`
}
