package evaluator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Assertion struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
	Path  string `yaml:"path"`
}

type AssertResult struct {
	Type   string
	Passed bool
	Reason string
}

func Validate(assertion Assertion, output string, latency time.Duration) AssertResult {
	switch assertion.Type {
	case "json-valid":
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(output), &js); err != nil {
			return AssertResult{Type: assertion.Type, Passed: false, Reason: "Malformed JSON"}
		}
		return AssertResult{Type: assertion.Type, Passed: true}

	case "contains":
		if !strings.Contains(output, assertion.Value) {
			return AssertResult{
				Type:   assertion.Type,
				Passed: false,
				Reason: fmt.Sprintf("Expected output to contain %q", assertion.Value),
			}
		}
		return AssertResult{Type: assertion.Type, Passed: true}

	case "regex":
		matched, err := regexp.MatchString(assertion.Value, output)
		if err != nil || !matched {
			return AssertResult{
				Type:   assertion.Type,
				Passed: false,
				Reason: fmt.Sprintf("Pattern %q failed to match", assertion.Value),
			}
		}
		return AssertResult{Type: assertion.Type, Passed: true}

	case "json-field-equals":
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(output), &js); err != nil {
			return AssertResult{Type: assertion.Type, Passed: false, Reason: "Malformed JSON"}
		}
		val, exists := js[assertion.Path]
		if !exists || fmt.Sprintf("%v", val) != assertion.Value {
			return AssertResult{
				Type:   assertion.Type,
				Passed: false,
				Reason: fmt.Sprintf("Field %q expected %q, got %v", assertion.Path, assertion.Value, val),
			}
		}
		return AssertResult{Type: assertion.Type, Passed: true}

	default:
		return AssertResult{Type: assertion.Type, Passed: false, Reason: "Unknown assertion type"}
	}
}
