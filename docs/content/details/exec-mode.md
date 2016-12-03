---
date: 2016-12-03T14:23:28+01:00
next: /details/commands/
prev: /details/template-resource/
title: exec mode
toc: true
weight: 10
---

Remco has the ability to run one arbitary child process per template resource.
When any of the provided templates change and the check command (if any) succeeds, remco will send the configurable reload signal to the child process.
Remco will kill and restart the child process if no reload signal is provided.
Additionally, every signal that remco receives will be forwarded to the child process.

The template resource will fail if the child process dies. It will be automatically restarted after a random amount of time (0-30s).
This also means that the child needs to remain in the foreground, otherwise the template resource will be restarted endlessly.

The exec configuration parameters can be found here: [exec configuration](/config/configuration-options/#exec-configuration-options).