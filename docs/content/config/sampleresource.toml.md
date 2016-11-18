---
date: 2016-11-06T17:24:57+02:00
title: Sample resource file
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
