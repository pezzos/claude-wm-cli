# Runtime Configuration (Auto-Generated)

⚡ **Configuration effective utilisée par claude-wm-cli**

## ⚠️ Important : NE PAS MODIFIER DIRECTEMENT

Ce dossier est **généré automatiquement** en fusionnant :
- `../system/` (templates par défaut)
- `../user/` (vos personnalisations)

## 📁 Contenu

- `commands/` - Commandes effectives (system + user)
- `hooks/` - Hooks effectifs (system + user)
- `settings.json` - Configuration effective (merge system + user)

## 🔄 Régénération

Ce dossier est recréé automatiquement :
- Au lancement de claude-wm-cli (si nécessaire)
- Avec `claude-wm config sync`
- Après `claude-wm config upgrade`

## 🎯 Logique de merge

1. **Fichiers** : user/ override system/ (remplacement complet)
2. **settings.json** : merge intelligent des objets JSON
3. **Dossiers** : combinaison des deux sources

## 🔍 Pour déboguer

Si quelque chose ne fonctionne pas comme attendu :

```bash
# Voir la config effective
claude-wm config show

# Voir un fichier spécifique
claude-wm config show settings.json

# Régénérer complètement
claude-wm config sync
```

## 🛠️ Pour modifier

**❌ NE PAS** éditer directement ici car vos changements seront perdus.

**✅ FAIRE** : Modifier dans `../user/` puis :
```bash
claude-wm config sync
```

## 📊 Exemple de merge

**system/settings.json.template** :
```json
{
  "version": "1.0.0",
  "hooks": {"PreToolUse": []},
  "model": "sonnet"
}
```

**user/settings.json** :
```json
{
  "model": "haiku",
  "hooks": {"PreToolUse": [{"matcher": "Bash", "hooks": ["mon-hook.sh"]}]}
}
```

**runtime/settings.json** (résultat) :
```json
{
  "version": "1.0.0",
  "model": "haiku",
  "hooks": {"PreToolUse": [{"matcher": "Bash", "hooks": ["mon-hook.sh"]}]}
}
```

---

*Configuration runtime - Générée automatiquement*