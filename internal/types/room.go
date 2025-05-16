package types

import (
	"context"
	"fmt"
	"time"

	"github.com/m1k1o/neko-rooms/internal/config"
)

type RoomsConfig struct {
	Connections    uint16   `json:"connections"`
	NekoImages     []string `json:"neko_images"`
	StorageEnabled bool     `json:"storage_enabled"`
	UsesMux        bool     `json:"uses_mux"`
}

type RoomEntry struct {
	ID             string            `json:"id"`
	URL            string            `json:"url"`
	Name           string            `json:"name"`
	NekoImage      string            `json:"neko_image"`
	IsOutdated     bool              `json:"is_outdated"`
	MaxConnections uint16            `json:"max_connections"` // 0 when using mux
	Running        bool              `json:"running"`
	Paused         bool              `json:"paused"`
	IsReady        bool              `json:"is_ready"`
	Status         string            `json:"status"`
	Created        time.Time         `json:"created"`
	Labels         map[string]string `json:"labels,omitempty"`

	ContainerLabels map[string]string `json:"-"` // for internal use
}

type MountType string

const (
	MountPrivate   MountType = "private"
	MountTemplate  MountType = "template"
	MountProtected MountType = "protected"
	MountPublic    MountType = "public"
)

type RoomMount struct {
	Type          MountType `json:"type"`
	HostPath      string    `json:"host_path"`
	ContainerPath string    `json:"container_path"`
}

type RoomResources struct {
	CPUShares int64    `json:"cpu_shares"` // relative weight vs. other containers
	NanoCPUs  int64    `json:"nano_cpus"`  // in units of 10^-9 CPUs
	ShmSize   int64    `json:"shm_size"`   // in bytes
	Memory    int64    `json:"memory"`     // in bytes
	Gpus      []string `json:"gpus"`       // gpu opts
	Devices   []string `json:"devices"`
}

type RoomSettings struct {
	ApiVersion int `json:"api_version"`

	Name           string `json:"name"`
	NekoImage      string `json:"neko_image"`
	MaxConnections uint16 `json:"max_connections"` // 0 when using mux

	ControlProtection bool `json:"control_protection"`
	ImplicitControl   bool `json:"implicit_control"`

	UserPass  string `json:"user_pass"`
	AdminPass string `json:"admin_pass"`

	Screen        string `json:"screen"`
	VideoCodec    string `json:"video_codec,omitempty"`
	VideoBitrate  int    `json:"video_bitrate,omitempty"`
	VideoPipeline string `json:"video_pipeline,omitempty"`
	VideoMaxFPS   int    `json:"video_max_fps"`

	AudioCodec    string `json:"audio_codec,omitempty"`
	AudioBitrate  int    `json:"audio_bitrate,omitempty"`
	AudioPipeline string `json:"audio_pipeline,omitempty"`

	BroadcastPipeline string `json:"broadcast_pipeline,omitempty"`

	Envs      map[string]string `json:"envs"`
	Labels    map[string]string `json:"labels"`
	Mounts    []RoomMount       `json:"mounts"`
	Resources RoomResources     `json:"resources"`

	Hostname string   `json:"hostname,omitempty"`
	DNS      []string `json:"dns,omitempty"`

	BrowserPolicy *BrowserPolicy `json:"browser_policy,omitempty"`
}

func (settings *RoomSettings) ToEnv(config *config.Room, ports PortSettings) ([]string, error) {
	switch settings.ApiVersion {
	case 2:
		return settings.toEnvV2(config, ports), nil
	case 3:
		return settings.toEnvV3(config, ports), nil
	default:
		return nil, fmt.Errorf("unsupported API version: %d", settings.ApiVersion)
	}
}

func (settings *RoomSettings) FromEnv(apiVersion int, envs []string) error {
	switch apiVersion {
	case 2:
		return settings.fromEnvV2(envs)
	case 3:
		return settings.fromEnvV3(envs)
	default:
		return fmt.Errorf("unsupported API version: %d", apiVersion)
	}
}

type PortSettings struct {
	FrontendPort   uint16
	EprMin, EprMax uint16
}

type RoomStats struct {
	Connections uint32        `json:"connections"`
	Host        string        `json:"host"`
	Members     []*RoomMember `json:"members"`

	Banned map[string]string `json:"banned"` // IP -> session ID (that banned it)
	Locked map[string]string `json:"locked"` // resource name -> session ID (that locked it)

	ServerStartedAt time.Time  `json:"server_started_at"`
	LastAdminLeftAt *time.Time `json:"last_admin_left_at"`
	LastUserLeftAt  *time.Time `json:"last_user_left_at"`

	ControlProtection bool `json:"control_protection"`
	ImplicitControl   bool `json:"implicit_control"`
}

type RoomMember struct {
	ID    string `json:"id"`
	Name  string `json:"displayname"`
	Admin bool   `json:"admin"`
	Muted bool   `json:"muted"`
}

type RoomEventAction string

const (
	RoomEventCreated   RoomEventAction = "created"
	RoomEventStarted   RoomEventAction = "started"
	RoomEventReady     RoomEventAction = "ready"
	RoomEventStopped   RoomEventAction = "stopped"
	RoomEventDestroyed RoomEventAction = "destroyed"
	RoomEventPaused    RoomEventAction = "paused"
)

type RoomEvent struct {
	ID     string          `json:"id"`
	Action RoomEventAction `json:"action"`

	ContainerLabels map[string]string `json:"-"` // for internal use
}

var ErrRoomNotFound = fmt.Errorf("room not found")

type RoomManager interface {
	Config() RoomsConfig
	List(ctx context.Context, labels map[string]string) ([]RoomEntry, error)
	ExportAsDockerCompose(ctx context.Context) ([]byte, error)

	Create(ctx context.Context, settings RoomSettings) (string, error)
	GetEntry(ctx context.Context, id string) (*RoomEntry, error)
	GetEntryByName(ctx context.Context, name string) (*RoomEntry, error)
	GetSettings(ctx context.Context, id string) (*RoomSettings, error)
	GetStats(ctx context.Context, id string) (*RoomStats, error)
	Remove(ctx context.Context, id string) error

	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Restart(ctx context.Context, id string) error
	Pause(ctx context.Context, id string) error

	EventsLoopStart()
	EventsLoopStop() error
	Events(ctx context.Context) (<-chan RoomEvent, <-chan error)
}
