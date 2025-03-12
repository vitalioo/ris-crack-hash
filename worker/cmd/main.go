package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/internal/api/worker/hash/crack/task", handleTask).Methods("POST")
	r.HandleFunc("/health", healthCheck).Methods("GET")
	log.Println("Starting worker on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", r))
}
