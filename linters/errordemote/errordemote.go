package errordemote

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `detect errors that are demoted to log statements instead of being returned

This analyzer identifies the "fail-fast vs resilient" pattern where:
1. A function returns (value, error)
2. The error is caught but NOT returned to the caller
3. The error is only logged (Info/Debug/Warn level)
4. Code continues with a default value

This pattern can hide critical failures. The linter requires explicit
documentation when demoting errors to logs.

Example flagged code:

	if value, err := getConfig(ctx, cli); err == nil {
		config.Value = value
	} else {
		log.Info("couldn't get config", "error", err)  // Error demoted to log
	}

To suppress, add a comment explaining why the error can be safely ignored:

	//nolint:errordemote // ConfigMap may not exist on non-OCP clusters
	if value, err := getConfig(ctx, cli); err == nil {
		config.Value = value
	} else {
		log.Info("couldn't get config", "error", err)
	}

Or document with an explicit comment:

	// RESILIENCE: config is optional; safe to continue with zero value
	if value, err := getConfig(ctx, cli); err == nil {
		config.Value = value
	} else {
		log.Info("couldn't get config", "error", err)
	}
`

var Analyzer = &analysis.Analyzer{
	Name:     "errordemote",
	Doc:      Doc,
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.IfStmt)(nil),
	}

	inspector.Preorder(nodeFilter, func(n ast.Node) {
		ifStmt := n.(*ast.IfStmt)

		// Check if this is the error demotion pattern:
		// if val, err := fn(); err == nil { ... } else { log... }
		if isErrorDemotionPattern(ifStmt, pass) {
			// Check for nolint comment
			if hasNolintComment(pass, ifStmt.Pos()) {
				return
			}

			// Check for explicit resilience documentation
			if hasResilienceDoc(pass, ifStmt.Pos()) {
				return
			}

			pass.Reportf(ifStmt.Pos(),
				"error demoted to log statement instead of being returned; add //nolint:errordemote with justification or return the error")
		}
	})

	return nil, nil
}

// isErrorDemotionPattern checks if this is the error demotion pattern
func isErrorDemotionPattern(ifStmt *ast.IfStmt, pass *analysis.Pass) bool {
	// Must have an assignment in the init section
	// Pattern: if val, err := fn(); err == nil { ... } else { ... }
	if ifStmt.Init == nil {
		return false
	}

	assignStmt, ok := ifStmt.Init.(*ast.AssignStmt)
	if !ok || assignStmt.Tok != token.DEFINE {
		return false
	}

	// Must assign to at least 2 values (value, error)
	if len(assignStmt.Lhs) < 2 {
		return false
	}

	// Last variable should be named "err" or "_err"
	lastVar, ok := assignStmt.Lhs[len(assignStmt.Lhs)-1].(*ast.Ident)
	if !ok {
		return false
	}
	if !strings.Contains(lastVar.Name, "err") && lastVar.Name != "_" {
		return false
	}

	// Condition should be "err == nil" or "err != nil"
	if !isErrCondition(ifStmt.Cond) {
		return false
	}

	// Must have an else branch
	if ifStmt.Else == nil {
		return false
	}

	// The else branch should contain logging but NOT return an error
	hasLog := containsLogCall(ifStmt.Else)
	returnsError := containsErrorReturn(ifStmt.Else)

	// Pattern: logs error but doesn't return it
	return hasLog && !returnsError
}

// isErrCondition checks if the condition is testing an error variable
func isErrCondition(cond ast.Expr) bool {
	switch expr := cond.(type) {
	case *ast.BinaryExpr:
		// err == nil or err != nil
		if (expr.Op == token.EQL || expr.Op == token.NEQ) && 
		   (isNilIdent(expr.Y) || isNilIdent(expr.X)) {
			// Check if the other side is an error variable
			if ident, ok := expr.X.(*ast.Ident); ok && strings.Contains(ident.Name, "err") {
				return true
			}
			if ident, ok := expr.Y.(*ast.Ident); ok && strings.Contains(ident.Name, "err") {
				return true
			}
		}
	}
	return false
}

// isNilIdent checks if an expression is the "nil" identifier
func isNilIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

// containsLogCall checks if a statement contains a log call
func containsLogCall(stmt ast.Stmt) bool {
	hasLog := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				// Common logging methods
				logMethods := map[string]bool{
					"Info":    true,
					"Debug":   true,
					"Warn":    true,
					"Warning": true,
					"Trace":   true,
					"V":       true, // klog verbosity
				}
				if logMethods[sel.Sel.Name] {
					hasLog = true
					return false
				}
			}
		}
		return true
	})
	return hasLog
}

// containsErrorReturn checks if a statement returns an error
func containsErrorReturn(stmt ast.Stmt) bool {
	hasReturn := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if ret, ok := n.(*ast.ReturnStmt); ok {
			// Check if any return value contains "err"
			for _, result := range ret.Results {
				if ident, ok := result.(*ast.Ident); ok && strings.Contains(ident.Name, "err") {
					hasReturn = true
					return false
				}
			}
		}
		return true
	})
	return hasReturn
}

// hasNolintComment checks if there's a //nolint:errordemote comment
func hasNolintComment(pass *analysis.Pass, pos token.Pos) bool {
	file := pass.Fset.File(pos)
	if file == nil {
		return false
	}

	line := file.Line(pos)
	
	// Check current line and previous line
	for _, commentGroup := range pass.Files[0].Comments {
		for _, comment := range commentGroup.List {
			commentLine := file.Line(comment.Pos())
			if commentLine == line || commentLine == line-1 {
				text := comment.Text
				if strings.Contains(text, "nolint:errordemote") || 
				   (strings.Contains(text, "nolint") && !strings.Contains(text, "nolint:")) {
					return true
				}
			}
		}
	}
	
	return false
}

// hasResilienceDoc checks if there's explicit documentation about resilience
func hasResilienceDoc(pass *analysis.Pass, pos token.Pos) bool {
	file := pass.Fset.File(pos)
	if file == nil {
		return false
	}

	line := file.Line(pos)
	
	// Check for comments in the 3 lines before the if statement
	for _, commentGroup := range pass.Files[0].Comments {
		for _, comment := range commentGroup.List {
			commentLine := file.Line(comment.Pos())
			if commentLine >= line-3 && commentLine < line {
				text := strings.ToLower(comment.Text)
				// Look for keywords indicating explicit resilience decision
				keywords := []string{
					"resilience:",
					"resilient:",
					"non-critical",
					"optional",
					"safe to ignore",
					"safe to continue",
					"safe default",
					"may not exist",
				}
				for _, keyword := range keywords {
					if strings.Contains(text, keyword) {
						return true
					}
				}
			}
		}
	}
	
	return false
}

