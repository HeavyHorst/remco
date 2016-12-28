---
date: 2016-12-03T15:13:50+01:00
next: /plugins/consul-plugin-example/
prev: /plugins/
title: env plugin example
toc: true
weight: 5
---

This is the env backend as a plugin.
If you want to try it yourself, then
just compile it and move the executable to /etc/remco/plugins.

```go
package main

import (
	"context"
	"log"
	"net/rpc/jsonrpc"

	"github.com/HeavyHorst/easyKV"
	"github.com/HeavyHorst/easyKV/env"
	"github.com/HeavyHorst/remco/backends/plugin"
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

func (e EnvRPCServer) WatchPrefix(args plugin.WatchConfig, resp *uint64) error {
	var err error
	*resp, err = e.Impl.WatchPrefix(context.Background(), args.Prefix, easyKV.WithKeys(args.Opts.Keys), easyKV.WithWaitIndex(args.Opts.WaitIndex))
	return err
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
