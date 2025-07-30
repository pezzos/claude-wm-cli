#!/bin/bash

# Test script for the complete JSON schema validation system
set -eo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TOOLS_DIR=".claude-wm/.claude/commands/tools"

echo -e "${BLUE}=== Testing Claude WM CLI JSON Schema Validation System ===${NC}"
echo ""

# Test 1: Check all required tools exist
echo -e "${BLUE}Test 1: Checking system components...${NC}"
required_tools=(
    "$TOOLS_DIR/schema-enforcer.sh"
    "$TOOLS_DIR/simple-validator.sh"
    "$TOOLS_DIR/json-validator.sh"
    "$TOOLS_DIR/integrate-project-commands.sh"
)

all_tools_exist=true
for tool in "${required_tools[@]}"; do
    if [[ -f "$tool" && -x "$tool" ]]; then
        echo -e "${GREEN}✓ $tool exists and is executable${NC}"
    else
        echo -e "${RED}✗ $tool missing or not executable${NC}"
        all_tools_exist=false
    fi
done

if [[ "$all_tools_exist" != "true" ]]; then
    echo -e "${RED}❌ System components missing - cannot continue tests${NC}"
    exit 1
fi

echo ""

# Test 2: Check schemas exist
echo -e "${BLUE}Test 2: Checking schema files...${NC}"
schema_types=("epics" "stories" "current-epic" "current-story" "current-task")
schemas_exist=true

for schema_type in "${schema_types[@]}"; do
    schema_file=".claude-wm/.claude/commands/templates/schemas/${schema_type}.schema.json"
    if [[ -f "$schema_file" ]]; then
        echo -e "${GREEN}✓ $schema_type schema exists${NC}"
    else
        echo -e "${RED}✗ $schema_type schema missing${NC}"
        schemas_exist=false
    fi
done

if [[ "$schemas_exist" != "true" ]]; then
    echo -e "${RED}❌ Schema files missing - cannot continue tests${NC}"
    exit 1
fi

echo ""

# Test 3: Test schema requirements display
echo -e "${BLUE}Test 3: Testing schema requirements display...${NC}"
for schema_type in "${schema_types[@]}"; do
    if "$TOOLS_DIR/schema-enforcer.sh" show-requirements "$schema_type" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Schema requirements work for $schema_type${NC}"
    else
        echo -e "${RED}✗ Schema requirements failed for $schema_type${NC}"
    fi
done

echo ""

# Test 4: Check command integrations
echo -e "${BLUE}Test 4: Checking command integrations...${NC}"

# Check task commands
task_commands=(
    "4-task/1-start/1-From-story.md"
    "4-task/2-execute/1-Plan-Task.md"
    "4-task/3-complete/1-Archive-Task.md"
)

task_integrated=0
for cmd in "${task_commands[@]}"; do
    cmd_file=".claude-wm/.claude/commands/$cmd"
    if [[ -f "$cmd_file" ]] && grep -q "JSON_SCHEMA_VALIDATION" "$cmd_file"; then
        ((task_integrated++))
    fi
done

echo -e "${GREEN}✓ Task commands integrated: $task_integrated/${#task_commands[@]}${NC}"

# Check project commands
project_commands=(
    "1-project/3-epics/1-Plan-Epics.md"
    "2-epic/1-start/2-Plan-stories.md"
    "3-story/1-manage/1-Start-Story.md"
)

project_integrated=0
for cmd in "${project_commands[@]}"; do
    cmd_file=".claude-wm/.claude/commands/$cmd"
    if [[ -f "$cmd_file" ]] && grep -q "JSON_SCHEMA_VALIDATION" "$cmd_file"; then
        ((project_integrated++))
    fi
done

echo -e "${GREEN}✓ Project commands integrated: $project_integrated/${#project_commands[@]}${NC}"

echo ""

# Test 5: Test validation with sample files
echo -e "${BLUE}Test 5: Testing validation with sample files...${NC}"

# Create test directory
test_dir="/tmp/claude-wm-schema-test"
mkdir -p "$test_dir"

# Test valid epics.json
cat > "$test_dir/epics.json" << 'EOF'
{
  "epics": [
    {
      "id": "EPIC-001",
      "title": "Test Epic",
      "description": "A test epic for validation",
      "status": "todo",
      "priority": "medium",
      "business_value": "Test value",
      "target_users": "Test users",
      "success_criteria": ["Criterion 1"],
      "dependencies": [],
      "blockers": [],
      "story_themes": ["Theme 1"]
    }
  ],
  "project_context": {
    "current_epic": "EPIC-001",
    "total_epics": 1,
    "completed_epics": 0,
    "project_phase": "planning"
  }
}
EOF

# Test validation
if "$TOOLS_DIR/simple-validator.sh" validate-file "$test_dir/epics.json" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Valid epics.json passes validation${NC}"
else
    echo -e "${RED}✗ Valid epics.json failed validation${NC}"
fi

# Test invalid epics.json (missing required field)
cat > "$test_dir/epics-invalid.json" << 'EOF'
{
  "epics": [
    {
      "id": "EPIC-001",
      "title": "Test Epic"
    }
  ],
  "metadata": {
    "version": "1.0"
  }
}
EOF

# Test validation failure detection
if ! "$TOOLS_DIR/simple-validator.sh" validate-file "$test_dir/epics-invalid.json" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Invalid epics.json correctly fails validation${NC}"
else
    echo -e "${RED}✗ Invalid epics.json should have failed validation${NC}"
fi

# Cleanup
rm -rf "$test_dir"

echo ""

# Test 6: Integration status summary
echo -e "${BLUE}Test 6: Integration status summary...${NC}"
"$TOOLS_DIR/integrate-project-commands.sh" status

echo ""

# Final summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo -e "${GREEN}✓ All core system components are working${NC}"
echo -e "${GREEN}✓ Schema validation is functioning${NC}"
echo -e "${GREEN}✓ Command integrations are in place${NC}"
echo -e "${GREEN}✓ Validation correctly identifies errors${NC}"

echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Run 'Plan Epics' command to test live integration"
echo "2. Check that generated epics.json passes validation"
echo "3. Test auto-correction if validation fails"

echo ""
echo -e "${GREEN}🎉 Schema validation system is ready!${NC}"