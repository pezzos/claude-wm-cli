package patterns

import (
	"regexp"
	"testing"
)

func TestPatternCompilation(t *testing.T) {
	patterns := GetPatterns()
	
	// Test that all patterns compile without errors
	tests := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		// Environment patterns
		{"EnvKeyValue", patterns.EnvKeyValue},
		{"EnvJS", patterns.EnvJS},
		{"EnvPython", patterns.EnvPython},
		{"EnvGo", patterns.EnvGo},
		{"EnvRust", patterns.EnvRust},
		{"EnvPHP", patterns.EnvPHP},
		{"EnvJava", patterns.EnvJava},
		{"EnvRuby", patterns.EnvRuby},
		{"EnvShell", patterns.EnvShell},
		{"EnvDocker", patterns.EnvDocker},
		{"EnvKubernetes", patterns.EnvKubernetes},
		{"EnvTerraform", patterns.EnvTerraform},
		{"EnvConfig", patterns.EnvConfig},
		{"EnvNext", patterns.EnvNext},
		{"EnvReact", patterns.EnvReact},
		{"EnvVue", patterns.EnvVue},
		{"EnvSvelte", patterns.EnvSvelte},
		{"EnvGeneral", patterns.EnvGeneral},
		
		// Security patterns
		{"SecretAPI", patterns.SecretAPI},
		{"SecretAuth", patterns.SecretAuth},
		{"SecretAWS", patterns.SecretAWS},
		{"SecretGCP", patterns.SecretGCP},
		{"SecretAzure", patterns.SecretAzure},
		{"SecretGeneral", patterns.SecretGeneral},
		
		// Git patterns
		{"GitCommitHash", patterns.GitCommitHash},
		{"GitBranchName", patterns.GitBranchName},
		{"GitTag", patterns.GitTag},
		{"GitRemoteURL", patterns.GitRemoteURL},
		{"GitMergeConflict", patterns.GitMergeConflict},
		{"GitIgnorePattern", patterns.GitIgnorePattern},
		{"GitDiffHunk", patterns.GitDiffHunk},
		
		// API patterns
		{"APINextJS", patterns.APINextJS},
		{"APIExpress", patterns.APIExpress},
		{"APIFastAPI", patterns.APIFastAPI},
		{"APIGin", patterns.APIGin},
		{"APIRoute", patterns.APIRoute},
		{"APIMethod", patterns.APIMethod},
		
		// File patterns
		{"FileCodeExt", patterns.FileCodeExt},
		{"FileConfigExt", patterns.FileConfigExt},
		{"FileDocExt", patterns.FileDocExt},
		{"FileTestExt", patterns.FileTestExt},
		{"FileAssetExt", patterns.FileAssetExt},
		{"FileIgnoreExt", patterns.FileIgnoreExt},
		
		// MCP patterns
		{"MCPToolCall", patterns.MCPToolCall},
		{"MCPFunction", patterns.MCPFunction},
		{"MCPParameter", patterns.MCPParameter},
		{"MCPImport", patterns.MCPImport},
		{"MCPExport", patterns.MCPExport},
		{"MCPAnthropicAPI", patterns.MCPAnthropicAPI},
		{"MCPToolDef", patterns.MCPToolDef},
		{"MCPToolUse", patterns.MCPToolUse},
		{"MCPToolResult", patterns.MCPToolResult},
		{"MCPToolError", patterns.MCPToolError},
		{"MCPToolMessage", patterns.MCPToolMessage},
		{"MCPToolRequest", patterns.MCPToolRequest},
		
		// Quality patterns
		{"QualityStyle", patterns.QualityStyle},
		{"QualityComment", patterns.QualityComment},
		{"QualityImport", patterns.QualityImport},
		{"QualityFunction", patterns.QualityFunction},
		{"QualityVariable", patterns.QualityVariable},
		{"QualityConstant", patterns.QualityConstant},
		{"QualityClass", patterns.QualityClass},
		{"QualityMethod", patterns.QualityMethod},
		{"QualityInterface", patterns.QualityInterface},
		{"QualityStruct", patterns.QualityStruct},
		
		// Database patterns
		{"DBModel", patterns.DBModel},
		{"DBCreateTable", patterns.DBCreateTable},
		{"DBQuery", patterns.DBQuery},
		{"DBConnection", patterns.DBConnection},
		{"DBMigration", patterns.DBMigration},
		
		// Cache patterns
		{"CacheKey", patterns.CacheKey},
		{"CacheInvalidate", patterns.CacheInvalidate},
		{"CacheExpire", patterns.CacheExpire},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.pattern == nil {
				t.Errorf("Pattern %s is nil", test.name)
			}
		})
	}
}

