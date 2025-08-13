# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /metrics:1-manage:1-Update

**Rôle**
Analyste performance projet spécialisé dans la collecte de données archives et la mise à jour de métriques de performance complètes.

**Contexte**
Analyse archives projet (docs/archive/) et historique git pour calculer vélocité, métriques qualité et tendances. Mise à jour METRICS.md avec indicateurs performance et santé actuels.

**MCP à utiliser**
- **serena** : analyser archives et structure documentaire
- **github** : collecter métadonnées historique git
- **sequential-thinking** : analyser tendances méthodiquement
- **time** : dater analyses pour tracking temporel

**Objectif**
Maintenir METRICS.md à jour avec données performance complètes et insights actionnables pour optimisation projet.

**Spécification détaillée**

### Processus de collecte métriques
1. Collecte données docs/archive/ (épiques, stories, tâches) et historique git
2. Calcul vélocité, métriques qualité et analyse tendances
3. Mise à jour METRICS.md avec indicateurs performance et santé actuels
4. Génération insights et recommandations optimisation

### Métriques cible
- Tendances vélocité et taux succès
- Efficacité itération et temps cycles
- Indicateurs qualité et performance
- Insights actionnables optimisation

**Bornes d'écriture**
* Autorisé : docs/1-project/METRICS.md
* Interdit : archives (lecture seule), configuration système

**Étapes**
1. [serena] Analyser docs/archive/ pour données historiques
2. [github] Collecter métadonnées historique git
3. [sequential-thinking] Calculer métriques vélocité et qualité
4. [time] Dater analyse pour tracking temporel
5. Mettre à jour METRICS.md avec indicateurs actuels
6. Générer insights et recommandations

**Points de vigilance**
- Tracking tendances vélocité précis
- Métriques qualité représentatives
- Insights actionnables pour optimisation
- Conservation historique pour comparaisons

**Tests/Validation**
- Vérification cohérence données collectées
- Validation calculs métriques
- Test pertinence insights générés

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