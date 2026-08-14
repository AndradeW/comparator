// Package requestlog agrega un request ID a cada petición y registra logs
// estructurados de inicio y fin (método, ruta, status y duración).
package requestlog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type ctxKey struct{}

// Middleware envuelve el handler, genera (o respeta) un request ID, lo expone
// en el header X-Request-ID y registra el inicio y el final de la petición.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newID()
		}
		r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, id))
		w.Header().Set("X-Request-ID", id)

		logger := slog.With("request_id", id)
		logger.Info("request iniciada", "method", r.Method, "path", r.URL.Path)

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		logger.Info("request completada",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// FromContext devuelve el request ID guardado en el contexto, o "" si no existe.
func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// statusWriter captura el status code para poder loguearlo al finalizar.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b)
}
