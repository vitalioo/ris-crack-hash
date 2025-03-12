package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	run()

	r := mux.NewRouter()
	r.HandleFunc("/internal/api/manager/hash/crack", crackHash).Methods("POST")
	r.HandleFunc("/internal/api/manager/hash/status", getHashStatus).Methods("GET")
	r.HandleFunc("/internal/api/manager/hash/crack/request", workerResult).Methods("PATCH")
	log.Fatal(http.ListenAndServe(":8080", r))
}
