package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/elm-as/elmasguard/reporting"
	"github.com/elm-as/elmasguard/rules"
	"github.com/elm-as/elmasguard/scanner"
)

// Execute runs the elmasguard CLI workflow from args.
func Execute(args []string) int {
	flags := flag.NewFlagSet("elmasguard", flag.ContinueOnError)
	jsonMode := flags.Bool("json", false, "Sortie au format JSON pour intégration CI/CD")
	strictMode := flags.Bool("strict", false, "Traiter les avertissements comme des erreurs bloquantes")
	targetDir := flags.String("path", ".", "Répertoire racine à auditer")

	if err := flags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur arguments : %v\n", err)
		return 2
	}

	violations, count, err := scanner.ScanWorkspace(*targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur lors du scan : %v\n", err)
		return 2
	}

	if *jsonMode {
		_ = reporting.RenderJsonReport(os.Stdout, violations, count)
	} else {
		reporting.RenderTerminalReport(os.Stdout, violations, count)
	}

	errorCount := 0
	warningCount := 0
	for _, v := range violations {
		if v.Severity == rules.SeverityError {
			errorCount++
		} else {
			warningCount++
		}
	}

	if errorCount > 0 || (*strictMode && warningCount > 0) {
		return 1
	}

	return 0
}
