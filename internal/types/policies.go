package types

type BrowserPolicyType string

const (
	ChromiumBrowserPolicy BrowserPolicyType = "chromium"
	FirefoxBrowserPolicy  BrowserPolicyType = "firefox"
)

type BrowserPolicy struct {
	Type    BrowserPolicyType    `json:"type"`
	Path    string               `json:"path"`
	Content BrowserPolicyContent `json:"content"`
}

type BrowserPolicyContent struct {
	Extensions     []BrowserPolicyExtension `json:"extensions"`
	DeveloperTools bool                     `json:"developer_tools"`
	PersistentData bool                     `json:"persistent_data"`
}

type BrowserPolicyExtension struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}
