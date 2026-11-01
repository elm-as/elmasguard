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
		RelativePath: "domain/logic.ts",
		FileName:     "logic.ts",
		Lines: []string{
			"function process() {",
			"    if (a) {",
			"        if (b) {",
			"            if (c) {",
			"                if (d) {",
			"                    doSomething();",
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
		RelativePath: "service/order.ts",
		FileName:     "order.ts",
		Lines: []string{
			"function calculateDiscount(amount, user, isPromo, taxRate) {",
		},
	}
	v := rules.CheckSignatureSLACompliance(badSignature)
	if len(v) == 0 {
		t.Errorf("Attendu violation SLA (>3 paramètres), obtenu 0")
	}
}

func TestCheckGoAstCompliance(t *testing.T) {
	goFileWithDeepNestingAndManyParams := rules.FileContext{
		RelativePath: "orders/processor.go",
		FileName:     "processor.go",
		Lines: []string{
			"package orders",
			"func ProcessOrder(id string, isDraft bool, shouldNotify bool, isAdmin bool) {",
			"	if id != \"\" {",
			"		for i := 0; i < 10; i++ {",
			"			if isDraft {",
			"				if shouldNotify {",
			"					println(id)",
			"				}",
			"			}",
			"		}",
			"	}",
			"}",
		},
	}

	violations, handled := rules.CheckGoAstCompliance(goFileWithDeepNestingAndManyParams)
	if !handled {
		t.Fatalf("Attendu prise en charge par l'analyseur AST Go")
	}
	if len(violations) < 3 {
		t.Errorf("Attendu au moins 3 violations (SLA > 3 params, >= 2 bool flags, nesting > 3), obtenu %d : %+v", len(violations), violations)
	}
}

func TestCheckCouplingAndSecurityCompliance(t *testing.T) {
	riskyFile := rules.FileContext{
		RelativePath: "auth/client.ts",
		FileName:     "client.ts",
		Lines: []string{
			`const apiKey = "sk_live_9876543210abcdef123456";`,
			`const city = order.getCustomer().getAddress().getCity().normalize();`,
		},
	}

	violations := rules.CheckCouplingAndSecurityCompliance(riskyFile)
	if len(violations) != 2 {
		t.Errorf("Attendu 2 violations (ELMAS-SEC-001 + ELMAS-COUP-001), obtenu %d : %+v", len(violations), violations)
	}
}
