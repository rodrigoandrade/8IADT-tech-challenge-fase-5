package domain

import (
	"testing"

	"github.com/google/uuid"
)

func validProperty() *Property {
	return &Property{
		ID:            uuid.New(),
		Title:         "Apartamento 3 quartos",
		Description:   "Apartamento moderno",
		Type:          PropertyTypeApartment,
		Transaction:   TransactionTypeSale,
		Price:         450000,
		Bedrooms:      3,
		Bathrooms:     2,
		ParkingSpaces: 2,
		Area:          120,
		Address:       "Rua das Flores, 100",
		Neighborhood:  "Centro",
		City:          "São Paulo",
		Status:        PropertyStatusAvailable,
	}
}

func TestProperty_Validate(t *testing.T) {
	property := validProperty()

	if err := property.Validate(); err != nil {
		t.Fatalf(
			"expected valid property, got error: %v",
			err,
		)
	}
}

func TestProperty_Validate_NilProperty(t *testing.T) {
	var property *Property

	err := property.Validate()

	if err != ErrInvalidProperty {
		t.Fatalf(
			"expected ErrInvalidProperty, got %v",
			err,
		)
	}
}

func TestProperty_Validate_DefaultStatus(t *testing.T) {
	property := validProperty()
	property.Status = ""

	if err := property.Validate(); err != nil {
		t.Fatalf(
			"expected valid property, got error: %v",
			err,
		)
	}

	if property.Status != PropertyStatusAvailable {
		t.Fatalf(
			"expected default status %q, got %q",
			PropertyStatusAvailable,
			property.Status,
		)
	}
}

func TestProperty_Validate_InvalidType(t *testing.T) {
	property := validProperty()
	property.Type = PropertyType("INVALID")

	err := property.Validate()

	if err != ErrInvalidPropertyType {
		t.Fatalf(
			"expected ErrInvalidPropertyType, got %v",
			err,
		)
	}
}

func TestProperty_Validate_InvalidTransactionType(t *testing.T) {
	property := validProperty()
	property.Transaction = TransactionType("INVALID")

	err := property.Validate()

	if err != ErrInvalidTransactionType {
		t.Fatalf(
			"expected ErrInvalidTransactionType, got %v",
			err,
		)
	}
}

func TestProperty_Validate_InvalidStatus(t *testing.T) {
	property := validProperty()
	property.Status = PropertyStatus("INVALID")

	err := property.Validate()

	if err != ErrInvalidPropertyStatus {
		t.Fatalf(
			"expected ErrInvalidPropertyStatus, got %v",
			err,
		)
	}
}

func TestProperty_Validate_InvalidPrice(t *testing.T) {
	property := validProperty()
	property.Price = 0

	if err := property.Validate(); err == nil {
		t.Fatal("expected error for invalid price")
	}
}

func TestProperty_Validate_NegativePrice(t *testing.T) {
	property := validProperty()
	property.Price = -1

	if err := property.Validate(); err == nil {
		t.Fatal("expected error for negative price")
	}
}

func TestProperty_Validate_InvalidArea(t *testing.T) {
	property := validProperty()
	property.Area = 0

	if err := property.Validate(); err == nil {
		t.Fatal("expected error for invalid area")
	}
}

func TestProperty_Validate_NegativeBedrooms(t *testing.T) {
	property := validProperty()
	property.Bedrooms = -1

	if err := property.Validate(); err == nil {
		t.Fatal("expected error for negative bedrooms")
	}
}

func TestProperty_Validate_NegativeBathrooms(t *testing.T) {
	property := validProperty()
	property.Bathrooms = -1

	if err := property.Validate(); err == nil {
		t.Fatal("expected error for negative bathrooms")
	}
}

func TestProperty_Validate_NegativeParkingSpaces(t *testing.T) {
	property := validProperty()
	property.ParkingSpaces = -1

	if err := property.Validate(); err == nil {
		t.Fatal("expected error for negative parking spaces")
	}
}

func TestProperty_Validate_RequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Property)
		wantErr string
	}{
		{
			name: "missing title",
			mutate: func(p *Property) {
				p.Title = ""
			},
			wantErr: "property title is required",
		},
		{
			name: "missing type",
			mutate: func(p *Property) {
				p.Type = ""
			},
			wantErr: "property type is required",
		},
		{
			name: "missing transaction",
			mutate: func(p *Property) {
				p.Transaction = ""
			},
			wantErr: "transaction type is required",
		},
		{
			name: "missing city",
			mutate: func(p *Property) {
				p.City = ""
			},
			wantErr: "property city is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property := validProperty()
			tt.mutate(property)

			err := property.Validate()

			if err == nil {
				t.Fatalf(
					"expected error %q, got nil",
					tt.wantErr,
				)
			}

			if err.Error() != tt.wantErr {
				t.Fatalf(
					"expected error %q, got %q",
					tt.wantErr,
					err.Error(),
				)
			}
		})
	}
}

func TestIsValidPropertyType(t *testing.T) {
	tests := []struct {
		propertyType PropertyType
		expected     bool
	}{
		{
			propertyType: PropertyTypeHouse,
			expected:     true,
		},
		{
			propertyType: PropertyTypeApartment,
			expected:     true,
		},
		{
			propertyType: PropertyTypeCondominium,
			expected:     true,
		},
		{
			propertyType: PropertyTypeCommercial,
			expected:     true,
		},
		{
			propertyType: PropertyTypeLand,
			expected:     true,
		},
		{
			propertyType: PropertyType("INVALID"),
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.propertyType), func(t *testing.T) {
			got := IsValidPropertyType(tt.propertyType)

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

func TestIsValidTransactionType(t *testing.T) {
	tests := []struct {
		transactionType TransactionType
		expected        bool
	}{
		{
			transactionType: TransactionTypeSale,
			expected:        true,
		},
		{
			transactionType: TransactionTypeRent,
			expected:        true,
		},
		{
			transactionType: TransactionType("INVALID"),
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.transactionType), func(t *testing.T) {
			got := IsValidTransactionType(tt.transactionType)

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

func TestIsValidPropertyStatus(t *testing.T) {
	tests := []struct {
		status   PropertyStatus
		expected bool
	}{
		{
			status:   PropertyStatusAvailable,
			expected: true,
		},
		{
			status:   PropertyStatusReserved,
			expected: true,
		},
		{
			status:   PropertyStatusSold,
			expected: true,
		},
		{
			status:   PropertyStatusRented,
			expected: true,
		},
		{
			status:   PropertyStatusInactive,
			expected: true,
		},
		{
			status:   PropertyStatus("INVALID"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			got := IsValidPropertyStatus(tt.status)

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
