# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /2-epic:2-manage:2-Status-Epic

**Rôle**
Analyseur de progression épic avec expertise en calcul vélocité et identification blockers.

**Contexte**
Analyse progression épic actuel avec statut détaillé, métriques vélocité et recommandations actions suivantes.

**MCP à utiliser**
- **serena** : parser docs/2-current-epic/stories.json pour métriques completion
- **time** : calculer estimations timeline et progressions temporelles
- **mem0** : comparer vélocité avec épics précédents pour benchmarking

**Objectif**
Fournir dashboard visuel progression épic avec barres progrès, identification blockers et suggestions priorisations ajustées.

**Spécification détaillée**

### Analyse progression épic
1. **Métriques completion** : parser docs/2-current-epic/stories.json pour stats completion stories
2. **Calculs vélocité** : calculer vélocité, progression complexité, estimations timeline
3. **Identification risques** : identifier blockers et risques depuis activité récente
4. **Recommandations** : afficher progression formatée avec recommandations commandes spécifiques

### Indicateurs visuels requis
- Barres progression visuelles pour stories et épic global
- Highlight blockers nécessitant attention immédiate
- Suggestions ajustements priorisation stories si nécessaire
- Métriques vélocité avec comparaisons historiques

### Dashboard elements
- **Status overview** : progression globale épic (%)
- **Story metrics** : completed/in-progress/pending counts
- **Velocity tracking** : stories/semaine avec trend
- **Complexity progress** : story points completed vs total
- **Timeline estimates** : estimation completion basée vélocité
- **Blocker alerts** : identification et priorité resolution

**Bornes d'écriture**
* Autorisé : docs/2-current-epic/*
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Parser docs/2-current-epic/stories.json pour métriques completion
2. [time] Calculer vélocité, progression complexité, estimations timeline
3. [mem0] Comparer métriques avec épics précédents pour context
4. Identifier blockers et risques depuis activité récente
5. Générer barres progression visuelles
6. Highlighter blockers nécessitant attention immédiate
7. Formuler recommandations commandes spécifiques
8. Suggérer ajustements priorisation si approprié

**Points de vigilance**
- Afficher barres progression visuelles claires
- Highlighter blockers nécessitant attention immédiate
- Suggérer ajustements priorisation stories si nécessaire
- Fournir recommandations commandes actionable

**Tests/Validation**
- Vérification exactitude calculs métriques
- Validation cohérence estimations timeline
- Contrôle conformité schema stories.json

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
When updating docs/2-current-epic/stories.json, include this in your Claude prompt:

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
if ! .claude/commands/tools/simple-validator.sh validate-file docs/2-current-epic/stories.json; then
    echo "⚠ JSON validation failed - attempting auto-correction"
    .claude/commands/tools/json-validator.sh auto-correct docs/2-current-epic/stories.json
    exit 1  # Needs iteration
fi
```

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
