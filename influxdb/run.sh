#!/bin/bash
set -x
/bin/sh -c "cd /app && ./influxconfig join"

source /etc/default/influxdb

influxd --config /etc/influxdb.toml $INFLUXD_OPTS