#!/bin/bash
# THIS ONLY WORK IN OUR CI!

echo "XXX: Jenkins variable:  " ${JOB_NAME}
echo "XXX: required path: /data/jenkins/jobs/docker-skvs-build/workspace"

docker run --rm -v /data/jenkins/jobs/docker-skvs-build/workspace:/usr/src/skvs -w /usr/src/skvs golang:1.4 /bin/bash -c 'go get -d && go build -v'
