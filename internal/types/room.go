package types

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

type RoomData struct {
	ID   string `json:"id"`
	RoomSettings
}

type RoomManager interface {
	List() (*[]RoomData, error)
	Create(settings RoomSettings) (*RoomData, error)
	Get(id string) (*RoomData, error)
	Update(id string, settings RoomSettings) error
	Remove(id string) error
}
