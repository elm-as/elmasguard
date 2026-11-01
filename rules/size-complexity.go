package rules

import "fmt"

const (
	HardLineLimit = 350
	WarnLineLimit = 300
)

// buildHardLimitViolation constructs a hard line limit violation.
func buildHardLimitViolation(ctx FileContext, lineCount int) []Violation {
	return []Violation{
		{
			RuleID:         "ELMAS-SIZE-001",
			Severity:       SeverityError,
			FilePath:       ctx.RelativePath,
			LineNumber:     lineCount,
			Message:        fmt.Sprintf("Fichier de %d lignes dépassant la limite stricte de %d lignes", lineCount, HardLineLimit),
			Recommendation: "Découper le fichier par responsabilité métier unique avant de l'étendre (ELMAS.md §2).",
		},
	}
}

// buildWarnLimitViolation constructs a warning limit violation.
func buildWarnLimitViolation(ctx FileContext, lineCount int) []Violation {
	return []Violation{
		{
			RuleID:         "ELMAS-SIZE-002",
			Severity:       SeverityWarning,
			FilePath:       ctx.RelativePath,
			LineNumber:     lineCount,
			Message:        fmt.Sprintf("Fichier de %d lignes approchant du plafond (%d/%d lignes)", lineCount, lineCount, HardLineLimit),
			Recommendation: "Ce fichier commence à faire plusieurs choses ; anticiper son découpage (ELMAS.md §2).",
		},
	}
}

// CheckFileSizeCompliance verifies that files respect line limits defined in ELMAS.md §2.
// Pure function: deterministic, max indentation depth 1.
func CheckFileSizeCompliance(ctx FileContext) []Violation {
	lineCount := len(ctx.Lines)
	if lineCount <= WarnLineLimit {
		return nil
	}

	if lineCount > HardLineLimit {
		return buildHardLimitViolation(ctx, lineCount)
	}

	return buildWarnLimitViolation(ctx, lineCount)
}
