package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type API struct {
	Lobby bool
}

func (API) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("lobby", true, "show lobby when room is not started yet; in order to show lobby, neko-rooms must run on the same subdomain where rooms are published")
	if err := viper.BindPFlag("lobby", cmd.PersistentFlags().Lookup("lobby")); err != nil {
		return err
	}

	return nil
}

func (s *API) Set() {
	s.Lobby = viper.GetBool("lobby")
}
