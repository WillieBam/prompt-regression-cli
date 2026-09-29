package eval

import "github.com/spf13/cobra"

var EvalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Run automated evaluation test suites against models",
}

func init() {
	EvalCmd.AddCommand(runCmd)
}

var RunCmd = runCmd
