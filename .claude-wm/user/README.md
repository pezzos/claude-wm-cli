# User Customizations

✏️ **C'est ICI que vous devez faire VOS modifications !**

## 🎯 Principe

Ce dossier contient VOS personnalisations qui s'appliquent par-dessus les templates système.

## 📁 Structure (optionnelle)

Vous pouvez créer la même structure que dans `../system/` :
```
user/
├── commands/       # Vos commandes personnalisées
├── hooks/          # Vos hooks personnalisés
├── templates/      # Vos templates personnalisés
└── settings.json   # Votre configuration personnalisée
```

## ⚡ Comment ça marche

1. Les fichiers ici **overrident** les templates système
2. Structure identique = remplacement complet
3. settings.json = merge intelligent (vos clés + clés système)
4. Tout est automatiquement mergé dans `../runtime/`

## 🔧 Workflow recommandé

### Pour modifier settings.json :
```bash
# Copiez le template système
cp ../system/settings.json.template ./settings.json

# Éditez votre copie
nano settings.json

# Appliquez les changements
claude-wm config sync
```

### Pour ajouter un hook personnalisé :
```bash
# Créez le dossier si nécessaire
mkdir -p hooks/

# Ajoutez votre hook
echo '#!/bin/bash\necho "Mon hook perso"' > hooks/mon-hook.sh
chmod +x hooks/mon-hook.sh

# Appliquez
claude-wm config sync
```

## ✅ Avantages

- ✅ **Vos modifs sont préservées** lors des mises à jour système
- ✅ **Versioning possible** - Vous pouvez versionner ce dossier
- ✅ **Merge intelligent** - Les configs se combinent automatiquement
- ✅ **Override granulaire** - Remplacez seulement ce que vous voulez

## 🔄 Après modification

Toujours lancer après vos changements :
```bash
claude-wm config sync
```

---

*Vos personnalisations - Modifiez ici en toute sécurité !*