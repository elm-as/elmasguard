package rules

// Severity defines the impact level of a rule violation.
type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
)

// Violation represents an individual charter rule infringement.
type Violation struct {
	RuleID      string   `json:"ruleId"`
	Severity    Severity `json:"severity"`
	FilePath    string   `json:"filePath"`
	LineNumber  int      `json:"lineNumber"`
	Message     string   `json:"message"`
	Recommendation string `json:"recommendation"`
}

// FileContext encapsulates file properties and content for inspection.
type FileContext struct {
	RelativePath string
	FileName     string
	Lines        []string
}