func TestEnvPatterns(t *testing.T) {
	patterns := GetEnvPatterns()
	
	testCases := []struct {
		name     string
		pattern  string
		input    string
		expected bool
	}{
		{"JS", "JS", "process.env.API_KEY", true},
		{"Python", "Python", "os.environ.get('DATABASE_URL')", true},
		{"Go", "Go", "os.Getenv(\"PORT\")", true},
		{"KeyValue", "KeyValue", "API_KEY=secret123", true},
		{"Shell", "Shell", "echo $DATABASE_URL", true},
		{"Docker", "Docker", "ENV NODE_ENV production", true},
		{"Next", "Next", "NEXT_PUBLIC_API_URL", true},
		{"React", "React", "REACT_APP_API_KEY", true},
		{"Vue", "Vue", "VUE_APP_BASE_URL", true},
		{"Svelte", "Svelte", "VITE_API_KEY", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern, exists := patterns[tc.pattern]
			if !exists {
				t.Fatalf("Pattern %s not found", tc.pattern)
			}
			
			matches := pattern.MatchString(tc.input)
			if matches != tc.expected {
				t.Errorf("Pattern %s with input %q: expected %v, got %v", tc.pattern, tc.input, tc.expected, matches)
			}
		})
	}
}

func TestSecurityPatterns(t *testing.T) {
	patterns := GetSecurityPatterns()
	
	testCases := []struct {
		name     string
		pattern  string
		input    string
		expected bool
	}{
		{"API", "API", "api_key", true},
		{"API", "API", "API-KEY", true},
		{"Auth", "Auth", "password", true},
		{"Auth", "Auth", "SECRET", true},
		{"AWS", "AWS", "aws_access_key", true},
		{"AWS", "AWS", "AWS-SECRET", true},
		{"GCP", "GCP", "gcp_key", true},
		{"GCP", "GCP", "GOOGLE_API_KEY", true},
		{"Azure", "Azure", "azure_key", true},
		{"Azure", "Azure", "SUBSCRIPTION_KEY", true},
		{"General", "General", "private_key", true},
		{"General", "General", "CLIENT-SECRET", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern, exists := patterns[tc.pattern]
			if !exists {
				t.Fatalf("Pattern %s not found", tc.pattern)
			}
			
			matches := pattern.MatchString(tc.input)
			if matches != tc.expected {
				t.Errorf("Pattern %s with input %q: expected %v, got %v", tc.pattern, tc.input, tc.expected, matches)
			}
		})
	}
}

func TestGitPatterns(t *testing.T) {
	patterns := GetGitPatterns()
	
	testCases := []struct {
		name     string
		pattern  string
		input    string
		expected bool
	}{
		{"CommitHash", "CommitHash", "a1b2c3d4e5f6789012345678901234567890abcd", true},
		{"CommitHash", "CommitHash", "invalid-hash", false},
		{"BranchName", "BranchName", "feature/new-feature", true},
		{"BranchName", "BranchName", "main", true},
		{"Tag", "Tag", "v1.0.0", true},
		{"Tag", "Tag", "1.2.3", true},
		{"Tag", "Tag", "v2.0.0-beta.1", true},
		{"MergeConflict", "MergeConflict", "<<<<<<<", true},
		{"MergeConflict", "MergeConflict", "=======", true},
		{"MergeConflict", "MergeConflict", ">>>>>>>", true},
		{"DiffHunk", "DiffHunk", "@@ -1,4 +1,4 @@", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern, exists := patterns[tc.pattern]
			if !exists {
				t.Fatalf("Pattern %s not found", tc.pattern)
			}
			
			matches := pattern.MatchString(tc.input)
			if matches != tc.expected {
				t.Errorf("Pattern %s with input %q: expected %v, got %v", tc.pattern, tc.input, tc.expected, matches)
			}
		})
	}
}

