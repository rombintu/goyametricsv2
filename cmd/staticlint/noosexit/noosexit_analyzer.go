// Package noosexit NoOsExit Analyzer
package noosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "reports direct calls to os.Exit in main function of main package",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		ident, ok := selectorExpr.X.(*ast.Ident)
		if !ok {
			return
		}

		if ident.Name == "os" && selectorExpr.Sel.Name == "Exit" {
			pass.Reportf(callExpr.Pos(), "direct call to os.Exit in main function of main package")
		}
	})

	return nil, nil
}
