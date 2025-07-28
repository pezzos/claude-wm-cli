package epic

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Dashboard provides epic progress visualization and analytics
type Dashboard struct {
	manager *Manager
}

// NewDashboard creates a new epic dashboard
func NewDashboard(manager *Manager) *Dashboard {
	return &Dashboard{
		manager: manager,
	}
}

// EpicDashboardData contains comprehensive epic progress data
type EpicDashboardData struct {
	Epic            *Epic
	ProgressMetrics ProgressSummary
	RiskLevel       RiskLevel
	Velocity        VelocityMetrics
	Timeline        TimelineMetrics
}

// ProgressSummary provides detailed progress information
type ProgressSummary struct {
	CompletionPercentage   float64
	StoriesCompleted       int
	StoriesInProgress      int
	StoriesPlanned         int
	TotalStories          int
	StoryPointsCompleted  int
	StoryPointsTotal      int
}

// RiskLevel indicates the risk status of an epic
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// VelocityMetrics tracks epic velocity and productivity
type VelocityMetrics struct {
	StoriesPerDay      float64
	StoryPointsPerDay  float64
	AverageStoryDays   float64
	CompletionTrend    string // "improving", "stable", "declining"
}

// TimelineMetrics provides timeline analysis
type TimelineMetrics struct {
	DaysActive           int
	EstimatedDaysRemaining int
	OriginalEstimate     string
	IsOverdue           bool
	DaysOverdue         int
}

// DisplayEpicDashboard shows a comprehensive dashboard for all epics
func (d *Dashboard) DisplayEpicDashboard() error {
	// Get all epics
	epics, err := d.manager.ListEpics(EpicListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get epics: %w", err)
	}

	if len(epics) == 0 {
		fmt.Println("📊 Epic Dashboard")
		fmt.Println("=================")
		fmt.Println()
		fmt.Println("No epics found. Create your first epic to get started!")
		fmt.Println()
		fmt.Println("💡 Next steps:")
		fmt.Println("   • Create an epic: claude-wm-cli epic create \"Epic Title\"")
		return nil
	}

	// Gather dashboard data for all epics
	var dashboardData []*EpicDashboardData
	for _, epic := range epics {
		data := d.GetEpicDashboardData(epic)
		dashboardData = append(dashboardData, data)
	}

	// Sort by priority and status
	sort.Slice(dashboardData, func(i, j int) bool {
		// Active epics first
		if dashboardData[i].Epic.Status == StatusInProgress && dashboardData[j].Epic.Status != StatusInProgress {
			return true
		}
		if dashboardData[i].Epic.Status != StatusInProgress && dashboardData[j].Epic.Status == StatusInProgress {
			return false
		}
		
		// Then by priority
		priorityOrder := map[Priority]int{
			PriorityCritical: 4,
			PriorityHigh:     3,
			PriorityMedium:   2,
			PriorityLow:      1,
		}
		return priorityOrder[dashboardData[i].Epic.Priority] > priorityOrder[dashboardData[j].Epic.Priority]
	})

	// Display header
	fmt.Println("📊 Epic Progress Dashboard")
	fmt.Println("==========================")
	fmt.Println()

	// Display summary
	d.displaySummary(dashboardData)
	fmt.Println()

	// Display each epic
	for _, data := range dashboardData {
		d.displayEpicCard(data)
		fmt.Println()
	}

	// Display risk analysis
	d.displayRiskAnalysis(dashboardData)

	return nil
}

// GetEpicDashboardData gathers comprehensive data for a specific epic
func (d *Dashboard) GetEpicDashboardData(epic *Epic) *EpicDashboardData {
	// Calculate progress metrics from epic's user stories
	progressMetrics := d.calculateProgressMetrics(epic)

	// Assess risk level
	riskLevel := d.assessRiskLevel(epic, progressMetrics)

	// Calculate velocity metrics
	velocityMetrics := d.calculateVelocityMetrics(epic)

	// Calculate timeline metrics
	timelineMetrics := d.calculateTimelineMetrics(epic, progressMetrics, velocityMetrics)

	return &EpicDashboardData{
		Epic:            epic,
		ProgressMetrics: progressMetrics,
		RiskLevel:       riskLevel,
		Velocity:        velocityMetrics,
		Timeline:        timelineMetrics,
	}
}

