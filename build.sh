#!/bin/bash

CGO_ENABLED=0 go build -a -tags netgo -ldflags "-w -X main.version=0.2.0-dev -X 'main.buildDate=$(date -u '+%Y-%m-%d %H:%M:%S')' -X main.commit=`git rev-parse HEAD`"
