# Claude WM CLI

A window management command-line tool that integrates with Claude AI to provide intelligent window management and command execution capabilities.

## 🚀 Quick Start

### Installation

```bash
# Build from source
git clone <repository-url>
cd claude-wm-cli
make build

# The binary will be available at ./build/claude-wm-cli
```

### Basic Usage

```bash
# Initialize a new project
claude-wm-cli init my-project
cd my-project

# Check project status
claude-wm-cli status

# Execute Claude commands
claude-wm-cli execute "claude --help"

# Use verbose mode for detailed output
claude-wm-cli --verbose status
```

## 📋 Commands

### `init` - Initialize Project
Initialize a new Claude WM CLI project with proper directory structure.

```bash
claude-wm-cli init [project-name] [flags]

# Examples:
claude-wm-cli init my-project          # Create in ./my-project
claude-wm-cli init                     # Initialize current directory
claude-wm-cli init --force my-project  # Overwrite existing files

# Flags:
  -f, --force   Force initialization (overwrite existing files)
```

### `status` - Show Project State
Display current project state including progress and configuration.

```bash
claude-wm-cli status [flags]

# Examples:
claude-wm-cli status                   # Show current status
claude-wm-cli --verbose status         # Show detailed status
```

### `execute` - Execute Claude Commands
Execute Claude AI commands with robust timeout and retry handling.

```bash
claude-wm-cli execute [command] [flags]

# Examples:
claude-wm-cli execute "claude --help"
claude-wm-cli execute "claude -p '/analyze code'"
claude-wm-cli execute --timeout 60 "claude build"
claude-wm-cli execute --retries 3 "claude test"

# Flags:
  -t, --timeout int   Command timeout in seconds (default 30)
  -r, --retries int   Maximum number of retries (default 2)
```

## 🔧 Configuration

### Configuration File
Claude WM CLI uses YAML configuration files. Default locations:
- `~/.claude-wm-cli.yaml` (global)
- `./.claude-wm-cli.yaml` (project-specific)

```yaml
# Example configuration
verbose: false

project:
  name: "my-project"
  initialized: true

defaults:
  timeout: 30
  retries: 2

# Window management settings
window:
  default_timeout: 30
  max_retries: 2

# Claude AI integration settings
claude:
  model: "claude-3-sonnet"
  max_tokens: 4000
  temperature: 0.1
```

### Environment Variables
Override configuration with environment variables:
- `CLAUDE_WM_VERBOSE=true` - Enable verbose output
- `CLAUDE_WM_TIMEOUT=60` - Set default timeout
- `CLAUDE_WM_RETRIES=3` - Set default retry count

### Global Flags
Available on all commands:
- `--config string` - Custom config file path
- `--verbose, -v` - Verbose output
- `--help, -h` - Show help
- `--version` - Show version information

## 🏗️ Project Structure

When you initialize a project, the following structure is created:

```
my-project/
├── .claude-wm-cli.yaml    # Project configuration
└── docs/
    ├── 1-project/          # Global project documentation
    ├── 2-current-epic/     # Current epic execution files
    ├── 3-current-task/     # Current task breakdown
    └── archive/            # Completed work archive
```

## 🔄 Workflow

1. **Initialize**: Create a new project with `claude-wm-cli init`
2. **Status**: Check current state with `claude-wm-cli status`
3. **Execute**: Run Claude commands with `claude-wm-cli execute`
4. **Configure**: Customize behavior with config files or flags

## 🚨 Error Handling

Claude WM CLI provides comprehensive error handling:

### Validation Errors
- **Empty commands**: Command cannot be empty
- **Invalid timeouts**: Must be 1-3600 seconds
- **Invalid project names**: No special characters allowed
- **File permissions**: Clear messages for access issues

### Execution Errors
- **Timeout handling**: 30-second default with configurable override
- **Retry logic**: Automatic retry for transient failures
- **Network failures**: Specific handling for connection issues
- **Command failures**: Detailed output and suggestions

### Error Examples

```bash
# Invalid command
$ claude-wm-cli execute ""
❌ Command cannot be empty. Please provide a valid command to execute.
💡 Try: claude-wm-cli execute "claude --help"

# Invalid timeout
$ claude-wm-cli execute --timeout -5 "claude test"
❌ Timeout must be at least 1 second.
💡 Try: claude-wm-cli execute --timeout 30 "your-command"

# Invalid project name
$ claude-wm-cli init "project/with/slashes"
❌ Project name contains invalid character '/'. Use only letters, numbers, hyphens, and underscores.
💡 Try: claude-wm-cli init my-project
```

## 🔧 Development

### Prerequisites
- Go 1.19+
- Make (for build targets)

### Building
```bash
make build          # Build binary
make test           # Run tests
make lint           # Run linter
make clean          # Clean build artifacts
```

### Development Setup
```bash
make dev            # Install development dependencies
make fmt            # Format code
```

## 📊 Performance

- **Startup time**: <500ms target
- **Command execution**: 2-5s typical, 30s maximum
- **State operations**: Sub-second target
- **Memory usage**: Monitored and limited

## 🐛 Troubleshooting

### Common Issues

**Command not found**:
```bash
# Make sure the binary is in your PATH or use full path
./build/claude-wm-cli --help
```

**Permission denied**:
```bash
# Check file permissions
chmod +x ./build/claude-wm-cli
```

**Config file errors**:
```bash
# Validate YAML syntax
claude-wm-cli --config /path/to/config.yaml status
```

**Timeout issues**:
```bash
# Increase timeout for slow commands
claude-wm-cli execute --timeout 120 "your-slow-command"
```

### Getting Help

- Use `--help` flag on any command for detailed usage
- Use `--verbose` flag for detailed execution information
- Check configuration with `claude-wm-cli status`

## 📄 License

[License information would go here]

## 🤝 Contributing

[Contributing guidelines would go here]