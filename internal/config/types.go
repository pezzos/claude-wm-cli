package config

// Configuration represents the runtime configuration structure
type Configuration struct {
	Version     string                 `json:"version"`
	Hooks       HooksConfig           `json:"hooks"`
	Permissions PermissionsConfig     `json:"permissions"`
	Env         map[string]string     `json:"env"`
}

// HooksConfig defines hook configuration
type HooksConfig struct {
	PreToolUse  []HookConfig `json:"PreToolUse"`
	PostToolUse []HookConfig `json:"PostToolUse"`
}

// HookConfig defines a single hook configuration
type HookConfig struct {
	Matcher string   `json:"matcher"`
	Hooks   []string `json:"hooks"`
}

// PermissionsConfig defines security permissions
type PermissionsConfig struct {
	AllowedTools []string `json:"allowed_tools"`
}

// PackageInfo contains metadata about the configuration package
type PackageInfo struct {
	Version     string `json:"version"`
	LastUpdated string `json:"last_updated"`
	Source      string `json:"source"`
}