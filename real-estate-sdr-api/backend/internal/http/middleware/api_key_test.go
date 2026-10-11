package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKey(t *testing.T) {
	const expectedKey = "test-secret"

	tests := []struct {
		name           string
		providedKey    string
		path           string
		expectedStatus int
	}{
		{
			name:           "missing key",
			path:           "/leads",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid key",
			providedKey:    "wrong-secret",
			path:           "/leads",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid key",
			providedKey:    expectedKey,
			path:           "/leads",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "health without key",
			path:           "/health",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "ready without key",
			path:           "/ready",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				},
			)

			handler := APIKey(expectedKey, next)

			req := httptest.NewRequest(
				http.MethodGet,
				tt.path,
				nil,
			)

			if tt.providedKey != "" {
				req.Header.Set(
					"X-API-Key",
					tt.providedKey,
				)
			}

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf(
					"expected status %d, got %d",
					tt.expectedStatus,
					recorder.Code,
				)
			}
		})
	}
}
