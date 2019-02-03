#!/bin/bash

etcdctl --endpoints=http://127.0.0.1:2379 set /appdata/database/host 127.0.0.1
etcdctl --endpoints=http://127.0.0.1:2379 set /appdata/database/password p@sSw0rd
etcdctl --endpoints=http://127.0.0.1:2379 set /appdata/database/port 3306
etcdctl --endpoints=http://127.0.0.1:2379 set /appdata/database/username remco
etcdctl --endpoints=http://127.0.0.1:2379 set /appdata/upstream/app1 10.0.1.10:8080
etcdctl --endpoints=http://127.0.0.1:2379 set /appdata/upstream/app2 10.0.1.11:8080

remco --config integration/etcdv2/etcd.toml
cat /tmp/remco-basic-test-etcdv2.conf
