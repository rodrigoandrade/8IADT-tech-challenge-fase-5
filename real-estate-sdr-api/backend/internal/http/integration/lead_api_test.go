package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"real-estate-sdr/internal/application"
	web "real-estate-sdr/internal/http"
	"real-estate-sdr/internal/infrastructure/postgres"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type leadResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func TestLeadAPI_CreateAndGet(t *testing.T) {
	pool := newTestPool(t)

	t.Cleanup(func() {
		pool.Close()
	})

	router := newTestRouter(t, pool)

	email := uniqueEmail("integration")

	createBody := fmt.Sprintf(`{
		"name": "Integration Test",
		"email": "%s",
		"phone": "11999999999"
	}`, email)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/leads",
		ToReader(createBody),
	)

	createRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	createResponse := httptest.NewRecorder()

	router.ServeHTTP(
		createResponse,
		createRequest,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			createResponse.Code,
			createResponse.Body.String(),
		)
	}

	var created leadResponse

	if err := json.NewDecoder(
		createResponse.Body,
	).Decode(&created); err != nil {
		t.Fatalf(
			"failed to decode create response: %v",
			err,
		)
	}

	t.Cleanup(func() {
		cleanupLead(t, pool, created.ID)
	})

	if created.ID == "" {
		t.Fatal("expected created lead id")
	}

	if created.Name != "Integration Test" {
		t.Errorf(
			"expected name %q, got %q",
			"Integration Test",
			created.Name,
		)
	}

	if created.Email != email {
		t.Errorf(
			"expected email %q, got %q",
			email,
			created.Email,
		)
	}

	if created.Phone != "11999999999" {
		t.Errorf(
			"expected phone %q, got %q",
			"11999999999",
			created.Phone,
		)
	}

	if created.Status != "NEW" {
		t.Errorf(
			"expected status %q, got %q",
			"NEW",
			created.Status,
		)
	}

	// GET

	getRequest := httptest.NewRequest(
		http.MethodGet,
		"/leads/"+created.ID,
		nil,
	)

	getResponse := httptest.NewRecorder()

	router.ServeHTTP(
		getResponse,
		getRequest,
	)

	if getResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			getResponse.Code,
			getResponse.Body.String(),
		)
	}

	var found leadResponse

	if err := json.NewDecoder(
		getResponse.Body,
	).Decode(&found); err != nil {
		t.Fatalf(
			"failed to decode get response: %v",
			err,
		)
	}

	if found.ID != created.ID {
		t.Errorf(
			"expected id %q, got %q",
			created.ID,
			found.ID,
		)
	}

	if found.Name != created.Name {
		t.Errorf(
			"expected name %q, got %q",
			created.Name,
			found.Name,
		)
	}

	if found.Email != created.Email {
		t.Errorf(
			"expected email %q, got %q",
			created.Email,
			found.Email,
		)
	}

	if found.Phone != created.Phone {
		t.Errorf(
			"expected phone %q, got %q",
			created.Phone,
			found.Phone,
		)
	}

	if found.Status != created.Status {
		t.Errorf(
			"expected status %q, got %q",
			created.Status,
			found.Status,
		)
	}
}

func TestLeadAPI_List(t *testing.T) {
	pool := newTestPool(t)

	t.Cleanup(func() {
		pool.Close()
	})

	router := newTestRouter(t, pool)

	email := uniqueEmail("list")

	createBody := fmt.Sprintf(`{
		"name": "List Integration Test",
		"email": "%s",
		"phone": "11888888888"
	}`, email)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/leads",
		ToReader(createBody),
	)

	createRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	createResponse := httptest.NewRecorder()

	router.ServeHTTP(
		createResponse,
		createRequest,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			createResponse.Code,
			createResponse.Body.String(),
		)
	}

	var created leadResponse

	if err := json.NewDecoder(
		createResponse.Body,
	).Decode(&created); err != nil {
		t.Fatalf(
			"failed to decode create response: %v",
			err,
		)
	}

	t.Cleanup(func() {
		cleanupLead(t, pool, created.ID)
	})

	// LIST

	listRequest := httptest.NewRequest(
		http.MethodGet,
		"/leads",
		nil,
	)

	listResponse := httptest.NewRecorder()

	router.ServeHTTP(
		listResponse,
		listRequest,
	)

	if listResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			listResponse.Code,
			listResponse.Body.String(),
		)
	}

	var leads []leadResponse

	if err := json.NewDecoder(
		listResponse.Body,
	).Decode(&leads); err != nil {
		t.Fatalf(
			"failed to decode list response: %v",
			err,
		)
	}

	found := false

	for _, lead := range leads {
		if lead.ID == created.ID {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf(
			"created lead %s was not found in list response",
			created.ID,
		)
	}
}

