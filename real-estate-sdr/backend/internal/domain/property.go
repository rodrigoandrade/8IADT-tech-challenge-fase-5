package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PropertyType string

const (
	PropertyTypeHouse       PropertyType = "HOUSE"
	PropertyTypeApartment   PropertyType = "APARTMENT"
	PropertyTypeCondominium PropertyType = "CONDOMINIUM"
	PropertyTypeCommercial  PropertyType = "COMMERCIAL"
	PropertyTypeLand        PropertyType = "LAND"
)

type TransactionType string

const (
	TransactionTypeSale TransactionType = "SALE"
	TransactionTypeRent TransactionType = "RENT"
)

type PropertyStatus string

const (
	PropertyStatusAvailable PropertyStatus = "AVAILABLE"
	PropertyStatusReserved  PropertyStatus = "RESERVED"
	PropertyStatusSold      PropertyStatus = "SOLD"
	PropertyStatusRented    PropertyStatus = "RENTED"
	PropertyStatusInactive  PropertyStatus = "INACTIVE"
)

var (
	ErrInvalidProperty        = errors.New("invalid property")
	ErrInvalidPropertyType    = errors.New("invalid property type")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
	ErrInvalidPropertyStatus  = errors.New("invalid property status")
)

type Property struct {
	ID            uuid.UUID       `json:"id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	Type          PropertyType    `json:"type"`
	Transaction   TransactionType `json:"transaction_type"`
	Price         float64         `json:"price"`
	Bedrooms      int             `json:"bedrooms"`
	Bathrooms     int             `json:"bathrooms"`
	ParkingSpaces int             `json:"parking_spaces"`
	Area          float64         `json:"area"`
	Address       string          `json:"address"`
	Neighborhood  string          `json:"neighborhood"`
	City          string          `json:"city"`
	State         string          `json:"state"`
	Status        PropertyStatus  `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func (p *Property) Validate() error {
	if p == nil {
		return ErrInvalidProperty
	}

	if strings.TrimSpace(p.Title) == "" {
		return errors.New("property title is required")
	}

	if p.Type == "" {
		return errors.New("property type is required")
	}

	if !IsValidPropertyType(p.Type) {
		return ErrInvalidPropertyType
	}

	if p.Transaction == "" {
		return errors.New("transaction type is required")
	}

	if !IsValidTransactionType(p.Transaction) {
		return ErrInvalidTransactionType
	}

	if p.Price <= 0 {
		return errors.New("property price must be greater than zero")
	}

	if p.Area <= 0 {
		return errors.New("property area must be greater than zero")
	}

	if p.Bedrooms < 0 {
		return errors.New("bedrooms cannot be negative")
	}

	if p.Bathrooms < 0 {
		return errors.New("bathrooms cannot be negative")
	}

	if p.ParkingSpaces < 0 {
		return errors.New("parking spaces cannot be negative")
	}

	if strings.TrimSpace(p.City) == "" {
		return errors.New("property city is required")
	}

	if p.Status == "" {
		p.Status = PropertyStatusAvailable
	}

	if !IsValidPropertyStatus(p.Status) {
		return ErrInvalidPropertyStatus
	}

	return nil
}

func IsValidPropertyType(propertyType PropertyType) bool {
	switch propertyType {
	case PropertyTypeHouse,
		PropertyTypeApartment,
		PropertyTypeCondominium,
		PropertyTypeCommercial,
		PropertyTypeLand:
		return true
	default:
		return false
	}
}

func IsValidTransactionType(transactionType TransactionType) bool {
	switch transactionType {
	case TransactionTypeSale,
		TransactionTypeRent:
		return true
	default:
		return false
	}
}

func IsValidPropertyStatus(status PropertyStatus) bool {
	switch status {
	case PropertyStatusAvailable,
		PropertyStatusReserved,
		PropertyStatusSold,
		PropertyStatusRented,
		PropertyStatusInactive:
		return true
	default:
		return false
	}
}
