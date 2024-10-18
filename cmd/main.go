package main

import (
	"log"
	"net/http"

	"comparator/internal/comparator/api"
)

func main() {

	http.HandleFunc("/compare", api.CompareHandler)
	
	log.Println("Iniciando servidor en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
