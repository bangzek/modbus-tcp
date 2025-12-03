package modbus_test

import (
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/bangzek/modbus-tcp"
)

var _ = Describe("ReadCoilsCmd", func() {
	var cmd *ReadCoilsCmd
	SetRx := func(b []byte) {
		BeforeEach(func() {
			rx := cmd.RxBytes()
			*rx = (*rx)[:len(b)]
			copy(*rx, b)
		})
	}

	String := func(x string) {
		It("has String", func() {
			Expect(cmd.String()).To(Equal(x))
		})
	}
	Count := func(x int) {
		It("has Count", func() {
			Expect(cmd.Count()).To(Equal(x))
		})
	}
	OnlyTx := func(dev byte, addr uint16, s string, count int, b []byte) {
		It("has Tx Bytes", func() {
			Expect(cmd.TxBytes()).To(Equal(b))
		})
		It("has Dev Addr", func() {
			Expect(cmd.DevAddr()).To(Equal(dev))
		})
		It("has Addr", func() {
			Expect(cmd.Addr()).To(Equal(addr))
		})
		It("has Tx String", func() {
			Expect(cmd.Tx()).To(Equal(s))
		})
		String(s + "\n[]")
		Count(count)
	}
	IsValidRx := func() {
		It("is Valid Rx", func() {
			Expect(cmd.IsValidRx()).To(BeTrue())
		})
	}
	RxBytes := func(x []byte) {
		It("has Rx Bytes", func() {
			Expect(*cmd.RxBytes()).To(Equal(x))
		})
	}
	RxString := func(x string) {
		It("has Rx String", func() {
			Expect(cmd.Rx()).To(Equal(x))
		})
	}
	GoodRx := func(rb []byte, tx, rx string, a []bool, b []byte) {
		IsValidRx()
		RxBytes(rb)
		RxString(rx)
		String(tx + "\n" + rx)
		Count(len(a))
		It("has no Err", func() {
			Expect(cmd.Err()).To(Succeed())
		})
		It("has Coils", func() {
			for i, x := range a {
				Expect(cmd.Coil(i)).To(Equal(x), "Coil "+strconv.Itoa(i))
			}
		})
		It("has Bytes", func() {
			Expect(cmd.Bytes()).To(Equal(b))
		})
	}

	Describe("Invalid New", func() {
		It("can't do no broadcast", func() {
			Expect(func() {
				NewReadCoilsCmd(0, 2, 1)
			}).Should(PanicWith("could not broadcast ReadCoilsCmd"))
		})
		It("can't read zero", func() {
			Expect(func() {
				NewReadCoilsCmd(1, 2, 0)
			}).Should(PanicWith("zero count"))
		})
		It("can't read beyond 2000", func() {
			Expect(func() {
				NewReadCoilsCmd(1, 2, 2001)
			}).Should(PanicWith("count too many: 2001"))
		})
		It("can't overflow address", func() {
			Expect(func() {
				NewReadCoilsCmd(1, 63537, 2000)
			}).Should(PanicWith("address overflow: 63537, 2000"))
		})
	})

	Context("one", func() {
		const dev byte = 3
		const addr uint16 = 2
		BeforeEach(func() {
			cmd = NewReadCoilsCmd(dev, addr, 1)
		})

		const tx = "0000 3<-RC  2:1"
		Context("New", func() {
			OnlyTx(dev, addr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 1, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})
		})

		Context("Tx Id changed", func() {
			const id uint16 = 12345
			const tx = "3039 3<-RC  2:1"
			BeforeEach(func() {
				Expect(cmd.TxId()).To(BeZero())
				cmd.SetTxId(id)
			})

			OnlyTx(dev, addr, tx, 1, []byte{
				byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
				dev, 1, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})

			Context("Rx", func() {
				b := []byte{
					byte(id >> 8), byte(id & 0xff), 0, 0, 0, 4,
					dev, 1, 1, 0b0,
				}
				const rx = "3039 3->RC  1[0]"

				SetRx(b)
				GoodRx(b, tx, rx, []bool{false}, []byte{0b0})
			})
		})

		Context("Dev Addr changed", func() {
			const ndev byte = 13
			const tx = "0000 13<-RC  2:1"
			BeforeEach(func() {
				cmd.SetDevAddr(ndev)
			})

			OnlyTx(ndev, addr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				ndev, 1, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})
		})

		Context("Addr changed", func() {
			const naddr uint16 = 23
			const tx = "0000 3<-RC  23:1"
			BeforeEach(func() {
				cmd.SetAddr(naddr)
			})

			OnlyTx(dev, naddr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 1, byte(naddr >> 8), byte(naddr & 0xFF), 0, 1,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 4, dev, 1, 1, 0b1}
			const rx = "0000 3->RC  1[1]"

			SetRx(b)
			GoodRx(b, tx, rx, []bool{true}, []byte{0b1})
			It("can't read -1 coil", func() {
				Expect(func() {
					cmd.Coil(-1)
				}).Should(PanicWith("invalid i: -1"))
			})
			It("can't read too many coil", func() {
				Expect(func() {
					cmd.Coil(1)
				}).Should(PanicWith("invalid i: 1"))
			})
		})

		Context("Err Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 3, dev, 0x81, 4}
			const rx = "0000 3->RC  Slave Device Failure"

			SetRx(b)
			IsValidRx()
			RxBytes(b)
			RxString(rx)
			String(tx + "\n" + rx)
			It("has error ", func() {
				Expect(cmd.Err()).To(Equal(SlaveDeviceFail))
			})
		})
	})

	Context("five", func() {
		const dev byte = 4
		const addr uint16 = 321
		const tx = "0000 4<-RC  321:5"
		BeforeEach(func() {
			cmd = NewReadCoilsCmd(dev, addr, 5)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 5, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 1, byte(addr >> 8), byte(addr & 0xFF), 0, 5,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 4, dev, 1, 1, 0b1_0110}
			const rx = "0000 4->RC  5[0 1 1 0 1]"

			SetRx(b)
			GoodRx(b, tx, rx,
				[]bool{false, true, true, false, true},
				[]byte{0b1_0110})
		})
	})

	Context("six", func() {
		const dev byte = 5
		const addr uint16 = 1234
		const tx = "0000 5<-RC  1234:6"
		BeforeEach(func() {
			cmd = NewReadCoilsCmd(dev, addr, 6)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 6, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 1, byte(addr >> 8), byte(addr & 0xFF), 0, 6,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 4, dev, 1, 1, 0b10_1101}
			const rx = "0000 5->RC  6[1 0 1 1 0  1]"

			SetRx(b)
			GoodRx(b, tx, rx,
				[]bool{true, false, true, true, false, true},
				[]byte{0b10_1101})
		})
	})

	Context("ten", func() {
		const dev byte = 11
		const addr uint16 = 23456
		const tx = "0000 11<-RC  23456:10"
		BeforeEach(func() {
			cmd = NewReadCoilsCmd(dev, addr, 10)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 10, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 1, byte(addr >> 8), byte(addr & 0xFF), 0, 10,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 5, dev, 1, 2, 0b0110_1101, 0b11}
			const rx = "0000 11->RC  10[1 0 1 1 0  1 1 0 1 1]"

			SetRx(b)
			GoodRx(b, tx, rx, []bool{
				true, false, true, true, false, true, true, false, true, true,
			}, []byte{
				0b0110_1101, 0b11,
			})
		})
	})

	Context("eleven", func() {
		const dev byte = 123
		const addr uint16 = 45
		const tx = "0000 123<-RC  45:11"
		BeforeEach(func() {
			cmd = NewReadCoilsCmd(dev, addr, 11)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 11, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 1, byte(addr >> 8), byte(addr & 0xFF), 0, 11,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 5, dev, 1, 2, 0b0110_1101, 0b011}
			const rx = `0000 123->RC  11[
 1 0 1 1 0  1 1 0 1 1
 0
]`

			SetRx(b)
			GoodRx(b, tx, rx, []bool{
				true, false, true, true, false, true, true, false, true, true,
				false,
			}, []byte{
				0b0110_1101, 0b011,
			})
		})
	})

	Context("2K", func() {
		const dev byte = 123
		const addr uint16 = 63536
		const tx = "0000 123<-RC  63536:2000"
		BeforeEach(func() {
			cmd = NewReadCoilsCmd(dev, addr, 2000)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 2000, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 1, byte(addr >> 8), byte(addr & 0xFF), 7, 208,
			})
		})

		Context("Valid Rx", func() {
			b := make([]byte, 259)
			b[5] = 253
			b[6] = dev
			b[7] = 1
			b[8] = 250
			for i := 9; i < 259; i++ {
				b[i] = 0xA5
			}
			rx := `0000 123->RC  2000[` + strings.Repeat(`
 1 0 1 0 0  1 0 1 1 0
 1 0 0 1 0  1 1 0 1 0
 0 1 0 1 1  0 1 0 0 1
 0 1 1 0 1  0 0 1 0 1`, 50) + `
]`
			a := make([]bool, 2000)
			for i := 0; i < 250; i++ {
				a[i*8+0] = true
				a[i*8+1] = false
				a[i*8+2] = true
				a[i*8+3] = false
				a[i*8+4] = false
				a[i*8+5] = true
				a[i*8+6] = false
				a[i*8+7] = true
			}

			SetRx(b)
			GoodRx(b, tx, rx, a, b[9:])
		})
	})
})

