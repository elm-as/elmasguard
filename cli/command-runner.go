package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/elm-as/elmasguard/reporting"
	"github.com/elm-as/elmasguard/rules"
	"github.com/elm-as/elmasguard/scanner"
)

const preCommitHookScript = `#!/bin/sh
# ElmasGuard Pre-Commit Hook — "From Data to Systems"
if [ -x "./elmasguard.exe" ]; then
    ./elmasguard.exe -strict
elif [ -x "./elmasguard" ]; then
    ./elmasguard -strict
else
    elmasguard -strict
fi

if [ $? -ne 0 ]; then
    echo "Commit rejeté : violations de la charte ELMAS.md détectées."
    exit 1
fi
`

// installGitPreCommitHook writes the pre-commit hook into .git/hooks/pre-commit.
func installGitPreCommitHook(rootDir string) int {
	hooksDir := filepath.Join(rootDir, ".git", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Erreur : aucun dépôt Git trouvé dans '%s'\n", rootDir)
		return 2
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(preCommitHookScript), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur écriture hook : %v\n", err)
		return 2
	}

	fmt.Printf("Hook Git pre-commit ElmasGuard installé avec succès dans %s\n", hookPath)
	return 0
}

// evaluateExitCode determines process exit code from violations and strict flag.
func evaluateExitCode(violations []rules.Violation, strict bool) int {
	errorCount := 0
	warningCount := 0
	for _, v := range violations {
		if v.Severity == rules.SeverityError {
			errorCount++
		} else {
			warningCount++
		}
	}

	if errorCount > 0 || (strict && warningCount > 0) {
		return 1
	}
	return 0
}

// normalizeCliArgs defaults dangling "-path" or "--path" flags to "." and extracts positional path.
func normalizeCliArgs(args []string) []string {
	normalized := make([]string, 0, len(args)+1)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		isPathFlag := arg == "-path" || arg == "--path"
		if isPathFlag && (i+1 >= len(args) || (len(args[i+1]) > 0 && args[i+1][0] == '-')) {
			normalized = append(normalized, arg, ".")
			continue
		}
		normalized = append(normalized, arg)
	}
	return normalized
}

// Execute runs the elmasguard CLI workflow from args.
func Execute(args []string) int {
	flags := flag.NewFlagSet("elmasguard", flag.ContinueOnError)
	jsonMode := flags.Bool("json", false, "Sortie au format JSON pour intégration CI/CD")
	sarifMode := flags.Bool("sarif", false, "Sortie au format SARIF 2.1.0 pour GitHub Code Scanning")
	strictMode := flags.Bool("strict", false, "Traiter les avertissements comme des erreurs bloquantes")
	installHook := flags.Bool("install-hook", false, "Installer automatiquement le hook Git pre-commit")
	targetDir := flags.String("path", ".", "Répertoire racine à auditer")

	if err := flags.Parse(normalizeCliArgs(args)); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur arguments : %v\n", err)
		return 2
	}

	resolvedDir := *targetDir
	if resolvedDir == "." && flags.NArg() > 0 {
		resolvedDir = flags.Arg(0)
	}

	if *installHook {
		return installGitPreCommitHook(resolvedDir)
	}

	violations, count, err := scanner.ScanWorkspace(resolvedDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur lors du scan : %v\n", err)
		return 2
	}

	switch {
	case *sarifMode:
		_ = reporting.RenderSarifReport(os.Stdout, violations)
	case *jsonMode:
		_ = reporting.RenderJsonReport(os.Stdout, violations, count)
	default:
		reporting.RenderTerminalReport(os.Stdout, violations, count)
	}

	return evaluateExitCode(violations, *strictMode)
}
