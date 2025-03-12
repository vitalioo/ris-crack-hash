package main

import (
	"os"
	"strconv"
	"time"
)

var taskQueue chan struct{}

func initializeTaskQueue() {
	queueSize := os.Getenv("QUEUE_SIZE")
	queueSizeInt, _ := strconv.Atoi(queueSize)
	taskQueue = make(chan struct{}, queueSizeInt)
	for i := 0; i < queueSizeInt; i++ {
		taskQueue <- struct{}{}
	}
}

func cleanupOldTasks() {
	for {
		updateTable()
		time.Sleep(24 * time.Hour)
	}
}

func run() {
	initializeTaskQueue()
	initDB()
	createTable()
	go cleanupOldTasks()
}
