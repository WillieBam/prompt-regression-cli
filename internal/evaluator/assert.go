package evaluator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
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
		cleaned := cleanJSON(output)
		var js any
		if err := json.Unmarshal([]byte(cleaned), &js); err != nil {
			return AssertResult{Type: assertion.Type, Passed: false, Reason: "Malformed JSON: " + err.Error()}
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
		cleaned := cleanJSON(output)
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(cleaned), &js); err != nil {
			return AssertResult{Type: assertion.Type, Passed: false, Reason: "Malformed JSON: " + err.Error()}
		}
		val, exists := extractJSON(js, assertion.Path)
		if !exists || fmt.Sprintf("%v", val) != assertion.Value {
			return AssertResult{
				Type:   assertion.Type,
				Passed: false,
				Reason: fmt.Sprintf("Field %q expected %q, got %v", assertion.Path, assertion.Value, val),
			}
		}
		return AssertResult{Type: assertion.Type, Passed: true}

	case "max-latency", "latency-less-than":
		maxDur, err := time.ParseDuration(assertion.Value)
		if err != nil {
			return AssertResult{
				Type:   assertion.Type,
				Passed: false,
				Reason: fmt.Sprintf("Invalid duration %q: %v", assertion.Value, err),
			}
		}

		if latency > maxDur {
			return AssertResult{
				Type:   assertion.Type,
				Passed: false,
				Reason: fmt.Sprintf("Latency exceed maximum allowed: %v > %v", latency, maxDur),
			}
		}
		return AssertResult{Type: assertion.Type, Passed: true}

	default:
		return AssertResult{Type: assertion.Type, Passed: false, Reason: "Unknown assertion type"}
	}
}

func extractJSON(obj any, path string) (any, bool) {
	if path == "" {
		return obj, true
	}

	// [llm, metrics, latency]
	parts := strings.Split(path, ".")
	current := obj
	for _, part := range parts {
		// any type in this case is current
		// compiler knows any as it has no key or able to perform indexing
		// in order to perform indexing for current e.g. val[idx]
		switch val := current.(type) {
		case map[string]any:
			next, ok := val[part]
			if !ok {
				return nil, false
			}
			current = next
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(val) {
				return nil, false
			}

			current = val[idx]
		default:
			return nil, false

		}
	}
	return current, true

}

func cleanJSON(s string) string {
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "```") {
		lines := strings.Split(trimmed, "\n")

		// check line len >=2 and content has prefix/suffix with ```
		if len(lines) >= 2 && strings.HasPrefix(lines[0], "```") && strings.HasSuffix(strings.TrimSpace(lines[len(lines)-1]), "```") {
			return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
		}
	}
	return trimmed
}
