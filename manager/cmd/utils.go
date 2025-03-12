package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"ris/manager/model"
	"strconv"
	"strings"
	"time"
)

func checkWorkerHealth(workerID int) bool {
	workerPort := os.Getenv("WORKER_PORT")
	resp, err := http.Get(fmt.Sprintf("http://ris-worker-%d:"+workerPort+"/health", workerID))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func monitorWorkers(requestId string, workerCount int) {
	var completedParts int
	var status string
	for {
		completedParts = countOfCompletedWorkers(requestId)
		allWorkersHealthy := true
		for i := 1; i <= workerCount; i++ {
			if !checkWorkerHealth(i) {
				allWorkersHealthy = false
				break
			}
		}
		if allWorkersHealthy {
			if completedParts == workerCount {
				status = model.PARTITION_READY
			} else {
				time.Sleep(10 * time.Second)
				continue
			}
		} else {
			if completedParts > 0 {
				status = model.PARTITION_READY
			} else {
				status = model.ERROR
			}
		}
		updateTaskStatus(requestId, status)
		break
	}
}

func processTask(requestId string, hash string, maxLength int) {
	workerCount := getWorkerCount()
	alphabet := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s",
		"t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	workerPort := os.Getenv("WORKER_PORT")
	for i := 1; i <= workerCount; i++ {
		xmlRequest := generateXMLRequest(requestId, hash, maxLength, i, workerCount, alphabet)
		workerUrl := fmt.Sprintf("http://ris-worker-%d:"+workerPort+"/internal/api/worker/hash/crack/task", i)
		sendTask(xmlRequest, workerUrl)
	}

	go monitorWorkers(requestId, workerCount)

	go func() {
		time.Sleep(2 * time.Minute)
		status, _ := getHashStatusById(requestId)
		if status == model.IN_PROGRESS {
			updateTaskStatus(requestId, "ERROR")
		}
	}()
}

func sendTask(xmlRequest, workerUrl string) {
	log.Printf("Sending task to worker: %v", workerUrl)
	response, err := http.Post(workerUrl, "text/xml", bytes.NewBufferString(xmlRequest))
	if err != nil {
		log.Printf("Error sending task to %s: %v", workerUrl, err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}(response.Body)
}

func getWorkerCount() int {
	workerCountStr := os.Getenv("WORKER_COUNT")
	if workerCountStr == "" {
		log.Println("WORKER_COUNT is not set, using default value: 2")
		return 2
	}

	workerCount, err := strconv.Atoi(workerCountStr)
	if err != nil {
		log.Printf("Invalid WORKER_COUNT value: %s, using default value: 2", workerCountStr)
		return 2
	}

	return workerCount
}

func generateXMLRequest(requestId, hash string, maxLength, partNumber, partCount int, alphabet []string) string {
	managerRequest := model.HashCrackManagerRequest{
		RequestId:  requestId,
		PartNumber: partNumber,
		PartCount:  partCount,
		Hash:       hash,
		MaxLength:  maxLength,
		Alphabet:   model.Alphabet{Symbols: alphabet},
	}
	xmlData, err := xml.MarshalIndent(managerRequest, " ", "")
	if err != nil {
		log.Printf("Error marshaling XML: %v", err)
		return ""
	}
	return string(xmlData)
}

func cleanData(data string) []string {
	if data == "" {
		return nil
	}

	parts := strings.Split(data, ",")

	var cleaned []string
	for _, part := range parts {
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}

	return cleaned
}
