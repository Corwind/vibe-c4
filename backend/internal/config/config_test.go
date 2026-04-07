package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	// Ensure no env var interferes
	t.Setenv("ANTHROPIC_API_KEY", "")

	cfg := defaults()

	assert.Equal(t, "", cfg.Claude.APIKey)
	assert.Equal(t, "claude-sonnet-4-20250514", cfg.Claude.Model)
	assert.Equal(t, 8192, cfg.Claude.MaxTokens)
	assert.Equal(t, 80000, cfg.Claude.TokenBudget)
	assert.Equal(t, 120, cfg.Claude.TimeoutSecs)
}

func TestLoadFromYAML(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")

	yamlContent := `
claude:
  api_key: "sk-yaml-key"
  model: "claude-opus-4-20250514"
  max_tokens: 4096
  token_budget: 50000
  timeout_secs: 60
`
	tmpFile := writeTemp(t, yamlContent)

	cfg, err := LoadFromFile(tmpFile)
	require.NoError(t, err)

	assert.Equal(t, "sk-yaml-key", cfg.Claude.APIKey)
	assert.Equal(t, "claude-opus-4-20250514", cfg.Claude.Model)
	assert.Equal(t, 4096, cfg.Claude.MaxTokens)
	assert.Equal(t, 50000, cfg.Claude.TokenBudget)
	assert.Equal(t, 60, cfg.Claude.TimeoutSecs)
}

func TestLoadEnvVarFallback(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-env-key")

	yamlContent := `
claude:
  model: "claude-sonnet-4-20250514"
`
	tmpFile := writeTemp(t, yamlContent)

	cfg, err := LoadFromFile(tmpFile)
	require.NoError(t, err)

	assert.Equal(t, "sk-env-key", cfg.Claude.APIKey)
}

func TestLoadYAMLOverridesDefaults(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")

	yamlContent := `
claude:
  model: "claude-opus-4-20250514"
`
	tmpFile := writeTemp(t, yamlContent)

	cfg, err := LoadFromFile(tmpFile)
	require.NoError(t, err)

	assert.Equal(t, "claude-opus-4-20250514", cfg.Claude.Model)
	// Other defaults should be preserved
	assert.Equal(t, 8192, cfg.Claude.MaxTokens)
	assert.Equal(t, 80000, cfg.Claude.TokenBudget)
	assert.Equal(t, 120, cfg.Claude.TimeoutSecs)
	assert.Equal(t, "", cfg.Claude.APIKey)
}

func TestLoadAPIKeyFromYAMLTakesPrecedence(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-env-key")

	yamlContent := `
claude:
  api_key: "sk-yaml-key"
`
	tmpFile := writeTemp(t, yamlContent)

	cfg, err := LoadFromFile(tmpFile)
	require.NoError(t, err)

	assert.Equal(t, "sk-yaml-key", cfg.Claude.APIKey)
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
	return path
}
