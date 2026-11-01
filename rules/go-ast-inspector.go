package rules

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// calculateAstStatementNesting recursively measures control-flow nesting depth inside a Go function.
// Pure function over an AST node.
func calculateAstStatementNesting(node ast.Node, currentDepth int) (int, token.Pos) {
	maxDepth := currentDepth
	maxPos := token.NoPos

	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil || n == node {
			return true
		}

		isBranching := false
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt:
			isBranching = true
		}

		if !isBranching {
			return true
		}

		nextDepth := currentDepth + 1
		subMax, subPos := calculateAstStatementNesting(n, nextDepth)
		if subMax > maxDepth {
			maxDepth = subMax
			if subPos != token.NoPos {
				maxPos = subPos
			} else {
				maxPos = n.Pos()
			}
		}
		return false
	})

	return maxDepth, maxPos
}

// countGoFieldListParams counts total parameters in an AST FieldList.
func countGoFieldListParams(fields *ast.FieldList) (int, int) {
	if fields == nil {
		return 0, 0
	}
	total := 0
	boolCount := 0
	for _, field := range fields.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		total += count
		if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
			boolCount += count
		}
	}
	return total, boolCount
}

// inspectGoFuncDecl checks a single Go AST FuncDecl for parameter SLA, boolean flags, and control-flow nesting.
func inspectGoFuncDecl(fn *ast.FuncDecl, fset *token.FileSet, relPath string) []Violation {
	var violations []Violation
	lineNum := fset.Position(fn.Pos()).Line

	paramCount, boolCount := countGoFieldListParams(fn.Type.Params)
	if paramCount > 3 {
		violations = append(violations, buildSLAViolation(relPath, lineNum, paramCount))
	}
	if boolCount >= 2 {
		violations = append(violations, buildBooleanFlagViolation(relPath, lineNum, boolCount))
	}

	if fn.Body == nil {
		return violations
	}

	depth, pos := calculateAstStatementNesting(fn.Body, 0)
	if depth > MaxAllowedNestingDepth {
		nestLine := lineNum
		if pos != token.NoPos {
			nestLine = fset.Position(pos).Line
		}
		violations = append(violations, Violation{
			RuleID:         "ELMAS-NEST-001",
			Severity:       SeverityError,
			FilePath:       relPath,
			LineNumber:     nestLine,
			Message:        fmt.Sprintf("Fonction '%s' avec imbrication cyclomatique de niveau %d (maximum autorisé : %d)", fn.Name.Name, depth, MaxAllowedNestingDepth),
			Recommendation: "Appliquer des clauses de garde (Guard Clauses) et sorties précoces (Early Return) en début de fonction (ELMAS.md §3).",
		})
	}

	return violations
}

// CheckGoAstCompliance performs exact native AST analysis on Go source files.
// Pure function: deterministic, zero side effects.
func CheckGoAstCompliance(ctx FileContext) ([]Violation, bool) {
	if !strings.HasSuffix(strings.ToLower(ctx.FileName), ".go") {
		return nil, false
	}

	src := strings.Join(ctx.Lines, "\n")
	fset := token.NewFileSet()
	fileAst, err := parser.ParseFile(fset, ctx.RelativePath, src, parser.ParseComments)
	if err != nil {
		return nil, false
	}

	var violations []Violation
	for _, decl := range fileAst.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		violations = append(violations, inspectGoFuncDecl(fn, fset, ctx.RelativePath)...)
	}

	return violations, true
}
