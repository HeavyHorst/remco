#!/bin/bash

etcdctl --endpoints=127.0.0.1:2379 set /appdata/database/host 127.0.0.1
etcdctl --endpoints=127.0.0.1:2379 set /appdata/database/password p@sSw0rd
etcdctl --endpoints=127.0.0.1:2379 set /appdata/database/port 3306
etcdctl --endpoints=127.0.0.1:2379 set /appdata/database/username remco
etcdctl --endpoints=127.0.0.1:2379 set /appdata/upstream/app1 10.0.1.10:8080
etcdctl --endpoints=127.0.0.1:2379 set /appdata/upstream/app2 10.0.1.11:8080
cat integration/etcdv2/etcd.toml | etcdctl --endpoints=127.0.0.1:2379 set /remco/config > /dev/null

remco config etcd -c /remco/config --api-version=2
cmp /tmp/remco-basic-test-etcdv2.conf ./integration/config/test.config || cat /tmp/remco-basic-test-etcdv2.conf


