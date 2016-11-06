---
date: 2016-11-06T17:24:57+02:00
title: Sample resource file
---

```
[[template]]
  src = "/etc/remco/templates/haproxy.cfg"
  dst = "/etc/haproxy/haproxy.cfg"
  mode = "0644"

[backend]
  [backend.etcd]
    nodes = ["http://localhost:2379"]
    keys = ["/"]
    watch = true
    interval = 1
    version = 3

```
