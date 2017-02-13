#!/bin/bash

version=0.8.0
date=`date -u '+%Y-%m-%d %H:%M:%S'`
commit=`git rev-parse HEAD`
os_list=( "linux" "darwin" "windows" )

for os in "${os_list[@]}"
do
    echo "Build remco for $os"
    GOOS=${os} CGO_ENABLED=0 go build -a -tags netgo -o bin/remco_${os} -ldflags \
		"-w -X 'main.version=$version' 
            -X 'main.buildDate=$date' 
			-X 'main.commit=$commit'" \
            github.com/HeavyHorst/remco/cmd/remco

    cd bin && zip -r remco_${version}_${os}_amd64.zip remco_${os} && cd ..
done
