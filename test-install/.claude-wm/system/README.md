# System Templates (Read-Only)

📦 **Ce dossier contient les templates par défaut de claude-wm-cli**

## ⚠️ Important : NE PAS MODIFIER

Ce dossier est géré automatiquement par claude-wm-cli. Vos modifications seraient écrasées lors des mises à jour.

## 📁 Contenu

- `commands/` - Commandes claude par défaut
- `hooks/` - Hooks système par défaut  
- `templates/` - Templates de fichiers (JSON, MD, etc.)
- `settings.json.template` - Configuration par défaut

## 🔄 Mise à jour

Les templates système sont mis à jour avec :
```bash
claude-wm config upgrade
```

## ✏️ Pour personnaliser

Si vous voulez modifier quelque chose :
1. **NE PAS** modifier ici
2. Copiez le fichier dans `../user/` avec la même structure
3. Modifiez votre copie dans `../user/`
4. Lancez `claude-wm config sync` pour appliquer

## 🎯 Exemple

Pour personnaliser `settings.json.template` :
```bash
# ❌ NE PAS FAIRE : modifier system/settings.json.template
# ✅ FAIRE : 
cp system/settings.json.template ../user/settings.json
# Éditez ../user/settings.json
claude-wm config sync
```

---

*Templates système - Version managée automatiquement*