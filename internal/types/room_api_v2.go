package types

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/m1k1o/neko-rooms/internal/config"
)

//
// m1k1o/neko v2 envs API
//

var blacklistedEnvsV2 = []string{
	// ignore bunch of default envs
	"DEBIAN_FRONTEND",
	"PULSE_SERVER",
	"XDG_RUNTIME_DIR",
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
	"NEKO_PROXY",
}

func (settings *RoomSettings) toEnvV2(config *config.Room, ports PortSettings) []string {
	env := []string{
		fmt.Sprintf("NEKO_BIND=:%d", ports.FrontendPort),
		"NEKO_ICELITE=true",
		"NEKO_PROXY=true",

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
		if !slices.Contains(blacklistedEnvsV2, key) {
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
			settings.ControlProtection, err = strconv.ParseBool(val)
		case "NEKO_IMPLICIT_CONTROL":
			settings.ImplicitControl, err = strconv.ParseBool(val)
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
			if !slices.Contains(blacklistedEnvsV2, key) {
				settings.Envs[key] = val
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}
