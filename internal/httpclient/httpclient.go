package httpclient

import (
	"net/http"

	"comparator/config"
)

type Httpclient struct {
	client http.Client
}

// TODO revisar
func NewHttpclient() *Httpclient {
	return &Httpclient{
		client: http.Client{
			Timeout: config.GetTimeout(),
		},
	}
}

func (c *Httpclient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
