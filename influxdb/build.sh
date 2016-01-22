#! /bin/bash
set -x
set -e
echo "building influxdb config executable for Linux"
GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o influxconfig main.go


echo "building influxdb docker image"
 pushd $( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
 docker build -t spatel/influxdb:stresstest .
 popd

rm influxconfig