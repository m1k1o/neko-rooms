package policies

import (
	"errors"
	"m1k1o/neko_rooms/internal/policies/chromium"
	"m1k1o/neko_rooms/internal/policies/firefox"
	"m1k1o/neko_rooms/internal/types"
)

func Generate(policies types.BrowserPolicyContent, policyType types.BrowserPolicyType) (string, error) {
	if policyType == types.ChromiumBrowserPolicy {
		return chromium.Generate(policies)
	}

	if policyType == types.FirefoxBrowserPolicy {
		return firefox.Generate(policies)
	}

	return "", errors.New("unknown policy type")
}

func Parse(policiesJson string, policyType types.BrowserPolicyType) (*types.BrowserPolicyContent, error) {
	if policyType == types.ChromiumBrowserPolicy {
		return chromium.Parse(policiesJson)
	}

	if policyType == types.FirefoxBrowserPolicy {
		return firefox.Parse(policiesJson)
	}

	return nil, errors.New("unknown policy type")
}