// displaySummary shows an overview of all epics
func (d *Dashboard) displaySummary(data []*EpicDashboardData) {
	var totalEpics, completedEpics, activeEpics, plannedEpics int
	var totalStories, completedStories int
	var totalPoints, completedPoints int

	for _, epic := range data {
		totalEpics++
		totalStories += epic.ProgressMetrics.TotalStories
		completedStories += epic.ProgressMetrics.StoriesCompleted
		totalPoints += epic.ProgressMetrics.StoryPointsTotal
		completedPoints += epic.ProgressMetrics.StoryPointsCompleted

		switch epic.Epic.Status {
		case StatusCompleted:
			completedEpics++
		case StatusInProgress:
			activeEpics++
		case StatusPlanned:
			plannedEpics++
		}
	}

	fmt.Printf("📈 Project Overview\n")
	fmt.Printf("   Epics:        %d total (%d active, %d completed, %d planned)\n", totalEpics, activeEpics, completedEpics, plannedEpics)
	fmt.Printf("   Stories:      %d/%d completed (%.1f%%)\n", completedStories, totalStories, percentage(completedStories, totalStories))
	fmt.Printf("   Story Points: %d/%d completed (%.1f%%)\n", completedPoints, totalPoints, percentage(completedPoints, totalPoints))
}

// displayEpicCard shows detailed information for one epic
func (d *Dashboard) displayEpicCard(data *EpicDashboardData) {
	epic := data.Epic
	metrics := data.ProgressMetrics

	// Header
	statusIcon := d.getStatusIcon(epic.Status)
	priorityIcon := d.getPriorityIcon(epic.Priority)
	riskIcon := d.getRiskIcon(data.RiskLevel)
	
	fmt.Printf("┌─ %s %s | %s %s | %s %s\n", statusIcon, epic.Status, priorityIcon, epic.Priority, riskIcon, data.RiskLevel)
	fmt.Printf("│  📋 %s\n", epic.Title)
	fmt.Printf("│  🆔 %s\n", epic.ID)

	// Progress bar
	progressBar := d.createProgressBar(metrics.CompletionPercentage, 30)
	fmt.Printf("│  %s %.1f%%\n", progressBar, metrics.CompletionPercentage)

	// Progress details
	if metrics.TotalStories > 0 {
		fmt.Printf("│  📊 Stories: %d/%d completed", metrics.StoriesCompleted, metrics.TotalStories)
		if metrics.StoriesInProgress > 0 {
			fmt.Printf(" (%d in progress)", metrics.StoriesInProgress)
		}
		fmt.Printf("\n")
	}

	if metrics.StoryPointsTotal > 0 {
		fmt.Printf("│  🎯 Points:  %d/%d completed\n", metrics.StoryPointsCompleted, metrics.StoryPointsTotal)
	}

	// Timeline information
	timeline := data.Timeline
	if timeline.DaysActive > 0 {
		fmt.Printf("│  ⏱️  Duration: %d days active", timeline.DaysActive)
		if timeline.EstimatedDaysRemaining > 0 {
			fmt.Printf(", ~%d days remaining", timeline.EstimatedDaysRemaining)
		}
		if timeline.IsOverdue {
			fmt.Printf(" (⚠️ %d days overdue)", timeline.DaysOverdue)
		}
		fmt.Printf("\n")
	}

	// Velocity information
	velocity := data.Velocity
	if velocity.StoriesPerDay > 0 {
		fmt.Printf("│  🚀 Velocity: %.1f stories/day", velocity.StoriesPerDay)
		if velocity.CompletionTrend != "stable" {
			fmt.Printf(" (%s)", velocity.CompletionTrend)
		}
		fmt.Printf("\n")
	}

	// Show user stories if available
	if len(epic.UserStories) > 0 {
		fmt.Printf("│  📚 User Stories:\n")
		count := len(epic.UserStories)
		if count > 3 {
			count = 3
		}
		for i := 0; i < count; i++ {
			story := epic.UserStories[i]
			storyIcon := d.getStatusIcon(story.Status)
			fmt.Printf("│     %s %s\n", storyIcon, truncateText(story.Title, 40))
		}
		if len(epic.UserStories) > 3 {
			fmt.Printf("│     ... and %d more\n", len(epic.UserStories)-3)
		}
	}

	fmt.Printf("└─\n")
}

