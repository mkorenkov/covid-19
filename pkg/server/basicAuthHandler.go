package server

import (
	"net/http"
)

// BasicAuthMiddleware holds userPasswords and can perform HTTP Basic Auth.
type BasicAuthMiddleware struct {
	userPasswords map[string]string
}

// BasicAuth makes sure HTTP requests are properly authenticated with HTTP Basic Auth.
func (b BasicAuthMiddleware) BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, _ := r.BasicAuth()

		expectedPassword, ok := b.userPasswords[user]
		if !ok || password != expectedPassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// NewBasicAuthMiddleware creates BasicAuthMiddleware object
func NewBasicAuthMiddleware(userPasswords map[string]string) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{
		userPasswords: userPasswords,
	}
}
