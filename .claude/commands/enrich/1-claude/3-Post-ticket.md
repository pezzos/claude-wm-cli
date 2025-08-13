# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /enrich:1-claude:3-Post-ticket

**Rôle**
Analyste post-tâche spécialisé dans l'extraction de patterns apprentissage et l'enrichissement documentaire multi-niveaux.

**Contexte**
Extraction de leçons apprises depuis archive tâche complétée pour enrichir documentation appropriée. Catégorisation apprentissages globaux vs épique-spécifiques avec mémorisation patterns réutilisables.

**MCP à utiliser**
- **serena** : analyser archive tâche complétée et code associé
- **mem0** : mémoriser patterns clés et techniques debugging
- **sequential-thinking** : catégoriser apprentissages méthodiquement

**Objectif**
Capitaliser sur apprentissages tâche complétée pour enrichir documentation appropriée avec patterns actionnables et techniques debugging.

**Spécification détaillée**

### Processus d'extraction apprentissages
1. Analyse archive tâche complétée pour patterns et insights réutilisables
2. Catégorisation apprentissages (globaux vs épique-spécifiques) et détermination CLAUDE.md cible
3. Stockage patterns clés **avec mem0** et mise à jour fichiers CLAUDE.md pertinents
4. Focus sur patterns actionnables, techniques debugging et améliorations processus

### Capture bi-directionnelle
- Patterns succès et leçons échecs
- Exemples code spécifiques avec contexte
- Techniques debugging éprouvées
- Améliorations processus identifiées

**Bornes d'écriture**
* Autorisé : CLAUDE.md global projet, CLAUDE.md épique courante
* Interdit : archives tâches, configuration système

**Étapes**
1. [serena] Analyser archive tâche complétée
2. [sequential-thinking] Catégoriser apprentissages par portée
3. [mem0] Mémoriser patterns clés et techniques
4. Déterminer CLAUDE.md cible (global/épique)
5. Enrichir documentation avec insights actionnables

**Points de vigilance**
- Capture patterns succès ET échecs
- Exemples code avec contexte suffisant
- Catégorisation appropriée global/épique
- Focus sur actionnabilité future

**Tests/Validation**
- Vérification pertinence patterns extraits
- Validation catégorisation apprentissages
- Test applicabilité insights futurs

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