# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /1-project:3-epics:1-Plan-Epics

**Rôle**
Planificateur d'épics avec expertise en architecture de livraison et découpage valeur utilisateur.

**Contexte**
Archivage de l'epics.json précédent et création nouveau planning épic basé sur vision projet au format JSON structuré.

**MCP à utiliser**
- **mem0** : rechercher patterns épic réussis pour insights réutilisables
- **serena** : accéder aux documents IMPLEMENTATION.md, ARCHITECTURE.md, README.md
- **time** : dater archivage et horodater création nouveau plan

**Objectif**
Créer plan épic structuré JSON avec séquençage dépendances, focus livraison valeur utilisateur et scope 2-4 semaines par épic.

**Spécification détaillée**

### Structure JSON requis
- Utiliser schema .claude/commands/templates/schemas/epics.schema.json
- Chaque épic doit inclure : id, title, description, status, priority, business_value, target_users, success_criteria, dependencies, blockers, story_themes
- NE PAS inclure userStories - appartiennent à docs/2-current-epic/stories.json avec liaison epic_id
- Inclure section project_context avec current_epic, total_epics, completed_epics, project_phase

### Processus planification épic
1. **Archivage** : sauvegarder epics.json existant vers docs/archive/epics-archive-{date}/
2. **Analyse état** : lire IMPLEMENTATION.md pour comprendre fonctionnalités opérationnelles
3. **Gap analysis** : comparer ARCHITECTURE.md et README.md pour identifier manques
4. **Pattern research** : rechercher avec mem0 patterns épic réussis
5. **Structuration JSON** : créer nouveau epics.json conforme schema
6. **Définition épics** : 3-5 épics avec story themes, dépendances, critères succès

### Critères qualité
- Baser épics sur livraison valeur utilisateur
- Séquencer par dépendances et risque
- Scope 2-4 semaines par épic
- Format JSON strict (pas markdown)

**Bornes d'écriture**
* Autorisé : docs/1-project/*, docs/archive/*
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [time] Générer timestamp pour archivage
2. [serena] Archiver epics.json existant vers docs/archive/epics-archive-{date}/
3. [serena] Lire IMPLEMENTATION.md pour état fonctionnalités actuelles
4. [serena] Analyser ARCHITECTURE.md et README.md pour gaps identification
5. [mem0] Rechercher patterns épic réussis dans contexte projet
6. Créer nouveau docs/1-project/epics.json conforme schema
7. Définir 3-5 épics avec story themes et dépendances
8. Valider conformité JSON schema

**Points de vigilance**
- Générer docs/1-project/epics.json (format JSON) vs epics.json (markdown)
- Respecter strictement schema requirements
- Séparer DONE/TODO lors archivage
- Focus livraison valeur vs features techniques

**Tests/Validation**
- Validation conformité schema JSON
- Vérification cohérence dépendances entre épics
- Contrôle scope 2-4 semaines par épic

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

### MANDATORY: Schema Compliance for epics.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements epics
```

### Schema-Aware Generation
When updating docs/1-project/epics.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements epics
```

**MANDATORY REQUIREMENTS:**
1. **$schema field**: The JSON file MUST contain a "$schema" field with the value ".claude/commands/templates/schemas/epics.schema.json"
2. All required fields must be present with correct types and values
3. No forbidden fields (like userStories) should be included

### Post-Generation Validation
After completing the main task, validate the generated JSON:

```bash
# Validate with auto-correction
if ! .claude/commands/tools/simple-validator.sh validate-file docs/1-project/epics.json; then
    echo "⚠ JSON validation failed - attempting auto-correction"
    .claude/commands/tools/json-validator.sh auto-correct docs/1-project/epics.json
    exit 1  # Needs iteration
fi
```

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
