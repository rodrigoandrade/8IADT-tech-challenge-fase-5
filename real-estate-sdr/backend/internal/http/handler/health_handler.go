package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		pool: pool,
	}
}

func (h *HealthHandler) Check(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(
		map[string]string{
			"status": "UP",
		},
	)
}

func (h *HealthHandler) Ready(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if err := h.pool.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)

		_ = json.NewEncoder(w).Encode(
			map[string]string{
				"status":   "DOWN",
				"database": "DOWN",
			},
		)

		return
	}

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(
		map[string]string{
			"status":   "UP",
			"database": "UP",
		},
	)
}
