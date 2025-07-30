package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"strings"
)

// CacheKeyGenerator provides standardized cache key generation
type CacheKeyGenerator struct {
	prefix string
}

// NewCacheKeyGenerator creates a new cache key generator
func NewCacheKeyGenerator(prefix string) *CacheKeyGenerator {
	return &CacheKeyGenerator{prefix: prefix}
}

// Common cache key patterns for hooks

// GitStatusKey generates a cache key for git status
func (ckg *CacheKeyGenerator) GitStatusKey(workingDir string) string {
	hash := ckg.hashString(workingDir)
	return fmt.Sprintf("%s_git_status_%s", ckg.prefix, hash)
}

// GitDiffKey generates a cache key for git diff
func (ckg *CacheKeyGenerator) GitDiffKey(workingDir, commitHash string) string {
	combined := fmt.Sprintf("%s_%s", workingDir, commitHash)
	hash := ckg.hashString(combined)
	return fmt.Sprintf("%s_git_diff_%s", ckg.prefix, hash)
}

// FileHashKey generates a cache key for file content hash
func (ckg *CacheKeyGenerator) FileHashKey(filePath string) string {
	hash := ckg.hashString(filePath)
	return fmt.Sprintf("%s_file_hash_%s", ckg.prefix, hash)
}

// FileModTimeKey generates a cache key for file modification time
func (ckg *CacheKeyGenerator) FileModTimeKey(filePath string) string {
	hash := ckg.hashString(filePath)
	return fmt.Sprintf("%s_file_modtime_%s", ckg.prefix, hash)
}

// RegexPatternKey generates a cache key for compiled regex patterns
func (ckg *CacheKeyGenerator) RegexPatternKey(pattern, source string) string {
	combined := fmt.Sprintf("%s_%s", pattern, source)
	hash := ckg.hashString(combined)
	return fmt.Sprintf("%s_regex_%s", ckg.prefix, hash)
}

// ConfigParseKey generates a cache key for parsed configuration files
func (ckg *CacheKeyGenerator) ConfigParseKey(filePath string) string {
	hash := ckg.hashString(filePath)
	return fmt.Sprintf("%s_config_%s", ckg.prefix, hash)
}

// SecretPatternsKey generates a cache key for secret detection patterns
func (ckg *CacheKeyGenerator) SecretPatternsKey(rulesFile string) string {
	hash := ckg.hashString(rulesFile)
	return fmt.Sprintf("%s_secret_patterns_%s", ckg.prefix, hash)
}

// StyleRulesKey generates a cache key for style consistency rules
func (ckg *CacheKeyGenerator) StyleRulesKey(language, rulesFile string) string {
	combined := fmt.Sprintf("%s_%s", language, rulesFile)
	hash := ckg.hashString(combined)
	return fmt.Sprintf("%s_style_rules_%s", ckg.prefix, hash)
}

// APIEndpointsKey generates a cache key for API endpoint detection
func (ckg *CacheKeyGenerator) APIEndpointsKey(frameworkType string) string {
	hash := ckg.hashString(frameworkType)
	return fmt.Sprintf("%s_api_endpoints_%s", ckg.prefix, hash)
}

// DependencyKey generates a cache key for dependency analysis
func (ckg *CacheKeyGenerator) DependencyKey(packageFile string) string {
	hash := ckg.hashString(packageFile)
	return fmt.Sprintf("%s_dependencies_%s", ckg.prefix, hash)
}

// DuplicateAnalysisKey generates a cache key for duplicate detection
func (ckg *CacheKeyGenerator) DuplicateAnalysisKey(directory string, filePattern string) string {
	combined := fmt.Sprintf("%s_%s", directory, filePattern)
	hash := ckg.hashString(combined)
	return fmt.Sprintf("%s_duplicates_%s", ckg.prefix, hash)
}

// TimestampValidationKey generates a cache key for timestamp validation
func (ckg *CacheKeyGenerator) TimestampValidationKey(filePath string) string {
	hash := ckg.hashString(filePath)
	return fmt.Sprintf("%s_timestamp_%s", ckg.prefix, hash)
}

// EnvironmentSyncKey generates a cache key for environment synchronization
func (ckg *CacheKeyGenerator) EnvironmentSyncKey(envFiles []string) string {
	combined := strings.Join(envFiles, "|")
	hash := ckg.hashString(combined)
	return fmt.Sprintf("%s_env_sync_%s", ckg.prefix, hash)
}

