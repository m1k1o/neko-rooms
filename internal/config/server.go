package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Admin struct {
	Static    string
	ProxyAuth string
	Username  string
	Password  string
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

	cmd.PersistentFlags().String("admin.proxy_auth", "", "require auth: proxy authentication URL, only allow if it returns 200")
	if err := viper.BindPFlag("admin.proxy_auth", cmd.PersistentFlags().Lookup("admin.proxy_auth")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("admin.username", "admin", "require auth: admin username")
	if err := viper.BindPFlag("admin.username", cmd.PersistentFlags().Lookup("admin.username")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("admin.password", "", "require auth: admin password")
	if err := viper.BindPFlag("admin.password", cmd.PersistentFlags().Lookup("admin.password")); err != nil {
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
	s.Admin.ProxyAuth = viper.GetString("admin.proxy_auth")
	s.Admin.Username = viper.GetString("admin.username")
	s.Admin.Password = viper.GetString("admin.password")
}
