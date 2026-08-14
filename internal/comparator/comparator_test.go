package comparator

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"comparator/internal/dtos"
	"github.com/stretchr/testify/assert"
)

type trackingBody struct {
	io.ReadCloser
	closed bool
}

func (tb *trackingBody) Close() error {
	tb.closed = true
	return tb.ReadCloser.Close()
}

type errorBody struct{}

func (errorBody) Read(p []byte) (int, error) { return 0, errors.New("lectura falló") }
func (errorBody) Close() error               { return nil }

type failingClient struct{}

func (failingClient) Do(req *http.Request) (*http.Response, error) {
	return nil, errors.New("connection refused")
}

// capturingClient captura el request recibido y devuelve una respuesta fija.
type capturingClient struct {
	req *http.Request
}

func (c *capturingClient) Do(req *http.Request) (*http.Response, error) {
	c.req = req
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}, nil
}

func TestService_makeRequest_enviaMetodoYBody(t *testing.T) {
	client := &capturingClient{}
	s := NewComparatorService(client)

	_, err := s.makeRequest(dtos.RequestDetails{
		URL:    "https://example.com/",
		Method: "post",
		Body:   `{"hola":"mundo"}`,
	})
	assert.NoError(t, err)

	assert.Equal(t, http.MethodPost, client.req.Method)

	body, err := io.ReadAll(client.req.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"hola":"mundo"}`, string(body))
}

func TestService_makeRequest_metodoPorDefectoGET(t *testing.T) {
	client := &capturingClient{}
	s := NewComparatorService(client)

	_, err := s.makeRequest(dtos.RequestDetails{URL: "https://example.com/"})
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, client.req.Method)
}

func TestService_makeRequest_metodoNoPermitido(t *testing.T) {
	s := NewComparatorService(nil)

	_, err := s.makeRequest(dtos.RequestDetails{URL: "https://example.com/", Method: "CONNECT"})

	var validationErr *ValidationError
	assert.ErrorAs(t, err, &validationErr)
}

func TestService_compareResponses_cierraBodies(t *testing.T) {
	body1 := &trackingBody{ReadCloser: io.NopCloser(strings.NewReader(`{"a": 1}`))}
	body2 := &trackingBody{ReadCloser: io.NopCloser(strings.NewReader(`{"a": 1}`))}

	resp1 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       body1,
	}
	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       body2,
	}

	s := NewComparatorService(nil)
	_, _ = s.compareResponses(resp1, resp2)

	if !body1.closed || !body2.closed {
		t.Fatal("los bodies deberían cerrarse al finalizar la comparación")
	}
}

func TestService_compareResponses_errorLectura(t *testing.T) {
	resp1 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       errorBody{},
	}
	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}

	s := NewComparatorService(nil)
	_, err := s.compareResponses(resp1, resp2)
	assert.Error(t, err)
}

func TestService_CompareRequest_errorUpstream(t *testing.T) {
	s := NewComparatorService(failingClient{})
	req := dtos.Request{
		Request1: dtos.RequestDetails{URL: "https://8.8.8.8/"},
		Request2: dtos.RequestDetails{URL: "https://8.8.8.8/"},
	}

	_, err := s.CompareRequest(req)

	var upstreamErr *UpstreamError
	assert.ErrorAs(t, err, &upstreamErr)
}

func TestService_compareResponses_bodyNoJSON(t *testing.T) {
	resp1 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`hola mundo`)),
	}
	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`otro texto`)),
	}

	s := NewComparatorService(nil)
	diff, err := s.compareResponses(resp1, resp2)
	assert.NoError(t, err)
	assert.Len(t, diff.BodyDifferences, 1)
	assert.Equal(t, "error", diff.BodyDifferences[0].Tipo)
}

func TestService_compareResponses_topLevelArray(t *testing.T) {
	// Un array como cuerpo raíz se compara elemento por elemento.
	resp1 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`[1,2,3]`)),
	}
	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`[1,2,4]`)),
	}

	s := NewComparatorService(nil)
	diff, err := s.compareResponses(resp1, resp2)
	assert.NoError(t, err)
	assert.Equal(t, []dtos.BodyDifference{
		{Path: "[2]", Tipo: "number", Values: []interface{}{3.0, 4.0}},
	}, diff.BodyDifferences)
}

func TestService_compareResponses_topLevelScalar(t *testing.T) {
	resp1 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`"hola"`)),
	}
	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`"chau"`)),
	}

	s := NewComparatorService(nil)
	diff, err := s.compareResponses(resp1, resp2)
	assert.NoError(t, err)
	assert.Equal(t, []dtos.BodyDifference{
		{Path: "", Tipo: "string", Values: []interface{}{"hola", "chau"}},
	}, diff.BodyDifferences)
}

func TestService_compareResponses_topLevelScalarIgual(t *testing.T) {
	resp1 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`"hola"`)),
	}
	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`"hola"`)),
	}

	s := NewComparatorService(nil)
	diff, err := s.compareResponses(resp1, resp2)
	assert.NoError(t, err)
	assert.Empty(t, diff.BodyDifferences)
}

func TestService_compareResponses_respuestaExcedeLimite(t *testing.T) {
	t.Setenv("MAX_RESPONSE_SIZE", "10")

	resp1 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"valor": "12345678901234567890"}`)),
	}
	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}

	s := NewComparatorService(nil)
	_, err := s.compareResponses(resp1, resp2)

	var upstreamErr *UpstreamError
	assert.ErrorAs(t, err, &upstreamErr)
	assert.Contains(t, upstreamErr.Message, "excede el tamaño máximo")
}

