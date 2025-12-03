package modbus_test

import (
	"strconv"
	"strings"
	"testing"

	. "github.com/bangzek/modbus-tcp"
)

var result string

func BenchmarkReadCoilsCmd(b *testing.B) {
	sda := func(c *ReadCoilsCmd, a byte) {
		c.SetDevAddr(a)
		b := c.RxBytes()
		(*b)[6] = a
	}
	srx := func(c *ReadCoilsCmd, b []byte) {
		r := c.RxBytes()
		*r = (*r)[:len(b)]
		copy(*r, b)
	}

	var cmd *ReadCoilsCmd

	rx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Rx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	tx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Tx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	str := func(x string) func(b *testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.String()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}

	runAll := func(x [3][5][3]string) {
		devA := cmd.DevAddr()
		addr := cmd.Addr()
		p := strconv.Itoa(cmd.Count())
		if len(*cmd.RxBytes()) == 5 {
			p += ",ERR"
		}

		ds := [3]byte{devA, 20 + devA, 120 + devA}
		as := [5]uint16{0, 10, 100, 1000, 10000}
		for i := 0; i < 3; i++ {
			sda(cmd, ds[i])
			si := strconv.Itoa(i + 1)
			for j := 0; j < 5; j++ {
				cmd.SetAddr(as[j] + addr)
				sj := strconv.Itoa(j + 1)
				b.Run(" tx:"+si+","+sj+","+p, tx(x[i][j][0]))
				b.Run(" rx:"+si+","+sj+","+p, rx(x[i][j][1]))
				b.Run("str:"+si+","+sj+","+p, str(x[i][j][2]))
			}
		}
	}
	msg := func(r string) (s [3][5][3]string) {
		d := strconv.Itoa(int(cmd.DevAddr()))
		a := strconv.Itoa(int(cmd.Addr()))
		c := strconv.Itoa(cmd.Count())
		ds := [3]string{d, "2" + d, "12" + d}
		as := [5]string{"", "1", "10", "100", "1000"}
		for i := 0; i < 3; i++ {
			for j := 0; j < 5; j++ {
				s[i][j][0] = "0000 " + ds[i] + "<-RC  " + as[j] + a + ":" + c
				s[i][j][1] = "0000 " + ds[i] + "->RC  " + r
				s[i][j][2] = s[i][j][0] + "\n" + s[i][j][1]
			}
		}
		return
	}

	const devA = 3
	const addr = 2
	cmd = NewReadCoilsCmd(devA, addr, 1)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x81, 1})
	runAll(msg("Illegal Function"))

	cmd = NewReadCoilsCmd(devA, addr, 5)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x81, 2})
	runAll(msg("Illegal Data Address"))

	cmd = NewReadCoilsCmd(devA, addr, 6)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x81, 3})
	runAll(msg("Illegal Data Value"))

	cmd = NewReadCoilsCmd(devA, addr, 10)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x81, 4})
	runAll(msg("Slave Device Failure"))

	//

	cmd = NewReadCoilsCmd(devA, addr, 1)
	srx(cmd, []byte{0, 0, 0, 0, 0, 4, devA, 1, 1, 0b1})
	runAll(msg("1[1]"))

	cmd = NewReadCoilsCmd(devA, addr, 5)
	srx(cmd, []byte{0, 0, 0, 0, 0, 4, devA, 1, 1, 0b0_1001})
	runAll(msg("5[1 0 0 1 0]"))

	cmd = NewReadCoilsCmd(devA, addr, 6)
	srx(cmd, []byte{0, 0, 0, 0, 0, 4, devA, 1, 1, 0b10_1101})
	runAll(msg("6[1 0 1 1 0  1]"))

	cmd = NewReadCoilsCmd(devA, addr, 10)
	srx(cmd, []byte{0, 0, 0, 0, 0, 5, devA, 1, 2, 0b0110_1101, 0b11})
	runAll(msg("10[1 0 1 1 0  1 1 0 1 1]"))

	cmd = NewReadCoilsCmd(devA, addr, 11)
	srx(cmd, []byte{0, 0, 0, 0, 0, 5, devA, 1, 2, 0b0110_1101, 0b011})
	runAll(msg(`11[
 1 0 1 1 0  1 1 0 1 1
 0
]`))

	cmd = NewReadCoilsCmd(devA, addr, 15)
	srx(cmd, []byte{0, 0, 0, 0, 0, 5, devA, 1, 2, 0b1001_0010, 0b010_0100})
	runAll(msg(`15[
 0 1 0 0 1  0 0 1 0 0
 1 0 0 1 0
]`))

	cmd = NewReadCoilsCmd(devA, addr, 16)
	srx(cmd, []byte{0, 0, 0, 0, 0, 5, devA, 1, 2, 0b0110_1101, 0b1101_1011})
	runAll(msg(`16[
 1 0 1 1 0  1 1 0 1 1
 0 1 1 0 1  1
]`))

	cmd = NewReadCoilsCmd(devA, addr, 20)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 6, devA, 1, 3, 0b1001_0010, 0b0010_0100, 0b1001,
	})
	runAll(msg(`20[
 0 1 0 0 1  0 0 1 0 0
 1 0 0 1 0  0 1 0 0 1
]`))

	cmd = NewReadCoilsCmd(devA, addr, 21)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 6, devA, 1, 3, 0b0110_1101, 0b1101_1011, 0b1_0110,
	})
	runAll(msg(`21[
 1 0 1 1 0  1 1 0 1 1
 0 1 1 0 1  1 0 1 1 0
 1
]`))

	cmd = NewReadCoilsCmd(devA, addr, 2000)
	brx := make([]byte, 259)
	brx[5] = 253
	brx[6] = devA
	brx[7] = 1
	brx[8] = 250
	for i := 9; i < 259; i++ {
		brx[i] = 0xA5
	}
	srx(cmd, brx)
	runAll(msg(`2000[` + strings.Repeat(`
 1 0 1 0 0  1 0 1 1 0
 1 0 0 1 0  1 1 0 1 0
 0 1 0 1 1  0 1 0 0 1
 0 1 1 0 1  0 0 1 0 1`, 50) + `
]`))
}

