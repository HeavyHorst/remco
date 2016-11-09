---
date: 2016-10-05T17:24:57+02:00
type: index
title: Configuring remco
---

The configuration file is in TOML format.<br>
TOML looks very similar to INI configuration formats, but with slightly more rich data structures and nesting support.

## Using Environment Variables
If you wish to use environmental variables in your config files as a way
to configure values, you can simply use $VARIABLE_NAME or ${VARIABLE_NAME} and the text will be replaced with the value of the environmental variable VARIABLE_NAME.

## Global configuration options
 - **log_level(string):** 
   - Valid levels are panic, fatal, error, warn, info and debug. Default is info.
 - **log_format(string):** 
   - The format of the log messages. Valid formats are *text* and *json*.
 - **include_dir(string):**
   - Specify an entire directory of resource configuration files to include.

## Resource configuration options
 - **name(string, optional):**
    - You can give the resource a name which is added to the logs as field *resource*. Default is the name of the resource file.

## Template configuration options
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
 - **UID(int, optional):**
    - The UID that should own the file. Defaults to the effective uid.
 - **GID(int, optional):**
    - The GID that should own the file. Defaults to the effective gid.

## Backend configuration options

<details>
<summary> **valid in every backend** </summary>

 - **watch(bool, optional):**
   - Enable watch support. Default is false.
 - **prefix(string):**
   - Key path prefix. Default is "".
 - **interval(int):**
   - The backend polling interval. Can be used as a reconcilation loop for watch or standalone.
 - **onetime(bool, optional):**
   - Render the config file and quit. Default is false.
 - **keys([]string):**
   - The backend keys that the template requires to be rendered correctly. The child keys are also loaded.
</details>

<details>
<summary> **etcd** </summary>

 - **nodes([]string):**
   - List of backend nodes.
 - **client_cert(string, optional):**
   - The client cert file.
 - **client_key(string, optional):**
   - The client key file.
 - **client_ca_keys(string, optional):**
   - The client CA key file.
 - **username(string, optional):**
   - The username for the basic_auth authentication.
 - **password(string, optional):**
   - The password for the basic_auth authentication.
 - **version(uint, optional):**
   - The etcd api-level to use (2 or 3). Default is 2.
</details>

<details>
<summary> **consul** </summary>

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
</details>

<details>
<summary> **file** </summary>

 - **filepath(string):**
   - The filepath to a yaml or json file containing the key-value pairs.
</details>

<details>
<summary> **redis** </summary>

 - **nodes([]string):**
   - List of backend nodes.
 - **password(string, optional):**
   - The redis password.
</details>

<details>
<summary> **vault** </summary>

 - **node(string):**
    - The backend node.
 - **auth_type(string):**
   - The vault authentication type. (token, approle, app-id, userpass, github)
 - **auth_token(string):**
   - The vault authentication token. Only used with auth_type=token or github.
 - **app_role(string):**
   - The vault app role. Only used with auth_type=approle.
 - **secret_id(string):**
   - The vault secret id. Only used with auth_type=approle.
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
</details>


<details>
<summary> **env** </summary>
</details>

<details>
<summary> **zookeeper** </summary>

 - **nodes([]string):**
   - List of backend nodes.
</details>