func TestService_compareResponses_respuestaDentroDelLimite(t *testing.T) {
	t.Setenv("MAX_RESPONSE_SIZE", "10")

	resp1 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"a": 1}`)),
	}
	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"a": 1}`)),
	}

	s := NewComparatorService(nil)
	_, err := s.compareResponses(resp1, resp2)
	assert.NoError(t, err)
}

func TestService_compareResponses_headersSimetricos(t *testing.T) {
	resp1 := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": {"application/json"},
			"X-Solo-1":     {"a"},
			"X-Multi":      {"1", "2"},
		},
		Body: io.NopCloser(strings.NewReader(`{}`)),
	}
	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": {"application/json"},
			"X-Solo-2":     {"b"},
			"X-Multi":      {"1", "3"},
		},
		Body: io.NopCloser(strings.NewReader(`{}`)),
	}

	s := NewComparatorService(nil)
	diff, _ := s.compareResponses(resp1, resp2)

	// Header idéntico no debe aparecer
	assert.NotContains(t, diff.Headers, "Content-Type")

	// Header solo en resp1
	assert.Equal(t, []string{"a", ""}, diff.Headers["X-Solo-1"])

	// Header solo en resp2
	assert.Equal(t, []string{"", "b"}, diff.Headers["X-Solo-2"])

	// Multi-valores comparados de forma completa
	assert.Equal(t, []string{"1, 2", "1, 3"}, diff.Headers["X-Multi"])
}

func TestService_compareResponses_headersIdenticos(t *testing.T) {
	resp1 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}
	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}

	s := NewComparatorService(nil)
	diff, _ := s.compareResponses(resp1, resp2)

	assert.Empty(t, diff.Headers)
}

