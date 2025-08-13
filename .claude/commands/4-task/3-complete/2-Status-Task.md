# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /4-task:3-complete:2-Status-Task

**Rôle**
Assistant d'analyse de progrès de tâche avec génération de rapport de statut complet et recommandations.

**Contexte**
Analyse complète du progrès de tâche incluant révision documentation, évaluation critères succès, calcul métriques et génération rapport intelligent avec recommandations actions.

**MCP à utiliser**
- **serena** : accéder documentation task pour analyse progrès
- **mem0** : capturer insights sur patterns de progrès et blockers

**Objectif**
Générer rapport de statut complet analysant progrès task vs critères succès, effort vs estimations, qualité métriques avec historique itérations et recommandations actions spécifiques.

**Spécification détaillée**

### Processus d'analyse statut
1. **Révision documentation task** : Analyser current-task.json, TEST.md, iterations.json pour métriques progrès
2. **Évaluation préparation complétion** : Évaluer préparation vs critères succès via analyse preprocessing
3. **Calcul métriques** : Calculer effort vs estimations et métriques qualité depuis données JSON
4. **Rapport intelligent** : Enrichir rapport avec analyse intelligente et recommandations actions spécifiques

**Bornes d'écriture**
* Autorisé : Aucune écriture - mode lecture seule pour analyse
* Interdit : Tous fichiers (analyse seulement)

**Étapes**
1. [serena] Réviser docs/3-current-task/current-task.json pour métriques progrès
2. [serena] Analyser docs/3-current-task/TEST.md pour statut validation
3. [serena] Examiner docs/3-current-task/iterations.json pour historique itérations
4. Évaluer préparation complétion vs critères succès
5. Calculer effort vs estimations et métriques qualité
6. Générer rapport avec analyse intelligente
7. [mem0] Capturer insights patterns progrès et blockers
8. Fournir recommandations actions spécifiques next steps

**Points de vigilance**
- Montrer historique itérations et leçons apprises
- Fournir pourcentage complétion clair et blockers
- Analyser écarts effort vs estimations
- Identifier patterns blocage récurrents

**Tests/Validation**
- Analyse complète documentation task (JSON, TEST.md, iterations.json)
- Évaluation préparation complétion vs critères succès
- Métriques effort/qualité calculées précisément
- Rapport enrichi avec recommandations actions intelligentes
- Insights patterns progrès capturés dans mem0

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
