// Package application provides base application service functionality.
package application

import (
	"context"
	"fmt"
	"time"

	"claude-wm-cli/internal/entity"
	"claude-wm-cli/internal/model"
)

// BaseApplicationService provides common functionality for all application services.
// This eliminates duplication across different entity application services.
type BaseApplicationService[T entity.BaseEntity] struct {
	entityManager   entity.EntityService[T]
	notificationSvc entity.EntityNotificationService
	metricsSvc      entity.EntityMetricsService
	cacheSvc        entity.EntityCacheService[T]
	entityType      string
}

// BaseApplicationServiceConfig provides common configuration for application services.
type BaseApplicationServiceConfig struct {
	EntityType      string
	DataDirectory   string
	EnableCache     bool
	EnableMetrics   bool
	EnableNotifications bool
	CacheTTL        time.Duration
}

// NewBaseApplicationService creates a new base application service.
func NewBaseApplicationService[T entity.BaseEntity](
	entityManager entity.EntityService[T],
	config BaseApplicationServiceConfig,
) *BaseApplicationService[T] {
	return &BaseApplicationService[T]{
		entityManager: entityManager,
		entityType:    config.EntityType,
		// In real implementation, inject notification, metrics, and cache services
	}
}

// CommonWorkflowRequest represents common fields across all workflow requests.
type CommonWorkflowRequest struct {
	SkipValidation  bool                   `json:"skip_validation,omitempty"`
	Force           bool                   `json:"force,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	RequestID       string                 `json:"request_id,omitempty"`
	UserID          string                 `json:"user_id,omitempty"`
	Timestamp       time.Time              `json:"timestamp,omitempty"`
}

// CommonWorkflowResponse represents common fields in all workflow responses.
type CommonWorkflowResponse struct {
	Success     bool                   `json:"success"`
	Warnings    []string               `json:"warnings,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	NextActions []string               `json:"next_actions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
}

// WorkflowContext provides context information for workflow operations.
type WorkflowContext struct {
	UserID        string                 `json:"user_id,omitempty"`
	RequestID     string                 `json:"request_id,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
	Operation     string                 `json:"operation"`
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	TraceID       string                 `json:"trace_id,omitempty"`
}

// ValidationResult represents the result of business validation.
type ValidationResult struct {
	Valid       bool     `json:"valid"`
	Errors      []string `json:"errors,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// BusinessRuleEngine provides common business rule validation.
type BusinessRuleEngine[T entity.BaseEntity] struct {
	rules []BusinessRule[T]
}

// BusinessRule represents a single business rule.
type BusinessRule[T entity.BaseEntity] struct {
	Name        string
	Description string
	Severity    string // "error", "warning", "info"
	Validate    func(ctx context.Context, entity T) ValidationResult
}

// NewBusinessRuleEngine creates a new business rule engine.
func NewBusinessRuleEngine[T entity.BaseEntity]() *BusinessRuleEngine[T] {
	return &BusinessRuleEngine[T]{
		rules: make([]BusinessRule[T], 0),
	}
}

// AddRule adds a business rule to the engine.
func (e *BusinessRuleEngine[T]) AddRule(rule BusinessRule[T]) {
	e.rules = append(e.rules, rule)
}

// ValidateEntity validates an entity against all business rules.
func (e *BusinessRuleEngine[T]) ValidateEntity(ctx context.Context, entity T) ValidationResult {
	result := ValidationResult{Valid: true}
	
	for _, rule := range e.rules {
		ruleResult := rule.Validate(ctx, entity)
		
		if !ruleResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, ruleResult.Errors...)
		}
		
		result.Warnings = append(result.Warnings, ruleResult.Warnings...)
		result.Suggestions = append(result.Suggestions, ruleResult.Suggestions...)
	}
	
	return result
}

// WorkflowOrchestrator coordinates complex workflows across multiple entities.
type WorkflowOrchestrator struct {
	services map[string]interface{} // Map of entity type to application service
}

// NewWorkflowOrchestrator creates a new workflow orchestrator.
func NewWorkflowOrchestrator() *WorkflowOrchestrator {
	return &WorkflowOrchestrator{
		services: make(map[string]interface{}),
	}
}

// RegisterService registers an application service with the orchestrator.
func (o *WorkflowOrchestrator) RegisterService(entityType string, service interface{}) {
	o.services[entityType] = service
}

// GetService retrieves an application service by entity type.
func (o *WorkflowOrchestrator) GetService(entityType string) (interface{}, bool) {
	service, exists := o.services[entityType]
	return service, exists
}

// CrossEntityWorkflowRequest represents a workflow that spans multiple entity types.
type CrossEntityWorkflowRequest struct {
	WorkflowType string                 `json:"workflow_type"`
	Entities     map[string]interface{} `json:"entities"` // entityType -> entity data
	Options      map[string]interface{} `json:"options,omitempty"`
	CommonWorkflowRequest
}

// CrossEntityWorkflowResponse represents the response from a cross-entity workflow.
type CrossEntityWorkflowResponse struct {
	Results     map[string]interface{} `json:"results"` // entityType -> operation result
	Summary     map[string]interface{} `json:"summary"`
	CommonWorkflowResponse
}

