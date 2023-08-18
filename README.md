![image](https://hub.steampipe.io/images/plugins/jreyesr/tfbridge-social-graphic.png)

# Terraform Bridge Plugin for Steampipe

Use SQL to query data from the datasources of any Terraform provider.

This repo contains a [Steampipe](https://steampipe.io/) plugin that lets the user call any (?) [data source that is exposed by a Terraform provider](https://developer.hashicorp.com/terraform/language/data-sources). This will expand the reach of Steampipe's plugins to also cover remote APIs that have a Terraform provider but no Steampipe plugin. This will also let users unify efforts: the Terraform provider can be used to manage resources, and querying on the current state of those resources can be done via Steampipe, using the same source code and provider.


- **[Get started →](https://hub.steampipe.io/plugins/jreyesr/tfbridge)**
- Documentation: [Table definitions & examples](https://hub.steampipe.io/plugins/jreyesr/tfbridge/tables)
- Community: [Join #steampipe on Slack →](https://turbot.com/community/join)
- Get involved: [Issues](https://github.com/jreyesr/steampipe-plugin-tfbridge/issues)

## Quick start

Install the plugin with [Steampipe](https://steampipe.io):

```shell
steampipe plugin install jreyesr/tfbridge
```

Configure your [config file](https://hub.steampipe.io/plugins/jreyesr/tfbridge#configuration) to point to a Terraform provider+version. If the Terraform provider requires configuration values, provide them too.

Run steampipe:

```shell
steampipe query
```

Run a query for whatever data source the Terraform provider exposes:

```sql
select
  attr1,
  attr2
from
  datasource_name;
```

## Developing

Prerequisites:

- [Steampipe](https://steampipe.io/downloads)
- [Golang](https://golang.org/doc/install)

Clone:

```sh
git clone https://github.com/jreyesr/steampipe-plugin-tfbridge.git
cd steampipe-plugin-tfbridge
```

Build, which automatically installs the new version to your `~/.steampipe/plugins` directory:

```
make
```

Configure the plugin:

```
cp config/* ~/.steampipe/config
vi ~/.steampipe/config/tfbridge.spc
```

Try it!

```
steampipe query
> .inspect tfbridge
```

Further reading:

- [Writing plugins](https://steampipe.io/docs/develop/writing-plugins)
- [Writing your first table](https://steampipe.io/docs/develop/writing-your-first-table)

## Contributing

Please see the [contribution guidelines](https://github.com/turbot/steampipe/blob/main/CONTRIBUTING.md) and our [code of conduct](https://github.com/turbot/steampipe/blob/main/CODE_OF_CONDUCT.md). All contributions are subject to the [Mozilla Public License 2.0 open source license](https://github.com/jreyesr/steampipe-plugin-tfbridge/blob/main/LICENSE).


## Old proof of concept

Previously (see [the `poc` tag](https://github.com/jreyesr/steampipe-plugin-tfbridge/tree/poc)), the repo only contained a proof of concept Go program that was used to test ways of driving a Terraform plugin. This wasn't too documented before, with [Terraform's docs](https://developer.hashicorp.com/terraform/plugin/best-practices/interacting-with-providers#using-the-rpc-protocol) merely stating that 

> For projects that actually want to drive the provider, the supported option is to use the gRPC protocol and the RPC calls the protocol supplies. This protocol is the same protocol that drives Terraform's CLI interface, and it is versioned using a protocol version.

However, no widely-known projects that I could find did so (i.e., the only major consumer of Terraform providers is the Terraform project itself).

[This series of posts](https://jreyesr.github.io/series/tfbridge/) contains much more information, discussions, pictures, screenshots of tests, comparisons with other tools, and more. Of particular interest may be the first (oldest) 3 posts, since they deal with driving Terraform providers outside of Terraform. Further posts cover the development of a Steampipe plugin, so they may be of less interest if you wish to drive Terraform providers yourself.

The `main.go.old` file implements a subset of Terraform's functionality as it pertains to providers: it spawns a provider, connects to it via gRPC over a Unix domain socket, and then it issues RPCs to the provider.

* `GetProviderSchema` retrieves the provider's schema (which configuration values it requires, which data sources it exposes, and the fields and data types of each data source)
* `ConfigureProvider` provides a set of configuration values to the provider
* `ReadDataSource` retrieves some data from one of the provider's data sources

More information about the RPCs that every provider implements (since Terraform the CLI uses them) can be found [here](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#rpcs-and-terraform-commands).

The `main.go.old` file contains a bunch of code that performs the RPCs on two Terraform plugins, DNS and Terraform Enterprise. Commented code was used before to test oter functionality or other use cases. You may use the file (and all the associated directories) as a starting point to develop your own consumers of Terraform providers.

## Licensing

This repo vendors some internal code from [the Terraform project](https://www.terraform.io/), since that can't be reused via the standard Go import system. Those files are [MPL 2.0](https://choosealicense.com/licenses/mpl-2.0/) and &copy; HashiCorp, Inc. The following directories contain those files:

* `configschema`
* `logging`
* `plugin`
* `providers`
* `tfdiags`
* `tfplugin5`

While [ChooseALicense](https://choosealicense.com/licenses/mpl-2.0/) seems to state that "a larger work using the licensed work may be distributed under different terms", I'm licensing the entire plugin under MPL 2.0 to be sure.