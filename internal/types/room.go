package types

import (
	"fmt"
	"strings"
)

type RoomSettings struct {
	MaxConnections uint   `json:"max_connections"`
	UserPass       string `json:"user_pass"`
	AdminPass      string `json:"admin_pass"`

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

func (settings *RoomSettings) Env(epr_start uint, epr_end uint, nat1to1 []string) []string {
	env := []string{
		fmt.Sprintf("NEKO_EPR=%d-%d", epr_start, epr_end),
		fmt.Sprintf("NEKO_NAT1TO1=%s", strings.Join(nat1to1, ",")),
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

type RoomData struct {
	ID string `json:"id"`

	RoomSettings
}

type RoomManager interface {
	List() ([]RoomData, error)
	Create(settings RoomSettings) (*RoomData, error)
	Get(id string) (*RoomData, error)
	Update(id string, settings RoomSettings) error
	Remove(id string) error
}
