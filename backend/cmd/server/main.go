package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/api"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/Corwind/vibe-c4/backend/internal/llm"
	"github.com/Corwind/vibe-c4/backend/internal/project"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	a := analyzer.NewGoAnalyzer()
	b := c4model.NewModelBuilder()

	var opts []project.ServiceOption

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config := llm.ClaudeConfig{
			APIKey: apiKey,
		}
		if model := os.Getenv("CLAUDE_MODEL"); model != "" {
			config.Model = model
		}
		adapter := llm.NewClaudeAdapter(config)
		opts = append(opts, project.WithInterpreter(adapter))
		modelName := config.Model
		if modelName == "" {
			modelName = "claude-sonnet-4-20250514"
		}
		log.Printf("AI interpretation enabled (model: %s)", modelName)
	}

	ps := project.NewService(a, b, opts...)

	router := api.NewDefaultRouter(ps)

	log.Printf("Starting vibe-c4 API server on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
