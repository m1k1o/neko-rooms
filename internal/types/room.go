package types

type RoomSettings struct {
	MaxConnections uint   `json:"max_connections"`
	UserPass       string `json:"user_pass"`
	AdminPass      string `json:"admin_pass"`

	ScreenWidth  string `json:"screen_width"`
	ScreenHeight string `json:"screen_height"`
	ScreenRate   string `json:"screen_rate"`

	BroadcastPipeline string `json:"broadcast_pipeline"`

	VideoCodec    string `json:"video_codec"`
	VideoBitrate  uint   `json:"video_bitrate"`
	VideoPipeline string `json:"video_pipeline"`
	VideoFPSMax   uint   `json:"video_fps_max"`

	AudioCodec    string `json:"audio_codec"`
	AudioBitrate  uint   `json:"audio_bitrate"`
	AudioPipeline string `json:"audio_pipeline"`
}

type RoomData struct {
	ID   string `json:"id"`
	RoomSettings
}

type RoomManager interface {
}
