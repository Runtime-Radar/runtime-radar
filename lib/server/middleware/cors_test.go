package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func handlerWithCORS(allowedOrigins []string) http.Handler {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return CORS(allowedOrigins, DefaultCORSHeaders)(next)
}

func originHeaderFor(t *testing.T, allowedOrigins []string, origin string) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rule", nil)
	req.Header.Set("Origin", origin)

	rec := httptest.NewRecorder()
	handlerWithCORS(allowedOrigins).ServeHTTP(rec, req)

	return rec.Header().Get("Access-Control-Allow-Origin")
}

func TestCORSEmptyOriginsAllowsNobody(t *testing.T) {
	// rs/cors reads an empty origin list as "allow any", which is exactly what must not happen here
	if got := originHeaderFor(t, nil, "https://attacker.tld"); got != "" {
		t.Fatalf("Expected no allow-origin header, got %q", got)
	}
	if got := originHeaderFor(t, []string{}, "https://attacker.tld"); got != "" {
		t.Fatalf("Expected no allow-origin header, got %q", got)
	}
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	allowed := []string{"https://central.example"}

	if got := originHeaderFor(t, allowed, "https://central.example"); got != "https://central.example" {
		t.Fatalf("Expected the configured origin to be allowed, got %q", got)
	}
	if got := originHeaderFor(t, allowed, "https://attacker.tld"); got != "" {
		t.Fatalf("Expected an unknown origin to be rejected, got %q", got)
	}
}

func TestCORSPassesRequestThroughWhenDisabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rule", nil)
	rec := httptest.NewRecorder()

	handlerWithCORS(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected same-origin requests to still be served, got %d", rec.Code)
	}
}
