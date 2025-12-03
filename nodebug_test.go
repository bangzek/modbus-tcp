//go:build !debug

package modbus

func Alloc() int {
	return 0
}

func Debugf(name, f string, a ...any) {
	// do nothing
}
