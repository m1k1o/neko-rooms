package firefox

// https://github.com/mozilla/policy-templates/blob/master/README.md#homepage

import (
	_ "embed"
	"encoding/json"

	"github.com/m1k1o/neko-rooms/internal/types"
)

//go:embed policies.json
var policiesJson string

func Generate(policies types.BrowserPolicyContent) (string, error) {
	policiesTmpl := struct {
		Policies map[string]interface{} `json:"policies"`
	}{}
	if err := json.Unmarshal([]byte(policiesJson), &policiesTmpl); err != nil {
		return "", err
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

	Preferences := policiesTmpl.Policies["Preferences"].(map[string]interface{})
	Preferences["browser.urlbar.suggest.history"] = policies.PersistentData
	Preferences["places.history.enabled"] = policies.PersistentData
	policiesTmpl.Policies["Preferences"] = Preferences
	policiesTmpl.Policies["SanitizeOnShutdown"] = !policies.PersistentData

	if policies.PersistentData {
		policiesTmpl.Policies["Homepage"] = map[string]interface{}{
			"StartPage": "previous-session",
		}
	} else {
		policiesTmpl.Policies["Homepage"] = map[string]interface{}{
			"StartPage": "homepage",
		}
	}

	data, err := json.MarshalIndent(policiesTmpl, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
