# Backends

Remco fetches configuration data from key-value stores via backends. Each backend can operate in two modes:

- **Watch mode** — the backend watches for changes in real time and triggers template re-rendering immediately.
- **Interval mode** — the backend polls at a fixed interval (in seconds). This is a reconciliation loop.

These modes are not mutually exclusive. You can enable both `watch` and `interval` simultaneously, so that watch provides low-latency updates and interval provides a safety net.

If neither `watch` nor `onetime` is set and `interval` is 0 or unset, the interval defaults to 60 seconds.

Every backend implements the [easykv](https://github.com/HeavyHorst/easykv) interface.

## Supported backends

| Backend | Watch | Interval | Notes |
|---------|-------|----------|-------|
| **etcd 2 & 3** | yes | yes | Use `version = 3` (default is 2) for etcd v3 API. The `scheme` setting only applies when nodes are discovered via DNS SRV records with etcd v2. |
| **consul** | yes | yes | Supports TLS client certificates. |
| **nats kv** | yes | yes | If no `nodes` are given, defaults to `nats://localhost:4222`. |
| **zookeeper** | yes | yes | |
| **redis** | no | yes | Polling only. |
| **vault** | no | yes | Supports multiple auth types: `token`, `approle`, `app-id`, `userpass`, `github`, `cert`, `kubernetes`. Polling only. |
| **env** | no | yes | Reads from environment variables. No config fields beyond the common backend options. |
| **file** | yes | yes | YAML or JSON files. Can be a local path or a remote HTTP/HTTPS URL. Use `httpheaders` to add custom headers to remote requests. |
| **mock** | no | yes | Intended for testing. Accepts an `error` field to simulate backend failures. |

## Backend configuration

The different configuration parameters can be found here: [backend configuration](../config/configuration-options.md#backend-configuration-options).

## Default backends

You can define shared backend defaults in a `[default_backends]` section. These values are deep-copied into every resource as a starting point. Resource-level backend settings then override the defaults.

```toml
[default_backends.etcd]
  nodes = ["http://etcd1:2379"]

[[resource]]
  [resource.backend.etcd]
    nodes = ["http://etcd2:2379"]   # overrides the default
    keys = ["/myapp"]
```

## Plugin backends

Remco also supports backends as plugins via JSON-RPC. See [plugins](plugins.md) for details.
