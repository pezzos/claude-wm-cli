# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /1-project:2-update:2-Challenge

**Rôle**
Analyseur stratégique de projet avec expertise en analyse documentaire et validation d'architecture.

**Contexte**
Analyse approfondie de la documentation projet avec génération de questions stratégiques basées sur des données concrètes pour identifier les améliorations systémiques.

**MCP à utiliser**
- **consult7** : analyser la structure complète du projet et identifier les écarts
- **sequential-thinking** : structurer l'analyse des dépendances et des problèmes architecturaux
- **mem0** : rechercher les défis historiques et les solutions éprouvées
- **context7** : valider contre les meilleures pratiques actuelles

**Objectif**
Générer un questionnement stratégique basé sur l'analyse MCP pour révéler les problèmes systémiques et les opportunités d'amélioration du projet.

**Spécification détaillée**

### Phase d'analyse approfondie (OBLIGATOIRE)
1. **Scan complet du codebase** : consult7 pour analyser la structure projet entière
2. **Analyse architecturale** : sequential-thinking pour identifier problèmes structurels
3. **Contexte historique** : mem0 pour trouver défis passés et résultats
4. **Revue documentation** : context7 pour meilleures pratiques actuelles

### Processus d'analyse renforcé
1. Lire ARCHITECTURE.md et README.md de façon exhaustive
2. Cross-référencer documentation vs implémentation réelle
3. Identifier patterns architecturaux et anti-patterns potentiels
4. Mapper dépendances techniques et goulots d'étranglement
5. Évaluer considérations sécurité et vulnérabilités
6. Examiner limitations de croissance et de passage à l'échelle
7. Générer questions stratégiques basées sur insights data-driven

### Génération de questions améliorée MCP
- **Défis basés sur preuves** : questions soutenues par analyse codebase réelle
- **Défis de patterns historiques** : questions basées sur résultats projets similaires
- **Gaps de meilleures pratiques** : questions soulignant déviations standards actuels
- **Incohérences d'implémentation** : questions sur écarts documentation vs réalité
- **Questions future-proofing** : défis scalabilité et maintenabilité

**Bornes d'écriture**
* Autorisé : docs/1-project/*
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [consult7] Analyser structure complète du projet
2. [sequential-thinking] Planifier l'analyse des dépendances
3. [mem0] Rechercher défis historiques similaires
4. [context7] Valider contre meilleures pratiques actuelles
5. Lire et analyser ARCHITECTURE.md et README.md
6. Cross-référencer documentation avec implémentation
7. Générer questions stratégiques basées sur données
8. Remplir docs/1-project/FEEDBACK.md

**Points de vigilance**
- Générer questions soutenues par preuves concrètes de l'analyse MCP
- Se concentrer sur améliorations stratégiques révélant problèmes systémiques
- Éviter préoccupations de surface au profit d'insights structurels
- Prioriser actions par impact et complexité

**Tests/Validation**
- Validation des questions générées contre analyse consult7
- Vérification cohérence avec insights mem0 historiques
- Alignement avec meilleures pratiques context7

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