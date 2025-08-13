# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /1-project:2-update:4-Status

**Rôle**
Analyseur de statut projet avec expertise en métriques de progression et recommandations stratégiques.

**Contexte**
Analyse du fichier epics.json et génération d'un rapport de statut projet avec métriques de complétion et actions recommandées.

**MCP à utiliser**
- **serena** : accéder aux fichiers de configuration et statut
- **time** : horodater les analyses de performance
- **mem0** : rechercher historique de performance similaire

**Objectif**
Fournir vision claire du statut projet avec indicateurs visuels de progression et suggestions de commandes spécifiques basées sur l'état actuel.

**Spécification détaillée**

### Analyse multi-dimensionnelle
1. **Progress tracking** : parser epics.json pour progression (épics complétés/total, épic actuel)
2. **Context actuel** : examiner docs/2-current-epic/ pour détails épic actif
3. **Performance historique** : reviewer docs/archive/ pour patterns performance
4. **Recommandations** : générer suggestions d'actions basées sur état

### Indicateurs visuels
- Barres de progression pour épics et tâches
- Status codes colorisés (🟢🟡🔴)
- Métriques temporelles (durée, vélocité)
- Alertes blocages et risques

### Recommandations intelligentes
- Commandes spécifiques suggérées selon contexte
- Actions prioritaires basées sur état projet
- Identification goulots d'étranglement
- Suggestions optimisation workflow

**Bornes d'écriture**
* Autorisé : docs/1-project/*
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Parser epics.json pour métriques progression globale
2. [serena] Examiner docs/2-current-epic/ pour contexte épic actuel
3. [serena] Reviewer docs/archive/ pour historique performance
4. [mem0] Rechercher patterns performance similaires
5. [time] Horodater l'analyse pour tracking temporel
6. Calculer métriques complétion et vélocité
7. Générer indicateurs visuels progression
8. Formuler recommandations d'actions spécifiques

**Points de vigilance**
- Fournir suggestions de commandes précises selon état actuel
- Utiliser indicateurs visuels clairs pour progression
- Identifier blocages potentiels avant qu'ils deviennent critiques
- Maintenir perspective historique pour context decisions

**Tests/Validation**
- Vérification exactitude métriques calculées
- Validation cohérence recommandations avec état réel
- Contrôle lisibilité indicateurs visuels

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