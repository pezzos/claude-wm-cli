# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /templates:TASK

**Rôle**
Structurateur tâche avec analyse approche, recherche mémoire similaire et documentation implémentation proposée.

**Contexte**
Template tâche standardisé avec structure observations, approche, recherche mémoire, raisonnement, implémentation proposée, diagramme Mermaid et changements fichiers.

**MCP à utiliser**
- **mem0** : recherche tâches similaires pour réutilisation patterns
- **serena** : analyse fichiers pour changements appropriés

**Objectif**
Fournir template tâche structuré pour analyse complète et documentation implémentation.

**Spécification détaillée**

## Task: {description}
### Observations
- Requirements, insights, warnings
### Approach
### Similar memory search
### Reasoning
### Proposed implementation
### Mermaid Diagram
### File changes
- {file}: {description of change}

**Bornes d'écriture**
* Autorisé : génération TASK.md avec structure et analyse complète
* Interdit : modification template sans validation approche

**Étapes**
1. [mem0] Rechercher tâches similaires pour patterns
2. Structurer tâche selon template avec analyse complète
3. [serena] Analyser fichiers pour changements appropriés
4. Documenter implémentation avec diagrammes Mermaid

**Points de vigilance**
- Observations complètes requirements et insights
- Recherche mémoire similaire systématique
- Raisonnement clair pour approche choisie
- Documentation changements fichiers précise

**Tests/Validation**
- Vérification complétude structure tâche
- Validation pertinence approche proposée
- Test clarté documentation implémentation

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