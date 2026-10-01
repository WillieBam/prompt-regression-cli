# Prompt Regression & Evaluation CLI

prompt-regression-cli is an evaluation and regression testing CLI for LLMs and prompt templates built in Go.

It helps developers and AI engineers prevent regressions in model behavior, accuracy, latency, and schema compliance when changing prompts, tweaking model hyperparameters, or switching model versions.

---

## Prerequisites


1. **Go**: Version `1.22+` (tested on `1.26+`). Check your version:
   ```bash
   go version
   ```
2. **Make**: (Optional, recommended) For running automated Makefile targets.
3. **LLM API Key**:
   * A Google Gemini API key or OpenAI-compatible provider key.

---

## Installation & Setup

1. **Clone the repository:**
   ```bash
   git clone <repo-url>
   cd prompt-regression-cli
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Configure Environment Variables:**
    Create a `.env` file in the repository root:
   ```bash
    touch .env
   ```
   Edit `.env` with your API credentials:
   ```env
  LLM_PROVIDER=gemini
   LLM_API_KEY=your_actual_api_key_here
   LLM_BASE_URL=url_to_access_llm
   ```

  Supported providers are `gemini` and `openai`. For OpenAI, use:
  ```env
  LLM_PROVIDER=openai
  LLM_API_KEY=sk-your-api-key
  LLM_BASE_URL=https://api.openai.com/v1
  ```

---

## Repository Structure

```text
prompt-regression-cli/
├── cmd/                          # CLI commands built with Cobra
│   ├── root.go                   # Root command definition & global flags
│   ├── bench.go                  # "bench" command: prompt latency & throughput
│   ├── lint.go                   # "lint" command: template syntax validation│   └── eval/
│       ├── eval.go               # "eval" parent command
│       └── run.go                # "eval run" command: executes test suites
├── internal/
│   ├── config/                   # Configuration management (Viper & .env)
│   │   └── config.go
│   ├── endpoint/                 # HTTP endpoint routing & request builders
│   │   └── endpoint.go
│   ├── evaluator/                # Assertion evaluation engine
│   │   └── assert.go             # json-valid, contains, regex, json-field-equals, max-latency
│   ├── provider/                 # LLM provider abstractions
│   │   ├── client.go             # ModelProvider interface & Completion types
│   │   └── gemini.go             # Gemini HTTP client with connection pooling & mock fallback
│   ├── report/                   # Output formatters & reporting
│   │   └── table.go              # ASCII table generator & Pass/Fail summary verdict
│   └── runner/                   # Concurrent test suite executor
│       └── runner.go             # Worker pool, job dispatcher & template renderer
│   └── tui/                      # Bubble Tea interactive menu and evaluation UI
|
├── types/                        # Wire format Data Transfer Objects (DTOs)
│   └── chat.go                   # OpenAI/Gemini Chat Completion request & response
├── testdata/                     # Test suites and mock inputs
│   └── eval.yaml                 # Sample test suite definition
├── Makefile                      # Quick shortcut commands
├── go.mod                        # Go module dependencies
└── main.go                       # Application entrypoint
```

---

## Interactive Menu

Run the CLI without a subcommand to open the Bubble Tea main menu:

```bash
go run .
```

Use the arrow keys or `j`/`k` to move through the menu, press `Enter` to select a command, and press `q` or `Ctrl+C` to quit. After an evaluation, benchmark, or lint command finishes, the menu opens again.

The menu provides shortcuts for:

* Running an evaluation suite
* Benchmarking a sample prompt
* Validating the sample YAML suite
* Quitting the application

The direct commands below remain available for automation and CI/CD workflows.

## Commands

### 1. Running Evaluation Suites (`eval run`)

Execute evaluation suites defined in YAML against the target LLM.

#### Use a custom YAML suite

Create a YAML file with a suite definition, prompt template, inputs, and assertions. For example, `testdata/testing.yaml` contains a custom intent classification test:

```yaml
version: "1"
name: "Custom Intent Test"
model: "gemini-3.1-flash-lite"
template: |
  Classify this customer request: {{ .query }}
  Respond strictly in JSON with category and urgency fields.
  Use category GENERAL and urgency LOW for ordinary informational requests.

tests:
  - id: "general_information_request"
    inputs:
      query: "What are your business hours?"
    assert:
      - type: json-valid
      - type: contains
        value: "GENERAL"
      - type: json-field-equals
        path: "urgency"
        value: "LOW"
