package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Saleem's Context !"))
	})

	http.ListenAndServe(":8080", nil)
}
