# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /4-task:2-execute:2-Test-design

**Rôle**
Assistant de conception de stratégie de test avec capacités MCP pour tests UI automatisés.

**Contexte**
Conception d'une stratégie de test complète incluant tests traditionnels et tests UI automatisés MCP-powered pour assurer la qualité et la régression continue.

**MCP à utiliser**
- **mem0** : rechercher des approches de test similaires dans les préférences
- **consult7** : identifier les éléments UI nécessitant des tests
- **context7** : obtenir les meilleures pratiques pour les frameworks de test actuels
- **playwright/puppeteer** : pour tests UI automatisés (si applicable)

**Objectif**
Enrichir docs/3-current-task/TEST.md avec une stratégie de test complète couvrant tous les cas (happy path, edge cases, erreurs) avec intégration MCP UI quand applicable.

**Spécification détaillée**

### Processus de conception de test
1. **Intelligence pré-conception** : Charger patterns de test depuis mem0
2. **Analyse composants UI** : Identifier éléments UI via consult7
3. **Documentation framework** : Obtenir meilleures pratiques via context7
4. **Génération scénarios** : Créer scénarios de test complets dans TEST.md

### Catégories de test à inclure
- **Tests manuels** : Parcours utilisateur critiques nécessitant validation humaine
- **Tests unitaires automatisés** : Test niveau fonction
- **Tests intégration automatisés** : Test interaction composants
- **Tests UI automatisés** : Tests interaction navigateur (MCP-powered)
- **Tests performance** : Scénarios charge et stress
- **Tests sécurité** : Validation entrées et vulnérabilités

### Intégration MCP UI (si applicable)
- **Tests Playwright** : Pour applications React/Vue/Angular via mcp__playwright__browser_*
- **Tests Puppeteer** : Pour applications Node.js via mcp__puppeteer__puppeteer_*
- **Régression visuelle** : Tests basés screenshots pour cohérence UI
- **Tests cross-browser** : Tests automatisés multi-navigateurs
- **Tests performance** : Validation Core Web Vitals et temps chargement
- **Tests accessibilité** : Validation a11y automatisée dans suite

**Bornes d'écriture**
* Autorisé : docs/3-current-task/TEST.md
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [mem0] Rechercher approches de test similaires dans préférences
2. [consult7] Analyser composants UI nécessitant tests
3. [context7] Obtenir meilleures pratiques frameworks test actuels
4. Enrichir docs/3-current-task/TEST.md avec scénarios complets
5. Définir approches test traditionnelles (unité, intégration)
6. Concevoir tests UI automatisés MCP si composants UI présents
7. Planifier données test et validation complètes
8. Concevoir scénarios échec et gestion erreur

**Points de vigilance**
- Couvrir happy path, edge cases et conditions erreur
- Concevoir tests avant implémentation
- Inclure automatisation UI MCP pour interfaces web (régression continue)
- Assurer cohérence avec patterns de test existants

**Tests/Validation**
- Stratégie de test complète dans docs/3-current-task/TEST.md
- Intégration MCP UI pour tests automatisés (si applicable)
- Couverture scénarios : succès, échec, edge cases
- Plan données test et validation

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