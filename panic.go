package gocket

import (
	"fmt"
	"runtime"
)

func RecoverPanic() error {
	if x := recover(); x != nil {
		i := 0
		panicMsg := fmt.Sprintf("%v\n", x)
		funcName, file, line, ok := runtime.Caller(i)

		for ok {
			i++
			panicMsg += fmt.Sprintf("PrintPanicStack. [func]: %s, [file]: %s, [line]: %d\n", runtime.FuncForPC(funcName).Name(), file, line)
			funcName, file, line, ok = runtime.Caller(i)
		}

		return NewErrorFromString(PanicError, panicMsg)
	}

	return nil
}
