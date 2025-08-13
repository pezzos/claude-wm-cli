# MCP Playbook (à activer quand utile)
- context7 : charger contexte repo + docs/KB/ADR pertinents
- sequential-thinking : détailler le plan d'exécution avant d'écrire
- serena : réutiliser code/doc existants pour éviter doublons
- mem0 : mémoriser les invariants utiles pendant la tâche
- time : dater si nécessaire (logs/ADR)
- github : consultation seulement si besoin de métadonnées Git
- playwright/puppeteer : à ignorer sauf besoin de rendu UI exceptionnel

# /learning:dashboard

**Rôle**
Analyste apprentissage système avec génération analytics complètes et insights mémoire pour optimisation performance continue.

**Contexte**
Analytics apprentissage complètes depuis mémoire avec patterns succès/échec, métriques performance, tendances vélocité et recommandations optimisation actionnables.

**MCP à utiliser**
- **mem0** : analyser toute mémoire pour patterns apprentissage
- **sequential-thinking** : traiter analytics méthodiquement
- **time** : tracking temporel pour tendances

**Objectif**
Générer dashboard apprentissage complet avec analytics performance, patterns succès/échec et recommandations optimisation système.

**Spécification détaillée**

Show comprehensive learning analytics and insights from memory.

## Model Selection
**Uses Claude Sonnet** for analytics processing:
- Pattern recognition and analysis
- Statistical calculations
- Insight generation
- Dashboard visualization

## What it does
- Analyzes complete memory for learning patterns
- Identifies success and failure patterns
- Calculates performance metrics and trends
- Provides actionable optimization recommendations
- Shows learning system effectiveness

## Implementation
```javascript
// Generate comprehensive learning dashboard
async function generateLearningDashboard() {
  console.log('🧠 LEARNING DASHBOARD');
  console.log('====================\n');
  
  // Get all memory for analysis
  const allMemory = await mcp__mem0__get_all_coding_preferences();
  const analysis = await analyzeMemoryLearnings();
  
  // 1. Performance Metrics
  console.log('📊 PERFORMANCE METRICS:');
  const iterationCount = countIterations(allMemory);
  const successRate = calculateSuccessRate(analysis);
  const avgDuration = calculateAverageDuration(analysis);
  const learningEfficacy = calculateLearningEfficacy(analysis);
  
  console.log(`- Total iterations analyzed: ${iterationCount}`);
  console.log(`- Success rate: ${successRate}%`);
  console.log(`- Average task duration: ${avgDuration}min`);
  console.log(`- Learning efficacy: ${learningEfficacy}/10\n`);
  
  // 2. Success Patterns
  console.log('🎯 TOP SUCCESS PATTERNS:');
  const topSuccessPatterns = analysis.successPatterns.slice(0, 3);
  topSuccessPatterns.forEach(pattern => {
    console.log(`- ${pattern.pattern}: ${pattern.factors.join(', ')}`);
  });
  console.log('');
  
  // 3. Failure Patterns
  console.log('⚠️ TOP FAILURE PATTERNS:');
  const topFailurePatterns = analysis.failurePatterns.slice(0, 3);
  topFailurePatterns.forEach(failure => {
    console.log(`- ${failure.failure}: ${failure.cause}`);
  });
  console.log('');
  
  // 4. Velocity Analysis
  console.log('📈 VELOCITY ANALYSIS:');
  const velocityTrends = analysis.velocityTrends;
  console.log(`- Trend: ${velocityTrends.direction} (${velocityTrends.change}%)`);
  console.log(`- Current velocity: ${calculateCurrentVelocity(allMemory)} tasks/hour`);
  console.log(`- Quality trend: ${calculateQualityTrend(allMemory)}`);
  console.log('');
  
  // 5. Learning Insights
  console.log('💡 LEARNING INSIGHTS:');
  const recentLearnings = extractRecentLearnings(allMemory);
  recentLearnings.forEach(learning => {
    console.log(`- ${learning.insight} (${learning.context})`);
  });
  console.log('');
  
  // 6. Recommendations
  console.log('🚀 RECOMMENDED OPTIMIZATIONS:');
  const optimizations = analysis.processOptimizations;
  optimizations.forEach(opt => {
    console.log(`- ${opt}`);
  });
  console.log('');
  
  // 7. Technology Effectiveness
  console.log('🔧 TECHNOLOGY EFFECTIVENESS:');
  const techEffectiveness = analyzeTechEffectiveness(allMemory);
  Object.entries(techEffectiveness).forEach(([tech, score]) => {
    console.log(`- ${tech}: ${score}/10`);
  });
  console.log('');
  
  // 8. Learning System Health
  console.log('⚡ LEARNING SYSTEM HEALTH:');
  const systemHealth = assessLearningSystemHealth(analysis);
  console.log(`- Pattern recognition: ${systemHealth.patternRecognition}/10`);
  console.log(`- Adaptation capability: ${systemHealth.adaptationCapability}/10`);
  console.log(`- Memory utilization: ${systemHealth.memoryUtilization}/10`);
  console.log(`- Overall health: ${systemHealth.overallHealth}/10\n`);
  
  // 9. Next Steps
  console.log('📋 NEXT STEPS:');
  const nextSteps = generateNextSteps(analysis);
  nextSteps.forEach(step => {
    console.log(`- ${step}`);
  });
  
  // Store dashboard generation in memory
  await mcp__mem0__add_coding_preference({
    text: `LEARNING_DASHBOARD_GENERATED: ${new Date().toISOString()}
