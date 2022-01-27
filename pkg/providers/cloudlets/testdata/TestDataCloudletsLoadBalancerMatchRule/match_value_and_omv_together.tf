provider "akamai" {
  edgerc = "~/.edgerc"
}

data "akamai_cloudlets_application_load_balancer_match_rule" "test" {
  match_rules {
    name      = "rule1"
    start     = 10
    end       = 10000
    match_url = "example.com"
    matches {
      match_type     = "clientip"
      match_operator = "equals"
      match_value    = "example.ex"
      object_match_value {
        type  = "simple"
        value = ["abc"]
      }
    }
    forward_settings {
      origin_id = "33"
    }
    disabled = false
  }
}