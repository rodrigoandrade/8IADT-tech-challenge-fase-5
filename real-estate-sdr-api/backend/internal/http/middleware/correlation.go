package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type correlationIDKey struct{}

func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		correlationID := r.Header.Get("X-Correlation-ID")

		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		ctx := context.WithValue(
			r.Context(),
			correlationIDKey{},
			correlationID,
		)

		w.Header().Set(
			"X-Correlation-ID",
			correlationID,
		)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}

func GetCorrelationID(ctx context.Context) string {
	value := ctx.Value(correlationIDKey{})

	if value == nil {
		return ""
	}

	correlationID, ok := value.(string)

	if !ok {
		return ""
	}

	return correlationID
}
