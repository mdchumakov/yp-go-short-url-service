package panic_check

func errBaseCheck() {
	panic("panic") // want "использование функции panic запрещено"
}

func errCheckInFunc() {
	func() {
		panic("panic") // want "использование функции panic запрещено"
	}()
}

func errCheckInIf() {
	if true {
		panic("error") // want "использование функции panic запрещено"
	}
}

func errCheckInFor() {
	for i := 0; i < 10; i++ {
		panic("error") // want "использование функции panic запрещено"
	}
}

func errCheckInSwitch() {
	switch {
	case true:
		panic("error") // want "использование функции panic запрещено"
	}
}

func errCheckWithVariable() {
	err := "error message"
	panic(err) // want "использование функции panic запрещено"
}

func errCheckInDefer() {
	defer func() {
		panic("error") // want "использование функции panic запрещено"
	}()
}

func errCheckInGoroutine() {
	go func() {
		panic("error") // want "использование функции panic запрещено"
	}()
}

func errCheckWithNil() {
	panic(nil) // want "использование функции panic запрещено"
}

// MyStruct Тестовая структура
type MyStruct struct{}

func (m MyStruct) errCheckInMethod() {
	panic("error") // want "использование функции panic запрещено"
}

func (m *MyStruct) errCheckInPointerMethod() {
	panic("error") // want "использование функции panic запрещено"
}

func errCheckInNestedBlocks() {
	{
		{
			panic("error") // want "использование функции panic запрещено"
		}
	}
}

func errCheckWithErrorType() {
	var err error
	panic(err) // want "использование функции panic запрещено"
}

func errCheckInSelect() {
	select {
	default:
		panic("error") // want "использование функции panic запрещено"
	}
}
