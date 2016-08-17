#Configuring remco
The configuration file is in TOML format. TOML looks very similar to INI configuration formats, but with slightly more rich data structures and nesting support.

##Using Environment Variables
If you wish to use environmental variables in your config files as a way
to configure values, you can simply use $VARIABLE_NAME or ${VARIABLE_NAME} and the text will be replaced with the value of the environmental variable VARIABLE_NAME.

##Global configuration options
 - **log_level(string):** 
   - Valid levels are panic, fatal, error, warn, info and debug. Default is info.
 - **log_format(string):** 
   - The format of the log messages. Valid formats are *text* and *json*.

##Template configuration options
 - **src(string):**
    - The path of the template that will be used to render the application's configuration file.
 - **dst(string):**
    - The location to place the rendered configuration file.
 - **checkCmd(string, optional):**
    - The command to check config. Use {{.src}} to reference the rendered source template.
 - **reloadCmd(string, optional):**
    - The command to reload config.
 - **mode(string, optional):**
    - The permission mode of the file. Default is "0644".

##Backend configuration options

###valid in every backend
 - **watch(bool, optional):**
   - Enable watch support. Default is false.
 - **prefix(string):**
   - Key path prefix. Default is "".
 - **interval(int):**
   - The backend polling interval. Only used when watch mode is disabled.
 - **onetime(bool, optional):**
   - Render the config file and quit. Only used when watch mode is disabled. Default is false.
 - **keys([]string):**
   - The backend keys that the template requires to be rendered correctly. The child keys are also loaded.

###etcd
 - **nodes([]string):**
   - List of backend nodes.
 - **client_cert(string, optional):**
   - The client cert file.
 - **client_key(string, optional):**
   - The client key file.
 - **client_ca_keys(string, optional):**
   - The client CA key file.
 - **basic_auth(bool, optional):**
   - Use Basic Auth to authenticate. Default is false.
 - **username(string, optional):**
   - The username for the basic_auth authentication.
 - **password(string, optional):**
   - The password for the basic_auth authentication.
 - **version(uint, optional):**
   - The etcd api-level to use (2 or 3). Default is 2.

###consul
 - **nodes([]string):**
    - List of backend nodes.
 - **scheme(string):**
    - the backend URI scheme (http or https).
 - **client_cert(string, optional):**
   - The client cert file.
 - **client_key(string, optional):**
   - The client key file.
 - **client_ca_keys(string, optional):**
   - The client CA key file.

###file
 - **filepath(string):**
   - The filepath to a yaml or json file containing the key-value pairs.

###redis
 - **nodes([]string):**
   - List of backend nodes.
 - **password(string, optional):**
   - The redis password.

###vault
 - **node(string):**
    - The backend node.
 - **auth_type(string):**
   - The vault authentication type. (token, app-id, userpass)
 - **auth_token(string):**
   - The vault authentication token. Only used with auth_type=token.
 - **app_id(string):**
   - The vault app ID. Only used with auth_type=app-id.
 - **user_id(string):**
   - The vault user ID. Only used with auth_type=app-id.
 - **username(string):**
   - The username for the userpass authentication.
 - **password(string):**
   - The password for the userpass authentication.
 - **client_cert(string, optional):**
   - The client cert file.
 - **client_key(string, optional):**
   - The client key file.
 - **client_ca_keys(string, optional):**
   - The client CA key file.

#Example
```TOML
log_level = "debug"
log_format = "text"

[[resource]]
  [[resource.template]]
    src = "/path/to/template"
    dst = "/path/to/destionation/file"
    checkCmd = ""
    reloadCmd = ""
    mode = "0644"
    
    [resource.backend.etcdconfig]
      nodes = ["127.0.0.1:2379"]
      version = 3
      watch = true
      prefix = "/production"
      keys = ["/some_key"]
      
    [resource.backend.vaultconfig]
      node = "http://127.0.0.1:8200"
      auth_type = "token"
      auth_token = "vault_token"
      client_cert = "/path/to/client_cert"
      client_key = "/path/to/client_key"
      client_ca_keys = "/path/to/client_ca_keys"
      interval = 60
      prefix = "/production"
      keys = ["/some_secret_key"]
```