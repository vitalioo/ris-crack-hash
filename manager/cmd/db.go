package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
	"time"
)

var db *sql.DB

func initDB() {
	host := os.Getenv("DB_HOST")
	portStr := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		port = 5432
		log.Printf("Invalid DB_PORT value: %s, using default: %d", portStr, port)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return
	}

	if err = db.Ping(); err != nil {
		log.Printf("Error connecting to database: %v", err)
		return
	}
	log.Println("Successfully connected to the database")
}

func createTable() {
	query := `CREATE TABLE IF NOT EXISTS tasks (
        request_id VARCHAR(36) PRIMARY KEY,
        hash VARCHAR(255) NOT NULL,
        max_length INTEGER NOT NULL,
        status VARCHAR(20) DEFAULT 'IN_PROGRESS',
        data VARCHAR(255),
        completed_parts INTEGER DEFAULT 0
        );`

	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating table: %v", err)
	}
}

func createTask(hash string, maxLength int) string {
	requestId := uuid.New().String()

	query := `INSERT INTO tasks (request_id, hash, max_length, created_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`
	_, err := db.Exec(query, requestId, hash, maxLength)
	if err != nil {
		log.Printf("Error creating task: %v", err)
		return ""
	}

	return requestId
}

func getHashStatusById(requestId string) (string, string) {
	query := `SELECT status, COALESCE(data, '') FROM tasks WHERE request_id=$1`
	row := db.QueryRow(query, requestId)

	var status string
	var data string
	err := row.Scan(&status, &data)
	if err != nil {
		log.Printf("Error getting hash status by ID: %v", err)
		return "", ""
	}

	return status, data
}

func appendTaskData(requestId, word string) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return
	}
	defer tx.Rollback()

	query := `UPDATE tasks 
				SET data = COALESCE(data || ',', '') || $1, 
    			completed_parts = completed_parts + 1 
				WHERE request_id = $2
				`
	_, err = tx.Exec(query, word, requestId)
	if err != nil {
		log.Printf("Error appending task data: %v", err)
		return
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
	}
}

func updateTaskStatus(requestId, status string) {
	query := `UPDATE tasks SET status=$1 WHERE request_id=$2`
	_, err := db.Exec(query, status, requestId)
	if err != nil {
		log.Printf("Error updating task status: %v", err)
	}
}

func countOfCompletedWorkers(requestId string) int {
	query := `SELECT completed_parts FROM tasks WHERE request_id=$1`
	row := db.QueryRow(query, requestId)

	var count int
	err := row.Scan(&count)
	if err != nil {
		log.Printf("Error counting completed workers: %v", err)
		return 0
	}

	return count
}

func updateTable() {
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	updateTable()
	query := `DELETE FROM tasks WHERE created_at < $1`
	_, err := db.Exec(query, oneWeekAgo)
	if err != nil {
		log.Printf("Error cleaning up old tasks: %v", err)
	}
}
