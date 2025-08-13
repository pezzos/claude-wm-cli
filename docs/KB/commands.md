# Commands Reference

## Config Commands

### `claude-wm config install`
**Purpose**: Initial setup - creates baseline and local configurations  
**Writes**: `.claude/`, `.wm/baseline/`, `.wm/meta.json`  
**Reads**: `internal/config/system/` (embedded)  
**Flags**: `--force` (overwrite existing)  

### `claude-wm config status`
**Purpose**: Show configuration state and differences  
**Writes**: None (read-only)  
**Reads**: All configuration spaces  
**Flags**: `--verbose`, `--json`

### `claude-wm config update`
**Purpose**: 3-way merge to apply system updates  
**Writes**: `.claude/`, `.wm/backups/`  
**Reads**: All configuration spaces  
**Flags**: `--dry-run`, `--force`, `--backup`

### `claude-wm config sync`
**Purpose**: Regenerate local config from templates  
**Writes**: `.claude/`  
**Reads**: `.wm/baseline/`  
**Flags**: `--dry-run`

### `claude-wm config show`
**Purpose**: Display configuration values  
**Writes**: None (read-only)  
**Reads**: Specified config files  
**Args**: `<config-path>`

### `claude-wm config migrate-legacy`
**Purpose**: Migrate from `.claude-wm/` to `.wm/`  
**Writes**: `.wm/`  
**Reads**: `.claude-wm/`  
**Flags**: `--dry-run`, `--archive`, `--force`

## Dev Commands

### `claude-wm dev sandbox`
**Purpose**: Create isolated development environment  
**Writes**: `.wm/sandbox/claude/`  
**Reads**: `internal/config/system/`  
**Flags**: `--clean` (reset sandbox)

### `claude-wm dev sandbox diff`
**Purpose**: Compare sandbox with source  
**Writes**: `internal/config/system/` (with `--apply`)  
**Reads**: `.wm/sandbox/claude/`, `internal/config/system/`  
**Flags**: `--apply`, `--pattern <regex>`

## Guard Commands

### `claude-wm guard check`
**Purpose**: Validate current configuration state  
**Writes**: None (read-only)  
**Reads**: All configuration files  
**Flags**: `--fix` (auto-repair minor issues)

### `claude-wm guard install-hook`
**Purpose**: Install pre-commit validation hook  
**Writes**: `.git/hooks/pre-commit`  
**Reads**: Hook template  
**Flags**: `--force` (overwrite existing)

## Global Flags

All commands support:
- `--verbose` / `-v`: Detailed output
- `--quiet` / `-q`: Minimal output  
- `--config <path>`: Custom config file
- `--workspace <path>`: Custom workspace root

## Exit Codes

- `0`: Success
- `1`: General error
- `2`: Validation error
- `3`: Conflict requiring manual resolution
- `4`: Permission denied
- `5`: File not found or inaccessible