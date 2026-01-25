package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

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
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("DB CONNECTED")
	return db
}

ffunc main() {
	db := connectDB()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Dealna backend running"))
	})

	http.HandleFunc("/db-test", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, note FROM health_test ORDER BY id DESC LIMIT 5")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type Row struct {
			ID   int    `json:"id"`
			Note string `json:"note"`
		}

		var results []Row

		for rows.Next() {
			var r Row
			if err := rows.Scan(&r.ID, &r.Note); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			results = append(results, r)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%v", results)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