func BenchmarkReadDInputsCmd(b *testing.B) {
	sda := func(c *ReadDInputsCmd, a byte) {
		c.SetDevAddr(a)
		b := c.RxBytes()
		(*b)[6] = a
	}
	srx := func(c *ReadDInputsCmd, b []byte) {
		r := c.RxBytes()
		*r = (*r)[:len(b)]
		copy(*r, b)
	}

	var cmd *ReadDInputsCmd

	rx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Rx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	tx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Tx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	str := func(x string) func(b *testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.String()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}

	runAll := func(x [3][5][3]string) {
		devA := cmd.DevAddr()
		addr := cmd.Addr()
		p := strconv.Itoa(cmd.Count())
		if len(*cmd.RxBytes()) == 5 {
			p += ",ERR"
		}

		ds := [3]byte{devA, 20 + devA, 120 + devA}
		as := [5]uint16{0, 10, 100, 1000, 10000}
		for i := 0; i < 3; i++ {
			sda(cmd, ds[i])
			si := strconv.Itoa(i + 1)
			for j := 0; j < 5; j++ {
				cmd.SetAddr(as[j] + addr)
				sj := strconv.Itoa(j + 1)
				b.Run(" tx:"+si+","+sj+","+p, tx(x[i][j][0]))
				b.Run(" rx:"+si+","+sj+","+p, rx(x[i][j][1]))
				b.Run("str:"+si+","+sj+","+p, str(x[i][j][2]))
			}
		}
	}
	msg := func(r string) (s [3][5][3]string) {
		d := strconv.Itoa(int(cmd.DevAddr()))
		a := strconv.Itoa(int(cmd.Addr()))
		c := strconv.Itoa(cmd.Count())
		ds := [3]string{d, "2" + d, "12" + d}
		as := [5]string{"", "1", "10", "100", "1000"}
		for i := 0; i < 3; i++ {
			for j := 0; j < 5; j++ {
				s[i][j][0] = "0000 " + ds[i] + "<-RDI " + as[j] + a + ":" + c
				s[i][j][1] = "0000 " + ds[i] + "->RDI " + r
				s[i][j][2] = s[i][j][0] + "\n" + s[i][j][1]
			}
		}
		return
	}

	const devA = 3
	const addr = 2

	cmd = NewReadDInputsCmd(devA, addr, 1)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x82, 1})
	runAll(msg("Illegal Function"))

	cmd = NewReadDInputsCmd(devA, addr, 5)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x82, 2})
	runAll(msg("Illegal Data Address"))

	cmd = NewReadDInputsCmd(devA, addr, 6)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x82, 3})
	runAll(msg("Illegal Data Value"))

	cmd = NewReadDInputsCmd(devA, addr, 10)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x82, 4})
	runAll(msg("Slave Device Failure"))

	//

	cmd = NewReadDInputsCmd(devA, addr, 1)
	srx(cmd, []byte{0, 0, 0, 0, 0, 4, devA, 2, 1, 0b1})
	runAll(msg("1[1]"))

	cmd = NewReadDInputsCmd(devA, addr, 5)
	srx(cmd, []byte{0, 0, 0, 0, 0, 4, devA, 2, 1, 0b0_1001})
	runAll(msg("5[1 0 0 1 0]"))

	cmd = NewReadDInputsCmd(devA, addr, 6)
	srx(cmd, []byte{0, 0, 0, 0, 0, 4, devA, 2, 1, 0b10_1101})
	runAll(msg("6[1 0 1 1 0  1]"))

	cmd = NewReadDInputsCmd(devA, addr, 10)
	srx(cmd, []byte{0, 0, 0, 0, 0, 5, devA, 2, 2, 0b0110_1101, 0b11})
	runAll(msg("10[1 0 1 1 0  1 1 0 1 1]"))

	cmd = NewReadDInputsCmd(devA, addr, 11)
	srx(cmd, []byte{0, 0, 0, 0, 0, 5, devA, 2, 2, 0b0110_1101, 0b011})
	runAll(msg(`11[
 1 0 1 1 0  1 1 0 1 1
 0
]`))

	cmd = NewReadDInputsCmd(devA, addr, 15)
	srx(cmd, []byte{0, 0, 0, 0, 0, 5, devA, 2, 2, 0b1001_0010, 0b010_0100})
	runAll(msg(`15[
 0 1 0 0 1  0 0 1 0 0
 1 0 0 1 0
]`))

	cmd = NewReadDInputsCmd(devA, addr, 16)
	srx(cmd, []byte{0, 0, 0, 0, 0, 5, devA, 2, 2, 0b0110_1101, 0b1101_1011})
	runAll(msg(`16[
 1 0 1 1 0  1 1 0 1 1
 0 1 1 0 1  1
]`))

	cmd = NewReadDInputsCmd(devA, addr, 20)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 6, devA, 2, 3, 0b1001_0010, 0b0010_0100, 0b1001,
	})
	runAll(msg(`20[
 0 1 0 0 1  0 0 1 0 0
 1 0 0 1 0  0 1 0 0 1
]`))

	cmd = NewReadDInputsCmd(devA, addr, 21)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 6, devA, 2, 3, 0b0110_1101, 0b1101_1011, 0b1_0110,
	})
	runAll(msg(`21[
 1 0 1 1 0  1 1 0 1 1
 0 1 1 0 1  1 0 1 1 0
 1
]`))

	cmd = NewReadDInputsCmd(devA, addr, 2000)
	brx := make([]byte, 259)
	brx[5] = 253
	brx[6] = devA
	brx[7] = 2
	brx[8] = 250
	for i := 9; i < 259; i++ {
		brx[i] = 0xA5
	}
	srx(cmd, brx)
	runAll(msg(`2000[` + strings.Repeat(`
 1 0 1 0 0  1 0 1 1 0
 1 0 0 1 0  1 1 0 1 0
 0 1 0 1 1  0 1 0 0 1
 0 1 1 0 1  0 0 1 0 1`, 50) + `
]`))
}

