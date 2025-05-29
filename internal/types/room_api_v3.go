package types

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/m1k1o/neko-rooms/internal/config"
)

//
// demodesk/neko v3 envs API
//

var blacklistedEnvsV3 = []string{
	// ignore bunch of default envs
	"DEBIAN_FRONTEND",
	"PULSE_SERVER",
	"XDG_RUNTIME_DIR",
	"DISPLAY",
	"USER",
	"PATH",

	// ignore bunch of envs managed by neko
	"NEKO_PLUGINS_ENABLED",
	"NEKO_PLUGINS_DIR",

	// ignore bunch of envs managed by neko-rooms
	"NEKO_SERVER_BIND",
	"NEKO_SERVER_PROXY",
	"NEKO_SESSION_API_TOKEN",
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
		"NEKO_SERVER_PROXY=true",

		// from settings
		"NEKO_MEMBER_PROVIDER=multiuser",
		fmt.Sprintf("NEKO_MEMBER_MULTIUSER_USER_PASSWORD=%s", settings.UserPass),
		fmt.Sprintf("NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD=%s", settings.AdminPass),
		fmt.Sprintf("NEKO_SESSION_API_TOKEN=%s", settings.AdminPass), // TODO: should be random and saved somewhere
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

	if settings.ControlProtection {
		env = append(env, "NEKO_SESSION_CONTROL_PROTECTION=true")
	}

	// implicit control - enabled by default but in legacy mode disabled by default
	// so we need to set it explicitly until legacy mode is removed
	if !settings.ImplicitControl {
		env = append(env, "NEKO_SESSION_IMPLICIT_HOSTING=false")
	} else {
		env = append(env, "NEKO_SESSION_IMPLICIT_HOSTING=true")
	}

	if settings.VideoCodec != "VP8" { // VP8 is default
		env = append(env, fmt.Sprintf("NEKO_CAPTURE_VIDEO_CODEC=%s", strings.ToLower(settings.VideoCodec)))
	}

	//if settings.VideoBitrate != 0 {
	//	env = append(env, fmt.Sprintf("NEKO_VIDEO_BITRATE=%d", settings.VideoBitrate)) // TODO: not supported yet
	//}

	if settings.VideoPipeline != "" {
		env = append(env, fmt.Sprintf("NEKO_CAPTURE_VIDEO_PIPELINE=%s", settings.VideoPipeline)) // TOOD: allow simulcast pipelines
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
		if !slices.Contains(blacklistedEnvsV3, key) {
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
		case "NEKO_SESSION_CONTROL_PROTECTION":
			settings.ControlProtection, err = strconv.ParseBool(val)
		case "NEKO_SESSION_IMPLICIT_HOSTING":
			settings.ImplicitControl, err = strconv.ParseBool(val)
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
		case "NEKO_CAPTURE_VIDEO_PIPELINE": // TOOD: allow simulcast pipelines
			settings.VideoPipeline = val
		case "NEKO_CAPTURE_AUDIO_CODEC":
			settings.AudioCodec = strings.ToUpper(val)
		//case "NEKO_AUDIO_BITRATE": // TODO: not supported yet
		//	settings.AudioBitrate, err = strconv.Atoi(val)
		case "NEKO_CAPTURE_AUDIO_PIPELINE":
			settings.AudioPipeline = val
		default:
			if !slices.Contains(blacklistedEnvsV3, key) {
				settings.Envs[key] = val
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}
