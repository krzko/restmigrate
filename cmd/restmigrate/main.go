package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/krzko/restmigrate/internal/executor"
	"github.com/krzko/restmigrate/internal/logger"
	"github.com/krzko/restmigrate/internal/migration"
	"github.com/urfave/cli/v2"
)

var (
	Version = "unknown"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	executor.SetConfig(executor.Config{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	})

	app := &cli.App{
		Name:    "restmigrate",
		Usage:   "Migrate REST API configurations",
		Version: VersionString(),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable debug logging",
			},
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "Path to migrations directory",
				Value:   ".",
			},
		},
		Before: func(c *cli.Context) error {
			if c.Bool("debug") {
				logger.SetLevel(log.DebugLevel)
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create a new migration",
				Action:  migration.CreateMigration,
			},
			{
				Name:    "up",
				Aliases: []string{"u"},
				Usage:   "Apply all pending migrations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "base-url",
						Aliases:  []string{"u"},
						Usage:    "Base URL for the API",
						EnvVars:  []string{"RESTMIGRATE_BASE_URL"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    "api-key",
						Aliases: []string{"k"},
						Usage:   "API Key for authentication",
						EnvVars: []string{"RESTMIGRATE_API_KEY"},
					},
					&cli.StringFlag{
						Name:    "type",
						Aliases: []string{"t"},
						Usage:   "API gateway type (apisix, kong, generic)",
						Value:   "generic",
						EnvVars: []string{"RESTMIGRATE_API_TYPE"},
					},
				},
				Action: executor.ExecuteUp,
			},
			{
				Name:    "down",
				Aliases: []string{"d", "rollback"},
				Usage:   "Revert migration/s",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Usage: "Revert all applied migrations",
					},
					&cli.StringFlag{
						Name:     "base-url",
						Aliases:  []string{"u"},
						Usage:    "Base URL for the API",
						EnvVars:  []string{"RESTMIGRATE_BASE_URL"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    "api-key",
						Aliases: []string{"k"},
						Usage:   "API Key for authentication",
						EnvVars: []string{"RESTMIGRATE_API_KEY"},
					},
					&cli.StringFlag{
						Name:    "type",
						Aliases: []string{"t"},
						Usage:   "API gateway type (apisix, kong, generic)",
						Value:   "generic",
						EnvVars: []string{"RESTMIGRATE_API_TYPE"},
					},
				},
				Action: executor.ExecuteDown,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Logger.Fatal(err)
	}
}

func VersionString() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
}
