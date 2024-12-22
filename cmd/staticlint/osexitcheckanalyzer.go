package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var OSExitCheckAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check os.Exit() in main()",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	var isMainPkg, isMainFunc, isCallOSExit bool

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.File:
				isMainPkg = x.Name.Name == "main"
			case *ast.FuncDecl:
				isMainFunc = x.Name.Name == "main"
			case *ast.CallExpr:
				isCallOSExit = types.ExprString(x.Fun) == "os.Exit"

				if isMainPkg && isMainFunc && isCallOSExit {
					pass.Reportf(x.Pos(), "os.Exit called")
				}
			}

			return true
		})
	}

	return nil, nil
}
