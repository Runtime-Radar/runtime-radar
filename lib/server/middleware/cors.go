package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

// DefaultCORSHeaders are the headers the UI sends to every service.
var DefaultCORSHeaders = []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "Authorization"}

var corsMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

// CORS allows cross-origin API calls from allowedOrigins, which is how a central CS UI reaches a
// child cluster. An empty list is skipped, since rs/cors would read it as "allow any".
func CORS(allowedOrigins, allowedHeaders []string) func(http.Handler) http.Handler {
	if len(allowedOrigins) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	return cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: corsMethods,
		AllowedHeaders: allowedHeaders,
	}).Handler
}
