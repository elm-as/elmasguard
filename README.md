# ElmasGuard (`elmasguard` v1.1)

> **"From Data to Systems" — Gardien automatisé et analyseur AST de la charte ElmasCore (`ELMAS.md`).**

`elmasguard` est un outil CLI natif ultra-rapide (< 15ms), concurrent (worker pool multi-cœurs) et à zéro dépendance externe, conçu pour analyser automatiquement le code source et garantir le respect strict des principes d'architecture, de pureté et de sécurité de la charte `ELMAS.md`.

---

## 1. Règles appliquées par le moteur

| Règle | Code | Sévérité | Description | Référence `ELMAS.md` |
| :--- | :--- | :--- | :--- | :--- |
| **Limite stricte de lignes** | `ELMAS-SIZE-001` | `ERROR` | **Rejet immédiat** pour tout fichier dépassant **350 lignes**. | §2 — Taille et complexité |
| **Seuil d'alerte de taille** | `ELMAS-SIZE-002` | `WARNING` | Avertissement dès **300 lignes** pour anticiper le découpage par domaine. | §2 — Taille et complexité |
| **Nommage & Domaines** | `ELMAS-NAME-001` | `ERROR` | Interdiction des dossiers et fichiers fourre-tout (`utils/`, `helpers/`, `misc/`, `common/`, `temp`, `v2`, `final`). | §1 & §6 — Où va le code |
| **Anti-Spaghetti (AST & Indentation)** | `ELMAS-NEST-001` | `ERROR` | **Maximum 3 niveaux d'imbrication** (analyse AST native `go/ast` pour Go + analyse lexicale multi-langages). | §3 — Clauses de garde |
| **SLA (Signature de fonction)** | `ELMAS-SLA-001` | `ERROR` | **Maximum 3 paramètres par fonction**. Au-delà, regrouper dans une structure/interface typée. | §3 — Signatures et SLA |
| **Drapeaux booléens multiples** | `ELMAS-FLAG-001` | `WARNING` | Détection des fonctions recevant $\ge 2$ paramètres booléens de contrôle (`isDraft`, `shouldNotify`...). | §3 — Pas de Boolean flags |
| **Loi de Déméter** | `ELMAS-COUP-001` | `WARNING` | Détection des chaînes d'appels profonds en cascade (`a.getB().getC().doD()`). | §3 — Couplage lâche |
| **Secrets & Clés en dur** | `ELMAS-SEC-001` | `ERROR` | Détection de clés API, tokens ou mots de passe codés en dur dans le code source. | §9 — Sécurité et données |
| **Gestion d'erreurs muette** | `ELMAS-ERR-001` | `ERROR` | Rejet des blocs `catch (...) {}` vides qui avalent silencieusement les erreurs. | §8 — Qualité générale |
| **Code mort & TODO orphelins** | `ELMAS-DEAD-001` / `ELMAS-TODO-001` | `WARNING` | Détection de code commenté laissé en place et des `// TODO` sans explication contextuelle. | §7 — Commentaires |

---

## 2. Architecture du Projet (`ELMAS.md` §1)

```
elmasguard/
├── rules/                      # Moteur de règles pures et analyseur AST
│   ├── violation-model.go      # Structures immuables Violation & FileContext
│   ├── go-ast-inspector.go     # Analyseur syntaxique abstrait natif (go/ast)
│   ├── size-complexity.go      # Vérification des plafonds 300 / 350 lignes
│   ├── naming-conventions.go   # Bannissement des noms génériques fourre-tout
│   ├── nesting-depth.go        # Calcul de profondeur d'imbrication
│   ├── signature-sla.go        # Validation SLA (<= 3 paramètres)
│   ├── coupling-security.go    # Loi de Déméter, drapeaux booléens & secrets
│   ├── code-hygiene.go         # Code mort commenté, TODO orphelins, catch vides
│   └── charter-rules_test.go   # Suite de tests unitaires automatisés
├── scanner/                    # Parcours système de fichiers & concurrence
│   └── file-tree-walker.go     # Worker pool multi-goroutines + support .gitignore
├── reporting/                  # Formateurs de sortie (Terminal, JSON, SARIF)
│   ├── terminal-formatter.go   # Affichage haute densité + Score ElmasCore (A+..F)
│   ├── json-reporter.go        # Export JSON structuré
│   └── sarif-reporter.go       # Export standard SARIF 2.1.0 (GitHub Security)
├── cli/                        # Orchestration de la ligne de commande
│   └── command-runner.go       # Drapeaux CLI & installateur de hook Git
└── main.go                     # Point d'entrée binaire
```

---

## 3. Installation & Utilisation

### Compilation du binaire natif
```bash
go build -o elmasguard.exe main.go
```

### Commandes principales
```bash
# Audit standard avec Score ElmasCore (0..100 [A+..F])
./elmasguard.exe -path .

# Mode strict (échoue dès le moindre avertissement)
./elmasguard.exe -strict -path .

# Export JSON pour pipelines CI/CD
./elmasguard.exe -json -path .

# Export SARIF 2.1.0 pour GitHub Code Scanning & IDE
./elmasguard.exe -sarif -path .

# Installation automatique du hook Git pre-commit en une commande
./elmasguard.exe -install-hook -path .
```
