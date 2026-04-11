# Commands

Each template can have two optional commands:

## Check command (`check_cmd`)

Executed *before* the rendered template is written to the destination path. The check command runs in a shell (`/bin/sh -c`).

- The rendered template is written to a temporary staging file first. You can reference this staging file with `{{ .src }}` in the check command.
- If the check command returns a non-zero exit code, the destination file is **not** overwritten. The old configuration is left in place.
- If no `check_cmd` is configured, the template is always written.

Example:

```toml
check_cmd = "nginx -t -c {{ .src }}"
```

## Reload command (`reload_cmd`)

Executed *after* the destination file has been updated. The reload command also runs in a shell.

- You can reference the destination path with `{{ .dst }}`.
- If the reload command returns a non-zero exit code, remco logs an error but the destination file remains updated.

Example:

```toml
reload_cmd = "systemctl reload nginx"
```

## Resource-level commands

A template resource also supports two higher-level commands:

- **`start_cmd`** — runs once when all templates in the resource have been processed successfully for the first time.
- **`reload_cmd`** (resource-level) — runs after any template in the resource is updated. This is distinct from the template-level `reload_cmd`.

## Interaction with watch mode

When a backend is in watch mode and a change is detected:

1. The template is re-rendered.
2. If configured, `check_cmd` runs against the staging file.
3. If the check passes, the staging file replaces the destination.
4. If configured, the template-level `reload_cmd` runs.
5. If configured, the resource-level `reload_cmd` runs.

The template configuration parameters can be found here: [template configuration](../config/configuration-options.md#template-configuration-options).