var _ = Describe("ReadDInputsCmd", func() {
	var cmd *ReadDInputsCmd
	SetRx := func(b []byte) {
		BeforeEach(func() {
			rx := cmd.RxBytes()
			*rx = (*rx)[:len(b)]
			copy(*rx, b)
		})
	}

	String := func(x string) {
		It("has String", func() {
			Expect(cmd.String()).To(Equal(x))
		})
	}
	Count := func(x int) {
		It("has Count", func() {
			Expect(cmd.Count()).To(Equal(x))
		})
	}
	OnlyTx := func(dev byte, addr uint16, s string, count int, b []byte) {
		It("has Tx Bytes", func() {
			Expect(cmd.TxBytes()).To(Equal(b))
		})
		It("has Dev Addr", func() {
			Expect(cmd.DevAddr()).To(Equal(dev))
		})
		It("has Addr", func() {
			Expect(cmd.Addr()).To(Equal(addr))
		})
		It("has Tx String", func() {
			Expect(cmd.Tx()).To(Equal(s))
		})
		String(s + "\n[]")
		Count(count)
	}
	IsValidRx := func() {
		It("is Valid Rx", func() {
			Expect(cmd.IsValidRx()).To(BeTrue())
		})
	}
	RxBytes := func(x []byte) {
		It("has Rx Bytes", func() {
			Expect(*cmd.RxBytes()).To(Equal(x))
		})
	}
	RxString := func(x string) {
		It("has Rx String", func() {
			Expect(cmd.Rx()).To(Equal(x))
		})
	}
	GoodRx := func(rb []byte, tx, rx string, a []bool, b []byte) {
		IsValidRx()
		RxBytes(rb)
		RxString(rx)
		String(tx + "\n" + rx)
		Count(len(a))
		It("has no Err", func() {
			Expect(cmd.Err()).To(Succeed())
		})
		It("has Inputs", func() {
			for i, x := range a {
				Expect(cmd.Input(i)).To(Equal(x), "Input "+strconv.Itoa(i))
			}
		})
		It("has Bytes", func() {
			Expect(cmd.Bytes()).To(Equal(b))
		})
	}

	Describe("Invalid New", func() {
		It("can't do no broadcast", func() {
			Expect(func() {
				NewReadDInputsCmd(0, 2, 1)
			}).Should(PanicWith("could not broadcast ReadDInputsCmd"))
		})
		It("can't read zero", func() {
			Expect(func() {
				NewReadDInputsCmd(1, 2, 0)
			}).Should(PanicWith("zero count"))
		})
		It("can't read beyond 2000", func() {
			Expect(func() {
				NewReadDInputsCmd(1, 2, 2001)
			}).Should(PanicWith("count too many: 2001"))
		})
		It("can't overflow address", func() {
			Expect(func() {
				NewReadDInputsCmd(1, 63537, 2000)
			}).Should(PanicWith("address overflow: 63537, 2000"))
		})
	})

	Context("one", func() {
		const dev byte = 3
		const addr uint16 = 2
		BeforeEach(func() {
			cmd = NewReadDInputsCmd(dev, addr, 1)
		})

		const tx = "0000 3<-RDI 2:1"
		Context("New", func() {
			OnlyTx(dev, addr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 2, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})
		})

		Context("Tx Id changed", func() {
			const id uint16 = 23456
			const tx = "5BA0 3<-RDI 2:1"
			BeforeEach(func() {
				Expect(cmd.TxId()).To(BeZero())
				cmd.SetTxId(id)
			})

			OnlyTx(dev, addr, tx, 1, []byte{
				byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
				dev, 2, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})

			Context("Rx", func() {
				b := []byte{
					byte(id >> 8), byte(id & 0xff), 0, 0, 0, 4,
					dev, 2, 1, 0b0,
				}
				const rx = "5BA0 3->RDI 1[0]"

				SetRx(b)
				GoodRx(b, tx, rx, []bool{false}, []byte{0b0})
			})
		})

		Context("Dev Addr changed", func() {
			const ndev byte = 13
			const tx = "0000 13<-RDI 2:1"
			BeforeEach(func() {
				cmd.SetDevAddr(ndev)
			})

			OnlyTx(ndev, addr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				ndev, 2, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})
		})

		Context("Addr changed", func() {
			const naddr uint16 = 23
			const tx = "0000 3<-RDI 23:1"
			BeforeEach(func() {
				cmd.SetAddr(naddr)
			})

			OnlyTx(dev, naddr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 2, byte(naddr >> 8), byte(naddr & 0xFF), 0, 1,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 4, dev, 2, 1, 0b1}
			const rx = "0000 3->RDI 1[1]"

			SetRx(b)
			GoodRx(b, tx, rx, []bool{true}, []byte{0b1})
			It("can't read -1 coil", func() {
				Expect(func() {
					cmd.Input(-1)
				}).Should(PanicWith("invalid i: -1"))
			})
			It("can't read too many coil", func() {
				Expect(func() {
					cmd.Input(1)
				}).Should(PanicWith("invalid i: 1"))
			})
		})

		Context("Err Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 3, dev, 0x82, 3}
			const rx = "0000 3->RDI Illegal Data Value"

			SetRx(b)
			IsValidRx()
			RxBytes(b)
			RxString(rx)
			String(tx + "\n" + rx)
			It("has error ", func() {
				Expect(cmd.Err()).To(Equal(IllegalDataValue))
			})
		})
	})

	Context("five", func() {
		const dev byte = 4
		const addr uint16 = 321
		const tx = "0000 4<-RDI 321:5"
		BeforeEach(func() {
			cmd = NewReadDInputsCmd(dev, addr, 5)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 5, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 2, byte(addr >> 8), byte(addr & 0xFF), 0, 5,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 4, dev, 2, 1, 0b1_0110}
			const rx = "0000 4->RDI 5[0 1 1 0 1]"

			SetRx(b)
			GoodRx(b, tx, rx,
				[]bool{false, true, true, false, true},
				[]byte{0b1_0110})
		})
	})

	Context("six", func() {
		const dev byte = 5
		const addr uint16 = 1234
		const tx = "0000 5<-RDI 1234:6"
		BeforeEach(func() {
			cmd = NewReadDInputsCmd(dev, addr, 6)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 6, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 2, byte(addr >> 8), byte(addr & 0xFF), 0, 6,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 4,
				dev, 2, 1, 0b10_1101}
			const rx = "0000 5->RDI 6[1 0 1 1 0  1]"

			SetRx(b)
			GoodRx(b, tx, rx,
				[]bool{true, false, true, true, false, true},
				[]byte{0b10_1101})
		})
	})

	Context("ten", func() {
		const dev byte = 11
		const addr uint16 = 23456
		const tx = "0000 11<-RDI 23456:10"
		BeforeEach(func() {
			cmd = NewReadDInputsCmd(dev, addr, 10)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 10, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 2, byte(addr >> 8), byte(addr & 0xFF), 0, 10,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 5, dev, 2, 2, 0b0110_1101, 0b11}
			const rx = "0000 11->RDI 10[1 0 1 1 0  1 1 0 1 1]"

			SetRx(b)
			GoodRx(b, tx, rx, []bool{
				true, false, true, true, false, true, true, false, true, true,
			}, []byte{
				0b0110_1101, 0b11,
			})
		})
	})

	Context("eleven", func() {
		const dev byte = 123
		const addr uint16 = 45
		const tx = "0000 123<-RDI 45:11"
		BeforeEach(func() {
			cmd = NewReadDInputsCmd(dev, addr, 11)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 11, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 2, byte(addr >> 8), byte(addr & 0xFF), 0, 11,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 5, dev, 2, 2, 0b0110_1101, 0b011}
			const rx = `0000 123->RDI 11[
 1 0 1 1 0  1 1 0 1 1
 0
]`

			SetRx(b)
			GoodRx(b, tx, rx, []bool{
				true, false, true, true, false, true, true, false, true, true,
				false,
			}, []byte{
				0b0110_1101, 0b011,
			})
		})
	})

	Context("2K", func() {
		const dev byte = 123
		const addr uint16 = 63536
		const tx = "0000 123<-RDI 63536:2000"
		BeforeEach(func() {
			cmd = NewReadDInputsCmd(dev, addr, 2000)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 2000, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 2, byte(addr >> 8), byte(addr & 0xFF), 7, 208,
			})
		})

		Context("Valid Rx", func() {
			b := make([]byte, 259)
			b[5] = 253
			b[6] = dev
			b[7] = 2
			b[8] = 250
			for i := 9; i < 259; i++ {
				b[i] = 0xA5
			}
			rx := `0000 123->RDI 2000[` + strings.Repeat(`
 1 0 1 0 0  1 0 1 1 0
 1 0 0 1 0  1 1 0 1 0
 0 1 0 1 1  0 1 0 0 1
 0 1 1 0 1  0 0 1 0 1`, 50) + `
]`
			c := make([]bool, 2000)
			for i := 0; i < 250; i++ {
				c[i*8+0] = true
				c[i*8+1] = false
				c[i*8+2] = true
				c[i*8+3] = false
				c[i*8+4] = false
				c[i*8+5] = true
				c[i*8+6] = false
				c[i*8+7] = true
			}

			SetRx(b)
			GoodRx(b, tx, rx, c, b[9:])
		})
	})
})

