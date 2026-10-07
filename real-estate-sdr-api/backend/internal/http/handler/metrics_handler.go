package handler

import (
	"encoding/json"
	"net/http"

	"real-estate-sdr/internal/observability"
)

type MetricsHandler struct {
	metrics *observability.Metrics
}

func NewMetricsHandler(
	metrics *observability.Metrics,
) *MetricsHandler {
	return &MetricsHandler{
		metrics: metrics,
	}
}

func (h *MetricsHandler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	snapshot := h.metrics.Snapshot()

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(snapshot)
}
