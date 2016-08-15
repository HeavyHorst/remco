#!/bin/bash

etcdctl --endpoints=127.0.0.1:2379 set /appdata/database/host 127.0.0.1
etcdctl --endpoints=127.0.0.1:2379 set /appdata/database/password p@sSw0rd
etcdctl --endpoints=127.0.0.1:2379 set /appdata/database/port 3306
etcdctl --endpoints=127.0.0.1:2379 set /appdata/database/username remco
etcdctl --endpoints=127.0.0.1:2379 set /appdata/upstream/app1 10.0.1.10:8080
etcdctl --endpoints=127.0.0.1:2379 set /appdata/upstream/app2 10.0.1.11:8080
cat integration/etcd/etcd.toml | etcdctl --endpoints=127.0.0.1:2379 set /remco/config

remco config etcd -c /remco/test --api-version=2
cmp /tmp/remco-basic-test-etcdv2.conf ./integration/config/test.config || cat /tmp/remco-basic-test-etcdv2.conf
cmp /tmp/remco-basic-test-etcdv3.conf ./integration/config/test.config || cat /tmp/remco-basic-test-etcdv3.conf

export ETCDCTL_API=3
etcdctl --endpoints=127.0.0.1:2379 put /appdata/database/host 127.0.0.1
etcdctl --endpoints=127.0.0.1:2379 put /appdata/database/password p@sSw0rd
etcdctl --endpoints=127.0.0.1:2379 put /appdata/database/port 3306
etcdctl --endpoints=127.0.0.1:2379 put /appdata/database/username remco
etcdctl --endpoints=127.0.0.1:2379 put /appdata/upstream/app1 10.0.1.10:8080
etcdctl --endpoints=127.0.0.1:2379 put /appdata/upstream/app2 10.0.1.11:8080
cat integration/etcd/etcd.toml | etcdctl --endpoints=127.0.0.1:2379 put /remco/config

remco config etcd -c /remco/test --api-version=3
cmp /tmp/remco-basic-test-etcdv2.conf ./integration/config/test.config || cat /tmp/remco-basic-test-etcdv2.conf
cmp /tmp/remco-basic-test-etcdv3.conf ./integration/config/test.config || cat /tmp/remco-basic-test-etcdv3.conf

