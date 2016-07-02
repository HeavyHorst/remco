#!/bin/bash

curl -L -X PUT http://127.0.0.1:2379/v2/keys/database/host -d value=127.0.0.1
curl -L -X PUT http://127.0.0.1:2379/v2/keys/database/password -d value=p@sSw0rd
curl -L -X PUT http://127.0.0.1:2379/v2/keys/database/port -d value=3306
curl -L -X PUT http://127.0.0.1:2379/v2/keys/database/username -d value=remco
curl -L -X PUT http://127.0.0.1:2379/v2/keys/upstream/app1 -d value=10.0.1.10:8080
curl -L -X PUT http://127.0.0.1:2379/v2/keys/upstream/app2 -d value=10.0.1.11:8080

remco poll --onetime  etcd \
    --log-level=debug \
    --src=./integration/templates/basic.conf.tmpl \
    --dst=/tmp/remco-basic-test.conf \
    --nodes=http://127.0.0.1:2379 \
    --apiversion=2

cmp /tmp/remco-basic-test.conf ./integration/config/test.config || cat /tmp/remco-basic-test.conf

alias etcdctl3="docker run -ti --rm -e ETCDCTL_API=3 --net=host quay.io/coreos/etcd:v3.0.1 etcdctl"

etcdctl3 --endpoints=127.0.0.1:2379 put /database/host 127.0.0.1
etcdctl3 --endpoints=127.0.0.1:2379 put /database/password p@sSw0rd
etcdctl3 --endpoints=127.0.0.1:2379 put /database/port 3306
etcdctl3 --endpoints=127.0.0.1:2379 put /database/username remco
etcdctl3 --endpoints=127.0.0.1:2379 put /upstream/app1 10.0.1.10:8080
etcdctl3 --endpoints=127.0.0.1:2379 put /upstream/app2 10.0.1.11:8080

remco poll --onetime  etcd \
    --log-level=debug \
    --src=./integration/templates/basic.conf.tmpl \
    --dst=/tmp/remco-basic-test.conf \
    --nodes=127.0.0.1:2379 \
    --apiversion=3

cmp /tmp/remco-basic-test.conf ./integration/config/test.config || cat /tmp/remco-basic-test.conf