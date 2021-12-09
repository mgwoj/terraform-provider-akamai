package cloudlets

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataCloudletsVisitorPrioritizationMatchRule(t *testing.T) {
	workdir := "testdata/TestDataCloudletsVisitorPrioritizationMatchRule"

	tests := map[string]struct {
		configPath       string
		expectedJSONPath string
		matchRulesSize   int
	}{
		"valid all vars map": {
			configPath:       fmt.Sprintf("%s/vars_map.tf", workdir),
			expectedJSONPath: fmt.Sprintf("%s/rules/rules_out.json", workdir),
			matchRulesSize:   3,
		},
		"valid minimal vars map": {
			configPath:       fmt.Sprintf("%s/minimal_vars_map.tf", workdir),
			expectedJSONPath: fmt.Sprintf("%s/rules/minimal_rules_out.json", workdir),
			matchRulesSize:   1,
		},
		"match criteria VP - ObjectMatchValue of Simple type": {
			configPath:       fmt.Sprintf("%s/omv_simple.tf", workdir),
			expectedJSONPath: fmt.Sprintf("%s/rules/omv_simple_rules.json", workdir),
			matchRulesSize:   1,
		},
		"match criteria VP - ObjectMatchValue of Object type": {
			configPath:       fmt.Sprintf("%s/omv_object.tf", workdir),
			expectedJSONPath: fmt.Sprintf("%s/rules/omv_object_rules.json", workdir),
			matchRulesSize:   1,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString(test.configPath),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(
								"data.akamai_cloudlets_visitor_prioritization_match_rule.test", "json",
								loadFixtureString(test.expectedJSONPath)),
							resource.TestCheckResourceAttr(
								"data.akamai_cloudlets_visitor_prioritization_match_rule.test", "match_rules.0.type", "vpMatchRule"),
							resource.TestCheckResourceAttr(
								"data.akamai_cloudlets_visitor_prioritization_match_rule.test", "match_rules.#", strconv.Itoa(test.matchRulesSize)),
						),
					},
				},
			})
		})
	}
}

func TestIncorrectDataCloudletsVisitorPrioritizationMatchRule(t *testing.T) {
	workdir := "testdata/TestDataCloudletsVisitorPrioritizationMatchRule"

	tests := map[string]struct {
		configPath     string
		withError      string
		matchRulesSize int
	}{
		"missing passThroughPercent": {
			configPath:     fmt.Sprintf("%s/missing_argument.tf", workdir),
			withError:      "Missing required argument",
			matchRulesSize: 1,
		},
		"match criteria VP - invalid type value for ObjectMatchValue": {
			configPath:     fmt.Sprintf("%s/omv_invalid_type.tf", workdir),
			withError:      `expected type to be one of \[simple object\], got range`,
			matchRulesSize: 1,
		},
		"match criteria VP - invalid matches_operator value": {
			configPath:     fmt.Sprintf("%s/matches_invalid_operator.tf", workdir),
			withError:      `expected match_operator to be one of \[contains exists equals \], got invalid`,
			matchRulesSize: 1,
		},
		"match criteria VP - invalid pass_through_percent value": {
			configPath:     fmt.Sprintf("%s/invalid_pass_through_percent.tf", workdir),
			withError:      `expected pass_through_percent to be in the range \(-1.000000 - 100.000000\), got -2.000000`,
			matchRulesSize: 1,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString(test.configPath),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(
								"data.akamai_cloudlets_visitor_prioritization_match_rule.test", "match_rules.0.type", "vpMatchRule"),
							resource.TestCheckResourceAttr(
								"data.akamai_cloudlets_visitor_prioritization_match_rule.test", "match_rules.#", strconv.Itoa(test.matchRulesSize)),
						),
						ExpectError: regexp.MustCompile(test.withError),
					},
				},
			})
		})
	}
}