func TestAPIPatterns(t *testing.T) {
	patterns := GetAPIPatterns()
	
	testCases := []struct {
		name     string
		pattern  string
		input    string
		expected bool
	}{
		{"NextJS", "NextJS", "export async function GET", true},
		{"NextJS", "NextJS", "export async function POST", true},
		{"NextJS", "NextJS", "export async function PUT", true},
		{"NextJS", "NextJS", "export async function DELETE", true},
		{"NextJS", "NextJS", "export async function PATCH", true},
		{"Express", "Express", "router.get('/api/users', handler)", true},
		{"Express", "Express", "router.post(\"/api/users\", handler)", true},
		{"Express", "Express", "app.put('/api/users/:id', handler)", true},
		{"FastAPI", "FastAPI", "@get('/api/users')", true},
		{"FastAPI", "FastAPI", "@post(\"/api/users\")", true},
		{"Method", "Method", "GET", true},
		{"Method", "Method", "POST", true},
		{"Method", "Method", "PUT", true},
		{"Method", "Method", "DELETE", true},
		{"Method", "Method", "PATCH", true},
		{"Method", "Method", "HEAD", true},
		{"Method", "Method", "OPTIONS", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern, exists := patterns[tc.pattern]
			if !exists {
				t.Fatalf("Pattern %s not found", tc.pattern)
			}
			
			matches := pattern.MatchString(tc.input)
			if matches != tc.expected {
				t.Errorf("Pattern %s with input %q: expected %v, got %v", tc.pattern, tc.input, tc.expected, matches)
			}
		})
	}
}

func TestFilePatterns(t *testing.T) {
	patterns := GetFilePatterns()
	
	testCases := []struct {
		name     string
		pattern  string
		input    string
		expected bool
	}{
		{"CodeExt", "CodeExt", "test.js", true},
		{"CodeExt", "CodeExt", "test.py", true},
		{"CodeExt", "CodeExt", "test.go", true},
		{"CodeExt", "CodeExt", "test.rs", true},
		{"CodeExt", "CodeExt", "test.php", true},
		{"CodeExt", "CodeExt", "test.java", true},
		{"CodeExt", "CodeExt", "test.rb", true},
		{"CodeExt", "CodeExt", "test.txt", false},
		{"ConfigExt", "ConfigExt", "config.json", true},
		{"ConfigExt", "ConfigExt", "config.yaml", true},
		{"ConfigExt", "ConfigExt", "config.yml", true},
		{"ConfigExt", "ConfigExt", "config.toml", true},
		{"ConfigExt", "ConfigExt", "config.ini", true},
		{"ConfigExt", "ConfigExt", ".env", true},
		{"DocExt", "DocExt", "README.md", true},
		{"DocExt", "DocExt", "guide.txt", true},
		{"DocExt", "DocExt", "manual.pdf", true},
		{"TestExt", "TestExt", "test.spec.js", true},
		{"TestExt", "TestExt", "example.test.py", true},
		{"TestExt", "TestExt", "userTest.java", true},
		{"AssetExt", "AssetExt", "image.png", true},
		{"AssetExt", "AssetExt", "photo.jpg", true},
		{"AssetExt", "AssetExt", "icon.svg", true},
		{"AssetExt", "AssetExt", "style.css", true},
		{"AssetExt", "AssetExt", "font.woff", true},
		{"IgnoreExt", "IgnoreExt", "debug.log", true},
		{"IgnoreExt", "IgnoreExt", "temp.tmp", true},
		{"IgnoreExt", "IgnoreExt", "cache.cache", true},
		{"IgnoreExt", "IgnoreExt", "backup.bak", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern, exists := patterns[tc.pattern]
			if !exists {
				t.Fatalf("Pattern %s not found", tc.pattern)
			}
			
			matches := pattern.MatchString(tc.input)
			if matches != tc.expected {
				t.Errorf("Pattern %s with input %q: expected %v, got %v", tc.pattern, tc.input, tc.expected, matches)
			}
		})
	}
}

func TestMCPPatterns(t *testing.T) {
	patterns := GetMCPPatterns()
	
	testCases := []struct {
		name     string
		pattern  string
		input    string
		expected bool
	}{
		{"ToolCall", "ToolCall", "mcp_tool_call", true},
		{"ToolCall", "ToolCall", "MCP-TOOL-CALL", true},
		{"Function", "Function", "function_call", true},
		{"Function", "Function", "FUNCTION-CALL", true},
		{"Parameter", "Parameter", "parameter", true},
		{"Parameter", "Parameter", "PARAMETER", true},
		{"Import", "Import", "import mcp", true},
		{"Import", "Import", "IMPORT MCP", true},
		{"Export", "Export", "export mcp", true},
		{"Export", "Export", "EXPORT MCP", true},
		{"AnthropicAPI", "AnthropicAPI", "anthropic_api", true},
		{"AnthropicAPI", "AnthropicAPI", "ANTHROPIC-API", true},
		{"ToolDef", "ToolDef", "tool_def", true},
		{"ToolDef", "ToolDef", "TOOL-DEF", true},
		{"ToolUse", "ToolUse", "tool_use", true},
		{"ToolUse", "ToolUse", "TOOL-USE", true},
		{"ToolResult", "ToolResult", "tool_result", true},
		{"ToolResult", "ToolResult", "TOOL-RESULT", true},
		{"ToolError", "ToolError", "tool_error", true},
		{"ToolError", "ToolError", "TOOL-ERROR", true},
		{"ToolMessage", "ToolMessage", "tool_message", true},
		{"ToolMessage", "ToolMessage", "TOOL-MESSAGE", true},
		{"ToolRequest", "ToolRequest", "tool_request", true},
		{"ToolRequest", "ToolRequest", "TOOL-REQUEST", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern, exists := patterns[tc.pattern]
			if !exists {
				t.Fatalf("Pattern %s not found", tc.pattern)
			}
			
			matches := pattern.MatchString(tc.input)
			if matches != tc.expected {
				t.Errorf("Pattern %s with input %q: expected %v, got %v", tc.pattern, tc.input, tc.expected, matches)
			}
		})
	}
}

