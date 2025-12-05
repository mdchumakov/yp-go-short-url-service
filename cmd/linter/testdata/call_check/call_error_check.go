package call_check

import (
	"log"
	"os"
)

func errLogFatalInFunction() {
	log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
}

func errOsExitInFunction() {
	os.Exit(1) // want "использование os.Exit запрещено вне функции main пакета main"
}

func errLogFatalInAnonymousFunc() {
	func() {
		log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
	}()
}

func errOsExitInAnonymousFunc() {
	func() {
		os.Exit(1) // want "использование os.Exit запрещено вне функции main пакета main"
	}()
}

func errLogFatalInIf() {
	if true {
		log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
	}
}

func errOsExitInIf() {
	if true {
		os.Exit(1) // want "использование os.Exit запрещено вне функции main пакета main"
	}
}

func errLogFatalInFor() {
	for i := 0; i < 10; i++ {
		log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
	}
}

func errOsExitInFor() {
	for i := 0; i < 10; i++ {
		os.Exit(1) // want "использование os.Exit запрещено вне функции main пакета main"
	}
}

func errLogFatalInSwitch() {
	switch {
	case true:
		log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
	}
}

func errOsExitInSwitch() {
	switch {
	case true:
		os.Exit(1) // want "использование os.Exit запрещено вне функции main пакета main"
	}
}

func errLogFatalInDefer() {
	defer func() {
		log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
	}()
}

func errOsExitInDefer() {
	defer func() {
		os.Exit(1) // want "использование os.Exit запрещено вне функции main пакета main"
	}()
}

func errLogFatalInGoroutine() {
	go func() {
		log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
	}()
}

func errOsExitInGoroutine() {
	go func() {
		os.Exit(1) // want "использование os.Exit запрещено вне функции main пакета main"
	}()
}

func errLogFatalWithArgs() {
	log.Fatal("error", "message") // want "использование log.Fatal запрещено вне функции main пакета main"
}

func errOsExitWithCode() {
	os.Exit(0) // want "использование os.Exit запрещено вне функции main пакета main"
}

// MyStruct тестовая структура для методов
type MyStruct struct{}

func (m MyStruct) errLogFatalInMethod() {
	log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
}

func (m MyStruct) errOsExitInMethod() {
	os.Exit(1) // want "использование os.Exit запрещено вне функции main пакета main"
}

func (m *MyStruct) errLogFatalInPointerMethod() {
	log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
}

func (m *MyStruct) errOsExitInPointerMethod() {
	os.Exit(1) // want "использование os.Exit запрещено вне функции main пакета main"
}

func errLogFatalInNestedBlocks() {
	{
		{
			log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
		}
	}
}

func errOsExitInNestedBlocks() {
	{
		{
			os.Exit(1) // want "использование os.Exit запрещено вне функции main пакета main"
		}
	}
}

func errLogFatalInSelect() {
	select {
	default:
		log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
	}
}

func errOsExitInSelect() {
	select {
	default:
		os.Exit(1) // want "использование os.Exit запрещено вне функции main пакета main"
	}
}
