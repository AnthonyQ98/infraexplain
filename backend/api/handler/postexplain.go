package handler

import (
	"encoding/json"
	"infraexplain/internal/explainer"
	"infraexplain/internal/parser"
	"log"
	"net/http"
)

type ExplainRequest struct {
	TextContent string `json:"text_content"`
}

type ExplainResponse struct {
	Summary string `json:"summary"`
}

func PostExplain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExplainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Parse Terraform code to structured data
	config, err := parser.ParseTerraform(req.TextContent)
	if err != nil {
		log.Printf("Error parsing Terraform: %v", err)
		http.Error(w, "failed to parse Terraform code", http.StatusInternalServerError)
		return
	}

	// Call API to explain the structured Terraform
	explanation, err := explainer.ExplainTerraform(config)
	if err != nil {
		log.Printf("Error explaining Terraform: %v", err)
		http.Error(w, "failed to generate explanation", http.StatusInternalServerError)
		return
	}

	// Respond with the explanation
	resp := ExplainResponse{
		Summary: explanation,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
