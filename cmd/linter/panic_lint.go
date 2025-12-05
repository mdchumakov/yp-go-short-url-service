package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func runPanicLinter(pass *analysis.Pass, n ast.Node) bool {
	// Проверяем, является ли узел вызовом функции
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return true
	}

	// Проверяем, что вызываемая функция - это идентификатор
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return true
	}

	// Проверяем, что идентификатор соответствует функции panic и что panic встроенный
	obj := pass.TypesInfo.Uses[ident]
	if builtin, ok := obj.(*types.Builtin); ok && builtin.Name() == "panic" {
		pass.Reportf(n.Pos(), "использование функции %s запрещено", builtin.Name())
	}

	return true
}
