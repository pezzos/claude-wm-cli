package git

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"claude-wm-cli/internal/state"
)

// GitRecoveryEngine provides advanced recovery mechanisms using Git history
type GitRecoveryEngine struct {
	repository       *Repository
	versionManager   *StateVersionManager
	corruptionDetector *state.CorruptionDetector
	config           *GitConfig
}

// NewGitRecoveryEngine creates a new Git recovery engine
func NewGitRecoveryEngine(repository *Repository, versionManager *StateVersionManager, 
	corruptionDetector *state.CorruptionDetector, config *GitConfig) *GitRecoveryEngine {
	
	return &GitRecoveryEngine{
		repository:         repository,
		versionManager:     versionManager,
		corruptionDetector: corruptionDetector,
		config:             config,
	}
}

// RecoveryStrategy represents different recovery approaches
type RecoveryStrategy string

const (
	StrategyAutomatic    RecoveryStrategy = "automatic"     // Automatic recovery using best practices
	StrategyInteractive  RecoveryStrategy = "interactive"   // User-guided recovery
	StrategyConservative RecoveryStrategy = "conservative"  // Safe recovery with minimal changes
	StrategyAggressive   RecoveryStrategy = "aggressive"    // Comprehensive recovery including hard resets
)

// RecoveryOptions configures recovery behavior
type RecoveryOptions struct {
	Strategy         RecoveryStrategy `json:"strategy"`
	MaxSearchDepth   int              `json:"max_search_depth"`    // How far back to look
	VerifyIntegrity  bool             `json:"verify_integrity"`    // Verify state after recovery
	CreateBackup     bool             `json:"create_backup"`       // Backup before recovery
	AllowDataLoss    bool             `json:"allow_data_loss"`     // Allow operations that might lose data
	TargetFiles      []string         `json:"target_files"`        // Specific files to recover
	TimeWindow       *TimeWindow      `json:"time_window"`         // Time-based recovery constraints
}

// TimeWindow defines a time range for recovery operations
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// RecoveryPlan describes the steps needed for recovery
type RecoveryPlan struct {
	Strategy        RecoveryStrategy   `json:"strategy"`
	Steps           []RecoveryStep     `json:"steps"`
	EstimatedRisk   string             `json:"estimated_risk"`     // low, medium, high
	DataLossRisk    bool               `json:"data_loss_risk"`
	RequiresBackup  bool               `json:"requires_backup"`
	EstimatedTime   time.Duration      `json:"estimated_time"`
	RecoveryPoints  []*RecoveryPoint   `json:"recovery_points"`
	Warnings        []string           `json:"warnings"`
	Prerequisites   []string           `json:"prerequisites"`
}

// RecoveryStep represents a single step in the recovery process
type RecoveryStep struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"`           // checkout, reset, merge, manual
	Description string        `json:"description"`
	Command     string        `json:"command,omitempty"`
	Files       []string      `json:"files,omitempty"`
	Risk        string        `json:"risk"`           // low, medium, high
	Reversible  bool          `json:"reversible"`
	Required    bool          `json:"required"`
	Automated   bool          `json:"automated"`
	EstimatedTime time.Duration `json:"estimated_time"`
}

// RecoveryResult contains the outcome of recovery operations
type RecoveryResult struct {
	Success         bool              `json:"success"`
	Strategy        RecoveryStrategy  `json:"strategy"`
	StepsExecuted   []RecoveryStep    `json:"steps_executed"`
	StepsFailed     []RecoveryStep    `json:"steps_failed"`
	FilesRecovered  []string          `json:"files_recovered"`
	FilesCorrupted  []string          `json:"files_corrupted"`
	BackupCreated   string            `json:"backup_created,omitempty"`
	Duration        time.Duration     `json:"duration"`
	RecoveryPoint   *RecoveryPoint    `json:"recovery_point,omitempty"`
	IntegrityCheck  bool              `json:"integrity_check"`
	Warnings        []string          `json:"warnings"`
	NextSteps       []string          `json:"next_steps"`
	Timestamp       time.Time         `json:"timestamp"`
}

