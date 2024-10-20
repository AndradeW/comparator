package routes

import (
	"net/http"

	"comparator/internal/api"
)

func SetUpRoutes(server *http.ServeMux, handler *api.Handler) {
	server.HandleFunc("POST /compare", handler.CompareHandler)
}
