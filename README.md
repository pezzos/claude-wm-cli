# Claude WM CLI

A **production-ready** Go-based CLI tool for intelligent window management and project workflow automation with AI integration. Built with **Clean Architecture principles** and comprehensive configuration management, providing enterprise-grade workflow management with atomic state operations, comprehensive validation, and intelligent guidance systems.

## ✨ Key Features

### Core Capabilities
- **🔧 Configuration Management**: Package manager approach with 3-way merge, baseline tracking, and migration tools
- **📦 Sandbox Environment**: Isolated testing environment for safe experimentation
- **🛡️ Guard System**: Pre-commit hooks and validation guards for code quality
- **⚡ Development Tools**: Comprehensive diff tools, migration utilities, and development commands
- **🏗️ Clean Architecture**: Domain-driven design with strict separation of concerns
- **🔒 Atomic Operations**: Corruption-resistant file operations with validation
- **📊 Comprehensive Testing**: L0-L3 testing protocol with extensive coverage

### Configuration System
- **Embedded Upstream**: Built-in system templates and configurations
- **Baseline Tracking**: `.wm/baseline/` for installation snapshots
- **Local Customization**: `.claude/` for runtime configurations
- **Migration Tools**: Seamless migration from legacy `.claude-wm/` to `.wm/` structure

## 🚀 Quick Start

### Prerequisites
- Go 1.21+ for building from source
- Git for version control
- (Optional) jq for JSON validation hooks

### Installation

#### Option 1: Download Binary
```bash
# Download latest release from GitHub
curl -L https://github.com/your-org/claude-wm-cli/releases/latest/download/claude-wm-cli-$(uname -s)-$(uname -m).tar.gz | tar xz
sudo mv claude-wm-cli /usr/local/bin/
```

#### Option 2: Build from Source
```bash
git clone https://github.com/your-org/claude-wm-cli.git
cd claude-wm-cli
make build
sudo cp build/claude-wm-cli /usr/local/bin/
```

### Initial Setup

1. **Install System Configuration**
   ```bash
   # Initialize with embedded system templates
   claude-wm-cli config install
   ```

2. **Check Configuration Status**
   ```bash
   # View differences between upstream, baseline, and local
   claude-wm-cli config status
   ```

3. **Preview Updates**
   ```bash
   # See what would change without applying
   claude-wm-cli config update --dry-run
   ```

4. **Apply Updates**
   ```bash
   # Apply configuration updates with backup
   claude-wm-cli config update
   ```

## 🧪 Sandbox Development

The sandbox provides an isolated environment for testing configuration changes:

### Create Sandbox
```bash
# Create sandbox from current system configuration
claude-wm-cli dev sandbox

# Reset existing sandbox
claude-wm-cli dev sandbox --reset
```

### Experiment and Compare
```bash
# Make changes in .wm/sandbox/claude/
cd .wm/sandbox/claude
# ... edit files ...

# Compare sandbox with source
claude-wm-cli dev sandbox diff

# Apply specific changes back to source
claude-wm-cli dev sandbox diff --apply --only "agents/**"

# Dry run to see what would be applied
claude-wm-cli dev sandbox diff --apply --dry-run
```

## 🛡️ Guard System

Install validation guards for code quality:

```bash
# Check current guard status
claude-wm-cli guard check

# Install pre-commit hook
claude-wm-cli guard install-hook
```

## 🔄 Legacy Migration

Migrate from old `.claude-wm/` structure to new `.wm/` structure:

```bash
# Analyze migration plan
claude-wm-cli config migrate-legacy

# Preview migration without applying
claude-wm-cli config migrate-legacy --dry-run

# Apply migration and archive old directory
claude-wm-cli config migrate-legacy --archive
```

## 📁 Directory Structure

```
your-project/
├── .claude/                    # Runtime configuration (auto-generated)
├── .wm/                       # Workspace management
│   ├── baseline/              # Installation baseline
│   ├── backups/               # Configuration backups
│   ├── meta.json              # Workspace metadata
│   └── sandbox/               # Isolated testing environment
│       └── claude/            # Sandbox instance
├── internal/config/system/    # Embedded system templates
└── .claude-wm.bak           # Legacy backup (if migrated)
```

## 🔧 Common Commands

### Configuration Management
```bash
# Install initial system configuration
claude-wm-cli config install

# Show configuration status
claude-wm-cli config status

# Update with 3-way merge
claude-wm-cli config update [--dry-run] [--no-backup]

# Regenerate runtime configuration
claude-wm-cli config sync

# Upgrade system templates
claude-wm-cli config upgrade

# Show effective configuration
claude-wm-cli config show [file]

# Migrate from legacy structure
claude-wm-cli config migrate-legacy [--dry-run] [--archive]
```

### Development Tools
```bash
# Create/reset development sandbox
claude-wm-cli dev sandbox [--reset]

# Compare and apply sandbox changes
claude-wm-cli dev sandbox diff [--apply] [--only pattern] [--dry-run] [--allow-delete]
```

### Guard System
```bash
# Check validation guards
claude-wm-cli guard check

# Install pre-commit hook
claude-wm-cli guard install-hook
```

### Project Management
```bash
# Initialize new project
claude-wm-cli init [project-name]

# Check project status
claude-wm-cli status

# Interactive mode
claude-wm-cli interactive
```

## 📖 Documentation

For detailed information, see the documentation in `docs/`:

- **[Architecture Guide](docs/ARCHITECTURE.md)** - System architecture, components, and data flows
- **[Configuration Guide](docs/CONFIG_GUIDE.md)** - Detailed configuration management reference
- **[Testing Guide](docs/TESTING.md)** - Testing protocols and validation procedures

## 🏗️ Architecture

The project follows **Clean Architecture** with strict separation of concerns:

```
internal/
├── domain/              # Core Business Logic (Zero Dependencies)
├── application/         # Use Cases & Orchestration
├── infrastructure/     # External Concerns
├── interfaces/         # External World Adapters
└── cmd/                # Command implementations
```

### Key Components
- **Configuration Manager**: Handles install, update, sync, and migration
- **Diff Engine**: Compares file trees and generates change plans
- **Sandbox System**: Isolated testing environment management
- **Guard System**: Pre-commit hooks and validation
- **Migration Engine**: Legacy structure migration tools

## 🧪 Testing

Run the test suite:

```bash
# Run all tests
make test-all

# Run specific test levels
make test-unit      # L1: Unit tests
make test-integ     # L2: Integration tests  
make test-system    # L3: End-to-end tests
```

See [docs/TESTING.md](docs/TESTING.md) for detailed testing procedures.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run the full test suite: `make test-all`
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Issue Tracker](https://github.com/your-org/claude-wm-cli/issues)
- [Discussions](https://github.com/your-org/claude-wm-cli/discussions)
- [Releases](https://github.com/your-org/claude-wm-cli/releases)