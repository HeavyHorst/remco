# Templates

Templates are written in flosch's [`pongo2`](https://github.com/flosch/pongo2) template engine.

> For a documentation on how the templating language works you can [head over to the Django documentation](https://docs.djangoproject.com/en/dev/topics/templates/). pongo2 aims to be compatible with it.

## Template Functions
### exists
Checks if the key exists. Return false if key is not found.
```
{% if exists("/key") %}
    value: {{ getv ("/key") }}
{% endif %}
```

### get
Returns the KVPair where key matches its argument.
```
{% with get("/key") as dat %}
    key: {{dat.Key}}
    value: {{dat.Value}}
{% endwith %}
```

### gets
Returns all KVPair, []KVPair, where key matches its argument.
```
{% for i in gets("/*") %}
    key: {{i.Key}}
    value: {{i.Value}}
{% endfor %}
```

### getv
Returns the value as a string where key matches its argument or an optional default value.
```
value: {{ getv("/key") }}
```

#### With a default value
```
value: {{ getv("/key", "default_value") }}
```

### getvs
Returns all values, []string, where key matches its argument.
```
{% for value in getvs("/*") %}
    value: {{value}}
{% endfor %}
```

### getenv
Retrieves the value of the environment variable named by the key. It returns the value, which will be empty if the variable is not present. Optionally, you can give a default value that will be returned if the key is not present.
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

### ls
Returns all subkeys, []string, where path matches its argument. Returns an empty list if path is not found.
```
{% for i in ls("/deis/services") %}
   value: {{i}}
{% endfor %}
```

### lsdir
Returns all subkeys, []string, where path matches its argument. It only returns subkeys that also have subkeys. Returns an empty list if path is not found.
```
{% for dir in lsdir("/deis/services") %}
   value: {{dir}}
{% endfor %}
```

### replace
Alias for the [strings.Replace](https://golang.org/pkg/strings/#Replace) function.
```
backend = {{ replace(getv("/services/backend/nginx"), "-", "_", -1) }}
```

### contains
Alias for the [strings.Contains](https://golang.org/pkg/strings/#Contains) function.
```
{% if contains(getv("/services/backend/nginx"), "something") %}
something
{% endif %}
```

### printf
Alias for the [fmt.Sprintf](https://golang.org/pkg/fmt/#Sprintf) function.
```
{{ getv (printf ("/config/%s/host_port", dir)) }}
```

## Template Filters
### base64
Encodes a string as base64
```
{{ "somestring" | base64}}
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

### parseJson
Returns an map[string]interface{} of the json value.

### parseJsonArray
Returns a []interface{} from a json array.

### toJson
Converts data, for example the result of gets or lsdir, into an JSON object.
```
{{ gets("/myapp/database/*") | toJson}}
```

### toPrettyJson
Converts data, for example the result of gets or lsdir, into an pretty-printed JSON object, indented by four spaces.
```
{{ gets("/myapp/database/*") | toPrettyJson}}
```

### sortByLength
Returns the sorted array. 
Works with []string and []KVPair.
```
{% for dir in lsdir("/config") | sortByLength %}
{{dir}}
{% endfor %}
```

### reverse
Returns the reversed array. 
Works with []string and []KVPair.
```
{% for dir in lsdir("/config") | sortByLength | reverse %}
{{dir}}
{% endfor %}
```
