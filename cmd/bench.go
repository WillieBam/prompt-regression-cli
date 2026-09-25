package cmd

import (
	"context"
	"prompt-regression-cli/internal/provider"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var benchCmd = &cobra.Command{
	Use:   "bench [prompt]",
	Short: "Benchmark latency and token throughput for a prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode, _ := cmd.Flags().GetBool("mock")
		apiKey := viper.GetString("llm_api_key")
		baseurl := viper.GetString("llm_base_url")
		if mockMode {
			apiKey = ""
		}
		p := provider.NewGeminiProvider(apiKey, baseurl)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd.Printf("Benchmarking prompt against %s...\n", p.Name())
		resp, err := p.Complete(ctx, provider.CompletionRequest{Prompt: args[0]})
		if err != nil {
			return err
		}
		cmd.Printf("Total Latency: %v\n", resp.Metrics.Latency)
		cmd.Printf("Estimated Input Tokens: %d\n", resp.Metrics.InputTokens)
		cmd.Printf("Estimated Output Token: %v\n", resp.Metrics.OutputTokens)
		return nil
	},
}

func init() {
	benchCmd.Flags().BoolP("mock", "m", false, "Force mock mode for offline testing")
}
