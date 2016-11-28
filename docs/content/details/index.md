---
type: index
title: remco details
---

## Template resource

A template resource in remco consists of the following parts:

  - **one optional exec command.**
  - **one or many templates.**
  - **one or many backends.** 
  
Please note that it is not possible to use the same backend more than once per template resource.
It is for example not possible to use two different redis servers.


## Exec mode

Remco has the ability to run one arbitary child process per template resource.
When any of the provided templates change and the check command (if any) succeeds, remco will send the configurable reload signal to the child process.
Remco will kill and restart the child process if no reload signal is provided.
Additionally, every signal that remco receives will be forwarded to the child process.

The template resource will fail if the child process dies. It will be automatically restarted after a random amount of time (0-30s).
This also means that the child needs to remain in the foreground, otherwise the template resource will be restarted endlessly.

The exec configuration parameters can be found here: [exec configuration](../config/#exec-configuration-options).

## Commands (reload & check)

Each template can have its own reload and check command. Both commands are executed in a sh-shell which means that operations like environment variable substitution or pipes should work correctly.

In the check command its additionally possible to reference to the rendered source template with {{ .src }}.

The check command must exit with status code 0 so that:

  - the reload command runs
  - the child process gets reloaded

The template configuration parameters can be found here: [template configuration](../config/#template-configuration-options).

## Zombie reaping (pid 1)

See: https://blog.phusion.nl/2015/01/20/docker-and-the-pid-1-zombie-reaping-problem/

If Remco detects that it runs as pid 1 (for example in a Docker container) it will automatically reap zombie processes.
No additional init system is needed.

## Backends

Remco can fetch configuration data from a bunch of different kv-stores.
Some backends can be configured to watch for changes in the store to immediately react to these changes.
The other way is to provide a backend polling interval. 
These two modes are not mutual exclusive, you can watch for changes and run the interval processor as a reconcilation loop.

Every Backend needs to implement the [easyKV](https://github.com/HeavyHorst/easyKV) interface.
This is also the repository where the current implementations live.


Currently supported are:

  - **etcd 2 and 3** (interval and watch)
  - **consul** (interval and watch)
  - **zookeeper** (interval and watch)
  - **redis** (only interval)
  - **vault** (only interval)
  - **environment** (only interval)
  - **yaml/json files** (interval and watch)

The different coniguration parameters can be found here: [backend configuration](../config/#backend-configuration-options).

## Plugins

Remco supports backends as plugins.
There is no requirement that plugins be written in Go.
Every language that can provide a JSON-RPC API is ok.

Example: [env plugin](../plugins/).

## Running and Process Lifecycle

Remcos lifecycle can be controlled with several syscalls.

  - os.Interrupt(SIGINT on linux) and SIGTERM: remco will gracefully shut down
  - SIGHUP: remco will reload all configuration files.
