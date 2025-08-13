# Claude WM CLI

Window Management CLI Tool with AI Integration using package manager approach for configuration management.

## 🚀 Quickstart

### Installation

```bash
# Build from source
git clone <repository-url>
cd claude-wm-cli
make build

# The binary will be available at ./build/claude-wm-cli
```

### First Steps

```bash
# 1. Install system configuration
claudewm config install

# 2. Check configuration status
claudewm config status

# 3. Update configuration (preview first)
claudewm config update --dry-run
claudewm config update

# 4. Initialize project workspace
claudewm init my-project
```

## 📋 Key Commands

### Configuration Management
- `config install` - Install system config to .claude/ and .wm/baseline/
- `config status` - Show differences between upstream, baseline, and local
- `config update --dry-run` - Preview configuration updates
- `config update --no-backup` - Update without creating backups
- `config migrate-legacy` - Migrate from .claude-wm to .wm structure

### Development & Testing
- `dev sandbox` - Create testing sandbox from upstream files
- `dev sandbox diff --apply` - Apply sandbox changes to baseline
- `dev sandbox diff --only <pattern>` - Apply specific files only
- `dev sandbox diff --allow-delete` - Allow file deletion during sync

### Safety & Validation
- `guard check` - Validate changes against writing restrictions
- `guard install-hook` - Install git pre-commit validation hook

### Project Management
- `init <project>` - Initialize new project with proper structure
- `status` - Show current project state and progress
- `execute "command"` - Execute Claude commands with timeout protection

## 🏗️ Architecture Overview

Claude WM CLI uses a **4-space configuration model**:

- **Upstream (A)**: `internal/config/system` - Embedded system templates
- **Baseline (B)**: `.wm/baseline` - Installed reference state
- **Local (L)**: `.claude` - User workspace and customizations
- **Sandbox (S)**: `.wm/sandbox/claude` - Testing environment

### 3-Way Merge Flow
```
Upstream (A) ──┐
               ├─→ 3-way merge ──→ Local (L)
Baseline (B) ──┘
```

## 🔧 Configuration Flags

- `--dry-run` - Preview changes without applying
- `--no-backup` - Skip backup creation during updates
- `--only <pattern>` - Apply only matching files/patterns
- `--allow-delete` - Allow file deletion during operations
- `--apply` - Apply changes after preview
- `--timeout <seconds>` - Custom timeout for execute commands
- `--retries <count>` - Custom retry count for execute commands

## 📖 Documentation

- [Architecture Guide](docs/ARCHITECTURE.md) - Detailed technical architecture
- [Configuration Guide](docs/CONFIG_GUIDE.md) - Complete command reference with flags
- [Testing Guide](docs/TESTING.md) - L0-L3 testing protocols and `make test-all`
- [Knowledge Base](docs/KB/) - Glossary, commands reference, file ownership rules
- [ADR](docs/ADR/) - Architectural Decision Records
- [MCP Playbook](docs/mcp-playbook.md) - MCP tools usage guide

## 🚨 Quick Examples

```bash
# Safe configuration update with preview
claudewm config update --dry-run  # Preview changes
claudewm config update             # Apply changes

# Sandbox testing workflow
claudewm dev sandbox                           # Create sandbox
claudewm dev sandbox diff                      # Show differences
claudewm dev sandbox diff --apply --only cmd/  # Apply only cmd/ changes

# Project initialization
claudewm init my-project
cd my-project
claudewm status
```

## 🔒 Safety Features

- **Atomic operations** with automatic backups
- **3-way merge** conflict detection
- **File ownership** boundary enforcement
- **Git pre-commit** hooks for validation
- **Dry-run mode** for safe previews

See [docs/](docs/) for complete documentation.