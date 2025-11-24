package main

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// osexit проверяет, что в функции main пакета main не используется прямой вызов os.Exit.
// Это позволяет корректно обрабатывать ошибки и выполнять defer-функции перед завершением программы.
//
// Анализатор проверяет:
//   - Прямые вызовы os.Exit в функции main
//   - Вызовы log.Fatal, log.Fatalf, log.Fatalln (которые внутри вызывают os.Exit)
//
// Рекомендуется использовать возврат ошибки из main или os.Exit только в крайних случаях
// после выполнения всех defer-функций.
var osexit = &analysis.Analyzer{
	Name:     "osexit",
	Doc:      "запрещает прямой вызов os.Exit в функции main пакета main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      osexitRun,
}

func osexitRun(pass *analysis.Pass) (interface{}, error) {
	// Проверяем, что TypesInfo доступен
	if pass.TypesInfo == nil {
		return nil, nil
	}

	for _, file := range pass.Files {
		// Пропускаем файлы не из пакета main
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Пропускаем, если это не функция main
			if fn.Name.Name != "main" {
				return true
			}

			// Пропускаем, если нет тела функции
			if fn.Body == nil {
				return true
			}

			// Обходим все узлы в теле функции main
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				callExpr, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := selExpr.X.(*ast.Ident)
				if !ok {
					return true
				}

				// Используем TypesInfo для проверки пакета
				obj, ok := pass.TypesInfo.Uses[ident]
				if !ok {
					return true
				}

				pkgName, ok := obj.(*types.PkgName)
				if !ok {
					return true
				}

				pkgPath := pkgName.Imported().Path()

				// Проверяем os.Exit
				if pkgPath == "os" && selExpr.Sel.Name == "Exit" {
					pass.Reportf(callExpr.Pos(),
						"прямой вызов os.Exit в функции main запрещен, используйте возврат ошибки")
					return true
				}

				// Проверяем log.Fatal, log.Fatalf, log.Fatalln
				if pkgPath == "log" && strings.HasPrefix(selExpr.Sel.Name, "Fatal") {
					pass.Reportf(callExpr.Pos(),
						"вызов log.%s в функции main запрещен (внутри вызывает os.Exit), используйте возврат ошибки", selExpr.Sel.Name)
					return true
				}

				return true
			})

			return true
		})
	}

	return nil, nil
}
