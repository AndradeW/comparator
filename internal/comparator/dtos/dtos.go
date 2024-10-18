package dtos

// Estructura para recibir datos desde el frontend
type CompareRequest struct {
	URL1 string `json:"url1"`
	URL2 string `json:"url2"`
}

// Estructura para capturar input del usuario
type RequestDetails struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Params  map[string]string `json:"params"`
}

// Estructura para almacenar las diferencias
type CompareResponse struct {
	StatusCodes     []int                    `json:"status_codes"`
	Headers         map[string][]string      `json:"headers"`
	BodyDifferences map[string][]interface{} `json:"body_differences"`
}
