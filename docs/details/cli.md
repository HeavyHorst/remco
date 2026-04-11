# Command-line reference

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-config` | `/etc/remco/config` | Path to the configuration file. |
| `-onetime` | `false` | Render all templates once and exit. Overrides the `onetime` setting on every backend to `true`. |
| `-version` | — | Print version information and exit. |

## Exit codes

When remco exits after finishing its work, the exit code reflects the number of resources that encountered errors. This applies to `-onetime` runs and any other run where all resources complete on their own. The exit code is capped at 125 — if more than 125 resources fail, remco exits with 125.

| Code | Meaning |
|------|---------|
| 0 | All resources completed successfully. |
| 1–125 | Number of resources that had errors. |
| 125+ | Clamped to 125. |

If remco receives `SIGINT` or `SIGTERM`, it performs a graceful shutdown and exits with code `0`.

## Version output

`remco -version` prints:

```
remco Version: <version>
UTC Build Time: <timestamp>
Git Commit Hash: <hash>
Go Version: <go version>
Go OS/Arch: <os>/<arch>
```

## Configuration reload

`-onetime` is not the only way to control remco's lifecycle. See [process lifecycle](process-lifecycle.md) for signal handling.
