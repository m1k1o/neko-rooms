package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"m1k1o/neko_rooms"
	"m1k1o/neko_rooms/internal/config"
)

func init() {
	command := &cobra.Command{
		Use:   "serve",
		Short: "serve neko_rooms server",
		Long:  `serve neko_rooms server`,
		Run:   neko_rooms.Service.ServeCommand,
	}

	configs := []config.Config{
		neko_rooms.Service.Configs.Server,
		neko_rooms.Service.Configs.API,
		neko_rooms.Service.Configs.Room,
	}

	cobra.OnInitialize(func() {
		for _, cfg := range configs {
			cfg.Set()
		}
		neko_rooms.Service.Preflight()
	})

	for _, cfg := range configs {
		if err := cfg.Init(command); err != nil {
			log.Panic().Err(err).Msg("unable to run serve command")
		}
	}

	root.AddCommand(command)
}
