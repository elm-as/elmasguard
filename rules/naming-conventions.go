package rules

import (
	"fmt"
	"path/filepath"
	"strings"
)

var forbiddenGenericNames = map[string]struct{}{
	"utils":   {},
	"util":    {},
	"helpers": {},
	"helper":  {},
	"misc":    {},
	"common":  {},
	"temp":    {},
	"tmp":     {},
	"final":   {},
	"v2":      {},
	"new":     {},
}

// isSegmentForbidden tests if a segment is a forbidden generic name.
func isSegmentForbidden(seg string) bool {
	lowerSeg := strings.ToLower(seg)
	dotIdx := strings.Index(lowerSeg, ".")
	nameWithoutExt := lowerSeg
	if dotIdx != -1 {
		nameWithoutExt = lowerSeg[:dotIdx]
	}

	_, forbidden := forbiddenGenericNames[nameWithoutExt]
	return forbidden
}

// buildNamingViolation creates a naming violation struct.
func buildNamingViolation(path string, seg string) Violation {
	return Violation{
		RuleID:         "ELMAS-NAME-001",
		Severity:       SeverityError,
		FilePath:       path,
		LineNumber:     1,
		Message:        fmt.Sprintf("Nom générique ou fourre-tout interdit détecté : '%s'", seg),
		Recommendation: "Donner un nom décrivant le domaine métier réel (ex: 'student-payment-status.go' au lieu de 'utils.go') (ELMAS.md §1 & §6).",
	}
}

// CheckNamingCompliance verifies files and folder names respect ELMAS.md §1 & §6.
// Pure function: early returns, max indentation depth 1.
func CheckNamingCompliance(ctx FileContext) []Violation {
	cleanPath := filepath.ToSlash(ctx.RelativePath)
	segments := strings.Split(cleanPath, "/")

	for _, seg := range segments {
		if !isSegmentForbidden(seg) {
			continue
		}
		return []Violation{buildNamingViolation(ctx.RelativePath, seg)}
	}

	return nil
}
