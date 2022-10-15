package chromium

// https://chromeenterprise.google/policies/

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/m1k1o/neko-rooms/internal/types"
)

//go:embed policies.json
var policiesJson string

func Generate(policies types.BrowserPolicyContent) (string, error) {
	policiesTmpl := map[string]interface{}{}
	if err := json.Unmarshal([]byte(policiesJson), &policiesTmpl); err != nil {
		return "", err
	}

	//
	// Extensions
	//

	ExtensionInstallForcelist := []interface{}{}
	for _, e := range policies.Extensions {
		URL := e.URL
		if URL == "" {
			URL = "https://clients2.google.com/service/update2/crx"
		}

		ExtensionInstallForcelist = append(
			ExtensionInstallForcelist,
			fmt.Sprintf("%s;%s", e.ID, URL),
		)
	}

	ExtensionInstallAllowlist := []interface{}{}
	for _, e := range policies.Extensions {
		ExtensionInstallAllowlist = append(
			ExtensionInstallAllowlist,
			e.ID,
		)
	}

	policiesTmpl["ExtensionInstallForcelist"] = ExtensionInstallForcelist
	policiesTmpl["ExtensionInstallAllowlist"] = ExtensionInstallAllowlist
	policiesTmpl["ExtensionInstallBlocklist"] = []interface{}{"*"}

	//
	// Developer Tools
	//

	if policies.DeveloperTools {
		// Allow usage of the Developer Tools
		policiesTmpl["DeveloperToolsAvailability"] = 1
	} else {
		// Disallow usage of the Developer Tools
		policiesTmpl["DeveloperToolsAvailability"] = 2
	}

	//
	// Persistent Data
	//

	if policies.PersistentData {
		// Allow all sites to set local data
		policiesTmpl["DefaultCookiesSetting"] = 1
		// Restore the last session
		policiesTmpl["RestoreOnStartup"] = 1
	} else {
		// Keep cookies for the duration of the session
		policiesTmpl["DefaultCookiesSetting"] = 4
		// Open New Tab Page
		policiesTmpl["RestoreOnStartup"] = 5
	}

	data, err := json.MarshalIndent(policiesTmpl, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
