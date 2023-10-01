package types

import "time"

type PullStart struct {
	NekoImage    string `json:"neko_image"`
	RegistryUser string `json:"registry_user"`
	RegistryPass string `json:"registry_pass"`
}

type PullLayer struct {
	Status         string `json:"status"`
	ProgressDetail *struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
	Progress string `json:"progress"`
	ID       string `json:"id"`
}

type PullStatus struct {
	Active   bool        `json:"active"`
	Started  *time.Time  `json:"started"`
	Layers   []PullLayer `json:"layers"`
	Status   []string    `json:"status"`
	Finished *time.Time  `json:"finished"`
}

type PullManager interface {
	Start(request PullStart) error
	Stop() error
	Status() PullStatus
	Subscribe(ch chan<- string) func()
}
