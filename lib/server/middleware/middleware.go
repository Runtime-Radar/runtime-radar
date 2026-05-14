package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/felixge/httpsnoop"
	"github.com/rs/zerolog/log"
)

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if e := log.Debug(); e.Enabled() {
			m := httpsnoop.CaptureMetrics(next, w, r)
			schema := "http"
			if r.TLS != nil {
				schema = "https"
			}
			log.Debug().Str("delay", m.Duration.String()).
				Int64("length", m.Written).
				Int("code", m.Code).
				Str("schema", schema).
				Interface("headers", sanitizeHeaders(r.Header)).
				Msgf("%s %s request", r.Method, r.URL.RequestURI())

			return
		}

		next.ServeHTTP(w, r)
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				log.Error().Str("stacktrace", string(debug.Stack())).Msgf("Panic caught while calling %s %s: %v", r.Method, r.URL.RequestURI(), p)

				if e := log.Debug(); e.Enabled() {
					debug.PrintStack()
				}

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func sanitizeHeaders(header http.Header) http.Header {
	if header == nil {
		return nil
	}

	h := header.Clone()

	if auth := h.Get("Authorization"); auth != "" {
		if scheme, _, ok := strings.Cut(auth, " "); ok {
			h.Set("Authorization", scheme+" ******")
		} else {
			h.Set("Authorization", "******")
		}
	}

	return h
}
