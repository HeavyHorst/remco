---
date: 2016-12-03T14:56:09+01:00
next: #
prev: /examples/
title: Dynamic haproxy configuration with docker, registrator and etcd
toc: true
weight: 5
---

## The haproxy template

We expect [registrator](http://gliderlabs.github.io/registrator/latest/) to write the service data in this format to etcd: 

    /services/<service-name>/<service-id> = <ip>:<port>

The scheme (tcp, http) and the host_port of the service is configurable over the following keys:

    /config/<service-name>/scheme
    /config/<service-name>/host_port

We create the template for the haproxy configuration file first.
Create the file **haproxy.tmpl** and add the following config blocks:

Some default configuration parameters:
```
global
    daemon
    maxconn 2048

defaults
    timeout connect 5000ms
    timeout client 500000ms
    timeout server 500000ms
    log global

frontend name_resolver_http
    bind *:80
    mode http
```


<hr>

This block creates the *http* acl's.
We itarate over all directories under /config  (the services) and create a **url_beg** and a **hdr_beg(host)** acl if the service has a scheme configured. 
Note that we sort the services by length and iterate in reversed order (longest services first). That way we can have services with the same prefix, for example redis_test, and redis.

```
{% for dir in lsdir("/config") | sortByLength reversed %}
  {% if exists(printf("/config/%s/scheme", dir)) %}
    {% if getv(printf("/config/%s/scheme", dir)) == "http" %}
      {% if ls(printf("/services/%s", dir)) %}
        acl is_{{ dir }} url_beg /{{ dir }}
        acl is_{{ dir }} hdr_beg(host) {{ dir }}
        use_backend {{ dir }}_servers if is_{{ dir }}
      {% endif %}
    {% endif %}
  {% endif %}
{% endfor %}
```

If we had one service named redis we would get:

```
acl is_redis url_beg /redis 
acl is_redis hdr_beg(host) redis
use_backend redis_servers if is_redis
```

<hr>

Optional template block to expose a service on an host port:
We iterate over all services under /config, test if the scheme and host_port is configured and create the host port configuration.


```
{% for dir in lsdir("/config") %}
  {% if exists (printf ("/config/%s/scheme", dir )) %}
    {% if exists (printf("/config/%s/host_port", dir )) %}
      {% if ls(printf("/services/%s", dir)) %}
                frontend {{ dir }}_port
                mode {{ getv (printf("/config/%s/scheme", dir)) }}
                bind *:{{ getv (printf ("/config/%s/host_port", dir)) }}
                default_backend {{ dir }}_servers
      {% endif %}
    {% endif %}
  {% endif %}
{% endfor %}
```

If we had one service named redis with scheme=tcp and host_port=6379 we would get:

```
frontend redis_port
mode tcp
bind *:6379 
default_backend redis_servers
```

<hr>

The last block creates the haproxy backends.
We iterate over all services and, if a scheme is set, create the backend *{service_name}_servers*.

```
{% for dir in lsdir("/services") %}
  {% if exists(printf("/config/%s/scheme", dir)) %}
backend {{ dir }}_servers
        mode {{ getv (printf ("/config/%s/scheme", dir)) }}
        {% for i in gets (printf("/services/%s/*", dir)) %}
            server server_{{ dir }}_{{ base (i.Key) }} {{ i.Value }}
        {% endfor %}
    {% endif %}
{% endfor %}
```

If we had one service named redis with scheme=tcp we could get for example:

```
backend redis_servers   
  mode tcp
    server server_redis_1 192.168.0.10:32012
    server server_redis_2 192.168.0.10:35013
```

<hr>

## The remco configuration file

We also need to create the remco configuration file.
Create a file named **config** and insert the following toml configuration.

```toml
################################################################
# Global configuration
################################################################
log_level = "debug"
log_format = "text"

[[resource]]
name = "haproxy"

[[resource.template]]
  src = "/etc/remco/templates/haproxy.tmpl"
  dst = "/etc/haproxy/haproxy.cfg"
  reload_cmd 	  = "haproxy -f /etc/haproxy/haproxy.cfg -p /var/run/haproxy.pid -D -sf `cat /var/run/haproxy.pid`"

  [resource.backend]
    [resource.backend.etcd]
      nodes = ["${ETCD_NODE}"]
      keys = ["/services", "/config"]
      watch = true
      interval = 60
```

## The Dockerfile

```
FROM alpine:3.4

ENV REMCO_VER 0.8.0

RUN apk --update add --no-cache haproxy bash ca-certificates
RUN wget https://github.com/HeavyHorst/remco/releases/download/v${REMCO_VER}/remco_${REMCO_VER}_linux_amd64.zip && \
    unzip remco_${REMCO_VER}_linux_amd64.zip && rm remco_${REMCO_VER}_linux_amd64.zip && \
    mv remco_linux /bin/remco

COPY config /etc/remco/config
COPY haproxy.tmpl /etc/remco/templates/haproxy.tmpl

ENTRYPOINT ["remco"]
```

## Build and Run the container

You should have three files at this point:

```
.
├── config
├── Dockerfile
└── haproxy.tmpl
```

### Build the docker container:

```bash
sudo docker build -t remcohaproxy .
```

### Optionally test the container:

#### put some data into etcd:

```bash
etcdctl set /services/exampleService/1 someip:port
etcdctl set /config/exampleService/scheme http
etcdctl set /config/exampleService/host_port 1234
```


In this example we connect to a local etcd cluster.

```bash
sudo docker run --rm -ti --net=host -e ETCD_NODE=http://localhost:2379 remcohaproxy
```

You should see something like this:

```
[Dec 16 18:26:20]  INFO remco[1]: Target config out of sync config=/etc/haproxy/haproxy.cfg resource=haproxy source=resource.go:66
[Dec 16 18:26:20] DEBUG remco[1]: Overwriting target config config=/etc/haproxy/haproxy.cfg resource=haproxy source=resource.go:66
[Dec 16 18:26:20] DEBUG remco[1]: Running haproxy -f /etc/haproxy/haproxy.cfg -p /var/run/haproxy.pid -D -sf `cat /var/run/haproxy.pid` resource=haproxy source=resource.go:66
[Dec 16 18:26:20] DEBUG remco[1]: "" resource=haproxy source=resource.go:66
[Dec 16 18:26:20]  INFO remco[1]: Target config has been updated config=/etc/haproxy/haproxy.cfg resource=haproxy source=resource.go:66
[Dec 16 18:26:20] DEBUG remco[1]: [Reaped child process 60] source=main.go:87
[Dec 16 18:26:24] DEBUG remco[1]: Retrieving keys backend=etcd key_prefix= resource=haproxy source=resource.go:66
[Dec 16 18:26:24] DEBUG remco[1]: Compiling source template resource=haproxy source=resource.go:66 template=/etc/remco/templates/haproxy.tmpl
[Dec 16 18:26:24] DEBUG remco[1]: Comparing staged and dest config files dest=/etc/haproxy/haproxy.cfg resource=haproxy source=resource.go:66 staged=.haproxy.cfg389124299
[Dec 16 18:26:24] DEBUG remco[1]: Target config in sync config=/etc/haproxy/haproxy.cfg resource=haproxy source=resource.go:66
```

### Run registrator

```
sudo docker run -d \
    --name=registrator \
    --net=host \
    --volume=/var/run/docker.sock:/tmp/docker.sock \
    gliderlabs/registrator:latest \
      etcd://localhost:2379/services
```

Now every container get automatically registered under /services.
You can then configure the scheme and optionally the host_port of each service that you want to expose.
