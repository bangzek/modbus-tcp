package modbus_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/bangzek/modbus-tcp"
)

var _ = DescribeTable("Modbus Err",
	func(e ModbusErr, s string) {
		Expect(e.Error()).To(Equal(s))
	},
	Entry(nil, IllegalFunction, "Illegal Function"),
	Entry(nil, IllegalDataAddress, "Illegal Data Address"),
	Entry(nil, IllegalDataValue, "Illegal Data Value"),
	Entry(nil, SlaveDeviceFail, "Slave Device Failure"),
	Entry(nil, ModbusErr(5), "Err: 5"),
)
