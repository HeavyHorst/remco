# Telemetry

Remco can expose different metrics about its state using [go-metrics](https://github.com/armon/go-metrics).
You can configure any type of sink supported by go-metrics through the configuration file.
All the configured sinks will be aggregated using FanoutSink.

Currently supported sinks are:

- **inmem**
- **prometheus**
- **statsd**
- **statsite**

The different configuration parameters can be found here: [telemetry configuration](../config/configuration-options.md#telemetry-configuration-options).

Exposed metrics:

- **files.template_execution_duration** — Duration of template execution
- **files.check_command_duration** — Duration of check_command execution
- **files.reload_command_duration** — Duration of reload_command execution
- **files.stage_errors_total** — Total number of errors in file staging action
- **files.staged_total** — Total number of successfully staged files
- **files.sync_errors_total** — Total number of errors in file syncing action
- **files.synced_total** — Total number of successfully synced files
- **backends.sync_errors_total** — Total errors in backend sync action
- **backends.synced_total** — Total number of successfully synced backends