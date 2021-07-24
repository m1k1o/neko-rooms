package types

type Policies struct {
	Homepage       string      `json:"homepage"`
	Extensions     []Extension `json:"extensions"`
	DeveloperTools bool        `json:"developer_tools"`
	PersistentData bool        `json:"persistent_data"`
}

type Extension struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}
