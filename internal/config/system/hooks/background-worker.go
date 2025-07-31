package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Job represents a background hook job
type Job struct {
	ID          int64     `json:"id"`
	HookName    string    `json:"hook_name"`
	Args        string    `json:"args"`
	Priority    int       `json:"priority"`
	Status      string    `json:"status"` // pending, running, completed, failed
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Output      string    `json:"output,omitempty"`
	Error       string    `json:"error,omitempty"`
	Retries     int       `json:"retries"`
	MaxRetries  int       `json:"max_retries"`
}

// BackgroundWorker manages background hook execution
type BackgroundWorker struct {
	db           *sql.DB
	hooksDir     string
	maxWorkers   int
	pollInterval time.Duration
	jobs         chan Job
	shutdown     chan struct{}
	done         chan struct{}
}

// NewBackgroundWorker creates a new background worker
func NewBackgroundWorker(dbPath, hooksDir string, maxWorkers int) (*BackgroundWorker, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	worker := &BackgroundWorker{
		db:           db,
		hooksDir:     hooksDir,
		maxWorkers:   maxWorkers,
		pollInterval: 1 * time.Second,
		jobs:         make(chan Job, maxWorkers*2),
		shutdown:     make(chan struct{}),
		done:         make(chan struct{}),
	}

	if err := worker.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return worker, nil
}

// initDatabase creates the jobs table if it doesn't exist
func (bw *BackgroundWorker) initDatabase() error {
	query := `
	CREATE TABLE IF NOT EXISTS jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hook_name TEXT NOT NULL,
		args TEXT,
		priority INTEGER DEFAULT 5,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		started_at DATETIME,
		completed_at DATETIME,
		output TEXT,
		error TEXT,
		retries INTEGER DEFAULT 0,
		max_retries INTEGER DEFAULT 3
	);
	
	CREATE INDEX IF NOT EXISTS idx_status_priority ON jobs(status, priority);
	CREATE INDEX IF NOT EXISTS idx_created_at ON jobs(created_at);
	`

	_, err := bw.db.Exec(query)
	return err
}

// EnqueueJob adds a new job to the background queue
func (bw *BackgroundWorker) EnqueueJob(hookName, args string, priority int) (int64, error) {
	query := `
	INSERT INTO jobs (hook_name, args, priority, max_retries)
	VALUES (?, ?, ?, ?)
	`

	// Set max retries based on hook type
	maxRetries := 3
	if strings.Contains(hookName, "log-") {
		maxRetries = 1 // Logging hooks don't need many retries
	}

	result, err := bw.db.Exec(query, hookName, args, priority, maxRetries)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Printf("🔄 Enqueued background job: %s (ID: %d, Priority: %d)", hookName, id, priority)
	return id, nil
}

// Start starts the background worker
func (bw *BackgroundWorker) Start(ctx context.Context) {
	log.Printf("🚀 Background Worker starting with %d workers", bw.maxWorkers)

	// Start worker goroutines
	for i := 0; i < bw.maxWorkers; i++ {
		go bw.worker(ctx, i+1)
	}

	// Start job poller
	go bw.pollJobs(ctx)

	// Start cleanup routine
	go bw.cleanupOldJobs(ctx)

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Printf("🛑 Received shutdown signal")
		bw.Shutdown()
	case <-ctx.Done():
		log.Printf("🛑 Context cancelled")
		bw.Shutdown()
	case <-bw.done:
		log.Printf("✅ Background worker finished")
	}
}

// worker processes jobs from the queue
func (bw *BackgroundWorker) worker(ctx context.Context, workerID int) {
	log.Printf("👷 Worker %d started", workerID)

	for {
		select {
		case job := <-bw.jobs:
			bw.processJob(ctx, job, workerID)
		case <-bw.shutdown:
			log.Printf("👷 Worker %d shutting down", workerID)
			return
		case <-ctx.Done():
			return
		}
	}
}

// pollJobs polls the database for pending jobs
func (bw *BackgroundWorker) pollJobs(ctx context.Context) {
	ticker := time.NewTicker(bw.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			jobs, err := bw.getPendingJobs()
			if err != nil {
				log.Printf("❌ Error fetching pending jobs: %v", err)
				continue
			}

			for _, job := range jobs {
				select {
				case bw.jobs <- job:
					// Job queued successfully
				default:
					// Queue is full, will be picked up next time
				}
			}

		case <-bw.shutdown:
			return
		case <-ctx.Done():
			return
		}
	}
}

