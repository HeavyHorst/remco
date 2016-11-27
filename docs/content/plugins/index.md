---
title: plugins
---

## ENV backend as plugin [Example]

This is the env backend as a plugin.
If you want to try it yourself, then
just compile it and move the executable to /etc/remco/plugins.

```go
package main

import (
	"log"
	"net/rpc/jsonrpc"

	"github.com/HeavyHorst/easyKV"
	"github.com/HeavyHorst/easyKV/env"
	"github.com/natefinch/pie"
)

func main() {
	p := pie.NewProvider()
	if err := p.RegisterName("Plugin", &EnvRPCServer{}); err != nil {
		log.Fatalf("failed to register Plugin: %s", err)
	}
	p.ServeCodec(jsonrpc.NewServerCodec)
}

type EnvRPCServer struct {
	// This is the real implementation
	Impl easyKV.ReadWatcher
}

func (e *EnvRPCServer) Init(args map[string]interface{}, resp *bool) error {
	// use the data in args to create the ReadWatcher
	// env var doesn't need any data

	var err error
	e.Impl, err = env.New()
	return err
}

func (e *EnvRPCServer) GetValues(args []string, resp *map[string]string) error {
	erg, err := e.Impl.GetValues(args)
	if err != nil {
		return err
	}
	*resp = erg
	return nil
}

func (e *EnvRPCServer) Close(args interface{}, resp *interface{}) error {
	e.Impl.Close()
	return nil
}
```

Then create a config file with this backend section.

```toml
[backend]
  [[backend.plugin]]
    path = "/etc/remco/plugins/env"
    keys = ["/"]
    interval = 60
	watch = false
	[backend.plugin.config]
	 # these parameters are not used in the env backend plugin
	 # but other plugins may need some data (password, prefix ...)
	 a = "hallo"
	 b = "moin"
```