// displayRiskAnalysis shows epics that need attention
func (d *Dashboard) displayRiskAnalysis(data []*EpicDashboardData) {
	var highRiskEpics []*EpicDashboardData
	var overdueEpics []*EpicDashboardData
	var stagnantEpics []*EpicDashboardData

	for _, epic := range data {
		if epic.RiskLevel == RiskHigh || epic.RiskLevel == RiskCritical {
			highRiskEpics = append(highRiskEpics, epic)
		}
		if epic.Timeline.IsOverdue {
			overdueEpics = append(overdueEpics, epic)
		}
		if epic.Velocity.CompletionTrend == "declining" && epic.Epic.Status == StatusInProgress {
			stagnantEpics = append(stagnantEpics, epic)
		}
	}

	if len(highRiskEpics) > 0 || len(overdueEpics) > 0 || len(stagnantEpics) > 0 {
		fmt.Println("⚠️  Risk Analysis")
		fmt.Println("================")
		fmt.Println()

		if len(highRiskEpics) > 0 {
			fmt.Printf("🔴 High Risk Epics (%d):\n", len(highRiskEpics))
			for _, epic := range highRiskEpics {
				fmt.Printf("   • %s - %s\n", epic.Epic.ID, epic.Epic.Title)
			}
			fmt.Println()
		}

		if len(overdueEpics) > 0 {
			fmt.Printf("⏰ Overdue Epics (%d):\n", len(overdueEpics))
			for _, epic := range overdueEpics {
				fmt.Printf("   • %s - %d days overdue\n", epic.Epic.ID, epic.Timeline.DaysOverdue)
			}
			fmt.Println()
		}

		if len(stagnantEpics) > 0 {
			fmt.Printf("📉 Declining Velocity (%d):\n", len(stagnantEpics))
			for _, epic := range stagnantEpics {
				fmt.Printf("   • %s - %.1f stories/day\n", epic.Epic.ID, epic.Velocity.StoriesPerDay)
			}
			fmt.Println()
		}

		fmt.Println("💡 Recommendations:")
		if len(highRiskEpics) > 0 {
			fmt.Println("   • Review high-risk epics for blockers")
		}
		if len(overdueEpics) > 0 {
			fmt.Println("   • Update timelines for overdue epics")
		}
		if len(stagnantEpics) > 0 {
			fmt.Println("   • Investigate velocity decline causes")
		}
	}
}

// Helper methods for calculations

func (d *Dashboard) calculateProgressMetrics(epic *Epic) ProgressSummary {
	var storiesCompleted, storiesInProgress, storiesPlanned int
	var storyPointsCompleted, storyPointsTotal int

	for _, story := range epic.UserStories {
		storyPointsTotal += story.StoryPoints
		
		switch story.Status {
		case StatusCompleted:
			storiesCompleted++
			storyPointsCompleted += story.StoryPoints
		case StatusInProgress:
			storiesInProgress++
		case StatusPlanned:
			storiesPlanned++
		}
	}

	totalStories := len(epic.UserStories)
	completionPercentage := 0.0
	if totalStories > 0 {
		completionPercentage = float64(storiesCompleted) / float64(totalStories) * 100.0
	}

	return ProgressSummary{
		CompletionPercentage:  completionPercentage,
		StoriesCompleted:      storiesCompleted,
		StoriesInProgress:     storiesInProgress,
		StoriesPlanned:        storiesPlanned,
		TotalStories:         totalStories,
		StoryPointsCompleted: storyPointsCompleted,
		StoryPointsTotal:     storyPointsTotal,
	}
}

func (d *Dashboard) assessRiskLevel(epic *Epic, progress ProgressSummary) RiskLevel {
	riskScore := 0

	// Check timeline risk
	if epic.StartDate != nil {
		daysActive := int(time.Since(*epic.StartDate).Hours() / 24)
		if daysActive > 30 && progress.CompletionPercentage < 50 {
			riskScore += 2
		}
		if daysActive > 60 && progress.CompletionPercentage < 80 {
			riskScore += 3
		}
	}

	// Check progress risk
	if progress.TotalStories > 0 {
		if progress.CompletionPercentage < 20 && epic.Status == StatusInProgress {
			riskScore += 1
		}
		if progress.StoriesInProgress == 0 && epic.Status == StatusInProgress {
			riskScore += 2
		}
	}

	// Check priority vs progress mismatch
	if epic.Priority == PriorityCritical && progress.CompletionPercentage < 50 {
		riskScore += 2
	}

	switch {
	case riskScore >= 5:
		return RiskCritical
	case riskScore >= 3:
		return RiskHigh
	case riskScore >= 1:
		return RiskMedium
	default:
		return RiskLow
	}
}

