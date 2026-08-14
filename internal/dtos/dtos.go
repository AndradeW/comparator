package dtos

// Estructura para recibir datos desde el frontend
type Request struct {
	Request1 RequestDetails `json:"request1"`
	Request2 RequestDetails `json:"request2"`
}

// Estructura para capturar input del usuario
type RequestDetails struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Params  map[string]string `json:"params"`
	Body    string            `json:"body"`
}

// BodyDifference representa una diferencia puntual en el cuerpo JSON:
// la ruta donde ocurre (path), el tipo de los valores (tipo) y ambos valores.
type BodyDifference struct {
	Path   string        `json:"path"`
	Tipo   string        `json:"tipo"`
	Values []interface{} `json:"values"`
}

// Estructura para almacenar las diferencias
type CompareResponse struct {
	StatusCodes     []int             `json:"status_codes"`
	Headers         map[string][]string `json:"headers"`
	BodyDifferences []BodyDifference  `json:"body_differences"`
}
