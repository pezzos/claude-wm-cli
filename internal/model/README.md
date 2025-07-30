# Model Package

The `model` package provides centralized type definitions, interfaces, and error handling for the Claude WM CLI application. This package serves as the foundation for consistent data modeling and behavior across all application components.

## Overview

This package eliminates code duplication by centralizing common types that were previously scattered across multiple packages. It follows domain-driven design principles and provides clear interfaces for extending functionality.

## Package Structure

```
internal/model/
├── entity.go     # Common entity types and workflow management
├── errors.go     # Standardized error handling and codes  
├── interfaces.go # Core interfaces and documentation
└── README.md     # This documentation
```

## Key Components

### 1. Common Entity Types

#### BaseEntity
Provides standard fields for all entities:
```go
type BaseEntity struct {
    ID          string    `json:"id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Description string    `json:"description,omitempty"`
}
```

#### Priority System
Standardized P0-P3 priority scale:
```go
const (
    PriorityP0 Priority = "P0" // Critical
    PriorityP1 Priority = "P1" // High  
    PriorityP2 Priority = "P2" // Medium
    PriorityP3 Priority = "P3" // Low
)
```

#### Status Workflow
Standardized status workflow with validation:
```go
const (
    StatusPlanned    Status = "planned"
    StatusInProgress Status = "in_progress"
    StatusBlocked    Status = "blocked"
    StatusOnHold     Status = "on_hold"
    StatusCompleted  Status = "completed"
    StatusCancelled  Status = "cancelled"
)
```

### 2. Error Management

#### Standardized Error Codes
HTTP-style error codes for consistent error handling:
```go
const (
    ErrCodeBadRequest      ErrorCode = 4000
    ErrCodeNotFound        ErrorCode = 4004
    ErrCodeValidation      ErrorCode = 4022
    ErrCodeInternal        ErrorCode = 5000
    ErrCodeWorkflowViolation ErrorCode = 6001
)
```

#### Rich Error Context
Errors include context, suggestions, and structured details:
```go
err := NewNotFoundError("epic").
    WithContext("epic-123").
    WithSuggestion("Check that the epic exists and you have access to it")
```

### 3. Core Interfaces

#### WorkflowState Interface
Entities that participate in workflow management:
```go
type WorkflowState interface {
    GetID() string
    GetStatus() Status
    SetStatus(Status) error
    GetPriority() Priority
    SetPriority(Priority) error
    GetMetadata() BaseEntity
    Validate() error
}
```

#### Repository Pattern
Generic repository interface for data access:
```go
type Repository[T any] interface {
    Create(ctx context.Context, entity T) error
    Read(ctx context.Context, id string) (T, error)
    Update(ctx context.Context, id string, entity T) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter Filter) ([]T, error)
}
```

## Migration Guide

### From Duplicate Types

**Before:**
```go
// internal/epic/types.go
type Priority string
const (
    PriorityLow Priority = "low"
    PriorityHigh Priority = "high"
)

// internal/story/types.go  
type Priority string
const (
    PriorityLow Priority = "low"
    PriorityHigh Priority = "high"
)
```

**After:**
```go
// Both packages now import
import "internal/model"

// Use model.Priority directly
priority := model.PriorityP1
```

### From Custom Errors

**Before:**
```go
return fmt.Errorf("epic not found: %s", id)
```

**After:**
```go
return model.NewNotFoundError("epic").WithContext(id)
```

## Usage Examples

### Creating Entities

```go
// Create a new entity with metadata
entity := model.BaseEntity{
    ID:          generateID(),
    CreatedAt:   time.Now(),
    UpdatedAt:   time.Now(),
    Description: "My epic description",
}

// Set priority and validate
priority := model.PriorityP1
if !priority.IsValid() {
    return model.NewValidationError("invalid priority")
}
```

### Status Transitions

```go
// Check if transition is valid
currentStatus := model.StatusPlanned
newStatus := model.StatusInProgress

if !currentStatus.CanTransitionTo(newStatus) {
    return model.NewWorkflowViolationError(currentStatus, newStatus)
}

