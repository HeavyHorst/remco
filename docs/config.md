#Configuring remco

The configuration file is in TOML format. TOML looks very similar to INI configuration formats, but with slightly more rich data structures and nesting support.

##Global configuration options
 - **log_level(string):** 
   - Valid levels are panic, fatal, error, warn, info and debug. Default is info.
 - **log_format(string):** 
   - The format of the log messages. Valid formats are *text* and *json*.

##Using Environment Variables
If you wish to use environmental variables in your config files as a way
to configure values, you can simply use $VARIABLE_NAME or ${VARIABLE_NAME} and the text will be replaced with the value of the environmental variable VARIABLE_NAME.

##Backend config options
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