Iterations Analyzed: ${iterationCount}
Success Rate: ${successRate}%
Learning Efficacy: ${learningEfficacy}/10
System Health: ${systemHealth.overallHealth}/10
Key Insights: ${recentLearnings.length} new insights
Optimizations: ${optimizations.length} recommendations
Status: Learning system operational and improving`
  });
}

// Helper functions for dashboard analytics
function countIterations(memory) {
  return memory.filter(entry => 
    entry.includes('TASK_PROGRESS:') || 
    entry.includes('LEARNING_ITERATION:')
  ).length;
}

function calculateSuccessRate(analysis) {
  const totalTasks = analysis.successPatterns.length + analysis.failurePatterns.length;
  if (totalTasks === 0) return 100;
  return Math.round((analysis.successPatterns.length / totalTasks) * 100);
}

function calculateAverageDuration(analysis) {
  const allEntries = [...analysis.successPatterns, ...analysis.failurePatterns];
  const durations = allEntries.map(entry => {
    const match = entry.pattern?.match(/Duration:\s*(\d+)\s*minutes/) || 
                  entry.failure?.match(/Duration:\s*(\d+)\s*minutes/);
    return match ? parseInt(match[1]) : 0;
  }).filter(d => d > 0);
  
  if (durations.length === 0) return 0;
  return Math.round(durations.reduce((a, b) => a + b, 0) / durations.length);
}

function calculateLearningEfficacy(analysis) {
  const hasSuccessPatterns = analysis.successPatterns.length > 0;
  const hasFailurePatterns = analysis.failurePatterns.length > 0;
  const hasVelocityTrends = analysis.velocityTrends.direction !== 'stable';
  const hasOptimizations = analysis.processOptimizations.length > 0;
  
  let score = 0;
  if (hasSuccessPatterns) score += 3;
  if (hasFailurePatterns) score += 2; // Failure patterns are learning too
  if (hasVelocityTrends) score += 2;
  if (hasOptimizations) score += 3;
  
  return Math.min(score, 10);
}

function calculateCurrentVelocity(memory) {
  const recentTasks = memory.filter(entry => 
    entry.includes('TASK_PROGRESS:') &&
    entry.includes('completed') &&
    isWithinLast24Hours(entry)
  );
  
  return recentTasks.length > 0 ? recentTasks.length : 0;
}

function calculateQualityTrend(memory) {
  const qualityScores = memory.filter(entry => 
    entry.includes('Quality:')
  ).map(entry => {
    const match = entry.match(/Quality:\s*(\d+)/);
    return match ? parseInt(match[1]) : 0;
  });
  
  if (qualityScores.length < 3) return 'stable';
  
  const recent = qualityScores.slice(-3);
  const previous = qualityScores.slice(-6, -3);
  
  const recentAvg = recent.reduce((a, b) => a + b, 0) / recent.length;
  const previousAvg = previous.reduce((a, b) => a + b, 0) / previous.length;
  
  const change = ((recentAvg - previousAvg) / previousAvg) * 100;
  
  return change > 10 ? 'improving' : change < -10 ? 'declining' : 'stable';
}

function extractRecentLearnings(memory) {
  return memory.filter(entry => 
    entry.includes('NEW_INSIGHTS:') ||
    entry.includes('LEARNING_ITERATION:')
  ).slice(-5).map(entry => {
    const insightMatch = entry.match(/- ([^:]+):/);
    const contextMatch = entry.match(/Context:\s*(.+)/);
    return {
      insight: insightMatch ? insightMatch[1] : 'Process improvement',
      context: contextMatch ? contextMatch[1] : 'general'
    };
  });
}

function analyzeTechEffectiveness(memory) {
  const techMentions = {
    'React': 0,
    'TypeScript': 0,
    'Python': 0,
    'FastAPI': 0,
    'PostgreSQL': 0,
    'Docker': 0,
    'Context7': 0,
    'Sub-agent': 0
  };
  
  const techSuccess = { ...techMentions };
  
  memory.forEach(entry => {
    Object.keys(techMentions).forEach(tech => {
      if (entry.includes(tech)) {
        techMentions[tech]++;
        if (entry.includes('SUCCESS') || entry.includes('Quality: 9') || entry.includes('Quality: 10')) {
          techSuccess[tech]++;
        }
      }
    });
  });
  
  const effectiveness = {};
  Object.keys(techMentions).forEach(tech => {
    if (techMentions[tech] > 0) {
      effectiveness[tech] = Math.round((techSuccess[tech] / techMentions[tech]) * 10);
    }
  });
  
  return effectiveness;
}

function assessLearningSystemHealth(analysis) {
  const patternRecognition = Math.min(analysis.successPatterns.length + analysis.failurePatterns.length, 10);
  const adaptationCapability = analysis.processOptimizations.length > 0 ? 8 : 5;
  const memoryUtilization = analysis.successPatterns.length > 0 && analysis.failurePatterns.length > 0 ? 9 : 6;
  const overallHealth = Math.round((patternRecognition + adaptationCapability + memoryUtilization) / 3);
  
  return {
    patternRecognition,
    adaptationCapability,
    memoryUtilization,
    overallHealth
  };
}

function generateNextSteps(analysis) {
  const steps = [];
  
  if (analysis.successPatterns.length === 0) {
    steps.push('Complete more iterations to build success pattern database');
  }
  
  if (analysis.failurePatterns.length === 0) {
    steps.push('Document failure patterns to improve learning');
  }
  
  if (analysis.velocityTrends.direction === 'declining') {
    steps.push('Focus on simplifying tasks and reducing scope');
  }
  
  if (analysis.processOptimizations.length === 0) {
    steps.push('Identify process improvements in current workflow');
  }
  
  if (steps.length === 0) {
    steps.push('Continue iterating to maintain learning momentum');
    steps.push('Consider implementing advanced learning features');
  }
  
  return steps;
}

function isWithinLast24Hours(entry) {
  const dateMatch = entry.match(/Date:\s*(.+)/);
  if (!dateMatch) return false;
  
  const entryDate = new Date(dateMatch[1]);
  const now = new Date();
  const dayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);
  
  return entryDate >= dayAgo;
}

// Execute the dashboard
await generateLearningDashboard();
```

