package appsec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/appsec"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/akamai"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/tools"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// appsec v1
//
// https://developer.akamai.com/api/cloud_security/application_security/v1.html
func resourceRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRuleCreate,
		ReadContext:   resourceRuleRead,
		UpdateContext: resourceRuleUpdate,
		DeleteContext: resourceRuleDelete,
		CustomizeDiff: customdiff.All(
			VerifyIDUnchanged,
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"config_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"security_policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rule_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"rule_action": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: ValidateActions,
			},
			"condition_exception": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsJSON),
				DiffSuppressFunc: suppressEquivalentJSONDiffsConditionException,
			},
		},
	}
}

func resourceRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceRuleCreate")
	logger.Debugf("in resourceRuleCreate")

	configID, err := tools.GetIntValue("config_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	version := getModifiableConfigVersion(ctx, configID, "rule", m)
	policyID, err := tools.GetStringValue("security_policy_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	ruleID, err := tools.GetIntValue("rule_id", d)
	if err != nil {
		return diag.FromErr(err)
	}

	conditionexception, err := tools.GetStringValue("condition_exception", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}

	jsonPayloadRaw := []byte(conditionexception)
	rawJSON := (json.RawMessage)(jsonPayloadRaw)

	getWAFMode := appsec.GetWAFModeRequest{}

	getWAFMode.ConfigID = configID
	getWAFMode.Version = version
	getWAFMode.PolicyID = policyID

	wafmode, err := client.GetWAFMode(ctx, getWAFMode)
	if err != nil {
		logger.Errorf("calling 'getWAFMode': %s", err.Error())
		return diag.FromErr(err)
	}

	if wafmode.Mode == AseAuto { // action is read only, only condition exception is writable
		ruleConditionException := appsec.RuleConditionException{}
		if conditionexception != "" {
			err = json.Unmarshal([]byte(rawJSON), &ruleConditionException)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		createRule := appsec.UpdateConditionExceptionRequest{
			ConfigID:               configID,
			Version:                version,
			PolicyID:               policyID,
			RuleID:                 ruleID,
			Conditions:             ruleConditionException.Conditions,
			Exception:              ruleConditionException.Exception,
			AdvancedExceptionsList: ruleConditionException.AdvancedExceptionsList,
		}

		resp, err := client.UpdateRuleConditionException(ctx, createRule)
		if err != nil {
			logger.Errorf("calling 'UpdateRule': %s", err.Error())
			return diag.FromErr(err)
		}
		logger.Debugf("calling 'UpdateRule Response': %s", resp)
		d.SetId(fmt.Sprintf("%d:%s:%d", createRule.ConfigID, createRule.PolicyID, createRule.RuleID))
	} else {
		action, err := tools.GetStringValue("rule_action", d)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := validateActionAndConditionException(action, conditionexception); err != nil {
			return diag.FromErr(err)
		}

		createRule := appsec.UpdateRuleRequest{
			ConfigID:       configID,
			Version:        version,
			PolicyID:       policyID,
			RuleID:         ruleID,
			Action:         action,
			JsonPayloadRaw: rawJSON,
		}

		resp, err := client.UpdateRule(ctx, createRule)
		if err != nil {
			logger.Errorf("calling 'UpdateRule': %s", err.Error())
			return diag.FromErr(err)
		}
		logger.Debugf("calling 'UpdateRule Response': %s", resp)
		d.SetId(fmt.Sprintf("%d:%s:%d", createRule.ConfigID, createRule.PolicyID, createRule.RuleID))

	}

	return resourceRuleRead(ctx, d, m)
}

func resourceRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceRuleRead")
	logger.Debugf("in resourceRuleRead")

	iDParts, err := splitID(d.Id(), 3, "configID:securityPolicyID:ruleID")
	if err != nil {
		return diag.FromErr(err)
	}
	configID, err := strconv.Atoi(iDParts[0])
	if err != nil {
		return diag.FromErr(err)
	}
	version := getLatestConfigVersion(ctx, configID, m)
	policyID := iDParts[1]
	ruleID, err := strconv.Atoi(iDParts[2])
	if err != nil {
		return diag.FromErr(err)
	}

	getRule := appsec.GetRuleRequest{
		ConfigID: configID,
		Version:  version,
		PolicyID: policyID,
		RuleID:   ruleID,
	}

	rule, err := client.GetRule(ctx, getRule)
	if err != nil {
		logger.Errorf("calling 'GetRule': %s", err.Error())
		return diag.FromErr(err)
	}

	if err := d.Set("config_id", getRule.ConfigID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("security_policy_id", getRule.PolicyID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("rule_id", getRule.RuleID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}

	if err := d.Set("rule_action", rule.Action); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}

	if !rule.IsEmptyConditionException() {
		jsonBody, err := json.Marshal(rule.ConditionException)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("condition_exception", string(jsonBody)); err != nil {
			return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
		}
	}

	return nil
}

func resourceRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceRuleUpdate")
	logger.Debugf("in resourceRuleUpdate")

	iDParts, err := splitID(d.Id(), 3, "configID:securityPolicyID:ruleID")
	if err != nil {
		return diag.FromErr(err)
	}
	configID, err := strconv.Atoi(iDParts[0])
	if err != nil {
		return diag.FromErr(err)
	}
	policyID := iDParts[1]
	version := getModifiableConfigVersion(ctx, configID, "rule", m)
	ruleID, err := strconv.Atoi(iDParts[2])
	if err != nil {
		return diag.FromErr(err)
	}
	conditionexception, err := tools.GetStringValue("condition_exception", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	jsonPayloadRaw := []byte(conditionexception)
	rawJSON := (json.RawMessage)(jsonPayloadRaw)

	getWAFMode := appsec.GetWAFModeRequest{}

	getWAFMode.ConfigID = configID
	getWAFMode.Version = version
	getWAFMode.PolicyID = policyID

	wafmode, err := client.GetWAFMode(ctx, getWAFMode)
	if err != nil {
		logger.Errorf("calling 'getWAFMode': %s", err.Error())
		return diag.FromErr(err)
	}

	if wafmode.Mode == AseAuto { // action is read only, only exception is writable
		ruleConditionException := appsec.RuleConditionException{}
		if conditionexception != "" {
			err = json.Unmarshal([]byte(rawJSON), &ruleConditionException)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		updateRule := appsec.UpdateConditionExceptionRequest{
			ConfigID:               configID,
			Version:                version,
			PolicyID:               policyID,
			RuleID:                 ruleID,
			Conditions:             ruleConditionException.Conditions,
			Exception:              ruleConditionException.Exception,
			AdvancedExceptionsList: ruleConditionException.AdvancedExceptionsList,
		}

		resp, err := client.UpdateRuleConditionException(ctx, updateRule)
		if err != nil {
			logger.Errorf("calling 'UpdateRule': %s", err.Error())
			return diag.FromErr(err)
		}
		logger.Debugf("calling 'UpdateRule Response': %s", resp)
	} else {

		action, err := tools.GetStringValue("rule_action", d)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := validateActionAndConditionException(action, conditionexception); err != nil {
			return diag.FromErr(err)
		}

		updateRule := appsec.UpdateRuleRequest{
			ConfigID:       configID,
			Version:        version,
			PolicyID:       policyID,
			RuleID:         ruleID,
			Action:         action,
			JsonPayloadRaw: rawJSON,
		}

		_, err = client.UpdateRule(ctx, updateRule)
		if err != nil {
			logger.Errorf("calling 'UpdateRule': %s", err.Error())
			return diag.FromErr(err)
		}
	}

	return resourceRuleRead(ctx, d, m)
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceRuleDelete")
	logger.Debugf("in resourceRuleDelete")

	iDParts, err := splitID(d.Id(), 3, "configID:securityPolicyID:ruleID")
	if err != nil {
		return diag.FromErr(err)
	}
	configID, err := strconv.Atoi(iDParts[0])
	if err != nil {
		return diag.FromErr(err)
	}
	version := getModifiableConfigVersion(ctx, configID, "rule", m)
	policyID := iDParts[1]
	ruleID, err := strconv.Atoi(iDParts[2])
	if err != nil {
		return diag.FromErr(err)
	}

	getWAFMode := appsec.GetWAFModeRequest{}

	getWAFMode.ConfigID = configID
	getWAFMode.Version = version
	getWAFMode.PolicyID = policyID

	wafmode, err := client.GetWAFMode(ctx, getWAFMode)
	if err != nil {
		logger.Errorf("calling 'getWAFMode': %s", err.Error())
		return diag.FromErr(err)
	}

	if wafmode.Mode == AseAuto {
		updateRule := appsec.UpdateConditionExceptionRequest{
			ConfigID: configID,
			Version:  version,
			PolicyID: policyID,
			RuleID:   ruleID,
		}

		_, err = client.UpdateRuleConditionException(ctx, updateRule)
		if err != nil {
			logger.Errorf("calling 'UpdateRule': %s", err.Error())
			return diag.FromErr(err)
		}
	} else {
		updateRule := appsec.UpdateRuleRequest{
			ConfigID: configID,
			Version:  version,
			PolicyID: policyID,
			RuleID:   ruleID,
			Action:   "none",
		}
		_, err = client.UpdateRule(ctx, updateRule)
		if err != nil {
			logger.Errorf("calling 'UpdateRule': %s", err.Error())
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}
