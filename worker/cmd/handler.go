package main

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"ris/worker/model"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleTask(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var task model.CrackHashManagerRequest
	err = xml.Unmarshal(body, &task)
	if err != nil {
		log.Printf("Error parsing XML: %v", err)
		http.Error(w, "Invalid XML format", http.StatusBadRequest)
		return
	}

	log.Printf("Received task: RequestID=%s, PartNumber=%d, PartCount=%d, Hash=%s, MaxLength=%d",
		task.RequestId, task.PartNumber, task.PartCount, task.Hash, task.MaxLength)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task received and is being processed"))

	go startTask(task)
}
