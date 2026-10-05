package postgres

import (
	"context"
	"os"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"real-estate-sdr/internal/domain"
)

func TestPropertyRepository(t *testing.T) {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not configured")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to create PostgreSQL pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to ping PostgreSQL: %v", err)
	}

	queries := generated.New(pool)
	repository := NewPropertyRepository(queries)

	property := &domain.Property{
		Title:         "Apartamento Teste",
		Description:   "Apartamento criado pelo teste de integração",
		Type:          domain.PropertyTypeApartment,
		Transaction:   domain.TransactionTypeSale,
		Price:         450000.00,
		Bedrooms:      3,
		Bathrooms:     2,
		ParkingSpaces: 2,
		Area:          120.50,
		City:          "São Paulo",
		Neighborhood:  "Vila Mariana",
		Address:       "Rua Teste, 100",
		Status:        domain.PropertyStatusAvailable,
	}

	// ---------------------------------------------------------
	// CREATE
	// ---------------------------------------------------------

	createdProperty, err := repository.Create(ctx, property)
	if err != nil {
		t.Fatalf("failed to create property: %v", err)
	}

	if createdProperty == nil {
		t.Fatal("expected created property, got nil")
	}

	if createdProperty.ID == uuid.Nil {
		t.Fatal("expected generated property id")
	}

	if createdProperty.Title != property.Title {
		t.Errorf(
			"expected title %q, got %q",
			property.Title,
			createdProperty.Title,
		)
	}

	if createdProperty.Description != property.Description {
		t.Errorf(
			"expected description %q, got %q",
			property.Description,
			createdProperty.Description,
		)
	}

	if createdProperty.Type != property.Type {
		t.Errorf(
			"expected type %q, got %q",
			property.Type,
			createdProperty.Type,
		)
	}

	if createdProperty.Transaction != property.Transaction {
		t.Errorf(
			"expected transaction %q, got %q",
			property.Transaction,
			createdProperty.Transaction,
		)
	}

	if createdProperty.Price != property.Price {
		t.Errorf(
			"expected price %.2f, got %.2f",
			property.Price,
			createdProperty.Price,
		)
	}

	if createdProperty.Bedrooms != property.Bedrooms {
		t.Errorf(
			"expected bedrooms %d, got %d",
			property.Bedrooms,
			createdProperty.Bedrooms,
		)
	}

	if createdProperty.Bathrooms != property.Bathrooms {
		t.Errorf(
			"expected bathrooms %d, got %d",
			property.Bathrooms,
			createdProperty.Bathrooms,
		)
	}

	if createdProperty.ParkingSpaces != property.ParkingSpaces {
		t.Errorf(
			"expected parking spaces %d, got %d",
			property.ParkingSpaces,
			createdProperty.ParkingSpaces,
		)
	}

	if createdProperty.Area != property.Area {
		t.Errorf(
			"expected area %.2f, got %.2f",
			property.Area,
			createdProperty.Area,
		)
	}

	if createdProperty.City != property.City {
		t.Errorf(
			"expected city %q, got %q",
			property.City,
			createdProperty.City,
		)
	}

	if createdProperty.Neighborhood != property.Neighborhood {
		t.Errorf(
			"expected neighborhood %q, got %q",
			property.Neighborhood,
			createdProperty.Neighborhood,
		)
	}

	if createdProperty.Address != property.Address {
		t.Errorf(
			"expected address %q, got %q",
			property.Address,
			createdProperty.Address,
		)
	}

	if createdProperty.Status != domain.PropertyStatusAvailable {
		t.Errorf(
			"expected status %q, got %q",
			domain.PropertyStatusAvailable,
			createdProperty.Status,
		)
	}

	// ---------------------------------------------------------
	// GET BY ID
	// ---------------------------------------------------------

	foundProperty, err := repository.GetByID(
		ctx,
		createdProperty.ID,
	)
	if err != nil {
		t.Fatalf("failed to get property: %v", err)
	}

	if foundProperty == nil {
		t.Fatal("expected found property, got nil")
	}

	if foundProperty.ID != createdProperty.ID {
		t.Errorf(
			"expected id %s, got %s",
			createdProperty.ID,
			foundProperty.ID,
		)
	}

	if foundProperty.Title != createdProperty.Title {
		t.Errorf(
			"expected title %q, got %q",
			createdProperty.Title,
			foundProperty.Title,
		)
	}

	if foundProperty.Price != createdProperty.Price {
		t.Errorf(
			"expected price %.2f, got %.2f",
			createdProperty.Price,
			foundProperty.Price,
		)
	}

	// ---------------------------------------------------------
	// LIST
	// ---------------------------------------------------------

	properties, err := repository.List(
		ctx,
		100,
		0,
	)
	if err != nil {
		t.Fatalf("failed to list properties: %v", err)
	}

	if len(properties) == 0 {
		t.Fatal("expected at least one property")
	}

	var listedProperty *domain.Property

	for _, item := range properties {
		if item.ID == createdProperty.ID {
			listedProperty = item
			break
		}
	}

	if listedProperty == nil {
		t.Fatalf(
			"created property %s was not found in list",
			createdProperty.ID,
		)
	}

	// ---------------------------------------------------------
	// UPDATE
	// ---------------------------------------------------------

	foundProperty.Title = "Apartamento Teste Atualizado"
	foundProperty.Description = "Descrição atualizada"
	foundProperty.Price = 475000.00
	foundProperty.Bedrooms = 4
	foundProperty.Bathrooms = 3
	foundProperty.ParkingSpaces = 2
	foundProperty.Area = 135.75
	foundProperty.City = "São Paulo"
	foundProperty.Neighborhood = "Moema"
	foundProperty.Address = "Rua Atualizada, 200"
	foundProperty.Status = domain.PropertyStatusAvailable

	updatedProperty, err := repository.Update(
		ctx,
		foundProperty,
	)
	if err != nil {
		t.Fatalf("failed to update property: %v", err)
	}

	if updatedProperty.Title != "Apartamento Teste Atualizado" {
		t.Errorf(
			"expected updated title %q, got %q",
			"Apartamento Teste Atualizado",
			updatedProperty.Title,
		)
	}

	if updatedProperty.Description != "Descrição atualizada" {
		t.Errorf(
			"expected updated description %q, got %q",
			"Descrição atualizada",
			updatedProperty.Description,
		)
	}

	if updatedProperty.Price != 475000.00 {
		t.Errorf(
			"expected updated price %.2f, got %.2f",
			475000.00,
			updatedProperty.Price,
		)
	}

	if updatedProperty.Bedrooms != 4 {
		t.Errorf(
			"expected updated bedrooms %d, got %d",
			4,
			updatedProperty.Bedrooms,
		)
	}

	if updatedProperty.Bathrooms != 3 {
		t.Errorf(
			"expected updated bathrooms %d, got %d",
			3,
			updatedProperty.Bathrooms,
		)
	}

	if updatedProperty.Area != 135.75 {
		t.Errorf(
			"expected updated area %.2f, got %.2f",
			135.75,
			updatedProperty.Area,
		)
	}

	if updatedProperty.Neighborhood != "Moema" {
		t.Errorf(
			"expected updated neighborhood %q, got %q",
			"Moema",
			updatedProperty.Neighborhood,
		)
	}

	// ---------------------------------------------------------
	// UPDATE STATUS
	// ---------------------------------------------------------

	statusProperty, err := repository.UpdateStatus(
		ctx,
		createdProperty.ID,
		domain.PropertyStatusInactive,
	)
	if err != nil {
		t.Fatalf(
			"failed to update property status: %v",
			err,
		)
	}

	if statusProperty.Status != domain.PropertyStatusInactive {
		t.Errorf(
			"expected status %q, got %q",
			domain.PropertyStatusInactive,
			statusProperty.Status,
		)
	}

	// ---------------------------------------------------------
	// GET BY ID AFTER STATUS UPDATE
	// ---------------------------------------------------------

	finalProperty, err := repository.GetByID(
		ctx,
		createdProperty.ID,
	)
	if err != nil {
		t.Fatalf(
			"failed to get property after status update: %v",
			err,
		)
	}

	if finalProperty.Status != domain.PropertyStatusInactive {
		t.Errorf(
			"expected final status %q, got %q",
			domain.PropertyStatusInactive,
			finalProperty.Status,
		)
	}

	// ---------------------------------------------------------
	// COUNT
	// ---------------------------------------------------------

	count, err := repository.Count(ctx)
	if err != nil {
		t.Fatalf(
			"failed to count properties: %v",
			err,
		)
	}

	if count <= 0 {
		t.Errorf(
			"expected property count > 0, got %d",
			count,
		)
	}

	// ---------------------------------------------------------
	// NOT FOUND
	// ---------------------------------------------------------

	_, err = repository.GetByID(
		ctx,
		uuid.New(),
	)

	if err == nil {
		t.Fatal(
			"expected error when getting nonexistent property",
		)
	}
}
