package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"comparator/internal/comparator"
	"comparator/internal/dtos"
	"comparator/internal/httpclient"
	"github.com/stretchr/testify/assert"
)

func TestHandler_CompareHandler(t *testing.T) {
	clientMock := httpclient.NewMockHTTPClient()
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
		BodyDifferences: make(map[string][]interface{}),
	}

	assert.Equal(t, expectedResponse, response)
}
