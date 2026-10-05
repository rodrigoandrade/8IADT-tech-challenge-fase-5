package domain

import (
	"time"

	"github.com/google/uuid"
)

type LeadStatus string

const (
	LeadStatusNew       LeadStatus = "NEW"
	LeadStatusContacted LeadStatus = "CONTACTED"
	LeadStatusQualified LeadStatus = "QUALIFIED"
	LeadStatusConverted LeadStatus = "CONVERTED"
	LeadStatusLost      LeadStatus = "LOST"
)

type Lead struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Status    LeadStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
