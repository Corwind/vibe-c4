package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ClaudeConfig struct {
	APIKey      string `yaml:"api_key"`
	Model       string `yaml:"model"`
	MaxTokens   int    `yaml:"max_tokens"`
	TokenBudget int    `yaml:"token_budget"`
	TimeoutSecs int    `yaml:"timeout_secs"`
}

type Config struct {
	Claude ClaudeConfig `yaml:"claude"`
}

func defaults() *Config {
	return &Config{
		Claude: ClaudeConfig{
			Model:       "claude-sonnet-4-20250514",
			MaxTokens:   8192,
			TokenBudget: 80000,
			TimeoutSecs: 120,
		},
	}
}

// Load reads configuration from YAML files and environment variables.
// It checks .vibe-c4.yaml in the current working directory first,
// then ~/.vibe-c4/config.yaml. Environment variable ANTHROPIC_API_KEY
// is used as a fallback for the API key.
func Load() (*Config, error) {
	cfg := defaults()

	// Try CWD config first
	cwd, err := os.Getwd()
	if err == nil {
		cwdPath := filepath.Join(cwd, ".vibe-c4.yaml")
		if loaded, loadErr := loadYAMLFile(cwdPath, cfg); loaded {
			applyEnvFallback(cfg)
			return cfg, loadErr
		}
	}

	// Try home directory config
	home, err := os.UserHomeDir()
	if err == nil {
		homePath := filepath.Join(home, ".vibe-c4", "config.yaml")
		if loaded, loadErr := loadYAMLFile(homePath, cfg); loaded {
			applyEnvFallback(cfg)
			return cfg, loadErr
		}
	}

	applyEnvFallback(cfg)
	return cfg, nil
}

// LoadFromFile loads configuration from a specific YAML file path.
// Defaults are applied first, then overridden by values from the file.
// Environment variable ANTHROPIC_API_KEY is used as a fallback for the API key.
func LoadFromFile(path string) (*Config, error) {
	cfg := defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	applyEnvFallback(cfg)
	return cfg, nil
}

// loadYAMLFile attempts to read and unmarshal a YAML file into cfg.
// Returns (true, nil) if file was found and parsed successfully,
// (true, err) if file was found but parsing failed,
// (false, nil) if the file does not exist.
func loadYAMLFile(path string, cfg *Config) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return true, fmt.Errorf("reading config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return true, err
	}

	return true, nil
}

func applyEnvFallback(cfg *Config) {
	if cfg.Claude.APIKey == "" {
		cfg.Claude.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	}
}
