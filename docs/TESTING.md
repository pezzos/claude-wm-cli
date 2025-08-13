# Testing Guide

Testing protocols for claude-wm-cli with L0-L3 classification and comprehensive test execution.

## Testing Levels

### L0: Unit Tests
**Scope:** Individual functions and methods  
**Duration:** <100ms per test  
**Environment:** In-memory, no I/O  
**Coverage:** Business logic, validation, algorithms  

```bash
# Run unit tests
go test -short ./...

# With coverage
go test -short -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific package
go test -short ./internal/config
go test -short ./internal/diff
go test -short ./internal/validation
```

### L1: Integration Tests
**Scope:** Component interactions  
**Duration:** <5s per test  
**Environment:** TempDir, filesystem I/O  
**Coverage:** Command flows, file operations, 3-way merge  

```bash
# Run integration tests
go test -run Integration ./...

# Specific integration suites
go test -run Integration ./cmd
go test -run Integration ./internal/config
go test -run Integration ./internal/update
```

### L2: System Tests
**Scope:** End-to-end command execution  
**Duration:** <30s per test  
**Environment:** Real CLI execution, isolated directories  
**Coverage:** Full command workflows, error scenarios  

```bash
# Run system tests
go test -tags=system ./test/system

# Specific system scenarios
go test -tags=system -run TestConfigInstallWorkflow ./test/system
go test -tags=system -run TestSandboxWorkflow ./test/system
go test -tags=system -run TestMigrationWorkflow ./test/system
```

### L3: Acceptance Tests
**Scope:** User scenarios and workflows  
**Duration:** <2min per test  
**Environment:** Real projects, Git repositories  
**Coverage:** Complete user journeys, multi-command flows  

```bash
# Run acceptance tests
go test -tags=acceptance ./test/acceptance

# Long-running tests
go test -tags=acceptance -timeout=10m ./test/acceptance
```

## Comprehensive Test Execution

### make test-all
Executes all test levels in sequence with proper setup/teardown:

```bash
make test-all
```

**Equivalent to:**
```bash
# L0: Unit tests
go test -short ./...

# L1: Integration tests  
go test -run Integration ./...

# L2: System tests
go test -tags=system ./test/system

# L3: Acceptance tests
go test -tags=acceptance ./test/acceptance

# Generate combined coverage report
go tool cover -html=coverage-combined.out
```

### Parallel Execution
```bash
# Run tests in parallel (faster)
make test-parallel

# Equivalent to:
go test -short -p 4 ./...         # L0 with 4 cores
go test -run Integration -p 2 ./...  # L1 with 2 cores
go test -tags=system ./test/system   # L2 sequential
go test -tags=acceptance ./test/acceptance  # L3 sequential
```

## Test Scenarios by Component

### Configuration Management Tests

#### config install
```bash
# L1: Integration
go test -run TestConfigInstall ./internal/config

# L2: System
go test -tags=system -run TestConfigInstallWorkflow ./test/system
```

**Test scenarios:**
- Fresh install to empty directory
- Install with existing .claude/ (conflict handling)
- Install with --force flag
- Install with --dry-run preview
- Atomic failure recovery

#### config status
```bash
# L0: Unit
go test -run TestDiffEngine ./internal/diff

# L1: Integration  
go test -run TestConfigStatus ./internal/config
```

**Test scenarios:**
- No differences (clean state)
- Files added in upstream
- Files modified in local
- Files deleted in upstream
- Complex 3-way merge scenarios

#### config update
```bash
# L1: Integration
go test -run TestConfigUpdate ./internal/config
go test -run TestThreeWayMerge ./internal/update

# L2: System
go test -tags=system -run TestUpdateWorkflow ./test/system
```

**Test scenarios:**
- Simple apply (no conflicts)
- Merge conflicts requiring resolution
- Partial updates with --only flag
- Updates with --allow-delete
- Backup creation and restoration

### Development Tests

#### dev sandbox
```bash
# L1: Integration
go test -run TestDevSandbox ./internal/cmd

# L2: System
go test -tags=system -run TestSandboxWorkflow ./test/system
```

**Test scenarios:**
- Create sandbox from upstream
- Sandbox isolation verification
- Partial sandbox with --only
- Sandbox overwrite with --force

#### dev sandbox diff
```bash
# L1: Integration
go test -run TestSandboxDiff ./internal/cmd
```

**Test scenarios:**
- Diff display formatting
- Apply changes to baseline
- Selective apply with --only
- File deletion with --allow-delete

### Guard System Tests

#### guard check
```bash
# L0: Unit
go test -run TestGuardRules ./internal/guard

# L1: Integration
go test -run TestGuardCheck ./internal/integration
```