func TestLeadAPI_Update(t *testing.T) {
	pool := newTestPool(t)

	t.Cleanup(func() {
		pool.Close()
	})

	router := newTestRouter(t, pool)

	email := uniqueEmail("update")

	createBody := fmt.Sprintf(`{
		"name": "Before Update",
		"email": "%s",
		"phone": "11666666666"
	}`, email)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/leads",
		ToReader(createBody),
	)

	createRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	createResponse := httptest.NewRecorder()

	router.ServeHTTP(
		createResponse,
		createRequest,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			createResponse.Code,
			createResponse.Body.String(),
		)
	}

	var created leadResponse

	if err := json.NewDecoder(
		createResponse.Body,
	).Decode(&created); err != nil {
		t.Fatalf(
			"failed to decode create response: %v",
			err,
		)
	}

	t.Cleanup(func() {
		cleanupLead(t, pool, created.ID)
	})

	updateEmail := uniqueEmail("updated")

	updateBody := fmt.Sprintf(`{
		"name": "After Update",
		"email": "%s",
		"phone": "11666666666",
		"status": "CONTACTED"
	}`, updateEmail)

	updateRequest := httptest.NewRequest(
		http.MethodPut,
		"/leads/"+created.ID,
		ToReader(updateBody),
	)

	updateRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	updateResponse := httptest.NewRecorder()

	router.ServeHTTP(
		updateResponse,
		updateRequest,
	)

	if updateResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			updateResponse.Code,
			updateResponse.Body.String(),
		)
	}

	var updated leadResponse

	if err := json.NewDecoder(
		updateResponse.Body,
	).Decode(&updated); err != nil {
		t.Fatalf(
			"failed to decode update response: %v",
			err,
		)
	}

	if updated.ID != created.ID {
		t.Errorf(
			"expected id %q, got %q",
			created.ID,
			updated.ID,
		)
	}

	if updated.Name != "After Update" {
		t.Errorf(
			"expected name %q, got %q",
			"After Update",
			updated.Name,
		)
	}

	if updated.Email != updateEmail {
		t.Errorf(
			"expected email %q, got %q",
			updateEmail,
			updated.Email,
		)
	}

	if updated.Status != "CONTACTED" {
		t.Errorf(
			"expected status %q, got %q",
			"CONTACTED",
			updated.Status,
		)
	}
}

func TestLeadAPI_UpdateStatus(t *testing.T) {
	pool := newTestPool(t)

	t.Cleanup(func() {
		pool.Close()
	})

	router := newTestRouter(t, pool)

	email := uniqueEmail("status")

	createBody := fmt.Sprintf(`{
		"name": "Status Integration Test",
		"email": "%s",
		"phone": "11777777777"
	}`, email)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/leads",
		ToReader(createBody),
	)

	createRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	createResponse := httptest.NewRecorder()

	router.ServeHTTP(
		createResponse,
		createRequest,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			createResponse.Code,
			createResponse.Body.String(),
		)
	}

	var created leadResponse

	if err := json.NewDecoder(
		createResponse.Body,
	).Decode(&created); err != nil {
		t.Fatalf(
			"failed to decode create response: %v",
			err,
		)
	}

	t.Cleanup(func() {
		cleanupLead(t, pool, created.ID)
	})

	statusBody := `{
		"status": "QUALIFIED"
	}`

	statusRequest := httptest.NewRequest(
		http.MethodPatch,
		"/leads/"+created.ID+"/status",
		ToReader(statusBody),
	)

	statusRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	statusResponse := httptest.NewRecorder()

	router.ServeHTTP(
		statusResponse,
		statusRequest,
	)

	if statusResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			statusResponse.Code,
			statusResponse.Body.String(),
		)
	}

	var updated leadResponse

	if err := json.NewDecoder(
		statusResponse.Body,
	).Decode(&updated); err != nil {
		t.Fatalf(
			"failed to decode status response: %v",
			err,
		)
	}

	if updated.ID != created.ID {
		t.Errorf(
			"expected id %q, got %q",
			created.ID,
			updated.ID,
		)
	}

	if updated.Status != "QUALIFIED" {
		t.Errorf(
			"expected status %q, got %q",
			"QUALIFIED",
			updated.Status,
		)
	}
}

func TestLeadAPI_NotFound(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	request := httptest.NewRequest(
		http.MethodGet,
		"/leads/11111111-1111-1111-1111-111111111111",
		nil,
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestLeadAPI_Health(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	request := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			response.Code,
		)
	}
}

