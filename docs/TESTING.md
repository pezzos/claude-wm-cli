# Testing Guide

This document describes the comprehensive testing strategy and procedures for Claude WM CLI.

## Testing Philosophy

Claude WM CLI uses a **4-level testing pyramid** (L0-L3) that ensures reliability through progressive validation layers:

```
    ┌─────┐
    │ L3  │  System Tests (End-to-End)
    ├─────┤
    │ L2  │  Integration Tests  
    ├─────┤
    │ L1  │  Unit Tests
    ├─────┤
    │ L0  │  Smoke Tests
    └─────┘
```

## Testing Protocol

### L0: Smoke Tests (Basic Functionality)

**Purpose**: Quick validation that core functionality works
**Duration**: < 30 seconds
**Frequency**: Every build, before commits

**Coverage:**
- Binary compilation
- Basic command execution
- Help system functionality
- Version information

**Commands:**
```bash
# Manual smoke tests
make smoke-test

# Or individual commands
claude-wm-cli --version
claude-wm-cli --help
claude-wm-cli config --help
```

**Expected Results:**
- All commands exit with status 0
- Help text displays correctly
- Version information is accurate

---

### L1: Unit Tests (Component Testing)

**Purpose**: Validate individual components in isolation
**Duration**: < 2 minutes
**Frequency**: Every commit, during development

**Coverage:**
- Domain logic validation
- Utility function correctness
- Error handling scenarios
- Data structure operations

**Commands:**
```bash
# Run all unit tests
make test-unit

# Or using Go directly
go test ./internal/... -short

# Specific packages
go test ./internal/diff/
go test ./internal/config/
go test ./internal/cmd/
```

**Key Test Areas:**

**Configuration Management:**
```bash
go test ./internal/config/ -v
go test ./internal/cmd/ -run TestConfig -v
```

**Diff Engine:**
```bash
go test ./internal/diff/ -v
```

**Validation Systems:**
```bash
go test ./internal/validation/ -v
go test ./internal/model/ -v
```

**File Operations:**
```bash
go test ./internal/fsutil/ -v
go test ./internal/ziputil/ -v
```

---

### L2: Integration Tests (Component Interaction)

**Purpose**: Test interactions between components
**Duration**: < 5 minutes
**Frequency**: Before major commits, in CI

**Coverage:**
- Configuration workflow end-to-end
- Sandbox operations
- Migration procedures
- Hook integration

**Commands:**
```bash
# Run all integration tests
make test-integ

# Or using Go directly
go test ./internal/integration/ -v
go test ./cmd/ -run Integration -v

# Specific integration areas
go test ./internal/integration/ -run TestConfigWorkflow -v
go test ./internal/integration/ -run TestSandboxIntegration -v
```

**Test Scenarios:**

**Configuration Lifecycle:**
```bash
# Test full config workflow
go test ./internal/integration/ -run TestConfigInstallUpdateCycle -v
```

**Sandbox Development:**
```bash
# Test sandbox creation and diff
go test ./internal/integration/ -run TestSandboxWorkflow -v
```

**Legacy Migration:**
```bash
# Test migration from .claude-wm to .wm
go test ./internal/cmd/ -run TestMigrateLegacy -v
```

**Guard System:**
```bash
# Test hook installation and execution
go test ./internal/integration/ -run TestGuardIntegration -v
```

---

### L3: System Tests (End-to-End)

**Purpose**: Full system validation in realistic scenarios
**Duration**: < 10 minutes
**Frequency**: Before releases, weekly

**Coverage:**
- Complete user workflows
- Cross-platform compatibility
- Performance validation
- Error recovery scenarios

**Commands:**
```bash
# Run all system tests
make test-system

# Or using Go directly
go test ./test/system/ -v -timeout 10m
```

**Test Scenarios:**

**New Project Setup:**
```bash
# Test complete project initialization
go test ./test/system/ -run TestNewProjectWorkflow -v
```

