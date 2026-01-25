package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type HealthRow struct {
	ID   int    `json:"id"`
	Note string `json:"note"`
}

type InsertRequest struct {
	Note string `json:"note"`
}

// --------------------
// Database Connection
// --------------------
func connectDB() *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("DB open error:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("DB ping error:", err)
	}

	log.Println("✅ DB CONNECTED")
	return db
}

// --------------------
// Main
// --------------------
func main() {
	db := connectDB()

	// Basic health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Dealna backend running"))
	})

	// Read from DB
	http.HandleFunc("/db-test", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(
			"SELECT id, note FROM health_test ORDER BY id DESC LIMIT 10",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var results []HealthRow

		for rows.Next() {
			var row HealthRow
			if err := rows.Scan(&row.ID, &row.Note); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	// Insert into DB
	http.HandleFunc("/db-insert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req InsertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		if req.Note == "" {
			http.Error(w, "note field is required", http.StatusBadRequest)
			return
		}

		_, err := db.Exec(
			"INSERT INTO health_test(note) VALUES ($1)",
			req.Note,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Row inserted successfully"))
	})

	log.Println("🚀 Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
