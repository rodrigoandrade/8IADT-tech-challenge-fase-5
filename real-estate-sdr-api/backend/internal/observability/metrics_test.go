package observability

import (
	"testing"
	"time"
)

func TestMetricsRecordSuccessfulRequest(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordRequest(
		200,
		100*time.Millisecond,
	)

	snapshot := metrics.Snapshot()

	if snapshot.RequestsTotal != 1 {
		t.Fatalf(
			"expected 1 total request, got %d",
			snapshot.RequestsTotal,
		)
	}

	if snapshot.RequestsSuccess != 1 {
		t.Fatalf(
			"expected 1 successful request, got %d",
			snapshot.RequestsSuccess,
		)
	}

	if snapshot.RequestsError != 0 {
		t.Fatalf(
			"expected 0 errors, got %d",
			snapshot.RequestsError,
		)
	}

	if snapshot.AverageDurationMs != 100 {
		t.Fatalf(
			"expected average duration 100ms, got %f",
			snapshot.AverageDurationMs,
		)
	}
}

func TestMetricsRecordFailedRequest(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordRequest(
		500,
		200*time.Millisecond,
	)

	snapshot := metrics.Snapshot()

	if snapshot.RequestsTotal != 1 {
		t.Fatalf(
			"expected 1 total request, got %d",
			snapshot.RequestsTotal,
		)
	}

	if snapshot.RequestsSuccess != 0 {
		t.Fatalf(
			"expected 0 successful requests, got %d",
			snapshot.RequestsSuccess,
		)
	}

	if snapshot.RequestsError != 1 {
		t.Fatalf(
			"expected 1 error, got %d",
			snapshot.RequestsError,
		)
	}
}

func TestMetricsCalculateAverageDuration(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordRequest(
		200,
		100*time.Millisecond,
	)

	metrics.RecordRequest(
		201,
		300*time.Millisecond,
	)

	snapshot := metrics.Snapshot()

	if snapshot.RequestsTotal != 2 {
		t.Fatalf(
			"expected 2 total requests, got %d",
			snapshot.RequestsTotal,
		)
	}

	if snapshot.AverageDurationMs != 200 {
		t.Fatalf(
			"expected average duration 200ms, got %f",
			snapshot.AverageDurationMs,
		)
	}
}
