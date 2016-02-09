#!/bin/bash
set -x
/bin/sh -c "cd /app && ./influxconfig join"

source /etc/default/influxdb

HOSTIP=$(hostname -i)
export INFLUXDB_META_HTTP_BIND_ADDRESS=$HOSTIP:8091
export INFLUXDB_META_BIND_ADDRESS=$HOSTIP:8088
export INFLUXDB_HTTP_BIND_ADDRESS=$HOSTIP:8086

influxd --config /etc/influxdb.toml $INFLUXD_OPTS