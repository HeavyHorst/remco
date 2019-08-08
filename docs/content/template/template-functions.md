---
date: 2016-12-03T15:06:33+01:00
next: /template/template-filters/
prev: /template/
title: template functions
toc: true
weight: 5
---

<details>
<summary> **exists** -- Checks if the key exists. Return false if key is not found.</summary>
```
{% if exists("/key") %}
    value: {{ getv ("/key") }}
{% endif %}
```
</details>

<details>
<summary> **get** -- Returns the KVPair where key matches its argument.</summary>
```
{% with get("/key") as dat %}
    key: {{dat.Key}}
    value: {{dat.Value}}
{% endwith %}
```
</details>

<details>
<summary> **gets** -- Returns all KVPair, []KVPair, where key matches its argument.</summary>
```
{% for i in gets("/*") %}
    key: {{i.Key}}
    value: {{i.Value}}
{% endfor %}
```
</details>

<details>
<summary> **getv** -- Returns the value as a string where key matches its argument or an optional default value.</summary>
```
value: {{ getv("/key") }}
```
#### With a default value
```
value: {{ getv("/key", "default_value") }}
```
</details>

<details>
<summary> **getvs** -- Returns all values, []string, where key matches its argument.</summary>
```
{% for value in getvs("/*") %}
    value: {{value}}
{% endfor %}
```
</details>

<details>
<summary> **getenv** -- Retrieves the value of the environment variable named by the key. It returns the value, which will be empty if the variable is not present. Optionally, you can give a default value that will be returned if the key is not present. </summary>
```
export HOSTNAME=`hostname`
```
```
hostname: {{getenv("HOSTNAME")}}
```
#### With a default value
```
ipaddr: {{ getenv("HOST_IP", "127.0.0.1") }}
```
</details>

<details>
<summary> **ls** -- Returns all subkeys, []string, where path matches its argument. Returns an empty list if path is not found. </summary>
```
{% for i in ls("/deis/services") %}
   value: {{i}}
{% endfor %}
```
</details>

<details>
<summary> **lsdir** -- Returns all subkeys, []string, where path matches its argument. It only returns subkeys that also have subkeys. Returns an empty list if path is not found. </summary>
```
{% for dir in lsdir("/deis/services") %}
   value: {{dir}}
{% endfor %}
```
</details>

<details>
<summary> **replace** -- Alias for the [strings.Replace](https://golang.org/pkg/strings/#Replace) function. </summary>
```
backend = {{ replace(getv("/services/backend/nginx"), "-", "_", -1) }}
```
</details>

<details>
<summary> **contains** -- Alias for the [strings.Contains](https://golang.org/pkg/strings/#Contains) function. </summary>
```
{% if contains(getv("/services/backend/nginx"), "something") %}
something
{% endif %}
```
</details>

<details>
<summary> **printf** -- Alias for the [fmt.Sprintf](https://golang.org/pkg/fmt/#Sprintf) function. </summary>
```
{{ getv (printf ("/config/%s/host_port", dir)) }}
```
</details>

<details>
<summary> **unixTS** -- Wrapper for [time.Now().Unix()](https://golang.org/pkg/time/#Unix). </summary>
```
{{ unixTS }}
```
</details>

<details>
<summary> **dateRFC3339** -- Wrapper for [time.Now().Format(time.RFC3339)](https://golang.org/pkg/time/). </summary>
```
{{ dateRFC3339 }}
```
</details>

<details>
<summary> **lookupIP** -- Wrapper for the [net.LookupIP](https://golang.org/pkg/net/#LookupIP) function. The wrapper returns the IP addresses in alphabetical order. </summary>
```
{% for ip in lookupIP("kube-master") %}
 {{ ip }}
{% endfor %}
```
</details>

<details>
<summary> **lookupSRV** -- Wrapper for the [net.LookupSRV](https://golang.org/pkg/net/#LookupSRV) function. The wrapper returns the SRV records in alphabetical order. </summary>
```
{% for srv in lookupSRV("xmpp-server", "tcp", "google.com") %}
  target: {{ srv.Target }}
  port: {{ srv.Port }}
  priority: {{ srv.Priority }}
  weight: {{ srv.Weight }}
{% endfor %}
```
</details>