**Configuration Update Cycle:**
```bash
# Test full update workflow with conflicts
go test ./test/system/ -run TestConfigUpdateWithConflicts -v
```

**Development Workflow:**
```bash
# Test sandbox development and upstreaming
go test ./test/system/ -run TestDevelopmentWorkflow -v
```

**Migration Scenarios:**
```bash
# Test legacy migration scenarios
go test ./test/system/ -run TestLegacyMigration -v
```

---

## Complete Test Suite

### `make test-all`

Runs the complete testing protocol in sequence:

```bash
make test-all
```

**Process:**
1. **Build Verification**: Ensures clean build
2. **L0 Smoke Tests**: Basic functionality validation
3. **L1 Unit Tests**: Component testing
4. **L2 Integration Tests**: Component interaction testing
5. **L3 System Tests**: End-to-end validation
6. **Coverage Report**: Test coverage analysis
7. **Performance Benchmarks**: Performance regression checks

**Target Execution Time**: < 15 minutes
**Success Criteria**: All test levels pass with >80% coverage

---

## Specialized Testing

### Guard and Hook Testing

**Pre-commit Hook Validation:**
```bash
# Test hook installation
go test ./internal/cmd/ -run TestGuardInstallHook -v

# Test hook execution
go test ./internal/integration/ -run TestGuardExecution -v

# Manual hook testing
claude-wm-cli guard install-hook
echo '{"invalid": json}' > test.json
git add test.json
git commit -m "Test commit"  # Should fail
```

**JSON Validation:**
```bash
# Test JSON validator hooks
go test ./internal/validation/ -run TestJSONValidator -v

# Manual validation
claude-wm-cli guard check
```

### Configuration Testing

**3-Way Merge Testing:**
```bash
# Test merge scenarios
go test ./internal/update/ -run TestMergePlanning -v

# Test conflict detection
go test ./internal/update/ -run TestConflictDetection -v
```

**Atomic Operations:**
```bash
# Test atomic file operations
go test ./internal/fsutil/ -run TestAtomicOperations -v

# Test backup and restore
go test ./internal/ziputil/ -run TestBackupRestore -v
```

### Performance Testing

**Benchmarks:**
```bash
# Run performance benchmarks
go test ./internal/diff/ -bench=. -benchmem
go test ./internal/config/ -bench=. -benchmem

# Specific benchmarks
go test ./internal/diff/ -bench=BenchmarkDiffTrees -v
```

**Load Testing:**
```bash
# Test with large file sets
go test ./test/performance/ -run TestLargeFileSet -v

# Memory usage testing
go test ./internal/diff/ -run TestMemoryUsage -v
```

---

## Manual Testing Scenarios

### Quick Validation (5 minutes)

For rapid validation during development:

```bash
# 1. Basic functionality
claude-wm-cli --version
claude-wm-cli --help

# 2. Configuration workflow
cd /tmp/test-project
claude-wm-cli config install
claude-wm-cli config status
claude-wm-cli config update --dry-run

# 3. Sandbox workflow
claude-wm-cli dev sandbox
echo "test change" >> .wm/sandbox/claude/test.md
claude-wm-cli dev sandbox diff

# 4. Guard system
claude-wm-cli guard check
```

### Comprehensive Validation (15 minutes)

For thorough manual validation:

```bash
# Setup test environment
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"

# 1. Fresh installation
claude-wm-cli config install
ls -la .claude/ .wm/

# 2. Status and updates
claude-wm-cli config status
claude-wm-cli config update --dry-run
claude-wm-cli config update

# 3. Sandbox development
claude-wm-cli dev sandbox
cd .wm/sandbox/claude/
echo "# Test modification" >> agents/test-agent.md
cd - 
claude-wm-cli dev sandbox diff
claude-wm-cli dev sandbox diff --apply --dry-run

# 4. Migration testing (if legacy exists)
# mkdir .claude-wm/system -p
# echo "legacy content" > .claude-wm/system/legacy.json
# claude-wm-cli config migrate-legacy --dry-run

# 5. Guard installation
claude-wm-cli guard install-hook
claude-wm-cli guard check

# Cleanup
rm -rf "$TEST_DIR"
```

