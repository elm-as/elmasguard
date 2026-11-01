package rules

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	demeterChainPattern = regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z_][a-zA-Z0-9_]*\([^)]*\)\.[a-zA-Z_][a-zA-Z0-9_]*\([^)]*\)\.[a-zA-Z_][a-zA-Z0-9_]*`)
	secretLiteralPattern = regexp.MustCompile(`(?i)\b(?:api[_-]?key|secret[_-]?key|password|passwd|private[_-]?key|access[_-]?token)\s*[:=]\s*["'][A-Za-z0-9_\-./+=]{12,}["']`)
	booleanParamPattern  = regexp.MustCompile(`\b(?:bool|boolean)\b`)
)

// isStructuralChainException excludes fluent builder or standard library chains.
func isStructuralChainException(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
		return true
	}
	return strings.Contains(trimmed, "regexp.") || strings.Contains(trimmed, "strings.")
}

// isTestFile checks whether a file is an automated test suite.
func isTestFile(fileName string) bool {
	lower := strings.ToLower(fileName)
	return strings.HasSuffix(lower, "_test.go") ||
		strings.HasSuffix(lower, ".test.ts") ||
		strings.HasSuffix(lower, ".spec.ts") ||
		strings.HasSuffix(lower, ".test.js")
}

// buildDemeterViolation constructs a Law of Demeter violation.
func buildDemeterViolation(path string, lineNum int) Violation {
	return Violation{
		RuleID:         "ELMAS-COUP-001",
		Severity:       SeverityWarning,
		FilePath:       path,
		LineNumber:     lineNum,
		Message:        "Chaîne d'appels profonds détectée (violation de la Loi de Déméter)",
		Recommendation: "Un module ne discute qu'avec ses collaborateurs directs ; éviter les déréférencements profonds en cascade (ELMAS.md §3).",
	}
}

// buildSecretViolation constructs a hardcoded secret violation.
func buildSecretViolation(path string, lineNum int) Violation {
	return Violation{
		RuleID:         "ELMAS-SEC-001",
		Severity:       SeverityError,
		FilePath:       path,
		LineNumber:     lineNum,
		Message:        "Secret, mot de passe ou clé API potentiellement codé en dur",
		Recommendation: "Aucune donnée sensible en dur dans le code ; utiliser des variables d'environnement ou un secret manager (ELMAS.md §9).",
	}
}

// buildBooleanFlagViolation constructs a multiple boolean control flags violation.
func buildBooleanFlagViolation(path string, lineNum int, count int) Violation {
	return Violation{
		RuleID:         "ELMAS-FLAG-001",
		Severity:       SeverityWarning,
		FilePath:       path,
		LineNumber:     lineNum,
		Message:        fmt.Sprintf("Fonction recevant %d drapeaux booléens dans sa signature", count),
		Recommendation: "Si une fonction reçoit plusieurs booléens pour bifurquer, elle cache plusieurs responsabilités ; séparer les cas d'usage (ELMAS.md §3).",
	}
}

// inspectBooleanFlags checks if a function declaration contains 2 or more boolean parameters.
func inspectBooleanFlags(line string, path string, lineNum int) []Violation {
	matches := funcDeclPattern.FindAllStringSubmatch(line, -1)
	var violations []Violation

	for _, match := range matches {
		if len(match) <= 1 {
			continue
		}
		boolMatches := booleanParamPattern.FindAllStringIndex(match[1], -1)
		if len(boolMatches) >= 2 {
			violations = append(violations, buildBooleanFlagViolation(path, lineNum, len(boolMatches)))
		}
	}

	return violations
}

// CheckCouplingAndSecurityCompliance enforces Law of Demeter, anti-boolean flags, and secret scanning per ELMAS.md §3 & §9.
// Pure function: deterministic, zero side effects.
func CheckCouplingAndSecurityCompliance(ctx FileContext) []Violation {
	if isTestFile(ctx.FileName) {
		return nil
	}

	isGoSource := strings.HasSuffix(strings.ToLower(ctx.FileName), ".go")
	var violations []Violation

	for idx, line := range ctx.Lines {
		if isStructuralChainException(line) {
			continue
		}

		if demeterChainPattern.MatchString(line) {
			violations = append(violations, buildDemeterViolation(ctx.RelativePath, idx+1))
		}

		if secretLiteralPattern.MatchString(line) {
			violations = append(violations, buildSecretViolation(ctx.RelativePath, idx+1))
		}

		if !isGoSource {
			violations = append(violations, inspectBooleanFlags(line, ctx.RelativePath, idx+1)...)
		}
	}

	return violations
}
