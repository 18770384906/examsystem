// config/ai_config.go
package config

import (
	"os"
)

type AIConfig struct {
	TongyiAPIKey   string
	DeepSeekAPIKey string
	TongyiAPIURL   string
	DeepSeekAPIURL string
}

func LoadAIConfig() AIConfig {
	return AIConfig{
		TongyiAPIKey:   os.Getenv("TONGYI_API_KEY"),
		DeepSeekAPIKey: os.Getenv("DEEPSEEK_API_KEY"),
		TongyiAPIURL:   os.Getenv("TONGYI_API_URL"),
		DeepSeekAPIURL: os.Getenv("DEEPSEEK_API_URL"),
	}
}
