package cors

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const testAllowedOrigin = "https://andradew.github.io"

func testHandler() http.Handler {
	return Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func TestPreflight_originPermitido(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", testAllowedOrigin)

	req := httptest.NewRequest(http.MethodOptions, "/compare", nil)
	req.Header.Set("Origin", testAllowedOrigin)
	rec := httptest.NewRecorder()

	testHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("código = %d, se esperaba 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != testAllowedOrigin {
		t.Fatalf("Allow-Origin = %q, se esperaba %q", got, testAllowedOrigin)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "POST, OPTIONS" {
		t.Fatalf("Allow-Methods = %q, se esperaba POST, OPTIONS", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
		t.Fatalf("Allow-Headers = %q, se esperaba Content-Type", got)
	}
}

func TestPreflight_originNoPermitido(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", testAllowedOrigin)

	req := httptest.NewRequest(http.MethodOptions, "/compare", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()

	testHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("código = %d, se esperaba 403", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Allow-Origin = %q, no debería estar seteado", got)
	}
}

func TestRequestNormal_originPermitido(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", testAllowedOrigin)

	req := httptest.NewRequest(http.MethodPost, "/compare", nil)
	req.Header.Set("Origin", testAllowedOrigin)
	rec := httptest.NewRecorder()

	testHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba 200", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != testAllowedOrigin {
		t.Fatalf("Allow-Origin = %q, se esperaba %q", got, testAllowedOrigin)
	}
}

func TestRequestNormal_originNoPermitido(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", testAllowedOrigin)

	req := httptest.NewRequest(http.MethodPost, "/compare", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()

	testHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba 200 (pasa el handler sin CORS)", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Allow-Origin = %q, no debería estar seteado", got)
	}
}

func TestCORSDeshabilitado(t *testing.T) {
	// Sin env CORS_ALLOWED_ORIGIN → CORS deshabilitado
	old, ok := os.LookupEnv("CORS_ALLOWED_ORIGIN")
	os.Unsetenv("CORS_ALLOWED_ORIGIN")
	defer func() {
		if ok {
			os.Setenv("CORS_ALLOWED_ORIGIN", old)
		}
	}()

	req := httptest.NewRequest(http.MethodOptions, "/compare", nil)
	req.Header.Set("Origin", testAllowedOrigin)
	rec := httptest.NewRecorder()
	testHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("preflight con CORS deshabilitado: código = %d, se esperaba 403", rec.Code)
	}
}
