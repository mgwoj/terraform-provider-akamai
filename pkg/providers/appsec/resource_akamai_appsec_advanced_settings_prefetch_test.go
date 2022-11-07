package appsec

import (
	"encoding/json"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v3/pkg/appsec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAkamaiAdvancedSettingsPrefetch_res_basic(t *testing.T) {
	t.Run("match by AdvancedSettingsPrefetch ID", func(t *testing.T) {
		client := &mockappsec{}

		updateResponse := appsec.UpdateAdvancedSettingsPrefetchResponse{}
		err := json.Unmarshal(loadFixtureBytes("testdata/TestResAdvancedSettingsPrefetch/AdvancedSettingsPrefetch.json"), &updateResponse)
		require.NoError(t, err)

		deleteResopnse := appsec.UpdateAdvancedSettingsPrefetchResponse{}
		err = json.Unmarshal(loadFixtureBytes("testdata/TestResAdvancedSettingsPrefetch/AdvancedSettingsPrefetch.json"), &deleteResopnse)
		require.NoError(t, err)

		getResponse := appsec.GetAdvancedSettingsPrefetchResponse{}
		err = json.Unmarshal(loadFixtureBytes("testdata/TestResAdvancedSettingsPrefetch/AdvancedSettingsPrefetch.json"), &getResponse)
		require.NoError(t, err)

		config := appsec.GetConfigurationResponse{}
		err = json.Unmarshal(loadFixtureBytes("testdata/TestResConfiguration/LatestConfiguration.json"), &config)
		require.NoError(t, err)

		client.On("GetConfiguration",
			mock.Anything,
			appsec.GetConfigurationRequest{ConfigID: 43253},
		).Return(&config, nil)

		client.On("GetAdvancedSettingsPrefetch",
			mock.Anything,
			appsec.GetAdvancedSettingsPrefetchRequest{ConfigID: 43253, Version: 7},
		).Return(&getResponse, nil)

		client.On("UpdateAdvancedSettingsPrefetch",
			mock.Anything,
			appsec.UpdateAdvancedSettingsPrefetchRequest{ConfigID: 43253, Version: 7, AllExtensions: false, EnableAppLayer: true, EnableRateControls: false, Extensions: []string{"cgi", "asp", "php", "jsp", "EMPTY_STRING", "py", "aspx"}},
		).Return(&updateResponse, nil)

		client.On("UpdateAdvancedSettingsPrefetch",
			mock.Anything,
			appsec.UpdateAdvancedSettingsPrefetchRequest{ConfigID: 43253, Version: 7, AllExtensions: false, EnableAppLayer: false, EnableRateControls: false, Extensions: []string(nil)},
		).Return(&updateResponse, nil)

		useClient(client, func() {
			resource.Test(t, resource.TestCase{
				IsUnitTest: true,
				Providers:  testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString("testdata/TestResAdvancedSettingsPrefetch/match_by_id.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("akamai_appsec_advanced_settings_prefetch.test", "id", "43253"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

}
