# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /2-epic:2-manage:1-Complete-Epic

**Rôle**
Gestionnaire de complétion épic avec expertise en validation critères succès et archivage structuré.

**Contexte**
Archivage épic complété avec mise à jour métriques performance et validation critères succès avant finalisation.

**MCP à utiliser**
- **serena** : valider complétion stories et accéder aux fichiers épic
- **time** : horodater archivage et métriques
- **mem0** : enregistrer patterns succès épic pour réutilisabilité

**Objectif**
Finaliser épic avec validation critères succès, archivage structuré et enrichissement métriques performance pour apprentissage continu.

**Spécification détaillée**

### Processus complétion épic
1. **Validation complétion** : vérifier toutes stories docs/2-current-epic/stories.json sont ✅ complétées
2. **Vérification critères** : valider critères succès épic avant finalisation
3. **Archivage structuré** : archiver docs/2-current-epic/ vers docs/archive/{epic-name}-{date}/
4. **Mise à jour statut** : mettre à jour epics.json statut à "✅ Completed" avec métriques
5. **Enrichissement métriques** : enrichir metrics.json avec stats performance épic

### Validation quality gates
- Toutes stories marquées ✅ completed
- Critères succès épic validés
- Métriques performance calculées et documentées
- Archivage complet et horodaté

**Bornes d'écriture**
* Autorisé : docs/2-current-epic/*, docs/1-project/epics.json, docs/archive/*
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Vérifier complétion toutes stories dans docs/2-current-epic/stories.json
2. [serena] Valider critères succès épic avant finalisation
3. [time] Générer timestamp pour archivage
4. [serena] Archiver docs/2-current-epic/ vers docs/archive/{epic-name}-{date}/
5. [serena] Mettre à jour epics.json statut à "✅ Completed" avec métriques
6. [serena] Enrichir metrics.json avec stats performance épic
7. [mem0] Enregistrer patterns succès épic pour apprentissage
8. Valider conformité schema JSON

**Points de vigilance**
- Valider critères succès épic avant complétion
- Mettre à jour metrics.json avec data performance épic
- Archivage complet avec horodatage précis
- Capturer learnings pour épics futurs

**Tests/Validation**
- Vérification complétion toutes stories requises
- Validation critères succès épic
- Contrôle conformité schema metrics.json

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

### MANDATORY: Schema Compliance for metrics.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements metrics
```

### Schema-Aware Generation
When updating docs/2-current-epic/metrics.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements metrics
```

**MANDATORY REQUIREMENTS:**
1. **$schema field**: The JSON file MUST contain a "$schema" field with the value ".claude/commands/templates/schemas/metrics.schema.json"
2. All required fields must be present with correct types and values
3. All nested objects must have their required fields
### Post-Generation Validation
After completing the main task, validate the generated JSON:

```bash
# Validate with auto-correction
if ! .claude/commands/tools/simple-validator.sh validate-file docs/2-current-epic/metrics.json; then
    echo "⚠ JSON validation failed - attempting auto-correction"
    .claude/commands/tools/json-validator.sh auto-correct docs/2-current-epic/metrics.json
    exit 1  # Needs iteration
fi
```

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
