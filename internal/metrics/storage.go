package metrics

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// MetricEntry represents a single performance metric
type MetricEntry struct {
	ID          int64     `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	ProjectPath string    `json:"project_path"` // Hashed
	ProjectName string    `json:"project_name"`
	CommandName string    `json:"command_name"`
	StepName    string    `json:"step_name,omitempty"`
	DurationMs  int64     `json:"duration_ms"`
	ContextData string    `json:"context_data,omitempty"`
	ToolVersion string    `json:"tool_version"`
	ExitCode    int       `json:"exit_code"`
}

// Storage handles SQLite operations for metrics
type Storage struct {
	db       *sql.DB
	dbPath   string
	initOnce bool
}

// NewStorage creates a new storage instance
func NewStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	metricsDir := filepath.Join(homeDir, ".claude-wm", "metrics")
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metrics directory: %w", err)
	}
	
	dbPath := filepath.Join(metricsDir, "performance.db")
	
	db, err := sql.Open("sqlite3", dbPath+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	storage := &Storage{
		db:     db,
		dbPath: dbPath,
	}
	
	if err := storage.initialize(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	
	return storage, nil
}

// initialize creates the database schema
func (s *Storage) initialize() error {
	if s.initOnce {
		return nil
	}
	
	schema := `
	CREATE TABLE IF NOT EXISTS performance_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		project_path TEXT NOT NULL,
		project_name TEXT NOT NULL,
		command_name TEXT NOT NULL,
		step_name TEXT,
		duration_ms INTEGER NOT NULL,
		context_data TEXT,
		tool_version TEXT NOT NULL,
		exit_code INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_command_step ON performance_metrics(command_name, step_name);
	CREATE INDEX IF NOT EXISTS idx_project ON performance_metrics(project_name);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON performance_metrics(timestamp);
	CREATE INDEX IF NOT EXISTS idx_duration ON performance_metrics(duration_ms);
	`
	
	_, err := s.db.Exec(schema)
	if err != nil {
		return err
	}
	
	s.initOnce = true
	return nil
}

// SaveMetric saves a single metric entry
func (s *Storage) SaveMetric(entry MetricEntry) error {
	query := `
	INSERT INTO performance_metrics (
		timestamp, project_path, project_name, command_name, step_name,
		duration_ms, context_data, tool_version, exit_code
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := s.db.Exec(query,
		entry.Timestamp,
		entry.ProjectPath,
		entry.ProjectName,
		entry.CommandName,
		entry.StepName,
		entry.DurationMs,
		entry.ContextData,
		entry.ToolVersion,
		entry.ExitCode,
	)
	
	return err
}

// GetCommandStats returns statistics for a specific command
func (s *Storage) GetCommandStats(commandName string, days int) (*CommandStats, error) {
	var query string
	var args []interface{}
	
	if commandName == "" {
		// Query for all commands combined
		query = `
		SELECT 
			COUNT(*) as count,
			MIN(duration_ms) as min_duration,
			AVG(duration_ms) as avg_duration,
			MAX(duration_ms) as max_duration
		FROM performance_metrics 
		WHERE step_name = ''
			AND timestamp >= datetime('now', '-' || ? || ' days')
		`
		args = []interface{}{days}
	} else {
		// Query for specific command
		query = `
		SELECT 
			COUNT(*) as count,
			MIN(duration_ms) as min_duration,
			AVG(duration_ms) as avg_duration,
			MAX(duration_ms) as max_duration
		FROM performance_metrics 
		WHERE command_name = ? 
			AND step_name = ''
			AND timestamp >= datetime('now', '-' || ? || ' days')
		`
		args = []interface{}{commandName, days}
	}
	
	var stats CommandStats
	var minDuration, avgDuration, maxDuration sql.NullFloat64
	
	err := s.db.QueryRow(query, args...).Scan(
		&stats.Count,
		&minDuration,
		&avgDuration,
		&maxDuration,
	)
	
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	
	// Set values from nullable fields
	if minDuration.Valid {
		stats.MinDuration = minDuration.Float64
	}
	if avgDuration.Valid {
		stats.AvgDuration = avgDuration.Float64
	}
	if maxDuration.Valid {
		stats.MaxDuration = maxDuration.Float64
	}
	
	// Calculate P95 separately (SQLite doesn't support PERCENTILE_CONT)
	if stats.Count > 0 {
		p95Query := `
		SELECT duration_ms FROM performance_metrics 
		WHERE command_name = ? 
			AND step_name = ''
			AND timestamp >= datetime('now', '-' || ? || ' days')
		ORDER BY duration_ms 
		LIMIT 1 OFFSET (SELECT COUNT(*) * 95 / 100 FROM performance_metrics 
			WHERE command_name = ? 
				AND step_name = ''
				AND timestamp >= datetime('now', '-' || ? || ' days'))
		`
		
		var p95Duration sql.NullFloat64
		if commandName != "" {
			s.db.QueryRow(p95Query, commandName, days, commandName, days).Scan(&p95Duration)
		}
		
		if p95Duration.Valid {
			stats.P95Duration = p95Duration.Float64
		}
	}
	
	stats.CommandName = commandName
	return &stats, nil
}

// GetStepStats returns statistics for steps within a command
func (s *Storage) GetStepStats(commandName string, days int) ([]StepStats, error) {
	query := `
	SELECT 
		step_name,
		COUNT(*) as count,
		MIN(duration_ms) as min_duration,
		AVG(duration_ms) as avg_duration,
		MAX(duration_ms) as max_duration
	FROM performance_metrics 
	WHERE command_name = ? 
		AND step_name != ''
		AND timestamp >= datetime('now', '-' || ? || ' days')
	GROUP BY step_name
	ORDER BY avg_duration DESC
	`
	
	rows, err := s.db.Query(query, commandName, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stats []StepStats
	for rows.Next() {
		var step StepStats
		err := rows.Scan(
			&step.StepName,
			&step.Count,
			&step.MinDuration,
			&step.AvgDuration,
			&step.MaxDuration,
		)
		if err != nil {
			return nil, err
		}
		step.CommandName = commandName
		stats = append(stats, step)
	}
	
	return stats, nil
}

// GetSlowCommands returns the slowest commands
func (s *Storage) GetSlowCommands(thresholdMs int64, days int) ([]CommandStats, error) {
	query := `
	SELECT 
		command_name,
		COUNT(*) as count,
		MIN(duration_ms) as min_duration,
		AVG(duration_ms) as avg_duration,
		MAX(duration_ms) as max_duration
	FROM performance_metrics 
	WHERE step_name = ''
		AND avg_duration > ?
		AND timestamp >= datetime('now', '-' || ? || ' days')
	GROUP BY command_name
	ORDER BY avg_duration DESC
	`
	
	rows, err := s.db.Query(query, thresholdMs, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var commands []CommandStats
	for rows.Next() {
		var cmd CommandStats
		err := rows.Scan(
			&cmd.CommandName,
			&cmd.Count,
			&cmd.MinDuration,
			&cmd.AvgDuration,
			&cmd.MaxDuration,
		)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}
	
	return commands, nil
}

// GetProjectComparison returns performance comparison across projects
func (s *Storage) GetProjectComparison(days int) ([]ProjectStats, error) {
	query := `
	SELECT 
		project_name,
		COUNT(*) as total_commands,
		AVG(duration_ms) as avg_duration,
		MAX(duration_ms) as max_duration
	FROM performance_metrics 
	WHERE step_name = ''
		AND timestamp >= datetime('now', '-' || ? || ' days')
	GROUP BY project_name
	ORDER BY avg_duration DESC
	`
	
	rows, err := s.db.Query(query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var projects []ProjectStats
	for rows.Next() {
		var project ProjectStats
		err := rows.Scan(
			&project.ProjectName,
			&project.TotalCommands,
			&project.AvgDuration,
			&project.MaxDuration,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	
	return projects, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Stats structures
type CommandStats struct {
	CommandName string  `json:"command_name"`
	Count       int     `json:"count"`
	MinDuration float64 `json:"min_duration_ms"`
	AvgDuration float64 `json:"avg_duration_ms"`
	MaxDuration float64 `json:"max_duration_ms"`
	P95Duration float64 `json:"p95_duration_ms"`
}

type StepStats struct {
	CommandName string  `json:"command_name"`
	StepName    string  `json:"step_name"`
	Count       int     `json:"count"`
	MinDuration float64 `json:"min_duration_ms"`
	AvgDuration float64 `json:"avg_duration_ms"`
	MaxDuration float64 `json:"max_duration_ms"`
}

type ProjectStats struct {
	ProjectName   string  `json:"project_name"`
	TotalCommands int     `json:"total_commands"`
	AvgDuration   float64 `json:"avg_duration_ms"`
	MaxDuration   float64 `json:"max_duration_ms"`
}

// GetAllCommandStats returns statistics for all commands
func (s *Storage) GetAllCommandStats(days int) ([]CommandStats, error) {
	query := `
	SELECT 
		command_name,
		COUNT(*) as count,
		MIN(duration_ms) as min_duration,
		AVG(duration_ms) as avg_duration,
		MAX(duration_ms) as max_duration
	FROM performance_metrics 
	WHERE step_name = ''
		AND timestamp >= datetime('now', '-' || ? || ' days')
	GROUP BY command_name
	ORDER BY avg_duration DESC
	`
	
	rows, err := s.db.Query(query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var commands []CommandStats
	for rows.Next() {
		var cmd CommandStats
		err := rows.Scan(
			&cmd.CommandName,
			&cmd.Count,
			&cmd.MinDuration,
			&cmd.AvgDuration,
			&cmd.MaxDuration,
		)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}
	
	return commands, nil
}

// hashProjectPath creates a consistent hash of the project path for anonymization
func hashProjectPath(path string) string {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars
}