var _ = Describe("ReadHRegsCmd", func() {
	var cmd *ReadHRegsCmd
	SetRx := func(b []byte) {
		BeforeEach(func() {
			rx := cmd.RxBytes()
			*rx = (*rx)[:len(b)]
			copy(*rx, b)
		})
	}

	String := func(x string) {
		It("has String", func() {
			Expect(cmd.String()).To(Equal(x))
		})
	}
	Count := func(x int) {
		It("has Count", func() {
			Expect(cmd.Count()).To(Equal(x))
		})
	}
	OnlyTx := func(dev byte, addr uint16, s string, count int, b []byte) {
		It("has Tx Bytes", func() {
			Expect(cmd.TxBytes()).To(Equal(b))
		})
		It("has Dev Addr", func() {
			Expect(cmd.DevAddr()).To(Equal(dev))
		})
		It("has Addr", func() {
			Expect(cmd.Addr()).To(Equal(addr))
		})
		It("has Tx String", func() {
			Expect(cmd.Tx()).To(Equal(s))
		})
		String(s + "\n[]")
		Count(count)
	}
	IsValidRx := func() {
		It("is Valid Rx", func() {
			Expect(cmd.IsValidRx()).To(BeTrue())
		})
	}
	RxBytes := func(x []byte) {
		It("has Rx Bytes", func() {
			Expect(*cmd.RxBytes()).To(Equal(x))
		})
	}
	RxString := func(x string) {
		It("has Rx String", func() {
			Expect(cmd.Rx()).To(Equal(x))
		})
	}
	GoodRx := func(rb []byte, tx, rx string, a []uint16, b []byte) {
		IsValidRx()
		RxBytes(rb)
		RxString(rx)
		String(tx + "\n" + rx)
		Count(len(a))
		It("has no Err", func() {
			Expect(cmd.Err()).To(Succeed())
		})
		It("has Regs", func() {
			for i, x := range a {
				Expect(cmd.Reg(i)).To(Equal(x), "Reg "+strconv.Itoa(i))
			}
		})
		It("has Bytes", func() {
			Expect(cmd.Bytes()).To(Equal(b))
		})
	}

	Describe("Invalid New", func() {
		It("can't do no broadcast", func() {
			Expect(func() {
				NewReadHRegsCmd(0, 2, 1)
			}).Should(PanicWith("could not broadcast ReadHRegsCmd"))
		})
		It("can't read zero", func() {
			Expect(func() {
				NewReadHRegsCmd(1, 2, 0)
			}).Should(PanicWith("zero count"))
		})
		It("can't read beyond 125", func() {
			Expect(func() {
				NewReadHRegsCmd(1, 2, 126)
			}).Should(PanicWith("count too many: 126"))
		})
		It("can't overflow address", func() {
			Expect(func() {
				NewReadHRegsCmd(1, 65412, 125)
			}).Should(PanicWith("address overflow: 65412, 125"))
		})
	})

	Context("one", func() {
		const dev byte = 3
		const addr uint16 = 2
		const tx = "0000 3<-RHR 2:1"
		BeforeEach(func() {
			cmd = NewReadHRegsCmd(dev, addr, 1)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 3, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})
		})

		Context("Tx Id changed", func() {
			const id uint16 = 34567
			const tx = "8707 3<-RHR 2:1"
			BeforeEach(func() {
				Expect(cmd.TxId()).To(BeZero())
				cmd.SetTxId(id)
			})

			OnlyTx(dev, addr, tx, 1, []byte{
				byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
				dev, 3, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})

			Context("Rx", func() {
				b := []byte{
					byte(id >> 8), byte(id & 0xff), 0, 0, 0, 5,
					dev, 3, 2, 0xDE, 0xAD,
				}
				const rx = "8707 3->RHR 1[57005]"

				SetRx(b)
				GoodRx(b, tx, rx, []uint16{0xDEAD}, []byte{0xDE, 0xAD})
			})
		})

		Context("Dev Addr changed", func() {
			const ndev byte = 13
			const tx = "0000 13<-RHR 2:1"
			BeforeEach(func() {
				cmd.SetDevAddr(ndev)
			})

			OnlyTx(ndev, addr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				ndev, 3, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})
		})

		Context("Addr changed", func() {
			const naddr uint16 = 23
			const tx = "0000 3<-RHR 23:1"
			BeforeEach(func() {
				cmd.SetAddr(naddr)
			})

			OnlyTx(dev, naddr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 3, byte(naddr >> 8), byte(naddr & 0xFF), 0, 1,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 5, dev, 3, 2, 0xDE, 0xAD}
			const rx = "0000 3->RHR 1[57005]"

			SetRx(b)
			GoodRx(b, tx, rx, []uint16{0xDEAD}, []byte{0xDE, 0xAD})
			It("can't read -1 coil", func() {
				Expect(func() {
					cmd.Reg(-1)
				}).Should(PanicWith("invalid i: -1"))
			})
			It("can't read too many coil", func() {
				Expect(func() {
					cmd.Reg(1)
				}).Should(PanicWith("invalid i: 1"))
			})
		})

		Context("Err Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 3, dev, 0x83, 2}
			const rx = "0000 3->RHR Illegal Data Address"

			SetRx(b)
			IsValidRx()
			RxBytes(b)
			RxString(rx)
			String(tx + "\n" + rx)
			It("has no Err", func() {
				Expect(cmd.Err()).To(Equal(IllegalDataAddress))
			})
		})
	})

	Context("five", func() {
		const dev byte = 4
		const addr uint16 = 321
		const tx = "0000 4<-RHR 321:5"
		BeforeEach(func() {
			cmd = NewReadHRegsCmd(dev, addr, 5)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 5, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 3, byte(addr >> 8), byte(addr & 0xFF), 0, 5,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 13,
				dev, 3, 10,
				0, 0, 255, 255, 0, 255, 255, 254, 1, 0,
			}
			const rx = "0000 4->RHR 5[    0 65535   255 65534   256]"

			SetRx(b)
			GoodRx(b, tx, rx,
				[]uint16{0, 65535, 255, 65534, 256},
				[]byte{0, 0, 255, 255, 0, 255, 255, 254, 1, 0})
		})
	})

	Context("six", func() {
		const dev byte = 5
		const addr uint16 = 1234
		const tx = "0000 5<-RHR 1234:6"
		BeforeEach(func() {
			cmd = NewReadHRegsCmd(dev, addr, 6)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 6, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 3, byte(addr >> 8), byte(addr & 0xFF), 0, 6,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 15,
				dev, 3, 12,
				0, 1, 0, 2, 0, 3, 1, 1, 1, 2, 1, 3,
			}
			const rx = "0000 5->RHR 6[    1     2     3   257   258 :   259]"

			SetRx(b)
			GoodRx(b, tx, rx,
				[]uint16{1, 2, 3, 257, 258, 259},
				[]byte{0, 1, 0, 2, 0, 3, 1, 1, 1, 2, 1, 3})
		})
	})

	Context("ten", func() {
		const dev byte = 11
		const addr uint16 = 23456
		const tx = "0000 11<-RHR 23456:10"
		BeforeEach(func() {
			cmd = NewReadHRegsCmd(dev, addr, 10)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 10, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 3, byte(addr >> 8), byte(addr & 0xFF), 0, 10,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 23,
				dev, 3, 20,
				0, 51, 53, 92, 153, 155, 169, 187, 195, 223,
				3, 65, 71, 75, 80, 86, 96, 171, 213, 248,
			}
			const rx = `0000 11->RHR 10[` +
				`   51 13660 39323 43451 50143 : ` +
				`  833 18251 20566 24747 54776]`

			SetRx(b)
			GoodRx(b, tx, rx, []uint16{
				51, 13660, 39323, 43451, 50143,
				833, 18251, 20566, 24747, 54776,
			}, []byte{
				0, 51, 53, 92, 153, 155, 169, 187, 195, 223,
				3, 65, 71, 75, 80, 86, 96, 171, 213, 248,
			})
		})
	})

	Context("eleven", func() {
		const dev byte = 123
		const addr uint16 = 45
		const tx = "0000 123<-RHR 45:11"
		BeforeEach(func() {
			cmd = NewReadHRegsCmd(dev, addr, 11)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 11, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 3, byte(addr >> 8), byte(addr & 0xFF), 0, 11,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 25,
				dev, 3, 22,
				0, 2, 2, 64, 3, 136, 1, 187, 231, 74,
				0, 6, 15, 128, 0, 48, 212, 110, 13, 212,
				9, 61,
			}
			const rx = `0000 123->RHR 11[
     2   576   904   443 59210 :     6  3968    48 54382  3540
  2365
]`

			SetRx(b)
			GoodRx(b, tx, rx, []uint16{
				2, 576, 904, 443, 59210,
				6, 3968, 48, 54382, 3540,
				2365,
			}, []byte{
				0, 2, 2, 64, 3, 136, 1, 187, 231, 74,
				0, 6, 15, 128, 0, 48, 212, 110, 13, 212,
				9, 61,
			})
		})
	})

	Context("125", func() {
		const dev byte = 123
		const addr uint16 = 65410
		const tx = "0000 123<-RHR 65410:125"
		BeforeEach(func() {
			cmd = NewReadHRegsCmd(dev, addr, 125)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 125, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 3, byte(addr >> 8), byte(addr & 0xFF), 0, 125,
			})
		})

		Context("Valid Rx", func() {
			b := make([]byte, 259)
			b[5] = 253
			b[6] = dev
			b[7] = 3
			b[8] = 250
			for i := 9; i < 259; i += 10 {
				b[i] = 0
				b[i+1] = 7
				b[i+2] = 0
				b[i+3] = 87
				b[i+4] = 0
				b[i+5] = 233
				b[i+6] = 5
				b[i+7] = 205
				b[i+8] = 220
				b[i+9] = 203
			}
			rx := `0000 123->RHR 125[` + strings.Repeat(`
     7    87   233  1485 56523 :     7    87   233  1485 56523`, 12) + `
     7    87   233  1485 56523
]`
			a := make([]uint16, 125)
			for i := 0; i < 125; i += 5 {
				a[i] = 7
				a[i+1] = 87
				a[i+2] = 233
				a[i+3] = 1485
				a[i+4] = 56523
			}

			SetRx(b)
			GoodRx(b, tx, rx, a, b[9:])
		})
	})
})

