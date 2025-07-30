// Package persistence provides generic repository implementations for data access.
// This package centralizes CRUD operations and reduces code duplication.
package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"claude-wm-cli/internal/model"
	"claude-wm-cli/internal/state"
)

// JSONRepository provides a generic repository implementation using JSON files.
// It implements atomic operations, validation, and error handling.
type JSONRepository[T any] struct {
	mu           sync.RWMutex
	filePath     string
	validator    ValidatorFunc[T]
	atomicWriter *state.AtomicWriter
	cache        map[string]*CacheEntry[T]
	cacheEnabled bool
	cacheTTL     time.Duration
}

// ValidatorFunc defines the signature for entity validation functions.
type ValidatorFunc[T any] func(T) error

// CacheEntry represents a cached entity with TTL.
type CacheEntry[T any] struct {
	Entity    T
	ExpiresAt time.Time
}

// RepositoryOptions configures repository behavior.
type RepositoryOptions struct {
	EnableCache   bool          // Enable in-memory caching
	CacheTTL      time.Duration // Cache time-to-live
	EnableBackup  bool          // Enable automatic backups
	EnableGit     bool          // Enable Git versioning
	FileMode      os.FileMode   // File permissions
}

// DefaultRepositoryOptions returns sensible defaults for repository configuration.
func DefaultRepositoryOptions() RepositoryOptions {
	return RepositoryOptions{
		EnableCache:  true,
		CacheTTL:     5 * time.Minute,
		EnableBackup: true,
		EnableGit:    true,
		FileMode:     0644,
	}
}

// NewJSONRepository creates a new JSON-based repository.
//
// Parameters:
//   - filePath: Path to the JSON file for persistence
//   - validator: Function to validate entities before persistence
//   - options: Repository configuration options
//
// Returns a repository instance that implements the model.Repository interface.
func NewJSONRepository[T any](filePath string, validator ValidatorFunc[T], options RepositoryOptions) model.Repository[T] {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		// Log error but don't fail - let the actual operation fail with context
	}

	// Create atomic writer with specified options
	atomicWriter := state.NewAtomicWriter(filepath.Join(dir, ".tmp"))

	var cache map[string]*CacheEntry[T]
	if options.EnableCache {
		cache = make(map[string]*CacheEntry[T])
	}

	return &JSONRepository[T]{
		filePath:     filePath,
		validator:    validator,
		atomicWriter: atomicWriter,
		cache:        cache,
		cacheEnabled: options.EnableCache,
		cacheTTL:     options.CacheTTL,
	}
}

// Create persists a new entity and returns an error if the operation fails.
func (r *JSONRepository[T]) Create(ctx context.Context, entity T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate entity
	if r.validator != nil {
		if err := r.validator(entity); err != nil {
			return model.NewValidationError(err.Error()).WithCause(err)
		}
	}

	// Get entity ID using reflection
	entityID, err := r.getEntityID(entity)
	if err != nil {
		return model.NewValidationError("entity must have an ID field").WithCause(err)
	}

	// Check if entity already exists
	exists, err := r.existsInternal(ctx, entityID)
	if err != nil {
		return model.NewInternalError("failed to check entity existence").WithCause(err)
	}
	if exists {
		return model.NewConflictError(fmt.Sprintf("entity with ID '%s' already exists", entityID))
	}

	// Load existing collection
	collection, err := r.loadCollection()
	if err != nil && !os.IsNotExist(err) {
		return model.NewInternalError("failed to load collection").WithCause(err)
	}

	// Initialize collection if it doesn't exist
	if collection == nil {
		collection = make(map[string]T)
	}

	// Add entity to collection
	collection[entityID] = entity

	// Save collection
	if err := r.saveCollection(collection); err != nil {
		return model.NewInternalError("failed to save entity").WithCause(err)
	}

	// Update cache
	r.updateCache(entityID, entity)

	return nil
}

// Read retrieves an entity by its ID.
func (r *JSONRepository[T]) Read(ctx context.Context, id string) (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var zero T

	// Check cache first
	if r.cacheEnabled {
		if entry, found := r.cache[id]; found && entry.ExpiresAt.After(time.Now()) {
			return entry.Entity, nil
		}
	}

	// Load from file
	collection, err := r.loadCollection()
	if err != nil {
		if os.IsNotExist(err) {
			return zero, model.NewNotFoundError("entity").WithContext(id)
		}
		return zero, model.NewInternalError("failed to load collection").WithCause(err)
	}

	entity, exists := collection[id]
	if !exists {
		return zero, model.NewNotFoundError("entity").WithContext(id)
	}

	// Update cache
	r.updateCache(id, entity)

	return entity, nil
}

// Update modifies an existing entity.
func (r *JSONRepository[T]) Update(ctx context.Context, id string, entity T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate entity
	if r.validator != nil {
		if err := r.validator(entity); err != nil {
			return model.NewValidationError(err.Error()).WithCause(err)
		}
	}

	// Load existing collection
	collection, err := r.loadCollection()
	if err != nil {
		if os.IsNotExist(err) {
			return model.NewNotFoundError("entity").WithContext(id)
		}
		return model.NewInternalError("failed to load collection").WithCause(err)
	}

	// Check if entity exists
	if _, exists := collection[id]; !exists {
		return model.NewNotFoundError("entity").WithContext(id)
	}

	// Update entity
	collection[id] = entity

	// Save collection
	if err := r.saveCollection(collection); err != nil {
		return model.NewInternalError("failed to update entity").WithCause(err)
	}

	// Update cache
	r.updateCache(id, entity)

	return nil
}

