package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"ris/worker/model"
	"strings"
	"sync"
)

func processTask(hash string, maxLength int, alphabet []string, partNumber, partCount int) []string {
	n := len(alphabet)
	total := 0
	for l := 1; l <= maxLength; l++ {
		total += int(math.Pow(float64(n), float64(l)))
	}

	startIndex := total * (partNumber - 1) / partCount
	endIndex := total * partNumber / partCount

	targetHash := strings.ToLower(hash)
	var results []string

	rangeSize := endIndex - startIndex
	segSize := rangeSize / partCount
	if segSize < 1 {
		segSize = 1
	}

	var wg sync.WaitGroup
	resChan := make(chan string, 100)

	for i := 0; i < partCount; i++ {
		segStart := startIndex + i*segSize
		segEnd := segStart + segSize
		if i == partCount-1 {
			segEnd = endIndex
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for idx := s; idx < e; idx++ {
				word := indexToWord(idx, maxLength, alphabet)
				if word == "" {
					continue
				}
				hashBytes := md5.Sum([]byte(word))
				hashStr := hex.EncodeToString(hashBytes[:])
				if hashStr == targetHash {
					resChan <- word
				}
			}
		}(segStart, segEnd)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for w := range resChan {
		results = append(results, w)
	}

	return results
}

func indexToWord(index int, maxLength int, alphabet []string) string {
	n := len(alphabet)
	sum := 0
	for l := 1; l <= maxLength; l++ {
		count := intPow(len(alphabet), l)
		if index < sum+count {
			rank := index - sum
			word := ""
			for i := 0; i < l; i++ {
				power := intPow(n, l-i-1)
				pos := rank / power
				word += alphabet[pos]
				rank = rank % power
			}
			return word
		}
		sum += count
	}
	return ""
}

func intPow(a, b int) int {
	result := 1
	for i := 0; i < b; i++ {
		result *= a
	}
	return result
}

func startTask(task model.CrackHashManagerRequest) {
	alphabet := task.Alphabet.Symbols
	results := processTask(task.Hash, task.MaxLength, alphabet, task.PartNumber, task.PartCount)

	managerUrl := fmt.Sprintf("http://ris-manager-1:8080/internal/api/manager/hash/crack/request")
	workerResult := model.WorkerResult{
		RequestID: task.RequestId,
		Word:      strings.Join(results, ","),
	}
	log.Printf("Worker result: %v", workerResult)

	jsonData, err := json.Marshal(workerResult)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	req, err := http.NewRequest("PATCH", managerUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending result to manager: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Result sent to manager: %s", workerResult.Word)
}
