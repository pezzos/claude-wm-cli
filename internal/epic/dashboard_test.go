package epic

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboard_NewDashboard(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)
	dashboard := NewDashboard(manager)

	assert.NotNil(t, dashboard)
	assert.Equal(t, manager, dashboard.manager)
}

func TestDashboard_GetEpicDashboardData(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)
	dashboard := NewDashboard(manager)

	// Create a test epic
	epic, err := manager.CreateEpic(EpicCreateOptions{
		Title:       "Test Epic",
		Description: "Test epic for dashboard",
		Priority:    PriorityHigh,
		Duration:    "2 weeks",
	})
	require.NoError(t, err)

	// Start the epic
	now := time.Now()
	epic.StartDate = &now
	statusInProgress := StatusInProgress
	epic, err = manager.UpdateEpic(epic.ID, EpicUpdateOptions{
		Status: &statusInProgress,
	})
	require.NoError(t, err)

	// Add some user stories to the epic
	epic.UserStories = []UserStory{
		{
			ID:          "STORY-1",
			Title:       "Story 1",
			Status:      StatusCompleted,
			Priority:    PriorityHigh,
			StoryPoints: 5,
		},
		{
			ID:          "STORY-2",
			Title:       "Story 2",
			Status:      StatusPlanned,
			Priority:    PriorityMedium,
			StoryPoints: 3,
		},
	}

	// Get dashboard data
	data := dashboard.GetEpicDashboardData(epic)
	require.NotNil(t, data)

	// Verify epic data
	assert.Equal(t, epic.ID, data.Epic.ID)
	assert.Equal(t, "Test Epic", data.Epic.Title)

	// Verify progress metrics
	assert.Equal(t, 2, data.ProgressMetrics.TotalStories)
	assert.Equal(t, 1, data.ProgressMetrics.StoriesCompleted)
	assert.Equal(t, 0, data.ProgressMetrics.StoriesInProgress)
	assert.Equal(t, 1, data.ProgressMetrics.StoriesPlanned)
	assert.Equal(t, 50.0, data.ProgressMetrics.CompletionPercentage)
	assert.Equal(t, 8, data.ProgressMetrics.StoryPointsTotal)
	assert.Equal(t, 5, data.ProgressMetrics.StoryPointsCompleted)

	// Verify risk assessment
	assert.NotEqual(t, "", string(data.RiskLevel))

	// Verify velocity metrics
	assert.True(t, data.Velocity.StoriesPerDay >= 0)

	// Verify timeline metrics
	assert.True(t, data.Timeline.DaysActive >= 0)
	assert.Equal(t, "2 weeks", data.Timeline.OriginalEstimate)
}

func TestDashboard_CalculateProgressMetrics(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)
	dashboard := NewDashboard(manager)

	// Create test epic with user stories
	epic := &Epic{
		ID:    "TEST-EPIC",
		Title: "Progress Test Epic",
		UserStories: []UserStory{
			{
				ID:          "STORY-1",
				Status:      StatusCompleted,
				StoryPoints: 5,
			},
			{
				ID:          "STORY-2",
				Status:      StatusInProgress,
				StoryPoints: 3,
			},
			{
				ID:          "STORY-3",
				Status:      StatusPlanned,
				StoryPoints: 2,
			},
		},
	}

	progress := dashboard.calculateProgressMetrics(epic)

	assert.Equal(t, 3, progress.TotalStories)
	assert.Equal(t, 1, progress.StoriesCompleted)
	assert.Equal(t, 1, progress.StoriesInProgress)
	assert.Equal(t, 1, progress.StoriesPlanned)
	assert.InDelta(t, 33.33, progress.CompletionPercentage, 0.1)
	assert.Equal(t, 10, progress.StoryPointsTotal)
	assert.Equal(t, 5, progress.StoryPointsCompleted)
}

