provider "akamai" {
  edgerc        = "../../test/edgerc"
  cache_enabled = false
}

resource "akamai_appsec_reputation_protection" "test" {
  config_id          = 43253
  security_policy_id = "AAAA_81230"
  enabled            = false
}