// ExecuteCrossEntityWorkflow executes a workflow that involves multiple entity types.
func (o *WorkflowOrchestrator) ExecuteCrossEntityWorkflow(
	ctx context.Context,
	req CrossEntityWorkflowRequest,
) (*CrossEntityWorkflowResponse, error) {
	startTime := time.Now()
	
	results := make(map[string]interface{})
	var allWarnings []string
	var allSuggestions []string
	
	// Execute workflow based on type
	switch req.WorkflowType {
	case "create_epic_with_stories":
		// Example: Create epic and initial stories in one workflow
		// Implementation would coordinate between EpicApplicationService and StoryApplicationService
		
	case "complete_feature":
		// Example: Mark all stories complete, then mark epic complete
		// Implementation would update stories first, then epic
		
	default:
		return nil, model.NewValidationError("unknown workflow type").
			WithContext(req.WorkflowType).
			WithSuggestions([]string{"create_epic_with_stories", "complete_feature"})
	}
	
	duration := time.Since(startTime)
	
	return &CrossEntityWorkflowResponse{
		Results: results,
		Summary: map[string]interface{}{
			"workflow_type":     req.WorkflowType,
			"entities_processed": len(req.Entities),
			"execution_time":    duration,
		},
		CommonWorkflowResponse: CommonWorkflowResponse{
			Success:     true,
			Warnings:    allWarnings,
			Suggestions: allSuggestions,
			Duration:    duration,
			RequestID:   req.RequestID,
		},
	}, nil
}

// ApplicationServiceRegistry provides centralized access to all application services.
type ApplicationServiceRegistry struct {
	epicService  *EpicApplicationService
	// storyService *StoryApplicationService  // To be implemented
	// ticketService *TicketApplicationService // To be implemented
	orchestrator *WorkflowOrchestrator
}

// ApplicationServiceRegistryConfig configures the service registry.
type ApplicationServiceRegistryConfig struct {
	DataDirectory       string
	EnableCache         bool
	EnableMetrics       bool
	EnableNotifications bool
	CacheTTL            time.Duration
}

// NewApplicationServiceRegistry creates a new service registry.
func NewApplicationServiceRegistry(config ApplicationServiceRegistryConfig) (*ApplicationServiceRegistry, error) {
	// Create Epic service
	epicService, err := NewEpicApplicationService(EpicApplicationServiceConfig{
		DataDirectory:       config.DataDirectory,
		EnableCache:         config.EnableCache,
		EnableMetrics:       config.EnableMetrics,
		EnableNotifications: config.EnableNotifications,
	})
	if err != nil {
		return nil, model.NewInternalError("failed to create epic service").WithCause(err)
	}
	
	// Create orchestrator
	orchestrator := NewWorkflowOrchestrator()
	orchestrator.RegisterService("epic", epicService)
	
	return &ApplicationServiceRegistry{
		epicService:  epicService,
		orchestrator: orchestrator,
	}, nil
}

// GetEpicService returns the Epic application service.
func (r *ApplicationServiceRegistry) GetEpicService() *EpicApplicationService {
	return r.epicService
}

// GetOrchestrator returns the workflow orchestrator.
func (r *ApplicationServiceRegistry) GetOrchestrator() *WorkflowOrchestrator {
	return r.orchestrator
}

// ApplicationServiceInterface defines the interface that all application services should implement.
type ApplicationServiceInterface interface {
	GetEntityType() string
	HealthCheck(ctx context.Context) error
	GetMetrics(ctx context.Context) (map[string]interface{}, error)
	GetStatus(ctx context.Context) (map[string]interface{}, error)
}

// HealthCheckResult represents the health status of an application service.
type HealthCheckResult struct {
	Service   string                 `json:"service"`
	Healthy   bool                   `json:"healthy"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// PerformHealthCheck performs health checks on all registered services.
func (r *ApplicationServiceRegistry) PerformHealthCheck(ctx context.Context) ([]HealthCheckResult, error) {
	results := make([]HealthCheckResult, 0)
	
	// Check Epic service
	epicHealth := HealthCheckResult{
		Service:   "epic",
		Timestamp: time.Now(),
	}
	
	if err := r.epicService.entityManager.Count(ctx, nil); err != nil {
		epicHealth.Healthy = false
		epicHealth.Message = err.Error()
	} else {
		epicHealth.Healthy = true
		epicHealth.Message = "Service operational"
	}
	
	results = append(results, epicHealth)
	
	return results, nil
}

// GetSystemMetrics returns overall system metrics across all services.
func (r *ApplicationServiceRegistry) GetSystemMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := map[string]interface{}{
		"timestamp": time.Now(),
		"services":  make(map[string]interface{}),
	}
	
	// Epic service metrics
	epicCount, err := r.epicService.entityManager.Count(ctx, nil)
	if err != nil {
		return nil, model.NewInternalError("failed to get epic metrics").WithCause(err)
	}
	
	metrics["services"].(map[string]interface{})["epic"] = map[string]interface{}{
		"total_count": epicCount,
		"entity_type": "epic",
	}
	
	return metrics, nil
}

// Common utility functions for application services

// GenerateRequestID generates a unique request ID for tracing.
func GenerateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

// CreateWorkflowContext creates a workflow context with common fields.
func CreateWorkflowContext(operation, entityType, entityID string) WorkflowContext {
	return WorkflowContext{
		RequestID:  GenerateRequestID(),
		Operation:  operation,
		EntityType: entityType,
		EntityID:   entityID,
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}
}

// ValidateCommonFields validates common fields across requests.
func ValidateCommonFields(req CommonWorkflowRequest) error {
	// Add common validation logic here
	return nil
}

// EnrichResponse enriches a response with common metadata.
func EnrichResponse(resp *CommonWorkflowResponse, ctx WorkflowContext, startTime time.Time) {
	resp.RequestID = ctx.RequestID
	resp.Duration = time.Since(startTime)
	resp.Success = true
	
	if resp.Metadata == nil {
		resp.Metadata = make(map[string]interface{})
	}
	
	resp.Metadata["operation"] = ctx.Operation
	resp.Metadata["entity_type"] = ctx.EntityType
	resp.Metadata["timestamp"] = ctx.Timestamp
}