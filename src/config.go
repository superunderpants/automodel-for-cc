package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ---- config struct ----

type Config struct {
	LLM       LLMConfig    `yaml:"llm"`
	Dangerous []string     `yaml:"dangerous"`
}

type LLMConfig struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
	BaseURL  string `yaml:"base_url"`
	Timeout  int    `yaml:"timeout"`
}

// ---- provider URL mapping ----

var providerBaseURLs = map[string]string{
	"deepseek":   "https://api.deepseek.com/v1/chat/completions",
	"openai":     "https://api.openai.com/v1/chat/completions",
	"openrouter": "https://openrouter.ai/api/v1/chat/completions",
	"ollama":     "http://localhost:11434/v1/chat/completions",
	"anthropic":  "https://api.anthropic.com/v1/messages",
}

func (c *LLMConfig) ResolveBaseURL() string {
	if c.Provider == "custom" || c.Provider == "" {
		return c.BaseURL
	}
	if url, ok := providerBaseURLs[c.Provider]; ok {
		return url
	}
	return c.BaseURL
}

func (c *LLMConfig) ResolveTimeout() time.Duration {
	if c.Timeout <= 0 {
		return 10 * time.Second
	}
	return time.Duration(c.Timeout) * time.Second
}

// ---- config loading ----

func configDir() string {
	if d := os.Getenv("AUTO_GUARD_CONFIG_DIR"); d != "" {
		return d
	}
	if d := os.Getenv("APPDATA"); d != "" {
		return filepath.Join(d, "auto-guard")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "auto-guard")
}

func LoadConfig() (*Config, error) {
	path := filepath.Join(configDir(), "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.LLM.Provider == "" {
		return nil, fmt.Errorf("llm.provider is required")
	}
	if cfg.LLM.APIKey == "" {
		cfg.LLM.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("llm.api_key is required (or set OPENAI_API_KEY)")
	}
	if cfg.LLM.Model == "" {
		cfg.LLM.Model = "deepseek-chat"
	}
	return &cfg, nil
}
