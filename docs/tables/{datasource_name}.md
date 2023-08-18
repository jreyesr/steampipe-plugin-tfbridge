# Table: {datasource_name}

Query data from Terraform data sources. A table is automatically created to represent each
data source found in the configured Terraform provider.

For instance, if `provider` is set to `integrations/github` and `version` is set to `5.33.0`, this plugin will create 57 tables (see [here](https://registry.terraform.io/providers/integrations/github/5.33.0/docs/data-sources/actions_environment_secrets) and below):

- github_actions_environment_secrets
- github_actions_environment_variables
- github_actions_organization_oidc_subject_claim_customization_template
- github_actions_organization_public_key
- github_actions_registration_token
- github_actions_organization_secrets
- ...

Which you can then query directly (see [here](https://registry.terraform.io/providers/integrations/github/5.33.0/docs/data-sources/repository) to compare):

```sql
select
  full_name, description, topics, html_url, repo_id
from
  github_repository
where
  full_name='turbot/steampipe';
```

All columns will have data types that match their Terraform types, if possible, or JSONB otherwise (such as nested attributes). For example, for the `github_repository` table, [`private` will be a BOOL](https://registry.terraform.io/providers/integrations/github/5.33.0/docs/data-sources/repository#private), while [`topics` will be a JSONB](https://registry.terraform.io/providers/integrations/github/5.33.0/docs/data-sources/repository#topics), since originally it's a list of strings.

## Examples

⚠ The following examples use different Terraform providers. This is intentional, since the Terraform Bridge plugin is able to work with _many_ different Terraform providers, each of which will generate a different set of tables.

### Inspect the table structure

Assuming your connection is called `tfbridge` (the default), and that you're using the `integrations/github` Terraform provider at version `5.33.0`, you can list all tables with:

```bash
.inspect tfbridge
+-----------------------------------------------------------------+-------------------------------------------------------------------------+
| table                                                           | description                                                             |
+-----------------------------------------------------------------+-------------------------------------------------------------------------+
| github_actions_environment_secrets                              | github_actions_environment_secrets:                                     |
| github_actions_environment_variables                            | github_actions_environment_variables:                                   |
| github_actions_organization_oidc_subject_claim_customization_te | github_actions_organization_oidc_subject_claim_customization_template:  |
| github_actions_organization_public_key                          | github_actions_organization_public_key:                                 |
| github_actions_organization_registration_token                  | github_actions_organization_registration_token:                         |
...
+-----------------------------------------------------------------+-------------------------------------------------------------------------+
```

To get defails for a specific table, inspect it by name:

```bash
.inspect tfbridge.github_repositories
+------------------+------------------+-------------------------------------------------------+
| column           | type             | description                                           |
+------------------+------------------+-------------------------------------------------------+
| _ctx             | jsonb            | Steampipe context in JSON form, e.g. connection_name. |
| full_names       | jsonb            |                                                       |
| id               | text             |                                                       |
| include_repo_id  | boolean          |                                                       |
| names            | jsonb            |                                                       |
| query            | text             |                                                       |
| repo_ids         | jsonb            |                                                       |
| results_per_page | double precision |                                                       |
| sort             | text             |                                                       |
+------------------+------------------+-------------------------------------------------------+
```

### Query for a list of Github repos

Assume that you're using the `integrations/github` Terraform provider at version `5.33.0`.

```sql
select 
  * 
from 
  github_repositories 
where 
  query='org:turbot' 
  and sort='stars'
```

```
+------------+------------+-----------------+------------------+-------+------------------------------------------------------------------------------------------------------->
| query      | id         | include_repo_id | results_per_page | sort  | full_names                                                                                            >
+------------+------------+-----------------+------------------+-------+------------------------------------------------------------------------------------------------------->
| org:turbot | org:turbot | false           | 100              | stars | ["turbot/steampipe","turbot/steampipe-mod-aws-compliance","turbot/steampipe-plugin-aws","turbot/steamp>
|            |            |                 |                  |       | plugin-whois","turbot/steampipe-plugin-prometheus","turbot/steampipe-plugin-csv","turbot/steampipe-plu>
|            |            |                 |                  |       | urbot/terraform-provider-turbot","turbot/terraform-provider-steampipecloud","turbot/steampipe-plugin-a>
|            |            |                 |                  |       | eampipe-mod-digitalocean-thrifty","turbot/steampipe-plugin-crtsh","turbot/steampipe-plugin-abuseipdb",>
|            |            |                 |                  |       | ","turbot/steampipe-plugin-hcloud","turbot/steampipe-plugin-fly","turbot/steampipe-plugin-equinix","tu>
|            |            |                 |                  |       | pipe-plugin-workos","turbot/steampipe-mod-oci-insights","turbot/steampipe-mod-digitalocean-insights",">
|            |            |                 |                  |       | hrifty","turbot/steampipe-mod-jira-sherlock","turbot/steampipe-plugin-ansible","turbot/steampipe-plugi>
|            |            |                 |                  |       | teampipe-plugin-chaosratelimit","turbot/steampipe-mod-googleworkspace-compliance","turbot/steampipe-mo>
+------------+------------+-----------------+------------------+-------+------------------------------------------------------------------------------------------------------->
```

### Getting details of an Algolia index

Assume that you're using the `k-yomo/algolia` Terraform provider at version `0.5.7`.


```sql
select 
  * 
from 
  algolia_index 
where
  name='indexname'
```

```
+-----------+-----------+-------------------------------------------------------------------------------------------------------------------------------------+
| name      | id        | advanced_config                                                                                                                     |
+-----------+-----------+-------------------------------------------------------------------------------------------------------------------------------------+
| yourindex | yourindex | [{"attribute_criteria_computed_by_min_proximity": true, "attribute_for_distinct": "url", "distinct": 1, "max_facet_hits": 10, ...}] |
+-----------+-----------+-------------------------------------------------------------------------------------------------------------------------------------+
```


### Getting information about your public IP

Assume that you're using the `dewhurstwill/whatsmyip` Terraform provider at version `1.0.3`.

```sql
select 
  * 
from 
  whatsmyip 
```

```
+----+---------------+---------+---------+
| cc | country       | id      | ip      |
+----+---------------+---------+---------+
| US | United States | 1.1.1.1 | 1.1.1.1 |
+----+---------------+---------+---------+
```

## General translation rules

The following rules are followed to translate from Terraform data sources to Steampipe/SQL tables:

* Every Terraform data source becomes a SQL table
    * The name and description of the table are taken directly from the Terraform schema
* Every attribute in the Terraform data source becomes a column in the table
    * Data types are translated in a best-effort basis: strings, numbers and booleans will become their corresponding Postgres types, and more complex Terraform types will become JSONB columns
* Required attributes in the Terraform data source (such as resource IDs if the data source returns data about a single object) _must_ be provided via `WHERE` clauses
    * This requires that the Terraform provider has actually marked the attributes as required
* Other/more complex `WHERE` conditions may also be expressed, but those won't be passed to the Terraform provider. 
    * For example, `LIKE` conditions on text fields, or numerical comparisons (such as greater-than or is-even), or comparisons between columns
    * If you use such conditions, be aware that the Terraform provider will receive a request to list _all_ data, and thus may incur on large API usage, even if most of that data will be discarded by a later condition
* It's up to the Terraform provider to implement "singular" and "plural" data sources (i.e. data sources that return information about a single item or about a list of items). If the Terraform data source is singular (i.e. you must provide a unique ID for the item that you wish to look up), this plugin won't be able to list all items, since it would need to somehow invent the IDs of the items
    * For a provider that does implement singular and plural data sources, see [the Grafana provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/data-sources/dashboard), in particular see the `dashboard`/`dashboards`, `folder`/`folders` and `user`/`users` pairs of data sources
    * However, if you somehow have a list of item IDs to query, you can use a `WHERE unique_id IN('id1', 'id2', ...)` condition, and it _will_ work as expected on a singular data source (since such queries are internally expanded into many parallel queries with `WHERE unique_id='id1'` and so on, which can be satisfied by a singular data source)