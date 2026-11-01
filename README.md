# ElmasGuard (`elmasguard`)

> **"From Data to Systems" — Gardien automatisé de la charte de développement ElmasCore.**

`elmasguard` est un outil CLI natif ultra-rapide (< 15ms), à zéro dépendance externe, conçu pour analyser automatiquement le code source et garantir le respect strict des principes d'architecture et de qualité de la charte `ELMAS.md`.

---

## 1. Règles appliquées par le moteur

| Règle | Code | Description | Référence `ELMAS.md` |
| :--- | :--- | :--- | :--- |
| **Limite de lignes** | `ELMAS-SIZE-001` | **Rejet immédiat** pour tout fichier dépassant **350 lignes**. Avertissement à 300 lignes. | §2 — Taille et complexité |
| **Nommage & Domaines** | `ELMAS-NAME-001` | Interdiction des dossiers et fichiers fourre-tout (`utils/`, `helpers/`, `misc/`, `common/`, `temp`, `v2`, `final`). | §1 & §6 — Où va le code / Nommage |
| **Anti-Spaghetti** | `ELMAS-NEST-001` | **Maximum 3 niveaux d'indentation**. Rejet des "pyramides du destin" ; obligation d'utiliser des Guard Clauses et Early Returns. | §3 — Prévention du code spaghetti |
| **SLA (Abstraction)** | `ELMAS-SLA-001` | **Maximum 3 paramètres par fonction**. Au-delà, obligation d'extraire un objet/interface typée. | §3 — Signatures et abstraction |
| **Code mort & Hygiène** | `ELMAS-DEAD-001` / `ELMAS-TODO-001` | Détection de code mort commenté (`// let ...`, `# import ...`), `catch` vides et `TODO` orphelins sans contexte. | §7 & §8 — Commentaires et qualité |

---

## 2. Installation & Compilation

```bash
# Compilation du binaire natif autonome
go build -o elmasguard.exe main.go
```

---

## 3. Utilisation

### Audit du projet courant
```bash
./elmasguard.exe
```

### Audit d'un répertoire spécifique
```bash
./elmasguard.exe -path C:/Users/.../MonProjet
```

### Mode strict (les avertissements échouent la commande)
```bash
./elmasguard.exe -strict
```

### Mode JSON (intégration CI/CD & GitHub Actions)
```bash
./elmasguard.exe -json
```

---

## 4. Intégration en Hook Git Pre-commit

Ajoutez simplement dans `.git/hooks/pre-commit` :

```bash
#!/bin/sh
elmasguard -strict
if [ $? -ne 0 ]; then
    echo "Commit rejeté : violations de la charte ELMAS.md détectées."
    exit 1
fi
```
