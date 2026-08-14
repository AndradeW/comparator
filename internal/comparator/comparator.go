package comparator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"comparator/config"
	"comparator/internal/dtos"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Service struct {
	client httpClient
}

func NewComparatorService(client httpClient) *Service {
	return &Service{client: client}
}

func (s *Service) CompareRequest(request dtos.Request) (dtos.CompareResponse, error) {

	response1, err := s.makeRequest(request.Request1)
	if err != nil {
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			return dtos.CompareResponse{}, err
		}
		slog.Error("error en la petición 1", "error", err)
		return dtos.CompareResponse{}, &UpstreamError{Message: fmt.Sprintf("error en la petición 1 : %s", err)}
	}

	response2, err := s.makeRequest(request.Request2)
	if err != nil {
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			return dtos.CompareResponse{}, err
		}
		slog.Error("error en la petición 2", "error", err)
		return dtos.CompareResponse{}, &UpstreamError{Message: fmt.Sprintf("error en la petición 2 : %s", err)}
	}

	differences, err := s.compareResponses(response1, response2)
	if err != nil {
		return dtos.CompareResponse{}, err
	}

	return differences, nil
}

// Métodos HTTP permitidos para la petición upstream.
var allowedMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodPatch:   {},
	http.MethodDelete:  {},
	http.MethodHead:    {},
	http.MethodOptions: {},
}

// Función para realizar la petición HTTP
func (s *Service) makeRequest(reqDetails dtos.RequestDetails) (*http.Response, error) {
	if err := validateURL(reqDetails.URL); err != nil {
		return nil, err
	}

	method := http.MethodGet
	if reqDetails.Method != "" {
		method = strings.ToUpper(reqDetails.Method)
		if _, ok := allowedMethods[method]; !ok {
			return nil, &ValidationError{Message: fmt.Sprintf("método HTTP no permitido: %s", reqDetails.Method)}
		}
	}

	// Construir la URL con parámetros
	req, err := http.NewRequest(method, reqDetails.URL, strings.NewReader(reqDetails.Body))
	if err != nil {
		return nil, err
	}

	// Agregar headers a la petición
	for key, value := range reqDetails.Headers {
		req.Header.Set(key, value)
	}

	// Agregar parámetros a la URL
	q := req.URL.Query()
	for key, value := range reqDetails.Params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// Hacer la petición
	return s.client.Do(req)
}

// Función para comparar las respuestas HTTP
func (s *Service) compareResponses(resp1, resp2 *http.Response) (dtos.CompareResponse, error) {
	defer func() { _ = resp1.Body.Close() }()
	defer func() { _ = resp2.Body.Close() }()

	differences := dtos.CompareResponse{
		Headers:         make(map[string][]string),
		BodyDifferences: []dtos.BodyDifference{},
	}

	// Comparar los códigos de estado
	if resp1.StatusCode != resp2.StatusCode {
		differences.StatusCodes = []int{resp1.StatusCode, resp2.StatusCode}
	}

	// Comparar los headers (unión de keys de ambas respuestas, preservando multi-valores)
	headerKeys := make(map[string]struct{})
	for key := range resp1.Header {
		headerKeys[key] = struct{}{}
	}
	for key := range resp2.Header {
		headerKeys[key] = struct{}{}
	}

	for key := range headerKeys {
		val1 := resp1.Header.Values(key)
		val2 := resp2.Header.Values(key)
		if !reflect.DeepEqual(val1, val2) {
			differences.Headers[key] = []string{strings.Join(val1, ", "), strings.Join(val2, ", ")}
		}
	}

	// Comparar los cuerpos de la respuesta (asumiendo que son JSON)
	body1, err := s.readBody(resp1.Body, "la respuesta 1")
	if err != nil {
		return dtos.CompareResponse{}, err
	}
	body2, err := s.readBody(resp2.Body, "la respuesta 2")
	if err != nil {
		return dtos.CompareResponse{}, err
	}

	var json1, json2 interface{}
	err1 := json.Unmarshal(body1, &json1)
	err2 := json.Unmarshal(body2, &json2)

	if err1 != nil || err2 != nil {
		// Si hay error al parsear JSON, agregar los cuerpos completos a las diferencias
		differences.BodyDifferences = append(differences.BodyDifferences, dtos.BodyDifference{
			Path:   "",
			Tipo:   "error",
			Values: []interface{}{string(body1), string(body2)},
		})
	} else {
		// Comparar los JSON (soporta objetos, arrays y escalares como raíz)
		s.compareBodyValues(json1, json2, "", &differences.BodyDifferences)
	}

	return differences, nil
}