var _ = Describe("ReadIRegsCmd", func() {
	var cmd *ReadIRegsCmd
	SetRx := func(b []byte) {
		BeforeEach(func() {
			rx := cmd.RxBytes()
			*rx = (*rx)[:len(b)]
			copy(*rx, b)
		})
	}

	String := func(x string) {
		It("has String", func() {
			Expect(cmd.String()).To(Equal(x))
		})
	}
	Count := func(x int) {
		It("has Count", func() {
			Expect(cmd.Count()).To(Equal(x))
		})
	}
	OnlyTx := func(dev byte, addr uint16, s string, count int, b []byte) {
		It("has Tx Bytes", func() {
			Expect(cmd.TxBytes()).To(Equal(b))
		})
		It("has Dev Addr", func() {
			Expect(cmd.DevAddr()).To(Equal(dev))
		})
		It("has Addr", func() {
			Expect(cmd.Addr()).To(Equal(addr))
		})
		It("has Tx String", func() {
			Expect(cmd.Tx()).To(Equal(s))
		})
		String(s + "\n[]")
		Count(count)
	}
	IsValidRx := func() {
		It("is Valid Rx", func() {
			Expect(cmd.IsValidRx()).To(BeTrue())
		})
	}
	RxBytes := func(x []byte) {
		It("has Rx Bytes", func() {
			Expect(*cmd.RxBytes()).To(Equal(x))
		})
	}
	RxString := func(x string) {
		It("has Rx String", func() {
			Expect(cmd.Rx()).To(Equal(x))
		})
	}
	GoodRx := func(rb []byte, tx, rx string, a []uint16, b []byte) {
		IsValidRx()
		RxBytes(rb)
		RxString(rx)
		String(tx + "\n" + rx)
		Count(len(a))
		It("has no Err", func() {
			Expect(cmd.Err()).To(Succeed())
		})
		It("has Regs", func() {
			for i, x := range a {
				Expect(cmd.Reg(i)).To(Equal(x), "Reg "+strconv.Itoa(i))
			}
		})
		It("has Bytes", func() {
			Expect(cmd.Bytes()).To(Equal(b))
		})
	}

	Describe("Invalid New", func() {
		It("can't do no broadcast", func() {
			Expect(func() {
				NewReadIRegsCmd(0, 2, 1)
			}).Should(PanicWith("could not broadcast ReadIRegsCmd"))
		})
		It("can't read zero", func() {
			Expect(func() {
				NewReadIRegsCmd(1, 2, 0)
			}).Should(PanicWith("zero count"))
		})
		It("can't read beyond 125", func() {
			Expect(func() {
				NewReadIRegsCmd(1, 2, 126)
			}).Should(PanicWith("count too many: 126"))
		})
		It("can't overflow address", func() {
			Expect(func() {
				NewReadIRegsCmd(1, 65412, 125)
			}).Should(PanicWith("address overflow: 65412, 125"))
		})
	})

	Context("one", func() {
		const dev byte = 3
		const addr uint16 = 2
		const tx = "0000 3<-RIR 2:1"
		BeforeEach(func() {
			cmd = NewReadIRegsCmd(dev, addr, 1)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 4, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})
		})

		Context("Tx Id changed", func() {
			const id uint16 = 45678
			const tx = "B26E 3<-RIR 2:1"
			BeforeEach(func() {
				Expect(cmd.TxId()).To(BeZero())
				cmd.SetTxId(id)
			})

			OnlyTx(dev, addr, tx, 1, []byte{
				byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
				dev, 4, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})

			Context("Rx", func() {
				b := []byte{
					byte(id >> 8), byte(id & 0xff), 0, 0, 0, 5,
					dev, 4, 2, 0xDE, 0xAD,
				}
				const rx = "B26E 3->RIR 1[57005]"

				SetRx(b)
				GoodRx(b, tx, rx, []uint16{0xDEAD}, []byte{0xDE, 0xAD})
			})
		})

		Context("Dev Addr changed", func() {
			const ndev byte = 13
			const tx = "0000 13<-RIR 2:1"
			BeforeEach(func() {
				cmd.SetDevAddr(ndev)
			})

			OnlyTx(ndev, addr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				ndev, 4, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			})
		})

		Context("Addr changed", func() {
			const naddr uint16 = 23
			const tx = "0000 3<-RIR 23:1"
			BeforeEach(func() {
				cmd.SetAddr(naddr)
			})

			OnlyTx(dev, naddr, tx, 1, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 4, byte(naddr >> 8), byte(naddr & 0xFF), 0, 1,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 5, dev, 4, 2, 0xDE, 0xAD}
			const rx = "0000 3->RIR 1[57005]"

			SetRx(b)
			GoodRx(b, tx, rx, []uint16{0xDEAD}, []byte{0xDE, 0xAD})
			It("can't read -1 coil", func() {
				Expect(func() {
					cmd.Reg(-1)
				}).Should(PanicWith("invalid i: -1"))
			})
			It("can't read too much coil", func() {
				Expect(func() {
					cmd.Reg(1)
				}).Should(PanicWith("invalid i: 1"))
			})
		})

		Context("Err Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 3, dev, 0x84, 1}
			const rx = "0000 3->RIR Illegal Function"

			SetRx(b)
			IsValidRx()
			RxBytes(b)
			RxString(rx)
			String(tx + "\n" + rx)
			It("has no Err", func() {
				Expect(cmd.Err()).To(Equal(IllegalFunction))
			})
		})
	})

	Context("five", func() {
		const dev byte = 4
		const addr uint16 = 321
		const tx = "0000 4<-RIR 321:5"
		BeforeEach(func() {
			cmd = NewReadIRegsCmd(dev, addr, 5)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 5, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 4, byte(addr >> 8), byte(addr & 0xFF), 0, 5,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 13,
				dev, 4, 10,
				0, 0, 255, 255, 0, 255, 255, 254, 1, 0,
			}
			const rx = "0000 4->RIR 5[    0 65535   255 65534   256]"

			SetRx(b)
			GoodRx(b, tx, rx,
				[]uint16{0, 65535, 255, 65534, 256},
				[]byte{0, 0, 255, 255, 0, 255, 255, 254, 1, 0})
		})
	})

	Context("six", func() {
		const dev byte = 5
		const addr uint16 = 1234
		const tx = "0000 5<-RIR 1234:6"
		BeforeEach(func() {
			cmd = NewReadIRegsCmd(dev, addr, 6)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 6, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 4, byte(addr >> 8), byte(addr & 0xFF), 0, 6,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 15,
				dev, 4, 12,
				0, 1, 0, 2, 0, 3, 1, 1, 1, 2, 1, 3,
			}
			const rx = "0000 5->RIR 6[    1     2     3   257   258 :   259]"

			SetRx(b)
			GoodRx(b, tx, rx,
				[]uint16{1, 2, 3, 257, 258, 259},
				[]byte{0, 1, 0, 2, 0, 3, 1, 1, 1, 2, 1, 3})
		})
	})

	Context("ten", func() {
		const dev byte = 11
		const addr uint16 = 23456
		const tx = "0000 11<-RIR 23456:10"
		BeforeEach(func() {
			cmd = NewReadIRegsCmd(dev, addr, 10)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 10, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 4, byte(addr >> 8), byte(addr & 0xFF), 0, 10,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 23,
				dev, 4, 20,
				0, 51, 53, 92, 153, 155, 169, 187, 195, 223,
				3, 65, 71, 75, 80, 86, 96, 171, 213, 248,
			}
			const rx = `0000 11->RIR 10[` +
				`   51 13660 39323 43451 50143 : ` +
				`  833 18251 20566 24747 54776]`

			SetRx(b)
			GoodRx(b, tx, rx, []uint16{
				51, 13660, 39323, 43451, 50143,
				833, 18251, 20566, 24747, 54776,
			}, []byte{
				0, 51, 53, 92, 153, 155, 169, 187, 195, 223,
				3, 65, 71, 75, 80, 86, 96, 171, 213, 248,
			})
		})
	})

	Context("eleven", func() {
		const dev byte = 123
		const addr uint16 = 45
		const tx = "0000 123<-RIR 45:11"
		BeforeEach(func() {
			cmd = NewReadIRegsCmd(dev, addr, 11)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 11, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 4, byte(addr >> 8), byte(addr & 0xFF), 0, 11,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 25,
				dev, 4, 22,
				0, 2, 2, 64, 3, 136, 1, 187, 231, 74,
				0, 6, 15, 128, 0, 48, 212, 110, 13, 212,
				9, 61,
			}
			const rx = `0000 123->RIR 11[
     2   576   904   443 59210 :     6  3968    48 54382  3540
  2365
]`

			SetRx(b)
			GoodRx(b, tx, rx, []uint16{
				2, 576, 904, 443, 59210,
				6, 3968, 48, 54382, 3540,
				2365,
			}, []byte{
				0, 2, 2, 64, 3, 136, 1, 187, 231, 74,
				0, 6, 15, 128, 0, 48, 212, 110, 13, 212,
				9, 61,
			})
		})
	})

	Context("125", func() {
		const dev byte = 123
		const addr uint16 = 65410
		const tx = "0000 123<-RIR 65410:125"
		BeforeEach(func() {
			cmd = NewReadIRegsCmd(dev, addr, 125)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, 125, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 4, byte(addr >> 8), byte(addr & 0xFF), 0, 125,
			})
		})

		Context("Valid Rx", func() {
			b := make([]byte, 259)
			b[5] = 253
			b[6] = dev
			b[7] = 4
			b[8] = 250
			for i := 9; i < 259; i += 10 {
				b[i] = 0
				b[i+1] = 7
				b[i+2] = 0
				b[i+3] = 87
				b[i+4] = 0
				b[i+5] = 233
				b[i+6] = 5
				b[i+7] = 205
				b[i+8] = 220
				b[i+9] = 203
			}
			rx := `0000 123->RIR 125[` + strings.Repeat(`
     7    87   233  1485 56523 :     7    87   233  1485 56523`, 12) + `
     7    87   233  1485 56523
]`
			a := make([]uint16, 125)
			for i := 0; i < 125; i += 5 {
				a[i] = 7
				a[i+1] = 87
				a[i+2] = 233
				a[i+3] = 1485
				a[i+4] = 56523
			}

			SetRx(b)
			GoodRx(b, tx, rx, a, b[9:])
		})
	})
})

