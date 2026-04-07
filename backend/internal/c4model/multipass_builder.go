package c4model

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/config"
	"github.com/Corwind/vibe-c4/backend/internal/llm"
	"golang.org/x/sync/errgroup"
)

// MultiPassBuilder implements ModelBuilder using a multi-pass Claude API pipeline
// that generates C4 architecture diagrams level by level:
//   - Pass 1: System Context + Containers (L1+L2)
//   - Pass 2: Components per container (L3) — parallel
//   - Pass 3: Code elements per component (L4) — parallel
//   - Merge: Combine results, rebuild parent-child links, validate
type MultiPassBuilder struct {
	llmClient llm.Client
	config    config.ClaudeConfig
	fallback  ModelBuilder
}

// NewMultiPassBuilder creates a new MultiPassBuilder.
func NewMultiPassBuilder(client llm.Client, cfg config.ClaudeConfig, fallback ModelBuilder) *MultiPassBuilder {
	return &MultiPassBuilder{
		llmClient: client,
		config:    cfg,
		fallback:  fallback,
	}
}

// BuildFromAnalysis implements ModelBuilder. It runs the multi-pass pipeline and
// falls back to the deterministic builder on failure.
func (b *MultiPassBuilder) BuildFromAnalysis(result *analyzer.AnalysisResult) (*C4Model, error) {
	if result == nil {
		return nil, fmt.Errorf("analysis result is nil")
	}

	model, err := b.runPipeline(context.Background(), result)
	if err != nil {
		log.Printf("WARN: multi-pass pipeline failed: %v; falling back to deterministic builder", err)
		return b.fallback.BuildFromAnalysis(result)
	}
	return model, nil
}

// runPipeline executes the three-pass pipeline and merges results.
func (b *MultiPassBuilder) runPipeline(ctx context.Context, result *analyzer.AnalysisResult) (*C4Model, error) {
	// Pass 1: System Context + Containers (L1+L2)
	pass1, repoContents, err := b.runPass1(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("pass 1 failed: %w", err)
	}

	// Pass 2: Components per container (L3) — parallel
	pass2s := b.runPass2(ctx, result, pass1, repoContents)

	// Pass 3: Code elements per component (L4) — parallel
	var allComponents []Component
	for _, p2 := range pass2s {
		allComponents = append(allComponents, p2.Components...)
	}
	pass3s := b.runPass3(ctx, result, allComponents, repoContents)

	// Merge
	model := MergeResults(pass1, pass2s, pass3s)

	if err := ValidateC4Model(model); err != nil {
		return nil, fmt.Errorf("merged model validation failed: %w", err)
	}

	return model, nil
}

// runPass1 executes the L1+L2 pass: reads the full repository and calls Claude.
func (b *MultiPassBuilder) runPass1(ctx context.Context, result *analyzer.AnalysisResult) (*Pass1Result, map[string]string, error) {
	repoContents, err := ReadFullRepository(result.ProjectPath, b.config.TokenBudget)
	if err != nil {
		return nil, nil, fmt.Errorf("read repository: %w", err)
	}

	systemPrompt := BuildContextPrompt()
	userPrompt := BuildContextUserPrompt(result, repoContents)

	resp, err := b.llmClient.Complete(ctx, llm.Request{
		SystemPrompt: systemPrompt,
		Messages:     []llm.Message{{Role: "user", Content: userPrompt}},
		MaxTokens:    b.config.MaxTokens,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("llm call: %w", err)
	}

	log.Printf("Pass 1 token usage: input=%d output=%d", resp.Usage.InputTokens, resp.Usage.OutputTokens)

	pass1, err := ParsePass1Response(resp.Content)
	if err != nil {
		return nil, nil, fmt.Errorf("parse pass1: %w", err)
	}

	return pass1, repoContents, nil
}

// runPass2 executes L3 passes in parallel, one per container.
func (b *MultiPassBuilder) runPass2(ctx context.Context, result *analyzer.AnalysisResult, pass1 *Pass1Result, repoContents map[string]string) []Pass2Result {
	if len(pass1.Containers) == 0 {
		return nil
	}

	maxConc := b.config.MaxConcurrency
	if maxConc <= 0 {
		maxConc = 5
	}

	maxTokens := b.config.MaxTokens / 2
	if maxTokens < 2048 {
		maxTokens = 2048
	}

	var mu sync.Mutex
	var results []Pass2Result

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConc)

	for _, container := range pass1.Containers {
		container := container
		g.Go(func() error {
			scopedFiles := ScopeFilesToPackage(repoContents, container.PackagePath, result.Module.ModulePath)
			scopedResult := ScopeAnalysis(result, []string{container.PackagePath})

			systemPrompt := BuildComponentPrompt(container, pass1.Containers)
			userPrompt := BuildComponentUserPrompt(scopedResult, container, scopedFiles)

			resp, err := b.llmClient.Complete(gctx, llm.Request{
				SystemPrompt: systemPrompt,
				Messages:     []llm.Message{{Role: "user", Content: userPrompt}},
				MaxTokens:    maxTokens,
			})
			if err != nil {
				log.Printf("WARN: pass2 failed for container %s: %v", container.ID, err)
				return nil
			}

			log.Printf("Pass 2 (%s) token usage: input=%d output=%d", container.ID, resp.Usage.InputTokens, resp.Usage.OutputTokens)

			pass2, err := ParsePass2Response(resp.Content)
			if err != nil {
				log.Printf("WARN: pass2 parse failed for container %s: %v", container.ID, err)
				return nil
			}

			mu.Lock()
			results = append(results, *pass2)
			mu.Unlock()

			return nil
		})
	}

	_ = g.Wait()
	return results
}

