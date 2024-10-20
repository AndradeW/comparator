package comparator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

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
		fmt.Println("Error en la petición 1:", err)
		return dtos.CompareResponse{}, fmt.Errorf("error en la petición 1 : %s", err)
	}

	response2, err := s.makeRequest(request.Request2)
	if err != nil {
		fmt.Println("Error en la petición 2:", err)
		return dtos.CompareResponse{}, fmt.Errorf("error en la petición 2 : %s", err)
	}

	differences := s.compareResponses(response1, response2)

	return differences, nil
}

// Función para realizar la petición HTTP
func (s *Service) makeRequest(reqDetails dtos.RequestDetails) (*http.Response, error) {

	// Construir la URL con parámetros
	req, err := http.NewRequest("GET", reqDetails.URL, nil)
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
func (s *Service) compareResponses(resp1, resp2 *http.Response) dtos.CompareResponse {
	differences := dtos.CompareResponse{
		Headers:         make(map[string][]string),
		BodyDifferences: make(map[string][]interface{}),
	}

	// Comparar los códigos de estado
	if resp1.StatusCode != resp2.StatusCode {
		differences.StatusCodes = []int{resp1.StatusCode, resp2.StatusCode}
	}

	// Comparar los headers
	for key, val1 := range resp1.Header {
		val2 := resp2.Header.Get(key)
		if !reflect.DeepEqual(val1, []string{val2}) {
			differences.Headers[key] = []string{val1[0], val2}
		}
	}

	// Comparar los cuerpos de la respuesta (asumiendo que son JSON)
	body1, _ := io.ReadAll(resp1.Body)
	body2, _ := io.ReadAll(resp2.Body)

	var json1, json2 map[string]interface{}
	err1 := json.Unmarshal(body1, &json1)
	err2 := json.Unmarshal(body2, &json2)

	if err1 != nil || err2 != nil {
		// Si hay error al parsear JSON, agregar los cuerpos completos a las diferencias
		differences.BodyDifferences["error"] = []interface{}{string(body1), string(body2)}
	} else {
		// Comparar los JSON
		s.compareJSON(json1, json2, "", differences.BodyDifferences)
	}

	return differences
}

// Función para comparar JSONs
func (s *Service) compareJSON(json1, json2 map[string]interface{}, prefix string, differences map[string][]interface{}) {
	for key, val1 := range json1 {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if val2, ok := json2[key]; ok {
			// Si la clave existe en ambos, comparar los valores
			if !reflect.DeepEqual(val1, val2) { //TODO Extender a json anidados y arrays
				differences[fullKey] = []interface{}{val1, val2}
			}
		} else {
			// Si la clave solo está en json1
			differences[fullKey] = []interface{}{val1, "key not found in second JSON"}
		}
	}

	// Verificar claves que están en json2 pero no en json1
	for key, val2 := range json2 {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if _, ok := json1[key]; !ok {
			differences[fullKey] = []interface{}{"key not found in first JSON", val2}
		}
	}
}