// readBody lee el cuerpo de una respuesta aplicando el límite MAX_RESPONSE_SIZE.
func (s *Service) readBody(body io.Reader, label string) ([]byte, error) {
	max := config.GetMaxResponseSize()
	data, err := io.ReadAll(io.LimitReader(body, max+1))
	if err != nil {
		return nil, &UpstreamError{Message: fmt.Sprintf("error al leer el cuerpo de %s: %s", label, err)}
	}
	if int64(len(data)) > max {
		return nil, &UpstreamError{Message: fmt.Sprintf("el cuerpo de %s excede el tamaño máximo (%d bytes)", label, max)}
	}
	return data, nil
}

// compareBodyValues compara dos valores JSON arbitrarios (objeto, array o escalar),
// incluso como cuerpo raíz.
func (s *Service) compareBodyValues(v1, v2 interface{}, prefix string, differences *[]dtos.BodyDifference) {
	switch a := v1.(type) {
	case map[string]interface{}:
		if b, ok := v2.(map[string]interface{}); ok {
			s.compareJSON(a, b, prefix, differences)
		} else {
			s.addDiff(differences, prefix, mixedType(a, v2), a, v2)
		}
	case []interface{}:
		if b, ok := v2.([]interface{}); ok {
			s.compareArray(a, b, prefix, differences)
		} else {
			s.addDiff(differences, prefix, mixedType(a, v2), a, v2)
		}
	default:
		if !reflect.DeepEqual(v1, v2) {
			s.addDiff(differences, prefix, mixedType(v1, v2), v1, v2)
		}
	}
}

// addDiff agrega una diferencia al resultado, normalizando su tipo.
func (s *Service) addDiff(differences *[]dtos.BodyDifference, path, tipo string, values ...interface{}) {
	*differences = append(*differences, dtos.BodyDifference{Path: path, Tipo: tipo, Values: values})
}

// mixedType devuelve el tipo JSON de los valores si coinciden, o "mixed" si difieren.
func mixedType(v1, v2 interface{}) string {
	t1, t2 := jsonType(v1), jsonType(v2)
	if t1 == t2 {
		return t1
	}
	return "mixed"
}

// jsonType devuelve el nombre del tipo JSON de un valor.
func jsonType(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// sortedUnionKeys devuelve la unión ordenada de las claves de ambos objetos,
// para que el resultado sea determinista.
func sortedUnionKeys(m1, m2 map[string]interface{}) []string {
	set := make(map[string]struct{})
	for key := range m1 {
		set[key] = struct{}{}
	}
	for key := range m2 {
		set[key] = struct{}{}
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Función para comparar JSONs
func (s *Service) compareJSON(m1, m2 map[string]interface{}, prefix string, differences *[]dtos.BodyDifference) {
	for _, key := range sortedUnionKeys(m1, m2) {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		val1, ok1 := m1[key]
		val2, ok2 := m2[key]

		switch {
		case ok1 && ok2:
			s.compareBodyValues(val1, val2, fullKey, differences)
		case ok1:
			s.addDiff(differences, fullKey, "missing", val1, "key not found in second JSON")
		default:
			s.addDiff(differences, fullKey, "missing", "key not found in first JSON", val2)
		}
	}
}

// Función para comparar arrays
func (s *Service) compareArray(arr1, arr2 []interface{}, prefix string, differences *[]dtos.BodyDifference) {
	if len(arr1) != len(arr2) {
		s.addDiff(differences, prefix, "array", "different lengths", len(arr1), len(arr2))
	}

	n := len(arr1)
	if len(arr2) < n {
		n = len(arr2)
	}
	for i := 0; i < n; i++ {
		s.compareBodyValues(arr1[i], arr2[i], fmt.Sprintf("%s[%d]", prefix, i), differences)
	}
}
