package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/metrics"
	"google.golang.org/grpc/metadata"
)

type accessTokenContextKeyType string

const (
	accessTokenContextKey accessTokenContextKeyType = "access_token"
)

const (
	AuthorizationHeader = "Authorization"
	AccessTokenHeader   = "Access-Token"
	CorrelationHeader   = "Grpc-Metadata-Correlation-Id"
)

func JWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if jwtHeader := r.Header.Get(AuthorizationHeader); jwtHeader != "" {
			ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", jwtHeader))
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if accessTokenHeader := r.Header.Get(AccessTokenHeader); accessTokenHeader != "" {
			ctx = context.WithValue(ctx, accessTokenContextKey, accessTokenHeader)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Correlation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrID := r.Header.Get(CorrelationHeader)
		if corrID == "" {
			corrID = uuid.New().String()
		}

		ctx := r.Context()
		ctx = metadata.AppendToOutgoingContext(ctx, interceptor.CorrelationHeader, corrID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AccessTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(accessTokenContextKey).(string)
	return token, ok
}

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w, statusCode: "2xx"}
		next.ServeHTTP(sr, r)

		metrics.RequestCounter.With(
			prometheus.Labels{
				metrics.Method: r.Method,
				metrics.Code:   sr.statusCode,
				metrics.Path:   normalizePath(r.URL.Path),
			},
		).Add(1)

		metrics.RequestDurationHistogram.Observe(time.Since(start).Seconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode string
}

func (rw *statusRecorder) WriteHeader(code int) {
	c := strconv.Itoa(code)
	rw.statusCode = string(c[0]) + "xx"
	rw.ResponseWriter.WriteHeader(code)
}

func normalizePath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if _, err := strconv.ParseUint(part, 10, 64); err == nil {
			parts[i] = "{num}"

			continue
		}

		if uuid.Validate(part) == nil {
			parts[i] = "{uuid}"
		}
	}

	return strings.Join(parts, "/")
}
