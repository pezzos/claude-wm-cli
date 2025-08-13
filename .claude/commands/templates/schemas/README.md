# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /templates:schemas:README

**Rôle**
Gestionnaire schémas validation JSON avec documentation structures données et intégration workflow système.

**Contexte**
Schémas JSON Schema pour validation structure fichiers données générés par commandes workflow Claude WM CLI. Documentation schémas disponibles avec patterns validation et intégration workflow.

**MCP à utiliser**
- **serena** : analyser schémas existants et structure validation
- **mem0** : mémoriser patterns validation efficaces

**Objectif**
Documenter système schémas validation JSON et guider intégration workflow pour validation données complète.

**Spécification détaillée**

# Schémas de Validation JSON

Ce dossier contient les schémas JSON Schema pour valider la structure des fichiers de données générés par les commandes du workflow Claude WM CLI.

## Schémas Disponibles

### 1. `current-epic.schema.json`
Valide la structure des données d'epic actuel.
- **Fichier validé**: `current-epic.json`
- **Structure**: Object avec propriété `epic` contenant id, title, description, status, dates, priority et dependencies

### 2. `current-story.schema.json`
Valide la structure des données de story actuelle.
- **Fichier validé**: `current-story.json`
- **Structure**: Object avec propriété `story` contenant id, title, description, epic_id, status et priority

### 3. `epics.schema.json`
Valide la collection complète des epics du projet.
- **Fichier validé**: `epics.json`
- **Structure**: Array d'epics avec contexte projet complet, success criteria, dependencies, blockers

### 4. `stories.schema.json`
Valide la collection complète des stories.
- **Fichier validé**: `stories.json`  
- **Structure**: Object avec stories indexées par ID et contexte epic parent

### 5. `iterations.schema.json`
Valide les données d'itération de tâches avec historique complet.
- **Fichier validé**: `iterations.json`
- **Structure**: Context de tâche, array d'itérations, outcome final et recommandations

### 6. `current-task.schema.json`
Valide les données de tâche critique en cours.
- **Fichier validé**: `docs/3-current-task/current-task.json`
- **Structure**: Tâche détaillée avec analysis, reproduction, investigation, implementation et resolution

### 7. `metrics.schema.json`
Valide les métriques de performance du projet.
- **Fichier validé**: `metrics.json`
- **Structure**: Overview projet, epic actuel, performance iterations, analytics temporelles, qualité et équipe

## Utilisation

### Validation avec JSON Schema

Les schémas suivent la spécification [JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12/schema).

#### Exemple avec Node.js et AJV:

```javascript
const Ajv = require('ajv');
const addFormats = require('ajv-formats');
const fs = require('fs');

const ajv = new Ajv();
addFormats(ajv);

// Charger le schéma
const schema = JSON.parse(fs.readFileSync('schemas/current-epic.schema.json'));
const validate = ajv.compile(schema);

// Valider les données
const data = JSON.parse(fs.readFileSync('current-epic.json'));
const valid = validate(data);

if (!valid) {
    console.error('Validation errors:', validate.errors);
} else {
    console.log('Data is valid!');
}
```

#### Exemple avec Python et jsonschema:

```python
import json
import jsonschema
from jsonschema import validate

# Charger le schéma
with open('schemas/current-epic.schema.json') as f:
    schema = json.load(f)

# Charger les données
with open('current-epic.json') as f:
    data = json.load(f)

# Valider
try:
    validate(instance=data, schema=schema)
    print("Data is valid!")
except jsonschema.exceptions.ValidationError as e:
    print(f"Validation error: {e}")
```

### Intégration dans le Workflow

Les schémas peuvent être utilisés pour:

1. **Validation à la génération**: Valider les données avant écriture des fichiers
2. **Validation à la lecture**: Vérifier l'intégrité des données existantes
3. **Tests automatisés**: Intégrer dans les tests unitaires
4. **IDE Support**: Auto-complétion et validation en temps réel
5. **CI/CD Pipeline**: Validation automatique dans les pipelines

### Exemple d'Intégration Workflow

```go
// Exemple d'intégration en Go
func ValidateCurrentEpic(data []byte) error {
    // Charger le schéma
    schema, err := os.ReadFile("schemas/current-epic.schema.json")
    if err != nil {
        return err
    }
    
    // Valider avec une lib JSON Schema Go
    loader := gojsonschema.NewBytesLoader(schema)
    documentLoader := gojsonschema.NewBytesLoader(data)
    
    result, err := gojsonschema.Validate(loader, documentLoader)
    if err != nil {
        return err
    }
    
    if !result.Valid() {
        return fmt.Errorf("validation failed: %v", result.Errors())
    }
    
    return nil
}
```

## Patterns de Validation

### IDs et References
- **Epic IDs**: Pattern `^EPIC-[0-9]{3}$` (ex: EPIC-001)
- **Story IDs**: Pattern `^STORY-[0-9]{3}$` (ex: STORY-001)
- **Task IDs**: Pattern `^TASK-[0-9]{3}$` (ex: TASK-001)
- **Story-Task IDs**: Pattern `^STORY-[0-9]{3}-TASK-[0-9]+$` (ex: STORY-001-TASK-1)

### Status Values
- Standard: `["todo", "in_progress", "done", "blocked"]`
- Task types: `["bug", "feature", "enhancement", "refactor", "documentation"]`
- Priorities: `["low", "medium", "high", "critical"]`

### Dates et Timestamps
- Format ISO 8601: `YYYY-MM-DDTHH:mm:ss+TZ`
- Dates simples: `YYYY-MM-DD`

### Emojis de Status
Les schémas acceptent les emojis standard pour les status visuels:
- ✅ `done` / `completed`
- 🚧 `in_progress`
- 📋 `todo` / `planned`
- ❌ `failed` / `error`

## Maintenance

Lors de modifications des structures de données:

1. **Mettre à jour le schéma correspondant**
2. **Tester la validation avec les données existantes**
3. **Documenter les changements dans ce README**
4. **Versionner les changements breaking**

Les schémas suivent le versioning sémantique et doivent maintenir la rétro-compatibilité autant que possible.

**Bornes d'écriture**
* Autorisé : documentation schémas, mise à jour patterns validation
* Interdit : modification schémas sans tests compatibilité

**Étapes**
1. [serena] Analyser schémas existants et structure validation
2. Documenter schémas disponibles et patterns
3. Valider intégration workflow appropriée
4. [mem0] Mémoriser patterns validation efficaces

**Points de vigilance**
- Documentation schémas complète et à jour
- Patterns validation appropriés pour structures données
- Intégration workflow systématique
- Versioning sémantique et rétro-compatibilité

**Tests/Validation**
- Vérification complétude documentation schémas
- Test patterns validation avec données existantes
- Validation intégration workflow

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