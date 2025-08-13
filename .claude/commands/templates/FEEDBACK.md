# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /templates:FEEDBACK

**Rôle**
Collecteur feedback utilisateur structuré avec questionnaire challengeant documentation et capture informations nouvelles.

**Contexte**
Template feedback utilisateur standardisé avec champs libres, questions architecture/techniques/scope et capture nouvelles informations features/contraintes.

**MCP à utiliser**
- **time** : dater feedback pour traçabilité
- **mem0** : mémoriser patterns feedback récurrents

**Objectif**
Fournir template feedback utilisateur structuré pour collecte information complète et amélioration continue.

**Spécification détaillée**

# Feedback - {Date}

## Free field for simple user input

## Questions raised from the docs challenging
### Architecture
- Q: {Question about architecture}
  A: {User response}

### Technical Choices
- Q: {Question about tech stack}
  A: {User response}

### Scope & Requirements
- Q: {Clarification needed}
  A: {User response}

## New Information
### Features
- {New feature requirement}
- {Changed requirement}

### Constraints
- {New technical constraint}
- {Business constraint}

**Bornes d'écriture**
* Autorisé : génération FEEDBACK.md avec date et structure appropriée
* Interdit : modification template sans validation utilisateur

**Étapes**
1. [time] Dater feedback pour traçabilité
2. Structurer feedback selon template standardisé
3. [mem0] Mémoriser patterns feedback efficaces
4. Valider complétude informations collectées

**Points de vigilance**
- Date feedback pour traçabilité temporelle
- Structure questions challengeant documentation
- Capture complète nouvelles informations
- Distinction features vs contraintes

**Tests/Validation**
- Vérification structure feedback complète
- Validation questions pertinentes posées
- Test utilisabilité informations collectées

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