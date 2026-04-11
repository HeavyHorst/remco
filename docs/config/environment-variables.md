# Environment variables

Environment variable substitution is applied to the entire configuration file before TOML parsing. You can use `$VARIABLE_NAME` or `${VARIABLE_NAME}` and the text will be replaced with the value of the environment variable.

```
[resource]
  [resource.backend.etcd]
    nodes = ["${ETCD_HOST}:2379"]
    username = "$ETCD_USER"
    password = "$ETCD_PASS"
```

## How it works

Substitution is performed by Go's `os.ExpandEnv`, which means:

- Only simple `$VAR` and `${VAR}` forms are supported.
- Bash-style defaults like `${VAR:-default}` are **not** supported. Use a template function like `getv` with a default value inside templates instead.
- Undefined variables expand to an empty string.

## Quoting

Because substitution happens before TOML parsing, values that may be empty or contain special characters should be quoted:

```toml
password = "${MY_PASSWORD}"
```

Without quotes, an empty expansion could produce invalid TOML.