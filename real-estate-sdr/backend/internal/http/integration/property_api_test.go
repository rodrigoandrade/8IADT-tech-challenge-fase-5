package integration

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"real-estate-sdr/internal/application"
	"real-estate-sdr/internal/domain"
	web "real-estate-sdr/internal/http"
	"real-estate-sdr/internal/infrastructure/postgres"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
)

type propertyResponse struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Type          string    `json:"type"`
	Transaction   string    `json:"transaction_type"`
	Price         float64   `json:"price"`
	Bedrooms      int       `json:"bedrooms"`
	Bathrooms     int       `json:"bathrooms"`
	ParkingSpaces int       `json:"parking_spaces"`
	Area          float64   `json:"area"`
	City          string    `json:"city"`
	Neighborhood  string    `json:"neighborhood"`
	Address       string    `json:"address"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func TestPropertyAPI_CreateAndGet(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	propertyID := uuid.Nil

	defer func() {
		if propertyID == uuid.Nil {
			return
		}

		_, err := pool.Exec(
			context.Background(),
			"DELETE FROM properties WHERE id = $1",
			propertyID,
		)

		if err != nil {
			t.Errorf(
				"failed to cleanup property %s: %v",
				propertyID,
				err,
			)
		}
	}()

	// ---------------------------------------------------------
	// CREATE
	// ---------------------------------------------------------

	createBody := `{
		"title": "Apartamento Integration Test",
		"description": "Apartamento criado pelo teste de integração",
		"type": "APARTMENT",
		"transaction_type": "SALE",
		"price": 450000,
		"bedrooms": 3,
		"bathrooms": 2,
		"parking_spaces": 2,
		"area": 120.5,
		"city": "São Paulo",
		"neighborhood": "Vila Mariana",
		"address": "Rua Integration Test, 100",
		"status": "AVAILABLE"
	}`

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/properties",
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

	var created propertyResponse

	if err := json.NewDecoder(
		createResponse.Body,
	).Decode(&created); err != nil {
		t.Fatalf(
			"failed to decode create response: %v",
			err,
		)
	}

	propertyID, err := uuid.Parse(created.ID)
	if err != nil {
		t.Fatalf(
			"invalid property id returned by API: %v",
			err,
		)
	}

	if created.Title != "Apartamento Integration Test" {
		t.Errorf(
			"expected title %q, got %q",
			"Apartamento Integration Test",
			created.Title,
		)
	}

	if created.Type != "APARTMENT" {
		t.Errorf(
			"expected type %q, got %q",
			"APARTMENT",
			created.Type,
		)
	}

	if created.Transaction != "SALE" {
		t.Errorf(
			"expected transaction %q, got %q",
			"SALE",
			created.Transaction,
		)
	}

	if created.Price != 450000 {
		t.Errorf(
			"expected price %.2f, got %.2f",
			450000.0,
			created.Price,
		)
	}

	if created.Area != 120.5 {
		t.Errorf(
			"expected area %.2f, got %.2f",
			120.5,
			created.Area,
		)
	}

	if created.Status != "AVAILABLE" {
		t.Errorf(
			"expected status %q, got %q",
			"AVAILABLE",
			created.Status,
		)
	}

	// ---------------------------------------------------------
	// GET BY ID
	// ---------------------------------------------------------

	getRequest := httptest.NewRequest(
		http.MethodGet,
		"/properties/"+created.ID,
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

	var found propertyResponse

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

	if found.Title != created.Title {
		t.Errorf(
			"expected title %q, got %q",
			created.Title,
			found.Title,
		)
	}

	if found.Price != created.Price {
		t.Errorf(
			"expected price %.2f, got %.2f",
			created.Price,
			found.Price,
		)
	}
}

func TestPropertyAPI_List(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	property := createTestProperty(
		t,
		pool,
		"Property List Integration Test",
	)

	defer cleanupProperty(
		t,
		pool,
		property.ID,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/properties?limit=20&offset=0",
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

	var properties []propertyResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&properties); err != nil {
		t.Fatalf(
			"failed to decode list response: %v",
			err,
		)
	}

	found := false

	for _, item := range properties {
		if item.ID == property.ID.String() {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf(
			"created property %s was not found in list",
			property.ID,
		)
	}
}

func TestPropertyAPI_Update(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	property := createTestProperty(
		t,
		pool,
		"Property Update Integration Test",
	)

	defer cleanupProperty(
		t,
		pool,
		property.ID,
	)

	updateBody := `{
		"title": "Property Updated",
		"description": "Updated description",
		"type": "APARTMENT",
		"transaction_type": "SALE",
		"price": 500000,
		"bedrooms": 4,
		"bathrooms": 3,
		"parking_spaces": 2,
		"area": 150.75,
		"city": "São Paulo",
		"neighborhood": "Moema",
		"address": "Rua Updated, 200",
		"status": "AVAILABLE"
	}`

	request := httptest.NewRequest(
		http.MethodPut,
		"/properties/"+property.ID.String(),
		ToReader(updateBody),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
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

	var updated propertyResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&updated); err != nil {
		t.Fatalf(
			"failed to decode update response: %v",
			err,
		)
	}

	if updated.Title != "Property Updated" {
		t.Errorf(
			"expected title %q, got %q",
			"Property Updated",
			updated.Title,
		)
	}

	if updated.Price != 500000 {
		t.Errorf(
			"expected price %.2f, got %.2f",
			500000.0,
			updated.Price,
		)
	}

	if updated.Bedrooms != 4 {
		t.Errorf(
			"expected bedrooms %d, got %d",
			4,
			updated.Bedrooms,
		)
	}

	if updated.Area != 150.75 {
		t.Errorf(
			"expected area %.2f, got %.2f",
			150.75,
			updated.Area,
		)
	}

	if updated.Neighborhood != "Moema" {
		t.Errorf(
			"expected neighborhood %q, got %q",
			"Moema",
			updated.Neighborhood,
		)
	}
}

func TestPropertyAPI_UpdateStatus(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	property := createTestProperty(
		t,
		pool,
		"Property Status Integration Test",
	)

	defer cleanupProperty(
		t,
		pool,
		property.ID,
	)

	updateBody := `{
		"status": "INACTIVE"
	}`

	request := httptest.NewRequest(
		http.MethodPatch,
		"/properties/"+property.ID.String()+"/status",
		ToReader(updateBody),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
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

	var updated propertyResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&updated); err != nil {
		t.Fatalf(
			"failed to decode status response: %v",
			err,
		)
	}

	if updated.Status != "INACTIVE" {
		t.Errorf(
			"expected status %q, got %q",
			"INACTIVE",
			updated.Status,
		)
	}
}

func TestPropertyAPI_NotFound(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	id := uuid.New()

	request := httptest.NewRequest(
		http.MethodGet,
		"/properties/"+id.String(),
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

func TestPropertyAPI_Create_InvalidJSON(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	request := httptest.NewRequest(
		http.MethodPost,
		"/properties",
		ToReader(`{"title":`),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestPropertyAPI_Create_EmptyBody(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	request := httptest.NewRequest(
		http.MethodPost,
		"/properties",
		nil,
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestPropertyAPI_Create_InvalidProperty(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing title",
			body: `{
				"description": "Test",
				"type": "APARTMENT",
				"transaction_type": "SALE",
				"price": 450000,
				"area": 100,
				"city": "São Paulo",
				"neighborhood": "Moema",
				"address": "Rua Teste, 1"
			}`,
		},
		{
			name: "missing type",
			body: `{
				"title": "Test",
				"description": "Test",
				"transaction_type": "SALE",
				"price": 450000,
				"area": 100,
				"city": "São Paulo",
				"neighborhood": "Moema",
				"address": "Rua Teste, 1"
			}`,
		},
		{
			name: "invalid type",
			body: `{
				"title": "Test",
				"description": "Test",
				"type": "INVALID",
				"transaction_type": "SALE",
				"price": 450000,
				"area": 100,
				"city": "São Paulo",
				"neighborhood": "Moema",
				"address": "Rua Teste, 1"
			}`,
		},
		{
			name: "invalid transaction",
			body: `{
				"title": "Test",
				"description": "Test",
				"type": "APARTMENT",
				"transaction_type": "INVALID",
				"price": 450000,
				"area": 100,
				"city": "São Paulo",
				"neighborhood": "Moema",
				"address": "Rua Teste, 1"
			}`,
		},
		{
			name: "invalid price",
			body: `{
				"title": "Test",
				"description": "Test",
				"type": "APARTMENT",
				"transaction_type": "SALE",
				"price": 0,
				"area": 100,
				"city": "São Paulo",
				"neighborhood": "Moema",
				"address": "Rua Teste, 1"
			}`,
		},
		{
			name: "invalid area",
			body: `{
				"title": "Test",
				"description": "Test",
				"type": "APARTMENT",
				"transaction_type": "SALE",
				"price": 450000,
				"area": 0,
				"city": "São Paulo",
				"neighborhood": "Moema",
				"address": "Rua Teste, 1"
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodPost,
				"/properties",
				ToReader(test.body),
			)

			request.Header.Set(
				"Content-Type",
				"application/json",
			)

			response := httptest.NewRecorder()

			router.ServeHTTP(
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
		})
	}
}

