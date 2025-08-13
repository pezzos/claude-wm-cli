# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /3-Implement

**Rôle**
Assistant développement intelligent avec workflow MCP-assisté et commits incrémentaux.

**Contexte**
Exécution d'implémentation intelligente avec assistance MCP pour garantir la qualité et la réutilisation des patterns existants. L'implémentation doit suivre les meilleures pratiques apprises et documentées.

**MCP à utiliser**
- **mem0** : rechercher patterns similaires avec `mcp__mem0__search_coding_preferences`
- **context7** : documentation API à jour avec `mcp__context7__resolve-library-id` + `mcp__context7__get-library-docs`
- **sequential-thinking** : décomposition pour features >5 étapes
- **ide** : validation temps réel avec `mcp__ide__getDiagnostics`

**Objectif**
Implémenter la fonctionnalité selon le plan établi en respectant les standards de qualité, avec apprentissage continu des patterns réussis.

**Spécification détaillée**
### Phases d'implémentation
1. **Foundation** : Structure core avec patterns validés de mem0
2. **Core Features** : Fonctionnalité principale avec docs actuelles de context7
3. **Integration** : Connexion composants avec validation temps réel IDE
4. **Polish** : Raffinement et optimisation basés sur diagnostics

### Workflow continu
- **Avant chaque phase** : Rechercher patterns pertinents dans mem0
- **Pendant implémentation** : Validation temps réel avec diagnostics IDE
- **Après chaque feature** : Capturer patterns réussis avec `mcp__mem0__add_coding_preference`
- **Pour appels librairie** : Toujours vérifier syntaxe avec context7

**Bornes d'écriture**
* Autorisé : tous fichiers projet selon spécification task
* Interdit : fichiers système, configuration IDE, .git/

**Étapes**
1. [mem0] Charger patterns d'implémentation similaires
2. [context7] Obtenir documentation API courante
3. [sequential-thinking] Décomposer si >5 étapes
4. Implémenter par phases avec validation continue
5. [mem0] Sauvegarder patterns réussis
6. Tests et commits incrémentaux

**Points de vigilance**
- Respecter patterns existants améliorés par apprentissage mem0
- Validation continue avec diagnostics IDE
- Commits fréquents avec messages clairs
- Documentation mise à jour pendant implémentation
- Standards de qualité maintenus

**Tests/Validation**
- Tests écrits parallèlement à l'implémentation
- Validation temps réel avec diagnostics IDE
- Conformité aux patterns appris stockés dans mem0
- Vérification syntaxe avec documentation context7 actuelle

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