## Usage
```
/project:learning:dashboard
```

Shows comprehensive learning analytics including:
- Performance metrics and trends
- Success and failure patterns
- Velocity analysis
- Learning insights
- Technology effectiveness
- System health assessment
- Actionable recommendations

## Example Output
```
🧠 LEARNING DASHBOARD
====================

📊 PERFORMANCE METRICS:
- Total iterations analyzed: 47
- Success rate: 89%
- Average task duration: 23min
- Learning efficacy: 8/10

🎯 TOP SUCCESS PATTERNS:
- Enhanced existing systems: Context7 guidance, iterative approach
- Simple focused tasks: Single responsibility, clear scope
- Context7 integration: Up-to-date docs, version matching

⚠️ TOP FAILURE PATTERNS:
- Over-engineering: Complexity creep, unnecessary abstractions
- Missing dependencies: Incomplete setup, version conflicts
- Scope expansion: Task creep, unclear requirements

📈 VELOCITY ANALYSIS:
- Trend: improving (15%)
- Current velocity: 3 tasks/hour
- Quality trend: stable

💡 LEARNING INSIGHTS:
- Context7 pre-checks reduce implementation errors
- Sub-agent delegation effective for simple tasks
- Incremental enhancement outperforms full rewrites

🚀 RECOMMENDED OPTIMIZATIONS:
- Continue Context7 integration for all implementations
- Increase sub-agent delegation for routine tasks
- Focus on incremental improvements over new features

🔧 TECHNOLOGY EFFECTIVENESS:
- React: 9/10
- TypeScript: 8/10
- Context7: 10/10
- Sub-agent: 7/10

⚡ LEARNING SYSTEM HEALTH:
- Pattern recognition: 8/10
- Adaptation capability: 9/10
- Memory utilization: 9/10
- Overall health: 9/10

📋 NEXT STEPS:
- Continue iterating to maintain learning momentum
- Consider implementing advanced learning features
```

**Bornes d'écriture**
* Autorisé : output console analytics, documentation insights
* Interdit : modification mémoire source, configuration système

**Étapes**
1. [mem0] Récupérer toute mémoire pour analyse
2. [sequential-thinking] Traiter patterns et métriques
3. Générer analytics performance complètes
4. Extraire insights et recommandations
5. Présenter dashboard structuré

**Points de vigilance**
- Analytics complètes depuis mémoire
- Patterns succès ET échecs capturés
- Recommandations actionnables prioritisées
- Health système apprentissage monitored

**Tests/Validation**
- Vérification complétude analytics mémoire
- Validation pertinence insights extraits
- Test actionnabilité recommandations

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