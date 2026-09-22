package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint [template-file]",
	Short: "Validate prompt template sytax and variables",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}

		_, err = template.New("lint").Parse(string(content))
		if err != nil {
			return fmt.Errorf("template parse error: %w", err)
		}
		cmd.Println("Template syntax valid.")
		return nil
	},
}