// getPendingJobs retrieves pending jobs from the database
func (bw *BackgroundWorker) getPendingJobs() ([]Job, error) {
	query := `
	SELECT id, hook_name, args, priority, status, created_at, retries, max_retries
	FROM jobs 
	WHERE status = 'pending' AND retries < max_retries
	ORDER BY priority ASC, created_at ASC
	LIMIT ?
	`

	rows, err := bw.db.Query(query, bw.maxWorkers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		err := rows.Scan(
			&job.ID, &job.HookName, &job.Args, &job.Priority,
			&job.Status, &job.CreatedAt, &job.Retries, &job.MaxRetries,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// processJob executes a single background job
func (bw *BackgroundWorker) processJob(ctx context.Context, job Job, workerID int) {
	startTime := time.Now()
	log.Printf("👷 Worker %d processing job %d: %s", workerID, job.ID, job.HookName)

	// Mark job as running
	if err := bw.updateJobStatus(job.ID, "running", &startTime, nil); err != nil {
		log.Printf("❌ Failed to update job status: %v", err)
		return
	}

	// Execute the hook
	hookPath := filepath.Join(bw.hooksDir, job.HookName)
	
	// Create command with timeout
	jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if strings.HasSuffix(job.HookName, ".py") {
		cmd = exec.CommandContext(jobCtx, "python3", hookPath)
	} else if strings.HasSuffix(job.HookName, ".sh") {
		cmd = exec.CommandContext(jobCtx, "bash", hookPath)
	} else {
		cmd = exec.CommandContext(jobCtx, hookPath)
	}

	// Set environment variables
	cmd.Env = append(os.Environ(),
		"BACKGROUND_MODE=true",
		fmt.Sprintf("WORKER_ID=%d", workerID),
		fmt.Sprintf("JOB_ID=%d", job.ID),
	)

	// Pass args via stdin if provided
	if job.Args != "" {
		cmd.Stdin = strings.NewReader(job.Args)
	}

	// Execute command
	output, err := cmd.CombinedOutput()
	completedAt := time.Now()

	if err != nil {
		// Job failed, increment retries
		log.Printf("❌ Job %d failed: %v", job.ID, err)
		if err := bw.handleJobFailure(job.ID, string(output), err.Error(), job.Retries+1); err != nil {
			log.Printf("❌ Failed to handle job failure: %v", err)
		}
	} else {
		// Job succeeded
		log.Printf("✅ Job %d completed successfully in %v", job.ID, completedAt.Sub(startTime))
		if err := bw.updateJobStatus(job.ID, "completed", nil, &completedAt); err != nil {
			log.Printf("❌ Failed to update job status: %v", err)
		}
		if err := bw.updateJobOutput(job.ID, string(output), ""); err != nil {
			log.Printf("❌ Failed to update job output: %v", err)
		}
	}
}

// updateJobStatus updates the status and timestamps of a job
func (bw *BackgroundWorker) updateJobStatus(jobID int64, status string, startedAt, completedAt *time.Time) error {
	query := `UPDATE jobs SET status = ?`
	args := []interface{}{status}

	if startedAt != nil {
		query += `, started_at = ?`
		args = append(args, startedAt)
	}

	if completedAt != nil {
		query += `, completed_at = ?`
		args = append(args, completedAt)
	}

	query += ` WHERE id = ?`
	args = append(args, jobID)

	_, err := bw.db.Exec(query, args...)
	return err
}

// updateJobOutput updates the output and error of a job
func (bw *BackgroundWorker) updateJobOutput(jobID int64, output, errorMsg string) error {
	query := `UPDATE jobs SET output = ?, error = ? WHERE id = ?`
	_, err := bw.db.Exec(query, output, errorMsg, jobID)
	return err
}

// handleJobFailure handles a failed job
func (bw *BackgroundWorker) handleJobFailure(jobID int64, output, errorMsg string, retries int) error {
	query := `UPDATE jobs SET status = ?, retries = ?, output = ?, error = ? WHERE id = ?`
	
	status := "pending" // Retry
	if retries >= 3 {    // Max retries reached
		status = "failed"
	}

	_, err := bw.db.Exec(query, status, retries, output, errorMsg, jobID)
	return err
}

// cleanupOldJobs removes old completed/failed jobs
func (bw *BackgroundWorker) cleanupOldJobs(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cutoff := time.Now().Add(-24 * time.Hour) // Keep jobs for 24 hours
			query := `DELETE FROM jobs WHERE status IN ('completed', 'failed') AND completed_at < ?`
			
			result, err := bw.db.Exec(query, cutoff)
			if err != nil {
				log.Printf("❌ Error cleaning up old jobs: %v", err)
				continue
			}

			if affected, err := result.RowsAffected(); err == nil && affected > 0 {
				log.Printf("🧹 Cleaned up %d old jobs", affected)
			}

		case <-bw.shutdown:
			return
		case <-ctx.Done():
			return
		}
	}
}

// GetJobStats returns statistics about background jobs
func (bw *BackgroundWorker) GetJobStats() (map[string]int, error) {
	query := `
	SELECT status, COUNT(*) as count
	FROM jobs
	WHERE created_at > datetime('now', '-1 hour')
	GROUP BY status
	`

	rows, err := bw.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}

	return stats, nil
}