// Perform transition
entity.SetStatus(newStatus)
```

### Error Handling

```go
func processEpic(id string) error {
    epic, err := repo.Read(ctx, id)
    if err != nil {
        return model.NewNotFoundError("epic").
            WithContext(id).
            WithSuggestion("Use 'epic list' to see available epics")
    }
    
    if !epic.GetStatus().IsActive() {
        return model.NewValidationError("epic is not active").
            WithDetails(map[string]interface{}{
                "current_status": epic.GetStatus(),
                "epic_id": id,
            })
    }
    
    return nil
}
```

### Progress Tracking

```go
// Calculate progress metrics
progress := model.ProgressMetrics{
    TotalItems:    10,
    CompletedItems: 7,
    TotalPoints:   50,
    CompletedPoints: 35,
}

progress.Calculate()
fmt.Printf("Progress: %.1f%%", progress.CompletionPercent)
```

## Interface Implementation

### Implementing WorkflowState

```go
type Epic struct {
    model.BaseEntity
    Title    string        `json:"title"`
    Priority model.Priority `json:"priority"`
    Status   model.Status   `json:"status"`
}

func (e *Epic) GetID() string { return e.ID }
func (e *Epic) GetStatus() model.Status { return e.Status }
func (e *Epic) SetStatus(s model.Status) error {
    if !e.Status.CanTransitionTo(s) {
        return model.NewWorkflowViolationError(e.Status, s)
    }
    e.Status = s
    e.UpdatedAt = time.Now()
    return nil
}

func (e *Epic) GetPriority() model.Priority { return e.Priority }
func (e *Epic) SetPriority(p model.Priority) error {
    if !p.IsValid() {
        return model.NewValidationError("invalid priority")
    }
    e.Priority = p
    e.UpdatedAt = time.Now()
    return nil
}

func (e *Epic) GetMetadata() model.BaseEntity { return e.BaseEntity }
func (e *Epic) Validate() error {
    var errors model.ValidationErrors
    
    if e.ID == "" {
        errors.Add("id", e.ID, "ID is required")
    }
    if e.Title == "" {
        errors.Add("title", e.Title, "title is required")
    }
    if !e.Priority.IsValid() {
        errors.Add("priority", string(e.Priority), "invalid priority")
    }
    if !e.Status.IsValid() {
        errors.Add("status", string(e.Status), "invalid status")
    }
    
    if errors.HasErrors() {
        return errors
    }
    return nil
}
```

## Testing Support

### Mock Implementations

The package provides interfaces that are easily mockable for testing:

```go
type MockRepository[T any] struct {
    data map[string]T
}

func (m *MockRepository[T]) Create(ctx context.Context, entity T) error {
    // Mock implementation
    return nil
}

func (m *MockRepository[T]) Read(ctx context.Context, id string) (T, error) {
    entity, exists := m.data[id]
    if !exists {
        var zero T
        return zero, model.NewNotFoundError("entity")
    }
    return entity, nil
}
```

### Test Helpers

```go
// Create test entities
func NewTestEpic() *Epic {
    return &Epic{
        BaseEntity: model.BaseEntity{
            ID:        "test-epic-1",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
        Title:    "Test Epic",
        Priority: model.PriorityP2,
        Status:   model.StatusPlanned,
    }
}
```

## Best Practices

### 1. Entity Design
- Always embed `BaseEntity` or `Metadata` for consistency
- Implement `WorkflowState` interface for workflow management
- Use validation methods to ensure data integrity

### 2. Error Handling
- Use appropriate error codes for different error types
- Include context and suggestions in errors
- Wrap underlying errors with `WithCause()`

### 3. Status Management
- Always validate status transitions before applying
- Use the provided helper methods (`IsActive()`, `IsTerminal()`)
- Implement proper state machine validation

### 4. Interface Implementation
- Follow interface contracts strictly
- Handle edge cases gracefully
- Provide comprehensive error messages

## Future Enhancements

### Planned Features
- [ ] Event sourcing support for audit trails
- [ ] Validation rule engine for complex business rules
- [ ] Schema versioning for entity evolution
- [ ] Performance metrics and monitoring hooks

### Breaking Changes
This package was created to centralize previously duplicated types. Migration should be straightforward, but will require updating imports across the codebase.

## Contributing

When adding new types or interfaces to this package:

1. **Documentation**: All public types must have comprehensive documentation
2. **Validation**: Include validation methods for new types
3. **Testing**: Provide test helpers and examples
4. **Compatibility**: Consider backward compatibility for existing code
5. **Standards**: Follow established naming and design patterns

## Dependencies

This package has minimal dependencies to ensure broad compatibility:
- Standard library only
- No external dependencies
- Compatible with Go 1.21+

---

For questions or contributions, please refer to the main project documentation.