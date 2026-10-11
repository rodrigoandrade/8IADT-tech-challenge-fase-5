package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"real-estate-sdr/internal/application"
	web "real-estate-sdr/internal/http"
	"real-estate-sdr/internal/infrastructure/postgres"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
)

type conversationTestServer struct {
	pool   *pgxpool.Pool
	router http.Handler
}

func setupConversationTestServer(
	t *testing.T,
) *conversationTestServer {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not configured")
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(
		ctx,
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"failed to connect to PostgreSQL: %v",
			err,
		)
	}

	queries := generated.New(pool)

	leadRepository := postgres.NewLeadRepository(
		queries,
	)

	propertyRepository := postgres.NewPropertyRepository(
		queries,
	)

	conversationRepository := postgres.NewConversationRepository(
		queries,
	)

	leadService := application.NewLeadService(
		leadRepository,
	)

	propertyService := application.NewPropertyService(
		propertyRepository,
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

	router := web.NewRouter(
		leadService,
		propertyService,
		conversationService,
		messageService,
		pool,
	)

	t.Cleanup(func() {
		pool.Close()
	})

	return &conversationTestServer{
		pool:   pool,
		router: router,
	}
}

func createTestLead(
	t *testing.T,
	server *conversationTestServer,
) uuid.UUID {
	t.Helper()

	leadID := uuid.New()

	_, err := server.pool.Exec(
		context.Background(),
		`
			INSERT INTO leads (
				id,
				name,
				email,
				phone,
				status
			)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				'NEW'
			)
		`,
		leadID,
		"Conversation Test Lead",
		"conversation-"+leadID.String()+"@example.com",
		"11999999999",
	)

	if err != nil {
		t.Fatalf(
			"failed to create test lead: %v",
			err,
		)
	}

	t.Cleanup(func() {
		_, _ = server.pool.Exec(
			context.Background(),
			"DELETE FROM leads WHERE id = $1",
			leadID,
		)
	})

	return leadID
}

func createConversationRequestBody(
	t *testing.T,
	leadID uuid.UUID,
) *bytes.Buffer {
	t.Helper()

	body := map[string]interface{}{
		"lead_id": leadID,
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf(
			"failed to marshal request: %v",
			err,
		)
	}

	return bytes.NewBuffer(data)
}

