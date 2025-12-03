#!/bin/sh
cd $(dirname $0)
env GOOS=linux GOARCH=arm GOARM=7 go build && \
    ls -l $(basename $PWD) &&
    file $(basename $PWD) | cut -d, -f1,3