// MCPToolsKey generates a cache key for MCP tool analysis
func (ckg *CacheKeyGenerator) MCPToolsKey(configFile string) string {
	hash := ckg.hashString(configFile)
	return fmt.Sprintf("%s_mcp_tools_%s", ckg.prefix, hash)
}

// DocumentationKey generates a cache key for documentation standards
func (ckg *CacheKeyGenerator) DocumentationKey(docType, language string) string {
	combined := fmt.Sprintf("%s_%s", docType, language)
	hash := ckg.hashString(combined)
	return fmt.Sprintf("%s_docs_%s", ckg.prefix, hash)
}

// DatabaseSchemaKey generates a cache key for database schema validation
func (ckg *CacheKeyGenerator) DatabaseSchemaKey(schemaFile string) string {
	hash := ckg.hashString(schemaFile)
	return fmt.Sprintf("%s_db_schema_%s", ckg.prefix, hash)
}

// Composite cache keys for complex operations

// FileAnalysisCompositeKey generates a composite key for complete file analysis
func (ckg *CacheKeyGenerator) FileAnalysisCompositeKey(filePath, analysisType string, options map[string]string) string {
	var parts []string
	parts = append(parts, filePath, analysisType)
	
	// Add sorted options
	for key, value := range options {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	
	combined := strings.Join(parts, "|")
	hash := ckg.hashString(combined)
	return fmt.Sprintf("%s_analysis_%s", ckg.prefix, hash)
}

// ProjectStateCompositeKey generates a composite key for project state
func (ckg *CacheKeyGenerator) ProjectStateCompositeKey(projectDir string, components []string) string {
	parts := append([]string{projectDir}, components...)
	combined := strings.Join(parts, "|")
	hash := ckg.hashString(combined)
	return fmt.Sprintf("%s_project_state_%s", ckg.prefix, hash)
}

// Utility functions for cache key validation and manipulation

// IsValidCacheKey checks if a cache key follows the expected pattern
func (ckg *CacheKeyGenerator) IsValidCacheKey(key string) bool {
	expectedPrefix := ckg.prefix + "_"
	return strings.HasPrefix(key, expectedPrefix) && len(key) > len(expectedPrefix)
}

// ExtractCacheType extracts the cache type from a cache key
func (ckg *CacheKeyGenerator) ExtractCacheType(key string) string {
	if !ckg.IsValidCacheKey(key) {
		return ""
	}
	
	parts := strings.Split(key, "_")
	if len(parts) >= 3 {
		return parts[1] // Type is the second part after prefix
	}
	
	return ""
}

// GenerateInvalidationPattern creates a pattern to invalidate related cache entries
func (ckg *CacheKeyGenerator) GenerateInvalidationPattern(baseKey string) string {
	parts := strings.Split(baseKey, "_")
	if len(parts) >= 2 {
		// Return pattern that matches all keys with same type
		return fmt.Sprintf("%s_%s_*", ckg.prefix, parts[1])
	}
	return baseKey
}

// Cache key categories for batch operations

// GetGitRelatedPatterns returns patterns for git-related cache keys
func (ckg *CacheKeyGenerator) GetGitRelatedPatterns() []string {
	return []string{
		fmt.Sprintf("%s_git_*", ckg.prefix),
		fmt.Sprintf("%s_file_*", ckg.prefix), // Files might be affected by git changes
	}
}

// GetFileRelatedPatterns returns patterns for file-related cache keys
func (ckg *CacheKeyGenerator) GetFileRelatedPatterns(filePath string) []string {
	hash := ckg.hashString(filePath)
	return []string{
		fmt.Sprintf("%s_file_hash_%s", ckg.prefix, hash),
		fmt.Sprintf("%s_file_modtime_%s", ckg.prefix, hash),
		fmt.Sprintf("%s_config_%s", ckg.prefix, hash),
		fmt.Sprintf("%s_analysis_%s", ckg.prefix, hash[:8]+"*"), // Partial match for composite keys
	}
}

// GetConfigRelatedPatterns returns patterns for configuration-related cache keys
func (ckg *CacheKeyGenerator) GetConfigRelatedPatterns() []string {
	return []string{
		fmt.Sprintf("%s_config_*", ckg.prefix),
		fmt.Sprintf("%s_dependencies_*", ckg.prefix),
		fmt.Sprintf("%s_env_sync_*", ckg.prefix),
		fmt.Sprintf("%s_mcp_tools_*", ckg.prefix),
	}
}

// Private helper methods

func (ckg *CacheKeyGenerator) hashString(input string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(input)))
}

// Hook-specific key generators

// SecretScannerKeys provides cache keys for secret-scanner.py
type SecretScannerKeys struct {
	*CacheKeyGenerator
}

