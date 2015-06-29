#!/bin/bash
set -e

SRC_PATH=$(pwd)

# we're calling docker from within a container with a path (/data/jenkins) mounted into this container (/var/jenkins)
# so the newly created container needs a different path (/data/jenkins).
docker run --rm -v /data${SRC_PATH#/var}:/usr/src/skvs -w /usr/src/skvs golang:1.4 /bin/bash -c 'go get -d && go test -v && go build -v'
