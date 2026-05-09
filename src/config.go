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
	LLM       LLMConfig `yaml:"llm"`
	Dangerous []string  `yaml:"dangerous"`
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

// ---- env-first config loading ----

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

// LoadConfig builds config from env vars, optionally overlaying config.yaml.
// config.yaml is NOT required — everything can be set via environment variables.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		LLM: LLMConfig{
			Provider: first(os.Getenv("AUTO_GUARD_PROVIDER"), "deepseek"),
			Model:    first(os.Getenv("AUTO_GUARD_MODEL"), "deepseek-chat"),
			BaseURL:  os.Getenv("AUTO_GUARD_BASE_URL"),
		},
	}

	// Try config.yaml — overlay on top of env defaults if it exists
	path := filepath.Join(configDir(), "config.yaml")
	if data, err := os.ReadFile(path); err == nil {
		var fileCfg Config
		if err := yaml.Unmarshal(data, &fileCfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
		// File values override defaults, but env vars still take priority below
		if fileCfg.LLM.Provider != "" {
			cfg.LLM.Provider = fileCfg.LLM.Provider
		}
		if fileCfg.LLM.Model != "" {
			cfg.LLM.Model = fileCfg.LLM.Model
		}
		if fileCfg.LLM.APIKey != "" {
			cfg.LLM.APIKey = fileCfg.LLM.APIKey
		}
		if fileCfg.LLM.BaseURL != "" {
			cfg.LLM.BaseURL = fileCfg.LLM.BaseURL
		}
		if fileCfg.LLM.Timeout > 0 {
			cfg.LLM.Timeout = fileCfg.LLM.Timeout
		}
		cfg.Dangerous = fileCfg.Dangerous
	}

	// API key: env takes priority over file
	if cfg.LLM.APIKey == "" {
		cfg.LLM.APIKey = os.Getenv("AUTO_GUARD_API_KEY")
	}
	if cfg.LLM.APIKey == "" {
		// Fall back to provider-specific env vars
		switch cfg.LLM.Provider {
		case "deepseek":
			cfg.LLM.APIKey = os.Getenv("DEEPSEEK_API_KEY")
		case "openai":
			cfg.LLM.APIKey = os.Getenv("OPENAI_API_KEY")
		case "anthropic":
			cfg.LLM.APIKey = os.Getenv("ANTHROPIC_API_KEY")
		case "openrouter":
			cfg.LLM.APIKey = os.Getenv("OPENROUTER_API_KEY")
		}
	}
	if cfg.LLM.APIKey == "" {
		// Generic fallback
		cfg.LLM.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf(
			"API key not found — set one of:\n" +
				"  $env:AUTO_GUARD_API_KEY (or DEEPSEEK_API_KEY)\n" +
				"  or create %s with llm.api_key field", path)
	}

	return cfg, nil
}

func first(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
