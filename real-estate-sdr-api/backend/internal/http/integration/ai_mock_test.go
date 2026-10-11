package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"real-estate-sdr/internal/infrastructure/ai"
)

const testAIServiceKey = "test-ai-service-key"

func newTestAIClient(t *testing.T) *ai.HTTPClient {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost ||
				r.URL.Path != "/ai/generate-response" {
				http.NotFound(w, r)
				return
			}

			if r.Header.Get("X-Internal-API-Key") != testAIServiceKey {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"message":     "Resposta simulada da IA.",
				"intent":      "property_interest",
				"lead_status": "QUALIFYING",
				"next_action": "ask_budget",
			})
		},
	))

	t.Cleanup(server.Close)

	return ai.NewHTTPClient(server.URL, nil, testAIServiceKey)
}