func (d *Dashboard) calculateVelocityMetrics(epic *Epic) VelocityMetrics {
	if epic.StartDate == nil || len(epic.UserStories) == 0 {
		return VelocityMetrics{}
	}

	daysActive := time.Since(*epic.StartDate).Hours() / 24
	if daysActive <= 0 {
		return VelocityMetrics{}
	}

	completedStories := 0
	completedPoints := 0
	for _, story := range epic.UserStories {
		if story.Status == StatusCompleted {
			completedStories++
			completedPoints += story.StoryPoints
		}
	}

	storiesPerDay := float64(completedStories) / daysActive
	storyPointsPerDay := float64(completedPoints) / daysActive
	
	avgStoryDays := 0.0
	if completedStories > 0 {
		avgStoryDays = daysActive / float64(completedStories)
	}

	// Simple trend analysis (would be more sophisticated with historical data)
	trend := "stable"
	if storiesPerDay > 0.5 {
		trend = "improving"
	} else if storiesPerDay < 0.1 && daysActive > 7 {
		trend = "declining"
	}

	return VelocityMetrics{
		StoriesPerDay:     storiesPerDay,
		StoryPointsPerDay: storyPointsPerDay,
		AverageStoryDays:  avgStoryDays,
		CompletionTrend:   trend,
	}
}

func (d *Dashboard) calculateTimelineMetrics(epic *Epic, progress ProgressSummary, velocity VelocityMetrics) TimelineMetrics {
	timeline := TimelineMetrics{
		OriginalEstimate: epic.Duration,
	}

	if epic.StartDate != nil {
		timeline.DaysActive = int(time.Since(*epic.StartDate).Hours() / 24)
		
		// Estimate remaining days based on velocity
		if velocity.StoriesPerDay > 0 && progress.TotalStories > 0 {
			remainingStories := progress.TotalStories - progress.StoriesCompleted
			timeline.EstimatedDaysRemaining = int(float64(remainingStories) / velocity.StoriesPerDay)
		}

		// Check if overdue (simplified logic)
		if epic.Duration != "" && strings.Contains(epic.Duration, "week") {
			var estimatedWeeks int
			fmt.Sscanf(epic.Duration, "%d", &estimatedWeeks)
			estimatedDays := estimatedWeeks * 7
			
			if timeline.DaysActive > estimatedDays {
				timeline.IsOverdue = true
				timeline.DaysOverdue = timeline.DaysActive - estimatedDays
			}
		}
	}

	return timeline
}

// Helper methods for display

func (d *Dashboard) createProgressBar(percentage float64, width int) string {
	filled := int(percentage / 100.0 * float64(width))
	empty := width - filled
	
	bar := "["
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	bar += "]"
	
	return bar
}

func (d *Dashboard) getStatusIcon(status Status) string {
	switch status {
	case StatusPlanned:
		return "📋"
	case StatusInProgress:
		return "🚧"
	case StatusOnHold:
		return "⏸️"
	case StatusCompleted:
		return "✅"
	case StatusCancelled:
		return "❌"
	default:
		return "❓"
	}
}

func (d *Dashboard) getPriorityIcon(priority Priority) string {
	switch priority {
	case PriorityLow:
		return "🟢"
	case PriorityMedium:
		return "🟡"
	case PriorityHigh:
		return "🟠"
	case PriorityCritical:
		return "🔴"
	default:
		return "⚪"
	}
}

func (d *Dashboard) getRiskIcon(risk RiskLevel) string {
	switch risk {
	case RiskLow:
		return "🟢"
	case RiskMedium:
		return "🟡"
	case RiskHigh:
		return "🟠"
	case RiskCritical:
		return "🔴"
	default:
		return "⚪"
	}
}

// Helper functions

func percentage(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0.0
	}
	return float64(numerator) / float64(denominator) * 100.0
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}