func TestPropertyAPI_GetByID_InvalidID(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	request := httptest.NewRequest(
		http.MethodGet,
		"/properties/not-a-uuid",
		nil,
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestPropertyAPI_Update_InvalidID(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	request := httptest.NewRequest(
		http.MethodPut,
		"/properties/not-a-uuid",
		ToReader(`{}`),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestPropertyAPI_UpdateStatus_InvalidID(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/properties/not-a-uuid/status",
		ToReader(`{"status":"SOLD"}`),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestPropertyAPI_UpdateStatus_InvalidStatus(t *testing.T) {
	pool := newPropertyTestPool(t)
	router := newPropertyTestRouter(pool)

	property := createTestProperty(
		t,
		pool,
		"Property Invalid Status Test",
	)

	defer cleanupProperty(
		t,
		pool,
		property.ID,
	)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/properties/"+property.ID.String()+"/status",
		ToReader(`{"status":"INVALID"}`),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
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

func newPropertyTestRouter(
	pool *pgxpool.Pool,
) http.Handler {
	queries := generated.New(pool)

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

	return web.NewRouter(
		nil,
		propertyService,
		conversationService,
		pool,
	)
}

func newPropertyTestPool(
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
			"failed to ping PostgreSQL: %v",
			err,
		)
	}

	return pool
}

func createTestProperty(
	t *testing.T,
	pool *pgxpool.Pool,
	title string,
) *domain.Property {
	t.Helper()

	queries := generated.New(pool)

	result, err := queries.CreateProperty(
		context.Background(),
		generated.CreatePropertyParams{
			Title:           title,
			Description:     "Integration test property",
			PropertyType:    string(domain.PropertyTypeApartment),
			TransactionType: string(domain.TransactionTypeSale),
			Price:           numericForTest(450000),
			Bedrooms:        3,
			Bathrooms:       2,
			ParkingSpaces:   2,
			Area:            numericForTest(120),
			City:            "São Paulo",
			Neighborhood:    "Moema",
			Address:         "Rua Teste, 100",
			Available:       true,
		},
	)
	if err != nil {
		t.Fatalf(
			"failed to create test property: %v",
			err,
		)
	}

	return &domain.Property{
		ID:            result.ID,
		Title:         result.Title,
		Description:   result.Description,
		Type:          domain.PropertyType(result.PropertyType),
		Transaction:   domain.TransactionType(result.TransactionType),
		Price:         450000,
		Bedrooms:      int(result.Bedrooms),
		Bathrooms:     int(result.Bathrooms),
		ParkingSpaces: int(result.ParkingSpaces),
		Area:          120,
		City:          result.City,
		Neighborhood:  result.Neighborhood,
		Address:       result.Address,
		Status:        domain.PropertyStatusAvailable,
		CreatedAt:     result.CreatedAt.Time,
		UpdatedAt:     result.UpdatedAt.Time,
	}
}

func cleanupProperty(
	t *testing.T,
	pool *pgxpool.Pool,
	id uuid.UUID,
) {
	t.Helper()

	_, err := pool.Exec(
		context.Background(),
		"DELETE FROM properties WHERE id = $1",
		id,
	)

	if err != nil {
		t.Errorf(
			"failed to cleanup property %s: %v",
			id,
			err,
		)
	}
}

func numericForTest(value float64) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   big.NewInt(int64(value * 100)),
		Exp:   -2,
		Valid: true,
	}
}