var _ = Describe("WriteCoilCmd", func() {
	var cmd *WriteCoilCmd
	SetRx := func(b []byte) {
		BeforeEach(func() {
			rx := cmd.RxBytes()
			*rx = (*rx)[:len(b)]
			copy(*rx, b)
		})
	}

	String := func(x string) {
		It("has String", func() {
			Expect(cmd.String()).To(Equal(x))
		})
	}
	OnlyTx := func(dev byte, addr uint16, s string, coil bool, b []byte) {
		It("has Tx Bytes", func() {
			Expect(cmd.TxBytes()).To(Equal(b))
		})
		It("has Dev Addr", func() {
			Expect(cmd.DevAddr()).To(Equal(dev))
		})
		It("has Addr", func() {
			Expect(cmd.Addr()).To(Equal(addr))
		})
		It("has Tx String", func() {
			Expect(cmd.Tx()).To(Equal(s))
		})
		It("has Coil", func() {
			Expect(cmd.Coil()).To(Equal(coil))
		})
		if dev == 0 {
			String(s)
		} else {
			String(s + "\n[]")
		}
	}
	IsValidRx := func() {
		It("is Valid Rx", func() {
			Expect(cmd.IsValidRx()).To(BeTrue())
		})
	}
	RxBytes := func(x []byte) {
		It("has Rx Bytes", func() {
			Expect(*cmd.RxBytes()).To(Equal(x))
		})
	}
	RxString := func(x string) {
		It("has Rx String", func() {
			Expect(cmd.Rx()).To(Equal(x))
		})
	}
	GoodRx := func(b []byte, tx, rx string) {
		IsValidRx()
		RxBytes(b)
		RxString(rx)
		String(tx + "\n" + rx)
		It("has no Err", func() {
			Expect(cmd.Err()).To(Succeed())
		})
	}
	ErrRx := func(b []byte, tx, rx string, err error) {
		IsValidRx()
		RxBytes(b)
		RxString(rx)
		String(tx + "\n" + rx)
		It("has Err", func() {
			Expect(cmd.Err()).To(Equal(err))
		})
	}

	Context("broadcast", func() {
		Context("true", func() {
			const addr uint16 = 258
			const tx = "0000 0<-W1C 258 true"
			BeforeEach(func() {
				cmd = NewWriteCoilCmd(0, addr, true)
			})

			Context("New", func() {
				OnlyTx(0, addr, tx, true, []byte{
					0, 0, 0, 0, 0, 6,
					0, 5, byte(addr >> 8), byte(addr & 0xFF), 0xFF, 0,
				})
			})

			Context("Dev Addr changed", func() {
				const ndev byte = 3
				const tx = "0000 3<-W1C 258 true"
				BeforeEach(func() {
					cmd.SetDevAddr(ndev)
				})

				OnlyTx(ndev, addr, tx, true, []byte{
					0, 0, 0, 0, 0, 6,
					ndev, 5, byte(addr >> 8), byte(addr & 0xFF), 0xFF, 0,
				})
			})

			Context("Addr changed", func() {
				const naddr uint16 = 23
				const tx = "0000 0<-W1C 23 true"
				BeforeEach(func() {
					cmd.SetAddr(naddr)
				})

				OnlyTx(0, naddr, tx, true, []byte{
					0, 0, 0, 0, 0, 6,
					0, 5, byte(naddr >> 8), byte(naddr & 0xFF), 0xFF, 0,
				})
			})

			Context("Coil changed", func() {
				const tx = "0000 0<-W1C 258 false"
				BeforeEach(func() {
					cmd.SetCoil(false)
				})

				OnlyTx(0, addr, tx, false, []byte{
					0, 0, 0, 0, 0, 6,
					0, 5, byte(addr >> 8), byte(addr & 0xFF), 0, 0,
				})
			})
		})

		Context("false", func() {
			const addr uint16 = 2
			BeforeEach(func() {
				cmd = NewWriteCoilCmd(0, addr, false)
			})

			const tx = "0000 0<-W1C 2 false"
			Context("New", func() {
				OnlyTx(0, addr, tx, false, []byte{
					0, 0, 0, 0, 0, 6,
					0, 5, byte(addr >> 8), byte(addr & 0xFF), 0, 0,
				})
			})

			Context("Dev Addr changed", func() {
				const ndev byte = 3
				const tx = "0000 3<-W1C 2 false"
				BeforeEach(func() {
					cmd.SetDevAddr(ndev)
				})

				OnlyTx(ndev, addr, tx, false, []byte{
					0, 0, 0, 0, 0, 6,
					ndev, 5, byte(addr >> 8), byte(addr & 0xFF), 0, 0,
				})
			})

			Context("Addr changed", func() {
				const naddr uint16 = 2003
				const tx = "0000 0<-W1C 2003 false"
				BeforeEach(func() {
					cmd.SetAddr(naddr)
				})

				OnlyTx(0, naddr, tx, false, []byte{
					0, 0, 0, 0, 0, 6,
					0, 5, byte(naddr >> 8), byte(naddr & 0xFF), 0, 0,
				})
			})

			Context("Coil changed", func() {
				const tx = "0000 0<-W1C 2 true"
				BeforeEach(func() {
					cmd.SetCoil(true)
				})

				OnlyTx(0, addr, tx, true, []byte{
					0, 0, 0, 0, 0, 6,
					0, 5, byte(addr >> 8), byte(addr & 0xFF), 0xFF, 0,
				})
			})
		})
	})

	Context("non broadcast", func() {
		Context("true", func() {
			const dev = 1
			const addr uint16 = 258
			const tx = "0000 1<-W1C 258 true"
			BeforeEach(func() {
				cmd = NewWriteCoilCmd(dev, addr, true)
			})

			Context("New", func() {
				OnlyTx(dev, addr, tx, true, []byte{
					0, 0, 0, 0, 0, 6,
					dev, 5, byte(addr >> 8), byte(addr & 0xFF), 0xFF, 0,
				})
			})

			Context("Tx Id changed", func() {
				const id uint16 = 56789
				const tx = "DDD5 1<-W1C 258 true"
				BeforeEach(func() {
					Expect(cmd.TxId()).To(BeZero())
					cmd.SetTxId(id)
				})

				OnlyTx(dev, addr, tx, true, []byte{
					byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
					dev, 5, byte(addr >> 8), byte(addr & 0xFF), 0xFF, 0,
				})

				Context("Rx", func() {
					b := []byte{
						byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
						dev, 5, byte(addr >> 8), byte(addr & 0xFF), 0xFF, 0,
					}
					const rx = "DDD5 1->W1C 258 true"

					SetRx(b)
					GoodRx(b, tx, rx)
				})
			})

			Context("Dev Addr changed", func() {
				const ndev byte = 3
				const tx = "0000 3<-W1C 258 true"
				BeforeEach(func() {
					cmd.SetDevAddr(ndev)
				})

				OnlyTx(ndev, addr, tx, true, []byte{
					0, 0, 0, 0, 0, 6,
					ndev, 5, byte(addr >> 8), byte(addr & 0xFF), 0xFF, 0,
				})
			})

			Context("Dev Addr changed to broadcast", func() {
				const tx = "0000 0<-W1C 258 true"
				BeforeEach(func() {
					cmd.SetDevAddr(0)
				})

				OnlyTx(0, addr, tx, true, []byte{
					0, 0, 0, 0, 0, 6,
					0, 5, byte(addr >> 8), byte(addr & 0xFF), 0xFF, 0,
				})
			})

			Context("Addr changed", func() {
				const naddr uint16 = 23
				const tx = "0000 1<-W1C 23 true"
				BeforeEach(func() {
					cmd.SetAddr(naddr)
				})

				OnlyTx(dev, naddr, tx, true, []byte{
					0, 0, 0, 0, 0, 6,
					dev, 5, byte(naddr >> 8), byte(naddr & 0xFF), 0xFF, 0,
				})
			})

			Context("Coil changed", func() {
				const tx = "0000 1<-W1C 258 false"
				BeforeEach(func() {
					cmd.SetCoil(false)
				})

				OnlyTx(dev, addr, tx, false, []byte{
					0, 0, 0, 0, 0, 6,
					dev, 5, byte(addr >> 8), byte(addr & 0xFF), 0, 0,
				})
			})

			Context("Valid Rx", func() {
				b := []byte{
					0, 0, 0, 0, 0, 6,
					dev, 5, byte(addr >> 8), byte(addr & 0xFF), 0xFF, 0,
				}
				const rx = "0000 1->W1C 258 true"

				SetRx(b)
				GoodRx(b, tx, rx)
			})

			Context("Err Rx", func() {
				b := []byte{0, 0, 0, 0, 0, 3, dev, 0x85, 4}
				const rx = "0000 1->W1C Slave Device Failure"

				SetRx(b)
				ErrRx(b, tx, rx, SlaveDeviceFail)
			})
		})

		Context("false", func() {
			const dev = 12
			const addr uint16 = 3456
			const tx = "0000 12<-W1C 3456 false"
			BeforeEach(func() {
				cmd = NewWriteCoilCmd(dev, addr, false)
			})

			Context("New", func() {
				OnlyTx(dev, addr, tx, false, []byte{
					0, 0, 0, 0, 0, 6,
					dev, 5, byte(addr >> 8), byte(addr & 0xFF), 0, 0,
				})
			})

			Context("Tx Id changed", func() {
				const id uint16 = 65432
				const tx = "FF98 12<-W1C 3456 false"
				BeforeEach(func() {
					Expect(cmd.TxId()).To(BeZero())
					cmd.SetTxId(id)
				})

				OnlyTx(dev, addr, tx, false, []byte{
					byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
					dev, 5, byte(addr >> 8), byte(addr & 0xFF), 0, 0,
				})

				Context("Rx", func() {
					b := []byte{
						byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
						dev, 5, byte(addr >> 8), byte(addr & 0xFF), 0, 0,
					}
					const rx = "FF98 12->W1C 3456 false"

					SetRx(b)
					GoodRx(b, tx, rx)
				})
			})

			Context("Dev Addr changed", func() {
				const ndev byte = 123
				const tx = "0000 123<-W1C 3456 false"
				BeforeEach(func() {
					cmd.SetDevAddr(ndev)
				})

				OnlyTx(ndev, addr, tx, false, []byte{
					0, 0, 0, 0, 0, 6,
					ndev, 5, byte(addr >> 8), byte(addr & 0xFF), 0, 0,
				})
			})

			Context("Dev Addr changed to broadcast", func() {
				const tx = "0000 0<-W1C 3456 false"
				BeforeEach(func() {
					cmd.SetDevAddr(0)
				})

				OnlyTx(0, addr, tx, false, []byte{
					0, 0, 0, 0, 0, 6,
					0, 5, byte(addr >> 8), byte(addr & 0xFF), 0, 0,
				})
			})

			Context("Addr changed", func() {
				const naddr uint16 = 3
				const tx = "0000 12<-W1C 3 false"
				BeforeEach(func() {
					cmd.SetAddr(naddr)
				})

				OnlyTx(dev, naddr, tx, false, []byte{
					0, 0, 0, 0, 0, 6,
					dev, 5, byte(naddr >> 8), byte(naddr & 0xFF), 0, 0,
				})
			})

			Context("Coil changed", func() {
				const tx = "0000 12<-W1C 3456 true"
				BeforeEach(func() {
					cmd.SetCoil(true)
				})

				OnlyTx(dev, addr, tx, true, []byte{
					0, 0, 0, 0, 0, 6,
					dev, 5, byte(addr >> 8), byte(addr & 0xFF), 0xFF, 0,
				})
			})

			Context("Valid Rx", func() {
				b := []byte{
					0, 0, 0, 0, 0, 6,
					dev, 5, byte(addr >> 8), byte(addr & 0xFF), 0, 0,
				}
				const rx = "0000 12->W1C 3456 false"

				SetRx(b)
				GoodRx(b, tx, rx)
			})

			Context("Err Rx", func() {
				b := []byte{0, 0, 0, 0, 0, 3, dev, 0x85, 2}
				const rx = "0000 12->W1C Illegal Data Address"

				SetRx(b)
				ErrRx(b, tx, rx, IllegalDataAddress)
			})
		})
	})
})

