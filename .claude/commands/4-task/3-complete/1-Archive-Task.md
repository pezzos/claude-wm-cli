# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /4-task:3-complete:1-Archive-Task

**Rôle**
Assistant d'archivage de tâche complétée avec capture d'apprentissages et nettoyage workspace.

**Contexte**
Archivage systématique d'une tâche complétée incluant révision documentation, mise à jour technique, capture apprentissages mem0 et validation statut complétion.

**MCP à utiliser**
- **mem0** : capturer patterns réutilisables et leçons apprises avec mcp__mem0__add_coding_preference
- **serena** : accéder documentation existante pour révision et mise à jour

**Objectif**
Archiver tâche complétée en capturant apprentissages réutilisables, mettant à jour documentation technique et nettoyant workspace current-task.

**Spécification détaillée**

### Processus d'archivage
1. **Révision documentation archivée** : Examiner docs pré-archivées depuis docs/3-current-task/
2. **Mise à jour technique** : Enrichir épic TECHNICAL.md avec décisions techniques et patterns
3. **Capture apprentissages** : Stocker key learnings via mem0 et enrichir CLAUDE.md global
4. **Validation complétion** : Vérifier statut complétion task pré-mis à jour dans PRD.md et stories.json

**Bornes d'écriture**
* Autorisé : docs/2-current-epic/TECHNICAL.md, CLAUDE.md, docs/2-current-epic/stories.json
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Réviser documentation tâche archivée depuis docs/3-current-task/
2. Mettre à jour docs/2-current-epic/TECHNICAL.md avec décisions techniques et patterns
3. [mem0] Capturer patterns réutilisables et leçons apprises
4. Enrichir CLAUDE.md global avec apprentissages contexte projet
5. Valider statut complétion task dans PRD.md et stories.json
6. Nettoyer workspace current-task si nécessaire

**Points de vigilance**
- Capturer patterns réutilisables pour projets futurs
- Documenter décisions techniques importantes
- Assurer apprentissages stockés dans mem0 pour réutilisation
- Nettoyer workspace current-task pour prochaine tâche

**Tests/Validation**
- Documentation technique enrichie avec patterns et décisions
- Apprentissages capturés dans mem0 avec patterns réutilisables
- CLAUDE.md global enrichi avec contexte projet
- Statut complétion task validé dans PRD.md et stories.json
- Workspace current-task nettoyé

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

### MANDATORY: Schema Compliance for docs/2-current-epic/stories.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements stories
```

### Schema-Aware Generation
When updating docs/3-current-task/stories.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements stories
```

**MANDATORY REQUIREMENTS:**
1. **$schema field**: The JSON file MUST contain a "$schema" field with the value ".claude/commands/templates/schemas/stories.schema.json"
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
