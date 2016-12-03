---
date: 2016-12-03T14:33:41+01:00
next: /config/
prev: /details/plugins/
title: process lifecycle
toc: true
weight: 40
---

Remcos lifecycle can be controlled with several syscalls.

  - os.Interrupt(SIGINT on linux) and SIGTERM: remco will gracefully shut down
  - SIGHUP: remco will reload all configuration files.
