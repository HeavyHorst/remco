# remco

remco is a lightweight configuration management tool that renders templates from backend data and reloads services when values change.

It watches backends like etcd, consul, vault, redis, zookeeper, NATS KV, or environment variables, pushes changes through template rendering, and optionally execs or signals a child process.

## Sections

### Runtime & Operations

- [Template resource](details/template-resource.md) — how a template resource is structured
- [Exec mode](details/exec-mode.md) — running a child process per resource
- [Commands](details/commands.md) — check and reload commands
- [Process lifecycle](details/process-lifecycle.md) — signal handling
- [CLI reference](details/cli.md) — flags, exit codes, and version info
- [Zombie reaping](details/zombie-reaping.md) — automatic reaping when running as PID 1
- [Telemetry](details/telemetry.md) — metrics sinks

### Configuration & Backends

- [Environment variables](config/environment-variables.md) — variable substitution in config
- [Configuration options](config/configuration-options.md) — full config reference
- [Sample config](config/sample-config.md) — complete TOML example
- [Sample resource](config/sample-resource.md) — resource block example
- [Backends](details/backends.md) — supported key-value stores
- [Plugins](details/plugins.md) — JSON-RPC backend plugins

### Template Authoring

- [Template engine](template/template-engine.md) — pongo2 syntax and whitespace handling
- [Template functions](template/template-functions.md) — available template functions
- [Template filters](template/template-filters.md) — built-in and custom filters

### Extensions & Examples

- [Env plugin example](plugins/env-plugin-example.md) — environment backend as a plugin
- [Consul plugin example](plugins/consul-plugin-example.md) — consul service endpoint plugin
- [Dynamic haproxy configuration](examples/haproxy.md) — full tutorial with Docker
