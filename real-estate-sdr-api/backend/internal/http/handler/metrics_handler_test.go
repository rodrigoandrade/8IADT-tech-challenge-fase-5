package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"real-estate-sdr/internal/observability"
)

func TestMetricsHandlerGet(t *testing.T) {
	metrics := observability.NewMetrics()

	metrics.RecordRequest(
		http.StatusOK,
		100*time.Millisecond,
	)

	metrics.RecordRequest(
		http.StatusCreated,
		300*time.Millisecond,
	)

	handler := NewMetricsHandler(metrics)

	request := httptest.NewRequest(
		http.MethodGet,
		"/metrics",
		nil,
	)

	response := httptest.NewRecorder()

	handler.Get(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			response.Code,
		)
	}

	contentType := response.Header().Get(
		"Content-Type",
	)

	if contentType != "application/json" {
		t.Fatalf(
			"expected application/json, got %s",
			contentType,
		)
	}

	expected := `{"RequestsTotal":2,"RequestsSuccess":2,"RequestsError":0,"AverageDurationMs":200,"PropertyMatchAttempts":0,"PropertyMatchSuccess":0,"PropertyMatchNotFound":0}` + "\n"

	if response.Body.String() != expected {
		t.Fatalf(
			"unexpected response body: %s",
			response.Body.String(),
		)
	}
}
