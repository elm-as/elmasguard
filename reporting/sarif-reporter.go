package reporting

import (
	"encoding/json"
	"io"

	"github.com/elm-as/elmasguard/rules"
)

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifDriver struct {
	Name           string `json:"name"`
	InformationURI string `json:"informationUri"`
	Version        string `json:"version"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifDocument struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

// mapSeverityToSarifLevel converts internal Severity to SARIF 2.1.0 level.
func mapSeverityToSarifLevel(sev rules.Severity) string {
	if sev == rules.SeverityError {
		return "error"
	}
	return "warning"
}

// RenderSarifReport serializes audit violations into SARIF 2.1.0 standard format.
func RenderSarifReport(writer io.Writer, violations []rules.Violation) error {
	results := make([]sarifResult, 0, len(violations))

	for _, v := range violations {
		results = append(results, sarifResult{
			RuleID:  v.RuleID,
			Level:   mapSeverityToSarifLevel(v.Severity),
			Message: sarifMessage{Text: v.Message + " — " + v.Recommendation},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{URI: v.FilePath},
						Region:           sarifRegion{StartLine: v.LineNumber},
					},
				},
			},
		})
	}

	doc := sarifDocument{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           "ElmasGuard",
						InformationURI: "https://github.com/elm-as/elmasguard",
						Version:        "1.1.0",
					},
				},
				Results: results,
			},
		},
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}
