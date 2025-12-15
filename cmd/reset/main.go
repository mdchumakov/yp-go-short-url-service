package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

func main() {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedName | packages.NeedFiles,
	}

	parsedPackages, err := packages.Load(cfg, "./...")
	if err != nil {
		panic(err)
	}
	for _, pkg := range parsedPackages {

		structs := make([]tmpStruct, 0)
		for _, file := range pkg.Syntax {
			processFile(pkg.Fset, pkg.TypesInfo, file, &structs)
		}

		if len(structs) > 0 {
			enum := tmpEnum{
				PackageName: pkg.Name,
				Structs:     structs,
			}
			if err := generateResetMethods(&enum, pkg.Dir); err != nil {
				panic(err)
			}
		}

	}
}

func processFile(fset *token.FileSet, typesInfo *types.Info, file *ast.File, structs *[]tmpStruct) {
	commentMap := ast.NewCommentMap(fset, file, file.Comments)

	ast.Inspect(file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}

		for _, spec := range genDecl.Specs {
			typeSpec, isTypeSpec := spec.(*ast.TypeSpec)
			if !isTypeSpec {
				continue
			}
			astStruct, isStruct := typeSpec.Type.(*ast.StructType)
			if !isStruct {
				continue
			}

			// Проверяем комментарии, связанные с этой структурой
			comments := commentMap[n]
			if comments == nil {
				continue
			}

			for _, commentGroup := range comments {
				for _, comment := range commentGroup.List {
					if strings.Contains(comment.Text, "generate:reset") {
						fmt.Println("Found struct with // reset:", typeSpec.Name.Name)
						structName := typeSpec.Name.Name
						structFields := make([]tmpStructFiled, 0)

						for _, field := range astStruct.Fields.List {

							fieldType := typesInfo.TypeOf(field.Type)
							if fieldType == nil {
								continue
							}

							// Обрабатываем все имена полей (может быть несколько: a, b int)
							if len(field.Names) > 0 {
								// Обычные именованные поля
								for _, fieldName := range field.Names {
									structFields = append(structFields, tmpStructFiled{
										Name: fieldName.Name,
										Type: fieldType,
									})
								}
							} else {
								// Встроенное поле (embedded field) - используем тип как имя
								structFields = append(structFields, tmpStructFiled{
									Name: types.TypeString(fieldType, nil), // для встроенных полей имя = тип
									Type: fieldType,
								})
							}
						}

						*structs = append(
							*structs,
							tmpStruct{
								Name:   structName,
								Fields: structFields,
							},
						)
					}
				}
			}
		}

		return true
	})

}

func generateResetMethods(enum *tmpEnum, pkgPath string) error {
	var buf bytes.Buffer

	tmp := template.Must(template.New("reset").Parse(tmpResetMethod))

	err := tmp.Execute(&buf, enum)
	if err != nil {
		return err
	}

	filePath := filepath.Join(pkgPath, genFileName)
	err = os.WriteFile(filePath, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
