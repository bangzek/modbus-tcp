package modbus_test

import (
	"errors"
	"fmt"
	"io"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bangzek/clock"
	. "github.com/bangzek/modbus-tcp"
)

var _ = Describe("Controller", func() {
	const dsn = clock.DefaultScriptNow
	Context("single send", func() {
		It("runs just fine", func() {
			var tid uint16 = 1234
			t := time.Date(2024, time.March, 2, 10, 11, 12, 0, time.UTC)
			mc := new(clock.Mock)
			mc.NowScripts = []time.Duration{0, time.Second}
			SetClock(mc)
			mc.Start(t)
			cmd := NewReadCoilsCmd(3, 2, 1)
			conn := &MockConn{
				Writes: []WriteScript{
					{12, nil},
				},
				Reads: []ReadScript{
					{[]byte{4, 210, 0, 0, 0, 4, 3, 1, 1, 0b1}, nil},
				},
			}
			dialer := &MockDialer{
				Dials: []DialScript{
					{conn, TIMEOUT, WAIT, tid, nil},
				},
			}
			con := &Controller{
				Dialer: dialer,
			}
			log := NewLog()
			Expect(con.Send(cmd)).To(Succeed())
			con.Close()
			Expect(dialer.Calls).To(Equal([]bool{false}))
			Expect(conn.Calls).To(Equal([]string{
				"SWD 2024-03-02T10:11:15.001Z",
				"WRITE [04 D2 00 00 00 06 03 01 00 02 00 01]",
				"SRD 2024-03-02T10:11:16.001Z",
				"READ",
				"CLOSE",
			}))
			mc.Stop()
			Expect(mc.Calls()).To(HaveExactElements(
				"now",
				"now",
			))
			Expect(mc.Times()).To(HaveExactElements(
				t.Add(dsn),
				t.Add(dsn+time.Second),
			))
			Expect(log.Msgs).To(Equal([]string{
				"D:tx: 04 D2 00 00 00 06 03 01 00 02 00 01",
				"D:TX: 04D2 3<-RC  2:1",
				"D:rx: 04 D2 00 00 00 04 03 01 01 01",
				"D:RX: 04D2 3->RC  1[1]",
			}))
		})
	})

	Context("two send", func() {
		It("runs just fine", func() {
			var tid uint16 = 2345
			t := time.Date(2024, time.March, 2, 10, 11, 12, 0, time.UTC)
			mc := new(clock.Mock)
			mc.NowScripts = []time.Duration{0, time.Second, time.Second}
			SetClock(mc)
			mc.Start(t)
			cmd1 := NewReadDInputsCmd(3, 2, 1)
			cmd2 := NewWriteCoilCmd(0, 258, true)
			rwc := &MockConn{
				Writes: []WriteScript{
					{12, nil},
					{12, nil},
				},
				Reads: []ReadScript{
					{[]byte{9, 41, 0, 0, 0, 4, 3, 2, 1, 0b1}, nil},
				},
			}
			dialer := &MockDialer{
				Dials: []DialScript{
					{rwc, TIMEOUT, WAIT, tid, nil},
				},
			}
			con := &Controller{
				Dialer: dialer,
			}
			log := NewLog()
			Expect(con.Send(cmd1)).To(Succeed())
			Expect(con.Send(cmd2)).To(Succeed())
			Expect(dialer.Calls).To(Equal([]bool{false}))
			Expect(rwc.Calls).To(Equal([]string{
				"SWD 2024-03-02T10:11:15.001Z",
				"WRITE [09 29 00 00 00 06 03 02 00 02 00 01]",
				"SRD 2024-03-02T10:11:16.001Z",
				"READ",
				"SWD 2024-03-02T10:11:17.001Z",
				"WRITE [09 2A 00 00 00 06 00 05 01 02 FF 00]",
			}))
			mc.Stop()
			Expect(mc.Calls()).To(HaveExactElements(
				"now",
				"now",
				"now",
			))
			Expect(mc.Times()).To(HaveExactElements(
				t.Add(dsn),
				t.Add(dsn+time.Second),
				t.Add(dsn+2*time.Second),
			))
			Expect(log.Msgs).To(Equal([]string{
				"D:tx: 09 29 00 00 00 06 03 02 00 02 00 01",
				"D:TX: 0929 3<-RDI 2:1",
				"D:rx: 09 29 00 00 00 04 03 02 01 01",
				"D:RX: 0929 3->RDI 1[1]",
				"D:tx: 09 2A 00 00 00 06 00 05 01 02 FF 00",
				"D:TX: 092A 0<-W1C 258 true",
			}))
		})
	})

	Context("error on open", func() {
		It("returns that err", func() {
			t := time.Date(2024, time.March, 2, 10, 11, 12, 0, time.UTC)
			mc := new(clock.Mock)
			SetClock(mc)
			mc.Start(t)
			cmd1 := NewReadDInputsCmd(3, 2, 1)
			err1 := errors.New("one")
			cmd2 := NewWriteCoilCmd(0, 258, true)
			err2 := errors.New("two")
			dialer := &MockDialer{
				Dials: []DialScript{
					{nil, TIMEOUT, WAIT, 0, err1},
					{nil, TIMEOUT, WAIT, 0, err2},
				},
			}
			con := &Controller{
				Dialer: dialer,
			}
			log := NewLog()
			Expect(con.Send(cmd1)).To(MatchError(err1))
			Expect(con.Send(cmd2)).To(MatchError(err2))
			Expect(dialer.Calls).To(Equal([]bool{false, true}))
			mc.Stop()
			Expect(mc.Calls()).To(BeEmpty())
			Expect(mc.Times()).To(BeEmpty())
			Expect(log.Msgs).To(BeEmpty())
		})
	})

	Context("error on set write deadline", func() {
		It("returns that err", func() {
			var tid uint16 = 1234
			t := time.Date(2024, time.March, 2, 10, 11, 12, 0, time.UTC)
			mc := new(clock.Mock)
			SetClock(mc)
			mc.Start(t)
			err := errors.New("something")
			cmd := NewReadCoilsCmd(3, 2, 1)
			conn := &MockConn{
				WDeadlines: []error{err},
			}
			dialer := &MockDialer{
				Dials: []DialScript{
					{conn, TIMEOUT, WAIT, tid, nil},
				},
			}
			con := &Controller{
				Dialer: dialer,
			}
			log := NewLog()
			Expect(con.Send(cmd)).To(MatchError(err))
			Expect(dialer.Calls).To(Equal([]bool{false}))
			Expect(conn.Calls).To(Equal([]string{
				"SWD 2024-03-02T10:11:15.001Z",
				"CLOSE",
			}))
			mc.Stop()
			Expect(mc.Calls()).To(HaveExactElements("now"))
			Expect(mc.Times()).To(HaveExactElements(t.Add(dsn)))
			Expect(log.Msgs).To(BeEmpty())
		})
	})

	Context("error on tx", func() {
		It("returns that err", func() {
			var tid1 uint16 = 12345
			var tid2 uint16 = 23456
			t := time.Date(2024, time.March, 2, 10, 11, 12, 0, time.UTC)
			mc := new(clock.Mock)
			mc.NowScripts = []time.Duration{0, time.Second}
			SetClock(mc)
			mc.Start(t)
			cmd1 := NewReadDInputsCmd(3, 2, 1)
			err1 := errors.New("one")
			cmd2 := NewWriteCoilCmd(0, 258, true)
			rwc1 := &MockConn{Writes: []WriteScript{{8, err1}}}
			rwc2 := &MockConn{Writes: []WriteScript{{5, nil}}}
			dialer := &MockDialer{
				Dials: []DialScript{
					{rwc1, TIMEOUT, WAIT, tid1, nil},
					{rwc2, TIMEOUT, WAIT, tid2, nil},
				},
			}
			con := &Controller{
				Dialer: dialer,
			}
			log := NewLog()
			Expect(con.Send(cmd1)).To(MatchError(err1))
			Expect(con.Send(cmd2)).To(MatchError(io.ErrShortWrite))
			Expect(dialer.Calls).To(Equal([]bool{false, false}))
			Expect(rwc1.Calls).To(Equal([]string{
				"SWD 2024-03-02T10:11:15.001Z",
				"WRITE [30 39 00 00 00 06 03 02 00 02 00 01]",
				"CLOSE",
			}))
			Expect(rwc2.Calls).To(Equal([]string{
				"SWD 2024-03-02T10:11:16.001Z",
				"WRITE [5B A0 00 00 00 06 00 05 01 02 FF 00]",
				"CLOSE",
			}))
			mc.Stop()
			Expect(mc.Calls()).To(HaveExactElements(
				"now",
				"now",
			))
			Expect(mc.Times()).To(HaveExactElements(
				t.Add(dsn),
				t.Add(dsn+time.Second),
			))
			Expect(log.Msgs).To(Equal([]string{
				"D:tx: 30 39 00 00 00 06 03 02 00 02 00 01",
				"D:TX: 3039 3<-RDI 2:1",
				"D:tx: 5B A0 00 00 00 06 00 05 01 02 FF 00",
				"D:TX: 5BA0 0<-W1C 258 true",
			}))
		})
	})

	Context("error on set read deadline", func() {
		It("returns that err", func() {
			var tid uint16 = 45678
			t := time.Date(2024, time.March, 2, 10, 11, 12, 0, time.UTC)
			mc := new(clock.Mock)
			mc.NowScripts = []time.Duration{0, time.Second}
			SetClock(mc)
			mc.Start(t)
			cmd := NewReadCoilsCmd(3, 2, 1)
			err := errors.New("something")
			rwc := &MockConn{
				Writes: []WriteScript{
					{12, nil},
				},
				RDeadlines: []error{err},
			}
			dialer := &MockDialer{
				Dials: []DialScript{
					{rwc, TIMEOUT, WAIT, tid, nil},
				},
			}
			con := &Controller{
				Dialer: dialer,
			}
			log := NewLog()
			Expect(con.Send(cmd)).To(MatchError(err))
			Expect(dialer.Calls).To(Equal([]bool{false}))
			Expect(rwc.Calls).To(Equal([]string{
				"SWD 2024-03-02T10:11:15.001Z",
				"WRITE [B2 6E 00 00 00 06 03 01 00 02 00 01]",
				"SRD 2024-03-02T10:11:16.001Z",
				"CLOSE",
			}))
			mc.Stop()
			Expect(mc.Calls()).To(HaveExactElements(
				"now",
				"now",
			))
			Expect(mc.Times()).To(HaveExactElements(
				t.Add(dsn),
				t.Add(dsn+time.Second),
			))
			Expect(log.Msgs).To(Equal([]string{
				"D:tx: B2 6E 00 00 00 06 03 01 00 02 00 01",
				"D:TX: B26E 3<-RC  2:1",
			}))
		})
	})

	Context("error on rx", func() {
		It("returns that err", func() {
			var tid uint16 = 45678
			t := time.Date(2024, time.March, 2, 10, 11, 12, 0, time.UTC)
			mc := new(clock.Mock)
			mc.NowScripts = []time.Duration{0, time.Second}
			SetClock(mc)
			mc.Start(t)
			cmd := NewReadCoilsCmd(3, 2, 1)
			err := errors.New("something")
			rwc := &MockConn{
				Writes: []WriteScript{
					{12, nil},
				},
				Reads: []ReadScript{
					{nil, err},
				},
			}
			dialer := &MockDialer{
				Dials: []DialScript{
					{rwc, TIMEOUT, WAIT, tid, nil},
				},
			}
			con := &Controller{
				Dialer: dialer,
			}
			log := NewLog()
			Expect(con.Send(cmd)).To(MatchError(err))
			Expect(dialer.Calls).To(Equal([]bool{false}))
			Expect(rwc.Calls).To(Equal([]string{
				"SWD 2024-03-02T10:11:15.001Z",
				"WRITE [B2 6E 00 00 00 06 03 01 00 02 00 01]",
				"SRD 2024-03-02T10:11:16.001Z",
				"READ",
				"CLOSE",
			}))
			mc.Stop()
			Expect(mc.Calls()).To(HaveExactElements(
				"now",
				"now",
			))
			Expect(mc.Times()).To(HaveExactElements(
				t.Add(dsn),
				t.Add(dsn+time.Second),
			))
			Expect(log.Msgs).To(Equal([]string{
				"D:tx: B2 6E 00 00 00 06 03 01 00 02 00 01",
				"D:TX: B26E 3<-RC  2:1",
			}))
		})
	})

	Context("bad rx", func() {
		It("returns BadRxErr", func() {
			var tid uint16 = 56789
			t := time.Date(2024, time.March, 2, 10, 11, 12, 0, time.UTC)
			mc := new(clock.Mock)
			mc.NowScripts = []time.Duration{0, time.Second}
			SetClock(mc)
			mc.Start(t)
			rx := []byte{221, 213, 0, 0, 0, 4, 3, 2, 1, 0b1}
			cmd := NewReadCoilsCmd(3, 2, 1)
			rwc := &MockConn{
				Writes: []WriteScript{
					{12, nil},
				},
				Reads: []ReadScript{
					{rx, nil},
				},
			}
			dialer := &MockDialer{
				Dials: []DialScript{
					{rwc, TIMEOUT, WAIT, tid, nil},
				},
			}
			con := &Controller{
				Dialer: dialer,
			}
			log := NewLog()
			Expect(con.Send(cmd)).To(MatchError(
				"invalid response: [DD D5 00 00 00 04 03 02 01 01]"))
			Expect(dialer.Calls).To(Equal([]bool{false}))
			Expect(rwc.Calls).To(Equal([]string{
				"SWD 2024-03-02T10:11:15.001Z",
				"WRITE [DD D5 00 00 00 06 03 01 00 02 00 01]",
				"SRD 2024-03-02T10:11:16.001Z",
				"READ",
				"CLOSE",
			}))
			mc.Stop()
			Expect(mc.Calls()).To(HaveExactElements(
				"now",
				"now",
			))
			Expect(mc.Times()).To(HaveExactElements(
				t.Add(dsn),
				t.Add(dsn+time.Second),
			))
			Expect(log.Msgs).To(Equal([]string{
				"D:tx: DD D5 00 00 00 06 03 01 00 02 00 01",
				"D:TX: DDD5 3<-RC  2:1",
				"D:rx: DD D5 00 00 00 04 03 02 01 01",
			}))
		})
	})

	Context("bad io.Reader", func() {
		It("returns ErrNoProgress", func() {
			var tid uint16 = 56789
			t := time.Date(2024, time.March, 2, 10, 11, 12, 0, time.UTC)
			mc := new(clock.Mock)
			mc.NowScripts = []time.Duration{0, time.Second}
			SetClock(mc)
			mc.Start(t)
			cmd := NewReadCoilsCmd(3, 2, 1)
			rwc := &MockConn{
				Writes: []WriteScript{
					{12, nil},
				},
			}
			dialer := &MockDialer{
				Dials: []DialScript{
					{rwc, TIMEOUT, WAIT, tid, nil},
				},
			}
			con := &Controller{
				Dialer: dialer,
			}
			log := NewLog()
			Expect(con.Send(cmd)).To(MatchError(io.ErrNoProgress))
			Expect(dialer.Calls).To(Equal([]bool{false}))
			Expect(rwc.Calls).To(Equal([]string{
				"SWD 2024-03-02T10:11:15.001Z",
				"WRITE [DD D5 00 00 00 06 03 01 00 02 00 01]",
				"SRD 2024-03-02T10:11:16.001Z",
				"READ",
				"CLOSE",
			}))
			mc.Stop()
			Expect(mc.Calls()).To(HaveExactElements(
				"now",
				"now",
			))
			Expect(mc.Times()).To(HaveExactElements(
				t.Add(dsn),
				t.Add(dsn+time.Second),
			))
			Expect(log.Msgs).To(Equal([]string{
				"D:tx: DD D5 00 00 00 06 03 01 00 02 00 01",
				"D:TX: DDD5 3<-RC  2:1",
			}))
		})
	})
})

