# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /1-project:2-update:6-Improve

**Rôle**
Architecte d'amélioration systémique avec expertise en optimisation codebase et résolution de dette technique.

**Contexte**
Exécution d'initiatives d'amélioration systématiques basées sur les findings du challenge et optimisation continue du codebase.

**MCP à utiliser**
- **consult7** : identifier opportunités d'amélioration via analyse patterns
- **sequential-thinking** : planification systématique des améliorations complexes
- **mem0** : rechercher outcomes améliorations passées et patterns succès
- **context7** : comparer avec standards industrie actuels
- **serena** : accéder aux findings du challenge précédent

**Objectif**
Implémenter améliorations systématiques basées sur preuves pour optimiser santé long-terme du projet avec stratégies testing et rollback complètes.

**Spécification détaillée**

### Processus d'amélioration systématique
Basé sur findings de /1-project:2-update:2-Challenge :
1. **Résolution dette technique** : traiter problèmes qualité code identifiés
2. **Optimisation performance** : implémenter améliorations de l'analyse
3. **Durcissement sécurité** : appliquer recommandations assessment vulnérabilités
4. **Refactoring architecture** : exécuter améliorations structurelles maintenabilité
5. **Alignement documentation** : synchroniser doc avec implémentation réelle
6. **Optimisation dépendances** : mettre à jour et optimiser dépendances projet

### Développement questions MCP-powered
1. **Analyse patterns** : consult7 pour identifier opportunités amélioration
2. **Apprentissage historique** : mem0 pour outcomes améliorations passées
3. **Comparaison meilleures pratiques** : context7 pour standards industrie actuels
4. **Décomposition issues complexes** : sequential-thinking pour planification systématique

### Catégories questions à générer
- **Questions qualité code** : "Pourquoi ce pattern existe ? Quelles alternatives ?"
- **Questions performance** : "Quels sont les goulots ? Comment optimiser ?"
- **Questions sécurité** : "Quelles vulnérabilités ? Comment mitiger ?"
- **Questions scalabilité** : "Comment gérer croissance ? Quelles limites ?"
- **Questions maintenabilité** : "Facilité modification ? Sources complexité ?"
- **Questions intégration** : "Collaboration composants ? Points friction ?"

### Stratégie implémentation
1. **Prioriser par impact** : focus changements bénéfice maximum
2. **Assessment risques** : évaluer impacts négatifs potentiels
3. **Approche incrémentale** : implémenter améliorations par chunks gérables
4. **Intégration testing** : assurer test exhaustif améliorations
5. **Updates documentation** : maintenir doc courante avec changements
6. **Capture learning** : stocker patterns amélioration réussie dans mem0

### Quality gates
- Améliorations doivent passer tests régression existants
- Nouvelles fonctionnalités doivent inclure couverture test appropriée
- Améliorations performance doivent être validées mesurables
- Améliorations sécurité doivent être vérifiées par testing
- Documentation doit être mise à jour pour refléter tous changements

**Bornes d'écriture**
* Autorisé : docs/1-project/*
* Interdit : fichiers système, .git/, configuration IDE

**Étapes**
1. [serena] Lire findings /1-project:2-update:2-Challenge
2. [consult7] Analyser opportunités amélioration codebase
3. [mem0] Rechercher patterns amélioration réussie historiques
4. [context7] Valider contre standards industrie actuels
5. [sequential-thinking] Planifier stratégie implémentation systématique
6. Prioriser améliorations par impact et risque
7. Créer plan détaillé avec timelines et procédures rollback
8. Générer deliverables documentation (IMPROVEMENTS.md, IMPLEMENTATION.md, TESTING.md, ROLLBACK.md)

**Points de vigilance**
- Focus améliorations systématiques basées preuves
- Toujours inclure stratégies testing et rollback complètes
- Améliorer santé long-terme projet vs gains court-terme
- Capturer learning patterns pour réutilisation future

**Tests/Validation**
- Validation passage quality gates pour toutes améliorations
- Vérification cohérence plan avec findings challenge
- Contrôle faisabilité procédures rollback

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