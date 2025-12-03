package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/bangzek/modbus-tcp"
)

func main() {
	modbus.InfoLogFunc = log.Printf
	modbus.DebugLogFunc = log.Printf

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s IP\n"+
			" e.g.: %s 172.16.17.18\n",
			os.Args[0],
			os.Args[0])
		os.Exit(1)
	}

	con := &modbus.Controller{
		Dialer: &modbus.Dialer{
			Host: os.Args[1],
		},
	}

	demoVegaScan(con)
}

// This is for VegaScan ATG
func demoVegaScan(con *modbus.Controller) {
	vals := modbus.NewReadIRegsCmd(1, 0, 60)
	if err := con.Send(vals); err != nil {
		fmt.Printf("ERR: %s\n", err)
		fmt.Println()
	}

	floats := modbus.NewReadIRegsCmd(1, 1000, 120)
	if err := con.Send(floats); err != nil {
		fmt.Printf("ERR: %s\n", err)
		fmt.Println()
	}

	for i := 0; i < 60; i++ {
		f := math.Float32frombits(
			uint32(floats.Reg(i*2+1))<<16 |
				uint32(floats.Reg(i*2)))
		fmt.Printf("%d %g\n", i, f)
	}
}
