package reporting

import (
	"fmt"
	"io"

	"github.com/elm-as/elmasguard/rules"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// RenderTerminalReport prints an intentional, high-density terminal report.
func RenderTerminalReport(writer io.Writer, violations []rules.Violation, filesScanned int) {
	fmt.Fprintf(writer, "\n%s%sELMASGUARD // Audit de Conformité de Charte ElmasCore%s\n", colorBold, colorCyan, colorReset)
	fmt.Fprintf(writer, "%s\"From Data to Systems\" — Rigueur, Clarté et Zéro Spaghetti%s\n\n", colorGray, colorReset)

	errorCount := 0
	warningCount := 0

	for _, v := range violations {
		if v.Severity == rules.SeverityError {
			errorCount++
			fmt.Fprintf(writer, "  %s[REJET]%s  %s:%d\n", colorRed, colorReset, v.FilePath, v.LineNumber)
		} else {
			warningCount++
			fmt.Fprintf(writer, "  %s[AVIS]%s   %s:%d\n", colorYellow, colorReset, v.FilePath, v.LineNumber)
		}
		fmt.Fprintf(writer, "          %s%s%s : %s\n", colorBold, v.RuleID, colorReset, v.Message)
		fmt.Fprintf(writer, "          %sConseil : %s%s\n\n", colorGray, v.Recommendation, colorReset)
	}

	fmt.Fprintf(writer, "------------------------------------------------------------------------\n")
	fmt.Fprintf(writer, "Fichiers audités : %d | ", filesScanned)
	if errorCount == 0 && warningCount == 0 {
		fmt.Fprintf(writer, "%s%sPARFAIT : 100%% Conforme à la charte ELMAS.md%s\n\n", colorBold, colorGreen, colorReset)
	} else if errorCount == 0 {
		fmt.Fprintf(writer, "%s%sVALIDE : 0 violation bloquante%s (%d avertissements)\n\n", colorBold, colorGreen, colorReset, warningCount)
	} else {
		fmt.Fprintf(writer, "%s%sÉCHEC : %d violation(s) bloquante(s)%s (%d avertissements)\n\n", colorBold, colorRed, errorCount, colorReset, warningCount)
	}
}