func TestConversationAPI_CreateAndGet(t *testing.T) {
	server := setupConversationTestServer(t)

	leadID := createTestLead(
		t,
		server,
	)

	body := createConversationRequestBody(
		t,
		leadID,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/conversations",
		body,
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	var created map[string]interface{}

	if err := json.Unmarshal(
		response.Body.Bytes(),
		&created,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	idString, ok := created["id"].(string)
	if !ok {
		t.Fatal("expected conversation id in response")
	}

	conversationID, err := uuid.Parse(idString)
	if err != nil {
		t.Fatalf(
			"invalid conversation id: %v",
			err,
		)
	}

	if created["status"] != "ACTIVE" {
		t.Fatalf(
			"expected status ACTIVE, got %v",
			created["status"],
		)
	}

	if created["lead_id"] != leadID.String() {
		t.Fatalf(
			"expected lead_id %s, got %v",
			leadID,
			created["lead_id"],
		)
	}

	getRequest := httptest.NewRequest(
		http.MethodGet,
		"/conversations/"+conversationID.String(),
		nil,
	)

	getResponse := httptest.NewRecorder()

	server.router.ServeHTTP(
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

	var result map[string]interface{}

	if err := json.Unmarshal(
		getResponse.Body.Bytes(),
		&result,
	); err != nil {
		t.Fatalf(
			"failed to decode GET response: %v",
			err,
		)
	}

	if result["id"] != conversationID.String() {
		t.Fatalf(
			"expected id %s, got %v",
			conversationID,
			result["id"],
		)
	}

	if result["lead_id"] != leadID.String() {
		t.Fatalf(
			"expected lead_id %s, got %v",
			leadID,
			result["lead_id"],
		)
	}

	if result["status"] != "ACTIVE" {
		t.Fatalf(
			"expected status ACTIVE, got %v",
			result["status"],
		)
	}
}

func TestConversationAPI_List(t *testing.T) {
	server := setupConversationTestServer(t)

	leadID := createTestLead(
		t,
		server,
	)

	for i := 0; i < 3; i++ {
		body := createConversationRequestBody(
			t,
			leadID,
		)

		request := httptest.NewRequest(
			http.MethodPost,
			"/conversations",
			body,
		)

		request.Header.Set(
			"Content-Type",
			"application/json",
		)

		response := httptest.NewRecorder()

		server.router.ServeHTTP(
			response,
			request,
		)

		if response.Code != http.StatusCreated {
			t.Fatalf(
				"expected status %d, got %d: %s",
				http.StatusCreated,
				response.Code,
				response.Body.String(),
			)
		}
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/conversations?limit=2&offset=0",
		nil,
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
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

	var conversations []map[string]interface{}

	if err := json.Unmarshal(
		response.Body.Bytes(),
		&conversations,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if len(conversations) != 2 {
		t.Fatalf(
			"expected 2 conversations, got %d",
			len(conversations),
		)
	}
}

func TestConversationAPI_ListByLead(t *testing.T) {
	server := setupConversationTestServer(t)

	leadID := createTestLead(
		t,
		server,
	)

	for i := 0; i < 2; i++ {
		body := createConversationRequestBody(
			t,
			leadID,
		)

		request := httptest.NewRequest(
			http.MethodPost,
			"/conversations",
			body,
		)

		request.Header.Set(
			"Content-Type",
			"application/json",
		)

		response := httptest.NewRecorder()

		server.router.ServeHTTP(
			response,
			request,
		)

		if response.Code != http.StatusCreated {
			t.Fatalf(
				"expected status %d, got %d: %s",
				http.StatusCreated,
				response.Code,
				response.Body.String(),
			)
		}
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/leads/"+leadID.String()+"/conversations",
		nil,
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
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

	var conversations []map[string]interface{}

	if err := json.Unmarshal(
		response.Body.Bytes(),
		&conversations,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if len(conversations) != 2 {
		t.Fatalf(
			"expected 2 conversations, got %d",
			len(conversations),
		)
	}

	for _, conversation := range conversations {
		if conversation["lead_id"] != leadID.String() {
			t.Fatalf(
				"expected lead_id %s, got %v",
				leadID,
				conversation["lead_id"],
			)
		}
	}
}

func TestConversationAPI_Close(t *testing.T) {
	server := setupConversationTestServer(t)

	leadID := createTestLead(
		t,
		server,
	)

	body := createConversationRequestBody(
		t,
		leadID,
	)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/conversations",
		body,
	)

	createRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	createResponse := httptest.NewRecorder()

	server.router.ServeHTTP(
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

	var created map[string]interface{}

	if err := json.Unmarshal(
		createResponse.Body.Bytes(),
		&created,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	conversationID, err := uuid.Parse(
		created["id"].(string),
	)
	if err != nil {
		t.Fatalf(
			"invalid conversation id: %v",
			err,
		)
	}

	closeRequest := httptest.NewRequest(
		http.MethodPatch,
		"/conversations/"+conversationID.String()+"/close",
		nil,
	)

	closeResponse := httptest.NewRecorder()

	server.router.ServeHTTP(
		closeResponse,
		closeRequest,
	)

	if closeResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			closeResponse.Code,
			closeResponse.Body.String(),
		)
	}

	var closed map[string]interface{}

	if err := json.Unmarshal(
		closeResponse.Body.Bytes(),
		&closed,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if closed["status"] != "CLOSED" {
		t.Fatalf(
			"expected status CLOSED, got %v",
			closed["status"],
		)
	}

	if closed["ended_at"] == nil {
		t.Fatal("expected ended_at to be populated")
	}

	getRequest := httptest.NewRequest(
		http.MethodGet,
		"/conversations/"+conversationID.String(),
		nil,
	)

	getResponse := httptest.NewRecorder()

	server.router.ServeHTTP(
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

	var result map[string]interface{}

	if err := json.Unmarshal(
		getResponse.Body.Bytes(),
		&result,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if result["status"] != "CLOSED" {
		t.Fatalf(
			"expected persisted status CLOSED, got %v",
			result["status"],
		)
	}

	if result["ended_at"] == nil {
		t.Fatal(
			"expected persisted ended_at",
		)
	}
}

func TestConversationAPI_CloseAlreadyClosed(t *testing.T) {
	server := setupConversationTestServer(t)

	leadID := createTestLead(
		t,
		server,
	)

	body := createConversationRequestBody(
		t,
		leadID,
	)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/conversations",
		body,
	)

	createRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	createResponse := httptest.NewRecorder()

	server.router.ServeHTTP(
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

	var created map[string]interface{}

	if err := json.Unmarshal(
		createResponse.Body.Bytes(),
		&created,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	conversationID := created["id"].(string)

	closePath := "/conversations/" +
		conversationID +
		"/close"

	firstRequest := httptest.NewRequest(
		http.MethodPatch,
		closePath,
		nil,
	)

	firstResponse := httptest.NewRecorder()

	server.router.ServeHTTP(
		firstResponse,
		firstRequest,
	)

	if firstResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected first close status %d, got %d: %s",
			http.StatusOK,
			firstResponse.Code,
			firstResponse.Body.String(),
		)
	}

	secondRequest := httptest.NewRequest(
		http.MethodPatch,
		closePath,
		nil,
	)

	secondResponse := httptest.NewRecorder()

	server.router.ServeHTTP(
		secondResponse,
		secondRequest,
	)

	if secondResponse.Code != http.StatusConflict {
		t.Fatalf(
			"expected second close status %d, got %d: %s",
			http.StatusConflict,
			secondResponse.Code,
			secondResponse.Body.String(),
		)
	}
}

func TestConversationAPI_CreateInvalidJSON(t *testing.T) {
	server := setupConversationTestServer(t)

	request := httptest.NewRequest(
		http.MethodPost,
		"/conversations",
		bytes.NewBufferString("{invalid"),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestConversationAPI_CreateInvalidLead(t *testing.T) {
	server := setupConversationTestServer(t)

	body := createConversationRequestBody(
		t,
		uuid.New(),
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/conversations",
		body,
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusInternalServerError &&
		response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400 or 500, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}
}

func TestConversationAPI_GetByIDInvalidID(t *testing.T) {
	server := setupConversationTestServer(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/conversations/not-a-uuid",
		nil,
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestConversationAPI_GetByIDNotFound(t *testing.T) {
	server := setupConversationTestServer(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/conversations/"+uuid.New().String(),
		nil,
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
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

func TestConversationAPI_ListInvalidLimit(t *testing.T) {
	server := setupConversationTestServer(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/conversations?limit=invalid",
		nil,
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestConversationAPI_ListInvalidOffset(t *testing.T) {
	server := setupConversationTestServer(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/conversations?offset=-1",
		nil,
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestConversationAPI_ListByLeadInvalidID(t *testing.T) {
	server := setupConversationTestServer(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/leads/not-a-uuid/conversations",
		nil,
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestConversationAPI_CloseInvalidID(t *testing.T) {
	server := setupConversationTestServer(t)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/conversations/not-a-uuid/close",
		nil,
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestConversationAPI_CloseNotFound(t *testing.T) {
	server := setupConversationTestServer(t)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/conversations/"+uuid.New().String()+"/close",
		nil,
	)

	response := httptest.NewRecorder()

	server.router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusNotFound &&
		response.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status 404 or 500, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}
}
