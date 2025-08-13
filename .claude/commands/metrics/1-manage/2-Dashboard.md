# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /metrics:1-manage:2-Dashboard

**Rôle**
Analyste visualisation performance avec génération de tableaux de bord ASCII et insights actionnables prioritisés.

**Contexte**
Parsing METRICS.md pour indicateurs clés performance avec création visualisations ASCII pour vélocité, qualité et tendances completion. Génération analyse corrélation performance et identification goulots étranglement.

**MCP à utiliser**
- **serena** : parser METRICS.md et données performance
- **sequential-thinking** : analyser corrélations et tendances
- **mem0** : capitaliser sur patterns visualisation efficaces

**Objectif**
Générer tableau de bord visuel avec graphiques ASCII et insights actionnables pour optimisation performance projet.

**Spécification détaillée**

### Processus génération dashboard
1. Parse METRICS.md données indicateurs clés performance
2. Création visualisations ASCII vélocité, qualité et tendances completion
3. Génération analyse corrélation performance et identification bottlenecks
4. Fourniture recommandations optimisation spécifiques avec ranking priorité

### Focus insights actionnables
- Insights actionnables prioritaires sur esthétique graphiques
- Mise en évidence issues critiques nécessitant attention immédiate
- Corrélations performance exploitables

**Bornes d'écriture**
* Autorisé : output console/dashboard, documentation insights
* Interdit : modification METRICS.md source, configuration système

**Étapes**
1. [serena] Parser données METRICS.md
2. [sequential-thinking] Analyser tendances et corrélations
3. Générer visualisations ASCII appropriées
4. Identifier bottlenecks et issues critiques
5. [mem0] Mémoriser patterns visualisation efficaces
6. Prioriser recommandations optimisation

**Points de vigilance**
- Priorité insights actionnables sur esthétique
- Issues critiques mise en évidence immédiate
- Corrélations performance exploitables
- Visualisations ASCII lisibles et informatives

**Tests/Validation**
- Vérification parsing données METRICS.md
- Validation pertinence visualisations générées
- Test actionnabilité recommendations optimisation

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