func TestQualityPatterns(t *testing.T) {
	patterns := GetQualityPatterns()
	
	testCases := []struct {
		name     string
		pattern  string
		input    string
		expected bool
	}{
		{"Style", "Style", "prettier", true},
		{"Style", "Style", "eslint", true},
		{"Style", "Style", "flake8", true},
		{"Style", "Style", "black", true},
		{"Style", "Style", "gofmt", true},
		{"Style", "Style", "rustfmt", true},
		{"Comment", "Comment", "// This is a comment", true},
		{"Comment", "Comment", "/* Block comment */", true},
		{"Comment", "Comment", "# Python comment", true},
		{"Comment", "Comment", "\"\"\"Docstring\"\"\"", true},
		{"Comment", "Comment", "<!-- HTML comment -->", true},
		{"Import", "Import", "import React from 'react'", true},
		{"Import", "Import", "from django import settings", true},
		{"Import", "Import", "use std::collections::HashMap", true},
		{"Import", "Import", "require('./utils')", true},
		{"Import", "Import", "include 'config.php'", true},
		{"Function", "Function", "function getName() {}", true},
		{"Function", "Function", "func GetName() string {}", true},
		{"Function", "Function", "def get_name():", true},
		{"Function", "Function", "fn get_name() -> String", true},
		{"Function", "Function", "proc getName() =", true},
		{"Function", "Function", "sub getName {}", true},
		{"Function", "Function", "method getName() {}", true},
		{"Variable", "Variable", "var name = 'John'", true},
		{"Variable", "Variable", "let age = 25", true},
		{"Variable", "Variable", "const PI = 3.14", true},
		{"Variable", "Variable", "dim username", true},
		{"Variable", "Variable", "my $variable", true},
		{"Variable", "Variable", "local result", true},
		{"Variable", "Variable", "global config", true},
		{"Constant", "Constant", "const MAX_SIZE = 100", true},
		{"Constant", "Constant", "final String NAME = \"test\"", true},
		{"Constant", "Constant", "static int COUNT = 0", true},
		{"Constant", "Constant", "readonly API_KEY = \"xyz\"", true},
		{"Class", "Class", "class User {}", true},
		{"Class", "Class", "interface UserInterface {}", true},
		{"Class", "Class", "struct UserStruct {}", true},
		{"Class", "Class", "enum UserType {}", true},
		{"Class", "Class", "type UserType = {}", true},
		{"Method", "Method", "public static getName() {}", true},
		{"Method", "Method", "private void setName() {}", true},
		{"Method", "Method", "protected String getName() {}", true},
		{"Method", "Method", "internal int getAge() {}", true},
		{"Interface", "Interface", "interface UserInterface {}", true},
		{"Interface", "Interface", "INTERFACE UserInterface {}", true},
		{"Struct", "Struct", "struct User {}", true},
		{"Struct", "Struct", "STRUCT User {}", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern, exists := patterns[tc.pattern]
			if !exists {
				t.Fatalf("Pattern %s not found", tc.pattern)
			}
			
			matches := pattern.MatchString(tc.input)
			if matches != tc.expected {
				t.Errorf("Pattern %s with input %q: expected %v, got %v", tc.pattern, tc.input, tc.expected, matches)
			}
		})
	}
}

