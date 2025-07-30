# User Commands

✏️ **Vos commandes personnalisées - Modifiez ici !**

## 🎯 Comment ajouter une commande

1. **Organisez par catégorie** (optionnel) :
   ```
   commands/
   ├── mon-workflow/
   ├── templates/
   └── outils/
   ```

2. **Créez vos fichiers** :
   - `.md` pour les commandes claude
   - `.json` pour les templates
   - `.sh` pour les scripts

3. **Appliquez** :
   ```bash
   claude-wm config sync
   ```

## 📝 Exemple : Template personnalisé

**templates/mon-template.json** :
```json
{
  "nom": "",
  "description": "",
  "created": "{{DATE}}",
  "author": "{{USER}}"
}
```

## 📝 Exemple : Commande personnalisée

**mon-workflow/deploy.md** :
```markdown
# Commande de déploiement

Execute les étapes de déploiement de l'application.

## Prerequisites
- Tests passés
- Code validé

## Steps
1. Build de production
2. Upload vers serveur
3. Redémarrage services
```

## 🔄 Override des commandes système

Pour remplacer une commande système :
1. Copiez depuis `../system/commands/`
2. Modifiez votre copie ici
3. `claude-wm config sync`

La version user prendra priorité !

---

*Vos commandes personnalisées - Zone d'édition libre*