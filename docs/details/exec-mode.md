# Exec mode

Remco can run one arbitrary child process per template resource. When any of the provided templates change and the check command (if any) succeeds, remco will notify or restart the child process.

## How it works

If a `reload_signal` is configured, remco sends that signal to the child when templates change. If no `reload_signal` is set, remco kills the child process (with the configured `kill_signal`) and restarts it.

The child process must remain in the foreground. If it forks into the background, remco will be unable to track it and will restart it endlessly.

## Child process failure and restart

If the child process dies, the template resource is marked as failed. Remco automatically restarts it after a random delay of 0 to 30 seconds.

This jitter helps prevent thundering-herd problems in large clusters where many instances might restart simultaneously.

## Signal forwarding

Every signal that remco receives and does not handle itself (SIGINT, SIGTERM, SIGHUP, SIGCHLD) is forwarded to the child process. This means sending SIGUSR2 (or any custom signal) to the remco process will relay it to the child.

See [process lifecycle](process-lifecycle.md) for the full signal handling table.

## Configuration

The exec configuration parameters can be found here: [exec configuration](../config/configuration-options.md#exec-configuration-options).
