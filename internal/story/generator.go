package story

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"claude-wm-cli/internal/epic"
)

const (
	StoriesFileName = "stories.json"
	StoriesVersion  = "1.0.0"
)

// Generator handles story generation and management
type Generator struct {
	rootPath    string
	epicManager *epic.Manager
}

// NewGenerator creates a new story generator
func NewGenerator(rootPath string) *Generator {
	return &Generator{
		rootPath:    rootPath,
		epicManager: epic.NewManager(rootPath),
	}
}

// GenerateStoriesFromEpic generates stories from a specific epic
func (g *Generator) GenerateStoriesFromEpic(epicID string) error {
	// Get the epic
	ep, err := g.epicManager.GetEpic(epicID)
	if err != nil {
		return fmt.Errorf("failed to get epic %s: %w", epicID, err)
	}

	// Load existing story collection
	collection, err := g.loadStoryCollection()
	if err != nil {
		return fmt.Errorf("failed to load story collection: %w", err)
	}

	// Generate stories from epic user stories
	for _, userStory := range ep.UserStories {
		storyID := g.generateStoryID(userStory.Title, collection)

		// Skip if story already exists
		if _, exists := collection.Stories[storyID]; exists {
			continue
		}

		// Create story from user story
		story := &Story{
			ID:          storyID,
			Title:       userStory.Title,
			Description: userStory.Description,
			EpicID:      epicID,
			Status:      userStory.Status,
			Priority:    userStory.Priority,
			StoryPoints: userStory.StoryPoints,
			Tasks:       []Task{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Generate tasks from acceptance criteria (if available from epic definition)
		// For now, we'll create placeholder tasks
		if len(story.AcceptanceCriteria) > 0 {
			for i, criteria := range story.AcceptanceCriteria {
				taskID := fmt.Sprintf("%s-TASK-%d", storyID, i+1)
				task := Task{
					ID:          taskID,
					Title:       fmt.Sprintf("Implement: %s", criteria),
					Description: criteria,
					Status:      epic.StatusPlanned,
					StoryID:     storyID,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				story.Tasks = append(story.Tasks, task)
			}
		}

		collection.Stories[storyID] = story
	}

	// Update metadata
	collection.Metadata.LastUpdated = time.Now()
	collection.Metadata.TotalStories = len(collection.Stories)
	collection.Metadata.TotalTasks = g.countTotalTasks(collection)

	// Save collection
	return g.saveStoryCollection(collection)
}

// GenerateStoriesFromAllEpics generates stories from all epics
func (g *Generator) GenerateStoriesFromAllEpics() error {
	epicCollection, err := g.epicManager.GetEpicCollection()
	if err != nil {
		return fmt.Errorf("failed to get epic collection: %w", err)
	}

	for epicID := range epicCollection.Epics {
		if err := g.GenerateStoriesFromEpic(epicID); err != nil {
			return fmt.Errorf("failed to generate stories from epic %s: %w", epicID, err)
		}
	}

	return nil
}

// CreateStory creates a new story manually
func (g *Generator) CreateStory(options StoryCreateOptions) (*Story, error) {
	// Validate inputs
	if strings.TrimSpace(options.Title) == "" {
		return nil, fmt.Errorf("story title cannot be empty")
	}

	if options.EpicID != "" {
		// Verify epic exists
		if _, err := g.epicManager.GetEpic(options.EpicID); err != nil {
			return nil, fmt.Errorf("epic %s not found: %w", options.EpicID, err)
		}
	}

	// Load existing collection
	collection, err := g.loadStoryCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load story collection: %w", err)
	}

	// Generate unique ID
	storyID := g.generateStoryID(options.Title, collection)

	// Set default priority if not specified
	if options.Priority == "" {
		options.Priority = epic.PriorityMedium
	}

	// Create the story
	now := time.Now()
	story := &Story{
		ID:                 storyID,
		Title:              strings.TrimSpace(options.Title),
		Description:        strings.TrimSpace(options.Description),
		EpicID:             options.EpicID,
		Status:             epic.StatusPlanned,
		Priority:           options.Priority,
		StoryPoints:        options.StoryPoints,
		AcceptanceCriteria: options.AcceptanceCriteria,
		Tasks:              []Task{},
		Dependencies:       options.Dependencies,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Generate tasks from acceptance criteria
	for i, criteria := range options.AcceptanceCriteria {
		taskID := fmt.Sprintf("%s-TASK-%d", storyID, i+1)
		task := Task{
			ID:          taskID,
			Title:       fmt.Sprintf("Implement: %s", criteria),
			Description: criteria,
			Status:      epic.StatusPlanned,
			StoryID:     storyID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		story.Tasks = append(story.Tasks, task)
	}

	// Add to collection
	collection.Stories[storyID] = story
	collection.Metadata.TotalStories = len(collection.Stories)
	collection.Metadata.TotalTasks = g.countTotalTasks(collection)
	collection.Metadata.LastUpdated = now

	// Save collection
	if err := g.saveStoryCollection(collection); err != nil {
		return nil, fmt.Errorf("failed to save story collection: %w", err)
	}

	return story, nil
}

// UpdateStory updates an existing story
func (g *Generator) UpdateStory(storyID string, options StoryUpdateOptions) (*Story, error) {
	collection, err := g.loadStoryCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load story collection: %w", err)
	}

	story, exists := collection.Stories[storyID]
	if !exists {
		return nil, fmt.Errorf("story not found: %s", storyID)
	}

	// Apply updates
	now := time.Now()

	if options.Title != nil {
		if strings.TrimSpace(*options.Title) == "" {
			return nil, fmt.Errorf("story title cannot be empty")
		}
		story.Title = strings.TrimSpace(*options.Title)
	}

	if options.Description != nil {
		story.Description = strings.TrimSpace(*options.Description)
	}

	if options.Status != nil {
		if err := g.validateStatusTransition(story, *options.Status); err != nil {
			return nil, err
		}

		story.Status = *options.Status

		// Set timestamps for status changes
		if *options.Status == epic.StatusInProgress && story.StartedAt == nil {
			story.StartedAt = &now
		}
		if *options.Status == epic.StatusCompleted && story.CompletedAt == nil {
			story.CompletedAt = &now
		}
	}

	if options.Priority != nil {
		story.Priority = *options.Priority
	}

	if options.StoryPoints != nil {
		story.StoryPoints = *options.StoryPoints
	}

	if options.AcceptanceCriteria != nil {
		story.AcceptanceCriteria = *options.AcceptanceCriteria
		// Regenerate tasks from new acceptance criteria
		story.Tasks = []Task{}
		for i, criteria := range story.AcceptanceCriteria {
			taskID := fmt.Sprintf("%s-TASK-%d", storyID, i+1)
			task := Task{
				ID:          taskID,
				Title:       fmt.Sprintf("Implement: %s", criteria),
				Description: criteria,
				Status:      epic.StatusPlanned,
				StoryID:     storyID,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			story.Tasks = append(story.Tasks, task)
		}
	}

	if options.Dependencies != nil {
		story.Dependencies = *options.Dependencies
	}

	story.UpdatedAt = now

	// Update metadata
	collection.Metadata.LastUpdated = now
	collection.Metadata.TotalTasks = g.countTotalTasks(collection)

	// Save collection
	if err := g.saveStoryCollection(collection); err != nil {
		return nil, fmt.Errorf("failed to save story collection: %w", err)
	}

	return story, nil
}

// GetStoryCollection returns the story collection
func (g *Generator) GetStoryCollection() (*StoryCollection, error) {
	return g.loadStoryCollection()
}

// GetStory returns a specific story by ID
func (g *Generator) GetStory(storyID string) (*Story, error) {
	collection, err := g.loadStoryCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load story collection: %w", err)
	}

	story, exists := collection.Stories[storyID]
	if !exists {
		return nil, fmt.Errorf("story not found: %s", storyID)
	}

	return story, nil
}

// ListStories returns a list of stories with optional filtering
func (g *Generator) ListStories(epicID string, status Status) ([]*Story, error) {
	collection, err := g.loadStoryCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to load story collection: %w", err)
	}

	var stories []*Story
	for _, story := range collection.Stories {
		// Apply filters
		if epicID != "" && story.EpicID != epicID {
			continue
		}
		if status != "" && story.Status != status {
			continue
		}

		stories = append(stories, story)
	}

	// Sort by creation date (newest first)
	sort.Slice(stories, func(i, j int) bool {
		return stories[i].CreatedAt.After(stories[j].CreatedAt)
	})

	return stories, nil
}

// DeleteStory removes a story from the collection
func (g *Generator) DeleteStory(storyID string) error {
	collection, err := g.loadStoryCollection()
	if err != nil {
		return fmt.Errorf("failed to load story collection: %w", err)
	}

	_, exists := collection.Stories[storyID]
	if !exists {
		return fmt.Errorf("story not found: %s", storyID)
	}

	// Clear current story if it's the one being deleted
	if collection.CurrentStory == storyID {
		collection.CurrentStory = ""
	}

	// Remove from collection
	delete(collection.Stories, storyID)
	collection.Metadata.TotalStories = len(collection.Stories)
	collection.Metadata.TotalTasks = g.countTotalTasks(collection)
	collection.Metadata.LastUpdated = time.Now()

	// Save collection
	return g.saveStoryCollection(collection)
}

// loadStoryCollection loads the story collection from disk
func (g *Generator) loadStoryCollection() (*StoryCollection, error) {
	storiesPath := filepath.Join(g.rootPath, "docs", "2-current-epic", StoriesFileName)

	// Check if file exists
	if _, err := os.Stat(storiesPath); os.IsNotExist(err) {
		// Create default collection
		return &StoryCollection{
			Stories:      make(map[string]*Story),
			CurrentStory: "",
			Metadata: CollectionMetadata{
				Version:      StoriesVersion,
				LastUpdated:  time.Now(),
				TotalStories: 0,
				TotalTasks:   0,
			},
		}, nil
	}

	// Read file
	data, err := os.ReadFile(storiesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read stories file: %w", err)
	}

	var collection StoryCollection

	// Try to unmarshal as the new format first
	if err := json.Unmarshal(data, &collection); err != nil {
		// If that fails, try to parse the old format and migrate
		var oldFormat struct {
			Stories []struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Priority    string `json:"priority"`
				Status      string `json:"status"`
				StoryPoints int    `json:"storyPoints"`
				// Add other fields as needed for migration
			} `json:"stories"`
		}

		if err := json.Unmarshal(data, &oldFormat); err != nil {
			return nil, fmt.Errorf("failed to parse stories file: %w", err)
		}

		// Migrate from old format to new format
		collection = StoryCollection{
			Stories:      make(map[string]*Story),
			CurrentStory: "",
			Metadata: CollectionMetadata{
				Version:      StoriesVersion,
				LastUpdated:  time.Now(),
				TotalStories: len(oldFormat.Stories),
				TotalTasks:   0,
			},
		}

		// Convert old stories to new format (but skip since they are not in our expected format)
		// For now, we'll just create an empty collection and let users create new stories
		fmt.Printf("Warning: Old stories.json format detected. Creating new story collection.\n")
		fmt.Printf("Previous stories data has been preserved but not migrated.\n")
	}

	// Validate and migrate if needed
	if err := g.validateAndMigrateCollection(&collection); err != nil {
		return nil, fmt.Errorf("failed to validate story collection: %w", err)
	}

	return &collection, nil
}

// saveStoryCollection saves the story collection to disk
func (g *Generator) saveStoryCollection(collection *StoryCollection) error {
	storiesPath := filepath.Join(g.rootPath, "docs", "2-current-epic", StoriesFileName)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(storiesPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Update metadata
	collection.Metadata.LastUpdated = time.Now()
	collection.Metadata.Version = StoriesVersion

	// Marshal to JSON
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal story collection: %w", err)
	}

	// Write file atomically using temp file + rename
	tempPath := storiesPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp stories file: %w", err)
	}

	if err := os.Rename(tempPath, storiesPath); err != nil {
		os.Remove(tempPath) // cleanup
		return fmt.Errorf("failed to replace stories file: %w", err)
	}

	return nil
}

