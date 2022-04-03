---
date: 2016-12-03T14:31:14+01:00
next: /details/plugins/
prev: /details/zombie-reaping/
title: backends
toc: true
weight: 30
---

Remco can fetch configuration data from a bunch of different kv-stores.
Some backends can be configured to watch for changes in the store to immediately react to these changes.
The other way is to provide a backend polling interval. 
These two modes are not mutual exclusive, you can watch for changes and run the interval processor as a reconciliation loop.

Every Backend needs to implement the [easyKV](https://github.com/HeavyHorst/easyKV) interface.
This is also the repository where the current implementations live.


Currently supported are:

  - **etcd 2 and 3** (interval and watch)
  - **consul** (interval and watch)
  - **nats kv** (interval and watch)
  - **zookeeper** (interval and watch)
  - **redis** (only interval)
  - **vault** (only interval)
  - **environment** (only interval)
  - **yaml/json files** (interval and watch)

The different configuration parameters can be found here: [backend configuration](/config/configuration-options/#backend-configuration-options).
