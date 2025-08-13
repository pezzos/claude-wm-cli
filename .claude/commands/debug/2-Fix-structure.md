# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /debug:2-Fix-structure

**Rôle**
Assistant de réparation structure projet avec restauration fichiers, alignement branches et consistency git.

**Contexte**
Réparation complète structure projet incluant restauration dossiers/fichiers manquants, correction misalignments branches et validation consistency git avec preservation travail existant.

**MCP à utiliser**
- **serena** : analyser structure existante et identifier réparations nécessaires
- **github** : gérer alignement branches et consistency git si nécessaire
- **mem0** : capturer patterns réparation pour prévention future

**Objectif**
Réparer structure projet manquante en restaurant dossiers/fichiers, corrigeant misalignments branches et assurant consistency git avec validation complète.

**Spécification détaillée**

### Processus réparation structure
1. **Backup et restauration** : Créer backup puis restaurer structure dossiers et fichiers requis manquants
2. **Templates documentation** : Utiliser templates depuis ./commands/templates/ pour recréer documentation manquante
3. **Alignement branches** : Corriger problèmes alignement branches et incohérences état git
4. **Validation réparations** : Valider réparations et commit fixes structurels avec message descriptif

**Bornes d'écriture**
* Autorisé : tous fichiers projet selon besoins réparation
* Interdit : modification travail utilisateur existant sans backup

**Étapes**
1. Créer backup complet avant réparations
2. [serena] Analyser structure existante et identifier manquants
3. Restaurer structure dossiers manquante (docs/, etc.)
4. Utiliser templates .claude/commands/templates/ pour recréer docs
5. [github] Corriger alignement branches et état git si nécessaire
6. Valider toutes réparations avant commit
7. [mem0] Capturer patterns réparation pour prévention
8. Commit fixes structurels avec message descriptif

**Points de vigilance**
- Toujours backup avant réparations
- Préserver travail existant et changements utilisateur
- Valider toutes réparations avant commit
- Utiliser templates pour consistency documentation

**Tests/Validation**
- Backup créé avant modifications
- Structure dossiers restaurée complètement
- Documentation recréée depuis templates valides
- Alignement branches et consistency git corrigés
- Réparations validées avant commit
- Patterns réparation capturés dans mem0

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