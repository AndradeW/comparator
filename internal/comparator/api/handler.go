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
	CompareRequest(request dtos.Request) (diff dtos.CompareResponse, error error)
}

func (h *Handler) CompareHandler(w http.ResponseWriter, r *http.Request) {

	var req dtos.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	differences, err := h.service.CompareRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //TODO validar si todo es 500
		return
	}

	response := dtos.CompareResponse{
		StatusCodes:     differences.StatusCodes,
		Headers:         differences.Headers,
		BodyDifferences: differences.BodyDifferences,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
