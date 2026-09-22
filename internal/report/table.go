package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"prompt-regression-cli/internal/runner"

	"github.com/olekukonko/tablewriter"
)

func PrintResults(w io.Writer, results []runner.RunResult, asJSON bool) {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(results)
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Status", "Test ID", "Duration", "Details"})
	table.SetBorder(true)

	for _, r := range results {
		status := "PASS"
		details := "-"
		if !r.Passed {
			status = "FAIL"
			if r.Err != nil {
				details = r.Err.Error()
			} else {
				details = strings.Join(r.Failures, "; ")
			}
		}
		table.Append([]string{status, r.ID, fmt.Sprintf("%v", r.Duration), details})
	}
	table.Render()
}
