package rules

import (
	"regexp"
)

var (
	commentedCodePattern = regexp.MustCompile(`^\s*(?://|#)\s*(?:const|let|var|import|return|func|function|class|if|for|while)\b`)
	orphanTodoPattern    = regexp.MustCompile(`(?i)^\s*(?://|#)\s*TODO\s*[:\-]?\s*$`)
	emptyCatchPattern    = regexp.MustCompile(`catch\s*\([^)]*\)\s*\{\s*\}`)
)

// inspectLineHygiene checks a single line for hygiene violations.
func inspectLineHygiene(line string, lineNum int, path string) []Violation {
	var violations []Violation

	if commentedCodePattern.MatchString(line) {
		violations = append(violations, Violation{
			RuleID:         "ELMAS-DEAD-001",
			Severity:       SeverityWarning,
			FilePath:       path,
			LineNumber:     lineNum,
			Message:        "Code mort commenté détecté",
			Recommendation: "Le contrôle de version garde l'historique ; supprimer le code mort au lieu de le commenter (ELMAS.md §7 & §8).",
		})
	}

	if orphanTodoPattern.MatchString(line) {
		violations = append(violations, Violation{
			RuleID:         "ELMAS-TODO-001",
			Severity:       SeverityWarning,
			FilePath:       path,
			LineNumber:     lineNum,
			Message:        "TODO orphelin sans explication contextuelle",
			Recommendation: "// TODO autorisé uniquement avec le contexte de ce qui manque (ELMAS.md §7).",
		})
	}

	if emptyCatchPattern.MatchString(line) {
		violations = append(violations, Violation{
			RuleID:         "ELMAS-ERR-001",
			Severity:       SeverityError,
			FilePath:       path,
			LineNumber:     lineNum,
			Message:        "Bloc 'catch' vide avalant silencieusement une erreur",
			Recommendation: "Toute erreur doit être gérée ou propagée explicitement (ELMAS.md §8).",
		})
	}

	return violations
}

// CheckCodeHygieneCompliance flags commented-out dead code and orphan TODOs per ELMAS.md §7 & §8.
// Pure function: early returns, max indentation depth 1.
func CheckCodeHygieneCompliance(ctx FileContext) []Violation {
	var violations []Violation

	for idx, line := range ctx.Lines {
		violations = append(violations, inspectLineHygiene(line, idx+1, ctx.RelativePath)...)
	}

	return violations
}