func TestDBPatterns(t *testing.T) {
	patterns := GetDBPatterns()
	
	testCases := []struct {
		name     string
		pattern  string
		input    string
		expected bool
	}{
		{"Model", "Model", "model User {", true},
		{"Model", "Model", "MODEL User {", true},
		{"CreateTable", "CreateTable", "CREATE TABLE users", true},
		{"CreateTable", "CreateTable", "CREATE TABLE IF NOT EXISTS users", true},
		{"CreateTable", "CreateTable", "create table `users`", true},
		{"CreateTable", "CreateTable", "CREATE TABLE \"users\"", true},
		{"Query", "Query", "SELECT * FROM users", true},
		{"Query", "Query", "INSERT INTO users VALUES", true},
		{"Query", "Query", "UPDATE users SET", true},
		{"Query", "Query", "DELETE FROM users", true},
		{"Query", "Query", "CREATE INDEX", true},
		{"Query", "Query", "DROP TABLE", true},
		{"Query", "Query", "ALTER TABLE", true},
		{"Query", "Query", "TRUNCATE TABLE", true},
		{"Connection", "Connection", "connection string", true},
		{"Connection", "Connection", "connect to database", true},
		{"Connection", "Connection", "database connection", true},
		{"Connection", "Connection", "db connection", true},
		{"Connection", "Connection", "sql connection", true},
		{"Connection", "Connection", "nosql database", true},
		{"Connection", "Connection", "mongo connection", true},
		{"Connection", "Connection", "postgres connection", true},
		{"Connection", "Connection", "mysql connection", true},
		{"Connection", "Connection", "sqlite connection", true},
		{"Connection", "Connection", "redis connection", true},
		{"Connection", "Connection", "memcached connection", true},
		{"Migration", "Migration", "migration file", true},
		{"Migration", "Migration", "migrate database", true},
		{"Migration", "Migration", "schema migration", true},
		{"Migration", "Migration", "ddl migration", true},
		{"Migration", "Migration", "dml migration", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern, exists := patterns[tc.pattern]
			if !exists {
				t.Fatalf("Pattern %s not found", tc.pattern)
			}
			
			matches := pattern.MatchString(tc.input)
			if matches != tc.expected {
				t.Errorf("Pattern %s with input %q: expected %v, got %v", tc.pattern, tc.input, tc.expected, matches)
			}
		})
	}
}

func TestCachePatterns(t *testing.T) {
	patterns := GetCachePatterns()
	
	testCases := []struct {
		name     string
		pattern  string
		input    string
		expected bool
	}{
		{"Key", "Key", "cache_key", true},
		{"Key", "Key", "CACHE-KEY", true},
		{"Invalidate", "Invalidate", "invalidate cache", true},
		{"Invalidate", "Invalidate", "clear cache", true},
		{"Invalidate", "Invalidate", "flush cache", true},
		{"Invalidate", "Invalidate", "expire cache", true},
		{"Invalidate", "Invalidate", "evict cache", true},
		{"Expire", "Expire", "expire time", true},
		{"Expire", "Expire", "ttl value", true},
		{"Expire", "Expire", "timeout setting", true},
		{"Expire", "Expire", "duration limit", true},
		{"Expire", "Expire", "age limit", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern, exists := patterns[tc.pattern]
			if !exists {
				t.Fatalf("Pattern %s not found", tc.pattern)
			}
			
			matches := pattern.MatchString(tc.input)
			if matches != tc.expected {
				t.Errorf("Pattern %s with input %q: expected %v, got %v", tc.pattern, tc.input, tc.expected, matches)
			}
		})
	}
}

func TestSingletonBehavior(t *testing.T) {
	// Test that GetPatterns returns the same instance
	patterns1 := GetPatterns()
	patterns2 := GetPatterns()
	
	if patterns1 != patterns2 {
		t.Error("GetPatterns() should return the same instance (singleton)")
	}
}

func TestCompilePatternHelper(t *testing.T) {
	// Test the CompilePattern helper function
	pattern, err := CompilePattern(`^test.*$`)
	if err != nil {
		t.Fatalf("CompilePattern failed: %v", err)
	}
	
	if !pattern.MatchString("test123") {
		t.Error("Pattern should match 'test123'")
	}
	
	if pattern.MatchString("notest") {
		t.Error("Pattern should not match 'notest'")
	}
}

func TestMustCompilePatternHelper(t *testing.T) {
	// Test the MustCompilePattern helper function
	pattern := MustCompilePattern(`^test.*$`)
	
	if !pattern.MatchString("test123") {
		t.Error("Pattern should match 'test123'")
	}
	
	if pattern.MatchString("notest") {
		t.Error("Pattern should not match 'notest'")
	}
}

func TestMustCompilePatternPanic(t *testing.T) {
	// Test that MustCompilePattern panics on invalid regex
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustCompilePattern should panic on invalid regex")
		}
	}()
	
	MustCompilePattern(`[invalid regex`)
}