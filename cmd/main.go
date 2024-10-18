package main

import (
	"log"
	"net/http"

	"comparator/config"
	"comparator/internal/comparator/api"
)

func main() {

	server := http.NewServeMux()

	handler := api.Handler{}

	server.HandleFunc("POST /compare", handler.CompareHandler)

	log.Println("Listening on " + config.GetPort())
	log.Fatal(http.ListenAndServe(config.GetPort(), server))
}
