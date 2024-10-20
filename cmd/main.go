package main

import (
	"log"
	"net/http"

	"comparator/config"
	"comparator/internal/api"
	"comparator/internal/comparator"
	"comparator/internal/httpclient"
	"comparator/internal/routes"
)

func main() {

	client := httpclient.NewHttpclient() //TODO create struct
	handler := api.NewHandler(comparator.NewComparatorService(client))

	server := http.NewServeMux()
	routes.SetUpRoutes(server, handler)

	log.Println("Listening on " + config.GetPort())
	log.Fatal(http.ListenAndServe(config.GetPort(), server))
}
