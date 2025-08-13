# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /4-task:1-start:1-From-story

**Rôle**
Analyseur de tâches avec expertise en décomposition technique et planification stratégique basée contexte story.

**Contexte**
Génération analyse intelligente tâche et planification depuis contexte story actuelle avec enrichissement insights techniques.

**Contexte disponible**
- docs/3-current-task/current-task.json - Données tâche actuelle (pré-peuplé par preprocessing)
- docs/2-current-epic/stories.json - Contexte story

**MCP à utiliser**
- **serena** : analyser contexte story et tâche pour insights techniques
- **sequential-thinking** : structurer approche implémentation et décomposition complexité
- **mem0** : rechercher patterns tâches similaires pour stratégies éprouvées
- **context7** : enrichir avec connaissances techniques spécifiques domaine

**Objectif**
Générer analyse intelligente tâche avec stratégie implémentation basée contexte story et enrichissement current-task.json avec insights techniques.

**Spécification détaillée**

### Focus analyse intelligente
Preprocessing a déjà géré gestion fichiers, sélection tâche, mises à jour statut.
Concentration sur génération contenu intelligent et analyse :
1. **Analyse complexité** : analyser complexité tâche et requirements depuis docs/3-current-task/current-task.json
2. **Description compréhensive** : générer description tâche compréhensive et approche
3. **Stratégie implémentation** : créer stratégie implémentation basée contexte story
4. **Enrichissement insights** : mettre à jour docs/3-current-task/current-task.json avec insights intelligents

### Analyse technique approfondie
- Évaluation complexité et dépendances techniques
- Identification patterns réutilisables depuis story contexte
- Création roadmap implémentation structurée
- Intégration best practices domaine-spécifiques

**Bornes d'écriture**
* Autorisé : docs/3-current-task/*
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Analyser docs/3-current-task/current-task.json pour contexte tâche pré-peuplé
2. [serena] Accéder docs/2-current-epic/stories.json pour contexte story complet
3. [sequential-thinking] Analyser complexité tâche et structurer approche
4. [mem0] Rechercher patterns tâches similaires pour insights réutilisables
5. [context7] Enrichir avec connaissances techniques spécifiques domaine
6. Générer description tâche compréhensive avec approche technique
7. Créer stratégie implémentation basée contexte story
8. Enrichir docs/3-current-task/current-task.json avec insights intelligents
9. Valider conformité schema JSON

**Points de vigilance**
- Preprocessing a déjà géré gestion fichiers et sélection tâche
- Focus génération contenu intelligent vs manipulation fichiers
- Enrichir avec insights techniques basés contexte story
- Respecter structure schema current-task.json

**Tests/Validation**
- Vérification conformité schema current-task.json
- Validation cohérence analyse avec contexte story
- Contrôle qualité insights techniques générés

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
## JSON Schema Validation
<!-- JSON_SCHEMA_VALIDATION -->

### MANDATORY: Schema Compliance for docs/3-current-task/current-task.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements current-task
```

### Schema-Aware Generation
When updating docs/3-current-task/current-task.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements current-task
```

**MANDATORY REQUIREMENTS:**
1. **$schema field**: The JSON file MUST contain a "$schema" field with the value ".claude/commands/templates/schemas/current-task.schema.json"
2. All required fields must be present with correct types and values
3. All nested objects must have their required fields

### Post-Generation Validation
After completing the main task, validate the generated JSON:

```bash
# Validate with auto-correction
if ! .claude/commands/tools/json-validator.sh validate; then
    echo "⚠ JSON validation failed - files auto-corrected"
    exit 1  # Needs iteration
fi
```

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
