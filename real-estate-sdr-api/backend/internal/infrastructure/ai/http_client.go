package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"real-estate-sdr/internal/domain"
	"real-estate-sdr/internal/http/middleware"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

func NewHTTPClient(
	baseURL string,
	httpClient *http.Client,
	apiKeys ...string,
) *HTTPClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	var apiKey string
	if len(apiKeys) > 0 {
		apiKey = apiKeys[0]
	}

	return &HTTPClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
		apiKey:     apiKey,
	}
}

type generateResponseRequest struct {
	ID       string                   `json:"id"`
	LeadID   string                   `json:"lead_id"`
	Messages []generateMessageRequest `json:"messages"`
}

type generateMessageRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *HTTPClient) GenerateResponse(
	ctx context.Context,
	conversation *domain.Conversation,
	messages []*domain.Message,
) (*domain.AIResponse, error) {
	if conversation == nil {
		return nil, fmt.Errorf("conversation is required")
	}

	if strings.TrimSpace(c.baseURL) == "" {
		return nil, fmt.Errorf("AI service URL is not configured")
	}

	if c.apiKey == "" {
		return nil, fmt.Errorf(
			"AI service API key is not configured",
		)
	}

	request := generateResponseRequest{
		ID:     conversation.ID.String(),
		LeadID: conversation.LeadID.String(),
	}

	for _, message := range messages {
		if message == nil {
			continue
		}

		request.Messages = append(
			request.Messages,
			generateMessageRequest{
				Role:    toAIRole(message.Role),
				Content: message.Content,
			},
		)
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal AI request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/ai/generate-response",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create AI request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	req.Header.Set(
		"X-Internal-API-Key",
		c.apiKey,
	)

	correlationID := middleware.GetCorrelationID(ctx)

	if correlationID != "" {
		req.Header.Set(
			"X-Correlation-ID",
			correlationID,
		)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call AI service: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		errorBody, readErr := io.ReadAll(
			io.LimitReader(response.Body, 4096),
		)

		if readErr != nil {
			return nil, fmt.Errorf(
				"AI service returned status %d; failed to read error body: %w",
				response.StatusCode,
				readErr,
			)
		}

		detail := strings.TrimSpace(string(errorBody))

		if detail == "" {
			return nil, fmt.Errorf(
				"AI service returned status %d with empty error body",
				response.StatusCode,
			)
		}

		return nil, fmt.Errorf(
			"AI service returned status %d: %s",
			response.StatusCode,
			detail,
		)
	}

	var aiResponse domain.AIResponse

	if err := json.NewDecoder(response.Body).Decode(&aiResponse); err != nil {
		return nil, fmt.Errorf(
			"decode AI response: %w",
			err,
		)
	}

	if strings.TrimSpace(aiResponse.Message) == "" {
		return nil, fmt.Errorf(
			"AI service returned empty message",
		)
	}

	return &aiResponse, nil
}

func toAIRole(role string) string {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case domain.MessageRoleUser:
		return "lead"

	case domain.MessageRoleAssistant:
		return "assistant"

	case domain.MessageRoleSystem:
		return "system"

	default:
		return strings.ToLower(strings.TrimSpace(role))
	}
}

var _ interface {
	GenerateResponse(
		context.Context,
		*domain.Conversation,
		[]*domain.Message,
	) (*domain.AIResponse, error)
} = (*HTTPClient)(nil)
