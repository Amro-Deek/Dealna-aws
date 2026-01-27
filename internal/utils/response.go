package utils

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, success bool, message string, data interface{}, err interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	var errStr interface{}
	if e, ok := err.(error); ok {
		errStr = e.Error()
	} else {
		errStr = err
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: success,
		Message: message,
		Data:    data,
		Error:   errStr,
	})
}
