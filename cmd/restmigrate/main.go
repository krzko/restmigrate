package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/krzko/restmigrate/internal/executor"
	"github.com/krzko/restmigrate/internal/logger"
	"github.com/krzko/restmigrate/internal/migration"
	"github.com/krzko/restmigrate/internal/telemetry"
	"github.com/urfave/cli/v2"
)

const appName = "restmigrate"

var (
	Version = "unknown"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	ctx := context.Background()
	var shutdownTelemetry func(context.Context) error

	app := &cli.App{
		Name:    appName,
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
		Before: func(cliCtx *cli.Context) error {
			if cliCtx.Bool("debug") {
				logger.GetLogger().SetLevel(log.DebugLevel)
			}

			var err error
			shutdownTelemetry, err = telemetry.InitTracer("restmigrate", nil)
			if err != nil {
				logger.Error("Failed to initialise telemetry", "error", err)
			} else {
				logger.Debug("Telemetry initialised successfully")
			}

			return nil
		},
		After: func(cliCtx *cli.Context) error {
			if shutdownTelemetry != nil {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
				defer cancel()

				logger.Debug("Starting telemetry shutdown")
				if err := shutdownTelemetry(shutdownCtx); err != nil {
					logger.Error("Failed to shutdown telemetry", "error", err)
				} else {
					logger.Debug("Telemetry shutdown completed")
				}
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create a new migration",
				Action:  wrapActionWithTelemetry(migration.CreateMigration),
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
				Action: wrapActionWithTelemetry(executor.ExecuteDown),
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List all applied migrations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "Path to migrations directory",
						Value:   ".",
					},
				},
				Action: wrapActionWithTelemetry(executor.ListMigrations),
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
				Action: wrapActionWithTelemetry(executor.ExecuteUp),
			},
		},
	}

	executor.SetConfig(executor.Config{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	})

	if err := app.RunContext(ctx, os.Args); err != nil {
		logger.Fatal("Failed to run application", "error", err)
	}
}

func VersionString() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
}

func wrapActionWithTelemetry(f func(context.Context, *cli.Context) error) cli.ActionFunc {
	return func(c *cli.Context) error {
		commandName := fmt.Sprintf("%s %s", appName, c.Command.Name)
		ctx, span := telemetry.StartSpan(c.Context, commandName)
		defer span.End()
		return f(ctx, c)
	}
}
