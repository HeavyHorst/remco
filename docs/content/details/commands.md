---
date: 2016-12-03T14:25:24+01:00
next: /details/zombie-reaping/
prev: /details/exec-mode/
title: commands
toc: true
weight: 15
---

Each template can have its own reload and check command. Both commands are executed in a sh-shell which means that operations like environment variable substitution or pipes should work correctly.

In the check command its additionally possible to reference to the rendered source template with {{ .src }}.

The check command must exit with status code 0 so that:

  - the reload command runs
  - the child process gets reloaded

The template configuration parameters can be found here: [template configuration](/config/configuration-options/#template-configuration-options).
