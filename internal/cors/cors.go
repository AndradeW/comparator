package cors

import (
	"net/http"

	"comparator/config"
)

// Middleware agrega los headers CORS a las respuestas y responde el preflight OPTIONS.
// Solo permite el origin configurado (CORS_ALLOWED_ORIGIN); orígenes distintos no reciben
// Access-Control-Allow-Origin (el navegador los bloquea). Si la env no está seteada,
// CORS queda deshabilitado.
func Middleware(next http.Handler) http.Handler {
	allowed := config.GetCORSAllowedOrigin()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Preflight del navegador
		if r.Method == http.MethodOptions {
			if !originAllowed(r, allowed) {
				http.Error(w, "origin no permitido", http.StatusForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", allowed)
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.Header().Set("Vary", "Origin")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if originAllowed(r, allowed) {
			w.Header().Set("Access-Control-Allow-Origin", allowed)
			w.Header().Set("Vary", "Origin")
		}

		next.ServeHTTP(w, r)
	})
}

// originAllowed indica si el origin de la request puede cruzar el dominio.
// Sin origin (curl/Postman) se permite; sin allowed configurado no se permite ninguno.
func originAllowed(r *http.Request, allowed string) bool {
	if allowed == "" {
		return false
	}
	origin := r.Header.Get("Origin")
	return origin == "" || origin == allowed
}