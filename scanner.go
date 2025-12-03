package modbus

import (
	"sync"
)

type CmdReq struct {
	Cmd Cmd
	Err chan<- error
}

func NewCmdReq(cmd Cmd) (CmdReq, <-chan error) {
	ch := make(chan error)
	return CmdReq{cmd, ch}, ch
}

type IController interface {
	Close()
	Send(Cmd) error
}

type SubScanner interface {
	Run(stop <-chan struct{}) <-chan CmdReq
}

type Scanner struct {
	Controller IController
	Subs       []SubScanner

	ch <-chan CmdReq
}

func (s *Scanner) Run(stop <-chan struct{}) {
	if len(s.Subs) == 0 {
		panic("empty Scanner.Subs")
	} else if len(s.Subs) == 1 {
		s.ch = s.Subs[0].Run(stop)
	} else {
		var wg sync.WaitGroup
		wg.Add(len(s.Subs))

		ch := make(chan CmdReq)
		for _, sub := range s.Subs {
			go func(sub SubScanner) {
				defer wg.Done()
				for req := range sub.Run(stop) {
					ch <- req
				}
			}(sub)
		}
		go func() {
			wg.Wait()
			close(ch)
		}()
		s.ch = ch
	}

	go s.run()
}

func (s *Scanner) run() {
	defer logPanic()
	defer s.Controller.Close()

	for req := range s.ch {
		req.Err <- s.Controller.Send(req.Cmd)
	}
}
