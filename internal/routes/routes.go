package routes

import (
	"net/http"

	"comparator/internal/api"
)

func SetupRoutes(server *http.ServeMux, handler *api.Handler) {
	server.HandleFunc("POST /compare", handler.CompareHandler)
}
