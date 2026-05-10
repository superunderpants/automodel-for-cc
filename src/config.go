package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LLM LLMConfig `yaml:"llm"`
}

type LLMConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	Timeout int    `yaml:"timeout"`
}

func (c *LLMConfig) ResolveTimeout() time.Duration {
	if c.Timeout <= 0 {
		return 10 * time.Second
	}
	return time.Duration(c.Timeout) * time.Second
}

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
	cfg := &Config{
		LLM: LLMConfig{
			BaseURL: "https://api.deepseek.com/anthropic",
			Model:   "deepseek-chat",
		},
	}

	// Env overrides
	if v := os.Getenv("AUTO_GUARD_BASE_URL"); v != "" {
		cfg.LLM.BaseURL = v
	}
	if v := os.Getenv("AUTO_GUARD_MODEL"); v != "" {
		cfg.LLM.Model = v
	}

	// config.yaml overrides env defaults (but not env vars set above)
	path := filepath.Join(configDir(), "config.yaml")
	if data, err := os.ReadFile(path); err == nil {
		var fileCfg Config
		if err := yaml.Unmarshal(data, &fileCfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
		if fileCfg.LLM.BaseURL != "" {
			cfg.LLM.BaseURL = fileCfg.LLM.BaseURL
		}
		if fileCfg.LLM.Model != "" {
			cfg.LLM.Model = fileCfg.LLM.Model
		}
		if fileCfg.LLM.APIKey != "" {
			cfg.LLM.APIKey = fileCfg.LLM.APIKey
		}
		if fileCfg.LLM.Timeout > 0 {
			cfg.LLM.Timeout = fileCfg.LLM.Timeout
		}
	}

	// API key: AUTO_GUARD_API_KEY → ANTHROPIC_AUTH_TOKEN → OPENAI_API_KEY
	if cfg.LLM.APIKey == "" {
		cfg.LLM.APIKey = os.Getenv("AUTO_GUARD_API_KEY")
	}
	if cfg.LLM.APIKey == "" {
		cfg.LLM.APIKey = os.Getenv("ANTHROPIC_AUTH_TOKEN")
	}
	if cfg.LLM.APIKey == "" {
		cfg.LLM.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf(
			"API key not found — set one of:\n"+
				"  $env:AUTO_GUARD_API_KEY\n"+
				"  or create %s with llm.api_key field", path)
	}

	return cfg, nil
}
