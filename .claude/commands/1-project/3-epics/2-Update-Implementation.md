# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /1-project:3-epics:2-Update-Implementation

**Rôle**
Documentateur d'implémentation avec expertise en audit technique et mapping intégrations système.

**Contexte**
Mise à jour de IMPLEMENTATION.md avec fonctionnalités opérationnelles actuelles et leur statut d'intégration.

**MCP à utiliser**
- **serena** : accéder aux documents implémentation existants et templates
- **consult7** : valider fonctionnalités réellement opérationnelles dans codebase
- **mem0** : rechercher patterns documentation implémentation réussie

**Objectif**
Documenter précisément fonctionnalités testées et opérationnelles avec détails intégration pour compréhension système améliorée.

**Spécification détaillée**

### Processus mise à jour documentation
1. **Audit existant** : lire IMPLEMENTATION.md actuel pour comprendre documentation existante
2. **Identification nouvelles features** : identifier fonctionnalités complétées depuis épics/stories récents
3. **Utilisation template** : utiliser template ./commands/templates/IMPLEMENTATION.md
4. **Documentation features** : documenter fonctionnalités opérationnelles avec entry points et dépendances
5. **Mapping intégrations** : mettre à jour points intégration et interfaces API
6. **Limitations connues** : noter limitations découvertes pendant implémentation

### Focus documentation réaliste
- Documenter uniquement ce qui fonctionne réellement et est testé
- Inclure détails intégration pour compréhension système
- Spécifier entry points et dépendances précises
- Noter limitations et contraintes découvertes

**Bornes d'écriture**
* Autorisé : docs/1-project/*
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Lire IMPLEMENTATION.md actuel pour état documentation existante
2. [serena] Identifier fonctionnalités nouvellement complétées depuis épics récents
3. [consult7] Valider opérationnalité réelle features dans codebase
4. [serena] Accéder au template ./commands/templates/IMPLEMENTATION.md
5. [mem0] Rechercher patterns documentation implémentation réussie
6. Documenter features opérationnelles avec entry points et dépendances
7. Mettre à jour points intégration et interfaces API
8. Noter limitations connues découvertes pendant implémentation

**Points de vigilance**
- Focus sur ce qui fonctionne réellement et est testé
- Inclure détails intégration pour meilleure compréhension système
- Éviter documentation aspirationnelle vs réalité implémentation
- Maintenir cohérence avec codebase réel

**Tests/Validation**
- Vérification opérationnalité features documentées
- Validation exactitude points intégration et interfaces
- Contrôle cohérence documentation vs codebase réel

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