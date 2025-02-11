// Package main is the entry point for the contracts-api application.
package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/CTO2BPublic/passage-server/pkg/api"
	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/crondriver"

	"github.com/urfave/cli/v2"
)

// @version         0.1.0
// @title 					passage-server
// @description 		powerful, open-source access control management solution built in Go
// @contact.name 		API Support
// @contact.url 		https://cto2b.eu
// @contact.email 	tomas@cto2b.eu
// @license.name 		Apache 2.0
// @securityDefinitions.apikey JWT
// @in header
// @name Authorization

func main() {

	config.InitConfig()
	Config := config.GetConfig()

	if Config.Log.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if Config.Log.Level == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		config.PrintConfig(Config)
	}
	if Config.Log.Level == "info" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	if Config.Log.Caller {
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			parts := strings.SplitAfter(file, "passage-server")
			return parts[1] + ":" + strconv.Itoa(line)
		}
		log.Logger = log.With().Caller().Logger()
	}

	app := &cli.App{

		// CLI metadata
		Name:                 "passage-server",
		Usage:                "powerful, open-source access control management solution built in Go",
		Version:              "0.1.0",
		EnableBashCompletion: true,

		// Default command if none is passed
		Action: func(cCtx *cli.Context) error {

			cron := crondriver.GetDriver()
			cron.Start()

			server := api.GetServer()
			server.SetupEngineWithDefaults()
			server.RunEngine()
			return nil
		},

		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "Manages API server",
				Subcommands: []*cli.Command{

					// SERVER start s
					{
						Name:    "start",
						Usage:   "Start API server",
						Aliases: []string{"s"},
						Action: func(c *cli.Context) error {

							cron := crondriver.GetDriver()
							cron.Start()

							server := api.GetServer()
							server.SetupEngineWithDefaults()
							server.RunEngine()

							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

}
