---
date: 2016-12-03T15:07:41+01:00
next: /plugins/
prev: /template/template-functions/
title: template filters
toc: true
weight: 20
---

## Builtin filters

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
<summary> **parseYAML** -- Returns an interface{} of the yaml/json value.</summary>
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


## Custom filters

It is possible to create custom filters in JavaScript.
If you want to create a 'toEnv' filter, which transforms file system paths to environment variables, you must create the file 'toEnv.js' in the configurable filter directory.

The filter code could look like:
```javascript
In.split("/").join("_").substr(1).toUpperCase();
```

There are two predifined variables:

  - In: the filter input
  - Param: the optional filter parameter
 
### Examples
**reverse filter**
 
```javascript
function reverse(s) {
     var o = "";
     for (var i = s.length - 1; i >= 0; i--)
        o += s[i];
     return o;
}

reverse(In);
```