func NewSecretScannerKeys() *SecretScannerKeys {
	return &SecretScannerKeys{NewCacheKeyGenerator("secret_scanner")}
}

func (ssk *SecretScannerKeys) SecretsInFileKey(filePath string) string {
	return ssk.FileAnalysisCompositeKey(filePath, "secrets", map[string]string{"type": "scan"})
}

func (ssk *SecretScannerKeys) SecretPatternsKey(rulesFile string) string {
	return ssk.CacheKeyGenerator.SecretPatternsKey(rulesFile)
}

// StyleConsistencyKeys provides cache keys for style-consistency.py
type StyleConsistencyKeys struct {
	*CacheKeyGenerator
}

func NewStyleConsistencyKeys() *StyleConsistencyKeys {
	return &StyleConsistencyKeys{NewCacheKeyGenerator("style_consistency")}
}

func (sck *StyleConsistencyKeys) StyleViolationsKey(filePath, language string) string {
	return sck.FileAnalysisCompositeKey(filePath, "style", map[string]string{"language": language})
}

func (sck *StyleConsistencyKeys) StyleRulesKey(language string) string {
	return sck.CacheKeyGenerator.StyleRulesKey(language, "default")
}

// GitValidatorKeys provides cache keys for git-comprehensive-validator.py
type GitValidatorKeys struct {
	*CacheKeyGenerator
}

func NewGitValidatorKeys() *GitValidatorKeys {
	return &GitValidatorKeys{NewCacheKeyGenerator("git_validator")}
}

func (gvk *GitValidatorKeys) RepoStateKey(repoPath string) string {
	return gvk.ProjectStateCompositeKey(repoPath, []string{"status", "branch", "commits"})
}

func (gvk *GitValidatorKeys) CommitValidationKey(commitHash string) string {
	return fmt.Sprintf("%s_commit_validation_%s", gvk.prefix, gvk.hashString(commitHash))
}

// Main CLI interface for testing cache keys
func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <generator-type> <operation> [args...]\n", os.Args[0])
		fmt.Printf("Generator types: general, secret-scanner, style-consistency, git-validator\n")
		fmt.Printf("Operations: generate, validate, extract-type\n")
		return
	}
	
	generatorType := os.Args[1]
	operation := os.Args[2]
	
	var ckg *CacheKeyGenerator
	
	switch generatorType {
	case "general":
		ckg = NewCacheKeyGenerator("general")
	case "secret-scanner":
		ckg = NewSecretScannerKeys().CacheKeyGenerator
	case "style-consistency":
		ckg = NewStyleConsistencyKeys().CacheKeyGenerator
	case "git-validator":
		ckg = NewGitValidatorKeys().CacheKeyGenerator
	default:
		fmt.Printf("Unknown generator type: %s\n", generatorType)
		return
	}
	
	switch operation {
	case "generate":
		if len(os.Args) < 4 {
			fmt.Printf("Usage: %s %s generate <key-type> [args...]\n", os.Args[0], generatorType)
			return
		}
		
		keyType := os.Args[3]
		switch keyType {
		case "git-status":
			if len(os.Args) < 5 {
				fmt.Printf("Usage: %s %s generate git-status <working-dir>\n", os.Args[0], generatorType)
				return
			}
			key := ckg.GitStatusKey(os.Args[4])
			fmt.Printf("Generated key: %s\n", key)
			
		case "file-hash":
			if len(os.Args) < 5 {
				fmt.Printf("Usage: %s %s generate file-hash <file-path>\n", os.Args[0], generatorType)
				return
			}
			key := ckg.FileHashKey(os.Args[4])
			fmt.Printf("Generated key: %s\n", key)
			
		default:
			fmt.Printf("Unknown key type: %s\n", keyType)
		}
		
	case "validate":
		if len(os.Args) < 4 {
			fmt.Printf("Usage: %s %s validate <cache-key>\n", os.Args[0], generatorType)
			return
		}
		
		key := os.Args[3]
		valid := ckg.IsValidCacheKey(key)
		fmt.Printf("Key '%s' is valid: %t\n", key, valid)
		
	case "extract-type":
		if len(os.Args) < 4 {
			fmt.Printf("Usage: %s %s extract-type <cache-key>\n", os.Args[0], generatorType)
			return
		}
		
		key := os.Args[3]
		keyType := ckg.ExtractCacheType(key)
		fmt.Printf("Key type: %s\n", keyType)
		
	default:
		fmt.Printf("Unknown operation: %s\n", operation)
	}
}