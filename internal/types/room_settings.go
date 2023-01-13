package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/m1k1o/neko-rooms/internal/config"
	"github.com/m1k1o/neko-rooms/internal/utils"
)

func (settings *RoomSettings) ToEnv(config *config.Room, ports PortSettings) ([]string, error) {
	switch config.ApiVersion {
	case 2:
		return settings.toEnvV2(config, ports), nil
	case 3:
		return settings.toEnvV3(config, ports), nil
	default:
		return nil, fmt.Errorf("unsupported API version: %d", config.ApiVersion)
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

//
// m1k1o/neko v2 envs API
//

var blacklistedEnvsV2 = []string{
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

func (settings *RoomSettings) toEnvV2(config *config.Room, ports PortSettings) []string {
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

	if settings.VideoCodec != "VP8" { // VP8 is default
		env = append(env, fmt.Sprintf("NEKO_VIDEO_CODEC=%s", strings.ToLower(settings.VideoCodec)))
	}

	if settings.VideoBitrate != 0 {
		env = append(env, fmt.Sprintf("NEKO_VIDEO_BITRATE=%d", settings.VideoBitrate))
	}

	if settings.VideoPipeline != "" {
		env = append(env, fmt.Sprintf("NEKO_VIDEO=%s", settings.VideoPipeline))
	}

	if settings.AudioCodec != "OPUS" { // OPUS is default
		env = append(env, fmt.Sprintf("NEKO_AUDIO_CODEC=%s", strings.ToLower(settings.AudioCodec)))
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
		if in, _ := utils.ArrayIn(key, blacklistedEnvsV2); !in {
			env = append(env, fmt.Sprintf("%s=%s", key, val))
		}
	}

	return env
}

func (settings *RoomSettings) fromEnvV2(envs []string) error {
	settings.Envs = map[string]string{}
	settings.VideoCodec = "VP8"  // default
	settings.AudioCodec = "OPUS" // default

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
		case "NEKO_VIDEO_CODEC":
			settings.VideoCodec = strings.ToUpper(val)
		case "NEKO_VIDEO_BITRATE":
			settings.VideoBitrate, err = strconv.Atoi(val)
		case "NEKO_VIDEO":
			settings.VideoPipeline = val
		case "NEKO_AUDIO_CODEC":
			settings.VideoCodec = strings.ToUpper(val)
		case "NEKO_AUDIO_BITRATE":
			settings.AudioBitrate, err = strconv.Atoi(val)
		case "NEKO_AUDIO":
			settings.AudioPipeline = val
		default:
			if in, _ := utils.ArrayIn(key, blacklistedEnvsV2); !in {
				settings.Envs[key] = val
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}

//
// demodesk/neko v3 envs API
//

var blacklistedEnvsV3 = []string{
	// ignore bunch of default envs
	"DEBIAN_FRONTEND",
	"DISPLAY",
	"USER",
	"PATH",

	// ignore bunch of envs managed by neko
	"NEKO_PLUGINS_ENABLED",
	"NEKO_PLUGINS_DIR",

	// ignore bunch of envs managed by neko-rooms
	"NEKO_SERVER_BIND",
	"NEKO_MEMBER_PROVIDER",
	"NEKO_WEBRTC_EPR",
	"NEKO_WEBRTC_UDPMUX",
	"NEKO_WEBRTC_TCPMUX",
	"NEKO_WEBRTC_NAT1TO1",
	"NEKO_WEBRTC_ICELITE",
}

func (settings *RoomSettings) toEnvV3(config *config.Room, ports PortSettings) []string {
	env := []string{
		fmt.Sprintf("NEKO_SERVER_BIND=:%d", ports.FrontendPort),
		"NEKO_WEBRTC_ICELITE=true",

		// from settings
		"NEKO_MEMBER_PROVIDER=multiuser",
		fmt.Sprintf("NEKO_MEMBER_MULTIUSER_USER_PASSWORD=%s", settings.UserPass),
		fmt.Sprintf("NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD=%s", settings.AdminPass),
		fmt.Sprintf("NEKO_DESKTOP_SCREEN=%s", settings.Screen),
		//fmt.Sprintf("NEKO_MAX_FPS=%d", settings.VideoMaxFPS), // TODO: not supported yet
	}

	if config.Mux {
		env = append(env,
			fmt.Sprintf("NEKO_WEBRTC_UDPMUX=%d", ports.EprMin),
			fmt.Sprintf("NEKO_WEBRTC_TCPMUX=%d", ports.EprMin),
		)
	} else {
		env = append(env,
			fmt.Sprintf("NEKO_WEBRTC_EPR=%d-%d", ports.EprMin, ports.EprMax),
		)
	}

	// optional nat mapping
	if len(config.NAT1To1IPs) > 0 {
		env = append(env, fmt.Sprintf("NEKO_WEBRTC_NAT1TO1=%s", strings.Join(config.NAT1To1IPs, ",")))
	}

	//if settings.ControlProtection {
	//	env = append(env, "NEKO_CONTROL_PROTECTION=true") // TODO: not supported yet
	//}

	// implicit control - enabled by default
	if !settings.ImplicitControl {
		env = append(env, "NEKO_SESSION_IMPLICIT_HOSTING=false")
	}

	if settings.VideoCodec != "VP8" { // VP8 is default
		env = append(env, fmt.Sprintf("NEKO_CAPTURE_VIDEO_CODEC=%s", strings.ToLower(settings.VideoCodec)))
	}

	//if settings.VideoBitrate != 0 {
	//	env = append(env, fmt.Sprintf("NEKO_VIDEO_BITRATE=%d", settings.VideoBitrate)) // TODO: not supported yet
	//}

	if settings.VideoPipeline != "" {
		env = append(env, fmt.Sprintf("NEKO_CAPTURE_VIDEO_PIPELINES=%s", settings.VideoPipeline)) // TOOD: multiple pipelines, as JSON
	}

	if settings.AudioCodec != "OPUS" { // OPUS is default
		env = append(env, fmt.Sprintf("NEKO_CAPTURE_AUDIO_CODEC=%s", strings.ToLower(settings.AudioCodec)))
	}

	//if settings.AudioBitrate != 0 {
	//	env = append(env, fmt.Sprintf("NEKO_AUDIO_BITRATE=%d", settings.AudioBitrate)) // TODO: not supported yet
	//}

	if settings.AudioPipeline != "" {
		env = append(env, fmt.Sprintf("NEKO_CAPTURE_AUDIO_PIPELINE=%s", settings.AudioPipeline))
	}

	if settings.BroadcastPipeline != "" {
		env = append(env, fmt.Sprintf("NEKO_CAPTURE_BROADCAST_PIPELINE=%s", settings.BroadcastPipeline))
	}

	for key, val := range settings.Envs {
		if in, _ := utils.ArrayIn(key, blacklistedEnvsV3); !in {
			env = append(env, fmt.Sprintf("%s=%s", key, val))
		}
	}

	return env
}

func (settings *RoomSettings) fromEnvV3(envs []string) error {
	settings.Envs = map[string]string{}
	// enabled implicit control by default
	settings.ImplicitControl = true
	settings.VideoCodec = "VP8"  // default
	settings.AudioCodec = "OPUS" // default

	var err error
	for _, env := range envs {
		r := strings.SplitN(env, "=", 2)
		key, val := r[0], r[1]

		switch key {
		case "NEKO_MEMBER_MULTIUSER_USER_PASSWORD":
			settings.UserPass = val
		case "NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD":
			settings.AdminPass = val
		//case "NEKO_CONTROL_PROTECTION": // TODO: not supported yet
		//	if ok, _ := strconv.ParseBool(val); ok {
		//		settings.ControlProtection = true
		//	}
		case "NEKO_SESSION_IMPLICIT_HOSTING":
			if ok, _ := strconv.ParseBool(val); !ok {
				settings.ImplicitControl = false
			}
		case "NEKO_DESKTOP_SCREEN":
			settings.Screen = val
		//case "NEKO_MAX_FPS": // TODO: not supported yet
		//	settings.VideoMaxFPS, err = strconv.Atoi(val)
		case "NEKO_CAPTURE_BROADCAST_PIPELINE":
			settings.BroadcastPipeline = val
		case "NEKO_CAPTURE_VIDEO_CODEC":
			settings.VideoCodec = strings.ToUpper(val)
		//case "NEKO_VIDEO_BITRATE": // TODO: not supported yet
		//	settings.VideoBitrate, err = strconv.Atoi(val)
		case "NEKO_CAPTURE_VIDEO_PIPELINES": // TOOD: multiple pipelines, as JSON
			settings.VideoPipeline = val
		case "NEKO_CAPTURE_AUDIO_CODEC":
			settings.AudioCodec = strings.ToUpper(val)
		//case "NEKO_AUDIO_BITRATE": // TODO: not supported yet
		//	settings.AudioBitrate, err = strconv.Atoi(val)
		case "NEKO_CAPTURE_AUDIO_PIPELINE":
			settings.AudioPipeline = val
		default:
			if in, _ := utils.ArrayIn(key, blacklistedEnvsV3); !in {
				settings.Envs[key] = val
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}