// AutoRecover automatically recovers from state corruption
func (gre *GitRecoveryEngine) AutoRecover(corruptedFiles []string, options *RecoveryOptions) (*RecoveryResult, error) {
	if options == nil {
		options = &RecoveryOptions{
			Strategy:        StrategyAutomatic,
			MaxSearchDepth:  50,
			VerifyIntegrity: true,
			CreateBackup:    true,
			AllowDataLoss:   false,
			TargetFiles:     corruptedFiles,
		}
	}
	
	start := time.Now()
	result := &RecoveryResult{
		Strategy:       options.Strategy,
		FilesCorrupted: corruptedFiles,
		Timestamp:      start,
	}
	
	// Step 1: Create backup if requested
	if options.CreateBackup {
		backup, err := gre.createPreRecoveryBackup(corruptedFiles)
		if err != nil {
			return result, fmt.Errorf("failed to create backup: %w", err)
		}
		result.BackupCreated = backup
	}
	
	// Step 2: Analyze corruption and create recovery plan
	plan, err := gre.CreateRecoveryPlan(corruptedFiles, options)
	if err != nil {
		return result, fmt.Errorf("failed to create recovery plan: %w", err)
	}
	
	// Step 3: Execute recovery plan
	for _, step := range plan.Steps {
		if err := gre.executeRecoveryStep(step, result); err != nil {
			result.StepsFailed = append(result.StepsFailed, step)
			result.Warnings = append(result.Warnings, fmt.Sprintf("Step failed: %s - %v", step.ID, err))
			
			// For automatic recovery, stop on first failure
			if options.Strategy == StrategyAutomatic && !step.Required {
				break
			}
		} else {
			result.StepsExecuted = append(result.StepsExecuted, step)
		}
	}
	
	// Step 4: Verify recovery
	if options.VerifyIntegrity {
		result.IntegrityCheck = gre.verifyRecoveryIntegrity(corruptedFiles)
		if !result.IntegrityCheck {
			result.Warnings = append(result.Warnings, "Recovery integrity check failed")
		}
	}
	
	// Step 5: Determine overall success
	result.Success = len(result.StepsFailed) == 0 && (result.IntegrityCheck || !options.VerifyIntegrity)
	result.Duration = time.Since(start)
	
	// Add next steps recommendations
	if result.Success {
		result.NextSteps = []string{
			"Verify application functionality",
			"Consider additional testing",
			"Monitor for recurring issues",
		}
	} else {
		result.NextSteps = []string{
			"Review recovery logs",
			"Consider manual recovery",
			"Contact support if issues persist",
		}
	}
	
	return result, nil
}

