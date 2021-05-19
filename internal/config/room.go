package config

import (
	"strconv"
	"strings"

	dockerNames "github.com/docker/docker/daemon/names"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Room struct {
	EprMin uint16
	EprMax uint16

	NAT1To1IPs []string
	NekoImages []string

	InstanceName string
	InstanceUrl  string
	InstanceData string

	TraefikDomain       string
	TraefikEntrypoint   string
	TraefikCertresolver string
	TraefikNetwork      string
	TraefikPort         string // deprecated
}

func (Room) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("epr", "59000-59999", "limits the pool of ephemeral ports that ICE UDP connections can allocate from")
	if err := viper.BindPFlag("epr", cmd.PersistentFlags().Lookup("epr")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("nat1to1", []string{}, "sets a list of external IP addresses of 1:1 (D)NAT and a candidate type for which the external IP address is used")
	if err := viper.BindPFlag("nat1to1", cmd.PersistentFlags().Lookup("nat1to1")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("neko_images", []string{
		"m1k1o/neko:latest",
		"m1k1o/neko:chromium",
		"m1k1o/neko:ungoogled-chromium",
		"m1k1o/neko:tor-browser",
		"m1k1o/neko:vlc",
		"m1k1o/neko:vncviewer",
		"m1k1o/neko:xfce",
	}, "neko images to be used")
	if err := viper.BindPFlag("neko_images", cmd.PersistentFlags().Lookup("neko_images")); err != nil {
		return err
	}

	// Instance

	cmd.PersistentFlags().String("instance.name", "neko-rooms", "unique instance name (if running muliple on the same host)")
	if err := viper.BindPFlag("instance.name", cmd.PersistentFlags().Lookup("instance.name")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("instance.url", "", "instance url that is prefixing room names (if different from `http(s)://{traefik_domain}/`)")
	if err := viper.BindPFlag("instance.url", cmd.PersistentFlags().Lookup("instance.url")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("instance.data", "", "absolute path on host to a folder, where peristent containers data will be stored")
	if err := viper.BindPFlag("instance.data", cmd.PersistentFlags().Lookup("instance.url")); err != nil {
		return err
	}

	// Traefik

	cmd.PersistentFlags().String("traefik.domain", "neko.lan", "traefik: domain on which will be container hosted")
	if err := viper.BindPFlag("traefik.domain", cmd.PersistentFlags().Lookup("traefik.domain")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("traefik.entrypoint", "web-secure", "traefik: router entrypoint")
	if err := viper.BindPFlag("traefik.entrypoint", cmd.PersistentFlags().Lookup("traefik.entrypoint")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("traefik.certresolver", "", "traefik: certificate resolver for router")
	if err := viper.BindPFlag("traefik.certresolver", cmd.PersistentFlags().Lookup("traefik.certresolver")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("traefik.network", "traefik", "traefik: docker network name")
	if err := viper.BindPFlag("traefik.network", cmd.PersistentFlags().Lookup("traefik.network")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("traefik.port", "", "traefik: external port (deprecated)")
	if err := viper.BindPFlag("traefik.port", cmd.PersistentFlags().Lookup("traefik.port")); err != nil {
		return err
	}

	return nil
}

func (s *Room) Set() {
	min := uint16(59000)
	max := uint16(59999)
	epr := viper.GetString("epr")
	ports := strings.SplitN(epr, "-", -1)
	if len(ports) > 1 {
		start, err := strconv.ParseUint(ports[0], 10, 16)
		if err == nil {
			min = uint16(start)
		}

		end, err := strconv.ParseUint(ports[1], 10, 16)
		if err == nil {
			max = uint16(end)
		}
	}

	if min > max {
		s.EprMin = max
		s.EprMax = min
	} else {
		s.EprMin = min
		s.EprMax = max
	}

	s.NAT1To1IPs = viper.GetStringSlice("nat1to1")
	s.NekoImages = viper.GetStringSlice("neko_images")

	s.InstanceName = viper.GetString("instance.name")
	if !dockerNames.RestrictedNamePattern.MatchString(s.InstanceName) {
		log.Panic().Msg("invalid `instance.name`, must match " + dockerNames.RestrictedNameChars)
	}

	s.InstanceUrl = viper.GetString("instance.url")
	s.InstanceData = viper.GetString("instance.data")
	if s.InstanceData != "" {
		if !strings.HasPrefix(s.InstanceData, "/") {
			log.Panic().Msg("invalid `instance.data`, must be absolute path starting with /")
		}

		if strings.Contains(s.InstanceData, ":") {
			log.Panic().Msg("invalid `instance.data`, cannot contain : character")
		}
	} else {
		log.Warn().Msg("missing `instance.data`, container mounts are unavailable")
	}

	s.TraefikDomain = viper.GetString("traefik.domain")
	s.TraefikEntrypoint = viper.GetString("traefik.entrypoint")
	s.TraefikCertresolver = viper.GetString("traefik.certresolver")
	s.TraefikNetwork = viper.GetString("traefik.network")

	// deprecated
	s.TraefikPort = viper.GetString("traefik.port")
	if s.TraefikPort != "" {
		if s.InstanceUrl != "" {
			log.Warn().Msg("deprecated `traefik.port` config item is ignored when `instance.url` is set")
		} else {
			log.Warn().Msg("you are using deprecated `traefik.port` config item, you should consider moving to `instance.url`")
		}
	}
}
