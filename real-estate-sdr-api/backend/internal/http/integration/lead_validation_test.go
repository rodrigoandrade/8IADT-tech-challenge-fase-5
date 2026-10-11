package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLeadAPI_Create_InvalidJSON(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	request := httptest.NewRequest(
		http.MethodPost,
		"/leads",
		strings.NewReader(`{"name":"JoÃ£o Silva"`),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestLeadAPI_Create_EmptyBody(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	request := httptest.NewRequest(
		http.MethodPost,
		"/leads",
		strings.NewReader(""),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestLeadAPI_Create_InvalidLead(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing name",
			body: `{
				"email": "validation@example.com",
				"phone": "11999999999"
			}`,
		},
		{
			name: "missing email",
			body: `{
				"name": "Validation Test",
				"phone": "11999999999"
			}`,
		},
		{
			name: "missing phone",
			body: `{
				"name": "Validation Test",
				"email": "validation@example.com"
			}`,
		},
		{
			name: "empty name",
			body: `{
				"name": "",
				"email": "validation@example.com",
				"phone": "11999999999"
			}`,
		},
		{
			name: "empty email",
			body: `{
				"name": "Validation Test",
				"email": "",
				"phone": "11999999999"
			}`,
		},
		{
			name: "empty phone",
			body: `{
				"name": "Validation Test",
				"email": "validation@example.com",
				"phone": ""
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := newTestPool(t)
			defer pool.Close()

			router := newTestRouter(t, pool)

			request := httptest.NewRequest(
				http.MethodPost,
				"/leads",
				strings.NewReader(tt.body),
			)

			request.Header.Set(
				"Content-Type",
				"application/json",
			)

			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected status %d, got %d: %s",
					http.StatusBadRequest,
					response.Code,
					response.Body.String(),
				)
			}
		})
	}
}

func TestLeadAPI_GetByID_InvalidID(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	request := httptest.NewRequest(
		http.MethodGet,
		"/leads/not-a-valid-uuid",
		nil,
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestLeadAPI_Update_InvalidID(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	body := `{
		"name": "Updated Lead",
		"email": "updated@example.com",
		"phone": "11988887777",
		"status": "CONTACTED"
	}`

	request := httptest.NewRequest(
		http.MethodPut,
		"/leads/not-a-valid-uuid",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestLeadAPI_Update_InvalidJSON(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	request := httptest.NewRequest(
		http.MethodPut,
		"/leads/11111111-1111-1111-1111-111111111111",
		strings.NewReader(`{"name":"invalid"`),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestLeadAPI_UpdateStatus_InvalidID(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	body := `{
		"status": "QUALIFIED"
	}`

	request := httptest.NewRequest(
		http.MethodPatch,
		"/leads/not-a-valid-uuid/status",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestLeadAPI_UpdateStatus_InvalidStatus(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	body := `{
		"status": "INVALID_STATUS"
	}`

	request := httptest.NewRequest(
		http.MethodPatch,
		"/leads/11111111-1111-1111-1111-111111111111/status",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestLeadAPI_UpdateStatus_InvalidJSON(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/leads/11111111-1111-1111-1111-111111111111/status",
		strings.NewReader(`{"status":`),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestLeadAPI_UpdateStatus_NotFound(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	body := `{
		"status": "QUALIFIED"
	}`

	request := httptest.NewRequest(
		http.MethodPatch,
		"/leads/11111111-1111-1111-1111-111111111111/status",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestLeadAPI_Update_NotFound(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	body := `{
		"name": "Updated Lead",
		"email": "updated@example.com",
		"phone": "11988887777",
		"status": "CONTACTED"
	}`

	request := httptest.NewRequest(
		http.MethodPut,
		"/leads/11111111-1111-1111-1111-111111111111",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}
}

func timestamp() int64 {
	return time.Now().UnixNano()
}
