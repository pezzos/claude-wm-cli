# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /templates:PRD

**Rôle**
Rédacteur PRD épique avec executive summary, background, goals objectives, user stories et requirements complets.

**Contexte**
Template Product Requirements Document standardisé pour épiques avec structure complète executive summary, background, goals/objectives, user stories, requirements fonctionnels/non-fonctionnels.

**MCP à utiliser**
- **serena** : analyser épiques existantes pour patterns
- **mem0** : capitaliser sur templates PRD efficaces

**Objectif**
Fournir template PRD épique structuré pour définition complète requirements et objectives mesurables.

**Spécification détaillée**

# Product Requirements Document - {Epic Name}

## Executive Summary
{1-2 sentences describing what this epic delivers}

## Background
### Problem Statement
{What problem are we solving?}

### Current State
{How things work today}

### Desired State
{How we want things to work}

## Goals & Objectives
### Primary Goals
1. {Measurable goal}
2. {Measurable goal}

### Success Metrics
- KPI 1: {metric} - Target: {value}
- KPI 2: {metric} - Target: {value}

## User Stories
### Story 1: {Title}
**As a** {user type}  
**I want** {action}  
**So that** {benefit}

**Acceptance Criteria:**
- [ ] {Specific criterion}
- [ ] {Specific criterion}

### Story 2: {Title}
**As a** {user type}  
**I want** {action}  
**So that** {benefit}

**Acceptance Criteria:**
- [ ] {Specific criterion}
- [ ] {Specific criterion}

## Requirements
### Functional Requirements
- FR1: System shall {requirement}
- FR2: System shall {requirement}

### Non-Functional Requirements
- NFR1: Performance - {requirement}
- NFR2: Security - {requirement}
- NFR3: Usability - {requirement}

## Constraints & Dependencies
### Technical Constraints
- {Constraint}

### Business Constraints
- {Constraint}

### Dependencies
- Depends on: {Other epic/system}
- Blocks: {Other epic/system}

## Out of Scope
- {What we're NOT doing}
- {What we're NOT doing}

## Timeline
- Start: {date}
- Key milestones:
  - {Milestone}: {date}
  - {Milestone}: {date}
- Target completion: {date}

## Risks
| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| {Risk} | High/Med/Low | High/Med/Low | {Plan} |

**Bornes d'écriture**
* Autorisé : génération PRD.md épique avec structure et objectives complets
* Interdit : modification template sans validation requirements

**Étapes**
1. [serena] Analyser épiques existantes pour patterns PRD
2. Structurer PRD selon template avec objectives mesurables
3. [mem0] Capitaliser sur templates PRD efficaces
4. Valider complétude requirements et user stories

**Points de vigilance**
- Executive summary clair et concis pour épique
- Goals objectives mesurables avec KPIs appropriés
- User stories avec acceptance criteria spécifiques
- Requirements fonctionnels/non-fonctionnels complets
- Timeline réaliste avec milestones claires
- Risques identifiés avec plans mitigation

**Tests/Validation**
- Vérification complétude structure PRD
- Validation mesurabilité objectives définis
- Test clarté requirements pour équipe développement

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