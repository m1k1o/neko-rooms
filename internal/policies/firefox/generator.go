package firefox

// https://github.com/mozilla/policy-templates/blob/master/README.md#homepage

import (
	_ "embed"
	"encoding/json"
	"m1k1o/neko_rooms/internal/types"
)

//go:embed policies.json
var policiesJson string

func Generate(policies types.Policies) (string, error) {
	policiesTmpl := struct {
		Policies map[string]interface{} `json:"policies"`
	}{}
	if err := json.Unmarshal([]byte(policiesJson), &policiesTmpl); err != nil {
		return "", err
	}

	//
	// Homepage
	//

	if policies.Homepage != "" {
		policiesTmpl.Policies["Homepage"] = map[string]interface{}{
			"URL":       policies.Homepage,
			"StartPage": "homepage",
		}
	}

	//
	// Extensions
	//

	ExtensionSettings := map[string]interface{}{}
	ExtensionSettings["*"] = map[string]interface{}{
		"installation_mode": "blocked",
	}

	for _, e := range policies.Extensions {
		ExtensionSettings[e.ID] = map[string]interface{}{
			"install_url":       e.URL,
			"installation_mode": "force_installed",
		}
	}

	policiesTmpl.Policies["ExtensionSettings"] = ExtensionSettings

	//
	// Developer Tools
	//

	policiesTmpl.Policies["DisableDeveloperTools"] = !policies.DeveloperTools

	//
	// Persistent Data
	//

	if policies.PersistentData {
		Preferences := policiesTmpl.Policies["Preferences"].(map[string]interface{})
		Preferences["browser.urlbar.suggest.history"] = true
		Preferences["places.history.enabled"] = true
		policiesTmpl.Policies["Preferences"] = Preferences
		policiesTmpl.Policies["SanitizeOnShutdown"] = false
	} else {
		policiesTmpl.Policies["SanitizeOnShutdown"] = true
	}

	data, err := json.MarshalIndent(policiesTmpl, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
