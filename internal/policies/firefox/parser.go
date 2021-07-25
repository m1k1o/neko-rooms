package firefox

// https://github.com/mozilla/policy-templates/blob/master/README.md#homepage

import (
	_ "embed"
	"encoding/json"
	"m1k1o/neko_rooms/internal/types"
)

func Parse(policiesJson string) (*types.Policies, error) {
	policies := types.Policies{}

	policiesTmpl := struct {
		Policies map[string]interface{} `json:"policies"`
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
		policies.Extensions = []types.Extension{}
		for id, val := range extensions.(map[string]interface{}) {
			if id == "*" {
				continue
			}

			data := val.(map[string]interface{})
			url, _ := data["install_url"].(string)

			policies.Extensions = append(
				policies.Extensions,
				types.Extension{
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
