// Package noosexit NoOsExit Analyzer
package noosexit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "reports direct calls to os.Exit in main function of main package",
	Run:  runNoOsExit,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.File)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		file, ok := n.(*ast.File)
		if !ok {
			return
		}

		// Пропускаем файлы в пакете "cache"
		if pass.Pkg.Path() == "cache" {
			return
		}

		// Пропускаем файлы в директории .cache
		if strings.Contains(pass.Fset.Position(file.Pos()).Filename, "/.cache/") {
			return
		}

		// Проверяем только пакет "main"
		if pass.Pkg.Name() != "main" {
			return
		}

		// Ищем функцию main
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" {
				continue
			}

			// Проверяем тело функции main на наличие вызова os.Exit
			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := selectorExpr.X.(*ast.Ident)
				if !ok {
					return true
				}

				if ident.Name == "os" && selectorExpr.Sel.Name == "Exit" {
					pass.Reportf(callExpr.Pos(), "direct call to os.Exit in main function of main package")
				}

				return true
			})
		}
	})

	return nil, nil
}
