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
<summary> **parseInt** -- Takes the given string and parses it as a base-10 integer (64bit) </summary>

```
{{ "12000" | parseInt }}
```
</details>

<details>
<summary> **parseFloat** -- Takes the given string and parses it as a float64 </summary>

```
{{ "12000.45" | parseFloat }}
```
</details>

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
<summary> **mapValue** -- Returns an map element by key  </summary>

```
{{ getv("/some_yaml_config") | parseYAML | mapValue:"key" }}
```
</details>

<details>
<summary> **index** -- Returns an array element by index  </summary>

```
{{ "/home/user/test" | split:"/" | index:"1" }}
```
</details>

<details>
<summary> **parseYAML** -- Returns an interface{} of the yaml value.</summary>

```
{% for value in getvs("/cache1/domains/*") %}
{% set data = value | parseYAML %}
{{ data.type }} {{ data.name }} {{ data.addr }}
{% endfor %}
```
</details>

<details>
<summary> **parseJSON** -- Returns an interface{} of the json value.</summary>

```
{% for value in getvs("/cache1/domains/*") %}
{% set data = value | parseJSON %}
{{ data.type }} {{ data.name }} {{ data.addr }}
{% endfor %}
```
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
{{ gets("/myapp/database/*") | toYAML}}
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

## Custom filters

It is possible to create custom filters in JavaScript.
If you want to create a 'toEnv' filter, which transforms file system paths to environment variables, you must create the file 'toEnv.js' in the configurable filter directory.

The filter code could look like:
```javascript
In.split("/").join("_").substr(1).toUpperCase();
```

There are two predefined variables:

  - In: the filter input (string)
  - Param: the optional filter parameter (string)

As the parameter one string is possible only, the parameter string is added with a double-colon to the filter name (""yadda" | filter:"paramstr").
When the filter function needs multiple parameter all of them must be put into this one string and parsed inside the filter to extract all 
parameter from this string (example "replace" filter below).

Remark:

* "console" object for logging does not exist, therefore no output (for debugging and similar) possible.
* variable declaration must use "var" as other keywords like "const" or "let" are not defined
* the main script must not use "return" keyword, last output is the filter result.

### Examples
**reverse filter**

Put file `reverse.js` into the configured "filter_dir" with following content: 

```javascript
function reverse(s) {
     var o = "";
     for (var i = s.length - 1; i >= 0; i--)
        o += s[i];
     return o;
}

reverse(In);
```
Call this filter inside your template (e,g, "my-reverse-template.tmpl") with

```
{% set myString = "hip-hip-hooray" %}
myString is {{ myString }}
reversed myString is {{ myString | reverse }}
```

Output is:

```text
myString is hip-hip-hooray
reversed myString is yarooh-pih-pih
```

**replace filter**

Put file `replace.js` into the configured "filter_dir" with following content:

```javascript
function replace(str, p) {
    var params = [' ','_'];  // default: replace all spaces with underscore
    if (p) {
        params = p.split(',');  // split all params given at comma
    }
  // if third param is a "g" like "global" change search string to regexp
    if (params.length > 2 && params[2] == 'g') {
        params[0] = new RegExp(params[0], params[2]);
    }
    // javascript string.replace replaces first occurence only if search param is a string
    // need regexp object to replace all occurences
    return str.replace(params[0], params[1]);
}
replace(In, Param)
```

Use this inside the template as:

```
{% set myString = "hip-hip-hooray" %}
myString is {{ myString }}
replace with default params (spaces): {{ myString | replace }}
only replace first "-" with underscore is {{ myString | replace:"-,_" }}
replace all "-" with underscore is {{ myString | replace:"-,_,g" }}
```

Output is:

```text
myString is hip-hip-hooray
replace with default params (spaces): hip-hip-hooray
only replace first "-" with underscore is hip_hip-hooray
replace all "-" with underscore is hip_hip_hooray
```
