# `steampipe-plugin-tfbridge`

This repo _will_ contain a [Steampipe](https://steampipe.io/) plugin that lets the user call any (?) [data source that is exposed by a Terraform provider](https://developer.hashicorp.com/terraform/language/data-sources). This will expand the reach of Steampipe's plugins to also cover remote APIs that have a Terraform provider but no Steampipe plugin. This will also let users unify efforts: the Terraform provider can be used to manage resources, and querying on the current state of those resources can be done via Steampipe, using the same source code and provider.

Currently, the repo contains a proof of concept Go program that has been used to test ways of driving a Terraform plugin. This wasn't too documented before, with [Terraform's docs](https://developer.hashicorp.com/terraform/plugin/best-practices/interacting-with-providers#using-the-rpc-protocol) merely stating that 

> For projects that actually want to drive the provider, the supported option is to use the gRPC protocol and the RPC calls the protocol supplies. This protocol is the same protocol that drives Terraform's CLI interface, and it is versioned using a protocol version.

However, no widely-known projects that I could find did so (i.e., the only major consumer of Terraform providers is the Terraform project itself).

[This series of posts](https://jreyesr.github.io/series/tfbridge/) contains much more information, discussions, pictures, screenshots of tests, comparisons with other tools, and more. Of particular interest may be the first (oldest) 3 posts, since they deal with driving Terraform providers outside of Terraform. Further posts cover the development of a Steampipe plugin, so they may be of less interest if you wish to drive Terraform providers yourself.

## Files

The `main.go` file implements a subset of Terraform's functionality as it pertains to providers: it spawns a provider, connects to it via gRPC over a Unix domain socket, and then it issues RPCs to the provider.

* `GetProviderSchema` retrieves the provider's schema (which configuration values it requires, which data sources it exposes, and the fields and data types of each data source)
* `ConfigureProvider` provides a set of configuration values to the provider
* `ReadDataSource` retrieves some data from one of the provider's data sources

More information about the RPCs that every provider implements (since Terraform the CLI uses them) can be found [here](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#rpcs-and-terraform-commands).

The `main.go` file contains a bunch of code that performs the RPCs on two Terraform plugins, DNS and Terraform Enterprise. Commented code was used before to test oter functionality or other use cases. You may use the file (and all the associated directories) as a starting point to develop your own consumers of Terraform providers.

## Licensing

This repo vendors some internal code from [the Terraform project](https://www.terraform.io/), since that can't be reused via the standard Go import system. Those files are [MPL 2.0](https://choosealicense.com/licenses/mpl-2.0/) and &copy; HashiCorp, Inc. The following directories contain those files:

* `configschema`
* `logging`
* `plugin`
* `providers`
* `tfdiags`
* `tfplugin5`

While [ChooseALicense](https://choosealicense.com/licenses/mpl-2.0/) seems to state that "a larger work using the licensed work may be distributed under different terms", I'm licensing the entire plugin under MPL 2.0 to be sure.