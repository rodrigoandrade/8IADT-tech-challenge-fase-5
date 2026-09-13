package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"real-estate-sdr/internal/infrastructure/ai"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/domain"
	web "real-estate-sdr/internal/http"
	"real-estate-sdr/internal/infrastructure/postgres"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
)

type messageTestServer struct {
	pool   *pgxpool.Pool
	router http.Handler
}

func setupMessageTestServer(
	t *testing.T,
) *messageTestServer {
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

	messageRepository := postgres.NewMessageRepository(
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

	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		log.Fatal("AI_SERVICE_URL is not configured")
	}

	aiService := ai.NewHTTPClient(aiServiceURL, nil)

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

	return &messageTestServer{
		pool:   pool,
		router: router,
	}
}

func createTestLeadForMessage(
	t *testing.T,
	server *messageTestServer,
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
		"Message Test Lead",
		"message-"+leadID.String()+"@example.com",
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

func createTestConversationForMessage(
	t *testing.T,
	server *messageTestServer,
	leadID uuid.UUID,
) uuid.UUID {
	t.Helper()

	conversationID := uuid.New()
	startedAt := time.Now()

	_, err := server.pool.Exec(
		context.Background(),
		`
			INSERT INTO conversations (
				id,
				lead_id,
				status,
				started_at,
				ended_at
			)
			VALUES (
				$1,
				$2,
				'ACTIVE',
				$3,
				NULL
			)
		`,
		conversationID,
		leadID,
		startedAt,
	)

	if err != nil {
		t.Fatalf(
			"failed to create test conversation: %v",
			err,
		)
	}

	t.Cleanup(func() {
		_, _ = server.pool.Exec(
			context.Background(),
			"DELETE FROM conversations WHERE id = $1",
			conversationID,
		)
	})

	return conversationID
}

func createTestMessage(
	t *testing.T,
	server *messageTestServer,
	conversationID uuid.UUID,
) uuid.UUID {
	t.Helper()

	messageID := uuid.New()

	_, err := server.pool.Exec(
		context.Background(),
		`
			INSERT INTO messages (
				id,
				conversation_id,
				role,
				content
			)
			VALUES (
				$1,
				$2,
				$3,
				$4
			)
		`,
		messageID,
		conversationID,
		"USER",
		"Mensagem de teste",
	)

	if err != nil {
		t.Fatalf(
			"failed to create test message: %v",
			err,
		)
	}

	t.Cleanup(func() {
		_, _ = server.pool.Exec(
			context.Background(),
			"DELETE FROM messages WHERE id = $1",
			messageID,
		)
	})

	return messageID
}

func TestMessageAPI_CreateAndGet(t *testing.T) {
	server := setupMessageTestServer(t)

	leadID := createTestLeadForMessage(
		t,
		server,
	)

	conversationID := createTestConversationForMessage(
		t,
		server,
		leadID,
	)

	message := domain.Message{
		ConversationID: conversationID,
		Role:           domain.MessageRoleUser,
		Content:        "Olá, gostaria de saber mais sobre o imóvel.",
	}

	data := []byte(`{
    	"conversation_id": "` + conversationID.String() + `",
    	"role": "USER",
    	"content": "Olá, gostaria de saber mais sobre o imóvel."
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/messages",
		bytes.NewReader(data),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var created domain.Message

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&created,
	); err != nil {
		t.Fatalf(
			"failed to decode created message: %v",
			err,
		)
	}

	if created.ID == uuid.Nil {
		t.Fatal("expected created message ID")
	}

	if created.ConversationID != conversationID {
		t.Fatalf(
			"expected conversation ID %s, got %s",
			conversationID,
			created.ConversationID,
		)
	}

	if created.Role != domain.MessageRoleUser {
		t.Fatalf(
			"expected role %q, got %q",
			domain.MessageRoleUser,
			created.Role,
		)
	}

	if created.Content != message.Content {
		t.Fatalf(
			"expected content %q, got %q",
			message.Content,
			created.Content,
		)
	}

	messageID := created.ID

	getRequest := httptest.NewRequest(
		http.MethodGet,
		"/messages/"+messageID.String(),
		nil,
	)

	getRecorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		getRecorder,
		getRequest,
	)

	if getRecorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			getRecorder.Code,
			getRecorder.Body.String(),
		)
	}

	var found domain.Message

	if err := json.Unmarshal(
		getRecorder.Body.Bytes(),
		&found,
	); err != nil {
		t.Fatalf(
			"failed to decode message: %v",
			err,
		)
	}

	if found.ID != messageID {
		t.Fatalf(
			"expected message ID %s, got %s",
			messageID,
			found.ID,
		)
	}

	if found.ConversationID != conversationID {
		t.Fatalf(
			"expected conversation ID %s, got %s",
			conversationID,
			found.ConversationID,
		)
	}

	if found.Role != domain.MessageRoleUser {
		t.Fatalf(
			"expected role %q, got %q",
			domain.MessageRoleUser,
			found.Role,
		)
	}

	if found.Content != message.Content {
		t.Fatalf(
			"expected content %q, got %q",
			message.Content,
			found.Content,
		)
	}
}

func TestMessageAPI_List(t *testing.T) {
	server := setupMessageTestServer(t)

	leadID := createTestLeadForMessage(
		t,
		server,
	)

	conversationID := createTestConversationForMessage(
		t,
		server,
		leadID,
	)

	message1 := createTestMessage(
		t,
		server,
		conversationID,
	)

	message2 := createTestMessage(
		t,
		server,
		conversationID,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/messages?limit=100&offset=0",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var messages []domain.Message

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&messages,
	); err != nil {
		t.Fatalf(
			"failed to decode messages: %v",
			err,
		)
	}

	foundMessage1 := false
	foundMessage2 := false

	for _, message := range messages {
		if message.ID == message1 {
			foundMessage1 = true
		}

		if message.ID == message2 {
			foundMessage2 = true
		}
	}

	if !foundMessage1 {
		t.Fatalf(
			"message %s not found in list",
			message1,
		)
	}

	if !foundMessage2 {
		t.Fatalf(
			"message %s not found in list",
			message2,
		)
	}
}

func TestMessageAPI_ListByConversation(t *testing.T) {
	server := setupMessageTestServer(t)

	leadID := createTestLeadForMessage(
		t,
		server,
	)

	conversationID := createTestConversationForMessage(
		t,
		server,
		leadID,
	)

	messageID := createTestMessage(
		t,
		server,
		conversationID,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/conversations/"+conversationID.String()+"/messages",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var messages []domain.Message

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&messages,
	); err != nil {
		t.Fatalf(
			"failed to decode messages: %v",
			err,
		)
	}

	found := false

	for _, message := range messages {
		if message.ID == messageID {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf(
			"message %s not found for conversation %s",
			messageID,
			conversationID,
		)
	}
}

func TestMessageAPI_Count(t *testing.T) {
	server := setupMessageTestServer(t)

	leadID := createTestLeadForMessage(
		t,
		server,
	)

	conversationID := createTestConversationForMessage(
		t,
		server,
		leadID,
	)

	createTestMessage(
		t,
		server,
		conversationID,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/messages/count",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response struct {
		Count int64 `json:"count"`
	}

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf(
			"failed to decode count response: %v",
			err,
		)
	}

	if response.Count <= 0 {
		t.Fatalf(
			"expected count greater than zero, got %d",
			response.Count,
		)
	}
}

func TestMessageAPI_CountByConversation(t *testing.T) {
	server := setupMessageTestServer(t)

	leadID := createTestLeadForMessage(
		t,
		server,
	)

	conversationID := createTestConversationForMessage(
		t,
		server,
		leadID,
	)

	createTestMessage(
		t,
		server,
		conversationID,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/conversations/"+conversationID.String()+"/messages/count",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response struct {
		Count int64 `json:"count"`
	}

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf(
			"failed to decode count response: %v",
			err,
		)
	}

	if response.Count <= 0 {
		t.Fatalf(
			"expected count greater than zero, got %d",
			response.Count,
		)
	}
}

func TestMessageAPI_CreateInvalidJSON(t *testing.T) {
	server := setupMessageTestServer(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/messages",
		bytes.NewBufferString(
			`{"conversation_id":`,
		),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestMessageAPI_CreateInvalidConversation(t *testing.T) {
	server := setupMessageTestServer(t)

	invalidConversationID := uuid.New()

	body := map[string]interface{}{
		"conversation_id": invalidConversationID,
		"role":            "USER",
		"content":         "Olá, gostaria de saber mais sobre o imóvel.",
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf(
			"failed to marshal request: %v",
			err,
		)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/messages",
		bytes.NewReader(data),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest &&
		recorder.Code != http.StatusNotFound &&
		recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status 400, 404 or 500, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestMessageAPI_GetByIDInvalidID(t *testing.T) {
	server := setupMessageTestServer(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/messages/not-a-uuid",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestMessageAPI_GetByIDNotFound(t *testing.T) {
	server := setupMessageTestServer(t)

	messageID := uuid.New()

	req := httptest.NewRequest(
		http.MethodGet,
		"/messages/"+messageID.String(),
		nil,
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestMessageAPI_ListByConversationInvalidID(t *testing.T) {
	server := setupMessageTestServer(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/conversations/not-a-uuid/messages",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestMessageAPI_CountByConversationInvalidID(t *testing.T) {
	server := setupMessageTestServer(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/conversations/not-a-uuid/messages/count",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestMessageAPI_ListInvalidLimit(t *testing.T) {
	server := setupMessageTestServer(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/messages?limit=abc",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestMessageAPI_ListInvalidOffset(t *testing.T) {
	server := setupMessageTestServer(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/messages?offset=-1",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
