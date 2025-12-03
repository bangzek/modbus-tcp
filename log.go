package modbus

import "runtime/debug"

var (
	ErrorLogFunc func(string, ...interface{})
	InfoLogFunc  func(string, ...interface{})
	DebugLogFunc func(string, ...interface{})
)

func logPanic() {
	if err := recover(); err != nil {
		errorLog("PANIC: %s\n%s", err, debug.Stack())
	}
}

func errorLog(format string, v ...interface{}) {
	if ErrorLogFunc != nil {
		ErrorLogFunc(format, v...)
	}
}

func log(format string, v ...interface{}) {
	if InfoLogFunc != nil {
		InfoLogFunc(format, v...)
	}
}

func debugLog(format string, v ...interface{}) {
	if DebugLogFunc != nil {
		DebugLogFunc(format, v...)
	}
}
