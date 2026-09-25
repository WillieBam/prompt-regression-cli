package config

import (
	"os"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	APIKey  string `mapstructure:"llm_api_key"`
	BaseURL string `mapstructure:"llm_base_url"`
	JSON    bool   `mapstructure:"json"`
}

var current = &Config{}

func Load(cfgFile string) (*Config, error) {
	viper.SetDefault("llm_api_key", "")
	viper.SetDefault("llm_base_url", "")
	viper.SetDefault("json", false)

	_ = gotenv.Load()

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

	_ = viper.BindEnv("llm_api_key", "LLM_API_KEY")
	_ = viper.BindEnv("llm_base_url", "LLM_BASE_URL")

	_ = viper.ReadInConfig()

	if err := viper.Unmarshal(current); err != nil {
		return nil, err
	}

	return current, nil
}

// Get returns the current configuration instance.
func Get() *Config {
	return current
}
