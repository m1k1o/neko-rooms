package policies

import (
	"errors"
	"m1k1o/neko_rooms/internal/policies/chromium"
	"m1k1o/neko_rooms/internal/policies/firefox"
	"m1k1o/neko_rooms/internal/types"
)

type PolicyType string

const (
	Chromium PolicyType = "chromium"
	Firefox  PolicyType = "friefox"
)

type PoliciesConfig struct {
	Type PolicyType
	Path string
}

func GetConfig(image string) *PoliciesConfig {
	switch image {
	case "m1k1o/neko:latest":
		fallthrough
	case "m1k1o/neko:firefox":
		return &PoliciesConfig{
			Type: Firefox,
			Path: "/usr/lib/firefox/distribution/policies.json",
		}
	case "m1k1o/neko:arm-firefox":
		return &PoliciesConfig{
			Type: Firefox,
			Path: "/usr/lib/firefox-esr/distribution/policies.json",
		}
	case "m1k1o/neko:chromium":
		fallthrough
	case "m1k1o/neko:arm-chromium":
		fallthrough
	case "m1k1o/neko:ungoogled-chromium":
		return &PoliciesConfig{
			Type: Chromium,
			Path: "/etc/chromium/policies/managed/policies.json",
		}
	case "m1k1o/neko:google-chrome":
		return &PoliciesConfig{
			Type: Chromium,
			Path: "/etc/opt/chrome/policies/managed/policies.json",
		}
	case "m1k1o/neko:brave":
		return &PoliciesConfig{
			Type: Chromium,
			Path: "/etc/brave/policies/managed/policies.json",
		}
	}

	return nil
}

func Generate(policies types.Policies, policyType PolicyType) (string, error) {
	if policyType == Chromium {
		return chromium.Generate(policies)
	}

	if policyType == Firefox {
		return firefox.Generate(policies)
	}

	return "", errors.New("unknown policy type")
}

func Parse(policiesJson string, policyType PolicyType) (*types.Policies, error) {
	if policyType == Chromium {
		return chromium.Parse(policiesJson)
	}

	if policyType == Firefox {
		return firefox.Parse(policiesJson)
	}

	return nil, errors.New("unknown policy type")
}
