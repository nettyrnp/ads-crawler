package middleware

import (
	"context"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

// RequestID generates an unique identifier
func RequestID() Middleware {

	type ctxKey int
	const ridKey ctxKey = ctxKey(0)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get("X-Request-ID")
			if rid == "" {
				rid = uuid.NewV4().String()
				r.Header.Set("X-Request-ID", rid)
			}
			ctx := context.WithValue(r.Context(), ridKey, rid)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
