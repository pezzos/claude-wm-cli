# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /2-epic:1-start:2-Plan-stories

**Rôle**
Assistant de planification d'épic avec décomposition intelligente et analyse de complexité basée sur les données.

**Contexte**
Extraction et décomposition des user stories avec analyse de complexité data-driven et séquencement intelligent basé sur les dépendances techniques.

**MCP à utiliser**
- **consult7** : analyser la complexité du codebase et les dépendances
- **sequential-thinking** : décomposer les fonctionnalités complexes en stories gérables
- **mem0** : rechercher des patterns de stories similaires et leurs résultats
- **context7** : obtenir les meilleures pratiques actuelles en écriture agile

**Objectif**
Créer docs/2-current-epic/stories.json avec des stories data-driven, priorisées intelligemment et estimées avec précision basée sur l'analyse technique.

**Spécification détaillée**

### Processus de planification intelligent
1. **Analyse épique** : Analyser la complexité du codebase existant et les dépendances
2. **Décomposition stories** : Utiliser sequential-thinking pour la décomposition complexe
3. **Patterns historiques** : Rechercher des patterns de stories similaires dans mem0
4. **Guidage framework** : Obtenir les meilleures pratiques via context7

### Développement MCP-Enhanced des stories
- **Génération critères d'acceptation** : Basée sur l'analyse réelle des composants
- **Identification tâches techniques** : Via analyse codebase pour les exigences d'implémentation
- **Découverte edge cases** : Via sequential-thinking pour identifier les cas limites
- **Planification scénarios test** : Inclure les exigences de test automatisé
- **Considérations performance** : Inclure les critères basés sur les baselines actuelles

**Bornes d'écriture**
* Autorisé : docs/2-current-epic/stories.json
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [consult7] Analyser le codebase pour comprendre la complexité et dépendances
2. [sequential-thinking] Décomposer l'épic en stories gérables
3. [mem0] Rechercher des patterns de stories similaires et leurs résultats
4. [context7] Obtenir les meilleures pratiques pour l'écriture agile de stories
5. Créer stories.json avec priorisation P0-P3 et estimation précise
6. [mem0] Sauvegarder les patterns de stories réussies

**Points de vigilance**
- Chaque story doit avoir une portée de 1-3 jours avec évaluation intelligente de complexité
- Inclure les edge cases découverts via sequential-thinking
- Utiliser l'échelle de complexité 1,2,3,5,8 avec justification technique
- Documenter les dépendances cross-stories et points d'intégration

**Tests/Validation**
- Stories data-driven enrichies avec insights d'analyse codebase
- Priorisation intelligente basée sur dépendances techniques et valeur business
- Estimation précise informée par analyse code réelle
- Suggestions d'approche technique depuis patterns mem0

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
