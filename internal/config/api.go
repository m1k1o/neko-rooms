package config

import (
	"github.com/spf13/cobra"
)

type API struct {
}

func (API) Init(cmd *cobra.Command) error {

	return nil
}

func (s *API) Set() {

}