func TestDashboard_AssessRiskLevel(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)
	dashboard := NewDashboard(manager)

	// Test low risk epic (new, high progress)
	lowRiskEpic := &Epic{
		Status:   StatusInProgress,
		Priority: PriorityMedium,
	}
	lowRiskProgress := ProgressSummary{
		CompletionPercentage: 80.0,
		TotalStories:         5,
		StoriesInProgress:    1,
	}
	risk := dashboard.assessRiskLevel(lowRiskEpic, lowRiskProgress)
	assert.Equal(t, RiskLow, risk)

	// Test high risk epic (old, low progress, critical priority)
	oldDate := time.Now().Add(-45 * 24 * time.Hour) // 45 days ago
	highRiskEpic := &Epic{
		Status:    StatusInProgress,
		Priority:  PriorityCritical,
		StartDate: &oldDate,
	}
	highRiskProgress := ProgressSummary{
		CompletionPercentage: 30.0,
		TotalStories:         5,
		StoriesInProgress:    0, // No active work
	}
	risk = dashboard.assessRiskLevel(highRiskEpic, highRiskProgress)
	assert.True(t, risk == RiskHigh || risk == RiskCritical)

	// Test medium risk epic
	mediumDate := time.Now().Add(-20 * 24 * time.Hour) // 20 days ago
	mediumRiskEpic := &Epic{
		Status:    StatusInProgress,
		Priority:  PriorityHigh,
		StartDate: &mediumDate,
	}
	mediumRiskProgress := ProgressSummary{
		CompletionPercentage: 15.0,
		TotalStories:         3,
		StoriesInProgress:    1,
	}
	risk = dashboard.assessRiskLevel(mediumRiskEpic, mediumRiskProgress)
	assert.True(t, risk == RiskMedium || risk == RiskHigh)
}

func TestDashboard_CalculateVelocityMetrics(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)
	dashboard := NewDashboard(manager)

	// Test epic with no start date (should return empty metrics)
	epic := &Epic{Status: StatusPlanned}
	velocity := dashboard.calculateVelocityMetrics(epic)
	assert.Equal(t, 0.0, velocity.StoriesPerDay)

	// Test epic with start date and completed stories
	startDate := time.Now().Add(-10 * 24 * time.Hour) // 10 days ago
	epicWithVelocity := &Epic{
		Status:    StatusInProgress,
		StartDate: &startDate,
		UserStories: []UserStory{
			{Status: StatusCompleted, StoryPoints: 5},
			{Status: StatusCompleted, StoryPoints: 3},
			{Status: StatusInProgress, StoryPoints: 2},
			{Status: StatusPlanned, StoryPoints: 1},
		},
	}

	velocity = dashboard.calculateVelocityMetrics(epicWithVelocity)
	assert.True(t, velocity.StoriesPerDay > 0)
	assert.InDelta(t, 2.0/10.0, velocity.StoriesPerDay, 0.01)     // 2 completed stories over 10 days
	assert.InDelta(t, 8.0/10.0, velocity.StoryPointsPerDay, 0.01) // 8 points over 10 days
	assert.InDelta(t, 5.0, velocity.AverageStoryDays, 0.01)       // 10 days / 2 stories
	assert.NotEqual(t, "", velocity.CompletionTrend)
}

func TestDashboard_CalculateTimelineMetrics(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)
	dashboard := NewDashboard(manager)

	// Test epic with no start date
	epic := &Epic{
		Duration: "2 weeks",
	}
	timeline := dashboard.calculateTimelineMetrics(epic, ProgressSummary{}, VelocityMetrics{})
	assert.Equal(t, "2 weeks", timeline.OriginalEstimate)
	assert.Equal(t, 0, timeline.DaysActive)

	// Test epic with start date
	startDate := time.Now().Add(-5 * 24 * time.Hour) // 5 days ago
	epicWithTimeline := &Epic{
		Duration:  "1 weeks", // Should be overdue
		StartDate: &startDate,
	}

	progress := ProgressSummary{
		TotalStories:     4,
		StoriesCompleted: 2,
	}

	velocity := VelocityMetrics{
		StoriesPerDay: 0.5, // 0.5 stories per day
	}

	timeline = dashboard.calculateTimelineMetrics(epicWithTimeline, progress, velocity)
	assert.Equal(t, "1 weeks", timeline.OriginalEstimate)
	assert.Equal(t, 5, timeline.DaysActive)
	assert.Equal(t, 4, timeline.EstimatedDaysRemaining) // 2 remaining stories / 0.5 per day
	assert.False(t, timeline.IsOverdue)                 // 5 days < 7 days (1 week)
}

