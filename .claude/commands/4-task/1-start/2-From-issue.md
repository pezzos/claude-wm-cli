# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /4-task:1-start:2-From-issue

**Rôle**
Assistant d'analyse d'issues intelligent avec planification de résolution et stratégie de débogage.

**Contexte**
Analyse approfondie d'une issue GitHub pré-sélectionnée avec génération d'une stratégie de résolution complète et plan d'implémentation basé sur les métadonnées et contexte disponibles.

**MCP à utiliser**
- **github** : accéder aux métadonnées de l'issue si nécessaire
- **serena** : analyser le code existant pour la reproduction
- **sequential-thinking** : décomposer la stratégie de résolution complexe
- **mem0** : mémoriser patterns d'issues similaires

**Objectif**
Générer une analyse intelligente d'issue avec reproduction steps, stratégie de débogage et plan de résolution complet dans docs/3-current-task/current-task.json.

**Spécification détaillée**

### Contexte disponible
- docs/3-current-task/current-task.json - Données issue (pré-populées par preprocessing)
- Métadonnées GitHub et contexte issue complet
- Templates prêts pour génération de contenu

### Focus analyse intelligente
1. **Analyse de complexité** : évaluer la complexité de l'issue et potentiel de cause racine depuis current-task.json
2. **Steps de reproduction** : générer des étapes de reproduction complètes et approche de débogage
3. **Stratégie de résolution** : créer stratégie et plan d'implémentation détaillé
4. **Mise à jour contexte** : enrichir docs/3-current-task/current-task.json avec analyse complète

**Bornes d'écriture**
* Autorisé : docs/3-current-task/, docs/templates/
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Lire et analyser docs/3-current-task/current-task.json
2. [github] Accéder aux métadonnées issue si nécessaire
3. [sequential-thinking] Analyser complexité et stratégie de résolution
4. [serena] Analyser code existant pour reproduction
5. Générer steps de reproduction et approche débogage
6. [mem0] Mémoriser patterns pour issues similaires
7. Enrichir current-task.json avec analyse complète

**Points de vigilance**
- Le preprocessing a déjà géré sélection, assignation et setup workspace
- Se concentrer sur analyse intelligente et planification solution
- Assurer conformité schéma JSON strict
- Intégrer validation post-génération

**Tests/Validation**
- Validation schéma current-task.json avec schema-enforcer.sh
- Cohérence entre analyse et métadonnées issue
- Complétude des steps de reproduction

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