type MockDialer struct {
	Dials []DialScript

	Calls []bool
	i     int
}

type DialScript struct {
	Conn    Conn
	Timeout time.Duration
	Wait    time.Duration
	TxId    uint16
	Err     error
}

func (m *MockDialer) Dial(
	repeat bool,
) (conn Conn, timeout, wait time.Duration, txId uint16, err error) {
	if m.i < len(m.Dials) {
		conn = m.Dials[m.i].Conn
		timeout = m.Dials[m.i].Timeout
		wait = m.Dials[m.i].Wait
		txId = m.Dials[m.i].TxId
		err = m.Dials[m.i].Err
	}
	m.i++
	m.Calls = append(m.Calls, repeat)
	return
}

type MockConn struct {
	WDeadlines []error
	Writes     []WriteScript
	RDeadlines []error
	Reads      []ReadScript

	Calls []string

	iWDeadline int
	iWrite     int
	iRDeadline int
	iRead      int
}

type WriteScript struct {
	N   int
	Err error
}

type ReadScript struct {
	Bytes []byte
	Err   error
}

func (m *MockConn) SetWriteDeadline(t time.Time) (err error) {
	if m.iWDeadline < len(m.WDeadlines) {
		err = m.WDeadlines[m.iWDeadline]
	}
	m.Calls = append(m.Calls, "SWD "+t.Format(time.RFC3339Nano))
	m.iWDeadline++
	return
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	if m.iWrite < len(m.Writes) {
		n = m.Writes[m.iWrite].N
		err = m.Writes[m.iWrite].Err
	}
	m.Calls = append(m.Calls, fmt.Sprintf("WRITE [% X]", b))
	m.iWrite++
	return
}

func (m *MockConn) SetReadDeadline(t time.Time) (err error) {
	if m.iRDeadline < len(m.RDeadlines) {
		err = m.RDeadlines[m.iRDeadline]
	}
	m.Calls = append(m.Calls, "SRD "+t.Format(time.RFC3339Nano))
	m.iRDeadline++
	return
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	if m.iRead < len(m.Reads) {
		s := m.Reads[m.iRead]
		if len(b) < len(s.Bytes) {
			panic(fmt.Sprintf("Invalid MockConn.ReadScript[%d].Bytes %d>%d",
				m.iRead, len(s.Bytes), len(b)))
		}
		if len(s.Bytes) > 0 {
			copy(b, s.Bytes)
			n = len(s.Bytes)
		}
		err = s.Err
	}
	m.Calls = append(m.Calls, "READ")
	m.iRead++
	return
}

func (m *MockConn) Close() error {
	m.Calls = append(m.Calls, "CLOSE")
	return nil
}
