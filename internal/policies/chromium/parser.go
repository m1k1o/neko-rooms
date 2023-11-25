package chromium

// https://chromeenterprise.google/policies/

import (
	"encoding/json"
	"strings"

	"github.com/m1k1o/neko-rooms/internal/types"
)

func Parse(policiesJson string) (*types.BrowserPolicyContent, error) {
	policies := types.BrowserPolicyContent{}

	policiesTmpl := map[string]any{}
	if err := json.Unmarshal([]byte(policiesJson), &policiesTmpl); err != nil {
		return nil, err
	}

	//
	// Extensions
	//

	if extensions, ok := policiesTmpl["ExtensionInstallForcelist"]; ok {
		policies.Extensions = []types.BrowserPolicyExtension{}
		for _, val := range extensions.([]any) {
			s := strings.Split(val.(string), ";")
			url := ""
			if len(s) > 1 {
				url = s[1]
			}
			policies.Extensions = append(
				policies.Extensions,
				types.BrowserPolicyExtension{
					ID:  s[0],
					URL: url,
				},
			)
		}
	}

	//
	// Developer Tools
	//

	if val, ok := policiesTmpl["DeveloperToolsAvailability"]; ok {
		policies.DeveloperTools = val.(float64) == 1
	}

	//
	// Persistent Data
	//

	if val, ok := policiesTmpl["DefaultCookiesSetting"]; ok {
		policies.PersistentData = val.(float64) == 1
	}

	return &policies, nil
}