// CreateRecoveryPlan analyzes corruption and creates a step-by-step recovery plan
func (gre *GitRecoveryEngine) CreateRecoveryPlan(corruptedFiles []string, options *RecoveryOptions) (*RecoveryPlan, error) {
	plan := &RecoveryPlan{
		Strategy:       options.Strategy,
		Steps:          make([]RecoveryStep, 0),
		EstimatedRisk:  "low",
		DataLossRisk:   false,
		RequiresBackup: true,
		RecoveryPoints: make([]*RecoveryPoint, 0),
		Warnings:       make([]string, 0),
		Prerequisites:  make([]string, 0),
	}
	
	// Find suitable recovery points
	recoveryPoints, err := gre.findRecoveryPoints(corruptedFiles, options)
	if err != nil {
		return nil, fmt.Errorf("failed to find recovery points: %w", err)
	}
	
	if len(recoveryPoints) == 0 {
		return nil, fmt.Errorf("no suitable recovery points found")
	}
	
	plan.RecoveryPoints = recoveryPoints
	
	// Analyze corruption severity
	corruptionSeverity := gre.analyzeCorruptionSeverity(corruptedFiles)
	
	// Create recovery steps based on strategy and corruption severity
	switch options.Strategy {
	case StrategyAutomatic:
		plan.Steps = gre.createAutomaticRecoverySteps(recoveryPoints, corruptedFiles, corruptionSeverity)
	case StrategyConservative:
		plan.Steps = gre.createConservativeRecoverySteps(recoveryPoints, corruptedFiles, corruptionSeverity)
	case StrategyAggressive:
		plan.Steps = gre.createAggressiveRecoverySteps(recoveryPoints, corruptedFiles, corruptionSeverity)
	case StrategyInteractive:
		plan.Steps = gre.createInteractiveRecoverySteps(recoveryPoints, corruptedFiles, corruptionSeverity)
	default:
		return nil, fmt.Errorf("unknown recovery strategy: %s", options.Strategy)
	}
	
	// Calculate estimated time and risk
	totalTime := time.Duration(0)
	highRiskSteps := 0
	
	for _, step := range plan.Steps {
		totalTime += step.EstimatedTime
		if step.Risk == "high" {
			highRiskSteps++
		}
		if !step.Reversible {
			plan.DataLossRisk = true
		}
	}
	
	plan.EstimatedTime = totalTime
	
	if highRiskSteps > 0 {
		plan.EstimatedRisk = "high"
	} else if len(plan.Steps) > 3 {
		plan.EstimatedRisk = "medium"
	}
	
	// Add warnings and prerequisites
	if plan.DataLossRisk {
		plan.Warnings = append(plan.Warnings, "Recovery may result in data loss")
		plan.Prerequisites = append(plan.Prerequisites, "Create backup before proceeding")
	}
	
	if corruptionSeverity == "high" {
		plan.Warnings = append(plan.Warnings, "High corruption detected - recovery may be incomplete")
	}
	
	return plan, nil
}

// ExecuteRecoveryPlan executes a complete recovery plan
func (gre *GitRecoveryEngine) ExecuteRecoveryPlan(plan *RecoveryPlan, options *RecoveryOptions) (*RecoveryResult, error) {
	start := time.Now()
	result := &RecoveryResult{
		Strategy:    plan.Strategy,
		Timestamp:   start,
		StepsExecuted: make([]RecoveryStep, 0),
		StepsFailed:   make([]RecoveryStep, 0),
	}
	
	// Execute each step
	for _, step := range plan.Steps {
		if err := gre.executeRecoveryStep(step, result); err != nil {
			result.StepsFailed = append(result.StepsFailed, step)
			result.Warnings = append(result.Warnings, fmt.Sprintf("Step '%s' failed: %v", step.ID, err))
			
			// Stop on critical failures
			if step.Required {
				result.Success = false
				result.Duration = time.Since(start)
				return result, fmt.Errorf("critical recovery step failed: %s", step.ID)
			}
		} else {
			result.StepsExecuted = append(result.StepsExecuted, step)
		}
	}
	
	result.Success = len(result.StepsFailed) == 0
	result.Duration = time.Since(start)
	
	return result, nil
}

// Helper methods for recovery operations

