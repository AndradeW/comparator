package dtos

// Estructura para recibir datos desde el frontend
type Request struct {
	Request1 RequestDetails `json:"request1"`
	Request2 RequestDetails `json:"request2"`
}

// Estructura para capturar input del usuario
type RequestDetails struct {
	URL     string            `json:"url" validate:"required"` //TODO rev
	Headers map[string]string `json:"headers"`
	Params  map[string]string `json:"params"`
}

// Estructura para almacenar las diferencias
type CompareResponse struct {
	StatusCodes     []int                    `json:"status_codes"`
	Headers         map[string][]string      `json:"headers"`
	BodyDifferences map[string][]interface{} `json:"body_differences"`
}
