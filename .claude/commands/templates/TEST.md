# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /templates:TEST

**Rôle**
Concepteur tests validation avec stratégies manuelles, automatisées et critères succès mesurables.

**Contexte**
Template tests validation standardisé avec tests manuels, automatisés et critères succès pour validation complète fonctionnalités.

**MCP à utiliser**
- **mem0** : réutiliser patterns tests efficaces
- **playwright/puppeteer** : tests automatisés UI si nécessaire

**Objectif**
Fournir template tests validation complet pour assurance qualité fonctionnalités.

**Spécification détaillée**

## Validation Tests
### Manual Test
1. Verification step
2. Expected result
### Automated Test
```code```
### Success Criteria
- Criterion 1

**Bornes d'écriture**
* Autorisé : génération TEST.md avec stratégies validation complètes
* Interdit : modification template sans validation approche tests

**Étapes**
1. [mem0] Rechercher patterns tests similaires efficaces
2. Structurer tests selon template avec manuel/automatisé
3. Définir critères succès mesurables
4. [playwright/puppeteer] Ajouter tests UI si approprié

**Points de vigilance**
- Tests manuels étapes claires et reproductibles
- Tests automatisés code fonctionnel et maintenable
- Critères succès mesurables et vérifiables
- Couverture test appropriée fonctionnalité

**Tests/Validation**
- Vérification reproductibilité tests manuels
- Validation fonctionnement tests automatisés
- Test mesurabilité critères succès

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