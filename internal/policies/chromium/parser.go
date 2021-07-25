package chromium

// https://chromeenterprise.google/policies/

import (
	"encoding/json"
	"m1k1o/neko_rooms/internal/types"
	"strings"
)

func Parse(policiesJson string) (*types.Policies, error) {
	policies := types.Policies{}

	policiesTmpl := map[string]interface{}{}
	if err := json.Unmarshal([]byte(policiesJson), &policiesTmpl); err != nil {
		return nil, err
	}

	//
	// Extensions
	//

	if extensions, ok := policiesTmpl["ExtensionInstallForcelist"]; ok {
		policies.Extensions = []types.Extension{}
		for _, val := range extensions.([]interface{}) {
			s := strings.Split(val.(string), ";")
			policies.Extensions = append(
				policies.Extensions,
				types.Extension{
					ID:  s[0],
					URL: s[1],
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
