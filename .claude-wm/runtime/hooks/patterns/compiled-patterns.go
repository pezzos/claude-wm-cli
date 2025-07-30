package patterns

import (
	"regexp"
	"sync"
)

// CompiledPatterns holds all pre-compiled regex patterns
type CompiledPatterns struct {
	// Environment Variable Patterns
	EnvKeyValue     *regexp.Regexp
	EnvJS           *regexp.Regexp
	EnvPython       *regexp.Regexp
	EnvGo           *regexp.Regexp
	EnvRust         *regexp.Regexp
	EnvPHP          *regexp.Regexp
	EnvJava         *regexp.Regexp
	EnvRuby         *regexp.Regexp
	EnvShell        *regexp.Regexp
	EnvDocker       *regexp.Regexp
	EnvKubernetes   *regexp.Regexp
	EnvTerraform    *regexp.Regexp
	EnvConfig       *regexp.Regexp
	EnvNext         *regexp.Regexp
	EnvReact        *regexp.Regexp
	EnvVue          *regexp.Regexp
	EnvSvelte       *regexp.Regexp
	EnvGeneral      *regexp.Regexp

	// Security Patterns
	SecretAPI        *regexp.Regexp
	SecretAuth       *regexp.Regexp
	SecretAWS        *regexp.Regexp
	SecretGCP        *regexp.Regexp
	SecretAzure      *regexp.Regexp
	SecretGeneral    *regexp.Regexp

	// Git Patterns
	GitCommitHash    *regexp.Regexp
	GitBranchName    *regexp.Regexp
	GitTag           *regexp.Regexp
	GitRemoteURL     *regexp.Regexp
	GitMergeConflict *regexp.Regexp
	GitIgnorePattern *regexp.Regexp
	GitDiffHunk      *regexp.Regexp

	// API Route Patterns
	APINextJS        *regexp.Regexp
	APIExpress       *regexp.Regexp
	APIFastAPI       *regexp.Regexp
	APIGin           *regexp.Regexp
	APIRoute         *regexp.Regexp
	APIMethod        *regexp.Regexp

	// File Extension Patterns
	FileCodeExt      *regexp.Regexp
	FileConfigExt    *regexp.Regexp
	FileDocExt       *regexp.Regexp
	FileTestExt      *regexp.Regexp
	FileAssetExt     *regexp.Regexp
	FileIgnoreExt    *regexp.Regexp

	// MCP Tool Patterns
	MCPToolCall      *regexp.Regexp
	MCPFunction      *regexp.Regexp
	MCPParameter     *regexp.Regexp
	MCPImport        *regexp.Regexp
	MCPExport        *regexp.Regexp
	MCPAnthropicAPI  *regexp.Regexp
	MCPToolDef       *regexp.Regexp
	MCPToolUse       *regexp.Regexp
	MCPToolResult    *regexp.Regexp
	MCPToolError     *regexp.Regexp
	MCPToolMessage   *regexp.Regexp
	MCPToolRequest   *regexp.Regexp

	// Code Quality Patterns
	QualityStyle     *regexp.Regexp
	QualityComment   *regexp.Regexp
	QualityImport    *regexp.Regexp
	QualityFunction  *regexp.Regexp
	QualityVariable  *regexp.Regexp
	QualityConstant  *regexp.Regexp
	QualityClass     *regexp.Regexp
	QualityMethod    *regexp.Regexp
	QualityInterface *regexp.Regexp
	QualityStruct    *regexp.Regexp

	// Database Patterns
	DBModel          *regexp.Regexp
	DBCreateTable    *regexp.Regexp
	DBQuery          *regexp.Regexp
	DBConnection     *regexp.Regexp
	DBMigration      *regexp.Regexp

	// Cache Patterns
	CacheKey         *regexp.Regexp
	CacheInvalidate  *regexp.Regexp
	CacheExpire      *regexp.Regexp
}

var (
	instance *CompiledPatterns
	once     sync.Once
)

