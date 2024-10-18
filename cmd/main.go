package main

import (
	"log"
	"net/http"

	"comparator/config"
	"comparator/internal/comparator/api"
	"comparator/internal/comparator/comparator"
)

func main() {

	server := http.NewServeMux()

	client := &http.Client{} //TODO create struct

	handler := api.NewHandler(comparator.NewComparatorService(client))

	server.HandleFunc("POST /compare", handler.CompareHandler)

	log.Println("Listening on " + config.GetPort())
	log.Fatal(http.ListenAndServe(config.GetPort(), server))
}
