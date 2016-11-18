---
date: 2016-10-05T17:33:14+02:00
title: Templates
---

Templates are written in flosch's [`pongo2`](https://github.com/flosch/pongo2) template engine.

> For a documentation on how the templating language works you can [head over to the Django documentation](https://docs.djangoproject.com/en/dev/topics/templates/). pongo2 aims to be compatible with it.

## Template Functions

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

## Template Filters

<details>
<summary> **base64** -- Encodes a string as base64 </summary>
```
{{ "somestring" | base64}}
```
</details>

<details>
<summary> **base** -- Alias for the [path.Base](https://golang.org/pkg/path/#Base) function. </summary>
```
{{ "/home/user/test" | base }}
```
</details>

<details>
<summary> **dir** -- Alias for the [path.Dir](https://golang.org/pkg/path/#Dir) function. </summary>
```
{{ "/home/user/test" | dir }}
```
</details>

<details>
<summary> **split** -- Alias for the [strings.Split](https://golang.org/pkg/strings/#Split) function. </summary>
```
{% for i in ("/home/user/test" | split:"/") %}
{{i}}
{% endfor %}
```
</details>

<details>
<summary> **parseJSON** -- Returns an map[string]interface{} of the json value.</summary>
</details>

<details>
<summary> **parseJSONArray** -- Returns a []interface{} from a json array. </summary>
</details>

<details>
<summary> **toJSON** -- Converts data, for example the result of gets or lsdir, into an JSON object. </summary>
```
{{ gets("/myapp/database/*") | toJson}}
```
</details>

<details>
<summary> **toPrettyJSON** -- Converts data, for example the result of gets or lsdir, into an pretty-printed JSON object, indented by four spaces. </summary>
```
{{ gets("/myapp/database/*") | toPrettyJson}}
```
</details>

<details>
<summary> **toYAML** -- Converts data, for example the result of gets or lsdir, into a YAML string. </summary>
```
{{ gets("/myapp/database/*") | toJson}}
```
</details>

<details>
<summary> **sortByLength** - Returns the sorted array. </summary>

Works with []string and []KVPair.
```
{% for dir in lsdir("/config") | sortByLength %}
{{dir}}
{% endfor %}
```
</details>

<details>
<summary> **reverse** -- Returns the reversed array. </summary>

Works with []string and []KVPair.
```
{% for dir in lsdir("/config") | sortByLength | reverse %}
{{dir}}
{% endfor %}
```
</details>

<details>
<summary> **decrypt** -- Decrypts the stored data. Data must follow the following format, `base64(gpg(gzip(data)))`. </summary>

This is compatible with [crypt](https://github.com/xordataexchange/crypt/tree/master/bin/crypt).

Works with string, []string, KVPair, KVPairs

```
{{ getv("/test/data") | decrypt:"/path/to/your/armored/private/key" }}
```

#### Storing data using gpg
```
data = `echo 'secret text' | gzip -c | gpg2 --compress-level 0 --encrypt --default-recipient <your-recipient> | base64`
ETCDCTL_API=3 etcdctl put /test/data $data
```
</details>
