package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func runCallLinter(pass *analysis.Pass, n ast.Node) bool {
	// Проверяем, является ли узел вызовом функции
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return true
	}

	// Проверяем, что вызывается селектор (например, log.Fatal, os.Exit)
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

		// Проверяем конкретные вызовы log.Fatal, log.Fatalf, log.Fatalln, os.Exit
		if (pkgPath == "log" && methodName == "Fatal") ||
			(pkgPath == "log" && methodName == "Fatalf") ||
			(pkgPath == "log" && methodName == "Fatalln") ||
			(pkgPath == "os" && methodName == "Exit") {
			// Проверяем, находимся ли мы внутри функции main пакета main
			if !isInsideMainFunction(pass, n) {
				pass.Reportf(n.Pos(), "использование %s.%s запрещено вне функции main пакета main", pkgPath, methodName)
			}
		}
	}

	return true
}

// isInsideMainFunction проверяет, находится ли узел внутри функции main пакета main
func isInsideMainFunction(pass *analysis.Pass, node ast.Node) bool {
	// Проверяем, что пакет называется main
	if pass.Pkg.Name() != "main" {
		return false
	}

	nodePos := node.Pos()
	found := false

	// Ищем функцию main в файлах пакета
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				// Проверяем, что это функция main
				if funcDecl.Name != nil && funcDecl.Name.Name == "main" {
					// Проверяем, что узел находится внутри тела функции main
					if funcDecl.Body != nil {
						bodyStart := funcDecl.Body.Pos()
						bodyEnd := funcDecl.Body.End()
						if nodePos >= bodyStart && nodePos <= bodyEnd {
							found = true
							return false // останавливаем обход, так как нашли
						}
					}
				}
			}
			return true
		})
		if found {
			break
		}
	}

	return found
}