// runPass3 executes L4 passes in parallel, one per component.
func (b *MultiPassBuilder) runPass3(ctx context.Context, result *analyzer.AnalysisResult, components []Component, repoContents map[string]string) []Pass3Result {
	if len(components) == 0 {
		return nil
	}

	maxConc := b.config.MaxConcurrency
	if maxConc <= 0 {
		maxConc = 5
	}

	maxTokens := b.config.MaxTokens / 4
	if maxTokens < 1024 {
		maxTokens = 1024
	}

	var mu sync.Mutex
	var results []Pass3Result

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConc)

	for _, component := range components {
		component := component
		g.Go(func() error {
			scopedFiles := ScopeFilesToPackage(repoContents, component.PackagePath, result.Module.ModulePath)
			scopedResult := ScopeAnalysis(result, []string{component.PackagePath})

			systemPrompt := BuildCodePrompt(component)
			userPrompt := BuildCodeUserPrompt(scopedResult, component, scopedFiles)

			resp, err := b.llmClient.Complete(gctx, llm.Request{
				SystemPrompt: systemPrompt,
				Messages:     []llm.Message{{Role: "user", Content: userPrompt}},
				MaxTokens:    maxTokens,
			})
			if err != nil {
				log.Printf("WARN: pass3 failed for component %s: %v", component.ID, err)
				return nil
			}

			log.Printf("Pass 3 (%s) token usage: input=%d output=%d", component.ID, resp.Usage.InputTokens, resp.Usage.OutputTokens)

			pass3, err := ParsePass3Response(resp.Content)
			if err != nil {
				log.Printf("WARN: pass3 parse failed for component %s: %v", component.ID, err)
				return nil
			}

			mu.Lock()
			results = append(results, *pass3)
			mu.Unlock()

			return nil
		})
	}

	_ = g.Wait()
	return results
}

// MergeResults combines results from all three passes into a single C4Model.
// It rebuilds parent-child ID lists and deduplicates relationships.
func MergeResults(pass1 *Pass1Result, pass2s []Pass2Result, pass3s []Pass3Result) *C4Model {
	model := &C4Model{
		Systems:    pass1.Systems,
		Containers: pass1.Containers,
	}

	var allRels []Relationship
	allRels = append(allRels, pass1.Relationships...)

	for _, p2 := range pass2s {
		model.Components = append(model.Components, p2.Components...)
		allRels = append(allRels, p2.Relationships...)
	}

	for _, p3 := range pass3s {
		model.CodeElements = append(model.CodeElements, p3.CodeElements...)
	}

	model.Relationships = DeduplicateRelationships(allRels)

	rebuildParentChildIDs(model)

	return model
}

// DeduplicateRelationships removes duplicate relationships keyed on sourceID->targetID@level.
func DeduplicateRelationships(rels []Relationship) []Relationship {
	seen := make(map[string]bool)
	var result []Relationship

	for _, r := range rels {
		key := r.SourceID + "->" + r.TargetID + "@" + r.Level
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, r)
	}

	return result
}

// rebuildParentChildIDs repopulates ContainerIDs, ComponentIDs, and CodeElements
// based on the foreign key references in each child element.
func rebuildParentChildIDs(model *C4Model) {
	systemContainers := make(map[string][]string)
	for _, c := range model.Containers {
		if c.SystemID != "" {
			systemContainers[c.SystemID] = append(systemContainers[c.SystemID], c.ID)
		}
	}
	for i := range model.Systems {
		model.Systems[i].ContainerIDs = systemContainers[model.Systems[i].ID]
	}

	containerComponents := make(map[string][]string)
	for _, c := range model.Components {
		if c.ContainerID != "" {
			containerComponents[c.ContainerID] = append(containerComponents[c.ContainerID], c.ID)
		}
	}
	for i := range model.Containers {
		model.Containers[i].ComponentIDs = containerComponents[model.Containers[i].ID]
	}

	componentCodes := make(map[string][]string)
	for _, ce := range model.CodeElements {
		if ce.ComponentID != "" {
			componentCodes[ce.ComponentID] = append(componentCodes[ce.ComponentID], ce.ID)
		}
	}
	for i := range model.Components {
		model.Components[i].CodeElements = componentCodes[model.Components[i].ID]
	}
}
