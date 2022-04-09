package config

import (
	"net/url"
	"path"
	"path/filepath"
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

	NAT1To1IPs           []string
	NekoImages           []string
	NekoPrivilegedImages []string
	PathPrefix           string
	Labels               []string

	StorageEnabled  bool
	StorageInternal string
	StorageExternal string

	MountsWhitelist []string

	InstanceName string
	InstanceUrl  *url.URL

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
		"m1k1o/neko:firefox",
		"m1k1o/neko:chromium",
		"m1k1o/neko:google-chrome",
		"m1k1o/neko:ungoogled-chromium",
		"m1k1o/neko:microsoft-edge",
		"m1k1o/neko:brave",
		"m1k1o/neko:tor-browser",
		"m1k1o/neko:vlc",
		"m1k1o/neko:remmina",
		"m1k1o/neko:xfce",
	}, "neko images to be used")
	if err := viper.BindPFlag("neko_images", cmd.PersistentFlags().Lookup("neko_images")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("neko_privileged_images", []string{}, "Whitelist of images allowed to be executed with Privileged mode")
	if err := viper.BindPFlag("neko_privileged_images", cmd.PersistentFlags().Lookup("neko_privileged_images")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("path_prefix", "", "path prefix that is added to every room path")
	if err := viper.BindPFlag("path_prefix", cmd.PersistentFlags().Lookup("path_prefix")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("labels", []string{}, "additional labels appended to every room")
	if err := viper.BindPFlag("labels", cmd.PersistentFlags().Lookup("labels")); err != nil {
		return err
	}

	// Data

	cmd.PersistentFlags().Bool("storage.enabled", true, "whether storage is enabled, where peristent containers data will be stored")
	if err := viper.BindPFlag("storage.enabled", cmd.PersistentFlags().Lookup("storage.enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("storage.external", "", "external absolute path (on the host) to storage folder")
	if err := viper.BindPFlag("storage.external", cmd.PersistentFlags().Lookup("storage.external")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("storage.internal", "/data", "internal absolute path (inside container) to storage folder")
	if err := viper.BindPFlag("storage.internal", cmd.PersistentFlags().Lookup("storage.internal")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("mounts.whitelist", []string{}, "whitelisted public mounts for containers")
	if err := viper.BindPFlag("mounts.whitelist", cmd.PersistentFlags().Lookup("mounts.whitelist")); err != nil {
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

	// Traefik

	cmd.PersistentFlags().String("traefik.domain", "", "traefik: domain on which will be container hosted (if empty or '*', match all; for neko-rooms as subdomain use '*.domain.tld')")
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
	s.NekoPrivilegedImages = viper.GetStringSlice("neko_privileged_images")
	s.PathPrefix = viper.GetString("path_prefix")
	s.Labels = viper.GetStringSlice("labels")

	s.StorageEnabled = viper.GetBool("storage.enabled")
	s.StorageInternal = viper.GetString("storage.internal")
	s.StorageExternal = viper.GetString("storage.external")

	if s.StorageInternal != "" && s.StorageExternal != "" {
		s.StorageInternal = filepath.Clean(s.StorageInternal)
		s.StorageExternal = filepath.Clean(s.StorageExternal)

		if !filepath.IsAbs(s.StorageInternal) || !filepath.IsAbs(s.StorageExternal) {
			log.Panic().Msg("invalid `storage.internal` or `storage.external`, must be an absolute path")
		}
	} else {
		log.Warn().Msg("missing `storage.internal` or `storage.external`, storage is unavailable")
		s.StorageEnabled = false
	}

	s.MountsWhitelist = viper.GetStringSlice("mounts.whitelist")
	for _, path := range s.MountsWhitelist {
		path = filepath.Clean(path)

		if !filepath.IsAbs(path) {
			log.Panic().Msg("invalid `mounts.whitelist`, must be an absolute path")
		}
	}

	s.InstanceName = viper.GetString("instance.name")
	if !dockerNames.RestrictedNamePattern.MatchString(s.InstanceName) {
		log.Panic().Msg("invalid `instance.name`, must match " + dockerNames.RestrictedNameChars)
	}

	instanceUrl := viper.GetString("instance.url")
	if instanceUrl != "" {
		var err error
		s.InstanceUrl, err = url.Parse(instanceUrl)
		if err != nil {
			log.Panic().Err(err).Msg("invalid `instance.url`")
		}
	}

	s.TraefikDomain = viper.GetString("traefik.domain")
	s.TraefikEntrypoint = viper.GetString("traefik.entrypoint")
	s.TraefikCertresolver = viper.GetString("traefik.certresolver")
	s.TraefikNetwork = viper.GetString("traefik.network")

	// deprecated
	s.TraefikPort = viper.GetString("traefik.port")
	if s.TraefikPort != "" {
		if s.InstanceUrl != nil {
			log.Warn().Msg("deprecated `traefik.port` config item is ignored when `instance.url` is set")
		} else {
			log.Warn().Msg("you are using deprecated `traefik.port` config item, you should consider moving to `instance.url`")
		}
	}
}

func (s *Room) GetInstanceUrl() url.URL {
	if s.InstanceUrl != nil {
		return *s.InstanceUrl
	}

	instanceUrl := url.URL{
		Scheme: "http",
		Host:   "127.0.0.1",
		Path:   "/",
	}

	if s.TraefikCertresolver != "" {
		instanceUrl.Scheme = "https"
	}

	if s.TraefikDomain != "" && s.TraefikDomain != "*" {
		instanceUrl.Host = s.TraefikDomain
	}

	// deprecated
	if s.TraefikPort != "" {
		instanceUrl.Host += ":" + s.TraefikPort
	}

	return instanceUrl
}

func (s *Room) GetRoomUrl(roomName string) string {
	instanceUrl := s.GetInstanceUrl()

	if strings.HasPrefix(instanceUrl.Host, "*.") {
		instanceUrl.Host = roomName + "." + strings.TrimPrefix(instanceUrl.Host, "*.")
	} else {
		instanceUrl.Path = path.Join(instanceUrl.Path, s.PathPrefix, roomName) + "/"
	}

	return instanceUrl.String()
}
