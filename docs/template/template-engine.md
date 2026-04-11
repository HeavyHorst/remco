# Template engine

Remco uses [pongo2](https://github.com/flosch/pongo2), a Django-syntax template engine for Go. This is different from confd's Go `text/template` syntax. If you are migrating from confd, templates must be rewritten.

## Syntax overview

Pongo2 uses `{% %}` for tags and `{{ }}` for variable output:

```
{% for key in gets("/config/*") %}
{{ key }} = {{ getv(key) }}
{% endfor %}
```

Auto-escaping is disabled, so HTML entities are not inserted.

## Whitespace handling

Remco enables pongo2's `TrimBlocks` and `LStripBlocks` options:

- **TrimBlocks** — the first newline after a block tag (`{% if %}`, `{% for %}`, `{% endfor %}`, etc.) is stripped.
- **LStripBlocks** — leading whitespace (including tabs) before a block tag on the same line is stripped.

These match the behavior of Jinja2's similarly-named options. Without them, templates that mix tags and literal content often produce unwanted blank lines.

Example with both options enabled:

```
{% if true %}
hello
{% endif %}
```

Produces `\nhello\n` (not `\n\nhello\n\n`).

## Available functions and filters

- [Template functions](template-functions.md) — `getv`, `getvs`, `ls`, `fileExists`, etc.
- [Template filters](template-filters.md) — `parseInt`, `toYAML`, `base64`, etc.

## memkv store functions

The functions `exists`, `get`, `gets`, `getv`, `getvs`, `ls`, and `lsdir` come from the [memkv](https://github.com/HeavyHorst/memkv) library, which remco uses as an in-memory cache of the backend key-value data. They are available in every template without any additional configuration.