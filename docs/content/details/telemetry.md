---
title: "Telemetry"
date: 2020-09-05T21:12:31+03:00
next: /config/
prev: /details/process-lifecycle/
toc: true
weight: 50
---

Remco can expose different metrics about it's state using [go-metrics](https://github.com/armon/go-metrics).
You can configure any type of sink supported by go-metric using configuration file.
All the configured sinks will be aggregated using FanoutSink.


Currently supported sinks are:

  - **inmem**
  - **prometheus**
  - **statsd**
  - **statsite**

The different coniguration parameters can be found here: [telemetry configuration](/config/configuration-options/#telemetry-configuration-options).

Exposed metrics:
  - **files.template_execution_duration**
    - Duration of template execution
  - **files.check_command_duration**
    - Duration of check_command execution
  - **files.reload_command_duration**
    - Duration of reload_command execution
  - **files.stage_errors_total**
    - Total number of errors in file staging action
  - **files.staged_total**
    - Total number of successfully files staged
  - **files.sync_errors_total**
    - Total number of errors in file syncing action
  - **files.synced_total**
    - Total number of successfully files synced
  - **backends.sync_errors_total**
    - Total errors in backend sync action
  - **backends.synced_total**
    - Total number of successfully synced backends
