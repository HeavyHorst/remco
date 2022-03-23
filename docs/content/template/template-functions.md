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

<details>
<summary> **createMap** -- create a hashMap to store values at runtime. This can be useful if you want to generate json/yaml files. </summary>

```
{% set map = createMap() %}
{{ map.Set("Moin", "Hallo2") }}
{{ map.Set("Test", 105) }}
{{ map | toYAML }}

{% set map2 = createMap() %}
{{ map2.Set("Moin", "Hallo") }}
{{ map2.Set("Test", 300) }}
{{ map2.Set("anotherMap", map) }}
{{ map2 | toYAML }}
```

The hashmap supports the following methods:
* `m.Set("key", value)` adds a new value of arbitrary type referenced by "key" to the map
* `m.Get("key")` get the value for the given "key"
* `m.Remove("key")` removes the key and value from the map
</details>

<details>
<summary> **createSet** -- create a Set to store values at runtime. This can be useful if you want to generate json/yaml files. </summary>

```
{% set s = createSet() %}
{{ s.Append("Moin") }}
{{ s.Append("Moin") }}
{{ s.Append("Hallo") }}
{{ s.Append(1) }}
{{ s.Remove("Hallo") }}
{{ s | toYAML }}
```
The set created supports the following methods:
* `s.Append("string")` adds a new string to the set. Attention - the set is not
  sorted or the order of appended elements guaranteed.
* `s.Remove("string")` removes the given element from the set.
* `s.Contains("string")` check if the given string is part of the set, returns
  true or false otherwise
* `s.SortedSet()` return a new list where all elements are sorted in increasing
  order. This method should be used inside the template with a for-in loop to generate
 a stable output file not changing order of elements on every run. 

```
{% set s = createSet() %}
{% s.Append("Moin") %}
{% s.Append("Hi") %}
{% s.Append("Hallo") %}

{% for greeting in s %}
{{ geeting }}
{% endfor %}

{% for greeting in s.SortedSet() %}
{{ geeting }}
{% endfor %}
```

The output of the first loop is not defined, it can be in every order (like `Moin Hallo Hi` or `Hi Hallo Moin` and so on)
The second loop returns every time `Hallo Hi Moin` (items sorted as string in increasing order)
</details>
