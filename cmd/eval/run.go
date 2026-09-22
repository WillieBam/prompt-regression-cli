package eval

import (
	"context"
	"fmt"
	"os"
	"time"

	"prompt-regression-cli/internal/provider"
	"prompt-regression-cli/internal/report"
	"prompt-regression-cli/internal/runner"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute an evaluation suite YAML file",
	RunE: func(cmd *cobra.Command, args []string) error {
		suitePath, _ := cmd.Flags().GetString("suite")
		concurrency, _ := cmd.Flags().GetInt("concurrency")
		timeout, _ := cmd.Flags().GetDuration("timeout")
		asJSON := viper.GetBool("json")

		suite, err := runner.LoadSuite(suitePath)
		if err != nil {
			return fmt.Errorf("failed to load suite: %w", err)
		}

		if !asJSON {
			fmt.Printf("Evaluating %q (%d tests) using %s...\n", suite.Name, len(suite.Tests), suite.Model)
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		p := provider.NewGeminiProvider(viper.GetString("api_key"))

		resultChan := runner.ExecuteSuite(ctx, suite, p, concurrency)

		var results []runner.RunResult
		failed := 0
		for r := range resultChan {
			if !r.Passed {
				failed++
			}
			results = append(results, r)
		}

		report.PrintResults(cmd.OutOrStdout(), results, asJSON)

		if failed > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	runCmd.Flags().StringP("suite", "s", "testdata/customer_eval.yaml", "Path to test suite definition")
	runCmd.Flags().IntP("concurrency", "c", 4, "Number of concurrent evaluation workers")
	runCmd.Flags().DurationP("timeout", "t", 30*time.Second, "Suite timeout")
}
