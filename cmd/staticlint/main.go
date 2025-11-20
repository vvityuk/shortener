// Package main предоставляет multichecker - объединенный статический анализатор кода.
//
// Multichecker включает в себя:
//
// Стандартные анализаторы из golang.org/x/tools/go/analysis/passes:
//   - asmdecl: проверка соответствия объявлений ассемблера
//   - assign: проверка бесполезных присваиваний
//   - atomic: проверка неправильного использования sync/atomic
//   - bools: проверка распространенных ошибок с булевыми значениями
//   - buildtag: проверка корректности build tags
//   - cgocall: проверка правильности вызовов CGO
//   - composite: проверка литералов структур
//   - copylock: проверка копирования мьютексов
//   - deepequalerrors: проверка использования reflect.DeepEqual с ошибками
//   - errorsas: проверка правильности использования errors.As
//   - httpresponse: проверка закрытия HTTP-ответов
//   - ifaceassert: проверка утверждений типов интерфейсов
//   - loopclosure: проверка замыканий в циклах
//   - lostcancel: проверка отмены контекста
//   - nilfunc: проверка вызовов nil-функций
//   - printf: проверка форматирования строк
//   - shift: проверка сдвигов вне допустимого диапазона
//   - stdmethods: проверка соответствия стандартным интерфейсам
//   - stringintconv: проверка преобразований строк в числа
//   - structtag: проверка тегов структур
//   - testinggoroutine: проверка использования горутин в тестах
//   - tests: проверка тестов
//   - unmarshal: проверка unmarshal операций
//   - unreachable: проверка недостижимого кода
//   - unsafeptr: проверка небезопасных указателей
//   - unusedresult: проверка неиспользуемых результатов функций
//
// Все анализаторы класса SA из staticcheck.io:
//   - SA1000-SA1030: неправильное использование стандартных пакетов
//   - SA2000-SA2003: проблемы с конкурентностью
//   - SA3000-SA3001: проблемы в тестах
//   - SA4000-SA4023: логические ошибки и сравнения
//   - SA5000-SA5012: проблемы с производительностью и корректностью
//   - SA6000-SA6005: неэффективные операции
//   - SA9000-SA9005: стилистические проблемы
//
// Анализаторы других классов из staticcheck.io:
//   - S1000-S1040: упрощение кода (все анализаторы класса S)
//   - ST1000: стиль именования (проверка имен пакетов)
//   - ST1001: стиль кода (проверка дублирования строк)
//
// Дополнительные публичные анализаторы:
//   - errcheck: проверка необработанных ошибок
//   - ineffassign: проверка неэффективных присваиваний
//
// Собственный анализатор:
//   - osexit: запрещает прямой вызов os.Exit в функции main пакета main
//
// Запуск:
//
//		go run ./cmd/staticlint ./...
//
// Или после сборки:
//
//		go build -o staticlint ./cmd/staticlint
//		./staticlint ./...
package main

import (
	"strings"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	// Собираем все анализаторы
	analyzers := []*analysis.Analyzer{
		// Стандартные анализаторы из golang.org/x/tools/go/analysis/passes
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}

	// Добавляем все анализаторы класса SA из staticcheck
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Добавляем все анализаторы класса S (упрощение кода)
	for _, v := range simple.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	// Добавляем только ST1000 и ST1001 из класса ST (стиль кода)
	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1000" || v.Analyzer.Name == "ST1001" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Дополнительные публичные анализаторы
	analyzers = append(analyzers,
		ineffassign.Analyzer,    // проверка неэффективных присваиваний
		errcheck.Analyzer,       // проверка необработанных ошибок
	)

	// Собственный анализатор
	analyzers = append(analyzers, osExitAnalyzer)

	// Запускаем multichecker
	multichecker.Main(analyzers...)
}

