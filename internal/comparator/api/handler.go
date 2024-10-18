package api

import (
	"encoding/json"
	"net/http"

	"comparator/internal/comparator/dtos"
)

type Handler struct {
	service comparatorService
}

func NewHandler(comparatorService comparatorService) *Handler {
	return &Handler{service: comparatorService}
}

type comparatorService interface {
	CompareRequest(request dtos.Request) dtos.CompareResponse
}

func (h *Handler) CompareHandler(w http.ResponseWriter, r *http.Request) {

	var req dtos.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	differences := h.service.CompareRequest(req)

	response := dtos.CompareResponse{
		StatusCodes:     differences.StatusCodes,
		Headers:         differences.Headers,
		BodyDifferences: differences.BodyDifferences,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
