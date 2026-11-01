package reporting

import (
	"encoding/json"
	"io"

	"github.com/elm-as/elmasguard/rules"
)

// AuditSummary contains aggregated metrics for machine-readable output.
type AuditSummary struct {
	TotalFilesScanned int               `json:"totalFilesScanned"`
	TotalViolations   int               `json:"totalViolations"`
	ErrorCount        int               `json:"errorCount"`
	WarningCount      int               `json:"warningCount"`
	IsCompliant       bool              `json:"isCompliant"`
	Violations        []rules.Violation `json:"violations"`
}

// RenderJsonReport serializes audit results into standard formatted JSON.
func RenderJsonReport(writer io.Writer, violations []rules.Violation, filesScanned int) error {
	errCount := 0
	warnCount := 0

	for _, v := range violations {
		if v.Severity == rules.SeverityError {
			errCount++
		} else {
			warnCount++
		}
	}

	summary := AuditSummary{
		TotalFilesScanned: filesScanned,
		TotalViolations:   len(violations),
		ErrorCount:        errCount,
		WarningCount:      warnCount,
		IsCompliant:       errCount == 0,
		Violations:        violations,
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}
