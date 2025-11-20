package main

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// osExitAnalyzer проверяет, что в функции main пакета main не используется прямой вызов os.Exit.
// Это позволяет корректно обрабатывать ошибки и выполнять defer-функции перед завершением программы.
//
// Анализатор проверяет:
//   - Прямые вызовы os.Exit в функции main
//   - Вызовы log.Fatal, log.Fatalf, log.Fatalln (которые внутри вызывают os.Exit)
//
// Рекомендуется использовать возврат ошибки из main или os.Exit только в крайних случаях
// после выполнения всех defer-функций.
var osExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "запрещает прямой вызов os.Exit в функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Проверяем, что это пакет main
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			// Ищем функцию main
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				return true
			}

			// Проверяем тело функции main
			if fn.Body == nil {
				return true
			}

			// Обходим все узлы в теле функции main
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				// Проверяем вызовы функций
				callExpr, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Проверяем вызовы через селектор (например, os.Exit, log.Fatal)
				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok {
						// Проверяем os.Exit
						if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
							pass.Reportf(callExpr.Pos(),
								"прямой вызов os.Exit в функции main запрещен, используйте возврат ошибки")
							return true
						}

						// Проверяем log.Fatal, log.Fatalf, log.Fatalln
						if ident.Name == "log" && strings.HasPrefix(selExpr.Sel.Name, "Fatal") {
							pass.Reportf(callExpr.Pos(),
								"вызов log.%s в функции main запрещен (внутри вызывает os.Exit), используйте возврат ошибки", selExpr.Sel.Name)
							return true
						}
					}
				}

				return true
			})

			return true
		})
	}

	return nil, nil
}
