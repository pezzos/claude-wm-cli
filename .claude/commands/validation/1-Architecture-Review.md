# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /validation:1-Architecture-Review

**Rôle**
Architecte analyste spécialisé dans l'évaluation architecture complète avec insights codebase MCP-powered et recommandations structurelles.

**Contexte**
Exécution analyse architecture complète avec insights codebase profonds MCP. Évaluation structure, patterns, dépendances, scalabilité, performance, sécurité et maintenabilité.

**MCP à utiliser**
- **consult7** : analyse complète structure codebase
- **context7** : standards architecturaux actuels
- **sequential-thinking** : évaluation architecture systématique
- **mem0** : patterns analyse architecture similaire

**Objectif**
Exécuter revue architecture complète avec insights MCP-powered et recommandations actionnables pour amélioration structurelle.

**Spécification détaillée**

Execute comprehensive architecture analysis with MCP-powered deep codebase insights.

## Pre-Review Intelligence (MANDATORY)
1. **Load Architecture Patterns**: Use `mcp__mem0__search_coding_preferences` to find similar architecture analysis
2. **Full Codebase Analysis**: Use `mcp__consult7__consultation` to map entire project structure
3. **Best Practice Reference**: Use `mcp__context7__` for current architectural standards
4. **Complex Analysis Planning**: Use `mcp__sequential-thinking__` for systematic architecture evaluation

## Comprehensive Architecture Analysis

### 1. Structure Mapping and Pattern Recognition
- **Project Structure Analysis**: Complete codebase organization assessment
- **Design Pattern Identification**: Identify implemented design patterns (MVC, Observer, Factory, etc.)
- **Anti-Pattern Detection**: Locate architectural anti-patterns and code smells
- **Component Relationships**: Map dependencies and coupling between components
- **Layer Architecture**: Evaluate separation of concerns and layer boundaries

### 2. Dependency Analysis
- **Dependency Graph**: Create comprehensive dependency mapping
- **Circular Dependencies**: Identify and flag circular dependency issues
- **Tight Coupling**: Locate tightly coupled components
- **External Dependencies**: Analyze third-party library usage and versions
- **Dependency Injection**: Evaluate DI pattern implementation

### 3. Scalability Assessment
- **Horizontal Scaling**: Evaluate ability to scale across multiple instances
- **Vertical Scaling**: Assess resource utilization and optimization potential
- **Database Scaling**: Analyze data layer scalability and performance
- **Caching Strategy**: Review caching implementation and effectiveness
- **Load Handling**: Evaluate system capacity and bottleneck identification

### 4. Performance Analysis
- **Code Performance**: Identify performance bottlenecks in algorithms and data structures
- **Memory Usage**: Analyze memory allocation and potential leaks
- **I/O Operations**: Review file system and network operation efficiency
- **Database Queries**: Evaluate query performance and optimization opportunities
- **Caching Effectiveness**: Assess current caching strategies

### 5. Security Architecture Review
- **Authentication/Authorization**: Review security implementation patterns
- **Input Validation**: Analyze data validation and sanitization
- **Encryption**: Evaluate data protection mechanisms
- **API Security**: Review API endpoint security measures
- **Vulnerability Assessment**: Identify potential security weaknesses

### 6. Maintainability Evaluation
- **Code Organization**: Assess code structure and modularity
- **Documentation Quality**: Evaluate inline and external documentation
- **Testing Coverage**: Analyze test coverage and testing strategies
- **Code Duplication**: Identify and quantify code duplication
- **Refactoring Opportunities**: Locate areas requiring code improvement

## MCP-Enhanced Analysis Outputs

### Automated Deliverables
- **ARCHITECTURE-ANALYSIS.md**: Comprehensive findings and recommendations
- **DEPENDENCY-GRAPH.md**: Visual and textual dependency mapping
- **PERFORMANCE-REPORT.md**: Performance bottleneck analysis and recommendations
- **SECURITY-AUDIT.md**: Security findings and improvement suggestions
- **MAINTAINABILITY-SCORE.md**: Code quality metrics and improvement roadmap

### Quality Metrics
- **Complexity Score**: Cyclomatic complexity analysis per module
- **Coupling Index**: Measurement of component interdependencies
- **Cohesion Rating**: Evaluation of component internal consistency
- **Technical Debt Score**: Quantified technical debt assessment
- **Maintainability Index**: Overall code maintainability rating

### Recommendation Categories
- **Immediate Actions**: Critical issues requiring urgent attention
- **Short-term Improvements**: Enhancements for next sprint/release
- **Long-term Strategy**: Architectural evolution recommendations
- **Performance Optimizations**: Specific performance improvement suggestions
- **Security Enhancements**: Security hardening recommendations

## Implementation Roadmap
1. **Priority Matrix**: Rank improvements by impact vs. effort
2. **Implementation Phases**: Break improvements into manageable phases
3. **Resource Requirements**: Estimate time and skill requirements
4. **Risk Assessment**: Identify risks associated with each improvement
5. **Success Metrics**: Define measurable success criteria

## Learning Integration
- **Store Architecture Insights**: Use `mcp__mem0__add_coding_preference` to capture analysis patterns
- **Best Practice Documentation**: Save effective architecture evaluation approaches
- **Pattern Recognition**: Document successful architecture patterns for future projects

**Bornes d'écriture**
* Autorisé : docs/1-project/ (ARCHITECTURE-ANALYSIS.md, rapports associés)
* Interdit : code source modification, configuration système

**Étapes**
1. [mem0] Charger patterns analyse architecture similaire
2. [consult7] Mapper structure complète projet
3. [context7] Référencer standards architecturaux actuels
4. [sequential-thinking] Évaluer architecture systématiquement
5. Générer deliverables automatisés
6. [mem0] Mémoriser insights architecture pour futures analyses

**Points de vigilance**
- Insights architecturaux data-driven avec analyse codebase complète
- Recommandations actionnables avec guidance implémentation claire
- Couverture complète : structure, performance, sécurité, maintenabilité
- Roadmap implémentation avec matrice priorité impact/effort

**Tests/Validation**
- Vérification complétude analyse architecture
- Validation pertinence recommandations
- Test actionnabilité roadmap implémentation

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
- 0: Success - Architecture review completed with actionable recommendations
- 1: Needs iteration - Analysis incomplete, requires additional investigation
- 2: Blocked - Unable to analyze due to missing dependencies or access issues
- 3: User input needed - Requires clarification on architecture goals or constraints