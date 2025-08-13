# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /templates:README

**Rôle**
Gestionnaire template README projet avec génération documentation standardisée et substitution variables dynamiques.

**Contexte**
Template README projet standard avec placeholders variables {{PROJECT_NAME}}, features, installation, usage et documentation support. Structure standardisée pour génération documentation projet.

**MCP à utiliser**
- **serena** : analyser structure projet pour variables appropriées
- **mem0** : mémoriser patterns README efficaces

**Objectif**
Fournir template README standardisé avec substitution variables pour génération documentation projet cohérente.

**Spécification détaillée**

### Template README standard
# {{PROJECT_NAME}}

> {{PROJECT_TAGLINE}}

## What is {{PROJECT_NAME}}?

{{PROJECT_DESCRIPTION}}

### Key Features
- ✨ **Feature 1**: Brief description of main feature
- 🚀 **Feature 2**: Brief description of second feature  
- 🔒 **Feature 3**: Brief description of third feature
- 📊 **Feature 4**: Brief description of fourth feature

## Quick Start

### Prerequisites
- {{PREREQUISITE_1}}
- {{PREREQUISITE_2}}
- {{PREREQUISITE_3}}

### Installation

```bash
# Clone the repository
git clone {{REPO_URL}}
cd {{PROJECT_NAME}}

# Install dependencies
{{INSTALL_COMMAND}}

# Setup environment
{{SETUP_COMMAND}}

# Start the application
{{START_COMMAND}}
```

### First Steps
1. **Setup your account**: {{FIRST_STEP_DESCRIPTION}}
2. **Configure settings**: {{SECOND_STEP_DESCRIPTION}}
3. **Start using**: {{THIRD_STEP_DESCRIPTION}}

## How to Use

### Basic Usage
{{BASIC_USAGE_DESCRIPTION}}

```bash
# Example command
{{EXAMPLE_COMMAND}}
```

### Common Tasks
- **{{TASK_1_NAME}}**: {{TASK_1_DESCRIPTION}}
- **{{TASK_2_NAME}}**: {{TASK_2_DESCRIPTION}}  
- **{{TASK_3_NAME}}**: {{TASK_3_DESCRIPTION}}

## Support & Documentation

- 📖 **User Guide**: [Link to detailed user guide]
- ❓ **FAQ**: [Link to frequently asked questions]
- 🐛 **Report Issues**: [Link to issue tracker]
- 💬 **Community**: [Link to community forum/chat]

## Project Status

{{PROJECT_STATUS_BADGE}}

**Current Version**: {{VERSION}}  
**Last Updated**: {{LAST_UPDATE}}

---

*For technical documentation and development setup, see [CONTRIBUTING.md](CONTRIBUTING.md)*

**Bornes d'écriture**
* Autorisé : génération README.md projet avec substitution variables
* Interdit : modification templates sans autorisation utilisateur

**Étapes**
1. [serena] Analyser structure projet pour extraction variables
2. Appliquer template avec substitution variables appropriées
3. [mem0] Mémoriser patterns génération README efficaces
4. Valider cohérence documentation générée

**Points de vigilance**
- Substitution complète variables {{PLACEHOLDER}}
- Cohérence structure projet analysée
- Documentation support appropriée
- Version et status projet actuels

**Tests/Validation**
- Vérification substitution variables complète
- Validation cohérence structure README
- Test lisibilité documentation générée

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