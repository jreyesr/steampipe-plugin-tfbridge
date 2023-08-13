---
name: Bug report
about: Report a Terraform provider that doesn't work
title: ''
labels: bug
assignees: ''

---

_replace anything in this template with your actual values, most sections have placeholders/examples_

### Terraform backing provider

**Name:** _the provider vendor/name, as you'd write in the Terraform file_
**Version:** _the provider version, as written in the SPC file_

###  SPC file

_(remember to remove any confidential tokens from the SPC file, and replace them with dummy data)_

```hcl
connection "tfbridge" {
  plugin = "jreyesr/tfbridge"

  provider = "???"
  version = "???"

  provider_config = <<EOT
    token = "token here"
  EOT
}
```

### When does the error happen?

- [ ] When loading the plugin (i.e. while Steampipe is booting)
- [ ] When `.inspect`ing the tables provided by the plugin
- [x] When running a SQL query that involves the plugin

### SQL query

_(if the error appears when running a SQL query)_

```sql
select * from github_users where usernames='[]'
```

### Error

```
type {{{} {{{} 83}}}} can't be handled by quals
```

![image](https://github.com/jreyesr/steampipe-plugin-tfbridge/assets/23390438/4c1ce602-2045-4943-b07a-4383471b611c)

### Steampipe log file

_(before running `steampipe query`, run `export STEAMPIPE_LOG_LEVEL=debug` and `export TF_LOG_PROVIDER=debug`, to increase the verbosity of the logs)_

_(search for the current day in `~/.steampipe/logs/plugin-*.log` and paste the relevant contents here, since the last boot of Steampipe. Alternatively, run `tail -f ~/.steampipe/logs/plugin-*.log`, then start Steampipe and cause the error, and then copy all the logs since Steampipe booted here. Feel free to remove any secret data such as API keys from the logs before pasting them)_

_(⚠ be careful, since high-verbosity logs WILL expose configuration items, such as API keys. You're encouraged to rotate the API credentials used after submitting the issue. I'll try to warn you if I find such credentials in the logs)_

<details>
  <summary>Steampipe logs</summary>
  
```
2023-08-12 22:05:54.002 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: Plugin execute table: github_users  (169187795474)
2023-08-12 22:05:54.002 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: Plugin execute, setting memory limit to -1Mb
2023-08-12 22:05:54.002 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: Setting max concurrent connections to 25
2023-08-12 22:05:54.002 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: executeForConnection callId: 169187795474, connectionCallId: tfbridge-169187795474, connection: tfbridge table: github_users cols: usernames,id,emails,logins,node_ids,unknown_logins,_ctx
2023-08-12 22:05:54.002 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: Setting free memory interval to 100 rows
2023-08-12 22:05:54.004 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: cacheEnabled, trying cache get (tfbridge-169187795474)
2023-08-12 22:05:54.004 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: getCachedQueryResult returned CACHE MISS - checking for pending transfers (tfbridge-169187795474)
2023-08-12 22:05:54.004 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: findAndSubscribeToPendingRequest returning error cache miss - this will be treated as a cache miss, so add pending item (tfbridge-169187795474)
2023-08-12 22:05:54.004 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: queryCacheGet returned CACHE MISS (tfbridge-169187795474)
2023-08-12 22:05:54.007 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: tfbridge.ListDataSource:
2023-08-12 22:05:54.007 UTC [DEBUG] steampipe-plugin-tfbridge.plugin:   equalsQuals=
2023-08-12 22:05:54.007 UTC [DEBUG] steampipe-plugin-tfbridge.plugin:   | usernames = [
2023-08-12 22:05:54.007 UTC [DEBUG] steampipe-plugin-tfbridge.plugin:   |     "jreyesr"
2023-08-12 22:05:54.007 UTC [DEBUG] steampipe-plugin-tfbridge.plugin:   | ]
2023-08-12 22:05:54.007 UTC [DEBUG] steampipe-plugin-tfbridge.plugin:   
2023-08-12 22:05:54.007 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: tfbridge.ListDataSource: location=/tmp/steampipe-plugin-tfbridge/tfbridge/2023-08-12T22:05:35Z/terraform-provider-github_5.33.0_linux_amd64.zip/terraform-provider-github_v5.33.0
2023-08-12 22:05:54.007 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: 2023-08-12T17:05:54.007-0500 [DEBUG] provider: starting plugin: path=/usr/bin/sh args=["sh", "-c", "/tmp/steampipe-plugin-tfbridge/tfbridge/2023-08-12T22:05:35Z/terraform-provider-github_5.33.0_linux_amd64.zip/terraform-provider-github_v5.33.0"]
2023-08-12 22:05:54.007 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: 2023-08-12T17:05:54.007-0500 [DEBUG] provider: plugin started: path=/usr/bin/sh pid=254307
2023-08-12 22:05:54.007 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: 2023-08-12T17:05:54.007-0500 [DEBUG] provider: waiting for RPC address: path=/usr/bin/sh
2023-08-12 22:05:54.023 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: 2023-08-12T17:05:54.023-0500 [DEBUG] provider: using plugin: version=5
2023-08-12 22:05:54.023 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: 2023-08-12T17:05:54.023-0500 [DEBUG] provider.sh: plugin address: address=/tmp/plugin3433056442 network=unix timestamp=2023-08-12T17:05:54.023-0500
2023-08-12 22:05:54.025 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: [DEBUG] 169187795474: configureProvider:
2023-08-12 22:05:54.025 UTC [DEBUG] steampipe-plugin-tfbridge.plugin:   rawConfig=
2023-08-12 22:05:54.025 UTC [DEBUG] steampipe-plugin-tfbridge.plugin:   |     token = "ITSAGITHUBPAT"
2023-08-12 22:05:54.025 UTC [DEBUG] steampipe-plugin-tfbridge.plugin:   
2023-08-12 22:05:54.033 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: [DEBUG] 169187795474: configureProvider: parsedConfigType="{{{} map[app_auth:{{{} {{{} map[id:{{{} 83}} installation_id:{{{} 83}} pem_file:{{{} 83}}] map[]}}}} base_url:{{{} 83}} insecure:{{{} 66}} organization:{{{} 83}} owner:{{{} 83}} parallel_requests:{{{} 66}} read_delay_ms:{{{} 78}} token:{{{} 83}} write_delay_ms:{{{} 78}}] map[]}}" parsedConfig="{{{{} map[app_auth:{{{} {{{} map[id:{{{} 83}} installation_id:{{{} 83}} pem_file:{{{} 83}}] map[]}}}} base_url:{{{} 83}} insecure:{{{} 66}} organization:{{{} 83}} owner:{{{} 83}} parallel_requests:{{{} 66}} read_delay_ms:{{{} 78}} token:{{{} 83}} write_delay_ms:{{{} 78}}] map[]}} map[app_auth:[] base_url:<nil> insecure:<nil> organization:<nil> owner:<nil> parallel_requests:<nil> read_delay_ms:<nil> token:ITSAGITHUBPAT write_delay_ms:<nil>]}"
2023-08-12 22:05:54.043 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: 2023-08-12T17:05:54.043-0500 [DEBUG] provider.sh: 2023/08/12 17:05:54 [INFO] Selecting owner  from GITHUB_OWNER environment variable
2023-08-12 22:05:54.043 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: 2023-08-12T17:05:54.043-0500 [DEBUG] provider.sh: 2023/08/12 17:05:54 [INFO] Setting write_delay_ms to 1000
2023-08-12 22:05:54.043 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: 2023-08-12T17:05:54.043-0500 [DEBUG] provider.sh: 2023/08/12 17:05:54 [DEBUG] Setting read_delay_ms to 0
2023-08-12 22:05:54.043 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: 2023-08-12T17:05:54.043-0500 [DEBUG] provider.sh: 2023/08/12 17:05:54 [DEBUG] Setting parallel_requests to false
2023-08-12 22:05:54.511 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: 2023-08-12T17:05:54.511-0500 [DEBUG] provider.sh: 2023/08/12 17:05:54 [INFO] Token present; configuring authenticated owner: jreyesr
2023-08-12 22:05:54.531 UTC [DEBUG] steampipe-plugin-tfbridge.plugin: [DEBUG] 169187795474: readDataSource: dsSchema="&{0 0xc014fff9b0}" dsSchemaType="{{{} map[emails:{{{} {{{} 83}}}} id:{{{} 83}} logins:{{{} {{{} 83}}}} node_ids:{{{} {{{} 83}}}} unknown_logins:{{{} {{{} 83}}}} usernames:{{{} {{{} 83}}}}] map[]}}"
2023-08-12 22:05:54.531 UTC [WARN]  steampipe-plugin-tfbridge.plugin: [WARN]  169187795474: readDataSource.makeSimpleQuals.unsupported: qualName=usernames qual="jsonb_value:\"[\n    \\"jreyesr\\"\n]\"" typeInSchema="{{{} {{{} 83}}}}" err="type {{{} {{{} 83}}}} can't be handled by quals"
2023-08-12 22:05:54.531 UTC [WARN]  steampipe-plugin-tfbridge.plugin: [WARN]  169187795474: tfbridge.ListDataSource.readDataSource: name=github_users
2023-08-12 22:05:54.531 UTC [WARN]  steampipe-plugin-tfbridge.plugin: [WARN]  169187795474: doList callHydrateWithRetries (tfbridge-169187795474) returned err type {{{} {{{} 83}}}} can't be handled by quals
2023-08-12 22:05:54.531 UTC [WARN]  steampipe-plugin-tfbridge.plugin: [WARN]  169187795474: QueryData StreamError type {{{} {{{} 83}}}} can't be handled by quals (tfbridge-169187795474)
2023-08-12 22:05:54.531 UTC [WARN]  steampipe-plugin-tfbridge.plugin: [WARN]  169187795474: streamRows for tfbridge-169187795474 - execution has failed (type {{{} {{{} 83}}}} can't be handled by quals) - calling queryCache.AbortSet
2023-08-12 22:05:54.531 UTC [WARN]  steampipe-plugin-tfbridge.plugin: [WARN]  169187795474: executeForConnection tfbridge returned error type {{{} {{{} 83}}}} can't be handled by quals
2023-08-12 22:05:54.531 UTC [WARN]  steampipe-plugin-tfbridge.plugin: [WARN]  169187795474: error channel received type {{{} {{{} 83}}}} can't be handled by quals
2023-08-12 22:05:54.531 UTC [INFO]  steampipe-plugin-tfbridge.plugin: [INFO]  169187795474: Plugin execute complete (169187795474)
```
</details>