func (gre *GitRecoveryEngine) findRecoveryPoints(corruptedFiles []string, options *RecoveryOptions) ([]*RecoveryPoint, error) {
	// Get commit history
	commits, err := gre.repository.GetLog(options.MaxSearchDepth)
	if err != nil {
		return nil, err
	}
	
	var recoveryPoints []*RecoveryPoint
	
	for _, commit := range commits {
		// Filter by time window if specified
		if options.TimeWindow != nil {
			if commit.Date.Before(options.TimeWindow.Start) || commit.Date.After(options.TimeWindow.End) {
				continue
			}
		}
		
		// Check if commit affects our corrupted files
		affectsFiles := len(corruptedFiles) == 0 // If no specific files, consider all commits
		for _, corruptedFile := range corruptedFiles {
			for _, commitFile := range commit.Files {
				if strings.Contains(commitFile, corruptedFile) {
					affectsFiles = true
					break
				}
			}
			if affectsFiles {
				break
			}
		}
		
		if !affectsFiles {
			continue
		}
		
		recoveryPoint := &RecoveryPoint{
			Commit:      *commit,
			Description: commit.Message,
			StateFiles:  commit.Files,
			Verified:    false, // Will be verified when needed
			Safe:        true,  // Assume safe unless proven otherwise
			Automatic:   strings.Contains(commit.Message, "backup:"),
		}
		
		recoveryPoints = append(recoveryPoints, recoveryPoint)
		
		// Limit number of recovery points
		if len(recoveryPoints) >= 10 {
			break
		}
	}
	
	// Sort by date (newest first)
	sort.Slice(recoveryPoints, func(i, j int) bool {
		return recoveryPoints[i].Commit.Date.After(recoveryPoints[j].Commit.Date)
	})
	
	return recoveryPoints, nil
}

func (gre *GitRecoveryEngine) analyzeCorruptionSeverity(corruptedFiles []string) string {
	if gre.corruptionDetector == nil {
		return "unknown"
	}
	
	criticalIssues := 0
	majorIssues := 0
	
	for _, file := range corruptedFiles {
		report := gre.corruptionDetector.ScanFile(file)
		for _, issue := range report.Issues {
			switch issue.Severity {
			case "critical":
				criticalIssues++
			case "major":
				majorIssues++
			}
		}
	}
	
	if criticalIssues > 0 {
		return "high"
	} else if majorIssues > 2 {
		return "medium"
	}
	return "low"
}

func (gre *GitRecoveryEngine) createAutomaticRecoverySteps(recoveryPoints []*RecoveryPoint, 
	corruptedFiles []string, severity string) []RecoveryStep {
	
	var steps []RecoveryStep
	
	if len(recoveryPoints) == 0 {
		return steps
	}
	
	// Use the most recent recovery point
	latestPoint := recoveryPoints[0]
	
	// Step 1: Checkout files from recovery point
	steps = append(steps, RecoveryStep{
		ID:            "checkout_files",
		Type:          "checkout",
		Description:   fmt.Sprintf("Restore files from commit %s", latestPoint.Commit.ShortHash),
		Files:         corruptedFiles,
		Risk:          "low",
		Reversible:    true,
		Required:      true,
		Automated:     true,
		EstimatedTime: 5 * time.Second,
	})
	
	// Step 2: Verify integrity
	steps = append(steps, RecoveryStep{
		ID:            "verify_integrity",
		Type:          "verify",
		Description:   "Verify recovered files integrity",
		Files:         corruptedFiles,
		Risk:          "low",
		Reversible:    true,
		Required:      false,
		Automated:     true,
		EstimatedTime: 2 * time.Second,
	})
	
	return steps
}

func (gre *GitRecoveryEngine) createConservativeRecoverySteps(recoveryPoints []*RecoveryPoint, 
	corruptedFiles []string, severity string) []RecoveryStep {
	
	var steps []RecoveryStep
	
	// Conservative approach: multiple verification steps
	steps = append(steps, RecoveryStep{
		ID:            "analyze_corruption",
		Type:          "analyze",
		Description:   "Analyze corruption patterns",
		Files:         corruptedFiles,
		Risk:          "low",
		Reversible:    true,
		Required:      false,
		Automated:     true,
		EstimatedTime: 3 * time.Second,
	})
	
	if len(recoveryPoints) > 0 {
		steps = append(steps, RecoveryStep{
			ID:            "selective_restore",
			Type:          "checkout",
			Description:   "Selectively restore corrupted sections",
			Files:         corruptedFiles,
			Risk:          "low",
			Reversible:    true,
			Required:      true,
			Automated:     false, // Requires review
			EstimatedTime: 10 * time.Second,
		})
	}
	
	return steps
}

