package requestlog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddleware_generaYExponeRequestID(t *testing.T) {
	var gotID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = FromContext(r.Context())
		assert.NotEmpty(t, gotID)
		w.WriteHeader(http.StatusCreated)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/compare", nil)

	Middleware(inner).ServeHTTP(rr, req)

	assert.Equal(t, gotID, rr.Header().Get("X-Request-ID"))
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestMiddleware_respetaRequestIDRecibido(t *testing.T) {
	var gotID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/compare", nil)
	req.Header.Set("X-Request-ID", "id-cliente-123")

	Middleware(inner).ServeHTTP(rr, req)

	assert.Equal(t, "id-cliente-123", rr.Header().Get("X-Request-ID"))
	assert.Equal(t, "id-cliente-123", gotID)
}

func TestMiddleware_capturaStatus(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream error", http.StatusBadGateway)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/compare", nil)

	Middleware(inner).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadGateway, rr.Code)
}

func TestFromContext_sinID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/compare", nil)
	assert.Equal(t, "", FromContext(req.Context()))
}
