# User Hooks

✏️ **Vos hooks personnalisés - Modifiez ici !**

## 🎯 Comment ajouter un hook

1. **Créez votre script** dans ce dossier :
   ```bash
   # Exemple : hook de notification personnalisé
   echo '#!/bin/bash
   echo "🎉 Hook personnalisé exécuté !"
   osascript -e "display notification \"Action terminée\" with title \"Claude WM\""' > mon-hook.sh
   
   chmod +x mon-hook.sh
   ```

2. **Référencez dans ../settings.json** :
   ```json
   {
     "hooks": {
       "PostToolUse": [
         {
           "matcher": "Write",
           "hooks": ["/chemin/absolu/vers/.claude-wm/user/hooks/mon-hook.sh"]
         }
       ]
     }
   }
   ```

3. **Appliquez vos changements** :
   ```bash
   claude-wm config sync
   ```

## 💡 Types de matchers courants

- `"Bash"` - Déclenché avant les commandes bash
- `"Write"` - Déclenché après les écritures de fichiers
- `"Edit"` - Déclenché après les modifications de fichiers
- `"MultiEdit"` - Déclenché après les modifications multiples
- `""` - Déclenché pour toute action (matcher vide)

## 📝 Exemples utiles

### Hook de sauvegarde automatique
```bash
#!/bin/bash
# backup-hook.sh
git add . && git commit -m "Auto-backup: $(date)"
```

### Hook de notification macOS
```bash
#!/bin/bash
# notify-hook.sh
osascript -e "display notification \"$1\" with title \"Claude WM\" sound name \"Glass\""
```

### Hook de validation
```bash
#!/bin/bash
# validate-hook.sh
if [[ "$1" == *.json ]]; then
    python -m json.tool "$1" > /dev/null || echo "⚠️ JSON invalide: $1"
fi
```

## 🔄 Après modification

Toujours lancer après vos changements :
```bash
claude-wm config sync
```

---

*Vos hooks personnalisés - Zone d'édition libre*