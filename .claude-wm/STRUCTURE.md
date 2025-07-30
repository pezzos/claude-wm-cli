# 📁 Structure de .claude-wm

## 🏗️ Vue d'ensemble

```
.claude-wm/
├── 📖 README.md                    # Guide principal
├── 📖 GUIDE-UTILISATEUR.md         # Guide rapide
├── 📖 STRUCTURE.md                 # Ce fichier
├── 📊 state.json                   # État du projet
│
├── 📦 system/                      # Templates système (READ-ONLY)
│   ├── 📖 README.md
│   ├── ⚙️  settings.json.template
│   ├── 📁 commands/
│   │   ├── 📖 README.md
│   │   └── 📁 templates/
│   └── 📁 hooks/
│       ├── 📖 README.md
│       ├── 📁 common/
│       ├── 📁 agile/  
│       └── 📁 config/
│
├── ✏️  user/                       # VOS MODIFICATIONS
│   ├── 📖 README.md
│   ├── ⚙️  settings.json            # Votre config
│   ├── 📁 commands/
│   │   ├── 📖 README.md
│   │   └── [vos commandes]
│   └── 📁 hooks/
│       ├── 📖 README.md
│       └── [vos hooks]
│
└── ⚡ runtime/                     # Config effective (AUTO-GÉNÉRÉ)
    ├── 📖 README.md
    ├── ⚙️  settings.json            # Config mergée
    ├── 📁 commands/                # system + user
    └── 📁 hooks/                   # system + user
```

## 🎯 Règles simples

| Dossier | Vous pouvez... | Ne pas... |
|---------|----------------|-----------|
| `system/` | ❌ Lire seulement | ❌ Modifier (écrasé lors des updates) |
| `user/` | ✅ **Modifier librement** | ✅ Zone d'édition sécurisée |
| `runtime/` | ❌ Lire seulement | ❌ Modifier (regénéré automatiquement) |

## 📚 Documentation disponible

Chaque dossier contient un `README.md` expliquant :
- ✅ Son rôle dans l'architecture
- ✅ Ce que vous pouvez y faire
- ✅ Des exemples concrets
- ✅ Les commandes utiles

## 🚀 Commandes de base

```bash
# Voir l'état général
claude-wm config show

# Appliquer vos modifications
claude-wm config sync  

# Mettre à jour les templates
claude-wm config upgrade
```

---

*Architecture Package Manager - Documentation complète*