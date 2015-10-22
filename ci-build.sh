#!/bin/bash
set -e

SRC_PATH=$(pwd)

docker run -v ${SRC_PATH}:/usr/src/platform-skvs -w /usr/src/platform-skvs golang:1.4 /bin/bash -c 'go get -d && go test -v && go build -v'
