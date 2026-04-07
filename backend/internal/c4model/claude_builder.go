package c4model

import (
	"context"
	"fmt"
	"log"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/config"
	"github.com/Corwind/vibe-c4/backend/internal/llm"
)

// ClaudeModelBuilder implements ModelBuilder using the Claude API for
// AI-assisted C4 model generation. It falls back to a static builder on failure.
type ClaudeModelBuilder struct {
	llmClient llm.Client
	config    config.ClaudeConfig
	fallback  ModelBuilder
}

// NewClaudeModelBuilder creates a new ClaudeModelBuilder.
func NewClaudeModelBuilder(client llm.Client, cfg config.ClaudeConfig, fallback ModelBuilder) *ClaudeModelBuilder {
	return &ClaudeModelBuilder{
		llmClient: client,
		config:    cfg,
		fallback:  fallback,
	}
}

// BuildFromAnalysis generates a C4 model by sending the analysis result and
// repository source code to Claude. Falls back to the static builder on failure.
func (b *ClaudeModelBuilder) BuildFromAnalysis(result *analyzer.AnalysisResult) (*C4Model, error) {
	if result == nil {
		return nil, fmt.Errorf("analysis result is nil")
	}

	model, err := b.buildWithClaude(context.Background(), result)
	if err != nil {
		log.Printf("WARN: Claude model building failed, falling back to static builder: %v", err)
		return b.fallback.BuildFromAnalysis(result)
	}

	return model, nil
}

func (b *ClaudeModelBuilder) buildWithClaude(ctx context.Context, result *analyzer.AnalysisResult) (*C4Model, error) {
	// Read the full repository source code
	repoContents, err := ReadFullRepository(result.ProjectPath, b.config.TokenBudget)
	if err != nil {
		return nil, fmt.Errorf("reading repository: %w", err)
	}

	// Build prompts
	systemPrompt := BuildSystemPrompt()
	userPrompt := BuildUserPrompt(result, repoContents)

	// Call Claude
	resp, err := b.llmClient.Complete(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		Messages: []llm.Message{
			{Role: "user", Content: userPrompt},
		},
		MaxTokens: b.config.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	// Parse response
	model, err := ParseClaudeResponse(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("parsing Claude response: %w", err)
	}

	// Validate
	if err := ValidateC4Model(model); err != nil {
		return nil, fmt.Errorf("validating C4 model: %w", err)
	}

	log.Printf("INFO: Claude generated C4 model (input: %d tokens, output: %d tokens)",
		resp.Usage.InputTokens, resp.Usage.OutputTokens)

	return model, nil
}
