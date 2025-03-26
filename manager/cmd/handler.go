package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ris/manager/model"
)

func crackHash(w http.ResponseWriter, r *http.Request) {
	select {
	case <-taskQueue:
		defer func() {
			taskQueue <- struct{}{}
		}()
	default:
		http.Error(w, "Too many requests. Queue limit exceeded.", http.StatusTooManyRequests)
		return
	}

	var hashCrackRequest model.HashCrackRequest
	err := json.NewDecoder(r.Body).Decode(&hashCrackRequest)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	requestId := createTask(hashCrackRequest.Hash, hashCrackRequest.MaxLength)
	go processTask(requestId, hashCrackRequest.Hash, hashCrackRequest.MaxLength)

	err = json.NewEncoder(w).Encode(model.HashCrackResponse{RequestID: requestId})
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func getHashStatus(w http.ResponseWriter, r *http.Request) {
	var hashStatusRequest model.HashStatusRequest
	err := json.NewDecoder(r.Body).Decode(&hashStatusRequest)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	status, data := getHashStatusById(hashStatusRequest.RequestID)
	cleanedData := cleanData(data)

	response := model.HashStatusResponse{
		Status: status,
		Data:   nil,
	}

	if status == model.READY || status == model.PARTITION_READY {
		response.Data = cleanedData
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func workerResult(w http.ResponseWriter, r *http.Request) {
	var workerResult model.WorkerResult
	err := json.NewDecoder(r.Body).Decode(&workerResult)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	log.Printf("Worker model result: %v", workerResult)

	appendTaskData(workerResult.RequestID, workerResult.Word)

	workerCount := getWorkerCount()
	if countOfCompletedWorkers(workerResult.RequestID) == workerCount {
		updateTaskStatus(workerResult.RequestID, "READY")
	}
}
