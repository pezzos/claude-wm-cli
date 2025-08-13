#!/usr/bin/env node

/**
 * Claude Code Output Schema Validator
 * Validates JSON output against the standardized schema
 */

const fs = require('fs');
const path = require('path');

// JSON Schema for Claude Code outputs
const OUTPUT_SCHEMA = {
  type: "object",
  required: ["plan", "changes", "patches", "summary", "notes"],
  properties: {
    plan: { 
      type: "string",
      minLength: 10,
      description: "Sequential steps executed in this task"
    },
    changes: {
      type: "array",
      description: "List of file changes made",
      items: {
        type: "object",
        required: ["path", "action", "content"],
        properties: {
          path: { 
            type: "string",
            minLength: 1,
            description: "Relative file path from project root"
          },
          action: { 
            type: "string", 
            enum: ["create", "update", "delete", "none"],
            description: "Action performed on the file"
          },
          content: { 
            type: "string",
            minLength: 5,
            description: "Brief description of changes made"
          }
        }
      }
    },
    patches: {
      type: "array",
      description: "Unified diff patches for each changed file",
      items: {
        type: "object",
        required: ["path", "diff"],
        properties: {
          path: { 
            type: "string",
            minLength: 1,
            description: "Relative file path from project root"
          },
          diff: { 
            type: "string",
            description: "Unified diff or empty for create/delete"
          }
        }
      }
    },
    summary: { 
      type: "string",
      minLength: 20,
      maxLength: 500,
      description: "5-line max TL;DR with file stats (#files, new/mod/del)"
    },
    notes: { 
      type: "string",
      minLength: 10,
      description: "Gotchas encountered, TODOs, limitations"
    }
  }
};

/**
 * Simple JSON Schema validator
 */
function validateSchema(data, schema, path = '') {
  const errors = [];
  
  if (schema.type === 'object') {
    if (typeof data !== 'object' || data === null || Array.isArray(data)) {
      errors.push(`${path}: Expected object, got ${typeof data}`);
      return errors;
    }
    
    // Check required fields
    if (schema.required) {
      for (const field of schema.required) {
        if (!(field in data)) {
          errors.push(`${path}.${field}: Required field missing`);
        }
      }
    }
    
    // Validate properties
    if (schema.properties) {
      for (const [key, subSchema] of Object.entries(schema.properties)) {
        if (key in data) {
          errors.push(...validateSchema(data[key], subSchema, `${path}.${key}`));
        }
      }
    }
  }
  
  else if (schema.type === 'array') {
    if (!Array.isArray(data)) {
      errors.push(`${path}: Expected array, got ${typeof data}`);
      return errors;
    }
    
    if (schema.items) {
      data.forEach((item, index) => {
        errors.push(...validateSchema(item, schema.items, `${path}[${index}]`));
      });
    }
  }
  
  else if (schema.type === 'string') {
    if (typeof data !== 'string') {
      errors.push(`${path}: Expected string, got ${typeof data}`);
      return errors;
    }
    
    if (schema.minLength && data.length < schema.minLength) {
      errors.push(`${path}: String too short (${data.length} < ${schema.minLength})`);
    }
    
    if (schema.maxLength && data.length > schema.maxLength) {
      errors.push(`${path}: String too long (${data.length} > ${schema.maxLength})`);
    }
    
    if (schema.enum && !schema.enum.includes(data)) {
      errors.push(`${path}: Invalid enum value '${data}', expected one of: ${schema.enum.join(', ')}`);
    }
  }
  
  return errors;
}

/**
 * Validate file statistics in summary
 */
function validateSummary(summary) {
  const errors = [];
  const lines = summary.split('\n');
  
  if (lines.length > 5) {
    errors.push('Summary exceeds 5 lines maximum');
  }
  
  // Check for file statistics pattern
  const hasFileStats = summary.match(/Files:\s*\d+\s*total/i);
  if (!hasFileStats) {
    errors.push('Summary missing file statistics (Files: X total ...)');
  }
  
  return errors;
}

/**
 * Validate changes and patches consistency
 */
