package parser

import (
	"regexp"
	"strings"
)

// TerraformResource represents a parsed Terraform resource
type TerraformResource struct {
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Properties map[string]string `json:"properties"`
}

// TerraformConfig represents the parsed Terraform configuration
type TerraformConfig struct {
	Resources []TerraformResource `json:"resources"`
	Variables []string             `json:"variables"`
	Outputs   []string             `json:"outputs"`
}

// ParseTerraform parses Terraform code and returns structured data
func ParseTerraform(terraformCode string) (*TerraformConfig, error) {
	config := &TerraformConfig{
		Resources: []TerraformResource{},
		Variables: []string{},
		Outputs:   []string{},
	}

	// Extract resources
	resourcePattern := regexp.MustCompile(`resource\s+"([^"]+)"\s+"([^"]+)"\s*\{`)
	resourceMatches := resourcePattern.FindAllStringSubmatch(terraformCode, -1)

	for _, match := range resourceMatches {
		if len(match) >= 3 {
			resourceType := match[1]
			resourceName := match[2]

			// Extract resource block
			resourceBlock := extractResourceBlock(terraformCode, match[0])
			properties := extractProperties(resourceBlock)

			config.Resources = append(config.Resources, TerraformResource{
				Type:       resourceType,
				Name:       resourceName,
				Properties: properties,
			})
		}
	}

	// Extract variables
	variablePattern := regexp.MustCompile(`variable\s+"([^"]+)"`)
	variableMatches := variablePattern.FindAllStringSubmatch(terraformCode, -1)
	for _, match := range variableMatches {
		if len(match) >= 2 {
			config.Variables = append(config.Variables, match[1])
		}
	}

	// Extract outputs
	outputPattern := regexp.MustCompile(`output\s+"([^"]+)"`)
	outputMatches := outputPattern.FindAllStringSubmatch(terraformCode, -1)
	for _, match := range outputMatches {
		if len(match) >= 2 {
			config.Outputs = append(config.Outputs, match[1])
		}
	}

	return config, nil
}

// extractResourceBlock extracts the full resource block from the Terraform code
func extractResourceBlock(code, startMatch string) string {
	startIdx := strings.Index(code, startMatch)
	if startIdx == -1 {
		return ""
	}

	// Find the matching closing brace
	braceCount := 0
	inString := false
	escapeNext := false

	for i := startIdx; i < len(code); i++ {
		char := code[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		if !inString {
			if char == '{' {
				braceCount++
			} else if char == '}' {
				braceCount--
				if braceCount == 0 {
					return code[startIdx : i+1]
				}
			}
		}
	}

	return code[startIdx:]
}

// extractProperties extracts key-value properties from a resource block
func extractProperties(block string) map[string]string {
	properties := make(map[string]string)

	// Simple pattern to extract key = value pairs
	// This is a basic implementation - a full parser would be more robust
	keyValuePattern := regexp.MustCompile(`(\w+)\s*=\s*"([^"]*)"`)
	matches := keyValuePattern.FindAllStringSubmatch(block, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			key := match[1]
			value := match[2]
			properties[key] = value
		}
	}

	return properties
}

