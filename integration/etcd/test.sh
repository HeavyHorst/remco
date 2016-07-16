#!/bin/bash

curl -L -X PUT http://127.0.0.1:2379/v2/keys/database/host -d value=127.0.0.1
curl -L -X PUT http://127.0.0.1:2379/v2/keys/database/password -d value=p@sSw0rd
curl -L -X PUT http://127.0.0.1:2379/v2/keys/database/port -d value=3306
curl -L -X PUT http://127.0.0.1:2379/v2/keys/database/username -d value=remco
curl -L -X PUT http://127.0.0.1:2379/v2/keys/upstream/app1 -d value=10.0.1.10:8080
curl -L -X PUT http://127.0.0.1:2379/v2/keys/upstream/app2 -d value=10.0.1.11:8080

export ETCDCTL_API=3

etcdctl --endpoints=127.0.0.1:2379 put /database/host 127.0.0.1
etcdctl --endpoints=127.0.0.1:2379 put /database/password p@sSw0rd
etcdctl --endpoints=127.0.0.1:2379 put /database/port 3306
etcdctl --endpoints=127.0.0.1:2379 put /database/username remco
etcdctl --endpoints=127.0.0.1:2379 put /upstream/app1 10.0.1.10:8080
etcdctl --endpoints=127.0.0.1:2379 put /upstream/app2 10.0.1.11:8080

remco config file -c integration/etcd/etcd.toml

cmp /tmp/remco-basic-test-etcdv2.conf ./integration/config/test.config || cat /tmp/remco-basic-test-etcdv2.conf
cmp /tmp/remco-basic-test-etcdv3.conf ./integration/config/test.config || cat /tmp/remco-basic-test-etcdv3.conf