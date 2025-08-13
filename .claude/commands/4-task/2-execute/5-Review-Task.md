# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /4-task:2-execute:5-Review-Task

**Rôle**
Assistant de revue complète avec validation des quality gates et vérification de cohérence du plan.

**Contexte**
Exécution de revue complète avec intelligence pré-revue, validation quality gates, vérification cohérence plan-implémentation et gestion d'itération basée sur résultats qualité.

**MCP à utiliser**
- **mem0** : charger patterns de revue efficaces et capturer approches réussies
- **sequential-thinking** : décomposer exigences complexes en critères vérifiables
- **context7** : obtenir standards qualité actuels et meilleures pratiques

**Objectif**
Effectuer revue complète validant cohérence plan-implémentation, qualité code, sécurité/performance, intégration et documentation avant approbation archivage.

**Spécification détaillée**

### Intelligence pré-revue (MANDATORY)
1. **Patterns revue** : Charger approches revue efficaces depuis mem0
2. **Analyse exigences complexes** : Décomposer exigences via sequential-thinking
3. **Documentation qualité** : Obtenir standards qualité actuels via context7

### Étapes de revue
1. **Cohérence plan-implémentation** : Comparer implémentation vs plan et exigences originales
2. **Assessment qualité code** : Réviser standards, maintenabilité, cohérence architecturale
3. **Revue sécurité & performance** : Valider exigences sécurité et benchmarks performance
4. **Intégration & compatibilité** : Vérifier points intégration et compatibilité backward
5. **Complétude documentation** : Assurer documentation complète et précise
6. **Validation quality gates** : Vérifier tous quality gates satisfaits avant approbation

### Analyse exigences complexes
- Décomposer exigences en critères vérifiables via sequential-thinking
- Mapper quality gates et checkpoints validation
- Vérifier aucune régression fonctionnalité existante
- Valider tous critères acceptation remplis
- Assurer intégrité solution end-to-end

**Bornes d'écriture**
* Autorisé : docs/3-current-task/iterations.json
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [mem0] Charger patterns revue efficaces
2. [sequential-thinking] Décomposer exigences complexes en critères vérifiables
3. [context7] Obtenir standards qualité actuels et meilleures pratiques
4. Comparer implémentation vs plan/exigences originales
5. Assessment qualité code (standards, maintenabilité, architecture)
6. Revue sécurité/performance et validation benchmarks
7. Vérifier intégration et compatibilité backward
8. Valider complétude/précision documentation
9. [mem0] Capturer patterns revue réussis
10. Gérer itération selon résultats

**Points de vigilance**
- Tous quality gates doivent passer avant approbation
- Utiliser sequential-thinking pour scénarios validation complexes
- Si échec revue : mettre à jour iterations.json avec guidance re-planning
- Si succès revue : approuver pour archivage et complétion task
- Aucune limite itération pour revue - continuer jusqu'à standards qualité atteints

**Tests/Validation**
- Cohérence complète plan-implémentation-exigences
- Standards qualité code satisfaits
- Exigences sécurité/performance validées
- Intégration et compatibilité backward vérifiées
- Documentation complète et précise
- Tous critères acceptation validés

**Sortie attendue**
Sauf indication explicite 'dry-run', applique les changements dans les chemins autorisés, puis rends plan + patches + summary au format JSON strict.

CRITICAL: Sortir avec code approprié selon résultats revue :
- EXIT_CODE=0 si revue passe complètement (ready for archiving)
- EXIT_CODE=1 si échec revue mais retry possible (triggers re-planning)
- EXIT_CODE=2 si échec revue pour problèmes fondamentaux (blocked)

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
- 0: Success - review passed completely, ready for archiving
- 1: Needs iteration - review failed but retryable, triggers re-planning  
- 2: Blocked - review failed due to fundamental issues
- 3: User input needed
## JSON Schema Validation
<!-- JSON_SCHEMA_VALIDATION -->

### MANDATORY: Schema Compliance for docs/3-current-task/iterations.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements iterations
```

### Schema-Aware Generation
When updating docs/3-current-task/iterations.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements iterations
```

**MANDATORY REQUIREMENTS:**
1. **$schema field**: The JSON file MUST contain a "$schema" field with the value ".claude/commands/templates/schemas/iterations.schema.json"
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
