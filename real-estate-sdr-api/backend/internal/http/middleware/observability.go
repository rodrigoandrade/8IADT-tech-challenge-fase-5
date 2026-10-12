package middleware

import (
	"log"
	"net/http"
	"time"

	"real-estate-sdr/internal/observability"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(body []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	return w.ResponseWriter.Write(body)
}

func Observability(
	metrics *observability.Metrics,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		start := time.Now()

		correlationID := GetCorrelationID(
			r.Context(),
		)

		log.Printf(
			"event=request_started correlation_id=%s method=%s path=%s",
			correlationID,
			r.Method,
			r.URL.Path,
		)

		writer := &responseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(
			writer,
			r,
		)

		duration := time.Since(start)

		if r.URL.Path != "/metrics" {
			metrics.RecordRequest(
				writer.statusCode,
				duration,
			)
		}

		metrics.RecordRequest(
			writer.statusCode,
			duration,
		)

		log.Printf(
			"event=request_completed correlation_id=%s method=%s path=%s status_code=%d duration_ms=%d",
			correlationID,
			r.Method,
			r.URL.Path,
			writer.statusCode,
			duration.Milliseconds(),
		)
	})
}
