# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /enrich:1-claude:1-Global

**Rôle**
Assistant d'analyse d'évolution projet avec enrichissement intelligent du CLAUDE.md global.

**Contexte**
Analyse des patterns de codebase, décisions architecturales et workflows de développement pour enrichir le CLAUDE.md global avec des patterns actionables et conventions réutilisables.

**MCP à utiliser**
- **mem0** : extraire patterns de succès et évolution projet récente
- **serena** : analyser patterns codebase et décisions architecturales
- **context7** : charger documentation existante pour contexte
- **sequential-thinking** : structurer insights pour réutilisabilité

**Objectif**
Analyser l'évolution projet et enrichir le CLAUDE.md global avec des patterns découverts, conventions et insights actionables pour améliorer l'efficacité de développement future.

**Spécification détaillée**

### Processus d'analyse évolution
1. **Analyse patterns** : analyser patterns codebase, décisions architecturales et workflows de développement
2. **Extraction mem0** : extraire patterns de succès depuis mem0 et évolution projet récente
3. **Enrichissement CLAUDE.md** : mettre à jour CLAUDE.md global avec patterns actionables, commandes et conventions
4. **Focus réutilisabilité** : se concentrer sur insights réutilisables qui améliorent l'efficacité de développement future

**Bornes d'écriture**
* Autorisé : CLAUDE.md, docs/*, fichiers documentation globale
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Analyser patterns codebase et décisions architecturales
2. [mem0] Extraire patterns de succès et insights récents
3. [context7] Charger documentation existante pour contexte
4. [sequential-thinking] Structurer insights pour maximum réutilisabilité
5. Enrichir CLAUDE.md global avec patterns actionables
6. Documenter conventions et commandes éprouvées
7. [mem0] Mémoriser nouvelles conventions pour sessions futures

**Points de vigilance**
- Documenter ce qui fonctionne en pratique, pas seulement la théorie
- Inclure exemples de code spécifiques et commandes
- Se concentrer sur insights réutilisables et actionables
- Maintenir cohérence avec patterns existants

**Tests/Validation**
- Vérification de la qualité des patterns documentés
- Validation des exemples de code et commandes
- Cohérence avec l'architecture projet existante

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