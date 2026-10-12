package observability

import (
	"sync/atomic"
	"time"
)

type Metrics struct {
	requestsTotal         atomic.Int64
	requestsSuccess       atomic.Int64
	requestsError         atomic.Int64
	totalDurationNs       atomic.Int64
	propertyMatchAttempts atomic.Int64
	propertyMatchSuccess  atomic.Int64
	propertyMatchNotFound atomic.Int64
}

type Snapshot struct {
	RequestsTotal         int64
	RequestsSuccess       int64
	RequestsError         int64
	AverageDurationMs     float64
	PropertyMatchAttempts int64
	PropertyMatchSuccess  int64
	PropertyMatchNotFound int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) RecordRequest(
	statusCode int,
	duration time.Duration,
) {
	m.requestsTotal.Add(1)

	if statusCode >= 200 && statusCode < 400 {
		m.requestsSuccess.Add(1)
	} else {
		m.requestsError.Add(1)
	}

	m.totalDurationNs.Add(
		duration.Nanoseconds(),
	)
}

func (m *Metrics) RecordPropertyMatch(
	found bool,
) {
	m.propertyMatchAttempts.Add(1)

	if found {
		m.propertyMatchSuccess.Add(1)
		return
	}

	m.propertyMatchNotFound.Add(1)
}

func (m *Metrics) Snapshot() Snapshot {
	total := m.requestsTotal.Load()

	var averageDurationMs float64

	if total > 0 {
		averageDurationMs =
			float64(m.totalDurationNs.Load()) /
				float64(total) /
				float64(time.Millisecond)
	}

	return Snapshot{
		RequestsTotal:         total,
		RequestsSuccess:       m.requestsSuccess.Load(),
		RequestsError:         m.requestsError.Load(),
		AverageDurationMs:     averageDurationMs,
		PropertyMatchAttempts: m.propertyMatchAttempts.Load(),
		PropertyMatchSuccess:  m.propertyMatchSuccess.Load(),
		PropertyMatchNotFound: m.propertyMatchNotFound.Load(),
	}
}
