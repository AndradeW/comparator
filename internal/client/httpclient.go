package client

import (
	"bytes"
	"io"
	"net/http"
)

type MockHTTPClient struct{}

func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{}
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	response := `{}`

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(response))),
		Header:     make(http.Header),
	}, nil
}
