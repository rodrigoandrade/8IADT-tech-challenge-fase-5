package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"real-estate-sdr/internal/domain"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(
	baseURL string,
	httpClient *http.Client,
) *HTTPClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &HTTPClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
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

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call AI service: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"AI service returned status %d",
			response.StatusCode,
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
