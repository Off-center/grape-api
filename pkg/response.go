package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type APIError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func WriteWithStatus(w http.ResponseWriter, status int, data any) {
	jsonData, err := json.Marshal(data)

	if err != nil {
		log.Printf("failed to marshal response: %v", err)
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(jsonData); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteWithStatus(w, status, APIError{
		Message: message,
		Status:  status,
	})
}

func GetIDFromPath(r *http.Request) (int64, error) {
	rawID := r.PathValue("id")

	id, err := strconv.ParseInt(rawID, 10, 64)

	if err != nil {
		return 0, fmt.Errorf("invalid todo id %q: %w", rawID, err)
	}

	if id <= 0 {
		return 0, fmt.Errorf("invalid todo id %d", id)
	}

	return id, nil
}
