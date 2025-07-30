package patterns

import (
	"regexp"
	"sync"
	"testing"
)

// BenchmarkOldRegexCompilation benchmarks the old way of compiling patterns
func BenchmarkOldRegexCompilation(b *testing.B) {
	patterns := []string{
		`process\.env\.([A-Z_][A-Z0-9_]*)`,
		`os\.environ\.get\(["']([A-Z_][A-Z0-9_]*)["']\)`,
		`os\.Getenv\(["']([A-Z_][A-Z0-9_]*)["']\)`,
		`(?i)(api[_-]?key|secret[_-]?key|access[_-]?key|auth[_-]?key)`,
		`(?i)(token|password|secret|key|credential)`,
		`^[a-f0-9]{40}$`,
		`^[a-zA-Z0-9._/-]+$`,
		`export\s+async\s+function\s+(GET|POST|PUT|DELETE|PATCH)`,
		`(router|app)\.(get|post|put|delete|patch)\(['"]([^'"]+)['"]`,
		`(?i)(function|func|def|fn|proc|sub|method)\s+([a-zA-Z_][a-zA-Z0-9_]*)`,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pattern := range patterns {
			regexp.MustCompile(pattern)
		}
	}
}

// BenchmarkNewPreCompiledPatterns benchmarks the new pre-compiled patterns
func BenchmarkNewPreCompiledPatterns(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patterns := GetPatterns()
		_ = patterns.EnvJS
		_ = patterns.EnvPython
		_ = patterns.EnvGo
		_ = patterns.SecretAPI
		_ = patterns.SecretAuth
		_ = patterns.GitCommitHash
		_ = patterns.GitBranchName
		_ = patterns.APINextJS
		_ = patterns.APIExpress
		_ = patterns.QualityFunction
	}
}

// BenchmarkPatternMatching benchmarks pattern matching performance
func BenchmarkPatternMatching(b *testing.B) {
	patterns := GetPatterns()
	testStrings := []string{
		"process.env.API_KEY",
		"os.environ.get('DATABASE_URL')",
		"os.Getenv(\"PORT\")",
		"api_key",
		"password",
		"a1b2c3d4e5f6789012345678901234567890abcd",
		"feature/new-feature",
		"export async function GET",
		"router.get('/api/users', handler)",
		"function getName() {}",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, str := range testStrings {
			patterns.EnvJS.MatchString(str)
			patterns.EnvPython.MatchString(str)
			patterns.EnvGo.MatchString(str)
			patterns.SecretAPI.MatchString(str)
			patterns.SecretAuth.MatchString(str)
			patterns.GitCommitHash.MatchString(str)
			patterns.GitBranchName.MatchString(str)
			patterns.APINextJS.MatchString(str)
			patterns.APIExpress.MatchString(str)
			patterns.QualityFunction.MatchString(str)
		}
	}
}

// BenchmarkOldPatternMatching benchmarks old pattern matching performance
func BenchmarkOldPatternMatching(b *testing.B) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]*)`),
		regexp.MustCompile(`os\.environ\.get\(["']([A-Z_][A-Z0-9_]*)["']\)`),
		regexp.MustCompile(`os\.Getenv\(["']([A-Z_][A-Z0-9_]*)["']\)`),
		regexp.MustCompile(`(?i)(api[_-]?key|secret[_-]?key|access[_-]?key|auth[_-]?key)`),
		regexp.MustCompile(`(?i)(token|password|secret|key|credential)`),
		regexp.MustCompile(`^[a-f0-9]{40}$`),
		regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`),
		regexp.MustCompile(`export\s+async\s+function\s+(GET|POST|PUT|DELETE|PATCH)`),
		regexp.MustCompile(`(router|app)\.(get|post|put|delete|patch)\(['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)(function|func|def|fn|proc|sub|method)\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
	}
	
	testStrings := []string{
		"process.env.API_KEY",
		"os.environ.get('DATABASE_URL')",
		"os.Getenv(\"PORT\")",
		"api_key",
		"password",
		"a1b2c3d4e5f6789012345678901234567890abcd",
		"feature/new-feature",
		"export async function GET",
		"router.get('/api/users', handler)",
		"function getName() {}",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, str := range testStrings {
			for _, pattern := range patterns {
				pattern.MatchString(str)
			}
		}
	}
}

// BenchmarkSingletonAccess benchmarks singleton pattern access
func BenchmarkSingletonAccess(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetPatterns()
	}
}

// BenchmarkEnvPatternAccess benchmarks environment pattern access
func BenchmarkEnvPatternAccess(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetEnvPatterns()
	}
}

