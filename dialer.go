package modbus

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	PORT    = 502
	TIMEOUT = 3 * time.Second
	WAIT    = 5 * time.Millisecond
)

type DialErr struct {
	Addr string
	Err  error
}

func (e DialErr) Error() string {
	return e.Err.Error() + " while opening " + e.Addr
}

func (e DialErr) Unwrap() error {
	return e.Err
}

type Dialer struct {
	Host    string
	Port    int
	Timeout time.Duration
	Wait    time.Duration
}

func (p *Dialer) Dial(
	repeat bool,
) (Conn, time.Duration, time.Duration, uint16, error) {
	if p.Host == "" {
		panic("empty Dialer.Host")
	}
	if p.Port <= 0 {
		p.Port = PORT
	}
	if p.Timeout <= 0 {
		p.Timeout = TIMEOUT
	}
	if p.Wait <= 0 {
		p.Wait = WAIT
	}

	a := fmt.Sprintf("%s:%d", p.Host, p.Port)
	if repeat {
		debugLog("Dialing %s", a)
	} else {
		log("Dialing %s", a)
	}
	conn, err := net.DialTimeout("tcp", a, p.Timeout)

	if err != nil {
		return nil, p.Timeout, p.Wait, 0, DialErr{a, err}
	}
	log("%s opened", a)
	t := time.Now().UnixNano()
	la := conn.LocalAddr().(*net.TCPAddr)
	ra := conn.RemoteAddr().(*net.TCPAddr)
	var mask int64
	if t%2 == 0 {
		mask = int64(la.Port)<<48 | int64(ra.Port)<<32
	} else {
		mask = int64(ra.Port)<<48 | int64(la.Port)<<32
	}
	txId := uint16(rand.New(rand.NewSource(t ^ mask)).Int31n(0xFFFF))

	return conn, p.Timeout, p.Wait, txId, nil
}
