# 🚀 Guide Rapide - Claude WM Package Manager

## ✅ Démarrage en 3 étapes

### 1. Initialiser (première fois seulement)
```bash
claude-wm config init
```

### 2. Voir votre configuration
```bash
claude-wm config show
```

### 3. Personnaliser si besoin
```bash
# Éditer votre configuration
nano .claude-wm/user/settings.json

# Ajouter un hook personnalisé
echo '#!/bin/bash\necho "Mon hook !"' > .claude-wm/user/hooks/test.sh
chmod +x .claude-wm/user/hooks/test.sh

# Appliquer vos changements
claude-wm config sync
```

## 📁 Où mettre quoi ?

| Dossier | Usage | Modifiable ? |
|---------|-------|--------------|
| `system/` | Templates par défaut | ❌ Non (écrasé lors des mises à jour) |
| `user/` | **VOS modifications** | ✅ **OUI - Modifiez ici !** |
| `runtime/` | Configuration effective | ❌ Non (généré automatiquement) |

## 🔧 Commandes utiles

```bash
# Voir l'état
claude-wm config show

# Régénérer la config
claude-wm config sync

# Mettre à jour les templates
claude-wm config upgrade

# Voir un fichier spécifique
claude-wm config show settings.json
```

## ❓ Questions fréquentes

**Q: Où modifier ma configuration ?**  
R: Dans `user/settings.json` - jamais dans `system/` !

**Q: Comment ajouter un hook ?**  
R: Créez le script dans `user/hooks/`, référencez dans `user/settings.json`, puis `claude-wm config sync`

**Q: J'ai perdu ma config !**  
R: Pas de panique ! Votre config user est préservée. Relancez `claude-wm config sync`

**Q: Comment revenir en arrière ?**  
R: Supprimez vos modifs dans `user/` et relancez `claude-wm config sync`

## 🆘 Besoin d'aide ?

1. Lisez les `README.md` dans chaque dossier
2. Utilisez `claude-wm config show` pour déboguer
3. Regardez les exemples dans les README

---

*Guide utilisateur - Architecture Package Manager*