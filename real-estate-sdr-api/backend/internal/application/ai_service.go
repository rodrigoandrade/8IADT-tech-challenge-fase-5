package application

import (
	"context"

	"real-estate-sdr/internal/domain"
)

type AIService interface {
	GenerateResponse(
		ctx context.Context,
		conversation *domain.Conversation,
		messages []*domain.Message,
	) (*domain.AIResponse, error)
}
