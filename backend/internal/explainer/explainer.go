package explainer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"infraexplain/internal/parser"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ExplainTerraform calls an API to explain the structured Terraform configuration
func ExplainTerraform(config *parser.TerraformConfig) (string, error) {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Fallback to a simple explanation if no API key is provided
		return generateSimpleExplanation(config), nil
	}

	// Prepare the prompt
	prompt := buildPrompt(config)

	// Call OpenAI API
	explanation, err := callOpenAI(apiKey, prompt)
	if err != nil {
		// Fallback to simple explanation on error
		return generateSimpleExplanation(config), nil
	}

	return explanation, nil
}

// buildPrompt creates a user-friendly prompt for explaining Terraform
func buildPrompt(config *parser.TerraformConfig) string {
	var prompt strings.Builder
	prompt.WriteString("Explain this Terraform configuration in simple, beginner-friendly terms:\n\n")

	if len(config.Resources) > 0 {
		prompt.WriteString("Resources:\n")
		for _, resource := range config.Resources {
			prompt.WriteString(fmt.Sprintf("- %s.%s", resource.Type, resource.Name))
			if len(resource.Properties) > 0 {
				prompt.WriteString(" with properties: ")
				first := true
				for key := range resource.Properties {
					if !first {
						prompt.WriteString(", ")
					}
					prompt.WriteString(key)
					first = false
				}
			}
			prompt.WriteString("\n")
		}
	}

	if len(config.Variables) > 0 {
		prompt.WriteString("\nVariables: ")
		for i, v := range config.Variables {
			if i > 0 {
				prompt.WriteString(", ")
			}
			prompt.WriteString(v)
		}
		prompt.WriteString("\n")
	}

	if len(config.Outputs) > 0 {
		prompt.WriteString("\nOutputs: ")
		for i, o := range config.Outputs {
			if i > 0 {
				prompt.WriteString(", ")
			}
			prompt.WriteString(o)
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("\nProvide a clear, concise explanation suitable for someone new to infrastructure as code.")
	return prompt.String()
}

// callOpenAI calls the OpenAI API to get an explanation
func callOpenAI(apiKey, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a helpful assistant that explains Terraform configurations in simple, beginner-friendly terms.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens": 500,
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", err
	}

	if len(apiResponse.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return apiResponse.Choices[0].Message.Content, nil
}

// generateSimpleExplanation provides a fallback explanation without API
func generateSimpleExplanation(config *parser.TerraformConfig) string {
	var explanation strings.Builder
	explanation.WriteString("This Terraform configuration defines the following:\n\n")

	if len(config.Resources) > 0 {
		explanation.WriteString("**Resources:**\n")
		for _, resource := range config.Resources {
			explanation.WriteString(fmt.Sprintf("- A %s resource named '%s'", resource.Type, resource.Name))
			if len(resource.Properties) > 0 {
				explanation.WriteString(" with configured properties")
			}
			explanation.WriteString(".\n")
		}
		explanation.WriteString("\n")
	}

	if len(config.Variables) > 0 {
		explanation.WriteString(fmt.Sprintf("**Variables:** %d input variable(s) that can be customized.\n\n", len(config.Variables)))
	}

	if len(config.Outputs) > 0 {
		explanation.WriteString(fmt.Sprintf("**Outputs:** %d output value(s) that provide information about the infrastructure.\n", len(config.Outputs)))
	}

	return explanation.String()
}

