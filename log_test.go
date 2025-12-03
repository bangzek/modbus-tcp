package modbus_test

import (
	"fmt"

	. "github.com/bangzek/modbus-tcp"
)

type Log struct {
	Msgs []string
}

func NewLog() *Log {
	l := &Log{}
	DebugLogFunc = l.debugLog
	InfoLogFunc = l.infoLog
	ErrorLogFunc = l.errorLog
	return l
}

func (l *Log) debugLog(format string, v ...interface{}) {
	l.Msgs = append(l.Msgs, "D:"+fmt.Sprintf(format, v...))
}

func (l *Log) infoLog(format string, v ...interface{}) {
	l.Msgs = append(l.Msgs, "I:"+fmt.Sprintf(format, v...))
}

func (l *Log) errorLog(format string, v ...interface{}) {
	l.Msgs = append(l.Msgs, "E:"+fmt.Sprintf(format, v...))
}
