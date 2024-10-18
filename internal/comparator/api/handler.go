package api

import (
	"encoding/json"
	"net/http"

	"comparator/internal/comparator/comparator"
)

// Estructura para recibir datos desde el frontend
type CompareRequest struct {
	URL1 string `json:"url1"`
	URL2 string `json:"url2"`
}

func CompareHandler(w http.ResponseWriter, r *http.Request) {

	var req CompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	differences := comparator.CompareURLs(req.URL1, req.URL2)

	response := comparator.CompareResponse{
		StatusCodes:     differences.StatusCodes,
		Headers:         differences.Headers,
		BodyDifferences: differences.BodyDifferences,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