// Shutdown gracefully shuts down the background worker
func (bw *BackgroundWorker) Shutdown() {
	log.Printf("🛑 Shutting down background worker...")
	close(bw.shutdown)
	
	// Wait a bit for workers to finish current jobs
	select {
	case <-time.After(5 * time.Second):
		log.Printf("⏰ Shutdown timeout reached")
	case <-bw.done:
		log.Printf("✅ Background worker shut down cleanly")
	}

	if err := bw.db.Close(); err != nil {
		log.Printf("❌ Error closing database: %v", err)
	}
}

// EnqueueHookAPI provides a simple API for enqueuing hooks
func EnqueueHookAPI(dbPath, hookName, args string, priority int) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO jobs (hook_name, args, priority) VALUES (?, ?, ?)`
	_, err = db.Exec(query, hookName, args, priority)
	return err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <command> [args...]\n", os.Args[0])
		fmt.Printf("Commands:\n")
		fmt.Printf("  start <hooks-dir>     - Start the background worker daemon\n")
		fmt.Printf("  enqueue <hook> <args> <priority> - Enqueue a background job\n")
		fmt.Printf("  stats                 - Show job statistics\n")
		os.Exit(1)
	}

	hooksDir := "/Users/a.pezzotta/.claude/hooks"
	dbPath := filepath.Join(hooksDir, "queue", "background-jobs.db")

	switch os.Args[1] {
	case "start":
		if len(os.Args) > 2 {
			hooksDir = os.Args[2]
		}

		worker, err := NewBackgroundWorker(dbPath, hooksDir, 3)
		if err != nil {
			log.Fatalf("Failed to create background worker: %v", err)
		}

		ctx := context.Background()
		worker.Start(ctx)

	case "enqueue":
		if len(os.Args) < 5 {
			fmt.Printf("Usage: %s enqueue <hook-name> <args> <priority>\n", os.Args[0])
			os.Exit(1)
		}

		hookName := os.Args[2]
		args := os.Args[3]
		priority := 5
		if len(os.Args) > 4 {
			fmt.Sscanf(os.Args[4], "%d", &priority)
		}

		if err := EnqueueHookAPI(dbPath, hookName, args, priority); err != nil {
			log.Fatalf("Failed to enqueue job: %v", err)
		}
		
		fmt.Printf("✅ Enqueued background job: %s\n", hookName)

	case "stats":
		worker, err := NewBackgroundWorker(dbPath, hooksDir, 1)
		if err != nil {
			log.Fatalf("Failed to create background worker: %v", err)
		}

		stats, err := worker.GetJobStats()
		if err != nil {
			log.Fatalf("Failed to get stats: %v", err)
		}

		fmt.Printf("📊 Background Job Statistics (last hour):\n")
		for status, count := range stats {
			fmt.Printf("  %s: %d\n", status, count)
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}