package types

import (
	"fmt"
	"strings"
	"time"
)

type RoomEntry struct {
	ID             string    `json:"id"`
	URL            string    `json:"url"`
	Name           string    `json:"name"`
	MaxConnections uint16    `json:"max_connections"`
	Image          string    `json:"image"`
	State          string    `json:"state"`
	Status         string    `json:"status"`
	Created        time.Time `json:"created"`
}

type RoomSettings struct {
	Name           string `json:"name"`
	MaxConnections uint16 `json:"max_connections"`

	UserPass  string `json:"user_pass"`
	AdminPass string `json:"admin_pass"`

	BroadcastPipeline string `json:"broadcast_pipeline"`

	Screen        string `json:"screen"`
	VideoCodec    string `json:"video_codec"`
	VideoBitrate  uint   `json:"video_bitrate"`
	VideoPipeline string `json:"video_pipeline"`
	VideoMaxFPS   uint   `json:"video_max_fps"`

	AudioCodec    string `json:"audio_codec"`
	AudioBitrate  uint   `json:"audio_bitrate"`
	AudioPipeline string `json:"audio_pipeline"`
}

func (settings *RoomSettings) ToEnv() []string {
	env := []string{
		fmt.Sprintf("NEKO_PASSWORD=%s", settings.UserPass),
		fmt.Sprintf("NEKO_PASSWORD_ADMIN=%s", settings.AdminPass),
		fmt.Sprintf("NEKO_SCREEN=%s", settings.Screen),
		fmt.Sprintf("NEKO_MAX_FPS=%d", settings.VideoMaxFPS),
	}

	if settings.BroadcastPipeline != "" {
		env = append(env, fmt.Sprintf("NEKO_BROADCAST_PIPELINE=%s", settings.BroadcastPipeline))
	}

	if settings.VideoCodec != "" {
		env = append(env, fmt.Sprintf("NEKO_%s=true", strings.ToUpper(settings.VideoCodec)))
	}

	if settings.VideoBitrate != 0 {
		env = append(env, fmt.Sprintf("NEKO_VIDEO_BITRATE=%d", settings.VideoBitrate))
	}

	if settings.VideoPipeline != "" {
		env = append(env, fmt.Sprintf("NEKO_VIDEO=%s", settings.VideoPipeline))
	}

	if settings.AudioCodec != "" {
		env = append(env, fmt.Sprintf("NEKO_%s=true", strings.ToUpper(settings.AudioCodec)))
	}

	if settings.AudioBitrate != 0 {
		env = append(env, fmt.Sprintf("NEKO_AUDIO_BITRATE=%d", settings.AudioBitrate))
	}

	if settings.AudioPipeline != "" {
		env = append(env, fmt.Sprintf("NEKO_AUDIO=%s", settings.AudioPipeline))
	}

	return env
}

type RoomManager interface {
	List() ([]RoomEntry, error)
	Create(settings RoomSettings) (string, error)
	Get(id string) (*RoomSettings, error)
	Update(id string, settings RoomSettings) error
	Remove(id string) error
}