func BenchmarkReadHRegsCmd(b *testing.B) {
	sda := func(c *ReadHRegsCmd, a byte) {
		c.SetDevAddr(a)
		b := c.RxBytes()
		(*b)[6] = a
	}
	srx := func(c *ReadHRegsCmd, b []byte) {
		r := c.RxBytes()
		*r = (*r)[:len(b)]
		copy(*r, b)
	}

	var cmd *ReadHRegsCmd

	rx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Rx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	tx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Tx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	str := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.String()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}

	runAll := func(x [3][5][3]string) {
		devA := cmd.DevAddr()
		addr := cmd.Addr()
		p := strconv.Itoa(cmd.Count())
		if len(*cmd.RxBytes()) == 5 {
			p += ",ERR"
		}

		ds := [3]byte{devA, 20 + devA, 120 + devA}
		as := [5]uint16{0, 10, 100, 1000, 10000}
		for i := 0; i < 3; i++ {
			sda(cmd, ds[i])
			si := strconv.Itoa(i + 1)
			for j := 0; j < 5; j++ {
				cmd.SetAddr(as[j] + addr)
				sj := strconv.Itoa(j + 1)
				b.Run(" tx:"+si+","+sj+","+p, tx(x[i][j][0]))
				b.Run(" rx:"+si+","+sj+","+p, rx(x[i][j][1]))
				b.Run("str:"+si+","+sj+","+p, str(x[i][j][2]))
			}
		}
	}

	msg := func(r string) (s [3][5][3]string) {
		d := strconv.Itoa(int(cmd.DevAddr()))
		a := strconv.Itoa(int(cmd.Addr()))
		c := strconv.Itoa(cmd.Count())
		ds := [3]string{d, "2" + d, "12" + d}
		as := [5]string{"", "1", "10", "100", "1000"}
		for i := 0; i < 3; i++ {
			for j := 0; j < 5; j++ {
				s[i][j][0] = "0000 " + ds[i] + "<-RHR " + as[j] + a + ":" + c
				s[i][j][1] = "0000 " + ds[i] + "->RHR " + r
				s[i][j][2] = s[i][j][0] + "\n" + s[i][j][1]
			}
		}
		return
	}

	const devA = 3
	const addr = 2

	cmd = NewReadHRegsCmd(devA, addr, 1)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x83, 1})
	runAll(msg("Illegal Function"))

	cmd = NewReadHRegsCmd(devA, addr, 5)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x83, 2})
	runAll(msg("Illegal Data Address"))

	cmd = NewReadHRegsCmd(devA, addr, 6)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x83, 3})
	runAll(msg("Illegal Data Value"))

	cmd = NewReadHRegsCmd(devA, addr, 10)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x83, 4})
	runAll(msg("Slave Device Failure"))

	//

	cmd = NewReadHRegsCmd(devA, addr, 1)
	srx(cmd, []byte{0, 0, 0, 0, 0, 5, devA, 3, 2, 0xDE, 0xAD})
	runAll(msg("1[57005]"))

	cmd = NewReadHRegsCmd(devA, addr, 5)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 13,
		devA, 3, 10,
		0, 0, 255, 255, 0, 255, 255, 254, 1, 0,
	})
	runAll(msg("5[    0 65535   255 65534   256]"))

	cmd = NewReadHRegsCmd(devA, addr, 6)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 15,
		devA, 3, 12,
		0, 1, 0, 2, 0, 3, 1, 1, 1, 2, 1, 3,
	})
	runAll(msg("6[    1     2     3   257   258 :   259]"))

	cmd = NewReadHRegsCmd(devA, addr, 10)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 23,
		devA, 3, 20,
		0, 51, 53, 92, 153, 155, 169, 187, 195, 223,
		3, 65, 71, 75, 80, 86, 96, 171, 213, 248,
	})
	runAll(msg(
		"10[   51 13660 39323 43451 50143 :   833 18251 20566 24747 54776]",
	))

	cmd = NewReadHRegsCmd(devA, addr, 11)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 25,
		devA, 3, 22,
		0, 2, 2, 64, 3, 136, 1, 187, 231, 74,
		0, 6, 15, 128, 0, 48, 212, 110, 13, 212,
		9, 61,
	})
	runAll(msg(`11[
     2   576   904   443 59210 :     6  3968    48 54382  3540
  2365
]`))

	cmd = NewReadHRegsCmd(devA, addr, 15)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 33,
		devA, 3, 30,
		18, 234, 0, 79, 3, 185, 9, 179, 50, 230,
		0, 15, 0, 5, 0, 6, 19, 140, 0, 12,
		202, 143, 0, 5, 0, 2, 23, 197, 20, 102,
	})
	runAll(msg(`15[
  4842    79   953  2483 13030 :    15     5     6  5004    12
 51855     5     2  6085  5222
]`))

	cmd = NewReadHRegsCmd(devA, addr, 16)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 35,
		devA, 3, 32,
		0, 162, 0, 4, 0, 9, 20, 127, 0, 37,
		21, 170, 1, 85, 0, 5, 193, 177, 179, 139,
		3, 96, 0, 2, 1, 67, 8, 143, 0, 75,
		146, 122,
	})
	runAll(msg(`16[
   162     4     9  5247    37 :  5546   341     5 49585 45963
   864     2   323  2191    75 : 37498
]`))

	cmd = NewReadHRegsCmd(devA, addr, 20)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 43,
		devA, 3, 40,
		116, 142, 199, 122, 31, 90, 3, 163, 3, 144,
		2, 253, 2, 96, 105, 160, 0, 183, 193, 74,
		0, 23, 79, 47, 0, 238, 0, 18, 182, 19,
		0, 97, 0, 9, 0, 8, 0, 7, 3, 23,
	})
	runAll(msg(`20[
 29838 51066  8026   931   912 :   765   608 27040   183 49482
    23 20271   238    18 46611 :    97     9     8     7   791
]`))

	cmd = NewReadHRegsCmd(devA, addr, 21)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 45,
		devA, 3, 42,
		0, 74, 34, 43, 154, 47, 0, 82, 0, 33,
		243, 61, 0, 45, 201, 186, 19, 130, 0, 1,
		0, 64, 19, 111, 0, 5, 0, 24, 0, 32,
		0, 3, 33, 159, 0, 11, 115, 73, 0, 1,
		5, 178,
	})
	runAll(msg(`21[
    74  8747 39471    82    33 : 62269    45 51642  4994     1
    64  4975     5    24    32 :     3  8607    11 29513     1
  1458
]`))

	cmd = NewReadHRegsCmd(devA, addr, 125)
	brx := make([]byte, 259)
	brx[5] = 253
	brx[6] = devA
	brx[7] = 3
	brx[8] = 250
	for i := 9; i < 259; i += 10 {
		brx[i] = 0
		brx[i+1] = 7
		brx[i+2] = 0
		brx[i+3] = 87
		brx[i+4] = 0
		brx[i+5] = 233
		brx[i+6] = 5
		brx[i+7] = 205
		brx[i+8] = 220
		brx[i+9] = 203
	}
	srx(cmd, brx)
	runAll(msg(`125[` + strings.Repeat(`
     7    87   233  1485 56523 :     7    87   233  1485 56523`, 12) + `
     7    87   233  1485 56523
]`))
}

