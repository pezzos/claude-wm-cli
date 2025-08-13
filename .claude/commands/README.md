# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /commands:index

**Rôle**
Guide d'organisation des commandes système Claude WM CLI avec gestion de la structure hiérarchique des commandes.

**Contexte**
Index principal des commandes Claude organisées par catégorie avec templates de fichiers intégrés. Documentation système read-only avec directives de personnalisation utilisateur.

**MCP à utiliser**
- **serena** : navigation structure commandes
- **mem0** : mémoriser patterns d'organisation commandes

**Objectif**
Documenter la structure des commandes système et orienter vers personnalisation utilisateur appropriée.

**Spécification détaillée**

### Structure des commandes
📦 **Commandes claude par défaut fournies par claude-wm-cli**

## 📁 Structure

- `templates/` - Templates de fichiers (JSON, MD) utilisés par le preprocessing
- Autres dossiers : Commandes claude organisées par catégorie

## ⚠️ Ne pas modifier

Ces fichiers sont gérés automatiquement. Pour personnaliser :

1. Copiez vers `../../user/commands/`
2. Modifiez votre copie
3. Lancez `claude-wm config sync`

**Bornes d'écriture**
* Autorisé : documentation utilisateur seulement
* Interdit : fichiers système .claude/commands/ (read-only)

**Étapes**
1. Consultation de la structure commandes
2. Orientation vers personnalisation appropriée
3. Documentation des patterns d'organisation

**Points de vigilance**
- Structure système en read-only
- Personnalisation via user/commands/ seulement
- Synchronisation obligatoire après modifications

**Tests/Validation**
- Vérification intégrité structure système
- Validation chemins personnalisation utilisateur

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