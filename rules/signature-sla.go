package rules

import (
	"fmt"
	"regexp"
	"strings"
)

var funcDeclPattern = regexp.MustCompile(`(?:func|function|def|\b[a-zA-Z0-9_]+\s+[a-zA-Z0-9_]+)\s*\(([^)]*)\)`)

// countParameters parses a raw parameter substring and counts comma-separated parameters.
func countParameters(rawParams string) int {
	trimmed := strings.TrimSpace(rawParams)
	if len(trimmed) == 0 {
		return 0
	}

	parts := strings.Split(trimmed, ",")
	return len(parts)
}

// buildSLAViolation creates a SLA violation struct.
func buildSLAViolation(path string, lineNum int, paramCount int) Violation {
	return Violation{
		RuleID:         "ELMAS-SLA-001",
		Severity:       SeverityError,
		FilePath:       path,
		LineNumber:     lineNum,
		Message:        fmt.Sprintf("Signature de fonction avec %d paramètres (maximum autorisé : 3)", paramCount),
		Recommendation: "Regrouper les arguments au-delà de 3 dans un objet ou une interface de configuration typée (ELMAS.md §3).",
	}
}

// inspectMatches evaluates regex matches on a single line.
func inspectMatches(matches [][]string, path string, lineNum int) []Violation {
	var violations []Violation
	for _, match := range matches {
		if len(match) <= 1 {
			continue
		}
		paramCount := countParameters(match[1])
		if paramCount <= 3 {
			continue
		}
		violations = append(violations, buildSLAViolation(path, lineNum, paramCount))
	}
	return violations
}

// CheckSignatureSLACompliance validates that functions do not exceed 3 parameters per ELMAS.md §3.
// Pure function: guard clauses, max indentation depth 1.
func CheckSignatureSLACompliance(ctx FileContext) []Violation {
	var violations []Violation

	for idx, line := range ctx.Lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		matches := funcDeclPattern.FindAllStringSubmatch(line, -1)
		violations = append(violations, inspectMatches(matches, ctx.RelativePath, idx+1)...)
	}

	return violations
}
