//go:build debug

package modbus

var alloc int

func noteAlloc(x int) {
	alloc = x
}
