# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /templates:IMPLEMENTATION

**Rôle**
Documenteur statut implémentation avec features fonctionnelles, points intégration, APIs/interfaces et limitations connues.

**Contexte**
Template statut implémentation actuel avec documentation features working, points intégration, APIs/interfaces et limitations pour traçabilité complète.

**MCP à utiliser**
- **serena** : analyser implémentation actuelle codebase
- **mem0** : capitaliser sur patterns implémentation réussis

**Objectif**
Fournir template documentation statut implémentation pour traçabilité et communication équipe.

**Spécification détaillée**

# Current Implementation Status
## Working Features
## Integration Points
## API/Interfaces
## Known Limitations

**Bornes d'écriture**
* Autorisé : génération IMPLEMENTATION.md avec statut actuel complet
* Interdit : modification template sans validation implémentation

**Étapes**
1. [serena] Analyser implémentation actuelle codebase
2. Documenter features working et statut actuel
3. Identifier points intégration et APIs/interfaces
4. [mem0] Capitaliser sur patterns implémentation efficaces
5. Documenter limitations connues et impacts

**Points de vigilance**
- Features working documentées précisément
- Points intégration identifiés complètement
- APIs/interfaces documentées avec détails
- Limitations connues avec impacts évalués

**Tests/Validation**
- Vérification exactitude statut features
- Validation complétude documentation intégrations
- Test utilisabilité documentation équipe

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