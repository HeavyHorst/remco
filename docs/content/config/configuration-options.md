---
date: 2016-12-03T14:57:13+01:00
next: /config/sample-config/
prev: /config/environment-variables/
title: configuration options
toc: true
weight: 10
---


## Global configuration options
 - **log_level(string):** 
   - Valid levels are panic, fatal, error, warn, info and debug. Default is info.
 - **log_format(string):** 
   - The format of the log messages. Valid formats are *text* and *json*.
 - **include_dir(string):**
   - Specify an entire directory of resource configuration files to include. Data from files will be imported directly into `resource` array.
 - **filter_dir(string):**
   - A folder with custom JavaScript template filters.
 - **pid_file(string):**
   - A filename to write the process-id to.
 - **log_file(string):**
   - Specify the log file name. The empty string means to log to stdout.

## Resource configuration options
 - **name(string, optional):**
    - You can give the resource a name which is added to the logs as field *resource*. Default is the name of the resource file.
 - **start_cmd(string, optional)**
    - An optional command which is executed once all templates have been processed successfully.
 - **reload_cmd(string, optional)**
    - An optional command which is executed as soon as a template belonging to the resource has been successfully recreated.

## Exec configuration options
 - **command(string):**
   - This is the command to exec as a child process. Note that the child process must remain in the foreground.
 - **kill_signal(string):**
   - This defines the signal sent to the child process when remco is gracefully shutting down. The application needs to exit before the `kill_timeout`,
     it will be terminated otherwise (like kill -9). The default value is "SIGTERM".
 - **kill_timeout(int):**
   - the maximum amount of time (seconds) to wait for the child process to gracefully terminate. Default is 10.
 - **reload_signal(string):**
   - This defines the signal sent to the child process when some configuration data is changed. If no signal is specified the child process will be killed (gracefully) and started again.
 - **splay(int):**
   - A random splay to wait before killing the command. May be useful in large clusters to prevent all child processes to reload at the same time when configuration changes occur. Default is 0.

## Template configuration options
 - **src(string):**
    - The path of the template that will be used to render the application's configuration file.
 - **dst(string):**
    - The location to place the rendered configuration file.
 - **make_directories(bool, optional):**
    - make parent directories for the dst path as needed. Default is false.
 - **check_cmd(string, optional):**
    - An optional command to check the rendered source template before writing it to the destination. If this command returns non-zero, the destination will not be overwritten by the rendered source template. We can use `{{.src}}` here to reference the rendered source template.
 - **reload_cmd(string, optional):**
    - An optional command to run after the destination is updated. We can use `{{.dst}}` here to reference the destination.
 - **mode(string, optional):**
    - The permission mode of the file. Default is "0644".
 - **UID(int, optional):**
    - The UID that should own the file. Defaults to the effective uid.
 - **GID(int, optional):**
    - The GID that should own the file. Defaults to the effective gid.

## Backend configuration options

See the example configuration to see how global default values can be set for individual backends.

<details>
<summary> **valid in every backend** </summary>

 - **keys([]string):**
   - The backend keys that the template requires to be rendered correctly. The child keys are also loaded.
 - **watch(bool, optional):**
   - Enable watch support. Default is false.
 - **prefix(string, optional):**
   - Key path prefix. Default is "".
 - **interval(int, optional):**
   - The backend polling interval. Can be used as a reconciliation loop for watch or standalone.
 - **onetime(bool, optional):**
   - Render the config file and quit. Default is false.
</details>

<details>
<summary> **etcd** </summary>

 - **nodes([]string):**
   - List of backend nodes.
 - **srv_record(string, optional):**
   - A DNS server record to discover the etcd nodes.
 - **scheme(string, optional):**
   - The backend URI scheme (http or https). This is only used when the nodes are discovered via DNS srv records and the api level is 2. Default is http.
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
 - **srv_record(string, optional):**
   - A DNS server record to discover the consul nodes.
 - **scheme(string):**
   - The backend URI scheme (http or https).
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
   - The filepath to a yaml or json file containing the key-value pairs. This can be a local file or a remote http/https location.
 - **httpheaders(map[string]string):**
   - Optional HTTP-headers to append to the request if the file path is a remote http/https location. 
</details>

<details>
<summary> **redis** </summary>

 - **nodes([]string):**
   - List of backend nodes.
 - **srv_record(string), optional:**
   - A DNS server record to discover the redis nodes.
 - **password(string, optional):**
   - The redis password.
 - **database(int, optional):**
   - The redis database.
</details>

<details>
<summary> **vault** </summary>

 - **node(string):**
   - The backend node.
 - **auth_type(string):**
   - The vault authentication type. (token, approle, app-id, userpass, github, cert, kubernetes)
 - **auth_token(string):**
   - The vault authentication token. Only used with auth_type=token or github.
 - **role_id(string):**
   - The vault app role. Only used with auth_type=approle and kubernetes.
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
 - **srv_record(string, optional):**
   - A DNS server record to discover the zookeeper nodes.
</details>

## Telemetry configuration options
 - **enabled(bool):**
   - Flag to enable telemetry.
 - **service_name(string):**
   - Service name to add to every metric name. "remco" by default

## Sink configuration options

<details>
<summary> **inmem** </summary>

 - **interval(int):**
   - How long is each aggregation interval (seconds).
 - **retain(int):**
   - Retain controls how many metrics interval we keep.
</details>

<details>
<summary> **prometheus** </summary>

 - **addr(string):**
   - Address to expose metrics on. Prometheus stats will be available at /metrics endpoint.
 - **expiration(int):**
   - Expiration is the duration a metric is valid for, after which it will be untracked. If the value is zero, a metric is never expired.
</details>

<details>
<summary> **statsd** </summary>

 - **addr(string):**
   - Statsd/Statsite server address
</details>

<details>
<summary> **statsite** </summary>

  - **addr(string):**
   - Statsd/Statsite server address
 </details>
