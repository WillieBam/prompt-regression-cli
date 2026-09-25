package cmd

import (
	"prompt-regression-cli/cmd/eval"
	"prompt-regression-cli/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "promptctl",
		Short: "High-concurrency LLM evaluation and regression testing CLI",
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.promptctl.yaml)")
	rootCmd.PersistentFlags().String("llm-api-key", "", "API Key for target provider (env: PROMPTCTL_API_KEY)")
	rootCmd.PersistentFlags().Bool("json", false, "Output results in JSON format")

	_ = viper.BindPFlag("llm-api-key", rootCmd.PersistentFlags().Lookup("llm-api-key"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))

	// register subcommands
	rootCmd.AddCommand(eval.EvalCmd)
	rootCmd.AddCommand(benchCmd)
	rootCmd.AddCommand(lintCmd)
}

func initConfig() {
	_, _ = config.Load(cfgFile)
}
