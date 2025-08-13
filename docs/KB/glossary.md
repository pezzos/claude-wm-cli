# Glossary

Definitive terminology for claude-wm-cli architecture and operations.

## Configuration Spaces

### Upstream (A)
**Definition:** System-provided templates and configurations embedded in the binary  
**Location:** `internal/config/system/`  
**Nature:** Read-only, version-controlled with binary  
**Update mechanism:** Binary rebuild/upgrade  
**Ownership:** System developers

### Baseline (B)
**Definition:** Installed reference state representing last known good configuration from Upstream  
**Location:** `.wm/baseline/`  
**Nature:** Package manager controlled, atomically updated  
**Update mechanism:** `config install`, `config update`  
**Ownership:** Package manager

### Local (L)
**Definition:** User workspace for customizations and local state  
**Location:** `.claude/`  
**Nature:** User-editable, preserved during updates  
**Update mechanism:** Manual editing, 3-way merge  
**Ownership:** End user

### Sandbox (S)
**Definition:** Isolated testing environment for safe experimentation  
**Location:** `.wm/sandbox/claude/`  
**Nature:** Temporary, experimental, no production impact  
**Update mechanism:** `dev sandbox` commands  
**Ownership:** Developer/tester

## Operation Types

### Plan
**Definition:** Analysis phase that determines what changes would be applied without making them  
**Usage:** `--dry-run` flag, `config status` command  
**Output:** Shows intended operations (preserve/apply/conflict/delete)  
**Safety:** Non-destructive, read-only analysis

### Apply
**Definition:** Execution phase that implements planned changes to disk  
**Usage:** Commands without `--dry-run`, explicit `--apply` flag  
**Output:** Actual file system modifications  
**Safety:** Destructive, creates backups when possible

### Preserve
**Definition:** Keep existing file unchanged during merge operation  
**Trigger:** No differences detected or user choice to maintain current version  
**Result:** File content remains identical  
**Status:** Safe operation, no backup needed

### Conflict
**Definition:** 3-way merge cannot be automatically resolved  
**Trigger:** Upstream, Baseline, and Local all differ in incompatible ways  
**Result:** Operation pauses, manual resolution required  
**Resolution:** User edits conflicted file, retries operation

### Delete
**Definition:** Remove file during update operation  
**Trigger:** File exists in Baseline/Local but removed in Upstream  
**Safety:** Requires explicit `--allow-delete` flag  
**Backup:** Always creates backup before deletion

## Merge Strategies

### 3-Way Merge
**Definition:** Conflict resolution using three versions: Upstream (A), Baseline (B), Local (L)  
**Algorithm:** Compare A-B-L to determine user vs. system changes  
**Logic:**
- If L == B: User hasn't modified, apply A
- If A == B: No system changes, preserve L  
- If A != B and L != B: Conflict, needs resolution  
**Advantage:** Distinguishes user changes from system updates

### Content-Based Diff
**Definition:** File comparison based on actual content, not timestamps  
**Benefit:** Reliable change detection regardless of file modification times  
**Implementation:** SHA-256 hashing for quick comparison, line-by-line for detailed diffs  
**Use case:** Determining preserve vs. apply decisions

## File Management

### Atomic Operations
**Definition:** Multi-file changes that either completely succeed or completely fail  
**Implementation:** Temporary staging, verification, then atomic move/rename  
**Benefit:** Never leaves system in inconsistent state  
**Recovery:** Automatic rollback on failure

### Backup Protection
**Definition:** Automatic creation of timestamped backups before destructive operations  
**Location:** `.wm/backups/YYYY-MM-DD_HH-MM-SS/`  
**Scope:** All files being modified or deleted  
**Override:** `--no-backup` flag skips backup creation  
**Retention:** Manual cleanup required

### File Ownership Boundaries
**Definition:** Strict rules about which commands can write to which spaces  
**Enforcement:** Guard system blocks unauthorized writes  
**Purpose:** Prevent accidental system file corruption  
**Examples:**
- User commands cannot write to Upstream space
- Only package manager can write to Baseline space
- Sandbox operations isolated from production spaces

## Command Categories