// BenchmarkPatternHelperFunctions benchmarks pattern helper functions
func BenchmarkPatternHelperFunctions(b *testing.B) {
	pattern := `test.*pattern`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompilePattern(pattern)
	}
}

// BenchmarkMustCompileHelperFunctions benchmarks MustCompile helper functions
func BenchmarkMustCompileHelperFunctions(b *testing.B) {
	pattern := `test.*pattern`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MustCompilePattern(pattern)
	}
}

// BenchmarkComplexPatternUsage benchmarks complex pattern usage simulation
func BenchmarkComplexPatternUsage(b *testing.B) {
	patterns := GetPatterns()
	
	// Simulate complex usage like in hooks
	testInputs := []string{
		"const API_KEY = process.env.API_KEY",
		"DATABASE_URL = os.environ.get('DATABASE_URL')",
		"port := os.Getenv(\"PORT\")",
		"export async function GET(request) { return new Response('Hello') }",
		"router.get('/api/users', (req, res) => { res.json(users) })",
		"@get('/api/items')",
		"function calculateTotal(items) { return items.reduce((sum, item) => sum + item.price, 0) }",
		"var userList = getUsersFromDatabase()",
		"class UserService { constructor() {} }",
		"interface UserRepository { findById(id: number): User }",
		"model User { id String @id @default(cuid()) }",
		"CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(255))",
		"SELECT * FROM users WHERE active = true",
		"cache_key = 'users:' + user.id",
		"// This is a comment explaining the function",
		"import React from 'react'",
		"from django.contrib.auth import authenticate",
		"use std::collections::HashMap",
		"require('./utils/database')",
		"#include <iostream>",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range testInputs {
			// Simulate typical hook processing
			patterns.EnvJS.MatchString(input)
			patterns.EnvPython.MatchString(input)
			patterns.EnvGo.MatchString(input)
			patterns.APINextJS.MatchString(input)
			patterns.APIExpress.MatchString(input)
			patterns.APIFastAPI.MatchString(input)
			patterns.QualityFunction.MatchString(input)
			patterns.QualityVariable.MatchString(input)
			patterns.QualityClass.MatchString(input)
			patterns.QualityInterface.MatchString(input)
			patterns.DBModel.MatchString(input)
			patterns.DBCreateTable.MatchString(input)
			patterns.DBQuery.MatchString(input)
			patterns.CacheKey.MatchString(input)
			patterns.QualityComment.MatchString(input)
			patterns.QualityImport.MatchString(input)
			patterns.SecretAPI.MatchString(input)
			patterns.SecretAuth.MatchString(input)
		}
	}
}

// BenchmarkMemoryUsage benchmarks memory usage of patterns
func BenchmarkMemoryUsage(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create multiple instances to test memory usage
		patterns1 := GetPatterns()
		patterns2 := GetPatterns()
		patterns3 := GetPatterns()
		
		// Ensure they're the same instance (singleton)
		if patterns1 != patterns2 || patterns2 != patterns3 {
			b.Fatal("Singleton pattern failed")
		}
	}
}

// BenchmarkConcurrentAccess benchmarks concurrent access to patterns
func BenchmarkConcurrentAccess(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			patterns := GetPatterns()
			patterns.EnvJS.MatchString("process.env.API_KEY")
			patterns.SecretAPI.MatchString("api_key")
			patterns.QualityFunction.MatchString("function test() {}")
		}
	})
}

// BenchmarkCompareOldVsNew compares old vs new approach directly
func BenchmarkCompareOldVsNew(b *testing.B) {
	testString := "process.env.API_KEY"
	
	b.Run("Old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Old approach: compile every time
			pattern := regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]*)`)
			pattern.MatchString(testString)
		}
	})
	
	b.Run("New", func(b *testing.B) {
		patterns := GetPatterns()
		for i := 0; i < b.N; i++ {
			// New approach: use pre-compiled pattern
			patterns.EnvJS.MatchString(testString)
		}
	})
}

// BenchmarkInitializationTime benchmarks initialization time
func BenchmarkInitializationTime(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset the singleton to force re-initialization
		instance = nil
		once = sync.Once{}
		GetPatterns()
	}
}