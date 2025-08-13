# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /templates:CLAUDE

**Rôle**
Gestionnaire contexte projet Claude avec documentation patterns architecturaux et standards développement.

**Contexte**
Template contexte projet standardisé pour configuration Claude avec patterns architecture, standards coding, environnement développement et standards qualité. Structure template pour substitution variables contexte projet.

**MCP à utiliser**
- **serena** : analyser patterns projet existants
- **mem0** : mémoriser contextes projet efficaces

**Objectif**
Fournir template contexte projet standardisé avec substitution variables pour configuration Claude cohérente.

**Spécification détaillée**

# Project Context for Claude

## Project Overview
{{PROJECT_NAME}} - {{PROJECT_DESCRIPTION}}

## Architecture Patterns
{{ARCHITECTURE_PATTERNS}}

## Coding Standards
{{CODING_STANDARDS}}

## Common Commands
{{COMMON_COMMANDS}}

## Lessons Learned
{{LESSONS_LEARNED}}

## Development Environment
{{DEV_ENVIRONMENT}}

## Quality Standards
{{QUALITY_STANDARDS}}

**Bornes d'écriture**
* Autorisé : génération CLAUDE.md avec substitution variables projet
* Interdit : modification template sans validation contexte

**Étapes**
1. [serena] Analyser patterns projet existants
2. Appliquer substitution variables contexte appropriées
3. [mem0] Mémoriser patterns contexte projet efficaces
4. Valider cohérence configuration Claude générée

**Points de vigilance**
- Substitution complète variables {{PLACEHOLDER}}
- Cohérence patterns architecture documentés
- Standards coding et qualité appropriés
- Contexte projet actualisé

**Tests/Validation**
- Vérification substitution variables complète
- Validation cohérence contexte projet
- Test utilisation configuration Claude

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