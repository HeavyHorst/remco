---
date: 2016-12-03T14:30:27+01:00
next: /details/backends/
prev: /details/commands/
title: zombie reaping
toc: true
weight: 20
---

See: https://blog.phusion.nl/2015/01/20/docker-and-the-pid-1-zombie-reaping-problem/

If Remco detects that it runs as pid 1 (for example in a Docker container) it will automatically reap zombie processes.
No additional init system is needed.
