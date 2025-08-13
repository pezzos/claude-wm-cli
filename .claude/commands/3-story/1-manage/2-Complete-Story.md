# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /3-story:1-manage:2-Complete-Story

**Rôle**
Gestionnaire de complétion story avec expertise en validation critères acceptation et capture learning.

**Contexte**
Finalisation story avec validation complétion tâches, mise à jour statut et documentation implémentation.

**MCP à utiliser**
- **serena** : accéder aux fichiers story et mettre à jour statuts
- **mem0** : stocker lessons learned pour réutilisation future
- **time** : horodater complétion pour métriques tracking

**Objectif**
Valider complétion story avec critères acceptation, mettre à jour documentation implémentation et capturer learnings.

**Spécification détaillée**

### Processus complétion story
1. **Validation tâches** : vérifier toutes tâches story actuelle (dans docs/2-current-epic/stories.json) sont ✅ complétées
2. **Validation acceptation** : vérifier critères acceptation rencontrés
3. **Exécution tests** : lancer tests et valider qualité avant marquage complet
4. **Mise à jour statut** : marquer story "✅ Completed" dans stories.json avec métriques complétion
5. **Nettoyage contexte** : supprimer docs/2-current-epic/current-story.json pour effacer sélection actuelle
6. **Documentation implémentation** : mettre à jour IMPLEMENTATION.md avec détails implémentation story
7. **Capture learning** : stocker lessons learned avec mem0

### Quality gates
- Toutes tâches story marquées ✅ completed
- Critères acceptation validés
- Tests exécutés avec succès
- Documentation implémentation mise à jour
- Lessons learned capturés pour apprentissage

### Gestion tâches
- Tâches stockées dans docs/2-current-epic/stories.json (pas todo.json séparés)
- Validation complétion à l'intérieur structure story
- Métriques complétion incluses dans mise à jour statut

**Bornes d'écriture**
* Autorisé : docs/2-current-epic/*, docs/1-project/IMPLEMENTATION.md
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Vérifier toutes tâches story actuelle dans docs/2-current-epic/stories.json sont ✅
2. Valider critères acceptation story rencontrés
3. Exécuter tests et valider qualité implémentation
4. [time] Générer timestamp et métriques complétion
5. [serena] Marquer story "✅ Completed" dans stories.json avec métriques
6. [serena] Supprimer docs/2-current-epic/current-story.json
7. [serena] Mettre à jour IMPLEMENTATION.md avec détails implémentation
8. [mem0] Stocker lessons learned pour réutilisation future
9. Valider conformité schema stories.json

**Points de vigilance**
- Lancer tests et valider critères acceptation avant marquage complet
- Stocker lessons learned avec mem0 pour apprentissage continu
- Tâches dans docs/2-current-epic/stories.json (pas todo.json séparés)
- Capturer métriques complétion précises

**Tests/Validation**
- Vérification complétion toutes tâches story
- Validation critères acceptation rencontrés
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