```

Run your custom suite by passing its file path with `-s` or `--suite`:

```bash
go run main.go eval run --suite testdata/testing.yaml
```

Use mock mode to validate the suite without an API key:

```bash
go run main.go eval run --suite testdata/testing.yaml --mock
```

You can also run `go run .`, choose **Run evaluation**, and enter the custom YAML path when prompted. Press Enter to use the default `testdata/eval.yaml`.

* **Run live evaluation against Gemini API:**
  ```bash
  go run main.go eval run -s testdata/eval.yaml
  ```
  *(Or use Make)*:
  ```bash
  make eval
  ```

* **Run in offline mock mode (no API key required):**
  ```bash
  go run main.go eval run -s testdata/eval.yaml --mock
  ```
  *(Or use Make)*:
  ```bash
  make mock
  ```

* **Run the interactive menu:**
  ```bash
  go run .
  ```
  When you choose **Run evaluation**, promptctl asks for the YAML suite path. Press Enter to use `testdata/eval.yaml`.

* **Output results as JSON (for CI/CD pipelines):**
  ```bash
  go run main.go eval run -s testdata/eval.yaml --json
  ```
  *(Or use Make)*:
  ```bash
  make pipeline
  ```

* **Configure concurrency and timeout:**
  ```bash
  # Run with 8 concurrent workers and a 45-second suite timeout
  go run main.go eval run -s testdata/eval.yaml -c 8 -t 45s
  ```

* **Select a provider from the command line:**
  ```bash
  go run main.go --provider openai eval run -s testdata/testing.yaml
  ```
  The provider can also be selected with the `LLM_PROVIDER` environment variable. The default is `gemini`.

---

### 2. Benchmarking Prompts (`bench`)

Measure latency and token throughput for a single prompt against the configured model:

```bash
go run main.go bench "Explain goroutines in one sentence"
```
*(Or use Make)*:
```bash
make bench
```

**Example Output:**
```text
Benchmarking prompt against LLM...
Total Latency: 1.25s
Estimated Input Tokens: 7
Estimated Output Token: 28
```

---

### 3. Linting Prompt Templates (`lint`)

Validate prompt template syntax before execution:

```bash
go run main.go lint testdata/eval.yaml
```
*(Or use Make)*:
```bash
make lint
```

---

## Makefile Command Summary

| Command | Description |
| :--- | :--- |
| `make eval` | Runs the test suite against the live LLM API and renders an ASCII table summary |
| `make mock` | Runs the test suite offline using fast simulated responses |
| `make pipeline` | Runs evaluation and outputs machine-readable JSON for CI/CD gates |
| `make bench` | Runs latency and token benchmark on a sample prompt |
| `make lint` | Validates prompt template syntax |

---

## Test Suite Configuration (`eval.yaml`)

Evaluation suites are defined as YAML files with prompt templates and assertions:

```yaml
version: "1"
name: "Customer Intent Classifier"
model: "gemini-3.1-flash-lite"
template: |
  You are an automated triage agent. Classify this request: {{ .query }}
  Allowed categories: [BILLING, TECHNICAL, GENERAL]
  Allowed urgencies: [LOW, HIGH]
  Respond strictly in JSON: {"category": "CATEGORY", "urgency": "URGENCY"}

tests:
  - id: "billing_duplicate_charge"
    inputs:
      query: "I was double charged on my card."
    assert:
      - type: json-valid
      - type: contains
        value: "BILLING"
      - type: json-field-equals
        path: "urgency"
        value: "HIGH"

  - id: "server_downtime_alert"
    inputs:
      query: "The service is returning 500 downtime errors."
    assert:
      - type: json-valid
      - type: contains
        value: "TECHNICAL"
      - type: json-field-equals
        path: "urgency"
        value: "HIGH"
      - type: max-latency
        value: "3s"
```

### Supported Assertion Types

| Assertion Type | Description | Example |
| :--- | :--- | :--- |
| `json-valid` | Verifies response is valid JSON (strips markdown code blocks automatically) | `- type: json-valid` |
| `contains` | Checks if output contains a specific substring | `- type: contains`<br>`  value: "BILLING"` |
| `regex` | Matches output against a regular expression pattern | `- type: regex`<br>`  value: "^\\{.*\\}$"` |
| `json-field-equals` | Checks field equality using dot-notation (e.g. `data.category` or `urgency`) | `- type: json-field-equals`<br>`  path: "urgency"`<br>`  value: "HIGH"` |
| `max-latency` | Verifies request completed within a specified duration | `- type: max-latency`<br>`  value: "2s"` |
