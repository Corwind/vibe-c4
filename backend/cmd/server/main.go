package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/api"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/Corwind/vibe-c4/backend/internal/config"
	"github.com/Corwind/vibe-c4/backend/internal/llm/claude"
	"github.com/Corwind/vibe-c4/backend/internal/project"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg, err := config.Load()
	if err != nil {
		log.Printf("WARN: Failed to load config: %v (using defaults)", err)
		cfg = &config.Config{}
	}

	a := analyzer.NewGoAnalyzer()
	staticBuilder := c4model.NewSmartModelBuilder()

	// Set up AI builder if API key is available
	var aiBuilder c4model.ModelBuilder
	if cfg.Claude.APIKey != "" {
		claudeClient := claude.New(claude.Config{
			APIKey:      cfg.Claude.APIKey,
			Model:       cfg.Claude.Model,
			MaxTokens:   cfg.Claude.MaxTokens,
			TimeoutSecs: cfg.Claude.TimeoutSecs,
		})
		aiBuilder = c4model.NewClaudeModelBuilder(claudeClient, cfg.Claude, staticBuilder)
		log.Printf("AI mode enabled (model: %s)", cfg.Claude.Model)
	} else {
		log.Printf("AI mode disabled (no API key configured)")
	}

	ps := project.NewService(a, staticBuilder, aiBuilder)

	router := api.NewDefaultRouter(ps)

	log.Printf("Starting vibe-c4 API server on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
