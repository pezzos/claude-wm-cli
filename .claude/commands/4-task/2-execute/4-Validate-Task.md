# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /4-task:2-execute:4-Validate-Task

**Rôle**
Assistant de validation complète avec analyse intelligente des parcours utilisateur complexes.

**Contexte**
Exécution de tests compréhensifs avec intelligence pré-validation, analyse de parcours utilisateur complexes et gestion d'itération basée sur les résultats.

**MCP à utiliser**
- **mem0** : charger patterns de validation et capturer approches efficaces
- **sequential-thinking** : décomposer parcours utilisateur complexes en étapes testables
- **context7** : obtenir meilleures pratiques de test actuelles
- **playwright/puppeteer** : tests UI automatisés (si applicable)

**Objectif**
Exécuter validation complète incluant tests automatisés, parcours utilisateur complexes, scénarios manuels avec gestion d'itération et capture d'apprentissage.

**Spécification détaillée**

### Intelligence pré-validation (MANDATORY)
1. **Patterns de test** : Charger approches de validation depuis mem0
2. **Analyse parcours complexes** : Décomposer workflows multi-étapes via sequential-thinking
3. **Documentation test** : Obtenir meilleures pratiques via context7

### Étapes de validation
1. **Révision résultats tests** : Analyser résultats tests automatisés
2. **Tests UI MCP** : Exécuter tests Playwright/Puppeteer automatisés
3. **Tests parcours complexes** : Valider workflows multi-étapes avec sequential-thinking
4. **Scénarios tests manuels** : Exécuter scénarios depuis docs/3-current-task/TEST.md
5. **Performance & sécurité** : Réviser baselines performance et exigences sécurité
6. **Régression visuelle** : Comparer screenshots UI pour cohérence

### Analyse parcours utilisateur complexes
- Décomposer en étapes testables via sequential-thinking
- Mapper dépendances étape par étape et points validation
- Tester scénarios échec à chaque étape
- Valider récupération système après échecs
- Assurer intégrité parcours end-to-end

**Bornes d'écriture**
* Autorisé : docs/3-current-task/iterations.json, docs/3-current-task/TEST.md
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [mem0] Charger patterns validation et approches efficaces
2. [sequential-thinking] Décomposer parcours utilisateur complexes
3. [context7] Obtenir meilleures pratiques test actuelles
4. Réviser résultats tests automatisés
5. Exécuter tests UI MCP si applicable
6. Valider parcours complexes décomposés
7. Exécuter scénarios tests manuels
8. [mem0] Capturer patterns validation réussis
9. Gérer itération selon résultats

**Points de vigilance**
- Tous tests doivent passer avant de continuer
- Utiliser sequential-thinking pour scénarios validation complexes
- Si échec et iterations < 3 : mettre à jour iterations.json et guider prochaine itération
- Si échec ET iterations = 3 : documenter blocage et demander aide
- Si succès : valider statut complétion et fournir rapport validation final

**Tests/Validation**
- Tests automatisés, manuels, performance et sécurité
- Validation parcours utilisateur complexes via sequential-thinking
- Capture patterns validation réussis dans mem0
- Métriques performance pour tests régression futurs

**Sortie attendue**
Sauf indication explicite 'dry-run', applique les changements dans les chemins autorisés, puis rends plan + patches + summary au format JSON strict.

CRITICAL: Sortir avec code approprié selon résultats validation :
- EXIT_CODE=0 si tous tests passent (succès)
- EXIT_CODE=1 si échec mais retry possible (needs iteration)
- EXIT_CODE=2 si blocage fondamental (blocked)

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
- 0: Success - validation passed completely
- 1: Needs iteration - validation failed but retryable
- 2: Blocked - validation failed due to fundamental issues
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
