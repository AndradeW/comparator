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

// defaultUserAgent identifica la herramienta ante los servidores destino;
// muchas APIs responden distinto al User-Agent por defecto de Go.
const defaultUserAgent = "comparator/1.0"

func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", defaultUserAgent)
	}
	return c.client.Do(req)
}
