# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /2-epic:1-start:1-Select-Stories

**Rôle**
Assistant de sélection d'épic avec analyse de priorités et initialisation d'espace de travail.

**Contexte**
Sélection automatique de l'épic suivant basée sur les priorités définies dans epics.json et initialisation complète de l'espace de travail pour commencer le développement.

**MCP à utiliser**
- **serena** : accéder aux fichiers de configuration existants
- **mem0** : mémoriser le contexte épic sélectionné pour la session

**Objectif**
Choisir le prochain épic prioritaire depuis epics.json et initialiser complètement l'espace de travail docs/2-current-epic/.

**Spécification détaillée**

### Processus de sélection
1. Parse epics.json pour trouver l'épic avec la plus haute priorité non démarré (P0 > P1 > P2 > P3)
2. Vérifier que les dépendances sont satisfaites avant sélection
3. Copier les informations épic vers docs/2-current-epic/current-epic.json
4. Créer docs/2-current-epic/CLAUDE.md avec contexte épic complet
5. Marquer l'épic comme "🚧 In Progress" avec date de démarrage

**Bornes d'écriture**
* Autorisé : docs/2-current-epic/, docs/1-project/epics.json
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Lire et parser docs/1-project/epics.json
2. Identifier l'épic prioritaire non démarré avec dépendances satisfaites
3. Copier métadonnées épic vers docs/2-current-epic/current-epic.json
4. Générer docs/2-current-epic/CLAUDE.md avec contexte complet
5. [mem0] Mémoriser le contexte épic pour la session
6. Mettre à jour le statut dans epics.json

**Points de vigilance**
- Vérifier les dépendances avant sélection
- Mettre à jour la date de démarrage dans epics.json
- Assurer la cohérence des métadonnées copiées

**Tests/Validation**
- Validation de l'épic sélectionné selon les critères de priorité
- Vérification de l'initialisation complète de l'espace de travail
- Cohérence entre epics.json et current-epic.json

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