package handler

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func GetHealth(w http.ResponseWriter, _ *http.Request) {

	resp := HealthResponse{
		Status: "available",
	}
	json.NewEncoder(w).Encode(resp)
}
