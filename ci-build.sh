#!/bin/bash

SRC_PATH=$(pwd)

docker run -v ${SRC_PATH}:/usr/src/skvs -w /usr/src/skvs golang:1.4 /bin/bash -c 'go get -d && go test -v && go build -v'
