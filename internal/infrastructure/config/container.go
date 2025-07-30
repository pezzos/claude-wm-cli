// Package config contains infrastructure configuration and dependency injection.
// This package wires together all layers of the Clean Architecture.
package config

import (
	"path/filepath"

	"claude-wm-cli/internal/domain/services"
	"claude-wm-cli/internal/infrastructure/persistence"
	"claude-wm-cli/internal/interfaces/cli"
)

// Container holds all application dependencies using dependency injection.
// This implements the Dependency Inversion Principle by injecting implementations.
type Container struct {
	// Infrastructure layer
	EpicRepository *persistence.JSONEpicRepository

	// Domain layer  
	EpicDomainService *services.EpicDomainService

	// Interface layer
	EpicCLIAdapter *cli.EpicCLIAdapter
}

// NewContainer creates and wires up all application dependencies.
// This is the composition root where we assemble the Clean Architecture layers.
func NewContainer(dataDir string) (*Container, error) {
	// Infrastructure layer - concrete implementations
	epicFilePath := filepath.Join(dataDir, "epics.json")
	epicRepo, err := persistence.NewJSONEpicRepository(epicFilePath)
	if err != nil {
		return nil, err
	}

	// Domain layer - business logic (depends only on interfaces)
	epicDomainService := services.NewEpicDomainService(epicRepo)

	// Interface layer - adapters (depends on domain interfaces)
	epicCLIAdapter := cli.NewEpicCLIAdapter(epicRepo, epicDomainService)

	return &Container{
		EpicRepository:    epicRepo,
		EpicDomainService: epicDomainService,
		EpicCLIAdapter:    epicCLIAdapter,
	}, nil
}

// GetEpicCLIAdapter returns the epic CLI adapter.
// This is the main entry point for CLI commands.
func (c *Container) GetEpicCLIAdapter() *cli.EpicCLIAdapter {
	return c.EpicCLIAdapter
}

// Example usage function showing how to use the Clean Architecture
/* 
func ExampleUsage() {
	// Initialize the container (composition root)
	container, err := NewContainer("./data")
	if err != nil {
		log.Fatal(err)
	}

	// Get the CLI adapter (interface layer)
	epicAdapter := container.GetEpicCLIAdapter()

	// Use the adapter for CLI operations
	ctx := context.Background()
	
	// Create an epic
	req := cli.CreateEpicRequest{
		ID:          "epic-1",
		Title:       "Implement Clean Architecture",
		Description: "Refactor codebase to follow Clean Architecture principles",
		Priority:    "P1",
		Tags:        []string{"architecture", "refactoring"},
		Duration:    "2 weeks",
	}
	
	epic, err := epicAdapter.CreateEpic(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Created epic: %s\n", epic.Title)
	
	// List epics
	epics, err := epicAdapter.ListEpics(ctx, cli.EpicListOptions{
		Status: "planned",
		Limit:  10,
	})
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Found %d planned epics\n", len(epics))
}
*/