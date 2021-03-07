package main

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"m1k1o/neko_rooms"
	"m1k1o/neko_rooms/cmd"
	"m1k1o/neko_rooms/internal/utils"
)

func main() {
	fmt.Print(utils.Colorf(neko_rooms.Header, "server", neko_rooms.Service.Version))
	if err := cmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("failed to execute command")
	}
}
