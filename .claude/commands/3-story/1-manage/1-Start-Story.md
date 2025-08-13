# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /3-story:1-manage:1-Start-Story

**Rôle**
Gestionnaire de démarrage story avec expertise en sélection priorité et validation dépendances.

**Contexte**
Sélection story priorité suivante et création docs/2-current-epic/current-story.json avec extraction tâches techniques.

**MCP à utiliser**
- **serena** : accéder aux fichiers stories.json pour sélection et mise à jour
- **time** : horodater démarrage story pour tracking progression
- **sequential-thinking** : décomposer story en tâches techniques structurées

**Objectif**
Initialiser story priorité avec validation dépendances, création current-story.json et extraction tâches techniques dans stories.json.

**Spécification détaillée**

### Processus sélection story
1. **Sélection priorité** : lire docs/2-current-epic/stories.json et identifier story priorité plus haute non démarrée (P0 > P1 > P2 > P3)
2. **Validation dépendances** : vérifier toutes dépendances story marquées complètes
3. **Création contexte** : créer docs/2-current-epic/current-story.json avec détails story sélectionnée
4. **Mise à jour statut** : mettre à jour docs/2-current-epic/stories.json : marquer story "🚧 In Progress - {date}"
5. **Extraction tâches** : extraire tâches techniques depuis story et mettre à jour champ tasks dans stories.json

### Gestion tâches
- Tâches stockées dans story au sein de docs/2-current-epic/stories.json
- PAS de fichier séparé todo.json
- Décomposition technique basée story requirements
- Tracking progression à l'intérieur structure story

**Bornes d'écriture**
* Autorisé : docs/2-current-epic/*
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Lire docs/2-current-epic/stories.json pour identifier stories disponibles
2. Identifier story priorité plus haute non démarrée (P0 > P1 > P2 > P3)
3. [serena] Vérifier toutes dépendances story marquées complètes
4. [time] Générer timestamp pour marquage démarrage
5. Créer docs/2-current-epic/current-story.json avec détails story
6. [serena] Mettre à jour stories.json : marquer story "🚧 In Progress - {date}"
7. [sequential-thinking] Extraire tâches techniques depuis story
8. Mettre à jour champ tasks dans docs/2-current-epic/stories.json
9. Valider conformité schema JSON

**Points de vigilance**
- Valider dépendances story complètes avant démarrage
- Tâches stockées dans story au sein stories.json (pas todo.json séparé)
- Respecter hiérarchie priorités P0 > P1 > P2 > P3
- Horodater précisément démarrage pour tracking

**Tests/Validation**
- Vérification conformité schema current-story.json
- Validation complétion dépendances story
- Contrôle cohérence mise à jour stories.json

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

### MANDATORY: Schema Compliance for docs/2-current-epic/current-story.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements current-story
```

### Schema-Aware Generation
When updating docs/2-current-epic/current-story.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements current-story
```

**MANDATORY REQUIREMENTS:**
1. **$schema field**: The JSON file MUST contain a "$schema" field with the value ".claude/commands/templates/schemas/current-story.schema.json"
2. All required fields must be present with correct types and values
3. All nested objects must have their required fields
### Post-Generation Validation
After completing the main task, validate the generated JSON:

```bash
# Validate with auto-correction
if ! .claude/commands/tools/simple-validator.sh validate-file docs/2-current-epic/current-story.json; then
    echo "⚠ JSON validation failed - attempting auto-correction"
    .claude/commands/tools/json-validator.sh auto-correct docs/2-current-epic/current-story.json
    exit 1  # Needs iteration
fi
```

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
