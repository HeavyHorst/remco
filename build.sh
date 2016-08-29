#!/bin/bash

CGO_ENABLED=0 go build -a -tags netgo -ldflags "-w -X github.com/HeavyHorst/remco/cmd.version=0.1.0 -X 'github.com/HeavyHorst/remco/cmd.buildDate=$(date -u '+%Y-%m-%d %H:%M:%S')' -X github.com/HeavyHorst/remco/cmd.commit=`git rev-parse HEAD`"