func TestDashboard_CreateProgressBar(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)
	dashboard := NewDashboard(manager)

	// Test 0% progress
	bar := dashboard.createProgressBar(0.0, 10)
	assert.Equal(t, "[░░░░░░░░░░]", bar)

	// Test 50% progress
	bar = dashboard.createProgressBar(50.0, 10)
	assert.Equal(t, "[█████░░░░░]", bar)

	// Test 100% progress
	bar = dashboard.createProgressBar(100.0, 10)
	assert.Equal(t, "[██████████]", bar)

	// Test custom width
	bar = dashboard.createProgressBar(25.0, 4)
	assert.Equal(t, "[█░░░]", bar)
}

func TestDashboard_IconMethods(t *testing.T) {
	tempDir := t.TempDir()
	setupTestDirs(t, tempDir)

	manager := NewManager(tempDir)
	dashboard := NewDashboard(manager)

	// Test status icons
	assert.Equal(t, "📋", dashboard.getStatusIcon(StatusPlanned))
	assert.Equal(t, "🚧", dashboard.getStatusIcon(StatusInProgress))
	assert.Equal(t, "✅", dashboard.getStatusIcon(StatusCompleted))
	assert.Equal(t, "❌", dashboard.getStatusIcon(StatusCancelled))
	assert.Equal(t, "⏸️", dashboard.getStatusIcon(StatusOnHold))

	// Test priority icons
	assert.Equal(t, "🟢", dashboard.getPriorityIcon(PriorityLow))
	assert.Equal(t, "🟡", dashboard.getPriorityIcon(PriorityMedium))
	assert.Equal(t, "🟠", dashboard.getPriorityIcon(PriorityHigh))
	assert.Equal(t, "🔴", dashboard.getPriorityIcon(PriorityCritical))

	// Test risk icons
	assert.Equal(t, "🟢", dashboard.getRiskIcon(RiskLow))
	assert.Equal(t, "🟡", dashboard.getRiskIcon(RiskMedium))
	assert.Equal(t, "🟠", dashboard.getRiskIcon(RiskHigh))
	assert.Equal(t, "🔴", dashboard.getRiskIcon(RiskCritical))
}

func TestDashboard_HelperFunctions(t *testing.T) {
	// Test percentage function
	assert.Equal(t, 50.0, percentage(1, 2))
	assert.Equal(t, 0.0, percentage(0, 5))
	assert.Equal(t, 0.0, percentage(1, 0)) // Division by zero
	assert.Equal(t, 100.0, percentage(5, 5))

	// Test truncateText function
	assert.Equal(t, "Hello", truncateText("Hello", 10))
	assert.Equal(t, "Hello W...", truncateText("Hello World", 10))
	assert.Equal(t, "Hello", truncateText("Hello", 5)) // Same length, no truncation needed
	assert.Equal(t, "H...", truncateText("Hello", 4))  // Truncation needed
}

// Helper function to setup test directories
func setupTestDirs(t *testing.T, tempDir string) {
	docsDir := filepath.Join(tempDir, "docs", "1-project")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	currentEpicDir := filepath.Join(tempDir, "docs", "2-current-epic")
	err = os.MkdirAll(currentEpicDir, 0755)
	require.NoError(t, err)

	currentTaskDir := filepath.Join(tempDir, "docs", "3-current-task")
	err = os.MkdirAll(currentTaskDir, 0755)
	require.NoError(t, err)
}
