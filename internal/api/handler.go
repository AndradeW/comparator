package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"comparator/config"
	"comparator/internal/comparator"
	"comparator/internal/dtos"
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
	r.Body = http.MaxBytesReader(w, r.Body, config.GetMaxBodySize())

	var req dtos.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	differences, err := h.service.CompareRequest(req)
	if err != nil {
		var validationErr *comparator.ValidationError
		if errors.As(err, &validationErr) {
			http.Error(w, validationErr.Error(), http.StatusBadRequest)
			return
		}
		slog.Error("error al comparar las peticiones", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError) //TODO distinguir errores upstream (502)
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
