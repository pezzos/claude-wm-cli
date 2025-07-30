# Claude WM Configuration Workspace

Bienvenue dans l'espace de configuration **Package Manager** de claude-wm-cli !

## 🏗️ Architecture

Cette nouvelle architecture remplace l'ancienne dualité `.claude/` + `.claude-wm/.claude/` par un système unifié inspiré des package managers modernes.

```
.claude-wm/
├── system/    # 📦 Templates système (read-only)
├── user/      # ✏️  Vos personnalisations
├── runtime/   # ⚡ Configuration effective (auto-générée)
└── state.json # 📊 État du projet
```

## 🎯 Comment ça marche

1. **System** contient les templates par défaut fournis par claude-wm-cli
2. **User** contient VOS modifications et personnalisations
3. **Runtime** est généré automatiquement en fusionnant system + user
4. Le code utilise toujours **runtime/** pour fonctionner

## 🔧 Commandes principales

```bash
# Initialiser la configuration (première fois)
claude-wm config init

# Régénérer la configuration runtime
claude-wm config sync

# Mettre à jour les templates système
claude-wm config upgrade

# Voir la configuration effective
claude-wm config show
```

## ✅ Avantages

- ✅ **Pas de synchronisation manuelle** - Tout est automatique
- ✅ **Personnalisations préservées** - Vos modifs restent dans user/
- ✅ **Mises à jour sans risque** - Les templates se mettent à jour sans écraser vos fichiers
- ✅ **Architecture claire** - Plus de confusion sur où modifier quoi
- ✅ **Commandes familières** - Style package manager moderne

## 🚀 Migration automatique

Si vous aviez l'ancienne structure `.claude-wm/.claude/`, elle a été automatiquement migrée :
- Vos fichiers sont maintenant dans les bons répertoires
- L'ancien dossier a été sauvegardé dans `~/.claude-backup-YYYYMMDD_HHMMSS/`
- Tout continue de fonctionner comme avant !

---

*Généré par claude-wm-cli Package Manager*