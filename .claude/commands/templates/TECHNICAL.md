# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /templates:TECHNICAL

**Rôle**
Documenteur décisions techniques avec log décisions, alternatives rejetées, dette technique et targets performance.

**Contexte**
Template documentation technique complète avec décisions projet/épique/tâche, patterns, standards code, alternatives rejetées, dette technique et targets performance.

**MCP à utiliser**
- **serena** : analyser décisions techniques existantes
- **context7** : référencer ADR et standards techniques
- **mem0** : capitaliser sur décisions techniques efficaces
- **time** : dater décisions pour traçabilité

**Objectif**
Fournir template documentation technique structuré pour traçabilité décisions et standards projet.

**Spécification détaillée**

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
```

**Bornes d'écriture**
* Autorisé : génération TECHNICAL.md avec documentation décisions complète
* Interdit : modification template sans traçabilité décisions

**Étapes**
1. [serena] Analyser décisions techniques existantes projet
2. [context7] Référencer ADR et standards techniques appropriés
3. [time] Dater décisions pour traçabilité historique
4. Structurer documentation selon template multi-niveaux
5. [mem0] Capitaliser sur décisions techniques efficaces

**Points de vigilance**
- Documentation décisions projet/épique/tâche appropriée
- Traçabilité historique décisions avec dates
- Alternatives rejetées documentées avec rationale
- Dette technique priorisée avec plans résolution
- Targets performance mesurables et réalistes

**Tests/Validation**
- Vérification complétude documentation décisions
- Validation traçabilité historique appropriée
- Test utilisabilité documentation pour équipe

**Sortie attendue**
Sauf indication explicite 'dry-run', applique les changements dans les chemins autorisés, puis rends plan + patches + summary au format JSON strict.

## Schéma JSON de sortie

```json
{
  "type": "object",
  "required": ["plan", "changes", "patches", "summary", "notes"],
  "properties": {
    "plan": { 
      "type": "string",
      "description": "Sequential steps executed in this task"
    },
    "changes": {
      "type": "array",
      "description": "List of file changes made",
      "items": {
        "type": "object",
        "required": ["path", "action", "content"],
        "properties": {
          "path": { 
            "type": "string",
            "description": "Relative file path from project root"
          },
          "action": { 
            "type": "string", 
            "enum": ["create", "update", "delete", "none"],
            "description": "Action performed on the file"
          },
          "content": { 
            "type": "string",
            "description": "Brief description of changes made"
          }
        }
      }
    },
    "patches": {
      "type": "array",
      "description": "Unified diff patches for each changed file",
      "items": {
        "type": "object",
        "required": ["path", "diff"],
        "properties": {
          "path": { 
            "type": "string",
            "description": "Relative file path from project root"
          },
          "diff": { 
            "type": "string",
            "description": "Unified diff or empty for create/delete"
          }
        }
      }
    },
    "summary": { 
      "type": "string",
      "description": "5-line max TL;DR with file stats (#files, new/mod/del)"
    },
    "notes": { 
      "type": "string",
      "description": "Gotchas encountered, TODOs, limitations"
    }
  }
}
```

## Exit Codes
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed