provider "akamai" {
  edgerc = "~/.edgerc"
}

resource "akamai_imaging_policy_image" "policy" {
  policy_id              = "test_policy"
  contract_id            = "test_contract"
  policyset_id           = "test_policy_set"
  activate_on_production = false
  json = jsonencode({
    "breakpoints" : {
      "widths" : [
        320,
        640,
        1024,
        2048,
        5000
      ]
    },
    "output" : {
      "perceptualQuality" : "mediumHigh"
    },
    "rolloutDuration" : 3600,
    "transformations" : [
      {
        "colors" : 2,
        "transformation" : "MaxColors"
      }
    ]
  })
}