func BenchmarkReadIRegsCmd(b *testing.B) {
	sda := func(c *ReadIRegsCmd, a byte) {
		c.SetDevAddr(a)
		b := c.RxBytes()
		(*b)[6] = a
	}
	srx := func(c *ReadIRegsCmd, b []byte) {
		r := c.RxBytes()
		*r = (*r)[:len(b)]
		copy(*r, b)
	}

	var cmd *ReadIRegsCmd

	rx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Rx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	tx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Tx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	str := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.String()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}

	runAll := func(x [3][5][3]string) {
		devA := cmd.DevAddr()
		addr := cmd.Addr()
		p := strconv.Itoa(cmd.Count())
		if len(*cmd.RxBytes()) == 5 {
			p += ",ERR"
		}

		ds := [3]byte{devA, 20 + devA, 120 + devA}
		as := [5]uint16{0, 10, 100, 1000, 10000}
		for i := 0; i < 3; i++ {
			sda(cmd, ds[i])
			si := strconv.Itoa(i + 1)
			for j := 0; j < 5; j++ {
				cmd.SetAddr(as[j] + addr)
				sj := strconv.Itoa(j + 1)
				b.Run(" tx:"+si+","+sj+","+p, tx(x[i][j][0]))
				b.Run(" rx:"+si+","+sj+","+p, rx(x[i][j][1]))
				b.Run("str:"+si+","+sj+","+p, str(x[i][j][2]))
			}
		}
	}

	msg := func(r string) (s [3][5][3]string) {
		d := strconv.Itoa(int(cmd.DevAddr()))
		a := strconv.Itoa(int(cmd.Addr()))
		c := strconv.Itoa(cmd.Count())
		ds := [3]string{d, "2" + d, "12" + d}
		as := [5]string{"", "1", "10", "100", "1000"}
		for i := 0; i < 3; i++ {
			for j := 0; j < 5; j++ {
				s[i][j][0] = "0000 " + ds[i] + "<-RIR " + as[j] + a + ":" + c
				s[i][j][1] = "0000 " + ds[i] + "->RIR " + r
				s[i][j][2] = s[i][j][0] + "\n" + s[i][j][1]
			}
		}
		return
	}

	const devA = 3
	const addr = 2

	cmd = NewReadIRegsCmd(devA, addr, 1)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x84, 1})
	runAll(msg("Illegal Function"))

	cmd = NewReadIRegsCmd(devA, addr, 5)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x84, 2})
	runAll(msg("Illegal Data Address"))

	cmd = NewReadIRegsCmd(devA, addr, 6)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x84, 3})
	runAll(msg("Illegal Data Value"))

	cmd = NewReadIRegsCmd(devA, addr, 10)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x84, 4})
	runAll(msg("Slave Device Failure"))

	//

	cmd = NewReadIRegsCmd(devA, addr, 1)
	srx(cmd, []byte{0, 0, 0, 0, 0, 5, devA, 4, 2, 0xDE, 0xAD})
	runAll(msg("1[57005]"))

	cmd = NewReadIRegsCmd(devA, addr, 5)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 13,
		devA, 4, 10,
		0, 0, 255, 255, 0, 255, 255, 254, 1, 0,
	})
	runAll(msg("5[    0 65535   255 65534   256]"))

	cmd = NewReadIRegsCmd(devA, addr, 6)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 15,
		devA, 4, 12,
		0, 1, 0, 2, 0, 3, 1, 1, 1, 2, 1, 3,
	})
	runAll(msg("6[    1     2     3   257   258 :   259]"))

	cmd = NewReadIRegsCmd(devA, addr, 10)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 23,
		devA, 4, 20,
		0, 51, 53, 92, 153, 155, 169, 187, 195, 223,
		3, 65, 71, 75, 80, 86, 96, 171, 213, 248,
	})
	runAll(msg(
		"10[   51 13660 39323 43451 50143 :   833 18251 20566 24747 54776]",
	))

	cmd = NewReadIRegsCmd(devA, addr, 11)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 25,
		devA, 4, 22,
		0, 2, 2, 64, 3, 136, 1, 187, 231, 74,
		0, 6, 15, 128, 0, 48, 212, 110, 13, 212,
		9, 61,
	})
	runAll(msg(`11[
     2   576   904   443 59210 :     6  3968    48 54382  3540
  2365
]`))

	cmd = NewReadIRegsCmd(devA, addr, 15)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 33,
		devA, 4, 30,
		18, 234, 0, 79, 3, 185, 9, 179, 50, 230,
		0, 15, 0, 5, 0, 6, 19, 140, 0, 12,
		202, 143, 0, 5, 0, 2, 23, 197, 20, 102,
	})
	runAll(msg(`15[
  4842    79   953  2483 13030 :    15     5     6  5004    12
 51855     5     2  6085  5222
]`))

	cmd = NewReadIRegsCmd(devA, addr, 16)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 35,
		devA, 4, 32,
		0, 162, 0, 4, 0, 9, 20, 127, 0, 37,
		21, 170, 1, 85, 0, 5, 193, 177, 179, 139,
		3, 96, 0, 2, 1, 67, 8, 143, 0, 75,
		146, 122,
	})
	runAll(msg(`16[
   162     4     9  5247    37 :  5546   341     5 49585 45963
   864     2   323  2191    75 : 37498
]`))

	cmd = NewReadIRegsCmd(devA, addr, 20)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 43,
		devA, 4, 40,
		116, 142, 199, 122, 31, 90, 3, 163, 3, 144,
		2, 253, 2, 96, 105, 160, 0, 183, 193, 74,
		0, 23, 79, 47, 0, 238, 0, 18, 182, 19,
		0, 97, 0, 9, 0, 8, 0, 7, 3, 23,
	})
	runAll(msg(`20[
 29838 51066  8026   931   912 :   765   608 27040   183 49482
    23 20271   238    18 46611 :    97     9     8     7   791
]`))

	cmd = NewReadIRegsCmd(devA, addr, 21)
	srx(cmd, []byte{
		0, 0, 0, 0, 0, 45,
		devA, 4, 42,
		0, 74, 34, 43, 154, 47, 0, 82, 0, 33,
		243, 61, 0, 45, 201, 186, 19, 130, 0, 1,
		0, 64, 19, 111, 0, 5, 0, 24, 0, 32,
		0, 3, 33, 159, 0, 11, 115, 73, 0, 1,
		5, 178,
	})
	runAll(msg(`21[
    74  8747 39471    82    33 : 62269    45 51642  4994     1
    64  4975     5    24    32 :     3  8607    11 29513     1
  1458
]`))

	cmd = NewReadIRegsCmd(devA, addr, 125)
	brx := make([]byte, 259)
	brx[5] = 253
	brx[6] = devA
	brx[7] = 4
	brx[8] = 250
	for i := 9; i < 259; i += 10 {
		brx[i] = 0
		brx[i+1] = 7
		brx[i+2] = 0
		brx[i+3] = 87
		brx[i+4] = 0
		brx[i+5] = 233
		brx[i+6] = 5
		brx[i+7] = 205
		brx[i+8] = 220
		brx[i+9] = 203
	}
	srx(cmd, brx)
	runAll(msg(`125[` + strings.Repeat(`
     7    87   233  1485 56523 :     7    87   233  1485 56523`, 12) + `
     7    87   233  1485 56523
]`))
}

