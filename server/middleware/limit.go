package middleware

import (
	"net/http"
)

type limitMiddleware struct {
	maxRequestBodySize int64
}

func NewLimitMiddleware(maxRequestBodySize int64) *limitMiddleware {
	return &limitMiddleware{
		maxRequestBodySize,
	}
}

// Middleware limits the request body size.
// This is done by first constraining to the ContentLength of the request headder,
// and then reading the actual Body to constraint it.
func (mw *limitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > mw.maxRequestBodySize {
			http.Error(w, "request body too large", http.StatusBadRequest)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, mw.maxRequestBodySize)
		defer r.Body.Close()

		next.ServeHTTP(w, r)
	})
}
