package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"prompt-regression-cli/internal/runner"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func PrintResults(w io.Writer, results []runner.RunResult, asJSON bool) {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(results)
		return
	}

	table := tablewriter.NewWriter(w)
	table.Header([]string{"Status", "Test ID", "Duration", "Details"})
	table.Options(tablewriter.WithRenderer(renderer.NewBlueprint((tw.Rendition{
		Borders: tw.Border{
			Top:    tw.On,
			Bottom: tw.On,
			Left:   tw.On,
			Right:  tw.On,
		},
		Symbols: tw.NewSymbols(tw.StyleASCII),
	}))))

	passCount := 0
	failCount := 0

	for _, r := range results {
		status := "PASS"
		details := "-"
		if !r.Passed {
			status = "FAIL"
			failCount++
			if r.Err != nil {
				details = r.Err.Error()
			} else {
				details = strings.Join(r.Failures, "; ")
			}
		} else {
			passCount++
		}
		table.Append([]string{status, r.ID, fmt.Sprintf("%v", r.Duration), details})
	}
	table.Render()

	// Print summary verdict
	if failCount == 0 {
		fmt.Fprintf(w, "\n\033[32m✔ SUITE PASSED\033[0m (%d/%d tests passed)\n", passCount, len(results))
	} else {
		fmt.Fprintf(w, "\n\033[31m✘ SUITE FAILED\033[0m (%d failed, %d passed, %d total)\n", failCount, passCount, len(results))
	}
}
