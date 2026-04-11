# Template filters

## Builtin filters

### parseInt

Takes the given string and parses it as a base-10 integer (64bit).

```
{{ "12000" | parseInt }}
```

### parseFloat

Takes the given string and parses it as a float64.

```
{{ "12000.45" | parseFloat }}
```

### base64

Encodes a string as base64.

```
{{ "somestring" | base64 }}
```

### base64decode

Decodes a base64-encoded string.

```
{{ "c29tZXN0cmluZw==" | base64decode }}
```

### base

Alias for the [path.Base](https://golang.org/pkg/path/#Base) function.

```
{{ "/home/user/test" | base }}
```

### dir

Alias for the [path.Dir](https://golang.org/pkg/path/#Dir) function.

```
{{ "/home/user/test" | dir }}
```

### split

Alias for the [strings.Split](https://golang.org/pkg/strings/#Split) function.

```
{% for i in ("/home/user/test" | split:"/") %}
{{i}}
{% endfor %}
```

### mapValue

Returns a map element by key.

```
{{ getv("/some_yaml_config") | parseYAML | mapValue:"key" }}
```

### index

Returns an array element by index.

```
{{ "/home/user/test" | split:"/" | index:"1" }}
```

### parseYAML

Returns an interface{} of the yaml value.

```
{% for value in getvs("/cache1/domains/*") %}
{% set data = value | parseYAML %}
{{ data.type }} {{ data.name }} {{ data.addr }}
{% endfor %}
```

### parseJSON

Returns an interface{} of the json value. (`parseYAMLArray` is a deprecated alias.)

```
{% for value in getvs("/cache1/domains/*") %}
{% set data = value | parseJSON %}
{{ data.type }} {{ data.name }} {{ data.addr }}
{% endfor %}
```

### toJSON

Converts data, for example the result of gets or lsdir, into a JSON object.

```
{{ gets("/myapp/database/*") | toJson}}
```

### toPrettyJSON

Converts data, for example the result of gets or lsdir, into a pretty-printed JSON object, indented by four spaces.

```
{{ gets("/myapp/database/*") | toPrettyJson}}
```

### toYAML

Converts data, for example the result of gets or lsdir, into a YAML string. Accepts an optional parameter to control indentation (e.g. `indent=4`).

```
{{ gets("/myapp/database/*") | toYAML }}
```

With custom indentation:

```
{{ gets("/myapp/database/*") | toYAML:"indent=4" }}
```

### sortByLength

Returns the sorted array. Works with []string and []KVPair.

```
{% for dir in lsdir("/config") | sortByLength %}
{{dir}}
{% endfor %}
```

## Custom filters

It is possible to create custom filters in JavaScript.
If you want to create a `toEnv` filter, which transforms file system paths to environment variables, you must create the file `toEnv.js` in the configurable filter directory.

The filter code could look like:

```javascript
In.split("/").join("_").substr(1).toUpperCase();
```

There are two predefined variables:

- **In:** the filter input (string)
- **Param:** the optional filter parameter (string)

As the parameter one string is possible only, the parameter string is added with a double-colon to the filter name (`"yadda" | filter:"paramstr"`).
When the filter function needs multiple parameter all of them must be put into this one string and parsed inside the filter to extract all parameter from this string (example "replace" filter below).

Remark:

- `console` object for logging does not exist, therefore no output (for debugging and similar) possible.
- variable declaration must use `var` as other keywords like `const` or `let` are not defined
- the main script must not use `return` keyword, last output is the filter result.

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

Call this filter inside your template (e.g. `my-reverse-template.tmpl`) with:

```
{% set myString = "hip-hip-hooray" %}
myString is {{ myString }}
reversed myString is {{ myString | reverse }}
```

Output is:

```
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

```
myString is hip-hip-hooray
replace with default params (spaces): hip-hip-hooray
only replace first "-" with underscore is hip_hip-hooray
replace all "-" with underscore is hip_hip_hooray
```
