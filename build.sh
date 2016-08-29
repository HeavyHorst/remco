#!/bin/bash

go build -ldflags "-X github.com/HeavyHorst/remco/cmd.version=0.1.0 -X 'github.com/HeavyHorst/remco/cmd.buildDate=$(date -u '+%Y-%m-%d %H:%M:%S')' -X github.com/HeavyHorst/remco/cmd.commit=`git rev-parse HEAD`"