var _ = Describe("WriteRegCmd", func() {
	var cmd *WriteRegCmd
	SetRx := func(b []byte) {
		BeforeEach(func() {
			rx := cmd.RxBytes()
			*rx = (*rx)[:len(b)]
			copy(*rx, b)
		})
	}

	String := func(x string) {
		It("has String", func() {
			Expect(cmd.String()).To(Equal(x))
		})
	}
	OnlyTx := func(dev byte, addr uint16, s string, reg uint16, b []byte) {
		It("has Tx Bytes", func() {
			Expect(cmd.TxBytes()).To(Equal(b))
		})
		It("has Dev Addr", func() {
			Expect(cmd.DevAddr()).To(Equal(dev))
		})
		It("has Addr", func() {
			Expect(cmd.Addr()).To(Equal(addr))
		})
		It("has Tx String", func() {
			Expect(cmd.Tx()).To(Equal(s))
		})
		It("has Reg", func() {
			Expect(cmd.Reg()).To(Equal(reg))
		})
		if dev == 0 {
			String(s)
		} else {
			String(s + "\n[]")
		}
	}
	IsValidRx := func() {
		It("is Valid Rx", func() {
			Expect(cmd.IsValidRx()).To(BeTrue())
		})
	}
	RxBytes := func(x []byte) {
		It("has Rx Bytes", func() {
			Expect(*cmd.RxBytes()).To(Equal(x))
		})
	}
	RxString := func(x string) {
		It("has Rx String", func() {
			Expect(cmd.Rx()).To(Equal(x))
		})
	}
	GoodRx := func(b []byte, tx, rx string) {
		IsValidRx()
		RxBytes(b)
		RxString(rx)
		String(tx + "\n" + rx)
		It("has no Err", func() {
			Expect(cmd.Err()).To(Succeed())
		})
	}

	Context("broadcast", func() {
		const addr uint16 = 258
		const val uint16 = 0xBEEF
		const tx = "0000 0<-W1R 258 48879"
		BeforeEach(func() {
			cmd = NewWriteRegCmd(0, addr, val)
		})

		Context("New", func() {
			OnlyTx(0, addr, tx, val, []byte{
				0, 0, 0, 0, 0, 6,
				0, 6, byte(addr >> 8), byte(addr & 0xFF), 0xBE, 0xEF,
			})
		})

		Context("Dev Addr changed", func() {
			const ndev byte = 3
			const tx = "0000 3<-W1R 258 48879"
			BeforeEach(func() {
				cmd.SetDevAddr(ndev)
			})

			OnlyTx(ndev, addr, tx, val, []byte{
				0, 0, 0, 0, 0, 6,
				ndev, 6, byte(addr >> 8), byte(addr & 0xFF), 0xBE, 0xEF,
			})
		})

		Context("Addr changed", func() {
			const naddr uint16 = 23
			const tx = "0000 0<-W1R 23 48879"
			BeforeEach(func() {
				cmd.SetAddr(naddr)
			})

			OnlyTx(0, naddr, tx, val, []byte{
				0, 0, 0, 0, 0, 6,
				0, 6, byte(naddr >> 8), byte(naddr & 0xFF), 0xBE, 0xEF,
			})
		})

		Context("Reg changed", func() {
			const nval uint16 = 0xDEAD
			const tx = "0000 0<-W1R 258 57005"
			BeforeEach(func() {
				cmd.SetReg(nval)
			})

			OnlyTx(0, addr, tx, nval, []byte{
				0, 0, 0, 0, 0, 6,
				0, 6, byte(addr >> 8), byte(addr & 0xFF), 0xDE, 0xAD,
			})
		})
	})

	Context("non broadcast", func() {
		const dev = 1
		const addr uint16 = 258
		const val uint16 = 0xDEAD
		BeforeEach(func() {
			cmd = NewWriteRegCmd(dev, addr, val)
		})

		const tx = "0000 1<-W1R 258 57005"
		Context("New", func() {
			OnlyTx(dev, addr, tx, val, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 6, byte(addr >> 8), byte(addr & 0xFF), 0xDE, 0xAD,
			})
		})

		Context("Tx Id changed", func() {
			const id uint16 = 48879
			const tx = "BEEF 1<-W1R 258 57005"
			BeforeEach(func() {
				Expect(cmd.TxId()).To(BeZero())
				cmd.SetTxId(id)
			})

			OnlyTx(dev, addr, tx, val, []byte{
				byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
				dev, 6, byte(addr >> 8), byte(addr & 0xFF), 0xDE, 0xAD,
			})

			Context("Rx", func() {
				b := []byte{
					byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
					dev, 6, byte(addr >> 8), byte(addr & 0xFF), 0xDE, 0xAD,
				}
				const rx = "BEEF 1->W1R 258 57005"

				SetRx(b)
				GoodRx(b, tx, rx)
			})
		})

		Context("Dev Addr changed", func() {
			const ndev byte = 3
			const tx = "0000 3<-W1R 258 57005"
			BeforeEach(func() {
				cmd.SetDevAddr(ndev)
			})

			OnlyTx(ndev, addr, tx, val, []byte{
				0, 0, 0, 0, 0, 6,
				ndev, 6, byte(addr >> 8), byte(addr & 0xFF), 0xDE, 0xAD,
			})
		})

		Context("Dev Addr changed to broadcast", func() {
			const tx = "0000 0<-W1R 258 57005"
			BeforeEach(func() {
				cmd.SetDevAddr(0)
			})

			OnlyTx(0, addr, tx, val, []byte{
				0, 0, 0, 0, 0, 6,
				0, 6, byte(addr >> 8), byte(addr & 0xFF), 0xDE, 0xAD,
			})
		})

		Context("Addr changed", func() {
			const naddr uint16 = 23
			const tx = "0000 1<-W1R 23 57005"
			BeforeEach(func() {
				cmd.SetAddr(naddr)
			})

			OnlyTx(dev, naddr, tx, val, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 6, byte(naddr >> 8), byte(naddr & 0xFF), 0xDE, 0xAD,
			})
		})

		Context("Reg changed", func() {
			const tx = "0000 1<-W1R 258 48879"
			const nval = 0xBEEF
			BeforeEach(func() {
				cmd.SetReg(nval)
			})

			OnlyTx(dev, addr, tx, nval, []byte{
				0, 0, 0, 0, 0, 6,
				dev, 6, byte(addr >> 8), byte(addr & 0xFF), 0xBE, 0xEF,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 6, byte(addr >> 8), byte(addr & 0xFF), 0xDE, 0xAD,
			}
			const rx = "0000 1->W1R 258 57005"

			SetRx(b)
			GoodRx(b, tx, rx)
		})

		Context("Err Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 3, dev, 0x86, 2}
			const rx = "0000 1->W1R Illegal Data Address"

			SetRx(b)
			IsValidRx()
			RxBytes(b)
			RxString(rx)
			String(tx + "\n" + rx)
			It("has no Err", func() {
				Expect(cmd.Err()).To(Equal(IllegalDataAddress))
			})
		})
	})
})

var _ = Describe("WriteCoilsCmd", func() {
	var cmd *WriteCoilsCmd
	SetRx := func(b []byte) {
		BeforeEach(func() {
			rx := cmd.RxBytes()
			*rx = (*rx)[:len(b)]
			copy(*rx, b)
		})
	}

	String := func(x string) {
		It("has String", func() {
			Expect(cmd.String()).To(Equal(x))
		})
	}
	Count := func(x int) {
		It("has Count", func() {
			Expect(cmd.Count()).To(Equal(x))
		})
	}
	OnlyTx := func(dev byte, addr uint16, s string, a []bool, b []byte) {
		It("has Tx Bytes", func() {
			Expect(cmd.TxBytes()).To(Equal(b))
		})
		It("has Dev Addr", func() {
			Expect(cmd.DevAddr()).To(Equal(dev))
		})
		It("has Addr", func() {
			Expect(cmd.Addr()).To(Equal(addr))
		})
		It("has Tx String", func() {
			Expect(cmd.Tx()).To(Equal(s))
		})
		if dev == 0 {
			String(s)
		} else {
			String(s + "\n[]")
		}
		Count(len(a))
		It("has Coils", func() {
			for i, x := range a {
				Expect(cmd.Coil(i)).To(Equal(x), "Coil "+strconv.Itoa(i))
			}
		})
		It("has ByteCount", func() {
			Expect(cmd.ByteCount()).To(Equal(len(b) - 13))
		})
		It("has Byte", func() {
			for i := 0; i < len(b)-13; i++ {
				Expect(cmd.Byte(i)).To(Equal(b[13+i]), "Byte "+strconv.Itoa(i))
			}
		})
		It("has Bytes", func() {
			Expect(cmd.Bytes()).To(Equal(b[13:]))
		})
	}
	IsValidRx := func() {
		It("is Valid Rx", func() {
			Expect(cmd.IsValidRx()).To(BeTrue())
		})
	}
	RxBytes := func(x []byte) {
		It("has Rx Bytes", func() {
			Expect(*cmd.RxBytes()).To(Equal(x))
		})
	}
	RxString := func(x string) {
		It("has Rx String", func() {
			Expect(cmd.Rx()).To(Equal(x))
		})
	}
	GoodRx := func(rb []byte, tx, rx string) {
		IsValidRx()
		RxBytes(rb)
		RxString(rx)
		String(tx + "\n" + rx)
		It("has no Err", func() {
			Expect(cmd.Err()).To(Succeed())
		})
	}

	Describe("Invalid New", func() {
		It("can't write zero", func() {
			Expect(func() {
				NewWriteCoilsCmd(1, 2, []bool{})
			}).Should(PanicWith("empty values"))
		})
		It("can't write beyond 2000", func() {
			Expect(func() {
				NewWriteCoilsCmd(1, 2, make([]bool, 1969))
			}).Should(PanicWith("values too many: 1969"))
		})
		It("can't overflow address", func() {
			Expect(func() {
				NewWriteCoilsCmd(1, 63569, make([]bool, 1968))
			}).Should(PanicWith("address overflow: 63569, 1968"))
		})
	})

	Context("broadcast", func() {
		const addr uint16 = 2
		const tx = "0000 0<-WC  2:1[0]"
		vals := []bool{false}
		BeforeEach(func() {
			cmd = NewWriteCoilsCmd(0, addr, vals)
		})

		Context("New", func() {
			OnlyTx(0, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 8,
				0, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b0,
			})
		})

		Context("Dev Addr changed", func() {
			const ndev byte = 13
			const tx = "0000 13<-WC  2:1[0]"
			BeforeEach(func() {
				cmd.SetDevAddr(ndev)
			})

			OnlyTx(ndev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 8,
				13, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b0,
			})
		})

		Context("Addr changed", func() {
			const naddr uint16 = 0
			const tx = "0000 0<-WC  0:1[0]"
			BeforeEach(func() {
				cmd.SetAddr(naddr)
			})

			OnlyTx(0, naddr, tx, vals, []byte{
				0, 0, 0, 0, 0, 8,
				0, 15, byte(naddr >> 8), byte(naddr & 0xFF), 0, 1, 1, 0b0,
			})
		})

		It("can't get -1 coil", func() {
			Expect(func() {
				cmd.Coil(-1)
			}).Should(PanicWith("invalid i: -1"))
		})
		It("can't get too many coil", func() {
			Expect(func() {
				cmd.Coil(1)
			}).Should(PanicWith("invalid i: 1"))
		})
		It("can't set -1 coil", func() {
			Expect(func() {
				cmd.SetCoil(-1, true)
			}).Should(PanicWith("invalid i: -1"))
		})
		It("can't set too many coil", func() {
			Expect(func() {
				cmd.SetCoil(1, true)
			}).Should(PanicWith("invalid i: 1"))
		})

		Context("Coil 0 changed", func() {
			const tx = "0000 0<-WC  2:1[1]"
			ncoils := []bool{true}
			BeforeEach(func() {
				cmd.SetCoil(0, true)
			})

			OnlyTx(0, addr, tx, ncoils, []byte{
				0, 0, 0, 0, 0, 8,
				0, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b1,
			})
		})

		It("can't get -1 byte", func() {
			Expect(func() {
				cmd.Byte(-1)
			}).Should(PanicWith("invalid i: -1"))
		})
		It("can't get too many byte", func() {
			Expect(func() {
				cmd.Byte(1)
			}).Should(PanicWith("invalid i: 1"))
		})
		It("can't set -1 byte", func() {
			Expect(func() {
				cmd.SetByte(-1, 0b1)
			}).Should(PanicWith("invalid i: -1"))
		})
		It("can't set too many byte", func() {
			Expect(func() {
				cmd.SetByte(1, 0b1)
			}).Should(PanicWith("invalid i: 1"))
		})

		Context("Byte 0 changed", func() {
			const tx = "0000 0<-WC  2:1[1]"
			ncoils := []bool{true}
			BeforeEach(func() {
				cmd.SetByte(0, 0b1)
			})

			OnlyTx(0, addr, tx, ncoils, []byte{
				0, 0, 0, 0, 0, 8,
				0, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b1,
			})
		})

		Context("Bytes Modified", func() {
			const tx = "0000 0<-WC  2:1[1]"
			ncoils := []bool{true}
			BeforeEach(func() {
				cmd.ModifyBytes(func(b []byte) {
					copy(b, []byte{0b1})
				})
			})

			OnlyTx(0, addr, tx, ncoils, []byte{
				0, 0, 0, 0, 0, 8,
				0, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b1,
			})
		})
	})

	Context("one", func() {
		const dev byte = 13
		const addr uint16 = 22
		vals := []bool{true}
		const tx = "0000 13<-WC  22:1[1]"
		BeforeEach(func() {
			cmd = NewWriteCoilsCmd(dev, addr, vals)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 8,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b1,
			})
		})

		Context("Tx Id changed", func() {
			const id uint16 = 291
			const tx = "0123 13<-WC  22:1[1]"
			BeforeEach(func() {
				Expect(cmd.TxId()).To(BeZero())
				cmd.SetTxId(id)
			})

			OnlyTx(dev, addr, tx, vals, []byte{
				byte(id >> 8), byte(id & 0xff), 0, 0, 0, 8,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b1,
			})

			Context("Rx", func() {
				b := []byte{
					byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
					dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
				}
				const rx = "0123 13->WC  22:1"

				SetRx(b)
				GoodRx(b, tx, rx)
			})
		})

		Context("Dev Addr changed", func() {
			const ndev byte = 123
			const tx = "0000 123<-WC  22:1[1]"
			BeforeEach(func() {
				cmd.SetDevAddr(ndev)
			})

			OnlyTx(ndev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 8,
				ndev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b1,
			})
		})

		Context("Dev Addr changed to broadcast", func() {
			const tx = "0000 0<-WC  22:1[1]"
			BeforeEach(func() {
				cmd.SetDevAddr(0)
			})

			OnlyTx(0, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 8,
				0, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b1,
			})
		})

		Context("Addr changed", func() {
			const naddr uint16 = 122
			const tx = "0000 13<-WC  122:1[1]"
			BeforeEach(func() {
				cmd.SetAddr(naddr)
			})

			OnlyTx(dev, naddr, tx, vals, []byte{
				0, 0, 0, 0, 0, 8,
				dev, 15, byte(naddr >> 8), byte(naddr & 0xFF), 0, 1, 1, 0b1,
			})
		})

		Context("Coil 0 changed", func() {
			const tx = "0000 13<-WC  22:1[0]"
			ncoils := []bool{false}
			BeforeEach(func() {
				cmd.SetCoil(0, false)
			})

			OnlyTx(dev, addr, tx, ncoils, []byte{
				0, 0, 0, 0, 0, 8,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b0,
			})
		})

		Context("Byte 0 changed", func() {
			const tx = "0000 13<-WC  22:1[0]"
			ncoils := []bool{false}
			BeforeEach(func() {
				cmd.SetByte(0, 0b0)
			})

			OnlyTx(dev, addr, tx, ncoils, []byte{
				0, 0, 0, 0, 0, 8,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b0,
			})
		})

		Context("Bytes Modified", func() {
			const tx = "0000 13<-WC  22:1[0]"
			ncoils := []bool{false}
			BeforeEach(func() {
				cmd.ModifyBytes(func(b []byte) {
					copy(b, []byte{0b0})
				})
			})

			OnlyTx(dev, addr, tx, ncoils, []byte{
				0, 0, 0, 0, 0, 8,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 1, 0b0,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			}
			const rx = "0000 13->WC  22:1"

			SetRx(b)
			GoodRx(b, tx, rx)
		})

		Context("Err Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 3, dev, 0x80 | 15, 1}
			const rx = "0000 13->WC  Illegal Function"

			SetRx(b)
			IsValidRx()
			RxBytes(b)
			RxString(rx)
			String(tx + "\n" + rx)
			It("has error ", func() {
				Expect(cmd.Err()).To(Equal(IllegalFunction))
			})
		})
	})

	Context("five", func() {
		const dev byte = 123
		const addr uint16 = 32
		vals := []bool{false, true, true, false, true}
		const tx = "0000 123<-WC  32:5[0 1 1 0 1]"
		BeforeEach(func() {
			cmd = NewWriteCoilsCmd(dev, addr, vals)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 8,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 5, 1, 0b1_0110,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 5,
			}
			const rx = "0000 123->WC  32:5"

			SetRx(b)
			GoodRx(b, tx, rx)
		})
	})

	Context("six", func() {
		const dev byte = 234
		const addr uint16 = 567
		vals := []bool{true, false, true, true, false, true}
		const tx = "0000 234<-WC  567:6[1 0 1 1 0  1]"
		BeforeEach(func() {
			cmd = NewWriteCoilsCmd(dev, addr, vals)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 8,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 6, 1, 0b10_1101,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 6,
			}
			const rx = "0000 234->WC  567:6"

			SetRx(b)
			GoodRx(b, tx, rx)
		})
	})

	Context("ten", func() {
		const dev byte = 123
		const addr uint16 = 4567
		vals := []bool{
			true, false, true, true, false, true, true, false, true, true,
		}
		const tx = "0000 123<-WC  4567:10[1 0 1 1 0  1 1 0 1 1]"
		BeforeEach(func() {
			cmd = NewWriteCoilsCmd(dev, addr, vals)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 9,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 10,
				2, 0b0110_1101, 0b11,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 10,
			}
			const rx = "0000 123->WC  4567:10"

			SetRx(b)
			GoodRx(b, tx, rx)
		})
	})

	Context("eleven", func() {
		const dev byte = 111
		const addr uint16 = 56789
		const tx = `0000 111<-WC  56789:11[
 1 0 1 1 0  1 1 0 1 1
 0
]`
		vals := []bool{
			true, false, true, true, false, true, true, false, true, true,
			false,
		}
		BeforeEach(func() {
			cmd = NewWriteCoilsCmd(dev, addr, vals)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 9,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 11,
				2, 0b0110_1101, 0b011,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF), 0, 11,
			}
			const rx = "0000 111->WC  56789:11"

			SetRx(b)
			GoodRx(b, tx, rx)
		})
	})

	Context("in 1968", func() {
		const dev byte = 1
		const addr uint16 = 63568
		tx := `0000 1<-WC  63568:1968[` + strings.Repeat(`
 1 0 1 0 0  1 0 1 1 0
 1 0 0 1 0  1 1 0 1 0
 0 1 0 1 1  0 1 0 0 1
 0 1 1 0 1  0 0 1 0 1`, 49) + `
 1 0 1 0 0  1 0 1
]`
		vals := make([]bool, 1968)
		for i := 0; i < 246; i++ {
			vals[i*8+0] = true
			vals[i*8+1] = false
			vals[i*8+2] = true
			vals[i*8+3] = false
			vals[i*8+4] = false
			vals[i*8+5] = true
			vals[i*8+6] = false
			vals[i*8+7] = true
		}
		BeforeEach(func() {
			cmd = NewWriteCoilsCmd(dev, addr, vals)
		})

		Context("New", func() {
			b := make([]byte, 259)
			b[5] = 253
			b[6] = dev
			b[7] = 15
			b[8] = byte(addr >> 8)
			b[9] = byte(addr & 0xFF)
			b[10] = byte(1968 >> 8)
			b[11] = byte(1968 & 0xFF)
			b[12] = 246
			for i := 13; i < 259; i++ {
				b[i] = 0xA5
			}

			OnlyTx(dev, addr, tx, vals, b)
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 15, byte(addr >> 8), byte(addr & 0xFF),
				byte(1968 >> 8), byte(1968 & 0xFF),
			}
			const rx = `0000 1->WC  63568:1968`

			SetRx(b)
			GoodRx(b, tx, rx)
		})
	})
})

