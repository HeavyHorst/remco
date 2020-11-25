---
date: 2016-12-03T14:59:49+01:00
next: /config/sample-resource/
prev: /config/configuration-options/
title: sample config file
toc: true
weight: 15
---

```
#remco.toml
################################################################
# Global configuration
################################################################
log_level   = "debug"
log_format  = "json"
include_dir = "/etc/remco/resource.d/"
pid_file    = "/var/run/remco/remco.pid"
log_file    = "/var/log/remco.log"

# default backend configurations.
# these settings can be overwritten in the individual resource backend settings.
[default_backends]
[default_backends.file]
    onetime  = true
    prefix   = "/bla"

################################################################
# Resource configuration
################################################################
[[resource]]
  name = "haproxy"
  start_cmd   = "echo 1"
  reload_cmd  = "echo 1"
  [[resource.template]]
    src         = "/etc/remco/templates/haproxy.cfg"
    dst         = "/etc/haproxy/haproxy.cfg"
    check_cmd   = "somecommand"
    reload_cmd  = "somecommand"
    mode        = "0644"

  [resource.backend]
    # you can use as many backends as you like
	  # in this example vault and file
    [resource.backend.vault]
      node           = "http://127.0.0.1:8200"
      ## Token based auth backend
      auth_type      = "token"
      auth_token     = "vault_token"
      ## AppID based auth backend
      # auth_type    = "app-id"
      # app_id       = "vault_app_id"
      # user_id      = "vault_user_id"
      ## userpass based auth backend
      # auth_type    = "userpass"
      # username     = "username"
      # password     = "password"
      client_cert    = "/path/to/client_cert"
      client_key     = "/path/to/client_key"
      client_ca_keys = "/path/to/client_ca_keys"
      
	  # These values are valid in every backend
      watch    = true
      prefix   = "/"
      onetime  = true
      interval = 1
      keys     = ["/"]
      watchKeys = ["/haproxy/reload"]

    [resource.backend.file]
      httpheader = { X-Test-Token = "XXX", X-Test-Token2 = "YYY" } 
      filepath = "/etc/remco/test.yml"
	    watch    = true
	    keys     = ["/prefix"]

################################################################
# Telemetry configuration
################################################################
[telemetry]
  enabled = true
  [telemetry.sinks.prometheus]
    addr = ":2112"
    expiration = 600
```      

