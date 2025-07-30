package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ProjectStatus struct {
	GitStatus      GitInfo
	TodoStatus     TodoInfo
	RecentFiles    []FileInfo
	Timestamp      time.Time
}

type GitInfo struct {
	Branch        string
	Status        string
	ModifiedFiles []string
	UntrackedFiles []string
	HasChanges    bool
}

type TodoInfo struct {
	FilePath      string
	TotalTasks    int
	CompletedTasks int
	PendingTasks  int
	Progress      float64
	LastModified  time.Time
}

type FileInfo struct {
	Path         string
	ModTime      time.Time
	Size         int64
}

func main() {
	// Get project status
	status, err := getProjectStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting project status: %v\n", err)
		os.Exit(0) // Non-blocking: don't fail the main operation
	}

	// Display status dashboard
	displayStatus(status)
}

func getProjectStatus() (*ProjectStatus, error) {
	status := &ProjectStatus{
		Timestamp: time.Now(),
	}

	// Get Git status
	gitInfo, err := getGitStatus()
	if err != nil {
		// Git might not be available, continue without it
		gitInfo = &GitInfo{Status: "unavailable"}
	}
	status.GitStatus = *gitInfo

	// Get TODO status
	todoInfo, err := getTodoStatus()
	if err != nil {
		// TODO file might not exist, continue without it
		todoInfo = &TodoInfo{}
	}
	status.TodoStatus = *todoInfo

	// Get recent files
	recentFiles, err := getRecentFiles()
	if err != nil {
		// Continue without recent files if error
		recentFiles = []FileInfo{}
	}
	status.RecentFiles = recentFiles

	return status, nil
}

func getGitStatus() (*GitInfo, error) {
	gitInfo := &GitInfo{}

	// Check if we're in a git repository
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return gitInfo, fmt.Errorf("not in git repository")
	}

	// Get current branch
	cmd = exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return gitInfo, fmt.Errorf("failed to get branch: %v", err)
	}
	gitInfo.Branch = strings.TrimSpace(string(output))

	// Get status
	cmd = exec.Command("git", "status", "--porcelain")
	output, err = cmd.Output()
	if err != nil {
		return gitInfo, fmt.Errorf("failed to get status: %v", err)
	}

	statusLines := strings.Split(string(output), "\n")
	var modifiedFiles, untrackedFiles []string

	for _, line := range statusLines {
		if len(line) < 3 {
			continue
		}
		
		statusCode := line[:2]
		fileName := line[3:]
		
		if statusCode == "??" {
			untrackedFiles = append(untrackedFiles, fileName)
		} else if statusCode[0] != ' ' || statusCode[1] != ' ' {
			modifiedFiles = append(modifiedFiles, fileName)
		}
	}

	gitInfo.ModifiedFiles = modifiedFiles
	gitInfo.UntrackedFiles = untrackedFiles
	gitInfo.HasChanges = len(modifiedFiles) > 0 || len(untrackedFiles) > 0

	if !gitInfo.HasChanges {
		gitInfo.Status = "clean"
	} else {
		gitInfo.Status = "dirty"
	}

	return gitInfo, nil
}

func getTodoStatus() (*TodoInfo, error) {
	// Look for TODO.md in common locations
	locations := []string{
		"TODO.md",
		"docs/current-epic/TODO.md",
		"docs/2-current-epic/TODO.md",
		"docs/3-current-task/TODO.md",
	}

	var todoPath string
	var todoStat os.FileInfo
	
	for _, location := range locations {
		if stat, err := os.Stat(location); err == nil {
			todoPath = location
			todoStat = stat
			break
		}
	}

	if todoPath == "" {
		return &TodoInfo{}, fmt.Errorf("no TODO.md file found")
	}

	content, err := os.ReadFile(todoPath)
	if err != nil {
		return &TodoInfo{}, fmt.Errorf("failed to read TODO.md: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	totalTasks := 0
	completedTasks := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
			completedTasks++
			totalTasks++
		} else if strings.HasPrefix(trimmed, "- [ ]") {
			totalTasks++
		}
	}

	progress := 0.0
	if totalTasks > 0 {
		progress = float64(completedTasks) / float64(totalTasks) * 100
	}

	return &TodoInfo{
		FilePath:      todoPath,
		TotalTasks:    totalTasks,
		CompletedTasks: completedTasks,
		PendingTasks:  totalTasks - completedTasks,
		Progress:      progress,
		LastModified:  todoStat.ModTime(),
	}, nil
}

func getRecentFiles() ([]FileInfo, error) {
	var files []FileInfo
	
	// Look for recently modified files (last 10 minutes)
	cutoff := time.Now().Add(-10 * time.Minute)
	
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}
		
		// Skip directories and hidden files
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		
		// Skip binary and large files
		if info.Size() > 1024*1024 { // 1MB
			return nil
		}
		
		// Only include recently modified files
		if info.ModTime().After(cutoff) {
			files = append(files, FileInfo{
				Path:    path,
				ModTime: info.ModTime(),
				Size:    info.Size(),
			})
		}
		
		return nil
	})
	
	if err != nil {
		return files, fmt.Errorf("failed to scan files: %v", err)
	}
	
	return files, nil
}

func displayStatus(status *ProjectStatus) {
	fmt.Printf("📊 Project Status Dashboard\n")
	fmt.Printf("═════════════════════════════\n")
	
	// Git Status
	if status.GitStatus.Status != "unavailable" {
		fmt.Printf("🔧 Git Status: ")
		if status.GitStatus.HasChanges {
			fmt.Printf("🔴 %s (%d modified, %d untracked)\n", 
				status.GitStatus.Branch,
				len(status.GitStatus.ModifiedFiles),
				len(status.GitStatus.UntrackedFiles))
		} else {
			fmt.Printf("🟢 %s (clean)\n", status.GitStatus.Branch)
		}
	}
	
	// TODO Status
	if status.TodoStatus.TotalTasks > 0 {
		fmt.Printf("📋 TODO Progress: %.1f%% (%d/%d tasks completed)\n",
			status.TodoStatus.Progress,
			status.TodoStatus.CompletedTasks,
			status.TodoStatus.TotalTasks)
		
		if status.TodoStatus.PendingTasks > 0 {
			fmt.Printf("   📌 %d pending tasks in %s\n",
				status.TodoStatus.PendingTasks,
				status.TodoStatus.FilePath)
		}
	}
	
	// Recent Files
	if len(status.RecentFiles) > 0 {
		fmt.Printf("📝 Recent Changes (%d files):\n", len(status.RecentFiles))
		for i, file := range status.RecentFiles {
			if i >= 5 { // Show max 5 recent files
				fmt.Printf("   ... and %d more files\n", len(status.RecentFiles)-i)
				break
			}
			fmt.Printf("   %s (%s)\n", 
				file.Path, 
				file.ModTime.Format("15:04:05"))
		}
	}
	
	fmt.Printf("\n")
}