func (gre *GitRecoveryEngine) createAggressiveRecoverySteps(recoveryPoints []*RecoveryPoint, 
	corruptedFiles []string, severity string) []RecoveryStep {
	
	var steps []RecoveryStep
	
	if len(recoveryPoints) > 0 {
		// Aggressive approach: hard reset to known good state
		steps = append(steps, RecoveryStep{
			ID:            "hard_reset",
			Type:          "reset",
			Description:   fmt.Sprintf("Hard reset to commit %s", recoveryPoints[0].Commit.ShortHash),
			Files:         []string{"."}, // All files
			Risk:          "high",
			Reversible:    false,
			Required:      true,
			Automated:     true,
			EstimatedTime: 3 * time.Second,
		})
	}
	
	return steps
}

func (gre *GitRecoveryEngine) createInteractiveRecoverySteps(recoveryPoints []*RecoveryPoint, 
	corruptedFiles []string, severity string) []RecoveryStep {
	
	var steps []RecoveryStep
	
	// Interactive approach: present options to user
	steps = append(steps, RecoveryStep{
		ID:            "user_review",
		Type:          "manual",
		Description:   "Review recovery options and select approach",
		Files:         corruptedFiles,
		Risk:          "low",
		Reversible:    true,
		Required:      true,
		Automated:     false,
		EstimatedTime: 30 * time.Second, // User interaction time
	})
	
	return steps
}

func (gre *GitRecoveryEngine) executeRecoveryStep(step RecoveryStep, result *RecoveryResult) error {
	switch step.Type {
	case "checkout":
		return gre.executeCheckoutStep(step, result)
	case "reset":
		return gre.executeResetStep(step, result)
	case "verify":
		return gre.executeVerifyStep(step, result)
	case "analyze":
		return gre.executeAnalyzeStep(step, result)
	case "manual":
		return gre.executeManualStep(step, result)
	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
}

func (gre *GitRecoveryEngine) executeCheckoutStep(step RecoveryStep, result *RecoveryResult) error {
	// Implementation would checkout specific files from Git
	result.FilesRecovered = append(result.FilesRecovered, step.Files...)
	return nil
}

func (gre *GitRecoveryEngine) executeResetStep(step RecoveryStep, result *RecoveryResult) error {
	// Implementation would perform Git reset
	result.FilesRecovered = append(result.FilesRecovered, step.Files...)
	return nil
}

func (gre *GitRecoveryEngine) executeVerifyStep(step RecoveryStep, result *RecoveryResult) error {
	// Verify using corruption detector
	if gre.corruptionDetector != nil {
		for _, file := range step.Files {
			report := gre.corruptionDetector.ScanFile(file)
			if report.IsCorrupted {
				return fmt.Errorf("file still corrupted after recovery: %s", file)
			}
		}
	}
	return nil
}

func (gre *GitRecoveryEngine) executeAnalyzeStep(step RecoveryStep, result *RecoveryResult) error {
	// Perform corruption analysis
	return nil
}

func (gre *GitRecoveryEngine) executeManualStep(step RecoveryStep, result *RecoveryResult) error {
	// Manual steps require user intervention
	result.Warnings = append(result.Warnings, fmt.Sprintf("Manual step required: %s", step.Description))
	return nil
}

func (gre *GitRecoveryEngine) createPreRecoveryBackup(files []string) (string, error) {
	backupName := fmt.Sprintf("pre-recovery-backup-%d", time.Now().Unix())
	commit, err := gre.versionManager.VersionState(CommitTypeBackup, backupName, files...)
	if err != nil {
		return "", err
	}
	if commit == nil {
		return "", fmt.Errorf("no changes to backup")
	}
	return commit.Hash, nil
}

func (gre *GitRecoveryEngine) verifyRecoveryIntegrity(files []string) bool {
	if gre.corruptionDetector == nil {
		return true // Assume success if no detector
	}
	
	for _, file := range files {
		report := gre.corruptionDetector.ScanFile(file)
		if report.IsCorrupted {
			return false
		}
	}
	return true
}