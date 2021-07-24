package firefox

import (
	_ "embed"
	"encoding/json"
	"net/url"
)

type Settings struct {
	Bookmarks      []string          `json:"bookmarks"`
	Extensions     map[string]string `json:"extensions"`
	DeveloperTools bool              `json:"developer_tools"`
}

//go:embed policies.json
var policiesJson string

func Generate(settings Settings) (string, error) {
	policiesObj := struct {
		Policies map[string]interface{} `json:"policies"`
	}{}
	if err := json.Unmarshal([]byte(policiesJson), &policiesObj); err != nil {
		return "", err
	}

	//
	// Bookmarks
	//

	// Do not append
	//Bookmarks := policiesObj.Policies["Bookmarks"].([]interface{})
	Bookmarks := []interface{}{}

	for _, URL := range settings.Bookmarks {
		title := URL

		if u, err := url.Parse(URL); err == nil {
			title = u.Hostname()
		}

		Bookmarks = append(Bookmarks, map[string]interface{}{
			"Title": title,
			"URL":   URL,
			//"Favicon":   URL + "/favicon.ico",
			"Folder":    "Pages",
			"Placement": "toolbar",
		})
	}

	if len(settings.Bookmarks) > 0 {
		policiesObj.Policies["Bookmarks"] = Bookmarks
	}

	//
	// Extensions
	//

	// Do not append
	//Extensions := policiesObj.Policies["ExtensionSettings"].(map[string]interface{})
	Extensions := map[string]interface{}{}

	// block all
	Extensions["*"] = map[string]interface{}{
		"installation_mode": "blocked",
	}

	for id, url := range settings.Extensions {
		Extensions[id] = map[string]interface{}{
			"install_url":       url,
			"installation_mode": "force_installed",
		}
	}

	if len(settings.Extensions) > 0 {
		policiesObj.Policies["ExtensionSettings"] = Extensions
	}

	//
	// DeveloperTools
	//

	policiesObj.Policies["DisableDeveloperTools"] = settings.DeveloperTools

	data, err := json.MarshalIndent(policiesObj, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
