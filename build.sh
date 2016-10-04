#!/bin/bash

version=0.3.0
date=`date -u '+%Y-%m-%d %H:%M:%S'`
commit=`git rev-parse HEAD`

CGO_ENABLED=0 go build -a -tags netgo -ldflags \
		"-w -X 'main.version=$version' 
            -X 'main.buildDate=$date' 
			-X 'main.commit=$commit'"
