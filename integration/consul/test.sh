#!/bin/bash

# Configure consul

curl -X PUT http://127.0.0.1:8500/v1/kv/appdata/database/host -d '127.0.0.1'
curl -X PUT http://127.0.0.1:8500/v1/kv/appdata/database/password -d 'p@sSw0rd'
curl -X PUT http://127.0.0.1:8500/v1/kv/appdata/database/port -d '3306'
curl -X PUT http://127.0.0.1:8500/v1/kv/appdata/database/username -d 'remco'
curl -X PUT http://127.0.0.1:8500/v1/kv/appdata/upstream/app1 -d '10.0.1.10:8080'
curl -X PUT http://127.0.0.1:8500/v1/kv/appdata/upstream/app2 -d '10.0.1.11:8080'

remco --config integration/consul/consul.toml
cmp /tmp/remco-basic-test.conf ./integration/config/test.config || cat /tmp/remco-basic-test.conf