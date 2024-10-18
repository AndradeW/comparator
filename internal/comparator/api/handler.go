package api

import (
	"encoding/json"
	"net/http"

	"comparator/internal/comparator/comparator"
)

type comparatorService interface {
	CompareURLs(url1 string, url2 string) comparator.CompareResponse
}

type Handler struct {
	ComparatorService comparatorService
}

func NewHandler(comparatorService comparatorService) *Handler {
	return &Handler{ComparatorService: comparatorService}
}

// Estructura para recibir datos desde el frontend
type CompareRequest struct {
	URL1 string `json:"url1"`
	URL2 string `json:"url2"`
}

func (h *Handler) CompareHandler(w http.ResponseWriter, r *http.Request) {

	var req CompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	differences := h.ComparatorService.CompareURLs(req.URL1, req.URL2)

	response := comparator.CompareResponse{
		StatusCodes:     differences.StatusCodes,
		Headers:         differences.Headers,
		BodyDifferences: differences.BodyDifferences,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
