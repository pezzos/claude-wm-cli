# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /debug:1-Check-state

**Rôle**
Assistant de diagnostic d'état projet avec vérification git, structure et cohérence workflow.

**Contexte**
Vérification complète état projet incluant statut git, intégrité structure dossiers et alignement workflow avec génération health score et recommandations actions.

**MCP à utiliser**
- **serena** : analyser structure projet et documentation existante
- **github** : consulter statut git si métadonnées nécessaires
- **mem0** : capturer patterns diagnostics pour réutilisation

**Objectif**
Vérifier état projet complet (git, structure, workflow) et fournir diagnostics actionnables avec score santé et recommandations spécifiques.

**Spécification détaillée**

### Processus diagnostic état
1. **Vérification git** : Contrôler statut repository et changements non commités
2. **Intégrité structure** : Vérifier docs/, fichiers requis, états workspace
3. **Cohérence workflow** : Analyser statut epic/story/task pour alignement
4. **Score santé** : Générer score santé et recommandations actions spécifiques

**Bornes d'écriture**
* Autorisé : Aucune écriture - mode diagnostic lecture seule
* Interdit : Tous fichiers (diagnostic seulement)

**Étapes**
1. Vérifier statut git repository et changements non commités
2. [serena] Analyser intégrité structure dossiers (docs/, fichiers requis)
3. [serena] Examiner états workspace pour cohérence
4. Analyser alignement epic/story/task status
5. Générer health score projet avec métriques claires
6. [mem0] Capturer patterns diagnostic pour réutilisation
7. Fournir recommandations actions spécifiques avec commandes

**Points de vigilance**
- Fournir diagnostics actionnables avec commandes spécifiques
- Montrer indicateurs santé projet clairs
- Identifier incohérences workflow et misalignements
- Capturer patterns diagnostic récurrents

**Tests/Validation**
- Statut git vérifié (uncommitted changes, branch status)
- Intégrité structure dossiers validée
- Cohérence workflow epic/story/task confirmée
- Health score généré avec métriques précises
- Recommandations actions avec commandes spécifiques

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