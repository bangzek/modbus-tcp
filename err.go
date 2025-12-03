package modbus

import (
	"fmt"
	"unsafe"
)

type ModbusErr byte

const (
	IllegalFunction ModbusErr = iota + 1
	IllegalDataAddress
	IllegalDataValue
	SlaveDeviceFail
)

func (e ModbusErr) Error() string {
	switch e {
	case IllegalFunction:
		return "Illegal Function"
	case IllegalDataAddress:
		return "Illegal Data Address"
	case IllegalDataValue:
		return "Illegal Data Value"
	case SlaveDeviceFail:
		// len:20
		return "Slave Device Failure"
	default:
		return fmt.Sprintf("Err: %d", e)
	}
}

type BadRxErr []byte

func (e BadRxErr) Error() string {
	// 12345678901234567890
	// invalid response: []
	h := hexs(e)
	l := 20 + h.Len()
	noteAlloc(l)
	b := make([]byte, 0, l)
	b = append(b, "invalid response: ["...)
	b = hexs(e).Append(b)
	b = append(b, ']')
	return unsafe.String(&b[0], len(b))
}
