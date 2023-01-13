package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/m1k1o/neko-rooms/internal/config"
	"github.com/m1k1o/neko-rooms/internal/utils"
)

var blacklistedEnvs = []string{
	// ignore bunch of default envs
	"DEBIAN_FRONTEND",
	"DISPLAY",
	"USER",
	"PATH",

	// ignore bunch of envs managed by neko-rooms
	"NEKO_BIND",
	"NEKO_EPR",
	"NEKO_UDPMUX",
	"NEKO_TCPMUX",
	"NEKO_NAT1TO1",
	"NEKO_ICELITE",
}

type RoomsConfig struct {
	Connections    uint16   `json:"connections"`
	NekoImages     []string `json:"neko_images"`
	StorageEnabled bool     `json:"storage_enabled"`
	UsesMux        bool     `json:"uses_mux"`
}

type RoomEntry struct {
	ID             string    `json:"id"`
	URL            string    `json:"url"`
	Name           string    `json:"name"`
	NekoImage      string    `json:"neko_image"`
	IsOutdated     bool      `json:"is_outdated"`
	MaxConnections uint16    `json:"max_connections"` // 0 when using mux
	Running        bool      `json:"running"`
	Status         string    `json:"status"`
	Created        time.Time `json:"created"`
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
}

type RoomSettings struct {
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
	Mounts    []RoomMount       `json:"mounts"`
	Resources RoomResources     `json:"resources"`

	BrowserPolicy *BrowserPolicy `json:"browser_policy,omitempty"`
}

type PortSettings struct {
	FrontendPort   uint16
	EprMin, EprMax uint16
}

func (settings *RoomSettings) ToEnv(config *config.Room, ports PortSettings) []string {
	env := []string{
		fmt.Sprintf("NEKO_BIND=:%d", ports.FrontendPort),
		"NEKO_ICELITE=true",

		// from settings
		fmt.Sprintf("NEKO_PASSWORD=%s", settings.UserPass),
		fmt.Sprintf("NEKO_PASSWORD_ADMIN=%s", settings.AdminPass),
		fmt.Sprintf("NEKO_SCREEN=%s", settings.Screen),
		fmt.Sprintf("NEKO_MAX_FPS=%d", settings.VideoMaxFPS),
	}

	if config.Mux {
		env = append(env,
			fmt.Sprintf("NEKO_UDPMUX=%d", ports.EprMin),
			fmt.Sprintf("NEKO_TCPMUX=%d", ports.EprMin),
		)
	} else {
		env = append(env,
			fmt.Sprintf("NEKO_EPR=%d-%d", ports.EprMin, ports.EprMax),
		)
	}

	// optional nat mapping
	if len(config.NAT1To1IPs) > 0 {
		env = append(env, fmt.Sprintf("NEKO_NAT1TO1=%s", strings.Join(config.NAT1To1IPs, ",")))
	}

	if settings.ControlProtection {
		env = append(env, "NEKO_CONTROL_PROTECTION=true")
	}

	if settings.ImplicitControl {
		env = append(env, "NEKO_IMPLICIT_CONTROL=true")
	}

	if settings.VideoCodec == "VP8" || settings.VideoCodec == "VP9" || settings.VideoCodec == "H264" {
		env = append(env, fmt.Sprintf("NEKO_%s=true", strings.ToUpper(settings.VideoCodec)))
	}

	if settings.VideoBitrate != 0 {
		env = append(env, fmt.Sprintf("NEKO_VIDEO_BITRATE=%d", settings.VideoBitrate))
	}

	if settings.VideoPipeline != "" {
		env = append(env, fmt.Sprintf("NEKO_VIDEO=%s", settings.VideoPipeline))
	}

	if settings.AudioCodec == "OPUS" || settings.AudioCodec == "G722" || settings.AudioCodec == "PCMU" || settings.AudioCodec == "PCMA" {
		env = append(env, fmt.Sprintf("NEKO_%s=true", strings.ToUpper(settings.AudioCodec)))
	}

	if settings.AudioBitrate != 0 {
		env = append(env, fmt.Sprintf("NEKO_AUDIO_BITRATE=%d", settings.AudioBitrate))
	}

	if settings.AudioPipeline != "" {
		env = append(env, fmt.Sprintf("NEKO_AUDIO=%s", settings.AudioPipeline))
	}

	if settings.BroadcastPipeline != "" {
		env = append(env, fmt.Sprintf("NEKO_BROADCAST_PIPELINE=%s", settings.BroadcastPipeline))
	}

	for key, val := range settings.Envs {
		if in, _ := utils.ArrayIn(key, blacklistedEnvs); !in {
			env = append(env, fmt.Sprintf("%s=%s", key, val))
		}
	}

	return env
}

func (settings *RoomSettings) FromEnv(envs []string) error {
	settings.Envs = map[string]string{}

	var err error
	for _, env := range envs {
		r := strings.SplitN(env, "=", 2)
		key, val := r[0], r[1]

		switch key {
		case "NEKO_PASSWORD":
			settings.UserPass = val
		case "NEKO_PASSWORD_ADMIN":
			settings.AdminPass = val
		case "NEKO_CONTROL_PROTECTION":
			if ok, _ := strconv.ParseBool(val); ok {
				settings.ControlProtection = true
			}
		case "NEKO_IMPLICIT_CONTROL":
			if ok, _ := strconv.ParseBool(val); ok {
				settings.ImplicitControl = true
			}
		case "NEKO_SCREEN":
			settings.Screen = val
		case "NEKO_MAX_FPS":
			settings.VideoMaxFPS, err = strconv.Atoi(val)
		case "NEKO_BROADCAST_PIPELINE":
			settings.BroadcastPipeline = val
		case "NEKO_VP8":
			if ok, _ := strconv.ParseBool(val); ok {
				settings.VideoCodec = "VP8"
			}
		case "NEKO_VP9":
			if ok, _ := strconv.ParseBool(val); ok {
				settings.VideoCodec = "VP9"
			}
		case "NEKO_H264":
			if ok, _ := strconv.ParseBool(val); ok {
				settings.VideoCodec = "H264"
			}
		case "NEKO_VIDEO_BITRATE":
			settings.VideoBitrate, err = strconv.Atoi(val)
		case "NEKO_VIDEO":
			settings.VideoPipeline = val
		case "NEKO_OPUS":
			if ok, _ := strconv.ParseBool(val); ok {
				settings.AudioCodec = "OPUS"
			}
		case "NEKO_G722":
			if ok, _ := strconv.ParseBool(val); ok {
				settings.AudioCodec = "G722"
			}
		case "NEKO_PCMU":
			if ok, _ := strconv.ParseBool(val); ok {
				settings.AudioCodec = "PCMU"
			}
		case "NEKO_PCMA":
			if ok, _ := strconv.ParseBool(val); ok {
				settings.AudioCodec = "PCMA"
			}
		case "NEKO_AUDIO_BITRATE":
			settings.AudioBitrate, err = strconv.Atoi(val)
		case "NEKO_AUDIO":
			settings.AudioPipeline = val
		default:
			if in, _ := utils.ArrayIn(key, blacklistedEnvs); !in {
				settings.Envs[key] = val
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
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

type RoomManager interface {
	Config() RoomsConfig
	List() ([]RoomEntry, error)
	FindByName(name string) (*RoomEntry, error)

	Create(settings RoomSettings) (string, error)
	GetEntry(id string) (*RoomEntry, error)
	GetSettings(id string) (*RoomSettings, error)
	GetStats(id string) (*RoomStats, error)
	Remove(id string) error

	Start(id string) error
	Stop(id string) error
	Restart(id string) error
}
