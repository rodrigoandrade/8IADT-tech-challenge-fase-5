package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLeadAPI_CreateConcurrent(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	const totalRequests = 20

	testID := fmt.Sprintf("%d", time.Now().UnixNano())

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		responses []leadResponse
		errors    []error
	)

	wg.Add(totalRequests)

	for i := 0; i < totalRequests; i++ {
		i := i

		go func() {
			defer wg.Done()

			email := fmt.Sprintf(
				"concurrent-%s-%d@example.com",
				testID,
				i,
			)

			body := fmt.Sprintf(
				`{
					"name": "Concurrent Lead %d",
					"email": "%s",
					"phone": "119%08d"
				}`,
				i,
				email,
				i,
			)

			request := httptest.NewRequest(
				http.MethodPost,
				"/leads",
				strings.NewReader(body),
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

			if response.Code != http.StatusCreated {
				mu.Lock()
				errors = append(
					errors,
					fmt.Errorf(
						"request %d: expected status %d, got %d: %s",
						i,
						http.StatusCreated,
						response.Code,
						response.Body.String(),
					),
				)
				mu.Unlock()

				return
			}

			var lead leadResponse

			if err := json.NewDecoder(
				response.Body,
			).Decode(&lead); err != nil {
				mu.Lock()
				errors = append(
					errors,
					fmt.Errorf(
						"request %d: failed to decode response: %w",
						i,
						err,
					),
				)
				mu.Unlock()

				return
			}

			if lead.ID == "" {
				mu.Lock()
				errors = append(
					errors,
					fmt.Errorf(
						"request %d: returned empty lead ID",
						i,
					),
				)
				mu.Unlock()

				return
			}

			if lead.Status != "NEW" {
				mu.Lock()
				errors = append(
					errors,
					fmt.Errorf(
						"request %d: expected status NEW, got %q",
						i,
						lead.Status,
					),
				)
				mu.Unlock()

				return
			}

			mu.Lock()
			responses = append(responses, lead)
			mu.Unlock()
		}()
	}

	wg.Wait()

	if len(errors) > 0 {
		for _, err := range errors {
			t.Error(err)
		}

		t.Fatalf(
			"%d concurrent requests failed",
			len(errors),
		)
	}

	if len(responses) != totalRequests {
		t.Fatalf(
			"expected %d successful responses, got %d",
			totalRequests,
			len(responses),
		)
	}

	ids := make(map[string]struct{})

	for _, lead := range responses {
		if _, exists := ids[lead.ID]; exists {
			t.Fatalf(
				"duplicate lead ID detected: %s",
				lead.ID,
			)
		}

		ids[lead.ID] = struct{}{}
	}

	if len(ids) != totalRequests {
		t.Fatalf(
			"expected %d unique IDs, got %d",
			totalRequests,
			len(ids),
		)
	}

	var count int

	err := pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*)
			FROM leads
			WHERE email LIKE $1
		`,
		fmt.Sprintf("concurrent-%s-%%@example.com", testID),
	).Scan(&count)

	if err != nil {
		t.Fatalf(
			"failed to count persisted leads: %v",
			err,
		)
	}

	if count != totalRequests {
		t.Fatalf(
			"expected %d persisted leads, got %d",
			totalRequests,
			count,
		)
	}
}

func TestLeadAPI_CreateConcurrent_DifferentEmails(t *testing.T) {
	pool := newTestPool(t)
	defer pool.Close()

	router := newTestRouter(t, pool)

	const totalRequests = 10

	testID := fmt.Sprintf("%d", time.Now().UnixNano())

	var wg sync.WaitGroup

	wg.Add(totalRequests)

	errors := make(chan error, totalRequests)

	for i := 0; i < totalRequests; i++ {
		i := i

		go func() {
			defer wg.Done()

			email := fmt.Sprintf(
				"unique-%s-%d@example.com",
				testID,
				i,
			)

			body := fmt.Sprintf(
				`{
					"name": "Email Test %d",
					"email": "%s",
					"phone": "118%08d"
				}`,
				i,
				email,
				i,
			)

			request := httptest.NewRequest(
				http.MethodPost,
				"/leads",
				strings.NewReader(body),
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

			if response.Code != http.StatusCreated {
				errors <- fmt.Errorf(
					"request %d: expected %d, got %d",
					i,
					http.StatusCreated,
					response.Code,
				)
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}

	var count int

	err := pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*)
			FROM leads
			WHERE email LIKE $1
		`,
		fmt.Sprintf("unique-%s-%%@example.com", testID),
	).Scan(&count)

	if err != nil {
		t.Fatalf(
			"failed to count persisted leads: %v",
			err,
		)
	}

	if count != totalRequests {
		t.Fatalf(
			"expected %d persisted leads, got %d",
			totalRequests,
			count,
		)
	}
}

// Evita que o compilador reclame caso o pool seja utilizado
// apenas em futuras extensÃµes do teste.
var _ *pgxpool.Pool
