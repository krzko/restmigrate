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

var (
	Version = "unknown"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	ctx := context.Background()
	var shutdownTelemetry func(context.Context) error

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
		Before: func(cliCtx *cli.Context) error {
			if cliCtx.Bool("debug") {
				logger.SetLevel(log.DebugLevel)
			}

			var err error
			shutdownTelemetry, err = telemetry.InitTracer("restmigrate", os.Getenv("DEPLOYMENT_ENVIRONMENT"), nil)
			if err != nil {
				return fmt.Errorf("failed to initialize telemetry: %w", err)
			}

			return nil
		},
		After: func(cliCtx *cli.Context) error {
			if shutdownTelemetry != nil {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := shutdownTelemetry(shutdownCtx); err != nil {
					logger.Error("Failed to shutdown telemetry", "error", err)
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
		},
	}

	executor.SetConfig(executor.Config{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	})

	if err := app.RunContext(ctx, os.Args); err != nil {
		logger.Logger.Fatal(err)
	}
}

func VersionString() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
}

func wrapActionWithTelemetry(f func(context.Context, *cli.Context) error) cli.ActionFunc {
	return func(c *cli.Context) error {
		ctx, span := telemetry.StartSpan(c.Context, c.Command.Name)
		defer span.End()
		return f(ctx, c)
	}
}
