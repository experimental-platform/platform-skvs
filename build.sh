#!/bin/bash
docker run --rm -v "$PWD":/usr/src/skvs -w /usr/src/skvs golang:1.4 go build -v
