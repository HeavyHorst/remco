#!/bin/bash

# Configure consul

curl -X PUT http://127.0.0.1:8500/v1/kv/database/host -d '127.0.0.1'
curl -X PUT http://127.0.0.1:8500/v1/kv/database/password -d 'p@sSw0rd'
curl -X PUT http://127.0.0.1:8500/v1/kv/database/port -d '3306'
curl -X PUT http://127.0.0.1:8500/v1/kv/database/username -d 'confd'
curl -X PUT http://127.0.0.1:8500/v1/kv/upstream/app1 -d '10.0.1.10:8080'
curl -X PUT http://127.0.0.1:8500/v1/kv/upstream/app2 -d '10.0.1.11:8080'

remco poll --onetime  consul \
    --log-level=debug \
    --src=./integration/templates/basic.conf.tmpl \
    --dst=/tmp/remco-basic-test.conf \
    --keys=/database/host, /database/password, /database/port, /database/username \
    --nodes=127.0.0.1:8500