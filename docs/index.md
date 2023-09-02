---
organization: jreyesr
category: ["software development"]
icon_url: "/images/plugins/jreyesr/tfbridge.svg"
brand_color: "#844FBA"
display_name: Terraform Bridge
name: tfbridge
description: Steampipe plugin for proxying requests to Terraform data sources.
og_description: Query any Terraform provider with SQL! Open source CLI. No DB required.
og_image: "/images/plugins/jreyesr/tfbridge-social-graphic.png"
---

# Terraform data sources + Steampipe

[Terraform](https://www.terraform.io/) lets you provision and manage resources in any cloud or data center. It has many providers, which let Terraform interface with external systems.

[Steampipe](https://steampipe.io) is an open source CLI to instantly query cloud APIs using SQL.

The Terraform Bridge plugin for Steampipe lets you reuse the data sources that some Terraform providers come with, so you can query them using Steampipe's SQL interface. If there is a Terraform provider that has a data source that already obtains the data that you need, you can consume that data from Steampipe (normally you'd have to use a Terraform file and `data {}` blocks)

For example:

(using the [Github Terraform provider's data sources](https://registry.terraform.io/providers/integrations/github/5.33.0))

```sql
select 
  * 
from 
  github_repositories
where 
  query='org:turbot';
```

| query      | id         | include_repo_id | results_per_page | sort    | full_names | names |
|---|---|---|---|---|---|---|
| org:turbot | org:turbot | false           | 100              | updated | ["turbot/steampipe", "turbot/steampipe-plugin-aws", "turbot/steampipe-mod-github-sherlock", ...] | ["steampipe", "steampipe-plugin-aws", "steampipe-mod-github-sherlock", ...] |

(using an [Algolia Terraform provider](https://registry.terraform.io/providers/k-yomo/algolia/latest))

```sql
select 
  * 
from 
  algolia_index 
where
  name='indexname'
```

| name      | id        | advanced_config |
|---|---|---|
| yourindex | yourindex | [{"attribute_criteria_computed_by_min_proximity": true, "attribute_for_distinct": "url", "distinct": 1, "max_facet_hits": 10, "min_proximity": 1, ...}]|

## Documentation

- **[Table definitions & examples →](/plugins/jreyesr/tfbridge/tables)**

## Get started

### Install

Download and install the latest Terraform Bridge plugin:

```bash
steampipe plugin install jreyesr/tfbridge
```

### Credentials

See [your Terraform provider](https://registry.terraform.io/) of choice for credentials. Anything that would normally go into the `provider "name" {...}` block in the Terraform file should be placed directly in the Steampipe config file (`~/.steampipe/config/tfbridge.spc`), in the `provider_config` value.

### Configuration

Installing the latest Terraform Bridge plugin will create a config file (`~/.steampipe/config/tfbridge.spc`) with a single connection named `tfbridge`:

```hcl
connection "tfbridge" {
  plugin = "jreyesr/tfbridge"

  # Write the name of a Terraform provider, in the same way that you'd write it
  # in the required_providers block in the Terraform main file
  # Examples: "hashicorp/aws", "TimDurward/slack", "hashicorp/random"
  # If using a private Terraform registry, also include the hostname: "registry.acme.com/acme/supercloud"
  # provider = "integrations/github"

  # Write a single version for a Terraform provider, in the same way that you'd write it
  # in the required_providers block in the Terraform main file
  # Note that, unlike in the Terraform file, you can't use a version constraint, such as "~> 1.0" or ">= 1.2.0, < 2.0.0",
  # only explicit versions are allowed
  # version = "5.33.0"

  # If the Terraform provider would require some configuration in its provider {...} block,
  # copy it here directly (just the _contents_ of the provider {} block, not the entire block!)
  # provider_config = <<EOT
  #   token = "github_pat_9fu38f0amil9FVOKmI0_0F8PmI0m0FNm"
  # EOT
}
```

Uncomment and edit the `provider`, `version` and `provider_config` parameters.

`provider` and `version` define which Terraform provider you want to use. You can find those in [the Terraform registry](https://registry.terraform.io/). If you have a working Terraform configuration, you may also run `terraform version` to retrieve the versions currently in use.

`provider_config` is a string that contains any configuration that must be forwarded to the Terraform provider. It should contain the entire contents of the `provider "yourprovname" {...}` block in the Terraform configuration (_only_ what is inside the curly braces, but not the `provider "yourprovname"` part). Those configuration values can be found in the Terraform Registry docs for your provider of choice.

## Get involved

* Open source: https://github.com/jreyesr/steampipe-plugin-tfbridge
* Community: [Join #steampipe on Slack →](https://turbot.com/community/join)