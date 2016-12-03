---
date: 2016-12-03T15:14:31+01:00
next: #
prev: /plugins/env-plugin-example/
title: consul plugin example
toc: true
weight: 10
---

Here is another simple example plugin that speaks to the consul service endpoint instead of the consul kv-store like the built in consul backend.

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc/jsonrpc"
	"path"
	"strconv"

	"github.com/HeavyHorst/easyKV"
	"github.com/HeavyHorst/remco/backends/plugin"
	consul "github.com/hashicorp/consul/api"
	"github.com/natefinch/pie"
)

func NewConsulClient(addr string) (*consul.Client, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	c, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type ConsulRPCServer struct {
	client *consul.Client
}

func main() {
	p := pie.NewProvider()
	if err := p.RegisterName("Plugin", &ConsulRPCServer{}); err != nil {
		log.Fatalf("failed to register Plugin: %s", err)
	}
	p.ServeCodec(jsonrpc.NewServerCodec)
}

func (c *ConsulRPCServer) Init(args map[string]string, resp *bool) error {
	var err error
	if addr, ok := args["addr"]; ok {
		c.client, err = NewConsulClient(addr)
		if err != nil {
			return err
		}
		*resp = true
		return nil
	}
	return fmt.Errorf("I need an Address !")
}

func (c *ConsulRPCServer) GetValues(args []string, resp *map[string]string) error {
	r := make(map[string]string)
	passingOnly := true
	for _, v := range args {
		addrs, _, err := c.client.Health().Service(v, "", passingOnly, nil)
		if len(addrs) == 0 && err == nil {
			log.Printf("service ( %s ) was not found", v)
		}
		if err != nil {
			return err
		}

		for idx, addr := range addrs {
			key := path.Join("/", "_consul", "service", addr.Service.Service, strconv.Itoa(idx))
			service_json, _ := json.Marshal(addr)
			r[key] = string(service_json)
		}
	}
	*resp = r
	return nil
}

func (c *ConsulRPCServer) Close(args interface{}, resp *interface{}) error {
	// consul client doesn't need to be closed
	return nil
}

func (c *ConsulRPCServer) WatchPrefix(args plugin.WatchConfig, resp *uint64) error {
	return easyKV.ErrWatchNotSupported
}
```
The config backend section could look like this:

```
[backend]
  [[backend.plugin]]
    path = "/etc/remco/plugins/consul-service"
    keys = ["consul"]
    interval = 60
    onetime = false
    [backend.plugin.config]
	 addr = "localhost:8500"
```

