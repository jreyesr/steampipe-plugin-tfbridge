connection "tfbridge" {
  plugin = "tfbridge"

  provider = "cloudflare/cloudflare"
  version = "4.11.0"
  
  provider_config = {
    api_user_service_key = "my_service_key"
  }
}