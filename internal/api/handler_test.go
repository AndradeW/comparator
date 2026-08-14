package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"comparator/internal/comparator"
	"comparator/internal/dtos"
	"github.com/stretchr/testify/assert"
)

type fakeService struct {
	err error
}

func (f fakeService) CompareRequest(request dtos.Request) (dtos.CompareResponse, error) {
	return dtos.CompareResponse{}, f.err
}

type mockClient struct{}

func (mockClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
		Header:     make(http.Header),
	}, nil
}

func TestHandler_CompareHandler(t *testing.T) {
	clientMock := mockClient{}
	handler := NewHandler(comparator.NewComparatorService(clientMock))

	server := httptest.NewServer(http.HandlerFunc(handler.CompareHandler))
	defer server.Close()

	body := `{
		"request1": {
			"url": "https://pokeapi.co/api/v2/pokemon/ditto"
		},
		"request2": {
			"url": "https://pokeapi.co/api/v2/pokemon/pikachu"
		}
	}`

	resp, err := http.Post(server.URL+"/compare", "application/json", bytes.NewBuffer([]byte(body)))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response dtos.CompareResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	expectedResponse := dtos.CompareResponse{
		StatusCodes:     nil,
		Headers:         make(map[string][]string),
		BodyDifferences: []dtos.BodyDifference{},
	}

	assert.Equal(t, expectedResponse, response)
}

func TestHandler_CompareHandler_bodyDemasiadoGrande(t *testing.T) {
	clientMock := mockClient{}
	handler := NewHandler(comparator.NewComparatorService(clientMock))

	server := httptest.NewServer(http.HandlerFunc(handler.CompareHandler))
	defer server.Close()

	// Body mayor a 1 MiB (default de MAX_BODY_SIZE)
	bigURL := strings.Repeat("a", (1<<20)+1024)
	body := `{"request1":{"url":"http://` + bigURL + `"},"request2":{"url":"http://` + bigURL + `"}}`

	resp, err := http.Post(server.URL+"/compare", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestHandler_CompareHandler_urlInvalida(t *testing.T) {
	clientMock := mockClient{}
	handler := NewHandler(comparator.NewComparatorService(clientMock))

	server := httptest.NewServer(http.HandlerFunc(handler.CompareHandler))
	defer server.Close()

	body := `{
		"request1": {"url": "ftp://example.com"},
		"request2": {"url": "https://example.com"}
	}`

	resp, err := http.Post(server.URL+"/compare", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CompareHandler_errorInterno_noExponeDetalles(t *testing.T) {
	handler := NewHandler(fakeService{err: errors.New("secreto interno de conexión")})

	server := httptest.NewServer(http.HandlerFunc(handler.CompareHandler))
	defer server.Close()

	body := `{
		"request1": {"url": "https://example.com"},
		"request2": {"url": "https://example.com"}
	}`

	resp, err := http.Post(server.URL+"/compare", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.NotContains(t, string(respBody), "secreto interno")
	assert.Contains(t, string(respBody), "internal server error")
}

func TestHandler_CompareHandler_errorUpstream(t *testing.T) {
	handler := NewHandler(fakeService{err: &comparator.UpstreamError{Message: "connection refused"}})

	server := httptest.NewServer(http.HandlerFunc(handler.CompareHandler))
	defer server.Close()

	body := `{
		"request1": {"url": "https://example.com"},
		"request2": {"url": "https://example.com"}
	}`

	resp, err := http.Post(server.URL+"/compare", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}