func BenchmarkWriteCoilCmd(b *testing.B) {
	sda := func(c *WriteCoilCmd, a byte) {
		c.SetDevAddr(a)
		b := c.RxBytes()
		(*b)[6] = a
	}
	srx := func(c *WriteCoilCmd, b []byte) {
		r := c.RxBytes()
		*r = (*r)[:len(b)]
		copy(*r, b)
	}
	sa := func(c *WriteCoilCmd, a uint16) {
		c.SetAddr(a)
		if r := c.RxBytes(); len(*r) != 9 {
			t := c.TxBytes()
			*r = (*r)[:12]
			copy(*r, t)
		}
	}

	var cmd *WriteCoilCmd

	rx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Rx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	tx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Tx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	str := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.String()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}

	trunAll := func(x [5]string) {
		s := "f"
		if cmd.Coil() {
			s = "t"
		}
		addr := cmd.Addr()
		as := [5]uint16{0, 10, 100, 1000, 10000}
		for i := 0; i < 5; i++ {
			si := strconv.Itoa(i + 1)
			cmd.SetAddr(as[i] + addr)
			b.Run(" tx:0,"+si+","+s, tx(x[i]))
			b.Run("str:0,"+si+","+s, str(x[i]))
		}
	}
	tmsg := func() [5]string {
		a := strconv.Itoa(int(cmd.Addr()))
		c := " false"
		if cmd.Coil() {
			c = " true"
		}
		return [5]string{
			"0000 0<-W1C " + a + c,
			"0000 0<-W1C 1" + a + c,
			"0000 0<-W1C 10" + a + c,
			"0000 0<-W1C 100" + a + c,
			"0000 0<-W1C 1000" + a + c,
		}
	}
	runAll := func(x [3][5][3]string) {
		devA := cmd.DevAddr()
		addr := cmd.Addr()
		s := "f"
		if cmd.Coil() {
			s = "t"
		}
		if rx := cmd.RxBytes(); len(*rx) == 5 {
			s += ",ERR:" + strconv.Itoa(int((*rx)[2]))
		}

		ds := [3]byte{devA, 20 + devA, 120 + devA}
		as := [5]uint16{0, 10, 100, 1000, 10000}
		for i := 0; i < 3; i++ {
			sda(cmd, ds[i])
			si := strconv.Itoa(i + 1)
			for j := 0; j < 5; j++ {
				sa(cmd, as[j]+addr)
				sj := strconv.Itoa(j + 1)
				b.Run(" tx:"+si+","+sj+","+s, tx(x[i][j][0]))
				b.Run(" rx:"+si+","+sj+","+s, rx(x[i][j][1]))
				b.Run("str:"+si+","+sj+","+s, str(x[i][j][2]))
			}
		}
	}
	msg := func(r string) (s [3][5][3]string) {
		d := strconv.Itoa(int(cmd.DevAddr()))
		a := strconv.Itoa(int(cmd.Addr()))
		c := " false"
		if cmd.Coil() {
			c = " true"
		}
		ds := [3]string{d, "2" + d, "12" + d}
		as := [5]string{"", "1", "10", "100", "1000"}
		for i := 0; i < 3; i++ {
			for j := 0; j < 5; j++ {
				s[i][j][0] = "0000 " + ds[i] + "<-W1C " + as[j] + a + c
				if r == "" {
					s[i][j][1] = "0000 " + ds[i] + "->W1C " + as[j] + a + c
				} else {
					s[i][j][1] = "0000 " + ds[i] + "->W1C " + r
				}
				s[i][j][2] = s[i][j][0] + "\n" + s[i][j][1]
			}
		}
		return
	}

	const addr = 2

	cmd = NewWriteCoilCmd(0, addr, true)
	trunAll(tmsg())

	cmd = NewWriteCoilCmd(0, addr, false)
	trunAll(tmsg())

	const devA = 3

	cmd = NewWriteCoilCmd(devA, addr, true)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x85, 1})
	runAll(msg("Illegal Function"))

	cmd = NewWriteCoilCmd(devA, addr, false)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x85, 1})
	runAll(msg("Illegal Function"))

	cmd = NewWriteCoilCmd(devA, addr, true)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x85, 2})
	runAll(msg("Illegal Data Address"))

	cmd = NewWriteCoilCmd(devA, addr, false)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x85, 2})
	runAll(msg("Illegal Data Address"))

	cmd = NewWriteCoilCmd(devA, addr, true)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x85, 3})
	runAll(msg("Illegal Data Value"))

	cmd = NewWriteCoilCmd(devA, addr, false)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x85, 3})
	runAll(msg("Illegal Data Value"))

	cmd = NewWriteCoilCmd(devA, addr, true)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x85, 4})
	runAll(msg("Slave Device Failure"))

	cmd = NewWriteCoilCmd(devA, addr, false)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x85, 4})
	runAll(msg("Slave Device Failure"))

	//

	cmd = NewWriteCoilCmd(devA, addr, true)
	srx(cmd, cmd.TxBytes())
	runAll(msg(""))

	cmd = NewWriteCoilCmd(devA, addr, false)
	srx(cmd, cmd.TxBytes())
	runAll(msg(""))
}

func BenchmarkWriteRegCmd(b *testing.B) {
	sda := func(c *WriteRegCmd, a byte) {
		c.SetDevAddr(a)
		b := c.RxBytes()
		(*b)[6] = a
	}
	srx := func(c *WriteRegCmd, b []byte) {
		r := c.RxBytes()
		*r = (*r)[:len(b)]
		copy(*r, b)
	}
	sa := func(c *WriteRegCmd, a uint16) {
		c.SetAddr(a)
		if r := c.RxBytes(); len(*r) != 9 {
			t := c.TxBytes()
			*r = (*r)[:12]
			copy(*r, t)
		}
	}

	var cmd *WriteRegCmd

	rx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Rx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	tx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Tx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	str := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.String()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}

	trunAll := func(x [5]string) {
		addr := cmd.Addr()
		as := [5]uint16{0, 10, 100, 1000, 10000}
		for i := 0; i < 5; i++ {
			si := strconv.Itoa(i + 1)
			cmd.SetAddr(as[i] + addr)
			b.Run(" tx:0,"+si, tx(x[i]))
			b.Run("str:0,"+si, str(x[i]))
		}
	}

	runAll := func(x [3][5][3]string) {
		devA := cmd.DevAddr()
		addr := cmd.Addr()
		var s string
		if rx := cmd.RxBytes(); len(*rx) == 5 {
			s = ",ERR:" + strconv.Itoa(int((*rx)[2]))
		}

		ds := [3]byte{devA, 20 + devA, 120 + devA}
		as := [5]uint16{0, 10, 100, 1000, 10000}
		for i := 0; i < 3; i++ {
			sda(cmd, ds[i])
			si := strconv.Itoa(i + 1)
			for j := 0; j < 5; j++ {
				sa(cmd, as[j]+addr)
				sj := strconv.Itoa(j + 1)
				b.Run(" tx:"+si+","+sj+s, tx(x[i][j][0]))
				b.Run(" rx:"+si+","+sj+s, rx(x[i][j][1]))
				b.Run("str:"+si+","+sj+s, str(x[i][j][2]))
			}
		}
	}

	msg := func(rx string) (s [3][5][3]string) {
		d := strconv.Itoa(int(cmd.DevAddr()))
		a := strconv.Itoa(int(cmd.Addr()))
		r := strconv.Itoa(int(cmd.Reg()))

		ds := [3]string{d, "2" + d, "12" + d}
		as := [5]string{"", "1", "10", "100", "1000"}
		for i := 0; i < 3; i++ {
			for j := 0; j < 5; j++ {
				s[i][j][0] = "0000 " + ds[i] + "<-W1R " + as[j] + a + " " + r
				if rx == "" {
					s[i][j][1] = "0000 " + ds[i] + "->W1R " + as[j] + a +
						" " + r
				} else {
					s[i][j][1] = "0000 " + ds[i] + "->W1R " + rx
				}
				s[i][j][2] = s[i][j][0] + "\n" + s[i][j][1]
			}
		}
		return
	}

	const addr = 2

	cmd = NewWriteRegCmd(0, addr, 0xBEEF)
	trunAll([5]string{
		"0000 0<-W1R 2 48879",
		"0000 0<-W1R 12 48879",
		"0000 0<-W1R 102 48879",
		"0000 0<-W1R 1002 48879",
		"0000 0<-W1R 10002 48879",
	})

	const devA = 3

	cmd = NewWriteRegCmd(devA, addr, 12345)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x86, 1})
	runAll(msg("Illegal Function"))

	cmd = NewWriteRegCmd(devA, addr, 1234)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x86, 2})
	runAll(msg("Illegal Data Address"))

	cmd = NewWriteRegCmd(devA, addr, 123)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x86, 3})
	runAll(msg("Illegal Data Value"))

	cmd = NewWriteRegCmd(devA, addr, 12)
	srx(cmd, []byte{0, 0, 0, 0, 0, 3, devA, 0x86, 4})
	runAll(msg("Slave Device Failure"))

	//

	cmd = NewWriteRegCmd(devA, addr, 1)
	srx(cmd, cmd.TxBytes())
	runAll(msg(""))
}

