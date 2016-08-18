#!/bin/bash

# Configure redis

redis-cli set /appdata/database/host 127.0.0.1
redis-cli set /appdata/database/password p@sSw0rd
redis-cli set /appdata/database/port 3306
redis-cli set /appdata/database/username remco
redis-cli set /appdata/upstream/app1 10.0.1.10:8080
redis-cli set /appdata/upstream/app2 10.0.1.11:8080
redis-cli -x set /remco/config < integration/redis/redis.toml

remco config redis -c /remco/config
cmp /tmp/remco-basic-test.conf ./integration/config/test.config || cat /tmp/remco-basic-test.conf