var _ = Describe("WriteRegsCmd", func() {
	var cmd *WriteRegsCmd
	SetRx := func(b []byte) {
		BeforeEach(func() {
			rx := cmd.RxBytes()
			*rx = (*rx)[:len(b)]
			copy(*rx, b)
		})
	}

	String := func(x string) {
		It("has String", func() {
			Expect(cmd.String()).To(Equal(x))
		})
	}
	Count := func(x int) {
		It("has Count", func() {
			Expect(cmd.Count()).To(Equal(x))
		})
	}
	OnlyTx := func(dev byte, addr uint16, s string, a []uint16, b []byte) {
		It("has Tx Bytes", func() {
			Expect(cmd.TxBytes()).To(Equal(b))
		})
		It("has Dev Addr", func() {
			Expect(cmd.DevAddr()).To(Equal(dev))
		})
		It("has Addr", func() {
			Expect(cmd.Addr()).To(Equal(addr))
		})
		It("has Tx String", func() {
			Expect(cmd.Tx()).To(Equal(s))
		})
		if dev == 0 {
			String(s)
		} else {
			String(s + "\n[]")
		}
		Count(len(a))
		It("has Regs", func() {
			for i, x := range a {
				Expect(cmd.Reg(i)).To(Equal(x), "Reg "+strconv.Itoa(i))
			}
		})
		It("has ByteCount", func() {
			Expect(cmd.ByteCount()).To(Equal(len(b) - 13))
		})
		It("has Byte", func() {
			for i := 0; i < len(b)-13; i++ {
				Expect(cmd.Byte(i)).To(Equal(b[13+i]), "Byte "+strconv.Itoa(i))
			}
		})
		It("has Bytes", func() {
			Expect(cmd.Bytes()).To(Equal(b[13:]))
		})
	}
	IsValidRx := func() {
		It("is Valid Rx", func() {
			Expect(cmd.IsValidRx()).To(BeTrue())
		})
	}
	RxBytes := func(x []byte) {
		It("has Rx Bytes", func() {
			Expect(*cmd.RxBytes()).To(Equal(x))
		})
	}
	RxString := func(x string) {
		It("has Rx String", func() {
			Expect(cmd.Rx()).To(Equal(x))
		})
	}
	GoodRx := func(rb []byte, tx, rx string) {
		IsValidRx()
		RxBytes(rb)
		RxString(rx)
		String(tx + "\n" + rx)
		It("has no Err", func() {
			Expect(cmd.Err()).To(Succeed())
		})
	}

	Describe("Invalid New", func() {
		It("can't write zero", func() {
			Expect(func() {
				NewWriteRegsCmd(1, 2, []uint16{})
			}).Should(PanicWith("empty values"))
		})
		It("can't write beyond 123", func() {
			Expect(func() {
				NewWriteRegsCmd(1, 2, make([]uint16, 124))
			}).Should(PanicWith("values too many: 124"))
		})
		It("can't overflow address", func() {
			Expect(func() {
				NewWriteRegsCmd(1, 65417, make([]uint16, 123))
			}).Should(PanicWith("address overflow: 65417, 123"))
		})
	})

	Context("broadcast", func() {
		const addr uint16 = 2
		const tx = "0000 0<-WR  2:1[    3]"
		vals := []uint16{3}
		BeforeEach(func() {
			cmd = NewWriteRegsCmd(0, addr, vals)
		})

		Context("New", func() {
			OnlyTx(0, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 9,
				0, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 0, 3,
			})
		})

		Context("Dev Addr changed", func() {
			const ndev byte = 13
			const tx = "0000 13<-WR  2:1[    3]"
			BeforeEach(func() {
				cmd.SetDevAddr(ndev)
			})

			OnlyTx(ndev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 9,
				13, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 0, 3,
			})
		})

		Context("Addr changed", func() {
			const naddr uint16 = 0
			const tx = "0000 0<-WR  0:1[    3]"
			BeforeEach(func() {
				cmd.SetAddr(naddr)
			})

			OnlyTx(0, naddr, tx, vals, []byte{
				0, 0, 0, 0, 0, 9,
				0, 16, byte(naddr >> 8), byte(naddr & 0xFF), 0, 1, 2, 0, 3,
			})
		})

		It("can't get -1 reg", func() {
			Expect(func() {
				cmd.Reg(-1)
			}).Should(PanicWith("invalid i: -1"))
		})
		It("can't get too many reg", func() {
			Expect(func() {
				cmd.Reg(1)
			}).Should(PanicWith("invalid i: 1"))
		})
		It("can't set -1 reg", func() {
			Expect(func() {
				cmd.SetReg(-1, 123)
			}).Should(PanicWith("invalid i: -1"))
		})
		It("can't set too many reg", func() {
			Expect(func() {
				cmd.SetReg(1, 456)
			}).Should(PanicWith("invalid i: 1"))
		})

		Context("Reg 0 changed", func() {
			const tx = "0000 0<-WR  2:1[ 4321]"
			nregs := []uint16{4321}
			BeforeEach(func() {
				cmd.SetReg(0, 4321)
			})

			OnlyTx(0, addr, tx, nregs, []byte{
				0, 0, 0, 0, 0, 9,
				0, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 16, 225,
			})
		})

		It("can't get -1 byte", func() {
			Expect(func() {
				cmd.Byte(-1)
			}).Should(PanicWith("invalid i: -1"))
		})
		It("can't get too many byte", func() {
			Expect(func() {
				cmd.Byte(2)
			}).Should(PanicWith("invalid i: 2"))
		})
		It("can't set -1 byte", func() {
			Expect(func() {
				cmd.SetByte(-1, 0b1)
			}).Should(PanicWith("invalid i: -1"))
		})
		It("can't set too many byte", func() {
			Expect(func() {
				cmd.SetByte(2, 3)
			}).Should(PanicWith("invalid i: 2"))
		})

		Context("Byte 0 changed", func() {
			const tx = "0000 0<-WR  2:1[  259]"
			nregs := []uint16{259}
			BeforeEach(func() {
				cmd.SetByte(0, 1)
			})

			OnlyTx(0, addr, tx, nregs, []byte{
				0, 0, 0, 0, 0, 9,
				0, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 1, 3,
			})
		})

		Context("Bytes Modified", func() {
			const tx = "0000 0<-WR  2:1[  259]"
			nregs := []uint16{259}
			BeforeEach(func() {
				cmd.ModifyBytes(func(b []byte) {
					copy(b, []byte{1, 3})
				})
			})

			OnlyTx(0, addr, tx, nregs, []byte{
				0, 0, 0, 0, 0, 9,
				0, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 1, 3,
			})
		})
	})

	Context("one", func() {
		const dev byte = 13
		const addr uint16 = 22
		vals := []uint16{44}
		const tx = "0000 13<-WR  22:1[   44]"
		BeforeEach(func() {
			cmd = NewWriteRegsCmd(dev, addr, vals)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 9,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 0, 44,
			})
		})

		Context("Tx Id changed", func() {
			const id uint16 = 43981
			const tx = "ABCD 13<-WR  22:1[   44]"
			BeforeEach(func() {
				Expect(cmd.TxId()).To(BeZero())
				cmd.SetTxId(id)
			})

			OnlyTx(dev, addr, tx, vals, []byte{
				byte(id >> 8), byte(id & 0xff), 0, 0, 0, 9,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 0, 44,
			})

			Context("Rx", func() {
				b := []byte{
					byte(id >> 8), byte(id & 0xff), 0, 0, 0, 6,
					dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
				}
				const rx = "ABCD 13->WR  22:1"

				SetRx(b)
				GoodRx(b, tx, rx)
			})
		})

		Context("Dev Addr changed", func() {
			const ndev byte = 123
			const tx = "0000 123<-WR  22:1[   44]"
			BeforeEach(func() {
				cmd.SetDevAddr(ndev)
			})

			OnlyTx(ndev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 9,
				ndev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 0, 44,
			})
		})

		Context("Dev Addr changed to broadcast", func() {
			const tx = "0000 0<-WR  22:1[   44]"
			BeforeEach(func() {
				cmd.SetDevAddr(0)
			})

			OnlyTx(0, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 9,
				0, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 0, 44,
			})
		})

		Context("Addr changed", func() {
			const naddr uint16 = 122
			const tx = "0000 13<-WR  122:1[   44]"
			BeforeEach(func() {
				cmd.SetAddr(naddr)
			})

			OnlyTx(dev, naddr, tx, vals, []byte{
				0, 0, 0, 0, 0, 9,
				dev, 16, byte(naddr >> 8), byte(naddr & 0xFF), 0, 1, 2, 0, 44,
			})
		})

		Context("Reg 0 changed", func() {
			const tx = "0000 13<-WR  22:1[   55]"
			nregs := []uint16{55}
			BeforeEach(func() {
				cmd.SetReg(0, 55)
			})

			OnlyTx(dev, addr, tx, nregs, []byte{
				0, 0, 0, 0, 0, 9,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 0, 55,
			})
		})

		Context("Byte 0 changed", func() {
			const tx = "0000 13<-WR  22:1[  300]"
			nregs := []uint16{300}
			BeforeEach(func() {
				cmd.SetByte(0, 1)
			})

			OnlyTx(dev, addr, tx, nregs, []byte{
				0, 0, 0, 0, 0, 9,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 1, 44,
			})
		})

		Context("Bytes Modified", func() {
			const tx = "0000 13<-WR  22:1[  300]"
			nregs := []uint16{300}
			BeforeEach(func() {
				cmd.ModifyBytes(func(b []byte) {
					copy(b, []byte{1})
				})
			})

			OnlyTx(dev, addr, tx, nregs, []byte{
				0, 0, 0, 0, 0, 9,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1, 2, 1, 44,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 1,
			}
			const rx = "0000 13->WR  22:1"

			SetRx(b)
			GoodRx(b, tx, rx)
		})

		Context("Err Rx", func() {
			b := []byte{0, 0, 0, 0, 0, 3, dev, 0x80 | 16, 1}
			const rx = "0000 13->WR  Illegal Function"

			SetRx(b)
			IsValidRx()
			RxBytes(b)
			RxString(rx)
			String(tx + "\n" + rx)
			It("has error ", func() {
				Expect(cmd.Err()).To(Equal(IllegalFunction))
			})
		})
	})

	Context("five", func() {
		const dev byte = 123
		const addr uint16 = 32
		vals := []uint16{44, 555, 6666, 71, 832}
		const tx = "0000 123<-WR  32:5[   44   555  6666    71   832]"
		BeforeEach(func() {
			cmd = NewWriteRegsCmd(dev, addr, vals)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 17,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 5, 10,
				0, 44, 2, 43, 26, 10, 0, 71, 3, 64,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 5,
			}
			const rx = "0000 123->WR  32:5"

			SetRx(b)
			GoodRx(b, tx, rx)
		})
	})

	Context("six", func() {
		const dev byte = 234
		const addr uint16 = 567
		vals := []uint16{11111, 2222, 333, 44, 5, 65432}
		const tx = "0000 234<-WR  567:6[11111  2222   333    44     5 : 65432]"
		BeforeEach(func() {
			cmd = NewWriteRegsCmd(dev, addr, vals)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 19,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 6, 12,
				43, 103, 8, 174, 1, 77, 0, 44, 0, 5, 255, 152,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 6,
			}
			const rx = "0000 234->WR  567:6"

			SetRx(b)
			GoodRx(b, tx, rx)
		})
	})

	Context("ten", func() {
		const dev byte = 123
		const addr uint16 = 4567
		vals := []uint16{
			11111, 2222, 333, 44, 5, 66, 777, 888, 9999, 10101,
		}
		const tx = "0000 123<-WR  4567:10[" +
			"11111  2222   333    44     5 :    66   777   888  9999 10101" +
			"]"
		BeforeEach(func() {
			cmd = NewWriteRegsCmd(dev, addr, vals)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 27,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 10, 20,
				43, 103, 8, 174, 1, 77, 0, 44, 0, 5,
				0, 66, 3, 9, 3, 120, 39, 15, 39, 117,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 10,
			}
			const rx = "0000 123->WR  4567:10"

			SetRx(b)
			GoodRx(b, tx, rx)
		})
	})

	Context("eleven", func() {
		const dev byte = 111
		const addr uint16 = 56789
		const tx = `0000 111<-WR  56789:11[
 11111  2222   333    44     5 :    66   777  8888   999 12345
 11011
]`
		vals := []uint16{
			11111, 2222, 333, 44, 5, 66, 777, 8888, 999, 12345,
			11011,
		}
		BeforeEach(func() {
			cmd = NewWriteRegsCmd(dev, addr, vals)
		})

		Context("New", func() {
			OnlyTx(dev, addr, tx, vals, []byte{
				0, 0, 0, 0, 0, 29,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 11, 22,
				43, 103, 8, 174, 1, 77, 0, 44, 0, 5,
				0, 66, 3, 9, 34, 184, 3, 231, 48, 57,
				43, 3,
			})
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 11,
			}
			const rx = "0000 111->WR  56789:11"

			SetRx(b)
			GoodRx(b, tx, rx)
		})
	})

	Context("in 123", func() {
		const dev byte = 1
		const addr uint16 = 65413
		tx := `0000 1<-WR  65413:123[
 10001 10002 10003 10004 10005 : 10006 10007 10008 10009 10010
 10011 10012 10013 10014 10015 : 10016 10017 10018 10019 10020
 10021 10022 10023 10024 10025 : 10026 10027 10028 10029 10030
 10031 10032 10033 10034 10035 : 10036 10037 10038 10039 10040
 10041 10042 10043 10044 10045 : 10046 10047 10048 10049 10050
 10051 10052 10053 10054 10055 : 10056 10057 10058 10059 10060
 10061 10062 10063 10064 10065 : 10066 10067 10068 10069 10070
 10071 10072 10073 10074 10075 : 10076 10077 10078 10079 10080
 10081 10082 10083 10084 10085 : 10086 10087 10088 10089 10090
 10091 10092 10093 10094 10095 : 10096 10097 10098 10099 10100
 10101 10102 10103 10104 10105 : 10106 10107 10108 10109 10110
 10111 10112 10113 10114 10115 : 10116 10117 10118 10119 10120
 10121 10122 10123
]`
		vals := make([]uint16, 123)
		for i := 0; i < 123; i++ {
			vals[i] = 10001 + uint16(i)
		}
		BeforeEach(func() {
			cmd = NewWriteRegsCmd(dev, addr, vals)
		})

		Context("New", func() {
			b := make([]byte, 259)
			b[5] = 253
			b[6] = dev
			b[7] = 16
			b[8] = byte(addr >> 8)
			b[9] = byte(addr & 0xFF)
			b[10] = 0
			b[11] = 123
			b[12] = 246
			for i := 0; i < 123; i++ {
				b[13+i*2] = byte((10001 + i) >> 8)
				b[14+i*2] = byte(10001 + i)
			}

			OnlyTx(dev, addr, tx, vals, b)
		})

		Context("Valid Rx", func() {
			b := []byte{
				0, 0, 0, 0, 0, 6,
				dev, 16, byte(addr >> 8), byte(addr & 0xFF), 0, 123,
			}
			const rx = `0000 1->WR  65413:123`

			SetRx(b)
			GoodRx(b, tx, rx)
		})
	})
})