func BenchmarkWriteCoilsCmd(b *testing.B) {
	sda := func(c *WriteCoilsCmd, a byte) {
		c.SetDevAddr(a)
		b := c.RxBytes()
		(*b)[6] = a
	}
	srx := func(c *WriteCoilsCmd, b []byte) {
		r := c.RxBytes()
		*r = (*r)[:len(b)]
		copy(*r, b)
	}
	sa := func(c *WriteCoilsCmd, a uint16) {
		c.SetAddr(a)
		if r := c.RxBytes(); len(*r) != 9 {
			t := c.TxBytes()
			*r = (*r)[:12]
			copy(*r, t)
			(*r)[5] = 6
		}
	}

	var cmd *WriteCoilsCmd

	rx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Rx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	tx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Tx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	str := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.String()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}

	trunAll := func(x [5]string) {
		addr := cmd.Addr()
		as := [5]uint16{0, 10, 100, 1000, 10000}
		c := strconv.Itoa(cmd.Count())
		for i := 0; i < 5; i++ {
			si := strconv.Itoa(i + 1)
			cmd.SetAddr(as[i] + addr)
			b.Run(" tx:0,"+si+","+c, tx(x[i]))
			b.Run("str:0,"+si+","+c, str(x[i]))
		}
	}
	tmsg := func(v string) (x [5]string) {
		a := strconv.Itoa(int(cmd.Addr()))
		c := strconv.Itoa(cmd.Count())
		as := [5]string{"", "1", "10", "100", "1000"}
		for i := 0; i < 5; i++ {
			x[i] = "0000 0<-WC  " + as[i] + a + ":" + c + "[" + v + "]"
		}
		return
	}

	runAll := func(x [3][5][3]string) {
		devA := cmd.DevAddr()
		addr := cmd.Addr()
		s := strconv.Itoa(cmd.Count())
		if rx := cmd.RxBytes(); len(*rx) == 5 {
			s += ",ERR:" + strconv.Itoa(int((*rx)[2]))
		}

		ds := [3]byte{devA, 20 + devA, 120 + devA}
		as := [5]uint16{0, 10, 100, 1000, 10000}
		for i := 0; i < 3; i++ {
			sda(cmd, ds[i])
			si := strconv.Itoa(i + 1)
			for j := 0; j < 5; j++ {
				sa(cmd, as[j]+addr)
				sj := strconv.Itoa(j + 1)
				b.Run(" tx:"+si+","+sj+","+s, tx(x[i][j][0]))
				b.Run(" rx:"+si+","+sj+","+s, rx(x[i][j][1]))
				b.Run("str:"+si+","+sj+","+s, str(x[i][j][2]))
			}
		}
	}

	msg := func(v, rx string) (s [3][5][3]string) {
		d := strconv.Itoa(int(cmd.DevAddr()))
		a := strconv.Itoa(int(cmd.Addr()))
		c := strconv.Itoa(cmd.Count())

		ds := [3]string{d, "2" + d, "12" + d}
		as := [5]string{"", "1", "10", "100", "1000"}
		for i := 0; i < 3; i++ {
			for j := 0; j < 5; j++ {
				s[i][j][0] = "0000 " + ds[i] + "<-WC  " + as[j] + a + ":" + c +
					"[" + v + "]"
				if rx == "" {
					s[i][j][1] = "0000 " + ds[i] + "->WC  " + as[j] + a + ":" +
						c
				} else {
					s[i][j][1] = "0000 " + ds[i] + "->WC  " + rx
				}
				s[i][j][2] = s[i][j][0] + "\n" + s[i][j][1]
			}
		}
		return
	}

	const addr = 2
	const t = true
	const f = false

	t01 := []bool{t}
	const st01 = "1"
	f01 := []bool{f}
	const sf01 = "0"

	t05 := []bool{t, f, t, t, f}
	const st05 = "1 0 1 1 0"
	f05 := []bool{f, t, f, f, t}
	const sf05 = "0 1 0 0 1"

	t06 := []bool{t, f, t, t, f /* */, t}
	const st06 = "1 0 1 1 0  1"
	f06 := []bool{f, t, f, f, t /* */, f}
	const sf06 = "0 1 0 0 1  0"

	t10 := []bool{t, f, t, t, f /* */, t, t, f, t, t}
	const st10 = "1 0 1 1 0  1 1 0 1 1"
	f10 := []bool{f, t, f, f, t /* */, f, f, t, f, f}
	const sf10 = "0 1 0 0 1  0 0 1 0 0"

	t11 := []bool{
		t, f, t, t, f /* */, t, t, f, t, t,
		f,
	}
	const st11 = `
 1 0 1 1 0  1 1 0 1 1
 0
`
	f11 := []bool{
		f, t, f, f, t /* */, f, f, t, f, f,
		t,
	}
	const sf11 = `
 0 1 0 0 1  0 0 1 0 0
 1
`

	t15 := []bool{
		t, f, t, t, f /* */, t, t, f, t, t,
		f, t, t, f, t,
	}
	const st15 = `
 1 0 1 1 0  1 1 0 1 1
 0 1 1 0 1
`
	f15 := []bool{
		f, t, f, f, t /* */, f, f, t, f, f,
		t, f, f, t, f,
	}
	const sf15 = `
 0 1 0 0 1  0 0 1 0 0
 1 0 0 1 0
`

	t16 := []bool{
		t, f, t, t, f /* */, t, t, f, t, t,
		f, t, t, f, t /* */, t,
	}
	const st16 = `
 1 0 1 1 0  1 1 0 1 1
 0 1 1 0 1  1
`
	f16 := []bool{
		f, t, f, f, t /* */, f, f, t, f, f,
		t, f, f, t, f /* */, f,
	}
	const sf16 = `
 0 1 0 0 1  0 0 1 0 0
 1 0 0 1 0  0
`

	t20 := []bool{
		t, f, t, t, f /* */, t, t, f, t, t,
		f, t, t, f, t /* */, t, f, t, t, f,
	}
	const st20 = `
 1 0 1 1 0  1 1 0 1 1
 0 1 1 0 1  1 0 1 1 0
`
	f20 := []bool{
		f, t, f, f, t /* */, f, f, t, f, f,
		t, f, f, t, f /* */, f, t, f, f, t,
	}
	const sf20 = `
 0 1 0 0 1  0 0 1 0 0
 1 0 0 1 0  0 1 0 0 1
`

	t21 := []bool{
		t, f, t, t, f /* */, t, t, f, t, t,
		f, t, t, f, t /* */, t, f, t, t, f,
		t,
	}
	const st21 = `
 1 0 1 1 0  1 1 0 1 1
 0 1 1 0 1  1 0 1 1 0
 1
`
	f21 := []bool{
		f, t, f, f, t /* */, f, f, t, f, f,
		t, f, f, t, f /* */, f, t, f, f, t,
		f,
	}
	const sf21 = `
 0 1 0 0 1  0 0 1 0 0
 1 0 0 1 0  0 1 0 0 1
 0
`

	v := make([]bool, 1968)
	for i := 0; i < 246; i++ {
		v[i*8+0] = true
		v[i*8+1] = false
		v[i*8+2] = true
		v[i*8+3] = false
		v[i*8+4] = false
		v[i*8+5] = true
		v[i*8+6] = false
		v[i*8+7] = true
	}
	sv := strings.Repeat(`
 1 0 1 0 0  1 0 1 1 0
 1 0 0 1 0  1 1 0 1 0
 0 1 0 1 1  0 1 0 0 1
 0 1 1 0 1  0 0 1 0 1`, 49) + `
 1 0 1 0 0  1 0 1
`

	tv := [10][]bool{t01, f05, t06, f10, t11, f15, t16, f20, t21, v}
	ts := [10]string{st01, sf05, st06, sf10, st11, sf15, st16, sf20, st21, sv}
	for i := 0; i < len(tv); i++ {
		cmd = NewWriteCoilsCmd(0, addr, tv[i])
		trunAll(tmsg(ts[i]))
	}

	const devA = 3
	ev := [10][4][]bool{
		{f01, t01, f01, t01},
		{t05, f05, t05, f05},
		{f06, t06, f06, t06},
		{t10, f10, t10, f10},
		{f11, t11, f11, t11},
		{t15, f15, t15, f15},
		{f16, t16, f16, t16},
		{t20, f20, t20, f20},
		{f21, t21, f21, t21},
		{v, v, v, v},
	}
	esv := [10][4]string{
		{sf01, st01, sf01, st01},
		{st05, sf05, st05, sf05},
		{sf06, st06, sf06, st06},
		{st10, sf10, st10, sf10},
		{sf11, st11, sf11, st11},
		{st15, sf15, st15, sf15},
		{sf16, st16, sf16, st16},
		{st20, sf20, st20, sf20},
		{sf21, st21, sf21, st21},
		{sv, sv, sv, sv},
	}
	eb := [4][]byte{
		{0, 0, 0, 0, 0, 3, devA, 0x80 + 15, 1},
		{0, 0, 0, 0, 0, 3, devA, 0x80 + 15, 2},
		{0, 0, 0, 0, 0, 3, devA, 0x80 + 15, 3},
		{0, 0, 0, 0, 0, 3, devA, 0x80 + 15, 4},
	}
	es := [4]string{
		"Illegal Function",
		"Illegal Data Address",
		"Illegal Data Value",
		"Slave Device Failure",
	}
	for i := 0; i < 10; i++ {
		for j := 0; j < 4; j++ {
			cmd = NewWriteCoilsCmd(devA, addr, ev[i][j])
			srx(cmd, eb[j])
			runAll(msg(esv[i][j], es[j]))
		}
	}

	//

	rv := [10][]bool{f01, t05, f06, t10, f11, t15, f16, t20, f21, v}
	rs := [10]string{sf01, st05, sf06, st10, sf11, st15, sf16, st20, sf21, sv}
	for i := 0; i < len(rv); i++ {
		cmd = NewWriteCoilsCmd(devA, addr, rv[i])
		b := make([]byte, 12)
		copy(b, cmd.TxBytes())
		b[5] = 6
		srx(cmd, b)
		runAll(msg(rs[i], ""))
	}
}