// Delete removes an entity by its ID.
func (r *JSONRepository[T]) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load existing collection
	collection, err := r.loadCollection()
	if err != nil {
		if os.IsNotExist(err) {
			return model.NewNotFoundError("entity").WithContext(id)
		}
		return model.NewInternalError("failed to load collection").WithCause(err)
	}

	// Check if entity exists
	if _, exists := collection[id]; !exists {
		return model.NewNotFoundError("entity").WithContext(id)
	}

	// Remove entity
	delete(collection, id)

	// Save collection
	if err := r.saveCollection(collection); err != nil {
		return model.NewInternalError("failed to delete entity").WithCause(err)
	}

	// Remove from cache
	if r.cacheEnabled {
		delete(r.cache, id)
	}

	return nil
}

// List retrieves entities based on filter criteria.
func (r *JSONRepository[T]) List(ctx context.Context, filter model.Filter) ([]T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Load collection
	collection, err := r.loadCollection()
	if err != nil {
		if os.IsNotExist(err) {
			return []T{}, nil // Return empty slice if no data exists
		}
		return nil, model.NewInternalError("failed to load collection").WithCause(err)
	}

	var result []T
	for _, entity := range collection {
		// Apply filter if provided
		if filter == nil || filter.Apply(entity) {
			result = append(result, entity)
		}
	}

	return result, nil
}

// Exists checks if an entity with the given ID exists.
func (r *JSONRepository[T]) Exists(ctx context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.existsInternal(ctx, id)
}

// Count returns the total number of entities matching the filter.
func (r *JSONRepository[T]) Count(ctx context.Context, filter model.Filter) (int, error) {
	entities, err := r.List(ctx, filter)
	if err != nil {
		return 0, err
	}
	return len(entities), nil
}

// Internal methods

// existsInternal checks existence without acquiring locks (internal use only).
func (r *JSONRepository[T]) existsInternal(ctx context.Context, id string) (bool, error) {
	// Check cache first
	if r.cacheEnabled {
		if entry, found := r.cache[id]; found && entry.ExpiresAt.After(time.Now()) {
			return true, nil
		}
	}

	// Load from file
	collection, err := r.loadCollection()
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, model.NewInternalError("failed to load collection").WithCause(err)
	}

	_, exists := collection[id]
	return exists, nil
}

// loadCollection loads the entity collection from file.
func (r *JSONRepository[T]) loadCollection() (map[string]T, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, err
	}

	var collection map[string]T
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, model.NewInternalError("failed to parse collection JSON").
			WithCause(err).
			WithContext(r.filePath).
			WithSuggestion("Check if the file contains valid JSON data")
	}

	return collection, nil
}

// saveCollection saves the entity collection to file using atomic operations.
func (r *JSONRepository[T]) saveCollection(collection map[string]T) error {
	options := &state.AtomicWriteOptions{
		Backup:    true,
		Verify:    true,
		GitCommit: true,
		CommitMsg: fmt.Sprintf("Update %s collection", r.getTypeName()),
	}

	if err := r.atomicWriter.WriteJSON(r.filePath, collection, options); err != nil {
		return model.NewInternalError("failed to write collection to disk").
			WithCause(err).
			WithContext(r.filePath).
			WithSuggestions([]string{
				"Check disk space and permissions",
				"Ensure the directory exists and is writable",
				"Verify file is not locked by another process",
			})
	}

	return nil
}

// updateCache updates the cache entry for an entity.
func (r *JSONRepository[T]) updateCache(id string, entity T) {
	if !r.cacheEnabled {
		return
	}

	r.cache[id] = &CacheEntry[T]{
		Entity:    entity,
		ExpiresAt: time.Now().Add(r.cacheTTL),
	}
}

// getEntityID extracts the ID from an entity using reflection.
func (r *JSONRepository[T]) getEntityID(entity T) (string, error) {
	v := reflect.ValueOf(entity)
	
	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", model.NewValidationError("entity cannot be nil")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", model.NewValidationError("entity must be a struct type")
	}

	// Try to find ID field (case-insensitive)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if field.Name == "ID" || field.Name == "Id" || field.Name == "id" {
			idValue := v.Field(i)
			if idValue.Kind() == reflect.String {
				return idValue.String(), nil
			}
		}
	}

	return "", model.NewValidationError("entity must have a string ID field").
		WithSuggestion("Add an 'ID string' field to your entity struct")
}

// getTypeName returns the type name for logging and Git commits.
func (r *JSONRepository[T]) getTypeName() string {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// CleanCache removes expired entries from the cache.
func (r *JSONRepository[T]) CleanCache() {
	if !r.cacheEnabled {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, entry := range r.cache {
		if entry.ExpiresAt.Before(now) {
			delete(r.cache, id)
		}
	}
}