# Claude WM CLI

An intelligent command-line interface that acts as a wrapper for Claude commands, providing both interactive and headless operational modes for sophisticated project workflow management.

## Project Structure

```
docs/
├── 1-project/          # Global project vision and roadmap
├── 2-current-epic/     # Current epic execution
├── 3-current-task/     # Current task breakdown
└── archive/            # Completed epics backup
```

## Features

- **Dual Operation Modes**: Interactive terminal usage and headless VSCode extension integration
- **Intelligent Command Wrapper**: Seamlessly interfaces with Claude commands from `$HOME/.claude/commands`
- **Hierarchical Workflow Management**: PROJECT → EPIC → STORY → TICKET progression
- **Context-Aware Suggestions**: Analyzes project state to recommend appropriate next actions
- **Integrated Git Workflows**: Automatic branch management and structured commits
- **MCP-Powered Intelligence**: Leverages Model Context Protocol for enhanced AI assistance

## Command Structure

Commands follow a hierarchical path-based structure:
`/{category}/{subcategory}/{command-name}` → `/{category}:{subcategory}:{command-name}`

### Core Workflow Commands

#### Project Level (`/1-project:*`)
- **Init**: `/1-project:1-start:1-Init-Project` - Initialize project structure
- **Update**: `/1-project:2-update:*` - Import feedback, challenge docs, enrich context
- **Epics**: `/1-project:3-epics:*` - Plan and manage epic roadmap

#### Epic Level (`/2-epic:*`)
- **Start**: `/2-epic:1-start:*` - Select and plan epic stories
- **Manage**: `/2-epic:2-manage:*` - Track progress and complete epics

#### Story Level (`/3-story:*`)
- **Manage**: `/3-story:1-manage:*` - Start stories and extract technical tasks

#### Ticket Level (`/4-ticket:*`)
- **Create**: `/4-ticket:1-start:*` - Generate tickets from stories, issues, or input
- **Execute**: `/4-ticket:2-execute:*` - 5-phase implementation process
- **Complete**: `/4-ticket:3-complete:*` - Archive and update status

#### Support Tools
- **DEBUG**: `/debug:*` - Project health monitoring and repair
- **ENRICH**: `/enrich:*` - Context enhancement and pattern discovery
- **METRICS**: `/metrics:*` - Performance tracking and analytics
- **LEARNING**: `/learning:*` - Pattern recognition and optimization
- **VALIDATION**: `/validation:*` - Architecture review and quality assurance

## Workflow Architecture

```mermaid
---
config:
  layout: dagre
---
flowchart TD
 subgraph TOOLS["Support Tools - Available Anytime"]
        TOOL_DEBUG["DEBUG<br>Check/Fix"]
        TOOL_ENRICH["ENRICH<br>Context"]
        TOOL_METRICS["METRICS<br>Track"]
  end
    P_START{"PROJECT"} --> P_INIT["1-Init-Project<br>Creates docs structure"]
    P_INIT --> P_UPDATE{"Project Update Cycle"} & E_MANAGE{"Epic Management"}
    P_UPDATE --> P_UPD1["1-Import-feedback<br>Read FEEDBACK.md"]
    P_UPD1 --> P_UPD2["2-Challenge<br>Challenge docs"]
    P_UPD2 --> P_UPD3["3-Enrich<br>Add context"]
    P_UPD3 --> P_UPD4["4-Status<br>Update status"]
    P_UPD4 --> P_UPD5["5-Implementation-Status<br>Review progress"]
    P_UPD5 -.-> P_UPD1
    P_EPICS["1-Plan-Epics<br>Create EPICS.md"] --> E_CYCLE{"Epic Cycle"}
    E_CYCLE --> E_SELECT["1-Select-Stories<br>Choose epic & create PRD"]
    E_SELECT --> E_PLAN["2-Plan-stories<br>Create STORIES.md"]
    E_MANAGE --> P_EPICS
    E_COMPLETE["1-Complete-Epic<br>Archive epic"] --> E_CLEAR["Clear Context"]
    E_CLEAR -- More epics --> E_SELECT
    E_CLEAR -- No more epics --> P_END["Project Complete"]
    E_STATUS["2-Status-Epic<br>Check progress"] --> E_CYCLE
    E_PLAN --> S_CYCLE{"Story Cycle"}
    S_CYCLE -- Has tasks --> S_START["1-Start-Story<br>Create TODO.md"]
    S_CYCLE -- All done --> E_COMPLETE
    S_START --> T_CYCLE{"Ticket Cycle"}
    S_COMPLETE["2-Complete-Story<br>Mark story done"] --> S_CYCLE
    T_CYCLE -- From story --> T_SRC1["1-From-story"]
    T_GITHUB["GitHub"] -- Issue --> T_SRC2["2-From-issue"]
    T_USER["User"] -- Input --> T_SRC3["3-From-input"]
    T_SRC1 --> T_EXEC{"Execute Cycle"}
    T_SRC2 --> T_EXEC
    T_SRC3 --> T_EXEC
    T_EXEC --> T_PLAN["1-Plan-Ticket<br>Implementation plan"]
    T_PLAN --> T_TEST["2-Test-design<br>Test strategy"]
    T_TEST --> T_IMPL["3-Implement<br>Code & test"]
    T_IMPL --> T_VALID["4-Validate-Ticket<br>Check criteria"]
    T_VALID -- Fail < 3 times --> T_PLAN
    T_VALID -- Success --> T_REVIEW["5-Review-Ticket<br>Final review"]
    T_REVIEW --> T_ARCHIVE["1-Archive-Ticket"]
    T_ARCHIVE --> T_STATUS["2-Status-Ticket"]
    T_STATUS --> T_CLEAR["Clear Context"]
    T_CLEAR -- More tickets --> T_SRC1
    T_CLEAR -- Story done --> S_COMPLETE
```

## Quick Start

### Initial Setup
1. **Initialize Project**: `/1-project:1-start:1-Init-Project`
2. **Import Feedback**: `/1-project:2-update:1-Import-feedback` (if FEEDBACK.md exists)
3. **Plan Epics**: `/1-project:3-epics:1-Plan-Epics`

### Development Cycle
1. **Select Epic**: `/2-epic:1-start:1-Select-Stories`
2. **Plan Stories**: `/2-epic:1-start:2-Plan-stories`
3. **Start Story**: `/3-story:1-manage:1-Start-Story`
4. **Execute Tickets**: `/4-ticket:1-start:1-From-story` → `/4-ticket:2-execute:*`
5. **Complete & Archive**: `/4-ticket:3-complete:*` → `/3-story:1-manage:2-Complete-Story`

### Context-Aware Operation
The CLI analyzes your project state and suggests appropriate next actions based on:
- Presence of `.claude-wm/state.json`
- Existing documentation structure
- Current workflow position
- Outstanding tasks and dependencies

## Development

Project is managed using Claude Code's agile workflow system with automatic role profiles and task management.

## License

[Add your license here]