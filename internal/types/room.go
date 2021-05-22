package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"m1k1o/neko_rooms/internal/utils"
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
	"NEKO_NAT1TO1",
	"NEKO_ICELITE",
}

type RoomsConfig struct {
	Connections uint16   `json:"connections"`
	NekoImages  []string `json:"neko_images"`
}

type RoomEntry struct {
	ID             string    `json:"id"`
	URL            string    `json:"url"`
	Name           string    `json:"name"`
	NekoImage      string    `json:"neko_image"`
	MaxConnections uint16    `json:"max_connections"`
	Running        bool      `json:"running"`
	Status         string    `json:"status"`
	Created        time.Time `json:"created"`
}

type MountType string

const (
	MountPrivate  MountType = "private"
	MountTemplate MountType = "template"
	MountPublic   MountType = "public"
)

type RoomMount struct {
	Type          MountType `json:"type"`
	HostPath      string    `json:"host_path"`
	ContainerPath string    `json:"container_path"`
}

type RoomSettings struct {
	Name           string `json:"name"`
	NekoImage      string `json:"neko_image"`
	MaxConnections uint16 `json:"max_connections"`

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

	Envs   map[string]string `json:"envs"`
	Mounts []RoomMount       `json:"mounts"`
}

func (settings *RoomSettings) ToEnv() []string {
	env := []string{
		fmt.Sprintf("NEKO_PASSWORD=%s", settings.UserPass),
		fmt.Sprintf("NEKO_PASSWORD_ADMIN=%s", settings.AdminPass),
		fmt.Sprintf("NEKO_SCREEN=%s", settings.Screen),
		fmt.Sprintf("NEKO_MAX_FPS=%d", settings.VideoMaxFPS),
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
