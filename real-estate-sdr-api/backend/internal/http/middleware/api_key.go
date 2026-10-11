package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

func APIKey(
	expectedKey string,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		// Mantém os endpoints operacionais públicos.
		if r.URL.Path == "/health" ||
			r.URL.Path == "/ready" {
			next.ServeHTTP(w, r)
			return
		}

		providedKey := strings.TrimSpace(
			r.Header.Get("X-API-Key"),
		)

		if expectedKey == "" ||
			providedKey == "" ||
			subtle.ConstantTimeCompare(
				[]byte(providedKey),
				[]byte(expectedKey),
			) != 1 {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)
			w.WriteHeader(http.StatusUnauthorized)

			_ = json.NewEncoder(w).Encode(
				map[string]string{
					"error": "unauthorized",
				},
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
