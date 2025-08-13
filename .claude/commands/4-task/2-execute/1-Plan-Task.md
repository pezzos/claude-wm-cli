# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /4-task:2-execute:1-Plan-Task

**Rôle**
Assistant planification intelligent avec analyse approfondie et stratégie de tests de régression.

**Contexte**
Génération de plan d'implémentation complet avec analyse intelligente basée sur les patterns existants. Le contexte de tâche est pré-populé dans docs/3-current-task/ par le preprocessing. L'accent est mis sur la planification intelligente et la stratégie de tests de régression obligatoire.

**MCP à utiliser**
- **mem0** : rechercher solutions similaires avec `mcp__mem0__search_coding_preferences`
- **context7** : documentation et patterns du repo via docs/KB/ADR
- **sequential-thinking** : décomposition complexe si >5 étapes
- **serena** : réutiliser documentation existante

**Objectif**
Enrichir docs/3-current-task/current-task.json avec approche complète, modifications de fichiers, étapes d'implémentation et stratégie de tests de régression obligatoire.

**Spécification détaillée**

### Contexte disponible
- `docs/3-current-task/current-task.json` - Contexte de tâche (pré-populé)
- `docs/3-current-task/iterations.json` - Suivi des itérations (pré-populé)
- Structures de templates prêtes pour génération de contenu

### Focus planification intelligente
1. **Recherche patterns** : mem0 pour solutions éprouvées
2. **Enrichissement contexte** : docs/3-current-task/current-task.json complet
3. **Planification tests régression** : OBLIGATOIRE pour validation continue
4. **Documentation risques** : hypothèses et limitations claires

### Tests de régression (OBLIGATOIRES)
- **Tests automatisés** : définir tests MCP UI (Playwright/Puppeteer)
- **Couverture tests** : parcours utilisateur nécessitant validation
- **Baselines performance** : métriques à maintenir (temps chargement, réactivité)
- **Régression visuelle** : composants UI nécessitant comparaison screenshots
- **Points d'intégration** : APIs/services nécessitant validation continue
- **Standards accessibilité** : exigences a11y pour conformité continue

### Stratégie intégration tests
- **Tests unitaires** : validation niveau fonction
- **Tests d'intégration** : validation interaction composants
- **Automatisation UI** : tests navigateur avec outils MCP
- **Tests performance** : monitoring automatisé performance
- **Tests sécurité** : scan vulnérabilités continu
- **Scénarios tests manuels** : parcours critiques nécessitant validation humaine

**Bornes d'écriture**
* Autorisé : docs/3-current-task/, docs/templates/, fichiers JSON/MD de suivi
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [mem0] Rechercher patterns existants avant nouvelle approche
2. [context7] Charger contexte repo et ADR pertinents  
3. [sequential-thinking] Décomposer si planification >5 étapes
4. Enrichir docs/3-current-task/current-task.json avec approche complète
5. Planifier stratégie tests régression (OBLIGATOIRE)
6. Documenter risques et hypothèses clairement

**Points de vigilance**
- Planifier pour maximum 3 itérations
- Toujours inclure stratégie tests de régression
- Rechercher patterns existants avant planifier nouvelle approche
- Documenter risques et hypothèses clairement
- Assurer conformité schéma JSON strict

**Tests/Validation**
- Validation schéma current-task.json avec `.claude/commands/tools/schema-enforcer.sh`
- Vérification inclusion stratégie tests de régression
- Cohérence avec patterns mem0 identifiés
- Plan réalisable en 3 itérations maximum

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