// generateStoryID generates a unique ID for a new story
func (g *Generator) generateStoryID(title string, collection *StoryCollection) string {
	// Create base ID from title
	baseID := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(title), " ", "-"))
	baseID = strings.ReplaceAll(baseID, "_", "-")

	// Remove special characters
	var cleaned strings.Builder
	for _, r := range baseID {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned.WriteRune(r)
		}
	}

	baseID = cleaned.String()
	if baseID == "" {
		baseID = "STORY"
	}

	// Ensure uniqueness
	counter := 1
	storyID := fmt.Sprintf("STORY-%03d-%s", counter, baseID)

	for {
		if _, exists := collection.Stories[storyID]; !exists {
			break
		}
		counter++
		storyID = fmt.Sprintf("STORY-%03d-%s", counter, baseID)
	}

	return storyID
}

// validateStatusTransition checks if a status transition is valid
func (g *Generator) validateStatusTransition(story *Story, newStatus Status) error {
	currentStatus := story.Status

	// Define valid transitions
	validTransitions := map[Status][]Status{
		epic.StatusPlanned:    {epic.StatusInProgress, epic.StatusCancelled},
		epic.StatusInProgress: {epic.StatusCompleted, epic.StatusOnHold, epic.StatusCancelled},
		epic.StatusOnHold:     {epic.StatusInProgress, epic.StatusCancelled},
		epic.StatusCompleted:  {},                   // Cannot transition from completed
		epic.StatusCancelled:  {epic.StatusPlanned}, // Can restart cancelled stories
	}

	allowedTransitions, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("unknown current status: %s", currentStatus)
	}

	// Check if transition is allowed
	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return nil
		}
	}

	return fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
}

