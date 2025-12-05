package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func runCallLinter(pass *analysis.Pass, n ast.Node) bool {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return true
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return true
	}

	// Получаем объект метода через Uses
	obj := pass.TypesInfo.Uses[sel.Sel]
	if obj == nil {
		return true
	}

	// Проверяем, что это функция из нужного пакета
	if pkg := obj.Pkg(); pkg != nil {
		pkgPath := pkg.Path()
		methodName := obj.Name()

		// Проверяем конкретные вызовы
		if (pkgPath == "log" && methodName == "Fatal") ||
			(pkgPath == "log" && methodName == "Fatalf") ||
			(pkgPath == "log" && methodName == "Fatalln") ||
			(pkgPath == "os" && methodName == "Exit") {
			pass.Reportf(n.Pos(), "использование %s.%s запрещено вне функции main пакета main", pkgPath, methodName)
		}
	}

	return true
}