function validateConsistency(output) {
  const errors = [];
  const changesPaths = new Set(output.changes.map(c => c.path));
  const patchesPaths = new Set(output.patches.map(p => p.path));
  
  // Check that all changed files have patches
  for (const change of output.changes) {
    if (change.action !== 'none' && !patchesPaths.has(change.path)) {
      errors.push(`Missing patch for changed file: ${change.path}`);
    }
  }
  
  // Check that all patches correspond to changes
  for (const patch of output.patches) {
    if (!changesPaths.has(patch.path)) {
      errors.push(`Patch provided for unchanged file: ${patch.path}`);
    }
  }
  
  // Check diff content for create/delete actions
  for (const patch of output.patches) {
    const change = output.changes.find(c => c.path === patch.path);
    if (change) {
      if ((change.action === 'create' || change.action === 'delete') && 
          patch.diff && !patch.diff.includes('no diff for')) {
        errors.push(`${patch.path}: Expected no diff for ${change.action} action`);
      }
    }
  }
  
  return errors;
}

/**
 * Main validation function
 */
function validateOutput(jsonString) {
  const results = {
    valid: false,
    errors: [],
    warnings: []
  };
  
  try {
    // Parse JSON
    const output = JSON.parse(jsonString);
    
    // Schema validation
    const schemaErrors = validateSchema(output, OUTPUT_SCHEMA);
    results.errors.push(...schemaErrors);
    
    if (schemaErrors.length === 0) {
      // Content validation
      const summaryErrors = validateSummary(output.summary);
      results.errors.push(...summaryErrors);
      
      const consistencyErrors = validateConsistency(output);
      results.errors.push(...consistencyErrors);
      
      // Warnings for best practices
      if (output.plan.split('\n').length < 2) {
        results.warnings.push('Plan seems very brief - consider more detailed steps');
      }
      
      if (output.changes.length > 10) {
        results.warnings.push('Large number of file changes - consider breaking into smaller tasks');
      }
      
      if (!output.summary.includes('Files:')) {
        results.warnings.push('Summary should include file statistics');
      }
    }
    
    results.valid = results.errors.length === 0;
    
  } catch (error) {
    results.errors.push(`JSON parsing error: ${error.message}`);
  }
  
  return results;
}

/**
 * CLI Interface
 */
function main() {
  const args = process.argv.slice(2);
  
  if (args.length === 0) {
    console.log('Usage: validate-output.js <json-file> [--verbose]');
    console.log('       validate-output.js --stdin [--verbose]');
    process.exit(1);
  }
  
  const verbose = args.includes('--verbose');
  let jsonInput;
  
  if (args.includes('--stdin')) {
    // Read from stdin
    let input = '';
    process.stdin.setEncoding('utf8');
    process.stdin.on('data', chunk => input += chunk);
    process.stdin.on('end', () => {
      validateAndOutput(input, verbose);
    });
  } else {
    // Read from file
    const filename = args[0];
    if (!fs.existsSync(filename)) {
      console.error(`Error: File '${filename}' not found`);
      process.exit(1);
    }
    
    jsonInput = fs.readFileSync(filename, 'utf8');
    validateAndOutput(jsonInput, verbose);
  }
}

function validateAndOutput(jsonInput, verbose) {
  const results = validateOutput(jsonInput);
  
  if (results.valid) {
    console.log('✅ Output validation passed');
    
    if (results.warnings.length > 0) {
      console.log('\n⚠️  Warnings:');
      results.warnings.forEach(warning => console.log(`  - ${warning}`));
    }
    
    if (verbose) {
      const output = JSON.parse(jsonInput);
      console.log(`\n📊 Summary:`);
      console.log(`  - Plan steps: ${output.plan.split('\n').length}`);
      console.log(`  - Files changed: ${output.changes.length}`);
      console.log(`  - Patches provided: ${output.patches.length}`);
      console.log(`  - Summary length: ${output.summary.length} chars`);
      console.log(`  - Notes length: ${output.notes.length} chars`);
    }
    
    process.exit(0);
  } else {
    console.log('❌ Output validation failed');
    console.log('\n🚫 Errors:');
    results.errors.forEach(error => console.log(`  - ${error}`));
    
    if (results.warnings.length > 0) {
      console.log('\n⚠️  Warnings:');
      results.warnings.forEach(warning => console.log(`  - ${warning}`));
    }
    
    process.exit(1);
  }
}

// Run CLI if called directly
if (require.main === module) {
  main();
}

module.exports = { validateOutput, OUTPUT_SCHEMA };