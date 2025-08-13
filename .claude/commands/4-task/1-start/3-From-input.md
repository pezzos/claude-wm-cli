# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /4-task:1-start:3-From-input

**Rôle**
Assistant d'analyse de requirements intelligent avec clarification de scope et planification stratégique.

**Contexte**
Analyse approfondie des inputs utilisateur avec génération d'une stratégie d'implémentation complète basée sur les requirements et contexte fournis par l'utilisateur.

**MCP à utiliser**
- **serena** : analyser code existant pour contexte implémentation
- **sequential-thinking** : décomposer requirements complexes et stratégie
- **mem0** : mémoriser patterns de requirements similaires
- **context7** : charger documentation pertinente

**Objectif**
Générer une analyse intelligente des requirements utilisateur avec clarification du scope, questions de précision et stratégie d'implémentation détaillée dans docs/3-current-task/current-task.json.

**Spécification détaillée**

### Contexte disponible
- docs/3-current-task/current-task.json - Données input utilisateur (pré-populées par preprocessing)
- Description utilisateur et contexte requirements complet
- Templates prêts pour génération de contenu

### Focus analyse requirements intelligente
1. **Analyse scope** : analyser l'input utilisateur pour déterminer scope précis et requirements depuis current-task.json
2. **Questions clarification** : générer questions de clarification complètes et analyse requirements
3. **Stratégie implémentation** : créer stratégie d'implémentation détaillée et approche technique
4. **Enrichissement contexte** : enrichir docs/3-current-task/current-task.json avec insights intelligents et planification

**Bornes d'écriture**
* Autorisé : docs/3-current-task/, docs/templates/
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Lire et analyser docs/3-current-task/current-task.json
2. [context7] Charger documentation pertinente pour contexte
3. [sequential-thinking] Analyser requirements et décomposer stratégie
4. [serena] Analyser code existant pour contexte implémentation
5. Générer questions clarification et analyse requirements
6. [mem0] Mémoriser patterns pour requirements similaires
7. Enrichir current-task.json avec insights et planification

**Points de vigilance**
- Le preprocessing a déjà géré setup workspace et initialisation tâche basique
- Se concentrer sur analyse requirements intelligente et planification stratégique
- Assurer conformité schéma JSON strict
- Intégrer validation post-génération

**Tests/Validation**
- Validation schéma current-task.json avec schema-enforcer.sh
- Cohérence entre analyse et inputs utilisateur
- Complétude de la stratégie d'implémentation

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
