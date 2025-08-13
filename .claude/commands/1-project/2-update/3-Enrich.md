# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /1-project:2-update:3-Enrich

**Rôle**
Architecte documentaire avec expertise en réorganisation technique et extraction de patterns réutilisables.

**Contexte**
Réorganisation de TECHNICAL.md basée sur l'évolution projet et mise à jour de CLAUDE.md global avec les meilleures pratiques découvertes.

**MCP à utiliser**
- **mem0** : rechercher patterns projet pour insights réutilisables
- **serena** : accéder aux fichiers techniques existants
- **time** : dater les versions archivées

**Objectif**
Extraire et organiser les patterns techniques actionnable pour enrichir la documentation globale avec des pratiques éprouvées.

**Spécification détaillée**

### Processus de réorganisation
1. **Review et archivage** : examiner TECHNICAL.md actuel et sauvegarder version précédente
2. **Restructuration** : réorganiser avec sections claires (décisions, patterns, outils, leçons apprises)
3. **Extraction patterns** : rechercher avec mem0 les insights réutilisables du projet
4. **Enrichissement global** : mettre à jour CLAUDE.md global avec patterns découverts

### Focus sur l'actionnable
- Privilégier patterns qui fonctionnent en pratique vs documentation théorique
- Extraire commandes et workflows éprouvés
- Documenter anti-patterns identifiés et leurs solutions
- Capturer métriques de succès et échecs

**Bornes d'écriture**
* Autorisé : docs/1-project/*
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Lire TECHNICAL.md actuel
2. [time] Créer archive avec timestamp de l'ancienne version
3. [mem0] Rechercher patterns techniques réutilisables du projet
4. Réorganiser TECHNICAL.md avec sections structurées
5. Identifier patterns actionnable vs documentation pure
6. Mettre à jour CLAUDE.md global avec découvertes
7. Valider cohérence entre documents

**Points de vigilance**
- Se concentrer sur ce qui fonctionne réellement en pratique
- Éviter documentation théorique sans valeur actionnable
- Préserver historique des versions précédentes
- Maintenir cohérence avec standards projet existants

**Tests/Validation**
- Vérification de la structuration claire de TECHNICAL.md
- Validation de l'enrichissement cohérent de CLAUDE.md
- Contrôle de préservation de l'archive précédente

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