#!/bin/bash

zookeeper-3.4.9/bin/zkCli.sh create /appdata ""
zookeeper-3.4.9/bin/zkCli.sh create /appdata/database ""
zookeeper-3.4.9/bin/zkCli.sh create /appdata/upstream ""

zookeeper-3.4.9/bin/zkCli.sh create /appdata/database/password "p@sSw0rd"
zookeeper-3.4.9/bin/zkCli.sh create /appdata/database/port "3306"
zookeeper-3.4.9/bin/zkCli.sh create /appdata/database/username "remco"

zookeeper-3.4.9/bin/zkCli.sh create /appdata/upstream/app1 "10.0.1.10:8080"
zookeeper-3.4.9/bin/zkCli.sh create /appdata/upstream/app2 "10.0.1.11:8080"

remco --config integration/zookeeper/zookeeper.toml
cat /tmp/remco-basic-test.conf
