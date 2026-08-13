package httpclient

import (
	"net/http"

	"comparator/config"
)

type HTTPClient struct {
	client http.Client
}

// TODO revisar
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: http.Client{
			Timeout: config.GetTimeout(),
		},
	}
}

func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
