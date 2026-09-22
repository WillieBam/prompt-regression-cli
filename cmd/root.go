package cmd

import (
	"os"

	"prompt-regression-cli/cmd/eval"

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
	rootCmd.PersistentFlags().String("api-key", "", "API Key for target provider (env: PROMPTCTL_API_KEY)")
	rootCmd.PersistentFlags().Bool("json", false, "Output results in JSON format")

	_ = viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))

	// register subcommands
	rootCmd.AddCommand(eval.EvalCmd)
	rootCmd.AddCommand(benchCmd)
	rootCmd.AddCommand(lintCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
			viper.SetConfigType("yaml")
			viper.SetConfigName(".promptctl")
		}
	}

	viper.SetEnvPrefix("PROMPTCTL")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}
