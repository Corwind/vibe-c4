package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/api"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/Corwind/vibe-c4/backend/internal/project"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	a := analyzer.NewGoAnalyzer()
	b := c4model.NewSmartModelBuilder()
	ps := project.NewService(a, b)

	router := api.NewDefaultRouter(ps)

	log.Printf("Starting vibe-c4 API server on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
