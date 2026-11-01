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

// calculateHealthScore computes a 0..100 architectural health score and letter grade.
// Pure function: deterministic, zero side effects.
func calculateHealthScore(errorCount int, warningCount int) (int, string) {
	score := 100 - (errorCount * 12) - (warningCount * 3)
	if score < 0 {
		score = 0
	}

	switch {
	case errorCount == 0 && warningCount == 0:
		return 100, "A+"
	case score >= 90 && errorCount == 0:
		return score, "A"
	case score >= 75:
		return score, "B"
	case score >= 55:
		return score, "C"
	default:
		return score, "F"
	}
}

// RenderTerminalReport prints an intentional, high-density terminal report.
func RenderTerminalReport(writer io.Writer, violations []rules.Violation, filesScanned int) {
	fmt.Fprintf(writer, "\n%s%sELMASGUARD v1.1 // Audit de Conformité de Charte ElmasCore%s\n", colorBold, colorCyan, colorReset)
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

	score, grade := calculateHealthScore(errorCount, warningCount)

	fmt.Fprintf(writer, "------------------------------------------------------------------------\n")
	fmt.Fprintf(writer, "Fichiers : %d | Score ElmasCore : %s%d/100 [%s]%s | ", filesScanned, colorBold, score, grade, colorReset)
	if errorCount == 0 && warningCount == 0 {
		fmt.Fprintf(writer, "%s%sPARFAIT : 100%% Conforme à ELMAS.md%s\n\n", colorBold, colorGreen, colorReset)
	} else if errorCount == 0 {
		fmt.Fprintf(writer, "%s%sVALIDE : 0 violation bloquante%s (%d avis)\n\n", colorBold, colorGreen, colorReset, warningCount)
	} else {
		fmt.Fprintf(writer, "%s%sÉCHEC : %d violation(s) bloquante(s)%s (%d avis)\n\n", colorBold, colorRed, errorCount, colorReset, warningCount)
	}
}