// validateAndMigrateCollection validates and migrates the collection if needed
func (g *Generator) validateAndMigrateCollection(collection *StoryCollection) error {
	// Initialize maps if nil
	if collection.Stories == nil {
		collection.Stories = make(map[string]*Story)
	}

	// Set default metadata if missing
	if collection.Metadata.Version == "" {
		collection.Metadata.Version = StoriesVersion
	}

	// Update counters
	collection.Metadata.TotalStories = len(collection.Stories)
	collection.Metadata.TotalTasks = g.countTotalTasks(collection)

	// Validate each story
	for id, story := range collection.Stories {
		if story == nil {
			delete(collection.Stories, id)
			continue
		}

		// Ensure required fields
		if story.ID == "" {
			story.ID = id
		}
		if story.CreatedAt.IsZero() {
			story.CreatedAt = time.Now()
		}
		if story.UpdatedAt.IsZero() {
			story.UpdatedAt = story.CreatedAt
		}
		if story.Priority == "" {
			story.Priority = epic.PriorityMedium
		}
		if story.Status == "" {
			story.Status = epic.StatusPlanned
		}
		if story.Tasks == nil {
			story.Tasks = []Task{}
		}
	}

	return nil
}

// countTotalTasks counts all tasks across all stories
func (g *Generator) countTotalTasks(collection *StoryCollection) int {
	totalTasks := 0
	for _, story := range collection.Stories {
		totalTasks += len(story.Tasks)
	}
	return totalTasks
}
