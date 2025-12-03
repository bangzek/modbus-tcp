//go:build debug

package modbus

import (
	"fmt"
	"os"
)

func Alloc() int {
	return alloc
}

var lname string

func Debugf(name, f string, a ...any) {
	if lname != name {
		fmt.Fprintf(os.Stderr, name+"\t"+f+"\n", a...)
		lname = name
	}
}
