# Process lifecycle

Remco's lifecycle can be controlled with signals.

## Supported signals

| Signal | Behavior |
|--------|----------|
| SIGINT / os.Interrupt | Graceful shutdown. Remco stops all watchers and exits. |
| SIGTERM | Graceful shutdown. Same as SIGINT. |
| SIGHUP | Reload configuration. Remco re-reads the config file and restarts all resources. |
| SIGCHLD | Ignored. Remco handles child process reaping internally. |
| SIGUSR1 | If an `inmem` telemetry sink is configured, dumps runtime metrics to stderr. |
| Any other | Forwarded to the child process (in exec mode). |

## Graceful shutdown

On SIGINT or SIGTERM, remco:

1. Stops all backend watchers/pollers.
2. Sends the configured `kill_signal` (default: SIGTERM) to the child process if running in exec mode.
3. Waits up to `kill_timeout` seconds for the child to exit.
4. If the child has not exited, kills it (equivalent to SIGKILL).
5. Exits with a code equal to the number of resources that had errors (capped at 125).

## Configuration reload (SIGHUP)

When remco receives SIGHUP, it re-reads the configuration file from disk (including environment variable expansion). The new configuration replaces the old one — resources are stopped and restarted to match the new config.

## Exec-mode signal forwarding

When running in exec mode, any signal not explicitly handled by remco is forwarded to the child process. This includes SIGUSR2 and any custom signals.