func BenchmarkWriteRegsCmd(b *testing.B) {
	sda := func(c *WriteRegsCmd, a byte) {
		c.SetDevAddr(a)
		b := c.RxBytes()
		(*b)[6] = a
	}
	srx := func(c *WriteRegsCmd, b []byte) {
		r := c.RxBytes()
		*r = (*r)[:len(b)]
		copy(*r, b)
	}
	sa := func(c *WriteRegsCmd, a uint16) {
		c.SetAddr(a)
		if r := c.RxBytes(); len(*r) != 9 {
			t := c.TxBytes()
			*r = (*r)[:12]
			copy(*r, t)
			(*r)[5] = 6
		}
	}

	var cmd *WriteRegsCmd

	rx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Rx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	tx := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.Tx()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}
	str := func(x string) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = cmd.String()
			}
			if result != x {
				b.Fatalf("want %q got %q", x, result)
			} else {
				a, l := Alloc(), len(result)
				Debugf(b.Name(), "%d-%d %d", a, l, a-l)
			}
		}
	}

	trunAll := func(x [5]string) {
		addr := cmd.Addr()
		as := [5]uint16{0, 10, 100, 1000, 10000}
		c := strconv.Itoa(cmd.Count())
		for i := 0; i < 5; i++ {
			si := strconv.Itoa(i + 1)
			cmd.SetAddr(as[i] + addr)
			b.Run(" tx:0,"+si+","+c, tx(x[i]))
			b.Run("str:0,"+si+","+c, str(x[i]))
		}
	}
	tmsg := func(v string) (x [5]string) {
		a := strconv.Itoa(int(cmd.Addr()))
		c := strconv.Itoa(cmd.Count())
		as := [5]string{"", "1", "10", "100", "1000"}
		for i := 0; i < 5; i++ {
			x[i] = "0000 0<-WR  " + as[i] + a + ":" + c + "[" + v + "]"
		}
		return
	}

	runAll := func(x [3][5][3]string) {
		devA := cmd.DevAddr()
		addr := cmd.Addr()
		s := strconv.Itoa(cmd.Count())
		if rx := cmd.RxBytes(); len(*rx) == 5 {
			s += ",ERR:" + strconv.Itoa(int((*rx)[2]))
		}

		ds := [3]byte{devA, 20 + devA, 120 + devA}
		as := [5]uint16{0, 10, 100, 1000, 10000}
		for i := 0; i < 3; i++ {
			sda(cmd, ds[i])
			si := strconv.Itoa(i + 1)
			for j := 0; j < 5; j++ {
				sa(cmd, as[j]+addr)
				sj := strconv.Itoa(j + 1)
				b.Run(" tx:"+si+","+sj+","+s, tx(x[i][j][0]))
				b.Run(" rx:"+si+","+sj+","+s, rx(x[i][j][1]))
				b.Run("str:"+si+","+sj+","+s, str(x[i][j][2]))
			}
		}
	}

	msg := func(v, rx string) (s [3][5][3]string) {
		d := strconv.Itoa(int(cmd.DevAddr()))
		a := strconv.Itoa(int(cmd.Addr()))
		c := strconv.Itoa(cmd.Count())

		ds := [3]string{d, "2" + d, "12" + d}
		as := [5]string{"", "1", "10", "100", "1000"}
		for i := 0; i < 3; i++ {
			for j := 0; j < 5; j++ {
				s[i][j][0] = "0000 " + ds[i] + "<-WR  " + as[j] + a + ":" + c +
					"[" + v + "]"
				if rx == "" {
					s[i][j][1] = "0000 " + ds[i] + "->WR  " + as[j] + a + ":" +
						c
				} else {
					s[i][j][1] = "0000 " + ds[i] + "->WR  " + rx
				}
				s[i][j][2] = s[i][j][0] + "\n" + s[i][j][1]
			}
		}
		return
	}

	const addr = 2

	v01 := []uint16{0}
	const s01 = "    0"

	v05 := []uint16{65535, 1, 99, 100, 9999}
	const s05 = "65535     1    99   100  9999"

	v06 := []uint16{2, 33, 444, 5555, 60000, 7}
	const s06 = "    2    33   444  5555 60000 :     7"

	v10 := []uint16{33333, 4, 5555, 66, 777, 888, 99, 1010, 1, 11111}
	const s10 = "33333     4  5555    66   777 :   888    99  1010     1 11111"

	v11 := []uint16{4, 55, 666, 7777, 18888, 9999, 101, 11, 2, 13, 141}
	const s11 = `
     4    55   666  7777 18888 :  9999   101    11     2    13
   141
`

	v15 := []uint16{
		55555, 6666, 777, 88, 9,
		1, 11, 121, 1313, 14141,
		15151, 1616, 171, 18, 9,
	}
	const s15 = `
 55555  6666   777    88     9 :     1    11   121  1313 14141
 15151  1616   171    18     9
`

	v16 := []uint16{
		66, 777, 8888, 19999, 1010,
		111, 12, 3, 14, 151,
		1616, 17171, 1818, 191, 20,
		1,
	}
	const s16 = `
    66   777  8888 19999  1010 :   111    12     3    14   151
  1616 17171  1818   191    20 :     1
`

	v20 := []uint16{
		777, 88, 9, 10, 111,
		1212, 13131, 1414, 151, 16,
		171, 1818, 19191, 2020, 212,
		22, 3, 24, 252, 2626,
	}
	const s20 = `
   777    88     9    10   111 :  1212 13131  1414   151    16
   171  1818 19191  2020   212 :    22     3    24   252  2626
`

	v21 := []uint16{
		8888, 19999, 1010, 111, 12,
		3, 14, 151, 1616, 17171,
		1818, 191, 20, 1, 22,
		232, 2424, 25252, 2626, 272,
		28,
	}
	const s21 = `
  8888 19999  1010   111    12 :     3    14   151  1616 17171
  1818   191    20     1    22 :   232  2424 25252  2626   272
    28
`

	p := []uint16{10000, 2000, 300, 40, 0}
	v := make([]uint16, 123)
	for i := uint16(0); i < 123; i++ {
		v[i] = i + p[i%5] + 1
	}
	sv := `
 10001  2002   303    44     5 : 10006  2007   308    49    10
 10011  2012   313    54    15 : 10016  2017   318    59    20
 10021  2022   323    64    25 : 10026  2027   328    69    30
 10031  2032   333    74    35 : 10036  2037   338    79    40
 10041  2042   343    84    45 : 10046  2047   348    89    50
 10051  2052   353    94    55 : 10056  2057   358    99    60
 10061  2062   363   104    65 : 10066  2067   368   109    70
 10071  2072   373   114    75 : 10076  2077   378   119    80
 10081  2082   383   124    85 : 10086  2087   388   129    90
 10091  2092   393   134    95 : 10096  2097   398   139   100
 10101  2102   403   144   105 : 10106  2107   408   149   110
 10111  2112   413   154   115 : 10116  2117   418   159   120
 10121  2122   423
`

	tv := [10][]uint16{v01, v05, v06, v10, v11, v15, v16, v20, v21, v}
	ts := [10]string{s01, s05, s06, s10, s11, s15, s16, s20, s21, sv}
	for i := 0; i < len(tv); i++ {
		cmd = NewWriteRegsCmd(0, addr, tv[i])
		trunAll(tmsg(ts[i]))
	}

	const devA = 3
	ev := [3][4][]uint16{
		{v01, v05, v06, v10},
		{v11, v15, v16, v20},
		{v21, v, v01, v05},
	}
	esv := [3][4]string{
		{s01, s05, s06, s10},
		{s11, s15, s16, s20},
		{s21, sv, s01, s05},
	}
	eb := [4][]byte{
		{0, 0, 0, 0, 0, 3, devA, 0x80 + 16, 1},
		{0, 0, 0, 0, 0, 3, devA, 0x80 + 16, 2},
		{0, 0, 0, 0, 0, 3, devA, 0x80 + 16, 3},
		{0, 0, 0, 0, 0, 3, devA, 0x80 + 16, 4},
	}
	es := [4]string{
		"Illegal Function",
		"Illegal Data Address",
		"Illegal Data Value",
		"Slave Device Failure",
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			cmd = NewWriteRegsCmd(devA, addr, ev[i][j])
			srx(cmd, eb[j])
			runAll(msg(esv[i][j], es[j]))
		}
	}

	//

	rv := [10][]uint16{v01, v05, v06, v10, v11, v15, v16, v20, v21, v}
	rs := [10]string{s01, s05, s06, s10, s11, s15, s16, s20, s21, sv}
	for i := 0; i < len(rv); i++ {
		cmd = NewWriteRegsCmd(devA, addr, rv[i])
		b := make([]byte, 8)
		copy(b, cmd.TxBytes())
		srx(cmd, b)
		runAll(msg(rs[i], ""))
	}
}
