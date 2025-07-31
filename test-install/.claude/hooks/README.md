# System Hooks (Package Manager)

📦 **Hooks système par défaut de claude-wm-cli**

## ⚠️ Important : Read-Only

Ce dossier contient les hooks fournis par claude-wm-cli. **Ne pas modifier directement**.

## 🎯 Pour personnaliser

1. Copiez le hook dans `../../user/hooks/`
2. Modifiez votre copie
3. Référencez dans `../../user/settings.json`
4. Lancez `claude-wm config sync`

## 📁 Structure

- `common/` - Hooks partagés (backup, git-status, etc.)
- `agile/` - Hooks spécifiques au workflow agile
- `config/` - Configuration des triggers et groupes parallèles
- `logs/` - Statistiques de performance et fiabilité
- `patterns/` - Patterns de sécurité et validation
- `*.sh`, `*.go`, `*.py` - Scripts individuels

## 🔧 Utilisation dans settings.json

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": ["chemin/vers/votre-hook.sh"]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write",
        "hooks": ["chemin/vers/post-hook.py"]
      }
    ]
  }
}
```

## 📝 Exemple de personnalisation

```bash
# 1. Copiez un hook système vers user/
cp smart-notify.sh ../../user/hooks/mon-notify.sh

# 2. Modifiez votre copie
nano ../../user/hooks/mon-notify.sh

# 3. Référencez dans user/settings.json
{
  "hooks": {
    "Notification": [
      {
        "matcher": "",
        "hooks": ["/chemin/absolu/vers/mon-notify.sh"]
      }
    ]
  }
}

# 4. Appliquez
claude-wm config sync
```

## 🔍 Hooks principaux

- `parallel-hook-runner.sh` - Orchestrateur principal
- `smart-notify.sh` - Notifications système
- `git-validator.go` - Validation Git
- `security-validator.go` - Validation sécurité
- `duplicate-detector.go` - Détection doublons

## 📊 Debugging

Runtime hooks sont dans `../../runtime/hooks/` - consultez les logs là-bas.

---

*Hooks système - Version managée automatiquement*