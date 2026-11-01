package rules_test

import (
	"testing"

	"github.com/elm-as/elmasguard/rules"
)

func TestCheckFileSizeCompliance(t *testing.T) {
	shortFile := rules.FileContext{
		RelativePath: "pkg/core.go",
		Lines:        make([]string, 100),
	}
	if v := rules.CheckFileSizeCompliance(shortFile); len(v) != 0 {
		t.Errorf("Attendu 0 violation pour 100 lignes, obtenu %d", len(v))
	}

	tooLongFile := rules.FileContext{
		RelativePath: "pkg/bloated.go",
		Lines:        make([]string, 360),
	}
	v := rules.CheckFileSizeCompliance(tooLongFile)
	if len(v) != 1 || v[0].Severity != rules.SeverityError {
		t.Errorf("Attendu 1 violation bloquante pour 360 lignes, obtenu %+v", v)
	}
}

func TestCheckNamingCompliance(t *testing.T) {
	forbiddenFile := rules.FileContext{
		RelativePath: "src/utils/helpers.go",
	}
	v := rules.CheckNamingCompliance(forbiddenFile)
	if len(v) == 0 {
		t.Errorf("Attendu une violation de nommage pour 'utils', obtenu 0")
	}

	cleanFile := rules.FileContext{
		RelativePath: "payments/invoice-generator.go",
	}
	if vClean := rules.CheckNamingCompliance(cleanFile); len(vClean) != 0 {
		t.Errorf("Attendu 0 violation pour nom de domaine propre, obtenu %d", len(vClean))
	}
}

func TestCheckNestingDepthCompliance(t *testing.T) {
	nestedFile := rules.FileContext{
		RelativePath: "domain/logic.go",
		Lines: []string{
			"func process() {",
			"    if a {",
			"        if b {",
			"            if c {",
			"                if d {",
			"                    doSomething()",
			"                }",
			"            }",
			"        }",
			"    }",
			"}",
		},
	}
	v := rules.CheckNestingDepthCompliance(nestedFile)
	if len(v) == 0 {
		t.Errorf("Attendu violation d'indentation excessive, obtenu 0")
	}
}

func TestCheckSignatureSLACompliance(t *testing.T) {
	badSignature := rules.FileContext{
		RelativePath: "service/order.go",
		Lines: []string{
			"func calculateDiscount(amount float64, user string, isPromo bool, taxRate float64) float64 {",
		},
	}
	v := rules.CheckSignatureSLACompliance(badSignature)
	if len(v) == 0 {
		t.Errorf("Attendu violation SLA (>3 paramètres), obtenu 0")
	}
}
