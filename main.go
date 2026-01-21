package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK from Dealna API"))
	})

	http.ListenAndServe(":8080", nil)
}