---

## Test Data and Fixtures

### Test Directory Structure

Tests use standardized fixtures in `testdata/`:

```
testdata/
├── configs/
│   ├── minimal/          # Minimal valid configuration
│   ├── complex/          # Complex multi-file configuration
│   └── invalid/          # Invalid configurations for error testing
├── legacy/
│   ├── claude-wm-v1/     # Legacy .claude-wm structure
│   └── claude-wm-v2/     # Updated legacy structure
└── scenarios/
    ├── fresh-install/    # Clean installation scenario
    ├── existing-config/  # Existing configuration scenario
    └── migration/        # Migration scenarios
```

### Test Utilities

**Temporary Environment Creation:**
```go
func setupTestEnv(t *testing.T) string {
    dir := t.TempDir()
    // Setup test environment
    return dir
}
```

**Configuration Fixtures:**
```go
func loadTestConfig(name string) (*Config, error) {
    path := filepath.Join("testdata", "configs", name)
    return LoadConfig(path)
}
```

---

## Continuous Integration

### GitHub Actions Workflow

The CI pipeline runs all test levels:

```yaml
name: Test Suite
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - run: make test-all
      - run: make coverage-report
```

### Coverage Requirements

**Minimum Coverage Targets:**
- **Overall**: 80%
- **Domain Layer**: 90%
- **Critical Paths**: 95%
- **Error Handling**: 85%

**Coverage Commands:**
```bash
# Generate coverage report
make coverage

# View HTML coverage report
make coverage-html
open coverage.html
```

---

## Debugging and Troubleshooting

### Test Debugging

**Verbose Test Output:**
```bash
go test ./internal/config/ -v -run TestSpecificFunction
```

**Test with Debug Logging:**
```bash
go test ./internal/config/ -v -args -debug
```

**Race Condition Detection:**
```bash
go test ./internal/... -race
```

### Common Issues

**Temporary Directory Cleanup:**
```bash
# Find leftover test directories
find /tmp -name "claude-wm-test-*" -type d

# Clean up
rm -rf /tmp/claude-wm-test-*
```

**Port Conflicts (Integration Tests):**
```bash
# Check for port usage
lsof -i :8080

# Kill conflicting processes
pkill -f "claude-wm-test"
```

**File Permission Issues:**
```bash
# Reset permissions in test directory
chmod -R 755 testdata/
```

---

## Performance Monitoring

### Benchmark Tracking

Regular benchmarking ensures performance stability:

```bash
# Run benchmarks with memory profiling
go test ./internal/diff/ -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

### Performance Regression Detection

**Automated Performance Testing:**
```bash
# Compare benchmark results
go test ./internal/... -bench=. -count=5 > new.bench
benchcmp old.bench new.bench
```

**Memory Usage Monitoring:**
```bash
# Monitor memory usage during tests
go test ./internal/config/ -memprofile=mem.prof
go tool pprof -web mem.prof
```

---

## Test Maintenance

### Regular Maintenance Tasks

**Weekly:**
- Review test coverage reports
- Update test data fixtures
- Performance regression analysis

**Monthly:**
- Audit test execution times
- Update integration test scenarios
- Review and update test documentation

**Per Release:**
- Full system test validation
- Cross-platform testing
- Performance benchmark validation
- Test data cleanup and updates

### Test Quality Metrics

**Target Metrics:**
- **Test Execution Time**: L0 <30s, L1 <2m, L2 <5m, L3 <10m
- **Test Reliability**: >99% pass rate
- **Coverage**: >80% overall, >90% critical paths
- **Maintainability**: Tests updated with code changes