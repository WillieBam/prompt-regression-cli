.PHONY: eval mock pipeline bench lint 

eval:
	go run main.go eval run -s testdata/eval.yaml

# run the test suite using mock mode
mock:
	go run main.go eval run -s testdata/eval.yaml --mock

# output results as JSON for pipeline integration
pipeline:
	go run main.go eval run -s testdata/eval.yaml --json

# test the benchmark command
bench:
	go run main.go bench "Explain goroutines in one sentence"

# check prompt linting
lint:
	go run main.go lint testdata/eval.yaml
