## TECHNICAL.md
```markdown
# Technical Decisions Log

## Project-Level Decisions
> Note: This file exists at multiple levels (project/epic/task)

### Tech Stack
| Category | Choice | Rationale | Date |
|----------|--------|-----------|------|
| Language | {e.g., Python 3.11} | {Why} | {Date} |
| Framework | {e.g., FastAPI} | {Why} | {Date} |
| Database | {e.g., PostgreSQL} | {Why} | {Date} |
| Testing | {e.g., pytest} | {Why} | {Date} |

### Patterns & Conventions
| Pattern | Example | Rationale |
|---------|---------|-----------|
| {e.g., Repository pattern} | `UserRepository.get()` | {Why} |
| {e.g., Error handling} | Try/catch with custom exceptions | {Why} |

### Code Standards
- Formatting: {tool and config}
- Linting: {tool and rules}
- Documentation: {standards}
- Commit messages: {format}

## Rejected Alternatives
| Considered | Rejected Because | Date |
|------------|------------------|------|
| {Technology} | {Reason} | {Date} |

## Technical Debt
| Item | Impact | Priority | Plan |
|------|--------|----------|------|
| {Debt item} | {Business impact} | {P0-P3} | {Resolution plan} |

## Performance Targets
- Response time: < {X}ms
- Throughput: {X} requests/sec
- Memory usage: < {X}GB
- Database queries: < {X}ms