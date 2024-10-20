package httpclient

import "net/http"

type Httpclient struct {
	client http.Client
}

// TODO revisar
func NewHttpclient() *Httpclient {
	return &Httpclient{
		client: http.Client{},
	}
}

func (c *Httpclient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
