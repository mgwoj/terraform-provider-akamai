package appsec

import (
	"encoding/json"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/appsec"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/common/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAkamaiSelectableHostnames_data_basic(t *testing.T) {
	t.Run("match by SelectableHostnames ID", func(t *testing.T) {
		client := &appsec.Mock{}

		config := appsec.GetConfigurationResponse{}
		err := json.Unmarshal(testutils.LoadFixtureBytes(t, "testdata/TestResConfiguration/LatestConfiguration.json"), &config)
		require.NoError(t, err)

		client.On("GetConfiguration",
			mock.Anything,
			appsec.GetConfigurationRequest{ConfigID: 43253},
		).Return(&config, nil)

		getSelectableHostnamesResponse := appsec.GetSelectableHostnamesResponse{}
		err = json.Unmarshal(testutils.LoadFixtureBytes(t, "testdata/TestDSSelectableHostnames/SelectableHostnames.json"), &getSelectableHostnamesResponse)
		require.NoError(t, err)

		client.On("GetSelectableHostnames",
			mock.Anything,
			appsec.GetSelectableHostnamesRequest{ConfigID: 43253, Version: 7},
		).Return(&getSelectableHostnamesResponse, nil)

		useClient(client, func() {
			resource.Test(t, resource.TestCase{
				IsUnitTest:        true,
				ProviderFactories: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: testutils.LoadFixtureString(t, "testdata/TestDSSelectableHostnames/match_by_id.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.akamai_appsec_selectable_hostnames.test", "id", "43253"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

}
