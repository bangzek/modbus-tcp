package modbus

import (
	"io"
	"time"

	"github.com/bangzek/clock"
)

var (
	ctime = clock.New()
)

type Conn interface {
	io.ReadWriteCloser
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

type ConnDialer interface {
	Dial(bool) (Conn, time.Duration, time.Duration, uint16, error)
}

type Controller struct {
	Dialer ConnDialer

	conn    Conn
	timeout time.Duration
	wait    time.Duration
	txId    uint16
	repeat  bool
}

func (c *Controller) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

func (c *Controller) Send(cmd Cmd) error {
	if c.conn == nil {
		var err error
		c.conn, c.timeout, c.wait, c.txId, err = c.Dialer.Dial(c.repeat)
		if err != nil {
			c.repeat = true
			return err
		}
		c.repeat = false
	}

	if err := c.conn.SetWriteDeadline(ctime.Now().Add(c.timeout)); err != nil {
		c.Close()
		return err
	}

	cmd.SetTxId(c.txId)
	c.txId++
	tx := cmd.TxBytes()
	debugLog("tx: % X", tx)
	debugLog("TX: %s", cmd.Tx())
	if n, err := c.conn.Write(tx); err != nil {
		c.Close()
		return err
	} else if n != len(tx) {
		c.Close()
		return io.ErrShortWrite
	}

	time.Sleep(c.wait)

	rx := cmd.RxBytes()
	if cap(*rx) == 0 {
		return nil
	}
	*rx = (*rx)[:cap(*rx)]

	if err := c.conn.SetReadDeadline(ctime.Now().Add(c.timeout)); err != nil {
		c.Close()
		return err
	}
	if n, err := c.conn.Read(*rx); err != nil {
		c.Close()
		return err
	} else if n == 0 {
		c.Close()
		return io.ErrNoProgress
	} else {
		*rx = (*rx)[:n]
	}
	debugLog("rx: % X", *rx)
	if cmd.IsValidRx() {
		debugLog("RX: %s", cmd.Rx())
	} else {
		c.Close()
		return BadRxErr(*rx)
	}
	return cmd.Err()
}
