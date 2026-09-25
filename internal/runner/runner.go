package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"sync"
	"text/template"
	"time"

	"prompt-regression-cli/internal/evaluator"
	"prompt-regression-cli/internal/provider"

	"gopkg.in/yaml.v3"
)

type SuiteConfig struct {
	Version  string     `yaml:"version"`
	Name     string     `yaml:"name"`
	Model    string     `yaml:"model"`
	Template string     `yaml:"template"`
	Tests    []TestCase `yaml:"tests"`
}

type TestCase struct {
	ID      string                `yaml:"id"`
	Inputs  map[string]string     `yaml:"inputs"`
	Asserts []evaluator.Assertion `yaml:"assert"`
}

type RunResult struct {
	ID       string
	Passed   bool
	Output   string
	Duration time.Duration
	Failures []string
	Err      error
}

func (r RunResult) MarshalJSON() ([]byte, error) {
	type Alias RunResult
	var errStr *string
	if r.Err != nil {
		s := r.Err.Error()
		errStr = &s
	}
	return json.Marshal(&struct {
		Alias
		Err *string `json:"Err"`
	}{
		Alias: Alias(r),
		Err:   errStr,
	})
}

func LoadSuite(path string) (*SuiteConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var suite SuiteConfig
	if err := yaml.Unmarshal(data, &suite); err != nil {
		return nil, err
	}
	return &suite, nil
}

func ExecuteSuite(ctx context.Context, suite *SuiteConfig, p provider.ModelProvider, concurrency int) <-chan RunResult {
	out := make(chan RunResult, len(suite.Tests))
	jobs := make(chan TestCase, len(suite.Tests))
	var wg sync.WaitGroup

	// Render prompt template
	tmpl, _ := template.New("prompt").Parse(suite.Template)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tc := range jobs {
				out <- runSingle(ctx, tc, tmpl, suite.Model, p)
			}
		}()
	}

	go func() {
		for _, tc := range suite.Tests {
			jobs <- tc
		}
		close(jobs)
		wg.Wait()
		close(out)
	}()

	return out
}

func runSingle(ctx context.Context, tc TestCase, tmpl *template.Template, model string, p provider.ModelProvider) RunResult {
	start := time.Now()

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tc.Inputs); err != nil {
		return RunResult{ID: tc.ID, Passed: false, Err: err}
	}

	resp, err := p.Complete(ctx, provider.CompletionRequest{
		Model:  model,
		Prompt: buf.String(),
	})
	if err != nil {
		return RunResult{ID: tc.ID, Passed: false, Err: err, Duration: time.Since(start)}
	}

	var failures []string
	for _, a := range tc.Asserts {
		res := evaluator.Validate(a, resp.Text, resp.Metrics.Latency)
		if !res.Passed {
			failures = append(failures, res.Reason)
		}
	}

	return RunResult{
		ID:       tc.ID,
		Passed:   len(failures) == 0,
		Output:   resp.Text,
		Duration: resp.Metrics.Latency,
		Failures: failures,
	}
}
