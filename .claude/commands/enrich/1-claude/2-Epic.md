# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /enrich:1-claude:2-Epic

**Rôle**
Analyste épique spécialisé dans l'enrichissement de documentation domaine et l'extraction de patterns techniques épique-spécifiques.

**Contexte**
Analyse des stories complétées de l'épique courante pour extraire patterns domaine, décisions techniques et optimisations spécifiques. Mise à jour de docs/2-current-epic/CLAUDE.md avec connaissances actionnables.

**MCP à utiliser**
- **serena** : analyser stories complétées et code associé
- **sequential-thinking** : extraire patterns techniques systématiquement
- **mem0** : capitaliser sur patterns domaine réutilisables

**Objectif**
Enrichir la documentation épique avec patterns domaine spécifiques et décisions techniques pour améliorer la cohérence et performance future.

**Spécification détaillée**

### Processus d'analyse épique
1. Review completed stories et extraction patterns domaine-spécifiques
2. Documentation décisions techniques et patterns expérience utilisateur
3. Mise à jour docs/2-current-epic/CLAUDE.md avec connaissances actionnables
4. Capture patterns intégration et optimisations performance épique

### Focus spécialisé
Patterns épique-spécifiques non applicables globalement :
- Terminologie domaine et règles métier
- Décisions architecture épique
- Optimisations performance spécifiques

**Bornes d'écriture**
* Autorisé : docs/2-current-epic/CLAUDE.md
* Interdit : documentation globale projet, configuration système

**Étapes**
1. [serena] Analyser stories complétées épique courante
2. [sequential-thinking] Extraire patterns techniques et domaine
3. Documenter décisions épique-spécifiques
4. [mem0] Mémoriser patterns réutilisables
5. Enrichir CLAUDE.md épique avec insights actionnables

**Points de vigilance**
- Focus épique uniquement (éviter généralisation excessive)
- Patterns actionnables pour stories futures
- Cohérence avec architecture globale projet
- Performance et intégration spécifiques domaine

**Tests/Validation**
- Vérification pertinence patterns extraits
- Validation cohérence avec architecture projet
- Test applicabilité insights futurs stories

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