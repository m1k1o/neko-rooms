package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Admin struct {
	Static     string
	Password   string
	PathPrefix string
}

type Server struct {
	Cert  string
	Key   string
	Bind  string
	Proxy bool
	PProf bool

	Admin Admin
}

func (Server) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("bind", "127.0.0.1:8080", "address/port/socket to serve neko_rooms")
	if err := viper.BindPFlag("bind", cmd.PersistentFlags().Lookup("bind")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("cert", "", "path to the SSL cert used to secure the neko_rooms server")
	if err := viper.BindPFlag("cert", cmd.PersistentFlags().Lookup("cert")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("key", "", "path to the SSL key used to secure the neko_rooms server")
	if err := viper.BindPFlag("key", cmd.PersistentFlags().Lookup("key")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("proxy", false, "trust reverse proxy headers")
	if err := viper.BindPFlag("proxy", cmd.PersistentFlags().Lookup("proxy")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("pprof", false, "enable pprof endpoint available at /debug/pprof")
	if err := viper.BindPFlag("pprof", cmd.PersistentFlags().Lookup("pprof")); err != nil {
		return err
	}

	// Admin

	cmd.PersistentFlags().String("admin.static", "", "path to neko_rooms admin client files to serve")
	if err := viper.BindPFlag("admin.static", cmd.PersistentFlags().Lookup("admin.static")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("admin.password", "", "admin password")
	if err := viper.BindPFlag("admin.password", cmd.PersistentFlags().Lookup("admin.password")); err != nil {
		return err
	}

	// TODO: Default in v2 will be '/admin'.
	cmd.PersistentFlags().String("admin.path_prefix", "", "set custom path prefix for admin")
	if err := viper.BindPFlag("admin.path_prefix", cmd.PersistentFlags().Lookup("admin.path_prefix")); err != nil {
		return err
	}

	return nil
}

func (s *Server) Set() {
	s.Cert = viper.GetString("cert")
	s.Key = viper.GetString("key")
	s.Bind = viper.GetString("bind")
	s.Proxy = viper.GetBool("proxy")
	s.PProf = viper.GetBool("pprof")

	s.Admin.Static = viper.GetString("admin.static")
	s.Admin.Password = viper.GetString("admin.password")
	s.Admin.PathPrefix = viper.GetString("admin.path_prefix")
}
