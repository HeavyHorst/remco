---
date: 2016-12-03T15:01:23+01:00
next: /template/
prev: /config/sample-config/
title: sample resource
toc: true
weight: 20
---

```
[exec]
  command       = "/path/to/program"
  kill_signal   = "SIGTERM"
  reload_signal = "SIGHUP"
  kill_timeout  = 10
  splay         = 10


[[template]]
  src           = "/etc/remco/templates/haproxy.cfg"
  dst           = "/etc/haproxy/haproxy.cfg"
  reload_cmd 	= "haproxy -f /etc/haproxy/haproxy.cfg -p /var/run/haproxy.pid -D -sf `cat /var/run/haproxy.pid`"
  mode          = "0644"

[backend]
  [backend.etcd]
    nodes    = ["http://localhost:2379"]
    keys     = ["/service-registry"]
    watch    = true
    interval = 60
    version  = 3

```

