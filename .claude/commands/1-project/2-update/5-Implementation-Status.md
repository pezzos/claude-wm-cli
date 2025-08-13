# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /1-project:2-update:5-Implementation-Status

**Rôle**
Analyseur de statut d'implémentation avec expertise en audit de fonctionnalités et couverture technique.

**Contexte**
Analyse détaillée de IMPLEMENTATION.md pour présenter l'état des fonctionnalités actives, points d'intégration et couverture globale.

**Prérequis**
- IMPLEMENTATION.md doit exister

**MCP à utiliser**
- **serena** : analyser IMPLEMENTATION.md et fichiers techniques liés
- **consult7** : cross-référencer avec codebase réelle
- **mem0** : comparer avec implémentations similaires passées

**Objectif**
Générer rapport d'état implémentation présentant fonctionnalités opérationnelles, points d'intégration et taux de couverture.

**Spécification détaillée**

### Analyse d'implémentation
1. **Audit fonctionnalités** : identifier features actives vs planifiées
2. **Points d'intégration** : mapper connections externes et APIs
3. **Couverture technique** : évaluer complétion par composant
4. **Gaps analysis** : identifier écarts entre spec et implémentation

### Présentation utilisateur
- **Dashboard fonctionnalités** : statut visuel des features
- **Carte intégrations** : schéma connections système
- **Métriques couverture** : pourcentages complétion
- **Roadmap implémentation** : prochaines priorités

**Bornes d'écriture**
* Autorisé : docs/1-project/*
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Vérifier existence IMPLEMENTATION.md
2. [serena] Analyser contenu et structure IMPLEMENTATION.md
3. [consult7] Cross-référencer avec codebase pour validation
4. [mem0] Comparer avec implémentations références
5. Identifier fonctionnalités opérationnelles vs en cours
6. Mapper points d'intégration et dépendances
7. Calculer métriques couverture par composant
8. Générer rapport de présentation utilisateur

**Points de vigilance**
- Distinguer fonctionnalités opérationnelles des prototypes
- Vérifier cohérence documentation vs implémentation réelle
- Identifier risques d'intégration et dépendances critiques
- Présenter informations de manière actionnable

**Tests/Validation**
- Vérification existence et accessibilité IMPLEMENTATION.md
- Validation cohérence métriques avec codebase réelle
- Contrôle exactitude mapping intégrations

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