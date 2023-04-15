package neko_rooms

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/m1k1o/neko-rooms/internal/api"
	"github.com/m1k1o/neko-rooms/internal/config"
	"github.com/m1k1o/neko-rooms/internal/proxy"
	"github.com/m1k1o/neko-rooms/internal/pull"
	"github.com/m1k1o/neko-rooms/internal/room"
	"github.com/m1k1o/neko-rooms/internal/server"
)

const Header = `&34
               __                                              
   ____  ___  / /______        _________  ____  ____ ___  _____
  / __ \/ _ \/ //_/ __ \      / ___/ __ \/ __ \/ __ '__ \/ ___/
 / / / /  __/ ,< / /_/ /_____/ /  / /_/ / /_/ / / / / / (__  ) 
/_/ /_/\___/_/|_|\____/_____/_/   \____/\____/_/ /_/ /_/____/  
                                                               
&1&37                    by m1k1o                   &33%s v%s&0
`

var (
	//
	buildDate = "dev"
	//
	gitCommit = "dev"
	//
	gitBranch = "dev"

	// Major version when you make incompatible API changes,
	major = "1"
	// Minor version when you add functionality in a backwards-compatible manner, and
	minor = "0"
	// Patch version when you make backwards-compatible bug fixes.
	patch = "0"
)

var Service *MainCtx

func init() {
	Service = &MainCtx{
		Version: &Version{
			Major:     major,
			Minor:     minor,
			Patch:     patch,
			GitCommit: gitCommit,
			GitBranch: gitBranch,
			BuildDate: buildDate,
			GoVersion: runtime.Version(),
			Compiler:  runtime.Compiler,
			Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		},
		Configs: &Configs{
			Root:   &config.Root{},
			Server: &config.Server{},
			Room:   &config.Room{},
		},
	}
}

type Version struct {
	Major     string
	Minor     string
	Patch     string
	GitCommit string
	GitBranch string
	BuildDate string
	GoVersion string
	Compiler  string
	Platform  string
}

func (i *Version) String() string {
	return fmt.Sprintf("%s.%s.%s %s", i.Major, i.Minor, i.Patch, i.GitCommit)
}

func (i *Version) Details() string {
	return fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
		fmt.Sprintf("Version %s.%s.%s", i.Major, i.Minor, i.Patch),
		fmt.Sprintf("GitCommit %s", i.GitCommit),
		fmt.Sprintf("GitBranch %s", i.GitBranch),
		fmt.Sprintf("BuildDate %s", i.BuildDate),
		fmt.Sprintf("GoVersion %s", i.GoVersion),
		fmt.Sprintf("Compiler %s", i.Compiler),
		fmt.Sprintf("Platform %s", i.Platform),
	)
}

type Configs struct {
	Root   *config.Root
	Server *config.Server
	Room   *config.Room
}

type MainCtx struct {
	Version *Version
	Configs *Configs

	logger        zerolog.Logger
	roomManager   *room.RoomManagerCtx
	pullManager   *pull.PullManagerCtx
	apiManager    *api.ApiManagerCtx
	proxyManager  *proxy.ProxyManagerCtx
	serverManager *server.ServerManagerCtx
}

func (main *MainCtx) Preflight() {
	main.logger = log.With().Str("service", "neko_rooms").Logger()
}

func (main *MainCtx) Start() {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		main.logger.Panic().Err(err).Msg("unable to connect to docker client")
	} else {
		main.logger.Info().Msg("successfully connected to docker client")
	}

	main.roomManager = room.New(
		client,
		main.Configs.Room,
	)

	main.pullManager = pull.New(
		client,
		main.Configs.Room.NekoImages,
	)

	main.apiManager = api.New(
		main.roomManager,
		main.pullManager,
	)

	main.proxyManager = proxy.New(
		client,
		main.Configs.Room.InstanceName,
		main.Configs.Room.WaitEnabled,
	)
	main.proxyManager.Start()

	main.serverManager = server.New(
		main.apiManager,
		main.Configs.Room,
		main.Configs.Server,
		main.proxyManager,
	)
	main.serverManager.Start()
}

func (main *MainCtx) Shutdown() {
	var err error

	err = main.serverManager.Shutdown()
	main.logger.Err(err).Msg("server manager shutdown")

	err = main.proxyManager.Shutdown()
	main.logger.Err(err).Msg("proxy manager shutdown")
}

func (main *MainCtx) ServeCommand(cmd *cobra.Command, args []string) {
	main.logger.Info().Msg("starting neko_rooms server")
	main.Start()
	main.logger.Info().Msg("neko_rooms ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit

	main.logger.Warn().Msgf("received %s, attempting graceful shutdown.", sig)
	main.Shutdown()
	main.logger.Info().Msg("shutdown complete")
}