// GetPatterns returns the singleton instance of compiled patterns
func GetPatterns() *CompiledPatterns {
	once.Do(func() {
		instance = &CompiledPatterns{
			// Environment Variable Patterns
			EnvKeyValue:     regexp.MustCompile(`^([A-Z_][A-Z0-9_]*)\s*=\s*(.*)$`),
			EnvJS:           regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]*)`),
			EnvPython:       regexp.MustCompile(`os\.environ(?:\.get)?\(['"]([A-Z_][A-Z0-9_]*)['"]`),
			EnvGo:           regexp.MustCompile(`os\.Getenv\(['"]([A-Z_][A-Z0-9_]*)['"]`),
			EnvRust:         regexp.MustCompile(`env::var\(['"]([A-Z_][A-Z0-9_]*)['"]`),
			EnvPHP:          regexp.MustCompile(`\$_ENV\[['"]([A-Z_][A-Z0-9_]*)['"]`),
			EnvJava:         regexp.MustCompile(`System\.getenv\(['"]([A-Z_][A-Z0-9_]*)['"]`),
			EnvRuby:         regexp.MustCompile(`ENV\[['"]([A-Z_][A-Z0-9_]*)['"]`),
			EnvShell:        regexp.MustCompile(`\$\{?([A-Z_][A-Z0-9_]*)\}?`),
			EnvDocker:       regexp.MustCompile(`ENV\s+([A-Z_][A-Z0-9_]*)`),
			EnvKubernetes:   regexp.MustCompile(`env:\s*-\s*name:\s*([A-Z_][A-Z0-9_]*)`),
			EnvTerraform:    regexp.MustCompile(`var\.([A-Z_][A-Z0-9_]*)`),
			EnvConfig:       regexp.MustCompile(`config\.([A-Z_][A-Z0-9_]*)`),
			EnvNext:         regexp.MustCompile(`NEXT_PUBLIC_([A-Z_][A-Z0-9_]*)`),
			EnvReact:        regexp.MustCompile(`REACT_APP_([A-Z_][A-Z0-9_]*)`),
			EnvVue:          regexp.MustCompile(`VUE_APP_([A-Z_][A-Z0-9_]*)`),
			EnvSvelte:       regexp.MustCompile(`VITE_([A-Z_][A-Z0-9_]*)`),
			EnvGeneral:      regexp.MustCompile(`([A-Z_][A-Z0-9_]*)`),

			// Security Patterns
			SecretAPI:       regexp.MustCompile(`(?i)(api[_-]?key|secret[_-]?key|access[_-]?key|auth[_-]?key)`),
			SecretAuth:      regexp.MustCompile(`(?i)(token|password|secret|key|credential)`),
			SecretAWS:       regexp.MustCompile(`(?i)(aws[_-]?access[_-]?key|aws[_-]?secret)`),
			SecretGCP:       regexp.MustCompile(`(?i)(gcp[_-]?key|google[_-]?api[_-]?key|service[_-]?account)`),
			SecretAzure:     regexp.MustCompile(`(?i)(azure[_-]?key|azure[_-]?secret|subscription[_-]?key)`),
			SecretGeneral:   regexp.MustCompile(`(?i)(private[_-]?key|client[_-]?secret|bearer[_-]?token)`),

			// Git Patterns
			GitCommitHash:    regexp.MustCompile(`^[a-f0-9]{40}$`),
			GitBranchName:    regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`),
			GitTag:           regexp.MustCompile(`^v?[0-9]+(\.[0-9]+)*(-[a-zA-Z0-9.-]+)?$`),
			GitRemoteURL:     regexp.MustCompile(`^(https?://|git@)[^/]+/[^/]+/[^/]+(?:\.git)?$`),
			GitMergeConflict: regexp.MustCompile(`^<{7}|^={7}|^>{7}`),
			GitIgnorePattern: regexp.MustCompile(`^[^#\n][^\n]*$`),
			GitDiffHunk:      regexp.MustCompile(`^@@\s+-(\d+)(?:,(\d+))?\s+\+(\d+)(?:,(\d+))?\s+@@`),

			// API Route Patterns
			APINextJS:        regexp.MustCompile(`export\s+async\s+function\s+(GET|POST|PUT|DELETE|PATCH)`),
			APIExpress:       regexp.MustCompile(`(router|app)\.(get|post|put|delete|patch)\(['"]([^'"]+)['"]`),
			APIFastAPI:       regexp.MustCompile(`@(get|post|put|delete|patch)\(['"]([^'"]+)['"]`),
			APIGin:           regexp.MustCompile(`router\.(GET|POST|PUT|DELETE|PATCH)\(['"]([^'"]+)['"]`),
			APIRoute:         regexp.MustCompile(`/app(/api/[^/]+(?:/[^/]+)*)`),
			APIMethod:        regexp.MustCompile(`(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)`),

			// File Extension Patterns
			FileCodeExt:      regexp.MustCompile(`\.(js|jsx|ts|tsx|py|go|rs|php|java|rb|c|cpp|h|hpp|cs|kt|swift|dart|scala|clj|hs|ml|fs|erl|ex|lua|pl|r|jl|nim|zig|odin|v|d|cr|pas|ada|cob|for|f90|asm|s|sh|bash|zsh|fish|ps1|bat|cmd|vbs|applescript|awk|sed|m|mm|pyx|pxd|pxi|pth|whl|egg|jar|class|dex|apk|ipa|exe|dll|so|dylib|a|lib|o|obj|pyc|pyo|pyd|gem|deb|rpm|dmg|pkg|msi|tar|zip|rar|7z|gz|bz2|xz|lz4|zst|br|lzma|lzo|rz|sz|Z|uu|xxe|b64|hex|bin|dat|db|sqlite|sqlite3|mdb|accdb|dbf|csv|tsv|json|xml|yaml|yml|toml|ini|cfg|conf|config|properties|env|plist|reg|bak|tmp|temp|cache|log|out|err|pid|lock|sock|fifo|pipe|link|alias|lnk|url|webloc|desktop|service|timer|socket|mount|automount|swap|device|path|target|slice|scope|busname|snapshot|network|netdev|link|dnssd|template|wants|requires|requisite|binds|part|conflict|before|after|on|install|user|group|passwd|shadow|gshadow|hosts|hostname|resolv|fstab|mtab|proc|sys|dev|run|var|tmp|boot|home|root|usr|bin|sbin|lib|lib32|lib64|libx32|opt|srv|mnt|media|cdrom|floppy|zip|rar|7z|tar|gz|bz2|xz|lz4|zst|br|lzma|lzo|rz|sz|Z|uu|xxe|b64|hex|bin|dat)$`),
			FileConfigExt:    regexp.MustCompile(`\.(json|yaml|yml|toml|ini|cfg|conf|config|properties|env|plist|reg)$`),
			FileDocExt:       regexp.MustCompile(`\.(md|txt|rst|adoc|tex|pdf|doc|docx|odt|rtf|html|htm|xml)$`),
			FileTestExt:      regexp.MustCompile(`\.(test|spec)\..*$|.*\.(test|spec)$|.*_test\..*$|.*Test\..*$`),
			FileAssetExt:     regexp.MustCompile(`\.(png|jpg|jpeg|gif|svg|webp|ico|bmp|tiff|tif|webm|mp4|avi|mov|wmv|flv|mkv|mp3|wav|flac|ogg|aac|wma|m4a|css|scss|sass|less|styl|woff|woff2|ttf|otf|eot)$`),
			FileIgnoreExt:    regexp.MustCompile(`\.(log|tmp|temp|cache|bak|swp|swo|orig|rej|pyc|pyo|pyd|class|o|obj|exe|dll|so|dylib|a|lib|lock|pid|sock|fifo|pipe)$`),

			// MCP Tool Patterns
			MCPToolCall:      regexp.MustCompile(`(?i)mcp[_-]?tool[_-]?call`),
			MCPFunction:      regexp.MustCompile(`(?i)function[_-]?call`),
			MCPParameter:     regexp.MustCompile(`(?i)parameter`),
			MCPImport:        regexp.MustCompile(`(?i)import.*mcp`),
			MCPExport:        regexp.MustCompile(`(?i)export.*mcp`),
			MCPAnthropicAPI:  regexp.MustCompile(`(?i)anthropic[_-]?api`),
			MCPToolDef:       regexp.MustCompile(`(?i)tool[_-]?def`),
			MCPToolUse:       regexp.MustCompile(`(?i)tool[_-]?use`),
			MCPToolResult:    regexp.MustCompile(`(?i)tool[_-]?result`),
			MCPToolError:     regexp.MustCompile(`(?i)tool[_-]?error`),
			MCPToolMessage:   regexp.MustCompile(`(?i)tool[_-]?message`),
			MCPToolRequest:   regexp.MustCompile(`(?i)tool[_-]?request`),

			// Code Quality Patterns
			QualityStyle:     regexp.MustCompile(`(?i)(style|format|lint|prettier|eslint|flake8|black|gofmt|rustfmt)`),
			QualityComment:   regexp.MustCompile(`^\s*//|^\s*/\*|^\s*\*|^\s*#|^\s*"""|(<!--.*-->)`),
			QualityImport:    regexp.MustCompile(`(?i)(import|from|use|require|include)[\s\(]+`),
			QualityFunction:  regexp.MustCompile(`(?i)(function|func|def|fn|proc|sub|method)\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
			QualityVariable:  regexp.MustCompile(`(?i)(var|let|const|dim|my|local|global)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`),
			QualityConstant:  regexp.MustCompile(`(?i)(const|final|static|readonly)\s+([A-Z_][A-Z0-9_]*)`),
			QualityClass:     regexp.MustCompile(`(?i)(class|interface|struct|enum|type)\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
			QualityMethod:    regexp.MustCompile(`(?i)(public|private|protected|internal)\s+(static\s+)?([a-zA-Z_][a-zA-Z0-9_]*\s+)?([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`),
			QualityInterface: regexp.MustCompile(`(?i)interface\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
			QualityStruct:    regexp.MustCompile(`(?i)struct\s+([a-zA-Z_][a-zA-Z0-9_]*)`),

			// Database Patterns
			DBModel:          regexp.MustCompile(`(?i)model\s+(\w+)\s*\{`),
			DBCreateTable:    regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?[` + "`" + `"]?(\w+)[` + "`" + `"]?`),
			DBQuery:          regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE|CREATE|DROP|ALTER|TRUNCATE)\s+`),
			DBConnection:     regexp.MustCompile(`(?i)(connection|connect|database|db|sql|nosql|mongo|postgres|mysql|sqlite|redis|memcached)`),
			DBMigration:      regexp.MustCompile(`(?i)(migration|migrate|schema|ddl|dml)`),

			// Cache Patterns
			CacheKey:         regexp.MustCompile(`(?i)cache[_-]?key`),
			CacheInvalidate:  regexp.MustCompile(`(?i)(invalidate|clear|flush|expire|evict)`),
			CacheExpire:      regexp.MustCompile(`(?i)(expire|ttl|timeout|duration|age)`),
		}
	})
	return instance
}

// GetEnvPatterns returns all environment variable patterns
func GetEnvPatterns() map[string]*regexp.Regexp {
	patterns := GetPatterns()
	return map[string]*regexp.Regexp{
		"KeyValue":   patterns.EnvKeyValue,
		"JS":         patterns.EnvJS,
		"Python":     patterns.EnvPython,
		"Go":         patterns.EnvGo,
		"Rust":       patterns.EnvRust,
		"PHP":        patterns.EnvPHP,
		"Java":       patterns.EnvJava,
		"Ruby":       patterns.EnvRuby,
		"Shell":      patterns.EnvShell,
		"Docker":     patterns.EnvDocker,
		"Kubernetes": patterns.EnvKubernetes,
		"Terraform":  patterns.EnvTerraform,
		"Config":     patterns.EnvConfig,
		"Next":       patterns.EnvNext,
		"React":      patterns.EnvReact,
		"Vue":        patterns.EnvVue,
		"Svelte":     patterns.EnvSvelte,
		"General":    patterns.EnvGeneral,
	}
}

// GetSecurityPatterns returns all security patterns
func GetSecurityPatterns() map[string]*regexp.Regexp {
	patterns := GetPatterns()
	return map[string]*regexp.Regexp{
		"API":     patterns.SecretAPI,
		"Auth":    patterns.SecretAuth,
		"AWS":     patterns.SecretAWS,
		"GCP":     patterns.SecretGCP,
		"Azure":   patterns.SecretAzure,
		"General": patterns.SecretGeneral,
	}
}

// GetGitPatterns returns all git patterns
func GetGitPatterns() map[string]*regexp.Regexp {
	patterns := GetPatterns()
	return map[string]*regexp.Regexp{
		"CommitHash":    patterns.GitCommitHash,
		"BranchName":    patterns.GitBranchName,
		"Tag":           patterns.GitTag,
		"RemoteURL":     patterns.GitRemoteURL,
		"MergeConflict": patterns.GitMergeConflict,
		"IgnorePattern": patterns.GitIgnorePattern,
		"DiffHunk":      patterns.GitDiffHunk,
	}
}

// GetAPIPatterns returns all API patterns
func GetAPIPatterns() map[string]*regexp.Regexp {
	patterns := GetPatterns()
	return map[string]*regexp.Regexp{
		"NextJS":  patterns.APINextJS,
		"Express": patterns.APIExpress,
		"FastAPI": patterns.APIFastAPI,
		"Gin":     patterns.APIGin,
		"Route":   patterns.APIRoute,
		"Method":  patterns.APIMethod,
	}
}

// GetFilePatterns returns all file patterns
func GetFilePatterns() map[string]*regexp.Regexp {
	patterns := GetPatterns()
	return map[string]*regexp.Regexp{
		"CodeExt":   patterns.FileCodeExt,
		"ConfigExt": patterns.FileConfigExt,
		"DocExt":    patterns.FileDocExt,
		"TestExt":   patterns.FileTestExt,
		"AssetExt":  patterns.FileAssetExt,
		"IgnoreExt": patterns.FileIgnoreExt,
	}
}

// GetMCPPatterns returns all MCP patterns
func GetMCPPatterns() map[string]*regexp.Regexp {
	patterns := GetPatterns()
	return map[string]*regexp.Regexp{
		"ToolCall":      patterns.MCPToolCall,
		"Function":      patterns.MCPFunction,
		"Parameter":     patterns.MCPParameter,
		"Import":        patterns.MCPImport,
		"Export":        patterns.MCPExport,
		"AnthropicAPI":  patterns.MCPAnthropicAPI,
		"ToolDef":       patterns.MCPToolDef,
		"ToolUse":       patterns.MCPToolUse,
		"ToolResult":    patterns.MCPToolResult,
		"ToolError":     patterns.MCPToolError,
		"ToolMessage":   patterns.MCPToolMessage,
		"ToolRequest":   patterns.MCPToolRequest,
	}
}

// GetQualityPatterns returns all code quality patterns
func GetQualityPatterns() map[string]*regexp.Regexp {
	patterns := GetPatterns()
	return map[string]*regexp.Regexp{
		"Style":     patterns.QualityStyle,
		"Comment":   patterns.QualityComment,
		"Import":    patterns.QualityImport,
		"Function":  patterns.QualityFunction,
		"Variable":  patterns.QualityVariable,
		"Constant":  patterns.QualityConstant,
		"Class":     patterns.QualityClass,
		"Method":    patterns.QualityMethod,
		"Interface": patterns.QualityInterface,
		"Struct":    patterns.QualityStruct,
	}
}

// GetDBPatterns returns all database patterns
func GetDBPatterns() map[string]*regexp.Regexp {
	patterns := GetPatterns()
	return map[string]*regexp.Regexp{
		"Model":       patterns.DBModel,
		"CreateTable": patterns.DBCreateTable,
		"Query":       patterns.DBQuery,
		"Connection":  patterns.DBConnection,
		"Migration":   patterns.DBMigration,
	}
}

// GetCachePatterns returns all cache patterns
func GetCachePatterns() map[string]*regexp.Regexp {
	patterns := GetPatterns()
	return map[string]*regexp.Regexp{
		"Key":         patterns.CacheKey,
		"Invalidate":  patterns.CacheInvalidate,
		"Expire":      patterns.CacheExpire,
	}
}

// CompilePattern compiles a single pattern with lazy loading
func CompilePattern(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}

// MustCompilePattern compiles a single pattern with panic on error
func MustCompilePattern(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}