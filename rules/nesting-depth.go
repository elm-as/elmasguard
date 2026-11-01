package rules

import (
	"fmt"
	"strings"
)

const MaxAllowedNestingDepth = 3

// calculateLineIndentationLevel determines the nesting depth of a line of code.
func calculateLineIndentationLevel(line string) int {
	trimmed := strings.TrimRight(line, " \t\r\n")
	if len(trimmed) == 0 {
		return 0
	}

	spaces := 0
	for _, ch := range trimmed {
		if ch == '\t' {
			spaces += 4
			continue
		}
		if ch == ' ' {
			spaces++
			continue
		}
		break
	}

	return spaces / 4
}

// isCommentOrBlank checks if a line is a comment or blank.
func isCommentOrBlank(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) == 0 {
		return true
	}
	return strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*")
}

// buildNestingViolation builds a violation for excessive nesting.
func buildNestingViolation(path string, lineNum int, depth int) Violation {
	return Violation{
		RuleID:         "ELMAS-NEST-001",
		Severity:       SeverityError,
		FilePath:       path,
		LineNumber:     lineNum,
		Message:        fmt.Sprintf("Indentation excessive de niveau %d (maximum autorisé : %d)", depth, MaxAllowedNestingDepth),
		Recommendation: "Appliquer des clauses de garde (Guard Clauses) et sorties précoces (Early Return) pour aplatir la pyramide (ELMAS.md §3).",
	}
}

// CheckNestingDepthCompliance flags deep nesting pyramids violating ELMAS.md §3.
// Pure function: max indentation depth 1.
func CheckNestingDepthCompliance(ctx FileContext) []Violation {
	var violations []Violation

	for idx, line := range ctx.Lines {
		if isCommentOrBlank(line) {
			continue
		}

		depth := calculateLineIndentationLevel(line)
		if depth <= MaxAllowedNestingDepth {
			continue
		}

		violations = append(violations, buildNestingViolation(ctx.RelativePath, idx+1, depth))
		if len(violations) >= 3 {
			break
		}
	}

	return violations
}
