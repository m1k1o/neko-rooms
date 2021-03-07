package config

import (
	"github.com/spf13/cobra"
)

type Room struct {
}

func (Room) Init(cmd *cobra.Command) error {

	return nil
}

func (s *Room) Set() {

}
