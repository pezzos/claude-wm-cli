# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /1-project:2-update:1-Import-feedback

**Rôle**
Assistant d'intégration de feedback avec analyse technique et mise à jour documentaire intelligente.

**Contexte**
Traitement et intégration du feedback projet depuis docs/1-project/FEEDBACK.md vers la documentation technique (ARCHITECTURE.md) et de présentation (README.md) avec préservation du contenu existant.

**MCP à utiliser**
- **serena** : accéder aux fichiers de documentation existants
- **mem0** : mémoriser les insights techniques pour cohérence
- **sequential-thinking** : analyser les contradictions potentielles

**Objectif**
Intégrer intelligemment le feedback depuis FEEDBACK.md dans ARCHITECTURE.md et README.md en préservant le contenu existant et en signalant les contradictions.

**Spécification détaillée**

### Processus d'intégration
1. Lecture et analyse complète du contenu docs/1-project/FEEDBACK.md
2. Extraction des insights techniques pour mise à jour ARCHITECTURE.md
3. Extraction des éléments de présentation pour mise à jour README.md
4. Fusion intelligente avec contenu existant (pas de remplacement)
5. Détection et signalement des contradictions pour révision utilisateur

**Bornes d'écriture**
* Autorisé : docs/1-project/ARCHITECTURE.md, docs/1-project/README.md
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Lire docs/1-project/FEEDBACK.md
2. [sequential-thinking] Analyser et catégoriser le feedback
3. [serena] Lire ARCHITECTURE.md et README.md existants
4. Extraire insights techniques pour ARCHITECTURE.md
5. Extraire éléments présentation pour README.md
6. [mem0] Mémoriser contradictions détectées
7. Fusionner contenu avec préservation existant

**Points de vigilance**
- Fusionner uniquement, ne jamais remplacer le contenu existant
- Signaler explicitement toute contradiction détectée
- Maintenir la cohérence architecturale
- Préserver la structure documentaire existante

**Tests/Validation**
- Vérification de la préservation du contenu original
- Détection effective des contradictions
- Cohérence entre feedback intégré et architecture existante

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