**Test scenarios:**
- Valid changes (no violations)
- Unauthorized writes to upstream space
- Boundary enforcement validation
- Git working tree scanning

#### guard install-hook
```bash
# L2: System
go test -tags=system -run TestGuardHook ./test/system
```

**Test scenarios:**
- Hook installation to .git/hooks/
- Hook execution during git commit
- Hook bypass with --no-verify
- Hook updates and versioning

## Test Environment Setup

### TempDir Pattern (L1/L2)
```go
// Standard test directory setup
func setupTestEnvironment(t *testing.T) string {
    tempDir := t.TempDir() // Automatically cleaned up
    
    // Create test structure
    testDirs := []string{
        ".claude",
        ".wm/baseline", 
        ".wm/sandbox/claude",
    }
    
    for _, dir := range testDirs {
        err := os.MkdirAll(filepath.Join(tempDir, dir), 0755)
        require.NoError(t, err)
    }
    
    return tempDir
}
```

### Git Repository Setup (L2/L3)
```go
// Git repository for integration tests
func setupGitRepository(t *testing.T, dir string) {
    cmd := exec.Command("git", "init")
    cmd.Dir = dir
    err := cmd.Run()
    require.NoError(t, err)
    
    // Configure git for testing
    configCmds := [][]string{
        {"git", "config", "user.name", "Test User"},
        {"git", "config", "user.email", "test@example.com"},
        {"git", "config", "init.defaultBranch", "main"},
    }
    
    for _, cmdArgs := range configCmds {
        cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
        cmd.Dir = dir
        err := cmd.Run()
        require.NoError(t, err)
    }
}
```

## Performance Testing

### Benchmark Tests
```bash
# Run benchmarks
go test -bench=. ./internal/diff
go test -bench=. ./internal/config
go test -bench=BenchmarkThreeWayMerge ./internal/update
```

**Performance targets:**
- Config status: <200ms for 100 files
- 3-way merge: <500ms for 100 files
- Sandbox creation: <1s for full upstream
- Backup creation: <2s for 10MB content

### Load Testing
```bash
# Stress test with large configurations
go test -tags=stress -run TestLargeConfiguration ./test/stress

# Concurrent operations
go test -tags=stress -run TestConcurrentOperations ./test/stress
```

## Error Scenario Testing

### Network Isolation
```bash
# Test without network access
GO_TEST_OFFLINE=true go test ./...
```

### Filesystem Errors
```bash
# Test with readonly filesystem
GO_TEST_READONLY=true go test ./test/system

# Test with disk full conditions
GO_TEST_DISKFULL=true go test ./test/system
```

### Interrupted Operations
```bash
# Test atomic operation recovery
go test -run TestAtomicRecovery ./internal/update
go test -run TestInterruptedBackup ./internal/backup
```

## Continuous Integration

### GitHub Actions Integration
```yaml
# .github/workflows/test.yml
- name: Run L0-L2 Tests
  run: make test-ci
  
- name: Run L3 Acceptance Tests  
  run: make test-acceptance
  if: github.event_name == 'pull_request'
```

### Test Matrix
```bash
# Cross-platform testing
make test-all GOOS=linux
make test-all GOOS=darwin 
make test-all GOOS=windows

# Go version matrix
make test-all GO_VERSION=1.19
make test-all GO_VERSION=1.20
make test-all GO_VERSION=1.21
```

## Test Data Management

### Fixtures
```
test/fixtures/
├── configs/
│   ├── minimal/
│   ├── complete/
│   └── conflicted/
├── repositories/
│   ├── clean/
│   ├── dirty/
│   └── legacy/
└── scenarios/
    ├── install/
    ├── update/
    └── migrate/
```

### Test Utilities
```bash
# Generate test fixtures
go run ./test/tools/generate-fixtures

# Validate test data
go run ./test/tools/validate-fixtures

# Clean test artifacts
make clean-test
```

## Debugging Failed Tests

### Verbose Test Output
```bash
# Debug specific test
go test -v -run TestConfigUpdate ./internal/config

# Debug with race detection
go test -race -run TestConcurrentOperations ./test/system

# Debug with CPU profiling
go test -cpuprofile cpu.prof -run TestLargeConfig ./test/stress
```

### Test Artifacts
```bash
# Preserve test directories for inspection
GO_TEST_PRESERVE=true go test -run TestFailingScenario ./test/system

# Enable debug logging during tests
GO_TEST_DEBUG=true go test -v ./...
```

### Common Debugging
```bash
# Check test coverage gaps
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -E '(0.0%|[1-7][0-9]\.[0-9]%)'

# Find flaky tests
for i in {1..10}; do go test -run TestSuspiciousTest ./... || echo "Failed on run $i"; done
```