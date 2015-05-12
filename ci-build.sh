#!/bin/bash
# THIS ONLY WORK IN OUR CI!
docker run --rm -v /data/jenkins/jobs/docker-skvs-build/workspace:/usr/src/skvs -w /usr/src/skvs golang:1.4 go build -v