func TestService_compareJSON(t *testing.T) {

	type args struct {
		json1       map[string]interface{}
		json2       map[string]interface{}
		prefix      string
		differences []dtos.BodyDifference
	}

	tests := []struct {
		name                      string
		args                      args
		differencesExpected       []dtos.BodyDifference
		amountDifferencesExpected int
	}{
		{
			name: "Ok",
			args: args{
				json1:       map[string]interface{}{"a": "1"},
				json2:       map[string]interface{}{"a": "2"},
				prefix:      "test",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "test.a", Tipo: "string", Values: []interface{}{"1", "2"}},
			},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok without prefix",
			args: args{
				json1:       map[string]interface{}{"a": "1"},
				json2:       map[string]interface{}{"a": "2"},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "a", Tipo: "string", Values: []interface{}{"1", "2"}},
			},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok with multiple differences",
			args: args{
				json1:       map[string]interface{}{"a": "1", "b": "2", "c": "3"},
				json2:       map[string]interface{}{"a": "11", "b": "12", "c": "13"},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "a", Tipo: "string", Values: []interface{}{"1", "11"}},
				{Path: "b", Tipo: "string", Values: []interface{}{"2", "12"}},
				{Path: "c", Tipo: "string", Values: []interface{}{"3", "13"}},
			},
			amountDifferencesExpected: 3,
		},
		{
			name: "Ok with Arrays without differences",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
				json2:       map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected:       make([]dtos.BodyDifference, 0),
			amountDifferencesExpected: 0,
		},
		{
			name: "Ok with Arrays",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
				json2:       map[string]interface{}{"a": []interface{}{"13", "22", "33"}},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "a[0]", Tipo: "string", Values: []interface{}{"1", "13"}},
				{Path: "a[1]", Tipo: "string", Values: []interface{}{"2", "22"}},
				{Path: "a[2]", Tipo: "string", Values: []interface{}{"3", "33"}},
			},
			amountDifferencesExpected: 3,
		},
		{
			name: "Ok with Arrays, second without array",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
				json2:       map[string]interface{}{"a": "13"},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "a", Tipo: "mixed", Values: []interface{}{[]interface{}{"1", "2", "3"}, "13"}},
			},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok with Arrays, length difference",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
				json2:       map[string]interface{}{"a": []interface{}{}},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "a", Tipo: "array", Values: []interface{}{"different lengths", 3, 0}},
			},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok with Arrays of Json",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": "1"}}},
				json2:       map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": "1"}}},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected:       make([]dtos.BodyDifference, 0),
			amountDifferencesExpected: 0,
		},
		{
			name: "Ok with Arrays of Json",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": "1", "c": "2"}}},
				json2:       map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": "11", "c": "12"}}},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "a[0].b", Tipo: "string", Values: []interface{}{"1", "11"}},
				{Path: "a[0].c", Tipo: "string", Values: []interface{}{"2", "12"}},
			},
			amountDifferencesExpected: 2,
		},
		{
			name: "Ok with Arrays of Json, second without Json",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": "1", "c": "2"}}},
				json2:       map[string]interface{}{"a": []interface{}{"11"}},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "a[0]", Tipo: "mixed", Values: []interface{}{map[string]interface{}{"b": "1", "c": "2"}, "11"}},
			},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok with Json of Json",
			args: args{
				json1:       map[string]interface{}{"a": map[string]interface{}{"b": "1"}},
				json2:       map[string]interface{}{"a": map[string]interface{}{"b": "1"}},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected:       make([]dtos.BodyDifference, 0),
			amountDifferencesExpected: 0,
		},
		{
			name: "Ok with Json of Json",
			args: args{
				json1:       map[string]interface{}{"a": map[string]interface{}{"b": "1"}},
				json2:       map[string]interface{}{"a": map[string]interface{}{"b": "12"}},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "a.b", Tipo: "string", Values: []interface{}{"1", "12"}},
			},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok with Json of Json, second without json",
			args: args{
				json1:       map[string]interface{}{"a": map[string]interface{}{"b": "1"}},
				json2:       map[string]interface{}{"a": "12"},
				prefix:      "",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "a", Tipo: "mixed", Values: []interface{}{map[string]interface{}{"b": "1"}, "12"}},
			},
			amountDifferencesExpected: 1,
		},
		{
			name: "Key not found in second JSON",
			args: args{
				json1:       map[string]interface{}{"a": "1"},
				json2:       make(map[string]interface{}),
				prefix:      "test",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "test.a", Tipo: "missing", Values: []interface{}{"1", "key not found in second JSON"}},
			},
			amountDifferencesExpected: 1,
		}, {
			name: "Key not found in first JSON",
			args: args{
				json1:       make(map[string]interface{}),
				json2:       map[string]interface{}{"a": "1"},
				prefix:      "test",
				differences: make([]dtos.BodyDifference, 0),
			},
			differencesExpected: []dtos.BodyDifference{
				{Path: "test.a", Tipo: "missing", Values: []interface{}{"key not found in first JSON", "1"}},
			},
			amountDifferencesExpected: 1,
		}}
	for _, tt := range tests {

		s := NewComparatorService(nil)

		t.Run(tt.name, func(t *testing.T) {
			s.compareJSON(tt.args.json1, tt.args.json2, tt.args.prefix, &tt.args.differences)

			assert.Len(t, tt.args.differences, tt.amountDifferencesExpected)
			assert.Equal(t, tt.differencesExpected, tt.args.differences)
		})
	}
}
