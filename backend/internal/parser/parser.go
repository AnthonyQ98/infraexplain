package parser

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// TerraformResource represents a parsed resource
type TerraformResource struct {
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Properties map[string]string `json:"properties"`
}

// TerraformConfig represents parsed TF code
type TerraformConfig struct {
	Resources []TerraformResource `json:"resources"`
	Variables []string            `json:"variables"`
	Outputs   []string            `json:"outputs"`
	Issues    []Issue             `json:"issues"`
}

// Issue represents a security or improvement finding
type Issue struct {
	Type    string `json:"type"` // "security" or "improvement"
	Message string `json:"message"`
}

// ParseTerraform parses Terraform HCL code and returns structured data with basic issues
func ParseTerraform(terraformCode string) (*TerraformConfig, error) {
	config := &TerraformConfig{
		Resources: []TerraformResource{},
		Variables: []string{},
		Outputs:   []string{},
		Issues:    []Issue{},
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(terraformCode), "input.tf")
	if diags.HasErrors() {
		return config, fmt.Errorf("failed to parse HCL: %v", diags)
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return config, fmt.Errorf("unexpected body type")
	}

	// Walk through blocks
	for _, block := range body.Blocks {
		switch block.Type {
		case "resource":
			if len(block.Labels) < 2 {
				continue
			}
			resType := block.Labels[0]
			resName := block.Labels[1]
			props := extractProperties(block.Body)

			config.Resources = append(config.Resources, TerraformResource{
				Type:       resType,
				Name:       resName,
				Properties: props,
			})

			// Basic security checks
			checkResourceSecurity(resType, resName, props, &config.Issues)
			log.Printf("issues: %v", config.Issues)

		case "variable":
			if len(block.Labels) >= 1 {
				config.Variables = append(config.Variables, block.Labels[0])
			}
		case "output":
			if len(block.Labels) >= 1 {
				config.Outputs = append(config.Outputs, block.Labels[0])
			}
		}
	}

	return config, nil
}

// extractProperties converts HCL attributes into a simple key=value map
func extractProperties(body *hclsyntax.Body) map[string]string {
	props := make(map[string]string)
	for name, attr := range body.Attributes {
		val, diag := attr.Expr.Value(nil)
		if diag.HasErrors() {
			props[name] = "<unknown>"
			continue
		}

		// Safely convert any type to string
		switch val.Type().FriendlyName() {
		case "string":
			props[name] = val.AsString()
		case "list", "tuple":
			elems := make([]string, val.LengthInt())
			for i := 0; i < val.LengthInt(); i++ {
				elem := val.Index(cty.NumberIntVal(int64(i)))
				elems[i] = elem.GoString()
			}
			props[name] = "[" + strings.Join(elems, ", ") + "]"
		case "bool":
			props[name] = fmt.Sprintf("%v", val.True())
		case "number":
			props[name] = val.GoString()
		default:
			props[name] = val.GoString()
		}
	}
	return props
}

// checkResourceSecurity adds basic security/improvement issues
func checkResourceSecurity(resType, resName string, props map[string]string, issues *[]Issue) {
	switch resType {
	case "aws_security_group":
		if cidr, ok := props["cidr_blocks"]; ok {
			if strings.Contains(cidr, "0.0.0.0/0") {
				*issues = append(*issues, Issue{
					Type:    "security",
					Message: fmt.Sprintf("Security group '%s' allows 0.0.0.0/0 ingress", resName),
				})
			}
		}
	case "aws_iam_role":
		if policy, ok := props["assume_role_policy"]; ok && strings.Contains(policy, "*") {
			*issues = append(*issues, Issue{
				Type:    "security",
				Message: fmt.Sprintf("IAM role '%s' grants '*' permissions", resName),
			})
		}
	default:
		// Improvement suggestions for all resources
		if len(props) == 0 {
			*issues = append(*issues, Issue{
				Type:    "improvement",
				Message: fmt.Sprintf("Resource '%s' has no properties defined", resName),
			})
		}
	}
}
