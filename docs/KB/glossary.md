# Glossary

## Core Terms

### **Upstream**
The system templates and configurations embedded in the binary (`internal/config/system/`). Read-only, updated only through binary updates.

### **Baseline** 
Immutable snapshot stored in `.wm/baseline/` created during installation. Used as reference point for 3-way merges. Never modified after creation.

### **Local**
Runtime configuration in `.claude/` directory used by Claude Code. Generated from templates and user customizations. Updated by config commands.

### **Sandbox**
Isolated testing environment in `.wm/sandbox/claude/` for safe experimentation without affecting production configs.

### **3-Way Merge**
Algorithm comparing three states (upstream, baseline, local) to reconcile system updates with user changes.

## File System Terms

### **Workspace Root**
The `.wm/` directory containing all window management metadata, baselines, backups, and sandbox.

### **Meta.json**
Workspace metadata file (`.wm/meta.json`) tracking installation version, timestamps, and configuration state.

### **Atomic Operations**
File operations using temp+rename pattern to ensure consistency and prevent corruption during updates.

## Operational Terms

### **Config Install**
Initial setup command that creates baseline and local configurations from upstream templates.

### **Config Update**  
3-way merge operation to apply system changes while preserving user customizations.

### **Config Sync**
Regenerate local configuration from templates without merging system changes.

### **Dev Sandbox**
Development workflow for testing changes to system templates before upstreaming.

### **Migrate Legacy**
Migration process from old `.claude-wm/` structure to new `.wm/` workspace format.

## Command Categories

### **Config Commands**
Commands managing configuration spaces: `install`, `status`, `update`, `sync`, `upgrade`, `show`, `migrate-legacy`

### **Dev Commands**  
Development and testing commands: `sandbox`, `sandbox diff`

### **Guard Commands**
Validation and safety commands: `check`, `install-hook`