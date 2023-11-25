package firefox

// https://github.com/mozilla/policy-templates/blob/master/README.md#homepage

import (
	_ "embed"
	"encoding/json"

	"github.com/m1k1o/neko-rooms/internal/types"
)

func Parse(policiesJson string) (*types.BrowserPolicyContent, error) {
	policies := types.BrowserPolicyContent{}

	policiesTmpl := struct {
		Policies map[string]any `json:"policies"`
	}{}
	if err := json.Unmarshal([]byte(policiesJson), &policiesTmpl); err != nil {
		return nil, err
	}

	// empty file
	if policiesTmpl.Policies == nil {
		return &policies, nil
	}

	//
	// Extensions
	//

	if extensions, ok := policiesTmpl.Policies["ExtensionSettings"]; ok {
		policies.Extensions = []types.BrowserPolicyExtension{}
		for id, val := range extensions.(map[string]any) {
			if id == "*" {
				continue
			}

			data := val.(map[string]any)
			url, _ := data["install_url"].(string)

			policies.Extensions = append(
				policies.Extensions,
				types.BrowserPolicyExtension{
					ID:  id,
					URL: url,
				},
			)
		}
	}

	//
	// Developer Tools
	//

	if val, ok := policiesTmpl.Policies["DisableDeveloperTools"]; ok {
		policies.DeveloperTools = !val.(bool)
	}

	//
	// Persistent Data
	//

	if val, ok := policiesTmpl.Policies["SanitizeOnShutdown"]; ok {
		policies.PersistentData = !val.(bool)
	}

	return &policies, nil
}
