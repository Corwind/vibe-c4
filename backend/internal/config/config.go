package config

// ClaudeConfig holds configuration for Claude API interactions.
type ClaudeConfig struct {
	APIKey         string `json:"api_key"`
	Model          string `json:"model"`
	MaxTokens      int    `json:"max_tokens"`
	TokenBudget    int    `json:"token_budget"`
	TimeoutSecs    int    `json:"timeout_secs"`
	MaxConcurrency int    `json:"max_concurrency"`
}
