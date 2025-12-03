#!/bin/sh
./bench.sh |tee new.txt
benchstat bench.txt new.txt
