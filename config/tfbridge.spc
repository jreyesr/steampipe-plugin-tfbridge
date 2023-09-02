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