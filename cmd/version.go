/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"claude-wm-cli/internal/meta"
	"claude-wm-cli/internal/metrics"

	"github.com/spf13/cobra"
)

var (
	versionOutput string
	versionShort  bool
	versionSimple bool
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version and build information",
	Long: `Display detailed version information including:
- Semantic version number
- Git commit hash and timestamp
- Go version and build information
- Operating system and architecture
- Dependency versions (when verbose)

This information is useful for debugging and support.`,
	Example: `  claude-wm-cli version              # Show full version info
  claude-wm-cli version --short       # Show version number only
  claude-wm-cli version --output json # Output as JSON`,
	Run: func(cmd *cobra.Command, args []string) {
		// Start performance monitoring
		timer := metrics.InstrumentCommand("version")
		defer timer.Stop()
		
		showVersionInfo()
		timer.SetExitCode(0)
	},
}

func showVersionInfo() {
	if versionShort {
		fmt.Println(meta.Version)
		return
	}
	
	if versionSimple {
		fmt.Printf("%s (commit %s, %s)\n", meta.Version, meta.Commit, meta.BuildDate)
		return
	}

	switch versionOutput {
	case "json":
		showVersionJSON()
	case "yaml":
		showVersionYAML()
	default:
		showVersionDefault()
	}
}

func showVersionDefault() {
	fmt.Printf("🚀 Claude WM CLI\n")
	fmt.Printf("================\n\n")

	// Core version info
	fmt.Printf("Version:     %s\n", getVersionString())
	fmt.Printf("Git Commit:  %s\n", meta.Commit)
	fmt.Printf("Built:       %s\n", meta.BuildDate)
	fmt.Printf("Go Version:  %s\n", runtime.Version())
	fmt.Printf("OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)

	if verbose {
		fmt.Printf("\n🔧 Build Details:\n")
		fmt.Printf("Compiler:    %s\n", runtime.Compiler)
		fmt.Printf("NumCPU:      %d\n", runtime.NumCPU())

		// Get build info including dependencies
		if info, ok := debug.ReadBuildInfo(); ok {
			fmt.Printf("\n📦 Dependencies:\n")
			for _, dep := range info.Deps {
				if dep.Path == "github.com/spf13/cobra" ||
					dep.Path == "github.com/spf13/viper" ||
					dep.Path == "github.com/stretchr/testify" {
					fmt.Printf("%-12s %s@%s\n", getShortName(dep.Path)+":", dep.Path, dep.Version)
				}
			}
		}
	}

	fmt.Printf("\n📖 Documentation: docs/README.md\n")
	fmt.Printf("🐛 Report issues: [repository-url]/issues\n")
}

func showVersionJSON() {
	fmt.Printf(`{
  "version": "%s",
  "git_commit": "%s",
  "build_time": "%s",
  "go_version": "%s",
  "os": "%s",
  "arch": "%s",
  "compiler": "%s"
}
`, getVersionString(), meta.Commit, meta.BuildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
}

func showVersionYAML() {
	fmt.Printf(`version: %s
git_commit: %s
build_time: %s
go_version: %s
os: %s
arch: %s
compiler: %s
`, getVersionString(), meta.Commit, meta.BuildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
}

func getVersionString() string {
	if meta.Version == "" || meta.Version == "dev" {
		return "dev (built from source)"
	}
	return meta.Version
}

func getShortName(path string) string {
	switch path {
	case "github.com/spf13/cobra":
		return "Cobra"
	case "github.com/spf13/viper":
		return "Viper"
	case "github.com/stretchr/testify":
		return "Testify"
	default:
		return path
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Command-specific flags
	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "Show version number only")
	versionCmd.Flags().BoolVar(&versionSimple, "simple", false, "Show simple version format: version (commit hash, date)")
	versionCmd.Flags().StringVarP(&versionOutput, "output", "o", "", "Output format: json, yaml (default: human-readable)")
}
