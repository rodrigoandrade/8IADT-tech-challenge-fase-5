package postgres

import (
	"context"
	"os"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
	"testing"
	"time"

	"real-estate-sdr/internal/domain"
)

func TestLeadRepository(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to create postgres pool: %v", err)
	}
	defer pool.Close()

	queries := generated.New(pool)
	repository := NewLeadRepository(queries)

	lead := &domain.Lead{
		Name:   "Lead Teste",
		Email:  "lead.teste@example.com",
		Phone:  "11999999999",
		Status: domain.LeadStatusNew,
	}

	// ---------------------------------------------------------
	// CREATE
	// ---------------------------------------------------------

	createdLead, err := repository.Create(ctx, lead)
	if err != nil {
		t.Fatalf("failed to create lead: %v", err)
	}

	if createdLead == nil {
		t.Fatal("created lead should not be nil")
	}

	if createdLead.ID == (createdLead.ID) {
		// ID foi gerado pelo PostgreSQL.
		if createdLead.ID.String() == "00000000-0000-0000-0000-000000000000" {
			t.Fatal("created lead should have a generated ID")
		}
	}

	if createdLead.Name != lead.Name {
		t.Errorf(
			"expected name %q, got %q",
			lead.Name,
			createdLead.Name,
		)
	}

	if createdLead.Email != lead.Email {
		t.Errorf(
			"expected email %q, got %q",
			lead.Email,
			createdLead.Email,
		)
	}

	if createdLead.Phone != lead.Phone {
		t.Errorf(
			"expected phone %q, got %q",
			lead.Phone,
			createdLead.Phone,
		)
	}

	if createdLead.Status != domain.LeadStatusNew {
		t.Errorf(
			"expected status %q, got %q",
			domain.LeadStatusNew,
			createdLead.Status,
		)
	}

	if createdLead.CreatedAt.IsZero() {
		t.Error("created_at should not be zero")
	}

	if createdLead.UpdatedAt.IsZero() {
		t.Error("updated_at should not be zero")
	}

	// Limpeza do registro criado pelo teste.
	defer func() {
		_, err := pool.Exec(
			context.Background(),
			"DELETE FROM leads WHERE id = $1",
			createdLead.ID,
		)

		if err != nil {
			t.Errorf("failed to cleanup test lead: %v", err)
		}
	}()

	// ---------------------------------------------------------
	// GET BY ID
	// ---------------------------------------------------------

	foundLead, err := repository.GetByID(ctx, createdLead.ID)
	if err != nil {
		t.Fatalf("failed to get lead by ID: %v", err)
	}

	if foundLead.ID != createdLead.ID {
		t.Errorf(
			"expected ID %s, got %s",
			createdLead.ID,
			foundLead.ID,
		)
	}

	if foundLead.Name != createdLead.Name {
		t.Errorf(
			"expected name %q, got %q",
			createdLead.Name,
			foundLead.Name,
		)
	}

	if foundLead.Email != createdLead.Email {
		t.Errorf(
			"expected email %q, got %q",
			createdLead.Email,
			foundLead.Email,
		)
	}

	// ---------------------------------------------------------
	// UPDATE
	// ---------------------------------------------------------

	foundLead.Name = "Lead Teste Atualizado"
	foundLead.Email = "lead.atualizado@example.com"
	foundLead.Phone = "11888888888"
	foundLead.Status = domain.LeadStatusContacted

	updatedLead, err := repository.Update(ctx, foundLead)
	if err != nil {
		t.Fatalf("failed to update lead: %v", err)
	}

	if updatedLead.Name != "Lead Teste Atualizado" {
		t.Errorf(
			"expected updated name %q, got %q",
			"Lead Teste Atualizado",
			updatedLead.Name,
		)
	}

	if updatedLead.Email != "lead.atualizado@example.com" {
		t.Errorf(
			"expected updated email %q, got %q",
			"lead.atualizado@example.com",
			updatedLead.Email,
		)
	}

	if updatedLead.Phone != "11888888888" {
		t.Errorf(
			"expected updated phone %q, got %q",
			"11888888888",
			updatedLead.Phone,
		)
	}

	if updatedLead.Status != domain.LeadStatusContacted {
		t.Errorf(
			"expected status %q, got %q",
			domain.LeadStatusContacted,
			updatedLead.Status,
		)
	}

	// ---------------------------------------------------------
	// UPDATE STATUS
	// ---------------------------------------------------------

	statusLead, err := repository.UpdateStatus(
		ctx,
		createdLead.ID,
		domain.LeadStatusQualified,
	)
	if err != nil {
		t.Fatalf("failed to update lead status: %v", err)
	}

	if statusLead.Status != domain.LeadStatusQualified {
		t.Errorf(
			"expected status %q, got %q",
			domain.LeadStatusQualified,
			statusLead.Status,
		)
	}

	// ---------------------------------------------------------
	// LIST
	// ---------------------------------------------------------

	leads, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("failed to list leads: %v", err)
	}

	found := false

	for _, item := range leads {
		if item.ID == createdLead.ID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf(
			"created lead %s was not found in lead list",
			createdLead.ID,
		)
	}

	// ---------------------------------------------------------
	// COUNT
	// ---------------------------------------------------------

	count, err := repository.Count(ctx)
	if err != nil {
		t.Fatalf("failed to count leads: %v", err)
	}

	if count < 1 {
		t.Errorf(
			"expected at least one lead, got %d",
			count,
		)
	}
}