func TestLeadAPI_Ready(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	request := httptest.NewRequest(
		http.MethodGet,
		"/ready",
		nil,
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	var body map[string]string

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"failed to decode ready response: %v",
			err,
		)
	}

	if body["status"] != "UP" {
		t.Fatalf(
			"expected status UP, got %q",
			body["status"],
		)
	}

	if body["database"] != "UP" {
		t.Fatalf(
			"expected database UP, got %q",
			body["database"],
		)
	}
}

func newTestRouter(
	t *testing.T,
	pool *pgxpool.Pool,
) http.Handler {
	t.Helper()
	queries := generated.New(pool)

	leadRepository := postgres.NewLeadRepository(
		queries,
	)

	leadService := application.NewLeadService(
		leadRepository,
	)

	propertyRepository := postgres.NewPropertyRepository(
		queries,
	)

	propertyService := application.NewPropertyService(
		propertyRepository,
	)

	conversationRepository := postgres.NewConversationRepository(
		queries,
	)

	conversationService := application.NewConversationService(
		conversationRepository,
	)

	messageRepository := postgres.NewMessageRepository(
		queries,
	)

	aiService := newTestAIClient(t)

	messageService := application.NewMessageService(
		messageRepository,
		conversationRepository,
		aiService,
	)

	return web.NewRouter(
		leadService,
		propertyService,
		conversationService,
		messageService,
		pool,
	)

}

func newTestPool(
	t *testing.T,
) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not configured")
	}

	pool, err := pgxpool.New(
		context.Background(),
		databaseURL,
	)

	if err != nil {
		t.Fatalf(
			"failed to create PostgreSQL pool: %v",
			err,
		)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()

		t.Fatalf(
			"failed to connect to PostgreSQL: %v",
			err,
		)
	}

	return pool
}

func ToReader(value string) *stringReader {
	return &stringReader{
		value: value,
	}
}

type stringReader struct {
	value string
}

func (r *stringReader) Read(p []byte) (int, error) {
	if len(r.value) == 0 {
		return 0, context.Canceled
	}

	n := copy(p, r.value)
	r.value = r.value[n:]

	return n, nil
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf(
		"%s-%s@example.com",
		prefix,
		uuid.NewString(),
	)
}

func cleanupLead(t *testing.T, pool *pgxpool.Pool, id string) {
	t.Helper()

	if id == "" {
		return
	}

	_, err := pool.Exec(
		context.Background(),
		"DELETE FROM leads WHERE id = $1",
		id,
	)

	if err != nil {
		t.Logf(
			"failed to cleanup lead %s: %v",
			id,
			err,
		)
	}
}

func TestLeadAPI_Create_DuplicateEmail(t *testing.T) {
	pool := newTestPool(t)

	t.Cleanup(func() {
		pool.Close()
	})

	router := newTestRouter(t, pool)

	email := uniqueEmail("duplicate")

	body := fmt.Sprintf(`{
		"name": "Duplicate Test",
		"email": "%s",
		"phone": "11999999999"
	}`, email)

	// Primeiro CREATE
	firstRequest := httptest.NewRequest(
		http.MethodPost,
		"/leads",
		ToReader(body),
	)

	firstRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	firstResponse := httptest.NewRecorder()

	router.ServeHTTP(
		firstResponse,
		firstRequest,
	)

	if firstResponse.Code != http.StatusCreated {
		t.Fatalf(
			"expected first create status %d, got %d: %s",
			http.StatusCreated,
			firstResponse.Code,
			firstResponse.Body.String(),
		)
	}

	var created leadResponse

	if err := json.NewDecoder(
		firstResponse.Body,
	).Decode(&created); err != nil {
		t.Fatalf(
			"failed to decode first create response: %v",
			err,
		)
	}

	t.Cleanup(func() {
		cleanupLead(t, pool, created.ID)
	})

	// Segundo CREATE com o MESMO email.
	secondRequest := httptest.NewRequest(
		http.MethodPost,
		"/leads",
		ToReader(body),
	)

	secondRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	secondResponse := httptest.NewRecorder()

	router.ServeHTTP(
		secondResponse,
		secondRequest,
	)

	if secondResponse.Code != http.StatusConflict {
		t.Fatalf(
			"expected duplicate email status %d, got %d: %s",
			http.StatusConflict,
			secondResponse.Code,
			secondResponse.Body.String(),
		)
	}

	var errorResponse map[string]string

	if err := json.NewDecoder(
		secondResponse.Body,
	).Decode(&errorResponse); err != nil {
		t.Fatalf(
			"failed to decode duplicate response: %v",
			err,
		)
	}

	if errorResponse["error"] != "lead already exists" {
		t.Errorf(
			"expected error %q, got %q",
			"lead already exists",
			errorResponse["error"],
		)
	}
}
