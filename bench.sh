#!/bin/sh
go test -test.run=^# -test.count=10 -test.cpu=1 \
	-test.bench=. -test.benchmem -test.benchtime=1000x

# other flags
# -tags debug
# -test.bench=WriteRegs
