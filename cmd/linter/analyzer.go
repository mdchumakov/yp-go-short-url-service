package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "shortener_lint",
	Doc:  "Линтер проверяющий использования функции в panic, log.Fatal, os.Exit",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool { return runPanicLinter(pass, n) })
		ast.Inspect(f, func(n ast.Node) bool { return runCallLinter(pass, n) })
	}
	return nil, nil
}