### Package Manager Commands
**Commands:** `config install`, `config update`, `config status`, `config migrate-legacy`  
**Purpose:** Manage system configuration distribution and updates  
**Pattern:** Follows package manager paradigm (install → update → manage)  
**Safety:** High (atomic, backed up, conflict detection)

### Development Commands
**Commands:** `dev sandbox`, `dev sandbox diff`  
**Purpose:** Safe experimentation and testing of configuration changes  
**Pattern:** Isolated environment → test → integrate  
**Safety:** Medium (isolated but can affect baseline)

### Guard Commands
**Commands:** `guard check`, `guard install-hook`  
**Purpose:** Validate and prevent unauthorized system modifications  
**Pattern:** Continuous validation and protection  
**Safety:** High (read-only validation, prevents corruption)

### Project Commands
**Commands:** `init`, `status`, `execute`  
**Purpose:** General project management and command execution  
**Pattern:** Traditional CLI tool workflow  
**Safety:** Variable (depends on specific command)

## Technical Terms

### Manifest
**Definition:** Metadata file describing configuration structure and versions  
**Location:** `manifest.json` in each space  
**Contents:** File checksums, version info, dependency data  
**Usage:** Validation, change detection, integrity verification

### Schema Validation
**Definition:** JSON/YAML structure validation against predefined schemas  
**Implementation:** Go struct tags, JSON Schema files  
**Trigger:** File read operations, startup validation  
**Benefit:** Early error detection, data consistency

### Hook System
**Definition:** Event-driven scripts that run during specific operations  
**Types:** Pre-commit hooks, post-update hooks, validation hooks  
**Location:** `.git/hooks/`, `internal/config/system/hooks/`  
**Purpose:** Automated quality control, continuous validation

### Working Tree
**Definition:** Git working directory state including staged and unstaged changes  
**Usage:** `guard check` analyzes working tree for violations  
**Scope:** All tracked and untracked files in Git repository  
**Commands:** `git status`, `git diff`, `git ls-files`

## Error Categories

### Validation Errors
**Definition:** Input data fails schema or business rule validation  
**Examples:** Invalid JSON, missing required fields, constraint violations  
**Handling:** Early detection, clear error messages, suggested fixes  
**Recovery:** User corrects input and retries

### Filesystem Errors
**Definition:** Operating system file/directory operation failures  
**Examples:** Permission denied, disk full, file not found  
**Handling:** System error translation, permission guidance  
**Recovery:** User fixes permissions or system issues

### Conflict Errors
**Definition:** 3-way merge cannot be automatically resolved  
**Examples:** Both user and system modified same lines differently  
**Handling:** Pause operation, show conflicts, await resolution  
**Recovery:** Manual editing, then retry operation

### Boundary Violations
**Definition:** Attempt to write files outside authorized space  
**Examples:** User command trying to modify Upstream files  
**Handling:** Guard system blocks operation, explains boundaries  
**Recovery:** Use appropriate command or space for the operation

## Process Workflows

### Installation Workflow
1. **Detection:** Check if Baseline exists
2. **Extraction:** Copy Upstream to Baseline
3. **Initialization:** Copy Baseline to Local
4. **Validation:** Verify all files copied correctly
5. **Completion:** Mark installation successful

### Update Workflow  
1. **Analysis:** Compare Upstream, Baseline, Local (3-way)
2. **Planning:** Determine preserve/apply/conflict/delete operations
3. **Preview:** Show planned changes (if --dry-run)
4. **Backup:** Create timestamped backup
5. **Execution:** Apply changes atomically
6. **Verification:** Validate final state

### Migration Workflow
1. **Discovery:** Locate legacy .claude-wm configuration
2. **Analysis:** Map legacy structure to new spaces
3. **Planning:** Determine transformation operations
4. **Backup:** Preserve legacy configuration
5. **Transform:** Convert to new structure
6. **Validation:** Verify migration correctness

## Quality Metrics

### Performance Targets
- **Config status**: <200ms for 100 files
- **3-way merge**: <500ms for 100 files  
- **Sandbox creation**: <1s for full upstream
- **Backup creation**: <2s for 10MB content

### Safety Measures
- **Atomic operations**: All-or-nothing guarantees
- **Backup protection**: Automatic before destructive operations
- **Validation chains**: Multiple layers of error checking
- **Boundary enforcement**: Strict file ownership rules