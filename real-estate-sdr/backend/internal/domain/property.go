package domain

import (
	"time"

	"github.com/google/uuid"
)

type PropertyType string

const (
	PropertyTypeApartment  PropertyType = "APARTMENT"
	PropertyTypeHouse      PropertyType = "HOUSE"
	PropertyTypeCommercial PropertyType = "COMMERCIAL"
	PropertyTypeLand       PropertyType = "LAND"
)

type TransactionType string

const (
	TransactionTypeSale TransactionType = "SALE"
	TransactionTypeRent TransactionType = "RENT"
)

type Property struct {
	ID              uuid.UUID       `json:"id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	PropertyType    PropertyType    `json:"property_type"`
	TransactionType TransactionType `json:"transaction_type"`
	Price           float64         `json:"price"`
	Bedrooms        int             `json:"bedrooms"`
	Bathrooms       int             `json:"bathrooms"`
	ParkingSpaces   int             `json:"parking_spaces"`
	Area            float64         `json:"area"`
	City            string          `json:"city"`
	Neighborhood    string          `json:"neighborhood"`
	Address         string          `json:"address"`
	Available       bool            `json:"available"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
